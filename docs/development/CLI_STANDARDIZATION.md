# CLI Standardization Guide

**Version:** v0.8.0  
**Status:** Sprint 9 Day 5  

This document defines the standards for r8s CLI commands.

---

## Exit Codes

All commands MUST use these exit codes:

| Code | Meaning | When to Use |
|------|---------|-------------|
| `0` | Success | Command completed, no issues found |
| `1` | Issues Found | Command completed but found issues (incomplete bundle, warnings) |
| `2` | Error | Command failed (invalid args, file not found, panic) |
| `130` | Cancelled | User pressed Ctrl+C (SIGINT) |

### Examples

```go
// Success
if everythingOK {
    os.Exit(0)
}

// Issues found (but command worked)
if bundleIncomplete {
    os.Exit(1)
}

// Error
if err != nil {
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    os.Exit(2)
}
```

---

## Output Formats

All commands that support `--format` MUST accept these values:

| Format | Description | Use Case |
|--------|-------------|----------|
| `human` | Pretty-printed, colorized | Interactive use (default) |
| `table` | Tabular format | Quick overview |
| `wide` | Extended table with more columns | Detailed overview |
| `json` | Machine-readable JSON | CI/CD, scripting |
| `yaml` | Machine-readable YAML | Config files, readability |

### Standard Format Flag

```go
// In your command
myCmd.Flags().StringVarP(&outputFormat, "format", "f", "human", cmd.FormatHelp())

// In RunE function
format := cmd.StandardizeFormat(outputFormat)
```

---

## Help Text Structure

All commands MUST include in their `Long` description:

1. **One-line description** (in `Short`)
2. **What it does** (in `Long`)
3. **EXAMPLES section** with copy-paste ready commands
4. **EXIT CODES section** documenting exit codes

### Template

```go
var myCmd = &cobra.Command{
    Use:   "mycommand [args]",
    Short: "Brief description",
    Long: `Detailed description of what this does.

EXAMPLES:
  # Basic usage
  r8s mycommand ./bundle/

  # With specific option
  r8s mycommand ./bundle/ --option=value

EXIT CODES:
  0 - Success
  1 - Issues found
  2 - Error`,
    RunE: runMyCommand,
}
```

---

## Flag Conventions

### Standard Flags

| Flag | Short | Type | Standard Description |
|------|-------|------|---------------------|
| `--format` | `-f` | string | Output format |
| `--output` | `-o` | string | Output file path |
| `--namespace` | `-n` | string | Filter by namespace |
| `--selector` | `-l` | string | Label selector |
| `--verbose` | `-v` | bool | Verbose output |
| `--help` | `-h` | bool | Show help |

### Flag Conflicts

**CRITICAL:** `logs` command uses `-f` for **follow mode** (like `kubectl`), NOT format.

This is the ONLY exception. All other commands use `-f` for `--format`.

---

## JSON Structure Standards

When outputting JSON, include these standard fields where applicable:

### Top-Level Structure

```json
{
  "meta": {
    "generated_at": "2026-02-19T12:00:00Z",
    "r8s_version": "v0.8.0"
  },
  "data": { /* command-specific data */ },
  "summary": {
    "total": 42,
    "status": "success"
  }
}
```

### Error Responses

```json
{
  "error": "descriptive error message",
  "code": 2,
  "details": { /* optional */ }
}
```

---

## Naming Conventions

### Commands

- Use lowercase: `describe`, not `Describe`
- Use single words where possible: `export`, not `export-findings`
- Use hyphen for compound: Not applicable (single words preferred)

### Resources

- Singular preferred: `pod`, not `pods`
- Accept aliases: `pods` → `pod`, `nodes` → `node`

### Arguments

- `[bundle-path]` always first positional arg
- `[name]` second positional arg (for specific resources)

---

## Testing Standards

### Exit Code Tests

```bash
# Test success
./bin/r8s command ./bundle/
[ $? -eq 0 ] && echo "PASS" || echo "FAIL"

# Test issues found
./bin/r8s command ./incomplete-bundle/
[ $? -eq 1 ] && echo "PASS" || echo "FAIL"

# Test error
./bin/r8s command /nonexistent/
[ $? -eq 2 ] && echo "PASS" || echo "FAIL"
```

### Format Tests

```bash
# All should produce valid output
./bin/r8s command ./bundle/ --format=json | jq .
./bin/r8s command ./bundle/ --format=yaml
./bin/r8s command ./bundle/ --format=table
```

---

## Implementation Checklist

When creating a new command:

- [ ] Add to `rootCmd` in `init()`
- [ ] Define `Use`, `Short`, `Long` with EXAMPLES and EXIT CODES
- [ ] Use standard exit codes (0, 1, 2)
- [ ] Support `--format` flag (except where inappropriate)
- [ ] Use `StandardizeFormat()` for format normalization
- [ ] Include `--help` support (automatic with Cobra)
- [ ] Add test plan to `docs/testing/`
- [ ] Test exit codes manually
- [ ] Test all format options

---

## Migration Notes (from v0.7.x)

### Changed in v0.8.0

| Old | New | Breaking? |
|-----|-----|-----------|
| Default TUI launch | Must use `r8s dashboard` | YES |
| `r8s ./bundle/` | `r8s validate ./bundle/` | YES |
| `--format` values varied | Now standardized | Minor |
| Exit codes inconsistent | Now standardized (0,1,2) | Minor |

---

*Standards enforced from v0.8.0 onward*
