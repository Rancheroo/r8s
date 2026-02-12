# CI/CD Pipeline Guide

**S4-CRITICAL-1: Automated Testing and Quality Gates**

This document describes the GitHub Actions CI/CD pipeline for r8s.

---

## Overview

The CI/CD pipeline automates testing, linting, and build verification for every code change. It ensures code quality and prevents regressions.

**Location:** `.github/workflows/ci.yml`

---

## Pipeline Jobs

### 1. Build & Test (`build-and-test`)
**Purpose:** Core verification that code compiles and passes tests

**Steps:**
1. Checkout repository
2. Setup Go 1.23
3. Download dependencies
4. Build the application (`make build`)
5. Run tests (`make test`)
6. Generate coverage report
7. Verify coverage ≥ 50%
8. Upload coverage artifacts

**Failure Conditions:**
- Build fails
- Tests fail
- Coverage below 50%

---

### 2. Lint (`lint`)
**Purpose:** Code quality and style checks

**Tool:** golangci-lint

**Checks:**
- Go code style
- Common errors
- Performance issues
- Security concerns

---

### 3. Cross-Platform Build (`cross-platform`)
**Purpose:** Verify builds work on multiple platforms

**Matrix:**
- Linux: amd64, arm64
- macOS: amd64, arm64

**Note:** Windows builds can be added if needed

---

### 4. Integration Test (`integration`)
**Purpose:** Test the built binary functions correctly

**Tests:**
- CLI help displays
- Version command works
- Basic smoke test

---

### 5. Documentation (`docs`)
**Purpose:** Ensure required documentation exists

**Checks:**
- README.md present
- ROADMAP_UPDATES.md present
- CHANGELOG.md present
- CONTRIBUTING.md present

---

## Triggers

The pipeline runs on:

| Event | Branches |
|-------|----------|
| Push | `main`, `sprint4` |
| Pull Request | `main`, `sprint4` |

---

## Configuration

### Environment Variables

| Variable | Default | Purpose |
|----------|---------|---------|
| `GO_VERSION` | `1.23` | Go version to use |
| `MIN_COVERAGE` | `50` | Minimum coverage percentage |

### Modifying Thresholds

Edit `.github/workflows/ci.yml`:

```yaml
env:
  GO_VERSION: '1.23'      # Change Go version
  MIN_COVERAGE: '50'      # Change coverage threshold
```

---

## Usage for Developers

### Before Pushing

Always run locally:

```bash
# Build
make build

# Test
make test

# Coverage
coverage
cat coverage.out | grep total

# Lint (if installed)
golangci-lint run
```

### Viewing CI Results

1. Go to GitHub repository
2. Click "Actions" tab
3. Select the workflow run
4. View individual job logs

### Understanding Failures

| Failure | Likely Cause | Fix |
|---------|--------------|-----|
| Build fails | Compilation error | Fix code, run `make build` |
| Tests fail | Test assertion failed | Check test output |
| Coverage fails | Coverage < 50% | Add tests |
| Lint fails | Style/quality issue | Run `golangci-lint run --fix` |

---

## Coverage Reports

### Accessing Reports

Coverage reports are uploaded as artifacts:

1. Go to Actions → Select run
2. Scroll to "Artifacts" section
3. Download "coverage-report-{sha}"
4. Open `coverage.html` in browser

### Coverage in PRs

Coverage summary appears in:
- GitHub Actions log
- Step summary (click job → Summary)
- Artifact upload

---

## Branch Protection

To enforce CI on PRs, enable branch protection:

1. Go to Settings → Branches
2. Add rule for `main` and `sprint4`
3. Check "Require status checks to pass"
4. Add checks:
   - `Build & Test`
   - `Lint`
   - `Cross-Platform Build`

---

## Extending the Pipeline

### Adding New Jobs

Example: Add performance benchmarks

```yaml
benchmark:
  name: Performance Benchmarks
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
      with:
        go-version: '1.23'
    - name: Run benchmarks
      run: go test -bench=. ./...
```

### Adding Notifications

Example: Slack notification on failure

```yaml
- name: Notify Slack on failure
  if: failure()
  uses: slackapi/slack-github-action@v1
  with:
    payload: |
      {"text": "CI failed for r8s: ${{ github.ref }}"}
  env:
    SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK }}
```

---

## Troubleshooting

### Job is stuck / queued

GitHub Actions may queue jobs during high demand. Wait or check [GitHub Status](https://www.githubstatus.com/).

### Cache not working

Go module caching is enabled. If you see cache misses:
- Check `go.sum` is committed
- Verify cache key in workflow

### Coverage report not uploading

Ensure coverage file exists:
```bash
ls -la coverage.out coverage.html
```

---

## Cost

GitHub Actions is **free** for public repositories:
- 2,000 minutes/month included
- r8s builds take ~2 minutes
- ~1,000 builds per month limit

Private repositories have different limits.

---

## Related

- Issue #33: S4-CRITICAL-1 CI/CD Pipeline
- Issue #35: S4-HIGH-5 50% Coverage Enforcement
- `Makefile` - Local build/test commands
- `.golangci.yml` - Lint configuration (if present)

---

**Last Updated:** 2026-02-12  
**Maintained By:** RancherSRE
