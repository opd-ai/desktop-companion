// optimization.go: Frame rate and performance optimization utilities
// Provides enhanced animation manager with caching and platform-aware optimizations

package performance

import (
	"image"
	"image/gif"
	"sync"
	"time"
)

// OptimizedAnimationManager enhances the standard animation manager with performance optimizations.
// Includes frame caching, smart refresh logic, and platform-aware frame rate adjustment.
type OptimizedAnimationManager struct {
	mu            sync.RWMutex
	frameCache    *FrameCache
	targetFPS     int
	lastFrameTime time.Time
	frameInterval time.Duration
	animationName string
	frameIndex    int
	totalFrames   int
	width         int
	height        int
}

// NewOptimizedAnimationManager creates an optimized animation manager.
// frameCache: LRU cache for processed frames
// targetFPS: desired frame rate (60 for desktop, 30 for mobile)
func NewOptimizedAnimationManager(frameCache *FrameCache, targetFPS int) *OptimizedAnimationManager {
	if targetFPS <= 0 {
		targetFPS = 60 // Default to 60 FPS
	}

	return &OptimizedAnimationManager{
		frameCache:    frameCache,
		targetFPS:     targetFPS,
		frameInterval: time.Second / time.Duration(targetFPS),
		// Don't set lastFrameTime here - let first frame be immediate
	}
}

// SetAnimation configures the animation parameters for caching and optimization.
// This should be called when switching to a new animation or changing display size.
func (oam *OptimizedAnimationManager) SetAnimation(name string, totalFrames int, width, height int) {
	oam.mu.Lock()
	defer oam.mu.Unlock()

	oam.animationName = name
	oam.totalFrames = totalFrames
	oam.width = width
	oam.height = height
	oam.frameIndex = 0
}

// GetOptimizedFrame returns the current frame using cache optimization.
// Returns cached frame if available, nil if frame should be skipped for performance.
func (oam *OptimizedAnimationManager) GetOptimizedFrame(gifData *gif.GIF) image.Image {
	oam.mu.Lock()
	defer oam.mu.Unlock()

	// Check if we should skip this frame for performance
	now := time.Now()
	if oam.shouldSkipFrame(now) {
		return nil
	}

	// Generate cache key for current frame
	cacheKey := GetCacheKey(oam.animationName, oam.frameIndex, oam.width, oam.height)

	// Try to get cached frame first
	if cachedFrame, found := oam.frameCache.Get(cacheKey); found {
		oam.lastFrameTime = now
		return cachedFrame
	}

	// Cache miss - process frame and cache it
	if gifData != nil && oam.frameIndex < len(gifData.Image) {
		frame := gifData.Image[oam.frameIndex]

		// Cache the processed frame for future use
		oam.frameCache.Put(cacheKey, frame)
		oam.lastFrameTime = now

		return frame
	}

	return nil
}

// shouldSkipFrame determines if we should skip rendering this frame for performance.
// Implements adaptive frame skipping based on target FPS and system performance.
func (oam *OptimizedAnimationManager) shouldSkipFrame(now time.Time) bool {
	// Don't skip first frame
	if oam.lastFrameTime.IsZero() {
		return false
	}

	// Skip if not enough time has passed for target FPS
	elapsed := now.Sub(oam.lastFrameTime)
	return elapsed < oam.frameInterval
}

// AdvanceFrame moves to the next animation frame.
// Should be called after successfully rendering a frame.
func (oam *OptimizedAnimationManager) AdvanceFrame() {
	oam.mu.Lock()
	defer oam.mu.Unlock()

	if oam.totalFrames > 0 {
		oam.frameIndex = (oam.frameIndex + 1) % oam.totalFrames
	}
}

// SetTargetFPS updates the target frame rate for the animation.
// Used for platform-aware optimization (desktop vs mobile).
func (oam *OptimizedAnimationManager) SetTargetFPS(fps int) {
	oam.mu.Lock()
	defer oam.mu.Unlock()

	if fps > 0 {
		oam.targetFPS = fps
		oam.frameInterval = time.Second / time.Duration(fps)
	}
}

// GetCacheStats returns performance statistics for the frame cache.
func (oam *OptimizedAnimationManager) GetCacheStats() FrameCacheStats {
	if oam.frameCache != nil {
		return oam.frameCache.GetStats()
	}
	return FrameCacheStats{}
}

// ClearCache clears all cached frames and resets performance counters.
// Useful when memory usage needs to be reduced or when switching contexts.
func (oam *OptimizedAnimationManager) ClearCache() {
	if oam.frameCache != nil {
		oam.frameCache.Clear()
	}
}

// FrameRateOptimizer provides platform-aware frame rate optimization.
// Adjusts rendering parameters based on device capabilities and power state.
type FrameRateOptimizer struct {
	mu                 sync.RWMutex
	isDesktop          bool
	isMobile           bool
	isBackgrounded     bool
	batteryLevel       float64
	currentFPS         int
	targetFPS          int
	adaptiveAdjustment bool
}

// NewFrameRateOptimizer creates a platform-aware frame rate optimizer.
func NewFrameRateOptimizer(isDesktop, isMobile bool) *FrameRateOptimizer {
	optimizer := &FrameRateOptimizer{
		isDesktop:          isDesktop,
		isMobile:           isMobile,
		adaptiveAdjustment: true,
		batteryLevel:       1.0, // Assume full battery if unknown
	}

	// Set initial target FPS based on platform
	if isDesktop {
		optimizer.targetFPS = 60
	} else if isMobile {
		optimizer.targetFPS = 30
	} else {
		optimizer.targetFPS = 45 // Conservative default
	}

	optimizer.currentFPS = optimizer.targetFPS
	return optimizer
}

// GetOptimalFPS returns the current optimal frame rate based on platform and power state.
func (fro *FrameRateOptimizer) GetOptimalFPS() int {
	fro.mu.RLock()
	defer fro.mu.RUnlock()

	baseFPS := fro.targetFPS

	// Reduce FPS when backgrounded
	if fro.isBackgrounded {
		if fro.isDesktop {
			return 10 // Desktop background FPS
		} else {
			return 5 // Mobile background FPS
		}
	}

	// Adaptive adjustment based on battery level (mobile only)
	if fro.isMobile && fro.adaptiveAdjustment {
		if fro.batteryLevel < 0.2 { // Low battery
			return baseFPS / 2
		} else if fro.batteryLevel < 0.5 { // Medium battery
			return baseFPS * 3 / 4
		}
	}

	return baseFPS
}

// SetBackgroundState updates whether the application is in background.
// Used to adjust frame rate for power saving.
func (fro *FrameRateOptimizer) SetBackgroundState(isBackground bool) {
	fro.mu.Lock()
	defer fro.mu.Unlock()
	fro.isBackgrounded = isBackground
}

// SetBatteryLevel updates the current battery level for mobile optimization.
// batteryLevel should be between 0.0 (empty) and 1.0 (full).
func (fro *FrameRateOptimizer) SetBatteryLevel(level float64) {
	fro.mu.Lock()
	defer fro.mu.Unlock()

	if level >= 0.0 && level <= 1.0 {
		fro.batteryLevel = level
	}
}

// UpdateCurrentFPS records the actual achieved frame rate for adaptive optimization.
func (fro *FrameRateOptimizer) UpdateCurrentFPS(fps int) {
	fro.mu.Lock()
	defer fro.mu.Unlock()
	fro.currentFPS = fps
}

// GetFrameRateRecommendation returns optimization recommendations.
// Returns target FPS and whether adaptive adjustment should be enabled.
func (fro *FrameRateOptimizer) GetFrameRateRecommendation() (targetFPS int, useAdaptive bool) {
	fro.mu.RLock()
	defer fro.mu.RUnlock()

	return fro.GetOptimalFPS(), fro.adaptiveAdjustment
}
