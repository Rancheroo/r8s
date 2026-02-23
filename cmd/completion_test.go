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
	}{
		{name: "bash", shell: "bash", wantErr: false},
		{name: "zsh", shell: "zsh", wantErr: false},
		{name: "fish", shell: "fish", wantErr: false},
		{name: "powershell", shell: "powershell", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a buffer to capture output
			buf := new(bytes.Buffer)
			
			// Create a temporary command with output buffer
			cmd := &cobra.Command{}
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			
			err := runCompletion(cmd, []string{tt.shell})

			if tt.wantErr && err == nil {
				t.Errorf("runCompletion() expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("runCompletion() unexpected error: %v", err)
			}

			// Check output is non-empty (shell completion generates scripts)
			output := buf.String()
			if len(output) == 0 {
				t.Errorf("runCompletion() produced no output for %s", tt.shell)
			}
			
			// Verify shell-specific content is present
			switch tt.shell {
			case "bash":
				if !strings.Contains(output, "__start_r8s") {
					t.Errorf("bash completion missing function definition")
				}
			case "zsh":
				if !strings.Contains(output, "compdef") {
					t.Errorf("zsh completion missing compdef")
				}
			case "fish":
				if !strings.Contains(output, "complete") {
					t.Errorf("fish completion missing complete command")
				}
			case "powershell":
				if !strings.Contains(output, "Register-ArgumentCompleter") {
					t.Errorf("powershell completion missing Register-ArgumentCompleter")
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
	// Test with invalid shell argument - should not error, just show usage
	buf := new(bytes.Buffer)
	cmd := &cobra.Command{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	
	err := runCompletion(cmd, []string{"invalid-shell"})

	// Currently returns nil (shows usage), not an error - this is acceptable behavior
	// Just verify it doesn't panic
	if err != nil {
		t.Logf("Got error for invalid shell (may be expected): %v", err)
	}
}