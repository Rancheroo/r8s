# 🚨 THIS REPOSITORY HAS MOVED 🚨

**The official `r8s` repository has moved to rancherlabs/r8s:**
👉 **[https://github.com/rancherlabs/r8s](https://github.com/rancherlabs/r8s)** 👈

*This personal repository is now archived and read-only. Please update your bookmarks, clones, and remotes to point to the new official location.*

---

# r8s

> **r8s — CLI-Powered Automation Engine & kubectl for RKE2 & K3S Bundles with AI (of course)**

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev)

**r8s** (pronounced "rates") is **The CLI Automation Engine for RKE2 & K3S Triage**.

Stop scrolling through thousands of text files. **r8s** turns static support bundles into a live, queryable environment. It combines `kubectl` muscle memory with an AI-driven analysis engine, allowing you to **validate**, **triage**, and **resolve** issues before you even open a log file.

It's not just a viewer. It's a **pipeline-ready tool** designed for First Response automation. It is the log bundle swiss army knife with lasers. 🔦

---

## 🚀 Installation

**Linux / macOS:**
```bash
# Download the latest release automatically
LATEST_VERSION=$(curl -s https://api.github.com/repos/Rancheroo/r8s/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m | sed 's/x86_64/amd64/')

curl -L -o r8s "https://github.com/Rancheroo/r8s/releases/download/${LATEST_VERSION}/r8s-${LATEST_VERSION}-${OS}-${ARCH}"
chmod +x r8s
sudo mv r8s /usr/local/bin/
```

**Docker (Coming Soon):**
> 🐳 **Help Wanted:** We are looking for a contributor to containerize `r8s` for CI/CD usage. This is a [Great First Issue](https://github.com/Rancheroo/r8s/issues/112)!

---

## ⚡ Quick Start

Experience **r8s** in 60 seconds.

### 1. Collect a Support Bundle
Use the official Rancher log collector tool:
*   [Rancher Log Collector (Script)](https://rancher.com/docs/rancher/v2.6/en/troubleshooting/log-collection/)

### 2. Validate the Bundle
Ensure the customer provided a complete and valid bundle before you start.

```bash
r8s validate ./support-bundle/
# Output: ✓ Bundle validated (RKE2, 100% complete)
```

### 3. Analyze the Bundle
Instantly detect 19+ common failure patterns (CrashLoops, OOMs, expired certs).

```bash
r8s analyze ./support-bundle/
# Output:
# 🔴 [CRITICAL] Certificate Expired: serving-kube-apiserver.crt
# 🔴 [CRITICAL] CrashLoopBackOff: rancher-webhook-5d9b7 (Restarts: 5)
```

### 4. Ask Questions in Plain English (BETA)
No need to grep. Just ask.

> ⚠️ **Beta Feature:** This feature is experimental. Verify results manually.

```bash
r8s ask ./support-bundle/ "why is nginx crashing?"
r8s ask ./support-bundle/ "show me all expired certificates"
```

### 5. Explore like kubectl
Navigate the static bundle using familiar commands.

```bash
# List pods in a namespace
r8s get pods ./support-bundle/ -n cattle-system

# View logs (automatically finds the right file)
r8s logs ./support-bundle/ rancher-webhook-5d9b7

# Describe a resource
r8s describe pod ./support-bundle/ rancher-webhook-5d9b7
```

### 6. Integrate & Automate
Output JSON for your CI/CD pipelines or scripts.

```bash
# Extract critical issues
r8s analyze ./support-bundle/ --format=json | jq '.issues[] | select(.severity=="critical")'
```

---

## ✨ Key Features

*   **Offline First:** Works on air-gapped machines. No cluster access required.
*   **Zero Data Leaks:** All analysis is local. No data sent to the cloud.
*   **kubectl Muscle Memory:** Use `get`, `logs`, `describe` exactly as you know them.
*   **AI Pattern Detection:** Automatically finds etcd quorum loss, CNI failures, OOMKilled, and more.
*   **CI/CD Ready:** Exit codes and JSON output make automation easy.

---

## 🔥 First Response & Automation

r8s is designed to shift left. Use it to automate triage workflows.

**1. Auto-Validate Incoming Tickets**
Reject invalid bundles automatically before an engineer even looks at them.
```bash
if ! r8s validate ./bundle/; then
  echo "Please upload a complete support bundle."
  exit 1
fi
```

**2. Auto-Triage Critical Issues**
Pipe analysis results directly to your ticketing system or Slack.
```bash
r8s analyze ./bundle/ --format=json | jq -r '.issues[] | select(.severity=="critical") | .summary' | \
  while read issue; do
    slack-send "🚨 Critical Issue Detected: $issue"
  done
```

---

## 📚 Documentation

*   [CLI Reference](docs/USAGE.md)
*   [Bundle Format Guide](docs/BUNDLE-FORMAT.md)
*   [Architecture](docs/ARCHITECTURE.md)
*   [Contributing](CONTRIBUTING.md)

---

**Made with ⚡ for Kubernetes troubleshooters**
[Report Bug](https://github.com/Rancheroo/r8s/issues/new) | [Releases](https://github.com/Rancheroo/r8s/releases)
