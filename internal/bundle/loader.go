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
	pods, err := InventoryPods(extractPath)
	if err != nil {
		// Pods are optional - log warning
		if opts.Verbose {
			fmt.Printf("⚠️  pods: missing\n")
		}
		health.MissingFiles = append(health.MissingFiles, "pods")
		pods = []PodInfo{} // Empty slice
	} else {
		if opts.Verbose {
			fmt.Printf("✓ pods: %d\n", len(pods))
		}
		health.FoundFiles++
	}

	// Inventory log files
	logFiles, err := InventoryLogFiles(extractPath)
	if err != nil {
		// Logs are optional - log warning
		if opts.Verbose {
			fmt.Printf("⚠️  logs: missing\n")
		}
		health.MissingFiles = append(health.MissingFiles, "logs")
		logFiles = []LogFileInfo{} // Empty slice
	} else {
		if opts.Verbose {
			fmt.Printf("✓ logs: %d\n", len(logFiles))
		}
		health.FoundFiles++
	}

	// Parse kubectl resources (all optional)
	crds, crdErr := ParseCRDs(extractPath)
	if crdErr != nil {
		health.MissingFiles = append(health.MissingFiles, "crds")
		if opts.Verbose {
			fmt.Printf("⚠️  crds: missing\n")
		}
	} else {
		health.FoundFiles++
		if opts.Verbose {
			fmt.Printf("✓ crds: %d\n", len(crds))
		}
	}

	deployments, depErr := ParseDeployments(extractPath)
	if depErr != nil {
		health.MissingFiles = append(health.MissingFiles, "deployments")
		if opts.Verbose {
			fmt.Printf("⚠️  deployments: missing\n")
		}
	} else {
		health.FoundFiles++
		if opts.Verbose {
			fmt.Printf("✓ deployments: %d\n", len(deployments))
		}
	}

	services, svcErr := ParseServices(extractPath)
	if svcErr != nil {
		health.MissingFiles = append(health.MissingFiles, "services")
		if opts.Verbose {
			fmt.Printf("⚠️  services: missing\n")
		}
	} else {
		health.FoundFiles++
		if opts.Verbose {
			fmt.Printf("✓ services: %d\n", len(services))
		}
	}

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

	events, evtErr := ParseEvents(extractPath)
	if evtErr != nil {
		health.MissingFiles = append(health.MissingFiles, "events")
		if opts.Verbose {
			fmt.Printf("⚠️  events: missing\n")
		}
	} else {
		health.FoundFiles++
		if opts.Verbose {
			fmt.Printf("✓ events: %d\n", len(events))
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
