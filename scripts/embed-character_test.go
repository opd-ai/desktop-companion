package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/opd-ai/desktop-companion/internal/embedding"
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

	// Create test GIF file (minimal valid GIF generated using Go's standard library)
	testGifData := []byte{
		71, 73, 70, 56, 57, 97, 1, 0, 1, 0, 128, 0, 0, 0, 0, 0, 255, 255, 255, 44, 0, 0, 0, 0,
		1, 0, 1, 0, 0, 2, 2, 68, 1, 0, 59,
	}

	animPath := filepath.Join(tempDir, "test.gif")
	err := os.WriteFile(animPath, testGifData, 0644)
	if err != nil {
		t.Fatalf("Failed to create test GIF: %v", err)
	}

	// Test LoadAnimations function
	animations, err := embedding.LoadAnimations(testCard, tempDir)
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
	animations, err := embedding.LoadAnimations(testCard, tempDir)
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
	animations, err := embedding.LoadAnimations(testCard, tempDir)
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
	animations, err := embedding.LoadAnimations(testCard, "")
	if err != nil {
		t.Errorf("Unexpected error for no animations: %v", err)
	}

	if len(animations) != 0 {
		t.Errorf("Expected 0 animations, got %d", len(animations))
	}
}

func TestIsValidGIF(t *testing.T) {
	// Test valid GIF data - generated using Go's standard library
	validGifData := []byte{
		71, 73, 70, 56, 57, 97, 1, 0, 1, 0, 128, 0, 0, 0, 0, 0, 255, 255, 255, 44, 0, 0, 0, 0,
		1, 0, 1, 0, 0, 2, 2, 68, 1, 0, 59,
	}

	if !embedding.IsValidGIF(validGifData) {
		t.Error("Expected valid GIF to be recognized as valid")
	}

	// Test invalid GIF data
	invalidData := []byte{0x00, 0x01, 0x02, 0x03}
	if embedding.IsValidGIF(invalidData) {
		t.Error("Expected invalid data to be recognized as invalid")
	}

	// Test empty data
	if embedding.IsValidGIF([]byte{}) {
		t.Error("Expected empty data to be recognized as invalid")
	}
}

func TestGenerateEmbeddedCharacter_Integration(t *testing.T) {
	// This would require actual character assets, so we'll skip for now
	// In a real scenario, you'd test with a known character like "default"
	t.Skip("Integration test requires actual character assets")
}

// Helper function to check if string contains substring (for compatibility with older Go versions)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || indexOfSubstring(s, substr) >= 0)
}

func indexOfSubstring(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	if len(substr) > len(s) {
		return -1
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
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
		_, err := embedding.LoadAnimations(testCard, tempDir)
		if err != nil {
			b.Fatalf("LoadAnimations failed: %v", err)
		}
	}
}

func BenchmarkIsValidGIF(b *testing.B) {
	// Create test data
	validGifData := []byte{
		0x47, 0x49, 0x46, 0x38, 0x37, 0x61, // GIF87a header
		0x01, 0x00, 0x01, 0x00, // 1x1 image
		0x00, 0x00, 0x00, // Global color table
		0x2C, 0x00, 0x00, 0x00, 0x00, // Image descriptor
		0x01, 0x00, 0x01, 0x00, 0x00, // 1x1 image with no color table
		0x02, 0x02, 0x04, 0x01, 0x00, 0x3B, // Image data and trailer
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		embedding.IsValidGIF(validGifData)
	}
}
