// Package cmd implements the CLI commands for r8s.
// v0.8.0: r8s describe - Show resource details (kubectl-style)
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/Rancheroo/r8s/internal/bundle"
)

// describeCmd represents the describe command
var describeCmd = &cobra.Command{
	Use:   "describe [resource] [bundle-path] [name]",
	Short: "Show details of a resource (kubectl-style)",
	Long: `Show detailed information about a resource from a bundle.

Similar to 'kubectl describe', but works offline with bundle data.

SUPPORTED RESOURCES:
  pod, pods, po          - Describe a pod
  node, nodes, no        - Describe a node
  deployment, deploy     - Describe a deployment
  service, svc           - Describe a service
  namespace, ns          - Describe a namespace
  event, events, ev      - Describe events

EXAMPLES:
  # Describe a pod
  r8s describe pod ./bundle/ nginx-pod

  # Describe pod (shorthand)
  r8s describe po ./bundle/ nginx-pod

  # Describe a node
  r8s describe node ./bundle/ my-node

  # Describe with JSON output
  r8s describe pod ./bundle/ nginx-pod -o json

  # Show events only
  r8s describe pod ./bundle/ nginx-pod --events

POD NAME MATCHING:
  Pod names support partial matching like 'r8s logs'.
  If multiple pods match, you'll be prompted to specify.

OUTPUT FORMATS:
  text   - Human-readable description (default)
  json   - Full JSON resource definition
  yaml   - Full YAML resource definition`,
	Args: cobra.RangeArgs(2, 3),
	RunE: runDescribe,
}

var (
	describeOutput string // Output format: text, json, yaml
	describeEvents bool   // Show only events
	describeWide   bool   // Show wide output
)

func init() {
	rootCmd.AddCommand(describeCmd)

	describeCmd.Flags().StringVarP(&describeOutput, "output", "o", "text", "Output format: text, json, yaml")
	describeCmd.Flags().BoolVar(&describeEvents, "events", false, "Show only events")
	describeCmd.Flags().BoolVar(&describeWide, "wide", false, "Show wide output (more details)")
}

// runDescribe executes the describe command
func runDescribe(cmd *cobra.Command, args []string) error {
	// Parse arguments: describe [resource] [bundle] [name]
	// OR: describe [resource] [name] (if bundle path from root)
	resource := strings.ToLower(args[0])
	
	var bundlePath, resourceName string
	
	if len(args) == 3 {
		// Full form: describe resource bundle name
		bundlePath = args[1]
		resourceName = args[2]
	} else {
		// Short form: describe resource name (use default bundle)
		bundlePath = tuiBundlePath
		resourceName = args[1]
		if bundlePath == "" {
			return fmt.Errorf("bundle path required: r8s describe %s [bundle-path] [name]", resource)
		}
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

	// Route to appropriate handler
	switch resource {
	case "pod", "pods", "po":
		return describePod(b, resourceName)
	case "node", "nodes", "no":
		return describeNode(b)
	case "deployment", "deploy":
		return describeDeployment(b, resourceName)
	case "service", "svc":
		return describeService(b, resourceName)
	case "namespace", "ns":
		return describeNamespace(b, resourceName)
	case "event", "events", "ev":
		return describeEventsResource(b)
	default:
		return fmt.Errorf("unknown resource type: %s (supported: pod, node, deploy, svc, ns, events)", resource)
	}
}

// ============================================
// DESCRIBE POD
// ============================================

func describePod(b *bundle.Bundle, name string) error {
	// Find matching pod
	matchedPod, err := findPodForDescribe(b, name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
		return err
	}

	// Output based on format
	switch describeOutput {
	case "json":
		return outputDescribeJSON(matchedPod)
	case "yaml":
		return outputDescribeYAML(matchedPod)
	default:
		return outputPodDescribe(matchedPod, b)
	}
}

func findPodForDescribe(b *bundle.Bundle, name string) (*bundle.PodInfo, error) {
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

func outputPodDescribe(pod *bundle.PodInfo, b *bundle.Bundle) error {
	fmt.Printf("Name:         %s\n", pod.Name)
	fmt.Printf("Namespace:    %s\n", pod.Namespace)
	fmt.Println()

	// Pod info
	fmt.Println("Status:")
	status := "Unknown"
	if pod.HasCurrentLogs {
		status = "Running"
	}
	fmt.Printf("  Phase:      %s\n", status)
	fmt.Printf("  Ready:      %t\n", pod.HasCurrentLogs)
	fmt.Println()

	// Containers
	fmt.Println("Containers:")
	for _, container := range pod.Containers {
		fmt.Printf("  %s:\n", container)
		fmt.Printf("    Current Logs:  %t\n", pod.HasCurrentLogs)
		fmt.Printf("    Previous Logs: %t\n", pod.HasPreviousLogs)
	}
	fmt.Println()

	// Log availability
	fmt.Println("Log Files:")
	foundLogs := false
	for _, logFile := range b.LogFiles {
		if logFile.Type == bundle.LogTypePod &&
			logFile.Namespace == pod.Namespace &&
			logFile.PodName == pod.Name {
			fmt.Printf("  %s\n", logFile.Path)
			foundLogs = true
		}
	}
	if !foundLogs {
		fmt.Println("  (no log files found)")
	}
	fmt.Println()

	// Commands
	fmt.Println("Commands:")
	fmt.Printf("  View logs:  r8s logs %s %s\n", b.Path, pod.Name)
	if pod.HasPreviousLogs {
		fmt.Printf("  Prev logs:  r8s logs %s %s -p\n", b.Path, pod.Name)
	}
	fmt.Println()

	return nil
}

// ============================================
// DESCRIBE NODE
// ============================================

func describeNode(b *bundle.Bundle) error {
	// Bundle only has one node (the one collected from)
	fmt.Println("Name:         ", b.Manifest.NodeName)
	fmt.Println()

	fmt.Println("System Info:")
	fmt.Printf("  RKE2 Version:  %s\n", b.Manifest.RKE2Version)
	fmt.Printf("  K8s Version:   %s\n", b.Manifest.K8sVersion)
	fmt.Println()

	fmt.Println("Bundle Info:")
	fmt.Printf("  Collected At:  %s\n", b.Manifest.CollectedAt)
	fmt.Printf("  Bundle Type:   %s\n", b.Manifest.BundleType)
	fmt.Printf("  File Count:    %d\n", b.Manifest.FileCount)
	fmt.Println()

	fmt.Println("Available Resources:")
	fmt.Printf("  Pods:          %d\n", len(b.Pods))
	fmt.Printf("  Log Files:     %d\n", len(b.LogFiles))
	fmt.Printf("  Namespaces:    %d\n", len(b.Namespaces))
	fmt.Printf("  Deployments:   %d\n", len(b.Deployments))
	fmt.Printf("  Services:      %d\n", len(b.Services))
	fmt.Println()

	return nil
}

// ============================================
// DESCRIBE DEPLOYMENT
// ============================================

func describeDeployment(b *bundle.Bundle, name string) error {
	// Stub - would need to search through b.Deployments
	if len(b.Deployments) == 0 {
		fmt.Println("No deployments found in bundle")
		return nil
	}

	fmt.Printf("Deployment:    %s\n", name)
	fmt.Println()
	fmt.Printf("(Found %d deployments in bundle - detailed view not yet implemented)\n", len(b.Deployments))
	fmt.Println("Use 'r8s get deploy ./bundle/' to list all deployments")
	return nil
}

// ============================================
// DESCRIBE SERVICE
// ============================================

func describeService(b *bundle.Bundle, name string) error {
	// Stub - would need to search through b.Services
	if len(b.Services) == 0 {
		fmt.Println("No services found in bundle")
		return nil
	}

	fmt.Printf("Service:       %s\n", name)
	fmt.Println()
	fmt.Printf("(Found %d services in bundle - detailed view not yet implemented)\n", len(b.Services))
	fmt.Println("Use 'r8s get svc ./bundle/' to list all services")
	return nil
}

// ============================================
// DESCRIBE NAMESPACE
// ============================================

func describeNamespace(b *bundle.Bundle, name string) error {
	// Collect pods in this namespace
	podCount := 0
	for _, pod := range b.Pods {
		if pod.Namespace == name {
			podCount++
		}
	}

	fmt.Printf("Name:          %s\n", name)
	fmt.Println()

	fmt.Println("Status:")
	fmt.Printf("  Phase:       Active\n")
	fmt.Println()

	fmt.Println("Resources:")
	fmt.Printf("  Pods:        %d\n", podCount)
	fmt.Println()

	return nil
}

// ============================================
// DESCRIBE EVENTS
// ============================================

func describeEventsResource(b *bundle.Bundle) error {
	if len(b.Events) == 0 {
		fmt.Println("No events found in bundle")
		return nil
	}

	fmt.Printf("Events (%d total):\n", len(b.Events))
	fmt.Println()
	fmt.Println("(Event details not yet implemented - showing count only)")
	return nil
}

// ============================================
// OUTPUT HELPERS
// ============================================

func outputDescribeJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func outputDescribeYAML(data interface{}) error {
	encoder := yaml.NewEncoder(os.Stdout)
	defer encoder.Close()
	return encoder.Encode(data)
}
