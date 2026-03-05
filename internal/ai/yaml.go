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
	registry *PatternRegistryV2
}

// NewYAMLLoader creates a new YAML pattern loader
func NewYAMLLoader(registry *PatternRegistryV2) *YAMLLoader {
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

	var pattern PatternV2
	if err := yaml.Unmarshal(data, &pattern); err != nil {
		return fmt.Errorf("cannot parse YAML: %w", err)
	}

	// Validate essential fields
	if pattern.ID == "" {
		return fmt.Errorf("pattern ID is required")
	}
	if len(pattern.Matchers) == 0 {
		return fmt.Errorf("at least one matcher is required")
	}

	if err := yl.registry.Register(pattern); err != nil {
		return fmt.Errorf("invalid pattern: %w", err)
	}

	return nil
}
