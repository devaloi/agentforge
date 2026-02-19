package provider

import "context"

// ToolDef describes a tool available to the LLM during a conversation.
type ToolDef struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Parameters  JSONSchema `json:"parameters"`
}

// JSONSchema is a minimal representation of a JSON Schema object.
type JSONSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]SchemaField `json:"properties,omitempty"`
	Required   []string               `json:"required,omitempty"`
}

// SchemaField describes a single property in a JSON Schema.
type SchemaField struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

// ChatConfig holds per-request configuration for a chat completion call.
type ChatConfig struct {
	Model       string  `json:"model"`
	Temperature float64 `json:"temperature,omitempty"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
}

// Provider is the interface for interacting with an LLM backend.
type Provider interface {
	// ChatComplete sends a conversation to the LLM and returns a response.
	ChatComplete(ctx context.Context, messages []Message, tools []ToolDef, cfg ChatConfig) (*Response, error)

	// Name returns the provider identifier (e.g. "openai", "anthropic").
	Name() string
}

// Registry maps provider names to their implementations.
type Registry struct {
	providers map[string]Provider
}

// NewRegistry creates an empty provider registry.
func NewRegistry() *Registry {
	return &Registry{providers: make(map[string]Provider)}
}

// Register adds a provider to the registry.
func (r *Registry) Register(p Provider) {
	r.providers[p.Name()] = p
}

// Get retrieves a provider by name.
func (r *Registry) Get(name string) (Provider, bool) {
	p, ok := r.providers[name]
	return p, ok
}
