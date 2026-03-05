# r8s

> **r8s v1.3.2 — AI-Powered kubectl for Rancher Bundles**

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev)

r8s (pronounced "rates") is an intelligent CLI tool for analyzing Rancher support bundles. It combines kubectl-like navigation with AI pattern detection to find root causes instantly — no cluster access required.

**Latest: v1.3.2** (March 2026) — Demo Ready & Polish

- **AI Analysis**: Detects 19+ issue patterns (CrashLoop, OOM, etcd, certs, CNI, storage)
- **Natural Language Queries**: Ask `r8s ask "why is nginx crashing?"` — no kubectl expertise needed
- **Root Cause Hints**: Explains *why* something broke and *how* to fix it
- **CI/CD Integration**: Export findings to JSON, YAML, Markdown
- **kubectl-Compatible**: `get`, `logs`, `describe` work exactly like kubectl — but offline
- **Never Blank**: Every command produces clear, helpful output

---

## 🚀 Quick Start

### Download Pre-built Binaries

**Linux:**
```bash
curl -L -o r8s https://github.com/Rancheroo/r8s/releases/download/v1.3.2/r8s-v1.3.2-linux-amd64
chmod +x r8s
sudo mv r8s /usr/local/bin/
```

**macOS (Intel):**
```bash
curl -L -o r8s https://github.com/Rancheroo/r8s/releases/download/v1.3.2/r8s-v1.3.2-darwin-amd64
chmod +x r8s
sudo mv r8s /usr/local/bin/
```

**macOS (Apple Silicon M1/M2/M3):**
```bash
curl -L -o r8s https://github.com/Rancheroo/r8s/releases/download/v1.3.2/r8s-v1.3.2-darwin-arm64
chmod +x r8s
sudo mv r8s /usr/local/bin/
```

### Build from Source

```bash
git clone https://github.com/Rancheroo/r8s.git && cd r8s
make build
# Binary: ./bin/r8s
```

---

## 💡 Tips & Tricks

r8s is designed to fit into the workflows you already use. Here are practical ways to get the most out of it.

### Ask Questions in Plain English

No need to memorize kubectl flags or log paths. Just ask:

```bash
r8s ask ./bundle/ "why is my pod crashing?"
r8s ask ./bundle/ "show me all certificate issues"
r8s ask ./bundle/ "what caused the outage?"
r8s ask ./bundle/ "which pods are pending and why?"
r8s ask ./bundle/ "explain the etcd problems"
```

This is the fastest way to triage — especially if you're new to a cluster or unfamiliar with the bundle structure.

### Use JSON Output for Scripting

Every command supports `--format=json` for automation:

```bash
# Extract critical issues
r8s analyze ./bundle/ --format=json | jq '.issues[] | select(.severity=="critical")'

# Count issues by severity
r8s analyze ./bundle/ --format=json | jq '[.issues[] | .severity] | group_by(.) | map({(.[0]): length}) | add'

# Get pod names in a namespace
r8s get pods ./bundle/ -n cattle-system -o json | jq -r '.[].name'
```

### Validate Before You Analyze

Bundles from customers are sometimes incomplete. Validate first:

```bash
r8s validate ./bundle/
# Shows completeness %, missing files, bundle type (RKE2/K3s)

# In scripts:
r8s validate ./bundle/ --format=json | jq '.completeness'
```

### Combine with Standard Unix Tools

r8s plays well with pipes, grep, and your existing toolkit:

```bash
# Find all crash-related pods
r8s get pods ./bundle/ | grep -i crash

# Search logs for specific errors
r8s logs ./bundle/ my-pod | grep -i "connection refused"

# Export and diff two bundles
diff <(r8s analyze ./bundle-old/ --format=json) <(r8s analyze ./bundle-new/ --format=json)
```

### Use Exit Codes in Scripts

r8s returns meaningful exit codes for automation:
- `0` — Success / bundle is healthy
- `1` — Issues found or incomplete bundle
- `2` — Error (invalid args, file not found)

```bash
r8s analyze ./bundle/
if [ $? -eq 1 ]; then
  echo "⚠️ Issues detected — escalating"
  r8s export ./bundle/ --format=markdown > incident-report.md
fi
```

### Onboard New Team Members

r8s is a great tool for engineers who aren't Kubernetes experts yet:

```bash
# They don't need to know kubectl commands
r8s ask ./bundle/ "what's wrong with this cluster?"

# Or explore like they would with kubectl
r8s get pods ./bundle/
r8s get nodes ./bundle/
r8s logs ./bundle/ problematic-pod
```

---

## 🔌 Integrate r8s Into Your Projects

### GitHub Actions

Run bundle analysis as part of your CI/CD pipeline:

```yaml
name: Bundle Analysis
on:
  workflow_dispatch:
    inputs:
      bundle_path:
        description: 'Path to support bundle'
        required: true

jobs:
  analyze:
    runs-on: ubuntu-latest
    steps:
      - name: Download r8s
        run: |
          curl -L -o r8s https://github.com/Rancheroo/r8s/releases/download/v1.2.1/r8s-v1.2.1-linux-amd64
          chmod +x r8s
          sudo mv r8s /usr/local/bin/

      - name: Validate Bundle
        run: |
          r8s validate ${{ inputs.bundle_path }} --format=json > validation.json
          completeness=$(jq '.completeness' validation.json)
          echo "Bundle completeness: ${completeness}%"
          if [ "$completeness" -lt 70 ]; then
            echo "::warning::Bundle is only ${completeness}% complete"
          fi

      - name: Analyze
        run: |
          r8s analyze ${{ inputs.bundle_path }} --format=json > analysis.json
          critical=$(jq '[.issues[] | select(.severity=="critical")] | length' analysis.json)
          echo "Critical issues: ${critical}"
          if [ "$critical" -gt 0 ]; then
            echo "::error::Found ${critical} critical issues"
            exit 1
          fi

      - name: Upload Report
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: bundle-analysis
          path: |
            validation.json
            analysis.json
```

### Cron Job for Nightly Analysis

Monitor bundles automatically:

```bash
#!/bin/bash
# /etc/cron.daily/r8s-nightly
BUNDLE_DIR="/var/log/rancher/bundles"
REPORT_DIR="/var/log/r8s-reports"
DATE=$(date +%Y-%m-%d)

mkdir -p "$REPORT_DIR"

for bundle in "$BUNDLE_DIR"/*/; do
  name=$(basename "$bundle")
  r8s analyze "$bundle" --format=json > "$REPORT_DIR/${name}-${DATE}.json"

  # Alert on critical issues
  critical=$(jq '[.issues[] | select(.severity=="critical")] | length' "$REPORT_DIR/${name}-${DATE}.json")
  if [ "$critical" -gt 0 ]; then
    echo "ALERT: ${critical} critical issues in ${name}" | mail -s "r8s Alert: ${name}" team@example.com
  fi
done
```

### Support Ticket Workflow

Generate reports for customer tickets:

```bash
#!/bin/bash
# Usage: ./triage.sh <bundle-path> <ticket-id>
BUNDLE="$1"
TICKET="$2"

echo "=== Triage for $TICKET ==="

# Step 1: Validate
r8s validate "$BUNDLE"
if [ $? -eq 2 ]; then
  echo "❌ Invalid bundle — ask customer to re-collect"
  exit 1
fi

# Step 2: Quick analysis
r8s analyze "$BUNDLE" --format=json > "/tmp/${TICKET}-analysis.json"

# Step 3: Generate human-readable report
r8s export "$BUNDLE" --format=markdown > "/tmp/${TICKET}-report.md"

# Step 4: Summarize
echo ""
echo "📋 Analysis complete. Files:"
echo "  JSON:     /tmp/${TICKET}-analysis.json"
echo "  Report:   /tmp/${TICKET}-report.md"
echo ""
echo "Quick look at critical issues:"
jq -r '.issues[] | select(.severity=="critical") | "  🔴 \(.title)"' "/tmp/${TICKET}-analysis.json"
```

### Slack / Webhook Integration

Post analysis results to Slack:

```bash
#!/bin/bash
BUNDLE="$1"
WEBHOOK_URL="${SLACK_WEBHOOK_URL}"

# Analyze
result=$(r8s analyze "$BUNDLE" --format=json)
issues=$(echo "$result" | jq '[.issues[]] | length')
critical=$(echo "$result" | jq '[.issues[] | select(.severity=="critical")] | length')

# Post to Slack
curl -s -X POST "$WEBHOOK_URL" \
  -H 'Content-type: application/json' \
  -d "{
    \"text\": \"r8s Bundle Analysis\",
    \"blocks\": [{
      \"type\": \"section\",
      \"text\": {
        \"type\": \"mrkdwn\",
        \"text\": \"*Bundle Analysis Complete*\n• Total issues: ${issues}\n• Critical: ${critical}\n• Bundle: \`$(basename $BUNDLE)\`\"
      }
    }]
  }"
```

---

## 📖 Commands

### `r8s analyze` — AI-Powered Issue Detection
```bash
r8s analyze ./bundle/
r8s analyze ./bundle/ --format=json | jq '.issues[] | select(.severity=="critical")'
r8s analyze ./bundle/ --format=yaml
```

### `r8s ask` — Natural Language Queries
```bash
r8s ask ./bundle/ "why is etcd slow?"
r8s ask ./bundle/ "show me all certificate issues"
r8s ask ./bundle/ "which pods are pending?"
```

### `r8s get` — List Resources (kubectl-style)
```bash
r8s get pods ./bundle/
r8s get nodes ./bundle/
r8s get pods ./bundle/ -n kube-system
r8s get deploy ./bundle/ -o json
r8s get events ./bundle/
```

Supported resources: `pods`, `nodes`, `namespaces`, `deployments`, `services`, `events`

### `r8s logs` — View Pod Logs
```bash
r8s logs ./bundle/ nginx-pod
r8s logs ./bundle/ nginx-pod -c sidecar
r8s logs ./bundle/ nginx-pod --tail=100
```

### `r8s describe` — Resource Details
```bash
r8s describe pod ./bundle/ nginx-pod
r8s describe node ./bundle/ worker-1
```

### `r8s validate` — Check Bundle Health
```bash
r8s validate ./bundle/
r8s validate ./bundle/ --format=json | jq '.completeness'
```

### `r8s export` — Export Findings
```bash
r8s export ./bundle/ --format=json > findings.json
r8s export ./bundle/ --format=yaml > findings.yaml
r8s export ./bundle/ --format=markdown > report.md
```

### `r8s generate prompt` — AI-Ready Diagnostics
```bash
r8s generate prompt ./bundle/ > ai-analysis.md
```

### `r8s test-cluster` — Run Diagnostic Checks
```bash
r8s test-cluster   # Runs 7 diagnostic checks automatically
```

### `r8s completion` — Shell Completion
```bash
# Bash
r8s completion bash | sudo tee /etc/bash_completion.d/r8s

# Zsh
r8s completion zsh > "${fpath[1]}/_r8s"
```

---

## 🧠 AI Pattern Detection

r8s automatically detects 19+ common Kubernetes issues:

**Workload Issues**
- CrashLoopBackOff / OOMKilled
- ImagePullBackOff / ErrImagePull
- Pods stuck in Terminating / Pending

**Cluster Infrastructure**
- etcd corruption / latency / quorum loss
- Certificate expiration / invalid CA
- Node NotReady / DiskPressure / MemoryPressure

**Networking & Storage**
- CNI plugin errors / DNS failures
- PVC binding failures / Storage pressure
- Service endpoint misconfigurations

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success, no issues |
| 1 | Issues found or incomplete bundle |
| 2 | Error (invalid args, file not found) |

---

## Troubleshooting

| Error | Solution |
|-------|----------|
| "bundle not found" | Check the path — r8s will show what you tried |
| "not a directory" | Extract the bundle first: `tar -xzf bundle.tar.gz` |
| "failed to load bundle" | Point to the extracted folder containing `rke2/` or `k3s/` |
| "go: command not found" | Install Go 1.23+ or download a pre-built binary |

---

## Documentation

- **[CLI Reference](docs/USAGE.md)** — Complete command guide
- **[Bundle Format](docs/BUNDLE-FORMAT.md)** — RKE2/K3s bundle structure
- **[Architecture](docs/ARCHITECTURE.md)** — Technical design
- **[Release Notes](docs/releases/v1.2.1.md)** — What's new in v1.2.1
- **[Contributing](CONTRIBUTING.md)** — Development guide

---

## Development

```bash
# Run from source
go run main.go --help

# Run tests
make test

# Build
make build

# Full CI checks (lint + test + coverage)
make ci

# Development checks (tidy + fmt + vet + lint)
make dev
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

[Report Bug](https://github.com/Rancheroo/r8s/issues/new) | [Request Feature](https://github.com/Rancheroo/r8s/issues/new) | [Releases](https://github.com/Rancheroo/r8s/releases)
