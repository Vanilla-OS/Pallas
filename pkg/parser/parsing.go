package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
)

// ParseEntitiesInPackage parses the entities in a given package and returns
// a slice of EntityInfo
//
// Example:
//
//	entities, err := parser.ParseEntitiesInPackage("github.com/gophercises/quiz")
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
func ParseEntitiesInPackage(pkgPath string) ([]EntityInfo, error) {
	var entities []EntityInfo
	var interfaces = make(map[string]EntityInfo)
	var methodsByType = make(map[string][]MethodInfo)

	fs := token.NewFileSet()
	pkgs, err := parser.ParseDir(fs, pkgPath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	extractors := map[string]EntityExtractor{
		"function":  FunctionExtractor{},
		"struct":    StructExtractor{},
		"interface": InterfaceExtractor{},
	}

	// here we parse all entities, store interfaces, and collect methods
	for pkgName, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				switch decl := decl.(type) {
				case *ast.FuncDecl:
					// Check if the function is a method
					if decl.Recv != nil {
						receiverType := formatExpr(decl.Recv.List[0].Type)
						method := MethodInfo{
							Name:       decl.Name.Name,
							Parameters: extractParameters(decl.Type.Params),
							Returns:    extractParameters(decl.Type.Results),
						}
						methodsByType[receiverType] = append(methodsByType[receiverType], method)
					} else {
						entities = append(entities, extractors["function"].Extract(decl, fs, interfaces, pkgName))
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
								ifaceInfo := extractors[entityType].Extract(decl, fs, interfaces, pkgName)
								ifaceInfo.Package = pkgName
								interfaces[spec.Name.Name] = ifaceInfo
								entities = append(entities, ifaceInfo)
							default:
								entityType = "type"
							}
							if entityType != "interface" {
								entity := extractors[entityType].Extract(decl, fs, interfaces, pkgName)
								entities = append(entities, entity)
							}
						}
					}
				}
			}
		}
	}

	// Here we associate methods with structs and resolve interfaces implementations
	for i, entity := range entities {
		if entity.Type == "struct" {
			if methods, ok := methodsByType[entity.Name]; ok {
				entity.Methods = append(entity.Methods, methods...)
			}
			entity.Implements = findImplementedInterfaces(entity, interfaces)
			entities[i] = entity
		}
	}

	return entities, nil
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
	methodSet := make(map[string]MethodInfo)
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
func methodsMatch(ifaceMethod, structMethod MethodInfo) bool {
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
