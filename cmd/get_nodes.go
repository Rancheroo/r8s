// Package cmd implements the CLI commands for r8s.
// Sprint 9: r8s get nodes - kubectl-compatible node listing
package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/Rancheroo/r8s/internal/bundle"
	"github.com/Rancheroo/r8s/output"
	"github.com/spf13/cobra"
)

// getNodesCmd represents the get nodes command
var getNodesCmd = &cobra.Command{
	Use:     "nodes [bundle-path]",
	Aliases: []string{"node", "no"},
	Short:   "List nodes from bundle",
	Long: `List nodes from a Rancher support bundle.

Similar to 'kubectl get nodes', but works offline with bundle data.
Note: Bundles typically contain data from a single node, so this will
show the node the bundle was collected from.

EXAMPLES:
  # List all nodes
  r8s get nodes ./bundle/

  # List nodes with wide output (includes IPs)
  r8s get nodes ./bundle/ -o wide

  # List nodes with JSON output
  r8s get nodes ./bundle/ -o json

OUTPUT COLUMNS:
  NAME       - Node name
  STATUS     - Node status (Ready, NotReady, etc.)
  ROLES      - Node roles (control-plane, worker, etcd)
  AGE        - Time since node joined
  VERSION    - Kubernetes version`,
	Args: cobra.RangeArgs(0, 1),
	RunE: runGetNodes,
}

var (
	nodesOutput    string
	nodesNoHeaders bool
	nodesSelector  string
)

func init() {
	getCmd.AddCommand(getNodesCmd)

	getNodesCmd.Flags().StringVarP(&nodesOutput, "output", "o", "table", "Output format: table, json, yaml, wide, name")
	getNodesCmd.Flags().BoolVar(&nodesNoHeaders, "no-headers", false, "Hide column headers")
	getNodesCmd.Flags().StringVarP(&nodesSelector, "selector", "l", "", "Label selector (not yet implemented)")
}

// runGetNodes executes the get nodes command
func runGetNodes(cmd *cobra.Command, args []string) error {
	// Validate output format
	if !output.IsValid(nodesOutput) {
		return fmt.Errorf("invalid output format: %q (supported: %v)", nodesOutput, output.ValidFormats())
	}

	// Determine bundle path
	bundlePath := ""
	if len(args) > 0 {
		bundlePath = args[0]
	} else {
		bundlePath = tuiBundlePath
	}

	if bundlePath == "" {
		return fmt.Errorf("bundle path required: r8s get nodes [bundle-path]\n\nExamples:\n  r8s get nodes ./bundle/")
	}

	// Validate bundle exists
	if _, err := os.Stat(bundlePath); err != nil {
		return fmt.Errorf("bundle not found: %w\n\nEnsure the bundle path is correct and the bundle has been extracted.", err)
	}

	// Load bundle
	b, err := loadBundleForNodes(bundlePath)
	if err != nil {
		return fmt.Errorf("failed to load bundle: %w\n\nRun 'r8s validate %s' to check bundle integrity.", err, bundlePath)
	}
	defer b.Close()

	// Get nodes from bundle
	nodes, err := getNodesFromBundle(b)
	if err != nil {
		return fmt.Errorf("failed to get nodes: %w", err)
	}

	if len(nodes) == 0 {
		return fmt.Errorf("no nodes found in bundle\n\nThe bundle may not contain node information. Try 'r8s analyze %s' for full bundle inspection.", bundlePath)
	}

	// Output nodes
	opts := output.Options{
		Format:    output.Format(nodesOutput),
		NoHeaders: nodesNoHeaders,
	}

	return output.OutputNodes(nodes, opts)
}

// loadBundleForNodes loads a bundle from path with specific options for nodes
func loadBundleForNodes(path string) (*bundle.Bundle, error) {
	importOpts := bundle.ImportOptions{
		Path:    path,
		Verbose: verbose,
	}

	b, err := bundle.Load(importOpts)
	if err != nil {
		return nil, err
	}

	if !b.Loaded {
		return nil, fmt.Errorf("bundle failed to load properly")
	}

	return b, nil
}

// getNodesFromBundle retrieves nodes from the bundle
func getNodesFromBundle(b *bundle.Bundle) ([]output.NodeRow, error) {
	// Try to parse from kubectl nodes output first
	nodeInfos, err := bundle.ParseNodes(b.ExtractPath)
	if err == nil && len(nodeInfos) > 0 {
		return convertNodeInfos(nodeInfos, b)
	}

	// Fallback to nodes from node describe
	nodeDescribes, err := bundle.ParseNodeDescribe(b.ExtractPath)
	if err == nil && len(nodeDescribes) > 0 {
		return convertNodeDescribes(nodeDescribes, b)
	}

	// Final fallback: single node from bundle manifest
	return getNodeFromManifest(b), nil
}

// convertNodeInfos converts bundle NodeInfo to output NodeRow
func convertNodeInfos(nodeInfos []bundle.NodeInfo, b *bundle.Bundle) ([]output.NodeRow, error) {
	var nodes []output.NodeRow

	for _, ni := range nodeInfos {
		// Determine roles
		roles := "<none>"
		if b.Manifest != nil {
			roles = determineRoles(ni.Name, b.Manifest.NodeName)
		}

		// Determine age - we don't have this from basic node info
		age := "-"

		// Get version from manifest
		version := "-"
		if b.Manifest != nil {
			version = b.Manifest.K8sVersion
		}

		nodes = append(nodes, output.NodeRow{
			Name:    ni.Name,
			Status:  ni.Status,
			Roles:   roles,
			Age:     age,
			Version: version,
		})
	}

	// Sort by name
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Name < nodes[j].Name
	})

	return nodes, nil
}

// convertNodeDescribes converts bundle NodeConditions to output NodeRow
func convertNodeDescribes(nodeConditions []bundle.NodeConditions, b *bundle.Bundle) ([]output.NodeRow, error) {
	var nodes []output.NodeRow

	for _, nc := range nodeConditions {
		// Determine status
		status := "NotReady"
		if nc.Ready {
			status = "Ready"
		}

		// Format roles
		roles := formatRoles(nc)

		nodes = append(nodes, output.NodeRow{
			Name:             nc.Name,
			Status:           status,
			Roles:            roles,
			Age:              "-", // Not available from describe output
			Version:          nc.KubeletVersion,
			InternalIP:       nc.InternalIP,
			OSImage:          nc.OSImage,
			KernelVersion:    nc.KernelVersion,
			ContainerRuntime: nc.ContainerRuntime,
		})
	}

	// Sort by name
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Name < nodes[j].Name
	})

	return nodes, nil
}

// getNodeFromManifest creates a single node row from bundle manifest
func getNodeFromManifest(b *bundle.Bundle) []output.NodeRow {
	if b.Manifest == nil {
		return []output.NodeRow{}
	}

	// Determine roles based on typical RKE2 naming
	roles := determineRoles(b.Manifest.NodeName, b.Manifest.NodeName)

	return []output.NodeRow{
		{
			Name:    b.Manifest.NodeName,
			Status:  "Ready",
			Roles:   roles,
			Age:     "-",
			Version: b.Manifest.K8sVersion,
		},
	}
}

// determineRoles determines node roles from name and context
func determineRoles(nodeName, manifestNodeName string) string {
	var roles []string

	// Check name patterns
	nameLower := strings.ToLower(nodeName)
	
	// Control plane indicators
	if strings.Contains(nameLower, "server") || 
	   strings.Contains(nameLower, "master") ||
	   strings.Contains(nameLower, "control") {
		roles = append(roles, "control-plane")
	}

	// etcd indicators
	if strings.Contains(nameLower, "etcd") {
		roles = append(roles, "etcd")
	}

	// Worker indicators
	if strings.Contains(nameLower, "worker") || 
	   strings.Contains(nameLower, "agent") {
		roles = append(roles, "worker")
	}

	// If no roles detected, assume worker for most nodes
	if len(roles) == 0 {
		// Check if this is the bundle collection node - likely control-plane
		if nodeName == manifestNodeName {
			roles = []string{"control-plane", "etcd"}
		} else {
			roles = []string{"worker"}
		}
	}

	return strings.Join(roles, ",")
}

// formatRoles formats roles from NodeConditions
func formatRoles(nc bundle.NodeConditions) string {
	var roles []string

	if nc.IsControlPlane {
		roles = append(roles, "control-plane")
	}
	if nc.IsEtcd {
		roles = append(roles, "etcd")
	}
	if nc.IsWorker {
		roles = append(roles, "worker")
	}

	// Fallback to parsing from Roles field
	if len(roles) == 0 && nc.Roles != "" {
		return nc.Roles
	}

	if len(roles) == 0 {
		return "<none>"
	}

	return strings.Join(roles, ",")
}
