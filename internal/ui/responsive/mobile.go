// Package responsive provides mobile-specific window management for cross-platform compatibility.
// This file implements mobile window modes including fullscreen and picture-in-picture support.
package responsive

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"desktop-companion/internal/platform"
)

// MobileWindowManager handles window lifecycle and mode transitions for mobile platforms.
// It provides picture-in-picture support, fullscreen mode, and background/foreground handling.
type MobileWindowManager struct {
	platform     *platform.PlatformInfo
	layout       *Layout
	currentMode  LayoutMode
	window       fyne.Window
	content      fyne.CanvasObject
	controlBar   *MobileControlBar
	isBackground bool
}

// MobileControlBar provides touch-friendly control buttons that replace keyboard shortcuts.
// This follows mobile UI patterns with bottom sheet navigation and large touch targets.
type MobileControlBar struct {
	platform      *platform.PlatformInfo
	container     *fyne.Container
	statsButton   *widget.Button
	chatButton    *widget.Button
	networkButton *widget.Button
	menuButton    *widget.Button
	visible       bool
}

// NewMobileWindowManager creates a window manager optimized for mobile platforms.
// It automatically configures appropriate window modes and control interfaces.
func NewMobileWindowManager(platform *platform.PlatformInfo, layout *Layout) *MobileWindowManager {
	currentMode := OverlayMode // Default to overlay mode
	if layout != nil {
		currentMode = layout.GetLayoutMode()
	}

	return &MobileWindowManager{
		platform:     platform,
		layout:       layout,
		currentMode:  currentMode,
		isBackground: false,
	}
}

// ConfigureWindow applies mobile-specific configuration to a Fyne window.
// This sets up fullscreen mode, removes desktop decorations, and optimizes for touch.
func (mwm *MobileWindowManager) ConfigureWindow(window fyne.Window) error {
	mwm.window = window

	if mwm.platform == nil || !mwm.platform.IsMobile() {
		return nil // No mobile configuration needed for desktop
	}

	config := mwm.layout.GetWindowConfig(128) // Use default character size

	switch config.Mode {
	case FullscreenMode:
		mwm.configureFullscreen(config)
	case PictureInPictureMode:
		mwm.configurePictureInPicture(config)
	default:
		// Use overlay mode as fallback
		mwm.configureOverlay(config)
	}

	return nil
}

// configureFullscreen sets up mobile fullscreen mode.
func (mwm *MobileWindowManager) configureFullscreen(config *WindowConfig) {
	mwm.window.Resize(config.WindowSize)
	mwm.window.SetFixedSize(true)
	mwm.currentMode = FullscreenMode

	// Mobile fullscreen apps should not be always on top
	// and should show standard mobile UI elements
	if desktop, ok := mwm.window.(interface{ SetAlwaysOnTop(bool) }); ok {
		desktop.SetAlwaysOnTop(false)
	}
}

// configurePictureInPicture sets up mobile picture-in-picture mode.
// This is for background operation while other apps are in foreground.
func (mwm *MobileWindowManager) configurePictureInPicture(config *WindowConfig) {
	mwm.window.Resize(config.WindowSize)
	mwm.window.SetFixedSize(true)
	mwm.currentMode = PictureInPictureMode

	// PiP mode should be on top but not take focus
	if desktop, ok := mwm.window.(interface{ SetAlwaysOnTop(bool) }); ok {
		desktop.SetAlwaysOnTop(true)
	}
}

// configureOverlay sets up desktop-style overlay mode (fallback for mobile).
func (mwm *MobileWindowManager) configureOverlay(config *WindowConfig) {
	mwm.window.Resize(config.WindowSize)
	mwm.currentMode = OverlayMode
}

// SetContent sets the main content and adds mobile controls if appropriate.
func (mwm *MobileWindowManager) SetContent(content fyne.CanvasObject) {
	mwm.content = content

	if mwm.window == nil {
		return // No window configured yet
	}

	// Use fallback if content is nil
	if content == nil {
		content = container.NewWithoutLayout() // Empty content fallback
	}

	if mwm.platform != nil && mwm.platform.IsMobile() && mwm.currentMode == FullscreenMode {
		// Mobile fullscreen: Add control bar
		mwm.createMobileControls()
		if mwm.controlBar != nil && mwm.controlBar.container != nil {
			fullContent := container.NewVBox(
				content,
				mwm.controlBar.container,
			)
			mwm.window.SetContent(fullContent)
		} else {
			mwm.window.SetContent(content)
		}
	} else {
		// Desktop or PiP mode: Use content directly
		mwm.window.SetContent(content)
	}
}

// createMobileControls creates the mobile control bar with touch-friendly buttons.
func (mwm *MobileWindowManager) createMobileControls() {
	if mwm.controlBar != nil {
		return // Already created
	}

	mwm.controlBar = NewMobileControlBar(mwm.platform)
}

// EnterPictureInPictureMode transitions to picture-in-picture mode.
// This is called when the app goes to background on mobile platforms.
func (mwm *MobileWindowManager) EnterPictureInPictureMode() error {
	if mwm.platform == nil || !mwm.platform.IsMobile() {
		return nil // Desktop doesn't support PiP mode
	}

	if mwm.currentMode == PictureInPictureMode {
		return nil // Already in PiP mode
	}

	// Transition to smaller PiP window
	config := mwm.layout.GetWindowConfig(128)
	config.Mode = PictureInPictureMode

	mwm.configurePictureInPicture(config)
	mwm.isBackground = true

	// Remove mobile controls in PiP mode to save space
	if mwm.content != nil {
		mwm.window.SetContent(mwm.content)
	}

	return nil
}

// ExitPictureInPictureMode transitions back to fullscreen mode.
// This is called when the app returns to foreground.
func (mwm *MobileWindowManager) ExitPictureInPictureMode() error {
	if mwm.platform == nil || !mwm.platform.IsMobile() {
		return nil
	}

	if mwm.currentMode != PictureInPictureMode {
		return nil // Not in PiP mode
	}

	// Transition back to fullscreen
	config := mwm.layout.GetWindowConfig(128)
	mwm.configureFullscreen(config)
	mwm.isBackground = false

	// Restore mobile controls
	mwm.SetContent(mwm.content)

	return nil
}

// HandleBackgroundTransition manages app lifecycle transitions.
// This optimizes performance when the app is not actively visible.
func (mwm *MobileWindowManager) HandleBackgroundTransition(isBackground bool) {
	mwm.isBackground = isBackground

	if mwm.platform != nil && mwm.platform.IsMobile() {
		if isBackground {
			// Enter PiP mode when backgrounded
			mwm.EnterPictureInPictureMode()
		} else {
			// Return to fullscreen when foregrounded
			mwm.ExitPictureInPictureMode()
		}
	}
}

// IsInPictureInPictureMode returns true if currently in PiP mode.
func (mwm *MobileWindowManager) IsInPictureInPictureMode() bool {
	return mwm.currentMode == PictureInPictureMode
}

// GetCurrentMode returns the current window layout mode.
func (mwm *MobileWindowManager) GetCurrentMode() LayoutMode {
	return mwm.currentMode
}

// NewMobileControlBar creates a mobile control bar with platform-appropriate buttons.
// Buttons are sized according to mobile touch target guidelines.
func NewMobileControlBar(platform *platform.PlatformInfo) *MobileControlBar {
	mcb := &MobileControlBar{
		platform: platform,
		visible:  true,
	}

	// Create touch-friendly buttons with proper sizing
	touchTargetSize := float32(44) // iOS Human Interface Guidelines

	mcb.statsButton = widget.NewButton("📊 Stats", func() {
		// Stats button callback - will be connected to main app logic
	})
	mcb.statsButton.Resize(fyne.NewSize(touchTargetSize*2, touchTargetSize))

	mcb.chatButton = widget.NewButton("💬 Chat", func() {
		// Chat button callback - will be connected to main app logic
	})
	mcb.chatButton.Resize(fyne.NewSize(touchTargetSize*2, touchTargetSize))

	mcb.networkButton = widget.NewButton("🌐 Network", func() {
		// Network button callback - will be connected to main app logic
	})
	mcb.networkButton.Resize(fyne.NewSize(touchTargetSize*2, touchTargetSize))

	mcb.menuButton = widget.NewButton("⚙️ Menu", func() {
		// Menu button callback - will be connected to main app logic
	})
	mcb.menuButton.Resize(fyne.NewSize(touchTargetSize*2, touchTargetSize))

	// Create horizontal layout for mobile control bar
	mcb.container = container.NewVBox(
		container.NewHBox(
			mcb.statsButton,
			mcb.chatButton,
			mcb.networkButton,
			mcb.menuButton,
		),
	)

	return mcb
}

// SetStatsCallback sets the callback for the stats button.
func (mcb *MobileControlBar) SetStatsCallback(callback func()) {
	if mcb.statsButton != nil {
		mcb.statsButton.OnTapped = callback
	}
}

// SetChatCallback sets the callback for the chat button.
func (mcb *MobileControlBar) SetChatCallback(callback func()) {
	if mcb.chatButton != nil {
		mcb.chatButton.OnTapped = callback
	}
}

// SetNetworkCallback sets the callback for the network button.
func (mcb *MobileControlBar) SetNetworkCallback(callback func()) {
	if mcb.networkButton != nil {
		mcb.networkButton.OnTapped = callback
	}
}

// SetMenuCallback sets the callback for the menu button.
func (mcb *MobileControlBar) SetMenuCallback(callback func()) {
	if mcb.menuButton != nil {
		mcb.menuButton.OnTapped = callback
	}
}

// Show makes the control bar visible.
func (mcb *MobileControlBar) Show() {
	mcb.visible = true
	if mcb.container != nil {
		mcb.container.Show()
	}
}

// Hide makes the control bar invisible.
func (mcb *MobileControlBar) Hide() {
	mcb.visible = false
	if mcb.container != nil {
		mcb.container.Hide()
	}
}

// IsVisible returns true if the control bar is currently visible.
func (mcb *MobileControlBar) IsVisible() bool {
	return mcb.visible
}

// GetContainer returns the container widget for embedding in the main UI.
func (mcb *MobileControlBar) GetContainer() *fyne.Container {
	return mcb.container
}
