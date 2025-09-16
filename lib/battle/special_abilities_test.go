package battle

import (
	"testing"
)

func TestNewAbilityManager(t *testing.T) {
	am := NewAbilityManager()
	
	if am == nil {
		t.Fatal("NewAbilityManager returned nil")
	}
	
	if am.participantAbilities == nil {
		t.Error("participantAbilities map not initialized")
	}
	
	if am.activeComboStates == nil {
		t.Error("activeComboStates map not initialized")
	}
	
	if len(am.availableCombos) == 0 {
		t.Error("no default combos loaded")
	}
	
	// Verify we have expected combos
	expectedCombos := []ComboAttackType{
		COMBO_STUN_ATTACK,
		COMBO_BOOST_STRIKE,
		COMBO_DRAIN_HEAL,
		COMBO_CHARGE_BOOST_ATTACK,
		COMBO_SHIELD_COUNTER_STUN,
		COMBO_BERSERKER_FURY,
		COMBO_DEFENSIVE_MASTERY,
	}
	
	comboMap := make(map[ComboAttackType]bool)
	for _, combo := range am.availableCombos {
		comboMap[combo.Type] = true
	}
	
	for _, expected := range expectedCombos {
		if !comboMap[expected] {
			t.Errorf("Expected combo %v not found", expected)
		}
	}
}

func TestInitializeParticipantAbilities(t *testing.T) {
	am := NewAbilityManager()
	participantID := "test_participant"
	
	tests := []struct {
		name          string
		characterLevel int
		minAbilities   int
		maxAbilities   int
	}{
		{"Level 1", 1, 1, 3},
		{"Level 3", 3, 3, 5},
		{"Level 5", 5, 5, 8},
		{"Level 10", 10, 7, 7}, // All abilities
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			am.InitializeParticipantAbilities(participantID, tt.characterLevel)
			
			abilities := am.participantAbilities[participantID]
			if abilities == nil {
				t.Fatal("No abilities initialized")
			}
			
			if len(abilities) < tt.minAbilities {
				t.Errorf("Expected at least %d abilities, got %d", tt.minAbilities, len(abilities))
			}
			
			if len(abilities) > tt.maxAbilities {
				t.Errorf("Expected at most %d abilities, got %d", tt.maxAbilities, len(abilities))
			}
			
			// Check that all abilities meet level requirement
			for _, ability := range abilities {
				if ability.RequiredLevel > tt.characterLevel {
					t.Errorf("Ability %v requires level %d but character is level %d", 
						ability.Type, ability.RequiredLevel, tt.characterLevel)
				}
			}
		})
	}
}

func TestSpecialAbilityExecution(t *testing.T) {
	am := NewAbilityManager()
	participantID := "test_participant"
	am.InitializeParticipantAbilities(participantID, 10) // Max level for all abilities
	
	// Create a mock battle state
	battleState := &BattleState{
		Participants: map[string]*BattleParticipant{
			participantID: {
				CharacterID: participantID,
				Stats: BattleStats{
					HP:      100,
					MaxHP:   100,
					Attack:  50,
					Defense: 30,
					Speed:   40,
				},
			},
		},
	}
	
	tests := []struct {
		name         string
		abilityType  SpecialAbilityType
		expectDamage bool
		expectHeal   bool
		expectMods   bool
		expectStatus bool
	}{
		{"Critical Strike", ABILITY_CRITICAL_STRIKE, true, false, false, true},
		{"Lightning Bolt", ABILITY_LIGHTNING_BOLT, true, false, false, true},
		{"Berserker Rage", ABILITY_BERSERKER_RAGE, false, false, true, false},
		{"Life Steal", ABILITY_LIFE_STEAL, true, true, false, true},
		{"Perfect Guard", ABILITY_PERFECT_GUARD, false, false, true, false},
		{"Sanctuary", ABILITY_SANCTUARY, false, true, true, false},
		{"Cleanse", ABILITY_CLEANSE, false, false, false, true},
		{"Time Freeze", ABILITY_TIME_FREEZE, false, false, false, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := am.UseSpecialAbility(participantID, tt.abilityType, battleState)
			
			if err != nil {
				t.Fatalf("UseSpecialAbility failed: %v", err)
			}
			
			if result == nil {
				t.Fatal("Result is nil")
			}
			
			if !result.Success {
				t.Error("Ability execution was not successful")
			}
			
			if tt.expectDamage && result.Damage <= 0 {
				t.Error("Expected damage but got none")
			}
			
			if !tt.expectDamage && result.Damage > 0 {
				t.Error("Unexpected damage value")
			}
			
			if tt.expectHeal && result.Healing <= 0 {
				t.Error("Expected healing but got none")
			}
			
			if !tt.expectHeal && result.Healing > 0 {
				t.Error("Unexpected healing value")
			}
			
			if tt.expectMods && len(result.ModifiersApplied) == 0 {
				t.Error("Expected modifiers but got none")
			}
			
			if tt.expectStatus && len(result.StatusEffects) == 0 {
				t.Error("Expected status effects but got none")
			}
			
			if result.Animation == "" {
				t.Error("No animation specified")
			}
			
			if result.Response == "" {
				t.Error("No response text specified")
			}
		})
	}
}

func TestSpecialAbilityCooldowns(t *testing.T) {
	am := NewAbilityManager()
	participantID := "test_participant"
	am.InitializeParticipantAbilities(participantID, 10)
	
	battleState := &BattleState{
		Participants: map[string]*BattleParticipant{
			participantID: {
				CharacterID: participantID,
				Stats: BattleStats{HP: 100, MaxHP: 100},
			},
		},
	}
	
	// Use an ability
	_, err := am.UseSpecialAbility(participantID, ABILITY_CRITICAL_STRIKE, battleState)
	if err != nil {
		t.Fatalf("First use failed: %v", err)
	}
	
	// Try to use it again immediately (should fail due to cooldown)
	_, err = am.UseSpecialAbility(participantID, ABILITY_CRITICAL_STRIKE, battleState)
	if err != ErrAbilityOnCooldown {
		t.Errorf("Expected cooldown error, got: %v", err)
	}
	
	// Check available abilities (critical strike should not be available)
	available := am.GetAvailableSpecialAbilities(participantID)
	for _, ability := range available {
		if ability.Type == ABILITY_CRITICAL_STRIKE {
			t.Error("Critical strike should not be available due to cooldown")
		}
	}
}

func TestSpecialAbilityCharges(t *testing.T) {
	am := NewAbilityManager()
	participantID := "test_participant"
	am.InitializeParticipantAbilities(participantID, 10)
	
	battleState := &BattleState{
		Participants: map[string]*BattleParticipant{
			participantID: {
				CharacterID: participantID,
				Stats: BattleStats{HP: 100, MaxHP: 100},
			},
		},
	}
	
	// Perfect Guard has limited charges (2)
	// Use it once
	_, err := am.UseSpecialAbility(participantID, ABILITY_PERFECT_GUARD, battleState)
	if err != nil {
		t.Fatalf("First use failed: %v", err)
	}
	
	// Reset cooldown and use again
	am.currentTurn += 10 // Advance beyond cooldown
	_, err = am.UseSpecialAbility(participantID, ABILITY_PERFECT_GUARD, battleState)
	if err != nil {
		t.Fatalf("Second use failed: %v", err)
	}
	
	// Reset cooldown and try to use third time (should fail - no charges left)
	am.currentTurn += 10
	_, err = am.UseSpecialAbility(participantID, ABILITY_PERFECT_GUARD, battleState)
	if err != ErrAbilityOnCooldown {
		t.Errorf("Expected cooldown/charges error, got: %v", err)
	}
}

func TestComboStart(t *testing.T) {
	am := NewAbilityManager()
	participantID := "test_participant"
	
	// Try starting a combo with the first action of COMBO_STUN_ATTACK (ACTION_STUN)
	combo, err := am.TrackComboAction(participantID, ACTION_STUN)
	
	if err != nil {
		t.Fatalf("TrackComboAction failed: %v", err)
	}
	
	if combo != nil {
		t.Error("First action should not complete a combo")
	}
	
	// Check that combo state was created
	comboState := am.GetActiveComboState(participantID)
	if comboState == nil {
		t.Fatal("No combo state created")
	}
	
	if comboState.Type != COMBO_STUN_ATTACK {
		t.Errorf("Expected combo type %v, got %v", COMBO_STUN_ATTACK, comboState.Type)
	}
	
	if len(comboState.ActionsCompleted) != 1 {
		t.Errorf("Expected 1 action completed, got %d", len(comboState.ActionsCompleted))
	}
	
	if comboState.ActionsCompleted[0] != ACTION_STUN {
		t.Errorf("Expected ACTION_STUN, got %v", comboState.ActionsCompleted[0])
	}
}

func TestComboCompletion(t *testing.T) {
	am := NewAbilityManager()
	participantID := "test_participant"
	
	// Start combo with ACTION_STUN
	_, err := am.TrackComboAction(participantID, ACTION_STUN)
	if err != nil {
		t.Fatalf("Starting combo failed: %v", err)
	}
	
	// Complete combo with ACTION_ATTACK
	combo, err := am.TrackComboAction(participantID, ACTION_ATTACK)
	if err != nil {
		t.Fatalf("Completing combo failed: %v", err)
	}
	
	if combo == nil {
		t.Fatal("Combo should be completed")
	}
	
	if combo.Type != COMBO_STUN_ATTACK {
		t.Errorf("Expected combo type %v, got %v", COMBO_STUN_ATTACK, combo.Type)
	}
	
	// Check that combo state was cleared
	comboState := am.GetActiveComboState(participantID)
	if comboState != nil {
		t.Error("Combo state should be cleared after completion")
	}
}

func TestComboInterruption(t *testing.T) {
	am := NewAbilityManager()
	participantID := "test_participant"
	
	// Start combo with ACTION_STUN (part of COMBO_STUN_ATTACK)
	_, err := am.TrackComboAction(participantID, ACTION_STUN)
	if err != nil {
		t.Fatalf("Starting combo failed: %v", err)
	}
	
	// Interrupt with wrong action (should break combo)
	combo, err := am.TrackComboAction(participantID, ACTION_HEAL)
	if err != ErrComboInterrupted {
		t.Errorf("Expected combo interrupted error, got: %v", err)
	}
	
	if combo != nil {
		t.Error("Combo should not be completed when interrupted")
	}
	
	// Check that combo state was cleared
	comboState := am.GetActiveComboState(participantID)
	if comboState != nil {
		t.Error("Combo state should be cleared after interruption")
	}
}

func TestComboTimeout(t *testing.T) {
	am := NewAbilityManager()
	participantID := "test_participant"
	
	// Start combo
	_, err := am.TrackComboAction(participantID, ACTION_STUN)
	if err != nil {
		t.Fatalf("Starting combo failed: %v", err)
	}
	
	// Verify combo state exists
	comboState := am.GetActiveComboState(participantID)
	if comboState == nil {
		t.Fatal("Combo state should exist after starting")
	}
	
	// Advance turns beyond window duration (window is 2 for stun_attack combo)
	for i := 0; i < 3; i++ {
		am.AdvanceTurn()
	}
	
	// Check that combo state was cleared due to timeout
	comboState = am.GetActiveComboState(participantID)
	if comboState != nil {
		t.Error("Combo state should be cleared after timeout")
	}
	
	// Try to continue combo (should start a new combo since old one timed out)
	combo, err := am.TrackComboAction(participantID, ACTION_ATTACK)
	if err != nil {
		t.Errorf("Unexpected error when starting new combo: %v", err)
	}
	if combo != nil {
		t.Error("Should not complete a combo when starting new one")
	}
}

func TestThreeHitCombo(t *testing.T) {
	am := NewAbilityManager()
	participantID := "test_participant"
	
	// Execute COMBO_CHARGE_BOOST_ATTACK sequence
	actions := []BattleActionType{ACTION_CHARGE, ACTION_BOOST, ACTION_ATTACK}
	
	for i, action := range actions {
		combo, err := am.TrackComboAction(participantID, action)
		
		if err != nil {
			t.Fatalf("Action %d failed: %v", i+1, err)
		}
		
		if i < len(actions)-1 {
			// Not the last action - combo should not be complete
			if combo != nil {
				t.Errorf("Combo should not be complete at action %d", i+1)
			}
		} else {
			// Last action - combo should be complete
			if combo == nil {
				t.Fatal("Combo should be completed")
			}
			if combo.Type != COMBO_CHARGE_BOOST_ATTACK {
				t.Errorf("Expected combo type %v, got %v", COMBO_CHARGE_BOOST_ATTACK, combo.Type)
			}
		}
	}
}

func TestApplyComboBonus(t *testing.T) {
	am := NewAbilityManager()
	
	// Create a base battle result
	result := &BattleResult{
		Success: true,
		Damage:  20.0,
		Healing: 15.0,
		Response: "attacks",
	}
	
	// Create a combo
	combo := &ComboAttack{
		Type:             COMBO_STUN_ATTACK,
		Name:             "Stunning Strike",
		DamageMultiplier: 1.5,
		EffectMultiplier: 1.2,
		BonusEffects:     []string{"guaranteed_hit", "armor_piercing"},
	}
	
	// Apply combo bonus
	enhanced := am.ApplyComboBonus(result, combo)
	
	if enhanced != result {
		t.Error("ApplyComboBonus should modify the result in place")
	}
	
	expectedDamage := 20.0 * 1.5
	if result.Damage != expectedDamage {
		t.Errorf("Expected damage %f, got %f", expectedDamage, result.Damage)
	}
	
	expectedHealing := 15.0 * 1.2
	if result.Healing != expectedHealing {
		t.Errorf("Expected healing %f, got %f", expectedHealing, result.Healing)
	}
	
	if len(result.StatusEffects) != 2 {
		t.Errorf("Expected 2 status effects, got %d", len(result.StatusEffects))
	}
	
	expectedEffects := map[string]bool{"guaranteed_hit": true, "armor_piercing": true}
	for _, effect := range result.StatusEffects {
		if !expectedEffects[effect] {
			t.Errorf("Unexpected status effect: %s", effect)
		}
	}
	
	if !contains(result.Response, combo.Name) {
		t.Error("Response should contain combo name")
	}
}

func TestResetParticipantAbilities(t *testing.T) {
	am := NewAbilityManager()
	participantID := "test_participant"
	am.InitializeParticipantAbilities(participantID, 10)
	
	battleState := &BattleState{
		Participants: map[string]*BattleParticipant{
			participantID: {
				CharacterID: participantID,
				Stats: BattleStats{HP: 100, MaxHP: 100},
			},
		},
	}
	
	// Use abilities with charges
	_, err := am.UseSpecialAbility(participantID, ABILITY_PERFECT_GUARD, battleState)
	if err != nil {
		t.Fatalf("UseSpecialAbility failed: %v", err)
	}
	
	// Start a combo
	_, err = am.TrackComboAction(participantID, ACTION_STUN)
	if err != nil {
		t.Fatalf("TrackComboAction failed: %v", err)
	}
	
	// Reset abilities
	am.ResetParticipantAbilities(participantID)
	
	// Check that charges are restored
	abilities := am.participantAbilities[participantID]
	for _, ability := range abilities {
		if ability.Type == ABILITY_PERFECT_GUARD {
			if ability.ChargesCurrent != ability.ChargesMax {
				t.Errorf("Charges not restored: current=%d, max=%d", 
					ability.ChargesCurrent, ability.ChargesMax)
			}
		}
	}
	
	// Check that combo state is cleared
	comboState := am.GetActiveComboState(participantID)
	if comboState != nil {
		t.Error("Combo state should be cleared after reset")
	}
}

func TestFairnessConstraints(t *testing.T) {
	am := NewAbilityManager()
	
	// Test that special abilities respect damage caps
	result := am.executeSpecialAbility(ABILITY_CRITICAL_STRIKE, "test", &BattleState{})
	
	if result.Damage > BASE_ATTACK_DAMAGE*MAX_DAMAGE_MODIFIER {
		t.Errorf("Critical strike damage %f exceeds fairness cap %f", 
			result.Damage, BASE_ATTACK_DAMAGE*MAX_DAMAGE_MODIFIER)
	}
	
	// Test berserker rage modifier values
	berserkerResult := am.executeSpecialAbility(ABILITY_BERSERKER_RAGE, "test", &BattleState{})
	
	for _, mod := range berserkerResult.ModifiersApplied {
		if mod.Type == MODIFIER_DAMAGE && mod.Value > MAX_DAMAGE_MODIFIER {
			t.Errorf("Berserker damage modifier %f exceeds cap %f", 
				mod.Value, MAX_DAMAGE_MODIFIER)
		}
	}
}

// Helper function for tests
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    s[:len(substr)] == substr || 
		    s[len(s)-len(substr):] == substr ||
		    (len(s) > len(substr) && containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Benchmark tests for performance validation
func BenchmarkSpecialAbilityExecution(b *testing.B) {
	am := NewAbilityManager()
	participantID := "test_participant"
	am.InitializeParticipantAbilities(participantID, 10)
	
	battleState := &BattleState{
		Participants: map[string]*BattleParticipant{
			participantID: {
				CharacterID: participantID,
				Stats: BattleStats{HP: 100, MaxHP: 100},
			},
		},
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// Reset abilities each iteration to avoid cooldown issues
		am.ResetParticipantAbilities(participantID)
		
		_, err := am.UseSpecialAbility(participantID, ABILITY_CRITICAL_STRIKE, battleState)
		if err != nil {
			b.Fatalf("UseSpecialAbility failed: %v", err)
		}
	}
}

func BenchmarkComboTracking(b *testing.B) {
	am := NewAbilityManager()
	participantID := "test_participant"
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// Reset combo state
		am.clearComboState(participantID)
		
		// Execute a two-hit combo
		_, _ = am.TrackComboAction(participantID, ACTION_STUN)
		_, _ = am.TrackComboAction(participantID, ACTION_ATTACK)
	}
}

func BenchmarkComboBonus(b *testing.B) {
	am := NewAbilityManager()
	
	result := &BattleResult{
		Success: true,
		Damage:  20.0,
		Healing: 15.0,
		Response: "attacks",
	}
	
	combo := &ComboAttack{
		Type:             COMBO_STUN_ATTACK,
		Name:             "Stunning Strike",
		DamageMultiplier: 1.5,
		EffectMultiplier: 1.2,
		BonusEffects:     []string{"guaranteed_hit"},
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// Reset result for each iteration
		result.Damage = 20.0
		result.Healing = 15.0
		result.StatusEffects = nil
		result.Response = "attacks"
		
		am.ApplyComboBonus(result, combo)
	}
}