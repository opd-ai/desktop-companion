package ui

import (
	"github.com/opd-ai/desktop-companion/internal/character"
	"testing"
	"time"
)

// TestGiftDialog_CooldownIntegration tests the integration between cooldown timers and gift dialog
func TestGiftDialog_CooldownIntegration(t *testing.T) {
	// Create test character with gift system
	testCharacter := &character.CharacterCard{
		Name: "Test Character",
		GiftSystem: &character.GiftSystemConfig{
			Enabled: true,
		},
	}

	gameState := &character.GameState{
		GiftMemories: make([]character.GiftMemory, 0),
	}

	giftManager := character.NewGiftManager(testCharacter, gameState)

	// Create test gift with cooldown
	testGift := &character.GiftDefinition{
		ID:          "cooldown_test_gift",
		Name:        "Cooldown Test Gift",
		Description: "A gift for testing cooldown UI integration",
		Rarity:      "common",
		Properties: character.GiftProperties{
			CooldownSeconds: 1, // 1 second cooldown for fast testing
		},
		GiftEffects: character.GiftEffects{
			Immediate: character.ImmediateEffects{
				Stats:      map[string]float64{"happiness": 5},
				Responses:  []string{"Thank you for the gift!"},
				Animations: []string{"happy"},
			},
		},
		Notes: character.GiftNotesConfig{
			Enabled:     true,
			MaxLength:   100,
			Placeholder: "Add a note",
		},
	}

	// Add gift to manager's catalog using test helper method
	giftManager.AddGiftToTestCatalog(testGift)

	// Create gift dialog
	dialog := NewGiftSelectionDialog(giftManager)

	if dialog == nil {
		t.Fatal("Failed to create gift dialog")
	}

	if dialog.cooldownTimers == nil {
		t.Error("Dialog should have cooldown timers map initialized")
	}

	if len(dialog.cooldownTimers) != 0 {
		t.Error("New dialog should have empty cooldown timers map")
	}

	// Initially gift should not be on cooldown
	if giftManager.IsGiftOnCooldown(testGift.ID) {
		t.Error("Gift should not be on cooldown initially")
	}

	// Give the gift to trigger cooldown
	_, err := giftManager.GiveGift(testGift.ID, "Test gift")
	if err != nil {
		t.Fatalf("Failed to give gift: %v", err)
	}

	// Now gift should be on cooldown
	if !giftManager.IsGiftOnCooldown(testGift.ID) {
		t.Error("Gift should be on cooldown after giving")
	}

	remaining := giftManager.GetGiftCooldownRemaining(testGift.ID)
	if remaining <= 0 || remaining > 1*time.Second {
		t.Errorf("Expected cooldown remaining between 0 and 1s, got %v", remaining)
	}

	// Wait for cooldown to expire
	time.Sleep(1*time.Second + 100*time.Millisecond)

	// Gift should no longer be on cooldown
	if giftManager.IsGiftOnCooldown(testGift.ID) {
		t.Error("Gift should not be on cooldown after waiting")
	}
}

// TestGiftDialog_UpdateGiveButtonState tests button state updates with cooldowns
func TestGiftDialog_UpdateGiveButtonState(t *testing.T) {
	testCharacter := &character.CharacterCard{
		Name: "Test Character",
		GiftSystem: &character.GiftSystemConfig{
			Enabled: true,
		},
	}

	gameState := &character.GameState{
		GiftMemories: make([]character.GiftMemory, 0),
	}

	giftManager := character.NewGiftManager(testCharacter, gameState)

	testGift := &character.GiftDefinition{
		ID:          "button_test_gift",
		Name:        "Button Test Gift",
		Description: "A gift for testing button states",
		Rarity:      "common",
		Properties: character.GiftProperties{
			CooldownSeconds: 1,
		},
		GiftEffects: character.GiftEffects{
			Immediate: character.ImmediateEffects{
				Stats:     map[string]float64{"happiness": 1},
				Responses: []string{"Thanks!"},
			},
		},
		Notes: character.GiftNotesConfig{
			Enabled: true,
		},
	}

	giftManager.AddGiftToTestCatalog(testGift)

	dialog := NewGiftSelectionDialog(giftManager)

	// Select the gift (should enable button initially)
	dialog.selectedGift = testGift
	dialog.updateGiveButtonState()

	if dialog.giveButton.Disabled() {
		t.Error("Give button should be enabled for available gift")
	}

	// Give the gift to put it on cooldown
	_, err := giftManager.GiveGift(testGift.ID, "")
	if err != nil {
		t.Fatalf("Failed to give gift: %v", err)
	}

	// Update button state - should be disabled due to cooldown
	dialog.updateGiveButtonState()

	if !dialog.giveButton.Disabled() {
		t.Error("Give button should be disabled when gift is on cooldown")
	}

	// Wait for cooldown to expire
	time.Sleep(1*time.Second + 100*time.Millisecond)

	// Update button state - should be enabled again
	dialog.updateGiveButtonState()

	if dialog.giveButton.Disabled() {
		t.Error("Give button should be enabled after cooldown expires")
	}
}

// TestGiftDialog_HandleGiftCooldownSafety tests safety checks in gift giving
func TestGiftDialog_HandleGiftCooldownSafety(t *testing.T) {
	testCharacter := &character.CharacterCard{
		Name: "Test Character",
		GiftSystem: &character.GiftSystemConfig{
			Enabled: true,
		},
	}

	gameState := &character.GameState{
		GiftMemories: make([]character.GiftMemory, 0),
	}

	giftManager := character.NewGiftManager(testCharacter, gameState)

	testGift := &character.GiftDefinition{
		ID:          "safety_test_gift",
		Name:        "Safety Test Gift",
		Description: "A gift for testing safety checks",
		Rarity:      "common",
		Properties: character.GiftProperties{
			CooldownSeconds: 1,
		},
		GiftEffects: character.GiftEffects{
			Immediate: character.ImmediateEffects{
				Stats:     map[string]float64{"happiness": 1},
				Responses: []string{"Safe!"},
			},
		},
		Notes: character.GiftNotesConfig{
			Enabled: true,
		},
	}

	giftManager.AddGiftToTestCatalog(testGift)

	dialog := NewGiftSelectionDialog(giftManager)

	// Give the gift to put it on cooldown
	_, err := giftManager.GiveGift(testGift.ID, "")
	if err != nil {
		t.Fatalf("Failed to give gift: %v", err)
	}

	// Select the gift that's on cooldown
	dialog.selectedGift = testGift

	// Try to give gift while on cooldown - should be safe (no crash)
	originalMemoryCount := len(gameState.GiftMemories)
	dialog.handleGiveGift() // Should return early due to cooldown check

	// Should not have created additional memory entry
	if len(gameState.GiftMemories) != originalMemoryCount {
		t.Error("handleGiveGift should not give gifts while on cooldown")
	}
}
