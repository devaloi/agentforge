package agent

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/devaloi/agentforge/internal/history"
	"github.com/devaloi/agentforge/internal/provider"
	"github.com/devaloi/agentforge/internal/tools"
)

// Agent represents an autonomous agent with a system prompt, tools, and memory.
type Agent struct {
	config   Config
	provider provider.Provider
	tools    *tools.Registry
	history  *history.History
	logger   *slog.Logger
}

// New creates a new Agent with the given configuration and dependencies.
func New(cfg Config, p provider.Provider, toolReg *tools.Registry, logger *slog.Logger) *Agent {
	if logger == nil {
		logger = slog.Default()
	}
	return &Agent{
		config:   cfg,
		provider: p,
		tools:    toolReg,
		history:  history.New(cfg.TokenBudget),
		logger:   logger,
	}
}

// Run executes the agent's task loop and returns the result.
func (a *Agent) Run(ctx context.Context, task string) (*Result, error) {
	start := time.Now()

	a.history.Append(provider.Message{Role: provider.RoleSystem, Content: a.config.SystemPrompt})
	a.history.Append(provider.Message{Role: provider.RoleUser, Content: task})

	a.logger.Info("agent started",
		"agent", a.config.Name,
		"task", truncate(task, 80),
	)

	result, err := runLoop(ctx, a)
	if err != nil {
		return &Result{
			Status:   StatusFailed,
			Error:    err.Error(),
			Duration: time.Since(start),
		}, err
	}

	result.Duration = time.Since(start)
	result.Status = StatusComplete

	a.logger.Info("agent completed",
		"agent", a.config.Name,
		"tool_calls", result.ToolCallCount,
		"tokens", result.TokensUsed,
		"duration", result.Duration,
	)

	return result, nil
}

// Name returns the agent's configured name.
func (a *Agent) Name() string { return a.config.Name }

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// AgentFactory creates agents by type name.
type AgentFactory func(agentType string) (*Agent, error)

// NewFactory creates an AgentFactory from a map of agent type → constructor.
func NewFactory(builders map[string]func() (*Agent, error)) AgentFactory {
	return func(agentType string) (*Agent, error) {
		build, ok := builders[agentType]
		if !ok {
			return nil, fmt.Errorf("unknown agent type: %s", agentType)
		}
		return build()
	}
}
