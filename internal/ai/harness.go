package ai

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Harness provides a framework for testing patterns against sample data
type Harness struct {
	engine *Engine
}

// NewHarness creates a new test harness
func NewHarness() *Harness {
	return &Harness{
		engine: NewEngine(),
	}
}

// TestCase represents a single test scenario for the AI engine
type TestCase struct {
	Name             string        `json:"name"`
	Content          string        `json:"content"`
	Metadata         MatchMetadata `json:"metadata"`
	ExpectedPatterns []string      `json:"expected_patterns"`
	ExpectedSeverity string        `json:"expected_severity"`
}

// TestResult contains the outcome of a test case execution
type TestResult struct {
	Name     string    `json:"name"`
	Passed   bool      `json:"passed"`
	Findings []Finding `json:"findings"`
	Errors   []string  `json:"errors"`
}

// Run executes a test case and validates the results
func (h *Harness) Run(tc TestCase) TestResult {
	findings := h.engine.Analyze(tc.Content, tc.Metadata)
	
	result := TestResult{
		Name:     tc.Name,
		Findings: findings,
		Errors:   []string{},
		Passed:   true,
	}

	// Check if expected patterns were found
	foundIDs := make(map[string]bool)
	for _, f := range findings {
		foundIDs[f.PatternID] = true
	}

	for _, expectedID := range tc.ExpectedPatterns {
		if !foundIDs[expectedID] {
			result.Passed = false
			result.Errors = append(result.Errors, fmt.Sprintf("Expected pattern %s not found", expectedID))
		}
	}

	// Check for unexpected extra patterns (optional, but good for precision)
	if len(findings) > len(tc.ExpectedPatterns) {
		result.Errors = append(result.Errors, fmt.Sprintf("Found %d findings, expected %d", len(findings), len(tc.ExpectedPatterns)))
	}

	return result
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
			Name:    "OOM Kill in dmesg",
			Content: " [123.456] oom-killer: ginkgo invoked oom-killer: gfp_mask=0x100cca(GFP_HIGHUSER_MOVABLE), order=0, oom_score_adj=0",
			Metadata: MatchMetadata{
				SourceType: "dmesg",
				NodeName:   "node-1",
			},
			ExpectedPatterns: []string{"oom-kill"},
			ExpectedSeverity: "critical",
		},
		{
			Name:    "Image Pull Backoff in Events",
			Content: "Normal  BackOff    4m22s (x12 over 9m11s)  kubelet  Back-off pulling image \"nginx:latest\" (ImagePullBackOff)",
			Metadata: MatchMetadata{
				SourceType:    "events",
				PodName:       "nginx-pod",
				Namespace:     "default",
				ContainerName: "nginx",
			},
			ExpectedPatterns: []string{"image-pull-backoff"},
			ExpectedSeverity: "high",
		},
		{
			Name:    "CrashLoopBackOff in Pod Logs",
			Content: "2024-02-17T12:00:00Z panic: application failed to connect to database. CrashLoopBackOff detected.",
			Metadata: MatchMetadata{
				SourceType:    "logs",
				PodName:       "db-app",
				Namespace:     "prod",
				ContainerName: "app",
			},
			ExpectedPatterns: []string{"crash-loop-backoff"},
			ExpectedSeverity: "high",
		},
	}
}
