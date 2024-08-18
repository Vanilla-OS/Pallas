package parser

import (
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

// EntityInfo contains relevant information about each entity in the package
// (functions, types, interfaces)
type EntityInfo struct {
	Name        string
	Description string
	Example     string
	Parameters  []string
	Returns     []string
	Body        string
	Type        string
}

// ParseEntitiesInPackage parses the entities in a given package and returns
// a slice of EntityInfo
//
// Example:
//
//	entities, err := parser.ParseEntitiesInPackage(pkgPath)
//	if err != nil {
//		log.Fatalf("Error parsing package %s: %v", err)
//	}
//	for _, entity := range entities {
//		fmt.Printf("Name: %s\n", entity.Name)
//		fmt.Printf("Type: %s\n", entity.Type)
//		fmt.Printf("Description: %s\n", entity.Description)
//		fmt.Printf("Example: %s\n", entity.Example)
//		fmt.Printf("Parameters: %v\n", entity.Parameters)
//		fmt.Printf("Returns: %v\n", entity.Returns)
//	}
func ParseEntitiesInPackage(pkgPath string) ([]EntityInfo, error) {
	var entities []EntityInfo

	// Create a new file set
	fs := token.NewFileSet()

	// Parse the package directory
	pkgs, err := parser.ParseDir(fs, pkgPath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				switch decl := decl.(type) {
				case *ast.FuncDecl:
					description, example := extractDescriptionAndExample(decl.Doc.Text())
					entity := EntityInfo{
						Name:        decl.Name.Name,
						Type:        "function",
						Body:        extractBody(fs, decl),
						Description: description,
						Example:     example,
					}
					entity.Parameters = extractParameters(decl)
					entity.Returns = extractReturns(decl)
					entities = append(entities, entity)

				case *ast.GenDecl:
					for _, spec := range decl.Specs {
						switch spec := spec.(type) {
						case *ast.TypeSpec:
							entity := EntityInfo{
								Name: spec.Name.Name,
							}
							switch spec.Type.(type) {
							case *ast.StructType:
								entity.Type = "type"
							case *ast.InterfaceType:
								entity.Type = "interface"
							}
							entity.Description = extractComment(decl.Doc)
							entities = append(entities, entity)
						}
					}
				}
			}
		}
	}

	return entities, nil
}

// extractDescriptionAndExample extracts the description and example code from a
// function's documentation comment
func extractDescriptionAndExample(doc string) (description string, example string) {
	lines := strings.Split(doc, "\n")
	var descLines []string
	var exampleLines []string
	isExample := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Example:") {
			isExample = true
			continue
		}

		if isExample {
			exampleLines = append(exampleLines, line)
		} else {
			if line != "" {
				descLines = append(descLines, line)
			}
		}
	}

	// Join the description lines with a <p> tag
	description = strings.Join(descLines, "</p>\n<p>")
	description = "<p>" + description + "</p>"
	description = strings.ReplaceAll(description, "\t", " ")

	// Clean up the example code
	example = strings.Join(exampleLines, "\n")
	example = strings.TrimLeft(example, " \t")
	example = strings.TrimLeft(example, "\n")

	// Format the example code
	example = formatExample(example)

	return description, example
}

// formatExample formats the example code using the go/format package
func formatExample(example string) string {
	// Convert the example string to a byte slice
	src := []byte(example)

	// Use format.Source to format the code
	formattedSrc, err := format.Source(src)
	if err != nil {
		return example
	}

	// Convert the formatted byte slice back to a string and return it
	return string(formattedSrc)
}

// extractParameters extracts the parameters from a function declaration
func extractParameters(fn *ast.FuncDecl) []string {
	var params []string
	for _, param := range fn.Type.Params.List {
		typeStr := formatExpr(param.Type)
		for _, name := range param.Names {
			params = append(params, name.Name+" "+typeStr)
		}
	}
	return params
}

// extractReturns extracts the return types from a function declaration
func extractReturns(fn *ast.FuncDecl) []string {
	var returns []string
	if fn.Type.Results != nil {
		for _, result := range fn.Type.Results.List {
			returns = append(returns, formatExpr(result.Type))
		}
	}
	return returns
}

// extractBody extracts the body of a function declaration
func extractBody(fs *token.FileSet, fn *ast.FuncDecl) string {
	if fn.Body == nil {
		return ""
	}
	start := fs.Position(fn.Body.Pos()).Offset
	end := fs.Position(fn.Body.End()).Offset
	fileContent, _ := os.ReadFile(fs.File(fn.Body.Pos()).Name())
	return string(fileContent[start:end])
}

// formatExpr formats an expression using the go/format package
func formatExpr(expr ast.Expr) string {
	var out strings.Builder
	if err := format.Node(&out, token.NewFileSet(), expr); err != nil {
		return ""
	}
	return out.String()
}

// extractComment extracts the comment text from a comment group
func extractComment(doc *ast.CommentGroup) string {
	if doc == nil {
		return ""
	}
	return strings.TrimSpace(doc.Text())
}
