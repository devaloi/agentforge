// Package main demonstrates a single researcher agent answering a question.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/devaloi/agentforge/internal/agent"
	"github.com/devaloi/agentforge/internal/agents"
	"github.com/devaloi/agentforge/internal/memory"
	"github.com/devaloi/agentforge/internal/provider"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	mock := provider.NewMockProvider(
		&provider.Response{
			ToolCalls: []provider.ToolCall{
				{ID: "call_1", Name: "web_search", Arguments: `{"query":"benefits of Go for microservices"}`},
			},
			Usage: provider.Usage{TotalTokens: 15},
		},
		&provider.Response{
			ToolCalls: []provider.ToolCall{
				{ID: "call_2", Name: "memory_write", Arguments: `{"key":"research_findings","value":"Go excels at microservices due to fast compilation, lightweight goroutines, and excellent standard library for HTTP servers."}`},
			},
			Usage: provider.Usage{TotalTokens: 30},
		},
		&provider.Response{
			Content: "Based on my research, Go is an excellent choice for microservices:\n\n1. **Fast compilation** — Go compiles to a single binary in seconds\n2. **Goroutines** — lightweight concurrency primitives perfect for handling many requests\n3. **Standard library** — net/http provides a production-grade HTTP server\n4. **Small footprint** — Go binaries are small and start fast, ideal for containers",
			Usage: provider.Usage{TotalTokens: 80},
		},
	)

	mem := memory.NewStore()

	cfg := agent.Config{
		Name:          "researcher",
		Model:         "gpt-4o",
		Provider:      "mock",
		MaxIterations: 5,
		TokenBudget:   8000,
	}

	researcher := agents.NewResearcher(cfg, mock, mem, logger)

	fmt.Println("🔍 Single Agent Example: Researcher")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("Question: What are the benefits of Go for microservices?")
	fmt.Println()

	result, err := researcher.Run(context.Background(), "What are the benefits of Go for microservices?")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📝 Answer:\n%s\n\n", result.Content)
	fmt.Printf("📊 Stats: %d tool calls, %d tokens, %s\n",
		result.ToolCallCount, result.TokensUsed, result.Duration)

	if val, ok := mem.Read("research_findings"); ok {
		fmt.Printf("\n💾 Shared Memory [research_findings]:\n%s\n", val)
	}
}
