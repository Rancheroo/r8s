// Package cmd implements error handling tests for r8s commands.
// Issue #86: Never Blank Output - Error handling tests
package cmd

import (
	"errors"
	"strings"
	"testing"
)

func TestNewUsageError(t *testing.T) {
	err := NewUsageError("ask", "r8s ask <bundle> <question>")

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !IsUsageError(err) {
		t.Error("expected IsUsageError to return true")
	}

	if !strings.Contains(err.Error(), "ask") {
		t.Errorf("error message should contain 'ask', got: %s", err.Error())
	}

	if !strings.Contains(err.Error(), "wrong argument order") {
		t.Errorf("error message should contain 'wrong argument order', got: %s", err.Error())
	}
}

func TestIsUsageError(t *testing.T) {
	// Test with UsageError
	usageErr := NewUsageError("test", "test usage")
	if !IsUsageError(usageErr) {
		t.Error("IsUsageError should return true for UsageError")
	}

	// Test with regular error
	regularErr := errors.New("regular error")
	if IsUsageError(regularErr) {
		t.Error("IsUsageError should return false for regular error")
	}

	// Test with nil
	if IsUsageError(nil) {
		t.Error("IsUsageError should return false for nil")
	}
}

func TestUsageError_Error(t *testing.T) {
	err := &UsageError{
		Command:      "analyze",
		CorrectUsage: "r8s analyze <bundle>",
	}

	expected := "wrong argument order for command 'analyze'"
	if err.Error() != expected {
		t.Errorf("expected '%s', got '%s'", expected, err.Error())
	}
}

// Test that error types exist and have proper structure
func TestErrorTypes(t *testing.T) {
	// Test ExitCodeError
	exitErr := NewExitError(ExitError, "test message")
	if exitErr == nil {
		t.Fatal("NewExitError should not return nil")
	}

	code := GetExitCode(exitErr)
	if code != ExitError {
		t.Errorf("expected exit code %d, got %d", ExitError, code)
	}

	// Test that nil returns ExitSuccess
	if GetExitCode(nil) != ExitSuccess {
		t.Error("GetExitCode(nil) should return ExitSuccess")
	}
}

// Test isKnownCommand helper
func TestIsKnownCommand(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected bool
	}{
		{"analyze", "analyze", true},
		{"ask", "ask", true},
		{"logs", "logs", true},
		{"unknown", "unknown", false},
		{"typo", "analize", true}, // typo that maps to analyze
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isKnownCommand(tt.command)
			if result != tt.expected {
				t.Errorf("isKnownCommand('%s') = %v, expected %v", tt.command, result, tt.expected)
			}
		})
	}
}

// Test isValidBundlePath helper
func TestIsValidBundlePath(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"./bundle/", true},
		{"/absolute/path", true},
		{"../parent", true},
		{"relative/path", true},
		{"somename", false},
		{"command", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isValidBundlePath(tt.path)
			if result != tt.expected {
				t.Errorf("isValidBundlePath('%s') = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}

// Test CommandSuggestions map entries
func TestCommandSuggestions(t *testing.T) {
	// Verify that common typos are present
	expectedTypos := []string{"analize", "analyse", "askk", "descibe", "gett", "logg", "validat"}

	for _, typo := range expectedTypos {
		if _, ok := CommandSuggestions[typo]; !ok {
			t.Errorf("CommandSuggestions missing typo: '%s'", typo)
		}
	}

	// Verify that suggestions have proper structure
	for typo, suggestion := range CommandSuggestions {
		if suggestion.Command == "" {
			t.Errorf("CommandSuggestions[%s] has empty Command", typo)
		}
		if suggestion.Description == "" {
			t.Errorf("CommandSuggestions[%s] has empty Description", typo)
		}
		if suggestion.Example == "" {
			t.Errorf("CommandSuggestions[%s] has empty Example", typo)
		}
	}
}

// Test AvailableCommands has entries
func TestAvailableCommands(t *testing.T) {
	if len(AvailableCommands) == 0 {
		t.Error("AvailableCommands should not be empty")
	}

	// Verify each command has required fields
	for _, cmd := range AvailableCommands {
		if cmd.Command == "" {
			t.Error("AvailableCommands entry has empty Command")
		}
		if cmd.Description == "" {
			t.Error("AvailableCommands entry has empty Description")
		}
	}
}
