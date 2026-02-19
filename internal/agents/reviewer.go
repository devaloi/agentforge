package agents

import (
	"log/slog"

	"github.com/devaloi/agentforge/internal/agent"
	"github.com/devaloi/agentforge/internal/memory"
	"github.com/devaloi/agentforge/internal/provider"
	"github.com/devaloi/agentforge/internal/tools"
)

const reviewerPrompt = `You are a senior code reviewer. Analyze code for bugs, security issues, performance problems, and style. Be constructive and specific.

When given a review task:
1. Read the code from shared memory using memory_read
2. Analyze it using code_review
3. Return a detailed review with specific, actionable feedback`

// NewReviewer creates a reviewer agent configured with code review and memory read tools.
func NewReviewer(cfg agent.Config, p provider.Provider, mem *memory.Store, logger *slog.Logger) *agent.Agent {
	cfg.SystemPrompt = reviewerPrompt
	reg := tools.NewRegistry()
	reg.Register(tools.NewCodeReview(p, cfg.Model))
	reg.Register(tools.NewMemoryRead(mem))
	return agent.New(cfg, p, reg, logger)
}
