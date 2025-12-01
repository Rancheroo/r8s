package bundle

import (
	"fmt"
	"os"
)

// Load loads a bundle from either a tar.gz archive or an extracted directory.
// It automatically detects the input type and validates the bundle structure.
//
// This function now delegates to LoadFromPath which handles both:
// - Compressed archives (.tar.gz, .tgz)
// - Extracted directories
func Load(opts ImportOptions) (*Bundle, error) {
	// Set default max size if not specified
	if opts.MaxSize == 0 {
		opts.MaxSize = DefaultMaxBundleSize
	}

	// Use the new bulletproof loader that handles both archives and directories
	return LoadFromPath(opts.Path, opts)
}

// Close is a no-op since bundles are always pre-extracted directories.
// Kept for backwards compatibility.
func (b *Bundle) Close() error {
	// No cleanup needed - users manage their own extracted directories
	return nil
}

// GetPod returns pod information by namespace and name.
func (b *Bundle) GetPod(namespace, name string) *PodInfo {
	for i := range b.Pods {
		if b.Pods[i].Namespace == namespace && b.Pods[i].Name == name {
			return &b.Pods[i]
		}
	}
	return nil
}

// GetLogFile returns log file information by path.
func (b *Bundle) GetLogFile(path string) *LogFileInfo {
	for i := range b.LogFiles {
		if b.LogFiles[i].Path == path {
			return &b.LogFiles[i]
		}
	}
	return nil
}

// ReadLogFile reads the contents of a log file from the bundle.
func (b *Bundle) ReadLogFile(logFile *LogFileInfo) ([]byte, error) {
	if logFile == nil {
		return nil, fmt.Errorf("log file info is nil")
	}
	return os.ReadFile(logFile.Path)
}

// Summary returns a human-readable summary of the bundle.
func (b *Bundle) Summary() string {
	return fmt.Sprintf(
		"Bundle: %s\nNode: %s\nRKE2: %s\nK8s: %s\nFiles: %d\nPods: %d\nLogs: %d",
		b.Manifest.NodeName,
		b.Manifest.NodeName,
		b.Manifest.RKE2Version,
		b.Manifest.K8sVersion,
		b.Manifest.FileCount,
		len(b.Pods),
		len(b.LogFiles),
	)
}
