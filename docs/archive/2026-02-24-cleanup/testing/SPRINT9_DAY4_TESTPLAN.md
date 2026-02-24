# Sprint 9 Day 4: Test Plan — r8s export

**Feature:** JSON/YAML findings export for CI/CD integration  
**Branch:** `feature/sprint9-cli-polish`  
**Commit:** `05fb797`  
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

# 3. Have jq installed (for JSON testing)
which jq

# 4. Have a test bundle
ls test-bundle/rke2/
```

---

## Test Cases

### TC-001: Export Command Exists
**Command:** `./bin/r8s export --help`  
**Expected:** Shows help with all flags listed  
**Verify Flags:** `--format`, `--severity`, `--pattern`, `--output`  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### TC-002: JSON Export (Default)
**Command:** `./bin/r8s export ./test-bundle/`  
**Expected:** Valid JSON output  
**Verify:** `./bin/r8s export ./test-bundle/ | jq .` succeeds  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### TC-003: YAML Export
**Command:** `./bin/r8s export ./test-bundle/ --format=yaml`  
**Expected:** Valid YAML output  
**Verify:** Contains `health:`, `findings:`, `summary:` sections  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### TC-004: Export to File
**Command:** `./bin/r8s export ./test-bundle/ --output=report.json`  
**Expected:** File created, success message displayed  
**Verify:** `cat report.json | jq .meta.r8s_version` shows v0.8.0-alpha  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### TC-005: Severity Filter (Critical)
**Command:** `./bin/r8s export ./test-bundle/ --severity=critical`  
**Expected:** Only critical findings exported  
**Verify:** `jq '.findings[] | select(.severity == "critical")' | wc -l` > 0  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### TC-006: Severity Filter (Warning)
**Command:** `./bin/r8s export ./test-bundle/ --severity=warning`  
**Expected:** Only warning and critical findings (no info)  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### TC-007: CI/CD Pipeline Integration
**Command:**
```bash
./bin/r8s export ./test-bundle/ | \
  jq -e '.summary.is_valid' || echo "FAIL"
```
**Expected:** Returns exit code 0 if valid, 1 if not  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### TC-008: Health Percentage Query
**Command:**
```bash
./bin/r8s export ./test-bundle/ | \
  jq '.summary.health_percentage'
```
**Expected:** Returns numeric value (0-100)  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### TC-009: Bundle Not Found
**Command:** `./bin/r8s export /nonexistent/`  
**Expected:** Error: "bundle path not found"  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### TC-010: Report Structure
**Command:** `./bin/r8s export ./test-bundle/ --format=json | jq keys`  
**Expected:** Shows `meta`, `health`, `findings`, `summary`  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

---

## CI/CD Integration Tests

### CT-001: CI Exit Code Check
**Script:**
```bash
#!/bin/bash
r8s export ./bundle/ | jq -e '.health.is_valid' || exit 1
echo "Bundle valid"
```
**Expected:** Exits with code 1 if bundle invalid  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### CT-002: Monitoring Integration
**Script:**
```bash
CRITICAL=$(r8s export ./bundle/ --severity=critical | \
  jq '.summary.critical_count')
echo "Critical issues: $CRITICAL"
```
**Expected:** Echoes number of critical issues  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

---

## Regression Tests

### RT-001: Other Commands Still Work
**Commands:**
```bash
./bin/r8s validate ./test-bundle/
./bin/r8s logs ./test-bundle/
./bin/r8s describe ./test-bundle/ node-1
./bin/r8s generate prompt ./test-bundle/
./bin/r8s completion bash
```
**Expected:** All work without errors  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### RT-002: Help Shows Export Command
**Command:** `./bin/r8s --help`  
**Expected:** "export" appears in Available Commands  
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

---

## Report Structure Validation

### RS-001: Meta Section
**Command:** `./bin/r8s export ./test-bundle/ | jq .meta`  
**Expected Fields:**
- generated_at
- bundle_path
- bundle_type
- r8s_version
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### RS-002: Health Section
**Command:** `./bin/r8s export ./test-bundle/ | jq .health`  
**Expected Fields:**
- completeness
- is_valid
- found_files
- total_files
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### RS-003: Findings Section
**Command:** `./bin/r8s export ./test-bundle/ | jq '.findings[0]'`  
**Expected Fields:**
- id
- severity
- category
- title
- description
- suggestion
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

### RS-004: Summary Section
**Command:** `./bin/r8s export ./test-bundle/ | jq .summary`  
**Expected Fields:**
- total_findings
- critical_count
- warning_count
- health_percentage
- is_valid
**Actual:** _________________________________  
**Status:** ⬅️ PASS / FAIL  

---

## Environment Info

```bash
OS: _____________
Bundle completeness: _____________
jq version: _____________
Go version: _____________
```

---

## Summary

| Category | Passed | Failed | Total |
|----------|--------|--------|-------|
| Test Cases | ___ | ___ | 10 |
| CI/CD Tests | ___ | ___ | 2 |
| Regression | ___ | ___ | 2 |
| Report Structure | ___ | ___ | 4 |

**Overall Status:** ⬅️ READY FOR DAY 5 / BLOCKED  

**CI/CD Ready:** YES / NO  

**Blockers:**
_________________________________

**Integration Notes:**
_________________________________

---

## How to Report Back

Reply with:
```
Status: PASS / FAIL (X tests)
CI/CD Ready: YES / NO
Blockers: None / [describe]
Ready for Day 5: YES / NO
```

---

*Template version: Sprint 9 Day 4*  
*Generated: 2026-02-19*
