package bundle

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckHealth(t *testing.T) {
	// Create temporary test directories
	tmpDir := t.TempDir()

	tests := []struct {
		name           string
		setup          func() string
		wantValid      bool
		wantComplete   float64
		wantBundleType string
		wantErr        bool
	}{
		{
			name: "empty directory",
			setup: func() string {
				path := filepath.Join(tmpDir, "empty")
				os.MkdirAll(path, 0755)
				return path
			},
			wantValid:      false,
			wantComplete:   0,
			wantBundleType: "unknown",
			wantErr:        false,
		},
		{
			name: "complete RKE2 bundle",
			setup: func() string {
				path := filepath.Join(tmpDir, "complete")
				os.MkdirAll(filepath.Join(path, "rke2/kubectl"), 0755)
				os.MkdirAll(filepath.Join(path, "rke2/etcd"), 0755)
				os.WriteFile(filepath.Join(path, "rke2/kubectl/pods"), []byte("test"), 0644)
				os.WriteFile(filepath.Join(path, "rke2/kubectl/nodes"), []byte("test"), 0644)
				os.WriteFile(filepath.Join(path, "rke2/etcd/endpoint_status"), []byte("test"), 0644)
				return path
			},
			wantValid:      true,
			wantComplete:   15.4, // 3/20 files - dmesg now consolidated (Issue #88)
			wantBundleType: "RKE2",
			wantErr:        false,
		},
		{
			name: "missing critical files",
			setup: func() string {
				path := filepath.Join(tmpDir, "missing-critical")
				os.MkdirAll(filepath.Join(path, "rke2/kubectl"), 0755)
				// Only create nodes, missing pods
				os.WriteFile(filepath.Join(path, "rke2/kubectl/nodes"), []byte("test"), 0644)
				return path
			},
			wantValid:      false,
			wantComplete:   7.7, // 1/13 approx - missing critical pods - RKE2 detected (Issue #88: dmesg consolidated)
			wantBundleType: "RKE2",
			wantErr:        false,
		},
		{
			name: "K3s bundle",
			setup: func() string {
				path := filepath.Join(tmpDir, "k3s")
				os.MkdirAll(filepath.Join(path, "k3s"), 0755)
				os.WriteFile(filepath.Join(path, "k3s/pods"), []byte("test"), 0644)
				return path
			},
			wantValid:      false,
			wantComplete:   0,
			wantBundleType: "K3s",
			wantErr:        false,
		},
		{
			name: "non-existent path",
			setup: func() string {
				return filepath.Join(tmpDir, "does-not-exist")
			},
			wantValid: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()
			health, err := CheckHealth(path)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CheckHealth() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("CheckHealth() unexpected error: %v", err)
				return
			}

			if health.IsValid != tt.wantValid {
				t.Errorf("IsValid = %v, want %v", health.IsValid, tt.wantValid)
			}

			if health.BundleType != tt.wantBundleType {
				t.Errorf("BundleType = %s, want %s", health.BundleType, tt.wantBundleType)
			}

			// Check completeness is roughly correct (allow for expected file count changes)
			if tt.wantComplete > 0 {
				// Use tolerance for float comparison
				diff := health.Completeness - tt.wantComplete
				if diff < 0 {
					diff = -diff
				}
				if diff > 1.0 { // Allow 1% difference
					t.Errorf("Completeness = %.1f, want %.1f (diff: %.1f)", health.Completeness, tt.wantComplete, diff)
				}
			}
		})
	}
}

func TestHealthCheck_Summary(t *testing.T) {
	tests := []struct {
		name   string
		health *HealthCheck
		want   string
	}{
		{
			name: "complete bundle",
			health: &HealthCheck{
				Completeness: 100,
				IsValid:      true,
			},
			want: "Bundle Health: 100% ✅ Complete",
		},
		{
			name: "mostly complete",
			health: &HealthCheck{
				Completeness: 75,
				IsValid:      true,
			},
			want: "Bundle Health: 75% ⚠️  Mostly complete",
		},
		{
			name: "partial bundle",
			health: &HealthCheck{
				Completeness: 30,
				IsValid:      true,
			},
			want: "Bundle Health: 30% ⚠️  Partial bundle",
		},
		{
			name: "invalid bundle",
			health: &HealthCheck{
				Completeness: 10,
				IsValid:      false,
			},
			want: "Bundle Health: 10% 🔴 CRITICAL - missing required files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.health.Summary()
			if got != tt.want {
				t.Errorf("Summary() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMissingFile_Impact(t *testing.T) {
	tests := []struct {
		importance FileImportance
		category   string
		want       string
	}{
		{
			importance: ImportanceCritical,
			category:   "pods",
			want:       "Bundle analysis severely limited without pods data",
		},
		{
			importance: ImportanceHigh,
			category:   "events",
			want:       "Major events analysis features unavailable",
		},
		{
			importance: ImportanceMedium,
			category:   "logs",
			want:       "Minor logs features may be limited",
		},
		{
			importance: ImportanceLow,
			category:   "system",
			want:       "Optional system data unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.want[:20], func(t *testing.T) {
			got := impactDescription(tt.importance, tt.category)
			if got != tt.want {
				t.Errorf("impactDescription() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExpectedFiles(t *testing.T) {
	// Create a temp bundle with rke2 structure
	tmpDir := t.TempDir()
	rke2Dir := filepath.Join(tmpDir, "rke2")
	if err := os.MkdirAll(rke2Dir, 0755); err != nil {
		t.Fatalf("failed to create rke2 dir: %v", err)
	}

	files := ExpectedFiles(tmpDir)
	if len(files) == 0 {
		t.Error("ExpectedFiles() returned empty slice")
	}

	// Check for critical files (with rke2 prefix since we created rke2 dir)
	hasPods := false
	hasNodes := false
	for _, f := range files {
		if f.Path == "rke2/kubectl/pods" && f.Importance == ImportanceCritical {
			hasPods = true
		}
		if f.Path == "rke2/kubectl/nodes" && f.Importance == ImportanceCritical {
			hasNodes = true
		}
	}

	if !hasPods {
		t.Error("ExpectedFiles() missing critical pods file")
	}
	if !hasNodes {
		t.Error("ExpectedFiles() missing critical nodes file")
	}
}

func TestHealthCheck_GetHighImpactMissing(t *testing.T) {
	tests := []struct {
		name     string
		missing  []MissingFile
		expected int
	}{
		{
			name: "only high impact",
			missing: []MissingFile{
				{Path: "file1", Importance: ImportanceHigh},
				{Path: "file2", Importance: ImportanceHigh},
			},
			expected: 2,
		},
		{
			name: "critical and high",
			missing: []MissingFile{
				{Path: "file1", Importance: ImportanceCritical},
				{Path: "file2", Importance: ImportanceHigh},
				{Path: "file3", Importance: ImportanceMedium},
			},
			expected: 2, // Only critical + high
		},
		{
			name: "mixed importance",
			missing: []MissingFile{
				{Path: "file1", Importance: ImportanceCritical},
				{Path: "file2", Importance: ImportanceHigh},
				{Path: "file3", Importance: ImportanceMedium},
				{Path: "file4", Importance: ImportanceLow},
			},
			expected: 2, // Only critical + high
		},
		{
			name:     "no missing files",
			missing:  []MissingFile{},
			expected: 0,
		},
		{
			name: "only low and medium",
			missing: []MissingFile{
				{Path: "file1", Importance: ImportanceMedium},
				{Path: "file2", Importance: ImportanceLow},
			},
			expected: 0, // Should be empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &HealthCheck{MissingFiles: tt.missing}
			got := h.GetHighImpactMissing()
			if len(got) != tt.expected {
				t.Errorf("GetHighImpactMissing() returned %d files, want %d", len(got), tt.expected)
			}
		})
	}
}

func TestDetectBundleType(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		setup    func() string
		expected string
	}{
		{
			name: "RKE2 bundle",
			setup: func() string {
				path := filepath.Join(tmpDir, "rke2-bundle")
				os.MkdirAll(filepath.Join(path, "rke2"), 0755)
				return path
			},
			expected: "RKE2",
		},
		{
			name: "K3s bundle",
			setup: func() string {
				path := filepath.Join(tmpDir, "k3s-bundle")
				os.MkdirAll(filepath.Join(path, "k3s"), 0755)
				return path
			},
			expected: "K3s",
		},
		{
			name: "kubectl bundle",
			setup: func() string {
				path := filepath.Join(tmpDir, "kubectl-bundle")
				os.MkdirAll(filepath.Join(path, "kubectl"), 0755)
				return path
			},
			expected: "kubectl",
		},
		{
			name: "unknown bundle",
			setup: func() string {
				path := filepath.Join(tmpDir, "unknown-bundle")
				os.MkdirAll(path, 0755)
				return path
			},
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()
			got := detectBundleType(path)
			if got != tt.expected {
				t.Errorf("detectBundleType() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestDmesgFallback tests Issue #88: dmesg location fallback
// Verifies that dmesg is found in either old (systemlogs/) or new (systeminfo/) location
func TestDmesgFallback(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name           string
		setup          func() string
		wantDmesgFound bool
		wantMissing    bool
	}{
		{
			name: "dmesg in new location (systeminfo/dmesg)",
			setup: func() string {
				path := filepath.Join(tmpDir, "new-location")
				os.MkdirAll(filepath.Join(path, "rke2/kubectl"), 0755)
				os.MkdirAll(filepath.Join(path, "systeminfo"), 0755)
				// Create required critical files
				os.WriteFile(filepath.Join(path, "rke2/kubectl/pods"), []byte("test"), 0644)
				os.WriteFile(filepath.Join(path, "rke2/kubectl/nodes"), []byte("test"), 0644)
				// Create dmesg in new location
				os.WriteFile(filepath.Join(path, "systeminfo/dmesg"), []byte("dmesg content"), 0644)
				return path
			},
			wantDmesgFound: true,
			wantMissing:    false,
		},
		{
			name: "dmesg in old location (systemlogs/dmesg)",
			setup: func() string {
				path := filepath.Join(tmpDir, "old-location")
				os.MkdirAll(filepath.Join(path, "rke2/kubectl"), 0755)
				os.MkdirAll(filepath.Join(path, "systemlogs"), 0755)
				// Create required critical files
				os.WriteFile(filepath.Join(path, "rke2/kubectl/pods"), []byte("test"), 0644)
				os.WriteFile(filepath.Join(path, "rke2/kubectl/nodes"), []byte("test"), 0644)
				// Create dmesg in old location
				os.WriteFile(filepath.Join(path, "systemlogs/dmesg"), []byte("dmesg content"), 0644)
				return path
			},
			wantDmesgFound: true,
			wantMissing:    false,
		},
		{
			name: "dmesg in both locations (should be found)",
			setup: func() string {
				path := filepath.Join(tmpDir, "both-locations")
				os.MkdirAll(filepath.Join(path, "rke2/kubectl"), 0755)
				os.MkdirAll(filepath.Join(path, "systeminfo"), 0755)
				os.MkdirAll(filepath.Join(path, "systemlogs"), 0755)
				// Create required critical files
				os.WriteFile(filepath.Join(path, "rke2/kubectl/pods"), []byte("test"), 0644)
				os.WriteFile(filepath.Join(path, "rke2/kubectl/nodes"), []byte("test"), 0644)
				// Create dmesg in both locations
				os.WriteFile(filepath.Join(path, "systeminfo/dmesg"), []byte("new dmesg"), 0644)
				os.WriteFile(filepath.Join(path, "systemlogs/dmesg"), []byte("old dmesg"), 0644)
				return path
			},
			wantDmesgFound: true,
			wantMissing:    false,
		},
		{
			name: "dmesg missing in both locations",
			setup: func() string {
				path := filepath.Join(tmpDir, "missing-dmesg")
				os.MkdirAll(filepath.Join(path, "rke2/kubectl"), 0755)
				// Create required critical files
				os.WriteFile(filepath.Join(path, "rke2/kubectl/pods"), []byte("test"), 0644)
				os.WriteFile(filepath.Join(path, "rke2/kubectl/nodes"), []byte("test"), 0644)
				// No dmesg created
				return path
			},
			wantDmesgFound: false,
			wantMissing:    true,
		},
		{
			name: "dmesg fallback with K3s bundle (new location)",
			setup: func() string {
				path := filepath.Join(tmpDir, "k3s-new-location")
				os.MkdirAll(filepath.Join(path, "k3s/kubectl"), 0755)
				os.MkdirAll(filepath.Join(path, "systeminfo"), 0755)
				// Create required critical files
				os.WriteFile(filepath.Join(path, "k3s/kubectl/pods"), []byte("test"), 0644)
				os.WriteFile(filepath.Join(path, "k3s/kubectl/nodes"), []byte("test"), 0644)
				// Create dmesg in new location
				os.WriteFile(filepath.Join(path, "systeminfo/dmesg"), []byte("dmesg content"), 0644)
				return path
			},
			wantDmesgFound: true,
			wantMissing:    false,
		},
		{
			name: "dmesg fallback with K3s bundle (old location)",
			setup: func() string {
				path := filepath.Join(tmpDir, "k3s-old-location")
				os.MkdirAll(filepath.Join(path, "k3s/kubectl"), 0755)
				os.MkdirAll(filepath.Join(path, "systemlogs"), 0755)
				// Create required critical files
				os.WriteFile(filepath.Join(path, "k3s/kubectl/pods"), []byte("test"), 0644)
				os.WriteFile(filepath.Join(path, "k3s/kubectl/nodes"), []byte("test"), 0644)
				// Create dmesg in old location
				os.WriteFile(filepath.Join(path, "systemlogs/dmesg"), []byte("dmesg content"), 0644)
				return path
			},
			wantDmesgFound: true,
			wantMissing:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()
			health, err := CheckHealth(path)
			if err != nil {
				t.Fatalf("CheckHealth() error = %v", err)
			}

			// Check if dmesg is in missing files
			dmesgMissing := false
			for _, missing := range health.MissingFiles {
				if missing.Path == "systeminfo/dmesg" {
					dmesgMissing = true
					break
				}
			}

			// Check system category stats
			sysCat, hasSystem := health.Categories["system"]

			if tt.wantDmesgFound {
				// Dmesg should be found (not in missing files)
				if dmesgMissing {
					t.Errorf("dmesg reported as missing but should be found in one of the locations")
				}
				// System category should show found
				if hasSystem && sysCat.Found == 0 {
					t.Errorf("system category has Found=0 but dmesg should be counted")
				}
			} else {
				// Dmesg should be missing
				if !dmesgMissing {
					t.Errorf("dmesg should be reported as missing but was not found in missing files")
				}
			}

			// Verify system category exists
			if !hasSystem {
				t.Errorf("system category not found in health check")
			}
		})
	}
}
