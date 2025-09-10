package battle

import (
	"math/rand"
	"time"
)

// AIDifficulty defines the sophistication level of AI decision making
type AIDifficulty string

const (
	AI_EASY   AIDifficulty = "easy"
	AI_NORMAL AIDifficulty = "normal"
	AI_HARD   AIDifficulty = "hard"
	AI_EXPERT AIDifficulty = "expert"
)

// AIStrategy represents different AI behavioral patterns
type AIStrategy string

const (
	STRATEGY_AGGRESSIVE AIStrategy = "aggressive" // Favors attack actions
	STRATEGY_DEFENSIVE  AIStrategy = "defensive"  // Favors defensive actions
	STRATEGY_BALANCED   AIStrategy = "balanced"   // Mix of offense and defense
	STRATEGY_SUPPORT    AIStrategy = "support"    // Favors healing and buffs
)

// BattleAI handles automated battle decisions for characters
type BattleAI struct {
	characterID  string
	difficulty   AIDifficulty
	strategy     AIStrategy
	giftProvider GiftProvider       // For item integration
	lastActions  []BattleActionType // Track recent actions to avoid repetition
	rng          *rand.Rand         // For deterministic but varied behavior
}

// NewBattleAI creates a new AI instance for a character
func NewBattleAI(characterID string, difficulty AIDifficulty, strategy AIStrategy) *BattleAI {
	return &BattleAI{
		characterID: characterID,
		difficulty:  difficulty,
		strategy:    strategy,
		lastActions: make([]BattleActionType, 0, 3),
		rng:         rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// NewBattleAIWithGifts creates a new AI instance with gift system integration
func NewBattleAIWithGifts(characterID string, difficulty AIDifficulty, strategy AIStrategy, giftProvider GiftProvider) *BattleAI {
	return &BattleAI{
		characterID:  characterID,
		difficulty:   difficulty,
		strategy:     strategy,
		giftProvider: giftProvider,
		lastActions:  make([]BattleActionType, 0, 3),
		rng:          rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// SelectAction chooses the best action for the current battle situation
func (ai *BattleAI) SelectAction(battleState *BattleState, timeRemaining time.Duration) BattleAction {
	// Emergency timeout decision (< 5 seconds)
	if timeRemaining < AI_EMERGENCY_TIMEOUT {
		return ai.selectQuickAction(battleState)
	}

	// Analyze current situation
	threat := ai.assessThreat(battleState)
	opportunity := ai.assessOpportunity(battleState)

	// Select strategy based on situation and AI personality
	action := ai.selectStrategicAction(battleState, threat, opportunity)

	// Enhance action with items if available
	action = ai.enhanceActionWithItem(action)

	// Track action history to avoid repetition
	ai.updateActionHistory(action.Type)

	return action
}

// selectQuickAction chooses a simple action when time is running out
func (ai *BattleAI) selectQuickAction(battleState *BattleState) BattleAction {
	participant := battleState.Participants[ai.characterID]
	if participant == nil {
		return BattleAction{Type: ACTION_ATTACK, ActorID: ai.characterID}
	}

	// Emergency healing if critically low on health
	healthRatio := participant.Stats.HP / participant.Stats.MaxHP
	if healthRatio < 0.3 {
		return BattleAction{Type: ACTION_HEAL, ActorID: ai.characterID}
	}

	// Default to attack
	return BattleAction{Type: ACTION_ATTACK, ActorID: ai.characterID, TargetID: ai.selectTarget(battleState)}
}

// selectStrategicAction chooses action based on strategy and situation analysis
func (ai *BattleAI) selectStrategicAction(battleState *BattleState, threat, opportunity float64) BattleAction {
	participant := battleState.Participants[ai.characterID]
	if participant == nil {
		return ai.selectQuickAction(battleState)
	}

	// Health-based decisions (override strategy if critical)
	healthRatio := participant.Stats.HP / participant.Stats.MaxHP
	if healthRatio < 0.25 {
		return BattleAction{Type: ACTION_HEAL, ActorID: ai.characterID}
	}

	// Strategy-based action selection
	switch ai.strategy {
	case STRATEGY_AGGRESSIVE:
		return ai.selectAggressiveAction(battleState, threat, opportunity)
	case STRATEGY_DEFENSIVE:
		return ai.selectDefensiveAction(battleState, threat, opportunity)
	case STRATEGY_SUPPORT:
		return ai.selectSupportAction(battleState, threat, opportunity)
	default: // STRATEGY_BALANCED
		return ai.selectBalancedAction(battleState, threat, opportunity)
	}
}

// selectAggressiveAction prioritizes offensive actions
func (ai *BattleAI) selectAggressiveAction(battleState *BattleState, threat, opportunity float64) BattleAction {
	target := ai.selectTarget(battleState)

	// High opportunity - go for powerful attacks
	if opportunity > 0.7 {
		if !ai.hasRecentAction(ACTION_CHARGE) {
			return BattleAction{Type: ACTION_CHARGE, ActorID: ai.characterID}
		}
		return BattleAction{Type: ACTION_ATTACK, ActorID: ai.characterID, TargetID: target}
	}

	// Medium threat - use varied attacks
	if threat > 0.5 {
		actions := []BattleActionType{ACTION_ATTACK, ACTION_STUN, ACTION_DRAIN, ACTION_TAUNT}
		actionType := ai.selectRandomFromList(actions)
		return BattleAction{Type: actionType, ActorID: ai.characterID, TargetID: target}
	}

	// Default attack
	return BattleAction{Type: ACTION_ATTACK, ActorID: ai.characterID, TargetID: target}
}

// selectDefensiveAction prioritizes defensive and control actions
func (ai *BattleAI) selectDefensiveAction(battleState *BattleState, threat, opportunity float64) BattleAction {
	// High threat - defensive measures
	if threat > 0.6 {
		if !ai.hasRecentAction(ACTION_SHIELD) {
			return BattleAction{Type: ACTION_SHIELD, ActorID: ai.characterID}
		}
		if !ai.hasRecentAction(ACTION_DEFEND) {
			return BattleAction{Type: ACTION_DEFEND, ActorID: ai.characterID}
		}
		return BattleAction{Type: ACTION_EVADE, ActorID: ai.characterID}
	}

	// Medium threat - control actions
	if threat > 0.3 {
		target := ai.selectTarget(battleState)
		return BattleAction{Type: ACTION_STUN, ActorID: ai.characterID, TargetID: target}
	}

	// Low threat - prepare for counter-attack
	return BattleAction{Type: ACTION_COUNTER, ActorID: ai.characterID}
}

// selectSupportAction prioritizes healing and buffs
func (ai *BattleAI) selectSupportAction(battleState *BattleState, threat, opportunity float64) BattleAction {
	participant := battleState.Participants[ai.characterID]

	// Heal if health is not full
	if participant.Stats.HP < participant.Stats.MaxHP*0.8 {
		return BattleAction{Type: ACTION_HEAL, ActorID: ai.characterID}
	}

	// Boost if not recently used
	if !ai.hasRecentAction(ACTION_BOOST) && opportunity > 0.4 {
		return BattleAction{Type: ACTION_BOOST, ActorID: ai.characterID}
	}

	// Shield if threatened
	if threat > 0.5 && !ai.hasRecentAction(ACTION_SHIELD) {
		return BattleAction{Type: ACTION_SHIELD, ActorID: ai.characterID}
	}

	// Fall back to light attack
	target := ai.selectTarget(battleState)
	return BattleAction{Type: ACTION_DRAIN, ActorID: ai.characterID, TargetID: target}
}

// selectBalancedAction provides a mix of offensive and defensive actions
func (ai *BattleAI) selectBalancedAction(battleState *BattleState, threat, opportunity float64) BattleAction {
	participant := battleState.Participants[ai.characterID]
	healthRatio := participant.Stats.HP / participant.Stats.MaxHP

	// Health management
	if healthRatio < 0.5 && !ai.hasRecentAction(ACTION_HEAL) {
		return BattleAction{Type: ACTION_HEAL, ActorID: ai.characterID}
	}

	// Threat response
	if threat > opportunity {
		if threat > 0.7 {
			return BattleAction{Type: ACTION_DEFEND, ActorID: ai.characterID}
		}
		target := ai.selectTarget(battleState)
		return BattleAction{Type: ACTION_STUN, ActorID: ai.characterID, TargetID: target}
	}

	// Opportunity exploitation
	if opportunity > 0.6 {
		target := ai.selectTarget(battleState)
		if !ai.hasRecentAction(ACTION_BOOST) {
			return BattleAction{Type: ACTION_BOOST, ActorID: ai.characterID}
		}
		return BattleAction{Type: ACTION_ATTACK, ActorID: ai.characterID, TargetID: target}
	}

	// Default balanced action
	target := ai.selectTarget(battleState)
	return BattleAction{Type: ACTION_ATTACK, ActorID: ai.characterID, TargetID: target}
}

// assessThreat evaluates the danger level from opponents
func (ai *BattleAI) assessThreat(battleState *BattleState) float64 {
	participant := battleState.Participants[ai.characterID]
	if participant == nil {
		return 0.5 // Neutral threat if can't assess
	}

	threat := 0.0

	// Factor 1: Own health status (lower health = higher perceived threat)
	healthRatio := participant.Stats.HP / participant.Stats.MaxHP
	threat += (1.0 - healthRatio) * 0.4

	// Factor 2: Opponent health and attack potential
	for id, opponent := range battleState.Participants {
		if id == ai.characterID {
			continue
		}

		opponentHealthRatio := opponent.Stats.HP / opponent.Stats.MaxHP
		opponentThreat := opponentHealthRatio * 0.3

		// Check for dangerous modifiers on opponent
		for _, modifier := range opponent.Stats.Modifiers {
			if modifier.Type == MODIFIER_DAMAGE {
				opponentThreat += 0.2
			}
		}

		threat += opponentThreat
	}

	// Factor 3: Negative modifiers on self
	for _, modifier := range participant.Stats.Modifiers {
		if modifier.Type == MODIFIER_STUN || modifier.Type == MODIFIER_DEFENSE {
			threat += 0.1 * float64(modifier.Duration)
		}
	}

	return ai.clampValue(threat, 0.0, 1.0)
}

// assessOpportunity evaluates the potential for successful offensive actions
func (ai *BattleAI) assessOpportunity(battleState *BattleState) float64 {
	participant := battleState.Participants[ai.characterID]
	if participant == nil {
		return 0.5
	}

	opportunity := 0.0

	// Factor 1: Own combat readiness
	healthRatio := participant.Stats.HP / participant.Stats.MaxHP
	opportunity += healthRatio * 0.3

	// Factor 2: Beneficial modifiers on self
	for _, modifier := range participant.Stats.Modifiers {
		if modifier.Type == MODIFIER_DAMAGE || modifier.Type == MODIFIER_SHIELD {
			opportunity += 0.2
		}
	}

	// Factor 3: Opponent vulnerability
	for id, opponent := range battleState.Participants {
		if id == ai.characterID {
			continue
		}

		opponentHealthRatio := opponent.Stats.HP / opponent.Stats.MaxHP
		vulnerability := (1.0 - opponentHealthRatio) * 0.3

		// Check for debuffs on opponent
		for _, modifier := range opponent.Stats.Modifiers {
			if modifier.Type == MODIFIER_STUN {
				vulnerability += 0.3
			}
		}

		opportunity += vulnerability
	}

	return ai.clampValue(opportunity, 0.0, 1.0)
}

// selectTarget chooses the best target for offensive actions
func (ai *BattleAI) selectTarget(battleState *BattleState) string {
	var bestTarget string
	var lowestHP float64 = 1000000 // Start with high value

	// Target opponent with lowest HP (focus fire strategy)
	for id, participant := range battleState.Participants {
		if id == ai.characterID {
			continue
		}
		if participant.Stats.HP < lowestHP && participant.Stats.HP > 0 {
			lowestHP = participant.Stats.HP
			bestTarget = id
		}
	}

	return bestTarget
}

// hasRecentAction checks if an action was used recently to avoid repetition
func (ai *BattleAI) hasRecentAction(actionType BattleActionType) bool {
	for _, action := range ai.lastActions {
		if action == actionType {
			return true
		}
	}
	return false
}

// updateActionHistory tracks recent actions for variety
func (ai *BattleAI) updateActionHistory(actionType BattleActionType) {
	ai.lastActions = append(ai.lastActions, actionType)
	if len(ai.lastActions) > 3 {
		ai.lastActions = ai.lastActions[1:] // Keep only last 3 actions
	}
}

// selectRandomFromList chooses a random action from the provided list
func (ai *BattleAI) selectRandomFromList(actions []BattleActionType) BattleActionType {
	if len(actions) == 0 {
		return ACTION_ATTACK
	}
	return actions[ai.rng.Intn(len(actions))]
}

// clampValue ensures a value stays within the specified range
func (ai *BattleAI) clampValue(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// GetDifficulty returns the AI's difficulty level
func (ai *BattleAI) GetDifficulty() AIDifficulty {
	return ai.difficulty
}

// GetStrategy returns the AI's strategic preference
func (ai *BattleAI) GetStrategy() AIStrategy {
	return ai.strategy
}

// SetStrategy updates the AI's strategic behavior
func (ai *BattleAI) SetStrategy(strategy AIStrategy) {
	ai.strategy = strategy
}

// ShouldActImmediately determines if AI should act without waiting for timeout
func (ai *BattleAI) ShouldActImmediately(battleState *BattleState) bool {
	participant := battleState.Participants[ai.characterID]
	if participant == nil {
		return true
	}

	// Act immediately if health is critical
	healthRatio := participant.Stats.HP / participant.Stats.MaxHP
	if healthRatio < 0.2 {
		return true
	}

	// Act immediately if stunned opponent (high opportunity)
	for id, opponent := range battleState.Participants {
		if id == ai.characterID {
			continue
		}
		for _, modifier := range opponent.Stats.Modifiers {
			if modifier.Type == MODIFIER_STUN {
				return true
			}
		}
	}

	// Based on difficulty, some AIs act faster
	switch ai.difficulty {
	case AI_EASY:
		return ai.rng.Float64() < 0.1 // 10% chance to act immediately
	case AI_NORMAL:
		return ai.rng.Float64() < 0.2 // 20% chance
	case AI_HARD:
		return ai.rng.Float64() < 0.4 // 40% chance
	case AI_EXPERT:
		return ai.rng.Float64() < 0.6 // 60% chance
	}

	return false
}

// selectBestItem chooses the most beneficial item for the given action type
func (ai *BattleAI) selectBestItem(actionType BattleActionType) string {
	if ai.giftProvider == nil {
		return "" // No gift system available
	}

	availableGifts := ai.giftProvider.GetAvailableGifts()
	if len(availableGifts) == 0 {
		return ""
	}

	bestItem := ""
	bestScore := 0.0

	for _, gift := range availableGifts {
		score := ai.calculateItemScore(gift, actionType)
		if score > bestScore {
			bestItem = gift.ID
			bestScore = score
		}
	}

	return bestItem
}

// calculateItemScore evaluates how beneficial an item is for the given action
func (ai *BattleAI) calculateItemScore(gift GiftDefinition, actionType BattleActionType) float64 {
	effect := gift.BattleEffect

	// Skip if item doesn't apply to this action type
	if effect.ActionType != "" && effect.ActionType != string(actionType) {
		return 0.0
	}

	score := 0.0

	// Score based on relevant modifiers for the action type
	switch actionType {
	case ACTION_ATTACK, ACTION_DRAIN:
		score += (effect.DamageModifier - 1.0) * 100 // Convert multiplier to score

	case ACTION_HEAL:
		score += (effect.HealModifier - 1.0) * 100

	case ACTION_DEFEND, ACTION_SHIELD:
		score += (effect.DefenseModifier - 1.0) * 100

	default:
		// For other actions, consider speed boost
		score += (effect.SpeedModifier - 1.0) * 50
	}

	// Bonus for longer duration effects
	if effect.Duration > 1 {
		score += float64(effect.Duration) * 10
	}

	// Penalty for consumable items (based on AI difficulty)
	if effect.Consumable {
		switch ai.difficulty {
		case AI_EASY:
			score *= 0.5 // Easy AI is reluctant to use consumables
		case AI_NORMAL:
			score *= 0.7
		case AI_HARD:
			score *= 0.9
		case AI_EXPERT:
			// Expert AI uses consumables strategically, no penalty
		}
	}

	return score
}

// enhanceActionWithItem adds the best available item to an action
func (ai *BattleAI) enhanceActionWithItem(action BattleAction) BattleAction {
	// Only consider items on higher difficulties or with specific chance
	useItemChance := 0.0
	switch ai.difficulty {
	case AI_EASY:
		useItemChance = 0.1 // 10% chance to use items
	case AI_NORMAL:
		useItemChance = 0.3 // 30% chance
	case AI_HARD:
		useItemChance = 0.6 // 60% chance
	case AI_EXPERT:
		useItemChance = 0.8 // 80% chance
	}

	if ai.rng.Float64() > useItemChance {
		return action // Don't use item this time
	}

	// Select best item for this action
	bestItem := ai.selectBestItem(action.Type)
	if bestItem != "" {
		action.ItemUsed = bestItem
	}

	return action
}
