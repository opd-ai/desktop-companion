package artifact

import (
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// RetentionPolicy defines how long artifacts should be kept
type RetentionPolicy struct {
	Name            string        `json:"name"`
	RetentionPeriod time.Duration `json:"retention_period"`
	MaxCount        int           `json:"max_count"`        // Maximum number of artifacts to keep
	CompressAfter   time.Duration `json:"compress_after"`   // Compress artifacts older than this
	CleanupInterval time.Duration `json:"cleanup_interval"` // How often to run cleanup
}

// ArtifactInfo contains metadata about a build artifact
type ArtifactInfo struct {
	Name         string            `json:"name"`
	Character    string            `json:"character"`
	Platform     string            `json:"platform"`
	Architecture string            `json:"architecture"`
	Size         int64             `json:"size"`
	Checksum     string            `json:"checksum"`
	CreatedAt    time.Time         `json:"created_at"`
	ModifiedAt   time.Time         `json:"modified_at"`
	Compressed   bool              `json:"compressed"`
	Metadata     map[string]string `json:"metadata"`
}

// Manager handles artifact storage, retention, and cleanup
type Manager struct {
	artifactsDir string
	policies     map[string]RetentionPolicy
}

// DefaultRetentionPolicies returns standard retention policies for different artifact types
func DefaultRetentionPolicies() map[string]RetentionPolicy {
	return map[string]RetentionPolicy{
		"development": {
			Name:            "development",
			RetentionPeriod: 7 * 24 * time.Hour,   // 7 days
			MaxCount:        50,                    // Keep max 50 development builds
			CompressAfter:   24 * time.Hour,       // Compress after 1 day
			CleanupInterval: 4 * time.Hour,        // Cleanup every 4 hours
		},
		"production": {
			Name:            "production",
			RetentionPeriod: 90 * 24 * time.Hour,  // 90 days
			MaxCount:        200,                   // Keep max 200 production builds
			CompressAfter:   7 * 24 * time.Hour,   // Compress after 1 week
			CleanupInterval: 24 * time.Hour,       // Cleanup daily
		},
		"release": {
			Name:            "release",
			RetentionPeriod: 365 * 24 * time.Hour, // 1 year
			MaxCount:        -1,                    // Unlimited count
			CompressAfter:   30 * 24 * time.Hour,  // Compress after 1 month
			CleanupInterval: 7 * 24 * time.Hour,   // Cleanup weekly
		},
	}
}

// NewManager creates a new artifact manager with the specified artifacts directory
func NewManager(artifactsDir string) (*Manager, error) {
	if err := os.MkdirAll(artifactsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create artifacts directory: %w", err)
	}

	return &Manager{
		artifactsDir: artifactsDir,
		policies:     DefaultRetentionPolicies(),
	}, nil
}

// SetRetentionPolicy updates or adds a retention policy
func (m *Manager) SetRetentionPolicy(name string, policy RetentionPolicy) {
	m.policies[name] = policy
}

// StoreArtifact stores a build artifact with metadata
func (m *Manager) StoreArtifact(srcPath, character, platform, arch string, metadata map[string]string) (*ArtifactInfo, error) {
	// Generate artifact info
	info, err := m.generateArtifactInfo(srcPath, character, platform, arch, metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to generate artifact info: %w", err)
	}

	// Create destination directory structure
	destDir := filepath.Join(m.artifactsDir, character, platform+"_"+arch)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Copy artifact to managed location
	destPath := filepath.Join(destDir, info.Name)
	if err := m.copyFile(srcPath, destPath); err != nil {
		return nil, fmt.Errorf("failed to copy artifact: %w", err)
	}

	// Store metadata
	if err := m.storeMetadata(destDir, info); err != nil {
		return nil, fmt.Errorf("failed to store metadata: %w", err)
	}

	return info, nil
}

// generateArtifactInfo creates metadata for an artifact
func (m *Manager) generateArtifactInfo(srcPath, character, platform, arch string, metadata map[string]string) (*ArtifactInfo, error) {
	stat, err := os.Stat(srcPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat artifact file: %w", err)
	}

	// Calculate checksum
	checksum, err := m.calculateChecksum(srcPath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate checksum: %w", err)
	}

	// Generate artifact name with timestamp
	timestamp := time.Now().Format("20060102-150405")
	ext := filepath.Ext(srcPath)
	name := fmt.Sprintf("%s_%s_%s_%s%s", character, platform, arch, timestamp, ext)

	if metadata == nil {
		metadata = make(map[string]string)
	}

	return &ArtifactInfo{
		Name:         name,
		Character:    character,
		Platform:     platform,
		Architecture: arch,
		Size:         stat.Size(),
		Checksum:     checksum,
		CreatedAt:    time.Now(),
		ModifiedAt:   stat.ModTime(),
		Compressed:   false,
		Metadata:     metadata,
	}, nil
}

// calculateChecksum computes SHA256 checksum of a file
func (m *Manager) calculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// copyFile copies a file from src to dst
func (m *Manager) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// storeMetadata writes artifact metadata to a JSON file
func (m *Manager) storeMetadata(dir string, info *ArtifactInfo) error {
	// Use Go's standard encoding/json for metadata storage
	// This follows the project's "standard library first" principle
	metadataPath := filepath.Join(dir, strings.TrimSuffix(info.Name, filepath.Ext(info.Name))+".json")
	
	file, err := os.Create(metadataPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write JSON metadata using standard library
	return writeJSON(file, info)
}

// ListArtifacts returns all artifacts for a character/platform combination
func (m *Manager) ListArtifacts(character, platform, arch string) ([]*ArtifactInfo, error) {
	searchDir := filepath.Join(m.artifactsDir, character)
	if platform != "" && arch != "" {
		searchDir = filepath.Join(searchDir, platform+"_"+arch)
	}

	var artifacts []*ArtifactInfo

	err := filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip metadata files and directories
		if info.IsDir() || strings.HasSuffix(path, ".json") {
			return nil
		}

		// Try to load metadata
		metadataPath := strings.TrimSuffix(path, filepath.Ext(path)) + ".json"
		if artifactInfo, err := m.loadMetadata(metadataPath); err == nil {
			artifacts = append(artifacts, artifactInfo)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk artifacts directory: %w", err)
	}

	// Sort by creation time (newest first)
	sort.Slice(artifacts, func(i, j int) bool {
		return artifacts[i].CreatedAt.After(artifacts[j].CreatedAt)
	})

	return artifacts, nil
}

// loadMetadata reads artifact metadata from a JSON file
func (m *Manager) loadMetadata(metadataPath string) (*ArtifactInfo, error) {
	file, err := os.Open(metadataPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var info ArtifactInfo
	if err := readJSON(file, &info); err != nil {
		return nil, err
	}

	return &info, nil
}

// CleanupArtifacts removes expired artifacts based on retention policies
func (m *Manager) CleanupArtifacts(policyName string) error {
	policy, exists := m.policies[policyName]
	if !exists {
		return fmt.Errorf("retention policy %q not found", policyName)
	}

	cutoffTime := time.Now().Add(-policy.RetentionPeriod)

	err := filepath.Walk(m.artifactsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and metadata files
		if info.IsDir() || strings.HasSuffix(path, ".json") {
			return nil
		}

		// Check if artifact is expired
		if info.ModTime().Before(cutoffTime) {
			// Remove both artifact and metadata
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("failed to remove expired artifact %s: %w", path, err)
			}

			metadataPath := strings.TrimSuffix(path, filepath.Ext(path)) + ".json"
			if err := os.Remove(metadataPath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to remove metadata %s: %w", metadataPath, err)
			}
		}

		return nil
	})

	return err
}

// CompressOldArtifacts compresses artifacts older than the policy threshold
func (m *Manager) CompressOldArtifacts(policyName string) error {
	policy, exists := m.policies[policyName]
	if !exists {
		return fmt.Errorf("retention policy %q not found", policyName)
	}

	compressAfter := time.Now().Add(-policy.CompressAfter)

	err := filepath.Walk(m.artifactsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories, metadata files, and already compressed files
		if info.IsDir() || strings.HasSuffix(path, ".json") || strings.HasSuffix(path, ".gz") {
			return nil
		}

		// Check if artifact should be compressed
		if info.ModTime().Before(compressAfter) {
			if err := m.compressFile(path); err != nil {
				return fmt.Errorf("failed to compress artifact %s: %w", path, err)
			}
		}

		return nil
	})

	return err
}

// compressFile compresses a file using gzip and removes the original
func (m *Manager) compressFile(filePath string) error {
	// Open source file
	srcFile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create compressed file
	compressedPath := filePath + ".gz"
	dstFile, err := os.Create(compressedPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Create gzip writer
	gzWriter := gzip.NewWriter(dstFile)
	defer gzWriter.Close()

	// Copy and compress
	if _, err := io.Copy(gzWriter, srcFile); err != nil {
		return err
	}

	// Close gzip writer to flush
	if err := gzWriter.Close(); err != nil {
		return err
	}

	// Remove original file
	return os.Remove(filePath)
}

// GetArtifactStats returns statistics about stored artifacts
func (m *Manager) GetArtifactStats() (map[string]interface{}, error) {
	stats := map[string]interface{}{
		"total_artifacts": 0,
		"total_size":      int64(0),
		"characters":      make(map[string]int),
		"platforms":       make(map[string]int),
		"compressed":      0,
	}

	err := filepath.Walk(m.artifactsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and metadata files
		if info.IsDir() || strings.HasSuffix(path, ".json") {
			return nil
		}

		stats["total_artifacts"] = stats["total_artifacts"].(int) + 1
		stats["total_size"] = stats["total_size"].(int64) + info.Size()

		if strings.HasSuffix(path, ".gz") {
			stats["compressed"] = stats["compressed"].(int) + 1
		}

		// Extract character and platform from path
		relPath, err := filepath.Rel(m.artifactsDir, path)
		if err == nil {
			parts := strings.Split(relPath, string(filepath.Separator))
			if len(parts) >= 2 {
				character := parts[0]
				platform := parts[1]

				charCounts := stats["characters"].(map[string]int)
				charCounts[character]++
				stats["characters"] = charCounts

				platCounts := stats["platforms"].(map[string]int)
				platCounts[platform]++
				stats["platforms"] = platCounts
			}
		}

		return nil
	})

	return stats, err
}
