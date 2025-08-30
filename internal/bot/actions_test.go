package bot

import (
	"testing"
	"time"
)

// TestNewActionExecutor tests creation of action executor
func TestNewActionExecutor(t *testing.T) {
	charController := NewMockCharacterController()
	netController := NewMockNetworkController()

	executor := NewActionExecutor(charController, netController)

	if executor == nil {
		t.Fatal("Expected non-nil action executor")
	}

	if executor.characterController != charController {
		t.Error("Character controller not set correctly")
	}

	if executor.networkController != netController {
		t.Error("Network controller not set correctly")
	}

	if executor.maxHistorySize != 100 {
		t.Errorf("Expected max history size 100, got %d", executor.maxHistorySize)
	}

	if len(executor.actionHistory) != 0 {
		t.Errorf("Expected empty action history, got %d items", len(executor.actionHistory))
	}
}

// TestExecuteClickAction tests click action execution
func TestExecuteClickAction(t *testing.T) {
	charController := NewMockCharacterController()
	charController.clickResponses = []string{"Hello there!"}
	charController.stats = map[string]float64{
		"happiness": 50.0,
		"health":    80.0,
	}

	executor := NewActionExecutor(charController, nil)

	decision := BotDecision{
		Action:      "click",
		Probability: 1.0,
		Priority:    1,
	}

	result, err := executor.ExecuteAction(decision)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.Action != ActionClick {
		t.Errorf("Expected action %s, got %s", ActionClick, result.Action)
	}

	if !result.Success {
		t.Error("Expected successful action")
	}

	if result.Response != "Hello there!" {
		t.Errorf("Expected response 'Hello there!', got '%s'", result.Response)
	}

	if !charController.clickCalled {
		t.Error("Expected HandleClick to be called")
	}
}

// TestExecuteFeedAction tests feed action execution
func TestExecuteFeedAction(t *testing.T) {
	charController := NewMockCharacterController()
	charController.rightClickResponses = []string{"Thanks for the food!"}
	charController.stats = map[string]float64{
		"hunger": 30.0,
		"health": 80.0,
	}

	executor := NewActionExecutor(charController, nil)

	decision := BotDecision{
		Action:      "feed",
		Probability: 1.0,
		Priority:    1,
	}

	result, err := executor.ExecuteAction(decision)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Action != ActionFeed {
		t.Errorf("Expected action %s, got %s", ActionFeed, result.Action)
	}

	if !result.Success {
		t.Error("Expected successful action")
	}

	if result.Response != "Thanks for the food!" {
		t.Errorf("Expected response 'Thanks for the food!', got '%s'", result.Response)
	}

	if !charController.rightClickCalled {
		t.Error("Expected HandleRightClick to be called")
	}
}

// TestExecutePlayAction tests play action execution
func TestExecutePlayAction(t *testing.T) {
	charController := NewMockCharacterController()
	charController.doubleClickResponses = []string{"Let's play!"}
	charController.stats = map[string]float64{
		"energy":    80.0,
		"happiness": 60.0,
	}

	executor := NewActionExecutor(charController, nil)

	decision := BotDecision{
		Action:      "play",
		Probability: 1.0,
		Priority:    1,
	}

	result, err := executor.ExecuteAction(decision)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Action != ActionPlay {
		t.Errorf("Expected action %s, got %s", ActionPlay, result.Action)
	}

	if !result.Success {
		t.Error("Expected successful action")
	}

	if result.Response != "Let's play!" {
		t.Errorf("Expected response 'Let's play!', got '%s'", result.Response)
	}

	if !charController.doubleClickCalled {
		t.Error("Expected HandleDoubleClick to be called")
	}
}

// TestExecuteChatActionWithNetwork tests chat action with network enabled
func TestExecuteChatActionWithNetwork(t *testing.T) {
	charController := NewMockCharacterController()
	netController := NewMockNetworkController()
	netController.SetEnabled(true)
	netController.SetPeers([]string{"peer1", "peer2"})

	executor := NewActionExecutor(charController, netController)

	decision := BotDecision{
		Action:      "chat",
		Target:      "peer1",
		Probability: 1.0,
		Priority:    1,
		Metadata: map[string]interface{}{
			"message": "Hello friend!",
		},
	}

	result, err := executor.ExecuteAction(decision)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Action != ActionChat {
		t.Errorf("Expected action %s, got %s", ActionChat, result.Action)
	}

	if !result.Success {
		t.Error("Expected successful action")
	}

	expectedResponse := "Sent message to peer1: Hello friend!"
	if result.Response != expectedResponse {
		t.Errorf("Expected response '%s', got '%s'", expectedResponse, result.Response)
	}

	// Check that message was sent
	messages := netController.GetSentMessages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message sent, got %d", len(messages))
	}

	if messages[0].PeerID != "peer1" {
		t.Errorf("Expected message sent to peer1, got %s", messages[0].PeerID)
	}
}

// TestExecuteChatActionWithoutNetwork tests chat action with network disabled
func TestExecuteChatActionWithoutNetwork(t *testing.T) {
	charController := NewMockCharacterController()
	netController := NewMockNetworkController()
	// Network disabled by default

	executor := NewActionExecutor(charController, netController)

	decision := BotDecision{
		Action:      "chat",
		Probability: 1.0,
		Priority:    1,
		Metadata: map[string]interface{}{
			"message": "Hello!",
		},
	}

	result, err := executor.ExecuteAction(decision)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Action != ActionChat {
		t.Errorf("Expected action %s, got %s", ActionChat, result.Action)
	}

	if !result.Success {
		t.Error("Expected successful action (with fallback message)")
	}

	expectedResponse := "Chat not available - network disabled"
	if result.Response != expectedResponse {
		t.Errorf("Expected response '%s', got '%s'", expectedResponse, result.Response)
	}
}

// TestExecuteWaitAction tests wait action execution
func TestExecuteWaitAction(t *testing.T) {
	charController := NewMockCharacterController()
	executor := NewActionExecutor(charController, nil)

	startTime := time.Now()
	decision := BotDecision{
		Action:      "wait",
		Delay:       50 * time.Millisecond,
		Probability: 1.0,
		Priority:    1,
	}

	result, err := executor.ExecuteAction(decision)
	duration := time.Since(startTime)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Action != ActionWait {
		t.Errorf("Expected action %s, got %s", ActionWait, result.Action)
	}

	if !result.Success {
		t.Error("Expected successful action")
	}

	expectedResponse := "Waited patiently"
	if result.Response != expectedResponse {
		t.Errorf("Expected response '%s', got '%s'", expectedResponse, result.Response)
	}

	// Check that wait actually happened (with some tolerance)
	if duration < 45*time.Millisecond {
		t.Errorf("Expected wait duration of at least 45ms, got %v", duration)
	}
}

// TestExecuteObserveAction tests observe action execution
func TestExecuteObserveAction(t *testing.T) {
	charController := NewMockCharacterController()
	netController := NewMockNetworkController()
	netController.SetEnabled(true)
	netController.SetPeers([]string{"peer1", "peer2"})

	executor := NewActionExecutor(charController, netController)

	decision := BotDecision{
		Action:      "observe",
		Probability: 1.0,
		Priority:    1,
	}

	result, err := executor.ExecuteAction(decision)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Action != ActionObserve {
		t.Errorf("Expected action %s, got %s", ActionObserve, result.Action)
	}

	if !result.Success {
		t.Error("Expected successful action")
	}

	expectedResponse := "Observed 2 peer interactions"
	if result.Response != expectedResponse {
		t.Errorf("Expected response '%s', got '%s'", expectedResponse, result.Response)
	}
}

// TestActionHistoryRecording tests that actions are properly recorded
func TestActionHistoryRecording(t *testing.T) {
	charController := NewMockCharacterController()
	charController.clickResponses = []string{"Click 1", "Click 2"}

	executor := NewActionExecutor(charController, nil)

	// Execute two actions
	decision1 := BotDecision{Action: "click", Probability: 1.0, Priority: 1}
	decision2 := BotDecision{Action: "click", Probability: 1.0, Priority: 1}

	_, err1 := executor.ExecuteAction(decision1)
	_, err2 := executor.ExecuteAction(decision2)

	if err1 != nil || err2 != nil {
		t.Fatalf("Unexpected errors: %v, %v", err1, err2)
	}

	// Check history
	history := executor.GetActionHistory()
	if len(history) != 2 {
		t.Fatalf("Expected 2 items in history, got %d", len(history))
	}

	if history[0].Response != "Click 1" {
		t.Errorf("Expected first response 'Click 1', got '%s'", history[0].Response)
	}

	if history[1].Response != "Click 2" {
		t.Errorf("Expected second response 'Click 2', got '%s'", history[1].Response)
	}
}

// TestActionStatsTracking tests that statistics are properly tracked
func TestActionStatsTracking(t *testing.T) {
	charController := NewMockCharacterController()
	charController.clickResponses = []string{"Success", "", "Success"} // Empty response = failure

	executor := NewActionExecutor(charController, nil)

	// Execute three click actions
	decision := BotDecision{Action: "click", Probability: 1.0, Priority: 1}

	for i := 0; i < 3; i++ {
		_, err := executor.ExecuteAction(decision)
		if err != nil {
			t.Fatalf("Unexpected error on execution %d: %v", i, err)
		}
	}

	// Check statistics
	stats := executor.GetActionStats()
	clickStats, exists := stats[ActionClick]
	if !exists {
		t.Fatal("Expected click action statistics")
	}

	if clickStats.TotalExecutions != 3 {
		t.Errorf("Expected 3 total executions, got %d", clickStats.TotalExecutions)
	}

	// Success rate should be 2/3 ≈ 0.67
	expectedSuccessRate := 2.0 / 3.0
	if abs(clickStats.SuccessRate-expectedSuccessRate) > 0.01 {
		t.Errorf("Expected success rate %.2f, got %.2f", expectedSuccessRate, clickStats.SuccessRate)
	}
}

// TestLearnFromPeerActions tests peer learning functionality
func TestLearnFromPeerActions(t *testing.T) {
	charController := NewMockCharacterController()
	executor := NewActionExecutor(charController, nil)

	// Create peer action events
	peerActions := []PeerActionEvent{
		{
			PeerID:    "peer1",
			Action:    ActionClick,
			Success:   true,
			Response:  "Great interaction!",
			Timestamp: time.Now(),
			CharacterStats: map[string]float64{
				"happiness": 75.0,
				"health":    80.0,
			},
		},
		{
			PeerID:    "peer2",
			Action:    ActionFeed,
			Success:   true,
			Response:  "Thanks for food!",
			Timestamp: time.Now(),
			CharacterStats: map[string]float64{
				"hunger": 25.0, // Low hunger triggered feed
				"health": 90.0,
			},
		},
	}

	// Learn from peer actions (should not panic or error)
	executor.LearnFromPeerActions(peerActions)

	// Test passes if no panic occurs and method completes
	t.Log("Successfully learned from peer actions")
}

// TestGetRecommendedAction tests action recommendation logic
func TestGetRecommendedAction(t *testing.T) {
	charController := NewMockCharacterController()
	charController.isGameMode = true
	charController.stats = map[string]float64{
		"hunger":    20.0, // Low hunger should trigger feed
		"happiness": 70.0,
		"energy":    60.0,
		"health":    80.0,
	}

	executor := NewActionExecutor(charController, nil)

	recommended := executor.GetRecommendedAction()

	// Low hunger should recommend feed action
	if recommended != ActionFeed {
		t.Errorf("Expected recommendation %s for low hunger, got %s", ActionFeed, recommended)
	}
}

// TestGetRecommendedActionNonGameMode tests recommendations in non-game mode
func TestGetRecommendedActionNonGameMode(t *testing.T) {
	charController := NewMockCharacterController()
	charController.isGameMode = false

	executor := NewActionExecutor(charController, nil)

	recommended := executor.GetRecommendedAction()

	// Non-game mode should prefer interaction actions
	if recommended != ActionClick {
		t.Errorf("Expected recommendation %s for non-game mode, got %s", ActionClick, recommended)
	}
}

// TestAnalyzeStatImpact tests stat impact analysis
func TestAnalyzeStatImpact(t *testing.T) {
	charController := NewMockCharacterController()
	executor := NewActionExecutor(charController, nil)

	// Manually add some action results with stat changes
	results := []ActionResult{
		{
			Action:      ActionClick,
			Success:     true,
			StatsBefore: map[string]float64{"happiness": 50.0},
			StatsAfter:  map[string]float64{"happiness": 55.0},
		},
		{
			Action:      ActionClick,
			Success:     true,
			StatsBefore: map[string]float64{"happiness": 60.0},
			StatsAfter:  map[string]float64{"happiness": 67.0},
		},
		{
			Action:      ActionFeed,
			Success:     true,
			StatsBefore: map[string]float64{"hunger": 30.0},
			StatsAfter:  map[string]float64{"hunger": 45.0},
		},
	}

	for _, result := range results {
		executor.recordActionResult(result)
	}

	// Test happiness impact of click actions
	happinessImpact := executor.AnalyzeStatImpact(ActionClick, "happiness")
	expectedImpact := (5.0 + 7.0) / 2.0 // Average of +5 and +7
	if abs(happinessImpact-expectedImpact) > 0.01 {
		t.Errorf("Expected happiness impact %.2f, got %.2f", expectedImpact, happinessImpact)
	}

	// Test hunger impact of feed actions
	hungerImpact := executor.AnalyzeStatImpact(ActionFeed, "hunger")
	expectedHungerImpact := 15.0 // +15 from the feed action
	if abs(hungerImpact-expectedHungerImpact) > 0.01 {
		t.Errorf("Expected hunger impact %.2f, got %.2f", expectedHungerImpact, hungerImpact)
	}
}

// TestResetHistory tests history reset functionality
func TestResetHistory(t *testing.T) {
	charController := NewMockCharacterController()
	charController.clickResponses = []string{"Click!"}

	executor := NewActionExecutor(charController, nil)

	// Execute an action to populate history
	decision := BotDecision{Action: "click", Probability: 1.0, Priority: 1}
	_, err := executor.ExecuteAction(decision)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify history has content
	if len(executor.GetActionHistory()) == 0 {
		t.Fatal("Expected non-empty history before reset")
	}

	if len(executor.GetActionStats()) == 0 {
		t.Fatal("Expected non-empty stats before reset")
	}

	// Reset history
	executor.ResetHistory()

	// Verify history is cleared
	if len(executor.GetActionHistory()) != 0 {
		t.Errorf("Expected empty history after reset, got %d items", len(executor.GetActionHistory()))
	}

	if len(executor.GetActionStats()) != 0 {
		t.Errorf("Expected empty stats after reset, got %d items", len(executor.GetActionStats()))
	}
}

// TestExecuteActionWithNilController tests error handling for nil controller
func TestExecuteActionWithNilController(t *testing.T) {
	executor := NewActionExecutor(nil, nil)

	decision := BotDecision{Action: "click", Probability: 1.0, Priority: 1}

	result, err := executor.ExecuteAction(decision)

	if err == nil {
		t.Error("Expected error for nil character controller")
	}

	if result == nil {
		t.Error("Expected non-nil result even on error")
	}

	if result.Success {
		t.Error("Expected unsuccessful result for nil controller")
	}
}

// TestExecuteActionUnknownType tests error handling for unknown action types
func TestExecuteActionUnknownType(t *testing.T) {
	charController := NewMockCharacterController()
	executor := NewActionExecutor(charController, nil)

	decision := BotDecision{Action: "unknown", Probability: 1.0, Priority: 1}

	result, err := executor.ExecuteAction(decision)

	if err == nil {
		t.Error("Expected error for unknown action type")
	}

	if result != nil {
		t.Error("Expected nil result for unknown action type")
	}
}

// Helper function for floating point comparison
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// BenchmarkExecuteAction benchmarks action execution performance
func BenchmarkExecuteAction(b *testing.B) {
	charController := NewMockCharacterController()
	charController.clickResponses = []string{"Benchmark click"}

	executor := NewActionExecutor(charController, nil)
	decision := BotDecision{Action: "click", Probability: 1.0, Priority: 1}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := executor.ExecuteAction(decision)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

// BenchmarkAnalyzeStatImpact benchmarks stat analysis performance
func BenchmarkAnalyzeStatImpact(b *testing.B) {
	charController := NewMockCharacterController()
	executor := NewActionExecutor(charController, nil)

	// Add sample data
	for i := 0; i < 100; i++ {
		result := ActionResult{
			Action:      ActionClick,
			Success:     true,
			StatsBefore: map[string]float64{"happiness": 50.0},
			StatsAfter:  map[string]float64{"happiness": 55.0},
		}
		executor.recordActionResult(result)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		executor.AnalyzeStatImpact(ActionClick, "happiness")
	}
}

// TestBotControllerActionExecutorIntegration tests the integration between bot controller and action executor
func TestBotControllerActionExecutorIntegration(t *testing.T) {
	charController := NewMockCharacterController()
	charController.clickResponses = []string{"Bot clicked successfully!"}
	charController.stats = map[string]float64{
		"happiness": 50.0,
		"health":    80.0,
	}
	charController.isGameMode = true

	netController := NewMockNetworkController()
	personality := DefaultPersonality()

	botController, err := NewBotController(personality, charController, netController)
	if err != nil {
		t.Fatalf("Failed to create bot controller: %v", err)
	}

	// Force create a scheduled action for testing
	decision := BotDecision{
		Action:      "click",
		Probability: 1.0,
		Priority:    1,
	}

	// Manually set the next action to test execution
	botController.mu.Lock()
	botController.nextScheduledAction = &decision
	botController.mu.Unlock()

	// Update should execute the action
	botController.Update()

	// Verify action was executed
	if !charController.clickCalled {
		t.Error("Expected HandleClick to be called via action executor")
	}

	// Check that action execution history is available
	executionHistory := botController.GetActionExecutionHistory()
	if len(executionHistory) == 0 {
		t.Error("Expected action execution history to be recorded")
	}

	// Check action stats
	actionStats := botController.GetActionStats()
	if clickStats, exists := actionStats[ActionClick]; !exists {
		t.Error("Expected click action statistics to be available")
	} else if clickStats.TotalExecutions != 1 {
		t.Errorf("Expected 1 click execution, got %d", clickStats.TotalExecutions)
	}

	// Test recommendation system
	recommended := botController.GetRecommendedAction()
	if recommended == "" {
		t.Error("Expected a recommendation from the action system")
	}

	t.Logf("Successfully executed action via executor: %d executions recorded", len(executionHistory))
	t.Logf("Recommended next action: %s", recommended)
}
