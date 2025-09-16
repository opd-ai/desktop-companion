package battle

import (
	"testing"
)

func TestNewPersonalityBasedAI(t *testing.T) {
	ai := NewPersonalityBasedAI()

	if ai == nil {
		t.Fatal("NewPersonalityBasedAI returned nil")
	}

	if ai.personalities == nil {
		t.Error("personalities map not initialized")
	}

	if ai.rand == nil {
		t.Error("random generator not initialized")
	}

	// Verify all personality types are initialized
	expectedPersonalities := []AIPersonality{
		PERSONALITY_AGGRESSIVE, PERSONALITY_DEFENSIVE, PERSONALITY_BALANCED, PERSONALITY_TACTICAL,
	}

	for _, personality := range expectedPersonalities {
		if _, exists := ai.personalities[personality]; !exists {
			t.Errorf("Personality %s not initialized", personality)
		}
	}
}

func TestPersonalityWeightCharacteristics(t *testing.T) {
	ai := NewPersonalityBasedAI()

	tests := []struct {
		personality   AIPersonality
		expectedTrait string
		checkFunction func(AIDecisionWeights) bool
	}{
		{
			PERSONALITY_AGGRESSIVE,
			"high attack weight",
			func(w AIDecisionWeights) bool { return w.AttackWeight > 0.7 },
		},
		{
			PERSONALITY_DEFENSIVE,
			"high defense weight",
			func(w AIDecisionWeights) bool { return w.DefenseWeight > 0.8 },
		},
		{
			PERSONALITY_BALANCED,
			"moderate weights",
			func(w AIDecisionWeights) bool {
				return w.AttackWeight >= 0.5 && w.DefenseWeight >= 0.5 && w.RiskTolerance == 0.5
			},
		},
		{
			PERSONALITY_TACTICAL,
			"high planning depth",
			func(w AIDecisionWeights) bool { return w.PlanningDepth >= 5 },
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.personality), func(t *testing.T) {
			weights, exists := ai.GetPersonalityWeights(tt.personality)
			if !exists {
				t.Fatalf("Personality %s not found", tt.personality)
			}

			if !tt.checkFunction(weights) {
				t.Errorf("Personality %s doesn't have expected trait: %s", tt.personality, tt.expectedTrait)
			}
		})
	}
}

func TestGetOptimalActionType_BasicScenarios(t *testing.T) {
	ai := NewPersonalityBasedAI()

	// Test aggressive scenario
	actionType, err := ai.GetOptimalActionType(PERSONALITY_AGGRESSIVE, 0.8, 0.6, 1)
	if err != nil {
		t.Fatalf("GetOptimalActionType failed: %v", err)
	}

	// Aggressive AI should prefer offensive actions
	if actionType != ACTION_ATTACK && actionType != ACTION_STUN && actionType != ACTION_DRAIN {
		t.Errorf("Aggressive AI should prefer offensive actions, got %s", actionType)
	}
}

func TestGetOptimalActionType_LowHPScenario(t *testing.T) {
	ai := NewPersonalityBasedAI()

	// Test defensive behavior when low HP
	actionType, err := ai.GetOptimalActionType(PERSONALITY_DEFENSIVE, 0.2, 0.8, 3)
	if err != nil {
		t.Fatalf("GetOptimalActionType failed: %v", err)
	}

	// Defensive AI with low HP should prioritize healing
	if actionType != ACTION_HEAL && actionType != ACTION_DEFEND && actionType != ACTION_SHIELD {
		t.Errorf("Defensive AI with low HP should prioritize healing/defense, got %s", actionType)
	}
}

func TestGetOptimalActionType_EnemyLowHPScenario(t *testing.T) {
	ai := NewPersonalityBasedAI()

	// Test finishing behavior when enemy is low HP
	actionType, err := ai.GetOptimalActionType(PERSONALITY_AGGRESSIVE, 0.7, 0.15, 5)
	if err != nil {
		t.Fatalf("GetOptimalActionType failed: %v", err)
	}

	// Should prefer attack to finish enemy
	if actionType != ACTION_ATTACK {
		t.Errorf("AI should try to finish weak enemy with attack, got %s", actionType)
	}
}

func TestGetOptimalActionType_BalancedPersonality(t *testing.T) {
	ai := NewPersonalityBasedAI()

	actionType, err := ai.GetOptimalActionType(PERSONALITY_BALANCED, 0.6, 0.7, 2)
	if err != nil {
		t.Fatalf("GetOptimalActionType failed: %v", err)
	}

	// Balanced AI should make reasonable decisions
	validActions := []BattleActionType{ACTION_ATTACK, ACTION_DEFEND, ACTION_HEAL}
	isValid := false
	for _, validAction := range validActions {
		if actionType == validAction {
			isValid = true
			break
		}
	}

	if !isValid {
		t.Errorf("Balanced AI should choose valid actions, got %s", actionType)
	}
}

func TestGetOptimalActionType_TacticalPersonality(t *testing.T) {
	ai := NewPersonalityBasedAI()

	actionType, err := ai.GetOptimalActionType(PERSONALITY_TACTICAL, 0.6, 0.7, 2)
	if err != nil {
		t.Fatalf("GetOptimalActionType failed: %v", err)
	}

	// Tactical AI should make strategic decisions
	if actionType == "" {
		t.Error("Tactical AI should return a valid action type")
	}
}

func TestAnalyzeBattleSituation(t *testing.T) {
	ai := NewPersonalityBasedAI()

	analysis := ai.AnalyzeBattleSituation(0.6, 0.4, 5)

	expectedKeys := []string{
		"hp_percentage", "enemy_hp_percentage", "turn_number",
		"danger_level", "victory_proximity",
	}

	for _, key := range expectedKeys {
		if _, exists := analysis[key]; !exists {
			t.Errorf("Analysis missing key: %s", key)
		}
	}

	// Verify calculated values
	if analysis["hp_percentage"] != 0.6 {
		t.Errorf("Expected hp_percentage 0.6, got %f", analysis["hp_percentage"])
	}

	if analysis["danger_level"] != 0.4 {
		t.Errorf("Expected danger_level 0.4, got %f", analysis["danger_level"])
	}

	if analysis["victory_proximity"] != 0.6 {
		t.Errorf("Expected victory_proximity 0.6, got %f", analysis["victory_proximity"])
	}
}

func TestGetPersonalityDescription(t *testing.T) {
	ai := NewPersonalityBasedAI()

	personalities := []AIPersonality{
		PERSONALITY_AGGRESSIVE, PERSONALITY_DEFENSIVE, PERSONALITY_BALANCED, PERSONALITY_TACTICAL,
	}

	for _, personality := range personalities {
		description := ai.GetPersonalityDescription(personality)
		if description == "" {
			t.Errorf("Empty description for personality %s", personality)
		}

		if len(description) < 10 {
			t.Errorf("Description for %s should be more descriptive, got: %s",
				personality, description)
		}
	}
}

func TestAdjustPersonality(t *testing.T) {
	ai := NewPersonalityBasedAI()

	// Get original weights
	originalWeights, exists := ai.GetPersonalityWeights(PERSONALITY_BALANCED)
	if !exists {
		t.Fatal("PERSONALITY_BALANCED should exist")
	}

	// Create new weights
	newWeights := AIDecisionWeights{
		AttackWeight:  0.9,
		DefenseWeight: 0.1,
		RiskTolerance: 0.8,
		PlanningDepth: 2,
	}

	// Adjust personality
	ai.AdjustPersonality(PERSONALITY_BALANCED, newWeights)

	// Verify changes
	updatedWeights, exists := ai.GetPersonalityWeights(PERSONALITY_BALANCED)
	if !exists {
		t.Fatal("PERSONALITY_BALANCED should still exist after adjustment")
	}

	if updatedWeights.AttackWeight != 0.9 {
		t.Errorf("Expected AttackWeight 0.9, got %f", updatedWeights.AttackWeight)
	}

	if updatedWeights.DefenseWeight != 0.1 {
		t.Errorf("Expected DefenseWeight 0.1, got %f", updatedWeights.DefenseWeight)
	}

	// Reset for other tests
	ai.AdjustPersonality(PERSONALITY_BALANCED, originalWeights)
}

func TestUnknownPersonality(t *testing.T) {
	ai := NewPersonalityBasedAI()

	// Test with unknown personality
	_, err := ai.GetOptimalActionType("unknown_personality", 0.5, 0.5, 1)
	if err == nil {
		t.Error("Should return error for unknown personality")
	}

	// Test GetPersonalityWeights with unknown personality
	_, exists := ai.GetPersonalityWeights("unknown_personality")
	if exists {
		t.Error("Should return false for unknown personality")
	}
}

func TestPersonalityConsistency(t *testing.T) {
	ai := NewPersonalityBasedAI()

	// Test that aggressive AI consistently chooses more aggressive actions
	aggressiveChoices := make(map[BattleActionType]int)
	defensiveChoices := make(map[BattleActionType]int)

	// Run multiple iterations
	for i := 0; i < 10; i++ {
		aggressiveAction, _ := ai.GetOptimalActionType(PERSONALITY_AGGRESSIVE, 0.7, 0.7, i)
		defensiveAction, _ := ai.GetOptimalActionType(PERSONALITY_DEFENSIVE, 0.7, 0.7, i)

		aggressiveChoices[aggressiveAction]++
		defensiveChoices[defensiveAction]++
	}

	// Aggressive should favor attack-type actions
	if aggressiveChoices[ACTION_ATTACK] == 0 && aggressiveChoices[ACTION_STUN] == 0 && aggressiveChoices[ACTION_DRAIN] == 0 {
		t.Error("Aggressive personality should sometimes choose offensive actions")
	}

	// Defensive should favor defensive actions
	if defensiveChoices[ACTION_DEFEND] == 0 && defensiveChoices[ACTION_HEAL] == 0 && defensiveChoices[ACTION_SHIELD] == 0 {
		t.Error("Defensive personality should sometimes choose defensive actions")
	}
}

// Benchmark tests for AI performance
func BenchmarkGetOptimalActionType(b *testing.B) {
	ai := NewPersonalityBasedAI()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ai.GetOptimalActionType(PERSONALITY_TACTICAL, 0.6, 0.7, 3)
	}
}

func BenchmarkAnalyzeBattleSituation(b *testing.B) {
	ai := NewPersonalityBasedAI()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ai.AnalyzeBattleSituation(0.6, 0.7, 3)
	}
}
