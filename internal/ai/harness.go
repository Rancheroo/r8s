package ai

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Harness provides a framework for testing patterns against sample data
type Harness struct {
	analyzer *Analyzer
}

// NewHarness creates a new test harness
func NewHarness() *Harness {
	return &Harness{
		analyzer: NewAnalyzer(),
	}
}

// TestCase represents a single test scenario for the AI engine
type TestCase struct {
	Name             string   `json:"name"`
	Content          string   `json:"content"`
	ExpectedPatterns []string `json:"expected_patterns"`
	ExpectedSeverity string   `json:"expected_severity"`
}

// TestResult contains the outcome of a test case execution
type TestResult struct {
	Name    string          `json:"name"`
	Passed  bool            `json:"passed"`
	Matches []MatchResultV2 `json:"matches"`
	Errors  []string        `json:"errors"`
}

// Run executes a test case and validates the results
func (h *Harness) Run(tc TestCase) TestResult {
	// Use default options for testing
	result, err := h.analyzer.Analyze(tc.Content, AnalysisOptions{})

	testResult := TestResult{
		Name:    tc.Name,
		Matches: []MatchResultV2{},
		Errors:  []string{},
		Passed:  true,
	}

	if err != nil {
		testResult.Passed = false
		testResult.Errors = append(testResult.Errors, fmt.Sprintf("Analysis failed: %v", err))
		return testResult
	}

	// Filter only matched patterns
	for _, m := range result.Patterns {
		if m.Matched {
			testResult.Matches = append(testResult.Matches, m)
		}
	}

	// Check if expected patterns were found
	foundIDs := make(map[string]bool)
	for _, m := range testResult.Matches {
		foundIDs[m.PatternID] = true
	}

	for _, expectedID := range tc.ExpectedPatterns {
		if !foundIDs[expectedID] {
			testResult.Passed = false
			testResult.Errors = append(testResult.Errors, fmt.Sprintf("Expected pattern %s not found", expectedID))
		}
	}

	// Check for unexpected extra patterns (optional, but good for precision)
	if len(testResult.Matches) > len(tc.ExpectedPatterns) {
		testResult.Errors = append(testResult.Errors, fmt.Sprintf("Found %d matches, expected %d", len(testResult.Matches), len(tc.ExpectedPatterns)))
	}

	return testResult
}

// RunSuite executes multiple test cases
func (h *Harness) RunSuite(cases []TestCase) []TestResult {
	results := []TestResult{}
	for _, tc := range cases {
		results = append(results, h.Run(tc))
	}
	return results
}

// LoadTestData reads test cases from a JSON file
func LoadTestData(path string) ([]TestCase, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cases []TestCase
	if err := json.Unmarshal(data, &cases); err != nil {
		return nil, err
	}

	return cases, nil
}

// SaveTestData writes test cases to a JSON file
func SaveTestData(path string, cases []TestCase) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cases, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// GetDefaultTestCases returns a standard set of tests for builtin patterns
func GetDefaultTestCases() []TestCase {
	return []TestCase{
		{
			Name:             "OOM Kill in dmesg",
			Content:          " [123.456] Memory cgroup out of memory: Killed process 12345 (java)",
			ExpectedPatterns: []string{"oomkill-v2"},
			ExpectedSeverity: "critical",
		},
		{
			Name:             "Image Pull Backoff in Events",
			Content:          "Normal  BackOff    4m22s (x12 over 9m11s)  kubelet  Back-off pulling image \"nginx:latest\" (ImagePullBackOff)",
			ExpectedPatterns: []string{"imagepullbackoff-v2"},
			ExpectedSeverity: "warning",
		},
		{
			Name:             "CrashLoopBackOff in Pod Logs",
			Content:          "2024-02-17T12:00:00Z panic: application failed to connect to database. CrashLoopBackOff detected.",
			ExpectedPatterns: []string{"crashloopbackoff-v2"},
			ExpectedSeverity: "critical",
		},
	}
}
