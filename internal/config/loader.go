package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Load reads YAML configuration files from a directory.
// It expects agents.yaml, tools.yaml, and providers.yaml.
func Load(dir string) (*Config, error) {
	cfg := &Config{}

	agents, err := loadAgents(filepath.Join(dir, "agents.yaml"))
	if err != nil {
		return nil, fmt.Errorf("load agents: %w", err)
	}
	cfg.Agents = agents

	tools, err := loadTools(filepath.Join(dir, "tools.yaml"))
	if err != nil {
		return nil, fmt.Errorf("load tools: %w", err)
	}
	cfg.Tools = tools

	providers, err := loadProviders(filepath.Join(dir, "providers.yaml"))
	if err != nil {
		return nil, fmt.Errorf("load providers: %w", err)
	}
	cfg.Providers = providers

	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}

func loadAgents(path string) (map[string]AgentConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var wrapper struct {
		Agents map[string]AgentConfig `yaml:"agents"`
	}
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return nil, err
	}
	return wrapper.Agents, nil
}

func loadTools(path string) (map[string]ToolConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var wrapper struct {
		Tools map[string]ToolConfig `yaml:"tools"`
	}
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return nil, err
	}
	return wrapper.Tools, nil
}

func loadProviders(path string) (map[string]ProviderConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var wrapper struct {
		Providers map[string]ProviderConfig `yaml:"providers"`
	}
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return nil, err
	}
	return wrapper.Providers, nil
}

func validate(cfg *Config) error {
	for name, agent := range cfg.Agents {
		if agent.Model == "" {
			return fmt.Errorf("agent %q: missing model", name)
		}
		if agent.Provider == "" {
			return fmt.Errorf("agent %q: missing provider", name)
		}
		if _, ok := cfg.Providers[agent.Provider]; !ok {
			return fmt.Errorf("agent %q: references unknown provider %q", name, agent.Provider)
		}
		for _, toolName := range agent.Tools {
			if _, ok := cfg.Tools[toolName]; !ok {
				return fmt.Errorf("agent %q: references unknown tool %q", name, toolName)
			}
		}
		if agent.MaxIterations <= 0 {
			return fmt.Errorf("agent %q: max_iterations must be positive", name)
		}
		if agent.TokenBudget <= 0 {
			return fmt.Errorf("agent %q: token_budget must be positive", name)
		}
	}

	for name, p := range cfg.Providers {
		if p.BaseURL == "" {
			return fmt.Errorf("provider %q: missing base_url", name)
		}
	}

	return nil
}

// Validate checks configuration files in a directory for errors.
func Validate(dir string) error {
	_, err := Load(dir)
	return err
}
