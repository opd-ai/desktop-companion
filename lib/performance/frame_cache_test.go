package performance

import (
	"image"
	"image/color"
	"image/gif"
	"testing"
	"time"
)

// TestFrameCacheBasic validates basic frame cache operations
func TestFrameCacheBasic(t *testing.T) {
	cache := NewFrameCache(5)

	// Test nil frame handling
	cache.Put("nil", nil)
	if _, found := cache.Get("nil"); found {
		t.Error("Should not cache nil frames")
	}

	// Test zero size cache
	zeroCache := NewFrameCache(0)
	img := createTestImage(32, 32, color.RGBA{255, 0, 0, 255})
	zeroCache.Put("test", img)

	stats := zeroCache.GetStats()
	if stats.MaxSize <= 0 {
		t.Error("Zero size cache should default to reasonable size")
	}
}

// TestFrameCacheEdgeCases tests edge cases and error conditions
func TestFrameCacheEdgeCases(t *testing.T) {
	cache := NewFrameCache(2)
	img1 := createTestImage(16, 16, color.RGBA{255, 0, 0, 255})
	img2 := createTestImage(16, 16, color.RGBA{0, 255, 0, 255})

	// Test updating existing key
	cache.Put("update", img1)
	cache.Put("update", img2) // Should update, not add new entry

	if cached, found := cache.Get("update"); !found || cached != img2 {
		t.Error("Should update existing cache entry")
	}

	stats := cache.GetStats()
	if stats.Size != 1 {
		t.Errorf("Cache should have 1 entry after update, got %d", stats.Size)
	}
}

// TestIntToString validates the custom integer to string conversion
func TestIntToString(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{123, "123"},
		{-1, "-1"},
		{-123, "-123"},
		{999999, "999999"},
	}

	for _, test := range tests {
		result := intToString(test.input)
		if result != test.expected {
			t.Errorf("intToString(%d) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

// TestFrameRateOptimizerEdgeCases tests optimizer edge cases
func TestFrameRateOptimizerEdgeCases(t *testing.T) {
	// Test invalid battery levels
	optimizer := NewFrameRateOptimizer(false, true)
	optimizer.SetBatteryLevel(-0.5) // Invalid negative
	optimizer.SetBatteryLevel(1.5)  // Invalid > 1.0

	// Should maintain reasonable FPS despite invalid inputs
	fps := optimizer.GetOptimalFPS()
	if fps <= 0 || fps > 120 {
		t.Errorf("FPS should be reasonable despite invalid battery input, got %d", fps)
	}

	// Test unknown platform (neither desktop nor mobile)
	unknown := NewFrameRateOptimizer(false, false)
	fps = unknown.GetOptimalFPS()
	if fps <= 0 || fps > 120 {
		t.Errorf("Unknown platform should have reasonable default FPS, got %d", fps)
	}
}

// TestOptimizedAnimationManagerEdgeCases tests edge cases for animation manager
func TestOptimizedAnimationManagerEdgeCases(t *testing.T) {
	frameCache := NewFrameCache(10)

	// Test invalid FPS
	oam := NewOptimizedAnimationManager(frameCache, 0)
	if oam.targetFPS <= 0 {
		t.Error("Should default to reasonable FPS when given invalid value")
	}

	// Test nil GIF data
	oam.SetAnimation("test", 3, 64, 64)
	frame := oam.GetOptimizedFrame(nil)
	if frame != nil {
		t.Error("Should return nil for nil GIF data")
	}

	// Test empty GIF
	emptyGIF := createEmptyGIF()
	frame = oam.GetOptimizedFrame(emptyGIF)
	if frame != nil {
		t.Error("Should return nil for empty GIF")
	}
}

// TestFrameSkippingPrecision validates precise frame timing
func TestFrameSkippingPrecision(t *testing.T) {
	frameCache := NewFrameCache(10)
	oam := NewOptimizedAnimationManager(frameCache, 100) // 100 FPS = 10ms intervals

	gifData := createTestGIF()
	oam.SetAnimation("precision", len(gifData.Image), 32, 32)

	// First frame should be immediate
	frame1 := oam.GetOptimizedFrame(gifData)
	if frame1 == nil {
		t.Error("First frame should be immediate")
	}

	// Frames within 10ms should be skipped
	for i := 0; i < 5; i++ {
		frame := oam.GetOptimizedFrame(gifData)
		if frame != nil {
			t.Error("Immediate frames should be skipped for high FPS target")
		}
	}

	// After waiting 12ms, frame should be available
	time.Sleep(12 * time.Millisecond)
	frame2 := oam.GetOptimizedFrame(gifData)
	if frame2 == nil {
		t.Error("Frame should be available after waiting")
	}
}

// TestCacheMemoryEfficiency validates memory usage patterns
func TestCacheMemoryEfficiency(t *testing.T) {
	cache := NewFrameCache(1000)

	// Add many frames and verify LRU eviction
	for i := 0; i < 1500; i++ {
		img := createTestImage(8, 8, color.RGBA{uint8(i % 255), 0, 0, 255})
		cache.Put(GetCacheKey("mem", i, 8, 8), img)
	}

	stats := cache.GetStats()
	if stats.Size > 1000 {
		t.Errorf("Cache should not exceed max size, got %d", stats.Size)
	}

	// Oldest entries should be evicted
	if _, found := cache.Get(GetCacheKey("mem", 0, 8, 8)); found {
		t.Error("Oldest entry should have been evicted")
	}

	// Recent entries should still be cached
	if _, found := cache.Get(GetCacheKey("mem", 1499, 8, 8)); !found {
		t.Error("Recent entry should still be cached")
	}
}

// Helper function for edge case testing
func createEmptyGIF() *gif.GIF {
	return &gif.GIF{
		Image: []*image.Paletted{},
		Delay: []int{},
	}
}
