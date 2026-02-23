## Problem

The `describe` command has rigid positional argument ordering that doesn't match user expectations. When flags like `-n` are placed between positional args, the parsing breaks.

## Current Behavior

**Command that fails:**
```bash
r8s describe pod -n cattle-system cattle-cluster-agent-dbb4889db-4qhf5 ~/Downloads/bundle/
Error: bundle path not found: stat cattle-cluster-agent-dbb4889db-4qhf5: no such file or directory
```

**What happens:**
- Cobra parses `-n cattle-system` correctly as flags
- Remaining positional args: `["pod", "cattle-cluster-agent...", "~/Downloads/bundle/"]`
- `parseDescribeArgs()` assigns:
  - `kind = args[0]` = "pod" ✓
  - `bundlePath = args[1]` = "cattle-cluster-agent..." ✗
  - `name = args[2]` = "~/Downloads/bundle/" ✗

## Expected Behavior

Either:
1. **Auto-detect bundle path** - Check if arg[1] exists as a path, if not try arg[2]
2. **Support flexible flag placement** - kubectl allows flags anywhere
3. **Better error message** - "Invalid bundle path 'cattle-cluster-agent-dbb4889db-4qhf5' - expected directory"

## Workaround

Use the correct argument order:
```bash
r8s describe pod ~/Downloads/bundle/ cattle-cluster-agent-dbb4889db-4qhf5 -n cattle-system
```

## Impact

- Users familiar with kubectl syntax will hit this
- Error message is confusing (suggests file doesn't exist rather than wrong order)
- Poor UX for a commonly-used command

## Proposed Fix

Update `parseDescribeArgs()` in `cmd/describe.go` to:

```go
func parseDescribeArgs(args []string) (kind, bundlePath, name string) {
    // Try to auto-detect bundle path by checking if it exists
    if len(args) == 2 {
        // Could be: [bundle, name] or [name, bundle]
        if _, err := os.Stat(args[0]); err == nil {
            bundlePath = args[0]
            name = args[1]
        } else if _, err := os.Stat(args[1]); err == nil {
            bundlePath = args[1]
            name = args[0]
        } else {
            // Default to original order with better error
            bundlePath = args[0]
            name = args[1]
        }
    } else {
        // 3 args: kind, bundle, name
        kind = strings.ToLower(args[0])
        // Try to detect which is bundle path
        if _, err := os.Stat(args[1]); err == nil {
            bundlePath = args[1]
            name = args[2]
        } else if _, err := os.Stat(args[2]); err == nil {
            bundlePath = args[2]
            name = args[1]
        } else {
            bundlePath = args[1]
            name = args[2]
        }
    }
    // ... normalize kind
}
```

## Acceptance Criteria

- [ ] `describe pod ./bundle/ mypod -n ns` works
- [ ] `describe pod -n ns mypod ./bundle/` works  
- [ ] `describe ./bundle/ mypod` (auto-detect kind) works
- [ ] Clear error when bundle path cannot be found
- [ ] Tests cover all argument order permutations

---

**Labels:** bug, cli, ux, sprint-8
**Priority:** Medium (has workaround, but affects common use case)
