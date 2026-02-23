// Package ai provides pattern matching and root cause analysis for Kubernetes issues.
// Sprint 11: Root Cause Hint System Tests
package ai

import (
	"strings"
	"testing"
)

func TestHintGeneratorGenerate(t *testing.T) {
	hg := NewHintGenerator()

	// Create a mock match result
	match := MatchResultV2{
		Matched:     true,
		PatternID:   "oomkill-v2",
		PatternName: "OOMKill Detected",
		Severity:    SeverityCritical,
		Confidence:  ConfidenceCertain,
		Message:     "[CRITICAL] OOMKill Detected: Container was killed due to memory limits",
		Evidence:    []string{"Out of memory: Kill process 1234 (nginx)"},
		Metadata: map[string]string{
			"ContainerName": "nginx",
			"PodName":       "web-0",
			"Namespace":     "default",
			"MemoryLimit":   "128Mi",
		},
	}

	// Get the pattern
	registry := NewRegistryV2()
	pattern, found := registry.GetByID("oomkill-v2")
	if !found {
		t.Fatal("oomkill-v2 pattern not found")
	}

	// Generate hint
	hint, err := hg.Generate(match, pattern)
	if err != nil {
		t.Fatalf("Failed to generate hint: %v", err)
	}

	// Verify hint fields
	if hint.PatternID != "oomkill-v2" {
		t.Errorf("Expected PatternID 'oomkill-v2', got %s", hint.PatternID)
	}
	if hint.Severity != SeverityCritical {
		t.Errorf("Expected Severity Critical, got %s", hint.Severity)
	}
	if hint.Confidence != ConfidenceCertain {
		t.Errorf("Expected Confidence Certain, got %s", hint.Confidence)
	}
	if hint.Summary == "" {
		t.Error("Expected non-empty Summary")
	}
	if hint.Suggestion == "" {
		t.Error("Expected non-empty Suggestion")
	}
	if hint.Command == "" {
		t.Error("Expected non-empty Command")
	}
	if len(hint.References) == 0 {
		t.Error("Expected at least one Reference")
	}

	// Check template substitution worked
	if !strings.Contains(hint.Summary, "nginx") {
		t.Errorf("Expected Summary to contain 'nginx' from template, got: %s", hint.Summary)
	}
	if !strings.Contains(hint.Summary, "web-0") {
		t.Errorf("Expected Summary to contain 'web-0' from template, got: %s", hint.Summary)
	}
}

func TestHintGeneratorApplyTemplate(t *testing.T) {
	hg := NewHintGenerator()

	tests := []struct {
		name     string
		tmpl     string
		data     map[string]string
		want     string
		wantErr  bool
	}{
		{
			name:    "simple substitution",
			tmpl:    "Container {{.Name}} in namespace {{.Namespace}}",
			data:    map[string]string{"Name": "nginx", "Namespace": "default"},
			want:    "Container nginx in namespace default",
			wantErr: false,
		},
		{
			name:    "no template variables",
			tmpl:    "Static message",
			data:    map[string]string{},
			want:    "Static message",
			wantErr: false,
		},
		{
			name:    "missing variable",
			tmpl:    "Container {{.Name}} in {{.Missing}}",
			data:    map[string]string{"Name": "nginx"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty template",
			tmpl:    "",
			data:    map[string]string{"Name": "nginx"},
			want:    "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := hg.applyTemplate(tt.tmpl, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("applyTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("applyTemplate() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHintGeneratorSeverityIcon(t *testing.T) {
	hg := NewHintGenerator()

	tests := []struct {
		severity Severity
		want     string
	}{
		{SeverityCritical, "🔴"},
		{SeverityWarning, "🟡"},
		{SeverityInfo, "🔵"},
		{"unknown", "⚪"},
	}

	for _, tt := range tests {
		t.Run(string(tt.severity), func(t *testing.T) {
			got := hg.severityIcon(tt.severity)
			if got != tt.want {
				t.Errorf("severityIcon(%s) = %s, want %s", tt.severity, got, tt.want)
			}
		})
	}
}

func TestHintGeneratorGenerateAll(t *testing.T) {
	hg := NewHintGenerator()
	registry := NewRegistryV2()

	// Simulate analyzing content with multiple patterns
	content := `
		Out of memory: Kill process 1234 (worker)
		Pull access denied for image 'myapp:latest'
	`

	matches := registry.AnalyzeV2(content)
	if len(matches) == 0 {
		t.Skip("No patterns matched in test content")
	}

	// Generate hints for all matches
	hints := hg.GenerateAll(matches, registry)

	if len(hints) == 0 {
		t.Fatal("Expected hints to be generated, got none")
	}

	// Verify each hint has required fields
	for _, hint := range hints {
		if hint.PatternID == "" {
			t.Error("Hint missing PatternID")
		}
		if hint.Summary == "" {
			t.Error("Hint missing Summary")
		}
		if hint.Suggestion == "" {
			t.Error("Hint missing Suggestion")
		}
	}
}

func TestHintFormatterFormatMarkdown(t *testing.T) {
	formatter := NewHintFormatter()

	hints := []*Hint{
		{
			PatternID:   "oomkill-v2",
			Severity:    SeverityCritical,
			Confidence:  ConfidenceCertain,
			Summary:     "Container nginx was killed",
			Explanation: "Container exceeded memory limit",
			Suggestion:  "Increase memory limit",
			Command:     "kubectl describe pod nginx",
			References:  []string{"https://example.com"},
		},
	}

	output := formatter.FormatMarkdown(hints)

	// Verify markdown structure
	expectedSections := []string{
		"# Root Cause Analysis Report",
		"## 1. Container nginx was killed",
		"**Confidence:** certain",
		"### Explanation",
		"### Suggestion",
		"### Command",
		"```bash",
		"### References",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("Expected markdown to contain %q", section)
		}
	}
}

func TestFilterHints(t *testing.T) {
	hints := []*Hint{
		{PatternID: "1", Severity: SeverityCritical},
		{PatternID: "2", Severity: SeverityWarning},
		{PatternID: "3", Severity: SeverityInfo},
		{PatternID: "4", Severity: SeverityCritical},
	}

	// Filter for Critical only
	critical := FilterHints(hints, SeverityCritical)
	if len(critical) != 2 {
		t.Errorf("Expected 2 critical hints, got %d", len(critical))
	}

	// Filter for Critical and Warning
	criticalAndWarning := FilterHints(hints, SeverityCritical, SeverityWarning)
	if len(criticalAndWarning) != 3 {
		t.Errorf("Expected 3 critical/warning hints, got %d", len(criticalAndWarning))
	}
}

func TestSortHintsBySeverity(t *testing.T) {
	hints := []*Hint{
		{PatternID: "1", Severity: SeverityInfo},
		{PatternID: "2", Severity: SeverityCritical},
		{PatternID: "3", Severity: SeverityWarning},
		{PatternID: "4", Severity: SeverityCritical},
	}

	sorted := SortHintsBySeverity(hints)

	// Verify order: Critical, Critical, Warning, Info
	expected := []Severity{SeverityCritical, SeverityCritical, SeverityWarning, SeverityInfo}
	for i, hint := range sorted {
		if hint.Severity != expected[i] {
			t.Errorf("Position %d: expected %s, got %s", i, expected[i], hint.Severity)
		}
	}
}

func TestHintGeneratorBuildExplanation(t *testing.T) {
	hg := NewHintGenerator()

	pattern := PatternV2{
		Description: "Test pattern description",
		Correlations: []Correlation{
			{PatternID: "related-1", Message: "Related to issue 1"},
			{PatternID: "related-2", Message: "Related to issue 2"},
		},
	}

	match := MatchResultV2{
		Evidence:   []string{"Evidence line 1", "Evidence line 2"},
		Correlated: []string{"related-1"},
		Metadata:   map[string]string{"Key": "Value"},
	}

	explanation := hg.buildExplanation(pattern, match)

	// Verify explanation contains expected sections
	expectedParts := []string{
		"Test pattern description",
		"Evidence:",
		"Evidence line 1",
		"Related Issues:",
		"Related to issue 1",
		"Context:",
		"Key: Value",
	}

	for _, part := range expectedParts {
		if !strings.Contains(explanation, part) {
			t.Errorf("Expected explanation to contain %q", part)
		}
	}
}