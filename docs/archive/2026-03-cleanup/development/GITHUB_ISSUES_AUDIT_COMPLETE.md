# GitHub Issues Audit — r8s Public Repo (COMPLETE)

**Audit Date:** 2026-02-13  
**Source:** GitHub CLI (`gh issue list`)  
**Total Open Issues:** 29
**Authenticated as:** Rancheroo

---

## 📊 Issue Summary by Category

| Category | Count | Issue Range | Priority |
|----------|-------|-------------|----------|
| Sprint 4 Active | 6 | #34, #35, #37, #38, #44, #45 | 🔴 High |
| Backlog (Sprint 5+) | 4 | #39-#42 | 🟡 Medium |
| Critical Bugs | 1 | #15 | 🔴 CRITICAL |
| Performance/Refactor | 4 | #11, #13, #14, #24 | 🟡 Medium |
| UI/UX Bugs | 4 | #16-#19 | 🟢 Low |
| Sprint 2 Follow-ups | 3 | #22, #23 | 🟢 Low |
| Cleanup/Tech Debt | 7 | #6-#10, #28-#30 | 🟢 Low |

---

## 🔴 CRITICAL: Immediate Attention Required

### Issue #15: Dashboard navigation Truth Only violation persists
- **Created:** 2026-01-12
- **Severity:** 🔴 **CRITICAL** — Truth Only™ violation
- **Problem:** Dashboard shows wrong pod diagnostics when navigating
- **Impact:** Users cannot trust dashboard navigation — shows wrong data
- **Attempted Fix:** v0.6.2 commit 3765576 did not resolve
- **Root Cause:** Cursor/table/viewport state desync during sorting
- **Recommendation:** 
  - **Option A:** Include in Sprint 4B (if we can fix quickly)
  - **Option B:** Document as known issue, fix in v0.7.0 with architecture rewrite
- **Effort:** 4-8 hours (requires architectural understanding)

**Musk's Law:** 🧠 QUESTION — "Is this the highest-impact fix we can ship?"

---

## 🔴 Sprint 4 Active Work (6 issues)

| # | Title | Created | Sprint Link | Quick Action |
|---|-------|---------|-------------|--------------|
| **#45** | CI/CD: Fix Go version compatibility | 2026-02-13 | S4-CRITICAL-1 | Comment: Pin deps to Go 1.23 |
| **#44** | CI/CD: Re-enable coverage threshold | 2026-02-13 | S4-HIGH-5 | Run `make coverage`, update |
| **#35** | S4-HIGH-5: 50% Coverage Enforcement | 2026-02-12 | S4-HIGH-5 | **CLOSE** — duplicate of #44 |
| **#34** | S4-HIGH-4: Bundle Completeness Indicator | 2026-02-12 | S4-HIGH-4 | On track |
| **#38** | Follow-up: Fix JSON format debug output | 2026-02-13 | Sprint 3 | Quick win — 0.5h |
| **#37** | Follow-up: Fix exit code documentation | 2026-02-13 | Sprint 3 | Quick win — 0.5h |

---

## 🟡 Backlog Items (4 issues) — README Promise Alignment

| # | Title | README Promise | Sprint Target | Status |
|---|-------|---------------|---------------|--------|
| **#39** | PV/PVC/StatefulSet Parser | Promise #1 | Sprint 6 | ✅ Aligned |
| **#40** | dmesg OOM Detection | Promise #2 | Sprint 4B? | 🟡 Consider quick win |
| **#41** | RKE2 Journald Parser | Promise #3 | Sprint 5+ | ✅ Aligned |
| **#42** | ConfigMaps and HelmCharts | Promise #4 | Sprint 6 | ✅ Aligned |

**Recommendation:** Offer to split #40 into #40-quick (OOM-only, 2h) for Sprint 4B

---

## 🟢 Quick Win Candidates (11 issues — 9.5h total)

| # | Title | Effort | Sprint 4B Fit |
|---|-------|--------|---------------|
| **#29** | Release script: Fix shell robustness | 0.5h | ✅ Yes |
| **#30** | Add docstrings to key functions | 2h | ✅ Yes |
| **#37** | Fix exit code documentation | 0.5h | ✅ Yes |
| **#38** | Fix JSON format debug output | 0.5h | ✅ Yes |
| **#9** | Fix markdown formatting in RELEASE_ROADMAP | 0.5h | ✅ Yes |
| **#10** | Replace hard tabs in V0.5.1_COMPLETION_GUIDE | 0.5h | ✅ Yes |
| **#22** | Fix pre-commit hook script issues (PR #21) | 1h | ✅ Yes |
| **#23** | Fix pre-commit hook script issues (PR #21) | 1h | ✅ Yes |
| **#24** | Improve error handling in syntheticDataSource | 2h | ✅ Yes |
| **#16** | Enhance example bundle (test scenarios) | 1h | ✅ Yes |
| **#7** | Markdown linting issues | 0.5h | ✅ Yes |

**Selected for Sprint 4B:** #29, #30, #37, #38 (4 issues, 4h)  
**Deferred to Sprint 6:** #9, #10, #16, #22, #23, #24, #7

---

## 🟡 Performance/Refactoring (4 issues)

| # | Title | Type | Action |
|---|-------|------|--------|
| **#11** | Inconsistent height calculations in helpers.go | Bug | Sprint 5 — UI alignment |
| **#13** | ComputeNamespaceHealth called on every render | Performance | Sprint 5 — caching |
| **#14** | Duplicated namespace extraction logic in table.go | Refactor | Sprint 5 — cleanup |
| **#24** | Improve error handling in syntheticDataSource | Enhancement | Sprint 4B — quick win |

---

## 🟢 UI/UX Bugs (4 issues)

| # | Title | Severity | Sprint |
|---|-------|----------|--------|
| **#17** | Dashboard navigation for kubelet items shows wrong panel | Medium | Sprint 5 |
| **#18** | Dashboard table alignment breaks with double-digit numbers | Low | Sprint 5 |
| **#19** | Bundle health percentage missing from status bar | Low | Sprint 4 (via #34) |

---

## 🔗 Duplicate / Merge Candidates

### Group 1: Coverage Enforcement (Duplicates)
- **#35** S4-HIGH-5: 50% Coverage Enforcement
- **#44** CI/CD: Re-enable coverage threshold enforcement

**Recommendation:** CLOSE #35, keep #44 (more detailed, newer context)

### Group 2: Pre-commit Hook Issues (Duplicates)
- **#22** Fix pre-commit hook script issues identified in PR #21  
- **#23** Fix pre-commit hook script issues identified in PR #21

**Recommendation:** CLOSE #23 as duplicate of #22

---

## 📋 Immediate Actions (This Week)

### Must Do (Critical Path)
1. **#15** — Decide: Fix in Sprint 4B or document as known issue?
2. **#45** — Comment with Go version fix ETA
3. **#44** — Run `make coverage`, update issue with current %
4. **#35** — Close as duplicate of #44
5. **#23** — Close as duplicate of #22

### Should Do (Quick Wins)
6. **#40** — Comment offering OOM-only quick win (2h vs 6h)
7. **#29, #30, #37, #38** — Add to Sprint 4B plan

### Can Wait (Backlog)
8. **#9, #10, #16** — Document cleanup for Sprint 6
9. **#11, #13, #14, #17, #18** — Sprint 5 performance/refactor bundle

---

## Sprint 4B Revised Plan (Including Quick Wins)

| ID | Issue | Task | Effort |
|----|-------|------|--------|
| S4B-1 | #29 | Release script shell fixes | 0.5h |
| S4B-2 | #30 | Add docstrings | 2h |
| S4B-3 | #37 | Exit code documentation | 0.5h |
| S4B-4 | #38 | JSON debug → stderr | 0.5h |
| S4B-5 | #40-quick | dmesg OOM only (if accepted) | 2h |

**Total:** 4-6 hours  
**Plus:** Decide on #15 (Critical navigation bug)

---

## Musk's Laws Applied

| Law | Application |
|-----|-------------|
| 🧠 **QUESTION** | Is #15 worth 4-8h now, or document and fix in v0.7.0 architecture? |
| 🗑️ **DELETE** | Close #35, #23 as duplicates — don't track twice |
| ⚙️ **SIMPLIFY** | Split #40 into OOM-only quick win |
| 🚀 **ACCELERATE** | Bundle #29, #30, #37, #38 into Sprint 4B |
| 🤖 **AUTOMATE** | Use `gh` CLI for issue management (now working!) |

---

## GH CLI Command Reference

```bash
# List all open issues
export GH_TOKEN="$GITHUB_TOKEN"
cd /workspace/r8s && gh issue list --repo Rancheroo/r8s --state open --limit 100

# Close duplicates
gh issue close 35 --comment "Closing as duplicate of #44 (newer, more detailed)"
gh issue close 23 --comment "Closing as duplicate of #22"

# Comment on issues
gh issue comment 45 --body "Fix: Pinning dependencies to Go 1.23 compatible versions. ETA: 2 hours."
gh issue comment 40 --body "Proposing quick win: OOM-only parser (2h). Keep full version for Sprint 6?"
```

---

*Audit Complete*  
*Next: User decision on #15 (critical bug) and permission to close duplicates*
