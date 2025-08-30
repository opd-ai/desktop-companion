package ui

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"desktop-companion/internal/network"
)

// NetworkManagerInterface defines the interface needed by NetworkOverlay
type NetworkManagerInterface interface {
	GetPeerCount() int
	GetPeers() []network.Peer
	GetNetworkID() string
	SendMessage(msgType network.MessageType, payload []byte, targetPeerID string) error
	RegisterMessageHandler(msgType network.MessageType, handler network.MessageHandler)
}

// CharacterInfo represents a character's location and status for UI display
type CharacterInfo struct {
	Name      string
	Location  string // "Local" or peer ID
	IsLocal   bool
	IsActive  bool
	CharType  string // Character archetype/type
}

// NetworkOverlay displays multiplayer network status as an optional UI overlay
// Uses Fyne widgets to avoid custom implementations - follows "lazy programmer" approach
type NetworkOverlay struct {
	widget.BaseWidget
	networkManager  NetworkManagerInterface
	container      *fyne.Container
	statusLabel    *widget.Label
	peerList       *widget.List
	characterList  *widget.List  // New: shows local vs network characters
	peerCount      *widget.Label
	chatLog        *widget.RichText
	chatInput      *widget.Entry
	sendButton     *widget.Button
	visible        bool
	updateTicker   *time.Ticker
	stopUpdate     chan bool
	mu             sync.RWMutex // Protects updateTicker and background goroutine state
	
	// Peer data for list widget
	peers          []network.Peer
	peerMutex      sync.RWMutex
	
	// Character data for visual distinction between local and network characters
	characters     []CharacterInfo
	characterMutex sync.RWMutex
	localCharName  string  // Name of the local character
}

// NewNetworkOverlay creates a new network overlay widget
// Only creates UI elements when network manager is provided
func NewNetworkOverlay(nm NetworkManagerInterface) *NetworkOverlay {
	no := &NetworkOverlay{
		networkManager: nm,
		visible:        false,
		stopUpdate:     make(chan bool, 1),
		peers:          make([]network.Peer, 0),
		characters:     make([]CharacterInfo, 0),
		localCharName:  "Local Character", // Default name, can be updated
	}

	no.ExtendBaseWidget(no)
	no.createNetworkWidgets()

	return no
}

// createNetworkWidgets creates the network status and peer communication UI
// Uses standard Fyne widgets to minimize custom code
func (no *NetworkOverlay) createNetworkWidgets() {
	// Network status indicator
	no.statusLabel = widget.NewLabel("Network: Disconnected")
	no.statusLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Peer count display
	no.peerCount = widget.NewLabel("Peers: 0")

	// Peer list widget
	no.peerList = widget.NewList(
		func() int {
			no.peerMutex.RLock()
			defer no.peerMutex.RUnlock()
			return len(no.peers)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Peer")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			no.peerMutex.RLock()
			defer no.peerMutex.RUnlock()
			
			if id < len(no.peers) {
				peer := no.peers[id]
				statusIcon := "🔴" // Disconnected
				if peer.Conn != nil {
					statusIcon = "🟢" // Connected
				}
				
				obj.(*widget.Label).SetText(fmt.Sprintf("%s %s", statusIcon, peer.ID))
			}
		},
	)
	no.peerList.Resize(fyne.NewSize(200, 80)) // Reduced height to make room for character list

	// Character list widget - clearly distinguishes local vs network characters
	no.characterList = widget.NewList(
		func() int {
			no.characterMutex.RLock()
			defer no.characterMutex.RUnlock()
			return len(no.characters)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Character")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			no.characterMutex.RLock()
			defer no.characterMutex.RUnlock()
			
			if id < len(no.characters) {
				char := no.characters[id]
				
				// Visual indicators for character type and status
				var locationIcon, statusIcon string
				if char.IsLocal {
					locationIcon = "🏠" // House icon for local
				} else {
					locationIcon = "🌐" // Globe icon for network/remote
				}
				
				if char.IsActive {
					statusIcon = "✅" // Active
				} else {
					statusIcon = "💤" // Idle
				}
				
				displayText := fmt.Sprintf("%s %s %s (%s)", 
					locationIcon, statusIcon, char.Name, char.Location)
				obj.(*widget.Label).SetText(displayText)
			}
		},
	)
	no.characterList.Resize(fyne.NewSize(200, 80))

	// Chat log display
	no.chatLog = widget.NewRichText()
	no.chatLog.Resize(fyne.NewSize(200, 80))
	no.chatLog.Scroll = container.ScrollVerticalOnly

	// Chat input controls
	no.chatInput = widget.NewEntry()
	no.chatInput.SetPlaceHolder("Type message...")
	no.chatInput.OnSubmitted = func(text string) {
		no.sendChatMessage(text)
	}

	no.sendButton = widget.NewButton("Send", func() {
		no.sendChatMessage(no.chatInput.Text)
	})

	// Layout components in a compact vertical arrangement
	headerContainer := container.NewHBox(no.statusLabel, layout.NewSpacer(), no.peerCount)
	
	peerSection := container.NewBorder(
		widget.NewLabel("Network Peers:"),
		nil, nil, nil,
		no.peerList,
	)

	// New character section to clearly show local vs network characters
	characterSection := container.NewBorder(
		widget.NewLabel("Characters (🏠=Local, 🌐=Network):"),
		nil, nil, nil,
		no.characterList,
	)

	chatControls := container.NewBorder(nil, nil, nil, no.sendButton, no.chatInput)
	chatSection := container.NewBorder(
		widget.NewLabel("Network Chat:"),
		chatControls, nil, nil,
		no.chatLog,
	)

	// Main container with all network UI elements - character section added
	no.container = container.NewVBox(
		headerContainer,
		widget.NewSeparator(),
		characterSection,  // Show characters first as this is most important for users
		widget.NewSeparator(),
		peerSection,
		widget.NewSeparator(),
		chatSection,
	)

	// Apply consistent styling
	no.styleNetworkWidgets()
}

// styleNetworkWidgets applies consistent visual styling to network overlay components
func (no *NetworkOverlay) styleNetworkWidgets() {
	// Set background color for better visibility over character
	// backgroundColor := color.RGBA{R: 0, G: 0, B: 0, A: 180} // Semi-transparent black
	
	// Style the main container - increased height to accommodate character list
	no.container.Resize(fyne.NewSize(220, 380))
	
	// Style status label with appropriate colors
	if no.networkManager != nil && no.networkManager.GetPeerCount() > 0 {
		no.statusLabel.SetText("Network: Connected")
		// Green tint for connected status could be added here if Fyne supports it
	}
	
	// Initialize with local character
	no.updateCharacterList()
}

// sendChatMessage handles sending messages through the network
func (no *NetworkOverlay) sendChatMessage(message string) {
	if message == "" || no.networkManager == nil {
		return
	}

	// Clear input field
	no.chatInput.SetText("")

	// Send message through network manager
	payload := []byte(fmt.Sprintf(`{"type":"chat","message":"%s","from":"%s"}`, 
		message, no.networkManager.GetNetworkID()))
	
	err := no.networkManager.SendMessage(network.MessageTypeCharacterAction, payload, "")
	if err != nil {
		no.addChatMessage("System", fmt.Sprintf("Failed to send message: %v", err))
		return
	}

	// Add to local chat log
	no.addChatMessage("You", message)
}

// addChatMessage adds a message to the chat log display
func (no *NetworkOverlay) addChatMessage(sender, message string) {
	timestamp := time.Now().Format("15:04")
	formattedMessage := fmt.Sprintf("[%s] %s: %s\n", timestamp, sender, message)
	
	// Get current text and append new message
	currentText := no.chatLog.String()
	newText := fmt.Sprintf("%s%s", currentText, formattedMessage)
	no.chatLog.ParseMarkdown(newText)
	
	// Note: Auto-scroll functionality would need custom implementation in Fyne
}

// Show makes the network overlay visible
func (no *NetworkOverlay) Show() {
	no.mu.Lock()
	defer no.mu.Unlock()

	if no.visible {
		return
	}

	no.visible = true
	no.container.Show()

	// Start periodic updates
	no.updateTicker = time.NewTicker(2 * time.Second)
	go no.updateLoop()
}

// Hide makes the network overlay invisible
func (no *NetworkOverlay) Hide() {
	no.mu.Lock()
	defer no.mu.Unlock()

	if !no.visible {
		return
	}

	no.visible = false
	no.container.Hide()

	// Stop updates
	if no.updateTicker != nil {
		no.updateTicker.Stop()
		no.updateTicker = nil
	}
	
	// Signal update loop to stop
	select {
	case no.stopUpdate <- true:
	default:
	}
}

// Toggle switches visibility state
func (no *NetworkOverlay) Toggle() {
	no.mu.RLock()
	visible := no.visible
	no.mu.RUnlock()

	if visible {
		no.Hide()
	} else {
		no.Show()
	}
}

// IsVisible returns current visibility state
func (no *NetworkOverlay) IsVisible() bool {
	no.mu.RLock()
	defer no.mu.RUnlock()
	return no.visible
}

// GetContainer returns the container for layout integration
func (no *NetworkOverlay) GetContainer() *fyne.Container {
	return no.container
}

// updateLoop periodically refreshes network status and peer information
func (no *NetworkOverlay) updateLoop() {
	no.mu.RLock()
	ticker := no.updateTicker
	no.mu.RUnlock()
	
	if ticker == nil {
		return
	}
	
	for {
		select {
		case <-no.stopUpdate:
			return
		case <-ticker.C:
			no.updateNetworkStatus()
		}
	}
}

// updateNetworkStatus refreshes displayed network information
func (no *NetworkOverlay) updateNetworkStatus() {
	if no.networkManager == nil {
		return
	}

	// Update peer count
	peerCount := no.networkManager.GetPeerCount()
	no.peerCount.SetText(fmt.Sprintf("Peers: %d", peerCount))

	// Update connection status
	if peerCount > 0 {
		no.statusLabel.SetText("Network: Connected")
	} else {
		no.statusLabel.SetText("Network: Searching...")
	}

	// Update peer list
	no.updatePeerList()
	
	// Update character list to show local vs network distinction
	no.updateCharacterList()
}

// updatePeerList refreshes the peer list display with current peer information
func (no *NetworkOverlay) updatePeerList() {
	if no.networkManager == nil {
		return
	}

	peers := no.networkManager.GetPeers()
	
	no.peerMutex.Lock()
	no.peers = peers
	no.peerMutex.Unlock()

	// Refresh the list widget
	no.peerList.Refresh()
}

// updateCharacterList refreshes the character list to clearly show local vs network characters
func (no *NetworkOverlay) updateCharacterList() {
	no.characterMutex.Lock()
	defer no.characterMutex.Unlock()
	
	// Clear existing character list
	no.characters = no.characters[:0]
	
	// Always add local character first
	localChar := CharacterInfo{
		Name:     no.localCharName,
		Location: "Local",
		IsLocal:  true,
		IsActive: true, // Assume local character is always active
		CharType: "Local",
	}
	no.characters = append(no.characters, localChar)
	
	// Add network characters from peers
	if no.networkManager != nil {
		peers := no.networkManager.GetPeers()
		for _, peer := range peers {
			// Each peer may have one or more characters
			// For now, assume one character per peer
			networkChar := CharacterInfo{
				Name:     fmt.Sprintf("%s's Character", peer.ID),
				Location: peer.ID,
				IsLocal:  false,
				IsActive: peer.Conn != nil, // Active if connected
				CharType: "Network",
			}
			no.characters = append(no.characters, networkChar)
		}
	}
	
	// Refresh the character list widget
	no.characterList.Refresh()
}

// SetLocalCharacterName updates the local character name for display
func (no *NetworkOverlay) SetLocalCharacterName(name string) {
	no.localCharName = name
	no.updateCharacterList()
}

// RegisterNetworkEvents sets up handlers for network events like peer join/leave
func (no *NetworkOverlay) RegisterNetworkEvents() {
	if no.networkManager == nil {
		return
	}

	// Register handler for character action messages (including chat)
	no.networkManager.RegisterMessageHandler(network.MessageTypeCharacterAction, 
		func(msg network.Message, from *network.Peer) error {
			// Try to parse as chat message
			var chatData map[string]interface{}
			if err := json.Unmarshal(msg.Payload, &chatData); err == nil {
				if msgType, ok := chatData["type"].(string); ok && msgType == "chat" {
					if message, ok := chatData["message"].(string); ok {
						no.addChatMessage(from.ID, message)
					}
				}
			}
			return nil
		})

	// Future: Add handlers for peer join/leave events when available
}

// GetCharacterList returns current character information (for testing)
func (no *NetworkOverlay) GetCharacterList() []CharacterInfo {
	no.characterMutex.RLock()
	defer no.characterMutex.RUnlock()
	
	// Return a copy to avoid race conditions
	characters := make([]CharacterInfo, len(no.characters))
	copy(characters, no.characters)
	return characters
}

// CreateObject implements fyne.Widget interface - required but not used
func (no *NetworkOverlay) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(no.container)
}

// Move positions the overlay at the specified coordinates
func (no *NetworkOverlay) Move(pos fyne.Position) {
	no.container.Move(pos)
}

// Resize changes the overlay size
func (no *NetworkOverlay) Resize(size fyne.Size) {
	no.container.Resize(size)
}
