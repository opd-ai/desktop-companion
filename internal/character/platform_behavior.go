// Package character provides platform-aware behavior adaptation for cross-platform compatibility.
// This module adapts character interactions, animations, and performance based on the detected platform,
// ensuring optimal user experience on both desktop and mobile devices.
package character

import (
	"time"

	"desktop-companion/internal/platform"
)

// PlatformBehaviorAdapter manages platform-specific behavior adaptations for characters.
// It uses the platform detection system to modify character behavior, interaction patterns,
// and performance characteristics based on the target platform capabilities.
type PlatformBehaviorAdapter struct {
	platform *platform.PlatformInfo
}

// BehaviorConfig contains platform-specific behavior configuration.
// These settings adapt character behavior to be appropriate for the target platform,
// considering factors like input methods, performance constraints, and user expectations.
type BehaviorConfig struct {
	// Interaction settings
	IdleTimeout         time.Duration // How long before character returns to idle state
	InteractionCooldown time.Duration // Minimum time between interactions
	MovementEnabled     bool          // Whether character can be moved/dragged

	// Animation settings
	AnimationFrameRate  int  // Target FPS for animations (lower on mobile for battery)
	AnimationCacheSize  int  // Number of frames to cache (smaller on mobile)
	AutoPauseAnimations bool // Pause animations when app is backgrounded

	// Interaction feedback
	HapticFeedback bool // Enable haptic feedback for touch interactions
	VisualFeedback bool // Enable visual feedback for interactions
	AudioFeedback  bool // Enable audio feedback (may be disabled on mobile)

	// Performance settings
	BackgroundFPS       int  // FPS when app is in background
	MemoryOptimization  bool // Enable memory optimization for mobile
	BatteryOptimization bool // Enable battery-saving optimizations
}

// NewPlatformBehaviorAdapter creates a new platform-aware behavior adapter.
// It uses the existing platform detection system to determine appropriate behavior settings.
func NewPlatformBehaviorAdapter(platformInfo *platform.PlatformInfo) *PlatformBehaviorAdapter {
	// Handle nil platform info gracefully (fallback to desktop behavior)
	if platformInfo == nil {
		platformInfo = &platform.PlatformInfo{
			OS:           "unknown",
			FormFactor:   "desktop",
			InputMethods: []string{"mouse", "keyboard"},
		}
	}

	return &PlatformBehaviorAdapter{
		platform: platformInfo,
	}
}

// GetBehaviorConfig returns platform-optimized behavior configuration.
// Desktop platforms prioritize responsiveness and full features, while mobile platforms
// focus on battery life, touch interaction, and performance optimization.
func (pba *PlatformBehaviorAdapter) GetBehaviorConfig() *BehaviorConfig {
	if pba.platform.IsMobile() {
		return pba.getMobileBehaviorConfig()
	}
	return pba.getDesktopBehaviorConfig()
}

// getDesktopBehaviorConfig returns behavior settings optimized for desktop platforms.
// Desktop behavior prioritizes responsiveness and full feature availability.
func (pba *PlatformBehaviorAdapter) getDesktopBehaviorConfig() *BehaviorConfig {
	return &BehaviorConfig{
		// Desktop: Responsive interactions with mouse precision
		IdleTimeout:         30 * time.Second,
		InteractionCooldown: 1 * time.Second,
		MovementEnabled:     true, // Mouse drag works well

		// Desktop: Full animation quality for better hardware
		AnimationFrameRate:  60,    // Smooth 60 FPS
		AnimationCacheSize:  100,   // Larger cache for smooth playback
		AutoPauseAnimations: false, // Don't pause animations

		// Desktop: All feedback types available
		HapticFeedback: false, // Most desktops don't have haptic feedback
		VisualFeedback: true,  // Mouse cursors and visual highlights
		AudioFeedback:  true,  // Desktop audio is always available

		// Desktop: Performance optimized for plugged-in systems
		BackgroundFPS:       30,    // Maintain some FPS when minimized
		MemoryOptimization:  false, // More memory available
		BatteryOptimization: false, // Power is not a concern
	}
}

// getMobileBehaviorConfig returns behavior settings optimized for mobile platforms.
// Mobile behavior prioritizes battery life, touch interaction, and performance.
func (pba *PlatformBehaviorAdapter) getMobileBehaviorConfig() *BehaviorConfig {
	return &BehaviorConfig{
		// Mobile: Longer timeouts for touch-based interaction
		IdleTimeout:         45 * time.Second, // Longer idle for mobile
		InteractionCooldown: 2 * time.Second,  // Prevent accidental rapid taps
		MovementEnabled:     false,            // Touch drag can be problematic

		// Mobile: Reduced animation quality for battery life
		AnimationFrameRate:  30,   // 30 FPS saves battery
		AnimationCacheSize:  50,   // Smaller cache for memory constraints
		AutoPauseAnimations: true, // Pause when app is backgrounded

		// Mobile: Touch-optimized feedback
		HapticFeedback: pba.platform.HasTouch(), // Enable haptic if touch available
		VisualFeedback: true,                    // Visual feedback important for touch
		AudioFeedback:  false,                   // Audio may disturb others on mobile

		// Mobile: Aggressive optimization for battery and memory
		BackgroundFPS:       5,    // Minimal FPS when backgrounded
		MemoryOptimization:  true, // Enable memory optimization
		BatteryOptimization: true, // Enable battery optimization
	}
}

// GetOptimalCharacterSize returns the optimal character size for the current platform.
// This integrates with the existing responsive layout system while providing character-specific logic.
func (pba *PlatformBehaviorAdapter) GetOptimalCharacterSize(screenWidth float32, defaultSize int) int {
	if pba.platform.IsMobile() {
		// Mobile: Use 25% of screen width for touch-friendly size
		// Ensure minimum size for visibility and maximum for performance
		mobileSize := int(screenWidth * 0.25)
		if mobileSize < 96 { // Minimum touch target size
			return 96
		}
		if mobileSize > 256 { // Maximum for performance
			return 256
		}
		return mobileSize
	}

	// Desktop: Use provided default size or reasonable fallback
	if defaultSize <= 0 {
		return 128 // Desktop default
	}
	return defaultSize
}

// GetInteractionDelayForEvent returns platform-appropriate delay for specific interaction types.
// This helps prevent accidental interactions on touch devices while maintaining responsiveness on desktop.
func (pba *PlatformBehaviorAdapter) GetInteractionDelayForEvent(eventType string) time.Duration {
	baseConfig := pba.GetBehaviorConfig()

	switch eventType {
	case "click", "tap":
		return baseConfig.InteractionCooldown
	case "rightclick", "longpress":
		// Longer cooldown for context menu actions
		return baseConfig.InteractionCooldown * 2
	case "doubleclick", "doubletap":
		// Shorter cooldown for double actions
		return baseConfig.InteractionCooldown / 2
	default:
		return baseConfig.InteractionCooldown
	}
}

// ShouldEnableFeature returns whether a specific feature should be enabled for the current platform.
// This allows character cards to conditionally enable features based on platform capabilities.
func (pba *PlatformBehaviorAdapter) ShouldEnableFeature(feature string) bool {
	switch feature {
	case "movement", "dragging":
		return pba.GetBehaviorConfig().MovementEnabled
	case "haptic":
		return pba.GetBehaviorConfig().HapticFeedback
	case "audio":
		return pba.GetBehaviorConfig().AudioFeedback
	case "background_animation":
		return !pba.GetBehaviorConfig().AutoPauseAnimations
	case "memory_optimization":
		return pba.GetBehaviorConfig().MemoryOptimization
	case "battery_optimization":
		return pba.GetBehaviorConfig().BatteryOptimization
	default:
		return true // Default to enabled for unknown features
	}
}

// GetAnimationFrameRate returns the optimal frame rate for animations on the current platform.
func (pba *PlatformBehaviorAdapter) GetAnimationFrameRate() int {
	return pba.GetBehaviorConfig().AnimationFrameRate
}

// GetAnimationCacheSize returns the optimal animation cache size for the current platform.
func (pba *PlatformBehaviorAdapter) GetAnimationCacheSize() int {
	return pba.GetBehaviorConfig().AnimationCacheSize
}

// GetBackgroundFrameRate returns the frame rate to use when the app is in the background.
func (pba *PlatformBehaviorAdapter) GetBackgroundFrameRate() int {
	return pba.GetBehaviorConfig().BackgroundFPS
}
