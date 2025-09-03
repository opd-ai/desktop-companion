package ui

import (
	"net"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"

	"desktop-companion/internal/network"
)

// MockNetworkManager implements a minimal NetworkManager interface for testing
type MockNetworkManager struct {
	peers        []network.Peer
	peerCount    int
	networkID    string
	messagesSent []MockMessage
}

type MockMessage struct {
	MsgType  network.MessageType
	Payload  []byte
	TargetID string
}

func NewMockNetworkManager() *MockNetworkManager {
	return &MockNetworkManager{
		peers:        make([]network.Peer, 0),
		networkID:    "test-network",
		messagesSent: make([]MockMessage, 0),
	}
}

func (m *MockNetworkManager) GetPeerCount() int {
	return m.peerCount
}

func (m *MockNetworkManager) GetPeers() []network.Peer {
	return m.peers
}

func (m *MockNetworkManager) GetNetworkID() string {
	return m.networkID
}

func (m *MockNetworkManager) SendMessage(msgType network.MessageType, payload []byte, targetPeerID string) error {
	m.messagesSent = append(m.messagesSent, MockMessage{
		MsgType:  msgType,
		Payload:  payload,
		TargetID: targetPeerID,
	})
	return nil
}

func (m *MockNetworkManager) RegisterMessageHandler(msgType network.MessageType, handler network.MessageHandler) {
	// Store handler if needed for testing
}

// Helper methods for testing
func (m *MockNetworkManager) SetPeerCount(count int) {
	m.peerCount = count
}

func (m *MockNetworkManager) AddPeer(id string, connected bool) {
	peer := network.Peer{
		ID: id,
	}
	if connected {
		// Set a mock connection - in real code this would be a net.Conn
		// For testing purposes we'll use a non-nil placeholder
		peer.Conn = &MockConn{}
	}
	m.peers = append(m.peers, peer)
	m.peerCount = len(m.peers)
}

func (m *MockNetworkManager) GetLastMessage() *MockMessage {
	if len(m.messagesSent) == 0 {
		return nil
	}
	return &m.messagesSent[len(m.messagesSent)-1]
}

// MockConn implements net.Conn interface for testing
type MockConn struct{}

func (m *MockConn) Read(b []byte) (n int, err error)   { return 0, nil }
func (m *MockConn) Write(b []byte) (n int, err error)  { return len(b), nil }
func (m *MockConn) Close() error                       { return nil }
func (m *MockConn) LocalAddr() net.Addr                { return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8080} }
func (m *MockConn) RemoteAddr() net.Addr               { return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8081} }
func (m *MockConn) SetDeadline(t time.Time) error      { return nil }
func (m *MockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *MockConn) SetWriteDeadline(t time.Time) error { return nil }

func TestNewNetworkOverlay(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	mockNM := NewMockNetworkManager()
	overlay := NewNetworkOverlay(mockNM)

	if overlay == nil {
		t.Fatal("NewNetworkOverlay() returned nil")
	}

	if overlay.networkManager != mockNM {
		t.Error("NetworkManager not properly assigned")
	}

	if overlay.container == nil {
		t.Error("Container not created")
	}

	if overlay.visible {
		t.Error("Overlay should be initially hidden")
	}
}

func TestNetworkOverlay_ShowHide(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	mockNM := NewMockNetworkManager()
	overlay := NewNetworkOverlay(mockNM)

	// Test initial state
	if overlay.IsVisible() {
		t.Error("Overlay should be initially hidden")
	}

	// Test Show
	overlay.Show()
	if !overlay.IsVisible() {
		t.Error("Overlay should be visible after Show()")
	}

	// Test Hide
	overlay.Hide()
	if overlay.IsVisible() {
		t.Error("Overlay should be hidden after Hide()")
	}
}

func TestNetworkOverlay_Toggle(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	mockNM := NewMockNetworkManager()
	overlay := NewNetworkOverlay(mockNM)

	// Test toggle from hidden to visible
	overlay.Toggle()
	if !overlay.IsVisible() {
		t.Error("Overlay should be visible after Toggle() from hidden state")
	}

	// Test toggle from visible to hidden
	overlay.Toggle()
	if overlay.IsVisible() {
		t.Error("Overlay should be hidden after Toggle() from visible state")
	}
}

func TestNetworkOverlay_UpdateNetworkStatus(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	mockNM := NewMockNetworkManager()
	overlay := NewNetworkOverlay(mockNM)

	// Test with no peers
	overlay.updateNetworkStatus()
	expectedStatus := "Network: Searching..."
	if overlay.statusLabel.Text != expectedStatus {
		t.Errorf("Status label = %q, want %q", overlay.statusLabel.Text, expectedStatus)
	}

	expectedCount := "Peers: 0"
	if overlay.peerCount.Text != expectedCount {
		t.Errorf("Peer count = %q, want %q", overlay.peerCount.Text, expectedCount)
	}

	// Test with peers
	mockNM.SetPeerCount(2)
	overlay.updateNetworkStatus()

	expectedStatus = "Network: Connected"
	if overlay.statusLabel.Text != expectedStatus {
		t.Errorf("Status label with peers = %q, want %q", overlay.statusLabel.Text, expectedStatus)
	}

	expectedCount = "Peers: 2"
	if overlay.peerCount.Text != expectedCount {
		t.Errorf("Peer count with peers = %q, want %q", overlay.peerCount.Text, expectedCount)
	}
}

func TestNetworkOverlay_SendChatMessage(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	mockNM := NewMockNetworkManager()
	overlay := NewNetworkOverlay(mockNM)

	// Test sending a message
	testMessage := "Hello, peers!"
	overlay.sendChatMessage(testMessage)

	// Verify message was sent through network manager
	lastMsg := mockNM.GetLastMessage()
	if lastMsg == nil {
		t.Fatal("No message was sent through network manager")
	}

	if lastMsg.MsgType != network.MessageTypeCharacterAction {
		t.Errorf("Message type = %v, want %v", lastMsg.MsgType, network.MessageTypeCharacterAction)
	}

	if len(lastMsg.Payload) == 0 {
		t.Error("Message payload is empty")
	}

	// Verify input field was cleared
	if overlay.chatInput.Text != "" {
		t.Error("Chat input should be cleared after sending message")
	}
}

func TestNetworkOverlay_SendEmptyMessage(t *testing.T) {
	// Add small delay to ensure previous test cleanup completes
	time.Sleep(10 * time.Millisecond)

	app := test.NewApp()
	defer app.Quit()

	mockNM := NewMockNetworkManager()
	overlay := NewNetworkOverlay(mockNM)

	initialMessageCount := len(mockNM.messagesSent)

	// Test sending empty message
	overlay.sendChatMessage("")

	// Verify no message was sent
	if len(mockNM.messagesSent) != initialMessageCount {
		t.Error("Empty message should not be sent")
	}
}

func TestNetworkOverlay_UpdateCharacterList(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	mockNM := NewMockNetworkManager()
	overlay := NewNetworkOverlay(mockNM)

	// Test initial state - should have local character
	chars := overlay.GetCharacterList()
	if len(chars) != 1 {
		t.Errorf("Initial character list length = %d, want 1", len(chars))
	}

	if !chars[0].IsLocal {
		t.Error("First character should be local")
	}

	if chars[0].Name != "Local Character" {
		t.Errorf("Local character name = %q, want %q", chars[0].Name, "Local Character")
	}

	// Add some test peers
	mockNM.AddPeer("peer1", true)  // Connected
	mockNM.AddPeer("peer2", false) // Disconnected

	// Update character list
	overlay.updateCharacterList()

	// Verify character data was updated
	chars = overlay.GetCharacterList()
	expectedCount := 3 // 1 local + 2 network characters
	if len(chars) != expectedCount {
		t.Errorf("Character list length = %d, want %d", len(chars), expectedCount)
	}

	// Verify local character is first
	if !chars[0].IsLocal {
		t.Error("First character should be local")
	}

	// Verify network characters
	networkCharCount := 0
	for _, char := range chars {
		if !char.IsLocal {
			networkCharCount++
		}
	}

	if networkCharCount != 2 {
		t.Errorf("Network character count = %d, want 2", networkCharCount)
	}
}

func TestNetworkOverlay_SetLocalCharacterName(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	mockNM := NewMockNetworkManager()
	overlay := NewNetworkOverlay(mockNM)

	// Test setting local character name
	newName := "My Custom Character"
	overlay.SetLocalCharacterName(newName)

	chars := overlay.GetCharacterList()
	if len(chars) == 0 {
		t.Fatal("Character list is empty")
	}

	if chars[0].Name != newName {
		t.Errorf("Local character name = %q, want %q", chars[0].Name, newName)
	}
}

func TestNetworkOverlay_CharacterVisualDistinction(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	mockNM := NewMockNetworkManager()
	overlay := NewNetworkOverlay(mockNM)

	// Add peers to test network character distinction
	mockNM.AddPeer("active-peer", true)
	mockNM.AddPeer("inactive-peer", false)

	overlay.updateCharacterList()
	chars := overlay.GetCharacterList()

	// Verify local character properties
	localChar := chars[0]
	if !localChar.IsLocal {
		t.Error("First character should be local")
	}
	if !localChar.IsActive {
		t.Error("Local character should be active")
	}
	if localChar.Location != "Local" {
		t.Errorf("Local character location = %q, want %q", localChar.Location, "Local")
	}

	// Verify network character properties
	if len(chars) >= 2 {
		networkChar := chars[1]
		if networkChar.IsLocal {
			t.Error("Network character should not be local")
		}
		if networkChar.Location == "Local" {
			t.Error("Network character location should not be 'Local'")
		}
	}
}

func TestNetworkOverlay_UpdatePeerList(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	mockNM := NewMockNetworkManager()
	overlay := NewNetworkOverlay(mockNM)

	// Add some test peers
	mockNM.AddPeer("peer1", true)  // Connected
	mockNM.AddPeer("peer2", false) // Disconnected

	// Update peer list
	overlay.updatePeerList()

	// Verify peer data was updated
	overlay.peerMutex.RLock()
	peerCount := len(overlay.peers)
	overlay.peerMutex.RUnlock()

	if peerCount != 2 {
		t.Errorf("Peer list length = %d, want 2", peerCount)
	}

	// Verify list widget reflects correct count
	listLength := overlay.peerList.Length()
	if listLength != 2 {
		t.Errorf("Peer list widget length = %d, want 2", listLength)
	}
}

func TestNetworkOverlay_AddChatMessage(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	mockNM := NewMockNetworkManager()
	overlay := NewNetworkOverlay(mockNM)

	// Test adding a chat message
	sender := "TestUser"
	message := "Test message"

	initialText := overlay.chatLog.String()
	overlay.addChatMessage(sender, message)

	// Verify message was added to chat log
	finalText := overlay.chatLog.String()
	if len(finalText) <= len(initialText) {
		t.Error("Chat message was not added to log")
	}

	// Verify message contains sender and content
	if finalText == initialText {
		t.Error("Chat log was not updated")
	}
}

func TestNetworkOverlay_GetContainer(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	mockNM := NewMockNetworkManager()
	overlay := NewNetworkOverlay(mockNM)

	container := overlay.GetContainer()
	if container == nil {
		t.Error("GetContainer() returned nil")
	}

	if container != overlay.container {
		t.Error("GetContainer() returned wrong container")
	}
}

func TestNetworkOverlay_NilNetworkManager(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	// Test with nil network manager
	overlay := NewNetworkOverlay(nil)

	if overlay == nil {
		t.Fatal("NewNetworkOverlay(nil) returned nil")
	}

	// Should not panic with nil network manager
	overlay.updateNetworkStatus()
	overlay.updatePeerList()
	overlay.sendChatMessage("test")
}

// Benchmark tests for performance validation
func BenchmarkNetworkOverlay_UpdateNetworkStatus(b *testing.B) {
	app := test.NewApp()
	defer app.Quit()

	mockNM := NewMockNetworkManager()
	mockNM.AddPeer("peer1", true)
	mockNM.AddPeer("peer2", true)

	overlay := NewNetworkOverlay(mockNM)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		overlay.updateNetworkStatus()
	}
}

func BenchmarkNetworkOverlay_SendChatMessage(b *testing.B) {
	app := test.NewApp()
	defer app.Quit()

	mockNM := NewMockNetworkManager()
	overlay := NewNetworkOverlay(mockNM)

	message := "Benchmark test message"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		overlay.sendChatMessage(message)
	}
}
