package parser

// Information about each entity in the package
// (functions, types, interfaces)
type EntityInfo struct {
	Name            string
	Description     string
	Example         string
	Notes           string
	DeprecationNote string
	Parameters      []string
	Returns         []string
	Body            string
	Type            string
	Fields          []FieldInfo
	Methods         []EntityInfo
	Implements      []ImplementationInfo
	Package         string
	PackageURL      string
	PackagePath     string
	References      []ReferenceInfo

	// Raw fields
	DescriptionRaw     string
	DeprecationNoteRaw string
}

// Information about references used by an entity
type ReferenceInfo struct {
	Name        string
	Package     string
	PackageURL  string
	PackagePath string
}

// Information about each field in a struct
type FieldInfo struct {
	Name string
	Type string
	Tag  string
}

// Information about an implemented interface
type ImplementationInfo struct {
	InterfaceName string
	Package       string
}

// Information about an imported package
type ImportInfo struct {
	URL     string
	Path    string
	Alias   string
	Doc     string
	Comment string
}
