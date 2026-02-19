package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadValidConfig(t *testing.T) {
	cfg, err := Load("../../testdata/configs")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(cfg.Agents) != 2 {
		t.Errorf("agents count = %d, want 2", len(cfg.Agents))
	}
	if len(cfg.Tools) != 4 {
		t.Errorf("tools count = %d, want 4", len(cfg.Tools))
	}
	if len(cfg.Providers) != 2 {
		t.Errorf("providers count = %d, want 2", len(cfg.Providers))
	}

	researcher := cfg.Agents["researcher"]
	if researcher.Model != "gpt-4o" {
		t.Errorf("researcher model = %q, want %q", researcher.Model, "gpt-4o")
	}
	if researcher.MaxIterations != 5 {
		t.Errorf("researcher max_iterations = %d, want 5", researcher.MaxIterations)
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for missing directory")
	}
}

func TestValidateInvalidProviderRef(t *testing.T) {
	dir := t.TempDir()

	writeYAML(t, filepath.Join(dir, "agents.yaml"), `
agents:
  test:
    model: gpt-4o
    provider: nonexistent
    system_prompt: "test"
    tools: []
    max_iterations: 5
    token_budget: 8000
`)
	writeYAML(t, filepath.Join(dir, "tools.yaml"), "tools: {}")
	writeYAML(t, filepath.Join(dir, "providers.yaml"), `
providers:
  openai:
    name: openai
    base_url: "https://api.openai.com/v1"
`)

	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected error for invalid provider reference")
	}
}

func TestValidateMissingModel(t *testing.T) {
	dir := t.TempDir()

	writeYAML(t, filepath.Join(dir, "agents.yaml"), `
agents:
  test:
    provider: openai
    system_prompt: "test"
    tools: []
    max_iterations: 5
    token_budget: 8000
`)
	writeYAML(t, filepath.Join(dir, "tools.yaml"), "tools: {}")
	writeYAML(t, filepath.Join(dir, "providers.yaml"), `
providers:
  openai:
    name: openai
    base_url: "https://api.openai.com/v1"
`)

	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected error for missing model")
	}
}

func TestValidateInvalidToolRef(t *testing.T) {
	dir := t.TempDir()

	writeYAML(t, filepath.Join(dir, "agents.yaml"), `
agents:
  test:
    model: gpt-4o
    provider: openai
    system_prompt: "test"
    tools: [nonexistent_tool]
    max_iterations: 5
    token_budget: 8000
`)
	writeYAML(t, filepath.Join(dir, "tools.yaml"), "tools: {}")
	writeYAML(t, filepath.Join(dir, "providers.yaml"), `
providers:
  openai:
    name: openai
    base_url: "https://api.openai.com/v1"
`)

	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected error for invalid tool reference")
	}
}

func writeYAML(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
