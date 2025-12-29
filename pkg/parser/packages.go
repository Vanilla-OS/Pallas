package parser

import (
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

// GetPackages identifies and returns absolute paths to all Go package directories in the project.
func GetPackages(projectRoot string) ([]string, error) {
	absRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return nil, err
	}

	cfg := &packages.Config{
		Mode: packages.NeedFiles,
		Dir:  absRoot,
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, err
	}

	var packageDirs []string
	for _, pkg := range pkgs {
		if len(pkg.GoFiles) > 0 {
			dir := filepath.Dir(pkg.GoFiles[0])

			// We intentionally skip the root directory to focus on subpackages
			if dir == absRoot {
				continue
			}
			packageDirs = append(packageDirs, dir)
		}
	}

	return packageDirs, nil
}
