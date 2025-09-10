package bot

import (
	"testing"
	"time"
)

// MockCharacterController implements CharacterController for testing
type MockCharacterController struct {
	clickCalled       bool
	rightClickCalled  bool
	doubleClickCalled bool

	currentState        string
	lastInteractionTime time.Time
	stats               map[string]float64
	mood                float64
	isGameMode          bool

	// Track method calls for verification
	clickResponses       []string
	rightClickResponses  []string
	doubleClickResponses []string
}

func NewMockCharacterController() *MockCharacterController {
	return &MockCharacterController{
		currentState:        "idle",
		lastInteractionTime: time.Now(),
		stats: map[string]float64{
			"hunger":    70.0,
			"happiness": 80.0,
			"health":    90.0,
			"energy":    60.0,
		},
		mood:       75.0,
		isGameMode: true,

		clickResponses:       []string{"Hello!", "Nice to see you!", "How are you?"},
		rightClickResponses:  []string{"Thank you for the food!", "Yummy!", "I was hungry!"},
		doubleClickResponses: []string{"Let's play!", "This is fun!", "Wheee!"},
	}
}

func (m *MockCharacterController) HandleClick() string {
	m.clickCalled = true
	m.lastInteractionTime = time.Now()

	if len(m.clickResponses) > 0 {
		response := m.clickResponses[0]
		m.clickResponses = m.clickResponses[1:] // Remove used response
		return response
	}
	return "Default click response"
}

func (m *MockCharacterController) HandleRightClick() string {
	m.rightClickCalled = true
	m.lastInteractionTime = time.Now()

	if len(m.rightClickResponses) > 0 {
		response := m.rightClickResponses[0]
		m.rightClickResponses = m.rightClickResponses[1:] // Remove used response
		return response
	}
	return "Default right-click response"
}

func (m *MockCharacterController) HandleDoubleClick() string {
	m.doubleClickCalled = true
	m.lastInteractionTime = time.Now()

	if len(m.doubleClickResponses) > 0 {
		response := m.doubleClickResponses[0]
		m.doubleClickResponses = m.doubleClickResponses[1:] // Remove used response
		return response
	}
	return "Default double-click response"
}

func (m *MockCharacterController) GetCurrentState() string {
	return m.currentState
}

func (m *MockCharacterController) GetLastInteractionTime() time.Time {
	return m.lastInteractionTime
}

func (m *MockCharacterController) GetStats() map[string]float64 {
	// Return a copy to prevent external modification
	stats := make(map[string]float64)
	for k, v := range m.stats {
		stats[k] = v
	}
	return stats
}

func (m *MockCharacterController) GetMood() float64 {
	return m.mood
}

func (m *MockCharacterController) IsGameMode() bool {
	return m.isGameMode
}

// Helper methods for testing
func (m *MockCharacterController) SetStats(stats map[string]float64) {
	m.stats = stats
}

func (m *MockCharacterController) SetMood(mood float64) {
	m.mood = mood
}

func (m *MockCharacterController) SetGameMode(enabled bool) {
	m.isGameMode = enabled
}

func (m *MockCharacterController) Reset() {
	m.clickCalled = false
	m.rightClickCalled = false
	m.doubleClickCalled = false
}

// MockNetworkController implements NetworkController for testing
type MockNetworkController struct {
	peerCount    int
	peerIDs      []string
	isEnabled    bool
	sentMessages []MockMessage
}

type MockMessage struct {
	PeerID  string
	Message interface{}
}

func NewMockNetworkController() *MockNetworkController {
	return &MockNetworkController{
		peerCount:    0,
		peerIDs:      []string{},
		isEnabled:    false,
		sentMessages: []MockMessage{},
	}
}

func (m *MockNetworkController) GetPeerCount() int {
	return m.peerCount
}

func (m *MockNetworkController) GetPeerIDs() []string {
	// Return a copy to prevent external modification
	ids := make([]string, len(m.peerIDs))
	copy(ids, m.peerIDs)
	return ids
}

func (m *MockNetworkController) SendMessage(peerID string, message interface{}) error {
	m.sentMessages = append(m.sentMessages, MockMessage{
		PeerID:  peerID,
		Message: message,
	})
	return nil
}

func (m *MockNetworkController) IsNetworkEnabled() bool {
	return m.isEnabled
}

// Helper methods for testing
func (m *MockNetworkController) SetPeers(peerIDs []string) {
	m.peerIDs = peerIDs
	m.peerCount = len(peerIDs)
}

func (m *MockNetworkController) SetEnabled(enabled bool) {
	m.isEnabled = enabled
}

func (m *MockNetworkController) GetSentMessages() []MockMessage {
	return m.sentMessages
}

func (m *MockNetworkController) Reset() {
	m.sentMessages = []MockMessage{}
}

// Test DefaultPersonality function
func TestDefaultPersonality(t *testing.T) {
	personality := DefaultPersonality()

	// Test basic fields are set
	if personality.ResponseDelay != 3*time.Second {
		t.Errorf("Expected ResponseDelay to be 3s, got %v", personality.ResponseDelay)
	}

	if personality.InteractionRate != 2.0 {
		t.Errorf("Expected InteractionRate to be 2.0, got %f", personality.InteractionRate)
	}

	if personality.Attention != 0.7 {
		t.Errorf("Expected Attention to be 0.7, got %f", personality.Attention)
	}

	// Test social tendencies are set
	if len(personality.SocialTendencies) == 0 {
		t.Error("Expected SocialTendencies to be populated")
	}

	if personality.SocialTendencies["helpfulness"] != 0.8 {
		t.Errorf("Expected helpfulness to be 0.8, got %f", personality.SocialTendencies["helpfulness"])
	}

	// Test emotional profile is set
	if len(personality.EmotionalProfile) == 0 {
		t.Error("Expected EmotionalProfile to be populated")
	}

	if personality.EmotionalProfile["empathy"] != 0.8 {
		t.Errorf("Expected empathy to be 0.8, got %f", personality.EmotionalProfile["empathy"])
	}

	// Test constraints
	if personality.MaxActionsPerMinute != 5 {
		t.Errorf("Expected MaxActionsPerMinute to be 5, got %d", personality.MaxActionsPerMinute)
	}

	if personality.MinTimeBetweenSame != 10 {
		t.Errorf("Expected MinTimeBetweenSame to be 10, got %d", personality.MinTimeBetweenSame)
	}

	// Test preferred actions
	if len(personality.PreferredActions) == 0 {
		t.Error("Expected PreferredActions to be populated")
	}
}

// Test NewBotController with valid parameters
func TestNewBotController_Valid(t *testing.T) {
	personality := DefaultPersonality()
	charController := NewMockCharacterController()
	netController := NewMockNetworkController()

	bot, err := NewBotController(personality, charController, netController)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if bot == nil {
		t.Fatal("Expected bot to be created, got nil")
	}

	if !bot.IsEnabled() {
		t.Error("Expected bot to be enabled by default")
	}

	if bot.characterController != charController {
		t.Error("Expected character controller to be set correctly")
	}

	if bot.networkController != netController {
		t.Error("Expected network controller to be set correctly")
	}
}

// Test NewBotController with nil character controller
func TestNewBotController_NilCharacterController(t *testing.T) {
	personality := DefaultPersonality()
	netController := NewMockNetworkController()

	_, err := NewBotController(personality, nil, netController)

	if err == nil {
		t.Error("Expected error for nil character controller")
	}

	expectedMessage := "character controller cannot be nil"
	if err.Error() != expectedMessage {
		t.Errorf("Expected error message '%s', got '%s'", expectedMessage, err.Error())
	}
}

// Test NewBotController with invalid personality
func TestNewBotController_InvalidPersonality(t *testing.T) {
	personality := DefaultPersonality()
	personality.InteractionRate = -1.0 // Invalid value

	charController := NewMockCharacterController()
	netController := NewMockNetworkController()

	_, err := NewBotController(personality, charController, netController)

	if err == nil {
		t.Error("Expected error for invalid personality")
	}

	if err.Error() == "" {
		t.Error("Expected non-empty error message")
	}
}

// Test validatePersonality function
func TestValidatePersonality(t *testing.T) {
	tests := []struct {
		name        string
		personality BotPersonality
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid personality",
			personality: DefaultPersonality(),
			expectError: false,
		},
		{
			name: "Invalid interaction rate - too low",
			personality: BotPersonality{
				InteractionRate:     0.05, // Too low
				Attention:           0.5,
				MaxActionsPerMinute: 5,
				MinTimeBetweenSame:  10,
			},
			expectError: true,
			errorMsg:    "interaction rate must be between 0.1 and 10.0",
		},
		{
			name: "Invalid interaction rate - too high",
			personality: BotPersonality{
				InteractionRate:     15.0, // Too high
				Attention:           0.5,
				MaxActionsPerMinute: 5,
				MinTimeBetweenSame:  10,
			},
			expectError: true,
			errorMsg:    "interaction rate must be between 0.1 and 10.0",
		},
		{
			name: "Invalid attention - negative",
			personality: BotPersonality{
				InteractionRate:     2.0,
				Attention:           -0.1, // Negative
				MaxActionsPerMinute: 5,
				MinTimeBetweenSame:  10,
			},
			expectError: true,
			errorMsg:    "attention must be between 0.0 and 1.0",
		},
		{
			name: "Invalid attention - too high",
			personality: BotPersonality{
				InteractionRate:     2.0,
				Attention:           1.5, // Too high
				MaxActionsPerMinute: 5,
				MinTimeBetweenSame:  10,
			},
			expectError: true,
			errorMsg:    "attention must be between 0.0 and 1.0",
		},
		{
			name: "Invalid max actions - too low",
			personality: BotPersonality{
				InteractionRate:     2.0,
				Attention:           0.5,
				MaxActionsPerMinute: 0, // Too low
				MinTimeBetweenSame:  10,
			},
			expectError: true,
			errorMsg:    "max actions per minute must be between 1 and 30",
		},
		{
			name: "Invalid max actions - too high",
			personality: BotPersonality{
				InteractionRate:     2.0,
				Attention:           0.5,
				MaxActionsPerMinute: 50, // Too high
				MinTimeBetweenSame:  10,
			},
			expectError: true,
			errorMsg:    "max actions per minute must be between 1 and 30",
		},
		{
			name: "Invalid min time between same - too low",
			personality: BotPersonality{
				InteractionRate:     2.0,
				Attention:           0.5,
				MaxActionsPerMinute: 5,
				MinTimeBetweenSame:  0, // Too low
			},
			expectError: true,
			errorMsg:    "min time between same action must be between 1 and 300 seconds",
		},
		{
			name: "Invalid min time between same - too high",
			personality: BotPersonality{
				InteractionRate:     2.0,
				Attention:           0.5,
				MaxActionsPerMinute: 5,
				MinTimeBetweenSame:  500, // Too high
			},
			expectError: true,
			errorMsg:    "min time between same action must be between 1 and 300 seconds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePersonality(tt.personality)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorMsg != "" && !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

// Helper function to check if string contains substring
func containsString(str, substr string) bool {
	return len(str) >= len(substr) && str[:len(substr)] == substr
}

// Test bot Update() method when disabled
func TestBotController_Update_Disabled(t *testing.T) {
	personality := DefaultPersonality()
	charController := NewMockCharacterController()
	netController := NewMockNetworkController()

	bot, err := NewBotController(personality, charController, netController)
	if err != nil {
		t.Fatalf("Failed to create bot: %v", err)
	}

	bot.Disable()

	actionTaken := bot.Update()

	if actionTaken {
		t.Error("Expected no action when bot is disabled")
	}

	if charController.clickCalled || charController.rightClickCalled || charController.doubleClickCalled {
		t.Error("Expected no character interactions when bot is disabled")
	}
}

// Test bot Update() method when enabled but no scheduled action
func TestBotController_Update_NoScheduledAction(t *testing.T) {
	personality := DefaultPersonality()
	personality.InteractionRate = 0.1 // Very low rate to prevent random actions

	charController := NewMockCharacterController()
	netController := NewMockNetworkController()

	bot, err := NewBotController(personality, charController, netController)
	if err != nil {
		t.Fatalf("Failed to create bot: %v", err)
	}

	// Update multiple times quickly - should not trigger actions due to low interaction rate
	for i := 0; i < 10; i++ {
		actionTaken := bot.Update()
		if actionTaken {
			t.Errorf("Did not expect action on update %d with low interaction rate", i)
		}
	}
}

// Test Enable/Disable functionality
func TestBotController_EnableDisable(t *testing.T) {
	personality := DefaultPersonality()
	charController := NewMockCharacterController()
	netController := NewMockNetworkController()

	bot, err := NewBotController(personality, charController, netController)
	if err != nil {
		t.Fatalf("Failed to create bot: %v", err)
	}

	// Should be enabled by default
	if !bot.IsEnabled() {
		t.Error("Expected bot to be enabled by default")
	}

	// Test disable
	bot.Disable()
	if bot.IsEnabled() {
		t.Error("Expected bot to be disabled after Disable()")
	}

	// Test enable
	bot.Enable()
	if !bot.IsEnabled() {
		t.Error("Expected bot to be enabled after Enable()")
	}
}

// Test GetPersonality/SetPersonality
func TestBotController_GetSetPersonality(t *testing.T) {
	personality := DefaultPersonality()
	charController := NewMockCharacterController()
	netController := NewMockNetworkController()

	bot, err := NewBotController(personality, charController, netController)
	if err != nil {
		t.Fatalf("Failed to create bot: %v", err)
	}

	// Test getting personality
	retrieved := bot.GetPersonality()
	if retrieved.InteractionRate != personality.InteractionRate {
		t.Error("Retrieved personality does not match original")
	}

	// Test setting valid personality
	newPersonality := DefaultPersonality()
	newPersonality.InteractionRate = 5.0

	err = bot.SetPersonality(newPersonality)
	if err != nil {
		t.Errorf("Expected no error setting valid personality, got %v", err)
	}

	retrieved = bot.GetPersonality()
	if retrieved.InteractionRate != 5.0 {
		t.Error("Personality was not updated correctly")
	}

	// Test setting invalid personality
	invalidPersonality := DefaultPersonality()
	invalidPersonality.InteractionRate = -1.0

	err = bot.SetPersonality(invalidPersonality)
	if err == nil {
		t.Error("Expected error setting invalid personality")
	}

	// Ensure personality was not changed after invalid update
	retrieved = bot.GetPersonality()
	if retrieved.InteractionRate != 5.0 {
		t.Error("Personality should not have changed after invalid update")
	}
}

// Test GetStats
func TestBotController_GetStats(t *testing.T) {
	personality := DefaultPersonality()
	charController := NewMockCharacterController()
	netController := NewMockNetworkController()

	bot, err := NewBotController(personality, charController, netController)
	if err != nil {
		t.Fatalf("Failed to create bot: %v", err)
	}

	stats := bot.GetStats()

	// Check expected fields
	expectedFields := []string{"enabled", "actionsInHistory", "hasScheduledAction", "timeSinceLastAction", "decisionsPerSecond"}

	for _, field := range expectedFields {
		if _, exists := stats[field]; !exists {
			t.Errorf("Expected stats to contain field '%s'", field)
		}
	}

	// Check enabled status
	if stats["enabled"] != true {
		t.Error("Expected enabled to be true")
	}

	// Check actions in history (should be 0 initially)
	if stats["actionsInHistory"] != 0 {
		t.Error("Expected actionsInHistory to be 0 initially")
	}

	// Check scheduled action (should be false initially)
	if stats["hasScheduledAction"] != false {
		t.Error("Expected hasScheduledAction to be false initially")
	}
}

// Test GetActionHistory
func TestBotController_GetActionHistory(t *testing.T) {
	personality := DefaultPersonality()
	charController := NewMockCharacterController()
	netController := NewMockNetworkController()

	bot, err := NewBotController(personality, charController, netController)
	if err != nil {
		t.Fatalf("Failed to create bot: %v", err)
	}

	// Initially should be empty
	history := bot.GetActionHistory()
	if len(history) != 0 {
		t.Error("Expected empty action history initially")
	}

	// Add some actions to history (simulate by accessing internal state)
	bot.mu.Lock()
	bot.actionHistory = append(bot.actionHistory, BotDecision{
		Action:      "click",
		Delay:       2 * time.Second,
		Probability: 0.8,
		Priority:    3,
	})
	bot.mu.Unlock()

	history = bot.GetActionHistory()
	if len(history) != 1 {
		t.Errorf("Expected 1 action in history, got %d", len(history))
	}

	if history[0].Action != "click" {
		t.Errorf("Expected action to be 'click', got '%s'", history[0].Action)
	}

	// Ensure we get a copy, not the original slice
	history[0].Action = "modified"
	originalHistory := bot.GetActionHistory()
	if originalHistory[0].Action != "click" {
		t.Error("Action history should return a copy, not the original")
	}
}

// Test generateBasicActions with different character states
func TestBotController_GenerateBasicActions(t *testing.T) {
	personality := DefaultPersonality()
	personality.SocialTendencies["playfulness"] = 0.8
	personality.SocialTendencies["helpfulness"] = 0.9
	personality.EmotionalProfile["enthusiasm"] = 0.7

	charController := NewMockCharacterController()
	netController := NewMockNetworkController()

	bot, err := NewBotController(personality, charController, netController)
	if err != nil {
		t.Fatalf("Failed to create bot: %v", err)
	}

	// Test with normal stats
	actions := bot.generateBasicActions(time.Now())

	if len(actions) == 0 {
		t.Error("Expected some basic actions to be generated")
	}

	// Check that click action is generated due to playfulness
	hasClickAction := false
	for _, action := range actions {
		if action.Action == "click" {
			hasClickAction = true
			if action.Probability <= 0 {
				t.Error("Expected click action to have positive probability")
			}
		}
	}

	if !hasClickAction {
		t.Error("Expected click action to be generated with playfulness > 0")
	}

	// Test with hungry character (should increase feed probability)
	charController.SetStats(map[string]float64{
		"hunger":    20.0, // Very hungry
		"happiness": 80.0,
		"health":    90.0,
		"energy":    60.0,
	})

	actions = bot.generateBasicActions(time.Now())

	hasFeedAction := false
	var feedAction BotDecision
	for _, action := range actions {
		if action.Action == "feed" {
			hasFeedAction = true
			feedAction = action
			break
		}
	}

	if !hasFeedAction {
		t.Error("Expected feed action to be generated when character is hungry")
	} else {
		// Feed action should have higher probability when character is hungry
		if feedAction.Probability <= 0.5 {
			t.Error("Expected higher probability for feed action when character is hungry")
		}
	}

	// Test with low energy (should not generate play action)
	charController.SetStats(map[string]float64{
		"hunger":    70.0,
		"happiness": 80.0,
		"health":    90.0,
		"energy":    20.0, // Very low energy
	})
	charController.SetMood(30.0) // Low mood

	actions = bot.generateBasicActions(time.Now())

	hasPlayAction := false
	for _, action := range actions {
		if action.Action == "play" {
			hasPlayAction = true
		}
	}

	if hasPlayAction {
		t.Error("Did not expect play action when character has low energy and mood")
	}
}

// Test generateNetworkActions
func TestBotController_GenerateNetworkActions(t *testing.T) {
	personality := DefaultPersonality()
	personality.SocialTendencies["chattiness"] = 0.9

	charController := NewMockCharacterController()
	netController := NewMockNetworkController()

	bot, err := NewBotController(personality, charController, netController)
	if err != nil {
		t.Fatalf("Failed to create bot: %v", err)
	}

	// Test with no peers
	actions := bot.generateNetworkActions(time.Now())
	if len(actions) != 0 {
		t.Error("Expected no network actions when no peers are available")
	}

	// Test with peers and network enabled
	netController.SetEnabled(true)
	netController.SetPeers([]string{"peer1", "peer2", "peer3"})

	actions = bot.generateNetworkActions(time.Now())

	if len(actions) == 0 {
		t.Error("Expected network actions when peers are available and chattiness > 0")
	}

	// Check that chat action is generated
	hasChatAction := false
	for _, action := range actions {
		if action.Action == "chat" {
			hasChatAction = true
			if action.Target == "" {
				t.Error("Expected chat action to have a target peer")
			}
			if action.Probability <= 0 {
				t.Error("Expected chat action to have positive probability")
			}
		}
	}

	if !hasChatAction {
		t.Error("Expected chat action to be generated with chattiness > 0")
	}

	// Test with network disabled
	netController.SetEnabled(false)
	actions = bot.generateNetworkActions(time.Now())

	// Should still generate actions because we have peers, but they won't execute
	// The network enabled check is done at execution time
}

// Test rate limiting functionality
func TestBotController_RateLimiting(t *testing.T) {
	personality := DefaultPersonality()
	personality.MaxActionsPerMinute = 2 // Very low limit

	charController := NewMockCharacterController()
	netController := NewMockNetworkController()

	bot, err := NewBotController(personality, charController, netController)
	if err != nil {
		t.Fatalf("Failed to create bot: %v", err)
	}

	// Manually add actions to history to simulate rate limiting
	now := time.Now()
	bot.mu.Lock()
	bot.actionHistory = []BotDecision{
		{
			Action: "click",
			Metadata: map[string]interface{}{
				"timestamp": now.Add(-30 * time.Second), // 30 seconds ago
			},
		},
		{
			Action: "feed",
			Metadata: map[string]interface{}{
				"timestamp": now.Add(-10 * time.Second), // 10 seconds ago
			},
		},
	}
	bot.mu.Unlock()

	// isRateLimited should return true since we have 2 actions in the last minute
	if !bot.isRateLimited() {
		t.Error("Expected bot to be rate limited with 2 actions in last minute and limit of 2")
	}

	// Test with old actions (should not be rate limited)
	bot.mu.Lock()
	bot.actionHistory = []BotDecision{
		{
			Action: "click",
			Metadata: map[string]interface{}{
				"timestamp": now.Add(-2 * time.Minute), // 2 minutes ago
			},
		},
	}
	bot.mu.Unlock()

	if bot.isRateLimited() {
		t.Error("Expected bot not to be rate limited with old actions")
	}
}

// Test action selection by probability
func TestBotController_SelectActionByProbability(t *testing.T) {
	personality := DefaultPersonality()
	charController := NewMockCharacterController()
	netController := NewMockNetworkController()

	bot, err := NewBotController(personality, charController, netController)
	if err != nil {
		t.Fatalf("Failed to create bot: %v", err)
	}

	// Test with empty actions list
	selected := bot.selectActionByProbability([]BotDecision{})
	if selected != nil {
		t.Error("Expected nil when selecting from empty actions list")
	}

	// Test with single action
	actions := []BotDecision{
		{
			Action:      "click",
			Probability: 0.8,
			Priority:    3,
		},
	}

	selected = bot.selectActionByProbability(actions)
	if selected == nil {
		t.Error("Expected action to be selected")
	} else if selected.Action != "click" {
		t.Errorf("Expected 'click' action, got '%s'", selected.Action)
	}

	// Test with multiple actions - higher probability should be more likely
	actions = []BotDecision{
		{
			Action:      "low_prob",
			Probability: 0.1,
			Priority:    1,
		},
		{
			Action:      "high_prob",
			Probability: 0.9,
			Priority:    5,
		},
	}

	// Run selection multiple times to test probability distribution
	highProbCount := 0
	totalSelections := 100

	for i := 0; i < totalSelections; i++ {
		selected = bot.selectActionByProbability(actions)
		if selected != nil && selected.Action == "high_prob" {
			highProbCount++
		}
	}

	// High probability action should be selected more often (allow some variance)
	if highProbCount < totalSelections/2 {
		t.Errorf("Expected high probability action to be selected more often, got %d/%d", highProbCount, totalSelections)
	}
}

// Test delay calculation
func TestBotController_CalculateRandomDelay(t *testing.T) {
	personality := DefaultPersonality()
	personality.ResponseDelay = 4 * time.Second

	charController := NewMockCharacterController()
	netController := NewMockNetworkController()

	bot, err := NewBotController(personality, charController, netController)
	if err != nil {
		t.Fatalf("Failed to create bot: %v", err)
	}

	// Test multiple delay calculations
	delays := make([]time.Duration, 10)
	for i := 0; i < 10; i++ {
		delays[i] = bot.calculateRandomDelay()
	}

	// All delays should be at least 1 second
	for i, delay := range delays {
		if delay < time.Second {
			t.Errorf("Delay %d should be at least 1 second, got %v", i, delay)
		}
	}

	// Delays should vary (not all the same)
	allSame := true
	for i := 1; i < len(delays); i++ {
		if delays[i] != delays[0] {
			allSame = false
			break
		}
	}

	if allSame {
		t.Error("Expected delays to vary, but all were the same")
	}

	// Delays should be roughly around the base delay (within reasonable range)
	baseDelay := personality.ResponseDelay
	for i, delay := range delays {
		// Allow 50% variation from base delay
		minDelay := baseDelay / 2
		maxDelay := baseDelay + (baseDelay / 2)

		if delay < minDelay || delay > maxDelay {
			t.Errorf("Delay %d (%v) outside expected range [%v, %v]", i, delay, minDelay, maxDelay)
		}
	}
}

// Benchmark the Update method for performance validation
func BenchmarkBotController_Update(b *testing.B) {
	personality := DefaultPersonality()
	charController := NewMockCharacterController()
	netController := NewMockNetworkController()

	bot, err := NewBotController(personality, charController, netController)
	if err != nil {
		b.Fatalf("Failed to create bot: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bot.Update()
	}
}

// Benchmark action generation for performance validation
func BenchmarkBotController_GenerateActions(b *testing.B) {
	personality := DefaultPersonality()
	charController := NewMockCharacterController()
	netController := NewMockNetworkController()
	netController.SetEnabled(true)
	netController.SetPeers([]string{"peer1", "peer2", "peer3"})

	bot, err := NewBotController(personality, charController, netController)
	if err != nil {
		b.Fatalf("Failed to create bot: %v", err)
	}

	now := time.Now()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = bot.generatePotentialActions()
		_ = bot.generateBasicActions(now)
		_ = bot.generateNetworkActions(now)
	}
}
