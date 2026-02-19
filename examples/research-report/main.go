// Package main demonstrates a multi-agent research report:
// research → write → synthesize.
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
	"github.com/devaloi/agentforge/internal/supervisor"
	"github.com/devaloi/agentforge/internal/tools"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	task := "Write a research report on the state of WebAssembly in 2025"

	supervisorMock := provider.NewMockProvider(
		&provider.Response{
			Content: `{
				"tasks": [
					{"id": "research_overview", "description": "Research WebAssembly current state and adoption", "agent": "researcher", "depends_on": []},
					{"id": "research_performance", "description": "Research WebAssembly performance benchmarks", "agent": "researcher", "depends_on": []},
					{"id": "write_report", "description": "Write comprehensive research report", "agent": "writer", "depends_on": ["research_overview", "research_performance"]}
				]
			}`,
			Usage: provider.Usage{TotalTokens: 50},
		},
		&provider.Response{
			Content: "# WebAssembly in 2025: A Research Report\n\nWebAssembly has matured significantly, with broad adoption across browsers and server-side runtimes. Performance benchmarks show near-native execution speed. The ecosystem includes WASI for system interfaces and component model for composition.",
			Usage: provider.Usage{TotalTokens: 60},
		},
	)

	mem := memory.NewStore()

	factory := func(agentType string) (*agent.Agent, error) {
		mock := provider.NewMockProvider(&provider.Response{
			Content: fmt.Sprintf("[%s] Research and analysis completed.", agentType),
			Usage:   provider.Usage{TotalTokens: 25},
		})
		cfg := agent.Config{
			Name:          agentType,
			Model:         "gpt-4o",
			MaxIterations: 5,
			TokenBudget:   8000,
		}
		switch agentType {
		case "researcher":
			return agents.NewResearcher(cfg, mock, mem, logger), nil
		case "writer":
			return agents.NewWriter(cfg, mock, mem, logger), nil
		default:
			return agent.New(cfg, mock, tools.NewRegistry(), logger), nil
		}
	}

	sv := supervisor.New(supervisorMock, "gpt-4o", factory, 4, logger)

	fmt.Println("📝 Multi-Agent Research Report Example")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Task: %s\n\n", task)

	result, err := sv.Run(context.Background(), task)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n🏁 Task Complete (%s)\n", result.Duration.Truncate(1e6))
	fmt.Printf("   Tasks: %d | Success: %d | Failed: %d\n\n",
		result.TaskCount, result.SuccessCount, result.FailureCount)
	fmt.Printf("📄 Final Output:\n%s\n", result.FinalOutput)
}
