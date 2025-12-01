<!-- This is a machine-generated file. To regenerate it, run: make docs -->
# r8s CLI Reference

Complete command-line reference for r8s (Rancheroos).

## Table of Contents
- [Global Flags](#global-flags)
- [Commands](#commands)
  - [r8s (root)](#r8s-root)
  - [r8s tui](#r8s-tui)
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

The root command provides the simplest way to launch r8s - just point it at what you want to view.

### Synopsis
```bash
r8s [bundle-path] [flags]
```

### Flags
| Flag | Default | Description |
|------|---------|-------------|
| `--mockdata` | `false` | Enable demo mode with mock data (no API required) |

All [global flags](#global-flags) also apply.

### Examples

**Live cluster mode:**
```bash
# Launch TUI with default profile
r8s

# Use specific profile
r8s --profile=production

# Start in specific namespace
r8s --namespace=kube-system
```

**Demo mode:**
```bash
# Launch with mock data (no configuration needed)
r8s --mockdata
```

**Bundle mode:**
```bash
# Extract bundle first
tar -xzf support-bundle.tar.gz

# Then analyze extracted folder
r8s ./extracted-bundle-folder/

# Works with any path format
r8s ./example-log-bundle/w-guard-wg-cp-xyz/
r8s /tmp/support-bundles/bundle-001/
```

**Why use the root command?**
- Simplest UX - just `r8s ./bundle/`
- Auto-detects bundle vs live mode
- No need to remember `tui` subcommand
- Matches user mental model: "analyze this thing"

---

## r8s tui

Launch the interactive Terminal UI for browsing clusters or bundles.

**Note:** The root command `r8s` is simpler for most use cases. Use `r8s tui` when you need TUI-specific flags like `--bundle`.

### Synopsis
```bash
r8s tui [flags]
```

### Flags
| Flag | Default | Description |
|------|---------|-------------|
| `--mockdata` | `false` | Enable demo mode with mock data (no API required) |
| `--bundle` | | Path to extracted log bundle folder |

### Examples

**Live cluster mode:**
```bash
# Use default profile from config
r8s tui

# Use specific profile
r8s tui --profile=production

# Start in specific namespace
r8s tui --namespace=kube-system

# Skip TLS verification (dev only!)
r8s tui --insecure
```

**Demo mode:**
```bash
# Launch with mock data (no configuration needed)
r8s tui --mockdata
```

**Bundle mode:**
```bash
# Extract bundle first
tar -xzf support-bundle.tar.gz

# Analyze extracted folder
r8s tui --bundle=./w-guard-wg-cp-xyz/

# Works with relative paths
r8s tui --bundle=./example-log-bundle/extracted-folder/

# Works with absolute paths
r8s tui --bundle=/tmp/support-bundles/bundle-001/
```

### Keyboard Shortcuts

**Navigation:**
- `↑` / `k` - Move up
- `↓` / `j` - Move down
- `Enter` - Navigate into resource
- `Esc` - Go back
- `q` / `Ctrl+C` - Quit

**Views:**
- `1` - Switch to Pods view
- `2` - Switch to Deployments view
- `3` - Switch to Services view
- `C` - Jump to CRDs view

**Actions:**
- `d` - Describe resource (JSON)
- `l` - View logs (pods only)
- `r` / `Ctrl+R` - Refresh
- `?` - Show help

**Log viewing:**
- `/` - Search logs
- `n` - Next search result
- `N` - Previous search result
- `Ctrl+E` - Filter ERROR logs
- `Ctrl+W` - Filter WARN logs
- `Ctrl+A` - Show all logs
- `t` - Toggle tail mode

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

## Common Workflows

### First-Time Setup

```bash
# 1. Initialize config
r8s config init

# 2. Edit config with your credentials
r8s config edit

# 3. Validate config
r8s config validate

# 4. Launch TUI
r8s
```

### Multiple Rancher Environments

```bash
# Production
r8s --profile=production

# Staging
r8s --profile=staging

# Development
r8s --profile=dev
```

### Troubleshooting a Cluster

```bash
# 1. Get support bundle from RKE2 node
ssh node.example.com
sudo rke2 support-bundle

# 2. Download bundle
scp node.example.com:/tmp/rke2-support-bundle-*.tar.gz ./

# 3. Extract bundle
tar -xzf rke2-support-bundle-*.tar.gz

# 4. Analyze
r8s ./rke2-support-bundle-*/
```

### Demo/Screenshot Mode

```bash
# Launch with realistic mock data
r8s --mockdata

# Take screenshots for documentation
# Navigate through UI as normal
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
