package supervisor

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/devaloi/agentforge/internal/planner"
	"github.com/devaloi/agentforge/internal/provider"
)

const plannerPrompt = `You are a task planner. Decompose the given task into sub-tasks that can be executed by specialized agents.

Available agent types:
- researcher: Searches the web for information, stores findings in shared memory
- coder: Generates code, reads/writes files and memory
- reviewer: Reviews code for bugs and issues
- writer: Generates documentation and structured text

Return a JSON object with this exact structure:
{
  "tasks": [
    {"id": "unique_id", "description": "what to do", "agent": "agent_type", "depends_on": []}
  ]
}

Rules:
- Each task must have a unique id (lowercase, underscore-separated)
- depends_on lists task IDs that must complete before this task starts
- Use minimal tasks needed - don't over-decompose
- Ensure dependencies form a valid DAG (no cycles)`

// planResponse is the expected JSON structure from the LLM planner.
type planResponse struct {
	Tasks []planTask `json:"tasks"`
}

type planTask struct {
	ID           string   `json:"id"`
	Description  string   `json:"description"`
	AgentType    string   `json:"agent"`
	Dependencies []string `json:"depends_on"`
}

// Plan decomposes a complex task into a DAG of sub-tasks using an LLM.
func Plan(ctx context.Context, task string, p provider.Provider, model string) (*planner.DAG, error) {
	messages := []provider.Message{
		{Role: provider.RoleSystem, Content: plannerPrompt},
		{Role: provider.RoleUser, Content: task},
	}

	resp, err := p.ChatComplete(ctx, messages, nil, provider.ChatConfig{Model: model})
	if err != nil {
		return nil, fmt.Errorf("planner LLM call: %w", err)
	}

	return parsePlan(resp.Content)
}

func parsePlan(content string) (*planner.DAG, error) {
	content = extractJSON(content)

	var plan planResponse
	if err := json.Unmarshal([]byte(content), &plan); err != nil {
		return nil, fmt.Errorf("parse plan: %w", err)
	}

	if len(plan.Tasks) == 0 {
		return nil, fmt.Errorf("plan contains no tasks")
	}

	dag := planner.NewDAG()
	validAgents := map[string]bool{
		"researcher": true,
		"coder":      true,
		"reviewer":   true,
		"writer":     true,
	}

	for _, t := range plan.Tasks {
		if !validAgents[t.AgentType] {
			return nil, fmt.Errorf("task %q: invalid agent type %q", t.ID, t.AgentType)
		}
		dag.AddNode(planner.SubTask{
			ID:           t.ID,
			Description:  t.Description,
			AgentType:    t.AgentType,
			Dependencies: t.Dependencies,
			Status:       planner.TaskPending,
		})
	}

	for _, t := range plan.Tasks {
		for _, dep := range t.Dependencies {
			if err := dag.AddEdge(dep, t.ID); err != nil {
				return nil, fmt.Errorf("invalid dependency: %w", err)
			}
		}
	}

	if _, err := dag.TopologicalSort(); err != nil {
		return nil, fmt.Errorf("plan validation: %w", err)
	}

	return dag, nil
}

// extractJSON finds a JSON object in the response content.
func extractJSON(s string) string {
	start := strings.Index(s, "{")
	if start < 0 {
		return s
	}
	depth := 0
	for i := start; i < len(s); i++ {
		switch s[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[start : i+1]
			}
		}
	}
	return s[start:]
}
