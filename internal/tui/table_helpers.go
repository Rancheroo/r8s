// Package tui - Table helper functions for dynamic column width calculation
// and consistent text truncation. These helpers implement the "Show, Don't Ask"
// principle by automatically adapting table layouts to terminal width.
package tui

import (
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

	// First pass: calculate proportional widths
	for i, spec := range specs {
		width := int(float64(availableWidth) * spec.Ratio)

		// Enforce minimum width
		if width < spec.MinWidth {
			width = spec.MinWidth
		}

		columns[i] = table.NewColumn(spec.Key, spec.Title, width)
	}

	return columns
}

// truncateWithEllipsis truncates text consistently across all tables
// If text exceeds maxWidth, truncate and add "..."
func getCRDColumnSpecs() []ColumnSpec {
	return []ColumnSpec{
		{"name", "NAME", 0.35, 20},          // 35% of width, min 20 chars
		{"group", "GROUP", 0.25, 15},        // 25% of width, min 15 chars
		{"kind", "KIND", 0.18, 12},          // 18% of width, min 12 chars
		{"scope", "SCOPE", 0.12, 8},         // 12% of width, min 8 chars
		{"instances", "INSTANCES", 0.10, 6}, // 10% of width, min 6 chars
	}
}

// getClusterColumnSpecs returns column specifications for the clusters table.
// Each ColumnSpec defines the column key, header title, proportional width ratio, and minimum width used to layout cluster list views.
func getClusterColumnSpecs() []ColumnSpec {
	return []ColumnSpec{
		{"name", "NAME", 0.40, 20},         // 40% for cluster names
		{"provider", "PROVIDER", 0.25, 12}, // 25% for provider
		{"state", "STATE", 0.20, 10},       // 20% for state
		{"created", "AGE", 0.15, 8},        // 15% for age
	}
}

// getProjectColumnSpecs returns the column specifications for the projects table.
// The slice defines columns: project name (key "name", 35% ratio, min width 20), display name (key "displayName", 35% ratio, min width 15), state (key "state", 15% ratio, min width 10), and namespaces (key "namespaces", 15% ratio, min width 8).
func getProjectColumnSpecs() []ColumnSpec {
	return []ColumnSpec{
		{"name", "NAME", 0.35, 20},                // 35% for project name
		{"displayName", "DISPLAY NAME", 0.35, 15}, // 35% for display name
		{"state", "STATE", 0.15, 10},              // 15% for state
		{"namespaces", "NAMESPACES", 0.15, 8},     // 15% for namespace count
	}
}

// getNamespaceColumnSpecs returns the column specifications for the namespace table view.
// Each ColumnSpec defines the column key, header title, proportional width (Ratio), and minimum width in characters.
// The specs allocate space as: name (35%, min 18), issues (20%, min 12), state (15%, min 8), project (20%, min 12), created/age (10%, min 6).
func getNamespaceColumnSpecs() []ColumnSpec {
	return []ColumnSpec{
		{"name", "NAME", 0.35, 18},       // 35% for namespace name
		{"issues", "ISSUES", 0.20, 12},   // 20% for issue summary
		{"state", "STATE", 0.15, 8},      // 15% for state
		{"project", "PROJECT", 0.20, 12}, // 20% for project
		{"created", "AGE", 0.10, 6},      // 10% for age
	}
}

// getPodColumnSpecs returns the default ColumnSpec slice used to render pod table views.
// Columns: "name" (NAME) — pod name, 30% ratio, min width 18; "namespace" (NAMESPACE) — 22% ratio, min 12;
// "state" (STATE) — 18% ratio, min 10; "we" (E/W) — error/warning count, 10% ratio, min 6;
// "node" (NODE) — 20% ratio, min 12.
func getPodColumnSpecs() []ColumnSpec {
	return []ColumnSpec{
		{"name", "NAME", 0.30, 18},           // 30% for pod name
		{"namespace", "NAMESPACE", 0.22, 12}, // 22% for namespace
		{"state", "STATE", 0.18, 10},         // 18% for state
		{"we", "E/W", 0.10, 6},               // 10% for error/warning count
		{"node", "NODE", 0.20, 12},           // 20% for node
	}
}

// getDeploymentColumnSpecs returns the default ColumnSpec slice used to render the deployments table.
// Each spec defines the column Key, Title, proportional Ratio (0.0–1.0) of available width, and a minimum width in characters.
func getDeploymentColumnSpecs() []ColumnSpec {
	return []ColumnSpec{
		{"name", "NAME", 0.35, 20},           // 35% for deployment name
		{"namespace", "NAMESPACE", 0.25, 12}, // 25% for namespace
		{"ready", "READY", 0.13, 8},          // 13% for ready status
		{"uptodate", "UP-TO-DATE", 0.13, 8},  // 13% for up-to-date
		{"available", "AVAILABLE", 0.14, 8},  // 14% for available
	}
}

// getServiceColumnSpecs returns column specifications for the services table view.
// Columns provided: "name", "namespace", "type", "cluster_ip", and "ports" with suggested proportional ratios and minimum widths for terminal layout.
func getServiceColumnSpecs() []ColumnSpec {
	return []ColumnSpec{
		{"name", "NAME", 0.25, 15},             // 25% for service name
		{"namespace", "NAMESPACE", 0.20, 12},   // 20% for namespace
		{"type", "TYPE", 0.15, 10},             // 15% for service type
		{"cluster_ip", "CLUSTER-IP", 0.20, 12}, // 20% for cluster IP
		{"ports", "PORT(S)", 0.20, 10},         // 20% for ports
	}
}

// getCRDInstanceColumnSpecs returns the column specifications for the CRD instance table view.
// The returned slice defines columns and their proportional widths and minimums: name (35%, min 20),
// namespace (25%, min 12), age (15%, min 8), and status (25%, min 12).
func getCRDInstanceColumnSpecs() []ColumnSpec {
	return []ColumnSpec{
		{"name", "NAME", 0.35, 20},           // 35% for instance name
		{"namespace", "NAMESPACE", 0.25, 12}, // 25% for namespace
		{"age", "AGE", 0.15, 8},              // 15% for age
		{"status", "STATUS", 0.25, 12},       // 25% for status
	}
}

// getAttentionColumnSpecs returns the default column specifications for the attention table view.
// The specs define columns for severity, title, context, and count, including each column's display title, proportional width ratio, and minimum character width.
func getAttentionColumnSpecs() []ColumnSpec {
	return []ColumnSpec{
		{"severity", "🚨", 0.05, 4},       // 5% for severity emoji
		{"title", "ISSUE", 0.55, 25},     // 55% for issue title
		{"context", "CONTEXT", 0.30, 15}, // 30% for context
		{"count", "COUNT", 0.10, 6},      // 10% for count
	}
}

// getContainerSelectColumnSpecs returns column specifications for the container selection table.
// S4-HIGH-2: Terminal-adaptive column widths for multi-container pod view
func getContainerSelectColumnSpecs() []ColumnSpec {
	return []ColumnSpec{
		{"container", "Container", 0.35, 20},  // 35% for container name
		{"status", "Status", 0.15, 10},        // 15% for status
		{"restarts", "Restarts", 0.15, 8},     // 15% for restart count
		{"resources", "Resources", 0.35, 20},  // 35% for resource limits
	}
}