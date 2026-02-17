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

func TestSeverityString(t *testing.T) {
	tests := []struct {
		input    AttentionSeverity
		expected string
	}{
		{SeverityCritical, "Critical"},
		{SeverityWarning, "Warning"},
		{SeverityInfo, "Info"},
		{AttentionSeverity(99), "Unknown"},
	}

	for _, tt := range tests {
		result := severityString(tt.input)
		if result != tt.expected {
			t.Errorf("severityString(%d) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}
