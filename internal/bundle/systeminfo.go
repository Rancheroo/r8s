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

	// Parse memory usage from memory or freem file
	// New format uses "memory" file, old format uses "freem"
	// Try "memory" first, fall back to "freem" for backward compatibility
	if err := parseMemoryFile(systeminfoDir, health); err == nil {
		// Successfully parsed memory info
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

// parseMemoryFile attempts to parse memory info from either "memory" (new format) or "freem" (old format)
func parseMemoryFile(systeminfoDir string, health *SystemHealthInfo) error {
	// Try "memory" file first (new v1.1+ format), then fall back to "freem" (old format)
	memoryPaths := []string{
		filepath.Join(systeminfoDir, "memory"),
		filepath.Join(systeminfoDir, "freem"),
	}

	for _, memPath := range memoryPaths {
		content, err := os.ReadFile(memPath)
		if err != nil {
			continue // Try next path
		}

		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "Mem:") {
				fields := strings.Fields(line)
				if len(fields) >= 3 {
					total := parseMemoryValue(fields[1])
					used := parseMemoryValue(fields[2])
					if total > 0 {
						health.MemoryUsedPercent = (used / total) * 100
					}
				}
				return nil // Successfully parsed
			}
		}
		// File exists but no Mem: line found - this is still a successful read
		return nil
	}

	return os.ErrNotExist
}

// parseMemoryValue parses a memory value that may have units (e.g., "3.8Gi", "3915", "2.0Gi")
func parseMemoryValue(s string) float64 {
	s = strings.TrimSpace(s)

	// Handle unit suffixes
	multiplier := 1.0
	if strings.HasSuffix(s, "Gi") {
		multiplier = 1.0 // Gi is our base unit
		s = strings.TrimSuffix(s, "Gi")
	} else if strings.HasSuffix(s, "Mi") {
		multiplier = 1.0 / 1024.0 // Convert Mi to Gi
		s = strings.TrimSuffix(s, "Mi")
	} else if strings.HasSuffix(s, "Ki") {
		multiplier = 1.0 / (1024.0 * 1024.0) // Convert Ki to Gi
		s = strings.TrimSuffix(s, "Ki")
	} else if strings.HasSuffix(s, "B") {
		// Bytes to GiB (e.g., "1024B" -> 1024 / 1024^3)
		multiplier = 1.0 / (1024.0 * 1024.0 * 1024.0)
		s = strings.TrimSuffix(s, "B")
	}

	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return value * multiplier
}
