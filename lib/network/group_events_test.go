package network

import (
	"encoding/json"
	"testing"
	"time"
)

// MockNetworkManager implements GroupNetworkManagerInterface for testing
type MockNetworkManager struct {
	connectedPeers  []string
	localPeerID     string
	sentMessages    []MockMessage
	messageHandlers map[string]func([]byte, string) error
}

type MockMessage struct {
	MessageType string
	Payload     []byte
	TargetPeer  string
	Timestamp   time.Time
}

func NewMockNetworkManager(localPeerID string, connectedPeers []string) *MockNetworkManager {
	return &MockNetworkManager{
		connectedPeers:  connectedPeers,
		localPeerID:     localPeerID,
		sentMessages:    []MockMessage{},
		messageHandlers: make(map[string]func([]byte, string) error),
	}
}

func (m *MockNetworkManager) BroadcastMessage(msgType string, payload []byte) error {
	m.sentMessages = append(m.sentMessages, MockMessage{
		MessageType: msgType,
		Payload:     payload,
		TargetPeer:  "broadcast",
		Timestamp:   time.Now(),
	})
	return nil
}

func (m *MockNetworkManager) SendMessage(msgType string, payload []byte, targetPeerID string) error {
	m.sentMessages = append(m.sentMessages, MockMessage{
		MessageType: msgType,
		Payload:     payload,
		TargetPeer:  targetPeerID,
		Timestamp:   time.Now(),
	})
	return nil
}

func (m *MockNetworkManager) RegisterMessageHandler(msgType string, handler func([]byte, string) error) {
	m.messageHandlers[msgType] = handler
}

func (m *MockNetworkManager) GetConnectedPeers() []string {
	return m.connectedPeers
}

func (m *MockNetworkManager) GetLocalPeerID() string {
	return m.localPeerID
}

func (m *MockNetworkManager) SimulateIncomingMessage(msgType string, payload []byte, senderID string) error {
	if handler, exists := m.messageHandlers[msgType]; exists {
		return handler(payload, senderID)
	}
	return nil
}

func (m *MockNetworkManager) GetSentMessages() []MockMessage {
	return m.sentMessages
}

func (m *MockNetworkManager) ClearSentMessages() {
	m.sentMessages = []MockMessage{}
}

// Test helper to create sample event templates
func createTestEventTemplates() []GroupEventTemplate {
	return []GroupEventTemplate{
		{
			ID:              "trivia_game",
			Name:            "Trivia Challenge",
			Description:     "A collaborative trivia game for 2-4 players",
			Category:        "minigame",
			MinParticipants: 2,
			MaxParticipants: 4,
			EstimatedTime:   5 * time.Minute,
			Phases: []EventPhase{
				{
					Name:        "question1",
					Description: "First trivia question",
					Type:        "choice",
					Duration:    30 * time.Second,
					MinVotes:    2, // Require 2 votes to advance
					AutoAdvance: true,
					Choices: []EventChoice{
						{ID: "a", Text: "Answer A", Points: 10},
						{ID: "b", Text: "Answer B", Points: 0},
						{ID: "c", Text: "Answer C", Points: 0},
					},
				},
				{
					Name:        "question2",
					Description: "Second trivia question",
					Type:        "choice",
					Duration:    30 * time.Second,
					MinVotes:    1,
					AutoAdvance: true,
					Choices: []EventChoice{
						{ID: "x", Text: "Answer X", Points: 15},
						{ID: "y", Text: "Answer Y", Points: 0},
					},
				},
			},
		},
		{
			ID:              "story_collaboration",
			Name:            "Collaborative Story",
			Description:     "Build a story together",
			Category:        "scenario",
			MinParticipants: 3,
			MaxParticipants: 6,
			EstimatedTime:   10 * time.Minute,
			Phases: []EventPhase{
				{
					Name:        "opening",
					Description: "Choose story opening",
					Type:        "vote",
					Duration:    60 * time.Second,
					MinVotes:    2,
					AutoAdvance: false,
					Choices: []EventChoice{
						{ID: "fantasy", Text: "Fantasy Adventure", Points: 5},
						{ID: "scifi", Text: "Sci-Fi Mystery", Points: 5},
						{ID: "romance", Text: "Romance Story", Points: 5},
					},
				},
			},
		},
	}
}

func TestNewGroupEventManager(t *testing.T) {
	mockNetwork := NewMockNetworkManager("peer1", []string{"peer2", "peer3"})
	templates := createTestEventTemplates()

	gem := NewGroupEventManager(mockNetwork, templates)

	if gem == nil {
		t.Fatal("Expected non-nil GroupEventManager")
	}

	if len(gem.eventTemplates) != 2 {
		t.Errorf("Expected 2 event templates, got %d", len(gem.eventTemplates))
	}

	if gem.minParticipants != 2 {
		t.Errorf("Expected minParticipants to be 2, got %d", gem.minParticipants)
	}

	if gem.maxParticipants != 8 {
		t.Errorf("Expected maxParticipants to be 8, got %d", gem.maxParticipants)
	}

	// Verify network handler was registered
	if len(mockNetwork.messageHandlers) == 0 {
		t.Error("Expected message handler to be registered")
	}
}

func TestStartGroupEvent_Success(t *testing.T) {
	mockNetwork := NewMockNetworkManager("peer1", []string{"peer2", "peer3"})
	templates := createTestEventTemplates()
	gem := NewGroupEventManager(mockNetwork, templates)

	sessionID, err := gem.StartGroupEvent("trivia_game", "peer1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if sessionID == "" {
		t.Error("Expected non-empty session ID")
	}

	// Verify event was created
	activeEvents := gem.GetActiveEvents()
	if len(activeEvents) != 1 {
		t.Errorf("Expected 1 active event, got %d", len(activeEvents))
	}

	// Verify invitation was broadcast
	sentMessages := mockNetwork.GetSentMessages()
	if len(sentMessages) != 1 {
		t.Errorf("Expected 1 broadcast message, got %d", len(sentMessages))
	}

	if sentMessages[0].MessageType != "group_event" {
		t.Errorf("Expected message type 'group_event', got %s", sentMessages[0].MessageType)
	}
}

func TestStartGroupEvent_TemplateNotFound(t *testing.T) {
	mockNetwork := NewMockNetworkManager("peer1", []string{"peer2", "peer3"})
	templates := createTestEventTemplates()
	gem := NewGroupEventManager(mockNetwork, templates)

	_, err := gem.StartGroupEvent("nonexistent_template", "peer1")

	if err == nil {
		t.Error("Expected error for nonexistent template")
	}

	expectedError := "group event template not found: nonexistent_template"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}

func TestStartGroupEvent_InsufficientParticipants(t *testing.T) {
	// Create network with only 1 peer (total 2 including local)
	mockNetwork := NewMockNetworkManager("peer1", []string{"peer2"})
	templates := createTestEventTemplates()
	gem := NewGroupEventManager(mockNetwork, templates)

	// Try to start story collaboration (requires 3 minimum)
	_, err := gem.StartGroupEvent("story_collaboration", "peer1")

	if err == nil {
		t.Error("Expected error for insufficient participants")
	}

	if !contains(err.Error(), "insufficient participants") {
		t.Errorf("Expected insufficient participants error, got %v", err)
	}
}

func TestJoinGroupEvent_Success(t *testing.T) {
	mockNetwork := NewMockNetworkManager("peer1", []string{"peer2", "peer3"})
	templates := createTestEventTemplates()
	gem := NewGroupEventManager(mockNetwork, templates)

	// Start an event
	sessionID, err := gem.StartGroupEvent("trivia_game", "peer1")
	if err != nil {
		t.Fatalf("Failed to start event: %v", err)
	}

	// Clear previous messages
	mockNetwork.ClearSentMessages()

	// Join the event
	err = gem.JoinGroupEvent(sessionID, "peer2")
	if err != nil {
		t.Fatalf("Expected no error joining event, got %v", err)
	}

	// Verify join message was broadcast
	sentMessages := mockNetwork.GetSentMessages()
	if len(sentMessages) != 1 {
		t.Errorf("Expected 1 broadcast message, got %d", len(sentMessages))
	}

	// Verify participant was added
	activeEvents := gem.GetActiveEvents()
	if len(activeEvents[0].Participants) != 2 {
		t.Errorf("Expected 2 participants, got %d", len(activeEvents[0].Participants))
	}
}

func TestJoinGroupEvent_EventNotFound(t *testing.T) {
	mockNetwork := NewMockNetworkManager("peer1", []string{"peer2", "peer3"})
	templates := createTestEventTemplates()
	gem := NewGroupEventManager(mockNetwork, templates)

	err := gem.JoinGroupEvent("nonexistent_session", "peer2")

	if err == nil {
		t.Error("Expected error for nonexistent event")
	}

	expectedError := "group event not found: nonexistent_session"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}

func TestJoinGroupEvent_AlreadyParticipant(t *testing.T) {
	mockNetwork := NewMockNetworkManager("peer1", []string{"peer2", "peer3"})
	templates := createTestEventTemplates()
	gem := NewGroupEventManager(mockNetwork, templates)

	sessionID, err := gem.StartGroupEvent("trivia_game", "peer1")
	if err != nil {
		t.Fatalf("Failed to start event: %v", err)
	}

	// Try to join as the initiator
	err = gem.JoinGroupEvent(sessionID, "peer1")

	if err == nil {
		t.Error("Expected error for already being participant")
	}

	if !contains(err.Error(), "already participant") {
		t.Errorf("Expected already participant error, got %v", err)
	}
}

func TestSubmitVote_Success(t *testing.T) {
	mockNetwork := NewMockNetworkManager("peer1", []string{"peer2", "peer3"})
	templates := createTestEventTemplates()
	gem := NewGroupEventManager(mockNetwork, templates)

	// Start story collaboration event (doesn't auto-advance)
	sessionID, err := gem.StartGroupEvent("story_collaboration", "peer1")
	if err != nil {
		t.Fatalf("Failed to start event: %v", err)
	}

	err = gem.JoinGroupEvent(sessionID, "peer2")
	if err != nil {
		t.Fatalf("Failed to join event: %v", err)
	}

	err = gem.JoinGroupEvent(sessionID, "peer3")
	if err != nil {
		t.Fatalf("Failed to join event: %v", err)
	}

	// Clear previous messages
	mockNetwork.ClearSentMessages()

	// Submit vote
	err = gem.SubmitVote(sessionID, "peer1", "fantasy")
	if err != nil {
		t.Fatalf("Expected no error submitting vote, got %v", err)
	}

	// Verify vote message was broadcast (might include advance message)
	sentMessages := mockNetwork.GetSentMessages()
	if len(sentMessages) == 0 {
		t.Error("Expected at least 1 broadcast message")
	}

	// Find the vote message
	var voteMessage GroupEventMessage
	found := false
	for _, msg := range sentMessages {
		if json.Unmarshal(msg.Payload, &voteMessage) == nil && voteMessage.Type == "vote" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find vote message in broadcast")
	}

	// Verify vote was recorded
	activeEvents := gem.GetActiveEvents()
	event := activeEvents[0]
	if event.Votes["fantasy"] != 1 {
		t.Errorf("Expected vote count 1 for choice 'fantasy', got %d", event.Votes["fantasy"])
	}

	if event.ParticipantVotes["peer1"] != "fantasy" {
		t.Errorf("Expected participant vote 'fantasy', got %s", event.ParticipantVotes["peer1"])
	}
}

func TestSubmitVote_InvalidChoice(t *testing.T) {
	mockNetwork := NewMockNetworkManager("peer1", []string{"peer2", "peer3"})
	templates := createTestEventTemplates()
	gem := NewGroupEventManager(mockNetwork, templates)

	sessionID, err := gem.StartGroupEvent("trivia_game", "peer1")
	if err != nil {
		t.Fatalf("Failed to start event: %v", err)
	}

	err = gem.SubmitVote(sessionID, "peer1", "invalid_choice")

	if err == nil {
		t.Error("Expected error for invalid choice")
	}

	expectedError := "invalid choice: invalid_choice"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}

func TestSubmitVote_AutoAdvance(t *testing.T) {
	mockNetwork := NewMockNetworkManager("peer1", []string{"peer2"})
	templates := createTestEventTemplates()
	gem := NewGroupEventManager(mockNetwork, templates)

	// Start event with 2 participants
	sessionID, err := gem.StartGroupEvent("trivia_game", "peer1")
	if err != nil {
		t.Fatalf("Failed to start event: %v", err)
	}

	err = gem.JoinGroupEvent(sessionID, "peer2")
	if err != nil {
		t.Fatalf("Failed to join event: %v", err)
	}

	// First participant votes
	err = gem.SubmitVote(sessionID, "peer1", "a")
	if err != nil {
		t.Fatalf("Failed to submit first vote: %v", err)
	}

	// Clear messages
	mockNetwork.ClearSentMessages()

	// Second participant votes (should trigger advance since MinVotes=2, AutoAdvance=true)
	err = gem.SubmitVote(sessionID, "peer2", "b")
	if err != nil {
		t.Fatalf("Failed to submit second vote: %v", err)
	}

	// Verify advance message was sent
	sentMessages := mockNetwork.GetSentMessages()
	found := false
	for _, msg := range sentMessages {
		var groupMsg GroupEventMessage
		if json.Unmarshal(msg.Payload, &groupMsg) == nil && groupMsg.Type == "advance" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected advance message to be sent")
	}

	// Verify phase advancement
	activeEvents := gem.GetActiveEvents()
	if len(activeEvents) > 0 {
		event := activeEvents[0]
		if event.CurrentPhase != "question2" {
			t.Errorf("Expected current phase 'question2', got %s", event.CurrentPhase)
		}
	}
}

func TestGetEventTemplates(t *testing.T) {
	mockNetwork := NewMockNetworkManager("peer1", []string{"peer2", "peer3"})
	templates := createTestEventTemplates()
	gem := NewGroupEventManager(mockNetwork, templates)

	retrievedTemplates := gem.GetEventTemplates()

	if len(retrievedTemplates) != 2 {
		t.Errorf("Expected 2 templates, got %d", len(retrievedTemplates))
	}

	if retrievedTemplates[0].ID != "trivia_game" {
		t.Errorf("Expected first template ID 'trivia_game', got %s", retrievedTemplates[0].ID)
	}
}

func TestGetActiveEvents(t *testing.T) {
	mockNetwork := NewMockNetworkManager("peer1", []string{"peer2", "peer3"})
	templates := createTestEventTemplates()
	gem := NewGroupEventManager(mockNetwork, templates)

	// Initially no active events
	activeEvents := gem.GetActiveEvents()
	if len(activeEvents) != 0 {
		t.Errorf("Expected 0 active events initially, got %d", len(activeEvents))
	}

	// Start an event
	sessionID, err := gem.StartGroupEvent("trivia_game", "peer1")
	if err != nil {
		t.Fatalf("Failed to start event: %v", err)
	}

	// Now should have 1 active event
	activeEvents = gem.GetActiveEvents()
	if len(activeEvents) != 1 {
		t.Errorf("Expected 1 active event, got %d", len(activeEvents))
	}

	if activeEvents[0].SessionID != sessionID {
		t.Errorf("Expected session ID %s, got %s", sessionID, activeEvents[0].SessionID)
	}
}

func TestGetParticipantHistory(t *testing.T) {
	mockNetwork := NewMockNetworkManager("peer1", []string{})
	templates := createTestEventTemplates()
	gem := NewGroupEventManager(mockNetwork, templates)

	// Initially no history
	history := gem.GetParticipantHistory("peer1")
	if len(history) != 0 {
		t.Errorf("Expected 0 history entries initially, got %d", len(history))
	}

	// History is updated when events complete
	// For now, just verify empty history returns empty slice
	if history == nil {
		t.Error("Expected non-nil history slice")
	}
}

func TestHandleGroupEventMessage_InvalidJSON(t *testing.T) {
	mockNetwork := NewMockNetworkManager("peer1", []string{"peer2"})
	templates := createTestEventTemplates()
	gem := NewGroupEventManager(mockNetwork, templates)

	// Test invalid JSON
	err := gem.handleGroupEventMessage([]byte("invalid json"), "peer2")

	if err == nil {
		t.Error("Expected error for invalid JSON")
	}

	if !contains(err.Error(), "failed to unmarshal") {
		t.Errorf("Expected unmarshal error, got %v", err)
	}
}

func TestHandleGroupEventMessage_UnknownType(t *testing.T) {
	mockNetwork := NewMockNetworkManager("peer1", []string{"peer2"})
	templates := createTestEventTemplates()
	gem := NewGroupEventManager(mockNetwork, templates)

	message := GroupEventMessage{
		Type:      "unknown_type",
		SessionID: "test_session",
		Sender:    "peer2",
		Timestamp: time.Now(),
	}

	data, _ := json.Marshal(message)
	err := gem.handleGroupEventMessage(data, "peer2")

	if err == nil {
		t.Error("Expected error for unknown message type")
	}

	expectedError := "unknown group event message type: unknown_type"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}

func TestConcurrentAccess(t *testing.T) {
	mockNetwork := NewMockNetworkManager("peer1", []string{"peer2", "peer3", "peer4"})
	templates := createTestEventTemplates()
	gem := NewGroupEventManager(mockNetwork, templates)

	// Start an event
	sessionID, err := gem.StartGroupEvent("trivia_game", "peer1")
	if err != nil {
		t.Fatalf("Failed to start event: %v", err)
	}

	// Simulate concurrent operations
	done := make(chan bool, 3)

	// Concurrent joins
	go func() {
		_ = gem.JoinGroupEvent(sessionID, "peer2")
		done <- true
	}()

	go func() {
		_ = gem.JoinGroupEvent(sessionID, "peer3")
		done <- true
	}()

	// Concurrent active events access
	go func() {
		_ = gem.GetActiveEvents()
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}

	// Verify final state
	activeEvents := gem.GetActiveEvents()
	if len(activeEvents) != 1 {
		t.Errorf("Expected 1 active event after concurrent access, got %d", len(activeEvents))
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || (len(s) > len(substr) &&
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
				containsHelper(s, substr))))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Benchmark tests
func BenchmarkStartGroupEvent(b *testing.B) {
	mockNetwork := NewMockNetworkManager("peer1", []string{"peer2", "peer3", "peer4"})
	templates := createTestEventTemplates()
	gem := NewGroupEventManager(mockNetwork, templates)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sessionID, err := gem.StartGroupEvent("trivia_game", "peer1")
		if err != nil {
			b.Fatalf("Failed to start event: %v", err)
		}
		// Clean up for next iteration
		delete(gem.activeEvents, sessionID)
		delete(gem.participants, sessionID)
		mockNetwork.ClearSentMessages()
	}
}

func BenchmarkSubmitVote(b *testing.B) {
	mockNetwork := NewMockNetworkManager("peer1", []string{"peer2", "peer3", "peer4"})
	templates := createTestEventTemplates()
	gem := NewGroupEventManager(mockNetwork, templates)

	// Setup
	sessionID, err := gem.StartGroupEvent("trivia_game", "peer1")
	if err != nil {
		b.Fatalf("Failed to start event: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		choice := "a"
		if i%2 == 1 {
			choice = "b"
		}
		_ = gem.SubmitVote(sessionID, "peer1", choice)
	}
}

func TestEventCompletion(t *testing.T) {
	mockNetwork := NewMockNetworkManager("peer1", []string{"peer2"})
	templates := createTestEventTemplates()
	gem := NewGroupEventManager(mockNetwork, templates)

	// Start event with simple 1-phase template
	simpleTemplate := GroupEventTemplate{
		ID:              "simple_test",
		Name:            "Simple Test",
		Description:     "Simple one-phase test",
		Category:        "test",
		MinParticipants: 2,
		MaxParticipants: 2,
		EstimatedTime:   1 * time.Minute,
		Phases: []EventPhase{
			{
				Name:        "only_phase",
				Description: "Only phase",
				Type:        "choice",
				Duration:    30 * time.Second,
				MinVotes:    2,
				AutoAdvance: true,
				Choices: []EventChoice{
					{ID: "yes", Text: "Yes", Points: 10},
					{ID: "no", Text: "No", Points: 5},
				},
			},
		},
	}

	gem.eventTemplates = append(gem.eventTemplates, simpleTemplate)

	sessionID, err := gem.StartGroupEvent("simple_test", "peer1")
	if err != nil {
		t.Fatalf("Failed to start event: %v", err)
	}

	err = gem.JoinGroupEvent(sessionID, "peer2")
	if err != nil {
		t.Fatalf("Failed to join event: %v", err)
	}

	// Both participants vote to complete the event
	err = gem.SubmitVote(sessionID, "peer1", "yes")
	if err != nil {
		t.Fatalf("Failed to submit first vote: %v", err)
	}

	err = gem.SubmitVote(sessionID, "peer2", "no")
	if err != nil {
		t.Fatalf("Failed to submit second vote: %v", err)
	}

	// Event should be completed and removed from active events
	activeEvents := gem.GetActiveEvents()
	if len(activeEvents) != 0 {
		t.Errorf("Expected 0 active events after completion, got %d", len(activeEvents))
	}

	// Check participant history
	history := gem.GetParticipantHistory("peer1")
	if len(history) != 1 {
		t.Errorf("Expected 1 history entry for peer1, got %d", len(history))
	}

	if len(history) > 0 && history[0].TemplateID != "simple_test" {
		t.Errorf("Expected template ID 'simple_test', got %s", history[0].TemplateID)
	}
}

func TestHandleIncomingMessages(t *testing.T) {
	mockNetwork := NewMockNetworkManager("peer1", []string{"peer2"})
	templates := createTestEventTemplates()
	_ = NewGroupEventManager(mockNetwork, templates) // Create manager to register handlers

	// Test invitation handling
	inviteMsg := GroupEventMessage{
		Type:      "invite",
		SessionID: "test_session",
		Sender:    "peer2",
		Data: map[string]interface{}{
			"templateName": "Test Event",
		},
		Timestamp: time.Now(),
	}

	inviteData, _ := json.Marshal(inviteMsg)
	err := mockNetwork.SimulateIncomingMessage("group_event", inviteData, "peer2")
	if err != nil {
		t.Errorf("Failed to handle invitation: %v", err)
	}

	// Test join message handling
	joinMsg := GroupEventMessage{
		Type:      "join",
		SessionID: "test_session",
		Sender:    "peer2",
		Data: map[string]interface{}{
			"participantCount": float64(2),
		},
		Timestamp: time.Now(),
	}

	joinData, _ := json.Marshal(joinMsg)
	err = mockNetwork.SimulateIncomingMessage("group_event", joinData, "peer2")
	if err != nil {
		t.Errorf("Failed to handle join message: %v", err)
	}

	// Test vote message handling
	voteMsg := GroupEventMessage{
		Type:      "vote",
		SessionID: "test_session",
		Sender:    "peer2",
		Data: map[string]interface{}{
			"choiceId": "test_choice",
		},
		Timestamp: time.Now(),
	}

	voteData, _ := json.Marshal(voteMsg)
	err = mockNetwork.SimulateIncomingMessage("group_event", voteData, "peer2")
	if err != nil {
		t.Errorf("Failed to handle vote message: %v", err)
	}

	// Test completion message handling
	endMsg := GroupEventMessage{
		Type:      "end",
		SessionID: "test_session",
		Sender:    "peer2",
		Data: map[string]interface{}{
			"duration": "2m30s",
		},
		Timestamp: time.Now(),
	}

	endData, _ := json.Marshal(endMsg)
	err = mockNetwork.SimulateIncomingMessage("group_event", endData, "peer2")
	if err != nil {
		t.Errorf("Failed to handle completion message: %v", err)
	}
}
