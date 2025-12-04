package bundle

import (
	"os"
	"path/filepath"
	"strings"
)

// EtcdHealthInfo contains parsed etcd health information
type EtcdHealthInfo struct {
	Healthy    bool
	HasAlarms  bool
	AlarmType  string
	AlarmCount int
}

// ParseEtcdHealth parses etcd health files from bundle
func ParseEtcdHealth(extractPath string) (*EtcdHealthInfo, error) {
	bundleRoot := getBundleRoot(extractPath)
	etcdDir := filepath.Join(bundleRoot, "etcd")

	health := &EtcdHealthInfo{
		Healthy: true, // Assume healthy unless proven otherwise
	}

	// Check for alarms in alarmlist file
	alarmPath := filepath.Join(etcdDir, "alarmlist")
	if content, err := os.ReadFile(alarmPath); err == nil {
		alarmText := strings.TrimSpace(string(content))
		// If file has content beyond just "memberID:" headers, we have alarms
		if alarmText != "" && !strings.HasPrefix(alarmText, "memberID:") {
			health.HasAlarms = true
			// Try to parse alarm type (format: "memberID:123 alarm:NOSPACE")
			for _, line := range strings.Split(alarmText, "\n") {
				if strings.Contains(line, "alarm:") {
					parts := strings.Split(line, "alarm:")
					if len(parts) > 1 {
						health.AlarmType = strings.TrimSpace(parts[1])
						health.AlarmCount++
					}
				}
			}
			if health.AlarmType == "" {
				health.AlarmType = "UNKNOWN"
			}
		}
	}

	// Check endpoint health
	healthPath := filepath.Join(etcdDir, "endpointhealth")
	if content, err := os.ReadFile(healthPath); err == nil {
		healthText := strings.ToLower(string(content))
		// Look for "is unhealthy" or "health: false"
		if strings.Contains(healthText, "unhealthy") ||
			strings.Contains(healthText, "health: false") ||
			strings.Contains(healthText, "\"health\":false") {
			health.Healthy = false
		}
	}

	return health, nil
}
