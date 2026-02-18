#!/bin/bash
# release.sh — Fully Automated Release Script for r8s
# Usage: ./scripts/release.sh v0.8.1
# Requires: git, go, curl, jq (optional but recommended)

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO="Rancheroo/r8s"
BIN_DIR="bin"
PLATFORMS=("linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64" "windows/amd64")

# Logging functions
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[✓]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[!]${NC} $1"; }
log_error() { echo -e "${RED}[✗]${NC} $1"; }
log_step() { echo -e "\n${BLUE}═══ $1 ═══${NC}"; }

# Error handler
error_exit() {
    log_error "$1"
    exit 1
}

# Extract GitHub PAT from git remote
get_github_token() {
    local remote_url
    remote_url=$(git remote get-url origin 2>/dev/null) || error_exit "Not a git repository or no origin remote"
    
    local pat
    pat=$(echo "$remote_url" | grep -o 'github_pat_[^@]*') || true
    
    if [[ -z "$pat" ]]; then
        # Try environment variable
        if [[ -n "${GITHUB_TOKEN:-}" ]]; then
            echo "$GITHUB_TOKEN"
            return 0
        fi
        error_exit "No GitHub PAT found in git remote or GITHUB_TOKEN env var"
    fi
    
    echo "$pat"
}

# Get commit info
get_commit_info() {
    COMMIT=$(git rev-parse --short HEAD)
    DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    log_info "Commit: $COMMIT, Date: $DATE"
}

# Validate version format
validate_version() {
    local version=$1
    if [[ ! "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9]+)?$ ]]; then
        error_exit "Invalid version format: $version. Expected: vX.Y.Z or vX.Y.Z-alpha"
    fi
}

# Pre-flight checks
check_prerequisites() {
    log_step "Pre-flight Checks"
    
    # Check required tools
    command -v git >/dev/null 2>&1 || error_exit "git is required"
    command -v go >/dev/null 2>&1 || error_exit "go is required"
    command -v curl >/dev/null 2>&1 || error_exit "curl is required"
    
    # Check jq (optional but recommended)
    if command -v jq >/dev/null 2>&1; then
        HAS_JQ=true
        log_success "jq found (will use for JSON parsing)"
    else
        HAS_JQ=false
        log_warn "jq not found (JSON parsing will be limited)"
    fi
    
    # Get GitHub token
    log_info "Extracting GitHub token..."
    GITHUB_TOKEN=$(get_github_token)
    log_success "GitHub token acquired"
    
    # Verify token works
    log_info "Verifying GitHub token..."
    local response
    response=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: token $GITHUB_TOKEN" \
        "https://api.github.com/repos/$REPO")
    
    if [[ "$response" != "200" ]]; then
        error_exit "GitHub token verification failed (HTTP $response)"
    fi
    log_success "GitHub token verified"
    
    # Check working directory is clean
    if [[ -n $(git status --porcelain 2>/dev/null) ]]; then
        log_warn "Working directory has uncommitted changes"
        read -p "Continue anyway? (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            error_exit "Aborted by user"
        fi
    fi
    
    # Check tag doesn't already exist
    if git rev-parse "$VERSION" >/dev/null 2>&1; then
        log_warn "Tag $VERSION already exists locally"
        read -p "Delete and recreate? (y/N) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            git tag -d "$VERSION" 2>/dev/null || true
        else
            error_exit "Aborted by user"
        fi
    fi
    
    log_success "Pre-flight checks passed"
}

# Run tests
run_tests() {
    log_step "Running Tests"
    
    log_info "Running make ci..."
    if make ci 2>&1 | tee /tmp/ci-output.log; then
        log_success "All tests passed"
    else
        log_error "Tests failed. See /tmp/ci-output.log"
        exit 1
    fi
}

# Build binaries
build_binaries() {
    log_step "Building Binaries"
    
    mkdir -p "$BIN_DIR"
    rm -f "$BIN_DIR"/r8s-v*"$VERSION"* 2>/dev/null || true
    
    local go_version
    go_version=$(go version | awk '{print $3}')
    log_info "Go version: $go_version"
    
    for platform in "${PLATFORMS[@]}"; do
        local goos=${platform%/*}
        local goarch=${platform#*/}
        local output_name="r8s-${VERSION}-${goos}-${goarch}"
        
        if [[ "$goos" == "windows" ]]; then
            output_name="${output_name}.exe"
        fi
        
        log_info "Building $output_name..."
        
        local cgo_flag="CGO_ENABLED=0"
        if [[ "$goos" == "darwin" && "$goarch" == "arm64" ]]; then
            cgo_flag="CGO_ENABLED=1"
        fi
        
        if $cgo_flag GOOS="$goos" GOARCH="$goarch" \
            go build \
            -ldflags "-X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE" \
            -o "$BIN_DIR/$output_name" \
            main.go 2>&1 | tee /tmp/build-${goos}-${goarch}.log; then
            
            local size
            size=$(du -h "$BIN_DIR/$output_name" | cut -f1)
            log_success "Built $output_name ($size)"
        else
            error_exit "Build failed for $platform. See /tmp/build-${goos}-${goarch}.log"
        fi
    done
    
    log_success "All binaries built successfully"
}

# Generate checksums
generate_checksums() {
    log_step "Generating Checksums"
    
    cd "$BIN_DIR"
    sha256sum r8s-"$VERSION"-* > checksums.txt
    cd ..
    
    log_info "Checksums:"
    cat "$BIN_DIR/checksums.txt" | while read -r line; do
        echo "  $line"
    done
    
    log_success "Checksums generated"
}

# Create GitHub release
create_release() {
    log_step "Creating GitHub Release"
    
    log_info "Creating release $VERSION..."
    
    local release_body="## 🎉 $VERSION — r8s Release

**Release Date:** $(date -u +"%Y-%m-%d")
**Commit:** $COMMIT

### Installation

**Linux (amd64):**
\`\`\`bash
curl -L https://github.com/$REPO/releases/download/$VERSION/r8s-$VERSION-linux-amd64 -o r8s
chmod +x r8s
sudo mv r8s /usr/local/bin/
\`\`\`

**macOS (Intel):**
\`\`\`bash
curl -L https://github.com/$REPO/releases/download/$VERSION/r8s-$VERSION-darwin-amd64 -o r8s
chmod +x r8s
sudo mv r8s /usr/local/bin/
\`\`\`

**macOS (Apple Silicon):**
\`\`\`bash
curl -L https://github.com/$REPO/releases/download/$VERSION/r8s-$VERSION-darwin-arm64 -o r8s
chmod +x r8s
sudo mv r8s /usr/local/bin/
\`\`\`

**Windows:**
Download \`r8s-$VERSION-windows-amd64.exe\` and add to PATH.

### Verification

\`\`\`bash
r8s version
# Expected: $VERSION
\`\`\`

### Checksums

See \`checksums.txt\` for SHA256 hashes.

---

**Full Changelog:** https://github.com/$REPO/compare/v0.8.0-alpha...$VERSION"

    local response
    response=$(curl -s -X POST \
        -H "Authorization: token $GITHUB_TOKEN" \
        -H "Accept: application/vnd.github+json" \
        -H "Content-Type: application/json" \
        "https://api.github.com/repos/$REPO/releases" \
        -d "{
            \"tag_name\": \"$VERSION\",
            \"name\": \"$VERSION\",
            \"body\": $(echo "$release_body" | jq -Rs .),
            \"draft\": false,
            \"prerelease\": $(echo "$VERSION" | grep -qE 'alpha|beta|rc' && echo "true" || echo "false")
        }")
    
    if $HAS_JQ; then
        RELEASE_ID=$(echo "$response" | jq -r '.id // empty')
        UPLOAD_URL=$(echo "$response" | jq -r '.upload_url // empty')
    else
        RELEASE_ID=$(echo "$response" | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2)
        UPLOAD_URL=$(echo "$response" | grep -o '"upload_url":"[^"]*"' | cut -d'"' -f4)
    fi
    
    if [[ -z "$RELEASE_ID" ]] || [[ "$RELEASE_ID" == "null" ]]; then
        log_error "Failed to create release"
        log_error "Response: $response"
        exit 1
    fi
    
    log_success "Release created (ID: $RELEASE_ID)"
    log_info "URL: https://github.com/$REPO/releases/tag/$VERSION"
}

# Upload asset
upload_asset() {
    local file=$1
    local name=$2
    local content_type=${3:-"application/octet-stream"}
    
    log_info "Uploading $name..."
    
    local response
    response=$(curl -s -X POST \
        -H "Authorization: token $GITHUB_TOKEN" \
        -H "Accept: application/vnd.github+json" \
        -H "Content-Type: $content_type" \
        "https://uploads.github.com/repos/$REPO/releases/$RELEASE_ID/assets?name=$name" \
        --data-binary @"$file" 2>&1)
    
    if $HAS_JQ; then
        local state
        state=$(echo "$response" | jq -r '.state // empty')
        if [[ "$state" == "uploaded" ]]; then
            log_success "$name uploaded"
            return 0
        fi
    else
        if echo "$response" | grep -q '"state":"uploaded"'; then
            log_success "$name uploaded"
            return 0
        fi
    fi
    
    log_error "Failed to upload $name"
    log_error "Response: $response"
    return 1
}

# Upload all assets
upload_assets() {
    log_step "Uploading Assets"
    
    local failed=0
    
    for platform in "${PLATFORMS[@]}"; do
        local goos=${platform%/*}
        local goarch=${platform#*/}
        local filename="r8s-$VERSION-${goos}-${goarch}"
        
        if [[ "$goos" == "windows" ]]; then
            filename="${filename}.exe"
        fi
        
        if ! upload_asset "$BIN_DIR/$filename" "$filename"; then
            ((failed++))
        fi
    done
    
    # Upload checksums
    if ! upload_asset "$BIN_DIR/checksums.txt" "checksums.txt" "text/plain"; then
        ((failed++))
    fi
    
    if [[ $failed -gt 0 ]]; then
        error_exit "$failed asset(s) failed to upload"
    fi
    
    log_success "All assets uploaded"
}

# Verify release
verify_release() {
    log_step "Verifying Release"
    
    log_info "Checking release on GitHub..."
    
    local response
    response=$(curl -s -H "Authorization: token $GITHUB_TOKEN" \
        "https://api.github.com/repos/$REPO/releases/$RELEASE_ID")
    
    local asset_count
    if $HAS_JQ; then
        asset_count=$(echo "$response" | jq '.assets | length')
    else
        asset_count=$(echo "$response" | grep -o '"assets":\[' | wc -l)
    fi
    
    log_info "Release has $asset_count assets"
    
    # Verify checksums.txt is downloadable
    log_info "Verifying checksums.txt download..."
    if curl -s -L -o /dev/null -w "%{http_code}" \
        "https://github.com/$REPO/releases/download/$VERSION/checksums.txt" | grep -q "200"; then
        log_success "checksums.txt is accessible"
    else
        log_warn "checksums.txt may not be accessible yet (GitHub propagation delay)"
    fi
    
    log_success "Release verified"
}

# Print summary
print_summary() {
    log_step "Release Summary"
    
    echo ""
    echo -e "${GREEN}✅ Release $VERSION Complete!${NC}"
    echo ""
    echo "Release URL:"
    echo "  https://github.com/$REPO/releases/tag/$VERSION"
    echo ""
    echo "Assets:"
    for platform in "${PLATFORMS[@]}"; do
        local goos=${platform%/*}
        local goarch=${platform#*/}
        local filename="r8s-$VERSION-${goos}-${goarch}"
        if [[ "$goos" == "windows" ]]; then
            filename="${filename}.exe"
        fi
        echo "  ✓ $filename"
    done
    echo "  ✓ checksums.txt"
    echo ""
    echo "Quick Install (Linux):"
    echo "  curl -L https://github.com/$REPO/releases/download/$VERSION/r8s-$VERSION-linux-amd64 -o r8s"
    echo "  chmod +x r8s && sudo mv r8s /usr/local/bin/"
    echo ""
    echo -e "${YELLOW}Next Steps:${NC}"
    echo "  1. Verify the release page looks correct"
    echo "  2. Test installation on a clean machine"
    echo "  3. Announce to team"
    echo ""
}

# Main function
main() {
    echo -e "${BLUE}"
    echo "╔════════════════════════════════════════════════════════════╗"
    echo "║        r8s Release Script — Fully Automated               ║"
    echo "╚════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    
    # Get version from argument
    if [[ $# -eq 0 ]]; then
        echo "Usage: $0 <version>"
        echo "Example: $0 v0.8.1"
        exit 1
    fi
    
    VERSION=$1
    validate_version "$VERSION"
    
    log_info "Preparing release: $VERSION"
    
    # Get commit info
    get_commit_info
    
    # Run all steps
    check_prerequisites
    run_tests
    build_binaries
    generate_checksums
    create_release
    upload_assets
    verify_release
    print_summary
    
    log_success "Release process complete!"
}

# Run main
main "$@"
