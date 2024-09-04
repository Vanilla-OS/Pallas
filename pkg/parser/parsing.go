package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

// ParseEntitiesInPackage parses the entities in a given package and returns
// a slice of EntityInfo
//
// Example:
//
//	entities, err := parser.ParseEntitiesInPackage("/home/me/myproject/pkg/mypackage")
//	if err != nil {
//		log.Fatalf("Error parsing entities: %v", err)
//	}
//	for _, entity := range entities {
//		fmt.Printf("Name: %s\n", entity.Name)
//		fmt.Printf("Type: %s\n", entity.Type)
//		fmt.Printf("Description: %s\n", entity.Description)
//		fmt.Printf("Package: %s\n", entity.Package)
//	}
//
// Notes:
// The package must be a full path to the package directory
func ParseEntitiesInPackage(projectPath string, pkgPath string, relativePath string) ([]EntityInfo, []ImportInfo, error) {
	var entities []EntityInfo
	var imports []ImportInfo
	var interfaces = make(map[string]EntityInfo)
	var methodsByType = make(map[string][]EntityInfo)
	var entityIndex = make(map[string]EntityInfo)

	fs := token.NewFileSet()
	pkgs, err := parser.ParseDir(fs, pkgPath, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}

	var pkgName string
	var pkg *ast.Package
	for k, v := range pkgs {
		pkgName = k
		pkg = v
		break
	}

	extractors := map[string]EntityExtractor{
		"function":  FunctionExtractor{},
		"method":    MethodExtractor{},
		"struct":    StructExtractor{},
		"interface": InterfaceExtractor{},
		"type":      TypeExtractor{},
	}

	// Replace slashes with hyphens to ensure unique filenames
	url := strings.ReplaceAll(relativePath, string(os.PathSeparator), "-")

	for _, file := range pkg.Files {

		// here we parse all imports
		for _, imp := range file.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			var importName string
			if imp.Name != nil {
				importName = imp.Name.Name
			} else {
				importName = ""
			}

			importURL := strings.ReplaceAll(importPath, "/", "-")

			imports = append(imports, ImportInfo{
				Path:  importPath,
				URL:   importURL,
				Alias: importName,
			})
		}

		// here we parse all entities types
		for _, decl := range file.Decls {
			switch decl := decl.(type) {
			case *ast.FuncDecl:
				if decl.Recv != nil {
					receiverType := formatExpr(decl.Recv.List[0].Type)
					method := extractors["method"].Extract(decl, fs, interfaces, pkgName, relativePath, url)
					methodsByType[receiverType] = append(methodsByType[receiverType], method)
				} else {
					entity := extractors["function"].Extract(decl, fs, interfaces, pkgName, relativePath, url)
					entities = append(entities, entity)
					entityIndex[pkgName+"."+entity.Name] = entity
				}
			case *ast.GenDecl:
				for _, spec := range decl.Specs {
					switch spec := spec.(type) {
					case *ast.TypeSpec:
						var entityType string
						switch spec.Type.(type) {
						case *ast.StructType:
							entityType = "struct"
						case *ast.InterfaceType:
							entityType = "interface"
							if _, exists := interfaces[spec.Name.Name]; !exists {
								ifaceInfo := extractors[entityType].Extract(decl, fs, interfaces, pkgName, relativePath, url)
								ifaceInfo.Package = pkgName
								interfaces[spec.Name.Name] = ifaceInfo
								entities = append(entities, ifaceInfo)
								entityIndex[pkgName+"."+ifaceInfo.Name] = ifaceInfo
							}
						default:
							entityType = "type"
						}

						if entityType != "interface" {
							entity := extractors[entityType].Extract(decl, fs, interfaces, pkgName, relativePath, url)
							entities = append(entities, entity)
							entityIndex[pkgName+"."+entity.Name] = entity
						}
					}
				}
			}
		}
	}

	// Here we associate methods with structs, resolve interfaces
	// implementations and find references for each entity
	for i, entity := range entities {
		references := findReferences(entity, entityIndex)
		entity.References = references

		// if the entity is a struct, we associate methods with it
		if entity.Type == "struct" {
			receiverName := entity.Name
			if methods, ok := methodsByType[receiverName]; ok {
				entity.Methods = append(entity.Methods, methods...)
			} else if methods, ok := methodsByType["*"+receiverName]; ok {
				entity.Methods = append(entity.Methods, methods...)
			}

			entity.Implements = findImplementedInterfaces(entity, interfaces)

			// and here we find references for each method if any
			for j, method := range entity.Methods {
				methodReferences := findReferences(method, entityIndex)
				entity.Methods[j].References = methodReferences
			}
		}

		entities[i] = entity
	}

	return entities, imports, nil
}

// findImplementedInterfaces checks which interfaces are implemented by a struct
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

// implementsInterface checks if a struct implements a given interface
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

// methodsMatch checks if the parameters and return types of two methods match
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
