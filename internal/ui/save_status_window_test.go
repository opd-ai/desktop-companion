package ui

import (
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"desktop-companion/internal/character"
)

func TestDesktopWindow_SaveStatusIndicatorIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode to avoid Fyne font cache race condition")
	}
	app := test.NewApp()
	defer app.Quit()

	// Create test character
	char := createTestCharacter()

	// Create window with save status indicator
	dw := NewDesktopWindow(app, char, false, nil, false, false, nil, false, false, false)

	// Verify save status indicator is created
	if dw.saveStatusIndicator == nil {
		t.Error("Expected save status indicator to be created")
	}

	// Verify indicator is positioned
	expectedSize := fyne.NewSize(16, 16)
	if dw.saveStatusIndicator.Size() != expectedSize {
		t.Errorf("Expected save status indicator size %v, got %v", expectedSize, dw.saveStatusIndicator.Size())
	}

	expectedPos := fyne.NewPos(float32(char.GetSize()-20), 4)
	if dw.saveStatusIndicator.Position() != expectedPos {
		t.Errorf("Expected save status indicator position %v, got %v", expectedPos, dw.saveStatusIndicator.Position())
	}
}

func TestDesktopWindow_SaveStatusCallbacks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode to avoid Fyne font cache race condition")
	}

	// Create test app
	app := test.NewApp()
	defer app.Quit()

	char := createTestCharacter()

	// Create desktop window
	dw := NewDesktopWindow(app, char, false, nil, false, false, nil, false, false, false)

	// Test that the window was created successfully
	if dw == nil {
		t.Error("Desktop window creation failed")
	}
}

func TestDesktopWindow_OnSaveCompleted_AutoReturnToIdle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode to avoid Fyne font cache race condition")
	}

	app := test.NewApp()
	defer app.Quit()

	char := createTestCharacter()
	dw := NewDesktopWindow(app, char, false, nil, false, false, nil, false, false, false)

	// Call onSaveCompleted
	dw.onSaveCompleted()

	// Verify status is initially saved
	if dw.saveStatusIndicator.GetStatus() != SaveStatusSaved {
		t.Errorf("Expected status SaveStatusSaved immediately after onSaveCompleted, got %v", dw.saveStatusIndicator.GetStatus())
	}

	// Wait for auto-return to idle (2+ seconds)
	time.Sleep(2100 * time.Millisecond)

	// Verify status returned to idle
	if dw.saveStatusIndicator.GetStatus() != SaveStatusIdle {
		t.Errorf("Expected status to return to SaveStatusIdle after 2 seconds, got %v", dw.saveStatusIndicator.GetStatus())
	}
}

func TestDesktopWindow_SaveStatusIndicator_NilSafety(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode to avoid Fyne font cache race condition")
	}

	app := test.NewApp()
	defer app.Quit()

	char := createTestCharacter()
	dw := NewDesktopWindow(app, char, false, nil, false, false, nil, false, false, false)

	// Set indicator to nil to test nil safety
	dw.saveStatusIndicator = nil

	// These should not crash
	dw.onSaveStarted()
	dw.onSaveCompleted()
	dw.onSaveError(nil)

	callback := dw.SetSaveStatusCallback()
	callback(SaveStatusSaving, "")
	callback(SaveStatusSaved, "")
	callback(SaveStatusError, "test")
	callback(SaveStatusIdle, "")
}

func TestDesktopWindow_SaveStatusIndicator_WindowContent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode to avoid Fyne font cache race condition")
	}

	app := test.NewApp()
	defer app.Quit()

	char := createTestCharacter()
	dw := NewDesktopWindow(app, char, false, nil, false, false, nil, false, false, false)

	// Verify window content includes save status indicator
	content := dw.window.Content()
	if content == nil {
		t.Fatal("Expected window to have content")
	}

	// Check that save status indicator is in the content
	// Note: Due to Fyne's container structure, we can't easily traverse objects
	// but we can verify the indicator exists and is properly positioned
	if dw.saveStatusIndicator == nil {
		t.Error("Expected save status indicator to exist in window")
	}

	// Verify indicator has the expected size and position
	expectedSize := fyne.NewSize(16, 16)
	if dw.saveStatusIndicator.Size() != expectedSize {
		t.Errorf("Expected size %v, got %v", expectedSize, dw.saveStatusIndicator.Size())
	}
}

func TestDesktopWindow_SaveStatusIndicator_DraggableCharacter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode to avoid Fyne font cache race condition")
	}

	app := test.NewApp()
	defer app.Quit()

	// Create character with movement enabled (draggable)
	char := createTestCharacter()
	char.GetCard().Behavior.MovementEnabled = true

	dw := NewDesktopWindow(app, char, false, nil, false, false, nil, false, false, false)

	// Verify save status indicator is still created for draggable characters
	if dw.saveStatusIndicator == nil {
		t.Error("Expected save status indicator to be created for draggable characters")
	}

	// Verify positioning is applied
	expectedPos := fyne.NewPos(float32(char.GetSize()-20), 4)
	if dw.saveStatusIndicator.Position() != expectedPos {
		t.Errorf("Expected position %v, got %v", expectedPos, dw.saveStatusIndicator.Position())
	}
}

func TestDesktopWindow_SaveStatusIndicator_GameModeIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode to avoid Fyne font cache race condition")
	}

	app := test.NewApp()
	defer app.Quit()

	char := createTestCharacter()

	// Create window in game mode
	dw := NewDesktopWindow(app, char, false, nil, true, false, nil, false, false, false)

	// Verify save status indicator exists alongside game features
	if dw.saveStatusIndicator == nil {
		t.Error("Expected save status indicator to exist in game mode")
	}

	// Verify it doesn't interfere with other game features
	// Note: statsOverlay may not be initialized in test mode
	if dw.statsOverlay != nil {
		// Both should be positioned without overlap
		indicatorPos := dw.saveStatusIndicator.Position()
		overlayContainer := dw.statsOverlay.GetContainer()

		// Verify they don't overlap (save status indicator should be separate)
		if overlayContainer != nil {
			// Save status indicator should be in top-right corner
			expectedIndicatorPos := fyne.NewPos(float32(char.GetSize()-20), 4)
			if indicatorPos != expectedIndicatorPos {
				t.Errorf("Expected save status indicator at %v, got %v", expectedIndicatorPos, indicatorPos)
			}
		}
	} else {
		// Just verify save status indicator positioning when overlay isn't available
		indicatorPos := dw.saveStatusIndicator.Position()
		expectedIndicatorPos := fyne.NewPos(float32(char.GetSize()-20), 4)
		if indicatorPos != expectedIndicatorPos {
			t.Errorf("Expected save status indicator at %v, got %v", expectedIndicatorPos, indicatorPos)
		}
	}

}

func TestDesktopWindow_SaveStatusIndicator_ThreadSafety(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode to avoid Fyne font cache race condition")
	}

	app := test.NewApp()
	defer app.Quit()

	char := createTestCharacter()
	dw := NewDesktopWindow(app, char, false, nil, false, false, nil, false, false, false)

	callback := dw.SetSaveStatusCallback()

	// Call callbacks from multiple goroutines
	done := make(chan bool, 4)

	go func() {
		callback(SaveStatusSaving, "")
		done <- true
	}()

	go func() {
		callback(SaveStatusSaved, "")
		done <- true
	}()

	go func() {
		callback(SaveStatusError, "error1")
		done <- true
	}()

	go func() {
		callback(SaveStatusIdle, "")
		done <- true
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 4; i++ {
		<-done
	}

	// Should not crash and indicator should have some final status
	if dw.saveStatusIndicator == nil {
		t.Error("Expected save status indicator to remain non-nil")
	}
}

// createTestCharacter creates a basic test character for window testing
func createTestCharacter() *character.Character {
	card := &character.CharacterCard{
		Name:        "Test Character",
		Description: "Test character for save status indicator",
		Animations:  map[string]string{"idle": "test.gif"},
		Behavior: character.Behavior{
			IdleTimeout:     10,
			MovementEnabled: false,
			DefaultSize:     128,
		},
	}

	char, _ := character.New(card, "test_data")
	return char
}
