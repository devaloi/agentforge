package observability

import (
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestLoggerEvents(t *testing.T) {
	logger := NewLogger(slog.Default())

	logger.AgentStart("researcher", "Research Go REST APIs")
	logger.ToolCall("researcher", "web_search", `{"query":"Go REST"}`)
	logger.ToolResult("researcher", "web_search", 500*time.Millisecond)
	logger.AgentComplete("researcher", 2*time.Second, 500)

	events := logger.Events()
	if len(events) != 4 {
		t.Fatalf("events count = %d, want 4", len(events))
	}

	if events[0].Type != EventAgentStart {
		t.Errorf("event[0].Type = %q, want %q", events[0].Type, EventAgentStart)
	}
	if events[0].Agent != "researcher" {
		t.Errorf("event[0].Agent = %q, want %q", events[0].Agent, "researcher")
	}
	if events[1].Type != EventToolCall {
		t.Errorf("event[1].Type = %q, want %q", events[1].Type, EventToolCall)
	}
}

func TestLoggerTaskEvents(t *testing.T) {
	logger := NewLogger(slog.Default())

	logger.TaskStarted("research", "researcher")
	logger.TaskCompleted("research", time.Second)
	logger.TaskFailed("code", "LLM error")

	events := logger.Events()
	if len(events) != 3 {
		t.Fatalf("events count = %d, want 3", len(events))
	}

	if events[2].Type != EventTaskFailed {
		t.Errorf("event[2].Type = %q, want %q", events[2].Type, EventTaskFailed)
	}
}

func TestTraceSummary(t *testing.T) {
	now := time.Now()
	events := []TraceEvent{
		{Type: EventAgentStart, Agent: "researcher", Message: "Research Go", Timestamp: now},
		{Type: EventToolCall, Agent: "researcher", Tool: "web_search", Message: "Go REST", Timestamp: now.Add(100 * time.Millisecond)},
		{Type: EventAgentComplete, Agent: "researcher", Duration: time.Second, Tokens: 200, Timestamp: now.Add(time.Second)},
	}

	trace := NewTrace(events)
	summary := trace.Summary()

	if !strings.Contains(summary, "researcher") {
		t.Error("summary should contain agent name")
	}
	if !strings.Contains(summary, "web_search") {
		t.Error("summary should contain tool name")
	}
}

func TestLoggerConcurrent(t *testing.T) {
	logger := NewLogger(slog.Default())

	done := make(chan struct{})
	for range 10 {
		go func() {
			logger.AgentStart("agent", "task")
			logger.AgentComplete("agent", time.Millisecond, 10)
			done <- struct{}{}
		}()
	}

	for range 10 {
		<-done
	}

	events := logger.Events()
	if len(events) != 20 {
		t.Errorf("events count = %d, want 20", len(events))
	}
}
