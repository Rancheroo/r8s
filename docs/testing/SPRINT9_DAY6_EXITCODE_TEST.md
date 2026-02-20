# Sprint 9 Day 6: Exit Code Fixes - Test Plan

**Changes:** export, describe, logs now use standard exit codes  
**Branch:** `feature/sprint9-cli-polish`  
**Commit:** `09ecd2f`  
**Goal:** 90%+ exit code compliance  

---

## Exit Code Standards (Retest)

| Code | Meaning | When Used |
|------|---------|-----------|
| 0 | Success | No issues, command completed |
| 1 | Issues Found | Incomplete bundle, resource not found, warnings |
| 2 | Error | Invalid args, file not found, system error |

---

## Test Cases

### TC-001: Export Bundle Not Found
**Command:** `./bin/r8s export /nonexistent/`  
**Expected Exit:** 2  
**Previous:** 1 ❌  
**Status:** ⬅️ TEST  

### TC-002: Describe Bundle Not Found  
**Command:** `./bin/r8s describe /nonexistent/ test`  
**Expected Exit:** 2  
**Previous:** 1 ❌  
**Status:** ⬅️ TEST  

### TC-003: Describe Resource Not Found
**Command:** `./bin/r8s describe ./bundle/ nonexistent-pod`  
**Expected Exit:** 1  
**Previous:** 0 ❌  
**Status:** ⬅️ TEST  

### TC-004: Logs Bundle Not Found
**Command:** `./bin/r8s logs /nonexistent/`  
**Expected Exit:** 2  
**Previous:** 1 ❌  
**Status:** ⬅️ TEST  

### TC-005: Export Valid Bundle
**Command:** `./bin/r8s export ./bundle/`  
**Expected Exit:** 0 or 1 (depending on bundle health)  
**Status:** ⬅️ TEST  

### TC-006: Describe Existing Resource
**Command:** `./bin/r8s describe ./bundle/ node-1`  
**Expected Exit:** 0  
**Status:** ⬅️ TEST  

---

## Quick Test Script

```bash
#!/bin/bash
R8S="./bin/r8s"
BUNDLE="./test-bundle"

echo "=== Exit Code Fix Verification ==="

# Test 1: export bundle not found → 2
$R8S export /nonexistent/ 2>/dev/null
if [ $? -eq 2 ]; then echo "✅ PASS: export /nonexistent/ = 2"; else echo "❌ FAIL"; fi

# Test 2: describe bundle not found → 2
$R8S describe /nonexistent/ test 2>/dev/null
if [ $? -eq 2 ]; then echo "✅ PASS: describe /nonexistent/ = 2"; else echo "❌ FAIL"; fi

# Test 3: describe not found → 1
$R8S describe $BUNDLE nonexistent-xyz 2>/dev/null
if [ $? -eq 1 ]; then echo "✅ PASS: describe not found = 1"; else echo "❌ FAIL"; fi

# Test 4: logs bundle not found → 2
$R8S logs /nonexistent/ 2>/dev/null
if [ $? -eq 2 ]; then echo "✅ PASS: logs /nonexistent/ = 2"; else echo "❌ FAIL"; fi

# Test 5: export valid → 0 or 1
$R8S export $BUNDLE >/dev/null 2>&1
EXIT=$?
if [ $EXIT -eq 0 ] || [ $EXIT -eq 1 ]; then 
    echo "✅ PASS: export valid = $EXIT (0 or 1 OK)"
else 
    echo "❌ FAIL: export valid = $EXIT (expected 0 or 1)"
fi

echo ""
echo "=== Done ==="
```

---

## Standards Compliance Calculation

| Category | Before | After (Expected) |
|----------|--------|------------------|
| Export errors | 1 instead of 2 | 2 ✅ |
| Describe errors | 1 instead of 2 | 2 ✅ |
| Describe not found | 0 instead of 1 | 1 ✅ |
| Logs errors | 1 instead of 2 | 2 ✅ |
| **Overall** | **80%** | **~95%** |

---

## How to Report

```bash
# Pull fixes
git pull origin feature/sprint9-cli-polish
make build

# Run quick test
./test-exit-codes.sh

# Or manual
./bin/r8s export /nonexistent/; echo "Exit: $?"
./bin/r8s describe /nonexistent/ test; echo "Exit: $?"
./bin/r8s describe ./bundle/ nonexistent; echo "Exit: $?"
./bin/r8s logs /nonexistent/; echo "Exit: $?"
```

Reply: `PASS (X%)` or `FAIL (details)`

---

*Fixes in commit: 09ecd2f*  
*Target: 90%+ compliance*
