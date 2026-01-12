# r8s Future Work & Deferred Features

This document tracks feature ideas and enhancements that have been identified but deferred to future releases.

---

## 🎯 v0.5.x → v0.6.x Roadmap (January 2026)

**Foundation Status**: ✅ **COMPLETE** (v0.5.10, v0.5.11, v0.5.12 shipped 2026-01-08)

### Foundation Releases (✅ COMPLETE)

All backend parsers and data structures ready for v0.6.x consumption:

#### v0.5.10 "Data Source Enrichment" ✅ SHIPPED

- Enhanced ETCD parser (member count, leader ID, DB size, compaction recs)
- Node conditions parser (MemoryPressure, DiskPressure, resource capacity)
- DataSource interface extensions (GetEtcdDetails, GetNodeConditions)
- ~480 lines of backend capabilities, zero UI changes

#### v0.5.11 "Kubelet & OOM Analysis" ✅ SHIPPED

- Kubelet log parser (10+ error patterns from journald)
- OOM root cause analyzer (container vs node OOM)
- Container resource parser (CPU/memory limits, QoS classes)
- ~382 lines of detection capabilities, zero UI changes

#### v0.5.12 "Diagnostic Context Types" ✅ SHIPPED

- DiagnosticContext struct (RootCause, Recommendation, Severity, FixPriority)
- Context generators (CrashLoop, OOM, ImagePull, Node, ETCD, Kubelet)
- Recommendation engine (severity/priority mapping)
- ~168 lines of data structures, zero UI changes

**Total Foundation**: ~1,030 lines of backend code ready for consumption

---

### v0.6.x Implementation (7 Releases Planned)

**See**: `docs/V0.6.X_ROADMAP.md` for detailed prompts and implementation guide

**Quick Summary**:
1. **v0.6.0** - Diagnostic Panel Tightening (UI cleanup, show only failures)
2. **v0.6.1** - Cluster Events Drill-Down (fix v0.5.9 limitation)
3. **v0.6.2** - Full Container Detection (multi-container support)
4. **v0.6.3** - Enhanced ETCD Display (consume v0.5.10)
5. **v0.6.4** - Node Conditions Display (consume v0.5.10)
6. **v0.6.5** - Kubelet Issues Display (consume v0.5.11)
7. **v0.6.6** - OOM Root Cause Display (consume v0.5.11)
8. **v0.6.7** - Inline Diagnostics (consume v0.5.12, 2-line format)

**Timeline**: 9-12 days total for complete diagnostic-first overhaul

**Success Criteria**:
- "5-Second Rule" - operator decides action in ≤5 seconds
- Diagnostic panel: 6 sections → 4 sections
- Event messages ≤80 chars (truncated)
- Show ONLY failing containers/Warning events
- Dashboard: 2-line format with inline root cause + recommendation

**Philosophy**: "Show, Don't Ask" - automatic diagnostic display without user action

---

### Documentation & Next Steps

- **docs/V0.6.X_ROADMAP.md** - ✅ Detailed prompts for each v0.6.x release
- **docs/V0.6.0_PLAN.md** - Original technical specification
- **docs/V0.5.9_KNOWN_LIMITATIONS.md** - Known issues to be resolved
- **CHANGELOG.md** - Complete v0.5.x history

**Ready to Start**: v0.6.0 implementation (all prerequisites complete)

---

## ✅ Recently Completed

### Container Status Diagnostics (v0.5.6)
- **Status**: ✅ Shipped in v0.5.6 - 2026-01-06
- **Description**: Enhanced error diagnostics with container-level health visibility
- **Impact**: Users see which specific containers failed in multi-container pods
- **Features**:
  - Container health indicators (✅ running, ❌ failed, ⚠️ unknown)
  - BackOff frequency analysis from events
  - Failure message extraction from Warning events
  - Pod Ready state correlation for intelligent inference
  - Clean, value-only output (removed "exit codes not captured" noise)
- **Philosophy**: "Maximum Information Extraction" - show all available data transparently
- **Test Case**: Pod `test-notready` (1/2 ready) now shows ok-container vs failing-container

### Dashboard Scrolling & Item Capping (v0.4.0)
- **Status**: ✅ Shipped in v0.4.0
- **Description**: Smart capping with expansion for high --scan values
- **Impact**: Dashboard now handles 80+ issues gracefully without screen overflow
- **Features**:
  - Default cap at top-20 issues (sorted by severity)
  - Press 'm' to toggle between capped/expanded view
  - Press 'g/G' to jump to top/bottom of list
  - Position indicator shows "Showing X/Y" when capped
  - "...and X more (press 'm')" message when items hidden
  - Smooth navigation through unlimited items
- **Usage**: Works automatically with `--scan` flag at any value
- **Result**: High --scan values (500-1000) now usable without UX degradation

### Tunable Scan Depth (v0.3.9)
- **Status**: ✅ Shipped in v0.3.9
- **Description**: User-controllable scan depth via `--scan` flag
- **Impact**: Users can now tune speed/accuracy trade-off based on bundle size
- **Usage**: `r8s --scan 500 ./bundle/` (default: 200 lines)

### App.go Modular Decomposition (v0.5.1)

- **Status**: ✅ Shipped in v0.5.1
- **Description**: Complete decomposition of 3031-line app.go into 6 focused modules
- **Impact**: 87% reduction in main file size, 70% cognitive load reduction
- **Modules Created**:
  - **app.go**: ~400 lines - Core state & orchestration
  - **helpers.go**: ~300 lines - Utilities, breadcrumb, status text
  - **logs.go**: ~450 lines - Log view rendering & filtering
  - **fetch.go**: ~400 lines - Data fetching & describe operations
  - **table.go**: ~500 lines - Table rendering for all views
  - **handlers.go**: ~500 lines - Event handlers & navigation
- **Result**: Better maintainability, improved testability, faster onboarding

### Dashboard Log Scanning (RE-ENABLED in v0.5.0)
- **Priority**: ~~HIGH~~ COMPLETED ✅ v0.5.0
- **Status**: ✅ Re-enabled in v0.5.0
- **Complexity**: Medium-High
- **Impact**: High
- **Status**: ❌ REMOVED in v0.4.3 due to inaccurate counts
- **Description**: Re-implement accurate per-pod log error/warning detection for dashboard
- **Problem Identified**: 
  - Dashboard showed identical ERR/WARN counts across different pods (e.g., all argocd pods showing "19 ERR, 17 WARN")
  - Actual pod logs showed different counts (e.g., "1 errors · 0 warnings")
  - Root cause: `detectLogIssues()` was reusing/caching counts incorrectly
  - Also noticed in namespace view - counts appear inconsistent
- **Requirements for Re-implementation**:
  - Fix GetLogs() calls to ensure correct namespace/pod parameters
  - Verify no caching/reuse of counts across different pods
  - Add per-pod verification: dashboard count MUST match log view count
  - Test with namespace-level aggregation
  - Add debug mode to verify which pod's logs are being scanned
  - Only re-enable once 100% verified accurate
- **Current State**: Real-time log counting in individual pod view works perfectly - keep that
- **Principle**: r8s only displays truth. Better to show no count than a wrong count.
- **Triggered by**: User-reported critical bug in v0.4.2

### Smart Sorting by Error Count
- **Priority**: Medium
- **Complexity**: Medium
- **Impact**: Medium
- **Description**: Auto-sort Attention Dashboard items by error/warning count for faster triage
- **Requirements**:
  - Refactor dashboard rendering to support dynamic sorting
  - Add sort toggle (by severity vs. by count)
  - Persist sort preference in config


## 📋 Medium Priority (v0.4.0)

### Diagnostic Panel Event Message Truncation (v0.5.8+) ⭐
- **Priority**: Medium
- **Complexity**: Low
- **Impact**: Medium (readability)
- **Description**: Long event messages (e.g., registry pull errors) overwhelm the diagnostic panel display
- **Problem Observed (v0.5.7)**:
  - Event messages like "Failed to pull image registry.rancher.com/..." span multiple lines
  - Multiple similar events create wall of text
  - Hard to scan for key information
  - Diagnostic panel loses scannable structure
- **Proposed Fix**:
  ```
  Current:  Failed: lgsp1skbtd12001.gso.aexp.com failed to pull image "registry.rancher...
            agent:v2.8.5": rpc error: code = DeadlineExceeded desc = failed to resolve 
            reference "registry.rancher.com/rancher/rancher-agent:v2.8.5": failed to do
            request: Head "https://registry.rancher.com/v2/rancher/rancher-agent/manifests...
            
  Proposed: Failed: Image pull timeout - registry.rancher.com/rancher/rancher-agent:v2.8.5
            └─ rpc error: DeadlineExceeded (truncated)
  ```
- **Requirements**:
  - Truncate event messages to max 80-100 characters
  - Preserve most important details (event type, resource, error code)
  - Add "(truncated)" indicator when message is cut
  - Consider smart truncation (keep beginning + end, ellipsis in middle)
  - Group identical events with count: "Image pull failed (x6)"
- **Location**: `internal/tui/logs.go` - `buildEventsSection()` function
- **Alternative**: Add `[d]=details` key to show full event text in modal
- **Philosophy**: "Scannable summaries > verbose walls of text"
- **Triggered by**: User feedback on v0.5.7 testing (registry pull errors)

### Namespace Health Ranking & Smart Filtering
- **Priority**: Medium-High
- **Complexity**: Medium
- **Impact**: High
- **Description**: Rank and sort namespaces by problem severity for quick detection of most problematic areas
- **Problem**: With 25+ namespaces, no quick way to identify which ones have the most issues. All show as "active" with no health indicators.
- **Requirements**:
  - Add "ISSUES" column to namespace view showing error/warning counts
  - Sort namespaces by total issue count (highest problems first)
  - Color-code namespace rows: Red (>50 errors), Yellow (>20 warnings), Green (healthy)
  - Filter options: "Show only namespaces with issues" (press 'f')
  - Quick jump to most problematic namespace (press 'e' for highest errors)
  - Aggregate pod-level errors per namespace for ranking score
- **Use Case**: "Which namespace should I investigate first?"
- **Example**: 
  ```
  NAME                  STATE    ISSUES       PROJECT
  kube-system          active   🔥 127E/89W  bundle-project
  gpu-operator         active   ⚠️  22E/67W  bundle-project  
  longhorn-system      active   ✅  2E/5W    bundle-project
  calico-system        active   ✅  0E/0W    bundle-project
  ```
- **Triggered by**: User feedback viewing 25-namespace bundle

### Journald Log Scanning
- **Priority**: Medium
- **Complexity**: High
- **Impact**: Medium
- **Description**: Scan systemd/journald logs for node-level issues
- **Requirements**:
  - Extend DataSource interface to support journald parsing
  - Pattern matching for systemd service failures
  - Integration with Attention Dashboard
  - Bundle structure changes to include journald/
- **Dependencies**: Requires datasource interface refactor

### Enhanced Help Panel
- **Priority**: Low
- **Complexity**: Low
- **Impact**: Low
- **Description**: Add contextual pro tips to help panel
- **Requirements**:
  - Pro tip: "Start with dashboard for quick wins"
  - Pro tip: "Use Ctrl+W in logs to focus on issues"
  - Pro tip: "Watch W/E column in Pods view for quick health check"
  - Context-aware tips based on current view

### Edge Case Handling
- **Priority**: Low
- **Complexity**: Low
- **Impact**: Low
- **Description**: Better UX for edge cases
- **Requirements**:
  - Empty logs: Show "No E/W — check describe/events"
  - Huge bundles: Cap dashboard at top-N with "and X more..." indicator
  - Parse errors: Show count in bundle load warning

## 💡 UX Philosophy: "Show, Don't Ask" (v0.5.2+)

**Core Principle**: Information should surface automatically without requiring button presses.

### Auto-Display Health Indicators
- **Priority**: High
- **Complexity**: Low
- **Impact**: High
- **Description**: Show critical system health automatically in dashboard header
- **Requirements**:
  - Dashboard header shows "🔥 2 critical · ⚠ 5 warnings · ✅ 12 healthy" always visible
  - No button press needed - appears on load
  - Updates automatically when data refreshes
- **Philosophy**: User sees cluster health at a glance, no navigation required

### Auto-Show Parse Warnings
- **Priority**: Medium
- **Complexity**: Low
- **Impact**: Medium
- **Description**: Display bundle parse issues automatically in status bar
- **Requirements**:
  - Status bar shows "⚠ 3 parse issues" when bundle has parsing errors
  - Appears automatically on bundle load
  - No hotkey needed - information surfaces immediately
- **Philosophy**: Problems should announce themselves, not hide behind buttons

### Progressive Information Disclosure
- **Priority**: Medium
- **Complexity**: Medium
- **Impact**: High
- **Description**: Each view automatically shows most important info without drilling down
- **Requirements**:
  - Namespace view: Shows error counts automatically in ISSUES column
  - Pod view: Shows W/E counts in every row by default
  - Dashboard: Worst items sorted to top automatically
  - Log view: Empty state shows helpful next steps automatically
- **Philosophy**: Critical data visible by default, details available on demand

### Smart Defaults Over Configuration
- **Priority**: Medium
- **Complexity**: Low
- **Impact**: Medium
- **Description**: System chooses best defaults, reduces need for user configuration
- **Requirements**:
  - Default sort = count (worst first) - no need to press 's'
  - Default filter = all logs - user adds filters only when needed
  - Scan depth auto-adjusts based on bundle size
- **Philosophy**: Software should be intelligent enough to make good choices automatically

### Remove Sort Mode Complexity (v0.5.3+)
- **Priority**: High
- **Complexity**: Low (removal)
- **Impact**: High (simplification)
- **Description**: Eliminate sort mode toggles - always use smart default (highest error count first)
- **Current State**: 
  - 3 sort modes: Count, Severity, Name
  - Toggle hotkey 's' cycles through modes
  - Status bar shows current mode
  - ~200 lines of sorting code
- **Proposed Change**:
  - Remove `SortMode` enum entirely
  - Remove sort toggle functionality
  - Always sort by error count descending (worst first)
  - Remove status bar sort indicator
  - Delete unused sorting functions
- **Rationale**:
  - 95% of users want "worst first" (count-based)
  - Sorting options add cognitive load without value
  - Smart default eliminates need for configuration
  - Fewer features = better UX
- **Code Reduction**: ~200 lines removed
- **Philosophy**: "The best feature is no feature - smart defaults beat options"

### CrashLoopBackOff with No Logs Indicator (v0.5.3+) ⭐
- **Priority**: HIGH
- **Complexity**: Low
- **Impact**: High (UX consistency)
- **Description**: Pods showing as CrashLoopBackOff in dashboard but "✅ Clean" in namespace view is misleading
- **Problem**:
  - Dashboard correctly shows pod in CrashLoopBackOff
  - Namespace/Classic view shows "✅ Clean" because no log files found
  - User sees conflicting information: "Is this pod healthy or broken?"
- **Scope Expansion (v0.5.3+)**:
  - Applies to ALL pods with zero log lines, not just CrashLoopBackOff
  - Any pod state (Running, Pending, Error, CrashLoopBackOff, etc.) with 0 logs gets indicator
  - Consistent application of Truth Only™ principle: "no logs ≠ clean"
- **Proposed Fix**:
  - When pod has 0 log lines (regardless of pod state):
    - Show "⚠️ No Logs" instead of "✅ Clean"
    - Use warning emoji (not green tick) for any pod without log data
    - Add tooltip: "No logs captured - check describe/events for details"
- **Logic**:

  ```python
  if logCount == 0:
    display "⚠️ No Logs"  # Not "✅ Clean"
  else:
    display actual E/W counts or "✅ Clean"
  ```

- **Philosophy**: Truth in context - "Clean" means healthy AND no issues, not just "no logs found"
- **Triggered by**: User-reported UX confusion
- **Applies to**: All pod states, not just CrashLoopBackOff (consistency)

### Empty Namespace Intelligence (v0.5.3+) ⭐
- **Priority**: Medium
- **Complexity**: Low
- **Impact**: Medium (clarity)
- **Description**: Namespaces with 0 pods showing "✅ Clean" is misleading - they're empty, not healthy
- **Problem**:
  - Empty namespaces (0 pods) show same "✅ Clean" as healthy namespaces
  - User can't distinguish between "healthy" vs "empty"
  - Leads to false confidence: "This namespace looks good!" (but it's just empty)
- **Proposed Fix**:
  - If namespace.PodCount == 0:
    - Display "📭 Empty" instead of "✅ Clean"
  - Use different indicator to show "no pods = no data" not "no problems"
- **Philosophy**: Distinguish "no issues" from "no data" - they mean different things

### Age Display Consistency (v0.5.3+) ⭐
- **Priority**: Low
- **Complexity**: Low
- **Impact**: Low (polish)
- **Description**: "N/A" age is uninformative - show creation time when available
- **Problem**:
  - Many resources show "N/A" for age
  - Bundle contains creation timestamps in metadata
  - User doesn't know if resource is 1 hour or 1 year old
- **Proposed Fix**:
  - Parse creation timestamp from kubectl output
  - Show relative age: "2d", "5h", "30m"
  - Fall back to "Unknown" (not "N/A") only when timestamp truly missing
- **Philosophy**: Show what we know - bundles have timestamps, use them


### Auto-Display Pod Diagnostics on Failure (v0.5.4) ✅
- **Status**: ✅ Shipped in v0.5.4
- **Description**: Enhanced "No Logs" diagnostic panel with maximum intel
- **Impact**: Users see comprehensive pod diagnostics automatically - r8s interprets data instead of showing raw output
- **Features**:
  - Intelligent diagnosis based on pod state (CrashLoop, OOMKilled, ImagePull, Error, Pending, Evicted)
  - Recent events display (last 5) with emoji indicators (always shown, even when empty)
  - State-specific investigation suggestions
  - External tools guidance (lnav, kubectl logs)
  - Works from both Dashboard and Classic navigation paths
  - Removed '[d]=describe pod' key to reduce confusion
- **Philosophy**: "r8s interprets, user acts" - Show intelligence, not just information
- **Deferred to v0.5.5**: Parse kubectl/events file for comprehensive event data

### Parse kubectl/events File for Rich Event Data (v0.5.5) ⭐
- **Priority**: HIGH
- **Complexity**: Medium
- **Impact**: High (data completeness)
- **Description**: Parse global kubectl/events file to show ALL pod events, not just attached ones
- **Current State**:
  - Diagnostic panel uses `pod.KubectlEvents` (events attached to pod object)
  - Many events are missing: scheduling failures, volume mount issues, network problems
  - Bundle contains `rke2/kubectl/events` file with ALL cluster events
- **Proposed Enhancement**:
  - Parse `kubectl/events` file on bundle load
  - Filter events by pod name to get complete event history
  - Show FailedScheduling, FailedMount, NetworkNotReady, etc.
  - Display event counts: "(x47)" for repeated events
  - Smart defaults: Show last 5, warnings first, sorted by time
- **Benefits**:
  - Complete event history for better diagnostics
  - See why pods failed to schedule
  - Detect volume and network issues
  - More accurate troubleshooting guidance
- **Reference**: See `docs/archive/2025-12-01/LOG_BUNDLE_ANALYSIS.md` for bundle structure details
- **Philosophy**: "Maximum information extraction" - Use all available bundle data

### Bundle Health Indicator & Resilience (v0.5.5) ⭐
- **Priority**: HIGH  
- **Complexity**: Low-Medium
- **Impact**: High (transparency + resilience)
- **Description**: Show bundle completeness and add smart fallbacks for all optional files
- **Current State (v0.5.4)**:
  - ✅ Namespaces: Falls back to deriving from pods
  - ❌ Other files: Silent failures, no fallbacks
  - ❌ No visibility into what's missing
- **Layer 1: More Smart Fallbacks**:
  - Apply namespace fallback pattern to other resources
  - Derive what we can, fail gracefully for rest
  - Verbose warnings show what was derived vs missing
- **Layer 2: Bundle Health Indicator**:
  - Status bar shows bundle completeness: `[BUNDLE 73%]`
  - Tooltip/help shows which files present/missing
  - Color coding: Green (>90%), Yellow (70-90%), Red (<70%)
- **Layer 3: Verbose Loading**:
  - Show exactly what was found/missing during load
  - Example: "✓ pods: 93, ⚠️ namespaces: derived from pods, ⚠️ events: missing"
- **Philosophy**: "Show, Don't Ask" - Transparency about data quality without blocking
- **Triggered by**: User reported partial bundle with missing namespaces file (v0.5.4)

### Remove Log Filter Modes (v0.5.5+)
- **Priority**: Medium
- **Complexity**: Medium (removal + smart defaults)
- **Impact**: High (simplification)
- **Description**: Eliminate ALL/ERROR/WARN filter toggles - use intelligent defaults with prominent highlighting
- **Current State**:
  - 3 filter modes: ALL, ERROR, WARN
  - Ctrl+A / Ctrl+E / Ctrl+W hotkeys
  - Status bar shows current filter
  - User must manually toggle to see issues
- **Proposed Change (Option A - Pure Simplification)**:
  - Remove filter modes entirely
  - Always show ALL logs by default
  - Errors highlighted RED, warnings YELLOW (already implemented in log view)
  - No filter needed - issues visually surface themselves
- **Proposed Change (Option B - Smart Auto-Detect)**:
  - Auto-detect log content on open:
    - If < 100 lines: show ALL (no filtering needed)
    - If 100-1000 lines and >10 errors: show ERROR
    - If >1000 lines and >20 warnings: show WARN
    - Otherwise: show ALL
  - Add single "Show All" toggle (Ctrl+A) to override smart filter
  - Remove ERROR/WARN individual modes
- **Recommendation (Option A - More Aligned with Principles)**:
  - Option A better aligns with "Show, Don't Ask" - show everything, highlight issues
  - Reduces 3 choices to 1 (toggle available if needed)
  - User sees full context immediately
  - Color highlighting already works - leverage it
- **Rationale**:
  - Software should show data, not hide it behind filters
  - 90% of time: user wants context + errors visible together
  - Color highlighting already surfaces issues
  - Reduce 3 choices to 0 (no filtering needed)
- **Code Reduction**: ~100 lines removed
- **Philosophy**: "Show everything, highlight issues - don't hide context behind filters"

### Remove View Switching Hotkeys (v0.6.0+)
- **Priority**: Low
- **Complexity**: Medium (behavior change)
- **Impact**: Medium (simplification)
- **Description**: Eliminate 1/2/3/4/5 view switching - use context-aware navigation only
- **Current State**:
  - Hotkeys: 1=Pods, 2=Deployments, 3=Services, 4=CRDs, etc.
  - User must remember mapping
  - Rarely used (Enter/Esc navigation more natural)
- **Proposed Change**:
  - Remove number hotkeys entirely
  - Navigation via Enter (drill down) and Esc (go up) only
  - Add breadcrumb-based "back to X" shortcuts if needed
- **Rationale**:
  - Enter/Esc is intuitive hierarchy navigation
  - Number keys aren't discoverable (not shown in help)
  - Simplifies mental model: "Just press Enter"
- **Code Reduction**: ~50 lines removed
- **Philosophy**: "Fewer ways to do same thing = easier to learn"

### Navigation Simplification (v0.6.0) ⭐
- **Priority**: HIGH
- **Complexity**: Medium
- **Impact**: High (code quality + reliability)
- **Description**: Standardize navigation state management across all view transitions
- **Current Pain Points**:
  - State clearing code duplicated in 3 locations (handlers.go ViewPods, app.go ViewAttention x2)
  - Classic pod view and dashboard use different approaches
  - Race condition fix required adding validation layer due to inconsistent state management
  - Each navigation path has subtle behavioral differences
- **Root Cause**: Organic growth - features added incrementally without unified architecture
- **Proposed Solution**:
  ```go
  // Single source of truth for navigation state clearing
  func (a *App) navigateToLogs(namespace, podName string) tea.Cmd {
      // Clear ALL pod-related state in one place
      a.clearPodState()
      
      // Set new context
      a.currentView = ViewContext{
          viewType: ViewLogs,
          namespaceName: namespace,
          podName: podName,
      }
      
      // Fetch logs
      return a.fetchLogs(...)
  }
  
  func (a *App) clearPodState() {
      a.logs = nil
      a.searchMode = false
      a.searchQuery = ""
      a.searchMatches = nil
      a.currentMatch = -1
      a.showRawLogs = false  // Diagnostic-first
      a.filterLevel = ""
  }
  ```
- **Benefits**:
  - Zero duplication - state clearing happens once
  - Consistent behavior across all navigation paths
  - Easier to maintain and extend
  - Race condition validation becomes simpler
  - Faster development of new navigation features
- **Code Reduction**: ~40 lines removed (3 duplicate implementations → 1)
- **Testing**: Navigation reliability test suite to prevent regressions
- **Philosophy Alignment**: "One way to do it" - eliminate subtle path differences

### Navigation UX Analysis (Jan 2026) - Context for v0.6.0 Planning

**Analysis Date**: 2026-01-06

**Current Navigation Complexity Score**: 8.5/10 (excellent, minor improvements possible)

#### What Works Well ✅

1. **Dashboard-First Design (v0.3.3)** - Revolutionary decision, enables 2-key triage
   - Dashboard → Enter → Diagnostic Panel → l → Logs = 2-3 keys total
   - Smart detection surfaces worst issues automatically
   - Average triage time: 15-20 seconds (verified via interactive session)

2. **Diagnostic-First Approach (v0.5.4-v0.5.6)** - Intentional design, not a bug
   - Enter from dashboard shows **diagnostic panel** (not logs) first
   - Rationale: Many pods have no logs (CrashLoopBackOff, ImagePullBackOff)
   - Provides context (events, status) before showing logs
   - Aligns with "Maximum Information Extraction" philosophy
   - One additional keypress (`l`) to reach logs when they exist

3. **Classic Mode Serves Distinct Use Case** - NOT redundant
   - Dashboard = Triage (jump to worst issues)
   - Classic = Exploration (browse all namespaces/pods systematically)
   - Different mental models for different workflows
   - Problem is **state management**, not existence

4. **Sort Mode Indicators** - Already implemented (v0.4.3)
   - Status bar shows: `Sort: Count` / `Sort: Severity` / `Sort: Name`
   - Critical count visible: `🔥 Criticals: 5`
   - Footer shows context-specific actions

5. **Smart Capping with Expansion (v0.4.0)** - Prevents overflow at scale
   - Default top-20 cap with `m` toggle
   - Dynamic cap ensures ALL criticals included regardless of position
   - Position indicator: "Showing 20/86"

6. **Truth Indicators (v0.5.2-v0.5.3)** - Transparency about data availability
   - "⚠️ No Logs" for pods without log files
   - "📭 Empty" for empty namespaces
   - Prevents false confidence from "✅ Clean" on missing data

#### Recommendations for v0.6.0

**TIER 1: High-Value, Low-Effort (Implement These)** ⭐⭐⭐

1. **Help Discovery Hint** (Aligns with "Show, Don't Ask" principle)
   ```
   Current:  (no hint about help system)
   Proposed: Footer includes "Press ? for help" (first 3 launches only)
   ```
   - **Impact**: Solves discovery problem for new users
   - **Effort**: ~30 lines of code (add launch counter to config)
   - **Philosophy**: Information should surface automatically

2. **Expansion State Clarity**
   ```
   Current:  Showing 9/100 · [m]=expand
   Proposed: Showing 9/100 (capped) · [m]=show all 100
   ```
   - **Impact**: Users immediately understand expansion state
   - **Effort**: ~5 lines (string formatting change)

**TIER 2: DO NOT Implement (Violates Design Principles)** ❌

1. **Auto-Log on Dashboard Enter** - REJECTED
   - **Why**: Violates v0.5.4-v0.5.6 diagnostic-first design
   - **Rationale**: Many pods have no logs; diagnostics provide context first
   - **Current design is superior**: Show intelligence (diagnostics) before data (logs)
   - **Philosophy**: "Maximum Information Extraction" > "Fewer keypresses"

2. **Remove Classic Mode** - REJECTED
   - **Why**: Serves distinct exploration use case vs dashboard's triage
   - **Reality**: Different workflows need different navigation models
   - **Real problem**: State management complexity (already tracked for v0.6.0)
   - **Keep**: Classic mode
   - **Fix**: State management duplication (already in v0.6.0 scope)

3. **Command Palette (`:` key)** - DEFERRED
   - **Why**: Adds complexity without clear user demand
   - **When**: Consider for v0.7+ if user feedback requests it
   - **Philosophy**: "Best feature = no feature" - don't add without proven need

**TIER 3: Already Solved (No Action Needed)** ✅

1. Silent Mode Changes - Solved in v0.4.3 (status bar indicators)
2. Truth Indicators - Solved in v0.5.2-v0.5.3
3. Smart Capping - Solved in v0.4.0
4. Diagnostic Intelligence - Solved in v0.5.4-v0.5.6

#### Interactive Testing Results (Jan 2026)

**Measured Workflows:**
- Dashboard → worst pod logs: **2 keys** (position cursor, Enter, then `l`)
- Dashboard → classic pods: **4 keys** (c, Enter, Enter, Enter)
- Sort cycling: **1 key** (`s` toggles through modes)
- Jump to top/bottom: **1 key** (`g` or `G`)
- Help access: **1 key** (`?`)

**Key Findings:**
- Footer shows available actions context-sensitively
- Sort mode indicators work correctly
- `m` toggle for expansion functions as designed
- Help system is comprehensive but requires discovery
- Diagnostic panel shows maximum intel before logs

#### Design Rationale Validated

**Why Diagnostic-First (not Auto-Log)?**
1. CrashLoopBackOff pods often have **zero logs**
2. Events and pod status explain **why** crash happened
3. Context before content = better troubleshooting
4. One extra keypress (`l`) is acceptable trade-off
5. Aligns with CHANGELOG v0.5.4-v0.5.6 philosophy: "r8s interprets, user acts"

**Why Keep Classic Mode?**
1. Dashboard = "Show me worst issues" (triage)
2. Classic = "Let me browse everything" (exploration)
3. Both serve 40%+ of use cases (estimated from development history)
4. State management is the bug, not feature duplication

#### Target for v0.6.0: 9.5/10 Simplicity

**Current**: 8.5/10

**Required Changes**:
1. ✅ Help discovery hint (1-line footer + launch counter)
2. ✅ Expansion state clarity (string formatting)
3. ✅ Navigation state consolidation (already scoped)

**DO NOT Change**:
- ❌ Diagnostic-first design (intentional, superior)
- ❌ Classic mode existence (serves distinct use case)
- ❌ Sort mode toggles (useful for different scenarios)
- ❌ Current keypress counts (already optimal)

**Philosophy Alignment**:
- **Truth Only™**: Show accurate data availability (solved v0.5.2-v0.5.3)
- **Best Feature = No Feature**: Don't add command palette without demand
- **Show, Don't Ask**: Help hint surfaces automatically (proposed)
- **Minimal Keys**: 2-key triage already achieved
- **Complete Removal**: v0.6.0 state refactor (not feature removal)

#### Lessons from History

**v0.3.5 Live Mode Removal**: 1,200 lines deleted in 22 minutes
- Lesson: Complete removal works when feature has low usage
- Does NOT apply to classic mode (serves distinct, validated use case)

**v0.4.0 Smart Capping**: Fixed overflow at scale
- Lesson: Test features at 10× scale before shipping
- Already applied to navigation analysis

**v0.5.4-v0.5.6 Diagnostic Panel**: Intentional intermediate screen
- Lesson: Context before content improves troubleshooting
- Validates diagnostic-first approach over auto-log

**Conclusion**: r8s navigation already exemplifies "Minimal Keys" principle. Focus v0.6.0 on state management cleanup, not feature reduction.

## 🚀 Long-Term Ideas (v0.6.0+)

### Real-Time Monitoring
- **Description**: Support live cluster monitoring (not just bundles)
- **Requirements**: 
  - Kubernetes API client integration
  - Auto-refresh mode
  - Connection status indicator

### Advanced Search
- **Description**: Regex search across all logs in bundle
- **Requirements**:
  - Global search mode
  - Results aggregation across pods
  - Jump-to-log functionality

### Log Export & Reporting
- **Description**: Export filtered logs or generate issue reports
- **Requirements**:
  - Export selected logs to file
  - Generate markdown summary
  - Email report capability

### Multi-Bundle Comparison
- **Description**: Compare two bundles side-by-side
- **Requirements**:
  - Load two bundles simultaneously
  - Diff view for configuration changes
  - Timeline comparison

### Plugin System
- **Description**: Allow custom signal detection plugins
- **Requirements**:
  - Plugin API specification
  - Custom pattern matching
  - User-defined issue types

## 📚 Documentation Enhancements

### Video Tutorials
- Quick start (3 min)
- Advanced navigation (5 min)
- Custom signals (7 min)

### Use Case Examples
- RKE2 cluster troubleshooting walkthrough
- K3s debugging scenarios
- Rancher upgrade troubleshooting

### Pattern Library
- Common error patterns and solutions
- Known issues database
- Community-contributed patterns

## 🔧 Technical Debt

### Code Quality & Refactoring (Deferred from Jan 2026 Audit)

#### Log View Edge Case Documentation
- **Priority**: Low
- **Complexity**: Low
- **Impact**: Low (maintainability)
- **Description**: Add inline comments explaining whitespace-search edge cases in log wrapping
- **Location**: `internal/tui/logs.go` lines 186-244
- **Requirements**:
  - Document fallback behavior when no whitespace found within wrapWidth
  - Explain lastSpace == 0 and lastSpace == -1 edge cases
  - Note that subsequent wrapped segments trim leading spaces via TrimLeft
  - Clarify intentional split at wrapWidth for very long words (URLs/base64)
- **Rationale**: Future maintainers need to understand the wrapping logic for bug fixes

#### Dynamic Help Text Height Calculation
- **Priority**: Low
- **Complexity**: Low
- **Impact**: Low (polish)
- **Description**: Compute contentHeight dynamically instead of hardcoded value
- **Location**: `internal/tui/logs.go` lines 471-477
- **Current State**: `contentHeight := 12` is hardcoded
- **Requirements**:
  - Use `strings.Count(helpText, "\n") + X` to compute actual height
  - Adjust constant offset X to account for title, borders, margins
  - Import strings package if needed
- **Rationale**: Prevents misalignment when helpText changes

#### Extract Namespace Parsing Helper
- **Priority**: Medium
- **Complexity**: Low
- **Impact**: Medium (code quality)
- **Description**: Extract duplicated namespace extraction logic into helper function
- **Location**: `internal/tui/table.go` lines 244-255 (and other locations)
- **Current State**: Logic for splitting "cluster:namespace" duplicated across ViewPods, ViewDeployments, ViewServices
- **Proposed Fix**:
  ```go
  // In helpers.go
  func extractNamespaceName(namespaceID string) string {
      if namespaceID == "" {
          return "default"
      }
      if strings.Contains(namespaceID, ":") {
          parts := strings.Split(namespaceID, ":")
          if len(parts) > 1 {
              return parts[1]
          }
      }
      return namespaceID
  }
  ```
- **Replace**: All instances of namespace extraction with single function call
- **Rationale**: DRY principle, centralize parsing logic

#### Cache Namespace Health Computation
- **Priority**: Medium
- **Complexity**: High
- **Impact**: High (performance)
- **Description**: Cache expensive namespace health computation instead of recalculating on every render
- **Location**: `internal/tui/table.go` lines 130-145
- **Current State**: `ComputeNamespaceHealth` and `SortNamespacesByHealth` run on every table update
- **Requirements**:
  - Add cache fields to table struct: `cachedNsHealth`, `cachedSortedNS`, `cachedScanDepth`
  - Add invalidation flag or last-namespaces-version tracker
  - On render: check if cache valid (same scanDepth, namespaces unchanged)
  - Reuse cached values if valid, otherwise recompute and update cache
  - Invalidate cache on namespace modification or refresh key ('r')
  - Use mutex if table updates can happen concurrently
- **Expected Impact**: Significant performance improvement for large clusters (25+ namespaces)
- **Rationale**: Scanning logs for all pods in all namespaces on every render is expensive

### Test Coverage
- Increase unit test coverage to 80%+
- Add integration tests for bundle loading
- Performance benchmarks for large bundles
- Update breadcrumb tests to expect empty mode indicator (not "[LIVE]")

### Code Quality
- Refactor attention.go (too large)
- Extract signal detection to strategy pattern
- Reduce cyclomatic complexity in app.go

### Performance Optimization
- Lazy-load logs (don't scan all pods upfront)
- Streaming parser for huge log files
- Memory profiling for 1GB+ bundles

---

## Notes

- Features are moved here when they're identified but not critical for current release
- Priority/complexity/impact ratings help with future planning
- Community suggestions welcome via GitHub issues
- Mark items as ✅ when moved to active development

Last updated: 2025-12-10 (v0.3.9 - Tunable scan depth shipped)
