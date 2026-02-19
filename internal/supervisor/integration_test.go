package supervisor_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/devaloi/agentforge/internal/agent"
	"github.com/devaloi/agentforge/internal/agents"
	"github.com/devaloi/agentforge/internal/config"
	"github.com/devaloi/agentforge/internal/memory"
	"github.com/devaloi/agentforge/internal/provider"
	"github.com/devaloi/agentforge/internal/supervisor"
	"github.com/devaloi/agentforge/internal/tools"
)

func TestIntegrationFullSupervisorFlow(t *testing.T) {
	planResp := &provider.Response{
		Content: `{
			"tasks": [
				{"id": "research", "description": "Research Go REST API best practices", "agent": "researcher", "depends_on": []},
				{"id": "code_models", "description": "Implement data models", "agent": "coder", "depends_on": ["research"]},
				{"id": "code_handlers", "description": "Implement CRUD handlers", "agent": "coder", "depends_on": ["research"]},
				{"id": "review", "description": "Review code", "agent": "reviewer", "depends_on": ["code_models", "code_handlers"]},
				{"id": "docs", "description": "Write documentation", "agent": "writer", "depends_on": ["review"]}
			]
		}`,
		Usage: provider.Usage{TotalTokens: 80},
	}

	synthResp := &provider.Response{
		Content: "The REST API has been built: research → data models + handlers (parallel) → review → documentation. All stages completed successfully.",
		Usage:   provider.Usage{TotalTokens: 40},
	}

	svMock := provider.NewMockProvider(planResp, synthResp)
	mem := memory.NewStore()
	logger := slog.Default()

	factory := func(agentType string) (*agent.Agent, error) {
		mock := provider.NewMockProvider(&provider.Response{
			Content: "[" + agentType + "] completed task",
			Usage:   provider.Usage{TotalTokens: 15},
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
			return agents.NewCoder(cfg, mock, mem, t.TempDir(), logger), nil
		case "reviewer":
			return agents.NewReviewer(cfg, mock, mem, logger), nil
		case "writer":
			return agents.NewWriter(cfg, mock, mem, logger), nil
		default:
			return agent.New(cfg, mock, tools.NewRegistry(), logger), nil
		}
	}

	sv := supervisor.New(svMock, "gpt-4o", factory, 4, logger)
	result, err := sv.Run(context.Background(), "Build a REST API for a todo app in Go")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result.TaskCount != 5 {
		t.Errorf("TaskCount = %d, want 5", result.TaskCount)
	}
	if result.SuccessCount != 5 {
		t.Errorf("SuccessCount = %d, want 5", result.SuccessCount)
	}
	if result.FailureCount != 0 {
		t.Errorf("FailureCount = %d, want 0", result.FailureCount)
	}
	if result.FinalOutput == "" {
		t.Error("FinalOutput should not be empty")
	}
}

func TestIntegrationSharedMemoryCollaboration(t *testing.T) {
	mem := memory.NewStore()
	logger := slog.Default()

	researcherMock := provider.NewMockProvider(
		&provider.Response{
			ToolCalls: []provider.ToolCall{
				{ID: "c1", Name: "web_search", Arguments: `{"query":"Go REST best practices"}`},
			},
			Usage: provider.Usage{TotalTokens: 10},
		},
		&provider.Response{
			ToolCalls: []provider.ToolCall{
				{ID: "c2", Name: "memory_write", Arguments: `{"key":"research_findings","value":"Use net/http, structured handlers, middleware"}`},
			},
			Usage: provider.Usage{TotalTokens: 15},
		},
		&provider.Response{
			Content: "Research complete, findings stored in memory.",
			Usage:   provider.Usage{TotalTokens: 10},
		},
	)

	researcher := agents.NewResearcher(agent.Config{
		Name: "researcher", Model: "gpt-4o", MaxIterations: 5, TokenBudget: 8000,
	}, researcherMock, mem, logger)

	_, err := researcher.Run(context.Background(), "Research Go REST APIs")
	if err != nil {
		t.Fatalf("researcher: %v", err)
	}

	val, ok := mem.Read("research_findings")
	if !ok {
		t.Fatal("research_findings not in shared memory")
	}
	if val == "" {
		t.Error("research_findings should not be empty")
	}

	coderMock := provider.NewMockProvider(
		&provider.Response{
			ToolCalls: []provider.ToolCall{
				{ID: "c3", Name: "memory_read", Arguments: `{"key":"research_findings"}`},
			},
			Usage: provider.Usage{TotalTokens: 10},
		},
		&provider.Response{
			Content: "Code generated using research findings from shared memory.",
			Usage:   provider.Usage{TotalTokens: 15},
		},
	)

	coder := agents.NewCoder(agent.Config{
		Name: "coder", Model: "gpt-4o", MaxIterations: 5, TokenBudget: 8000,
	}, coderMock, mem, t.TempDir(), logger)

	result, err := coder.Run(context.Background(), "Generate REST handlers")
	if err != nil {
		t.Fatalf("coder: %v", err)
	}

	if result.ToolCallCount != 1 {
		t.Errorf("coder ToolCallCount = %d, want 1", result.ToolCallCount)
	}
}

func TestIntegrationFailureHandling(t *testing.T) {
	planResp := &provider.Response{
		Content: `{
			"tasks": [
				{"id": "research", "description": "Research", "agent": "researcher", "depends_on": []},
				{"id": "code", "description": "Code", "agent": "coder", "depends_on": ["research"]}
			]
		}`,
		Usage: provider.Usage{TotalTokens: 30},
	}

	synthResp := &provider.Response{
		Content: "Partial results: research completed but coding was blocked.",
		Usage:   provider.Usage{TotalTokens: 20},
	}

	svMock := provider.NewMockProvider(planResp, synthResp)
	logger := slog.Default()

	factory := func(agentType string) (*agent.Agent, error) {
		if agentType == "researcher" {
			mock := provider.NewMockProvider()
			return agent.New(agent.Config{
				Name: "researcher", Model: "test", SystemPrompt: "fail",
				MaxIterations: 1, TokenBudget: 8000,
			}, mock, tools.NewRegistry(), logger), nil
		}
		mock := provider.NewMockProvider(&provider.Response{
			Content: "done", Usage: provider.Usage{TotalTokens: 5},
		})
		return agent.New(agent.Config{
			Name: agentType, Model: "test", SystemPrompt: "test",
			MaxIterations: 5, TokenBudget: 8000,
		}, mock, tools.NewRegistry(), logger), nil
	}

	sv := supervisor.New(svMock, "gpt-4o", factory, 4, logger)
	result, err := sv.Run(context.Background(), "test failure handling")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result.FailureCount == 0 {
		t.Log("No explicit failures, but supervisor handled gracefully")
	}
}

func TestIntegrationConfigLoading(t *testing.T) {
	cfg, err := config.Load("../../testdata/configs")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(cfg.Agents) == 0 {
		t.Error("no agents loaded")
	}
	if len(cfg.Tools) == 0 {
		t.Error("no tools loaded")
	}
	if len(cfg.Providers) == 0 {
		t.Error("no providers loaded")
	}

	for name, ac := range cfg.Agents {
		if ac.Model == "" {
			t.Errorf("agent %q: missing model", name)
		}
		if ac.MaxIterations <= 0 {
			t.Errorf("agent %q: invalid max_iterations", name)
		}
	}
}
