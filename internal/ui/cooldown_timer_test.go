package ui

import (
	"sync/atomic"
	"testing"
	"time"
)

// TestNewCooldownTimer tests the creation of a new cooldown timer
func TestNewCooldownTimer(t *testing.T) {
	timer := NewCooldownTimer()

	if timer == nil {
		t.Fatal("NewCooldownTimer returned nil")
	}

	if timer.progressBar == nil {
		t.Error("CooldownTimer progressBar is nil")
	}

	if timer.timeLabel == nil {
		t.Error("CooldownTimer timeLabel is nil")
	}

	if timer.IsActive() {
		t.Error("New timer should not be active initially")
	}

	if timer.GetRemainingTime() != 0 {
		t.Error("New timer should have 0 remaining time")
	}
}

// TestCooldownTimer_StartCooldown tests starting a cooldown
func TestCooldownTimer_StartCooldown(t *testing.T) {
	timer := NewCooldownTimer()

	// Start a 1 second cooldown
	duration := 1 * time.Second
	timer.StartCooldown(duration)

	if !timer.IsActive() {
		t.Error("Timer should be active after starting cooldown")
	}

	remaining := timer.GetRemainingTime()
	if remaining <= 0 || remaining > duration {
		t.Errorf("Expected remaining time between 0 and %v, got %v", duration, remaining)
	}
}

// TestCooldownTimer_SetOnComplete tests the completion callback
func TestCooldownTimer_SetOnComplete(t *testing.T) {
	timer := NewCooldownTimer()

	var callbackCalled atomic.Bool
	timer.SetOnComplete(func() {
		callbackCalled.Store(true)
	})

	// Start a very short cooldown
	timer.StartCooldown(10 * time.Millisecond)

	// Wait for cooldown to complete with generous margin
	time.Sleep(200 * time.Millisecond)

	if !callbackCalled.Load() {
		t.Error("Completion callback should have been called")
	}

	if timer.IsActive() {
		t.Error("Timer should not be active after completion")
	}

	if timer.GetRemainingTime() != 0 {
		t.Error("Timer should have 0 remaining time after completion")
	}
}

// TestCooldownTimer_ThreadSafety tests concurrent access to timer methods
func TestCooldownTimer_ThreadSafety(t *testing.T) {
	timer := NewCooldownTimer()

	// Start multiple goroutines accessing timer methods
	done := make(chan bool, 3)

	go func() {
		timer.StartCooldown(100 * time.Millisecond)
		done <- true
	}()

	go func() {
		for i := 0; i < 10; i++ {
			timer.IsActive()
			timer.GetRemainingTime()
			time.Sleep(5 * time.Millisecond)
		}
		done <- true
	}()

	go func() {
		timer.SetOnComplete(func() {})
		done <- true
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		select {
		case <-done:
		case <-time.After(500 * time.Millisecond):
			t.Fatal("Timeout waiting for goroutines to complete")
		}
	}
}

// TestCooldownTimer_CreateRenderer tests the widget renderer creation
func TestCooldownTimer_CreateRenderer(t *testing.T) {
	timer := NewCooldownTimer()

	renderer := timer.CreateRenderer()
	if renderer == nil {
		t.Error("CreateRenderer should not return nil")
	}

	// Test that renderer can be called multiple times
	renderer2 := timer.CreateRenderer()
	if renderer2 == nil {
		t.Error("CreateRenderer should work on subsequent calls")
	}
}

// TestCooldownTimer_ZeroDuration tests handling of zero duration cooldown
func TestCooldownTimer_ZeroDuration(t *testing.T) {
	timer := NewCooldownTimer()

	timer.StartCooldown(0)

	// Should complete immediately
	time.Sleep(10 * time.Millisecond)

	if timer.IsActive() {
		t.Error("Timer should not be active for zero duration cooldown")
	}
}

// TestCooldownTimer_NegativeDuration tests handling of negative duration
func TestCooldownTimer_NegativeDuration(t *testing.T) {
	timer := NewCooldownTimer()

	timer.StartCooldown(-1 * time.Second)

	// Should complete immediately
	time.Sleep(10 * time.Millisecond)

	if timer.IsActive() {
		t.Error("Timer should not be active for negative duration cooldown")
	}
}
