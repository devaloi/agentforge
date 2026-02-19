package agents

import (
	"log/slog"

	"github.com/devaloi/agentforge/internal/agent"
	"github.com/devaloi/agentforge/internal/memory"
	"github.com/devaloi/agentforge/internal/provider"
	"github.com/devaloi/agentforge/internal/tools"
)

const writerPrompt = `You are a technical writer. Create clear, well-structured documentation. Read context from shared memory to understand what was built.

When given a writing task:
1. Read relevant context from shared memory using memory_read
2. Generate well-structured text using text_gen
3. Return the formatted documentation`

// NewWriter creates a writer agent configured with text generation and memory read tools.
func NewWriter(cfg agent.Config, p provider.Provider, mem *memory.Store, logger *slog.Logger) *agent.Agent {
	cfg.SystemPrompt = writerPrompt
	reg := tools.NewRegistry()
	reg.Register(tools.NewTextGen(p, cfg.Model))
	reg.Register(tools.NewMemoryRead(mem))
	return agent.New(cfg, p, reg, logger)
}
