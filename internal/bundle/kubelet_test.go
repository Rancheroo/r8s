package bundle

import (
	"os"
	"path/filepath"
	"testing"
)

// createKubeletTestBundle creates a minimal bundle with journald logs
func createKubeletTestBundle(t *testing.T) (string, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "r8s-kubelet-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create journald directory
	journalDir := filepath.Join(tmpDir, "journald")
	if err := os.MkdirAll(journalDir, 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create journald dir: %v", err)
	}

	cleanup := func() { os.RemoveAll(tmpDir) }
	return tmpDir, cleanup
}

// --- ParseKubeletLogs integration tests ---

func TestParseKubeletLogs_MissingFile(t *testing.T) {
	bundleRoot, cleanup := createKubeletTestBundle(t)
	defer cleanup()

	// No journal file exists
	issues, err := ParseKubeletLogs(bundleRoot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if issues != nil {
		t.Errorf("expected nil for missing file, got %v", issues)
	}
}

func TestParseKubeletLogs_NoIssues(t *testing.T) {
	bundleRoot, cleanup := createKubeletTestBundle(t)
	defer cleanup()

	content := `Dec 13 14:06:17 node1 rke2[1234]: time="2024-12-13T14:06:17Z" level=info msg="Kubelet started successfully"
Dec 13 14:06:18 node1 rke2[1234]: time="2024-12-13T14:06:18Z" level=info msg="Normal operation"
`
	journalPath := filepath.Join(bundleRoot, "journald", "rke2-server")
	os.WriteFile(journalPath, []byte(content), 0644)

	issues, err := ParseKubeletLogs(bundleRoot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("expected 0 issues, got %d", len(issues))
	}
}

func TestParseKubeletLogs_HTTP502(t *testing.T) {
	bundleRoot, cleanup := createKubeletTestBundle(t)
	defer cleanup()

	content := `Dec 13 14:06:17 node1 rke2[1234]: time="2024-12-13T14:06:17Z" level=error msg="Sending HTTP/1.1 502 response to 10.0.0.1:12345"
`
	journalPath := filepath.Join(bundleRoot, "journald", "rke2-server")
	os.WriteFile(journalPath, []byte(content), 0644)

	issues, err := ParseKubeletLogs(bundleRoot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].Pattern != "HTTP 502" {
		t.Errorf("expected pattern 'HTTP 502', got '%s'", issues[0].Pattern)
	}
	if issues[0].Count != 1 {
		t.Errorf("expected count 1, got %d", issues[0].Count)
	}
}

func TestParseKubeletLogs_ConnectionTimeout(t *testing.T) {
	bundleRoot, cleanup := createKubeletTestBundle(t)
	defer cleanup()

	content := `Dec 13 14:06:17 node1 rke2[1234]: time="2024-12-13T14:06:17Z" level=error msg="dial tcp 10.0.0.1:6443: connect: connection timed out"
`
	journalPath := filepath.Join(bundleRoot, "journald", "rke2-server")
	os.WriteFile(journalPath, []byte(content), 0644)

	issues, err := ParseKubeletLogs(bundleRoot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].Pattern != "connection timeout" {
		t.Errorf("expected pattern 'connection timeout', got '%s'", issues[0].Pattern)
	}
}

func TestParseKubeletLogs_MultiplePatterns(t *testing.T) {
	bundleRoot, cleanup := createKubeletTestBundle(t)
	defer cleanup()

	content := `Dec 13 14:06:17 node1 rke2[1234]: time="2024-12-13T14:06:17Z" level=error msg="Sending HTTP/1.1 502 response"
Dec 13 14:06:18 node1 rke2[1234]: time="2024-12-13T14:06:18Z" level=error msg="dial tcp: connect: connection timed out"
Dec 13 14:06:19 node1 rke2[1234]: time="2024-12-13T14:06:19Z" level=error msg="TLS handshake error from 10.0.0.1:12345"
`
	journalPath := filepath.Join(bundleRoot, "journald", "rke2-server")
	os.WriteFile(journalPath, []byte(content), 0644)

	issues, err := ParseKubeletLogs(bundleRoot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 3 {
		t.Fatalf("expected 3 issues, got %d", len(issues))
	}
}

func TestParseKubeletLogs_DuplicatePatternCounts(t *testing.T) {
	bundleRoot, cleanup := createKubeletTestBundle(t)
	defer cleanup()

	content := `Dec 13 14:06:17 node1 rke2[1234]: time="2024-12-13T14:06:17Z" level=error msg="Sending HTTP/1.1 502 response"
Dec 13 14:06:18 node1 rke2[1234]: time="2024-12-13T14:06:18Z" level=error msg="Sending HTTP/1.1 502 response again"
Dec 13 14:06:19 node1 rke2[1234]: time="2024-12-13T14:06:19Z" level=error msg="Sending HTTP/1.1 502 response third"
`
	journalPath := filepath.Join(bundleRoot, "journald", "rke2-server")
	os.WriteFile(journalPath, []byte(content), 0644)

	issues, err := ParseKubeletLogs(bundleRoot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue type, got %d", len(issues))
	}
	if issues[0].Count != 3 {
		t.Errorf("expected count 3, got %d", issues[0].Count)
	}
}

// --- parseJournaldLogs unit tests ---

func TestParseJournaldLogs_Empty(t *testing.T) {
	issues, err := parseJournaldLogs("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("expected 0 issues, got %d", len(issues))
	}
}

func TestParseJournaldLogs_OnlyWhitespace(t *testing.T) {
	issues, err := parseJournaldLogs("   \n\n   \n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("expected 0 issues, got %d", len(issues))
	}
}

func TestParseJournaldLogs_AllPatterns(t *testing.T) {
	tests := []struct {
		name     string
		logLine  string
		expected string // expected pattern name
	}{
		{"HTTP 502", `time="2024-01-01T00:00:00Z" level=error msg="Sending HTTP/1.1 502 response"`, "HTTP 502"},
		{"connection timeout", `time="2024-01-01T00:00:00Z" level=error msg="connect: connection timed out"`, "connection timeout"},
		{"remotedialer timeout", `time="2024-01-01T00:00:00Z" level=error msg="remotedialer server connection i/o timeout"`, "remotedialer timeout"},
		{"DNS nameserver limits", `time="2024-01-01T00:00:00Z" level=warning msg="Nameserver limits exceeded"`, "DNS nameserver limits"},
		{"TLS handshake error", `time="2024-01-01T00:00:00Z" level=error msg="TLS handshake error from client"`, "TLS handshake error"},
		{"EOF during TLS", `time="2024-01-01T00:00:00Z" level=error msg="EOF during TLS handshake"`, "EOF during TLS"},
		{"unsupported protocol", `time="2024-01-01T00:00:00Z" level=error msg="protocol not supported"`, "unsupported protocol"},
		{"version not supported", `time="2024-01-01T00:00:00Z" level=error msg="version not supported"`, "version not supported"},
		{"no route to host", `time="2024-01-01T00:00:00Z" level=error msg="no route to host"`, "no route to host"},
		{"network unreachable", `time="2024-01-01T00:00:00Z" level=error msg="network is unreachable"`, "network unreachable"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logContent := "Dec 13 14:06:17 node1 rke2[1234]: " + tt.logLine
			issues, err := parseJournaldLogs(logContent)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(issues) != 1 {
				t.Fatalf("expected 1 issue, got %d", len(issues))
			}
			if issues[0].Pattern != tt.expected {
				t.Errorf("expected pattern '%s', got '%s'", tt.expected, issues[0].Pattern)
			}
		})
	}
}

// --- parseJournaldLine unit tests ---

func TestParseJournaldLine(t *testing.T) {
	tests := []struct {
		name            string
		line            string
		wantTimestamp   string
		wantLevel       string
		wantMessage     string
	}{
		{
			name:          "standard format",
			line:          `Dec 13 14:06:17 node1 rke2[1234]: time="2024-12-13T14:06:17Z" level=error msg="Something went wrong"`,
			wantTimestamp: "2024-12-13T14:06:17Z",
			wantLevel:     "error",
			wantMessage:   "Something went wrong",
		},
		{
			name:          "warning level",
			line:          `Dec 13 14:06:17 node1 rke2[1234]: time="2024-12-13T14:06:17Z" level=warning msg="This is a warning"`,
			wantTimestamp: "2024-12-13T14:06:17Z",
			wantLevel:     "warning",
			wantMessage:   "This is a warning",
		},
		{
			name:          "info level",
			line:          `Dec 13 14:06:17 node1 rke2[1234]: time="2024-12-13T14:06:17Z" level=info msg="Info message"`,
			wantTimestamp: "2024-12-13T14:06:17Z",
			wantLevel:     "info",
			wantMessage:   "Info message",
		},
		{
			name:          "no structured format",
			line:          `Dec 13 14:06:17 node1 rke2[1234]: Plain text message`,
			wantTimestamp: "Dec 13 14:06:17 node1 rke2[1234]",
			wantLevel:     "info",
			wantMessage:   "Plain text message",
		},
		{
			name:          "empty line",
			line:          "",
			wantTimestamp: "",
			wantLevel:     "",
			wantMessage:   "",
		},
		{
			name:          "no colon",
			line:          "This line has no colon separator",
			wantTimestamp: "",
			wantLevel:     "",
			wantMessage:   "",
		},
		{
			name:          "too short",
			line:          "short",
			wantTimestamp: "",
			wantLevel:     "",
			wantMessage:   "",
		},
		{
			name:          "msg with quotes inside",
			line:          `Dec 13 14:06:17 node1 rke2[1234]: time="2024-12-13T14:06:17Z" level=error msg="Error: connection refused"`,
			wantTimestamp: "2024-12-13T14:06:17Z",
			wantLevel:     "error",
			wantMessage:   "Error: connection refused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timestamp, level, message := parseJournaldLine(tt.line)
			if timestamp != tt.wantTimestamp {
				t.Errorf("timestamp = %q, want %q", timestamp, tt.wantTimestamp)
			}
			if level != tt.wantLevel {
				t.Errorf("level = %q, want %q", level, tt.wantLevel)
			}
			if message != tt.wantMessage {
				t.Errorf("message = %q, want %q", message, tt.wantMessage)
			}
		})
	}
}
