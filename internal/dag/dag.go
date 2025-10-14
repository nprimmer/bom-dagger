package dag

import (
	"fmt"

	"github.com/nprimmer/bom-dagger/internal/sbom"
)

// Node represents a node in the DAG
type Node struct {
	ID           string
	Component    *sbom.Component
	Service      *sbom.Service  // For CycloneDX 1.6 services
	Dependencies []*Node
	Dependents   []*Node
}

// Graph represents the dependency DAG
type Graph struct {
	Nodes map[string]*Node
	Roots []*Node // Components with no dependencies
}

// New creates a new Graph
func New() *Graph {
	return &Graph{
		Nodes: make(map[string]*Node),
		Roots: []*Node{},
	}
}

// BuildFromSBOM builds a DAG from a CycloneDX SBOM
func (g *Graph) BuildFromSBOM(bom *sbom.CycloneDX, componentMap map[string]*sbom.Component) error {
	// Create nodes for all components
	for ref, component := range componentMap {
		node := &Node{
			ID:           ref,
			Component:    component,
			Dependencies: []*Node{},
			Dependents:   []*Node{},
		}
		g.Nodes[ref] = node
	}

	// Create nodes for all services (CycloneDX 1.6)
	for i := range bom.Services {
		service := &bom.Services[i]
		if service.BOMRef != "" {
			node := &Node{
				ID:           service.BOMRef,
				Service:      service,
				Dependencies: []*Node{},
				Dependents:   []*Node{},
			}
			g.Nodes[service.BOMRef] = node
		}
	}

	// Build dependency relationships
	for _, dep := range bom.Dependencies {
		node, exists := g.Nodes[dep.Ref]
		if !exists {
			// Skip dependencies for components not in our map
			continue
		}

		for _, depRef := range dep.DependsOn {
			depNode, exists := g.Nodes[depRef]
			if !exists {
				// Skip missing dependencies
				continue
			}

			// Add edge from node to dependency
			node.Dependencies = append(node.Dependencies, depNode)
			// Add reverse edge for dependents
			depNode.Dependents = append(depNode.Dependents, node)
		}
	}

	// Identify root nodes (components with no dependencies)
	for _, node := range g.Nodes {
		if len(node.Dependencies) == 0 {
			g.Roots = append(g.Roots, node)
		}
	}

	// Check for cycles
	if g.hasCycle() {
		return fmt.Errorf("dependency graph contains cycles")
	}

	return nil
}

// hasCycle detects if the graph has any cycles using DFS
func (g *Graph) hasCycle() bool {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for id := range g.Nodes {
		if g.hasCycleDFS(id, visited, recStack) {
			return true
		}
	}

	return false
}

// hasCycleDFS is a helper function for cycle detection
func (g *Graph) hasCycleDFS(nodeID string, visited, recStack map[string]bool) bool {
	visited[nodeID] = true
	recStack[nodeID] = true

	node := g.Nodes[nodeID]
	for _, dep := range node.Dependencies {
		if !visited[dep.ID] {
			if g.hasCycleDFS(dep.ID, visited, recStack) {
				return true
			}
		} else if recStack[dep.ID] {
			return true
		}
	}

	recStack[nodeID] = false
	return false
}

// GetNodeCount returns the number of nodes in the graph
func (g *Graph) GetNodeCount() int {
	return len(g.Nodes)
}

// GetEdgeCount returns the number of edges in the graph
func (g *Graph) GetEdgeCount() int {
	count := 0
	for _, node := range g.Nodes {
		count += len(node.Dependencies)
	}
	return count
}