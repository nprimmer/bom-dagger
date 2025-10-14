package dag

import (
	"testing"

	"github.com/nprimmer/bom-dagger/internal/sbom"
)

func createTestGraph() *Graph {
	g := New()

	// Create components
	compA := &sbom.Component{Name: "App", Version: "1.0"}
	compB := &sbom.Component{Name: "Database", Version: "2.0"}
	compC := &sbom.Component{Name: "Cache", Version: "1.5"}
	compD := &sbom.Component{Name: "MessageQueue", Version: "3.0"}

	// Create nodes
	nodeA := &Node{ID: "app", Component: compA}
	nodeB := &Node{ID: "db", Component: compB}
	nodeC := &Node{ID: "cache", Component: compC}
	nodeD := &Node{ID: "mq", Component: compD}

	// Set up dependencies: App depends on DB and Cache, Cache depends on MQ
	nodeA.Dependencies = []*Node{nodeB, nodeC}
	nodeB.Dependents = []*Node{nodeA}
	nodeC.Dependencies = []*Node{nodeD}
	nodeC.Dependents = []*Node{nodeA}
	nodeD.Dependents = []*Node{nodeC}

	// Add to graph
	g.Nodes["app"] = nodeA
	g.Nodes["db"] = nodeB
	g.Nodes["cache"] = nodeC
	g.Nodes["mq"] = nodeD

	// Set roots (nodes with no dependencies)
	g.Roots = []*Node{nodeB, nodeD}

	return g
}

func TestTopologicalSort(t *testing.T) {
	g := createTestGraph()

	order, err := g.TopologicalSort()
	if err != nil {
		t.Fatalf("TopologicalSort failed: %v", err)
	}

	if len(order) != 4 {
		t.Errorf("Expected 4 items in order, got %d", len(order))
	}

	// Create a map to track deployment order
	orderMap := make(map[string]int)
	for _, item := range order {
		orderMap[item.BOMRef] = item.Step
	}

	// Verify dependencies are satisfied
	// App should come after DB and Cache
	if orderMap["app"] <= orderMap["db"] {
		t.Error("App should be deployed after Database")
	}
	if orderMap["app"] <= orderMap["cache"] {
		t.Error("App should be deployed after Cache")
	}
	// Cache should come after MQ
	if orderMap["cache"] <= orderMap["mq"] {
		t.Error("Cache should be deployed after MessageQueue")
	}
}

func TestReverseTopologicalSort(t *testing.T) {
	g := createTestGraph()

	order, err := g.ReverseTopologicalSort()
	if err != nil {
		t.Fatalf("ReverseTopologicalSort failed: %v", err)
	}

	if len(order) != 4 {
		t.Errorf("Expected 4 items in order, got %d", len(order))
	}

	// Create a map to track teardown order
	orderMap := make(map[string]int)
	for _, item := range order {
		orderMap[item.BOMRef] = item.Step
	}

	// Verify reverse order
	// App should come before its dependencies in teardown
	if orderMap["app"] >= orderMap["db"] {
		t.Error("App should be torn down before Database")
	}
	if orderMap["app"] >= orderMap["cache"] {
		t.Error("App should be torn down before Cache")
	}
	// Cache should come before MQ in teardown
	if orderMap["cache"] >= orderMap["mq"] {
		t.Error("Cache should be torn down before MessageQueue")
	}
}

func TestGetDeploymentGroups(t *testing.T) {
	g := createTestGraph()

	groups, err := g.GetDeploymentGroups()
	if err != nil {
		t.Fatalf("GetDeploymentGroups failed: %v", err)
	}

	// Should have 3 groups:
	// Group 1: DB and MQ (can deploy in parallel)
	// Group 2: Cache (depends on MQ)
	// Group 3: App (depends on DB and Cache)
	if len(groups) != 3 {
		t.Errorf("Expected 3 deployment groups, got %d", len(groups))
	}

	// First group should have 2 items (DB and MQ)
	if len(groups[0]) != 2 {
		t.Errorf("Expected 2 items in first group, got %d", len(groups[0]))
	}

	// Second group should have 1 item (Cache)
	if len(groups[1]) != 1 {
		t.Errorf("Expected 1 item in second group, got %d", len(groups[1]))
	}

	// Third group should have 1 item (App)
	if len(groups[2]) != 1 {
		t.Errorf("Expected 1 item in third group, got %d", len(groups[2]))
	}
}

func TestTopologicalSortWithCycle(t *testing.T) {
	g := New()

	// Create a graph with a cycle
	compA := &sbom.Component{Name: "A", Version: "1.0"}
	compB := &sbom.Component{Name: "B", Version: "1.0"}

	nodeA := &Node{ID: "a", Component: compA}
	nodeB := &Node{ID: "b", Component: compB}

	// Create cycle: A depends on B, B depends on A
	nodeA.Dependencies = []*Node{nodeB}
	nodeB.Dependencies = []*Node{nodeA}
	nodeA.Dependents = []*Node{nodeB}
	nodeB.Dependents = []*Node{nodeA}

	g.Nodes["a"] = nodeA
	g.Nodes["b"] = nodeB

	_, err := g.TopologicalSort()
	if err == nil {
		t.Error("Expected error for cyclic graph, but got none")
	}
}
