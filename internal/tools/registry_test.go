package tools

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/devaloi/agentforge/internal/provider"
)

// stubTool is a simple tool for testing.
type stubTool struct {
	name    string
	output  string
	delay   time.Duration
	failing bool
}

func (s *stubTool) Name() string        { return s.name }
func (s *stubTool) Description() string { return "stub tool for testing" }
func (s *stubTool) Schema() provider.JSONSchema {
	return NewSchemaBuilder().AddString("input", "test input", true).Build()
}

func (s *stubTool) Execute(ctx context.Context, _ map[string]any) (string, error) {
	if s.delay > 0 {
		select {
		case <-time.After(s.delay):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	if s.failing {
		return "", fmt.Errorf("stub tool failed")
	}
	return s.output, nil
}

func TestRegistryRegisterAndGet(t *testing.T) {
	r := NewRegistry()
	tool := &stubTool{name: "test_tool", output: "result"}
	r.Register(tool)

	got, ok := r.Get("test_tool")
	if !ok {
		t.Fatal("tool not found")
	}
	if got.Name() != "test_tool" {
		t.Errorf("Name = %q, want %q", got.Name(), "test_tool")
	}
}

func TestRegistryGetUnknown(t *testing.T) {
	r := NewRegistry()
	_, ok := r.Get("nonexistent")
	if ok {
		t.Error("should not find nonexistent tool")
	}
}

func TestRegistryList(t *testing.T) {
	r := NewRegistry()
	r.Register(&stubTool{name: "a", output: "x"})
	r.Register(&stubTool{name: "b", output: "y"})

	names := r.List()
	if len(names) != 2 {
		t.Errorf("List len = %d, want 2", len(names))
	}
}

func TestRegistryInvoke(t *testing.T) {
	r := NewRegistry()
	r.Register(&stubTool{name: "greet", output: "hello"})

	out, err := r.Invoke(context.Background(), "greet", nil, 5*time.Second)
	if err != nil {
		t.Fatalf("Invoke: %v", err)
	}
	if out != "hello" {
		t.Errorf("output = %q, want %q", out, "hello")
	}
}

func TestRegistryInvokeTimeout(t *testing.T) {
	r := NewRegistry()
	r.Register(&stubTool{name: "slow", output: "done", delay: 5 * time.Second})

	_, err := r.Invoke(context.Background(), "slow", nil, 50*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestRegistryInvokeUnknown(t *testing.T) {
	r := NewRegistry()
	_, err := r.Invoke(context.Background(), "unknown", nil, time.Second)
	if err == nil {
		t.Fatal("expected error for unknown tool")
	}
}

func TestRegistryToolDefs(t *testing.T) {
	r := NewRegistry()
	r.Register(&stubTool{name: "tool_a", output: "a"})
	r.Register(&stubTool{name: "tool_b", output: "b"})

	defs := r.ToolDefs()
	if len(defs) != 2 {
		t.Errorf("ToolDefs len = %d, want 2", len(defs))
	}
}

func TestDuplicateRegisterPanics(t *testing.T) {
	r := NewRegistry()
	r.Register(&stubTool{name: "dup", output: "x"})

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on duplicate register")
		}
	}()
	r.Register(&stubTool{name: "dup", output: "y"})
}
