package bundle

import (
	"fmt"
	"log"
	"os"
)

// Load loads a bundle from a tar.gz file and returns a Bundle structure.
func Load(opts ImportOptions) (*Bundle, error) {
	// Validate options
	if opts.Path == "" {
		return nil, fmt.Errorf("bundle path is required")
	}

	// Check if file exists
	if _, err := os.Stat(opts.Path); os.IsNotExist(err) {
		return nil, fmt.Errorf("bundle file not found: %s", opts.Path)
	}

	// Set default max size if not specified
	if opts.MaxSize == 0 {
		opts.MaxSize = DefaultMaxBundleSize
	}

	// Extract the bundle
	extractPath, err := Extract(opts.Path, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to extract bundle: %w", err)
	}

	// Parse manifest
	manifest, err := ParseManifest(extractPath)
	if err != nil {
		Cleanup(extractPath)
		return nil, fmt.Errorf("failed to parse bundle manifest: %w", err)
	}

	// Inventory pods
	pods, err := InventoryPods(extractPath)
	if err != nil {
		Cleanup(extractPath)
		return nil, fmt.Errorf("failed to inventory pods: %w", err)
	}

	// Inventory log files
	logFiles, err := InventoryLogFiles(extractPath)
	if err != nil {
		Cleanup(extractPath)
		return nil, fmt.Errorf("failed to inventory log files: %w", err)
	}

	// Parse kubectl resources (ignore errors - these are optional)
	// Storing as interface{} to avoid import cycle - will be type-asserted in datasource
	crds, err := ParseCRDs(extractPath)
	if err != nil {
		log.Printf("Warning: Failed to parse CRDs from bundle: %v", err)
	}
	deployments, err := ParseDeployments(extractPath)
	if err != nil {
		log.Printf("Warning: Failed to parse Deployments from bundle: %v", err)
	}
	services, err := ParseServices(extractPath)
	if err != nil {
		log.Printf("Warning: Failed to parse Services from bundle: %v", err)
	}
	namespaces, err := ParseNamespaces(extractPath)
	if err != nil {
		log.Printf("Warning: Failed to parse Namespaces from bundle: %v", err)
	}

	// Get bundle file size
	stat, _ := os.Stat(opts.Path)
	bundleSize := int64(0)
	if stat != nil {
		bundleSize = stat.Size()
	}

	// Convert to []interface{} to avoid import cycle
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

	// Create bundle structure
	bundle := &Bundle{
		Path:        opts.Path,
		ExtractPath: extractPath,
		Manifest:    manifest,
		Pods:        pods,
		LogFiles:    logFiles,
		CRDs:        crdsI,
		Deployments: deploymentsI,
		Services:    servicesI,
		Namespaces:  namespacesI,
		Loaded:      true,
		Size:        bundleSize,
	}

	return bundle, nil
}

// Close cleans up the bundle's extracted files.
func (b *Bundle) Close() error {
	if b.ExtractPath == "" {
		return nil
	}
	return Cleanup(b.ExtractPath)
}

// GetPod returns pod information by namespace and name.
func (b *Bundle) GetPod(namespace, name string) *PodInfo {
	for i := range b.Pods {
		if b.Pods[i].Namespace == namespace && b.Pods[i].Name == name {
			return &b.Pods[i]
		}
	}
	return nil
}

// GetLogFile returns log file information by path.
func (b *Bundle) GetLogFile(path string) *LogFileInfo {
	for i := range b.LogFiles {
		if b.LogFiles[i].Path == path {
			return &b.LogFiles[i]
		}
	}
	return nil
}

// ReadLogFile reads the contents of a log file from the bundle.
func (b *Bundle) ReadLogFile(logFile *LogFileInfo) ([]byte, error) {
	if logFile == nil {
		return nil, fmt.Errorf("log file info is nil")
	}
	return os.ReadFile(logFile.Path)
}

// Summary returns a human-readable summary of the bundle.
func (b *Bundle) Summary() string {
	return fmt.Sprintf(
		"Bundle: %s\nNode: %s\nRKE2: %s\nK8s: %s\nFiles: %d\nPods: %d\nLogs: %d",
		b.Manifest.NodeName,
		b.Manifest.NodeName,
		b.Manifest.RKE2Version,
		b.Manifest.K8sVersion,
		b.Manifest.FileCount,
		len(b.Pods),
		len(b.LogFiles),
	)
}
