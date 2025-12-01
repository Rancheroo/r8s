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
r8s --verbose [your-bundle-path]

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
ssh -t user@host r8s ./bundle-folder/

# CI/CD: r8s is not designed for automation
# Use kubectl/k9s/other tools for automation

# Check if you have a TTY
tty
# Should output: /dev/pts/0 (or similar)
# If "not a tty", run from interactive terminal
```

---

### 2. "not a directory"

**Error message:**
```
Error: ./bundle.tar.gz is not a directory

r8s only supports extracted bundle folders.

Extract the bundle first:
  tar -xzf bundle.tar.gz
  r8s ./extracted-folder/
```

**Cause:** Trying to analyze a .tar.gz file directly (no longer supported).

**Solution:**
```bash
# Extract the bundle first
tar -xzf bundle.tar.gz

# Then analyze the extracted folder
r8s ./extracted-folder/
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
   r8s --verbose
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
1. **Check directory structure:**
   ```bash
   # Wrong: pointing to parent folder
   r8s ./bundles/
   
   # Right: pointing to extracted bundle
   r8s ./bundles/w-guard-wg-cp-xyz/
   ```

2. **Verify structure:**
   ```bash
   ls bundle-name/
   # Should see: rke2/ directory
   
   ls bundle-name/rke2/
   # Should see: kubectl/, podlogs/, etc.
   ```

3. **Make sure bundle is extracted:**
   ```bash
   # List tarball contents first
   tar -tzf bundle.tar.gz | head -20
   
   # Extract if needed
   tar -xzf bundle.tar.gz
   
   # Analyze extracted folder
   r8s ./extracted-folder/
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
   r8s --verbose ./bundle/
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
   r8s
   ```

3. **Or use demo mode (no config needed):**
   ```bash
   r8s --mockdata
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
   r8s --insecure
   
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
   r8s --verbose ./bundle/
   # Shows parse errors
   ```

3. **Wrong directory:**
   ```bash
   # Make sure you're pointing to the extracted bundle folder
   find . -type d -name "rke2"
   ```

---

### Slow bundle loading

**Symptom:** Bundle takes >10 seconds to load.

**Causes:**
- Very large bundle (hundreds of MBs extracted)
- Thousands of pods/resources
- Slow disk (networked storage)

**Solutions:**
```bash
# Extract to faster disk if possible
tar -xzf bundle.tar.gz -C /tmp/
r8s /tmp/extracted-bundle/

# Use local SSD instead of networked storage
# (extraction is one-time cost)
```

---

### Bundle path not found

**Error:**
```
Error: bundle path not found: ./my-bundle/
```

**Solutions:**
1. **Check path is correct:**
   ```bash
   ls -la ./my-bundle/
   # Does it exist?
   
   # Use absolute path if needed
   r8s /full/path/to/bundle/
   ```

2. **Check bundle is extracted:**
   ```bash
   # If you have bundle.tar.gz, extract it first
   tar -xzf bundle.tar.gz
   r8s ./extracted-folder/
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
   r8s
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
   r8s --verbose ./bundle/
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
r8s --verbose ./bundle/ 2>&1 | tee debug.log
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
   r8s ./my-bundle/
   ```

3. **Verbose output:**
   ```bash
   r8s --verbose ./my-bundle/ 2>&1
   ```

4. **Environment:**
   - OS: `uname -a`
   - Go version: `go version`
   - Terminal: `echo $TERM`

5. **Bundle structure (if bundle issue):**
   ```bash
   ls -la bundle-folder/
   ls -la bundle-folder/rke2/
   ```

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
4. **Directory-only bundles:**
   - Bundles must be extracted first
   - No automatic tarball extraction
5. **TTY required:**
   - Cannot run without terminal
   - Not suitable for cronjobs/automation

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
   r8s --mockdata
   
   # If this works, issue is with config/bundle
   ```

4. **Ask for help:**
   - GitHub Discussions
   - GitHub Issues
   - See [Contributing Guide](CONTRIBUTING.md)

---

**Last Updated:** 2025-12-01  
**r8s Version:** 0.2.1+
