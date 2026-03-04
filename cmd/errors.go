// Package cmd implements friendly error handling and command suggestions.
// Sprint 12: Better error messages with helpful suggestions
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// SuggestedCommand represents a command suggestion
type SuggestedCommand struct {
	Command     string
	Description string
	Example     string
}

// CommandSuggestions - Map of common typos to correct commands
var CommandSuggestions = map[string]SuggestedCommand{
	"analize":   {Command: "analyze", Description: "Analyze a Rancher support bundle", Example: "r8s analyze ./bundle/"},
	"analyse":   {Command: "analyze", Description: "Analyze a Rancher support bundle", Example: "r8s analyze ./bundle/"},
	"analaysis": {Command: "analyze", Description: "Analyze a Rancher support bundle", Example: "r8s analyze ./bundle/"},
	"askk":      {Command: "ask", Description: "Ask natural language questions", Example: "r8s ask ./bundle/ 'why is pod crashing?'"},
	"descibe":   {Command: "describe", Description: "Describe resources in bundle", Example: "r8s describe ./bundle/ pod nginx"},
	"desc":      {Command: "describe", Description: "Describe resources in bundle", Example: "r8s describe ./bundle/ pod nginx"},
	"gett":      {Command: "get", Description: "Get resources (like kubectl)", Example: "r8s get pods ./bundle/"},
	"log":       {Command: "logs", Description: "View pod logs", Example: "r8s logs ./bundle/ nginx-pod"},
	"logg":      {Command: "logs", Description: "View pod logs", Example: "r8s logs ./bundle/ nginx-pod"},
	"loogs":     {Command: "logs", Description: "View pod logs", Example: "r8s logs ./bundle/ nginx-pod"},
	"validat":   {Command: "validate", Description: "Validate bundle completeness", Example: "r8s validate ./bundle/"},
	"val":       {Command: "validate", Description: "Validate bundle completeness", Example: "r8s validate ./bundle/"},
	"check":     {Command: "validate", Description: "Validate bundle completeness", Example: "r8s validate ./bundle/"},
	"exportt":   {Command: "export", Description: "Export analysis results", Example: "r8s export ./bundle/ --format=json"},
	"generat":   {Command: "generate", Description: "Generate prompts/reports", Example: "r8s generate prompt ./bundle/"},
	"gen":       {Command: "generate", Description: "Generate prompts/reports", Example: "r8s generate prompt ./bundle/"},
	"configg":   {Command: "config", Description: "Manage r8s configuration", Example: "r8s config init"},
	"versionn":  {Command: "version", Description: "Show version information", Example: "r8s version"},
	"ver":       {Command: "version", Description: "Show version information", Example: "r8s version"},
	"v":         {Command: "version", Description: "Show version information", Example: "r8s version"},
	"helpp":     {Command: "help", Description: "Show help information", Example: "r8s help"},
	"h":         {Command: "help", Description: "Show help information", Example: "r8s help"},
}

// AvailableCommands - List of all available commands for suggestions
var AvailableCommands = []SuggestedCommand{
	{Command: "analyze", Description: "Analyze bundle and detect issues", Example: "r8s analyze ./bundle/"},
	{Command: "ask", Description: "Ask natural language questions", Example: "r8s ask ./bundle/ 'why is pod crashing?'"},
	{Command: "describe", Description: "Describe resources (like kubectl)", Example: "r8s describe ./bundle/ pod nginx"},
	{Command: "export", Description: "Export analysis results", Example: "r8s export ./bundle/ -o results.json"},
	{Command: "generate", Description: "Generate AI prompts/reports", Example: "r8s generate prompt ./bundle/"},
	{Command: "get", Description: "Get resources (pods, nodes, etc.)", Example: "r8s get pods ./bundle/"},
	{Command: "logs", Description: "View pod logs", Example: "r8s logs ./bundle/ nginx-pod"},
	{Command: "validate", Description: "Validate bundle completeness", Example: "r8s validate ./bundle/"},
	{Command: "version", Description: "Show version information", Example: "r8s version"},
	{Command: "config", Description: "Manage configuration", Example: "r8s config init"},
	{Command: "completion", Description: "Generate shell completion", Example: "r8s completion bash"},
}

// ShowUnknownCommandError displays a friendly error for unknown commands
func ShowUnknownCommandError(cmdName string) {
	fmt.Fprintln(os.Stderr)
	
	// Header
	header := color.New(color.Bold, color.FgRed)
	header.Fprintf(os.Stderr, "🤠 Howdy! Unknown command: '%s'\n", cmdName)
	fmt.Fprintln(os.Stderr)
	
	// Check for exact match in suggestions
	if suggestion, found := CommandSuggestions[strings.ToLower(cmdName)]; found {
		suggestColor := color.New(color.Bold, color.FgYellow)
		suggestColor.Fprintln(os.Stderr, "Did you mean:")
		fmt.Fprintln(os.Stderr)
		
		cmdColor := color.New(color.Bold, color.FgCyan)
		descColor := color.New(color.FgWhite)
		exampleColor := color.New(color.FgHiBlack)
		
		cmdColor.Fprintf(os.Stderr, "  r8s %s", suggestion.Command)
		descColor.Fprintf(os.Stderr, " - %s\n", suggestion.Description)
		exampleColor.Fprintf(os.Stderr, "    Example: %s\n", suggestion.Example)
		fmt.Fprintln(os.Stderr)
		return
	}
	
	// Try to find similar commands
	similar := findSimilarCommands(cmdName)
	if len(similar) > 0 {
		suggestColor := color.New(color.Bold, color.FgYellow)
		suggestColor.Fprintln(os.Stderr, "Did you mean one of these?")
		fmt.Fprintln(os.Stderr)
		
		for _, cmd := range similar {
			cmdColor := color.New(color.Bold, color.FgCyan)
			descColor := color.New(color.FgWhite)
			
			cmdColor.Fprintf(os.Stderr, "  r8s %s", cmd.Command)
			descColor.Fprintf(os.Stderr, " - %s\n", cmd.Description)
		}
		fmt.Fprintln(os.Stderr)
	}
	
	// Show all available commands
	showAvailableCommands()
	
	// Show tip
	tipColor := color.New(color.Italic, color.FgHiBlack)
	tipColor.Fprintln(os.Stderr, "💡 Tip: Run 'r8s --help' to see all available commands")
	fmt.Fprintln(os.Stderr)
}

// findSimilarCommands finds commands similar to the input
func findSimilarCommands(input string) []SuggestedCommand {
	var similar []SuggestedCommand
	input = strings.ToLower(input)
	
	for _, cmd := range AvailableCommands {
		cmdName := strings.ToLower(cmd.Command)
		
		// Check for substring match
		if strings.Contains(cmdName, input) || strings.Contains(input, cmdName) {
			similar = append(similar, cmd)
			continue
		}
		
		// Check Levenshtein distance for typos (simplified)
		if levenshteinDistance(input, cmdName) <= 2 {
			similar = append(similar, cmd)
		}
	}
	
	// Limit to 3 suggestions
	if len(similar) > 3 {
		similar = similar[:3]
	}
	
	return similar
}

// levenshteinDistance calculates the edit distance between two strings
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create matrix
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
	}

	// Initialize first row and column
	for i := 0; i <= len(s1); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}
			
			deletion := matrix[i-1][j] + 1
			insertion := matrix[i][j-1] + 1
			substitution := matrix[i-1][j-1] + cost
			
			matrix[i][j] = min(deletion, min(insertion, substitution))
		}
	}

	return matrix[len(s1)][len(s2)]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// showAvailableCommands displays all available commands
func showAvailableCommands() {
	header := color.New(color.Bold, color.FgCyan)
	header.Fprintln(os.Stderr, "Available commands:")
	fmt.Fprintln(os.Stderr)
	
	for _, cmd := range AvailableCommands {
		cmdColor := color.New(color.Bold, color.FgGreen)
		descColor := color.New(color.FgWhite)
		
		cmdColor.Fprintf(os.Stderr, "  %-12s", cmd.Command)
		descColor.Fprintln(os.Stderr, cmd.Description)
	}
	fmt.Fprintln(os.Stderr)
}

// ShowFriendlyError displays a user-friendly error message
func ShowFriendlyError(err error) {
	if err == nil {
		return
	}
	
	fmt.Fprintln(os.Stderr)
	
	// Check for specific error types
	errMsg := err.Error()
	
	// Bundle path errors
	if strings.Contains(errMsg, "cannot access bundle path") {
		showBundlePathError(err)
		return
	}
	
	// File not found errors
	if strings.Contains(errMsg, "no such file") || strings.Contains(errMsg, "cannot find") {
		showFileNotFoundError(err)
		return
	}
	
	// Permission errors
	if strings.Contains(errMsg, "permission denied") {
		showPermissionError(err)
		return
	}
	
	// Default error display
	header := color.New(color.Bold, color.FgRed)
	header.Fprintln(os.Stderr, "✗ Error:")
	
	msgColor := color.New(color.FgWhite)
	msgColor.Fprintf(os.Stderr, "  %s\n", errMsg)
	fmt.Fprintln(os.Stderr)
	
	// Add helpful context based on error
	if strings.Contains(errMsg, "analyze") {
		tip := color.New(color.Italic, color.FgHiBlack)
		tip.Fprintln(os.Stderr, "💡 Tip: Make sure you're providing a valid bundle path")
		tip.Fprintln(os.Stderr, "   Example: r8s analyze ./path/to/extracted-bundle/")
	}
}

// showBundlePathError shows specific help for bundle path errors
func showBundlePathError(err error) {
	header := color.New(color.Bold, color.FgRed)
	header.Fprintln(os.Stderr, "🤠 Whoops! Can't find that bundle path")
	fmt.Fprintln(os.Stderr)
	
	msgColor := color.New(color.FgWhite)
	msgColor.Fprintf(os.Stderr, "  %s\n", err.Error())
	fmt.Fprintln(os.Stderr)
	
	tipColor := color.New(color.FgYellow)
	tipColor.Fprintln(os.Stderr, "Make sure you:")
	fmt.Fprintln(os.Stderr)
	
	infoColor := color.New(color.FgWhite)
	infoColor.Fprintln(os.Stderr, "  1. Extracted the support bundle tarball first")
	infoColor.Fprintln(os.Stderr, "  2. Provided the correct path to the extracted folder")
	infoColor.Fprintln(os.Stderr, "  3. Have read permissions on that directory")
	fmt.Fprintln(os.Stderr)
	
	exampleColor := color.New(color.FgHiBlack)
	exampleColor.Fprintln(os.Stderr, "Example:")
	exampleColor.Fprintln(os.Stderr, "  tar -xzf support-bundle.tar.gz")
	exampleColor.Fprintln(os.Stderr, "  r8s analyze ./support-bundle/")
	fmt.Fprintln(os.Stderr)
}

// showFileNotFoundError shows specific help for file not found errors
func showFileNotFoundError(err error) {
	header := color.New(color.Bold, color.FgYellow)
	header.Fprintln(os.Stderr, "📁 File not found")
	fmt.Fprintln(os.Stderr)
	
	msgColor := color.New(color.FgWhite)
	msgColor.Fprintf(os.Stderr, "  %s\n", err.Error())
	fmt.Fprintln(os.Stderr)
	
	tipColor := color.New(color.FgHiBlack)
	tipColor.Fprintln(os.Stderr, "💡 Double-check the file path and try again")
	fmt.Fprintln(os.Stderr)
}

// showPermissionError shows specific help for permission errors
func showPermissionError(err error) {
	header := color.New(color.Bold, color.FgRed)
	header.Fprintln(os.Stderr, "🚫 Permission denied")
	fmt.Fprintln(os.Stderr)
	
	msgColor := color.New(color.FgWhite)
	msgColor.Fprintf(os.Stderr, "  %s\n", err.Error())
	fmt.Fprintln(os.Stderr)
	
	tipColor := color.New(color.FgHiBlack)
	tipColor.Fprintln(os.Stderr, "💡 You may need to:")
	tipColor.Fprintln(os.Stderr, "   - Check file/directory permissions")
	tipColor.Fprintln(os.Stderr, "   - Run with appropriate user privileges")
	tipColor.Fprintln(os.Stderr, "   - Use sudo if accessing system directories")
	fmt.Fprintln(os.Stderr)
}

// SetUnknownCommandHandler configures Cobra to use our custom error handler
func SetUnknownCommandHandler(root *cobra.Command) {
	root.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		fmt.Fprintln(os.Stderr)
		header := color.New(color.Bold, color.FgRed)
		header.Fprintln(os.Stderr, "🤠 Invalid flag")
		fmt.Fprintln(os.Stderr)
		
		msgColor := color.New(color.FgWhite)
		msgColor.Fprintf(os.Stderr, "  %s\n", err.Error())
		fmt.Fprintln(os.Stderr)
		
		tipColor := color.New(color.FgHiBlack)
		tipColor.Fprintln(os.Stderr, "💡 Run 'r8s --help' to see available flags")
		fmt.Fprintln(os.Stderr)
		
		return err
	})
}

// ShowBundleNotFoundError displays a helpful error when bundle path doesn't exist
func ShowBundleNotFoundError(bundlePath string) {
	fmt.Fprintln(os.Stderr)

	// Header
	header := color.New(color.Bold, color.FgRed)
	header.Fprintf(os.Stderr, "Bundle not found: '%s'\n", bundlePath)

	// Suggestion
	suggestColor := color.New(color.FgYellow)
	suggestColor.Fprintln(os.Stderr, "Try: r8s analyze ./bundle/")
	fmt.Fprintln(os.Stderr)
}

// ShowUsageError displays usage hint when argument order is wrong
func ShowUsageError(command, correctUsage string) {
	fmt.Fprintln(os.Stderr)

	// Header
	header := color.New(color.Bold, color.FgRed)
	header.Fprintln(os.Stderr, "Wrong argument order")
	fmt.Fprintln(os.Stderr)

	// Suggestion
	suggestColor := color.New(color.FgYellow)
	suggestColor.Fprintf(os.Stderr, "Usage: %s\n", correctUsage)
	fmt.Fprintln(os.Stderr)
}

// ShowNoIssuesFound displays a friendly message when analysis finds no issues
func ShowNoIssuesFound(bundlePath string) {
	fmt.Fprintln(os.Stderr)

	// Header
	header := color.New(color.Bold, color.FgGreen)
	header.Fprintln(os.Stderr, "No issues found")
	fmt.Fprintln(os.Stderr)

	// Suggestion
	suggestColor := color.New(color.FgYellow)
	suggestColor.Fprintln(os.Stderr, "Try 'r8s ask' for natural language queries")
	fmt.Fprintln(os.Stderr)

	exampleColor := color.New(color.FgHiBlack)
	exampleColor.Fprintf(os.Stderr, "  r8s ask %s \"which pods are crashing?\"\n", bundlePath)
	fmt.Fprintln(os.Stderr)
}

// NewUsageError creates an error with usage hint
func NewUsageError(command, correctUsage string) error {
	return &UsageError{Command: command, CorrectUsage: correctUsage}
}

// UsageError represents an error with incorrect argument usage
type UsageError struct {
	Command      string
	CorrectUsage string
}

func (e *UsageError) Error() string {
	return fmt.Sprintf("wrong argument order for command '%s'", e.Command)
}

// IsUsageError checks if an error is a UsageError
func IsUsageError(err error) bool {
	_, ok := err.(*UsageError)
	return ok
}
