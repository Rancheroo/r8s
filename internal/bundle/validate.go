package bundle

import (
	"fmt"
	"os"
	"path/filepath"
)

// ValidateBundle checks if the given path points to a valid RKE2 bundle
// Returns error with helpful message if bundle structure is invalid
func ValidateBundle(path string) error {
	// Check if path exists
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("bundle path does not exist: %s", path)
		}
		return fmt.Errorf("cannot access bundle path: %w", err)
	}

	// Must be a directory
	if !info.IsDir() {
		return fmt.Errorf("bundle path is not a directory: %s\n\n"+
			"Hint: Did you forget to extract the bundle?\n"+
			"  tar -xzf your-bundle.tar.gz", path)
	}

	// Check for required subdirectories (at least one must exist)
	// Support RKE2, K3s, and kubectl-only bundles
	requiredDirs := []string{"rke2", "k3s", "kubectl"}
	hasRequired := false

	for _, dir := range requiredDirs {
		fullPath := filepath.Join(path, dir)
		if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
			hasRequired = true
			break
		}
	}

	if !hasRequired {
		return fmt.Errorf("not a valid support bundle at: %s\n\n"+
			"Expected structure:\n"+
			"  bundle-dir/\n"+
			"    rke2/ or k3s/ or kubectl/  (required)\n"+
			"    etcd/                      (optional)\n"+
			"    journald/                  (optional)\n"+
			"    systeminfo/                (optional)\n\n"+
			"Hint: Point to the extracted bundle directory, not the tarball", path)
	}

	return nil
}
