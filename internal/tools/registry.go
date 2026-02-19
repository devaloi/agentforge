package tools

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/devaloi/agentforge/internal/provider"
)

// Registry manages a set of named tools and provides invocation with timeouts.
type Registry struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

// NewRegistry creates an empty tool registry.
func NewRegistry() *Registry {
	return &Registry{tools: make(map[string]Tool)}
}

// Register adds a tool to the registry. Panics on duplicate name.
func (r *Registry) Register(t Tool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.tools[t.Name()]; exists {
		panic(fmt.Sprintf("tool already registered: %s", t.Name()))
	}
	r.tools[t.Name()] = t
}

// Get returns a tool by name.
func (r *Registry) Get(name string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tools[name]
	return t, ok
}

// List returns all registered tool names.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}

// Invoke executes a named tool with a timeout.
func (r *Registry) Invoke(ctx context.Context, name string, params map[string]any, timeout time.Duration) (string, error) {
	t, ok := r.Get(name)
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", name)
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	type result struct {
		output string
		err    error
	}

	ch := make(chan result, 1)
	go func() {
		out, err := t.Execute(ctx, params)
		ch <- result{out, err}
	}()

	select {
	case <-ctx.Done():
		return "", fmt.Errorf("tool %s: %w", name, ctx.Err())
	case res := <-ch:
		return res.output, res.err
	}
}

// ToolDefs returns provider.ToolDef for all registered tools.
func (r *Registry) ToolDefs() []provider.ToolDef {
	r.mu.RLock()
	defer r.mu.RUnlock()
	defs := make([]provider.ToolDef, 0, len(r.tools))
	for _, t := range r.tools {
		defs = append(defs, ToToolDef(t))
	}
	return defs
}
