# Future Work

This document tracks known issues and planned enhancements for future releases.

## Known Issues

### Emoji Alignment in Attention Dashboard

**Issue**: Column alignment in the attention dashboard can be inconsistent when emoji display widths vary across different terminal emulators and fonts.

**Root Cause**: 
- Emojis have varying display widths (1-cell vs 2-cell) depending on terminal/font
- Current implementation uses `runewidth` library and manual padding
- However, actual rendering varies by:
  - Terminal emulator (iTerm2, Alacritty, gnome-terminal, etc.)
  - Font choice and version
  - Unicode version support
  - System locale settings

**Impact**: LOW - Cosmetic only, does not affect functionality

**Status**: 
- v0.6.8.1: Attempted fix using runewidth.StringWidth() + manual padding
- Needs testing across multiple terminal environments
- May require terminal-specific detection or user configuration

**Potential Solutions**:
1. Use monospaced ASCII art instead of emojis (sacrifices visual appeal)
2. Detect terminal capabilities and adjust rendering
3. Add user configuration option for emoji display mode
4. Switch to table library with better Unicode width support
5. Accept as terminal-dependent behavior and document it

**Target**: v0.7.x or later (after core functionality is stable)

## Planned Enhancements

### Native Bundle Comparison (`r8s diff`)

**Feature**: Compare two support bundles to identify changes in cluster state, node health, or configuration over time.

**Value**: Critical for "what changed?" root cause analysis.

**Proposed Usage**:
```bash
r8s diff ./bundle-old/ ./bundle-new/
r8s diff ./bundle-old/ ./bundle-new/ --format=markdown > diff_report.md
```

### v0.7.x Roadmap

- TBD: Add more issues as they are identified

---

*Last Updated: v0.6.8.1*
