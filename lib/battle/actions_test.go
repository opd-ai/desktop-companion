package battle

import (
	"testing"
)

// TestBattleManager_PerformAction tests action execution
func TestBattleManager_PerformAction(t *testing.T) {
	bm := NewBattleManager()
	bm.InitiateBattle("opponent_1")

	stats := BattleStats{HP: 100, MaxHP: 100, Attack: 20, Defense: 15, Speed: 10}
	bm.AddParticipant("char_1", "", true, stats)
	bm.AddParticipant("char_2", "", false, stats)

	// Set battle to active
	bm.mu.Lock()
	bm.currentBattle.Phase = PHASE_ACTIVE
	bm.mu.Unlock()

	action := BattleAction{
		Type:    ACTION_ATTACK,
		ActorID: "char_1",
	}

	result, err := bm.PerformAction(action, "char_2")
	if err != nil {
		t.Fatalf("Failed to perform action: %v", err)
	}

	if !result.Success {
		t.Error("Action should have succeeded")
	}
	if result.Damage != BASE_ATTACK_DAMAGE {
		t.Errorf("Expected damage %.1f, got %.1f", BASE_ATTACK_DAMAGE, result.Damage)
	}
}

// TestBattleManager_PerformAction_NoBattle tests error when no battle active
func TestBattleManager_PerformAction_NoBattle(t *testing.T) {
	bm := NewBattleManager()

	action := BattleAction{Type: ACTION_ATTACK, ActorID: "char_1"}
	_, err := bm.PerformAction(action, "char_2")
	if err != ErrBattleNotActive {
		t.Errorf("Expected ErrBattleNotActive, got %v", err)
	}
}

// TestBattleActions_Attack tests attack action calculation
func TestBattleActions_Attack(t *testing.T) {
	bm := NewBattleManager()
	bm.InitiateBattle("test")
	bm.mu.Lock()
	bm.currentBattle.Phase = PHASE_ACTIVE
	bm.mu.Unlock()

	action := BattleAction{Type: ACTION_ATTACK, ActorID: "char_1"}
	result := bm.calculateBaseEffect(action)

	if !result.Success {
		t.Error("Attack should succeed")
	}
	if result.Damage != BASE_ATTACK_DAMAGE {
		t.Errorf("Expected damage %.1f, got %.1f", BASE_ATTACK_DAMAGE, result.Damage)
	}
	if result.Animation != "attack" {
		t.Errorf("Expected animation 'attack', got '%s'", result.Animation)
	}
}

// TestBattleActions_Heal tests heal action calculation
func TestBattleActions_Heal(t *testing.T) {
	bm := NewBattleManager()
	bm.InitiateBattle("test")

	action := BattleAction{Type: ACTION_HEAL, ActorID: "char_1"}
	result := bm.calculateBaseEffect(action)

	if !result.Success {
		t.Error("Heal should succeed")
	}
	if result.Healing != BASE_HEAL_AMOUNT {
		t.Errorf("Expected healing %.1f, got %.1f", BASE_HEAL_AMOUNT, result.Healing)
	}
	if result.Animation != "heal" {
		t.Errorf("Expected animation 'heal', got '%s'", result.Animation)
	}
}

// TestBattleActions_Defend tests defend action calculation
func TestBattleActions_Defend(t *testing.T) {
	bm := NewBattleManager()
	bm.InitiateBattle("test")

	action := BattleAction{Type: ACTION_DEFEND, ActorID: "char_1"}
	result := bm.calculateBaseEffect(action)

	if !result.Success {
		t.Error("Defend should succeed")
	}
	if len(result.ModifiersApplied) != 1 {
		t.Errorf("Expected 1 modifier, got %d", len(result.ModifiersApplied))
	}

	modifier := result.ModifiersApplied[0]
	if modifier.Type != MODIFIER_DEFENSE {
		t.Errorf("Expected defense modifier, got %s", modifier.Type)
	}
	if modifier.Value != BASE_DEFEND_REDUCTION {
		t.Errorf("Expected value %.1f, got %.1f", BASE_DEFEND_REDUCTION, modifier.Value)
	}
}

// TestBattleActions_Stun tests stun action calculation
func TestBattleActions_Stun(t *testing.T) {
	bm := NewBattleManager()
	bm.InitiateBattle("test")

	action := BattleAction{Type: ACTION_STUN, ActorID: "char_1"}
	result := bm.calculateBaseEffect(action)

	if !result.Success {
		t.Error("Stun should succeed")
	}
	if len(result.ModifiersApplied) != 1 {
		t.Errorf("Expected 1 modifier, got %d", len(result.ModifiersApplied))
	}

	modifier := result.ModifiersApplied[0]
	if modifier.Type != MODIFIER_STUN {
		t.Errorf("Expected stun modifier, got %s", modifier.Type)
	}
	if modifier.Duration != BASE_STUN_DURATION {
		t.Errorf("Expected duration %d, got %d", BASE_STUN_DURATION, modifier.Duration)
	}
}

// TestBattleActions_Drain tests drain action calculation
func TestBattleActions_Drain(t *testing.T) {
	bm := NewBattleManager()
	bm.InitiateBattle("test")

	action := BattleAction{Type: ACTION_DRAIN, ActorID: "char_1"}
	result := bm.calculateBaseEffect(action)

	expectedDamage := BASE_ATTACK_DAMAGE * BASE_DRAIN_RATIO
	if result.Damage != expectedDamage {
		t.Errorf("Expected damage %.1f, got %.1f", expectedDamage, result.Damage)
	}
	if result.Healing != expectedDamage {
		t.Errorf("Expected healing %.1f, got %.1f", expectedDamage, result.Healing)
	}
}

// TestValidateAction tests action validation
func TestValidateAction(t *testing.T) {
	bm := NewBattleManager()
	bm.InitiateBattle("test")

	stats := BattleStats{HP: 100, MaxHP: 100}
	bm.AddParticipant("char_1", "", true, stats)
	bm.AddParticipant("char_2", "", false, stats)

	bm.mu.Lock()
	bm.currentBattle.Phase = PHASE_ACTIVE
	bm.mu.Unlock()

	// Valid action
	action := BattleAction{Type: ACTION_ATTACK, ActorID: "char_1", TargetID: "char_2"}
	err := bm.validateAction(action)
	if err != nil {
		t.Errorf("Valid action should not return error: %v", err)
	}

	// Invalid actor
	action.ActorID = "nonexistent"
	err = bm.validateAction(action)
	if err != ErrInvalidParticipant {
		t.Errorf("Expected ErrInvalidParticipant, got %v", err)
	}

	// Invalid target
	action.ActorID = "char_1"
	action.TargetID = "nonexistent"
	err = bm.validateAction(action)
	if err != ErrInvalidParticipant {
		t.Errorf("Expected ErrInvalidParticipant, got %v", err)
	}
}

// TestValidateAction_Stunned tests validation with stunned participant
func TestValidateAction_Stunned(t *testing.T) {
	bm := NewBattleManager()
	bm.InitiateBattle("test")

	stats := BattleStats{
		HP:    100,
		MaxHP: 100,
		Modifiers: []BattleModifier{
			{Type: MODIFIER_STUN, Duration: 1},
		},
	}
	bm.AddParticipant("char_1", "", true, stats)

	bm.mu.Lock()
	bm.currentBattle.Phase = PHASE_ACTIVE
	bm.mu.Unlock()

	action := BattleAction{Type: ACTION_ATTACK, ActorID: "char_1"}
	err := bm.validateAction(action)
	if err != ErrActionNotAllowed {
		t.Errorf("Expected ErrActionNotAllowed for stunned participant, got %v", err)
	}
}

// TestApplyFairnessCaps tests fairness constraint enforcement
func TestApplyFairnessCaps(t *testing.T) {
	bm := NewBattleManager()

	// Test damage capping
	result := &BattleResult{
		Damage:  BASE_ATTACK_DAMAGE * 2, // Exceeds cap
		Healing: BASE_HEAL_AMOUNT * 2,   // Exceeds cap
	}

	capped := bm.applyFairnessCaps(result)

	expectedMaxDamage := BASE_ATTACK_DAMAGE * MAX_DAMAGE_MODIFIER
	if capped.Damage != expectedMaxDamage {
		t.Errorf("Expected capped damage %.1f, got %.1f", expectedMaxDamage, capped.Damage)
	}

	expectedMaxHealing := BASE_HEAL_AMOUNT * MAX_HEAL_MODIFIER
	if capped.Healing != expectedMaxHealing {
		t.Errorf("Expected capped healing %.1f, got %.1f", expectedMaxHealing, capped.Healing)
	}
}

// TestExecuteEffect tests effect application to targets
func TestExecuteEffect(t *testing.T) {
	bm := NewBattleManager()
	bm.InitiateBattle("test")

	stats := BattleStats{HP: 100, MaxHP: 100}
	bm.AddParticipant("char_1", "", true, stats)
	bm.AddParticipant("char_2", "", false, stats)

	action := BattleAction{Type: ACTION_ATTACK, ActorID: "char_1", TargetID: "char_2"}
	result := &BattleResult{
		Success: true,
		Damage:  20,
	}

	finalResult := bm.executeEffect(action, result)

	// Check target took damage
	target := bm.currentBattle.Participants["char_2"]
	expectedHP := 100 - 20
	if target.Stats.HP != float64(expectedHP) {
		t.Errorf("Expected target HP %.1f, got %.1f", float64(expectedHP), target.Stats.HP)
	}

	// Check result reports actual damage dealt
	if finalResult.Damage != 20 {
		t.Errorf("Expected result damage 20, got %.1f", finalResult.Damage)
	}
}

// TestExecuteEffect_Healing tests healing effect application
func TestExecuteEffect_Healing(t *testing.T) {
	bm := NewBattleManager()
	bm.InitiateBattle("test")

	stats := BattleStats{HP: 50, MaxHP: 100} // Damaged character
	bm.AddParticipant("char_1", "", true, stats)

	action := BattleAction{Type: ACTION_HEAL, ActorID: "char_1", TargetID: "char_1"}
	result := &BattleResult{
		Success: true,
		Healing: 25,
	}

	finalResult := bm.executeEffect(action, result)

	// Check character was healed
	character := bm.currentBattle.Participants["char_1"]
	expectedHP := 75.0
	if character.Stats.HP != expectedHP {
		t.Errorf("Expected character HP %.1f, got %.1f", expectedHP, character.Stats.HP)
	}

	// Check result reports actual healing done
	if finalResult.Healing != 25 {
		t.Errorf("Expected result healing 25, got %.1f", finalResult.Healing)
	}
}

// TestExecuteEffect_HealingCap tests healing doesn't exceed max HP
func TestExecuteEffect_HealingCap(t *testing.T) {
	bm := NewBattleManager()
	bm.InitiateBattle("test")

	stats := BattleStats{HP: 90, MaxHP: 100}
	bm.AddParticipant("char_1", "", true, stats)

	action := BattleAction{Type: ACTION_HEAL, ActorID: "char_1", TargetID: "char_1"}
	result := &BattleResult{
		Success: true,
		Healing: 25, // Would exceed max HP
	}

	finalResult := bm.executeEffect(action, result)

	// Check HP capped at maximum
	character := bm.currentBattle.Participants["char_1"]
	if character.Stats.HP != 100 {
		t.Errorf("Expected character HP 100, got %.1f", character.Stats.HP)
	}

	// Check result reports actual healing done (10, not 25)
	if finalResult.Healing != 10 {
		t.Errorf("Expected result healing 10, got %.1f", finalResult.Healing)
	}
}

// TestCalculateActualDamage tests damage calculation with modifiers
func TestCalculateActualDamage(t *testing.T) {
	bm := NewBattleManager()

	// Target with defense modifier
	target := &BattleParticipant{
		Stats: BattleStats{
			HP:    100,
			MaxHP: 100,
			Modifiers: []BattleModifier{
				{Type: MODIFIER_DEFENSE, Value: 0.5, Duration: 1}, // 50% damage reduction
			},
		},
	}

	damage := bm.calculateActualDamage(20, target)
	expectedDamage := 20 * (1 - 0.5) // 10
	if damage != expectedDamage {
		t.Errorf("Expected damage %.1f, got %.1f", expectedDamage, damage)
	}
}

// TestUpdateModifierDurations tests modifier duration management
func TestUpdateModifierDurations(t *testing.T) {
	bm := NewBattleManager()

	participant := &BattleParticipant{
		Stats: BattleStats{
			Modifiers: []BattleModifier{
				{Type: MODIFIER_DAMAGE, Duration: 3},
				{Type: MODIFIER_DEFENSE, Duration: 2},
				{Type: MODIFIER_STUN, Duration: 1}, // Will become 0 and be removed
			},
		},
	}

	bm.updateModifierDurations(participant)

	// Should have 2 modifiers left (durations 2 and 1)
	if len(participant.Stats.Modifiers) != 2 {
		t.Errorf("Expected 2 modifiers after update, got %d", len(participant.Stats.Modifiers))
	}

	// Check remaining modifiers have reduced duration
	for _, modifier := range participant.Stats.Modifiers {
		if modifier.Type == MODIFIER_DAMAGE && modifier.Duration != 2 {
			t.Errorf("Expected damage modifier duration 2, got %d", modifier.Duration)
		}
		if modifier.Type == MODIFIER_DEFENSE && modifier.Duration != 1 {
			t.Errorf("Expected defense modifier duration 1, got %d", modifier.Duration)
		}
	}
}

// TestAdvanceTurn tests turn order progression
func TestAdvanceTurn(t *testing.T) {
	bm := NewBattleManager()
	bm.InitiateBattle("test")

	// Add participants to create turn order
	stats := BattleStats{HP: 100, MaxHP: 100}
	bm.AddParticipant("char_1", "", true, stats)
	bm.AddParticipant("char_2", "", false, stats)

	// Should start at turn 0 (char_1)
	current := bm.GetCurrentTurnParticipant()
	if current != "char_1" {
		t.Errorf("Expected 'char_1' first, got '%s'", current)
	}

	// Advance turn
	bm.advanceTurn()
	current = bm.GetCurrentTurnParticipant()
	if current != "char_2" {
		t.Errorf("Expected 'char_2' after advance, got '%s'", current)
	}

	// Advance again (should wrap to char_1)
	bm.advanceTurn()
	current = bm.GetCurrentTurnParticipant()
	if current != "char_1" {
		t.Errorf("Expected 'char_1' after wrap, got '%s'", current)
	}
}

// TestActionHistory tests action history tracking
func TestActionHistory(t *testing.T) {
	bm := NewBattleManager()
	bm.InitiateBattle("test")

	stats := BattleStats{HP: 100, MaxHP: 100}
	bm.AddParticipant("char_1", "", true, stats)
	bm.AddParticipant("char_2", "", false, stats)

	bm.mu.Lock()
	bm.currentBattle.Phase = PHASE_ACTIVE
	bm.mu.Unlock()

	action := BattleAction{Type: ACTION_ATTACK, ActorID: "char_1"}
	_, err := bm.PerformAction(action, "char_2")
	if err != nil {
		t.Fatalf("Failed to perform action: %v", err)
	}

	// Check action was recorded in history
	participant := bm.currentBattle.Participants["char_1"]
	if len(participant.ActionHistory) != 1 {
		t.Errorf("Expected 1 action in history, got %d", len(participant.ActionHistory))
	}

	historyAction := participant.ActionHistory[0]
	if historyAction.Type != ACTION_ATTACK {
		t.Errorf("Expected ACTION_ATTACK in history, got %s", historyAction.Type)
	}
	if historyAction.ActorID != "char_1" {
		t.Errorf("Expected actor 'char_1' in history, got '%s'", historyAction.ActorID)
	}
}
