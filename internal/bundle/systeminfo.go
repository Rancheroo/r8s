package bundle

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// SystemHealthInfo contains parsed system health information
type SystemHealthInfo struct {
	MemoryUsedPercent float64
	DiskUsedPercent   float64
	VirtType          string // Virtualization type from systemd-detect-virt (e.g., kvm, vmware, docker, lxc, wsl, none)
}

// ParseSystemHealth parses system info files from bundle
func ParseSystemHealth(extractPath string) (*SystemHealthInfo, error) {
	bundleRoot := getBundleRoot(extractPath)
	systeminfoDir := filepath.Join(bundleRoot, "systeminfo")

	health := &SystemHealthInfo{}

	// Parse memory usage from freem file
	// Format: "Mem:      total    used    free   shared  buff/cache   available"
	freemPath := filepath.Join(systeminfoDir, "freem")
	if content, err := os.ReadFile(freemPath); err == nil {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "Mem:") {
				fields := strings.Fields(line)
				if len(fields) >= 3 {
					total, _ := strconv.ParseFloat(fields[1], 64)
					used, _ := strconv.ParseFloat(fields[2], 64)
					if total > 0 {
						health.MemoryUsedPercent = (used / total) * 100
					}
				}
				break
			}
		}
	}

	// Parse disk usage from dfh file
	// Format: "Filesystem      Size  Used Avail Use% Mounted on"
	// Look for root filesystem (/)
	dfhPath := filepath.Join(systeminfoDir, "dfh")
	if content, err := os.ReadFile(dfhPath); err == nil {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			if strings.HasSuffix(strings.TrimSpace(line), " /") ||
				strings.Contains(line, " / ") {
				fields := strings.Fields(line)
				if len(fields) >= 5 {
					// Use% field is 4th field (0-indexed)
					usePercent := strings.TrimSuffix(fields[4], "%")
					if percent, err := strconv.ParseFloat(usePercent, 64); err == nil {
						health.DiskUsedPercent = percent
						break
					}
				}
			}
		}
	}

	// Parse virtualization type from systemd-detect-virt file
	// Output is a single line: "kvm", "vmware", "docker", "lxc", "wsl", "none", or error
	virtPath := filepath.Join(systeminfoDir, "systemd-detect-virt")
	if content, err := os.ReadFile(virtPath); err == nil {
		virtType := strings.TrimSpace(string(content))
		// Clean up any error messages - only take first line if multiple
		lines := strings.Split(virtType, "\n")
		if len(lines) > 0 {
			// Only accept valid virt types (alphanumeric/hyphen), ignore errors
			firstLine := strings.ToLower(strings.TrimSpace(lines[0]))
			// Check if it looks like a valid virt type (no spaces, reasonable length)
			if len(firstLine) > 0 && len(firstLine) < 50 && !strings.Contains(firstLine, " ") {
				health.VirtType = firstLine
			}
		}
	}

	return health, nil
}
