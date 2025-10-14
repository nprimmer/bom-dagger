package dag

import (
	"testing"

	"github.com/nprimmer/bom-dagger/internal/sbom"
)

func TestNewGraph(t *testing.T) {
	g := New()
	if g == nil {
		t.Fatal("New() returned nil")
	}
	if g.Nodes == nil {
		t.Error("Nodes map is nil")
	}
	if g.Roots == nil {
		t.Error("Roots slice is nil")
	}
}

func TestBuildFromSBOM(t *testing.T) {
	// Create a sample SBOM
	bom := &sbom.CycloneDX{
		BOMFormat:   "CycloneDX",
		SpecVersion: "1.4",
		Components: []sbom.Component{
			{BOMRef: "comp-a", Name: "Component A", Version: "1.0"},
			{BOMRef: "comp-b", Name: "Component B", Version: "1.0"},
			{BOMRef: "comp-c", Name: "Component C", Version: "1.0"},
		},
		Dependencies: []sbom.Dependency{
			{Ref: "comp-a", DependsOn: []string{"comp-b"}},
			{Ref: "comp-b", DependsOn: []string{"comp-c"}},
			{Ref: "comp-c", DependsOn: []string{}},
		},
	}

	// Create component map
	componentMap := map[string]*sbom.Component{
		"comp-a": &bom.Components[0],
		"comp-b": &bom.Components[1],
		"comp-c": &bom.Components[2],
	}

	g := New()
	err := g.BuildFromSBOM(bom, componentMap)
	if err != nil {
		t.Fatalf("BuildFromSBOM failed: %v", err)
	}

	// Check nodes were created
	if len(g.Nodes) != 3 {
		t.Errorf("Expected 3 nodes, got %d", len(g.Nodes))
	}

	// Check dependencies
	if len(g.Nodes["comp-a"].Dependencies) != 1 {
		t.Errorf("comp-a should have 1 dependency, got %d", len(g.Nodes["comp-a"].Dependencies))
	}
	if len(g.Nodes["comp-b"].Dependencies) != 1 {
		t.Errorf("comp-b should have 1 dependency, got %d", len(g.Nodes["comp-b"].Dependencies))
	}
	if len(g.Nodes["comp-c"].Dependencies) != 0 {
		t.Errorf("comp-c should have 0 dependencies, got %d", len(g.Nodes["comp-c"].Dependencies))
	}

	// Check roots
	if len(g.Roots) != 1 {
		t.Errorf("Expected 1 root node, got %d", len(g.Roots))
	}
	if g.Roots[0].ID != "comp-c" {
		t.Errorf("Expected comp-c as root, got %s", g.Roots[0].ID)
	}
}

func TestCycleDetection(t *testing.T) {
	// Create a SBOM with a cycle
	bom := &sbom.CycloneDX{
		BOMFormat:   "CycloneDX",
		SpecVersion: "1.4",
		Components: []sbom.Component{
			{BOMRef: "comp-a", Name: "Component A", Version: "1.0"},
			{BOMRef: "comp-b", Name: "Component B", Version: "1.0"},
			{BOMRef: "comp-c", Name: "Component C", Version: "1.0"},
		},
		Dependencies: []sbom.Dependency{
			{Ref: "comp-a", DependsOn: []string{"comp-b"}},
			{Ref: "comp-b", DependsOn: []string{"comp-c"}},
			{Ref: "comp-c", DependsOn: []string{"comp-a"}}, // Creates a cycle
		},
	}

	// Create component map
	componentMap := map[string]*sbom.Component{
		"comp-a": &bom.Components[0],
		"comp-b": &bom.Components[1],
		"comp-c": &bom.Components[2],
	}

	g := New()
	err := g.BuildFromSBOM(bom, componentMap)
	if err == nil {
		t.Error("Expected error for cyclic dependency, but got none")
	}
	if err.Error() != "dependency graph contains cycles" {
		t.Errorf("Expected cycle error message, got: %v", err)
	}
}

func TestGetNodeCount(t *testing.T) {
	g := New()
	g.Nodes["a"] = &Node{ID: "a"}
	g.Nodes["b"] = &Node{ID: "b"}
	g.Nodes["c"] = &Node{ID: "c"}

	if g.GetNodeCount() != 3 {
		t.Errorf("Expected 3 nodes, got %d", g.GetNodeCount())
	}
}

func TestGetEdgeCount(t *testing.T) {
	g := New()
	nodeA := &Node{ID: "a"}
	nodeB := &Node{ID: "b"}
	nodeC := &Node{ID: "c"}

	nodeA.Dependencies = []*Node{nodeB, nodeC}
	nodeB.Dependencies = []*Node{nodeC}

	g.Nodes["a"] = nodeA
	g.Nodes["b"] = nodeB
	g.Nodes["c"] = nodeC

	if g.GetEdgeCount() != 3 {
		t.Errorf("Expected 3 edges, got %d", g.GetEdgeCount())
	}
}
