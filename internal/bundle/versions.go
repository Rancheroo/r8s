package bundle

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// VersionInfo holds parsed data from the root-level versions file
type VersionInfo struct {
	// System info
	CollectionDate string
	MemoryTotal    string
	MemoryUsed     string
	KernelVersion  string
	OSName         string
	Hostname       string

	// RKE2/K3s version
	DistroVersion string // e.g., "rke2 version v1.33.7+rke2r1"

	// Container images grouped by category
	RKE2Images   []ContainerImage // RKE2 system images
	CattleImages []ContainerImage // Rancher/cattle images
	CustomImages []ContainerImage // Other user workloads

	// Helm releases
	HelmReleases []HelmRelease
}

// ContainerImage represents a container image running in the cluster
type ContainerImage struct {
	PodName string
	Image   string
}

// HelmRelease represents a Helm chart deployment
type HelmRelease struct {
	Name           string
	Chart          string
	Version        string
	ReleaseName    string
	ReleaseVersion string
	Status         string
}

// ParseVersions parses the root-level versions file from a bundle
func ParseVersions(bundleRoot string) (*VersionInfo, error) {
	versionsPath := filepath.Join(bundleRoot, "versions")

	file, err := os.Open(versionsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open versions file: %w", err)
	}
	defer file.Close()

	info := &VersionInfo{}
	scanner := bufio.NewScanner(file)

	// State machine for parsing sections
	section := "header"

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Skip empty lines
		if trimmed == "" {
			continue
		}

		// Skip table headers (but use them for section detection)
		isHeaderLine := false
		switch trimmed {
		case "NAME", "NAME IMAGE",
			"NAME                           CHART                          VERSION     RELEASE NAME                   RELEASE VERSION   STATUS":
			isHeaderLine = true
		default:
			// Check for shorter header match
			if strings.HasPrefix(trimmed, "NAME ") && strings.Contains(trimmed, "CHART") && strings.Contains(trimmed, "VERSION") {
				isHeaderLine = true
			}
		}

		if isHeaderLine {
			// Use headers to detect section changes
			if strings.Contains(trimmed, "CHART") && strings.Contains(trimmed, "VERSION") {
				section = "helm"
			} else if trimmed == "NAME IMAGE" || trimmed == "NAME" {
				// Stay in current section, just a separator
			}
			continue
		}

		// Detect other section changes
		switch {
		case strings.Contains(line, "rke2 version") || strings.Contains(line, "k3s version"):
			info.DistroVersion = trimmed
			section = "images"
			continue
		}

		// Parse based on current section
		switch section {
		case "header":
			parseHeaderLine(info, line)
		case "images":
			parseImageLine(info, line)
		case "helm":
			parseHelmLine(info, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading versions file: %w", err)
	}

	return info, nil
}

// parseHeaderLine extracts system info from header lines
func parseHeaderLine(info *VersionInfo, line string) {
	trimmed := strings.TrimSpace(line)

	// Collection date (first line typically)
	if info.CollectionDate == "" && !strings.Contains(line, "Mem:") &&
		!strings.HasPrefix(line, "Linux") && !strings.HasPrefix(line, "Swap:") &&
		!strings.Contains(line, "total") && !strings.Contains(line, "version") {
		info.CollectionDate = trimmed
		return
	}

	// Memory: Mem: 3.8Gi 2.0Gi ...
	if strings.HasPrefix(trimmed, "Mem:") {
		fields := strings.Fields(trimmed)
		if len(fields) >= 3 {
			info.MemoryTotal = fields[1]
			info.MemoryUsed = fields[2]
		}
		return
	}

	// Linux kernel line: Linux <hostname> <kernel>...
	if strings.HasPrefix(trimmed, "Linux ") {
		fields := strings.Fields(trimmed)
		if len(fields) >= 2 {
			info.Hostname = fields[1]
		}
		if len(fields) >= 3 {
			info.KernelVersion = fields[2]
		}
		return
	}

	// OS Name: Ubuntu 24.04.3 LTS
	osRe := regexp.MustCompile(`^(Ubuntu|CentOS|RHEL|SLES|Debian|Alpine)`)
	if osRe.MatchString(trimmed) {
		info.OSName = trimmed
	}
}

// parseImageLine extracts container image info
func parseImageLine(info *VersionInfo, line string) {
	trimmed := strings.TrimSpace(line)

	// Image lines have format: <pod-name> <image>
	// With varying whitespace (columnar format)
	fields := strings.Fields(trimmed)
	if len(fields) < 2 {
		return
	}

	podName := fields[0]
	image := fields[len(fields)-1]

	// Skip metadata lines
	if strings.HasPrefix(podName, "go") ||
		!strings.Contains(image, "/") ||
		(!strings.Contains(image, ":") && !strings.Contains(image, "v")) {
		return
	}

	// Skip if looks like a header or partial line
	if podName == "NAME" || image == "IMAGE" {
		return
	}

	img := ContainerImage{
		PodName: podName,
		Image:   image,
	}

	// Categorize based on pod/image name
	switch {
	case strings.Contains(podName, "cattle-") || strings.Contains(podName, "rancher-"):
		info.CattleImages = append(info.CattleImages, img)
	case strings.HasPrefix(podName, "helm-install-"):
		// Skip helm-install jobs
		return
	default:
		info.RKE2Images = append(info.RKE2Images, img)
	}
}

// parseHelmLine extracts Helm release info
func parseHelmLine(info *VersionInfo, line string) {
	trimmed := strings.TrimSpace(line)
	fields := strings.Fields(trimmed)

	// Helm table format: NAME CHART VERSION RELEASE_NAME RELEASE_VERSION STATUS
	// Need at least 6 fields
	if len(fields) < 6 {
		return
	}

	release := HelmRelease{
		Name:           fields[0],
		Chart:          fields[1],
		Version:        fields[2],
		ReleaseName:    fields[3],
		ReleaseVersion: fields[4],
		Status:         fields[5],
	}

	info.HelmReleases = append(info.HelmReleases, release)
}
