# Log Bundle Analysis - r8s Project

**Date:** November 27, 2025  
**Bundle Analyzed:** `example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09.tar.gz`  
**Bundle Type:** Rancher/RKE2 Cluster Support Bundle  
**Total Files:** 337

---

## 📋 Executive Summary

The log bundle is a comprehensive **Rancher/RKE2 cluster support bundle** containing diagnostic data from a single cluster node. It includes system information, network configuration, Kubernetes resources, pod logs, etcd data, and certificates - everything needed for deep troubleshooting.

**Key Insight:** This bundle format represents a critical offline troubleshooting capability that r8s should natively support, allowing engineers to analyze cluster issues without live access.

---

## 📁 Bundle Structure Analysis

### Directory Breakdown

```
Total Files: 337
├── rke2/               246 files (73%) ⭐ LARGEST SECTION
├── systeminfo/          30 files (9%)
├── networking/          28 files (8%)
├── systemlogs/           4 files (1%)
├── etcd/                 8 files (2%)
├── journald/             3 files (1%)
└── collector-output.log  1 file  (<1%)
```

---

## 🔍 Detailed Section Analysis

### 1. etcd/ Directory (8 files) - Cluster State Database

**Purpose:** etcd cluster health and configuration data

**Files:**
- `memberlist` - List of etcd cluster members and their roles
- `endpointstatus` - Endpoint status information
- `endpointhealth` - Health check results for etcd endpoints
- `alarmlist` - Active etcd alarms (critical for troubleshooting)
- `findserverdbetcd` - etcd database location and info
- `etcd-metrics-127.0.0.1:2381.txt` - Prometheus metrics from etcd
- `findserverdbsnapshots` - etcd backup/snapshot information

**Troubleshooting Value:** 🔴 CRITICAL
- etcd health directly impacts cluster stability
- Common issues: split-brain, disk pressure, member down
- Metrics reveal performance problems

**UX Considerations:**
- Parse health status into Red/Yellow/Green indicators
- Extract and highlight alarms prominently
- Show member quorum status
- Graph key metrics (latency, DB size)

---

### 2. journald/ Directory (3 files) - Systemd Service Logs

**Purpose:** System service journal logs

**Files:**
- `rke2-server` - RKE2 server service logs (main K8s control plane)
- `cloud-init` - Cloud initialization logs
- `rancher-system-agent` - Rancher agent logs

**Troubleshooting Value:** 🟡 HIGH
- Service start/stop/crash information
- Initialization problems
- Agent communication issues

**UX Considerations:**
- Timestamp-based navigation
- Error/warning highlighting
- Search across all journal logs
- Show service restart history

---

### 3. networking/ Directory (28 files) - Network Configuration

**Purpose:** Complete network stack configuration

**Files:**

**iptables Rules (7 files):**
- `iptables`, `iptablessave`, `iptablesnat`, `iptablesmangle`
- `ip6tables`, `ip6tablessave`, `ip6tablesnat`, `ip6tablesmangle`

**IP Configuration (9 files):**
- `iproute`, `ipv6route` - Routing tables
- `ipaddrshow`, `ipv6addrshow` - IP addresses
- `iplinkshow` - Network interfaces
- `ipneighbour`, `ipv6neighbour` - ARP/ND tables
- `iprule`, `ipv6rule` - Policy routing

**Socket Statistics (8 files):**
- `ss4apn`, `ss6apn` - TCP/UDP sockets (IPv4/IPv6, all, processes, numeric)
- `ssxapn`, `ssuapn`, `ssitan`, `sswapn` - Various socket states
- `sstunlp4`, `sstunlp6` - Listening sockets

**Other (4 files):**
- `cni/10-calico.conflist` - Calico CNI configuration
- `nft_ruleset` - nftables rules
- `ethtool` - Network interface details
- `procnetxfrmstat` - IPsec statistics

**Troubleshooting Value:** 🟡 HIGH
- Network connectivity issues
- Firewall rule problems
- CNI misconfiguration
- Port conflicts

**UX Considerations:**
- Parse iptables rules into readable format
- Show active connections grouped by port/service
- Highlight CNI configuration issues
- Searchable rules database

---

### 4. rke2/ Directory (246 files - 73%) - RKE2/Kubernetes Data ⭐

**Purpose:** Complete Kubernetes cluster state and configuration

#### 4.1 Configuration Files (8 files)
- `50-rancher.yaml` - Rancher configuration
- `version` - RKE2 version information
- `rke2-server.service` - Systemd service definition
- `rke2-agent.service` - Agent service definition
- `directories/rke2agent` - Agent directory listing
- `directories/rke2servermanifests` - Server manifest listing
- `directories/rke2servertls` - TLS directory listing

#### 4.2 Logs (2 files)
- `agent-logs/kubelet.log` - Kubelet logs (pod lifecycle)
- `containerd.log` - Container runtime logs

#### 4.3 Certificates (12 files in certs/agent/ and certs/server/)
**Agent Certificates:**
- `client-rke2-controller.crt`
- `client-kube-proxy.crt`
- `serving-kubelet.crt`
- `client-kubelet.crt`

**Server Certificates:**
- `client-scheduler.crt`
- `client-ca.nochain.crt`
- `client-kube-apiserver.crt`
- `server-ca.nochain.crt`
- `client-controller.crt`
- `client-supervisor.crt`
- `serving-kube-apiserver.crt`
- `client-rke2-cloud-controller.crt`
- `client-auth-proxy.crt`
- `client-admin.crt`

#### 4.4 Static Pod Manifests (6 files in pod-manifests/)
- `etcd.yaml`
- `kube-apiserver.yaml`
- `kube-controller-manager.yaml`
- `kube-scheduler.yaml`
- `kube-proxy.yaml`
- `cloud-controller-manager.yaml`

#### 4.5 Container Runtime Info (10 files in crictl/)
- `pods` - Running pods
- `statsa` - Container statistics
- `psa` - Pod sandboxes
- `images` - Container images
- `version` - Runtime version
- `crictl-version` - crictl version
- `containerd-version` - containerd version
- `imagefsinfo` - Image filesystem info
- `info` - Runtime information
- `runc-version` - runc version

#### 4.6 Kubernetes Resources (30 files in kubectl/) 🎯
**Core Workloads:**
- `pods` - All pods across all namespaces
- `deployments` - Deployment resources
- `replicasets` - ReplicaSets
- `statefulsets` - StatefulSets
- `daemonsets` - DaemonSets
- `jobs` - Batch jobs
- `cronjobs` - Scheduled jobs

**Networking:**
- `services` - Service resources
- `endpoints` - Service endpoints
- `ingress` - Ingress resources
- `networkpolicies` - Network policies

**Configuration:**
- `configmaps` - ConfigMaps
- `secrets` - Secrets (metadata only)
- `namespaces` - Namespace list

**RBAC:**
- `roles` - Roles
- `rolebindings` - RoleBindings
- `clusterroles` - ClusterRoles
- `clusterrolebindings` - ClusterRoleBindings

**Storage:**
- `pv` - PersistentVolumes
- `pvc` - PersistentVolumeClaims
- `volumeattachments` - Volume attachments

**Cluster:**
- `nodes` - Node information
- `nodesdescribe` - Detailed node descriptions
- `events` - Cluster events (critical for debugging!)
- `crds` - CustomResourceDefinitions
- `apiservices` - API services
- `api-resources` - Available API resources
- `helmcharts` - Helm charts
- `hpa` - HorizontalPodAutoscalers
- `leases` - Lease objects
- `validatingwebhookconfigurations` - Validating webhooks
- `mutatingwebhookconfigurations` - Mutating webhooks
- `version` - Kubernetes version

#### 4.7 Pod Logs (160+ files in podlogs/) 📝
**Structure:** Logs for each pod, with both current and previous (-previous suffix) logs

**Sample Pods:**
- **kube-system:** etcd, kube-apiserver, kube-controller-manager, kube-scheduler, kube-proxy, coredns, ingress-nginx
- **calico-system:** calico-node, calico-typha, calico-kube-controllers
- **cattle-system:** cattle-cluster-agent, rancher-webhook, system-upgrade-controller
- **cattle-monitoring-system:** prometheus, alertmanager, grafana, node-exporter, kube-state-metrics, pushprox components
- **cattle-fleet-system:** fleet-agent
- **longhorn-system:** longhorn-manager, longhorn-ui, longhorn-driver-deployer, CSI components, instance-managers
- **tigera-operator:** tigera-operator

**Log Count:**
- Current logs: ~80 files
- Previous logs: ~80 files (for crash investigation)

**Troubleshooting Value:** 🔴 CRITICAL
- Most valuable data for troubleshooting
- Crash loops visible via -previous logs
- Error messages and stack traces
- Container startup issues

**UX Considerations:**
- **Group by namespace** for organization
- **Current vs Previous toggle** for crash analysis
- **Full-text search** across all logs
- **Timestamp navigation**
- **Error/Warning highlighting**
- **Link to related resources** (pod → deployment)

---

### 5. systeminfo/ Directory (30 files) - System Information

**Purpose:** Complete system state snapshot

**Hardware & Resources (5 files):**
- `cpuinfo` - CPU information
- `lsblk` - Block devices
- `freem` - Memory usage (free -m)
- `dfh` - Disk usage (df -h)
- `dfi` - Disk inodes (df -i)

**Operating System (4 files):**
- `osrelease` - OS release information
- `uname` - Kernel information
- `hostname` - Hostname
- `hostnamefqdn` - Fully qualified domain name

**Processes (3 files):**
- `ps` - Process list
- `lsof` - Open files
- `top` - Process snapshot

**Disk & Filesystem (2 files):**
- `mount` - Mounted filesystems
- `lsblk` - (duplicate above)

**Security (3 files):**
- `ubuntu-ufw` - UFW firewall status
- `ubuntu-apparmorstatus` - AppArmor status
- `lsmod` - Loaded kernel modules

**Network (2 files):**
- `etchosts` - /etc/hosts file
- `etcresolvconf` - /etc/resolv.conf

**Systemd (3 files):**
- `systemd-units` - Active systemd units
- `systemd-unit-files` - All unit files
- `service-statusall` - All service statuses
- `systemd-resolved` - systemd-resolved status

**Kernel (4 files):**
- `dmesg` - Kernel ring buffer
- `sysctla` - Kernel parameters (sysctl -a)
- `vmstat` - Virtual memory statistics

**System State (4 files):**
- `date` - Current date/time
- `uptime` - System uptime
- `file-max` - File descriptor limits
- `file-nr` - Current file descriptors
- `ulimit-hard` - Hard ulimits

**Package Management (1 file):**
- `packages-dpkg` - Installed packages (dpkg -l)

**Troubleshooting Value:** 🟡 HIGH
- Resource exhaustion issues
- Kernel problems
- Security configuration
- Service status

**UX Considerations:**
- Parse into structured data
- Highlight resource warnings (high CPU, low memory, disk full)
- Show security posture summary
- Link processes to pods/containers

---

### 6. systemlogs/ Directory (4 files) - OS System Logs

**Purpose:** Operating system log files

**Files:**
- `syslog` - Current syslog
- `syslog.1` - Previous syslog
- `kern.log` - Current kernel log
- `kern.log.1` - Previous kernel log

**Troubleshooting Value:** 🟡 MEDIUM
- OS-level errors
- Kernel panics
- Hardware issues
- Boot problems

**UX Considerations:**
- Timestamp-based navigation
- Error/warning filtering
- Search functionality
- Link to related systeminfo

---

### 7. collector-output.log (1 file) - Collection Metadata

**Purpose:** Log of the collection process itself

**Troubleshooting Value:** 🟢 LOW
- Verify bundle completeness
- Collection errors

---

## 🎨 UX Architecture Design

### A. Dual-Mode Navigation

```
r8s
├── 🌐 Live Mode (existing)
│   └── API → Clusters → Projects → Namespaces → Resources
│
└── 📦 Bundle Mode (NEW)
    ├── Bundle Selection
    │   ├── Recent bundles
    │   ├── Browse filesystem
    │   └── Drag & drop support
    │
    └── Bundle Browser
        ├── Overview Dashboard
        ├── System Analysis
        ├── Network Analysis
        ├── Kubernetes Resources
        ├── Logs Viewer
        ├── etcd Status
        └── Security Review
```

### B. Overview Dashboard (Landing Page)

```
┌─ Bundle Overview ─────────────────────────────────────┐
│ Bundle: w-guard-wg-cp-svtk6-lqtxw                     │
│ Collected: 2025-11-27 04:19:09                        │
│ Node: w-guard-wg-cp-svtk6-lqtxw                      │
│ RKE2 Version: v1.28.x                                 │
│                                                        │
│ ┌─ Health Summary ────────────────────────────────┐  │
│ │ ● etcd:        HEALTHY    (3/3 members)         │  │
│ │ ● API Server:  HEALTHY    (responding)          │  │
│ │ ● Pods:        WARNING    (5 CrashLoopBackOff)  │  │
│ │ ● Disk:        CRITICAL   (92% full)            │  │
│ │ ● Memory:      HEALTHY    (45% used)            │  │
│ └─────────────────────────────────────────────────┘  │
│                                                        │
│ ┌─ Quick Stats ───────────────────────────────────┐  │
│ │ Namespaces:     12          Pods:        147    │  │
│ │ Deployments:    28          Nodes:       3      │  │
│ │ Services:       45          PVs:         8      │  │
│ │ Events (24h):   342         Warnings:    23     │  │
│ └─────────────────────────────────────────────────┘  │
│                                                        │
│ [1] System  [2] Network  [3] K8s  [4] Logs  [5] etcd │
└────────────────────────────────────────────────────────┘
```

### C. Log Viewer Interface

```
┌─ Pod Logs ─────────────────────────────────────────────┐
│ Namespace: cattle-monitoring-system                    │
│ Pod: prometheus-rancher-monitoring-prometheus-0        │
│ Container: prometheus         [Current] [Previous]     │
│                                                         │
│ 🔍 Search: error ▂▂▂▂▂▂▂▂  [3 matches]  [n] Next      │
│                                                         │
│ ┌─────────────────────────────────────────────────────┤
│ │ 2025-11-27T04:15:23.123Z level=info msg="Starting" │
│ │ 2025-11-27T04:15:24.456Z level=info msg="Loaded"   │
│ │ 2025-11-27T04:16:01.789Z level=error msg="Failed   │ ←
│ │   to scrape target" err="context deadline exceeded"│
│ │ 2025-11-27T04:16:02.012Z level=warn msg="Retry"    │
│ │ 2025-11-27T04:17:15.345Z level=info msg="Success"  │
│ │ ...                                                  │
│ └─────────────────────────────────────────────────────┘
│                                                         │
│ [/] Search  [g] Go to time  [f] Filter  [e] Export    │
└────────────────────────────────────────────────────────┘
```

### D. Kubernetes Resource Browser

```
┌─ Deployments ──────────────────────────────────────────┐
│ Namespace: All                       [Filter: ▂▂▂▂▂]   │
│                                                         │
│ NAME                          READY    STATUS   AGE    │
│ ────────────────────────────────────────────────────── │
│ cattle-cluster-agent          1/1      Running  30d    │
│ longhorn-ui                   1/1      Running  25d    │
│ prometheus                    0/1 ⚠️    Pending  2h     │ ←
│ grafana                       1/1      Running  20d    │
│ alertmanager                  1/1      Running  20d    │
│                                                         │
│ [Enter] Describe  [d] Details  [l] Logs  [e] Events   │
└────────────────────────────────────────────────────────┘
```

---

## 🏗️ Implementation Architecture

### Package Structure

```
internal/
├── logbundle/
│   ├── bundle.go           # Bundle type and loader
│   ├── extractor.go        # tar.gz extraction
│   ├── parser.go           # Generic file parser
│   ├── indexer.go          # Search index builder
│   │
│   ├── parsers/
│   │   ├── kubectl.go      # Parse kubectl outputs
│   │   ├── etcd.go         # Parse etcd data
│   │   ├── systeminfo.go   # Parse system files
│   │   ├── networking.go   # Parse network configs
│   │   ├── logs.go         # Log file parser
│   │   └── certs.go        # Certificate parser
│   │
│   └── analyzer/
│       ├── health.go       # Health scoring
│       ├── problems.go     # Problem detection
│       └── recommendations.go  # Suggestions
│
└── tui/
    └── views/
        ├── bundle_overview.go    # Dashboard view
        ├── bundle_logs.go        # Log viewer
        ├── bundle_resources.go   # K8s resources
        ├── bundle_system.go      # System info
        └── bundle_network.go     # Network info
```

### Core Types

```go
// Bundle represents a loaded support bundle
type Bundle struct {
    Path          string
    ExtractedPath string
    Manifest      *BundleManifest
    
    // Parsed data
    Resources     *K8sResources
    SystemInfo    *SystemInfo
    NetworkInfo   *NetworkInfo
    EtcdInfo      *EtcdInfo
    Logs          *LogIndex
    Certificates  []*Certificate
    
    // Analysis
    Health        *HealthStatus
    Problems      []*Problem
}

// BundleManifest describes bundle contents
type BundleManifest struct {
    NodeName      string
    CollectedAt   time.Time
    RKE2Version   string
    K8sVersion    string
    FileCount     int
    TotalSize     int64
}

// K8sResources contains parsed Kubernetes resources
type K8sResources struct {
    Pods          []Pod
    Deployments   []Deployment
    Services      []Service
    Nodes         []Node
    Events        []Event
    // ... other resource types
}

// LogIndex provides fast log searching
type LogIndex struct {
    Logs          []*LogFile
    Index         *bleve.Index  // Full-text search
}

// LogFile represents a single log file
type LogFile struct {
    Path          string
    Type          LogType  // Pod, System, Journal
    Namespace     string
    PodName       string
    Container     string
    IsPrevious    bool
    Size          int64
    LineCount     int
}
```

---

## 🎯 Feature Priorities

### Phase 1: Foundation (Week 1) - MUST HAVE
**Goal:** Basic bundle loading and browsing

1. **Bundle Loader**
   - [ ] tar.gz extraction to temp directory
   - [ ] Bundle manifest detection
   - [ ] File inventory creation
   - [ ] Basic validation

2. **Simple File Browser**
   - [ ] Tree view of bundle contents
   - [ ] File preview (text files)
   - [ ] Basic searching (grep-style)
   - [ ] Navigation (up/down directory tree)

3. **Integration**
   - [ ] Add `--bundle=/path/to.tar.gz` flag
   - [ ] Switch between Live and Bundle mode
   - [ ] Status bar indicator for mode

**Success Criteria:**
- Can load and browse bundle
- Can view file contents
- Can search within files
- <5s load time for example bundle

---

### Phase 2: Kubernetes Resources (Week 2) - HIGH VALUE
**Goal:** Parse and display K8s resources

1. **Resource Parsers**
   - [ ] Parse kubectl/ output files (JSON/YAML)
   - [ ] Build resource index
   - [ ] Extract relationships (pod → deployment)

2. **Resource Views**
   - [ ] Pods table (reuse existing table component)
   - [ ] Deployments table
   - [ ] Services table
   - [ ] Events viewer (critical!)
   - [ ] Nodes viewer

3. **Navigation**
   - [ ] Jump between related resources
   - [ ] Link to pod logs
   - [ ] Describe resource (from kubectl output)

**Success Criteria:**
- All major resource types displayed
- Can navigate resource relationships
- Events linked to resources
- Sub-second navigation between views

---

### Phase 3: Log Viewer (Week 3) - HIGH VALUE
**Goal:** Advanced log viewing and searching

1. **Log Indexing**
   - [ ] Parse all pod logs
   - [ ] Build full-text search index
   - [ ] Extract metadata (namespace, pod, container)
   - [ ] Timestamp parsing

2. **Log Viewer UI**
   - [ ] Display logs with syntax highlighting
   - [ ] Multi-log search (across all logs)
   - [ ] Current vs Previous toggle
   - [ ] Timestamp navigation
   - [ ] Error/Warning highlighting

3. **Smart Features**
   - [ ] Auto-link to related resources
   - [ ] Crash pattern detection
   - [ ] Log correlation (same timestamp across pods)

**Success Criteria:**
- Can search across 160+ log files in <1s
- Can view current and previous logs
- Error highlighting works
- Timestamp navigation accurate

---

### Phase 4: Health Analysis (Week 4) - HIGH VALUE
**Goal:** Automated problem detection

1. **Health Checks**
   - [ ] etcd health analysis
   - [ ] Pod status analysis (CrashLoopBackOff detection)
   - [ ] Resource usage warnings (disk, memory)
   - [ ] Certificate expiry checking
   - [ ] Event analysis (warning/error counts)

2. **Problem Detection**
   - [ ] Common failure patterns
   - [ ] Configuration issues
   - [ ] Resource exhaustion
   - [ ] Network problems

3. **Recommendations**
   - [ ] Suggested fixes
   - [ ] Related documentation links
   - [ ] Severity scoring

**Success Criteria:**
- Dashboard shows accurate health status
- Top 5 problems identified automatically
- Actionable recommendations provided

---

### Phase 5: System & Network (Week 5) - MEDIUM VALUE
**Goal:** System and network analysis

1. **System Info Display**
   - [ ] Parse systeminfo files
   - [ ] Display CPU/memory/disk
   - [ ] Show running processes
   - [ ] Security status
   - [ ] Service status

2. **Network Display**
   - [ ] Parse iptables rules
   - [ ] Show active connections
   - [ ] CNI configuration
   - [ ] Routing tables

**Success Criteria:**
- All systeminfo displayed clearly
- Network rules searchable
- Resource warnings highlighted

---

### Phase 6: Advanced Features (Week 6+) - NICE TO HAVE
**Goal:** Power-user features

1. **Export & Reports**
   - [ ] Generate HTML report
   - [ ] Export filtered logs
   - [ ] Save search results
   - [ ] Markdown summary export

2. **Comparison**
   - [ ] Compare two bundles
   - [ ] Diff resource states
   - [ ] Track changes over time

3. **Multi-bundle Support**
   - [ ] Load multiple node bundles
   - [ ] Correlate across nodes
   - [ ] Cluster-wide view

---

## 🔧 Technical Decisions

### 1. Bundle Extraction Strategy

**Options:**
- **A. Extract to temp directory** (RECOMMENDED)
  - Pros: Fast subsequent access, easier to work with
  - Cons: Disk space usage, cleanup required
  - Implementation: `os.MkdirTemp()`, defer cleanup

- **B. Stream from tar.gz**
  - Pros: Lower disk usage
  - Cons: Slower, must re-extract for each access
  - Implementation: `archive/tar` package

**Decision:** Option A - Extract to temp directory
- Bundle is only ~100MB compressed, ~500MB uncompressed
- Much faster for repeated file access
- Simpler code

### 2. Search Implementation

**Options:**
- **A. Full-text index (bleve)** (RECOMMENDED for logs)
  - Pros: Instant search, advanced queries
  - Cons: Index creation time, memory usage
  - Implementation: Build index on bundle load

- **B. Grep-style on-demand**
  - Pros: No index overhead
  - Cons: Slow for large logs, no advanced queries
  - Implementation: Parallel grep across files

**Decision:** Hybrid approach
- Full-text index for pod logs (160+ files)
- Grep-style for other files (one-off searches)

### 3. Resource Parsing

**Format:** kubectl outputs are in custom table format, not JSON/YAML

**Strategy:**
- Parse table format (column-based)
- Extract structured data
- Build internal types matching rancher package types
- Reuse existing table rendering components

### 4. Memory Management

**Challenge:** Large bundles, many logs

**Strategy:**
- Lazy load log contents (only when viewed)
- Keep metadata in memory
- Stream large files instead of loading entirely
- Use memory-mapped files for search index

---

## 📊 Success Metrics

### Performance Targets
- Bundle load time: <10s
- Search across logs: <1s
- Navigation between views: <100ms
- Memory usage: <200MB for typical bundle

### UX Targets
- Health dashboard shows critical issues immediately
- Can find relevant logs in <3 clicks
- Search returns useful results (precision >80%)
- Problem detection catches top 10 common issues

### Code Quality
- Test coverage: 70%+
- Documentation: All public APIs
- Error handling: Graceful degradation
- No crashes on malformed bundles

---

## 🤔 Open Questions

### 1. Bundle Format Standardization
- Should we support other bundle formats? (kubectl cluster-info dump, etc.)
- Should we define a standard r8s bundle format?

### 2. Multi-Node Correlation
- How to handle bundles from different nodes in same cluster?
- Should we auto-detect and correlate?

### 3. Real-time vs Bundle
- Should Live Mode be able to "snapshot" to a bundle?
- Should Bundle Mode allow "connecting" if credentials available?

### 4. Storage
- Save loaded bundles for quick re-access?
- Cache parsed data between sessions?

---

## 📝 Implementation Notes

### File Format Detection
```go
// Detect bundle format by structure
func DetectBundleFormat(path string) (BundleFormat, error) {
    // Check for rke2/ directory (RKE2 support bundle)
    // Check for kubectl/ directory (kubectl cluster-info dump)
    // Check for rancher-logs/ (Rancher support bundle)
}
```

### Health Scoring Algorithm
```go
// Calculate overall health score
func CalculateHealth(bundle *Bundle) *HealthScore {
    score := 100
    
    // etcd checks
    if !bundle.EtcdInfo.AllMembersHealthy() {
        score -= 30  // Critical
    }
    
    // Pod checks
    crashLoopCount := bundle.Resources.CountCrashLoopBackOff()
    score -= min(crashLoopCount * 5, 20)
    
    // Resource checks
    if bundle.SystemInfo.DiskUsagePercent > 90 {
        score -= 20  // Critical
    }
    
    // ... more checks
    
    return &HealthScore{
        Score: score,
        Grade: getGrade(score),
        Issues: collectIssues(),
    }
}
```

---

## 🚀 Next Steps

1. **Review and approve** this analysis
2. **Toggle to Act mode** to begin implementation
3. **Start with Phase 1** - Basic bundle loading
4. **Iterate based on feedback** - Build incrementally
5. **Target delivery** - 6 weeks for all phases

---

**Status:** 📋 Analysis Complete - Ready for Implementation  
**Estimated Effort:** 6 weeks (Phases 1-6)  
**Risk Level:** 🟡 Medium (new feature area, but well-scoped)  
**Value Proposition:** 🔴 High (enables offline troubleshooting, unique feature)
