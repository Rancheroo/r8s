#!/bin/bash
# Automated release script for r8s
# Usage: ./scripts/release.sh v0.7.0-sprint1

set -e

VERSION=$1

if [ -z "$VERSION" ]; then
    echo "Usage: ./scripts/release.sh <version>"
    echo "Example: ./scripts/release.sh v0.7.0-sprint1"
    exit 1
fi

echo "🚀 Building release $VERSION..."

# Ensure clean state
echo "📥 Fetching tags..."
git fetch --tags --force

# Checkout the tag
echo "📋 Checking out $VERSION..."
git checkout $VERSION

# Verify we're on the right version
CURRENT_VERSION=$(git describe --tags)
if [ "$CURRENT_VERSION" != "$VERSION" ]; then
    echo "⚠️  Warning: Current version is $CURRENT_VERSION, expected $VERSION"
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Clean previous builds
echo "🧹 Cleaning previous builds..."
rm -f bin/r8s

# Build
echo "🔨 Building binary..."
make build

# Verify binary works
echo "✅ Verifying binary..."
./bin/r8s version

# Check if gh CLI is available
if ! command -v gh &> /dev/null; then
    echo "❌ GitHub CLI (gh) not found. Install it:"
    echo "   https://cli.github.com/"
    exit 1
fi

# Check if logged in
echo "🔐 Checking GitHub authentication..."
gh auth status || exit 1

# Create release if it doesn't exist, otherwise upload
echo "📤 Creating/Uploading release $VERSION..."
if gh release view "$VERSION" &>/dev/null; then
    echo "   Release exists, uploading binary..."
    gh release upload "$VERSION" ./bin/r8s --clobber
else
    echo "   Creating new release..."
    gh release create "$VERSION" ./bin/r8s --generate-notes
fi

echo ""
echo "✅ Release $VERSION binary uploaded successfully!"
echo "🌐 Release URL: https://github.com/Rancheroo/r8s/releases/tag/$VERSION"
echo ""
echo "Next steps:"
echo "  1. Verify the binary downloads correctly from the release page"
echo "  2. Update CHANGELOG.md if not already done"
echo "  3. Announce the release"
