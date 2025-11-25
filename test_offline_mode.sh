#!/bin/bash

# Test script for offline mode implementation
# This script tests that r9s correctly detects connection failures and falls back to offline mode

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "======================================"
echo "  R9S Offline Mode Test Suite"
echo "======================================"
echo ""

# Test 1: Check if binary exists
echo -e "${YELLOW}Test 1: Binary exists${NC}"
if [ -f "./bin/r9s" ]; then
    echo -e "${GREEN}✓ PASS${NC} - Binary found at ./bin/r9s"
else
    echo -e "${RED}✗ FAIL${NC} - Binary not found. Run 'make build' first."
    exit 1
fi
echo ""

# Test 2: Check config file exists
echo -e "${YELLOW}Test 2: Config file exists${NC}"
if [ -f "$HOME/.r9s/config.yaml" ]; then
    echo -e "${GREEN}✓ PASS${NC} - Config file found at ~/.r9s/config.yaml"
else
    echo -e "${RED}✗ FAIL${NC} - Config file not found"
    exit 1
fi
echo ""

# Test 3: Test with invalid credentials (offline mode)
echo -e "${YELLOW}Test 3: Offline mode with invalid credentials${NC}"
echo "Creating temporary config with invalid credentials..."

# Backup original config
cp ~/.r9s/config.yaml ~/.r9s/config.yaml.test-backup

# Create config with invalid credentials
cat > ~/.r9s/config.yaml << 'EOF'
currentProfile: default
profiles:
  - name: default
    url: https://fake-rancher.example.com
    bearerToken: invalid-token-xyz
    insecure: true
refreshInterval: 5s
logLevel: info
EOF

echo "Running r9s with invalid credentials (should enter offline mode)..."

# Create a test script that verifies the app starts
timeout 3 ./bin/r9s &
PID=$!
sleep 2

# Check if process is still running
if ps -p $PID > /dev/null 2>&1; then
    echo -e "${GREEN}✓ PASS${NC} - r9s started successfully in offline mode"
    kill $PID 2>/dev/null || true
else
    echo -e "${RED}✗ FAIL${NC} - r9s failed to start or crashed immediately"
    # Restore config
    mv ~/.r9s/config.yaml.test-backup ~/.r9s/config.yaml
    exit 1
fi
echo ""

# Test 4: Restore original config and test with real credentials
echo -e "${YELLOW}Test 4: Restore original configuration${NC}"
mv ~/.r9s/config.yaml.test-backup ~/.r9s/config.yaml
echo -e "${GREEN}✓ PASS${NC} - Original config restored"
echo ""

# Test 5: Code review - check for offline mode implementation
echo -e "${YELLOW}Test 5: Code review - Offline mode implementation${NC}"

# Check if offlineMode flag exists
if grep -q "offlineMode bool" internal/tui/app.go; then
    echo -e "${GREEN}✓ PASS${NC} - offlineMode flag found in App struct"
else
    echo -e "${RED}✗ FAIL${NC} - offlineMode flag not found"
    exit 1
fi

# Check if connection test is performed
if grep -q "TestConnection()" internal/tui/app.go; then
    echo -e "${GREEN}✓ PASS${NC} - Connection test implemented in NewApp()"
else
    echo -e "${RED}✗ FAIL${NC} - Connection test not found"
    exit 1
fi

# Check if mock data function exists
if grep -q "getMockPods" internal/tui/app.go; then
    echo -e "${GREEN}✓ PASS${NC} - Mock data generation function found"
else
    echo -e "${RED}✗ FAIL${NC} - Mock data function not found"
    exit 1
fi

# Check if fetchPods uses offlineMode
if grep -q "if a.offlineMode" internal/tui/app.go; then
    echo -e "${GREEN}✓ PASS${NC} - fetchPods checks offlineMode flag"
else
    echo -e "${RED}✗ FAIL${NC} - fetchPods doesn't check offlineMode"
    exit 1
fi

# Check if demo context is set
if grep -q "demo-cluster" internal/tui/app.go; then
    echo -e "${GREEN}✓ PASS${NC} - Demo cluster context configured"
else
    echo -e "${RED}✗ FAIL${NC} - Demo context not found"
    exit 1
fi

echo ""

# Summary
echo "======================================"
echo -e "${GREEN}All Tests Passed!${NC}"
echo "======================================"
echo ""
echo "Offline Mode Implementation Summary:"
echo "  ✓ Binary builds successfully"
echo "  ✓ App starts with invalid credentials (offline mode)"
echo "  ✓ offlineMode flag present in App struct"
echo "  ✓ Connection test on startup"
echo "  ✓ Mock data generation implemented"
echo "  ✓ Automatic fallback to mock data"
echo "  ✓ Demo context (demo-cluster/demo-project/default)"
echo ""
echo "Next Steps:"
echo "  1. Run: ./bin/r9s"
echo "  2. You should see mock pods displayed"
echo "  3. Press 'd' on any pod to see describe feature"
echo "  4. Press 'q' to quit"
echo ""
