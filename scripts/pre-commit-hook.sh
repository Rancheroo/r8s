#!/bin/bash
# Pre-commit hook for r8s project
# Automatically runs quality checks before each commit
# Install: ln -s ../../scripts/pre-commit-hook.sh .git/hooks/pre-commit

set -e

echo "🔍 Running pre-commit checks for r8s..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

FAILED=0

# 1. Format check
echo "  📝 Checking Go formatting..."
if [ -n "$(gofmt -l .)" ]; then
    echo "${RED}  ❌ Code formatting issues detected:${NC}"
    gofmt -l .
    echo ""
    echo "  Run: ${YELLOW}gofmt -w .${NC} to fix"
    FAILED=1
else
    echo "${GREEN}  ✅ Formatting OK${NC}"
fi

# 2. Go vet
echo "  🔍 Running go vet..."
if ! go vet ./... 2>&1; then
    echo "${RED}  ❌ go vet failed${NC}"
    FAILED=1
else
    echo "${GREEN}  ✅ go vet passed${NC}"
fi

# 3. Run short tests
echo "  🧪 Running tests (short mode)..."
if ! go test ./... -short 2>&1; then
    echo "${RED}  ❌ Tests failed${NC}"
    FAILED=1
else
    echo "${GREEN}  ✅ Tests passed${NC}"
fi

# 4. Check for TODO/FIXME in staged changes
echo "  📋 Checking for TODO/FIXME in staged changes..."
TODO_COUNT=$(git diff --cached | grep -E '^\+.*TODO|FIXME' | wc -l)
if [ "$TODO_COUNT" -gt 0 ]; then
    echo "${YELLOW}  ⚠️  Found $TODO_COUNT TODO/FIXME comments in staged changes${NC}"
    echo "     Consider documenting in CODE_REVIEW_LOG.md"
    # Don't fail, just warn
else
    echo "${GREEN}  ✅ No TODO/FIXME in staged changes${NC}"
fi

# 5. Check for large files (>1MB)
echo "  📦 Checking for large files..."
LARGE_FILES=$(git diff --cached --stat | awk '$1 > 1000 {print $NF}' | grep -v '|' || true)
if [ -n "$LARGE_FILES" ]; then
    echo "${YELLOW}  ⚠️  Large files detected (>1MB):${NC}"
    echo "$LARGE_FILES"
    echo "     Consider using Git LFS for large binaries"
fi

# Final result
if [ $FAILED -eq 0 ]; then
    echo ""
    echo "${GREEN}✅ All pre-commit checks passed!${NC}"
    exit 0
else
    echo ""
    echo "${RED}❌ Pre-commit checks failed${NC}"
    echo ""
    echo "To bypass (not recommended): ${YELLOW}git commit --no-verify${NC}"
    exit 1
fi
