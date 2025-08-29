package bot

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// CharacterController defines the interface for character behavior control.
// This allows the bot to interact with characters while maintaining clean separation.
type CharacterController interface {
	// Basic interactions - mirror existing Character methods
	HandleClick() string
	HandleRightClick() string
	HandleDoubleClick() string

	// State queries for decision making
	GetCurrentState() string
	GetLastInteractionTime() time.Time

	// Game state queries (if game features enabled)
	GetStats() map[string]float64
	GetMood() float64
	IsGameMode() bool
}

// NetworkController defines the interface for network coordination.
// Allows bot to coordinate with multiplayer features when available.
type NetworkController interface {
	GetPeerCount() int
	GetPeerIDs() []string
	SendMessage(peerID string, message interface{}) error
	IsNetworkEnabled() bool
}

// BotDecision represents an action the bot wants to take.
// Follows the project's preference for explicit, simple data structures.
type BotDecision struct {
	Action      string                 `json:"action"`             // "click", "feed", "play", "chat", "wait"
	Target      string                 `json:"target"`             // Which peer to interact with (empty for self)
	Delay       time.Duration          `json:"delay"`              // When to execute (relative to now)
	Probability float64                `json:"probability"`        // Likelihood of this action (0.0-1.0)
	Priority    int                    `json:"priority"`           // Action priority (higher = more important)
	Metadata    map[string]interface{} `json:"metadata,omitempty"` // Additional context
}

// BotPersonality defines personality traits that drive bot behavior.
// Uses simple float64 values for easy configuration and modification.
type BotPersonality struct {
	// Timing characteristics - how the bot behaves over time
	ResponseDelay   time.Duration `json:"responseDelay"`   // Average delay between actions
	InteractionRate float64       `json:"interactionRate"` // Actions per minute (0.1-10.0)
	Attention       float64       `json:"attention"`       // How quickly bot notices events (0.0-1.0)

	// Social tendencies - how bot interacts with others
	SocialTendencies map[string]float64 `json:"socialTendencies"` // chattiness, helpfulness, playfulness

	// Emotional profile - internal emotional drives
	EmotionalProfile map[string]float64 `json:"emotionalProfile"` // curiosity, empathy, assertiveness

	// Behavioral constraints - limits and preferences
	MaxActionsPerMinute int      `json:"maxActionsPerMinute"` // Rate limiting
	MinTimeBetweenSame  int      `json:"minTimeBetweenSame"`  // Seconds between same action type
	PreferredActions    []string `json:"preferredActions"`    // Actions this bot prefers
}

// BotController manages autonomous character behavior using personality-driven decision making.
// Integrates with the existing Character.Update() cycle following the project's embedding pattern.
//
// Design Philosophy:
// - Uses standard library only (following project guidelines)
// - Simple decision engine with personality traits driving behavior
// - Maintains clean separation from Character implementation
// - Provides natural, human-like delays and actions
type BotController struct {
	mu                  sync.RWMutex
	personality         BotPersonality
	characterController CharacterController
	networkController   NetworkController

	// Decision engine state
	actionHistory       []BotDecision
	lastActionTime      time.Time
	nextScheduledAction *BotDecision
	isEnabled           bool

	// Random source for personality-driven randomness
	rng *rand.Rand

	// Performance tracking
	decisionsPerSecond float64
	lastDecisionTime   time.Time
}

// DefaultPersonality returns a balanced personality suitable for most bot characters.
// Provides sensible defaults that feel natural and responsive.
func DefaultPersonality() BotPersonality {
	return BotPersonality{
		ResponseDelay:   3 * time.Second,
		InteractionRate: 2.0, // 2 actions per minute
		Attention:       0.7, // Moderately attentive

		SocialTendencies: map[string]float64{
			"chattiness":  0.6,
			"helpfulness": 0.8,
			"playfulness": 0.5,
			"curiosity":   0.7,
		},

		EmotionalProfile: map[string]float64{
			"empathy":       0.8,
			"assertiveness": 0.4,
			"patience":      0.7,
			"enthusiasm":    0.6,
		},

		MaxActionsPerMinute: 5,
		MinTimeBetweenSame:  10, // 10 seconds between same action
		PreferredActions:    []string{"click", "chat"},
	}
}

// NewBotController creates a new bot controller with personality-driven behavior.
// Uses dependency injection for clean testing and maintains project's interface-based approach.
func NewBotController(personality BotPersonality, charController CharacterController, netController NetworkController) (*BotController, error) {
	if charController == nil {
		return nil, fmt.Errorf("character controller cannot be nil")
	}

	// Validate personality constraints
	if err := validatePersonality(personality); err != nil {
		return nil, fmt.Errorf("invalid personality: %w", err)
	}

	return &BotController{
		personality:         personality,
		characterController: charController,
		networkController:   netController,
		actionHistory:       make([]BotDecision, 0, 100), // Limit history size
		isEnabled:           true,
		rng:                 rand.New(rand.NewSource(time.Now().UnixNano())),
		lastActionTime:      time.Now(),
		lastDecisionTime:    time.Now(),
	}, nil
}

// validatePersonality ensures personality values are within reasonable ranges.
// Prevents configuration errors that could cause poor bot behavior.
func validatePersonality(p BotPersonality) error {
	if p.InteractionRate < 0.1 || p.InteractionRate > 10.0 {
		return fmt.Errorf("interaction rate must be between 0.1 and 10.0, got %.2f", p.InteractionRate)
	}

	if p.Attention < 0.0 || p.Attention > 1.0 {
		return fmt.Errorf("attention must be between 0.0 and 1.0, got %.2f", p.Attention)
	}

	if p.MaxActionsPerMinute < 1 || p.MaxActionsPerMinute > 30 {
		return fmt.Errorf("max actions per minute must be between 1 and 30, got %d", p.MaxActionsPerMinute)
	}

	if p.MinTimeBetweenSame < 1 || p.MinTimeBetweenSame > 300 {
		return fmt.Errorf("min time between same action must be between 1 and 300 seconds, got %d", p.MinTimeBetweenSame)
	}

	return nil
}

// Update integrates with Character.Update() cycle to provide autonomous behavior.
// Called regularly (e.g., 60 FPS) to maintain responsive bot decision making.
// Returns true if bot performed an action that may have visual impact.
func (bc *BotController) Update() bool {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	if !bc.isEnabled {
		return false
	}

	now := time.Now()

	// Execute scheduled action if it's time
	if bc.nextScheduledAction != nil && now.After(bc.lastActionTime.Add(bc.nextScheduledAction.Delay)) {
		actionTaken := bc.executeScheduledAction()
		bc.nextScheduledAction = nil
		if actionTaken {
			bc.lastActionTime = now
		}
		return actionTaken
	}

	// Make new decision if no action is scheduled
	if bc.nextScheduledAction == nil {
		bc.makeDecision(now)
	}

	return false
}

// makeDecision uses personality traits to decide what action to take next.
// Implements a simple but effective decision engine based on:
// - Time since last action
// - Current character state
// - Personality preferences
// - Network context (if available)
func (bc *BotController) makeDecision(now time.Time) {
	// Calculate time since last action
	timeSinceLastAction := now.Sub(bc.lastActionTime)

	// Check if we should act based on personality
	shouldAct := bc.shouldTakeAction(timeSinceLastAction)
	if !shouldAct {
		return
	}

	// Generate potential actions based on personality
	actions := bc.generatePotentialActions()
	if len(actions) == 0 {
		return
	}

	// Select best action using weighted random selection
	selectedAction := bc.selectActionByProbability(actions)
	if selectedAction != nil {
		bc.nextScheduledAction = selectedAction
	}
}

// shouldTakeAction determines if the bot should take an action based on personality and timing.
// Uses probabilistic decision making to create natural, varied behavior patterns.
func (bc *BotController) shouldTakeAction(timeSinceLastAction time.Duration) bool {
	// Base probability increases with time since last action
	baseProbability := float64(timeSinceLastAction.Seconds()) / bc.personality.ResponseDelay.Seconds()

	// Modify by attention trait - more attentive bots act more frequently
	attentionModifier := bc.personality.Attention * 0.5
	totalProbability := (baseProbability + attentionModifier) * bc.personality.InteractionRate

	// Cap probability to prevent excessive actions
	if totalProbability > 1.0 {
		totalProbability = 1.0
	}

	// Random decision based on calculated probability
	return bc.rng.Float64() < totalProbability
}

// generatePotentialActions creates a list of possible actions based on current context.
// Considers personality preferences, character state, and available interactions.
func (bc *BotController) generatePotentialActions() []BotDecision {
	actions := make([]BotDecision, 0, 10)
	now := time.Now()

	// Check rate limiting
	if bc.isRateLimited() {
		return actions
	}

	// Generate basic interaction actions
	actions = append(actions, bc.generateBasicActions(now)...)

	// Generate network actions if network is available
	if bc.networkController != nil && bc.networkController.IsNetworkEnabled() {
		actions = append(actions, bc.generateNetworkActions(now)...)
	}

	// Filter out actions that are too recent
	return bc.filterRecentActions(actions)
}

// generateBasicActions creates potential basic character interactions.
// Uses personality traits to weight different types of actions.
func (bc *BotController) generateBasicActions(now time.Time) []BotDecision {
	actions := make([]BotDecision, 0, 5)

	// Click action - weighted by playfulness
	playfulness := bc.personality.SocialTendencies["playfulness"]
	if playfulness > 0 {
		actions = append(actions, BotDecision{
			Action:      "click",
			Delay:       bc.calculateRandomDelay(),
			Probability: playfulness * 0.8,
			Priority:    3,
		})
	}

	// Feed action (right-click) - weighted by helpfulness and character needs
	helpfulness := bc.personality.SocialTendencies["helpfulness"]
	if helpfulness > 0 && bc.characterController.IsGameMode() {
		stats := bc.characterController.GetStats()
		hungerLevel := stats["hunger"]

		// Higher probability if character is hungry
		hungerModifier := 1.0
		if hungerLevel < 50 {
			hungerModifier = 2.0
		}

		actions = append(actions, BotDecision{
			Action:      "feed",
			Delay:       bc.calculateRandomDelay(),
			Probability: helpfulness * 0.6 * hungerModifier,
			Priority:    4,
		})
	}

	// Play action (double-click) - weighted by energy and mood
	if bc.characterController.IsGameMode() {
		stats := bc.characterController.GetStats()
		energy := stats["energy"]
		mood := bc.characterController.GetMood()

		// Only play if character has energy and mood is decent
		if energy > 30 && mood > 40 {
			enthusiasm := bc.personality.EmotionalProfile["enthusiasm"]
			actions = append(actions, BotDecision{
				Action:      "play",
				Delay:       bc.calculateRandomDelay(),
				Probability: enthusiasm * 0.7,
				Priority:    2,
			})
		}
	}

	return actions
}

// generateNetworkActions creates potential network-based actions.
// Considers peer availability and social personality traits.
func (bc *BotController) generateNetworkActions(now time.Time) []BotDecision {
	actions := make([]BotDecision, 0, 3)

	peerCount := bc.networkController.GetPeerCount()
	if peerCount == 0 {
		return actions
	}

	chattiness := bc.personality.SocialTendencies["chattiness"]
	if chattiness > 0 {
		peerIDs := bc.networkController.GetPeerIDs()

		// Select a random peer to interact with
		if len(peerIDs) > 0 {
			targetPeer := peerIDs[bc.rng.Intn(len(peerIDs))]

			actions = append(actions, BotDecision{
				Action:      "chat",
				Target:      targetPeer,
				Delay:       bc.calculateRandomDelay(),
				Probability: chattiness * 0.9,
				Priority:    5,
				Metadata:    map[string]interface{}{"peerCount": peerCount},
			})
		}
	}

	return actions
}

// filterRecentActions removes actions that violate minimum time constraints.
// Prevents repetitive behavior that feels unnatural.
func (bc *BotController) filterRecentActions(actions []BotDecision) []BotDecision {
	filtered := make([]BotDecision, 0, len(actions))
	minTime := time.Duration(bc.personality.MinTimeBetweenSame) * time.Second

	for _, action := range actions {
		if bc.canPerformAction(action.Action, minTime) {
			filtered = append(filtered, action)
		}
	}

	return filtered
}

// canPerformAction checks if enough time has passed since the last instance of this action type.
func (bc *BotController) canPerformAction(actionType string, minTime time.Duration) bool {
	for i := len(bc.actionHistory) - 1; i >= 0; i-- {
		if bc.actionHistory[i].Action == actionType {
			return time.Since(bc.lastActionTime) >= minTime
		}
	}
	return true
}

// selectActionByProbability chooses an action using weighted random selection.
// Higher probability and priority actions are more likely to be selected.
func (bc *BotController) selectActionByProbability(actions []BotDecision) *BotDecision {
	if len(actions) == 0 {
		return nil
	}

	// Calculate total weight (probability * priority)
	totalWeight := 0.0
	for _, action := range actions {
		weight := action.Probability * float64(action.Priority)
		totalWeight += weight
	}

	if totalWeight <= 0 {
		return nil
	}

	// Select random point in weight distribution
	randomPoint := bc.rng.Float64() * totalWeight

	// Find selected action
	currentWeight := 0.0
	for _, action := range actions {
		weight := action.Probability * float64(action.Priority)
		currentWeight += weight

		if randomPoint <= currentWeight {
			return &action
		}
	}

	// Fallback to first action if selection fails
	return &actions[0]
}

// calculateRandomDelay generates a human-like delay based on personality.
// Adds natural variation to prevent mechanical behavior.
func (bc *BotController) calculateRandomDelay() time.Duration {
	baseDelay := bc.personality.ResponseDelay

	// Add random variation (±50% of base delay)
	variation := bc.rng.Float64() - 0.5 // -0.5 to +0.5
	adjustedDelay := baseDelay + time.Duration(float64(baseDelay)*variation)

	// Ensure minimum delay of 1 second
	if adjustedDelay < time.Second {
		adjustedDelay = time.Second
	}

	return adjustedDelay
}

// executeScheduledAction performs the planned action and records it in history.
// Returns true if action was successfully executed.
func (bc *BotController) executeScheduledAction() bool {
	if bc.nextScheduledAction == nil {
		return false
	}

	action := *bc.nextScheduledAction
	success := false

	// Execute action based on type
	switch action.Action {
	case "click":
		bc.characterController.HandleClick()
		success = true

	case "feed":
		bc.characterController.HandleRightClick()
		success = true

	case "play":
		bc.characterController.HandleDoubleClick()
		success = true

	case "chat":
		if bc.networkController != nil && action.Target != "" {
			err := bc.networkController.SendMessage(action.Target, map[string]interface{}{
				"type":    "bot_chat",
				"message": "Hello! How are you?",
			})
			success = (err == nil)
		}

	default:
		// Unknown action type - do nothing
		return false
	}

	// Record action in history (limit size to prevent memory growth)
	bc.recordAction(action)

	return success
}

// recordAction adds an action to history with size limiting.
func (bc *BotController) recordAction(action BotDecision) {
	bc.actionHistory = append(bc.actionHistory, action)

	// Limit history size to prevent unbounded growth
	maxHistory := 50
	if len(bc.actionHistory) > maxHistory {
		bc.actionHistory = bc.actionHistory[len(bc.actionHistory)-maxHistory:]
	}
}

// isRateLimited checks if the bot has exceeded its maximum actions per minute.
func (bc *BotController) isRateLimited() bool {
	if bc.personality.MaxActionsPerMinute <= 0 {
		return false
	}

	// Count actions in the last minute
	oneMinuteAgo := time.Now().Add(-time.Minute)
	recentActions := 0

	for i := len(bc.actionHistory) - 1; i >= 0; i-- {
		if bc.actionHistory[i].Metadata != nil {
			if timestamp, ok := bc.actionHistory[i].Metadata["timestamp"].(time.Time); ok {
				if timestamp.After(oneMinuteAgo) {
					recentActions++
				} else {
					break // History is chronological, so we can stop here
				}
			}
		}
	}

	return recentActions >= bc.personality.MaxActionsPerMinute
}

// GetPersonality returns a copy of the current personality configuration.
func (bc *BotController) GetPersonality() BotPersonality {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	return bc.personality
}

// SetPersonality updates the bot's personality and validates the new configuration.
func (bc *BotController) SetPersonality(personality BotPersonality) error {
	if err := validatePersonality(personality); err != nil {
		return fmt.Errorf("invalid personality: %w", err)
	}

	bc.mu.Lock()
	bc.personality = personality
	bc.mu.Unlock()

	return nil
}

// Enable activates the bot controller.
func (bc *BotController) Enable() {
	bc.mu.Lock()
	bc.isEnabled = true
	bc.mu.Unlock()
}

// Disable deactivates the bot controller and clears any scheduled actions.
func (bc *BotController) Disable() {
	bc.mu.Lock()
	bc.isEnabled = false
	bc.nextScheduledAction = nil
	bc.mu.Unlock()
}

// IsEnabled returns whether the bot controller is currently active.
func (bc *BotController) IsEnabled() bool {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	return bc.isEnabled
}

// GetActionHistory returns a copy of the recent action history for debugging.
func (bc *BotController) GetActionHistory() []BotDecision {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	// Return a copy to prevent external modification
	history := make([]BotDecision, len(bc.actionHistory))
	copy(history, bc.actionHistory)

	return history
}

// GetStats returns performance and behavior statistics for monitoring.
func (bc *BotController) GetStats() map[string]interface{} {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	return map[string]interface{}{
		"enabled":             bc.isEnabled,
		"actionsInHistory":    len(bc.actionHistory),
		"hasScheduledAction":  bc.nextScheduledAction != nil,
		"timeSinceLastAction": time.Since(bc.lastActionTime),
		"decisionsPerSecond":  bc.decisionsPerSecond,
	}
}
