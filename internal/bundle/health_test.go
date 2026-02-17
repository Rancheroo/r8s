package bundle

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewHealthChecker(t *testing.T) {
	bundle := &Bundle{ExtractPath: "/tmp/test"}
	hc := NewHealthChecker(bundle)

	if hc == nil {
		t.Fatal("NewHealthChecker returned nil")
	}
	if hc.bundle != bundle {
		t.Error("HealthChecker.bundle mismatch")
	}
}

func TestHealthChecker_Check_EmptyBundle(t *testing.T) {
	hc := NewHealthChecker(nil)
	health := hc.Check()

	if len(health.Warnings) == 0 {
		t.Error("Expected warning for nil bundle")
	}
	if health.Percentage() != 100 {
		t.Errorf("Empty bundle should be 100%%, got %d%%", health.Percentage())
	}
}

func TestHealthChecker_Check_NoPath(t *testing.T) {
	bundle := &Bundle{ExtractPath: ""}
	hc := NewHealthChecker(bundle)
	health := hc.Check()

	if len(health.Warnings) == 0 {
		t.Error("Expected warning for empty path")
	}
}

func TestHealthChecker_Check_FullRKE2Bundle(t *testing.T) {
	// Create temp directory structure for full RKE2 bundle
	tmpDir := t.TempDir()

	// Create all critical directories
	dirs := []string{
		"rke2/kubectl",
		"rke2/podlogs",
		"rke2/pod-manifests",
		"rke2/agent-logs",
		"rke2/etcd",
		"systemlogs",
		"systeminfo",
	}
	for _, dir := range dirs {
		path := filepath.Join(tmpDir, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatalf("Failed to create %s: %v", dir, err)
		}
	}

	bundle := &Bundle{ExtractPath: tmpDir}
	hc := NewHealthChecker(bundle)
	health := hc.Check()

	// Should find all critical + optional files
	if health.FoundFiles == 0 {
		t.Error("Expected found files in full bundle")
	}
	if len(health.Warnings) > 0 {
		t.Errorf("Full bundle should have no warnings, got: %v", health.Warnings)
	}
	if health.BundleType != string(FormatRKE2) {
		t.Errorf("Expected format %s, got %s", FormatRKE2, health.BundleType)
	}
}

func TestHealthChecker_Check_PartialBundle(t *testing.T) {
	// Create temp directory with only some files (partial bundle)
	tmpDir := t.TempDir()

	// Only create kubectl, missing podlogs
	if err := os.MkdirAll(filepath.Join(tmpDir, "rke2", "kubectl"), 0755); err != nil {
		t.Fatalf("Failed to create kubectl dir: %v", err)
	}

	bundle := &Bundle{ExtractPath: tmpDir}
	hc := NewHealthChecker(bundle)
	health := hc.Check()

	// Should have warnings for missing data
	if len(health.Warnings) == 0 {
		t.Error("Expected warnings for partial bundle")
	}

	// Should detect missing pod logs warning
	hasPodLogWarning := false
	for _, w := range health.Warnings {
		if w == "Missing pod logs: container log analysis limited" {
			hasPodLogWarning = true
			break
		}
	}
	if !hasPodLogWarning {
		t.Errorf("Expected pod log warning, got: %v", health.Warnings)
	}
}

func TestHealthChecker_Check_K3sBundle(t *testing.T) {
	tmpDir := t.TempDir()

	// Create K3s structure
	dirs := []string{
		"k3s/kubectl",
		"k3s/podlogs",
	}
	for _, dir := range dirs {
		path := filepath.Join(tmpDir, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatalf("Failed to create %s: %v", dir, err)
		}
	}

	bundle := &Bundle{ExtractPath: tmpDir}
	hc := NewHealthChecker(bundle)
	health := hc.Check()

	if health.BundleType != string(FormatK3s) {
		t.Errorf("Expected format %s, got %s", FormatK3s, health.BundleType)
	}
}

func TestHealthChecker_Check_KubectlOnlyBundle(t *testing.T) {
	tmpDir := t.TempDir()

	// Create kubectl-only structure (no rke2/k3s)
	if err := os.MkdirAll(filepath.Join(tmpDir, "kubectl"), 0755); err != nil {
		t.Fatalf("Failed to create kubectl dir: %v", err)
	}

	bundle := &Bundle{ExtractPath: tmpDir}
	hc := NewHealthChecker(bundle)
	health := hc.Check()

	if health.BundleType != string(FormatKubectl) {
		t.Errorf("Expected format %s, got %s", FormatKubectl, health.BundleType)
	}
}

func TestHealthChecker_Check_UnknownBundle(t *testing.T) {
	tmpDir := t.TempDir()

	// Create random files without known structure
	if err := os.MkdirAll(filepath.Join(tmpDir, "random"), 0755); err != nil {
		t.Fatalf("Failed to create random dir: %v", err)
	}

	bundle := &Bundle{ExtractPath: tmpDir}
	hc := NewHealthChecker(bundle)
	health := hc.Check()

	if health.BundleType != string(FormatUnknown) {
		t.Errorf("Expected format %s, got %s", FormatUnknown, health.BundleType)
	}
}

func TestHealthChecker_pathExists(t *testing.T) {
	tmpDir := t.TempDir()
	hc := &HealthChecker{}

	// Test existing path
	if !hc.pathExists(tmpDir) {
		t.Error("pathExists should return true for existing directory")
	}

	// Test non-existing path
	fakePath := filepath.Join(tmpDir, "does-not-exist")
	if hc.pathExists(fakePath) {
		t.Error("pathExists should return false for non-existing path")
	}
}

func TestHealthChecker_hasKubectlData(t *testing.T) {
	tmpDir := t.TempDir()
	hc := &HealthChecker{}

	// No kubectl data initially
	if hc.hasKubectlData(tmpDir) {
		t.Error("Expected false with no kubectl data")
	}

	// Create RKE2 kubectl
	os.MkdirAll(filepath.Join(tmpDir, "rke2", "kubectl"), 0755)
	if !hc.hasKubectlData(tmpDir) {
		t.Error("Expected true with RKE2 kubectl")
	}

	// Test with K3s kubectl
	tmpDir2 := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir2, "k3s", "kubectl"), 0755)
	if !hc.hasKubectlData(tmpDir2) {
		t.Error("Expected true with K3s kubectl")
	}
}

func TestHealthChecker_hasPodLogs(t *testing.T) {
	tmpDir := t.TempDir()
	hc := &HealthChecker{}

	// Create RKE2 podlogs
	os.MkdirAll(filepath.Join(tmpDir, "rke2", "podlogs"), 0755)
	if !hc.hasPodLogs(tmpDir) {
		t.Error("Expected true with RKE2 podlogs")
	}

	// Test with K3s podlogs
	tmpDir2 := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir2, "k3s", "podlogs"), 0755)
	if !hc.hasPodLogs(tmpDir2) {
		t.Error("Expected true with K3s podlogs")
	}
}

func TestHealthChecker_hasSystemLogs(t *testing.T) {
	tmpDir := t.TempDir()
	hc := &HealthChecker{}

	// Create systemlogs
	os.MkdirAll(filepath.Join(tmpDir, "systemlogs"), 0755)
	if !hc.hasSystemLogs(tmpDir) {
		t.Error("Expected true with systemlogs")
	}

	// Test with agent-logs (RKE2)
	tmpDir2 := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir2, "rke2", "agent-logs"), 0755)
	if !hc.hasSystemLogs(tmpDir2) {
		t.Error("Expected true with agent-logs")
	}
}

func TestHealthChecker_detectFormat(t *testing.T) {
	hc := &HealthChecker{}

	// Test RKE2
	tmpDir1 := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir1, "rke2"), 0755)
	if hc.detectFormat(tmpDir1) != FormatRKE2 {
		t.Error("Expected FormatRKE2")
	}

	// Test K3s
	tmpDir2 := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir2, "k3s"), 0755)
	if hc.detectFormat(tmpDir2) != FormatK3s {
		t.Error("Expected FormatK3s")
	}

	// Test kubectl-only
	tmpDir3 := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir3, "kubectl"), 0755)
	if hc.detectFormat(tmpDir3) != FormatKubectl {
		t.Error("Expected FormatKubectl")
	}

	// Test unknown
	tmpDir4 := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir4, "random"), 0755)
	if hc.detectFormat(tmpDir4) != FormatUnknown {
		t.Error("Expected FormatUnknown")
	}
}

func TestBundleHealth_IsCriticalMissing(t *testing.T) {
	tests := []struct {
		name     string
		health   BundleHealth
		expected bool
	}{
		{
			name:     "Excellent - not critical",
			health:   BundleHealth{TotalFiles: 10, FoundFiles: 10},
			expected: false,
		},
		{
			name:     "Good - not critical",
			health:   BundleHealth{TotalFiles: 10, FoundFiles: 8},
			expected: false,
		},
		{
			name:     "Fair - critical",
			health:   BundleHealth{TotalFiles: 10, FoundFiles: 5},
			expected: true,
		},
		{
			name:     "Poor - critical",
			health:   BundleHealth{TotalFiles: 10, FoundFiles: 3},
			expected: true,
		},
		{
			name:     "Empty - not critical (100%)",
			health:   BundleHealth{TotalFiles: 0, FoundFiles: 0},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.health.IsCriticalMissing()
			if result != tt.expected {
				t.Errorf("IsCriticalMissing() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestBundleHealth_Summary(t *testing.T) {
	tests := []struct {
		name     string
		health   BundleHealth
		expected string
	}{
		{
			name:     "Excellent",
			health:   BundleHealth{TotalFiles: 10, FoundFiles: 10},
			expected: "Excellent (100%)",
		},
		{
			name:     "Good",
			health:   BundleHealth{TotalFiles: 10, FoundFiles: 8},
			expected: "Good (80%)",
		},
		{
			name:     "Fair",
			health:   BundleHealth{TotalFiles: 10, FoundFiles: 6},
			expected: "Fair (60%) — Some data may be incomplete",
		},
		{
			name:     "Poor",
			health:   BundleHealth{TotalFiles: 10, FoundFiles: 3},
			expected: "Poor (30%) — Critical data missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.health.Summary()
			if result != tt.expected {
				t.Errorf("Summary() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestCriticalFiles_NotEmpty(t *testing.T) {
	if len(CriticalFiles) == 0 {
		t.Error("CriticalFiles should not be empty")
	}
}

func TestOptionalFiles_NotEmpty(t *testing.T) {
	if len(OptionalFiles) == 0 {
		t.Error("OptionalFiles should not be empty")
	}
}
