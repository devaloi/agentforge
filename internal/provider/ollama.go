package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Ollama implements the Provider interface for the local Ollama API.
type Ollama struct {
	baseURL string
	client  *http.Client
}

// NewOllama creates an Ollama provider pointed at the given base URL.
func NewOllama(baseURL string) *Ollama {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	return &Ollama{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

// Name returns "ollama".
func (o *Ollama) Name() string { return "ollama" }

// ChatComplete sends a chat request to the Ollama API.
func (o *Ollama) ChatComplete(ctx context.Context, messages []Message, tools []ToolDef, cfg ChatConfig) (*Response, error) {
	body := o.buildRequest(messages, tools, cfg)

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.baseURL+"/api/chat", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return o.parseResponse(respBody)
}

type ollamaRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Tools    []ollamaToolDef `json:"tools,omitempty"`
	Stream   bool            `json:"stream"`
}

type ollamaMessage struct {
	Role      string           `json:"role"`
	Content   string           `json:"content"`
	ToolCalls []ollamaToolCall `json:"tool_calls,omitempty"`
}

type ollamaToolCall struct {
	Function ollamaToolCallFunc `json:"function"`
}

type ollamaToolCallFunc struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type ollamaToolDef struct {
	Type     string        `json:"type"`
	Function ollamaFuncDef `json:"function"`
}

type ollamaFuncDef struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Parameters  JSONSchema `json:"parameters"`
}

type ollamaResponse struct {
	Message ollamaMessage `json:"message"`
}

func (o *Ollama) buildRequest(messages []Message, tools []ToolDef, cfg ChatConfig) ollamaRequest {
	ollamaMessages := make([]ollamaMessage, 0, len(messages))
	for _, m := range messages {
		msg := ollamaMessage{
			Role:    string(m.Role),
			Content: m.Content,
		}
		ollamaMessages = append(ollamaMessages, msg)
	}

	ollamaTools := make([]ollamaToolDef, 0, len(tools))
	for _, t := range tools {
		ollamaTools = append(ollamaTools, ollamaToolDef{
			Type:     "function",
			Function: ollamaFuncDef(t),
		})
	}

	return ollamaRequest{
		Model:    cfg.Model,
		Messages: ollamaMessages,
		Tools:    ollamaTools,
		Stream:   false,
	}
}

func (o *Ollama) parseResponse(data []byte) (*Response, error) {
	var ollamaResp ollamaResponse
	if err := json.Unmarshal(data, &ollamaResp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	resp := &Response{
		Content: ollamaResp.Message.Content,
	}

	for i, tc := range ollamaResp.Message.ToolCalls {
		args, _ := json.Marshal(tc.Function.Arguments)
		resp.ToolCalls = append(resp.ToolCalls, ToolCall{
			ID:        fmt.Sprintf("ollama_call_%d", i),
			Name:      tc.Function.Name,
			Arguments: string(args),
		})
	}

	return resp, nil
}
