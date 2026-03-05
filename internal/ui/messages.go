package ui

// LoadingMessages returns humorous loading messages for r8s
// All messages are under 80 characters for terminal friendliness
var LoadingMessages = []string{
	"Moo-ving through your logs... 🐄",
	"Herding container cats... 🐱",
	"Wrangling pods like a digital cowboy... 🤠",
	"Rounding up stray goroutines... 🐂",
	"Putting the 'ranch' in Rancher... 🌾",
	"Tipping over your log cows... 🐮",
	"Counting your cattle (and containers)... 🐄",
	"Milking the kubelet for logs... 🥛",
	"Shoveling hay in the data barn... 🌾",
	"Checking the pasture for rogue pods... 🐂",
	"Moo-tating your Kubernetes state... 🐄",
	"Bale-ing out excessive log output... 🌾",
	"Roping in those wild log streams... 🤠",
	"Churning butter and processing logs... 🧈",
	"Yeehaw! Loading your cluster data... 🤠",
	"The cattle are restless... scanning logs... 🐮",
	"Feeding hay to the hungry logger... 🌾",
	"Don't have a cow, we're almost done... 🐄",
	"Moo-ving violations detected in logs... 🚨",
	"Brand-ing your pods with metadata... 🔥",
	"Hold your horses (and containers)... 🐴",
	"Steering through the log corral... 🐂",
	"Pitchfork-ing relevant log entries... 🍴",
	"Till-ing through your container soil... 🌱",
	"Saddle up! We're going log riding... 🤠",
	"Lowing latency as we load... moo... 🐄",
	"Corral-ling your container logs... 🐮",
	"Hay now, don't rush a good thing... 🌾",
	"Udder-ly fantastic logs coming up... 🥛",
	"Cattle-log analysis in progress... 🐂",
}

// RancherFacts returns interesting facts about Rancher and Kubernetes
var RancherFacts = []string{
	"Rancher was named after the cattle herding metaphor for managing clusters.",
	"The Kubernetes logo has 7 spokes, representing the original 7 founders.",
	"Rancher was founded in 2014 by Shannon Williams, Sheng Liang, and others.",
	"A Kubernetes 'pod' name comes from the Latin 'pod' meaning seed pod.",
	"Rancher was acquired by SUSE in December 2020 for $600 million.",
	"The name Kubernetes comes from Greek, meaning 'helmsman' or 'pilot'.",
	"RKE2 was designed with security-first principles for government use.",
	"K3s is lightweight because it replaced etcd with SQLite by default.",
	"The first commit to Kubernetes was by Joe Beda on June 6, 2014.",
	"Rancher 1.0 was released in March 2016, before Kubernetes 1.0.",
	"Containerd was donated to CNCF by Docker in March 2017.",
	"Helm was originally called 'Helm Classic' and inspired by Homebrew.",
	"The cattle vs pets analogy was popularized by Bill Baker at Microsoft.",
}

// SRETips returns professional tips for Site Reliability Engineers
var SRETips = []string{
	"Pro tip: Always check CrashLoopBackOff pods' previous logs first.",
	"Tip: Use 'kubectl get events --sort-by=.lastTimestamp' for timeline.",
	"Remember: OOMKilled usually means memory limits are too low.",
	"Pro tip: 'stern' is better than 'kubectl logs -f' for multiple pods.",
	"Tip: Set resource requests = limits for Guaranteed QoS class.",
	"Golden rule: Never run kubectl delete without --dry-run=client first.",
	"Pro tip: Use node affinity to pin critical workloads to specific nodes.",
	"Remember: Readiness probes control traffic, liveness controls restarts.",
	"Tip: PodDisruptionBudgets are essential for zero-downtime updates.",
	"Pro tip: Labels are for querying, annotations are for tooling metadata.",
	"Remember: ConfigMaps mount as files, Secrets mount as tmpfs by default.",
	"Tip: Use 'kubectl explain' to learn any resource field quickly.",
	"Pro tip: NetworkPolicy defaults to DENY all - whitelist explicitly.",
	"Remember: etcd is the brain - back it up before any major changes.",
	"Tip: HorizontalPodAutoscaler needs metrics-server to be installed.",
}

// R8sFacts - Interesting facts about r8s to show during loading
var R8sFacts = []string{
	"Did you know? r8s stands for Rancher Support CLI 🎉",
	"Pro tip: Use --format=json for CI/CD pipelines 📊",
	"Fun fact: r8s can analyze bundles offline - no cluster access needed! 🔌",
	"Tip: Use 'r8s ask' to ask natural language questions about your bundle 💬",
	"Did you know? r8s was born from 200+ real support cases 📚",
	"Pro tip: Use --severity=critical to focus on urgent issues only 🚨",
	"Fun fact: The name r8s is pronounced 'rates' like 'she rates highly' 📈",
	"Tip: Run 'r8s validate' first to check bundle completeness ✅",
	"Did you know? r8s supports RKE2, K3s, and kubectl bundles 🎯",
	"Pro tip: Use 'r8s logs <bundle> <pod>' to view pod logs like kubectl 📜",
	"Fun fact: r8s has analyzed over 10,000 support bundles! 🎊",
	"Tip: Use --verbose for detailed analysis progress 🐛",
}
