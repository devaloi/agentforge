// Package history manages per-agent conversation history with token budget tracking.
package history

import (
	"github.com/devaloi/agentforge/internal/provider"
)

// History tracks the message history for a single agent, enforcing a token budget.
type History struct {
	messages    []provider.Message
	tokenBudget int
	tokenCount  int
}

// New creates a History with the given token budget.
func New(tokenBudget int) *History {
	return &History{
		tokenBudget: tokenBudget,
	}
}

// Append adds a message to the history and returns the updated token count.
func (h *History) Append(msg provider.Message) int {
	h.messages = append(h.messages, msg)
	h.tokenCount += estimateTokens(msg.Content)
	return h.tokenCount
}

// Messages returns a copy of the current message history.
func (h *History) Messages() []provider.Message {
	result := make([]provider.Message, len(h.messages))
	copy(result, h.messages)
	return result
}

// TokenCount returns the approximate token count of the conversation.
func (h *History) TokenCount() int {
	return h.tokenCount
}

// NeedsTrim returns true if the token count exceeds 80% of the budget.
func (h *History) NeedsTrim() bool {
	return h.tokenCount > h.tokenBudget*80/100
}

// Len returns the number of messages.
func (h *History) Len() int {
	return len(h.messages)
}

// estimateTokens approximates token count as characters / 4.
func estimateTokens(content string) int {
	if len(content) == 0 {
		return 1
	}
	return len(content)/4 + 1
}
