package datasource

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// NewEmbeddedDataSource creates a data source from the demo bundle
// Auto-discovers the newest bundle in example-log-bundle/ directory
func NewEmbeddedDataSource(verbose bool) (DataSource, error) {
	bundlePath, err := findNewestBundle("example-log-bundle", verbose)
	if err != nil {
		return nil, fmt.Errorf("failed to find demo bundle: %w\n\n"+
			"The example-log-bundle/ directory may be missing or empty.\n"+
			"Try using --bundle with a specific bundle path instead", err)
	}

	if verbose {
		fmt.Printf("🎯 Auto-discovered bundle: %s\n", bundlePath)
	}

	return NewBundleDataSource(bundlePath, verbose)
}

// findNewestBundle scans a directory for bundle folders and returns the newest by timestamp
func findNewestBundle(dir string, verbose bool) (string, error) {
	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return "", fmt.Errorf("directory not found: %s", dir)
	}

	// Read directory entries
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %w", err)
	}

	// Filter for bundle directories (contain timestamp pattern)
	var bundles []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Bundle pattern: *-YYYY-MM-DD_HH_MM_SS
		// Example: w-guard-wg-cp-svtk6-lqtxw-2025-12-04_09_15_57
		if strings.Contains(name, "-20") && strings.Contains(name, "_") {
			bundles = append(bundles, name)
		}
	}

	if len(bundles) == 0 {
		return "", fmt.Errorf("no bundle directories found in %s", dir)
	}

	// Sort bundles by name (timestamp is at end, so lexical sort works)
	sort.Strings(bundles)

	// Return the newest (last in sorted list)
	newestBundle := bundles[len(bundles)-1]
	bundlePath := filepath.Join(dir, newestBundle)

	if verbose {
		fmt.Printf("Found %d bundle(s), selected newest: %s\n", len(bundles), newestBundle)
	}

	return bundlePath, nil
}
