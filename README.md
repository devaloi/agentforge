# agentforge

A multi-agent orchestration framework in Go — a supervisor agent decomposes tasks into a dependency DAG, delegates to specialized sub-agents (researcher, coder, reviewer, writer), manages shared memory, and synthesizes results. Built from primitives, no frameworks.

## Architecture

```
User Task
    ↓
┌─ Supervisor ─────────────────────────────────────┐
│  1. Plan: decompose task into sub-tasks (DAG)    │
│  2. Execute: run DAG with parallel execution     │
│  3. Synthesize: merge results into final output  │
└──────────────────────────────────────────────────┘
    ↓                ↓                ↓
 Researcher        Coder           Reviewer
 (web search)    (code gen)      (code review)
    ↓                ↓                ↓
         Shared Memory (key-value store)
```

### Agent Execution Loop

```
System Prompt + User Message
        ↓
Call LLM (with tool schemas)
        ↓
Response has tool_call? → NO → Done
        ↓ YES
Execute tool (with timeout)
        ↓
Append result to history
        ↓
Token budget check → trim if needed
        ↓
Loop back to "Call LLM"
```

## Tech Stack

| Component | Choice |
|-----------|--------|
| Language | Go 1.26 |
| HTTP | `net/http` (raw API calls, no SDKs) |
| CLI | `cobra` |
| YAML | `gopkg.in/yaml.v3` |
| Concurrency | goroutines + semaphore channel |
| LLM Providers | OpenAI, Anthropic, Ollama |
| Testing | stdlib + mock provider + table-driven |
| Linting | golangci-lint |

## Prerequisites

- Go 1.22+
- golangci-lint (optional, for linting)
- One of: OpenAI API key, Anthropic API key, or local Ollama instance

## Installation

```bash
git clone https://github.com/devaloi/agentforge.git
cd agentforge
go build -o bin/agentforge ./cmd/agentforge/
```

## Configuration

Copy the environment file and add your API keys:

```bash
cp .env.example .env
# Edit .env with your API keys
```

Agent definitions, tool registrations, and provider settings are in `config/`:

- `config/agents.yaml` — agent types with system prompts, tools, and token budgets
- `config/tools.yaml` — tool schemas with JSON Schema parameter definitions
- `config/providers.yaml` — LLM provider endpoints and API key env var references

## Usage

### Run a task

```bash
# With configured providers
agentforge run "Build a REST API for a todo app in Go"

# Override provider and model
agentforge run --provider anthropic --model claude-sonnet-4-20250514 "Review this code"

# Custom config directory
agentforge run --config ./my-config/ "Research quantum computing"
```

### List agents and tools

```bash
agentforge agents list
agentforge tools list
```

### Validate configuration

```bash
agentforge config validate ./config/
```

### Run examples

```bash
go run ./examples/simple/              # Single researcher agent
go run ./examples/code-task/           # Multi-agent code generation
go run ./examples/research-report/     # Multi-agent research report
```

## Agent Types

| Agent | Purpose | Tools |
|-------|---------|-------|
| **Researcher** | Web search, information gathering | `web_search`, `memory_write` |
| **Coder** | Code generation, file I/O | `code_gen`, `read_file`, `write_file`, `memory_read`, `memory_write` |
| **Reviewer** | Code analysis, bug detection | `code_review`, `memory_read` |
| **Writer** | Documentation, text formatting | `text_gen`, `memory_read` |

## Tool Reference

| Tool | Parameters | Description |
|------|-----------|-------------|
| `web_search` | `query` | Search the web for information |
| `code_gen` | `language`, `task`, `context` | Generate code with LLM |
| `code_review` | `code`, `language` | Analyze code for issues |
| `text_gen` | `topic`, `style`, `context` | Generate structured text |
| `read_file` | `path` | Read file from sandboxed directory |
| `write_file` | `path`, `content` | Write file to sandboxed directory |
| `memory_read` | `key` | Read from shared memory |
| `memory_write` | `key`, `value` | Write to shared memory |

## Shared Memory

Agents collaborate through a thread-safe key-value store:

```go
// Researcher stores findings
memory.Write("research_findings", "Go REST APIs use net/http...", "researcher")

// Coder reads findings
findings, _ := memory.Read("research_findings")

// Reviewer reads generated code
code, _ := memory.Read("generated_code")
```

## Adding Custom Agents

Add a new agent definition to `config/agents.yaml`:

```yaml
agents:
  my_agent:
    model: gpt-4o
    provider: openai
    system_prompt: |
      You are a specialized agent for...
    tools: [web_search, memory_read, memory_write]
    max_iterations: 5
    token_budget: 8000
```

## Adding Custom Tools

Implement the `Tool` interface:

```go
type Tool interface {
    Name() string
    Description() string
    Schema() provider.JSONSchema
    Execute(ctx context.Context, params map[string]any) (string, error)
}
```

Register it in the tool registry:

```go
registry.Register(&MyCustomTool{})
```

## Example Output

```
🔵 Supervisor: Planning task decomposition...
   → Created 4 sub-tasks in dependency DAG

📋 Execution Plan:
   Step 1: [researcher] Research Go REST API best practices
   Step 2: [coder] Implement data models  |  [coder] Implement handlers
   Step 3: [reviewer] Review generated code
   Step 4: [writer] Write API documentation

🔍 Researcher: Completed in 4.2s (2 tool calls, 500 tokens)
💻 Coder: Completed in 8.1s (3 tool calls, 1200 tokens)
💻 Coder: Completed in 12.3s (4 tool calls, 1500 tokens)
🔎 Reviewer: Completed in 5.7s (1 tool call, 400 tokens)
📝 Writer: Completed in 6.8s (2 tool calls, 800 tokens)

🏁 Task Complete (37.1s)
   Agents: 5 | Tool calls: 12 | Tokens: 4,400
```

## Design Decisions

**Why no LangChain/CrewAI?** This IS the framework, built from primitives. Using an existing framework would defeat the purpose — the value is in understanding and implementing agent orchestration patterns from scratch.

**Why DAG execution?** Running sub-tasks sequentially is trivial. The DAG with parallel execution of independent tasks and dependency resolution demonstrates real orchestration. Topological sort ensures correctness; the semaphore controls concurrency.

**Why shared memory?** Without it, agents run in isolation. Shared memory is what enables collaboration — the researcher stores findings that the coder reads, the coder stores code that the reviewer reads. This is the difference between "multiple single agents" and a "multi-agent system."

**Why raw HTTP?** Using SDKs hides the protocol. Raw `net/http` calls to OpenAI/Anthropic/Ollama demonstrate understanding of the underlying APIs — message formats, function calling, streaming, error handling.

**Why YAML configuration?** Agents are configurable without code changes. A new agent type is a YAML block, not a new Go file. This shows production-grade thinking about configurability and extensibility.

## Testing

```bash
make test       # Run all tests
make lint       # Run linter
make build      # Build binary
make all        # Format, vet, lint, test, build
```

All tests use a mock LLM provider for deterministic, fast, CI-friendly execution. No test hits a real LLM API.

## License

MIT
