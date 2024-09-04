package generator

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/vanilla-os/pallas/pkg/parser"
)

//go:embed templates/entities.html
var htmlTemplate string

//go:embed templates/static/*
var staticAssets embed.FS

// GenerateHTML generates an HTML file for the given package and entities
func GenerateHTML(projectPath string, packagePath string, entities []parser.EntityInfo, imports []parser.ImportInfo, outputDir string, docTitle string) error {
	// Create the output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return fmt.Errorf("error creating output directory: %v", err)
	}

	tmpl, err := template.New("package").Parse(htmlTemplate)
	if err != nil {
		return err
	}

	// Extract the relative path of the package based on the project path
	relativePackagePath, err := filepath.Rel(projectPath, packagePath)
	if err != nil {
		return err
	}

	// Replace slashes with hyphens to ensure unique filenames
	safeFileName := strings.ReplaceAll(relativePackagePath, string(os.PathSeparator), "-")

	// Generate the HTML file
	filePath := filepath.Join(outputDir, fmt.Sprintf("%s.html", safeFileName))
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Determine if the package has functions, types, and structs
	hasFunctions := false
	hasTypes := false
	hasStructs := false
	hasInterfaces := false
	hasImports := len(imports) > 0
	for _, entity := range entities {
		if entity.Type == "function" {
			hasFunctions = true
		} else if entity.Type == "type" {
			hasTypes = true
		} else if entity.Type == "struct" {
			hasStructs = true
		} else if entity.Type == "interface" {
			hasInterfaces = true
		}
	}

	data := struct {
		PackageName   string
		Entities      []parser.EntityInfo
		Imports       []parser.ImportInfo
		Title         string
		HasFunctions  bool
		HasTypes      bool
		HasStructs    bool
		HasInterfaces bool
		HasImports    bool
	}{
		PackageName:   relativePackagePath,
		Entities:      entities,
		Imports:       imports,
		Title:         docTitle,
		HasFunctions:  hasFunctions,
		HasTypes:      hasTypes,
		HasStructs:    hasStructs,
		HasInterfaces: hasInterfaces,
		HasImports:    hasImports,
	}

	return tmpl.Execute(file, data)
}
