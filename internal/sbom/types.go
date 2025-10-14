package sbom

// CycloneDX represents a CycloneDX SBOM document (supports 1.6)
type CycloneDX struct {
	BOMFormat    string         `json:"bomFormat"`
	SpecVersion  string         `json:"specVersion"`
	SerialNumber string         `json:"serialNumber,omitempty"`
	Version      int            `json:"version"`
	Metadata     *Metadata      `json:"metadata,omitempty"`
	Components   []Component    `json:"components"`
	Services     []Service      `json:"services,omitempty"`
	Dependencies []Dependency   `json:"dependencies,omitempty"`
	Compositions []Composition  `json:"compositions,omitempty"`
}

// Metadata contains metadata about the BOM
type Metadata struct {
	Timestamp   string      `json:"timestamp"`
	Authors     []Author    `json:"authors,omitempty"`
	Component   *Component  `json:"component,omitempty"`
	Supplier    *Supplier   `json:"supplier,omitempty"`
	Tools       []Tool      `json:"tools,omitempty"`
}

// Author represents an author
type Author struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

// Supplier represents a supplier
type Supplier struct {
	Name string `json:"name"`
	URL  []string `json:"url,omitempty"`
}

// Tool represents a tool used to create the BOM
type Tool struct {
	Vendor  string `json:"vendor,omitempty"`
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// Component represents a component in the BOM
type Component struct {
	Type        string       `json:"type"`
	BOMRef      string       `json:"bom-ref"`
	Name        string       `json:"name"`
	Version     string       `json:"version"`
	Description string       `json:"description,omitempty"`
	Scope       string       `json:"scope,omitempty"`
	Group       string       `json:"group,omitempty"`
	Purl        string       `json:"purl,omitempty"`
	Components  []Component  `json:"components,omitempty"`
	Properties  []Property   `json:"properties,omitempty"`
}

// Service represents a service in CycloneDX 1.6
type Service struct {
	BOMRef      string       `json:"bom-ref"`
	Name        string       `json:"name"`
	Version     string       `json:"version,omitempty"`
	Description string       `json:"description,omitempty"`
	Endpoints   []string     `json:"endpoints,omitempty"`
	Properties  []Property   `json:"properties,omitempty"`
}

// Property represents a key-value property
type Property struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Dependency represents a dependency relationship
type Dependency struct {
	Ref       string   `json:"ref"`
	DependsOn []string `json:"dependsOn,omitempty"`
}

// Composition represents component composition in CycloneDX 1.6
type Composition struct {
	Aggregate    string   `json:"aggregate"`
	Assemblies   []string `json:"assemblies,omitempty"`
	Dependencies []string `json:"dependencies,omitempty"`
}