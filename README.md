# r8s

> **r8s v1.3.3 — AI-Powered kubectl for Rancher Bundles**

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev)

**r8s** (pronounced "rates") is an intelligent CLI tool for analyzing Rancher support bundles. It combines kubectl-like navigation with AI pattern detection to find root causes instantly — no cluster access required.

---

## 🚀 Installation

**Linux / macOS:**
```bash
curl -L -o r8s https://github.com/Rancheroo/r8s/releases/download/v1.3.3/r8s-v1.3.3-linux-amd64
chmod +x r8s
sudo mv r8s /usr/local/bin/
```
*(Replace `linux-amd64` with `darwin-amd64` (Intel) or `darwin-arm64` (Apple Silicon) as needed)*

---

## ⚡ Quick Start

Experience **r8s** in 60 seconds.

### 1. Collect a Support Bundle
Use the official Rancher log collector tool:
*   [Rancher Log Collector (Script)](https://rancher.com/docs/rancher/v2.6/en/troubleshooting/log-collection/)

### 2. Analyze the Bundle
Instantly detect 19+ common failure patterns (CrashLoops, OOMs, expired certs).

```bash
r8s analyze ./support-bundle/
# Output:
# 🔴 [CRITICAL] Certificate Expired: serving-kube-apiserver.crt
# 🔴 [CRITICAL] CrashLoopBackOff: rancher-webhook-5d9b7 (Restarts: 5)
```

### 3. Ask Questions in Plain English (BETA)
No need to grep. Just ask.

> ⚠️ **Beta Feature:** This feature is experimental. Verify results manually.

```bash
r8s ask ./support-bundle/ "why is nginx crashing?"
r8s ask ./support-bundle/ "show me all expired certificates"
```

### 4. Explore like kubectl
Navigate the static bundle using familiar commands.

```bash
# List pods in a namespace
r8s get pods ./support-bundle/ -n cattle-system

# View logs (automatically finds the right file)
r8s logs ./support-bundle/ rancher-webhook-5d9b7

# Describe a resource
r8s describe pod ./support-bundle/ rancher-webhook-5d9b7
```

### 5. Integrate & Automate
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

## 📚 Documentation

*   [CLI Reference](docs/USAGE.md)
*   [Bundle Format Guide](docs/BUNDLE-FORMAT.md)
*   [Architecture](docs/ARCHITECTURE.md)
*   [Contributing](CONTRIBUTING.md)

---

**Made with ⚡ for Kubernetes troubleshooters**
[Report Bug](https://github.com/Rancheroo/r8s/issues/new) | [Releases](https://github.com/Rancheroo/r8s/releases)
