package bundle

import (
	"fmt"
	"os"
	"path/filepath"
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
	if err := validateBundleStructure(absPath, opts.Verbose); err != nil {
		return nil, fmt.Errorf("invalid bundle directory: %w", err)
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

// validateBundleStructure verifies a directory contains valid bundle structure
func validateBundleStructure(dir string, verbose bool) error {
	// Check for RKE2 bundle markers
	rke2Dir := filepath.Join(dir, "rke2")
	if _, err := os.Stat(rke2Dir); os.IsNotExist(err) {
		if verbose {
			return fmt.Errorf("not a valid RKE2 bundle\n\n"+
				"Missing: rke2/ directory\n"+
				"Path checked: %s\n\n"+
				"EXPECTED STRUCTURE:\n"+
				"  bundle-folder/\n"+
				"    ├── rke2/\n"+
				"    │   ├── kubectl/\n"+
				"    │   ├── podlogs/\n"+
				"    │   └── ...\n"+
				"    └── (other bundle files)\n\n"+
				"HINT: This folder doesn't appear to be an extracted RKE2 support bundle", rke2Dir)
		}
		return fmt.Errorf("missing rke2/ directory - not a valid bundle")
	}

	// Check for kubectl data or podlogs
	kubectlDir := filepath.Join(dir, "rke2", "kubectl")
	podlogsDir := filepath.Join(dir, "rke2", "podlogs")

	hasKubectl := false
	hasPodlogs := false

	if info, err := os.Stat(kubectlDir); err == nil && info.IsDir() {
		hasKubectl = true
	}
	if info, err := os.Stat(podlogsDir); err == nil && info.IsDir() {
		hasPodlogs = true
	}

	if !hasKubectl && !hasPodlogs {
		if verbose {
			return fmt.Errorf("bundle appears incomplete\n\n" +
				"Missing both:\n" +
				"  - rke2/kubectl/ (for resource data)\n" +
				"  - rke2/podlogs/ (for pod logs)\n\n" +
				"HINT: This may be a partial or corrupted bundle extraction")
		}
		return fmt.Errorf("missing kubectl/ and podlogs/ - bundle appears incomplete")
	}

	return nil
}

// loadFromExtractedPath loads bundle data from an extracted directory
func loadFromExtractedPath(extractPath, originalPath string, size int64, opts ImportOptions) (*Bundle, error) {
	if opts.Verbose {
		fmt.Println("Parsing bundle data...")
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
	pods, err := InventoryPods(extractPath)
	if err != nil {
		// Pods are optional - log warning
		if opts.Verbose {
			fmt.Printf("⚠ Warning: No pods found (%v)\n", err)
		}
		pods = []PodInfo{} // Empty slice
	}

	// Inventory log files
	logFiles, err := InventoryLogFiles(extractPath)
	if err != nil {
		// Logs are optional - log warning
		if opts.Verbose {
			fmt.Printf("⚠ Warning: No log files found (%v)\n", err)
		}
		logFiles = []LogFileInfo{} // Empty slice
	}

	// Parse kubectl resources (all optional)
	crds, _ := ParseCRDs(extractPath)
	deployments, _ := ParseDeployments(extractPath)
	services, _ := ParseServices(extractPath)
	namespaces, _ := ParseNamespaces(extractPath)

	// Convert to interfaces for storage
	var crdsI, deploymentsI, servicesI, namespacesI []interface{}
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

	if opts.Verbose {
		fmt.Printf("✓ Loaded: %d pods, %d logs, %d deployments, %d services, %d CRDs, %d namespaces\n",
			len(pods), len(logFiles), len(deployments), len(services), len(crds), len(namespaces))
	}

	// Create bundle
	bundle := &Bundle{
		Path:        originalPath,
		ExtractPath: extractPath,
		Manifest:    manifest,
		Pods:        pods,
		LogFiles:    logFiles,
		CRDs:        crdsI,
		Deployments: deploymentsI,
		Services:    servicesI,
		Namespaces:  namespacesI,
		Loaded:      true,
		Size:        size,
		IsTemporary: false, // Bundles are already extracted, never temporary
	}

	return bundle, nil
}
