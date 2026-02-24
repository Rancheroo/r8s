# Sprint 9 Final: Standardization Testing — All Commands

**Scope:** CLI standardization verification across all commands  
**Branch:** `feature/sprint9-cli-polish`  
**Commit:** `303da2b`  
**Tester:** _____________  
**Date:** _____________  

---

## Quick Test Script

Save this as `test-standards.sh`:

```bash
#!/bin/bash
set -e

echo "=== Sprint 9 Standards Testing ==="
echo ""

BUNDLE="./test-bundle"
R8S="./bin/r8s"

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

pass_count=0
fail_count=0

run_test() {
    name="$1"
    cmd="$2"
    expected_exit="$3"
    
    echo -n "Testing: $name ... "
    if eval "$cmd" > /dev/null 2>&1; then
        actual_exit=0
    else
        actual_exit=$?
    fi
    
    if [ "$actual_exit" -eq "$expected_exit" ]; then
        echo -e "${GREEN}PASS${NC} (exit $actual_exit)"
        ((pass_count++))
    else
        echo -e "${RED}FAIL${NC} (expected $expected_exit, got $actual_exit)"
        ((fail_count++))
    fi
}

echo "--- Exit Code Standards ---"

# validate: exit 0 = success, 1 = incomplete, 2 = error
run_test "validate valid bundle" "$R8S validate $BUNDLE" 0
run_test "validate nonexistent" "$R8S validate /nonexistent/" 2

# export: exit 0 = success, 1 = issues, 2 = error
run_test "export valid bundle" "$R8S export $BUNDLE" 0
run_test "export nonexistent" "$R8S export /nonexistent/" 2

# describe: exit 0 = success, 1 = not found, 2 = error
run_test "describe existing resource" "$R8S describe $BUNDLE node-1" 0
run_test "describe nonexistent resource" "$R8S describe $BUNDLE nonexistent-xyz" 1
run_test "describe nonexistent bundle" "$R8S describe /nonexistent/ test" 2

# logs: exit 0 = success, 1 = no logs, 2 = error
run_test "logs valid bundle" "$R8S logs $BUNDLE" 0
run_test "logs nonexistent bundle" "$R8S logs /nonexistent/" 2

echo ""
echo "--- Format Standards ---"

# JSON format should work across all commands
run_test "validate --format=json" "$R8S validate $BUNDLE --format=json | jq ." 0
run_test "describe --format=json" "$R8S describe $BUNDLE node-1 --format=json | jq ." 0
run_test "export --format=json" "$R8S export $BUNDLE --format=json | jq ." 0

# YAML format
run_test "export --format=yaml" "$R8S export $BUNDLE --format=yaml" 0
run_test "describe --format=yaml" "$R8S describe $BUNDLE node-1 --format=yaml" 0

echo ""
echo "--- Help Standards ---"

# All commands should have help
run_test "validate --help" "$R8S validate --help | grep -q 'EXIT CODES'" 0
run_test "describe --help" "$R8S describe --help | grep -q 'EXAMPLES'" 0
run_test "export --help" "$R8S export --help | grep -q 'EXIT CODES'" 0
run_test "logs --help" "$R8S logs --help | grep -q 'EXAMPLES'" 0
run_test "completion --help" "$R8S completion --help" 0
run_test "generate prompt --help" "$R8S generate prompt --help" 0

echo ""
echo "--- Integration Tests ---"

# Pipeline test
run_test "export | jq pipeline" "$R8S export $BUNDLE | jq -e '.meta.r8s_version'" 0
run_test "export validation" "$R8S export $BUNDLE | jq -e '.summary.is_valid'" 0

echo ""
echo "=== Results ==="
echo "Passed: $pass_count"
echo "Failed: $fail_count"

if [ $fail_count -eq 0 ]; then
    echo -e "${GREEN}ALL TESTS PASSED${NC}"
    exit 0
else
    echo -e "${RED}SOME TESTS FAILED${NC}"
    exit 1
fi
```

Run: `chmod +x test-standards.sh && ./test-standards.sh`

---

## Manual Verification

### Standard 1: Exit Codes

| Scenario | Expected | Test Command | Result |
|----------|----------|--------------|--------|
| Success | 0 | `r8s validate ./bundle/; echo $?` | ⬅️ |
| Issues Found | 1 | `r8s describe ./bundle/ nonexistent; echo $?` | ⬅️ |
| Error | 2 | `r8s validate /nonexistent/; echo $?` | ⬅️ |

### Standard 2: Format Flag

| Command | JSON | YAML | Human | Result |
|---------|------|------|-------|--------|
| `validate --format` | ✅ | ⚪️ | ✅ | ⬅️ |
| `describe --format` | ✅ | ✅ | ✅ | ⬅️ |
| `export --format` | ✅ | ✅ | ⚪️ | ⬅️ |

(⚪️ = not applicable or not implemented)

### Standard 3: Help Structure

Check each command has:
- [ ] Short description
- [ ] Long description  
- [ ] EXAMPLES section
- [ ] EXIT CODES section

| Command | Short | Long | EXAMPLES | EXIT CODES | Result |
|---------|-------|------|----------|------------|--------|
| validate | ⬅️ | ⬅️ | ⬅️ | ⬅️ | |
| logs | ⬅️ | ⬅️ | ⬅️ | ⬅️ | |
| describe | ⬅️ | ⬅️ | ⬅️ | ⬅️ | |
| export | ⬅️ | ⬅️ | ⬅️ | ⬅️ | |
| completion | ⬅️ | ⬅️ | ⬅️ | ⚪️ | |
| generate prompt | ⬅️ | ⬅️ | ⬅️ | ⬅️ | |

(⚪️ = completion has usage instead of EXIT CODES, which is OK)

### Standard 4: JSON Structure

| Field | export | describe (json) | validate (json) | Result |
|-------|--------|-----------------|-----------------|--------|
| Structured output | ⬅️ | ⬅️ | ⬅️ | |
| Parseable by jq | ⬅️ | ⬅️ | ⬅️ | |
| Consistent types | ⬅️ | ⬅️ | ⬅️ | |

---

## Summary

| Category | Passed | Failed | Total |
|----------|--------|--------|-------|
| Exit Codes | ___ | ___ | 10 |
| Format Flags | ___ | ___ | 6 |
| Help Structure | ___ | ___ | 6 |
| JSON Structure | ___ | ___ | 3 |
| Integration | ___ | ___ | 3 |

**Overall:** ___% compliance  

**Status:** ⬅️ READY FOR DAY 6 / NEEDS FIX  

---

## Coverage Notes

Current test coverage by command:

| Command | Unit Tests | Integration | Standards | Overall |
|---------|------------|-------------|-----------|---------|
| validate | ✅ | ✅ | ✅ | High |
| logs | ⚪️ | ✅ | ✅ | Medium |
| describe | ⚪️ | ✅ | ✅ | Medium |
| export | ⚪️ | ✅ | ✅ | Medium |
| completion | N/A | ✅ | ✅ | N/A |
| generate prompt | ⚪️ | ⚪️ | ⚪️ | Low |

**Recommendation:** After Sprint 9, add unit tests for:
- Log file parsing edge cases
- Describe resource finding
- Export report generation
- Pattern matching in generate prompt

---

## Report Back

Reply with:
```
Standards Testing: PASS / NEEDS FIX
Exit Code Compliance: X%
Format Compliance: X%
Help Compliance: X%
Ready for Day 6: YES / NO
Coverage Concerns: None / [list]
```

---

*Template: Sprint 9 Final Standards Testing*  
*Generated: 2026-02-19*
