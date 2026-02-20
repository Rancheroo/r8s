## 🎉 v0.8.0-alpha — kubectl for Rancher Bundles

**The CLI-first pivot is complete.**

After deleting 9,940 lines of TUI code, r8s is now a pure CLI tool that brings kubectl compatibility to Rancher support bundles.

---

### ✨ What's New

#### kubectl-Compatible Commands

```bash
# Resource listing
r8s get pods ./bundle/
r8s get pods ./bundle/ -n kube-system
r8s get nodes ./bundle/
r8s get namespaces ./bundle/
r8s get deployments ./bundle/
r8s get services ./bundle/
r8s get events ./bundle/

# View logs (RKE2 flat structure fully supported)
r8s logs ./bundle/ nginx-pod
r8s logs ./bundle/ nginx-pod -c container-name
r8s logs ./bundle/ nginx-pod --previous
r8s logs ./bundle/ nginx-pod --tail=100

# Describe resources
r8s describe pod ./bundle/ nginx-pod
r8s describe node ./bundle/ my-node
r8s describe deployment ./bundle/ my-deploy

# Bundle analysis
r8s validate ./bundle/
r8s analyze ./bundle/ --format=json | jq '.critical'
r8s test-cluster ./bundle/

# AI integration
r8s generate prompt ./bundle/
r8s generate prompt ./bundle/ --format=terminal | claude
```

---

### 🔧 Critical Fixes

#### Bundle Path Detection (100% Accuracy)

| File | Old Path | New Path |
|------|----------|----------|
| etcd status | `rke2/etcd/endpoint_status` | `etcd/endpointstatus` |
| journald logs | `rke2/logs/journald.log` | `journald/` (directory) |
| dmesg | `rke2/dmesg` | `systeminfo/dmesg` |
| sysstat | `rke2/sysstat/` | `systemlogs/sysstat-data/` |

**Impact:** Bundle validation now detects all 13 critical files with 100% accuracy.

#### Build System Protection
- `make sync` — Sync vendor/check embedded scripts before build
- `make check-sync` — Verify local files match embedded content
- Prevents building with outdated or stale code

---

### 📊 Stats

- **Lines Deleted:** 9,940 (TUI code)
- **Binary Size:** 7.4MB (30% smaller)
- **Test Coverage:** 61.4%
- **Bundle Validation:** 100% (13/13 files)
- **Manual Tests:** 56 tests, 52 passed

---

### 🚀 Quickstart

```bash
# 1. Download
curl -L https://github.com/Rancheroo/r8s/releases/download/v0.8.0-alpha/r8s-v0.8.0-alpha-linux-amd64 -o r8s
chmod +x r8s
sudo mv r8s /usr/local/bin/

# 2. Validate bundle
r8s validate ./extracted-bundle/

# 3. Get pods
r8s get pods ./bundle/

# 4. Check logs for a crashing pod
r8s logs ./bundle/ my-crashing-pod --previous

# 5. Generate AI troubleshooting prompt
r8s generate prompt ./bundle/ --format=terminal | claude
```

---

### 📝 Documentation

- [ROADMAP_v1.0_FINAL.md](docs/development/ROADMAP_v1.0_FINAL.md) — v1.0 scope locked
- [QUICK_WINS_RESEARCH.md](docs/development/QUICK_WINS_RESEARCH.md) — Feature research
- [BUNDLE-FORMAT.md](docs/development/BUNDLE-FORMAT.md) — Bundle structure reference

---

### 🗺️ Roadmap

- **v0.8.1** (Mar 9) — K3s support, `r8s top`, label selectors
- **v0.9.0** (Mar 23) — AI Intelligence (10+ patterns, root cause hints)
- **v1.0.0** (Apr 20) — Stable release

---

### 🙏 Acknowledgments

This release represents a significant milestone:
- TUI deleted (9,940 lines)
- Pure CLI implementation complete
- kubectl compatibility achieved
- Bundle validation locked at 100%

**We live and die by the bundle. This release honors that.**

---

**Full Changelog:** [v0.7.0...v0.8.0-alpha](https://github.com/Rancheroo/r8s/compare/v0.7.0...v0.8.0-alpha)
