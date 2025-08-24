package ui

import (
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"desktop-companion/internal/character"
)

// DesktopWindow represents the transparent overlay window containing the character
// Uses Fyne for cross-platform window management - avoiding custom windowing code
type DesktopWindow struct {
	window    fyne.Window
	character *character.Character
	renderer  *CharacterRenderer
	dialog    *DialogBubble
	debug     bool
}

// NewDesktopWindow creates a new transparent desktop window
// Uses Fyne's desktop app interface for always-on-top and transparency
func NewDesktopWindow(app fyne.App, char *character.Character, debug bool) *DesktopWindow {
	// Create window with transparency support
	window := app.NewWindow("Desktop Companion")

	// Configure window for desktop overlay
	window.SetFixedSize(true)
	window.Resize(fyne.NewSize(float32(char.GetSize()), float32(char.GetSize())))

	dw := &DesktopWindow{
		window:    window,
		character: char,
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
		log.Printf("Created desktop window: %dx%d", char.GetSize(), char.GetSize())
	}

	return dw
}

// setupContent configures the window's visual content
func (dw *DesktopWindow) setupContent() {
	// Create container with character renderer and dialog overlay
	content := container.NewWithoutLayout(
		dw.renderer,
		dw.dialog,
	)

	dw.window.SetContent(content)
}

// setupInteractions configures mouse interactions with the character
func (dw *DesktopWindow) setupInteractions() {
	// Wrap renderer in a button for click detection
	// This is simpler than implementing custom gesture detection
	clickable := widget.NewButton("", func() {
		dw.handleClick()
	})
	clickable.Resize(fyne.NewSize(float32(dw.character.GetSize()), float32(dw.character.GetSize())))

	// Make button transparent by removing background
	clickable.Importance = widget.LowImportance

	// Add right-click support if available
	if dw.supportsRightClick() {
		dw.setupRightClick(clickable)
	}

	// Add dragging support if character allows movement
	if dw.character.IsMovementEnabled() {
		dw.setupDragging()
	}

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

// handleRightClick processes right-click interactions
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

		// Refresh renderer to show new animation frame
		dw.renderer.Refresh()
	}
}

// supportsRightClick checks if the platform supports right-click detection
func (dw *DesktopWindow) supportsRightClick() bool {
	// Fyne supports right-click on desktop platforms
	// This could be extended to check specific platform capabilities
	return true
}

// setupRightClick configures right-click interaction
func (dw *DesktopWindow) setupRightClick(widget fyne.Widget) {
	// Fyne doesn't have built-in right-click support for widgets
	// This would need platform-specific implementation or custom gesture detection
	// For now, we'll use a simple approach with a context menu

	// TODO: Implement proper right-click detection
	// This is a placeholder for platform-specific right-click handling
}

// setupDragging configures character dragging behavior
func (dw *DesktopWindow) setupDragging() {
	// TODO: Implement window dragging
	// This requires platform-specific window manipulation
	// Fyne provides some drag support, but for desktop overlay windows
	// we might need to handle this at the OS level

	if dw.debug {
		log.Println("Dragging enabled for character (implementation pending)")
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
// Note: Fyne doesn't directly support window positioning on all platforms
func (dw *DesktopWindow) SetPosition(x, y int) {
	// Store position in character for reference
	dw.character.SetPosition(float32(x), float32(y))

	// Note: Window positioning may not be supported on all platforms
	if dw.debug {
		log.Printf("Position set to (%d, %d) - actual window positioning may not be supported", x, y)
	}
}

// GetPosition returns the current window position
// Note: Fyne doesn't directly support window position queries on all platforms
func (dw *DesktopWindow) GetPosition() (int, int) {
	// Return stored character position as fallback
	x, y := dw.character.GetPosition()
	return int(x), int(y)
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
