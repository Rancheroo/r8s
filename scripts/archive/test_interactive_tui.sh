#!/bin/bash
# Interactive TUI Testing Script for r8s
# Tests critical functionality systematically

set -e

BINARY="./bin/r8s"
TEST_RESULTS="/tmp/r8s_test_results.txt"

echo "=== r8s TUI Critical Bug Testing ===" > "$TEST_RESULTS"
echo "Date: $(date)" >> "$TEST_RESULTS"
echo "" >> "$TEST_RESULTS"

# Test 1: Startup with no args
echo "TEST 1: Startup with no args shows help"
if $BINARY 2>&1 | grep -q "r8s (Rancheroos)"; then
    echo "✅ PASS: Help displayed" | tee -a "$TEST_RESULTS"
else
    echo "❌ FAIL: Help not displayed" | tee -a "$TEST_RESULTS"
fi

# Test 2: Invalid flag handling
echo ""
echo "TEST 2: Invalid flag shows error"
if $BINARY tui --invalid-flag 2>&1 | grep -q "unknown flag"; then
    echo "✅ PASS: Error message for invalid flag" | tee -a "$TEST_RESULTS"
else
    echo "❌ FAIL: No error for invalid flag" | tee -a "$TEST_RESULTS"
fi

# Test 3: Version command
echo ""
echo "TEST 3: Version command works"
if $BINARY version 2>&1 | grep -q "version"; then
    echo "✅ PASS: Version displayed" | tee -a "$TEST_RESULTS"
else
    echo "❌ FAIL: Version not displayed" | tee -a "$TEST_RESULTS"
fi

# Test 4: Help command
echo ""
echo "TEST 4: Help command works"
if $BINARY help 2>&1 | grep -q "Available Commands"; then
    echo "✅ PASS: Help command works" | tee -a "$TEST_RESULTS"
else
    echo "❌ FAIL: Help command failed" | tee -a "$TEST_RESULTS"
fi

# Test 5: Config init (non-destructive check)
echo ""
echo "TEST 5: Config command exists"
if $BINARY config --help 2>&1 | grep -q "Manage r8s configuration"; then
    echo "✅ PASS: Config command available" | tee -a "$TEST_RESULTS"
else
    echo "❌ FAIL: Config command missing" | tee -a "$TEST_RESULTS"
fi

# Test 6: Bundle command exists
echo ""
echo "TEST 6: Bundle command exists"
if $BINARY bundle --help 2>&1 | grep -q "Work with support bundles"; then
    echo "✅ PASS: Bundle command available" | tee -a "$TEST_RESULTS"
else
    echo "❌ FAIL: Bundle command missing" | tee -a "$TEST_RESULTS"
fi

# Test 7: TUI mockdata flag is recognized
echo ""
echo "TEST 7: TUI accepts --mockdata flag"
if timeout 1 $BINARY tui --mockdata 2>&1 || [ $? -eq 124 ]; then
    echo "✅ PASS: --mockdata flag accepted (TUI launched)" | tee -a "$TEST_RESULTS"
else
    echo "❌ FAIL: --mockdata flag rejected" | tee -a "$TEST_RESULTS"
fi

# Test 8: Verbose flag works
echo ""
echo "TEST 8: Verbose flag accepted"
if $BINARY -v version 2>&1 | grep -q "version"; then
    echo "✅ PASS: Verbose flag accepted" | tee -a "$TEST_RESULTS"
else
    echo "❌ FAIL: Verbose flag failed" | tee -a "$TEST_RESULTS"
fi

echo ""
echo "=== Basic CLI Tests Complete ===" | tee -a "$TEST_RESULTS"
echo ""
cat "$TEST_RESULTS"

echo ""
echo "For interactive TUI testing, run:"
echo "  $BINARY tui --mockdata"
echo ""
echo "Test these keyboard shortcuts manually:"
echo "  - Arrow keys for navigation"
echo "  - Enter to drill down"
echo "  - Esc to go back"
echo "  - 'C' for CRDs from Cluster view"
echo "  - 'd' to describe resources"
echo "  - 'l' to view logs"
echo "  - '?' for help"
echo "  - 'q' to quit"
