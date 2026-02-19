package agent

import (
	"context"
	"log/slog"
	"testing"

	"github.com/devaloi/agentforge/internal/provider"
	"github.com/devaloi/agentforge/internal/tools"
)

func newTestAgent(responses ...*provider.Response) *Agent {
	mock := provider.NewMockProvider(responses...)
	reg := tools.NewRegistry()
	reg.Register(&tools.WebSearch{})

	return New(Config{
		Name:          "test-agent",
		Model:         "gpt-4o",
		SystemPrompt:  "You are a helpful assistant.",
		MaxIterations: 5,
		TokenBudget:   8000,
	}, mock, reg, slog.Default())
}

func TestAgentRunNoToolCalls(t *testing.T) {
	agent := newTestAgent(&provider.Response{
		Content: "The answer is 42.",
		Usage:   provider.Usage{TotalTokens: 20},
	})

	result, err := agent.Run(context.Background(), "What is the meaning of life?")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result.Content != "The answer is 42." {
		t.Errorf("Content = %q", result.Content)
	}
	if result.ToolCallCount != 0 {
		t.Errorf("ToolCallCount = %d, want 0", result.ToolCallCount)
	}
	if result.Status != StatusComplete {
		t.Errorf("Status = %q, want %q", result.Status, StatusComplete)
	}
}

func TestAgentRunWithToolCall(t *testing.T) {
	agent := newTestAgent(
		&provider.Response{
			ToolCalls: []provider.ToolCall{
				{ID: "call_1", Name: "web_search", Arguments: `{"query":"Go REST API"}`},
			},
			Usage: provider.Usage{TotalTokens: 15},
		},
		&provider.Response{
			Content: "Here are the search results summarized.",
			Usage:   provider.Usage{TotalTokens: 25},
		},
	)

	result, err := agent.Run(context.Background(), "Search for Go REST APIs")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result.ToolCallCount != 1 {
		t.Errorf("ToolCallCount = %d, want 1", result.ToolCallCount)
	}
	if result.TokensUsed != 40 {
		t.Errorf("TokensUsed = %d, want 40", result.TokensUsed)
	}
}

func TestAgentRunMaxIterations(t *testing.T) {
	// Every response triggers a tool call, never completing
	responses := make([]*provider.Response, 10)
	for i := range responses {
		responses[i] = &provider.Response{
			ToolCalls: []provider.ToolCall{
				{ID: "call_loop", Name: "web_search", Arguments: `{"query":"infinite"}`},
			},
			Usage: provider.Usage{TotalTokens: 5},
		}
	}

	agent := newTestAgent(responses...)
	agent.config.MaxIterations = 3

	_, err := agent.Run(context.Background(), "Loop forever")
	if err == nil {
		t.Fatal("expected max iterations error")
	}
}

func TestAgentRunInvalidToolArgs(t *testing.T) {
	agent := newTestAgent(
		&provider.Response{
			ToolCalls: []provider.ToolCall{
				{ID: "call_bad", Name: "web_search", Arguments: `{invalid}`},
			},
			Usage: provider.Usage{TotalTokens: 10},
		},
		&provider.Response{
			Content: "I could not parse the tool arguments.",
			Usage:   provider.Usage{TotalTokens: 10},
		},
	)

	result, err := agent.Run(context.Background(), "Test invalid args")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Content == "" {
		t.Error("should have a content response after error recovery")
	}
}
