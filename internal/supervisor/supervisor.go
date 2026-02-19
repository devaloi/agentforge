// Package supervisor orchestrates the full multi-agent flow:
// plan → delegate → execute → synthesize.
package supervisor

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/devaloi/agentforge/internal/agent"
	"github.com/devaloi/agentforge/internal/executor"
	"github.com/devaloi/agentforge/internal/provider"
)

// Result holds the complete output of a supervisor run.
type Result struct {
	FinalOutput  string                  `json:"final_output"`
	TaskResults  map[string]*agent.Result `json:"task_results"`
	TaskCount    int                     `json:"task_count"`
	Duration     time.Duration           `json:"duration"`
	SuccessCount int                     `json:"success_count"`
	FailureCount int                     `json:"failure_count"`
}

// Supervisor decomposes a task into sub-tasks and orchestrates their execution.
type Supervisor struct {
	provider    provider.Provider
	model       string
	factory     agent.AgentFactory
	concurrency int
	logger      *slog.Logger
}

// New creates a Supervisor with the given provider and agent factory.
func New(p provider.Provider, model string, factory agent.AgentFactory, concurrency int, logger *slog.Logger) *Supervisor {
	if logger == nil {
		logger = slog.Default()
	}
	return &Supervisor{
		provider:    p,
		model:       model,
		factory:     factory,
		concurrency: concurrency,
		logger:      logger,
	}
}

// Run executes the full supervisor flow: plan → execute → synthesize.
func (s *Supervisor) Run(ctx context.Context, task string) (*Result, error) {
	start := time.Now()

	s.logger.Info("supervisor: planning task decomposition", "task", truncate(task, 80))

	dag, err := Plan(ctx, task, s.provider, s.model)
	if err != nil {
		return nil, fmt.Errorf("planning: %w", err)
	}

	s.logger.Info("supervisor: execution plan created", "tasks", dag.Size())

	layers, _ := dag.TopologicalSort()
	for i, layer := range layers {
		s.logger.Info("supervisor: execution layer", "step", i+1, "tasks", layer)
	}

	exec := executor.New(s.factory, s.concurrency, s.logger)
	execResult, err := exec.Execute(ctx, dag)
	if err != nil {
		return nil, fmt.Errorf("execution: %w", err)
	}

	s.logger.Info("supervisor: synthesizing results",
		"success", execResult.SuccessCount,
		"failed", execResult.FailureCount,
	)

	finalOutput, err := Synthesize(ctx, task, execResult.TaskResults, s.provider, s.model)
	if err != nil {
		return nil, fmt.Errorf("synthesis: %w", err)
	}

	return &Result{
		FinalOutput:  finalOutput,
		TaskResults:  execResult.TaskResults,
		TaskCount:    dag.Size(),
		Duration:     time.Since(start),
		SuccessCount: execResult.SuccessCount,
		FailureCount: execResult.FailureCount,
	}, nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
