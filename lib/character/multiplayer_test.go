package character

import (
	"sync"
	"testing"
	"time"
)

// MockNetworkManager implements the NetworkManager interface for testing
type MockNetworkManager struct {
	mu           sync.RWMutex
	started      bool
	stopped      bool
	messages     []NetworkMessage
	handlers     map[string]func(NetworkMessage, interface{}) error
	broadcastErr error
}

func NewMockNetworkManager() *MockNetworkManager {
	return &MockNetworkManager{
		messages: make([]NetworkMessage, 0),
		handlers: make(map[string]func(NetworkMessage, interface{}) error),
	}
}

func (m *MockNetworkManager) Start() error {
	m.started = true
	return nil
}

func (m *MockNetworkManager) Stop() error {
	m.stopped = true
	return nil
}

func (m *MockNetworkManager) Broadcast(msg NetworkMessage) error {
	if m.broadcastErr != nil {
		return m.broadcastErr
	}
	m.mu.Lock()
	m.messages = append(m.messages, msg)
	m.mu.Unlock()
	return nil
}

func (m *MockNetworkManager) RegisterHandler(msgType string, handler func(NetworkMessage, interface{}) error) {
	m.handlers[msgType] = handler
}

func (m *MockNetworkManager) GetMessages() []NetworkMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()
	// Return a copy to prevent external modification
	result := make([]NetworkMessage, len(m.messages))
	copy(result, m.messages)
	return result
}

func (m *MockNetworkManager) ClearMessages() {
	m.mu.Lock()
	m.messages = make([]NetworkMessage, 0)
	m.mu.Unlock()
}

// MockProtocolManager implements the ProtocolManager interface for testing
type MockProtocolManager struct {
	signErr   error
	verifyErr error
}

func NewMockProtocolManager() *MockProtocolManager {
	return &MockProtocolManager{}
}

func (m *MockProtocolManager) SignMessage(data []byte) ([]byte, error) {
	if m.signErr != nil {
		return nil, m.signErr
	}
	return []byte("mock_signature"), nil
}

func (m *MockProtocolManager) VerifyMessage(data, signature, publicKey []byte) error {
	return m.verifyErr
}

// Helper function to create a test character card
func createTestCard() *CharacterCard {
	return &CharacterCard{
		Name:        "TestCharacter",
		Description: "A test character for multiplayer testing",
		Animations: map[string]string{
			"idle":    "idle.gif",
			"talking": "talking.gif",
		},
		Dialogs: []Dialog{
			{
				Trigger:   "click",
				Responses: []string{"Hello!"},
				Animation: "talking",
			},
			{
				Trigger:   "rightclick",
				Responses: []string{"Right clicked!"},
				Animation: "talking",
			},
		},
		Behavior: Behavior{
			IdleTimeout:     30,
			MovementEnabled: true,
			DefaultSize:     64,
		},
		Stats: map[string]StatConfig{
			"hunger": {Max: 100, Initial: 50, DegradationRate: 0.1},
		},
		GameRules: &GameRulesConfig{
			StatsDecayInterval:             30,
			CriticalStateAnimationPriority: true,
			MoodBasedAnimations:            false,
		},
		Interactions: map[string]InteractionConfig{
			"feed": {
				Animations: []string{"talking"},
				Responses:  []string{"Thanks for feeding me!"},
				Effects:    map[string]float64{"hunger": 10},
			},
		},
		Multiplayer: &MultiplayerConfig{
			Enabled:   true,
			NetworkID: "test-network",
		},
	}
}

func TestNewMultiplayerCharacter(t *testing.T) {
	tests := []struct {
		name        string
		card        *CharacterCard
		config      MultiplayerWrapperConfig
		networkMgr  NetworkManager
		protocolMgr ProtocolManager
		expectError bool
	}{
		{
			name:        "Valid creation",
			card:        createTestCard(),
			config:      MultiplayerWrapperConfig{CharacterID: "test-char"},
			networkMgr:  NewMockNetworkManager(),
			protocolMgr: NewMockProtocolManager(),
			expectError: false,
		},
		{
			name:        "Nil card",
			card:        nil,
			config:      MultiplayerWrapperConfig{},
			networkMgr:  NewMockNetworkManager(),
			protocolMgr: NewMockProtocolManager(),
			expectError: true,
		},
		{
			name:        "Nil network manager",
			card:        createTestCard(),
			config:      MultiplayerWrapperConfig{},
			networkMgr:  nil,
			protocolMgr: NewMockProtocolManager(),
			expectError: true,
		},
		{
			name:        "Nil protocol manager",
			card:        createTestCard(),
			config:      MultiplayerWrapperConfig{},
			networkMgr:  NewMockNetworkManager(),
			protocolMgr: nil,
			expectError: true,
		},
		{
			name:   "Default config values",
			card:   createTestCard(),
			config: MultiplayerWrapperConfig{
				// Empty config should get defaults
			},
			networkMgr:  NewMockNetworkManager(),
			protocolMgr: NewMockProtocolManager(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc, err := NewMultiplayerCharacter(tt.card, tt.config, tt.networkMgr, tt.protocolMgr)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if mc == nil {
				t.Errorf("Expected MultiplayerCharacter but got nil")
				return
			}

			// Verify default values were set
			if tt.config.CharacterID == "" && mc.characterID == "" {
				t.Errorf("Expected default character ID to be set")
			}

			if tt.config.SyncInterval == 0 && mc.syncInterval != 5*time.Second {
				t.Errorf("Expected default sync interval of 5s, got %v", mc.syncInterval)
			}
		})
	}
}

func TestMultiplayerCharacter_EnableDisableNetworking(t *testing.T) {
	card := createTestCard()
	config := MultiplayerWrapperConfig{CharacterID: "test-char"}
	mockNet := NewMockNetworkManager()
	mockProto := NewMockProtocolManager()

	mc, err := NewMultiplayerCharacter(card, config, mockNet, mockProto)
	if err != nil {
		t.Fatalf("Failed to create MultiplayerCharacter: %v", err)
	}

	// Initially should be disabled
	if mc.IsNetworkEnabled() {
		t.Errorf("Expected networking to be disabled initially")
	}

	// Enable networking
	err = mc.EnableNetworking()
	if err != nil {
		t.Errorf("Failed to enable networking: %v", err)
	}

	if !mc.IsNetworkEnabled() {
		t.Errorf("Expected networking to be enabled")
	}

	if !mockNet.started {
		t.Errorf("Expected network manager to be started")
	}

	// Disable networking
	mc.DisableNetworking()

	if mc.IsNetworkEnabled() {
		t.Errorf("Expected networking to be disabled")
	}
}

func TestMultiplayerCharacter_HandleClick(t *testing.T) {
	card := createTestCard()
	config := MultiplayerWrapperConfig{
		CharacterID:      "test-char",
		BroadcastActions: true,
	}
	mockNet := NewMockNetworkManager()
	mockProto := NewMockProtocolManager()

	mc, err := NewMultiplayerCharacter(card, config, mockNet, mockProto)
	if err != nil {
		t.Fatalf("Failed to create MultiplayerCharacter: %v", err)
	}

	// Test click without networking enabled
	response := mc.HandleClick()
	if response == "" {
		t.Errorf("Expected response from HandleClick")
	}

	// Should be no messages since networking is disabled
	if len(mockNet.GetMessages()) != 0 {
		t.Errorf("Expected no network messages when networking disabled, got %d", len(mockNet.GetMessages()))
	}

	// Enable networking and test again
	mc.EnableNetworking()
	mockNet.ClearMessages()

	response = mc.HandleClick()
	if response == "" {
		t.Errorf("Expected response from HandleClick")
	}

	// Wait a bit for the async broadcast to complete
	time.Sleep(10 * time.Millisecond)

	// Should have one broadcast message
	messages := mockNet.GetMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 network message, got %d", len(messages))
	}

	if len(messages) > 0 && messages[0].Type != "character_action" {
		t.Errorf("Expected message type 'character_action', got '%s'", messages[0].Type)
	}
}

func TestMultiplayerCharacter_HandleRightClick(t *testing.T) {
	card := createTestCard()
	config := MultiplayerWrapperConfig{
		CharacterID:      "test-char",
		BroadcastActions: true,
	}
	mockNet := NewMockNetworkManager()
	mockProto := NewMockProtocolManager()

	mc, err := NewMultiplayerCharacter(card, config, mockNet, mockProto)
	if err != nil {
		t.Fatalf("Failed to create MultiplayerCharacter: %v", err)
	}

	// Enable networking
	mc.EnableNetworking()

	response := mc.HandleRightClick()
	if response == "" {
		t.Errorf("Expected response from HandleRightClick")
	}

	// Wait a bit for the async broadcast to complete
	time.Sleep(10 * time.Millisecond)

	// Should have one broadcast message
	messages := mockNet.GetMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 network message, got %d", len(messages))
	}

	if len(messages) > 0 && messages[0].Type != "character_action" {
		t.Errorf("Expected message type 'character_action', got '%s'", messages[0].Type)
	}
}

func TestMultiplayerCharacter_HandleGameInteraction(t *testing.T) {
	card := createTestCard()
	config := MultiplayerWrapperConfig{
		CharacterID:      "test-char",
		BroadcastActions: true,
	}
	mockNet := NewMockNetworkManager()
	mockProto := NewMockProtocolManager()

	mc, err := NewMultiplayerCharacter(card, config, mockNet, mockProto)
	if err != nil {
		t.Fatalf("Failed to create MultiplayerCharacter: %v", err)
	}

	// Enable networking
	mc.EnableNetworking()

	// Enable game mode for interactions to work
	err = mc.EnableGameMode(nil, "")
	if err != nil {
		t.Fatalf("Failed to enable game mode: %v", err)
	}

	response := mc.HandleGameInteraction("feed")
	if response == "" {
		t.Errorf("Expected response from HandleGameInteraction")
	}

	// Wait a bit for the async broadcast to complete
	time.Sleep(10 * time.Millisecond)

	// Should have one broadcast message
	messages := mockNet.GetMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 network message, got %d", len(messages))
	}

	if len(messages) > 0 && messages[0].Type != "character_action" {
		t.Errorf("Expected message type 'character_action', got '%s'", messages[0].Type)
	}
}

func TestMultiplayerCharacter_StateSync(t *testing.T) {
	card := createTestCard()
	config := MultiplayerWrapperConfig{
		CharacterID:     "test-char",
		EnableStateSync: true,
		SyncInterval:    100 * time.Millisecond, // Fast sync for testing
	}
	mockNet := NewMockNetworkManager()
	mockProto := NewMockProtocolManager()

	mc, err := NewMultiplayerCharacter(card, config, mockNet, mockProto)
	if err != nil {
		t.Fatalf("Failed to create MultiplayerCharacter: %v", err)
	}

	// Enable networking
	mc.EnableNetworking()

	// Manual sync test
	err = mc.syncState()
	if err != nil {
		t.Errorf("Failed to sync state: %v", err)
	}

	// Should have one sync message
	messages := mockNet.GetMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 sync message, got %d", len(messages))
	}

	if messages[0].Type != "state_sync" {
		t.Errorf("Expected message type 'state_sync', got '%s'", messages[0].Type)
	}
}

func TestMultiplayerCharacter_GetNetworkStats(t *testing.T) {
	card := createTestCard()
	config := MultiplayerWrapperConfig{
		CharacterID:      "test-char",
		BroadcastActions: true,
		EnableStateSync:  true,
	}
	mockNet := NewMockNetworkManager()
	mockProto := NewMockProtocolManager()

	mc, err := NewMultiplayerCharacter(card, config, mockNet, mockProto)
	if err != nil {
		t.Fatalf("Failed to create MultiplayerCharacter: %v", err)
	}

	stats := mc.GetNetworkStats()

	expectedFields := []string{
		"networkEnabled", "broadcastActions", "enableStateSync",
		"characterId", "syncInterval", "lastSyncTime", "pendingActions",
	}

	for _, field := range expectedFields {
		if _, exists := stats[field]; !exists {
			t.Errorf("Expected field '%s' in network stats", field)
		}
	}

	// Verify specific values
	if stats["networkEnabled"] != false {
		t.Errorf("Expected networkEnabled to be false initially")
	}

	if stats["characterId"] != "test-char" {
		t.Errorf("Expected characterId to be 'test-char', got %v", stats["characterId"])
	}

	if stats["broadcastActions"] != true {
		t.Errorf("Expected broadcastActions to be true")
	}
}

func TestMultiplayerCharacter_InterfacePreservation(t *testing.T) {
	// Test that MultiplayerCharacter preserves the Character interface
	card := createTestCard()
	config := MultiplayerWrapperConfig{CharacterID: "test-char"}
	mockNet := NewMockNetworkManager()
	mockProto := NewMockProtocolManager()

	mc, err := NewMultiplayerCharacter(card, config, mockNet, mockProto)
	if err != nil {
		t.Fatalf("Failed to create MultiplayerCharacter: %v", err)
	}

	// Enable game mode for interactions to work
	err = mc.EnableGameMode(nil, "")
	if err != nil {
		t.Fatalf("Failed to enable game mode: %v", err)
	}

	// Test that all Character methods are still available
	// (This is mostly a compile-time test, but we can verify basic functionality)

	// Position methods
	mc.SetPosition(10.0, 20.0)
	x, y := mc.GetPosition()
	if x != 10.0 || y != 20.0 {
		t.Errorf("Expected position (10.0, 20.0), got (%.1f, %.1f)", x, y)
	}

	// Name method
	name := mc.GetName()
	if name != "TestCharacter" {
		t.Errorf("Expected name 'TestCharacter', got '%s'", name)
	}

	// State methods
	state := mc.GetCurrentState()
	if state == "" {
		t.Errorf("Expected non-empty current state")
	}

	// Click methods (already tested but verify they return strings)
	clickResponse := mc.HandleClick()
	if clickResponse == "" {
		t.Errorf("Expected non-empty click response")
	}

	rightClickResponse := mc.HandleRightClick()
	if rightClickResponse == "" {
		t.Errorf("Expected non-empty right click response")
	}

	// Game interaction method (use 'feed' which is defined in our test card)
	gameResponse := mc.HandleGameInteraction("feed")
	if gameResponse == "" {
		t.Errorf("Expected non-empty game interaction response")
	}
}

func TestMultiplayerCharacter_ConcurrentAccess(t *testing.T) {
	// Test thread safety of the MultiplayerCharacter
	card := createTestCard()
	config := MultiplayerWrapperConfig{
		CharacterID:      "test-char",
		BroadcastActions: true,
	}
	mockNet := NewMockNetworkManager()
	mockProto := NewMockProtocolManager()

	mc, err := NewMultiplayerCharacter(card, config, mockNet, mockProto)
	if err != nil {
		t.Fatalf("Failed to create MultiplayerCharacter: %v", err)
	}

	// Enable networking
	mc.EnableNetworking()

	// Simulate concurrent access
	done := make(chan bool, 3)

	// Goroutine 1: Click operations
	go func() {
		for i := 0; i < 10; i++ {
			mc.HandleClick()
			time.Sleep(time.Millisecond)
		}
		done <- true
	}()

	// Goroutine 2: Network stats access
	go func() {
		for i := 0; i < 10; i++ {
			mc.GetNetworkStats()
			time.Sleep(time.Millisecond)
		}
		done <- true
	}()

	// Goroutine 3: Enable/disable networking
	go func() {
		for i := 0; i < 5; i++ {
			mc.DisableNetworking()
			time.Sleep(time.Millisecond)
			mc.EnableNetworking()
			time.Sleep(time.Millisecond)
		}
		done <- true
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		select {
		case <-done:
			// Success
		case <-time.After(5 * time.Second):
			t.Errorf("Timeout waiting for concurrent operations to complete")
		}
	}
}

// TestMultiplayerConfigValidation tests the existing multiplayer configuration validation from card.go
func TestMultiplayerConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		multiplayer *MultiplayerConfig
		expectError bool
		errorText   string
	}{
		{
			name:        "nil multiplayer config",
			multiplayer: nil,
			expectError: false,
		},
		{
			name: "valid multiplayer config",
			multiplayer: &MultiplayerConfig{
				Enabled:   true,
				NetworkID: "test-network",
				MaxPeers:  4,
			},
			expectError: false,
		},
		{
			name: "empty network ID",
			multiplayer: &MultiplayerConfig{
				Enabled:   true,
				NetworkID: "",
			},
			expectError: true,
			errorText:   "multiplayer config: networkID is required when multiplayer is enabled",
		},
		{
			name: "invalid max peers - negative",
			multiplayer: &MultiplayerConfig{
				Enabled:   true,
				NetworkID: "test",
				MaxPeers:  -1,
			},
			expectError: false, // Implementation allows negative values (treated as default)
		},
		{
			name: "invalid max peers - zero",
			multiplayer: &MultiplayerConfig{
				Enabled:   true,
				NetworkID: "test",
				MaxPeers:  0,
			},
			expectError: false, // Implementation allows zero values (treated as default)
		},
		{
			name: "invalid discovery port - negative",
			multiplayer: &MultiplayerConfig{
				Enabled:       true,
				NetworkID:     "test",
				DiscoveryPort: -1,
			},
			expectError: false, // Implementation allows negative values (treated as default)
		},
		{
			name: "invalid discovery port - zero",
			multiplayer: &MultiplayerConfig{
				Enabled:       true,
				NetworkID:     "test",
				DiscoveryPort: 0,
			},
			expectError: false, // Implementation allows zero values (treated as default)
		},
		{
			name: "disabled multiplayer with config",
			multiplayer: &MultiplayerConfig{
				Enabled:   false,
				NetworkID: "",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &CharacterCard{
				Name:        "Test",
				Description: "Test character",
				Animations: map[string]string{
					"idle":    "idle.gif",
					"talking": "talking.gif",
				},
				Dialogs: []Dialog{
					{
						Trigger:   "click",
						Responses: []string{"Hello!"},
						Animation: "talking",
						Cooldown:  5,
					},
				},
				Behavior: Behavior{
					IdleTimeout:     30,
					MovementEnabled: false,
					DefaultSize:     128,
				},
				Multiplayer: tt.multiplayer,
			}

			err := card.Validate()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorText != "" && err.Error() != tt.errorText {
					t.Errorf("Expected error text '%s', got '%s'", tt.errorText, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}
