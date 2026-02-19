package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/devaloi/agentforge/internal/memory"
	"github.com/devaloi/agentforge/internal/provider"
)

// MemoryWrite writes a value to shared memory with agent attribution.
type MemoryWrite struct {
	store *memory.Store
	agent string
}

// NewMemoryWrite creates a MemoryWrite tool for the given agent.
func NewMemoryWrite(store *memory.Store, agentName string) *MemoryWrite {
	return &MemoryWrite{store: store, agent: agentName}
}

func (m *MemoryWrite) Name() string        { return "memory_write" }
func (m *MemoryWrite) Description() string { return "Write a key-value pair to shared memory" }

func (m *MemoryWrite) Schema() provider.JSONSchema {
	return NewSchemaBuilder().
		AddString("key", "The memory key to write", true).
		AddString("value", "The value to store", true).
		Build()
}

func (m *MemoryWrite) Execute(_ context.Context, params map[string]any) (string, error) {
	key, _ := params["key"].(string)
	value, _ := params["value"].(string)

	if key == "" || value == "" {
		return "", fmt.Errorf("memory_write: 'key' and 'value' are required")
	}

	m.store.Write(key, value, m.agent)
	result := map[string]any{"success": true}
	data, _ := json.Marshal(result)
	return string(data), nil
}
