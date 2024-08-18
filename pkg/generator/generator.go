package generator

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/vanilla-os/pallas/pkg/parser"
)

//go:embed templates/entities.html
var htmlTemplate string

// GenerateHTML generates an HTML file for the given package and entities
func GenerateHTML(projectPath string, packagePath string, entities []parser.EntityInfo, outputDir string, docTitle string) error {
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

	data := struct {
		PackageName string
		Entities    []parser.EntityInfo
		Title       string
	}{
		PackageName: relativePackagePath,
		Entities:    entities,
		Title:       docTitle,
	}

	return tmpl.Execute(file, data)
}
