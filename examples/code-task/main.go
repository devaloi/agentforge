// Package main demonstrates a multi-agent code generation task:
// plan → research → code → review → deliver.
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

	task := "Build a REST API for a todo app in Go with CRUD endpoints"

	supervisorMock := provider.NewMockProvider(
		&provider.Response{
			Content: `{
				"tasks": [
					{"id": "research", "description": "Research Go REST API best practices", "agent": "researcher", "depends_on": []},
					{"id": "code_models", "description": "Implement todo data models", "agent": "coder", "depends_on": ["research"]},
					{"id": "code_handlers", "description": "Implement CRUD handlers", "agent": "coder", "depends_on": ["research"]},
					{"id": "review", "description": "Review generated code", "agent": "reviewer", "depends_on": ["code_models", "code_handlers"]}
				]
			}`,
			Usage: provider.Usage{TotalTokens: 60},
		},
		&provider.Response{
			Content: "The REST API has been successfully built with data models, CRUD handlers, and code review. All components follow Go best practices.",
			Usage:   provider.Usage{TotalTokens: 50},
		},
	)

	mem := memory.NewStore()

	factory := func(agentType string) (*agent.Agent, error) {
		mock := provider.NewMockProvider(&provider.Response{
			Content: fmt.Sprintf("[%s] Task completed successfully with high-quality output.", agentType),
			Usage:   provider.Usage{TotalTokens: 20},
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
		case "coder":
			return agents.NewCoder(cfg, mock, mem, "/tmp/agentforge-example", logger), nil
		case "reviewer":
			return agents.NewReviewer(cfg, mock, mem, logger), nil
		default:
			return agent.New(cfg, mock, tools.NewRegistry(), logger), nil
		}
	}

	sv := supervisor.New(supervisorMock, "gpt-4o", factory, 4, logger)

	fmt.Println("💻 Multi-Agent Code Task Example")
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
