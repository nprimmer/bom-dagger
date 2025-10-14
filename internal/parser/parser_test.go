package parser

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nprimmer/bom-dagger/internal/sbom"
)

func TestNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Error("New() returned nil")
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid simple SBOM",
			json: `{
				"bomFormat": "CycloneDX",
				"specVersion": "1.6",
				"version": 1,
				"components": []
			}`,
			wantErr: false,
		},
		{
			name: "invalid BOM format",
			json: `{
				"bomFormat": "InvalidFormat",
				"specVersion": "1.6",
				"version": 1,
				"components": []
			}`,
			wantErr: true,
			errMsg:  "invalid BOM format",
		},
		{
			name:    "invalid JSON",
			json:    `{"invalid json`,
			wantErr: true,
			errMsg:  "failed to decode JSON",
		},
		{
			name:    "empty input",
			json:    ``,
			wantErr: true,
			errMsg:  "failed to decode JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New()
			reader := strings.NewReader(tt.json)
			bom, err := p.Parse(reader)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if bom == nil {
					t.Error("Expected non-nil BOM")
				}
			}
		})
	}
}

func TestParseFile(t *testing.T) {
	// Test with valid file
	t.Run("valid file", func(t *testing.T) {
		p := New()
		testFile := filepath.Join("..", "..", "testdata", "sboms", "simple-1.6.json")

		bom, err := p.ParseFile(testFile)
		if err != nil {
			t.Fatalf("ParseFile failed: %v", err)
		}

		if bom.BOMFormat != "CycloneDX" {
			t.Errorf("Expected BOMFormat 'CycloneDX', got '%s'", bom.BOMFormat)
		}
		if bom.SpecVersion != "1.6" {
			t.Errorf("Expected SpecVersion '1.6', got '%s'", bom.SpecVersion)
		}
		if len(bom.Components) != 3 {
			t.Errorf("Expected 3 components, got %d", len(bom.Components))
		}
	})

	// Test with non-existent file
	t.Run("non-existent file", func(t *testing.T) {
		p := New()
		_, err := p.ParseFile("/non/existent/file.json")
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})
}

func TestGetComponentMap(t *testing.T) {
	tests := []struct {
		name          string
		bom           *sbom.CycloneDX
		expectedCount int
	}{
		{
			name: "simple components",
			bom: &sbom.CycloneDX{
				Components: []sbom.Component{
					{BOMRef: "comp-1", Name: "Component 1"},
					{BOMRef: "comp-2", Name: "Component 2"},
					{BOMRef: "comp-3", Name: "Component 3"},
				},
			},
			expectedCount: 3,
		},
		{
			name: "nested components",
			bom: &sbom.CycloneDX{
				Components: []sbom.Component{
					{
						BOMRef: "parent",
						Name:   "Parent",
						Components: []sbom.Component{
							{BOMRef: "child-1", Name: "Child 1"},
							{BOMRef: "child-2", Name: "Child 2"},
						},
					},
				},
			},
			expectedCount: 3,
		},
		{
			name: "metadata component",
			bom: &sbom.CycloneDX{
				Metadata: &sbom.Metadata{
					Component: &sbom.Component{
						BOMRef: "metadata-comp",
						Name:   "Metadata Component",
					},
				},
				Components: []sbom.Component{
					{BOMRef: "regular-comp", Name: "Regular Component"},
				},
			},
			expectedCount: 2,
		},
		{
			name: "components without BOMRef",
			bom: &sbom.CycloneDX{
				Components: []sbom.Component{
					{BOMRef: "comp-1", Name: "Component 1"},
					{Name: "Component without ref"}, // No BOMRef
					{BOMRef: "comp-3", Name: "Component 3"},
				},
			},
			expectedCount: 2,
		},
		{
			name:          "empty BOM",
			bom:           &sbom.CycloneDX{},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New()
			componentMap := p.GetComponentMap(tt.bom)

			if len(componentMap) != tt.expectedCount {
				t.Errorf("Expected %d components in map, got %d", tt.expectedCount, len(componentMap))
			}

			// Verify that components are correctly mapped by their BOMRef
			for ref, comp := range componentMap {
				if comp.BOMRef != ref {
					t.Errorf("Component BOMRef mismatch: map key '%s' != component ref '%s'", ref, comp.BOMRef)
				}
			}
		})
	}
}

func TestGetServiceMap(t *testing.T) {
	tests := []struct {
		name          string
		bom           *sbom.CycloneDX
		expectedCount int
	}{
		{
			name: "multiple services",
			bom: &sbom.CycloneDX{
				Services: []sbom.Service{
					{BOMRef: "svc-1", Name: "Service 1"},
					{BOMRef: "svc-2", Name: "Service 2"},
					{BOMRef: "svc-3", Name: "Service 3"},
				},
			},
			expectedCount: 3,
		},
		{
			name: "services without BOMRef",
			bom: &sbom.CycloneDX{
				Services: []sbom.Service{
					{BOMRef: "svc-1", Name: "Service 1"},
					{Name: "Service without ref"}, // No BOMRef
					{BOMRef: "svc-3", Name: "Service 3"},
				},
			},
			expectedCount: 2,
		},
		{
			name:          "no services",
			bom:           &sbom.CycloneDX{},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New()
			serviceMap := p.GetServiceMap(tt.bom)

			if len(serviceMap) != tt.expectedCount {
				t.Errorf("Expected %d services in map, got %d", tt.expectedCount, len(serviceMap))
			}

			// Verify that services are correctly mapped by their BOMRef
			for ref, svc := range serviceMap {
				if svc.BOMRef != ref {
					t.Errorf("Service BOMRef mismatch: map key '%s' != service ref '%s'", ref, svc.BOMRef)
				}
			}
		})
	}
}

func TestParseCycloneDX16Features(t *testing.T) {
	json := `{
		"bomFormat": "CycloneDX",
		"specVersion": "1.6",
		"serialNumber": "urn:uuid:test-serial",
		"version": 1,
		"metadata": {
			"timestamp": "2024-01-15T10:00:00Z",
			"authors": [
				{"name": "Test Author", "email": "test@example.com"}
			],
			"supplier": {
				"name": "Test Supplier",
				"url": ["https://example.com"]
			},
			"tools": [
				{"vendor": "Test Vendor", "name": "Test Tool", "version": "1.0"}
			]
		},
		"components": [
			{
				"type": "library",
				"bom-ref": "comp-1",
				"name": "Component 1",
				"version": "1.0.0",
				"description": "Test component",
				"scope": "required",
				"group": "com.example",
				"properties": [
					{"name": "key1", "value": "value1"}
				]
			}
		],
		"services": [
			{
				"bom-ref": "svc-1",
				"name": "Service 1",
				"version": "2.0.0",
				"description": "Test service",
				"endpoints": ["https://api.example.com"],
				"properties": [
					{"name": "key2", "value": "value2"}
				]
			}
		],
		"compositions": [
			{
				"aggregate": "complete",
				"assemblies": ["comp-1"],
				"dependencies": ["svc-1"]
			}
		]
	}`

	p := New()
	reader := strings.NewReader(json)
	bom, err := p.Parse(reader)
	if err != nil {
		t.Fatalf("Failed to parse CycloneDX 1.6: %v", err)
	}

	// Verify 1.6 specific fields
	if bom.SerialNumber != "urn:uuid:test-serial" {
		t.Errorf("Expected SerialNumber 'urn:uuid:test-serial', got '%s'", bom.SerialNumber)
	}

	// Verify metadata
	if bom.Metadata == nil {
		t.Fatal("Expected metadata to be present")
	}
	if len(bom.Metadata.Authors) != 1 {
		t.Errorf("Expected 1 author, got %d", len(bom.Metadata.Authors))
	}
	if bom.Metadata.Supplier == nil || bom.Metadata.Supplier.Name != "Test Supplier" {
		t.Error("Supplier not parsed correctly")
	}
	if len(bom.Metadata.Tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(bom.Metadata.Tools))
	}

	// Verify component properties
	if len(bom.Components) != 1 {
		t.Fatalf("Expected 1 component, got %d", len(bom.Components))
	}
	comp := bom.Components[0]
	if comp.Description != "Test component" {
		t.Errorf("Expected description 'Test component', got '%s'", comp.Description)
	}
	if comp.Scope != "required" {
		t.Errorf("Expected scope 'required', got '%s'", comp.Scope)
	}
	if comp.Group != "com.example" {
		t.Errorf("Expected group 'com.example', got '%s'", comp.Group)
	}
	if len(comp.Properties) != 1 {
		t.Errorf("Expected 1 property, got %d", len(comp.Properties))
	}

	// Verify services
	if len(bom.Services) != 1 {
		t.Fatalf("Expected 1 service, got %d", len(bom.Services))
	}
	svc := bom.Services[0]
	if svc.Name != "Service 1" {
		t.Errorf("Expected service name 'Service 1', got '%s'", svc.Name)
	}
	if len(svc.Endpoints) != 1 {
		t.Errorf("Expected 1 endpoint, got %d", len(svc.Endpoints))
	}

	// Verify compositions
	if len(bom.Compositions) != 1 {
		t.Fatalf("Expected 1 composition, got %d", len(bom.Compositions))
	}
	comp0 := bom.Compositions[0]
	if comp0.Aggregate != "complete" {
		t.Errorf("Expected aggregate 'complete', got '%s'", comp0.Aggregate)
	}
}

func TestParseAllTestFiles(t *testing.T) {
	testDir := filepath.Join("..", "..", "testdata", "sboms")
	files, err := os.ReadDir(testDir)
	if err != nil {
		t.Skipf("Test data directory not found: %v", err)
	}

	p := New()

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		t.Run(file.Name(), func(t *testing.T) {
			filePath := filepath.Join(testDir, file.Name())
			bom, err := p.ParseFile(filePath)

			// Special handling for files expected to fail
			if strings.Contains(file.Name(), "cycle") {
				// Cycle detection happens in DAG building, not parsing
				if err != nil {
					t.Errorf("Parsing should succeed for cycle test: %v", err)
				}
				return
			}

			if err != nil {
				t.Errorf("Failed to parse %s: %v", file.Name(), err)
				return
			}

			// Basic validation
			if bom.BOMFormat != "CycloneDX" {
				t.Errorf("Invalid BOMFormat in %s: %s", file.Name(), bom.BOMFormat)
			}
			if bom.SpecVersion != "1.6" {
				t.Errorf("Expected SpecVersion 1.6 in %s, got %s", file.Name(), bom.SpecVersion)
			}
		})
	}
}

func BenchmarkParse(b *testing.B) {
	json := bytes.Repeat([]byte(`{
		"bomFormat": "CycloneDX",
		"specVersion": "1.6",
		"version": 1,
		"components": [
			{"bom-ref": "comp-1", "name": "Component 1", "version": "1.0.0", "type": "library"}
		]
	}`), 1)

	p := New()
	reader := bytes.NewReader(json)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = reader.Seek(0, 0)
		_, _ = p.Parse(reader)
	}
}
