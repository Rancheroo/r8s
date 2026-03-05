package ui

// LoadingMessages returns professional loading messages
var LoadingMessages = []string{
	"Analyzing bundle contents...",
	"Parsing container logs...",
	"Scanning for error patterns...",
	"Validating cluster state...",
	"Checking resource limits...",
	"Correlating system events...",
	"Examining control plane health...",
	"Processing Kubernetes objects...",
	"Identifying root causes...",
	"Generating analysis report...",
}

// R8sFacts contains useful tips for automation and advanced usage
var R8sFacts = []string{
	// CI/CD & Automation
	"Use 'r8s export --format=sarif' for CI/CD pipeline integration",
	"r8s outputs JSON - perfect for automation scripts: r8s analyze bundle/ --format=json | jq '.issues'",
	"Integrate r8s into your GitHub Actions for automated bundle analysis",
	"Use r8s in pre-commit hooks to validate cluster state before deployments",
	"r8s SARIF output integrates with GitHub Advanced Security and other SAST tools",
	"Automate nightly bundle analysis with cron: r8s analyze /var/log/bundles/ --format=json",
	"Export to Markdown for inclusion in incident reports: r8s export bundle/ --format=markdown",
	"Use r8s test-cluster in your deployment pipeline for pre-flight checks",
	"r8s exit codes (0=healthy, 1=issues, 2=error) enable script automation",
	"Chain r8s commands: r8s analyze bundle/ && r8s export bundle/ --format=sarif",

	// AI & Natural Language
	"Ask natural language questions: r8s ask bundle/ 'why is my pod crashing?'",
	"r8s integrates with Claude Code for AI-powered troubleshooting workflows",
	"Use r8s ask to get root cause analysis without knowing kubectl commands",
	"Ask comparative questions: r8s ask bundle/ 'what changed between these events?'",
	"r8s AI patterns detect issues you might miss in manual log review",
	"Combine r8s ask with your knowledge base for instant troubleshooting runbooks",
	"Use r8s ask to onboard new team members - no Kubernetes expertise required",
	"r8s patterns learn from your bundles - the more you use it, the smarter it gets",
	"Ask r8s to summarize complex issues: r8s ask bundle/ 'explain the etcd problems'",
	"Use r8s ask for post-mortems: r8s ask bundle/ 'what caused the outage?'",

	// Familiar Syntax
	"r8s commands mirror kubectl: get pods, get nodes, logs, describe",
	"r8s supports familiar output: -o json, -o yaml, -o wide",
	"Filter by namespace: r8s get pods bundle/ -n cattle-system",
	"r8s logs works like you'd expect: r8s logs bundle/ pod-name",
	"Use r8s describe for detailed resource info: r8s describe pod bundle/ nginx",
	"r8s get nodes shows cluster topology from bundle",
	"Use labels with r8s get: r8s get pods bundle/ -l app=nginx",
	"r8s works with standard Unix pipes: r8s analyze bundle/ | grep ERROR",
	"Create aliases for common tasks: alias r8s-check='r8s test-cluster'",
	"r8s supports all standard output formats for integration",

	// Debugging & Troubleshooting
	"r8s test-cluster runs 7 diagnostic checks automatically",
	"Check bundle completeness: r8s validate bundle/ (should be >90%)",
	"Look for patterns: r8s analyze bundle/ | grep -i 'crash\\|oom\\|error'",
	"Use r8s offline - analyze production issues without cluster access",
	"r8s detects OOM kills, crash loops, image pull failures automatically",
	"Compare bundles: analyze old/ and new/ to spot what changed",
	"r8s finds issues in etcd, networking, storage, and workloads",
	"Use r8s validate to ensure complete bundle collection",
	"r8s patterns identify root causes, not just symptoms",
	"Export analysis for team review: r8s export bundle/ --format=markdown",

	// Swiss Army Knife Features
	"r8s is your Swiss Army knife for Rancher log automation",
	"One tool for analysis, export, querying, and validation",
	"r8s works with RKE2, K3s, and standard Kubernetes bundles",
	"Use r8s in air-gapped environments - no internet required",
	"r8s bundles can be archived for compliance and auditing",
	"Share r8s analysis: export to Markdown and attach to tickets",
	"r8s supports both interactive and scripted workflows",
	"Master r8s once, use it for any Kubernetes distribution",
	"r8s scales from single-node to multi-cluster analysis",
	"The complete toolkit: analyze, ask, export, get, logs, validate, test-cluster",
}

// RancherFacts - Empty to prefer R8sFacts
var RancherFacts = []string{}

// SRETips - Empty to prefer R8sFacts
var SRETips = []string{}
