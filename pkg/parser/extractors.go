package parser

import (
	"go/ast"
	"go/token"
	"strings"
)

// EntityExtractor defines an interface for extracting information from AST declarations.
type EntityExtractor interface {
	// Extract extracts entity information from an AST node.
	Extract(node ast.Node, fset *token.FileSet, pkgPath string, fileContent []byte) ([]EntityInfo, error)
}

// FunctionExtractor extracts details from function declarations.
type FunctionExtractor struct{}

// Extract implements the EntityExtractor interface for functions.
func (e FunctionExtractor) Extract(node ast.Node, fset *token.FileSet, pkgPath string, fileContent []byte) ([]EntityInfo, error) {
	decl, ok := node.(*ast.FuncDecl)
	if !ok {
		return nil, nil
	}

	doc := decl.Doc.Text()
	descData := extractDescriptionData(doc)

	pos := fset.Position(decl.Pos())
	end := fset.Position(decl.End())

	signature := "func " + decl.Name.Name + strings.TrimPrefix(formatExpr(decl.Type), "func")

	entity := EntityInfo{
		Name:               decl.Name.Name,
		Description:        descData.Description,
		DescriptionRaw:     descData.DescriptionRaw,
		DeprecationNote:    descData.DeprecationNote,
		DeprecationNoteRaw: descData.DeprecationNoteRaw,
		Parameters:         extractParameters(decl.Type.Params),
		Returns:            extractParameters(decl.Type.Results),
		Body:               extractBody(decl.Body, fset, fileContent),
		Example:            descData.Example,
		Notes:              descData.Notes,
		Type:               "function",
		Signature:          signature,
		File:               pos.Filename,
		LineStart:          pos.Line,
		LineEnd:            end.Line,
		Package:            pkgPath,
		PackagePath:        pkgPath,
	}

	return []EntityInfo{entity}, nil
}

// MethodExtractor extracts details from method declarations.
type MethodExtractor struct{}

// Extract implements the EntityExtractor interface for methods.
func (e MethodExtractor) Extract(node ast.Node, fset *token.FileSet, pkgPath string, fileContent []byte) ([]EntityInfo, error) {
	decl, ok := node.(*ast.FuncDecl)
	if !ok {
		return nil, nil
	}

	doc := decl.Doc.Text()
	descData := extractDescriptionData(doc)

	pos := fset.Position(decl.Pos())
	end := fset.Position(decl.End())

	recv := ""
	if decl.Recv != nil && len(decl.Recv.List) > 0 {
		recv = "(" + formatExpr(decl.Recv.List[0].Type) + ") "
	}
	signature := "func " + recv + decl.Name.Name + strings.TrimPrefix(formatExpr(decl.Type), "func")

	entity := EntityInfo{
		Name:               decl.Name.Name,
		Description:        descData.Description,
		DescriptionRaw:     descData.DescriptionRaw,
		DeprecationNote:    descData.DeprecationNote,
		DeprecationNoteRaw: descData.DeprecationNoteRaw,
		Parameters:         extractParameters(decl.Type.Params),
		Returns:            extractParameters(decl.Type.Results),
		Body:               extractBody(decl.Body, fset, fileContent),
		Example:            descData.Example,
		Notes:              descData.Notes,
		Type:               "method",
		Signature:          signature,
		File:               pos.Filename,
		LineStart:          pos.Line,
		LineEnd:            end.Line,
		Package:            pkgPath,
		PackagePath:        pkgPath,
	}

	return []EntityInfo{entity}, nil
}

// StructExtractor extracts details from struct declarations.
type StructExtractor struct{}

// Extract implements the EntityExtractor interface for structs.
func (e StructExtractor) Extract(node ast.Node, fset *token.FileSet, pkgPath string, fileContent []byte) ([]EntityInfo, error) {
	decl, ok := node.(*ast.GenDecl)
	if !ok {
		return nil, nil
	}

	var entities []EntityInfo

	for _, spec := range decl.Specs {
		typeSpec, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			continue
		}

		doc := decl.Doc.Text()
		if typeSpec.Doc != nil {
			doc = typeSpec.Doc.Text()
		}
		descData := extractDescriptionData(doc)

		pos := fset.Position(typeSpec.Pos())
		end := fset.Position(typeSpec.End())

		signature := "type " + typeSpec.Name.Name + " struct"

		entity := EntityInfo{
			Name:               typeSpec.Name.Name,
			Description:        descData.Description,
			DescriptionRaw:     descData.DescriptionRaw,
			DeprecationNote:    descData.DeprecationNote,
			DeprecationNoteRaw: descData.DeprecationNoteRaw,
			Fields:             extractFields(structType),
			Example:            descData.Example,
			Notes:              descData.Notes,
			Type:               "struct",
			Signature:          signature,
			File:               pos.Filename,
			LineStart:          pos.Line,
			LineEnd:            end.Line,
			Package:            pkgPath,
			PackagePath:        pkgPath,
		}
		entities = append(entities, entity)
	}

	return entities, nil
}

// InterfaceExtractor extracts details from interface declarations.
type InterfaceExtractor struct{}

// Extract implements the EntityExtractor interface for interfaces.
func (e InterfaceExtractor) Extract(node ast.Node, fset *token.FileSet, pkgPath string, fileContent []byte) ([]EntityInfo, error) {
	decl, ok := node.(*ast.GenDecl)
	if !ok {
		return nil, nil
	}

	var entities []EntityInfo

	for _, spec := range decl.Specs {
		typeSpec, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
		if !ok {
			continue
		}

		doc := decl.Doc.Text()
		if typeSpec.Doc != nil {
			doc = typeSpec.Doc.Text()
		}
		descData := extractDescriptionData(doc)

		pos := fset.Position(typeSpec.Pos())
		end := fset.Position(typeSpec.End())

		signature := "type " + typeSpec.Name.Name + " interface"

		entity := EntityInfo{
			Name:               typeSpec.Name.Name,
			Description:        descData.Description,
			DescriptionRaw:     descData.DescriptionRaw,
			DeprecationNote:    descData.DeprecationNote,
			DeprecationNoteRaw: descData.DeprecationNoteRaw,
			Methods:            extractMethods(interfaceType),
			Example:            descData.Example,
			Notes:              descData.Notes,
			Type:               "interface",
			Signature:          signature,
			File:               pos.Filename,
			LineStart:          pos.Line,
			LineEnd:            end.Line,
			Package:            pkgPath,
			PackagePath:        pkgPath,
		}
		entities = append(entities, entity)
	}

	return entities, nil
}

// TypeExtractor extracts details from general type declarations.
type TypeExtractor struct{}

// Extract implements the EntityExtractor interface for types.
func (e TypeExtractor) Extract(node ast.Node, fset *token.FileSet, pkgPath string, fileContent []byte) ([]EntityInfo, error) {
	decl, ok := node.(*ast.GenDecl)
	if !ok {
		return nil, nil
	}

	var entities []EntityInfo

	for _, spec := range decl.Specs {
		typeSpec, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		if _, ok := typeSpec.Type.(*ast.StructType); ok {
			continue
		}
		if _, ok := typeSpec.Type.(*ast.InterfaceType); ok {
			continue
		}

		doc := decl.Doc.Text()
		if typeSpec.Doc != nil {
			doc = typeSpec.Doc.Text()
		}
		descData := extractDescriptionData(doc)

		pos := fset.Position(typeSpec.Pos())
		end := fset.Position(typeSpec.End())

		signature := "type " + typeSpec.Name.Name + " " + formatExpr(typeSpec.Type)

		entity := EntityInfo{
			Name:               typeSpec.Name.Name,
			Description:        descData.Description,
			DescriptionRaw:     descData.DescriptionRaw,
			DeprecationNote:    descData.DeprecationNote,
			DeprecationNoteRaw: descData.DeprecationNoteRaw,
			Example:            descData.Example,
			Notes:              descData.Notes,
			Type:               "type",
			Signature:          signature,
			File:               pos.Filename,
			LineStart:          pos.Line,
			LineEnd:            end.Line,
			Package:            pkgPath,
			PackagePath:        pkgPath,
		}
		entities = append(entities, entity)
	}
	return entities, nil
}
