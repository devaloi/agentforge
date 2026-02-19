package planner

import (
	"fmt"
	"sort"
)

// DAG represents a directed acyclic graph of sub-tasks with dependency edges.
type DAG struct {
	nodes map[string]*SubTask
	edges map[string][]string // from → []to (from must complete before to)
	inDeg map[string]int
}

// NewDAG creates an empty DAG.
func NewDAG() *DAG {
	return &DAG{
		nodes: make(map[string]*SubTask),
		edges: make(map[string][]string),
		inDeg: make(map[string]int),
	}
}

// AddNode adds a sub-task to the DAG.
func (d *DAG) AddNode(task SubTask) {
	t := task
	d.nodes[task.ID] = &t
	if _, ok := d.inDeg[task.ID]; !ok {
		d.inDeg[task.ID] = 0
	}
}

// AddEdge adds a dependency: from must complete before to can start.
func (d *DAG) AddEdge(from, to string) error {
	if _, ok := d.nodes[from]; !ok {
		return fmt.Errorf("unknown node: %s", from)
	}
	if _, ok := d.nodes[to]; !ok {
		return fmt.Errorf("unknown node: %s", to)
	}
	d.edges[from] = append(d.edges[from], to)
	d.inDeg[to]++
	return nil
}

// Nodes returns all sub-tasks in the DAG.
func (d *DAG) Nodes() []*SubTask {
	tasks := make([]*SubTask, 0, len(d.nodes))
	for _, t := range d.nodes {
		tasks = append(tasks, t)
	}
	return tasks
}

// GetNode returns a sub-task by ID.
func (d *DAG) GetNode(id string) (*SubTask, bool) {
	t, ok := d.nodes[id]
	return t, ok
}

// TopologicalSort returns execution layers — groups of tasks that can run in parallel.
// Returns an error if the DAG contains a cycle.
func (d *DAG) TopologicalSort() ([][]string, error) {
	inDeg := make(map[string]int)
	for id, deg := range d.inDeg {
		inDeg[id] = deg
	}

	var layers [][]string
	remaining := len(d.nodes)

	for remaining > 0 {
		var layer []string
		for id, deg := range inDeg {
			if deg == 0 {
				layer = append(layer, id)
			}
		}

		if len(layer) == 0 {
			return nil, fmt.Errorf("cycle detected in DAG")
		}

		sort.Strings(layer)
		layers = append(layers, layer)

		for _, id := range layer {
			delete(inDeg, id)
			remaining--
			for _, to := range d.edges[id] {
				inDeg[to]--
			}
		}
	}

	return layers, nil
}

// Ready returns task IDs whose dependencies are all in the completed set.
func (d *DAG) Ready(completed map[string]bool) []string {
	var ready []string
	for id, task := range d.nodes {
		if completed[id] || task.Status == TaskComplete || task.Status == TaskRunning {
			continue
		}
		allDeps := true
		for _, dep := range task.Dependencies {
			if !completed[dep] {
				allDeps = false
				break
			}
		}
		if allDeps {
			ready = append(ready, id)
		}
	}
	sort.Strings(ready)
	return ready
}

// Size returns the number of nodes in the DAG.
func (d *DAG) Size() int {
	return len(d.nodes)
}
