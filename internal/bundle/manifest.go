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

	// Also parse pod manifests to get container names
	// This handles cases where log filenames don't include container names
	manifestsDir := filepath.Join(bundleRoot, "rke2", "pod-manifests")
	if _, err := os.Stat(manifestsDir); err == nil {
		parsePodManifestsForContainers(manifestsDir, podMap)
	}

	// Also parse poddescribe output (PR #418) for container names
	poddescribeDir := filepath.Join(bundleRoot, "rke2", "kubectl", "poddescribe")
	if _, err := os.Stat(poddescribeDir); err == nil {
		parsePodDescribeForContainers(poddescribeDir, podMap)
	}


	// Convert map to slice
	for _, pod := range podMap {
		pods = append(pods, *pod)
	}

	return pods, nil
}

// parsePodLogFilename extracts pod information from a log filename.
func parsePodLogFilename(filename string) *LogFileInfo {
	// Check for -previous suffix first
	isPrevious := strings.HasSuffix(filename, "-previous")
	if isPrevious {
		filename = strings.TrimSuffix(filename, "-previous")
	}

	// The format is: namespace-podname (no separate container field)
	// Example: calico-system-calico-kube-controllers-8889b866f-jtlsb
	// namespace: calico-system
	// podname: calico-kube-controllers-8889b866f-jtlsb

	// Find the first dash to separate namespace from pod
	// Namespaces typically follow pattern: xxx-system, xxx-xxx, etc
	// We need to identify where namespace ends and pod begins
	// Common patterns: kube-system, calico-system, cattle-system, longhorn-system

	parts := strings.Split(filename, "-")
	if len(parts) < 2 {
		return nil // Invalid format
	}

	// Try to identify namespace boundary
	// Common namespace patterns end with: -system, -operator
	var namespace, podName string

	// Check for common namespace patterns
	if len(parts) >= 2 && (parts[1] == "system" || parts[1] == "operator") {
		// Pattern: xxx-system or xxx-operator
		namespace = parts[0] + "-" + parts[1]
		if len(parts) > 2 {
			podName = strings.Join(parts[2:], "-")
		}
	} else if len(parts) >= 3 && parts[2] == "system" {
		// Pattern: xxx-xxx-system
		namespace = parts[0] + "-" + parts[1] + "-" + parts[2]
		if len(parts) > 3 {
			podName = strings.Join(parts[3:], "-")
		}
	} else {
		// Fallback: assume first part is namespace
		namespace = parts[0]
		podName = strings.Join(parts[1:], "-")
	}

	if podName == "" {
		return nil // No pod name found
	}

	return &LogFileInfo{
		Path:          filename,
		Type:          LogTypePod,
		Namespace:     namespace,
		PodName:       podName,
		ContainerName: "", // Not available in this format
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

// parsePodManifestsForContainers scans pod manifest YAMLs to extract container names
// This supplements container info that may be missing from log filenames
func parsePodManifestsForContainers(manifestsDir string, podMap map[string]*PodInfo) {
	files, err := os.ReadDir(manifestsDir)
	if err != nil {
		return
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".yaml") && !strings.HasSuffix(file.Name(), ".yml") {
			continue
		}

		content, err := os.ReadFile(filepath.Join(manifestsDir, file.Name()))
		if err != nil {
			continue
		}

		// Parse YAML to extract pod name and containers
		containers, podNamespace, podName := parsePodYAMLForContainers(string(content))
		if podName == "" || len(containers) == 0 {
			continue
		}

		// Find matching pod in map
		key := podNamespace + "/" + podName
		if pod, exists := podMap[key]; exists {
			// Add containers that aren't already in the list
			for _, container := range containers {
				if !contains(pod.Containers, container) {
					pod.Containers = append(pod.Containers, container)
				}
			}
		}
	}
}

// parsePodYAMLForContainers extracts container names from pod YAML
func parsePodYAMLForContainers(yamlContent string) ([]string, string, string) {
	var containers []string
	var podName, namespace string

	lines := strings.Split(yamlContent, "\n")
	var inMetadata, inSpec, inContainers bool
	var indentLevel int

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Track section entry
		if trimmed == "metadata:" {
			inMetadata = true
			inSpec = false
			continue
		}
		if trimmed == "spec:" {
			inMetadata = false
			inSpec = true
			continue
		}
		if inSpec && trimmed == "containers:" {
			inContainers = true
			indentLevel = len(line) - len(trimmed)
			continue
		}

		// Extract pod name from metadata
		if inMetadata && strings.HasPrefix(trimmed, "name:") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				podName = strings.TrimSpace(parts[1])
			}
		}

		// Extract namespace from metadata
		if inMetadata && strings.HasPrefix(trimmed, "namespace:") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				namespace = strings.TrimSpace(parts[1])
			}
		}

		// Extract container names from containers section
		if inContainers {
			currentIndent := len(line) - len(trimmed)
			// Check if we've exited the containers section
			if currentIndent <= indentLevel && trimmed != "" {
				inContainers = false
				continue
			}

			// Container entry starts with "- name:"
			if strings.HasPrefix(trimmed, "- name:") {
				parts := strings.SplitN(trimmed, ":", 2)
				if len(parts) == 2 {
					containerName := strings.TrimSpace(parts[1])
					containers = append(containers, containerName)
				}
			}
		}

		// Simple heuristic: stop at end of file or next top-level section
		if i > 0 && (strings.HasPrefix(trimmed, "status:") || strings.HasPrefix(trimmed, "---")) {
			break
		}
	}

	return containers, namespace, podName
}

// parsePodDescribeForContainers parses poddescribe output to extract container names
// Uses PR #418 format: rke2/kubectl/poddescribe/<namespace>
func parsePodDescribeForContainers(poddescribeDir string, podMap map[string]*PodInfo) {
	files, err := os.ReadDir(poddescribeDir)
	if err != nil {
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		content, err := os.ReadFile(filepath.Join(poddescribeDir, file.Name()))
		if err != nil {
			continue
		}

		// Parse each pod in the describe output
		pods := strings.Split(string(content), "\n\n")
		for _, podBlock := range pods {
			lines := strings.Split(podBlock, "\n")
			var podName, namespace string
			var containers []string
			var inContainers bool

			for _, line := range lines {
				line = strings.TrimSpace(line)

				if strings.HasPrefix(line, "Name:") {
					podName = strings.TrimSpace(strings.TrimPrefix(line, "Name:"))
				}

				if strings.HasPrefix(line, "Namespace:") {
					namespace = strings.TrimSpace(strings.TrimPrefix(line, "Namespace:"))
				}

				if line == "Containers:" {
					inContainers = true
					continue
				}

				// Container names are lines ending with ":" in the Containers section
				if inContainers && strings.HasSuffix(line, ":") && !strings.Contains(line, "/") {
					containerName := strings.TrimSuffix(line, ":")
					if containerName != "" && containerName != "Containers" {
						containers = append(containers, containerName)
					}
				}

				// Exit containers section on empty line or new section
				if inContainers && line == "" {
					inContainers = false
				}
			}

			// Update pod in map
			if podName != "" && namespace != "" {
				key := namespace + "/" + podName
				if pod, exists := podMap[key]; exists {
					for _, container := range containers {
						if !contains(pod.Containers, container) {
							pod.Containers = append(pod.Containers, container)
						}
					}
				}
			}
		}
	}
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
