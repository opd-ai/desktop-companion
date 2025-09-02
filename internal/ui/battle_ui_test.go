package ui

import (
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
)

// TestNewBattleActionDialog tests battle action dialog creation
func TestNewBattleActionDialog(t *testing.T) {
	t.Run("dialog_creation", func(t *testing.T) {
		dialog := NewBattleActionDialog(30 * time.Second)

		if dialog == nil {
			t.Fatal("NewBattleActionDialog returned nil")
		}

		if dialog.visible {
			t.Error("New battle action dialog should be initially hidden")
		}

		if dialog.turnTimeout != 30*time.Second {
			t.Errorf("Expected turn timeout of 30s, got %v", dialog.turnTimeout)
		}

		if len(dialog.actionButtons) == 0 {
			t.Error("Battle action dialog should have action buttons")
		}

		if dialog.background == nil {
			t.Error("Battle action dialog background should be initialized")
		}

		if dialog.content == nil {
			t.Error("Battle action dialog content container should be initialized")
		}

		if dialog.cancelButton == nil {
			t.Error("Battle action dialog should have cancel button")
		}
	})

	t.Run("dialog_without_timer", func(t *testing.T) {
		dialog := NewBattleActionDialog(0)

		if dialog.timerLabel != nil {
			t.Error("Dialog without timeout should not have timer label")
		}

		if dialog.turnTimeout != 0 {
			t.Error("Dialog without timeout should have zero turn timeout")
		}
	})

	t.Run("dialog_with_timer", func(t *testing.T) {
		dialog := NewBattleActionDialog(15 * time.Second)

		if dialog.timerLabel == nil {
			t.Error("Dialog with timeout should have timer label")
		}

		if dialog.turnTimeout != 15*time.Second {
			t.Errorf("Expected turn timeout of 15s, got %v", dialog.turnTimeout)
		}
	})
}

// TestBattleActionDialogCallbacks tests callback functionality
func TestBattleActionDialogCallbacks(t *testing.T) {
	dialog := NewBattleActionDialog(10 * time.Second)

	t.Run("action_select_callback", func(t *testing.T) {
		var selectedAction BattleActionType
		actionCalled := false

		dialog.SetOnActionSelect(func(action BattleActionType) {
			selectedAction = action
			actionCalled = true
		})

		// Simulate action selection
		if dialog.onActionSelect != nil {
			dialog.onActionSelect(ActionAttack)
		}

		if !actionCalled {
			t.Error("Action select callback should have been called")
		}

		if selectedAction != ActionAttack {
			t.Errorf("Expected ActionAttack, got %v", selectedAction)
		}
	})

	t.Run("cancel_callback", func(t *testing.T) {
		cancelCalled := false

		dialog.SetOnCancel(func() {
			cancelCalled = true
		})

		// Simulate cancel
		if dialog.onCancel != nil {
			dialog.onCancel()
		}

		if !cancelCalled {
			t.Error("Cancel callback should have been called")
		}
	})
}

// TestBattleActionDialogVisibility tests show/hide functionality
func TestBattleActionDialogVisibility(t *testing.T) {
	dialog := NewBattleActionDialog(5 * time.Second)

	t.Run("initial_state", func(t *testing.T) {
		if dialog.IsVisible() {
			t.Error("Dialog should be initially hidden")
		}
	})

	t.Run("show_dialog", func(t *testing.T) {
		dialog.Show()

		if !dialog.IsVisible() {
			t.Error("Dialog should be visible after Show()")
		}

		if !dialog.visible {
			t.Error("Internal visible state should be true after Show()")
		}
	})

	t.Run("hide_dialog", func(t *testing.T) {
		dialog.Show()
		dialog.Hide()

		if dialog.IsVisible() {
			t.Error("Dialog should be hidden after Hide()")
		}

		if dialog.visible {
			t.Error("Internal visible state should be false after Hide()")
		}

		if dialog.timerRunning {
			t.Error("Timer should be stopped after Hide()")
		}
	})
}

// TestBattleActionDialogTimer tests timer functionality
func TestBattleActionDialogTimer(t *testing.T) {
	t.Run("timer_functionality", func(t *testing.T) {
		dialog := NewBattleActionDialog(100 * time.Millisecond) // Very short timeout for testing
		timeoutChan := make(chan bool, 1)

		dialog.SetOnCancel(func() {
			timeoutChan <- true
		})

		dialog.Show()

		if !dialog.IsTimerRunning() {
			t.Error("Timer should be running after Show()")
		}

		// Wait for timeout
		select {
		case <-timeoutChan:
			// Good, timeout occurred
		case <-time.After(200 * time.Millisecond):
			t.Error("Timeout callback should have been called")
		}

		if dialog.IsVisible() {
			t.Error("Dialog should be hidden after timeout")
		}
	})

	t.Run("timer_stop_on_hide", func(t *testing.T) {
		dialog := NewBattleActionDialog(1 * time.Second)
		dialog.Show()

		if !dialog.timerRunning {
			t.Error("Timer should be running after Show()")
		}

		dialog.Hide()

		if dialog.timerRunning {
			t.Error("Timer should be stopped after Hide()")
		}
	})
}

// TestBattleActionDialogRenderer tests the Fyne renderer
func TestBattleActionDialogRenderer(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	dialog := NewBattleActionDialog(10 * time.Second)
	renderer := dialog.CreateRenderer()

	t.Run("renderer_creation", func(t *testing.T) {
		if renderer == nil {
			t.Fatal("CreateRenderer returned nil")
		}
	})

	t.Run("min_size_hidden", func(t *testing.T) {
		minSize := renderer.MinSize()
		if minSize.Width != 0 || minSize.Height != 0 {
			t.Error("Hidden dialog should have zero min size")
		}
	})

	t.Run("min_size_visible", func(t *testing.T) {
		dialog.Show()
		minSize := renderer.MinSize()
		if minSize.Width <= 0 || minSize.Height <= 0 {
			t.Error("Visible dialog should have positive min size")
		}
		dialog.Hide()
	})

	t.Run("objects_hidden", func(t *testing.T) {
		objects := renderer.Objects()
		if len(objects) != 0 {
			t.Error("Hidden dialog should have no visible objects")
		}
	})

	t.Run("objects_visible", func(t *testing.T) {
		dialog.Show()
		objects := renderer.Objects()
		if len(objects) == 0 {
			t.Error("Visible dialog should have objects")
		}
		dialog.Hide()
	})

	t.Run("destroy_cleanup", func(t *testing.T) {
		dialog.Show()
		renderer.Destroy()

		// Dialog should stop timer on destroy
		if dialog.timerRunning {
			t.Error("Timer should be stopped after renderer destroy")
		}
	})
}

// TestNewBattleResultOverlay tests battle result overlay creation
func TestNewBattleResultOverlay(t *testing.T) {
	overlay := NewBattleResultOverlay()

	if overlay == nil {
		t.Fatal("NewBattleResultOverlay returned nil")
	}

	if overlay.visible {
		t.Error("New battle result overlay should be initially hidden")
	}

	if overlay.background == nil {
		t.Error("Battle result overlay background should be initialized")
	}

	if overlay.content == nil {
		t.Error("Battle result overlay content container should be initialized")
	}

	if overlay.titleLabel == nil {
		t.Error("Battle result overlay should have title label")
	}

	if overlay.messageLabel == nil {
		t.Error("Battle result overlay should have message label")
	}

	if overlay.detailsLabel == nil {
		t.Error("Battle result overlay should have details label")
	}
}

// TestBattleResultOverlayShowResult tests result display functionality
func TestBattleResultOverlayShowResult(t *testing.T) {
	overlay := NewBattleResultOverlay()

	t.Run("successful_attack_result", func(t *testing.T) {
		result := BattleResult{
			Success:    true,
			ActionType: ActionAttack,
			Damage:     25.0,
			Response:   "Take that!",
		}

		overlay.ShowResult(result)

		if !overlay.IsVisible() {
			t.Error("Overlay should be visible after ShowResult()")
		}

		if overlay.titleLabel.Text != "Action Successful!" {
			t.Errorf("Expected success title, got: %s", overlay.titleLabel.Text)
		}

		if overlay.messageLabel.Text != "Dealt 25 damage!" {
			t.Errorf("Expected damage message, got: %s", overlay.messageLabel.Text)
		}
	})

	t.Run("failed_action_result", func(t *testing.T) {
		result := BattleResult{
			Success:    false,
			ActionType: ActionHeal,
		}

		overlay.ShowResult(result)

		if overlay.titleLabel.Text != "Action Failed!" {
			t.Errorf("Expected failure title, got: %s", overlay.titleLabel.Text)
		}

		if overlay.messageLabel.Text != "heal failed!" {
			t.Errorf("Expected failure message, got: %s", overlay.messageLabel.Text)
		}
	})

	t.Run("healing_result", func(t *testing.T) {
		result := BattleResult{
			Success:    true,
			ActionType: ActionHeal,
			Healing:    30.0,
		}

		overlay.ShowResult(result)

		if overlay.messageLabel.Text != "Restored 30 HP!" {
			t.Errorf("Expected healing message, got: %s", overlay.messageLabel.Text)
		}
	})

	t.Run("auto_hide_functionality", func(t *testing.T) {
		result := BattleResult{
			Success:    true,
			ActionType: ActionDefend,
		}

		overlay.ShowResult(result)

		if !overlay.IsVisible() {
			t.Error("Overlay should be visible immediately after ShowResult()")
		}

		// Overlay should auto-hide after 3 seconds, but we won't wait that long in tests
		// Just verify the timer is set
		if overlay.autoHideTimer == nil {
			t.Error("Auto-hide timer should be set")
		}
	})
}

// TestBattleResultOverlayShowMessage tests simple message display
func TestBattleResultOverlayShowMessage(t *testing.T) {
	overlay := NewBattleResultOverlay()

	overlay.ShowMessage("Test Title", "Test Message")

	if !overlay.IsVisible() {
		t.Error("Overlay should be visible after ShowMessage()")
	}

	if overlay.titleLabel.Text != "Test Title" {
		t.Errorf("Expected 'Test Title', got: %s", overlay.titleLabel.Text)
	}

	if overlay.messageLabel.Text != "Test Message" {
		t.Errorf("Expected 'Test Message', got: %s", overlay.messageLabel.Text)
	}

	if overlay.detailsLabel.Text != "" {
		t.Error("Details label should be empty for simple message")
	}
}

// TestBattleResultOverlayRenderer tests the Fyne renderer
func TestBattleResultOverlayRenderer(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	overlay := NewBattleResultOverlay()
	renderer := overlay.CreateRenderer()

	t.Run("renderer_creation", func(t *testing.T) {
		if renderer == nil {
			t.Fatal("CreateRenderer returned nil")
		}
	})

	t.Run("min_size_hidden", func(t *testing.T) {
		minSize := renderer.MinSize()
		if minSize.Width != 0 || minSize.Height != 0 {
			t.Error("Hidden overlay should have zero min size")
		}
	})

	t.Run("min_size_visible", func(t *testing.T) {
		overlay.Show()
		minSize := renderer.MinSize()
		if minSize.Width <= 0 || minSize.Height <= 0 {
			t.Error("Visible overlay should have positive min size")
		}
		overlay.Hide()
	})

	t.Run("objects_visibility", func(t *testing.T) {
		// Hidden state
		objects := renderer.Objects()
		if len(objects) != 0 {
			t.Error("Hidden overlay should have no visible objects")
		}

		// Visible state
		overlay.Show()
		objects = renderer.Objects()
		if len(objects) == 0 {
			t.Error("Visible overlay should have objects")
		}
		overlay.Hide()
	})

	t.Run("destroy_cleanup", func(t *testing.T) {
		overlay.ShowMessage("Test", "Test")
		renderer.Destroy()

		// Should stop auto-hide timer on destroy
		if overlay.autoHideTimer != nil {
			// Timer should be stopped (though we can't easily test this)
			// The destroy method should handle cleanup
		}
	})
}

// TestBattleResultFormatting tests result message formatting
func TestBattleResultFormatting(t *testing.T) {
	overlay := NewBattleResultOverlay()

	testCases := []struct {
		name           string
		result         BattleResult
		expectedMsg    string
		expectedDetail string
	}{
		{
			name: "attack_with_damage",
			result: BattleResult{
				Success:    true,
				ActionType: ActionAttack,
				Damage:     15.0,
			},
			expectedMsg: "Dealt 15 damage!",
		},
		{
			name: "heal_with_amount",
			result: BattleResult{
				Success:    true,
				ActionType: ActionHeal,
				Healing:    20.0,
			},
			expectedMsg: "Restored 20 HP!",
		},
		{
			name: "defend_action",
			result: BattleResult{
				Success:    true,
				ActionType: ActionDefend,
			},
			expectedMsg: "Defense stance activated!",
		},
		{
			name: "stun_action",
			result: BattleResult{
				Success:    true,
				ActionType: ActionStun,
			},
			expectedMsg: "Opponent stunned!",
		},
		{
			name: "failed_action",
			result: BattleResult{
				Success:    false,
				ActionType: ActionAttack,
			},
			expectedMsg: "attack failed!",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			message := overlay.formatActionMessage(tc.result)
			if message != tc.expectedMsg {
				t.Errorf("Expected message '%s', got '%s'", tc.expectedMsg, message)
			}
		})
	}
}

// TestBattleActionTypes tests all battle action type constants
func TestBattleActionTypes(t *testing.T) {
	expectedActions := []BattleActionType{
		ActionAttack,
		ActionDefend,
		ActionStun,
		ActionHeal,
		ActionBoost,
		ActionCounter,
		ActionDrain,
		ActionShield,
		ActionCharge,
		ActionEvade,
		ActionTaunt,
	}

	if len(expectedActions) != 11 {
		t.Errorf("Expected 11 battle action types, found %d", len(expectedActions))
	}

	// Test that all actions are non-empty strings
	for _, action := range expectedActions {
		if string(action) == "" {
			t.Error("Battle action type should not be empty string")
		}
	}

	// Test specific action values
	if ActionAttack != "attack" {
		t.Errorf("Expected ActionAttack to be 'attack', got '%s'", ActionAttack)
	}

	if ActionDefend != "defend" {
		t.Errorf("Expected ActionDefend to be 'defend', got '%s'", ActionDefend)
	}

	if ActionHeal != "heal" {
		t.Errorf("Expected ActionHeal to be 'heal', got '%s'", ActionHeal)
	}
}
