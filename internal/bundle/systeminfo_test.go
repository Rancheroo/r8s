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
