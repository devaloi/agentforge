package planner

import "testing"

func TestDAGSimpleChain(t *testing.T) {
	dag := NewDAG()
	dag.AddNode(SubTask{ID: "a", Description: "Step A", AgentType: "researcher"})
	dag.AddNode(SubTask{ID: "b", Description: "Step B", AgentType: "coder"})
	dag.AddNode(SubTask{ID: "c", Description: "Step C", AgentType: "reviewer"})

	_ = dag.AddEdge("a", "b")
	_ = dag.AddEdge("b", "c")

	layers, err := dag.TopologicalSort()
	if err != nil {
		t.Fatalf("TopologicalSort: %v", err)
	}

	if len(layers) != 3 {
		t.Fatalf("layers = %d, want 3", len(layers))
	}
	if layers[0][0] != "a" {
		t.Errorf("layer 0 = %v, want [a]", layers[0])
	}
	if layers[1][0] != "b" {
		t.Errorf("layer 1 = %v, want [b]", layers[1])
	}
	if layers[2][0] != "c" {
		t.Errorf("layer 2 = %v, want [c]", layers[2])
	}
}

func TestDAGDiamondDependency(t *testing.T) {
	dag := NewDAG()
	dag.AddNode(SubTask{ID: "research", AgentType: "researcher"})
	dag.AddNode(SubTask{ID: "code_a", AgentType: "coder"})
	dag.AddNode(SubTask{ID: "code_b", AgentType: "coder"})
	dag.AddNode(SubTask{ID: "review", AgentType: "reviewer"})

	_ = dag.AddEdge("research", "code_a")
	_ = dag.AddEdge("research", "code_b")
	_ = dag.AddEdge("code_a", "review")
	_ = dag.AddEdge("code_b", "review")

	layers, err := dag.TopologicalSort()
	if err != nil {
		t.Fatalf("TopologicalSort: %v", err)
	}

	if len(layers) != 3 {
		t.Fatalf("layers = %d, want 3", len(layers))
	}
	if len(layers[0]) != 1 || layers[0][0] != "research" {
		t.Errorf("layer 0 = %v, want [research]", layers[0])
	}
	if len(layers[1]) != 2 {
		t.Errorf("layer 1 = %v, want 2 parallel tasks", layers[1])
	}
	if len(layers[2]) != 1 || layers[2][0] != "review" {
		t.Errorf("layer 2 = %v, want [review]", layers[2])
	}
}

func TestDAGCycleDetection(t *testing.T) {
	dag := NewDAG()
	dag.AddNode(SubTask{ID: "a"})
	dag.AddNode(SubTask{ID: "b"})
	dag.AddNode(SubTask{ID: "c"})

	_ = dag.AddEdge("a", "b")
	_ = dag.AddEdge("b", "c")
	_ = dag.AddEdge("c", "a")

	_, err := dag.TopologicalSort()
	if err == nil {
		t.Fatal("expected cycle detection error")
	}
}

func TestDAGReady(t *testing.T) {
	dag := NewDAG()
	dag.AddNode(SubTask{ID: "a", Dependencies: nil})
	dag.AddNode(SubTask{ID: "b", Dependencies: []string{"a"}})
	dag.AddNode(SubTask{ID: "c", Dependencies: []string{"a"}})
	dag.AddNode(SubTask{ID: "d", Dependencies: []string{"b", "c"}})

	ready := dag.Ready(map[string]bool{})
	if len(ready) != 1 || ready[0] != "a" {
		t.Errorf("ready = %v, want [a]", ready)
	}

	ready = dag.Ready(map[string]bool{"a": true})
	if len(ready) != 2 {
		t.Errorf("ready = %v, want [b, c]", ready)
	}

	ready = dag.Ready(map[string]bool{"a": true, "b": true, "c": true})
	if len(ready) != 1 || ready[0] != "d" {
		t.Errorf("ready = %v, want [d]", ready)
	}
}

func TestDAGInvalidEdge(t *testing.T) {
	dag := NewDAG()
	dag.AddNode(SubTask{ID: "a"})

	err := dag.AddEdge("a", "nonexistent")
	if err == nil {
		t.Error("expected error for invalid edge target")
	}
}

func TestDAGParallelGroups(t *testing.T) {
	dag := NewDAG()
	dag.AddNode(SubTask{ID: "a"})
	dag.AddNode(SubTask{ID: "b"})
	dag.AddNode(SubTask{ID: "c"})

	layers, err := dag.TopologicalSort()
	if err != nil {
		t.Fatalf("TopologicalSort: %v", err)
	}

	if len(layers) != 1 {
		t.Fatalf("layers = %d, want 1 (all parallel)", len(layers))
	}
	if len(layers[0]) != 3 {
		t.Errorf("layer 0 = %v, want 3 parallel tasks", layers[0])
	}
}

func TestDAGSize(t *testing.T) {
	dag := NewDAG()
	if dag.Size() != 0 {
		t.Errorf("Size = %d, want 0", dag.Size())
	}
	dag.AddNode(SubTask{ID: "a"})
	dag.AddNode(SubTask{ID: "b"})
	if dag.Size() != 2 {
		t.Errorf("Size = %d, want 2", dag.Size())
	}
}
