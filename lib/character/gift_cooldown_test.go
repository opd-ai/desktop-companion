package character

import (
	"testing"
	"time"
)

// TestGiftManager_CooldownFunctionality tests the complete cooldown system
func TestGiftManager_CooldownFunctionality(t *testing.T) {
	// Create test character and game state
	character := &CharacterCard{
		Name: "Test Character",
		GiftSystem: &GiftSystemConfig{
			Enabled: true,
		},
	}

	gameState := &GameState{
		GiftMemories: make([]GiftMemory, 0),
	}

	manager := NewGiftManager(character, gameState)

	// Create test gift with cooldown
	testGift := &GiftDefinition{
		ID:          "test_gift",
		Name:        "Test Gift",
		Description: "A test gift with cooldown",
		Properties: GiftProperties{
			CooldownSeconds: 2, // 2 second cooldown
		},
		GiftEffects: GiftEffects{
			Immediate: ImmediateEffects{
				Stats:      map[string]float64{"happiness": 10},
				Responses:  []string{"Thank you!"},
				Animations: []string{"happy"},
			},
			Memory: MemoryEffects{
				Importance:    0.5,
				EmotionalTone: "positive",
			},
		},
	}

	// Add gift to catalog
	manager.giftCatalog[testGift.ID] = testGift

	// Initially gift should not be on cooldown
	if manager.IsGiftOnCooldown(testGift.ID) {
		t.Error("Gift should not be on cooldown initially")
	}

	if manager.GetGiftCooldownRemaining(testGift.ID) != 0 {
		t.Error("Gift should have 0 cooldown remaining initially")
	}

	// Give the gift
	_, err := manager.GiveGift(testGift.ID, "Test note")
	if err != nil {
		t.Fatalf("Failed to give gift: %v", err)
	}

	// Now gift should be on cooldown
	if !manager.IsGiftOnCooldown(testGift.ID) {
		t.Error("Gift should be on cooldown after giving")
	}

	remaining := manager.GetGiftCooldownRemaining(testGift.ID)
	if remaining <= 0 || remaining > 2*time.Second {
		t.Errorf("Expected cooldown remaining between 0 and 2s, got %v", remaining)
	}

	// Gift should not be giveable while on cooldown
	if manager.canGiveGift(testGift) {
		t.Error("Gift should not be giveable while on cooldown")
	}

	// Wait for cooldown to expire
	time.Sleep(2*time.Second + 100*time.Millisecond)

	// Gift should no longer be on cooldown
	if manager.IsGiftOnCooldown(testGift.ID) {
		t.Error("Gift should not be on cooldown after waiting")
	}

	if manager.GetGiftCooldownRemaining(testGift.ID) != 0 {
		t.Error("Gift should have 0 cooldown remaining after waiting")
	}

	// Gift should be giveable again
	if !manager.canGiveGift(testGift) {
		t.Error("Gift should be giveable again after cooldown expires")
	}
}

// TestGiftManager_NoCooldownGift tests gifts without cooldown
func TestGiftManager_NoCooldownGift(t *testing.T) {
	character := &CharacterCard{
		Name: "Test Character",
		GiftSystem: &GiftSystemConfig{
			Enabled: true,
		},
	}

	gameState := &GameState{
		GiftMemories: make([]GiftMemory, 0),
	}

	manager := NewGiftManager(character, gameState)

	// Create test gift without cooldown
	testGift := &GiftDefinition{
		ID:          "no_cooldown_gift",
		Name:        "No Cooldown Gift",
		Description: "A gift without cooldown",
		Properties: GiftProperties{
			CooldownSeconds: 0, // No cooldown
		},
		GiftEffects: GiftEffects{
			Immediate: ImmediateEffects{
				Stats:      map[string]float64{"happiness": 5},
				Responses:  []string{"Thanks!"},
				Animations: []string{"happy"},
			},
		},
	}

	manager.giftCatalog[testGift.ID] = testGift

	// Give the gift multiple times
	for i := 0; i < 3; i++ {
		if manager.IsGiftOnCooldown(testGift.ID) {
			t.Errorf("Gift without cooldown should never be on cooldown (iteration %d)", i)
		}

		if !manager.canGiveGift(testGift) {
			t.Errorf("Gift without cooldown should always be giveable (iteration %d)", i)
		}

		_, err := manager.GiveGift(testGift.ID, "")
		if err != nil {
			t.Fatalf("Failed to give gift (iteration %d): %v", i, err)
		}
	}
}

// TestGiftManager_NonexistentGift tests cooldown check for nonexistent gifts
func TestGiftManager_NonexistentGift(t *testing.T) {
	character := &CharacterCard{Name: "Test Character"}
	gameState := &GameState{}
	manager := NewGiftManager(character, gameState)

	// Check cooldown for nonexistent gift
	if manager.IsGiftOnCooldown("nonexistent") {
		t.Error("Nonexistent gift should not be on cooldown")
	}

	if manager.GetGiftCooldownRemaining("nonexistent") != 0 {
		t.Error("Nonexistent gift should have 0 cooldown remaining")
	}
}

// TestGiftManager_NilGameState tests cooldown handling without game state
func TestGiftManager_NilGameState(t *testing.T) {
	character := &CharacterCard{Name: "Test Character"}
	manager := NewGiftManager(character, nil) // No game state

	// Check cooldown without game state
	if manager.IsGiftOnCooldown("any_gift") {
		t.Error("Gift should not be on cooldown without game state")
	}

	if manager.GetGiftCooldownRemaining("any_gift") != 0 {
		t.Error("Gift should have 0 cooldown remaining without game state")
	}
}

// TestGiftManager_CooldownAfterMultipleGifts tests cooldown tracking with multiple gift types
func TestGiftManager_CooldownAfterMultipleGifts(t *testing.T) {
	character := &CharacterCard{
		Name: "Test Character",
		GiftSystem: &GiftSystemConfig{
			Enabled: true,
		},
	}

	gameState := &GameState{
		GiftMemories: make([]GiftMemory, 0),
	}

	manager := NewGiftManager(character, gameState)

	// Create multiple gifts with different cooldowns
	gift1 := &GiftDefinition{
		ID:          "gift1",
		Name:        "Gift 1",
		Description: "First gift",
		Properties: GiftProperties{
			CooldownSeconds: 1,
		},
		GiftEffects: GiftEffects{
			Immediate: ImmediateEffects{
				Stats:     map[string]float64{"happiness": 5},
				Responses: []string{"Thanks for gift 1!"},
			},
		},
	}

	gift2 := &GiftDefinition{
		ID:          "gift2",
		Name:        "Gift 2",
		Description: "Second gift",
		Properties: GiftProperties{
			CooldownSeconds: 2,
		},
		GiftEffects: GiftEffects{
			Immediate: ImmediateEffects{
				Stats:     map[string]float64{"happiness": 3},
				Responses: []string{"Thanks for gift 2!"},
			},
		},
	}

	manager.giftCatalog[gift1.ID] = gift1
	manager.giftCatalog[gift2.ID] = gift2

	// Give both gifts
	_, err := manager.GiveGift(gift1.ID, "")
	if err != nil {
		t.Fatalf("Failed to give gift1: %v", err)
	}

	_, err = manager.GiveGift(gift2.ID, "")
	if err != nil {
		t.Fatalf("Failed to give gift2: %v", err)
	}

	// Both should be on cooldown
	if !manager.IsGiftOnCooldown(gift1.ID) {
		t.Error("Gift1 should be on cooldown")
	}

	if !manager.IsGiftOnCooldown(gift2.ID) {
		t.Error("Gift2 should be on cooldown")
	}

	// Wait for gift1 cooldown to expire
	time.Sleep(1*time.Second + 100*time.Millisecond)

	// Gift1 should no longer be on cooldown, but gift2 should still be
	if manager.IsGiftOnCooldown(gift1.ID) {
		t.Error("Gift1 should not be on cooldown after 1 second")
	}

	if !manager.IsGiftOnCooldown(gift2.ID) {
		t.Error("Gift2 should still be on cooldown after 1 second")
	}

	// Wait for gift2 cooldown to expire
	time.Sleep(1*time.Second + 100*time.Millisecond)

	// Both gifts should no longer be on cooldown
	if manager.IsGiftOnCooldown(gift1.ID) {
		t.Error("Gift1 should not be on cooldown after 2 seconds")
	}

	if manager.IsGiftOnCooldown(gift2.ID) {
		t.Error("Gift2 should not be on cooldown after 2 seconds")
	}
}

// TestGiftManager_CooldownThreadSafety tests concurrent access to cooldown methods
func TestGiftManager_CooldownThreadSafety(t *testing.T) {
	character := &CharacterCard{
		Name: "Test Character",
		GiftSystem: &GiftSystemConfig{
			Enabled: true,
		},
	}

	gameState := &GameState{
		GiftMemories: make([]GiftMemory, 0),
	}

	manager := NewGiftManager(character, gameState)

	testGift := &GiftDefinition{
		ID:          "thread_test_gift",
		Name:        "Thread Test Gift",
		Description: "A gift for testing thread safety",
		Properties: GiftProperties{
			CooldownSeconds: 1,
		},
		GiftEffects: GiftEffects{
			Immediate: ImmediateEffects{
				Stats:     map[string]float64{"happiness": 1},
				Responses: []string{"Thread safe!"},
			},
		},
	}

	manager.giftCatalog[testGift.ID] = testGift

	// Give the gift to put it on cooldown
	_, err := manager.GiveGift(testGift.ID, "")
	if err != nil {
		t.Fatalf("Failed to give gift: %v", err)
	}

	// Start multiple goroutines checking cooldown status
	done := make(chan bool, 5)

	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 20; j++ {
				manager.IsGiftOnCooldown(testGift.ID)
				manager.GetGiftCooldownRemaining(testGift.ID)
				time.Sleep(10 * time.Millisecond)
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 5; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for goroutines to complete")
		}
	}
}
