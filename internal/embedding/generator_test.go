package embedding

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAnimations(t *testing.T) {
	// Create temporary test directory
	tempDir := t.TempDir()

	// Create test character card
	testCard := map[string]interface{}{
		"name": "Test Character",
		"animations": map[string]interface{}{
			"test": "test.gif",
		},
	}

	// Copy a real GIF file from the assets for testing
	srcGifPath := "../../assets/characters/default/animations/idle.gif"
	testGifData, err := os.ReadFile(srcGifPath)
	if err != nil {
		t.Skipf("Skipping test - no test GIF available: %v", err)
	}

	animPath := filepath.Join(tempDir, "test.gif")
	err = os.WriteFile(animPath, testGifData, 0644)
	if err != nil {
		t.Fatalf("Failed to create test GIF: %v", err)
	} // Test LoadAnimations function
	animations, err := LoadAnimations(testCard, tempDir)
	if err != nil {
		t.Fatalf("LoadAnimations failed: %v", err)
	}

	if len(animations) != 1 {
		t.Errorf("Expected 1 animation, got %d", len(animations))
	}

	if _, exists := animations["test"]; !exists {
		t.Error("Expected 'test' animation to be loaded")
	}

	if len(animations["test"]) != len(testGifData) {
		t.Errorf("Expected animation data length %d, got %d", len(testGifData), len(animations["test"]))
	}
}

func TestLoadAnimations_InvalidGIF(t *testing.T) {
	// Create temporary test directory
	tempDir := t.TempDir()

	// Create test character card
	testCard := map[string]interface{}{
		"name": "Test Character",
		"animations": map[string]interface{}{
			"invalid": "invalid.gif",
		},
	}

	// Create invalid GIF file
	invalidData := []byte{0x00, 0x01, 0x02, 0x03}
	animPath := filepath.Join(tempDir, "invalid.gif")
	err := os.WriteFile(animPath, invalidData, 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid GIF: %v", err)
	}

	// Test LoadAnimations function with invalid GIF
	animations, err := LoadAnimations(testCard, tempDir)
	if err == nil {
		t.Error("Expected error for invalid GIF, but got none")
	}

	if animations != nil {
		t.Error("Expected nil animations for invalid GIF")
	}
}

func TestLoadAnimations_MissingFile(t *testing.T) {
	// Create temporary test directory
	tempDir := t.TempDir()

	// Create test character card with non-existent file
	testCard := map[string]interface{}{
		"name": "Test Character",
		"animations": map[string]interface{}{
			"missing": "missing.gif",
		},
	}

	// Test LoadAnimations function with missing file
	animations, err := LoadAnimations(testCard, tempDir)
	if err == nil {
		t.Error("Expected error for missing file, but got none")
	}

	if animations != nil {
		t.Error("Expected nil animations for missing file")
	}
}

func TestLoadAnimations_NoAnimations(t *testing.T) {
	// Create test character card without animations
	testCard := map[string]interface{}{
		"name": "Test Character",
	}

	// Test LoadAnimations function with no animations
	animations, err := LoadAnimations(testCard, "")
	if err != nil {
		t.Errorf("Unexpected error for no animations: %v", err)
	}

	if len(animations) != 0 {
		t.Errorf("Expected 0 animations, got %d", len(animations))
	}
}

func TestIsValidGIF(t *testing.T) {
	// Test with real GIF data from assets
	srcGifPath := "../../assets/characters/default/animations/idle.gif"
	validGifData, err := os.ReadFile(srcGifPath)
	if err != nil {
		t.Skipf("Skipping test - no test GIF available: %v", err)
	}

	if !IsValidGIF(validGifData) {
		t.Error("Expected valid GIF to be recognized as valid")
	} // Test invalid GIF data
	invalidData := []byte{0x00, 0x01, 0x02, 0x03}
	if IsValidGIF(invalidData) {
		t.Error("Expected invalid data to be recognized as invalid")
	}

	// Test empty data
	if IsValidGIF([]byte{}) {
		t.Error("Expected empty data to be recognized as invalid")
	}
}

func BenchmarkLoadAnimations(b *testing.B) {
	// Setup benchmark data
	tempDir := b.TempDir()

	testCard := map[string]interface{}{
		"name": "Benchmark Character",
		"animations": map[string]interface{}{
			"test": "test.gif",
		},
	}

	// Create a larger test GIF for more realistic benchmark
	testGifData := make([]byte, 1024)                             // 1KB test file
	copy(testGifData, []byte{0x47, 0x49, 0x46, 0x38, 0x37, 0x61}) // GIF header

	animPath := filepath.Join(tempDir, "test.gif")
	err := os.WriteFile(animPath, testGifData, 0644)
	if err != nil {
		b.Fatalf("Failed to create test GIF: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := LoadAnimations(testCard, tempDir)
		if err != nil {
			b.Fatalf("LoadAnimations failed: %v", err)
		}
	}
}

func BenchmarkIsValidGIF(b *testing.B) {
	// Use real GIF data from assets
	srcGifPath := "../../assets/characters/default/animations/idle.gif"
	validGifData, err := os.ReadFile(srcGifPath)
	if err != nil {
		b.Skipf("Skipping benchmark - no test GIF available: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		IsValidGIF(validGifData)
	}
}
