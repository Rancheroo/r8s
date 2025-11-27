package bundle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ParseManifest analyzes a bundle directory and extracts metadata.
func ParseManifest(extractPath string) (*BundleManifest, error) {
	manifest := &BundleManifest{
		CollectedAt: time.Now(), // Default, will try to parse from filename
		BundleType:  string(DetectFormat(extractPath)),
	}

	// Detect bundle format
	format := DetectFormat(extractPath)
	if format == FormatUnknown {
		return nil, fmt.Errorf("unknown bundle format")
	}

	// Extract node name from directory structure or filename
	manifest.NodeName = extractNodeName(extractPath)

	// Count files and calculate total size
	fileCount, totalSize, err := calculateBundleStats(extractPath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate bundle stats: %w", err)
	}
	manifest.FileCount = fileCount
	manifest.TotalSize = totalSize

	// Parse RKE2 version if available
	if format == FormatRKE2 {
		manifest.RKE2Version = parseRKE2Version(extractPath)
		manifest.K8sVersion = parseK8sVersion(extractPath)
	}

	return manifest, nil
}

// DetectFormat determines the bundle format by examining directory structure.
func DetectFormat(extractPath string) BundleFormat {
	// Check for RKE2 support bundle structure (direct)
	rke2Dir := filepath.Join(extractPath, "rke2")
	if stat, err := os.Stat(rke2Dir); err == nil && stat.IsDir() {
		return FormatRKE2
	}

	// Check for RKE2 with wrapper directory (common in tar.gz bundles)
	entries, err := os.ReadDir(extractPath)
	if err == nil && len(entries) == 1 && entries[0].IsDir() {
		// Single top-level directory - check inside it
		wrapperDir := filepath.Join(extractPath, entries[0].Name())
		rke2Dir = filepath.Join(wrapperDir, "rke2")
		if stat, err := os.Stat(rke2Dir); err == nil && stat.IsDir() {
			return FormatRKE2
		}
	}

	// Check for kubectl cluster-info dump structure
	namespacesDir := filepath.Join(extractPath, "namespaces")
	if stat, err := os.Stat(namespacesDir); err == nil && stat.IsDir() {
		return FormatKubectl
	}

	return FormatUnknown
}

// getBundleRoot returns the actual bundle root, handling wrapper directories.
func getBundleRoot(extractPath string) string {
	// Check if there's a single wrapper directory
	entries, err := os.ReadDir(extractPath)
	if err == nil && len(entries) == 1 && entries[0].IsDir() {
		// Check if this wrapper contains the bundle
		wrapperDir := filepath.Join(extractPath, entries[0].Name())
		rke2Dir := filepath.Join(wrapperDir, "rke2")
		if _, err := os.Stat(rke2Dir); err == nil {
			return wrapperDir
		}
	}
	return extractPath
}

// extractNodeName attempts to extract the node name from the bundle.
func extractNodeName(extractPath string) string {
	bundleRoot := getBundleRoot(extractPath)

	// Try to get from directory name (e.g., w-guard-wg-cp-svtk6-lqtxw)
	baseName := filepath.Base(bundleRoot)

	// RKE2 bundles often have pattern: <nodename>-<timestamp>
	// Example: w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09
	parts := strings.Split(baseName, "-")
	if len(parts) >= 6 {
		// Assume last 3 parts are timestamp, rest is node name
		nodeParts := parts[:len(parts)-3]
		return strings.Join(nodeParts, "-")
	}

	// Try reading from systeminfo/hostname file
	hostnameFile := filepath.Join(bundleRoot, "systeminfo", "hostname")
	if data, err := os.ReadFile(hostnameFile); err == nil {
		hostname := strings.TrimSpace(string(data))
		if hostname != "" {
			return hostname
		}
	}

	// Fallback to directory name
	return baseName
}

// parseRKE2Version attempts to read the RKE2 version from the bundle.
func parseRKE2Version(extractPath string) string {
	bundleRoot := getBundleRoot(extractPath)
	versionFile := filepath.Join(bundleRoot, "rke2", "version")
	if data, err := os.ReadFile(versionFile); err == nil {
		return strings.TrimSpace(string(data))
	}
	return "unknown"
}

// parseK8sVersion attempts to read the Kubernetes version from the bundle.
func parseK8sVersion(extractPath string) string {
	bundleRoot := getBundleRoot(extractPath)
	// Try kubectl version file
	versionFile := filepath.Join(bundleRoot, "rke2", "kubectl", "version")
	if data, err := os.ReadFile(versionFile); err == nil {
		// Parse version output (could be JSON or text)
		version := strings.TrimSpace(string(data))
		// Extract version number if present
		if strings.Contains(version, "GitVersion") {
			// JSON format: extract version
			lines := strings.Split(version, "\n")
			for _, line := range lines {
				if strings.Contains(line, "GitVersion") {
					parts := strings.Split(line, ":")
					if len(parts) >= 2 {
						ver := strings.Trim(parts[1], `", `)
						return ver
					}
				}
			}
		}
		return version
	}
	return "unknown"
}

// calculateBundleStats walks the directory tree and counts files/sizes.
func calculateBundleStats(extractPath string) (fileCount int, totalSize int64, err error) {
	err = filepath.Walk(extractPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fileCount++
			totalSize += info.Size()
		}
		return nil
	})
	return
}

// InventoryPods scans the bundle for pod information.
func InventoryPods(extractPath string) ([]PodInfo, error) {
	var pods []PodInfo
	bundleRoot := getBundleRoot(extractPath)

	// Look for pod logs in rke2/podlogs/
	podlogsDir := filepath.Join(bundleRoot, "rke2", "podlogs")
	if _, err := os.Stat(podlogsDir); os.IsNotExist(err) {
		return pods, nil // No pod logs directory
	}

	// Map to track pods we've seen
	podMap := make(map[string]*PodInfo)

	// Walk the podlogs directory
	err := filepath.Walk(podlogsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		// Parse filename to extract pod info
		// Format: <namespace>_<podname>_<container>.log
		// or: <namespace>_<podname>_<container>-previous.log
		relPath, _ := filepath.Rel(podlogsDir, path)
		podInfo := parsePodLogFilename(relPath)
		if podInfo == nil {
			return nil
		}

		// Create key for pod
		key := podInfo.Namespace + "/" + podInfo.PodName

		// Get or create pod entry
		pod, exists := podMap[key]
		if !exists {
			pod = &PodInfo{
				Namespace:  podInfo.Namespace,
				Name:       podInfo.PodName,
				Containers: []string{},
			}
			podMap[key] = pod
		}

		// Add container if not already present
		if !contains(pod.Containers, podInfo.ContainerName) {
			pod.Containers = append(pod.Containers, podInfo.ContainerName)
		}

		// Track log availability
		if podInfo.IsPrevious {
			pod.HasPreviousLogs = true
		} else {
			pod.HasCurrentLogs = true
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Convert map to slice
	for _, pod := range podMap {
		pods = append(pods, *pod)
	}

	return pods, nil
}

// parsePodLogFilename extracts pod information from a log filename.
func parsePodLogFilename(filename string) *LogFileInfo {
	// Remove .log extension
	name := strings.TrimSuffix(filename, ".log")

	// Check for -previous suffix
	isPrevious := strings.HasSuffix(name, "-previous")
	if isPrevious {
		name = strings.TrimSuffix(name, "-previous")
	}

	// Split by underscore: namespace_podname_container
	parts := strings.Split(name, "_")
	if len(parts) < 3 {
		return nil // Invalid format
	}

	return &LogFileInfo{
		Path:          filename,
		Type:          LogTypePod,
		Namespace:     parts[0],
		PodName:       strings.Join(parts[1:len(parts)-1], "_"),
		ContainerName: parts[len(parts)-1],
		IsPrevious:    isPrevious,
	}
}

// contains checks if a string slice contains a value.
func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// InventoryLogFiles scans the bundle for all log files.
func InventoryLogFiles(extractPath string) ([]LogFileInfo, error) {
	var logFiles []LogFileInfo
	bundleRoot := getBundleRoot(extractPath)

	// Scan pod logs
	podlogsDir := filepath.Join(bundleRoot, "rke2", "podlogs")
	if stat, err := os.Stat(podlogsDir); err == nil && stat.IsDir() {
		err := filepath.Walk(podlogsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}

			relPath, _ := filepath.Rel(podlogsDir, path)
			logInfo := parsePodLogFilename(relPath)
			if logInfo != nil {
				logInfo.Path = path
				logInfo.Size = info.Size()
				logFiles = append(logFiles, *logInfo)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	// Scan system logs
	systemlogsDir := filepath.Join(bundleRoot, "systemlogs")
	if stat, err := os.Stat(systemlogsDir); err == nil && stat.IsDir() {
		err := filepath.Walk(systemlogsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}

			logInfo := LogFileInfo{
				Path: path,
				Type: LogTypeSystem,
				Size: info.Size(),
			}
			logFiles = append(logFiles, logInfo)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return logFiles, nil
}
