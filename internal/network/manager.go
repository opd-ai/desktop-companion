package network

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

// NetworkManager handles peer discovery and communication for multiplayer functionality.
// Uses Go standard library for networking following the project's "library-first" philosophy.
type NetworkManager struct {
	mu sync.RWMutex

	// Network configuration
	discoveryPort int
	maxPeers      int
	networkID     string

	// Connection management - using interface types for testability
	discoveryConn net.PacketConn // UDP for peer discovery
	tcpListener   net.Listener   // TCP for reliable messaging

	// Peer tracking
	peers     map[string]*Peer
	localAddr net.Addr

	// Message handling
	messageQueue chan Message
	handlers     map[MessageType]MessageHandler

	// Lifecycle management
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Discovery state
	discoveryInterval time.Duration
}

// Peer represents a connected peer in the network
type Peer struct {
	ID       string    `json:"id"`
	Addr     net.Addr  `json:"-"`    // Don't serialize net.Addr
	AddrStr  string    `json:"addr"` // Serializable address
	LastSeen time.Time `json:"lastSeen"`
	Conn     net.Conn  `json:"-"` // TCP connection, nil if not connected
}

// MessageType defines the type of network message
type MessageType string

const (
	MessageTypeDiscovery       MessageType = "discovery"
	MessageTypeCharacterAction MessageType = "character_action"
	MessageTypeStateSync       MessageType = "state_sync"
	MessageTypePeerList        MessageType = "peer_list"
	// Battle system message types (Phase 2)
	MessageTypeBattleInvite    MessageType = "battle_invite"
	MessageTypeBattleAction    MessageType = "battle_action"
	MessageTypeBattleResult    MessageType = "battle_result"
	MessageTypeBattleEnd       MessageType = "battle_end"
)

// Message represents a network message between peers
type Message struct {
	Type      MessageType `json:"type"`
	From      string      `json:"from"`
	To        string      `json:"to,omitempty"` // Empty for broadcast
	Payload   []byte      `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
}

// MessageHandler processes incoming messages of a specific type
type MessageHandler func(msg Message, from *Peer) error

// DiscoveryPayload is sent during peer discovery
type DiscoveryPayload struct {
	NetworkID string `json:"networkId"`
	PeerID    string `json:"peerId"`
	TCPPort   int    `json:"tcpPort"`
}

// NetworkManagerConfig contains configuration for the NetworkManager
type NetworkManagerConfig struct {
	DiscoveryPort     int           `json:"discoveryPort"`
	MaxPeers          int           `json:"maxPeers"`
	NetworkID         string        `json:"networkId"`
	DiscoveryInterval time.Duration `json:"discoveryInterval"`
}

// NewNetworkManager creates a new NetworkManager with the given configuration.
// Uses standard library networking interfaces for testability and follows
// the project's principle of minimal external dependencies.
func NewNetworkManager(config NetworkManagerConfig) (*NetworkManager, error) {
	if config.DiscoveryPort <= 0 {
		config.DiscoveryPort = 8080
	}
	if config.MaxPeers <= 0 {
		config.MaxPeers = 8
	}
	if config.NetworkID == "" {
		config.NetworkID = "dds-default"
	}
	if config.DiscoveryInterval <= 0 {
		config.DiscoveryInterval = 5 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())

	nm := &NetworkManager{
		discoveryPort:     config.DiscoveryPort,
		maxPeers:          config.MaxPeers,
		networkID:         config.NetworkID,
		peers:             make(map[string]*Peer),
		messageQueue:      make(chan Message, 100), // Buffered channel for async processing
		handlers:          make(map[MessageType]MessageHandler),
		ctx:               ctx,
		cancel:            cancel,
		discoveryInterval: config.DiscoveryInterval,
	}

	// Register default message handlers
	nm.handlers[MessageTypeDiscovery] = nm.handleDiscoveryMessage
	nm.handlers[MessageTypePeerList] = nm.handlePeerListMessage

	return nm, nil
}

// Start initializes the network manager and begins peer discovery.
// Returns error if network initialization fails.
func (nm *NetworkManager) Start() error {
	// Start UDP discovery listener
	discoveryAddr := fmt.Sprintf(":%d", nm.discoveryPort)
	conn, err := net.ListenPacket("udp", discoveryAddr)
	if err != nil {
		return fmt.Errorf("failed to start discovery listener: %w", err)
	}
	nm.discoveryConn = conn

	// Start TCP listener for peer connections
	tcpListener, err := net.Listen("tcp", ":0") // Let OS assign port
	if err != nil {
		nm.discoveryConn.Close()
		return fmt.Errorf("failed to start TCP listener: %w", err)
	}
	nm.tcpListener = tcpListener

	// Store local address for discovery messages
	nm.localAddr = tcpListener.Addr()

	// Start background goroutines
	nm.wg.Add(3)
	go nm.discoveryListener()
	go nm.tcpConnectionHandler()
	go nm.messageProcessor()

	// Start periodic discovery broadcasts
	nm.wg.Add(1)
	go nm.discoveryBroadcaster()

	return nil
}

// Stop gracefully shuts down the network manager
func (nm *NetworkManager) Stop() error {
	nm.cancel() // Signal all goroutines to stop

	// Close network connections
	if nm.discoveryConn != nil {
		nm.discoveryConn.Close()
	}
	if nm.tcpListener != nil {
		nm.tcpListener.Close()
	}

	// Close peer connections
	nm.mu.Lock()
	for _, peer := range nm.peers {
		if peer.Conn != nil {
			peer.Conn.Close()
		}
	}
	nm.mu.Unlock()

	// Wait for goroutines to finish
	nm.wg.Wait()

	return nil
}

// GetPeers returns a copy of the current peer list
func (nm *NetworkManager) GetPeers() []Peer {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	peers := make([]Peer, 0, len(nm.peers))
	for _, peer := range nm.peers {
		peers = append(peers, *peer)
	}
	return peers
}

// GetPeerCount returns the number of connected peers
func (nm *NetworkManager) GetPeerCount() int {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	return len(nm.peers)
}

// GetNetworkID returns the local network identifier
func (nm *NetworkManager) GetNetworkID() string {
	return nm.networkID
}

// RegisterMessageHandler registers a handler for a specific message type
func (nm *NetworkManager) RegisterMessageHandler(msgType MessageType, handler MessageHandler) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	nm.handlers[msgType] = handler
}

// SendMessage sends a message to a specific peer or broadcasts to all peers
func (nm *NetworkManager) SendMessage(msgType MessageType, payload []byte, targetPeerID string) error {
	message := Message{
		Type:      msgType,
		From:      nm.networkID,
		To:        targetPeerID,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	select {
	case nm.messageQueue <- message:
		return nil
	case <-nm.ctx.Done():
		return fmt.Errorf("network manager stopped")
	default:
		return fmt.Errorf("message queue full")
	}
}

// discoveryListener handles incoming UDP discovery messages
func (nm *NetworkManager) discoveryListener() {
	defer nm.wg.Done()

	buffer := make([]byte, 1024)
	for {
		select {
		case <-nm.ctx.Done():
			return
		default:
		}

		// Set read timeout to allow periodic context checking
		nm.discoveryConn.SetReadDeadline(time.Now().Add(time.Second))
		n, addr, err := nm.discoveryConn.ReadFrom(buffer)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue // Timeout is expected, check context and continue
			}
			continue // Other errors are logged but don't stop the listener
		}

		var msg Message
		if err := json.Unmarshal(buffer[:n], &msg); err != nil {
			continue // Invalid JSON, skip
		}

		// Process discovery message
		if msg.Type == MessageTypeDiscovery {
			nm.processDiscoveryMessage(msg, addr)
		}
	}
}

// processDiscoveryMessage handles a discovery message from a peer
func (nm *NetworkManager) processDiscoveryMessage(msg Message, from net.Addr) {
	var payload DiscoveryPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return
	}

	// Ignore messages from ourselves
	if payload.PeerID == nm.networkID {
		return
	}

	// Ignore messages from different networks
	if payload.NetworkID != nm.networkID {
		return
	}

	// Check if we're at max peers
	nm.mu.RLock()
	atMaxPeers := len(nm.peers) >= nm.maxPeers
	nm.mu.RUnlock()

	if atMaxPeers {
		return
	}

	// Add or update peer
	nm.mu.Lock()
	peer, exists := nm.peers[payload.PeerID]
	if !exists {
		peer = &Peer{
			ID:      payload.PeerID,
			Addr:    from,
			AddrStr: from.String(),
		}
		nm.peers[payload.PeerID] = peer
	}
	peer.LastSeen = time.Now()
	nm.mu.Unlock()

	// Attempt TCP connection if not already connected
	if peer.Conn == nil {
		go nm.connectToPeer(peer, payload.TCPPort)
	}
}

// connectToPeer establishes a TCP connection to a discovered peer
func (nm *NetworkManager) connectToPeer(peer *Peer, tcpPort int) {
	// Extract host from UDP address and connect to TCP port
	host, _, err := net.SplitHostPort(peer.AddrStr)
	if err != nil {
		return
	}

	tcpAddr := net.JoinHostPort(host, fmt.Sprintf("%d", tcpPort))
	conn, err := net.DialTimeout("tcp", tcpAddr, 5*time.Second)
	if err != nil {
		return
	}

	nm.mu.Lock()
	peer.Conn = conn
	nm.mu.Unlock()

	// Handle the connection in a separate goroutine
	go nm.handlePeerConnection(peer)
}

// handlePeerConnection manages a TCP connection to a peer
func (nm *NetworkManager) handlePeerConnection(peer *Peer) {
	defer func() {
		if peer.Conn != nil {
			peer.Conn.Close()
			nm.mu.Lock()
			peer.Conn = nil
			nm.mu.Unlock()
		}
	}()

	decoder := json.NewDecoder(peer.Conn)
	for {
		select {
		case <-nm.ctx.Done():
			return
		default:
		}

		var msg Message
		if err := decoder.Decode(&msg); err != nil {
			return // Connection error, cleanup and exit
		}

		// Forward message to handler
		if handler, exists := nm.handlers[msg.Type]; exists {
			go handler(msg, peer) // Handle in separate goroutine to avoid blocking
		}
	}
}

// tcpConnectionHandler accepts incoming TCP connections from peers
func (nm *NetworkManager) tcpConnectionHandler() {
	defer nm.wg.Done()

	for {
		select {
		case <-nm.ctx.Done():
			return
		default:
		}

		conn, err := nm.tcpListener.Accept()
		if err != nil {
			select {
			case <-nm.ctx.Done():
				return // Expected error during shutdown
			default:
				continue // Other errors, keep listening
			}
		}

		// Handle incoming connection
		go nm.handleIncomingConnection(conn)
	}
}

// handleIncomingConnection processes an incoming TCP connection
func (nm *NetworkManager) handleIncomingConnection(conn net.Conn) {
	defer conn.Close()

	// Set a timeout for initial handshake
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	decoder := json.NewDecoder(conn)
	var msg Message
	if err := decoder.Decode(&msg); err != nil {
		return
	}

	// First message should identify the peer
	nm.mu.RLock()
	peer, exists := nm.peers[msg.From]
	nm.mu.RUnlock()

	if !exists {
		return // Unknown peer
	}

	// Update peer connection
	nm.mu.Lock()
	peer.Conn = conn
	nm.mu.Unlock()

	// Continue handling messages on this connection
	nm.handlePeerConnection(peer)
}

// discoveryBroadcaster periodically sends discovery messages
func (nm *NetworkManager) discoveryBroadcaster() {
	defer nm.wg.Done()

	ticker := time.NewTicker(nm.discoveryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-nm.ctx.Done():
			return
		case <-ticker.C:
			nm.sendDiscoveryBroadcast()
		}
	}
}

// sendDiscoveryBroadcast sends a UDP broadcast discovery message
func (nm *NetworkManager) sendDiscoveryBroadcast() {
	// Get TCP port from listener
	tcpPort := 0
	if nm.tcpListener != nil {
		if addr, ok := nm.tcpListener.Addr().(*net.TCPAddr); ok {
			tcpPort = addr.Port
		}
	}

	payload := DiscoveryPayload{
		NetworkID: nm.networkID,
		PeerID:    nm.networkID,
		TCPPort:   tcpPort,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return
	}

	msg := Message{
		Type:      MessageTypeDiscovery,
		From:      nm.networkID,
		Payload:   payloadBytes,
		Timestamp: time.Now(),
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return
	}

	// Broadcast to local network
	broadcastAddr := net.JoinHostPort("255.255.255.255", fmt.Sprintf("%d", nm.discoveryPort))
	addr, err := net.ResolveUDPAddr("udp", broadcastAddr)
	if err != nil {
		return
	}

	nm.discoveryConn.WriteTo(msgBytes, addr)
}

// messageProcessor handles outgoing messages from the queue
func (nm *NetworkManager) messageProcessor() {
	defer nm.wg.Done()

	for {
		select {
		case <-nm.ctx.Done():
			return
		case msg := <-nm.messageQueue:
			nm.processOutgoingMessage(msg)
		}
	}
}

// processOutgoingMessage sends a message to the appropriate peer(s)
func (nm *NetworkManager) processOutgoingMessage(msg Message) {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	if msg.To == "" {
		// Broadcast to all connected peers
		for _, peer := range nm.peers {
			if peer.Conn != nil {
				nm.sendMessageToPeer(msg, peer)
			}
		}
	} else {
		// Send to specific peer
		if peer, exists := nm.peers[msg.To]; exists && peer.Conn != nil {
			nm.sendMessageToPeer(msg, peer)
		}
	}
}

// sendMessageToPeer sends a message to a specific peer over TCP
func (nm *NetworkManager) sendMessageToPeer(msg Message, peer *Peer) {
	encoder := json.NewEncoder(peer.Conn)
	encoder.Encode(msg) // Ignore errors for now, connection will be cleaned up by handler
}

// Default message handlers

// handleDiscoveryMessage processes discovery messages (should not be called for UDP discovery)
func (nm *NetworkManager) handleDiscoveryMessage(msg Message, from *Peer) error {
	// Discovery is handled via UDP, this is for TCP-based discovery messages
	return nil
}

// handlePeerListMessage processes peer list synchronization messages
func (nm *NetworkManager) handlePeerListMessage(msg Message, from *Peer) error {
	var peerList []Peer
	if err := json.Unmarshal(msg.Payload, &peerList); err != nil {
		return fmt.Errorf("failed to unmarshal peer list: %w", err)
	}

	// Update our peer list with information from remote peer
	nm.mu.Lock()
	for _, remotePeer := range peerList {
		if _, exists := nm.peers[remotePeer.ID]; !exists && len(nm.peers) < nm.maxPeers {
			nm.peers[remotePeer.ID] = &remotePeer
		}
	}
	nm.mu.Unlock()

	return nil
}
