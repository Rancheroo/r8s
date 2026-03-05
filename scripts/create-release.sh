#!/bin/bash
# GitHub Release Script for r8s
# Usage: ./create-release.sh [version]
# Example: ./create-release.sh v1.2.1

set -e

VERSION="${1:-v1.2.1}"
REPO="Rancheroo/r8s"
RELEASE_DIR="releases"

echo "Creating GitHub release ${VERSION}..."
echo ""

# Verify binaries exist
for PLATFORM in linux-amd64 darwin-amd64 darwin-arm64; do
  BINARY="${RELEASE_DIR}/r8s-${VERSION}-${PLATFORM}"
  if [ ! -f "${BINARY}" ]; then
    echo "❌ Missing binary: ${BINARY}"
    echo "Run: GOOS=<os> GOARCH=<arch> go build -o ${BINARY} main.go"
    exit 1
  fi
  echo "✓ Found ${BINARY}"
done

# Generate checksums
echo ""
echo "Generating checksums..."
(cd ${RELEASE_DIR} && sha256sum r8s-${VERSION}-* > checksums-${VERSION}.txt)
echo "✓ checksums-${VERSION}.txt"

# Create the release
echo ""
echo "Creating GitHub release..."
gh release create ${VERSION} \
  --repo ${REPO} \
  --title "${VERSION} — Simplify & Never Blank" \
  --notes-file docs/releases/${VERSION#v}.md \
  --target main

echo "Uploading binary artifacts..."

# Upload binaries + checksums
gh release upload ${VERSION} \
  --repo ${REPO} \
  ${RELEASE_DIR}/r8s-${VERSION}-linux-amd64 \
  ${RELEASE_DIR}/r8s-${VERSION}-darwin-amd64 \
  ${RELEASE_DIR}/r8s-${VERSION}-darwin-arm64 \
  ${RELEASE_DIR}/checksums-${VERSION}.txt

echo ""
echo "✅ Release ${VERSION} created successfully!"
echo ""
echo "Verify at: https://github.com/${REPO}/releases/tag/${VERSION}"
