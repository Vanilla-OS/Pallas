package parser

import (
	"go/ast"
	"go/token"
)

// EntityExtractor defines an interface for extracting information from AST declarations
type EntityExtractor interface {
	Extract(decl ast.Decl, fs *token.FileSet, interfaces map[string]EntityInfo, pkgName string, packagePath string, url string) EntityInfo
}

// FunctionExtractor extracts information from function declarations
type FunctionExtractor struct{}

func (f FunctionExtractor) Extract(decl ast.Decl, fs *token.FileSet, interfaces map[string]EntityInfo, pkgName string, packagePath string, url string) EntityInfo {
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
		PackageURL:      url,
		PackagePath:     packagePath,

		// Raw fields
		DescriptionRaw:     descriptionData.DescriptionRaw,
		DeprecationNoteRaw: descriptionData.DeprecationNoteRaw,
	}
}

// MethodExtractor extracts information from method declarations
type MethodExtractor struct{}

func (m MethodExtractor) Extract(decl ast.Decl, fs *token.FileSet, interfaces map[string]EntityInfo, pkgName string, packagePath string, url string) EntityInfo {
	funcDecl := decl.(*ast.FuncDecl)
	descriptionData := extractDescriptionData(funcDecl.Doc.Text())

	return EntityInfo{
		Name:            funcDecl.Name.Name,
		Type:            "method",
		Body:            extractBody(fs, funcDecl),
		Description:     descriptionData.Description,
		Example:         descriptionData.Example,
		Notes:           descriptionData.Notes,
		DeprecationNote: descriptionData.DeprecationNote,
		Parameters:      extractParameters(funcDecl.Type.Params),
		Returns:         extractParameters(funcDecl.Type.Results),
		Package:         pkgName,
		PackageURL:      url,
		PackagePath:     packagePath,

		// Raw fields
		DescriptionRaw:     descriptionData.DescriptionRaw,
		DeprecationNoteRaw: descriptionData.DeprecationNoteRaw,
	}
}

// StructExtractor extracts information from struct declarations
type StructExtractor struct{}

func (s StructExtractor) Extract(decl ast.Decl, fs *token.FileSet, interfaces map[string]EntityInfo, pkgName string, packagePath string, url string) EntityInfo {
	spec := decl.(*ast.GenDecl).Specs[0].(*ast.TypeSpec)
	structType := spec.Type.(*ast.StructType)

	descriptionData := extractDescriptionData(decl.(*ast.GenDecl).Doc.Text())

	return EntityInfo{
		Name:            spec.Name.Name,
		Type:            "struct",
		Description:     descriptionData.Description,
		Notes:           descriptionData.Notes,
		DeprecationNote: descriptionData.DeprecationNote,
		Fields:          extractFields(structType),
		Package:         pkgName,
		PackageURL:      url,
		PackagePath:     packagePath,

		// Raw fields
		DescriptionRaw:     descriptionData.DescriptionRaw,
		DeprecationNoteRaw: descriptionData.DeprecationNoteRaw,
	}
}

// InterfaceExtractor extracts information from interface declarations
type InterfaceExtractor struct{}

func (i InterfaceExtractor) Extract(decl ast.Decl, fs *token.FileSet, interfaces map[string]EntityInfo, pkgName string, packagePath string, url string) EntityInfo {
	spec := decl.(*ast.GenDecl).Specs[0].(*ast.TypeSpec)
	interfaceType := spec.Type.(*ast.InterfaceType)

	descriptionData := extractDescriptionData(decl.(*ast.GenDecl).Doc.Text())

	return EntityInfo{
		Name:            spec.Name.Name,
		Description:     descriptionData.Description,
		Notes:           descriptionData.Notes,
		DeprecationNote: descriptionData.DeprecationNote,
		Type:            "interface",
		Methods:         extractMethods(interfaceType),
		Package:         pkgName,
		PackageURL:      url,
		PackagePath:     packagePath,

		// Raw fields
		DescriptionRaw:     descriptionData.DescriptionRaw,
		DeprecationNoteRaw: descriptionData.DeprecationNoteRaw,
	}
}

// TypeExtractor extracts information from type declarations
type TypeExtractor struct{}

func (t TypeExtractor) Extract(decl ast.Decl, fs *token.FileSet, interfaces map[string]EntityInfo, pkgName string, packagePath string, url string) EntityInfo {
	spec := decl.(*ast.GenDecl).Specs[0].(*ast.TypeSpec)
	typeExpr := formatExpr(spec.Type)

	descriptionData := extractDescriptionData(decl.(*ast.GenDecl).Doc.Text())

	return EntityInfo{
		Name:            spec.Name.Name,
		Description:     descriptionData.Description,
		Notes:           descriptionData.Notes,
		DeprecationNote: descriptionData.DeprecationNote,
		Type:            "type",
		Body:            typeExpr,
		Package:         pkgName,
		PackageURL:      url,
		PackagePath:     packagePath,

		// Raw fields
		DescriptionRaw:     descriptionData.DescriptionRaw,
		DeprecationNoteRaw: descriptionData.DeprecationNoteRaw,
	}
}
