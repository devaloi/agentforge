package supervisor

import (
	"context"
	"testing"

	"github.com/devaloi/agentforge/internal/provider"
)

func TestPlanSimple(t *testing.T) {
	mock := provider.NewMockProvider(&provider.Response{
		Content: `{
			"tasks": [
				{"id": "research", "description": "Research REST APIs", "agent": "researcher", "depends_on": []},
				{"id": "code", "description": "Write code", "agent": "coder", "depends_on": ["research"]},
				{"id": "review", "description": "Review code", "agent": "reviewer", "depends_on": ["code"]}
			]
		}`,
		Usage: provider.Usage{TotalTokens: 50},
	})

	dag, err := Plan(context.Background(), "Build a REST API", mock, "gpt-4o")
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if dag.Size() != 3 {
		t.Errorf("DAG size = %d, want 3", dag.Size())
	}

	layers, _ := dag.TopologicalSort()
	if len(layers) != 3 {
		t.Errorf("layers = %d, want 3", len(layers))
	}
}

func TestPlanComplex(t *testing.T) {
	mock := provider.NewMockProvider(&provider.Response{
		Content: `{
			"tasks": [
				{"id": "research", "description": "Research", "agent": "researcher", "depends_on": []},
				{"id": "code_models", "description": "Models", "agent": "coder", "depends_on": ["research"]},
				{"id": "code_handlers", "description": "Handlers", "agent": "coder", "depends_on": ["research"]},
				{"id": "review", "description": "Review", "agent": "reviewer", "depends_on": ["code_models", "code_handlers"]},
				{"id": "docs", "description": "Documentation", "agent": "writer", "depends_on": ["review"]}
			]
		}`,
		Usage: provider.Usage{TotalTokens: 80},
	})

	dag, err := Plan(context.Background(), "Build a complete API", mock, "gpt-4o")
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if dag.Size() != 5 {
		t.Errorf("DAG size = %d, want 5", dag.Size())
	}

	layers, _ := dag.TopologicalSort()
	if len(layers) != 4 {
		t.Errorf("layers = %d, want 4", len(layers))
	}
	if len(layers[1]) != 2 {
		t.Errorf("layer 1 (parallel) = %d tasks, want 2", len(layers[1]))
	}
}

func TestPlanInvalidAgent(t *testing.T) {
	mock := provider.NewMockProvider(&provider.Response{
		Content: `{
			"tasks": [
				{"id": "x", "description": "test", "agent": "invalid_agent", "depends_on": []}
			]
		}`,
	})

	_, err := Plan(context.Background(), "test", mock, "gpt-4o")
	if err == nil {
		t.Fatal("expected error for invalid agent type")
	}
}

func TestPlanEmptyTasks(t *testing.T) {
	mock := provider.NewMockProvider(&provider.Response{
		Content: `{"tasks": []}`,
	})

	_, err := Plan(context.Background(), "test", mock, "gpt-4o")
	if err == nil {
		t.Fatal("expected error for empty tasks")
	}
}

func TestPlanWithMarkdown(t *testing.T) {
	mock := provider.NewMockProvider(&provider.Response{
		Content: "Here is the plan:\n```json\n" + `{
			"tasks": [
				{"id": "research", "description": "Do research", "agent": "researcher", "depends_on": []}
			]
		}` + "\n```\n",
		Usage: provider.Usage{TotalTokens: 30},
	})

	dag, err := Plan(context.Background(), "test", mock, "gpt-4o")
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if dag.Size() != 1 {
		t.Errorf("DAG size = %d, want 1", dag.Size())
	}
}
