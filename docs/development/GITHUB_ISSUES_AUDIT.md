# GitHub Issues Audit — r8s Public Repo

**Audit Date:** 2026-02-13  
**Source:** https://github.com/Rancheroo/r8s/issues  
**Status:** 6 Open Issues (all created today)

---

## Summary

All 6 open issues were created today (2026-02-13) as part of Sprint 4 planning and backlog organization. **No stale issues** — all are active and relevant.

| # | Issue | Priority | Status | Sprint Alignment | Action Needed |
|---|-------|----------|--------|------------------|---------------|
| #45 | CI/CD: Fix Go version compatibility | 🔴 **High** | Open | Sprint 4A (S4-CRITICAL-1) | **Active work** — blocks full CI enablement |
| #44 | CI/CD: Re-enable coverage threshold | 🔴 **High** | Open | Sprint 4A (S4-HIGH-5) | **Active work** — part of quality gates |
| #42 | ConfigMaps and HelmCharts Parser | 🟡 Medium | Backlog | Sprint 6 (deferred) | Track, don't action yet |
| #41 | RKE2 Journald Parser | 🟡 Medium | Backlog | Sprint 5+ (deferred) | Track, don't action yet |
| #40 | dmesg OOM Detection | 🟢 Low/Med | Backlog | Sprint 4B Quick Win candidate | **Reprioritize?** — 2h OOM-only version |
| #39 | *(partial data)* | — | Open | Unknown | Need full details |

---

## Detailed Analysis

### 🔴 Sprint 4 Active Work (Address Immediately)

#### Issue #45: CI/CD: Fix Go version compatibility
- **URL:** https://github.com/Rancheroo/r8s/issues/45
- **Created:** 2026-02-13 02:37 UTC
- **Priority:** Medium (blocks full CI)
- **Effort:** 2 hours
- **Problem:** Dependencies require Go 1.24, GitHub Actions only has 1.23.x
- **Options:**
  1. Pin dependencies to Go 1.23 compatible versions
  2. Wait for GitHub Actions Go 1.24 support
  3. Use alternative linter config
- **Sprint Link:** S4-CRITICAL-1 (CI/CD Pipeline)
- **Recommendation:** Option 1 — pin dependencies. 2 hours, unblocks CI.

#### Issue #44: CI/CD: Re-enable coverage threshold enforcement
- **URL:** https://github.com/Rancheroo/r8s/issues/44
- **Created:** 2026-02-13 02:37 UTC
- **Priority:** High
- **Effort:** 4-8 hours
- **Problem:** Coverage check disabled in PR #43, coverage below 50%
- **Tasks:**
  - Measure current coverage (`make coverage`)
  - Add tests to reach 50%
  - Re-enable in `.github/workflows/ci.yml`
- **Sprint Link:** S4-HIGH-5 (50% Coverage Enforcement)
- **Recommendation:** Check current coverage first. May be closer than expected.

---

### 🟡 Backlog Items (Track, Don't Action)

#### Issue #42: ConfigMaps and HelmCharts Parser
- **URL:** https://github.com/Rancheroo/r8s/issues/42
- **Priority:** Low
- **Effort:** 5 hours
- **Sprint:** Sprint 6 (deferred from Sprint 4)
- **Note:** Already correctly deferred — matches our roadmap

#### Issue #41: RKE2 Journald Parser
- **URL:** https://github.com/Rancheroo/r8s/issues/41
- **Priority:** Medium
- **Effort:** 6 hours
- **Sprint:** Sprint 5+ (deferred from Sprint 4)
- **Note:** README #3 promise — complex parser, correct deferral

---

### 🟢 Quick Win Candidate (Reprioritize?)

#### Issue #40: dmesg OOM Detection
- **URL:** https://github.com/Rancheroo/r8s/issues/40
- **Priority:** Low (marked in issue)
- **Effort:** 6 hours (full implementation)
- **Sprint:** Deferred from Sprint 4
- **README Promise:** #2 (System Health)

**⚠️ RECOMMENDATION:** Consider for Sprint 4B Quick Win

Current issue scope: 6 hours (full dmesg analysis)  
**Proposed quick win:** 2 hours (OOM kills only)

| Approach | Effort | Value | Sprint Fit |
|----------|--------|-------|------------|
| Full dmesg (#40 as-is) | 6h | High | Sprint 6 |
| **OOM-only (quick win)** | **2h** | **Medium** | **Sprint 4B** |

**Action:** Comment on #40 offering to:
1. Split into two issues: #40-quick (OOM-only) and #40-full (complete)
2. Take #40-quick for Sprint 4B
3. Keep #40-full for Sprint 6

---

## Recommendations

### Immediate Actions (This Week)

1. **#45 Go Version:** Comment with decision — pin dependencies to Go 1.23
2. **#44 Coverage:** Run `make coverage`, update issue with current %
3. **#40 dmesg:** Comment offering to split into quick win + full version

### Sprint Planning Updates

Add to Sprint 4 plan:
| Issue | Sprint Task | Effort |
|-------|-------------|--------|
| #45 | S4-CRITICAL-1 CI/CD Pipeline | 2h |
| #44 | S4-HIGH-5 50% Coverage | 4-8h |
| #40-quick | S4-QUICK-2 Minimal dmesg | 2h |

### Issue Hygiene

All issues are fresh (today) — no stale issues to close.  
Consider for future:
- Add labels: `sprint-4`, `backlog`, `quick-win`
- Close #40 if we create #40-quick and #40-full as replacements

---

## Cross-Reference with ROADMAP_UPDATES.md

| GitHub Issue | Roadmap Task | Status |
|--------------|--------------|--------|
| #45 Go version | S4-CRITICAL-1 CI/CD | ✅ Linked |
| #44 Coverage | S4-HIGH-5 50% Coverage | ✅ Linked |
| #42 ConfigMaps | ~~S4-MEDIUM-1~~ → Sprint 6 | ✅ Deferred |
| #41 RKE2 journald | ~~S4-HIGH-3~~ → Sprint 5 | ✅ Deferred |
| #40 dmesg | ~~S4-HIGH-2~~ → Sprint 6 | 🟡 Quick win candidate |

---

*Audit Complete*  
*Next Review: Weekly during Sprint 4*
