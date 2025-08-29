package network

import (
	"crypto/ed25519"
	"encoding/json"
	"testing"
	"time"
)

func TestNewProtocolManager(t *testing.T) {
	pm, err := NewProtocolManager()
	if err != nil {
		t.Fatalf("NewProtocolManager() error = %v", err)
	}

	if pm == nil {
		t.Fatal("NewProtocolManager() returned nil")
	}

	// Verify key generation
	if len(pm.publicKey) != ed25519.PublicKeySize {
		t.Errorf("public key size = %d, want %d", len(pm.publicKey), ed25519.PublicKeySize)
	}
	if len(pm.privateKey) != ed25519.PrivateKeySize {
		t.Errorf("private key size = %d, want %d", len(pm.privateKey), ed25519.PrivateKeySize)
	}

	// Verify internal structures
	if pm.peerKeys == nil {
		t.Error("peerKeys map not initialized")
	}

	// Verify GetPublicKey
	pubKey := pm.GetPublicKey()
	if len(pubKey) != ed25519.PublicKeySize {
		t.Errorf("GetPublicKey() size = %d, want %d", len(pubKey), ed25519.PublicKeySize)
	}
}

func TestProtocolManager_AddPeerKey(t *testing.T) {
	pm, err := NewProtocolManager()
	if err != nil {
		t.Fatalf("NewProtocolManager() error = %v", err)
	}

	// Test valid key addition
	validKey := make([]byte, ed25519.PublicKeySize)
	err = pm.AddPeerKey("test-peer", validKey)
	if err != nil {
		t.Errorf("AddPeerKey() with valid key error = %v", err)
	}

	// Verify key was stored
	if !pm.IsPeerVerified("test-peer") {
		t.Error("peer should be verified after adding key")
	}

	// Test invalid key size
	invalidKey := make([]byte, 16) // Wrong size
	err = pm.AddPeerKey("invalid-peer", invalidKey)
	if err == nil {
		t.Error("AddPeerKey() with invalid key should return error")
	}

	// Verify invalid peer not stored
	if pm.IsPeerVerified("invalid-peer") {
		t.Error("invalid peer should not be verified")
	}
}

func TestProtocolManager_SignAndVerifyMessage(t *testing.T) {
	pm, err := NewProtocolManager()
	if err != nil {
		t.Fatalf("NewProtocolManager() error = %v", err)
	}

	// Create test message
	msg := Message{
		Type:      MessageTypeCharacterAction,
		From:      "test-sender",
		To:        "test-receiver",
		Payload:   []byte(`{"action": "click"}`),
		Timestamp: time.Now(),
	}

	// Test signing
	signedMsg, err := pm.SignMessage(msg)
	if err != nil {
		t.Fatalf("SignMessage() error = %v", err)
	}

	if signedMsg == nil {
		t.Fatal("SignMessage() returned nil")
	}

	// Verify signature exists
	if len(signedMsg.Signature) == 0 {
		t.Error("signature is empty")
	}

	// Verify public key is included
	if len(signedMsg.PublicKey) != ed25519.PublicKeySize {
		t.Errorf("public key size = %d, want %d", len(signedMsg.PublicKey), ed25519.PublicKeySize)
	}

	// Test verification
	err = pm.VerifyMessage(signedMsg)
	if err != nil {
		t.Errorf("VerifyMessage() error = %v", err)
	}

	// Test verification with tampered message
	tamperedMsg := *signedMsg
	tamperedMsg.Message.Payload = []byte(`{"action": "tampered"}`)
	err = pm.VerifyMessage(&tamperedMsg)
	if err == nil {
		t.Error("VerifyMessage() should fail with tampered message")
	}

	// Test verification with nil message
	err = pm.VerifyMessage(nil)
	if err == nil {
		t.Error("VerifyMessage() should fail with nil message")
	}
}

func TestProtocolManager_CreateCharacterActionMessage(t *testing.T) {
	pm, err := NewProtocolManager()
	if err != nil {
		t.Fatalf("NewProtocolManager() error = %v", err)
	}

	payload := CharacterActionPayload{
		Action:        "click",
		CharacterID:   "test-character",
		InteractionID: "interaction-123",
		Position:      &Position{X: 100, Y: 200},
		Animation:     "happy",
		Response:      "Hello!",
		Stats:         map[string]float64{"happiness": 5.0},
	}

	signedMsg, err := pm.CreateCharacterActionMessage("sender", "receiver", payload)
	if err != nil {
		t.Fatalf("CreateCharacterActionMessage() error = %v", err)
	}

	// Verify message type
	if signedMsg.Message.Type != MessageTypeCharacterAction {
		t.Errorf("message type = %v, want %v", signedMsg.Message.Type, MessageTypeCharacterAction)
	}

	// Verify from/to fields
	if signedMsg.Message.From != "sender" {
		t.Errorf("from = %v, want sender", signedMsg.Message.From)
	}
	if signedMsg.Message.To != "receiver" {
		t.Errorf("to = %v, want receiver", signedMsg.Message.To)
	}

	// Verify payload can be parsed
	var parsedPayload CharacterActionPayload
	err = json.Unmarshal(signedMsg.Message.Payload, &parsedPayload)
	if err != nil {
		t.Errorf("failed to unmarshal payload: %v", err)
	}

	if parsedPayload.Action != payload.Action {
		t.Errorf("action = %v, want %v", parsedPayload.Action, payload.Action)
	}
}

func TestProtocolManager_ParseCharacterActionPayload(t *testing.T) {
	pm, err := NewProtocolManager()
	if err != nil {
		t.Fatalf("NewProtocolManager() error = %v", err)
	}

	tests := []struct {
		name      string
		payload   CharacterActionPayload
		wantError bool
	}{
		{
			name: "valid payload",
			payload: CharacterActionPayload{
				Action:        "click",
				CharacterID:   "char-1",
				InteractionID: "int-1",
			},
			wantError: false,
		},
		{
			name: "missing action",
			payload: CharacterActionPayload{
				CharacterID:   "char-1",
				InteractionID: "int-1",
			},
			wantError: true,
		},
		{
			name: "missing character ID",
			payload: CharacterActionPayload{
				Action:        "click",
				InteractionID: "int-1",
			},
			wantError: true,
		},
		{
			name: "missing interaction ID",
			payload: CharacterActionPayload{
				Action:      "click",
				CharacterID: "char-1",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payloadBytes, _ := json.Marshal(tt.payload)
			msg := Message{
				Type:    MessageTypeCharacterAction,
				Payload: payloadBytes,
			}

			parsed, err := pm.ParseCharacterActionPayload(msg)
			if (err != nil) != tt.wantError {
				t.Errorf("ParseCharacterActionPayload() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && parsed == nil {
				t.Error("ParseCharacterActionPayload() returned nil for valid payload")
			}
		})
	}

	// Test wrong message type
	wrongTypeMsg := Message{
		Type:    MessageTypeStateSync,
		Payload: []byte(`{}`),
	}
	_, err = pm.ParseCharacterActionPayload(wrongTypeMsg)
	if err == nil {
		t.Error("ParseCharacterActionPayload() should fail with wrong message type")
	}
}

func TestProtocolManager_CreateAndParseStateSyncMessage(t *testing.T) {
	pm, err := NewProtocolManager()
	if err != nil {
		t.Fatalf("NewProtocolManager() error = %v", err)
	}

	payload := StateSyncPayload{
		CharacterID:  "char-1",
		Position:     Position{X: 50, Y: 75},
		Animation:    "idle",
		CurrentState: "happy",
		GameStats:    map[string]float64{"hunger": 80.0, "happiness": 90.0},
		RomanceStats: map[string]float64{"affection": 70.0},
		LastUpdate:   time.Now(),
	}

	// Create message
	signedMsg, err := pm.CreateStateSyncMessage("sender", payload)
	if err != nil {
		t.Fatalf("CreateStateSyncMessage() error = %v", err)
	}

	// Verify message type and broadcast
	if signedMsg.Message.Type != MessageTypeStateSync {
		t.Errorf("message type = %v, want %v", signedMsg.Message.Type, MessageTypeStateSync)
	}
	if signedMsg.Message.To != "" {
		t.Errorf("to = %v, want empty (broadcast)", signedMsg.Message.To)
	}

	// Parse message back
	parsed, err := pm.ParseStateSyncPayload(signedMsg.Message)
	if err != nil {
		t.Fatalf("ParseStateSyncPayload() error = %v", err)
	}

	// Verify parsed data
	if parsed.CharacterID != payload.CharacterID {
		t.Errorf("characterId = %v, want %v", parsed.CharacterID, payload.CharacterID)
	}
	if parsed.CurrentState != payload.CurrentState {
		t.Errorf("currentState = %v, want %v", parsed.CurrentState, payload.CurrentState)
	}
	if parsed.Checksum == "" {
		t.Error("checksum should not be empty")
	}
}

func TestProtocolManager_CreateSecureDiscoveryMessage(t *testing.T) {
	pm, err := NewProtocolManager()
	if err != nil {
		t.Fatalf("NewProtocolManager() error = %v", err)
	}

	basicPayload := DiscoveryPayload{
		NetworkID: "test-network",
		PeerID:    "test-peer",
		TCPPort:   8080,
	}

	capabilities := []string{"bot", "game", "romance"}

	signedMsg, err := pm.CreateSecureDiscoveryMessage(basicPayload, capabilities)
	if err != nil {
		t.Fatalf("CreateSecureDiscoveryMessage() error = %v", err)
	}

	// Verify message structure
	if signedMsg.Message.Type != MessageTypeDiscovery {
		t.Errorf("message type = %v, want %v", signedMsg.Message.Type, MessageTypeDiscovery)
	}

	// Parse extended payload
	parsed, err := pm.ParseExtendedDiscoveryPayload(signedMsg.Message)
	if err != nil {
		t.Fatalf("ParseExtendedDiscoveryPayload() error = %v", err)
	}

	// Verify extended fields
	if parsed.NetworkID != basicPayload.NetworkID {
		t.Errorf("networkId = %v, want %v", parsed.NetworkID, basicPayload.NetworkID)
	}
	if len(parsed.PublicKey) != ed25519.PublicKeySize {
		t.Errorf("public key size = %d, want %d", len(parsed.PublicKey), ed25519.PublicKeySize)
	}
	if len(parsed.Capabilities) != len(capabilities) {
		t.Errorf("capabilities length = %d, want %d", len(parsed.Capabilities), len(capabilities))
	}
	if parsed.Version == "" {
		t.Error("version should not be empty")
	}
}

func TestProtocolManager_ValidateMessageAge(t *testing.T) {
	pm, err := NewProtocolManager()
	if err != nil {
		t.Fatalf("NewProtocolManager() error = %v", err)
	}

	tests := []struct {
		name      string
		timestamp time.Time
		maxAge    time.Duration
		wantError bool
	}{
		{
			name:      "recent message",
			timestamp: time.Now().Add(-30 * time.Second),
			maxAge:    time.Minute,
			wantError: false,
		},
		{
			name:      "old message",
			timestamp: time.Now().Add(-2 * time.Minute),
			maxAge:    time.Minute,
			wantError: true,
		},
		{
			name:      "future message",
			timestamp: time.Now().Add(2 * time.Minute),
			maxAge:    time.Minute,
			wantError: true,
		},
		{
			name:      "slightly future message (allowed)",
			timestamp: time.Now().Add(30 * time.Second),
			maxAge:    time.Minute,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := Message{
				Timestamp: tt.timestamp,
			}

			err := pm.ValidateMessageAge(msg, tt.maxAge)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateMessageAge() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestProtocolManager_GetVerifiedPeers(t *testing.T) {
	pm, err := NewProtocolManager()
	if err != nil {
		t.Fatalf("NewProtocolManager() error = %v", err)
	}

	// Initially no verified peers
	peers := pm.GetVerifiedPeers()
	if len(peers) != 0 {
		t.Errorf("initial verified peers = %d, want 0", len(peers))
	}

	// Add some peer keys
	key1 := make([]byte, ed25519.PublicKeySize)
	key2 := make([]byte, ed25519.PublicKeySize)

	pm.AddPeerKey("peer1", key1)
	pm.AddPeerKey("peer2", key2)

	// Check verified peers
	peers = pm.GetVerifiedPeers()
	if len(peers) != 2 {
		t.Errorf("verified peers count = %d, want 2", len(peers))
	}

	// Verify individual peers
	if !pm.IsPeerVerified("peer1") {
		t.Error("peer1 should be verified")
	}
	if !pm.IsPeerVerified("peer2") {
		t.Error("peer2 should be verified")
	}
	if pm.IsPeerVerified("unknown-peer") {
		t.Error("unknown peer should not be verified")
	}
}

func TestProtocolManager_ChecksumGeneration(t *testing.T) {
	pm, err := NewProtocolManager()
	if err != nil {
		t.Fatalf("NewProtocolManager() error = %v", err)
	}

	payload := StateSyncPayload{
		CharacterID:  "test-char",
		CurrentState: "idle",
		Position:     Position{X: 10, Y: 20},
	}

	// Generate checksum
	checksum1 := pm.generateChecksum(payload)
	if checksum1 == "" {
		t.Error("checksum should not be empty")
	}

	// Same payload should generate same checksum
	checksum2 := pm.generateChecksum(payload)
	if checksum1 != checksum2 {
		t.Errorf("checksums differ: %s != %s", checksum1, checksum2)
	}

	// Different payload should generate different checksum
	payload.CurrentState = "happy"
	checksum3 := pm.generateChecksum(payload)
	if checksum1 == checksum3 {
		t.Error("different payloads should generate different checksums")
	}
}

func TestSignedMessage_Serialization(t *testing.T) {
	pm, err := NewProtocolManager()
	if err != nil {
		t.Fatalf("NewProtocolManager() error = %v", err)
	}

	msg := Message{
		Type:      MessageTypeCharacterAction,
		From:      "test-sender",
		To:        "test-receiver",
		Payload:   []byte(`{"test": "data"}`),
		Timestamp: time.Now(),
	}

	signedMsg, err := pm.SignMessage(msg)
	if err != nil {
		t.Fatalf("SignMessage() error = %v", err)
	}

	// Test JSON serialization
	data, err := json.Marshal(signedMsg)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// Test JSON deserialization
	var decoded SignedMessage
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	// Verify fields
	if decoded.Message.Type != signedMsg.Message.Type {
		t.Errorf("Type = %v, want %v", decoded.Message.Type, signedMsg.Message.Type)
	}
	if decoded.Message.From != signedMsg.Message.From {
		t.Errorf("From = %v, want %v", decoded.Message.From, signedMsg.Message.From)
	}
	if len(decoded.Signature) != len(signedMsg.Signature) {
		t.Errorf("Signature length = %d, want %d", len(decoded.Signature), len(signedMsg.Signature))
	}
}

func TestProtocolManager_CreatePeerListMessage(t *testing.T) {
	pm, err := NewProtocolManager()
	if err != nil {
		t.Fatalf("NewProtocolManager() error = %v", err)
	}

	peers := []SecurePeer{
		{
			ID:        "peer1",
			AddrStr:   "127.0.0.1:8080",
			PublicKey: make([]byte, ed25519.PublicKeySize),
			LastSeen:  time.Now(),
			Verified:  true,
		},
		{
			ID:        "peer2",
			AddrStr:   "127.0.0.1:8081",
			PublicKey: make([]byte, ed25519.PublicKeySize),
			LastSeen:  time.Now(),
			Verified:  false,
		},
	}

	signedMsg, err := pm.CreatePeerListMessage("sender", peers)
	if err != nil {
		t.Fatalf("CreatePeerListMessage() error = %v", err)
	}

	// Verify message type and broadcast
	if signedMsg.Message.Type != MessageTypePeerList {
		t.Errorf("message type = %v, want %v", signedMsg.Message.Type, MessageTypePeerList)
	}
	if signedMsg.Message.To != "" {
		t.Errorf("to = %v, want empty (broadcast)", signedMsg.Message.To)
	}

	// Parse payload
	var payload PeerListPayload
	err = json.Unmarshal(signedMsg.Message.Payload, &payload)
	if err != nil {
		t.Fatalf("failed to unmarshal peer list payload: %v", err)
	}

	// Verify peer data
	if len(payload.Peers) != 2 {
		t.Errorf("peers count = %d, want 2", len(payload.Peers))
	}
	if payload.Peers[0].ID != "peer1" {
		t.Errorf("first peer ID = %v, want peer1", payload.Peers[0].ID)
	}
}

// Benchmark tests for performance validation
func BenchmarkProtocolManager_SignMessage(b *testing.B) {
	pm, err := NewProtocolManager()
	if err != nil {
		b.Fatalf("NewProtocolManager() error = %v", err)
	}

	msg := Message{
		Type:      MessageTypeCharacterAction,
		From:      "sender",
		To:        "receiver",
		Payload:   []byte(`{"action": "click", "characterId": "test"}`),
		Timestamp: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pm.SignMessage(msg)
		if err != nil {
			b.Fatalf("SignMessage() error = %v", err)
		}
	}
}

func BenchmarkProtocolManager_VerifyMessage(b *testing.B) {
	pm, err := NewProtocolManager()
	if err != nil {
		b.Fatalf("NewProtocolManager() error = %v", err)
	}

	msg := Message{
		Type:      MessageTypeCharacterAction,
		From:      "sender",
		To:        "receiver",
		Payload:   []byte(`{"action": "click", "characterId": "test"}`),
		Timestamp: time.Now(),
	}

	signedMsg, err := pm.SignMessage(msg)
	if err != nil {
		b.Fatalf("SignMessage() error = %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := pm.VerifyMessage(signedMsg)
		if err != nil {
			b.Fatalf("VerifyMessage() error = %v", err)
		}
	}
}
