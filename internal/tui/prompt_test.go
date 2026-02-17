package tui

import (
	"strings"
	"testing"

	"github.com/Rancheroo/r8s/internal/bundle"
)

func TestPromptGenerator_GenerateSupportPrompt(t *testing.T) {
	health := &bundle.BundleHealth{
		BundleType: "rke2-support-bundle",
		TotalFiles: 10,
		FoundFiles: 8,
		Warnings:   []string{"Missing system logs"},
	}

	findings := []AttentionItem{
		{
			Severity:     SeverityCritical,
			Title:        "OOM Kill",
			Description:  "Container exceeded memory limit",
			Namespace:    "prod",
			PodName:      "memory-hog",
			ResourceType: "pod",
		},
		{
			Severity:     SeverityWarning,
			Title:        "Image Pull Failed",
			Description:  "Back-off pulling image",
			Namespace:    "default",
			ResourceType: "pod",
		},
	}

	pg := NewPromptGenerator(health, findings, "/path/to/bundle.tar.gz")
	prompt := pg.GenerateSupportPrompt()

	// Check key sections exist
	if !strings.Contains(prompt, "R8S Support Bundle Analysis") {
		t.Error("Missing title")
	}
	if !strings.Contains(prompt, "bundle.tar.gz") {
		t.Error("Missing bundle path")
	}
	if !strings.Contains(prompt, "OOM Kill") {
		t.Error("Missing critical finding")
	}
	if !strings.Contains(prompt, "1 Critical, 1 Warning") {
		t.Error("Missing summary count")
	}
	if !strings.Contains(prompt, "Step-by-Step Remediation") {
		t.Error("Missing request section")
	}
}

func TestPromptGenerator_GenerateFindingPrompt(t *testing.T) {
	health := &bundle.BundleHealth{
		BundleType: "rke2-support-bundle",
	}

	item := AttentionItem{
		Severity:      SeverityCritical,
		Title:         "OOM Kill",
		Description:   "Exceeded limit: 1Gi",
		Namespace:     "prod",
		PodName:       "app-pod",
		ContainerName: "main",
		ResourceType:  "pod",
	}

	pg := NewPromptGenerator(health, []AttentionItem{item}, "bundle.tar.gz")
	prompt := pg.GenerateFindingPrompt(item)

	if !strings.Contains(prompt, "OOM Kill") {
		t.Error("Missing title")
	}
	if !strings.Contains(prompt, "prod") {
		t.Error("Missing namespace")
	}
	if !strings.Contains(prompt, "app-pod") {
		t.Error("Missing pod name")
	}
	if !strings.Contains(prompt, "Immediate Fix") {
		t.Error("Missing request section")
	}
}

func TestPromptView(t *testing.T) {
	pv := NewPromptView()

	if pv.IsVisible() {
		t.Error("New view should not be visible")
	}

	pv.Show("test content", "support")
	if !pv.IsVisible() {
		t.Error("View should be visible after Show()")
	}

	if pv.GetContent() != "test content" {
		t.Error("Content not stored correctly")
	}

	pv.Hide()
	if pv.IsVisible() {
		t.Error("View should not be visible after Hide()")
	}
}

func TestPromptGenerator_GenerateTerminalPrompt(t *testing.T) {
	health := &bundle.BundleHealth{
		BundleType: "rke2-support-bundle",
	}

	findings := []AttentionItem{
		{
			Severity:     SeverityCritical,
			Title:        "OOM Kill",
			Description:  "Container exceeded memory limit",
			Namespace:    "prod",
			PodName:      "memory-hog",
			ResourceType: "pod",
		},
	}

	pg := NewPromptGenerator(health, findings, "/path/to/bundle.tar.gz")
	prompt := pg.GenerateTerminalPrompt()

	// Check for terminal-specific content
	if !strings.Contains(prompt, "R8S Terminal AI Analysis") {
		t.Error("Missing terminal prompt title")
	}
	if !strings.Contains(prompt, "kubectl") {
		t.Error("Missing kubectl references")
	}
	if !strings.Contains(prompt, "Diagnostic command") {
		t.Error("Missing diagnostic command section")
	}
	if !strings.Contains(prompt, "Fix command") {
		t.Error("Missing fix command section")
	}
}

func TestPromptGenerator_GetPromptForType(t *testing.T) {
	health := &bundle.BundleHealth{}
	findings := []AttentionItem{
		{Severity: SeverityCritical, Title: "Test Issue", Description: "Test"},
	}

	pg := NewPromptGenerator(health, findings, "bundle.tar.gz")

	chatbotPrompt := pg.GetPromptForType(PromptTypeChatbot)
	if !strings.Contains(chatbotPrompt, "Support Bundle Analysis") {
		t.Error("Chatbot prompt should contain support bundle analysis")
	}

	terminalPrompt := pg.GetPromptForType(PromptTypeTerminal)
	if !strings.Contains(terminalPrompt, "Terminal AI Analysis") {
		t.Error("Terminal prompt should contain terminal AI analysis")
	}
}

func TestPromptGenerator_getPrimaryNamespace(t *testing.T) {
	findings := []AttentionItem{
		{Namespace: "prod"},
		{Namespace: "prod"},
		{Namespace: "default"},
	}

	pg := NewPromptGenerator(nil, findings, "")
	ns := pg.getPrimaryNamespace()

	if ns != "prod" {
		t.Errorf("Expected primary namespace 'prod', got '%s'", ns)
	}
}

func TestGetPromptTypeDescription(t *testing.T) {
	chatbotDesc := GetPromptTypeDescription(PromptTypeChatbot)
	if !strings.Contains(chatbotDesc, "Chatbot") {
		t.Error("Chatbot description should contain 'Chatbot'")
	}

	terminalDesc := GetPromptTypeDescription(PromptTypeTerminal)
	if !strings.Contains(terminalDesc, "Terminal") {
		t.Error("Terminal description should contain 'Terminal'")
	}
}

func TestListAvailablePromptTypes(t *testing.T) {
	types := ListAvailablePromptTypes()
	if len(types) != 2 {
		t.Errorf("Expected 2 prompt types, got %d", len(types))
	}

	hasChatbot := false
	hasTerminal := false
	for _, t := range types {
		if t == PromptTypeChatbot {
			hasChatbot = true
		}
		if t == PromptTypeTerminal {
			hasTerminal = true
		}
	}

	if !hasChatbot {
		t.Error("Should have Chatbot prompt type")
	}
	if !hasTerminal {
		t.Error("Should have Terminal prompt type")
	}
}
