package parser

import (
	"go/ast"
	"go/format"
	"go/token"
	"html"
	"os"
	"strings"
)

// Extract methods from an interface declaration
//
// Returns: EntityInfo representing the methods of the interface
func extractMethods(interfaceType *ast.InterfaceType) []EntityInfo {
	var methods []EntityInfo
	for _, field := range interfaceType.Methods.List {
		if funcType, ok := field.Type.(*ast.FuncType); ok {
			methodInfo := EntityInfo{
				Name:       field.Names[0].Name,
				Parameters: extractParameters(funcType.Params),
				Returns:    extractParameters(funcType.Results),
			}
			methods = append(methods, methodInfo)
		}
	}
	return methods
}

// Extract fields from a struct
//
// Returns: FieldInfo representing the fields of the struct
func extractFields(structType *ast.StructType) []FieldInfo {
	var fields []FieldInfo
	for _, field := range structType.Fields.List {
		typeStr := formatExpr(field.Type)
		for _, name := range field.Names {
			fieldInfo := FieldInfo{
				Name: name.Name,
				Type: typeStr,
				Tag:  extractTag(field),
			}
			fields = append(fields, fieldInfo)
		}
	}
	return fields
}

// Extract a structs tags
//
// Returns: Tag value as a string or an empty string if no tag is present
func extractTag(field *ast.Field) string {
	if field.Tag != nil {
		return strings.Trim(field.Tag.Value, "`")
	}
	return ""
}

// Holds different parts of a function's documentation comment
type DescriptionData struct {
	Description     string
	Example         string
	Notes           string
	DeprecationNote string
	Returns         string

	// Raw fields
	DescriptionRaw     string
	DeprecationNoteRaw string
}

// Extract and format description data and example code from a
// documentation comment string
//
// Returns: A DescriptionData struct containing formatted and raw description,
// example, notes, deprecation note, and returns information
func extractDescriptionData(doc string) DescriptionData {
	lines := strings.Split(doc, "\n")

	var descLines []string
	var exampleLines []string
	var notesLines []string
	var deprecationNoteLines []string
	var returnsLines []string

	var description string
	var example string
	var notes string
	var deprecationNote string
	var returns string

	isExample := false
	isNotes := false
	isDeprecationNote := false
	isReturns := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Example:") {
			isExample = true
			isNotes = false
			isDeprecationNote = false
			isReturns = false
			continue
		}
		if strings.HasPrefix(line, "Notes:") {
			isNotes = true
			isExample = false
			isDeprecationNote = false
			isReturns = false
			continue
		}
		if strings.HasPrefix(line, "Deprecated:") {
			isDeprecationNote = true
			isExample = false
			isNotes = false
			isReturns = false
			continue
		}
		if strings.HasPrefix(line, "Returns:") {
			isReturns = true
			isExample = false
			isNotes = false
			isDeprecationNote = false
			continue
		}

		if isExample {
			exampleLines = append(exampleLines, line)
		} else if isNotes {
			notesLines = append(notesLines, line)
		} else if isDeprecationNote {
			deprecationNoteLines = append(deprecationNoteLines, line)
		} else if isReturns {
				returnsLines = append(returnsLines, line)
		} else {
			descLines = append(descLines, line)
		}
	}

	// Description
	descriptionRaw := strings.Join(descLines, "\n")
	description = strings.Join(descLines, "</p>\n<p>")
	description = "<p>" + description + "</p>"
	description = strings.ReplaceAll(description, "\t", " ")
	if description == "<p></p>" {
		description = ""
	}

	// Example
	example = strings.Join(exampleLines, "\n")
	example = strings.TrimLeft(example, " \t")
	example = strings.TrimLeft(example, "\n")
	example = formatExample(example)

	// Notes
	notes = strings.Join(notesLines, "</p>\n<p>")
	notes = "<p>" + notes + "</p>"
	notes = strings.ReplaceAll(notes, "\t", " ")
	if notes == "<p></p>" {
		notes = ""
	}

	// Deprecation Note
	deprecationNoteRaw := strings.Join(deprecationNoteLines, "\n")
	deprecationNote = strings.Join(deprecationNoteLines, "</p>\n<p>")
	deprecationNote = "<p>" + deprecationNote + "</p>"
	deprecationNote = strings.ReplaceAll(deprecationNote, "\t", " ")
	if deprecationNote == "<p></p>" {
		deprecationNote = ""
	}

	// Returns
	returns = strings.Join(returnsLines, "</p>\n<p>")
	returns = "<p>" + returns + "</p>"
	returns = strings.ReplaceAll(returns, "\t", " ")
	if returns == "<p></p>" {
		returns = ""
	}
	
	return DescriptionData{
		Description:     description,
		Example:         example,
		Notes:           notes,
		DeprecationNote: deprecationNote,
		Returns:         returns,

		// Raw fields
		DescriptionRaw:     descriptionRaw,
		DeprecationNoteRaw: deprecationNoteRaw,
	}
}

// Format the example code snippet to be properly indented and aligned
// using the go/format package.
//
// Returns: Formatted example code as a string; if formatting fails, returns
// the original example string
func formatExample(example string) string {
	src := []byte(example)
	formattedSrc, err := format.Source(src)
	if err != nil {
		return example
	}

	return string(formattedSrc)
}

// Extract the parameters from a function or method declaration
//
// Returns: Strings representing parameter names and their types
func extractParameters(fieldList *ast.FieldList) []string {
	var params []string
	if fieldList != nil {
		for _, param := range fieldList.List {
			typeStr := formatExpr(param.Type)
			for _, name := range param.Names {
				params = append(params, name.Name+" "+typeStr)
			}
			if len(param.Names) == 0 {
				params = append(params, typeStr)
			}
		}
	}
	return params
}

// Extracts the body of a function declaration
//
// Returns: Strings where each string represents a parameter name
// and type, or just the type if no name is provided
func extractBody(fs *token.FileSet, fn *ast.FuncDecl) string {
	if fn.Body == nil {
		return ""
	}

	start := fs.Position(fn.Body.Pos()).Offset
	end := fs.Position(fn.Body.End()).Offset
	fileContent, _ := os.ReadFile(fs.File(fn.Body.Pos()).Name())
	body := string(fileContent[start:end])

	// before returning we have to escape possible html snippets in it since
	// those snippets are rendered by highlighting.js which has an issue with
	// unescaped html snippets (yeah even if inside a Go string, what a pleasure)
	return html.EscapeString(body)
}

// Format an expression into a string representation
// using the go/format package.
//
// Returns: Formatted string of the expression
func formatExpr(expr ast.Expr) string {
	var out strings.Builder
	if err := format.Node(&out, token.NewFileSet(), expr); err != nil {
		return ""
	}
	return out.String()
}

// Identify references to other entities within the given entity.
//
// Returns: ReferenceInfo with details of each referenced entity
func findReferences(entity EntityInfo, entityIndex map[string]EntityInfo) []ReferenceInfo {
	var references []ReferenceInfo

	// Check for parameters
	for _, param := range entity.Parameters {
		paramType := strings.Split(param, " ")[1]
		if refEntity, found := entityIndex[entity.Package+"."+paramType]; found {
			references = append(references, ReferenceInfo{
				Name:        paramType,
				Package:     refEntity.Package,
				PackageURL:  refEntity.PackageURL,
				PackagePath: refEntity.PackagePath,
			})
		}
	}

	// Check for returns
	for _, ret := range entity.Returns {
		if refEntity, found := entityIndex[entity.Package+"."+ret]; found {
			references = append(references, ReferenceInfo{
				Name:        ret,
				Package:     refEntity.Package,
				PackageURL:  refEntity.PackageURL,
				PackagePath: refEntity.PackagePath,
			})
		}
	}

	// Check for fields
	for _, field := range entity.Fields {
		if refEntity, found := entityIndex[entity.Package+"."+field.Type]; found {
			references = append(references, ReferenceInfo{
				Name:        field.Type,
				Package:     refEntity.Package,
				PackageURL:  refEntity.PackageURL,
				PackagePath: refEntity.PackagePath,
			})
		}
	}

	return references
}
