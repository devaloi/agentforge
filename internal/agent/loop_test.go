package agent

import (
	"context"
	"log/slog"
	"testing"

	"github.com/devaloi/agentforge/internal/provider"
	"github.com/devaloi/agentforge/internal/tools"
)

func TestLoopSingleIteration(t *testing.T) {
	mock := provider.NewMockProvider(&provider.Response{
		Content: "Direct answer",
		Usage:   provider.Usage{TotalTokens: 10},
	})
	reg := tools.NewRegistry()
	a := New(Config{
		Name:          "loop-test",
		Model:         "test",
		SystemPrompt:  "You help.",
		MaxIterations: 5,
		TokenBudget:   8000,
	}, mock, reg, slog.Default())

	a.history.Append(provider.Message{Role: provider.RoleSystem, Content: "System"})
	a.history.Append(provider.Message{Role: provider.RoleUser, Content: "Question"})

	result, err := runLoop(context.Background(), a)
	if err != nil {
		t.Fatalf("runLoop: %v", err)
	}
	if result.Content != "Direct answer" {
		t.Errorf("Content = %q", result.Content)
	}
}

func TestLoopMultiToolChain(t *testing.T) {
	mock := provider.NewMockProvider(
		&provider.Response{
			ToolCalls: []provider.ToolCall{
				{ID: "c1", Name: "web_search", Arguments: `{"query":"step 1"}`},
			},
			Usage: provider.Usage{TotalTokens: 5},
		},
		&provider.Response{
			ToolCalls: []provider.ToolCall{
				{ID: "c2", Name: "web_search", Arguments: `{"query":"step 2"}`},
			},
			Usage: provider.Usage{TotalTokens: 5},
		},
		&provider.Response{
			Content: "Done after 2 tool calls",
			Usage:   provider.Usage{TotalTokens: 10},
		},
	)

	reg := tools.NewRegistry()
	reg.Register(&tools.WebSearch{})
	a := New(Config{
		Name:          "chain-test",
		Model:         "test",
		SystemPrompt:  "You help.",
		MaxIterations: 10,
		TokenBudget:   8000,
	}, mock, reg, slog.Default())

	a.history.Append(provider.Message{Role: provider.RoleSystem, Content: "System"})
	a.history.Append(provider.Message{Role: provider.RoleUser, Content: "Multi-step"})

	result, err := runLoop(context.Background(), a)
	if err != nil {
		t.Fatalf("runLoop: %v", err)
	}
	if result.ToolCallCount != 2 {
		t.Errorf("ToolCallCount = %d, want 2", result.ToolCallCount)
	}
}
