package bundle

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Rancheroo/r8s/internal/rancher"
)

// LoadFromPath loads a bundle from an extracted directory.
// Tarball support has been removed - users must extract bundles first.
func LoadFromPath(path string, opts ImportOptions) (*Bundle, error) {
	// Step 1: Validate and resolve path
	absPath, pathInfo, err := validateAndResolvePath(path, opts.Verbose)
	if err != nil {
		return nil, err
	}

	// Step 2: Verify it's a directory
	if !pathInfo.IsDir() {
		if opts.Verbose {
			return nil, fmt.Errorf("%s is not a directory\n\n"+
				"r8s only supports extracted bundle folders.\n\n"+
				"If you have a .tar.gz file, extract it first:\n"+
				"  tar -xzf %s\n"+
				"  r8s ./extracted-folder/\n\n"+
				"HINT: Point r8s at the extracted bundle directory, not the archive file",
				path, filepath.Base(path))
		}
		return nil, fmt.Errorf("%s is not a directory - extract the bundle first (tar -xzf bundle.tar.gz)", path)
	}

	if opts.Verbose {
		fmt.Printf("📁 Loading bundle from: %s\n", absPath)
	}

	// Step 3: Validate bundle structure
	if err := ValidateBundle(absPath); err != nil {
		return nil, err
	}

	// Step 4: Load bundle from directory
	bundle, err := loadFromExtractedPath(absPath, absPath, 0, opts)
	if err != nil {
		return nil, err
	}

	// Bundle is already extracted, no cleanup needed
	bundle.IsTemporary = false

	return bundle, nil
}

// validateAndResolvePath validates the path exists and resolves it to absolute path
func validateAndResolvePath(path string, verbose bool) (string, os.FileInfo, error) {
	if path == "" {
		if verbose {
			return "", nil, fmt.Errorf("bundle path is required\n\n" +
				"USAGE:\n" +
				"  r8s ./extracted-bundle-folder/\n\n" +
				"HINT: Provide an extracted bundle directory")
		}
		return "", nil, fmt.Errorf("bundle path is required")
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		if verbose {
			return "", nil, fmt.Errorf("invalid path format: %w\n\n"+
				"Provided: %s\n"+
				"HINT: Check for special characters or invalid path syntax", err, path)
		}
		return "", nil, fmt.Errorf("invalid path: %w", err)
	}

	// Check if exists
	info, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		if verbose {
			cwd, _ := os.Getwd()
			return "", nil, fmt.Errorf("path not found: %s\n\n"+
				"Current directory: %s\n"+
				"Absolute path tried: %s\n\n"+
				"TROUBLESHOOTING:\n"+
				"  1. Check the path is correct\n"+
				"  2. Ensure folder exists\n"+
				"  3. Check directory permissions\n"+
				"  4. Try using an absolute path\n\n"+
				"REMINDER: If you have a .tar.gz file, extract it first:\n"+
				"  tar -xzf bundle.tar.gz", path, cwd, absPath)
		}
		return "", nil, fmt.Errorf("path not found: %s", path)
	}
	if err != nil {
		return "", nil, fmt.Errorf("failed to access path: %w", err)
	}

	return absPath, info, nil
}

func loadFromExtractedPath(extractPath, originalPath string, size int64, opts ImportOptions) (*Bundle, error) {
	if opts.Verbose {
		fmt.Println("Parsing bundle data...")
	}

	// Initialize health tracking
	health := &BundleHealth{
		TotalFiles:   7, // pods, deployments, services, namespaces, events, crds, logs
		FoundFiles:   0,
		DerivedFiles: 0,
		MissingFiles: make([]string, 0),
		Warnings:     make([]string, 0),
	}

	// Parse manifest
	manifest, err := ParseManifest(extractPath)
	if err != nil {
		if opts.Verbose {
			return nil, fmt.Errorf("failed to parse manifest: %w\n\n"+
				"Expected: metadata.json in bundle root\n"+
				"Searched: %s\n\n"+
				"This may not be a valid RKE2 support bundle", err, extractPath)
		}
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Inventory pods
	pods := loadResource(extractPath, InventoryPods, "pods", health, opts.Verbose)

	// Inventory log files
	logFiles := loadResource(extractPath, InventoryLogFiles, "logs", health, opts.Verbose)

	// Parse kubectl resources (all optional)
	crds := loadResource(extractPath, ParseCRDs, "crds", health, opts.Verbose)
	deployments := loadResource(extractPath, ParseDeployments, "deployments", health, opts.Verbose)
	services := loadResource(extractPath, ParseServices, "services", health, opts.Verbose)
	events := loadResource(extractPath, ParseEvents, "events", health, opts.Verbose)

	namespaces, nsErr := ParseNamespaces(extractPath)
	kubectlPods, _ := ParsePods(extractPath)

	// RESILIENCE: If namespaces file missing (common in partial bundles), derive from pods
	if nsErr != nil && len(namespaces) == 0 && len(kubectlPods) > 0 {
		// Extract unique namespaces from pod list
		nsMap := make(map[string]bool)
		for _, pod := range kubectlPods {
			if pod.NamespaceID != "" {
				nsMap[pod.NamespaceID] = true
			}
		}

		// Create namespace objects
		for nsName := range nsMap {
			namespaces = append(namespaces, rancher.Namespace{
				Name:      nsName,
				State:     "active",
				ClusterID: "bundle",
				ProjectID: "bundle-project",
			})
		}

		if len(namespaces) > 0 {
			health.DerivedFiles++
			health.Warnings = append(health.Warnings, fmt.Sprintf("namespaces: derived %d from pods", len(namespaces)))
			if opts.Verbose {
				fmt.Printf("⚠️  namespaces: derived %d from pods\n", len(namespaces))
			}
		} else {
			health.MissingFiles = append(health.MissingFiles, "namespaces")
		}
	} else if nsErr != nil {
		health.MissingFiles = append(health.MissingFiles, "namespaces")
		if opts.Verbose {
			fmt.Printf("⚠️  namespaces: missing\n")
		}
	} else {
		health.FoundFiles++
		if opts.Verbose {
			fmt.Printf("✓ namespaces: %d\n", len(namespaces))
		}
	}

	// Convert to interfaces for storage
	var crdsI, deploymentsI, servicesI, namespacesI, eventsI []interface{}
	for i := range crds {
		crdsI = append(crdsI, crds[i])
	}
	for i := range deployments {
		deploymentsI = append(deploymentsI, deployments[i])
	}
	for i := range services {
		servicesI = append(servicesI, services[i])
	}
	for i := range namespaces {
		namespacesI = append(namespacesI, namespaces[i])
	}
	for i := range events {
		eventsI = append(eventsI, events[i])
	}

	if opts.Verbose {
		fmt.Printf("✓ Loaded: %d pods, %d logs, %d kubectl pods, %d events, %d deployments, %d services, %d CRDs, %d namespaces\n",
			len(pods), len(logFiles), len(kubectlPods), len(events), len(deployments), len(services), len(crds), len(namespaces))
	}

	// Detect format and create path resolver
	format := DetectFormat(extractPath)
	bundleRoot := getBundleRoot(extractPath)
	pathResolver := NewPathResolver(bundleRoot, format)

	if opts.Verbose {
		fmt.Printf("✓ Format: %s, Distro: %s\n", format, pathResolver.GetDistro())
	}

	// Create bundle
	bundle := &Bundle{
		Path:         originalPath,
		ExtractPath:  extractPath,
		Manifest:     manifest,
		Pods:         pods,
		LogFiles:     logFiles,
		CRDs:         crdsI,
		Deployments:  deploymentsI,
		Services:     servicesI,
		Namespaces:   namespacesI,
		Events:       eventsI,
		Loaded:       true,
		Size:         size,
		IsTemporary:  false, // Bundles are already extracted, never temporary
		Health:       health,
		PathResolver: pathResolver,
	}

	return bundle, nil
}

// loadResource is a generic helper to load resources and handle errors
func loadResource[T any](
	path string,
	parseFunc func(string) ([]T, error),
	name string,
	health *BundleHealth,
	verbose bool,
) []T {
	items, err := parseFunc(path)
	if err != nil {
		health.MissingFiles = append(health.MissingFiles, name)
		if verbose {
			fmt.Printf("⚠️  %s: missing\n", name)
		}
		return []T{} // Return empty slice
	}
	health.FoundFiles++
	if verbose {
		fmt.Printf("✓ %s: %d\n", name, len(items))
	}
	return items
}
