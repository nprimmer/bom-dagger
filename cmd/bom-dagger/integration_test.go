package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Helper to run the tool with arguments and capture output
func runBomDagger(t *testing.T, args ...string) (string, string, error) {
	cmd := exec.Command("go", append([]string{"run", "main.go"}, args...)...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func TestIntegrationSimpleSBOM(t *testing.T) {
	sbomPath := filepath.Join("..", "..", "testdata", "sboms", "simple-1.6.json")

	tests := []struct {
		name     string
		args     []string
		wantOut  []string
		wantErr  bool
	}{
		{
			name: "deployment order",
			args: []string{"-i", sbomPath},
			wantOut: []string{
				"Deployment Order",
				"Component C",
				"Component B",
				"Component A",
			},
		},
		{
			name: "reverse order",
			args: []string{"-i", sbomPath, "-r"},
			wantOut: []string{
				"Teardown Order",
				"Component A",
				"Component B",
				"Component C",
			},
		},
		{
			name: "deployment groups",
			args: []string{"-i", sbomPath, "-g"},
			wantOut: []string{
				"Deployment Groups",
				"Component C",
				"Component B",
				"Component A",
			},
		},
		{
			name: "statistics",
			args: []string{"-i", sbomPath, "-s"},
			wantOut: []string{
				"Graph Statistics",
				"Total Components: 3",
				"Total Dependencies: 3",
			},
		},
		{
			name: "DOT output",
			args: []string{"-i", sbomPath, "-o", "dot"},
			wantOut: []string{
				"digraph dependencies",
				"comp-a",
				"comp-b",
				"comp-c",
				"->",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runBomDagger(t, tt.args...)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v\nStderr: %s", err, stderr)
				}
			}

			for _, want := range tt.wantOut {
				if !strings.Contains(stdout, want) {
					t.Errorf("Output missing '%s'\nGot: %s", want, stdout)
				}
			}
		})
	}
}

func TestIntegrationServicesSBOM(t *testing.T) {
	sbomPath := filepath.Join("..", "..", "testdata", "sboms", "services-1.6.json")

	stdout, stderr, err := runBomDagger(t, "-i", sbomPath)
	if err != nil {
		t.Fatalf("Failed to process services SBOM: %v\nStderr: %s", err, stderr)
	}

	// Check that services are included in output
	expectedServices := []string{
		"REST API Service",
		"Authentication Service",
		"Redis Cache Service",
	}

	for _, svc := range expectedServices {
		if !strings.Contains(stdout, svc) {
			t.Errorf("Service '%s' not found in output", svc)
		}
	}

	// Check that components are also included
	expectedComponents := []string{
		"Frontend Application",
		"PostgreSQL",
	}

	for _, comp := range expectedComponents {
		if !strings.Contains(stdout, comp) {
			t.Errorf("Component '%s' not found in output", comp)
		}
	}
}

func TestIntegrationNestedSBOM(t *testing.T) {
	sbomPath := filepath.Join("..", "..", "testdata", "sboms", "nested-1.6.json")

	stdout, stderr, err := runBomDagger(t, "-i", sbomPath)
	if err != nil {
		t.Fatalf("Failed to process nested SBOM: %v\nStderr: %s", err, stderr)
	}

	// Check that nested components are included
	expectedComponents := []string{
		"Main Application",
		"Embedded Library",
		"Module A",
		"Submodule A1",
		"Submodule A2",
		"Module B",
		"Submodule B1",
		"Shared Library",
	}

	for _, comp := range expectedComponents {
		if !strings.Contains(stdout, comp) {
			t.Errorf("Nested component '%s' not found in output", comp)
		}
	}
}

func TestIntegrationCycleSBOM(t *testing.T) {
	sbomPath := filepath.Join("..", "..", "testdata", "sboms", "cycle-1.6.json")

	_, stderr, err := runBomDagger(t, "-i", sbomPath)

	if err == nil {
		t.Error("Expected error for cyclic dependencies")
	}

	if !strings.Contains(stderr, "cycle") {
		t.Errorf("Expected cycle error message in stderr, got: %s", stderr)
	}
}

func TestIntegrationNoDependencies(t *testing.T) {
	sbomPath := filepath.Join("..", "..", "testdata", "sboms", "no-deps-1.6.json")

	stdout, stderr, err := runBomDagger(t, "-i", sbomPath, "-g")
	if err != nil {
		t.Fatalf("Failed to process no-deps SBOM: %v\nStderr: %s", err, stderr)
	}

	// All components should be in the same group (can deploy in parallel)
	if !strings.Contains(stdout, "Group 1") {
		t.Error("Expected all components in Group 1")
	}

	expectedComponents := []string{
		"Standalone Component 1",
		"Standalone Component 2",
		"Standalone Component 3",
	}

	for _, comp := range expectedComponents {
		if !strings.Contains(stdout, comp) {
			t.Errorf("Component '%s' not found in output", comp)
		}
	}
}

func TestIntegrationMicroservices(t *testing.T) {
	sbomPath := filepath.Join("..", "..", "testdata", "sboms", "microservices-1.6.json")

	tests := []struct {
		name     string
		args     []string
		validate func(t *testing.T, stdout, stderr string)
	}{
		{
			name: "deployment order with stats",
			args: []string{"-i", sbomPath, "-s"},
			validate: func(t *testing.T, stdout, stderr string) {
				// Check statistics
				if !strings.Contains(stdout, "Total Components:") {
					t.Error("Missing component count")
				}
				if !strings.Contains(stdout, "Total Dependencies:") {
					t.Error("Missing dependency count")
				}

				// Infrastructure should be deployed first
				infraComponents := []string{
					"Zookeeper",
					"PostgreSQL Primary",
					"MongoDB",
					"Redis Master",
					"Elasticsearch",
					"Prometheus",
				}

				output := stdout
				for _, comp := range infraComponents {
					if !strings.Contains(output, comp) {
						t.Errorf("Infrastructure component '%s' not found", comp)
					}
				}
			},
		},
		{
			name: "parallel deployment groups",
			args: []string{"-i", sbomPath, "-g"},
			validate: func(t *testing.T, stdout, stderr string) {
				// Check for multiple deployment groups
				if !strings.Contains(stdout, "Group 1") {
					t.Error("Missing Group 1")
				}
				if !strings.Contains(stdout, "Group 2") {
					t.Error("Missing Group 2")
				}

				// Infrastructure should be in early groups
				if !strings.Contains(stdout, "can deploy in parallel") {
					t.Error("Missing parallel deployment indication")
				}
			},
		},
		{
			name: "DOT visualization",
			args: []string{"-i", sbomPath, "-o", "dot"},
			validate: func(t *testing.T, stdout, stderr string) {
				// Check DOT format elements
				if !strings.Contains(stdout, "digraph dependencies") {
					t.Error("Missing DOT graph declaration")
				}
				if !strings.Contains(stdout, "rankdir=BT") {
					t.Error("Missing rankdir setting")
				}
				if !strings.Contains(stdout, "[label=") {
					t.Error("Missing node labels")
				}
				if !strings.Contains(stdout, "->") {
					t.Error("Missing edge declarations")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runBomDagger(t, tt.args...)
			if err != nil {
				t.Fatalf("Unexpected error: %v\nStderr: %s", err, stderr)
			}

			tt.validate(t, stdout, stderr)
		})
	}
}

func TestIntegrationInvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "missing input file",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "non-existent file",
			args:    []string{"-i", "/non/existent/file.json"},
			wantErr: true,
			errMsg:  "Error parsing SBOM",
		},
		{
			name:    "help flag",
			args:    []string{"-h"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, stderr, err := runBomDagger(t, tt.args...)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errMsg != "" && !strings.Contains(stderr, tt.errMsg) {
					t.Errorf("Expected error message containing '%s', got: %s", tt.errMsg, stderr)
				}
			}
		})
	}
}

func TestIntegrationExampleSBOM(t *testing.T) {
	// Test with the original example SBOM (updated to 1.6)
	sbomPath := filepath.Join("..", "..", "example-sbom.json")

	// Check if file exists
	if _, err := os.Stat(sbomPath); os.IsNotExist(err) {
		t.Skip("example-sbom.json not found")
	}

	stdout, stderr, err := runBomDagger(t, "-i", sbomPath)
	if err != nil {
		t.Fatalf("Failed to process example SBOM: %v\nStderr: %s", err, stderr)
	}

	// Verify the output contains expected components
	expectedComponents := []string{
		"React Frontend",
		"API Gateway",
		"PostgreSQL Database",
		"Redis Cache",
		"RabbitMQ",
		"Nginx Load Balancer",
	}

	for _, comp := range expectedComponents {
		if !strings.Contains(stdout, comp) {
			t.Errorf("Component '%s' not found in output", comp)
		}
	}
}

func BenchmarkProcessSimpleSBOM(b *testing.B) {
	sbomPath := filepath.Join("..", "..", "testdata", "sboms", "simple-1.6.json")

	for i := 0; i < b.N; i++ {
		cmd := exec.Command("go", "run", "main.go", "-i", sbomPath)
		_ = cmd.Run()
	}
}

func BenchmarkProcessMicroservicesSBOM(b *testing.B) {
	sbomPath := filepath.Join("..", "..", "testdata", "sboms", "microservices-1.6.json")

	for i := 0; i < b.N; i++ {
		cmd := exec.Command("go", "run", "main.go", "-i", sbomPath)
		_ = cmd.Run()
	}
}