package bundle

import (
	"fmt"
	"os"
	"path/filepath"
)

// HealthChecker provides bundle completeness and quality analysis
type HealthChecker struct {
	bundle *Bundle
}

// CriticalFiles lists directories/files that are essential for accurate analysis
var CriticalFiles = []string{
	"rke2/kubectl",      // kubectl output (RKE2)
	"k3s/kubectl",       // kubectl output (K3s)
	"rke2/podlogs",      // pod logs (RKE2)
	"k3s/podlogs",       // pod logs (K3s)
}

// OptionalFiles lists files that enhance analysis but aren't required
var OptionalFiles = []string{
	"rke2/pod-manifests",
	"k3s/pod-manifests",
	"rke2/agent-logs",
	"k3s/agent-logs",
	"rke2/etcd",
	"k3s/etcd",
	"systemlogs",
	"systeminfo",
}

// NewHealthChecker creates a health checker for a bundle
func NewHealthChecker(b *Bundle) *HealthChecker {
	return &HealthChecker{bundle: b}
}

// Check analyzes bundle completeness and returns health metrics
func (hc *HealthChecker) Check() BundleHealth {
	health := BundleHealth{
		MissingFiles: []string{},
		Warnings:     []string{},
	}

	if hc.bundle == nil || hc.bundle.ExtractPath == "" {
		health.Warnings = append(health.Warnings, "No bundle loaded")
		return health
	}

	root := hc.bundle.ExtractPath

	// Check critical files
	for _, critical := range CriticalFiles {
		path := filepath.Join(root, critical)
		exists := hc.pathExists(path)
		health.TotalFiles++
		if exists {
			health.FoundFiles++
		} else {
			health.MissingFiles = append(health.MissingFiles, critical)
		}
	}

	// Check optional files
	for _, optional := range OptionalFiles {
		path := filepath.Join(root, optional)
		exists := hc.pathExists(path)
		health.TotalFiles++
		if exists {
			health.FoundFiles++
		}
		// Optional files don't count as "missing" for health calculation
	}

	// Add warnings for missing critical data
	if !hc.hasKubectlData(root) {
		health.Warnings = append(health.Warnings, "Missing kubectl data: pod/namespace information unavailable")
	}
	if !hc.hasPodLogs(root) {
		health.Warnings = append(health.Warnings, "Missing pod logs: container log analysis limited")
	}
	if !hc.hasSystemLogs(root) {
		health.Warnings = append(health.Warnings, "Missing system logs: node-level diagnostics unavailable")
	}

	// Detect bundle format
	format := hc.detectFormat(root)
	health.BundleType = string(format)

	return health
}

// pathExists checks if a path exists (file or directory)
func (hc *HealthChecker) pathExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// hasKubectlData checks if kubectl directory exists in any format
func (hc *HealthChecker) hasKubectlData(root string) bool {
	paths := []string{
		filepath.Join(root, "rke2", "kubectl"),
		filepath.Join(root, "k3s", "kubectl"),
	}
	for _, p := range paths {
		if hc.pathExists(p) {
			return true
		}
	}
	return false
}

// hasPodLogs checks if pod logs directory exists in any format
func (hc *HealthChecker) hasPodLogs(root string) bool {
	paths := []string{
		filepath.Join(root, "rke2", "podlogs"),
		filepath.Join(root, "k3s", "podlogs"),
	}
	for _, p := range paths {
		if hc.pathExists(p) {
			return true
		}
	}
	return false
}

// hasSystemLogs checks if system logs exist
func (hc *HealthChecker) hasSystemLogs(root string) bool {
	paths := []string{
		filepath.Join(root, "systemlogs"),
		filepath.Join(root, "rke2", "agent-logs"),
		filepath.Join(root, "k3s", "agent-logs"),
	}
	for _, p := range paths {
		if hc.pathExists(p) {
			return true
		}
	}
	return false
}

// detectFormat determines the bundle type from directory structure
func (hc *HealthChecker) detectFormat(root string) BundleFormat {
	if hc.pathExists(filepath.Join(root, "rke2")) {
		return FormatRKE2
	}
	if hc.pathExists(filepath.Join(root, "k3s")) {
		return FormatK3s
	}
	if hc.pathExists(filepath.Join(root, "kubectl")) {
		return FormatKubectl
	}
	return FormatUnknown
}

// IsCriticalMissing returns true if critical data is missing
func (health *BundleHealth) IsCriticalMissing() bool {
	return health.Percentage() < 70
}

// Summary returns a human-readable health summary
func (health *BundleHealth) Summary() string {
	pct := health.Percentage()
	switch {
	case pct >= 90:
		return fmt.Sprintf("Excellent (%d%%)", pct)
	case pct >= 70:
		return fmt.Sprintf("Good (%d%%)", pct)
	case pct >= 50:
		return fmt.Sprintf("Fair (%d%%) — Some data may be incomplete", pct)
	default:
		return fmt.Sprintf("Poor (%d%%) — Critical data missing", pct)
	}
}
