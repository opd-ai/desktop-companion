package battle

import (
	"errors"
	"testing"
)

// MockGiftProvider implements GiftProvider interface for testing
type MockGiftProvider struct {
	gifts map[string]GiftDefinition
}

func NewMockGiftProvider() *MockGiftProvider {
	return &MockGiftProvider{
		gifts: make(map[string]GiftDefinition),
	}
}

func (mgp *MockGiftProvider) AddGift(gift GiftDefinition) {
	mgp.gifts[gift.ID] = gift
}

func (mgp *MockGiftProvider) GetGiftDefinition(giftID string) (GiftDefinition, error) {
	gift, exists := mgp.gifts[giftID]
	if !exists {
		return GiftDefinition{}, errors.New("gift not found")
	}
	return gift, nil
}

func (mgp *MockGiftProvider) GetAvailableGifts() []GiftDefinition {
	gifts := make([]GiftDefinition, 0, len(mgp.gifts))
	for _, gift := range mgp.gifts {
		gifts = append(gifts, gift)
	}
	return gifts
}

// Test gift definitions for various scenarios
func createTestGifts() map[string]GiftDefinition {
	return map[string]GiftDefinition{
		"damage_potion": {
			ID:   "damage_potion",
			Name: "Damage Potion",
			BattleEffect: BattleItemEffect{
				ActionType:     "attack",
				DamageModifier: 1.15, // +15% damage
				Consumable:     true,
			},
		},
		"heal_herb": {
			ID:   "heal_herb",
			Name: "Healing Herb",
			BattleEffect: BattleItemEffect{
				ActionType:   "heal",
				HealModifier: 1.20, // +20% healing
				Consumable:   true,
			},
		},
		"shield_ring": {
			ID:   "shield_ring",
			Name: "Shield Ring",
			BattleEffect: BattleItemEffect{
				DefenseModifier: 1.10, // +10% defense
				Duration:        3,    // 3 turns
				Consumable:      false,
			},
		},
		"speed_boots": {
			ID:   "speed_boots",
			Name: "Speed Boots",
			BattleEffect: BattleItemEffect{
				SpeedModifier: 1.08, // +8% speed
				Duration:      2,    // 2 turns
				Consumable:    false,
			},
		},
		"overpowered_weapon": {
			ID:   "overpowered_weapon",
			Name: "Overpowered Weapon",
			BattleEffect: BattleItemEffect{
				ActionType:     "attack",
				DamageModifier: 2.5, // +150% damage (should be capped)
				Consumable:     false,
			},
		},
	}
}

func TestItemIntegration_DamageModifier(t *testing.T) {
	// Setup
	giftProvider := NewMockGiftProvider()
	testGifts := createTestGifts()
	for _, gift := range testGifts {
		giftProvider.AddGift(gift)
	}

	bm := NewBattleManagerWithGifts(giftProvider)

	// Create a simple battle state
	bm.currentBattle = &BattleState{
		BattleID:     "test_battle",
		Participants: make(map[string]*BattleParticipant),
		Phase:        PHASE_ACTIVE,
	}

	// Add participants
	bm.currentBattle.Participants["player1"] = &BattleParticipant{
		CharacterID: "player1",
		IsLocal:     true,
		Stats: BattleStats{
			HP:    100,
			MaxHP: 100,
		},
	}

	// Test attack action with damage potion
	action := BattleAction{
		Type:     ACTION_ATTACK,
		ActorID:  "player1",
		TargetID: "player1", // Self-target for testing
		ItemUsed: "damage_potion",
	}

	result, err := bm.PerformAction(action, "player1")
	if err != nil {
		t.Fatalf("PerformAction failed: %v", err)
	}

	// Verify damage was modified (BASE_ATTACK_DAMAGE * 1.15)
	expectedDamage := BASE_ATTACK_DAMAGE * 1.15
	if result.Damage != expectedDamage {
		t.Errorf("Expected damage %f, got %f", expectedDamage, result.Damage)
	}

	if !result.Success {
		t.Error("Action should have succeeded")
	}
}

func TestItemIntegration_HealModifier(t *testing.T) {
	// Setup
	giftProvider := NewMockGiftProvider()
	testGifts := createTestGifts()
	for _, gift := range testGifts {
		giftProvider.AddGift(gift)
	}

	bm := NewBattleManagerWithGifts(giftProvider)

	// Create a simple battle state
	bm.currentBattle = &BattleState{
		BattleID:     "test_battle",
		Participants: make(map[string]*BattleParticipant),
		Phase:        PHASE_ACTIVE,
	}

	// Add participants with reduced health
	bm.currentBattle.Participants["player1"] = &BattleParticipant{
		CharacterID: "player1",
		IsLocal:     true,
		Stats: BattleStats{
			HP:    50, // Reduced health
			MaxHP: 100,
		},
	}

	// Test heal action with healing herb
	action := BattleAction{
		Type:     ACTION_HEAL,
		ActorID:  "player1",
		TargetID: "player1",
		ItemUsed: "heal_herb",
	}

	result, err := bm.PerformAction(action, "player1")
	if err != nil {
		t.Fatalf("PerformAction failed: %v", err)
	}

	// Verify healing was modified (BASE_HEAL_AMOUNT * 1.20)
	expectedHealing := BASE_HEAL_AMOUNT * 1.20
	if result.Healing != expectedHealing {
		t.Errorf("Expected healing %f, got %f", expectedHealing, result.Healing)
	}
}

func TestItemIntegration_DefenseModifier(t *testing.T) {
	// Setup
	giftProvider := NewMockGiftProvider()
	testGifts := createTestGifts()
	for _, gift := range testGifts {
		giftProvider.AddGift(gift)
	}

	bm := NewBattleManagerWithGifts(giftProvider)

	// Create a simple battle state
	bm.currentBattle = &BattleState{
		BattleID:     "test_battle",
		Participants: make(map[string]*BattleParticipant),
		Phase:        PHASE_ACTIVE,
	}

	// Add participants
	bm.currentBattle.Participants["player1"] = &BattleParticipant{
		CharacterID: "player1",
		IsLocal:     true,
		Stats: BattleStats{
			HP:    100,
			MaxHP: 100,
		},
	}

	// Test defend action with shield ring
	action := BattleAction{
		Type:     ACTION_DEFEND,
		ActorID:  "player1",
		TargetID: "player1",
		ItemUsed: "shield_ring",
	}

	result, err := bm.PerformAction(action, "player1")
	if err != nil {
		t.Fatalf("PerformAction failed: %v", err)
	}

	// Check that defense modifier was applied
	foundDefenseModifier := false
	for _, modifier := range result.ModifiersApplied {
		if modifier.Type == MODIFIER_DEFENSE {
			// Should have both the base defend modifier and the item modifier
			foundDefenseModifier = true
			break
		}
	}

	if !foundDefenseModifier {
		t.Error("Expected defense modifier to be applied")
	}

	// Should have modifiers from both defend action and shield ring
	if len(result.ModifiersApplied) < 2 {
		t.Errorf("Expected at least 2 modifiers, got %d", len(result.ModifiersApplied))
	}
}

func TestItemIntegration_FairnessCaps(t *testing.T) {
	// Setup
	giftProvider := NewMockGiftProvider()
	testGifts := createTestGifts()
	for _, gift := range testGifts {
		giftProvider.AddGift(gift)
	}

	bm := NewBattleManagerWithGifts(giftProvider)

	// Create a simple battle state
	bm.currentBattle = &BattleState{
		BattleID:     "test_battle",
		Participants: make(map[string]*BattleParticipant),
		Phase:        PHASE_ACTIVE,
	}

	// Add participants
	bm.currentBattle.Participants["player1"] = &BattleParticipant{
		CharacterID: "player1",
		IsLocal:     true,
		Stats: BattleStats{
			HP:    100,
			MaxHP: 100,
		},
	}

	// Test overpowered weapon (should be capped)
	action := BattleAction{
		Type:     ACTION_ATTACK,
		ActorID:  "player1",
		TargetID: "player1",
		ItemUsed: "overpowered_weapon",
	}

	result, err := bm.PerformAction(action, "player1")
	if err != nil {
		t.Fatalf("PerformAction failed: %v", err)
	}

	// Damage should be capped at MAX_DAMAGE_MODIFIER
	maxAllowedDamage := BASE_ATTACK_DAMAGE * MAX_DAMAGE_MODIFIER
	if result.Damage > maxAllowedDamage {
		t.Errorf("Damage %f exceeds maximum allowed %f", result.Damage, maxAllowedDamage)
	}
}

func TestItemIntegration_InvalidItem(t *testing.T) {
	// Setup
	giftProvider := NewMockGiftProvider()
	bm := NewBattleManagerWithGifts(giftProvider)

	// Create a simple battle state
	bm.currentBattle = &BattleState{
		BattleID:     "test_battle",
		Participants: make(map[string]*BattleParticipant),
		Phase:        PHASE_ACTIVE,
	}

	// Add participants
	bm.currentBattle.Participants["player1"] = &BattleParticipant{
		CharacterID: "player1",
		IsLocal:     true,
		Stats: BattleStats{
			HP:    100,
			MaxHP: 100,
		},
	}

	// Test with non-existent item
	action := BattleAction{
		Type:     ACTION_ATTACK,
		ActorID:  "player1",
		TargetID: "player1",
		ItemUsed: "non_existent_item",
	}

	result, err := bm.PerformAction(action, "player1")
	if err != nil {
		t.Fatalf("PerformAction failed: %v", err)
	}

	// Should proceed with base damage (no item effects)
	if result.Damage != BASE_ATTACK_DAMAGE {
		t.Errorf("Expected base damage %f, got %f", BASE_ATTACK_DAMAGE, result.Damage)
	}
}

func TestItemIntegration_WrongActionType(t *testing.T) {
	// Setup
	giftProvider := NewMockGiftProvider()
	testGifts := createTestGifts()
	for _, gift := range testGifts {
		giftProvider.AddGift(gift)
	}

	bm := NewBattleManagerWithGifts(giftProvider)

	// Create a simple battle state
	bm.currentBattle = &BattleState{
		BattleID:     "test_battle",
		Participants: make(map[string]*BattleParticipant),
		Phase:        PHASE_ACTIVE,
	}

	// Add participants with reduced health to allow healing
	bm.currentBattle.Participants["player1"] = &BattleParticipant{
		CharacterID: "player1",
		IsLocal:     true,
		Stats: BattleStats{
			HP:    50, // Reduced health to allow healing
			MaxHP: 100,
		},
	}

	// Test using damage potion (for "attack") with heal action
	action := BattleAction{
		Type:     ACTION_HEAL,
		ActorID:  "player1",
		TargetID: "player1",
		ItemUsed: "damage_potion", // Wrong action type
	}

	result, err := bm.PerformAction(action, "player1")
	if err != nil {
		t.Fatalf("PerformAction failed: %v", err)
	}

	// Should proceed with base healing (no item effects)
	if result.Healing != BASE_HEAL_AMOUNT {
		t.Errorf("Expected base healing %f, got %f", BASE_HEAL_AMOUNT, result.Healing)
	}
}

// AI Tests

func TestAI_ItemSelection(t *testing.T) {
	// Setup with only basic, reasonable items
	giftProvider := NewMockGiftProvider()

	// Add only reasonable items (not overpowered ones)
	basicGifts := map[string]GiftDefinition{
		"damage_potion": {
			ID:   "damage_potion",
			Name: "Damage Potion",
			BattleEffect: BattleItemEffect{
				ActionType:     "attack",
				DamageModifier: 1.15, // +15% damage
				Consumable:     true,
			},
		},
		"heal_herb": {
			ID:   "heal_herb",
			Name: "Healing Herb",
			BattleEffect: BattleItemEffect{
				ActionType:   "heal",
				HealModifier: 1.20, // +20% healing
				Consumable:   true,
			},
		},
		"shield_ring": {
			ID:   "shield_ring",
			Name: "Shield Ring",
			BattleEffect: BattleItemEffect{
				DefenseModifier: 1.10, // +10% defense
				Duration:        3,    // 3 turns
				Consumable:      false,
			},
		},
	}

	for _, gift := range basicGifts {
		giftProvider.AddGift(gift)
	}

	// Create AI with gift provider
	ai := NewBattleAIWithGifts("player1", AI_EXPERT, STRATEGY_AGGRESSIVE, giftProvider)

	// Test item selection for attack action
	bestItem := ai.selectBestItem(ACTION_ATTACK)
	if bestItem != "damage_potion" {
		t.Errorf("Expected 'damage_potion' for attack, got '%s'", bestItem)
	}

	// Test item selection for heal action
	bestItem = ai.selectBestItem(ACTION_HEAL)
	if bestItem != "heal_herb" {
		t.Errorf("Expected 'heal_herb' for heal, got '%s'", bestItem)
	}
}

func TestAI_ItemScoring(t *testing.T) {
	// Setup
	giftProvider := NewMockGiftProvider()
	testGifts := createTestGifts()
	for _, gift := range testGifts {
		giftProvider.AddGift(gift)
	}

	ai := NewBattleAIWithGifts("player1", AI_EXPERT, STRATEGY_AGGRESSIVE, giftProvider)

	// Test scoring for damage potion with attack action
	score := ai.calculateItemScore(testGifts["damage_potion"], ACTION_ATTACK)
	if score <= 0 {
		t.Error("Damage potion should have positive score for attack action")
	}

	// Test scoring for damage potion with heal action (should be 0)
	score = ai.calculateItemScore(testGifts["damage_potion"], ACTION_HEAL)
	if score != 0 {
		t.Error("Damage potion should have zero score for heal action")
	}

	// Test scoring for shield ring (duration bonus)
	score = ai.calculateItemScore(testGifts["shield_ring"], ACTION_DEFEND)
	if score <= 0 {
		t.Error("Shield ring should have positive score for defend action")
	}
}

func TestAI_ItemUsageByDifficulty(t *testing.T) {
	// Setup
	giftProvider := NewMockGiftProvider()
	testGifts := createTestGifts()
	for _, gift := range testGifts {
		giftProvider.AddGift(gift)
	}

	// Test different difficulty levels
	difficulties := []AIDifficulty{AI_EASY, AI_NORMAL, AI_HARD, AI_EXPERT}

	for _, difficulty := range difficulties {
		ai := NewBattleAIWithGifts("player1", difficulty, STRATEGY_BALANCED, giftProvider)

		// Test item enhancement multiple times to check probability
		itemUsageCount := 0
		totalTests := 100

		for i := 0; i < totalTests; i++ {
			baseAction := BattleAction{
				Type:    ACTION_ATTACK,
				ActorID: "player1",
			}

			enhancedAction := ai.enhanceActionWithItem(baseAction)
			if enhancedAction.ItemUsed != "" {
				itemUsageCount++
			}
		}

		itemUsageRate := float64(itemUsageCount) / float64(totalTests)

		// Verify item usage rate matches expected difficulty behavior
		switch difficulty {
		case AI_EASY:
			if itemUsageRate > 0.3 { // Should be around 10% but allow variance
				t.Errorf("Easy AI uses items too frequently: %f", itemUsageRate)
			}
		case AI_EXPERT:
			if itemUsageRate < 0.5 { // Should be around 80% but allow variance
				t.Errorf("Expert AI doesn't use items frequently enough: %f", itemUsageRate)
			}
		}
	}
}

func TestBattleManager_NoGiftProvider(t *testing.T) {
	// Test that battle system works without gift provider
	bm := NewBattleManager() // No gift provider

	// Create a simple battle state
	bm.currentBattle = &BattleState{
		BattleID:     "test_battle",
		Participants: make(map[string]*BattleParticipant),
		Phase:        PHASE_ACTIVE,
	}

	// Add participants
	bm.currentBattle.Participants["player1"] = &BattleParticipant{
		CharacterID: "player1",
		IsLocal:     true,
		Stats: BattleStats{
			HP:    100,
			MaxHP: 100,
		},
	}

	// Test action with item specified (should be ignored)
	action := BattleAction{
		Type:     ACTION_ATTACK,
		ActorID:  "player1",
		TargetID: "player1",
		ItemUsed: "some_item", // Should be ignored
	}

	result, err := bm.PerformAction(action, "player1")
	if err != nil {
		t.Fatalf("PerformAction failed: %v", err)
	}

	// Should use base damage (no item effects)
	if result.Damage != BASE_ATTACK_DAMAGE {
		t.Errorf("Expected base damage %f, got %f", BASE_ATTACK_DAMAGE, result.Damage)
	}
}
