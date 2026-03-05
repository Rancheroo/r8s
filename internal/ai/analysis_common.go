package ai

// ShouldIncludeMatch determines if a match should be included based on options
func ShouldIncludeMatch(match MatchResultV2, opts AnalysisOptions) bool {
	// Check severity
	if !opts.IncludeInfo && match.Severity == SeverityInfo {
		return false
	}

	// Check confidence
	if match.Confidence == ConfidencePossible && opts.MinConfidence != "" {
		if opts.MinConfidence == ConfidenceCertain || opts.MinConfidence == ConfidenceLikely {
			return false
		}
	}
	if match.Confidence == ConfidenceLikely && opts.MinConfidence == ConfidenceCertain {
		return false
	}

	// Check minimum severity
	if opts.MinSeverity != "" {
		severityOrder := map[Severity]int{
			SeverityInfo:     0,
			SeverityWarning:  1,
			SeverityCritical: 2,
		}
		if severityOrder[match.Severity] < severityOrder[opts.MinSeverity] {
			return false
		}
	}

	return true
}

// ShouldIncludeHint determines if a hint should be included
func ShouldIncludeHint(hint *Hint, opts AnalysisOptions) bool {
	if !opts.IncludeInfo && hint.Severity == SeverityInfo {
		return false
	}

	if opts.MinSeverity != "" {
		severityOrder := map[Severity]int{
			SeverityInfo:     0,
			SeverityWarning:  1,
			SeverityCritical: 2,
		}
		if severityOrder[hint.Severity] < severityOrder[opts.MinSeverity] {
			return false
		}
	}

	return true
}

// ShouldAnalyzePattern determines if a pattern should be analyzed based on options
func ShouldAnalyzePattern(p PatternV2, opts AnalysisOptions) bool {
	// Check severity filter
	if opts.MinSeverity != "" {
		severityOrder := map[Severity]int{
			SeverityInfo:     0,
			SeverityWarning:  1,
			SeverityCritical: 2,
		}
		if severityOrder[p.Severity] < severityOrder[opts.MinSeverity] {
			return false
		}
	}

	return true
}

// DetectCorrelations finds all correlations between matched patterns
func DetectCorrelations(matches []MatchResultV2, registry *PatternRegistryV2) []CorrelationMatch {
	var correlations []CorrelationMatch
	matchedIDs := make(map[string]bool)

	// Build set of matched IDs
	for _, match := range matches {
		if match.Matched {
			matchedIDs[match.PatternID] = true
		}
	}

	// Find correlations
	for _, match := range matches {
		if !match.Matched {
			continue
		}

		pattern, found := registry.GetByID(match.PatternID)
		if !found {
			continue
		}

		for _, corr := range pattern.Correlations {
			if matchedIDs[corr.PatternID] {
				// Both patterns matched - this is a real correlation
				correlations = append(correlations, CorrelationMatch{
					PatternID1: match.PatternID,
					PatternID2: corr.PatternID,
					Message:    corr.Message,
				})
			}
		}
	}

	return correlations
}

// BuildSummary creates analysis summary statistics
func BuildSummary(matches []MatchResultV2, correlations []CorrelationMatch, registry *PatternRegistryV2) AnalysisSummary {
	summary := AnalysisSummary{
		TotalPatterns: len(registry.GetAll()),
		MatchesFound:  0,
		Correlations:  len(correlations),
	}

	for _, match := range matches {
		if match.Matched {
			summary.MatchesFound++
			switch match.Severity {
			case SeverityCritical:
				summary.CriticalIssues++
			case SeverityWarning:
				summary.WarningIssues++
			case SeverityInfo:
				summary.InfoIssues++
			}
		}
	}

	return summary
}
