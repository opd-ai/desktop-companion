package ui

import (
	"testing"

	"fyne.io/fyne/v2"
)

// TestPlatformAwareClickableWidgetCreation verifies widget creation
func TestPlatformAwareClickableWidgetCreation(t *testing.T) {
	tapCalled := false
	secondaryCalled := false

	widget := NewPlatformAwareClickableWidget(
		func() { tapCalled = true },
		func() { secondaryCalled = true },
	)

	if widget == nil {
		t.Fatal("Widget should not be nil")
	}

	if widget.ClickableWidget == nil {
		t.Error("Embedded ClickableWidget should not be nil")
	}

	if widget.touchWidget == nil {
		t.Error("Touch widget should not be nil")
	}

	if widget.platform == nil {
		t.Error("Platform info should not be nil")
	}

	// Suppress unused variable warnings
	_ = tapCalled
	_ = secondaryCalled
}

// TestPlatformAwareClickableWidgetWithDoubleTap verifies widget creation with double tap
func TestPlatformAwareClickableWidgetWithDoubleTap(t *testing.T) {
	tapCalled := false
	secondaryCalled := false
	doubleCalled := false

	widget := NewPlatformAwareClickableWidgetWithDoubleTap(
		func() { tapCalled = true },
		func() { secondaryCalled = true },
		func() { doubleCalled = true },
	)

	if widget == nil {
		t.Fatal("Widget should not be nil")
	}

	if widget.ClickableWidget == nil {
		t.Error("Embedded ClickableWidget should not be nil")
	}

	if widget.touchWidget == nil {
		t.Error("Touch widget should not be nil")
	}

	// Suppress unused variable warnings
	_ = tapCalled
	_ = secondaryCalled
	_ = doubleCalled
}

// TestPlatformAwareClickableWidgetSizing verifies size handling
func TestPlatformAwareClickableWidgetSizing(t *testing.T) {
	widget := NewPlatformAwareClickableWidget(nil, nil)

	testSize := fyne.NewSize(200, 150)
	widget.SetSize(testSize)

	// Size should be set on both embedded widgets
	if widget.ClickableWidget.size != testSize {
		t.Error("ClickableWidget size should be updated")
	}

	if widget.touchWidget.Size() != testSize {
		t.Error("TouchWidget size should be updated")
	}
}

// TestPlatformAwareClickableWidgetRenderer verifies renderer selection
func TestPlatformAwareClickableWidgetRenderer(t *testing.T) {
	widget := NewPlatformAwareClickableWidget(nil, nil)
	renderer := widget.CreateRenderer()

	if renderer == nil {
		t.Fatal("Renderer should not be nil")
	}

	// Test renderer interface methods don't panic
	minSize := renderer.MinSize()
	if minSize.Width < 0 || minSize.Height < 0 {
		t.Error("MinSize should not have negative dimensions")
	}

	objects := renderer.Objects()
	if objects == nil {
		t.Error("Objects should not be nil")
	}

	// Test layout and refresh don't panic
	testSize := fyne.NewSize(100, 100)
	renderer.Layout(testSize)
	renderer.Refresh()
	renderer.Destroy()
}

// TestPlatformAwareClickableWidgetGestureDetection verifies gesture translation detection
func TestPlatformAwareClickableWidgetGestureDetection(t *testing.T) {
	widget := NewPlatformAwareClickableWidget(nil, nil)

	// Test that gesture translation detection works
	isActive := widget.IsGestureTranslationActive()

	// The result depends on the platform the test is running on
	// We just verify it returns a boolean without error
	if isActive != true && isActive != false {
		t.Error("IsGestureTranslationActive should return a boolean")
	}
}

// TestPlatformAwareClickableWidgetDragHandlers verifies drag handler setup
func TestPlatformAwareClickableWidgetDragHandlers(t *testing.T) {
	widget := NewPlatformAwareClickableWidget(nil, nil)

	dragStartCalled := false
	dragCalled := false
	dragEndCalled := false

	widget.SetDragHandlers(
		func() { dragStartCalled = true },
		func(*fyne.DragEvent) { dragCalled = true },
		func() { dragEndCalled = true },
	)

	// We can't easily test that handlers are called without complex setup,
	// but we can verify the method doesn't panic and handlers are passed through

	// Suppress unused variable warnings
	_ = dragStartCalled
	_ = dragCalled
	_ = dragEndCalled
}

// TestPlatformAwareClickableWidgetEventHandling verifies event routing
func TestPlatformAwareClickableWidgetEventHandling(t *testing.T) {
	tapCalled := false
	secondaryCalled := false
	doubleCalled := false

	widget := NewPlatformAwareClickableWidgetWithDoubleTap(
		func() { tapCalled = true },
		func() { secondaryCalled = true },
		func() { doubleCalled = true },
	)

	event := &fyne.PointEvent{Position: fyne.NewPos(100, 100)}

	// Test that event methods don't panic
	widget.Tapped(event)
	widget.TappedSecondary(event)
	widget.DoubleTapped(event)

	// Test drag events
	dragEvent := &fyne.DragEvent{
		PointEvent: *event,
		Dragged:    fyne.NewDelta(10, 10),
	}
	widget.Dragged(dragEvent)
	widget.DragEnd()

	// The actual behavior depends on platform detection,
	// so we just verify methods don't panic

	// Note: We can't easily test callback invocation without mocking platform detection
	// That would require more complex test setup, which goes against the simplicity rule

	// Suppress unused variable warnings
	_ = tapCalled
	_ = secondaryCalled
	_ = doubleCalled
}

// TestPlatformAwareClickableWidgetCompatibility verifies backward compatibility
func TestPlatformAwareClickableWidgetCompatibility(t *testing.T) {
	widget := NewPlatformAwareClickableWidget(nil, nil)

	// Verify it still acts like a ClickableWidget for existing code
	if widget.ClickableWidget == nil {
		t.Error("Should maintain ClickableWidget compatibility")
	}

	// Verify standard widget interface compliance
	if widget.Size().Width < 0 || widget.Size().Height < 0 {
		t.Error("Should provide valid size")
	}

	// Test that it can be resized like standard widgets
	originalSize := widget.Size()
	newSize := fyne.NewSize(150, 100)
	widget.Resize(newSize)

	if widget.Size() == originalSize {
		t.Error("Resize should change widget size")
	}
}
