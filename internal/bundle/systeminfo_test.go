package bundle

import (
	"os"
	"path/filepath"
	"testing"
)

func createSystemInfoTestBundle(t *testing.T, freemContent, dfhContent, virtContent string) (string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "r8s-systeminfo-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	systeminfoDir := filepath.Join(tmpDir, "systeminfo")
	os.MkdirAll(systeminfoDir, 0755)

	if freemContent != "" {
		os.WriteFile(filepath.Join(systeminfoDir, "freem"), []byte(freemContent), 0644)
	}
	if dfhContent != "" {
		os.WriteFile(filepath.Join(systeminfoDir, "dfh"), []byte(dfhContent), 0644)
	}
	if virtContent != "" {
		os.WriteFile(filepath.Join(systeminfoDir, "systemd-detect-virt"), []byte(virtContent), 0644)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func TestParseSystemHealth_BothFiles(t *testing.T) {
	freem := `              total        used        free      shared  buff/cache   available
Mem:       16384       8192        4096        1024        4096        6144
Swap:       2048          0        2048`

	dfh := `Filesystem      Size  Used Avail Use% Mounted on
/dev/sda1       100G   50G   50G  50% /
/dev/sdb1       200G  100G  100G  50% /data`

	bundlePath, cleanup := createSystemInfoTestBundle(t, freem, dfh, "")
	defer cleanup()

	health, err := ParseSystemHealth(bundlePath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Memory: 8192/16384 = 50%
	if health.MemoryUsedPercent != 50.0 {
		t.Errorf("Expected memory used 50%%, got: %f", health.MemoryUsedPercent)
	}

	// Disk: 50%
	if health.DiskUsedPercent != 50.0 {
		t.Errorf("Expected disk used 50%%, got: %f", health.DiskUsedPercent)
	}
}

func TestParseSystemHealth_OnlyMemory(t *testing.T) {
	freem := `              total        used        free      shared  buff/cache   available
Mem:       8192        2048        6144        0        0        6144`

	bundlePath, cleanup := createSystemInfoTestBundle(t, freem, "", "")
	defer cleanup()

	health, err := ParseSystemHealth(bundlePath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Memory: 2048/8192 = 25%
	if health.MemoryUsedPercent != 25.0 {
		t.Errorf("Expected memory used 25%%, got: %f", health.MemoryUsedPercent)
	}

	// Disk should be 0 (file missing)
	if health.DiskUsedPercent != 0 {
		t.Errorf("Expected disk used 0%%, got: %f", health.DiskUsedPercent)
	}
}

func TestParseSystemHealth_OnlyDisk(t *testing.T) {
	dfh := `Filesystem      Size  Used Avail Use% Mounted on
/dev/sda1       500G  400G  100G  80% /`

	bundlePath, cleanup := createSystemInfoTestBundle(t, "", dfh, "")
	defer cleanup()

	health, err := ParseSystemHealth(bundlePath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Memory should be 0 (file missing)
	if health.MemoryUsedPercent != 0 {
		t.Errorf("Expected memory used 0%%, got: %f", health.MemoryUsedPercent)
	}

	// Disk: 80%
	if health.DiskUsedPercent != 80.0 {
		t.Errorf("Expected disk used 80%%, got: %f", health.DiskUsedPercent)
	}
}

func TestParseSystemHealth_MissingFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "r8s-systeminfo-missing-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	os.MkdirAll(filepath.Join(tmpDir, "systeminfo"), 0755)

	health, err := ParseSystemHealth(tmpDir)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Both should be 0 when files missing
	if health.MemoryUsedPercent != 0 {
		t.Errorf("Expected memory 0%% when file missing, got: %f", health.MemoryUsedPercent)
	}
	if health.DiskUsedPercent != 0 {
		t.Errorf("Expected disk 0%% when file missing, got: %f", health.DiskUsedPercent)
	}
}

func TestParseSystemHealth_MalformedFiles(t *testing.T) {
	freem := `this is not valid freem output`

	dfh := `neither is this`

	bundlePath, cleanup := createSystemInfoTestBundle(t, freem, dfh, "")
	defer cleanup()

	// Should not error, just return 0s for malformed data
	health, err := ParseSystemHealth(bundlePath)
	if err != nil {
		t.Fatalf("Expected no error for malformed files, got: %v", err)
	}

	// Values should be 0 (parsing failed gracefully)
	if health.MemoryUsedPercent != 0 {
		t.Errorf("Expected memory 0%% for malformed data, got: %f", health.MemoryUsedPercent)
	}
}

func TestParseSystemHealth_DetectVirt_KVM(t *testing.T) {
	virt := "kvm\n"

	bundlePath, cleanup := createSystemInfoTestBundle(t, "", "", virt)
	defer cleanup()

	health, err := ParseSystemHealth(bundlePath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if health.VirtType != "kvm" {
		t.Errorf("Expected VirtType 'kvm', got: %s", health.VirtType)
	}
}

func TestParseSystemHealth_DetectVirt_Docker(t *testing.T) {
	virt := "docker\n"

	bundlePath, cleanup := createSystemInfoTestBundle(t, "", "", virt)
	defer cleanup()

	health, err := ParseSystemHealth(bundlePath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if health.VirtType != "docker" {
		t.Errorf("Expected VirtType 'docker', got: %s", health.VirtType)
	}
}

func TestParseSystemHealth_DetectVirt_None(t *testing.T) {
	virt := "none\n"

	bundlePath, cleanup := createSystemInfoTestBundle(t, "", "", virt)
	defer cleanup()

	health, err := ParseSystemHealth(bundlePath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if health.VirtType != "none" {
		t.Errorf("Expected VirtType 'none', got: %s", health.VirtType)
	}
}

func TestParseSystemHealth_DetectVirt_MissingFile(t *testing.T) {
	bundlePath, cleanup := createSystemInfoTestBundle(t, "", "", "")
	defer cleanup()

	health, err := ParseSystemHealth(bundlePath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// VirtType should be empty when file missing
	if health.VirtType != "" {
		t.Errorf("Expected VirtType empty when file missing, got: %s", health.VirtType)
	}
}

// TestParseSystemHealth_NewMemoryFormat tests the new v1.1+ memory file with Gi units
func TestParseSystemHealth_NewMemoryFormat(t *testing.T) {
	// New format uses "memory" file with Gi units
	memory := `              total        used        free      shared  buff/cache   available
Mem:           3.8Gi       2.0Gi       183Mi       6.4Mi       2.0Gi       1.8Gi
Swap:             0B          0B          0B`

	tmpDir, err := os.MkdirTemp("", "r8s-systeminfo-new-format-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	systeminfoDir := filepath.Join(tmpDir, "systeminfo")
	os.MkdirAll(systeminfoDir, 0755)
	os.WriteFile(filepath.Join(systeminfoDir, "memory"), []byte(memory), 0644)

	health, err := ParseSystemHealth(tmpDir)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Memory: 2.0Gi / 3.8Gi = ~52.6%
	expected := 52.63157894736842
	if health.MemoryUsedPercent != expected {
		t.Errorf("Expected memory used %.2f%%, got: %.2f%%", expected, health.MemoryUsedPercent)
	}
}

// TestParseSystemHealth_MemoryFileFallback tests that it tries "memory" first, then "freem"
func TestParseSystemHealth_MemoryFileFallback(t *testing.T) {
	// Create bundle with only "freem" file (old format)
	freem := `              total        used        free      shared  buff/cache   available
Mem:       16384        8192        4096        1024        4096        6144`

	tmpDir, err := os.MkdirTemp("", "r8s-systeminfo-fallback-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	systeminfoDir := filepath.Join(tmpDir, "systeminfo")
	os.MkdirAll(systeminfoDir, 0755)
	os.WriteFile(filepath.Join(systeminfoDir, "freem"), []byte(freem), 0644)

	health, err := ParseSystemHealth(tmpDir)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Memory: 8192/16384 = 50%
	if health.MemoryUsedPercent != 50.0 {
		t.Errorf("Expected memory used 50%%, got: %f%%", health.MemoryUsedPercent)
	}
}

// TestParseMemoryValue tests the memory value parser with various units
func TestParseMemoryValue(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
		epsilon  float64 // Allow for floating point comparison
	}{
		// Gi units
		{"3.8Gi", 3.8, 0.001},
		{"2.0Gi", 2.0, 0.001},
		{"10Gi", 10.0, 0.001},
		// Mi units
		{"183Mi", 183.0 / 1024.0, 0.001},
		{"1024Mi", 1.0, 0.001},
		{"6.4Mi", 6.4 / 1024.0, 0.001},
		// Ki units
		{"1048576Ki", 1.0, 0.001},
		// B units
		{"0B", 0.0, 0.001},
		{"1024B", 1024.0 / (1024.0 * 1024.0 * 1024.0), 1e-10}, // Small value, small epsilon
		// Raw numbers (no units)
		{"16384", 16384.0, 0.001},
		{"8192", 8192.0, 0.001},
		{"0", 0.0, 0.001},
		// With whitespace
		{"  3.8Gi  ", 3.8, 0.001},
		{"  8192", 8192.0, 0.001},
		// Invalid
		{"invalid", 0.0, 0.001},
		{"", 0.0, 0.001},
	}

	for _, tt := range tests {
		result := parseMemoryValue(tt.input)
		diff := result - tt.expected
		if diff < 0 {
			diff = -diff
		}
		if diff > tt.epsilon {
			t.Errorf("parseMemoryValue(%q) = %v, want %v (diff %v > epsilon %v)", 
				tt.input, result, tt.expected, diff, tt.epsilon)
		}
	}
}
