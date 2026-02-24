# Sprint 11 Deferred Issues

**Status:** v0.9.0 Release Deferred Items
**Date:** February 24, 2026

The following issues were identified during Sprint 11 testing but deferred to v0.9.1/v0.9.5 to unblock the release.

---

## 🐛 Bugs

### [Medium] Export to stdout not working
- **Description:** Running `r8s export` without `--output` flag should print to stdout, but currently outputs nothing.
- **Impact:** Cannot pipe export results to other tools.
- **Workaround:** Use `--output /dev/stdout` or temporary file.
- **Target:** v0.9.1

### [Low] Namespace Data Consistency
- **Description:** Some detected pods (e.g., `r8s-test-crash-segfault`) reference namespaces that don't appear in `r8s get ns`.
- **Impact:** Potential confusion if investigating non-existent namespaces.
- **Action:** Add validation step to verify namespaces exist.
- **Target:** v0.9.5

---

## 🚀 Enhancements

### [High] Template Variable Validation
- **Description:** Add build-time check to ensure all template variables (`{{.Var}}`) have corresponding regex capture groups (`(?P<Var>...)`).
- **Impact:** Prevents `<no value>` regressions.
- **Target:** v0.9.1

### [Medium] Progress Bar Visibility
- **Description:** Progress indicator in verbose mode is sometimes overwritten or not visible.
- **Action:** Use a proper progress bar library (e.g., `schollz/progressbar`).
- **Target:** v0.9.5

### [Low] Custom Pattern Support
- **Description:** Allow users to provide their own YAML pattern files via `--patterns-file`.
- **Target:** v0.10.0
