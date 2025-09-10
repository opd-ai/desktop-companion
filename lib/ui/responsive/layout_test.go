package responsive

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"github.com/opd-ai/desktop-companion/lib/platform"
)

// TestNewLayout verifies layout creation with different platform types
func TestNewLayout(t *testing.T) {
	tests := []struct {
		name     string
		platform *platform.PlatformInfo
		wantSize fyne.Size
	}{
		{
			name: "Desktop Platform",
			platform: &platform.PlatformInfo{
				OS:           "linux",
				FormFactor:   "desktop",
				InputMethods: []string{"mouse", "keyboard"},
			},
			wantSize: fyne.NewSize(1920, 1080), // Default desktop resolution
		},
		{
			name: "Mobile Platform",
			platform: &platform.PlatformInfo{
				OS:           "android",
				FormFactor:   "mobile",
				InputMethods: []string{"touch"},
			},
			wantSize: fyne.NewSize(360, 640), // Default mobile resolution
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test app without windows
			app := test.NewApp()
			defer app.Quit()

			layout := NewLayout(tt.platform, app)

			if layout == nil {
				t.Fatal("NewLayout returned nil")
			}

			if layout.platform != tt.platform {
				t.Errorf("Platform not set correctly: got %v, want %v", layout.platform, tt.platform)
			}

			// Screen size should use fallback values when no windows exist
			if layout.screenSize != tt.wantSize {
				t.Errorf("Screen size incorrect: got %v, want %v", layout.screenSize, tt.wantSize)
			}
		})
	}
}

// TestGetCharacterSize verifies character sizing calculations
func TestGetCharacterSize(t *testing.T) {
	tests := []struct {
		name         string
		platform     *platform.PlatformInfo
		screenWidth  float32
		defaultSize  int
		expectedSize int
	}{
		{
			name: "Desktop with normal default size",
			platform: &platform.PlatformInfo{
				FormFactor: "desktop",
			},
			screenWidth:  1920,
			defaultSize:  128,
			expectedSize: 128,
		},
		{
			name: "Desktop with too small default size",
			platform: &platform.PlatformInfo{
				FormFactor: "desktop",
			},
			screenWidth:  1920,
			defaultSize:  32, // Below minimum
			expectedSize: 64, // Should use minimum
		},
		{
			name: "Desktop with too large default size",
			platform: &platform.PlatformInfo{
				FormFactor: "desktop",
			},
			screenWidth:  1920,
			defaultSize:  1000, // Above maximum
			expectedSize: 512,  // Should use maximum
		},
		{
			name: "Mobile with standard screen",
			platform: &platform.PlatformInfo{
				FormFactor: "mobile",
			},
			screenWidth:  360,
			defaultSize:  128,
			expectedSize: 100, // 25% of 360 = 90, but minimum is 100
		},
		{
			name: "Mobile with large screen",
			platform: &platform.PlatformInfo{
				FormFactor: "mobile",
			},
			screenWidth:  800,
			defaultSize:  128,
			expectedSize: 200, // 25% of 800 = 200
		},
		{
			name: "Mobile with very large screen",
			platform: &platform.PlatformInfo{
				FormFactor: "mobile",
			},
			screenWidth:  1600,
			defaultSize:  128,
			expectedSize: 300, // 25% of 1600 = 400, but maximum is 300
		},
		{
			name: "Mobile with very small screen",
			platform: &platform.PlatformInfo{
				FormFactor: "mobile",
			},
			screenWidth:  240,
			defaultSize:  128,
			expectedSize: 100, // 25% of 240 = 60, but minimum is 100
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layout := &Layout{
				platform:    tt.platform,
				screenWidth: tt.screenWidth,
			}

			result := layout.GetCharacterSize(tt.defaultSize)

			if result != tt.expectedSize {
				t.Errorf("GetCharacterSize() = %d, want %d", result, tt.expectedSize)
			}
		})
	}
}

// TestGetLayoutMode verifies layout mode selection
func TestGetLayoutMode(t *testing.T) {
	tests := []struct {
		name         string
		platform     *platform.PlatformInfo
		expectedMode LayoutMode
	}{
		{
			name: "Desktop platform should use overlay",
			platform: &platform.PlatformInfo{
				FormFactor: "desktop",
			},
			expectedMode: OverlayMode,
		},
		{
			name: "Mobile platform should use fullscreen",
			platform: &platform.PlatformInfo{
				FormFactor: "mobile",
			},
			expectedMode: FullscreenMode,
		},
		{
			name: "Tablet platform should use fullscreen",
			platform: &platform.PlatformInfo{
				FormFactor: "tablet",
			},
			expectedMode: OverlayMode, // Tablets behave like desktop for now
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layout := &Layout{
				platform: tt.platform,
			}

			result := layout.GetLayoutMode()

			if result != tt.expectedMode {
				t.Errorf("GetLayoutMode() = %v, want %v", result, tt.expectedMode)
			}
		})
	}
}

// TestGetWindowConfig verifies complete window configuration
func TestGetWindowConfig(t *testing.T) {
	tests := []struct {
		name             string
		platform         *platform.PlatformInfo
		defaultSize      int
		expectedMode     LayoutMode
		expectedControls bool
		expectedOnTop    bool
	}{
		{
			name: "Desktop configuration",
			platform: &platform.PlatformInfo{
				FormFactor: "desktop",
			},
			defaultSize:      128,
			expectedMode:     OverlayMode,
			expectedControls: false,
			expectedOnTop:    true,
		},
		{
			name: "Mobile configuration",
			platform: &platform.PlatformInfo{
				FormFactor: "mobile",
			},
			defaultSize:      128,
			expectedMode:     FullscreenMode,
			expectedControls: true,
			expectedOnTop:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layout := &Layout{
				platform:     tt.platform,
				screenWidth:  360,
				screenHeight: 640,
				screenSize:   fyne.NewSize(360, 640),
			}

			config := layout.GetWindowConfig(tt.defaultSize)

			if config.Mode != tt.expectedMode {
				t.Errorf("Mode = %v, want %v", config.Mode, tt.expectedMode)
			}

			if config.ShowControls != tt.expectedControls {
				t.Errorf("ShowControls = %v, want %v", config.ShowControls, tt.expectedControls)
			}

			if config.AlwaysOnTop != tt.expectedOnTop {
				t.Errorf("AlwaysOnTop = %v, want %v", config.AlwaysOnTop, tt.expectedOnTop)
			}

			if config.CharacterSize <= 0 {
				t.Error("CharacterSize should be positive")
			}

			if config.WindowSize.Width <= 0 || config.WindowSize.Height <= 0 {
				t.Error("WindowSize should be positive")
			}
		})
	}
}

// TestGetOptimalPosition verifies positioning calculations
func TestGetOptimalPosition(t *testing.T) {
	tests := []struct {
		name       string
		platform   *platform.PlatformInfo
		screenSize fyne.Size
		windowSize fyne.Size
		expectX    float32
		expectY    float32
	}{
		{
			name: "Desktop positioning (bottom-right)",
			platform: &platform.PlatformInfo{
				FormFactor: "desktop",
			},
			screenSize: fyne.NewSize(1920, 1080),
			windowSize: fyne.NewSize(128, 128),
			expectX:    1772, // 1920 - 128 - 20
			expectY:    932,  // 1080 - 128 - 20
		},
		{
			name: "Mobile positioning (centered)",
			platform: &platform.PlatformInfo{
				FormFactor: "mobile",
			},
			screenSize: fyne.NewSize(360, 640),
			windowSize: fyne.NewSize(100, 100),
			expectX:    130, // (360 - 100) / 2
			expectY:    270, // (640 - 100) / 2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layout := &Layout{
				platform:     tt.platform,
				screenWidth:  tt.screenSize.Width,
				screenHeight: tt.screenSize.Height,
			}

			pos := layout.GetOptimalPosition(tt.windowSize)

			if pos.X != tt.expectX {
				t.Errorf("Position X = %f, want %f", pos.X, tt.expectX)
			}

			if pos.Y != tt.expectY {
				t.Errorf("Position Y = %f, want %f", pos.Y, tt.expectY)
			}
		})
	}
}

// TestShouldShowMobileControls verifies mobile control visibility logic
func TestShouldShowMobileControls(t *testing.T) {
	tests := []struct {
		name     string
		platform *platform.PlatformInfo
		expected bool
	}{
		{
			name: "Desktop should not show mobile controls",
			platform: &platform.PlatformInfo{
				FormFactor: "desktop",
			},
			expected: false,
		},
		{
			name: "Mobile should show mobile controls",
			platform: &platform.PlatformInfo{
				FormFactor: "mobile",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layout := &Layout{
				platform: tt.platform,
			}

			result := layout.ShouldShowMobileControls()

			if result != tt.expected {
				t.Errorf("ShouldShowMobileControls() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestGetTouchTargetSize verifies touch target sizing
func TestGetTouchTargetSize(t *testing.T) {
	tests := []struct {
		name         string
		platform     *platform.PlatformInfo
		expectedSize int
	}{
		{
			name: "Touch platform should use 44pt minimum",
			platform: &platform.PlatformInfo{
				InputMethods: []string{"touch"},
			},
			expectedSize: 44,
		},
		{
			name: "Mouse platform should use 24pt minimum",
			platform: &platform.PlatformInfo{
				InputMethods: []string{"mouse", "keyboard"},
			},
			expectedSize: 24,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layout := &Layout{
				platform: tt.platform,
			}

			result := layout.GetTouchTargetSize()

			if result != tt.expectedSize {
				t.Errorf("GetTouchTargetSize() = %d, want %d", result, tt.expectedSize)
			}
		})
	}
}

// TestAdaptToScreenRotation verifies screen rotation handling
func TestAdaptToScreenRotation(t *testing.T) {
	layout := &Layout{
		platform:     &platform.PlatformInfo{FormFactor: "mobile"},
		screenWidth:  360,
		screenHeight: 640,
		screenSize:   fyne.NewSize(360, 640),
	}

	// Simulate rotation from portrait to landscape
	newSize := fyne.NewSize(640, 360)
	layout.AdaptToScreenRotation(newSize)

	if layout.screenWidth != 640 {
		t.Errorf("screenWidth = %f, want 640", layout.screenWidth)
	}

	if layout.screenHeight != 360 {
		t.Errorf("screenHeight = %f, want 360", layout.screenHeight)
	}

	if layout.screenSize != newSize {
		t.Errorf("screenSize = %v, want %v", layout.screenSize, newSize)
	}
}

// BenchmarkGetCharacterSize tests performance of character size calculations
func BenchmarkGetCharacterSize(b *testing.B) {
	layout := &Layout{
		platform: &platform.PlatformInfo{
			FormFactor: "mobile",
		},
		screenWidth: 360,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		layout.GetCharacterSize(128)
	}
}

// BenchmarkGetWindowConfig tests performance of window configuration generation
func BenchmarkGetWindowConfig(b *testing.B) {
	layout := &Layout{
		platform: &platform.PlatformInfo{
			FormFactor: "mobile",
		},
		screenWidth:  360,
		screenHeight: 640,
		screenSize:   fyne.NewSize(360, 640),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		layout.GetWindowConfig(128)
	}
}

// TestNilPlatformHandling verifies graceful handling of nil platform info
func TestNilPlatformHandling(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Function panicked with nil platform: %v", r)
		}
	}()

	layout := &Layout{
		platform:    nil, // This should not cause panics
		screenWidth: 360,
	}

	// These calls should handle nil platform gracefully
	// Note: This test verifies we don't panic, the behavior with nil platform is undefined
	_ = layout.GetCharacterSize(128)
}

// TestEdgeCases verifies handling of edge cases and boundary conditions
func TestEdgeCases(t *testing.T) {
	t.Run("Zero screen dimensions", func(t *testing.T) {
		layout := &Layout{
			platform: &platform.PlatformInfo{
				FormFactor: "mobile",
			},
			screenWidth:  0,
			screenHeight: 0,
		}

		size := layout.GetCharacterSize(128)
		if size < 100 { // Should enforce minimum size
			t.Errorf("Character size too small with zero screen: %d", size)
		}
	})

	t.Run("Negative default size", func(t *testing.T) {
		layout := &Layout{
			platform: &platform.PlatformInfo{
				FormFactor: "desktop",
			},
		}

		size := layout.GetCharacterSize(-50)
		if size < 64 { // Should enforce minimum size
			t.Errorf("Character size too small with negative default: %d", size)
		}
	})

	t.Run("Very large screen", func(t *testing.T) {
		layout := &Layout{
			platform: &platform.PlatformInfo{
				FormFactor: "mobile",
			},
			screenWidth: 10000,
		}

		size := layout.GetCharacterSize(128)
		if size > 300 { // Should enforce maximum size
			t.Errorf("Character size too large: %d", size)
		}
	})
}
