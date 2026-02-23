// Package output provides standardized output formatting for r8s CLI commands.
// Supports JSON, YAML, table, and wide formats with kubectl-compatible styling.
package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"gopkg.in/yaml.v3"
)

// Format represents the output format type
type Format string

const (
	// FormatTable is the default human-readable table format
	FormatTable Format = "table"
	// FormatJSON outputs as JSON
	FormatJSON Format = "json"
	// FormatYAML outputs as YAML
	FormatYAML Format = "yaml"
	// FormatWide outputs extra columns like kubectl wide
	FormatWide Format = "wide"
	// FormatName outputs just resource names (one per line)
	FormatName Format = "name"
)

// ValidFormats returns all valid format strings
func ValidFormats() []string {
	return []string{"table", "json", "yaml", "wide", "name"}
}

// IsValid checks if a format string is valid
func IsValid(format string) bool {
	switch Format(format) {
	case FormatTable, FormatJSON, FormatYAML, FormatWide, FormatName:
		return true
	default:
		return false
	}
}

// Options contains output formatting options
type Options struct {
	Format        Format
	ShowNamespace bool
	AllNamespaces bool
	NoHeaders     bool
}

// PodRow represents a pod in kubectl get pods output
type PodRow struct {
	Namespace string `json:"namespace" yaml:"namespace"`
	Name      string `json:"name" yaml:"name"`
	Ready     string `json:"ready" yaml:"ready"`
	Status    string `json:"status" yaml:"status"`
	Restarts  int    `json:"restarts" yaml:"restarts"`
	Age       string `json:"age" yaml:"age"`
	Node      string `json:"node,omitempty" yaml:"node,omitempty"`
	IP        string `json:"ip,omitempty" yaml:"ip,omitempty"`
}

// NodeRow represents a node in kubectl get nodes output
type NodeRow struct {
	Name             string `json:"name" yaml:"name"`
	Status           string `json:"status" yaml:"status"`
	Roles            string `json:"roles" yaml:"roles"`
	Age              string `json:"age" yaml:"age"`
	Version          string `json:"version" yaml:"version"`
	InternalIP       string `json:"internalIP,omitempty" yaml:"internalIP,omitempty"`
	ExternalIP       string `json:"externalIP,omitempty" yaml:"externalIP,omitempty"`
	OSImage          string `json:"osImage,omitempty" yaml:"osImage,omitempty"`
	KernelVersion    string `json:"kernelVersion,omitempty" yaml:"kernelVersion,omitempty"`
	ContainerRuntime string `json:"containerRuntime,omitempty" yaml:"containerRuntime,omitempty"`
}

// NamespaceRow represents a namespace in kubectl get ns output
type NamespaceRow struct {
	Name   string `json:"name" yaml:"name"`
	Status string `json:"status" yaml:"status"`
	Age    string `json:"age" yaml:"age"`
}

// EventRow represents an event in kubectl get events output
type EventRow struct {
	Namespace     string `json:"namespace" yaml:"namespace"`
	LastSeen      string `json:"lastSeen" yaml:"lastSeen"`
	Type          string `json:"type" yaml:"type"`
	Reason        string `json:"reason" yaml:"reason"`
	Object        string `json:"object" yaml:"object"`
	Message       string `json:"message" yaml:"message"`
}

// DeploymentRow represents a deployment in kubectl get deploy output
type DeploymentRow struct {
	Namespace string `json:"namespace" yaml:"namespace"`
	Name      string `json:"name" yaml:"name"`
	Ready     string `json:"ready" yaml:"ready"`
	UpToDate  int    `json:"upToDate" yaml:"upToDate"`
	Available int    `json:"available" yaml:"available"`
	Age       string `json:"age" yaml:"age"`
}

// ServiceRow represents a service in kubectl get svc output
type ServiceRow struct {
	Namespace string `json:"namespace" yaml:"namespace"`
	Name      string `json:"name" yaml:"name"`
	Type      string `json:"type" yaml:"type"`
	ClusterIP string `json:"clusterIP" yaml:"clusterIP"`
	ExternalIP string `json:"externalIP" yaml:"externalIP"`
	Ports     string `json:"ports" yaml:"ports"`
	Age       string `json:"age" yaml:"age"`
}

// OutputPods outputs pod data in the specified format
func OutputPods(pods []PodRow, opts Options) error {
	switch opts.Format {
	case FormatJSON:
		return outputJSON(pods)
	case FormatYAML:
		return outputYAML(pods)
	case FormatWide:
		return outputPodsWide(pods, opts)
	case FormatName:
		return outputNames(pods)
	default:
		return outputPodsTable(pods, opts)
	}
}

// OutputNodes outputs node data in the specified format
func OutputNodes(nodes []NodeRow, opts Options) error {
	switch opts.Format {
	case FormatJSON:
		return outputJSON(nodes)
	case FormatYAML:
		return outputYAML(nodes)
	case FormatWide:
		return outputNodesWide(nodes, opts)
	case FormatName:
		return outputNames(nodes)
	default:
		return outputNodesTable(nodes, opts)
	}
}

// OutputNamespaces outputs namespace data in the specified format
func OutputNamespaces(namespaces []NamespaceRow, opts Options) error {
	switch opts.Format {
	case FormatJSON:
		return outputJSON(namespaces)
	case FormatYAML:
		return outputYAML(namespaces)
	case FormatName:
		return outputNames(namespaces)
	default:
		return outputNamespacesTable(namespaces, opts)
	}
}

// OutputEvents outputs event data in the specified format
func OutputEvents(events []EventRow, opts Options) error {
	switch opts.Format {
	case FormatJSON:
		return outputJSON(events)
	case FormatYAML:
		return outputYAML(events)
	default:
		return outputEventsTable(events, opts)
	}
}

// OutputDeployments outputs deployment data in the specified format
func OutputDeployments(deployments []DeploymentRow, opts Options) error {
	switch opts.Format {
	case FormatJSON:
		return outputJSON(deployments)
	case FormatYAML:
		return outputYAML(deployments)
	case FormatName:
		return outputNames(deployments)
	default:
		return outputDeploymentsTable(deployments, opts)
	}
}

// OutputServices outputs service data in the specified format
func OutputServices(services []ServiceRow, opts Options) error {
	switch opts.Format {
	case FormatJSON:
		return outputJSON(services)
	case FormatYAML:
		return outputYAML(services)
	case FormatWide:
		return outputServicesWide(services, opts)
	case FormatName:
		return outputNames(services)
	default:
		return outputServicesTable(services, opts)
	}
}

// outputPodsTable outputs pods in table format
func outputPodsTable(pods []PodRow, opts Options) error {
	const padding = 4
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', 0)

	showNamespace := opts.ShowNamespace || opts.AllNamespaces

	if !opts.NoHeaders {
		if showNamespace {
			fmt.Fprintln(w, "NAMESPACE\tNAME\tREADY\tSTATUS\tRESTARTS\tAGE")
		} else {
			fmt.Fprintln(w, "NAME\tREADY\tSTATUS\tRESTARTS\tAGE")
		}
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

// outputPodsWide outputs pods in wide format (extra columns)
func outputPodsWide(pods []PodRow, opts Options) error {
	const padding = 4
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', 0)

	if !opts.NoHeaders {
		fmt.Fprintln(w, "NAMESPACE\tNAME\tREADY\tSTATUS\tRESTARTS\tAGE\tIP\tNODE")
	}

	for _, pod := range pods {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\t%s\t%s\n",
			pod.Namespace, pod.Name, pod.Ready, pod.Status, pod.Restarts, pod.Age, pod.IP, pod.Node)
	}

	w.Flush()
	return nil
}

// outputNodesTable outputs nodes in table format
func outputNodesTable(nodes []NodeRow, opts Options) error {
	const padding = 4
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', 0)

	if !opts.NoHeaders {
		fmt.Fprintln(w, "NAME\tSTATUS\tROLES\tAGE\tVERSION")
	}

	for _, node := range nodes {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			node.Name, node.Status, node.Roles, node.Age, node.Version)
	}

	w.Flush()
	fmt.Printf("\n%d nodes found\n", len(nodes))
	return nil
}

// outputNodesWide outputs nodes in wide format
func outputNodesWide(nodes []NodeRow, opts Options) error {
	const padding = 4
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', 0)

	if !opts.NoHeaders {
		fmt.Fprintln(w, "NAME\tSTATUS\tROLES\tAGE\tVERSION\tINTERNAL-IP\tEXTERNAL-IP\tOS-IMAGE\tKERNEL-VERSION\tCONTAINER-RUNTIME")
	}

	for _, node := range nodes {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			node.Name, node.Status, node.Roles, node.Age, node.Version,
			node.InternalIP, node.ExternalIP, node.OSImage, node.KernelVersion, node.ContainerRuntime)
	}

	w.Flush()
	return nil
}

// outputNamespacesTable outputs namespaces in table format
func outputNamespacesTable(namespaces []NamespaceRow, opts Options) error {
	const padding = 4
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', 0)

	if !opts.NoHeaders {
		fmt.Fprintln(w, "NAME\tSTATUS\tAGE")
	}

	for _, ns := range namespaces {
		fmt.Fprintf(w, "%s\t%s\t%s\n", ns.Name, ns.Status, ns.Age)
	}

	w.Flush()
	fmt.Printf("\n%d namespaces found\n", len(namespaces))
	return nil
}

// outputEventsTable outputs events in table format
func outputEventsTable(events []EventRow, opts Options) error {
	const padding = 4
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', 0)

	if !opts.NoHeaders {
		fmt.Fprintln(w, "NAMESPACE\tLAST SEEN\tTYPE\tREASON\tOBJECT\tMESSAGE")
	}

	for _, event := range events {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			event.Namespace, event.LastSeen, event.Type, event.Reason, event.Object, event.Message)
	}

	w.Flush()
	fmt.Printf("\n%d events found\n", len(events))
	return nil
}

// outputDeploymentsTable outputs deployments in table format
func outputDeploymentsTable(deployments []DeploymentRow, opts Options) error {
	const padding = 4
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', 0)

	showNamespace := opts.ShowNamespace || opts.AllNamespaces

	if !opts.NoHeaders {
		if showNamespace {
			fmt.Fprintln(w, "NAMESPACE\tNAME\tREADY\tUP-TO-DATE\tAVAILABLE\tAGE")
		} else {
			fmt.Fprintln(w, "NAME\tREADY\tUP-TO-DATE\tAVAILABLE\tAGE")
		}
	}

	for _, deploy := range deployments {
		if showNamespace {
			fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\t%s\n",
				deploy.Namespace, deploy.Name, deploy.Ready, deploy.UpToDate, deploy.Available, deploy.Age)
		} else {
			fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%s\n",
				deploy.Name, deploy.Ready, deploy.UpToDate, deploy.Available, deploy.Age)
		}
	}

	w.Flush()
	fmt.Printf("\n%d deployments found\n", len(deployments))
	return nil
}

// outputServicesTable outputs services in table format
func outputServicesTable(services []ServiceRow, opts Options) error {
	const padding = 4
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', 0)

	showNamespace := opts.ShowNamespace || opts.AllNamespaces

	if !opts.NoHeaders {
		if showNamespace {
			fmt.Fprintln(w, "NAMESPACE\tNAME\tTYPE\tCLUSTER-IP\tEXTERNAL-IP\tPORT(S)\tAGE")
		} else {
			fmt.Fprintln(w, "NAME\tTYPE\tCLUSTER-IP\tEXTERNAL-IP\tPORT(S)\tAGE")
		}
	}

	for _, svc := range services {
		if showNamespace {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				svc.Namespace, svc.Name, svc.Type, svc.ClusterIP, svc.ExternalIP, svc.Ports, svc.Age)
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				svc.Name, svc.Type, svc.ClusterIP, svc.ExternalIP, svc.Ports, svc.Age)
		}
	}

	w.Flush()
	fmt.Printf("\n%d services found\n", len(services))
	return nil
}

// outputServicesWide outputs services in wide format (includes selector)
func outputServicesWide(services []ServiceRow, opts Options) error {
	// For services, wide format just shows SELECTOR column
	// For now, delegate to table format (selector not parsed from bundles)
	return outputServicesTable(services, opts)
}

// outputJSON outputs data as indented JSON
func outputJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// outputYAML outputs data as YAML
func outputYAML(data interface{}) error {
	encoder := yaml.NewEncoder(os.Stdout)
	defer encoder.Close()
	return encoder.Encode(data)
}

// outputNames outputs just resource names
func outputNames(data interface{}) error {
	switch v := data.(type) {
	case []PodRow:
		for _, item := range v {
			fmt.Println(item.Name)
		}
	case []NodeRow:
		for _, item := range v {
			fmt.Println(item.Name)
		}
	case []NamespaceRow:
		for _, item := range v {
			fmt.Println(item.Name)
		}
	case []DeploymentRow:
		for _, item := range v {
			fmt.Println(item.Name)
		}
	case []ServiceRow:
		for _, item := range v {
			fmt.Println(item.Name)
		}
	default:
		return fmt.Errorf("unsupported type for name output")
	}
	return nil
}

// FormatAge formats a duration as kubectl-style age string
func FormatAge(created time.Time) string {
	if created.IsZero() {
		return "<unknown>"
	}

	duration := time.Since(created)

	switch {
	case duration < time.Minute:
		return fmt.Sprintf("%ds", int(duration.Seconds()))
	case duration < time.Hour:
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	case duration < 24*time.Hour:
		return fmt.Sprintf("%dh", int(duration.Hours()))
	case duration < 365*24*time.Hour:
		return fmt.Sprintf("%dd", int(duration.Hours()/24))
	default:
		return fmt.Sprintf("%dy", int(duration.Hours()/24/365))
	}
}

// TruncateMessage truncates a message to max length with ellipsis
func TruncateMessage(msg string, maxLen int) string {
	if len(msg) <= maxLen {
		return msg
	}
	return msg[:maxLen-3] + "..."
}

// FormatError formats an error message for CLI output
func FormatError(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	
	// Provide actionable error messages
	switch {
	case strings.Contains(msg, "bundle not loaded") || strings.Contains(msg, "no such file"):
		return "Bundle not loaded. Run: r8s get pods <bundle-path>"
	case strings.Contains(msg, "no pods found"):
		return "No pods found in bundle. Check the bundle path and namespace filter."
	case strings.Contains(msg, "namespace") && strings.Contains(msg, "not found"):
		return fmt.Sprintf("Namespace not found. Use -A to see all namespaces or check namespace name.")
	default:
		return fmt.Sprintf("Error: %s", msg)
	}
}

// Spinner represents a progress indicator
type Spinner struct {
	frames []string
	index  int
	active bool
}

// NewSpinner creates a new spinner
func NewSpinner() *Spinner {
	return &Spinner{
		frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		index:  0,
	}
}

// Start starts the spinner animation
func (s *Spinner) Start(message string) {
	s.active = true
	fmt.Printf("%s %s", s.frames[0], message)
}

// Stop stops the spinner
func (s *Spinner) Stop() {
	s.active = false
	fmt.Print("\r\033[K") // Clear line
}

// Next advances to the next frame
func (s *Spinner) Next() {
	if !s.active {
		return
	}
	s.index = (s.index + 1) % len(s.frames)
}
