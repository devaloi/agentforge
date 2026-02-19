package executor

import (
	"context"
	"log/slog"
	"sync/atomic"
	"testing"

	"github.com/devaloi/agentforge/internal/agent"
	"github.com/devaloi/agentforge/internal/planner"
	"github.com/devaloi/agentforge/internal/provider"
	"github.com/devaloi/agentforge/internal/tools"
)

func mockFactory(responses map[string][]*provider.Response) agent.AgentFactory {
	return func(agentType string) (*agent.Agent, error) {
		resps := responses[agentType]
		if resps == nil {
			resps = []*provider.Response{
				{Content: "Default response for " + agentType, Usage: provider.Usage{TotalTokens: 10}},
			}
		}
		mock := provider.NewMockProvider(resps...)
		reg := tools.NewRegistry()
		return agent.New(agent.Config{
			Name:          agentType,
			Model:         "test",
			SystemPrompt:  "You are a " + agentType,
			MaxIterations: 5,
			TokenBudget:   8000,
		}, mock, reg, slog.Default()), nil
	}
}

func TestExecuteSequential(t *testing.T) {
	dag := planner.NewDAG()
	dag.AddNode(planner.SubTask{ID: "a", Description: "Step A", AgentType: "researcher"})
	dag.AddNode(planner.SubTask{ID: "b", Description: "Step B", AgentType: "coder", Dependencies: []string{"a"}})
	dag.AddEdge("a", "b")

	exec := New(mockFactory(nil), 4, slog.Default())
	result, err := exec.Execute(context.Background(), dag)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if result.SuccessCount != 2 {
		t.Errorf("SuccessCount = %d, want 2", result.SuccessCount)
	}
	if result.FailureCount != 0 {
		t.Errorf("FailureCount = %d, want 0", result.FailureCount)
	}
}

func TestExecuteParallel(t *testing.T) {
	dag := planner.NewDAG()
	dag.AddNode(planner.SubTask{ID: "a", Description: "Task A", AgentType: "researcher"})
	dag.AddNode(planner.SubTask{ID: "b", Description: "Task B", AgentType: "coder"})
	dag.AddNode(planner.SubTask{ID: "c", Description: "Task C", AgentType: "writer"})

	var counter atomic.Int32
	factory := func(agentType string) (*agent.Agent, error) {
		counter.Add(1)
		mock := provider.NewMockProvider(&provider.Response{
			Content: "Done: " + agentType,
			Usage:   provider.Usage{TotalTokens: 10},
		})
		return agent.New(agent.Config{
			Name:          agentType,
			Model:         "test",
			SystemPrompt:  "Test",
			MaxIterations: 5,
			TokenBudget:   8000,
		}, mock, tools.NewRegistry(), slog.Default()), nil
	}

	exec := New(factory, 4, slog.Default())
	result, err := exec.Execute(context.Background(), dag)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if result.SuccessCount != 3 {
		t.Errorf("SuccessCount = %d, want 3", result.SuccessCount)
	}
	if counter.Load() != 3 {
		t.Errorf("agents created = %d, want 3", counter.Load())
	}
}

func TestExecuteFailurePropagation(t *testing.T) {
	dag := planner.NewDAG()
	dag.AddNode(planner.SubTask{ID: "a", Description: "Fail", AgentType: "failing"})
	dag.AddNode(planner.SubTask{ID: "b", Description: "Depends on fail", AgentType: "coder", Dependencies: []string{"a"}})
	dag.AddEdge("a", "b")

	factory := func(agentType string) (*agent.Agent, error) {
		if agentType == "failing" {
			mock := provider.NewMockProvider()
			return agent.New(agent.Config{
				Name:          "failing",
				Model:         "test",
				SystemPrompt:  "Fail",
				MaxIterations: 1,
				TokenBudget:   8000,
			}, mock, tools.NewRegistry(), slog.Default()), nil
		}
		return mockFactory(nil)(agentType)
	}

	exec := New(factory, 4, slog.Default())
	result, err := exec.Execute(context.Background(), dag)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if result.FailureCount != 1 {
		t.Errorf("FailureCount = %d, want 1", result.FailureCount)
	}
	if result.BlockedCount != 1 {
		t.Errorf("BlockedCount = %d, want 1", result.BlockedCount)
	}
}
