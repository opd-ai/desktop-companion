package responsive

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"github.com/opd-ai/desktop-companion/internal/platform"
)

// TestNewMobileWindowManager verifies mobile window manager creation
func TestNewMobileWindowManager(t *testing.T) {
	platform := &platform.PlatformInfo{
		FormFactor: "mobile",
	}

	layout := &Layout{
		platform:     platform,
		screenWidth:  360,
		screenHeight: 640,
	}

	mwm := NewMobileWindowManager(platform, layout)

	if mwm == nil {
		t.Fatal("NewMobileWindowManager returned nil")
	}

	if mwm.platform != platform {
		t.Error("Platform not set correctly")
	}

	if mwm.layout != layout {
		t.Error("Layout not set correctly")
	}

	if mwm.currentMode != FullscreenMode {
		t.Errorf("Expected FullscreenMode for mobile, got %v", mwm.currentMode)
	}

	if mwm.isBackground {
		t.Error("Should not start in background mode")
	}
}

// TestConfigureWindow verifies window configuration for different platforms
func TestConfigureWindow(t *testing.T) {
	tests := []struct {
		name            string
		platform        *platform.PlatformInfo
		expectedMode    LayoutMode
		shouldConfigure bool
	}{
		{
			name: "Mobile platform should configure window",
			platform: &platform.PlatformInfo{
				FormFactor: "mobile",
			},
			expectedMode:    FullscreenMode,
			shouldConfigure: true,
		},
		{
			name: "Desktop platform should not configure window",
			platform: &platform.PlatformInfo{
				FormFactor: "desktop",
			},
			expectedMode:    OverlayMode,
			shouldConfigure: false,
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

			mwm := NewMobileWindowManager(tt.platform, layout)

			app := test.NewApp()
			defer app.Quit()
			window := app.NewWindow("Test")

			err := mwm.ConfigureWindow(window)

			if err != nil {
				t.Errorf("ConfigureWindow returned error: %v", err)
			}

			if mwm.window != window {
				t.Error("Window not set correctly")
			}

			if mwm.currentMode != tt.expectedMode {
				t.Errorf("Expected mode %v, got %v", tt.expectedMode, mwm.currentMode)
			}
		})
	}
}

// TestSetContent verifies content setting with mobile controls
func TestSetContent(t *testing.T) {
	tests := []struct {
		name           string
		platform       *platform.PlatformInfo
		mode           LayoutMode
		expectControls bool
	}{
		{
			name: "Mobile fullscreen should add controls",
			platform: &platform.PlatformInfo{
				FormFactor: "mobile",
			},
			mode:           FullscreenMode,
			expectControls: true,
		},
		{
			name: "Desktop should not add controls",
			platform: &platform.PlatformInfo{
				FormFactor: "desktop",
			},
			mode:           OverlayMode,
			expectControls: false,
		},
		{
			name: "Mobile PiP should not add controls",
			platform: &platform.PlatformInfo{
				FormFactor: "mobile",
			},
			mode:           PictureInPictureMode,
			expectControls: false,
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

			mwm := NewMobileWindowManager(tt.platform, layout)
			mwm.currentMode = tt.mode // Set mode directly for testing

			app := test.NewApp()
			defer app.Quit()
			window := app.NewWindow("Test")
			mwm.window = window

			content := widget.NewLabel("Test Content")
			mwm.SetContent(content)

			if mwm.content != content {
				t.Error("Content not set correctly")
			}

			if tt.expectControls {
				if mwm.controlBar == nil {
					t.Error("Expected control bar to be created")
				}
			} else {
				if mwm.controlBar != nil {
					t.Error("Did not expect control bar to be created")
				}
			}
		})
	}
}

// TestPictureInPictureMode verifies PiP mode transitions
func TestPictureInPictureMode(t *testing.T) {
	platform := &platform.PlatformInfo{
		FormFactor: "mobile",
	}

	layout := &Layout{
		platform:     platform,
		screenWidth:  360,
		screenHeight: 640,
		screenSize:   fyne.NewSize(360, 640),
	}

	mwm := NewMobileWindowManager(platform, layout)

	app := test.NewApp()
	defer app.Quit()
	window := app.NewWindow("Test")
	mwm.window = window

	// Start in fullscreen mode
	mwm.currentMode = FullscreenMode

	// Set some initial content
	initialContent := widget.NewLabel("Test Content")
	mwm.SetContent(initialContent) // Enter PiP mode
	err := mwm.EnterPictureInPictureMode()
	if err != nil {
		t.Errorf("EnterPictureInPictureMode returned error: %v", err)
	}

	if mwm.currentMode != PictureInPictureMode {
		t.Error("Should be in PictureInPictureMode")
	}

	if !mwm.isBackground {
		t.Error("Should be in background state")
	}

	// Exit PiP mode
	err = mwm.ExitPictureInPictureMode()
	if err != nil {
		t.Errorf("ExitPictureInPictureMode returned error: %v", err)
	}

	if mwm.currentMode != FullscreenMode {
		t.Error("Should return to FullscreenMode")
	}

	if mwm.isBackground {
		t.Error("Should not be in background state")
	}
}

// TestPictureInPictureModeDesktop verifies PiP mode behavior on desktop
func TestPictureInPictureModeDesktop(t *testing.T) {
	platform := &platform.PlatformInfo{
		FormFactor: "desktop",
	}

	layout := &Layout{
		platform: platform,
	}

	mwm := NewMobileWindowManager(platform, layout)

	// Desktop should not enter PiP mode
	err := mwm.EnterPictureInPictureMode()
	if err != nil {
		t.Errorf("EnterPictureInPictureMode should not error on desktop: %v", err)
	}

	// Mode should remain unchanged
	if mwm.currentMode == PictureInPictureMode {
		t.Error("Desktop should not enter PictureInPictureMode")
	}
}

// TestHandleBackgroundTransition verifies background/foreground handling
func TestHandleBackgroundTransition(t *testing.T) {
	platform := &platform.PlatformInfo{
		FormFactor: "mobile",
	}

	layout := &Layout{
		platform:     platform,
		screenWidth:  360,
		screenHeight: 640,
		screenSize:   fyne.NewSize(360, 640),
	}

	mwm := NewMobileWindowManager(platform, layout)

	app := test.NewApp()
	defer app.Quit()
	window := app.NewWindow("Test")
	mwm.window = window
	mwm.currentMode = FullscreenMode

	// Go to background
	mwm.HandleBackgroundTransition(true)

	if !mwm.isBackground {
		t.Error("Should be in background state")
	}

	if !mwm.IsInPictureInPictureMode() {
		t.Error("Should be in PiP mode when backgrounded")
	}

	// Return to foreground
	mwm.HandleBackgroundTransition(false)

	if mwm.isBackground {
		t.Error("Should not be in background state")
	}

	if mwm.IsInPictureInPictureMode() {
		t.Error("Should not be in PiP mode when foregrounded")
	}
}

// TestNewMobileControlBar verifies mobile control bar creation
func TestNewMobileControlBar(t *testing.T) {
	platform := &platform.PlatformInfo{
		FormFactor: "mobile",
	}

	mcb := NewMobileControlBar(platform)

	if mcb == nil {
		t.Fatal("NewMobileControlBar returned nil")
	}

	if mcb.platform != platform {
		t.Error("Platform not set correctly")
	}

	if !mcb.visible {
		t.Error("Control bar should be visible by default")
	}

	if mcb.statsButton == nil {
		t.Error("Stats button should be created")
	}

	if mcb.chatButton == nil {
		t.Error("Chat button should be created")
	}

	if mcb.networkButton == nil {
		t.Error("Network button should be created")
	}

	if mcb.menuButton == nil {
		t.Error("Menu button should be created")
	}

	if mcb.container == nil {
		t.Error("Container should be created")
	}
}

// TestMobileControlBarCallbacks verifies callback setting and invocation
func TestMobileControlBarCallbacks(t *testing.T) {
	platform := &platform.PlatformInfo{
		FormFactor: "mobile",
	}

	mcb := NewMobileControlBar(platform)

	// Test stats callback
	statsCallbackCalled := false
	mcb.SetStatsCallback(func() {
		statsCallbackCalled = true
	})

	if mcb.statsButton.OnTapped != nil {
		mcb.statsButton.OnTapped()
	}

	if !statsCallbackCalled {
		t.Error("Stats callback was not called")
	}

	// Test chat callback
	chatCallbackCalled := false
	mcb.SetChatCallback(func() {
		chatCallbackCalled = true
	})

	if mcb.chatButton.OnTapped != nil {
		mcb.chatButton.OnTapped()
	}

	if !chatCallbackCalled {
		t.Error("Chat callback was not called")
	}

	// Test network callback
	networkCallbackCalled := false
	mcb.SetNetworkCallback(func() {
		networkCallbackCalled = true
	})

	if mcb.networkButton.OnTapped != nil {
		mcb.networkButton.OnTapped()
	}

	if !networkCallbackCalled {
		t.Error("Network callback was not called")
	}

	// Test menu callback
	menuCallbackCalled := false
	mcb.SetMenuCallback(func() {
		menuCallbackCalled = true
	})

	if mcb.menuButton.OnTapped != nil {
		mcb.menuButton.OnTapped()
	}

	if !menuCallbackCalled {
		t.Error("Menu callback was not called")
	}
}

// TestMobileControlBarVisibility verifies show/hide functionality
func TestMobileControlBarVisibility(t *testing.T) {
	platform := &platform.PlatformInfo{
		FormFactor: "mobile",
	}

	mcb := NewMobileControlBar(platform)

	// Should be visible by default
	if !mcb.IsVisible() {
		t.Error("Control bar should be visible by default")
	}

	// Hide the control bar
	mcb.Hide()

	if mcb.IsVisible() {
		t.Error("Control bar should be hidden after Hide()")
	}

	// Show the control bar
	mcb.Show()

	if !mcb.IsVisible() {
		t.Error("Control bar should be visible after Show()")
	}
}

// TestGetContainer verifies container retrieval
func TestGetContainer(t *testing.T) {
	platform := &platform.PlatformInfo{
		FormFactor: "mobile",
	}

	mcb := NewMobileControlBar(platform)
	container := mcb.GetContainer()

	if container == nil {
		t.Error("GetContainer should return a container")
	}

	if container != mcb.container {
		t.Error("GetContainer should return the internal container")
	}
}

// TestModeTransitions verifies proper mode state tracking
func TestModeTransitions(t *testing.T) {
	platform := &platform.PlatformInfo{
		FormFactor: "mobile",
	}

	layout := &Layout{
		platform:     platform,
		screenWidth:  360,
		screenHeight: 640,
		screenSize:   fyne.NewSize(360, 640),
	}

	mwm := NewMobileWindowManager(platform, layout)

	app := test.NewApp()
	defer app.Quit()
	window := app.NewWindow("Test")
	mwm.window = window

	// Start in fullscreen
	mwm.currentMode = FullscreenMode

	if mwm.GetCurrentMode() != FullscreenMode {
		t.Error("Should start in FullscreenMode")
	}

	// Enter PiP
	mwm.EnterPictureInPictureMode()

	if mwm.GetCurrentMode() != PictureInPictureMode {
		t.Error("Should be in PictureInPictureMode")
	}

	// Exit PiP
	mwm.ExitPictureInPictureMode()

	if mwm.GetCurrentMode() != FullscreenMode {
		t.Error("Should return to FullscreenMode")
	}
}

// BenchmarkMobileWindowManager tests performance of mobile window operations
func BenchmarkMobileWindowManager(b *testing.B) {
	platform := &platform.PlatformInfo{
		FormFactor: "mobile",
	}

	layout := &Layout{
		platform:     platform,
		screenWidth:  360,
		screenHeight: 640,
		screenSize:   fyne.NewSize(360, 640),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mwm := NewMobileWindowManager(platform, layout)

		app := test.NewApp()
		window := app.NewWindow("Test")
		mwm.ConfigureWindow(window)
		app.Quit()
	}
}

// BenchmarkControlBarCreation tests performance of control bar creation
func BenchmarkControlBarCreation(b *testing.B) {
	platform := &platform.PlatformInfo{
		FormFactor: "mobile",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewMobileControlBar(platform)
	}
}

// TestNilHandling verifies graceful handling of nil inputs
func TestNilHandling(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Function panicked with nil input: %v", r)
		}
	}()

	// Test with nil platform
	layout := &Layout{
		platform: nil,
	}

	mwm := NewMobileWindowManager(nil, layout)
	if mwm == nil {
		t.Error("Should handle nil platform gracefully")
	}

	// Test configuration with nil window (should not panic)
	err := mwm.ConfigureWindow(nil)
	if err != nil {
		t.Errorf("Should handle nil window gracefully: %v", err)
	}
}

// TestMobileEdgeCases verifies handling of edge cases in mobile functionality
func TestMobileEdgeCases(t *testing.T) {
	t.Run("Already in PiP mode", func(t *testing.T) {
		platform := &platform.PlatformInfo{
			FormFactor: "mobile",
		}

		layout := &Layout{platform: platform}
		mwm := NewMobileWindowManager(platform, layout)
		mwm.currentMode = PictureInPictureMode

		// Should not error when already in PiP mode
		err := mwm.EnterPictureInPictureMode()
		if err != nil {
			t.Errorf("Should handle already in PiP mode: %v", err)
		}
	})

	t.Run("Not in PiP mode when exiting", func(t *testing.T) {
		platform := &platform.PlatformInfo{
			FormFactor: "mobile",
		}

		layout := &Layout{platform: platform}
		mwm := NewMobileWindowManager(platform, layout)
		mwm.currentMode = FullscreenMode

		// Should not error when not in PiP mode
		err := mwm.ExitPictureInPictureMode()
		if err != nil {
			t.Errorf("Should handle not in PiP mode: %v", err)
		}
	})
}

// TestCallbacksWithNilButtons verifies callback setting with nil buttons
func TestCallbacksWithNilButtons(t *testing.T) {
	mcb := &MobileControlBar{
		platform: &platform.PlatformInfo{FormFactor: "mobile"},
		// Buttons are nil
	}

	// These should not panic with nil buttons
	mcb.SetStatsCallback(func() {})
	mcb.SetChatCallback(func() {})
	mcb.SetNetworkCallback(func() {})
	mcb.SetMenuCallback(func() {})

	// Visibility operations should also be safe
	mcb.Show()
	mcb.Hide()

	visible := mcb.IsVisible()
	_ = visible // Use the variable to avoid unused warning
}
