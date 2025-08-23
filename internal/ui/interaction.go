package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// DialogBubble displays speech bubbles for character interactions
// Uses Fyne's text and shape components for simple bubble rendering
type DialogBubble struct {
	widget.BaseWidget
	text       *widget.RichText
	background *canvas.Rectangle
	container  *container.Container
	visible    bool
}

// NewDialogBubble creates a new dialog bubble widget
func NewDialogBubble() *DialogBubble {
	bubble := &DialogBubble{}

	// Create background rectangle with rounded corners effect
	bubble.background = canvas.NewRectangle(color.RGBA{R: 255, G: 255, B: 255, A: 230})
	bubble.background.StrokeColor = color.RGBA{R: 100, G: 100, B: 100, A: 255}
	bubble.background.StrokeWidth = 1

	// Create text widget for dialog content
	bubble.text = widget.NewRichText()
	bubble.text.Wrapping = fyne.TextWrapWord
	
	// Set up text styling for readability
	bubble.text.Segments = []widget.RichTextSegment{
		&widget.TextSegment{
			Text: "",
			Style: widget.RichTextStyle{
				ColorName: "foreground",
				SizeName:  "default",
			},
		},
	}

	// Create container with background and text
	bubble.container = container.NewBorder(nil, nil, nil, nil, bubble.background, bubble.text)
	
	// Initially hidden
	bubble.Hide()

	bubble.ExtendBaseWidget(bubble)
	return bubble
}

// CreateRenderer creates the Fyne renderer for the dialog bubble
func (b *DialogBubble) CreateRenderer() fyne.WidgetRenderer {
	return &dialogBubbleRenderer{
		bubble:    b,
		container: b.container,
	}
}

// Show displays the dialog bubble with the specified text
func (b *DialogBubble) Show(text string) {
	// Update text content
	b.text.Segments = []widget.RichTextSegment{
		&widget.TextSegment{
			Text: text,
			Style: widget.RichTextStyle{
				ColorName: "foreground",
				SizeName:  "default",
			},
		},
	}
	b.text.Refresh()

	// Calculate bubble size based on text content
	b.updateSize(text)

	// Make visible
	b.visible = true
	b.container.Show()
	b.Refresh()
}

// Hide hides the dialog bubble
func (b *DialogBubble) Hide() {
	b.visible = false
	b.container.Hide()
	b.Refresh()
}

// IsVisible returns whether the bubble is currently visible
func (b *DialogBubble) IsVisible() bool {
	return b.visible
}

// updateSize calculates appropriate bubble size for the text content
func (b *DialogBubble) updateSize(text string) {
	// Simple size calculation based on text length
	// This could be enhanced with proper text measurement
	textLen := len(text)
	
	// Base size calculations
	minWidth := float32(100)
	minHeight := float32(40)
	
	// Calculate width based on text length (rough estimate)
	width := minWidth + float32(textLen)*2
	if width > 300 { // Max width
		width = 300
	}
	
	// Calculate height based on estimated line wrapping
	lines := 1 + textLen/30 // Rough estimate of lines needed
	height := minHeight + float32(lines-1)*20
	if height > 150 { // Max height
		height = 150
	}

	// Position bubble above character (offset by character size + margin)
	bubbleX := float32(10) // Small offset from character
	bubbleY := float32(-height - 10) // Above character with margin

	// Update container size and position
	b.container.Resize(fyne.NewSize(width, height))
	b.container.Move(fyne.NewPos(bubbleX, bubbleY))
	
	// Update background to match container
	b.background.Resize(fyne.NewSize(width, height))
	
	// Update text area with padding
	textPadding := float32(8)
	b.text.Resize(fyne.NewSize(width-textPadding*2, height-textPadding*2))
	b.text.Move(fyne.NewPos(textPadding, textPadding))
}

// SetBackgroundColor updates the bubble background color
func (b *DialogBubble) SetBackgroundColor(c color.Color) {
	b.background.FillColor = c
	b.background.Refresh()
}

// SetTextColor updates the bubble text color
func (b *DialogBubble) SetTextColor(colorName string) {
	if len(b.text.Segments) > 0 {
		if segment, ok := b.text.Segments[0].(*widget.TextSegment); ok {
			segment.Style.ColorName = colorName
			b.text.Refresh()
		}
	}
}

// dialogBubbleRenderer implements fyne.WidgetRenderer for dialog bubbles
type dialogBubbleRenderer struct {
	bubble    *DialogBubble
	container *container.Container
}

// Layout arranges the dialog bubble components
func (r *dialogBubbleRenderer) Layout(size fyne.Size) {
	if r.bubble.visible {
		r.container.Resize(r.container.Size())
		r.container.Move(r.container.Position())
	}
}

// MinSize returns the minimum size for the dialog bubble
func (r *dialogBubbleRenderer) MinSize() fyne.Size {
	if r.bubble.visible {
		return fyne.NewSize(100, 40)
	}
	return fyne.NewSize(0, 0)
}

// Objects returns the canvas objects for rendering
func (r *dialogBubbleRenderer) Objects() []fyne.CanvasObject {
	if r.bubble.visible {
		return []fyne.CanvasObject{r.container}
	}
	return []fyne.CanvasObject{}
}

// Refresh redraws the dialog bubble
func (r *dialogBubbleRenderer) Refresh() {
	if r.bubble.visible {
		r.container.Refresh()
	}
}

// Destroy cleans up dialog bubble resources
func (r *dialogBubbleRenderer) Destroy() {
	// No special cleanup needed
}
