package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Anthropic implements the Provider interface for Anthropic's messages API.
type Anthropic struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewAnthropic creates an Anthropic provider with the given API key.
func NewAnthropic(apiKey, baseURL string) *Anthropic {
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}
	return &Anthropic{
		apiKey:  apiKey,
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

// Name returns "anthropic".
func (a *Anthropic) Name() string { return "anthropic" }

// ChatComplete sends a message to the Anthropic messages API.
func (a *Anthropic) ChatComplete(ctx context.Context, messages []Message, tools []ToolDef, cfg ChatConfig) (*Response, error) {
	body := a.buildRequest(messages, tools, cfg)

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/v1/messages", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anthropic API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return a.parseResponse(respBody)
}

type anthropicRequest struct {
	Model     string              `json:"model"`
	MaxTokens int                 `json:"max_tokens"`
	System    string              `json:"system,omitempty"`
	Messages  []anthropicMessage  `json:"messages"`
	Tools     []anthropicToolDef  `json:"tools,omitempty"`
}

type anthropicMessage struct {
	Role    string              `json:"role"`
	Content json.RawMessage     `json:"content"`
}

type anthropicContentBlock struct {
	Type      string `json:"type"`
	Text      string `json:"text,omitempty"`
	ID        string `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	Input     any    `json:"input,omitempty"`
	ToolUseID string `json:"tool_use_id,omitempty"`
	Content   string `json:"content,omitempty"`
}

type anthropicToolDef struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	InputSchema JSONSchema `json:"input_schema"`
}

type anthropicResponse struct {
	Content    []anthropicContentBlock `json:"content"`
	StopReason string                  `json:"stop_reason"`
	Usage      anthropicUsage          `json:"usage"`
}

type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

func (a *Anthropic) buildRequest(messages []Message, tools []ToolDef, cfg ChatConfig) anthropicRequest {
	var system string
	antMessages := make([]anthropicMessage, 0, len(messages))

	for _, m := range messages {
		if m.Role == RoleSystem {
			system = m.Content
			continue
		}

		if m.Role == RoleTool {
			blocks := []anthropicContentBlock{
				{Type: "tool_result", ToolUseID: m.ToolCallID, Content: m.Content},
			}
			raw, _ := json.Marshal(blocks)
			antMessages = append(antMessages, anthropicMessage{Role: "user", Content: raw})
			continue
		}

		if m.Role == RoleAssistant && len(m.ToolCalls) > 0 {
			blocks := make([]anthropicContentBlock, 0, len(m.ToolCalls)+1)
			if m.Content != "" {
				blocks = append(blocks, anthropicContentBlock{Type: "text", Text: m.Content})
			}
			for _, tc := range m.ToolCalls {
				var input any
				json.Unmarshal([]byte(tc.Arguments), &input)
				blocks = append(blocks, anthropicContentBlock{
					Type:  "tool_use",
					ID:    tc.ID,
					Name:  tc.Name,
					Input: input,
				})
			}
			raw, _ := json.Marshal(blocks)
			antMessages = append(antMessages, anthropicMessage{Role: "assistant", Content: raw})
			continue
		}

		raw, _ := json.Marshal(m.Content)
		antMessages = append(antMessages, anthropicMessage{Role: string(m.Role), Content: raw})
	}

	maxTokens := cfg.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 4096
	}

	antTools := make([]anthropicToolDef, 0, len(tools))
	for _, t := range tools {
		antTools = append(antTools, anthropicToolDef{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.Parameters,
		})
	}

	return anthropicRequest{
		Model:     cfg.Model,
		MaxTokens: maxTokens,
		System:    system,
		Messages:  antMessages,
		Tools:     antTools,
	}
}

func (a *Anthropic) parseResponse(data []byte) (*Response, error) {
	var antResp anthropicResponse
	if err := json.Unmarshal(data, &antResp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	resp := &Response{
		Usage: Usage{
			PromptTokens:     antResp.Usage.InputTokens,
			CompletionTokens: antResp.Usage.OutputTokens,
			TotalTokens:      antResp.Usage.InputTokens + antResp.Usage.OutputTokens,
		},
	}

	for _, block := range antResp.Content {
		switch block.Type {
		case "text":
			resp.Content += block.Text
		case "tool_use":
			args, _ := json.Marshal(block.Input)
			resp.ToolCalls = append(resp.ToolCalls, ToolCall{
				ID:        block.ID,
				Name:      block.Name,
				Arguments: string(args),
			})
		}
	}

	return resp, nil
}
