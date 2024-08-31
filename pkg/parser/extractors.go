package parser

import (
	"go/ast"
	"go/token"
)

// EntityExtractor defines an interface for extracting information from AST declarations
type EntityExtractor interface {
	Extract(decl ast.Decl, fs *token.FileSet, interfaces map[string]EntityInfo, pkgName string) EntityInfo
}

// FunctionExtractor extracts information from function declarations
type FunctionExtractor struct{}

func (f FunctionExtractor) Extract(decl ast.Decl, fs *token.FileSet, interfaces map[string]EntityInfo, pkgName string) EntityInfo {
	funcDecl := decl.(*ast.FuncDecl)
	descriptionData := extractDescriptionData(funcDecl.Doc.Text())
	return EntityInfo{
		Name:            funcDecl.Name.Name,
		Type:            "function",
		Body:            extractBody(fs, funcDecl),
		Description:     descriptionData.Description,
		Example:         descriptionData.Example,
		Notes:           descriptionData.Notes,
		DeprecationNote: descriptionData.DeprecationNote,
		Parameters:      extractParameters(funcDecl.Type.Params),
		Returns:         extractParameters(funcDecl.Type.Results),
		Package:         pkgName,

		// Raw fields
		DescriptionRaw:     descriptionData.DescriptionRaw,
		DeprecationNoteRaw: descriptionData.DeprecationNoteRaw,
	}
}

// StructExtractor extracts information from struct declarations
type StructExtractor struct{}

func (s StructExtractor) Extract(decl ast.Decl, fs *token.FileSet, interfaces map[string]EntityInfo, pkgName string) EntityInfo {
	spec := decl.(*ast.GenDecl).Specs[0].(*ast.TypeSpec)
	structType := spec.Type.(*ast.StructType)
	return EntityInfo{
		Name:        spec.Name.Name,
		Type:        "struct",
		Description: extractComment(decl.(*ast.GenDecl).Doc),
		Fields:      extractFields(structType),
		Package:     pkgName,
	}
}

// InterfaceExtractor extracts information from interface declarations
type InterfaceExtractor struct{}

func (i InterfaceExtractor) Extract(decl ast.Decl, fs *token.FileSet, interfaces map[string]EntityInfo, pkgName string) EntityInfo {
	spec := decl.(*ast.GenDecl).Specs[0].(*ast.TypeSpec)
	interfaceType := spec.Type.(*ast.InterfaceType)
	return EntityInfo{
		Name:        spec.Name.Name,
		Type:        "interface",
		Description: extractComment(decl.(*ast.GenDecl).Doc),
		Methods:     extractMethods(interfaceType),
		Package:     pkgName,
	}
}

// TypeExtractor extracts information from type declarations
type TypeExtractor struct{}

func (t TypeExtractor) Extract(decl ast.Decl, fs *token.FileSet, interfaces map[string]EntityInfo, pkgName string) EntityInfo {
	spec := decl.(*ast.GenDecl).Specs[0].(*ast.TypeSpec)
	typeExpr := formatExpr(spec.Type)

	return EntityInfo{
		Name:        spec.Name.Name,
		Type:        "type",
		Description: extractComment(decl.(*ast.GenDecl).Doc),
		Body:        typeExpr,
		Package:     pkgName,
	}
}
