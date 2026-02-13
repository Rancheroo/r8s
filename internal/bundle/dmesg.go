package bundle

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// DMesgOOMKill represents an OOM kill event detected in dmesg
type DMesgOOMKill struct {
	Timestamp   string // Kernel timestamp (not wall clock)
	VictimPID   int
	VictimName  string // Process/command name that was killed
	OOMScore    int
	TotalRAM    string // Human readable like "16384MB"
	KilledAt    string // When the kill happened
	ContainerID string // If detectable from process name
}

// DMesgAnalysis contains analysis results from dmesg
type DMesgAnalysis struct {
	OOMKills        []DMesgOOMKill
	MemoryPressure  bool // System-wide memory pressure detected
	KernelWarnings  []string
}

// ParseDMesg parses dmesg output from bundle
// Bundle path: systeminfo/dmesg
func ParseDMesg(extractPath string) (*DMesgAnalysis, error) {
	bundleRoot := getBundleRoot(extractPath)
	dmesgPath := filepath.Join(bundleRoot, "systeminfo", "dmesg")

	file, err := os.Open(dmesgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open dmesg: %w", err)
	}
	defer file.Close()

	analysis := &DMesgAnalysis{
		OOMKills:       make([]DMesgOOMKill, 0),
		KernelWarnings: make([]string, 0),
	}

	scanner := bufio.NewScanner(file)
	
	// OOM kill patterns
	// Example: [1234567.890123] Out of memory: Killed process 12345 (nginx) total-vm:131072kB, anon-rss:65536kB, file-rss:0kB, shmem-rss:0kB, UID:0 pgtables:128kB oom_score_adj:0
	oomKillRegex := regexp.MustCompile(`\[\s*([\d.]+)\s*\]\s*Out of memory:\s*Killed process\s+(\d+)\s+\(([^)]+)\).*oom_score_adj:(-?\d+)`)
	
	// Memory pressure pattern
	// Example: [1234567.890123] Memory cgroup out of memory: Kill process 12345 (nginx) score 999 or sacrifice child
	cgroupOOMRegex := regexp.MustCompile(`\[\s*([\d.]+)\s*\]\s*Memory cgroup out of memory`)
	
	// Process killed summary
	// Example: [1234567.890123] oom-kill:constraint=CONSTRAINT_NONE,nodemask=(null),cpuset=cri-containerd-abc123.scope,mems_allowed=0,oom_memcg=/system.slice/containerd.service,task_memcg=/kubepods.slice/...,task=nginx,pid=12345,uid=0
	oomKillDetailRegex := regexp.MustCompile(`\[\s*([\d.]+)\s*\]\s*oom-kill:.*task=([^,]+),pid=(\d+)`)

	for scanner.Scan() {
		line := scanner.Text()

		// Check for OOM kill
		if matches := oomKillRegex.FindStringSubmatch(line); matches != nil {
			kill := DMesgOOMKill{
				Timestamp:  matches[1],
				VictimPID:  parseIntSafe(matches[2]),
				VictimName: matches[3],
				OOMScore:   parseIntSafe(matches[4]),
			}
			analysis.OOMKills = append(analysis.OOMKills, kill)
			continue
		}

		// Check for cgroup OOM (container-level)
		if cgroupOOMRegex.MatchString(line) {
			analysis.MemoryPressure = true
			continue
		}

		// Check for OOM kill detail line
		if matches := oomKillDetailRegex.FindStringSubmatch(line); matches != nil {
			// Extract container ID if present
			taskName := matches[2]
			if strings.Contains(taskName, "cri-containerd-") || strings.Contains(taskName, "cri-o-") {
				// Extract container runtime info
				analysis.KernelWarnings = append(analysis.KernelWarnings, 
					fmt.Sprintf("Container runtime OOM at %s: %s", matches[1], taskName))
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading dmesg: %w", err)
	}

	return analysis, nil
}

// HasOOMKills returns true if any OOM kills were detected
func (d *DMesgAnalysis) HasOOMKills() bool {
	return len(d.OOMKills) > 0
}

// GetOOMKillSummary returns a human-readable summary
func (d *DMesgAnalysis) GetOOMKillSummary() string {
	if !d.HasOOMKills() {
		return "No OOM kills detected in dmesg"
	}

	var parts []string
	parts = append(parts, fmt.Sprintf("Detected %d OOM kill(s):", len(d.OOMKills)))
	
	for _, kill := range d.OOMKills {
		parts = append(parts, fmt.Sprintf("  - %s (PID %d, score %d)", 
			kill.VictimName, kill.VictimPID, kill.OOMScore))
	}

	if d.MemoryPressure {
		parts = append(parts, "\nSystem memory pressure detected (cgroup OOM)")
	}

	return strings.Join(parts, "\n")
}

// CorrelateWithPods attempts to match dmesg OOM kills with pod events
// Returns a map of pod names that were likely OOM killed
func (d *DMesgAnalysis) CorrelateWithPods(podEvents map[string]int) map[string]bool {
	correlated := make(map[string]bool)

	for _, kill := range d.OOMKills {
		// Try to match process name with pod/container names
		// Common patterns: "nginx", "java", "python", etc.
		victimLower := strings.ToLower(kill.VictimName)
		
		for podName := range podEvents {
			podLower := strings.ToLower(podName)
			// Simple heuristic: if victim name appears in pod name
			if strings.Contains(podLower, victimLower) || strings.Contains(victimLower, podLower) {
				correlated[podName] = true
			}
		}
	}

	return correlated
}

func parseIntSafe(s string) int {
	var n int
	_, _ = fmt.Sscanf(s, "%d", &n) // Tolerant parsing - ignore errors, return 0 on failure
	return n
}
