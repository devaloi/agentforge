package history

import (
	"github.com/devaloi/agentforge/internal/provider"
)

// Trim removes the oldest non-system messages until the token count
// is within the budget threshold (80%).
func Trim(h *History) {
	target := h.tokenBudget * 70 / 100

	for h.tokenCount > target && len(h.messages) > 1 {
		idx := firstNonSystemIndex(h.messages)
		if idx < 0 {
			break
		}
		removed := h.messages[idx]
		h.messages = append(h.messages[:idx], h.messages[idx+1:]...)
		h.tokenCount -= estimateTokens(removed.Content)
		if h.tokenCount < 0 {
			h.tokenCount = 0
		}
	}
}

func firstNonSystemIndex(messages []provider.Message) int {
	for i, m := range messages {
		if m.Role != provider.RoleSystem {
			return i
		}
	}
	return -1
}
