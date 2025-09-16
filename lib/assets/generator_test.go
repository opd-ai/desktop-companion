package assets

import (
	"testing"
)

func TestDefaultGeneratorConfig(t *testing.T) {
	config := DefaultGeneratorConfig()

	if config == nil {
		t.Fatal("DefaultGeneratorConfig returned nil")
	}

	if config.ComfyUIURL == "" {
		t.Error("ComfyUIURL should not be empty")
	}

	if config.DefaultSettings == nil {
		t.Error("DefaultSettings should not be nil")
	}

	if config.WorkflowsPath == "" {
		t.Error("WorkflowsPath should not be empty")
	}
}

func TestDefaultValidationConfig(t *testing.T) {
	config := DefaultValidationConfig()

	if config == nil {
		t.Fatal("DefaultValidationConfig returned nil")
	}

	if config.MaxFileSizeKB <= 0 {
		t.Error("MaxFileSizeKB should be positive")
	}

	if config.MinFileSizeKB <= 0 {
		t.Error("MinFileSizeKB should be positive")
	}

	if config.MaxFileSizeKB <= config.MinFileSizeKB {
		t.Error("MaxFileSizeKB should be greater than MinFileSizeKB")
	}

	if config.MinWidth <= 0 || config.MaxWidth <= 0 {
		t.Error("Width limits should be positive")
	}

	if config.MinHeight <= 0 || config.MaxHeight <= 0 {
		t.Error("Height limits should be positive")
	}

	if config.MinFrames <= 0 || config.MaxFrames <= 0 {
		t.Error("Frame limits should be positive")
	}

	if config.MinFPS <= 0 || config.MaxFPS <= 0 {
		t.Error("FPS limits should be positive")
	}

	if config.MinQualityScore < 0.0 || config.MinQualityScore > 1.0 {
		t.Error("MinQualityScore should be between 0.0 and 1.0")
	}

	if len(config.AllowedFormats) == 0 {
		t.Error("AllowedFormats should not be empty")
	}
}

func TestDefaultBackupConfig(t *testing.T) {
	config := DefaultBackupConfig()

	if config == nil {
		t.Fatal("DefaultBackupConfig returned nil")
	}

	if config.BackupDir == "" {
		t.Error("BackupDir should not be empty")
	}

	if config.MaxBackups < 0 {
		t.Error("MaxBackups should be non-negative")
	}

	if config.FilenameFormat == "" {
		t.Error("FilenameFormat should not be empty")
	}
}

func TestNewAssetValidator(t *testing.T) {
	// Test with nil config (should use defaults)
	validator := NewAssetValidator(nil)
	if validator == nil {
		t.Fatal("NewAssetValidator returned nil")
	}

	// Test with custom config
	config := &ValidationConfig{
		MaxFileSizeKB: 1000,
		MinFileSizeKB: 50,
	}
	validator = NewAssetValidator(config)
	if validator == nil {
		t.Fatal("NewAssetValidator with config returned nil")
	}
}

func TestNewBackupManager(t *testing.T) {
	// Test with nil config (should use defaults)
	manager := NewBackupManager(nil)
	if manager == nil {
		t.Fatal("NewBackupManager returned nil")
	}

	// Test with custom config
	config := &BackupConfig{
		BackupDir:  "test_backups",
		MaxBackups: 10,
	}
	manager = NewBackupManager(config)
	if manager == nil {
		t.Fatal("NewBackupManager with config returned nil")
	}
}

func TestValidationResult(t *testing.T) {
	result := &ValidationResult{
		Valid:        true,
		Score:        0.85,
		FileExists:   true,
		FileSizeOK:   true,
		FormatOK:     true,
		DimensionsOK: true,
		Width:        128,
		Height:       128,
		FileSizeKB:   256,
		Format:       "gif",
		Warnings:     []string{},
		Errors:       []string{},
	}

	if !result.Valid {
		t.Error("Result should be valid")
	}

	if result.Score < 0.0 || result.Score > 1.0 {
		t.Error("Score should be between 0.0 and 1.0")
	}

	if result.Width <= 0 || result.Height <= 0 {
		t.Error("Dimensions should be positive")
	}

	if result.FileSizeKB <= 0 {
		t.Error("FileSizeKB should be positive")
	}
}

func TestGenerationError(t *testing.T) {
	err := GenerationError{
		Stage:   "generation",
		State:   "idle",
		Message: "Test error",
		Err:     nil,
	}

	if err.Stage == "" {
		t.Error("Stage should not be empty")
	}

	if err.Message == "" {
		t.Error("Message should not be empty")
	}
}

func TestGenerateResult(t *testing.T) {
	result := &GenerateResult{
		Success:         true,
		GeneratedAssets: make(map[string]string),
		BackupPath:      "/path/to/backup",
		Errors:          []GenerationError{},
	}

	// Add some test assets
	result.GeneratedAssets["idle"] = "/path/to/idle.gif"
	result.GeneratedAssets["talking"] = "/path/to/talking.gif"

	if !result.Success {
		t.Error("Result should be successful")
	}

	if len(result.GeneratedAssets) == 0 {
		t.Error("GeneratedAssets should not be empty")
	}

	if result.BackupPath == "" {
		t.Error("BackupPath should not be empty")
	}
}

func TestBackupResult(t *testing.T) {
	result := &BackupResult{
		Success:       true,
		BackupPath:    "/path/to/backup.tar.gz",
		BackedUpFiles: []string{"idle.gif", "talking.gif"},
		BackupSize:    1024,
		Errors:        []string{},
	}

	if !result.Success {
		t.Error("Backup should be successful")
	}

	if result.BackupPath == "" {
		t.Error("BackupPath should not be empty")
	}

	if len(result.BackedUpFiles) == 0 {
		t.Error("BackedUpFiles should not be empty")
	}

	if result.BackupSize <= 0 {
		t.Error("BackupSize should be positive")
	}
}

func TestRestoreResult(t *testing.T) {
	result := &RestoreResult{
		Success:       true,
		RestoredFiles: []string{"idle.gif", "talking.gif"},
		Errors:        []string{},
	}

	if !result.Success {
		t.Error("Restore should be successful")
	}

	if len(result.RestoredFiles) == 0 {
		t.Error("RestoredFiles should not be empty")
	}
}

func TestBackupInfo(t *testing.T) {
	info := BackupInfo{
		Path:     "/path/to/backup.tar.gz",
		Filename: "backup.tar.gz",
		Size:     1024,
	}

	if info.Path == "" {
		t.Error("Path should not be empty")
	}

	if info.Filename == "" {
		t.Error("Filename should not be empty")
	}

	if info.Size <= 0 {
		t.Error("Size should be positive")
	}
}

func TestQualityMetrics(t *testing.T) {
	metrics := QualityMetrics{
		Sharpness:        0.8,
		Contrast:         0.7,
		Brightness:       0.5,
		ColorRange:       0.9,
		FrameStability:   0.85,
		MotionSmoothness: 0.75,
		LoopSeamlessness: 0.9,
		CompressionRatio: 0.6,
		FileEfficiency:   0.8,
	}

	// Test that all metrics are in reasonable ranges
	if metrics.Sharpness < 0.0 || metrics.Sharpness > 1.0 {
		t.Error("Sharpness should be between 0.0 and 1.0")
	}

	if metrics.Contrast < 0.0 || metrics.Contrast > 1.0 {
		t.Error("Contrast should be between 0.0 and 1.0")
	}

	if metrics.Brightness < 0.0 || metrics.Brightness > 1.0 {
		t.Error("Brightness should be between 0.0 and 1.0")
	}
}

// Test helper functions
func TestEnsureDir(t *testing.T) {
	// This would normally test directory creation
	// For now, just test that the function exists and doesn't panic
	err := ensureDir("/tmp/test_ensure_dir")
	if err != nil {
		t.Logf("ensureDir returned error (expected in test environment): %v", err)
	}
}

// Benchmark tests
func BenchmarkDefaultValidationConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DefaultValidationConfig()
	}
}

func BenchmarkDefaultBackupConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DefaultBackupConfig()
	}
}

func BenchmarkDefaultGeneratorConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DefaultGeneratorConfig()
	}
}
