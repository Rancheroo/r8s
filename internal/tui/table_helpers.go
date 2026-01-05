// Package tui - Table helper functions for dynamic column width calculation
// and consistent text truncation. These helpers implement the "Show, Don't Ask"
// principle by automatically adapting table layouts to terminal width.
package tui

import (
	"strings"

	"github.com/evertras/bubble-table/table"
)

// ColumnSpec defines a column with proportional width
type ColumnSpec struct {
	Key      string  // Column key
	Title    string  // Column header
	Ratio    float64 // Proportion of total width (0.0-1.0)
	MinWidth int     // Minimum width in characters
}

// calculateColumnWidths computes actual column widths based on terminal width
// Implements "Show, Don't Ask" - automatically adapts to any terminal size
func (a *App) calculateColumnWidths(specs []ColumnSpec) []table.Column {
	// Calculate available width (subtract borders, padding, margins)
	availableWidth := a.width - 8
	if availableWidth < 40 {
		availableWidth = 40 // Minimum usable width
	}

	columns := make([]table.Column, len(specs))
	widths := make([]int, len(specs))

	// First pass: calculate proportional widths and enforce minimum
	for i, spec := range specs {
		width := int(float64(availableWidth) * spec.Ratio)

		// Enforce minimum width
		if width < spec.MinWidth {
			width = spec.MinWidth
		}

		widths[i] = width
	}

	// Second pass: adjust if total exceeds available width
	totalWidth := 0
	for _, w := range widths {
		totalWidth += w
	}

	if totalWidth > availableWidth {
		excess := totalWidth - availableWidth

		// Find adjustable columns (those with width > MinWidth)
		adjustable := make([]int, 0, len(specs))
		totalAdjustable := 0
		for i, spec := range specs {
			if widths[i] > spec.MinWidth {
				adjustable = append(adjustable, i)
				totalAdjustable += (widths[i] - spec.MinWidth)
			}
		}

		if len(adjustable) > 0 && totalAdjustable >= excess {
			// Reduce adjustable columns proportionally
			for _, i := range adjustable {
				reduction := int(float64(excess) * float64(widths[i]-specs[i].MinWidth) / float64(totalAdjustable))
				widths[i] -= reduction
			}
		} else {
			// Scale all columns down proportionally to fit
			scaleFactor := float64(availableWidth) / float64(totalWidth)
			for i := range widths {
				widths[i] = int(float64(widths[i]) * scaleFactor)
				if widths[i] < 1 {
					widths[i] = 1
				}
			}
		}
	}

	// Build columns with adjusted widths
	for i, spec := range specs {
		columns[i] = table.NewColumn(spec.Key, spec.Title, widths[i])
	}

	return columns
}

// truncateWithEllipsis truncates text consistently across all tables
// If text exceeds maxWidth, truncate and add "..."
// Implements consistent truncation behavior (audit issue #9)
// Uses rune-safe truncation for proper UTF-8 multi-byte character handling
func truncateWithEllipsis(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	// Convert to runes for proper UTF-8 character handling
	rs := []rune(text)

	if len(rs) <= maxWidth {
		return text
	}

	if maxWidth <= 3 {
		return strings.Repeat(".", maxWidth)
	}

	return string(rs[:maxWidth-3]) + "..."
}

// Common column specifications for different view types
// These define the proportional layouts that auto-adapt to terminal width

func getCRDColumnSpecs() []ColumnSpec {
	return []ColumnSpec{
		{"name", "NAME", 0.35, 20},          // 35% of width, min 20 chars
		{"group", "GROUP", 0.25, 15},        // 25% of width, min 15 chars
		{"kind", "KIND", 0.18, 12},          // 18% of width, min 12 chars
		{"scope", "SCOPE", 0.12, 8},         // 12% of width, min 8 chars
		{"instances", "INSTANCES", 0.10, 6}, // 10% of width, min 6 chars
	}
}

func getClusterColumnSpecs() []ColumnSpec {
	return []ColumnSpec{
		{"name", "NAME", 0.40, 20},         // 40% for cluster names
		{"provider", "PROVIDER", 0.25, 12}, // 25% for provider
		{"state", "STATE", 0.20, 10},       // 20% for state
		{"created", "AGE", 0.15, 8},        // 15% for age
	}
}

func getProjectColumnSpecs() []ColumnSpec {
	return []ColumnSpec{
		{"name", "NAME", 0.35, 20},                // 35% for project name
		{"displayName", "DISPLAY NAME", 0.35, 15}, // 35% for display name
		{"state", "STATE", 0.15, 10},              // 15% for state
		{"namespaces", "NAMESPACES", 0.15, 8},     // 15% for namespace count
	}
}

func getNamespaceColumnSpecs() []ColumnSpec {
	return []ColumnSpec{
		{"name", "NAME", 0.35, 18},       // 35% for namespace name
		{"issues", "ISSUES", 0.20, 12},   // 20% for issue summary
		{"state", "STATE", 0.15, 8},      // 15% for state
		{"project", "PROJECT", 0.20, 12}, // 20% for project
		{"created", "AGE", 0.10, 6},      // 10% for age
	}
}

func getPodColumnSpecs() []ColumnSpec {
	return []ColumnSpec{
		{"name", "NAME", 0.30, 18},           // 30% for pod name
		{"namespace", "NAMESPACE", 0.22, 12}, // 22% for namespace
		{"state", "STATE", 0.18, 10},         // 18% for state
		{"we", "E/W", 0.10, 6},               // 10% for error/warning count
		{"node", "NODE", 0.20, 12},           // 20% for node
	}
}

func getDeploymentColumnSpecs() []ColumnSpec {
	return []ColumnSpec{
		{"name", "NAME", 0.35, 20},           // 35% for deployment name
		{"namespace", "NAMESPACE", 0.25, 12}, // 25% for namespace
		{"ready", "READY", 0.13, 8},          // 13% for ready status
		{"uptodate", "UP-TO-DATE", 0.13, 8},  // 13% for up-to-date
		{"available", "AVAILABLE", 0.14, 8},  // 14% for available
	}
}

func getServiceColumnSpecs() []ColumnSpec {
	return []ColumnSpec{
		{"name", "NAME", 0.25, 15},             // 25% for service name
		{"namespace", "NAMESPACE", 0.20, 12},   // 20% for namespace
		{"type", "TYPE", 0.15, 10},             // 15% for service type
		{"cluster_ip", "CLUSTER-IP", 0.20, 12}, // 20% for cluster IP
		{"ports", "PORT(S)", 0.20, 10},         // 20% for ports
	}
}

func getCRDInstanceColumnSpecs() []ColumnSpec {
	return []ColumnSpec{
		{"name", "NAME", 0.35, 20},           // 35% for instance name
		{"namespace", "NAMESPACE", 0.25, 12}, // 25% for namespace
		{"age", "AGE", 0.15, 8},              // 15% for age
		{"status", "STATUS", 0.25, 12},       // 25% for status
	}
}

func getAttentionColumnSpecs() []ColumnSpec {
	return []ColumnSpec{
		{"severity", "🚨", 0.05, 4},       // 5% for severity emoji
		{"title", "ISSUE", 0.55, 25},     // 55% for issue title
		{"context", "CONTEXT", 0.30, 15}, // 30% for context
		{"count", "COUNT", 0.10, 6},      // 10% for count
	}
}
