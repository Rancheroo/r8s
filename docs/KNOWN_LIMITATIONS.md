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
