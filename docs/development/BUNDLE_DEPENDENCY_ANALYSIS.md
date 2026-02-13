# R8S Bundle Dependency Analysis
## rancherlabs/support-tools: rancher2_logs_collector.sh

**Analysis Date:** 2026-02-12  
**Critical Dependency:** HIGH - R8s parsing depends entirely on this script's output format  
**Repository:** https://github.com/rancherlabs/support-tools  
**Path:** `collection/rancher/v2.x/logs-collector/rancher2_logs_collector.sh`

---

## Executive Summary

This script is the **single source of truth** for RKE2/Rancher support bundle creation. Any change to:
- Directory structure
- File naming conventions  
- Command output formats
- New data sources added

...will **break R8s parsing** or create opportunities for enhanced data extraction.

**Critical Finding:** PR #418 (merged 2026-02-10) added `kubectl describe pod` output - a major new data source R8s doesn't yet parse.

---

## Script Architecture

### Entry Points & Distribution Detection

```bash
# The script auto-detects OR accepts manual distro flag:
DISTRO_FLAG="${1}"  # Optional: rke2, k3s, rke

# Detection logic (sherlock function):
if command -v rke2 >/dev/null 2>&1
  -> rke2-setup()
elif systemctl is-active --quiet k3s  
  -> k3s-setup()
elif [ -d /opt/rke ]
  -> rke-setup()
```

**R8s Impact:** Bundle structure varies slightly by distro (path prefixes change)

### Core Collection Functions

| Function | Purpose | R8s Relevance |
|----------|---------|---------------|
| `rke2-k8s()` | Collect RKE2-specific K8s data | **PRIMARY** - Most customer bundles |
| `k3s-k8s()` | Collect K3s data | Secondary |
| `rke-k8s()` | Collect RKE1 data | Legacy |
| `system-info()` | OS-level diagnostics | Used for node pressure analysis |

---

## Directory Structure (The Contract)

### Standard RKE2 Bundle Layout

```
${TMPDIR}/${LOGNAME}/
├── rke2/
│   ├── kubectl/
│   │   ├── pods              # kubectl get pods -o wide
│   │   ├── events            # kubectl get events
│   │   ├── nodes             # kubectl get nodes
│   │   ├── deployments       # kubectl get deployments
│   │   ├── [all K8S_OBJECTS_NAMESPACED]
│   │   └── poddescribe/      # ⚠️ NEW in PR #418
│   │       └── poddescribe-${NAMESPACE}
│   ├── podlogs/              # Container log files
│   │   └── ${NAMESPACE}-${POD_NAME}  # Plain text, no container suffix!
│   ├── pod-manifests/        # Static pod YAMLs (if present)
│   ├── crictl/
│   ├── agent-logs/
│   └── [cert dirs]
├── systeminfo/
│   ├── cpuinfo, meminfo, dfh
│   ├── dmesg                 # Kernel messages (OOM clues)
│   └── ps, top, lsof
├── systemlogs/
│   └── syslog, messages
└── metadata.json
```

### Critical File Formats for R8s

#### 1. kubectl/pods (Space-Delimited Table)
```
NAMESPACE     NAME                     READY  STATUS   RESTARTS  AGE  IP
kube-system   etcd-masternode-1        1/1    Running  0         3d   10.0.0.1
```
**Parser:** `internal/bundle/kubectl.go:ParsePods()`
**Risk:** Column alignment changes break parsing

#### 2. kubectl/events (Space-Delimited with Messages)
```
LAST SEEN   TYPE     REASON      OBJECT           MESSAGE
3m          Warning  OOMKilling  pod/oom-test     Memory cgroup out of memory
```
**Parser:** `internal/bundle/oom.go:parseOOMEvents()`
**Critical for:** OOMKill detection

#### 3. podlogs/ (Filename Convention)
```
${NAMESPACE}-${POD_NAME}
# NOT: ${NAMESPACE}-${POD_NAME}-${CONTAINER}
```
**Key Finding:** Container names are NOT in log filenames!
**R8s Solution:** Must parse pod-manifests/ or poddescribe/ for container names

#### 4. NEW: poddescribe/ (Full "kubectl describe" Output)
```
Name:         my-pod
Namespace:    default
Status:       Running
Containers:
  nginx:
    Image: nginx:1.19
    Limits: {memory: 512Mi}
    Requests: {memory: 256Mi}
  istio-proxy:
    Image: istio/proxyv2:1.12
Events:
  Type    Reason     Age   From     Message
  ----    ------     ----  ----     -------
  Normal  Scheduled  3m    default-scheduler
```
**Parser:** NOT YET IMPLEMENTED in R8s
**Gold Mine:** Contains Events, resource requests/limits, container states

---

## Namespace Coverage

### System Namespaces Array (Hardcoded)
```bash
SYSTEM_NAMESPACES=(
  kube-system
  cattle-system
  cattle-monitoring-system
  longhorn-system
  fleet-system
  calico-system
  rke2-control-plane-system
  # ... 20+ more
)
```

**R8s Risk:** If customer adds custom namespaces, they're NOT collected by default.
**Mitigation:** R8s should handle missing namespaces gracefully.

---

## Command Timeouts & Resource Limits

```bash
TIMEOUT=60                    # Commands timeout after 60s
DEFAULT_LOG_DAYS=7            # Only collect 7 days of logs
SPACE=1536                    # Require 1.5GB free disk
PRIORITY_NICE=19              # Run at lowest CPU priority
PRIORITY_IONICE="idle"        # Run at lowest I/O priority
```

**R8s Impact:** Bundle may be incomplete if commands timeout.
**Design Principle:** R8s must work with **partial bundles** (Principle #11: Empty is Valid).

---

## Change Detection Strategy

### What to Monitor Weekly

1. **New kubectl Commands Added**
   - Check for new `kubectl get` or `kubectl describe` calls
   - Adds new data sources we can leverage

2. **Output Format Changes**
   - `kubectl` table format changes (rare but breaking)
   - JSON vs table output switches

3. **Directory Structure Changes**
   - New subdirectories added
   - Files moved between locations

4. **Namespace Array Updates**
   - New system namespaces added
   - Rancher adds new features

### Monitoring Script (For Cron)

```bash
#!/bin/bash
# Weekly check for rancher2_logs_collector.sh changes

SCRIPT_URL="https://raw.githubusercontent.com/rancherlabs/support-tools/master/collection/rancher/v2.x/logs-collector/rancher2_logs_collector.sh"
LOCAL_COPY="/r8s/refs/logs-collector-reference.sh"

# Download current version
curl -s "$SCRIPT_URL" -o /tmp/logs-collector-latest.sh

# Check if changed
if ! diff -q "$LOCAL_COPY" /tmp/logs-collector-latest.sh >/dev/null 2>&1; then
  echo "ALERT: logs-collector.sh has changed!"
  diff "$LOCAL_COPY" /tmp/logs-collector-latest.sh | head -50
  # Create GitHub issue or alert
fi
```

---

## Integration Points for R8s

### Currently Parsed (Working)

| Data Source | File Path | Parser Status |
|-------------|-----------|---------------|
| Pod list | `rke2/kubectl/pods` | ✅ `ParsePods()` |
| Events | `rke2/kubectl/events` | ✅ `parseOOMEvents()` |
| Pod logs | `rke2/podlogs/*` | ✅ `GetLogs()` |
| Pod manifests | `rke2/pod-manifests/*` | ⚠️ Partial (just added) |
| Node info | `systeminfo/*` | ⚠️ Partial |

### NOT YET Parsed (Opportunity)

| Data Source | File Path | Priority |
|-------------|-----------|----------|
| Pod describe | `rke2/kubectl/poddescribe/*` | 🔴 HIGH (PR #418) |
| Deployments | `rke2/kubectl/deployments` | 🟡 Medium |
| Services | `rke2/kubectl/services` | 🟡 Medium |
| PVC/PV | `rke2/kubectl/pvc`, `pv` | 🟡 Medium |
| dmesg | `systeminfo/dmesg` | 🔴 HIGH (OOM kernel traces) |

---

## Critical Risk Scenarios

### Scenario 1: kubectl Output Format Change
**Risk:** `kubectl get pods` switches from table to JSON  
**Impact:** R8s `ParsePods()` breaks, no pods display  
**Mitigation:** Detect format in first line, switch parser

### Scenario 2: podlogs Filename Change
**Risk:** Script adds container name to filename: `ns-pod-container.log`  
**Impact:** Log correlation breaks, can't fetch logs  
**Mitigation:** Flexible filename parsing with fallback

### Scenario 3: pod-manifests Removed
**Risk:** RKE2 stops writing static pod manifests  
**Impact:** Can't get container names for multi-container pods  
**Mitigation:** Fallback to poddescribe/ (PR #418)

### Scenario 4: PR #418 poddescribe Changes
**Risk:** Output format of `kubectl describe pod` changes  
**Impact:** New parser (when written) breaks  
**Mitigation:** Version detection, multiple parser strategies

---

## Action Items for Dev Team

### Immediate (This Sprint)

1. **Fix Multi-Container Detection**
   - Debug why `parsePodManifestsForContainers()` isn't finding containers
   - Verify path resolution: `rke2/pod-manifests/` vs `rke2/kubectl/poddescribe/`

2. **Implement poddescribe Parser**
   - New function: `ParsePodDescribe(bundleRoot, namespace, podName)`
   - Extract: Events, container states, resource limits

3. **Add dmesg Parser**
   - `systeminfo/dmesg` contains kernel OOM traces
   - Correlates with container OOMs for root cause

### Short-Term (Next 2 Sprints)

4. **Version Detection**
   - Detect bundle format version from collector script
   - Switch parsing strategies based on detected format

5. **Change Monitoring**
   - Implement weekly check script
   - Auto-create issues when changes detected

6. **Graceful Degradation**
   - Test R8s with partial bundles (missing files)
   - Ensure core functionality works with minimal data

### Long-Term (Ongoing)

7. **Upstream Communication**
   - Subscribe to support-tools repo notifications
   - Propose R8s-friendly changes (e.g., structured JSON output)

---

## Appendix: Key Code Sections from Script

### RKE2 Collection Function (Simplified)
```bash
rke2-k8s() {
  # 1. Collect cluster-level objects
  for OBJECT in "${K8S_OBJECTS[@]}"; do
    ${RKE2_BIN} kubectl get "$OBJECT" -o wide > "${TMPDIR}/rke2/kubectl/${OBJECT}"
  done

  # 2. Collect namespaced objects  
  for NAMESPACE in "${SYSTEM_NAMESPACES[@]}"; do
    for OBJECT in "${K8S_OBJECTS_NAMESPACED[@]}"; do
      ${RKE2_BIN} kubectl get "$OBJECT" -n "$NAMESPACE" > "${TMPDIR}/rke2/kubectl/${OBJECT}.${NAMESPACE}"
    done
    
    # ⚠️ NEW in PR #418:
    ${RKE2_BIN} kubectl describe pod -n "$NAMESPACE" > "${TMPDIR}/rke2/kubectl/poddescribe/poddescribe-${NAMESPACE}"
  done

  # 3. Collect pod logs
  for NAMESPACE in "${SYSTEM_NAMESPACES[@]}"; do
    PODS=$(${RKE2_BIN} kubectl get pods -n "$NAMESPACE" -o jsonpath='{.items[*].metadata.name}')
    for POD in $PODS; do
      ${RKE2_BIN} kubectl logs "$POD" -n "$NAMESPACE" > "${TMPDIR}/rke2/podlogs/${NAMESPACE}-${POD}"
    done
  done
}
```

### Design Philosophy Notes

From script comments and structure:
1. **Low priority execution** - Doesn't impact production workloads
2. **Timeout protection** - Won't hang on broken clusters  
3. **Partial collection** - Gets what it can, skips what it can't
4. **Distro abstraction** - Same structure for RKE2/K3s/RKE

These align perfectly with R8s Principles #11 (Empty is Valid) and resilience goals.

---

## Reference Links

- **Script Source:** https://github.com/rancherlabs/support-tools/blob/master/collection/rancher/v2.x/logs-collector/rancher2_logs_collector.sh
- **PR #418 (describe pod):** https://github.com/rancherlabs/support-tools/pull/418
- **Documentation:** https://rancher.com/docs/rancher/v2.x/en/troubleshooting/

---

*Analysis compiled for R8s Development Team*  
*Next review: 2026-02-19*
