# GitHub Issues Audit — r8s Public Repo (REVISED)

**Audit Date:** 2026-02-13  
**Source:** https://github.com/Rancheroo/r8s/issues  
**User Reported:** 30 open issues  
**API Accessible:** 14+ issues (API returning subset)

---

## ⚠️ API Limitation Detected

The GitHub API is only returning issues created in the last 2 days (2026-02-12 to 2026-02-13). **16 issues are not visible** via the API query (likely due to pagination defaults or filtering).

**Action Required:** Browser access to see full issue list: https://github.com/Rancheroo/r8s/issues

---

## Issues Accessible via API (14 visible)

### 🔴 Sprint 4 Active Work (6 issues)

| Issue | Title | Created | Sprint Link | Status |
|-------|-------|---------|-------------|--------|
| **#45** | CI/CD: Fix Go version compatibility | 2026-02-13 | S4-CRITICAL-1 | 🔴 Active — blocks CI |
| **#44** | CI/CD: Re-enable coverage threshold | 2026-02-13 | S4-HIGH-5 | 🔴 Active — quality gate |
| **#35** | S4-HIGH-5: 50% Coverage Enforcement | 2026-02-12 | S4-HIGH-5 | 🟡 Duplicate of #44? |
| **#34** | S4-HIGH-4: Bundle Completeness Indicator | 2026-02-12 | S4-HIGH-4 | 🟢 Planned |
| **#38** | Follow-up: Fix JSON format debug output | 2026-02-13 | Sprint 3 | 🟢 Low priority — quick fix |
| **#37** | Follow-up: Fix exit code documentation | 2026-02-13 | Sprint 3 | 🟢 Low priority — doc fix |

### 🟡 Backlog Items (4 issues)

| Issue | Title | Created | Sprint Target |
|-------|-------|---------|---------------|
| **#42** | BACKLOG-7: ConfigMaps and HelmCharts | 2026-02-13 | Sprint 6 |
| **#41** | BACKLOG-6: RKE2 Journald Parser | 2026-02-13 | Sprint 5+ |
| **#40** | BACKLOG-5: dmesg OOM Detection | 2026-02-13 | Sprint 4B? |
| **#39** | BACKLOG-4: PV/PVC/StatefulSet Parser | 2026-02-13 | Sprint 6 |

### 🟢 Quick Win Candidates (4 issues)

| Issue | Title | Created | Effort | Sprint Fit |
|-------|-------|---------|--------|------------|
| **#30** | Add docstrings to key functions | 2026-02-12 | 2h | Sprint 4B |
| **#29** | Release script: Fix shell robustness | 2026-02-12 | 1h | Sprint 4B |
| **#28** | Release script: Add cross-platform builds | 2026-02-12 | 2h | Backlog (v0.8.0) |
| **#24** | Improve error handling in syntheticDataSource | Earlier | TBD | Needs review |

---

## 🔍 Issues Requiring Manual Review (Not in API)

**Status:** 16 issues not accessible via API  
**Likely Cause:** GitHub API pagination / rate limiting / age filtering

**Manual Check Required For:**
1. Issues older than 2026-02-12
2. Issues with specific labels
3. Stale issues that may need closure
4. Duplicates that should be merged
5. Feature requests that overlap with README promises

---

## Immediate Observations

### Possible Duplicates
- **#44** and **#35** both cover "50% Coverage Enforcement"
  - #44: More detailed (created 2026-02-13 02:37)
  - #35: Sprint-tagged (created 2026-02-12 11:50)
  - **Action:** Close #35, keep #44 (more current/context)

### Quick Win Opportunities
| Issue | Why Quick Win | Sprint 4B Fit |
|-------|---------------|---------------|
| #30 Docstrings | 2h, CodeRabbit flagged | ✅ Yes |
| #29 Release script | 1h, shell quote fixes | ✅ Yes |
| #38 JSON format | Low, debug→stderr | ✅ Yes |
| #37 Exit code docs | Low, doc update | ✅ Yes |

### Backlog Alignment Check
| Issue | README Promise | Sprint Target | Aligned? |
|-------|---------------|---------------|----------|
| #39 PV/PVC | Promise #1 | Sprint 6 | ✅ Yes |
| #40 dmesg | Promise #2 | Sprint 4B? | 🟡 Consider quick win |
| #41 RKE2 journald | Promise #3 | Sprint 5+ | ✅ Yes |
| #42 ConfigMaps | Promise #4 | Sprint 6 | ✅ Yes |

---

## Recommended Actions

### This Week (Sprint 4)

1. **#45 Go Version** — Comment with fix ETA (2h)
2. **#44 Coverage** — Check `make coverage`, update with current %
3. **#35** — Close as duplicate of #44
4. **#40** — Comment offering to split into quick win (OOM-only)

### Sprint 4B Quick Wins (Add These)

| Issue | Task | Effort |
|-------|------|--------|
| #30 | Add docstrings | 2h |
| #29 | Fix release script | 1h |
| #38 | Fix JSON debug output | 1h |
| #37 | Fix exit code docs | 0.5h |

**Total:** 4.5h — fits Sprint 4b budget

### Manual Review Needed (16 Hidden Issues)

**Required:** Browser access to https://github.com/Rancheroo/r8s/issues to:
- [ ] Count total open issues (confirm 30)
- [ ] Identify stale issues (>3 months)
- [ ] Find duplicates
- [ ] Check for critical bugs
- [ ] Review feature requests against README promises

---

## Tool Limitation Note

**Why API shows only 14:**
- GitHub API by default shows issues updated recently
- Older issues may not appear in `/issues?state=open`
- Some issues may have "pull request" associations filtering them out
- Default pagination limits to 30 per page (we requested 100)

**Workaround:**
- Direct browser access needed
- Or use `gh` CLI with authentication
- Or use GitHub GraphQL API with explicit pagination

---

*Partial Audit — API Limited*  
*Next Step: Manual browser review of remaining 16 issues*
