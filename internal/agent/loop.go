package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/devaloi/agentforge/internal/history"
	"github.com/devaloi/agentforge/internal/provider"
)

// runLoop executes the core agent loop: prompt → LLM → tool call → result → repeat.
func runLoop(ctx context.Context, a *Agent) (*Result, error) {
	var totalTokens int
	var toolCallCount int

	for i := range a.config.MaxIterations {
		if a.history.NeedsTrim() {
			history.Trim(a.history)
		}

		toolDefs := a.tools.ToolDefs()
		resp, err := a.provider.ChatComplete(ctx, a.history.Messages(), toolDefs, provider.ChatConfig{
			Model: a.config.Model,
		})
		if err != nil {
			return nil, fmt.Errorf("iteration %d: LLM call failed: %w", i, err)
		}

		totalTokens += resp.Usage.TotalTokens

		if len(resp.ToolCalls) == 0 {
			return &Result{
				Content:       resp.Content,
				ToolCallCount: toolCallCount,
				TokensUsed:    totalTokens,
			}, nil
		}

		a.history.Append(provider.Message{
			Role:      provider.RoleAssistant,
			Content:   resp.Content,
			ToolCalls: resp.ToolCalls,
		})

		for _, tc := range resp.ToolCalls {
			toolCallCount++
			a.logger.Info("tool call",
				"agent", a.config.Name,
				"tool", tc.Name,
				"iteration", i,
			)

			args, err := tc.ParseArguments()
			if err != nil {
				a.history.Append(provider.Message{
					Role:       provider.RoleTool,
					Content:    fmt.Sprintf("Error parsing arguments: %v", err),
					ToolCallID: tc.ID,
				})
				continue
			}

			output, err := a.tools.Invoke(ctx, tc.Name, args, 30*time.Second)
			if err != nil {
				a.history.Append(provider.Message{
					Role:       provider.RoleTool,
					Content:    fmt.Sprintf("Error: %v", err),
					ToolCallID: tc.ID,
				})
				continue
			}

			a.history.Append(provider.Message{
				Role:       provider.RoleTool,
				Content:    output,
				ToolCallID: tc.ID,
			})
		}
	}

	return nil, fmt.Errorf("agent %q: exceeded max iterations (%d)", a.config.Name, a.config.MaxIterations)
}
