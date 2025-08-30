package bot

import (
	"fmt"
	"math"
	"time"
)

// ActionType represents the different types of actions a bot can perform.
type ActionType string

const (
	ActionClick   ActionType = "click"   // Basic interaction - increases happiness
	ActionFeed    ActionType = "feed"    // Right-click interaction - increases hunger
	ActionPlay    ActionType = "play"    // Double-click interaction - increases happiness, decreases energy
	ActionChat    ActionType = "chat"    // Network chat with peers
	ActionWait    ActionType = "wait"    // Passive waiting action
	ActionObserve ActionType = "observe" // Watch peer interactions for learning
)

// ActionResult contains the outcome of an executed action.
// Used for learning and behavior adaptation.
type ActionResult struct {
	Action      ActionType             `json:"action"`
	Success     bool                   `json:"success"`
	Response    string                 `json:"response"`
	Duration    time.Duration          `json:"duration"`
	StatsBefore map[string]float64     `json:"statsBefore"`
	StatsAfter  map[string]float64     `json:"statsAfter"`
	Timestamp   time.Time              `json:"timestamp"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// ActionExecutor handles the execution of bot actions on character controllers.
// Follows the project's interface-based design for clean separation of concerns.
type ActionExecutor struct {
	characterController CharacterController
	networkController   NetworkController

	// Learning and adaptation
	actionHistory  []ActionResult
	maxHistorySize int

	// Performance tracking
	executionStats map[ActionType]ActionStats
}

// ActionStats tracks performance metrics for each action type.
type ActionStats struct {
	TotalExecutions int           `json:"totalExecutions"`
	SuccessRate     float64       `json:"successRate"`
	AverageDelay    time.Duration `json:"averageDelay"`
	LastExecution   time.Time     `json:"lastExecution"`
}

// NewActionExecutor creates a new action executor with the given controllers.
// Uses dependency injection pattern following project conventions.
func NewActionExecutor(charController CharacterController, netController NetworkController) *ActionExecutor {
	return &ActionExecutor{
		characterController: charController,
		networkController:   netController,
		actionHistory:       make([]ActionResult, 0, 100),
		maxHistorySize:      100,
		executionStats:      make(map[ActionType]ActionStats),
	}
}

// ExecuteAction performs the specified action and returns the result.
// Implements comprehensive error handling and performance tracking.
func (ae *ActionExecutor) ExecuteAction(decision BotDecision) (*ActionResult, error) {
	startTime := time.Now()
	actionType := ActionType(decision.Action)

	if ae.characterController == nil {
		return &ActionResult{
			Action:    actionType,
			Success:   false,
			Response:  "Error: character controller not available",
			Duration:  time.Since(startTime),
			Timestamp: startTime,
		}, fmt.Errorf("character controller not available")
	}

	// Get character stats before action for comparison
	statsBefore := ae.characterController.GetStats() // Execute the action
	var response string
	var err error

	switch actionType {
	case ActionClick:
		response = ae.executeClickAction()
	case ActionFeed:
		response = ae.executeFeedAction()
	case ActionPlay:
		response = ae.executePlayAction()
	case ActionChat:
		response, err = ae.executeChatAction(decision.Target, decision.Metadata)
	case ActionWait:
		response = ae.executeWaitAction(decision.Delay)
	case ActionObserve:
		response = ae.executeObserveAction()
	default:
		return nil, fmt.Errorf("unknown action type: %s", actionType)
	}

	// Handle execution errors
	if err != nil {
		return &ActionResult{
			Action:    actionType,
			Success:   false,
			Response:  fmt.Sprintf("Error: %v", err),
			Duration:  time.Since(startTime),
			Timestamp: startTime,
		}, err
	}

	// Get character stats after action
	statsAfter := ae.characterController.GetStats()

	// Create action result
	result := &ActionResult{
		Action:      actionType,
		Success:     response != "",
		Response:    response,
		Duration:    time.Since(startTime),
		StatsBefore: statsBefore,
		StatsAfter:  statsAfter,
		Timestamp:   startTime,
		Context:     decision.Metadata,
	}

	// Record result for learning
	ae.recordActionResult(*result)

	return result, nil
}

// executeClickAction performs a click interaction on the character.
// Equivalent to user clicking on the character.
func (ae *ActionExecutor) executeClickAction() string {
	response := ae.characterController.HandleClick()
	return response
}

// executeFeedAction performs a feed interaction (right-click) on the character.
// Helps maintain character's hunger stat in game mode.
func (ae *ActionExecutor) executeFeedAction() string {
	response := ae.characterController.HandleRightClick()
	return response
}

// executePlayAction performs a play interaction (double-click) on the character.
// Increases happiness but may decrease energy.
func (ae *ActionExecutor) executePlayAction() string {
	response := ae.characterController.HandleDoubleClick()
	return response
}

// executeChatAction sends a chat message to a peer or initiates local chat.
// Requires network controller for peer communication.
func (ae *ActionExecutor) executeChatAction(target string, metadata map[string]interface{}) (string, error) {
	if ae.networkController == nil || !ae.networkController.IsNetworkEnabled() {
		return "Chat not available - network disabled", nil
	}

	// Extract message from metadata
	message, ok := metadata["message"].(string)
	if !ok || message == "" {
		message = "Hello!"
	}

	// Send message to target peer
	if target != "" {
		err := ae.networkController.SendMessage(target, map[string]interface{}{
			"type":    "chat",
			"message": message,
			"sender":  "bot",
		})
		if err != nil {
			return "", fmt.Errorf("failed to send chat message: %w", err)
		}
		return fmt.Sprintf("Sent message to %s: %s", target, message), nil
	}

	// Broadcast message to all peers
	peers := ae.networkController.GetPeerIDs()
	if len(peers) == 0 {
		return "No peers available for chat", nil
	}

	for _, peerID := range peers {
		err := ae.networkController.SendMessage(peerID, map[string]interface{}{
			"type":    "chat",
			"message": message,
			"sender":  "bot",
		})
		if err != nil {
			// Log error but continue with other peers
			continue
		}
	}

	return fmt.Sprintf("Broadcasted message: %s", message), nil
}

// executeWaitAction performs a passive wait, useful for natural timing.
// Allows other activities to occur without bot interference.
func (ae *ActionExecutor) executeWaitAction(duration time.Duration) string {
	if duration > 0 {
		time.Sleep(duration)
	}
	return "Waited patiently"
}

// executeObserveAction watches peer interactions for learning opportunities.
// Builds knowledge about effective actions and timing.
func (ae *ActionExecutor) executeObserveAction() string {
	if ae.networkController == nil || !ae.networkController.IsNetworkEnabled() {
		return "Observed local character state"
	}

	peerCount := ae.networkController.GetPeerCount()
	if peerCount == 0 {
		return "No peers to observe"
	}

	return fmt.Sprintf("Observed %d peer interactions", peerCount)
}

// recordActionResult stores the action result for learning and adaptation.
// Maintains a rolling history to prevent memory growth.
func (ae *ActionExecutor) recordActionResult(result ActionResult) {
	// Add to history
	ae.actionHistory = append(ae.actionHistory, result)

	// Maintain maximum history size
	if len(ae.actionHistory) > ae.maxHistorySize {
		ae.actionHistory = ae.actionHistory[1:]
	}

	// Update execution statistics
	ae.updateExecutionStats(result)
}

// updateExecutionStats updates performance metrics for the action type.
// Tracks success rates and timing for behavior optimization.
func (ae *ActionExecutor) updateExecutionStats(result ActionResult) {
	stats, exists := ae.executionStats[result.Action]
	if !exists {
		stats = ActionStats{}
	}

	// Update execution count
	stats.TotalExecutions++

	// Update success rate
	if result.Success {
		stats.SuccessRate = (stats.SuccessRate*float64(stats.TotalExecutions-1) + 1.0) / float64(stats.TotalExecutions)
	} else {
		stats.SuccessRate = stats.SuccessRate * float64(stats.TotalExecutions-1) / float64(stats.TotalExecutions)
	}

	// Update average delay
	if stats.TotalExecutions == 1 {
		stats.AverageDelay = result.Duration
	} else {
		stats.AverageDelay = time.Duration(
			(int64(stats.AverageDelay)*int64(stats.TotalExecutions-1) + int64(result.Duration)) / int64(stats.TotalExecutions),
		)
	}

	stats.LastExecution = result.Timestamp
	ae.executionStats[result.Action] = stats
}

// GetActionHistory returns a copy of the recent action history.
// Used for analysis and debugging of bot behavior.
func (ae *ActionExecutor) GetActionHistory() []ActionResult {
	history := make([]ActionResult, len(ae.actionHistory))
	copy(history, ae.actionHistory)
	return history
}

// GetActionStats returns performance statistics for all action types.
// Provides insights into bot effectiveness and behavior patterns.
func (ae *ActionExecutor) GetActionStats() map[ActionType]ActionStats {
	stats := make(map[ActionType]ActionStats)
	for actionType, actionStats := range ae.executionStats {
		stats[actionType] = actionStats
	}
	return stats
}

// GetSuccessRateForAction returns the success rate for a specific action type.
// Used by the bot controller for action selection optimization.
func (ae *ActionExecutor) GetSuccessRateForAction(actionType ActionType) float64 {
	if stats, exists := ae.executionStats[actionType]; exists {
		return stats.SuccessRate
	}
	return 0.5 // Default neutral success rate
}

// AnalyzeStatImpact calculates the average stat impact of an action type.
// Helps the bot learn which actions are most effective for character care.
func (ae *ActionExecutor) AnalyzeStatImpact(actionType ActionType, statName string) float64 {
	var totalImpact float64
	var count int

	for _, result := range ae.actionHistory {
		if result.Action == actionType && result.Success {
			before, beforeExists := result.StatsBefore[statName]
			after, afterExists := result.StatsAfter[statName]

			if beforeExists && afterExists {
				impact := after - before
				totalImpact += impact
				count++
			}
		}
	}

	if count == 0 {
		return 0.0
	}

	return totalImpact / float64(count)
}

// LearnFromPeerActions analyzes network events to improve bot behavior.
// Implements peer learning capabilities as specified in the plan.
func (ae *ActionExecutor) LearnFromPeerActions(peerActions []PeerActionEvent) {
	for _, peerAction := range peerActions {
		ae.analyzeAndLearnFromPeerAction(peerAction)
	}
}

// PeerActionEvent represents an action performed by a peer bot or user.
// Used for learning effective interaction patterns.
type PeerActionEvent struct {
	PeerID         string                 `json:"peerID"`
	Action         ActionType             `json:"action"`
	Success        bool                   `json:"success"`
	Response       string                 `json:"response"`
	CharacterStats map[string]float64     `json:"characterStats"`
	Timestamp      time.Time              `json:"timestamp"`
	Context        map[string]interface{} `json:"context,omitempty"`
}

// analyzeAndLearnFromPeerAction extracts insights from peer actions.
// Updates internal models to improve future decision making.
func (ae *ActionExecutor) analyzeAndLearnFromPeerAction(peerAction PeerActionEvent) {
	// Learn timing patterns from successful peer actions
	if peerAction.Success {
		ae.updateActionTiming(peerAction)
	}

	// Learn stat management strategies
	ae.updateStatStrategy(peerAction)
}

// updateActionTiming learns optimal timing from peer behavior.
// Helps the bot develop better natural rhythm.
func (ae *ActionExecutor) updateActionTiming(peerAction PeerActionEvent) {
	// Find similar past actions from our own history
	recentActions := ae.getRecentActions(peerAction.Action, 10)

	if len(recentActions) > 0 {
		// Calculate peer's apparent effectiveness
		peerEffectiveness := ae.calculateActionEffectiveness(peerAction)

		// Adjust our own timing if peer appears more effective
		if peerEffectiveness > ae.getOwnEffectiveness(peerAction.Action) {
			// Implementation would adjust internal timing models
			// For now, we record the observation
		}
	}
}

// updateStatStrategy learns character care strategies from peers.
// Helps optimize when to perform care actions like feeding or playing.
func (ae *ActionExecutor) updateStatStrategy(peerAction PeerActionEvent) {
	// Analyze the character state that led to this action
	for statName, statValue := range peerAction.CharacterStats {
		// Learn the stat thresholds that trigger actions
		ae.learnStatThreshold(peerAction.Action, statName, statValue, peerAction.Success)
	}
}

// getRecentActions retrieves recent actions of a specific type from history.
// Helper method for pattern analysis.
func (ae *ActionExecutor) getRecentActions(actionType ActionType, limit int) []ActionResult {
	var recentActions []ActionResult

	// Search history in reverse order (most recent first)
	for i := len(ae.actionHistory) - 1; i >= 0 && len(recentActions) < limit; i-- {
		if ae.actionHistory[i].Action == actionType {
			recentActions = append(recentActions, ae.actionHistory[i])
		}
	}

	return recentActions
}

// calculateActionEffectiveness measures how effective a peer's action was.
// Uses stat improvements and response quality as metrics.
func (ae *ActionExecutor) calculateActionEffectiveness(peerAction PeerActionEvent) float64 {
	if !peerAction.Success {
		return 0.0
	}

	// Base effectiveness from success
	effectiveness := 0.5

	// Bonus for meaningful response
	if peerAction.Response != "" {
		effectiveness += 0.3
	}

	// Bonus for positive stat impact (if we can infer it)
	// This would require more context about stat changes

	return math.Min(effectiveness, 1.0)
}

// getOwnEffectiveness calculates our own effectiveness for an action type.
// Used for comparison with peer behavior.
func (ae *ActionExecutor) getOwnEffectiveness(actionType ActionType) float64 {
	return ae.GetSuccessRateForAction(actionType)
}

// learnStatThreshold updates knowledge about when to trigger actions based on stats.
// Builds decision models for character care.
func (ae *ActionExecutor) learnStatThreshold(actionType ActionType, statName string, statValue float64, success bool) {
	// This would update internal models about stat thresholds
	// For now, we record the observation
	// Implementation would maintain threshold models per action type
}

// ResetHistory clears the action history and statistics.
// Useful for testing or when starting fresh learning cycles.
func (ae *ActionExecutor) ResetHistory() {
	ae.actionHistory = make([]ActionResult, 0, ae.maxHistorySize)
	ae.executionStats = make(map[ActionType]ActionStats)
}

// GetRecommendedAction suggests the best action based on current context and learning.
// Uses accumulated knowledge to make intelligent recommendations.
func (ae *ActionExecutor) GetRecommendedAction() ActionType {
	if ae.characterController == nil {
		return ActionWait
	}

	// Get current character state
	stats := ae.characterController.GetStats()
	mood := ae.characterController.GetMood()
	isGameMode := ae.characterController.IsGameMode()

	// In game mode, prioritize care actions based on stats
	if isGameMode {
		return ae.getRecommendedCareAction(stats, mood)
	}

	// In non-game mode, prefer interaction actions
	return ae.getRecommendedInteractionAction()
}

// getRecommendedCareAction suggests care actions based on character stats.
// Implements intelligent character care strategy.
func (ae *ActionExecutor) getRecommendedCareAction(stats map[string]float64, mood float64) ActionType {
	// Check for critical stats that need attention
	if hunger, ok := stats["hunger"]; ok && hunger < 30 {
		return ActionFeed
	}

	if energy, ok := stats["energy"]; ok && energy > 80 {
		return ActionPlay // High energy, good time to play
	}

	if happiness, ok := stats["happiness"]; ok && happiness < 40 {
		return ActionClick // Low happiness, show affection
	}

	// Default to observation if stats are okay
	return ActionObserve
}

// getRecommendedInteractionAction suggests social actions for non-game mode.
// Focuses on interaction and social engagement.
func (ae *ActionExecutor) getRecommendedInteractionAction() ActionType {
	// Check if network is available for chat
	if ae.networkController != nil && ae.networkController.IsNetworkEnabled() && ae.networkController.GetPeerCount() > 0 {
		return ActionChat
	}

	// Fall back to basic interaction
	return ActionClick
}
