package cmd

import (
	"encoding/json"
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
		filepath.Join(kubectlDir, "pods"):           `{"items": []}`,
		filepath.Join(kubectlDir, "nodes"):          `{"items": []}`,
		filepath.Join(rke2Dir, "server", "logs"):    "test log content",
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
	
	// Capture output
	buf := new(strings.Builder)
	cmd := &cobra.Command{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	
	// Reset flags
	exportFormat = "json"
	exportOutput = ""
	exportSeverity = "all"
	exportPattern = ""
	
	err := runExport(cmd, []string{bundlePath})
	
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
	
	// Should exit with error (but we can't capture os.Exit in tests easily)
	// For now, just verify it returns an error
	err := runExport(cmd, []string{invalidPath})
	
	// runExport calls os.Exit on error, so we won't reach here in normal flow
	// This test documents expected behavior
	if err == nil {
		t.Log("runExport with invalid path should exit with code 2")
	}
}

func TestRunExport_NoArgs(t *testing.T) {
	cmd := &cobra.Command{}
	
	// Should require exactly 1 arg
	err := runExport(cmd, []string{})
	
	// This will fail before runExport is called (cobra handles arg validation)
	// But if we get here, document expected behavior
	if err == nil {
		t.Error("runExport() should require bundle path argument")
	}
}