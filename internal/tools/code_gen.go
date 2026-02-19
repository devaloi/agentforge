package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/devaloi/agentforge/internal/provider"
)

// CodeGen generates code by delegating to an LLM provider.
type CodeGen struct {
	provider provider.Provider
	model    string
}

// NewCodeGen creates a CodeGen tool backed by the given provider.
func NewCodeGen(p provider.Provider, model string) *CodeGen {
	return &CodeGen{provider: p, model: model}
}

func (c *CodeGen) Name() string { return "code_gen" }
func (c *CodeGen) Description() string {
	return "Generate code in a specified language for a given task"
}

func (c *CodeGen) Schema() provider.JSONSchema {
	return NewSchemaBuilder().
		AddString("language", "Programming language", true).
		AddString("task", "Description of the code to generate", true).
		AddString("context", "Additional context for generation", false).
		Build()
}

func (c *CodeGen) Execute(ctx context.Context, params map[string]any) (string, error) {
	language, _ := params["language"].(string)
	task, _ := params["task"].(string)
	ctxStr, _ := params["context"].(string)

	if language == "" || task == "" {
		return "", fmt.Errorf("code_gen: 'language' and 'task' are required")
	}

	prompt := fmt.Sprintf("Generate %s code for the following task: %s", language, task)
	if ctxStr != "" {
		prompt += "\n\nContext:\n" + ctxStr
	}
	prompt += "\n\nReturn valid, production-quality code with comments."

	resp, err := c.provider.ChatComplete(ctx, []provider.Message{
		{Role: provider.RoleUser, Content: prompt},
	}, nil, provider.ChatConfig{Model: c.model})
	if err != nil {
		return "", fmt.Errorf("code_gen: %w", err)
	}

	result := map[string]string{
		"code":        resp.Content,
		"explanation": fmt.Sprintf("Generated %s code for: %s", language, task),
	}
	data, _ := json.Marshal(result)
	return string(data), nil
}
