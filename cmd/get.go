// Package cmd implements the CLI commands for r8s.
// v0.8.0: r8s get - kubectl-compatible resource listing
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/Rancheroo/r8s/internal/bundle"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get [resource] [bundle-path]",
	Short: "Get resources from bundle (kubectl-style)",
	Long: `List resources from a Rancher support bundle like kubectl.

SUPPORTED RESOURCES:
  pods, pod, po          - List pods
  nodes, node, no        - List nodes (from bundle metadata)
  namespaces, ns         - List namespaces
  deployments, deploy    - List deployments
  services, svc          - List services
  events, ev             - List cluster events

EXAMPLES:
  # List all pods
  r8s get pods ./bundle/

  # List pods in specific namespace
  r8s get pods ./bundle/ -n kube-system

  # List all namespaces
  r8s get ns ./bundle/

  # List deployments with JSON output
  r8s get deploy ./bundle/ -o json

OUTPUT FORMATS:
  table  - Human-readable columns (default)
  json   - JSON for piping to jq
  yaml   - YAML output
  wide   - Extra columns like kubectl wide`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runGet,
}

var (
	getOutput       string // Output format: table, json, yaml, wide
	getNamespace    string // Filter by namespace (-n)
	getAllNamespaces bool  // Show all namespaces (-A)
	getSelector     string // Label selector (-l)
)

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.Flags().StringVarP(&getOutput, "output", "o", "table", "Output format: table, json, yaml, wide, name")
	getCmd.Flags().StringVarP(&getNamespace, "namespace", "n", "", "Filter by namespace")
	getCmd.Flags().BoolVarP(&getAllNamespaces, "all-namespaces", "A", false, "Show resources in all namespaces")
	getCmd.Flags().StringVarP(&getSelector, "selector", "l", "", "Label selector (not yet implemented)")
}

// runGet executes the get command
func runGet(cmd *cobra.Command, args []string) error {
	resource := strings.ToLower(args[0])
	
	// Determine bundle path
	var bundlePath string
	if len(args) > 1 {
		bundlePath = args[1]
	} else {
		// Check if bundle path is in tuiBundlePath from root
		bundlePath = tuiBundlePath
		if bundlePath == "" {
			return fmt.Errorf("bundle path required: r8s get %s [bundle-path]", resource)
		}
	}

	// Load bundle
	b, err := loadBundle(bundlePath)
	if err != nil {
		return fmt.Errorf("failed to load bundle: %w", err)
	}
	defer b.Close()

	// Route to appropriate handler
	switch resource {
	case "pods", "pod", "po":
		return getPods(b, getNamespace, getAllNamespaces)
	case "nodes", "node", "no":
		return getNodes(b)
	case "namespaces", "ns":
		return getNamespaces(b)
	case "deployments", "deployment", "deploy":
		return getDeployments(b, getNamespace, getAllNamespaces)
	case "services", "service", "svc":
		return getServices(b, getNamespace, getAllNamespaces)
	case "events", "event", "ev":
		return getEvents(b, getNamespace, getAllNamespaces)
	default:
		return fmt.Errorf("unknown resource type: %s (supported: pods, nodes, ns, deploy, svc, events)", resource)
	}
}

// loadBundle loads a bundle from path
func loadBundle(path string) (*bundle.Bundle, error) {
	// Use existing load logic
	importOpts := bundle.ImportOptions{
		Path:    path,
		Verbose: verbose,
	}
	
	b, err := bundle.Load(importOpts)
	if err != nil {
		return nil, err
	}
	
	if !b.Loaded {
		return nil, fmt.Errorf("bundle failed to load")
	}
	
	return b, nil
}

// ============================================
// GET PODS
// ============================================

type PodRow struct {
	Namespace   string `json:"namespace"`
	Name        string `json:"name"`
	Ready       string `json:"ready"`
	Status      string `json:"status"`
	Restarts    int    `json:"restarts"`
	Age         string `json:"age"`
	Node        string `json:"node,omitempty"`
}

func getPods(b *bundle.Bundle, namespace string, allNamespaces bool) error {
	// Collect pods
	pods := []PodRow{}
	
	for _, pod := range b.Pods {
		// Filter by namespace
		if namespace != "" && pod.Namespace != namespace {
			continue
		}
		if !allNamespaces && namespace == "" && pod.Namespace != "default" && pod.Namespace != "cattle-system" {
			// If no namespace specified, show interesting namespaces
			// Actually, show all for now
		}
		
		// Determine status (simplified - would need yaml parsing for real status)
		status := "Unknown"
		if pod.HasCurrentLogs {
			status = "Running"
		}
		
		pods = append(pods, PodRow{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			Ready:     fmt.Sprintf("%d/%d", len(pod.Containers), len(pod.Containers)),
			Status:    status,
			Restarts:  0, // Would need to parse from yaml
			Age:       "-", // Would need timestamp from yaml
			Node:      b.Manifest.NodeName,
		})
	}
	
	// Sort by namespace, then name
	sort.Slice(pods, func(i, j int) bool {
		if pods[i].Namespace != pods[j].Namespace {
			return pods[i].Namespace < pods[j].Namespace
		}
		return pods[i].Name < pods[j].Name
	})
	
	// Output
	switch getOutput {
	case "json":
		return outputGetJSON(pods)
	case "yaml":
		return outputGetYAML(pods)
	case "wide":
		return outputPodsWide(pods)
	case "name":
		return outputNames(pods)
	default:
		return outputPodsTable(pods, allNamespaces || namespace == "")
	}
}

func outputPodsTable(pods []PodRow, showNamespace bool) error {
	const padding = 4
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', 0)
	
	if showNamespace {
		fmt.Fprintln(w, "NAMESPACE\tNAME\tREADY\tSTATUS\tRESTARTS\tAGE")
	} else {
		fmt.Fprintln(w, "NAME\tREADY\tSTATUS\tRESTARTS\tAGE")
	}
	
	for _, pod := range pods {
		if showNamespace {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\n",
				pod.Namespace, pod.Name, pod.Ready, pod.Status, pod.Restarts, pod.Age)
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\n",
				pod.Name, pod.Ready, pod.Status, pod.Restarts, pod.Age)
		}
	}
	
	w.Flush()
	fmt.Printf("\n%d pods found\n", len(pods))
	return nil
}

func outputPodsWide(pods []PodRow) error {
	const padding = 4
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', 0)
	
	fmt.Fprintln(w, "NAMESPACE\tNAME\tREADY\tSTATUS\tRESTARTS\tAGE\tNODE")
	
	for _, pod := range pods {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\t%s\n",
			pod.Namespace, pod.Name, pod.Ready, pod.Status, pod.Restarts, pod.Age, pod.Node)
	}
	
	w.Flush()
	return nil
}

// ============================================
// GET NODES
// ============================================

type NodeRow struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Version string `json:"version"`
	OS      string `json:"os,omitempty"`
}

func getNodes(b *bundle.Bundle) error {
	// Bundle only has one node's data (the node it was collected from)
	// For future: support multi-node bundles
	
	nodes := []NodeRow{
		{
			Name:    b.Manifest.NodeName,
			Status:  "Ready",
			Version: b.Manifest.K8sVersion,
			OS:      "-", // Could parse from node yaml
		},
	}
	
	switch getOutput {
	case "json":
		return outputGetJSON(nodes)
	case "yaml":
		return outputGetYAML(nodes)
	default:
		return outputNodesTable(nodes)
	}
}

func outputNodesTable(nodes []NodeRow) error {
	const padding = 4
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', 0)
	
	fmt.Fprintln(w, "NAME\tSTATUS\tVERSION\tOS-IMAGE")
	
	for _, node := range nodes {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			node.Name, node.Status, node.Version, node.OS)
	}
	
	w.Flush()
	return nil
}

// ============================================
// GET NAMESPACES
// ============================================

type NamespaceRow struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Age    string `json:"age"`
}

func getNamespaces(b *bundle.Bundle) error {
	// Collect unique namespaces from pods
	namespaceMap := make(map[string]bool)
	for _, pod := range b.Pods {
		namespaceMap[pod.Namespace] = true
	}
	
	namespaces := []NamespaceRow{}
	for ns := range namespaceMap {
		namespaces = append(namespaces, NamespaceRow{
			Name:   ns,
			Status: "Active",
			Age:    "-",
		})
	}
	
	sort.Slice(namespaces, func(i, j int) bool {
		return namespaces[i].Name < namespaces[j].Name
	})
	
	switch getOutput {
	case "json":
		return outputGetJSON(namespaces)
	case "yaml":
		return outputGetYAML(namespaces)
	default:
		return outputNamespacesTable(namespaces)
	}
}

func outputNamespacesTable(namespaces []NamespaceRow) error {
	const padding = 4
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', 0)
	
	fmt.Fprintln(w, "NAME\tSTATUS\tAGE")
	
	for _, ns := range namespaces {
		fmt.Fprintf(w, "%s\t%s\t%s\n", ns.Name, ns.Status, ns.Age)
	}
	
	w.Flush()
	return nil
}

// ============================================
// GET DEPLOYMENTS
// ============================================

func getDeployments(b *bundle.Bundle, namespace string, allNamespaces bool) error {
	// Stub - would need to parse deployment yaml from bundle
	fmt.Println("deployments: not yet implemented (stub)")
	return nil
}

// ============================================
// GET SERVICES
// ============================================

func getServices(b *bundle.Bundle, namespace string, allNamespaces bool) error {
	// Stub - would need to parse service yaml from bundle
	fmt.Println("services: not yet implemented (stub)")
	return nil
}

// ============================================
// GET EVENTS
// ============================================

type EventRow struct {
	Namespace   string `json:"namespace"`
	Type        string `json:"type"`
	Reason      string `json:"reason"`
	Message     string `json:"message"`
	Source      string `json:"source"`
	Count       int    `json:"count"`
	Age         string `json:"age"`
}

func getEvents(b *bundle.Bundle, namespace string, allNamespaces bool) error {
	// Stub - would need to parse events from bundle
	fmt.Println("events: not yet implemented (stub)")
	fmt.Println("Run 'r8s analyze' for issue detection instead.")
	return nil
}

// ============================================
// OUTPUT HELPERS
// ============================================

func outputGetJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func outputGetYAML(data interface{}) error {
	// Simple YAML output using json intermediate
	fmt.Println("# YAML output - using JSON for now")
	return outputGetJSON(data)
}

func outputNames(data interface{}) error {
	// Output just names, one per line
	// Type assertion needed
	switch v := data.(type) {
	case []PodRow:
		for _, item := range v {
			fmt.Println(item.Name)
		}
	case []NamespaceRow:
		for _, item := range v {
			fmt.Println(item.Name)
		}
	case []NodeRow:
		for _, item := range v {
			fmt.Println(item.Name)
		}
	}
	return nil
}

func outputTableHeader(title string) {
	header := color.New(color.Bold, color.FgCyan)
	header.Printf("\n%s\n", title)
	header.Println(strings.Repeat("═", 60))
}
