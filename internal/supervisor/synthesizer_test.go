package supervisor

import (
	"context"
	"testing"

	"github.com/devaloi/agentforge/internal/agent"
	"github.com/devaloi/agentforge/internal/provider"
)

func TestSynthesize(t *testing.T) {
	mock := provider.NewMockProvider(&provider.Response{
		Content: "Here is the synthesized result combining research and code.",
		Usage:   provider.Usage{TotalTokens: 30},
	})

	results := map[string]*agent.Result{
		"research": {Content: "Go REST APIs use net/http", Status: agent.StatusComplete},
		"code":     {Content: "func main() {}", Status: agent.StatusComplete},
	}

	output, err := Synthesize(context.Background(), "Build a REST API", results, mock, "gpt-4o")
	if err != nil {
		t.Fatalf("Synthesize: %v", err)
	}

	if output == "" {
		t.Error("output should not be empty")
	}

	calls := mock.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 LLM call, got %d", len(calls))
	}
}

func TestSynthesizeWithFailures(t *testing.T) {
	mock := provider.NewMockProvider(&provider.Response{
		Content: "Partial results synthesized despite some failures.",
		Usage:   provider.Usage{TotalTokens: 20},
	})

	results := map[string]*agent.Result{
		"research": {Content: "Found info", Status: agent.StatusComplete},
		"code":     {Status: agent.StatusFailed, Error: "LLM error"},
	}

	output, err := Synthesize(context.Background(), "test", results, mock, "gpt-4o")
	if err != nil {
		t.Fatalf("Synthesize: %v", err)
	}

	if output == "" {
		t.Error("should produce output even with failures")
	}
}
