package history

import (
	"strings"
	"testing"

	"github.com/devaloi/agentforge/internal/provider"
)

func TestAppendAndMessages(t *testing.T) {
	h := New(8000)

	h.Append(provider.Message{Role: provider.RoleSystem, Content: "You are a helper."})
	h.Append(provider.Message{Role: provider.RoleUser, Content: "Hello"})
	h.Append(provider.Message{Role: provider.RoleAssistant, Content: "Hi there!"})

	msgs := h.Messages()
	if len(msgs) != 3 {
		t.Fatalf("len = %d, want 3", len(msgs))
	}
	if msgs[0].Role != provider.RoleSystem {
		t.Errorf("msgs[0].Role = %q, want system", msgs[0].Role)
	}
	if h.TokenCount() <= 0 {
		t.Error("TokenCount should be positive")
	}
}

func TestNeedsTrim(t *testing.T) {
	h := New(100)
	if h.NeedsTrim() {
		t.Error("empty history should not need trim")
	}

	h.Append(provider.Message{Role: provider.RoleUser, Content: strings.Repeat("x", 400)})
	if !h.NeedsTrim() {
		t.Error("should need trim after exceeding 80% budget")
	}
}

func TestLen(t *testing.T) {
	h := New(8000)
	if h.Len() != 0 {
		t.Errorf("Len = %d, want 0", h.Len())
	}
	h.Append(provider.Message{Role: provider.RoleUser, Content: "hello"})
	if h.Len() != 1 {
		t.Errorf("Len = %d, want 1", h.Len())
	}
}

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		input string
		min   int
	}{
		{"", 1},
		{"hello", 2},
		{strings.Repeat("x", 100), 25},
	}

	for _, tt := range tests {
		got := estimateTokens(tt.input)
		if got < tt.min {
			t.Errorf("estimateTokens(%q) = %d, want >= %d", tt.input, got, tt.min)
		}
	}
}
