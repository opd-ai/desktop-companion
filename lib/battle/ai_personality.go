package battle

import (
	"errors"
	"math/rand"
	"time"
)

// AIPersonality defines different AI behavioral patterns
type AIPersonality string

const (
	PERSONALITY_AGGRESSIVE AIPersonality = "aggressive"
	PERSONALITY_DEFENSIVE  AIPersonality = "defensive"
	PERSONALITY_BALANCED   AIPersonality = "balanced"
	PERSONALITY_TACTICAL   AIPersonality = "tactical"
)

// AIDecisionWeights defines how much an AI values different strategies
type AIDecisionWeights struct {
	AttackWeight     float64 // Weight for offensive actions
	DefenseWeight    float64 // Weight for defensive actions
	HealingWeight    float64 // Weight for healing actions
	SpecialWeight    float64 // Weight for special abilities
	ComboWeight      float64 // Weight for combo attempts
	RiskTolerance    float64 // Willingness to take risks (0.0 - 1.0)
	PlanningDepth    int     // How many turns ahead AI considers
	SpecialThreshold float64 // HP threshold to prefer special abilities
	ComboAggression  float64 // How aggressively to pursue combos
}

// AIDecision represents a potential AI action with its priority
type AIDecision struct {
	ActionType     BattleActionType
	SpecialAbility SpecialAbilityType
	Priority       float64
	Reasoning      string
	IsSpecial      bool
	StartsCombo    bool
	ContinuesCombo bool
}

// PersonalityBasedAI manages AI decision making based on personality
type PersonalityBasedAI struct {
	personalities map[AIPersonality]AIDecisionWeights
	rand          *rand.Rand
}

// NewPersonalityBasedAI creates a new AI system with predefined personalities
func NewPersonalityBasedAI() *PersonalityBasedAI {
	ai := &PersonalityBasedAI{
		personalities: make(map[AIPersonality]AIDecisionWeights),
		rand:          rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	ai.initializePersonalities()
	return ai
}

// initializePersonalities sets up the default personality types
func (ai *PersonalityBasedAI) initializePersonalities() {
	// Aggressive: Prioritizes offense and special attacks
	ai.personalities[PERSONALITY_AGGRESSIVE] = AIDecisionWeights{
		AttackWeight:     0.8,
		DefenseWeight:    0.2,
		HealingWeight:    0.3,
		SpecialWeight:    0.9, // High preference for special abilities
		ComboWeight:      0.7, // Moderate combo preference
		RiskTolerance:    0.8,
		PlanningDepth:    2,
		SpecialThreshold: 0.8, // Use specials when above 80% HP
		ComboAggression:  0.8, // Aggressive combo attempts
	}

	// Defensive: Prioritizes survival and defensive specials
	ai.personalities[PERSONALITY_DEFENSIVE] = AIDecisionWeights{
		AttackWeight:     0.3,
		DefenseWeight:    0.9,
		HealingWeight:    0.8,
		SpecialWeight:    0.5, // Moderate special preference
		ComboWeight:      0.3, // Low combo preference (too risky)
		RiskTolerance:    0.2,
		PlanningDepth:    4,
		SpecialThreshold: 0.5, // Use specials when above 50% HP
		ComboAggression:  0.2, // Conservative combo attempts
	}

	// Balanced: Equal consideration of all options
	ai.personalities[PERSONALITY_BALANCED] = AIDecisionWeights{
		AttackWeight:     0.6,
		DefenseWeight:    0.6,
		HealingWeight:    0.5,
		SpecialWeight:    0.6, // Balanced special usage
		ComboWeight:      0.5, // Balanced combo usage
		RiskTolerance:    0.5,
		PlanningDepth:    3,
		SpecialThreshold: 0.6, // Use specials when above 60% HP
		ComboAggression:  0.5, // Moderate combo attempts
	}

	// Tactical: Strategic, high planning with smart special usage
	ai.personalities[PERSONALITY_TACTICAL] = AIDecisionWeights{
		AttackWeight:     0.5,
		DefenseWeight:    0.4,
		HealingWeight:    0.4,
		SpecialWeight:    0.7, // Strategic special usage
		ComboWeight:      0.8, // High combo preference (tactical advantage)
		RiskTolerance:    0.4,
		PlanningDepth:    5,
		SpecialThreshold: 0.7, // Use specials when above 70% HP
		ComboAggression:  0.6, // Calculated combo attempts
	}
}

// GetOptimalActionType determines the best action type for an AI based on personality
func (ai *PersonalityBasedAI) GetOptimalActionType(personality AIPersonality, hpPercentage, enemyHPPercentage float64, turnNumber int) (BattleActionType, error) {
	weights, exists := ai.personalities[personality]
	if !exists {
		return "", errors.New("unknown AI personality")
	}

	isLowHP := hpPercentage < 0.3
	isEnemyLowHP := enemyHPPercentage < 0.3

	// Calculate action priorities
	actionPriorities := make(map[BattleActionType]float64)

	// Attack actions
	actionPriorities[ACTION_ATTACK] = weights.AttackWeight
	if isEnemyLowHP {
		actionPriorities[ACTION_ATTACK] *= 1.5 // Boost when enemy is weak
	}

	// Defensive actions
	actionPriorities[ACTION_DEFEND] = weights.DefenseWeight
	if isLowHP {
		actionPriorities[ACTION_DEFEND] *= 1.3 // Boost when we're in danger
	}

	// Healing actions
	actionPriorities[ACTION_HEAL] = weights.HealingWeight
	if isLowHP {
		actionPriorities[ACTION_HEAL] *= 2.0 // Strong boost when low HP
	}

	// Special actions based on personality
	if weights.RiskTolerance > 0.6 {
		actionPriorities[ACTION_STUN] = weights.AttackWeight * 0.7
		actionPriorities[ACTION_DRAIN] = weights.AttackWeight * 0.6
	}

	if weights.DefenseWeight > 0.7 {
		actionPriorities[ACTION_SHIELD] = weights.DefenseWeight * 0.8
		actionPriorities[ACTION_COUNTER] = weights.DefenseWeight * 0.6
	}

	// Add some randomness based on risk tolerance
	if weights.RiskTolerance > ai.rand.Float64() {
		// Sometimes make suboptimal choices for unpredictability
		for action := range actionPriorities {
			actionPriorities[action] += (ai.rand.Float64() - 0.5) * 0.2
		}
	}

	// Find action with highest priority
	var bestAction BattleActionType
	var bestPriority float64

	for action, priority := range actionPriorities {
		if priority > bestPriority {
			bestAction = action
			bestPriority = priority
		}
	}

	if bestAction == "" {
		return ACTION_ATTACK, nil // Fallback to attack
	}

	return bestAction, nil
}

// GetOptimalDecision provides enhanced AI decision making including special abilities and combos
func (ai *PersonalityBasedAI) GetOptimalDecision(personality AIPersonality, hpPercentage, enemyHPPercentage float64, turnNumber int, availableAbilities []SpecialAbility, activeCombo *ComboState, availableCombos []ComboAttack) (*AIDecision, error) {
	weights, exists := ai.personalities[personality]
	if !exists {
		return nil, errors.New("unknown AI personality")
	}

	isLowHP := hpPercentage < 0.3
	isEnemyLowHP := enemyHPPercentage < 0.3
	isHealthyEnough := hpPercentage >= weights.SpecialThreshold

	var decisions []*AIDecision

	// Evaluate continuing active combo (highest priority if active)
	if activeCombo != nil && activeCombo.IsActive {
		comboDecision := ai.evaluateComboContination(activeCombo, weights, availableCombos)
		if comboDecision != nil {
			comboDecision.Priority *= 2.0 // High priority for combo continuation
			decisions = append(decisions, comboDecision)
		}
	}

	// Evaluate special abilities if healthy enough
	if isHealthyEnough && len(availableAbilities) > 0 {
		specialDecisions := ai.evaluateSpecialAbilities(availableAbilities, weights, isLowHP, isEnemyLowHP)
		decisions = append(decisions, specialDecisions...)
	}

	// Evaluate combo starters
	if len(availableCombos) > 0 {
		comboDecisions := ai.evaluateComboStarters(availableCombos, weights, isLowHP, isEnemyLowHP)
		decisions = append(decisions, comboDecisions...)
	}

	// Evaluate basic actions
	basicDecisions := ai.evaluateBasicActions(weights, isLowHP, isEnemyLowHP)
	decisions = append(decisions, basicDecisions...)

	// Add personality-based randomization
	if weights.RiskTolerance > 0.5 {
		for _, decision := range decisions {
			randomFactor := (ai.rand.Float64() - 0.5) * 0.3 * weights.RiskTolerance
			decision.Priority += randomFactor
		}
	}

	// Find best decision
	if len(decisions) == 0 {
		// Fallback to basic attack
		return &AIDecision{
			ActionType: ACTION_ATTACK,
			Priority:   1.0,
			Reasoning:  "fallback to basic attack",
		}, nil
	}

	var bestDecision *AIDecision
	for _, decision := range decisions {
		if bestDecision == nil || decision.Priority > bestDecision.Priority {
			bestDecision = decision
		}
	}

	return bestDecision, nil
}

// evaluateComboContination evaluates continuing an active combo
func (ai *PersonalityBasedAI) evaluateComboContination(activeCombo *ComboState, weights AIDecisionWeights, availableCombos []ComboAttack) *AIDecision {
	// Find the combo definition
	var comboDef *ComboAttack
	for i := range availableCombos {
		if availableCombos[i].Type == activeCombo.Type {
			comboDef = &availableCombos[i]
			break
		}
	}

	if comboDef == nil {
		return nil
	}

	// Get next expected action
	nextActionIndex := len(activeCombo.ActionsCompleted)
	if nextActionIndex >= len(comboDef.Sequence) {
		return nil
	}

	nextAction := comboDef.Sequence[nextActionIndex]
	priority := weights.ComboWeight * weights.ComboAggression * 1.5 // High priority for continuation

	return &AIDecision{
		ActionType:     nextAction,
		Priority:       priority,
		Reasoning:      "continuing active combo: " + comboDef.Name,
		ContinuesCombo: true,
	}
}

// evaluateSpecialAbilities evaluates available special abilities
func (ai *PersonalityBasedAI) evaluateSpecialAbilities(abilities []SpecialAbility, weights AIDecisionWeights, isLowHP, isEnemyLowHP bool) []*AIDecision {
	var decisions []*AIDecision

	for _, ability := range abilities {
		priority := ai.calculateSpecialAbilityPriority(ability, weights, isLowHP, isEnemyLowHP)
		if priority > 0 {
			decisions = append(decisions, &AIDecision{
				SpecialAbility: ability.Type,
				Priority:       priority,
				Reasoning:      "special ability: " + ability.Name,
				IsSpecial:      true,
			})
		}
	}

	return decisions
}

// calculateSpecialAbilityPriority determines priority for a special ability
func (ai *PersonalityBasedAI) calculateSpecialAbilityPriority(ability SpecialAbility, weights AIDecisionWeights, isLowHP, isEnemyLowHP bool) float64 {
	basePriority := weights.SpecialWeight

	switch ability.Type {
	// Offensive abilities
	case ABILITY_CRITICAL_STRIKE, ABILITY_LIGHTNING_BOLT, ABILITY_BERSERKER_RAGE:
		priority := basePriority * weights.AttackWeight
		if isEnemyLowHP {
			priority *= 1.5 // Boost for finishing moves
		}
		return priority

	case ABILITY_LIFE_STEAL:
		priority := basePriority * (weights.AttackWeight + weights.HealingWeight) * 0.5
		if isLowHP {
			priority *= 1.3 // Good when we need health
		}
		return priority

	// Defensive abilities
	case ABILITY_PERFECT_GUARD, ABILITY_SANCTUARY:
		priority := basePriority * weights.DefenseWeight
		if isLowHP {
			priority *= 2.0 // Critical when low HP
		}
		return priority

	// Utility abilities
	case ABILITY_CLEANSE:
		return basePriority * weights.DefenseWeight * 0.7

	case ABILITY_TIME_FREEZE:
		priority := basePriority * weights.RiskTolerance
		if isEnemyLowHP {
			priority *= 0.5 // Less useful when enemy is weak
		}
		return priority

	default:
		return basePriority * 0.5
	}
}

// evaluateComboStarters evaluates actions that start combos
func (ai *PersonalityBasedAI) evaluateComboStarters(combos []ComboAttack, weights AIDecisionWeights, isLowHP, isEnemyLowHP bool) []*AIDecision {
	var decisions []*AIDecision

	for _, combo := range combos {
		if len(combo.Sequence) == 0 {
			continue
		}

		startAction := combo.Sequence[0]
		priority := ai.calculateComboStarterPriority(combo, weights, isLowHP, isEnemyLowHP)

		if priority > 0 {
			decisions = append(decisions, &AIDecision{
				ActionType:  startAction,
				Priority:    priority,
				Reasoning:   "starting combo: " + combo.Name,
				StartsCombo: true,
			})
		}
	}

	return decisions
}

// calculateComboStarterPriority determines priority for starting a combo
func (ai *PersonalityBasedAI) calculateComboStarterPriority(combo ComboAttack, weights AIDecisionWeights, isLowHP, isEnemyLowHP bool) float64 {
	basePriority := weights.ComboWeight * weights.ComboAggression

	// Reduce priority if we're low on HP (risky)
	if isLowHP && weights.RiskTolerance < 0.5 {
		basePriority *= 0.5
	}

	// Consider combo type
	switch combo.Type {
	// Offensive combos
	case COMBO_STUN_ATTACK, COMBO_BOOST_STRIKE, COMBO_BERSERKER_FURY:
		priority := basePriority * weights.AttackWeight
		if isEnemyLowHP {
			priority *= 1.3 // Good for finishing
		}
		return priority

	// Defensive combos
	case COMBO_DEFENSIVE_MASTERY, COMBO_SHIELD_COUNTER_STUN:
		priority := basePriority * weights.DefenseWeight
		if isLowHP {
			priority *= 1.2 // Good when we need defense
		}
		return priority

	// Utility combos
	case COMBO_DRAIN_HEAL:
		priority := basePriority * (weights.AttackWeight + weights.HealingWeight) * 0.5
		if isLowHP {
			priority *= 1.4 // Excellent when we need health
		}
		return priority

	// Complex combos
	case COMBO_CHARGE_BOOST_ATTACK:
		priority := basePriority * weights.AttackWeight * float64(weights.PlanningDepth) / 5.0
		if weights.PlanningDepth < 3 {
			priority *= 0.5 // Too complex for simple personalities
		}
		return priority

	default:
		return basePriority * 0.7
	}
}

// evaluateBasicActions evaluates standard battle actions
func (ai *PersonalityBasedAI) evaluateBasicActions(weights AIDecisionWeights, isLowHP, isEnemyLowHP bool) []*AIDecision {
	var decisions []*AIDecision

	// Attack actions
	attackPriority := weights.AttackWeight
	if isEnemyLowHP {
		attackPriority *= 1.5
	}
	decisions = append(decisions, &AIDecision{
		ActionType: ACTION_ATTACK,
		Priority:   attackPriority,
		Reasoning:  "basic attack",
	})

	// Defense actions
	defensePriority := weights.DefenseWeight
	if isLowHP {
		defensePriority *= 1.3
	}
	decisions = append(decisions, &AIDecision{
		ActionType: ACTION_DEFEND,
		Priority:   defensePriority,
		Reasoning:  "basic defense",
	})

	// Healing actions
	healPriority := weights.HealingWeight
	if isLowHP {
		healPriority *= 2.0
	}
	decisions = append(decisions, &AIDecision{
		ActionType: ACTION_HEAL,
		Priority:   healPriority,
		Reasoning:  "basic heal",
	})

	// Other basic actions based on personality
	if weights.RiskTolerance > 0.6 {
		decisions = append(decisions, &AIDecision{
			ActionType: ACTION_STUN,
			Priority:   weights.AttackWeight * 0.7,
			Reasoning:  "aggressive stun",
		})
		decisions = append(decisions, &AIDecision{
			ActionType: ACTION_DRAIN,
			Priority:   weights.AttackWeight * 0.6,
			Reasoning:  "aggressive drain",
		})
	}

	if weights.DefenseWeight > 0.7 {
		decisions = append(decisions, &AIDecision{
			ActionType: ACTION_SHIELD,
			Priority:   weights.DefenseWeight * 0.8,
			Reasoning:  "defensive shield",
		})
		decisions = append(decisions, &AIDecision{
			ActionType: ACTION_COUNTER,
			Priority:   weights.DefenseWeight * 0.6,
			Reasoning:  "defensive counter",
		})
	}

	return decisions
}

// GetPersonalityDescription returns a human-readable description of a personality
func (ai *PersonalityBasedAI) GetPersonalityDescription(personality AIPersonality) string {
	switch personality {
	case PERSONALITY_AGGRESSIVE:
		return "Focuses on offense, willing to take risks for higher damage"
	case PERSONALITY_DEFENSIVE:
		return "Prioritizes survival and healing, cautious and strategic"
	case PERSONALITY_BALANCED:
		return "Considers all options equally, adaptable to different situations"
	case PERSONALITY_TACTICAL:
		return "Strategic approach, plans several moves ahead"
	default:
		return "Unknown personality type"
	}
}

// GetPersonalityWeights returns the current weights for a personality
func (ai *PersonalityBasedAI) GetPersonalityWeights(personality AIPersonality) (AIDecisionWeights, bool) {
	weights, exists := ai.personalities[personality]
	return weights, exists
}

// AdjustPersonality allows dynamic modification of personality weights
func (ai *PersonalityBasedAI) AdjustPersonality(personality AIPersonality, weights AIDecisionWeights) {
	ai.personalities[personality] = weights
}

// AnalyzeBattleSituation provides tactical analysis for AI decision making
func (ai *PersonalityBasedAI) AnalyzeBattleSituation(hpPercentage, enemyHPPercentage float64, turnNumber int) map[string]float64 {
	analysis := make(map[string]float64)

	analysis["hp_percentage"] = hpPercentage
	analysis["enemy_hp_percentage"] = enemyHPPercentage
	analysis["turn_number"] = float64(turnNumber)
	analysis["danger_level"] = 1.0 - hpPercentage
	analysis["victory_proximity"] = 1.0 - enemyHPPercentage

	return analysis
}
