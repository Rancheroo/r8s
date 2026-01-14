# Future Work

This document tracks known issues and planned enhancements for future releases.

## Critical Issues (Blocking)

### Pod Diagnostics Broken in Drill-Down View (v0.6.8.1 Regression)

**Issue**: Number keys (1-9) in cluster event drill-down panel do not navigate to pod diagnostics. Issue #20

**Root Cause**: 
- `handleClusterEventPodSelection()` function has a bug in pod selection logic
- Introduced in v0.6.8.1 hotfix
- Suspected issues:
  - Namespace extraction from pod data
  - Event item matching logic
  - Navigation command not being returned properly

**Impact**: **CRITICAL** - Breaks primary use case for cluster event investigation

**Symptoms**:
1. Navigate to cluster event item (e.g., "FailedKillPod")
2. Press Enter → Drill-down panel appears ✅
3. Press "1" to view first pod → **Nothing happens** ❌
4. User remains on drill-down panel with no feedback

**Workaround**: Navigate to pods directly from attention dashboard (if pod item exists)

**Fix Priority**: **URGENT** - Must be fixed in v0.6.8.2 immediately

**Testing Gap Identified**:
- No automated tests for drill-down navigation
- No test coverage for number key handling in cluster event view
- Pod selection flow not validated

**Target**: v0.6.8.2 (immediate hotfix)

**GitHub Issue**: https://github.com/Rancheroo/r8s/issues/20

---

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

### v0.7.x Roadmap

- TBD: Add more issues as they are identified

---

*Last Updated: v0.6.8.1*
