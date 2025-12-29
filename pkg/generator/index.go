package generator

import (
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/vanilla-os/pallas/pkg/parser"
)

// TOCEntry represents an entry in the documentation's Table of Contents.
type TOCEntry struct {
	ID    string
	Title string
	Level int
}

// IndexPageData contains the data necessary to render the documentation home page.
type IndexPageData struct {
	Title    string
	Entities []parser.EntityInfo
	Readme   template.HTML
	Initials string
	TOC      []TOCEntry
	Packages []PackageLink
}

// GenerateIndex creates the main entry point (index.html) of the documentation.
func GenerateIndex(title string, outputDir string, entities []parser.EntityInfo, readme string, initials string, toc []TOCEntry, packages []PackageLink) error {
	tmplPath := "pkg/generator/templates/index.html"
	tmplName := filepath.Base(tmplPath)

	tmpl, err := template.New(tmplName).Funcs(template.FuncMap{
		"contains": strings.Contains,
		"replace":  strings.ReplaceAll,
		"lower":    strings.ToLower,
		"html": func(s string) template.HTML {
			return template.HTML(s)
		},
	}).ParseFiles(tmplPath)
	if err != nil {
		return err
	}

	data := IndexPageData{
		Title:    title,
		Entities: entities,
		Readme:   template.HTML(readme),
		Initials: initials,
		TOC:      toc,
		Packages: packages,
	}

	outputPath := filepath.Join(outputDir, "index.html")
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}
