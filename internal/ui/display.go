package ui

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/fatih/color"
)

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
	if len(LoadingMessages) == 0 {
		return
	}
	msg := LoadingMessages[rand.Intn(len(LoadingMessages))]
	showFunMessage(msg)
}

// ShowSpecificLoadingMessage shows a specific loading message by index
func (ld *LoadingDisplay) ShowSpecificLoadingMessage(index int) {
	if len(LoadingMessages) == 0 {
		return
	}
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
	if len(R8sFacts) == 0 {
		return
	}

	fact := R8sFacts[rand.Intn(len(R8sFacts))]
	c := color.New(color.Italic, color.FgHiBlack)
	c.Fprintln(os.Stderr, "   💡 "+fact)
	ld.lastFactShown = time.Now()
}

// ShowFactAlways always shows a random fact
func (ld *LoadingDisplay) ShowFactAlways() {
	if len(R8sFacts) == 0 {
		return
	}
	fact := R8sFacts[rand.Intn(len(R8sFacts))]
	c := color.New(color.Italic, color.FgHiBlack)
	c.Fprintln(os.Stderr, "   💡 "+fact)
}

// UpdateCurrentFile updates the currently processing file
func (ld *LoadingDisplay) UpdateCurrentFile(file string) {
	ld.currentFile = file
	if ld.showProgress {
		fmt.Fprintf(os.Stderr, "   📄 %s\r", TruncateString(file, 50))
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
	fmt.Fprintf(os.Stderr, "   📁 Bundle: %s\n", TruncateString(bundlePath, 45))
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
	if total <= 0 {
		return
	}
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
	if len(LoadingMessages) == 0 {
		return ""
	}
	return LoadingMessages[rand.Intn(len(LoadingMessages))]
}

// GetRandomR8sFact returns a random r8s fact
func GetRandomR8sFact() string {
	if len(R8sFacts) == 0 {
		return ""
	}
	return R8sFacts[rand.Intn(len(R8sFacts))]
}

// GetRandomSRETip returns a random SRE tip
func GetRandomSRETip() string {
	if len(SRETips) == 0 {
		return ""
	}
	return SRETips[rand.Intn(len(SRETips))]
}

// GetRandomRancherFact returns a random Rancher fact
func GetRandomRancherFact() string {
	if len(RancherFacts) == 0 {
		return ""
	}
	return RancherFacts[rand.Intn(len(RancherFacts))]
}

// ShowRandomTip prints a random tip
func ShowRandomTip() {
	if len(R8sFacts) == 0 {
		return
	}
	rand.Seed(time.Now().UnixNano())
	tip := R8sFacts[rand.Intn(len(R8sFacts))]
	tipColor := color.New(color.Italic, color.FgCyan)
	tipColor.Fprintln(os.Stderr, "💡 "+tip)
}

// TruncateString truncates a string to max length
func TruncateString(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
