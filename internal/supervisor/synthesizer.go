package supervisor

import (
	"context"
	"fmt"
	"strings"

	"github.com/devaloi/agentforge/internal/agent"
	"github.com/devaloi/agentforge/internal/provider"
)

const synthesisPrompt = `You are a result synthesizer. Given the original task and the results from multiple specialized agents, produce a comprehensive final output that integrates all results coherently.

Original task: %s

Sub-agent results:
%s

Synthesize these results into a single, well-structured response that addresses the original task completely.`

// Synthesize merges sub-agent results into a final coherent output.
func Synthesize(ctx context.Context, task string, results map[string]*agent.Result, p provider.Provider, model string) (string, error) {
	var parts []string
	for name, r := range results {
		if r == nil || r.Status == agent.StatusFailed {
			parts = append(parts, fmt.Sprintf("[%s]: FAILED - %s", name, r.Error))
			continue
		}
		parts = append(parts, fmt.Sprintf("[%s]: %s", name, r.Content))
	}

	prompt := fmt.Sprintf(synthesisPrompt, task, strings.Join(parts, "\n\n"))

	resp, err := p.ChatComplete(ctx, []provider.Message{
		{Role: provider.RoleUser, Content: prompt},
	}, nil, provider.ChatConfig{Model: model})
	if err != nil {
		return "", fmt.Errorf("synthesis LLM call: %w", err)
	}

	return resp.Content, nil
}
