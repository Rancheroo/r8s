// Package cmd implements export format generators for r8s.
// Sprint 11 Day 8: SARIF, JUnit XML, and Markdown export formats.
package cmd

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/Rancheroo/r8s/internal/ai"
	"github.com/Rancheroo/r8s/internal/bundle"
)

// SARIF Types
// https://docs.oasis-open.org/sarif/sarif/v2.1.0/cs01/sarif-v2.1.0-cs01.html

type SARIFLog struct {
	Version string     `json:"version"`
	Schema  string     `json:"$schema"`
	Runs    []SARIFRun `json:"runs"`
}

type SARIFRun struct {
	Tool      SARIFTool       `json:"tool"`
	Results   []SARIFResult   `json:"results"`
	Invocations []SARIFInvocation `json:"invocations,omitempty"`
}

type SARIFTool struct {
	Driver SARIFDriver `json:"driver"`
}

type SARIFDriver struct {
	Name            string       `json:"name"`
	Version         string       `json:"version"`
	InformationURI  string       `json:"informationUri"`
	Rules           []SARIFRule  `json:"rules"`
}

type SARIFRule struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	ShortDescription SARIFMessage      `json:"shortDescription"`
	FullDescription  SARIFMessage      `json:"fullDescription"`
	DefaultConfiguration SARIFConfig   `json:"defaultConfiguration"`
	HelpURI          string            `json:"helpUri,omitempty"`
}

type SARIFMessage struct {
	Text string `json:"text"`
}

type SARIFConfig struct {
	Level string `json:"level"`
}

type SARIFResult struct {
	RuleID    string          `json:"ruleId"`
	Level     string          `json:"level"`
	Message   SARIFMessage    `json:"message"`
	Locations []SARIFLocation `json:"locations,omitempty"`
}

type SARIFLocation struct {
	PhysicalLocation SARIFPhysicalLocation `json:"physicalLocation"`
}

type SARIFPhysicalLocation struct {
	ArtifactLocation SARIFArtifactLocation `json:"artifactLocation"`
	Region           *SARIFRegion          `json:"region,omitempty"`
}

type SARIFArtifactLocation struct {
	URI string `json:"uri"`
}

type SARIFRegion struct {
	StartLine int `json:"startLine"`
}

type SARIFInvocation struct {
	ExecutionSuccessful bool      `json:"executionSuccessful"`
	StartTimeUTC        string    `json:"startTimeUtc"`
	EndTimeUTC          string    `json:"endTimeUtc"`
}

// severityToSARIFLevel converts internal severity to SARIF level
func severityToSARIFLevel(s ai.Severity) string {
	switch s {
	case ai.SeverityCritical:
		return "error"
	case ai.SeverityWarning:
		return "warning"
	case ai.SeverityInfo:
		return "note"
	default:
		return "none"
	}
}

// exportSARIF generates SARIF output for GitHub Advanced Security
func exportSARIF(bundlePath string, health *bundle.HealthCheck, hints []*ai.Hint) ([]byte, error) {
	// Build rules from hints
	rulesMap := make(map[string]SARIFRule)
	var results []SARIFResult

	for _, hint := range hints {
		// Add rule if not exists
		if _, exists := rulesMap[hint.PatternID]; !exists {
			rulesMap[hint.PatternID] = SARIFRule{
				ID:   hint.PatternID,
				Name: hint.PatternID,
				ShortDescription: SARIFMessage{
					Text: hint.Summary,
				},
				FullDescription: SARIFMessage{
					Text: hint.Explanation,
				},
				DefaultConfiguration: SARIFConfig{
					Level: severityToSARIFLevel(hint.Severity),
				},
				HelpURI: getFirstReference(hint.References),
			}
		}

		// Add result
		results = append(results, SARIFResult{
			RuleID: hint.PatternID,
			Level:  severityToSARIFLevel(hint.Severity),
			Message: SARIFMessage{
				Text: fmt.Sprintf("%s\n\nSuggestion: %s", hint.Summary, hint.Suggestion),
			},
			Locations: []SARIFLocation{
				{
					PhysicalLocation: SARIFPhysicalLocation{
						ArtifactLocation: SARIFArtifactLocation{
							URI: bundlePath,
						},
					},
				},
			},
		})
	}

	// Convert rules map to slice
	var rules []SARIFRule
	for _, rule := range rulesMap {
		rules = append(rules, rule)
	}

	log := SARIFLog{
		Version: "2.1.0",
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Runs: []SARIFRun{
			{
				Tool: SARIFTool{
					Driver: SARIFDriver{
						Name:           "r8s",
						Version:        "v0.9.0",
						InformationURI: "https://github.com/Rancheroo/r8s",
						Rules:          rules,
					},
				},
				Results: results,
				Invocations: []SARIFInvocation{
					{
						ExecutionSuccessful: true,
						StartTimeUTC:        time.Now().UTC().Format(time.RFC3339),
						EndTimeUTC:          time.Now().UTC().Format(time.RFC3339),
					},
				},
			},
		},
	}

	return json.MarshalIndent(log, "", "  ")
}

// getFirstReference returns first reference URL or empty
func getFirstReference(refs []string) string {
	if len(refs) > 0 {
		return refs[0]
	}
	return ""
}

// JUnit Types
type JUnitTestSuites struct {
	XMLName    xml.Name         `xml:"testsuites"`
	Name       string           `xml:"name,attr"`
	Tests      int              `xml:"tests,attr"`
	Failures   int              `xml:"failures,attr"`
	Errors     int              `xml:"errors,attr"`
	Time       float64          `xml:"time,attr"`
	TestSuites []JUnitTestSuite `xml:"testsuite"`
}

type JUnitTestSuite struct {
	Name      string           `xml:"name,attr"`
	Tests     int              `xml:"tests,attr"`
	Failures  int              `xml:"failures,attr"`
	Errors    int              `xml:"errors,attr"`
	Time      float64          `xml:"time,attr"`
	Timestamp string           `xml:"timestamp,attr"`
	TestCases []JUnitTestCase  `xml:"testcase"`
}

type JUnitTestCase struct {
	Name      string         `xml:"name,attr"`
	ClassName string         `xml:"classname,attr"`
	Time      float64        `xml:"time,attr"`
	Failure   *JUnitFailure  `xml:"failure,omitempty"`
	Error     *JUnitError    `xml:"error,omitempty"`
}

type JUnitFailure struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Text    string `xml:",chardata"`
}

type JUnitError struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Text    string `xml:",chardata"`
}

// exportJUnit generates JUnit XML output for CI/CD integration
func exportJUnit(bundlePath string, health *bundle.HealthCheck, hints []*ai.Hint) ([]byte, error) {
	now := time.Now().UTC().Format(time.RFC3339)

	// Group hints by category
	byCategory := make(map[string][]*ai.Hint)
	for _, hint := range hints {
		// Extract category from pattern ID or use "general"
		category := "general"
		if hint.PatternID != "" {
			parts := strings.Split(hint.PatternID, "-")
			if len(parts) > 0 {
				category = parts[0]
			}
		}
		byCategory[category] = append(byCategory[category], hint)
	}

	var suites []JUnitTestSuite
	failures := 0
	errors := 0

	for category, catHints := range byCategory {
		var testCases []JUnitTestCase
		catFailures := 0
		catErrors := 0

		for _, hint := range catHints {
			tc := JUnitTestCase{
				Name:      hint.PatternID,
				ClassName: "r8s.analysis." + category,
				Time:      0.001, // Placeholder
			}

			switch hint.Severity {
			case ai.SeverityCritical:
				tc.Error = &JUnitError{
					Message: hint.Summary,
					Type:    "critical",
					Text:    fmt.Sprintf("%s\n\nSuggestion: %s", hint.Explanation, hint.Suggestion),
				}
				errors++
				catErrors++
			case ai.SeverityWarning:
				tc.Failure = &JUnitFailure{
					Message: hint.Summary,
					Type:    "warning",
					Text:    fmt.Sprintf("%s\n\nSuggestion: %s", hint.Explanation, hint.Suggestion),
				}
				failures++
				catFailures++
			}

			testCases = append(testCases, tc)
		}

		suites = append(suites, JUnitTestSuite{
			Name:      category,
			Tests:     len(catHints),
			Failures:  catFailures,
			Errors:    catErrors,
			Time:      0.001,
			Timestamp: now,
			TestCases: testCases,
		})
	}

	suitesContainer := JUnitTestSuites{
		Name:       "r8s-bundle-analysis",
		Tests:      len(hints),
		Failures:   failures,
		Errors:     errors,
		Time:       0.001,
		TestSuites: suites,
	}

	// Add XML header
	output := []byte(`<?xml version="1.0" encoding="UTF-8"?>
`)
	xmlBytes, err := xml.MarshalIndent(suitesContainer, "", "  ")
	if err != nil {
		return nil, err
	}

	return append(output, xmlBytes...), nil
}

// exportMarkdown generates Markdown report for human reading
func exportMarkdown(bundlePath string, health *bundle.HealthCheck, hints []*ai.Hint) ([]byte, error) {
	var sb strings.Builder

	// Header
	sb.WriteString("# R8S Bundle Analysis Report\n\n")
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("**Bundle:** `%s`\n\n", bundlePath))
	sb.WriteString(fmt.Sprintf("**Type:** %s\n\n", health.BundleType))
	sb.WriteString(fmt.Sprintf("**Completeness:** %.0f%%\n\n", health.Completeness))

	// Summary
	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Total Issues:** %d\n", len(hints)))

	// Count by severity
	critical := 0
	warning := 0
	info := 0
	for _, hint := range hints {
		switch hint.Severity {
		case ai.SeverityCritical:
			critical++
		case ai.SeverityWarning:
			warning++
		case ai.SeverityInfo:
			info++
		}
	}

	if critical > 0 {
		sb.WriteString(fmt.Sprintf("- **🔴 Critical:** %d\n", critical))
	}
	if warning > 0 {
		sb.WriteString(fmt.Sprintf("- **🟡 Warning:** %d\n", warning))
	}
	if info > 0 {
		sb.WriteString(fmt.Sprintf("- **🔵 Info:** %d\n", info))
	}

	sb.WriteString("\n")

	// Detailed findings
	if len(hints) > 0 {
		sb.WriteString("## Findings\n\n")

		// Group by severity
		for i, hint := range hints {
			severityIcon := "🔵"
			if hint.Severity == ai.SeverityCritical {
				severityIcon = "🔴"
			} else if hint.Severity == ai.SeverityWarning {
				severityIcon = "🟡"
			}

			sb.WriteString(fmt.Sprintf("### %d. %s %s\n\n", i+1, severityIcon, hint.PatternID))
			sb.WriteString(fmt.Sprintf("**Summary:** %s\n\n", hint.Summary))

			if hint.Explanation != "" {
				sb.WriteString("**Details:**\n")
				sb.WriteString(hint.Explanation)
				sb.WriteString("\n\n")
			}

			if hint.Suggestion != "" {
				sb.WriteString("**Suggestion:**\n")
				sb.WriteString(hint.Suggestion)
				sb.WriteString("\n\n")
			}

			if hint.Command != "" {
				sb.WriteString("**Command:**\n")
				sb.WriteString("```bash\n")
				sb.WriteString(hint.Command)
				sb.WriteString("\n```\n\n")
			}

			if len(hint.References) > 0 {
				sb.WriteString("**References:**\n")
				for _, ref := range hint.References {
					sb.WriteString(fmt.Sprintf("- <%s>\n", ref))
				}
				sb.WriteString("\n")
			}

			sb.WriteString("---\n\n")
		}
	}

	return []byte(sb.String()), nil
}
