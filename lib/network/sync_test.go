package network

import (
	"crypto/ed25519"
	"encoding/json"
	"sync"
	"testing"
	"time"
)

// Mock implementations for testing

type mockNetworkManager struct {
	mu                sync.RWMutex
	networkID         string
	handlers          map[MessageType]MessageHandler
	sentMessages      []Message
	shouldFailSend    bool
	shouldFailHandler bool
}

func newMockNetworkManager(networkID string) *mockNetworkManager {
	return &mockNetworkManager{
		networkID:    networkID,
		handlers:     make(map[MessageType]MessageHandler),
		sentMessages: make([]Message, 0),
	}
}

func (m *mockNetworkManager) RegisterMessageHandler(msgType MessageType, handler MessageHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[msgType] = handler
}

func (m *mockNetworkManager) SendMessage(msgType MessageType, payload []byte, targetPeerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldFailSend {
		return &MockError{msg: "mock send error"}
	}

	message := Message{
		Type:      msgType,
		From:      m.networkID,
		To:        targetPeerID,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	m.sentMessages = append(m.sentMessages, message)
	return nil
}

func (m *mockNetworkManager) GetNetworkID() string {
	return m.networkID
}

func (m *mockNetworkManager) GetSentMessages() []Message {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]Message, len(m.sentMessages))
	copy(result, m.sentMessages)
	return result
}

func (m *mockNetworkManager) ClearSentMessages() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sentMessages = m.sentMessages[:0]
}

func (m *mockNetworkManager) SimulateIncomingMessage(msgType MessageType, payload []byte, fromPeer string) error {
	m.mu.RLock()
	handler, exists := m.handlers[msgType]
	m.mu.RUnlock()

	if !exists {
		return &MockError{msg: "no handler registered"}
	}

	if m.shouldFailHandler {
		return &MockError{msg: "mock handler error"}
	}

	message := Message{
		Type:      msgType,
		From:      fromPeer,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	peer := &Peer{
		ID:       fromPeer,
		AddrStr:  "127.0.0.1:8080",
		LastSeen: time.Now(),
	}

	return handler(message, peer)
}

type mockProtocolManager struct {
	publicKey ed25519.PublicKey
}

func newMockProtocolManager() *mockProtocolManager {
	// Create a mock ed25519 public key (32 bytes)
	publicKey := make(ed25519.PublicKey, ed25519.PublicKeySize)
	copy(publicKey, []byte("mock-public-key-32-bytes-long!!!"))

	return &mockProtocolManager{
		publicKey: publicKey,
	}
}

func (m *mockProtocolManager) GetPublicKey() ed25519.PublicKey {
	return m.publicKey
}

// MockError implements error interface for controlled testing
type MockError struct {
	msg string
}

func (e *MockError) Error() string {
	return e.msg
}

// Test helper functions

func createTestStateSynchronizer(networkID string) (*StateSynchronizer, *mockNetworkManager) {
	mockNetwork := newMockNetworkManager(networkID)
	mockProtocol := newMockProtocolManager()

	ss := NewStateSynchronizer(mockNetwork, mockProtocol)
	return ss, mockNetwork
}

func createTestCharacterState(characterID string) *CharacterState {
	return &CharacterState{
		CharacterID:  characterID,
		Position:     Position{X: 100.0, Y: 200.0},
		Animation:    "idle",
		CurrentState: "active",
		GameStats: map[string]float64{
			"happiness": 0.8,
			"hunger":    0.6,
		},
		RomanceStats: map[string]float64{
			"affection": 0.7,
			"trust":     0.9,
		},
		LastUpdate:   time.Now(),
		UpdateSource: "test-peer",
		Version:      1,
		Checksum:     "mock-checksum",
	}
}

// Test StateSynchronizer creation and initialization

func TestNewStateSynchronizer(t *testing.T) {
	mockNetwork := newMockNetworkManager("test-peer")
	mockProtocol := newMockProtocolManager()

	ss := NewStateSynchronizer(mockNetwork, mockProtocol)

	if ss == nil {
		t.Fatal("StateSynchronizer should not be nil")
	}

	if ss.networkManager == nil {
		t.Error("NetworkManager should not be nil")
	}

	if ss.protocolManager == nil {
		t.Error("ProtocolManager should not be nil")
	}

	if ss.syncInterval != 30*time.Second {
		t.Errorf("Default sync interval = %v, want %v", ss.syncInterval, 30*time.Second)
	}

	if ss.characterStates == nil {
		t.Error("Character states map should be initialized")
	}

	if ss.conflictResolver == nil {
		t.Error("Conflict resolver should be initialized")
	}
}

func TestStateSynchronizer_Start(t *testing.T) {
	ss, mockNetwork := createTestStateSynchronizer("test-peer")

	err := ss.Start()
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	// Verify message handler was registered
	mockNetwork.mu.RLock()
	_, exists := mockNetwork.handlers[MessageTypeStateSync]
	mockNetwork.mu.RUnlock()

	if !exists {
		t.Error("State sync message handler not registered")
	}

	// Clean shutdown
	err = ss.Stop()
	if err != nil {
		t.Errorf("Stop() failed: %v", err)
	}
}

func TestStateSynchronizer_UpdateCharacterState(t *testing.T) {
	ss, _ := createTestStateSynchronizer("test-peer")

	characterID := "test-character"
	position := Position{X: 150.0, Y: 250.0}
	animation := "walking"
	currentState := "moving"
	gameStats := map[string]float64{"energy": 0.5}
	romanceStats := map[string]float64{"intimacy": 0.3}

	err := ss.UpdateCharacterState(characterID, position, animation, currentState, gameStats, romanceStats)
	if err != nil {
		t.Fatalf("UpdateCharacterState() failed: %v", err)
	}

	// Verify state was stored
	state, exists := ss.GetCharacterState(characterID)
	if !exists {
		t.Fatal("Character state not found after update")
	}

	if state.Position.X != position.X || state.Position.Y != position.Y {
		t.Errorf("Position = %v, want %v", state.Position, position)
	}

	if state.Animation != animation {
		t.Errorf("Animation = %s, want %s", state.Animation, animation)
	}

	if state.CurrentState != currentState {
		t.Errorf("CurrentState = %s, want %s", state.CurrentState, currentState)
	}

	if state.Version != 1 {
		t.Errorf("Version = %d, want 1", state.Version)
	}

	if state.UpdateSource != "test-peer" {
		t.Errorf("UpdateSource = %s, want test-peer", state.UpdateSource)
	}
}

func TestStateSynchronizer_UpdateCharacterState_VersionIncrement(t *testing.T) {
	ss, _ := createTestStateSynchronizer("test-peer")

	characterID := "test-character"
	position := Position{X: 100.0, Y: 200.0}

	// First update
	err := ss.UpdateCharacterState(characterID, position, "idle", "active", nil, nil)
	if err != nil {
		t.Fatalf("First UpdateCharacterState() failed: %v", err)
	}

	state1, _ := ss.GetCharacterState(characterID)
	if state1.Version != 1 {
		t.Errorf("First version = %d, want 1", state1.Version)
	}

	// Second update
	err = ss.UpdateCharacterState(characterID, position, "walking", "moving", nil, nil)
	if err != nil {
		t.Fatalf("Second UpdateCharacterState() failed: %v", err)
	}

	state2, _ := ss.GetCharacterState(characterID)
	if state2.Version != 2 {
		t.Errorf("Second version = %d, want 2", state2.Version)
	}
}

func TestStateSynchronizer_SetSyncInterval(t *testing.T) {
	ss, _ := createTestStateSynchronizer("test-peer")

	newInterval := 10 * time.Second
	ss.SetSyncInterval(newInterval)

	if ss.syncInterval != newInterval {
		t.Errorf("Sync interval = %v, want %v", ss.syncInterval, newInterval)
	}
}

func TestStateSynchronizer_PeriodicSync(t *testing.T) {
	ss, mockNetwork := createTestStateSynchronizer("test-peer")

	// Set short interval for testing
	ss.SetSyncInterval(50 * time.Millisecond)

	// Add a character state
	characterID := "test-character"
	err := ss.UpdateCharacterState(characterID, Position{X: 10, Y: 20}, "idle", "active", nil, nil)
	if err != nil {
		t.Fatalf("UpdateCharacterState() failed: %v", err)
	}

	// Start sync
	err = ss.Start()
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	// Wait for periodic sync to trigger
	time.Sleep(100 * time.Millisecond)

	// Check if message was sent
	messages := mockNetwork.GetSentMessages()
	if len(messages) == 0 {
		t.Error("No sync messages sent during periodic sync")
	}

	// Verify message type
	if len(messages) > 0 && messages[0].Type != MessageTypeStateSync {
		t.Errorf("Message type = %v, want %v", messages[0].Type, MessageTypeStateSync)
	}

	// Clean shutdown
	err = ss.Stop()
	if err != nil {
		t.Errorf("Stop() failed: %v", err)
	}
}

func TestStateSynchronizer_HandleIncomingStateSync(t *testing.T) {
	ss, mockNetwork := createTestStateSynchronizer("local-peer")

	// Start the synchronizer
	err := ss.Start()
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer ss.Stop()

	// Create incoming state sync payload
	payload := StateSyncPayload{
		CharacterID:  "remote-character",
		Position:     Position{X: 300.0, Y: 400.0},
		Animation:    "running",
		CurrentState: "excited",
		GameStats:    map[string]float64{"happiness": 0.9},
		RomanceStats: map[string]float64{"affection": 0.8},
		LastUpdate:   time.Now(),
		Checksum:     "",
	}

	// Calculate checksum
	ss.mu.Lock()
	payload.Checksum = ss.calculatePayloadChecksum(payload)
	ss.mu.Unlock()

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	// Simulate incoming message
	err = mockNetwork.SimulateIncomingMessage(MessageTypeStateSync, payloadBytes, "remote-peer")
	if err != nil {
		t.Fatalf("SimulateIncomingMessage() failed: %v", err)
	}

	// Verify state was updated
	state, exists := ss.GetCharacterState("remote-character")
	if !exists {
		t.Fatal("Remote character state not found after sync")
	}

	if state.Position.X != 300.0 || state.Position.Y != 400.0 {
		t.Errorf("Position = %v, want {300, 400}", state.Position)
	}

	if state.Animation != "running" {
		t.Errorf("Animation = %s, want running", state.Animation)
	}

	if state.UpdateSource != "remote-peer" {
		t.Errorf("UpdateSource = %s, want remote-peer", state.UpdateSource)
	}
}

func TestStateSynchronizer_HandleIncomingStateSync_ChecksumMismatch(t *testing.T) {
	ss, mockNetwork := createTestStateSynchronizer("local-peer")

	err := ss.Start()
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer ss.Stop()

	// Create payload with incorrect checksum
	payload := StateSyncPayload{
		CharacterID:  "test-character",
		Position:     Position{X: 100.0, Y: 200.0},
		Animation:    "idle",
		CurrentState: "active",
		Checksum:     "incorrect-checksum",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	// Should return error due to checksum mismatch
	err = mockNetwork.SimulateIncomingMessage(MessageTypeStateSync, payloadBytes, "remote-peer")
	if err == nil {
		t.Error("Expected checksum mismatch error, got nil")
	}

	// Verify state was not created
	_, exists := ss.GetCharacterState("test-character")
	if exists {
		t.Error("Character state should not exist after checksum failure")
	}
}

// Test ConflictResolver

func TestNewConflictResolver(t *testing.T) {
	cr := NewConflictResolver(TimestampWins)

	if cr.strategy != TimestampWins {
		t.Errorf("Strategy = %v, want %v", cr.strategy, TimestampWins)
	}

	if cr.peerPriorities == nil {
		t.Error("Peer priorities map should be initialized")
	}

	if cr.resolvedConflicts == nil {
		t.Error("Resolved conflicts map should be initialized")
	}
}

func TestConflictResolver_TimestampWins(t *testing.T) {
	cr := NewConflictResolver(TimestampWins)

	now := time.Now()
	localState := &CharacterState{
		CharacterID:  "test-character",
		LastUpdate:   now,
		UpdateSource: "local-peer",
		Version:      1,
	}

	incomingState := &CharacterState{
		CharacterID:  "test-character",
		LastUpdate:   now.Add(time.Second), // Newer timestamp
		UpdateSource: "remote-peer",
		Version:      2,
	}

	resolvedState, err := cr.ResolveConflict(localState, incomingState)
	if err != nil {
		t.Fatalf("ResolveConflict() failed: %v", err)
	}

	if resolvedState.UpdateSource != "remote-peer" {
		t.Errorf("Winning source = %s, want remote-peer", resolvedState.UpdateSource)
	}

	if cr.conflictCount != 1 {
		t.Errorf("Conflict count = %d, want 1", cr.conflictCount)
	}
}

func TestConflictResolver_TimestampWins_LocalNewer(t *testing.T) {
	cr := NewConflictResolver(TimestampWins)

	now := time.Now()
	localState := &CharacterState{
		CharacterID:  "test-character",
		LastUpdate:   now.Add(time.Second), // Newer timestamp
		UpdateSource: "local-peer",
		Version:      2,
	}

	incomingState := &CharacterState{
		CharacterID:  "test-character",
		LastUpdate:   now,
		UpdateSource: "remote-peer",
		Version:      1,
	}

	resolvedState, err := cr.ResolveConflict(localState, incomingState)
	if err != nil {
		t.Fatalf("ResolveConflict() failed: %v", err)
	}

	if resolvedState.UpdateSource != "local-peer" {
		t.Errorf("Winning source = %s, want local-peer", resolvedState.UpdateSource)
	}
}

func TestConflictResolver_PeerPriorityWins(t *testing.T) {
	cr := NewConflictResolver(PeerPriorityWins)
	cr.SetPeerPriority("local-peer", 5)
	cr.SetPeerPriority("remote-peer", 10) // Higher priority

	localState := &CharacterState{
		CharacterID:  "test-character",
		UpdateSource: "local-peer",
	}

	incomingState := &CharacterState{
		CharacterID:  "test-character",
		UpdateSource: "remote-peer",
	}

	resolvedState, err := cr.ResolveConflict(localState, incomingState)
	if err != nil {
		t.Fatalf("ResolveConflict() failed: %v", err)
	}

	if resolvedState.UpdateSource != "remote-peer" {
		t.Errorf("Winning source = %s, want remote-peer", resolvedState.UpdateSource)
	}
}

func TestConflictResolver_LastWriteWins(t *testing.T) {
	cr := NewConflictResolver(LastWriteWins)

	localState := &CharacterState{
		CharacterID:  "test-character",
		UpdateSource: "local-peer",
	}

	incomingState := &CharacterState{
		CharacterID:  "test-character",
		UpdateSource: "remote-peer",
	}

	resolvedState, err := cr.ResolveConflict(localState, incomingState)
	if err != nil {
		t.Fatalf("ResolveConflict() failed: %v", err)
	}

	// LastWriteWins always chooses incoming state
	if resolvedState.UpdateSource != "remote-peer" {
		t.Errorf("Winning source = %s, want remote-peer", resolvedState.UpdateSource)
	}
}

func TestConflictResolver_GetConflictStats(t *testing.T) {
	cr := NewConflictResolver(TimestampWins)

	// Initially no conflicts
	count, conflicts := cr.GetConflictStats()
	if count != 0 {
		t.Errorf("Initial conflict count = %d, want 0", count)
	}
	if len(conflicts) != 0 {
		t.Errorf("Initial conflicts length = %d, want 0", len(conflicts))
	}

	// Simulate conflict
	localState := &CharacterState{CharacterID: "test", UpdateSource: "local"}
	incomingState := &CharacterState{CharacterID: "test", UpdateSource: "remote", LastUpdate: time.Now()}

	_, err := cr.ResolveConflict(localState, incomingState)
	if err != nil {
		t.Fatalf("ResolveConflict() failed: %v", err)
	}

	// Check updated stats
	count, conflicts = cr.GetConflictStats()
	if count != 1 {
		t.Errorf("Conflict count after resolution = %d, want 1", count)
	}
	if len(conflicts) != 1 {
		t.Errorf("Conflicts length = %d, want 1", len(conflicts))
	}
}

// Test merge behavior

func TestStateSynchronizer_MergeIncomingState_NoLocalState(t *testing.T) {
	ss, _ := createTestStateSynchronizer("local-peer")

	incomingState := createTestCharacterState("new-character")

	err := ss.mergeIncomingState(incomingState)
	if err != nil {
		t.Fatalf("mergeIncomingState() failed: %v", err)
	}

	// Verify state was accepted
	state, exists := ss.GetCharacterState("new-character")
	if !exists {
		t.Fatal("Incoming state should be accepted when no local state exists")
	}

	if state.UpdateSource != incomingState.UpdateSource {
		t.Errorf("UpdateSource = %s, want %s", state.UpdateSource, incomingState.UpdateSource)
	}
}

func TestStateSynchronizer_MergeIncomingState_NoConflict(t *testing.T) {
	ss, _ := createTestStateSynchronizer("local-peer")

	characterID := "test-character"

	// Create local state
	localTime := time.Now()
	err := ss.UpdateCharacterState(characterID, Position{X: 100, Y: 200}, "idle", "active", nil, nil)
	if err != nil {
		t.Fatalf("UpdateCharacterState() failed: %v", err)
	}

	// Create incoming state with much newer timestamp (no conflict)
	incomingState := &CharacterState{
		CharacterID:  characterID,
		Position:     Position{X: 300, Y: 400},
		Animation:    "running",
		CurrentState: "excited",
		LastUpdate:   localTime.Add(10 * time.Second), // Much newer
		UpdateSource: "remote-peer",
		Version:      2,
		Checksum:     "new-checksum",
	}

	err = ss.mergeIncomingState(incomingState)
	if err != nil {
		t.Fatalf("mergeIncomingState() failed: %v", err)
	}

	// Verify incoming state was accepted
	state, exists := ss.GetCharacterState(characterID)
	if !exists {
		t.Fatal("Character state should exist after merge")
	}

	if state.Position.X != 300 {
		t.Errorf("Position.X = %f, want 300", state.Position.X)
	}

	if state.UpdateSource != "remote-peer" {
		t.Errorf("UpdateSource = %s, want remote-peer", state.UpdateSource)
	}
}

// Performance and stress tests

func TestStateSynchronizer_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	ss, _ := createTestStateSynchronizer("test-peer")

	// Measure performance of state updates
	start := time.Now()
	numUpdates := 1000

	for i := 0; i < numUpdates; i++ {
		characterID := "character-" + string(rune(i%10))
		position := Position{X: float32(i), Y: float32(i * 2)}
		err := ss.UpdateCharacterState(characterID, position, "idle", "active", nil, nil)
		if err != nil {
			t.Fatalf("UpdateCharacterState() failed at iteration %d: %v", i, err)
		}
	}

	duration := time.Since(start)
	avgPerUpdate := duration / time.Duration(numUpdates)

	t.Logf("Performance: %d state updates in %v (avg: %v per update)", numUpdates, duration, avgPerUpdate)

	// Performance target: should handle 1000 updates in reasonable time
	if avgPerUpdate > time.Millisecond {
		t.Errorf("Performance degraded: average time per update %v > 1ms", avgPerUpdate)
	}
}

func TestStateSynchronizer_ConcurrentAccess(t *testing.T) {
	ss, _ := createTestStateSynchronizer("test-peer")

	numGoroutines := 10
	numUpdatesPerGoroutine := 100
	done := make(chan bool, numGoroutines)

	// Start concurrent state updates
	for g := 0; g < numGoroutines; g++ {
		go func(goroutineID int) {
			defer func() { done <- true }()

			for i := 0; i < numUpdatesPerGoroutine; i++ {
				characterID := "character-" + string(rune(goroutineID))
				position := Position{X: float32(i), Y: float32(goroutineID)}
				err := ss.UpdateCharacterState(characterID, position, "idle", "active", nil, nil)
				if err != nil {
					t.Errorf("Goroutine %d: UpdateCharacterState() failed: %v", goroutineID, err)
					return
				}
			}
		}(g)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify final state count
	ss.mu.RLock()
	stateCount := len(ss.characterStates)
	ss.mu.RUnlock()

	if stateCount != numGoroutines {
		t.Errorf("Expected %d character states, got %d", numGoroutines, stateCount)
	}
}

// Edge cases and error handling

func TestStateSynchronizer_Stop_WithoutStart(t *testing.T) {
	ss, _ := createTestStateSynchronizer("test-peer")

	// Should handle stop without start gracefully
	err := ss.Stop()
	if err != nil {
		t.Errorf("Stop() without Start() failed: %v", err)
	}
}

func TestStateSynchronizer_Checksum_Generation(t *testing.T) {
	ss, _ := createTestStateSynchronizer("test-peer")

	state1 := &CharacterState{
		CharacterID:  "test",
		Position:     Position{X: 100, Y: 200},
		Animation:    "idle",
		CurrentState: "active",
		Version:      1,
	}

	state2 := &CharacterState{
		CharacterID:  "test",
		Position:     Position{X: 100, Y: 200},
		Animation:    "idle",
		CurrentState: "active",
		Version:      1,
	}

	checksum1, err1 := ss.calculateStateChecksum(state1)
	checksum2, err2 := ss.calculateStateChecksum(state2)

	if err1 != nil || err2 != nil {
		t.Fatalf("Checksum calculation failed: %v, %v", err1, err2)
	}

	if checksum1 != checksum2 {
		t.Error("Identical states should have identical checksums")
	}

	// Modify state2 and verify different checksum
	state2.Position.X = 150
	checksum3, err3 := ss.calculateStateChecksum(state2)
	if err3 != nil {
		t.Fatalf("Checksum calculation failed: %v", err3)
	}

	if checksum1 == checksum3 {
		t.Error("Different states should have different checksums")
	}
}
