package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// ClickableWidget is a custom widget that supports both left and right click
type ClickableWidget struct {
	widget.BaseWidget
	OnTapped          func()
	OnTappedSecondary func()
	size              fyne.Size
}

// NewClickableWidget creates a new clickable widget with left and right click support
func NewClickableWidget(onTapped, onTappedSecondary func()) *ClickableWidget {
	w := &ClickableWidget{
		OnTapped:          onTapped,
		OnTappedSecondary: onTappedSecondary,
	}
	w.ExtendBaseWidget(w)
	return w
}

// SetSize sets the size of the clickable widget
func (w *ClickableWidget) SetSize(size fyne.Size) {
	w.size = size
	w.Resize(size)
}

// Tapped handles left mouse clicks
func (w *ClickableWidget) Tapped(*fyne.PointEvent) {
	if w.OnTapped != nil {
		w.OnTapped()
	}
}

// TappedSecondary handles right mouse clicks
func (w *ClickableWidget) TappedSecondary(*fyne.PointEvent) {
	if w.OnTappedSecondary != nil {
		w.OnTappedSecondary()
	}
}

// CreateRenderer creates the renderer for this widget
func (w *ClickableWidget) CreateRenderer() fyne.WidgetRenderer {
	return &clickableWidgetRenderer{widget: w}
}

// clickableWidgetRenderer renders the clickable widget (invisible overlay)
type clickableWidgetRenderer struct {
	widget *ClickableWidget
}

// Layout arranges the widget components
func (r *clickableWidgetRenderer) Layout(size fyne.Size) {
	// This is an invisible overlay, no layout needed
}

// MinSize returns the minimum size
func (r *clickableWidgetRenderer) MinSize() fyne.Size {
	return r.widget.size
}

// Refresh updates the widget appearance
func (r *clickableWidgetRenderer) Refresh() {
	// This is an invisible overlay, no refresh needed
}

// Objects returns the objects that make up this renderer
func (r *clickableWidgetRenderer) Objects() []fyne.CanvasObject {
	// Return empty slice for invisible overlay
	return []fyne.CanvasObject{}
}

// Destroy cleans up the renderer
func (r *clickableWidgetRenderer) Destroy() {
	// Nothing to clean up
}
