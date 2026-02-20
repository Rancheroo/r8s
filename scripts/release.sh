#!/bin/bash
# r8s Release Automation Script
# Usage: ./scripts/release.sh [version] [branch]
# Example: ./scripts/release.sh v0.8.0-alpha feature/sprint10-ci-cleanup

set -e

VERSION=${1:-"v0.8.0-alpha"}
BRANCH=${2:-"$(git branch --show-current)"}

echo "═══ r8s Release Automation ═══"
echo "Version: $VERSION"
echo "Branch: $BRANCH"
echo ""

# 1. Verify clean working directory
if [ -n "$(git status --porcelain)" ]; then
    echo "❌ ERROR: Working directory not clean"
    git status --short
    exit 1
fi
echo "✓ Working directory clean"

# 2. Verify branch
CURRENT_BRANCH=$(git branch --show-current)
if [ "$CURRENT_BRANCH" != "$BRANCH" ]; then
    echo "❌ ERROR: Not on branch $BRANCH (currently on $CURRENT_BRANCH)"
    exit 1
fi
echo "✓ On branch: $BRANCH"

# 3. Pull latest
echo ""
echo "Pulling latest changes..."
git pull origin $BRANCH
echo "✓ Latest changes pulled"

# 4. Build verification
echo ""
echo "Building r8s..."
make clean
make build
echo "✓ Build successful"

# 5. Quick smoke test
echo ""
echo "Running smoke tests..."
./bin/r8s --help > /dev/null
./bin/r8s version > /dev/null
echo "✓ Smoke tests passed"

# 6. Create tag
echo ""
echo "Creating tag $VERSION..."
git tag -a $VERSION -m "Release $VERSION

CLI-First Architecture:
- 6 kubectl-style commands (validate, logs, describe, export, generate, dashboard)
- 80%+ kubectl parity
- Standardized exit codes (0/1/2)
- TUI stripped to dashboard only (4,898 lines deleted)

CI/CD:
- Simplified workflow (build + cross-platform)
- Tests run locally
- Feature branch CI triggers enabled

See CHANGELOG.md for details."

echo "✓ Tag created"

# 7. Push tag
echo ""
echo "Pushing tag to origin..."
git push origin $VERSION
echo "✓ Tag pushed"

# 8. Summary
echo ""
echo "═══ Release Complete ═══"
echo "Version: $VERSION"
echo "Tag: https://github.com/Rancheroo/r8s/releases/tag/$VERSION"
echo ""
echo "Next steps:"
echo "1. GitHub Actions will build release artifacts"
echo "2. Create release notes at: https://github.com/Rancheroo/r8s/releases/new?tag=$VERSION"
echo "3. Attach binaries from GitHub Actions artifacts"
echo ""
echo "🎉 $VERSION is live!"
