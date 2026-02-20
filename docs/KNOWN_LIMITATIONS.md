# Known Limitations

**Version:** v0.8.0-alpha  
**Last Updated:** 2026-02-19  

This document tracks known issues and limitations in r8s. For bugs, see GitHub Issues.

---

## Sprint 9 (v0.8.0-alpha)

### 001: Namespace Parsing in `r8s logs`

**Status:** Known limitation, workaround exists  
**Priority:** Low  
**Fix Target:** v0.8.1

#### Description
When log filenames use hyphen-delimited format without explicit namespace metadata, the parser may split incorrectly on the first hyphen.

**Example:**
- Filename: `calico-system-calico-node-88rjf.log`
- Parsed as: `namespace=calico`, `pod=system-calico-node-88rjf`
- Expected: `namespace=calico-system`, `pod=calico-node-88rjf`

#### Impact
- Using `-n cattle-system` may return no logs
- Using `-n cattle` (first segment) works as workaround
- Pod name filtering (`r8s logs ./bundle/ calico-node`) works correctly

#### Workaround
Use the first segment of the namespace:
```bash
# Instead of:
r8s logs ./bundle/ -n cattle-system

# Use:
r8s logs ./bundle/ -n cattle

# Or filter by pod name (recommended):
r8s logs ./bundle/ calico-node
```

#### Root Cause
Hyphen-delimited log filenames are ambiguous without additional bundle metadata. The parser splits on the first hyphen to extract namespace.

#### Future Fix
- Parse namespace from bundle metadata (if available)
- Require explicit namespace/pod directory structure
- Support bundle format version detection

---

### 002: Exit Code Inconsistencies (v0.8.0-alpha)

**Status:** Known limitation, documented  
**Priority:** Low (non-blocking)  
**Fix Target:** v0.8.1  
**Compliance:** 80% (8/10 tests pass)

#### Description
Some commands don't consistently return exit code 2 for errors as per CLI_STANDARDIZATION.md.

**Current Behavior:**
| Scenario | Expected | Actual | Commands Affected |
|----------|----------|--------|-------------------|
| Bundle not found | 2 | 1 | `export`, `describe`, `logs` |
| Resource not found | 1 | 0 | `describe` |

**Impact:**
- Low for interactive use
- CI/CD pipelines checking exit codes may get false positives
- Scripts expecting exit 2 for errors will see exit 1 instead

**Workaround:**
```bash
# Check for both exit 1 and 2 for errors
r8s export ./bundle/ || [ $? -le 2 ] || exit 1

# Or check stderr for specific errors
r8s describe ./bundle/ nonexistent 2>&1 | grep -q "No resources found"
```

#### Root Cause
Error handling in describe, export, and logs commands returns exit 1 instead of standard exit 2 for file/bundle errors.

#### Future Fix
- Update `runExport()`, `runDescribe()`, `runLogs()` to use `os.Exit(ExitError)` consistently
- Add unit tests for exit code validation

---

### 003: Flag Naming Inconsistency (v0.8.0-alpha)

**Status:** Known limitation, documented  
**Priority:** Medium  
**Fix Target:** v0.8.1  
**Compliance:** 83% (5/6 tests pass)

#### Description
`describe` command uses `-o` for output format instead of `--format` like other commands.

**Current Behavior:**
| Command | Format Flag |
|---------|-------------|
| `validate` | `--format` |
| `describe` | `-o` / `--output` |
| `export` | `--format` |

**Impact:**
- Users may try `describe --format=json` and it won't work
- Inconsistent with kubectl (which uses `-o` for output)
- Confusing for CLI consistency

**Workaround:**
```bash
# For describe, use -o instead of --format
r8s describe ./bundle/ node-1 -o json

# Not:
r8s describe ./bundle/ node-1 --format=json  # Doesn't work
```

#### Root Cause
`describe` was modeled after `kubectl describe -o json`, which uses `-o` for output format. Other commands followed `validate` pattern with `--format`.

#### Future Fix
- Add `--format` as alias to `-o` in describe command
- Or standardize all commands on `-o` (kubectl style)

---

### 004: Missing EXIT CODES in Help (v0.8.0-alpha)

**Status:** Known limitation, documented  
**Priority:** Low  
**Fix Target:** v0.8.1  
**Compliance:** 75% (3/4 tests pass)

#### Description
Some commands don't include EXIT CODES section in help text.

**Current State:**
| Command | EXIT CODES Section | Status |
|---------|-------------------|--------|
| `validate` | ✅ Yes | Complete |
| `describe` | ⚪️ No | Missing |
| `export` | ⚪️ No | Missing |
| `logs` | ⚪️ No | Missing |
| `completion` | ⚪️ N/A | Usage pattern |
| `generate prompt` | ✅ Yes | Complete |

**Impact:**
- Low - users can still use the commands
- Scripts need to test to discover exit codes
- Inconsistent documentation experience

**Workaround:**
Refer to CLI_STANDARDIZATION.md for standard exit codes:
- 0 = Success
- 1 = Issues found / Not found
- 2 = Error

#### Root Cause
Help text written before standardization was complete.

#### Future Fix
- Add EXIT CODES section to describe, export, logs help text
- Consider generating help text from templates

---

## Historical Limitations

### Sprint 8

No documented limitations. All features working as designed.

---

## Reporting New Limitations

If you discover behavior that:
- Works as designed but has edge cases
- Has workarounds but isn't ideal
- Needs documentation for clarity

Please:
1. Check this document first
2. If not listed, discuss in Code Sprint channel
3. If confirmed, we'll add it here

For bugs (unexpected errors, crashes), please file a GitHub Issue instead.

---

*This file is updated per release. See SPRINT9_CRIBSHEET.md for current status.*
