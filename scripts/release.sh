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
rm -f bin/r8s*

# Build for multiple platforms
echo "🔨 Building binaries..."
PLATFORMS=("linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64" "windows/amd64")

for PLATFORM in "${PLATFORMS[@]}"; do
    GOOS=${PLATFORM%/*}
    GOARCH=${PLATFORM#*/}
    OUTPUT_NAME="r8s-${GOOS}-${GOARCH}"
    
    if [ "$GOOS" == "windows" ]; then
        OUTPUT_NAME+=".exe"
    fi

    echo "   Building for $GOOS/$GOARCH..."
    env GOOS=$GOOS GOARCH=$GOARCH go build -ldflags "-X main.version=$VERSION" -o "bin/$OUTPUT_NAME" main.go
done

# Verify local binary (linux/amd64 or current)
if [ "$(go env GOOS)" == "linux" ]; then
    cp "bin/r8s-linux-$(go env GOARCH)" bin/r8s
    echo "✅ Verifying binary..."
    ./bin/r8s version
fi

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

# Find all binaries in bin/
FILES=$(find bin -type f -name "r8s-*")

if gh release view "$VERSION" &>/dev/null; then
    echo "   Release exists, updating notes and uploading binaries..."
    # If RELEASE_NOTES.md exists, update the notes
    if [ -f "RELEASE_NOTES.md" ]; then
        echo "   Updating release notes from RELEASE_NOTES.md..."
        gh release edit "$VERSION" --notes-file RELEASE_NOTES.md
    fi
    gh release upload "$VERSION" $FILES --clobber
else
    echo "   Creating new release..."
    NOTES_FLAG="--generate-notes"
    if [ -f "RELEASE_NOTES.md" ]; then
        NOTES_FLAG="--notes-file RELEASE_NOTES.md"
    fi
    gh release create "$VERSION" $FILES $NOTES_FLAG
fi

echo ""
echo "✅ Release $VERSION binary uploaded successfully!"
echo "🌐 Release URL: https://github.com/Rancheroo/r8s/releases/tag/$VERSION"
echo ""
echo "Next steps:"
echo "  1. Verify the binary downloads correctly from the release page"
echo "  2. Update CHANGELOG.md if not already done"
echo "  3. Announce the release"
