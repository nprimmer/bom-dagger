package dag

import (
	"fmt"
)

// DeploymentOrder represents a deployment step
type DeploymentOrder struct {
	Step      int
	Component string
	BOMRef    string
}

// TopologicalSort performs a topological sort using Kahn's algorithm
// Returns the deployment order (components with no dependencies first)
func (g *Graph) TopologicalSort() ([]DeploymentOrder, error) {
	// Create a copy of in-degrees
	inDegree := make(map[string]int)
	for id, node := range g.Nodes {
		inDegree[id] = len(node.Dependencies)
	}

	// Queue for nodes with no dependencies
	queue := []*Node{}
	for _, node := range g.Nodes {
		if inDegree[node.ID] == 0 {
			queue = append(queue, node)
		}
	}

	var result []DeploymentOrder
	step := 1

	for len(queue) > 0 {
		// Process all nodes at the current level
		levelSize := len(queue)
		levelNodes := queue[:levelSize]
		queue = queue[levelSize:]

		// Add all nodes at this level to the result
		for _, node := range levelNodes {
			name := getNodeName(node)
			result = append(result, DeploymentOrder{
				Step:      step,
				Component: name,
				BOMRef:    node.ID,
			})

			// Reduce in-degree for dependent nodes
			for _, dependent := range node.Dependents {
				inDegree[dependent.ID]--
				if inDegree[dependent.ID] == 0 {
					queue = append(queue, dependent)
				}
			}
		}

		step++
	}

	// Check if all nodes were processed
	if len(result) != len(g.Nodes) {
		return nil, fmt.Errorf("cycle detected in dependency graph")
	}

	return result, nil
}

// GetDeploymentGroups returns components grouped by deployment order
// Components in the same group can be deployed in parallel
func (g *Graph) GetDeploymentGroups() ([][]string, error) {
	// Create a copy of in-degrees
	inDegree := make(map[string]int)
	for id, node := range g.Nodes {
		inDegree[id] = len(node.Dependencies)
	}

	// Queue for nodes with no dependencies
	queue := []*Node{}
	for _, node := range g.Nodes {
		if inDegree[node.ID] == 0 {
			queue = append(queue, node)
		}
	}

	var groups [][]string
	processedCount := 0

	for len(queue) > 0 {
		// Process all nodes at the current level
		levelSize := len(queue)
		levelNodes := queue[:levelSize]
		queue = queue[levelSize:]

		// Create a group for this level
		group := make([]string, 0, levelSize)
		for _, node := range levelNodes {
			name := getNodeName(node)
			version := getNodeVersion(node)
			if version != "" {
				group = append(group, fmt.Sprintf("%s (%s)", name, version))
			} else {
				group = append(group, name)
			}
			processedCount++

			// Reduce in-degree for dependent nodes
			for _, dependent := range node.Dependents {
				inDegree[dependent.ID]--
				if inDegree[dependent.ID] == 0 {
					queue = append(queue, dependent)
				}
			}
		}

		groups = append(groups, group)
	}

	// Check if all nodes were processed
	if processedCount != len(g.Nodes) {
		return nil, fmt.Errorf("cycle detected in dependency graph")
	}

	return groups, nil
}

// ReverseTopologicalSort returns the reverse deployment order (teardown order)
func (g *Graph) ReverseTopologicalSort() ([]DeploymentOrder, error) {
	order, err := g.TopologicalSort()
	if err != nil {
		return nil, err
	}

	// Reverse the order
	for i, j := 0, len(order)-1; i < j; i, j = i+1, j-1 {
		order[i], order[j] = order[j], order[i]
	}

	// Update step numbers
	for i := range order {
		order[i].Step = i + 1
	}

	return order, nil
}

// Helper functions to get node name and version for both components and services
func getNodeName(node *Node) string {
	if node.Component != nil {
		return node.Component.Name
	}
	if node.Service != nil {
		return node.Service.Name
	}
	return node.ID
}

func getNodeVersion(node *Node) string {
	if node.Component != nil {
		return node.Component.Version
	}
	if node.Service != nil {
		return node.Service.Version
	}
	return ""
}
