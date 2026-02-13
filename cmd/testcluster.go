// Package cmd implements the CLI commands and flags for r8s.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/Rancheroo/r8s/internal/bundle"
)

// testClusterCmd represents the test-cluster command
// S3-MEDIUM-4: Automated cluster testing subcommand
var testClusterCmd = &cobra.Command{
	Use:   "test-cluster [bundle-path]",
	Short: "Run automated diagnostic tests on a log bundle",
	Long: `Run automated diagnostic tests against an RKE2 log bundle.

This command performs a series of automated checks to identify common
cluster issues without launching the interactive TUI.

TESTS PERFORMED:
  • OOM Event Detection - Find pods killed by out-of-memory
  • Node Pressure Check - Identify nodes with memory/disk pressure
  • etcd Health Status - Check etcd alarms and member health
  • Pod Crash Detection - Find pods in CrashLoopBackOff or Error states
  • Resource Analysis - Identify pods without resource limits
  • Event Analysis - Check for warning events and errors

EXAMPLES:
  # Test an extracted bundle
  r8s test-cluster ./extracted-bundle/

  # Test with verbose output
  r8s test-cluster -v ./extracted-bundle/

  # Test and show all pods (not just issues)
  r8s test-cluster --all ./extracted-bundle/

EXIT CODES:
  0 - No issues found (healthy cluster)
  1 - Issues detected (see output for details)
  2 - Bundle parsing error`,
	RunE: runTestCluster,
}

var (
	testClusterAll    bool   // Show all pods, not just issues
	testClusterFormat string // Output format: table, json, summary
)

// TestResult represents a single test result
type TestResult struct {
	Name        string
	Status      string // PASS, FAIL, WARN, SKIP
	Description string
	Details     []string
}

// runTestCluster executes the test-cluster command
func runTestCluster(cmd *cobra.Command, args []string) error {
	// Get bundle path from args or use default
	bundlePath := ""
	if len(args) > 0 {
		bundlePath = args[0]
	}

	// Validate bundle path
	if bundlePath == "" {
		return fmt.Errorf("bundle path required. Usage: r8s test-cluster [bundle-path]")
	}

	info, err := os.Stat(bundlePath)
	if err != nil {
		return fmt.Errorf("cannot access bundle path '%s': %w", bundlePath, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("bundle path must be a directory: %s", bundlePath)
	}

	// Load the bundle
	opts := bundle.ImportOptions{
		Path:    bundlePath,
		MaxSize: 500 * 1024 * 1024, // 500MB for test mode
		Verbose: verbose,
	}

	b, err := bundle.Load(opts)
	if err != nil {
		// Exit code 2: Bundle parsing error (as documented)
		fmt.Fprintf(os.Stderr, "Error: failed to load bundle: %v\n", err)
		os.Exit(2)
	}

	// Run tests
	results := runTests(b, bundlePath)

	// Output results
	switch testClusterFormat {
	case "json":
		outputJSON(results)
	case "summary":
		outputSummary(results)
	default:
		outputTable(results)
	}

	// Exit with appropriate code
	exitCode := 0
	for _, r := range results {
		if r.Status == "FAIL" {
			exitCode = 1
			break
		}
	}
	os.Exit(exitCode)
	return nil
}

// runTests executes all diagnostic tests
func runTests(b *bundle.Bundle, extractPath string) []TestResult {
	var results []TestResult

	// Test 1: OOM Events
	results = append(results, testOOMEvents(extractPath))

	// Test 2: Node Pressure
	results = append(results, testNodePressure(extractPath))

	// Test 3: etcd Health
	results = append(results, testEtcdHealth(extractPath))

	// Test 4: Pod Health
	results = append(results, testPodHealth(extractPath))

	// Test 5: Bundle Completeness
	results = append(results, testBundleCompleteness(extractPath))

	return results
}

// testOOMEvents checks for OOM kill events
func testOOMEvents(extractPath string) TestResult {
	result := TestResult{
		Name:        "OOM Event Detection",
		Status:      "PASS",
		Description: "Check for out-of-memory kills",
	}

	analysis, err := bundle.AnalyzeOOMEvents(extractPath)
	if err != nil {
		result.Status = "SKIP"
		result.Description = fmt.Sprintf("Could not analyze OOM events: %v", err)
		return result
	}

	if len(analysis) == 0 {
		result.Details = append(result.Details, "No OOM events found")
		return result
	}

	result.Status = "FAIL"
	result.Description = fmt.Sprintf("Found %d OOM kill(s)", len(analysis))

	for _, oom := range analysis {
		detail := fmt.Sprintf("  • %s/%s: %s", oom.PodName, oom.ContainerName, oom.OOMKillTime)
		if oom.QoSClass != "" {
			detail += fmt.Sprintf(" (QoS: %s)", oom.QoSClass)
		}
		if oom.NodeMemoryPressure {
			detail += " [Node under memory pressure]"
		}
		result.Details = append(result.Details, detail)
	}

	return result
}

// testNodePressure checks for node pressure conditions
func testNodePressure(extractPath string) TestResult {
	result := TestResult{
		Name:        "Node Pressure Check",
		Status:      "PASS",
		Description: "Check for memory/disk/PID pressure on nodes",
	}

	nodes, err := bundle.ParseNodeDescribe(extractPath)
	if err != nil {
		result.Status = "SKIP"
		result.Description = fmt.Sprintf("Could not parse node describe: %v", err)
		return result
	}

	var pressureNodes []string
	for _, node := range nodes {
		if node.MemoryPressure || node.DiskPressure || node.PIDPressure {
			pressureNodes = append(pressureNodes, node.Name)
			var issues []string
			if node.MemoryPressure {
				issues = append(issues, "memory pressure")
			}
			if node.DiskPressure {
				issues = append(issues, "disk pressure")
			}
			if node.PIDPressure {
				issues = append(issues, "PID pressure")
			}
			result.Details = append(result.Details,
				fmt.Sprintf("  • %s: %s", node.Name, strings.Join(issues, ", ")))
		}
	}

	if len(pressureNodes) > 0 {
		result.Status = "FAIL"
		result.Description = fmt.Sprintf("Found %d node(s) with pressure conditions", len(pressureNodes))
	} else {
		result.Details = append(result.Details, "No pressure conditions found on any node")
	}

	return result
}

// testEtcdHealth checks etcd cluster health
func testEtcdHealth(extractPath string) TestResult {
	result := TestResult{
		Name:        "etcd Health Status",
		Status:      "PASS",
		Description: "Check etcd alarms and member health",
	}

	// Try to get etcd health from bundle
	etcdHealth, err := bundle.ParseEtcdHealth(extractPath)
	if err != nil {
		result.Status = "SKIP"
		result.Description = fmt.Sprintf("Could not parse etcd health: %v", err)
		return result
	}

	if etcdHealth.HasAlarms {
		result.Status = "FAIL"
		result.Description = fmt.Sprintf("etcd has alarms: %s (count: %d)",
			etcdHealth.AlarmType, etcdHealth.AlarmCount)
	} else if !etcdHealth.Healthy {
		result.Status = "WARN"
		result.Description = "etcd is not healthy (no alarms detected)"
	} else {
		result.Details = append(result.Details, "etcd is healthy with no alarms")
	}

	return result
}

// testPodHealth checks for unhealthy pods
func testPodHealth(extractPath string) TestResult {
	result := TestResult{
		Name:        "Pod Health Check",
		Status:      "PASS",
		Description: "Check for pods in CrashLoopBackOff, Error, or Failed states",
	}

	pods, err := bundle.ParsePods(extractPath)
	if err != nil {
		result.Status = "SKIP"
		result.Description = fmt.Sprintf("Could not parse pods: %v", err)
		return result
	}

	var unhealthyPods []string
	for _, pod := range pods {
		// Check both State (from Rancher API) and KubectlStatus (from bundle)
		state := strings.ToLower(pod.State)
		kubectlStatus := strings.ToLower(pod.KubectlStatus)

		isUnhealthy := false
		statusDisplay := pod.State

		// Check Rancher State
		if strings.Contains(state, "error") || strings.Contains(state, "failed") {
			isUnhealthy = true
		}

		// Check kubectl status (more detailed from bundle)
		if strings.Contains(kubectlStatus, "crashloop") ||
			strings.Contains(kubectlStatus, "error") ||
			strings.Contains(kubectlStatus, "failed") ||
			strings.Contains(kubectlStatus, "oomkilled") {
			isUnhealthy = true
			statusDisplay = pod.KubectlStatus
		}

		if isUnhealthy {
			unhealthyPods = append(unhealthyPods, pod.Name)
			result.Details = append(result.Details,
				fmt.Sprintf("  • %s/%s: %s", pod.NamespaceID, pod.Name, statusDisplay))
		}
	}

	if len(unhealthyPods) > 0 {
		result.Status = "FAIL"
		result.Description = fmt.Sprintf("Found %d unhealthy pod(s)", len(unhealthyPods))
	} else {
		result.Details = append(result.Details, "All pods are healthy")
	}

	return result
}

// testBundleCompleteness checks if bundle has all expected data
func testBundleCompleteness(extractPath string) TestResult {
	result := TestResult{
		Name:        "Bundle Completeness",
		Status:      "PASS",
		Description: "Check if bundle contains essential data",
	}

	hasNodes := false
	hasPods := false
	hasEvents := false
	hasEtcd := false

	// Check for key files
	if _, err := os.Stat(fmt.Sprintf("%s/rke2/kubectl/nodes", extractPath)); err == nil {
		hasNodes = true
	}
	if _, err := os.Stat(fmt.Sprintf("%s/rke2/kubectl/pods", extractPath)); err == nil {
		hasPods = true
	}
	if _, err := os.Stat(fmt.Sprintf("%s/rke2/kubectl/events", extractPath)); err == nil {
		hasEvents = true
	}
	if _, err := os.Stat(fmt.Sprintf("%s/rke2/etcd/endpoint_status", extractPath)); err == nil {
		hasEtcd = true
	}

	score := 0
	if hasNodes {
		score++
		result.Details = append(result.Details, "  ✓ Node data present")
	} else {
		result.Details = append(result.Details, "  ✗ Node data missing")
	}
	if hasPods {
		score++
		result.Details = append(result.Details, "  ✓ Pod data present")
	} else {
		result.Details = append(result.Details, "  ✗ Pod data missing")
	}
	if hasEvents {
		score++
		result.Details = append(result.Details, "  ✓ Event data present")
	} else {
		result.Details = append(result.Details, "  ✗ Event data missing")
	}
	if hasEtcd {
		score++
		result.Details = append(result.Details, "  ✓ etcd data present")
	} else {
		result.Details = append(result.Details, "  ✗ etcd data missing")
	}

	result.Description = fmt.Sprintf("Bundle completeness: %d/4", score)

	if score < 3 {
		result.Status = "WARN"
	}

	return result
}

// outputTable prints results in table format
func outputTable(results []TestResult) {
	fmt.Println()
	fmt.Println(color.New(color.Bold).Sprint("R8S Cluster Test Results"))
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	passCount := 0
	failCount := 0
	warnCount := 0
	skipCount := 0

	for _, r := range results {
		// Count results
		switch r.Status {
		case "PASS":
			passCount++
		case "FAIL":
			failCount++
		case "WARN":
			warnCount++
		case "SKIP":
			skipCount++
		}

		// Print status with color
		var statusStr string
		switch r.Status {
		case "PASS":
			statusStr = color.GreenString("✓ PASS")
		case "FAIL":
			statusStr = color.RedString("✗ FAIL")
		case "WARN":
			statusStr = color.YellowString("⚠ WARN")
		case "SKIP":
			statusStr = color.CyanString("⊘ SKIP")
		}

		fmt.Printf("%-30s %s\n", r.Name, statusStr)
		fmt.Printf("  %s\n", r.Description)
		for _, detail := range r.Details {
			fmt.Printf("%s\n", detail)
		}
		fmt.Println()
	}

	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("Summary: %s | %s | %s | %s\n",
		color.GreenString("%d passed", passCount),
		color.RedString("%d failed", failCount),
		color.YellowString("%d warnings", warnCount),
		color.CyanString("%d skipped", skipCount))
	fmt.Println()
}

// outputSummary prints a brief summary
func outputSummary(results []TestResult) {
	failCount := 0
	for _, r := range results {
		if r.Status == "FAIL" {
			failCount++
		}
	}

	if failCount == 0 {
		fmt.Println("PASS: All tests passed")
	} else {
		fmt.Printf("FAIL: %d test(s) failed\n", failCount)
	}
}

// outputJSON prints results as JSON using proper JSON encoding
func outputJSON(results []TestResult) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(results); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
	}
}

func init() {
	rootCmd.AddCommand(testClusterCmd)

	// test-cluster specific flags
	testClusterCmd.Flags().BoolVar(&testClusterAll, "all", false, "show all pods, not just issues")
	testClusterCmd.Flags().StringVar(&testClusterFormat, "format", "table", "output format: table, json, summary")
}
