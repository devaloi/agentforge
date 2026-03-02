# Changelog

All notable changes to agentforge are documented here.

## [0.2.0] - 2026-02-20

### Added
- Batch 9 depth improvements: richer tool execution tracing and token budget visibility
- Benchmark suite for DAG planner and parallel executor (`internal/bench/`)
- GitHub Actions CI with Go race detector and golangci-lint

### Changed
- Improved mock provider for deterministic, faster tests
- Supervisor planner now retries on malformed JSON from LLM

## [0.1.0] - 2026-02-18

### Added
- Supervisor agent with DAG-based task decomposition
- Four built-in agent types: Researcher, Coder, Reviewer, Writer
- Shared thread-safe memory store for agent collaboration
- Tool system with JSON Schema validation
- Multi-provider support: OpenAI, Anthropic, Ollama
- YAML-driven configuration for agents, tools, and providers
- CLI via `cobra`: `run`, `agents list`, `tools list`, `config validate`
- Three runnable examples: simple, code-task, research-report
- MIT License
