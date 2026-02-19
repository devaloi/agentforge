package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/devaloi/agentforge/internal/provider"
)

// TextGen generates text by delegating to an LLM provider.
type TextGen struct {
	provider provider.Provider
	model    string
}

// NewTextGen creates a TextGen tool backed by the given provider.
func NewTextGen(p provider.Provider, model string) *TextGen {
	return &TextGen{provider: p, model: model}
}

func (t *TextGen) Name() string        { return "text_gen" }
func (t *TextGen) Description() string { return "Generate structured text on a given topic" }

func (t *TextGen) Schema() provider.JSONSchema {
	return NewSchemaBuilder().
		AddString("topic", "The topic to write about", true).
		AddString("style", "Writing style (e.g. technical, casual)", false).
		AddString("context", "Additional context for generation", false).
		Build()
}

func (t *TextGen) Execute(ctx context.Context, params map[string]any) (string, error) {
	topic, _ := params["topic"].(string)
	style, _ := params["style"].(string)
	ctxStr, _ := params["context"].(string)

	if topic == "" {
		return "", fmt.Errorf("text_gen: 'topic' is required")
	}

	prompt := fmt.Sprintf("Write about: %s", topic)
	if style != "" {
		prompt += fmt.Sprintf("\nStyle: %s", style)
	}
	if ctxStr != "" {
		prompt += fmt.Sprintf("\n\nContext:\n%s", ctxStr)
	}

	resp, err := t.provider.ChatComplete(ctx, []provider.Message{
		{Role: provider.RoleUser, Content: prompt},
	}, nil, provider.ChatConfig{Model: t.model})
	if err != nil {
		return "", fmt.Errorf("text_gen: %w", err)
	}

	result := map[string]string{"text": resp.Content}
	data, _ := json.Marshal(result)
	return string(data), nil
}
