// Package recommendations provides mapping from issues to actionable suggestions
package tui

import "github.com/Rancheroo/r8s/internal/datasource"

// GetRecommendation returns recommendation based on diagnostic context
func GetRecommendation(ctx *datasource.DiagnosticContext) string {
	switch ctx.Severity {
	case "critical":
		return ctx.Recommendation + " 🚨 IMMEDIATE ACTION REQUIRED"
	case "high":
		return ctx.Recommendation + " ⚠️ PRIORITIZE INVESTIGATION"
	case "medium":
		return ctx.Recommendation + " 🔍 SCHEDULE REVIEW"
	default:
		return ctx.Recommendation + " ✅ MONITORING OK"
	}
}

// SeverityToEmoji maps severity to emoji indicator
func SeverityToEmoji(severity string) string {
	switch severity {
	case "critical":
		return "🚨"
	case "high":
		return "⚠️"
	case "medium":
		return "🔍"
	default:
		return "✅"
	}
}

// FixPriorityToEmoji maps priority to emoji
func FixPriorityToEmoji(priority string) string {
	switch priority {
	case "immediate":
		return "🚨"
	case "investigate":
		return "🔍"
	case "monitor":
		return "👀"
	default:
		return "✅"
	}
}
