package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/vanilla-os/pallas/pkg/generator"
	"github.com/vanilla-os/pallas/pkg/parser"
)

func main() {
	// Flags
	destDir := flag.String("dest", "", "Specify a custom destination directory for the output (default is './dist')")
	title := flag.String("title", "", "Specify a custom title for the documentation (default is the project root name)")
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

	// Here is were the magic happens (parsing and generating the documentation)
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

	err = generator.GenerateIndex(absProjectPath, packageNamesFull, outputDir, docTitle)
	if err != nil {
		log.Fatalf("Error generating index.html: %v", err)
	}

	fmt.Printf("Documentation index generated in %s/index.html\n", outputDir)
}
