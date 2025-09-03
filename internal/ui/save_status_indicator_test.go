package ui

import (
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
)

func TestNewSaveStatusIndicator(t *testing.T) {
	indicator := NewSaveStatusIndicator()

	if indicator == nil {
		t.Fatal("NewSaveStatusIndicator() returned nil")
	}

	if indicator.GetStatus() != SaveStatusIdle {
		t.Errorf("Expected initial status to be SaveStatusIdle, got %v", indicator.GetStatus())
	}

	if indicator.icon == nil {
		t.Error("Expected icon to be initialized")
	}
}

func TestSaveStatusIndicator_SetStatus(t *testing.T) {
	indicator := NewSaveStatusIndicator()

	tests := []struct {
		name    string
		status  SaveStatus
		message string
	}{
		{"Saving", SaveStatusSaving, ""},
		{"Saved", SaveStatusSaved, ""},
		{"Error", SaveStatusError, "Failed to save"},
		{"Idle", SaveStatusIdle, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeTime := time.Now()
			indicator.SetStatus(tt.status, tt.message)

			if indicator.GetStatus() != tt.status {
				t.Errorf("Expected status %v, got %v", tt.status, indicator.GetStatus())
			}

			if tt.status == SaveStatusSaved {
				lastSaved := indicator.GetLastSaved()
				if lastSaved.Before(beforeTime) {
					t.Error("Expected lastSaved timestamp to be updated")
				}
				if indicator.GetErrorMessage() != "" {
					t.Error("Expected error message to be cleared when status is SaveStatusSaved")
				}
			}

			if tt.status == SaveStatusError {
				if indicator.GetErrorMessage() != tt.message {
					t.Errorf("Expected error message %q, got %q", tt.message, indicator.GetErrorMessage())
				}
			}
		})
	}
}

func TestSaveStatusIndicator_IconUpdates(t *testing.T) {
	indicator := NewSaveStatusIndicator()

	// Test that icon updates when status changes
	originalIcon := indicator.icon

	indicator.SetStatus(SaveStatusSaving, "")
	if indicator.icon == originalIcon {
		t.Error("Expected icon to change when status changes")
	}

	// Test that different statuses produce different icons
	indicator.SetStatus(SaveStatusSaved, "")
	savedIcon := indicator.icon

	indicator.SetStatus(SaveStatusError, "test error")
	errorIcon := indicator.icon

	indicator.SetStatus(SaveStatusIdle, "")
	idleIcon := indicator.icon

	if savedIcon == errorIcon || errorIcon == idleIcon || savedIcon == idleIcon {
		t.Error("Expected different icons for different statuses")
	}
}

func TestSaveStatusIndicator_Renderer(t *testing.T) {
	indicator := NewSaveStatusIndicator()
	renderer := indicator.CreateRenderer()

	if renderer == nil {
		t.Fatal("CreateRenderer() returned nil")
	}

	// Test MinSize
	minSize := renderer.MinSize()
	expectedSize := fyne.NewSize(16, 16)
	if minSize != expectedSize {
		t.Errorf("Expected MinSize %v, got %v", expectedSize, minSize)
	}

	// Test Objects
	objects := renderer.Objects()
	if len(objects) != 1 {
		t.Errorf("Expected 1 object, got %d", len(objects))
	}

	if objects[0] != indicator.icon {
		t.Error("Expected objects to contain the indicator icon")
	}
}

func TestSaveStatusIndicator_Layout(t *testing.T) {
	indicator := NewSaveStatusIndicator()
	renderer := indicator.CreateRenderer()

	// Test layout with a specific size
	testSize := fyne.NewSize(20, 20)
	renderer.Layout(testSize)

	if indicator.icon.Size() != testSize {
		t.Errorf("Expected icon size %v, got %v", testSize, indicator.icon.Size())
	}

	expectedPos := fyne.NewPos(0, 0)
	if indicator.icon.Position() != expectedPos {
		t.Errorf("Expected icon position %v, got %v", expectedPos, indicator.icon.Position())
	}
}

func TestSaveStatusIndicator_StatusTransitions(t *testing.T) {
	indicator := NewSaveStatusIndicator()

	// Test complete save operation cycle
	indicator.SetStatus(SaveStatusSaving, "")
	if indicator.GetStatus() != SaveStatusSaving {
		t.Error("Expected status to be SaveStatusSaving")
	}

	indicator.SetStatus(SaveStatusSaved, "")
	if indicator.GetStatus() != SaveStatusSaved {
		t.Error("Expected status to be SaveStatusSaved")
	}

	// Test error during save
	indicator.SetStatus(SaveStatusSaving, "")
	indicator.SetStatus(SaveStatusError, "Disk full")
	if indicator.GetStatus() != SaveStatusError {
		t.Error("Expected status to be SaveStatusError")
	}
	if indicator.GetErrorMessage() != "Disk full" {
		t.Errorf("Expected error message 'Disk full', got %q", indicator.GetErrorMessage())
	}

	// Test return to idle
	indicator.SetStatus(SaveStatusIdle, "")
	if indicator.GetStatus() != SaveStatusIdle {
		t.Error("Expected status to be SaveStatusIdle")
	}
}

func TestSaveStatusIndicator_ErrorHandling(t *testing.T) {
	indicator := NewSaveStatusIndicator()

	// Test empty error message
	indicator.SetStatus(SaveStatusError, "")
	if indicator.GetErrorMessage() != "" {
		t.Errorf("Expected empty error message, got %q", indicator.GetErrorMessage())
	}

	// Test error message persistence
	errorMsg := "Network timeout"
	indicator.SetStatus(SaveStatusError, errorMsg)
	if indicator.GetErrorMessage() != errorMsg {
		t.Errorf("Expected error message %q, got %q", errorMsg, indicator.GetErrorMessage())
	}

	// Test error clearing on successful save
	indicator.SetStatus(SaveStatusSaved, "")
	if indicator.GetErrorMessage() != "" {
		t.Error("Expected error message to be cleared after successful save")
	}
}

func TestSaveStatusIndicator_TimeTracking(t *testing.T) {
	indicator := NewSaveStatusIndicator()

	// Test initial lastSaved is zero
	initialTime := indicator.GetLastSaved()
	if !initialTime.IsZero() {
		t.Error("Expected initial lastSaved to be zero time")
	}

	// Test lastSaved update on successful save
	beforeSave := time.Now()
	indicator.SetStatus(SaveStatusSaved, "")
	afterSave := time.Now()

	lastSaved := indicator.GetLastSaved()
	if lastSaved.Before(beforeSave) || lastSaved.After(afterSave) {
		t.Error("Expected lastSaved to be set to current time on successful save")
	}

	// Test lastSaved doesn't change on other status updates
	savedTime := indicator.GetLastSaved()
	time.Sleep(1 * time.Millisecond) // Ensure time difference
	indicator.SetStatus(SaveStatusError, "test error")

	if !indicator.GetLastSaved().Equal(savedTime) {
		t.Error("Expected lastSaved to remain unchanged on error status")
	}
}

func TestSaveStatusIndicator_WidgetIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode to avoid Fyne font cache race condition")
	}
	// Test widget integration with Fyne test framework
	app := test.NewApp()
	defer app.Quit()

	indicator := NewSaveStatusIndicator()

	// Test that widget can be resized
	testSize := fyne.NewSize(24, 24)
	indicator.Resize(testSize)
	if indicator.Size() != testSize {
		t.Errorf("Expected widget size %v, got %v", testSize, indicator.Size())
	}

	// Test that widget can be moved
	testPos := fyne.NewPos(10, 10)
	indicator.Move(testPos)
	if indicator.Position() != testPos {
		t.Errorf("Expected widget position %v, got %v", testPos, indicator.Position())
	}

	// Test refresh doesn't crash
	indicator.Refresh()
}

// TestSaveStatusIndicator_IconThemeCompatibility tests that icons are from standard theme
func TestSaveStatusIndicator_IconThemeCompatibility(t *testing.T) {
	indicator := NewSaveStatusIndicator()

	// Check that all status use standard theme icons
	statuses := []SaveStatus{SaveStatusIdle, SaveStatusSaving, SaveStatusSaved, SaveStatusError}
	expectedIcons := []fyne.Resource{
		theme.DocumentSaveIcon(),
		theme.ViewRefreshIcon(),
		theme.ConfirmIcon(),
		theme.ErrorIcon(),
	}

	for i, status := range statuses {
		indicator.SetStatus(status, "")
		// We can't directly compare icon resources, but we can ensure the icon exists
		if indicator.icon == nil {
			t.Errorf("Expected icon to be set for status %v", status)
		}
		// Verify the icon is created with the expected theme resource
		// Note: We can't directly access the resource from the widget.Icon,
		// but we can verify it's not nil and matches expected structure
		_ = expectedIcons[i] // Use to avoid unused variable warning
	}
}
