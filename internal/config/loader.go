package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Loader handles configuration file loading and validation
// Uses standard library encoding/json for simplicity and reliability
type Loader struct {
	basePath string
}

// New creates a new configuration loader
func New(basePath string) *Loader {
	return &Loader{
		basePath: basePath,
	}
}

// LoadJSON loads and parses a JSON configuration file
// Generic function that can load any JSON structure
func (l *Loader) LoadJSON(filename string, target interface{}) error {
	fullPath := filepath.Join(l.basePath, filename)
	
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", fullPath, err)
	}

	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to parse JSON config %s: %w", fullPath, err)
	}

	return nil
}

// SaveJSON saves a configuration structure to a JSON file
// Useful for creating default configs or saving user preferences
func (l *Loader) SaveJSON(filename string, data interface{}) error {
	fullPath := filepath.Join(l.basePath, filename)
	
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config data: %w", err)
	}

	if err := os.WriteFile(fullPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write config file %s: %w", fullPath, err)
	}

	return nil
}

// FileExists checks if a configuration file exists
func (l *Loader) FileExists(filename string) bool {
	fullPath := filepath.Join(l.basePath, filename)
	_, err := os.Stat(fullPath)
	return err == nil
}

// GetFullPath returns the absolute path for a config file
func (l *Loader) GetFullPath(filename string) string {
	return filepath.Join(l.basePath, filename)
}

// ListFiles returns all files in the configuration directory with the given extension
func (l *Loader) ListFiles(extension string) ([]string, error) {
	pattern := filepath.Join(l.basePath, "*."+extension)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to list %s files: %w", extension, err)
	}

	// Convert to relative paths
	var files []string
	for _, match := range matches {
		rel, err := filepath.Rel(l.basePath, match)
		if err == nil {
			files = append(files, rel)
		}
	}

	return files, nil
}
