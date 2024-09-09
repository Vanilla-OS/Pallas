package generator

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Copy static assets (such as CSS, JS, images) from the "templates/static" directory 
// to the output directory.
// Create the required directories and copy files recursively.
//
// Returns: An error if any occurs during the directory creation or file copying, otherwise nil
func CopyStaticAssets(outputDir string) error {
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return fmt.Errorf("error creating output directory: %v", err)
	}

	// Create the 'static' directory in the output location
	dstPath := filepath.Join(outputDir, "static")
	if err := os.MkdirAll(dstPath, os.ModePerm); err != nil {
		return fmt.Errorf("error creating static directory: %v", err)
	}

	assets, err := staticAssets.ReadDir("templates/static")
	if err != nil {
		return fmt.Errorf("error reading static assets: %v", err)
	}

	for _, asset := range assets {
		srcPath := filepath.Join("templates/static", asset.Name())
		dstAssetPath := filepath.Join(dstPath, asset.Name())

		if asset.IsDir() {
			// If it's a directory, recursively copy its contents
			if err := CopyStaticAssets(filepath.Join(outputDir, "static", asset.Name())); err != nil {
				return fmt.Errorf("error copying directory: %v", err)
			}
		} else {
			// If it's a file, copy it
			if err := copyFile(srcPath, dstAssetPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// Copy a file from the source path to the destination path.
// Handle opening, creating, and copying the file contents.
//
// Returns: An error if any occurs during file operations, otherwise nil
func copyFile(srcPath, dstPath string) error {
	src, err := staticAssets.Open(srcPath)
	if err != nil {
		return fmt.Errorf("error opening static asset: %v", err)
	}
	defer src.Close()

	dst, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("error copying file: %v", err)
	}

	return nil
}
