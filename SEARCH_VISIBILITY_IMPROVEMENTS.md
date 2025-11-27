# Search Visibility Improvements

**Date**: 2025-11-27  
**Issue**: Search works functionally but lacks visual feedback  
**Status**: Mock logs improved, highlighting planned for Phase 3

---

## The Problem

User reported that search functionality is hard to verify because:

1. **No visual highlighting** - Matched text isn't highlighted or marked
2. **Small dataset** - Only 16 mock log lines made navigation hard to see
3. **No clear indicators** - Status bar shows "Match 1/3" but viewport shows no visual distinction

**Root Cause**: Search navigates viewport correctly but provides minimal visual feedback.

---

## What We Fixed (Phase 2)

### 1. Increased Mock Log Dataset ✅

**Before**: 16 lines of simple app logs  
**After**: 50 lines of realistic Kubernetes kubelet logs

**Benefits**:
- Easier to see viewport navigation when searching
- More realistic testing scenarios
- Better representation of actual pod logs
- More varied log levels (INFO, WARN, ERROR)

**Example searches that now work better**:
- `ERROR` → 6 matches (was 1)
- `WARN` → 6 matches (was 1)  
- `connection refused` → Multiple matches across different operations
- `register node` → 3 retry attempts visible
- `volume` → Multiple volume-related operations

### 2. Realistic Log Format ✅

Now using actual Kubernetes log format:
```
I1127 00:44:40.476206 [INFO] Kubelet starting up...
E1127 00:44:40.479579 [ERROR] Skipping pod synchronization - PLEG is not healthy
W1127 00:44:40.483015 [WARN] Failed to list RuntimeClass: connection refused
```

This matches real `kubectl logs` output, making the UI feel more authentic.

---

## Current Search Behavior

### What Works

✅ **Functional Search**:
- Press '/' to enter search mode
- Type full query (all characters captured correctly)
- Press Enter to execute search
- Viewport navigates to first match
- Press 'n' / 'N' to navigate between matches
- Status bar shows "Match X/Y"

✅ **Filter Integration**:
- Search respects active filters (ERROR, WARN)
- Line count shows filtered count correctly
- Search indices match visible content

✅ **Escape Handling**:
- Esc cancels search without exiting logs view
- Clean state management on view exit

### What's Missing (Phase 3 Feature)

❌ **Visual Highlighting**:
- Matched text is not highlighted in the viewport
- Current match line has no visual distinction
- Hard to tell which line you're on when all look the same

❌ **Search Term Visibility**:
- No indication of which words matched
- No color coding or emphasis

---

## Testing the Current Implementation

### Good Search Terms to Try

1. **`ERROR`** - 6 matches spread throughout logs
   - Tests filter integration (should show all ERRORs)
   - Tests navigation through multiple matches

2. **`connection refused`** - Multiple matches
   - Tests phrase search
   - Shows viewport navigation clearly with 50 lines

3. **`volume`** - Volume-related operations
   - Tests specific keyword search
   - Multiple contexts (mount, attach, verify)

4. **`register node`** - Node registration retries
   - Tests repeated patterns
   - Good for testing 'n' navigation

5. **Combined: ERROR filter + search `node`**
   - Tests filter + search integration
   - Should only match ERROR lines containing "node"

### How to Verify Search is Working

Even without highlighting, you can verify search works by:

1. **Status bar changes**: Watch "Match 1/6" update as you press 'n'
2. **Scroll position**: Viewport scrolls to different positions
3. **Line numbers**: Count lines from top to verify position
4. **Known matches**: Search "ERROR" after applying ERROR filter (should navigate through all filtered lines)

---

## Phase 3: Visual Highlighting Plan

### Planned Improvements

#### 1. Highlight Current Match Line

Use lipgloss to style the entire line containing current match:

```go
if lineIndex == currentMatchIndex {
    styledLine = lipgloss.NewStyle().
        Background(lipgloss.Color("33")).  // Yellow background
        Foreground(lipgloss.Color("0")).   // Black text
        Render(line)
}
```

#### 2. Highlight Search Term Within Line

Use ANSI escape codes or lipgloss to highlight just the matched text:

```go
// Replace matched text with highlighted version
highlighted := lipgloss.NewStyle().
    Background(lipgloss.Color("11")).  // Bright yellow
    Foreground(lipgloss.Color("0")).   // Black text  
    Bold(true).
    Render(matchedText)

line = strings.Replace(line, matchedText, highlighted, 1)
```

#### 3. Current Match vs Other Matches

- **Current match**: Bright yellow background (like vim's search)
- **Other matches**: Subtle highlight (dim yellow or underline)
- **Dimmed non-matches**: Slightly gray to emphasize matches

#### 4. Search Status Enhancements

Instead of just "Match 1/6", show:
```
🔍 "ERROR" | Match 1/6 | n=next N=prev /=new Esc=clear
```

#### 5. Match Preview

Show a small preview of upcoming/previous matches:
```
Matches: [... ← ERROR in line 7 | *ERROR in line 13* | ERROR in line 23 → ...]
```

---

## Implementation Notes for Phase 3

### Challenge: Viewport Content Modification

The `viewport` component renders plain text. To add highlighting:

**Option 1**: Pre-process log content before setting viewport content
- Inject ANSI codes or lipgloss styles
- Update content on each search navigation
- **Pro**: Works with existing viewport
- **Con**: Need to regenerate styled content frequently

**Option 2**: Custom log renderer
- Build custom scrollable view
- Full control over line rendering
- **Pro**: More flexibility
- **Con**: More complex, lose viewport features

**Recommendation**: Use Option 1 with caching

### Implementation Sketch

```go
func (a *App) renderLogsWithHighlight() string {
    visibleLogs := a.getVisibleLogs()
    styledLines := make([]string, len(visibleLogs))
    
    for i, line := range visibleLogs {
        // Check if this line has a match
        isMatch := false
        isCurrentMatch := false
        for j, matchIdx := range a.searchMatches {
            if matchIdx == i {
                isMatch = true
                if j == a.currentMatch {
                    isCurrentMatch = true
                }
                break
            }
        }
        
        // Apply styling
        if isCurrentMatch {
            // Bright highlight for current match
            styledLines[i] = currentMatchStyle.Render(line)
        } else if isMatch {
            // Subtle highlight for other matches
            styledLines[i] = otherMatchStyle.Render(line)
        } else {
            // Normal line
            styledLines[i] = line
        }
    }
    
    return strings.Join(styledLines, "\n")
}
```

---

## Dependencies for Phase 3

Current stack already has everything needed:
- ✅ `lipgloss` - For styling/colors
- ✅ `viewport` - For scrolling
- ✅ Search state tracking - Already implemented

**No new dependencies required!**

---

## Testing Checklist for Phase 3 (When Implemented)

- [ ] Current match line has distinct background color
- [ ] Search term within line is highlighted
- [ ] Other matches have subtle highlight
- [ ] Highlighting updates when pressing 'n' / 'N'
- [ ] Highlighting works with filters active
- [ ] Highlighting clears when exiting search
- [ ] Performance is acceptable with highlighting (< 10ms)
- [ ] Colors are readable in different terminal themes
- [ ] Highlighting works with long lines (no truncation issues)

---

## User Feedback Addressed

**Original Issue**: "I was not sure if it was working as the mock logs are small and I couldn't see any highlighting"

**Resolution**:
1. ✅ **Mock logs expanded** - 16 → 50 lines for easier testing
2. ✅ **Realistic format** - Now uses actual Kubernetes log format
3. 📋 **Highlighting planned** - Scheduled for Phase 3 (ANSI color support)

**Current Status**: Search is fully functional but visual feedback is minimal. Improved mock logs make it easier to verify that search is working, but Phase 3 highlighting will provide the professional UX expected.

---

## Comparison: Before & After

### Before (16 lines)
```
2025-11-27T16:30:00Z [INFO] Application starting...
2025-11-27T16:30:01Z [INFO] Connecting to database at db:5432
... (14 more simple lines)
```

- Hard to see navigation
- Generic app logs
- Only 1 ERROR, 1 WARN

### After (50 lines)  
```
I1127 00:44:40.476206 [INFO] Kubelet starting up...
E1127 00:44:40.479579 [ERROR] Skipping pod synchronization - PLEG is not healthy
W1127 00:44:40.483015 [WARN] Failed to list RuntimeClass: connection refused
... (47 more realistic Kubernetes logs)
```

- Clear viewport navigation visible
- Realistic Kubernetes kubelet logs
- 6 ERRORs, 6 WARNs - better testing

---

## Recommendations

### For Current Testing (Phase 2)

1. **Use specific searches**: Try "connection refused", "volume", "ERROR"
2. **Watch status bar**: Match count updates as you navigate
3. **Use filters first**: Apply ERROR/WARN filter, then search
4. **Count lines manually**: Verify viewport position changes

### For Phase 3 Development

1. **Implement line highlighting first** - Biggest UX win
2. **Then add term highlighting** - More complex but valuable
3. **Test with various terminal themes** - Ensure readability
4. **Add color scheme options** - Light/dark mode support

---

## Conclusion

✅ **Phase 2 Complete**: Search is fully functional with improved mock data  
📋 **Phase 3 Planned**: Visual highlighting for professional UX  
🎯 **User Need Met**: Larger dataset makes search behavior visible and testable

**Current state is production-ready for functionality, Phase 3 will add polish.**
