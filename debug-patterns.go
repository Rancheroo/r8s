package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"github.com/Rancheroo/r8s/internal/ai"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: debug-patterns <bundle-dir>")
		os.Exit(1)
	}

	bundleDir := os.Args[1]
	analyzer := ai.NewAnalyzer()

	// Check what files exist
	patterns := []string{
		"pod-logs/*.log",
		"pod-logs/*/*.log",
		"journald/*.log",
		"logs/*.log",
	}

	fmt.Println("=== Files in bundle ===")
	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(bundleDir, pattern))
		if err != nil {
			fmt.Printf("  %s: ERROR %v\n", pattern, err)
			continue
		}
		fmt.Printf("  %s: %d files\n", pattern, len(matches))
		if len(matches) > 0 && len(matches) < 10 {
			for _, m := range matches {
				fmt.Printf("    - %s\n", filepath.Base(m))
			}
		}
	}

	// Try to find pods with crashloop/imagepull
	fmt.Println("\n=== Scanning for patterns ===")
	allMatches, _ := filepath.Glob(filepath.Join(bundleDir, "*/*"))
	for _, file := range allMatches {
		if strings.Contains(file, "crash") || strings.Contains(file, "imagepull") || strings.Contains(file, "panic") {
			fmt.Printf("  Found file: %s\n", file)
			
			// Read and check content
			content, err := os.ReadFile(file)
			if err != nil {
				continue
			}
			
			result, _ := analyzer.Analyze(string(content), ai.AnalysisOptions{})
			fmt.Printf("    Matches: %d\n", len(result.Hints))
			for _, hint := range result.Hints {
				fmt.Printf("    - %s: %s\n", hint.PatternID, hint.Summary)
			}
		}
	}

	// Now try reading describe pods
	describeFiles, _ := filepath.Glob(filepath.Join(bundleDir, "kubectl/describe/pods_*.txt"))
	fmt.Println("\n=== Describe pod files ===")
	fmt.Printf("Found %d describe pod files\n", len(describeFiles))
	for _, file := range describeFiles {
		content, _ := os.ReadFile(file)
		if strings.Contains(string(content), "CrashLoop") || strings.Contains(string(content), "ImagePull") {
			fmt.Printf("  %s contains issues\n", filepath.Base(file))
			
			result, _ := analyzer.Analyze(string(content), ai.AnalysisOptions{})
			fmt.Printf("    Pattern matches: %d\n", len(result.Hints))
			for _, hint := range result.Hints {
				fmt.Printf("    - %s\n", hint.PatternID)
			}
		}
	}
}