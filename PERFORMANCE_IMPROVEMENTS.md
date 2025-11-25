# Performance Improvements Backlog

This document tracks known performance issues and potential optimizations for future implementation.

## Known Performance Issues

### 1. Slow ESC Navigation from CRD Instances to CRD List
**Status:** Identified - Not Implemented  
**Priority:** Medium  
**Component:** CRD Views - Navigation

**Issue:**
When navigating back from the CRD instances view to the CRD list view using ESC, there is a noticeable delay.

**Current Behavior:**
- User presses ESC in CRD instances view
- App pops view from stack and refreshes CRD list
- During refresh, `updateTable()` is called which triggers `getCRDInstanceCount()` for each CRD
- `getCRDInstanceCount()` makes an API call to fetch instances for each CRD to display the count
- With many CRDs, this results in many sequential API calls

**Root Cause:**
- The `getCRDInstanceCount()` method is called synchronously for each CRD during table rendering
- Each call makes a live API request to Rancher
- With N CRDs, this results in N API calls happening during the table update
- API calls are blocking the UI render

**Potential Solutions:**

1. **Caching Strategy** (Recommended)
   - Cache instance counts after first fetch
   - Invalidate cache on refresh (Ctrl+R) or after a time period
   - Pros: Simple to implement, significant performance gain
   - Cons: Counts may be slightly stale

2. **Async/Background Loading**
   - Display CRDs immediately with loading indicators for counts
   - Fetch counts asynchronously in the background
   - Update table rows as counts become available
   - Pros: Instant navigation, progressive loading
   - Cons: More complex UI state management

3. **Batch API Calls**
   - If Rancher API supports it, batch multiple resource queries
   - Single API call instead of N calls
   - Pros: Optimal network usage
   - Cons: Requires API support, may not be available

4. **Lazy Loading**
   - Only fetch instance counts when user hovers/selects a CRD
   - Show counts on-demand rather than for all CRDs
   - Pros: Minimal API calls
   - Cons: Counts not immediately visible

**Recommended Implementation:**
- **Phase 1:** Implement simple caching (Solution 1)
  - Cache counts in memory with TTL of 30 seconds
  - Clear cache on explicit refresh (Ctrl+R)
  - This should provide immediate improvement

- **Phase 2:** Consider async loading (Solution 2)
  - If caching is insufficient
  - Provides better UX with progressive loading

**Estimated Effort:** 2-4 hours for caching implementation

**Related Code:**
- File: `internal/tui/app.go`
- Method: `getCRDInstanceCount()`
- Method: `updateTable()` - CRD case

**Reported:** 2025-11-25
**Reported By:** User testing

---

## Future Performance Considerations

### General API Call Optimization
- Consider implementing a request queue/throttling mechanism
- Implement connection pooling if not already done
- Add request timeout handling
- Consider parallel API calls where safe to do so

### UI Rendering
- Profile table rendering performance with large datasets
- Consider virtualization for very long lists
- Optimize re-renders to only update changed data
