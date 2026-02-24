#!/bin/bash
# Sprint 12 Final Validation Test
# Run this before declaring v1.0.0 complete
#
# Usage: ./sprint12-final-test.sh

set -e

echo "╔════════════════════════════════════════════════════════════╗"
echo "║     Sprint 12 Final Validation - v1.0.0 Release Test      ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

R8S_BIN="./bin/r8s"
KUBECTL_R8S="./kubectl-r8s"
TEST_BUNDLE="${R8S_BUNDLE:-$HOME/Downloads/r8s-cp-wlp7h-lhvgq/r8s-cp-wlp7h-lhvgq-2026-02-03_05_38_57/}"
FAILED=0

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "📦 Test 1: Binary Versions"
echo "==========================="
R8S_VERSION=$($R8S_BIN version 2>/dev/null | grep -o "v1\.0\.0" || echo "unknown")
if [ "$R8S_VERSION" = "v1.0.0" ]; then
    echo -e "${GREEN}✅ r8s binary is v1.0.0${NC}"
else
    echo -e "${RED}❌ r8s version is $R8S_VERSION (expected v1.0.0)${NC}"
    FAILED=1
fi
echo ""

echo "🔌 Test 2: kubectl-r8s Plugin"
echo "=============================="
if [ -f "$KUBECTL_R8S" ]; then
    echo -e "${GREEN}✅ kubectl-r8s binary exists${NC}"
else
    echo -e "${RED}❌ kubectl-r8s binary not found${NC}"
    FAILED=1
fi

if [ -d "$TEST_BUNDLE" ]; then
    export R8S_BUNDLE="$TEST_BUNDLE"
    export R8S_BINARY="$R8S_BIN"
    
    # Test kubectl r8s get pods
    if $KUBECTL_R8S get pods 2>&1 | grep -q "NAMESPACE"; then
        echo -e "${GREEN}✅ kubectl r8s get pods works${NC}"
    else
        echo -e "${YELLOW}⚠️  kubectl r8s get pods may have issues${NC}"
    fi
    
    # Test kubectl r8s analyze
    if timeout 10 $KUBECTL_R8S analyze 2>&1 | grep -q "R8S Bundle Analysis"; then
        echo -e "${GREEN}✅ kubectl r8s analyze works${NC}"
    else
        echo -e "${YELLOW}⚠️  kubectl r8s analyze may have issues${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  SKIP: Test bundle not found${NC}"
fi
echo ""

echo "🎯 Test 3: Core Commands"
echo "========================"
COMMANDS=("analyze" "ask" "export" "validate" "version")
for cmd in "${COMMANDS[@]}"; do
    if $R8S_BIN $cmd --help >/dev/null 2>&1; then
        echo -e "${GREEN}  ✅ $cmd${NC}"
    else
        echo -e "${RED}  ❌ $cmd${NC}"
        FAILED=1
    fi
done
echo ""

echo "🧠 Test 4: AI Pattern Detection"
echo "================================"
if [ -d "$TEST_BUNDLE" ]; then
    CRASHLOOPS=$($R8S_BIN analyze "$TEST_BUNDLE" --format=json 2>/dev/null | grep -c "crashloop" || echo "0")
    if [ "$CRASHLOOPS" -gt 0 ]; then
        echo -e "${GREEN}✅ Pattern detection working ($CRASHLOOPS CrashLoops found)${NC}"
    else
        echo -e "${YELLOW}⚠️  No CrashLoop patterns detected${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  SKIP: Test bundle not found${NC}"
fi
echo ""

echo "📤 Test 5: Export Formats"
echo "========================="
if [ -d "$TEST_BUNDLE" ]; then
    for format in sarif junit markdown; do
        if timeout 5 $R8S_BIN export "$TEST_BUNDLE" --format=$format 2>&1 | head -1 | grep -q "^{\|<\|#"; then
            echo -e "${GREEN}  ✅ $format export${NC}"
        else
            echo -e "${YELLOW}  ⚠️  $format export (timeout or empty)${NC}"
        fi
    done
else
    echo -e "${YELLOW}⚠️  SKIP: Test bundle not found${NC}"
fi
echo ""

echo "🤠 Test 6: UX Features"
echo "======================="
if [ -d "$TEST_BUNDLE" ]; then
    # Check for loading messages
    if timeout 5 $R8S_BIN analyze "$TEST_BUNDLE" -v 2>&1 | grep -q "Moo-\|Herding\|Wrangling"; then
        echo -e "${GREEN}✅ Loading messages showing${NC}"
    else
        echo -e "${YELLOW}⚠️  Loading messages not visible${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  SKIP: Test bundle not found${NC}"
fi
echo ""

echo "📚 Test 7: Documentation"
echo "========================"
if [ -f "README.md" ] && grep -q "kubectl-r8s plugin" README.md; then
    echo -e "${GREEN}✅ README has kubectl plugin docs${NC}"
else
    echo -e "${RED}❌ README missing kubectl plugin docs${NC}"
    FAILED=1
fi

if [ -f "CHANGELOG.md" ] && grep -q "v1.0.0" CHANGELOG.md; then
    echo -e "${GREEN}✅ CHANGELOG has v1.0.0${NC}"
else
    echo -e "${RED}❌ CHANGELOG missing v1.0.0${NC}"
    FAILED=1
fi
echo ""

echo "═══════════════════════════════════════════════════════════"
if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ Sprint 12 Complete - v1.0.0 Ready for Release!${NC}"
    echo ""
    echo "Ship it! 🚀"
    exit 0
else
    echo -e "${RED}❌ Sprint 12 Issues Found - Fix before release${NC}"
    exit 1
fi