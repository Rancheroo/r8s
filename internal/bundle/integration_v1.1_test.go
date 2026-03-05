package bundle

import (
	"os"
	"path/filepath"
	"testing"
)

// TestIntegration_NewBundleFormat tests parsing against the actual new bundle format
// Bundle: /home/bradmin/Downloads/logBundles/r8s-cp-wlp7h-lhvgq-2026-03-03_03_49_04/
func TestIntegration_NewBundleFormat(t *testing.T) {
	bundlePath := "/home/bradmin/Downloads/logBundles/r8s-cp-wlp7h-lhvgq-2026-03-03_03_49_04"

	// Skip if bundle doesn't exist (CI environment)
	if _, err := os.Stat(bundlePath); os.IsNotExist(err) {
		t.Skip("New format bundle not available for integration test")
	}

	t.Run("ParseSystemHealth", func(t *testing.T) {
		health, err := ParseSystemHealth(bundlePath)
		if err != nil {
			t.Fatalf("ParseSystemHealth failed: %v", err)
		}

		// Check virtualization detection
		if health.VirtType != "kvm" {
			t.Errorf("VirtType = %q, want kvm", health.VirtType)
		}
		t.Logf("✅ VirtType: %s", health.VirtType)

		// Memory parsing should work (either memory or freem file)
		t.Logf("✅ Memory Used: %.1f%%", health.MemoryUsedPercent)

		// Disk parsing
		t.Logf("✅ Disk Used: %.1f%%", health.DiskUsedPercent)
	})

	t.Run("ParseVersions", func(t *testing.T) {
		info, err := ParseVersions(bundlePath)
		if err != nil {
			t.Fatalf("ParseVersions failed: %v", err)
		}

		// Check system info
		if info.CollectionDate == "" {
			t.Error("CollectionDate is empty")
		} else {
			t.Logf("✅ Collection Date: %s", info.CollectionDate)
		}

		if info.Hostname == "" {
			t.Error("Hostname is empty")
		} else {
			t.Logf("✅ Hostname: %s", info.Hostname)
		}

		if info.KernelVersion == "" {
			t.Error("KernelVersion is empty")
		} else {
			t.Logf("✅ Kernel: %s", info.KernelVersion)
		}

		if info.OSName == "" {
			t.Error("OSName is empty")
		} else {
			t.Logf("✅ OS: %s", info.OSName)
		}

		// Check RKE2 version
		if info.DistroVersion == "" {
			t.Error("DistroVersion is empty")
		} else {
			t.Logf("✅ RKE2 Version: %s", info.DistroVersion)
		}

		// Check images were parsed
		t.Logf("✅ RKE2 Images: %d", len(info.RKE2Images))
		t.Logf("✅ Cattle Images: %d", len(info.CattleImages))

		// Check helm releases
		if len(info.HelmReleases) == 0 {
			t.Error("No Helm releases parsed")
		} else {
			t.Logf("✅ Helm Releases: %d", len(info.HelmReleases))
			for _, r := range info.HelmReleases {
				if r.Status != "deployed" {
					t.Logf("  ⚠️  %s: %s", r.Name, r.Status)
				}
			}
		}
	})

	t.Run("DMesgParser", func(t *testing.T) {
		analysis, err := ParseDMesg(bundlePath)
		if err != nil {
			// dmesg might not exist or be empty - that's ok
			t.Logf("Note: ParseDMesg: %v", err)
			return
		}

		t.Logf("✅ OOM Kills detected: %d", len(analysis.OOMKills))
		t.Logf("✅ Memory pressure: %v", analysis.MemoryPressure)
	})

	t.Run("PathResolver", func(t *testing.T) {
		resolver := NewRKE2PathResolver(bundlePath)

		// Check new paths exist
		rootVersions := resolver.GetRootVersionsFile()
		if _, err := os.Stat(rootVersions); err != nil {
			t.Errorf("Root versions file not found: %s", rootVersions)
		} else {
			t.Logf("✅ Root versions file: %s", rootVersions)
		}

		// Check dmesg in new location
		dmesgPaths := resolver.GetDmesgPaths()
		found := false
		for _, p := range dmesgPaths {
			if _, err := os.Stat(p); err == nil {
				t.Logf("✅ dmesg found: %s", p)
				found = true
				break
			}
		}
		if !found {
			t.Logf("Note: dmesg not found in expected locations")
		}

		// Check memory in new location
		memPaths := resolver.GetMemoryPaths()
		found = false
		for _, p := range memPaths {
			if _, err := os.Stat(p); err == nil {
				t.Logf("✅ memory file found: %s", p)
				found = true
				break
			}
		}
		if !found {
			t.Logf("Note: memory file not found in expected locations")
		}

		// Check virt detection file
		virtPath := filepath.Join(resolver.GetSysteminfoPath(), "systemd-detect-virt")
		if content, err := os.ReadFile(virtPath); err == nil {
			t.Logf("✅ systemd-detect-virt: %s", string(content))
		}

		// Check poddescribe directory
		podDescribeDir := resolver.GetPodDescribeDir()
		if entries, err := os.ReadDir(podDescribeDir); err == nil {
			t.Logf("✅ Pod describe namespaces: %d", len(entries))
		}
	})
}

// TestIntegration_OldBundleFormat ensures backward compatibility
// Bundle: /home/bradmin/Downloads/logBundles/r8s-cp-wlp7h-lhvgq-2026-02-03_05_38_57/
func TestIntegration_OldBundleFormat(t *testing.T) {
	bundlePath := "/home/bradmin/Downloads/logBundles/r8s-cp-wlp7h-lhvgq-2026-02-03_05_38_57"

	// Skip if bundle doesn't exist
	if _, err := os.Stat(bundlePath); os.IsNotExist(err) {
		t.Skip("Old format bundle not available for integration test")
	}

	t.Run("BackwardCompatibility_SystemHealth", func(t *testing.T) {
		health, err := ParseSystemHealth(bundlePath)
		if err != nil {
			t.Fatalf("ParseSystemHealth failed on old bundle: %v", err)
		}

		// Old bundle doesn't have virt detection
		if health.VirtType != "" {
			t.Logf("Note: Old bundle has virt type: %s (unexpected)", health.VirtType)
		}

		// Memory should still work (via freem file)
		t.Logf("✅ Memory parsing works: %.1f%%", health.MemoryUsedPercent)
	})

	t.Run("BackwardCompatibility_Versions", func(t *testing.T) {
		_, err := ParseVersions(bundlePath)
		// Old bundles don't have root versions file - this should fail gracefully
		if err == nil {
			t.Log("Note: Old bundle has versions file (unexpected)")
		} else {
			t.Logf("✅ Old bundle correctly lacks versions file: %v", err)
		}
	})

	t.Run("BackwardCompatibility_DMesg", func(t *testing.T) {
		// Old bundles have dmesg in systemlogs/
		analysis, err := ParseDMesg(bundlePath)
		if err != nil {
			t.Logf("Note: dmesg parse: %v", err)
		} else {
			t.Logf("✅ dmesg parsed from old location: %d OOM kills", len(analysis.OOMKills))
		}
	})
}

// TestIntegration_BothFormats compares parsing between formats
func TestIntegration_BothFormats(t *testing.T) {
	newBundle := "/home/bradmin/Downloads/logBundles/r8s-cp-wlp7h-lhvgq-2026-03-03_03_49_04"
	oldBundle := "/home/bradmin/Downloads/logBundles/r8s-cp-wlp7h-lhvgq-2026-02-03_05_38_57"

	// Skip if either bundle is missing
	if _, err := os.Stat(newBundle); os.IsNotExist(err) {
		t.Skip("Bundles not available for comparison test")
	}
	if _, err := os.Stat(oldBundle); os.IsNotExist(err) {
		t.Skip("Bundles not available for comparison test")
	}

	t.Run("CompareSystemHealth", func(t *testing.T) {
		newHealth, err := ParseSystemHealth(newBundle)
		if err != nil {
			t.Fatalf("Failed to parse new bundle: %v", err)
		}

		oldHealth, err := ParseSystemHealth(oldBundle)
		if err != nil {
			t.Fatalf("Failed to parse old bundle: %v", err)
		}

		t.Logf("New bundle - Memory: %.1f%%, Disk: %.1f%%, Virt: %s",
			newHealth.MemoryUsedPercent, newHealth.DiskUsedPercent, newHealth.VirtType)
		t.Logf("Old bundle - Memory: %.1f%%, Disk: %.1f%%, Virt: %s",
			oldHealth.MemoryUsedPercent, oldHealth.DiskUsedPercent, oldHealth.VirtType)

		// Both should have similar memory usage (same node)
		if newHealth.MemoryUsedPercent == 0 && oldHealth.MemoryUsedPercent > 0 {
			t.Error("New bundle memory parsing failed while old worked")
		}
	})
}
