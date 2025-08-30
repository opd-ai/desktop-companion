package character

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"desktop-companion/internal/platform"
)

// TestNewWithPlatform tests the platform-aware character constructor
func TestNewWithPlatform(t *testing.T) {
	// Create test animation file
	idleGifPath := createTestGIF(t, "idle.gif", 2, nil)
	defer os.RemoveAll(filepath.Dir(idleGifPath)) // Clean up temp directory

	// Create test character card with just the filename
	card := &CharacterCard{
		Name:        "Test Character",
		Description: "A test character for platform behavior",
		Animations: map[string]string{
			"idle": "idle.gif", // Just the filename
		},
		Behavior: Behavior{
			IdleTimeout:     30,
			MovementEnabled: true,
			DefaultSize:     128,
		},
	}

	tests := []struct {
		name           string
		platformInfo   *platform.PlatformInfo
		expectedSize   int
		expectedMoving bool
		expectError    bool
	}{
		{
			name:           "desktop platform",
			platformInfo:   &platform.PlatformInfo{OS: "windows", FormFactor: "desktop"},
			expectedSize:   128,  // Should use default size
			expectedMoving: true, // Desktop allows movement
			expectError:    false,
		},
		{
			name:           "mobile platform",
			platformInfo:   &platform.PlatformInfo{OS: "android", FormFactor: "mobile"},
			expectedSize:   100,   // 25% of 400px mobile screen
			expectedMoving: false, // Mobile disables movement
			expectError:    false,
		},
		{
			name:           "nil platform (defaults to desktop)",
			platformInfo:   nil,
			expectedSize:   128,  // Should use default size
			expectedMoving: true, // Desktop behavior
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use the directory containing the GIF as the basePath
			basePath := filepath.Dir(idleGifPath)
			char, err := NewWithPlatform(card, basePath, tt.platformInfo)

			if (err != nil) != tt.expectError {
				t.Errorf("NewWithPlatform() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if err != nil {
				return // Skip further checks if we expected an error
			}

			if char == nil {
				t.Fatal("NewWithPlatform() returned nil character")
			}

			// Check platform adapter was created
			if char.platformAdapter == nil {
				t.Error("Character should have platform adapter")
			}

			// Check size adaptation
			if char.size != tt.expectedSize {
				t.Errorf("Character size = %d, expected %d", char.size, tt.expectedSize)
			}

			// Check movement setting
			if char.movementEnabled != tt.expectedMoving {
				t.Errorf("Movement enabled = %v, expected %v", char.movementEnabled, tt.expectedMoving)
			}
		})
	}
}

// TestCharacterPlatformMethods tests platform-aware methods on Character
func TestCharacterPlatformMethods(t *testing.T) {
	// Create test animation file
	idleGifPath := createTestGIF(t, "idle.gif", 2, nil)
	defer os.RemoveAll(filepath.Dir(idleGifPath)) // Clean up temp directory

	// Create test character for desktop
	desktopCard := &CharacterCard{
		Name:        "Desktop Character",
		Description: "A test character for desktop",
		Animations:  map[string]string{"idle": "idle.gif"},
		Behavior: Behavior{
			IdleTimeout:     30,
			MovementEnabled: true,
			DefaultSize:     128,
		},
	}

	basePath := filepath.Dir(idleGifPath)
	desktopChar, err := NewWithPlatform(desktopCard, basePath, &platform.PlatformInfo{
		OS:         "windows",
		FormFactor: "desktop",
	})
	if err != nil {
		t.Fatalf("Failed to create desktop character: %v", err)
	}

	// Create test character for mobile
	mobileChar, err := NewWithPlatform(desktopCard, basePath, &platform.PlatformInfo{
		OS:         "android",
		FormFactor: "mobile",
	})
	if err != nil {
		t.Fatalf("Failed to create mobile character: %v", err)
	}

	t.Run("GetPlatformBehaviorConfig", func(t *testing.T) {
		desktopConfig := desktopChar.GetPlatformBehaviorConfig()
		mobileConfig := mobileChar.GetPlatformBehaviorConfig()

		if desktopConfig == nil || mobileConfig == nil {
			t.Fatal("GetPlatformBehaviorConfig() returned nil")
		}

		// Desktop should have different settings than mobile
		if desktopConfig.MovementEnabled == mobileConfig.MovementEnabled {
			t.Error("Desktop and mobile should have different movement settings")
		}

		if desktopConfig.AnimationFrameRate == mobileConfig.AnimationFrameRate {
			t.Error("Desktop and mobile should have different frame rates")
		}
	})

	t.Run("GetInteractionDelay", func(t *testing.T) {
		desktopDelay := desktopChar.GetInteractionDelay("click")
		mobileDelay := mobileChar.GetInteractionDelay("tap")

		// Mobile should have longer delays to prevent accidental interactions
		if mobileDelay <= desktopDelay {
			t.Errorf("Mobile delay (%v) should be longer than desktop delay (%v)", mobileDelay, desktopDelay)
		}

		// Test different event types
		longPressDelay := mobileChar.GetInteractionDelay("longpress")
		if longPressDelay <= mobileDelay {
			t.Errorf("Long press delay (%v) should be longer than tap delay (%v)", longPressDelay, mobileDelay)
		}
	})

	t.Run("ShouldEnableFeature", func(t *testing.T) {
		// Movement should be enabled on desktop but not mobile
		if !desktopChar.ShouldEnableFeature("movement") {
			t.Error("Desktop should enable movement")
		}

		if mobileChar.ShouldEnableFeature("movement") {
			t.Error("Mobile should disable movement")
		}

		// Battery optimization should be disabled on desktop but enabled on mobile
		if desktopChar.ShouldEnableFeature("battery_optimization") {
			t.Error("Desktop should disable battery optimization")
		}

		if !mobileChar.ShouldEnableFeature("battery_optimization") {
			t.Error("Mobile should enable battery optimization")
		}

		// Unknown features should default to enabled
		if !desktopChar.ShouldEnableFeature("unknown_feature") {
			t.Error("Unknown features should default to enabled")
		}
	})

	t.Run("GetOptimalSize", func(t *testing.T) {
		// Test different screen sizes
		smallScreen := float32(320)
		largeScreen := float32(1920)

		desktopSmall := desktopChar.GetOptimalSize(smallScreen)
		desktopLarge := desktopChar.GetOptimalSize(largeScreen)

		// Desktop should use same size regardless of screen size (uses default)
		if desktopSmall != desktopLarge {
			t.Errorf("Desktop should use consistent size, got %d vs %d", desktopSmall, desktopLarge)
		}

		mobileSmall := mobileChar.GetOptimalSize(smallScreen)
		mobileLarge := mobileChar.GetOptimalSize(largeScreen)

		// Mobile should adapt to screen size
		if mobileSmall >= mobileLarge {
			t.Errorf("Mobile should scale with screen size, got %d vs %d", mobileSmall, mobileLarge)
		}

		// Mobile should respect minimum size
		if mobileSmall < 96 {
			t.Errorf("Mobile size should respect minimum of 96, got %d", mobileSmall)
		}
	})

	t.Run("GetAnimationFrameRate", func(t *testing.T) {
		desktopFPS := desktopChar.GetAnimationFrameRate()
		mobileFPS := mobileChar.GetAnimationFrameRate()

		// Desktop should have higher frame rate
		if desktopFPS <= mobileFPS {
			t.Errorf("Desktop FPS (%d) should be higher than mobile FPS (%d)", desktopFPS, mobileFPS)
		}

		// Check expected values
		if desktopFPS != 60 {
			t.Errorf("Desktop FPS should be 60, got %d", desktopFPS)
		}

		if mobileFPS != 30 {
			t.Errorf("Mobile FPS should be 30, got %d", mobileFPS)
		}
	})

	t.Run("GetBackgroundFrameRate", func(t *testing.T) {
		desktopBgFPS := desktopChar.GetBackgroundFrameRate()
		mobileBgFPS := mobileChar.GetBackgroundFrameRate()

		// Mobile should have much lower background frame rate
		if mobileBgFPS >= desktopBgFPS {
			t.Errorf("Mobile background FPS (%d) should be lower than desktop (%d)", mobileBgFPS, desktopBgFPS)
		}

		// Check expected values
		if desktopBgFPS != 30 {
			t.Errorf("Desktop background FPS should be 30, got %d", desktopBgFPS)
		}

		if mobileBgFPS != 5 {
			t.Errorf("Mobile background FPS should be 5, got %d", mobileBgFPS)
		}
	})

	t.Run("UpdatePlatformSize", func(t *testing.T) {
		initialSize := mobileChar.GetSize()

		// Update to larger screen
		mobileChar.UpdatePlatformSize(800)
		newSize := mobileChar.GetSize()

		// Size should increase for larger screen on mobile
		if newSize <= initialSize {
			t.Errorf("Size should increase when screen gets larger, got %d vs %d", newSize, initialSize)
		}

		// Desktop character should not change size (uses fixed default)
		desktopInitialSize := desktopChar.GetSize()
		desktopChar.UpdatePlatformSize(800)
		desktopNewSize := desktopChar.GetSize()

		if desktopNewSize != desktopInitialSize {
			t.Errorf("Desktop size should not change, got %d vs %d", desktopNewSize, desktopInitialSize)
		}
	})
}

// TestCharacterWithNilPlatformAdapter tests behavior when platform adapter is nil
func TestCharacterWithNilPlatformAdapter(t *testing.T) {
	// Create test animation file
	idleGifPath := createTestGIF(t, "idle.gif", 2, nil)
	defer os.RemoveAll(filepath.Dir(idleGifPath)) // Clean up temp directory

	// Create character with old constructor (no platform adapter)
	card := &CharacterCard{
		Name:        "Legacy Character",
		Description: "A character created without platform awareness",
		Animations:  map[string]string{"idle": "idle.gif"},
		Behavior: Behavior{
			IdleTimeout:     30,
			MovementEnabled: true,
			DefaultSize:     128,
		},
	}

	basePath := filepath.Dir(idleGifPath)
	char, err := New(card, basePath)
	if err != nil {
		t.Fatalf("Failed to create character: %v", err)
	}

	// Manually set platform adapter to nil to test fallback behavior
	char.platformAdapter = nil

	t.Run("GetPlatformBehaviorConfig with nil adapter", func(t *testing.T) {
		config := char.GetPlatformBehaviorConfig()
		if config == nil {
			t.Fatal("Should return fallback config when adapter is nil")
		}

		// Should default to desktop behavior
		if !config.MovementEnabled {
			t.Error("Fallback config should enable movement (desktop behavior)")
		}
	})

	t.Run("GetInteractionDelay with nil adapter", func(t *testing.T) {
		delay := char.GetInteractionDelay("click")
		if delay != 1*time.Second {
			t.Errorf("Should return default delay of 1s, got %v", delay)
		}
	})

	t.Run("ShouldEnableFeature with nil adapter", func(t *testing.T) {
		if !char.ShouldEnableFeature("any_feature") {
			t.Error("Should default to enabled when adapter is nil")
		}
	})

	t.Run("GetOptimalSize with nil adapter", func(t *testing.T) {
		size := char.GetOptimalSize(1920)
		if size != char.GetSize() {
			t.Errorf("Should return current size when adapter is nil, got %d vs %d", size, char.GetSize())
		}
	})

	t.Run("GetAnimationFrameRate with nil adapter", func(t *testing.T) {
		fps := char.GetAnimationFrameRate()
		if fps != 60 {
			t.Errorf("Should return default 60 FPS when adapter is nil, got %d", fps)
		}
	})
}

// TestCharacterBackwardCompatibility tests that existing code still works
func TestCharacterBackwardCompatibility(t *testing.T) {
	card := &CharacterCard{
		Name:        "Legacy Character",
		Description: "Testing backward compatibility",
		Animations:  map[string]string{"idle": "testdata/idle.gif"},
		Behavior: Behavior{
			IdleTimeout:     30,
			MovementEnabled: true,
			DefaultSize:     128,
		},
	}

	// Old constructor should still work
	char, err := New(card, ".")
	if err != nil {
		t.Fatalf("Legacy constructor failed: %v", err)
	}

	// Should have platform adapter (defaults to desktop behavior)
	if char.platformAdapter == nil {
		t.Error("Legacy constructor should create platform adapter")
	}

	// Should use original behavior when no platform info provided
	if char.GetSize() != 128 {
		t.Errorf("Legacy character should use default size, got %d", char.GetSize())
	}

	if !char.movementEnabled {
		t.Error("Legacy character should preserve movement setting")
	}
}

// BenchmarkPlatformAwareMethods benchmarks the performance of platform-aware methods
func BenchmarkPlatformAwareMethods(b *testing.B) {
	card := &CharacterCard{
		Name:        "Benchmark Character",
		Description: "For performance testing",
		Animations:  map[string]string{"idle": "testdata/idle.gif"},
		Behavior: Behavior{
			IdleTimeout:     30,
			MovementEnabled: true,
			DefaultSize:     128,
		},
	}

	char, err := NewWithPlatform(card, ".", &platform.PlatformInfo{
		OS:         "android",
		FormFactor: "mobile",
	})
	if err != nil {
		b.Fatalf("Failed to create character: %v", err)
	}

	b.Run("GetPlatformBehaviorConfig", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = char.GetPlatformBehaviorConfig()
		}
	})

	b.Run("GetInteractionDelay", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = char.GetInteractionDelay("tap")
		}
	})

	b.Run("ShouldEnableFeature", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = char.ShouldEnableFeature("movement")
		}
	})

	b.Run("GetOptimalSize", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = char.GetOptimalSize(400.0)
		}
	})
}
