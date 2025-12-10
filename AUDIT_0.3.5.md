# r8s 0.3.5 "Bundle-Only Bliss" — Pre-Release Audit

**Date:** 2025-12-10  
**Goal:** Remove live mode, optimize bundle mode, ship the best bundle analyzer possible  
**Time Budget:** 2 hours total development time

---

## 10-POINT AUDIT FINDINGS

### 1. **LIVE MODE SPRAWL**
**Impact:** HIGH | **Lines:** ~800 | **Files:** 8+

Live mode touches nearly every layer of the application:
- `internal/datasource/live.go` (230 lines) — Rancher API wrapper
- `internal/rancher/client.go` (300 lines) — HTTP client with auth
- `internal/rancher/client_test.go` (100 lines) — Client tests
- `internal/rancher/types.go` (200+ lines) — API response types
- `internal/config/config.go` — Profile/credential management
- `cmd/root.go` — Profile, insecure, config flags
- `cmd/tui.go` — Live connection initialization
- `internal/tui/app.go` — Client field, mock data fallbacks

**Removal Plan:** Delete live.go, client.go entirely. Move essential types to bundle package. Simplify config to runtime flags only.

---

### 2. **TEST FAILURES & COVERAGE GAPS**
**Impact:** HIGH | **Coverage:** 6.6% TUI, 0% bundle, 0% datasource

Current test failures:
```
--- FAIL: TestNewApp/valid_config_creates_app (0.39s)
    app_test.go:64: Expected view type 1, got 0
--- FAIL: TestNewApp/no_profiles_creates_app_with_error (0.00s)
    app_test.go:64: Expected view type 1, got 0
```

Root cause: Tests expect `ViewClusters` (1) but bundle mode defaults to `ViewAttention` (0).

**Coverage by package:**
- `internal/tui`: 6.6% (tests exist but fail)
- `internal/config`: 47.1% (decent)
- `internal/rancher`: 55.7% (becomes 0% after removal)
- `internal/bundle`: 0% (critical gap!)
- `internal/datasource`: 0% (critical gap!)

**Fix Plan:** Update tests to expect ViewAttention. Add bundle parsing tests, attention computation tests, error detection tests. Target: 85%+ on core paths.

---

### 3. **150MB BUNDLE SIZE LIMIT**
**Impact:** MEDIUM | **Current:** 100MB | **Target:** 200MB+

Location: `internal/datasource/bundle.go:19`
```go
opts := bundle.ImportOptions{
    Path:    bundlePath,
    MaxSize: 100 * 1024 * 1024, // 100MB for TUI mode
    Verbose: verbose,
}
```

**Issues:**
- Real-world support bundles often exceed 100MB
- No streaming tar support (must extract manually)
- All bundle content loaded into memory at once

**Optimization Plan:**
- Increase limit to 200MB
- Add lazy log loading (don't read log files until viewed)
- (Stretch) Stream tar.gz extraction

---

### 4. **NO STARTUP SUMMARY BAR**
**Impact:** MEDIUM | **Status:** Missing entirely

Users get zero feedback on bundle load success:
- How many pods were found?
- How many log lines analyzed?
- How many issues detected?

**UX Goal:**
```
┌───────────────────────────────────────────────────────────────────────┐
│ ✓ Loaded w-guard-wg-cp-svtk6-lqtxw • 428 pods • 1.8M lines • 142 issues │
└───────────────────────────────────────────────────────────────────────┘
```

**Implementation:** Compute during bundle load, display before attention dashboard.

---

### 5. **MODE INDICATOR INCOMPLETE**
**Impact:** LOW | **Current:** `[BUNDLE]` | **Target:** `Bundle • 2025-11-27 04:19`

Current breadcrumb shows mode but no bundle timestamp:
```
[BUNDLE] r8s - Attention Dashboard
```

**Desired:**
```
Bundle • 2025-12-04 09:15 | Attention Dashboard
```

**Implementation:** Extract timestamp from bundle manifest or directory name.

---

### 6. **ATTENTION DASHBOARD NOT CACHED**
**Impact:** MEDIUM | **Performance:** Re-scans all logs on every refresh

Location: `internal/tui/app.go` → `fetchAttention()`

Every time the dashboard refreshes, it:
1. Re-reads all pod logs
2. Re-scans every line for errors/warnings
3. Re-computes attention items

On a 100MB bundle with 1M log lines, this takes 2-3 seconds.

**Optimization:** Compute once on bundle load, cache in App struct, only refresh on explicit user request.

---

### 7. **DEAD CODE: MOCK DATA GENERATORS**
**Impact:** LOW | **Lines:** ~200 | **Status:** Legacy

Locations in `internal/tui/app.go`:
- `getMockPods()` — 50 lines
- `getMockDeployments()` — 40 lines
- `getMockServices()` — 60 lines
- `getMockClusters()` — 30 lines
- `getMockProjects()` — 20 lines
- `getMockNamespaces()` — 30 lines
- `getMockCRDs()` — 60 lines
- `getMockCRDInstances()` — 80 lines

These were used before bundle mode worked. Now they're dead weight — live mode will be removed, and embedded demo uses actual bundle data.

**Cleanup:** Delete all `getMock*()` functions.

---

### 8. **UX WINS ALREADY IMPLEMENTED**
**Impact:** POSITIVE | **Status:** Just need docs

Already working (but not documented):
- ✅ `g` / `G` — Jump to first/last log line (vim muscle memory)
- ✅ `w` — Toggle word wrap for long lines
- ✅ `b` — Back navigation (in addition to Esc)
- ✅ Dashboard → logs drill-down works

**Small tweak needed:** Dashboard drill-down currently pre-filters to WARN. Should show ALL logs with cursor at first error.

---

### 9. **README PROMINENTLY FEATURES LIVE MODE**
**Impact:** HIGH | **Documentation debt:** 15+ live references

Current README hero:
```
✅ **Live Mode** - Browse Rancher clusters, projects, namespaces  
✅ **Bundle Mode** - Analyze RKE2 bundles offline (no API needed) 
```

**New hero:**
```
r8s — the fastest way to understand a broken Kubernetes cluster from a log bundle.
```

Live mode mentioned in:
- Features section
- Quick start (config init for Rancher credentials)
- Common workflows
- Troubleshooting (connection errors)

**Cleanup:** Remove all live references, rewrite for bundle-first workflow.

---

### 10. **RANCHER PACKAGE MOSTLY UNUSED POST-REMOVAL**
**Impact:** LOW | **Future cleanup opportunity**

After live mode removal, `internal/rancher/` only needed for types:
- `Pod`, `Deployment`, `Service`, `CRD` structs
- `Event`, `Node`, `Namespace` structs

**Options:**
1. Move types to `internal/bundle/types.go` (clean break)
2. Keep `internal/rancher/types.go` but delete client code (maintains import paths)
3. Inline minimal types directly into datasource package

**Recommendation:** Option 2 for minimal disruption. Can refactor later if needed.

---

## BASELINE METRICS (Pre-0.3.5)

### Build
```bash
$ make build
real    0m0.579s
user    0m0.800s
sys     0m0.255s
```

### Test Coverage
```
internal/config:      47.1%
internal/rancher:     55.7%
internal/tui:         6.6% (FAILING)
internal/bundle:      0.0%
internal/datasource:  0.0%
```

### Code Stats
```
Total lines (internal/ + cmd/): 10,222 lines of Go code
```

**Files with live mode code to be removed:**
- internal/datasource/live.go: ~230 lines
- internal/rancher/client.go: ~300 lines
- internal/rancher/client_test.go: ~100 lines
- Mock generators in app.go: ~200 lines
**Expected reduction: ~830 lines (8.1% codebase reduction)**

---

## SUCCESS CRITERIA FOR 0.3.5

### Must Have
- [ ] Live mode completely removed (0 references in code)
- [ ] All tests passing (`make test` returns 0)
- [ ] Bundle mode works identically to 0.3.4
- [ ] Default `r8s` command launches embedded demo instantly
- [ ] Tests cover 85%+ of bundle/datasource/attention code paths

### Should Have
- [ ] Startup summary bar showing pod/log/issue counts
- [ ] Bundle timestamp in mode indicator
- [ ] Attention dashboard cached (1-2s faster on large bundles)
- [ ] All dead code removed (mock generators)
- [ ] Documentation updated (README, USAGE, TROUBLESHOOTING)

### Nice to Have
- [ ] 200MB bundle support
- [ ] Lazy log loading
- [ ] Dashboard→logs shows all logs with smart cursor positioning

---

## BRANCH SEQUENCE

1. ✅ `audit` — This document + baseline metrics
2. `remove-live-mode` — Delete ~800 lines, simplify everything
3. `bundle-optimizations` — Performance + caching
4. `ux-perfection` — Startup bar, mode indicator, help updates
5. `tests` — Fix failures, add coverage, reach 85%+
6. `docs-and-release` — README, CHANGELOG, LESSONS-LEARNED, version bump

---

## NOTES

- **Time pressure is real:** 2 hours to ship is ambitious but doable
- **Test-first for phase 5:** Write failing tests, then fix
- **No feature creep:** This is removal + optimization, not new features
- **User value:** Instant demo launch + faster bundle analysis = big UX win

---

**Status:** Audit complete. Ready for Phase 2: Remove Live Mode.
