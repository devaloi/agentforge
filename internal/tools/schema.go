package tools

import (
	"github.com/devaloi/agentforge/internal/provider"
)

// SchemaBuilder constructs JSON Schema definitions for tool parameters.
type SchemaBuilder struct {
	properties map[string]provider.SchemaField
	required   []string
}

// NewSchemaBuilder creates a new SchemaBuilder.
func NewSchemaBuilder() *SchemaBuilder {
	return &SchemaBuilder{
		properties: make(map[string]provider.SchemaField),
	}
}

// AddString adds a string property to the schema.
func (b *SchemaBuilder) AddString(name, description string, required bool) *SchemaBuilder {
	b.properties[name] = provider.SchemaField{Type: "string", Description: description}
	if required {
		b.required = append(b.required, name)
	}
	return b
}

// AddBool adds a boolean property to the schema.
func (b *SchemaBuilder) AddBool(name, description string, required bool) *SchemaBuilder {
	b.properties[name] = provider.SchemaField{Type: "boolean", Description: description}
	if required {
		b.required = append(b.required, name)
	}
	return b
}

// Build produces the final JSONSchema.
func (b *SchemaBuilder) Build() provider.JSONSchema {
	return provider.JSONSchema{
		Type:       "object",
		Properties: b.properties,
		Required:   b.required,
	}
}
