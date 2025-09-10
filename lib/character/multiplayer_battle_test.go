// Unit tests for multiplayer battle integration
// Tests the MultiplayerCharacter battle extensions
// WHY: Ensures battle invitation and action handling works correctly in multiplayer

package character

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/opd-ai/desktop-companion/lib/battle"
	"github.com/opd-ai/desktop-companion/lib/network"
)

// mockBattleNetworkManager implements NetworkManager for testing
type mockBattleNetworkManager struct {
	messages []NetworkMessage
	handlers map[string]func(NetworkMessage, interface{}) error
}

func newMockBattleNetworkManager() *mockBattleNetworkManager {
	return &mockBattleNetworkManager{
		messages: make([]NetworkMessage, 0),
		handlers: make(map[string]func(NetworkMessage, interface{}) error),
	}
}

func (m *mockBattleNetworkManager) Start() error { return nil }
func (m *mockBattleNetworkManager) Stop() error  { return nil }
func (m *mockBattleNetworkManager) Broadcast(msg NetworkMessage) error {
	m.messages = append(m.messages, msg)
	return nil
}

func (m *mockBattleNetworkManager) RegisterHandler(msgType string, handler func(NetworkMessage, interface{}) error) {
	m.handlers[msgType] = handler
}

func (m *mockBattleNetworkManager) GetLastMessage() *NetworkMessage {
	if len(m.messages) == 0 {
		return nil
	}
	return &m.messages[len(m.messages)-1]
}

// mockBattleProtocolManager implements ProtocolManager for testing
type mockBattleProtocolManager struct{}

func (m *mockBattleProtocolManager) SignMessage(data []byte) ([]byte, error) {
	return data, nil
}

func (m *mockBattleProtocolManager) VerifyMessage(data, signature, publicKey []byte) error {
	return nil
}

func TestMultiplayerCharacter_InitiateBattle(t *testing.T) {
	// Create a test character card
	card := &CharacterCard{
		Name: "TestChar",
		Stats: map[string]StatConfig{
			"happiness": {Initial: 50.0, Max: 100.0},
		},
	}

	// Create mock network and protocol managers
	mockNetwork := newMockBattleNetworkManager()
	mockProtocol := &mockBattleProtocolManager{}

	// Create multiplayer character
	config := MultiplayerWrapperConfig{
		CharacterID:      "test-char-1",
		BroadcastActions: true,
		EnableStateSync:  true,
	}

	mc, err := NewMultiplayerCharacter(card, config, mockNetwork, mockProtocol)
	if err != nil {
		t.Fatalf("Failed to create multiplayer character: %v", err)
	}

	// Enable networking
	mc.mu.Lock()
	mc.networkEnabled = true
	mc.mu.Unlock()

	// Test battle initiation
	err = mc.InitiateBattle("target-peer-123")
	if err != nil {
		t.Fatalf("Failed to initiate battle: %v", err)
	}

	// Verify message was sent
	lastMsg := mockNetwork.GetLastMessage()
	if lastMsg == nil {
		t.Fatal("No message was sent")
	}

	if lastMsg.Type != "battle_invite" {
		t.Errorf("Expected message type 'battle_invite', got '%s'", lastMsg.Type)
	}

	if lastMsg.From != "test-char-1" {
		t.Errorf("Expected message from 'test-char-1', got '%s'", lastMsg.From)
	}

	if lastMsg.To != "target-peer-123" {
		t.Errorf("Expected message to 'target-peer-123', got '%s'", lastMsg.To)
	}

	// Verify payload
	var payload network.BattleInvitePayload
	err = json.Unmarshal(lastMsg.Payload, &payload)
	if err != nil {
		t.Fatalf("Failed to unmarshal battle invite payload: %v", err)
	}

	if payload.FromCharacterID != "test-char-1" {
		t.Errorf("Expected FromCharacterID 'test-char-1', got '%s'", payload.FromCharacterID)
	}

	if payload.ToCharacterID != "target-peer-123" {
		t.Errorf("Expected ToCharacterID 'target-peer-123', got '%s'", payload.ToCharacterID)
	}
}

func TestMultiplayerCharacter_HandleBattleInvite(t *testing.T) {
	// Create a test character card
	card := &CharacterCard{
		Name: "TestChar",
		Stats: map[string]StatConfig{
			"happiness": {Initial: 50.0, Max: 100.0},
		},
	}

	// Create mock network and protocol managers
	mockNetwork := newMockBattleNetworkManager()
	mockProtocol := &mockBattleProtocolManager{}

	// Create multiplayer character
	config := MultiplayerWrapperConfig{
		CharacterID:      "test-char-2",
		BroadcastActions: true,
		EnableStateSync:  true,
	}

	mc, err := NewMultiplayerCharacter(card, config, mockNetwork, mockProtocol)
	if err != nil {
		t.Fatalf("Failed to create multiplayer character: %v", err)
	}

	// Test battle invite handling
	invite := network.BattleInvitePayload{
		FromCharacterID: "opponent-123",
		ToCharacterID:   "test-char-2",
		BattleID:        "battle-456",
		Timestamp:       time.Now(),
	}

	err = mc.HandleBattleInvite(invite)
	if err != nil {
		t.Errorf("HandleBattleInvite failed: %v", err)
	}
}

func TestMultiplayerCharacter_PerformBattleAction(t *testing.T) {
	// Create a test character card
	card := &CharacterCard{
		Name: "TestChar",
		Stats: map[string]StatConfig{
			"happiness": {Initial: 50.0, Max: 100.0},
		},
	}

	// Create mock network and protocol managers
	mockNetwork := newMockBattleNetworkManager()
	mockProtocol := &mockBattleProtocolManager{}

	// Create multiplayer character
	config := MultiplayerWrapperConfig{
		CharacterID:      "test-char-3",
		BroadcastActions: true,
		EnableStateSync:  true,
	}

	mc, err := NewMultiplayerCharacter(card, config, mockNetwork, mockProtocol)
	if err != nil {
		t.Fatalf("Failed to create multiplayer character: %v", err)
	}

	// Enable networking
	mc.mu.Lock()
	mc.networkEnabled = true
	mc.mu.Unlock()

	// First initiate a battle to set currentBattleID
	err = mc.InitiateBattle("opponent-456")
	if err != nil {
		t.Fatalf("Failed to initiate battle: %v", err)
	}

	// Test battle action
	action := battle.BattleAction{
		Type:     battle.ACTION_ATTACK,
		ActorID:  "test-char-3",
		TargetID: "opponent-456",
		ItemUsed: "",
	}

	err = mc.PerformBattleAction(action)
	if err != nil {
		t.Fatalf("Failed to perform battle action: %v", err)
	}

	// Verify message was sent
	lastMsg := mockNetwork.GetLastMessage()
	if lastMsg == nil {
		t.Fatal("No message was sent")
	}

	if lastMsg.Type != "battle_action" {
		t.Errorf("Expected message type 'battle_action', got '%s'", lastMsg.Type)
	}

	// Verify payload
	var payload network.BattleActionPayload
	err = json.Unmarshal(lastMsg.Payload, &payload)
	if err != nil {
		t.Fatalf("Failed to unmarshal battle action payload: %v", err)
	}

	if payload.ActionType != "attack" {
		t.Errorf("Expected ActionType 'attack', got '%s'", payload.ActionType)
	}

	if payload.ActorID != "test-char-3" {
		t.Errorf("Expected ActorID 'test-char-3', got '%s'", payload.ActorID)
	}
}

func TestMultiplayerCharacter_SetupBattleHandlers(t *testing.T) {
	// Create a test character card
	card := &CharacterCard{
		Name: "TestChar",
		Stats: map[string]StatConfig{
			"happiness": {Initial: 50.0, Max: 100.0},
		},
	}

	// Create mock network and protocol managers
	mockNetwork := newMockBattleNetworkManager()
	mockProtocol := &mockBattleProtocolManager{}

	// Create multiplayer character
	config := MultiplayerWrapperConfig{
		CharacterID:      "test-char-4",
		BroadcastActions: true,
		EnableStateSync:  true,
	}

	mc, err := NewMultiplayerCharacter(card, config, mockNetwork, mockProtocol)
	if err != nil {
		t.Fatalf("Failed to create multiplayer character: %v", err)
	}

	// Test handler setup
	err = mc.setupBattleHandlers()
	if err != nil {
		t.Fatalf("Failed to setup battle handlers: %v", err)
	}

	// Verify handlers were registered
	expectedHandlers := []string{"battle_invite", "battle_action", "battle_result", "battle_end"}
	for _, handlerType := range expectedHandlers {
		if _, exists := mockNetwork.handlers[handlerType]; !exists {
			t.Errorf("Handler '%s' was not registered", handlerType)
		}
	}
}

func TestMultiplayerCharacter_BattleHandlerErrors(t *testing.T) {
	// Create a test character card
	card := &CharacterCard{
		Name: "TestChar",
		Stats: map[string]StatConfig{
			"happiness": {Initial: 50.0, Max: 100.0},
		},
	}

	// Create mock network and protocol managers
	mockNetwork := newMockBattleNetworkManager()
	mockProtocol := &mockBattleProtocolManager{}

	// Create multiplayer character
	config := MultiplayerWrapperConfig{
		CharacterID:      "test-char-5",
		BroadcastActions: true,
		EnableStateSync:  true,
	}

	mc, err := NewMultiplayerCharacter(card, config, mockNetwork, mockProtocol)
	if err != nil {
		t.Fatalf("Failed to create multiplayer character: %v", err)
	}

	// Setup battle handlers
	err = mc.setupBattleHandlers()
	if err != nil {
		t.Fatalf("Failed to setup battle handlers: %v", err)
	}

	// Test battle invite handler with invalid payload
	invalidMsg := NetworkMessage{
		Type:    "battle_invite",
		From:    "test-peer",
		Payload: []byte("invalid json"),
	}

	handler := mockNetwork.handlers["battle_invite"]
	err = handler(invalidMsg, nil)
	if err == nil {
		t.Error("Expected error for invalid battle invite payload")
	}
}
