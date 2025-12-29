package generator

// CopyStaticAssets is a legacy function. Pallas now uses CDNs for static assets.
//
// Deprecated: No longer needed with the transition to CDN-based assets.
func CopyStaticAssets(outputDir string) error {
	return nil
}

// copyFile is a legacy helper function.
//
// Deprecated: No longer needed.
func copyFile(srcPath, dstPath string) error {
	return nil
}
