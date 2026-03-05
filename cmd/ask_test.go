// Package cmd implements tests for the ask command.
// Issue #86: Never Blank Output - Test error message improvements
package cmd

import (
	"strings"
	"testing"

	"github.com/Rancheroo/r8s/internal/ai"
)

// TestIsLikelyQuestion verifies the heuristic for detecting natural language questions.
func TestIsLikelyQuestion(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"why is this happening?", true},
		{"what is the problem", true},
		{"which pods are crashing", true},
		{"show me errors", true},
		{"find the issue", true},
		{"list all pods", true},
		{"how to fix this", true},
		{"is the pod ready", true},
		{"are certificates expired", true},
		{"can you help", true},
		{"will this work", true},
		{"what?", true},
		{"./bundle/", false},
		{"/path/to/bundle", false},
		{"mypod", false},
		{"rancher", false},
		{"127.0.0.1", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isLikelyQuestion(tt.input)
			if result != tt.expected {
				t.Errorf("isLikelyQuestion('%s') = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestIsLikelyPath verifies the heuristic for detecting file paths.
func TestIsLikelyPath(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"./bundle/", true},
		{"../parent", true},
		{"/absolute/path", true},
		{"relative/path", true},
		{"path\\with\\backslash", true}, // Windows style
		{"why is this happening", false},
		{"what is wrong", false},
		{"show errors", false},
		{"mypod", false},
		{"rancher", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isLikelyPath(tt.input)
			if result != tt.expected {
				t.Errorf("isLikelyPath('%s') = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestParseQueryIntent verifies that natural language queries are correctly parsed into structured intents.
func TestParseQueryIntent(t *testing.T) {
	tests := []struct {
		name     string
		question string
		wantType string
		resource string
		cond     string
	}{
		{
			name:     "why crashing",
			question: "why is nginx-pod crashing?",
			wantType: "why",
			resource: "pod",
			cond:     "crashing",
		},
		{
			name:     "show imagepull",
			question: "show me imagepullbackoff issues",
			wantType: "show",
			resource: "image",
			cond:     "imagepull",
		},
		{
			name:     "which expired",
			question: "which certificates are expired?",
			wantType: "which",
			resource: "certificate",
			cond:     "expired",
		},
		{
			name:     "what wrong",
			question: "what is wrong with worker-1?",
			wantType: "what",
			resource: "", // "worker" doesn't match any known resource type
			cond:     "",
		},
		{
			name:     "unknown",
			question: "tell me something",
			wantType: "unknown",
			resource: "",
			cond:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			intent := parseQueryIntent(tt.question)
			if intent.Type != tt.wantType {
				t.Errorf("parseQueryIntent('%s').Type = %s, expected %s", tt.question, intent.Type, tt.wantType)
			}
			if intent.Resource != tt.resource {
				t.Errorf("parseQueryIntent('%s').Resource = %s, expected %s", tt.question, intent.Resource, tt.resource)
			}
			if intent.Condition != tt.cond {
				t.Errorf("parseQueryIntent('%s').Condition = %s, expected %s", tt.question, intent.Condition, tt.cond)
			}
		})
	}
}

// TestFormatNoResultsResponse verifies the output format when no issues are found.
func TestFormatNoResultsResponse(t *testing.T) {
	intent := QueryIntent{
		Type:      "why",
		Resource:  "pod",
		Condition: "crashing",
	}

	response := formatNoResultsResponse(intent)

	// Check that response contains expected elements
	if !strings.Contains(response, "crashing") {
		t.Error("response should mention condition (crashing)")
	}
	if !strings.Contains(response, "pod") {
		t.Error("response should mention resource (pod)")
	}
	if !strings.Contains(response, "No") {
		t.Error("response should mention No issues found")
	}
	if !strings.Contains(response, "r8s analyze") {
		t.Error("response should suggest using r8s analyze")
	}
}

// TestFormatUnknownResponse verifies the help message returned for unknown queries.
func TestFormatUnknownResponse(t *testing.T) {
	response := formatUnknownResponse()

	// Check that response contains help text
	if !strings.Contains(response, "didn't understand") {
		t.Error("response should indicate understanding failure")
	}
	if !strings.Contains(response, "r8s analyze") {
		t.Error("response should suggest r8s analyze")
	}
	if !strings.Contains(response, "r8s export") {
		t.Error("response should suggest r8s export")
	}
}

// TestMatchesIntent verifies the logic for matching AI hints against query intents.
func TestMatchesIntent(t *testing.T) {
	// Create a test hint
	hint := &ai.Hint{
		PatternID: "crashloop_backoff",
		Metadata:  map[string]string{"PodName": "test-pod"},
	}

	tests := []struct {
		name     string
		intent   QueryIntent
		expected bool
	}{
		{
			name:     "match crashing condition",
			intent:   QueryIntent{Condition: "crashing"},
			expected: true,
		},
		{
			name:     "match pod resource",
			intent:   QueryIntent{Resource: "pod", Condition: "crashing"},
			expected: true,
		},
		{
			// When no condition matches and no resource specified, matchesIntent returns true
			// This is the current behavior - let the caller filter more specifically
			name:     "no resource specified returns true",
			intent:   QueryIntent{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesIntent(hint, tt.intent)
			if result != tt.expected {
				t.Errorf("matchesIntent() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestResponseFormatters ensures that all response formatting functions produce non-empty output.
func TestResponseFormatters(t *testing.T) {
	hints := []*ai.Hint{
		{
			PatternID:   "test_pattern",
			Summary:     "Test summary",
			Explanation: "Test explanation",
			Suggestion:  "Test suggestion",
			Command:     "kubectl get pods",
		},
	}

	intent := QueryIntent{
		Type:      "why",
		Resource:  "pod",
		Condition: "crashing",
	}

	// Test that all formatters produce output
	if formatWhyResponse(intent, hints) == "" {
		t.Error("formatWhyResponse should produce output")
	}
	if formatShowResponse(intent, hints) == "" {
		t.Error("formatShowResponse should produce output")
	}
	if formatWhichResponse(intent, hints) == "" {
		t.Error("formatWhichResponse should produce output")
	}
	// formatWhatResponse takes intent and list of hints
	singleHintIntent := QueryIntent{Type: "what", Resource: "pod"}
	if formatWhatResponse(singleHintIntent, hints) == "" {
		t.Error("formatWhatResponse should produce output")
	}
	if formatGeneralResponse(hints) == "" {
		t.Error("formatGeneralResponse should produce output")
	}
}
