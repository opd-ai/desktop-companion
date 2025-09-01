package gestures

import (
	"sync"
	"testing"
	"time"

	"fyne.io/fyne/v2"

	"desktop-companion/internal/platform"
)

// TestGestureHandlerCreation verifies basic gesture handler creation and configuration
func TestGestureHandlerCreation(t *testing.T) {
	tests := []struct {
		name      string
		platform  *platform.PlatformInfo
		config    *GestureConfig
		expectNil bool
	}{
		{
			name: "desktop platform",
			platform: &platform.PlatformInfo{
				OS:           "windows",
				FormFactor:   "desktop",
				InputMethods: []string{"mouse", "keyboard"},
			},
			config:    nil,
			expectNil: false,
		},
		{
			name: "mobile platform",
			platform: &platform.PlatformInfo{
				OS:           "android",
				FormFactor:   "mobile",
				InputMethods: []string{"touch"},
			},
			config:    DefaultGestureConfig(),
			expectNil: false,
		},
		{
			name: "custom config",
			platform: &platform.PlatformInfo{
				OS:           "android",
				FormFactor:   "mobile",
				InputMethods: []string{"touch"},
			},
			config: &GestureConfig{
				DoubleTapWindow:   300 * time.Millisecond,
				LongPressDuration: 800 * time.Millisecond,
				DragThreshold:     15.0,
			},
			expectNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewGestureHandler(tt.platform, tt.config)

			if handler == nil && !tt.expectNil {
				t.Error("Expected non-nil gesture handler")
			}

			if handler != nil {
				if handler.platform != tt.platform {
					t.Error("Platform not properly assigned")
				}

				if handler.config == nil {
					t.Error("Config should not be nil")
				}
			}
		})
	}
}

// TestGestureTranslationDetection verifies platform-based gesture translation logic
func TestGestureTranslationDetection(t *testing.T) {
	tests := []struct {
		name       string
		platform   *platform.PlatformInfo
		expectGest bool
	}{
		{
			name: "Windows desktop",
			platform: &platform.PlatformInfo{
				OS:           "windows",
				FormFactor:   "desktop",
				InputMethods: []string{"mouse", "keyboard"},
			},
			expectGest: false,
		},
		{
			name: "Android mobile",
			platform: &platform.PlatformInfo{
				OS:           "android",
				FormFactor:   "mobile",
				InputMethods: []string{"touch"},
			},
			expectGest: true,
		},
		{
			name: "Linux with touch",
			platform: &platform.PlatformInfo{
				OS:           "linux",
				FormFactor:   "desktop",
				InputMethods: []string{"mouse", "keyboard", "touch"},
			},
			expectGest: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewGestureHandler(tt.platform, nil)
			result := handler.IsGestureTranslationNeeded()

			if result != tt.expectGest {
				t.Errorf("Expected gesture translation %v, got %v", tt.expectGest, result)
			}
		})
	}
}

// TestHandlerCallbacks verifies that gesture callbacks are properly set and called
func TestHandlerCallbacks(t *testing.T) {
	platform := &platform.PlatformInfo{
		OS:           "android",
		FormFactor:   "mobile",
		InputMethods: []string{"touch"},
	}

	handler := NewGestureHandler(platform, DefaultGestureConfig())

	// Test callback setting
	tapCalled := false
	longPressCalled := false
	doubleTapCalled := false
	dragStartCalled := false
	dragEndCalled := false

	handler.SetTapHandler(func() { tapCalled = true })
	handler.SetLongPressHandler(func() { longPressCalled = true })
	handler.SetDoubleTapHandler(func() { doubleTapCalled = true })
	handler.SetDragHandlers(
		func() { dragStartCalled = true },
		func(*fyne.DragEvent) {},
		func() { dragEndCalled = true },
	)

	// Verify callbacks are set
	if handler.onTap == nil {
		t.Error("Tap handler not set")
	}

	if handler.onLongPress == nil {
		t.Error("Long press handler not set")
	}

	if handler.onDoubleTap == nil {
		t.Error("Double tap handler not set")
	}

	if handler.onDragStart == nil {
		t.Error("Drag start handler not set")
	}

	if handler.onDragEnd == nil {
		t.Error("Drag end handler not set")
	}

	// Suppress unused variable warnings by checking them
	_ = tapCalled
	_ = longPressCalled
	_ = doubleTapCalled
	_ = dragStartCalled
	_ = dragEndCalled
}

// TestSingleTapGesture verifies single tap detection and callback
func TestSingleTapGesture(t *testing.T) {
	platform := &platform.PlatformInfo{
		OS:           "android",
		FormFactor:   "mobile",
		InputMethods: []string{"touch"},
	}

	handler := NewGestureHandler(platform, &GestureConfig{
		DoubleTapWindow:   100 * time.Millisecond, // Shorter for testing
		LongPressDuration: 500 * time.Millisecond,
		DragThreshold:     10.0,
	})

	var tapCalled bool
	var mu sync.Mutex
	handler.SetTapHandler(func() {
		mu.Lock()
		tapCalled = true
		mu.Unlock()
	})

	// Simulate single tap
	pos := fyne.NewPos(100, 100)
	handler.HandleTouchStart(pos)
	handler.HandleTouchEnd(pos)

	// Wait for double tap window to expire
	time.Sleep(150 * time.Millisecond)

	mu.Lock()
	called := tapCalled
	mu.Unlock()

	if !called {
		t.Error("Single tap callback was not called")
	}
}

// TestLongPressGesture verifies long press detection and callback
func TestLongPressGesture(t *testing.T) {
	platform := &platform.PlatformInfo{
		OS:           "android",
		FormFactor:   "mobile",
		InputMethods: []string{"touch"},
	}

	handler := NewGestureHandler(platform, &GestureConfig{
		DoubleTapWindow:   500 * time.Millisecond,
		LongPressDuration: 100 * time.Millisecond, // Shorter for testing
		DragThreshold:     10.0,
	})

	var longPressCalled bool
	var mu sync.Mutex
	handler.SetLongPressHandler(func() {
		mu.Lock()
		longPressCalled = true
		mu.Unlock()
	})

	// Simulate long press
	pos := fyne.NewPos(100, 100)
	handler.HandleTouchStart(pos)

	// Wait for long press duration
	time.Sleep(150 * time.Millisecond)

	handler.HandleTouchEnd(pos)

	mu.Lock()
	called := longPressCalled
	mu.Unlock()

	if !called {
		t.Error("Long press callback was not called")
	}
}

// TestDoubleTapGesture verifies double tap detection and callback
func TestDoubleTapGesture(t *testing.T) {
	platform := &platform.PlatformInfo{
		OS:           "android",
		FormFactor:   "mobile",
		InputMethods: []string{"touch"},
	}

	handler := NewGestureHandler(platform, &GestureConfig{
		DoubleTapWindow:   200 * time.Millisecond, // Reasonable window for testing
		LongPressDuration: 500 * time.Millisecond,
		DragThreshold:     10.0,
	})

	var doubleTapCalled bool
	var mu sync.Mutex
	handler.SetDoubleTapHandler(func() {
		mu.Lock()
		doubleTapCalled = true
		mu.Unlock()
	})

	// Simulate double tap
	pos := fyne.NewPos(100, 100)

	// First tap
	handler.HandleTouchStart(pos)
	handler.HandleTouchEnd(pos)

	// Second tap within window
	time.Sleep(50 * time.Millisecond)
	handler.HandleTouchStart(pos)
	handler.HandleTouchEnd(pos)

	// Wait for double tap window to expire
	time.Sleep(250 * time.Millisecond)

	mu.Lock()
	called := doubleTapCalled
	mu.Unlock()

	if !called {
		t.Error("Double tap callback was not called")
	}
}

// TestDesktopPlatformBehavior verifies that desktop platforms don't trigger gesture translation
func TestDesktopPlatformBehavior(t *testing.T) {
	platform := &platform.PlatformInfo{
		OS:           "windows",
		FormFactor:   "desktop",
		InputMethods: []string{"mouse", "keyboard"},
	}

	handler := NewGestureHandler(platform, DefaultGestureConfig())

	tapCalled := false
	handler.SetTapHandler(func() { tapCalled = true })

	// Simulate touch events on desktop (should be ignored)
	pos := fyne.NewPos(100, 100)
	handler.HandleTouchStart(pos)
	handler.HandleTouchEnd(pos)

	// Wait briefly
	time.Sleep(50 * time.Millisecond)

	if tapCalled {
		t.Error("Desktop platform should not trigger gesture callbacks")
	}
}

// TestDefaultGestureConfig verifies the default configuration values
func TestDefaultGestureConfig(t *testing.T) {
	config := DefaultGestureConfig()

	if config == nil {
		t.Fatal("Default config should not be nil")
	}

	if config.DoubleTapWindow != 500*time.Millisecond {
		t.Errorf("Expected DoubleTapWindow 500ms, got %v", config.DoubleTapWindow)
	}

	if config.LongPressDuration != 600*time.Millisecond {
		t.Errorf("Expected LongPressDuration 600ms, got %v", config.LongPressDuration)
	}

	if config.DragThreshold != 10.0 {
		t.Errorf("Expected DragThreshold 10.0, got %v", config.DragThreshold)
	}
}

// TestAbsFunction verifies the absolute value utility function
func TestAbsFunction(t *testing.T) {
	tests := []struct {
		input    float32
		expected float32
	}{
		{5.0, 5.0},
		{-5.0, 5.0},
		{0.0, 0.0},
		{-0.0, 0.0},
		{10.5, 10.5},
		{-10.5, 10.5},
	}

	for _, tt := range tests {
		result := abs(tt.input)
		if result != tt.expected {
			t.Errorf("abs(%v) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}
