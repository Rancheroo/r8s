package cmd

import (
	"fmt"
	"testing"
)

// TestAskCombinations verifies that common NLQ patterns are parsed into correct intents.
// This matrix test covers 750+ combinations of prefixes, resources, and conditions.
func TestAskCombinations(t *testing.T) {
	// 1. Define components
	resources := []struct {
		text     string
		expected string
	}{
		{"pod", "pod"},
		{"container", "pod"},
		{"certificate", "certificate"},
		{"cert", "certificate"},
		{"image", "image"},
		{"docker", "image"},
		{"node", "node"},
	}

	conditions := []struct {
		text     string
		expected string
	}{
		{"crashing", "crashing"},
		{"crashloop", "crashing"},
		{"restarting", "crashing"},
		{"imagepull", "imagepull"},
		{"imagepullbackoff", "imagepull"},
		{"can't pull", "imagepull"},
		{"pending", "pending"},
		{"stuck", "pending"},
		{"expired", "expired"},
		{"not ready", "notready"},
		{"unhealthy", "notready"},
		{"oom", "oom"},
		{"out of memory", "oom"},
		{"failed", "failed"},
		{"broken", "failed"},
		{"ready", "ready"},
		{"running", "ready"},
		{"slow", "latency"},
		{"latency", "latency"},
	}

	prefixes := []struct {
		text     string
		typeExp  string
	}{
		{"why is", "why"},
		{"show me", "show"},
		{"find", "show"},
		{"list", "show"},
		{"which", "which"},
		{"what is wrong with", "what"}, // This one is tricky if we add condition words, but let's see
	}

	var totalTests int
	var passedTests int

	// 2. Generate and test combinations
	for _, p := range prefixes {
		for _, r := range resources {
			for _, c := range conditions {
				// Construct query: "prefix resource condition"
				// e.g. "why is pod crashing"
				query := fmt.Sprintf("%s %s %s", p.text, r.text, c.text)
				
				totalTests++
				t.Run(query, func(t *testing.T) {
					intent := parseQueryIntent(query)

					// Check Type
					if intent.Type != p.typeExp {
						t.Errorf("Query: '%s' -> Type: %s, Want: %s", query, intent.Type, p.typeExp)
						return
					}

					// Check Resource
					if intent.Resource != r.expected {
						t.Errorf("Query: '%s' -> Resource: %s, Want: %s", query, intent.Resource, r.expected)
						return
					}

					// Check Condition
					if intent.Condition != c.expected {
						t.Errorf("Query: '%s' -> Condition: %s, Want: %s", query, intent.Condition, c.expected)
						return
					}
					passedTests++
				})
			}
		}
	}
	t.Logf("Total Combinations Tested: %d", totalTests)
}
