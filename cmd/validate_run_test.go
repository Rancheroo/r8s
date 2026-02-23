package cmd

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// createTestBundleForValidate creates a minimal bundle for validate tests
func createTestBundleForValidate(t *testing.T, withNodes bool) string {
	tmpDir := t.TempDir()

	rke2Dir := filepath.Join(tmpDir, "rke2")
	kubectlDir := filepath.Join(rke2Dir, "kubectl")

	if err := os.MkdirAll(kubectlDir, 0755); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}

	// Always create pods file
	podsFile := filepath.Join(kubectlDir, "pods")
	if err := os.WriteFile(podsFile, []byte(`{"items": [{"metadata": {"name": "test-pod"}}]}`), 0644); err != nil {
		t.Fatalf("failed to write pods file: %v", err)
	}

	// Create nodes if requested
	if withNodes {
		nodesFile := filepath.Join(kubectlDir, "nodes")
		if err := os.WriteFile(nodesFile, []byte(`{"items": [{"metadata": {"name": "test-node"}}]}`), 0644); err != nil {
			t.Fatalf("failed to write nodes file: %v", err)
		}
	}

	return tmpDir
}

func TestRunValidate_IncompleteBundle(t *testing.T) {
	// Bundle with pods + nodes is incomplete (not 100% complete) but usable
	bundlePath := createTestBundleForValidate(t, true)

	// Save and restore global flags
	origFormat, origSummary := validateFormat, validateSummary
	t.Cleanup(func() {
		validateFormat, validateSummary = origFormat, origSummary
	})

	buf := new(strings.Builder)
	cmd := &cobra.Command{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	validateFormat = "table"
	validateSummary = false

	err := runValidate(cmd, []string{bundlePath})

	// Incomplete bundle should return ExitIssuesFound (1)
	if err == nil {
		t.Fatal("runValidate() with incomplete bundle should return error")
	}

	var exitErr *ExitCodeError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected ExitCodeError, got %T: %v", err, err)
	}

	if exitErr.Code != ExitIssuesFound {
		t.Errorf("expected exit code %d (issues found), got %d", ExitIssuesFound, exitErr.Code)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Error("runValidate() produced no output")
	}
}

func TestRunValidate_JSON(t *testing.T) {
	bundlePath := createTestBundleForValidate(t, true)

	// Save and restore global flags
	origFormat, origSummary := validateFormat, validateSummary
	t.Cleanup(func() {
		validateFormat, validateSummary = origFormat, origSummary
	})

	buf := new(strings.Builder)
	cmd := &cobra.Command{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	validateFormat = "json"
	validateSummary = false

	err := runValidate(cmd, []string{bundlePath})

	// Incomplete bundle returns error (exit 1)
	if err == nil {
		t.Fatal("runValidate() with incomplete bundle should return error")
	}

	output := buf.String()
	if len(output) == 0 {
		t.Fatal("runValidate() produced no output")
	}

	// Verify it's valid JSON
	var health map[string]interface{}
	if err := json.Unmarshal([]byte(output), &health); err != nil {
		t.Errorf("output is not valid JSON: %v\nOutput: %s", err, output)
	}
}

func TestRunValidate_Summary(t *testing.T) {
	bundlePath := createTestBundleForValidate(t, true)

	// Save and restore global flags
	origFormat, origSummary := validateFormat, validateSummary
	t.Cleanup(func() {
		validateFormat, validateSummary = origFormat, origSummary
	})

	buf := new(strings.Builder)
	cmd := &cobra.Command{}
	cmd.SetOut(buf)

	// Use "summary" format (not "table" with validateSummary flag)
	validateFormat = "summary"
	validateSummary = false

	err := runValidate(cmd, []string{bundlePath})

	// Incomplete bundle returns error
	if err == nil {
		t.Fatal("runValidate() with incomplete bundle should return error")
	}

	// Summary should be short (typically one line with summary text)
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	// Summary mode should output a brief summary (usually 1-2 lines)
	if len(output) > 0 && len(lines) > 5 {
		t.Errorf("Summary output expected ≤5 lines, got %d:\n%s", len(lines), output)
	}
}

func TestRunValidate_InvalidPath(t *testing.T) {
	cmd := &cobra.Command{}

	err := runValidate(cmd, []string{"/nonexistent/path"})

	// Should return an ExitCodeError with code 2
	if err == nil {
		t.Fatal("runValidate() with invalid path should return error")
	}

	var exitErr *ExitCodeError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected ExitCodeError, got %T: %v", err, err)
	}

	if exitErr.Code != ExitError {
		t.Errorf("expected exit code %d, got %d", ExitError, exitErr.Code)
	}
}

func TestRunValidate_NoArgs(t *testing.T) {
	cmd := &cobra.Command{}

	err := runValidate(cmd, []string{})

	// Should return error for missing args
	if err == nil {
		t.Fatal("runValidate() with no args should return error")
	}

	var exitErr *ExitCodeError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected ExitCodeError, got %T: %v", err, err)
	}

	if exitErr.Code != ExitError {
		t.Errorf("expected exit code %d, got %d", ExitError, exitErr.Code)
	}
}

func TestRunValidate_WithNodes(t *testing.T) {
	// Bundle with pods + nodes is incomplete (exit 1) but not invalid (exit 2)
	bundlePath := createTestBundleForValidate(t, true)

	// Save and restore global flags
	origFormat, origSummary := validateFormat, validateSummary
	t.Cleanup(func() {
		validateFormat, validateSummary = origFormat, origSummary
	})

	buf := new(strings.Builder)
	cmd := &cobra.Command{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	validateFormat = "json"
	validateSummary = false

	err := runValidate(cmd, []string{bundlePath})

	// Bundle with nodes is incomplete but usable (exit 1)
	if err == nil {
		t.Fatal("runValidate() should return error for incomplete bundle")
	}

	var exitErr *ExitCodeError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected ExitCodeError, got %T: %v", err, err)
	}

	// Should be exit 1 (incomplete), not exit 2 (invalid)
	if exitErr.Code != ExitIssuesFound {
		t.Errorf("expected exit code %d (incomplete), got %d", ExitIssuesFound, exitErr.Code)
	}
}
