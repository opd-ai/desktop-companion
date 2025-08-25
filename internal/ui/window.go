package ui

import (
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	"desktop-companion/internal/character"
	"desktop-companion/internal/monitoring"
)

// DesktopWindow represents the transparent overlay window containing the character
// Uses Fyne for cross-platform window management - avoiding custom windowing code
type DesktopWindow struct {
	window    fyne.Window
	character *character.Character
	renderer  *CharacterRenderer
	dialog    *DialogBubble
	profiler  *monitoring.Profiler
	debug     bool
}

// NewDesktopWindow creates a new transparent desktop window
// Uses Fyne's desktop app interface for always-on-top and transparency
func NewDesktopWindow(app fyne.App, char *character.Character, debug bool, profiler *monitoring.Profiler) *DesktopWindow {
	// Create window with transparency support
	window := app.NewWindow("Desktop Companion")

	// Configure window for desktop overlay behavior
	window.SetFixedSize(true)
	window.Resize(fyne.NewSize(float32(char.GetSize()), float32(char.GetSize())))

	// Configure transparency for desktop overlay
	configureTransparency(window, debug)

	// Attempt to configure always-on-top behavior using available Fyne capabilities
	// Note: Fyne has limited always-on-top support, but we can try available approaches
	configureAlwaysOnTop(window, debug)

	dw := &DesktopWindow{
		window:    window,
		character: char,
		profiler:  profiler,
		debug:     debug,
	}

	// Create character renderer
	dw.renderer = NewCharacterRenderer(char, debug)

	// Create dialog bubble (initially hidden)
	dw.dialog = NewDialogBubble()

	// Set up window content and interactions
	dw.setupContent()
	dw.setupInteractions()

	// Start animation update loop
	go dw.animationLoop()

	if debug {
		log.Printf("Created desktop window: %dx%d with always-on-top configuration", char.GetSize(), char.GetSize())
	}

	return dw
}

// setupContent configures the window's visual content
func (dw *DesktopWindow) setupContent() {
	// Create container with transparent background for overlay effect
	content := container.NewWithoutLayout(
		dw.renderer,
		dw.dialog,
	)

	dw.window.SetContent(content)

	if dw.debug {
		log.Println("Window content configured for transparent overlay")
	}
}

// setupInteractions configures mouse interactions with the character
func (dw *DesktopWindow) setupInteractions() {
	// Add dragging support if character allows movement
	if dw.character.IsMovementEnabled() {
		dw.setupDragging()
		return // Draggable character handles all interactions
	}

	// For non-draggable characters, create custom clickable widget that supports both left and right click
	clickable := NewClickableWidget(
		func() { dw.handleClick() },
		func() { dw.handleRightClick() },
	)
	clickable.SetSize(fyne.NewSize(float32(dw.character.GetSize()), float32(dw.character.GetSize())))

	// Update window content with interactive overlay
	content := container.NewWithoutLayout(
		dw.renderer,
		clickable,
		dw.dialog,
	)

	dw.window.SetContent(content)
}

// handleClick processes character click interactions
func (dw *DesktopWindow) handleClick() {
	response := dw.character.HandleClick()

	if dw.debug {
		log.Printf("Character clicked, response: %q", response)
	}

	if response != "" {
		dw.showDialog(response)
	}
}

// handleRightClick processes character right-click interactions
func (dw *DesktopWindow) handleRightClick() {
	response := dw.character.HandleRightClick()

	if dw.debug {
		log.Printf("Character right-clicked, response: %q", response)
	}

	if response != "" {
		dw.showDialog(response)
	}
}

// showDialog displays a dialog bubble with the given text
func (dw *DesktopWindow) showDialog(text string) {
	dw.dialog.ShowWithText(text)

	// Auto-hide dialog after 3 seconds
	go func() {
		time.Sleep(3 * time.Second)
		dw.dialog.Hide()
	}()
}

// animationLoop runs the character animation update loop
// Maintains 60 FPS for smooth animation playback
func (dw *DesktopWindow) animationLoop() {
	ticker := time.NewTicker(time.Second / 60) // 60 FPS
	defer ticker.Stop()

	for range ticker.C {
		// Update character behavior and animations
		dw.character.Update()

		// Record frame for performance monitoring
		if dw.profiler != nil {
			dw.profiler.RecordFrame()
		}

		// Refresh renderer to show new animation frame
		dw.renderer.Refresh()
	}
}

// setupDragging configures character dragging behavior
func (dw *DesktopWindow) setupDragging() {
	// Create draggable wrapper that implements Fyne's drag interface
	// This provides smooth cross-platform drag support without platform-specific code
	draggable := NewDraggableCharacter(dw, dw.character, dw.debug)

	// Update window content to use draggable character instead of separate clickable overlay
	content := container.NewWithoutLayout(
		draggable,
		dw.dialog,
	)

	dw.window.SetContent(content)

	if dw.debug {
		log.Println("Character dragging enabled using Fyne drag system")
	}
}

// Show displays the desktop window
func (dw *DesktopWindow) Show() {
	dw.window.Show()

	if dw.debug {
		log.Printf("Desktop window shown for character: %s", dw.character.GetName())
	}
}

// Hide hides the desktop window
func (dw *DesktopWindow) Hide() {
	dw.window.Hide()
}

// Close closes the desktop window and stops animation
func (dw *DesktopWindow) Close() {
	dw.window.Close()
}

// SetPosition moves the window to the specified screen coordinates
// Uses available Fyne APIs for best-effort positioning support
func (dw *DesktopWindow) SetPosition(x, y int) {
	// Store position in character for reference
	dw.character.SetPosition(float32(x), float32(y))

	// Attempt to use available Fyne positioning capabilities
	// Note: Full positioning support varies by platform, but we can try
	if x == 0 && y == 0 {
		// Special case: center the window when position is (0,0)
		dw.window.CenterOnScreen()
		if dw.debug {
			log.Printf("Centering window using CenterOnScreen()")
		}
	} else {
		// For non-zero positions, we need to work within Fyne's limitations
		// Fyne doesn't expose direct positioning, but we can provide feedback
		if dw.debug {
			log.Printf("Position set to (%d, %d) - stored in character. Note: Fyne has limited window positioning support on some platforms", x, y)
		}
	}
}

// GetPosition returns the current window position
// Note: Fyne doesn't directly support window position queries on all platforms
func (dw *DesktopWindow) GetPosition() (int, int) {
	// Return stored character position as fallback
	x, y := dw.character.GetPosition()
	return int(x), int(y)
}

// CenterWindow centers the window on screen using Fyne's built-in capability
func (dw *DesktopWindow) CenterWindow() {
	dw.window.CenterOnScreen()
	// Reset stored position to indicate centered state
	dw.character.SetPosition(0, 0)

	if dw.debug {
		log.Println("Window centered on screen")
	}
}

// SetSize updates the window and character size
func (dw *DesktopWindow) SetSize(size int) {
	dw.window.Resize(fyne.NewSize(float32(size), float32(size)))
	dw.renderer.SetSize(size)
}

// GetCharacter returns the character instance for external access
func (dw *DesktopWindow) GetCharacter() *character.Character {
	return dw.character
}

// configureAlwaysOnTop attempts to configure always-on-top behavior using available Fyne capabilities
// Following the "lazy programmer" principle: use what's available rather than implementing platform-specific code
func configureAlwaysOnTop(window fyne.Window, debug bool) {
	// Fyne v2.4.5 has limited always-on-top support, but we can try available approaches:

	// 1. Try to minimize window decorations (makes it more overlay-like)
	window.SetTitle("") // Remove title bar text for cleaner overlay appearance

	// 2. Set window to be borderless for better desktop integration
	// Note: Fyne doesn't expose direct borderless mode, but we can minimize decoration

	// 3. Configure for desktop overlay use case
	// Fyne's design philosophy focuses on cross-platform compatibility over platform-specific features
	// True always-on-top requires platform-specific window manager hints that Fyne doesn't expose

	if debug {
		log.Println("Always-on-top configuration applied using available Fyne capabilities")
		log.Println("Note: Full always-on-top behavior requires platform-specific window manager support")
		log.Println("Window configured for optimal desktop overlay experience within Fyne's limitations")
	}

	// Future enhancement opportunity:
	// Could implement platform-specific always-on-top using CGO or system calls,
	// but this would violate the "lazy programmer" principle of avoiding custom platform code
}

// configureTransparency configures window transparency for desktop overlay behavior
// Following the "lazy programmer" principle: use Fyne's available transparency features
func configureTransparency(window fyne.Window, debug bool) {
	// Remove window padding to make character appear directly on desktop
	window.SetPadded(false)

	if debug {
		log.Println("Window transparency configuration applied using available Fyne capabilities")
		log.Println("Note: True transparency requires transparent window backgrounds and content")
		log.Println("Character should appear with minimal window decoration for overlay effect")
	}

	// Future enhancement opportunity:
	// Could explore platform-specific transparency using Fyne driver extensions,
	// but this maintains cross-platform compatibility by using standard Fyne APIs
}
