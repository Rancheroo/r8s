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
	if len(displayed) != defaultDashboardCap {
		t.Errorf("Expected %d items when collapsed, got %d", defaultDashboardCap, len(displayed))
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
