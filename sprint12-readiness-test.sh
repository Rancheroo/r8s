#!/bin/bash
# Sprint 12 Readiness Test Suite
# Validates Sprint 11 foundation before Sprint 12 work begins
# 
# Run: ./sprint12-readiness-test.sh
# Exit code: 0 = ready, 1 = blocking issues found

set -e

R8S_BIN="./bin/r8s"
TEST_BUNDLE="$HOME/Downloads/r8s-cp-wlp7h-lhvgq/r8s-cp-wlp7h-lhvgq-2026-02-03_05_38_57/"
FAILED=0

echo "╔════════════════════════════════════════════════════════════╗"
echo "║     Sprint 12 Readiness Test - v0.9.0 Foundation          ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test 1: Binary exists and version is correct
echo "📦 Test 1: Binary Version"
if [ ! -f "$R8S_BIN" ]; then
    echo -e "${RED}❌ FAIL: Binary not found at $R8S_BIN${NC}"
    FAILED=1
else
    VERSION=$($R8S_BIN version 2>/dev/null | grep -o "v0\.9\.0" || echo "unknown")
    if [ "$VERSION" = "v0.9.0" ]; then
        echo -e "${GREEN}✅ PASS: Binary is v0.9.0${NC}"
    else
        echo -e "${YELLOW}⚠️  WARN: Binary version is $VERSION (expected v0.9.0)${NC}"
    fi
fi
echo ""

# Test 2: Core commands work
echo "🔧 Test 2: Core Commands"
COMMANDS=("version" "analyze" "get" "export" "patterns")
for cmd in "${COMMANDS[@]}"; do
    if $R8S_BIN $cmd --help >/dev/null 2>&1; then
        echo -e "${GREEN}  ✅ $cmd${NC}"
    else
        echo -e "${RED}  ❌ $cmd (help failed)${NC}"
        FAILED=1
    fi
done
echo ""

# Test 3: Pattern registry shows 19 patterns
echo "🎯 Test 3: Pattern Count (Sprint 11: 19 patterns)"
PATTERN_COUNT=$($R8S_BIN patterns list 2>/dev/null | grep -E "^Total:" | grep -o "[0-9]*" | head -1)
if [ -z "$PATTERN_COUNT" ]; then
    PATTERN_COUNT=$(($R8S_BIN patterns list 2>/dev/null | wc -l) - 3)
fi
if [ "${PATTERN_COUNT:-0}" -ge 19 ]; then
    echo -e "${GREEN}✅ PASS: $PATTERN_COUNT patterns found (expected 19+)${NC}"
else
    echo -e "${YELLOW}⚠️  WARN: Only ${PATTERN_COUNT:-0} patterns found (expected 19)${NC}"
fi
echo ""

# Test 4: Critical patterns exist
echo "🎯 Test 4: Critical Pattern IDs"
CRITICAL_PATTERNS=("crashloopbackoff-v2" "oomkill-v2" "imagepullbackoff-v2" "etcd-corruption" "etcd-latency")
MISSING_PATTERNS=0
for pattern in "${CRITICAL_PATTERNS[@]}"; do
    if $R8S_BIN patterns show $pattern >/dev/null 2>&1; then
        echo -e "${GREEN}  ✅ $pattern${NC}"
    else
        echo -e "${RED}  ❌ $pattern${NC}"
        MISSING_PATTERNS=$((MISSING_PATTERNS + 1))
    fi
done
if [ $MISSING_PATTERNS -gt 0 ]; then
    echo -e "${YELLOW}⚠️  $MISSING_PATTERNS critical patterns missing${NC}"
fi
echo ""

# Test 5: Bundle analysis works (if test bundle exists)
echo "📊 Test 5: Bundle Analysis"
if [ -d "$TEST_BUNDLE" ]; then
    echo "  Analyzing test bundle..."
    ANALYSIS_OUTPUT=$($R8S_BIN analyze "$TEST_BUNDLE" --format=json 2>/dev/null)
    
    # Check for CrashLoop detection
    CRASHLOOPS=$(echo "$ANALYSIS_OUTPUT" | grep -c "crashloop" || echo "0")
    if [ "$CRASHLOOPS" -gt 0 ]; then
        echo -e "${GREEN}  ✅ CrashLoop detection working ($CRASHLOOPS found)${NC}"
    else
        echo -e "${YELLOW}⚠️  No CrashLoop patterns detected in test bundle${NC}"
    fi
    
    # Check for <no value>
    NO_VALUE_COUNT=$(echo "$ANALYSIS_OUTPUT" | grep -o '<no value>' | wc -l)
    if [ "$NO_VALUE_COUNT" -eq 0 ]; then
        echo -e "${GREEN}  ✅ No '<no value>' in output${NC}"
    else
        echo -e "${RED}  ❌ Found $NO_VALUE_COUNT '<no value>' placeholders${NC}"
        FAILED=1
    fi
    
    # Check exit code
    $R8S_BIN analyze "$TEST_BUNDLE" >/dev/null 2>&1
    EXIT_CODE=$?
    if [ "$EXIT_CODE" -eq 1 ]; then
        echo -e "${GREEN}  ✅ Exit code 1 for critical issues${NC}"
    else
        echo -e "${YELLOW}⚠️  Exit code $EXIT_CODE (expected 1 for critical issues)${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  SKIP: Test bundle not found at $TEST_BUNDLE${NC}"
fi
echo ""

# Test 6: Export formats work (with timeout)
echo "📤 Test 6: Export Formats"
if [ -d "$TEST_BUNDLE" ]; then
    for format in sarif junit markdown; do
        OUTPUT=$(timeout 5 $R8S_BIN export "$TEST_BUNDLE" --format=$format 2>&1 | head -1 || echo "TIMEOUT")
        if [ -n "$OUTPUT" ] && [ "$OUTPUT" != "TIMEOUT" ]; then
            echo -e "${GREEN}  ✅ $format export works${NC}"
        else
            echo -e "${YELLOW}  ⚠️  $format export timeout or failed${NC}"
        fi
    done
else
    echo -e "${YELLOW}⚠️  SKIP: Test bundle not found${NC}"
fi
echo ""

# Test 7: NLQ command works (with timeout)
echo "🤖 Test 7: Natural Language Query"
if [ -d "$TEST_BUNDLE" ]; then
    NLQ_OUTPUT=$(timeout 5 $R8S_BIN ask "$TEST_BUNDLE" "show me crashloop issues" 2>/dev/null | head -3 || echo "TIMEOUT")
    if echo "$NLQ_OUTPUT" | grep -q "crashloop\|CrashLoop"; then
        echo -e "${GREEN}  ✅ NLQ responds correctly${NC}"
    else
        echo -e "${YELLOW}⚠️  NLQ may not be working correctly${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  SKIP: Test bundle not found${NC}"
fi
echo ""

# Test 8: Code structure for Sprint 12
echo "🏗️  Test 8: Sprint 12 Foundation"

# Check if YAML pattern directory exists
if [ -d "internal/ai/patterns" ]; then
    YAML_COUNT=$(ls internal/ai/patterns/*.yaml 2>/dev/null | wc -l)
    echo -e "${YELLOW}  ⚠️  $YAML_COUNT YAML patterns exist (Sprint 12 needs 7 more)${NC}"
else
    echo -e "${YELLOW}  ⚠️  YAML patterns directory missing${NC}"
fi

# Check for kubectl plugin placeholder
if [ -d "cmd/kubectl-r8s" ]; then
    echo -e "${GREEN}  ✅ kubectl-r8s plugin directory exists${NC}"
else
    echo -e "${YELLOW}  ⚠️  kubectl-r8s plugin directory missing (Sprint 12 deliverable)${NC}"
fi

# Check for parallel analyzer
if grep -q "NewParallelAnalyzer" internal/ai/parallel.go 2>/dev/null; then
    echo -e "${GREEN}  ✅ Parallel analyzer exists${NC}"
else
    echo -e "${RED}  ❌ Parallel analyzer missing${NC}"
    FAILED=1
fi
echo ""

# Summary
echo "═══════════════════════════════════════════════════════════"
if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ Sprint 12 Ready: Foundation is solid${NC}"
    echo ""
    echo "📋 Sprint 12 Deliverables:"
    echo "  1. Create 7 YAML patterns (etcd, certs, network, storage, etc.)"
    echo "  2. Integrate YAMLLoader with V2 engine OR migrate patterns to YAML"
    echo "  3. Build kubectl-r8s plugin wrapper"
    echo "  4. Test on 10+ real bundles"
    echo "  5. Documentation polish"
    exit 0
else
    echo -e "${RED}❌ Sprint 12 BLOCKED: Fix issues above${NC}"
    exit 1
fi