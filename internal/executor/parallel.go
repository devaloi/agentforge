package executor

import (
	"context"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"github.com/devaloi/agentforge/internal/agent"
	"github.com/devaloi/agentforge/internal/planner"
	"github.com/devaloi/agentforge/internal/provider"
	"github.com/devaloi/agentforge/internal/tools"
)

func TestParallelConcurrencyLimit(t *testing.T) {
	var maxConcurrent atomic.Int32
	var current atomic.Int32

	factory := func(agentType string) (*agent.Agent, error) {
		c := current.Add(1)
		for {
			old := maxConcurrent.Load()
			if c <= old || maxConcurrent.CompareAndSwap(old, c) {
				break
			}
		}
		time.Sleep(50 * time.Millisecond)
		current.Add(-1)

		mock := provider.NewMockProvider(&provider.Response{
			Content: "Done",
			Usage:   provider.Usage{TotalTokens: 5},
		})
		return agent.New(agent.Config{
			Name:          agentType,
			Model:         "test",
			SystemPrompt:  "Test",
			MaxIterations: 5,
			TokenBudget:   8000,
		}, mock, tools.NewRegistry(), slog.Default()), nil
	}

	dag := planner.NewDAG()
	for _, id := range []string{"a", "b", "c", "d", "e"} {
		dag.AddNode(planner.SubTask{ID: id, Description: "Task " + id, AgentType: "worker"})
	}

	exec := New(factory, 2, slog.Default())
	result, err := exec.Execute(context.Background(), dag)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if result.SuccessCount != 5 {
		t.Errorf("SuccessCount = %d, want 5", result.SuccessCount)
	}

	if maxConcurrent.Load() > 2 {
		t.Errorf("max concurrent = %d, want <= 2", maxConcurrent.Load())
	}
}
