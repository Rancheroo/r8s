package bundle

import (
	"os"
	"path/filepath"
	"testing"
)

// createManifestTestBundle creates a minimal bundle structure for testing
func createManifestTestBundle(t *testing.T, structure map[string][]string) (string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "r8s-manifest-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create structure
	for dir, files := range structure {
		dirPath := filepath.Join(tmpDir, dir)
		os.MkdirAll(dirPath, 0755)
		for _, file := range files {
			if file != "" {
				os.WriteFile(filepath.Join(dirPath, file), []byte("test"), 0644)
			}
		}
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func TestDetectFormat_RKE2Direct(t *testing.T) {
	// Direct rke2/ structure
	structure := map[string][]string{
		"rke2/kubectl": {"nodes", "pods"},
	}
	
	bundlePath, cleanup := createManifestTestBundle(t, structure)
	defer cleanup()

	format := DetectFormat(bundlePath)
	if format != FormatRKE2 {
		t.Errorf("Expected FormatRKE2, got: %v", format)
	}
}

func TestDetectFormat_RKE2Wrapped(t *testing.T) {
	// Wrapped structure: wrapper-dir/rke2/
	structure := map[string][]string{
		"node-name-2025-01-15/rke2/kubectl": {"nodes"},
	}
	
	bundlePath, cleanup := createManifestTestBundle(t, structure)
	defer cleanup()

	format := DetectFormat(bundlePath)
	if format != FormatRKE2 {
		t.Errorf("Expected FormatRKE2 for wrapped bundle, got: %v", format)
	}
}

func TestDetectFormat_KubectlDump(t *testing.T) {
	// kubectl cluster-info dump structure
	structure := map[string][]string{
		"namespaces/kube-system": {"pod1.yaml"},
	}
	
	bundlePath, cleanup := createManifestTestBundle(t, structure)
	defer cleanup()

	format := DetectFormat(bundlePath)
	if format != FormatKubectl {
		t.Errorf("Expected FormatKubectl, got: %v", format)
	}
}

func TestDetectFormat_Unknown(t *testing.T) {
	// Empty or unrecognized structure
	structure := map[string][]string{
		"random-dir": {"file.txt"},
	}
	
	bundlePath, cleanup := createManifestTestBundle(t, structure)
	defer cleanup()

	format := DetectFormat(bundlePath)
	if format != FormatUnknown {
		t.Errorf("Expected FormatUnknown, got: %v", format)
	}
}

func TestDetectFormat_EmptyDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "r8s-empty-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	format := DetectFormat(tmpDir)
	if format != FormatUnknown {
		t.Errorf("Expected FormatUnknown for empty dir, got: %v", format)
	}
}

func TestParseManifest_RKE2Bundle(t *testing.T) {
	structure := map[string][]string{
		"rke2/kubectl":              {"nodes", "pods", "version"},
		"rke2":                      {"server", "agent"},
		"systemlogs":                {"syslog"},
	}
	
	bundlePath, cleanup := createManifestTestBundle(t, structure)
	defer cleanup()

	manifest, err := ParseManifest(bundlePath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if manifest.BundleType != string(FormatRKE2) {
		t.Errorf("Expected bundle type %s, got: %s", FormatRKE2, manifest.BundleType)
	}

	if manifest.FileCount == 0 {
		t.Error("Expected non-zero file count")
	}

	if manifest.TotalSize == 0 {
		t.Error("Expected non-zero total size")
	}
}

func TestParseManifest_UnknownFormat(t *testing.T) {
	structure := map[string][]string{
		"random": {"file.txt"},
	}
	
	bundlePath, cleanup := createManifestTestBundle(t, structure)
	defer cleanup()

	_, err := ParseManifest(bundlePath)
	if err == nil {
		t.Error("Expected error for unknown format")
	}
}

func TestExtractNodeName_FromTimestampedDir(t *testing.T) {
	// Create wrapped bundle with timestamped name
	tmpDir, err := os.MkdirTemp("", "r8s-node-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create: wrapper-dir/rke2 structure
	wrapperDir := filepath.Join(tmpDir, "w-guard-wg-cp-svtk6-lqtxw-2025-01-15_10_30_00")
	os.MkdirAll(filepath.Join(wrapperDir, "rke2", "kubectl"), 0755)

	nodeName := extractNodeName(tmpDir)
	expected := "w-guard-wg-cp-svtk6-lqtxw"
	if nodeName != expected {
		t.Errorf("Expected node name %s, got: %s", expected, nodeName)
	}
}

func TestExtractNodeName_FromHostnameFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "r8s-hostname-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create rke2 structure
	os.MkdirAll(filepath.Join(tmpDir, "rke2"), 0755)
	
	// Create systeminfo/hostname file
	systeminfoDir := filepath.Join(tmpDir, "systeminfo")
	os.MkdirAll(systeminfoDir, 0755)
	os.WriteFile(filepath.Join(systeminfoDir, "hostname"), []byte("my-custom-node\n"), 0644)

	nodeName := extractNodeName(tmpDir)
	if nodeName != "my-custom-node" {
		t.Errorf("Expected node name from hostname file, got: %s", nodeName)
	}
}

func TestExtractNodeName_FallbackToBaseName(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "r8s-fallback-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create simple rke2 structure without hostname file
	os.MkdirAll(filepath.Join(tmpDir, "rke2"), 0755)

	nodeName := extractNodeName(tmpDir)
	// Should fallback to directory base name
	if nodeName == "" {
		t.Error("Expected non-empty node name from fallback")
	}
}

func TestCalculateBundleStats(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "r8s-stats-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content1"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("content2-more-data"), 0644)
	os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "subdir", "file3.txt"), []byte("x"), 0644)

	count, size, err := calculateBundleStats(tmpDir)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected 3 files, got: %d", count)
	}

	// Content sizes: 8 + 18 + 1 = 27 bytes
	if size != 27 {
		t.Errorf("Expected size 27, got: %d", size)
	}
}

func TestParseRKE2Version(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "r8s-version-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create rke2 structure with version file
	os.MkdirAll(filepath.Join(tmpDir, "rke2"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "rke2", "version"), []byte("v1.28.5+rke2r1\n"), 0644)

	version := parseRKE2Version(tmpDir)
	if version != "v1.28.5+rke2r1" {
		t.Errorf("Expected version v1.28.5+rke2r1, got: %s", version)
	}
}

func TestParseRKE2Version_Missing(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "r8s-version-missing-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	os.MkdirAll(filepath.Join(tmpDir, "rke2"), 0755)

	version := parseRKE2Version(tmpDir)
	if version != "unknown" {
		t.Errorf("Expected 'unknown' for missing version file, got: %s", version)
	}
}

func TestParseK8sVersion(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "r8s-k8s-version-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create rke2/kubectl structure with version file
	os.MkdirAll(filepath.Join(tmpDir, "rke2", "kubectl"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "rke2", "kubectl", "version"), 
		[]byte(`{
  "clientVersion": {
    "major": "1",
    "minor": "28",
    "gitVersion": "v1.28.5",
    "gitCommit": "..."
  }
}`), 0644)

	version := parseK8sVersion(tmpDir)
	// Should extract GitVersion from JSON
	if version == "" || version == "unknown" {
		t.Errorf("Expected to parse k8s version from JSON, got: %s", version)
	}
}

func TestParseK8sVersion_PlainText(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "r8s-k8s-plain-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	os.MkdirAll(filepath.Join(tmpDir, "rke2", "kubectl"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "rke2", "kubectl", "version"), 
		[]byte("Client Version: v1.28.5\nServer Version: v1.28.5+rke2r1"), 0644)

	version := parseK8sVersion(tmpDir)
	// Should return raw content if not JSON
	if version == "" {
		t.Error("Expected non-empty version from plain text")
	}
}
