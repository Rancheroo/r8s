# Phase 2 Steps 3-5: Advanced Log Viewing Test Plan

**Date**: 2025-11-27  
**Features**: Container Selection, Tail Mode, Log Level Filtering  
**Build Status**: ✅ Compiles successfully

---

## Overview

This test plan validates the implementation of Steps 3, 4, and 5 of Phase 2 (Pager Integration), which adds advanced log viewing capabilities to r8s.

### Implemented Features

1. **Step 3: Container Selection** - 'c' key to cycle through containers in multi-container pods
2. **Step 4: Tail Mode** - 't' key to toggle auto-refresh tail mode
3. **Step 5: Log Level Filtering** - Ctrl+E/W/A to filter logs by severity

---

## Test Scenarios

### Test 1: Container Selection Feature

**Objective**: Verify container cycling works correctly

**Prerequisites**:
- r8s binary built successfully
- In offline mode (uses mock data)

**Steps**:
1. Run `./r8s`
2. Navigate to: Clusters > Projects > Namespaces > Pods
3. Select any pod and press 'l' to view logs
4. Press 'c' to cycle containers
5. Observe status bar changes
6. Press 'c' again (should cycle to next container)
7. Press 'c' third time (should wrap to first container)

**Expected Results**:
- ✅ First 'c' press: Status shows "Container: app"
- ✅ Second 'c' press: Status shows "Container: sidecar"
- ✅ Third 'c' press: Status shows "Container: init"
- ✅ Fourth 'c' press: Wraps back to "Container: app"
- ✅ No crashes or errors

**Actual Results**: *(To be filled during testing)*

**Status**: ⬜ Not Tested | ✅ Pass | ❌ Fail

---

### Test 2: Tail Mode Toggle

**Objective**: Verify tail mode can be toggled on/off

**Prerequisites**:
- r8s running with logs view open

**Steps**:
1. From logs view, press 't' to enable tail mode
2. Observe status bar shows "TAIL MODE"
3. Observe viewport scrolls to bottom
4. Press 't' again to disable tail mode
5. Observe "TAIL MODE" indicator disappears

**Expected Results**:
- ✅ 't' key toggles tail mode on
- ✅ Status bar shows "TAIL MODE" indicator when active
- ✅ Viewport auto-scrolls to bottom when enabled
- ✅ Second 't' press disables tail mode
- ✅ Indicator removed from status bar

**Actual Results**: *(To be filled during testing)*

**Status**: ⬜ Not Tested | ✅ Pass | ❌ Fail

---

### Test 3: Error-Only Filtering (Ctrl+E)

**Objective**: Filter logs to show only ERROR level entries

**Prerequisites**:
- r8s running with logs view open
- Mock logs contain [ERROR] entries

**Steps**:
1. From logs view, verify you can see multiple log levels
2. Press Ctrl+E to filter to errors only
3. Observe only lines with [ERROR] are shown
4. Count visible lines (should be 1: line 13)
5. Press Ctrl+E again to toggle off
6. Observe all logs return

**Expected Results**:
- ✅ Ctrl+E filters to show only [ERROR] lines
- ✅ Status bar shows "Filter: ERROR"
- ✅ Only 1 line visible (the ERROR line from mock data)
- ✅ Second Ctrl+E press clears filter
- ✅ All 16 log lines visible again
- ✅ Filter indicator removed from status

**Actual Results**: *(To be filled during testing)*

**Status**: ⬜ Not Tested | ✅ Pass | ❌ Fail

---

### Test 4: Warning/Error Filtering (Ctrl+W)

**Objective**: Filter logs to show WARN and ERROR level entries

**Prerequisites**:
- r8s running with logs view open

**Steps**:
1. From logs view, press Ctrl+W
2. Observe status bar shows "Filter: WARN"
3. Count visible lines (should show WARN + ERROR = 2 lines)
4. Verify line 10 ([WARN]) is visible
5. Verify line 13 ([ERROR]) is visible
6. Press Ctrl+W again to toggle off

**Expected Results**:
- ✅ Ctrl+W filters to [WARN] and [ERROR] lines
- ✅ Status bar shows "Filter: WARN"
- ✅ 2 lines visible (WARN line + ERROR line)
- ✅ Both warning and error content displayed
- ✅ Second Ctrl+W press clears filter

**Actual Results**: *(To be filled during testing)*

**Status**: ⬜ Not Tested | ✅ Pass | ❌ Fail

---

### Test 5: Clear All Filters (Ctrl+A)

**Objective**: Verify Ctrl+A clears any active filter

**Prerequisites**:
- r8s running with logs view open
- Active filter (ERROR or WARN)

**Steps**:
1. Press Ctrl+E to enable error filter
2. Verify only error lines shown
3. Press Ctrl+A to clear filter
4. Observe all logs return
5. Repeat with Ctrl+W filter
6. Press Ctrl+A again

**Expected Results**:
- ✅ Ctrl+A clears ERROR filter
- ✅ All 16 log lines visible
- ✅ No filter indicator in status
- ✅ Ctrl+A also clears WARN filter
- ✅ Works from any filter state

**Actual Results**: *(To be filled during testing)*

**Status**: ⬜ Not Tested | ✅ Pass | ❌ Fail

---

### Test 6: Combined Features - Multi-State Test

**Objective**: Verify all features work together correctly

**Prerequisites**:
- r8s running with logs view open

**Steps**:
1. Press 'c' to cycle containers → Status shows "Container: app"
2. Press 't' to enable tail mode → Status shows "TAIL MODE"
3. Press Ctrl+E for error filter → Status shows "Filter: ERROR"
4. Verify status bar shows all 3 indicators: "Container: app | TAIL MODE | Filter: ERROR"
5. Press '/' to search for "INFO"
6. Verify search works with filter active
7. Press 'n' to jump to next match (should work even with filter)
8. Press Esc to exit search
9. Press Ctrl+A to clear filter
10. Press 't' to disable tail
11. Verify status returns to basics

**Expected Results**:
- ✅ Multiple features can be active simultaneously
- ✅ Status bar correctly shows all active features
- ✅ Search works with filters active
- ✅ Features don't interfere with each other
- ✅ Each feature can be toggled independently
- ✅ Status bar updates correctly as features change

**Actual Results**: *(To be filled during testing)*

**Status**: ⬜ Not Tested | ✅ Pass | ❌ Fail

---

### Test 7: Filter with No Matches

**Objective**: Verify graceful handling when filter has no matches

**Prerequisites**:
- r8s running with custom logs that have no [ERROR] tags

**Steps**:
1. Navigate to logs view
2. Press Ctrl+E to filter errors
3. Observe message "No logs match the current filter"
4. Press Ctrl+A to clear

**Expected Results**:
- ✅ Displays helpful message instead of blank screen
- ✅ No crash or error
- ✅ Can exit filter mode cleanly

**Actual Results**: *(To be filled during testing)*

**Status**: ⬜ Not Tested | ✅ Pass | ❌ Fail

---

### Test 8: Status Bar Dynamic Updates

**Objective**: Verify status bar correctly reflects all active states

**Prerequisites**:
- r8s running

**Steps**:
1. View logs (default state)
2. Note status: "16 lines | 't'=tail ..."
3. Press 't' → Note status includes "TAIL MODE"
4. Press 'c' → Note status includes "Container: app"
5. Press Ctrl+E → Note status includes "Filter: ERROR"
6. Verify format: "{count} lines | TAIL MODE | Filter: ERROR | Container: app | ..."
7. Toggle each off and verify status updates

**Expected Results**:
- ✅ Line count always shown first
- ✅ Each active feature shown with pipe separator
- ✅ Status updates immediately when toggling
- ✅ Clean format, readable
- ✅ Help commands shown at end

**Actual Results**: *(To be filled during testing)*

**Status**: ⬜ Not Tested | ✅ Pass | ❌ Fail

---

## Performance Tests

### Test 9: Filter Performance

**Objective**: Ensure filtering is fast even with many log lines

**Prerequisites**:
- Logs view open (16 lines in mock data)

**Steps**:
1. Press Ctrl+E to filter
2. Measure response time (should be instant)
3. Press Ctrl+W
4. Measure response time
5. Press Ctrl+A
6. Measure response time

**Expected Results**:
- ✅ All filter operations < 100ms
- ✅ No visible lag
- ✅ Smooth user experience

**Actual Results**: *(To be filled during testing)*

**Status**: ⬜ Not Tested | ✅ Pass | ❌ Fail

---

## Edge Cases

### Test 10: Empty Logs

**Objective**: Handle edge case of pod with no logs

**Prerequisites**:
- Modified to have empty log array

**Steps**:
1. View logs for pod with no output
2. Try all features: container cycling, tail, filtering
3. Verify no crashes

**Expected Results**:
- ✅ Displays "0 lines" in status
- ✅ Features can be toggled without error
- ✅ Graceful handling of empty state

**Actual Results**: *(To be filled during testing)*

**Status**: ⬜ Not Tested | ✅ Pass | ❌ Fail

---

### Test 11: Single Container Pod

**Objective**: Verify container cycling gracefully handles single-container pods

**Prerequisites**:
- Default pod (mock has empty containers list initially)

**Steps**:
1. View logs
2. Press 'c' to cycle containers
3. Verify behavior with only 1 container

**Expected Results**:
- ✅ Container list initialized with defaults
- ✅ Cycling works (shows app→sidecar→init)
- ✅ No crash or error

**Actual Results**: *(To be filled during testing)*

**Status**: ⬜ Not Tested | ✅ Pass | ❌ Fail

---

## Integration Tests

### Test 12: Search + Filter Integration

**Objective**: Verify search works correctly with active filters

**Prerequisites**:
- Logs view open

**Steps**:
1. Press Ctrl+W to show WARN/ERROR only (2 lines)
2. Press '/' to enter search mode
3. Search for "Slow" (in WARN line)
4. Press Enter
5. Verify match found
6. Press 'n' to try next match
7. Clear filter with Ctrl+A
8. Search again for "INFO"

**Expected Results**:
- ✅ Search operates on filtered view
- ✅ Match count reflects filtered lines only
- ✅ Navigation works correctly
- ✅ Clearing filter allows searching full logs
- ✅ No confusion between search/filter

**Actual Results**: *(To be filled during testing)*

**Status**: ⬜ Not Tested | ✅ Pass | ❌ Fail

---

## Regression Tests

### Test 13: Existing Features Still Work

**Objective**: Ensure new features didn't break existing functionality

**Prerequisites**:
- r8s running

**Steps**:
1. Test basic navigation (Enter on resources)
2. Test 'd' for describe
3. Test Esc to go back
4. Test '?' for help
5. Test 'q' to quit (don't actually quit)
6. Test arrow keys for scrolling
7. Test existing search ('/')

**Expected Results**:
- ✅ All previous features work unchanged
- ✅ No regressions introduced

**Actual Results**: *(To be filled during testing)*

**Status**: ⬜ Not Tested | ✅ Pass | ❌ Fail

---

## Summary Template

### Test Execution Summary

**Date**: _____________  
**Tester**: _____________  
**Duration**: _____ minutes  

**Results**:
- Total Tests: 13
- ✅ Passed: ____
- ❌ Failed: ____  
- ⬜ Skipped: ____

**Critical Issues Found**: *(List any blockers)*

**Minor Issues Found**: *(List any cosmetic/UX issues)*

**Notes**:

---

## Success Criteria

Phase 2 Steps 3-5 are considered complete when:

- [ ] All 13 tests pass
- [ ] No critical bugs found
- [ ] Build compiles without errors
- [ ] Status bar correctly shows all active features
- [ ] Features work independently and together
- [ ] No regressions in existing functionality
- [ ] Performance is acceptable (< 100ms for operations)

---

## Known Limitations (By Design)

1. **Tail Mode Auto-Refresh**: Currently stubs - would need real API integration for live updates
2. **Container Detection**: Uses mock data - real implementation would parse actual pod spec
3. **Log Levels**: Simple string matching for [ERROR]/[WARN] - could be enhanced with regex
4. **Filter Persistence**: Filters reset when leaving logs view (by design for simplicity)

---

## Next Steps After Testing

Once all tests pass:

1. Document any findings in test report
2. Create issues for any bugs found
3. Plan Phase 3: Log Highlighting & Filtering enhancements
4. Consider adding ANSI color support for log levels
5. Evaluate need for regex-based filtering
