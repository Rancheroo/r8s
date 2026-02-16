package bundle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateAndResolvePath(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "r8s-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name        string
		path        string
		verbose     bool
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid directory",
			path:    tmpDir,
			verbose: false,
			wantErr: false,
		},
		{
			name:        "empty path",
			path:        "",
			verbose:     false,
			wantErr:     true,
			errContains: "path is required",
		},
		{
			name:        "non-existent path",
			path:        filepath.Join(tmpDir, "nonexistent"),
			verbose:     false,
			wantErr:     true,
			errContains: "path not found",
		},
		{
			name:        "empty path with verbose",
			path:        "",
			verbose:     true,
			wantErr:     true,
			errContains: "USAGE:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			absPath, info, err := validateAndResolvePath(tt.path, tt.verbose)

			if tt.wantErr {
				if err == nil {
					t.Errorf("validateAndResolvePath() error = nil, wantErr = true")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("validateAndResolvePath() error = %v, want containing %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("validateAndResolvePath() unexpected error = %v", err)
				return
			}

			if absPath == "" {
				t.Error("validateAndResolvePath() returned empty absPath")
			}

			if info == nil {
				t.Error("validateAndResolvePath() returned nil info")
			}
		})
	}
}

func TestLoadFromPath(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir, err := os.MkdirTemp("", "r8s-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name        string
		setup       func() (string, func())
		opts        ImportOptions
		wantErr     bool
		errContains string
	}{
		{
			name: "file instead of directory",
			setup: func() (string, func()) {
				file := filepath.Join(tmpDir, "bundle.tar.gz")
				os.WriteFile(file, []byte("fake"), 0644)
				return file, func() { os.Remove(file) }
			},
			opts:        ImportOptions{},
			wantErr:     true,
			errContains: "not a directory",
		},
		{
			name: "file with verbose",
			setup: func() (string, func()) {
				file := filepath.Join(tmpDir, "bundle.tar.gz")
				os.WriteFile(file, []byte("fake"), 0644)
				return file, func() { os.Remove(file) }
			},
			opts:        ImportOptions{Verbose: true},
			wantErr:     true,
			errContains: "tar -xzf",
		},
		{
			name: "invalid bundle structure",
			setup: func() (string, func()) {
				dir := filepath.Join(tmpDir, "empty-bundle")
				os.MkdirAll(dir, 0755)
				return dir, func() { os.RemoveAll(dir) }
			},
			opts:        ImportOptions{},
			wantErr:     true,
			errContains: "not a valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, cleanup := tt.setup()
			defer cleanup()

			bundle, err := LoadFromPath(path, tt.opts)

			if tt.wantErr {
				if err == nil {
					t.Errorf("LoadFromPath() error = nil, wantErr = true")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("LoadFromPath() error = %v, want containing %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("LoadFromPath() unexpected error = %v", err)
				return
			}

			if bundle == nil {
				t.Error("LoadFromPath() returned nil bundle")
			}
		})
	}
}
