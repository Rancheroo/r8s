<!-- This is a machine-generated file. To regenerate it, run: make docs -->
# Troubleshooting Guide

Common issues and solutions for r8s (Rancheroos).

## Table of Contents
- [Quick Diagnostics](#quick-diagnostics)
- [Common Errors](#common-errors)
- [Bundle Issues](#bundle-issues)
- [Configuration Issues](#configuration-issues)
- [TUI Issues](#tui-issues)
- [Getting Help](#getting-help)

---

## Quick Diagnostics

Before troubleshooting, gather this information:

```bash
# Version
r8s version

# Verbose mode
r8s --verbose tui [your flags]

# Config validation
r8s config validate
```

---

## Common Errors

### 1. "could not open a new TTY"

**Error message:**
```
Error: TUI error: could not open a new TTY: open /dev/tty: no such device or address
```

**Cause:** r8s requires a real terminal (TTY) to run. This occurs when:
- Running in CI/CD pipeline
- Using SSH without pseudo-TTY
- Piping input/output
- Running in non-interactive environment

**Solutions:**
```bash
# SSH: Force pseudo-TTY allocation
ssh -t user@host r8s tui

# CI/CD: Use bundle info instead (no TUI)
r8s bundle info ./bundle.tar.gz

# Check if you have a TTY
tty
# Should output: /dev/pts/0 (or similar)
# If "not a tty", run from interactive terminal
```

---

### 2. "bundle size exceeds limit"

**Error message:**
```
Error: bundle size (75.4 MB) exceeds limit (50.0 MB)
Solution: Use --limit=80 to increase (e.g. 'r8s tui --bundle=bundle.tar.gz --limit=80')
```

**Cause:** Default 50MB size limit for .tar.gz files.

**Solutions:**
```bash
# Option 1: Increase limit
r8s tui --bundle=bundle.tar.gz --limit=100

# Option 2: Extract first (recommended, no limits)
tar -xzf bundle.tar.gz
r8s tui --bundle=./extracted-folder/

# Option 3: Unlimited (use with caution)
r8s tui --bundle=bundle.tar.gz --limit=0
```

---

### 3. "connection refused"

**Error message:**
```
Error: failed to connect to Rancher: dial tcp 192.168.1.100:443: connection refused
```

**Cause:** Cannot reach Rancher API.

**Solutions:**
1. **Verify URL:**
   ```bash
   # Check Rancher URL in config
   r8s config view
   
   # Test connectivity
   curl -k https://rancher.example.com/ping
   ```

2. **Check network:**
   ```bash
   # Can you ping the server?
   ping rancher.example.com
   
   # Is port 443 open?
   telnet rancher.example.com 443
   ```

3. **Firewall rules:**
   ```bash
   # Check if firewall is blocking
   sudo iptables -L -n | grep 443
   ```

4. **Use verbose mode:**
   ```bash
   r8s --verbose tui
   # Shows detailed connection attempts
   ```

---

### 4. "authentication failed"

**Error message:**
```
Error: authentication failed: 401 Unauthorized
```

**Cause:** Invalid, expired, or incorrectly formatted token.

**Solutions:**
1. **Generate new token:**
   - Log in to Rancher UI
   - Avatar → Account & API Keys
   - Create API Key
   - Copy token immediately

2. **Update config:**
   ```bash
   r8s config set token token-xxxxx:yyyyyyyy
   ```

3. **Verify token format:**
   ```yaml
   # Correct format:
   bearerToken: token-xxxxx:yyyyyyyy
   
   # NOT this:
   bearerToken: token-xxxxx
   ```

4. **Check expiration:**
   - Tokens can expire
   - Create new token without expiration
   - Or set long expiration (1 year)

---

### 5. "not a valid RKE2 bundle"

**Error message:**
```
Error: not a valid RKE2 bundle
Missing: rke2/ directory
```

**Cause:** Pointing to wrong directory or unsupported bundle type.

**Solutions:**
1. **Check directory:**
   ```bash
   # Wrong: pointing to parent
   r8s tui --bundle=./bundles/
   
   # Right: pointing to extracted bundle
   r8s tui --bundle=./bundles/w-guard-wg-cp-xyz/
   ```

2. **Verify structure:**
   ```bash
   ls bundle-name/
   # Should see: rke2/ directory
   
   ls bundle-name/rke2/
   # Should see: kubectl/, podlogs/, etc.
   ```

3. **Extract if needed:**
   ```bash
   tar -tzf bundle.tar.gz | head -20
   # Shows bundle contents
   
   tar -xzf bundle.tar.gz
   ```

---

### 6. "panic: interface conversion: nil"

**Status:** ✅ **FIXED in v0.2.1**

**Historical error:**
```
panic: interface conversion: interface {} is nil, not string
```

**Cause:** Malformed kubectl YAML in older bundles.

**If you encounter this:**
1. Update to latest r8s version:
   ```bash
   cd r8s
   git pull
   make build
   ```

2. Use verbose mode:
   ```bash
   r8s --verbose tui --bundle=./bundle/
   ```

3. Report bug if still occurring (should not happen!)

---

### 7. "config file not found"

**Error message:**
```
Error: failed to load config: open /home/user/.r8s/config.yaml: no such file or directory
```

**Cause:** Config file doesn't exist yet.

**Solutions:**
1. **Initialize config:**
   ```bash
   r8s config init
   ```

2. **Or use environment variables:**
   ```bash
   export RANCHER_URL=https://rancher.example.com
   export RANCHER_TOKEN=token-xxxxx:yyyyyyyy
   r8s tui
   ```

3. **Or use demo mode (no config needed):**
   ```bash
   r8s tui --mockdata
   ```

---

### 8. "TLS certificate verification failed"

**Error message:**
```
Error: x509: certificate signed by unknown authority
```

**Cause:** Self-signed certificate or corporate CA.

**Solutions:**
1. **Development environment (self-signed cert):**
   ```bash
   # Option 1: Use insecure flag
   r8s tui --insecure
   
   # Option 2: Set in config
   r8s config set insecure true
   ```

2. **Production environment:**
   ```bash
   # Install CA certificate
   sudo cp ca-cert.crt /usr/local/share/ca-certificates/
   sudo update-ca-certificates
   ```

⚠️ **Security Warning:** Only use `--insecure` in development. Never in production!

---

## Bundle Issues

### Bundle loaded with warnings

**Message:**
```
⚠️ Warning: Bundle loaded with warnings
• Skipped 3 pod entries due to parse errors
• Missing rke2/kubectl/services file
```

**Meaning:** r8s loaded the bundle but encountered non-critical issues.

**Actions:**
- This is usually OK - r8s loads what it can
- Use `--verbose` to see details
- If too many warnings, bundle may be incomplete

---

### Empty resource lists

**Symptom:** TUI shows "No pods available" but bundle should have pods.

**Causes:**
1. **File actually empty:**
   ```bash
   ls -lh bundle/rke2/kubectl/pods
   # Shows 0 bytes
   ```

2. **Malformed YAML:**
   ```bash
   r8s --verbose tui --bundle=./bundle/
   # Shows parse errors
   ```

3. **Wrong directory:**
   ```bash
   # Make sure you're pointing to the extracted folder
   find . -type d -name "rke2"
   ```

---

### Slow bundle loading

**Symptom:** Bundle takes >10 seconds to load.

**Causes:**
- Very large bundle (100MB+)
- Thousands of pods
- Slow disk (networked storage)

**Solutions:**
```bash
# Use extracted folder (faster than .tar.gz)
tar -xzf bundle.tar.gz
r8s tui --bundle=./extracted/

# Reduce bundle size by filtering
# (when generating bundle on cluster)
```

---

## Configuration Issues

### Profile not found

**Error:**
```
Error: profile 'production' not found
```

**Solutions:**
1. **List available profiles:**
   ```bash
   r8s config view
   ```

2. **Create missing profile:**
   ```bash
   r8s config edit
   # Add new profile in YAML
   ```

3. **Use default profile:**
   ```bash
   r8s tui
   # Uses currentProfile from config
   ```

---

### Invalid YAML syntax

**Error:**
```
Error: failed to parse config file: yaml: line 5: mapping values are not allowed in this context
```

**Cause:** Syntax error in config.yaml.

**Solutions:**
1. **Validate YAML:**
   ```bash
   r8s config validate
   ```

2. **Check indentation:**
   ```yaml
   # Wrong:
   profiles:
   - name: test
     url: https://example.com
   
   # Right:
   profiles:
     - name: test
       url: https://example.com
   ```

3. **Recreate config:**
   ```bash
   mv ~/.r8s/config.yaml ~/.r8s/config.yaml.backup
   r8s config init
   ```

---

## TUI Issues

### Blank screen / no output

**Symptoms:** TUI starts but shows nothing.

**Solutions:**
1. **Check terminal size:**
   ```bash
   tput cols  # Should be >= 80
   tput lines # Should be >= 24
   ```

2. **Try different terminal:**
   - gnome-terminal
   - iTerm2 (macOS)
   - Windows Terminal

3. **Use verbose mode:**
   ```bash
   r8s --verbose tui
   ```

---

### Colors not working

**Symptoms:** No colors, or strange characters displayed.

**Solutions:**
1. **Check TERM variable:**
   ```bash
   echo $TERM
   # Should be: xterm-256color or similar
   
   # Set if needed:
   export TERM=xterm-256color
   ```

2. **Enable color support:**
   ```bash
   # Add to ~/.bashrc or ~/.zshrc
   export COLORTERM=truecolor
   ```

---

### Keyboard shortcuts not working

**Problem:** Keys like `j/k` or `Ctrl+R` don't work.

**Solutions:**
1. **Check terminal key bindings:**
   - Some terminals intercept certain key combos
   - Try alternate keys (arrow keys instead of j/k)

2. **Disable terminal shortcuts:**
   - gnome-terminal: Edit → Preferences → Shortcuts
   - iTerm2: Preferences → Keys

3. **Use mouse:**
   - r8s supports mouse clicks for navigation

---

## Getting Help

### Enable Verbose Mode

Always use `--verbose` when reporting issues:

```bash
r8s --verbose tui --bundle=./bundle/ 2>&1 | tee debug.log
```

This shows:
- Detailed error messages
- File paths being checked
- Parse warnings
- Connection attempts

---

### Collect Debug Information

When asking for help, include:

1. **Version:**
   ```bash
   r8s version
   ```

2. **Command used:**
   ```bash
   # Exact command that failed
   r8s tui --bundle=./my-bundle/
   ```

3. **Verbose output:**
   ```bash
   r8s --verbose tui --bundle=./my-bundle/ 2>&1
   ```

4. **Environment:**
   - OS: `uname -a`
   - Go version: `go version`
   - Terminal: `echo $TERM`

---

### Report a Bug

Use the bug report template:

```bash
# On GitHub
https://github.com/Rancheroo/r8s/issues/new?template=bug_report.md
```

Or:
```bash
# Command to open issue (if gh CLI installed)
gh issue create --repo Rancheroo/r8s --template bug_report.md
```

Required information:
- r8s version
- Mode (live/bundle/mock)
- Steps to reproduce
- Verbose output
- Bundle details (if applicable)

---

## Known Limitations

These are expected behavior, not bugs:

1. **Read-only:** r8s cannot modify resources
2. **No real-time watch:** Must manually refresh (press `r`)
3. **Bundle mode limitations:**
   - Only shows snapshot from collection time
   - No live updates
   - Describe limited to static YAML

4. **Size limits:**
   - 50MB default for .tar.gz
   - No limit for extracted folders

5. **TTY required:**
   - Cannot run without terminal
   - Not suitable for cronjobs/automation
   - Use `bundle info` instead

---

## Still Having Issues?

1. **Check existing issues:**
   ```bash
   # Search GitHub issues
   https://github.com/Rancheroo/r8s/issues
   ```

2. **Read documentation:**
   - [CLI Usage Guide](docs/USAGE.md)
   - [Bundle Format](docs/BUNDLE-FORMAT.md)
   - [Architecture](docs/ARCHITECTURE.md)

3. **Try demo mode:**
   ```bash
   # Test r8s without any configuration
   r8s tui --mockdata
   
   # If this works, issue is with config/bundle
   ```

4. **Ask for help:**
   - GitHub Discussions
   - GitHub Issues
   - See [Contributing Guide](CONTRIBUTING.md)

---

**Last Updated:** 2025-12-01  
**r8s Version:** 0.2.1+
