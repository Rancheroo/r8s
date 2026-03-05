package ui

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
