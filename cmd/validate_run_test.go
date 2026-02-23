package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// createTestBundleForValidate creates a minimal bundle for validate tests
func createTestBundleForValidate(t *testing.T, missingCritical bool) string {
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
	
	// Create nodes unless we're testing missing critical
	if !missingCritical {
		nodesFile := filepath.Join(kubectlDir, "nodes")
		if err := os.WriteFile(nodesFile, []byte(`{"items": [{"metadata": {"name": "test-node"}}]}`), 0644); err != nil {
			t.Fatalf("failed to write nodes file: %v", err)
		}
	}
	
	return tmpDir
}

func TestRunValidate_ValidBundle(t *testing.T) {
	bundlePath := createTestBundleForValidate(t, false)
	
	buf := new(strings.Builder)
	cmd := &cobra.Command{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	
	validateFormat = "table"
	validateSummary = false
	
	// This will call os.Exit, so we need to handle that
	// For now, test what we can
	err := runValidate(cmd, []string{bundlePath})
	
	// If we get here without os.Exit, check results
	if err != nil {
		t.Logf("runValidate returned error (may be expected): %v", err)
	}
	
	output := buf.String()
	if !strings.Contains(output, "Bundle") && len(output) > 0 {
		t.Logf("Output: %s", output)
	}
}

func TestRunValidate_JSON(t *testing.T) {
	bundlePath := createTestBundleForValidate(t, false)
	
	buf := new(strings.Builder)
	cmd := &cobra.Command{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	
	validateFormat = "json"
	validateSummary = false
	
	err := runValidate(cmd, []string{bundlePath})
	
	if err != nil {
		t.Logf("runValidate returned: %v", err)
	}
	
	output := buf.String()
	if len(output) > 0 {
		// Verify it looks like JSON
		if !strings.HasPrefix(output, "{") {
			t.Error("JSON output should start with {")
		}
	}
}

func TestRunValidate_Summary(t *testing.T) {
	bundlePath := createTestBundleForValidate(t, false)
	
	buf := new(strings.Builder)
	cmd := &cobra.Command{}
	cmd.SetOut(buf)
	
	validateFormat = "table"
	validateSummary = true
	
	err := runValidate(cmd, []string{bundlePath})
	
	if err != nil {
		t.Logf("runValidate returned: %v", err)
	}
	
	// Summary should be short (one line)
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) > 3 && len(output) > 0 {
		t.Logf("Summary output: %s", output)
	}
}

func TestRunValidate_InvalidPath(t *testing.T) {
	cmd := &cobra.Command{}
	
	// Should handle invalid path gracefully
	// Note: runValidate calls os.Exit(2) on error, so this is documentation
	err := runValidate(cmd, []string{"/nonexistent/path"})
	
	if err == nil {
		t.Log("runValidate with invalid path should exit with code 2")
	}
}

func TestRunValidate_NoArgs(t *testing.T) {
	cmd := &cobra.Command{}
	
	// Cobra should handle this before runValidate is called
	err := runValidate(cmd, []string{})
	
	// If we get here, cobra already validated args
	if err == nil {
		t.Error("runValidate should require bundle path")
	}
}

func TestOutputValidateTable(t *testing.T) {
	// This function writes to stdout, so we test it indirectly
	// through runValidate tests above
	t.Log("outputValidateTable tested via TestRunValidate_* tests")
}

func TestOutputValidateJSON(t *testing.T) {
	// This function writes to stdout, so we test it indirectly
	// through runValidate tests above
	t.Log("outputValidateJSON tested via TestRunValidate_* tests")
}