package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

// ParseEntitiesInPackage parses all Go entities (functions, structs, etc.) within a specified package directory.
// It returns a list of entities, associated imports, and any encountered error.
func ParseEntitiesInPackage(projectRoot string, pkgPath string, relPath string) ([]EntityInfo, []ImportInfo, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, pkgPath, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}

	var entities []EntityInfo
	var imports []ImportInfo
	var methodsByType = make(map[string][]EntityInfo)
	var interfaces = make(map[string]EntityInfo)
	var entityIndex = make(map[string]EntityInfo)

	extractors := map[string]EntityExtractor{
		"function":  FunctionExtractor{},
		"method":    MethodExtractor{},
		"struct":    StructExtractor{},
		"interface": InterfaceExtractor{},
		"type":      TypeExtractor{},
	}

	for pkgName, pkg := range pkgs {
		for filename, file := range pkg.Files {
			content, err := os.ReadFile(filename)
			if err != nil {
				return nil, nil, err
			}

			// Process file imports
			for _, imp := range file.Imports {
				path := strings.Trim(imp.Path.Value, "\"")
				name := ""
				if imp.Name != nil {
					name = imp.Name.Name
				}

				importURL := strings.ReplaceAll(path, "/", "-")
				doc := ""
				comment := ""

				if imp.Doc != nil {
					doc = imp.Doc.Text()
				}
				if imp.Comment != nil {
					comment = imp.Comment.Text()
				}

				imports = append(imports, ImportInfo{
					Path:    path,
					Alias:   name,
					URL:     importURL,
					Doc:     doc,
					Comment: comment,
					Package: pkgName,
					File:    filename,
				})
			}

			// Process declarations
			for _, decl := range file.Decls {
				switch d := decl.(type) {
				case *ast.FuncDecl:
					if d.Recv != nil {
						exs, err := extractors["method"].Extract(d, fset, pkgName, content)
						if err == nil && len(exs) > 0 {
							receiverType := formatExpr(d.Recv.List[0].Type)
							methodsByType[receiverType] = append(methodsByType[receiverType], exs[0])
						}
					} else {
						exs, err := extractors["function"].Extract(d, fset, pkgName, content)
						if err == nil && len(exs) > 0 {
							entities = append(entities, exs[0])
							entityIndex[pkgName+"."+exs[0].Name] = exs[0]
						}
					}
				case *ast.GenDecl:
					if d.Tok == token.TYPE {
						structs, _ := extractors["struct"].Extract(d, fset, pkgName, content)
						for _, s := range structs {
							entities = append(entities, s)
							entityIndex[pkgName+"."+s.Name] = s
						}

						ifaces, _ := extractors["interface"].Extract(d, fset, pkgName, content)
						for _, i := range ifaces {
							entities = append(entities, i)
							interfaces[i.Name] = i
							entityIndex[pkgName+"."+i.Name] = i
						}

						types, _ := extractors["type"].Extract(d, fset, pkgName, content)
						for _, t := range types {
							entities = append(entities, t)
							entityIndex[pkgName+"."+t.Name] = t
						}
					}
				}
			}
		}
	}

	// Link methods and calculate relationships between entities
	for i, entity := range entities {
		entity.References = findReferences(entity, entityIndex)

		if entity.Type == "struct" {
			receiverName := entity.Name

			if methods, ok := methodsByType[receiverName]; ok {
				entity.Methods = append(entity.Methods, methods...)
			}
			if methods, ok := methodsByType["*"+receiverName]; ok {
				entity.Methods = append(entity.Methods, methods...)
			}

			for j, method := range entity.Methods {
				entity.Methods[j].References = findReferences(method, entityIndex)
			}

			entity.Implements = findImplementedInterfaces(entity, interfaces)
		}

		entities[i] = entity
	}

	return entities, imports, nil
}

// findImplementedInterfaces identifies which interfaces are implemented by a given struct.
func findImplementedInterfaces(entity EntityInfo, interfaces map[string]EntityInfo) []ImplementationInfo {
	var implemented []ImplementationInfo

	for ifaceName, ifaceInfo := range interfaces {
		if implementsInterface(entity, ifaceInfo) {
			implemented = append(implemented, ImplementationInfo{
				InterfaceName: ifaceName,
				Package:       ifaceInfo.Package,
			})
		}
	}

	return implemented
}

// implementsInterface checks if a struct implements all methods of a given interface.
func implementsInterface(entity EntityInfo, iface EntityInfo) bool {
	methodSet := make(map[string]EntityInfo)
	for _, method := range entity.Methods {
		methodSet[method.Name] = method
	}

	for _, ifaceMethod := range iface.Methods {
		if method, ok := methodSet[ifaceMethod.Name]; !ok {
			return false
		} else {
			if !methodsMatch(ifaceMethod, method) {
				return false
			}
		}
	}

	return true
}

// methodsMatch verifies if the signatures of two methods match.
func methodsMatch(ifaceMethod, structMethod EntityInfo) bool {
	if len(ifaceMethod.Parameters) != len(structMethod.Parameters) ||
		len(ifaceMethod.Returns) != len(structMethod.Returns) {
		return false
	}

	for i, param := range ifaceMethod.Parameters {
		if param != structMethod.Parameters[i] {
			return false
		}
	}

	for i, ret := range ifaceMethod.Returns {
		if ret != structMethod.Returns[i] {
			return false
		}
	}

	return true
}
