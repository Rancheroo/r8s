package bundle

import (
	"testing"
	"time"
)

func TestBundleFormatConstants(t *testing.T) {
	tests := []struct {
		format   BundleFormat
		expected string
	}{
		{FormatRKE2, "rke2-support-bundle"},
		{FormatK3s, "k3s-support-bundle"},
		{FormatKubectl, "kubectl-cluster-info"},
		{FormatUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if string(tt.format) != tt.expected {
				t.Errorf("BundleFormat = %v, want %v", tt.format, tt.expected)
			}
		})
	}
}

func TestBundleStruct(t *testing.T) {
	// Test Bundle struct initialization
	bundle := &Bundle{
		Path:        "/test/bundle",
		ExtractPath: "/tmp/extracted",
		Loaded:      true,
		Size:        1024,
		IsTemporary: false,
	}

	if bundle.Path != "/test/bundle" {
		t.Errorf("Bundle.Path = %v, want /test/bundle", bundle.Path)
	}

	if !bundle.Loaded {
		t.Error("Bundle.Loaded should be true")
	}

	if bundle.Size != 1024 {
		t.Errorf("Bundle.Size = %v, want 1024", bundle.Size)
	}
}

func TestBundleManifest(t *testing.T) {
	now := time.Now()
	manifest := &BundleManifest{
		NodeName:    "test-node",
		CollectedAt: now,
		RKE2Version: "v1.28.0+rke2r1",
		K8sVersion:  "v1.28.0",
		FileCount:   100,
		TotalSize:   1048576,
		BundleType:  "rke2-support-bundle",
	}

	if manifest.NodeName != "test-node" {
		t.Errorf("BundleManifest.NodeName = %v, want test-node", manifest.NodeName)
	}

	if manifest.RKE2Version != "v1.28.0+rke2r1" {
		t.Errorf("BundleManifest.RKE2Version = %v", manifest.RKE2Version)
	}

	if manifest.FileCount != 100 {
		t.Errorf("BundleManifest.FileCount = %v, want 100", manifest.FileCount)
	}
}

func TestPodInfo(t *testing.T) {
	pod := PodInfo{
		Namespace:       "default",
		Name:            "test-pod",
		Containers:      []string{"container1", "container2"},
		HasCurrentLogs:  true,
		HasPreviousLogs: false,
	}

	if pod.Namespace != "default" {
		t.Errorf("PodInfo.Namespace = %v, want default", pod.Namespace)
	}

	if len(pod.Containers) != 2 {
		t.Errorf("PodInfo.Containers length = %v, want 2", len(pod.Containers))
	}

	if !pod.HasCurrentLogs {
		t.Error("PodInfo.HasCurrentLogs should be true")
	}
}

func TestLogFileInfo(t *testing.T) {
	logFile := LogFileInfo{
		Path:          "rke2/logs/kube-system/kube-proxy.log",
		Type:          LogTypePod,
		Namespace:     "kube-system",
		PodName:       "kube-proxy",
		ContainerName: "kube-proxy",
		IsPrevious:    false,
		Size:          4096,
		LineCount:     100,
	}

	if logFile.Namespace != "kube-system" {
		t.Errorf("LogFileInfo.Namespace = %v, want kube-system", logFile.Namespace)
	}

	if logFile.Size != 4096 {
		t.Errorf("LogFileInfo.Size = %v, want 4096", logFile.Size)
	}

	if logFile.Type != LogTypePod {
		t.Errorf("LogFileInfo.Type = %v, want LogTypePod", logFile.Type)
	}
}

func TestLogTypeConstants(t *testing.T) {
	if LogTypePod != "pod" {
		t.Errorf("LogTypePod = %v, want pod", LogTypePod)
	}

	if LogTypeSystem != "system" {
		t.Errorf("LogTypeSystem = %v, want system", LogTypeSystem)
	}

	if LogTypeJournald != "journald" {
		t.Errorf("LogTypeJournald = %v, want journald", LogTypeJournald)
	}
}

func TestBundleHealth(t *testing.T) {
	health := &BundleHealth{
		TotalFiles:   100,
		FoundFiles:   90,
		DerivedFiles: 5,
		MissingFiles: []string{"optional.txt"},
		Warnings:     []string{"Warning: some logs truncated"},
	}

	if health.TotalFiles != 100 {
		t.Errorf("BundleHealth.TotalFiles = %v, want 100", health.TotalFiles)
	}

	if health.FoundFiles != 90 {
		t.Errorf("BundleHealth.FoundFiles = %v, want 90", health.FoundFiles)
	}

	if len(health.MissingFiles) != 1 {
		t.Errorf("BundleHealth.MissingFiles length = %v, want 1", len(health.MissingFiles))
	}
}

func TestBundleHealthPercentage(t *testing.T) {
	tests := []struct {
		name     string
		health   BundleHealth
		expected int
	}{
		{
			name:     "100% - all found",
			health:   BundleHealth{TotalFiles: 100, FoundFiles: 100, DerivedFiles: 0},
			expected: 100,
		},
		{
			name:     "90% - some missing",
			health:   BundleHealth{TotalFiles: 100, FoundFiles: 90, DerivedFiles: 0},
			expected: 90,
		},
		{
			name:     "95% - with derived",
			health:   BundleHealth{TotalFiles: 100, FoundFiles: 90, DerivedFiles: 5},
			expected: 95,
		},
		{
			name:     "0% - empty",
			health:   BundleHealth{TotalFiles: 0, FoundFiles: 0, DerivedFiles: 0},
			expected: 100, // Edge case: no files = 100%
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pct := tt.health.Percentage()
			if pct != tt.expected {
				t.Errorf("Percentage() = %v, want %v", pct, tt.expected)
			}
		})
	}
}

func TestBundleHealthColor(t *testing.T) {
	tests := []struct {
		name     string
		health   BundleHealth
		expected string
	}{
		{
			name:     "green - excellent",
			health:   BundleHealth{TotalFiles: 100, FoundFiles: 95},
			expected: "green",
		},
		{
			name:     "yellow - good",
			health:   BundleHealth{TotalFiles: 100, FoundFiles: 75},
			expected: "yellow",
		},
		{
			name:     "red - poor",
			health:   BundleHealth{TotalFiles: 100, FoundFiles: 50},
			expected: "red",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			color := tt.health.Color()
			if color != tt.expected {
				t.Errorf("Color() = %v, want %v", color, tt.expected)
			}
		})
	}
}

func TestImportOptions(t *testing.T) {
	// Test ImportOptions struct
	opts := ImportOptions{
		Path:          "/test/bundle",
		MaxSize:       50 * 1024 * 1024,
		KeepExtracted: true,
		ExtractTo:     "/custom/extract",
		Verbose:       true,
	}

	if !opts.Verbose {
		t.Error("ImportOptions.Verbose should be true")
	}

	if !opts.KeepExtracted {
		t.Error("ImportOptions.KeepExtracted should be true")
	}

	if opts.MaxSize != 50*1024*1024 {
		t.Errorf("ImportOptions.MaxSize = %v, want 50MB", opts.MaxSize)
	}
}

func TestDefaultMaxBundleSize(t *testing.T) {
	expected := int64(50 * 1024 * 1024) // 50MB
	if DefaultMaxBundleSize != expected {
		t.Errorf("DefaultMaxBundleSize = %v, want %v", DefaultMaxBundleSize, expected)
	}
}
