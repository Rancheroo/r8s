// Package cmd implements the CLI commands for r8s.
// Sprint 9 Day 2: r8s logs - Stream pod logs from bundles
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs [bundle-path] [pod-name]",
	Short: "Stream pod logs from a bundle",
	Long: `Stream logs from pods in a support bundle.

Similar to 'kubectl logs', but works offline with bundle data.

EXAMPLES:
  # Stream all logs from all pods
  r8s logs ./bundle/

  # Stream logs from specific pod
  r8s logs ./bundle/ rancher-7c4c7b8f5-x2v9p

  # Stream logs from namespace
  r8s logs ./bundle/ -n cattle-system

  # Stream logs and follow (like tail -f)
  r8s logs ./bundle/ rancher-xyz -f

  # Stream with timestamps
  r8s logs ./bundle/ --timestamps

  # Show last N lines
  r8s logs ./bundle/ --tail=100`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runLogs,
}

var (
	logsNamespace  string
	logsFollow     bool
	logsTimestamps bool
	logsTail       int
	logsContainer  string
)

func init() {
	rootCmd.AddCommand(logsCmd)

	logsCmd.Flags().StringVarP(&logsNamespace, "namespace", "n", "", "Filter by namespace")
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Follow log output (stream in real-time)")
	logsCmd.Flags().BoolVarP(&logsTimestamps, "timestamps", "t", false, "Show timestamps")
	logsCmd.Flags().IntVar(&logsTail, "tail", 0, "Show last N lines (0 = all)")
	logsCmd.Flags().StringVarP(&logsContainer, "container", "c", "", "Specific container (for multi-container pods)")
}

// LogEntry represents a single log line with metadata
type LogEntry struct {
	Timestamp   time.Time
	PodName     string
	Namespace   string
	Container   string
	Message     string
	HasError    bool
	HasWarning  bool
}

func runLogs(cmd *cobra.Command, args []string) error {
	bundlePath := args[0]
	podFilter := ""
	if len(args) > 1 {
		podFilter = args[1]
	}

	// Validate bundle exists
	if _, err := os.Stat(bundlePath); err != nil {
		return fmt.Errorf("bundle path not found: %w", err)
	}

	// Find log files
	logFiles, err := findLogFiles(bundlePath, logsNamespace, podFilter)
	if err != nil {
		return fmt.Errorf("failed to find logs: %w", err)
	}

	if len(logFiles) == 0 {
		fmt.Fprintf(os.Stderr, "No logs found")
		if logsNamespace != "" {
			fmt.Fprintf(os.Stderr, " in namespace '%s'", logsNamespace)
		}
		if podFilter != "" {
			fmt.Fprintf(os.Stderr, " for pod '%s'", podFilter)
		}
		fmt.Fprintln(os.Stderr)
		return nil
	}

	// Stream logs
	if logsFollow {
		return streamLogsFollow(logFiles)
	}

	return streamLogsOnce(logFiles)
}

// findLogFiles discovers log files in the bundle
func findLogFiles(bundlePath, namespaceFilter, podFilter string) ([]string, error) {
	var files []string

	// Check for podlogs directory
	podlogsDir := filepath.Join(bundlePath, "rke2", "podlogs")
	if _, err := os.Stat(podlogsDir); err != nil {
		// Try alternative paths
		podlogsDir = filepath.Join(bundlePath, "podlogs")
		if _, err := os.Stat(podlogsDir); err != nil {
			return files, nil // No logs directory
		}
	}

	// Walk podlogs directory
	err := filepath.Walk(podlogsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() {
			return nil
		}

		// Parse filename: namespace_podname_container.log
		filename := info.Name()
		parts := strings.Split(filename, "_")
		if len(parts) < 2 {
			return nil
		}

		ns := parts[0]
		pod := strings.Join(parts[1:len(parts)-1], "_")

		// Apply filters
		if namespaceFilter != "" && ns != namespaceFilter {
			return nil
		}
		if podFilter != "" && !strings.Contains(pod, podFilter) {
			return nil
		}

		files = append(files, path)
		return nil
	})

	return files, err
}

// streamLogsOnce outputs logs once (non-following)
func streamLogsOnce(files []string) error {
	for _, file := range files {
		if err := outputLogFile(file, false); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		}
	}
	return nil
}

// streamLogsFollow implements tail -f behavior
func streamLogsFollow(files []string) error {
	fmt.Println(color.YellowString("Following logs (Ctrl+C to stop)..."))
	fmt.Println()

	// Track file positions
	positions := make(map[string]int64)

	for {
		newData := false
		for _, file := range files {
			pos := positions[file]
			newPos, hasNew, err := tailFile(file, pos)
			if err != nil {
				continue
			}
			positions[file] = newPos
			if hasNew {
				newData = true
			}
		}

		if !newData {
			time.Sleep(500 * time.Millisecond)
		}
	}
}

// outputLogFile outputs a single log file
func outputLogFile(path string, isFollowing bool) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Get metadata from filename
	filename := filepath.Base(path)
	parts := strings.Split(filename, "_")
	namespace := parts[0]
	pod := strings.Join(parts[1:len(parts)-1], "_")
	container := strings.TrimSuffix(parts[len(parts)-1], ".log")

	// Print header if not following
	if !isFollowing {
		fmt.Printf("\n%s %s/%s [%s]\n",
			color.CyanString("==>"),
			color.YellowString(namespace),
			color.GreenString(pod),
			color.MagentaString(container),
		)
		fmt.Println(color.CyanString(strings.Repeat("-", 60)))
	}

	scanner := bufio.NewScanner(file)
	lineCount := 0
	maxLines := logsTail

	// If tail mode, read all lines first
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Apply tail filter
	startIdx := 0
	if maxLines > 0 && len(lines) > maxLines {
		startIdx = len(lines) - maxLines
	}

	// Output lines
	for i := startIdx; i < len(lines); i++ {
		line := lines[i]
		lineCount++

		// Colorize errors/warnings
		formattedLine := formatLogLine(line)
		fmt.Println(formattedLine)
	}

	return scanner.Err()
}

// tailFile reads new content from a file position
func tailFile(path string, startPos int64) (int64, bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return startPos, false, err
	}
	defer file.Close()

	// Seek to start position
	_, err = file.Seek(startPos, 0)
	if err != nil {
		return startPos, false, err
	}

	// Read new content
	scanner := bufio.NewScanner(file)
	hasNew := false
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(formatLogLine(line))
		hasNew = true
	}

	// Get current position
	pos, _ := file.Seek(0, 1)
	return pos, hasNew, scanner.Err()
}

// formatLogLine applies formatting to log lines
func formatLogLine(line string) string {
	upper := strings.ToUpper(line)

	// Error highlighting
	if strings.Contains(upper, "ERROR") || strings.Contains(upper, "FATAL") ||
		strings.Contains(upper, "PANIC") || strings.Contains(upper, "EXCEPTION") {
		return color.RedString(line)
	}

	// Warning highlighting
	if strings.Contains(upper, "WARN") || strings.Contains(upper, "WARNING") {
		return color.YellowString(line)
	}

	// Info highlighting
	if strings.Contains(upper, "INFO") {
		return color.CyanString(line)
	}

	return line
}
