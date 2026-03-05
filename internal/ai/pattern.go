// Package ai provides pattern matching for Kubernetes issues.
// Sprint 11: Pattern Engine v2 - Confidence scoring, correlations, root cause hints.
package ai

import (
	"fmt"
	"regexp"
	"strings"
)

// Confidence represents detection confidence levels
type Confidence string

const (
	ConfidenceCertain  Confidence = "certain"  // Direct evidence, unambiguous
	ConfidenceLikely   Confidence = "likely"   // Strong evidence, minor ambiguity
	ConfidencePossible Confidence = "possible" // Some evidence, needs verification
)

// PatternV2 represents a detection pattern with confidence and correlation support
// Sprint 11: Enhanced pattern schema for AI Intelligence
type PatternV2 struct {
	ID            string        `yaml:"id"`             // Unique identifier
	Name          string        `yaml:"name"`           // Human-readable name
	Category      string        `yaml:"category"`       // e.g., "OOM", "ImagePull", "etcd"
	Severity      Severity      `yaml:"severity"`       // Critical, Warning, Info
	Confidence    Confidence    `yaml:"confidence"`     // Certain, Likely, Possible
	Matchers      []Matcher     `yaml:"matchers"`       // Match criteria
	Correlations  []Correlation `yaml:"correlations"`   // Related patterns
	HintGenerator HintGenerator `yaml:"hint_generator"` // Root cause hint template
	Description   string        `yaml:"description"`    // What this pattern detects
}

// Matcher defines how to match a pattern
type Matcher struct {
	Type    string  `yaml:"type"`    // "keyword", "regex"
	Pattern string  `yaml:"pattern"` // Pattern string
	Weight  float64 `yaml:"weight"`  // 0.0-1.0, contribution to confidence
}

// Correlation links patterns together for root cause analysis
type Correlation struct {
	PatternID string `yaml:"pattern_id"` // Related pattern ID
	Message   string `yaml:"message"`    // Explanation of relationship
}

// HintGenerator produces root cause hints
type HintGenerator struct {
	Template   string   `yaml:"template"`   // Go template for hint
	Variables  []string `yaml:"variables"`  // Required variables
	Suggestion string   `yaml:"suggestion"` // Recommended fix
	Command    string   `yaml:"command"`    // kubectl command to run
	References []string `yaml:"references"` // Doc links
}

// Severity represents issue severity
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityWarning  Severity = "warning"
	SeverityInfo     Severity = "info"
)

// MatchResultV2 represents pattern matching outcome with correlation support
// Sprint 11: Enhanced match result with confidence levels and correlations
type MatchResultV2 struct {
	Matched     bool              // Whether pattern matched
	PatternID   string            // Pattern identifier
	PatternName string            // Human-readable name
	Severity    Severity          // Issue severity
	Confidence  Confidence        // Detection confidence (Certain/Likely/Possible)
	Message     string            // Detection message
	Resources   []Resource        // Affected resources
	Evidence    []string          // Evidence snippets
	Correlated  []string          // IDs of correlated patterns found
	Metadata    map[string]string // Additional context (for hint generation)
}

// Resource identifies a Kubernetes resource affected by a pattern
type Resource struct {
	Kind      string // e.g., "Pod", "Node"
	Name      string // Resource name
	Namespace string // Namespace (if applicable)
}

// MatcherV2 provides pattern matching for PatternV2 (Sprint 11)
type MatcherV2 struct {
	pattern PatternV2
}

// NewMatcherV2 creates a new v2 pattern matcher
func NewMatcherV2(p PatternV2) *MatcherV2 {
	return &MatcherV2{pattern: p}
}

// Match checks if content matches the v2 pattern
// Sprint 11: Supports multiple matchers and multiple occurrences
func (m *MatcherV2) Match(content string) []MatchResultV2 {
	var results []MatchResultV2
	contentLower := strings.ToLower(content)
	hasRegexMatch := false

	for _, matcher := range m.pattern.Matchers {
		if matcher.Type == "regex" {
			// Regex handling: find all non-overlapping matches
			re, err := regexp.Compile(matcher.Pattern)
			if err != nil {
				continue
			}

			matches := re.FindAllStringSubmatch(content, -1)
			if matches == nil {
				continue
			}

			if len(matches) > 0 {
				hasRegexMatch = true
			}

			for _, match := range matches {
				metadata := make(map[string]string)
				for i, name := range re.SubexpNames() {
					if i != 0 && name != "" && i < len(match) {
						metadata[name] = match[i]
					}
				}

				// Create a result for each match
				results = append(results, MatchResultV2{
					Matched:     true,
					PatternID:   m.pattern.ID,
					PatternName: m.pattern.Name,
					Severity:    m.pattern.Severity,
					Confidence:  m.pattern.Confidence, // Regex match is high confidence
					Message:     m.detectedMessageV2(),
					Evidence:    []string{match[0]}, // Full match text
					Metadata:    metadata,
				})
			}
		} else {
			// Keyword handling (legacy behavior: one match per file)
			// Skip if regex already matched to avoid duplicates with less info
			if hasRegexMatch {
				continue
			}

			if strings.Contains(contentLower, strings.ToLower(matcher.Pattern)) {
				// Only add if not already matched by regex to avoid duplicates
				// For now, simple append
				results = append(results, MatchResultV2{
					Matched:     true,
					PatternID:   m.pattern.ID,
					PatternName: m.pattern.Name,
					Severity:    m.pattern.Severity,
					Confidence:  m.pattern.Confidence,
					Message:     m.detectedMessageV2(),
					Evidence:    []string{m.extractEvidence(matcher, content)},
					Metadata:    make(map[string]string),
				})
				// Break after first keyword match to avoid flooding
				break
			}
		}
	}

	// If no matches found, return empty (but we need to change return type from struct to slice)
	if len(results) == 0 {
		return []MatchResultV2{{Matched: false}}
	}

	return results
}

// extractEvidence extracts matching evidence from content
func (m *MatcherV2) extractEvidence(matcher Matcher, content string) string {
	// Find the line containing the match
	lines := strings.Split(content, "\n")
	patternLower := strings.ToLower(matcher.Pattern)
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), patternLower) {
			// Return trimmed line as evidence
			trimmed := strings.TrimSpace(line)
			if len(trimmed) > 200 {
				return trimmed[:200] + "..."
			}
			return trimmed
		}
	}
	return matcher.Pattern
}

// detectedMessageV2 returns a human-readable detection message
func (m *MatcherV2) detectedMessageV2() string {
	return fmt.Sprintf("[%s] %s: %s",
		strings.ToUpper(string(m.pattern.Severity)),
		m.pattern.Name,
		m.pattern.Description)
}

// BuiltinPatternsV2 contains Sprint 11 pattern definitions with confidence and correlations
var BuiltinPatternsV2 = []PatternV2{
	{
		ID:         "oomkill-v2",
		Name:       "OOMKill Detected",
		Category:   "OOM",
		Severity:   SeverityCritical,
		Confidence: ConfidenceCertain,
		Matchers: []Matcher{
			{
				Type:    "regex",
				Pattern: `TaskId: (?P<ContainerName>\S+) Error: OOM Killed`,
				Weight:  1.0,
			},
			{
				Type:    "regex",
				Pattern: `Memory cgroup out of memory: Kill process \d+ \((?P<Process>\S+)\) score \d+ or sacrifice child`,
				Weight:  1.0,
			},
			{Type: "keyword", Pattern: "out of memory", Weight: 1.0},
			{Type: "keyword", Pattern: "oomkill", Weight: 1.0},
			{Type: "keyword", Pattern: "oom_kill_process", Weight: 1.0},
			{Type: "keyword", Pattern: "killed process", Weight: 0.8},
		},
		Description: "Container was killed due to memory limits",
		HintGenerator: HintGenerator{
			Template:   "Container {{.ContainerName}} was killed due to out-of-memory.",
			Suggestion: "Increase memory limit in pod spec or optimize application memory usage",
			Command:    "kubectl describe pod {{.PodName}} -n {{.Namespace}} | grep -A5 'Last State'",
			References: []string{"https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/"},
		},
	},
	{
		ID:         "imagepullbackoff-v2",
		Name:       "ImagePullBackOff",
		Category:   "Image",
		Severity:   SeverityWarning,
		Confidence: ConfidenceCertain,
		Matchers: []Matcher{
			{
				Type:    "regex",
				Pattern: `Failed to pull image "(?P<Image>[^"]+)": rpc error: code = (?P<Code>\S+) desc = (?P<PullError>.+)`,
				Weight:  1.0,
			},
			{Type: "keyword", Pattern: "imagepullbackoff", Weight: 1.0},
			{Type: "keyword", Pattern: "pull access denied", Weight: 1.0},
			{Type: "keyword", Pattern: "failed to pull image", Weight: 1.0},
			{Type: "keyword", Pattern: "image not found", Weight: 0.9},
		},
		Description: "Cannot pull container image from registry",
		HintGenerator: HintGenerator{
			Template:   "Failed to pull image '{{.Image}}'. {{.PullError}}",
			Suggestion: "Check image name/tag exists, registry credentials are configured, and network connectivity to registry",
			Command:    "kubectl describe pod {{.PodName}} -n {{.Namespace}} | grep -A10 'Events'",
			References: []string{"https://kubernetes.io/docs/concepts/containers/images/"},
		},
	},
	{
		ID:         "crashloopbackoff-v2",
		Name:       "CrashLoopBackOff",
		Category:   "Crash",
		Severity:   SeverityCritical,
		Confidence: ConfidenceCertain,
		Matchers: []Matcher{
			// Regex to capture details from kubectl get pods output
			{
				Type:    "regex",
				Pattern: `(?P<Namespace>\S+)\s+(?P<PodName>\S+)\s+\S+\s+CrashLoopBackOff\s+(?P<RestartCount>\d+)`,
				Weight:  1.0,
			},
			{Type: "keyword", Pattern: "crashloopbackoff", Weight: 1.0},
			{Type: "keyword", Pattern: "back-off restarting", Weight: 1.0},
			{Type: "keyword", Pattern: "crash loop", Weight: 1.0},
		},
		Correlations: []Correlation{
			{PatternID: "oomkill-v2", Message: "Crash loop may be caused by OOM kills"},
			{PatternID: "imagepullbackoff-v2", Message: "Crash loop may be caused by image pull failures"},
		},
		Description: "Container repeatedly crashing and restarting",
		HintGenerator: HintGenerator{
			Template:   "Pod {{.PodName}} in namespace {{.Namespace}} is in CrashLoopBackOff. Restarts: {{.RestartCount}}.",
			Suggestion: "Check container logs for application errors. Common causes: missing env vars, config errors, dependency failures",
			Command:    "kubectl logs {{.PodName}} -n {{.Namespace}} --previous",
			References: []string{"https://kubernetes.io/docs/tasks/debug/debug-application/debug-running-pod/"},
		},
	},
	// Sprint 11 Day 2: etcd Patterns
	{
		ID:         "etcd-corruption",
		Name:       "etcd Data Corruption",
		Category:   "etcd",
		Severity:   SeverityCritical,
		Confidence: ConfidenceCertain,
		Matchers: []Matcher{
			{
				Type:    "regex",
				Pattern: `etcdserver: (?P<ErrorType>mvcc: database space exceeded)`,
				Weight:  1.0,
			},
			{
				Type:    "regex",
				Pattern: `(?P<ErrorType>etcd data corruption)`,
				Weight:  1.0,
			},
			{Type: "keyword", Pattern: "etcdserver: mvcc: database space exceeded", Weight: 1.0},
			{Type: "keyword", Pattern: "etcd data corruption", Weight: 1.0},
			{Type: "keyword", Pattern: "etcdserver: corrupt", Weight: 1.0},
			{Type: "keyword", Pattern: "raft: log corruption", Weight: 1.0},
		},
		Description: "etcd data corruption or storage limit exceeded",
		HintGenerator: HintGenerator{
			Template:   "etcd corruption detected: {{.ErrorType}}. Cluster may be unstable or unable to write new data.",
			Suggestion: "For space exceeded: run etcd compaction and defrag. For corruption: restore from backup or rebuild cluster",
			Command:    "ETCDCTL_API=3 etcdctl defrag && ETCDCTL_API=3 etcdctl compact $(ETCDCTL_API=3 etcdctl endpoint status --write-out=json | jq -r '.[0].Status.header.revision')",
			References: []string{"https://etcd.io/docs/v3.5/op-guide/maintenance/", "https://docs.rke2.io/backup_restore/"},
		},
	},
	{
		ID:         "etcd-latency",
		Name:       "etcd High Latency",
		Category:   "etcd",
		Severity:   SeverityWarning,
		Confidence: ConfidenceLikely,
		Matchers: []Matcher{
			{
				Type:    "regex",
				Pattern: `etcd: (?P<LatencyType>[\w\s]+) took (?P<Duration>[\d\.]+m?s)`,
				Weight:  1.0,
			},
			{Type: "keyword", Pattern: "etcd: slow", Weight: 1.0},
			{Type: "keyword", Pattern: "took too long", Weight: 0.9},
			{Type: "keyword", Pattern: "etcd: read index took", Weight: 1.0},
			{Type: "keyword", Pattern: "etcd: apply entries took", Weight: 1.0},
			{Type: "keyword", Pattern: "disk io latency", Weight: 0.8},
		},
		Description: "etcd experiencing high latency, potential disk or network issues",
		HintGenerator: HintGenerator{
			Template:   "etcd latency detected: {{.LatencyType}}. Operations taking {{.Duration}}.",
			Suggestion: "Check disk I/O latency (should be <10ms), network connectivity between etcd nodes, and consider dedicated disk for etcd",
			Command:    "iostat -x 1 10 && ETCDCTL_API=3 etcdctl endpoint status --write-out=table",
			References: []string{"https://etcd.io/docs/v3.5/tuning/", "https://docs.rke2.io/security/hardening_guide/"},
		},
	},
	{
		ID:         "etcd-quorum",
		Name:       "etcd Quorum Loss",
		Category:   "etcd",
		Severity:   SeverityCritical,
		Confidence: ConfidenceCertain,
		Matchers: []Matcher{
			{Type: "keyword", Pattern: "etcdserver: no leader", Weight: 1.0},
			{Type: "keyword", Pattern: "etcd: lost quorum", Weight: 1.0},
			{Type: "keyword", Pattern: "raft: no elected leader", Weight: 1.0},
			{Type: "keyword", Pattern: "etcdserver: publish error", Weight: 0.9},
		},
		Description: "etcd cluster has lost quorum, control plane is down",
		HintGenerator: HintGenerator{
			Template:   "etcd quorum lost. Cluster may be down.",
			Suggestion: "Restore failed etcd nodes or restore cluster from snapshot. Check network connectivity between control plane nodes.",
			Command:    "ETCDCTL_API=3 etcdctl member list && ETCDCTL_API=3 etcdctl endpoint health --cluster",
			References: []string{"https://etcd.io/docs/v3.5/op-guide/recovery/", "https://docs.rke2.io/backup_restore/"},
		},
	},
	// Sprint 11 Day 2: Certificate Patterns
	{
		ID:         "certificate-expired",
		Name:       "Certificate Expired",
		Category:   "Certificate",
		Severity:   SeverityCritical,
		Confidence: ConfidenceCertain,
		Matchers: []Matcher{
			{
				Type:    "regex",
				Pattern: `Serving cert is expired: (?P<Cert>\S+)`,
				Weight:  1.0,
			},
			{Type: "keyword", Pattern: "certificate has expired", Weight: 1.0},
			{Type: "keyword", Pattern: "x509: certificate has expired", Weight: 1.0},
			{Type: "keyword", Pattern: "certificate is not valid", Weight: 0.9},
		},
		Correlations: []Correlation{
			{PatternID: "node-not-ready", Message: "Node may be NotReady due to expired certificate"},
		},
		Description: "Kubernetes certificate has expired",
		HintGenerator: HintGenerator{
			Template:   "Certificate expired{{if .Cert}}: {{.Cert}}{{end}}",
			Suggestion: "Approve pending CSR to renew certificate: 'kubectl get csr', then 'kubectl certificate approve <csr-name>'",
			Command:    "kubectl get csr && kubectl certificate approve $(kubectl get csr -o json | jq -r '.items[] | select(.status.conditions == null) | .metadata.name')",
			References: []string{"https://kubernetes.io/docs/tasks/tls/certificate-issue/", "https://docs.rke2.io/security/certificates/"},
		},
	},
	{
		ID:         "certificate-expiring",
		Name:       "Certificate Expiring Soon",
		Category:   "Certificate",
		Severity:   SeverityWarning,
		Confidence: ConfidenceLikely,
		Matchers: []Matcher{
			{Type: "keyword", Pattern: "certificate will expire", Weight: 1.0},
			{Type: "keyword", Pattern: "certificate expiring", Weight: 1.0},
			{Type: "regex", Pattern: `(?i)notafter.*202[0-5]`, Weight: 0.8}, // Old dates in 2020-2025 range
		},
		Description: "Kubernetes certificate is expiring soon",
		HintGenerator: HintGenerator{
			Template:   "Certificate is expiring soon.",
			Suggestion: "Plan certificate rotation. For automatic rotation, ensure node is running and can request new certificates from API server",
			Command:    "openssl x509 -in /var/lib/rancher/rke2/server/tls/client-ca.crt -noout -dates",
			References: []string{"https://docs.rke2.io/security/certificates/", "https://kubernetes.io/docs/tasks/tls/certificate-issue/"},
		},
	},
	{
		ID:         "certificate-invalid-ca",
		Name:       "Invalid Certificate Authority",
		Category:   "Certificate",
		Severity:   SeverityCritical,
		Confidence: ConfidenceCertain,
		Matchers: []Matcher{
			{Type: "keyword", Pattern: "x509: unknown authority", Weight: 1.0},
			{Type: "keyword", Pattern: "certificate signed by unknown authority", Weight: 1.0},
			{Type: "keyword", Pattern: "x509: certificate is not standards compliant", Weight: 1.0},
		},
		Correlations: []Correlation{
			{PatternID: "etcd-quorum", Message: "Certificate issues may prevent etcd communication"},
		},
		Description: "Certificate signed by unknown or untrusted CA",
		HintGenerator: HintGenerator{
			Template:   "Certificate authority mismatch. The certificate is not trusted.",
			Suggestion: "Check CA certificates are consistent across cluster. If CA was rotated, ensure all nodes have updated CA bundle.",
			Command:    "openssl s_client -connect localhost:6443 -CAfile /var/lib/rancher/rke2/server/tls/server-ca.crt 2>&1 | grep 'Verify return code'",
			References: []string{"https://docs.rke2.io/security/certificates/", "https://kubernetes.io/docs/setup/best-practices/certificates/"},
		},
	},
	// Sprint 11 Day 3: Networking Patterns
	{
		ID:         "dns-failure",
		Name:       "DNS Resolution Failure",
		Category:   "Networking",
		Severity:   SeverityWarning,
		Confidence: ConfidenceLikely,
		Matchers: []Matcher{
			{
				Type:    "regex",
				Pattern: `lookup (?P<Hostname>[\w\.\-]+)( on \S+)?(:)? (?P<Error>no such host|i/o timeout|server misbehaving|nxdomain)`,
				Weight:  1.0,
			},
			{Type: "keyword", Pattern: "dns resolution failed", Weight: 1.0},
			{Type: "keyword", Pattern: "lookup failed", Weight: 0.9},
			{Type: "keyword", Pattern: "could not resolve", Weight: 0.9},
			{Type: "keyword", Pattern: "name not known", Weight: 0.9},
			{Type: "keyword", Pattern: "failed to resolve dns", Weight: 1.0},
			{Type: "keyword", Pattern: "nxdomain", Weight: 0.9},
		},
		Correlations: []Correlation{
			{PatternID: "cni-error", Message: "DNS failures may be caused by CNI connectivity issues"},
		},
		Description: "DNS resolution failing for services or external hosts",
		HintGenerator: HintGenerator{
			Template:   "DNS resolution failed. Error: {{.Error}}",
			Suggestion: "Check CoreDNS pods are running in kube-system. Verify DNS configuration and network policies allow DNS traffic.",
			Command:    "kubectl get pods -n kube-system -l k8s-app=kube-dns && kubectl logs -n kube-system -l k8s-app=kube-dns --tail=50",
			References: []string{"https://kubernetes.io/docs/tasks/administer-cluster/dns-debugging-resolution/", "https://coredns.io/manual/troubleshooting/"},
		},
	},
	{
		ID:         "cni-error",
		Name:       "CNI Plugin Error",
		Category:   "Networking",
		Severity:   SeverityCritical,
		Confidence: ConfidenceCertain,
		Matchers: []Matcher{
			{Type: "keyword", Pattern: "cni plugin failed", Weight: 1.0},
			{Type: "keyword", Pattern: "failed to set up sandbox", Weight: 1.0},
			{Type: "keyword", Pattern: "networkplugin cni", Weight: 1.0},
			{Type: "keyword", Pattern: "cni setup error", Weight: 1.0},
			{Type: "keyword", Pattern: "cni configuration error", Weight: 1.0},
			// Note: Removed "calico", "flannel", "cilium" - they match pod names, not errors
		},
		Correlations: []Correlation{
			{PatternID: "pod-stuck-pending", Message: "CNI errors prevent pods from starting"},
		},
		Description: "CNI (Container Network Interface) plugin errors preventing pod networking",
		HintGenerator: HintGenerator{
			Template:   "CNI plugin error detected.",
			Suggestion: "Check CNI daemonset pods and logs. Verify CNI configuration files in /etc/cni/net.d/",
			Command:    "kubectl get pods -n kube-system | grep -E 'calico|flannel|cilium|canal' && kubectl logs -n kube-system -l k8s-app=calico-node --tail=100",
			References: []string{"https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/network-plugins/", "https://projectcalico.docs.tigera.io/troubleshooting/"},
		},
	},
	{
		ID:         "connectivity-timeout",
		Name:       "Network Connectivity Timeout",
		Category:   "Networking",
		Severity:   SeverityWarning,
		Confidence: ConfidenceLikely,
		Matchers: []Matcher{
			{
				Type:    "regex",
				Pattern: `dial tcp (?P<Target>[\w\.\-:]+): i/o timeout`,
				Weight:  1.0,
			},
			{Type: "keyword", Pattern: "connection timed out", Weight: 1.0},
			{Type: "keyword", Pattern: "timeout awaiting response", Weight: 1.0},
			{Type: "keyword", Pattern: "i/o timeout", Weight: 1.0},
			{Type: "keyword", Pattern: "context deadline exceeded", Weight: 0.9},
			{Type: "keyword", Pattern: "no route to host", Weight: 0.9},
		},
		Description: "Network connections timing out, potential connectivity or firewall issues",
		HintGenerator: HintGenerator{
			Template:   "Connection timeout to {{.Target}}. No response received.",
			Suggestion: "Check network connectivity, firewall rules, and service endpoints. Verify target service is running and accessible.",
			Command:    "kubectl get endpoints {{.Service}} -n {{.Namespace}} && kubectl get svc {{.Service}} -n {{.Namespace}}",
			References: []string{"https://kubernetes.io/docs/tasks/debug/debug-application/debug-service/", "https://kubernetes.io/docs/concepts/services-networking/"},
		},
	},
	// Sprint 11 Day 3: Storage Patterns
	{
		ID:         "pvc-binding-failure",
		Name:       "PVC Binding Failure",
		Category:   "Storage",
		Severity:   SeverityWarning,
		Confidence: ConfidenceLikely,
		Matchers: []Matcher{
			{Type: "keyword", Pattern: "persistentvolumeclaim is not bound", Weight: 1.0},
			{Type: "keyword", Pattern: "pvc is not bound", Weight: 1.0},
			{Type: "keyword", Pattern: "failed to provision volume", Weight: 1.0},
			{Type: "keyword", Pattern: "no persistent volumes available", Weight: 1.0},
			{Type: "keyword", Pattern: "waiting for a volume", Weight: 0.9},
		},
		Correlations: []Correlation{
			{PatternID: "pod-stuck-pending", Message: "PVC binding failure prevents pod from starting"},
		},
		Description: "Persistent Volume Claim cannot bind to a storage volume",
		HintGenerator: HintGenerator{
			Template:   "PVC binding failed.",
			Suggestion: "Check storage class exists and has provisioner running. Verify PVs available or dynamic provisioning configured.",
			Command:    "kubectl get pvc {{.PVCName}} -n {{.Namespace}} && kubectl get storageclass && kubectl get pods -n kube-system | grep -E 'csi|provisioner'",
			References: []string{"https://kubernetes.io/docs/concepts/storage/persistent-volumes/", "https://kubernetes.io/docs/tasks/configure-pod-container/configure-persistent-volume-storage/"},
		},
	},
	{
		ID:         "storage-pressure",
		Name:       "Storage Pressure",
		Category:   "Storage",
		Severity:   SeverityCritical,
		Confidence: ConfidenceCertain,
		Matchers: []Matcher{
			{Type: "keyword", Pattern: "disk pressure", Weight: 1.0},
			{Type: "keyword", Pattern: "filesystem has no space left", Weight: 1.0},
			{Type: "keyword", Pattern: "no space left on device", Weight: 1.0},
			{Type: "keyword", Pattern: "InodePressure", Weight: 1.0},
			{Type: "keyword", Pattern: "runtime garbage collection failed", Weight: 0.8},
		},
		Correlations: []Correlation{
			{PatternID: "node-pressure", Message: "Storage pressure is a type of node pressure condition"},
		},
		Description: "Node or container experiencing disk space pressure",
		HintGenerator: HintGenerator{
			Template:   "Storage pressure detected.",
			Suggestion: "Free up disk space by cleaning images (crictl rmi), logs (truncate), or unused volumes. Check kubelet disk limits.",
			Command:    "df -h && crictl ps -a | wc -l && crictl rmi --prune && journalctl --vacuum-time=24h",
			References: []string{"https://kubernetes.io/docs/concepts/architecture/nodes/", "https://kubernetes.io/docs/tasks/administer-cluster/out-of-resource/"},
		},
	},
	// Sprint 11 Day 4: Node & Pod State Patterns
	{
		ID:         "node-pressure",
		Name:       "Node Pressure Conditions",
		Category:   "Node",
		Severity:   SeverityCritical,
		Confidence: ConfidenceCertain,
		Matchers: []Matcher{
			{Type: "keyword", Pattern: "diskpressure", Weight: 1.0},
			{Type: "keyword", Pattern: "memorypressure", Weight: 1.0},
			{Type: "keyword", Pattern: "pidpressure", Weight: 1.0},
			{Type: "keyword", Pattern: "outofdisk", Weight: 1.0},
			{Type: "keyword", Pattern: "node has insufficient", Weight: 0.9},
		},
		Correlations: []Correlation{
			{PatternID: "storage-pressure", Message: "Disk pressure often accompanies storage exhaustion"},
			{PatternID: "oomkill-v2", Message: "Memory pressure may cause OOM kills"},
		},
		Description: "Node experiencing pressure conditions (disk, memory, or PID)",
		HintGenerator: HintGenerator{
			Template:   "Node has pressure condition.",
			Suggestion: "For DiskPressure: free disk space. For MemoryPressure: stop non-critical pods or add nodes. For PIDPressure: check for PID leaks.",
			Command:    "kubectl describe node {{.NodeName}} | grep -A20 'Conditions' && df -h /var/lib/rancher",
			References: []string{"https://kubernetes.io/docs/concepts/architecture/nodes/", "https://kubernetes.io/docs/tasks/administer-cluster/out-of-resource/"},
		},
	},
	{
		ID:         "pod-stuck-pending",
		Name:       "Pod Stuck Pending",
		Category:   "Pod",
		Severity:   SeverityWarning,
		Confidence: ConfidenceLikely,
		Matchers: []Matcher{
			{Type: "keyword", Pattern: "pod status pending", Weight: 1.0},
			{Type: "keyword", Pattern: "0/1 nodes available", Weight: 1.0},
			{Type: "keyword", Pattern: "insufficient cpu", Weight: 0.9},
			{Type: "keyword", Pattern: "insufficient memory", Weight: 0.9},
			{Type: "keyword", Pattern: "failed to pull image", Weight: 0.7},
		},
		Correlations: []Correlation{
			{PatternID: "pvc-binding-failure", Message: "PVC binding failure prevents pod scheduling"},
			{PatternID: "cni-error", Message: "CNI errors prevent pod sandbox creation"},
			{PatternID: "imagepullbackoff-v2", Message: "Image pull failure prevents pod start"},
		},
		Description: "Pod cannot be scheduled or initialized, stuck in Pending state",
		HintGenerator: HintGenerator{
			Template:   "Pod stuck Pending.",
			Suggestion: "Check: 1) Resource limits fit node capacity, 2) PVCs are bound, 3) Image can be pulled, 4) CNI is working",
			Command:    "kubectl describe pod {{.PodName}} -n {{.Namespace}} | grep -A20 'Events' && kubectl get nodes -o yaml | grep -A5 'allocatable'",
			References: []string{"https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/", "https://kubernetes.io/docs/tasks/debug/debug-application/debug-pods/"},
		},
	},
	{
		ID:         "pod-stuck-terminating",
		Name:       "Pod Stuck Terminating",
		Category:   "Pod",
		Severity:   SeverityWarning,
		Confidence: ConfidenceLikely,
		Matchers: []Matcher{
			{
				Type:    "regex",
				Pattern: `(?P<Namespace>[a-z0-9-]+)\s+(?P<PodName>[a-z0-9-]+)\s+\S+\s+Terminating\s+\S+\s+(?P<Duration>[0-9dhmsy]+)`,
				Weight:  1.0,
			},
			{Type: "keyword", Pattern: "pod status terminating", Weight: 1.0},
			{Type: "keyword", Pattern: "terminating", Weight: 0.8},
			{Type: "keyword", Pattern: "deletiontimestamp", Weight: 0.9},
			{Type: "keyword", Pattern: "failed to delete", Weight: 0.8},
		},
		Description: "Pod is stuck in Terminating state, cannot be deleted",
		HintGenerator: HintGenerator{
			Template:   "Pod {{.PodName}} stuck Terminating for {{.Duration}}.",
			Suggestion: "Common causes: 1) Stuck finalizers, 2) Node unreachable, 3) Volume unmount stuck. Force delete with --grace-period=0 --force as last resort.",
			Command:    "kubectl describe pod {{.PodName}} -n {{.Namespace}} | grep -E 'Finalizers|Node' && kubectl delete pod {{.PodName}} -n {{.Namespace}} --grace-period=0 --force",
			References: []string{"https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/", "https://kubernetes.io/docs/tasks/run-application/force-delete-stateful-set-pod/"},
		},
	},
	{
		ID:         "leader-election-failure",
		Name:       "Leader Election Failure",
		Category:   "ControlPlane",
		Severity:   SeverityCritical,
		Confidence: ConfidenceCertain,
		Matchers: []Matcher{
			{Type: "keyword", Pattern: "failed to acquire lease", Weight: 1.0},
			{Type: "keyword", Pattern: "leader election lost", Weight: 1.0},
			{Type: "keyword", Pattern: "leader election failed", Weight: 1.0},
			{Type: "keyword", Pattern: "cancelled leader election", Weight: 1.0},
			{Type: "keyword", Pattern: "failed to renew lease", Weight: 0.9},
		},
		Correlations: []Correlation{
			{PatternID: "etcd-quorum", Message: "Leader election requires healthy etcd"},
			{PatternID: "connectivity-timeout", Message: "Network issues may prevent leader election"},
		},
		Description: "Control plane component failing leader election (controller-manager, scheduler)",
		HintGenerator: HintGenerator{
			Template:   "Leader election failed.",
			Suggestion: "Check: 1) etcd health and quorum, 2) Network connectivity to API server, 3) Lease objects exist in kube-system. May indicate split-brain or network partition.",
			Command:    "kubectl get leases -n kube-system && ETCDCTL_API=3 etcdctl endpoint health --cluster && kubectl get endpoints kube-controller-manager -n kube-system",
			References: []string{"https://kubernetes.io/docs/concepts/architecture/leases/", "https://docs.rke2.io/architecture/architecture/"},
		},
	},
	{
		ID:         "node-not-ready",
		Name:       "Node Not Ready",
		Category:   "Node",
		Severity:   SeverityCritical,
		Confidence: ConfidenceCertain,
		Matchers: []Matcher{
			{Type: "keyword", Pattern: "node notready", Weight: 1.0},
			{Type: "keyword", Pattern: "status notready", Weight: 1.0},
			{Type: "keyword", Pattern: "kubelet stopped posting node status", Weight: 1.0},
			{Type: "keyword", Pattern: "node controller marked node", Weight: 0.9},
		},
		Correlations: []Correlation{
			{PatternID: "certificate-expired", Message: "Node NotReady often caused by expired certificates"},
			{PatternID: "node-pressure", Message: "Node pressure conditions may cause NotReady"},
		},
		Description: "Node is in NotReady state, not accepting pods",
		HintGenerator: HintGenerator{
			Template:   "Node is NotReady.",
			Suggestion: "Check: 1) Kubelet status on node, 2) Certificate expiration, 3) Network connectivity to API server, 4) Disk/memory pressure on node",
			Command:    "kubectl describe node {{.NodeName}} && systemctl status kubelet && journalctl -u kubelet -n 50",
			References: []string{"https://kubernetes.io/docs/concepts/architecture/nodes/", "https://kubernetes.io/docs/tasks/debug/debug-cluster/"},
		},
	},
}

// PatternRegistryV2 manages PatternV2 definitions with correlation support
// Sprint 11: Enhanced registry for AI Intelligence
type PatternRegistryV2 struct {
	patterns []PatternV2
}

// NewRegistryV2 creates a new v2 pattern registry with built-in patterns
func NewRegistryV2() *PatternRegistryV2 {
	patterns := make([]PatternV2, len(BuiltinPatternsV2))
	copy(patterns, BuiltinPatternsV2)
	return &PatternRegistryV2{
		patterns: patterns,
	}
}

// Register adds a new pattern to the v2 registry
func (r *PatternRegistryV2) Register(p PatternV2) error {
	if p.ID == "" {
		return fmt.Errorf("pattern ID is required")
	}
	if p.Name == "" {
		return fmt.Errorf("pattern name is required")
	}
	if len(p.Matchers) == 0 {
		return fmt.Errorf("at least one matcher is required")
	}

	r.patterns = append(r.patterns, p)
	return nil
}

// GetByID retrieves a v2 pattern by ID
func (r *PatternRegistryV2) GetByID(id string) (PatternV2, bool) {
	for _, p := range r.patterns {
		if p.ID == id {
			return p, true
		}
	}
	return PatternV2{}, false
}

// GetByCategory retrieves all v2 patterns in a category
func (r *PatternRegistryV2) GetByCategory(category string) []PatternV2 {
	var result []PatternV2
	for _, p := range r.patterns {
		if p.Category == category {
			result = append(result, p)
		}
	}
	return result
}

// GetAll returns all v2 patterns
func (r *PatternRegistryV2) GetAll() []PatternV2 {
	return r.patterns
}

// AnalyzeV2 scans content against all v2 patterns and returns matches with correlations
func (r *PatternRegistryV2) AnalyzeV2(content string) []MatchResultV2 {
	var matches []MatchResultV2
	matchedIDs := make(map[string]bool)

	// First pass: collect all matches
	for _, pattern := range r.patterns {
		matcher := NewMatcherV2(pattern)
		results := matcher.Match(content)
		for _, result := range results {
			if result.Matched {
				matches = append(matches, result)
				matchedIDs[pattern.ID] = true
			}
		}
	}

	// Second pass: add correlations
	for i := range matches {
		pattern, found := r.GetByID(matches[i].PatternID)
		if found {
			for _, corr := range pattern.Correlations {
				if matchedIDs[corr.PatternID] {
					matches[i].Correlated = append(matches[i].Correlated, corr.PatternID)
				}
			}
		}
	}

	return matches
}
