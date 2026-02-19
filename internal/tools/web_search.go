package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/devaloi/agentforge/internal/provider"
)

// WebSearch simulates web search (or delegates to a real API when configured).
type WebSearch struct{}

func (w *WebSearch) Name() string { return "web_search" }
func (w *WebSearch) Description() string {
	return "Search the web for information and return relevant results"
}

func (w *WebSearch) Schema() provider.JSONSchema {
	return NewSchemaBuilder().
		AddString("query", "The search query", true).
		Build()
}

func (w *WebSearch) Execute(_ context.Context, params map[string]any) (string, error) {
	query, ok := params["query"].(string)
	if !ok || query == "" {
		return "", fmt.Errorf("web_search: missing required parameter 'query'")
	}

	results := []map[string]string{
		{
			"title":   fmt.Sprintf("Results for: %s", query),
			"snippet": fmt.Sprintf("Comprehensive guide on %s with best practices and examples.", query),
			"url":     fmt.Sprintf("https://example.com/search?q=%s", query),
		},
		{
			"title":   fmt.Sprintf("Advanced %s techniques", query),
			"snippet": fmt.Sprintf("Deep dive into %s covering architecture, patterns, and implementation details.", query),
			"url":     fmt.Sprintf("https://example.com/advanced/%s", query),
		},
	}

	data, err := json.Marshal(map[string]any{"results": results})
	if err != nil {
		return "", fmt.Errorf("web_search: marshal results: %w", err)
	}
	return string(data), nil
}
