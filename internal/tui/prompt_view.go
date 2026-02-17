package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// PromptView handles displaying and exporting AI prompts
type PromptView struct {
	content    string
	scrollY    int
	height     int
	width      int
	visible    bool
	promptType string // "support", "finding", "comparison"
}

// NewPromptView creates a new prompt view
func NewPromptView() *PromptView {
	return &PromptView{
		content: "",
		scrollY: 0,
		visible: false,
	}
}

// Show displays the prompt view with content
func (pv *PromptView) Show(content string, promptType string) {
	pv.content = content
	pv.promptType = promptType
	pv.visible = true
	pv.scrollY = 0
}

// Hide hides the prompt view
func (pv *PromptView) Hide() {
	pv.visible = false
}

// IsVisible returns true if the prompt view is showing
func (pv *PromptView) IsVisible() bool {
	return pv.visible
}

// ScrollDown scrolls the view down
func (pv *PromptView) ScrollDown() {
	lines := strings.Count(pv.content, "\n")
	maxScroll := lines - pv.height + 5
	if pv.scrollY < maxScroll {
		pv.scrollY++
	}
}

// ScrollUp scrolls the view up
func (pv *PromptView) ScrollUp() {
	if pv.scrollY > 0 {
		pv.scrollY--
	}
}

// SetSize updates the view dimensions
func (pv *PromptView) SetSize(width, height int) {
	pv.width = width
	pv.height = height
}

// GetContent returns the raw content (for clipboard copy)
func (pv *PromptView) GetContent() string {
	return pv.content
}

// Render draws the prompt view
func (pv *PromptView) Render() string {
	if !pv.visible {
		return ""
	}

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorCyan).
		Background(colorDarkGray).
		Padding(0, 2).
		Width(pv.width)

	var title string
	switch pv.promptType {
	case "support":
		title = "🤖 AI Support Prompt - Full Analysis"
	case "finding":
		title = "🤖 AI Prompt - Single Issue"
	case "comparison":
		title = "🤖 AI Prompt - Health Comparison"
	default:
		title = "🤖 AI Prompt"
	}

	// Content area with scrolling
	contentStyle := lipgloss.NewStyle().
		Foreground(colorWhite).
		Padding(1, 2).
		Width(pv.width - 4).
		Height(pv.height - 6)

	// Split content into lines and scroll
	lines := strings.Split(pv.content, "\n")
	startLine := pv.scrollY
	endLine := startLine + pv.height - 6
	if endLine > len(lines) {
		endLine = len(lines)
	}
	if startLine < len(lines) {
		lines = lines[startLine:endLine]
	}
	visibleContent := strings.Join(lines, "\n")

	// Help footer
	footerStyle := lipgloss.NewStyle().
		Foreground(colorGray).
		Background(colorDarkGray).
		Padding(0, 2).
		Width(pv.width)

	footer := "[↑/↓] Scroll | [c] Copy to clipboard | [q/Esc] Close"

	// Assemble view
	var b strings.Builder
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n")
	b.WriteString(contentStyle.Render(visibleContent))
	b.WriteString("\n")
	b.WriteString(footerStyle.Render(footer))

	return b.String()
}

// GetCopyInstructions returns platform-specific copy instructions
func GetCopyInstructions() string {
	return `To copy this prompt:

Linux/macOS:
  1. Select all text (Ctrl+A or Cmd+A)
  2. Copy (Ctrl+C or Cmd+C)
  3. Paste into your AI chatbot

Or use the [c] key to auto-copy (if xclip/xsel available)

Recommended AI Tools:
  • Grok (grok.com) - Great for technical analysis
  • ChatGPT (chat.openai.com) - Good for step-by-step guides
  • Claude (claude.ai) - Excellent for complex debugging`
}
