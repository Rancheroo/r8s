package ui

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

// SpinnerFrames - Classic spinning animation frames
var SpinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// RancherSpinnerFrames - Fun Rancher-themed spinner frames
var RancherSpinnerFrames = []string{"🐄", "🤠", "🐮", "🐕", "🌾", "🐴", "🌵", "⭐"}

// Spinner provides animated loading indication
type Spinner struct {
	frames      []string
	current     int
	message     string
	stopChan    chan struct{}
	stoppedChan chan struct{}
	mu          sync.Mutex
	running     bool
}

// NewSpinner creates a classic spinner
func NewSpinner(message string) *Spinner {
	return &Spinner{
		frames:      SpinnerFrames,
		message:     message,
		stopChan:    make(chan struct{}),
		stoppedChan: make(chan struct{}),
	}
}

// NewRancherSpinner creates a fun Rancher-themed spinner
func NewRancherSpinner(message string) *Spinner {
	return &Spinner{
		frames:      RancherSpinnerFrames,
		message:     message,
		stopChan:    make(chan struct{}),
		stoppedChan: make(chan struct{}),
	}
}

// Start begins the spinner animation
func (s *Spinner) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	// Initialize fresh channels for restart
	s.stopChan = make(chan struct{})
	s.stoppedChan = make(chan struct{})
	s.running = true
	s.mu.Unlock()

	go s.run()
}

// Stop ends the spinner animation
func (s *Spinner) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	close(s.stopChan)
	<-s.stoppedChan
}

// UpdateMessage changes the spinner message
func (s *Spinner) UpdateMessage(message string) {
	s.mu.Lock()
	s.message = message
	s.mu.Unlock()
}

// run is the main spinner loop
func (s *Spinner) run() {
	defer close(s.stoppedChan)

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			// Clear the line
			fmt.Fprint(os.Stderr, "\r\033[K")
			return
		case <-ticker.C:
			s.mu.Lock()
			frame := s.frames[s.current]
			msg := s.message
			s.current = (s.current + 1) % len(s.frames)
			s.mu.Unlock()

			c := color.New(color.FgCyan)
			c.Fprintf(os.Stderr, "\r   %s %s", frame, msg)
		}
	}
}

// ProgressBar represents a visual progress bar
type ProgressBar struct {
	width   int
	total   int
	current int
	prefix  string
	suffix  string
}

// NewProgressBar creates a new progress bar
func NewProgressBar(total int, prefix string) *ProgressBar {
	return &ProgressBar{
		width:  40,
		total:  total,
		prefix: prefix,
	}
}

// Update updates the progress bar
func (pb *ProgressBar) Update(current int, suffix string) {
	pb.current = current
	pb.suffix = suffix
	pb.render()
}

// Increment increments the progress
func (pb *ProgressBar) Increment(suffix string) {
	pb.current++
	pb.suffix = suffix
	pb.render()
}

// render draws the progress bar
func (pb *ProgressBar) render() {
	if pb.total <= 0 {
		return
	}
	percent := float64(pb.current) * 100 / float64(pb.total)
	filled := int(percent / 100 * float64(pb.width))

	barColor := color.New(color.FgGreen)
	emptyColor := color.New(color.FgHiBlack)
	textColor := color.New(color.FgCyan)

	// Build bar using strings.Builder for efficiency
	var builder strings.Builder
	builder.Grow(pb.width * 10) // Pre-allocate buffer

	for i := 0; i < pb.width; i++ {
		if i < filled {
			builder.WriteString(barColor.Sprint("█"))
		} else {
			builder.WriteString(emptyColor.Sprint("░"))
		}
	}

	// Format suffix
	suffixStr := ""
	if pb.suffix != "" {
		suffixStr = " " + pb.suffix
	}

	textColor.Fprintf(os.Stderr, "\r   %s [%s] %3.0f%%%s",
		pb.prefix, builder.String(), percent, suffixStr)
}

// Finish marks the progress as complete
func (pb *ProgressBar) Finish(message string) {
	pb.current = pb.total
	success := color.New(color.FgGreen)
	fullBar := repeatChar("█", pb.width)
	fmt.Fprintf(os.Stderr, "\r   %s [%s] %s\n",
		pb.prefix, success.Sprint(fullBar), message)
}

// repeatChar repeats a character n times
func repeatChar(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}

// SimpleProgressTracker tracks progress without animation
type SimpleProgressTracker struct {
	total       int
	current     int
	lastUpdate  time.Time
	updateDelay time.Duration
}

// NewSimpleProgressTracker creates a simple progress tracker
func NewSimpleProgressTracker(total int) *SimpleProgressTracker {
	return &SimpleProgressTracker{
		total:       total,
		lastUpdate:  time.Now(),
		updateDelay: 500 * time.Millisecond,
	}
}

// Increment updates progress with throttling
func (spt *SimpleProgressTracker) Increment(message string) bool {
	spt.current++

	if time.Since(spt.lastUpdate) < spt.updateDelay {
		return false
	}

	spt.lastUpdate = time.Now()

	percent := float64(spt.current) * 100 / float64(spt.total)
	textColor := color.New(color.FgHiBlack)
	textColor.Fprintf(os.Stderr, "   📊 %3.0f%% (%d/%d) - %s\r",
		percent, spt.current, spt.total, message)

	return true
}

// Complete finishes the progress tracking
func (spt *SimpleProgressTracker) Complete(message string) {
	fmt.Fprintln(os.Stderr)
	if message != "" {
		success := color.New(color.FgGreen)
		success.Fprintf(os.Stderr, "   ✓ %s\n", message)
	}
}

// ClearLine clears the current terminal line
func ClearLine() {
	fmt.Fprint(os.Stderr, "\r\033[K")
}

// ShowStep shows a step in a multi-step process
func ShowStep(stepNum, totalSteps int, message string) {
	stepColor := color.New(color.Bold, color.FgYellow)
	msgColor := color.New(color.FgCyan)

	stepColor.Fprintf(os.Stderr, "   [%d/%d] ", stepNum, totalSteps)
	msgColor.Fprintln(os.Stderr, message)
}

// ShowSuccess shows a success message
func ShowSuccess(message string) {
	success := color.New(color.Bold, color.FgGreen)
	success.Fprintf(os.Stderr, "   ✓ %s\n", message)
}

// ShowWarning shows a warning message
func ShowWarning(message string) {
	warning := color.New(color.Bold, color.FgYellow)
	warning.Fprintf(os.Stderr, "   ⚠️  %s\n", message)
}

// ShowError shows an error message
func ShowError(message string) {
	errColor := color.New(color.Bold, color.FgRed)
	errColor.Fprintf(os.Stderr, "   ✗ %s\n", message)
}

// ShowInfo shows an info message
func ShowInfo(message string) {
	info := color.New(color.FgHiBlack)
	info.Fprintf(os.Stderr, "   ℹ️  %s\n", message)
}
