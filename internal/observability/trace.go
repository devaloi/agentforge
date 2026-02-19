package observability

import (
	"fmt"
	"strings"
	"time"
)

// Trace collects events into a timeline and provides summary formatting.
type Trace struct {
	events []TraceEvent
	start  time.Time
}

// NewTrace creates a Trace from recorded events.
func NewTrace(events []TraceEvent) *Trace {
	var start time.Time
	if len(events) > 0 {
		start = events[0].Timestamp
	}
	return &Trace{events: events, start: start}
}

// Summary returns a human-readable timeline of the execution.
func (t *Trace) Summary() string {
	var sb strings.Builder

	sb.WriteString("Execution Trace:\n")
	sb.WriteString(strings.Repeat("─", 60) + "\n")

	for _, evt := range t.events {
		elapsed := evt.Timestamp.Sub(t.start).Truncate(time.Millisecond)
		icon := eventIcon(evt.Type)

		switch evt.Type {
		case EventAgentStart:
			fmt.Fprintf(&sb, "%s [%s] %s %s: %s\n", elapsed, evt.Agent, icon, evt.Type, evt.Message)
		case EventAgentComplete:
			fmt.Fprintf(&sb, "%s [%s] %s %s (%s, %d tokens)\n", elapsed, evt.Agent, icon, evt.Type, evt.Duration, evt.Tokens)
		case EventToolCall:
			fmt.Fprintf(&sb, "%s [%s] %s %s(%s)\n", elapsed, evt.Agent, icon, evt.Tool, evt.Message)
		case EventToolResult:
			fmt.Fprintf(&sb, "%s [%s] %s %s result (%s)\n", elapsed, evt.Agent, icon, evt.Tool, evt.Duration)
		case EventTaskStarted:
			fmt.Fprintf(&sb, "%s [%s] %s task %s started\n", elapsed, evt.Agent, icon, evt.TaskID)
		case EventTaskCompleted:
			fmt.Fprintf(&sb, "%s %s task %s completed (%s)\n", elapsed, icon, evt.TaskID, evt.Duration)
		case EventTaskFailed:
			fmt.Fprintf(&sb, "%s %s task %s FAILED: %s\n", elapsed, icon, evt.TaskID, evt.Message)
		default:
			fmt.Fprintf(&sb, "%s %s %s\n", elapsed, icon, evt.Message)
		}
	}

	sb.WriteString(strings.Repeat("─", 60) + "\n")
	return sb.String()
}

func eventIcon(t EventType) string {
	switch t {
	case EventAgentStart:
		return "🔵"
	case EventAgentComplete:
		return "✅"
	case EventToolCall:
		return "🔧"
	case EventToolResult:
		return "📝"
	case EventPlanCreated:
		return "📋"
	case EventTaskStarted:
		return "▶️"
	case EventTaskCompleted:
		return "✅"
	case EventTaskFailed:
		return "❌"
	default:
		return "•"
	}
}
