package parser

import (
	"os/exec"
	"strings"
)

// GetPackages returns a list of all packages in the project
//
// Example:
//
//	packages, err := parser.GetPackages()
//	if err != nil {
//		log.Fatalf("Error fetching packages: %v", err)
//	}
//	for _, pkgPath := range packages {
//		fmt.Printf("- %s\n", pkgPath)
//	}
func GetPackages() ([]string, error) {
	cmd := exec.Command("go", "list", "-f", "{{.Dir}}", "./...")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Split the output by newline to get each package path
	packages := strings.Split(string(out), "\n")

	// Remove any empty strings from the list
	var cleanedPackages []string
	for _, pkg := range packages {
		if pkg != "" {
			cleanedPackages = append(cleanedPackages, pkg)
		}
	}

	return cleanedPackages, nil
}
