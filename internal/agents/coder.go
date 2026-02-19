package agents

import (
	"log/slog"

	"github.com/devaloi/agentforge/internal/agent"
	"github.com/devaloi/agentforge/internal/memory"
	"github.com/devaloi/agentforge/internal/provider"
	"github.com/devaloi/agentforge/internal/tools"
)

const coderPrompt = `You are an expert software engineer. Write clean, well-documented, production-quality code. Read research findings from shared memory for context. Write generated code to files.

When given a coding task:
1. Read relevant context from shared memory using memory_read
2. Generate code using code_gen
3. Write generated code to files using write_file
4. Store code references in shared memory using memory_write
5. Return a summary of what you built`

// NewCoder creates a coder agent configured with code generation, file I/O, and memory tools.
func NewCoder(cfg agent.Config, p provider.Provider, mem *memory.Store, outputDir string, logger *slog.Logger) *agent.Agent {
	cfg.SystemPrompt = coderPrompt
	reg := tools.NewRegistry()
	reg.Register(tools.NewCodeGen(p, cfg.Model))
	reg.Register(tools.NewReadFile(outputDir))
	reg.Register(tools.NewWriteFile(outputDir))
	reg.Register(tools.NewMemoryRead(mem))
	reg.Register(tools.NewMemoryWrite(mem, cfg.Name))
	return agent.New(cfg, p, reg, logger)
}
