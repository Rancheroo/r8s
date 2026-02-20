// Package cmd implements shared CLI helpers for r8s.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Rancheroo/r8s/internal/bundle"
	"gopkg.in/yaml.v3"
)

// findPodInBundle finds a pod by name with partial matching support
// Extracted to avoid duplication between logs.go and describe.go
func findPodInBundle(b *bundle.Bundle, name string) (*bundle.PodInfo, error) {
	var matches []bundle.PodInfo

	for _, pod := range b.Pods {
		// Exact match
		if pod.Name == name {
			return &pod, nil
		}
		// Partial match
		if strings.Contains(pod.Name, name) {
			matches = append(matches, pod)
		}
	}

	// If no exact match but one partial match, use it
	if len(matches) == 1 {
		return &matches[0], nil
	}

	// Multiple partial matches
	if len(matches) > 1 {
		fmt.Fprintf(os.Stderr, "Multiple pods match '%s':\n", name)
		for _, pod := range matches {
			fmt.Fprintf(os.Stderr, "  - %s (namespace: %s)\n", pod.Name, pod.Namespace)
		}
		return nil, fmt.Errorf("ambiguous pod name '%s' - use full name", name)
	}

	return nil, fmt.Errorf("pod '%s' not found in bundle", name)
}

// outputJSON outputs data as indented JSON
func outputEncodeJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// outputYAML outputs data as YAML
func outputEncodeYAML(data interface{}) error {
	encoder := yaml.NewEncoder(os.Stdout)
	defer encoder.Close()
	return encoder.Encode(data)
}
