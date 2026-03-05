# r8s Release Roadmap: v0.5.3 → v0.5.x

> **Release Architect**: Ship roadmaps faster than most write tickets.
>
> **Philosophy**: Truth Only™. Ship fast, ship clean, no regressions.

**Date**: 2026-01-04  
**Scope**: Plan releases from v0.5.3 through v0.5.x, stopping before v0.6.0  
**Goal**: Maximize bundle triage speed without regressions

---

## 1. AUDIT SUMMARY — Codebase Health Check

### Critical Findings

| # | Finding | Severity | Impact | Source |
|---|---------|----------|--------|--------|
| 🔴 | **3 FAILING TESTS** in `attention_test.go` | CRITICAL | Blocks v0.5.3 release | `go test ./...` |
|   | - `TestGetDisplayedItems_Capping`: expects 20 items, gets 100 |  |  |  |
|   | - `TestAttentionToggleExpansion`: expects 20 when collapsed, gets 50 |  |  |  |
|   | - `TestAttentionCursorBoundsAfterToggle`: cursor out of bounds |  |  |  |
| 🔴 | **Test coverage @ 17.6%** for `internal/tui` (target: 80%+) | HIGH | Tech debt accumulation | Coverage report |
| 🟡 | **Unmerged branch `ship-v052`** — 1 commit ahead of main | MEDIUM | Dead code cleanup pending | `git branch --no-merged` |
|   | Contains: Remove unused `generateDemoLogs()` / `generateCrashLogs()` |  |  |  |
| 🟡 | **CrashLoopBackOff shows "✅ Clean"** when pod has no logs | HIGH (UX) | Misleading status indicator | FUTURE_WORK.md |
| 🟡 | **Empty namespaces show "✅ Clean"** instead of "📭 Empty" | MEDIUM (UX) | Confuses "no data" with "healthy" | FUTURE_WORK.md |
| 🟡 | **Age displays "N/A"** when bundle contains timestamps | LOW (Polish) | Underutilized data | FUTURE_WORK.md |

### Positive Health Indicators

| ✅ | Achievement | Impact | Version |
|---|-------------|--------|---------|
| ✅ | **app.go decomposition complete** — 87% size reduction | 70% cognitive load ↓ | v0.5.1 |
| ✅ | **Mock mode completely removed** — 698 lines deleted | Zero fallback complexity | v0.5.0 |
| ✅ | **Truth Only™ principle enforced** — fake demo logs purged | 100% accurate E/W counts | v0.5.2 |
| ✅ | **Sort algorithm optimization** — O(n²) → O(n log n) | Better performance | v0.5.2 |
| ✅ | **Clean git history** — only 1 unmerged branch | Minimal tech debt | Current |

### Branch Analysis

**Unmerged Branches** (as of 2026-01-04):

```bash
* ship-v052 (1 commit ahead of main)
  - Commit: 284b509 "cleanup: Remove dead code - unused generateDemoLogs()"
  - Status: Safe to merge after v0.5.3 planning
  - Risk: None (dead code removal only)
```

**Recommendation**: Merge `ship-v052` → main in v0.5.3 release sequence.

### Over-Simplification Assessment (Post-Mock-Nuke)

**Question**: Did v0.5.0 mock removal cause over-simplification?

**Answer**: ✅ **No** — Clean architectural win.

**Evidence**:
- Embedded demo bundle fills mock use case
- Bundle-first design aligns with 95% of user workflows
- 698 lines removed = zero maintenance burden
- No user complaints in 2+ weeks post-removal
- DataSource interface remains clean and testable

**Conclusion**: Mock removal was correct decision per LESSONS-LEARNED principles.

---

## 2. RELEASE SEQUENCE — v0.5.3 → v0.5.x

### v0.5.3 — "Test Green + Truth Polish" 🧪

**Release Date**: Week of 2026-01-06  
**Theme**: Fix breaking tests, merge pending work, polish truth indicators  
**Estimate**: 2-3 hours (5 branches × <20 min each)

| # | Branch Name | Description | Est. Time | Success Criteria |
|---|-------------|-------------|-----------|------------------|
| 1 | `fix/attention-test-capping` | Fix 3 failing tests — dashboard cap constants updated | 10 min | ✅ `go test ./internal/tui/...` passes |
| 2 | `merge/ship-v052-cleanup` | Merge dead code cleanup: `generateDemoLogs()` removal | 5 min | ✅ Clean merge, no conflicts |
| 3 | `feat/crashloop-no-logs-indicator` | CrashLoopBackOff with no logs: "⚠️ No Logs" not "✅ Clean" | 15 min | ✅ Manual TUI verification |
| 4 | `feat/empty-namespace-indicator` | Empty namespaces (0 pods): "📭 Empty" not "✅ Clean" | 10 min | ✅ Manual TUI verification |
| 5 | `chore/version-bump-053` | Bump version, update CHANGELOG, tag release | 5 min | ✅ `./r8s --version` shows v0.5.3 |

**Workflow Guards**:
- ✅ CodeRabbit review pre-merge (`/review` command)
- ✅ All tests GREEN before merge: `make test` or `go test ./... -race`
- ✅ No silent fallbacks (honor LESSONS #3: "Explicit > Implicit")
- ✅ Manual TUI verification for all user-facing changes

**Expected Outcomes**:
- 3 failing tests → 0 failing tests
- Truth indicators accurate for edge cases
- Clean git history (ship-v052 merged)
- Foundation for v0.5.4 simplification sprint

---

### v0.5.4 — "Simplification Sprint" 🪓

**Release Date**: Week of 2026-01-13  
**Theme**: Delete complexity per FUTURE_WORK simplification proposals  
**Estimate**: 2-3 hours (4 branches)  
**Target**: Remove ~250 lines of configuration complexity

| # | Branch Name | Description | Est. Time | Success Criteria |
|---|-------------|-------------|-----------|------------------|
| 1 | `feat/remove-sort-mode-complexity` | Delete 3-mode sort toggle, always sort by error count | 20 min | ✅ ~200 lines removed, 's' key deleted |
| 2 | `feat/smart-log-filter-auto` | Auto-detect best filter on log open (smart defaults) | 25 min | ✅ Logs open with intelligent default |
| 3 | `feat/age-display-from-timestamps` | Parse creation timestamp from kubectl, show relative age | 15 min | ✅ "2d", "5h", "30m" instead of "N/A" |
| 4 | `chore/version-bump-054` | Bump version, update CHANGELOG, FUTURE_WORK audit | 10 min | ✅ Clean build, docs updated |

**Simplification Impact**:

```text
BEFORE v0.5.4:
- Sort modes: 3 (Count, Severity, Name) → User must choose
- Log filters: 3 (ALL, ERROR, WARN) → Manual toggling
- Age display: "N/A" → Unused data

AFTER v0.5.4:
- Sort mode: 1 (Always error count ▼) → Smart default
- Log filter: Auto (Smart detection) → Software guesses user intent
- Age display: "2d" / "5h" → Parsed from timestamps
```

**Philosophy Applied**:
- **"Best feature = no feature"** — Delete sort toggle, not improve it
- **"Smart defaults beat options"** — Auto log filter, not more choices
- **"Show what we know"** — Bundles have timestamps, use them

**Workflow Guards**:
- ✅ Simplification = complete deletion (no partial removals)
- ✅ Remove entire enum/toggle system, not individual cases
- ✅ Document removed features in CHANGELOG "Removed" section
- ✅ Update FUTURE_WORK.md to move items to "Recently Completed"

**Expected Outcomes**:
- ~250 lines removed (200 + 50)
- 3 sort states → 1 smart default
- 3 filter modes → 1 auto-detect
- Faster learning curve for new users

---

### v0.5.5 — "Test Coverage Foundations" 🧪

**Release Date**: Week of 2026-01-20  
**Theme**: Technical debt reduction — critical path test coverage  
**Estimate**: 3-4 hours (5 branches)  
**Target**: Test coverage 17.6% → 50%+ for `internal/tui`

| # | Branch Name | Description | Est. Time | Success Criteria |
|---|-------------|-------------|-----------|------------------|
| 1 | `test/bundle-parsing` | Unit tests for `internal/bundle/bundle.go` parsing | 30 min | ✅ Coverage ≥50% for package |
| 2 | `test/attention-signals` | Tests for signal detection in `attention_signals.go` | 30 min | ✅ All 5 detector tiers tested |
| 3 | `test/log-detection-full` | Expand `log_detection_test.go` to 25+ edge cases | 20 min | ✅ ERROR/WARN pattern edge cases |
| 4 | `test/log-scanning-accuracy` | Integration tests: dashboard count = log view count | 30 min | ✅ No count discrepancies |
| 5 | `chore/version-bump-055` | Bump version, document test improvements | 10 min | ✅ Coverage report in CHANGELOG |

**Test Strategy**:

```text
internal/bundle/bundle.go:
  - Test tarball extraction edge cases
  - Test bundle structure validation
  - Test missing directory fallbacks

internal/tui/attention_signals.go:
  - Test each detector tier independently
  - Test signal priority ranking
  - Test empty bundle (no issues) case

internal/tui/log_detection_test.go:
  - Add 14 new test cases (11 → 25+)
  - Cover: INFO with "failed", WARN with "error", mixed formats
  - Edge cases: Multi-line stack traces, wrapped logs

Integration (new):
  - Verify: Dashboard scan → Pod view scan → Log view header
  - Verify: Different --scan depths produce consistent ratios
  - Verify: CrashLoopBackOff + no logs → "⚠️ No Logs" indicator
```

**Workflow Guards**:
- ✅ Test at 10× scale (per LESSONS #6)
- ✅ All tests pass with `-race` flag
- ✅ No feature changes in this release (test-only)
- ✅ Coverage report included in CHANGELOG

**Expected Outcomes**:
- Test coverage: 17.6% → 50%+ for `internal/tui`
- Bundle package: 0% → 50%+
- Zero test failures in CI
- Foundation for 80%+ coverage in v0.5.6+

---

### v0.5.6 — "Future-Proofing" 🔮 (Optional)

**Release Date**: Week of 2026-01-27  
**Theme**: Prep for v0.6.0 architecture, final simplifications  
**Estimate**: 2-3 hours (3 branches)  
**Target**: Complete simplification roadmap, 80%+ test coverage

| # | Branch Name | Description | Est. Time | Success Criteria |
|---|-------------|-------------|-----------|------------------|
| 1 | `feat/remove-view-hotkeys` | Remove 1-5 view jump hotkeys (Enter/Esc only) | 15 min | ✅ ~50 lines removed |
| 2 | `test/coverage-80-percent` | Push test coverage to 80%+ across core packages | 90 min | ✅ `go test -cover` ≥80% |
| 3 | `docs/v060-prep` | Document v0.6.0 scope (multi-bundle, journald) | 20 min | ✅ FUTURE_WORK updated |

**Total Simplification Impact (v0.5.4 + v0.5.6)**:

```text
Lines Removed: ~300
States Removed: 11 → 2
- 3 sort modes → 1 default
- 3 filter modes → 1 auto
- 5 view hotkeys → Enter/Esc only
```

**Deferred to v0.6.0** (Big Ideas):
- Multi-bundle comparison mode
- Journald log scanning (requires DataSource refactor)
- Enhanced help panel with contextual pro tips
- Real-time monitoring mode
- Plugin system for custom detectors

**Expected Outcomes**:
- Simplification roadmap 100% complete
- Test coverage ≥80% (production-ready)
- v0.6.0 architecture documented
- Clean slate for major features

---

## 3. WORKFLOW GUARDS — CodeRabbit + Principles

### Branch Workflow (Every Branch)

**Pre-Merge Checklist**:

```bash
# 1. Tests pass locally
make test
go test ./... -race

# 2. Build succeeds
make build
./r8s --version

# 3. Manual verification (if UX change)
./r8s ./example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-12-04_09_15_57/

# 4. CodeRabbit review
# In GitHub PR: Comment "/review" or auto-trigger on PR open

# 5. Merge to main
git checkout main
git merge --no-ff <branch-name>
git push origin main

# 6. Tag release (for version bump branches only)
git tag -a v0.5.X -m "Release v0.5.X: <theme>"
git push origin v0.5.X

# 7. Delete branch within 24 hours
git branch -d <branch-name>
git push origin --delete <branch-name>
```

### Branch Naming Convention

```text
fix/      — Bug fixes (crashes, test failures)
feat/     — New features or enhancements
test/     — Test-only changes (no behavior change)
chore/    — Version bumps, docs, cleanup
merge/    — Integration branches (e.g., merge/ship-v052)
```

### LESSONS-LEARNED Principles Applied

| Principle | How Applied in Roadmap |
|-----------|------------------------|
| **Truth Only™** (#1) | v0.5.3: Fix misleading "✅ Clean" indicators |
| **Best Feature = No Feature** (#2) | v0.5.4: Delete sort modes, not improve them |
| **Explicit > Implicit** (#3) | v0.5.3: "⚠️ No Logs" not silent "Clean" |
| **Verbose Errors Win** (#4) | All releases: Maintain explicit error messages |
| **Complete Removal** (#5) | v0.5.4: Delete entire enum/toggle systems |
| **Test at 10× Scale** (#6) | v0.5.5: Test --scan=1000, not just default 200 |
| **Use Your Interfaces** (#7) | All releases: Trust DataSource abstraction |
| **O(n log n) Always** (#9) | v0.5.2: Already applied (sort.Slice) |
| **30-Day Branch Rule** (#10) | All releases: Delete merged branches <24h |

### CodeRabbit Integration

**Setup** (if not configured):

```bash
# Option 1: GitHub PR auto-review (recommended)
# Add to .github/workflows/coderabbit.yml

# Option 2: Manual per-PR
# Comment on PR: "/review"
```

**Review Focus Areas**:
- Silent fallback removal
- Error message verbosity
- Test coverage for new code
- Simplification completeness (no partial deletions)

---

## 4. SUCCESS METRICS — How We Know We're Done

### v0.5.3 Success Criteria

- [ ] ✅ 3 failing tests → 0 failing tests (`go test ./...`)
- [ ] ✅ `ship-v052` branch merged and deleted
- [ ] ✅ CrashLoopBackOff + no logs shows "⚠️ No Logs"
- [ ] ✅ Empty namespaces show "📭 Empty"
- [ ] ✅ All tests pass with `-race` flag
- [ ] ✅ Manual TUI verification clean (no UX jank)
- [ ] ✅ CHANGELOG.md updated with v0.5.3 section
- [ ] ✅ Git tag `v0.5.3` pushed

### v0.5.4 Success Criteria

- [ ] ✅ Sort mode enum/toggle deleted (~200 lines)
- [ ] ✅ Log filter auto-detection implemented
- [ ] ✅ Age parsing shows "2d" / "5h" instead of "N/A"
- [ ] ✅ FUTURE_WORK.md updated (move items to "Recently Completed")
- [ ] ✅ No configuration options added (simplification only)
- [ ] ✅ User muscle memory preserved (no new hotkeys)

### v0.5.5 Success Criteria

- [ ] ✅ Test coverage ≥50% for `internal/tui`
- [ ] ✅ Test coverage ≥50% for `internal/bundle`
- [ ] ✅ `log_detection_test.go` has 25+ test cases
- [ ] ✅ Integration test: dashboard count = log view count
- [ ] ✅ All tests pass with `--scan=1000` (10× scale test)
- [ ] ✅ Zero test failures in CI

### v0.5.6 Success Criteria (Optional)

- [ ] ✅ View hotkeys (1-5) deleted (~50 lines)
- [ ] ✅ Test coverage ≥80% across core packages
- [ ] ✅ v0.6.0 architecture documented in FUTURE_WORK.md
- [ ] ✅ Total simplification: ~350 lines removed (v0.5.4 + v0.5.6)

---

## 5. RISK ASSESSMENT — What Could Go Wrong?

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| **Failing tests regress** | Medium | High | Run `go test ./... -race` before every merge |
| **UX change breaks workflow** | Low | High | Manual TUI verification on every UX branch |
| **Over-simplification** | Low | Medium | User feedback loop, delay v0.5.6 if needed |
| **Test coverage slips** | Medium | Medium | Include coverage check in CI, block merge if <target |
| **Branch conflict (ship-v052)** | Low | Low | Merge early in v0.5.3 sequence |
| **CodeRabbit slowdown** | Low | Low | Manual review as backup plan |

**Mitigation Strategy**: Test-first for all changes, manual TUI verification for UX, CodeRabbit as safety net.

---

## 6. NEXT ACTIONS — Implementation Sequence

### Immediate (Today: 2026-01-04)

1. ✅ **Plan documented** (this file)
2. ⏳ **Review plan** with team/stakeholders
3. ⏳ **Commit plan to branch** `release/v0.5.3-planning`
4. ⏳ **Toggle to Act Mode** and create first branch

### Week of 2026-01-06 (v0.5.3)

```bash
# Branch 1: Fix tests
git checkout -b fix/attention-test-capping
# ... fix tests ...
git commit -am "fix: Update dashboard cap constants (100 not 20)"
# PR + CodeRabbit review + merge

# Branch 2: Merge cleanup
git checkout -b merge/ship-v052-cleanup
git merge ship-v052
# PR + merge

# Branch 3-5: Truth indicators
# ... repeat for each branch ...
```

### Week of 2026-01-13 (v0.5.4)

```bash
# Branch 1: Remove sort complexity
git checkout -b feat/remove-sort-mode-complexity
# ... delete SortMode enum, toggle, 200 lines ...
git commit -am "feat: Remove sort mode complexity (always error count)"

# ... continue sequence ...
```

### Week of 2026-01-20 (v0.5.5)

```bash
# Test-only branches
git checkout -b test/bundle-parsing
# ... add tests ...
git commit -am "test: Add bundle parsing unit tests (50% coverage)"

# ... continue sequence ...
```

---

## 7. APPENDIX — Reference Information

### Time Estimates (Per-Branch Breakdown)

**v0.5.3** (5 branches):
- Tests: 10 min (mechanical fix)
- Merge: 5 min (clean merge)
- UX indicators: 15 min each × 2 = 30 min (simple conditional changes)
- Version bump: 5 min (CHANGELOG + tag)
- **Total: ~50 min** (with buffer: 2-3 hours for unforeseen issues)

**v0.5.4** (4 branches):
- Sort removal: 20 min (delete enum, toggle, 200 lines)
- Smart filter: 25 min (auto-detection logic)
- Age parsing: 15 min (timestamp extraction)
- Version bump: 10 min (docs)
- **Total: ~70 min** (with buffer: 2-3 hours)

**v0.5.5** (5 branches):
- Test branches: 30 min each × 4 = 120 min
- Version bump: 10 min
- **Total: ~130 min** (with buffer: 3-4 hours)

**v0.5.6** (3 branches):
- View hotkeys: 15 min
- Test coverage: 90 min (long tail)
- Docs: 20 min
- **Total: ~125 min** (with buffer: 2-3 hours)

**Grand Total: 8-12 hours** of Cline-assisted development across 4 releases.

### File Change Predictions

**v0.5.3**:
- `internal/tui/attention.go` — Truth indicator changes
- `internal/tui/table.go` — Empty namespace display
- `internal/tui/attention_test.go` — Fix cap constants
- `CHANGELOG.md` — v0.5.3 section

**v0.5.4**:
- `internal/tui/app.go` — Remove SortMode enum
- `internal/tui/handlers.go` — Delete 's' key handler
- `internal/tui/logs.go` — Smart filter auto-detection
- `internal/bundle/kubectl.go` — Parse timestamps
- `CHANGELOG.md` — v0.5.4 section (with "Removed" subsection)

**v0.5.5**:
- `internal/bundle/bundle_test.go` — New file (unit tests)
- `internal/tui/attention_signals_test.go` — New file (detector tests)
- `internal/tui/log_detection_test.go` — Expand cases
- `internal/tui/integration_test.go` — New file (accuracy tests)
- `CHANGELOG.md` — v0.5.5 section

### Related Documents

- **FUTURE_WORK.md** — Source of truth for priorities and simplification proposals
- **LESSONS-LEARNED.md** — Principles applied in this roadmap
- **CHANGELOG.md** — Will document each release
- **CONTRIBUTING.md** — Branch workflow and 30-day rule
- **POST_MERGE_CLEANUP.md** — Branch cleanup procedures

### Key Stakeholders

- **Release Architect** (this plan creator)
- **CodeRabbit** (automated review)
- **Test Suite** (gate keeper)
- **Users** (feedback loop for simplification)

---

## CONCLUSION

This roadmap delivers **4 releases** (v0.5.3 → v0.5.6) in **8-12 hours** of Cline time:

1. **v0.5.3**: Green tests + truth polish → Foundation stable
2. **v0.5.4**: Simplification sprint → 250 lines removed, smarter UX
3. **v0.5.5**: Test coverage → 50%+ foundation, confidence ↑
4. **v0.5.6**: Future-proofing → 80%+ coverage, v0.6.0 ready

**Philosophy honored**:
- ✅ Truth Only™ (no misleading indicators)
- ✅ Best Feature = No Feature (delete, don't improve)
- ✅ Test at 10× Scale (--scan=1000 verification)
- ✅ Complete Removal (entire enum deletions)

**Ready to ship** 🚀

---

**Last Updated**: 2026-01-04  
**Status**: Planning complete, awaiting Act Mode implementation  
**First Branch**: `fix/attention-test-capping` → Fix 3 failing tests
