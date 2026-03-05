# r8s v1.3.0 Release Notes

**Release Date:** 2026-03-05
**Codename:** Deterministic State & Strict Structure

## Overview

v1.3.0 introduces a major architectural improvement to the Natural Language Query (NLQ) engine (`r8s ask`). We have implemented **Strict Structure** for state queries, ensuring 100% reliability when asking about resource states like "running" or "ready".

## What's New

### 🧠 Strict State Handling (Refactor)
The `ask` command now distinguishes between **Issue Queries** (probabilistic AI) and **State Queries** (deterministic parsing):

- **"Which pods are running?"** → Now uses exact Kubernetes status parsing. 100% accurate.
- **"Which nodes are ready?"** → Now checks exact node conditions.
- **"Why is nginx crashing?"** → Continues to use the AI Pattern Engine for root cause analysis.

Previously, asking "which pods are running?" might have confusingly listed *failed* pods (because the AI was biased towards finding problems). Now it correctly lists healthy resources.

### 🛠️ Automation & Tooling
- **Automated PR Reviews:** New tooling to manage CodeRabbit feedback efficiently.
- **Release Automation:** Improved release scripts to support auto-creation of GitHub releases and multi-platform builds.

## Bug Fixes
- **Exit Codes:** `r8s analyze` now correctly returns exit code `1` for warnings (was incorrectly `0`).
- **Describe:** `r8s describe` now returns a proper error exit code when arguments are invalid.
- **Spinners:** Fixed a potential division-by-zero panic in the progress bar.
- **Tests:** Stricter test harness validation fails on unexpected pattern matches.

## Upgrade Notes
- No breaking changes.
- `r8s ask` is now reliable for positive assertions ("is it working?") as well as negative ones ("why is it broken?").

## Quick Start

```bash
# Download (Linux amd64)
curl -L -o r8s https://github.com/Rancheroo/r8s/releases/download/v1.3.0/r8s-linux-amd64
chmod +x r8s && sudo mv r8s /usr/local/bin/

# Download (Mac ARM64)
curl -L -o r8s https://github.com/Rancheroo/r8s/releases/download/v1.3.0/r8s-darwin-arm64
chmod +x r8s && sudo mv r8s /usr/local/bin/
```
