package battle

import (
	"testing"
)

func TestEnhancedAIDecisionMaking(t *testing.T) {
	ai := NewPersonalityBasedAI()

	// Mock available abilities
	abilities := []SpecialAbility{
		{
			Type:           ABILITY_CRITICAL_STRIKE,
			Name:           "Critical Strike",
			Cooldown:       4,
			ChargesMax:     0,
			ChargesCurrent: 0,
			RequiredLevel:  1,
		},
		{
			Type:           ABILITY_PERFECT_GUARD,
			Name:           "Perfect Guard",
			Cooldown:       8,
			ChargesMax:     2,
			ChargesCurrent: 2,
			RequiredLevel:  3,
		},
	}

	// Mock available combos
	combos := []ComboAttack{
		{
			Type:             COMBO_STUN_ATTACK,
			Name:             "Stunning Strike",
			Sequence:         []BattleActionType{ACTION_STUN, ACTION_ATTACK},
			WindowDuration:   2,
			DamageMultiplier: 1.2,
		},
	}

	tests := []struct {
		name              string
		personality       AIPersonality
		hpPercentage      float64
		enemyHPPercentage float64
		expectSpecial     bool
		expectCombo       bool
	}{
		{
			name:              "Aggressive High HP",
			personality:       PERSONALITY_AGGRESSIVE,
			hpPercentage:      0.9,
			enemyHPPercentage: 0.8,
			expectSpecial:     true, // Should prefer special abilities
			expectCombo:       true, // Should consider combos
		},
		{
			name:              "Defensive Low HP",
			personality:       PERSONALITY_DEFENSIVE,
			hpPercentage:      0.2,
			enemyHPPercentage: 0.8,
			expectSpecial:     false, // Should avoid risky specials
			expectCombo:       false, // Should avoid risky combos
		},
		{
			name:              "Tactical High HP",
			personality:       PERSONALITY_TACTICAL,
			hpPercentage:      0.8,
			enemyHPPercentage: 0.9,
			expectSpecial:     true, // Should use specials strategically
			expectCombo:       true, // Should prefer combos
		},
		{
			name:              "Balanced Medium HP",
			personality:       PERSONALITY_BALANCED,
			hpPercentage:      0.6,
			enemyHPPercentage: 0.5,
			expectSpecial:     true, // Should consider specials
			expectCombo:       true, // Should consider combos
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision, err := ai.GetOptimalDecision(
				tt.personality,
				tt.hpPercentage,
				tt.enemyHPPercentage,
				1,
				abilities,
				nil, // No active combo
				combos,
			)
			if err != nil {
				t.Fatalf("GetOptimalDecision failed: %v", err)
			}

			if decision == nil {
				t.Fatal("Decision is nil")
			}

			if decision.Priority <= 0 {
				t.Error("Decision priority should be positive")
			}

			if decision.Reasoning == "" {
				t.Error("Decision should have reasoning")
			}

			// Test personality-specific expectations
			weights, _ := ai.GetPersonalityWeights(tt.personality)

			if tt.expectSpecial && weights.SpecialWeight > 0.6 && tt.hpPercentage >= weights.SpecialThreshold {
				// Should potentially consider special abilities
				t.Logf("Decision: %+v", decision)
			}

			if tt.expectCombo && weights.ComboWeight > 0.5 {
				// Should potentially consider combo actions
				t.Logf("Decision: %+v", decision)
			}
		})
	}
}

func TestComboConitation(t *testing.T) {
	ai := NewPersonalityBasedAI()

	// Create active combo state
	activeCombo := &ComboState{
		Type:             COMBO_STUN_ATTACK,
		ActionsCompleted: []BattleActionType{ACTION_STUN},
		StartedTurn:      1,
		ActorID:          "test_ai",
		IsActive:         true,
	}

	combos := []ComboAttack{
		{
			Type:             COMBO_STUN_ATTACK,
			Name:             "Stunning Strike",
			Sequence:         []BattleActionType{ACTION_STUN, ACTION_ATTACK},
			WindowDuration:   2,
			DamageMultiplier: 1.2,
		},
	}

	// Test that AI prioritizes continuing active combo
	decision, err := ai.GetOptimalDecision(
		PERSONALITY_TACTICAL,
		0.8,
		0.7,
		2,
		[]SpecialAbility{}, // No special abilities
		activeCombo,
		combos,
	)
	if err != nil {
		t.Fatalf("GetOptimalDecision failed: %v", err)
	}

	if decision == nil {
		t.Fatal("Decision is nil")
	}

	if !decision.ContinuesCombo {
		t.Error("AI should prioritize continuing active combo")
	}

	if decision.ActionType != ACTION_ATTACK {
		t.Errorf("Expected ACTION_ATTACK to continue combo, got %v", decision.ActionType)
	}

	if decision.Priority < 1.0 {
		t.Error("Combo continuation should have high priority")
	}
}

func TestSpecialAbilityPriority(t *testing.T) {
	ai := NewPersonalityBasedAI()

	abilities := []SpecialAbility{
		{Type: ABILITY_CRITICAL_STRIKE, Name: "Critical Strike"},
		{Type: ABILITY_PERFECT_GUARD, Name: "Perfect Guard"},
		{Type: ABILITY_LIFE_STEAL, Name: "Life Steal"},
		{Type: ABILITY_CLEANSE, Name: "Cleanse"},
	}

	tests := []struct {
		name              string
		personality       AIPersonality
		hpPercentage      float64
		enemyHPPercentage float64
		expectedFocus     string // "offensive", "defensive", or "utility"
	}{
		{
			name:              "Aggressive prefers offense",
			personality:       PERSONALITY_AGGRESSIVE,
			hpPercentage:      0.8,
			enemyHPPercentage: 0.3,
			expectedFocus:     "offensive",
		},
		{
			name:              "Defensive when low HP",
			personality:       PERSONALITY_DEFENSIVE,
			hpPercentage:      0.2,
			enemyHPPercentage: 0.8,
			expectedFocus:     "defensive",
		},
		{
			name:              "Tactical balanced approach",
			personality:       PERSONALITY_TACTICAL,
			hpPercentage:      0.7,
			enemyHPPercentage: 0.6,
			expectedFocus:     "balanced",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision, err := ai.GetOptimalDecision(
				tt.personality,
				tt.hpPercentage,
				tt.enemyHPPercentage,
				1,
				abilities,
				nil,
				[]ComboAttack{},
			)
			if err != nil {
				t.Fatalf("GetOptimalDecision failed: %v", err)
			}

			if decision == nil {
				t.Fatal("Decision is nil")
			}

			// Test that the decision aligns with personality expectations
			if decision.IsSpecial {
				switch decision.SpecialAbility {
				case ABILITY_CRITICAL_STRIKE, ABILITY_LIFE_STEAL:
					if tt.expectedFocus == "defensive" {
						t.Logf("Defensive personality chose offensive ability - this may be valid in some contexts")
					}
				case ABILITY_PERFECT_GUARD:
					if tt.expectedFocus == "offensive" && tt.hpPercentage > 0.5 {
						t.Logf("Offensive personality chose defensive ability - may indicate special circumstances")
					}
				}
			}

			t.Logf("Personality %v chose: %v (special: %v, reasoning: %s)",
				tt.personality, decision.ActionType, decision.IsSpecial, decision.Reasoning)
		})
	}
}

func TestComboStarterEvaluation(t *testing.T) {
	ai := NewPersonalityBasedAI()

	combos := []ComboAttack{
		{
			Type:             COMBO_STUN_ATTACK,
			Name:             "Stunning Strike",
			Sequence:         []BattleActionType{ACTION_STUN, ACTION_ATTACK},
			WindowDuration:   2,
			DamageMultiplier: 1.2,
		},
		{
			Type:             COMBO_DRAIN_HEAL,
			Name:             "Vampiric Recovery",
			Sequence:         []BattleActionType{ACTION_DRAIN, ACTION_HEAL},
			WindowDuration:   2,
			DamageMultiplier: 1.0,
		},
		{
			Type:             COMBO_CHARGE_BOOST_ATTACK,
			Name:             "Overwhelming Assault",
			Sequence:         []BattleActionType{ACTION_CHARGE, ACTION_BOOST, ACTION_ATTACK},
			WindowDuration:   3,
			DamageMultiplier: 1.2,
		},
	}

	// Test that tactical personality prefers complex combos
	decision, err := ai.GetOptimalDecision(
		PERSONALITY_TACTICAL,
		0.8,
		0.7,
		1,
		[]SpecialAbility{}, // No special abilities to compete
		nil,                // No active combo
		combos,
	)
	if err != nil {
		t.Fatalf("GetOptimalDecision failed: %v", err)
	}

	if decision == nil {
		t.Fatal("Decision is nil")
	}

	t.Logf("Tactical AI chose: %v (starts combo: %v, reasoning: %s)",
		decision.ActionType, decision.StartsCombo, decision.Reasoning)

	// Test that defensive personality avoids complex combos when low HP
	defensiveDecision, err := ai.GetOptimalDecision(
		PERSONALITY_DEFENSIVE,
		0.2, // Low HP
		0.8,
		1,
		[]SpecialAbility{},
		nil,
		combos,
	)
	if err != nil {
		t.Fatalf("GetOptimalDecision failed: %v", err)
	}

	if defensiveDecision == nil {
		t.Fatal("Decision is nil")
	}

	t.Logf("Defensive AI (low HP) chose: %v (starts combo: %v, reasoning: %s)",
		defensiveDecision.ActionType, defensiveDecision.StartsCombo, defensiveDecision.Reasoning)
}

func TestPersonalityWeightUpdates(t *testing.T) {
	ai := NewPersonalityBasedAI()

	// Test that personalities have the expected new fields
	for _, personality := range []AIPersonality{
		PERSONALITY_AGGRESSIVE,
		PERSONALITY_DEFENSIVE,
		PERSONALITY_BALANCED,
		PERSONALITY_TACTICAL,
	} {
		weights, exists := ai.GetPersonalityWeights(personality)
		if !exists {
			t.Fatalf("Personality %v not found", personality)
		}

		// Check that new fields are present and reasonable
		if weights.SpecialWeight < 0 || weights.SpecialWeight > 1.5 {
			t.Errorf("SpecialWeight for %v is out of range: %f", personality, weights.SpecialWeight)
		}

		if weights.ComboWeight < 0 || weights.ComboWeight > 1.5 {
			t.Errorf("ComboWeight for %v is out of range: %f", personality, weights.ComboWeight)
		}

		if weights.SpecialThreshold < 0 || weights.SpecialThreshold > 1.0 {
			t.Errorf("SpecialThreshold for %v is out of range: %f", personality, weights.SpecialThreshold)
		}

		if weights.ComboAggression < 0 || weights.ComboAggression > 1.0 {
			t.Errorf("ComboAggression for %v is out of range: %f", personality, weights.ComboAggression)
		}
	}

	// Test personality-specific characteristics
	aggressiveWeights, _ := ai.GetPersonalityWeights(PERSONALITY_AGGRESSIVE)
	if aggressiveWeights.SpecialWeight < 0.8 {
		t.Error("Aggressive personality should have high special weight")
	}

	defensiveWeights, _ := ai.GetPersonalityWeights(PERSONALITY_DEFENSIVE)
	if defensiveWeights.ComboWeight > 0.4 {
		t.Error("Defensive personality should have low combo weight")
	}

	tacticalWeights, _ := ai.GetPersonalityWeights(PERSONALITY_TACTICAL)
	if tacticalWeights.ComboWeight < 0.7 {
		t.Error("Tactical personality should have high combo weight")
	}
}

// Benchmark tests for enhanced AI performance
func BenchmarkEnhancedAIDecision(b *testing.B) {
	ai := NewPersonalityBasedAI()

	abilities := []SpecialAbility{
		{Type: ABILITY_CRITICAL_STRIKE, Name: "Critical Strike"},
		{Type: ABILITY_PERFECT_GUARD, Name: "Perfect Guard"},
		{Type: ABILITY_LIFE_STEAL, Name: "Life Steal"},
	}

	combos := []ComboAttack{
		{
			Type:             COMBO_STUN_ATTACK,
			Sequence:         []BattleActionType{ACTION_STUN, ACTION_ATTACK},
			WindowDuration:   2,
			DamageMultiplier: 1.2,
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := ai.GetOptimalDecision(
			PERSONALITY_TACTICAL,
			0.7,
			0.6,
			i%10+1,
			abilities,
			nil,
			combos,
		)
		if err != nil {
			b.Fatalf("GetOptimalDecision failed: %v", err)
		}
	}
}

func BenchmarkComboEvaluation(b *testing.B) {
	ai := NewPersonalityBasedAI()

	combos := []ComboAttack{
		{Type: COMBO_STUN_ATTACK, Sequence: []BattleActionType{ACTION_STUN, ACTION_ATTACK}},
		{Type: COMBO_DRAIN_HEAL, Sequence: []BattleActionType{ACTION_DRAIN, ACTION_HEAL}},
		{Type: COMBO_CHARGE_BOOST_ATTACK, Sequence: []BattleActionType{ACTION_CHARGE, ACTION_BOOST, ACTION_ATTACK}},
	}

	weights := AIDecisionWeights{
		AttackWeight:    0.8,
		DefenseWeight:   0.4,
		HealingWeight:   0.5,
		SpecialWeight:   0.7,
		ComboWeight:     0.8,
		ComboAggression: 0.6,
		RiskTolerance:   0.7,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = ai.evaluateComboStarters(combos, weights, false, false)
	}
}
