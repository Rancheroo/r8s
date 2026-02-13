# Open Issues Verification Plan

**Date:** 2026-02-13  
**Total Open Issues:** 27 (after closing 2 duplicates)  
**Goal:** Verify which issues are resolved vs still valid

---

## Verification Strategy

### Method 1: Code archaeology (git log)
Search commits for issue references to find "fixes" that weren't linked properly.

### Method 2: Version correlation  
Check if issue was reported before v0.6.8/v0.6.9 and mentions bugs that may be fixed.

### Method 3: Test reproduction
For critical bugs, actually test if reproducible in current build.

---

## Issues Grouped by Verification Method

### 🔴 Method 1: Git Log Search (Find "stealth fixes")

| Issue | Title | Likely Fixed By | Action |
|-------|-------|-----------------|--------|
| **#15** | Dashboard navigation Truth Only violation | Commit `2d61b87` (cursor/visual order fix) | Search git log for "#15" or "navigation" |
| **#17** | *(from commit msg)* | Referenced in commit `2d61b87` | Check if #17 exists or was #15 typo |
| **#11** | Inconsistent height calculations | Check `helpers.go` commits | Search git log for height fixes |
| **#18** | Table alignment breaks with double digits | Check if fixed in v0.6.8+ | Git log search |
| **#19** | Bundle health percentage missing | May be superseded by #34 | Verify #34 covers this |

**Command to run:**
```bash
cd /workspace/r8s && git log --all --oneline --grep="#15\|#11\|#17\|navigation\|height\|alignment" | head -20
```

---

### 🟡 Method 2: Version Correlation (Pre-v0.6.8 Issues)

| Issue | Created | Reported Version | Current Version | Likely Status |
|-------|---------|------------------|-----------------|---------------|
| **#15** | 2026-01-12 | v0.6.2 | v0.6.9 | 🔴 Likely FIXED in v0.6.8.x |
| **#11** | 2026-01-05 | Pre-v0.6.8 | v0.6.9 | 🟡 Unknown - check commits |
| **#12** | 2026-01-05 | Pre-v0.6.8 | v0.6.9 | 🟡 Unknown - check commits |
| **#13** | 2026-01-05 | Pre-v0.6.8 | v0.6.9 | 🟡 Unknown - check commits |
| **#14** | 2026-01-05 | Pre-v0.6.8 | v0.6.9 | 🟡 Unknown - check commits |
| **#16** | 2026-01-12 | Pre-v0.6.8 | v0.6.9 | 🟢 Likely STILL VALID (enhancement) |
| **#18** | 2026-01-13 | Pre-v0.6.8 | v0.6.9 | 🟡 Unknown - alignment fixes may have resolved |
| **#19** | 2026-01-13 | Pre-v0.6.8 | v0.6.9 | 🟡 Check if #34 duplicates this |

**Rule:** Issues from January pre-v0.6.8 likely need re-testing.

---

### 🟢 Method 3: Active Verification (Test Reproduction)

**Critical Issues to Test:**
1. **#15** - Dashboard navigation bug
   - Build current r8s
   - Load example bundle
   - Navigate to attention dashboard item, press Enter
   - Verify correct pod shows

2. **#11** - Height calculations inconsistent
   - Resize terminal
   - Check help text layout in log view
   - Compare with expected behavior

3. **#17, #18** - Dashboard alignment issues  
   - Navigate to items #10+
   - Check if table alignment correct

**Quick Test Commands:**
```bash
cd /workspace/r8s && make build
./bin/r8s ./example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-12-04_09_15_57/
# Then manually test navigation
```

---

## Pragmatic Verification Approach

### Phase 1: Comment First (No Closures Yet)

For each pre-v0.6.8 issue, post a comment:

```
**Status Check:** This issue was reported for v0.6.2/v0.6.6. 
We're now at v0.6.9 with significant navigation fixes.

Can anyone confirm if this is still reproducible in v0.6.9?

If no response in 7 days, we'll assume fixed and close.
```

### Phase 2: Test & Close Confirmed Fixed

Actually test issues that are:
- Critical (#15)
- Easy to verify (#18 alignment)
- Suspected fixed by commit history

### Phase 3: Bulk Close Stale Issues

After 7-day comment period, close issues with:
- Reference to suspected fix commits
- Explanation of why believed fixed
- Invitation to reopen if still reproducible

---

## Recommended Immediate Actions

### 1. Search Git History (Do this first)
```bash
cd /workspace/r8s
git log --all --oneline | grep -E "#15|#11|#12|#13|#14|#17|#18|navigation|cursor|alignment|height"
```

### 2. Build & Test Critical Bug (#15)
```bash
cd /workspace/r8s && make build
./bin/r8s ./example-log-bundle/...
# Test dashboard navigation specifically
```

### 3. Comment on Pre-v0.6.8 Issues
Post status-check comments on:
- #15 (critical navigation)
- #11 (height calculations)  
- #12 (hardcoded height)
- #13 (performance - namespace health)
- #14 (refactor - duplicated logic)
- #18 (table alignment)

### 4. Check for Duplicates
- #19 vs #34 (bundle health indicator)
- May be able to close #19 in favor of #34

---

## Risk Assessment

| Risk | Mitigation |
|------|------------|
| Closing valid issue | Comment first, 7-day wait period |
| Missing stealth fix | Search git logs thoroughly |
| User frustration | Explain why believed fixed |
| Reopened issues | Welcome them, re-triage quickly |

---

## Verification Tracker

| Issue | Method | Status | Git Evidence | Test Result | Decision |
|-------|--------|--------|--------------|-------------|----------|
| #15 | Git + Test | ⏳ | `2d61b87` cursor fix | ⏳ Pending | ⏳ Wait for test |
| #11 | Git | ⏳ | Search needed | ⏳ Pending | ⏳ Research |
| #12 | Git | ⏳ | Search needed | ⏳ Pending | ⏳ Research |
| #13 | Git | ⏳ | Search needed | ⏳ Pending | ⏳ Research |
| #14 | Git | ⏳ | Search needed | ⏳ Pending | ⏳ Research |
| #18 | Git + Test | ⏳ | Check alignment commits | ⏳ Pending | ⏳ Wait for test |
| #19 | Compare | ⏳ | Check #34 overlap | N/A | ⏳ Research |

---

*Ready to execute — need user go-ahead*
*Estimated time: 30 minutes to search git logs + 15 minutes to test #15*
