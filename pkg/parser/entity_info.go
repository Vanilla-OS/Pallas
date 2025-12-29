package parser

// EntityInfo represents a Go entity such as a function, struct, interface, or type alias.
type EntityInfo struct {
	Name            string
	Description     string
	Example         string
	Notes           string
	DeprecationNote string
	Parameters      []string
	Returns         []string
	Body            string
	Type            string // "function", "struct", "interface", "type"
	Fields          []FieldInfo
	Methods         []EntityInfo
	Implements      []ImplementationInfo
	Package         string
	PackageURL      string
	PackagePath     string
	References      []ReferenceInfo

	DescriptionRaw     string
	DeprecationNoteRaw string

	Signature string
	File      string
	LineStart int
	LineEnd   int
}

// ReferenceInfo contains details about a type referenced by an entity.
type ReferenceInfo struct {
	Name        string
	Package     string
	PackageURL  string
	PackagePath string
}

// FieldInfo represents a single field within a struct.
type FieldInfo struct {
	Name string
	Type string
	Tag  string
}

// ImplementationInfo identifies an interface implemented by a struct.
type ImplementationInfo struct {
	InterfaceName string
	Package       string
}

// ImportInfo captures details about a package import.
type ImportInfo struct {
	URL     string
	Path    string
	Alias   string
	Doc     string
	Comment string
	Package string
	File    string
}
