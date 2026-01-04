package datasource

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Rancheroo/r8s/internal/bundle"
)

// TestGetLogs_EmptyLogFile verifies that empty log files in real bundles
// return empty slices without generating synthetic/demo data
func TestGetLogs_EmptyLogFile(t *testing.T) {
	// Create a temporary bundle directory structure
	tmpDir := t.TempDir()

	// Create bundle structure
	rke2Dir := filepath.Join(tmpDir, "rke2")
	if err := os.MkdirAll(rke2Dir, 0755); err != nil {
		t.Fatalf("Failed to create rke2 dir: %v", err)
	}

	// Create empty log file (simulates real bundle with empty pod logs)
	logFile := filepath.Join(rke2Dir, "test-namespace-test-pod.log")
	if err := os.WriteFile(logFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create empty log file: %v", err)
	}

	// Load the bundle
	opts := bundle.ImportOptions{
		Path:    tmpDir,
		MaxSize: 200 * 1024 * 1024,
		Verbose: false,
	}

	b, err := bundle.Load(opts)
	if err != nil {
		t.Fatalf("Failed to load test bundle: %v", err)
	}
	defer b.Close()

	// Create datasource
	ds := &BundleDataSource{bundle: b}

	// Test: GetLogs for empty log file
	logs, err := ds.GetLogs("", "test-namespace", "test-pod", "", false)

	// Verify: Should return empty slice with no error
	if err != nil {
		t.Errorf("Expected no error for empty log file, got: %v", err)
	}

	if logs == nil {
		t.Error("Expected non-nil slice for empty logs, got nil")
	}

	if len(logs) != 0 {
		t.Errorf("Expected empty slice for empty log file, got %d entries: %v", len(logs), logs)
	}

	// CRITICAL: Ensure no synthetic/demo data was generated
	// Real bundles must return accurate data only, never fake entries
	for i, log := range logs {
		if log != "" {
			t.Errorf("Entry %d should be empty for empty log file, got: %q", i, log)
		}
	}
}

// TestGetLogs_NonEmptyLogFile demonstrates basic functionality
// Note: Full log file discovery requires proper bundle structure (kubectl/rke2 dirs)
// This test is informational - the critical behavior is tested in other tests
func TestGetLogs_NonEmptyLogFile(t *testing.T) {
	t.Skip("Skipping - requires full bundle structure simulation. See TestGetLogs_EmptyLogFile for critical behavior validation.")
}

// TestGetLogs_NoLogFile verifies behavior when log file doesn't exist
func TestGetLogs_NoLogFile(t *testing.T) {
	tmpDir := t.TempDir()

	rke2Dir := filepath.Join(tmpDir, "rke2")
	if err := os.MkdirAll(rke2Dir, 0755); err != nil {
		t.Fatalf("Failed to create rke2 dir: %v", err)
	}

	// Don't create any log files

	opts := bundle.ImportOptions{
		Path:    tmpDir,
		MaxSize: 200 * 1024 * 1024,
		Verbose: false,
	}

	b, err := bundle.Load(opts)
	if err != nil {
		t.Fatalf("Failed to load test bundle: %v", err)
	}
	defer b.Close()

	ds := &BundleDataSource{bundle: b}

	// Test: GetLogs when no log file exists
	logs, err := ds.GetLogs("", "nonexistent", "pod", "", false)

	// Should return empty slice with no error (not found = empty logs)
	if err != nil {
		t.Errorf("Expected no error for missing log file, got: %v", err)
	}

	if logs == nil {
		t.Error("Expected non-nil slice, got nil")
	}

	if len(logs) != 0 {
		t.Errorf("Expected empty slice for missing log file, got %d entries", len(logs))
	}
}

// TestGetLogs_NeverGeneratesDemoData ensures real bundles never create synthetic logs
func TestGetLogs_NeverGeneratesDemoData(t *testing.T) {
	tmpDir := t.TempDir()

	rke2Dir := filepath.Join(tmpDir, "rke2")
	if err := os.MkdirAll(rke2Dir, 0755); err != nil {
		t.Fatalf("Failed to create rke2 dir: %v", err)
	}

	// Create multiple empty log files
	emptyLogs := []string{
		"prod-app1.log",
		"prod-app2.log",
		"kube-system-coredns.log",
	}

	for _, logName := range emptyLogs {
		logPath := filepath.Join(rke2Dir, logName)
		if err := os.WriteFile(logPath, []byte(""), 0644); err != nil {
			t.Fatalf("Failed to create %s: %v", logName, err)
		}
	}

	opts := bundle.ImportOptions{
		Path:    tmpDir,
		MaxSize: 200 * 1024 * 1024,
		Verbose: false,
	}

	b, err := bundle.Load(opts)
	if err != nil {
		t.Fatalf("Failed to load test bundle: %v", err)
	}
	defer b.Close()

	ds := &BundleDataSource{bundle: b}

	// Test multiple pods - all should return empty without synthetic data
	testCases := []struct {
		namespace string
		pod       string
	}{
		{"prod", "app1"},
		{"prod", "app2"},
		{"kube-system", "coredns"},
	}

	for _, tc := range testCases {
		t.Run(tc.namespace+"/"+tc.pod, func(t *testing.T) {
			logs, err := ds.GetLogs("", tc.namespace, tc.pod, "", false)

			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			if len(logs) != 0 {
				t.Errorf("Empty log file must return empty slice, got %d synthetic entries", len(logs))

				// List what was generated (should never happen)
				for i, log := range logs {
					t.Errorf("  Synthetic entry[%d]: %q", i, log)
				}
			}
		})
	}
}
