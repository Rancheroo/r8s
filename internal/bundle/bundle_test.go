package bundle

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBundle_Close(t *testing.T) {
	b := &Bundle{}
	err := b.Close()
	if err != nil {
		t.Errorf("Close() should be no-op, got error: %v", err)
	}
}

func TestBundle_ReadLogFile(t *testing.T) {
	// Create temp bundle with log file
	tmpDir, err := os.MkdirTemp("", "r8s-bundle-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a log file
	logContent := "test log content"
	logPath := filepath.Join(tmpDir, "test.log")
	os.WriteFile(logPath, []byte(logContent), 0644)

	b := &Bundle{}
	
	// Test with nil log file
	_, err = b.ReadLogFile(nil)
	if err == nil {
		t.Error("Expected error for nil log file")
	}

	// Test with valid log file
	logFile := &LogFileInfo{
		Path: logPath,
	}
	content, err := b.ReadLogFile(logFile)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if string(content) != logContent {
		t.Errorf("Expected %q, got: %q", logContent, string(content))
	}
}
