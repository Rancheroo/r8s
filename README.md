# r8s

> **r8s v0.6.9 "Principle Compliance Sprint" — the fastest way to understand a broken Kubernetes cluster from a log bundle**

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev)

r8s (pronounced "rates") is a terminal UI for analyzing RKE2 support bundles. The **Attention Dashboard** instantly highlights critical issues the moment you open a bundle — no configuration needed.

**Latest: v0.7.0.1** (February 2026)
- **CI Stability**: Quality gates restored, realistic coverage thresholds
- **Test Coverage**: 33% total repo, 61.4% internal/bundle package
- **Quality**: golangci-lint warnings addressed, all critical checks passing

**Previous: v0.7.0** (February 2026)
- **Bundle Coverage**: 90% data extraction (up from 30%)
- **New Parsers**: Storage (PV/PVC/StatefulSet), ConfigMaps, HelmCharts
- **System Health**: dmesg OOM detection, RKE2 journald analysis

**Recent v0.6.x highlights:**
- **0.6.8** - Smart event truncation + cluster-level drill-down for node/ETCD issues
- **0.6.7** - Inline diagnostics (2-line root cause + fix recommendations)
- **0.6.6** - OOM root cause detection and display
- **0.6.5** - Kubelet issue detection from journald logs

---

## 🚀 What's New

### v0.7.0.1 "CI Stability Patch" (February 2026) ✅ RELEASED
**For Developers**: Quality gates restored with realistic thresholds

**CI/CD Improvements**:
- **📊 Realistic Coverage**: 33% total repo threshold (was incorrectly set to 60%)
- **✅ Quality Gates Restored**: Build, test, coverage, and cross-platform checks
- **🧹 Lint Cleanup**: golangci-lint warnings addressed
- **🤖 CodeRabbit Integration**: Automated review feedback on every PR

**Key Lesson**: Package-level coverage (61.4% for internal/bundle) ≠ total repo coverage (~33%).

### v0.7.0 "Maximum Information Extraction" (February 2026) ✅ RELEASED
**For Support Teams**: Extract 60% more data from bundles + faster performance

**New Data Sources** (90% bundle coverage vs current 30%):
- **💾 Storage**: PersistentVolumes, PVCs, StatefulSets with health indicators
- **🖥️ System Health**: dmesg analysis (OOM kills, kernel panics), disk space monitoring
- **📋 Control Plane**: RKE2 server logs from journald with error detection
- **⚙️ Configuration**: ConfigMaps, HelmCharts (Rancher apps) with linkage to pods
- **📊 Bundle Completeness**: Indicator showing which data is available (e.g., "76% complete")

**Performance Improvements**:
- **⚡ Cache Optimization**: Eliminates 2-5 second dashboard loads on 800MB+ bundles
- **🚀 Async Operations**: Non-blocking CRD counting and resource fetching
- **📈 Namespace Health**: Pre-computed rankings show worst namespaces first

**Quality Infrastructure**:
- **🔄 CI/CD Pipeline**: Automated testing prevents regressions
- **📊 Quality Gates**: 30% total coverage, lint checks, cross-platform builds
- **🧪 Principle Compliance**: Automated checks on every commit

### v0.8.0 "Production Hardening + Advanced Diagnostics" (Target: August 2026)
**For Support Teams**: Enterprise-ready reliability + complete bundle analysis

**Additional Data Extraction** (90%+ bundle coverage):
- **🌐 Networking**: Ingress rules, Endpoints, Service→Pod mapping
- **⚙️ Workloads**: Jobs, CronJobs, ReplicaSets, HorizontalPodAutoscalers
- **🔧 etcd Health**: Complete cluster health, alarms (space quota), member status
- **🔍 Network Debugging**: iptables rules, routing tables for advanced troubleshooting

**Production Quality**:
- **🏆 80% Test Coverage**: Production-grade quality assurance
- **⚡ <2s Dashboard Load**: Even with 1000 pods (40-50% faster than v0.6.x)
- **📈 Memory Optimization**: <500MB for large bundles (vs current ~1GB)
- **💪 Zero Known Bugs**: All critical/high severity issues resolved

**Enterprise Features**:
- **📖 Complete Documentation**: Troubleshooting playbooks, deployment guides
- **🔒 Security Audit**: Complete security review
- **📊 Bundle Health Scoring**: Automated quality assessment (e.g., "72% complete - missing journald logs")
- **🎯 Stress Testing**: Validated with 1000+ pod clusters, 10M+ log lines

**Why These Matter for Support:**
- Faster troubleshooting = Faster resolution
- No crashes on large bundles = Analyze any cluster
- Better performance = Handle escalations under pressure
- Automated quality = Fewer tool bugs to work around

See [docs/development/ROADMAP_V0.6-V0.8.md](docs/development/ROADMAP_V0.6-V0.8.md) for complete roadmap.

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
| `m` | Expand/collapse dashboard | `c` | Classic cluster view |
| `/` | Search logs | `?` | Help |
| `g` | Jump to top | `G` | Jump to bottom |
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

# 2. Launch r8s (optionally with deep log scanning)
./bin/r8s ./w-guard-wg-cp-xyz-*/
# Or: ./bin/r8s --scan=1000 ./bundle/  # Scan 1000 lines per pod

# 3. Navigate the Attention Dashboard
#    - Shows top-20 critical issues by default
#    - Press 'm' to expand and see all issues (scroll with j/k)
#    - Press Enter on any issue to view pod logs
#    - Use Ctrl+E to filter to errors only
#    - Press ? for help
```

**Pro tip:** Use `--scan=500` or higher for large clusters. The dashboard smartly caps display to top-20 issues but you can press `m` to expand and scroll through all detected problems.

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
