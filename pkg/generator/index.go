package generator

import (
	_ "embed"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

//go:embed templates/index.html
var indexTemplate string

type PackageLink struct {
	Name string
	Link string
}

// GenerateIndex generates the index.html file listing all the documented packages
func GenerateIndex(projectPath string, packages []string, outputDir string, docTitle string) error {
	// Create the output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return err
	}

	tmpl, err := template.New("index").Parse(indexTemplate)
	if err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(outputDir, "index.html"))
	if err != nil {
		return err
	}
	defer file.Close()

	// Group packages by their prefix
	groupedPackages := make(map[string][]PackageLink)
	var totalPackages int

	for _, pkg := range packages {
		parts := strings.Split(pkg, string(os.PathSeparator))
		prefix := parts[0]

		safeFileName := strings.ReplaceAll(pkg, string(os.PathSeparator), "-")
		groupedPackages[prefix] = append(groupedPackages[prefix], PackageLink{
			Name: pkg,
			Link: safeFileName + ".html",
		})

		totalPackages++
	}

	// Sort packages within each group
	for _, packages := range groupedPackages {
		sort.Slice(packages, func(i, j int) bool {
			return packages[i].Name < packages[j].Name
		})
	}

	// Execute template with data
	return tmpl.Execute(file, struct {
		Title           string
		GroupedPackages map[string][]PackageLink
		TotalPackages   int
	}{
		Title:           docTitle,
		GroupedPackages: groupedPackages,
		TotalPackages:   totalPackages,
	})
}
