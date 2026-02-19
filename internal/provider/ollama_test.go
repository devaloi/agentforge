package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOllamaChatComplete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/api/chat" {
			t.Errorf("path = %s, want /api/chat", r.URL.Path)
		}

		resp := ollamaResponse{
			Message: ollamaMessage{Role: "assistant", Content: "Hello from Ollama"},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewOllama(server.URL)
	resp, err := p.ChatComplete(context.Background(), []Message{
		{Role: RoleUser, Content: "Hello"},
	}, nil, ChatConfig{Model: "llama3"})

	if err != nil {
		t.Fatalf("ChatComplete: %v", err)
	}
	if resp.Content != "Hello from Ollama" {
		t.Errorf("Content = %q, want %q", resp.Content, "Hello from Ollama")
	}
}

func TestOllamaToolCall(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := ollamaResponse{
			Message: ollamaMessage{
				Role: "assistant",
				ToolCalls: []ollamaToolCall{
					{Function: ollamaToolCallFunc{
						Name:      "web_search",
						Arguments: map[string]any{"query": "Go REST API"},
					}},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewOllama(server.URL)
	resp, err := p.ChatComplete(context.Background(), []Message{
		{Role: RoleUser, Content: "Search for Go REST APIs"},
	}, []ToolDef{
		{Name: "web_search", Description: "Search the web"},
	}, ChatConfig{Model: "llama3"})

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

func TestOllamaError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "model not found"}`))
	}))
	defer server.Close()

	p := NewOllama(server.URL)
	_, err := p.ChatComplete(context.Background(), []Message{
		{Role: RoleUser, Content: "Hello"},
	}, nil, ChatConfig{Model: "nonexistent"})

	if err == nil {
		t.Fatal("expected error for server error")
	}
}

func TestOllamaName(t *testing.T) {
	p := NewOllama("")
	if p.Name() != "ollama" {
		t.Errorf("Name = %q, want %q", p.Name(), "ollama")
	}
}
