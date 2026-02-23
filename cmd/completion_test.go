package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRunCompletion(t *testing.T) {
	tests := []struct {
		name    string
		shell   string
		wantErr bool
		wantIn  []string // Substrings expected in output
	}{
		{
			name:    "bash completion",
			shell:   "bash",
			wantErr: false,
			wantIn:  []string{"r8s", "completion", "bash"},
		},
		{
			name:    "zsh completion",
			shell:   "zsh",
			wantErr: false,
			wantIn:  []string{"r8s", "completion", "zsh"},
		},
		{
			name:    "fish completion",
			shell:   "fish",
			wantErr: false,
			wantIn:  []string{"r8s", "completion", "fish"},
		},
		{
			name:    "powershell completion",
			shell:   "powershell",
			wantErr: false,
			wantIn:  []string{"r8s", "completion", "powershell"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a buffer to capture output
			buf := new(bytes.Buffer)
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)

			// Run completion
			cmd := &cobra.Command{}
			err := runCompletion(cmd, []string{tt.shell})

			if tt.wantErr && err == nil {
				t.Errorf("runCompletion() expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("runCompletion() unexpected error: %v", err)
			}

			// Check output contains expected strings
			output := buf.String()
			for _, want := range tt.wantIn {
				if !strings.Contains(output, want) {
					t.Errorf("runCompletion() output missing %q", want)
				}
			}
		})
	}
}

func TestCompletionCommand_Integration(t *testing.T) {
	// Test that completion command is registered
	if completionCmd == nil {
		t.Fatal("completionCmd should be registered")
	}

	// Verify command properties
	if completionCmd.Use != "completion [bash|zsh|fish|powershell]" {
		t.Errorf("unexpected Use: %s", completionCmd.Use)
	}

	// Verify valid args
	validArgs := completionCmd.ValidArgs
	expectedArgs := []string{"bash", "zsh", "fish", "powershell"}
	if len(validArgs) != len(expectedArgs) {
		t.Errorf("expected %d valid args, got %d", len(expectedArgs), len(validArgs))
	}

	for i, arg := range expectedArgs {
		if i >= len(validArgs) || validArgs[i] != arg {
			t.Errorf("expected valid arg %q at position %d", arg, i)
		}
	}
}

func TestCompletion_InvalidShell(t *testing.T) {
	// Test with invalid shell argument
	cmd := &cobra.Command{}
	err := runCompletion(cmd, []string{"invalid-shell"})

	// Should return usage error
	if err == nil {
		t.Error("expected error for invalid shell, got nil")
	}
}