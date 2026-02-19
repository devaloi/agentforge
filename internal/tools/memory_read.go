package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/devaloi/agentforge/internal/memory"
	"github.com/devaloi/agentforge/internal/provider"
)

// MemoryRead reads a value from shared memory.
type MemoryRead struct {
	store *memory.Store
}

// NewMemoryRead creates a MemoryRead tool backed by the given store.
func NewMemoryRead(store *memory.Store) *MemoryRead {
	return &MemoryRead{store: store}
}

func (m *MemoryRead) Name() string        { return "memory_read" }
func (m *MemoryRead) Description() string { return "Read a value from shared memory by key" }

func (m *MemoryRead) Schema() provider.JSONSchema {
	return NewSchemaBuilder().
		AddString("key", "The memory key to read", true).
		Build()
}

func (m *MemoryRead) Execute(_ context.Context, params map[string]any) (string, error) {
	key, _ := params["key"].(string)
	if key == "" {
		return "", fmt.Errorf("memory_read: 'key' is required")
	}

	value, found := m.store.Read(key)
	result := map[string]any{"value": value, "found": found}
	data, _ := json.Marshal(result)
	return string(data), nil
}
