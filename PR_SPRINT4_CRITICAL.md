## Sprint 4: Critical Fixes (S4-CRITICAL-1)

Addresses all 6 high-priority CodeRabbit review items from Sprint 3 PR #25.

### Fixes Implemented

| # | File | Issue | Fix |
|---|------|-------|-----|
| 1 | `internal/bundle/oom.go` | QoS enrichment missed matches | Added `normalizePodName()` helper |
| 2 | `internal/bundle/oom.go` | Variable shadowing | Renamed variables, added request=limit defaulting |
| 3 | `internal/tui/handlers.go` | Silent GetContainers error | Now logs errors properly |
| 4 | `internal/tui/handlers.go` | State leakage | Added `clearPodState()` call |
| 5 | `internal/tui/helpers.go` | Pod-level restarts | Documented, distribute proportionally |
| 6 | `internal/tui/helpers.go` | Ignored GetPodResources errors | Check error + length |

### Bonus: Async CRD Loading

- Implemented `processPendingCRDCounts()` with batching (5 per tick)
- Implemented `fetchCRDCountAsync()` for non-blocking fetches
- Increased tick interval: 100ms → 250ms

### Pre-Checks
- [x] File count: 4 files (under 150 limit)
- [x] No example-log-bundle files
- [x] All tests pass
- [x] Build successful

### Documentation
- `SPRINT4_CODERABBIT_PLAN.md` - Full roadmap for all 17 items
- `WORKFLOW_IMPROVEMENTS.md` - Process improvements and lessons

Refs: #25 (Sprint 3), Sprint 4 planning
