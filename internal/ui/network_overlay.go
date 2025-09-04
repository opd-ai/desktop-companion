package ui

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/opd-ai/desktop-companion/internal/character"
	"github.com/opd-ai/desktop-companion/internal/network"
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
	Name        string
	Location    string // "Local" or peer ID
	IsLocal     bool
	IsActive    bool
	CharType    string                       // Character archetype/type
	PeerID      string                       // Network peer identifier for compatibility tracking
	Personality *character.PersonalityConfig // For compatibility calculations
}

// NetworkOverlay displays multiplayer network status as an optional UI overlay
// Uses Fyne widgets to avoid custom implementations - follows "lazy programmer" approach
type NetworkOverlay struct {
	widget.BaseWidget
	networkManager NetworkManagerInterface
	container      *fyne.Container
	statusLabel    *widget.Label
	peerList       *widget.List
	characterList  *widget.List // New: shows local vs network characters
	peerCount      *widget.Label
	chatLog        *widget.RichText
	chatScroll     *container.Scroll // Auto-scroll container for chat
	chatInput      *widget.Entry
	sendButton     *widget.Button
	visible        bool
	updateTicker   *time.Ticker
	stopUpdate     chan bool
	mu             sync.RWMutex // Protects updateTicker and background goroutine state

	// Peer data for list widget
	peers     []network.Peer
	peerMutex sync.RWMutex

	// Character data for visual distinction between local and network characters
	characters     []CharacterInfo
	characterMutex sync.RWMutex
	localCharName  string // Name of the local character

	// Feature 5: Friendship Compatibility Scoring
	compatibilityCalculator *character.CompatibilityCalculator
	compatibilityScores     map[string]float64 // peer ID -> compatibility score
	compatibilityMutex      sync.RWMutex       // Protects compatibility data

	// Feature 9: Network Peer Activity Feed
	activityTracker *network.ActivityTracker
	activityFeed    *ActivityFeed
}

// NewNetworkOverlay creates a new network overlay widget
// Only creates UI elements when network manager is provided
func NewNetworkOverlay(nm NetworkManagerInterface) *NetworkOverlay {
	no := &NetworkOverlay{
		networkManager:      nm,
		visible:             false,
		stopUpdate:          make(chan bool, 1),
		peers:               make([]network.Peer, 0),
		characters:          make([]CharacterInfo, 0),
		localCharName:       "Local Character", // Default name, can be updated
		compatibilityScores: make(map[string]float64),
	}

	// Feature 9: Initialize activity tracker and feed
	no.activityTracker = network.NewActivityTracker(100) // Store up to 100 events
	no.activityFeed = NewActivityFeed(no.activityTracker)

	no.ExtendBaseWidget(no)
	no.createNetworkWidgets()

	return no
}

// GetNetworkManager returns the network manager interface
func (no *NetworkOverlay) GetNetworkManager() NetworkManagerInterface {
	return no.networkManager
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
				var locationIcon, statusIcon, compatibilityText string
				if char.IsLocal {
					locationIcon = "🏠" // House icon for local
				} else {
					locationIcon = "🌐" // Globe icon for network/remote

					// Feature 5: Show compatibility score for network characters
					if char.PeerID != "" {
						score := no.GetCompatibilityScore(char.PeerID)
						category := no.GetCompatibilityCategory(char.PeerID)

						// Color-coded compatibility indicator
						var compatIcon string
						switch {
						case score >= 0.8:
							compatIcon = "💚" // Green heart - high compatibility
						case score >= 0.6:
							compatIcon = "💛" // Yellow heart - good compatibility
						case score >= 0.4:
							compatIcon = "🧡" // Orange heart - fair compatibility
						default:
							compatIcon = "❤️" // Red heart - low compatibility
						}

						compatibilityText = fmt.Sprintf(" %s %s", compatIcon, category)
					}
				}

				if char.IsActive {
					statusIcon = "✅" // Active
				} else {
					statusIcon = "💤" // Idle
				}

				displayText := fmt.Sprintf("%s %s %s (%s)%s",
					locationIcon, statusIcon, char.Name, char.Location, compatibilityText)
				obj.(*widget.Label).SetText(displayText)
			}
		},
	)
	no.characterList.Resize(fyne.NewSize(200, 80))

	// Chat log display with auto-scroll container
	no.chatLog = widget.NewRichText()
	no.chatLog.Resize(fyne.NewSize(200, 80))
	no.chatScroll = container.NewScroll(no.chatLog)
	no.chatScroll.Resize(fyne.NewSize(200, 80))

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
		no.chatScroll,
	)

	// Activity feed section (Feature 9)
	activitySection := container.NewBorder(
		widget.NewLabel("Recent Activity:"),
		nil, nil, nil,
		no.activityFeed.GetContainer(),
	)

	// Main container with all network UI elements - character section added
	no.container = container.NewVBox(
		headerContainer,
		widget.NewSeparator(),
		characterSection, // Show characters first as this is most important for users
		widget.NewSeparator(),
		peerSection,
		widget.NewSeparator(),
		activitySection, // Activity feed before chat for better UX
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

	// Style the main container - increased height to accommodate activity feed
	no.container.Resize(fyne.NewSize(220, 480))

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

	// Track chat activity
	no.TrackChatMessage("local", no.localCharName, message)
}

// addChatMessage adds a message to the chat log display
func (no *NetworkOverlay) addChatMessage(sender, message string) {
	timestamp := time.Now().Format("15:04")
	formattedMessage := fmt.Sprintf("[%s] %s: %s\n", timestamp, sender, message)

	// Get current text and append new message
	currentText := no.chatLog.String()
	newText := fmt.Sprintf("%s%s", currentText, formattedMessage)
	no.chatLog.ParseMarkdown(newText)

	// Auto-scroll to bottom when new message arrives
	if no.chatScroll != nil {
		// Scroll to bottom after content update
		no.chatScroll.ScrollToBottom()
	}
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
	// Clear existing character list
	no.characters = no.characters[:0]

	// Always add local character first
	localChar := CharacterInfo{
		Name:        no.localCharName,
		Location:    "Local",
		IsLocal:     true,
		IsActive:    true, // Assume local character is always active
		CharType:    "Local",
		PeerID:      "",  // No peer ID for local character
		Personality: nil, // Local personality managed by compatibility calculator
	}
	no.characters = append(no.characters, localChar)

	// Add network characters from peers
	if no.networkManager != nil {
		peers := no.networkManager.GetPeers()
		for _, peer := range peers {
			// Each peer may have one or more characters
			// For now, assume one character per peer
			networkChar := CharacterInfo{
				Name:        fmt.Sprintf("%s's Character", peer.ID),
				Location:    peer.ID,
				IsLocal:     false,
				IsActive:    peer.Conn != nil, // Active if connected
				CharType:    "Network",
				PeerID:      peer.ID,
				Personality: no.getPersonalityFromPeer(peer), // Get personality from peer data when available
			}
			no.characters = append(no.characters, networkChar)
		}
	}
	no.characterMutex.Unlock()

	// Update compatibility scores after character list changes
	no.UpdateCompatibilityScores()

	// Refresh the character list widget AFTER releasing the mutex
	// This prevents deadlock when Fyne calls back into list functions
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

// SetCompatibilityCalculator sets the compatibility calculator for personality-based scoring
// Feature 5: Friendship Compatibility Scoring implementation
func (no *NetworkOverlay) SetCompatibilityCalculator(calculator *character.CompatibilityCalculator) {
	no.compatibilityMutex.Lock()
	defer no.compatibilityMutex.Unlock()
	no.compatibilityCalculator = calculator
}

// UpdateCompatibilityScores calculates compatibility scores for all network characters
// Uses the personality-based compatibility calculator to determine relationship potential
func (no *NetworkOverlay) UpdateCompatibilityScores() {
	no.compatibilityMutex.Lock()
	calculator := no.compatibilityCalculator
	no.compatibilityMutex.Unlock()

	if calculator == nil {
		return
	}

	// Get current characters and calculate compatibility
	no.characterMutex.RLock()
	characters := make([]CharacterInfo, len(no.characters))
	copy(characters, no.characters)
	no.characterMutex.RUnlock()

	no.compatibilityMutex.Lock()
	defer no.compatibilityMutex.Unlock()

	for _, char := range characters {
		if !char.IsLocal && char.Personality != nil {
			score := calculator.CalculateCompatibility(char.Personality)
			no.compatibilityScores[char.PeerID] = score
		}
	}
}

// GetCompatibilityScore returns the compatibility score for a specific peer
// Returns 0.5 (neutral) if no score is available
func (no *NetworkOverlay) GetCompatibilityScore(peerID string) float64 {
	no.compatibilityMutex.RLock()
	defer no.compatibilityMutex.RUnlock()

	if score, exists := no.compatibilityScores[peerID]; exists {
		return score
	}
	return 0.5 // Neutral compatibility
}

// GetCompatibilityCategory returns a human-readable compatibility category
func (no *NetworkOverlay) GetCompatibilityCategory(peerID string) string {
	score := no.GetCompatibilityScore(peerID)

	no.compatibilityMutex.RLock()
	calculator := no.compatibilityCalculator
	no.compatibilityMutex.RUnlock()

	if calculator != nil {
		return calculator.GetCompatibilityCategory(score)
	}

	// Fallback categorization
	switch {
	case score >= 0.8:
		return "Very Good"
	case score >= 0.6:
		return "Good"
	case score >= 0.4:
		return "Fair"
	default:
		return "Poor"
	}
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

// Feature 9: Activity Tracking Methods

// TrackPeerJoined records when a peer joins the network
func (no *NetworkOverlay) TrackPeerJoined(peerID, characterName string) {
	if no.activityTracker == nil {
		return
	}

	event := network.CreatePeerJoinedEvent(peerID, characterName)
	no.activityTracker.AddEvent(event)
}

// TrackPeerLeft records when a peer leaves the network
func (no *NetworkOverlay) TrackPeerLeft(peerID, characterName string) {
	if no.activityTracker == nil {
		return
	}

	event := network.CreatePeerLeftEvent(peerID, characterName)
	no.activityTracker.AddEvent(event)
}

// TrackCharacterAction records character interactions
func (no *NetworkOverlay) TrackCharacterAction(peerID, characterName, action string, details interface{}) {
	if no.activityTracker == nil {
		return
	}

	event := network.CreateCharacterActionEvent(peerID, characterName, action, details)
	no.activityTracker.AddEvent(event)
}

// TrackChatMessage records chat messages
func (no *NetworkOverlay) TrackChatMessage(peerID, characterName, message string) {
	if no.activityTracker == nil {
		return
	}

	event := network.CreateChatEvent(peerID, characterName, message)
	no.activityTracker.AddEvent(event)
}

// TrackBattleAction records battle-related activities
func (no *NetworkOverlay) TrackBattleAction(peerID, characterName, battleAction string) {
	if no.activityTracker == nil {
		return
	}

	event := network.CreateBattleEvent(peerID, characterName, battleAction)
	no.activityTracker.AddEvent(event)
}

// GetActivityTracker returns the activity tracker for external use
func (no *NetworkOverlay) GetActivityTracker() *network.ActivityTracker {
	return no.activityTracker
}

// GetActivityFeed returns the activity feed widget for external use
func (no *NetworkOverlay) GetActivityFeed() *ActivityFeed {
	return no.activityFeed
}

// getPersonalityFromPeer retrieves personality data from peer information
// Implements basic personality inference from peer behavior when exchange is not available
func (no *NetworkOverlay) getPersonalityFromPeer(peer network.Peer) *character.PersonalityConfig {
	// Future enhancement: Check for personality data in network protocol
	// When personality exchange is implemented, this would parse structured personality data

	// Fallback: Generate basic personality from peer ID patterns
	// This provides immediate functionality while personality exchange is being developed
	fallbackPersonality := &character.PersonalityConfig{
		Traits:        make(map[string]float64),
		Compatibility: make(map[string]float64),
	}

	// Infer basic traits from peer ID naming patterns
	peerID := strings.ToLower(peer.ID)
	switch {
	case strings.Contains(peerID, "shy") || strings.Contains(peerID, "quiet"):
		fallbackPersonality.Traits["shyness"] = 0.8
		fallbackPersonality.Traits["openness"] = 0.3
		fallbackPersonality.Compatibility["gentle"] = 0.9
	case strings.Contains(peerID, "flirty") || strings.Contains(peerID, "social"):
		fallbackPersonality.Traits["extroversion"] = 0.9
		fallbackPersonality.Traits["openness"] = 0.8
		fallbackPersonality.Compatibility["outgoing"] = 0.9
	case strings.Contains(peerID, "tsundere"):
		fallbackPersonality.Traits["defensiveness"] = 0.7
		fallbackPersonality.Traits["hidden_affection"] = 0.6
		fallbackPersonality.Compatibility["patient"] = 0.8
	default:
		// Default balanced personality for unknown peers
		fallbackPersonality.Traits["openness"] = 0.5
		fallbackPersonality.Traits["friendliness"] = 0.6
		fallbackPersonality.Compatibility["balanced"] = 0.7
	}

	return fallbackPersonality
}
