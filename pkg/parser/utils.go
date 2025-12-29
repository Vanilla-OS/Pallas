package parser

import (
	"go/ast"
	"go/format"
	"go/token"
	"strings"

	"github.com/russross/blackfriday/v2"
)

// extractMethods collects method information from an interface.
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

// extractFields collects field information from a struct.
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

// extractTag returns the backtick-wrapped tag string of a struct field.
func extractTag(field *ast.Field) string {
	if field.Tag != nil {
		return strings.Trim(field.Tag.Value, "`")
	}
	return ""
}

// DescriptionData holds both formatted (HTML) and raw documentation components.
type DescriptionData struct {
	Description     string
	Example         string
	Notes           string
	DeprecationNote string
	Returns         string

	DescriptionRaw     string
	DeprecationNoteRaw string
}

// extractDescriptionData parses a Go documentation comment and extracts metadata sections.
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
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, "Example:") {
			isExample = true
			isNotes = false
			isDeprecationNote = false
			isReturns = false
			continue
		}
		if strings.HasPrefix(trimmedLine, "Notes:") {
			isNotes = true
			isExample = false
			isDeprecationNote = false
			isReturns = false
			continue
		}
		if strings.HasPrefix(trimmedLine, "Deprecated:") {
			isDeprecationNote = true
			isExample = false
			isNotes = false
			isReturns = false
			continue
		}
		if strings.HasPrefix(trimmedLine, "Returns:") {
			isReturns = true
			isExample = false
			isNotes = false
			isDeprecationNote = false
			continue
		}

		if isExample {
			exampleLines = append(exampleLines, line)
		} else if isNotes {
			notesLines = append(notesLines, trimmedLine)
		} else if isDeprecationNote {
			deprecationNoteLines = append(deprecationNoteLines, trimmedLine)
		} else if isReturns {
			returnsLines = append(returnsLines, trimmedLine)
		} else {
			descLines = append(descLines, trimmedLine)
		}
	}

	descriptionRaw := strings.Join(descLines, "\n")
	description = markdownToHTML(descriptionRaw)

	example = strings.Join(exampleLines, "\n")
	example = formatExample(example)

	notesRaw := strings.Join(notesLines, "\n")
	notes = markdownToHTML(notesRaw)

	deprecationNoteRaw := strings.Join(deprecationNoteLines, "\n")
	deprecationNote = markdownToHTML(deprecationNoteRaw)

	returnsRaw := strings.Join(returnsLines, "\n")
	returns = markdownToHTML(returnsRaw)

	return DescriptionData{
		Description:     description,
		Example:         example,
		Notes:           notes,
		DeprecationNote: deprecationNote,
		Returns:         returns,

		DescriptionRaw:     descriptionRaw,
		DeprecationNoteRaw: deprecationNoteRaw,
	}
}

// markdownToHTML converts markdown text to HTML using the blackfriday engine.
func markdownToHTML(md string) string {
	renderer := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
		Flags: blackfriday.CommonHTMLFlags,
	})
	extensions := blackfriday.CommonExtensions | blackfriday.AutoHeadingIDs | blackfriday.HardLineBreak | blackfriday.Autolink
	output := blackfriday.Run([]byte(md), blackfriday.WithRenderer(renderer), blackfriday.WithExtensions(extensions))
	return string(output)
}

// formatExample cleans up indentation in code examples while preserving relative structure.
func formatExample(example string) string {
	lines := strings.Split(example, "\n")

	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}

	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}

	if len(lines) == 0 {
		return ""
	}

	minIndent := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " \t"))
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	if minIndent <= 0 {
		return strings.Join(lines, "\n")
	}

	dedentedLines := make([]string, len(lines))
	for i, line := range lines {
		if len(line) >= minIndent {
			dedentedLines[i] = line[minIndent:]
		} else {
			dedentedLines[i] = strings.TrimLeft(line, " \t")
		}
	}

	return strings.Join(dedentedLines, "\n")
}

// extractParameters returns a slice of formatted parameter strings.
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

// formatExpr converts an AST expression إلى its Go source representation.
func formatExpr(expr ast.Expr) string {
	var out strings.Builder
	if err := format.Node(&out, token.NewFileSet(), expr); err != nil {
		return ""
	}
	return out.String()
}

// extractBody retrieves the source code block of a function or method.
func extractBody(node ast.Node, fset *token.FileSet, fileContent []byte) string {
	if node == nil {
		return ""
	}

	// Try to slice directly from the original file content for accuracy
	if len(fileContent) > 0 {
		start := fset.Position(node.Pos()).Offset
		end := fset.Position(node.End()).Offset
		if start >= 0 && end <= len(fileContent) && start < end {
			return string(fileContent[start:end])
		}
	}

	var buf strings.Builder
	if err := format.Node(&buf, fset, node); err != nil {
		return ""
	}
	return buf.String()
}

// findReferences scans an entity for references to other known types within the same package.
func findReferences(entity EntityInfo, entityIndex map[string]EntityInfo) []ReferenceInfo {
	var references []ReferenceInfo

	for _, param := range entity.Parameters {
		parts := strings.Fields(param)
		if len(parts) == 0 {
			continue
		}
		paramType := parts[len(parts)-1]
		if refEntity, found := entityIndex[entity.Package+"."+paramType]; found {
			references = append(references, ReferenceInfo{
				Name:        paramType,
				Package:     refEntity.Package,
				PackageURL:  refEntity.PackageURL,
				PackagePath: refEntity.PackagePath,
			})
		}
	}

	for _, ret := range entity.Returns {
		parts := strings.Fields(ret)
		if len(parts) == 0 {
			continue
		}
		retType := parts[len(parts)-1]
		if refEntity, found := entityIndex[entity.Package+"."+retType]; found {
			references = append(references, ReferenceInfo{
				Name:        retType,
				Package:     refEntity.Package,
				PackageURL:  refEntity.PackageURL,
				PackagePath: refEntity.PackagePath,
			})
		}
	}

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
