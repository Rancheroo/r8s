package ai

import (
	"testing"
)

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
