#!/bin/bash
# Diagnostic test for exit code fixes

echo "=== Exit Code Diagnostic ==="
echo ""

R8S="./bin/r8s"

echo "Test 1: Check error message format (should say 'Error:' in stderr)"
$R8S export /nonexistent/ 2>&1 | head -1
echo ""

echo "Test 2: Capture exit code immediately"
$R8S export /nonexistent/ 2>/dev/null
EXIT_CODE=$?
echo "Raw exit code: $EXIT_CODE"
echo ""

echo "Test 3: Compare with validate (should be same)"
$R8S validate /nonexistent/ 2>/dev/null
VALIDATE_EXIT=$?
echo "Validate exit code: $VALIDATE_EXIT"
echo ""

echo "Test 4: Check if describe not found works (should be 1)"
$R8S describe ./test-bundle/ nonexistent-xyz 2>/dev/null
DESCRIBE_EXIT=$?
echo "Describe not found exit code: $DESCRIBE_EXIT"
echo ""

echo "=== Expected Results ==="
echo "Export nonexistent: should be 2, got $EXIT_CODE"
echo "Validate nonexistent: should be 2, got $VALIDATE_EXIT"  
echo "Describe not found: should be 1, got $DESCRIBE_EXIT"
echo ""

if [ $EXIT_CODE -eq 2 ] && [ $VALIDATE_EXIT -eq 2 ] && [ $DESCRIBE_EXIT -eq 1 ]; then
    echo "✅ ALL TESTS PASS"
else
    echo "❌ SOME TESTS FAIL"
    echo ""
    echo "Debugging: Check if binary is rebuilt"
    echo "Run: make clean && make build"
fi
