# Release v0.9.0 — AI Intelligence

**Date:** February 24, 2026
**Theme:** Smart Analysis & Root Cause Detection

r8s v0.9.0 transforms the tool from a simple bundle inspector into an intelligent diagnostic assistant. It introduces a powerful pattern matching engine that automatically detects 19+ common Kubernetes issues and provides actionable root cause analysis.

## 🌟 Headline Features

### 🧠 AI Pattern Engine
Automatically scans support bundles for known issue patterns.
- **CrashLoopBackOff**: Detects crash loops, extracts exit codes, and suggests fixes.
- **OOMKilled**: Identifies memory pressure and specific container kills.
- **etcd Issues**: Detects database corruption, high latency, and quorum loss.
- **Certificate Expiry**: Warns about expired or soon-to-expire certificates.
- **Networking**: Finds DNS failures, CNI plugin errors, and connectivity timeouts.
- **Storage**: Identifies PVC binding failures and disk pressure.

### 🗣️ Natural Language Queries
Ask questions in plain English:
```bash
r8s ask ./bundle/ "why is nginx crashing?"
r8s ask ./bundle/ "show me certificate issues"
```

### 📤 CI/CD Integration
Export analysis results to standard formats:
- **SARIF**: Native integration with GitHub Advanced Security.
- **JUnit**: Test reporting for Jenkins/GitLab CI.
- **Markdown**: Human-readable reports for ticket attachments.

## 🛠️ Changes

### New Commands
- `r8s analyze`: Now runs the AI engine (previously just checked file health).
- `r8s ask`: Query the bundle using natural language.
- `r8s export`: Export findings to SARIF, JUnit, or Markdown.
- `r8s patterns`: List and inspect the built-in pattern registry.

### Improvements
- **Parallel Analysis**: Uses goroutines for fast scanning of large bundles.
- **Regex Support**: Advanced extraction of pod names, namespaces, and error details.
- **False Positive Reduction**: Tuned patterns to avoid matching pod names as errors.

## 📦 Installation

```bash
git clone https://github.com/Rancheroo/r8s.git
cd r8s
make build
```
