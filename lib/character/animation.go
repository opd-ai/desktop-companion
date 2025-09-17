package character

import (
	"fmt"
	"image"
	"image/gif"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// AnimationManager handles GIF animation playback using Go's standard library
// This follows the "lazy programmer" principle - leveraging built-in image/gif
// instead of writing custom GIF decoders
type AnimationManager struct {
	mu          sync.RWMutex
	animations  map[string]*gif.GIF // Loaded GIF animations
	currentAnim string              // Currently playing animation name
	frameIndex  int                 // Current frame index
	lastUpdate  time.Time           // Last frame update time
	playing     bool                // Whether animation is playing
}

// NewAnimationManager creates a new animation manager
func NewAnimationManager() *AnimationManager {
	caller := getCaller()
	logrus.WithFields(logrus.Fields{
		"caller": caller,
	}).Info("Creating new animation manager")

	manager := &AnimationManager{
		animations: make(map[string]*gif.GIF),
		playing:    true,
		lastUpdate: time.Now(),
	}

	logrus.WithFields(logrus.Fields{
		"caller":  caller,
		"playing": manager.playing,
	}).Info("Animation manager created successfully")

	return manager
}

// LoadAnimation loads a GIF animation from file using standard library
// Returns error if file cannot be loaded or is not a valid GIF
func (am *AnimationManager) LoadAnimation(name, filepath string) error {
	caller := getCaller()
	logrus.WithFields(logrus.Fields{
		"caller":   caller,
		"name":     name,
		"filepath": filepath,
	}).Info("Loading GIF animation from file")

	am.mu.Lock()
	defer am.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"caller":   caller,
		"filepath": filepath,
	}).Debug("Opening animation file")

	file, err := os.Open(filepath)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"caller":   caller,
			"filepath": filepath,
			"error":    err.Error(),
		}).Error("Failed to open animation file")
		return fmt.Errorf("failed to open animation file %s: %w", filepath, err)
	}
	defer file.Close()

	logrus.WithFields(logrus.Fields{
		"caller": caller,
	}).Debug("Decoding GIF animation data")

	// Use standard library gif.DecodeAll - no external dependencies
	gifData, err := gif.DecodeAll(file)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"caller":   caller,
			"filepath": filepath,
			"error":    err.Error(),
		}).Error("Failed to decode GIF animation")
		return fmt.Errorf("failed to decode GIF %s: %w", filepath, err)
	}

	// Validate GIF has frames
	if len(gifData.Image) == 0 {
		logrus.WithFields(logrus.Fields{
			"caller":   caller,
			"filepath": filepath,
		}).Error("GIF animation contains no frames")
		return fmt.Errorf("GIF file %s contains no frames", filepath)
	}

	logrus.WithFields(logrus.Fields{
		"caller":     caller,
		"name":       name,
		"frameCount": len(gifData.Image),
	}).Debug("GIF animation decoded successfully")

	am.animations[name] = gifData

	// Set as current animation if this is the first one loaded
	if am.currentAnim == "" {
		am.currentAnim = name
		logrus.WithFields(logrus.Fields{
			"caller": caller,
			"name":   name,
		}).Debug("Set as current animation (first loaded)")
	}

	logrus.WithFields(logrus.Fields{
		"caller":          caller,
		"name":            name,
		"totalAnimations": len(am.animations),
	}).Info("Animation loaded successfully")

	return nil
}

// LoadEmbeddedAnimation loads a pre-decoded GIF animation into the manager
// This method supports embedded asset loading for standalone binaries
// Uses standard library gif.GIF structure - no additional dependencies
func (am *AnimationManager) LoadEmbeddedAnimation(name string, gifData *gif.GIF) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if len(gifData.Image) == 0 {
		return fmt.Errorf("embedded GIF animation %s contains no frames", name)
	}

	am.animations[name] = gifData

	// Set as current animation if this is the first one loaded
	if am.currentAnim == "" {
		am.currentAnim = name
	}

	return nil
}

// SetCurrentAnimation switches to a different loaded animation
func (am *AnimationManager) SetCurrentAnimation(name string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if _, exists := am.animations[name]; !exists {
		return fmt.Errorf("animation '%s' not loaded", name)
	}

	am.currentAnim = name
	am.frameIndex = 0
	am.lastUpdate = time.Now()

	return nil
}

// GetCurrentFrame returns the current frame image and whether a new frame is available
// This method provides timing information without modifying state
// Frame advancement timing should be handled by calling Update() regularly
func (am *AnimationManager) GetCurrentFrame() (image.Image, bool) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	if !am.playing || am.currentAnim == "" {
		return nil, false
	}

	currentGif, exists := am.animations[am.currentAnim]
	if !exists || len(currentGif.Image) == 0 {
		return nil, false
	}

	// Calculate frame timing to determine if update is needed
	frameDelay := time.Duration(currentGif.Delay[am.frameIndex]) * 10 * time.Millisecond
	if frameDelay == 0 {
		frameDelay = 100 * time.Millisecond // Default to 10 FPS
	}

	// Check if enough time has passed for next frame (read-only timing check)
	needsUpdate := time.Since(am.lastUpdate) >= frameDelay

	return currentGif.Image[am.frameIndex], needsUpdate
}

// GetCurrentFrameImage returns just the current frame without timing logic
// Useful for rendering when timing is handled externally
func (am *AnimationManager) GetCurrentFrameImage() image.Image {
	am.mu.RLock()
	defer am.mu.RUnlock()

	if am.currentAnim == "" {
		return nil
	}

	currentGif, exists := am.animations[am.currentAnim]
	if !exists || len(currentGif.Image) == 0 {
		return nil
	}

	return currentGif.Image[am.frameIndex]
}

// Update advances animation frames based on timing
// Call this regularly (e.g., in a render loop) to maintain smooth animation
func (am *AnimationManager) Update() bool {
	am.mu.Lock()
	defer am.mu.Unlock()

	if !am.playing || am.currentAnim == "" {
		return false
	}

	currentGif, exists := am.animations[am.currentAnim]
	if !exists || len(currentGif.Image) == 0 {
		return false
	}

	// Calculate frame timing
	frameDelay := time.Duration(currentGif.Delay[am.frameIndex]) * 10 * time.Millisecond
	if frameDelay == 0 {
		frameDelay = 100 * time.Millisecond // Default to 10 FPS
	}

	// Advance frame if enough time has passed
	if time.Since(am.lastUpdate) >= frameDelay {
		am.frameIndex = (am.frameIndex + 1) % len(currentGif.Image)
		am.lastUpdate = time.Now()
		return true // Frame changed
	}

	return false // No frame change
}

// Play starts animation playback
func (am *AnimationManager) Play() {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.playing = true
}

// Pause stops animation playback
func (am *AnimationManager) Pause() {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.playing = false
}

// IsPlaying returns whether animation is currently playing
func (am *AnimationManager) IsPlaying() bool {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.playing
}

// GetCurrentAnimationName returns the name of currently playing animation
func (am *AnimationManager) GetCurrentAnimationName() string {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.currentAnim
}

// GetLoadedAnimations returns a list of all loaded animation names
func (am *AnimationManager) GetLoadedAnimations() []string {
	am.mu.RLock()
	defer am.mu.RUnlock()

	names := make([]string, 0, len(am.animations))
	for name := range am.animations {
		names = append(names, name)
	}
	return names
}

// GetAnimationFrameCount returns the number of frames in the specified animation
func (am *AnimationManager) GetAnimationFrameCount(name string) int {
	am.mu.RLock()
	defer am.mu.RUnlock()

	if anim, exists := am.animations[name]; exists {
		return len(anim.Image)
	}
	return 0
}

// Reset resets the current animation to the first frame
func (am *AnimationManager) Reset() {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.frameIndex = 0
	am.lastUpdate = time.Now()
}
