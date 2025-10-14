package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/nprimmer/bom-dagger/internal/dag"
	"github.com/nprimmer/bom-dagger/internal/parser"
	"github.com/nprimmer/bom-dagger/internal/sbom"
)

// Version is set at build time via -ldflags
var Version = "dev"

func main() {
	var (
		inputFile   string
		outputMode  string
		showReverse bool
		showGroups  bool
		showStats   bool
		showHelp    bool
		showVersion bool
	)

	flag.StringVar(&inputFile, "input", "", "Path to CycloneDX SBOM file (JSON)")
	flag.StringVar(&inputFile, "i", "", "Path to CycloneDX SBOM file (JSON) (shorthand)")
	flag.StringVar(&outputMode, "output", "order", "Output mode: order, groups, dot")
	flag.StringVar(&outputMode, "o", "order", "Output mode: order, groups, dot (shorthand)")
	flag.BoolVar(&showReverse, "reverse", false, "Show reverse order (teardown sequence)")
	flag.BoolVar(&showReverse, "r", false, "Show reverse order (teardown sequence) (shorthand)")
	flag.BoolVar(&showGroups, "groups", false, "Show deployment groups (components that can be deployed in parallel)")
	flag.BoolVar(&showGroups, "g", false, "Show deployment groups (shorthand)")
	flag.BoolVar(&showStats, "stats", false, "Show graph statistics")
	flag.BoolVar(&showStats, "s", false, "Show graph statistics (shorthand)")
	flag.BoolVar(&showHelp, "help", false, "Show help message")
	flag.BoolVar(&showHelp, "h", false, "Show help message (shorthand)")
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.BoolVar(&showVersion, "v", false, "Show version information (shorthand)")

	flag.Parse()

	if showVersion {
		printVersion()
		os.Exit(0)
	}

	if showHelp || inputFile == "" {
		printUsage()
		if inputFile == "" && !showHelp {
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Parse the SBOM file
	p := parser.New()
	bom, err := p.ParseFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing SBOM: %v\n", err)
		os.Exit(1)
	}

	// Get component map
	componentMap := p.GetComponentMap(bom)

	// Build the DAG
	graph := dag.New()
	if err := graph.BuildFromSBOM(bom, componentMap); err != nil {
		fmt.Fprintf(os.Stderr, "Error building DAG: %v\n", err)
		os.Exit(1)
	}

	// Show statistics if requested
	if showStats {
		printStatistics(graph, bom)
		fmt.Println()
	}

	// Handle different output modes
	if showGroups || outputMode == "groups" {
		printDeploymentGroups(graph)
	} else if outputMode == "dot" {
		printDotFormat(graph)
	} else {
		// Default: show deployment order
		if showReverse {
			printReverseOrder(graph)
		} else {
			printDeploymentOrder(graph)
		}
	}
}

func printVersion() {
	fmt.Printf("bom-dagger version %s\n", Version)
	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}

func printUsage() {
	fmt.Printf("bom-dagger %s - Creates a DAG for deployment order from a CycloneDX SBOM\n", Version)
	fmt.Println()
	fmt.Println("Usage: bom-dagger -i <sbom-file> [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -i, --input <file>     Path to CycloneDX SBOM file (JSON)")
	fmt.Println("  -o, --output <mode>    Output mode: order (default), groups, dot")
	fmt.Println("  -r, --reverse          Show reverse order (teardown sequence)")
	fmt.Println("  -g, --groups           Show deployment groups (parallel deployment)")
	fmt.Println("  -s, --stats            Show graph statistics")
	fmt.Println("  -h, --help             Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  bom-dagger -i sbom.json                    # Show deployment order")
	fmt.Println("  bom-dagger -i sbom.json -r                 # Show teardown order")
	fmt.Println("  bom-dagger -i sbom.json -g                 # Show parallel groups")
	fmt.Println("  bom-dagger -i sbom.json -o dot > graph.dot # Generate DOT format")
}

func printStatistics(graph *dag.Graph, bom *sbom.CycloneDX) {
	fmt.Println("=== Graph Statistics ===")
	fmt.Printf("Total Components: %d\n", graph.GetNodeCount())
	fmt.Printf("Total Dependencies: %d\n", graph.GetEdgeCount())
	fmt.Printf("Root Components: %d\n", len(graph.Roots))
	fmt.Printf("SBOM Format: %s %s\n", bom.BOMFormat, bom.SpecVersion)
}

func printDeploymentOrder(graph *dag.Graph) {
	order, err := graph.TopologicalSort()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error computing deployment order: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== Deployment Order ===")
	fmt.Println("Deploy components in this sequence:")
	fmt.Println()

	currentStep := 0
	for _, item := range order {
		if item.Step != currentStep {
			if currentStep > 0 {
				fmt.Println()
			}
			currentStep = item.Step
			fmt.Printf("Step %d:\n", currentStep)
		}
		fmt.Printf("  - %s (ref: %s)\n", item.Component, item.BOMRef)
	}
}

func printReverseOrder(graph *dag.Graph) {
	order, err := graph.ReverseTopologicalSort()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error computing reverse order: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== Teardown Order ===")
	fmt.Println("Remove/stop components in this sequence:")
	fmt.Println()

	currentStep := 0
	for _, item := range order {
		if item.Step != currentStep {
			if currentStep > 0 {
				fmt.Println()
			}
			currentStep = item.Step
			fmt.Printf("Step %d:\n", currentStep)
		}
		fmt.Printf("  - %s (ref: %s)\n", item.Component, item.BOMRef)
	}
}

func printDeploymentGroups(graph *dag.Graph) {
	groups, err := graph.GetDeploymentGroups()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error computing deployment groups: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== Deployment Groups ===")
	fmt.Println("Components in the same group can be deployed in parallel:")
	fmt.Println()

	for i, group := range groups {
		fmt.Printf("Group %d (can deploy in parallel):\n", i+1)
		for _, component := range group {
			fmt.Printf("  - %s\n", component)
		}
		if i < len(groups)-1 {
			fmt.Println("    ↓")
		}
	}
}

func printDotFormat(graph *dag.Graph) {
	fmt.Println("digraph dependencies {")
	fmt.Println("  rankdir=BT;")
	fmt.Println("  node [shape=box];")
	fmt.Println()

	// Print all nodes
	for id, node := range graph.Nodes {
		var label string
		if node.Component != nil {
			label = fmt.Sprintf("%s\\n%s", node.Component.Name, node.Component.Version)
		} else if node.Service != nil {
			label = fmt.Sprintf("%s\\n%s", node.Service.Name, node.Service.Version)
			if node.Service.Version == "" {
				label = node.Service.Name
			}
		} else {
			label = id
		}
		fmt.Printf("  \"%s\" [label=\"%s\"];\n", id, label)
	}
	fmt.Println()

	// Print all edges
	for _, node := range graph.Nodes {
		for _, dep := range node.Dependencies {
			fmt.Printf("  \"%s\" -> \"%s\";\n", node.ID, dep.ID)
		}
	}

	fmt.Println("}")
}