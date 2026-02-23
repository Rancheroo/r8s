## Problem

The `ExpectedFiles()` function in `internal/bundle/health.go` uses incorrect hardcoded paths that don't match the actual bundle structure collected by `rancher2_logs_collector.sh`.

## Current Behavior

Health check reports these files as missing, but they actually exist at different paths:

| Health Check Looks For | Actually At | Status |
|------------------------|-------------|--------|
| `rke2/etcd/endpoint_status` | `etcd/endpointstatus` | Wrong path |
| `rke2/logs/journald.log` | `journald/rke2-server` | Wrong path |
| `rke2/dmesg` | `systemlogs/*` or `systeminfo/dmesg` | Wrong location |
| `rke2/sysstat/` | Doesn't exist in bundles | Wrong location |

## Impact

- Bundle health shows ~69% complete when it should be higher
- False "missing file" warnings confuse users
- Incorrect health score affects trust in analysis output

## Bundle Structure (from actual RKE2 bundle)

```
bundle/
├── etcd/
│   ├── endpointstatus          # NOT rke2/etcd/endpoint_status
│   ├── endpointhealth
│   └── ...
├── journald/
│   ├── rke2-server             # NOT rke2/logs/journald.log
│   ├── cloud-init
│   └── ...
├── rke2/
│   ├── kubectl/
│   ├── podlogs/
│   └── ...
├── systemlogs/
│   ├── kern.log
│   └── syslog
└── systeminfo/
    └── ...
```

## Inconsistency Across Codebase

| Component | Path Used |
|-----------|-----------|
| `health.go` | `rke2/etcd/endpoint_status` (wrong) |
| `health.go` (fixed locally) | `etcd/endpointstatus` (correct) |
| `dmesg.go` parser | `systeminfo/dmesg` |
| BUNDLE-FORMAT.md | Documents `systemlogs/` but not `sysstat/` |

## Proposed Fix

1. **Audit all parsers** for hardcoded paths
2. **Align with collector script** - verify paths from `rancher2_logs_collector.sh`
3. **Centralize path definitions** or document canonical locations
4. **Test with real bundles** from different RKE2 versions

## Acceptance Criteria

- [ ] All paths in `ExpectedFiles()` match actual bundle structure
- [ ] Parsers and health check use consistent paths
- [ ] BUNDLE-FORMAT.md documents paths accurately
- [ ] Tested with bundles from RKE2 v1.28+ and v1.29+

## Additional Context

Reported during Sprint 8 validation with real customer bundle:
```
Bundle Type: RKE2
Health: ● (69% complete)
Issues Found: WARNING missing_file: rke2/etcd/endpoint_status
```

Files verified to exist at correct paths, but health check looks in wrong places.

---

**Labels:** bug, sprint-8, bundle-health
**Priority:** High (affects user trust in health metrics)
