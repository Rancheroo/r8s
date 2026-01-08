package bundle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// EtcdHealthInfo contains parsed etcd health information (legacy)
type EtcdHealthInfo struct {
	Healthy    bool
	HasAlarms  bool
	AlarmType  string
	AlarmCount int
}

// EtcdMember represents a single etcd cluster member
type EtcdMember struct {
	ID         string
	State      string // "started"
	Name       string
	PeerURLs   string
	ClientURLs string
	IsLearner  bool
}

// EtcdDetails contains comprehensive etcd cluster information
type EtcdDetails struct {
	// Health status (from alarmlist and endpointhealth)
	Healthy    bool
	HasAlarms  bool
	AlarmType  string
	AlarmCount int

	// Cluster membership (from memberlist)
	MemberCount int
	Members     []EtcdMember

	// Endpoint status (from endpointstatus)
	LeaderID    string
	DBSize      string // "50 MB" - human readable
	DBSizeBytes int64  // 52428800 - for calculations
	Version     string // "3.5.21"
	RaftTerm    int
	RaftIndex   int64

	// Computed recommendations
	NeedsCompaction  bool
	CompactionReason string
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

// ParseEtcdMemberList parses etcd/memberlist file
// Format: ID, state, name, peer-urls, client-urls, is-learner
// Example: 15e9d2d844399be2, started, w-guard-wg-cp-svtk6-lqtxw-fecd6ae7, https://134.199.165.191:2380, https://134.199.165.191:2379, false
func ParseEtcdMemberList(extractPath string) ([]EtcdMember, error) {
	bundleRoot := getBundleRoot(extractPath)
	path := filepath.Join(bundleRoot, "etcd/memberlist")

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	var members []EtcdMember

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Split CSV format
		fields := strings.Split(line, ",")
		if len(fields) < 6 {
			continue
		}

		member := EtcdMember{
			ID:         strings.TrimSpace(fields[0]),
			State:      strings.TrimSpace(fields[1]),
			Name:       strings.TrimSpace(fields[2]),
			PeerURLs:   strings.TrimSpace(fields[3]),
			ClientURLs: strings.TrimSpace(fields[4]),
			IsLearner:  strings.TrimSpace(fields[5]) == "true",
		}

		members = append(members, member)
	}

	return members, nil
}

// ParseEtcdEndpointStatus parses etcd/endpointstatus file
// Format: ASCII table with columns: ENDPOINT, ID, VERSION, DB SIZE, IS LEADER, etc.
func ParseEtcdEndpointStatus(extractPath string) (string, string, int64, string, int, int64, error) {
	bundleRoot := getBundleRoot(extractPath)
	path := filepath.Join(bundleRoot, "etcd/endpointstatus")

	content, err := os.ReadFile(path)
	if err != nil {
		return "", "", 0, "", 0, 0, err
	}

	lines := strings.Split(string(content), "\n")

	// Find the data row (skip header and separator lines)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "+") || strings.HasPrefix(line, "|") && strings.Contains(line, "ENDPOINT") {
			continue
		}

		// Parse data row: | endpoint | ID | VERSION | DB SIZE | IS LEADER | ...
		if strings.HasPrefix(line, "|") {
			fields := strings.Split(line, "|")
			if len(fields) < 8 {
				continue
			}

			// Extract fields (1-indexed because of leading |)
			leaderID := strings.TrimSpace(fields[2])  // ID column
			version := strings.TrimSpace(fields[3])   // VERSION column
			dbSize := strings.TrimSpace(fields[4])    // DB SIZE column
			isLeader := strings.TrimSpace(fields[5])  // IS LEADER column
			raftTerm := strings.TrimSpace(fields[7])  // RAFT TERM column
			raftIndex := strings.TrimSpace(fields[8]) // RAFT INDEX column

			// Parse DB size to bytes (e.g., "50 MB" -> 52428800)
			dbSizeBytes := parseDBSize(dbSize)

			// Parse raft term and index
			var term int
			var index int64
			fmt.Sscanf(raftTerm, "%d", &term)
			fmt.Sscanf(raftIndex, "%d", &index)

			// Only return leader ID if this is the leader
			if isLeader != "true" {
				leaderID = ""
			}

			return leaderID, dbSize, dbSizeBytes, version, term, index, nil
		}
	}

	return "", "", 0, "", 0, 0, nil
}

// parseDBSize converts "50 MB" to bytes
func parseDBSize(sizeStr string) int64 {
	sizeStr = strings.TrimSpace(sizeStr)
	parts := strings.Fields(sizeStr)
	if len(parts) != 2 {
		return 0
	}

	var size float64
	fmt.Sscanf(parts[0], "%f", &size)
	unit := strings.ToUpper(parts[1])

	switch unit {
	case "B":
		return int64(size)
	case "KB":
		return int64(size * 1024)
	case "MB":
		return int64(size * 1024 * 1024)
	case "GB":
		return int64(size * 1024 * 1024 * 1024)
	default:
		return 0
	}
}

// ParseEtcdDetails parses all etcd information from bundle
func ParseEtcdDetails(extractPath string) (*EtcdDetails, error) {
	details := &EtcdDetails{}

	// Parse basic health (backward compatibility)
	healthInfo, err := ParseEtcdHealth(extractPath)
	if err == nil {
		details.Healthy = healthInfo.Healthy
		details.HasAlarms = healthInfo.HasAlarms
		details.AlarmType = healthInfo.AlarmType
		details.AlarmCount = healthInfo.AlarmCount
	}

	// Parse member list
	members, err := ParseEtcdMemberList(extractPath)
	if err == nil {
		details.Members = members
		details.MemberCount = len(members)
	}

	// Parse endpoint status
	leaderID, dbSize, dbSizeBytes, version, raftTerm, raftIndex, err := ParseEtcdEndpointStatus(extractPath)
	if err == nil {
		details.LeaderID = leaderID
		details.DBSize = dbSize
		details.DBSizeBytes = dbSizeBytes
		details.Version = version
		details.RaftTerm = raftTerm
		details.RaftIndex = raftIndex
	}

	// Compute compaction recommendation
	// ETCD recommends compaction when DB exceeds 100MB for optimal performance
	const compactionThreshold = 100 * 1024 * 1024 // 100 MB
	if details.DBSizeBytes > compactionThreshold {
		details.NeedsCompaction = true
		details.CompactionReason = fmt.Sprintf("DB size %s exceeds recommended 100MB threshold", details.DBSize)
	}

	return details, nil
}
