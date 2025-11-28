package bundle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LoadFromPath is the new bulletproof entry point that handles both
// extracted directories and compressed archives (.tar.gz, .zip).
//
// It auto-detects the input type and validates bundle structure.
func LoadFromPath(path string, opts ImportOptions) (*Bundle, error) {
	// Step 1: Validate and resolve path
	absPath, pathInfo, err := validateAndResolvePath(path, opts.Verbose)
	if err != nil {
		return nil, err
	}

	// Step 2: Determine if path is directory or archive
	var extractPath string
	var bundleSize int64
	var isTemporary bool

	if pathInfo.IsDir() {
		// DIRECTORY MODE: Use extracted folder directly
		if opts.Verbose {
			fmt.Printf("📁 Detected extracted bundle directory: %s\n", absPath)
		}

		// Validate it's actually a bundle
		if err := validateBundleStructure(absPath, opts.Verbose); err != nil {
			return nil, fmt.Errorf("invalid bundle directory: %w", err)
		}

		extractPath = absPath
		bundleSize = 0 // Unknown for directories
		isTemporary = false

	} else {
		// ARCHIVE MODE: Extract compressed bundle
		if opts.Verbose {
			fmt.Printf("📦 Detected bundle archive: %s (%.2f MB)\n",
				filepath.Base(absPath),
				float64(pathInfo.Size())/(1024*1024))
		}

		// Validate archive type
		if err := validateArchiveType(absPath, opts.Verbose); err != nil {
			return nil, err
		}

		// Extract the archive
		extractPath, err = extractArchive(absPath, opts)
		if err != nil {
			return nil, err
		}

		bundleSize = pathInfo.Size()
		isTemporary = true
	}

	// Step 3: Load bundle from extracted path (common for both modes)
	bundle, err := loadFromExtractedPath(extractPath, absPath, bundleSize, opts)
	if err != nil {
		// Cleanup temp extraction if archive mode
		if isTemporary {
			Cleanup(extractPath)
		}
		return nil, err
	}

	// Mark whether cleanup is needed
	bundle.IsTemporary = isTemporary

	return bundle, nil
}

// validateAndResolvePath validates the path exists and resolves it to absolute path
func validateAndResolvePath(path string, verbose bool) (string, os.FileInfo, error) {
	if path == "" {
		if verbose {
			return "", nil, fmt.Errorf("bundle path is required\n\n" +
				"USAGE:\n" +
				"  r8s --bundle=/path/to/bundle.tar.gz    # Archive file\n" +
				"  r8s --bundle=/path/to/extracted/       # Extracted directory\n\n" +
				"HINT: Provide either a .tar.gz archive or an extracted bundle folder")
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
				"  2. Ensure file/folder exists\n"+
				"  3. Check file permissions\n"+
				"  4. Try using an absolute path", path, cwd, absPath)
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
				"    └── metadata.json\n\n"+
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

// validateArchiveType checks if the file is a supported archive format
func validateArchiveType(path string, verbose bool) error {
	ext := strings.ToLower(filepath.Ext(path))

	// Check for supported extensions
	supported := []string{".gz", ".tgz"}
	isSupported := false

	for _, s := range supported {
		if strings.HasSuffix(strings.ToLower(path), s) {
			isSupported = true
			break
		}
	}

	if !isSupported {
		if verbose {
			return fmt.Errorf("unsupported archive format: %s\n\n"+
				"Supported formats:\n"+
				"  • .tar.gz  (RKE2 support bundles)\n"+
				"  • .tgz     (compressed tar)\n\n"+
				"Current file: %s\n\n"+
				"SOLUTIONS:\n"+
				"  1. If bundle is already extracted, point to the folder:\n"+
				"     r8s --bundle=/path/to/extracted-folder/\n"+
				"  2. If you have a different archive format, extract it first\n"+
				"  3. Ensure the file extension is preserved", ext, filepath.Base(path))
		}
		return fmt.Errorf("unsupported format %s (expected .tar.gz or .tgz)", ext)
	}

	return nil
}

// extractArchive handles archive extraction (wraps existing Extract function)
func extractArchive(path string, opts ImportOptions) (string, error) {
	if opts.Verbose {
		fmt.Println("Extracting archive...")
	}

	// Use existing Extract function
	extractPath, err := Extract(path, opts)
	if err != nil {
		if opts.Verbose {
			return "", fmt.Errorf("extraction failed: %w\n\n"+
				"TROUBLESHOOTING:\n"+
				"  • Ensure the file is a valid .tar.gz archive\n"+
				"  • Check file isn't corrupted (try: tar -tzf bundle.tar.gz)\n"+
				"  • Verify sufficient disk space\n"+
				"  • Check file permissions\n\n"+
				"ALTERNATIVE:\n"+
				"  Extract manually and use folder mode:\n"+
				"  $ tar -xzf bundle.tar.gz\n"+
				"  $ r8s --bundle=./extracted-folder/", err)
		}
		return "", err
	}

	if opts.Verbose {
		fmt.Printf("✓ Extracted to: %s\n", extractPath)
	}

	return extractPath, nil
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
		IsTemporary: false, // Will be set by caller
	}

	return bundle, nil
}
