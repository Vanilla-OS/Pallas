package generator

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/vanilla-os/pallas/pkg/parser"
)

// PageData describes the data structure passed to the package entities template.
type PageData struct {
	Title           string
	Entities        []parser.EntityInfo
	Imports         []parser.ImportInfo
	Readme          template.HTML
	FileGroups      map[string][]parser.EntityInfo
	Initials        string
	PackageName     string
	AllPackages     []PackageLink
	TypeIndex       map[string]TypeInfo
	ProjectPackages []string
	ModulePath      string
	DocPages        []DocPage
	DocSections     []DocSection
}

// DocPage represents a documentation page (e.g., from docs/ folder).
type DocPage struct {
	Title    string
	Filename string
	Content  template.HTML
	ID       string
	Section  string // Parent section name (empty for root docs)
}

// DocSection represents a collapsible section of documentation pages.
type DocSection struct {
	Name     string
	Pages    []DocPage
	HasIndex bool
}

// TypeInfo stores cross-reference data for a Go type.
type TypeInfo struct {
	Name    string
	Package string
	Link    string
}

// PackageLink represents a navigation link to a package documentation page.
type PackageLink struct {
	Name      string
	Link      string
	IsCurrent bool
}

// GenerateHTML orchestrates the generation of HTML documentation for all packages in the project.
func GenerateHTML(title string, outputDir string, entities []parser.EntityInfo, imports []parser.ImportInfo, readme string, initials string, modulePath string, docPages []DocPage, docSections []DocSection) error {
	packageGroups := make(map[string][]parser.EntityInfo)
	for _, e := range entities {
		pkg := e.Package
		if pkg == "" {
			pkg = "main"
		}
		packageGroups[pkg] = append(packageGroups[pkg], e)
	}

	var allPackages []PackageLink
	for pkg := range packageGroups {
		allPackages = append(allPackages, PackageLink{
			Name: pkg,
			Link: packageFilename(pkg),
		})
	}

	typeIndex := make(map[string]TypeInfo)
	for _, e := range entities {
		pkg := e.Package
		if pkg == "" {
			pkg = "main"
		}
		if e.Type == "struct" || e.Type == "interface" || e.Type == "type" {
			info := TypeInfo{
				Name:    e.Name,
				Package: pkg,
				Link:    packageFilename(pkg) + "#" + e.Name,
			}
			typeIndex[pkg+"."+e.Name] = info
			if _, exists := typeIndex[e.Name]; !exists {
				typeIndex[e.Name] = info
			}
		}
	}

	var projectPackages []string
	for pkg := range packageGroups {
		projectPackages = append(projectPackages, pkg)
	}

	// docPages here contains ALL pages for generation, but we filter root-only for sidebar
	var rootDocPages []DocPage
	for _, p := range docPages {
		if p.Section == "" {
			rootDocPages = append(rootDocPages, p)
		}
	}

	for pkg, pkgEntities := range packageGroups {
		if err := generatePackagePage(title, outputDir, pkg, pkgEntities, imports, initials, allPackages, modulePath, typeIndex, projectPackages, rootDocPages, docSections); err != nil {
			return err
		}
	}

	if err := GenerateDocPages(title, outputDir, docPages, rootDocPages, initials, allPackages, modulePath, typeIndex, projectPackages, docSections); err != nil {
		return err
	}

	return GenerateSearchIndex(outputDir, entities, docPages)
}

// packageFilename generates a safe filename for a package's documentation page.
func packageFilename(pkg string) string {
	safe := regexp.MustCompile(`[^a-zA-Z0-9_-]`).ReplaceAllString(pkg, "_")
	return "pkg_" + safe + ".html"
}

// generatePackagePage generates a single HTML page for a specific package.
func generatePackagePage(title string, outputDir string, pkg string, entities []parser.EntityInfo, imports []parser.ImportInfo, initials string, allPackages []PackageLink, modulePath string, typeIndex map[string]TypeInfo, projectPackages []string, docPages []DocPage, docSections []DocSection) error {
	tmplPath := "templates/entities.html"
	tmplName := filepath.Base(tmplPath)

	tmpl, err := template.New(tmplName).Funcs(template.FuncMap{
		"contains": strings.Contains,
		"replace":  strings.ReplaceAll,
		"lower":    strings.ToLower,
		"lineCount": func(s string) int {
			return strings.Count(s, "\n") + 1
		},
		"html": func(s string) template.HTML {
			return template.HTML(s)
		},
		"json": func(v interface{}) string {
			b, _ := json.Marshal(v)
			return string(b)
		},
		"pkgGoDevURL": func(importPath string) string {
			return "https://pkg.go.dev/" + importPath
		},
		"isStdLib": func(importPath string) bool {
			parts := strings.Split(importPath, "/")
			return len(parts) > 0 && !strings.Contains(parts[0], ".")
		},
		"isInternalImport": func(importPath string) bool {
			return modulePath != "" && strings.HasPrefix(importPath, modulePath)
		},
		"internalImportLink": func(importPath string) string {
			parts := strings.Split(importPath, "/")
			if len(parts) > 0 {
				pkgName := parts[len(parts)-1]
				return packageFilename(pkgName)
			}
			return "#"
		},
		"packageLink": func(pkgName string) string {
			return packageFilename(pkgName)
		},
		"splitParam": func(param string) struct{ Name, Type string } {
			parts := strings.Fields(param)
			if len(parts) == 1 {
				return struct{ Name, Type string }{Name: "", Type: parts[0]}
			}
			if len(parts) >= 2 {
				return struct{ Name, Type string }{Name: parts[0], Type: strings.Join(parts[1:], " ")}
			}
			return struct{ Name, Type string }{Name: "", Type: ""}
		},
		"linkType": func(typeName string) template.HTML {
			orig := typeName
			prefix := ""
			base := typeName

			for {
				if strings.HasPrefix(base, "*") {
					prefix += "*"
					base = base[1:]
				} else if strings.HasPrefix(base, "[]") {
					prefix += "[]"
					base = base[2:]
				} else {
					break
				}
			}

			if info, ok := typeIndex[base]; ok {
				return template.HTML(fmt.Sprintf(`%s<a href="%s" class="text-brand-600 dark:text-brand-400 hover:underline">%s</a>`, prefix, info.Link, base))
			}

			if strings.Contains(base, ".") {
				parts := strings.Split(base, ".")
				pkgName := parts[0]
				typeNameOnly := parts[1]

				for _, imp := range imports {
					isMatch := (imp.Alias != "" && imp.Alias == pkgName) ||
						(imp.Alias == "" && (strings.HasSuffix(imp.Path, "/"+pkgName) || imp.Path == pkgName))

					if isMatch {
						url := "https://pkg.go.dev/" + imp.Path + "#" + typeNameOnly
						return template.HTML(fmt.Sprintf(`%s<a href="%s" target="_blank" class="text-zinc-600 dark:text-zinc-400 hover:text-brand-600 dark:hover:text-brand-400 underline decoration-dotted underline-offset-2">%s</a>`, prefix, url, base))
					}
				}
			}

			return template.HTML(template.HTMLEscapeString(orig))
		},
		"linkTypes": func(code string) template.HTML {
			result := template.HTMLEscapeString(code)
			for typeName, info := range typeIndex {
				pattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(typeName) + `\b`)
				link := fmt.Sprintf(`<a href="%s" class="text-brand-600 dark:text-brand-400 hover:underline">%s</a>`, info.Link, typeName)
				result = pattern.ReplaceAllString(result, link)
			}
			return template.HTML(result)
		},
	}).ParseFS(templatesFS, tmplPath)
	if err != nil {
		return err
	}

	fileGroups := make(map[string][]parser.EntityInfo)
	for _, e := range entities {
		f := filepath.Base(e.File)
		if f == "" || e.File == "" {
			f = "Globals"
		}
		fileGroups[f] = append(fileGroups[f], e)
	}

	seenImports := make(map[string]bool)
	var pkgImports []parser.ImportInfo
	for _, imp := range imports {
		if imp.Package == pkg && !seenImports[imp.Path] {
			seenImports[imp.Path] = true
			pkgImports = append(pkgImports, imp)
		}
	}

	navPackages := make([]PackageLink, len(allPackages))
	for i, p := range allPackages {
		navPackages[i] = PackageLink{
			Name:      p.Name,
			Link:      p.Link,
			IsCurrent: p.Name == pkg,
		}
	}

	data := PageData{
		Title:           title,
		Entities:        entities,
		Imports:         pkgImports,
		FileGroups:      fileGroups,
		Initials:        initials,
		PackageName:     pkg,
		AllPackages:     navPackages,
		TypeIndex:       typeIndex,
		ProjectPackages: projectPackages,
		ModulePath:      modulePath,
		DocPages:        docPages,
		DocSections:     docSections,
	}

	outputPath := filepath.Join(outputDir, packageFilename(pkg))
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}

// GenerateDocPages generates HTML pages for the documentation files found in the docs/ directory.
func GenerateDocPages(title string, outputDir string, allDocPages []DocPage, rootDocPages []DocPage, initials string, allPackages []PackageLink, modulePath string, typeIndex map[string]TypeInfo, projectPackages []string, docSections []DocSection) error {
	tmplPath := "templates/page.html"
	tmplName := filepath.Base(tmplPath)

	tmpl, err := template.New(tmplName).Funcs(template.FuncMap{
		"contains": strings.Contains,
		"replace":  strings.ReplaceAll,
		"lower":    strings.ToLower,
		"json": func(v interface{}) string {
			b, _ := json.Marshal(v)
			return string(b)
		},
		"html": func(s string) template.HTML {
			return template.HTML(s)
		},
		"linkTypes": func(input interface{}) template.HTML {
			var code string
			switch v := input.(type) {
			case string:
				code = v
			case template.HTML:
				code = string(v)
			default:
				return ""
			}

			// If it matches [Package.Type], replace it
			re := regexp.MustCompile(`\[([a-zA-Z0-9_]+)\.([a-zA-Z0-9_]+)\]`)
			code = re.ReplaceAllStringFunc(code, func(match string) string {
				parts := re.FindStringSubmatch(match)
				if len(parts) == 3 {
					pkg := parts[1]
					typ := parts[2]
					key := pkg + "." + typ
					if info, ok := typeIndex[key]; ok {
						return fmt.Sprintf(`<a href="%s" class="text-brand-600 dark:text-brand-400 hover:underline">%s</a>`, info.Link, typ)
					}
				}
				return match
			})

			return template.HTML(code)
		},
	}).ParseFS(templatesFS, tmplPath)
	if err != nil {
		return err
	}

	for _, page := range allDocPages {
		data := PageData{
			Title:           title,
			Readme:          page.Content,
			Initials:        initials,
			PackageName:     page.Title,
			AllPackages:     allPackages,
			TypeIndex:       typeIndex,
			ProjectPackages: projectPackages,
			ModulePath:      modulePath,
			DocPages:        rootDocPages,
			DocSections:     docSections,
		}

		outputPath := filepath.Join(outputDir, page.Filename)
		f, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("error creating doc page %s: %v", page.Filename, err)
		}

		if err := tmpl.Execute(f, data); err != nil {
			f.Close()
			return fmt.Errorf("error executing template for %s: %v", page.Filename, err)
		}
		f.Close()
	}

	return nil
}

// GenerateSearchIndex creates a JSON file used for client-side documentation search.
func GenerateSearchIndex(outputDir string, entities []parser.EntityInfo, docPages []DocPage) error {
	type SearchItem struct {
		Name        string `json:"name"`
		Type        string `json:"type"`
		Description string `json:"description"`
		Link        string `json:"link"`
		Package     string `json:"package"`
	}

	var items []SearchItem
	for _, e := range entities {
		pkg := e.Package
		if pkg == "" {
			pkg = "main"
		}
		link := packageFilename(pkg) + "#" + e.Name

		items = append(items, SearchItem{
			Name:        e.Name,
			Type:        e.Type,
			Description: e.DescriptionRaw,
			Link:        link,
			Package:     pkg,
		})

		for _, m := range e.Methods {
			items = append(items, SearchItem{
				Name:        e.Name + "." + m.Name,
				Type:        "method",
				Description: m.DescriptionRaw,
				Link:        link,
				Package:     pkg,
			})
		}
	}

	for _, p := range docPages {
		items = append(items, SearchItem{
			Name:        p.Title,
			Type:        "page",
			Description: "Documentation page: " + p.Title,
			Link:        p.Filename,
			Package:     "docs",
		})
	}

	f, err := os.Create(filepath.Join(outputDir, "search.json"))
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(items)
}
