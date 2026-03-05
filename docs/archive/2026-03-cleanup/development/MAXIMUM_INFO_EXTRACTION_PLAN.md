# Maximum Information Extraction Implementation Plan

**Created**: 2026-01-15  
**Principle**: "Extract and display all available bundle data, even when incomplete. Be transparent about data gaps."  
**Current State**: r8s parses ~30% of bundle data (9 of 90+ files)  
**Target**: 90%+ data extraction by v0.8.0

---

## Implementation Phases

### Phase 1: Critical Data (v0.7.0) - 3-4 weeks

**Priority**: HIGH - Most impactful troubleshooting data

#### Storage & Workloads (Week 1, ~15 hours)

**Task 1: StatefulSet Parser**
```go
// File: internal/bundle/kubectl.go
func ParseStatefulSets(extractPath string) ([]StatefulSetInfo, error) {
    // Parse kubectl get statefulsets output
    // Format: NAMESPACE NAME READY AGE
    // Extract replicas, ready status
}

// Add to loader.go
statefulSets, err := ParseStatefulSets(extractPath)
```

**Task 2: PV/PVC Parser**
```go
func ParsePersistentVolumes(extractPath string) ([]PVInfo, error) {
    // Parse kubectl get pv output
    // Extract: name, capacity, access modes, status, claim
}

func ParsePersistentVolumeClaims(extractPath string) ([]PVCInfo, error) {
    // Parse kubectl get pvc output
    // Extract: namespace, name, status, volume, capacity
}
```

**Task 3: TUI Views for Storage**
- Add "Storage" view to main menu
- Show PV/PVC status with health indicators
- Link PVCs to pods using them
- Show "Pending" PVCs prominently in dashboard

**Estimated Time**: 15 hours  
**Impact**: CRITICAL - Storage issues are top support topics

#### System Health Data (Week 2, ~12 hours)

**Task 4: dmesg Parser**
```go
// File: internal/bundle/systeminfo.go (NEW)
func ParseDmesg(extractPath string) ([]DmesgEntry, error) {
    // Parse systeminfo/dmesg
    // Extract OOM kills, hardware errors, kernel panics
    // Return structured entries with timestamp, level, message
}
```

**Task 5: Disk Space Parser**
```go
func ParseDiskSpace(extractPath string) (*DiskSpaceInfo, error) {
    // Parse systeminfo/dfh
    // Calculate % used per mount point
    // Flag critical (>90%), warning (>80%)
}
```

**Task 6: Dashboard Integration**
- Show disk space warnings in attention dashboard
- "⚠️  Disk 95% full on /var/lib/rancher/rke2"
- Show OOM events from dmesg
- "🔥 3 OOM kills detected in last 24h"

**Estimated Time**: 12 hours  
**Impact**: HIGH - Proactive issue detection

#### Journald Logs (Week 3, ~10 hours)

**Task 7: RKE2 Server Log Parser**
```go
// File: internal/bundle/journald.go (NEW)
func ParseJournaldRKE2Server(extractPath string) ([]LogEntry, error) {
    // Parse journald/rke2-server
    // Extract errors, warnings
    // Return structured log entries
}
```

**Task 8: Log Integration**
- Add "Control Plane" to log source selector
- Enable log filtering/search on journald logs
- Show RKE2 startup errors in dashboard

**Estimated Time**: 10 hours  
**Impact**: CRITICAL - Control plane troubleshooting

#### Configuration Data (Week 4, ~8 hours)

**Task 9: ConfigMap & HelmChart Parser**
```go
func ParseConfigMaps(extractPath string) ([]ConfigMapInfo, error) {
    // Parse kubectl get configmaps
}

func ParseHelmCharts(extractPath string) ([]HelmChartInfo, error) {
    // Parse kubectl get helmcharts (Rancher-specific)
    // Extract chart name, version, status
}
```

**Task 10: Configuration View**
- Add "Configuration" view
- Show ConfigMaps by namespace
- Show Rancher apps (HelmCharts) with status
- Link to pods using ConfigMaps

**Estimated Time**: 8 hours  
**Impact**: MEDIUM-HIGH - Configuration debugging

---

### Phase 2: Network & Additional Resources (v0.7.5) - 2 weeks

**Priority**: MEDIUM-HIGH

#### Network Resources (Week 1, ~12 hours)

**Task 11: Ingress & Endpoints Parser**
```go
func ParseIngress(extractPath string) ([]IngressInfo, error) {
    // Parse kubectl get ingress
    // Extract hosts, backends, TLS
}

func ParseEndpoints(extractPath string) ([]EndpointsInfo, error) {
    // Parse kubectl get endpoints
    // Map to services, show ready addresses
}
```

**Task 12: Networking View**
- Add "Networking" view
- Show ingress rules
- Show service → endpoints mapping
- Highlight services with no ready endpoints

**Estimated Time**: 12 hours

#### Additional Workloads (Week 2, ~10 hours)

**Task 13: Jobs, CronJobs, ReplicaSets, HPA**
```go
func ParseJobs(extractPath string) ([]JobInfo, error)
func ParseCronJobs(extractPath string) ([]CronJobInfo, error)
func ParseReplicaSets(extractPath string) ([]ReplicaSetInfo, error)
func ParseHPA(extractPath string) ([]HPAInfo, error)
```

**Task 14: Workload View Expansion**
- Show all workload types in workloads view
- Filter by type (pod, deployment, statefulset, job, etc.)

**Estimated Time**: 10 hours

---

### Phase 3: Advanced Diagnostics (v0.8.0) - 2 weeks

**Priority**: MEDIUM

#### etcd Health (Week 1, ~8 hours)

**Task 15: Complete etcd Parser**
```go
// Expand internal/bundle/etcd.go
func ParseEtcdAlarmList(extractPath string) ([]EtcdAlarm, error)
func ParseEtcdMemberList(extractPath string) ([]EtcdMember, error)
func ParseEtcdEndpointStatus(extractPath string) ([]EtcdEndpoint, error)
```

**Task 16: etcd Health View**
- Add dedicated etcd view
- Show cluster health
- Show alarms (space quota exceeded)
- Show member list & roles

**Estimated Time**: 8 hours

#### Network Troubleshooting (Week 2, ~10 hours)

**Task 17: iptables & Routing Parser**
```go
// File: internal/bundle/networking.go (NEW)
func ParseIptables(extractPath string) (*IPTablesRules, error)
func ParseRoutes(extractPath string) ([]RouteInfo, error)
```

**Task 18: Network Diagnostics View**
- Show iptables rule count
- Show routing table
- Flag unusual configurations

**Estimated Time**: 10 hours

---

## UI Transparency Enhancements

### Bundle Completeness Indicator (v0.7.0)

**Implementation** (4 hours):
```go
// File: internal/bundle/health.go
type BundleCompleteness struct {
    TotalCategories int
    PresentCategories int
    MissingCategories []string
    PercentComplete int
}

func CalculateCompleteness(bundle *Bundle) *BundleCompleteness {
    // Check for presence of:
    // - kubectl resources (8 categories)
    // - systeminfo (6 critical files)
    // - journald logs (3 files)
    // - networking (4 categories)
    // - etcd (4 files)
    // Total: 25 categories
}
```

**UI Display**:
```
Status Bar: 📦 BUNDLE 76% (19/25)  [Press 'i' for details]

Bundle Info Panel (press 'i'):
┌─ Bundle Completeness ─────────────────────┐
│ ✅ kubectl resources     8/8   100%        │
│ ⚠️  System info          4/6    67%        │
│     Missing: dmesg, top                    │
│ ❌ Journald logs         0/3     0%        │
│     Missing: rke2-server, rancher-agent    │
│ ✅ Networking           4/4   100%        │
│ ⚠️  etcd diagnostics     2/4    50%        │
│     Missing: alarmlist, memberlist         │
│                                            │
│ Overall: 18/25 (72%)                       │
│                                            │
│ 💡 To collect missing data:               │
│   rancher-support-bundle --include-all    │
└────────────────────────────────────────────┘
```

### Data Gap Hints (v0.7.0)

**Implementation**:
- When viewing pods with storage issues, show: "💡 PV/PVC data available - press 's' for Storage view"
- When control plane pods have issues: "💡 RKE2 server logs available - press 'j' for Journald"
- In dashboard: "ℹ️  ConfigMap data found - may help debug configuration issues"

---

## Testing Strategy

### Phase 1 Testing (v0.7.0)
- Test with complete bundles (all files present)
- Test with partial bundles (selected files missing)
- Test parser resilience to malformed data
- Performance: <500ms to parse all Phase 1 files

### Phase 2 Testing (v0.7.5)
- Network troubleshooting scenarios
- Workload failure scenarios
- Bundle completeness indicator accuracy

### Phase 3 Testing (v0.8.0)
- etcd failure scenarios
- Network policy debugging
- Complete end-to-end coverage

---

## Success Metrics

| Release | Files Parsed | % Coverage | Key Additions |
|---------|-------------|-----------|---------------|
| v0.6.8 (current) | 9 | 30% | Baseline |
| v0.7.0 | 25 | 60% | Storage, systeminfo, journald, configmaps |
| v0.7.5 | 35 | 70% | Networking, additional workloads |
| v0.8.0 | 40+ | 90%+ | etcd, RBAC, webhooks, crictl |

**Target**: Parse 90%+ of high-value bundle data by v0.8.0

---

## Effort Summary

| Phase | Release | Hours | Key Deliverables |
|-------|---------|-------|------------------|
| Phase 1 | v0.7.0 | 45 | Storage, systeminfo, journald, config |
| Phase 2 | v0.7.5 | 22 | Network, workloads |
| Phase 3 | v0.8.0 | 18 | etcd, advanced diagnostics |
| **Total** | | **85** | **90% data extraction** |

---

## References

- **PRINCIPLES.md** - Maximum Information Extraction principle
- **FUTURE_WORK.md** - Detailed file inventory
- **ROADMAP_V0.6-V0.8.md** - Release schedule integration

---

**Maintained By**: Development Team  
**Last Updated**: 2026-01-15
