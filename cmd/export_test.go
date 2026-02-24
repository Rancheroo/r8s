package cmd

import (
	"encoding/json"
	"testing"
)

func TestSARIFOutput(t *testing.T) {
	// Test SARIF structure that actually exists
	sarif := SARIFLog{
		Version: "2.1.0",
		Runs: []SARIFRun{
			{
				Tool: SARIFTool{
					Driver: SARIFDriver{
						Name: "r8s",
						Rules: []SARIFRule{
							{
								ID:   "test-rule",
								Name: "Test Rule",
							},
						},
					},
				},
				Results: []SARIFResult{
					{
						RuleID:  "test-rule",
						Level:   "error",
						Message: SARIFMessage{Text: "Test message"},
					},
				},
			},
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(sarif)
	if err != nil {
		t.Fatalf("failed to marshal SARIF: %v", err)
	}

	// Verify it's valid JSON
	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal SARIF: %v", err)
	}

	// Check version
	if decoded["version"] != "2.1.0" {
		t.Errorf("unexpected version: got %v", decoded["version"])
	}
}