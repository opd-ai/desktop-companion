package character

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// NetworkManager interface to avoid circular imports
type NetworkManager interface {
	Start() error
	Stop() error
	Broadcast(msg NetworkMessage) error
	RegisterHandler(msgType string, handler func(NetworkMessage, interface{}) error)
}

// ProtocolManager interface to avoid circular imports
type ProtocolManager interface {
	SignMessage(data []byte) ([]byte, error)
	VerifyMessage(data []byte, signature []byte, publicKey []byte) error
}

// NetworkMessage represents a network message
type NetworkMessage struct {
	Type      string    `json:"type"`
	From      string    `json:"from"`
	To        string    `json:"to,omitempty"`
	Payload   []byte    `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
}

// CharacterActionPayload represents character interaction data
type CharacterActionPayload struct {
	Action        string                 `json:"action"`
	CharacterID   string                 `json:"characterId"`
	Position      *NetworkPosition       `json:"position,omitempty"`
	Animation     string                 `json:"animation,omitempty"`
	Response      string                 `json:"response,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	InteractionID string                 `json:"interactionId"`
}

// NetworkPosition represents 2D coordinates
type NetworkPosition struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}

// MultiplayerCharacter wraps the existing Character struct with network coordination
// capabilities while preserving 100% backward compatibility with the Character interface.
// Follows the project's embedding pattern and "library-first" philosophy.
type MultiplayerCharacter struct {
	*Character // Embed existing Character - preserves all methods

	// Network coordination
	networkManager  NetworkManager
	protocolManager ProtocolManager
	characterID     string
	syncInterval    time.Duration
	lastSyncTime    time.Time
	pendingActions  []CharacterActionPayload

	// Multiplayer state
	mu               sync.RWMutex
	networkEnabled   bool
	broadcastActions bool
	enableStateSync  bool   // Renamed to avoid conflict
	currentBattleID  string // Track active battle ID (Finding #3 fix)
}

// MultiplayerWrapperConfig configures network behavior for the character wrapper
type MultiplayerWrapperConfig struct {
	CharacterID      string        `json:"characterId"`
	SyncInterval     time.Duration `json:"syncInterval"`
	BroadcastActions bool          `json:"broadcastActions"`
	EnableStateSync  bool          `json:"enableStateSync"`
}

// NewMultiplayerCharacter creates a new multiplayer-enabled character wrapper.
// Uses existing Character constructor and adds network coordination.
func NewMultiplayerCharacter(card *CharacterCard, config MultiplayerWrapperConfig,
	networkManager NetworkManager, protocolManager ProtocolManager) (*MultiplayerCharacter, error) {

	if card == nil {
		return nil, fmt.Errorf("character card cannot be nil")
	}
	if networkManager == nil {
		return nil, fmt.Errorf("network manager cannot be nil")
	}
	if protocolManager == nil {
		return nil, fmt.Errorf("protocol manager cannot be nil")
	}

	// Create base character using existing constructor
	baseChar, err := New(card, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create base character: %w", err)
	}

	// Set default config values
	if config.CharacterID == "" {
		config.CharacterID = fmt.Sprintf("char_%d", time.Now().UnixNano())
	}
	if config.SyncInterval <= 0 {
		config.SyncInterval = 5 * time.Second
	}

	mc := &MultiplayerCharacter{
		Character:        baseChar,
		networkManager:   networkManager,
		protocolManager:  protocolManager,
		characterID:      config.CharacterID,
		syncInterval:     config.SyncInterval,
		broadcastActions: config.BroadcastActions,
		enableStateSync:  config.EnableStateSync,
		pendingActions:   make([]CharacterActionPayload, 0),
	}

	// Register message handlers
	if err := mc.setupNetworkHandlers(); err != nil {
		return nil, fmt.Errorf("failed to setup network handlers: %w", err)
	}

	return mc, nil
}

// EnableNetworking activates multiplayer functionality
func (mc *MultiplayerCharacter) EnableNetworking() error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if mc.networkEnabled {
		return nil // Already enabled
	}

	// Start network manager if not already running
	if err := mc.networkManager.Start(); err != nil {
		return fmt.Errorf("failed to start network manager: %w", err)
	}

	mc.networkEnabled = true

	// Start periodic state sync if enabled
	if mc.enableStateSync {
		go mc.periodicStateSync()
	}

	return nil
}

// DisableNetworking deactivates multiplayer functionality
func (mc *MultiplayerCharacter) DisableNetworking() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.networkEnabled = false
}

// HandleClick overrides the base Character's HandleClick to add network broadcasting
func (mc *MultiplayerCharacter) HandleClick() string {
	// Call original HandleClick
	response := mc.Character.HandleClick()

	// Broadcast action if networking enabled
	mc.mu.RLock()
	shouldBroadcast := mc.networkEnabled && mc.broadcastActions
	mc.mu.RUnlock()

	if shouldBroadcast {
		action := CharacterActionPayload{
			Action:        "click",
			CharacterID:   mc.characterID,
			Response:      response,
			InteractionID: fmt.Sprintf("click_%d", time.Now().UnixNano()),
		}
		mc.broadcastAction(action)
	}

	return response
}

// HandleRightClick overrides with network broadcasting
func (mc *MultiplayerCharacter) HandleRightClick() string {
	response := mc.Character.HandleRightClick()

	mc.mu.RLock()
	shouldBroadcast := mc.networkEnabled && mc.broadcastActions
	mc.mu.RUnlock()

	if shouldBroadcast {
		action := CharacterActionPayload{
			Action:        "rightclick",
			CharacterID:   mc.characterID,
			Response:      response,
			InteractionID: fmt.Sprintf("rightclick_%d", time.Now().UnixNano()),
		}
		mc.broadcastAction(action)
	}

	return response
}

// HandleGameInteraction overrides with network broadcasting
func (mc *MultiplayerCharacter) HandleGameInteraction(actionType string) string {
	response := mc.Character.HandleGameInteraction(actionType)

	mc.mu.RLock()
	shouldBroadcast := mc.networkEnabled && mc.broadcastActions
	mc.mu.RUnlock()

	if shouldBroadcast {
		actionPayload := CharacterActionPayload{
			Action:        actionType,
			CharacterID:   mc.characterID,
			Response:      response,
			InteractionID: fmt.Sprintf("game_%s_%d", actionType, time.Now().UnixNano()),
		}
		mc.broadcastAction(actionPayload)
	}

	return response
}

// broadcastAction sends character action to all connected peers
func (mc *MultiplayerCharacter) broadcastAction(action CharacterActionPayload) {
	payload, err := json.Marshal(action)
	if err != nil {
		return // Silently ignore marshaling errors
	}

	msg := NetworkMessage{
		Type:      "character_action",
		From:      mc.characterID,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	// Use network manager to broadcast (non-blocking)
	go func() {
		if err := mc.networkManager.Broadcast(msg); err != nil {
			// Log error in production, ignore for now
		}
	}()
}

// syncState broadcasts current character state to peers
func (mc *MultiplayerCharacter) syncState() error {
	mc.mu.RLock()
	if !mc.networkEnabled || !mc.enableStateSync {
		mc.mu.RUnlock()
		return nil
	}
	mc.mu.RUnlock()

	// Get current character state
	posX, posY := mc.Character.GetPosition()
	state := mc.Character.GetCurrentState()

	syncPayload := map[string]interface{}{
		"characterId":  mc.characterID,
		"position":     NetworkPosition{X: posX, Y: posY},
		"currentState": state,
		"lastUpdate":   time.Now(),
	}

	payload, err := json.Marshal(syncPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal sync payload: %w", err)
	}

	msg := NetworkMessage{
		Type:      "state_sync",
		From:      mc.characterID,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	return mc.networkManager.Broadcast(msg)
}

// periodicStateSync runs state synchronization at regular intervals
func (mc *MultiplayerCharacter) periodicStateSync() {
	ticker := time.NewTicker(mc.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mc.mu.RLock()
			enabled := mc.networkEnabled && mc.enableStateSync
			mc.mu.RUnlock()

			if !enabled {
				return
			}

			if err := mc.syncState(); err != nil {
				// Log error in production, continue for now
			}
		}
	}
}

// setupNetworkHandlers registers message handlers for network events
func (mc *MultiplayerCharacter) setupNetworkHandlers() error {
	// Handle incoming character actions from peers
	actionHandler := func(msg NetworkMessage, from interface{}) error {
		var action CharacterActionPayload
		if err := json.Unmarshal(msg.Payload, &action); err != nil {
			return fmt.Errorf("failed to unmarshal action payload: %w", err)
		}

		// Only process actions for this character
		if action.CharacterID != mc.characterID {
			return nil
		}

		return mc.handleRemoteAction(action)
	}

	// Handle incoming state sync from peers
	syncHandler := func(msg NetworkMessage, from interface{}) error {
		var sync map[string]interface{}
		if err := json.Unmarshal(msg.Payload, &sync); err != nil {
			return fmt.Errorf("failed to unmarshal sync payload: %w", err)
		}

		// Only process sync for this character
		if charID, ok := sync["characterId"].(string); !ok || charID != mc.characterID {
			return nil
		}

		return mc.handleRemoteSync(sync)
	}

	// Register handlers with network manager
	mc.networkManager.RegisterHandler("character_action", actionHandler)
	mc.networkManager.RegisterHandler("state_sync", syncHandler)

	return nil
}

// handleRemoteAction processes character actions received from peers
func (mc *MultiplayerCharacter) handleRemoteAction(action CharacterActionPayload) error {
	switch action.Action {
	case "click":
		mc.Character.HandleClick()
	case "rightclick":
		mc.Character.HandleRightClick()
	default:
		mc.Character.HandleGameInteraction(action.Action)
	}
	return nil
}

// handleRemoteSync processes state synchronization from peers
func (mc *MultiplayerCharacter) handleRemoteSync(sync map[string]interface{}) error {
	// Update character position if provided
	if posData, ok := sync["position"].(map[string]interface{}); ok {
		if x, okX := posData["x"].(float64); okX {
			if y, okY := posData["y"].(float64); okY {
				currentX, currentY := mc.Character.GetPosition()
				if currentX != float32(x) || currentY != float32(y) {
					mc.Character.SetPosition(float32(x), float32(y))
				}
			}
		}
	}

	return nil
}

// GetCharacterID returns the unique character identifier for networking
func (mc *MultiplayerCharacter) GetCharacterID() string {
	return mc.characterID
}

// IsNetworkEnabled returns whether networking is currently active
func (mc *MultiplayerCharacter) IsNetworkEnabled() bool {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.networkEnabled
}

// GetNetworkStats returns networking statistics and status
func (mc *MultiplayerCharacter) GetNetworkStats() map[string]interface{} {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	stats := map[string]interface{}{
		"networkEnabled":   mc.networkEnabled,
		"broadcastActions": mc.broadcastActions,
		"enableStateSync":  mc.enableStateSync,
		"characterId":      mc.characterID,
		"syncInterval":     mc.syncInterval.String(),
		"lastSyncTime":     mc.lastSyncTime,
		"pendingActions":   len(mc.pendingActions),
	}

	return stats
}
