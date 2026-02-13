package bundle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseDMesg_OOMKills(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()
	systeminfoDir := filepath.Join(tmpDir, "systeminfo")
	if err := os.MkdirAll(systeminfoDir, 0755); err != nil {
		t.Fatalf("Failed to create systeminfo dir: %v", err)
	}

	// Test data: dmesg with OOM kills
	dmesgData := `[    0.000000] Linux version 5.15.0-generic
[   12.345678] Initializing cgroup subsys cpuset
[ 1234.567890] Out of memory: Killed process 12345 (nginx) total-vm:131072kB, anon-rss:65536kB, file-rss:0kB, shmem-rss:0kB, UID:0 pgtables:128kB oom_score_adj:0
[ 1234.567901] oom-kill:constraint=CONSTRAINT_NONE,nodemask=(null),cpuset=cri-containerd-abc123.scope,mems_allowed=0,oom_memcg=/system.slice/containerd.service,task_memcg=/kubepods.slice,task=nginx,pid=12345,uid=0
[ 5678.901234] Memory cgroup out of memory: Kill process 67890 (java) score 999 or sacrifice child
[ 5678.901245] Out of memory: Killed process 67890 (java) total-vm:4194304kB, anon-rss:3145728kB, file-rss:0kB, shmem-rss:0kB, UID:1000 pgtables:8192kB oom_score_adj:999
[ 9999.999999] Normal system operation
`

	dmesgPath := filepath.Join(systeminfoDir, "dmesg")
	if err := os.WriteFile(dmesgPath, []byte(dmesgData), 0644); err != nil {
		t.Fatalf("Failed to write test dmesg file: %v", err)
	}

	// Parse dmesg
	analysis, err := ParseDMesg(tmpDir)
	if err != nil {
		t.Fatalf("ParseDMesg failed: %v", err)
	}

	// Should have 2 OOM kills
	if len(analysis.OOMKills) != 2 {
		t.Errorf("Expected 2 OOM kills, got %d", len(analysis.OOMKills))
	}

	// Check first OOM kill
	if analysis.OOMKills[0].VictimName != "nginx" {
		t.Errorf("Expected first victim to be 'nginx', got '%s'", analysis.OOMKills[0].VictimName)
	}
	if analysis.OOMKills[0].VictimPID != 12345 {
		t.Errorf("Expected first victim PID to be 12345, got %d", analysis.OOMKills[0].VictimPID)
	}
	if analysis.OOMKills[0].OOMScore != 0 {
		t.Errorf("Expected first OOM score to be 0, got %d", analysis.OOMKills[0].OOMScore)
	}

	// Check second OOM kill
	if analysis.OOMKills[1].VictimName != "java" {
		t.Errorf("Expected second victim to be 'java', got '%s'", analysis.OOMKills[1].VictimName)
	}
	if analysis.OOMKills[1].VictimPID != 67890 {
		t.Errorf("Expected second victim PID to be 67890, got %d", analysis.OOMKills[1].VictimPID)
	}
	if analysis.OOMKills[1].OOMScore != 999 {
		t.Errorf("Expected second OOM score to be 999, got %d", analysis.OOMKills[1].OOMScore)
	}

	// Should have memory pressure detected
	if !analysis.MemoryPressure {
		t.Error("Expected MemoryPressure to be true")
	}
}

func TestParseDMesg_NoOOM(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()
	systeminfoDir := filepath.Join(tmpDir, "systeminfo")
	if err := os.MkdirAll(systeminfoDir, 0755); err != nil {
		t.Fatalf("Failed to create systeminfo dir: %v", err)
	}

	// Test data: dmesg without OOM kills
	dmesgData := `[    0.000000] Linux version 5.15.0-generic
[   12.345678] Initializing cgroup subsys cpuset
[   45.678901] device eth0 entered promiscuous mode
[ 9999.999999] Normal system operation
`

	dmesgPath := filepath.Join(systeminfoDir, "dmesg")
	if err := os.WriteFile(dmesgPath, []byte(dmesgData), 0644); err != nil {
		t.Fatalf("Failed to write test dmesg file: %v", err)
	}

	// Parse dmesg
	analysis, err := ParseDMesg(tmpDir)
	if err != nil {
		t.Fatalf("ParseDMesg failed: %v", err)
	}

	// Should have 0 OOM kills
	if len(analysis.OOMKills) != 0 {
		t.Errorf("Expected 0 OOM kills, got %d", len(analysis.OOMKills))
	}

	// Should not have memory pressure
	if analysis.MemoryPressure {
		t.Error("Expected MemoryPressure to be false")
	}
}

func TestDMesgAnalysis_HasOOMKills(t *testing.T) {
	// Test with OOM kills
	analysisWithOOM := &DMesgAnalysis{
		OOMKills: []DMesgOOMKill{
			{VictimName: "nginx", VictimPID: 12345},
		},
	}
	if !analysisWithOOM.HasOOMKills() {
		t.Error("Expected HasOOMKills() to return true")
	}

	// Test without OOM kills
	analysisWithoutOOM := &DMesgAnalysis{
		OOMKills: []DMesgOOMKill{},
	}
	if analysisWithoutOOM.HasOOMKills() {
		t.Error("Expected HasOOMKills() to return false")
	}
}

func TestDMesgAnalysis_GetOOMKillSummary(t *testing.T) {
	// Test with OOM kills
	analysisWithOOM := &DMesgAnalysis{
		OOMKills: []DMesgOOMKill{
			{VictimName: "nginx", VictimPID: 12345, OOMScore: 0},
			{VictimName: "java", VictimPID: 67890, OOMScore: 999},
		},
		MemoryPressure: true,
	}
	
	summary := analysisWithOOM.GetOOMKillSummary()
	if summary == "" {
		t.Error("Expected non-empty summary")
	}
	if !strings.Contains(summary, "2 OOM kill(s)") {
		t.Errorf("Summary should mention '2 OOM kill(s)', got: %s", summary)
	}

	// Test without OOM kills
	analysisWithoutOOM := &DMesgAnalysis{
		OOMKills: []DMesgOOMKill{},
	}
	
	summary = analysisWithoutOOM.GetOOMKillSummary()
	if !strings.Contains(summary, "No OOM kills detected") {
		t.Errorf("Summary should say 'No OOM kills detected', got: %s", summary)
	}
}

