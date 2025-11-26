// Package main provides the entry point for r9s, a k9s-inspired terminal UI for managing
// Rancher-based Kubernetes clusters. It initializes version information and executes the
// root Cobra command.
package main

import (
	"fmt"
	"os"

	"github.com/4realtech/r9s/cmd"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.SetVersionInfo(version, commit, date)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
