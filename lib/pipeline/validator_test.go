package pipeline

import (
	"context"
	"image"
	"image/color"
	"image/gif"
	"os"
	"path/filepath"
	"testing"
)

func TestNewValidator(t *testing.T) {
	validator := NewValidator()
	if validator == nil {
		t.Fatal("NewValidator returned nil")
	}
}

func TestValidateAsset(t *testing.T) {
	validator := NewValidator()

	// Create temporary directory and test GIF
	tmpDir := t.TempDir()
	testGIF := filepath.Join(tmpDir, "test.gif")

	// Create a valid test GIF
	createTestGIF(t, testGIF, 6, 128, 128, true)

	config := &ValidationConfig{
		MaxFileSize:          500000,
		MinFrameRate:         10,
		TransparencyRequired: true,
	}

	result, err := validator.ValidateAsset(context.Background(), testGIF, config)
	if err != nil {
		t.Fatalf("ValidateAsset failed: %v", err)
	}

	if !result.Valid {
		t.Errorf("Expected valid result, got invalid. Errors: %v", result.Errors)
	}

	if result.Metrics == nil {
		t.Fatal("Expected metrics to be populated")
	}

	// Check metrics
	if result.Metrics.FrameCount != 6 {
		t.Errorf("Expected 6 frames, got %d", result.Metrics.FrameCount)
	}
	if result.Metrics.Dimensions[0] != 128 || result.Metrics.Dimensions[1] != 128 {
		t.Errorf("Expected 128x128 dimensions, got %dx%d",
			result.Metrics.Dimensions[0], result.Metrics.Dimensions[1])
	}
	if !result.Metrics.HasTransparency {
		t.Error("Expected transparency to be detected")
	}
}

func TestValidateAssetInvalidFile(t *testing.T) {
	validator := NewValidator()
	config := &ValidationConfig{MaxFileSize: 500000}

	result, err := validator.ValidateAsset(context.Background(), "nonexistent.gif", config)
	if err != nil {
		t.Fatalf("Expected no error for missing file, got: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid result for missing file")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected errors for missing file")
	}

	// Check for FILE_NOT_FOUND error
	found := false
	for _, err := range result.Errors {
		if err.Code == "FILE_NOT_FOUND" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected FILE_NOT_FOUND error")
	}
}

func TestValidateAssetFileSize(t *testing.T) {
	validator := NewValidator()
	tmpDir := t.TempDir()
	testGIF := filepath.Join(tmpDir, "large.gif")

	// Create a valid GIF
	createTestGIF(t, testGIF, 6, 128, 128, true)

	// Set a very small max file size
	config := &ValidationConfig{
		MaxFileSize:          100, // Very small
		TransparencyRequired: false,
	}

	result, err := validator.ValidateAsset(context.Background(), testGIF, config)
	if err != nil {
		t.Fatalf("ValidateAsset failed: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid result for large file")
	}

	// Check for FILE_SIZE_EXCEEDED error
	found := false
	for _, err := range result.Errors {
		if err.Code == "FILE_SIZE_EXCEEDED" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected FILE_SIZE_EXCEEDED error")
	}
}

func TestValidateAssetFrameCount(t *testing.T) {
	validator := NewValidator()
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		frameCount  int
		expectValid bool
	}{
		{"too few frames", 2, false},
		{"minimum frames", 4, true},
		{"normal frames", 6, true},
		{"maximum frames", 8, true},
		{"too many frames", 10, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testGIF := filepath.Join(tmpDir, tt.name+".gif")
			createTestGIF(t, testGIF, tt.frameCount, 128, 128, false)

			config := &ValidationConfig{
				MaxFileSize:          500000,
				TransparencyRequired: false,
			}

			result, err := validator.ValidateAsset(context.Background(), testGIF, config)
			if err != nil {
				t.Fatalf("ValidateAsset failed: %v", err)
			}

			if result.Valid != tt.expectValid {
				t.Errorf("Expected valid=%v, got valid=%v", tt.expectValid, result.Valid)
			}

			if !tt.expectValid {
				// Check for INVALID_FRAME_COUNT error
				found := false
				for _, err := range result.Errors {
					if err.Code == "INVALID_FRAME_COUNT" {
						found = true
						break
					}
				}
				if !found {
					t.Error("Expected INVALID_FRAME_COUNT error")
				}
			}
		})
	}
}

func TestValidateCharacterSet(t *testing.T) {
	validator := NewValidator()
	tmpDir := t.TempDir()

	// Create character directory structure
	characterDir := filepath.Join(tmpDir, "test_character")
	animationsDir := filepath.Join(characterDir, "animations")
	if err := os.MkdirAll(animationsDir, 0755); err != nil {
		t.Fatalf("Failed to create animations directory: %v", err)
	}

	// Create test assets for required states
	states := []string{"idle", "talking", "happy", "sad"}
	for _, state := range states {
		gifPath := filepath.Join(animationsDir, state+".gif")
		createTestGIF(t, gifPath, 6, 128, 128, true)
	}

	config := DefaultCharacterConfig("test")
	config.States = states
	config.Deployment.OutputDir = characterDir

	result, err := validator.ValidateCharacterSet(context.Background(), characterDir, config)
	if err != nil {
		t.Fatalf("ValidateCharacterSet failed: %v", err)
	}

	if !result.Valid {
		t.Errorf("Expected valid character set, got invalid. Missing states: %v", result.MissingStates)
	}

	if len(result.MissingStates) > 0 {
		t.Errorf("Expected no missing states, got: %v", result.MissingStates)
	}

	if len(result.AssetResults) != len(states) {
		t.Errorf("Expected %d asset results, got %d", len(states), len(result.AssetResults))
	}

	// Check that all required states have results
	for _, state := range states {
		if _, exists := result.AssetResults[state]; !exists {
			t.Errorf("Missing asset result for state: %s", state)
		}
	}
}

func TestValidateCharacterSetMissingStates(t *testing.T) {
	validator := NewValidator()
	tmpDir := t.TempDir()

	// Create character directory with only some assets
	characterDir := filepath.Join(tmpDir, "incomplete_character")
	animationsDir := filepath.Join(characterDir, "animations")
	if err := os.MkdirAll(animationsDir, 0755); err != nil {
		t.Fatalf("Failed to create animations directory: %v", err)
	}

	// Create only idle and talking states
	createTestGIF(t, filepath.Join(animationsDir, "idle.gif"), 6, 128, 128, true)
	createTestGIF(t, filepath.Join(animationsDir, "talking.gif"), 6, 128, 128, true)

	config := DefaultCharacterConfig("test")
	config.States = []string{"idle", "talking", "happy", "sad"}
	config.Deployment.OutputDir = characterDir

	result, err := validator.ValidateCharacterSet(context.Background(), characterDir, config)
	if err != nil {
		t.Fatalf("ValidateCharacterSet failed: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid character set due to missing states")
	}

	expectedMissing := []string{"happy", "sad"}
	if len(result.MissingStates) != len(expectedMissing) {
		t.Errorf("Expected %d missing states, got %d", len(expectedMissing), len(result.MissingStates))
	}

	for _, expected := range expectedMissing {
		found := false
		for _, missing := range result.MissingStates {
			if missing == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected missing state %s not found in %v", expected, result.MissingStates)
		}
	}
}

func TestValidateBatch(t *testing.T) {
	validator := NewValidator()
	tmpDir := t.TempDir()

	// Create two character sets
	characters := []string{"character1", "character2"}
	configs := []*CharacterConfig{}

	for _, char := range characters {
		characterDir := filepath.Join(tmpDir, char)
		animationsDir := filepath.Join(characterDir, "animations")
		if err := os.MkdirAll(animationsDir, 0755); err != nil {
			t.Fatalf("Failed to create animations directory for %s: %v", char, err)
		}

		// Create test assets
		states := []string{"idle", "talking"}
		for _, state := range states {
			gifPath := filepath.Join(animationsDir, state+".gif")
			createTestGIF(t, gifPath, 6, 128, 128, true)
		}

		config := DefaultCharacterConfig(char)
		config.States = states
		config.Deployment.OutputDir = characterDir
		configs = append(configs, config)
	}

	result, err := validator.ValidateBatch(context.Background(), configs)
	if err != nil {
		t.Fatalf("ValidateBatch failed: %v", err)
	}

	if !result.OverallValid {
		t.Error("Expected valid batch result")
	}

	if len(result.Characters) != len(characters) {
		t.Errorf("Expected %d character results, got %d", len(characters), len(result.Characters))
	}

	if result.Summary == nil {
		t.Fatal("Expected summary to be populated")
	}

	if result.Summary.TotalAssets != 4 { // 2 characters * 2 states each
		t.Errorf("Expected 4 total assets, got %d", result.Summary.TotalAssets)
	}

	if result.ProcessingTime <= 0 {
		t.Error("Expected positive processing time")
	}
}

func TestStyleConsistency(t *testing.T) {
	validator := &assetValidator{}

	// Create test asset results with consistent dimensions
	assetResults := map[string]*ValidationResult{
		"idle": {
			Valid:   true,
			Metrics: &AssetMetrics{Dimensions: [2]int{128, 128}, Colors: 100},
		},
		"talking": {
			Valid:   true,
			Metrics: &AssetMetrics{Dimensions: [2]int{128, 128}, Colors: 110},
		},
	}

	result := validator.checkStyleConsistency(assetResults)
	if !result.Consistent {
		t.Error("Expected consistent style for similar assets")
	}
	if result.Score <= 0.7 {
		t.Errorf("Expected high consistency score, got %.2f", result.Score)
	}

	// Test inconsistent dimensions
	assetResults["happy"] = &ValidationResult{
		Valid:   true,
		Metrics: &AssetMetrics{Dimensions: [2]int{64, 64}, Colors: 100},
	}

	result = validator.checkStyleConsistency(assetResults)
	if result.Consistent {
		t.Error("Expected inconsistent style for different dimensions")
	}
	if len(result.Inconsistencies) == 0 {
		t.Error("Expected style inconsistencies to be reported")
	}
}

func TestValidationConfigDefaults(t *testing.T) {
	config := DefaultPipelineConfig()
	validation := config.Validation

	if validation.MaxFileSize != 500000 {
		t.Errorf("Expected max file size 500000, got %d", validation.MaxFileSize)
	}
	if validation.MinFrameRate != 10 {
		t.Errorf("Expected min frame rate 10, got %d", validation.MinFrameRate)
	}
	if !validation.TransparencyRequired {
		t.Error("Expected transparency to be required")
	}
	if len(validation.RequiredStates) != 4 {
		t.Errorf("Expected 4 required states, got %d", len(validation.RequiredStates))
	}
}

func TestExtractMetrics(t *testing.T) {
	validator := &assetValidator{}
	tmpDir := t.TempDir()
	testGIF := filepath.Join(tmpDir, "metrics_test.gif")

	// Create test GIF with known properties
	frameCount := 6
	width, height := 128, 128
	transparency := true
	createTestGIF(t, testGIF, frameCount, width, height, transparency)

	metrics, err := validator.extractMetrics(testGIF)
	if err != nil {
		t.Fatalf("extractMetrics failed: %v", err)
	}

	if metrics.FrameCount != frameCount {
		t.Errorf("Expected %d frames, got %d", frameCount, metrics.FrameCount)
	}
	if metrics.Dimensions[0] != width || metrics.Dimensions[1] != height {
		t.Errorf("Expected %dx%d dimensions, got %dx%d",
			width, height, metrics.Dimensions[0], metrics.Dimensions[1])
	}
	if metrics.HasTransparency != transparency {
		t.Errorf("Expected transparency=%v, got %v", transparency, metrics.HasTransparency)
	}
	if metrics.FileSize <= 0 {
		t.Error("Expected positive file size")
	}
	if metrics.Colors <= 0 {
		t.Error("Expected positive color count")
	}
}

// createTestGIF creates a test GIF file with specified properties.
func createTestGIF(t *testing.T, path string, frameCount, width, height int, transparency bool) {
	t.Helper()

	gifData := &gif.GIF{}

	for i := 0; i < frameCount; i++ {
		// Create a simple paletted image
		palette := color.Palette{
			color.RGBA{255, 255, 255, 255}, // White
			color.RGBA{0, 0, 0, 255},       // Black
			color.RGBA{255, 0, 0, 255},     // Red
		}

		if transparency {
			palette = append(palette, color.RGBA{0, 0, 0, 0}) // Transparent
		}

		bounds := image.Rect(0, 0, width, height)
		img := image.NewPaletted(bounds, palette)

		// Fill with a simple pattern that varies by frame
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				colorIndex := uint8((x + y + i) % len(palette))
				img.SetColorIndex(x, y, colorIndex)
			}
		}

		gifData.Image = append(gifData.Image, img)
		gifData.Delay = append(gifData.Delay, 10) // 100ms delay
	}

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create test GIF: %v", err)
	}
	defer file.Close()

	if err := gif.EncodeAll(file, gifData); err != nil {
		t.Fatalf("Failed to encode test GIF: %v", err)
	}
}
