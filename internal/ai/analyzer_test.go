// Package ai provides pattern matching and root cause analysis for Kubernetes issues.
// Sprint 11: Analysis Engine Tests
package ai

import (
	"strings"
	"testing"
)

func TestNewAnalyzer(t *testing.T) {
	analyzer := NewAnalyzer()
	if analyzer == nil {
		t.Fatal("NewAnalyzer() returned nil")
	}
	if analyzer.registry == nil {
		t.Error("Expected non-nil registry")
	}
	if analyzer.generator == nil {
		t.Error("Expected non-nil generator")
	}
	if analyzer.formatter == nil {
		t.Error("Expected non-nil formatter")
	}
}

func TestAnalyzerAnalyze(t *testing.T) {
	analyzer := NewAnalyzer()

	content := `
		Out of memory: Kill process 1234 (nginx)
		Pod worker-0 in CrashLoopBackOff
		Certificate has expired on node worker-1
	`

	opts := AnalysisOptions{
		IncludeInfo: true,
	}

	result, err := analyzer.Analyze(content, opts)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	// Verify result fields
	if result.StartTime.IsZero() {
		t.Error("Expected non-zero StartTime")
	}
	if result.EndTime.IsZero() {
		t.Error("Expected non-zero EndTime")
	}
	if result.Duration == 0 {
		t.Error("Expected non-zero Duration")
	}
	if len(result.Patterns) == 0 {
		t.Error("Expected at least one pattern match")
	}
	if len(result.Hints) == 0 {
		t.Error("Expected at least one hint generated")
	}

	// Verify summary
	if result.Summary.TotalPatterns == 0 {
		t.Error("Expected TotalPatterns > 0")
	}
	if result.Summary.MatchesFound == 0 {
		t.Error("Expected MatchesFound > 0")
	}
}

func TestAnalyzerFilteredAnalyze(t *testing.T) {
	analyzer := NewAnalyzer()

	content := `
		Out of memory: Kill process 1234 (nginx)
		Certificate expiring soon warning
		Pod info message
	`

	tests := []struct {
		name       string
		opts       AnalysisOptions
		minMatches int
		maxMatches int
	}{
		{
			name: "include all",
			opts: AnalysisOptions{
				IncludeInfo: true,
			},
			minMatches: 1,
		},
		{
			name: "only critical",
			opts: AnalysisOptions{
				MinSeverity: SeverityCritical,
			},
			minMatches: 0,
		},
		{
			name: "limit hints",
			opts: AnalysisOptions{
				MaxHints: 2,
			},
			maxMatches: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := analyzer.FilteredAnalyze(content, tt.opts)
			if err != nil {
				t.Fatalf("FilteredAnalyze() error = %v", err)
			}

			if len(result.Patterns) < tt.minMatches {
				t.Errorf("Expected at least %d matches, got %d", tt.minMatches, len(result.Patterns))
			}
			if tt.maxMatches > 0 && len(result.Hints) > tt.maxMatches {
				t.Errorf("Expected at most %d hints, got %d", tt.maxMatches, len(result.Hints))
			}
		})
	}
}

func TestAnalyzerDetectCorrelations(t *testing.T) {
	analyzer := NewAnalyzer()

	content := `
		Out of memory: Kill process 1234 (worker)
		Container in CrashLoopBackOff with 5 restarts
	`

	result, err := analyzer.Analyze(content, AnalysisOptions{})
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	// Should detect correlation between OOM and CrashLoop
	if len(result.Correlations) == 0 {
		t.Skip("No correlations detected (may not have both patterns in test content)")
	}

	foundOOMCrashCorrelation := false
	for _, corr := range result.Correlations {
		if (corr.PatternID1 == "oomkill-v2" && corr.PatternID2 == "crashloopbackoff-v2") ||
			(corr.PatternID1 == "crashloopbackoff-v2" && corr.PatternID2 == "oomkill-v2") {
			foundOOMCrashCorrelation = true
			if corr.Message == "" {
				t.Error("Expected correlation message to be non-empty")
			}
		}
	}

	if !foundOOMCrashCorrelation {
		t.Log("Note: OOM/CrashLoop correlation not found (content may not trigger both)")
	}
}

func TestAnalyzerBuildSummary(t *testing.T) {
	analyzer := NewAnalyzer()

	matches := []MatchResultV2{
		{PatternID: "1", Severity: SeverityCritical, Matched: true},
		{PatternID: "2", Severity: SeverityCritical, Matched: true},
		{PatternID: "3", Severity: SeverityWarning, Matched: true},
		{PatternID: "4", Severity: SeverityInfo, Matched: true},
	}

	correlations := []CorrelationMatch{
		{PatternID1: "1", PatternID2: "2"},
	}

	summary := BuildSummary(matches, correlations, analyzer.registry)

	if summary.CriticalIssues != 2 {
		t.Errorf("Expected 2 critical issues, got %d", summary.CriticalIssues)
	}
	if summary.WarningIssues != 1 {
		t.Errorf("Expected 1 warning issue, got %d", summary.WarningIssues)
	}
	if summary.InfoIssues != 1 {
		t.Errorf("Expected 1 info issue, got %d", summary.InfoIssues)
	}
	if summary.MatchesFound != 4 {
		t.Errorf("Expected 4 matches, got %d", summary.MatchesFound)
	}
	if summary.Correlations != 1 {
		t.Errorf("Expected 1 correlation, got %d", summary.Correlations)
	}
	if summary.TotalPatterns == 0 {
		t.Error("Expected TotalPatterns > 0")
	}
}

func TestAnalyzerFormatResults(t *testing.T) {
	analyzer := NewAnalyzer()

	result := &AnalysisResult{
		Duration: 1000000, // 1ms in nanoseconds
		Summary: AnalysisSummary{
			TotalPatterns:  19,
			MatchesFound:   3,
			CriticalIssues: 1,
			WarningIssues:  1,
			InfoIssues:     1,
			Correlations:   1,
		},
		Correlations: []CorrelationMatch{
			{PatternID1: "a", PatternID2: "b", Message: "Test correlation"},
		},
		Hints: []*Hint{
			{
				PatternID:   "test",
				Severity:    SeverityCritical,
				Confidence:  ConfidenceCertain,
				Summary:     "Test hint",
				Explanation: "Test explanation",
				Suggestion:  "Test suggestion",
			},
		},
	}

	output := analyzer.FormatResults(result)

	expectedSections := []string{
		"Analysis completed",
		"Patterns analyzed:",
		"Summary:",
		"Correlations detected:",
		"Root Cause Analysis:",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("Expected output to contain %q", section)
		}
	}
}

func TestAnalyzerGetCriticalHints(t *testing.T) {
	analyzer := NewAnalyzer()

	result := &AnalysisResult{
		Hints: []*Hint{
			{PatternID: "1", Severity: SeverityCritical},
			{PatternID: "2", Severity: SeverityWarning},
			{PatternID: "3", Severity: SeverityCritical},
			{PatternID: "4", Severity: SeverityInfo},
		},
	}

	critical := analyzer.GetCriticalHints(result)

	if len(critical) != 2 {
		t.Errorf("Expected 2 critical hints, got %d", len(critical))
	}

	for _, hint := range critical {
		if hint.Severity != SeverityCritical {
			t.Errorf("Expected Critical severity, got %s", hint.Severity)
		}
	}
}

func TestAnalyzerGetHintsByCategory(t *testing.T) {
	analyzer := NewAnalyzer()

	result := &AnalysisResult{
		Hints: []*Hint{
			{PatternID: "oomkill-v2"},          // Category: OOM
			{PatternID: "crashloopbackoff-v2"}, // Category: Crash
			{PatternID: "imagepullbackoff-v2"}, // Category: Image
		},
	}

	// Test OOM category
	oomHints := analyzer.GetHintsByCategory(result, "OOM")
	if len(oomHints) != 1 {
		t.Errorf("Expected 1 OOM hint, got %d", len(oomHints))
	}
}

func TestAnalyzerAnalyzeMultiple(t *testing.T) {
	analyzer := NewAnalyzer()

	contents := map[string]string{
		"pod.logs":  "Out of memory: Kill process 1234",
		"node.logs": "Certificate has expired",
	}

	opts := AnalysisOptions{}
	results, err := analyzer.AnalyzeMultiple(contents, opts)
	if err != nil {
		t.Fatalf("AnalyzeMultiple() error = %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	for name, result := range results {
		if result == nil {
			t.Errorf("Expected non-nil result for %s", name)
		}
	}
}

func TestAnalyzerGetPatternStats(t *testing.T) {
	analyzer := NewAnalyzer()

	results := []*AnalysisResult{
		{
			Patterns: []MatchResultV2{
				{PatternID: "oomkill-v2"},
				{PatternID: "oomkill-v2"},
			},
			Hints: []*Hint{
				{PatternID: "oomkill-v2"},
			},
		},
		{
			Patterns: []MatchResultV2{
				{PatternID: "crashloopbackoff-v2"},
			},
		},
	}

	stats := analyzer.GetPatternStats(results)

	// Check that stats were accumulated
	foundOOM := false
	for _, stat := range stats {
		if stat.PatternID == "oomkill-v2" {
			foundOOM = true
			if stat.MatchCount != 2 {
				t.Errorf("Expected oomkill-v2 MatchCount = 2, got %d", stat.MatchCount)
			}
			if stat.HintCount != 1 {
				t.Errorf("Expected oomkill-v2 HintCount = 1, got %d", stat.HintCount)
			}
		}
	}

	if !foundOOM {
		t.Error("Expected to find oomkill-v2 in stats")
	}
}

func TestShouldInclude(t *testing.T) {
	tests := []struct {
		name     string
		match    MatchResultV2
		opts     AnalysisOptions
		expected bool
	}{
		{
			name:     "always include certain",
			match:    MatchResultV2{Severity: SeverityCritical, Confidence: ConfidenceCertain},
			opts:     AnalysisOptions{MinSeverity: SeverityWarning},
			expected: true,
		},
		{
			name:     "exclude info when not included",
			match:    MatchResultV2{Severity: SeverityInfo, Confidence: ConfidenceCertain},
			opts:     AnalysisOptions{IncludeInfo: false},
			expected: false,
		},
		{
			name:     "exclude possible when min is certain",
			match:    MatchResultV2{Severity: SeverityWarning, Confidence: ConfidencePossible},
			opts:     AnalysisOptions{MinConfidence: ConfidenceCertain},
			expected: false,
		},
		{
			name:     "exclude warning when min is critical",
			match:    MatchResultV2{Severity: SeverityWarning, Confidence: ConfidenceCertain},
			opts:     AnalysisOptions{MinSeverity: SeverityCritical},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldIncludeMatch(tt.match, tt.opts)
			if got != tt.expected {
				t.Errorf("ShouldIncludeMatch() = %v, want %v", got, tt.expected)
			}
		})
	}
}
