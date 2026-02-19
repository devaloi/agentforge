package planner

import (
	"encoding/json"
	"testing"
)

func TestSubTaskJSON(t *testing.T) {
	task := SubTask{
		ID:           "research",
		Description:  "Research Go REST APIs",
		AgentType:    "researcher",
		Dependencies: []string{},
		Status:       TaskPending,
	}

	data, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded SubTask
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ID != task.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, task.ID)
	}
	if decoded.AgentType != task.AgentType {
		t.Errorf("AgentType = %q, want %q", decoded.AgentType, task.AgentType)
	}
	if decoded.Status != TaskPending {
		t.Errorf("Status = %q, want %q", decoded.Status, TaskPending)
	}
}

func TestSubTaskWithDependencies(t *testing.T) {
	task := SubTask{
		ID:           "code",
		Description:  "Write code",
		AgentType:    "coder",
		Dependencies: []string{"research", "design"},
		Status:       TaskPending,
	}

	data, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded SubTask
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(decoded.Dependencies) != 2 {
		t.Fatalf("Dependencies len = %d, want 2", len(decoded.Dependencies))
	}
	if decoded.Dependencies[0] != "research" {
		t.Errorf("Dependencies[0] = %q, want %q", decoded.Dependencies[0], "research")
	}
}
