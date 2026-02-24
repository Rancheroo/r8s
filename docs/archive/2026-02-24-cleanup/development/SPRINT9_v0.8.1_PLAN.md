# Sprint 9: v0.8.1 — K3s Support + kubectl Polish
**K3s Format Detection & kubectl Quick Wins**

**Duration:** 1 week  
**Target Date:** March 9, 2026  
**Base Branch:** `main` (post-v0.8.0-alpha)  
**Theme:** Multi-Distro Foundation

---

## 🎯 Sprint Goal

Add K3s support to r8s and implement the high-value kubectl quick wins identified in research. This sprint extends bundle compatibility beyond RKE2 and polishes the CLI experience.

---

## 📋 Scope

### P0: K3s Support (Days 1-3)
**Impact: HIGH | Effort: MEDIUM**

**Deliverables:**
- [ ] K3s format detection in `DetectFormat()`
- [ ] Path abstraction for 5 core files:
  - `types.go` — Add `FormatK3s`, path helpers
  - `manifest.go` — Use `b.KubectlPath()` with K3s paths
  - `validate.go` — Dynamic distro validation for K3s
  - `journald.go` — Service name mapping (k3s-agent vs rke2-server)
  - `completeness.go` — Dynamic path generation for K3s
- [ ] K3s sample bundle acquisition and testing
- [ ] Documentation update for K3s users

**K3s Path Differences:**
| Component | RKE2 Path | K3s Path |
|-----------|-----------|----------|
| Service logs | `journald/rke2-server.log` | `journald/k3s-agent.log` |
| Pod manifests | `rke2/pod-manifests/` | `k3s/pod-manifests/` |
| Data dir | `/var/lib/rancher/rke2/` | `/var/lib/rancher/k3s/` |

**Success Criteria:**
- K3s bundles analyze correctly
- No regression on RKE2 bundles
- 55%+ test coverage

---

### P1: `r8s top` — Resource Usage (Days 3-4)
**Impact: HIGH | Effort: MEDIUM**

**Deliverables:**
- [ ] `r8s top pods ./bundle/` — CPU/memory per pod
- [ ] `r8s top pods ./bundle/ --containers` — Per-container metrics
- [ ] `r8s top nodes ./bundle/` — Node-level resources

**Data Sources:**
- `kubectl top` output in bundle
- Container resource specs in pod manifests
- Node capacity/allocatable from node info

**Output Format:**
```
NAMESPACE    NAME            CPU(cores)   MEMORY(bytes)
kube-system  coredns-7c4...  5m           15Mi
default      nginx-pod       10m          64Mi
```

---

### P2: Label Selector (`-l` flag) (Day 4)
**Impact: HIGH | Effort: LOW**

**Deliverables:**
- [ ] `r8s get pods ./bundle/ -l app=nginx`
- [ ] `r8s get pods ./bundle/ -l 'tier in (frontend, backend)'`
- [ ] Support common selector operators: `=`, `!=`, `in`, `notin`

**Implementation:** Filter parsed resources by labels before output.

---

### P3: Export Formats Expansion (Day 5)
**Impact: MEDIUM | Effort: LOW**

**Deliverables:**
- [ ] `r8s generate export ./bundle/ --format=sarif` — Security scanners
- [ ] `r8s generate export ./bundle/ --format=junit` — CI/CD integration
- [ ] `r8s generate export ./bundle/ --format=markdown` — Human reports

---

## 📅 Day-by-Day Plan

### Day 1: K3s Foundation
- [ ] Acquire K3s sample bundle
- [ ] Document K3s path structure differences
- [ ] Add `FormatK3s` to types.go
- [ ] Update `DetectFormat()` for K3s detection

### Day 2: K3s Path Abstraction
- [ ] Implement path resolver for K3s
- [ ] Update `journald.go` for k3s-agent service names
- [ ] Update `completeness.go` for K3s expected files
- [ ] Test with K3s sample bundle

### Day 3: K3s Testing & `r8s top`
- [ ] K3s smoke tests
- [ ] RKE2 regression tests
- [ ] Start `r8s top` implementation
- [ ] Parse resource specs from pod manifests

### Day 4: `r8s top` & Label Selectors
- [ ] Complete `r8s top pods/nodes`
- [ ] Implement `-l` label selector flag
- [ ] Add selector parser

### Day 5: Export Formats & Polish
- [ ] SARIF export format
- [ ] JUnit export format
- [ ] Markdown export format
- [ ] Documentation updates
- [ ] Tag v0.8.1

---

## ✅ Success Criteria

| Metric | Target |
|--------|--------|
| K3s Support | 100% bundle detection |
| `r8s top` | Working for pods and nodes |
| Label Selector | `-l` flag functional |
| Export Formats | 3 new formats (SARIF, JUnit, Markdown) |
| Test Coverage | 55%+ |
| RKE2 Regression | 0 failures |

---

## 🛠️ Implementation Notes

### K3s Format Detection
```go
// In DetectFormat()
k3sDir := filepath.Join(extractPath, "k3s")
if stat, err := os.Stat(k3sDir); err == nil && stat.IsDir() {
    return FormatK3s
}
```

### Path Resolver Interface
```go
type PathResolver interface {
    GetPodManifestsDir() string      // rke2/pod-manifests/ vs k3s/pod-manifests/
    GetJournaldServiceName() string  // rke2-server vs k3s-agent
    GetDataDir() string              // /var/lib/rancher/rke2 vs /var/lib/rancher/k3s
}
```

### Label Selector Parsing
```go
// Support: app=nginx, tier!=frontend, env in (prod, staging)
func ParseLabelSelector(selector string) (LabelFilter, error)
```

---

## 📚 Resources

### K3s Documentation
- K3s support bundle structure: https://docs.k3s.io/
- Sample bundles needed for testing

### Quick Wins Reference
- `docs/development/QUICK_WINS_RESEARCH.md` — Full feature analysis

---

## 🚫 Out of Scope

**Deferred to v0.9.0:**
- `r8s exec` — Command replay simulation
- `r8s diff` — Bundle comparison
- `r8s wait` — Condition checking
- Complex label selectors (set-based)

---

## 🎯 Definition of Done

- [ ] K3s bundles analyze correctly
- [ ] `r8s top pods/nodes` working
- [ ] `-l` label selector functional
- [ ] 3 new export formats working
- [ ] All tests passing
- [ ] Coverage ≥ 55%
- [ ] Documentation updated
- [ ] v0.8.1 tagged and released

---

*"Multi-distro support begins with K3s. The foundation for future distros."*
