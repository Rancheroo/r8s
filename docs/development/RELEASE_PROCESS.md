# Release Process — One Command

## Quick Start

```bash
cd /home/node/.openclaw/workspace/r8s
./scripts/release.sh v0.8.1
```

That's it. The script handles everything.

---

## What the Script Does

1. **Pre-flight Checks**
   - Verifies git, go, curl are installed
   - Extracts GitHub PAT from git remote URL
   - Validates token works
   - Checks working directory is clean
   - Validates version format

2. **Tests**
   - Runs `make ci` (lint + test + build)
   - Fails fast if tests don't pass

3. **Build**
   - Compiles for 5 platforms:
     - linux/amd64
     - linux/arm64
     - darwin/amd64 (Intel)
     - darwin/arm64 (Apple Silicon)
     - windows/amd64
   - Injects version, commit, date into binary

4. **Checksums**
   - Generates SHA256 for all binaries
   - Creates `checksums.txt`

5. **GitHub Release**
   - Creates release via API
   - Uploads all 6 assets
   - Verifies release is accessible

6. **Summary**
   - Prints release URL
   - Shows quick install command
   - Lists all assets

---

## Prerequisites

- Git repository with GitHub PAT in remote URL
- `git`, `go`, `curl` installed
- `jq` recommended (optional, for better JSON parsing)

---

## Environment Variables

| Variable | Description | Fallback |
|----------|-------------|----------|
| `GITHUB_TOKEN` | GitHub PAT | Extracted from git remote |

---

## Example Output

```
╔════════════════════════════════════════════════════════════╗
║        r8s Release Script — Fully Automated               ║
╚════════════════════════════════════════════════════════════╝

[INFO] Preparing release: v0.8.1
[INFO] Commit: 9dbdfc8, Date: 2026-02-18T08:30:00Z

═══ Pre-flight Checks ═══
[✓] GitHub token verified
[✓] Pre-flight checks passed

═══ Running Tests ═══
[INFO] Running make ci...
[✓] All tests passed

═══ Building Binaries ═══
[INFO] Building r8s-v0.8.1-linux-amd64...
[✓] Built r8s-v0.8.1-linux-amd64 (7.4M)
...
[✓] All binaries built successfully

═══ Generating Checksums ═══
[✓] Checksums generated

═══ Creating GitHub Release ═══
[INFO] Creating release v0.8.1...
[✓] Release created (ID: 287573746)
[INFO] URL: https://github.com/Rancheroo/r8s/releases/tag/v0.8.1

═══ Uploading Assets ═══
[INFO] Uploading r8s-v0.8.1-linux-amd64...
[✓] r8s-v0.8.1-linux-amd64 uploaded
...
[✓] All assets uploaded

═══ Verifying Release ═══
[✓] Release verified

═══ Release Summary ═══

✅ Release v0.8.1 Complete!

Release URL:
  https://github.com/Rancheroo/r8s/releases/tag/v0.8.1

Quick Install (Linux):
  curl -L https://github.com/Rancheroo/r8s/releases/download/v0.8.1/r8s-v0.8.1-linux-amd64 -o r8s
  chmod +x r8s && sudo mv r8s /usr/local/bin/
```

---

## Error Handling

The script fails fast with clear error messages:

```
[✗] GitHub token verification failed (HTTP 401)
```

```
[✗] Tests failed. See /tmp/ci-output.log
```

```
[✗] Build failed for linux/arm64. See /tmp/build-linux-arm64.log
```

---

## Dry Run (Test Without Publishing)

To test the build process without creating a release:

```bash
# Just build binaries
make ci
make build
ls -lh bin/

# Or manually run script steps
./scripts/release.sh v0.8.1-test
# Then delete the test tag: git tag -d v0.8.1-test
```

---

## Manual Fallback

If the automated script fails, use the manual process:

1. Build: `make ci && make build`
2. Create release via GitHub web UI
3. Upload binaries manually

See `RELEASE_INSTRUCTIONS.md` for detailed steps.

---

## Troubleshooting

### "No GitHub PAT found"
- Ensure git remote URL contains PAT: `git remote get-url origin`
- Or set `GITHUB_TOKEN` environment variable

### "GitHub token verification failed"
- Token may be expired
- Check token has `repo` scope

### "Working directory has uncommitted changes"
- Commit or stash changes before releasing
- Or use `--force` flag (not recommended)

---

## CI/CD Integration

For fully automated releases via GitHub Actions:

```yaml
# .github/workflows/release.yml
on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Release
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          ./scripts/release.sh ${{ github.ref_name }}
```

---

**Next release: Just run `./scripts/release.sh v0.8.1`**
