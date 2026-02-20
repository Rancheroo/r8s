# GitHub Release Instructions — v0.8.0-alpha

## Option 1: Automated (Recommended)

Run the release script:

```bash
cd /home/node/.openclaw/workspace/r8s
chmod +x scripts/create-release.sh
gh auth login  # If not already authenticated
./scripts/create-release.sh
```

## Option 2: Manual Web UI

### Step 1: Go to Releases Page
Navigate to: https://github.com/Rancheroo/r8s/releases

### Step 2: Create New Release
Click "Draft a new release"

### Step 3: Fill in Release Details

**Choose a tag:**
- Select existing: `v0.8.0-alpha`

**Release title:**
```
v0.8.0-alpha — kubectl for Rancher Bundles
```

**Description:**
Copy from `GITHUB_RELEASE_v0.8.0-alpha.md`

**Set as:**
- ☑️ Set as a pre-release

### Step 4: Upload Binaries

Attach the following files from `bin/` directory:

1. `r8s-v0.8.0-alpha-linux-amd64` (7.4 MB)
2. `r8s-v0.8.0-alpha-linux-arm64` (7.0 MB)
3. `r8s-v0.8.0-alpha-darwin-amd64` (7.5 MB)
4. `r8s-v0.8.0-alpha-darwin-arm64` (7.2 MB)
5. `r8s-v0.8.0-alpha-windows-amd64.exe` (7.7 MB)
6. `checksums.txt` (SHA256 checksums)

### Step 5: Publish Release

Click "Publish release"

---

## Binary Details

| Platform | Architecture | Size | SHA256 |
|----------|-------------|------|--------|
| Linux | amd64 | 7.4 MB | `a89e776871208cddaccc91c7ec06d661d7b5be6c23db5feb03b9cde7592a4540` |
| Linux | arm64 | 7.0 MB | `970c7c9b0f05fe254614e47de0b6c40fb37ac48906ef0ed516641d5598824769` |
| macOS | amd64 | 7.5 MB | `d8ae991ae04c185308ab26f797e7a658eda4559a8cfb648c4cc6dfbfa013a831` |
| macOS | arm64 (M1/M2) | 7.2 MB | `7c8b4abe1bd31272b462702442c9c5922e7a07dca5d00545a35118bb84d647b6` |
| Windows | amd64 | 7.7 MB | `fded68d55830c7b939211057b0b391a3790401b984035e1db6bef15193a0b6c2` |

---

## Verification Commands

### Linux/macOS
```bash
# Download
curl -L https://github.com/Rancheroo/r8s/releases/download/v0.8.0-alpha/r8s-v0.8.0-alpha-linux-amd64 -o r8s

# Verify checksum
sha256sum r8s
# Expected: a89e776871208cddaccc91c7ec06d661d7b5be6c23db5feb03b9cde7592a4540

# Make executable and test
chmod +x r8s
./r8s version
```

### Windows
```powershell
# Download
Invoke-WebRequest -Uri "https://github.com/Rancheroo/r8s/releases/download/v0.8.0-alpha/r8s-v0.8.0-alpha-windows-amd64.exe" -OutFile "r8s.exe"

# Verify checksum (PowerShell)
Get-FileHash r8s.exe -Algorithm SHA256
# Expected: fded68d55830c7b939211057b0b391a3790401b984035e1db6bef15193a0b6c2
```

---

## Post-Release Checklist

- [ ] Release page shows all 5 binaries + checksums
- [ ] Release notes are complete and formatted
- [ ] Tagged as pre-release
- [ ] Installation instructions work
- [ ] Announce in team channels
