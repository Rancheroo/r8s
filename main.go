// Package main provides the entry point for r8s, a Rancher-focused log viewer and cluster
// simulator. It initializes version information and executes the root Cobra command.
package main

import (
	"fmt"
	"os"

	"github.com/Rancheroo/r8s/cmd"
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
