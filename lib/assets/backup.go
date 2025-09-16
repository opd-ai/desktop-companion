package assets

// backup.go implements a comprehensive backup system for character assets
// with versioning, compression, and restoration capabilities.

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// BackupManager handles asset backup and restoration operations.
type BackupManager struct {
	config *BackupConfig
}

// BackupConfig defines backup system configuration.
type BackupConfig struct {
	// Backup directory (relative to character directory)
	BackupDir string

	// Maximum number of backup sets to retain (0 = unlimited)
	MaxBackups int

	// Whether to compress backup archives
	EnableCompression bool

	// Backup filename format (supports time formatting)
	FilenameFormat string

	// Include metadata in backups
	IncludeMetadata bool

	// Backup verification
	VerifyBackups bool
}

// BackupResult contains information about a backup operation.
type BackupResult struct {
	// Success indicates if backup completed successfully
	Success bool

	// Path to the created backup
	BackupPath string

	// Files included in backup
	BackedUpFiles []string

	// Backup size in bytes
	BackupSize int64

	// Time taken to create backup
	Duration time.Duration

	// Any errors that occurred
	Errors []string
}

// RestoreResult contains information about a restore operation.
type RestoreResult struct {
	// Success indicates if restore completed successfully
	Success bool

	// Files that were restored
	RestoredFiles []string

	// Time taken to restore
	Duration time.Duration

	// Any errors that occurred
	Errors []string
}

// NewBackupManager creates a new backup manager with the given configuration.
func NewBackupManager(config *BackupConfig) *BackupManager {
	if config == nil {
		config = DefaultBackupConfig()
	}
	return &BackupManager{config: config}
}

// DefaultBackupConfig returns sensible default backup settings.
func DefaultBackupConfig() *BackupConfig {
	return &BackupConfig{
		BackupDir:         "backups",
		MaxBackups:        5,
		EnableCompression: true,
		FilenameFormat:    "backup_20060102_150405", // Go time format
		IncludeMetadata:   true,
		VerifyBackups:     true,
	}
}

// CreateBackup creates a backup of character assets.
func (bm *BackupManager) CreateBackup(basePath string, animations map[string]string) (*BackupResult, error) {
	startTime := time.Now()

	result := &BackupResult{
		BackedUpFiles: []string{},
		Errors:        []string{},
	}

	// Create backup directory
	backupDir := filepath.Join(basePath, bm.config.BackupDir)
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Create backup directory: %v", err))
		return result, nil
	}

	// Generate backup filename
	timestamp := time.Now()
	backupName := timestamp.Format(bm.config.FilenameFormat)
	if bm.config.EnableCompression {
		backupName += ".tar.gz"
	} else {
		backupName += ".tar"
	}

	result.BackupPath = filepath.Join(backupDir, backupName)

	// Create backup archive
	if err := bm.createArchive(result.BackupPath, basePath, animations, result); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Create archive: %v", err))
		return result, nil
	}

	// Get backup file size
	if info, err := os.Stat(result.BackupPath); err == nil {
		result.BackupSize = info.Size()
	}

	// Verify backup if enabled
	if bm.config.VerifyBackups {
		if err := bm.verifyBackup(result.BackupPath, result); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Backup verification failed: %v", err))
		}
	}

	// Clean up old backups
	if bm.config.MaxBackups > 0 {
		if err := bm.cleanupOldBackups(backupDir); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Cleanup old backups: %v", err))
		}
	}

	result.Duration = time.Since(startTime)
	result.Success = len(result.Errors) == 0

	return result, nil
}

// createArchive creates a tar archive (optionally compressed) containing the assets.
func (bm *BackupManager) createArchive(archivePath, basePath string, animations map[string]string, result *BackupResult) error {
	// Create archive file
	archiveFile, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("create archive file: %w", err)
	}
	defer archiveFile.Close()

	var writer io.Writer = archiveFile

	// Add compression if enabled
	var gzWriter *gzip.Writer
	if bm.config.EnableCompression {
		gzWriter = gzip.NewWriter(archiveFile)
		defer gzWriter.Close()
		writer = gzWriter
	}

	// Create tar writer
	tarWriter := tar.NewWriter(writer)
	defer tarWriter.Close()

	// Add animation files to archive
	for state, relativePath := range animations {
		fullPath := filepath.Join(basePath, relativePath)
		if err := bm.addFileToArchive(tarWriter, fullPath, relativePath, result); err != nil {
			return fmt.Errorf("add %s to archive: %w", state, err)
		}
	}

	// Add character.json if metadata is enabled
	if bm.config.IncludeMetadata {
		characterPath := filepath.Join(basePath, "character.json")
		if _, err := os.Stat(characterPath); err == nil {
			if err := bm.addFileToArchive(tarWriter, characterPath, "character.json", result); err != nil {
				return fmt.Errorf("add character.json to archive: %w", err)
			}
		}
	}

	return nil
}

// addFileToArchive adds a single file to the tar archive.
func (bm *BackupManager) addFileToArchive(tarWriter *tar.Writer, filePath, archivePath string, result *BackupResult) error {
	// Check if file exists
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, skip it (might be missing animation)
			return nil
		}
		return fmt.Errorf("stat file: %w", err)
	}

	// Skip directories
	if info.IsDir() {
		return nil
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	// Create tar header
	header := &tar.Header{
		Name:     archivePath,
		Size:     info.Size(),
		Mode:     int64(info.Mode()),
		ModTime:  info.ModTime(),
		Typeflag: tar.TypeReg,
	}

	// Write header
	if err := tarWriter.WriteHeader(header); err != nil {
		return fmt.Errorf("write tar header: %w", err)
	}

	// Copy file content
	if _, err := io.Copy(tarWriter, file); err != nil {
		return fmt.Errorf("copy file content: %w", err)
	}

	result.BackedUpFiles = append(result.BackedUpFiles, archivePath)
	return nil
}

// verifyBackup performs basic verification of a backup archive.
func (bm *BackupManager) verifyBackup(backupPath string, result *BackupResult) error {
	// Open archive
	archiveFile, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("open backup archive: %w", err)
	}
	defer archiveFile.Close()

	var reader io.Reader = archiveFile

	// Handle compression
	if bm.config.EnableCompression {
		gzReader, err := gzip.NewReader(archiveFile)
		if err != nil {
			return fmt.Errorf("create gzip reader: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	// Create tar reader
	tarReader := tar.NewReader(reader)

	// Count files in archive
	fileCount := 0
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar header: %w", err)
		}

		// Verify header is valid
		if header.Name == "" {
			return fmt.Errorf("empty filename in archive")
		}

		if header.Size < 0 {
			return fmt.Errorf("invalid file size in archive: %d", header.Size)
		}

		fileCount++
	}

	// Verify we have the expected number of files
	if fileCount == 0 {
		return fmt.Errorf("backup archive is empty")
	}

	return nil
}

// cleanupOldBackups removes old backup files to maintain the configured limit.
func (bm *BackupManager) cleanupOldBackups(backupDir string) error {
	// Find all backup files
	pattern := filepath.Join(backupDir, "backup_*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("find backup files: %w", err)
	}

	// If we don't exceed the limit, nothing to do
	if len(matches) <= bm.config.MaxBackups {
		return nil
	}

	// Sort by modification time (oldest first)
	type fileInfo struct {
		path    string
		modTime time.Time
	}

	var files []fileInfo
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			continue // Skip files we can't stat
		}
		files = append(files, fileInfo{
			path:    match,
			modTime: info.ModTime(),
		})
	}

	// Sort by modification time
	for i := 0; i < len(files)-1; i++ {
		for j := i + 1; j < len(files); j++ {
			if files[i].modTime.After(files[j].modTime) {
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	// Remove oldest files
	toRemove := len(files) - bm.config.MaxBackups
	for i := 0; i < toRemove; i++ {
		if err := os.Remove(files[i].path); err != nil {
			return fmt.Errorf("remove old backup %s: %w", files[i].path, err)
		}
	}

	return nil
}

// ListBackups returns a list of available backup files for a character.
func (bm *BackupManager) ListBackups(basePath string) ([]BackupInfo, error) {
	backupDir := filepath.Join(basePath, bm.config.BackupDir)

	// Check if backup directory exists
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return []BackupInfo{}, nil
	}

	// Find backup files
	pattern := filepath.Join(backupDir, "backup_*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("find backup files: %w", err)
	}

	var backups []BackupInfo
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			continue
		}

		backup := BackupInfo{
			Path:     match,
			Filename: filepath.Base(match),
			Size:     info.Size(),
			Created:  info.ModTime(),
		}

		backups = append(backups, backup)
	}

	return backups, nil
}

// BackupInfo contains information about a backup file.
type BackupInfo struct {
	Path     string
	Filename string
	Size     int64
	Created  time.Time
}

// RestoreFromBackup restores assets from a backup archive.
func (bm *BackupManager) RestoreFromBackup(backupPath, targetPath string) (*RestoreResult, error) {
	startTime := time.Now()

	result := &RestoreResult{
		RestoredFiles: []string{},
		Errors:        []string{},
	}

	// Open backup archive
	archiveFile, err := os.Open(backupPath)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Open backup: %v", err))
		return result, nil
	}
	defer archiveFile.Close()

	var reader io.Reader = archiveFile

	// Handle compression
	if strings.HasSuffix(backupPath, ".gz") {
		gzReader, err := gzip.NewReader(archiveFile)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Create gzip reader: %v", err))
			return result, nil
		}
		defer gzReader.Close()
		reader = gzReader
	}

	// Create tar reader
	tarReader := tar.NewReader(reader)

	// Extract files
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Read tar header: %v", err))
			continue
		}

		// Skip directories
		if header.Typeflag == tar.TypeDir {
			continue
		}

		// Construct target path
		targetFilePath := filepath.Join(targetPath, header.Name)

		// Create target directory if needed
		if err := os.MkdirAll(filepath.Dir(targetFilePath), 0o755); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Create directory for %s: %v", header.Name, err))
			continue
		}

		// Create target file
		targetFile, err := os.Create(targetFilePath)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Create file %s: %v", header.Name, err))
			continue
		}

		// Copy file content
		_, err = io.Copy(targetFile, tarReader)
		targetFile.Close()

		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Extract file %s: %v", header.Name, err))
			continue
		}

		result.RestoredFiles = append(result.RestoredFiles, header.Name)
	}

	result.Duration = time.Since(startTime)
	result.Success = len(result.Errors) == 0

	return result, nil
}

// GenerateBackupReport creates a human-readable backup report.
func (bm *BackupManager) GenerateBackupReport(backups []BackupInfo) string {
	var report strings.Builder

	report.WriteString("Asset Backup Report\n")
	report.WriteString("===================\n\n")

	if len(backups) == 0 {
		report.WriteString("No backups found.\n")
		return report.String()
	}

	totalSize := int64(0)
	for _, backup := range backups {
		report.WriteString(fmt.Sprintf("Backup: %s\n", backup.Filename))
		report.WriteString(fmt.Sprintf("  Created: %s\n", backup.Created.Format("2006-01-02 15:04:05")))
		report.WriteString(fmt.Sprintf("  Size: %.2f KB\n", float64(backup.Size)/1024))
		report.WriteString("\n")
		totalSize += backup.Size
	}

	report.WriteString(fmt.Sprintf("Total: %d backups using %.2f KB\n", len(backups), float64(totalSize)/1024))

	return report.String()
}
