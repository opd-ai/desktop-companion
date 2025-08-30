// Package main demonstrates touch gesture integration with existing DDS components.
// This shows how to update existing interaction handlers to support mobile gestures.
package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	"desktop-companion/internal/character"
	"desktop-companion/internal/platform"
	"desktop-companion/internal/ui"
)

// TouchEnabledWindow demonstrates how to modify the existing DesktopWindow
// to support touch gestures while maintaining desktop compatibility.
type TouchEnabledWindow struct {
	window    fyne.Window
	character *character.Character
	clickable *ui.PlatformAwareClickableWidget
	platform  *platform.PlatformInfo
	debug     bool
}

// NewTouchEnabledWindow creates a window with touch gesture support.
// This is an example of how to modify the existing window creation code.
func NewTouchEnabledWindow(app fyne.App, char *character.Character, debug bool) *TouchEnabledWindow {
	window := app.NewWindow("Desktop Companion - Touch Enabled")
	platform := platform.GetPlatformInfo()

	tw := &TouchEnabledWindow{
		window:    window,
		character: char,
		platform:  platform,
		debug:     debug,
	}

	tw.setupWindow()
	tw.setupInteractions()

	return tw
}

// setupWindow configures basic window properties
func (tw *TouchEnabledWindow) setupWindow() {
	// Configure window size based on platform
	size := tw.getOptimalWindowSize()
	tw.window.Resize(size)
	tw.window.SetFixedSize(true)

	if tw.debug {
		log.Printf("Touch-enabled window created for platform: %s", tw.platform.OS)
	}
}

// setupInteractions demonstrates the key change: replacing ClickableWidget
// with PlatformAwareClickableWidget for automatic gesture support
func (tw *TouchEnabledWindow) setupInteractions() {
	// Create platform-aware clickable widget (replaces NewClickableWidget)
	tw.clickable = ui.NewPlatformAwareClickableWidgetWithDoubleTap(
		func() { tw.handleTap() },       // Single tap -> left click
		func() { tw.handleLongPress() }, // Long press -> right click
		func() { tw.handleDoubleTap() }, // Double tap -> double click
	)

	// Set up drag support for character movement
	tw.clickable.SetDragHandlers(
		func() { tw.handleDragStart() },
		func(event *fyne.DragEvent) { tw.handleDrag(event) },
		func() { tw.handleDragEnd() },
	)

	// Size the clickable area to match character
	size := float32(tw.character.GetSize())
	tw.clickable.SetSize(fyne.NewSize(size, size))

	// Create window content
	content := container.NewWithoutLayout(tw.clickable)
	tw.window.SetContent(content)

	if tw.debug && tw.clickable.IsGestureTranslationActive() {
		log.Println("Touch gesture translation enabled")
	}
}

// getOptimalWindowSize returns platform-appropriate window size
func (tw *TouchEnabledWindow) getOptimalWindowSize() fyne.Size {
	baseSize := float32(tw.character.GetSize())

	if tw.platform.IsMobile() {
		// Mobile devices need larger touch targets
		return fyne.NewSize(baseSize*1.5, baseSize*1.5)
	}

	// Desktop uses standard size
	return fyne.NewSize(baseSize, baseSize)
}

// Event handlers - these remain unchanged from the original implementation

func (tw *TouchEnabledWindow) handleTap() {
	if tw.debug {
		log.Println("Character tapped (single tap/left click)")
	}

	// Existing click handling logic
	response := tw.character.HandleClick()
	if response != "" {
		tw.showDialog(response)
	}
}

func (tw *TouchEnabledWindow) handleLongPress() {
	if tw.debug {
		log.Println("Character long pressed (long press/right click)")
	}

	// Existing right-click handling logic
	tw.showContextMenu()
}

func (tw *TouchEnabledWindow) handleDoubleTap() {
	if tw.debug {
		log.Println("Character double tapped (double tap/double click)")
	}

	// Handle double tap as a special interaction
	// For example, could trigger play interaction
	response := tw.character.HandleGameInteraction("play")
	if response != "" {
		tw.showDialog(response)
	}
}

func (tw *TouchEnabledWindow) handleDragStart() {
	if tw.debug {
		log.Println("Character drag started")
	}

	// Existing drag start logic
}

func (tw *TouchEnabledWindow) handleDrag(event *fyne.DragEvent) {
	if tw.debug {
		log.Printf("Character dragged: delta (%.1f, %.1f)", event.Dragged.DX, event.Dragged.DY)
	}

	// Existing drag handling logic
	if tw.character.IsMovementEnabled() {
		currentX, currentY := tw.character.GetPosition()
		newX := currentX + event.Dragged.DX
		newY := currentY + event.Dragged.DY
		tw.character.SetPosition(newX, newY)
	}
}

func (tw *TouchEnabledWindow) handleDragEnd() {
	if tw.debug {
		log.Println("Character drag ended")
	}

	// Existing drag end logic
}

func (tw *TouchEnabledWindow) showDialog(text string) {
	if tw.debug {
		log.Printf("Showing dialog: %q", text)
	}
	// Dialog display logic would go here
}

func (tw *TouchEnabledWindow) showContextMenu() {
	if tw.debug {
		log.Println("Showing context menu")
	}
	// Context menu logic would go here
}

// Show displays the window
func (tw *TouchEnabledWindow) Show() {
	tw.window.Show()
}

// Migration demonstrates how to update existing DesktopWindow code
func MigrationExample() {
	// BEFORE (existing code):
	// clickable := NewClickableWidget(
	//     func() { dw.handleClick() },
	//     func() { dw.handleRightClick() },
	// )

	// AFTER (with touch support):
	clickable := ui.NewPlatformAwareClickableWidget(
		func() { /* dw.handleClick() */ },
		func() { /* dw.handleRightClick() */ },
	)

	// Add double tap support if needed
	clickableWithDouble := ui.NewPlatformAwareClickableWidgetWithDoubleTap(
		func() { /* single tap */ },
		func() { /* long press */ },
		func() { /* double tap */ },
	)

	// Add drag support for mobile
	clickableWithDouble.SetDragHandlers(
		func() { /* drag start */ },
		func(event *fyne.DragEvent) { /* drag move */ },
		func() { /* drag end */ },
	)

	// Check if running on touch platform
	if clickable.IsGestureTranslationActive() {
		log.Println("Touch gestures enabled")
	}

	// Rest of the code remains unchanged
	_ = clickable
	_ = clickableWithDouble
}

// main demonstrates the touch integration system
func main() {
	log.Println("Touch gesture integration example")
	log.Println("This example shows how to integrate touch gestures with existing DDS components")

	// Show migration example
	MigrationExample()

	log.Println("Integration complete - see function implementations above")
}
