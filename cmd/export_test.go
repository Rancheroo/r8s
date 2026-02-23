package cmd

import (
	"encoding/json"
	"testing"

	"github.com/Rancheroo/r8s/internal/bundle"
)

func TestExportMeta(t *testing.T) {
	meta := ExportMeta{
		GeneratedAt: "2026-02-23T10:00:00Z",
		BundlePath:  "/test/bundle",
		BundleType:  "rke2",
		R8SVersion:  "v0.8.0",
	}

	// Test JSON marshaling
	data, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("failed to marshal ExportMeta: %v", err)
	}

	// Verify all fields present
	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal ExportMeta: %v", err)
	}

	requiredFields := []string{"generated_at", "bundle_path", "bundle_type", "r8s_version"}
	for _, field := range requiredFields {
		if _, ok := decoded[field]; !ok {
			t.Errorf("ExportMeta missing field: %s", field)
		}
	}
}

func TestExportFinding(t *testing.T) {
	finding := ExportFinding{
		ID:          "test-001",
		Severity:    "critical",
		Category:    "OOM",
		Title:       "OOMKill Detected",
		Description: "Container killed due to memory limits",
		Namespace:   "default",
		Resource:    "pod/app-123",
		Suggestion:  "Increase memory limit",
	}

	// Test JSON marshaling
	data, err := json.Marshal(finding)
	if err != nil {
		t.Fatalf("failed to marshal ExportFinding: %v", err)
	}

	// Verify structure
	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal ExportFinding: %v", err)
	}

	if decoded["id"] != "test-001" {
		t.Errorf("unexpected ID: got %v", decoded["id"])
	}
	if decoded["severity"] != "critical" {
		t.Errorf("unexpected severity: got %v", decoded["severity"])
	}
}

func TestExportSummary(t *testing.T) {
	summary := ExportSummary{
		TotalFindings:    5,
		CriticalCount:    2,
		WarningCount:     2,
		InfoCount:        1,
		HealthPercentage: 75.5,
		IsValid:          true,
	}

	// Test JSON marshaling
	data, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("failed to marshal ExportSummary: %v", err)
	}

	// Verify counts
	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal ExportSummary: %v", err)
	}

	if decoded["total_findings"] != float64(5) {
		t.Errorf("unexpected total_findings: got %v", decoded["total_findings"])
	}
	if decoded["critical_count"] != float64(2) {
		t.Errorf("unexpected critical_count: got %v", decoded["critical_count"])
	}
}

func TestExportReport(t *testing.T) {
	report := ExportReport{
		Meta: ExportMeta{
			GeneratedAt: "2026-02-23T10:00:00Z",
			BundlePath:  "/test/bundle",
			BundleType:  "rke2",
			R8SVersion:  "v0.8.0",
		},
		Health: &bundle.HealthCheck{
			IsValid:      true,
			Completeness: 85.5,
		},
		Findings: []ExportFinding{
			{
				ID:       "test-001",
				Severity: "critical",
				Title:    "Test Finding",
			},
		},
		Summary: ExportSummary{
			TotalFindings: 1,
			CriticalCount: 1,
		},
	}

	// Test JSON marshaling
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal ExportReport: %v", err)
	}

	// Verify it's valid JSON and has expected structure
	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal ExportReport: %v", err)
	}

	requiredSections := []string{"meta", "health", "findings", "summary"}
	for _, section := range requiredSections {
		if _, ok := decoded[section]; !ok {
			t.Errorf("ExportReport missing section: %s", section)
		}
	}
}

func TestStandardizeFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"json", "json"},
		{"JSON", "json"},
		{"Json", "json"},
		{"yaml", "yaml"},
		{"YAML", "yaml"},
		{"yml", "yaml"},
		{"table", "table"},
		{"wide", "wide"},
		{"human", "human"},
		{"invalid", "table"}, // Default fallback
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := StandardizeFormat(tt.input)
			if got != tt.expected {
				t.Errorf("StandardizeFormat(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}