package bundle

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseJournald(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create systemlogs directory structure
	systemlogsDir := filepath.Join(tmpDir, "systemlogs")
	if err := os.MkdirAll(systemlogsDir, 0755); err != nil {
		t.Fatalf("Failed to create systemlogs dir: %v", err)
	}

	// Sample log data with mixed issues
	// Note: Using current year to match parser behavior
	currentYear := time.Now().Year()
	logData := `
Jan 01 10:00:00 node1 rke2-server[1234]: I0101 10:00:00.000000 1234 server.go:1] Starting rke2-server v1.24.0
Jan 01 10:05:00 node1 rke2-server[1234]: E0101 10:05:00.000000 1234 etcd.go:1] etcd cluster is unhealthy: member 123 is unreachable
Jan 01 10:10:00 node1 rke2-server[1234]: E0101 10:10:00.000000 1234 cert.go:1] x509: certificate has expired or is not yet valid
Jan 01 10:15:00 node1 rke2-server[1234]: F0101 10:15:00.000000 1234 panic.go:1] panic: critical error in controller
`

	logPath := filepath.Join(systemlogsDir, "journald-rke2-server")
	if err := os.WriteFile(logPath, []byte(logData), 0644); err != nil {
		t.Fatalf("Failed to write log file: %v", err)
	}

	events, err := ParseJournald(tmpDir)
	if err != nil {
		t.Fatalf("ParseJournald failed: %v", err)
	}

	// Check Server Restarts
	if len(events.ServerRestarts) != 1 {
		t.Errorf("Expected 1 server restart, got %d", len(events.ServerRestarts))
	} else if events.ServerRestarts[0].Unit != "rke2-server" {
		t.Errorf("Expected unit rke2-server, got %s", events.ServerRestarts[0].Unit)
	}

	// Check Etcd Issues
	if len(events.EtcdIssues) != 1 {
		t.Errorf("Expected 1 etcd issue, got %d", len(events.EtcdIssues))
	} else if events.EtcdIssues[0].Severity != "error" {
		t.Errorf("Expected error severity, got %s", events.EtcdIssues[0].Severity)
	}

	// Check Cert Issues
	if len(events.CertificateIssues) != 1 {
		t.Errorf("Expected 1 cert issue, got %d", len(events.CertificateIssues))
	}

	// Check Fatal/Panic (AgentIssues bucket for now based on implementation)
	if len(events.AgentIssues) != 1 {
		t.Errorf("Expected 1 critical error, got %d", len(events.AgentIssues))
	} else if events.AgentIssues[0].Severity != "critical" {
		t.Errorf("Expected critical severity, got %s", events.AgentIssues[0].Severity)
	}

	// Verify timestamps parsed correctly (should be current year)
	if events.ServerRestarts[0].Timestamp.Year() != currentYear {
		t.Errorf("Expected year %d, got %d", currentYear, events.ServerRestarts[0].Timestamp.Year())
	}
}

func TestParseJournald_NoIssues(t *testing.T) {
	tmpDir := t.TempDir()
	systemlogsDir := filepath.Join(tmpDir, "systemlogs")
	if err := os.MkdirAll(systemlogsDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	logData := `
Jan 01 10:00:00 node1 rke2-server[1234]: I0101 10:00:00.000000 1234 server.go:1] Normal operation
Jan 01 10:01:00 node1 rke2-server[1234]: I0101 10:01:00.000000 1234 server.go:1] Still running
`
	if err := os.WriteFile(filepath.Join(systemlogsDir, "journald-rke2-server"), []byte(logData), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	events, _ := ParseJournald(tmpDir)
	
	if events.HasIssues() {
		t.Error("Expected no issues, but HasIssues() returned true")
	}
}
