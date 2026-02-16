package bundle

import (
	"testing"
)

func TestRKE2ControlPlaneEvents_GetSummary(t *testing.T) {
	// Test with no issues
	noIssues := &RKE2ControlPlaneEvents{}
	summary := noIssues.GetSummary()
	if summary != "No RKE2 control plane issues detected" {
		t.Errorf("Expected no issues message, got: %s", summary)
	}

	// Test with issues
	withIssues := &RKE2ControlPlaneEvents{
		ServerRestarts: []ControlPlaneEvent{
			{Severity: "warning", Message: "restart"},
			{Severity: "warning", Message: "restart"},
		},
		EtcdIssues: []ControlPlaneEvent{
			{Severity: "error", Message: "etcd timeout"},
		},
		CertificateIssues: []ControlPlaneEvent{
			{Severity: "critical", Message: "cert expired"},
		},
	}

	summary = withIssues.GetSummary()
	if summary == "No RKE2 control plane issues detected" {
		t.Error("Expected issues summary, got no issues message")
	}

	// Check that all issue types are mentioned
	if !containsStr(summary, "server restart") {
		t.Error("Expected server restart in summary")
	}
	if !containsStr(summary, "etcd issue") {
		t.Error("Expected etcd issue in summary")
	}
	if !containsStr(summary, "certificate issue") {
		t.Error("Expected certificate issue in summary")
	}
}

func TestRKE2ControlPlaneEvents_GetCriticalIssues(t *testing.T) {
	events := &RKE2ControlPlaneEvents{
		ServerRestarts: []ControlPlaneEvent{
			{Severity: "warning", Message: "normal restart"},
			{Severity: "critical", Message: "crash loop"},
		},
		EtcdIssues: []ControlPlaneEvent{
			{Severity: "info", Message: "etcd info"},
			{Severity: "error", Message: "etcd timeout"},
		},
	}

	critical := events.GetCriticalIssues()

	// Should only get critical and error severity
	if len(critical) != 2 {
		t.Fatalf("Expected 2 critical issues, got: %d", len(critical))
	}

	// Verify we got the right ones
	hasCrashLoop := false
	hasEtcdTimeout := false
	for _, issue := range critical {
		if issue.Message == "crash loop" {
			hasCrashLoop = true
		}
		if issue.Message == "etcd timeout" {
			hasEtcdTimeout = true
		}
	}

	if !hasCrashLoop {
		t.Error("Expected crash loop (critical) in results")
	}
	if !hasEtcdTimeout {
		t.Error("Expected etcd timeout (error) in results")
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStrHelper(s, substr))
}

func containsStrHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
