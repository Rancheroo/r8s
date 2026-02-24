package cmd

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
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

func TestRunExport_SARIF(t *testing.T) {
	bundlePath := createTestBundle(t)

	// Save and restore global flags
	origFormat, origOutput, origMinSev := exportFormat, exportOutput, exportMinSev
	t.Cleanup(func() {
		exportFormat, exportOutput, exportMinSev = origFormat, origOutput, origMinSev
	})

	// Set test flags - SARIF format
	exportFormat = "sarif"
	exportOutput = ""  // stdout
	exportMinSev = "info"

	cmd := &cobra.Command{}

	err := runExport(cmd, []string{bundlePath})

	// Valid bundle should return nil (exit 0)
	if err != nil {
		t.Errorf("runExport() unexpected error: %v", err)
	}

	// Note: Function outputs to os.Stdout directly, not to cmd.Out
	// Test passes if no error (output would go to test runner stdout)
}

func TestRunExport_Markdown(t *testing.T) {
	bundlePath := createTestBundle(t)

	// Save and restore global flags
	origFormat, origOutput := exportFormat, exportOutput
	t.Cleanup(func() {
		exportFormat, exportOutput = origFormat, origOutput
	})

	exportFormat = "markdown"
	exportOutput = ""

	cmd := &cobra.Command{}

	err := runExport(cmd, []string{bundlePath})

	if err != nil {
		t.Errorf("runExport() unexpected error: %v", err)
	}

	// Note: Function outputs to os.Stdout directly
	// Test passes if no error
}

func TestRunExport_WithOutputFile(t *testing.T) {
	bundlePath := createTestBundle(t)
	outputPath := filepath.Join(t.TempDir(), "output.sarif")

	// Save and restore global flags
	origFormat, origOutput := exportFormat, exportOutput
	t.Cleanup(func() {
		exportFormat, exportOutput = origFormat, origOutput
	})

	cmd := &cobra.Command{}

	exportFormat = "sarif"
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

	var report map[string]interface{}
	if err := json.Unmarshal(content, &report); err != nil {
		t.Errorf("output file is not valid JSON: %v", err)
	}
}

func TestRunExport_InvalidBundle(t *testing.T) {
	invalidPath := "/nonexistent/path/to/bundle"

	cmd := &cobra.Command{}

	err := runExport(cmd, []string{invalidPath})

	// Should return an error
	if err == nil {
		t.Fatal("runExport() with invalid path should return error")
	}
}

func TestRunExport_NoArgs(t *testing.T) {
	// The function accesses args[0] without checking length
	// This is expected to panic or error - we test the behavior
	cmd := &cobra.Command{}

	// Use defer/recover to catch panic
	defer func() {
		if r := recover(); r != nil {
			t.Logf("runExport() panicked with no args (expected): %v", r)
		}
	}()

	err := runExport(cmd, []string{})

	// If it doesn't panic, it should return an error
	if err == nil {
		t.Fatal("runExport() with no args should return error or panic")
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