// Package cmd implements the CLI commands for r8s.
// v0.8.0: r8s get - kubectl-compatible resource listing
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/Rancheroo/r8s/internal/bundle"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get [resource] [bundle-path]",
	Short: "Get resources from bundle (kubectl-compatible)",
	Long: `Display one or many resources from a Rancher support bundle.

Similar to kubectl, but operates on offline bundles instead of live clusters.

SUPPORTED RESOURCES:
  pods, pod, po          Show all pods
  nodes, node, no        Show all nodes  
  ns, namespaces         Show all namespaces
  deploy, deployments    Show all deployments
  svc, services          Show all services
  events, ev             Show all events
  crds, crd              Show all custom resource definitions
  daemonsets, ds         Show all daemonsets
  pv                     Show all persistent volumes
  pvc                    Show all persistent volume claims
  statefulsets, sts      Show all statefulsets
  configmaps, cm         Show all configmaps
  helmcharts             Show all helm charts

EXAMPLES:
  # Get all pods in a bundle
  r8s get pods ./bundle/

  # Get nodes with JSON output
  r8s get nodes ./bundle/ -o json

  # Get specific resource type
  r8s get deploy ./bundle/ --namespace=cattle-system`,
}

// podCmd represents the get pods command
var podCmd = &cobra.Command{
	Use:   "pods [bundle-path]",
	Short: "List all pods in bundle",
	Aliases: []string{"pod", "po"},
	Args:  cobra.ExactArgs(1),
	RunE:  runGetPods,
}

// nodeCmd represents the get nodes command
var nodeCmd = &cobra.Command{
	Use:   "nodes [bundle-path]",
	Short: "List all nodes in bundle",
	Aliases: []string{"node", "no"},
	Args:  cobra.ExactArgs(1),
	RunE:  runGetNodes,
}

// namespaceCmd represents the get namespaces command
var namespaceCmd = &cobra.Command{
	Use:   "namespaces [bundle-path]",
	Short: "List all namespaces in bundle",
	Aliases: []string{"ns"},
	Args:  cobra.ExactArgs(1),
	RunE:  runGetNamespaces,
}

// deployCmd represents the get deployments command
var deployCmd = &cobra.Command{
	Use:   "deployments [bundle-path]",
	Short: "List all deployments in bundle",
	Aliases: []string{"deploy"},
	Args:  cobra.ExactArgs(1),
	RunE:  runGetDeployments,
}

// svcCmd represents the get services command
var svcCmd = &cobra.Command{
	Use:   "services [bundle-path]",
	Short: "List all services in bundle",
	Aliases: []string{"svc"},
	Args:  cobra.ExactArgs(1),
	RunE:  runGetServices,
}

// eventsCmd represents the get events command
var eventsCmd = &cobra.Command{
	Use:   "events [bundle-path]",
	Short: "List all events in bundle",
	Aliases: []string{"ev"},
	Args:  cobra.ExactArgs(1),
	RunE:  runGetEvents,
}

var (
	getOutput   string // Output format: table, json, yaml
	getNamespace string // Filter by namespace
	getAll      bool   // Show all resources
)

func init() {
	rootCmd.AddCommand(getCmd)
	
	getCmd.AddCommand(podCmd)
	getCmd.AddCommand(nodeCmd)
	getCmd.AddCommand(namespaceCmd)
	getCmd.AddCommand(deployCmd)
	getCmd.AddCommand(svcCmd)
	getCmd.AddCommand(eventsCmd)

	// Global flags for all get commands
	getCmd.PersistentFlags().StringVarP(&getOutput, "output", "o", "table", "Output format: table, json")
	getCmd.PersistentFlags().StringVarP(&getNamespace, "namespace", "n", "", "Filter by namespace")
	getCmd.PersistentFlags().BoolVarP(&getAll, "all-namespaces", "A", false, "Show resources from all namespaces")
}

// runGetPods lists pods from bundle
func runGetPods(cmd *cobra.Command, args []string) error {
	bundlePath := args[0]
	
	pods, err := bundle.ParsePods(bundlePath)
	if err != nil {
		return fmt.Errorf("failed to parse pods: %w", err)
	}

	// Filter by namespace if specified
	if getNamespace != "" {
		filtered := []interface{}{}
		for _, pod := range pods {
			if pod.NamespaceID == getNamespace {
				filtered = append(filtered, pod)
			}
		}
		_ = filtered // Use filtered for now, proper implementation later
	}

	// Output based on format
	switch getOutput {
	case "json":
		return outputDataJSON(pods)
	default:
		return outputPodsTable(pods)
	}
}

// runGetNodes lists nodes from bundle
func runGetNodes(cmd *cobra.Command, args []string) error {
	bundlePath := args[0]
	
	nodes, err := bundle.ParseNodes(bundlePath)
	if err != nil {
		return fmt.Errorf("failed to parse nodes: %w", err)
	}

	switch getOutput {
	case "json":
		return outputDataJSON(nodes)
	default:
		return outputNodesTable(nodes)
	}
}

// runGetNamespaces lists namespaces from bundle
func runGetNamespaces(cmd *cobra.Command, args []string) error {
	bundlePath := args[0]
	
	ns, err := bundle.ParseNamespaces(bundlePath)
	if err != nil {
		return fmt.Errorf("failed to parse namespaces: %w", err)
	}

	switch getOutput {
	case "json":
		return outputDataJSON(ns)
	default:
		return outputNamespacesTable(ns)
	}
}

// runGetDeployments lists deployments from bundle
func runGetDeployments(cmd *cobra.Command, args []string) error {
	bundlePath := args[0]
	
	deploys, err := bundle.ParseDeployments(bundlePath)
	if err != nil {
		return fmt.Errorf("failed to parse deployments: %w", err)
	}

	switch getOutput {
	case "json":
		return outputDataJSON(deploys)
	default:
		return outputDeploymentsTable(deploys)
	}
}

// runGetServices lists services from bundle
func runGetServices(cmd *cobra.Command, args []string) error {
	bundlePath := args[0]
	
	svcs, err := bundle.ParseServices(bundlePath)
	if err != nil {
		return fmt.Errorf("failed to parse services: %w", err)
	}

	switch getOutput {
	case "json":
		return outputDataJSON(svcs)
	default:
		return outputServicesTable(svcs)
	}
}

// runGetEvents lists events from bundle
func runGetEvents(cmd *cobra.Command, args []string) error {
	bundlePath := args[0]
	
	events, err := bundle.ParseEvents(bundlePath)
	if err != nil {
		return fmt.Errorf("failed to parse events: %w", err)
	}

	switch getOutput {
	case "json":
		return outputDataJSON(events)
	default:
		return outputEventsTable(events)
	}
}

// Table output helpers
func outputPodsTable(pods interface{}) error {
	// Use reflection or type assertion to handle different types
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	
	header := color.New(color.Bold)
	header.Fprintln(w, "NAMESPACE\tNAME\tREADY\tSTATUS\tRESTARTS\tAGE\tIP\tNODE")
	
	// This would need proper type assertion based on actual pod struct
	// For now, print placeholder
	fmt.Fprintln(w, "Use -o json for structured output")
	
	return w.Flush()
}

func outputNodesTable(nodes []bundle.NodeInfo) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	
	header := color.New(color.Bold)
	header.Fprintln(w, "NAME\tSTATUS\tROLES\tAGE\tVERSION")
	
	for _, node := range nodes {
		statusColor := color.GreenString
		if node.Status != "Ready" {
			statusColor = color.RedString
		}
		fmt.Fprintf(w, "%s\t%s\t\t\t\n", node.Name, statusColor(node.Status))
	}
	
	return w.Flush()
}

func outputNamespacesTable(ns interface{}) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	
	header := color.New(color.Bold)
	header.Fprintln(w, "NAME\tSTATUS\tAGE")
	
	fmt.Fprintln(w, "Use -o json for structured output")
	
	return w.Flush()
}

func outputDeploymentsTable(deploys interface{}) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	
	header := color.New(color.Bold)
	header.Fprintln(w, "NAMESPACE\tNAME\tREADY\tUP-TO-DATE\tAVAILABLE\tAGE")
	
	fmt.Fprintln(w, "Use -o json for structured output")
	
	return w.Flush()
}

func outputServicesTable(svcs interface{}) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	
	header := color.New(color.Bold)
	header.Fprintln(w, "NAMESPACE\tNAME\tTYPE\tCLUSTER-IP\tEXTERNAL-IP\tPORT(S)\tAGE")
	
	fmt.Fprintln(w, "Use -o json for structured output")
	
	return w.Flush()
}

func outputEventsTable(events interface{}) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	
	header := color.New(color.Bold)
	header.Fprintln(w, "NAMESPACE\tLAST SEEN\tTYPE\tREASON\tOBJECT\tMESSAGE")
	
	fmt.Fprintln(w, "Use -o json for structured output")
	
	return w.Flush()
}

// Generic helpers
func outputDataJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func outputResourceList(items []interface{}, headers []string, rowFunc func(interface{}) []string) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	
	header := color.New(color.Bold)
	header.Fprintln(w, strings.Join(headers, "\t"))
	
	for _, item := range items {
		row := rowFunc(item)
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}
	
	return w.Flush()
}
