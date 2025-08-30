package battle

import (
	"math"
	"time"
)

// PerformAction executes a battle action and returns the result
func (bm *BattleManager) PerformAction(action BattleAction, targetID string) (*BattleResult, error) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if bm.currentBattle == nil || bm.currentBattle.Phase != PHASE_ACTIVE {
		return nil, ErrBattleNotActive
	}

	// Set target and timestamp
	action.TargetID = targetID
	action.Timestamp = time.Now()

	// Process the action through the fairness pipeline
	result, err := bm.processActionPipeline(action)
	if err != nil {
		return nil, err
	}

	// Store the action result
	action.Result = result
	bm.currentBattle.LastAction = &action

	// Update participant action history
	if participant := bm.currentBattle.Participants[action.ActorID]; participant != nil {
		participant.ActionHistory = append(participant.ActionHistory, action)
		participant.LastActionTime = time.Now()
	}

	return result, nil
}

// processActionPipeline implements the turn resolution pipeline from the design
func (bm *BattleManager) processActionPipeline(action BattleAction) (*BattleResult, error) {
	// 1. Validate action legality
	if err := bm.validateAction(action); err != nil {
		return nil, err
	}

	// 2. Apply item modifiers with caps (placeholder for future item integration)
	modifiedAction := bm.applyItemModifiers(action)

	// 3. Calculate base effect
	baseResult := bm.calculateBaseEffect(modifiedAction)

	// 4. Apply fairness constraints
	cappedResult := bm.applyFairnessCaps(baseResult)

	// 5. Execute effect on target
	finalResult := bm.executeEffect(modifiedAction, cappedResult)

	// 6. Advance turn order (will be implemented in future PR)
	bm.advanceTurn()

	return finalResult, nil
}

// validateAction checks if an action is legal in the current battle state
func (bm *BattleManager) validateAction(action BattleAction) error {
	// Check if actor exists and is a participant
	actor := bm.currentBattle.Participants[action.ActorID]
	if actor == nil {
		return ErrInvalidParticipant
	}

	// Check if target exists (for targeted actions)
	if action.TargetID != "" && action.TargetID != action.ActorID {
		if bm.currentBattle.Participants[action.TargetID] == nil {
			return ErrInvalidParticipant
		}
	}

	// Check if actor is stunned (cannot act)
	for _, modifier := range actor.Stats.Modifiers {
		if modifier.Type == MODIFIER_STUN && modifier.Duration > 0 {
			return ErrActionNotAllowed
		}
	}

	// Check if actor has reached maximum modifiers for buff actions
	if bm.isBuffAction(action.Type) && len(actor.Stats.Modifiers) >= MAX_EFFECT_STACKING {
		return ErrMaxModifiersReached
	}

	return nil
}

// applyItemModifiers applies item effects to actions (placeholder for item system integration)
func (bm *BattleManager) applyItemModifiers(action BattleAction) BattleAction {
	// This will be extended when integrating with the gift/item system
	// For now, return the action unchanged
	return action
}

// calculateBaseEffect computes the base effect of an action before modifiers
func (bm *BattleManager) calculateBaseEffect(action BattleAction) *BattleResult {
	result := &BattleResult{
		Success:   true,
		Animation: bm.getActionAnimation(action.Type),
		Response:  bm.getActionResponse(action.Type),
	}

	switch action.Type {
	case ACTION_ATTACK:
		result.Damage = BASE_ATTACK_DAMAGE

	case ACTION_DEFEND:
		// Defend adds a defense modifier
		result.ModifiersApplied = []BattleModifier{
			{
				Type:     MODIFIER_DEFENSE,
				Value:    BASE_DEFEND_REDUCTION,
				Duration: 1,
				Source:   "defend_action",
			},
		}

	case ACTION_HEAL:
		result.Healing = BASE_HEAL_AMOUNT

	case ACTION_STUN:
		result.ModifiersApplied = []BattleModifier{
			{
				Type:     MODIFIER_STUN,
				Value:    1,
				Duration: BASE_STUN_DURATION,
				Source:   "stun_action",
			},
		}

	case ACTION_BOOST:
		result.ModifiersApplied = []BattleModifier{
			{
				Type:     MODIFIER_DAMAGE,
				Value:    BASE_BOOST_AMOUNT,
				Duration: 3,
				Source:   "boost_action",
			},
		}

	case ACTION_DRAIN:
		result.Damage = BASE_ATTACK_DAMAGE * BASE_DRAIN_RATIO
		result.Healing = result.Damage

	case ACTION_SHIELD:
		result.ModifiersApplied = []BattleModifier{
			{
				Type:     MODIFIER_SHIELD,
				Value:    BASE_SHIELD_ABSORPTION,
				Duration: 3,
				Source:   "shield_action",
			},
		}

	case ACTION_CHARGE:
		result.ModifiersApplied = []BattleModifier{
			{
				Type:     MODIFIER_DAMAGE,
				Value:    BASE_CHARGE_MULTIPLIER,
				Duration: 1,
				Source:   "charge_action",
			},
		}

	case ACTION_COUNTER, ACTION_EVADE, ACTION_TAUNT:
		// Special actions with situational effects
		result.ModifiersApplied = []BattleModifier{
			{
				Type:     ModifierType(string(action.Type)),
				Value:    1,
				Duration: 2,
				Source:   string(action.Type) + "_action",
			},
		}

	default:
		result.Success = false
		result.Response = "Unknown action"
	}

	return result
}

// applyFairnessCaps enforces maximum effect limits to maintain balance
func (bm *BattleManager) applyFairnessCaps(result *BattleResult) *BattleResult {
	// Cap damage modifications
	if result.Damage > BASE_ATTACK_DAMAGE*MAX_DAMAGE_MODIFIER {
		result.Damage = BASE_ATTACK_DAMAGE * MAX_DAMAGE_MODIFIER
	}

	// Cap healing modifications
	if result.Healing > BASE_HEAL_AMOUNT*MAX_HEAL_MODIFIER {
		result.Healing = BASE_HEAL_AMOUNT * MAX_HEAL_MODIFIER
	}

	// Validate modifier stacking
	if len(result.ModifiersApplied) > MAX_EFFECT_STACKING {
		result.ModifiersApplied = result.ModifiersApplied[:MAX_EFFECT_STACKING]
	}

	return result
}

// executeEffect applies the calculated effects to the target participant
func (bm *BattleManager) executeEffect(action BattleAction, result *BattleResult) *BattleResult {
	// Get target (default to actor for self-targeting actions)
	targetID := action.TargetID
	if targetID == "" {
		targetID = action.ActorID
	}

	target := bm.currentBattle.Participants[targetID]
	if target == nil {
		result.Success = false
		result.Response = "Target not found"
		return result
	}

	// Apply damage
	if result.Damage > 0 {
		damage := bm.calculateActualDamage(result.Damage, target)
		target.Stats.HP = math.Max(0, target.Stats.HP-damage)
		result.Damage = damage
	}

	// Apply healing
	if result.Healing > 0 {
		healing := math.Min(result.Healing, target.Stats.MaxHP-target.Stats.HP)
		target.Stats.HP = math.Min(target.Stats.MaxHP, target.Stats.HP+healing)
		result.Healing = healing
	}

	// Apply modifiers
	for _, modifier := range result.ModifiersApplied {
		bm.applyModifierToTarget(modifier, target)
	}

	// Update turn-based modifiers (reduce duration)
	bm.updateModifierDurations(target)

	return result
}

// calculateActualDamage computes damage after defense modifiers
func (bm *BattleManager) calculateActualDamage(baseDamage float64, target *BattleParticipant) float64 {
	damage := baseDamage

	// Apply defense modifiers
	for _, modifier := range target.Stats.Modifiers {
		if modifier.Type == MODIFIER_DEFENSE && modifier.Duration > 0 {
			damage *= (1 - modifier.Value)
		}
		if modifier.Type == MODIFIER_SHIELD && modifier.Duration > 0 {
			absorbed := math.Min(damage, modifier.Value)
			damage -= absorbed
			// Reduce shield strength (this would update the modifier)
		}
	}

	return math.Max(0, damage)
}

// applyModifierToTarget adds a modifier to the target's active effects
func (bm *BattleManager) applyModifierToTarget(modifier BattleModifier, target *BattleParticipant) {
	// Check for existing modifiers of the same type from the same source
	for i, existing := range target.Stats.Modifiers {
		if existing.Type == modifier.Type && existing.Source == modifier.Source {
			// Replace existing modifier
			target.Stats.Modifiers[i] = modifier
			return
		}
	}

	// Add new modifier if under the stacking limit
	if len(target.Stats.Modifiers) < MAX_EFFECT_STACKING {
		target.Stats.Modifiers = append(target.Stats.Modifiers, modifier)
	}
}

// updateModifierDurations reduces duration of active modifiers and removes expired ones
func (bm *BattleManager) updateModifierDurations(participant *BattleParticipant) {
	activeModifiers := make([]BattleModifier, 0, len(participant.Stats.Modifiers))

	for _, modifier := range participant.Stats.Modifiers {
		modifier.Duration--
		if modifier.Duration > 0 {
			activeModifiers = append(activeModifiers, modifier)
		}
	}

	participant.Stats.Modifiers = activeModifiers
}

// advanceTurn moves to the next participant in turn order
func (bm *BattleManager) advanceTurn() {
	if len(bm.currentBattle.TurnOrder) == 0 {
		return
	}

	bm.currentBattle.CurrentTurn = (bm.currentBattle.CurrentTurn + 1) % len(bm.currentBattle.TurnOrder)
}

// isBuffAction determines if an action provides beneficial effects
func (bm *BattleManager) isBuffAction(actionType BattleActionType) bool {
	buffActions := map[BattleActionType]bool{
		ACTION_HEAL:   true,
		ACTION_BOOST:  true,
		ACTION_SHIELD: true,
		ACTION_CHARGE: true,
	}
	return buffActions[actionType]
}

// getActionAnimation returns the animation name for a battle action
func (bm *BattleManager) getActionAnimation(actionType BattleActionType) string {
	// Maps to the GIF files defined in the plan
	animationMap := map[BattleActionType]string{
		ACTION_ATTACK:  "attack",
		ACTION_DEFEND:  "defend",
		ACTION_STUN:    "stun",
		ACTION_HEAL:    "heal",
		ACTION_BOOST:   "boost",
		ACTION_COUNTER: "counter",
		ACTION_DRAIN:   "drain",
		ACTION_SHIELD:  "shield",
		ACTION_CHARGE:  "charge",
		ACTION_EVADE:   "evade",
		ACTION_TAUNT:   "taunt",
	}

	if animation, exists := animationMap[actionType]; exists {
		return animation
	}
	return "idle" // Fallback
}

// getActionResponse returns a text response for a battle action
func (bm *BattleManager) getActionResponse(actionType BattleActionType) string {
	responseMap := map[BattleActionType]string{
		ACTION_ATTACK:  "attacks with determination!",
		ACTION_DEFEND:  "takes a defensive stance!",
		ACTION_STUN:    "attempts to stun the opponent!",
		ACTION_HEAL:    "recovers health!",
		ACTION_BOOST:   "powers up for increased damage!",
		ACTION_COUNTER: "prepares a counter-attack!",
		ACTION_DRAIN:   "drains energy from the opponent!",
		ACTION_SHIELD:  "creates a protective barrier!",
		ACTION_CHARGE:  "charges up energy!",
		ACTION_EVADE:   "prepares to dodge!",
		ACTION_TAUNT:   "taunts the opponent!",
	}

	if response, exists := responseMap[actionType]; exists {
		return response
	}
	return "takes an action!"
}

// GetCurrentTurnParticipant returns the participant whose turn it is
func (bm *BattleManager) GetCurrentTurnParticipant() string {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	if bm.currentBattle == nil || len(bm.currentBattle.TurnOrder) == 0 {
		return ""
	}

	return bm.currentBattle.TurnOrder[bm.currentBattle.CurrentTurn]
}

// AddParticipant adds a character to the current battle
func (bm *BattleManager) AddParticipant(characterID, peerID string, isLocal bool, initialStats BattleStats) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if bm.currentBattle == nil {
		return ErrBattleNotActive
	}

	participant := &BattleParticipant{
		CharacterID:    characterID,
		PeerID:         peerID,
		IsLocal:        isLocal,
		Stats:          initialStats,
		ActiveItems:    make([]string, 0),
		ActionHistory:  make([]BattleAction, 0),
		LastActionTime: time.Now(),
		IsReady:        true,
	}

	bm.currentBattle.Participants[characterID] = participant
	bm.currentBattle.TurnOrder = append(bm.currentBattle.TurnOrder, characterID)

	return nil
}

// IsParticipantDefeated checks if a participant has been defeated (HP <= 0)
func (bm *BattleManager) IsParticipantDefeated(characterID string) bool {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	if participant := bm.currentBattle.Participants[characterID]; participant != nil {
		return participant.Stats.HP <= 0
	}
	return false
}

// GetWinner returns the winner if battle is finished, empty string if ongoing
func (bm *BattleManager) GetWinner() string {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	if bm.currentBattle == nil || bm.currentBattle.Phase != PHASE_FINISHED {
		return ""
	}

	var activeParticipants []string
	for id, participant := range bm.currentBattle.Participants {
		if participant.Stats.HP > 0 {
			activeParticipants = append(activeParticipants, id)
		}
	}

	if len(activeParticipants) == 1 {
		return activeParticipants[0]
	}

	return "" // Draw or battle not concluded
}
