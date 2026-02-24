# Sprint 9 Day 1: Test Plan — r8s completion

**Feature:** Shell completion support  
**Branch:** `feature/sprint9-cli-polish`  
**Commit:** `17e2677`  
**Tester:** _____________  
**Date:** _____________  

---

## Prerequisites

```bash
# 1. Pull latest
git fetch origin
git checkout feature/sprint9-cli-polish

# 2. Build
make build

# 3. Verify binary exists
ls -la bin/r8s
./bin/r8s version
```

---

## Test Cases

### TC-001: Completion Command Exists
**Command:** `./bin/r8s completion --help`  
**Expected:** Shows help text with usage examples  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### TC-002: Bash Completion Generation
**Command:** `./bin/r8s completion bash`  
**Expected:** Outputs bash script (no errors)  
**Verify:** Script contains `_r8s_root_command()` function  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### TC-003: Zsh Completion Generation
**Command:** `./bin/r8s completion zsh`  
**Expected:** Outputs zsh script (no errors)  
**Verify:** Script starts with `#compdef r8s`  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### TC-004: Fish Completion Generation
**Command:** `./bin/r8s completion fish`  
**Expected:** Outputs fish script (no errors)  
**Verify:** Script contains `complete -c r8s`  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### TC-005: PowerShell Completion Generation
**Command:** `./bin/r8s completion powershell`  
**Expected:** Outputs PowerShell script (no errors)  
**Verify:** Script contains `Register-ArgumentCompleter`  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### TC-006: Invalid Shell Error
**Command:** `./bin/r8s completion invalid_shell`  
**Expected:** Shows usage/help (graceful handling)  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### TC-007: Bash Completion Functional (if tested)
**Setup:**
```bash
source <(./bin/r8s completion bash)
```
**Test:** Type `./bin/r8s val` + TAB  
**Expected:** Auto-completes to `validate`  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL / N/A  

### TC-008: Zsh Completion Functional (if tested)
**Setup:**
```zsh
source <(./bin/r8s completion zsh)
```
**Test:** Type `./bin/r8s gen` + TAB  
**Expected:** Auto-completes to `generate`  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL / N/A  

---

## Regression Tests

### RT-001: Other Commands Still Work
**Commands:**
```bash
./bin/r8s validate --help
./bin/r8s generate prompt --help
./bin/r8s test-cluster --help
```
**Expected:** All show help without errors  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### RT-002: Root Command Unchanged
**Command:** `./bin/r8s --help`  
**Expected:** Shows all commands including `completion`  
**Verify:** "completion" appears in Available Commands list  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

---

## Environment Info

```bash
# Fill this in
OS: _____________
Shell: _____________
Go version: _____________
```

---

## Summary

| Category | Passed | Failed | Total |
|----------|--------|--------|-------|
| Test Cases | ___ | ___ | 8 |
| Regression | ___ | ___ | 2 |

**Overall Status:** ⬅️ READY FOR DAY 2 / BLOCKED  

**Blockers (if any):**
_________________________________
_________________________________
_________________________________

**Additional Notes:**
_________________________________
_________________________________

---

## How to Report Back

**Option 1:** Fill out this form and send back  
**Option 2:** Reply with summary:
```
Status: PASS (all tests) / FAIL (X tests)
Blockers: None / [describe]
Ready for Day 2: YES / NO
```

**Send to:** SuseBot / RancherSRE  
**Channel:** Code Sprint  

---

*Template version: Sprint 9 Day 1*  
*Generated: 2026-02-17*
