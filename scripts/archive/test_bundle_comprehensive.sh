#!/bin/bash
# Comprehensive Bundle Testing Script
# Tests all bundle functionality including edge cases

BINARY="./bin/r8s"
BUNDLE_ARCHIVE="example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09.tar.gz"
BUNDLE_DIR="/tmp/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09"
RESULTS="/tmp/bundle_test_results.txt"

echo "=== Comprehensive Bundle Testing ===" > "$RESULTS"
echo "Date: $(date)" >> "$RESULTS"
echo "" >> "$RESULTS"

PASS=0
FAIL=0

test_result() {
    local test_name="$1"
    local result="$2"
    local details="$3"
    
    if [ "$result" = "PASS" ]; then
        echo "✅ PASS: $test_name" | tee -a "$RESULTS"
        ((PASS++))
    else
        echo "❌ FAIL: $test_name" | tee -a "$RESULTS"
        echo "   Details: $details" | tee -a "$RESULTS"
        ((FAIL++))
    fi
    echo "" >> "$RESULTS"
}

# TEST 1: Bundle Import (Archive Mode)
echo "TEST 1: Bundle import with archive file"
OUTPUT=$($BINARY bundle import --path="$BUNDLE_ARCHIVE" --limit=100 --verbose 2>&1)
if echo "$OUTPUT" | grep -q "Bundle Import Successful"; then
    if echo "$OUTPUT" | grep -q "96 CRDs"; then
        test_result "Archive import with kubectl parsing" "PASS" ""
    else
        test_result "Archive import with kubectl parsing" "FAIL" "CRDs not parsed (should be 96)"
    fi
else
    test_result "Archive import" "FAIL" "Bundle import failed"
fi

# TEST 2: Bundle Import (Directory Mode)
echo "TEST 2: Bundle import with directory"
if [ -d "$BUNDLE_DIR" ]; then
    OUTPUT=$($BINARY bundle import --path="$BUNDLE_DIR" --verbose 2>&1)
    if echo "$OUTPUT" | grep -q "📁 Detected extracted bundle directory"; then
        if echo "$OUTPUT" | grep -q "96 CRDs"; then
            test_result "Directory import with kubectl parsing" "PASS" ""
        else
            test_result "Directory import with kubectl parsing" "FAIL" "CRDs not parsed"
        fi
    else
        test_result "Directory import detection" "FAIL" "Not detected as directory"
    fi
else
    test_result "Directory import" "FAIL" "Directory not found: $BUNDLE_DIR"
fi

# TEST 3: TUI Launch (Archive)
echo "TEST 3: TUI launch with archive (headless timeout test)"
timeout 2 $BINARY tui --bundle="$BUNDLE_ARCHIVE" 2>&1 &
PID=$!
sleep 1
if kill -0 $PID 2>/dev/null; then
    kill $PID 2>/dev/null
    wait $PID 2>/dev/null
    test_result "TUI archive launch (no panic)" "PASS" ""
else
    test_result "TUI archive launch" "FAIL" "Process crashed immediately"
fi

# TEST 4: TUI Launch (Directory)
echo "TEST 4: TUI launch with directory (headless timeout test)"
if [ -d "$BUNDLE_DIR" ]; then
    timeout 2 $BINARY tui --bundle="$BUNDLE_DIR" 2>&1 &
    PID=$!
    sleep 1
    if kill -0 $PID 2>/dev/null; then
        kill $PID 2>/dev/null
        wait $PID 2>/dev/null
        test_result "TUI directory launch (no panic)" "PASS" ""
    else
        test_result "TUI directory launch" "FAIL" "Process crashed immediately"
    fi
else
    test_result "TUI directory launch" "FAIL" "Directory not found"
fi

# TEST 5: Error Handling - Invalid Path
echo "TEST 5: Error handling for invalid path"
OUTPUT=$($BINARY bundle import --path="/nonexistent/bundle.tar.gz" --verbose 2>&1)
if echo "$OUTPUT" | grep -q "path not found"; then
    if echo "$OUTPUT" | grep -q "TROUBLESHOOTING"; then
        test_result "Invalid path error handling" "PASS" ""
    else
        test_result "Invalid path error handling" "FAIL" "Missing troubleshooting guide"
    fi
else
    test_result "Invalid path error" "FAIL" "Wrong error message"
fi

# TEST 6: Error Handling - Invalid Directory Structure
echo "TEST 6: Error handling for invalid directory"
mkdir -p /tmp/not-a-bundle
OUTPUT=$($BINARY bundle import --path=/tmp/not-a-bundle/ --verbose 2>&1)
if echo "$OUTPUT" | grep -q "invalid bundle directory\|missing rke2"; then
    test_result "Invalid directory error handling" "PASS" ""
else
    test_result "Invalid directory error" "FAIL" "Should reject non-bundle directory"
fi
rmdir /tmp/not-a-bundle 2>/dev/null

# TEST 7: Size Limit Check (Archive)
echo "TEST 7: Size limit enforcement"
OUTPUT=$($BINARY bundle import --path="$BUNDLE_ARCHIVE" --limit=1 --verbose 2>&1)
if echo "$OUTPUT" | grep -q "exceeds limit"; then
    test_result "Size limit enforcement" "PASS" ""
else
    test_result "Size limit enforcement" "FAIL" "Should reject bundle over limit"
fi

# TEST 8: Resource Counts Verification
echo "TEST 8: Verify all resource types parsed"
OUTPUT=$($BINARY bundle import --path="$BUNDLE_ARCHIVE" --limit=100 2>&1)
ERRORS=""
echo "$OUTPUT" | grep -q "86 pods" || ERRORS="${ERRORS}pods missing; "
echo "$OUTPUT" | grep -q "176 logs" || ERRORS="${ERRORS}logs missing; "
echo "$OUTPUT" | grep -q "29 deployments" || ERRORS="${ERRORS}deployments missing; "
echo "$OUTPUT" | grep -q "37 services" || ERRORS="${ERRORS}services missing; "
echo "$OUTPUT" | grep -q "96 CRDs" || ERRORS="${ERRORS}CRDs missing; "
echo "$OUTPUT" | grep -q "17 namespaces" || ERRORS="${ERRORS}namespaces missing; "

if [ -z "$ERRORS" ]; then
    test_result "All resource types parsed" "PASS" ""
else
    test_result "All resource types parsed" "FAIL" "$ERRORS"
fi

# TEST 9: No Warnings Check
echo "TEST 9: No warning messages during import"
OUTPUT=$($BINARY bundle import --path="$BUNDLE_ARCHIVE" --limit=100 --verbose 2>&1)
if echo "$OUTPUT" | grep -qi "warning.*failed to parse"; then
    test_result "No kubectl parsing warnings" "FAIL" "Warnings present in output"
else
    test_result "No kubectl parsing warnings" "PASS" ""
fi

# TEST 10: Bundle Metadata Parsing
echo "TEST 10: Bundle metadata correctly parsed"
OUTPUT=$($BINARY bundle import --path="$BUNDLE_ARCHIVE" --limit=100 2>&1)
ERRORS=""
echo "$OUTPUT" | grep -q "Node Name:.*w-guard" || ERRORS="${ERRORS}node name; "
echo "$OUTPUT" | grep -q "RKE2 Version:" || ERRORS="${ERRORS}RKE2 version; "
echo "$OUTPUT" | grep -q "Bundle Type:.*rke2" || ERRORS="${ERRORS}bundle type; "

if [ -z "$ERRORS" ]; then
    test_result "Bundle metadata parsing" "PASS" ""
else
    test_result "Bundle metadata parsing" "FAIL" "Missing: $ERRORS"
fi

# Summary
echo "" | tee -a "$RESULTS"
echo "================================" | tee -a "$RESULTS"
echo "Test Results Summary" | tee -a "$RESULTS"
echo "================================" | tee -a "$RESULTS"
echo "PASSED: $PASS" | tee -a "$RESULTS"
echo "FAILED: $FAIL" | tee -a "$RESULTS"
echo "" | tee -a "$RESULTS"

if [ $FAIL -eq 0 ]; then
    echo "✅ ALL TESTS PASSED" | tee -a "$RESULTS"
    exit 0
else
    echo "❌ SOME TESTS FAILED" | tee -a "$RESULTS"
    echo "" | tee -a "$RESULTS"
    echo "Full results saved to: $RESULTS"
    exit 1
fi
