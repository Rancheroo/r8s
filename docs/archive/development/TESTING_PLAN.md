# Testing Plan for Recent r9s Changes

## Overview
This document outlines the testing plan for recent feature additions:
1. Offline Mode Warning Banner (commit d7ebe2f)
2. Deployments and Services Views with Keyboard Navigation (commit 319a2ae)

## Test Environment Setup

### Prerequisites
- Go 1.21+ installed
- Build the latest version: `make build`
- Configuration file at `~/.r9s/config.yaml` with a valid profile

### Test Scenarios

## 1. Offline Mode Warning Banner Tests

### Test 1.1: Offline Mode Detection
**Objective:** Verify the banner appears when Rancher API is unreachable

**Steps:**
1. Configure invalid Rancher URL in `~/.r9s/config.yaml`:
   ```yaml
   profiles:
     - name: offline-test
       url: https://invalid-rancher-url.example.com
       token: fake-token
   current_profile: offline-test
   ```
2. Run `./bin/r9s`
3. Observe the application startup

**Expected Results:**
- Application loads successfully (doesn't crash)
- Red warning banner appears: `⚠️  OFFLINE MODE - DISPLAYING MOCK DATA  ⚠️`
- Banner is positioned between breadcrumb and table
- Banner has red background, white text, and blinks
- Mock data is displayed in all views

**Pass Criteria:**
- ✓ Banner visible and blinking
- ✓ Mock data loads correctly
- ✓ No error crashes

### Test 1.2: Online Mode (No Banner)
**Objective:** Verify the banner does NOT appear with live Rancher connection

**Steps:**
1. Configure valid Rancher URL and token in `~/.r9s/config.yaml`
2. Run `./bin/r9s`
3. Observe the application startup

**Expected Results:**
- No warning banner appears
- Real cluster data is displayed
- Normal navigation works

**Pass Criteria:**
- ✓ No offline banner visible
- ✓ Live data loads from Rancher API

### Test 1.3: Banner Visibility Across Views
**Objective:** Verify banner appears in all views when offline

**Steps:**
1. Start r9s in offline mode (invalid config)
2. Navigate through: Clusters → Projects → Namespaces → Pods
3. Press '2' to switch to Deployments
4. Press '3' to switch to Services
5. Press 'Esc' to go back to Namespaces
6. Navigate to different namespace

**Expected Results:**
- Banner visible in Clusters view
- Banner visible in Projects view
- Banner visible in Namespaces view
- Banner visible in Pods view
- Banner visible in Deployments view
- Banner visible in Services view
- Banner persists across all navigation

**Pass Criteria:**
- ✓ Banner consistently visible across all views
- ✓ Banner position remains consistent

## 2. Deployments View Tests

### Test 2.1: Access Deployments View
**Objective:** Navigate to Deployments view and verify display

**Steps:**
1. Start r9s (offline or online mode)
2. Navigate: Clusters → Projects → Namespaces → Select a namespace (Enter)
3. Press '2' to switch to Deployments view

**Expected Results:**
- View switches to Deployments
- Breadcrumb shows: `Cluster: X > Project: Y > Namespace: Z > Deployments`
- Table shows columns: NAME, NAMESPACE, READY, UP-TO-DATE, AVAILABLE
- Status bar shows: `X deployments | Press '1'=Pods '2'=Deployments '3'=Services | '?' for help | 'q' to quit`

**Pass Criteria:**
- ✓ Deployments table displays correctly
- ✓ Breadcrumb updates properly
- ✓ Column headers are correct
- ✓ Status bar shows deployment count

### Test 2.2: Deployments Data Display
**Objective:** Verify deployment data is formatted correctly

**Steps:**
1. Navigate to Deployments view (offline mode for predictable data)
2. Observe the mock deployment data

**Expected Results:**
For offline mode, should see deployments like:
- `nginx-deployment`: READY 3/3, UP-TO-DATE 3, AVAILABLE 3
- `redis-deployment`: READY 2/2, UP-TO-DATE 2, AVAILABLE 2
- `api-server`: READY 5/5, UP-TO-DATE 5, AVAILABLE 5
- `worker-deployment` (updating state): READY 3/4, UP-TO-DATE 1, AVAILABLE 3

**Pass Criteria:**
- ✓ All deployments display with correct replica counts
- ✓ READY column shows format: X/Y
- ✓ UP-TO-DATE and AVAILABLE show numbers
- ✓ Namespace names extracted correctly

### Test 2.3: Deployments Refresh
**Objective:** Test refresh functionality in Deployments view

**Steps:**
1. Navigate to Deployments view
2. Press 'r' or 'Ctrl+R' to refresh

**Expected Results:**
- View refreshes
- Loading indicator appears briefly
- Data reloads (same data in offline mode)

**Pass Criteria:**
- ✓ Refresh triggers without errors
- ✓ Data reloads successfully

## 3. Services View Tests

### Test 3.1: Access Services View
**Objective:** Navigate to Services view and verify display

**Steps:**
1. Start r9s
2. Navigate: Clusters → Projects → Namespaces → Select a namespace
3. Press '3' to switch to Services view

**Expected Results:**
- View switches to Services
- Breadcrumb shows: `Cluster: X > Project: Y > Namespace: Z > Services`
- Table shows columns: NAME, NAMESPACE, TYPE, CLUSTER-IP, PORT(S)
- Status bar shows: `X services | Press '1'=Pods '2'=Deployments '3'=Services | '?' for help | 'q' to quit`

**Pass Criteria:**
- ✓ Services table displays correctly
- ✓ Breadcrumb updates properly
- ✓ Column headers are correct
- ✓ Status bar shows service count

### Test 3.2: Services Data Display
**Objective:** Verify service data formatting, especially ports

**Steps:**
1. Navigate to Services view (offline mode)
2. Observe the mock service data

**Expected Results:**
For offline mode, should see services like:
- `nginx-service`: TYPE ClusterIP, CLUSTER-IP 10.43.100.50, PORT(S) 80/TCP
- `redis-service`: TYPE ClusterIP, CLUSTER-IP 10.43.100.51, PORT(S) 6379/TCP
- `api-service`: TYPE NodePort, PORT(S) 8080:30080/TCP (shows NodePort)
- `loadbalancer-service`: TYPE LoadBalancer, PORT(S) 80/TCP,443/TCP (multiple ports)

**Pass Criteria:**
- ✓ Service types display correctly (ClusterIP, NodePort, LoadBalancer)
- ✓ Cluster IPs show correctly
- ✓ Port formats are correct:
  - Simple: `80/TCP`
  - NodePort: `8080:30080/TCP`
  - Multiple: `80/TCP,443/TCP`

### Test 3.3: Services Refresh
**Objective:** Test refresh in Services view

**Steps:**
1. Navigate to Services view
2. Press 'r' to refresh

**Expected Results:**
- Services data refreshes
- No errors occur

**Pass Criteria:**
- ✓ Refresh completes successfully

## 4. Keyboard Navigation Tests

### Test 4.1: Switch Between Resource Views (1/2/3 Keys)
**Objective:** Test keyboard shortcuts to switch between Pods/Deployments/Services

**Steps:**
1. Navigate to a namespace and enter it (will show Pods by default)
2. Press '2' → Should switch to Deployments
3. Press '3' → Should switch to Services
4. Press '1' → Should switch back to Pods
5. Repeat the sequence multiple times

**Expected Results:**
- Each keypress switches to the appropriate view instantly
- Loading indicator shows briefly
- Breadcrumb updates to show current resource type
- Status bar updates to show correct count
- View context (cluster/project/namespace) remains the same

**Pass Criteria:**
- ✓ All three keys (1/2/3) work correctly
- ✓ View switches are instant and smooth
- ✓ Context is preserved across switches
- ✓ No errors or crashes

### Test 4.2: Navigation Keys Only Work in Namespace Resource Views
**Objective:** Verify 1/2/3 keys only work when in namespace resource views

**Steps:**
1. In Clusters view, press '1', '2', '3'
2. In Projects view, press '1', '2', '3'
3. In Namespaces view, press '1', '2', '3'
4. In Pods view, press '1', '2', '3' → Should work
5. Navigate back with Esc to Namespaces

**Expected Results:**
- Keys 1/2/3 do nothing in Clusters view
- Keys 1/2/3 do nothing in Projects view
- Keys 1/2/3 do nothing in Namespaces view
- Keys 1/2/3 work in Pods/Deployments/Services views

**Pass Criteria:**
- ✓ Keys only active in correct views
- ✓ No unintended behavior in other views

### Test 4.3: Switching with Different Namespaces
**Objective:** Test view switching works across different namespaces

**Steps:**
1. Navigate to namespace "default", press Enter
2. Press '2' for Deployments
3. Press Esc to go back to Namespaces
4. Navigate to namespace "monitoring" (or another), press Enter
5. Press '3' for Services
6. Verify data is filtered by namespace

**Expected Results:**
- Each namespace shows only its own resources
- Switching views (1/2/3) filters data correctly per namespace
- No data leakage between namespaces

**Pass Criteria:**
- ✓ Namespace filtering works correctly
- ✓ Each view shows only resources from current namespace

## 5. Integration Tests

### Test 5.1: Offline Mode + Keyboard Navigation
**Objective:** Test keyboard navigation with offline mode banner

**Steps:**
1. Start in offline mode (invalid config)
2. Navigate through all views using keyboard shortcuts
3. Verify offline banner appears in Pods/Deployments/Services views

**Expected Results:**
- Offline banner visible in all views
- Keyboard navigation (1/2/3) works correctly
- Mock data displays correctly in all resource views

**Pass Criteria:**
- ✓ Banner + navigation work together seamlessly
- ✓ No layout issues with banner present

### Test 5.2: Describe Feature with New Views
**Objective:** Ensure describe ('d' key) works appropriately in new views

**Steps:**
1. Navigate to Deployments view
2. Select a deployment, press 'd'
3. Navigate to Services view
4. Select a service, press 'd'
5. Navigate to Pods view
6. Select a pod, press 'd'

**Expected Results:**
- Deployments: Shows error "Describe is not yet implemented for this resource type"
- Services: Shows error "Describe is not yet implemented for this resource type"
- Pods: Shows pod details in JSON format

**Pass Criteria:**
- ✓ Describe only works for Pods (as designed)
- ✓ Appropriate error messages for other resources
- ✓ No crashes

### Test 5.3: Esc Navigation with New Views
**Objective:** Test Esc key navigation from new views

**Steps:**
1. Navigate: Clusters → Projects → Namespaces → Namespace (Pods view)
2. Press '2' for Deployments
3. Press Esc → Should go back to Namespaces view
4. Enter namespace again
5. Press '3' for Services
6. Press Esc → Should go back to Namespaces view

**Expected Results:**
- Esc from Deployments returns to Namespaces view
- Esc from Services returns to Namespaces view
- Navigation stack works correctly

**Pass Criteria:**
- ✓ Esc navigation works from all new views
- ✓ Returns to correct parent view
- ✓ Navigation stack is correct

## 6. Loading State Tests

### Test 6.1: Loading Indicator in Offline Mode
**Objective:** Verify loading states display correctly in offline mode

**Steps:**
1. Start in offline mode
2. Observe initial loading
3. Navigate through views, observing loading states

**Expected Results:**
- Loading message shows: "Loading mock data (OFFLINE MODE)..."
- Loading states are brief (mock data loads quickly)
- Offline banner appears after loading completes

**Pass Criteria:**
- ✓ Offline-specific loading message appears
- ✓ Loading completes successfully
- ✓ No infinite loading states

### Test 6.2: Refresh Loading States
**Objective:** Verify refresh loading works with new views

**Steps:**
1. In Deployments view, press 'r'
2. Observe loading state
3. In Services view, press 'r'
4. Observe loading state

**Expected Results:**
- Loading indicator appears
- View refreshes successfully
- Data reloads

**Pass Criteria:**
- ✓ Refresh loading works in new views
- ✓ No stuck loading states

## 7. Error Handling Tests

### Test 7.1: Graceful Fallback to Mock Data
**Objective:** Test graceful degradation when API fails

**Steps:**
1. Start with valid config
2. Simulate connection failure (disconnect network or change to invalid URL)
3. Navigate through views
4. Refresh views with 'r'

**Expected Results:**
- App doesn't crash
- Automatically falls back to mock data
- Offline banner appears
- All views still navigable

**Pass Criteria:**
- ✓ No crashes on connection failure
- ✓ Graceful fallback to offline mode
- ✓ User can continue using app with mock data

## 8. Visual/UI Tests

### Test 8.1: Banner Styling
**Objective:** Verify offline banner styling is correct

**Steps:**
1. Start in offline mode
2. Observe banner appearance

**Expected Results:**
- Red background (easily distinguishable)
- White text (high contrast)
- Bold text (readable)
- Blinking effect (attention-grabbing)
- Centered text
- Warning emoji (⚠️) visible
- Full-width banner

**Pass Criteria:**
- ✓ All styling elements present
- ✓ Banner stands out visually
- ✓ Text is readable

### Test 8.2: Table Layout with Banner
**Objective:** Ensure table layout isn't broken by banner

**Steps:**
1. Start in offline mode
2. Navigate to various views
3. Verify table renders correctly below banner

**Expected Results:**
- Table positioned correctly below banner
- No overlap between banner and table
- Table columns aligned properly
- Status bar at bottom of screen

**Pass Criteria:**
- ✓ Clean layout with banner present
- ✓ No visual artifacts or overlap

## 9. Performance Tests

### Test 9.1: View Switching Performance
**Objective:** Verify quick response when switching views with 1/2/3

**Steps:**
1. Navigate to Pods view
2. Rapidly press 2, 3, 1, 2, 3, 1 (switch views quickly)
3. Observe response time

**Expected Results:**
- Each view switch is near-instant
- No lag or delay
- No rendering issues
- Memory doesn't leak

**Pass Criteria:**
- ✓ View switches < 100ms
- ✓ Smooth user experience
- ✓ No memory issues

## 10. Regression Tests

### Test 10.1: Existing Features Still Work
**Objective:** Verify no regressions in existing functionality

**Steps:**
Test the following existing features:
1. CRD navigation (Select cluster, press 'C')
2. CRD instance browsing
3. CRD description toggle ('i' key)
4. Help screen ('?')
5. Quit ('q')
6. General Esc navigation

**Expected Results:**
- All existing features work as before
- No new bugs introduced

**Pass Criteria:**
- ✓ CRD navigation works
- ✓ All keyboard shortcuts work
- ✓ Help screen displays
- ✓ Quit works cleanly

## Test Execution Checklist

### Pre-Test Setup
- [ ] Build latest version: `make build`
- [ ] Backup config: `cp ~/.r9s/config.yaml ~/.r9s/config.yaml.backup`
- [ ] Create offline test config
- [ ] Create online test config (if available)

### Test Execution
- [ ] Section 1: Offline Mode Warning Banner Tests (1.1-1.3)
- [ ] Section 2: Deployments View Tests (2.1-2.3)
- [ ] Section 3: Services View Tests (3.1-3.3)
- [ ] Section 4: Keyboard Navigation Tests (4.1-4.3)
- [ ] Section 5: Integration Tests (5.1-5.3)
- [ ] Section 6: Loading State Tests (6.1-6.2)
- [ ] Section 7: Error Handling Tests (7.1)
- [ ] Section 8: Visual/UI Tests (8.1-8.2)
- [ ] Section 9: Performance Tests (9.1)
- [ ] Section 10: Regression Tests (10.1)

### Post-Test
- [ ] Document any issues found
- [ ] Restore config: `cp ~/.r9s/config.yaml.backup ~/.r9s/config.yaml`
- [ ] File bug reports if needed
- [ ] Update STATUS.md with test results

## Bug Report Template

When issues are found, use this template:

```
**Test Case:** [Test number and name]
**Steps to Reproduce:**
1. 
2. 
3. 

**Expected Result:**
[What should happen]

**Actual Result:**
[What actually happened]

**Screenshots/Logs:**
[If applicable]

**Environment:**
- OS: 
- r9s version: 
- Mode: Online/Offline
```

## Success Criteria

All tests must pass with:
- ✓ No crashes or panics
- ✓ Correct data display
- ✓ Proper navigation
- ✓ Expected visual appearance
- ✓ No regressions

Any failing tests should be documented and prioritized for fixes.
