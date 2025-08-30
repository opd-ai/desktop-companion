// Package ui provides platform-aware UI integration for touch gesture support.
// This module extends existing UI components with mobile gesture capabilities.
package ui

import (
	"fyne.io/fyne/v2"

	"desktop-companion/internal/platform"
	"desktop-companion/internal/ui/gestures"
)

// PlatformAwareClickableWidget extends ClickableWidget with touch gesture support.
// Uses platform detection to provide appropriate interaction handling for both
// desktop and mobile environments.
type PlatformAwareClickableWidget struct {
	*ClickableWidget                            // Embedded for backward compatibility
	touchWidget      *gestures.TouchAwareWidget // Touch-specific behavior
	platform         *platform.PlatformInfo
}

// NewPlatformAwareClickableWidget creates a clickable widget that adapts to the platform.
// On desktop platforms, uses traditional mouse events. On mobile platforms, translates
// touch gestures to equivalent mouse events for seamless compatibility.
func NewPlatformAwareClickableWidget(onTapped, onTappedSecondary func()) *PlatformAwareClickableWidget {
	// Get platform information for adaptive behavior
	platformInfo := platform.GetPlatformInfo()

	// Create traditional clickable widget for desktop compatibility
	clickableWidget := NewClickableWidget(onTapped, onTappedSecondary)

	// Create touch-aware widget for mobile platforms
	touchWidget := gestures.NewTouchAwareWidget(
		platformInfo,
		onTapped,
		onTappedSecondary,
		nil, // Double tap handled separately if needed
	)

	return &PlatformAwareClickableWidget{
		ClickableWidget: clickableWidget,
		touchWidget:     touchWidget,
		platform:        platformInfo,
	}
}

// NewPlatformAwareClickableWidgetWithDoubleTap creates a clickable widget with double tap support.
// Provides all three interaction types: tap, secondary tap (long press), and double tap.
func NewPlatformAwareClickableWidgetWithDoubleTap(onTapped, onTappedSecondary, onDoubleTapped func()) *PlatformAwareClickableWidget {
	platformInfo := platform.GetPlatformInfo()

	clickableWidget := NewClickableWidget(onTapped, onTappedSecondary)
	touchWidget := gestures.NewTouchAwareWidget(
		platformInfo,
		onTapped,
		onTappedSecondary,
		onDoubleTapped,
	)

	return &PlatformAwareClickableWidget{
		ClickableWidget: clickableWidget,
		touchWidget:     touchWidget,
		platform:        platformInfo,
	}
}

// SetSize sets the size for both traditional and touch widgets
func (w *PlatformAwareClickableWidget) SetSize(size fyne.Size) {
	w.ClickableWidget.SetSize(size)
	w.touchWidget.SetSize(size)
}

// Tapped handles tap events using platform-appropriate method
func (w *PlatformAwareClickableWidget) Tapped(event *fyne.PointEvent) {
	if w.platform.IsMobile() || w.platform.HasTouch() {
		// Use touch-aware handling on mobile platforms
		w.touchWidget.Tapped(event)
	} else {
		// Use traditional handling on desktop platforms
		w.ClickableWidget.Tapped(event)
	}
}

// TappedSecondary handles secondary tap events using platform-appropriate method
func (w *PlatformAwareClickableWidget) TappedSecondary(event *fyne.PointEvent) {
	if w.platform.IsMobile() || w.platform.HasTouch() {
		// Use touch-aware handling on mobile platforms
		w.touchWidget.TappedSecondary(event)
	} else {
		// Use traditional handling on desktop platforms
		w.ClickableWidget.TappedSecondary(event)
	}
}

// DoubleTapped handles double tap events using platform-appropriate method
func (w *PlatformAwareClickableWidget) DoubleTapped(event *fyne.PointEvent) {
	if w.platform.IsMobile() || w.platform.HasTouch() {
		// Use touch-aware handling on mobile platforms
		w.touchWidget.DoubleTapped(event)
	}
	// Desktop platforms can add double-click support here if needed
}

// Dragged handles drag events using platform-appropriate method
func (w *PlatformAwareClickableWidget) Dragged(event *fyne.DragEvent) {
	if w.platform.IsMobile() || w.platform.HasTouch() {
		// Use touch-aware handling on mobile platforms
		w.touchWidget.Dragged(event)
	}
	// Desktop drag handling can be added here if needed
}

// DragEnd handles drag end events using platform-appropriate method
func (w *PlatformAwareClickableWidget) DragEnd() {
	if w.platform.IsMobile() || w.platform.HasTouch() {
		// Use touch-aware handling on mobile platforms
		w.touchWidget.DragEnd()
	}
	// Desktop drag end handling can be added here if needed
}

// SetDragHandlers configures drag behavior for touch platforms
func (w *PlatformAwareClickableWidget) SetDragHandlers(onStart func(), onDrag func(*fyne.DragEvent), onEnd func()) {
	w.touchWidget.SetDragHandlers(onStart, onDrag, onEnd)
}

// CreateRenderer creates the appropriate renderer based on platform
func (w *PlatformAwareClickableWidget) CreateRenderer() fyne.WidgetRenderer {
	if w.platform.IsMobile() || w.platform.HasTouch() {
		// Use touch-aware renderer on mobile platforms
		return w.touchWidget.CreateRenderer()
	}
	// Use traditional renderer on desktop platforms
	return w.ClickableWidget.CreateRenderer()
}

// IsGestureTranslationActive returns true if gesture translation is being used
func (w *PlatformAwareClickableWidget) IsGestureTranslationActive() bool {
	return w.platform.IsMobile() || w.platform.HasTouch()
}
