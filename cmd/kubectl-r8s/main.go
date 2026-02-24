// Package main implements the kubectl-r8s plugin
// Sprint 12: kubectl plugin wrapper for r8s
//
// Installation:
//   cp kubectl-r8s ~/.local/bin/
//   kubectl plugin list
//
// Usage:
//   export R8S_BUNDLE=./support-bundle.tar.gz
//   kubectl r8s get pods
//   kubectl r8s logs nginx-pod -n default
//   kubectl r8s describe node worker-1
//
// This plugin translates kubectl commands to r8s bundle analysis commands.

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	// Build r8s command
	// kubectl passes: [kubectl-r8s, get, pods, -n, namespace]
	// We need: r8s get pods ./bundle -n namespace
	args := os.Args[1:] // Skip program name

	// Detect subcommand (first non-flag token)
	subcommand := ""
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			subcommand = arg
			break
		}
	}

	// Check if this command needs a bundle
	needsBundle := commandsNeedingBundle()[subcommand]

	var bundlePath string
	if needsBundle {
		// Get bundle path from environment or auto-detect
		bundlePath = os.Getenv("R8S_BUNDLE")
		if bundlePath == "" {
			bundlePath = findBundle()
		}

		if bundlePath == "" {
			fmt.Fprintln(os.Stderr, "Error: No bundle found. Set R8S_BUNDLE or run in directory with bundle.")
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, "Examples:")
			fmt.Fprintln(os.Stderr, "  export R8S_BUNDLE=./support-bundle.tar.gz")
			fmt.Fprintln(os.Stderr, "  kubectl r8s get pods")
			os.Exit(1)
		}
	}

	// Build r8s args
	r8sArgs := []string{"r8s"}
	bundleInserted := false

	for _, arg := range args {
		if needsBundle && strings.HasPrefix(arg, "-") && !bundleInserted {
			// Found first flag, insert bundle before it
			r8sArgs = append(r8sArgs, bundlePath)
			bundleInserted = true
		}
		r8sArgs = append(r8sArgs, arg)
	}

	// If no flags found, append bundle at end
	if needsBundle && !bundleInserted {
		r8sArgs = append(r8sArgs, bundlePath)
	}

	// Find r8s binary
	r8sBinary := findR8SBinary()
	if r8sBinary == "" {
		fmt.Fprintln(os.Stderr, "Error: Cannot find r8s binary.")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Solutions:")
		fmt.Fprintln(os.Stderr, "  1. Place r8s binary in same directory as kubectl-r8s")
		fmt.Fprintln(os.Stderr, "  2. Run from directory containing ./r8s")
		fmt.Fprintln(os.Stderr, "  3. Set R8S_BINARY=/path/to/r8s")
		fmt.Fprintln(os.Stderr, "  4. Install r8s to PATH: cp r8s /usr/local/bin/")
		os.Exit(1)
	}

	// Execute r8s
	cmd := exec.Command(r8sBinary, r8sArgs[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// findR8SBinary looks for the r8s binary in common locations
// Priority: 1) R8S_BINARY env, 2) ./r8s in current dir, 3) Same dir as kubectl-r8s, 4) PATH
func findR8SBinary() string {
	// Check if R8S_BINARY env var is set
	if envPath := os.Getenv("R8S_BINARY"); envPath != "" {
		if _, err := os.Stat(envPath); err == nil {
			return envPath
		}
	}

	// Check current directory first (./r8s)
	if _, err := os.Stat("./r8s"); err == nil {
		return "./r8s"
	}

	// Look in same directory as kubectl-r8s
	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)
		r8sPath := filepath.Join(execDir, "r8s")
		if _, err := os.Stat(r8sPath); err == nil {
			return r8sPath
		}
	}

	// Try PATH (last resort - may find old version)
	if path, err := exec.LookPath("r8s"); err == nil {
		return path
	}

	return ""
}

// commandsNeedingBundle returns the set of commands that require a bundle path
func commandsNeedingBundle() map[string]bool {
	return map[string]bool{
		"analyze":      true,
		"ask":          true,
		"describe":     true,
		"export":       true,
		"get":          true,
		"logs":         true,
		"validate":     true,
		"test-cluster": true,
	}
}

// findBundle looks for bundle files in current directory
func findBundle() string {
	// Common bundle patterns
	patterns := []string{
		"*.tar.gz",
		"*.tgz",
		"support-bundle*",
		"rancher-bundle*",
		"r8s-bundle*",
		"*/support-bundle*",
	}

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err == nil && len(matches) > 0 {
			// Return first match
			return matches[0]
		}
	}

	// Look in common subdirectories
	dirs := []string{".", "bundle", "bundles", "support-bundle", "logs"}
	for _, dir := range dirs {
		if _, err := os.Stat(dir); err != nil {
			continue
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			name := entry.Name()
			if strings.Contains(name, "bundle") ||
			   strings.HasSuffix(name, ".tar.gz") ||
			   strings.HasSuffix(name, ".tgz") {
				return filepath.Join(dir, name)
			}
		}
	}

	return ""
}