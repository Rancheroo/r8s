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

	// Create directory structure
	for dir, files := range structure {
		dirPath := filepath.Join(tmpDir, dir)
		os.MkdirAll(dirPath, 0755)
		for _, file := range files {
			filepath.Join(dirPath, file)
			os.WriteFile(filepath.Join(dirPath, file), []byte("test"), 0644)
		}
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func TestDetectFormat_RKE2Direct(t *testing.T) {
	// RKE2 bundle with direct rke2/ directory
	structure := map[string][]string{
		"rke2/kubectl": {"nodes", "pods"},
		"rke2/server":  {"kube-apiserver.log"},
	}
	
	bundlePath, cleanup := createManifestTestBundle(t, structure)
	defer cleanup()

	format := DetectFormat(bundlePath)
	if format != FormatRKE2 {
		t.Errorf("Expected FormatRKE2, got: %v", format)
	}
}

func TestDetectFormat_RKE2Wrapped(t *testing.T) {
	// RKE2 bundle with wrapper directory (common in tar.gz)
	tmpDir, err := os.MkdirTemp("", "r8s-manifest-wrap-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create wrapper directory with single entry
	wrapperDir := filepath.Join(tmpDir, "w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09")
	os.MkdirAll(wrapperDir, 0755)
	
	// Create rke2 inside wrapper
	os.MkdirAll(filepath.Join(wrapperDir, "rke2", "kubectl"), 0755)
	os.WriteFile(filepath.Join(wrapperDir, "rke2", "kubectl", "nodes"), []byte("test"), 0644)

	format := DetectFormat(tmpDir)
	if format != FormatRKE2 {
		t.Errorf("Expected FormatRKE2 for wrapped bundle, got: %v", format)
	}
}

func TestDetectFormat_KubectlDump(t *testing.T) {
	// kubectl cluster-info dump structure
	structure := map[string][]string{
		"namespaces/default": {"pod.yaml"},
		"namespaces/kube-system": {"pod.yaml"},
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
		"random": {"file.txt"},
	}
	
	bundlePath, cleanup := createManifestTestBundle(t, structure)
	defer cleanup()

	format := DetectFormat(bundlePath)
	if format != FormatUnknown {
		t.Errorf("Expected FormatUnknown, got: %v", format)
	}
}

func TestDetectFormat_NonExistent(t *testing.T) {
	format := DetectFormat("/nonexistent/path/that/does/not/exist")
	if format != FormatUnknown {
		t.Errorf("Expected FormatUnknown for non-existent path, got: %v", format)
	}
}

func TestParseManifest_RKE2(t *testing.T) {
	structure := map[string][]string{
		"rke2/kubectl": {"nodes", "pods", "services"},
		"rke2/server":  {"kube-apiserver.log"},
		"systemlogs":   {"syslog"},
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
		t.Error("Expected FileCount > 0")
	}
	
	if manifest.TotalSize == 0 {
		t.Error("Expected TotalSize > 0")
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

func TestParseManifest_WithVersion(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "r8s-manifest-ver-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create rke2 structure
	os.MkdirAll(filepath.Join(tmpDir, "rke2", "kubectl"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "rke2", "kubectl", "nodes"), []byte("test"), 0644)
	
	// Add version file
	os.WriteFile(filepath.Join(tmpDir, "rke2", "version"), []byte("v1.28.5+rke2r1"), 0644)

	manifest, err := ParseManifest(tmpDir)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if manifest.RKE2Version != "v1.28.5+rke2r1" {
		t.Errorf("Expected RKE2 version v1.28.5+rke2r1, got: %s", manifest.RKE2Version)
	}
}

func TestExtractNodeName_FromDirectory(t *testing.T) {
	// Create bundle with node name in directory
	tmpDir, err := os.MkdirTemp("", "r8s-manifest-node-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create wrapper with node name pattern: nodename-timestamp
	wrapperDir := filepath.Join(tmpDir, "w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09")
	os.MkdirAll(wrapperDir, 0755)
	os.MkdirAll(filepath.Join(wrapperDir, "rke2", "kubectl"), 0755)
	os.WriteFile(filepath.Join(wrapperDir, "rke2", "kubectl", "nodes"), []byte("test"), 0644)

	manifest, err := ParseManifest(tmpDir)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	expectedNodeName := "w-guard-wg-cp-svtk6-lqtxw"
	if manifest.NodeName != expectedNodeName {
		t.Errorf("Expected node name %s, got: %s", expectedNodeName, manifest.NodeName)
	}
}

func TestExtractNodeName_FromHostname(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "r8s-manifest-host-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create simple structure
	os.MkdirAll(filepath.Join(tmpDir, "rke2", "kubectl"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "rke2", "kubectl", "nodes"), []byte("test"), 0644)
	
	// Add hostname file
	os.MkdirAll(filepath.Join(tmpDir, "systeminfo"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "systeminfo", "hostname"), []byte("my-server-01\n"), 0644)

	manifest, err := ParseManifest(tmpDir)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if manifest.NodeName != "my-server-01" {
		t.Errorf("Expected node name my-server-01, got: %s", manifest.NodeName)
	}
}

func TestCalculateBundleStats(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "r8s-manifest-stats-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create some files with known sizes
	os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("hello"), 0644) // 5 bytes
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("world!"), 0644) // 6 bytes
	os.WriteFile(filepath.Join(tmpDir, "subdir", "file3.txt"), []byte("test"), 0644) // 4 bytes

	count, size, err := calculateBundleStats(tmpDir)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected 3 files, got: %d", count)
	}

	if size != 15 {
		t.Errorf("Expected 15 bytes total, got: %d", size)
	}
}

func TestParseRKE2Version(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "r8s-manifest-rke2ver-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create rke2 structure
	os.MkdirAll(filepath.Join(tmpDir, "rke2"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "rke2", "version"), []byte("v1.28.5+rke2r1\n"), 0644)

	version := parseRKE2Version(tmpDir)
	if version != "v1.28.5+rke2r1" {
		t.Errorf("Expected version v1.28.5+rke2r1, got: %s", version)
	}
}

func TestParseRKE2Version_Missing(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "r8s-manifest-nover-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	version := parseRKE2Version(tmpDir)
	if version != "unknown" {
		t.Errorf("Expected 'unknown' for missing version, got: %s", version)
	}
}

func TestParseK8sVersion(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "r8s-manifest-k8sver-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create rke2 structure
	os.MkdirAll(filepath.Join(tmpDir, "rke2", "kubectl"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "rke2", "kubectl", "version"), []byte("v1.28.5"), 0644)

	version := parseK8sVersion(tmpDir)
	if version != "v1.28.5" {
		t.Errorf("Expected version v1.28.5, got: %s", version)
	}
}

func TestParseK8sVersion_JSONFormat(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "r8s-manifest-k8sjson-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create rke2 structure with JSON version output
	os.MkdirAll(filepath.Join(tmpDir, "rke2", "kubectl"), 0755)
	jsonVersion := `{
  "clientVersion": {
    "major": "1",
    "minor": "28",
    "gitVersion": "v1.28.5",
    "gitCommit": "abc123"
  }
}`
	os.WriteFile(filepath.Join(tmpDir, "rke2", "kubectl", "version"), []byte(jsonVersion), 0644)

	version := parseK8sVersion(tmpDir)
	if version != "v1.28.5" {
		t.Errorf("Expected version v1.28.5 from JSON, got: %s", version)
	}
}

func TestParseK8sVersion_Missing(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "r8s-manifest-nok8s-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	version := parseK8sVersion(tmpDir)
	if version != "unknown" {
		t.Errorf("Expected 'unknown' for missing version, got: %s", version)
	}
}
