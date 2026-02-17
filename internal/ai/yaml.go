// Package ai provides pattern loading from YAML files.
// Sprint 8: Custom pattern loading via YAML definitions.
package ai

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// YAMLLoader loads patterns from YAML files
type YAMLLoader struct {
	registry *PatternRegistry
}

// NewYAMLLoader creates a new YAML pattern loader
func NewYAMLLoader(registry *PatternRegistry) *YAMLLoader {
	return &YAMLLoader{registry: registry}
}

// LoadDirectory loads all YAML patterns from a directory
func (yl *YAMLLoader) LoadDirectory(dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("cannot read directory %s: %w", dir, err)
	}
	
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		
		// Only process .yaml files
		if !strings.HasSuffix(file.Name(), ".yaml") {
			continue
		}
		
		path := filepath.Join(dir, file.Name())
		if err := yl.LoadFile(path); err != nil {
			return fmt.Errorf("failed to load %s: %w", path, err)
		}
	}
	
	return nil
}

// LoadFile loads a single YAML pattern file
func (yl *YAMLLoader) LoadFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot read file: %w", err)
	}
	
	var pattern Pattern
	if err := yaml.Unmarshal(data, &pattern); err != nil {
		return fmt.Errorf("cannot parse YAML: %w", err)
	}
	
	if err := yl.registry.Register(pattern); err != nil {
		return fmt.Errorf("invalid pattern: %w", err)
	}
	
	return nil
}

// PatternDefinition is the YAML structure for pattern files
type PatternDefinition struct {
	ID          string   `yaml:"id"`
	Name        string   `yaml:"name"`
	Category    string   `yaml:"category"`
	Severity    string   `yaml:"severity"`
	Keywords    []string `yaml:"keywords"`
	Description string   `yaml:"description"`
	Suggestion  string   `yaml:"suggestion"`
}

// Validate checks if a pattern definition is valid
func (pd *PatternDefinition) Validate() error {
	if pd.ID == "" {
		return fmt.Errorf("pattern ID is required")
	}
	if pd.Name == "" {
		return fmt.Errorf("pattern name is required")
	}
	if len(pd.Keywords) == 0 {
		return fmt.Errorf("at least one keyword is required")
	}
	if pd.Severity == "" {
		pd.Severity = "warning" // Default
	}
	return nil
}

// ToPattern converts a definition to a Pattern
func (pd *PatternDefinition) ToPattern() Pattern {
	return Pattern{
		ID:          pd.ID,
		Name:        pd.Name,
		Category:    pd.Category,
		Severity:    Severity(pd.Severity),
		Keywords:    pd.Keywords,
		Description: pd.Description,
		Suggestion:  pd.Suggestion,
	}
}
