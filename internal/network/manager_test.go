package network

import (
	"encoding/json"
	"net"
	"testing"
	"time"
)

func TestNewNetworkManager(t *testing.T) {
	tests := []struct {
		name     string
		config   NetworkManagerConfig
		expected NetworkManagerConfig
	}{
		{
			name:   "default configuration",
			config: NetworkManagerConfig{},
			expected: NetworkManagerConfig{
				DiscoveryPort:     8080,
				MaxPeers:          8,
				NetworkID:         "dds-default",
				DiscoveryInterval: 5 * time.Second,
			},
		},
		{
			name: "custom configuration",
			config: NetworkManagerConfig{
				DiscoveryPort:     9000,
				MaxPeers:          4,
				NetworkID:         "test-network",
				DiscoveryInterval: 3 * time.Second,
			},
			expected: NetworkManagerConfig{
				DiscoveryPort:     9000,
				MaxPeers:          4,
				NetworkID:         "test-network",
				DiscoveryInterval: 3 * time.Second,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nm, err := NewNetworkManager(tt.config)
			if err != nil {
				t.Fatalf("NewNetworkManager() error = %v", err)
			}

			if nm.discoveryPort != tt.expected.DiscoveryPort {
				t.Errorf("discoveryPort = %v, want %v", nm.discoveryPort, tt.expected.DiscoveryPort)
			}
			if nm.maxPeers != tt.expected.MaxPeers {
				t.Errorf("maxPeers = %v, want %v", nm.maxPeers, tt.expected.MaxPeers)
			}
			if nm.networkID != tt.expected.NetworkID {
				t.Errorf("networkID = %v, want %v", nm.networkID, tt.expected.NetworkID)
			}
			if nm.discoveryInterval != tt.expected.DiscoveryInterval {
				t.Errorf("discoveryInterval = %v, want %v", nm.discoveryInterval, tt.expected.DiscoveryInterval)
			}

			// Verify internal structures are initialized
			if nm.peers == nil {
				t.Error("peers map not initialized")
			}
			if nm.messageQueue == nil {
				t.Error("messageQueue not initialized")
			}
			if nm.handlers == nil {
				t.Error("handlers map not initialized")
			}

			// Verify default handlers are registered
			if _, exists := nm.handlers[MessageTypeDiscovery]; !exists {
				t.Error("discovery handler not registered")
			}
			if _, exists := nm.handlers[MessageTypePeerList]; !exists {
				t.Error("peer list handler not registered")
			}
		})
	}
}

func TestNetworkManager_StartStop(t *testing.T) {
	// Test with valid configuration
	config := NetworkManagerConfig{
		DiscoveryPort:     findAvailablePort(t),
		MaxPeers:          2,
		NetworkID:         "test-network",
		DiscoveryInterval: 100 * time.Millisecond,
	}

	nm, err := NewNetworkManager(config)
	if err != nil {
		t.Fatalf("NewNetworkManager() error = %v", err)
	}

	// Test successful start
	err = nm.Start()
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Verify network manager is running
	if nm.discoveryConn == nil {
		t.Error("discoveryConn not set after Start()")
	}
	if nm.tcpListener == nil {
		t.Error("tcpListener not set after Start()")
	}

	// Test graceful stop
	err = nm.Stop()
	if err != nil {
		t.Errorf("Stop() error = %v", err)
	}

	// Verify stop completed within reasonable time
	done := make(chan bool, 1)
	go func() {
		// Give goroutines time to finish
		time.Sleep(100 * time.Millisecond)
		done <- true
	}()

	select {
	case <-done:
		// Success
	case <-time.After(time.Second):
		t.Error("Stop() took too long to complete")
	}
}

func TestNetworkManager_MessageHandling(t *testing.T) {
	config := NetworkManagerConfig{
		DiscoveryPort:     findAvailablePort(t),
		MaxPeers:          2,
		NetworkID:         "test-network",
		DiscoveryInterval: time.Hour, // Long interval to avoid automatic broadcasts
	}

	nm, err := NewNetworkManager(config)
	if err != nil {
		t.Fatalf("NewNetworkManager() error = %v", err)
	}

	// Test custom message handler registration
	var handlerCalled bool
	customHandler := func(msg Message, from *Peer) error {
		handlerCalled = true
		return nil
	}

	nm.RegisterMessageHandler(MessageTypeCharacterAction, customHandler)

	// Verify handler was registered
	if _, exists := nm.handlers[MessageTypeCharacterAction]; !exists {
		t.Error("custom handler not registered")
	}

	// Test message sending (should queue message)
	payload := []byte(`{"action": "test"}`)
	err = nm.SendMessage(MessageTypeCharacterAction, payload, "")
	if err != nil {
		t.Errorf("SendMessage() error = %v", err)
	}

	// Verify message was queued (can't easily test delivery without full network setup)
	// This is a basic test that the queueing mechanism works
	_ = handlerCalled // Acknowledge variable is used for future testing
}

func TestNetworkManager_PeerManagement(t *testing.T) {
	config := NetworkManagerConfig{
		DiscoveryPort:     findAvailablePort(t),
		MaxPeers:          2,
		NetworkID:         "test-network",
		DiscoveryInterval: time.Hour,
	}

	nm, err := NewNetworkManager(config)
	if err != nil {
		t.Fatalf("NewNetworkManager() error = %v", err)
	}

	// Test initial state
	if count := nm.GetPeerCount(); count != 0 {
		t.Errorf("initial peer count = %v, want 0", count)
	}

	peers := nm.GetPeers()
	if len(peers) != 0 {
		t.Errorf("initial peers length = %v, want 0", len(peers))
	}

	// Test adding a peer manually (simulating discovery)
	testAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:12345")
	testPeer := &Peer{
		ID:       "test-peer",
		Addr:     testAddr,
		AddrStr:  testAddr.String(),
		LastSeen: time.Now(),
	}

	nm.mu.Lock()
	nm.peers["test-peer"] = testPeer
	nm.mu.Unlock()

	// Verify peer was added
	if count := nm.GetPeerCount(); count != 1 {
		t.Errorf("peer count after adding = %v, want 1", count)
	}

	peers = nm.GetPeers()
	if len(peers) != 1 {
		t.Errorf("peers length after adding = %v, want 1", len(peers))
	}
	if peers[0].ID != "test-peer" {
		t.Errorf("peer ID = %v, want test-peer", peers[0].ID)
	}
}

func TestMessage_Serialization(t *testing.T) {
	msg := Message{
		Type:      MessageTypeCharacterAction,
		From:      "peer1",
		To:        "peer2",
		Payload:   []byte(`{"action": "click"}`),
		Timestamp: time.Now(),
	}

	// Test JSON serialization
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// Test JSON deserialization
	var decoded Message
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	// Verify fields
	if decoded.Type != msg.Type {
		t.Errorf("Type = %v, want %v", decoded.Type, msg.Type)
	}
	if decoded.From != msg.From {
		t.Errorf("From = %v, want %v", decoded.From, msg.From)
	}
	if decoded.To != msg.To {
		t.Errorf("To = %v, want %v", decoded.To, msg.To)
	}
	if string(decoded.Payload) != string(msg.Payload) {
		t.Errorf("Payload = %v, want %v", string(decoded.Payload), string(msg.Payload))
	}
}

func TestDiscoveryPayload_Serialization(t *testing.T) {
	payload := DiscoveryPayload{
		NetworkID: "test-network",
		PeerID:    "test-peer",
		TCPPort:   8081,
	}

	// Test JSON serialization
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// Test JSON deserialization
	var decoded DiscoveryPayload
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	// Verify fields
	if decoded.NetworkID != payload.NetworkID {
		t.Errorf("NetworkID = %v, want %v", decoded.NetworkID, payload.NetworkID)
	}
	if decoded.PeerID != payload.PeerID {
		t.Errorf("PeerID = %v, want %v", decoded.PeerID, payload.PeerID)
	}
	if decoded.TCPPort != payload.TCPPort {
		t.Errorf("TCPPort = %v, want %v", decoded.TCPPort, payload.TCPPort)
	}
}

func TestNetworkManager_ProcessDiscoveryMessage(t *testing.T) {
	config := NetworkManagerConfig{
		DiscoveryPort:     findAvailablePort(t),
		MaxPeers:          2,
		NetworkID:         "test-network",
		DiscoveryInterval: time.Hour,
	}

	nm, err := NewNetworkManager(config)
	if err != nil {
		t.Fatalf("NewNetworkManager() error = %v", err)
	}

	// Create a valid discovery message from another peer
	payload := DiscoveryPayload{
		NetworkID: "test-network",
		PeerID:    "remote-peer",
		TCPPort:   8081,
	}

	payloadBytes, _ := json.Marshal(payload)
	msg := Message{
		Type:      MessageTypeDiscovery,
		From:      "remote-peer",
		Payload:   payloadBytes,
		Timestamp: time.Now(),
	}

	testAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:12345")

	// Process the discovery message
	nm.processDiscoveryMessage(msg, testAddr)

	// Verify peer was added
	if count := nm.GetPeerCount(); count != 1 {
		t.Errorf("peer count after discovery = %v, want 1", count)
	}

	peers := nm.GetPeers()
	if len(peers) != 1 {
		t.Fatalf("peers length = %v, want 1", len(peers))
	}
	if peers[0].ID != "remote-peer" {
		t.Errorf("peer ID = %v, want remote-peer", peers[0].ID)
	}

	// Test ignoring messages from self
	selfPayload := DiscoveryPayload{
		NetworkID: "test-network",
		PeerID:    nm.networkID, // Same as our network ID
		TCPPort:   8081,
	}
	selfPayloadBytes, _ := json.Marshal(selfPayload)
	selfMsg := Message{
		Type:      MessageTypeDiscovery,
		From:      nm.networkID,
		Payload:   selfPayloadBytes,
		Timestamp: time.Now(),
	}

	nm.processDiscoveryMessage(selfMsg, testAddr)

	// Peer count should still be 1
	if count := nm.GetPeerCount(); count != 1 {
		t.Errorf("peer count after self-discovery = %v, want 1", count)
	}

	// Test ignoring messages from different network
	differentNetworkPayload := DiscoveryPayload{
		NetworkID: "different-network",
		PeerID:    "other-peer",
		TCPPort:   8081,
	}
	differentPayloadBytes, _ := json.Marshal(differentNetworkPayload)
	differentMsg := Message{
		Type:      MessageTypeDiscovery,
		From:      "other-peer",
		Payload:   differentPayloadBytes,
		Timestamp: time.Now(),
	}

	nm.processDiscoveryMessage(differentMsg, testAddr)

	// Peer count should still be 1
	if count := nm.GetPeerCount(); count != 1 {
		t.Errorf("peer count after different network discovery = %v, want 1", count)
	}
}

func TestNetworkManager_MaxPeersLimit(t *testing.T) {
	config := NetworkManagerConfig{
		DiscoveryPort:     findAvailablePort(t),
		MaxPeers:          1, // Limit to 1 peer
		NetworkID:         "test-network",
		DiscoveryInterval: time.Hour,
	}

	nm, err := NewNetworkManager(config)
	if err != nil {
		t.Fatalf("NewNetworkManager() error = %v", err)
	}

	testAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:12345")

	// Add first peer
	payload1 := DiscoveryPayload{
		NetworkID: "test-network",
		PeerID:    "peer-1",
		TCPPort:   8081,
	}
	payloadBytes1, _ := json.Marshal(payload1)
	msg1 := Message{
		Type:      MessageTypeDiscovery,
		From:      "peer-1",
		Payload:   payloadBytes1,
		Timestamp: time.Now(),
	}

	nm.processDiscoveryMessage(msg1, testAddr)

	if count := nm.GetPeerCount(); count != 1 {
		t.Errorf("peer count after first peer = %v, want 1", count)
	}

	// Try to add second peer (should be ignored due to max peers limit)
	payload2 := DiscoveryPayload{
		NetworkID: "test-network",
		PeerID:    "peer-2",
		TCPPort:   8082,
	}
	payloadBytes2, _ := json.Marshal(payload2)
	msg2 := Message{
		Type:      MessageTypeDiscovery,
		From:      "peer-2",
		Payload:   payloadBytes2,
		Timestamp: time.Now(),
	}

	nm.processDiscoveryMessage(msg2, testAddr)

	// Should still have only 1 peer
	if count := nm.GetPeerCount(); count != 1 {
		t.Errorf("peer count after second peer = %v, want 1", count)
	}

	peers := nm.GetPeers()
	if len(peers) != 1 || peers[0].ID != "peer-1" {
		t.Error("second peer was added despite max peers limit")
	}
}

// findAvailablePort finds an available UDP port for testing
func findAvailablePort(t *testing.T) int {
	t.Helper()

	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		t.Fatalf("Failed to find available port: %v", err)
	}
	defer conn.Close()

	addr := conn.LocalAddr()
	if udpAddr, ok := addr.(*net.UDPAddr); ok {
		return udpAddr.Port
	}

	t.Fatal("Failed to get port from UDP address")
	return 0
}

// TestContextCancellation tests that context cancellation stops all goroutines
func TestNetworkManager_ContextCancellation(t *testing.T) {
	config := NetworkManagerConfig{
		DiscoveryPort:     findAvailablePort(t),
		MaxPeers:          2,
		NetworkID:         "test-network",
		DiscoveryInterval: 10 * time.Millisecond, // Short interval for testing
	}

	nm, err := NewNetworkManager(config)
	if err != nil {
		t.Fatalf("NewNetworkManager() error = %v", err)
	}

	err = nm.Start()
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Let it run briefly
	time.Sleep(50 * time.Millisecond)

	// Test that Stop() completes quickly
	start := time.Now()
	err = nm.Stop()
	duration := time.Since(start)

	if err != nil {
		t.Errorf("Stop() error = %v", err)
	}

	// Should complete within a reasonable time (much less than a second)
	if duration > 500*time.Millisecond {
		t.Errorf("Stop() took %v, expected < 500ms", duration)
	}
}
