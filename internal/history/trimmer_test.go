package history

import (
	"strings"
	"testing"

	"github.com/devaloi/agentforge/internal/provider"
)

func TestTrimPreservesSystemMessage(t *testing.T) {
	h := New(100)

	h.Append(provider.Message{Role: provider.RoleSystem, Content: "System prompt"})
	h.Append(provider.Message{Role: provider.RoleUser, Content: strings.Repeat("a", 200)})
	h.Append(provider.Message{Role: provider.RoleAssistant, Content: strings.Repeat("b", 200)})
	h.Append(provider.Message{Role: provider.RoleUser, Content: strings.Repeat("c", 200)})

	Trim(h)

	msgs := h.Messages()
	if len(msgs) == 0 {
		t.Fatal("should have at least system message")
	}
	if msgs[0].Role != provider.RoleSystem {
		t.Errorf("first message should be system, got %q", msgs[0].Role)
	}
}

func TestTrimReducesTokens(t *testing.T) {
	h := New(100)

	h.Append(provider.Message{Role: provider.RoleSystem, Content: "Be helpful."})
	for range 10 {
		h.Append(provider.Message{Role: provider.RoleUser, Content: strings.Repeat("x", 40)})
	}

	before := h.Len()
	Trim(h)
	after := h.Len()

	if after >= before {
		t.Errorf("after trim: %d messages, before: %d, should be fewer", after, before)
	}
}

func TestTrimNoOpWhenUnderBudget(t *testing.T) {
	h := New(10000)

	h.Append(provider.Message{Role: provider.RoleSystem, Content: "System"})
	h.Append(provider.Message{Role: provider.RoleUser, Content: "Hello"})

	before := h.Len()
	Trim(h)
	after := h.Len()

	if after != before {
		t.Errorf("should not trim when under budget: %d -> %d", before, after)
	}
}
