package bundle

import (
	"fmt"
	"os"
	"path/filepath"
)

// BundleFile represents an expected file in the bundle
type BundleFile struct {
	Path        string
	Required    bool // If true, bundle is incomplete without this
	Description string
	Weight      int // Contribution to completeness percentage
}

// CompletenessResult contains the analysis of bundle completeness
type CompletenessResult struct {
	TotalFiles      int
	PresentFiles    int
	RequiredFiles   int
	RequiredPresent int
	Percentage      int // 0-100
	MissingRequired []string
	MissingOptional []string
	PresentFileList []string // List of what was found
}

// IsComplete returns true if all required files are present
func (c *CompletenessResult) IsComplete() bool {
	return len(c.MissingRequired) == 0
}

// GetStatus returns a human-readable status string
// Thresholds align with GetStatusColor() for consistent UX
func (c *CompletenessResult) GetStatus() string {
	if c.Percentage == 100 {
		return "Complete"
	} else if c.Percentage >= 70 {
		return "Good"
	} else if c.Percentage >= 40 {
		return "Partial"
	}
	return "Minimal"
}

// GetStatusColor returns the appropriate color for the status
func (c *CompletenessResult) GetStatusColor() string {
	switch {
	case c.Percentage >= 90:
		return "green"
	case c.Percentage >= 70:
		return "yellow"
	case c.Percentage >= 40:
		return "orange"
	default:
		return "red"
	}
}

// Define expected bundle files for RKE2
// Based on rancher2_logs_collector.sh output structure
func getRKE2ExpectedFiles() []BundleFile {
	return []BundleFile{
		// Required files (core diagnostic data)
		{Path: "rke2/kubectl/pods", Required: true, Description: "Pod list", Weight: 15},
		{Path: "rke2/kubectl/nodes", Required: true, Description: "Node list", Weight: 10},
		{Path: "rke2/kubectl/events", Required: true, Description: "Cluster events", Weight: 10},

		// Important but not required
		{Path: "rke2/kubectl/deployments", Required: false, Description: "Deployments", Weight: 5},
		{Path: "rke2/kubectl/daemonsets", Required: false, Description: "DaemonSets", Weight: 5},
		{Path: "rke2/kubectl/services", Required: false, Description: "Services", Weight: 5},
		{Path: "rke2/kubectl/pv", Required: false, Description: "PersistentVolumes", Weight: 3},
		{Path: "rke2/kubectl/pvc", Required: false, Description: "PVCs", Weight: 3},
		{Path: "rke2/kubectl/statefulsets", Required: false, Description: "StatefulSets", Weight: 3},
		{Path: "rke2/kubectl/configmaps", Required: false, Description: "ConfigMaps", Weight: 2},
		{Path: "rke2/kubectl/helmcharts", Required: false, Description: "HelmCharts", Weight: 2},

		// Pod logs (any presence counts)
		{Path: "rke2/podlogs", Required: false, Description: "Pod logs", Weight: 10},

		// System information
		{Path: "systeminfo/dmesg", Required: false, Description: "Kernel logs (dmesg)", Weight: 5},
		{Path: "systeminfo/meminfo", Required: false, Description: "Memory info", Weight: 3},
		{Path: "systeminfo/cpuinfo", Required: false, Description: "CPU info", Weight: 2},
		{Path: "systeminfo/dfh", Required: false, Description: "Disk usage", Weight: 3},

		// etcd data
		{Path: "rke2/etcd/endpoint_status", Required: false, Description: "etcd status", Weight: 5},

		// Journald logs
		{Path: "systemlogs", Required: false, Description: "System logs", Weight: 5},
	}
}

// getK3sExpectedFiles defines expected files for K3s bundles
func getK3sExpectedFiles() []BundleFile {
	return []BundleFile{
		// Required files (core diagnostic data)
		{Path: "k3s/kubectl/pods", Required: true, Description: "Pod list", Weight: 15},
		{Path: "k3s/kubectl/nodes", Required: true, Description: "Node list", Weight: 10},
		{Path: "k3s/kubectl/events", Required: true, Description: "Cluster events", Weight: 10},

		// Important but not required
		{Path: "k3s/kubectl/deployments", Required: false, Description: "Deployments", Weight: 5},
		{Path: "k3s/kubectl/daemonsets", Required: false, Description: "DaemonSets", Weight: 5},
		{Path: "k3s/kubectl/services", Required: false, Description: "Services", Weight: 5},
		{Path: "k3s/kubectl/pv", Required: false, Description: "PersistentVolumes", Weight: 3},
		{Path: "k3s/kubectl/pvc", Required: false, Description: "PVCs", Weight: 3},
		{Path: "k3s/kubectl/statefulsets", Required: false, Description: "StatefulSets", Weight: 3},
		{Path: "k3s/kubectl/configmaps", Required: false, Description: "ConfigMaps", Weight: 2},
		{Path: "k3s/kubectl/helmcharts", Required: false, Description: "HelmCharts", Weight: 2},

		// Pod logs (any presence counts)
		{Path: "k3s/podlogs", Required: false, Description: "Pod logs", Weight: 10},

		// System information (same as RKE2)
		{Path: "systeminfo/dmesg", Required: false, Description: "Kernel logs (dmesg)", Weight: 5},
		{Path: "systeminfo/meminfo", Required: false, Description: "Memory info", Weight: 3},
		{Path: "systeminfo/cpuinfo", Required: false, Description: "CPU info", Weight: 2},
		{Path: "systeminfo/dfh", Required: false, Description: "Disk usage", Weight: 3},

		// etcd data
		{Path: "k3s/etcd/endpoint_status", Required: false, Description: "etcd status", Weight: 5},

		// Journald logs
		{Path: "systemlogs", Required: false, Description: "System logs", Weight: 5},
	}
}

// getExpectedFiles returns format-appropriate expected files
func getExpectedFiles(format BundleFormat) []BundleFile {
	switch format {
	case FormatK3s:
		return getK3sExpectedFiles()
	case FormatRKE2:
		return getRKE2ExpectedFiles()
	default:
		// Default to RKE2 for backward compatibility
		return getRKE2ExpectedFiles()
	}
}

// AnalyzeCompleteness checks bundle completeness
func AnalyzeCompleteness(extractPath string) (*CompletenessResult, error) {
	bundleRoot := getBundleRoot(extractPath)
	format := DetectFormat(extractPath)
	expectedFiles := getExpectedFiles(format)

	result := &CompletenessResult{
		TotalFiles:      len(expectedFiles),
		PresentFiles:    0,
		RequiredFiles:   0,
		RequiredPresent: 0,
		MissingRequired: make([]string, 0),
		MissingOptional: make([]string, 0),
		PresentFileList: make([]string, 0),
	}

	totalWeight := 0
	presentWeight := 0

	for _, file := range expectedFiles {
		fullPath := filepath.Join(bundleRoot, file.Path)
		
		// Check if file or directory exists
		info, err := os.Stat(fullPath)
		exists := err == nil

		if file.Required {
			result.RequiredFiles++
		}

		if exists {
			result.PresentFiles++
			presentWeight += file.Weight
			result.PresentFileList = append(result.PresentFileList, file.Description)
			
			if file.Required {
				result.RequiredPresent++
			}

			// For directories (like podlogs), check if they have content
			if info.IsDir() {
				entries, err := os.ReadDir(fullPath)
				if err != nil || len(entries) == 0 {
					// Directory exists but is empty - count as partial (50% weight)
					// Integer truncation is intentional: odd weights round down
					presentWeight -= file.Weight / 2
				}
			}
		} else {
			if file.Required {
				result.MissingRequired = append(result.MissingRequired, file.Description)
			} else {
				result.MissingOptional = append(result.MissingOptional, file.Description)
			}
		}

		totalWeight += file.Weight
	}

	// Calculate percentage
	if totalWeight > 0 {
		result.Percentage = (presentWeight * 100) / totalWeight
	}

	return result, nil
}

// QuickCompletenessCheck returns just the percentage for fast display
func QuickCompletenessCheck(extractPath string) int {
	result, err := AnalyzeCompleteness(extractPath)
	if err != nil {
		return 0
	}
	return result.Percentage
}

// FormatCompleteness returns a formatted string for display
func FormatCompleteness(result *CompletenessResult) string {
	status := result.GetStatus()
	return fmt.Sprintf("Bundle: %d%% (%s)", result.Percentage, status)
}
