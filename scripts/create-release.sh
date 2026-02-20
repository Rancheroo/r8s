#!/bin/bash
# GitHub Release Script for v0.8.0-alpha
# Run this to create a proper GitHub release with binaries

set -e

VERSION="v0.8.0-alpha"
REPO="Rancheroo/r8s"
BIN_DIR="bin"

echo "Creating GitHub release ${VERSION}..."

# Create the release with notes
gh release create ${VERSION} \
  --repo ${REPO} \
  --title "v0.8.0-alpha — kubectl for Rancher Bundles" \
  --notes-file GITHUB_RELEASE_v0.8.0-alpha.md \
  --prerelease

echo "Uploading binary artifacts..."

# Upload binaries
gh release upload ${VERSION} \
  --repo ${REPO} \
  ${BIN_DIR}/r8s-v0.8.0-alpha-linux-amd64 \
  ${BIN_DIR}/r8s-v0.8.0-alpha-linux-arm64 \
  ${BIN_DIR}/r8s-v0.8.0-alpha-darwin-amd64 \
  ${BIN_DIR}/r8s-v0.8.0-alpha-darwin-arm64 \
  ${BIN_DIR}/r8s-v0.8.0-alpha-windows-amd64.exe \
  ${BIN_DIR}/checksums.txt

echo "✅ Release ${VERSION} created successfully!"
echo ""
echo "Verify at: https://github.com/${REPO}/releases/tag/${VERSION}"
