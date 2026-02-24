# Sprint 12: v1.0 Final Push — Pattern Completion & kubectl Plugin

**Status:** Planning  
**Target Date:** March 9, 2026 (2 weeks)  
**Theme:** Content Completion & Integration  
**Base Branch:** `main` (post-Sprint 11 merge)  

---

## 🎯 Sprint Goal

**Ship r8s v1.0: The complete kubectl-for-bundles experience.**

Sprints 10-11 delivered the engine. Sprint 12 delivers the content and integration that makes it production-ready.

**What's Done (Sprint 10-11):**
- ✅ AI Pattern Engine v2 with confidence scoring
- ✅ Root Cause Hint system
- ✅ Natural Language Queries (`r8s ask`)
- ✅ Export formats (SARIF, JUnit, Markdown, JSON)
- ✅ kubectl-style commands (`r8s get`, `r8s logs`, `r8s describe`)
- ✅ Pure CLI foundation

**What's Missing (Sprint 12):**
- 🔴 Pattern library content (only 3/10+ patterns exist)
- 🔴 kubectl-r8s plugin wrapper
- 🔴 Real-world bundle validation
- 🟡 Documentation polish

---

## 📋 Scope

### P0: Pattern Library Completion (Days 1-6)
**Impact: CRITICAL | Effort: MEDIUM | User Value: 🔥🔥🔥**

**THE GAP:** We have a Ferrari engine with no fuel. The AI pattern engine is ready, but only 3 patterns exist.

**Current Patterns (3/10+):**
- ✅ `oomkill.yaml` — Out of memory kills
- ✅ `imagepullbackoff.yaml` — Image pull failures
- ✅ `crashloopbackoff.yaml` — Crash loop detection

**Required New Patterns (7+):**

```yaml
# 1. etcd.yaml
name: "etcd issues"
description: "Detects etcd corruption, latency, space exceeded"
severity: critical
sources:
  - "*/etcd.log"
  - "*/etcd/server.log"
patterns:
  - regex: "etcdserver: mvcc: database space exceeded"
    weight: 1.0
  - regex: "etcdserver: read-only range request .* took too long"
    weight: 0.9
  - regex: "etcdserver: corrupted"
    weight: 1.0
root_cause_hint: "etcd database space exceeded. Defragment with: etcdctl defrag"

# 2. certificates.yaml
name: "Certificate Expiration"
description: "Detects expired or expiring certificates"
severity: critical
sources:
  - "*/secrets.yaml"
  - "*/nodes/*.yaml"
patterns:
  - regex: "certificate has expired"
    weight: 1.0
  - regex: "x509: certificate has expired"
    weight: 1.0
  - regex: "NotAfter.*\\d{4}-\\d{2}-\\d{2}"
    weight: 0.5  # Check date logic in code
root_cause_hint: "Certificate expired {{.DaysAgo}} days ago. Renew with: kubectl certificate approve {{.CSRName}}"

# 3. networking.yaml
name: "Network Issues"
description: "DNS failures, CNI errors, connectivity issues"
severity: warning
sources:
  - "*/pods/*.log"
  - "*/events.yaml"
patterns:
  - regex: "cni plugin failed to set up sandbox"
    weight: 1.0
  - regex: "failed to resolve DNS"
    weight: 0.9
  - regex: "no such host"
    weight: 0.8
  - regex: "connection refused"
    weight: 0.7
root_cause_hint: "CNI plugin failed. Check CoreDNS pods: kubectl get pods -n kube-system -l k8s-app=kube-dns"

# 4. storage.yaml
name: "Storage Issues"
description: "PVC binding failures, storage pressure"
severity: warning
sources:
  - "*/events.yaml"
  - "*/pvc/*.yaml"
patterns:
  - regex: "Failed to provision volume"
    weight: 1.0
  - regex: "persistentvolumeclaim .* not found"
    weight: 0.9
  - regex: "Volume is already exclusively attached"
    weight: 0.8
root_cause_hint: "PVC binding failed. Check storage class: kubectl get storageclass"

# 5. node-pressure.yaml
name: "Node Pressure"
description: "Disk, memory, PID pressure conditions"
severity: warning
sources:
  - "*/nodes/*.yaml"
  - "*/events.yaml"
patterns:
  - regex: "\\bDiskPressure\\b"
    weight: 1.0
  - regex: "\\bMemoryPressure\\b"
    weight: 1.0
  - regex: "\\bPIDPressure\\b"
    weight: 1.0
  - regex: "\\bOutOfDisk\\b"
    weight: 1.0
root_cause_hint: "Node {{.NodeName}} under pressure ({{.PressureType}}). Clean up: docker system prune -a"

# 6. pending-pods.yaml
name: "Pod Scheduling Failures"
description: "Pods stuck in Pending, Terminating, Unknown"
severity: warning
sources:
  - "*/pods/*.yaml"
  - "*/events.yaml"
patterns:
  - regex: "\\bPending\\b.*\\bd+ (d+h)?\\b"  # Pending for time
    weight: 0.8
  - regex: "Insufficient (cpu|memory)"
    weight: 0.9
  - regex: "didn't match pod anti-affinity rules"
    weight: 0.8
  - regex: "\\bTerminating\\b.*\\bd+ (d+h)?\\b"
    weight: 0.7
root_cause_hint: "Pod stuck pending. Check resources: kubectl describe pod {{.PodName}} -n {{.Namespace}}"

# 7. leader-election.yaml
name: "Leader Election Issues"
description: "Control plane leader failures"
severity: critical
sources:
  - "*/kube-system/*.log"
  - "*/kubernetes/*.log"
  - "*/etcd.log"
patterns:
  - regex: "leader election lost"
    weight: 1.0
  - regex: "failed to renew lease"
    weight: 0.9
  - regex: "leader changed"
    weight: 0.7
root_cause_hint: "Control plane leader instability. Check etcd quorum: kubectl get endpoints kube-controller-manager -n kube-system"

# BONUS Patterns (if time permits):
# 8. api-server-timeout.yaml
# 9. resource-quota.yaml
# 10. security-context.yaml
```

**Success Criteria:**
- All 10 patterns detect real issues in test bundles
- >80% precision on known-bad bundles
- Patterns load dynamically from YAML (no rebuild needed)

---

### P0: kubectl-r8s Plugin (Days 4-8)
**Impact: CRITICAL | Effort: MEDIUM | User Value: 🔥🔥🔥**

**THE GAP:** We have `r8s get pods`, but users want `kubectl r8s get pods`.

**Implementation Options:**

**Option A: kubectl Plugin (Recommended)**
Create a kubectl plugin that translates kubectl commands to r8s.

```go
// cmd/kubectl-r8s/main.go
// kubectl plugin entry point

package main

import (
  "fmt"
  "os"
  "os/exec"
  "strings"
)

func main() {
  // kubectl passes: [kubectl-r8s, get, pods, -n, namespace]
  // We need to translate to: r8s get pods <bundle-path> -n namespace
  
  args := os.Args[1:] // Skip program name
  
  // Find bundle path (should be in env or default)
  bundlePath := os.Getenv("R8S_BUNDLE")
  if bundlePath == "" {
    // Try to infer from context or use default
    bundlePath = findBundle()
  }
  
  // Reconstruct r8s command
  r8sArgs := []string{"r8s"}
  r8sArgs = append(r8sArgs, args...)
  r8sArgs = append(r8sArgs, bundlePath)
  
  // Execute r8s
  cmd := exec.Command("r8s", r8sArgs[1:]...)
  cmd.Stdout = os.Stdout
  cmd.Stderr = os.Stderr
  cmd.Run()
}

func findBundle() string {
  // Look for bundle in current dir
  // Return first .tar.gz that looks like a bundle
}
```

**Installation:**
```bash
# Install plugin
cp kubectl-r8s ~/.local/bin/
kubectl plugin list
# Output: kubectl-r8s

# Usage
export R8S_BUNDLE=./support-bundle.tar.gz
kubectl r8s get pods -n cattle-system
kubectl r8s logs nginx-pod -n default
kubectl r8s describe node worker-1
```

**Option B: Wrapper Script (Quick Win)**
```bash
#!/bin/bash
# kubectl-r8s wrapper
# Faster to implement, less flexible

BUNDLE_PATH=${R8S_BUNDLE:-$(find . -name "*.tar.gz" -o -name "support-bundle*" | head -1)}

# Remove 'r8s' from args if present
ARGS=($@)
if [ "${ARGS[0]}" = "r8s" ]; then
  ARGS=(${ARGS[@]:1})
fi

r8s "${ARGS[@]}" "$BUNDLE_PATH"
```

**Success Criteria:**
- User can type `kubectl r8s get pods` and it works
- Plugin available via `kubectl krew install r8s` (stretch goal)
- README documents installation

---

### P1: Real-World Validation (Days 6-10)
**Impact: HIGH | Effort: MEDIUM | User Value: 🔥🔥🔥**

**THE GAP:** New AI code needs testing on real bundles before v1.0.

**Test Matrix:**

| Bundle Type | Source | Expected Result |
|-------------|--------|-----------------|
| RKE2 v1.28 | Internal test | All patterns detect correctly |
| RKE2 v1.29 | Internal test | All patterns detect correctly |
| K3s v1.28 | Customer sample | K3s format detection works |
| K3s v1.29 | Customer sample | K3s format detection works |
| Rancher provisioned | Internal | Rancher-specific patterns |
| Known-bad bundle | Support case | Patterns detect actual issues |

**Validation Process:**
```bash
# For each bundle:
r8s analyze ./bundle --format=json
# Verify: Correct distro detected
# Verify: Patterns find real issues (no false positives)
# Verify: Root cause hints are accurate
# Verify: Performance <2s
```

**Success Criteria:**
- 10+ real bundles tested
- <5% false positive rate
- <5% false negative rate on known issues
- Performance <2s on all bundles

---

### P1: Documentation & Polish (Days 8-12)
**Impact: MEDIUM | Effort: LOW | User Value: 🔥🔥**

**README Updates:**
- [ ] Quickstart guide (30-second demo)
- [ ] Installation instructions (kubectl plugin)
- [ ] Pattern library documentation
- [ ] `r8s ask` examples
- [ ] Export format examples (CI/CD integration)

**Website/Docs:**
- [ ] Landing page copy
- [ ] Feature highlights
- [ ] Comparison table (r8s vs kubectl vs dashboard)

**CLI Help:**
- [ ] All commands have `--help` with examples
- [ ] Man pages generated

---

### P2: Performance Hardening (Days 10-13)
**Impact: MEDIUM | Effort: LOW | User Value: 🔥🔥**

**Benchmarking:**
```bash
# Performance targets
r8s analyze ./100mb-bundle    # <2s
r8s get pods ./bundle         # <500ms
r8s logs ./bundle pod         # <1s
```

**Optimizations (if needed):**
- [ ] Parallel pattern matching (already in engine)
- [ ] Bundle index caching
- [ ] Lazy loading for large files

**Success Criteria:**
- All commands meet performance targets
- Memory usage <500MB for 100MB bundle
- No goroutine leaks

---

## 📅 Day-by-Day Plan

### Week 1: Content & Integration

| Day | Focus | Deliverables |
|-----|-------|--------------|
| **Day 1** | Pattern Library (etcd, certs) | 2 new patterns with tests |
| **Day 2** | Pattern Library (network, storage) | 2 new patterns with tests |
| **Day 3** | Pattern Library (node, pending, leader) | 3 new patterns with tests |
| **Day 4** | kubectl Plugin Design | Design doc, choose Option A or B |
| **Day 5** | kubectl Plugin Implementation | Working plugin |
| **Day 6** | Pattern Testing | All 10 patterns validated |
| **Day 7** | Buffer / Integration Testing | Plugin + patterns working together |

### Week 2: Validation & Release

| Day | Focus | Deliverables |
|-----|-------|--------------|
| **Day 8** | Real-World Validation | 5 real bundles tested |
| **Day 9** | Real-World Validation | 5 more bundles tested |
| **Day 10** | Documentation | README, help, man pages |
| **Day 11** | Performance Tuning | Benchmarks, optimizations |
| **Day 12** | Final Testing | Full integration test |
| **Day 13** | Bug Fixes | Address any issues |
| **Day 14** | **RELEASE v1.0** | Tag, release notes, announcement |

---

## ✅ Definition of Done (v1.0)

**Code:**
- [ ] 10+ patterns in `internal/ai/patterns/`
- [ ] kubectl-r8s plugin working
- [ ] All tests passing (60%+ coverage)
- [ ] Performance targets met

**Validation:**
- [ ] 10+ real bundles tested
- [ ] <5% false positive rate
- [ ] <5% false negative rate

**Documentation:**
- [ ] README updated
- [ ] All commands have `--help`
- [ ] kubectl plugin installation docs

**Release:**
- [ ] CHANGELOG.md updated
- [ ] Git tag `v1.0.0`
- [ ] GitHub Release with binaries
- [ ] Release announcement

---

## 🎯 Success Metrics

| Metric | Target | Current | Gap |
|--------|--------|---------|-----|
| Patterns | 10+ | 3 | **+7** |
| kubectl plugin | Working | Missing | **New** |
| Test bundles | 10+ | Unknown | **Need** |
| False positive rate | <5% | Unknown | **Measure** |
| Performance | <2s | Unknown | **Validate** |

---

## 📊 What's Actually Left

**The REAL v1.0 Gap:**
1. **Content:** Write 7 YAML pattern files
2. **Integration:** Build kubectl plugin wrapper
3. **Validation:** Test on 10+ real bundles
4. **Docs:** Update README for new features

**Time Estimate:** 2 weeks (this sprint)

**Risk Level:** LOW
- Patterns are YAML files (low risk)
- Plugin is wrapper code (medium risk)
- Validation is testing (low risk)

---

## 🚀 After v1.0 (Post-Sprint 12)

**v1.1 Ideas:**
- More patterns (aim for 20+)
- kubectl krew distribution
- VS Code extension
- Rancher UI integration

**But First:** Ship v1.0. Perfect is the enemy of done.

---

## 📝 For the Sprint Team

**This is it.** The final push. Everything else is done.

**Focus:**
1. Write the pattern YAML files (Days 1-3)
2. Build the kubectl plugin (Days 4-5)
3. Test on real bundles (Days 8-9)
4. Ship v1.0 (Day 14)

**For Elon. For the kubectl compatibility. For v1.0.** 🚀