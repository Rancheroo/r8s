# Sprint 9 Day 2: Test Plan — r8s logs

**Feature:** kubectl-style log streaming from bundles  
**Branch:** `feature/sprint9-cli-polish`  
**Commit:** `2601537`  
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

# 3. Have a test bundle with podlogs
test-bundle/
└── rke2/
    └── podlogs/
        └── cattle-system_rancher-7c4c7b8f5-x2v9p_rancher.log
```

---

## Test Cases

### TC-001: Logs Command Exists
**Command:** `./bin/r8s logs --help`  
**Expected:** Shows help with all flags listed  
**Verify Flags:** `--namespace`, `--follow`, `--timestamps`, `--tail`, `--container`  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### TC-002: Basic Log Output
**Command:** `./bin/r8s logs ./test-bundle/`  
**Expected:** Shows logs from all pods in bundle  
**Verify:** Headers show namespace/pod/container  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### TC-003: Filter by Namespace
**Command:** `./bin/r8s logs ./test-bundle/ -n cattle-system`  
**Expected:** Only shows logs from cattle-system namespace  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### TC-004: Filter by Pod Name
**Command:** `./bin/r8s logs ./test-bundle/ rancher-7c4c7b8f5`  
**Expected:** Only shows logs matching pod name (partial match OK)  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### TC-005: Tail Last N Lines
**Command:** `./bin/r8s logs ./test-bundle/ --tail=10`  
**Expected:** Shows only last 10 lines per log file  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### TC-006: Color Coding
**Command:** `./bin/r8s logs ./test-bundle/`  
**Expected:** 
- ERROR/FATAL lines in RED
- WARN lines in YELLOW  
- INFO lines in CYAN
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### TC-007: No Logs Found
**Command:** `./bin/r8s logs ./test-bundle/ -n nonexistent`  
**Expected:** Graceful message "No logs found in namespace 'nonexistent'"  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### TC-008: Bundle Not Found
**Command:** `./bin/r8s logs /nonexistent/`  
**Expected:** Error: "bundle path not found"  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### TC-009: Follow Mode (Basic)
**Command:** `./bin/r8s logs ./test-bundle/ -f`  
**Setup:** Run in terminal, wait 2 seconds, Ctrl+C  
**Expected:** Shows "Following logs (Ctrl+C to stop)..." then streams  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL / N/A  

---

## Regression Tests

### RT-001: Other Commands Still Work
**Commands:**
```bash
./bin/r8s validate ./test-bundle/
./bin/r8s generate prompt ./test-bundle/
./bin/r8s completion bash
```
**Expected:** All work without errors  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### RT-002: Help Shows Logs Command
**Command:** `./bin/r8s --help`  
**Expected:** "logs" appears in Available Commands  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

---

## Comparison to kubectl

| Feature | kubectl logs | r8s logs | Status |
|---------|--------------|----------|--------|
| Stream logs | ✅ | ✅ | ⬅️ TEST |
| Filter by namespace | `-n` | `-n` | ⬅️ TEST |
| Follow mode | `-f` | `-f` | ⬅️ TEST |
| Tail last N | `--tail` | `--tail` | ⬅️ TEST |
| Timestamps | `--timestamps` | `--timestamps` | ⬅️ TEST |
| Specific pod | `pod-name` | `pod-name` | ⬅️ TEST |

---

## Performance Tests

### PT-001: Large Log File
**Setup:** Bundle with 100MB+ log file  
**Command:** `./bin/r8s logs ./large-bundle/ --tail=100`  
**Expected:** Returns in <2 seconds  
**Actual Time:** _____________  
**Status:** ⬅️ PASS / FAIL / N/A  

### PT-002: Many Log Files
**Setup:** Bundle with 100+ pod log files  
**Command:** `./bin/r8s logs ./many-logs-bundle/`  
**Expected:** Returns in <5 seconds  
**Actual Time:** _____________  
**Status:** ⬅️ PASS / FAIL / N/A  

---

## Environment Info

```bash
# Fill this in
OS: _____________
Shell: _____________
Bundle Size: _____________
Number of Log Files: _____________
```

---

## Summary

| Category | Passed | Failed | Total |
|----------|--------|--------|-------|
| Test Cases | ___ | ___ | 9 |
| Regression | ___ | ___ | 2 |
| Performance | ___ | ___ | 2 |

**Overall Status:** ⬅️ READY FOR DAY 3 / BLOCKED  

**Blockers (if any):**
_________________________________
_________________________________

**Comparison to kubectl:**
- Missing features: _________________________________
- Extra features: _________________________________ (color coding)

**Additional Notes:**
_________________________________
_________________________________

---

## How to Report Back

**Option 1:** Fill out this form  
**Option 2:** Reply with:
```
Status: PASS / FAIL (X tests)
Blockers: None / [describe]
Ready for Day 3: YES / NO
kubectl parity: Y% (X/Y features)
```

**Send to:** SuseBot / RancherSRE  
**Channel:** Code Sprint  

---

*Template version: Sprint 9 Day 2*  
*Generated: 2026-02-19*
