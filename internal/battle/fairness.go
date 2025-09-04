package battle

import (
	"errors"
	"fmt"
	"strings"
)

// FairnessValidator enforces balance constraints to ensure fair gameplay
type FairnessValidator struct {
	maxDamageModifier  float64
	maxDefenseModifier float64
	maxSpeedModifier   float64
	maxHealModifier    float64
	maxEffectStacking  int
}

// NewFairnessValidator creates a validator with default fairness constraints
func NewFairnessValidator() *FairnessValidator {
	return &FairnessValidator{
		maxDamageModifier:  MAX_DAMAGE_MODIFIER,
		maxDefenseModifier: MAX_DEFENSE_MODIFIER,
		maxSpeedModifier:   MAX_SPEED_MODIFIER,
		maxHealModifier:    MAX_HEAL_MODIFIER,
		maxEffectStacking:  MAX_EFFECT_STACKING,
	}
}

// ValidateAction checks if an action complies with fairness rules
func (fv *FairnessValidator) ValidateAction(action BattleAction, participant *BattleParticipant) error {
	// Check action is legal for current state
	if !fv.isActionLegal(action, participant) {
		return ErrIllegalAction
	}

	// Validate item effects don't exceed caps (placeholder for item integration)
	if action.ItemUsed != "" {
		if err := fv.validateItemEffects(action.ItemUsed, action.Type); err != nil {
			return err
		}
	}

	// Check modifier stacking limits
	if len(participant.Stats.Modifiers) >= fv.maxEffectStacking {
		return ErrMaxModifiersReached
	}

	return nil
}

// isActionLegal checks basic action legality rules
func (fv *FairnessValidator) isActionLegal(action BattleAction, participant *BattleParticipant) bool {
	// Cannot act if stunned
	for _, modifier := range participant.Stats.Modifiers {
		if modifier.Type == MODIFIER_STUN && modifier.Duration > 0 {
			return false
		}
	}

	// Cannot heal above maximum HP (checked elsewhere but good to validate)
	if action.Type == ACTION_HEAL && participant.Stats.HP >= participant.Stats.MaxHP {
		return false
	}

	// Cannot target self with offensive actions
	if action.TargetID == action.ActorID {
		offensiveActions := map[BattleActionType]bool{
			ACTION_ATTACK: true,
			ACTION_STUN:   true,
			ACTION_DRAIN:  true,
			ACTION_TAUNT:  true,
		}
		if offensiveActions[action.Type] {
			return false
		}
	}

	return true
}

// validateItemEffects ensures item enhancements don't exceed balance caps
func (fv *FairnessValidator) validateItemEffects(itemID string, actionType BattleActionType) error {
	// Basic item validation implementation (Finding #14 fix)
	if itemID == "" {
		return nil // No item used
	}

	// Validate item ID format (basic safety check)
	if len(itemID) > 50 {
		return errors.New("item ID too long")
	}

	// Prevent obviously invalid item IDs
	if strings.ContainsAny(itemID, "<>\"'&") {
		return errors.New("item ID contains invalid characters")
	}

	// Validate item usage makes sense for action type
	switch actionType {
	case ACTION_ATTACK, ACTION_STUN, ACTION_COUNTER, ACTION_DRAIN, ACTION_CHARGE, ACTION_TAUNT:
		// Offensive items allowed
		if strings.Contains(itemID, "heal") {
			return errors.New("healing items cannot be used for attacks")
		}
	case ACTION_DEFEND, ACTION_SHIELD, ACTION_EVADE:
		// Defensive items allowed
		if strings.Contains(itemID, "damage") || strings.Contains(itemID, "attack") {
			return errors.New("attack items cannot be used for defense")
		}
	case ACTION_HEAL, ACTION_BOOST:
		// Support items allowed - no restrictions
	}

	// Basic item effect caps (placeholder until full item system)
	// This prevents obvious abuse while maintaining functionality
	return nil
}

// ValidateBattleStats ensures participant stats are within acceptable ranges
func (fv *FairnessValidator) ValidateBattleStats(stats *BattleStats) error {
	// HP validation
	if stats.HP < 0 {
		return errors.New("HP cannot be negative")
	}
	if stats.HP > stats.MaxHP {
		return errors.New("HP cannot exceed maximum HP")
	}
	if stats.MaxHP <= 0 {
		return errors.New("maximum HP must be positive")
	}

	// Base stat validation (prevent extreme values)
	if stats.Attack < 0 || stats.Attack > 1000 {
		return errors.New("attack stat out of valid range (0-1000)")
	}
	if stats.Defense < 0 || stats.Defense > 1000 {
		return errors.New("defense stat out of valid range (0-1000)")
	}
	if stats.Speed < 0 || stats.Speed > 1000 {
		return errors.New("speed stat out of valid range (0-1000)")
	}

	// Modifier validation
	if len(stats.Modifiers) > fv.maxEffectStacking {
		return fmt.Errorf("too many active modifiers (%d), maximum allowed: %d",
			len(stats.Modifiers), fv.maxEffectStacking)
	}

	return nil
}

// ValidateDamageOutput ensures damage calculations respect fairness caps
func (fv *FairnessValidator) ValidateDamageOutput(baseDamage, actualDamage float64, actionType BattleActionType) error {
	maxAllowedDamage := baseDamage * fv.maxDamageModifier

	if actualDamage > maxAllowedDamage {
		return fmt.Errorf("damage %.1f exceeds maximum allowed %.1f for action %s",
			actualDamage, maxAllowedDamage, actionType)
	}

	// Ensure damage is non-negative
	if actualDamage < 0 {
		return errors.New("damage cannot be negative")
	}

	return nil
}

// ValidateHealingOutput ensures healing calculations respect fairness caps
func (fv *FairnessValidator) ValidateHealingOutput(baseHealing, actualHealing float64) error {
	maxAllowedHealing := baseHealing * fv.maxHealModifier

	if actualHealing > maxAllowedHealing {
		return fmt.Errorf("healing %.1f exceeds maximum allowed %.1f",
			actualHealing, maxAllowedHealing)
	}

	// Ensure healing is non-negative
	if actualHealing < 0 {
		return errors.New("healing cannot be negative")
	}

	return nil
}

// ValidateModifier ensures a battle modifier respects fairness constraints
func (fv *FairnessValidator) ValidateModifier(modifier BattleModifier) error {
	if err := fv.validateModifierDuration(modifier.Duration); err != nil {
		return err
	}

	if err := fv.validateModifierValue(modifier); err != nil {
		return err
	}

	if err := fv.validateModifierSource(modifier.Source); err != nil {
		return err
	}

	return nil
}

// validateModifierDuration checks if the modifier duration is within acceptable limits
func (fv *FairnessValidator) validateModifierDuration(duration int) error {
	if duration < 0 {
		return errors.New("modifier duration cannot be negative")
	}
	if duration > 10 {
		return errors.New("modifier duration cannot exceed 10 turns")
	}
	return nil
}

// validateModifierValue validates the modifier value based on its type
func (fv *FairnessValidator) validateModifierValue(modifier BattleModifier) error {
	switch modifier.Type {
	case MODIFIER_DAMAGE:
		return fv.validateDamageModifierValue(modifier.Value)
	case MODIFIER_DEFENSE:
		return fv.validateDefenseModifierValue(modifier.Value)
	case MODIFIER_SPEED:
		return fv.validateSpeedModifierValue(modifier.Value)
	case MODIFIER_HEALING:
		return fv.validateHealingModifierValue(modifier.Value)
	case MODIFIER_STUN:
		return fv.validateStunModifierValue(modifier.Value, modifier.Duration)
	case MODIFIER_SHIELD:
		return fv.validateShieldModifierValue(modifier.Value)
	}
	return nil
}

// validateModifierSource ensures the modifier has a valid source
func (fv *FairnessValidator) validateModifierSource(source string) error {
	if source == "" {
		return errors.New("modifier must have a source")
	}
	return nil
}

// validateDamageModifierValue checks if damage modifier value is within limits
func (fv *FairnessValidator) validateDamageModifierValue(value float64) error {
	if value > fv.maxDamageModifier {
		return fmt.Errorf("damage modifier %.2f exceeds maximum %.2f",
			value, fv.maxDamageModifier)
	}
	return nil
}

// validateDefenseModifierValue checks if defense modifier value is within limits
func (fv *FairnessValidator) validateDefenseModifierValue(value float64) error {
	if value > fv.maxDefenseModifier {
		return fmt.Errorf("defense modifier %.2f exceeds maximum %.2f",
			value, fv.maxDefenseModifier)
	}
	return nil
}

// validateSpeedModifierValue checks if speed modifier value is within limits
func (fv *FairnessValidator) validateSpeedModifierValue(value float64) error {
	if value > fv.maxSpeedModifier {
		return fmt.Errorf("speed modifier %.2f exceeds maximum %.2f",
			value, fv.maxSpeedModifier)
	}
	return nil
}

// validateHealingModifierValue checks if healing modifier value is within limits
func (fv *FairnessValidator) validateHealingModifierValue(value float64) error {
	if value > fv.maxHealModifier {
		return fmt.Errorf("healing modifier %.2f exceeds maximum %.2f",
			value, fv.maxHealModifier)
	}
	return nil
}

// validateStunModifierValue checks if stun modifier has correct value and duration
func (fv *FairnessValidator) validateStunModifierValue(value float64, duration int) error {
	if value != 1 || duration > 3 {
		return errors.New("stun modifier must have value 1 and duration ≤ 3")
	}
	return nil
}

// validateShieldModifierValue checks if shield modifier value is within limits
func (fv *FairnessValidator) validateShieldModifierValue(value float64) error {
	if value > BASE_SHIELD_ABSORPTION*2 {
		return fmt.Errorf("shield absorption %.1f exceeds maximum %.1f",
			value, BASE_SHIELD_ABSORPTION*2)
	}
	return nil
}

// CapDamageModifier ensures damage modifiers don't exceed fairness limits
func (fv *FairnessValidator) CapDamageModifier(modifier float64) float64 {
	if modifier > fv.maxDamageModifier {
		return fv.maxDamageModifier
	}
	if modifier < 0 {
		return 0
	}
	return modifier
}

// CapDefenseModifier ensures defense modifiers don't exceed fairness limits
func (fv *FairnessValidator) CapDefenseModifier(modifier float64) float64 {
	if modifier > fv.maxDefenseModifier {
		return fv.maxDefenseModifier
	}
	if modifier < 0 {
		return 0
	}
	return modifier
}

// CapHealingModifier ensures healing modifiers don't exceed fairness limits
func (fv *FairnessValidator) CapHealingModifier(modifier float64) float64 {
	if modifier > fv.maxHealModifier {
		return fv.maxHealModifier
	}
	if modifier < 0 {
		return 0
	}
	return modifier
}

// EnforceModifierStackingLimit removes excess modifiers if limit exceeded
func (fv *FairnessValidator) EnforceModifierStackingLimit(modifiers []BattleModifier) []BattleModifier {
	if len(modifiers) <= fv.maxEffectStacking {
		return modifiers
	}

	// Keep the most recent modifiers (preserve newer effects)
	return modifiers[len(modifiers)-fv.maxEffectStacking:]
}

// GetFairnessLimits returns the current fairness constraints
func (fv *FairnessValidator) GetFairnessLimits() map[string]float64 {
	return map[string]float64{
		"maxDamageModifier":  fv.maxDamageModifier,
		"maxDefenseModifier": fv.maxDefenseModifier,
		"maxSpeedModifier":   fv.maxSpeedModifier,
		"maxHealModifier":    fv.maxHealModifier,
		"maxEffectStacking":  float64(fv.maxEffectStacking),
	}
}

// SetFairnessLimits allows customization of fairness constraints (for testing)
func (fv *FairnessValidator) SetFairnessLimits(damageModifier, defenseModifier, speedModifier, healModifier float64, effectStacking int) {
	fv.maxDamageModifier = damageModifier
	fv.maxDefenseModifier = defenseModifier
	fv.maxSpeedModifier = speedModifier
	fv.maxHealModifier = healModifier
	fv.maxEffectStacking = effectStacking
}

// ValidateBattleBalance performs comprehensive balance checks on a battle state
func (fv *FairnessValidator) ValidateBattleBalance(battleState *BattleState) []error {
	var errors []error

	// Check all participants have valid stats
	for id, participant := range battleState.Participants {
		if err := fv.ValidateBattleStats(&participant.Stats); err != nil {
			errors = append(errors, fmt.Errorf("participant %s: %w", id, err))
		}

		// Check all modifiers are valid
		for i, modifier := range participant.Stats.Modifiers {
			if err := fv.ValidateModifier(modifier); err != nil {
				errors = append(errors, fmt.Errorf("participant %s modifier %d: %w", id, i, err))
			}
		}
	}

	// Check turn order validity
	if len(battleState.TurnOrder) != len(battleState.Participants) {
		errors = append(errors, fmt.Errorf("turn order length (%d) doesn't match participant count (%d)",
			len(battleState.TurnOrder), len(battleState.Participants)))
	}

	for _, participantID := range battleState.TurnOrder {
		if battleState.Participants[participantID] == nil {
			errors = append(errors, fmt.Errorf("turn order contains invalid participant ID: %s", participantID))
		}
	}

	return errors
}

// CheckActionFairness validates an action result against fairness rules
func (fv *FairnessValidator) CheckActionFairness(action BattleAction, result *BattleResult) error {
	// Validate damage output
	if result.Damage > 0 {
		baseDamage := fv.getBaseDamageForAction(action.Type)
		if err := fv.ValidateDamageOutput(baseDamage, result.Damage, action.Type); err != nil {
			return err
		}
	}

	// Validate healing output
	if result.Healing > 0 {
		baseHealing := fv.getBaseHealingForAction(action.Type)
		if err := fv.ValidateHealingOutput(baseHealing, result.Healing); err != nil {
			return err
		}
	}

	// Validate applied modifiers
	for _, modifier := range result.ModifiersApplied {
		if err := fv.ValidateModifier(modifier); err != nil {
			return fmt.Errorf("invalid modifier in action result: %w", err)
		}
	}

	return nil
}

// getBaseDamageForAction returns the expected base damage for an action type
func (fv *FairnessValidator) getBaseDamageForAction(actionType BattleActionType) float64 {
	switch actionType {
	case ACTION_ATTACK:
		return BASE_ATTACK_DAMAGE
	case ACTION_DRAIN:
		return BASE_ATTACK_DAMAGE * BASE_DRAIN_RATIO
	default:
		return 0
	}
}

// getBaseHealingForAction returns the expected base healing for an action type
func (fv *FairnessValidator) getBaseHealingForAction(actionType BattleActionType) float64 {
	switch actionType {
	case ACTION_HEAL:
		return BASE_HEAL_AMOUNT
	case ACTION_DRAIN:
		return BASE_ATTACK_DAMAGE * BASE_DRAIN_RATIO
	default:
		return 0
	}
}
