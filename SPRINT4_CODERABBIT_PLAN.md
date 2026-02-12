# Sprint 4 CodeRabbit Review Plan

**Goal:** Maintain clean, high-quality code with zero blockers from CodeRabbit reviews.

---

## Current State (Post-Sprint 3)

| Metric | Value |
|--------|-------|
| Open PRs | 1 (Sprint 3) |
| CodeRabbit Status | ✅ Passed (17 suggestions) |
| Open Branches | 3 (main, release/v0.7.0-sprint1, sprint3) |
| Local Branches Cleaned | ✅ 15 deleted |

---

## The 17 CodeRabbit Review Items

### 🔴 High Priority (Fix Before Sprint 4)

| # | File | Issue | Why It Matters |
|---|------|-------|----------------|
| 1 | `internal/bundle/oom.go:191-199` | QoS enrichment loop may miss matches due to PodName format mismatch | OOM analysis could show "Unknown" QoS when data exists |
| 2 | `internal/bundle/oom.go:308-310` | Variable shadowing in type assertions | Code maintainability, subtle bugs |
| 3 | `internal/tui/handlers.go:371-376` | Silent error when GetContainers fails | User experience - no feedback on failures |
| 4 | `internal/tui/handlers.go:252-275` | Missing clearPodState() when switching containers | State leakage between container views |
| 5 | `internal/tui/helpers.go:556-575` | Using pod-level restarts instead of per-container | Diagnostics display wrong restart counts |
| 6 | `internal/tui/helpers.go:519-524` | Errors from GetPodResources silently discarded | Missing resource data silently |

### 🟡 Medium Priority (Sprint 4 Early)

| # | File | Issue | Action |
|---|------|-------|--------|
| 7 | `internal/tui/app.go:1020-1026` | CRD batching tick too aggressive (100ms) | Tune to 250-500ms |
| 8 | `internal/tui/app.go:765-774` | processPendingCRDCounts method missing | Implement async CRD count fetching |
| 9 | `internal/tui/app.go:254-256` | processPendingCRDCounts needs implementation | Complete the TODO |
| 10 | `internal/tui/table.go:580-586` | Fixed column widths break terminal-adaptive pattern | Use calculateColumnWidths() |
| 11 | `internal/tui/table.go:569-578` | Empty-state guard checks wrong variable | Check a.containerDetails not a.containers |

### 🟢 Low Priority (Polish)

| # | File | Issue | Action |
|---|------|-------|--------|
| 12 | `Makefile:44-49` | clean target misses coverage artifacts | Add coverage.out coverage.html removal |
| 13 | `ROADMAP_UPDATES.md:273-274` | "Next Review" timestamp in past | Update to future date |
| 14 | `ROADMAP_UPDATES.md:237-258` | Missing language specifier on code block | Add ```text |
| 15 | `scripts/release.sh:59-61` | Only uploads host binary | Add cross-platform builds |
| 16 | `scripts/release.sh:21-34` | Unquoted variables, no branch restore | Quote VERSION, restore original branch |
| 17 | Docstrings | Various missing | Generate docstrings checkbox |

---

## Sprint 4 Workflow Rules

### Before Creating PR

1. **Run local checks:**
   ```bash
   make test
   make build
   make clean && make coverage  # Verify clean removes everything
   ```

2. **File count check:**
   ```bash
   git diff --stat main...HEAD | wc -l  # Must be < 150
   ```

3. **No log bundles:**
   ```bash
   git diff --name-only main...HEAD | grep -c "example-log-bundle"  # Must be 0
   ```

### PR Template (Mandatory)

```markdown
## Sprint 4: [Title]

### Changes
- [ ] Feature 1
- [ ] Feature 2

### CodeRabbit Pre-Checks
- [ ] File count < 150
- [ ] No example-log-bundle files
- [ ] Tests pass
- [ ] Coverage generated

### Related
- Closes #[issue]
- Previous: #25 (Sprint 3)
```

---

## Sprint 4 Proposed Tasks

### S4-CRITICAL-1: Fix High Priority CodeRabbit Items
**Effort:** 4 hours  
**Files:** `oom.go`, `handlers.go`, `helpers.go`  
**Deliverable:** PR with fixes, all 6 items resolved

### S4-HIGH-1: Complete CRD Async Loading
**Effort:** 3 hours  
**Files:** `app.go`  
**Deliverable:** processPendingCRDCounts fully implemented

### S4-HIGH-2: Terminal-Adaptive Column Widths
**Effort:** 2 hours  
**Files:** `table.go`  
**Deliverable:** All tables use calculateColumnWidths()

### S4-MEDIUM-1: Polish Release Script
**Effort:** 2 hours  
**Files:** `scripts/release.sh`, `Makefile`  
**Deliverable:** Cross-platform builds, clean coverage

### S4-MEDIUM-2: Documentation Cleanup
**Effort:** 1 hour  
**Files:** `ROADMAP_UPDATES.md`  
**Deliverable:** Markdown linting fixes

---

## Branch Management Rules

1. **Naming:** `sprint4-[task-id]-brief-description`
2. **Lifetime:** Delete after merge (enforced via cron job)
3. **Max open:** 3 branches per developer
4. **Sync:** `git pull origin main` before each work session

### Cleanup Command

```bash
# Run weekly
git checkout main
git pull origin main
git branch --merged main | grep -v "^\*" | xargs -n 1 git branch -d
git remote prune origin
```

---

## CodeRabbit Response Strategy

### Review Received → Action Matrix

| CodeRabbit Verdict | Action | Timeframe |
|-------------------|--------|-----------|
| 🟢 Approve (0 issues) | Merge immediately | Same day |
| 🟢 Approve (minor issues) | Merge, create follow-up issue | Same day |
| 🟡 Comment (improvements) | Address in branch, re-push | 24 hours |
| 🔴 Changes requested | Fix blockers, re-request review | 24 hours |

### Review Checklist for Each PR

- [ ] No "Too many files!" warnings (< 150)
- [ ] No log bundle pollution
- [ ] All pre-merge checks pass
- [ ] Title check passes
- [ ] Docstring coverage > 80%

---

## Success Metrics for Sprint 4

| Metric | Target |
|--------|--------|
| CodeRabbit blockers | 0 |
| CodeRabbit suggestions per PR | < 5 |
| Re-review cycles | 1 |
| Open branches | ≤ 2 |
| Sprint 3 follow-up issues closed | 6/6 (high priority) |

---

*Created: 2026-02-12*  
*Next Review: 2026-02-19*
