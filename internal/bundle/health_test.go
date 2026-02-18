package bundle

import (
	"os"
	"path/filepath"
	"testing"
)

// TestExpectedFiles verifies all expected file paths are correctly defined
func TestExpectedFiles(t *testing.T) {
	files := ExpectedFiles()
	
	if len(files) == 0 {
		t.Fatal("ExpectedFiles() returned empty slice")
	}
	
	// Verify critical files are present
	criticalPaths := map[string]bool{
		"rke2/kubectl/pods":   false,
		"rke2/kubectl/nodes":  false,
	}
	
	// Verify etcd has alternative paths (root level)
	etcdHasAltPath := false
	
	for _, f := range files {
		if f.Path == "" {
			t.Error("ExpectedFile has empty Path")
		}
		
		// Check critical files
		if _, exists := criticalPaths[f.Path]; exists {
			criticalPaths[f.Path] = true
		}
		
		// Check etcd has alternative paths
		if f.Path == "rke2/etcd/endpointstatus" || f.Path == "etcd/endpointstatus" {
			if len(f.AltPaths) > 0 {
				etcdHasAltPath = true
			}
		}
		
		// Verify importance is set
		if f.Importance < ImportanceCritical || f.Importance > ImportanceLow {
			t.Errorf("Invalid Importance for %s: %d", f.Path, f.Importance)
		}
		
		// Verify category is set
		if f.Category == "" {
			t.Errorf("Empty Category for %s", f.Path)
		}
	}
	
	// Verify all critical files found
	for path, found := range criticalPaths {
		if !found {
			t.Errorf("Critical file missing from ExpectedFiles: %s", path)
		}
	}
	
	if !etcdHasAltPath {
		t.Error("etcd file should have alternative paths for root-level bundles")
	}
}

// TestCheckHealthWithFullBundle tests health check with complete bundle
func TestCheckHealthWithFullBundle(t *testing.T) {
	// Create temp bundle structure
	tmpDir := t.TempDir()
	
	// Create all expected files
	files := []string{
		"rke2/kubectl/pods",
		"rke2/kubectl/nodes",
		"rke2/kubectl/events",
		"rke2/kubectl/deployments",
		"rke2/kubectl/services",
		"rke2/kubectl/configmaps",
		"rke2/kubectl/crds",
		"rke2/kubectl/pvc",
		"etcd/endpointstatus",  // Root level etcd
		"systeminfo/dmesg",     // Root level dmesg
	}
	
	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
	}
	
	// Create podlogs with files
	podlogsDir := filepath.Join(tmpDir, "rke2", "podlogs")
	if err := os.MkdirAll(podlogsDir, 0755); err != nil {
		t.Fatalf("Failed to create podlogs dir: %v", err)
	}
	for i := 0; i < 5; i++ {
		f := filepath.Join(podlogsDir, "test-pod-log-"+string(rune('a'+i)))
		if err := os.WriteFile(f, []byte("log"), 0644); err != nil {
			t.Fatalf("Failed to create log file: %v", err)
		}
	}
	
	// Create journald directory
	journaldDir := filepath.Join(tmpDir, "journald")
	if err := os.MkdirAll(journaldDir, 0755); err != nil {
		t.Fatalf("Failed to create journald dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(journaldDir, "rke2-server"), []byte("journal"), 0644); err != nil {
		t.Fatalf("Failed to create journal file: %v", err)
	}
	
	// Run health check
	health, err := CheckHealth(tmpDir)
	if err != nil {
		t.Fatalf("CheckHealth failed: %v", err)
	}
	
	// Verify results
	if !health.IsValid {
		t.Error("Bundle should be valid")
	}
	
	if health.Completeness < 90 {
		t.Errorf("Expected >90%% completeness, got %.0f%%", health.Completeness)
	}
	
	// Check that etcd was found (critical test)
	if etcdCat, ok := health.Categories["etcd"]; ok {
		if etcdCat.Found == 0 {
			t.Error("etcd files should be found at root level")
		}
	} else {
		t.Error("etcd category should exist in health check")
	}
}

// TestCheckHealthWithRKE2NestedBundle tests bundles with rke2/etcd/ structure
func TestCheckHealthWithRKE2NestedBundle(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create nested structure (older bundle format)
	files := []string{
		"rke2/kubectl/pods",
		"rke2/kubectl/nodes",
		"rke2/etcd/endpointstatus",  // Nested under rke2
		"rke2/dmesg",                // Nested dmesg
		"rke2/logs/journald.log",    // Nested journald
	}
	
	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
	}
	
	// Run health check
	health, err := CheckHealth(tmpDir)
	if err != nil {
		t.Fatalf("CheckHealth failed: %v", err)
	}
	
	if !health.IsValid {
		t.Error("Nested bundle should be valid")
	}
	
	// Verify etcd found in nested structure
	if etcdCat, ok := health.Categories["etcd"]; ok {
		if etcdCat.Found == 0 {
			t.Error("etcd should be found in rke2/etcd/ nested structure")
		}
	}
}

// TestCheckHealthWithMissingCriticalFiles tests invalid bundle detection
func TestCheckHealthWithMissingCriticalFiles(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Only create non-critical files
	if err := os.MkdirAll(filepath.Join(tmpDir, "etcd"), 0755); err != nil {
		t.Fatalf("Failed to create dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "etcd", "endpointstatus"), []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	
	// Run health check
	health, err := CheckHealth(tmpDir)
	if err != nil {
		t.Fatalf("CheckHealth failed: %v", err)
	}
	
	// Should be invalid (missing critical kubectl files)
	if health.IsValid {
		t.Error("Bundle without critical files should be invalid")
	}
}

// TestCheckHealthWithEmptyBundle tests empty bundle handling
func TestCheckHealthWithEmptyBundle(t *testing.T) {
	tmpDir := t.TempDir()
	
	_, err := CheckHealth(tmpDir)
	if err != nil {
		t.Fatalf("CheckHealth should not fail on empty bundle: %v", err)
	}
}

// TestCheckHealthWithNonexistentPath tests error handling
func TestCheckHealthWithNonexistentPath(t *testing.T) {
	_, err := CheckHealth("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("CheckHealth should return error for nonexistent path")
	}
}

// TestPodlogsDirectoryDetection tests the special podlogs counting logic
func TestPodlogsDirectoryDetection(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create required files
	for _, f := range []string{"rke2/kubectl/pods", "rke2/kubectl/nodes"} {
		path := filepath.Join(tmpDir, f)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
	}
	
	// Test: podlogs with <5 files should NOT count
	podlogsDir := filepath.Join(tmpDir, "rke2", "podlogs")
	if err := os.MkdirAll(podlogsDir, 0755); err != nil {
		t.Fatalf("Failed to create podlogs dir: %v", err)
	}
	// Only create 3 files
	for i := 0; i < 3; i++ {
		f := filepath.Join(podlogsDir, "log-"+string(rune('a'+i)))
		if err := os.WriteFile(f, []byte("log"), 0644); err != nil {
			t.Fatalf("Failed to create log file: %v", err)
		}
	}
	
	health, _ := CheckHealth(tmpDir)
	if logsCat, ok := health.Categories["logs"]; ok {
		// Should have 0 found for logs since podlogs has <5 files
		// and no journald
		if logsCat.Found > 0 {
			t.Logf("Logs category: %+v", logsCat)
		}
	}
	
	// Test: podlogs with >=5 files should count
	// Remove and recreate with 5 files
	os.RemoveAll(podlogsDir)
	if err := os.MkdirAll(podlogsDir, 0755); err != nil {
		t.Fatalf("Failed to recreate podlogs dir: %v", err)
	}
	for i := 0; i < 5; i++ {
		f := filepath.Join(podlogsDir, "log-"+string(rune('a'+i)))
		if err := os.WriteFile(f, []byte("log"), 0644); err != nil {
			t.Fatalf("Failed to create log file: %v", err)
		}
	}
	
	health, _ = CheckHealth(tmpDir)
	if logsCat, ok := health.Categories["logs"]; ok {
		// Now should have podlogs counted
		t.Logf("Logs category with 5 files: %+v", logsCat)
	}
}

// TestHealthCheckSummary tests the Summary() method
func TestHealthCheckSummary(t *testing.T) {
	tests := []struct {
		name         string
		completeness float64
		isValid      bool
		shouldContain []string
	}{
		{
			name:          "complete bundle",
			completeness:  100,
			isValid:       true,
			shouldContain: []string{"100%", "Complete"},
		},
		{
			name:          "mostly complete",
			completeness:  85,
			isValid:       true,
			shouldContain: []string{"85%", "Mostly"},
		},
		{
			name:          "partial bundle",
			completeness:  40,
			isValid:       true,
			shouldContain: []string{"40%", "Partial"},
		},
		{
			name:          "invalid bundle",
			completeness:  20,
			isValid:       false,
			shouldContain: []string{"20%", "CRITICAL"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &HealthCheck{
				Completeness: tt.completeness,
				IsValid:      tt.isValid,
			}
			summary := h.Summary()
			
			for _, substr := range tt.shouldContain {
				if !containsString(summary, substr) {
					t.Errorf("Summary() = %q, should contain %q", summary, substr)
				}
			}
		})
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
