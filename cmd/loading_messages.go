package cmd

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/fatih/color"
)

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

// LoadingDisplay manages the loading UI
type LoadingDisplay struct {
	startTime     time.Time
	currentFile   string
	showProgress  bool
	lastFactShown time.Time
}

// NewLoadingDisplay creates a new loading display
func NewLoadingDisplay(showProgress bool) *LoadingDisplay {
	return &LoadingDisplay{
		startTime:     time.Now(),
		showProgress:  showProgress,
		lastFactShown: time.Now(),
	}
}

// ShowRandomLoadingMessage prints a random fun loading message
func (ld *LoadingDisplay) ShowRandomLoadingMessage() {
	msg := LoadingMessages[rand.Intn(len(LoadingMessages))]
	showFunMessage(msg)
}

// ShowSpecificLoadingMessage shows a specific loading message by index
func (ld *LoadingDisplay) ShowSpecificLoadingMessage(index int) {
	if index >= 0 && index < len(LoadingMessages) {
		showFunMessage(LoadingMessages[index])
	}
}

// showFunMessage displays a message with fun formatting
func showFunMessage(message string) {
	border := color.New(color.FgYellow).Sprint("< ") +
		color.New(color.FgCyan).Sprint(message) +
		color.New(color.FgYellow).Sprint(" >")

	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, border)
	fmt.Fprintln(os.Stderr)
}

// ShowFact shows a random r8s fact if enough time has passed
func (ld *LoadingDisplay) ShowFact() {
	if time.Since(ld.lastFactShown) < 5*time.Second {
		return
	}

	fact := R8sFacts[rand.Intn(len(R8sFacts))]
	c := color.New(color.Italic, color.FgHiBlack)
	c.Fprintln(os.Stderr, "   💡 "+fact)
	ld.lastFactShown = time.Now()
}

// ShowFactAlways always shows a random fact
func (ld *LoadingDisplay) ShowFactAlways() {
	fact := R8sFacts[rand.Intn(len(R8sFacts))]
	c := color.New(color.Italic, color.FgHiBlack)
	c.Fprintln(os.Stderr, "   💡 "+fact)
}

// UpdateCurrentFile updates the currently processing file
func (ld *LoadingDisplay) UpdateCurrentFile(file string) {
	ld.currentFile = file
	if ld.showProgress {
		fmt.Fprintf(os.Stderr, "   📄 %s\r", truncateString(file, 50))
	}
}

// ShowElapsedTime displays elapsed time
func (ld *LoadingDisplay) ShowElapsedTime() {
	elapsed := time.Since(ld.startTime)
	c := color.New(color.FgHiBlack)
	c.Fprintf(os.Stderr, "   ⏱️  Elapsed: %s\n", formatDuration(elapsed))
}

// ShowCompletionMessage displays a fun completion message
func (ld *LoadingDisplay) ShowCompletionMessage() {
	elapsed := time.Since(ld.startTime)
	fmt.Fprintln(os.Stderr)

	successColor := color.New(color.Bold, color.FgGreen)
	if elapsed < 2*time.Second {
		successColor.Fprintln(os.Stderr, "   ✨ That was faster than a jackrabbit! Analysis complete!")
	} else if elapsed < 10*time.Second {
		successColor.Fprintln(os.Stderr, "   🎯 Analysis lassoed and complete!")
	} else {
		successColor.Fprintln(os.Stderr, "   🏆 Whew! That was a big bundle. Analysis complete!")
	}
	fmt.Fprintln(os.Stderr)
}

// ShowStartMessage displays the initial loading message
func (ld *LoadingDisplay) ShowStartMessage(bundlePath string) {
	fmt.Fprintln(os.Stderr)
	header := color.New(color.Bold, color.FgCyan)
	header.Fprintln(os.Stderr, "┌─────────────────────────────────────────┐")
	header.Fprintln(os.Stderr, "│     🤠 R8S Bundle Analysis Starting     │")
	header.Fprintln(os.Stderr, "└─────────────────────────────────────────┘")
	fmt.Fprintf(os.Stderr, "   📁 Bundle: %s\n", truncateString(bundlePath, 45))
	fmt.Fprintln(os.Stderr)
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
}

// ShowFileProgress shows progress for a specific file
func ShowFileProgress(fileName string, current, total int) {
	percent := float64(current) * 100 / float64(total)
	barWidth := 30
	filled := int(percent / 100 * float64(barWidth))

	bar := color.New(color.FgGreen).Sprint("█")
	empty := color.New(color.FgHiBlack).Sprint("░")

	progressBar := ""
	for i := 0; i < barWidth; i++ {
		if i < filled {
			progressBar += bar
		} else {
			progressBar += empty
		}
	}

	c := color.New(color.FgHiBlack)
	c.Fprintf(os.Stderr, "   %s %3.0f%% (%d/%d) %s\r",
		progressBar, percent, current, total, fileName)
}

// GetRandomLoadingMessage returns a random loading message
func GetRandomLoadingMessage() string {
	return LoadingMessages[rand.Intn(len(LoadingMessages))]
}

// GetRandomR8sFact returns a random r8s fact
func GetRandomR8sFact() string {
	return R8sFacts[rand.Intn(len(R8sFacts))]
}

// GetRandomSRETip returns a random SRE tip
func GetRandomSRETip() string {
	return SRETips[rand.Intn(len(SRETips))]
}

// GetRandomRancherFact returns a random Rancher fact
func GetRandomRancherFact() string {
	return RancherFacts[rand.Intn(len(RancherFacts))]
}
