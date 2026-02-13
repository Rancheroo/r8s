package bundle

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzeCompleteness_Complete(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create all required files
	requiredFiles := []string{
		"rke2/kubectl/pods",
		"rke2/kubectl/nodes",
		"rke2/kubectl/events",
	}
	
	for _, file := range requiredFiles {
		path := filepath.Join(tmpDir, file)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
	}

	result, err := AnalyzeCompleteness(tmpDir)
	if err != nil {
		t.Fatalf("AnalyzeCompleteness failed: %v", err)
	}

	if !result.IsComplete() {
		t.Error("Expected bundle to be complete")
	}

	if len(result.MissingRequired) > 0 {
		t.Errorf("Expected no missing required files, got: %v", result.MissingRequired)
	}
}

func TestAnalyzeCompleteness_Partial(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create only some files
	files := []string{
		"rke2/kubectl/pods",
		"rke2/kubectl/nodes",
		// Missing: rke2/kubectl/events
	}
	
	for _, file := range files {
		path := filepath.Join(tmpDir, file)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
	}

	result, err := AnalyzeCompleteness(tmpDir)
	if err != nil {
		t.Fatalf("AnalyzeCompleteness failed: %v", err)
	}

	if result.IsComplete() {
		t.Error("Expected bundle to be incomplete")
	}

	if len(result.MissingRequired) != 1 {
		t.Errorf("Expected 1 missing required file, got: %v", result.MissingRequired)
	}
}

func TestAnalyzeCompleteness_WithPodLogs(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create required files
	requiredFiles := []string{
		"rke2/kubectl/pods",
		"rke2/kubectl/nodes",
		"rke2/kubectl/events",
	}
	
	for _, file := range requiredFiles {
		path := filepath.Join(tmpDir, file)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
	}
	
	// Create podlogs directory with some files
	podlogsDir := filepath.Join(tmpDir, "rke2", "podlogs")
	if err := os.MkdirAll(podlogsDir, 0755); err != nil {
		t.Fatalf("Failed to create podlogs dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(podlogsDir, "test.log"), []byte("log"), 0644); err != nil {
		t.Fatalf("Failed to write log file: %v", err)
	}

	result, err := AnalyzeCompleteness(tmpDir)
	if err != nil {
		t.Fatalf("AnalyzeCompleteness failed: %v", err)
	}

	// Should have higher percentage with podlogs
	if result.Percentage < 50 {
		t.Errorf("Expected higher percentage with podlogs, got %d", result.Percentage)
	}
}

func TestCompletenessResult_GetStatus(t *testing.T) {
	tests := []struct {
		percentage int
		expected   string
	}{
		{100, "Complete"},
		{90, "Good"},
		{75, "Good"},
		{74, "Partial"},
		{50, "Partial"},
		{49, "Minimal"},
		{0, "Minimal"},
	}

	for _, test := range tests {
		result := &CompletenessResult{Percentage: test.percentage}
		status := result.GetStatus()
		if status != test.expected {
			t.Errorf("For percentage %d, expected status '%s', got '%s'",
				test.percentage, test.expected, status)
		}
	}
}

func TestFormatCompleteness(t *testing.T) {
	result := &CompletenessResult{
		Percentage: 85,
	}
	
	formatted := FormatCompleteness(result)
	expected := "Bundle: 85% (Good)"
	if formatted != expected {
		t.Errorf("Expected '%s', got '%s'", expected, formatted)
	}
}
