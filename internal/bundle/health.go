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

// ExpectedFile represents a file that should be present in a bundle
type ExpectedFile struct {
	Path       string
	AltPaths   []string // Alternative paths to check (for bundle format variations)
	Importance FileImportance
	Category   string // e.g., "pods", "nodes", "logs", "system"
}

// HealthCheck represents the result of a bundle health check
type HealthCheck struct {
	TotalFiles     int
	FoundFiles     int
	MissingFiles   []MissingFile
	Completeness   float64 // 0.0 to 100.0
	IsValid        bool    // Can be loaded (critical files present)
	BundleType     string  // "RKE2", "K3s", "kubectl", "unknown"
	Categories     map[string]CategoryHealth
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
// NOTE: Bundle format varies by collector version - some files are at root, some under rke2/
func ExpectedFiles() []ExpectedFile {
	return []ExpectedFile{
		// Critical files
		{Path: "rke2/kubectl/pods", Importance: ImportanceCritical, Category: "pods"},
		{Path: "rke2/kubectl/nodes", Importance: ImportanceCritical, Category: "nodes"},

		// High importance
		{Path: "rke2/kubectl/events", Importance: ImportanceHigh, Category: "events"},
		{Path: "rke2/kubectl/deployments", Importance: ImportanceHigh, Category: "workloads"},
		{Path: "rke2/etcd/endpointstatus", AltPaths: []string{"etcd/endpointstatus"}, Importance: ImportanceHigh, Category: "etcd"},

		// Medium importance
		{Path: "rke2/kubectl/services", Importance: ImportanceMedium, Category: "networking"},
		{Path: "rke2/kubectl/configmaps", Importance: ImportanceMedium, Category: "config"},
		{Path: "rke2/dmesg", AltPaths: []string{"systeminfo/dmesg", "systemlogs/dmesg"}, Importance: ImportanceMedium, Category: "system"},
		{Path: "rke2/logs/journald.log", AltPaths: []string{"journald/", "systemlogs/journald.log"}, Importance: ImportanceMedium, Category: "logs"},

		// Low importance (nice to have)
		{Path: "rke2/kubectl/crds", Importance: ImportanceLow, Category: "crds"},
		{Path: "rke2/kubectl/pvc", Importance: ImportanceLow, Category: "storage"},
		{Path: "rke2/sysstat/", AltPaths: []string{"sysstat/"}, Importance: ImportanceLow, Category: "system"},
		{Path: "rke2/podlogs/", Importance: ImportanceLow, Category: "logs"},
	}
}

// CheckHealth performs a health check on a bundle at the given path
func CheckHealth(bundlePath string) (*HealthCheck, error) {
	if _, err := os.Stat(bundlePath); err != nil {
		return nil, fmt.Errorf("cannot access bundle path: %w", err)
	}

	expected := ExpectedFiles()
	health := &HealthCheck{
		TotalFiles:   len(expected),
		Categories:   make(map[string]CategoryHealth),
		MissingFiles: []MissingFile{},
	}

	// Detect bundle type first for type-specific checks
	bundleType := detectBundleType(bundlePath)
	health.BundleType = bundleType

	// Check each expected file
	for _, file := range expected {
		found := false

		// Check primary path and all alternative paths
		pathsToCheck := append([]string{file.Path}, file.AltPaths...)
		for _, p := range pathsToCheck {
			fullPath := filepath.Join(bundlePath, p)
			info, err := os.Stat(fullPath)
			if err == nil {
				// For directories, check if they have contents
				if info.IsDir() {
					entries, err := os.ReadDir(fullPath)
					if err == nil && len(entries) > 0 {
						found = true
						break
					}
				} else {
					found = true
					break
				}
			}
		}

		// Special handling for podlogs directory - check if it actually contains log files
		if file.Path == "rke2/podlogs/" || file.Path == "k3s/podlogs/" || file.Path == "podlogs/" {
			podlogsPath := filepath.Join(bundlePath, file.Path)
			if info, err := os.Stat(podlogsPath); err == nil && info.IsDir() {
				// Count actual log files
				entries, err := os.ReadDir(podlogsPath)
				if err == nil {
					logFileCount := 0
					for _, entry := range entries {
						if !entry.IsDir() {
							logFileCount++
						}
					}
					// Consider podlogs present if it has at least 5 log files
					if logFileCount >= 5 {
						found = true
					}
				}
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

	// Calculate completeness
	if health.TotalFiles > 0 {
		health.Completeness = float64(health.FoundFiles) / float64(health.TotalFiles) * 100
	}

	// Determine if bundle is valid (critical files must be present)
	health.IsValid = health.hasCriticalFiles()

	return health, nil
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
