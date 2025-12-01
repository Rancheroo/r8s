---
name: Bug Report
about: Report a bug or issue with r8s
title: '[BUG] '
labels: bug
assignees: ''
---

## Bug Description
<!-- A clear and concise description of what the bug is -->

## r8s Version
```bash
# Run this command and paste the output:
r8s version
```

## Mode Used
<!-- Check ONE of the following -->
- [ ] Live cluster (`r8s tui`)
- [ ] Bundle mode (`r8s tui --bundle=...`)
- [ ] Mock mode (`r8s tui --mockdata`)

## Bundle Information (if applicable)
<!-- If using bundle mode, provide bundle details -->
- Bundle size: 
- Bundle source: <!-- e.g., RKE2, kubectl cluster-info dump -->
- Can you share the bundle? (yes/no): 
- Bundle download link (if yes): 

## Steps to Reproduce
<!-- Exact steps to reproduce the behavior -->
1. 
2. 
3. 

## Expected Behavior
<!-- What you expected to happen -->

## Actual Behavior
<!-- What actually happened -->

## Verbose Output
<!-- Run the command with --verbose flag and paste the output -->
```bash
r8s --verbose tui [your flags here] 2>&1
```

<details>
<summary>Verbose output</summary>

```
<!-- Paste verbose output here -->
```

</details>

## Screenshots
<!-- If applicable, add screenshots to help explain your problem -->

## Environment
- OS: <!-- e.g., Ubuntu 22.04, macOS 14.0, Windows 11 -->
- Terminal: <!-- e.g., gnome-terminal, iTerm2, Windows Terminal -->
- Go version: <!-- Run: go version -->

## Additional Context
<!-- Add any other context about the problem here -->

## Checklist
- [ ] I have searched existing issues to ensure this is not a duplicate
- [ ] I have included the r8s version output
- [ ] I have included verbose output showing the error
- [ ] I have tested with the latest version of r8s
