<!-- This is a machine-generated file. To regenerate it, run: make docs -->
# r8s CLI Reference

Complete command-line reference for **r8s**, the CLI Automation Engine for Kubernetes Triage.

## Table of Contents
- [Global Flags](#global-flags)
- [Commands](#commands)
  - [r8s analyze](#r8s-analyze)
  - [r8s get](#r8s-get)
  - [r8s describe](#r8s-describe)
  - [r8s logs](#r8s-logs)
  - [r8s config](#r8s-config)
  - [r8s version](#r8s-version)
- [Environment Variables](#environment-variables)
- [Configuration File](#configuration-file)
- [Common Workflows](#common-workflows)

---

## Global Flags

These flags work with any r8s command:

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--config` | | `~/.r8s/config.yaml` | Path to config file |
| `--profile` | | Current profile | Rancher profile to use |
| `--insecure` | | `false` | Skip TLS certificate verification |
| `--verbose` | `-v` | `false` | Enable verbose error output |
| `--context` | | | Cluster context to start in |
| `--namespace` | `-n` | | Namespace to start in |

---

## r8s (root)

The root command provides the simplest way to launch r8s analysis.

### Synopsis
```bash
r8s [bundle-path] [flags]
```

### Examples
```bash
# Analyze a bundle
r8s ./support-bundle/

# Show help
r8s --help
```

---

## r8s analyze

**The core automation command.** Scans the bundle or cluster for known failure patterns (CrashLoops, OOMs, Etcd issues, Certificate expiries).

### Synopsis
```bash
r8s analyze [bundle-path] [flags]
```

### Flags
| Flag | Default | Description |
|------|---------|-------------|
| `--format` | `text` | Output format: `text`, `json`, `yaml`, `sarif` |
| `--severity` | | Filter by severity (`critical`, `warning`, `info`) |

### Examples
```bash
# Basic analysis
r8s analyze ./bundle/

# JSON output for automation
r8s analyze ./bundle/ --format=json | jq '.issues[] | select(.severity=="critical")'

# Generate SARIF report for GitHub Advanced Security
r8s analyze ./bundle/ --format=sarif > results.sarif
```

---

## r8s get

List resources in the bundle, mimicking `kubectl get`.

### Synopsis
```bash
r8s get [resource] [bundle-path] [flags]
```

### Examples
```bash
# List pods
r8s get pods ./bundle/ -n cattle-system

# List all nodes
r8s get nodes ./bundle/

# List all resources (slow)
r8s get all ./bundle/
```

---

## r8s describe

Show detailed information about a specific resource, mimicking `kubectl describe`.

### Synopsis
```bash
r8s describe [resource] [name] [bundle-path] [flags]
```

### Examples
```bash
# Describe a pod
r8s describe pod rancher-webhook-5d9b7 ./bundle/ -n cattle-system

# Describe a node
r8s describe node worker-1 ./bundle/
```

---

## r8s logs

View logs for a container, automatically resolving the correct log file from the bundle.

### Synopsis
```bash
r8s logs [pod-name] [bundle-path] [flags]
```

### Flags
| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--container` | `-c` | | Specific container to read logs from |
| `--previous` | `-p` | `false` | Read previous (crashed) container logs |

### Examples
```bash
# View logs
r8s logs rancher-webhook-5d9b7 ./bundle/

# View previous logs for a crashing pod
r8s logs rancher-webhook-5d9b7 ./bundle/ -p
```

---

## r8s ask (BETA)

Ask natural language questions about your bundle. This uses an AI-driven pattern engine to find issues and explain root causes.

> **Warning:** This command is in BETA. Results may vary. Always verify with `r8s analyze`.

### Synopsis
```bash
r8s ask [bundle-path] [question]
```

### Examples
```bash
# Root Cause Analysis
r8s ask ./bundle/ "why is nginx-pod crashing?"
r8s ask ./bundle/ "what caused the outage?"

# Issue Discovery
r8s ask ./bundle/ "show me imagepullbackoff issues"
r8s ask ./bundle/ "which certificates are expired?"

# Resource Status
r8s ask ./bundle/ "which nodes are not ready?"
r8s ask ./bundle/ "which pods are running?"
r8s ask ./bundle/ "what is wrong with worker-1?"
```

**Note:** Queries for "running" or "ready" states use strict checking against cluster data, while issue queries use the AI pattern engine.

---

## r8s config

Manage r8s configuration files.

### Subcommands

#### config init

Initialize a new configuration file with template.

```bash
r8s config init [flags]
```

**Examples:**
```bash
# Create config at default location (~/.r8s/config.yaml)
r8s config init

# Create config at custom location
r8s config init --config=/path/to/custom-config.yaml
```

#### config view

Display current configuration (tokens are masked).

```bash
r8s config view [flags]
```

**Examples:**
```bash
# View current config
r8s config view

# View specific profile
r8s config view --profile=production
```

#### config edit

Edit configuration in your $EDITOR.

```bash
r8s config edit [flags]
```

**Examples:**
```bash
# Edit in default EDITOR (vi/vim/nano/etc)
r8s config edit

# Set custom editor first
export EDITOR="code --wait"
r8s config edit
```

#### config validate

Validate configuration file syntax.

```bash
r8s config validate [flags]
```

**Examples:**
```bash
# Validate current config
r8s config validate

# Validate custom config
r8s config validate --config=/path/to/config.yaml
```

#### config set

Set a configuration value.

```bash
r8s config set <key> <value> [flags]
```

**Supported keys:**
- `url` - Rancher server URL
- `token` / `bearerToken` - Bearer token
- `insecure` - Skip TLS verification (true/false)
- `currentProfile` - Default profile name

**Examples:**
```bash
# Set URL for current profile
r8s config set url https://rancher.example.com

# Set bearer token
r8s config set token token-xxxxx:yyyyyyyy

# Enable insecure mode (dev only!)
r8s config set insecure true

# Change default profile
r8s config set currentProfile production

# Set value for specific profile
r8s config set url https://staging.example.com --profile=staging
```

---

## r8s version

Print version information.

```bash
r8s version
```

**Example output:**
```
r8s v0.2.1 (commit: ecd8967, built: 2025-11-28)
```

---

## Environment Variables

R8s respects the following environment variables:

| Variable | Description | Example |
|----------|-------------|---------|
| `RANCHER_URL` | Rancher server URL | `https://rancher.example.com` |
| `RANCHER_TOKEN` | Bearer token | `token-xxxxx:yyyyyyyy` |
| `EDITOR` | Text editor for `config edit` | `vim`, `code --wait`, `nano` |
| `HOME` | Used for default config location | `/home/user` |

**Priority order** (highest to lowest):
1. CLI flags (`--profile`, `--insecure`, etc.)
2. Environment variables
3. Config file values

---

## Configuration File

### Default Location
```
~/.r8s/config.yaml
```

### Format
```yaml
currentProfile: production

profiles:
  - name: production
    url: https://rancher.example.com
    bearerToken: token-xxxxx:yyyyyyyy
    insecure: false

  - name: staging
    url: https://rancher-staging.example.com
    bearerToken: token-zzzzz:staging-secret
    insecure: false

  - name: dev
    url: https://rancher-dev.local
    bearerToken: token-aaaaa:dev-secret
    insecure: true  # OK for dev with self-signed certs

refreshInterval: 5s
logLevel: info
```

### Profile Fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Profile identifier |
| `url` | Yes | Rancher server URL |
| `bearerToken` | Yes* | Bearer token (format: `token-xxxxx:secret`) |
| `accessKey` | Yes* | Access key (alternative to bearerToken) |
| `secretKey` | Yes* | Secret key (alternative to bearerToken) |
| `insecure` | No | Skip TLS verification (default: false) |

*Either `bearerToken` OR `accessKey`+`secretKey` required

---
### 1. Automated Triage (CI/CD)
```bash
# Check if bundle is valid
if ! r8s validate ./bundle/; then
  exit 1
fi

# Fail build on critical issues
r8s analyze ./bundle/ --severity=critical --exit-code
```

### 2. Rapid Root Cause Analysis
```bash
# 1. Ask the engine
r8s ask ./bundle/ "why is the ingress controller failing?"

# 2. Verify with logs
r8s logs ingress-controller-x8s7 ./bundle/ | grep "Error"
```

### 3. Deep Dive Investigation
```bash
# Launch TUI to explore
r8s ./bundle/
```
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error (see stderr for details) |
| 2 | Invalid command-line arguments |

---

## Getting Help

```bash
# General help
r8s --help

# Command-specific help
r8s tui --help
r8s config --help

# Subcommand help
r8s config set --help
```

---

## See Also

- [Bundle Format Documentation](BUNDLE-FORMAT.md)
- [Troubleshooting Guide](../TROUBLESHOOTING.md)
- [Architecture Guide](ARCHITECTURE.md)
- [Contributing Guide](../CONTRIBUTING.md)
