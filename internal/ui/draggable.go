package ui

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"desktop-companion/internal/character"
)

// DraggableCharacter implements character dragging using Fyne's event system
// Follows the "lazy programmer" principle by leveraging Fyne's built-in drag support
type DraggableCharacter struct {
	widget.BaseWidget
	window    *DesktopWindow
	character *character.Character
	debug     bool

	// Drag state
	dragging   bool
	dragStartX float32
	dragStartY float32
	startPosX  float32
	startPosY  float32
}

// NewDraggableCharacter creates a new draggable character widget
func NewDraggableCharacter(window *DesktopWindow, char *character.Character, debug bool) *DraggableCharacter {
	dc := &DraggableCharacter{
		window:    window,
		character: char,
		debug:     debug,
	}

	dc.ExtendBaseWidget(dc)

	if debug {
		log.Println("Created draggable character wrapper")
	}

	return dc
}

// CreateRenderer creates the Fyne renderer for the draggable character
func (dc *DraggableCharacter) CreateRenderer() fyne.WidgetRenderer {
	return &draggableCharacterRenderer{
		draggable: dc,
		renderer:  dc.window.renderer,
	}
}

// Dragged handles drag events to move the character
// This implements fyne.Draggable interface for built-in drag support
func (dc *DraggableCharacter) Dragged(event *fyne.DragEvent) {
	if !dc.character.IsMovementEnabled() {
		return
	}

	if !dc.dragging {
		// Start dragging
		dc.dragging = true
		dc.dragStartX = event.Position.X
		dc.dragStartY = event.Position.Y
		dc.startPosX, dc.startPosY = dc.character.GetPosition()

		if dc.debug {
			log.Printf("Started dragging at (%.1f, %.1f)", event.Position.X, event.Position.Y)
		}
		return
	}

	// Calculate new position based on drag delta
	deltaX := event.Position.X - dc.dragStartX
	deltaY := event.Position.Y - dc.dragStartY

	newX := dc.startPosX + deltaX
	newY := dc.startPosY + deltaY

	// Update character position
	dc.character.SetPosition(newX, newY)

	// Move the window to follow the character
	// Note: Fyne window positioning may be limited on some platforms
	dc.moveWindow(newX, newY)

	if dc.debug {
		log.Printf("Dragging to (%.1f, %.1f), delta (%.1f, %.1f)", newX, newY, deltaX, deltaY)
	}
}

// DragEnd handles the end of a drag operation
func (dc *DraggableCharacter) DragEnd() {
	if dc.dragging {
		dc.dragging = false

		finalX, finalY := dc.character.GetPosition()
		if dc.debug {
			log.Printf("Drag ended at final position (%.1f, %.1f)", finalX, finalY)
		}
	}
}

// moveWindow attempts to move the window to follow character position
// Uses improved positioning logic with available Fyne capabilities
func (dc *DraggableCharacter) moveWindow(x, y float32) {
	// Store position in character for consistency
	dc.character.SetPosition(x, y)

	// Use the improved SetPosition method from DesktopWindow
	// This provides better positioning support using available Fyne APIs
	if dc.window != nil {
		dc.window.SetPosition(int(x), int(y))
	}

	if dc.debug {
		log.Printf("Character position updated to (%.1f, %.1f) via improved positioning", x, y)
	}
}

// MouseIn handles mouse enter events
func (dc *DraggableCharacter) MouseIn(event *fyne.PointEvent) {
	// Trigger hover interaction when mouse enters character area
	response := dc.character.HandleHover()

	if dc.debug {
		log.Printf("Mouse entered character area at (%.1f, %.1f), hover response: %q", event.Position.X, event.Position.Y, response)
	}

	if response != "" {
		dc.window.showDialog(response)
	}
}

// MouseOut handles mouse exit events
func (dc *DraggableCharacter) MouseOut() {
	if dc.debug {
		log.Println("Mouse left character area")
	}
}

// MouseMoved handles mouse movement over the character
func (dc *DraggableCharacter) MouseMoved(event *fyne.PointEvent) {
	// Optional: Could be used for hover effects
	// For now, we'll keep it simple to avoid performance overhead
}

// Tapped handles tap/click events on the character
func (dc *DraggableCharacter) Tapped(event *fyne.PointEvent) {
	// Delegate to the window's click handler
	dc.window.handleClick()

	if dc.debug {
		log.Printf("Character tapped at (%.1f, %.1f)", event.Position.X, event.Position.Y)
	}
}

// TappedSecondary handles right-click/secondary tap events
func (dc *DraggableCharacter) TappedSecondary(event *fyne.PointEvent) {
	// Delegate to the window's right-click handler for context menu
	dc.window.handleRightClick()

	if dc.debug {
		log.Printf("Character right-clicked at (%.1f, %.1f)", event.Position.X, event.Position.Y)
	}
}

// draggableCharacterRenderer implements fyne.WidgetRenderer for the draggable character
type draggableCharacterRenderer struct {
	draggable *DraggableCharacter
	renderer  *CharacterRenderer
}

// Layout arranges the draggable character within the widget bounds
func (r *draggableCharacterRenderer) Layout(size fyne.Size) {
	// Delegate to the character renderer
	if r.renderer != nil {
		r.renderer.Resize(size)
	}
}

// MinSize returns the minimum size for the draggable character
func (r *draggableCharacterRenderer) MinSize() fyne.Size {
	if r.renderer != nil {
		return fyne.NewSize(float32(r.renderer.GetSize()), float32(r.renderer.GetSize()))
	}
	return fyne.NewSize(64, 64)
}

// Objects returns the canvas objects for rendering
func (r *draggableCharacterRenderer) Objects() []fyne.CanvasObject {
	if r.renderer != nil {
		return []fyne.CanvasObject{r.renderer}
	}
	return []fyne.CanvasObject{}
}

// Refresh redraws the draggable character
func (r *draggableCharacterRenderer) Refresh() {
	if r.renderer != nil {
		r.renderer.Refresh()
	}
}

// Destroy cleans up draggable character resources
func (r *draggableCharacterRenderer) Destroy() {
	// No special cleanup needed
}
