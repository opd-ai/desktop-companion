package ui

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"

	"github.com/opd-ai/desktop-companion/lib/character"
)

// CharacterRenderer renders the animated character using Fyne's canvas
// Leverages Fyne's built-in image rendering instead of custom graphics code
type CharacterRenderer struct {
	widget.BaseWidget
	character *character.Character
	image     *canvas.Image
	debug     bool
	size      int
}

// NewCharacterRenderer creates a new character renderer widget
func NewCharacterRenderer(char *character.Character, debug bool) *CharacterRenderer {
	r := &CharacterRenderer{
		character: char,
		debug:     debug,
		size:      char.GetSize(),
	}

	// Create canvas image for character display
	r.image = canvas.NewImageFromImage(nil)
	r.image.FillMode = canvas.ImageFillContain
	r.image.ScaleMode = canvas.ImageScaleSmooth

	// Set initial size
	r.image.Resize(fyne.NewSize(float32(r.size), float32(r.size)))

	// Load initial frame
	r.updateFrame()

	r.ExtendBaseWidget(r)

	if debug {
		log.Printf("Character renderer created with size: %d", r.size)
	}

	return r
}

// CreateRenderer creates the Fyne renderer for this widget
func (r *CharacterRenderer) CreateRenderer() fyne.WidgetRenderer {
	return &characterWidgetRenderer{
		image: r.image,
	}
}

// updateFrame updates the displayed animation frame
func (r *CharacterRenderer) updateFrame() {
	frame := r.character.GetCurrentFrame()
	if frame != nil {
		r.image.Image = frame
		r.image.Refresh()

		if r.debug {
			state := r.character.GetCurrentState()
			log.Printf("Updated frame for state: %s", state)
		}
	}
}

// Refresh updates the character display with the current animation frame
func (r *CharacterRenderer) Refresh() {
	r.updateFrame()
	r.BaseWidget.Refresh()
}

// SetSize updates the character display size
func (r *CharacterRenderer) SetSize(size int) {
	r.size = size
	r.image.Resize(fyne.NewSize(float32(size), float32(size)))
	r.Refresh()
}

// GetSize returns the current character display size
func (r *CharacterRenderer) GetSize() int {
	return r.size
}

// characterWidgetRenderer implements fyne.WidgetRenderer for the character
type characterWidgetRenderer struct {
	image *canvas.Image
}

// Layout arranges the character image within the widget bounds
func (r *characterWidgetRenderer) Layout(size fyne.Size) {
	r.image.Resize(size)
	r.image.Move(fyne.NewPos(0, 0))
}

// MinSize returns the minimum size required for the character
func (r *characterWidgetRenderer) MinSize() fyne.Size {
	return fyne.NewSize(64, 64) // Minimum 64x64 pixels
}

// Objects returns the list of canvas objects to render
func (r *characterWidgetRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.image}
}

// Refresh redraws the character renderer
func (r *characterWidgetRenderer) Refresh() {
	r.image.Refresh()
}

// Destroy cleans up renderer resources
func (r *characterWidgetRenderer) Destroy() {
	// No special cleanup needed for image canvas object
}
