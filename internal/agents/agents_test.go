package agents

import (
	"context"
	"log/slog"
	"testing"

	"github.com/devaloi/agentforge/internal/agent"
	"github.com/devaloi/agentforge/internal/memory"
	"github.com/devaloi/agentforge/internal/provider"
)

func baseCfg(name string) agent.Config {
	return agent.Config{
		Name:          name,
		Model:         "gpt-4o",
		Provider:      "mock",
		MaxIterations: 5,
		TokenBudget:   8000,
	}
}

func TestResearcherRun(t *testing.T) {
	mock := provider.NewMockProvider(
		&provider.Response{
			ToolCalls: []provider.ToolCall{
				{ID: "c1", Name: "web_search", Arguments: `{"query":"Go best practices"}`},
			},
			Usage: provider.Usage{TotalTokens: 10},
		},
		&provider.Response{
			ToolCalls: []provider.ToolCall{
				{ID: "c2", Name: "memory_write", Arguments: `{"key":"findings","value":"Go is great"}`},
			},
			Usage: provider.Usage{TotalTokens: 10},
		},
		&provider.Response{
			Content: "Research complete. Stored findings in memory.",
			Usage:   provider.Usage{TotalTokens: 15},
		},
	)

	mem := memory.NewStore()
	a := NewResearcher(baseCfg("researcher"), mock, mem, slog.Default())

	result, err := a.Run(context.Background(), "Research Go best practices")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.ToolCallCount != 2 {
		t.Errorf("ToolCallCount = %d, want 2", result.ToolCallCount)
	}

	val, ok := mem.Read("findings")
	if !ok {
		t.Fatal("findings not in memory")
	}
	if val != "Go is great" {
		t.Errorf("findings = %q", val)
	}
}

func TestCoderRun(t *testing.T) {
	mock := provider.NewMockProvider(
		&provider.Response{
			ToolCalls: []provider.ToolCall{
				{ID: "c1", Name: "memory_read", Arguments: `{"key":"findings"}`},
			},
			Usage: provider.Usage{TotalTokens: 10},
		},
		&provider.Response{
			Content: "Code generated successfully.",
			Usage:   provider.Usage{TotalTokens: 15},
		},
	)

	mem := memory.NewStore()
	mem.Write("findings", "Go REST best practices", "researcher")
	a := NewCoder(baseCfg("coder"), mock, mem, t.TempDir(), slog.Default())

	result, err := a.Run(context.Background(), "Generate REST handlers")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.ToolCallCount != 1 {
		t.Errorf("ToolCallCount = %d, want 1", result.ToolCallCount)
	}
}

func TestReviewerRun(t *testing.T) {
	mock := provider.NewMockProvider(
		&provider.Response{
			Content: "Code looks good, approved.",
			Usage:   provider.Usage{TotalTokens: 15},
		},
	)

	mem := memory.NewStore()
	a := NewReviewer(baseCfg("reviewer"), mock, mem, slog.Default())

	result, err := a.Run(context.Background(), "Review the generated code")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Status != agent.StatusComplete {
		t.Errorf("Status = %q", result.Status)
	}
}

func TestWriterRun(t *testing.T) {
	mock := provider.NewMockProvider(
		&provider.Response{
			Content: "# API Documentation\n\nThis is the todo API.",
			Usage:   provider.Usage{TotalTokens: 20},
		},
	)

	mem := memory.NewStore()
	a := NewWriter(baseCfg("writer"), mock, mem, slog.Default())

	result, err := a.Run(context.Background(), "Write API documentation")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Content == "" {
		t.Error("expected non-empty content")
	}
}
