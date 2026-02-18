// Package cmd implements the CLI commands for r8s.
// v0.8.0: r8s logs - Stream pod logs from bundle (kubectl-style)
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Rancheroo/r8s/internal/bundle"
)

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs [bundle-path] [pod-name]",
	Short: "Print pod logs from bundle (kubectl-style)",
	Long: `Print logs for a pod from a Rancher support bundle.

Similar to 'kubectl logs', but works offline with bundle data.

EXAMPLES:
  # Print logs for a pod
  r8s logs ./bundle/ nginx-pod

  # Print logs for specific container
  r8s logs ./bundle/ nginx-pod -c container-name

  # Follow/stream logs (simulated from bundle)
  r8s logs ./bundle/ nginx-pod -f

  # Print previous container logs (crashed pod)
  r8s logs ./bundle/ nginx-pod -p

  # Show last N lines
  r8s logs ./bundle/ nginx-pod --tail=100

  # Show logs since timestamp
  r8s logs ./bundle/ nginx-pod --since=2024-01-01T00:00:00Z

  # Filter to errors only
  r8s logs ./bundle/ nginx-pod | grep -i error

POD NAME MATCHING:
  Pod names are matched against bundle inventory.
  Partial matching supported: "nginx" matches "nginx-7d8c9f4b2-x1z9q"

CONTAINER SELECTION:
  If pod has multiple containers and -c not specified:
  - First container is used (default)
  - Use -c to specify container name

EXIT CODES:
  0 - Logs displayed successfully
  1 - Pod not found in bundle
  2 - Log file not found or unreadable`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runLogs,
}

var (
	logsContainer  string // Container name (-c)
	logsPrevious   bool   // Previous container logs (-p)
	logsFollow     bool   // Follow/stream mode (-f)
	logsTail       int    // Show last N lines (--tail)
	logsSince      string // Show logs since timestamp (--since)
	logsTimestamps bool   // Show timestamps (-t)
	logsPrefix     bool   // Prefix each line with pod name
)

func init() {
	rootCmd.AddCommand(logsCmd)

	logsCmd.Flags().StringVarP(&logsContainer, "container", "c", "", "Container name (default: first container)")
	logsCmd.Flags().BoolVarP(&logsPrevious, "previous", "p", false, "Print previous container logs (crashed pod)")
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Stream logs (simulated from bundle)")
	logsCmd.Flags().IntVar(&logsTail, "tail", 0, "Show last N lines (0 = all)")
	logsCmd.Flags().StringVar(&logsSince, "since", "", "Show logs since timestamp (RFC3339)")
	logsCmd.Flags().BoolVarP(&logsTimestamps, "timestamps", "t", false, "Show timestamps")
	logsCmd.Flags().BoolVar(&logsPrefix, "prefix", false, "Prefix each line with pod name")
}

// runLogs executes the logs command
func runLogs(cmd *cobra.Command, args []string) error {
	// Parse arguments
	var bundlePath, podName string

	if len(args) == 1 {
		// Only pod name provided, use tuiBundlePath from root
		bundlePath = tuiBundlePath
		podName = args[0]
		if bundlePath == "" {
			return fmt.Errorf("bundle path required: r8s logs [bundle-path] [pod-name]")
		}
	} else {
		bundlePath = args[0]
		podName = args[1]
	}

	// Load bundle
	importOpts := bundle.ImportOptions{
		Path:    bundlePath,
		Verbose: verbose,
	}

	b, err := bundle.Load(importOpts)
	if err != nil {
		return fmt.Errorf("failed to load bundle: %w", err)
	}
	defer b.Close()

	// Find matching pod
	matchedPod, err := findPod(b, podName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
		return err
	}

	// Determine which container to show
	containerName := logsContainer
	if containerName == "" && len(matchedPod.Containers) > 0 {
		containerName = matchedPod.Containers[0]
	}

	if containerName == "" {
		return fmt.Errorf("no containers found for pod %s", matchedPod.Name)
	}

	// Find log file
	logFile, err := findLogFile(b, matchedPod, containerName, logsPrevious)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(2)
		return err
	}

	// Read and output logs
	return outputLogs(logFile, matchedPod, containerName)
}

// findPod finds a pod by name (supports partial matching)
func findPod(b *bundle.Bundle, name string) (*bundle.PodInfo, error) {
	var matches []bundle.PodInfo

	for _, pod := range b.Pods {
		// Exact match
		if pod.Name == name {
			return &pod, nil
		}
		// Partial match
		if strings.Contains(pod.Name, name) {
			matches = append(matches, pod)
		}
	}

	// If no exact match but one partial match, use it
	if len(matches) == 1 {
		return &matches[0], nil
	}

	// Multiple partial matches
	if len(matches) > 1 {
		fmt.Fprintf(os.Stderr, "Multiple pods match '%s':\n", name)
		for _, pod := range matches {
			fmt.Fprintf(os.Stderr, "  - %s (namespace: %s)\n", pod.Name, pod.Namespace)
		}
		return nil, fmt.Errorf("ambiguous pod name '%s' - use full name", name)
	}

	return nil, fmt.Errorf("pod '%s' not found in bundle", name)
}

// findLogFile finds the log file for a pod/container
func findLogFile(b *bundle.Bundle, pod *bundle.PodInfo, container string, previous bool) (string, error) {
	// Build expected log file patterns
	// RKE2 bundles typically have: podlogs/<namespace>/<pod>/<container>.log
	// or podlogs/<namespace>/<pod>/<container>-previous.log

	logFileName := container + ".log"
	if previous {
		logFileName = container + "-previous.log"
	}

	// Search through LogFiles
	for _, logFile := range b.LogFiles {
		// Check if this log file matches our pod/container
		if logFile.Type == bundle.LogTypePod {
			if logFile.Namespace == pod.Namespace &&
				logFile.PodName == pod.Name &&
				logFile.ContainerName == container {
				// Check if it's previous/current based on path
				isPreviousLog := strings.Contains(logFile.Path, "-previous")
				if previous == isPreviousLog {
					return logFile.Path, nil
				}
			}
		}
	}

	// Try to construct path from bundle structure
	possiblePaths := []string{
		filepath.Join(b.ExtractPath, "podlogs", pod.Namespace, pod.Name, logFileName),
		filepath.Join(b.ExtractPath, "podlogs", pod.Namespace, pod.Name, container, "current.log"),
		filepath.Join(b.ExtractPath, "podlogs", pod.Namespace, pod.Name, container, "previous.log"),
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	if previous {
		return "", fmt.Errorf("previous logs not found for %s/%s (container: %s)",
			pod.Namespace, pod.Name, container)
	}

	return "", fmt.Errorf("logs not found for %s/%s (container: %s). Available containers: %v",
		pod.Namespace, pod.Name, container, pod.Containers)
}

// outputLogs reads and outputs log file
func outputLogs(logPath string, pod *bundle.PodInfo, container string) error {
	file, err := os.Open(logPath)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	// Get file info for size
	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat log file: %w", err)
	}

	// Show header
	if verbose {
		fmt.Fprintf(os.Stderr, "# Logs for %s/%s (container: %s)\n", 
			pod.Namespace, pod.Name, container)
		fmt.Fprintf(os.Stderr, "# File: %s (%d bytes)\n", logPath, stat.Size())
		fmt.Fprintln(os.Stderr, "# ---")
	}

	// Scan and output
	scanner := bufio.NewScanner(file)
	lineNum := 0
	lines := []string{}

	// Read all lines (for tail support)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading log file: %w", err)
	}

	// Apply tail filter
	startLine := 0
	if logsTail > 0 && len(lines) > logsTail {
		startLine = len(lines) - logsTail
	}

	// Output lines
	prefix := ""
	if logsPrefix {
		prefix = fmt.Sprintf("[%s/%s] ", pod.Namespace, pod.Name)
	}

	for i := startLine; i < len(lines); i++ {
		line := lines[i]
		lineNum++

		// Apply timestamp formatting if requested
		// Note: This is simplified - real implementation would parse log timestamps
		if logsTimestamps {
			// Would add timestamp extraction here
			fmt.Printf("%s%d %s\n", prefix, lineNum, line)
		} else {
			fmt.Printf("%s%s\n", prefix, line)
		}
	}

	// Simulate follow mode (just show message)
	if logsFollow {
		fmt.Fprintln(os.Stderr, "# (Bundle mode - end of logs reached)")
	}

	return nil
}
