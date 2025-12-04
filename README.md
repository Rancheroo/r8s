# r8s

> **r8s = fastest bundle troubleshooter on earth** · Live cluster browser

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev)

r8s (pronounced "rates" or "Rancheroos") is a keyboard-driven terminal UI for browsing Rancher-managed Kubernetes clusters and analyzing RKE2 support bundles offline.

---

## Quick Start

### Demo Mode (Zero Setup)
```bash
git clone https://github.com/Rancheroo/r8s.git && cd r8s
make build
./bin/r8s --mockdata
```

### Live Cluster
```bash
# 1. Initialize config
./bin/r8s config init

# 2. Add your credentials
export EDITOR=vim  # or nano, code, etc.
./bin/r8s config edit

# 3. Launch
./bin/r8s
```

### Bundle Analysis (Offline)
```bash
# Extract bundle first
tar -xzf example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09.tar.gz

# Then analyze
./bin/r8s ./example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09/
```

---

## Keyboard Shortcuts

| Key | Action | Key | Action |
|-----|--------|-----|--------|
| `↑`/`↓` or `j`/`k` | Navigate | `Enter` | Select |
| `Esc` or `b` | Back | `q` | Quit |
| `d` | Describe (JSON) | `l` | View logs |
| `1`/`2`/`3` | Pods/Deploys/Services | `r` | Refresh |
| `/` | Search logs | `?` | Help |
| `g` | Jump to top (logs) | `G` | Jump to bottom (logs) |
| `w` | Toggle wrap (logs) | | |

---

## Installation

**Requirements:** Go 1.23+

```bash
# Build from source
git clone https://github.com/Rancheroo/r8s.git
cd r8s
make build

# Binary location: ./bin/r8s
```

---

## Features

✅ **Live Mode** - Browse Rancher clusters, projects, namespaces  
✅ **Bundle Mode** - Analyze RKE2 bundles offline (no API needed)  
✅ **Demo Mode** - Test with mock data (`--mockdata`)  
✅ **Resource Views** - Pods, Deployments, Services, CRDs  
✅ **Log Viewer** - Search, filter (ERROR/WARN), tail mode  
✅ **Describe** - Full JSON details for resources  
✅ **Multi-Profile** - Switch between Rancher environments  

---

## Common Workflows

**First-Time Setup:**
```bash
./bin/r8s config init && ./bin/r8s config edit && ./bin/r8s
```

**Bundle Troubleshooting:**
```bash
# Extract bundle first
tar -xzf rke2-support-bundle-*.tar.gz

# Then analyze
./bin/r8s ./rke2-support-bundle-*/
```

**Multiple Environments:**
```bash
./bin/r8s --profile=production
./bin/r8s --profile=staging  
./bin/r8s --profile=dev
```

---

## Documentation

- **[CLI Reference](docs/USAGE.md)** - Complete command documentation
- **[Bundle Format](docs/BUNDLE-FORMAT.md)** - RKE2 bundle structure
- **[Troubleshooting](TROUBLESHOOTING.md)** - Common issues & solutions
- **[Architecture](docs/ARCHITECTURE.md)** - Technical design
- **[Contributing](CONTRIBUTING.md)** - Development guide
- **[Lessons Learned](LESSONS-LEARNED.md)** - Project insights

---

## Configuration

**Default location:** `~/.r8s/config.yaml`

```yaml
currentProfile: production
profiles:
  - name: production
    url: https://rancher.example.com
    bearerToken: token-xxxxx:yyyyyyyy
    insecure: false
```

**Environment variables:**
```bash
export RANCHER_URL=https://rancher.example.com
export RANCHER_TOKEN=token-xxxxx:yyyyyyyy
```

See [docs/USAGE.md](docs/USAGE.md) for details.

---

## Troubleshooting

**Common issues:**

| Error | Solution |
|-------|----------|
| "could not open TTY" | Run from interactive terminal, not CI/pipe |
| "not a directory" | Extract bundle: `tar -xzf bundle.tar.gz` |
| "connection refused" | Check Rancher URL and network |
| "authentication failed" | Regenerate API token |
| "not a valid bundle" | Point to extracted folder with `rke2/` dir |

See [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for complete guide.

---

## Bug Reports

Found a bug? [Report it](https://github.com/Rancheroo/r8s/issues/new?template=bug_report.md) with:
- `r8s version` output
- Mode used (live/bundle/mock)
- Verbose output (`r8s --verbose tui ...`)
- Bundle details (if applicable)

---

## Development

```bash
# Run from source
go run main.go tui --mockdata

# Run tests
make test

# Build
make build
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

---

## License

Apache License 2.0 - See [LICENSE](LICENSE)

---

## Acknowledgments

- [k9s](https://k9scli.io/) - Inspiration
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Rancher](https://rancher.com/) - Multi-cluster management

---

**Made with ❤️ for the Rancher community**
