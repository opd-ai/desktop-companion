// Package gestures provides touch gesture translation for mobile platforms.
// This module converts touch gestures to traditional mouse events to maintain
// compatibility with existing desktop interaction patterns.
package gestures

import (
	"sync"
	"time"

	"fyne.io/fyne/v2"

	"github.com/opd-ai/desktop-companion/internal/platform"
)

// GestureHandler translates touch gestures to mouse events based on platform capabilities.
// Uses platform detection to apply appropriate gesture translation only when needed.
type GestureHandler struct {
	platform    *platform.PlatformInfo
	config      *GestureConfig
	onTap       func()                // Single tap -> left click
	onLongPress func()                // Long press -> right click
	onDoubleTap func()                // Double tap -> double click
	onDragStart func()                // Pan start -> drag start
	onDrag      func(*fyne.DragEvent) // Pan -> drag
	onDragEnd   func()                // Pan end -> drag end

	// Gesture detection state - protected by mutex
	mu              sync.RWMutex
	lastTapTime     time.Time
	tapCount        int
	longPressTimer  *time.Timer
	longPressActive bool
	isDragging      bool
}

// GestureConfig defines timing and threshold parameters for gesture recognition.
// All values use conservative defaults that work well across different devices.
type GestureConfig struct {
	// DoubleTapWindow is the maximum time between taps to count as double tap
	DoubleTapWindow time.Duration

	// LongPressDuration is the minimum hold time to trigger long press
	LongPressDuration time.Duration

	// DragThreshold is the minimum distance to start a drag operation
	DragThreshold float32
}

// DefaultGestureConfig returns platform-appropriate gesture configuration.
// Uses conservative timing that works well for both experienced and new users.
func DefaultGestureConfig() *GestureConfig {
	return &GestureConfig{
		DoubleTapWindow:   500 * time.Millisecond, // Standard double-click timing
		LongPressDuration: 600 * time.Millisecond, // Slightly longer than iOS (500ms)
		DragThreshold:     10.0,                   // 10 pixels minimum drag distance
	}
}

// NewGestureHandler creates a gesture handler with platform-aware behavior.
// Only applies gesture translation on mobile platforms, maintaining desktop compatibility.
func NewGestureHandler(platform *platform.PlatformInfo, config *GestureConfig) *GestureHandler {
	if config == nil {
		config = DefaultGestureConfig()
	}

	return &GestureHandler{
		platform: platform,
		config:   config,
	}
}

// SetTapHandler sets the callback for single tap events (translated to left click).
func (gh *GestureHandler) SetTapHandler(handler func()) {
	gh.onTap = handler
}

// SetLongPressHandler sets the callback for long press events (translated to right click).
func (gh *GestureHandler) SetLongPressHandler(handler func()) {
	gh.onLongPress = handler
}

// SetDoubleTapHandler sets the callback for double tap events (translated to double click).
func (gh *GestureHandler) SetDoubleTapHandler(handler func()) {
	gh.onDoubleTap = handler
}

// SetDragHandlers sets the callbacks for drag operations (translated from pan gestures).
func (gh *GestureHandler) SetDragHandlers(onStart func(), onDrag func(*fyne.DragEvent), onEnd func()) {
	gh.onDragStart = onStart
	gh.onDrag = onDrag
	gh.onDragEnd = onEnd
}

// IsGestureTranslationNeeded returns true if the platform requires gesture translation.
// Desktop platforms return false to maintain existing behavior unchanged.
func (gh *GestureHandler) IsGestureTranslationNeeded() bool {
	return gh.platform.IsMobile() || gh.platform.HasTouch()
}

// HandleTouchStart processes the start of a touch event.
// This begins gesture detection for tap, long press, and drag operations.
func (gh *GestureHandler) HandleTouchStart(pos fyne.Position) {
	if !gh.IsGestureTranslationNeeded() {
		return // Desktop platforms use existing mouse handling
	}

	gh.mu.Lock()
	defer gh.mu.Unlock()

	now := time.Now()

	// Cancel any existing long press timer
	if gh.longPressTimer != nil {
		gh.longPressTimer.Stop()
		gh.longPressTimer = nil
	}

	// Check for double tap
	if now.Sub(gh.lastTapTime) <= gh.config.DoubleTapWindow {
		gh.tapCount++
	} else {
		gh.tapCount = 1
	}

	gh.lastTapTime = now
	gh.longPressActive = false

	// Start long press timer
	gh.longPressTimer = time.AfterFunc(gh.config.LongPressDuration, func() {
		gh.handleLongPress()
	})
}

// HandleTouchEnd processes the end of a touch event.
// This triggers tap or double tap events if appropriate conditions are met.
func (gh *GestureHandler) HandleTouchEnd(pos fyne.Position) {
	if !gh.IsGestureTranslationNeeded() {
		return
	}

	gh.mu.Lock()

	// Cancel long press timer if still active
	if gh.longPressTimer != nil {
		gh.longPressTimer.Stop()
		gh.longPressTimer = nil
	}

	// End drag if active
	if gh.isDragging {
		gh.isDragging = false
		gh.mu.Unlock()
		if gh.onDragEnd != nil {
			gh.onDragEnd()
		}
		return
	}

	// Don't trigger tap if long press was already handled
	if gh.longPressActive {
		gh.longPressActive = false
		gh.mu.Unlock()
		return
	}

	// Capture state for the delayed tap handler
	tapCount := gh.tapCount
	onTap := gh.onTap
	onDoubleTap := gh.onDoubleTap
	gh.mu.Unlock()

	// Handle taps with a small delay to detect double taps
	go func() {
		time.Sleep(gh.config.DoubleTapWindow)

		if tapCount >= 2 && onDoubleTap != nil {
			onDoubleTap()
		} else if tapCount == 1 && onTap != nil {
			onTap()
		}

		gh.mu.Lock()
		gh.tapCount = 0
		gh.mu.Unlock()
	}()
}

// HandleTouchMove processes touch movement for drag operations.
// Converts pan gestures to drag events compatible with existing drag handlers.
func (gh *GestureHandler) HandleTouchMove(event *fyne.DragEvent) {
	if !gh.IsGestureTranslationNeeded() {
		return
	}

	gh.mu.Lock()
	defer gh.mu.Unlock()

	// Cancel long press if touch moves beyond threshold
	if gh.longPressTimer != nil && !gh.isDragging {
		if abs(event.Dragged.DX) > gh.config.DragThreshold || abs(event.Dragged.DY) > gh.config.DragThreshold {
			gh.longPressTimer.Stop()
			gh.longPressTimer = nil

			// Start drag operation
			gh.isDragging = true
			onDragStart := gh.onDragStart
			gh.mu.Unlock()
			if onDragStart != nil {
				onDragStart()
			}
			gh.mu.Lock()
		}
	}

	// Continue drag operation
	if gh.isDragging {
		onDrag := gh.onDrag
		gh.mu.Unlock()
		if onDrag != nil {
			onDrag(event)
		}
		gh.mu.Lock()
	}
}

// handleLongPress processes long press gesture detection.
// This is called by the timer when long press duration is reached.
func (gh *GestureHandler) handleLongPress() {
	gh.mu.Lock()
	defer gh.mu.Unlock()

	if gh.isDragging {
		return // Don't trigger long press during drag
	}

	gh.longPressActive = true
	onLongPress := gh.onLongPress

	// Call handler outside of lock to avoid potential deadlock
	gh.mu.Unlock()
	if onLongPress != nil {
		onLongPress()
	}
	gh.mu.Lock() // Reacquire for defer
}

// abs returns the absolute value of a float32.
// Simple utility function to avoid importing math package for one function.
func abs(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}
