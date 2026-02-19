package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// OpenAI implements the Provider interface for OpenAI's chat completions API.
type OpenAI struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewOpenAI creates an OpenAI provider with the given API key.
func NewOpenAI(apiKey, baseURL string) *OpenAI {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	return &OpenAI{
		apiKey:  apiKey,
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

// Name returns "openai".
func (o *OpenAI) Name() string { return "openai" }

// ChatComplete sends a chat completion request to the OpenAI API.
func (o *OpenAI) ChatComplete(ctx context.Context, messages []Message, tools []ToolDef, cfg ChatConfig) (*Response, error) {
	body := o.buildRequest(messages, tools, cfg)

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return o.parseResponse(respBody)
}

type openaiRequest struct {
	Model    string            `json:"model"`
	Messages []openaiMessage   `json:"messages"`
	Tools    []openaiTool      `json:"tools,omitempty"`
}

type openaiMessage struct {
	Role       string           `json:"role"`
	Content    string           `json:"content,omitempty"`
	ToolCalls  []openaiToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
}

type openaiToolCall struct {
	ID       string               `json:"id"`
	Type     string               `json:"type"`
	Function openaiToolCallFunc   `json:"function"`
}

type openaiToolCallFunc struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type openaiTool struct {
	Type     string           `json:"type"`
	Function openaiToolDef    `json:"function"`
}

type openaiToolDef struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Parameters  JSONSchema `json:"parameters"`
}

type openaiResponse struct {
	Choices []openaiChoice `json:"choices"`
	Usage   openaiUsage    `json:"usage"`
}

type openaiChoice struct {
	Message openaiMessage `json:"message"`
}

type openaiUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func (o *OpenAI) buildRequest(messages []Message, tools []ToolDef, cfg ChatConfig) openaiRequest {
	oaiMessages := make([]openaiMessage, 0, len(messages))
	for _, m := range messages {
		msg := openaiMessage{
			Role:       string(m.Role),
			Content:    m.Content,
			ToolCallID: m.ToolCallID,
		}
		for _, tc := range m.ToolCalls {
			msg.ToolCalls = append(msg.ToolCalls, openaiToolCall{
				ID:   tc.ID,
				Type: "function",
				Function: openaiToolCallFunc{
					Name:      tc.Name,
					Arguments: tc.Arguments,
				},
			})
		}
		oaiMessages = append(oaiMessages, msg)
	}

	oaiTools := make([]openaiTool, 0, len(tools))
	for _, t := range tools {
		oaiTools = append(oaiTools, openaiTool{
			Type: "function",
			Function: openaiToolDef{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
			},
		})
	}

	return openaiRequest{
		Model:    cfg.Model,
		Messages: oaiMessages,
		Tools:    oaiTools,
	}
}

func (o *OpenAI) parseResponse(data []byte) (*Response, error) {
	var oaiResp openaiResponse
	if err := json.Unmarshal(data, &oaiResp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if len(oaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice := oaiResp.Choices[0]
	resp := &Response{
		Content: choice.Message.Content,
		Usage: Usage{
			PromptTokens:     oaiResp.Usage.PromptTokens,
			CompletionTokens: oaiResp.Usage.CompletionTokens,
			TotalTokens:      oaiResp.Usage.TotalTokens,
		},
	}

	for _, tc := range choice.Message.ToolCalls {
		resp.ToolCalls = append(resp.ToolCalls, ToolCall{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: tc.Function.Arguments,
		})
	}

	return resp, nil
}
