package cmd

import (
	"testing"

	"github.com/Rancheroo/r8s/internal/bundle"
)

func TestValidateFlags(t *testing.T) {
	// Test that flags are properly initialized
	if validateCmd == nil {
		t.Fatal("validateCmd should be registered")
	}

	// Verify command properties
	if validateCmd.Use != "validate [bundle-path]" {
		t.Errorf("unexpected Use: %s", validateCmd.Use)
	}

	// Check required flags exist
	formatFlag := validateCmd.Flags().Lookup("format")
	if formatFlag == nil {
		t.Error("format flag should exist")
	}
	if formatFlag.DefValue != "table" {
		t.Errorf("unexpected format default: %s", formatFlag.DefValue)
	}

	summaryFlag := validateCmd.Flags().Lookup("summary")
	if summaryFlag == nil {
		t.Error("summary flag should exist")
	}
}

func TestValidateExitCodeConstants(t *testing.T) {
	// Verify exit codes are properly defined
	if ExitSuccess != 0 {
		t.Errorf("ExitSuccess should be 0, got %d", ExitSuccess)
	}
	if ExitIssuesFound != 1 {
		t.Errorf("ExitIssuesFound should be 1, got %d", ExitIssuesFound)
	}
	if ExitError != 2 {
		t.Errorf("ExitError should be 2, got %d", ExitError)
	}
}

func TestHealthCheckStatus(t *testing.T) {
	tests := []struct {
		name       string
		isValid    bool
		completeness float64
	}{
		{
			name:       "valid complete bundle",
			isValid:    true,
			completeness: 100.0,
		},
		{
			name:       "valid incomplete bundle",
			isValid:    true,
			completeness: 75.0,
		},
		{
			name:       "invalid bundle",
			isValid:    false,
			completeness: 50.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			health := &bundle.HealthCheck{
				IsValid:      tt.isValid,
				Completeness: tt.completeness,
			}

			if health.IsValid != tt.isValid {
				t.Errorf("IsValid mismatch: got %v, want %v", health.IsValid, tt.isValid)
			}
			if health.Completeness != tt.completeness {
				t.Errorf("Completeness mismatch: got %f, want %f", health.Completeness, tt.completeness)
			}
		})
	}
}