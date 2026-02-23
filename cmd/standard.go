// Package cmd provides standardization utilities for CLI commands.
// Sprint 9 Day 5: Standardized output formats and exit codes.
package cmd

import (
	"fmt"
	"os"
	"strings"
)

// OutputFormat represents standardized output formats
type OutputFormat string

const (
	FormatJSON   OutputFormat = "json"
	FormatYAML   OutputFormat = "yaml"
	FormatTable  OutputFormat = "table"
	FormatWide   OutputFormat = "wide"
	FormatHuman  OutputFormat = "human"
)

// ValidFormats returns list of valid format strings
func ValidFormats() []string {
	return []string{"json", "yaml", "table", "wide", "human"}
}

// IsValidFormat checks if format string is valid
func IsValidFormat(format string) bool {
	for _, f := range ValidFormats() {
		if f == format {
			return true
		}
	}
	return false
}

// Exit codes for standardized CLI behavior
const (
	ExitSuccess       = 0 // No issues, command completed successfully
	ExitIssuesFound   = 1 // Issues found but command completed (e.g., bundle incomplete)
	ExitError         = 2 // Error occurred (invalid args, file not found, etc.)
	ExitCancelled     = 130 // Ctrl+C (SIGINT)
)

// HandleExit sets exit code and returns error message
func HandleExit(code int, format string, err error) {
	if err != nil {
		if format == "json" || format == "yaml" {
			// Structured error output
			fmt.Fprintf(os.Stderr, `{"error": "%s"}`, err.Error())
		} else {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		}
	}
	os.Exit(code)
}

// ExitCodeError is an error that carries a specific exit code.
// Used to communicate exit codes from RunE functions without calling os.Exit.
type ExitCodeError struct {
	Code    int
	Message string
}

func (e *ExitCodeError) Error() string {
	return e.Message
}

// NewExitError creates an error with a specific exit code
func NewExitError(code int, message string) error {
	return &ExitCodeError{Code: code, Message: message}
}

// GetExitCode extracts the exit code from an error.
// Returns ExitError (2) for regular errors, or the custom code for ExitCodeError.
func GetExitCode(err error) int {
	if err == nil {
		return ExitSuccess
	}
	if exitErr, ok := err.(*ExitCodeError); ok {
		return exitErr.Code
	}
	return ExitError
}

// FormatHelp generates standardized help text for format flag
func FormatHelp() string {
	return fmt.Sprintf("Output format: %s", strings.Join(ValidFormats(), ", "))
}

// StandardFlags provides common flag descriptions
type StandardFlags struct {
	Format   string
	Output   string
	Help     bool
	Verbose  bool
}

// StandardizeFormat normalizes format string
func StandardizeFormat(format string) string {
	format = strings.ToLower(strings.TrimSpace(format))
	switch format {
	case "json":
		return "json"
	case "yaml", "yml":
		return "yaml"
	case "table", "tabular":
		return "table"
	case "wide", "long":
		return "wide"
	case "summary":
		return "summary"
	case "human", "pretty", "default":
		return "human"
	default:
		return "human" // default
	}
}
