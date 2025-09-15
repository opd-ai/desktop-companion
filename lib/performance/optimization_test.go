package performance

import (
	"image"
	"image/color"
	"image/gif"
	"testing"
	"time"
)

// TestFrameCache validates LRU frame caching functionality
func TestFrameCache(t *testing.T) {
	cache := NewFrameCache(3) // Small cache for testing

	// Create test images
	img1 := createTestImage(64, 64, color.RGBA{255, 0, 0, 255})
	img2 := createTestImage(64, 64, color.RGBA{0, 255, 0, 255})
	img3 := createTestImage(64, 64, color.RGBA{0, 0, 255, 255})
	img4 := createTestImage(64, 64, color.RGBA{255, 255, 0, 255})

	// Test initial empty cache
	if _, found := cache.Get("key1"); found {
		t.Error("Empty cache should not contain key1")
	}

	// Test adding items
	cache.Put("key1", img1)
	cache.Put("key2", img2)
	cache.Put("key3", img3)

	// Test retrieval
	if cachedImg, found := cache.Get("key1"); !found || cachedImg != img1 {
		t.Error("Should retrieve cached image for key1")
	}

	// Test LRU eviction - key1 was just accessed, so key2 should be evicted
	cache.Put("key4", img4)

	if _, found := cache.Get("key2"); found {
		t.Error("key2 should have been evicted (LRU)")
	}

	if _, found := cache.Get("key1"); !found {
		t.Error("key1 should still be cached (recently accessed)")
	}

	// Test stats
	stats := cache.GetStats()
	if stats.Size != 3 {
		t.Errorf("Expected cache size 3, got %d", stats.Size)
	}
	if stats.MaxSize != 3 {
		t.Errorf("Expected max size 3, got %d", stats.MaxSize)
	}

	// Test clear
	cache.Clear()
	stats = cache.GetStats()
	if stats.Size != 0 {
		t.Error("Cache should be empty after clear")
	}
	if stats.HitCount != 0 || stats.MissCount != 0 {
		t.Error("Stats should be reset after clear")
	}
}

// TestFrameCacheStats validates cache performance metrics
func TestFrameCacheStats(t *testing.T) {
	cache := NewFrameCache(10)
	img := createTestImage(32, 32, color.RGBA{128, 128, 128, 255})

	// Record hits and misses
	cache.Get("nonexistent") // miss
	cache.Put("key1", img)
	cache.Get("key1") // hit
	cache.Get("key1") // hit
	cache.Get("key2") // miss

	stats := cache.GetStats()
	if stats.HitCount != 2 {
		t.Errorf("Expected 2 hits, got %d", stats.HitCount)
	}
	if stats.MissCount != 2 {
		t.Errorf("Expected 2 misses, got %d", stats.MissCount)
	}
	if stats.HitRatio != 0.5 {
		t.Errorf("Expected hit ratio 0.5, got %f", stats.HitRatio)
	}
}

// TestGetCacheKey validates cache key generation
func TestGetCacheKey(t *testing.T) {
	tests := []struct {
		animName   string
		frameIndex int
		width      int
		height     int
		expected   string
	}{
		{"idle", 0, 64, 64, "idle:0:64x64"},
		{"talking", 5, 128, 128, "talking:5:128x128"},
		{"happy", 10, 256, 256, "happy:10:256x256"},
	}

	for _, test := range tests {
		key := GetCacheKey(test.animName, test.frameIndex, test.width, test.height)
		if key != test.expected {
			t.Errorf("Expected key %s, got %s", test.expected, key)
		}
	}
}

// TestOptimizedAnimationManager validates optimized animation performance
func TestOptimizedAnimationManager(t *testing.T) {
	frameCache := NewFrameCache(50)
	oam := NewOptimizedAnimationManager(frameCache, 60)

	// Create test GIF data
	gifData := createTestGIF()
	if gifData == nil || len(gifData.Image) == 0 {
		t.Fatal("Test GIF data is invalid")
	}

	oam.SetAnimation("test", len(gifData.Image), 64, 64)

	// Test frame retrieval
	frame := oam.GetOptimizedFrame(gifData)
	if frame == nil {
		t.Error("Should get frame from GIF data")
	}

	// Wait to allow next frame request (60 FPS = ~16.7ms interval)
	time.Sleep(20 * time.Millisecond)

	// Test cache utilization - second call should hit cache since we use same frame index
	initialStats := oam.GetCacheStats()
	frame2 := oam.GetOptimizedFrame(gifData)
	newStats := oam.GetCacheStats()

	if frame2 == nil {
		t.Error("Should get cached frame")
	}

	// Since we're using the same frame index, this should be a cache hit
	if newStats.HitCount <= initialStats.HitCount {
		t.Logf("Initial hits: %d, new hits: %d", initialStats.HitCount, newStats.HitCount)
		t.Error("Second frame access should increase hit count")
	}

	// Test frame advancement
	oam.AdvanceFrame()
	time.Sleep(20 * time.Millisecond) // Wait for frame rate limit
	frame3 := oam.GetOptimizedFrame(gifData)
	if frame3 == nil {
		t.Error("Should get next frame after advancement")
	}
}

// TestFrameRateOptimizer validates platform-aware frame rate optimization
func TestFrameRateOptimizer(t *testing.T) {
	// Test desktop optimizer
	desktop := NewFrameRateOptimizer(true, false)
	if fps := desktop.GetOptimalFPS(); fps != 60 {
		t.Errorf("Desktop should default to 60 FPS, got %d", fps)
	}

	// Test mobile optimizer
	mobile := NewFrameRateOptimizer(false, true)
	if fps := mobile.GetOptimalFPS(); fps != 30 {
		t.Errorf("Mobile should default to 30 FPS, got %d", fps)
	}

	// Test background state
	desktop.SetBackgroundState(true)
	if fps := desktop.GetOptimalFPS(); fps != 10 {
		t.Errorf("Backgrounded desktop should use 10 FPS, got %d", fps)
	}

	mobile.SetBackgroundState(true)
	if fps := mobile.GetOptimalFPS(); fps != 5 {
		t.Errorf("Backgrounded mobile should use 5 FPS, got %d", fps)
	}

	// Test battery optimization
	mobile.SetBackgroundState(false)
	mobile.SetBatteryLevel(0.1) // Low battery
	if fps := mobile.GetOptimalFPS(); fps >= 30 {
		t.Errorf("Low battery should reduce FPS, got %d", fps)
	}
}

// TestFrameSkipping validates frame skipping logic for performance
func TestFrameSkipping(t *testing.T) {
	frameCache := NewFrameCache(10)
	oam := NewOptimizedAnimationManager(frameCache, 30) // 30 FPS target

	gifData := createTestGIF()
	oam.SetAnimation("skiptest", len(gifData.Image), 64, 64)

	// First frame should not be skipped
	frame1 := oam.GetOptimizedFrame(gifData)
	if frame1 == nil {
		t.Error("First frame should not be skipped")
	}

	// Immediate second call should be skipped (not enough time elapsed)
	frame2 := oam.GetOptimizedFrame(gifData)
	if frame2 != nil {
		t.Error("Second immediate frame should be skipped for performance")
	}

	// After waiting, frame should be available
	time.Sleep(35 * time.Millisecond) // Wait longer than 30 FPS interval
	frame3 := oam.GetOptimizedFrame(gifData)
	if frame3 == nil {
		t.Error("Frame should be available after waiting")
	}
}

// TestCacheConcurrency validates thread safety of frame cache
func TestCacheConcurrency(t *testing.T) {
	cache := NewFrameCache(100)
	img := createTestImage(32, 32, color.RGBA{255, 255, 255, 255})

	// Run concurrent operations
	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			cache.Put("key"+string(rune('0'+i%10)), img)
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			cache.Get("key" + string(rune('0'+i%10)))
		}
		done <- true
	}()

	// Wait for completion
	<-done
	<-done

	// Should not panic and should have reasonable stats
	stats := cache.GetStats()
	if stats.Size > 100 {
		t.Error("Cache size exceeded maximum")
	}
}

// BenchmarkFrameCache benchmarks cache performance
func BenchmarkFrameCache(b *testing.B) {
	cache := NewFrameCache(1000)
	img := createTestImage(64, 64, color.RGBA{100, 100, 100, 255})

	// Pre-populate cache
	for i := 0; i < 100; i++ {
		cache.Put(GetCacheKey("bench", i, 64, 64), img)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := GetCacheKey("bench", i%100, 64, 64)
			cache.Get(key)
			i++
		}
	})
}

// BenchmarkOptimizedAnimationManager benchmarks optimized frame retrieval
func BenchmarkOptimizedAnimationManager(b *testing.B) {
	frameCache := NewFrameCache(1000)
	oam := NewOptimizedAnimationManager(frameCache, 60)
	gifData := createTestGIF()

	oam.SetAnimation("benchmark", len(gifData.Image), 64, 64)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		oam.GetOptimizedFrame(gifData)
		if i%len(gifData.Image) == 0 {
			oam.AdvanceFrame()
		}
	}
}

// BenchmarkCacheKeyGeneration benchmarks cache key creation performance
func BenchmarkCacheKeyGeneration(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			GetCacheKey("animation", i%10, 64+i%64, 64+i%64)
			i++
		}
	})
}

// Helper functions for testing

// createTestImage creates a simple test image with specified dimensions and color
func createTestImage(width, height int, c color.RGBA) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

// createTestGIF creates a simple test GIF with multiple frames
func createTestGIF() *gif.GIF {
	// Create 3 frames with different colors
	frame1 := image.NewPaletted(image.Rect(0, 0, 64, 64), color.Palette{
		color.RGBA{255, 0, 0, 255}, // Red
		color.RGBA{0, 0, 0, 0},     // Transparent
	})
	frame2 := image.NewPaletted(image.Rect(0, 0, 64, 64), color.Palette{
		color.RGBA{0, 255, 0, 255}, // Green
		color.RGBA{0, 0, 0, 0},     // Transparent
	})
	frame3 := image.NewPaletted(image.Rect(0, 0, 64, 64), color.Palette{
		color.RGBA{0, 0, 255, 255}, // Blue
		color.RGBA{0, 0, 0, 0},     // Transparent
	})

	// Fill frames with their respective colors
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			frame1.SetColorIndex(x, y, 0)
			frame2.SetColorIndex(x, y, 0)
			frame3.SetColorIndex(x, y, 0)
		}
	}

	return &gif.GIF{
		Image: []*image.Paletted{frame1, frame2, frame3},
		Delay: []int{10, 10, 10}, // 100ms per frame
	}
}
