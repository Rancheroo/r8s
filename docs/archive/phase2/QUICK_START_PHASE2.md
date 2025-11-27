# Phase 2 Quick Start Guide

**Date:** November 27, 2025  
**Current Status:** Phase 1 Complete ✅ - Ready for Phase 2  
**Next Goal:** Pager Integration for Log Viewing

---

## 🎯 Phase 2 Objective

Add scrolling, search, and advanced log viewing capabilities using a pager-like interface (similar to `less`).

---

## ⚡ Quick Context

### What's Working (Phase 1)
- ✅ Press 'l' on any pod → view logs
- ✅ 16 mock log lines displayed
- ✅ Clean bordered display
- ✅ Full navigation integration

### What's Missing (Phase 2 Goals)
- ❌ Scrolling through long logs
- ❌ Search/filter within logs
- ❌ Live tail mode (-f)
- ❌ Container selection (multi-container pods)
- ❌ Log level filtering
- ❌ Export logs to file

---

## 🚨 CRITICAL: Fix This First

**Before starting Phase 2, fix the selection indicator issue:**

**Problem:** Users cannot see which row is selected in lists  
**Impact:** CRITICAL - Makes navigation confusing  
**Effort:** 1-2 hours  
**File:** `internal/tui/app.go`

**Implementation Hint:**
```go
// In renderPodsView(), renderDeploymentsView(), etc.
selectedStyle := lipgloss.NewStyle().
    Background(lipgloss.Color("62")).  // Dark blue
    Foreground(lipgloss.Color("230"))  // White

// Apply to selected row
if i == selectedIndex {
    row = selectedStyle.Render(row)
}
```

**Test Plan:** See TEST_REPORT_V2.md Section "Critical Issues #1"

---

## 📋 Phase 2 Implementation Plan

### Step 1: Add Viewport for Scrolling (4 hours)

**Goal:** Make logs scrollable with j/k keys

**Tasks:**
1. Add `github.com/charmbracelet/bubbles/viewport` dependency
2. Create viewport in App struct for logs view
3. Handle j/k/↑/↓ keys for scrolling
4. Add g/G for top/bottom
5. Show scroll position indicator

**Files to modify:**
- `internal/tui/app.go` - Add viewport field, update renderLogsView()

**Success criteria:**
- Can scroll through logs with j/k
- Scroll indicator shows position (e.g., "45%")
- g jumps to top, G jumps to bottom

### Step 2: Add Search Functionality (3 hours)

**Goal:** Press '/' to search logs (like less)

**Tasks:**
1. Add search mode state
2. Handle '/' key to enter search mode
3. Capture search query input
4. Highlight matches in logs
5. n/N for next/previous match

**Files to modify:**
- `internal/tui/app.go` - Add search state, search logic

**Success criteria:**
- '/' enters search mode
- Search highlights matches
- n/N navigate matches
- Esc exits search mode

### Step 3: Add Container Selection (2 hours)

**Goal:** Multi-container pod support

**Tasks:**
1. Detect pods with multiple containers
2. Show container selection menu if >1 container
3. Add 'c' key to switch containers
4. Update breadcrumb to show container

**Files to modify:**
- `internal/tui/app.go` - Container detection, selection UI
- `internal/rancher/client.go` - Fetch container list

**Success criteria:**
- Multi-container pods show selector
- 'c' key switches between containers
- Breadcrumb shows active container

### Step 4: Add Live Tail Mode (3 hours)

**Goal:** Follow logs in real-time (like tail -f)

**Tasks:**
1. Add tail mode toggle ('t' key)
2. Implement auto-scrolling to bottom
3. Poll for new logs every 2s
4. Show "Following logs..." indicator

**Files to modify:**
- `internal/tui/app.go` - Tail mode state, auto-scroll
- `internal/rancher/client.go` - Stream logs API

**Success criteria:**
- 't' toggles tail mode
- Auto-scrolls to bottom
- New logs appear automatically
- 't' again stops following

### Step 5: Add Log Level Filtering (2 hours)

**Goal:** Filter by ERROR, WARN, INFO, DEBUG

**Tasks:**
1. Parse log level from lines
2. Add filter state (bitmask: ERROR|WARN|INFO|DEBUG)
3. Add hotkeys: E (errors), W (warnings), I (info), D (debug)
4. Show active filters in status bar

**Files to modify:**
- `internal/tui/app.go` - Filter state, rendering logic

**Success criteria:**
- E shows only ERROR lines
- W shows only WARN lines
- Filters can combine (E+W shows both)
- Clear indication of active filters

### Step 6: Add Export Functionality (2 hours)

**Goal:** Save logs to file

**Tasks:**
1. Add 's' key for save
2. Prompt for filename
3. Write logs to file (with filters applied)
4. Show success/error message

**Files to modify:**
- `internal/tui/app.go` - Export logic, file I/O

**Success criteria:**
- 's' prompts for filename
- Saves current filtered logs
- Shows confirmation message
- Handles errors gracefully

---

## 📊 Estimated Effort

| Step | Feature | Hours | Priority |
|------|---------|-------|----------|
| 0 | **Fix selection indicator** | 2 | **CRITICAL** |
| 1 | Viewport scrolling | 4 | High |
| 2 | Search functionality | 3 | High |
| 3 | Container selection | 2 | Medium |
| 4 | Live tail mode | 3 | High |
| 5 | Log level filtering | 2 | Medium |
| 6 | Export logs | 2 | Low |
| **Total** | | **18 hours** | |

**Realistic timeline:** 2-3 days of focused work

---

## 🧪 Testing Checklist

After implementing each step, verify:

- [ ] **Selection Indicator**
  - [ ] Selected row is highlighted
  - [ ] Highlight moves with j/k
  - [ ] Works in all views (Pods, Deployments, Services, CRDs)

- [ ] **Scrolling**
  - [ ] j/k scroll through logs
  - [ ] g goes to top
  - [ ] G goes to bottom
  - [ ] Scroll indicator shows position
  - [ ] Works with 100+ log lines

- [ ] **Search**
  - [ ] '/' enters search mode
  - [ ] Search query input works
  - [ ] Matches are highlighted
  - [ ] n/N navigate matches
  - [ ] Case-insensitive search
  - [ ] Regex support (optional)

- [ ] **Container Selection**
  - [ ] Multi-container pods detected
  - [ ] Container menu appears
  - [ ] 'c' switches containers
  - [ ] Logs update on switch
  - [ ] Breadcrumb shows container

- [ ] **Live Tail**
  - [ ] 't' toggles tail mode
  - [ ] Auto-scrolls to bottom
  - [ ] New logs appear (mock: simulate new lines)
  - [ ] Status shows "Following..."
  - [ ] 't' again stops following

- [ ] **Filtering**
  - [ ] E filters to ERROR only
  - [ ] W filters to WARN only
  - [ ] I filters to INFO only
  - [ ] D filters to DEBUG only
  - [ ] Filters combine (E+W)
  - [ ] Status shows active filters

- [ ] **Export**
  - [ ] 's' prompts for filename
  - [ ] File is created
  - [ ] Contains expected logs
  - [ ] Respects active filters
  - [ ] Error handling works

---

## 📚 Key References

### Code Files
- `internal/tui/app.go` - Main TUI logic (currently 500+ lines)
- `internal/rancher/client.go` - API client
- `internal/tui/styles.go` - Styling

### Documentation
- `PHASE1_COMPLETE.md` - What was just built
- `TEST_REPORT_V2.md` - Test results and issues
- `R8S_MIGRATION_PLAN.md` - Overall migration plan
- `STATUS.md` - Current project status

### Dependencies to Add
```go
// go.mod additions needed:
github.com/charmbracelet/bubbles/viewport  // For scrolling
```

---

## 🎯 Success Criteria for Phase 2

**Phase 2 is complete when:**
1. ✅ Selection indicator is visible (CRITICAL)
2. ✅ Logs are scrollable with j/k/g/G
3. ✅ Search works with '/' and n/N
4. ✅ Multi-container pods supported
5. ✅ Live tail mode works
6. ✅ Log level filtering works
7. ✅ Export to file works
8. ✅ All features tested (manual + unit tests)
9. ✅ Build passes with zero errors
10. ✅ Documentation updated

**Then ready for:** Phase 3 (Log Highlighting & Advanced Filtering)

---

## 🚀 Getting Started

### 1. Fix Selection Indicator (Do This First!)
```bash
# Edit internal/tui/app.go
# Add selectedStyle to renderPodsView, renderDeploymentsView, etc.
# Test: Navigate with j/k and see highlight move
```

### 2. Start Phase 2 Implementation
```bash
# Add viewport dependency
go get github.com/charmbracelet/bubbles/viewport

# Start with Step 1 (scrolling)
# Edit internal/tui/app.go
```

### 3. Test Incrementally
```bash
# After each step
go build ./...
./bin/r8s
# Manual testing
```

### 4. Document Progress
```bash
# Create PHASE2_PROGRESS.md as you go
# Update when each step completes
```

---

## 💡 Implementation Tips

### Viewport Integration
```go
import "github.com/charmbracelet/bubbles/viewport"

// In App struct
type App struct {
    // ... existing fields
    logViewport viewport.Model
}

// Initialize
func (a *App) handleViewLogs(...) {
    vp := viewport.New(width, height)
    vp.SetContent(strings.Join(logs, "\n"))
    a.logViewport = vp
}

// In Update()
case key.Matches(msg, a.keys.Down):
    if a.currentView == ViewLogs {
        a.logViewport, cmd = a.logViewport.Update(msg)
        return a, cmd
    }
```

### Search Highlighting
```go
// Use lipgloss to highlight matches
matchStyle := lipgloss.NewStyle().
    Background(lipgloss.Color("11")).  // Yellow
    Foreground(lipgloss.Color("0"))    // Black

// Highlight search term in log line
highlighted := strings.ReplaceAll(line, searchTerm, 
    matchStyle.Render(searchTerm))
```

### Log Level Parsing
```go
func parseLogLevel(line string) string {
    if strings.Contains(line, "ERROR") { return "ERROR" }
    if strings.Contains(line, "WARN")  { return "WARN" }
    if strings.Contains(line, "INFO")  { return "INFO" }
    if strings.Contains(line, "DEBUG") { return "DEBUG" }
    return "UNKNOWN"
}
```

---

## 📞 Need Help?

**Stuck on something?**
1. Check existing code patterns in `internal/tui/app.go`
2. Review Bubble Tea examples: https://github.com/charmbracelet/bubbletea/tree/master/examples
3. See viewport docs: https://github.com/charmbracelet/bubbles/tree/master/viewport
4. Reference TEST_REPORT_V2.md for expected behavior

**Found a bug?**
1. Note it in PHASE2_PROGRESS.md
2. Continue with implementation
3. Fix bugs at end of phase

---

**Good luck with Phase 2! 🚀**

*Remember: Fix the selection indicator first - it's critical for UX!*
