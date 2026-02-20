# Sprint 9 Day 5: Test Plan — CLI Standardization

**Feature:** Standardized flags, exit codes, and help across all commands  
**Branch:** `feature/sprint9-cli-polish`  
**Focus:** Consistency, polish, integration  
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

# 3. Have test bundle ready
ls test-bundle/
```

---

## Standardization Tests

### ST-001: Format Flag Consistency

Test that `--format` accepts same values across all commands:

| Command | Test | Expected | Status |
|---------|------|----------|--------|
| `validate` | `--format=json` | Works | ⬅️ |
| `validate` | `--format=yaml` | Works (or graceful error) | ⬅️ |
| `logs` | `--format=json` | N/A (logs doesn't have format) | ⬅️ |
| `describe` | `--format=json` | Works | ⬅️ |
| `describe` | `--format=yaml` | Works | ⬅️ |
| `describe` | `--format=wide` | Works | ⬅️ |
| `export` | `--format=json` | Works | ⬅️ |
| `export` | `--format=yaml` | Works | ⬅️ |

**Command:**
```bash
./bin/r8s validate ./bundle/ --format=json | jq .
./bin/r8s describe ./bundle/ node-1 --format=json | jq .
./bin/r8s export ./bundle/ --format=yaml
```

---

### ST-002: Exit Code Standardization

**Standard:**
- `0` = Success (no issues)
- `1` = Issues found but command completed (e.g., incomplete bundle)
- `2` = Error (invalid args, file not found, etc.)

| Scenario | Command | Expected Exit | Status |
|----------|---------|---------------|--------|
| Valid bundle | `validate ./good-bundle/` | 0 | ⬅️ |
| Incomplete bundle | `validate ./partial/` | 1 | ⬅️ |
| Invalid bundle | `validate /nonexistent/` | 2 | ⬅️ |
| Valid export | `export ./bundle/` | 0 | ⬅️ |
| Export error | `export /nonexistent/` | 2 | ⬅️ |
| Valid describe | `describe ./bundle/ node-1` | 0 | ⬅️ |
| Not found | `describe ./bundle/ nonexistent` | 1 | ⬅️ |

**Test Script:**
```bash
#!/bin/bash
test_exit() {
    cmd="$1"
    expected="$2"
    desc="$3"
    
    eval "$cmd" > /dev/null 2>&1
    actual=$?
    
    if [ "$actual" -eq "$expected" ]; then
        echo "✅ PASS: $desc (exit $actual)"
    else
        echo "❌ FAIL: $desc (expected $expected, got $actual)"
    fi
}

test_exit "./bin/r8s validate ./bundle/" 0 "Valid bundle"
test_exit "./bin/r8s validate /nonexistent/" 2 "Invalid bundle path"
# etc.
```

---

### ST-003: Help Text Consistency

Check that all commands follow same pattern:

| Element | validate | describe | export | logs | Status |
|---------|----------|----------|--------|------|--------|
| Short description | ⬅️ | ⬅️ | ⬅️ | ⬅️ | |
| Long description | ⬅️ | ⬅️ | ⬅️ | ⬅️ | |
| EXAMPLES section | ⬅️ | ⬅️ | ⬅️ | ⬅️ | |
| EXIT CODES section | ⬅️ | ⬅️ | ⬅️ | ⬅️ | |
| Flags documented | ⬅️ | ⬅️ | ⬅️ | ⬅️ | |

**Command:** `./bin/r8s [command] --help`

---

### ST-004: Flag Naming Consistency

| Flag | Used In | Expected | Status |
|------|---------|----------|--------|
| `--format` | validate, describe, export | Same behavior | ⬅️ |
| `--output` / `-o` | export | File output | ⬅️ |
| `--namespace` / `-n` | logs, describe | Namespace filter | ⬅️ |
| `-f` | logs | Follow mode | ⬅️ (don't confuse with format) |

**Note:** `logs` uses `-f` for follow, NOT format. This is intentional (matches `kubectl`).

---

## Integration Tests

### IT-001: Pipeline Integration

```bash
#!/bin/bash
# Full pipeline test

# Test 1: Validate, export, check
./bin/r8s validate ./bundle/ || exit 2
./bin/r8s export ./bundle/ --format=json | jq -e '.health.is_valid' || exit 2

# Test 2: Export critical, describe, logs
CRITICAL=$(./bin/r8s export ./bundle/ --severity=critical | jq '.summary.critical_count')
if [ "$CRITICAL" -gt 0 ]; then
    ./bin/r8s describe ./bundle/ node-1
fi

echo "✅ Pipeline test passed"
```

**Status:** ⬅️ PASS / FAIL

### IT-002: JSON Consistency

All JSON outputs should have consistent structure:

| Command | Has `meta`? | Has timestamps? | Consistent fields? | Status |
|---------|-------------|-----------------|-------------------|--------|
| `export` | ⬅️ | ⬅️ | ⬅️ | |
| `describe -o json` | ⬅️ N/A | ⬅️ N/A | ⬅️ | |
| `validate -o json` | ⬅️ N/A | ⬅️ N/A | ⬅️ | |

---

## Regression Tests

### RT-001: All Commands Still Work

```bash
./bin/r8s validate ./bundle/
./bin/r8s logs ./bundle/
./bin/r8s describe ./bundle/ node-1
./bin/r8s export ./bundle/
./bin/r8s generate prompt ./bundle/
./bin/r8s completion bash
./bin/r8s --help
```

**Status:** ⬅️ PASS / FAIL

### RT-002: Backward Compatibility

Old command patterns should still work:

```bash
./bin/r8s validate ./bundle/ --format=table
./bin/r8s describe ./bundle/ --format=yaml
./bin/r8s export ./bundle/ --format=json
```

**Status:** ⬅️ PASS / FAIL

---

## Performance Tests

### PT-001: Command Startup Time

Each command should return help in <100ms:

```bash
time ./bin/r8s validate --help
time ./bin/r8s logs --help
time ./bin/r8s describe --help
time ./bin/r8s export --help
```

**Expected:** All <100ms
**Status:** ⬅️ PASS / FAIL

### PT-002: Large Bundle Export

```bash
time ./bin/r8s export ./large-bundle/ --format=json > /dev/null
```

**Expected:** <2 seconds for 500MB bundle
**Status:** ⬅️ PASS / FAIL / N/A

---

## Standards Checklist

| Standard | Status | Notes |
|----------|--------|-------|
| Exit code 0 = success | ⬅️ | |
| Exit code 1 = issues found | ⬅️ | |
| Exit code 2 = error | ⬅️ | |
| `--format` accepts same values | ⬅️ | |
| Help text has EXAMPLES | ⬅️ | |
| Help text has EXIT CODES | ⬅️ | |
| JSON structure consistent | ⬅️ | |
| Flags use consistent short forms | ⬅️ | |
| All commands have --help | ⬅️ | |

---

## Summary

| Category | Passed | Failed | Total |
|----------|--------|--------|-------|
| Standardization | ___ | ___ | 4 |
| Integration | ___ | ___ | 2 |
| Regression | ___ | ___ | 2 |
| Performance | ___ | ___ | 2 |
| Standards | ___ | ___ | 10 |

**Overall Status:** ⬅️ READY FOR DAY 6 / NEEDS WORK  

**Standards Compliance:** ___%  

**Blockers:**
_________________________________

---

## How to Report Back

Reply with:
```
Status: PASS / NEEDS WORK
Standards Compliance: X%
Critical Issues: None / [list]
Ready for Day 6: YES / NO
```

---

*Template version: Sprint 9 Day 5*  
*Focus: CLI consistency and polish*
