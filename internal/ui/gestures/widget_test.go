package gestures

import (
	"testing"

	"fyne.io/fyne/v2"

	"github.com/opd-ai/desktop-companion/internal/platform"
)

// TestTouchAwareWidgetCreation verifies basic widget creation and configuration
func TestTouchAwareWidgetCreation(t *testing.T) {
	platform := &platform.PlatformInfo{
		OS:           "android",
		FormFactor:   "mobile",
		InputMethods: []string{"touch"},
	}

	tapCalled := false
	secondaryCalled := false
	doubleCalled := false

	widget := NewTouchAwareWidget(
		platform,
		func() { tapCalled = true },
		func() { secondaryCalled = true },
		func() { doubleCalled = true },
	)

	if widget == nil {
		t.Fatal("Widget should not be nil")
	}

	if widget.gestureHandler == nil {
		t.Error("Gesture handler should be initialized")
	}

	if widget.onTapped == nil {
		t.Error("Tap callback should be set")
	}

	if widget.onTappedSecondary == nil {
		t.Error("Secondary tap callback should be set")
	}

	if widget.onDoubleTapped == nil {
		t.Error("Double tap callback should be set")
	}

	// Suppress unused variable warnings
	_ = tapCalled
	_ = secondaryCalled
	_ = doubleCalled
}

// TestTouchAwareWidgetDesktopBehavior verifies desktop platform behavior
func TestTouchAwareWidgetDesktopBehavior(t *testing.T) {
	platform := &platform.PlatformInfo{
		OS:           "windows",
		FormFactor:   "desktop",
		InputMethods: []string{"mouse", "keyboard"},
	}

	tapCalled := false
	secondaryCalled := false
	doubleCalled := false

	widget := NewTouchAwareWidget(
		platform,
		func() { tapCalled = true },
		func() { secondaryCalled = true },
		func() { doubleCalled = true },
	)

	// Test that desktop doesn't need gesture translation
	if widget.gestureHandler.IsGestureTranslationNeeded() {
		t.Error("Desktop platform should not need gesture translation")
	}

	// Test direct tap (should work on desktop)
	event := &fyne.PointEvent{Position: fyne.NewPos(100, 100)}
	widget.Tapped(event)

	if !tapCalled {
		t.Error("Desktop tap should call callback directly")
	}

	// Test secondary tap (should work on desktop)
	widget.TappedSecondary(event)

	if !secondaryCalled {
		t.Error("Desktop secondary tap should call callback directly")
	}

	// Test double tap (should work on desktop)
	widget.DoubleTapped(event)

	if !doubleCalled {
		t.Error("Desktop double tap should call callback directly")
	}
}

// TestTouchAwareWidgetMobileBehavior verifies mobile platform behavior
func TestTouchAwareWidgetMobileBehavior(t *testing.T) {
	platform := &platform.PlatformInfo{
		OS:           "android",
		FormFactor:   "mobile",
		InputMethods: []string{"touch"},
	}

	tapCalled := false
	secondaryCalled := false
	doubleCalled := false

	widget := NewTouchAwareWidget(
		platform,
		func() { tapCalled = true },
		func() { secondaryCalled = true },
		func() { doubleCalled = true },
	)

	// Test that mobile needs gesture translation
	if !widget.gestureHandler.IsGestureTranslationNeeded() {
		t.Error("Mobile platform should need gesture translation")
	}

	// Test that secondary tap on mobile doesn't call callback directly
	// (it should go through gesture system instead)
	event := &fyne.PointEvent{Position: fyne.NewPos(100, 100)}
	widget.TappedSecondary(event)

	if secondaryCalled {
		t.Error("Mobile secondary tap should not call callback directly")
	}

	// Test that double tap on mobile doesn't call callback directly
	widget.DoubleTapped(event)

	if doubleCalled {
		t.Error("Mobile double tap should not call callback directly")
	}

	// Suppress unused variable warning
	_ = tapCalled
}

// TestTouchAwareWidgetRenderer verifies widget rendering functionality
func TestTouchAwareWidgetRenderer(t *testing.T) {
	platform := &platform.PlatformInfo{
		OS:           "android",
		FormFactor:   "mobile",
		InputMethods: []string{"touch"},
	}

	widget := NewTouchAwareWidget(platform, nil, nil, nil)
	renderer := widget.CreateRenderer()

	if renderer == nil {
		t.Fatal("Renderer should not be nil")
	}

	// Test renderer interface methods
	minSize := renderer.MinSize()
	if minSize.Width < 0 || minSize.Height < 0 {
		t.Error("MinSize should not have negative dimensions")
	}

	objects := renderer.Objects()
	if objects == nil {
		t.Error("Objects should not be nil (even if empty)")
	}

	// Test that it's an invisible overlay (no objects)
	if len(objects) != 0 {
		t.Error("Touch-aware widget should be an invisible overlay")
	}

	// Test layout and refresh don't panic
	testSize := fyne.NewSize(100, 100)
	renderer.Layout(testSize)
	renderer.Refresh()
	renderer.Destroy()
}

// TestTouchAwareWidgetSizing verifies widget sizing functionality
func TestTouchAwareWidgetSizing(t *testing.T) {
	platform := &platform.PlatformInfo{
		OS:           "android",
		FormFactor:   "mobile",
		InputMethods: []string{"touch"},
	}

	widget := NewTouchAwareWidget(platform, nil, nil, nil)

	// Test setting size
	testSize := fyne.NewSize(200, 150)
	widget.SetSize(testSize)

	if widget.size != testSize {
		t.Error("Widget size should be updated by SetSize")
	}
}

// TestTouchAwareWidgetDragHandlers verifies drag handler configuration
func TestTouchAwareWidgetDragHandlers(t *testing.T) {
	platform := &platform.PlatformInfo{
		OS:           "android",
		FormFactor:   "mobile",
		InputMethods: []string{"touch"},
	}

	widget := NewTouchAwareWidget(platform, nil, nil, nil)

	dragStartCalled := false
	dragCalled := false
	dragEndCalled := false

	widget.SetDragHandlers(
		func() { dragStartCalled = true },
		func(*fyne.DragEvent) { dragCalled = true },
		func() { dragEndCalled = true },
	)

	// Verify handlers are set on the gesture handler
	if widget.gestureHandler.onDragStart == nil {
		t.Error("Drag start handler should be set")
	}

	if widget.gestureHandler.onDrag == nil {
		t.Error("Drag handler should be set")
	}

	if widget.gestureHandler.onDragEnd == nil {
		t.Error("Drag end handler should be set")
	}

	// Suppress unused variable warnings
	_ = dragStartCalled
	_ = dragCalled
	_ = dragEndCalled
}
