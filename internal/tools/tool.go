// Package tools provides the tool interface, registry, and built-in tools
// that agents use to interact with the world.
package tools

import (
	"context"

	"github.com/devaloi/agentforge/internal/provider"
)

// Tool defines the interface for an executable agent tool.
type Tool interface {
	// Name returns the tool's unique identifier.
	Name() string

	// Description returns a human-readable description for the LLM.
	Description() string

	// Schema returns the JSON Schema for the tool's parameters.
	Schema() provider.JSONSchema

	// Execute runs the tool with the given parameters.
	Execute(ctx context.Context, params map[string]any) (string, error)
}

// ToToolDef converts a Tool to a provider.ToolDef for sending to the LLM.
func ToToolDef(t Tool) provider.ToolDef {
	return provider.ToolDef{
		Name:        t.Name(),
		Description: t.Description(),
		Parameters:  t.Schema(),
	}
}
