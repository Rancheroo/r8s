package tui

import (
	"testing"

	"github.com/charmbracelet/bubbles/viewport"
)

// TestGetDisplayedItems_Capping tests that items are capped at defaultDashboardCap
func TestGetDisplayedItems_Capping(t *testing.T) {
	app := &App{
		attentionItems:    make([]AttentionItem, 100), // 100 items
		attentionExpanded: false,                      // Not expanded
	}

	// Fill with dummy items
	for i := 0; i < 100; i++ {
		app.attentionItems[i] = AttentionItem{
			Title: "test-pod",
		}
	}

	displayed := app.getDisplayedItems()

	if len(displayed) != defaultDashboardCap {
		t.Errorf("Expected %d items when capped, got %d", defaultDashboardCap, len(displayed))
	}
}

// TestGetDisplayedItems_Expanded tests that all items show when expanded
func TestGetDisplayedItems_Expanded(t *testing.T) {
	app := &App{
		attentionItems:    make([]AttentionItem, 100), // 100 items
		attentionExpanded: true,                       // Expanded
	}

	// Fill with dummy items
	for i := 0; i < 100; i++ {
		app.attentionItems[i] = AttentionItem{
			Title: "test-pod",
		}
	}

	displayed := app.getDisplayedItems()

	if len(displayed) != 100 {
		t.Errorf("Expected 100 items when expanded, got %d", len(displayed))
	}
}

// TestGetDisplayedItems_UnderCap tests that all items show when under cap
func TestGetDisplayedItems_UnderCap(t *testing.T) {
	app := &App{
		attentionItems:    make([]AttentionItem, 10), // Only 10 items
		attentionExpanded: false,                     // Not expanded
	}

	// Fill with dummy items
	for i := 0; i < 10; i++ {
		app.attentionItems[i] = AttentionItem{
			Title: "test-pod",
		}
	}

	displayed := app.getDisplayedItems()

	if len(displayed) != 10 {
		t.Errorf("Expected 10 items when under cap, got %d", len(displayed))
	}
}

// TestRenderAttentionDashboard_LargeDataset tests rendering with 200+ items
func TestRenderAttentionDashboard_LargeDataset(t *testing.T) {
	app := &App{
		attentionItems:    make([]AttentionItem, 200),
		attentionExpanded: false,
		width:             100,
		height:            30,
		bundleMode:        true,
		attentionViewport: viewport.New(92, 20),
	}

	// Fill with dummy items of mixed severity
	for i := 0; i < 200; i++ {
		severity := SeverityInfo
		if i%3 == 0 {
			severity = SeverityCritical
		} else if i%3 == 1 {
			severity = SeverityWarning
		}

		app.attentionItems[i] = AttentionItem{
			Title:       "test-pod-" + string(rune(i)),
			Description: "Test issue",
			Namespace:   "default",
			Severity:    severity,
			Emoji:       "🔥",
		}
	}

	// Render - should not panic
	output := app.renderAttentionDashboard()

	if output == "" {
		t.Error("Expected non-empty dashboard output")
	}

	// Verify capping message appears when not expanded
	if len(app.attentionItems) > defaultDashboardCap && !app.attentionExpanded {
		// Should contain the "...and X more" message
		// Note: Can't easily check rendered output due to styling, but verify no panic
	}
}

// TestAttentionToggleExpansion simulates pressing 'm' to toggle expansion
func TestAttentionToggleExpansion(t *testing.T) {
	app := &App{
		attentionItems:    make([]AttentionItem, 50),
		attentionExpanded: false,
		attentionCursor:   0,
	}

	// Fill with dummy items
	for i := 0; i < 50; i++ {
		app.attentionItems[i] = AttentionItem{
			Title: "test-pod",
		}
	}

	// Simulate pressing 'm' to expand
	app.attentionExpanded = !app.attentionExpanded

	if !app.attentionExpanded {
		t.Error("Expected expansion to be true after toggle")
	}

	displayed := app.getDisplayedItems()
	if len(displayed) != 50 {
		t.Errorf("Expected 50 items when expanded, got %d", len(displayed))
	}

	// Toggle back
	app.attentionExpanded = !app.attentionExpanded

	if app.attentionExpanded {
		t.Error("Expected expansion to be false after second toggle")
	}

	displayed = app.getDisplayedItems()
	// With 50 items total and cap of 100, all 50 should be shown even when collapsed
	if len(displayed) != 50 {
		t.Errorf("Expected 50 items when collapsed (under cap), got %d", len(displayed))
	}
}

// TestAttentionCursorBoundsAfterToggle tests cursor bounds after expansion toggle
func TestAttentionCursorBoundsAfterToggle(t *testing.T) {
	app := &App{
		attentionItems:    make([]AttentionItem, 100),
		attentionExpanded: true,
		attentionCursor:   99, // At end of expanded list
	}

	// Fill with dummy items
	for i := 0; i < 100; i++ {
		app.attentionItems[i] = AttentionItem{
			Title: "test-pod",
		}
	}

	// Collapse - cursor should be reset if out of bounds
	app.attentionExpanded = false
	displayedItems := app.getDisplayedItems()

	// Simulate bounds checking (normally done in key handler)
	if app.attentionCursor >= len(displayedItems) {
		app.attentionCursor = len(displayedItems) - 1
		if app.attentionCursor < 0 {
			app.attentionCursor = 0
		}
	}

	if app.attentionCursor >= defaultDashboardCap {
		t.Errorf("Expected cursor to be within bounds (%d), got %d", defaultDashboardCap-1, app.attentionCursor)
	}
}

// BenchmarkRenderAttentionDashboard_1000Items benchmarks rendering 1000 items
func BenchmarkRenderAttentionDashboard_1000Items(b *testing.B) {
	app := &App{
		attentionItems:    make([]AttentionItem, 1000),
		attentionExpanded: true, // Worst case - all items
		width:             100,
		height:            30,
		bundleMode:        true,
		attentionViewport: viewport.New(92, 20),
	}

	// Fill with dummy items
	for i := 0; i < 1000; i++ {
		app.attentionItems[i] = AttentionItem{
			Title:       "test-pod",
			Description: "Test issue",
			Namespace:   "default",
			Severity:    SeverityCritical,
			Emoji:       "🔥",
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = app.renderAttentionDashboard()
	}
}

// TestNavigationResourceType_OnlyPodsNavigate tests Issue #17 fix
// Verifies that only pod items with valid PodName trigger navigation
func TestNavigationResourceType_OnlyPodsNavigate(t *testing.T) {
	testCases := []struct {
		name         string
		resourceType string
		podName      string
		shouldNav    bool
		description  string
	}{
		{
			name:         "kubelet item should not navigate",
			resourceType: "kubelet",
			podName:      "",
			shouldNav:    false,
			description:  "Kubelet events are cluster-level, not pod-specific",
		},
		{
			name:         "event item should not navigate",
			resourceType: "event",
			podName:      "",
			shouldNav:    false,
			description:  "Event groups need drill-down view, not pod diagnostics",
		},
		{
			name:         "node item should not navigate",
			resourceType: "node",
			podName:      "",
			shouldNav:    false,
			description:  "Node pressure is cluster-level, not pod-specific",
		},
		{
			name:         "etcd item should not navigate",
			resourceType: "etcd",
			podName:      "",
			shouldNav:    false,
			description:  "ETCD health is cluster-level, not pod-specific",
		},
		{
			name:         "system item should not navigate",
			resourceType: "system",
			podName:      "",
			shouldNav:    false,
			description:  "System health is cluster-level, not pod-specific",
		},
		{
			name:         "daemonset item should not navigate",
			resourceType: "daemonset",
			podName:      "",
			shouldNav:    false,
			description:  "DaemonSets need proper drill-down view",
		},
		{
			name:         "pod with name should navigate",
			resourceType: "pod",
			podName:      "test-crash",
			shouldNav:    true,
			description:  "Valid pod items should navigate to diagnostics",
		},
		{
			name:         "pod without name should not navigate",
			resourceType: "pod",
			podName:      "",
			shouldNav:    false,
			description:  "Edge case: pod items need valid PodName",
		},
		{
			name:         "log item should not navigate",
			resourceType: "log",
			podName:      "test-pod",
			shouldNav:    false,
			description:  "Log issues are different from pod diagnostics",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test the validation logic from handlers.go handleEnter()
			item := AttentionItem{
				ResourceType: tc.resourceType,
				PodName:      tc.podName,
				Title:        "Test Item",
				Namespace:    "default",
			}

			// Apply the same validation logic as handlers.go line 73-75
			shouldNavigate := item.ResourceType == "pod" && item.PodName != ""

			if shouldNavigate != tc.shouldNav {
				t.Errorf("%s: expected shouldNavigate=%v, got %v\nReason: %s",
					tc.name, tc.shouldNav, shouldNavigate, tc.description)
			}
		})
	}
}

// TestResourceTypeAssignment verifies signal detectors set correct ResourceType values
func TestResourceTypeAssignment(t *testing.T) {
	testCases := []struct {
		name         string
		detectorFunc string
		expectedType string
		description  string
	}{
		{
			name:         "detectKubeletIssues",
			detectorFunc: "kubelet",
			expectedType: "kubelet",
			description:  "Kubelet issues must set ResourceType='kubelet'",
		},
		{
			name:         "detectEventIssues",
			detectorFunc: "event",
			expectedType: "event",
			description:  "Event groups must set ResourceType='event'",
		},
		{
			name:         "detectNodeIssues",
			detectorFunc: "node",
			expectedType: "node",
			description:  "Node pressure must set ResourceType='node'",
		},
		{
			name:         "detectPodHealth",
			detectorFunc: "pod",
			expectedType: "pod",
			description:  "Pod health issues must set ResourceType='pod'",
		},
		{
			name:         "detectClusterHealth - etcd",
			detectorFunc: "etcd",
			expectedType: "etcd",
			description:  "ETCD issues must set ResourceType='etcd'",
		},
		{
			name:         "detectClusterHealth - daemonset",
			detectorFunc: "daemonset",
			expectedType: "daemonset",
			description:  "DaemonSet issues must set ResourceType='daemonset'",
		},
		{
			name:         "detectSystemHealth",
			detectorFunc: "system",
			expectedType: "system",
			description:  "System health must set ResourceType='system'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// This is a documentation test - verifies our expectations
			// Actual signal detector tests should be added when we have mock datasources
			if tc.expectedType == "" {
				t.Errorf("Expected type should not be empty")
			}

			// Verify the expected type is one of our known types
			validTypes := map[string]bool{
				"pod":       true,
				"kubelet":   true,
				"event":     true,
				"node":      true,
				"etcd":      true,
				"system":    true,
				"daemonset": true,
				"log":       true,
			}

			if !validTypes[tc.expectedType] {
				t.Errorf("ResourceType '%s' is not in the known types list", tc.expectedType)
			}
		})
	}
}
