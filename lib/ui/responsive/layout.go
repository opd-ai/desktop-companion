// Package responsive provides adaptive UI layout management for cross-platform compatibility.
// This package implements responsive design patterns that adapt to different screen sizes
// and form factors while maintaining consistent user experience across desktop and mobile.
package responsive

import (
	"math"

	"fyne.io/fyne/v2"

	"github.com/opd-ai/desktop-companion/lib/platform"
)

// Layout provides responsive design calculations for UI components.
// It adapts character sizing, window modes, and layout positioning based on
// platform capabilities and screen dimensions.
type Layout struct {
	platform     *platform.PlatformInfo
	screenWidth  float32
	screenHeight float32
	screenSize   fyne.Size
}

// LayoutMode defines how the application window should be displayed.
type LayoutMode string

const (
	// OverlayMode displays the character as a desktop overlay (traditional desktop pet behavior)
	OverlayMode LayoutMode = "overlay"

	// FullscreenMode displays the character in a fullscreen application (mobile-friendly)
	FullscreenMode LayoutMode = "fullscreen"

	// PictureInPictureMode displays the character in a small floating window (mobile PiP)
	PictureInPictureMode LayoutMode = "pip"
)

// WindowConfig contains platform-specific window configuration.
type WindowConfig struct {
	Mode          LayoutMode
	CharacterSize int
	WindowSize    fyne.Size
	AlwaysOnTop   bool
	Transparent   bool
	Resizable     bool
	ShowControls  bool
	ShowStatusBar bool
}

// NewLayout creates a new responsive layout manager using Fyne's built-in screen detection.
// It automatically detects screen dimensions and platform capabilities to provide
// appropriate layout calculations.
func NewLayout(platform *platform.PlatformInfo, app fyne.App) *Layout {
	// Use Fyne's driver to get screen information - standard library approach
	var screenSize fyne.Size
	if app != nil && len(app.Driver().AllWindows()) > 0 {
		// Get screen size from the first available window
		screenSize = app.Driver().AllWindows()[0].Canvas().Size()
	}

	// Fallback to reasonable defaults if screen detection fails or in test environment
	if screenSize.Width <= 10 || screenSize.Height <= 10 {
		if platform != nil && platform.IsMobile() {
			screenSize = fyne.NewSize(360, 640) // Common mobile resolution
		} else {
			screenSize = fyne.NewSize(1920, 1080) // Common desktop resolution
		}
	}

	return &Layout{
		platform:     platform,
		screenWidth:  screenSize.Width,
		screenHeight: screenSize.Height,
		screenSize:   screenSize,
	}
} // GetCharacterSize calculates the optimal character size based on platform and screen dimensions.
// Mobile platforms get larger touch targets, desktop platforms use fixed sizes.
func (l *Layout) GetCharacterSize(defaultSize int) int {
	if l.platform != nil && l.platform.IsMobile() {
		// Mobile: Character should be 25% of screen width for easy touch interaction
		// This follows mobile UI guidelines for minimum touch target sizes
		mobileSize := int(l.screenWidth * 0.25)

		// Ensure minimum size for readability and maximum size for usability
		const minMobileSize = 100
		const maxMobileSize = 300

		if mobileSize < minMobileSize {
			return minMobileSize
		}
		if mobileSize > maxMobileSize {
			return maxMobileSize
		}

		return mobileSize
	} else {
		// Desktop: Use configured default size, with reasonable bounds
		const minDesktopSize = 64
		const maxDesktopSize = 512

		if defaultSize < minDesktopSize {
			return minDesktopSize
		}
		if defaultSize > maxDesktopSize {
			return maxDesktopSize
		}

		return defaultSize
	}
}

// GetLayoutMode determines the appropriate window display mode for the current platform.
func (l *Layout) GetLayoutMode() LayoutMode {
	if l.platform != nil && l.platform.IsMobile() {
		return FullscreenMode // Mobile apps should take the full screen
	} else {
		return OverlayMode // Desktop pets are overlay windows
	}
}

// GetWindowConfig generates complete window configuration for the current platform.
// This provides all the settings needed to create an appropriately configured window.
func (l *Layout) GetWindowConfig(defaultCharacterSize int) *WindowConfig {
	mode := l.GetLayoutMode()
	characterSize := l.GetCharacterSize(defaultCharacterSize)

	config := &WindowConfig{
		Mode:          mode,
		CharacterSize: characterSize,
	}

	switch mode {
	case OverlayMode:
		// Desktop overlay configuration
		config.WindowSize = fyne.NewSize(float32(characterSize), float32(characterSize))
		config.AlwaysOnTop = true
		config.Transparent = true
		config.Resizable = false
		config.ShowControls = false
		config.ShowStatusBar = false

	case FullscreenMode:
		// Mobile fullscreen configuration
		config.WindowSize = l.screenSize
		config.AlwaysOnTop = false
		config.Transparent = false
		config.Resizable = false
		config.ShowControls = true
		config.ShowStatusBar = true

	case PictureInPictureMode:
		// Mobile picture-in-picture configuration
		pipSize := int(math.Min(float64(characterSize)*1.5, float64(l.screenWidth)*0.3))
		config.WindowSize = fyne.NewSize(float32(pipSize), float32(pipSize))
		config.AlwaysOnTop = true
		config.Transparent = false
		config.Resizable = false
		config.ShowControls = false
		config.ShowStatusBar = false
	}

	return config
}

// GetOptimalPosition calculates the best position for the character window.
// Desktop: Uses bottom-right corner, Mobile: Centers the character.
func (l *Layout) GetOptimalPosition(windowSize fyne.Size) fyne.Position {
	if l.platform != nil && l.platform.IsMobile() {
		// Mobile: Center the character for best visibility
		x := (l.screenWidth - windowSize.Width) / 2
		y := (l.screenHeight - windowSize.Height) / 2
		return fyne.NewPos(x, y)
	} else {
		// Desktop: Bottom-right corner with padding
		const padding float32 = 20
		x := l.screenWidth - windowSize.Width - padding
		y := l.screenHeight - windowSize.Height - padding
		return fyne.NewPos(x, y)
	}
}

// ShouldShowMobileControls determines if mobile control UI should be displayed.
func (l *Layout) ShouldShowMobileControls() bool {
	return l.platform != nil && l.platform.IsMobile() && l.GetLayoutMode() == FullscreenMode
}

// GetTouchTargetSize returns the minimum size for touch-interactive elements.
// This follows platform-specific UI guidelines for accessibility.
func (l *Layout) GetTouchTargetSize() int {
	if l.platform != nil && l.platform.HasTouch() {
		return 44 // iOS Human Interface Guidelines minimum touch target (44x44 points)
	} else {
		return 24 // Desktop minimum for mouse precision
	}
}

// AdaptToScreenRotation updates layout calculations when screen orientation changes.
// This is primarily relevant for mobile devices that support rotation.
func (l *Layout) AdaptToScreenRotation(newSize fyne.Size) {
	l.screenWidth = newSize.Width
	l.screenHeight = newSize.Height
	l.screenSize = newSize
}
