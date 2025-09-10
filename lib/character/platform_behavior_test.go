package character

import (
	"testing"
	"time"

	"github.com/opd-ai/desktop-companion/internal/platform"
)

// TestNewPlatformBehaviorAdapter tests the creation of platform behavior adapters
func TestNewPlatformBehaviorAdapter(t *testing.T) {
	tests := []struct {
		name         string
		platformInfo *platform.PlatformInfo
		expectNil    bool
	}{
		{
			name: "valid desktop platform",
			platformInfo: &platform.PlatformInfo{
				OS:           "windows",
				FormFactor:   "desktop",
				InputMethods: []string{"mouse", "keyboard"},
			},
			expectNil: false,
		},
		{
			name: "valid mobile platform",
			platformInfo: &platform.PlatformInfo{
				OS:           "android",
				FormFactor:   "mobile",
				InputMethods: []string{"touch"},
			},
			expectNil: false,
		},
		{
			name:         "nil platform info - should handle gracefully",
			platformInfo: nil,
			expectNil:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewPlatformBehaviorAdapter(tt.platformInfo)

			if (adapter == nil) != tt.expectNil {
				t.Errorf("NewPlatformBehaviorAdapter() returned nil = %v, expected nil = %v", adapter == nil, tt.expectNil)
			}

			if adapter != nil && adapter.platform == nil {
				t.Error("PlatformBehaviorAdapter.platform should not be nil")
			}
		})
	}
}

// TestGetBehaviorConfig tests platform-specific behavior configuration
func TestGetBehaviorConfig(t *testing.T) {
	tests := []struct {
		name                        string
		platformInfo                *platform.PlatformInfo
		expectedMovementEnabled     bool
		expectedAnimationFrameRate  int
		expectedHapticFeedback      bool
		expectedBatteryOptimization bool
	}{
		{
			name: "desktop platform configuration",
			platformInfo: &platform.PlatformInfo{
				OS:           "windows",
				FormFactor:   "desktop",
				InputMethods: []string{"mouse", "keyboard"},
			},
			expectedMovementEnabled:     true,
			expectedAnimationFrameRate:  60,
			expectedHapticFeedback:      false,
			expectedBatteryOptimization: false,
		},
		{
			name: "mobile platform configuration",
			platformInfo: &platform.PlatformInfo{
				OS:           "android",
				FormFactor:   "mobile",
				InputMethods: []string{"touch"},
			},
			expectedMovementEnabled:     false,
			expectedAnimationFrameRate:  30,
			expectedHapticFeedback:      true,
			expectedBatteryOptimization: true,
		},
		{
			name: "mobile without touch",
			platformInfo: &platform.PlatformInfo{
				OS:           "android",
				FormFactor:   "mobile",
				InputMethods: []string{"keyboard"}, // No touch
			},
			expectedMovementEnabled:     false,
			expectedAnimationFrameRate:  30,
			expectedHapticFeedback:      false, // No touch, no haptic
			expectedBatteryOptimization: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewPlatformBehaviorAdapter(tt.platformInfo)
			config := adapter.GetBehaviorConfig()

			if config == nil {
				t.Fatal("GetBehaviorConfig() returned nil")
			}

			if config.MovementEnabled != tt.expectedMovementEnabled {
				t.Errorf("MovementEnabled = %v, expected %v", config.MovementEnabled, tt.expectedMovementEnabled)
			}

			if config.AnimationFrameRate != tt.expectedAnimationFrameRate {
				t.Errorf("AnimationFrameRate = %d, expected %d", config.AnimationFrameRate, tt.expectedAnimationFrameRate)
			}

			if config.HapticFeedback != tt.expectedHapticFeedback {
				t.Errorf("HapticFeedback = %v, expected %v", config.HapticFeedback, tt.expectedHapticFeedback)
			}

			if config.BatteryOptimization != tt.expectedBatteryOptimization {
				t.Errorf("BatteryOptimization = %v, expected %v", config.BatteryOptimization, tt.expectedBatteryOptimization)
			}
		})
	}
}

// TestGetDesktopBehaviorConfig tests desktop-specific configuration
func TestGetDesktopBehaviorConfig(t *testing.T) {
	adapter := NewPlatformBehaviorAdapter(&platform.PlatformInfo{
		OS:           "windows",
		FormFactor:   "desktop",
		InputMethods: []string{"mouse", "keyboard"},
	})

	config := adapter.getDesktopBehaviorConfig()

	// Test desktop-specific settings
	if config.IdleTimeout != 30*time.Second {
		t.Errorf("Desktop IdleTimeout = %v, expected %v", config.IdleTimeout, 30*time.Second)
	}

	if config.InteractionCooldown != 1*time.Second {
		t.Errorf("Desktop InteractionCooldown = %v, expected %v", config.InteractionCooldown, 1*time.Second)
	}

	if !config.MovementEnabled {
		t.Error("Desktop MovementEnabled should be true")
	}

	if config.AnimationFrameRate != 60 {
		t.Errorf("Desktop AnimationFrameRate = %d, expected 60", config.AnimationFrameRate)
	}

	if config.BackgroundFPS != 30 {
		t.Errorf("Desktop BackgroundFPS = %d, expected 30", config.BackgroundFPS)
	}

	if config.MemoryOptimization {
		t.Error("Desktop MemoryOptimization should be false")
	}

	if config.BatteryOptimization {
		t.Error("Desktop BatteryOptimization should be false")
	}
}

// TestGetMobileBehaviorConfig tests mobile-specific configuration
func TestGetMobileBehaviorConfig(t *testing.T) {
	adapter := NewPlatformBehaviorAdapter(&platform.PlatformInfo{
		OS:           "android",
		FormFactor:   "mobile",
		InputMethods: []string{"touch"},
	})

	config := adapter.getMobileBehaviorConfig()

	// Test mobile-specific settings
	if config.IdleTimeout != 45*time.Second {
		t.Errorf("Mobile IdleTimeout = %v, expected %v", config.IdleTimeout, 45*time.Second)
	}

	if config.InteractionCooldown != 2*time.Second {
		t.Errorf("Mobile InteractionCooldown = %v, expected %v", config.InteractionCooldown, 2*time.Second)
	}

	if config.MovementEnabled {
		t.Error("Mobile MovementEnabled should be false")
	}

	if config.AnimationFrameRate != 30 {
		t.Errorf("Mobile AnimationFrameRate = %d, expected 30", config.AnimationFrameRate)
	}

	if config.BackgroundFPS != 5 {
		t.Errorf("Mobile BackgroundFPS = %d, expected 5", config.BackgroundFPS)
	}

	if !config.MemoryOptimization {
		t.Error("Mobile MemoryOptimization should be true")
	}

	if !config.BatteryOptimization {
		t.Error("Mobile BatteryOptimization should be true")
	}

	if !config.AutoPauseAnimations {
		t.Error("Mobile AutoPauseAnimations should be true")
	}
}

// TestGetOptimalCharacterSize tests character size calculation for different platforms
func TestGetOptimalCharacterSize(t *testing.T) {
	tests := []struct {
		name         string
		platformInfo *platform.PlatformInfo
		screenWidth  float32
		defaultSize  int
		expectedMin  int
		expectedMax  int
	}{
		{
			name: "desktop with default size",
			platformInfo: &platform.PlatformInfo{
				OS:         "windows",
				FormFactor: "desktop",
			},
			screenWidth: 1920,
			defaultSize: 128,
			expectedMin: 128,
			expectedMax: 128,
		},
		{
			name: "desktop with zero default size",
			platformInfo: &platform.PlatformInfo{
				OS:         "linux",
				FormFactor: "desktop",
			},
			screenWidth: 1920,
			defaultSize: 0,
			expectedMin: 128,
			expectedMax: 128,
		},
		{
			name: "mobile with normal screen",
			platformInfo: &platform.PlatformInfo{
				OS:         "android",
				FormFactor: "mobile",
			},
			screenWidth: 400, // 25% = 100, should return 100
			defaultSize: 64,  // Ignored on mobile
			expectedMin: 100,
			expectedMax: 100,
		},
		{
			name: "mobile with small screen",
			platformInfo: &platform.PlatformInfo{
				OS:         "android",
				FormFactor: "mobile",
			},
			screenWidth: 300, // 25% = 75, should return minimum 96
			defaultSize: 64,
			expectedMin: 96,
			expectedMax: 96,
		},
		{
			name: "mobile with large screen",
			platformInfo: &platform.PlatformInfo{
				OS:         "android",
				FormFactor: "mobile",
			},
			screenWidth: 1200, // 25% = 300, should return maximum 256
			defaultSize: 64,
			expectedMin: 256,
			expectedMax: 256,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewPlatformBehaviorAdapter(tt.platformInfo)
			size := adapter.GetOptimalCharacterSize(tt.screenWidth, tt.defaultSize)

			if size < tt.expectedMin || size > tt.expectedMax {
				t.Errorf("GetOptimalCharacterSize() = %d, expected between %d and %d", size, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

// TestGetInteractionDelayForEvent tests interaction delay calculation
func TestGetInteractionDelayForEvent(t *testing.T) {
	adapter := NewPlatformBehaviorAdapter(&platform.PlatformInfo{
		OS:         "android",
		FormFactor: "mobile",
	})

	baseDelay := adapter.GetBehaviorConfig().InteractionCooldown // 2 seconds for mobile

	tests := []struct {
		eventType     string
		expectedDelay time.Duration
	}{
		{"click", baseDelay},
		{"tap", baseDelay},
		{"rightclick", baseDelay * 2},
		{"longpress", baseDelay * 2},
		{"doubleclick", baseDelay / 2},
		{"doubletap", baseDelay / 2},
		{"unknown", baseDelay},
	}

	for _, tt := range tests {
		t.Run(tt.eventType, func(t *testing.T) {
			delay := adapter.GetInteractionDelayForEvent(tt.eventType)
			if delay != tt.expectedDelay {
				t.Errorf("GetInteractionDelayForEvent(%s) = %v, expected %v", tt.eventType, delay, tt.expectedDelay)
			}
		})
	}
}

// TestShouldEnableFeature tests feature enablement logic
func TestShouldEnableFeature(t *testing.T) {
	desktopAdapter := NewPlatformBehaviorAdapter(&platform.PlatformInfo{
		OS:         "windows",
		FormFactor: "desktop",
	})

	mobileAdapter := NewPlatformBehaviorAdapter(&platform.PlatformInfo{
		OS:           "android",
		FormFactor:   "mobile",
		InputMethods: []string{"touch"},
	})

	tests := []struct {
		feature         string
		desktopExpected bool
		mobileExpected  bool
	}{
		{"movement", true, false},
		{"dragging", true, false},
		{"haptic", false, true},
		{"audio", true, false},
		{"background_animation", true, false},
		{"memory_optimization", false, true},
		{"battery_optimization", false, true},
		{"unknown_feature", true, true}, // Default to enabled
	}

	for _, tt := range tests {
		t.Run(tt.feature, func(t *testing.T) {
			desktopResult := desktopAdapter.ShouldEnableFeature(tt.feature)
			if desktopResult != tt.desktopExpected {
				t.Errorf("Desktop ShouldEnableFeature(%s) = %v, expected %v", tt.feature, desktopResult, tt.desktopExpected)
			}

			mobileResult := mobileAdapter.ShouldEnableFeature(tt.feature)
			if mobileResult != tt.mobileExpected {
				t.Errorf("Mobile ShouldEnableFeature(%s) = %v, expected %v", tt.feature, mobileResult, tt.mobileExpected)
			}
		})
	}
}

// TestAnimationSettings tests animation-related methods
func TestAnimationSettings(t *testing.T) {
	desktopAdapter := NewPlatformBehaviorAdapter(&platform.PlatformInfo{
		OS:         "windows",
		FormFactor: "desktop",
	})

	mobileAdapter := NewPlatformBehaviorAdapter(&platform.PlatformInfo{
		OS:         "android",
		FormFactor: "mobile",
	})

	// Test frame rates
	if desktopAdapter.GetAnimationFrameRate() != 60 {
		t.Errorf("Desktop GetAnimationFrameRate() = %d, expected 60", desktopAdapter.GetAnimationFrameRate())
	}

	if mobileAdapter.GetAnimationFrameRate() != 30 {
		t.Errorf("Mobile GetAnimationFrameRate() = %d, expected 30", mobileAdapter.GetAnimationFrameRate())
	}

	// Test cache sizes
	if desktopAdapter.GetAnimationCacheSize() != 100 {
		t.Errorf("Desktop GetAnimationCacheSize() = %d, expected 100", desktopAdapter.GetAnimationCacheSize())
	}

	if mobileAdapter.GetAnimationCacheSize() != 50 {
		t.Errorf("Mobile GetAnimationCacheSize() = %d, expected 50", mobileAdapter.GetAnimationCacheSize())
	}

	// Test background frame rates
	if desktopAdapter.GetBackgroundFrameRate() != 30 {
		t.Errorf("Desktop GetBackgroundFrameRate() = %d, expected 30", desktopAdapter.GetBackgroundFrameRate())
	}

	if mobileAdapter.GetBackgroundFrameRate() != 5 {
		t.Errorf("Mobile GetBackgroundFrameRate() = %d, expected 5", mobileAdapter.GetBackgroundFrameRate())
	}
}

// TestNilPlatformHandling tests handling of nil platform information
func TestNilPlatformHandling(t *testing.T) {
	adapter := NewPlatformBehaviorAdapter(nil)

	// Should not panic and should provide reasonable defaults
	config := adapter.GetBehaviorConfig()
	if config == nil {
		t.Fatal("GetBehaviorConfig() should not return nil even with nil platform")
	}

	// Should default to desktop behavior for nil platform
	if !config.MovementEnabled {
		t.Error("Nil platform should default to desktop behavior with movement enabled")
	}

	size := adapter.GetOptimalCharacterSize(1920, 128)
	if size != 128 {
		t.Errorf("Nil platform should default to desktop sizing, got %d, expected 128", size)
	}
}

// BenchmarkGetBehaviorConfig benchmarks the behavior configuration retrieval
func BenchmarkGetBehaviorConfig(b *testing.B) {
	adapter := NewPlatformBehaviorAdapter(&platform.PlatformInfo{
		OS:         "android",
		FormFactor: "mobile",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = adapter.GetBehaviorConfig()
	}
}

// BenchmarkGetOptimalCharacterSize benchmarks character size calculation
func BenchmarkGetOptimalCharacterSize(b *testing.B) {
	adapter := NewPlatformBehaviorAdapter(&platform.PlatformInfo{
		OS:         "android",
		FormFactor: "mobile",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = adapter.GetOptimalCharacterSize(400.0, 128)
	}
}
