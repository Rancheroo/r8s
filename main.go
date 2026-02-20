// Package main provides the entry point for r8s, a Rancher-focused log viewer and cluster
// simulator. It initializes version information and executes the root Cobra command.
package main

import (
	"fmt"
	"os"

	"github.com/Rancheroo/r8s/cmd"
)

var (
	version = "0.8.0-alpha" // Version number
	commit  = "dev"   // Git commit (set via ldflags)
	date    = "now"   // Build date (set via ldflags)
)

func main() {
	cmd.SetVersionInfo(version, commit, date)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}