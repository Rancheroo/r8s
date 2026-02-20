// Package cmd implements the CLI commands for r8s.
// Sprint 9 Day 3: r8s describe - kubectl-style resource description
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// describeCmd represents the describe command
var describeCmd = &cobra.Command{
	Use:   "describe [kind] [bundle-path] [name]",
	Short: "Show detailed resource information from bundle",
	Long: `Show detailed information about Kubernetes resources from a bundle.

Similar to 'kubectl describe', but works offline with bundle data.

EXAMPLES:
  # Describe a specific pod
  r8s describe pod ./bundle/ rancher-7c4c7b8f5-x2v9p

  # Describe a node
  r8s describe node ./bundle/ node-1

  # Describe with YAML output
  r8s describe pod ./bundle/ rancher-xyz -o yaml

  # Describe all pods in a namespace
  r8s describe pods ./bundle/ -n cattle-system

  # Auto-detect resource type from name
  r8s describe ./bundle/ rancher-xyz

Supported kinds: pod, pods, node, nodes, deployment, deployments, 
service, services, configmap, configmaps, event, events`,
	Args: cobra.RangeArgs(2, 3),
	RunE: runDescribe,
}

var (
	describeNamespace string
	describeOutput    string
	describeSelector  string
)

func init() {
	rootCmd.AddCommand(describeCmd)

	describeCmd.Flags().StringVarP(&describeNamespace, "namespace", "n", "", "Filter by namespace")
	describeCmd.Flags().StringVarP(&describeOutput, "output", "o", "human", "Output format: human, json, yaml, wide")
	describeCmd.Flags().StringVarP(&describeSelector, "selector", "l", "", "Label selector (e.g. app=rancher)")
}

func runDescribe(cmd *cobra.Command, args []string) error {
	// Parse arguments
	kind, bundlePath, name := parseDescribeArgs(args)

	// Validate bundle
	if _, err := os.Stat(bundlePath); err != nil {
		return fmt.Errorf("bundle path not found: %w", err)
	}

	// Find and describe resources
	resources, err := findResources(bundlePath, kind, name, describeNamespace, describeSelector)
	if err != nil {
		return fmt.Errorf("failed to find resources: %w", err)
	}

	if len(resources) == 0 {
		fmt.Fprintf(os.Stderr, "No resources found")
		if kind != "" {
			fmt.Fprintf(os.Stderr, " of kind '%s'", kind)
		}
		if name != "" {
			fmt.Fprintf(os.Stderr, " with name '%s'", name)
		}
		if describeNamespace != "" {
			fmt.Fprintf(os.Stderr, " in namespace '%s'", describeNamespace)
		}
		fmt.Fprintln(os.Stderr)
		return nil
	}

	// Output based on format
	switch describeOutput {
	case "json":
		return outputDescribeJSON(resources)
	case "yaml":
		return outputDescribeYAML(resources)
	case "wide":
		return outputDescribeWide(resources)
	default:
		return outputDescribeHuman(resources)
	}
}

// parseDescribeArgs handles flexible argument order
func parseDescribeArgs(args []string) (kind, bundlePath, name string) {
	if len(args) == 2 {
		// Format: describe ./bundle/ name (auto-detect kind)
		bundlePath = args[0]
		name = args[1]
	} else {
		// Format: describe kind ./bundle/ name
		kind = strings.ToLower(args[0])
		bundlePath = args[1]
		name = args[2]
	}

	// Normalize kind aliases
	switch kind {
	case "pods":
		kind = "pod"
	case "nodes":
		kind = "node"
	case "deployments":
		kind = "deployment"
	case "services":
		kind = "service"
	case "configmaps":
		kind = "configmap"
	case "events":
		kind = "event"
	}

	return
}

// ResourceInfo holds parsed resource data
type ResourceInfo struct {
	Kind       string                 `json:"kind" yaml:"kind"`
	Name       string                 `json:"name" yaml:"name"`
	Namespace  string                 `json:"namespace" yaml:"namespace"`
	Labels     map[string]string      `json:"labels" yaml:"labels"`
	Status     string                 `json:"status" yaml:"status"`
	Details    map[string]interface{} `json:"details" yaml:"details"`
	SourceFile string                 `json:"-" yaml:"-"`
}

// findResources discovers resources in bundle
func findResources(bundlePath, kind, name, namespace, selector string) ([]ResourceInfo, error) {
	var resources []ResourceInfo

	// Map of kind to file patterns
	kindPatterns := map[string][]string{
		"pod":        {"rke2/kubectl/pods", "kubectl/pods"},
		"node":       {"rke2/kubectl/nodes", "kubectl/nodes"},
		"deployment": {"rke2/kubectl/deployments", "kubectl/deployments"},
		"service":    {"rke2/kubectl/services", "kubectl/services"},
		"configmap":  {"rke2/kubectl/configmaps", "kubectl/configmaps"},
		"event":      {"rke2/kubectl/events", "kubectl/events"},
	}

	// If specific kind requested, search only that
	if kind != "" {
		patterns := kindPatterns[kind]
		for _, pattern := range patterns {
			file := filepath.Join(bundlePath, pattern)
			if _, err := os.Stat(file); err == nil {
				rs, err := parseResourceFile(file, kind, name, namespace, selector)
				if err == nil {
					resources = append(resources, rs...)
				}
			}
		}
	} else {
		// Auto-detect: search all kinds
		for k := range kindPatterns {
			rs, _ := findResources(bundlePath, k, name, namespace, selector)
			resources = append(resources, rs...)
		}
	}

	return resources, nil
}

// parseResourceFile extracts resources from kubectl output files
func parseResourceFile(path, kind, nameFilter, namespaceFilter, selector string) ([]ResourceInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	content := string(data)
	var resources []ResourceInfo

	// Simple parsing: split by resource separators
	// kubectl describe output uses "Name:" as separator
	sections := strings.Split(content, "Name:\s+")

	for _, section := range sections {
		if strings.TrimSpace(section) == "" {
			continue
		}

		res := ResourceInfo{
			Kind:       kind,
			SourceFile: path,
			Labels:     make(map[string]string),
			Details:    make(map[string]interface{}),
		}

		// Parse name (first line)
		lines := strings.Split(section, "\n")
		if len(lines) > 0 {
			res.Name = strings.TrimSpace(lines[0])
		}

		// Parse namespace
		for _, line := range lines {
			if strings.HasPrefix(line, "Namespace:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					res.Namespace = strings.TrimSpace(parts[1])
				}
			}
			if strings.HasPrefix(line, "Status:") || strings.HasPrefix(line, "Phase:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					res.Status = strings.TrimSpace(parts[1])
				}
			}
		}

		// Apply filters
		if nameFilter != "" && !strings.Contains(res.Name, nameFilter) {
			continue
		}
		if namespaceFilter != "" && res.Namespace != namespaceFilter {
			continue
		}

		resources = append(resources, res)
	}

	return resources, nil
}

// outputDescribeHuman outputs human-readable format
func outputDescribeHuman(resources []ResourceInfo) error {
	for i, res := range resources {
		if i > 0 {
			fmt.Println()
			fmt.Println(strings.Repeat("-", 60))
			fmt.Println()
		}

		fmt.Printf("%s %s/%s\n",
			color.CyanString("==>"),
			color.YellowString(strings.Title(res.Kind)),
			color.GreenString(res.Name),
		)

		if res.Namespace != "" {
			fmt.Printf("Namespace:  %s\n", res.Namespace)
		}
		if res.Status != "" {
			statusColor := color.GreenString
			if res.Status == "Error" || res.Status == "Failed" {
				statusColor = color.RedString
			} else if res.Status == "Pending" || res.Status == "Unknown" {
				statusColor = color.YellowString
			}
			fmt.Printf("Status:     %s\n", statusColor(res.Status))
		}
		if len(res.Labels) > 0 {
			fmt.Println("Labels:")
			for k, v := range res.Labels {
				fmt.Printf("  %s=%s\n", k, v)
			}
		}
		fmt.Printf("Source:     %s\n", res.SourceFile)
	}

	return nil
}

// outputDescribeJSON outputs JSON format
func outputDescribeJSON(resources []ResourceInfo) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(resources)
}

// outputDescribeYAML outputs YAML format
func outputDescribeYAML(resources []ResourceInfo) error {
	encoder := yaml.NewEncoder(os.Stdout)
	defer encoder.Close()
	return encoder.Encode(resources)
}

// outputDescribeWide outputs wide table format
func outputDescribeWide(resources []ResourceInfo) error {
	// Header
	fmt.Printf("%-15s %-30s %-20s %-15s\n", "KIND", "NAME", "NAMESPACE", "STATUS")
	fmt.Println(strings.Repeat("-", 80))

	for _, res := range resources {
		fmt.Printf("%-15s %-30s %-20s %-15s\n",
			res.Kind,
			res.Name,
			res.Namespace,
			res.Status,
		)
	}

	return nil
}
