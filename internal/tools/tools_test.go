package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/devaloi/agentforge/internal/provider"
)

func TestWebSearchExecute(t *testing.T) {
	ws := &WebSearch{}

	out, err := ws.Execute(context.Background(), map[string]any{"query": "Go REST API"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("parse output: %v", err)
	}

	results, ok := result["results"].([]any)
	if !ok || len(results) == 0 {
		t.Error("expected non-empty results array")
	}
}

func TestWebSearchMissingQuery(t *testing.T) {
	ws := &WebSearch{}
	_, err := ws.Execute(context.Background(), map[string]any{})
	if err == nil {
		t.Fatal("expected error for missing query")
	}
}

func TestWebSearchSchema(t *testing.T) {
	ws := &WebSearch{}
	schema := ws.Schema()
	if schema.Type != "object" {
		t.Errorf("Type = %q, want %q", schema.Type, "object")
	}
	if _, ok := schema.Properties["query"]; !ok {
		t.Error("missing 'query' property")
	}
}

func TestCodeGenExecute(t *testing.T) {
	mock := provider.NewMockProvider(&provider.Response{
		Content: "func main() { fmt.Println(\"hello\") }",
		Usage:   provider.Usage{TotalTokens: 10},
	})

	cg := NewCodeGen(mock, "gpt-4o")
	out, err := cg.Execute(context.Background(), map[string]any{
		"language": "go",
		"task":     "hello world program",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("parse output: %v", err)
	}
	if result["code"] == "" {
		t.Error("expected non-empty code")
	}
}

func TestCodeGenMissingParams(t *testing.T) {
	mock := provider.NewMockProvider()
	cg := NewCodeGen(mock, "gpt-4o")
	_, err := cg.Execute(context.Background(), map[string]any{})
	if err == nil {
		t.Fatal("expected error for missing params")
	}
}

func TestCodeReviewExecute(t *testing.T) {
	mock := provider.NewMockProvider(&provider.Response{
		Content: "No issues found",
		Usage:   provider.Usage{TotalTokens: 10},
	})

	cr := NewCodeReview(mock, "gpt-4o")
	out, err := cr.Execute(context.Background(), map[string]any{
		"code":     "func main() {}",
		"language": "go",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("parse output: %v", err)
	}
}

func TestTextGenExecute(t *testing.T) {
	mock := provider.NewMockProvider(&provider.Response{
		Content: "Go is a great language for building APIs.",
		Usage:   provider.Usage{TotalTokens: 10},
	})

	tg := NewTextGen(mock, "gpt-4o")
	out, err := tg.Execute(context.Background(), map[string]any{
		"topic": "Go programming",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("parse output: %v", err)
	}
	if result["text"] == "" {
		t.Error("expected non-empty text")
	}
}

func TestTextGenMissingTopic(t *testing.T) {
	mock := provider.NewMockProvider()
	tg := NewTextGen(mock, "gpt-4o")
	_, err := tg.Execute(context.Background(), map[string]any{})
	if err == nil {
		t.Fatal("expected error for missing topic")
	}
}
