// Package cmd implements the CLI commands for r8s.
// Sprint 9: r8s get pods - kubectl-compatible pod listing
package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/Rancheroo/r8s/internal/bundle"
	"github.com/Rancheroo/r8s/output"
	"github.com/spf13/cobra"
)

// getPodsCmd represents the get pods command
var getPodsCmd = &cobra.Command{
	Use:     "pods [bundle-path]",
	Aliases: []string{"pod", "po"},
	Short:   "List pods from bundle",
	Long: `List pods from a Rancher support bundle.

Similar to 'kubectl get pods', but works offline with bundle data.

EXAMPLES:
  # List all pods
  r8s get pods ./bundle/

  # List pods in specific namespace
  r8s get pods ./bundle/ -n cattle-system

  # List pods in all namespaces
  r8s get pods ./bundle/ -A

  # List pods with JSON output
  r8s get pods ./bundle/ -o json

  # List pods with wide output (includes node)
  r8s get pods ./bundle/ -o wide

OUTPUT COLUMNS:
  NAME       - Pod name
  READY      - Ready containers / Total containers
  STATUS     - Pod status (Running, Pending, Failed, etc.)
  RESTARTS   - Number of container restarts
  AGE        - Time since pod creation
  NODE       - Node name (in wide mode)`,
	Args: cobra.RangeArgs(0, 1),
	RunE: runGetPods,
}

var (
	podsNamespace    string
	podsOutput       string
	podsAllNamespaces bool
	podsSelector     string
	podsShowLabels   bool
	podsNoHeaders    bool
	podsFieldSelector string
)

func init() {
	getCmd.AddCommand(getPodsCmd)

	getPodsCmd.Flags().StringVarP(&podsNamespace, "namespace", "n", "", "Filter by namespace")
	getPodsCmd.Flags().StringVarP(&podsOutput, "output", "o", "table", "Output format: table, json, yaml, wide, name")
	getPodsCmd.Flags().BoolVarP(&podsAllNamespaces, "all-namespaces", "A", false, "List pods in all namespaces")
	getPodsCmd.Flags().StringVarP(&podsSelector, "selector", "l", "", "Label selector (not yet implemented)")
	getPodsCmd.Flags().BoolVar(&podsShowLabels, "show-labels", false, "Show all labels (not yet implemented)")
	getPodsCmd.Flags().BoolVar(&podsNoHeaders, "no-headers", false, "Hide column headers")
	getPodsCmd.Flags().StringVar(&podsFieldSelector, "field-selector", "", "Field selector (not yet implemented)")
}

// runGetPods executes the get pods command
func runGetPods(cmd *cobra.Command, args []string) error {
	// Validate output format
	if !output.IsValid(podsOutput) {
		return fmt.Errorf("invalid output format: %q (supported: %v)", podsOutput, output.ValidFormats())
	}

	// Determine bundle path
	bundlePath := ""
	if len(args) > 0 {
		bundlePath = args[0]
	} else {
		// Check if bundle path is in tuiBundlePath from root
		bundlePath = tuiBundlePath
	}

	if bundlePath == "" {
		return fmt.Errorf("bundle path required: r8s get pods [bundle-path]\n\nExamples:\n  r8s get pods ./bundle/\n  r8s get pods ./bundle/ -n kube-system")
	}

	// Validate bundle exists
	if _, err := os.Stat(bundlePath); err != nil {
		return fmt.Errorf("bundle not found: %w\n\nEnsure the bundle path is correct and the bundle has been extracted.", err)
	}

	// Load bundle
	b, err := loadBundleForPods(bundlePath)
	if err != nil {
		return fmt.Errorf("failed to load bundle: %w\n\nRun 'r8s validate %s' to check bundle integrity.", err, bundlePath)
	}
	defer b.Close()

	// Get pods from bundle
	pods, err := getPodsFromBundle(b, podsNamespace, podsAllNamespaces)
	if err != nil {
		return fmt.Errorf("failed to get pods: %w", err)
	}

	if len(pods) == 0 {
		nsMsg := ""
		if podsNamespace != "" {
			nsMsg = fmt.Sprintf(" in namespace '%s'", podsNamespace)
		}
		return fmt.Errorf("no pods found%s\n\nTry:\n  - Using -A to see all namespaces\n  - Checking namespace name with 'r8s get namespaces %s'", nsMsg, bundlePath)
	}

	// Output pods
	opts := output.Options{
		Format:        output.Format(podsOutput),
		ShowNamespace: podsNamespace != "",
		AllNamespaces: podsAllNamespaces,
		NoHeaders:     podsNoHeaders,
	}

	return output.OutputPods(pods, opts)
}

// loadBundleForPods loads a bundle from path with specific options for pods
func loadBundleForPods(path string) (*bundle.Bundle, error) {
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

// getPodsFromBundle retrieves pods from the bundle, filtering as needed
func getPodsFromBundle(b *bundle.Bundle, namespace string, allNamespaces bool) ([]output.PodRow, error) {
	// Use the kubectl parser for enriched pod data
	rancherPods, err := bundle.ParsePods(b.ExtractPath)
	if err != nil {
		// Fallback to basic PodInfo
		return getPodsFromPodInfo(b, namespace, allNamespaces)
	}

	var pods []output.PodRow

	for _, pod := range rancherPods {
		// Filter by namespace if specified and not showing all
		if !allNamespaces && namespace != "" && pod.NamespaceID != namespace {
			continue
		}

		// Determine ready status
		ready := pod.KubectlReady
		if ready == "" {
			// Fallback: assume 1/1
			ready = "1/1"
		}

		// Determine node
		node := pod.NodeName
		if node == "" {
			node = pod.NodeID
			if node == "" {
				node = pod.Node
				if node == "" {
					node = b.Manifest.NodeName
				}
			}
		}

		pods = append(pods, output.PodRow{
			Namespace: pod.NamespaceID,
			Name:      pod.Name,
			Ready:     ready,
			Status:    pod.KubectlStatus,
			Restarts:  pod.KubectlRestarts,
			Age:       pod.KubectlAge,
			Node:      node,
			IP:        pod.KubectlIP,
		})
	}

	// Sort by namespace, then name
	sort.Slice(pods, func(i, j int) bool {
		if pods[i].Namespace != pods[j].Namespace {
			return pods[i].Namespace < pods[j].Namespace
		}
		return pods[i].Name < pods[j].Name
	})

	return pods, nil
}

// getPodsFromPodInfo extracts pods from basic PodInfo (fallback)
func getPodsFromPodInfo(b *bundle.Bundle, namespace string, allNamespaces bool) ([]output.PodRow, error) {
	var pods []output.PodRow

	for _, podInfo := range b.Pods {
		// Filter by namespace if specified and not showing all
		if !allNamespaces && namespace != "" && podInfo.Namespace != namespace {
			continue
		}

		// Determine status based on available data
		status := "Unknown"
		if podInfo.HasCurrentLogs {
			status = "Running"
		}
		if podInfo.HasPreviousLogs && !podInfo.HasCurrentLogs {
			status = "CrashLoopBackOff"
		}

		pods = append(pods, output.PodRow{
			Namespace: podInfo.Namespace,
			Name:      podInfo.Name,
			Ready:     fmt.Sprintf("%d/%d", len(podInfo.Containers), len(podInfo.Containers)),
			Status:    status,
			Restarts:  0,
			Age:       "-",
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

	return pods, nil
}
