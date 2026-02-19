package supervisor

import (
	"context"
	"log/slog"
	"testing"

	"github.com/devaloi/agentforge/internal/agent"
	"github.com/devaloi/agentforge/internal/provider"
	"github.com/devaloi/agentforge/internal/tools"
)

func TestSupervisorFullFlow(t *testing.T) {
	planResp := &provider.Response{
		Content: `{
			"tasks": [
				{"id": "research", "description": "Research Go REST APIs", "agent": "researcher", "depends_on": []},
				{"id": "code", "description": "Write handlers", "agent": "coder", "depends_on": ["research"]},
				{"id": "review", "description": "Review code", "agent": "reviewer", "depends_on": ["code"]}
			]
		}`,
		Usage: provider.Usage{TotalTokens: 50},
	}

	// Sub-agent responses
	agentResp := &provider.Response{
		Content: "Task completed successfully.",
		Usage:   provider.Usage{TotalTokens: 20},
	}

	synthResp := &provider.Response{
		Content: "Final synthesized output: The REST API has been built with research, code, and review.",
		Usage:   provider.Usage{TotalTokens: 40},
	}

	// The supervisor's provider handles planning + synthesis
	supervisorMock := provider.NewMockProvider(planResp, synthResp)

	factory := func(agentType string) (*agent.Agent, error) {
		mock := provider.NewMockProvider(agentResp)
		return agent.New(agent.Config{
			Name:          agentType,
			Model:         "test",
			SystemPrompt:  "You are a " + agentType,
			MaxIterations: 5,
			TokenBudget:   8000,
		}, mock, tools.NewRegistry(), slog.Default()), nil
	}

	sv := New(supervisorMock, "gpt-4o", factory, 4, slog.Default())
	result, err := sv.Run(context.Background(), "Build a REST API for a todo app")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result.TaskCount != 3 {
		t.Errorf("TaskCount = %d, want 3", result.TaskCount)
	}
	if result.SuccessCount != 3 {
		t.Errorf("SuccessCount = %d, want 3", result.SuccessCount)
	}
	if result.FinalOutput == "" {
		t.Error("FinalOutput should not be empty")
	}
	if result.Duration <= 0 {
		t.Error("Duration should be positive")
	}
}

func TestSupervisorPlanFailure(t *testing.T) {
	mock := provider.NewMockProvider(&provider.Response{
		Content: "not valid json",
	})

	sv := New(mock, "gpt-4o", func(string) (*agent.Agent, error) {
		return nil, nil
	}, 4, slog.Default())

	_, err := sv.Run(context.Background(), "test")
	if err == nil {
		t.Fatal("expected planning error")
	}
}
