# Sprint 9 Day 3: Test Plan — r8s describe

**Feature:** kubectl-style resource description from bundles  
**Branch:** `feature/sprint9-cli-polish`  
**Commit:** `7318712`  
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

# 3. Have a test bundle with kubectl output
ls test-bundle/rke2/kubectl/
# Should see: pods, nodes, deployments, services, etc.
```

---

## Test Cases

### TC-001: Describe Command Exists
**Command:** `./bin/r8s describe --help`  
**Expected:** Shows help with all flags listed  
**Verify Flags:** `--namespace`, `--output`, `--selector`  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### TC-002: Describe Pod
**Command:** `./bin/r8s describe pod ./test-bundle/ rancher-xyz`  
**Expected:** Shows pod details including name, namespace, status  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### TC-003: Describe Node
**Command:** `./bin/r8s describe node ./test-bundle/ node-1`  
**Expected:** Shows node details  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### TC-004: JSON Output
**Command:** `./bin/r8s describe pod ./test-bundle/ -o json`  
**Expected:** Valid JSON output, parseable by `jq`  
**Verify:** `./bin/r8s describe ./test-bundle/ -o json | jq .` works  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### TC-005: YAML Output
**Command:** `./bin/r8s describe node ./test-bundle/ -o yaml`  
**Expected:** Valid YAML output  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### TC-006: Wide Format
**Command:** `./bin/r8s describe pod ./test-bundle/ -o wide`  
**Expected:** Table format (KIND, NAME, NAMESPACE, STATUS columns)  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### TC-007: Namespace Filter
**Command:** `./bin/r8s describe pod ./test-bundle/ -n cattle-system`  
**Expected:** Only shows pods from cattle-system namespace  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### TC-008: Auto-Detect Kind
**Command:** `./bin/r8s describe ./test-bundle/ rancher-xyz`  
**Expected:** Auto-detects "pod" kind and shows details  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### TC-009: Resource Not Found
**Command:** `./bin/r8s describe pod ./test-bundle/ nonexistent-pod-12345`  
**Expected:** Graceful message "No resources found of kind 'pod' with name..."  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### TC-010: Bundle Not Found
**Command:** `./bin/r8s describe pod /nonexistent/ test`  
**Expected:** Error: "bundle path not found"  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### TC-011: Describe All (No Name Filter)
**Command:** `./bin/r8s describe pod ./test-bundle/`  
**Expected:** Shows all pods in bundle (no name filter)  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL / N/A  

---

## Regression Tests

### RT-001: Other Commands Still Work
**Commands:**
```bash
./bin/r8s validate ./test-bundle/
./bin/r8s logs ./test-bundle/
./bin/r8s generate prompt ./test-bundle/
./bin/r8s completion bash
```
**Expected:** All work without errors  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### RT-002: Help Shows Describe Command
**Command:** `./bin/r8s --help`  
**Expected:** "describe" appears in Available Commands  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

---

## kubectl Parity Check

| Feature | kubectl describe | r8s describe | Status |
|---------|------------------|--------------|--------|
| Describe specific resource | `kubectl describe pod <name>` | `./bin/r8s describe pod ./bundle/ <name>` | ⬅️ TEST |
| JSON output | `-o json` | `-o json` | ⬅️ TEST |
| YAML output | `-o yaml` | `-o yaml` | ⬅️ TEST |
| Wide format | `-o wide` | `-o wide` | ⬅️ TEST |
| Namespace filter | `-n` | `-n` | ⬅️ TEST |
| Auto-detect kind | N/A | `./bin/r8s describe ./bundle/ name` | ⬅️ EXTRA |

---

## Environment Info

```bash
OS: _____________
Bundle contents: _____________
Go version: _____________
```

---

## Summary

| Category | Passed | Failed | Total |
|----------|--------|--------|-------|
| Test Cases | ___ | ___ | 11 |
| Regression | ___ | ___ | 2 |

**Overall Status:** ⬅️ READY FOR DAY 4 / BLOCKED  

**kubectl Parity:** ___%  

**Blockers:**
_________________________________

**Additional Notes:**
_________________________________

---

## How to Report Back

Reply with:
```
Status: PASS / FAIL (X tests)
Blockers: None / [describe]
Ready for Day 4: YES / NO
kubectl parity: Y%
```

---

*Template version: Sprint 9 Day 3*  
*Generated: 2026-02-19*
