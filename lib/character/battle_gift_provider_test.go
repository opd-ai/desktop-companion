package character

import (
	"github.com/opd-ai/desktop-companion/internal/battle"
	"testing"
)

func TestBattleGiftProvider_Integration(t *testing.T) {
	// Create a character card and game state for testing
	character := &CharacterCard{
		Name: "Test Character",
		Stats: map[string]StatConfig{
			"health": {Initial: 100, Max: 100},
		},
	}

	gameState := &GameState{
		Stats: map[string]*Stat{
			"health": {Current: 100, Max: 100},
		},
	}

	// Create gift manager
	giftManager := NewGiftManager(character, gameState)

	// Create battle gift provider
	battleProvider := NewBattleGiftProvider(giftManager)

	// Test with empty catalog
	gifts := battleProvider.GetAvailableGifts()
	if len(gifts) != 0 {
		t.Errorf("Expected 0 gifts in empty catalog, got %d", len(gifts))
	}

	// Test getting non-existent gift
	_, err := battleProvider.GetGiftDefinition("non_existent")
	if err == nil {
		t.Error("Expected error for non-existent gift")
	}
}

func TestBattleGiftProvider_GiftConversion(t *testing.T) {
	// Create test gift definition with battle effects
	testGift := &GiftDefinition{
		ID:   "test_battle_item",
		Name: "Test Battle Item",
		GiftEffects: GiftEffects{
			Battle: BattleItemEffect{
				ActionType:     "attack",
				DamageModifier: 1.15,
				Duration:       2,
				Consumable:     true,
			},
		},
	}

	// Create a gift manager with the test gift
	character := &CharacterCard{
		Name: "Test Character",
		Stats: map[string]StatConfig{
			"health": {Initial: 100, Max: 100},
		},
	}

	gameState := &GameState{
		Stats: map[string]*Stat{
			"health": {Current: 100, Max: 100},
		},
	}

	giftManager := NewGiftManager(character, gameState)

	// Manually add the gift to the catalog for testing
	giftManager.giftCatalog["test_battle_item"] = testGift

	// Create battle gift provider
	battleProvider := NewBattleGiftProvider(giftManager)

	// Test getting the gift
	battleGift, err := battleProvider.GetGiftDefinition("test_battle_item")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify conversion
	if battleGift.ID != testGift.ID {
		t.Errorf("Expected ID %s, got %s", testGift.ID, battleGift.ID)
	}

	if battleGift.Name != testGift.Name {
		t.Errorf("Expected name %s, got %s", testGift.Name, battleGift.Name)
	}

	if battleGift.BattleEffect.ActionType != testGift.GiftEffects.Battle.ActionType {
		t.Errorf("Expected action type %s, got %s",
			testGift.GiftEffects.Battle.ActionType,
			battleGift.BattleEffect.ActionType)
	}

	if battleGift.BattleEffect.DamageModifier != testGift.GiftEffects.Battle.DamageModifier {
		t.Errorf("Expected damage modifier %f, got %f",
			testGift.GiftEffects.Battle.DamageModifier,
			battleGift.BattleEffect.DamageModifier)
	}
}

func TestBattleGiftProvider_HasBattleEffects(t *testing.T) {
	giftManager := NewGiftManager(&CharacterCard{}, &GameState{})
	battleProvider := NewBattleGiftProvider(giftManager)

	// Test gift with battle effects
	giftWithEffects := &GiftDefinition{
		ID: "battle_gift",
		GiftEffects: GiftEffects{
			Battle: BattleItemEffect{
				DamageModifier: 1.10,
			},
		},
	}

	if !battleProvider.hasBattleEffects(giftWithEffects) {
		t.Error("Expected gift to have battle effects")
	}

	// Test gift without battle effects
	giftWithoutEffects := &GiftDefinition{
		ID: "normal_gift",
		GiftEffects: GiftEffects{
			Immediate: ImmediateEffects{
				Stats: map[string]float64{"health": 10},
			},
		},
	}

	if battleProvider.hasBattleEffects(giftWithoutEffects) {
		t.Error("Expected gift to not have battle effects")
	}
}

// Test that the provider properly implements the battle.GiftProvider interface
func TestBattleGiftProvider_InterfaceCompliance(t *testing.T) {
	giftManager := NewGiftManager(&CharacterCard{}, &GameState{})
	battleProvider := NewBattleGiftProvider(giftManager)

	// This should compile without errors if the interface is properly implemented
	var _ battle.GiftProvider = battleProvider
}
