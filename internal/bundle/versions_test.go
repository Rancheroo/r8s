package bundle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseVersions_Success(t *testing.T) {
	// Create temporary test file
	tmpDir := t.TempDir()
	versionsContent := `Tue Mar  3 03:49:05 UTC 2026
               total        used        free      shared  buff/cache   available
Mem:           3.8Gi       2.0Gi       183Mi       6.4Mi       2.0Gi       1.8Gi
Swap:             0B          0B          0B
Linux r8s-cp-wlp7h-lhvgq 6.8.0-71-generic #71-Ubuntu SMP PREEMPT_DYNAMIC Tue Jul 22 16:52:38 UTC 2025 x86_64 x86_64 x86_64 GNU/Linux
Ubuntu 24.04.3 LTS
rke2 version v1.33.7+rke2r1 (b0a4ec8463abd1e23e41f213fdb54ad8006c693b)
go version go1.24.11 X:boringcrypto
NAME                                                    IMAGE
kube-apiserver-r8s-cp-wlp7h-lhvgq                       index.docker.io/rancher/hardened-kubernetes:v1.33.7-rke2r1-build20251210
cattle-cluster-agent-78b58bc994-dmc2f                   rancher/rancher-agent:v2.12.3
NAME                           CHART                          VERSION     RELEASE NAME                   RELEASE VERSION   STATUS
rke2-calico                    rke2-calico                    v3.31.200   rke2-calico                    1                 deployed
`
	
	versionsPath := filepath.Join(tmpDir, "versions")
	if err := os.WriteFile(versionsPath, []byte(versionsContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	info, err := ParseVersions(tmpDir)
	if err != nil {
		t.Fatalf("ParseVersions failed: %v", err)
	}
	
	// Verify header parsing
	if info.CollectionDate != "Tue Mar  3 03:49:05 UTC 2026" {
		t.Errorf("CollectionDate = %q, want %q", info.CollectionDate, "Tue Mar  3 03:49:05 UTC 2026")
	}
	
	if info.MemoryTotal != "3.8Gi" {
		t.Errorf("MemoryTotal = %q, want %q", info.MemoryTotal, "3.8Gi")
	}
	
	if info.MemoryUsed != "2.0Gi" {
		t.Errorf("MemoryUsed = %q, want %q", info.MemoryUsed, "2.0Gi")
	}
	
	if info.Hostname != "r8s-cp-wlp7h-lhvgq" {
		t.Errorf("Hostname = %q, want %q", info.Hostname, "r8s-cp-wlp7h-lhvgq")
	}
	
	if info.KernelVersion != "6.8.0-71-generic" {
		t.Errorf("KernelVersion = %q, want %q", info.KernelVersion, "6.8.0-71-generic")
	}
	
	if info.OSName != "Ubuntu 24.04.3 LTS" {
		t.Errorf("OSName = %q, want %q", info.OSName, "Ubuntu 24.04.3 LTS")
	}
	
	if info.DistroVersion != "rke2 version v1.33.7+rke2r1 (b0a4ec8463abd1e23e41f213fdb54ad8006c693b)" {
		t.Errorf("DistroVersion = %q, want RKE2 version", info.DistroVersion)
	}
	
	// Verify RKE2 images
	if len(info.RKE2Images) != 1 {
		t.Errorf("RKE2Images count = %d, want 1", len(info.RKE2Images))
	} else {
		if info.RKE2Images[0].PodName != "kube-apiserver-r8s-cp-wlp7h-lhvgq" {
			t.Errorf("RKE2Images[0].PodName = %q", info.RKE2Images[0].PodName)
		}
		if !strings.Contains(info.RKE2Images[0].Image, "hardened-kubernetes") {
			t.Errorf("RKE2Images[0].Image = %q", info.RKE2Images[0].Image)
		}
	}
	
	// Verify Cattle images
	if len(info.CattleImages) != 1 {
		t.Errorf("CattleImages count = %d, want 1", len(info.CattleImages))
	} else {
		if info.CattleImages[0].PodName != "cattle-cluster-agent-78b58bc994-dmc2f" {
			t.Errorf("CattleImages[0].PodName = %q", info.CattleImages[0].PodName)
		}
	}
	
	// Verify Helm releases
	if len(info.HelmReleases) != 1 {
		t.Errorf("HelmReleases count = %d, want 1", len(info.HelmReleases))
	} else {
		if info.HelmReleases[0].Name != "rke2-calico" {
			t.Errorf("HelmReleases[0].Name = %q, want rke2-calico", info.HelmReleases[0].Name)
		}
		if info.HelmReleases[0].Status != "deployed" {
			t.Errorf("HelmReleases[0].Status = %q, want deployed", info.HelmReleases[0].Status)
		}
	}
}

func TestParseVersions_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	
	_, err := ParseVersions(tmpDir)
	if err == nil {
		t.Error("Expected error for missing file, got nil")
	}
}

func TestParseVersions_K3sVersion(t *testing.T) {
	tmpDir := t.TempDir()
	versionsContent := `Test Date
Linux test-node 5.15.0-generic
k3s version v1.28.5+k3s1
NAME IMAGE
coredns-abc123 rancher/mirrored-coredns:1.10.1
`
	
	versionsPath := filepath.Join(tmpDir, "versions")
	os.WriteFile(versionsPath, []byte(versionsContent), 0644)
	
	info, err := ParseVersions(tmpDir)
	if err != nil {
		t.Fatalf("ParseVersions failed: %v", err)
	}
	
	if !strings.Contains(info.DistroVersion, "k3s version") {
		t.Errorf("Expected k3s version, got %q", info.DistroVersion)
	}
}

func TestParseVersions_MultipleHelmReleases(t *testing.T) {
	tmpDir := t.TempDir()
	versionsContent := `Test
Linux test 5.0.0
rke2 version v1.0.0
NAME IMAGE
kube-apiserver test:latest
NAME CHART VERSION RELEASE RELEASE_VER STATUS
rke2-calico calico v1.0 calico 1 deployed
rke2-coredns coredns v2.0 coredns 1 deployed
rke2-metrics metrics v3.0 metrics 1 failed
`
	
	versionsPath := filepath.Join(tmpDir, "versions")
	os.WriteFile(versionsPath, []byte(versionsContent), 0644)
	
	info, err := ParseVersions(tmpDir)
	if err != nil {
		t.Fatalf("ParseVersions failed: %v", err)
	}
	
	if len(info.HelmReleases) != 3 {
		t.Errorf("HelmReleases count = %d, want 3", len(info.HelmReleases))
	}
	
	// Check failed status is captured
	foundFailed := false
	for _, r := range info.HelmReleases {
		if r.Status == "failed" {
			foundFailed = true
			break
		}
	}
	if !foundFailed {
		t.Error("Expected to find a failed helm release")
	}
}
