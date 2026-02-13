# Sprint 5 Kickoff: Quality & Stability

**Date:** 2026-02-13  
**Status:** KICKOFF  
**Previous:** [Sprint 4 Complete](./SPRINT4_COMPLETE.md)

---

## Sprint 5 Goal

**Make CI/CD rock-solid, then resume feature work.**

We have a working CI pipeline now, but with lint and coverage temporarily disabled. Sprint 5's job is to:
1. Fix the underlying issues that made us disable those checks
2. Re-enable them permanently
3. THEN pick up the deferred parser work

---

## Current State: GitHub Issues Board

### P1: CI/CD (Must Fix First)
| Issue | Title | Effort | Blocking |
|-------|-------|--------|----------|
| #44 | CI/CD: Re-enable lint job | 2-4h | ✅ All future PRs |
| #45 | CI/CD: Fix Go version compatibility | 1-2h | ✅ Lint depends on this |
| #35 | S4-HIGH-5: 50% Coverage Enforcement | 4-8h | ✅ Quality gate |

### P2: Feature Work (After CI Clean)
| Issue | Title | Effort | Status |
|-------|-------|--------|--------|
| #39 | BACKLOG-4: PV/PVC/StatefulSet Parser | 4h | Deferred from Sprint 4 |
| #40 | BACKLOG-5: dmesg OOM Detection | 3h | Deferred from Sprint 4 |
| #41 | BACKLOG-6: RKE2 Journald Parser | 3h | Deferred from Sprint 4 |
| #34 | S4-HIGH-4: Bundle Completeness | 4h | Sprint 4 completed, OR move to S5? |

### P3: Follow-ups (Quick Fixes)
| Issue | Title | Effort | Priority |
|-------|-------|--------|----------|
| #37 | Exit code documentation fix | 30min | Low |
| #38 | JSON format debug output fix | 30min | Low |

---

## Sprint 5 Execution Plan

### Phase 1: CI Stabilization (Days 1-2)
**Goal:** Re-enable all CI checks

**Tasks:**
1. **Fix Go version compatibility (#45)**
   - Identify which dep requires Go 1.24
   - Either pin to older version OR wait for GitHub Actions update
   - **Owner:** Whoever can run Go locally
   - **Effort:** 1-2 hours

2. **Fix lint warnings (#44)**
   - Run `golangci-lint run ./...` locally
   - Fix or remove unused functions
   - Re-enable lint job in CI
   - **Owner:** Same as above
   - **Effort:** 2-4 hours

3. **Achieve 50% coverage (#35)**
   - Run `make coverage` to get current %
   - Write tests for uncovered critical paths
   - Re-enable coverage threshold
   - **Owner:** Test-writer
   - **Effort:** 4-8 hours (depends on current coverage)

### Phase 2: Resume Feature Work (Days 3-5)
**Goal:** Pick up deferred parser work

**Tasks:**
4. **PV/PVC/StatefulSet Parser (#39)**
   - Parse PVC/PV/StatefulSet from bundle
   - Enhance TUI to display storage resources
   - **Effort:** 4h

5. **dmesg OOM Detection (#40)**
   - Parse dmesg logs for OOM kills
   - Correlate with pods
   - **Effort:** 3h

6. **RKE2 Journald Parser (#41)**
   - Parse `journald/rke2-server` logs
   - Extract control plane events
   - **Effort:** 3h

---

## Sprint 5 Constraints

### CodeRabbit Limits
- **Max 150 files per PR**
- Split parser work into separate PRs from CI fixes
- File deletions should be their own PR

### Order Matters
1. ✅ Merge CI fixes first (PR passes, main stays green)
2. ✅ Verify CI is fully green on main
3. ✅ Then start feature branches from green main

### Definition of Done
- [ ] Lint job passes on every PR
- [ ] Coverage threshold enforced (50%)
- [ ] All CI checks green (no skips)
- [ ] At least one parser delivered (PV/PVC preferred)

---

## Success Criteria

| Metric | Target |
|--------|--------|
| CI Passing | 100% (no skipped jobs) |
| Test Coverage | ≥50% |
| Lint Warnings | 0 |
| Features Delivered | 1-3 parsers |
| PR Merge Time | <24h after approval |

---

## Risk Mitigation

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Go 1.24 not available in CI | Medium | Pin deps to Go 1.23 compatible versions |
| Coverage gap >20% | Low | Prioritize critical path tests only |
| Parser complexity underestimated | Medium | Start with PVC (simplest), defer others |

---

## Notes

### From ROADMAP_UPDATES.md
The original Sprint 4 had these parsers deferred:
- S4-HIGH-1: Storage (PV/PVC) → **Pick up in Sprint 5**
- S4-HIGH-2: dmesg → **Pick up in Sprint 5**
- S4-HIGH-3: journald → **Pick up in Sprint 5 or Sprint 6**

### CI Lessons Learned
- Don't upgrade Go versions mid-sprint
- Check CI pipeline health BEFORE starting feature work
- Disable failing gates temporarily, fix quickly

---

**Ready to start Phase 1?**

First task: Check current test coverage and lint status locally.
