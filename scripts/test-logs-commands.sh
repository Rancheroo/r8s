#!/bin/bash
set -e

BUNDLE="/tmp/r8s-demo-deepdive"
R8S="./bin/r8s"

echo "Testing r8s logs commands..."

# 1. Namespace Filter (cattle-system)
echo "1. Testing -n cattle-system..."
if $R8S logs $BUNDLE -n cattle-system 2>&1 | grep -q "No logs found"; then
    echo "❌ Failed: Namespace filter returned no logs"
    FAIL=1
else
    echo "✅ Success"
fi

# 2. Pod Name Filter (rancher-webhook)
echo "2. Testing pod filter 'rancher-webhook'..."
if $R8S logs $BUNDLE "rancher-webhook" 2>&1 | grep -q "No logs found"; then
    echo "❌ Failed: Pod filter 'rancher-webhook' returned no logs"
    FAIL=1
else
    echo "✅ Success"
fi

# 3. Fuzzy Filter (webhook)
echo "3. Testing fuzzy filter 'webhook'..."
if $R8S logs $BUNDLE "webhook" 2>&1 | grep -q "No logs found"; then
    echo "❌ Failed: Fuzzy filter 'webhook' returned no logs"
    FAIL=1
else
    echo "✅ Success"
fi

if [ "$FAIL" == "1" ]; then
    echo "Tests FAILED"
    exit 1
fi

echo "All tests PASSED"
exit 0
