package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/devaloi/agentforge/internal/provider"
)

// CodeReview analyzes code for issues by delegating to an LLM provider.
type CodeReview struct {
	provider provider.Provider
	model    string
}

// NewCodeReview creates a CodeReview tool backed by the given provider.
func NewCodeReview(p provider.Provider, model string) *CodeReview {
	return &CodeReview{provider: p, model: model}
}

func (c *CodeReview) Name() string { return "code_review" }
func (c *CodeReview) Description() string {
	return "Review code for bugs, security issues, and style problems"
}

func (c *CodeReview) Schema() provider.JSONSchema {
	return NewSchemaBuilder().
		AddString("code", "The code to review", true).
		AddString("language", "Programming language of the code", true).
		Build()
}

func (c *CodeReview) Execute(ctx context.Context, params map[string]any) (string, error) {
	code, _ := params["code"].(string)
	language, _ := params["language"].(string)

	if code == "" || language == "" {
		return "", fmt.Errorf("code_review: 'code' and 'language' are required")
	}

	prompt := fmt.Sprintf(`Review the following %s code for bugs, security issues, performance problems, and style:

%s

Return a JSON response with "issues" (array of {severity, line, message}) and "approved" (boolean).`, language, code)

	resp, err := c.provider.ChatComplete(ctx, []provider.Message{
		{Role: provider.RoleUser, Content: prompt},
	}, nil, provider.ChatConfig{Model: c.model})
	if err != nil {
		return "", fmt.Errorf("code_review: %w", err)
	}

	result := map[string]any{
		"review":   resp.Content,
		"approved": true,
	}
	data, _ := json.Marshal(result)
	return string(data), nil
}
