package dag

import (
	"testing"

	"github.com/nprimmer/bom-dagger/internal/sbom"
)

func TestEdgeCases(t *testing.T) {
	t.Run("empty graph", func(t *testing.T) {
		g := New()

		order, err := g.TopologicalSort()
		if err != nil {
			t.Errorf("TopologicalSort on empty graph failed: %v", err)
		}
		if len(order) != 0 {
			t.Errorf("Expected empty order for empty graph, got %d items", len(order))
		}

		groups, err := g.GetDeploymentGroups()
		if err != nil {
			t.Errorf("GetDeploymentGroups on empty graph failed: %v", err)
		}
		if len(groups) != 0 {
			t.Errorf("Expected empty groups for empty graph, got %d groups", len(groups))
		}
	})

	t.Run("single node", func(t *testing.T) {
		g := New()
		comp := &sbom.Component{Name: "Single", Version: "1.0"}
		node := &Node{ID: "single", Component: comp}
		g.Nodes["single"] = node
		g.Roots = []*Node{node}

		order, err := g.TopologicalSort()
		if err != nil {
			t.Errorf("TopologicalSort failed: %v", err)
		}
		if len(order) != 1 {
			t.Errorf("Expected 1 item in order, got %d", len(order))
		}
		if order[0].BOMRef != "single" {
			t.Errorf("Expected 'single' in order, got %s", order[0].BOMRef)
		}
	})

	t.Run("self-dependency", func(t *testing.T) {
		g := New()
		comp := &sbom.Component{Name: "SelfRef", Version: "1.0"}
		node := &Node{ID: "self", Component: comp}

		// Create self-reference
		node.Dependencies = []*Node{node}
		node.Dependents = []*Node{node}

		g.Nodes["self"] = node

		_, err := g.TopologicalSort()
		if err == nil {
			t.Error("Expected error for self-referencing node")
		}
	})

	t.Run("disconnected components", func(t *testing.T) {
		g := New()

		// Create two disconnected subgraphs
		compA1 := &sbom.Component{Name: "A1", Version: "1.0"}
		compA2 := &sbom.Component{Name: "A2", Version: "1.0"}
		compB1 := &sbom.Component{Name: "B1", Version: "1.0"}
		compB2 := &sbom.Component{Name: "B2", Version: "1.0"}

		nodeA1 := &Node{ID: "a1", Component: compA1}
		nodeA2 := &Node{ID: "a2", Component: compA2}
		nodeB1 := &Node{ID: "b1", Component: compB1}
		nodeB2 := &Node{ID: "b2", Component: compB2}

		// A1 -> A2
		nodeA1.Dependencies = []*Node{nodeA2}
		nodeA2.Dependents = []*Node{nodeA1}

		// B1 -> B2
		nodeB1.Dependencies = []*Node{nodeB2}
		nodeB2.Dependents = []*Node{nodeB1}

		g.Nodes["a1"] = nodeA1
		g.Nodes["a2"] = nodeA2
		g.Nodes["b1"] = nodeB1
		g.Nodes["b2"] = nodeB2
		g.Roots = []*Node{nodeA2, nodeB2}

		order, err := g.TopologicalSort()
		if err != nil {
			t.Errorf("TopologicalSort failed: %v", err)
		}
		if len(order) != 4 {
			t.Errorf("Expected 4 items in order, got %d", len(order))
		}

		groups, err := g.GetDeploymentGroups()
		if err != nil {
			t.Errorf("GetDeploymentGroups failed: %v", err)
		}
		// A2 and B2 should be in the same group (both are roots)
		// A1 and B1 should be in the same group (both depend on roots)
		if len(groups) != 2 {
			t.Errorf("Expected 2 groups, got %d", len(groups))
		}
	})

	t.Run("long chain", func(t *testing.T) {
		g := New()

		// Create a long chain: A -> B -> C -> D -> E
		chainLength := 5
		nodes := make([]*Node, chainLength)

		for i := 0; i < chainLength; i++ {
			comp := &sbom.Component{
				Name:    string(rune('A' + i)),
				Version: "1.0",
			}
			nodes[i] = &Node{
				ID:           string(rune('a' + i)),
				Component:    comp,
				Dependencies: []*Node{},
				Dependents:   []*Node{},
			}
			g.Nodes[nodes[i].ID] = nodes[i]
		}

		// Set up chain dependencies
		for i := 0; i < chainLength-1; i++ {
			nodes[i].Dependencies = []*Node{nodes[i+1]}
			nodes[i+1].Dependents = []*Node{nodes[i]}
		}

		g.Roots = []*Node{nodes[chainLength-1]}

		order, err := g.TopologicalSort()
		if err != nil {
			t.Errorf("TopologicalSort failed: %v", err)
		}
		if len(order) != chainLength {
			t.Errorf("Expected %d items in order, got %d", chainLength, len(order))
		}

		groups, err := g.GetDeploymentGroups()
		if err != nil {
			t.Errorf("GetDeploymentGroups failed: %v", err)
		}
		// Each node should be in its own group (sequential dependencies)
		if len(groups) != chainLength {
			t.Errorf("Expected %d groups, got %d", chainLength, len(groups))
		}
	})

	t.Run("diamond dependency", func(t *testing.T) {
		g := New()

		// Create diamond: A -> B,C -> D
		compA := &sbom.Component{Name: "A", Version: "1.0"}
		compB := &sbom.Component{Name: "B", Version: "1.0"}
		compC := &sbom.Component{Name: "C", Version: "1.0"}
		compD := &sbom.Component{Name: "D", Version: "1.0"}

		nodeA := &Node{ID: "a", Component: compA}
		nodeB := &Node{ID: "b", Component: compB}
		nodeC := &Node{ID: "c", Component: compC}
		nodeD := &Node{ID: "d", Component: compD}

		// A depends on B and C
		nodeA.Dependencies = []*Node{nodeB, nodeC}
		nodeB.Dependents = []*Node{nodeA}
		nodeC.Dependents = []*Node{nodeA}

		// B and C both depend on D
		nodeB.Dependencies = []*Node{nodeD}
		nodeC.Dependencies = []*Node{nodeD}
		nodeD.Dependents = []*Node{nodeB, nodeC}

		g.Nodes["a"] = nodeA
		g.Nodes["b"] = nodeB
		g.Nodes["c"] = nodeC
		g.Nodes["d"] = nodeD
		g.Roots = []*Node{nodeD}

		order, err := g.TopologicalSort()
		if err != nil {
			t.Errorf("TopologicalSort failed: %v", err)
		}
		if len(order) != 4 {
			t.Errorf("Expected 4 items in order, got %d", len(order))
		}

		groups, err := g.GetDeploymentGroups()
		if err != nil {
			t.Errorf("GetDeploymentGroups failed: %v", err)
		}
		// Should have 3 groups: D, then B&C, then A
		if len(groups) != 3 {
			t.Errorf("Expected 3 groups, got %d", len(groups))
		}
		if len(groups[1]) != 2 {
			t.Errorf("Expected 2 items in second group (B and C), got %d", len(groups[1]))
		}
	})

	t.Run("missing dependency references", func(t *testing.T) {
		bom := &sbom.CycloneDX{
			BOMFormat:   "CycloneDX",
			SpecVersion: "1.6",
			Components: []sbom.Component{
				{BOMRef: "exists", Name: "Existing Component", Version: "1.0"},
			},
			Dependencies: []sbom.Dependency{
				{Ref: "exists", DependsOn: []string{"missing-ref"}},
				{Ref: "missing-component", DependsOn: []string{"exists"}},
			},
		}

		componentMap := map[string]*sbom.Component{
			"exists": &bom.Components[0],
		}

		g := New()
		err := g.BuildFromSBOM(bom, componentMap)
		// Should not error, just skip missing references
		if err != nil {
			t.Errorf("BuildFromSBOM should handle missing references gracefully: %v", err)
		}

		// Should only have the existing component
		if len(g.Nodes) != 1 {
			t.Errorf("Expected 1 node, got %d", len(g.Nodes))
		}
	})

	t.Run("duplicate bomrefs", func(t *testing.T) {
		bom := &sbom.CycloneDX{
			BOMFormat:   "CycloneDX",
			SpecVersion: "1.6",
			Components: []sbom.Component{
				{BOMRef: "duplicate", Name: "First", Version: "1.0"},
				{BOMRef: "duplicate", Name: "Second", Version: "2.0"},
			},
		}

		// In the component map, the second one would overwrite the first
		componentMap := make(map[string]*sbom.Component)
		for i := range bom.Components {
			componentMap[bom.Components[i].BOMRef] = &bom.Components[i]
		}

		g := New()
		err := g.BuildFromSBOM(bom, componentMap)
		if err != nil {
			t.Errorf("BuildFromSBOM failed: %v", err)
		}

		// Should have one node (last one wins in map)
		if len(g.Nodes) != 1 {
			t.Errorf("Expected 1 node for duplicate refs, got %d", len(g.Nodes))
		}
	})
}

func TestServiceNodeHandling(t *testing.T) {
	t.Run("mixed components and services", func(t *testing.T) {
		bom := &sbom.CycloneDX{
			BOMFormat:   "CycloneDX",
			SpecVersion: "1.6",
			Components: []sbom.Component{
				{BOMRef: "comp-1", Name: "Component 1", Version: "1.0"},
			},
			Services: []sbom.Service{
				{BOMRef: "svc-1", Name: "Service 1", Version: "2.0"},
				{BOMRef: "svc-2", Name: "Service 2"}, // No version
			},
			Dependencies: []sbom.Dependency{
				{Ref: "comp-1", DependsOn: []string{"svc-1"}},
				{Ref: "svc-1", DependsOn: []string{"svc-2"}},
				{Ref: "svc-2", DependsOn: []string{}},
			},
		}

		componentMap := map[string]*sbom.Component{
			"comp-1": &bom.Components[0],
		}

		g := New()
		err := g.BuildFromSBOM(bom, componentMap)
		if err != nil {
			t.Errorf("BuildFromSBOM failed: %v", err)
		}

		// Should have 3 nodes total
		if len(g.Nodes) != 3 {
			t.Errorf("Expected 3 nodes, got %d", len(g.Nodes))
		}

		// Verify service nodes were created
		svcNode1, exists := g.Nodes["svc-1"]
		if !exists {
			t.Error("Service node svc-1 not created")
		} else if svcNode1.Service == nil {
			t.Error("Node svc-1 should have Service field set")
		}

		svcNode2, exists := g.Nodes["svc-2"]
		if !exists {
			t.Error("Service node svc-2 not created")
		} else if svcNode2.Service == nil {
			t.Error("Node svc-2 should have Service field set")
		}

		// Test topological sort with mixed types
		order, err := g.TopologicalSort()
		if err != nil {
			t.Errorf("TopologicalSort failed: %v", err)
		}
		if len(order) != 3 {
			t.Errorf("Expected 3 items in order, got %d", len(order))
		}

		// Service 2 should come first (no dependencies)
		if order[0].Component != "Service 2" {
			t.Errorf("Expected 'Service 2' first, got '%s'", order[0].Component)
		}
	})

	t.Run("service without bomref", func(t *testing.T) {
		bom := &sbom.CycloneDX{
			BOMFormat:   "CycloneDX",
			SpecVersion: "1.6",
			Services: []sbom.Service{
				{Name: "No Ref Service"}, // No BOMRef
				{BOMRef: "svc-with-ref", Name: "Service With Ref"},
			},
		}

		g := New()
		err := g.BuildFromSBOM(bom, make(map[string]*sbom.Component))
		if err != nil {
			t.Errorf("BuildFromSBOM failed: %v", err)
		}

		// Should only have one node (the one with BOMRef)
		if len(g.Nodes) != 1 {
			t.Errorf("Expected 1 node, got %d", len(g.Nodes))
		}

		if _, exists := g.Nodes["svc-with-ref"]; !exists {
			t.Error("Service with BOMRef should be in graph")
		}
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Run("getNodeName", func(t *testing.T) {
		// Component node
		compNode := &Node{
			ID:        "comp-id",
			Component: &sbom.Component{Name: "Component Name"},
		}
		if name := getNodeName(compNode); name != "Component Name" {
			t.Errorf("Expected 'Component Name', got '%s'", name)
		}

		// Service node
		svcNode := &Node{
			ID:      "svc-id",
			Service: &sbom.Service{Name: "Service Name"},
		}
		if name := getNodeName(svcNode); name != "Service Name" {
			t.Errorf("Expected 'Service Name', got '%s'", name)
		}

		// Node with neither (fallback to ID)
		emptyNode := &Node{ID: "node-id"}
		if name := getNodeName(emptyNode); name != "node-id" {
			t.Errorf("Expected 'node-id', got '%s'", name)
		}
	})

	t.Run("getNodeVersion", func(t *testing.T) {
		// Component with version
		compNode := &Node{
			Component: &sbom.Component{Version: "1.2.3"},
		}
		if ver := getNodeVersion(compNode); ver != "1.2.3" {
			t.Errorf("Expected '1.2.3', got '%s'", ver)
		}

		// Service with version
		svcNode := &Node{
			Service: &sbom.Service{Version: "2.0.0"},
		}
		if ver := getNodeVersion(svcNode); ver != "2.0.0" {
			t.Errorf("Expected '2.0.0', got '%s'", ver)
		}

		// Node with no version
		emptyNode := &Node{
			Component: &sbom.Component{},
		}
		if ver := getNodeVersion(emptyNode); ver != "" {
			t.Errorf("Expected empty string, got '%s'", ver)
		}
	})
}
