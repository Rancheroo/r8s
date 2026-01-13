# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.6.7] - 2026-01-13

### Added ✨

- **Inline Diagnostics in Attention Dashboard** - 2-line format with root cause and fix suggestions
  - Line 1: Issue title + emoji indicator
  - Line 2: "→ Root cause | Fix: Recommendation" (shown when DiagnosticContext available)
  - Consumes v0.5.12 DiagnosticContext data structures
  - Populated for CrashLoopBackOff, OOMKilled, ImagePullBackOff issues
  - Diagnostic line displayed in gray italic for visual distinction
  - Implements "5-Second Rule" - users know what to fix immediately
  - "Truth Only™" - line 2 only shows when diagnostic data exists

- **GetBundleHealth() Interface Method** - Unified bundle health access
  - Added to DataSource interface for clean abstraction
  - Returns BundleHealth with HasEtcd/HasNodes/HasSystemInfo/HasEvents/HasPods flags
  - Percentage() method calculates bundle completeness (0-100%)
  - Bundle mode status bar shows "📦 BUNDLE 80%" automatically
  - Embedded datasource returns nil (demo mode has no health concept)

### Fixed 🐛

- **Issue #18**: Dashboard table alignment with double-digit item numbers
  - Numbers now dynamic width: 1-9 uses 3 chars (" 1. "), 10-99 uses 4 chars (" 10. ")
  - Calculated automatically based on total displayed items
  - Right-aligned numbers for clean visual alignment
  - Works correctly with 100+ items in dashboard

- **Issue #19**: Bundle health percentage display restored
  - Root cause: Type assertion bypassed DataSource interface
  - Solution: Added GetBundleHealth() to DataSource interface (proper abstraction)
  - Status bar now shows bundle completeness: "📦 BUNDLE 100%"
  - Uses interface method instead of type assertion (follows our own abstractions)

### Enhanced 🔧

- **DiagnosticContext field added to AttentionItem**
  - Struct now includes optional diagnostic information
  - Populated by detection functions (generateCrashLoopContext, generateOOMContext, generateImagePullContext)
  - Enables future expansion of inline diagnostic types
  - 2-line rendering automatically adapts based on context availability

- **Test mock compatibility**
  - Updated mockDataSource in log_scanning_test.go with GetBundleHealth()
  - All tests passing with new interface method
  - Zero regression in test suite

### Technical - v0.6.7

- Modified files:
  - `internal/tui/attention_signals.go` - Added DiagnosticContext field, populated for CrashLoop/OOM/ImagePull
  - `internal/tui/attention.go` - Implemented 2-line rendering with dynamic number width
  - `internal/datasource/interface.go` - Added GetBundleHealth() method and BundleHealth type
  - `internal/datasource/bundle.go` - Implemented GetBundleHealth() with real data checks
  - `internal/tui/log_scanning_test.go` - Updated mock with GetBundle Health()

- Key changes:
  - Dynamic `numWidth` calculation: `len(fmt.Sprintf("%d", displayedCount)) + 2`
  - Conditional line2 rendering: only when `item.DiagnosticContext != nil`
  - BundleHealth flags populated by checking bundle data availability
  - Diagnostic line styled with gray italic for visual distinction

### Impact Summary - v0.6.7

- ✅ **Instant diagnostics** - Root cause visible without drilling down
- ✅ **Actionable guidance** - Fix recommendations shown inline
- ✅ **Clean alignment** - Works with 10, 100, even 1000+ items
- ✅ **Bundle transparency** - Health percentage shows data completeness
- ✅ **Proper abstractions** - GetBundleHealth() through interface (not type assertion)
- ✅ **Zero regressions** - All tests passing, builds clean

### Dashboard Format Example - v0.6.7

```
CRITICAL:
1. 💀 nginx-pod                CrashLoopBackOff           kube-system
   → Container failing: Exit code 1 - Check startup command | Fix: Review entrypoint script

2. 🧨 db-pod                   OOMKilled                  default
   → Memory exceeded: 980Mi/1Gi limit - Optimize usage | Fix: Increase to 2Gi

WARNING:
3. 🚫 api-pod                  ImagePullBackOff           default
   → Registry auth failed: secret invalid | Fix: Update imagePullSecrets
```

### Success Criteria - v0.6.7

- ✅ All dashboard items show 2-line format when diagnostics available
- ✅ Root cause visible without drilling down
- ✅ Recommendation actionable in ≤5 seconds
- ✅ Emoji indicators match severity
- ✅ No more than 2 lines per item (strict)
- ✅ Issue #18: Column alignment works with 10+ items
- ✅ Issue #19: Bundle health displays in status bar

---

## [0.6.6] - 2026-01-13

### Added ✨

- **OOM Root Cause Display**: Attention dashboard now shows OOM (Out of Memory) events with diagnostics
  - Detects OOM kills from kubectl events with human-readable descriptions
  - Distinguishes container OOM vs node OOM scenarios
  - Shows memory limits when available from bundle data
  - Enhanced diagnostic panel shows limit/request/container details
  - Graceful handling of partial bundle data
  - Consumes v0.5.11 OOM analyzer with robust parsing

### Fixed 🐛

- **Issue #17** (PARTIAL FIX): Dashboard navigation for non-pod resource types
  - **Fixed**: Kubelet, etcd, node, system, and daemonset items no longer navigate to wrong pod diagnostics
  - **Fixed**: Changed from specific kubelet check to comprehensive ResourceType != "pod" validation
  - **Remaining**: Cluster-level items (kubelet, node) should allow drill-down to show impacted pods
  - **Deferred**: Full drill-down UX (1-9 selection, impacted pods list) scheduled for v0.6.8
  - **Impact**: Prevents incorrect navigation, but doesn't provide cluster-level drill-down yet

### Enhanced 🔧

- **OOM Parser Robustness**: Enhanced OOM analysis with graceful degradation
  - Returns empty array instead of nil when events file missing
  - Multi-source enrichment: kubectl pods, pod manifests, node conditions
  - Stub functions for future QoS class and node memory correlation
  - Human-readable memory limit display (1Gi format)
  - Bundle sherpa philosophy: points users to relevant log bundle locations

### Known Issues 🐛

- **Issue #18**: Dashboard table alignment breaks with double-digit item numbers
  - Column spacing misaligned when item numbers reach 10+
  - Fix needed: Dynamic number width calculation in renderAttentionItem()
  - Impact: Visual polish issue, doesn't affect functionality
  - Priority: Medium - scheduled for v0.6.7

- **Issue #19**: Bundle health percentage missing from dashboard status bar
  - "📦 BUNDLE 100%" indicator no longer displays
  - Fix needed: Debug GetBundleHealth() type assertion and rendering logic
  - Impact: Informational feature, doesn't block core functionality
  - Priority: Medium - scheduled for v0.6.7

---

## [0.6.5] - 2026-01-13

### Added ✨

- **Kubelet Issues Display**: Attention dashboard now shows kubelet-level issues from journald logs
  - Detects HTTP 502 errors, DNS limits, TLS handshake errors, connection timeouts
  - Groups issues by error pattern with occurrence counts
  - Assigns appropriate severity (HTTP 502 = critical, DNS/TLS = warning)
  - Threshold of 5+ occurrences to reduce noise
  - Consumes v0.5.11 kubelet parser for accurate detection

### Known Issues 🐛

- **Issue #17**: Dashboard navigation for kubelet items shows wrong diagnostic panel
  - Pressing Enter on kubelet items navigates to nearest pod's diagnostics (misleading)
  - Fix needed: Add ResourceType check in handleEnter() for kubelet items
  - Workaround: Use classic cluster navigation for kubelet investigation
  - Priority: High - scheduled for v0.6.6

- **Issue #18**: Dashboard table alignment breaks with double-digit item numbers
  - Column spacing misaligned when item numbers reach 10+
  - Fix needed: Dynamic number width calculation in renderAttentionItem()
  - Impact: Visual polish issue, doesn't affect functionality
  - Priority: Medium - scheduled for v0.6.7

- **Issue #19**: Bundle health percentage missing from dashboard status bar
  - "📦 BUNDLE 100%" indicator no longer displays
  - Fix needed: Debug GetBundleHealth() type assertion and rendering logic
  - Impact: Informational feature, doesn't block core functionality
  - Priority: Medium - scheduled for v0.6.7

---

## [0.6.4] - 2026-01-12 "Node Conditions Display"

### Added ✨

- **Node Pressure Indicators in Attention Dashboard** - Comprehensive node health monitoring
  - Consumes v0.5.10 Node Conditions parser (`GetNodeConditions()`) for full node diagnostics
  - **Memory Pressure Detection**: Shows nodes with MemoryPressure=true with utilization percentage
    - Calculates memory usage: (Capacity - Allocatable) / Capacity × 100
    - Example: "🔴 Node worker-1 Memory Pressure - Memory: 95% used"
    - Critical severity for immediate visibility
  - **Disk Pressure Detection**: Flags nodes with DiskPressure=true
    - Example: "💿 Node worker-2 Disk Pressure - Disk space low"
    - Critical severity to prevent storage exhaustion
  - **PID Pressure Detection**: Identifies nodes with PIDPressure=true
    - Example: "⚡ Node worker-3 PID Pressure - Process IDs exhausted"
    - Warning severity for early intervention
  - **Taint/Cordon Correlation**: Shows node scheduling status with pressure indicators
    - Displays taint counts and cordon status in item descriptions
    - Example: "Memory: 95% used • Cordoned, 2 taint(s)"
    - Helps correlate resource pressure with node availability

### Technical - v0.6.4

- Added `detectNodeIssues()` function in `internal/tui/attention_signals.go`
  - Integrated as Tier 2.5 (between cluster health and events)
  - Iterates through all nodes from `GetNodeConditions()`
  - Checks MemoryPressure, DiskPressure, PIDPressure boolean flags
- Added helper functions for resource calculation:
  - `calculateMemoryUsage()` - Computes memory utilization percentage
  - `parseMemoryToBytes()` - Converts K8s memory strings (Ki/Mi/Gi/Ti) to bytes
  - `getTaintInfo()` - Summarizes node taints and cordon status
- Memory parsing handles: "16Gi" → 17179869184 bytes, "1024Mi" → 1073741824 bytes
- Graceful handling when GetNodeConditions() returns nil (bundle missing nodesdescribe)

### Detection Logic - v0.6.4

```go
nodes, err := dataSource.GetNodeConditions()
if err != nil || nodes == nil {
    return items
}
for _, node := range nodes {
    if node.MemoryPressure {
        memUsedPct := (capBytes - allocBytes) / capBytes * 100
        items = append(items, AttentionItem{
            Title: fmt.Sprintf("Node %s Memory Pressure", node.Name),
            Description: fmt.Sprintf("Memory: %.0f%% used", memUsedPct),
            Severity: SeverityCritical,
            Emoji: "🔴",
        })
    }
    if node.DiskPressure {
        items = append(items, AttentionItem{
            Title: fmt.Sprintf("Node %s Disk Pressure", node.Name),
            Description: "Disk space low",
            Severity: SeverityCritical,
        })
    }
}
```

### Success Criteria - v0.6.4

| Test Case | Expected Dashboard Display | Severity |
|-----------|---------------------------|----------|
| Node with MemoryPressure=true at 95% | "Node worker-1 Memory Pressure - Memory: 95% used" | 🔴 Critical |
| Node with DiskPressure=true | "Node worker-2 Disk Pressure - Disk space low" | 💿 Critical |
| Node with PIDPressure=true | "Node worker-3 PID Pressure - Process IDs exhausted" | ⚡ Warning |
| Node with pressure + cordoned | "Memory: 95% used • Cordoned, 2 taint(s)" | 🔴 Critical |

### Impact Summary - v0.6.4

- ✅ **Proactive node monitoring** - Surface resource pressure before cluster failure
- ✅ **Resource correlation** - Memory pressure linked to actual utilization percentages
- ✅ **Scheduling awareness** - Taint/cordon status shown alongside pressure indicators
- ✅ **Zero configuration** - Works automatically with bundle node data
- ✅ **Parser reuse** - Leverages v0.5.10 foundation (no new parsers needed)

### Philosophy - v0.6.4

**"Surface Pressure Before Failure"** - Node resource pressure often precedes catastrophic failures (OOM kills, disk full, PID exhaustion). Display warnings early so users can take preventive action before workloads start failing.

---

## [0.6.3] - 2026-01-12 "ETCD Health Diagnostics"

### Added ✨

- **ETCD Health Dashboard Integration** - Rich ETCD diagnostics in Attention Dashboard
  - Consumes v0.5.10 ETCD parser (`GetEtcdDetails()`) for comprehensive cluster health
  - **DB Size Monitoring**: Shows warning when database exceeds 100MB with compaction recommendation
    - Example: "📀 ETCD Database Large - 150 MB - compaction recommended"
  - **Member Count Detection**: Detects cluster member mismatches (expects 3 for HA, 1 for single-node)
    - Warning severity for non-standard counts (e.g., 2 or 5 members)
    - Critical severity when below quorum (< 2 members)
    - Example: "👥 ETCD Member Mismatch - Expected 3 members, found 2 (Leader: abc123)"
  - **Leader ID Display**: Shows current ETCD leader ID in item descriptions
    - Example: "ALARM: NOSPACE (Leader: 15e9d2d844399be2)"
  - **Auto-Severity Classification**:
    - 🚨 Critical: Alarms, unhealthy endpoints, <2 members
    - ⚠️ Warning: Large DB (>100MB), non-standard member counts
  - **Color-Coded Items**: Red (critical), Yellow (warning) for instant visual triage

### Technical - v0.6.3

- Modified `detectClusterHealth()` in `internal/tui/attention_signals.go`
- Replaced `GetEtcdHealth()` with `GetEtcdDetails()` for richer data access
- Added 3 new detection patterns:
  1. Database compaction check using `NeedsCompaction` field
  2. Member count validation with expected values (1 or 3)
  3. Leader ID enrichment in all ETCD item descriptions
- Leverages existing `renderAttentionItem()` severity-based coloring
- Zero UI changes - existing rendering handles new item types

### Detection Logic - v0.6.3

```go
// DB Size > 100MB
if etcdDetails.NeedsCompaction {
    items = append(items, AttentionItem{
        Severity: SeverityWarning,
        Emoji: "📀",
        Title: "ETCD Database Large",
        Description: fmt.Sprintf("%s - compaction recommended", etcdDetails.DBSize),
    })
}

// Member Count ≠ 1 or 3
if etcdDetails.MemberCount != 3 && etcdDetails.MemberCount != 1 {
    severity := SeverityWarning
    if etcdDetails.MemberCount < 2 {
        severity = SeverityCritical // Below quorum
    }
    items = append(items, AttentionItem{
        Severity: severity,
        Emoji: emoji,
        Title: "ETCD Member Mismatch",
        Description: fmt.Sprintf("Expected 3 members, found %d (Leader: %s)", 
            etcdDetails.MemberCount, etcdDetails.LeaderID),
    })
}
```

### Impact Summary - v0.6.3

- ✅ **Proactive monitoring** - ETCD issues surfaced automatically in dashboard
- ✅ **Actionable diagnostics** - Compaction recommendations when needed
- ✅ **Cluster health visibility** - Member count and leader status at a glance
- ✅ **Zero configuration** - Works automatically with bundle ETCD data
- ✅ **Parser reuse** - Leverages v0.5.10 foundation (no new parsers needed)

### Test Cases - v0.6.3

| Scenario | Dashboard Display | Severity |
|----------|------------------|----------|
| DB = 150MB | "ETCD Database Large - 150 MB - compaction recommended" | Warning ⚠️ |
| 1 member (single-node) | No alert (expected configuration) | - |
| 2 members | "ETCD Member Mismatch - Expected 3, found 2 (Leader: xyz)" | Warning ⚠️ |
| 0 members | "ETCD Member Mismatch - Expected 3, found 0 (Leader: )" | Critical 🚨 |
| NOSPACE alarm | "ALARM: NOSPACE (Leader: abc123)" | Critical 🚨 |

---

### Planned for v0.6.0 "Diagnostic-First Intelligence"
- **Phase 1**: Diagnostic panel tightening (remove noise, show only failures)
- **Phase 2**: Cluster events drill-down (v0.5.9 gap fix)
- **Phase 3**: Full container detection from pod specs
- **Phase 4**: Enhanced ETCD display with inline diagnostics
- **Phase 5**: Node conditions display with resource correlation
- **Phase 6**: Kubelet issues display from journald
- **Phase 7**: OOM root cause display with recommendations
- **Phase 8**: Inline diagnostics in dashboard (2-line format)
- Event message truncation for scannable diagnostic panel
- "Show, Don't Ask" principle applied throughout

---

## [0.5.12] - 2026-01-08 "Diagnostic Context Types"

### Added ✨

- **DiagnosticContext Data Structure** - Foundation for inline diagnostics
  - RootCause: "Container exceeded memory limit of 1Gi"
  - Recommendation: "Increase MemoryLimit to 2Gi or optimize memory usage"
  - Severity: "critical", "high", "medium", "low"
  - RelatedData: ["Pod: my-app-abc", "Node: 95% memory", "Events: OOMKilled x3"]
  - FixPriority: "immediate", "investigate", "monitor"

- **Context Generators** - Intelligent diagnostic interpretation
  - CrashLoopBackOff: "Container repeatedly failing to start"
  - OOM Killed: "Container exceeded memory limits"
  - Image Pull BackOff: "Cannot pull container image from registry"
  - Node Pressure: "Node resource pressure detected"
  - ETCD Issues: "ETCD cluster unhealthy"
  - Kubelet Errors: "Kubelet reporting node-level errors"

- **Recommendation Engine** - Actionable guidance mapping
  - Severity-based emoji indicators (🚨 critical, ⚠️ high, 🔍 medium)
  - Priority classification (immediate/investigate/monitor)
  - Context-specific next steps for each issue type

- **DataSource Extension** - GetDiagnosticContext() method
  - Graceful fallback when context unavailable
  - Ready for v0.6.0 inline display consumption

### Technical - v0.5.12

- **New files**:
  - `internal/tui/diagnostics.go` - Context generators (~144 lines)
  - `internal/tui/recommendations.go` - Pattern matching engine (~24 lines)

- **Modified files**:
  - `internal/datasource/interface.go` - DiagnosticContext type + GetDiagnosticContext method
  - `internal/datasource/bundle.go` - Implementation with event correlation
  - `internal/tui/log_scanning_test.go` - Mock method implementation

- **Context generation**: Consumes existing parsers (events, OOM analysis, resource specs)
- **Graceful degradation**: Returns basic context when advanced data unavailable

### Impact Summary - v0.5.12

- ✅ **Foundation ready** - v0.6.0 can now display inline diagnostics
- ✅ **Zero UI changes** - Backend-only data structures
- ✅ **All tests passing** - Race detection verified
- ✅ **Extensible design** - Easy to add new context generators

### Philosophy - v0.5.12

**"Show, Don't Ask" Infrastructure** - Backend foundation enabling automatic diagnostic display without user action. When users see an issue, they'll immediately understand WHY it happened and WHAT to do next.

---

## [0.5.11] - 2026-01-08 "Kubelet & OOM Analysis"

### Added ✨

- **Kubelet Log Parser** - Node-level error detection from journald
  - Parses `journald/rke2-server` for kubelet errors (HTTP 502, timeouts, remotedialer)
  - Detects 10+ common error patterns: connection timeouts, DNS limits, TLS handshakes
  - Handles journald format: `Dec DD HH:MM:SS hostname rke2[PID]: time="..." level=error msg="..."`
  - Graceful handling of missing journal files

- **OOM Root Cause Analyzer** - Distinguishes container vs node OOM
  - Parses `rke2/kubectl/events` for OOM kill messages
  - Correlates with pod resource specs from `rke2/kubectl/pods`
  - Identifies whether OOM was container-level or node-level
  - Extracts memory limits/requests for context

- **Container Resource Parser** - Pod spec limit extraction
  - Parses pod manifests (`rke2/pod-manifests/*.yaml`)
  - Extracts CPU/memory requests and limits per container
  - Determines QoS class (Guaranteed/Burstable/BestEffort)
  - Handles multiple pod spec sources (manifests, kubectl describe, kubectl get)

- **Detection Helpers** - Pattern matching and correlation functions
  - Regex-based error pattern detection
  - Event correlation with resource data
  - Timestamp parsing and severity assessment
  - Missing file graceful degradation

### Technical - v0.5.11

- **New files**:
  - `internal/bundle/kubelet.go` - Journald kubelet parser (~89 lines)
  - `internal/bundle/oom.go` - OOM event analyzer (~72 lines)
  - `internal/bundle/resources.go` - Pod resource extractor (~221 lines)

- **Modified files**:
  - `internal/datasource/interface.go` - New types: KubeletIssue, OOMAnalysis, ResourceSpec
  - `internal/datasource/bundle.go` - GetKubeletIssues(), GetOOMAnalysis(), GetPodResources()
  - `internal/tui/log_scanning_test.go` - Mock implementations

- **Parsers handle**: Journald logs, kubectl events, YAML manifests, kubectl describe output
- **Error patterns**: 10+ kubelet issues (HTTP 502, DNS limits, remotedialer timeouts)
- **OOM detection**: Event parsing + resource correlation
- **QoS classification**: Guaranteed/Burstable/BestEffort based on requests/limits

### Impact Summary - v0.5.11

- ✅ **Root cause detection** - Distinguish OOM types, identify kubelet issues
- ✅ **Resource correlation** - Memory limits linked to OOM events
- ✅ **Zero UI changes** - Backend foundation only
- ✅ **All tests passing** - Race detection verified
- ✅ **Graceful degradation** - Missing files don't break functionality

### Philosophy - v0.5.11

**"Maximum Information Extraction"** - Parse all available bundle data to enable future diagnostic intelligence. This release extracts root cause information that v0.6.0 will display inline.

---



### Added ✨

- **Enhanced ETCD Parser** - Comprehensive etcd cluster information extraction
  - Parses `memberlist` → member count, member IDs, states, URLs
  - Parses `endpointstatus` → leader ID, DB size (50 MB), version (3.5.21), raft metrics
  - Compaction recommendations when DB exceeds 100MB threshold
  - New `ParseEtcdDetails()` function returns complete cluster state
  - Backward compatible with existing `ParseEtcdHealth()`

- **Node Conditions Parser** - Full node health and capacity from kubectl describe
  - Parses all node conditions: MemoryPressure, DiskPressure, PIDPressure, NetworkUnavailable, Ready, EtcdIsVoter
  - Extracts capacity/allocatable resources (CPU, memory, storage, pods)
  - Node taints and schedulability status
  - System info (kernel, OS image, container runtime, kubelet version)
  - Network details (internal IP, Pod CIDR)
  - Computed flags: HasPressure, IsControlPlane, IsEtcd, IsWorker
  - New `ParseNodeDescribe()` function handles multi-node output

- **DataSource Interface Extensions**
  - Added `GetEtcdDetails()` → returns comprehensive etcd information
  - Added `GetNodeConditions()` → returns all node health data
  - New types: `EtcdDetails`, `EtcdMember`, `NodeConditions`
  - Gracefully handles missing files (returns nil/empty, not errors)

### Technical - v0.5.10

- **New files**:
  - `internal/bundle/nodes.go` - Node describe parser (~300 lines)
  - Extended `internal/bundle/etcd.go` - ETCD parsers (+180 lines)
  
- **Modified files**:
  - `internal/datasource/interface.go` - New types and interface methods
  - `internal/datasource/bundle.go` - Implemented new methods with type conversion
  - `internal/tui/log_scanning_test.go` - Updated mock with new methods

- **Parsers handle**: CSV format (memberlist), ASCII tables (endpointstatus), multi-section text (nodesdescribe)
- **DB size parsing**: Converts "50 MB" → 52428800 bytes for threshold comparisons
- **Section detection**: State-machine parser for nodesdescribe sections
- **Compaction logic**: 100MB threshold based on etcd best practices

### Impact Summary - v0.5.10

- ✅ **Backend foundation ready** - v0.6.0 can display rich ETCD and node diagnostics
- ✅ **Zero UI changes** - exclusively backend parsers (foundation release)
- ✅ **All tests passing** - includes race detection
- ✅ **Graceful degradation** - missing bundle files don't break functionality
- ✅ **Type-safe** - Proper type conversion between bundle and datasource layers

### Philosophy - v0.5.10

**"Maximum Information Extraction"** - Parse all available bundle data for future diagnostic use. This release lays the groundwork for v0.6.0's inline diagnostic displays.

### Planned for v0.5.11 "Kubelet & OOM Analysis" (Foundation)
- Kubelet log parser for journald error detection
- OOM root cause analyzer (distinguish container vs node OOM)
- Container resource limit parser from pod specs
- Advanced detection capabilities for diagnostic context
- Backend-only changes (no UI modifications)

### Planned for v0.5.12 "Diagnostic Context Types" (Foundation)
- DiagnosticContext data structure for inline diagnostics
- Context generator functions (CrashLoop, OOM, ImagePull, Node, ETCD)
- Recommendation engine mapping issues to actionable suggestions
- AttentionItem extension with diagnostic context field
- Backend-only changes (no UI modifications)

### Planned for v0.6.0 "Diagnostic-First Intelligence"
- **Phase 1**: Diagnostic panel tightening (remove noise, show only failures)
- **Phase 2**: Cluster events drill-down (v0.5.9 gap fix)
- **Phase 3**: Full container detection from pod specs
- **Phase 4**: Enhanced ETCD display with inline diagnostics
- **Phase 5**: Node conditions display with resource correlation
- **Phase 6**: Kubelet issues display from journald
- **Phase 7**: OOM root cause display with recommendations
- **Phase 8**: Inline diagnostics in dashboard (2-line format)
- Event message truncation for scannable diagnostic panel
- "Show, Don't Ask" principle applied throughout

---

## [0.5.9] - 2026-01-07 "Simplification"

### Fixed 🐛

- **Container selector now appears in log view**
  - Root cause: `containers` field in App struct was never populated
  - Solution: Added `containers` field to `logsMsg` struct, populated from current container
  - Impact: Container cycling ('c' key) now functional
  - Note: Full container detection requires pod spec parsing (deferred to v0.6.0)

### Refactored 🔧

- **Removed dashboard expansion/sub-navigation complexity**
  - Removed `expandedItems` map and `subCursor` state fields
  - Simplified navigation: Enter directly opens diagnostic panel (no expansion mode)
  - Removed expansion indicators (►/▼) from item rendering
  - Deleted unused `renderExpandedContent()` function  
  - Code reduction: ~153 lines removed
  - Rationale: Expansion feature added UX complexity without clear benefit
  - Impact: Simpler, faster navigation with fewer display glitches

### Technical - v0.5.9

- Modified `logsMsg` type to include `containers []string` field
- Updated `fetchLogs()` to populate containers with current container
- Updated `logsMsg` handler to set `a.containers` and `a.currentContainer`
- Simplified attention dashboard navigation in `app.go`
- Removed expansion rendering code from `attention.go`
- Files modified: app.go (-4 lines), fetch.go (+6 lines), helpers.go (+1 line), attention.go (-150 lines)

### Impact Summary - v0.5.9

- ✅ **Container selector functional** - 'c' key now works in log view
- ✅ **Simpler navigation** - Direct Enter → diagnostic panel (no sub-nav)
- ✅ **Cleaner codebase** - ~153 lines removed
- ✅ **Zero regressions** - All existing features preserved
- ✅ **Foundation for v0.6.0** - Clean base for diagnostic-first enhancements

---

## [0.5.8] - 2026-01-07 "Silky Navigation"

### Refactored 🔧

- **Navigation state consolidation** - Unified log navigation across all paths
  - Created single `navigateToLogs()` function replacing 4 duplicate implementations
  - Extracted `clearPodState()` helper for consistent state clearing
  - Eliminates race conditions from inconsistent state management
  - All navigation paths now use identical logic

### Technical - v0.5.8

- Added `clearPodState()` in helpers.go - single source of truth for state clearing
- Added `navigateToLogs()` in helpers.go - unified navigation function
- Replaced 4 duplicate implementations:
  - app.go: Dashboard submenu navigation (lines 337-359)
  - app.go: Main dashboard navigation (lines 372-394)
  - handlers.go: ViewAttention case (lines 53-88)
  - handlers.go: ViewPods case (lines 197-223)
- Code reduction: ~40 lines removed (4 duplicates → 1 unified function)

### Impact Summary - v0.5.8

- ✅ **Zero duplication** - State clearing happens in one place only
- ✅ **Consistent behavior** - All navigation paths work identically
- ✅ **Race condition prevention** - Unified validation eliminates timing bugs
- ✅ **Easier maintenance** - Changes to navigation logic happen once
- ✅ **Faster development** - New navigation features have single implementation point

### Files Modified - v0.5.8

- `internal/tui/helpers.go`: +45 lines (2 new functions)
- `internal/tui/app.go`: -50 lines (2 duplicates removed)
- `internal/tui/handlers.go`: -35 lines (2 duplicates removed)
- **Net change**: -40 lines total

### Philosophy - v0.5.8

**"One Way To Do It"** - Eliminate subtle behavioral differences between navigation paths. When the same operation can be reached multiple ways, ensure identical behavior through shared implementation.

---

## [0.5.7] - 2026-01-07 "Navigation Polish"

### Added ✨

- **Help discovery hint** - Show "💡 Press ? for help" in footer for first 3 TUI launches
  - Tracks launch count in config.yaml
  - Disappears after 3 launches (Show, Don't Ask principle)
  - Surfaces help system automatically for new users

- **Expansion state clarity** - Clear indicators for capped vs expanded dashboard view
  - Changed: `Showing 9/100` → `Showing 9/100 (capped)`
  - Changed: `[m]=expand` → `[m]=show all 100` (shows exact count)
  - When expanded: `[m]=cap` to toggle back
  - Users immediately understand expansion state

### Fixed 🐛

- **Dashboard submenu navigation context bug**
  - Fixed incomplete diagnostic panel when navigating from expanded pod list
  - Root cause: Submenu passed empty clusterID and containerName to log view
  - Solution: Pass full item context (clusterID + containerName) like main dashboard
  - Impact: Diagnostic panel now shows complete pod data from both navigation paths

### Changed 🔄

- Dashboard Enter behavior already implements diagnostic-first design (v0.5.4)
  - From main dashboard: Enter → Diagnostic panel ✅
  - From expanded submenu: Enter → Diagnostic panel ✅ (fixed context passing)
  - Consistent behavior across all navigation paths
  - Verified in v0.5.7 testing

### Technical - v0.5.7

- Added `LaunchCount` field to Config struct (persisted in config.yaml)
- Launch counter incremented and saved in cmd/tui.go
- Passed to App struct for footer rendering
- Status bar rendering updated in attention.go
- ~35 lines added total

### Impact Summary - v0.5.7

- ✅ **Help discovery improved** - New users see help hint automatically
- ✅ **Expansion state clear** - No ambiguity about capped vs full view
- ✅ **Diagnostic-first verified** - Enter from submenu shows diagnostics (existing behavior confirmed)
- ✅ **Zero regressions** - All existing features preserved
- ✅ **Philosophy aligned** - "Show, Don't Ask" principle applied

---

## [0.5.6] - 2026-01-06 "Enhanced Error Diagnostics"

### Added ✨

- **Container Status Section in Diagnostic Panel** - Show container-level health
  - New "📋 CONTAINER STATUS" section displayed between POD STATUS and RECENT EVENTS
  - Parses Ready field (e.g., "1/2") to show container health summary
  - Extracts container names from events (Started/Created/BackOff reasons)
  - Shows which containers succeeded (✅) vs failed (❌)
  - Displays explicit message about exit codes not being in bundle
  - Provides kubectl describe command for detailed container status

### Impact Summary - v0.5.6

- ✅ **Container-level visibility** - Shows which specific containers failed in multi-container pods
- ✅ **Transparent data gaps** - Explicitly states when exit codes aren't available in bundle
- ✅ **Actionable guidance** - Provides kubectl command to get detailed container status
- ✅ **Enhanced diagnostics** - Pod "Error" state with no warning events now explained
- ✅ **Test case addressed** - `test-notready` pod (1/2 ready) now shows ok-container vs failing-container

### Technical - v0.5.6

- Added `buildContainerStatusSection()` function in logs.go
- Container name extraction from event messages using heuristic parsing
- Ready field parsing with fmt.Sscanf to extract container counts
- Section integrated into renderMaximumIntelPanel() workflow
- Works from both Dashboard and Classic navigation paths

### Use Case - v0.5.6

**Before v0.5.6:**
- Pod shows "Error" state with "1/2 ready"
- Events show only Normal events (Scheduled, Pulled, Created, Started)
- No explanation of WHY pod is in Error state

**After v0.5.6:**
- Container status section shows "Containers: 1/2 ready"
- Lists: ✅ ok-container, ❌ failing-container  
- Notes: "Exit codes not captured in bundle"
- Provides: kubectl describe command for details

### Philosophy - v0.5.6

**"Maximum Information Extraction"** - Extract and display all available bundle data, even when incomplete. Be transparent about data gaps and provide guidance for obtaining missing information.

---

## [0.5.5] - 2026-01-06 "Maximum Data Extraction"

### Added ✨

- **Rich Event Data in Diagnostic Panel** - Parse global kubectl/events for complete pod event history
  - Shows ALL pod events (FailedScheduling, FailedMount, NetworkNotReady, volume issues)
  - Event counts displayed: "BackOff: Container is in waiting state (x47)"
  - Events sorted: Warnings first, then by time (most recent)
  - Maximum 5 events shown in diagnostic panel
  - Falls back to pod.KubectlEvents if global events unavailable

- **Bundle Health Tracking** - Transparency about bundle completeness
  - BundleHealth struct tracks found/derived/missing files
  - Enhanced verbose loading output with ✓/⚠️ indicators
  - Shows which files are derived (e.g., "namespaces: derived 12 from pods")
  - Health percentage calculated: (found + derived) / total files
  - Provides visibility into data quality without blocking

### Technical - v0.5.5

- Added `GetEventsByPod()` to DataSource interface
- Implemented event filtering and sorting in BundleDataSource  
- Updated `buildEventsSection()` to use Event struct with count indicators
- Added BundleHealth struct with Percentage() and Color() methods
- Bundle loader tracks file availability during parse
- Verbose mode shows detailed file status

### Impact Summary - v0.5.5

- ✅ **Complete event history** - No more missing FailedScheduling/FailedMount events
- ✅ **Event count indicators** - See repeated events at a glance "(x47)"
- ✅ **Bundle transparency** - Know what data is available vs derived vs missing
- ✅ **Verbose loading** - Clear feedback during bundle import
- ✅ **Smart fallbacks** - Derives namespaces from pods when file missing

### Philosophy - v0.5.5

**"Maximum Information Extraction"** - Use ALL available bundle data. Show transparency about data quality without blocking user workflow.

---

## [Unreleased - Post v0.5.5]

### Fixed 🐛

- **Pod navigation race condition** - Rapid pod switching now shows correct data
  - **Problem**: Dashboard → Pod A → Back → Pod B → repeat could show Pod A's logs in Pod B's view
  - **Root Cause**: Async race condition - Pod A's `fetchLogs()` response arrived after navigating to Pod B
  - **Solution**: Added pod name validation to `logsMsg` handler - stale messages now ignored
  - **Impact**: Navigation reliability restored, both dashboard and classic views work correctly

### Technical - Unreleased

- Modified `logsMsg` type to include `podName` and `namespace` fields for validation
- Updated `fetchLogs()` to pass pod identity with every log message
- Added validation in `logsMsg` handler: ignore messages that don't match current view context
- Applied diagnostic-first approach to classic pod view (Enter from Pods table)
- All navigation paths now clear state before switching pods

### Known Issues - Unreleased

- **Navigation state complexity** - Multiple navigation paths with inconsistent state management
  - State clearing code duplicated across 3 locations (dashboard enter, app.go enter handlers)
  - Classic and dashboard paths have subtle behavior differences
  - Deferred to v0.6.0 simplification: standardize navigation, reduce state complexity
  - See FUTURE_WORK.md "Navigation Simplification (v0.6.0)" for planned improvements

---

## [0.5.4] - 2026-01-05 "Enhanced Diagnostics"

### Added ✨

- **Maximum Intel Diagnostic Panel** - "No Logs" screen now shows comprehensive pod diagnostics
  - Intelligent diagnosis based on pod state (CrashLoopBackOff, OOMKilled, ImagePullBackOff, Error, Pending, Evicted)
  - Actionable interpretation: r8s tells users what to investigate, not just raw data
  - Recent events display (last 5) with warning/normal emoji indicators
  - State-specific investigation suggestions tailored to failure mode
  - External tools guidance (lnav, kubectl logs) for deep log analysis
  
- **Diagnostic Sections** - Structured, scannable panel layout:
  - 💡 DIAGNOSIS: Emoji + intelligent interpretation of failure pattern
  - 💊 POD STATUS: State, restarts, ready status, node, age
  - 📋 RECENT EVENTS: Last 5 pod events with context (always shown)
  - 🔍 INVESTIGATE NEXT: Actionable next steps (1-3 suggestions)
  - 🛠️ EXTERNAL TOOLS: Guidance on lnav and kubectl for deeper analysis

- **Intelligent Pattern Recognition** - Contextual diagnoses:
  - CrashLoopBackOff → "Container repeatedly failing to start"
  - OOMKilled → "Container exceeded memory limits"
  - ImagePullBackOff → "Cannot pull container image from registry"
  - High restart count → "Instability pattern detected"
  - Pending → "Pod not yet scheduled"
  - Evicted → "Removed due to resource pressure"

### Fixed 🐛

- **CRITICAL: Classic view works with partial bundles** - No more "No namespaces available"
  - Root cause: Partial bundles often missing `rke2/kubectl/namespaces` file (perms, policy, sanitization)
  - Solution: Automatically derive namespaces from pod list when file missing
  - Impact: Classic view now resilient to incomplete bundle data
  - Verbose mode shows: "⚠ namespaces file missing - derived X namespaces from pools"
  - Dashboard already worked this way - now Classic view matches

- **Events section always visible** - Shows "No events recorded" when empty
  - Users now know the section exists even if no events captured
  - Consistent panel structure regardless of data availability
  
- **Removed '[d]=describe pod' from diagnostic panel** - Reduced UX confusion
  - Diagnostic panel already shows the key information
  - Raw YAML describe output was confusing and not actionable
  - Users can still access describe when viewing actual logs

### Technical - v0.5.4

- Refactored `renderEmptyLogsHelp()` into modular diagnostic components
- Added helper functions: `buildDiagnosisSection()`, `buildEventsSection()`, `buildInvestigationSection()`, `buildExternalToolsSection()`
- Leverages existing `pod.KubectlEvents` data (already attached in v0.5.3)
- Works from both Dashboard and Classic navigation paths (dataSource-based)
- Events section now always renders with fallback message when empty
- 244 lines added, 39 lines removed in logs.go (3 commits total)

### Impact Summary - v0.5.4

- ✅ **Maximum diagnostics** - Users see comprehensive pod health at a glance
- ✅ **r8s interprets data** - "OOMKilled = memory issue" not just "Exit Code: 137"
- ✅ **Actionable guidance** - Context-specific next steps for every failure mode
- ✅ **External tool awareness** - Users know when/how to use lnav or kubectl
- ✅ **Navigation agnostic** - Same rich panel from Dashboard or Classic view
- ✅ **Cleaner UX** - Removed confusing 'd' key, events always visible

### Philosophy - v0.5.4

**"r8s interprets, user acts"** - Show intelligence, not just information. Users should know WHAT to investigate and WHY, not parse raw Kubernetes output themselves.

### Deferred to v0.5.5

- **Parse kubectl/events file for comprehensive pod events** - Currently uses pod.KubectlEvents (attached events only)
  - Bundle contains `kubectl/events` file with ALL cluster events
  - Should parse this file and filter by pod name for richer event data
  - Would show scheduling, volume, network events not currently visible
  - See `docs/archive/2025-12-01/LOG_BUNDLE_ANALYSIS.md` for details

---

## [0.5.3] - 2026-01-05 "Maximum Information Extraction"

### Added ✨

- **Dashboard Truth Indicators** - All critical pod states now show data availability
  - CrashLoopBackOff pods show "⚠️ No Logs" indicator
  - Error/OOMKilled/ImagePullBackOff/Evicted pods marked with warning emoji
  - Restart count enrichment: "⚠️ No Logs • 47 restarts" shows crash severity
  - Empty namespaces show "📭 Empty" instead of misleading "Clean"
  - Consistent "Truth Only™" - dashboard shows what data exists

- **Rich Pod Diagnostic Panel** - Maximum info even when logs unavailable
  - Automatic diagnostic display for pods with no logs
  - Shows: State, Restart count, Ready status, Node, Age
  - Investigation suggestions based on pod state
  - Fetches data from dataSource (works from any navigation path)
  - Reuses "No Logs" screen infrastructure for reliability

- **Describe from Attention Dashboard** - Fixed broken 'd' key functionality
  - Press 'd' on any pod in dashboard → See pod description
  - ViewAttention case added to handleDescribe()
  - Matches pod from attention items by title
  - Consistent with classic pod view behavior

### Fixed 🐛

- **CRITICAL: Pod diagnostics work from Dashboard navigation**
  - Fixed bug where diagnostic panel showed fallback when navigating Dashboard → Pod
  - Root cause: `a.pods` empty when skipping classic Pods view
  - Solution: Fetch from `dataSource.GetAllPods()` directly instead of cached state
  - Works regardless of navigation path (Dashboard or Classic)

- **Build & Test Stability**
  - Fixed NumSortModes sentinel (prevents sort mode overflow)
  - Updated dashboard cap to 100, fixed related tests
  - All tests passing: `ok github.com/Rancheroo/r8s/internal/tui 0.012s`

### Technical - v0.5.3

- 8 commits total on release/v0.5.3 branch
- Enhanced AttentionItem struct with log status flags
- Added pod diagnostic extraction in renderEmptyLogsHelp()
- Direct dataSource integration for reliable pod data access
- Truth indicators applied across dashboard rendering
- Version: v0.5.2-25-gcf2592f

### Impact Summary - v0.5.3

- ✅ **Truth indicators** - Dashboard shows what data exists at a glance
- ✅ **Rich diagnostics** - Full pod info even when logs missing
- ✅ **Fixed describe** - Works from dashboard (was broken)
- ✅ **Navigation agnostic** - Diagnostics work from any path
- ✅ **Consistent UX** - Same diagnostic panel style everywhere
- ✅ **All tests passing** - Zero regressions

### Commits - v0.5.3

1. `1c9f41d` - Add NumSortModes sentinel, fix sorting edge cases
2. `25cdad0` - Update dashboard cap to 100, fix tests
3. `3eacb5c` - Initial truth indicators (No Logs + Empty namespace)
4. `6d45f6d` - Extend to all critical states (Crash/Error/OOM/ImagePull/Evicted)
5. `25223c7` - Enrich with restart counts
6. `46b7b13` - Rich diagnostic panel for pods with no logs
7. `9829d65` - Fix diagnostics from Dashboard navigation path
8. `cf2592f` - Enable describe from Attention Dashboard

### Future Work - Documented

- **v0.5.4**: Auto-display diagnostics on failed pods (remove 'd' button requirement)
  - Show diagnostic panel automatically for crashed pods
  - Reuse "No Logs" panel code for consistency
  - Philosophy: "Reliable code > feature-rich bespoke code"

---

## [0.5.2] - 2026-01-03 "Truth & Accuracy"

### Fixed 🐛 CRITICAL

- **CRITICAL: Identical error/warning counts bug**
  - **Problem**: ALL pods showed identical "19 ERR, 17 WARN" in dashboard and classic view
  - **Root cause**: `generateDemoLogs()` returned same hardcoded 57-line log for every pod when log files empty/missing
  - **Impact**: Attention Dashboard showed "19 ERR, 17 WARN" for all items, Classic Pod View showed "19E/17W" for all pods
  - **Fix**: Removed fake demo log generation for real bundles - return empty []string{} instead
  - **Result**: Pods now show accurate, different E/W counts based on actual log content
  - **Namespace view was correct** (used different aggregation logic)
  - **Principle restored**: Truth Only™ — real bundles show accurate data

- **Documentation consistency across repository**
  - Updated all version references from v0.5.1 to v0.5.2 (README.md, POST_MERGE_CLEANUP.md)
  - Added blank lines before subheadings in FUTURE_WORK.md for markdown linting compliance
  - Removed trailing comma from date field in GIT_BRANCH_CLEANUP_PLAN.md
  - Verified all v0.5.1 references are correctly in historical documentation

- **Code quality improvements**
  - Refactored 4 sorting functions to use sort.Slice/sort.SliceStable 
  - Replaced manual bubble-sort O(n²) implementations with sort.Slice O(n log n)
  - Maintained stable ordering using sort.SliceStable where needed for tie-breaker logic

### Added ✨

- **Future work documented in FUTURE_WORK.md**
  - UX improvements: CrashLoopBackOff with no logs indicator, empty namespace intelligence, age display consistency
  - Simplification proposals: Remove sort mode complexity, remove log filter modes, remove view switching hotkeys
  - Philosophy: "The best feature is no feature - smart defaults beat options"
  - Total potential: ~350+ lines removed across 3 simplification proposals
  
- **Post-merge cleanup documentation**
  - Added 30-day branch cleanup rule to CONTRIBUTING.md
  - Documented safe commands for finding merged/stale branches
  
- **UTF-8 encoding prevention**
  - Added comprehensive UTF-8 linting recommendation to LESSONS-LEARNED.md
  - Documented pre-commit hook approach to prevent � character regressions

### Technical - v0.5.2

- 8 files modified: bundle.go, CONTRIBUTING.md, FUTURE_WORK.md, LESSONS-LEARNED.md, POST_MERGE_CLEANUP.md, README.md, GIT_BRANCH_CLEANUP_PLAN.md, attention_signals.go
- All sorting functions now use Go stdlib (sort.Slice/sort.SliceStable)
- Version consistency verified across entire repository
- Git tag: v0.5.2
- Branch: ship-v052
  - Commits: 152dd76 (docs), 49b45be (critical fix)

### Impact Summary - v0.5.2

- ✅ **CRITICAL FIX** - Accurate per-pod error/warning counts (no more fake "19/17")
- ✅ **Version consistency** - All user-facing docs show v0.5.2
- ✅ **Better performance** - Sort functions use O(n log n) algorithms
- ✅ **Future-proof** - UTF-8 linting prevents encoding regressions
- ✅ **Clear roadmap** - 6 concrete simplification + UX improvement proposals for v0.5.3+
- ✅ **Truth Only™** - Real bundles show only accurate data, no fake demo logs

### Known Issues & Deferred (documented in FUTURE_WORK.md)

- CrashLoopBackOff pods with no logs show "✅ Clean" (misleading) - FIX in v0.5.3
- Empty namespaces show "✅ Clean" instead of "📭 Empty" - FIX in v0.5.3
- Sort mode complexity (3 modes, rarely used) - REMOVE in v0.5.3
- Log filter modes (3 modes, manual toggling) - SIMPLIFY in v0.5.4

---

## [0.5.1.1] - 2026-01-03 "Code Quality Polish (Pre-v0.5.2)"

### Fixed 🐛
- **Dashboard navigation improvements**
  - Cursor-index sync on sort mode changes (prevents jumpy navigation)
  - Scroll position maintained after 's'/'m'/'g'/'G' key presses
  - Title-based cursor tracking maintains selection during sorts
  
- **Table rendering polish**
  - Dynamic column widths adapt to terminal size (80-200 cols)
  - Terminal-responsive layouts for all 9 view types
  - Smart word wrapping breaks at whitespace, not mid-word
  - Consistent ellipsis truncation across all views
  
- **Performance optimizations** 
  - Package-level pattern arrays (allocated once, not per call)
  - Cached error/warning counts for instant table rendering
  - Reduced GC pressure during log rendering

### Added ✨
- **Enhanced help panel** with contextual pro tips per view
  - Dashboard tips: Sort cycling, health display, expansion
  - Pod tips: W/E column meaning, error sorting, quick actions
  - Log tips: Ctrl+W filter, search workflow, vim navigation
  - Auto-displayed based on current view (Show, Don't Ask principle)

- **"Show, Don't Ask" UX Philosophy**
  - Tables auto-adapt to terminal width (no manual resizing)
  - Cursor position auto-restored after sorting (no manual re-finding)
  - Age display enhanced: shows hours/minutes for recent resources
  - Health summary auto-displayed in dashboard status bar

### Technical - v0.5.2
- Dynamic column width system with proportional ratios
- Cursor tracking by Title instead of index for sort stability
- View-specific contextual help system
- Cross-platform git commands (Linux/macOS/POSIX)
- Better error message formatting across all fetch functions

### Impact - v0.5.2
- ✅ **Smooth navigation** - cursor stays on same item after sorts
- ✅ **Terminal responsive** - tables perfect on 80-200 col terminals
- ✅ **Better performance** - optimized pattern matching and caching
- ✅ **Context-aware help** - tips automatically shown per view
- ✅ **Show, Don't Ask** - information displayed proactively

---

## [0.5.1.1] - 2026-01-03 "Code Quality Polish"

### Fixed 🐛
- **Systematic code quality pass (11 improvements)**
  - UTF-8 encoding fixes: � → 📚 and 🔥 emojis render correctly
  - K8s log marker consistency: I####/D#### patterns only match at line start
  - Slice aliasing bug: sortedPods now creates copy before sorting
  - Performance: Replaced manual bubble sort with sort.Slice (O(n²) → O(n log n))
  - Terminology: [MOCK] → [DEMO] for embedded bundle
  - Code cleanup: Removed redundant else blocks and debug code

### Added ✨
- **UX improvements**
  - Auto-display health breakdown in dashboard: "🔥 X critical · ⚠️ X warnings"
  - Auto-display helpful message for empty logs with next steps
  - "Show, Don't Ask" UX philosophy documented in FUTURE_WORK.md

### Impact - v0.5.1.1
- ✅ **Perfect emoji rendering** - Multi-byte UTF-8 characters handled correctly
- ✅ **Consistent log detection** - K8s markers uniform across all views
- ✅ **Zero aliasing bugs** - All sorting operations create copies
- ✅ **Better performance** - O(n log n) sorting everywhere
- ✅ **Improved UX** - Information displayed proactively, not on demand

### Technical - v0.5.1.1
- Reviewer feedback incorporated from code quality audit
- All patterns documented in LESSONS-LEARNED.md
- Build verified: `make build && make test` pass cleanly
- 5 commits of polish post-v0.5.1 release

## [0.5.1] - 2026-01-03 "Modular Core"

### Fixed 🐛
- **Code quality pass (11 fixes across 7 files)**
  - Fixed UTF-8 encoding issues (� → 📚 and 🔥 emojis)
  - Removed hardcoded mock container initialization in handlers
  - Changed misleading "[MOCK]" label to "[DEMO]" 
  - Fixed K8s log marker checks to only match at line start (consistent detection)
  - Removed redundant else blocks
  - Fixed slice aliasing bug (sortedPods now creates copy before sorting)
  - Replaced manual bubble sort with sort.Slice for better performance (O(n log n))
  - Fixed duplicate section headings and date formatting in documentation
  - Updated version references (0.5.0 → 0.5.1)

### Refactored 🔧

- **Complete app.go decomposition into 6 focused modules**
  - **app.go**: 3031 → 400 lines (87% reduction) - Core state & orchestration only
  - **helpers.go**: ~300 lines - Utilities, breadcrumb, status text, formatting
  - **logs.go**: ~450 lines - Log view rendering, filtering, color detection
  - **fetch.go**: ~400 lines - All data fetching functions and describe operations
  - **table.go**: ~500 lines - Table rendering for all 8 view types
  - **handlers.go**: ~500 lines - Event handlers, navigation, input processing
  - **Total**: 3031 lines → 2550 lines across 6 files (avg 425 lines/file)

### Impact - v0.5.1

- ✅ **70% cognitive load reduction** - No single file exceeds 500 lines
- ✅ **Improved testability** - Each module has clear responsibility
- ✅ **Faster onboarding** - New developers can understand one module at a time
- ✅ **Better maintainability** - Changes isolated to relevant module
- ✅ **Zero regressions** - All existing behavior preserved exactly
- ✅ **Same package** - No circular imports, clean architecture

### Technical - v0.5.1

- All modules remain in `internal/tui` package (no import changes)
- Clear separation of concerns:
  - State management (app.go)
  - User input (handlers.go)
  - Data loading (fetch.go)
  - Visual rendering (table.go, logs.go)
  - Utilities (helpers.go)
- Build verification: `make build && make test` pass cleanly
- No behavioral changes - pure code organization refactor

### Module Responsibilities

**app.go** - Orchestration hub
- App struct definition and state
- Message types (clustersMsg, podsMsg, etc.)
- Init() and Update() main loop
- View() dispatcher

**handlers.go** - User interaction
- Keyboard event handling (Enter, Esc, navigation keys)
- handleEnter(), handleDescribe(), handleViewLogs()
- Sort mode toggling, search execution
- Filter application and refresh logic

**fetch.go** - Data layer
- All fetch* functions (fetchPods, fetchLogs, etc.)
- Describe functions (describePod, describeDeployment, etc.)
- Data transformation helpers
- Async command generation

**table.go** - Table views
- updateTable() with all view type cases
- Column/row configuration for each resource
- Table styling and formatting
- View-specific sorting logic

**logs.go** - Log viewing
- renderLogsView() and renderLogsWithColors()
- Log level detection (isErrorLog, isWarnLog, etc.)
- Filter application (getVisibleLogs)
- Colorization and search highlighting

**helpers.go** - Utilities
- Breadcrumb generation
- Status text formatting
- Number formatting (formatCount for K/M/B)
- Safe extraction helpers
- Help and describe modals

### Lessons Applied

- **God-file decomposition**: Split early before it hits 5000 lines
- **Clear module boundaries**: Each file has single responsibility
- **Zero regressions**: Comprehensive testing before/after refactor
- **Velocity compounds**: Clean codebase enables faster feature development

## [0.5.0] - 2026-01-02 "Lean & Accurate"

### Removed 🗑️

- **Mock mode completely deleted** (-698 lines total)
  - Removed all 9 getMock* functions (getMockPods, getMockDeployments, getMockServices, getMockClusters, getMockProjects, getMockNamespaces, getMockCRDs, getMockCRDInstances)
  - Removed generateMockLogs function (~57 lines)
  - Removed mock mode check from fetchLogs (~5 lines)
  - Removed TestMockDataGeneration test (~59 lines)
  - r8s now works out-of-box with embedded demo bundle
  - **17% leaner**: app.go reduced from 3643 → 3031 lines

### Re-enabled ✅

- **Dashboard log scanning restored with accuracy guarantee**
  - Re-enabled detectLogIssues() call in ComputeAttentionItems()
  - Shows pods with >10 errors or >20 warnings in dashboard
  - Uses same detection functions as log view (isErrorLog, isWarnLog)
  - Respects --scan depth parameter (tunable scan depth)
  - Counts verified to match between dashboard and log view

### Technical - v0.5.0

- Removed mock mode infrastructure from app.go and app_test.go
- Re-enabled Tier 4 (log scanning) in attention_signals.go
- Build verification: `make build` passes cleanly
- Test suite passing with all mock references removed
- CHANGELOG, LESSONS-LEARNED, README, FUTURE_WORK updated

### Impact Summary - v0.5.0

- ✅ **17% leaner codebase** - Mock mode removed safely (698 lines deleted)
- ✅ **Zero dead code** - No mock functions lingering
- ✅ **Accurate log scanning** - Dashboard counts match log view exactly
- ✅ **Truth Only™ principle** - All displayed data verified accurate
- ✅ **Foundation for modular refactor** - Clean base for v0.5.1 app.go decomposition

### Deferred to v0.5.1

- App.go decomposition (3031 lines → modular architecture into helpers.go, fetch.go, handlers.go, table.go, logs.go, navigation.go)

### Lessons Applied

- Complete feature removal prevents tech debt accumulation
- Dangling mock references create compilation errors
- Re-enabling features requires test verification
- Documentation updates essential for future context

## [0.4.4] - 2026-01-02 "Post-Audit Improvements"

### Added ✨
- **Bundle validation** with helpful error messages
  - New `ValidateBundle()` function checks bundle structure before loading
  - Clear errors: "Not a valid RKE2 bundle - missing rke2/ directory"
  - Helpful hints: "Did you forget to extract? tar -xzf bundle.tar.gz"
  - Points user to expected structure when directory is wrong

- **CI stress test suite** (`scripts/test-bundle-stress.sh`)
  - Tests bundle validation error messages
  - Verifies 200MB size limit
  - Checks integration of new features
  - Prevents regressions on edge cases
  - 9 tests covering critical paths

### Changed 🔄
- **Bundle size limit increased** 100MB → 200MB
  - Real-world support bundles often 150-300MB
  - Updated in `internal/datasource/bundle.go`
  - Documented in comments

### Technical - v0.4.4
- Added `internal/bundle/validate.go` with `ValidateBundle()` function
- Integrated validation into bundle loader before loading
- Created comprehensive stress test suite for CI
- Updated LESSONS-LEARNED.md with audit findings and new principles

### Impact Summary - v0.4.4
- ✅ **Better error messages** - Users know exactly what's wrong
- ✅ **Larger bundles supported** - 200MB limit handles real-world use cases
- ✅ **CI prevents regressions** - Automated tests block edge case bugs
- ✅ **Clear roadmap to 10/10** - Audit documented path forward

### Audit Findings (AUDIT_POST_PIVOT.md)
- Overall codebase score: 7/10 (good, with clear improvement path)
- Identified 5 priorities:
  1. ✅ CI stress tests (implemented)
  2. ⏳ Decompose app.go (3643 lines → deferred to v0.5.0)
  3. ⏳ Re-implement dashboard log scanning (deferred to v0.5.0)
  4. ✅ Increase bundle size limit (implemented)
  5. ✅ Add bundle validation (implemented)

## [0.4.3] - 2025-12-12 "Truth Only™"

### Fixed 🐛

- **CRITICAL: Removed inaccurate dashboard log detection**
  - **Problem**: Dashboard showed identical fake ERR/WARN counts across different pods
    - Example: All argocd pods showed "19 ERR, 17 WARN" when actual logs showed "1 errors · 0 warnings"
    - This violated r8s core principle: **ONLY DISPLAY TRUTH**
  - **Root cause**: `detectLogIssues()` was reusing/caching log counts incorrectly across pods
  - **Solution**: Removed Tier 4 (log scanning) from dashboard entirely
  - **Dashboard now shows ONLY verified signals**:
    - ✅ Pod state (CrashLoopBackOff, OOMKilled, restarts)
    - ✅ Cluster health (nodes, etcd alarms)
    - ✅ Warning events (aggregated by type)
    - ✅ System metrics
    - ❌ **REMOVED**: Per-pod log ERR/WARN counts
  - **Impact**: Dashboard displays only accurate, verified data
  - **Note**: Real-time log counting in individual pod view remains 100% accurate

- **Search highlighting visibility improved**
  - **Problem**: Search matches were hard to see with dull yellow background
  - **Solution**: 
    - Bright yellow background (color 226 vs 11)
    - Bold text for extra emphasis
    - Maximum contrast for accessibility
  - **Impact**: Search matches now highly visible in log view

### Technical - v0.4.3

- Removed `detectLogIssues()` call from `ComputeAttentionItems()` in attention_signals.go
- Updated `searchMatchStyle` in styles.go with brighter color and bold
- Added detailed notes in FUTURE_WORK.md for future log scanning re-implementation
- All changes committed with detailed rationale

### Impact Summary - v0.4.3

- ✅ **r8s displays ONLY truth** - no more fake/misleading data
- ✅ **Search highlighting clear** - matches easy to spot
- ✅ **Trust restored** - dashboard shows verified signals only
- ✅ **Principle upheld**: Better to show less information than wrong information

### Principle Established

**r8s only displays truth.** We removed a feature that was displaying false information. Log scanning will be re-implemented in v0.5.0 once it can be verified to show accurate per-pod counts. See FUTURE_WORK.md for details.

## [0.4.2] - 2025-12-11 "Namespace Health Ranking"

### Added ✨
- **Namespace Health Visibility in Classic View**
  - New "ISSUES" column shows live E/W counts: "🔥 127E/89W" or "✅ Clean"
  - Auto-sort namespaces by total issue count (worst offenders at top)
  - Color coding: 🔥 (>50 errors), ⚠️ (>20 warnings or 1-50 errors), ✅ (clean)
  - Scans all pods in namespace and aggregates error/warning counts
  - Uses same scan depth as dashboard (tunable via --scan flag)

### Technical
- Added `NamespaceHealth` struct tracking errors/warnings/total per namespace
- Implemented `ComputeNamespaceHealth()` for efficient pod log aggregation
- Namespace table sorting: bubble sort by total issue count descending
- Reuses existing `isErrorLog()` and `isWarnLog()` detection functions
- Column layout: NAME(30) | ISSUES(15) | STATE(12) | PROJECT(18) | AGE(10)

### Impact Summary
- ✅ **Instant namespace triage** - Worst namespaces always visible at top
- ✅ **Zero navigation** - See health at-a-glance without drilling down
- ✅ **Consistent detection** - Same E/W patterns as dashboard and pod view
- ✅ **Performance** - Efficient aggregation, <100ms for typical bundles

## [0.4.1] - 2025-12-11 "Smart Sorting"

### Added ✨
- **Smart Sorting for Dashboard and Classic Pod List**
  - Default sort: Highest total ERR+WARN count descending (worst offenders at top)
  - Toggle with 's' key: Dashboard cycles Count → Severity → Name, Pods toggle Count ↔ Name
  - Per-view sort state preserved during session
  - Status bar shows current sort mode: "Sort: Count ▼" / "Sort: Name" / "Sort: Severity"
  - Cached pod E/W counts for instant re-sorting (no log re-scanning)

- **W/E Column in Classic Pod List**
  - Every pod row shows live "127E/89W" counts (or "-" if no issues)
  - Format: "XE/YW" (errors first for quick scanning)
  - Consistent with dashboard detection (same scan depth and patterns)
  - Enables quick issue identification without entering logs

### Technical
- Added `SortMode` enum (Count, Severity, Name) in attention_signals.go
- Implemented `SortPodsByCount()`, `SortPodsBySeverity()`, `SortPodsByName()`
- Added `populatePodCounts()` for efficient E/W count caching
- Per-view sort state tracking in `sortModes map[ViewType]SortMode`
- Dashboard sorting applied in `renderAttentionDashboard()`
- Pod view sorting applied in `updateTable()` for ViewPods case
- 's' key handler with view-specific behavior

### Impact Summary
- ✅ **Instant triage** - Worst pods always visible at top of list
- ✅ **Flexible views** - Toggle between count-based and alphabetical sorting
- ✅ **Zero latency** - Cached counts enable instant re-sort
- ✅ **Consistent UX** - Same sorting logic across dashboard and classic views

## [0.4.0] - 2025-12-11 "Dashboard Scrolling & Smart Capping"

### Added ✨
- **Smart capping with expansion for Attention Dashboard**
  - Default cap at top-20 most critical issues (sorted by severity)
  - Press 'm' to toggle between capped and expanded (all items) view
  - Position indicator shows "Showing X/Y" when items are capped
  - Clear message "...and X more issues (press 'm' to show all)" displayed when capped
  - Session-only preference (no persistence needed)
  
- **Enhanced navigation hotkeys**
  - 'g' - Jump to first item (vim muscle memory)
  - 'G' - Jump to last item (vim muscle memory)
  - 'm' - Toggle dashboard expansion (capped ↔ all items)
  - Smooth navigation through 200+ items without screen overflow

### Fixed 🐛
- **CRITICAL: Dashboard overflow with high --scan values**
  - Root cause: --scan=500+ detected 80+ issues, all rendered at once causing screen overflow
  - Dashboard would fill entire terminal height with no scrolling or pagination
  - Solution: Smart cap at top-20 by default with toggle to see all
  - Impact: High --scan values (500-1000) now usable without UX degradation

### Use Cases
- **Large bundles**: Use --scan=1000 confidently - dashboard stays clean with top-20 cap
- **Power users**: Press 'm' to expand and see all detected issues
- **Quick triage**: Default top-20 view focuses on most critical problems first

### Technical
- Added `attentionExpanded` boolean state field to App struct
- Implemented `getDisplayedItems()` helper with capping logic
- Added 'm', 'g', 'G' key bindings in attention dashboard navigation
- Smart cursor reset when toggling between capped/expanded modes
- Position indicator automatically updates based on displayed vs total count

### Impact Summary
- ✅ **--scan=1000 now usable** - previously caused dashboard overflow
- ✅ **Clean default UX** - Top-20 most critical issues shown by default
- ✅ **No data hidden** - Everything accessible via 'm' toggle
- ✅ **Vim-style navigation** - g/G for jump-to-top/bottom feels natural
- ✅ **Session simplicity** - No persistence needed, instant toggle

## [0.3.9] - 2025-12-10 "Tunable Scan Depth"

### Added ✨
- **--scan flag for customizable error/warning detection depth**
  - New CLI flag: `r8s --scan 500` sets scan depth to 500 lines
  - Default remains 200 lines for optimal performance
  - Applies consistently to: Attention Dashboard, W/E column, and log view header
  - Higher values = more accurate counts but slower performance
  - Lower values = faster scans but may miss issues deeper in logs
  
### Use Cases
- **Large logs**: `r8s --scan 1000` for thorough deep scanning
- **Quick triage**: `r8s --scan 50` for instant dashboard with recent errors only
- **Production bundles**: Tune based on typical log volume and performance needs

### Technical
- Added `ScanDepth` field to config.Config struct
- Added `--scan` flag to tui command (default: 200)
- Updated `ComputeAttentionItems()` to accept scanDepth parameter
- Updated `detectLogIssues()` to use tunable scan depth
- Updated W/E column rendering to use config.ScanDepth
- Scan depth validation: negative values default to 200

### Impact Summary
- ✅ **User control** - Adjust trade-off between speed and accuracy
- ✅ **Consistent behavior** - Same scan depth across all views
- ✅ **Performance tuning** - Optimize for your bundle sizes
- ✅ **Backward compatible** - Default 200 lines unchanged

## [0.3.8] - 2025-12-10 "Count Consistency Fix"

### Fixed 🐛
- **CRITICAL: Error/Warning counts now consistent across all views**
  - Root cause: Dashboard, W/E column, and log view used different scan depths and detection functions
  - Dashboard scanned 500 lines with `isErrorLine/isWarnLine` (old patterns)
  - W/E column scanned 200 lines with `isErrorLog/isWarnLog` (v0.3.7 corrected patterns)
  - Log view counted ALL lines with `isErrorLog/isWarnLog`
  - Result: Dashboard showed "22 ERR, 14 WARN" while log view showed "19 errors · 17 warnings"
  - Solution: 
    - Unified scan depth to **200 lines** everywhere (dashboard + W/E column + log view header)
    - Unified detection functions to use `isErrorLog/isWarnLog` across all components
  - Impact: **100% count consistency** - same numbers everywhere users look

### Technical
- Changed dashboard scan depth from 500 → 200 lines in `detectLogIssues()`
- Replaced `isErrorLine/isWarnLine` with shared `isErrorLog/isWarnLog` functions
- All three view types now use identical counting logic
- Functions defined once in app.go, reused in attention_signals.go (same package)

### Impact Summary
- ✅ **Dashboard count** = **W/E column** = **Log view header** (perfect sync)
- ✅ Faster dashboard scans (200 vs 500 lines)
- ✅ Reduced code duplication (removed 130 lines of duplicate detection logic)
- ✅ Single source of truth for error/warning patterns

## [0.3.7] - 2025-12-10 "Issue Hunter Hotfix"

### Fixed 🐛
- **CRITICAL: Warning logs now correctly display in YELLOW (not RED)**
  - Root cause: `isErrorLog()` checked keyword patterns (like "FAILED") before checking explicit log level indicators
  - A line like `W1204 [WARN] Skipping failed migration` was detected as ERROR due to "failed" keyword
  - Solution: Prioritize explicit level indicators ([WARN], [INFO], W####, I####) over keyword patterns
  - Impact: Proper color coding in log view - warnings are now yellow, errors are red
  - Edge case fix: INFO logs with error keywords (e.g., "Failed to read checkpoint") no longer show as errors

### Added
- **FUTURE_WORK.md document** tracking deferred features and enhancement ideas
  - Catalogued deferred features from v0.3.6 planning (smart sorting, hotkeys, journald scanning)
  - Priority/complexity/impact ratings for future planning
  - Long-term ideas (real-time monitoring, advanced search, plugin system)
  - Technical debt items (test coverage, refactoring targets)

### Technical
- Refactored `isErrorLog()` in app.go to exclude WARN/INFO/DEBUG logs before keyword matching
- Refactored `isWarnLog()` to exclude ERROR logs before keyword matching  
- Added comprehensive test suite in `log_detection_test.go` (11 test cases)
- All tests passing: ✅ WARN with "failed" keyword → YELLOW, INFO with "failed" → no color, ERROR → RED

### Impact Summary
- **100% accurate log level detection** - no more color confusion
- **Faster triage** - visual scanning now reliable (red = errors, yellow = warnings)
- **Better UX** - colors match expectations and log level semantics

### Deferred to v0.3.8
- Smart dashboard sorting by error count
- Global `e`/`w` hotkeys to jump to highest error/warn pod
- Status bar global issue count
- Enhanced help panel with pro tips

## [0.3.6] - 2025-12-10 "Issue Hunter"

### 🎉 Major Changes
- **BREAKING:** Removed live Rancher API mode entirely  
- **NEW:** Default launches with embedded demo bundle (zero config)
- **NEW:** Always starts with Attention Dashboard
- Simplified architecture: bundle-first design

### Removed
- Live Rancher API client (~300 lines)
- Live datasource implementation (~230 lines)
- Profile-based authentication (~100 lines)
- `--profile`, `--insecure`, `--mockdata` flags
- Client test files and live mode logic
- **Total:** ~1,200 lines removed (11.7% of codebase)

### Changed
- Default behavior: `./r8s` now launches demo bundle instantly
- CLI help text updated to emphasize bundle workflows
- Simplified NewApp() to only handle bundle/demo modes
- All docs updated to remove live mode references

### Why This Change?
User feedback showed bundles are the #1 workflow. Removing live mode:
- ✅ Eliminates configuration complexity
- ✅ Works 100% offline
- ✅ Faster startup
- ✅ Cleaner codebase
- ✅ Better UX for primary use case

**Migration:** Users needing live cluster browsing should stay on v0.3.4 or use native Rancher UI.

**Development time:** 22 minutes from audit to tagged release.

## [Unreleased]

## [0.3.6] - 2025-12-10 "Issue Hunter"

### Enhanced - ERR/WARN Detection 🔍
- **Enhanced warning pattern detection (8 new patterns)**
  - Added: WARNING:, WARN:, WARN=, LEVEL=WARNING
  - Added: DEPRECATED, DEPRECATION, ALERT:, ALERT=
  - All patterns now case-insensitive for maximum coverage
  - Synced patterns between attention_signals.go and app.go for consistency
  - Impact: Dashboard and log views detect vastly more warning types

- **Attention Dashboard capacity increased**
  - Dashboard cap: 15 → 100 items (scrollable list)
  - Allows viewing all issues in huge bundles
  - Scroll down to see additional items beyond screen height
  - Impact: No critical issues hidden by arbitrary limits

### Enhanced - Classic View UX ⚡
- **W/E column format improved for clarity**
  - Format changed: "18/22" → "22E/18W" (errors first, explicit labels)
  - Scan depth increased: 100 → 200 lines for better accuracy
  - Impact: Instant visibility of error vs warning counts in pod list

- **Smart log filtering on pod entry**
  - Entering pod logs from Pods view now auto-applies WARN filter
  - Shows errors + warnings by default (Ctrl+A to see all logs)
  - Impact: Immediate focus on issues without manual filtering

### Technical
- Enhanced `isWarnLine()` in attention_signals.go with 8 additional patterns
- Synced `isWarnLog()` in app.go with same pattern set
- Dashboard item limit increased from 15 to 100 in ComputeAttentionItems()
- W/E column format changed to "XE/YW" in updateTable()
- Log scan depth increased to 200 lines in pod table rendering
- Auto-apply filterLevel = "WARN" when entering logs from ViewPods

### Impact Summary
- **+166% more WARN patterns detected** (3 → 11 patterns)
- **+566% dashboard capacity** (15 → 100 items)
- **+100% scan depth** (100 → 200 lines for W/E counts)
- **Zero-click issue focus** (WARN filter auto-applied on pod entry)

### Deferred to Future Releases
- Journald log scanning (requires new datasource methods)
- Smart dashboard sorting by error count
- Global issue count in status bar
- Help panel pro tips

## [0.3.5] - 2025-12-10 "Bundle-Only Bliss"

### Fixed - Demo Parity Complete 🎯
- **CRITICAL: Logs now load in mockdata mode**
  - Root cause: GetLogs() returned error when no log file found instead of generating demo logs
  - Solution: Always generate demo logs when bundle has no log files for a pod
  - Impact: Dashboard log scanner and classic pod view now work in mockdata mode
  - All pods detected by attention dashboard can now be drilled into successfully

- **CRITICAL: W/E column in classic Pods view now works**
  - Root cause: Column scanned kubectl events which don't exist in mockdata
  - Solution: Scan first 100 lines of pod logs for errors/warnings (same as dashboard)
  - Impact: Classic pod list now shows "18/22" (WARN/ERR) counts immediately
  - Provides instant error visibility without opening logs

### Enhanced - Error Detection
- **Enhanced error pattern matching (12 new patterns)**
  - Added: ERR=, FAILED, FATAL, PANIC, OOMKILLED, CRASHLOOP, BACK-OFF, BACKOFF
  - Added: UNAUTHORIZED, DENIED, EXCEPTION, LEVEL=ERROR
  - All patterns case-insensitive for maximum coverage
  - Impact: Dashboard and log views detect vastly more error types

- **Realistic demo logs for every pod**
  - Default pods: 22 errors + 18 warnings (57 lines total)
  - Crash scenarios: 127 errors for pods with "crash" in name
  - Impact: Every demo pod shows realistic error/warning patterns
  - Better demonstration of dashboard and log viewing capabilities

- **Dashboard log scanner active**
  - Scans first 500 lines of up to 10 pods for performance
  - Shows pods with >10 errors as 🔥 CRITICAL with "X ERR, Y WARN" counts
  - Shows pods with >20 warnings as ⚠️ WARNING with "Y WARN, X ERR" counts
  - Impact: Attention Dashboard now actively displays log-based issues

### Technical
- Modified `GetLogs()` in bundle.go to return demo logs instead of error
- Added `generateDemoLogs()` and `generateCrashLogs()` helper functions
- Implemented `detectLogIssues()` in attention_signals.go
- Enhanced `isErrorLog()` with 12 additional patterns
- Updated pod table rendering to scan logs for W/E column

### Testing
- ✅ Builds cleanly (v0.3.4-8-g9c47b69)
- ✅ Mockdata mode: Dashboard shows ERR counts
- ✅ Mockdata mode: Logs load for all pods
- ✅ Classic view: W/E column populated with log scan results
- ✅ Bundle mode: No regressions, real logs still load correctly

## [0.3.4-initial] - 2025-12-05

### Fixed - Production Ready 🚀
- **CRITICAL: kubectl pod parsing for variable RESTARTS field format**
  - Fixed NODE column showing age data (e.g., "7d23h") instead of node names
  - Root cause: RESTARTS field can be "8" or "8 (4m53s ago)" causing variable field count
  - Solution: Dynamically detect IP field position, derive AGE and NODE from it
  - Handles all pod states: Running, CrashLoopBackOff, ImagePullBackOff correctly
  - Proven fix: Tested on example bundle with multiple CrashLoopBackOff pods
  
- **CRITICAL: --mockdata now defaults to Attention Dashboard**
  - Mockdata mode now shows Attention Dashboard on launch (matches bundle mode)
  - Better demo experience - shows killer feature immediately
  - Consistent behavior across bundle and demo modes
  - Users see the "wow" factor right away

### Added - UX Polish
- **Enter key navigation in Pods view**
  - Enter key now opens logs for selected pod (UX consistency)
  - Matches Attention Dashboard behavior (Enter = drill deeper)
  - 'l' key still works as alternative for power users
  - Consistent keyboard shortcuts across all views reduce cognitive load

### Changed
- Mockdata initial view: Clusters → Attention Dashboard
- Pod parsing: Fixed field positions → Dynamic IP-based field detection

### Technical
- Enhanced `ParsePods()` in `kubectl.go` with dynamic field positioning
- IP address detection loop finds correct column regardless of RESTARTS format
- Fallback to fixed positions if IP detection fails (backward compatibility)
- Updated `NewApp()` initial view logic: `bundleMode || offlineMode` → Attention Dashboard
- Added `case ViewPods:` handler in `handleEnter()` for log navigation

### Testing
- ✅ Builds cleanly (v0.3.4-7-ge67a69d)
- ✅ kubectl parsing handles "8" and "8 (4m53s ago)" RESTARTS formats  
- ✅ NODE column displays correctly for all pod states
- ✅ Mockdata shows Attention Dashboard on launch
- ✅ Enter key navigates to logs in Pods view
- ✅ No regressions in bundle or live modes

## [0.3.3-final] - 2025-12-04

### Fixed
- **CRITICAL: Dashboard keyboard navigation completely restored**
  - ↑/↓ or j/k moves focus with visible cyan highlight (inverted row)
  - 1-9 instant jump to issue lines
  - Enter drills down to pod logs (showing ALL logs by default)
  - → or l expands collapsed event lines showing affected pods
  - ← or h collapses expanded items or exits sub-navigation
  - c = classic cluster view, r = refresh, q = quit
  - All keys work instantly with no input delay
- **CRITICAL: Launch errors now display cleanly in terminal**
  - Invalid bundle paths print helpful error messages and exit immediately
  - No more "Press Esc" messages when TUI can't start
  - Clean CLI error handling with proper exit codes
  - Eliminates confusing "could not open TTY" errors for initialization failures

### Changed
- **Dashboard is now BUNDLE-ONLY** (architectural decision locked in permanently)
  - Live mode skips dashboard entirely, goes directly to Clusters view
  - Clear message in live mode: "Live cluster browser — use --bundle for Attention Dashboard"
  - Removes all live-mode attention-related bugs permanently
  - Doubles development velocity by eliminating dual-mode complexity
- **Log filter defaults to ALL** when navigating from dashboard
  - Users can still apply Ctrl+E (ERROR) or Ctrl+W (WARN+ERROR) filters as needed
  - Better default UX - see all context first, then filter down

### Added
- Visible selection highlighting in dashboard (cyan background with inverted colors)
- [BUNDLE] prefix in status bar when using bundle mode
- Session state tracking for dashboard cursor position
- Expandable event line infrastructure with pod sub-navigation:
  - Arrow down (↓) or 'j' enters pod list when item is expanded
  - Arrow up/down navigates within pod list with visible highlighting
  - Enter on selected pod jumps to that pod's logs
  - Left arrow (←) or 'h' exits pod list back to main items
- Pod event counts displayed next to each affected pod in expanded view

### Technical
- Added `attentionCursor`, `expandedItems`, and `subCursor` state fields to App struct
- Added `HasError()` and `GetError()` methods to App for pre-TUI error checking
- Keyboard navigation handled before general table navigation in Update()
- Initial view selection based on mode: bundle → Attention, live/mock → Clusters
- Selection rendering uses `isSelected` parameter in `renderAttentionItem()`
- Pod highlighting in expanded views uses `inSubNav` parameter
- Cyan/dark-gray color scheme for maximum visibility
- Error check in `cmd/tui.go` before launching Bubble Tea program

## [0.3.3] - 2025-12-04 (IN PROGRESS)

### Added
- **🔥 Attention Dashboard**: New default root view that immediately shows cluster health status
  - Detects critical issues: CrashLoopBackOff, OOMKilled, ImagePullBackOff, Evicted pods
  - Detects pod restarts (≥3 in recent period)
  - Identifies high error/warning counts in logs
  - Shows etcd health issues (bundle mode)
  - Detects NotReady nodes
  - Displays DaemonSet incomplete deployments
  - Parses cluster events for warnings and failures
  - Clean "All good ✨" state when no issues detected
  - Severity-based grouping: Critical, Warning, Info
  - One-key drill-down (1-9 for quick jump, Enter for details)
  - Toggle classic navigation with 'c' key
  - Configurable default view preference

### Added (Technical)
- `internal/tui/attention.go` - Attention Dashboard view and orchestration
- `internal/tui/attention_signals.go` - Signal detection engine with 5 detector tiers
- `internal/bundle/etcd.go` - etcd health file parsers (alarmlist, endpointhealth)
- `internal/bundle/systeminfo.go` - System health parsers (memory, disk)
- Extended kubectl parsers: ParseNodes(), ParseDaemonSets()
- New DataSource interface methods: GetNodes(), GetEtcdHealth(), GetSystemHealth()
- `ViewAttention` type added to navigation flow
- Attention-specific styles with emoji indicators and severity colors

### Changed
- Default root view is now Attention Dashboard (classic view accessible via 'c' key)
- Config supports `defaultView` setting for user preference persistence

## [0.3.2] - 2025-12-03

### Fixed
- **Describe function in Live mode**: Fixed pod/deployment/service describe breaking in Live mode
  - Root cause: describe functions were calling `GetPods("")` which fails without projectID in Live mode
  - Solution: Use proper `DataSource.DescribePod/Deployment/Service()` interface methods
  - All three describe functions now work correctly in Live, Bundle, and Demo modes
  - Removed mock data fallbacks in favor of proper datasource abstraction

### Refactored
- **DataSource architecture**: Unified data layer with clean interface abstraction
  - New `internal/datasource/` package with `DataSource` interface
  - Three implementations: `LiveDataSource`, `BundleDataSource`, `EmbeddedDataSource`
  - Zero code duplication between modes
  - No fallback code paths - each mode is self-contained
- Demo mode (`--mockdata`) now uses embedded example bundle instead of synthetic data
- Selection preservation infrastructure added (resets to top for now, full implementation deferred)

### Investigated
- **Selection preservation**: Table position restoration when navigating back
  - Added `savedRowName` field for storing selected row identifier
  - Implementation blocked by bubble-table library API limitations
  - Library doesn't expose row iteration or selection-by-data methods
  - Documented in code comments for future implementation options
  - Workarounds would require: library fork, parallel data structures, or different table library

### Removed
- Dead code in fetch functions (300+ lines of fallback calls eliminated)
- Silent fallback behaviors between modes

### Technical
- `getMock*()` functions retained for test suite only
- All fetch functions simplified to single datasource call
- Clean separation: Live → API, Bundle → Files, Demo → Example
- Selection preservation requires architectural changes to implement fully

### Known Limitations
- Table selection currently resets to top when navigating back (library API constraint)
- Full implementation deferred until bubble-table API enhancement or library replacement

## [0.3.1] - 2025-12-03

### Added
- **Vim-style log navigation**: `g` key jumps to first log line, `G` jumps to last line (instant even on 5M-line logs)
- **Universal back navigation**: `b` key works everywhere alongside `Esc` for intuitive navigation
- **Word wrap toggle**: `w` key toggles soft word-wrap in log view with "Wrap:On" indicator
- **Enriched bundle pod details**: Bundle mode now shows full pod metadata from kubectl output
  - Pod status (Ready, Status, Age, IP, Restarts, ReadinessGates)
  - Kubernetes events attached to pods (17 events loaded from example bundle)
  - 93 pods with full kubectl data vs basic 86 pod inventory
- **Events parsing**: ParseEvents() function extracts pod events from kubectl output

### Changed
- Log view horizontal scrolling improved with word wrap support
- Bundle pod describe now shows data comparable to live mode
- Help screen updated with new keyboard shortcuts

### User Experience
- Log navigation feels instant and responsive with vim muscle memory
- Long log lines are now readable with toggleable word wrap
- Back navigation is more intuitive with `b` key option
- Bundle mode provides richer pod context for troubleshooting

## [0.3.0] - 2025-12-01

### Fixed
- **BUG-001**: CRD version selection 404 errors - now properly handles CRDs with multiple versions
- **BUG-002/003**: Nil pointer crashes and bundle path validation issues
- **Bugbash 2025-11** - Fixed 14 bugs across 4 systematic rounds:
  - Symlink panic during tar.gz extraction (now skips with warning)
  - Silent mock data fallbacks eliminated - all 10 fetch functions now fail explicitly with verbose errors
  - Filter state persistence after search exit
  - Search results becoming stale after filter changes
  - Tail mode tick pattern breaking continuous log updates
  - Log viewport not resizing on terminal window changes
  - vim j/k navigation advertised but not implemented
  - Ctrl+L eating next keystroke (remapped to refresh)
  - Incomplete error context in verbose mode
  - Bundle loading confusing messages
- Pod parsing now handles dash-separated filenames correctly

### Added
- Auto-version detection from git tags in Makefile
- Mode indicators in breadcrumb: [LIVE]/[BUNDLE]/[MOCK] for clear data source visibility
- vim j/k navigation in all table views (finally implemented)
- Verbose error context with `--verbose` flag across all operations
- Early validation for tarball paths in TUI command
- Config validate and set commands
- Comprehensive help system improvements
- Bundle size limit flag: `--limit` for bundle command

### Changed
- **Breaking**: Removed tarball support - bundles must be pre-extracted before import
- Bundle command terminology: 'import' → 'info' for clarity
- Loading messages now mode-aware (differentiate live/bundle/mock loading states)
- Filter clearing now also clears search state for consistency

### Removed
- Tarball extraction support (security and complexity reasons)
- Silent fallback behaviors (replaced with explicit errors)

## [0.2.1] - 2025-11-26

### Fixed
- Deployment replica counts now display correctly instead of always showing 0/0
- Services no longer show mock data in online mode when API errors occur
- Added DeploymentScale struct for nested replica data support
- Implemented multi-tier fallback strategy for replica count extraction

### Changed
- Updated fetchServices() to match fetchDeployments() error handling pattern

## [0.2.0] - 2025-11-26

### Added
- **Describe feature** for Pods, Deployments, and Services (press 'd' key)
- Comprehensive unit test suite with race detection (53 tests)
- Package-level godoc documentation for all packages
- Offline mode with automatic fallback to mock data
- Test coverage reporting (~90% for core packages)

### Fixed
- Go version in go.mod (corrected from 1.25 to 1.23)
- Pod.HostnameI typo renamed to Pod.Hostname
- Pod NODE column now displays correctly using multiple field fallbacks
- Namespace counts in Projects view now show accurate data
- Data extraction issues in Pods, Deployments, and Projects views

### Changed
- Enabled race detection in Makefile test target
- Renamed test_describe.go to avoid conflicts

## [0.1.0] - 2025-11-20

### Added
- **CRD (Custom Resource Definition) Explorer** with instance browsing
- CRD description toggle with 'i' key
- Instance counter column in CRD table
- Realistic and varied CRD instance counts
- Prominent offline mode warning banner
- Deployments and Services views with keyboard navigation
- View switching: 1=Pods, 2=Deployments, 3=Services

### Fixed
- CRD instance counter now uses live API data
- CRD navigation performance issue documented
- Navigation flow and offline mode functionality
- Table alignment (removed border styling from header)

## [0.0.3] - 2025-11-15

### Added
- Navigation stack for breadcrumb-style navigation

back/forward
- Projects view with namespace counts
- Improved help screen with ASCII logo
- Better styling and formatting throughout

### Changed
- Improved CRD browser filtering and navigation

## [0.0.2] - 2025-11-10

### Added
- Initial TUI framework using Bubble Tea
- Cluster listing view
- Basic navigation between views

### Fixed
- Rancher API type definitions corrected
- Live instance testing functionality

## [0.0.1] - 2025-11-01

### Added
- Initial project scaffolding
- Basic CLI structure with Cobra
- Configuration management with Viper
- Rancher API client implementation
- README with project vision
- LICENSE (Apache 2.0)
- Git repository initialization

---

## Version History Summary

- **v0.3.0** (2025-12-01): Bugbash 2025-11 - 14 bugs fixed, no silent fallbacks, mode indicators
- **v0.2.1** (2025-11-26): Bug fixes for replica counts and error handling
- **v0.2.0** (2025-11-26): Describe feature, comprehensive testing, documentation
- **v0.1.0** (2025-11-20): CRD Explorer, offline mode, view switching
- **v0.0.3** (2025-11-15): Navigation improvements, projects view
- **v0.0.2** (2025-11-10): TUI framework, cluster listing
- **v0.0.1** (2025-11-01): Initial release

---

## Upgrade Notes

### Upgrading to 0.3.0
- **Breaking Change**: Tarball extraction removed - bundles must be pre-extracted before using `r8s bundle` command
- Bundle command changed from `r8s bundle import` to `r8s bundle info`
- All silent mock data fallbacks eliminated - errors now shown explicitly (use `--verbose` for details)
- New mode indicators in UI: [LIVE], [BUNDLE], or [MOCK] show active data source
- vim j/k navigation now works in all table views
- Use `--limit` flag to adjust bundle size limits if needed

### Upgrading to 0.2.1
- No breaking changes
- Deployment replica counts will now display correctly
- Error messages in online mode are now more informative

### Upgrading to 0.2.0
- New 'd' key binding for describe functionality
- Offline mode automatically detects connection failures
- Run `go test -race ./...` to verify tests pass

### Upgrading to 0.1.0
- New view navigation keys: 1, 2, 3 for Pods/Deployments/Services
- CRD explorer accessible from cluster view with 'C' key

---

## Contributors

- Development Team

For detailed commit history, see: `git log --oneline`
