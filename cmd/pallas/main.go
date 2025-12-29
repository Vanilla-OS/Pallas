package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mirkobrombin/go-cli-builder/v2/cli"
	"github.com/russross/blackfriday/v2"
	"github.com/vanilla-os/pallas/pkg/generator"
	"github.com/vanilla-os/pallas/pkg/parser"
)

// PallasConfig holds the configuration for the documentation generation.
type PallasConfig struct {
	Dest    string
	Title   string
	Readme  string
	Project string
}

type CLI struct {
	Generate GenerateCmd `cmd:"" help:"Generate documentation"`
	Version  VersionCmd  `cmd:"" help:"Print version information"`

	cli.Base
}

type VersionCmd struct {
	cli.Base
}

func (c *VersionCmd) Run() error {
	fmt.Println("0.0.1")
	return nil
}

type GenerateCmd struct {
	Dest        string `cli:"dest,d" help:"Destination directory" default:"./dist"`
	Title       string `cli:"title,t" help:"Project title" default:"Pallas"`
	Readme      string `cli:"readme,r" help:"Readme file" default:"README.md"`
	ProjectFlag string `cli:"project,p" help:"Project root" default:"."`
	ProjectArg  string `arg:"" optional:"true" name:"project-path" help:"Project root path"`

	cli.Base
}

func (c *GenerateCmd) Run() error {
	config := &PallasConfig{
		Dest:    c.Dest,
		Title:   c.Title,
		Readme:  c.Readme,
		Project: c.ProjectFlag,
	}

	if c.ProjectArg != "" {
		config.Project = c.ProjectArg
	}

	projectPath := config.Project
	if projectPath == "" {
		projectPath = "."
	}
	projectPath, _ = filepath.Abs(projectPath)

	outputDir := config.Dest
	if outputDir == "." || outputDir == "" || outputDir == "/" {
		outputDir = "./dist"
	}
	outputDir, _ = filepath.Abs(outputDir)

	if strings.HasPrefix(projectPath, outputDir) && projectPath == outputDir {
		return fmt.Errorf("output directory cannot be the same as project directory")
	}

	fmt.Println("Generating documentation for:", projectPath)
	fmt.Println("Output directory:", outputDir)

	if err := os.RemoveAll(outputDir); err != nil {
		return fmt.Errorf("error cleaning output directory: %v", err)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("error creating output directory: %v", err)
	}

	pkgDirs, err := parser.GetPackages(projectPath)
	if err != nil {
		return fmt.Errorf("error finding packages: %v", err)
	}

	var entities []parser.EntityInfo
	var imports []parser.ImportInfo
	seenPackages := make(map[string]bool)

	for _, pkgDir := range pkgDirs {
		rel, _ := filepath.Rel(projectPath, pkgDir)
		ents, imps, err := parser.ParseEntitiesInPackage(projectPath, pkgDir, rel)
		if err != nil {
			fmt.Printf("Warning: failed to parse package %s: %v\n", rel, err)
			continue
		}

		if len(ents) > 0 {
			for i := range ents {
				if relFile, err := filepath.Rel(projectPath, ents[i].File); err == nil {
					ents[i].File = relFile
				}
			}
		}

		entities = append(entities, ents...)
		imports = append(imports, imps...)
	}

	var packageLinks []generator.PackageLink
	for _, e := range entities {
		pkg := e.Package
		if pkg == "" {
			pkg = "main"
		}
		if !seenPackages[pkg] {
			seenPackages[pkg] = true
			packageLinks = append(packageLinks, generator.PackageLink{
				Name: pkg,
				Link: "pkg_" + sanitizePackageName(pkg) + ".html",
			})
		}
	}

	fmt.Printf("Found %d entities and %d imports\n", len(entities), len(imports))

	readmePath := filepath.Join(projectPath, config.Readme)
	readmeContent := ""
	readmeRaw := ""

	// Attempt to locate and read the README file
	if _, err := os.Stat(readmePath); err == nil {
		content, _ := os.ReadFile(readmePath)
		readmeRaw = string(content)
		readmeContent = markdownToHTML(readmeRaw)
	} else {
		readmePath = filepath.Join(projectPath, strings.ToLower(config.Readme))
		if _, err := os.Stat(readmePath); err == nil {
			content, _ := os.ReadFile(readmePath)
			readmeRaw = string(content)
			readmeContent = markdownToHTML(readmeRaw)
		}
	}

	toc := extractTOC(readmeRaw)

	// Auto-detect project title from directory name if default is used
	if config.Title == "Pallas" {
		base := filepath.Base(projectPath)
		if base != "." && base != "/" {
			if len(base) > 0 {
				clean := strings.ReplaceAll(base, "-", " ")
				clean = strings.ReplaceAll(clean, "_", " ")
				clean = strings.ReplaceAll(clean, ".", " ")
				config.Title = strings.Title(clean)
			} else {
				config.Title = base
			}
		}
	}

	initials := getInitials(config.Title)
	modulePath := getModulePath(projectPath)

	if err := generator.GenerateHTML(config.Title, config.Dest, entities, imports, readmeContent, initials, modulePath); err != nil {
		return fmt.Errorf("error generating HTML: %v", err)
	}

	if err := generator.GenerateIndex(config.Title, outputDir, entities, readmeContent, initials, toc, packageLinks); err != nil {
		return fmt.Errorf("error generating index: %v", err)
	}

	fmt.Println("Documentation generated successfully!")
	return nil
}

func main() {
	args := os.Args[1:]
	inject := true

	// If the user provided a command, we don't inject "generate"
	if len(args) > 0 {
		cmd := args[0]
		if !strings.HasPrefix(cmd, "-") {
			switch cmd {
			case "version", "completion", "help", "generate":
				inject = false
			}
		}
	}

	if inject {
		newArgs := append([]string{os.Args[0], "generate"}, args...)
		os.Args = newArgs
	}

	app := &CLI{}
	if err := cli.Run(app); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// getInitials extracts a short representation (initials) from the project title.
func getInitials(title string) string {
	title = strings.TrimPrefix(title, "http://")
	title = strings.TrimPrefix(title, "https://")

	f := func(c rune) bool {
		return c == ' ' || c == '-' || c == '_' || c == '.' || c == '/'
	}
	parts := strings.FieldsFunc(title, f)

	var validParts []string
	for _, p := range parts {
		if len(p) > 0 {
			validParts = append(validParts, p)
		}
	}

	if len(validParts) == 0 {
		return "P"
	}

	if len(validParts) >= 2 {
		return strings.ToUpper(string(validParts[0][0]) + string(validParts[1][0]))
	}

	s := validParts[0]
	if len(s) == 0 {
		return "P"
	}

	var uppers []rune
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			uppers = append(uppers, r)
		}
	}

	if len(uppers) >= 2 {
		return string(uppers[0]) + string(uppers[1])
	}

	return strings.ToUpper(string(s[0]))
}

// markdownToHTML converts markdown text to HTML.
func markdownToHTML(md string) string {
	renderer := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
		Flags: blackfriday.CommonHTMLFlags,
	})
	extensions := blackfriday.CommonExtensions | blackfriday.AutoHeadingIDs | blackfriday.HardLineBreak | blackfriday.Autolink
	output := blackfriday.Run([]byte(md), blackfriday.WithRenderer(renderer), blackfriday.WithExtensions(extensions))
	return string(output)
}

// extractTOC parses the markdown and extracts a Table of Contents.
func extractTOC(md string) []generator.TOCEntry {
	var toc []generator.TOCEntry
	node := blackfriday.New(blackfriday.WithExtensions(blackfriday.CommonExtensions | blackfriday.AutoHeadingIDs)).Parse([]byte(md))

	node.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if entering && node.Type == blackfriday.Heading {
			anchor := string(node.HeadingData.HeadingID)

			var titleParts []string
			for child := node.FirstChild; child != nil; child = child.Next {
				if child.Type == blackfriday.Text || child.Type == blackfriday.Code {
					titleParts = append(titleParts, string(child.Literal))
				}
			}
			title := strings.Join(titleParts, "")

			if anchor == "" {
				anchor = blackfriday.SanitizedAnchorName(title)
			}

			toc = append(toc, generator.TOCEntry{
				ID:    anchor,
				Title: title,
				Level: node.HeadingData.Level,
			})
		}
		return blackfriday.GoToNext
	})
	return toc
}

// sanitizePackageName ensures the package name is safe for file names.
func sanitizePackageName(pkg string) string {
	return regexp.MustCompile(`[^a-zA-Z0-9_-]`).ReplaceAllString(pkg, "_")
}

// getModulePath reads the go.mod file and extracts the module path.
func getModulePath(projectPath string) string {
	goModPath := filepath.Join(projectPath, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return ""
	}
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return ""
}
