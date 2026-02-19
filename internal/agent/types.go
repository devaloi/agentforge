// Package agent implements the core agent execution loop and configuration.
package agent

import "time"

// Status represents the current state of an agent's execution.
type Status string

const (
	StatusPending  Status = "pending"
	StatusRunning  Status = "running"
	StatusComplete Status = "complete"
	StatusFailed   Status = "failed"
)

// Result holds the output of an agent execution.
type Result struct {
	Content       string        `json:"content"`
	ToolCallCount int           `json:"tool_call_count"`
	TokensUsed    int           `json:"tokens_used"`
	Duration      time.Duration `json:"duration"`
	Status        Status        `json:"status"`
	Error         string        `json:"error,omitempty"`
}
