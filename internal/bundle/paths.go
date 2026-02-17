package bundle

import "path/filepath"

// PathResolver abstracts distro-specific paths (RKE2, K3s, etc.)
type PathResolver interface {
	// GetDistro returns the distro identifier
	GetDistro() string

	// GetKubectlDir returns the path to kubectl output directory
	GetKubectlDir() string

	// GetPodLogsDir returns the path to pod logs directory
	GetPodLogsDir() string

	// GetPodManifestsDir returns the path to pod manifests directory
	GetPodManifestsDir() string

	// GetPodDescribeDir returns the path to pod describe output directory
	GetPodDescribeDir() string

	// GetAgentLogsDir returns the path to agent logs directory
	GetAgentLogsDir() string

	// GetEtcdDir returns the path to etcd data directory
	GetEtcdDir() string

	// GetVersionFile returns the path to version file
	GetVersionFile() string

	// GetJournaldPaths returns possible paths for journald logs
	GetJournaldPaths() []string
}

// rke2PathResolver implements PathResolver for RKE2
type rke2PathResolver struct {
	bundleRoot string
}

// NewRKE2PathResolver creates an RKE2 path resolver
func NewRKE2PathResolver(bundleRoot string) PathResolver {
	return &rke2PathResolver{bundleRoot: bundleRoot}
}

func (r *rke2PathResolver) GetDistro() string {
	return "rke2"
}

func (r *rke2PathResolver) GetKubectlDir() string {
	return filepath.Join(r.bundleRoot, "rke2", "kubectl")
}

func (r *rke2PathResolver) GetPodLogsDir() string {
	return filepath.Join(r.bundleRoot, "rke2", "podlogs")
}

func (r *rke2PathResolver) GetPodManifestsDir() string {
	return filepath.Join(r.bundleRoot, "rke2", "pod-manifests")
}

func (r *rke2PathResolver) GetPodDescribeDir() string {
	return filepath.Join(r.bundleRoot, "rke2", "kubectl", "poddescribe")
}

func (r *rke2PathResolver) GetAgentLogsDir() string {
	return filepath.Join(r.bundleRoot, "rke2", "agent-logs")
}

func (r *rke2PathResolver) GetEtcdDir() string {
	return filepath.Join(r.bundleRoot, "rke2", "etcd")
}

func (r *rke2PathResolver) GetVersionFile() string {
	return filepath.Join(r.bundleRoot, "rke2", "version")
}

func (r *rke2PathResolver) GetJournaldPaths() []string {
	return []string{
		filepath.Join(r.bundleRoot, "systemlogs", "journald-rke2-server"),
		filepath.Join(r.bundleRoot, "systemlogs", "journald-rke2-agent"),
		filepath.Join(r.bundleRoot, "rke2", "agent-logs", "rke2-server"),
		filepath.Join(r.bundleRoot, "rke2", "agent-logs", "rke2-agent"),
		filepath.Join(r.bundleRoot, "systemlogs", "syslog"),
	}
}

// k3sPathResolver implements PathResolver for K3s
type k3sPathResolver struct {
	bundleRoot string
}

// NewK3sPathResolver creates a K3s path resolver
func NewK3sPathResolver(bundleRoot string) PathResolver {
	return &k3sPathResolver{bundleRoot: bundleRoot}
}

func (k *k3sPathResolver) GetDistro() string {
	return "k3s"
}

func (k *k3sPathResolver) GetKubectlDir() string {
	return filepath.Join(k.bundleRoot, "k3s", "kubectl")
}

func (k *k3sPathResolver) GetPodLogsDir() string {
	return filepath.Join(k.bundleRoot, "k3s", "podlogs")
}

func (k *k3sPathResolver) GetPodManifestsDir() string {
	return filepath.Join(k.bundleRoot, "k3s", "pod-manifests")
}

func (k *k3sPathResolver) GetPodDescribeDir() string {
	return filepath.Join(k.bundleRoot, "k3s", "kubectl", "poddescribe")
}

func (k *k3sPathResolver) GetAgentLogsDir() string {
	return filepath.Join(k.bundleRoot, "k3s", "agent-logs")
}

func (k *k3sPathResolver) GetEtcdDir() string {
	return filepath.Join(k.bundleRoot, "k3s", "etcd")
}

func (k *k3sPathResolver) GetVersionFile() string {
	return filepath.Join(k.bundleRoot, "k3s", "version")
}

func (k *k3sPathResolver) GetJournaldPaths() []string {
	return []string{
		filepath.Join(k.bundleRoot, "systemlogs", "journald-k3s"),
		filepath.Join(k.bundleRoot, "k3s", "agent-logs", "k3s"),
		filepath.Join(k.bundleRoot, "systemlogs", "syslog"),
	}
}

// NewPathResolver creates the appropriate PathResolver based on format
func NewPathResolver(bundleRoot string, format BundleFormat) PathResolver {
	switch format {
	case FormatK3s:
		return NewK3sPathResolver(bundleRoot)
	case FormatRKE2:
		return NewRKE2PathResolver(bundleRoot)
	default:
		// Default to RKE2 for backward compatibility
		return NewRKE2PathResolver(bundleRoot)
	}
}
