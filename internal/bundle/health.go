// Package bundle provides health checking for support bundles.
// Sprint 8: Bundle Health v2 - partial bundle support with impact scoring.
package bundle

import (
	"fmt"
	"os"
	"path/filepath"
)

// FileImportance indicates how critical a missing file is
type FileImportance int

const (
	ImportanceCritical FileImportance = iota // Bundle unusable without this
	ImportanceHigh                           // Major feature degraded
	ImportanceMedium                         // Minor feature affected
	ImportanceLow                            // Cosmetic only
)

// String returns the string representation of FileImportance
func (fi FileImportance) String() string {
	switch fi {
	case ImportanceCritical:
		return "Critical"
	case ImportanceHigh:
		return "High"
	case ImportanceMedium:
		return "Medium"
	case ImportanceLow:
		return "Low"
	default:
		return "Unknown"
	}
}

// ExpectedFile represents a file that should be present in a bundle
type ExpectedFile struct {
	Path       string
	Importance FileImportance
	Category   string // e.g., "pods", "nodes", "logs", "system"
}

// HealthCheck represents the result of a bundle health check
type HealthCheck struct {
	TotalFiles   int
	FoundFiles   int
	MissingFiles []MissingFile
	Completeness float64 // 0.0 to 100.0
	IsValid      bool    // Can be loaded (critical files present)
	BundleType   string  // "RKE2", "K3s", "kubectl", "unknown"
	Categories   map[string]CategoryHealth
}

// MissingFile represents a missing file with its impact
type MissingFile struct {
	Path       string
	Importance FileImportance
	Category   string
	Impact     string // Human-readable impact description
}

// CategoryHealth represents health for a specific category
type CategoryHealth struct {
	Total    int
	Found    int
	Missing  int
	Complete bool // 100% complete
}

// ExpectedFiles returns the list of files we expect in a bundle
// Sprint 8: RKE2 only for now, K3s to be added
// Sprint 10.1: Fixed to use actual collector script paths
func ExpectedFiles(bundlePath string) []ExpectedFile {
	// Detect if bundle has rke2 subdirectory
	hasRKE2 := false
	if _, err := os.Stat(filepath.Join(bundlePath, "rke2")); err == nil {
		hasRKE2 = true
	}

	// Build kubectl paths based on detected structure
	kubectlPrefix := ""
	if hasRKE2 {
		kubectlPrefix = "rke2/"
	}

	return []ExpectedFile{
		// Critical files (in rke2/kubectl/ if exists)
		{Path: kubectlPrefix + "kubectl/pods", Importance: ImportanceCritical, Category: "pods"},
		{Path: kubectlPrefix + "kubectl/nodes", Importance: ImportanceCritical, Category: "nodes"},

		// High importance (in rke2/kubectl/ if exists)
		{Path: kubectlPrefix + "kubectl/events", Importance: ImportanceHigh, Category: "events"},
		{Path: kubectlPrefix + "kubectl/deployments", Importance: ImportanceHigh, Category: "workloads"},
		// etcd is at ROOT level, not under rke2/
		{Path: "etcd/endpointstatus", Importance: ImportanceHigh, Category: "etcd"},

		// Medium importance
		{Path: kubectlPrefix + "kubectl/services", Importance: ImportanceMedium, Category: "networking"},
		{Path: kubectlPrefix + "kubectl/configmaps", Importance: ImportanceMedium, Category: "config"},
		// journald logs are in journald/ at root
		{Path: "journald/rke2-server", Importance: ImportanceMedium, Category: "logs"},

		// Low importance (nice to have)
		{Path: kubectlPrefix + "kubectl/crds", Importance: ImportanceLow, Category: "crds"},
		{Path: kubectlPrefix + "kubectl/pvc", Importance: ImportanceLow, Category: "storage"},
		// sysstat is in systemlogs/sysstat-data/
		{Path: "systemlogs/sysstat-data/", Importance: ImportanceLow, Category: "system"},
		// podlogs are in rke2/podlogs/ if rke2 exists
		{Path: kubectlPrefix + "podlogs/", Importance: ImportanceLow, Category: "logs"},
	}
}

// CheckHealth performs a health check on a bundle at the given path
func CheckHealth(bundlePath string) (*HealthCheck, error) {
	if _, err := os.Stat(bundlePath); err != nil {
		return nil, fmt.Errorf("cannot access bundle path: %w", err)
	}

	expected := ExpectedFiles(bundlePath)
	health := &HealthCheck{
		TotalFiles:   len(expected),
		Categories:   make(map[string]CategoryHealth),
		MissingFiles: []MissingFile{},
	}

	// Check each expected file
	for _, file := range expected {
		fullPath := filepath.Join(bundlePath, file.Path)
		found := false

		// Check if file or directory exists
		info, err := os.Stat(fullPath)
		if err == nil {
			// For directories, check if they have contents
			if info.IsDir() {
				entries, err := os.ReadDir(fullPath)
				if err == nil && len(entries) > 0 {
					found = true
				}
			} else {
				found = true
			}
		}

		if found {
			health.FoundFiles++
		} else {
			health.MissingFiles = append(health.MissingFiles, MissingFile{
				Path:       file.Path,
				Importance: file.Importance,
				Category:   file.Category,
				Impact:     impactDescription(file.Importance, file.Category),
			})
		}

		// Update category stats
		cat := health.Categories[file.Category]
		cat.Total++
		if found {
			cat.Found++
		} else {
			cat.Missing++
		}
		cat.Complete = (cat.Found == cat.Total)
		health.Categories[file.Category] = cat
	}

	// Check dmesg in both old and new locations (Issue #88)
	// Use path resolver to check both systeminfo/dmesg and systemlogs/dmesg
	if err := checkDmesgHealth(bundlePath, health); err != nil {
		// Log error but don't fail - dmesg is medium importance
		// Continue with health check
	}

	// Calculate completeness
	if health.TotalFiles > 0 {
		health.Completeness = float64(health.FoundFiles) / float64(health.TotalFiles) * 100
	}

	// Determine if bundle is valid (critical files must be present)
	health.IsValid = health.hasCriticalFiles()

	// Detect bundle type
	health.BundleType = detectBundleType(bundlePath)

	return health, nil
}

// checkDmesgHealth checks if dmesg exists in either old or new location (Issue #88)
// Uses GetDmesgPaths() to check both systeminfo/dmesg (new) and systemlogs/dmesg (old)
func checkDmesgHealth(bundlePath string, health *HealthCheck) error {
	// Detect bundle format and create appropriate resolver
	format := DetectFormat(bundlePath)
	resolver := NewPathResolver(bundlePath, format)

	// Get all possible dmesg paths
	dmesgPaths := resolver.GetDmesgPaths()

	// Check if dmesg exists in any location
	dmesgFound := false
	foundPath := ""
	for _, path := range dmesgPaths {
		if _, err := os.Stat(path); err == nil {
			dmesgFound = true
			foundPath = path
			break
		}
	}

	// Update health check with dmesg status
	// Use the first (new) path as the canonical path for reporting
	canonicalPath := "systeminfo/dmesg"
	if len(dmesgPaths) > 0 {
		// Extract relative path from the first dmesg path
		if relPath, err := filepath.Rel(bundlePath, dmesgPaths[0]); err == nil {
			canonicalPath = relPath
		}
	}

	health.TotalFiles++
	cat := health.Categories["system"]
	cat.Total++

	if dmesgFound {
		health.FoundFiles++
		cat.Found++
	} else {
		// Only report as missing if not found in either location
		health.MissingFiles = append(health.MissingFiles, MissingFile{
			Path:       canonicalPath,
			Importance: ImportanceMedium,
			Category:   "system",
			Impact:     impactDescription(ImportanceMedium, "system"),
		})
		cat.Missing++
	}
	cat.Complete = (cat.Found == cat.Total)
	health.Categories["system"] = cat

	// Suppress unused variable warning - foundPath indicates which path was found
	_ = foundPath

	return nil
}

// hasCriticalFiles returns true if all critical files are present
func (h *HealthCheck) hasCriticalFiles() bool {
	for _, missing := range h.MissingFiles {
		if missing.Importance == ImportanceCritical {
			return false
		}
	}
	return true
}

// impactDescription returns a human-readable impact description
func impactDescription(importance FileImportance, category string) string {
	switch importance {
	case ImportanceCritical:
		return fmt.Sprintf("Bundle analysis severely limited without %s data", category)
	case ImportanceHigh:
		return fmt.Sprintf("Major %s analysis features unavailable", category)
	case ImportanceMedium:
		return fmt.Sprintf("Minor %s features may be limited", category)
	case ImportanceLow:
		return fmt.Sprintf("Optional %s data unavailable", category)
	default:
		return "Unknown impact"
	}
}

// detectBundleType attempts to determine the bundle type
func detectBundleType(bundlePath string) string {
	// Check for RKE2 paths
	if _, err := os.Stat(filepath.Join(bundlePath, "rke2")); err == nil {
		return "RKE2"
	}
	// Check for K3s paths (Sprint 8: placeholder for future)
	if _, err := os.Stat(filepath.Join(bundlePath, "k3s")); err == nil {
		return "K3s"
	}
	// Check for generic kubectl
	if _, err := os.Stat(filepath.Join(bundlePath, "kubectl")); err == nil {
		return "kubectl"
	}
	return "unknown"
}

// Summary returns a one-line summary of the health check
func (h *HealthCheck) Summary() string {
	if !h.IsValid {
		return fmt.Sprintf("Bundle Health: %.0f%% 🔴 CRITICAL - missing required files", h.Completeness)
	}
	if h.Completeness < 50 {
		return fmt.Sprintf("Bundle Health: %.0f%% ⚠️  Partial bundle", h.Completeness)
	}
	if h.Completeness < 100 {
		return fmt.Sprintf("Bundle Health: %.0f%% ⚠️  Mostly complete", h.Completeness)
	}
	return fmt.Sprintf("Bundle Health: %.0f%% ✅ Complete", h.Completeness)
}

// GetHighImpactMissing returns missing files with high importance
func (h *HealthCheck) GetHighImpactMissing() []MissingFile {
	var high []MissingFile
	for _, m := range h.MissingFiles {
		if m.Importance == ImportanceCritical || m.Importance == ImportanceHigh {
			high = append(high, m)
		}
	}
	return high
}
