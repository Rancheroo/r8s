// Package ai provides pattern matching and root cause analysis for Kubernetes issues.
// Sprint 11: Analysis Engine - Integration of patterns, hints, and results
package ai

import (
	"fmt"
	"time"
)

// AnalysisResult represents the complete analysis of a bundle
type AnalysisResult struct {
	StartTime    time.Time          // When analysis started
	EndTime      time.Time          // When analysis completed
	Duration     time.Duration      // Analysis duration
	Patterns     []MatchResultV2    // All pattern matches
	Hints        []*Hint            // Generated hints
	Correlations []CorrelationMatch // Detected correlations
	Summary      AnalysisSummary    // High-level summary
}

// CorrelationMatch represents a detected correlation between patterns
type CorrelationMatch struct {
	PatternID1 string // First pattern
	PatternID2 string // Second pattern
	Message    string // Correlation message
}

// AnalysisSummary provides high-level statistics
type AnalysisSummary struct {
	TotalPatterns  int // Total patterns analyzed
	MatchesFound   int // Number of pattern matches
	CriticalIssues int // Critical severity matches
	WarningIssues  int // Warning severity matches
	InfoIssues     int // Info severity matches
	Correlations   int // Number of correlations detected
}

// Analyzer orchestrates pattern matching and hint generation
type Analyzer struct {
	registry  *PatternRegistryV2
	generator *HintGeneratorV2
	formatter *HintFormatter
}

// NewAnalyzer creates a new analysis engine
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		registry:  NewRegistryV2(),
		generator: NewHintGenerator(),
		formatter: NewHintFormatter(),
	}
}

// Analyze performs complete analysis of bundle content
func (a *Analyzer) Analyze(content string, opts AnalysisOptions) (*AnalysisResult, error) {
	startTime := time.Now()

	// Run pattern analysis
	matches := a.registry.AnalyzeV2(content)

	// Generate hints for matches
	hints := a.generator.GenerateAll(matches, a.registry)

	// Detect correlations
	correlations := DetectCorrelations(matches, a.registry)

	// Build summary
	summary := BuildSummary(matches, correlations, a.registry)

	endTime := time.Now()

	result := &AnalysisResult{
		StartTime:    startTime,
		EndTime:      endTime,
		Duration:     endTime.Sub(startTime),
		Patterns:     matches,
		Hints:        hints,
		Correlations: correlations,
		Summary:      summary,
	}

	return result, nil
}

// AnalysisOptions provides options for analysis
type AnalysisOptions struct {
	MinSeverity   Severity   // Minimum severity to include
	MinConfidence Confidence // Minimum confidence level
	MaxHints      int        // Maximum hints to generate (0 = no limit)
	IncludeInfo   bool       // Include info-level patterns
}

// FilteredAnalyze analyzes with options and filtering
func (a *Analyzer) FilteredAnalyze(content string, opts AnalysisOptions) (*AnalysisResult, error) {
	result, err := a.Analyze(content, opts)
	if err != nil {
		return nil, err
	}

	// Filter patterns by severity
	var filteredPatterns []MatchResultV2
	for _, match := range result.Patterns {
		if ShouldIncludeMatch(match, opts) {
			filteredPatterns = append(filteredPatterns, match)
		}
	}
	result.Patterns = filteredPatterns

	// Regenerate hints for filtered patterns
	result.Hints = a.generator.GenerateAll(filteredPatterns, a.registry)

	// Filter hints by severity too
	var filteredHints []*Hint
	for _, hint := range result.Hints {
		if ShouldIncludeHint(hint, opts) {
			filteredHints = append(filteredHints, hint)
		}
	}
	result.Hints = filteredHints

	// Limit hints if specified
	if opts.MaxHints > 0 && len(result.Hints) > opts.MaxHints {
		result.Hints = result.Hints[:opts.MaxHints]
	}

	// Rebuild summary with filtered results
	result.Summary = BuildSummary(filteredPatterns, result.Correlations, a.registry)

	return result, nil
}

// FormatResults formats analysis results as human-readable string
func (a *Analyzer) FormatResults(result *AnalysisResult) string {
	var output string

	// Header
	output += fmt.Sprintf("Analysis completed in %v\n", result.Duration)
	output += fmt.Sprintf("Patterns analyzed: %d | Matches: %d\n",
		result.Summary.TotalPatterns, result.Summary.MatchesFound)

	// Summary
	if result.Summary.CriticalIssues > 0 || result.Summary.WarningIssues > 0 {
		output += "\n📊 Summary:\n"
		if result.Summary.CriticalIssues > 0 {
			output += fmt.Sprintf("  🔴 %d Critical\n", result.Summary.CriticalIssues)
		}
		if result.Summary.WarningIssues > 0 {
			output += fmt.Sprintf("  🟡 %d Warning\n", result.Summary.WarningIssues)
		}
		if result.Summary.InfoIssues > 0 {
			output += fmt.Sprintf("  🔵 %d Info\n", result.Summary.InfoIssues)
		}
	}

	// Correlations
	if len(result.Correlations) > 0 {
		output += fmt.Sprintf("\n🔗 Correlations detected: %d\n", len(result.Correlations))
		for _, corr := range result.Correlations {
			output += fmt.Sprintf("  • %s ↔ %s: %s\n", corr.PatternID1, corr.PatternID2, corr.Message)
		}
	}

	// Hints
	if len(result.Hints) > 0 {
		output += "\n📝 Root Cause Analysis:\n"
		for i, hint := range result.Hints {
			output += fmt.Sprintf("\n━━━ Hint %d ━━━\n", i+1)
			output += a.generator.FormatHint(hint)
			output += "\n"
		}
	} else {
		output += "\n✅ No issues detected.\n"
	}

	return output
}

// GetCriticalHints returns only critical hints
func (a *Analyzer) GetCriticalHints(result *AnalysisResult) []*Hint {
	return FilterHints(result.Hints, SeverityCritical)
}

// GetHintsByCategory returns hints filtered by category
func (a *Analyzer) GetHintsByCategory(result *AnalysisResult, category string) []*Hint {
	var filtered []*Hint
	for _, hint := range result.Hints {
		pattern, found := a.registry.GetByID(hint.PatternID)
		if found && pattern.Category == category {
			filtered = append(filtered, hint)
		}
	}
	return filtered
}

// AnalyzeMultiple analyzes multiple content sources
func (a *Analyzer) AnalyzeMultiple(contents map[string]string, opts AnalysisOptions) (map[string]*AnalysisResult, error) {
	results := make(map[string]*AnalysisResult)

	for name, content := range contents {
		result, err := a.FilteredAnalyze(content, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze %s: %w", name, err)
		}
		results[name] = result
	}

	return results, nil
}

// PatternStats provides statistics about pattern usage
type PatternStats struct {
	PatternID     string
	Name          string
	Category      string
	MatchCount    int
	HintCount     int
	AvgConfidence Confidence
}

// GetPatternStats returns usage statistics for all patterns
func (a *Analyzer) GetPatternStats(results []*AnalysisResult) []PatternStats {
	stats := make(map[string]*PatternStats)
	confidenceSum := make(map[string]int) // Track confidence sums for averaging

	// Initialize stats for all patterns
	for _, pattern := range a.registry.GetAll() {
		stats[pattern.ID] = &PatternStats{
			PatternID: pattern.ID,
			Name:      pattern.Name,
			Category:  pattern.Category,
		}
	}

	// Accumulate stats from results
	for _, result := range results {
		for _, match := range result.Patterns {
			if stat, found := stats[match.PatternID]; found {
				stat.MatchCount++
				// Accumulate confidence for averaging
				confidenceSum[match.PatternID] += confidenceValue(match.Confidence)
			}
		}
		for _, hint := range result.Hints {
			if stat, found := stats[hint.PatternID]; found {
				stat.HintCount++
			}
		}
	}

	// Convert to slice and compute averages
	var result []PatternStats
	for _, stat := range stats {
		if stat.MatchCount > 0 {
			avg := confidenceSum[stat.PatternID] / stat.MatchCount
			stat.AvgConfidence = confidenceFromValue(avg)
		}
		result = append(result, *stat)
	}
	return result
}

// confidenceValue converts Confidence to numeric value for averaging
func confidenceValue(c Confidence) int {
	switch c {
	case ConfidenceCertain:
		return 3
	case ConfidenceLikely:
		return 2
	case ConfidencePossible:
		return 1
	default:
		return 0
	}
}

// confidenceFromValue converts numeric value back to Confidence
func confidenceFromValue(v int) Confidence {
	switch {
	case v >= 3:
		return ConfidenceCertain
	case v >= 2:
		return ConfidenceLikely
	case v >= 1:
		return ConfidencePossible
	default:
		return ""
	}
}
