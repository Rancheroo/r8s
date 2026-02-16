package bundle

import (
	"os"
	"path/filepath"
	"testing"
)

// createEtcdTestBundle creates a minimal bundle with etcd test files
func createEtcdTestBundle(t *testing.T, files map[string]string) (string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "r8s-etcd-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create rke2 structure for getBundleRoot to work properly
	rke2Dir := filepath.Join(tmpDir, "rke2")
	os.MkdirAll(rke2Dir, 0755)

	// Create etcd directory inside rke2 (where functions look for it)
	etcdDir := filepath.Join(rke2Dir, "etcd")
	os.MkdirAll(etcdDir, 0755)

	// Write test files
	for filename, content := range files {
		path := filepath.Join(etcdDir, filename)
		os.WriteFile(path, []byte(content), 0644)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	// Return the rke2 directory as bundle root (getBundleRoot uses this)
	return rke2Dir, cleanup
}

func TestParseEtcdHealth_NoAlarms(t *testing.T) {
	files := map[string]string{
		"alarmlist":   "memberID:0",
		"endpointhealth": "https://127.0.0.1:2379 is healthy: successfully committed proposal",
	}
	
	bundlePath, cleanup := createEtcdTestBundle(t, files)
	defer cleanup()

	health, err := ParseEtcdHealth(bundlePath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !health.Healthy {
		t.Error("Expected healthy to be true")
	}
	if health.HasAlarms {
		t.Error("Expected no alarms")
	}
}

func TestParseEtcdHealth_WithAlarms(t *testing.T) {
	// Alarm file format: alarms listed without memberID prefix
	files := map[string]string{
		"alarmlist": "alarm:NOSPACE\nalarm:NOSPACE",
		"endpointhealth": "https://127.0.0.1:2379 is healthy",
	}
	
	bundlePath, cleanup := createEtcdTestBundle(t, files)
	defer cleanup()

	health, err := ParseEtcdHealth(bundlePath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !health.HasAlarms {
		t.Error("Expected alarms to be detected")
	}
	if health.AlarmType != "NOSPACE" {
		t.Errorf("Expected alarm type NOSPACE, got: %s", health.AlarmType)
	}
	if health.AlarmCount != 2 {
		t.Errorf("Expected 2 alarms, got: %d", health.AlarmCount)
	}
}

func TestParseEtcdHealth_Unhealthy(t *testing.T) {
	files := map[string]string{
		"alarmlist":      "memberID:0",
		"endpointhealth": "https://127.0.0.1:2379 is unhealthy",
	}
	
	bundlePath, cleanup := createEtcdTestBundle(t, files)
	defer cleanup()

	health, err := ParseEtcdHealth(bundlePath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if health.Healthy {
		t.Error("Expected healthy to be false for unhealthy endpoint")
	}
}

func TestParseEtcdHealth_MissingFiles(t *testing.T) {
	// Empty bundle - no etcd files
	tmpDir, err := os.MkdirTemp("", "r8s-etcd-empty-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create rke2/etcd but no files
	os.MkdirAll(filepath.Join(tmpDir, "rke2", "etcd"), 0755)

	health, err := ParseEtcdHealth(tmpDir)
	if err != nil {
		t.Fatalf("Expected no error for missing files, got: %v", err)
	}

	// Should default to healthy when no data
	if !health.Healthy {
		t.Error("Expected healthy=true when files missing (default assumption)")
	}
}

func TestParseEtcdMemberList(t *testing.T) {
	content := `15e9d2d844399be2, started, node1, https://10.0.0.1:2380, https://10.0.0.1:2379, false
2a8f4b9c12345678, started, node2, https://10.0.0.2:2380, https://10.0.0.2:2379, false
3b1c5d0e87654321, started, learner-node, https://10.0.0.3:2380, https://10.0.0.3:2379, true
`
	
	files := map[string]string{
		"memberlist": content,
	}
	
	bundlePath, cleanup := createEtcdTestBundle(t, files)
	defer cleanup()

	members, err := ParseEtcdMemberList(bundlePath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(members) != 3 {
		t.Fatalf("Expected 3 members, got: %d", len(members))
	}

	// Check first member
	if members[0].ID != "15e9d2d844399be2" {
		t.Errorf("Expected ID 15e9d2d844399be2, got: %s", members[0].ID)
	}
	if members[0].Name != "node1" {
		t.Errorf("Expected name node1, got: %s", members[0].Name)
	}
	if members[0].State != "started" {
		t.Errorf("Expected state started, got: %s", members[0].State)
	}
	if members[0].IsLearner {
		t.Error("Expected first member to not be a learner")
	}

	// Check learner
	if !members[2].IsLearner {
		t.Error("Expected third member to be a learner")
	}
}

func TestParseEtcdMemberList_MissingFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "r8s-etcd-missing-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	os.MkdirAll(filepath.Join(tmpDir, "rke2", "etcd"), 0755)

	_, err = ParseEtcdMemberList(tmpDir)
	if err == nil {
		t.Error("Expected error for missing memberlist file")
	}
}

func TestParseEtcdEndpointStatus(t *testing.T) {
	content := `+----------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
|   ENDPOINT     |        ID        | VERSION | DB SIZE | IS LEADER | IS LEARNER | RAFT TERM | RAFT INDEX | RAFT APPLIED INDEX | ERRORS |
+----------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
| 127.0.0.1:2379 | 15e9d2d844399be2 |  3.5.21 |   50 MB |     true  |    false   |         5 |     12345 |              12345 |        |
+----------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
`
	
	files := map[string]string{
		"endpointstatus": content,
	}
	
	bundlePath, cleanup := createEtcdTestBundle(t, files)
	defer cleanup()

	leaderID, dbSize, dbSizeBytes, version, raftTerm, raftIndex, err := ParseEtcdEndpointStatus(bundlePath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if leaderID != "15e9d2d844399be2" {
		t.Errorf("Expected leader ID 15e9d2d844399be2, got: %s", leaderID)
	}
	if dbSize != "50 MB" {
		t.Errorf("Expected DB size '50 MB', got: %s", dbSize)
	}
	if dbSizeBytes != 52428800 {
		t.Errorf("Expected DB size bytes 52428800, got: %d", dbSizeBytes)
	}
	if version != "3.5.21" {
		t.Errorf("Expected version 3.5.21, got: %s", version)
	}
	if raftTerm != 5 {
		t.Errorf("Expected raft term 5, got: %d", raftTerm)
	}
	if raftIndex != 12345 {
		t.Errorf("Expected raft index 12345, got: %d", raftIndex)
	}
}

func TestParseEtcdEndpointStatus_NotLeader(t *testing.T) {
	content := `+----------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
|   ENDPOINT     |        ID        | VERSION | DB SIZE | IS LEADER | IS LEARNER | RAFT TERM | RAFT INDEX | RAFT APPLIED INDEX | ERRORS |
+----------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
| 127.0.0.1:2379 | 15e9d2d844399be2 |  3.5.21 |   50 MB |    false  |    false   |         5 |     12345 |              12345 |        |
+----------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
`
	
	files := map[string]string{
		"endpointstatus": content,
	}
	
	bundlePath, cleanup := createEtcdTestBundle(t, files)
	defer cleanup()

	leaderID, _, _, _, _, _, err := ParseEtcdEndpointStatus(bundlePath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if leaderID != "" {
		t.Errorf("Expected empty leader ID for non-leader, got: %s", leaderID)
	}
}

func TestParseDBSize(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"50 MB", 52428800},
		{"100 MB", 104857600},
		{"1 GB", 1073741824},
		{"500 KB", 512000},
		{"1000 B", 1000},
		{"invalid", 0},
		{"", 0},
	}

	for _, tc := range tests {
		result := parseDBSize(tc.input)
		if result != tc.expected {
			t.Errorf("parseDBSize(%q) = %d, expected %d", tc.input, result, tc.expected)
		}
	}
}

func TestParseEtcdDetails(t *testing.T) {
	files := map[string]string{
		"alarmlist":      "memberID:0",
		"endpointhealth": "https://127.0.0.1:2379 is healthy",
		"memberlist":     "15e9d2d844399be2, started, node1, https://10.0.0.1:2380, https://10.0.0.1:2379, false",
		"endpointstatus": `+----------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
|   ENDPOINT     |        ID        | VERSION | DB SIZE | IS LEADER | IS LEARNER | RAFT TERM | RAFT INDEX | RAFT APPLIED INDEX | ERRORS |
+----------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
| 127.0.0.1:2379 | 15e9d2d844399be2 |  3.5.21 |   50 MB |     true  |    false   |         5 |     12345 |              12345 |        |
+----------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
`,
	}
	
	bundlePath, cleanup := createEtcdTestBundle(t, files)
	defer cleanup()

	details, err := ParseEtcdDetails(bundlePath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !details.Healthy {
		t.Error("Expected healthy cluster")
	}
	if details.HasAlarms {
		t.Error("Expected no alarms")
	}
	if details.MemberCount != 1 {
		t.Errorf("Expected 1 member, got: %d", details.MemberCount)
	}
	if details.DBSize != "50 MB" {
		t.Errorf("Expected DB size 50 MB, got: %s", details.DBSize)
	}
	// Should not need compaction at 50MB
	if details.NeedsCompaction {
		t.Error("Expected no compaction needed at 50MB")
	}
}

func TestParseEtcdDetails_NeedsCompaction(t *testing.T) {
	files := map[string]string{
		"alarmlist":      "memberID:0",
		"endpointhealth": "https://127.0.0.1:2379 is healthy",
		"endpointstatus": `+----------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
|   ENDPOINT     |        ID        | VERSION | DB SIZE | IS LEADER | IS LEARNER | RAFT TERM | RAFT INDEX | RAFT APPLIED INDEX | ERRORS |
+----------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
| 127.0.0.1:2379 | 15e9d2d844399be2 |  3.5.21 |  150 MB |     true  |    false   |         5 |     12345 |              12345 |        |
+----------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
`,
	}
	
	bundlePath, cleanup := createEtcdTestBundle(t, files)
	defer cleanup()

	details, err := ParseEtcdDetails(bundlePath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !details.NeedsCompaction {
		t.Error("Expected compaction needed at 150MB")
	}
	if details.CompactionReason == "" {
		t.Error("Expected compaction reason to be set")
	}
}
