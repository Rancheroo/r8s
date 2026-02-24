# Sprint 9 Week 2 Day 8: Dashboard Test Plan

**Feature:** `r8s dashboard` - Minimal TUI  
**Branch:** `feature/sprint9-cli-polish`  
**Commit:** `72f025f`  
**Tester:** _____________  
**Date:** _____________  

---

## Quick Test

```bash
# Pull and build
git pull origin feature/sprint9-cli-polish
make build

# Test 1: Launch with bundle
./bin/r8s dashboard ~/Downloads/r8s-bundle/

# Test 2: Launch demo mode (no bundle)
./bin/r8s dashboard

# Test 3: Invalid bundle
./bin/r8s dashboard /nonexistent/
```

---

## Test Cases

### TC-001: Dashboard Command Exists
**Command:** `./bin/r8s dashboard --help`  
**Expected:** Shows help with keyboard shortcuts  
**Status:** ⬅️ PASS / FAIL  

### TC-002: Launch with Valid Bundle
**Command:** `./bin/r8s dashboard ~/Downloads/r8s-bundle/`  
**Expected:** Dashboard opens, shows bundle info and items  
**Keyboard Test:** Press `q` to quit  
**Status:** ⬅️ PASS / FAIL  

### TC-003: Launch Demo Mode
**Command:** `./bin/r8s dashboard`  
**Expected:** Dashboard opens in demo mode with sample items  
**Status:** ⬅️ PASS / FAIL  

### TC-004: Invalid Bundle Path
**Command:** `./bin/r8s dashboard /nonexistent/`  
**Expected:** Error message, exit code 2  
**Status:** ⬅️ PASS / FAIL  

### TC-005: Navigation
**In dashboard:**
- Press `↓` or `j` - Should move cursor down
- Press `↑` or `k` - Should move cursor up  
- Press `Enter` - Should show details
- Press `?` - Should show help
- Press `q` - Should quit

**Status:** ⬅️ PASS / FAIL  

---

## Regression Tests

### RT-001: Other Commands Still Work
```bash
./bin/r8s validate ./bundle/
./bin/r8s logs ./bundle/
./bin/r8s describe ./bundle/ node-1
./bin/r8s export ./bundle/
```

**Status:** ⬅️ PASS / FAIL  

### RT-002: Help Shows Dashboard
**Command:** `./bin/r8s --help | grep dashboard`  
**Expected:** "dashboard" appears in commands  
**Status:** ⬅️ PASS / FAIL  

---

## Summary

| Test | Status |
|------|--------|
| TC-001 | ⬅️ |
| TC-002 | ⬅️ |
| TC-003 | ⬅️ |
| TC-004 | ⬅️ |
| TC-005 | ⬅️ |
| RT-001 | ⬅️ |
| RT-002 | ⬅️ |

**Overall:** ⬅️ READY FOR DAY 9 / NEEDS FIX  

---

## Note for Day 9

After dashboard is verified, proceed with mass TUI deletion per:
`docs/development/TUI_SUNSET_PLAN.md`

**Target:** Delete ~8,000 lines of TUI code, keep only dashboard

---

*Day 8 complete when TC-002 and TC-003 pass*
