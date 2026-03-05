package ui

import (
	"github.com/Rancheroo/r8s/internal/bundle"
)

// AnalysisResult represents the output of bundle analysis
type AnalysisResult struct {
	BundlePath   string              `json:"bundle_path"`
	BundleType   string              `json:"bundle_type"`
	Completeness float64             `json:"completeness"`
	Issues       []Issue             `json:"issues"`
	Critical     int                 `json:"critical_count"`
	Warning      int                 `json:"warning_count"`
	Info         int                 `json:"info_count"`
	Health       *bundle.HealthCheck `json:"health,omitempty"`
}

// Issue represents a single detected issue
type Issue struct {
	Severity   string `json:"severity"`
	Type       string `json:"type"`
	Resource   string `json:"resource"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
}
