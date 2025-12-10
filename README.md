# r8s

> **r8s 0.3.5 — the fastest way to understand a broken Kubernetes cluster from a log bundle**

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev)

r8s (pronounced "rates") is a terminal UI for analyzing RKE2 support bundles. The **Attention Dashboard** instantly highlights critical issues the moment you open a bundle — no configuration needed.

**What's new in 0.3.5:** Bundle-only bliss · Live mode removed · Zero config · Instant demo bundle · Attention Dashboard first

---

## Quick Start

```bash
# 1. Install
git clone https://github.com/Rancheroo/r8s.git && cd r8s
make build

# 2. Try the demo
./bin/r8s  # Instantly loads embedded demo bundle

# 3. Analyze your bundle
tar -xzf support-bundle.tar.gz
./bin/r8s ./extracted-bundle/
```

**That's it.** No configuration, no API keys, no clusters needed.

---

## Features

✅ **Attention Dashboard** - See all cluster issues ranked by severity  
✅ **Bundle Analysis** - Works offline, no API required  
✅ **Demo Mode** - Embedded demo bundle (zero setup)  
✅ **Smart Log Analysis** - Detects crashes, OOM kills, connection failures  
✅ **Log Viewer** - Search, filter (ERROR/WARN), color-coded, word wrap  
✅ **Resource Views** - Pods, Deployments, Services, CRDs  
✅ **Describe** - Full JSON details for any resource  

---

## Keyboard Shortcuts

| Key | Action | Key | Action |
|-----|--------|-----|--------|
| `↑`/`↓` or `j`/`k` | Navigate | `Enter` | Drill down / View logs |
| `Esc` or `b` | Back | `q` | Quit |
| `d` | Describe (JSON) | `r` | Refresh |
| `/` | Search logs | `?` | Help |
| `g` | Jump to top (logs) | `G` | Jump to bottom (logs) |
| `w` | Toggle wrap (logs) | `Ctrl+E` | Filter errors only |

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

## Workflows

### First-Time Demo
```bash
./bin/r8s  # Instant embedded demo — no setup needed
```

### Analyze Production Bundle
```bash
# 1. Extract the bundle
tar -xzf rke2-support-bundle-*.tar.gz

# 2. Launch r8s
./bin/r8s ./w-guard-wg-cp-xyz-*/

# 3. Navigate the Attention Dashboard
#    - Press Enter on any issue to view pod logs
#    - Use Ctrl+E to filter to errors only
#    - Press ? for help
```

### Using the Example Bundle
```bash
./bin/r8s ./example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-12-04_09_15_57/
```

---

## Documentation

- **[Bundle Format](docs/BUNDLE-FORMAT.md)** - RKE2 bundle structure
- **[CLI Reference](docs/USAGE.md)** - Complete command documentation
- **[Troubleshooting](TROUBLESHOOTING.md)** - Common issues & solutions
- **[Architecture](docs/ARCHITECTURE.md)** - Technical design
- **[Contributing](CONTRIBUTING.md)** - Development guide
- **[Lessons Learned](LESSONS-LEARNED.md)** - Project wisdom

---

## Troubleshooting

**Common issues:**

| Error | Solution |
|-------|----------|
| "could not open TTY" | Run from interactive terminal, not CI/pipe |
| "not a directory" | Extract bundle: `tar -xzf bundle.tar.gz` |
| "failed to load bundle" | Point to extracted folder with `rke2/` dir |
| "no logs captured" | Some pods may not have logs in the bundle |

See [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for complete guide.

---

## Bug Reports

Found a bug? [Report it](https://github.com/Rancheroo/r8s/issues/new?template=bug_report.md) with:
- `r8s version` output
- Bundle details (if using custom bundle)
- Verbose output (`r8s -v /path/to/bundle`)

---

## Development

```bash
# Run from source (demo mode)
go run main.go

# Run with example bundle
go run main.go ./example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-12-04_09_15_57/

# Run tests
make test

# Build
make build
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

---

## What Happened to Live Mode?

As of v0.3.5, we removed live Rancher API support to focus 100% on bundle analysis. This decision came from user feedback: bundles are captured when clusters are broken, making them the #1 troubleshooting workflow.

**Benefits:**
- ✅ Zero configuration (no API tokens)
- ✅ Works offline
- ✅ Faster startup
- ✅ Simpler codebase (-1,200 lines)
- ✅ Better UX for the primary use case

If you need live cluster browsing, use v0.3.4 or earlier.

---

## License

Apache License 2.0 - See [LICENSE](LICENSE)

---

## Acknowledgments

- [k9s](https://k9scli.io/) - Inspiration
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Rancher](https://rancher.com/) - Kubernetes management platform

---

**Made with ❤️ for Kubernetes troubleshooters**
