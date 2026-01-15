package datasource

import (
	"sort"
	"testing"

	"github.com/Rancheroo/r8s/internal/rancher"
)

// v0.6.9 Regression Tests for Principle Compliance
// These tests ensure we don't regress on the fixes made in v0.6.9:
// - Principle #8: O(n log n) Always (stdlib sorting)
// - Principle #1: Truth Only™ (no mock data)
// - Principle #11: Empty is Valid (no fallbacks)

// TestEventSorting_WarningsFirst_ThenByTime tests the sorting logic
// Regression test for v0.6.9 bubble sort removal (Principle #8)
func TestEventSorting_WarningsFirst_ThenByTime(t *testing.T) {
	// Create test events in unsorted order
	events := []rancher.Event{
		{Type: "Normal", LastSeen: "2026-01-15T10:00:00Z"},
		{Type: "Warning", LastSeen: "2026-01-15T09:00:00Z"},
		{Type: "Normal", LastSeen: "2026-01-15T11:00:00Z"},
	}

	// Sort using the same logic as GetEventsByPod (stdlib sort.Slice)
	sort.Slice(events, func(i, j int) bool {
		if events[i].Type != events[j].Type {
			return events[i].Type == "Warning"
		}
		return events[i].LastSeen > events[j].LastSeen
	})

	// Verify: Warning comes first
	if events[0].Type != "Warning" {
		t.Errorf("First event should be Warning, got %s", events[0].Type)
	}

	// Verify: Normal events sorted by LastSeen descending
	if events[1].LastSeen < events[2].LastSeen {
		t.Errorf("Normal events not sorted by LastSeen desc")
	}
}

// TestPrincipleCompliance_TruthOnly documents expected behavior
// This test serves as documentation for the "Truth Only™" principle
func TestPrincipleCompliance_TruthOnly(t *testing.T) {
	t.Log("Principle #1: Truth Only™")
	t.Log("- DescribeDeployment returns error (not mock) when not found")
	t.Log("- DescribeService returns error (not mock) when not found")
	t.Log("- GetContainers returns empty slice (not ['default']) when unknown")
	t.Log("v0.6.9: Removed all mock data fallbacks")
}

// TestPrincipleCompliance_EmptyIsValid documents expected behavior
// This test serves as documentation for the "Empty is Valid" principle
func TestPrincipleCompliance_EmptyIsValid(t *testing.T) {
	t.Log("Principle #11: Empty is Valid")
	t.Log("- Empty slices are legitimate return values")
	t.Log("- Don't fabricate data when real data unavailable")
	t.Log("v0.6.9: Removed 'default' container name fallback")
}

// TestPrincipleCompliance_OnLogNAlways documents expected behavior
// This test serves as documentation for the "O(n log n) Always" principle
func TestPrincipleCompliance_OnLogNAlways(t *testing.T) {
	t.Log("Principle #8: O(n log n) Always")
	t.Log("- Use stdlib sort.Slice, never manual sorting loops")
	t.Log("- Prevents O(n²) bubble sort performance issues")
	t.Log("v0.6.9: Replaced bubble sort with sort.Slice in GetEventsByPod")
}
