// Package planner defines the DAG data structure and sub-task types
// used for task decomposition and execution planning.
package planner

// TaskStatus represents the execution state of a sub-task.
type TaskStatus string

const (
	TaskPending  TaskStatus = "pending"
	TaskRunning  TaskStatus = "running"
	TaskComplete TaskStatus = "complete"
	TaskFailed   TaskStatus = "failed"
	TaskBlocked  TaskStatus = "blocked"
)

// SubTask represents a single unit of work within a task decomposition.
type SubTask struct {
	ID           string     `json:"id"`
	Description  string     `json:"description"`
	AgentType    string     `json:"agent"`
	Dependencies []string   `json:"depends_on"`
	Status       TaskStatus `json:"status"`
	Result       string     `json:"result,omitempty"`
}
