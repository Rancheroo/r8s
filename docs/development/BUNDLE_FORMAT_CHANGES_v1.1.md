# Bundle Format Changes Audit

**Date:** 2026-03-03  
**Analyzed Bundles:**
- **New Format:** `r8s-cp-wlp7h-lhvgq-2026-03-03_03_49_04/` (March 3, 2026)
- **Old Format:** `r8s-cp-wlp7h-lhvgq-2026-02-03_05_38_57/` (February 3, 2026)
- **Location:** `/home/bradmin/Downloads/logBundles/`

---

## Summary

The collector script has been updated with significant structural changes. This document captures all differences to inform v1.1 parser development.

---

## NEW Files Added

### Critical New Files (Require New Parsers)

| File | Description | Priority |
|------|-------------|----------|
| `systeminfo/systemd-detect-virt` | Virtualization detection (e.g., "kvm") | **HIGH** |
| `versions` | Comprehensive component versions, helm charts, images | **HIGH** |
| `systeminfo/lscpu` | CPU architecture information | Medium |
| `systeminfo/memory` | Memory info (replaces `freem`) | Medium |

### New Directory Structure

| Directory | Description |
|-----------|-------------|
| `rke2/kubectl/poddescribe/` | Per-namespace pod describe outputs |
| `rke2/kubectl/poddescribe/<namespace>/` | Individual namespace subdirectories |
| `systemlogs/sysstat-data/` | Performance data (sa##, sar## files) |

### New Pod Log Files
- `rke2/agent-logs/kubelet-*.log.gz` (timestamped kubelet logs)
- `journald/rke2-agent` (agent journald logs)
- Updated `cattle-cluster-agent` pod logs (new pod names)

### New System Logs
- `systemlogs/kern.log.1`
- `systemlogs/syslog.1`
- `systemlogs/sysstat-data/sa01`, `sa02`, `sa24-28`
- `systemlogs/sysstat-data/sar01`, `sar02`, `sar24-28`

### New Kubectl Outputs
- `rke2/kubectl/apps`

---

## Files Removed

| File | Replacement | Action Required |
|------|-------------|-----------------|
| `systeminfo/freem` | `systeminfo/memory` | Update path mappings |
| `systemlogs/cloud-init.log` | None | Remove from parser |
| `systemlogs/cloud-init-output.log` | None | Remove from parser |
| `systemlogs/dmesg` | `systeminfo/dmesg` | Update path mappings |

---

## Files Moved

| Old Location | New Location | Notes |
|--------------|--------------|-------|
| `systemlogs/dmesg` | `systeminfo/dmesg` | Now in systeminfo directory |

---

## Structural Changes

### Pod Describe Directory
The new format includes per-namespace pod describe files:
```
rke2/kubectl/poddescribe/
├── calico-system/
├── capi-system/
├── cattle-alerting/
├── cattle-fleet-local-system/
├── cattle-fleet-system/
├── cattle-global-data/
├── cattle-logging/
├── cattle-logging-system/
├── cattle-monitoring-system/
├── cattle-pipeline/
├── cattle-prometheus/
├── cattle-provisioning-capi-system/
├── cattle-resources-system/
├── cattle-system/
├── fleet-default/
├── fleet-system/
├── ingress-nginx/
├── istio-system/
├── kube-public/
├── kube-system/
├── longhorn-system/
├── rancher-operator-system/
├── rancher-turtles-system/
├── rke2-bootstrap-system/
├── rke2-control-plane-system/
├── suse-observability/
└── tigera-operator/
```

### Versions File Content
The new `versions` file at bundle root contains:
- System info (date, memory, kernel, OS)
- RKE2 version
- Container images (rke2, cattle, system-upgrade-controller)
- Helm releases (rke2-* charts)
- Fleet agent info

---

## Cilium Status

**Finding:** No Cilium-specific files detected in either bundle format.

The cluster is using Calico (`networking/cni/10-calico.conflist`). The collector may be preparing for Cilium support, but this cluster is not using it.

---

## Collector Changes Detected

From `collector-output.log` comparison:
- Same collection phases
- Timing differs slightly (new collector takes ~1 minute longer)
- Both detect: Ubuntu 24.04, RKE2, systemd init

---

## Parser Impact Assessment

### High Priority (Breaking Changes)
1. **New `systemd-detect-virt` file** - Add parser for virtualization detection
2. **New `versions` file** - Add comprehensive version parser
3. **`freem` → `memory` migration** - Update path mapping
4. **`dmesg` location change** - Update path mapping

### Medium Priority (New Features)
1. **Pod describe directory** - Add namespace-aware pod describe parser
2. **Sysstat data** - Add performance metrics parsing (sa/sar files)
3. **Kubelet logs** - Add timestamped kubelet log parsing

### Low Priority (Cleanup)
1. Remove cloud-init log parsers
2. Update file existence checks to handle both old/new formats

---

## Implementation Notes

### Backward Compatibility
- Parsers should check for both old and new file locations
- Use fallback logic: try new path first, fall back to old path
- Maintain support for both bundle formats during transition

### New Parsers Needed
1. `VirtDetector` - Parse `systemd-detect-virt` output
2. `VersionsParser` - Parse root-level `versions` file
3. `PodDescribeParser` - Handle per-namespace pod describe directories
4. `SysstatParser` - Parse sar/sa performance data

---

## Related Files

- Collector output logs analyzed:
  - `/home/bradmin/Downloads/logBundles/r8s-cp-wlp7h-lhvgq-2026-03-03_03_49_04/collector-output.log`
  - `/home/bradmin/Downloads/logBundles/r8s-cp-wlp7h-lhvgq-2026-02-03_05_38_57/collector-output.log`
