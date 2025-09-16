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
	AttackWeight  float64 // Weight for offensive actions
	DefenseWeight float64 // Weight for defensive actions
	HealingWeight float64 // Weight for healing actions
	RiskTolerance float64 // Willingness to take risks (0.0 - 1.0)
	PlanningDepth int     // How many turns ahead AI considers
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
	// Aggressive: Prioritizes offense
	ai.personalities[PERSONALITY_AGGRESSIVE] = AIDecisionWeights{
		AttackWeight:  0.8,
		DefenseWeight: 0.2,
		HealingWeight: 0.3,
		RiskTolerance: 0.8,
		PlanningDepth: 2,
	}

	// Defensive: Prioritizes survival
	ai.personalities[PERSONALITY_DEFENSIVE] = AIDecisionWeights{
		AttackWeight:  0.3,
		DefenseWeight: 0.9,
		HealingWeight: 0.8,
		RiskTolerance: 0.2,
		PlanningDepth: 4,
	}

	// Balanced: Equal consideration
	ai.personalities[PERSONALITY_BALANCED] = AIDecisionWeights{
		AttackWeight:  0.6,
		DefenseWeight: 0.6,
		HealingWeight: 0.5,
		RiskTolerance: 0.5,
		PlanningDepth: 3,
	}

	// Tactical: Strategic, high planning
	ai.personalities[PERSONALITY_TACTICAL] = AIDecisionWeights{
		AttackWeight:  0.5,
		DefenseWeight: 0.4,
		HealingWeight: 0.4,
		RiskTolerance: 0.4,
		PlanningDepth: 5,
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
