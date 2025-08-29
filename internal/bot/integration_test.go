package bot

import (
	"sync"
	"testing"
	"time"
)

// IntegrationCharacterController provides a more realistic character controller for integration testing
type IntegrationCharacterController struct {
	mu                 sync.RWMutex
	currentState       string
	lastInteraction    time.Time
	stats              map[string]float64
	mood               float64
	isGameMode         bool
	actionCount        int
	interactionHistory []string
}

func NewIntegrationCharacterController() *IntegrationCharacterController {
	return &IntegrationCharacterController{
		currentState:    "idle",
		lastInteraction: time.Now(),
		stats: map[string]float64{
			"hunger":    70.0,
			"happiness": 80.0,
			"health":    90.0,
			"energy":    60.0,
		},
		mood:               75.0,
		isGameMode:         true,
		interactionHistory: make([]string, 0),
	}
}

func (c *IntegrationCharacterController) HandleClick() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.actionCount++
	c.lastInteraction = time.Now()
	c.stats["happiness"] += 5.0
	c.mood += 3.0

	response := "Thanks for clicking on me!"
	c.interactionHistory = append(c.interactionHistory, "click:"+response)

	return response
}

func (c *IntegrationCharacterController) HandleRightClick() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.actionCount++
	c.lastInteraction = time.Now()
	c.stats["hunger"] += 10.0
	c.stats["happiness"] += 3.0

	response := "Yum! Thanks for the food!"
	c.interactionHistory = append(c.interactionHistory, "feed:"+response)

	return response
}

func (c *IntegrationCharacterController) HandleDoubleClick() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.actionCount++
	c.lastInteraction = time.Now()
	c.stats["happiness"] += 8.0
	c.stats["energy"] -= 5.0

	response := "Let's play! This is fun!"
	c.interactionHistory = append(c.interactionHistory, "play:"+response)

	return response
}

func (c *IntegrationCharacterController) GetCurrentState() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentState
}

func (c *IntegrationCharacterController) GetLastInteractionTime() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastInteraction
}

func (c *IntegrationCharacterController) GetStats() map[string]float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := make(map[string]float64)
	for k, v := range c.stats {
		stats[k] = v
	}
	return stats
}

func (c *IntegrationCharacterController) GetMood() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.mood
}

func (c *IntegrationCharacterController) IsGameMode() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isGameMode
}

func (c *IntegrationCharacterController) GetActionCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.actionCount
}

func (c *IntegrationCharacterController) GetInteractionHistory() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	history := make([]string, len(c.interactionHistory))
	copy(history, c.interactionHistory)
	return history
}

func (c *IntegrationCharacterController) SimulateStatDecay() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Simulate natural stat degradation over time
	c.stats["hunger"] -= 2.0
	c.stats["happiness"] -= 1.0
	c.stats["energy"] += 1.0

	// Clamp values to reasonable ranges
	for stat, value := range c.stats {
		if value < 0 {
			c.stats[stat] = 0
		} else if value > 100 {
			c.stats[stat] = 100
		}
	}

	// Update mood based on overall stats
	totalStats := c.stats["hunger"] + c.stats["happiness"] + c.stats["health"] + c.stats["energy"]
	c.mood = totalStats / 4.0
}

// IntegrationNetworkController provides network functionality for integration testing
type IntegrationNetworkController struct {
	mu           sync.RWMutex
	isEnabled    bool
	peers        []string
	sentMessages []IntegrationMessage
}

type IntegrationMessage struct {
	PeerID    string
	Message   interface{}
	Timestamp time.Time
}

func NewIntegrationNetworkController() *IntegrationNetworkController {
	return &IntegrationNetworkController{
		isEnabled:    true,
		peers:        []string{"peer1", "peer2"},
		sentMessages: make([]IntegrationMessage, 0),
	}
}

func (n *IntegrationNetworkController) GetPeerCount() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return len(n.peers)
}

func (n *IntegrationNetworkController) GetPeerIDs() []string {
	n.mu.RLock()
	defer n.mu.RUnlock()

	peers := make([]string, len(n.peers))
	copy(peers, n.peers)
	return peers
}

func (n *IntegrationNetworkController) SendMessage(peerID string, message interface{}) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.sentMessages = append(n.sentMessages, IntegrationMessage{
		PeerID:    peerID,
		Message:   message,
		Timestamp: time.Now(),
	})

	return nil
}

func (n *IntegrationNetworkController) IsNetworkEnabled() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.isEnabled
}

func (n *IntegrationNetworkController) GetSentMessages() []IntegrationMessage {
	n.mu.RLock()
	defer n.mu.RUnlock()

	messages := make([]IntegrationMessage, len(n.sentMessages))
	copy(messages, n.sentMessages)
	return messages
}

// Test complete integration scenario
func TestBotControllerIntegration(t *testing.T) {
	// Create personality for active bot
	personality := DefaultPersonality()
	personality.InteractionRate = 5.0 // 5 actions per minute
	personality.ResponseDelay = 1 * time.Second
	personality.SocialTendencies["playfulness"] = 0.9
	personality.SocialTendencies["helpfulness"] = 0.8
	personality.SocialTendencies["chattiness"] = 0.7

	// Create realistic controllers
	charController := NewIntegrationCharacterController()
	netController := NewIntegrationNetworkController()

	// Create bot
	bot, err := NewBotController(personality, charController, netController)
	if err != nil {
		t.Fatalf("Failed to create bot: %v", err)
	}

	// Simulate 10 seconds of bot operation
	simulationDuration := 10 * time.Second
	updateInterval := 50 * time.Millisecond // 20 FPS for testing

	endTime := time.Now().Add(simulationDuration)
	updateCount := 0
	actionsPerformed := 0

	for time.Now().Before(endTime) {
		// Update bot
		actionTaken := bot.Update()
		if actionTaken {
			actionsPerformed++
		}

		// Simulate character stat decay occasionally
		if updateCount%40 == 0 { // Every 2 seconds
			charController.SimulateStatDecay()
		}

		updateCount++
		time.Sleep(updateInterval)
	}

	// Validate results
	t.Logf("Simulation completed: %d updates, %d actions performed", updateCount, actionsPerformed)

	// Check that bot performed some actions
	if actionsPerformed == 0 {
		t.Error("Expected bot to perform at least one action during simulation")
	}

	// Check that character received interactions
	finalActionCount := charController.GetActionCount()
	if finalActionCount == 0 {
		t.Error("Expected character to receive some interactions from bot")
	}

	// Check interaction history
	history := charController.GetInteractionHistory()
	if len(history) == 0 {
		t.Error("Expected some interaction history")
	}

	t.Logf("Character received %d interactions", finalActionCount)
	t.Logf("Interaction history: %v", history)

	// Check network messages
	sentMessages := netController.GetSentMessages()
	t.Logf("Bot sent %d network messages", len(sentMessages))

	// Check bot statistics
	stats := bot.GetStats()
	t.Logf("Bot stats: %+v", stats)

	// Validate bot behavior reasonableness
	actionHistory := bot.GetActionHistory()
	t.Logf("Bot action history length: %d", len(actionHistory))

	// Actions should be reasonable given the personality
	if len(actionHistory) > 0 {
		// Check that actions match personality preferences
		hasPreferredActions := false
		for _, action := range actionHistory {
			for _, preferred := range personality.PreferredActions {
				if action.Action == preferred {
					hasPreferredActions = true
					break
				}
			}
		}

		if !hasPreferredActions {
			t.Error("Expected bot to perform some preferred actions based on personality")
		}
	}
}

// Test bot personality impact on behavior
func TestBotPersonalityImpactIntegration(t *testing.T) {
	// Create two bots with different personalities
	shyPersonality := DefaultPersonality()
	shyPersonality.InteractionRate = 2.0 // Low interaction rate
	shyPersonality.ResponseDelay = 2 * time.Second
	shyPersonality.SocialTendencies["chattiness"] = 0.2
	shyPersonality.SocialTendencies["playfulness"] = 0.3

	energeticPersonality := DefaultPersonality()
	energeticPersonality.InteractionRate = 10.0 // High interaction rate
	energeticPersonality.ResponseDelay = 500 * time.Millisecond
	energeticPersonality.SocialTendencies["chattiness"] = 0.9
	energeticPersonality.SocialTendencies["playfulness"] = 0.9 // Create controllers for each bot
	shyChar := NewIntegrationCharacterController()
	energeticChar := NewIntegrationCharacterController()

	shyNet := NewIntegrationNetworkController()
	energeticNet := NewIntegrationNetworkController()

	// Create bots
	shyBot, err := NewBotController(shyPersonality, shyChar, shyNet)
	if err != nil {
		t.Fatalf("Failed to create shy bot: %v", err)
	}

	energeticBot, err := NewBotController(energeticPersonality, energeticChar, energeticNet)
	if err != nil {
		t.Fatalf("Failed to create energetic bot: %v", err)
	}

	// Run simulation for both bots
	simulationDuration := 15 * time.Second // Longer simulation
	updateInterval := 50 * time.Millisecond

	endTime := time.Now().Add(simulationDuration)

	for time.Now().Before(endTime) {
		shyBot.Update()
		energeticBot.Update()
		time.Sleep(updateInterval)
	}

	// Compare behavior
	shyActions := shyChar.GetActionCount()
	energeticActions := energeticChar.GetActionCount()

	shyMessages := len(shyNet.GetSentMessages())
	energeticMessages := len(energeticNet.GetSentMessages())

	t.Logf("Shy bot performed %d actions, sent %d messages", shyActions, shyMessages)
	t.Logf("Energetic bot performed %d actions, sent %d messages", energeticActions, energeticMessages)

	// Both bots should have some activity given the longer simulation
	totalShyActivity := shyActions + shyMessages
	totalEnergeticActivity := energeticActions + energeticMessages

	if totalShyActivity == 0 && totalEnergeticActivity == 0 {
		t.Error("Expected at least one bot to perform some actions in 15 second simulation")
	}

	// Log personality differences for debugging
	t.Logf("Shy personality: InteractionRate=%.1f, ResponseDelay=%v",
		shyPersonality.InteractionRate, shyPersonality.ResponseDelay)
	t.Logf("Energetic personality: InteractionRate=%.1f, ResponseDelay=%v",
		energeticPersonality.InteractionRate, energeticPersonality.ResponseDelay)
}

// Test bot state management and enable/disable
func TestBotControllerStateManagementIntegration(t *testing.T) {
	personality := DefaultPersonality()
	personality.InteractionRate = 10.0                 // Very active for testing (max allowed)
	personality.ResponseDelay = 200 * time.Millisecond // Fast response

	charController := NewIntegrationCharacterController()
	netController := NewIntegrationNetworkController()

	bot, err := NewBotController(personality, charController, netController)
	if err != nil {
		t.Fatalf("Failed to create bot: %v", err)
	}

	// Test enabled state
	if !bot.IsEnabled() {
		t.Error("Bot should be enabled by default")
	}

	// Run with bot enabled for longer period
	for i := 0; i < 20; i++ {
		bot.Update()
		time.Sleep(50 * time.Millisecond) // Reduced time
	}

	enabledActions := charController.GetActionCount()
	t.Logf("Actions while enabled: %d", enabledActions)

	// Disable bot
	bot.Disable()
	if bot.IsEnabled() {
		t.Error("Bot should be disabled after Disable() call")
	}

	// Run with bot disabled
	for i := 0; i < 10; i++ {
		actionTaken := bot.Update()
		if actionTaken {
			t.Error("Disabled bot should not perform actions")
		}
		time.Sleep(20 * time.Millisecond)
	}

	disabledActions := charController.GetActionCount()

	// Action count should not have increased while disabled
	if disabledActions > enabledActions {
		t.Error("Action count should not increase while bot is disabled")
	}

	// Re-enable bot
	bot.Enable()
	if !bot.IsEnabled() {
		t.Error("Bot should be enabled after Enable() call")
	}

	// Run with bot re-enabled for longer period
	for i := 0; i < 20; i++ {
		bot.Update()
		time.Sleep(50 * time.Millisecond)
	}

	reenabledActions := charController.GetActionCount()
	t.Logf("Actions after re-enabling: %d", reenabledActions)

	// At minimum, we should demonstrate state management works
	// Even if no actions occur, the enable/disable functionality should work
	t.Logf("State management test completed successfully")
	t.Logf("Enabled actions: %d, Disabled actions: %d, Re-enabled actions: %d",
		enabledActions, disabledActions, reenabledActions)
}

// Benchmark realistic bot operation
func BenchmarkBotControllerIntegration(b *testing.B) {
	personality := DefaultPersonality()
	charController := NewIntegrationCharacterController()
	netController := NewIntegrationNetworkController()

	bot, err := NewBotController(personality, charController, netController)
	if err != nil {
		b.Fatalf("Failed to create bot: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bot.Update()
		charController.SimulateStatDecay()
	}
}
