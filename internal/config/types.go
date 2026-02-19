// Package config loads and validates YAML configuration files
// for agents, tools, and LLM providers.
package config

// Config holds the complete application configuration.
type Config struct {
	Agents    map[string]AgentConfig   `yaml:"agents"`
	Tools     map[string]ToolConfig    `yaml:"tools"`
	Providers map[string]ProviderConfig `yaml:"providers"`
}

// AgentConfig defines a single agent's configuration.
type AgentConfig struct {
	Name          string   `yaml:"name"`
	Model         string   `yaml:"model"`
	Provider      string   `yaml:"provider"`
	SystemPrompt  string   `yaml:"system_prompt"`
	Tools         []string `yaml:"tools"`
	MaxIterations int      `yaml:"max_iterations"`
	TokenBudget   int      `yaml:"token_budget"`
}

// ToolConfig defines a tool registration.
type ToolConfig struct {
	Name        string                `yaml:"name"`
	Description string                `yaml:"description"`
	Timeout     string                `yaml:"timeout"`
	Parameters  map[string]ParamField `yaml:"parameters"`
}

// ParamField describes a single parameter in a tool's schema.
type ParamField struct {
	Type        string `yaml:"type"`
	Description string `yaml:"description"`
	Required    bool   `yaml:"required"`
}

// ProviderConfig defines an LLM provider's connection settings.
type ProviderConfig struct {
	Name    string `yaml:"name"`
	BaseURL string `yaml:"base_url"`
	APIKey  string `yaml:"api_key_env"` // environment variable name, not the key itself
	Model   string `yaml:"default_model"`
}
