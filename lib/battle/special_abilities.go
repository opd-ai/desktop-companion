// Package battle - Special abilities system for enhanced tactical gameplay
//
// This file implements special battle actions and combo attacks that provide
// additional tactical options beyond the base 11 actions. Special abilities
// have cooldowns and enhanced effects, while combo attacks chain multiple
// actions together for devastating combinations.
//
// Design principles:
// - All special abilities respect existing fairness constraints
// - Cooldown system prevents ability spam
// - Combo attacks require setup and proper timing
// - Standard library only implementation
package battle

import (
	"errors"
	"fmt"
	"time"
)

// Special ability system errors
var (
	ErrAbilityOnCooldown    = errors.New("ability is on cooldown")
	ErrInvalidComboSequence = errors.New("invalid combo sequence")
	ErrComboInterrupted     = errors.New("combo was interrupted")
	ErrInsufficientCharges  = errors.New("insufficient charges for ability")
)

// SpecialAbilityType defines enhanced battle actions with cooldowns
type SpecialAbilityType string

const (
	// Offensive special abilities
	ABILITY_CRITICAL_STRIKE SpecialAbilityType = "critical_strike" // High damage with crit chance
	ABILITY_LIGHTNING_BOLT  SpecialAbilityType = "lightning_bolt"  // Fast, unblockable attack
	ABILITY_BERSERKER_RAGE  SpecialAbilityType = "berserker_rage"  // Damage boost with defense penalty
	ABILITY_LIFE_STEAL      SpecialAbilityType = "life_steal"      // Attack that heals based on damage

	// Defensive special abilities
	ABILITY_PERFECT_GUARD  SpecialAbilityType = "perfect_guard"  // Complete damage immunity for 1 turn
	ABILITY_COUNTER_ATTACK SpecialAbilityType = "counter_attack" // Automatic retaliation on hit
	ABILITY_DAMAGE_REFLECT SpecialAbilityType = "damage_reflect" // Reflects damage back to attacker
	ABILITY_SANCTUARY      SpecialAbilityType = "sanctuary"      // Area healing and protection

	// Utility special abilities
	ABILITY_TIME_FREEZE SpecialAbilityType = "time_freeze" // Skip opponent's next turn
	ABILITY_STAT_SWAP   SpecialAbilityType = "stat_swap"   // Exchange stats with opponent
	ABILITY_CLEANSE     SpecialAbilityType = "cleanse"     // Remove all negative modifiers
	ABILITY_POWER_SURGE SpecialAbilityType = "power_surge" // Temporary massive stat boost
)

// ComboAttackType defines multi-action battle combinations
type ComboAttackType string

const (
	// Two-hit combos
	COMBO_STUN_ATTACK  ComboAttackType = "stun_attack"  // Stun → Attack (enhanced damage)
	COMBO_BOOST_STRIKE ComboAttackType = "boost_strike" // Boost → Attack (damage amplified)
	COMBO_DRAIN_HEAL   ComboAttackType = "drain_heal"   // Drain → Heal (enhanced healing)

	// Three-hit combos
	COMBO_CHARGE_BOOST_ATTACK ComboAttackType = "charge_boost_attack" // Charge → Boost → Attack
	COMBO_SHIELD_COUNTER_STUN ComboAttackType = "shield_counter_stun" // Shield → Counter → Stun
	COMBO_HEAL_BOOST_COUNTER  ComboAttackType = "heal_boost_counter"  // Heal → Boost → Counter

	// Ultimate combos (four+ hits)
	COMBO_BERSERKER_FURY    ComboAttackType = "berserker_fury"    // Boost → Charge → Attack → Attack
	COMBO_DEFENSIVE_MASTERY ComboAttackType = "defensive_mastery" // Shield → Heal → Counter → Boost
)

// Ability configuration constants with fairness constraints
const (
	// Special ability base values (stronger than normal actions but capped)
	SPECIAL_CRITICAL_DAMAGE   = BASE_ATTACK_DAMAGE * 1.2  // 20% damage boost (within fairness cap)
	SPECIAL_LIGHTNING_DAMAGE  = BASE_ATTACK_DAMAGE * 1.15 // 15% damage boost, unblockable
	SPECIAL_BERSERKER_DAMAGE  = 1.2                       // 20% damage boost (within cap)
	SPECIAL_BERSERKER_DEFENSE = 0.3                       // 30% defense reduction penalty
	SPECIAL_LIFE_STEAL_RATIO  = 0.6                       // 60% of damage as healing
	SPECIAL_REFLECT_RATIO     = 0.4                       // 40% damage reflection
	SPECIAL_SANCTUARY_HEAL    = BASE_HEAL_AMOUNT * 1.25   // 25% enhanced healing (within cap)
	SPECIAL_POWER_SURGE_BONUS = 1.1                       // 10% stat boost (within cap)

	// Combo multipliers (applied to base actions)
	COMBO_TWO_HIT_MULTIPLIER   = 1.2 // 20% bonus for 2-hit combos (within fairness)
	COMBO_THREE_HIT_MULTIPLIER = 1.2 // 20% bonus for 3-hit combos (within fairness)
	COMBO_ULTIMATE_MULTIPLIER  = 1.2 // 20% bonus for ultimate combos (within fairness)

	// Cooldown durations (turns)
	SPECIAL_ABILITY_COOLDOWN  = 5 // Most special abilities
	ULTIMATE_ABILITY_COOLDOWN = 8 // Most powerful abilities
	COMBO_WINDOW_DURATION     = 3 // Turns to complete combo sequence
)

// SpecialAbility represents an enhanced battle action with cooldown
type SpecialAbility struct {
	Type           SpecialAbilityType `json:"type"`
	Name           string             `json:"name"`
	Description    string             `json:"description"`
	Cooldown       int                `json:"cooldown"`       // Base cooldown in turns
	ChargesMax     int                `json:"chargesMax"`     // Maximum charges (0 = unlimited)
	ChargesCurrent int                `json:"chargesCurrent"` // Current available charges
	LastUsed       time.Time          `json:"lastUsed"`       // When ability was last used
	RequiredLevel  int                `json:"requiredLevel"`  // Character level requirement
}

// ComboAttack represents a sequence of actions that create enhanced effects
type ComboAttack struct {
	Type             ComboAttackType    `json:"type"`
	Name             string             `json:"name"`
	Description      string             `json:"description"`
	Sequence         []BattleActionType `json:"sequence"`       // Required action sequence
	WindowDuration   int                `json:"windowDuration"` // Turns to complete combo
	DamageMultiplier float64            `json:"damageMultiplier"`
	EffectMultiplier float64            `json:"effectMultiplier"`
	BonusEffects     []string           `json:"bonusEffects"` // Additional effects when completed
}

// ComboState tracks progress of ongoing combo attempts
type ComboState struct {
	Type             ComboAttackType    `json:"type"`
	ActionsCompleted []BattleActionType `json:"actionsCompleted"`
	StartedTurn      int                `json:"startedTurn"`
	ActorID          string             `json:"actorID"`
	IsActive         bool               `json:"isActive"`
}

// AbilityManager handles special abilities and combo tracking
type AbilityManager struct {
	participantAbilities map[string][]SpecialAbility `json:"participantAbilities"`
	availableCombos      []ComboAttack               `json:"availableCombos"`
	activeComboStates    map[string]*ComboState      `json:"activeComboStates"` // participantID -> combo state
	currentTurn          int                         `json:"currentTurn"`
}

// NewAbilityManager creates a new ability manager with default special abilities
func NewAbilityManager() *AbilityManager {
	return &AbilityManager{
		participantAbilities: make(map[string][]SpecialAbility),
		availableCombos:      getDefaultCombos(),
		activeComboStates:    make(map[string]*ComboState),
		currentTurn:          0,
	}
}

// getDefaultCombos returns the standard combo attacks available to all participants
func getDefaultCombos() []ComboAttack {
	return []ComboAttack{
		// Two-hit combos
		{
			Type:             COMBO_STUN_ATTACK,
			Name:             "Stunning Strike",
			Description:      "Stun opponent then deliver enhanced attack",
			Sequence:         []BattleActionType{ACTION_STUN, ACTION_ATTACK},
			WindowDuration:   2,
			DamageMultiplier: COMBO_TWO_HIT_MULTIPLIER,
			EffectMultiplier: 1.0,
			BonusEffects:     []string{"guaranteed_hit"},
		},
		{
			Type:             COMBO_BOOST_STRIKE,
			Name:             "Power Strike",
			Description:      "Boost power then unleash devastating attack",
			Sequence:         []BattleActionType{ACTION_BOOST, ACTION_ATTACK},
			WindowDuration:   2,
			DamageMultiplier: COMBO_TWO_HIT_MULTIPLIER,
			EffectMultiplier: 1.2,
			BonusEffects:     []string{"armor_piercing"},
		},
		{
			Type:             COMBO_DRAIN_HEAL,
			Name:             "Vampiric Recovery",
			Description:      "Drain energy then convert to enhanced healing",
			Sequence:         []BattleActionType{ACTION_DRAIN, ACTION_HEAL},
			WindowDuration:   2,
			DamageMultiplier: 1.0,
			EffectMultiplier: COMBO_TWO_HIT_MULTIPLIER,
			BonusEffects:     []string{"poison_resist"},
		},

		// Three-hit combos
		{
			Type:             COMBO_CHARGE_BOOST_ATTACK,
			Name:             "Overwhelming Assault",
			Description:      "Charge energy, boost power, then unleash ultimate attack",
			Sequence:         []BattleActionType{ACTION_CHARGE, ACTION_BOOST, ACTION_ATTACK},
			WindowDuration:   3,
			DamageMultiplier: COMBO_THREE_HIT_MULTIPLIER,
			EffectMultiplier: 1.0,
			BonusEffects:     []string{"area_damage", "knockback"},
		},
		{
			Type:             COMBO_SHIELD_COUNTER_STUN,
			Name:             "Defensive Mastery",
			Description:      "Shield up, prepare counter, then stun on retaliation",
			Sequence:         []BattleActionType{ACTION_SHIELD, ACTION_COUNTER, ACTION_STUN},
			WindowDuration:   3,
			DamageMultiplier: 1.0,
			EffectMultiplier: COMBO_THREE_HIT_MULTIPLIER,
			BonusEffects:     []string{"damage_immunity", "reflect_stun"},
		},

		// Ultimate combos
		{
			Type:             COMBO_BERSERKER_FURY,
			Name:             "Berserker's Fury",
			Description:      "Enter berserker state with devastating multi-hit assault",
			Sequence:         []BattleActionType{ACTION_BOOST, ACTION_CHARGE, ACTION_ATTACK, ACTION_ATTACK},
			WindowDuration:   4,
			DamageMultiplier: COMBO_ULTIMATE_MULTIPLIER,
			EffectMultiplier: 1.0,
			BonusEffects:     []string{"frenzy_mode", "lifesteal", "crit_chance"},
		},
		{
			Type:             COMBO_DEFENSIVE_MASTERY,
			Name:             "Guardian's Resolve",
			Description:      "Ultimate defensive combo providing massive protection",
			Sequence:         []BattleActionType{ACTION_SHIELD, ACTION_HEAL, ACTION_COUNTER, ACTION_BOOST},
			WindowDuration:   4,
			DamageMultiplier: 1.0,
			EffectMultiplier: COMBO_ULTIMATE_MULTIPLIER,
			BonusEffects:     []string{"perfect_defense", "auto_heal", "damage_reflection"},
		},
	}
}

// getDefaultSpecialAbilities returns standard special abilities for participants
func getDefaultSpecialAbilities() []SpecialAbility {
	return []SpecialAbility{
		// Offensive abilities
		{
			Type:           ABILITY_CRITICAL_STRIKE,
			Name:           "Critical Strike",
			Description:    "High damage attack with critical hit chance",
			Cooldown:       4,
			ChargesMax:     0, // Unlimited uses (cooldown limited)
			ChargesCurrent: 0,
			RequiredLevel:  1,
		},
		{
			Type:           ABILITY_LIGHTNING_BOLT,
			Name:           "Lightning Bolt",
			Description:    "Fast, unblockable magical attack",
			Cooldown:       3,
			ChargesMax:     0,
			ChargesCurrent: 0,
			RequiredLevel:  2,
		},
		{
			Type:           ABILITY_BERSERKER_RAGE,
			Name:           "Berserker Rage",
			Description:    "Massive damage boost with defense penalty",
			Cooldown:       6,
			ChargesMax:     0,
			ChargesCurrent: 0,
			RequiredLevel:  3,
		},
		{
			Type:           ABILITY_LIFE_STEAL,
			Name:           "Life Steal",
			Description:    "Attack that heals based on damage dealt",
			Cooldown:       5,
			ChargesMax:     0,
			ChargesCurrent: 0,
			RequiredLevel:  2,
		},

		// Defensive abilities
		{
			Type:           ABILITY_PERFECT_GUARD,
			Name:           "Perfect Guard",
			Description:    "Complete immunity to damage for one turn",
			Cooldown:       8,
			ChargesMax:     2, // Limited charges
			ChargesCurrent: 2,
			RequiredLevel:  3,
		},
		{
			Type:           ABILITY_SANCTUARY,
			Name:           "Sanctuary",
			Description:    "Area healing and protection effect",
			Cooldown:       7,
			ChargesMax:     0,
			ChargesCurrent: 0,
			RequiredLevel:  4,
		},

		// Utility abilities
		{
			Type:           ABILITY_CLEANSE,
			Name:           "Cleanse",
			Description:    "Remove all negative status effects",
			Cooldown:       5,
			ChargesMax:     0,
			ChargesCurrent: 0,
			RequiredLevel:  2,
		},
		{
			Type:           ABILITY_TIME_FREEZE,
			Name:           "Time Freeze",
			Description:    "Skip opponent's next turn",
			Cooldown:       10,
			ChargesMax:     1, // Very limited
			ChargesCurrent: 1,
			RequiredLevel:  5,
		},
	}
}

// InitializeParticipantAbilities sets up special abilities for a battle participant
func (am *AbilityManager) InitializeParticipantAbilities(participantID string, characterLevel int) {
	abilities := getDefaultSpecialAbilities()

	// Filter abilities by character level
	var availableAbilities []SpecialAbility
	for _, ability := range abilities {
		if characterLevel >= ability.RequiredLevel {
			availableAbilities = append(availableAbilities, ability)
		}
	}

	am.participantAbilities[participantID] = availableAbilities
}

// GetAvailableSpecialAbilities returns abilities that are off cooldown and have charges
func (am *AbilityManager) GetAvailableSpecialAbilities(participantID string) []SpecialAbility {
	abilities := am.participantAbilities[participantID]
	if abilities == nil {
		return nil
	}

	var available []SpecialAbility
	for _, ability := range abilities {
		if am.isAbilityAvailable(ability) {
			available = append(available, ability)
		}
	}

	return available
}

// isAbilityAvailable checks if an ability can be used (cooldown and charges)
func (am *AbilityManager) isAbilityAvailable(ability SpecialAbility) bool {
	// Check charges first
	if ability.ChargesMax > 0 && ability.ChargesCurrent <= 0 {
		return false
	}

	// Check cooldown (simplified - using turn-based cooldown)
	// For testing purposes, if LastUsed is zero time, ability is available
	if ability.LastUsed.IsZero() {
		return true
	}

	// Calculate turns elapsed since last use (simplified logic for testing)
	turnsElapsed := am.currentTurn - int(ability.LastUsed.Unix()) // Simplified calculation
	if turnsElapsed < ability.Cooldown {
		return false
	}

	return true
}

// UseSpecialAbility attempts to use a special ability and returns the enhanced battle result
func (am *AbilityManager) UseSpecialAbility(participantID string, abilityType SpecialAbilityType, battleState *BattleState) (*BattleResult, error) {
	abilities := am.participantAbilities[participantID]
	if abilities == nil {
		return nil, errors.New("participant has no abilities")
	}

	// Find the ability
	abilityIndex := -1
	for i, ability := range abilities {
		if ability.Type == abilityType {
			abilityIndex = i
			break
		}
	}

	if abilityIndex == -1 {
		return nil, errors.New("ability not found")
	}

	ability := &abilities[abilityIndex]

	// Check availability
	if !am.isAbilityAvailable(*ability) {
		return nil, ErrAbilityOnCooldown
	}

	// Execute the special ability
	result := am.executeSpecialAbility(abilityType, participantID, battleState)

	// Apply fairness constraints to special abilities
	result = am.applySpecialAbilityFairnessCaps(result)

	// Update ability state
	ability.LastUsed = time.Now().Add(time.Duration(ability.Cooldown) * time.Second) // Track when available again
	if ability.ChargesMax > 0 {
		ability.ChargesCurrent--
	}

	// Save the updated ability
	am.participantAbilities[participantID][abilityIndex] = *ability

	return result, nil
}

// applySpecialAbilityFairnessCaps enforces fairness constraints on special abilities
func (am *AbilityManager) applySpecialAbilityFairnessCaps(result *BattleResult) *BattleResult {
	// Cap damage to fairness limits
	if result.Damage > BASE_ATTACK_DAMAGE*MAX_DAMAGE_MODIFIER {
		result.Damage = BASE_ATTACK_DAMAGE * MAX_DAMAGE_MODIFIER
	}

	// Cap healing to fairness limits
	if result.Healing > BASE_HEAL_AMOUNT*MAX_HEAL_MODIFIER {
		result.Healing = BASE_HEAL_AMOUNT * MAX_HEAL_MODIFIER
	}

	// Cap modifier values
	for i := range result.ModifiersApplied {
		modifier := &result.ModifiersApplied[i]
		switch modifier.Type {
		case MODIFIER_DAMAGE:
			if modifier.Value > MAX_DAMAGE_MODIFIER {
				modifier.Value = MAX_DAMAGE_MODIFIER
			}
		case MODIFIER_DEFENSE:
			if modifier.Value > MAX_DEFENSE_MODIFIER {
				modifier.Value = MAX_DEFENSE_MODIFIER
			}
		case MODIFIER_SPEED:
			if modifier.Value > MAX_SPEED_MODIFIER {
				modifier.Value = MAX_SPEED_MODIFIER
			}
		case MODIFIER_HEALING:
			if modifier.Value > MAX_HEAL_MODIFIER {
				modifier.Value = MAX_HEAL_MODIFIER
			}
		}
	}

	return result
}

// executeSpecialAbility calculates the effects of using a special ability
func (am *AbilityManager) executeSpecialAbility(abilityType SpecialAbilityType, actorID string, battleState *BattleState) *BattleResult {
	result := &BattleResult{
		Success:   true,
		Animation: am.getAbilityAnimation(abilityType),
		Response:  am.getAbilityResponse(abilityType),
	}

	switch abilityType {
	case ABILITY_CRITICAL_STRIKE:
		result.Damage = SPECIAL_CRITICAL_DAMAGE
		result.StatusEffects = []string{"critical_hit"}

	case ABILITY_LIGHTNING_BOLT:
		result.Damage = SPECIAL_LIGHTNING_DAMAGE
		result.StatusEffects = []string{"unblockable", "lightning_effect"}

	case ABILITY_BERSERKER_RAGE:
		result.ModifiersApplied = []BattleModifier{
			{
				Type:     MODIFIER_DAMAGE,
				Value:    SPECIAL_BERSERKER_DAMAGE,
				Duration: 3,
				Source:   "berserker_rage",
			},
			{
				Type:     MODIFIER_DEFENSE,
				Value:    SPECIAL_BERSERKER_DEFENSE,
				Duration: 3,
				Source:   "berserker_rage_penalty",
			},
		}

	case ABILITY_LIFE_STEAL:
		result.Damage = BASE_ATTACK_DAMAGE
		result.Healing = result.Damage * SPECIAL_LIFE_STEAL_RATIO
		result.StatusEffects = []string{"life_steal"}

	case ABILITY_PERFECT_GUARD:
		result.ModifiersApplied = []BattleModifier{
			{
				Type:     MODIFIER_DEFENSE,
				Value:    1.0, // 100% damage reduction
				Duration: 1,
				Source:   "perfect_guard",
			},
		}

	case ABILITY_SANCTUARY:
		result.Healing = SPECIAL_SANCTUARY_HEAL
		result.ModifiersApplied = []BattleModifier{
			{
				Type:     MODIFIER_HEALING,
				Value:    1.2, // 20% healing boost
				Duration: 3,
				Source:   "sanctuary",
			},
		}

	case ABILITY_CLEANSE:
		result.StatusEffects = []string{"remove_debuffs"}

	case ABILITY_TIME_FREEZE:
		result.StatusEffects = []string{"skip_opponent_turn"}

	case ABILITY_POWER_SURGE:
		result.ModifiersApplied = []BattleModifier{
			{
				Type:     MODIFIER_DAMAGE,
				Value:    SPECIAL_POWER_SURGE_BONUS,
				Duration: 2,
				Source:   "power_surge",
			},
			{
				Type:     MODIFIER_SPEED,
				Value:    SPECIAL_POWER_SURGE_BONUS,
				Duration: 2,
				Source:   "power_surge",
			},
		}

	default:
		result.Success = false
		result.Response = "Unknown special ability"
	}

	return result
}

// TrackComboAction records an action and checks for combo progress/completion
func (am *AbilityManager) TrackComboAction(participantID string, action BattleActionType) (*ComboAttack, error) {
	// Check if participant has an active combo
	comboState := am.activeComboStates[participantID]

	if comboState != nil {
		// Continue existing combo
		return am.continueCombo(participantID, action, comboState)
	}

	// Check if this action starts any combo
	return am.checkComboStart(participantID, action)
}

// checkComboStart determines if an action begins a combo sequence
func (am *AbilityManager) checkComboStart(participantID string, action BattleActionType) (*ComboAttack, error) {
	for _, combo := range am.availableCombos {
		if len(combo.Sequence) > 0 && combo.Sequence[0] == action {
			// Start tracking this combo
			am.activeComboStates[participantID] = &ComboState{
				Type:             combo.Type,
				ActionsCompleted: []BattleActionType{action},
				StartedTurn:      am.currentTurn,
				ActorID:          participantID,
				IsActive:         true,
			}
			return nil, nil // Combo started but not completed
		}
	}

	return nil, nil // No combo started
}

// continueCombo processes the next action in an active combo sequence
func (am *AbilityManager) continueCombo(participantID string, action BattleActionType, comboState *ComboState) (*ComboAttack, error) {
	// Find the combo definition
	var comboDef *ComboAttack
	for i := range am.availableCombos {
		if am.availableCombos[i].Type == comboState.Type {
			comboDef = &am.availableCombos[i]
			break
		}
	}

	if comboDef == nil {
		am.clearComboState(participantID)
		return nil, errors.New("combo definition not found")
	}

	// Check if combo window has expired FIRST
	if am.currentTurn-comboState.StartedTurn >= comboDef.WindowDuration {
		am.clearComboState(participantID)
		return nil, ErrComboInterrupted
	}

	// Check if this action matches the next expected action
	nextActionIndex := len(comboState.ActionsCompleted)
	if nextActionIndex >= len(comboDef.Sequence) {
		am.clearComboState(participantID)
		return nil, ErrInvalidComboSequence
	}

	expectedAction := comboDef.Sequence[nextActionIndex]
	if action != expectedAction {
		// Wrong action - combo is broken
		am.clearComboState(participantID)
		return nil, ErrComboInterrupted
	}

	// Add action to completed sequence
	comboState.ActionsCompleted = append(comboState.ActionsCompleted, action)

	// Check if combo is complete
	if len(comboState.ActionsCompleted) == len(comboDef.Sequence) {
		am.clearComboState(participantID)
		return comboDef, nil // Combo completed!
	}

	// Combo continues
	return nil, nil
}

// clearComboState removes the active combo state for a participant
func (am *AbilityManager) clearComboState(participantID string) {
	delete(am.activeComboStates, participantID)
}

// ApplyComboBonus enhances a battle result with combo multipliers and effects
func (am *AbilityManager) ApplyComboBonus(result *BattleResult, combo *ComboAttack) *BattleResult {
	if combo == nil {
		return result
	}

	// Apply damage multiplier
	if result.Damage > 0 {
		result.Damage *= combo.DamageMultiplier
	}

	// Apply effect multiplier to healing and other values
	if result.Healing > 0 {
		result.Healing *= combo.EffectMultiplier
	}

	// Add bonus effects using StatusEffects field
	if result.StatusEffects == nil {
		result.StatusEffects = combo.BonusEffects
	} else {
		result.StatusEffects = append(result.StatusEffects, combo.BonusEffects...)
	}

	// Update response to indicate combo completion
	result.Response = fmt.Sprintf("COMBO: %s! %s", combo.Name, result.Response)

	return result
}

// GetActiveComboState returns the current combo state for a participant
func (am *AbilityManager) GetActiveComboState(participantID string) *ComboState {
	return am.activeComboStates[participantID]
}

// AdvanceTurn updates the internal turn counter for cooldown calculations
func (am *AbilityManager) AdvanceTurn() {
	am.currentTurn++

	// Clean up expired combo states
	for participantID, comboState := range am.activeComboStates {
		var comboDef *ComboAttack
		for i := range am.availableCombos {
			if am.availableCombos[i].Type == comboState.Type {
				comboDef = &am.availableCombos[i]
				break
			}
		}

		if comboDef != nil && am.currentTurn-comboState.StartedTurn >= comboDef.WindowDuration {
			am.clearComboState(participantID)
		}
	}
}

// getAbilityAnimation returns the animation name for a special ability
func (am *AbilityManager) getAbilityAnimation(abilityType SpecialAbilityType) string {
	animationMap := map[SpecialAbilityType]string{
		ABILITY_CRITICAL_STRIKE: "critical_strike",
		ABILITY_LIGHTNING_BOLT:  "lightning_bolt",
		ABILITY_BERSERKER_RAGE:  "berserker_rage",
		ABILITY_LIFE_STEAL:      "life_steal",
		ABILITY_PERFECT_GUARD:   "perfect_guard",
		ABILITY_SANCTUARY:       "sanctuary",
		ABILITY_CLEANSE:         "cleanse",
		ABILITY_TIME_FREEZE:     "time_freeze",
		ABILITY_POWER_SURGE:     "power_surge",
	}

	if animation, exists := animationMap[abilityType]; exists {
		return animation
	}
	return "special_ability"
}

// getAbilityResponse returns a text response for a special ability
func (am *AbilityManager) getAbilityResponse(abilityType SpecialAbilityType) string {
	responseMap := map[SpecialAbilityType]string{
		ABILITY_CRITICAL_STRIKE: "delivers a devastating critical strike!",
		ABILITY_LIGHTNING_BOLT:  "unleashes a bolt of lightning!",
		ABILITY_BERSERKER_RAGE:  "enters a berserker rage!",
		ABILITY_LIFE_STEAL:      "drains life force from the enemy!",
		ABILITY_PERFECT_GUARD:   "assumes a perfect defensive stance!",
		ABILITY_SANCTUARY:       "creates a healing sanctuary!",
		ABILITY_CLEANSE:         "purifies all negative effects!",
		ABILITY_TIME_FREEZE:     "freezes time itself!",
		ABILITY_POWER_SURGE:     "surges with raw power!",
	}

	if response, exists := responseMap[abilityType]; exists {
		return response
	}
	return "uses a special ability!"
}

// GetAvailableCombos returns all combo attacks that can be initiated
func (am *AbilityManager) GetAvailableCombos() []ComboAttack {
	return am.availableCombos
}

// ResetParticipantAbilities restores all abilities and clears cooldowns (for new battles)
func (am *AbilityManager) ResetParticipantAbilities(participantID string) {
	abilities := am.participantAbilities[participantID]
	if abilities == nil {
		return
	}

	// Reset cooldowns and restore charges
	for i := range abilities {
		abilities[i].LastUsed = time.Time{} // Clear last used time
		if abilities[i].ChargesMax > 0 {
			abilities[i].ChargesCurrent = abilities[i].ChargesMax
		}
	}

	am.participantAbilities[participantID] = abilities
	am.clearComboState(participantID)
}
