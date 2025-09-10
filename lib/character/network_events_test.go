package character

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

// MockNetworkInterface implements NetworkInterface for testing
type MockNetworkInterface struct {
	sentMessages     []MockMessage
	messageHandlers  map[string]func([]byte, string) error
	connectedPeers   []string
	localPeerID      string
	broadcastError   error
	sendMessageError error
	mu               sync.RWMutex
}

type MockMessage struct {
	MessageType  string
	Payload      []byte
	TargetPeerID string
	IsBroadcast  bool
}

func NewMockNetworkInterface(localPeerID string) *MockNetworkInterface {
	return &MockNetworkInterface{
		sentMessages:    make([]MockMessage, 0),
		messageHandlers: make(map[string]func([]byte, string) error),
		connectedPeers:  make([]string, 0),
		localPeerID:     localPeerID,
	}
}

func (m *MockNetworkInterface) SendMessage(msgType string, payload []byte, targetPeerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.sendMessageError != nil {
		return m.sendMessageError
	}

	m.sentMessages = append(m.sentMessages, MockMessage{
		MessageType:  msgType,
		Payload:      payload,
		TargetPeerID: targetPeerID,
		IsBroadcast:  false,
	})
	return nil
}

func (m *MockNetworkInterface) RegisterMessageHandler(msgType string, handler func([]byte, string) error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messageHandlers[msgType] = handler
}

func (m *MockNetworkInterface) BroadcastMessage(msgType string, payload []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.broadcastError != nil {
		return m.broadcastError
	}

	m.sentMessages = append(m.sentMessages, MockMessage{
		MessageType: msgType,
		Payload:     payload,
		IsBroadcast: true,
	})
	return nil
}

func (m *MockNetworkInterface) GetConnectedPeers() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]string{}, m.connectedPeers...)
}

func (m *MockNetworkInterface) GetLocalPeerID() string {
	return m.localPeerID
}

func (m *MockNetworkInterface) AddConnectedPeer(peerID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connectedPeers = append(m.connectedPeers, peerID)
}

func (m *MockNetworkInterface) GetSentMessages() []MockMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]MockMessage{}, m.sentMessages...)
}

func (m *MockNetworkInterface) ClearSentMessages() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sentMessages = make([]MockMessage, 0)
}

func (m *MockNetworkInterface) TriggerMessageHandler(msgType string, payload []byte, fromPeerID string) error {
	m.mu.RLock()
	handler, exists := m.messageHandlers[msgType]
	m.mu.RUnlock()

	if !exists {
		return errors.New("no handler registered")
	}
	return handler(payload, fromPeerID)
}

// MockPeerManager implements PeerManagerInterface for testing
type MockPeerManager struct {
	peers           map[string]*PeerInfo
	eventListeners  []func(eventType PeerEventType, peerID string)
	isValidPeerFunc func(string) bool
	mu              sync.RWMutex
}

func NewMockPeerManager() *MockPeerManager {
	return &MockPeerManager{
		peers:           make(map[string]*PeerInfo),
		eventListeners:  make([]func(eventType PeerEventType, peerID string), 0),
		isValidPeerFunc: func(peerID string) bool { return true }, // Default: all peers valid
	}
}

func (m *MockPeerManager) GetPeerInfo(peerID string) (*PeerInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if peerInfo, exists := m.peers[peerID]; exists {
		return peerInfo, nil
	}
	return nil, errors.New("peer not found")
}

func (m *MockPeerManager) IsValidPeer(peerID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isValidPeerFunc(peerID)
}

func (m *MockPeerManager) AddPeerEventListener(callback func(eventType PeerEventType, peerID string)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.eventListeners = append(m.eventListeners, callback)
}

func (m *MockPeerManager) AddPeer(peerID string, peerInfo *PeerInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.peers[peerID] = peerInfo
}

func (m *MockPeerManager) TriggerPeerEvent(eventType PeerEventType, peerID string) {
	m.mu.RLock()
	listeners := append([]func(eventType PeerEventType, peerID string){}, m.eventListeners...)
	m.mu.RUnlock()

	for _, listener := range listeners {
		listener(eventType, peerID)
	}
}

func (m *MockPeerManager) SetValidPeerFunc(fn func(string) bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isValidPeerFunc = fn
}

// Test helper functions
func createTestGeneralEventManager() *GeneralEventManager {
	events := []GeneralDialogEvent{
		{
			RandomEventConfig: RandomEventConfig{
				Name:        "test_conversation",
				Description: "Test conversation",
				Responses:   []string{"Hello!", "How are you?"},
				Effects:     map[string]float64{"happiness": 5.0},
			},
			Category:    "conversation",
			Trigger:     "manual",
			Interactive: true,
			Choices: []EventChoice{
				{
					Text:    "Option 1",
					Effects: map[string]float64{"happiness": 1.0},
				},
				{
					Text:    "Option 2",
					Effects: map[string]float64{"happiness": 2.0},
				},
			},
			Keywords: []string{"friendly", "chat"},
		},
		{
			RandomEventConfig: RandomEventConfig{
				Name:        "group_activity",
				Description: "Group activity",
				Responses:   []string{"Let's play together!"},
				Effects:     map[string]float64{"maxParticipants": 4.0},
			},
			Category:    "group",
			Trigger:     "manual",
			Interactive: true,
			Keywords:    []string{"multiplayer", "group"},
		},
	}
	return NewGeneralEventManager(events, true)
}

func createTestNetworkEventManager() (*NetworkEventManager, *MockNetworkInterface, *MockPeerManager) {
	baseManager := createTestGeneralEventManager()
	mockNetwork := NewMockNetworkInterface("local_peer_123")
	mockPeerManager := NewMockPeerManager()

	nem := NewNetworkEventManager(baseManager, mockNetwork, mockPeerManager, true)
	return nem, mockNetwork, mockPeerManager
}

// Tests for NetworkEventManager

func TestNewNetworkEventManager(t *testing.T) {
	baseManager := createTestGeneralEventManager()
	mockNetwork := NewMockNetworkInterface("test_peer")
	mockPeerManager := NewMockPeerManager()

	nem := NewNetworkEventManager(baseManager, mockNetwork, mockPeerManager, true)

	if nem == nil {
		t.Fatal("Expected non-nil NetworkEventManager")
	}

	if nem.GeneralEventManager == nil {
		t.Error("Expected embedded GeneralEventManager to be set")
	}

	if nem.networkInterface != mockNetwork {
		t.Error("Expected network interface to be set")
	}

	if nem.peerManager != mockPeerManager {
		t.Error("Expected peer manager to be set")
	}

	if !nem.enabled {
		t.Error("Expected NetworkEventManager to be enabled")
	}
}

func TestNewNetworkEventManager_Disabled(t *testing.T) {
	baseManager := createTestGeneralEventManager()
	mockNetwork := NewMockNetworkInterface("test_peer")
	mockPeerManager := NewMockPeerManager()

	nem := NewNetworkEventManager(baseManager, mockNetwork, mockPeerManager, false)

	if nem.enabled {
		t.Error("Expected NetworkEventManager to be disabled")
	}
}

func TestTriggerNetworkEvent_RegularEvent(t *testing.T) {
	nem, _, _ := createTestNetworkEventManager()
	gameState := &GameState{}

	// Trigger a non-network event
	event, err := nem.TriggerNetworkEvent("test_conversation", gameState, nil)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if event == nil {
		t.Fatal("Expected event to be returned")
	}

	if event.Name != "test_conversation" {
		t.Errorf("Expected event name 'test_conversation', got '%s'", event.Name)
	}
}

func TestTriggerNetworkEvent_GroupEvent(t *testing.T) {
	nem, mockNetwork, mockPeerManager := createTestNetworkEventManager()
	gameState := &GameState{}

	// Add some peers
	mockNetwork.AddConnectedPeer("peer1")
	mockNetwork.AddConnectedPeer("peer2")
	mockPeerManager.AddPeer("peer1", &PeerInfo{ID: "peer1", CharacterID: "char1"})
	mockPeerManager.AddPeer("peer2", &PeerInfo{ID: "peer2", CharacterID: "char2"})

	// Trigger a group event
	event, err := nem.TriggerNetworkEvent("group_activity", gameState, []string{"peer1", "peer2"})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if event == nil {
		t.Fatal("Expected event to be returned")
	}

	// Check that invitations were sent
	sentMessages := mockNetwork.GetSentMessages()
	if len(sentMessages) != 2 {
		t.Errorf("Expected 2 invitation messages, got %d", len(sentMessages))
	}

	// Verify group session was created
	sessions := nem.GetActiveGroupSessions()
	if len(sessions) != 1 {
		t.Errorf("Expected 1 active session, got %d", len(sessions))
	}

	for _, session := range sessions {
		if session.EventName != "group_activity" {
			t.Errorf("Expected session event name 'group_activity', got '%s'", session.EventName)
		}
		if len(session.Participants) != 1 { // Only local peer initially
			t.Errorf("Expected 1 participant initially, got %d", len(session.Participants))
		}
	}
}

func TestJoinGroupSession(t *testing.T) {
	nem, _, _ := createTestNetworkEventManager()

	// Create a test session manually
	sessionID := "test_session_123"
	session := &GroupSession{
		ID:               sessionID,
		EventName:        "group_activity",
		Participants:     []string{"local_peer_123"},
		InitiatorID:      "local_peer_123",
		CurrentState:     "waiting",
		MaxParticipants:  4,
		StartTime:        time.Now(),
		LastActivity:     time.Now(),
		VoteChoices:      make(map[string]int),
		ParticipantVotes: make(map[string]int),
	}
	nem.groupSessions[sessionID] = session

	// Test joining the session
	err := nem.JoinGroupSession(sessionID, "peer1")

	if err != nil {
		t.Fatalf("Expected no error joining session, got: %v", err)
	}

	// Verify participant was added
	if len(session.Participants) != 2 {
		t.Errorf("Expected 2 participants after join, got %d", len(session.Participants))
	}

	found := false
	for _, pid := range session.Participants {
		if pid == "peer1" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'peer1' to be in participants list")
	}
}

func TestJoinGroupSession_SessionNotFound(t *testing.T) {
	nem, _, _ := createTestNetworkEventManager()

	err := nem.JoinGroupSession("nonexistent_session", "peer1")

	if err == nil {
		t.Fatal("Expected error for nonexistent session")
	}

	expectedError := "session not found"
	if err.Error()[:len(expectedError)] != expectedError {
		t.Errorf("Expected error to start with '%s', got: %v", expectedError, err)
	}
}

func TestJoinGroupSession_AtCapacity(t *testing.T) {
	nem, _, _ := createTestNetworkEventManager()

	// Create a session at capacity
	sessionID := "full_session"
	session := &GroupSession{
		ID:               sessionID,
		EventName:        "group_activity",
		Participants:     []string{"peer1", "peer2"},
		MaxParticipants:  2,
		StartTime:        time.Now(),
		LastActivity:     time.Now(),
		VoteChoices:      make(map[string]int),
		ParticipantVotes: make(map[string]int),
	}
	nem.groupSessions[sessionID] = session

	err := nem.JoinGroupSession(sessionID, "peer3")

	if err == nil {
		t.Fatal("Expected error for session at capacity")
	}

	expectedError := "session at maximum capacity"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got: %v", expectedError, err)
	}
}

func TestSubmitGroupChoice(t *testing.T) {
	nem, _, _ := createTestNetworkEventManager()

	// Create a test session with participants
	sessionID := "voting_session"
	session := &GroupSession{
		ID:               sessionID,
		EventName:        "group_activity",
		Participants:     []string{"local_peer_123", "peer1", "peer2"},
		MaxParticipants:  4,
		CurrentState:     "active",
		StartTime:        time.Now(),
		LastActivity:     time.Now(),
		VoteChoices:      make(map[string]int),
		ParticipantVotes: make(map[string]int),
	}
	nem.groupSessions[sessionID] = session

	// Submit a vote
	err := nem.SubmitGroupChoice(sessionID, "peer1", 0)

	if err != nil {
		t.Fatalf("Expected no error submitting choice, got: %v", err)
	}

	// Verify vote was recorded
	if session.ParticipantVotes["peer1"] != 0 {
		t.Errorf("Expected vote from peer1 to be 0, got %d", session.ParticipantVotes["peer1"])
	}

	if session.VoteChoices["0"] != 1 {
		t.Errorf("Expected 1 vote for choice 0, got %d", session.VoteChoices["0"])
	}
}

func TestSubmitGroupChoice_AllParticipantsVoted(t *testing.T) {
	nem, _, _ := createTestNetworkEventManager()

	// Create a session with 2 participants
	sessionID := "voting_session"
	session := &GroupSession{
		ID:               sessionID,
		EventName:        "group_activity",
		Participants:     []string{"peer1", "peer2"},
		MaxParticipants:  4,
		CurrentState:     "active",
		StartTime:        time.Now(),
		LastActivity:     time.Now(),
		VoteChoices:      make(map[string]int),
		ParticipantVotes: make(map[string]int),
	}
	nem.groupSessions[sessionID] = session

	// First vote
	err := nem.SubmitGroupChoice(sessionID, "peer1", 0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Second vote should complete the session
	err = nem.SubmitGroupChoice(sessionID, "peer2", 1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Session should be completed
	if session.CurrentState != "completed" {
		t.Errorf("Expected session state 'completed', got '%s'", session.CurrentState)
	}
}

func TestSubmitGroupChoice_NotParticipant(t *testing.T) {
	nem, _, _ := createTestNetworkEventManager()

	// Create a session
	sessionID := "voting_session"
	session := &GroupSession{
		ID:               sessionID,
		EventName:        "group_activity",
		Participants:     []string{"peer1", "peer2"},
		MaxParticipants:  4,
		CurrentState:     "active",
		StartTime:        time.Now(),
		LastActivity:     time.Now(),
		VoteChoices:      make(map[string]int),
		ParticipantVotes: make(map[string]int),
	}
	nem.groupSessions[sessionID] = session

	// Try to vote as non-participant
	err := nem.SubmitGroupChoice(sessionID, "peer3", 0)

	if err == nil {
		t.Fatal("Expected error for non-participant")
	}

	expectedError := "not a participant in session"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got: %v", expectedError, err)
	}
}

func TestAddPeerEventListener(t *testing.T) {
	nem, _, _ := createTestNetworkEventManager()

	callbackTriggered := false
	var capturedEventType PeerEventType
	var capturedPeerID string

	callback := func(eventType PeerEventType, peerID string, peerInfo *PeerInfo) {
		callbackTriggered = true
		capturedEventType = eventType
		capturedPeerID = peerID
	}

	nem.AddPeerEventListener(PeerEventJoined, callback)

	// Verify callback was registered
	if len(nem.peerEventCallbacks[PeerEventJoined]) != 1 {
		t.Errorf("Expected 1 callback registered, got %d", len(nem.peerEventCallbacks[PeerEventJoined]))
	}

	// Trigger the callback manually to test
	for _, cb := range nem.peerEventCallbacks[PeerEventJoined] {
		cb(PeerEventJoined, "test_peer", &PeerInfo{ID: "test_peer"})
	}

	if !callbackTriggered {
		t.Error("Expected callback to be triggered")
	}

	if capturedEventType != PeerEventJoined {
		t.Errorf("Expected event type PeerEventJoined, got %v", capturedEventType)
	}

	if capturedPeerID != "test_peer" {
		t.Errorf("Expected peer ID 'test_peer', got '%s'", capturedPeerID)
	}
}

func TestIsNetworkEvent(t *testing.T) {
	nem, _, _ := createTestNetworkEventManager()

	tests := []struct {
		name     string
		event    *GeneralDialogEvent
		expected bool
	}{
		{
			name: "Event with multiplayer keyword",
			event: &GeneralDialogEvent{
				Keywords: []string{"multiplayer", "fun"},
			},
			expected: true,
		},
		{
			name: "Event with group keyword",
			event: &GeneralDialogEvent{
				Keywords: []string{"social", "group"},
			},
			expected: true,
		},
		{
			name: "Event with group category",
			event: &GeneralDialogEvent{
				Category: "group",
			},
			expected: true,
		},
		{
			name: "Event with multiplayer category",
			event: &GeneralDialogEvent{
				Category: "multiplayer",
			},
			expected: true,
		},
		{
			name: "Regular conversation event",
			event: &GeneralDialogEvent{
				Category: "conversation",
				Keywords: []string{"chat", "talk"},
			},
			expected: false,
		},
		{
			name:     "Nil event",
			event:    nil,
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := nem.isNetworkEvent(test.event)
			if result != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, result)
			}
		})
	}
}

func TestGetMaxParticipants(t *testing.T) {
	nem, _, _ := createTestNetworkEventManager()

	tests := []struct {
		name     string
		event    *GeneralDialogEvent
		expected int
	}{
		{
			name: "Event with maxParticipants in effects",
			event: &GeneralDialogEvent{
				RandomEventConfig: RandomEventConfig{
					Effects: map[string]float64{"maxParticipants": 6.0},
				},
			},
			expected: 6,
		},
		{
			name: "Event without maxParticipants",
			event: &GeneralDialogEvent{
				RandomEventConfig: RandomEventConfig{
					Effects: map[string]float64{"happiness": 5.0},
				},
			},
			expected: 4, // default
		},
		{
			name: "Event with zero maxParticipants",
			event: &GeneralDialogEvent{
				RandomEventConfig: RandomEventConfig{
					Effects: map[string]float64{"maxParticipants": 0.0},
				},
			},
			expected: 4, // default when invalid
		},
		{
			name: "Event with nil effects",
			event: &GeneralDialogEvent{
				RandomEventConfig: RandomEventConfig{
					Effects: nil,
				},
			},
			expected: 4, // default
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := nem.getMaxParticipants(test.event)
			if result != test.expected {
				t.Errorf("Expected %d, got %d", test.expected, result)
			}
		})
	}
}

func TestHandleNetworkEventMessage(t *testing.T) {
	nem, _, mockPeerManager := createTestNetworkEventManager()

	// Add a valid peer
	mockPeerManager.AddPeer("test_peer", &PeerInfo{
		ID:          "test_peer",
		CharacterID: "test_char",
		LastSeen:    time.Now(),
	})

	tests := []struct {
		name        string
		payload     NetworkEventPayload
		expectError bool
	}{
		{
			name: "Valid event invitation",
			payload: NetworkEventPayload{
				Type:        "event_invite",
				EventName:   "test_conversation",
				InitiatorID: "test_peer",
				SessionID:   "session_123",
				Timestamp:   time.Now(),
			},
			expectError: false,
		},
		{
			name: "Unknown event type",
			payload: NetworkEventPayload{
				Type:      "unknown_type",
				Timestamp: time.Now(),
			},
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payloadBytes, err := json.Marshal(test.payload)
			if err != nil {
				t.Fatalf("Failed to marshal payload: %v", err)
			}

			err = nem.handleNetworkEventMessage(payloadBytes, "test_peer")

			if test.expectError && err == nil {
				t.Error("Expected error but got none")
			} else if !test.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestHandleGroupSessionMessage(t *testing.T) {
	nem, _, _ := createTestNetworkEventManager()

	// Create an existing session
	sessionID := "test_session"
	session := &GroupSession{
		ID:               sessionID,
		EventName:        "group_activity",
		Participants:     []string{"local_peer_123"},
		MaxParticipants:  4,
		CurrentState:     "active",
		StartTime:        time.Now(),
		LastActivity:     time.Now(),
		VoteChoices:      make(map[string]int),
		ParticipantVotes: make(map[string]int),
	}
	nem.groupSessions[sessionID] = session

	tests := []struct {
		name        string
		payload     GroupSessionPayload
		expectError bool
	}{
		{
			name: "Valid join action",
			payload: GroupSessionPayload{
				SessionID:     sessionID,
				Action:        "join",
				ParticipantID: "peer1",
				Timestamp:     time.Now(),
			},
			expectError: false,
		},
		{
			name: "Nonexistent session",
			payload: GroupSessionPayload{
				SessionID:     "nonexistent",
				Action:        "join",
				ParticipantID: "peer1",
				Timestamp:     time.Now(),
			},
			expectError: true,
		},
		{
			name: "Unknown action",
			payload: GroupSessionPayload{
				SessionID:     sessionID,
				Action:        "unknown_action",
				ParticipantID: "peer1",
				Timestamp:     time.Now(),
			},
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payloadBytes, err := json.Marshal(test.payload)
			if err != nil {
				t.Fatalf("Failed to marshal payload: %v", err)
			}

			err = nem.handleGroupSessionMessage(payloadBytes, "peer1")

			if test.expectError && err == nil {
				t.Error("Expected error but got none")
			} else if !test.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// Benchmark tests for performance validation

func BenchmarkTriggerNetworkEvent(b *testing.B) {
	nem, _, _ := createTestNetworkEventManager()
	gameState := &GameState{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := nem.TriggerNetworkEvent("test_conversation", gameState, nil)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkJoinGroupSession(b *testing.B) {
	nem, _, _ := createTestNetworkEventManager()

	// Create test sessions
	sessions := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		sessionID := fmt.Sprintf("session_%d", i)
		sessions[i] = sessionID
		session := &GroupSession{
			ID:               sessionID,
			EventName:        "group_activity",
			Participants:     []string{"local_peer_123"},
			MaxParticipants:  4,
			StartTime:        time.Now(),
			LastActivity:     time.Now(),
			VoteChoices:      make(map[string]int),
			ParticipantVotes: make(map[string]int),
		}
		nem.groupSessions[sessionID] = session
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := nem.JoinGroupSession(sessions[i], "peer1")
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkSubmitGroupChoice(b *testing.B) {
	nem, _, _ := createTestNetworkEventManager()

	// Create test sessions
	sessions := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		sessionID := fmt.Sprintf("session_%d", i)
		sessions[i] = sessionID
		session := &GroupSession{
			ID:               sessionID,
			EventName:        "group_activity",
			Participants:     []string{"local_peer_123", "peer1"},
			MaxParticipants:  4,
			CurrentState:     "active",
			StartTime:        time.Now(),
			LastActivity:     time.Now(),
			VoteChoices:      make(map[string]int),
			ParticipantVotes: make(map[string]int),
		}
		nem.groupSessions[sessionID] = session
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := nem.SubmitGroupChoice(sessions[i], "peer1", i%2)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}
