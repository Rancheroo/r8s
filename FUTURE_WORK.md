# r8s Future Work & Deferred Features

**Last Updated**: 2026-01-13 (post v0.6.7 release)

This document tracks feature ideas and enhancements deferred to future releases.

---

## 🎯 Active Roadmap

### v0.6.8 "Event Truncation + Cluster Drill-Down" (Next Release)

**See**: `V0.6.8_PROMPT.md` for complete implementation prompt

**Priority Items**:

1. **Event Message Truncation** (MEDIUM-HIGH) - Issue #20
   - Truncate long registry pull errors to ~80 chars
   - Smart truncation preserves key info (event type, resource, error code)
   - Add "(truncated)" indicator
   - **Impact**: Diagnostic panel readability, container status section visibility

2. **Cluster-Level Drill-Down** (MEDIUM-HIGH) - Issue #17 continuation
   - Extend v0.6.1 cluster event drill-down to kubelet/node/etcd items
   - Show impacted pods list with 1-9 selection
   - Map kubelet errors → pods on affected nodes
   - Map node pressure → pods scheduled on node
   - **Impact**: Complete Issue #17 fix

3. **Dashboard Re-Sort on Navigation Return** (LOW) - Issue #21
   - Cache dashboard order between navigations for visual stability
   - Optional polish item (only if time permits)

---

## 📋 Backlog (v0.6.9+)

### High Priority

**Namespace Health Ranking** (v0.6.9)
- Add "ISSUES" column showing E/W counts per namespace
- Auto-sort by total issue count (worst first)
- Color-code: 🔥 (>50 errors), ⚠️ (>20 warnings), ✅ (clean)
- **Use Case**: "Which namespace should I investigate first?"
- **Impact**: High - Quick identification of problematic namespaces

**Parse kubectl/events File** (v0.6.9)
- Show ALL pod events (FailedScheduling, FailedMount, NetworkNotReady)
- Display event counts: "(x47)" for repeated events
- **Impact**: High - Complete event history for better diagnostics

### Medium Priority

**Bundle Health Indicator** (v0.7.0)
- Status bar shows completeness: "📦 BUNDLE 73%"
- Tooltip shows which files present/missing
- Color coding: Green (>90%), Yellow (70-90%), Red (<70%)
- **Impact**: Medium - Transparency about data quality

**Enhanced Help Panel** (v0.7.0)
- Context-aware pro tips per view
- "Start with dashboard for quick wins"
- "Use Ctrl+W in logs to focus on issues"
- **Impact**: Medium - Better discoverability

### Low Priority

**Age Display Consistency**
- Parse creation timestamps from bundle metadata
- Show relative age: "2d", "5h", "30m"
- Fall back to "Unknown" (not "N/A")

**Edge Case Handling**
- Empty logs: "No E/W — check describe/events"
- Parse errors: Show count in bundle load warning

---

## 🔧 Technical Debt

### Code Quality

**Extract Namespace Parsing Helper** (v0.7.0)
- Duplicated namespace extraction logic across ViewPods/Deployments/Services
- Create `extractNamespaceName()` helper function
- **Impact**: Medium - DRY principle, centralize logic

**Cache Namespace Health Computation** (v0.7.0)
- Expensive computation runs on every render
- Add cache with invalidation on refresh/modification
- **Impact**: High - Performance improvement for large clusters

**Log View Edge Case Documentation**
- Add inline comments for whitespace-search edge cases (lines 186-244)
- Document fallback behavior for long words

**Dynamic Help Text Height**
- Compute contentHeight dynamically vs hardcoded value
- Use `strings.Count(helpText, "\n") + X`

### Test Coverage

- Increase unit test coverage to 80%+
- Add integration tests for bundle loading
- Performance benchmarks for large bundles

---

## 🚀 Long-Term Ideas (v0.7+)

### Real-Time Monitoring
- Support live cluster monitoring (not just bundles)
- Kubernetes API client integration
- Auto-refresh mode

### Advanced Search
- Regex search across all logs in bundle
- Results aggregation across pods
- Jump-to-log functionality

### Log Export & Reporting
- Export filtered logs to file
- Generate markdown summary
- Email report capability

### Multi-Bundle Comparison
- Load two bundles simultaneously
- Diff view for configuration changes
- Timeline comparison

### Plugin System
- Custom signal detection plugins
- User-defined detection rules
- Community-contributed patterns

---

## 📚 Documentation

### Video Tutorials (v0.7+)
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

---

## 💡 Design Philosophy

**Core Principles** (see LESSONS-LEARNED.md for details):
- **Truth Only™** - Display only accurate, verified data
- **Show, Don't Ask** - Information surfaces automatically
- **Best Feature = No Feature** - Smart defaults beat options
- **5-Second Rule** - Users know what to fix in ≤5 seconds

**UX Principles** (v0.5.2+):
- Auto-display health indicators (no button presses needed)
- Auto-show parse warnings in status bar
- Smart defaults over configuration
- Progressive information disclosure

---

## ✅ Recently Completed (v0.6.7)

### Inline Diagnostics
- 2-line format with root cause + fix recommendations
- DiagnosticContext for CrashLoopBackOff, OOMKilled, ImagePullBackOff
- "5-Second Rule" validation

### Navigation Fixes
- Issue #17 (partial): Non-pod items no longer navigate to wrong diagnostics
- Issue #17 (partial): Cursor/visual order mismatch fixed
- 9 comprehensive navigation tests

### UX Polish
- Issue #18: Dynamic number width for 10+ items
- Issue #19: Bundle health percentage display restored

**See**: CHANGELOG.md for complete v0.6.7 details

---

## 📖 References

- **V0.6.8_PROMPT.md** - Next release prompt
- **docs/V0.6.X_ROADMAP.md** - Complete roadmap with all prompts
- **CHANGELOG.md** - Release history
- **LESSONS-LEARNED.md** - Design principles

---

## Notes

- Items marked ✅ in CHANGELOG.md are removed from this file
- Priority ratings: HIGH > MEDIUM > LOW
- Complexity: Low (1-2 days) > Medium (3-5 days) > High (1+ weeks)
- Impact: User-facing benefit (High > Medium > Low)

Last updated: 2026-01-13 (v0.6.7 shipped, v0.6.8 next)
