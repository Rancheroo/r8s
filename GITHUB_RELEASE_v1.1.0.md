# GitHub Release — v1.1.0

## r8s v1.1.0 — Swiss Army Knife for Rancher Log Automation

**Release Date:** 2026-03-04

---

## What's New

### 🆕 Bundle Format v1.1 Support

r8s now supports the latest Rancher log bundle format while maintaining full backward compatibility:

- **Virtualization Detection**: Automatically detects VM/container platform (KVM, VMware, Docker, WSL, etc.)
- **Versions File Parser**: Parses comprehensive system info, container images, and Helm releases
- **Memory Unit Parsing**: Handles Gi, Mi, Ki, B suffixes in memory files
- **Smart Path Fallback**: Automatically tries new locations first, falls back to legacy paths

### 🐛 Bug Fixes

- **Fixed**: `test-cluster` command now properly recognized
- **Fixed**: Command registration for `patterns` and `completion` commands

### 📊 Quality & Testing

- **Test Coverage**: 69.6% (exceeds 45% target)
- **Manual Testing**: 30/30 test cases passed
- **Integration Tests**: Verified with both old and new bundle formats
- **Code Review**: Approved by CodeRabbit AI

---

## Installation

### Linux

```bash
# Download
curl -LO https://github.com/Rancheroo/r8s/releases/download/v1.1.0/r8s-v1.1.0-linux-amd64

# Make executable
chmod +x r8s-v1.1.0-linux-amd64

# Move to PATH
sudo mv r8s-v1.1.0-linux-amd64 /usr/local/bin/r8s

# Verify
r8s version
```

### macOS

```bash
# Intel
curl -LO https://github.com/Rancheroo/r8s/releases/download/v1.1.0/r8s-v1.1.0-darwin-amd64

# Apple Silicon (M1/M2)
curl -LO https://github.com/Rancheroo/r8s/releases/download/v1.1.0/r8s-v1.1.0-darwin-arm64

chmod +x r8s-v1.1.0-darwin-*
sudo mv r8s-v1.1.0-darwin-* /usr/local/bin/r8s
```

### Windows

```powershell
# Download
Invoke-WebRequest -Uri "https://github.com/Rancheroo/r8s/releases/download/v1.1.0/r8s-v1.1.0-windows-amd64.exe" -OutFile "r8s.exe"

# Add to PATH or move to directory in PATH
```

---

## Quick Start

```bash
# Analyze a bundle
r8s analyze ./my-bundle/

# Ask natural language questions
r8s ask ./my-bundle/ "why is nginx crashing?"

# Export findings for CI/CD
r8s export ./my-bundle/ --format=sarif --output=findings.sarif

# Run automated diagnostics
r8s test-cluster ./my-bundle/

# Get kubectl-style output
r8s get pods ./my-bundle/ -n cattle-system
```

---

## Features

### Core Commands
- `analyze` — Detect issues with AI-powered pattern matching
- `ask` — Natural language queries about your bundle
- `export` — Export findings (SARIF, Markdown, JSON, JUnit)
- `get` — kubectl-compatible resource queries (pods, nodes)
- `logs` — Stream pod logs
- `validate` — Check bundle completeness
- `test-cluster` — Automated diagnostic tests
- `patterns` — Browse and search pattern definitions

### Output Formats
- Table (human-readable)
- JSON (for automation)
- SARIF (for CI/CD integration)
- Markdown (for reports)
- JUnit (for test frameworks)

---

## CI/CD Integration

```yaml
# Example GitHub Actions
- name: Analyze Bundle
  run: |
    r8s analyze ./bundle/ --format=json | jq '.critical'
    r8s export ./bundle/ --format=sarif --output=findings.sarif
    
- name: Upload SARIF
  uses: github/codeql-action/upload-sarif@v2
  with:
    sarif_file: findings.sarif
```

---

## Documentation

- [Full Documentation](https://github.com/Rancheroo/r8s/tree/main/docs)
- [Manual Test Plan](https://github.com/Rancheroo/r8s/blob/main/docs/testing/V1.1.0_MANUAL_TEST_PLAN.md)
- [Bundle Format Changes](https://github.com/Rancheroo/r8s/blob/main/docs/development/BUNDLE_FORMAT_CHANGES_v1.1.md)
- [CHANGELOG](https://github.com/Rancheroo/r8s/blob/main/CHANGELOG.md)

---

## Known Issues

See [GitHub Issues](https://github.com/Rancheroo/r8s/issues) for:
- UX improvements planned for v1.2
- Feature requests
- Bug reports

---

## Verification

Verify your download:

```bash
# Linux/macOS
sha256sum r8s-v1.1.0-linux-amd64
# Expected: [CHECKSUM_FROM_RELEASE]

# Windows
Get-FileHash r8s-v1.1.0-windows-amd64.exe -Algorithm SHA256
```

---

## Support

- 🐛 [Report Issues](https://github.com/Rancheroo/r8s/issues)
- 📖 [Documentation](https://github.com/Rancheroo/r8s/tree/main/docs)
- 💬 Discussions: GitHub Discussions

---

**r8s = Swiss Army Knife for Rancher Log Automation** ⚡
