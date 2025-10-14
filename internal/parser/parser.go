package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/nprimmer/bom-dagger/internal/sbom"
)

// Parser handles parsing of CycloneDX SBOM files
type Parser struct{}

// New creates a new Parser instance
func New() *Parser {
	return &Parser{}
}

// ParseFile parses a CycloneDX SBOM from a file path
func (p *Parser) ParseFile(filePath string) (*sbom.CycloneDX, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return p.Parse(file)
}

// Parse parses a CycloneDX SBOM from a reader
func (p *Parser) Parse(reader io.Reader) (*sbom.CycloneDX, error) {
	var bom sbom.CycloneDX

	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&bom); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	// Validate the BOM format
	if bom.BOMFormat != "CycloneDX" {
		return nil, fmt.Errorf("invalid BOM format: %s (expected CycloneDX)", bom.BOMFormat)
	}

	return &bom, nil
}

// GetComponentMap creates a map of component references to components
func (p *Parser) GetComponentMap(bom *sbom.CycloneDX) map[string]*sbom.Component {
	componentMap := make(map[string]*sbom.Component)

	// Add all components to the map
	for i := range bom.Components {
		addComponentToMap(&bom.Components[i], componentMap)
	}

	// Add metadata component if present
	if bom.Metadata != nil && bom.Metadata.Component != nil {
		addComponentToMap(bom.Metadata.Component, componentMap)
	}

	return componentMap
}

// GetServiceMap creates a map of service references to services (CycloneDX 1.6)
func (p *Parser) GetServiceMap(bom *sbom.CycloneDX) map[string]*sbom.Service {
	serviceMap := make(map[string]*sbom.Service)

	for i := range bom.Services {
		if bom.Services[i].BOMRef != "" {
			serviceMap[bom.Services[i].BOMRef] = &bom.Services[i]
		}
	}

	return serviceMap
}

// addComponentToMap recursively adds components to the map
func addComponentToMap(component *sbom.Component, componentMap map[string]*sbom.Component) {
	if component.BOMRef != "" {
		componentMap[component.BOMRef] = component
	}

	// Recursively add nested components
	for i := range component.Components {
		addComponentToMap(&component.Components[i], componentMap)
	}
}
