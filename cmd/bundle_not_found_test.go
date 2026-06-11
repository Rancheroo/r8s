package cmd

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()

	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stderr pipe: %v", err)
	}
	os.Stderr = w

	fn()

	_ = w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		_ = r.Close()
		t.Fatalf("failed to read captured stderr: %v", err)
	}
	_ = r.Close()

	return buf.String()
}

func newMissingPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "does-not-exist")
}

func assertBundleNotFoundOutput(t *testing.T, output string) {
	t.Helper()
	if !strings.Contains(output, "Bundle not found:") {
		t.Fatalf("expected bundle-not-found message, got: %q", output)
	}
	if strings.Count(output, "Bundle not found:") != 1 {
		t.Fatalf("expected bundle-not-found message exactly once, got output: %q", output)
	}
	if strings.TrimSpace(output) == "" {
		t.Fatal("expected non-blank output for bundle-not-found path")
	}
	if strings.Contains(output, "✗ Error:") {
		t.Fatalf("did not expect friendly error wrapper for bundle-not-found path, got: %q", output)
	}
}

func assertExitCodeError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var exitErr *ExitCodeError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected ExitCodeError, got %T (%v)", err, err)
	}
	if exitErr.Code != ExitError {
		t.Fatalf("expected exit code %d, got %d", ExitError, exitErr.Code)
	}
}

func TestRunGet_BundleNotFound_ShowsSingleMessageAndExitCodeError(t *testing.T) {
	missingPath := newMissingPath(t)

	prevOutput := getOutput
	getOutput = "table"
	defer func() { getOutput = prevOutput }()

	var runErr error
	output := captureStderr(t, func() {
		runErr = runGet(nil, []string{"pods", missingPath})
	})

	assertExitCodeError(t, runErr)
	assertBundleNotFoundOutput(t, output)
}

func TestRunAnalyze_BundleNotFound_ShowsSingleMessageAndExitCodeError(t *testing.T) {
	missingPath := newMissingPath(t)

	var runErr error
	output := captureStderr(t, func() {
		runErr = runAnalyze(nil, []string{missingPath})
	})

	assertExitCodeError(t, runErr)
	assertBundleNotFoundOutput(t, output)
}

func TestRunAsk_BundleNotFound_ShowsSingleMessageAndExitCodeError(t *testing.T) {
	missingPath := newMissingPath(t)

	var runErr error
	output := captureStderr(t, func() {
		runErr = runAsk(nil, []string{missingPath, "test question"})
	})

	assertExitCodeError(t, runErr)
	assertBundleNotFoundOutput(t, output)
}

func TestRunExport_BundleNotFound_ShowsSingleMessageAndExitCodeError(t *testing.T) {
	missingPath := newMissingPath(t)

	var runErr error
	output := captureStderr(t, func() {
		runErr = runExport(nil, []string{missingPath})
	})

	assertExitCodeError(t, runErr)
	assertBundleNotFoundOutput(t, output)
}
