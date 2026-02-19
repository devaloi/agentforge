// Package observability provides structured logging and execution tracing
// for the multi-agent orchestration system.
package observability

import "time"

// EventType classifies trace events.
type EventType string

const (
	EventAgentStart    EventType = "agent_start"
	EventAgentComplete EventType = "agent_complete"
	EventToolCall      EventType = "tool_call"
	EventToolResult    EventType = "tool_result"
	EventPlanCreated   EventType = "plan_created"
	EventTaskStarted   EventType = "task_started"
	EventTaskCompleted EventType = "task_completed"
	EventTaskFailed    EventType = "task_failed"
)

// TraceEvent represents a single event in the execution timeline.
type TraceEvent struct {
	Type      EventType     `json:"type"`
	Agent     string        `json:"agent,omitempty"`
	Tool      string        `json:"tool,omitempty"`
	TaskID    string        `json:"task_id,omitempty"`
	Message   string        `json:"message,omitempty"`
	Duration  time.Duration `json:"duration,omitempty"`
	Tokens    int           `json:"tokens,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
}

// TraceSpan represents a logical span of execution (e.g., an entire agent run).
type TraceSpan struct {
	Name   string       `json:"name"`
	Start  time.Time    `json:"start"`
	End    time.Time    `json:"end"`
	Events []TraceEvent `json:"events"`
}
