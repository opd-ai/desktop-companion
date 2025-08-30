package network

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"time"
)

// ProtocolManager handles message signing, verification, and structured payloads
// for the DDS multiplayer protocol. Uses Go's standard library Ed25519 implementation
// following the project's "library-first" philosophy.
type ProtocolManager struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
	peerKeys   map[string]ed25519.PublicKey // peerID -> publicKey
}

// SignedMessage represents a message with Ed25519 signature verification
type SignedMessage struct {
	Message   Message `json:"message"`
	Signature []byte  `json:"signature"`
	PublicKey []byte  `json:"publicKey"` // Sender's public key for verification
}

// CharacterActionPayload represents character interaction data
type CharacterActionPayload struct {
	Action        string                 `json:"action"`              // "click", "feed", "play", "pet"
	CharacterID   string                 `json:"characterId"`         // Target character identifier
	Position      *Position              `json:"position,omitempty"`  // Optional position data
	Animation     string                 `json:"animation,omitempty"` // Animation triggered
	Response      string                 `json:"response,omitempty"`  // Dialog response
	Stats         map[string]float64     `json:"stats,omitempty"`     // Stat changes
	Metadata      map[string]interface{} `json:"metadata,omitempty"`  // Additional data
	InteractionID string                 `json:"interactionId"`       // Unique interaction identifier
}

// StateSyncPayload represents character state synchronization data
type StateSyncPayload struct {
	CharacterID  string             `json:"characterId"`
	Position     Position           `json:"position"`
	Animation    string             `json:"animation"`
	CurrentState string             `json:"currentState"`
	GameStats    map[string]float64 `json:"gameStats,omitempty"`    // Hunger, happiness, etc.
	RomanceStats map[string]float64 `json:"romanceStats,omitempty"` // Affection, trust, etc.
	LastUpdate   time.Time          `json:"lastUpdate"`
	Checksum     string             `json:"checksum"` // Data integrity verification
}

// Position represents 2D coordinates for character positioning
type Position struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}

// PeerListPayload extends the basic peer list with security information
type PeerListPayload struct {
	Peers     []SecurePeer `json:"peers"`
	Timestamp time.Time    `json:"timestamp"`
}

// BattleInvitePayload represents a battle invitation between peers
type BattleInvitePayload struct {
	FromCharacterID string    `json:"fromCharacterId"`
	ToCharacterID   string    `json:"toCharacterId"`
	BattleID        string    `json:"battleId"`
	Timestamp       time.Time `json:"timestamp"`
}

// BattleActionPayload represents a battle action performed by a participant
type BattleActionPayload struct {
	BattleID   string    `json:"battleId"`
	ActionType string    `json:"actionType"` // attack, heal, defend, etc.
	ActorID    string    `json:"actorId"`
	TargetID   string    `json:"targetId"`
	ItemUsed   string    `json:"itemUsed,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}

// BattleResultPayload represents the result of a battle action
type BattleResultPayload struct {
	BattleID         string                            `json:"battleId"`
	ActionType       string                            `json:"actionType"`
	ActorID          string                            `json:"actorId"`
	TargetID         string                            `json:"targetId"`
	Success          bool                              `json:"success"`
	Damage           float64                           `json:"damage"`
	Healing          float64                           `json:"healing"`
	Animation        string                            `json:"animation,omitempty"`
	Response         string                            `json:"response,omitempty"`
	ParticipantStats map[string]BattleParticipantStats `json:"participantStats"`
	Timestamp        time.Time                         `json:"timestamp"`
}

// BattleEndPayload represents the end of a battle
type BattleEndPayload struct {
	BattleID  string    `json:"battleId"`
	Winner    string    `json:"winner,omitempty"` // Empty for draw
	Reason    string    `json:"reason"`           // "defeat", "forfeit", "timeout"
	Timestamp time.Time `json:"timestamp"`
}

// BattleParticipantStats represents current battle stats for network sync
type BattleParticipantStats struct {
	HP      float64 `json:"hp"`
	MaxHP   float64 `json:"maxHp"`
	Attack  float64 `json:"attack"`
	Defense float64 `json:"defense"`
	Speed   float64 `json:"speed"`
}

// SecurePeer extends Peer with cryptographic identity
type SecurePeer struct {
	ID        string    `json:"id"`
	AddrStr   string    `json:"addr"`
	PublicKey []byte    `json:"publicKey"`
	LastSeen  time.Time `json:"lastSeen"`
	Verified  bool      `json:"verified"` // Whether the peer's identity is verified
}

// ExtendedDiscoveryPayload extends basic discovery with security features
type ExtendedDiscoveryPayload struct {
	DiscoveryPayload          // Embed basic discovery payload
	PublicKey        []byte   `json:"publicKey"`
	Capabilities     []string `json:"capabilities,omitempty"` // "bot", "game", "romance", etc.
	Version          string   `json:"version,omitempty"`      // Protocol version
}

// NewProtocolManager creates a new ProtocolManager with Ed25519 key generation.
// Uses crypto/rand for secure key generation following security best practices.
func NewProtocolManager() (*ProtocolManager, error) {
	// Generate Ed25519 key pair using Go's standard library
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Ed25519 keys: %w", err)
	}

	return &ProtocolManager{
		privateKey: privateKey,
		publicKey:  publicKey,
		peerKeys:   make(map[string]ed25519.PublicKey),
	}, nil
}

// GetPublicKey returns the public key for this peer
func (pm *ProtocolManager) GetPublicKey() ed25519.PublicKey {
	return pm.publicKey
}

// AddPeerKey stores a peer's public key for future verification
func (pm *ProtocolManager) AddPeerKey(peerID string, publicKey ed25519.PublicKey) error {
	if len(publicKey) != ed25519.PublicKeySize {
		return fmt.Errorf("invalid public key size: expected %d, got %d",
			ed25519.PublicKeySize, len(publicKey))
	}
	pm.peerKeys[peerID] = publicKey
	return nil
}

// SignMessage creates a signed message with Ed25519 signature
func (pm *ProtocolManager) SignMessage(msg Message) (*SignedMessage, error) {
	// Serialize message for signing
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	// Sign the serialized message
	signature := ed25519.Sign(pm.privateKey, msgBytes)

	return &SignedMessage{
		Message:   msg,
		Signature: signature,
		PublicKey: pm.publicKey,
	}, nil
}

// VerifyMessage verifies a signed message using Ed25519 signature verification
func (pm *ProtocolManager) VerifyMessage(signedMsg *SignedMessage) error {
	if signedMsg == nil {
		return fmt.Errorf("signed message is nil")
	}

	// Validate public key size
	if len(signedMsg.PublicKey) != ed25519.PublicKeySize {
		return fmt.Errorf("invalid public key size: %d", len(signedMsg.PublicKey))
	}

	// Serialize the message for verification
	msgBytes, err := json.Marshal(signedMsg.Message)
	if err != nil {
		return fmt.Errorf("failed to marshal message for verification: %w", err)
	}

	// Verify the signature
	if !ed25519.Verify(signedMsg.PublicKey, msgBytes, signedMsg.Signature) {
		return fmt.Errorf("signature verification failed")
	}

	// Store peer's public key for future use
	if signedMsg.Message.From != "" {
		pm.peerKeys[signedMsg.Message.From] = signedMsg.PublicKey
	}

	return nil
}

// CreateCharacterActionMessage creates a signed character action message
func (pm *ProtocolManager) CreateCharacterActionMessage(fromPeerID, toPeerID string, payload CharacterActionPayload) (*SignedMessage, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal character action payload: %w", err)
	}

	msg := Message{
		Type:      MessageTypeCharacterAction,
		From:      fromPeerID,
		To:        toPeerID,
		Payload:   payloadBytes,
		Timestamp: time.Now(),
	}

	return pm.SignMessage(msg)
}

// CreateStateSyncMessage creates a signed state synchronization message
func (pm *ProtocolManager) CreateStateSyncMessage(fromPeerID string, payload StateSyncPayload) (*SignedMessage, error) {
	// Generate checksum for data integrity
	payload.Checksum = pm.generateChecksum(payload)

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal state sync payload: %w", err)
	}

	msg := Message{
		Type:      MessageTypeStateSync,
		From:      fromPeerID,
		To:        "", // Broadcast to all peers
		Payload:   payloadBytes,
		Timestamp: time.Now(),
	}

	return pm.SignMessage(msg)
}

// CreateSecureDiscoveryMessage creates a signed discovery message with public key
func (pm *ProtocolManager) CreateSecureDiscoveryMessage(basicPayload DiscoveryPayload, capabilities []string) (*SignedMessage, error) {
	extendedPayload := ExtendedDiscoveryPayload{
		DiscoveryPayload: basicPayload,
		PublicKey:        pm.publicKey,
		Capabilities:     capabilities,
		Version:          "1.0", // Protocol version
	}

	payloadBytes, err := json.Marshal(extendedPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal extended discovery payload: %w", err)
	}

	msg := Message{
		Type:      MessageTypeDiscovery,
		From:      basicPayload.PeerID,
		To:        "", // Broadcast
		Payload:   payloadBytes,
		Timestamp: time.Now(),
	}

	return pm.SignMessage(msg)
}

// ParseCharacterActionPayload parses and validates a character action payload
func (pm *ProtocolManager) ParseCharacterActionPayload(msg Message) (*CharacterActionPayload, error) {
	if msg.Type != MessageTypeCharacterAction {
		return nil, fmt.Errorf("message type must be character_action, got %s", msg.Type)
	}

	var payload CharacterActionPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal character action payload: %w", err)
	}

	// Validate required fields
	if payload.Action == "" {
		return nil, fmt.Errorf("action field is required")
	}
	if payload.CharacterID == "" {
		return nil, fmt.Errorf("characterId field is required")
	}
	if payload.InteractionID == "" {
		return nil, fmt.Errorf("interactionId field is required")
	}

	return &payload, nil
}

// ParseStateSyncPayload parses and validates a state synchronization payload
func (pm *ProtocolManager) ParseStateSyncPayload(msg Message) (*StateSyncPayload, error) {
	if msg.Type != MessageTypeStateSync {
		return nil, fmt.Errorf("message type must be state_sync, got %s", msg.Type)
	}

	var payload StateSyncPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state sync payload: %w", err)
	}

	// Validate required fields
	if payload.CharacterID == "" {
		return nil, fmt.Errorf("characterId field is required")
	}
	if payload.CurrentState == "" {
		return nil, fmt.Errorf("currentState field is required")
	}

	// Verify checksum for data integrity
	expectedChecksum := pm.generateChecksum(payload)
	if payload.Checksum != expectedChecksum {
		return nil, fmt.Errorf("checksum verification failed: expected %s, got %s",
			expectedChecksum, payload.Checksum)
	}

	return &payload, nil
}

// ParseExtendedDiscoveryPayload parses a secure discovery payload
func (pm *ProtocolManager) ParseExtendedDiscoveryPayload(msg Message) (*ExtendedDiscoveryPayload, error) {
	if msg.Type != MessageTypeDiscovery {
		return nil, fmt.Errorf("message type must be discovery, got %s", msg.Type)
	}

	var payload ExtendedDiscoveryPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal extended discovery payload: %w", err)
	}

	// Validate required fields
	if payload.NetworkID == "" {
		return nil, fmt.Errorf("networkId field is required")
	}
	if payload.PeerID == "" {
		return nil, fmt.Errorf("peerId field is required")
	}
	if len(payload.PublicKey) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key size: %d", len(payload.PublicKey))
	}

	return &payload, nil
}

// IsPeerVerified checks if a peer's identity has been cryptographically verified
func (pm *ProtocolManager) IsPeerVerified(peerID string) bool {
	_, exists := pm.peerKeys[peerID]
	return exists
}

// GetVerifiedPeers returns a list of all cryptographically verified peers
func (pm *ProtocolManager) GetVerifiedPeers() []string {
	peers := make([]string, 0, len(pm.peerKeys))
	for peerID := range pm.peerKeys {
		peers = append(peers, peerID)
	}
	return peers
}

// generateChecksum creates a simple checksum for state sync data integrity
// This is a basic implementation - in production, consider using a stronger hash
func (pm *ProtocolManager) generateChecksum(payload StateSyncPayload) string {
	// Create a copy without the checksum field to avoid circular dependency
	temp := payload
	temp.Checksum = ""

	data, err := json.Marshal(temp)
	if err != nil {
		return "" // Return empty string on error
	}

	// Simple checksum: sum of bytes modulo a large prime
	var sum uint64
	for _, b := range data {
		sum += uint64(b)
	}
	return fmt.Sprintf("%x", sum%982451653) // Large prime for distribution
}

// ValidateMessageAge checks if a message is within acceptable time bounds
// Helps prevent replay attacks by rejecting old messages
func (pm *ProtocolManager) ValidateMessageAge(msg Message, maxAge time.Duration) error {
	age := time.Since(msg.Timestamp)
	if age > maxAge {
		return fmt.Errorf("message too old: %v (max age: %v)", age, maxAge)
	}
	if age < -time.Minute {
		return fmt.Errorf("message from future: %v", age)
	}
	return nil
}

// CreatePeerListMessage creates a signed peer list message
func (pm *ProtocolManager) CreatePeerListMessage(fromPeerID string, peers []SecurePeer) (*SignedMessage, error) {
	payload := PeerListPayload{
		Peers:     peers,
		Timestamp: time.Now(),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal peer list payload: %w", err)
	}

	msg := Message{
		Type:      MessageTypePeerList,
		From:      fromPeerID,
		To:        "", // Broadcast
		Payload:   payloadBytes,
		Timestamp: time.Now(),
	}

	return pm.SignMessage(msg)
}

// CreateBattleInviteMessage creates a signed battle invitation message
func (pm *ProtocolManager) CreateBattleInviteMessage(fromPeerID, toPeerID string, payload BattleInvitePayload) (*SignedMessage, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal battle invite payload: %w", err)
	}

	msg := Message{
		Type:      MessageTypeBattleInvite,
		From:      fromPeerID,
		To:        toPeerID,
		Payload:   payloadBytes,
		Timestamp: time.Now(),
	}

	return pm.SignMessage(msg)
}

// CreateBattleActionMessage creates a signed battle action message
func (pm *ProtocolManager) CreateBattleActionMessage(fromPeerID, toPeerID string, payload BattleActionPayload) (*SignedMessage, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal battle action payload: %w", err)
	}

	msg := Message{
		Type:      MessageTypeBattleAction,
		From:      fromPeerID,
		To:        toPeerID,
		Payload:   payloadBytes,
		Timestamp: time.Now(),
	}

	return pm.SignMessage(msg)
}

// CreateBattleResultMessage creates a signed battle result message
func (pm *ProtocolManager) CreateBattleResultMessage(fromPeerID, toPeerID string, payload BattleResultPayload) (*SignedMessage, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal battle result payload: %w", err)
	}

	msg := Message{
		Type:      MessageTypeBattleResult,
		From:      fromPeerID,
		To:        toPeerID,
		Payload:   payloadBytes,
		Timestamp: time.Now(),
	}

	return pm.SignMessage(msg)
}

// CreateBattleEndMessage creates a signed battle end message
func (pm *ProtocolManager) CreateBattleEndMessage(fromPeerID, toPeerID string, payload BattleEndPayload) (*SignedMessage, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal battle end payload: %w", err)
	}

	msg := Message{
		Type:      MessageTypeBattleEnd,
		From:      fromPeerID,
		To:        toPeerID,
		Payload:   payloadBytes,
		Timestamp: time.Now(),
	}

	return pm.SignMessage(msg)
}
