package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus"
)

// getCaller returns the calling function name for structured logging
func getCaller() string {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return "unknown"
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}
	return fn.Name()
}

// Loader handles configuration file loading and validation
// Uses standard library encoding/json for simplicity and reliability
type Loader struct {
	basePath string
}

// New creates a new configuration loader
func New(basePath string) *Loader {
	caller := getCaller()
	logrus.WithFields(logrus.Fields{
		"caller":   caller,
		"basePath": basePath,
	}).Info("Creating new configuration loader")

	loader := &Loader{
		basePath: basePath,
	}

	logrus.WithFields(logrus.Fields{
		"caller":   caller,
		"basePath": basePath,
	}).Info("Configuration loader created successfully")

	return loader
}

// LoadJSON loads and parses a JSON configuration file
// Generic function that can load any JSON structure
func (l *Loader) LoadJSON(filename string, target interface{}) error {
	caller := getCaller()
	logrus.WithFields(logrus.Fields{
		"caller":   caller,
		"filename": filename,
		"basePath": l.basePath,
	}).Info("Loading JSON configuration file")

	fullPath := filepath.Join(l.basePath, filename)

	logrus.WithFields(logrus.Fields{
		"caller":   caller,
		"fullPath": fullPath,
	}).Debug("Reading configuration file from disk")

	data, err := os.ReadFile(fullPath)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"caller":   caller,
			"fullPath": fullPath,
			"error":    err.Error(),
		}).Error("Failed to read configuration file")
		return fmt.Errorf("failed to read config file %s: %w", fullPath, err)
	}

	logrus.WithFields(logrus.Fields{
		"caller":   caller,
		"fileSize": len(data),
	}).Debug("Configuration file read successfully")

	logrus.WithFields(logrus.Fields{
		"caller": caller,
	}).Debug("Unmarshaling JSON data")

	if err := json.Unmarshal(data, target); err != nil {
		logrus.WithFields(logrus.Fields{
			"caller":   caller,
			"fullPath": fullPath,
			"error":    err.Error(),
		}).Error("Failed to parse JSON configuration")
		return fmt.Errorf("failed to parse JSON config %s: %w", fullPath, err)
	}

	logrus.WithFields(logrus.Fields{
		"caller":   caller,
		"filename": filename,
	}).Info("JSON configuration loaded successfully")

	return nil
}

// SaveJSON saves a configuration structure to a JSON file
// Useful for creating default configs or saving user preferences
func (l *Loader) SaveJSON(filename string, data interface{}) error {
	caller := getCaller()
	logrus.WithFields(logrus.Fields{
		"caller":   caller,
		"filename": filename,
		"basePath": l.basePath,
	}).Info("Saving configuration to JSON file")

	fullPath := filepath.Join(l.basePath, filename)

	logrus.WithFields(logrus.Fields{
		"caller":   caller,
		"fullPath": fullPath,
	}).Debug("Creating directory structure if needed")

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		logrus.WithFields(logrus.Fields{
			"caller":    caller,
			"directory": filepath.Dir(fullPath),
			"error":     err.Error(),
		}).Error("Failed to create configuration directory")
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"caller": caller,
	}).Debug("Marshaling data to JSON")

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"caller": caller,
			"error":  err.Error(),
		}).Error("Failed to marshal configuration data to JSON")
		return fmt.Errorf("failed to marshal config data: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"caller":   caller,
		"dataSize": len(jsonData),
	}).Debug("JSON data marshaled successfully")

	logrus.WithFields(logrus.Fields{
		"caller":   caller,
		"fullPath": fullPath,
	}).Debug("Writing JSON data to file")

	if err := os.WriteFile(fullPath, jsonData, 0o644); err != nil {
		logrus.WithFields(logrus.Fields{
			"caller":   caller,
			"fullPath": fullPath,
			"error":    err.Error(),
		}).Error("Failed to write configuration file")
		return fmt.Errorf("failed to write config file %s: %w", fullPath, err)
	}

	logrus.WithFields(logrus.Fields{
		"caller":   caller,
		"filename": filename,
		"fullPath": fullPath,
	}).Info("Configuration saved successfully")

	return nil
}

// FileExists checks if a configuration file exists
func (l *Loader) FileExists(filename string) bool {
	caller := getCaller()
	logrus.WithFields(logrus.Fields{
		"caller":   caller,
		"filename": filename,
	}).Debug("Checking if configuration file exists")

	fullPath := filepath.Join(l.basePath, filename)
	_, err := os.Stat(fullPath)
	exists := err == nil

	logrus.WithFields(logrus.Fields{
		"caller":   caller,
		"filename": filename,
		"exists":   exists,
	}).Debug("File existence check completed")

	return exists
}

// GetFullPath returns the absolute path for a config file
func (l *Loader) GetFullPath(filename string) string {
	caller := getCaller()
	logrus.WithFields(logrus.Fields{
		"caller":   caller,
		"filename": filename,
	}).Debug("Getting full path for configuration file")

	fullPath := filepath.Join(l.basePath, filename)

	logrus.WithFields(logrus.Fields{
		"caller":   caller,
		"filename": filename,
		"fullPath": fullPath,
	}).Debug("Full path generated")

	return fullPath
}

// ListFiles returns all files in the configuration directory with the given extension
func (l *Loader) ListFiles(extension string) ([]string, error) {
	caller := getCaller()
	logrus.WithFields(logrus.Fields{
		"caller":    caller,
		"extension": extension,
		"basePath":  l.basePath,
	}).Info("Listing configuration files by extension")

	pattern := filepath.Join(l.basePath, "*."+extension)

	logrus.WithFields(logrus.Fields{
		"caller":  caller,
		"pattern": pattern,
	}).Debug("Searching for files matching pattern")

	matches, err := filepath.Glob(pattern)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"caller":    caller,
			"extension": extension,
			"pattern":   pattern,
			"error":     err.Error(),
		}).Error("Failed to list files by extension")
		return nil, fmt.Errorf("failed to list %s files: %w", extension, err)
	}

	logrus.WithFields(logrus.Fields{
		"caller":     caller,
		"matchCount": len(matches),
	}).Debug("Files found, converting to relative paths")

	// Convert to relative paths
	var files []string
	for i, match := range matches {
		logrus.WithFields(logrus.Fields{
			"caller":    caller,
			"index":     i,
			"matchPath": match,
		}).Debug("Processing matched file")

		rel, err := filepath.Rel(l.basePath, match)
		if err == nil {
			files = append(files, rel)
			logrus.WithFields(logrus.Fields{
				"caller":       caller,
				"relativePath": rel,
			}).Debug("File added to result list")
		} else {
			logrus.WithFields(logrus.Fields{
				"caller":    caller,
				"matchPath": match,
				"error":     err.Error(),
			}).Warn("Failed to convert file path to relative path")
		}
	}

	logrus.WithFields(logrus.Fields{
		"caller":    caller,
		"extension": extension,
		"fileCount": len(files),
	}).Info("Configuration file listing completed")

	return files, nil
}
