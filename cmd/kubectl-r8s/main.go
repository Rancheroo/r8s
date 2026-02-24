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
	// Get bundle path from environment or auto-detect
	bundlePath := os.Getenv("R8S_BUNDLE")
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

	// Build r8s command
	// kubectl passes: [kubectl-r8s, get, pods, -n, namespace]
	// We need: r8s get pods ./bundle -n namespace
	args := os.Args[1:] // Skip program name

	// Insert bundle path before flags
	r8sArgs := []string{"r8s"}
	bundleInserted := false

	for i, arg := range args {
		if strings.HasPrefix(arg, "-") && !bundleInserted {
			// Found first flag, insert bundle before it
			r8sArgs = append(r8sArgs, bundlePath)
			bundleInserted = true
		}
		r8sArgs = append(r8sArgs, arg)
	}

	// If no flags found, append bundle at end
	if !bundleInserted {
		r8sArgs = append(r8sArgs, bundlePath)
	}

	// Execute r8s
	cmd := exec.Command("r8s", r8sArgs[1:]...)
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