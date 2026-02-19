package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenAIChatComplete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/chat/completions" {
			t.Errorf("path = %s, want /chat/completions", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("auth header = %q", r.Header.Get("Authorization"))
		}

		resp := openaiResponse{
			Choices: []openaiChoice{
				{Message: openaiMessage{Role: "assistant", Content: "Hello from OpenAI"}},
			},
			Usage: openaiUsage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewOpenAI("test-key", server.URL)
	resp, err := p.ChatComplete(context.Background(), []Message{
		{Role: RoleUser, Content: "Hello"},
	}, nil, ChatConfig{Model: "gpt-4o"})

	if err != nil {
		t.Fatalf("ChatComplete: %v", err)
	}
	if resp.Content != "Hello from OpenAI" {
		t.Errorf("Content = %q, want %q", resp.Content, "Hello from OpenAI")
	}
	if resp.Usage.TotalTokens != 15 {
		t.Errorf("TotalTokens = %d, want 15", resp.Usage.TotalTokens)
	}
}

func TestOpenAIToolCall(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req openaiRequest
		json.NewDecoder(r.Body).Decode(&req)

		if len(req.Tools) != 1 {
			t.Errorf("tools count = %d, want 1", len(req.Tools))
		}

		resp := openaiResponse{
			Choices: []openaiChoice{
				{Message: openaiMessage{
					Role: "assistant",
					ToolCalls: []openaiToolCall{
						{
							ID:   "call_123",
							Type: "function",
							Function: openaiToolCallFunc{
								Name:      "web_search",
								Arguments: `{"query":"Go REST API"}`,
							},
						},
					},
				}},
			},
			Usage: openaiUsage{PromptTokens: 20, CompletionTokens: 10, TotalTokens: 30},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewOpenAI("test-key", server.URL)
	tools := []ToolDef{
		{
			Name:        "web_search",
			Description: "Search the web",
			Parameters: JSONSchema{
				Type: "object",
				Properties: map[string]SchemaField{
					"query": {Type: "string", Description: "Search query"},
				},
				Required: []string{"query"},
			},
		},
	}

	resp, err := p.ChatComplete(context.Background(), []Message{
		{Role: RoleUser, Content: "Search for Go REST APIs"},
	}, tools, ChatConfig{Model: "gpt-4o"})

	if err != nil {
		t.Fatalf("ChatComplete: %v", err)
	}
	if len(resp.ToolCalls) != 1 {
		t.Fatalf("ToolCalls count = %d, want 1", len(resp.ToolCalls))
	}
	if resp.ToolCalls[0].Name != "web_search" {
		t.Errorf("ToolCall name = %q, want %q", resp.ToolCalls[0].Name, "web_search")
	}
}

func TestOpenAIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error": {"message": "rate limit exceeded"}}`))
	}))
	defer server.Close()

	p := NewOpenAI("test-key", server.URL)
	_, err := p.ChatComplete(context.Background(), []Message{
		{Role: RoleUser, Content: "Hello"},
	}, nil, ChatConfig{Model: "gpt-4o"})

	if err == nil {
		t.Fatal("expected error for rate limit")
	}
}

func TestOpenAIName(t *testing.T) {
	p := NewOpenAI("key", "")
	if p.Name() != "openai" {
		t.Errorf("Name = %q, want %q", p.Name(), "openai")
	}
}
