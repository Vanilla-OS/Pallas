package parser

// EntityInfo contains relevant information about each entity in the package
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
	Methods         []MethodInfo
	Implements      []ImplementationInfo
	Package         string

	// Raw fields
	DescriptionRaw     string
	DeprecationNoteRaw string
}

// FieldInfo contains relevant information about each field in a struct
type FieldInfo struct {
	Name string
	Type string
	Tag  string
}

// MethodInfo contains relevant information about each method in an interface
type MethodInfo struct {
	Name       string
	Parameters []string
	Returns    []string
}

// ImplementationInfo contains information about an implemented interface
type ImplementationInfo struct {
	InterfaceName string
	Package       string
}
