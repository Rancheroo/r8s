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
			wantComplete:   23.0, // 3/13 files - exact match
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
			wantComplete:   7.7, // 1/13 but missing critical pods - RKE2 detected
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
	files := ExpectedFiles()
	if len(files) == 0 {
		t.Error("ExpectedFiles() returned empty slice")
	}

	// Check for critical files
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
