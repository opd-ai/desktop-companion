package ui

import (
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// CooldownTimer provides visual indication of remaining cooldown time
// Follows existing UI patterns from StatsOverlay and AchievementNotification
// Uses standard library time.Duration and Fyne widgets only
type CooldownTimer struct {
	widget.BaseWidget
	progressBar *widget.ProgressBar
	timeLabel   *widget.Label
	endTime     time.Time
	isActive    bool
	mu          sync.RWMutex
	onComplete  func() // Optional callback when cooldown completes
}

// NewCooldownTimer creates a new cooldown timer widget
// Returns a fully configured timer that integrates with existing UI patterns
func NewCooldownTimer() *CooldownTimer {
	timer := &CooldownTimer{
		progressBar: widget.NewProgressBar(),
		timeLabel:   widget.NewLabel("Ready"),
		isActive:    false,
	}

	// Style the progress bar to match existing UI
	timer.progressBar.TextFormatter = func() string {
		return "" // Hide default progress text
	}

	timer.ExtendBaseWidget(timer)
	return timer
}

// StartCooldown begins a cooldown period for the specified duration
// Uses goroutine-based updating following existing animation patterns
func (ct *CooldownTimer) StartCooldown(duration time.Duration) {
	ct.mu.Lock()
	ct.endTime = time.Now().Add(duration)
	ct.isActive = true
	ct.mu.Unlock()

	// Start update loop in background (follows existing timer patterns)
	go ct.updateLoop()
}

// SetOnComplete sets a callback function to execute when cooldown finishes
// Allows integration with gift button state management
func (ct *CooldownTimer) SetOnComplete(callback func()) {
	ct.mu.Lock()
	ct.onComplete = callback
	ct.mu.Unlock()
}

// IsActive returns whether the cooldown timer is currently running
// Thread-safe check for external state management
func (ct *CooldownTimer) IsActive() bool {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	return ct.isActive
}

// GetRemainingTime returns the remaining cooldown duration
// Thread-safe method for external cooldown queries
func (ct *CooldownTimer) GetRemainingTime() time.Duration {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	if !ct.isActive {
		return 0
	}

	remaining := time.Until(ct.endTime)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// updateLoop handles the timer update cycle
// Follows existing animation update patterns with 100ms intervals
func (ct *CooldownTimer) updateLoop() {
	ct.mu.RLock()
	startTime := time.Now()
	totalDuration := ct.endTime.Sub(startTime)
	ct.mu.RUnlock()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		ct.mu.RLock()
		if !ct.isActive {
			ct.mu.RUnlock()
			return
		}

		remaining := time.Until(ct.endTime)
		ct.mu.RUnlock()

		if remaining <= 0 {
			// Cooldown complete
			ct.mu.Lock()
			ct.isActive = false
			ct.progressBar.SetValue(1.0)
			ct.timeLabel.SetText("Ready")
			callback := ct.onComplete
			ct.mu.Unlock()

			ct.Refresh()

			// Execute completion callback if set
			if callback != nil {
				callback()
			}
			return
		}

		// Calculate progress (0.0 = just started, 1.0 = complete)
		progress := 1.0 - (remaining.Seconds() / totalDuration.Seconds())
		if progress < 0 {
			progress = 0
		} else if progress > 1 {
			progress = 1
		}

		ct.progressBar.SetValue(progress)
		ct.timeLabel.SetText(fmt.Sprintf("%ds", int(remaining.Seconds())+1))
		ct.Refresh()

		<-ticker.C
	}
}

// CreateRenderer creates the visual representation of the cooldown timer
// Uses container layout following existing widget patterns
func (ct *CooldownTimer) CreateRenderer() fyne.WidgetRenderer {
	content := container.NewVBox(
		ct.progressBar,
		ct.timeLabel,
	)

	return widget.NewSimpleRenderer(content)
}
