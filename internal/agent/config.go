package agent

// Config defines the configuration for an agent instance.
type Config struct {
	Name          string   `yaml:"name"          json:"name"`
	Model         string   `yaml:"model"         json:"model"`
	Provider      string   `yaml:"provider"      json:"provider"`
	SystemPrompt  string   `yaml:"system_prompt" json:"system_prompt"`
	Tools         []string `yaml:"tools"         json:"tools"`
	MaxIterations int      `yaml:"max_iterations" json:"max_iterations"`
	TokenBudget   int      `yaml:"token_budget"  json:"token_budget"`
}
