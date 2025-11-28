// Package bundle provides support for importing and working with Rancher/RKE2 support bundles.
// It enables offline analysis of cluster diagnostics without requiring a live API connection.
package bundle

import (
	"time"
)

// Bundle represents a loaded support bundle with all its metadata and contents.
type Bundle struct {
	// Path is the original path to the tar.gz bundle file
	Path string

	// ExtractPath is the temporary directory where bundle contents are extracted
	ExtractPath string

	// Manifest contains parsed metadata about the bundle
	Manifest *BundleManifest

	// Pods contains inventory of all pods found in the bundle
	Pods []PodInfo

	// LogFiles contains inventory of all log files in the bundle
	LogFiles []LogFileInfo

	// kubectl resources parsed from bundle
	CRDs        []interface{} // Will be []rancher.CRD when imported
	Deployments []interface{} // Will be []rancher.Deployment
	Services    []interface{} // Will be []rancher.Service
	Namespaces  []interface{} // Will be []rancher.Namespace

	// Loaded indicates whether the bundle has been successfully loaded
	Loaded bool

	// Size is the total size of the bundle in bytes
	Size int64
}

// BundleManifest contains metadata extracted from a support bundle.
type BundleManifest struct {
	// NodeName is the name of the node this bundle was collected from
	NodeName string

	// CollectedAt is the timestamp when the bundle was collected
	CollectedAt time.Time

	// RKE2Version is the version of RKE2 running on the node
	RKE2Version string

	// K8sVersion is the Kubernetes version
	K8sVersion string

	// FileCount is the total number of files in the bundle
	FileCount int

	// TotalSize is the uncompressed size of the bundle in bytes
	TotalSize int64

	// BundleType identifies the format (e.g., "rke2-support-bundle")
	BundleType string
}

// PodInfo contains metadata about a pod found in the bundle.
type PodInfo struct {
	// Namespace is the Kubernetes namespace
	Namespace string

	// Name is the pod name
	Name string

	// Containers is a list of container names in this pod
	Containers []string

	// HasCurrentLogs indicates if current logs are available
	HasCurrentLogs bool

	// HasPreviousLogs indicates if previous (crashed) logs are available
	HasPreviousLogs bool
}

// LogFileInfo contains metadata about a log file in the bundle.
type LogFileInfo struct {
	// Path is the relative path within the bundle
	Path string

	// Type indicates the log type (pod, system, journald)
	Type LogType

	// Namespace for pod logs
	Namespace string

	// PodName for pod logs
	PodName string

	// ContainerName for pod logs
	ContainerName string

	// IsPrevious indicates if this is a -previous log (crashed container)
	IsPrevious bool

	// Size is the file size in bytes
	Size int64

	// LineCount is an estimate of log lines (if parsed)
	LineCount int
}

// LogType identifies different types of log files in a bundle.
type LogType string

const (
	// LogTypePod represents pod container logs
	LogTypePod LogType = "pod"

	// LogTypeSystem represents system logs (syslog, kern.log)
	LogTypeSystem LogType = "system"

	// LogTypeJournald represents systemd journal logs
	LogTypeJournald LogType = "journald"

	// LogTypeContainerd represents containerd logs
	LogTypeContainerd LogType = "containerd"

	// LogTypeKubelet represents kubelet logs
	LogTypeKubelet LogType = "kubelet"
)

// BundleFormat identifies the type of support bundle.
type BundleFormat string

const (
	// FormatRKE2 represents an RKE2 support bundle
	FormatRKE2 BundleFormat = "rke2-support-bundle"

	// FormatKubectl represents a kubectl cluster-info dump
	FormatKubectl BundleFormat = "kubectl-cluster-info"

	// FormatUnknown represents an unrecognized bundle format
	FormatUnknown BundleFormat = "unknown"
)

// ImportOptions contains configuration for bundle import.
type ImportOptions struct {
	// Path is the path to the bundle tar.gz file
	Path string

	// MaxSize is the maximum allowed bundle size in bytes (0 = unlimited)
	MaxSize int64

	// KeepExtracted keeps the extracted directory after processing
	KeepExtracted bool

	// ExtractTo specifies a custom extraction directory (empty = temp)
	ExtractTo string

	// Verbose enables detailed error messages for debugging
	Verbose bool
}

// DefaultMaxBundleSize is 50MB by default to handle typical RKE2 log bundles.
// Increased from 10MB based on real-world bundle sizes (often 20-40MB uncompressed).
// Users can override with --limit flag for larger bundles.
const DefaultMaxBundleSize int64 = 50 * 1024 * 1024 // 50MB
