# Sprint 10 Verification Test Plan

**Branch:** `feature/sprint10-ci-cleanup`  
**Commit:** `3c46ee8`  
**Goal:** Verify TUI Phase 1 deletion didn't break anything  
**Estimated Time:** 10 minutes

---

## Pre-Test Checklist

```bash
# 1. Switch to branch
git checkout feature/sprint10-ci-cleanup
git pull origin feature/sprint10-ci-cleanup

# 2. Check status
git status  # Should be clean
```

---

## Test Cases

### TC-001: Build Verification
**Command:** `make build`  
**Expected:** Clean build, binary created at `bin/r8s`  
**Pass Criteria:** No errors, exit code 0  
**Status:** ⬅️ PASS / FAIL

### TC-002: Binary Help
**Command:** `./bin/r8s --help`  
**Expected:** Shows help with all 6 commands listed:
- validate, logs, describe, export, generate, dashboard  
**Pass Criteria:** All commands visible, no errors  
**Status:** ⬅️ PASS / FAIL

### TC-003: Dashboard Command Works
**Command:** `./bin/r8s dashboard` (no bundle - demo mode)  
**Expected:** Dashboard opens with demo items  
**Keyboard Test:** Press `q` to quit  
**Pass Criteria:** Dashboard launches, quits cleanly  
**Status:** ⬅️ PASS / FAIL

### TC-004: Dashboard with Bundle
**Command:** `./bin/r8s dashboard ~/Downloads/r8s-bundle/`  
**Expected:** Dashboard opens, shows bundle path  
**Keyboard Test:** Press `q` to quit  
**Pass Criteria:** Dashboard loads, shows bundle info  
**Status:** ⬅️ PASS / FAIL

### TC-005: Other CLI Commands Work
**Commands:**
```bash
./bin/r8s validate ~/Downloads/r8s-bundle/
./bin/r8s describe ~/Downloads/r8s-bundle/ node-1
./bin/r8s export ~/Downloads/r8s-bundle/ --format json
```
**Expected:** All commands execute without panic  
**Pass Criteria:** No crashes, reasonable output  
**Status:** ⬅️ PASS / FAIL

### TC-006: No Orphaned Imports
**Command:** `grep -r "app\.go\|fetch\.go\|handlers\.go\|logs\.go\|table\.go" internal/ cmd/`  
**Expected:** No references to deleted files  
**Pass Criteria:** Empty result or only in comments  
**Status:** ⬅️ PASS / FAIL

### TC-007: Remaining TUI Files Compile
**Command:** `ls internal/tui/*.go`  
**Expected:** These files exist:
- dashboard.go
- attention.go (if present)
- attention_ai.go (if present)
- styles.go
- helpers.go  
**Pass Criteria:** No missing critical files  
**Status:** ⬅️ PASS / FAIL

---

## Quick Regression Tests

### RT-001: Version Command
**Command:** `./bin/r8s version`  
**Expected:** Shows version, commit, date  
**Status:** ⬅️ PASS / FAIL

### RT-002: Completion Command
**Command:** `./bin/r8s completion bash`  
**Expected:** Outputs bash completion script  
**Status:** ⬅️ PASS / FAIL

---

## Summary

| Test | Status | Notes |
|------|--------|-------|
| TC-001 | ⬅️ | |
| TC-002 | ⬅️ | |
| TC-003 | ⬅️ | |
| TC-004 | ⬅️ | |
| TC-005 | ⬅️ | |
| TC-006 | ⬅️ | |
| TC-007 | ⬅️ | |
| RT-001 | ⬅️ | |
| RT-002 | ⬅️ | |

**Overall:** ⬅️ READY FOR PHASE 2 / NEEDS FIX

---

## Next Steps

- **All PASS:** Say "GO Phase 2" — Continue TUI deletion
- **Any FAIL:** Report which test failed, I'll fix
- **Partial PASS:** Decide if blockers are critical

---

*Run these tests, report results, then we proceed.*
