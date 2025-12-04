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

	return health, nil
}
