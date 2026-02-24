# r8s

> **r8s v0.9.0 — AI-Powered kubectl for Rancher bundles.**

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev)

r8s (pronounced "rates") is an intelligent CLI tool for analyzing Rancher support bundles. It combines kubectl-like navigation with AI pattern detection to find root causes instantly.

**Latest: v0.9.0** (March 2026) — AI Intelligence

- **AI Analysis**: Detects 19+ issue patterns (CrashLoop, OOM, etcd, certs)
- **Natural Language Queries**: Ask `r8s ask "why is nginx crashing?"`
- **Root Cause Hints**: Explains *why* something broke and *how* to fix it
- **CI/CD Integration**: Export findings to SARIF, JUnit, Markdown
- **kubectl-compatible**: `get`, `logs`, `describe` work offline

---

## 🚀 Quick Start

```bash
# 1. Install
git clone https://github.com/Rancheroo/r8s.git && cd r8s
make build

# 2. Analyze Bundle (AI Powered)
./bin/r8s analyze ./support-bundle/
# 🔴 Found: CrashLoopBackOff (Container nginx has crashed 5 times...)

# 3. Ask Questions (Natural Language)
./bin/r8s ask ./support-bundle/ "why is nginx crashing?"

# 4. Export for GitHub Security
./bin/r8s export ./support-bundle/ --format=sarif --output=results.sarif

# 5. Traditional kubectl commands
./bin/r8s get pods ./support-bundle/
./bin/r8s logs ./support-bundle/ nginx-pod
```

---

## ✨ New AI Features (v0.9.0)

### 🧠 Pattern Detection
Automatically detects 19+ common Kubernetes issues:
- **CrashLoopBackOff / OOMKilled**
- **ImagePullBackOff / ErrImagePull**
- **etcd corruption / latency / quorum loss**
- **Certificate expiration / invalid CA**
- **CNI plugin errors / DNS failures**
- **PVC binding failures / Storage pressure**
- **Node NotReady / DiskPressure**

### 🗣️ Natural Language Queries
Troubleshoot like a human:
```bash
r8s ask ./bundle/ "why is etcd slow?"
r8s ask ./bundle/ "show me all certificate issues"
r8s ask ./bundle/ "which pods are pending?"
```

### 📤 Export Formats
Integrate with your ecosystem:
- **SARIF**: GitHub Advanced Security
- **JUnit**: Jenkins / GitHub Actions / GitLab CI
- **Markdown**: Human-readable reports
- **JSON**: Custom automation

```bash
r8s export ./bundle/ --format=sarif > results.sarif
r8s export ./bundle/ --format=junit > test-results.xml
```

---

## kubectl Compatibility

| kubectl | r8s | Status |
|---------|-----|--------|
| `kubectl get pods` | `r8s get pods ./bundle/` | ✅ Full |
| `kubectl logs pod` | `r8s logs ./bundle/ pod` | ✅ Full |
| `kubectl describe pod` | `r8s describe pod ./bundle/ pod` | ✅ Full |
| `kubectl get nodes` | `r8s get nodes ./bundle/` | ✅ Full |

---

## Commands

### `r8s validate` — Check Bundle Health
```bash
r8s validate ./support-bundle/
# Output: Bundle completeness %, missing files, type detection

r8s validate ./bundle/ --format=json | jq '.completeness'
# CI-friendly validation

# Exit codes:
# 0 = Valid bundle (all critical files present)
# 1 = Incomplete but usable (missing optional files)
# 2 = Error (invalid bundle or path)
```

### `r8s get` — Get Resources
```bash
r8s get pods ./bundle/
r8s get nodes ./bundle/
r8s get pods ./bundle/ -n kube-system
r8s get pods ./bundle/ -o yaml
```

### `r8s logs` — Stream Pod Logs
```bash
r8s logs ./bundle/ nginx-pod
r8s logs ./bundle/ nginx-pod -c sidecar
r8s logs ./bundle/ nginx-pod --tail=100
r8s logs ./bundle/ nginx-pod --follow
```

### `r8s describe` — Resource Details
```bash
r8s describe pod ./bundle/ nginx-pod
r8s describe node ./bundle/ worker-1
r8s describe deployment ./bundle/ app -o json
```

### `r8s export` — Export Findings
```bash
r8s export ./bundle/ --format=json > findings.json
r8s export ./bundle/ --format=yaml > findings.yaml
```

### `r8s analyze` — Detect Issues
```bash
r8s analyze ./bundle/
r8s analyze ./bundle/ --format=json | jq '.issues[] | select(.severity=="critical")'
```

### `r8s generate` — AI Prompts
```bash
r8s generate prompt ./bundle/ > ai-analysis.md
```

### `r8s completion` — Shell Completion
```bash
r8s completion bash > /etc/bash_completion.d/r8s
r8s completion zsh > "${fpath[1]}/_r8s"
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success, no issues |
| 1 | Issues found or incomplete bundle |
| 2 | Error (invalid args, file not found) |

**CI/CD Example:**
```yaml
- name: Validate Bundle
  run: |
    r8s validate ./bundle/
    if [ $? -eq 2 ]; then exit 1; fi
```

---

## Installation

**Requirements:** Go 1.23+

```bash
# Build from source
git clone https://github.com/Rancheroo/r8s.git
cd r8s
make build

# Binary: ./bin/r8s
```

**Shell Completion:**
```bash
# Bash
r8s completion bash | sudo tee /etc/bash_completion.d/r8s

# Zsh
r8s completion zsh > "${fpath[1]}/_r8s"
```

---

## Workflows

### Support Engineer
```bash
# 1. Customer sends bundle
tar -xzf rke2-support-bundle-*.tar.gz

# 2. Quick health check
r8s validate ./extracted-bundle/

# 3. Find crashed pods
r8s get pods ./bundle/ | grep -i crash

# 4. Check logs
r8s logs ./bundle/ problematic-pod --tail=50

# 5. Export for escalation
r8s export ./bundle/ --format=json > case-12345.json
```

### CI/CD Pipeline
```yaml
- name: Bundle Validation
  run: |
    r8s validate ./artifacts/bundle/ --format=json > validation.json
    completeness=$(jq '.completeness' validation.json)
    if [ "$completeness" -lt 70 ]; then
      echo "Bundle only $completeness% complete"
      exit 1
    fi
```

### Automation Script
```bash
#!/bin/bash
BUNDLE="$1"

# Validate
r8s validate "$BUNDLE"
if [ $? -eq 2 ]; then exit 1; fi

# Extract critical issues
r8s analyze "$BUNDLE" --format=json | \
  jq '.issues[] | select(.severity=="critical")' > critical.json

# Generate AI summary
if [ -s critical.json ]; then
  r8s generate prompt "$BUNDLE" > summary.md
fi
```

---

## Documentation

- **[CLI Reference](docs/USAGE.md)** — Complete command guide
- **[Bundle Format](docs/BUNDLE-FORMAT.md)** — RKE2/K3s bundle structure
- **[Troubleshooting](TROUBLESHOOTING.md)** — Common issues
- **[Architecture](docs/ARCHITECTURE.md)** — Technical design
- **[Contributing](CONTRIBUTING.md)** — Development guide

---

## Troubleshooting

| Error | Solution |
|-------|----------|
| "not a directory" | Extract bundle: `tar -xzf bundle.tar.gz` |
| "failed to load bundle" | Point to extracted folder with `rke2/` or `k3s/` dir |
| "go: command not found" | Install Go 1.23+ or use Docker builder |

---

## What Happened to the Dashboard?

**v0.8.0 removed the TUI** (Bubble Tea dashboard). Why?

- **CLI is 80% of the value** — users want kubectl-compatible commands
- **Scriptability** — can't automate a TUI
- **Performance** — no startup delay
- **Simplicity** — faster development, fewer bugs

**Migration:**
- `r8s dashboard` → `r8s analyze` or `r8s validate`
- TUI navigation → `r8s get`, `r8s logs`, `r8s describe`
- Need the old dashboard? Use v0.7.1 or earlier

---

## Development

```bash
# Run from source
go run main.go --help

# Run tests
make test

# Build
make build

# Cross-compile
make build-all
```

---

## License

Apache License 2.0 — See [LICENSE](LICENSE)

---

## Acknowledgments

- [kubectl](https://kubernetes.io/docs/reference/kubectl/) — UX inspiration
- [Rancher](https://rancher.com/) — Kubernetes management platform

---

**Made with ⚡ for Kubernetes troubleshooters**

[Report Bug](https://github.com/Rancheroo/r8s/issues/new) | [Request Feature](https://github.com/Rancheroo/r8s/issues/new)