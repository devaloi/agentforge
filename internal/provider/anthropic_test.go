package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAnthropicChatComplete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/messages" {
			t.Errorf("path = %s, want /v1/messages", r.URL.Path)
		}
		if r.Header.Get("x-api-key") != "test-key" {
			t.Errorf("x-api-key = %q", r.Header.Get("x-api-key"))
		}
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Errorf("anthropic-version = %q", r.Header.Get("anthropic-version"))
		}

		resp := anthropicResponse{
			Content:    []anthropicContentBlock{{Type: "text", Text: "Hello from Claude"}},
			StopReason: "end_turn",
			Usage:      anthropicUsage{InputTokens: 10, OutputTokens: 5},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewAnthropic("test-key", server.URL)
	resp, err := p.ChatComplete(context.Background(), []Message{
		{Role: RoleUser, Content: "Hello"},
	}, nil, ChatConfig{Model: "claude-sonnet-4-20250514"})

	if err != nil {
		t.Fatalf("ChatComplete: %v", err)
	}
	if resp.Content != "Hello from Claude" {
		t.Errorf("Content = %q, want %q", resp.Content, "Hello from Claude")
	}
	if resp.Usage.TotalTokens != 15 {
		t.Errorf("TotalTokens = %d, want 15", resp.Usage.TotalTokens)
	}
}

func TestAnthropicToolUse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := anthropicResponse{
			Content: []anthropicContentBlock{
				{Type: "text", Text: "Let me search for that."},
				{
					Type:  "tool_use",
					ID:    "toolu_123",
					Name:  "web_search",
					Input: map[string]any{"query": "Go REST API"},
				},
			},
			StopReason: "tool_use",
			Usage:      anthropicUsage{InputTokens: 20, OutputTokens: 10},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewAnthropic("test-key", server.URL)
	resp, err := p.ChatComplete(context.Background(), []Message{
		{Role: RoleUser, Content: "Search for Go REST APIs"},
	}, []ToolDef{
		{Name: "web_search", Description: "Search the web"},
	}, ChatConfig{Model: "claude-sonnet-4-20250514"})

	if err != nil {
		t.Fatalf("ChatComplete: %v", err)
	}
	if resp.Content != "Let me search for that." {
		t.Errorf("Content = %q", resp.Content)
	}
	if len(resp.ToolCalls) != 1 {
		t.Fatalf("ToolCalls count = %d, want 1", len(resp.ToolCalls))
	}
	if resp.ToolCalls[0].Name != "web_search" {
		t.Errorf("ToolCall name = %q, want %q", resp.ToolCalls[0].Name, "web_search")
	}
}

func TestAnthropicError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": {"message": "invalid request"}}`))
	}))
	defer server.Close()

	p := NewAnthropic("test-key", server.URL)
	_, err := p.ChatComplete(context.Background(), []Message{
		{Role: RoleUser, Content: "Hello"},
	}, nil, ChatConfig{Model: "claude-sonnet-4-20250514"})

	if err == nil {
		t.Fatal("expected error for bad request")
	}
}

func TestAnthropicName(t *testing.T) {
	p := NewAnthropic("key", "")
	if p.Name() != "anthropic" {
		t.Errorf("Name = %q, want %q", p.Name(), "anthropic")
	}
}
