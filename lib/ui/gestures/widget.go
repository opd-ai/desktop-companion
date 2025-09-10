package gestures

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"github.com/opd-ai/desktop-companion/internal/platform"
)

// TouchAwareWidget extends the functionality of a standard widget to support
// touch gesture translation on mobile platforms while maintaining desktop compatibility.
type TouchAwareWidget struct {
	widget.BaseWidget
	gestureHandler    *GestureHandler
	onTapped          func()
	onTappedSecondary func()
	onDoubleTapped    func()
	size              fyne.Size
}

// NewTouchAwareWidget creates a widget that handles both traditional mouse events
// and touch gestures based on the platform capabilities.
func NewTouchAwareWidget(platform *platform.PlatformInfo, onTapped, onTappedSecondary, onDoubleTapped func()) *TouchAwareWidget {
	w := &TouchAwareWidget{
		onTapped:          onTapped,
		onTappedSecondary: onTappedSecondary,
		onDoubleTapped:    onDoubleTapped,
	}

	// Create gesture handler for touch platforms
	w.gestureHandler = NewGestureHandler(platform, DefaultGestureConfig())
	w.gestureHandler.SetTapHandler(onTapped)
	w.gestureHandler.SetLongPressHandler(onTappedSecondary)
	w.gestureHandler.SetDoubleTapHandler(onDoubleTapped)

	w.ExtendBaseWidget(w)
	return w
}

// SetSize sets the size of the touch-aware widget
func (w *TouchAwareWidget) SetSize(size fyne.Size) {
	w.size = size
	w.Resize(size)
}

// Tapped handles traditional mouse clicks and touch taps
func (w *TouchAwareWidget) Tapped(event *fyne.PointEvent) {
	if w.gestureHandler.IsGestureTranslationNeeded() {
		// On touch platforms, handle through gesture system
		w.gestureHandler.HandleTouchStart(event.Position)
		w.gestureHandler.HandleTouchEnd(event.Position)
	} else {
		// On desktop platforms, use direct callback
		if w.onTapped != nil {
			w.onTapped()
		}
	}
}

// TappedSecondary handles right-click events on desktop platforms
func (w *TouchAwareWidget) TappedSecondary(event *fyne.PointEvent) {
	if !w.gestureHandler.IsGestureTranslationNeeded() {
		// Only handle direct right-click on desktop platforms
		if w.onTappedSecondary != nil {
			w.onTappedSecondary()
		}
	}
	// On touch platforms, right-click is handled via long press gesture
}

// DoubleTapped handles double-click events on desktop platforms
func (w *TouchAwareWidget) DoubleTapped(event *fyne.PointEvent) {
	if !w.gestureHandler.IsGestureTranslationNeeded() {
		// Only handle direct double-click on desktop platforms
		if w.onDoubleTapped != nil {
			w.onDoubleTapped()
		}
	}
	// On touch platforms, double-click is handled via double tap gesture
}

// Dragged handles drag events for both mouse and touch platforms
func (w *TouchAwareWidget) Dragged(event *fyne.DragEvent) {
	if w.gestureHandler.IsGestureTranslationNeeded() {
		// On touch platforms, handle through gesture system
		w.gestureHandler.HandleTouchMove(event)
	}
	// Desktop drag handling can be added here if needed
}

// DragEnd handles the end of drag operations
func (w *TouchAwareWidget) DragEnd() {
	if w.gestureHandler.IsGestureTranslationNeeded() {
		// Touch drag end is handled in HandleTouchEnd
		// This could be extended for desktop drag end if needed
	}
}

// SetDragHandlers configures drag behavior for the widget
func (w *TouchAwareWidget) SetDragHandlers(onStart func(), onDrag func(*fyne.DragEvent), onEnd func()) {
	w.gestureHandler.SetDragHandlers(onStart, onDrag, onEnd)
}

// CreateRenderer creates the renderer for this widget
func (w *TouchAwareWidget) CreateRenderer() fyne.WidgetRenderer {
	return &touchAwareWidgetRenderer{widget: w}
}

// touchAwareWidgetRenderer renders the touch-aware widget (invisible overlay)
type touchAwareWidgetRenderer struct {
	widget *TouchAwareWidget
}

// Layout positions the widget components
func (r *touchAwareWidgetRenderer) Layout(size fyne.Size) {
	r.widget.size = size
}

// MinSize returns the minimum size for the widget
func (r *touchAwareWidgetRenderer) MinSize() fyne.Size {
	return r.widget.size
}

// Refresh updates the widget appearance
func (r *touchAwareWidgetRenderer) Refresh() {
	// Nothing to refresh for invisible overlay
}

// Objects returns the objects to be rendered (none for invisible overlay)
func (r *touchAwareWidgetRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{} // Invisible overlay
}

// Destroy cleans up the renderer resources
func (r *touchAwareWidgetRenderer) Destroy() {
	// Nothing to destroy
}
