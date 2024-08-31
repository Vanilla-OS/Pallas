package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/russross/blackfriday/v2"
	"github.com/vanilla-os/pallas/pkg/generator"
	"github.com/vanilla-os/pallas/pkg/parser"
)

func main() {
	// Flags
	destDir := flag.String("dest", "", "Specify a custom destination directory for the output (default is './dist')")
	title := flag.String("title", "", "Specify a custom title for the documentation (default is the project root name)")
	readmePath := flag.String("readme", "", "Specify a custom README.md file to use for the index page (default is to search in project root)")
	flag.Parse()

	// Here we assume the project path is the first argument (if provided)
	projectPath := "."
	if len(flag.Args()) > 0 {
		projectPath = flag.Args()[0]
	}

	// Determine the absolute path of the project
	absProjectPath, err := filepath.Abs(projectPath)
	if err != nil {
		log.Fatalf("Error determining absolute path: %v", err)
	}

	// Determine the output directory
	var outputDir string
	if *destDir != "" {
		outputDir, err = filepath.Abs(*destDir)
		if err != nil {
			log.Fatalf("Error determining absolute path for output directory: %v", err)
		}
	} else {
		// Default to a 'dist' directory in the current working directory
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatalf("Error determining current working directory: %v", err)
		}
		outputDir = filepath.Join(cwd, "dist")
	}

	fmt.Printf("Documentation will be generated in: %s\n", outputDir)

	// Clean the output directory
	if err := os.RemoveAll(outputDir); err != nil {
		log.Fatalf("Error cleaning output directory: %v", err)
	}

	// Change to the project directory so as to correctly parse the packages
	if err := os.Chdir(absProjectPath); err != nil {
		log.Fatalf("Error changing directory: %v", err)
	}

	// Determine the title for the documentation
	docTitle := *title
	if docTitle == "" {
		docTitle = filepath.Base(absProjectPath)
	}

	// Read and convert README.md content to HTML
	readmeContent := readReadme(*readmePath, absProjectPath)

	// Here is where the magic happens (parsing and generating the documentation)
	fmt.Printf("Parsing project at path: %s\n", absProjectPath)
	packages, err := parser.GetPackages()
	if err != nil {
		log.Fatalf("Error fetching packages: %v", err)
	}

	// Generate HTML for each package
	for _, pkgPath := range packages {
		fmt.Printf("Parsing package: %s\n", pkgPath)
		entities, err := parser.ParseEntitiesInPackage(pkgPath)
		if err != nil {
			log.Fatalf("Error parsing package %s: %v", pkgPath, err)
		}

		err = generator.GenerateHTML(absProjectPath, pkgPath, entities, outputDir, docTitle)
		if err != nil {
			log.Fatalf("Error generating HTML for package %s: %v", pkgPath, err)
		}

		fmt.Printf("HTML generated for package: %s\n", pkgPath)
	}

	// Generate the index.html file
	packageNamesFull := make([]string, 0, len(packages))
	for _, pkgPath := range packages {
		relativePath, err := filepath.Rel(absProjectPath, pkgPath)
		if err != nil {
			log.Fatalf("Error determining relative path: %v", err)
		}
		packageNamesFull = append(packageNamesFull, relativePath)
	}

	err = generator.GenerateIndex(absProjectPath, packageNamesFull, outputDir, docTitle, readmeContent)
	if err != nil {
		log.Fatalf("Error generating index.html: %v", err)
	}

	fmt.Printf("Documentation index generated in %s/index.html\n", outputDir)
}

// markdownToHTML converts markdown content to HTML and applies Tailwind CSS classes
func markdownToHTML(markdown string) string {
	htmlContent := blackfriday.Run([]byte(markdown))
	htmlString := string(htmlContent)

	// Tailwind CSS classes
	htmlString = strings.ReplaceAll(htmlString, "<h1", `<h1 class="text-3xl font-bold mb-4"`)
	htmlString = strings.ReplaceAll(htmlString, "<h2", `<h2 class="text-2xl font-bold mb-4"`)
	htmlString = strings.ReplaceAll(htmlString, "<h3", `<h3 class="text-xl font-bold mb-4"`)
	htmlString = strings.ReplaceAll(htmlString, "<p>", `<p class="text-gray-700 dark:text-gray-300 mb-4">`)
	htmlString = strings.ReplaceAll(htmlString, "<ul>", `<ul class="list-disc ml-6 mb-4">`)
	htmlString = strings.ReplaceAll(htmlString, "<ol>", `<ol class="list-decimal ml-6 mb-4">`)
	htmlString = strings.ReplaceAll(htmlString, "<li>", `<li class="mb-2">`)
	htmlString = strings.ReplaceAll(htmlString, "<pre>", `<pre class="bg-gray-800 text-white rounded-lg p-4 overflow-auto mb-4">`)
	htmlString = strings.ReplaceAll(htmlString, "<code>", `<code class="bg-gray-100 dark:bg-gray-800 rounded px-1 hljs">`)
	htmlString = strings.ReplaceAll(htmlString, "<blockquote>", `<blockquote class="border-l-4 border-gray-300 pl-4 italic mb-4">`)
	htmlString = strings.ReplaceAll(htmlString, "<table>", `<table class="border-collapse border border-gray-300 w-full mb-4">`)
	htmlString = strings.ReplaceAll(htmlString, "<th>", `<th class="border border-gray-300 bg-gray-100 dark:bg-gray-800 p-2">`)
	htmlString = strings.ReplaceAll(htmlString, "<td>", `<td class="border border-gray-300 p-2">`)
	htmlString = strings.ReplaceAll(htmlString, "<a ", `<a class="text-blue-500 hover:underline" target="_blank" `)

	// Fixes
	re := regexp.MustCompile(`<img[^>]*src="([^"]*)"[^>]*>`)
	htmlString = re.ReplaceAllStringFunc(htmlString, func(imgTag string) string {
		matches := re.FindStringSubmatch(imgTag)
		if len(matches) > 1 && !(strings.HasPrefix(matches[1], "http://") || strings.HasPrefix(matches[1], "https://")) {
			return ""
		}
		return imgTag
	})

	return htmlString
}

// readReadme reads the README.md file and converts it to HTML
func readReadme(customPath, projectRoot string) string {
	var readmePath string
	if customPath != "" {
		readmePath = customPath
	} else {
		readmePath = filepath.Join(projectRoot, "README.md")
	}

	content, err := os.ReadFile(readmePath)
	if err != nil {
		fmt.Println("README.md not found, generating default content...")
		return markdownToHTML(generateDefaultReadme())
	}

	return markdownToHTML(string(content))
}

// generateDefaultReadme generates a default README.md content
func generateDefaultReadme() string {
	return `# Welcome to the Documentation

This is the autogenerated documentation for the project. You can provide a custom README.md file to replace this content.

## Instructions to Replace

1. Create a README.md file in the root of your project.
2. Add content to it following standard Markdown syntax.
3. Re-run the documentation generation with the --readme flag pointing to your README.md file (Pallas will detect it automatically if it's in the root).

For more information, visit the [Pallas](https://github.com/vanilla-os/pallas) GitHub repository.`
}
