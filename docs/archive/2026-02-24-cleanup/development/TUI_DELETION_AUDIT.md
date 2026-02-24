# TUI Deletion Audit: Musk's 5 Laws Analysis

**Date:** 2026-02-17  
**Author:** RancherSRE  
**Status:** Pending Review  
**Related:** Sprint 8 planning, Law #2 (Delete)

---

## Executive Summary

The r8s TUI currently contains **9,360 lines across 19 files**. Analysis through Musk's 5 Laws identifies **~860 lines of deletable code** (9.2% reduction) from dead features, unused view types, and over-engineered sorting.

**Key Finding:** The TUI still carries code from the pre-v0.3.5 live Rancher API mode, which was removed from the product but not from the codebase.

---

## Current State

| Metric | Value |
|--------|-------|
| Total TUI Lines | 9,360 |
| TUI Files | 19 |
| View Types | 12 |
| State Fields | 25+ |
| Sort Modes | 3 |

---

## Musk's 5 Laws Applied

### 🔴 Law #2: Delete What's Not Needed

| Candidate | Location | Lines | Rationale |
|-----------|----------|-------|-----------|
| `ViewClusters` constant | `app.go:28` | 1 | Live API mode only |
| `ViewProjects` constant | `app.go:29` | 1 | Live API mode only |
| `ViewNamespaces` constant | `app.go:30` | 1 | Live API mode only |
| `clusters` state field | `app.go:~85` | 1 | Unreferenced in bundle mode |
| `projects` state field | `app.go:~86` | 1 | Unreferenced in bundle mode |
| `namespaces` state field | `app.go:~87` | 1 | Unreferenced in bundle mode |
| `fetchClusters()` method | `fetch.go` | ~80 | Live API only |
| `fetchProjects()` method | `fetch.go` | ~80 | Live API only |
| `fetchNamespaces()` method | `fetch.go` | ~80 | Live API only |
| Clusters/Projects/Namespaces table renders | `table.go`, `helpers.go` | ~150 | Dead code paths |
| Associated message types | `app.go` | ~30 | `clustersMsg`, `projectsMsg`, etc. |
| **Subtotal** | | **~405** | |

**Evidence:** README states *"As of v0.3.5, we removed live Rancher API support"*, yet the TUI still carries clusters/projects/namespaces view infrastructure.

### 🟡 Law #1: Question Every Requirement

| Feature | Current State | Question | Verdict | Lines |
|---------|---------------|----------|---------|-------|
| Classic view (`c` key) | Toggle between dashboard and classic pod view | Does anyone use non-dashboard mode? | **DELETE** | ~50 |
| Pod sort by count | 3rd sort mode (Count → Severity → Name) | Is count sorting useful? | **DELETE** | ~30 |
| Mock data loading message | "Loading mock data (OFFLINE MODE)..." | Misleading terminology? | **REWORD** | ~5 |
| Full ViewContext struct | Contains Rancher IDs (`clusterID`, `projectID`) | Needed for bundle-only? | **SLIM** | ~20 |
| **Subtotal** | | | | **~105** |

### 🟢 Law #3: Simplify

| Current | Simplified | Benefit | Effort |
|---------|------------|---------|--------|
| 19 TUI files | 15-16 files (merge small ones) | Easier navigation | Medium |
| 3 sort modes | 2 modes (Severity, Name) | Less cognitive load | Low |
| Full bubble-table | Simple list for small datasets | Faster rendering | Medium |
| **Note:** File merging deferred to avoid merge conflicts | | | |

---

## Deletion Plan

### Phase 1: Dead View Types (Safe, High Impact)

**Files to modify:**
- `internal/tui/app.go` — Remove constants, state fields, message types
- `internal/tui/fetch.go` — Remove fetch methods
- `internal/tui/table.go` — Remove table renderers
- `internal/tui/helpers.go` — Remove helper functions

**Steps:**
1. Comment out `ViewClusters`, `ViewProjects`, `ViewNamespaces` constants
2. Remove `clusters`, `projects`, `namespaces` fields from `App` struct
3. Remove `clustersMsg`, `projectsMsg`, `namespacesMsg` types
4. Delete `fetchClusters()`, `fetchProjects()`, `fetchNamespaces()`
5. Delete associated table column definitions
6. Build and verify — no compilation errors
7. Run tests — ensure no regressions
8. Commit

**Risk:** Low — These code paths are unreachable in bundle-only mode.

### Phase 2: Classic View Toggle (Product Decision)

**Files to modify:**
- `internal/tui/app.go` — Remove `case "c":` handler
- `internal/tui/handlers.go` — Remove classic view toggle logic
- `README.md` — Update keyboard shortcuts

**Question for Product:** Is the classic pod view used by anyone?

**Evidence against:**
- Dashboard is default on launch
- Classic view requires knowing `c` key exists
- No documentation mentions classic view
- Attention Dashboard supersedes it

**Risk:** Low-Medium — Verify with user feedback before deletion.

### Phase 3: Sort Simplification (UX Improvement)

**Files to modify:**
- `internal/tui/attention.go` — Remove count sort mode
- `internal/tui/app.go` — Update sort cycle logic

**Current cycle:** Count → Severity → Name → Count  
**Proposed cycle:** Severity → Name → Severity

**Risk:** Very Low — Simple change, easy to reverse.

### Phase 4: Code Hygiene (Nice to Have)

- Reword "mock data" → "demo bundle" in loading message
- Slim ViewContext struct (remove Rancher-specific IDs)
- Merge small files (`styles.go` into `helpers.go`?)

---

## Expected Outcomes

| Metric | Before | After Phase 1-3 | Reduction |
|--------|--------|-----------------|-----------|
| TUI Lines | 9,360 | ~8,500 | 9.2% |
| View Types | 12 | 9 | 25% |
| State Fields | 25+ | 20 | 20% |
| Sort Modes | 3 | 2 | 33% |
| Binary Size | ~10.7MB | ~10.5MB | ~2% |

---

## Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Compilation errors from deleted references | Medium | Low | Build after each deletion batch |
| Test failures | Low | Low | Run full test suite |
| User backlash (classic view) | Low | Medium | Add deprecation warning first |
| Merge conflicts | Medium | Medium | Coordinate with active PRs |

---

## Related Work

- **Sprint 8:** Bundle Health v2 + CLI commands — Good time for cleanup
- **Issue #44:** golangci-lint — Dead code detection may flag these
- **MEMORY.md:** v0.3.5 live mode removal decision

---

## Next Steps

1. **Review** — Product confirmation on classic view deletion
2. **Prioritize** — Bundle with Sprint 8 or defer to v0.7.3
3. **Implement** — Phase 1 (safe deletions) in Sprint 8
4. **Evaluate** — Phase 2-4 based on Sprint 8 capacity

---

## Musk's Law Checklist

- [ ] **Law #1 (Question):** Do users need classic view? — A: Probably not
- [ ] **Law #2 (Delete):** Remove dead view types — Ready to implement
- [ ] **Law #3 (Simplify):** Reduce sort modes — Ready to implement
- [ ] **Law #4 (Accelerate):** Less code = faster builds — Side benefit
- [ ] **Law #5 (Automate):** Dead code detection in CI — Future improvement

---

*Generated by RancherSRE during Sprint 8 planning*  
*Musk's Laws: Question → Delete → Simplify → Accelerate → Automate*
