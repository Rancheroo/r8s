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

// createTestBundle creates a minimal bundle structure for testing
func createTestBundle(t *testing.T) string {
	tmpDir := t.TempDir()

	// Create RKE2 bundle structure
	rke2Dir := filepath.Join(tmpDir, "rke2")
	kubectlDir := filepath.Join(rke2Dir, "kubectl")

	if err := os.MkdirAll(kubectlDir, 0755); err != nil {
		t.Fatalf("failed to create bundle dirs: %v", err)
	}

	// Create minimal files
	files := map[string]string{
		filepath.Join(kubectlDir, "pods"):        `{"items": []}`,
		filepath.Join(kubectlDir, "nodes"):       `{"items": []}`,
		filepath.Join(rke2Dir, "server", "logs"): "test log content",
	}

	for path, content := range files {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create dir %s: %v", dir, err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file %s: %v", path, err)
		}
	}

	return tmpDir
}

func TestRunExport_JSON(t *testing.T) {
	bundlePath := createTestBundle(t)

	// Save and restore global flags
	origFormat, origOutput, origSeverity, origPattern := exportFormat, exportOutput, exportSeverity, exportPattern
	t.Cleanup(func() {
		exportFormat, exportOutput, exportSeverity, exportPattern = origFormat, origOutput, origSeverity, origPattern
	})

	// Capture output
	buf := new(strings.Builder)
	cmd := &cobra.Command{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// Set test flags
	exportFormat = "json"
	exportOutput = ""
	exportSeverity = "all"
	exportPattern = ""

	err := runExport(cmd, []string{bundlePath})

	// Valid bundle should return nil (exit 0)
	if err != nil {
		t.Errorf("runExport() unexpected error: %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Fatal("runExport() produced no output")
	}

	// Verify it's valid JSON
	var report ExportReport
	if err := json.Unmarshal([]byte(output), &report); err != nil {
		t.Errorf("output is not valid JSON: %v\nOutput: %s", err, output)
	}

	// Check structure
	if report.Meta.BundlePath != bundlePath {
		t.Errorf("expected bundle path %q, got %q", bundlePath, report.Meta.BundlePath)
	}
}

func TestRunExport_YAML(t *testing.T) {
	bundlePath := createTestBundle(t)

	// Save and restore global flags
	origFormat, origOutput := exportFormat, exportOutput
	t.Cleanup(func() {
		exportFormat, exportOutput = origFormat, origOutput
	})

	buf := new(strings.Builder)
	cmd := &cobra.Command{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	exportFormat = "yaml"
	exportOutput = ""

	err := runExport(cmd, []string{bundlePath})

	if err != nil {
		t.Errorf("runExport() unexpected error: %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Fatal("runExport() produced no output")
	}

	// Verify it's YAML (contains YAML markers)
	if !strings.Contains(output, "meta:") {
		t.Error("output doesn't look like YAML (missing 'meta:')")
	}
}

func TestRunExport_WithOutputFile(t *testing.T) {
	bundlePath := createTestBundle(t)
	outputPath := filepath.Join(t.TempDir(), "output.json")

	// Save and restore global flags
	origFormat, origOutput := exportFormat, exportOutput
	t.Cleanup(func() {
		exportFormat, exportOutput = origFormat, origOutput
	})

	cmd := &cobra.Command{}

	exportFormat = "json"
	exportOutput = outputPath

	err := runExport(cmd, []string{bundlePath})

	if err != nil {
		t.Errorf("runExport() unexpected error: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("output file was not created: %s", outputPath)
	}

	// Verify file content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	var report ExportReport
	if err := json.Unmarshal(content, &report); err != nil {
		t.Errorf("output file is not valid JSON: %v", err)
	}
}

func TestRunExport_InvalidBundle(t *testing.T) {
	invalidPath := "/nonexistent/path/to/bundle"

	cmd := &cobra.Command{}

	err := runExport(cmd, []string{invalidPath})

	// Should return an ExitCodeError with code 2
	if err == nil {
		t.Fatal("runExport() with invalid path should return error")
	}

	var exitErr *ExitCodeError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected ExitCodeError, got %T: %v", err, err)
	}

	if exitErr.Code != ExitError {
		t.Errorf("expected exit code %d, got %d", ExitError, exitErr.Code)
	}
}

func TestRunExport_NoArgs(t *testing.T) {
	cmd := &cobra.Command{}

	err := runExport(cmd, []string{})

	// Should return error for missing args
	if err == nil {
		t.Fatal("runExport() with no args should return error")
	}

	var exitErr *ExitCodeError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected ExitCodeError, got %T: %v", err, err)
	}

	if exitErr.Code != ExitError {
		t.Errorf("expected exit code %d, got %d", ExitError, exitErr.Code)
	}
}

func TestGetExitCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{"nil error", nil, ExitSuccess},
		{"regular error", errors.New("some error"), ExitError},
		{"exit success", NewExitError(ExitSuccess, "success"), ExitSuccess},
		{"exit issues", NewExitError(ExitIssuesFound, "issues"), ExitIssuesFound},
		{"exit error", NewExitError(ExitError, "error"), ExitError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetExitCode(tt.err)
			if got != tt.expected {
				t.Errorf("GetExitCode() = %d, want %d", got, tt.expected)
			}
		})
	}
}
