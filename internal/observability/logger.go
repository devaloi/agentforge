package observability

import (
	"log/slog"
	"sync"
	"time"
)

// Logger provides structured logging for agent orchestration events.
type Logger struct {
	slog   *slog.Logger
	mu     sync.Mutex
	events []TraceEvent
}

// NewLogger creates a Logger wrapping the given slog.Logger.
func NewLogger(sl *slog.Logger) *Logger {
	if sl == nil {
		sl = slog.Default()
	}
	return &Logger{slog: sl}
}

// Event records a trace event and logs it.
func (l *Logger) Event(evt TraceEvent) {
	evt.Timestamp = time.Now()
	l.mu.Lock()
	l.events = append(l.events, evt)
	l.mu.Unlock()

	l.slog.Info(string(evt.Type),
		"agent", evt.Agent,
		"tool", evt.Tool,
		"task_id", evt.TaskID,
		"message", evt.Message,
		"duration", evt.Duration,
		"tokens", evt.Tokens,
	)
}

// AgentStart logs an agent starting execution.
func (l *Logger) AgentStart(agent, task string) {
	l.Event(TraceEvent{
		Type:    EventAgentStart,
		Agent:   agent,
		Message: task,
	})
}

// AgentComplete logs an agent completing execution.
func (l *Logger) AgentComplete(agent string, duration time.Duration, tokens int) {
	l.Event(TraceEvent{
		Type:     EventAgentComplete,
		Agent:    agent,
		Duration: duration,
		Tokens:   tokens,
	})
}

// ToolCall logs a tool invocation.
func (l *Logger) ToolCall(agent, tool, args string) {
	l.Event(TraceEvent{
		Type:    EventToolCall,
		Agent:   agent,
		Tool:    tool,
		Message: args,
	})
}

// ToolResult logs a tool result.
func (l *Logger) ToolResult(agent, tool string, duration time.Duration) {
	l.Event(TraceEvent{
		Type:     EventToolResult,
		Agent:    agent,
		Tool:     tool,
		Duration: duration,
	})
}

// TaskStarted logs a DAG task starting.
func (l *Logger) TaskStarted(taskID, agentType string) {
	l.Event(TraceEvent{
		Type:   EventTaskStarted,
		TaskID: taskID,
		Agent:  agentType,
	})
}

// TaskCompleted logs a DAG task completing.
func (l *Logger) TaskCompleted(taskID string, duration time.Duration) {
	l.Event(TraceEvent{
		Type:     EventTaskCompleted,
		TaskID:   taskID,
		Duration: duration,
	})
}

// TaskFailed logs a DAG task failing.
func (l *Logger) TaskFailed(taskID, reason string) {
	l.Event(TraceEvent{
		Type:    EventTaskFailed,
		TaskID:  taskID,
		Message: reason,
	})
}

// Events returns all recorded trace events.
func (l *Logger) Events() []TraceEvent {
	l.mu.Lock()
	defer l.mu.Unlock()
	result := make([]TraceEvent, len(l.events))
	copy(result, l.events)
	return result
}
