package bundle

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		name           string
		setupDirs      []string
		expectedFormat BundleFormat
	}{
		{
			name:           "RKE2 direct structure",
			setupDirs:      []string{"rke2", "rke2/kubectl"},
			expectedFormat: FormatRKE2,
		},
		{
			name:           "K3s direct structure",
			setupDirs:      []string{"k3s", "k3s/kubectl"},
			expectedFormat: FormatK3s,
		},
		{
			name:           "Kubectl cluster-info structure",
			setupDirs:      []string{"namespaces", "nodes"},
			expectedFormat: FormatKubectl,
		},
		{
			name:           "RKE2 with wrapper directory",
			setupDirs:      []string{"wrapper", "wrapper/rke2", "wrapper/rke2/kubectl"},
			expectedFormat: FormatRKE2,
		},
		{
			name:           "K3s with wrapper directory",
			setupDirs:      []string{"wrapper", "wrapper/k3s", "wrapper/k3s/kubectl"},
			expectedFormat: FormatK3s,
		},
		{
			name:           "Unknown structure",
			setupDirs:      []string{"random", "other"},
			expectedFormat: FormatUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tempDir := t.TempDir()

			// Setup directory structure
			for _, dir := range tt.setupDirs {
				path := filepath.Join(tempDir, dir)
				if err := os.MkdirAll(path, 0755); err != nil {
					t.Fatalf("failed to create dir %s: %v", path, err)
				}
			}

			format := DetectFormat(tempDir)
			if format != tt.expectedFormat {
				t.Errorf("DetectFormat() = %v, want %v", format, tt.expectedFormat)
			}
		})
	}
}

func TestRKE2PathResolver(t *testing.T) {
	resolver := NewRKE2PathResolver("/test/bundle")

	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"GetDistro", resolver.GetDistro(), "rke2"},
		{"GetKubectlDir", resolver.GetKubectlDir(), filepath.Join("/test/bundle", "rke2", "kubectl")},
		{"GetPodLogsDir", resolver.GetPodLogsDir(), filepath.Join("/test/bundle", "rke2", "podlogs")},
		{"GetPodManifestsDir", resolver.GetPodManifestsDir(), filepath.Join("/test/bundle", "rke2", "pod-manifests")},
		{"GetPodDescribeDir", resolver.GetPodDescribeDir(), filepath.Join("/test/bundle", "rke2", "kubectl", "poddescribe")},
		{"GetAgentLogsDir", resolver.GetAgentLogsDir(), filepath.Join("/test/bundle", "rke2", "agent-logs")},
		{"GetEtcdDir", resolver.GetEtcdDir(), filepath.Join("/test/bundle", "rke2", "etcd")},
		{"GetVersionFile", resolver.GetVersionFile(), filepath.Join("/test/bundle", "rke2", "version")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.expected)
			}
		})
	}

	// Test journald paths
	paths := resolver.GetJournaldPaths()
	if len(paths) != 5 {
		t.Errorf("GetJournaldPaths() returned %d paths, want 5", len(paths))
	}
}

func TestK3sPathResolver(t *testing.T) {
	resolver := NewK3sPathResolver("/test/bundle")

	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"GetDistro", resolver.GetDistro(), "k3s"},
		{"GetKubectlDir", resolver.GetKubectlDir(), filepath.Join("/test/bundle", "k3s", "kubectl")},
		{"GetPodLogsDir", resolver.GetPodLogsDir(), filepath.Join("/test/bundle", "k3s", "podlogs")},
		{"GetPodManifestsDir", resolver.GetPodManifestsDir(), filepath.Join("/test/bundle", "k3s", "pod-manifests")},
		{"GetPodDescribeDir", resolver.GetPodDescribeDir(), filepath.Join("/test/bundle", "k3s", "kubectl", "poddescribe")},
		{"GetAgentLogsDir", resolver.GetAgentLogsDir(), filepath.Join("/test/bundle", "k3s", "agent-logs")},
		{"GetEtcdDir", resolver.GetEtcdDir(), filepath.Join("/test/bundle", "k3s", "etcd")},
		{"GetVersionFile", resolver.GetVersionFile(), filepath.Join("/test/bundle", "k3s", "version")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.expected)
			}
		})
	}

	// Test journald paths
	paths := resolver.GetJournaldPaths()
	if len(paths) != 3 {
		t.Errorf("GetJournaldPaths() returned %d paths, want 3", len(paths))
	}
}

func TestNewPathResolver(t *testing.T) {
	tests := []struct {
		name         string
		format       BundleFormat
		expectedType string
	}{
		{"RKE2 format", FormatRKE2, "rke2"},
		{"K3s format", FormatK3s, "k3s"},
		{"Unknown defaults to RKE2", FormatUnknown, "rke2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewPathResolver("/test", tt.format)
			if resolver.GetDistro() != tt.expectedType {
				t.Errorf("NewPathResolver(%v).GetDistro() = %v, want %v",
					tt.format, resolver.GetDistro(), tt.expectedType)
			}
		})
	}
}

func TestParseK3sVersion(t *testing.T) {
	// Create temp directory structure
	tempDir := t.TempDir()
	bundleDir := filepath.Join(tempDir, "test-bundle")
	k3sDir := filepath.Join(bundleDir, "k3s")
	if err := os.MkdirAll(k3sDir, 0755); err != nil {
		t.Fatalf("failed to create k3s dir: %v", err)
	}

	// Write version file
	versionFile := filepath.Join(k3sDir, "version")
	if err := os.WriteFile(versionFile, []byte("v1.28.4+k3s1"), 0644); err != nil {
		t.Fatalf("failed to write version file: %v", err)
	}

	version := parseK3sVersion(bundleDir)
	if version != "v1.28.4+k3s1" {
		t.Errorf("parseK3sVersion() = %v, want %v", version, "v1.28.4+k3s1")
	}
}

func TestParseK3sVersion_Missing(t *testing.T) {
	tempDir := t.TempDir()
	version := parseK3sVersion(tempDir)
	if version != "unknown" {
		t.Errorf("parseK3sVersion() with missing file = %v, want %v", version, "unknown")
	}
}
