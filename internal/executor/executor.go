// Package executor runs a DAG of sub-tasks with parallel execution
// and dependency resolution.
package executor

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/devaloi/agentforge/internal/agent"
	"github.com/devaloi/agentforge/internal/planner"
)

// Result holds the outcome of executing an entire DAG.
type Result struct {
	TaskResults  map[string]*agent.Result `json:"task_results"`
	Duration     time.Duration            `json:"duration"`
	SuccessCount int                      `json:"success_count"`
	FailureCount int                      `json:"failure_count"`
	BlockedCount int                      `json:"blocked_count"`
}

// Executor runs a DAG of sub-tasks using an agent factory.
type Executor struct {
	factory     agent.AgentFactory
	concurrency int
	logger      *slog.Logger
}

// New creates an Executor with the given factory and concurrency limit.
func New(factory agent.AgentFactory, concurrency int, logger *slog.Logger) *Executor {
	if concurrency <= 0 {
		concurrency = 4
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Executor{
		factory:     factory,
		concurrency: concurrency,
		logger:      logger,
	}
}

// Execute runs all tasks in the DAG respecting dependencies and parallelism.
func (e *Executor) Execute(ctx context.Context, dag *planner.DAG) (*Result, error) {
	start := time.Now()

	layers, err := dag.TopologicalSort()
	if err != nil {
		return nil, fmt.Errorf("topological sort: %w", err)
	}

	result := &Result{
		TaskResults: make(map[string]*agent.Result),
	}

	for _, layer := range layers {
		if err := e.executeLayer(ctx, dag, layer, result); err != nil {
			e.logger.Warn("layer execution error", "error", err)
		}
	}

	result.Duration = time.Since(start)
	return result, nil
}

func (e *Executor) executeLayer(ctx context.Context, dag *planner.DAG, taskIDs []string, result *Result) error {
	var mu sync.Mutex
	sem := make(chan struct{}, e.concurrency)

	var wg sync.WaitGroup
	for _, id := range taskIDs {
		task, ok := dag.GetNode(id)
		if !ok {
			continue
		}

		if e.isBlocked(task, result) {
			mu.Lock()
			task.Status = planner.TaskBlocked
			result.BlockedCount++
			result.TaskResults[id] = &agent.Result{
				Status: agent.StatusFailed,
				Error:  "blocked by failed dependency",
			}
			mu.Unlock()
			continue
		}

		wg.Add(1)
		go func(taskID string, t *planner.SubTask) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			t.Status = planner.TaskRunning
			e.logger.Info("task started", "task", taskID, "agent", t.AgentType)

			agentResult, err := e.runTask(ctx, t)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				t.Status = planner.TaskFailed
				result.FailureCount++
				result.TaskResults[taskID] = &agent.Result{
					Status: agent.StatusFailed,
					Error:  err.Error(),
				}
				e.logger.Error("task failed", "task", taskID, "error", err)
			} else {
				t.Status = planner.TaskComplete
				t.Result = agentResult.Content
				result.SuccessCount++
				result.TaskResults[taskID] = agentResult
				e.logger.Info("task completed", "task", taskID, "duration", agentResult.Duration)
			}
		}(id, task)
	}

	wg.Wait()
	return nil
}

func (e *Executor) runTask(ctx context.Context, task *planner.SubTask) (*agent.Result, error) {
	a, err := e.factory(task.AgentType)
	if err != nil {
		return nil, fmt.Errorf("create agent: %w", err)
	}
	return a.Run(ctx, task.Description)
}

func (e *Executor) isBlocked(task *planner.SubTask, result *Result) bool {
	for _, dep := range task.Dependencies {
		r, ok := result.TaskResults[dep]
		if ok && r.Status == agent.StatusFailed {
			return true
		}
	}
	return false
}
