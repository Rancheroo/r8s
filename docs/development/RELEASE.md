# Release Process

This document outlines the standard release process for r8s.

## Versioning

We follow [Semantic Versioning](https://semver.org/).
- **Major (X.y.z)**: Breaking changes
- **Minor (x.Y.z)**: New features (backwards compatible)
- **Patch (x.y.Z)**: Bug fixes

## Release Steps

### 1. Pre-Release Checks
- Ensure the main branch builds and passes tests: `make test`
- Check that `CHANGELOG.md` is updated (if manual)
- Ensure all PRs are merged and CodeRabbit reviews are addressed

### 2. Create Release
Use the automated script to create a release. This script handles:
- Tagging (if you haven't already) - *Note: currently script expects tag to exist*
- Building the binary
- Verifying the binary
- Creating the GitHub Release
- Uploading artifacts

```bash
# 1. Tag the release
git tag v1.3.0
git push origin v1.3.0

# 2. Run release script
./scripts/release.sh v1.3.0
```

The script will ask for confirmation if the tag doesn't match the current commit description.

## Automation Tools

### PR Review Bot
We use a script to help manage CodeRabbit review comments. This ensures we acknowledge every suggestion, even if just to say "Deferred".

```bash
# Interactive mode - iterate through unanswered comments
./scripts/review-reply.sh <PR_NUMBER>

# Batch mode - reply to ALL unanswered comments with same message
./scripts/review-reply.sh <PR_NUMBER> "Acknowledged. Will fix in next sprint."
```

## CI/CD
Releases are built locally by the release manager (you) using `scripts/release.sh` to ensure signing keys and environment are secure, then uploaded to GitHub.
Future improvements will move this to GitHub Actions.
