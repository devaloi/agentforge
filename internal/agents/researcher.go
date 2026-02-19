// Package agents provides constructors for specialized agent types
// (researcher, coder, reviewer, writer) with pre-configured tools and prompts.
package agents

import (
	"log/slog"

	"github.com/devaloi/agentforge/internal/agent"
	"github.com/devaloi/agentforge/internal/memory"
	"github.com/devaloi/agentforge/internal/provider"
	"github.com/devaloi/agentforge/internal/tools"
)

const researcherPrompt = `You are a research assistant. Your job is to find accurate, relevant information using web search. Summarize findings clearly and store them in shared memory for other agents to use.

When given a research task:
1. Search for relevant information using web_search
2. Analyze and summarize the results
3. Store key findings in shared memory using memory_write
4. Return a clear summary of what you found`

// NewResearcher creates a researcher agent configured with web search and memory tools.
func NewResearcher(cfg agent.Config, p provider.Provider, mem *memory.Store, logger *slog.Logger) *agent.Agent {
	cfg.SystemPrompt = researcherPrompt
	reg := tools.NewRegistry()
	reg.Register(&tools.WebSearch{})
	reg.Register(tools.NewMemoryWrite(mem, cfg.Name))
	return agent.New(cfg, p, reg, logger)
}
