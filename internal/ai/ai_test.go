package ai

import (
	"testing"
)

func TestEngine_Analyze(t *testing.T) {
	engine := NewEngine()
	
	tests := []struct {
		name     string
		content  string
		metadata MatchMetadata
		expected string // pattern ID
	}{
		{
			name:    "OOM Match",
			content: "Memory cgroup out of memory: Killed process 1234 (java)",
			metadata: MatchMetadata{SourceType: "dmesg"},
			expected: "oomkill",
		},
		{
			name:    "Image Pull Match",
			content: "Failed to pull image: rpc error: code = Unknown desc = Error response from daemon",
			metadata: MatchMetadata{SourceType: "events"},
			expected: "imagepullbackoff",
		},
		{
			name:    "Crash Loop Match",
			content: "Back-off restarting failed container",
			metadata: MatchMetadata{SourceType: "logs"},
			expected: "crashloopbackoff",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			findings := engine.Analyze(tt.content, tt.metadata)
			
			found := false
			for _, f := range findings {
				if f.PatternID == tt.expected {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("expected pattern %s not found in %v", tt.expected, findings)
			}
		})
	}
}

func TestHarness_RunSuite(t *testing.T) {
	harness := NewHarness()
	cases := GetDefaultTestCases()
	results := harness.RunSuite(cases)

	if len(results) != len(cases) {
		t.Errorf("expected %d results, got %d", len(cases), len(results))
	}

	for _, r := range results {
		if !r.Passed {
			t.Errorf("test case %s failed: %v", r.Name, r.Errors)
		}
	}
}

func TestIsHigherSeverity(t *testing.T) {
	if !IsHigherSeverity("critical", "warning") {
		t.Error("critical should be higher than warning")
	}
	if IsHigherSeverity("info", "warning") {
		t.Error("info should not be higher than warning")
	}
	if IsHigherSeverity("warning", "warning") {
		t.Error("warning should not be higher than warning")
	}
}
