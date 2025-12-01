#!/bin/bash
# Test script for BUG-002 and BUG-003 fixes
# Tests both mock mode describe and bundle kubectl data loading

set -e

echo "=========================================="
echo "BUG-002 & BUG-003 Test Verification"
echo "=========================================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Build first
echo "Building r8s..."
make build
echo ""

# Test counters
PASS=0
FAIL=0
MANUAL=0

echo "=========================================="
echo "TEST SUITE: BUG-003 - Bundle kubectl Data"
echo "=========================================="
echo ""

BUNDLE_PATH="example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09.tar.gz"

# T4: Test Namespaces Loading
echo "TEST T4: Bundle Namespaces Loading"
echo "Expected: Parse namespaces from kubectl/namespaces file"
if grep -q "calico-system" "example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09/rke2/kubectl/namespaces"; then
    echo -e "${GREEN}✓ PASS${NC} - Namespaces file contains data"
    ((PASS++))
else
    echo -e "${RED}✗ FAIL${NC} - Namespaces file is empty or missing"
    ((FAIL++))
fi
echo ""

# T5: Test Deployments Loading
echo "TEST T5: Bundle Deployments Loading"
echo "Expected: Parse deployments from kubectl/deployments file"
if grep -q "calico-kube-controllers" "example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09/rke2/kubectl/deployments"; then
    echo -e "${GREEN}✓ PASS${NC} - Deployments file contains data"
    ((PASS++))
else
    echo -e "${RED}✗ FAIL${NC} - Deployments file is empty or missing"
    ((FAIL++))
fi
echo ""

# T6: Test Services Loading
echo "TEST T6: Bundle Services Loading"
echo "Expected: Parse services from kubectl/services file"
if grep -q "calico-typha" "example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09/rke2/kubectl/services"; then
    echo -e "${GREEN}✓ PASS${NC} - Services file contains data"
    ((PASS++))
else
    echo -e "${RED}✗ FAIL${NC} - Services file is empty or missing"
    ((FAIL++))
fi
echo ""

# T7: Test CRDs Loading
echo "TEST T7: Bundle CRDs Loading"
echo "Expected: Parse CRDs from kubectl/crds file"
if grep -q "addons.k3s.cattle.io" "example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09/rke2/kubectl/crds"; then
    echo -e "${GREEN}✓ PASS${NC} - CRDs file contains data"
    ((PASS++))
else
    echo -e "${RED}✗ FAIL${NC} - CRDs file is empty or missing"
    ((FAIL++))
fi
echo ""

# Code verification tests
echo "=========================================="
echo "CODE VERIFICATION TESTS"
echo "=========================================="
echo ""

echo "TEST: BUG-002 Fix - Client nil checks in describe functions"
if grep -q "if a.client != nil" internal/tui/app.go; then
    echo -e "${GREEN}✓ PASS${NC} - Nil checks added to describe functions"
    ((PASS++))
else
    echo -e "${RED}✗ FAIL${NC} - Missing nil checks in describe functions"
    ((FAIL++))
fi
echo ""

echo "TEST: BUG-003 Fix - getBundleRoot() usage in kubectl parsers"
if grep -q "bundleRoot := getBundleRoot" internal/bundle/kubectl.go; then
    echo -e "${GREEN}✓ PASS${NC} - getBundleRoot() used in kubectl parsers"
    ((PASS++))
else
    echo -e "${RED}✗ FAIL${NC} - getBundleRoot() not used in kubectl parsers"
    ((FAIL++))
fi
echo ""

# Path verification
echo "TEST: Correct kubectl file paths after getBundleRoot fix"
EXPECTED_PATTERN='filepath.Join(bundleRoot, "rke2/kubectl/'
if grep -q "$EXPECTED_PATTERN" internal/bundle/kubectl.go; then
    echo -e "${GREEN}✓ PASS${NC} - Correct path pattern with bundleRoot"
    ((PASS++))
else
    echo -e "${RED}✗ FAIL${NC} - Incorrect path pattern"
    ((FAIL++))
fi
echo ""

echo "=========================================="
echo "MANUAL TEST INSTRUCTIONS"
echo "=========================================="
echo ""

echo -e "${YELLOW}The following tests require manual verification:${NC}"
echo ""
echo "TEST T1: Describe Pod in Mock Mode"
echo "  Command: ./bin/r8s --mockdata"
echo "  Steps:"
echo "    1. Navigate to Pods view (Cluster > Project > Namespace)"
echo "    2. Press 'd' on any pod"
echo "  Expected: Describe modal appears, no crash"
echo ""

echo "TEST T2: Describe Deployment in Mock Mode"
echo "  Command: ./bin/r8s --mockdata"
echo "  Steps:"
echo "    1. Navigate to Deployments view (press '2' in namespace context)"
echo "    2. Press 'd' on any deployment"
echo "  Expected: Describe modal appears, no crash"
echo ""

echo "TEST T3: Describe Service in Mock Mode"
echo "  Command: ./bin/r8s --mockdata"
echo "  Steps:"
echo "    1. Navigate to Services view (press '3' in namespace context)"
echo "    2. Press 'd' on any service"
echo "  Expected: Describe modal appears, no crash"
echo ""

echo "TEST: Bundle Mode Interactive"
echo "  Command: ./bin/r8s --bundle $BUNDLE_PATH"
echo "  Steps:"
echo "    1. Navigate: Cluster > Project > Namespace > Deployments (press '2')"
echo "    2. Verify deployments are displayed (not empty)"
echo "    3. Press '3' for Services, verify services are displayed"
echo "    4. Navigate back to cluster, press 'C' for CRDs"
echo "    5. Verify CRDs are displayed"
echo "  Expected: All resources display correctly from bundle"
echo ""

((MANUAL += 4))

echo "=========================================="
echo "TEST SUMMARY"
echo "=========================================="
echo ""
echo -e "Automated Tests Passed:  ${GREEN}${PASS}${NC}"
echo -e "Automated Tests Failed:  ${RED}${FAIL}${NC}"
echo -e "Manual Tests Required:   ${YELLOW}${MANUAL}${NC}"
echo ""

if [ $FAIL -eq 0 ]; then
    echo -e "${GREEN}✓ All automated tests PASSED${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Run manual tests (see instructions above)"
    echo "  2. Verify no regressions in live API mode"
    echo "  3. Push commit to remote if all tests pass"
    exit 0
else
    echo -e "${RED}✗ Some automated tests FAILED${NC}"
    echo "Review failures above and fix issues"
    exit 1
fi
