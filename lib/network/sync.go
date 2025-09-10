package network

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// NetworkManagerInterface defines the interface for network management operations
type NetworkManagerInterface interface {
	RegisterMessageHandler(msgType MessageType, handler MessageHandler)
	SendMessage(msgType MessageType, payload []byte, targetPeerID string) error
	GetNetworkID() string
}

// ProtocolManagerInterface defines the interface for protocol management operations
type ProtocolManagerInterface interface {
	GetPublicKey() ed25519.PublicKey
}

// StateSynchronizer manages real-time character state synchronization across peers.
// Uses Go standard library for hashing and JSON serialization following the project's
// "library-first" philosophy. Implements conflict resolution for simultaneous actions.
type StateSynchronizer struct {
	mu sync.RWMutex

	// Network components - using interfaces for testability
	networkManager  NetworkManagerInterface
	protocolManager ProtocolManagerInterface

	// Synchronization state
	characterStates  map[string]*CharacterState // characterID -> state
	lastSyncTimes    map[string]time.Time       // characterID -> last sync time
	syncInterval     time.Duration
	conflictResolver *ConflictResolver

	// Lifecycle management
	ctx    context.Context
	cancel context.CancelFunc
	ticker *time.Ticker
	wg     sync.WaitGroup
}

// CharacterState represents the current state of a character for synchronization.
// Contains all data needed for real-time state sharing across peers.
type CharacterState struct {
	CharacterID  string             `json:"characterId"`
	Position     Position           `json:"position"`
	Animation    string             `json:"animation"`
	CurrentState string             `json:"currentState"`
	GameStats    map[string]float64 `json:"gameStats,omitempty"`
	RomanceStats map[string]float64 `json:"romanceStats,omitempty"`
	LastUpdate   time.Time          `json:"lastUpdate"`
	UpdateSource string             `json:"updateSource"` // Which peer updated this
	Version      int64              `json:"version"`      // Monotonic version counter
	Checksum     string             `json:"checksum"`     // SHA256 for integrity
}

// ConflictResolver handles conflicts when multiple peers update the same character
// simultaneously. Uses timestamp-based resolution with fallback to peer priority.
type ConflictResolver struct {
	mu sync.RWMutex

	// Conflict resolution strategy
	strategy       ConflictStrategy
	peerPriorities map[string]int // peerID -> priority level (higher = more priority)

	// Conflict tracking
	conflictCount     int64
	resolvedConflicts map[string]ConflictResolution // characterID -> last resolution
}

// ConflictStrategy defines how to resolve state conflicts
type ConflictStrategy int

const (
	// TimestampWins uses the most recent timestamp to resolve conflicts
	TimestampWins ConflictStrategy = iota
	// PeerPriorityWins uses configured peer priorities
	PeerPriorityWins
	// LastWriteWins accepts the last received update (least safe)
	LastWriteWins
)

// ConflictResolution tracks how a conflict was resolved
type ConflictResolution struct {
	ConflictTime   time.Time `json:"conflictTime"`
	WinningPeer    string    `json:"winningPeer"`
	LosingPeer     string    `json:"losingPeer"`
	Strategy       string    `json:"strategy"`
	CharacterState string    `json:"characterState"`
}

// NewStateSynchronizer creates a new state synchronizer with default settings.
// Uses 30-second sync intervals for balance between responsiveness and network usage.
func NewStateSynchronizer(networkManager NetworkManagerInterface, protocolManager ProtocolManagerInterface) *StateSynchronizer {
	ctx, cancel := context.WithCancel(context.Background())

	return &StateSynchronizer{
		networkManager:   networkManager,
		protocolManager:  protocolManager,
		characterStates:  make(map[string]*CharacterState),
		lastSyncTimes:    make(map[string]time.Time),
		syncInterval:     30 * time.Second, // Default sync interval
		conflictResolver: NewConflictResolver(TimestampWins),
		ctx:              ctx,
		cancel:           cancel,
	}
}

// NewConflictResolver creates a new conflict resolver with the specified strategy
func NewConflictResolver(strategy ConflictStrategy) *ConflictResolver {
	return &ConflictResolver{
		strategy:          strategy,
		peerPriorities:    make(map[string]int),
		resolvedConflicts: make(map[string]ConflictResolution),
	}
}

// Start begins the state synchronization process with periodic sync broadcasts
func (ss *StateSynchronizer) Start() error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	// Register message handler for incoming state sync messages
	if ss.networkManager != nil {
		ss.networkManager.RegisterMessageHandler(MessageTypeStateSync, ss.handleStateSyncMessage)
	}

	// Start periodic sync ticker
	ss.ticker = time.NewTicker(ss.syncInterval)
	ss.wg.Add(1)

	go ss.syncLoop()

	return nil
}

// Stop gracefully shuts down the state synchronizer
func (ss *StateSynchronizer) Stop() error {
	ss.cancel()

	if ss.ticker != nil {
		ss.ticker.Stop()
	}

	// Wait for sync loop to finish with timeout
	done := make(chan struct{})
	go func() {
		ss.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout waiting for sync loop to stop")
	}
}

// UpdateCharacterState updates the local state for a character and marks it for sync
func (ss *StateSynchronizer) UpdateCharacterState(characterID string, position Position,
	animation, currentState string, gameStats, romanceStats map[string]float64,
) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	now := time.Now()

	// Get existing state or create new one
	state, exists := ss.characterStates[characterID]
	if !exists {
		state = &CharacterState{
			CharacterID: characterID,
			Version:     1,
		}
		ss.characterStates[characterID] = state
	} else {
		state.Version++
	}

	// Update state fields
	state.Position = position
	state.Animation = animation
	state.CurrentState = currentState
	state.GameStats = gameStats
	state.RomanceStats = romanceStats
	state.LastUpdate = now
	state.UpdateSource = ss.getLocalPeerID()

	// Calculate checksum for integrity verification
	checksum, err := ss.calculateStateChecksum(state)
	if err != nil {
		return fmt.Errorf("failed to calculate state checksum: %w", err)
	}
	state.Checksum = checksum

	// Mark for immediate sync if significant change
	if ss.isSignificantChange(state, exists) {
		ss.lastSyncTimes[characterID] = time.Time{} // Force immediate sync
	}

	return nil
}

// GetCharacterState returns the current synchronized state for a character
func (ss *StateSynchronizer) GetCharacterState(characterID string) (*CharacterState, bool) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	state, exists := ss.characterStates[characterID]
	if !exists {
		return nil, false
	}

	// Return a copy to prevent external modification
	stateCopy := *state
	return &stateCopy, true
}

// SetSyncInterval configures how frequently states are synchronized
func (ss *StateSynchronizer) SetSyncInterval(interval time.Duration) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	ss.syncInterval = interval

	// Restart ticker if running
	if ss.ticker != nil {
		ss.ticker.Stop()
		ss.ticker = time.NewTicker(interval)
	}
}

// syncLoop runs the periodic state synchronization
func (ss *StateSynchronizer) syncLoop() {
	defer ss.wg.Done()

	for {
		select {
		case <-ss.ctx.Done():
			return
		case <-ss.ticker.C:
			ss.performPeriodicSync()
		}
	}
}

// performPeriodicSync sends state updates for characters that need synchronization
func (ss *StateSynchronizer) performPeriodicSync() {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	now := time.Now()

	for characterID, state := range ss.characterStates {
		lastSync, exists := ss.lastSyncTimes[characterID]

		// Sync if never synced or interval elapsed
		if !exists || now.Sub(lastSync) >= ss.syncInterval {
			ss.sendStateSync(state)
			ss.lastSyncTimes[characterID] = now
		}
	}
}

// sendStateSync broadcasts a state sync message to all peers
func (ss *StateSynchronizer) sendStateSync(state *CharacterState) {
	if ss.networkManager == nil {
		return
	}

	// Create state sync payload
	payload := StateSyncPayload{
		CharacterID:  state.CharacterID,
		Position:     state.Position,
		Animation:    state.Animation,
		CurrentState: state.CurrentState,
		GameStats:    state.GameStats,
		RomanceStats: state.RomanceStats,
		LastUpdate:   state.LastUpdate,
		Checksum:     state.Checksum,
	}

	// Serialize payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return // Log error in production
	}

	// Create and send message
	message := Message{
		Type:      MessageTypeStateSync,
		From:      ss.getLocalPeerID(),
		Payload:   payloadBytes,
		Timestamp: time.Now(),
	}

	// Send message to all peers (broadcast)
	msgBytes, marshalErr := json.Marshal(message)
	if marshalErr != nil {
		return // Log error in production
	}

	ss.networkManager.SendMessage(message.Type, msgBytes, "")
}

// handleStateSyncMessage processes incoming state sync messages from peers
func (ss *StateSynchronizer) handleStateSyncMessage(message Message, peer *Peer) error {
	var payload StateSyncPayload
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal state sync payload: %w", err)
	}

	// Verify checksum integrity
	expectedChecksum := ss.calculatePayloadChecksum(payload)
	if payload.Checksum != expectedChecksum {
		return fmt.Errorf("state sync checksum mismatch for character %s", payload.CharacterID)
	}

	// Convert to internal state representation
	incomingState := &CharacterState{
		CharacterID:  payload.CharacterID,
		Position:     payload.Position,
		Animation:    payload.Animation,
		CurrentState: payload.CurrentState,
		GameStats:    payload.GameStats,
		RomanceStats: payload.RomanceStats,
		LastUpdate:   payload.LastUpdate,
		UpdateSource: message.From,
		Checksum:     payload.Checksum,
	}

	return ss.mergeIncomingState(incomingState)
}

// mergeIncomingState integrates incoming state with local state, resolving conflicts
func (ss *StateSynchronizer) mergeIncomingState(incomingState *CharacterState) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	characterID := incomingState.CharacterID
	localState, exists := ss.characterStates[characterID]

	if !exists {
		// No local state - accept incoming state
		ss.characterStates[characterID] = incomingState
		return nil
	}

	// Check for conflicts (updates within conflict window)
	conflictWindow := 5 * time.Second
	timeDiff := incomingState.LastUpdate.Sub(localState.LastUpdate)

	if timeDiff.Abs() < conflictWindow && incomingState.UpdateSource != localState.UpdateSource {
		// Conflict detected - resolve using conflict resolver
		resolvedState, err := ss.conflictResolver.ResolveConflict(localState, incomingState)
		if err != nil {
			return fmt.Errorf("failed to resolve state conflict: %w", err)
		}
		ss.characterStates[characterID] = resolvedState
	} else if incomingState.LastUpdate.After(localState.LastUpdate) {
		// No conflict - incoming state is newer
		ss.characterStates[characterID] = incomingState
	}
	// If local state is newer, keep local state

	return nil
}

// ResolveConflict resolves conflicts between two character states
func (cr *ConflictResolver) ResolveConflict(localState, incomingState *CharacterState) (*CharacterState, error) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	cr.conflictCount++

	var winningState *CharacterState
	var winningPeer, losingPeer string

	switch cr.strategy {
	case TimestampWins:
		if incomingState.LastUpdate.After(localState.LastUpdate) {
			winningState = incomingState
			winningPeer = incomingState.UpdateSource
			losingPeer = localState.UpdateSource
		} else {
			winningState = localState
			winningPeer = localState.UpdateSource
			losingPeer = incomingState.UpdateSource
		}

	case PeerPriorityWins:
		localPriority := cr.peerPriorities[localState.UpdateSource]
		incomingPriority := cr.peerPriorities[incomingState.UpdateSource]

		if incomingPriority > localPriority {
			winningState = incomingState
			winningPeer = incomingState.UpdateSource
			losingPeer = localState.UpdateSource
		} else {
			winningState = localState
			winningPeer = localState.UpdateSource
			losingPeer = incomingState.UpdateSource
		}

	case LastWriteWins:
		winningState = incomingState
		winningPeer = incomingState.UpdateSource
		losingPeer = localState.UpdateSource

	default:
		return nil, fmt.Errorf("unknown conflict resolution strategy: %v", cr.strategy)
	}

	// Record the conflict resolution
	resolution := ConflictResolution{
		ConflictTime:   time.Now(),
		WinningPeer:    winningPeer,
		LosingPeer:     losingPeer,
		Strategy:       cr.getStrategyName(),
		CharacterState: winningState.CharacterID,
	}
	cr.resolvedConflicts[winningState.CharacterID] = resolution

	return winningState, nil
}

// Helper methods

// calculateStateChecksum generates SHA256 checksum for state integrity verification
func (ss *StateSynchronizer) calculateStateChecksum(state *CharacterState) (string, error) {
	// Create copy without checksum and volatile fields for consistent hashing
	hashData := struct {
		CharacterID  string             `json:"characterId"`
		Position     Position           `json:"position"`
		Animation    string             `json:"animation"`
		CurrentState string             `json:"currentState"`
		GameStats    map[string]float64 `json:"gameStats,omitempty"`
		RomanceStats map[string]float64 `json:"romanceStats,omitempty"`
		Version      int64              `json:"version"`
	}{
		CharacterID:  state.CharacterID,
		Position:     state.Position,
		Animation:    state.Animation,
		CurrentState: state.CurrentState,
		GameStats:    state.GameStats,
		RomanceStats: state.RomanceStats,
		Version:      state.Version,
	}

	data, err := json.Marshal(hashData)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

// calculatePayloadChecksum generates checksum for StateSyncPayload verification
func (ss *StateSynchronizer) calculatePayloadChecksum(payload StateSyncPayload) string {
	// Create consistent hash data
	hashData := struct {
		CharacterID  string             `json:"characterId"`
		Position     Position           `json:"position"`
		Animation    string             `json:"animation"`
		CurrentState string             `json:"currentState"`
		GameStats    map[string]float64 `json:"gameStats,omitempty"`
		RomanceStats map[string]float64 `json:"romanceStats,omitempty"`
	}{
		CharacterID:  payload.CharacterID,
		Position:     payload.Position,
		Animation:    payload.Animation,
		CurrentState: payload.CurrentState,
		GameStats:    payload.GameStats,
		RomanceStats: payload.RomanceStats,
	}

	data, _ := json.Marshal(hashData)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// isSignificantChange determines if a state change warrants immediate synchronization
func (ss *StateSynchronizer) isSignificantChange(state *CharacterState, hadPreviousState bool) bool {
	if !hadPreviousState {
		return true // First state always significant
	}

	// Position changes are significant (character movement)
	// Animation changes are significant (visual feedback)
	// State changes are significant (idle -> active, etc.)
	// Large stat changes are significant (>10% change)

	return true // For now, consider all changes significant for real-time feel
}

// getLocalPeerID returns the local peer's identifier
func (ss *StateSynchronizer) getLocalPeerID() string {
	if ss.networkManager != nil {
		return ss.networkManager.GetNetworkID()
	}
	return "local"
}

// getStrategyName returns human-readable strategy name
func (cr *ConflictResolver) getStrategyName() string {
	switch cr.strategy {
	case TimestampWins:
		return "timestamp_wins"
	case PeerPriorityWins:
		return "peer_priority_wins"
	case LastWriteWins:
		return "last_write_wins"
	default:
		return "unknown"
	}
}

// SetPeerPriority configures priority for conflict resolution
func (cr *ConflictResolver) SetPeerPriority(peerID string, priority int) {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	cr.peerPriorities[peerID] = priority
}

// GetConflictStats returns conflict resolution statistics
func (cr *ConflictResolver) GetConflictStats() (int64, map[string]ConflictResolution) {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	// Return copy of conflicts map
	conflicts := make(map[string]ConflictResolution)
	for k, v := range cr.resolvedConflicts {
		conflicts[k] = v
	}

	return cr.conflictCount, conflicts
}
