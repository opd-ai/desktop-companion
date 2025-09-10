package character

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// NetworkEventManager extends GeneralEventManager with multiplayer capabilities
// following the project's principle of minimal invasive changes
type NetworkEventManager struct {
	*GeneralEventManager                          // Embed existing manager
	networkInterface     NetworkInterface         // Network communication interface
	peerManager          PeerManagerInterface     // Peer state management interface
	groupSessions        map[string]*GroupSession // Active group conversations
	peerEventCallbacks   map[PeerEventType][]PeerEventCallback
	mu                   sync.RWMutex // Protects concurrent access
	enabled              bool         // Whether network events are enabled
}

// NetworkInterface defines methods required for network communication
// Uses interface pattern for testability following project standards
type NetworkInterface interface {
	SendMessage(msgType string, payload []byte, targetPeerID string) error
	RegisterMessageHandler(msgType string, handler func([]byte, string) error)
	BroadcastMessage(msgType string, payload []byte) error
	GetConnectedPeers() []string
	GetLocalPeerID() string
}

// PeerManagerInterface defines methods for peer state management
type PeerManagerInterface interface {
	GetPeerInfo(peerID string) (*PeerInfo, error)
	IsValidPeer(peerID string) bool
	AddPeerEventListener(callback func(eventType PeerEventType, peerID string))
}

// PeerInfo contains information about a connected peer
type PeerInfo struct {
	ID           string                 `json:"id"`
	CharacterID  string                 `json:"characterId"`
	Nickname     string                 `json:"nickname,omitempty"`
	Capabilities []string               `json:"capabilities,omitempty"` // e.g., "group_events", "voice_chat"
	LastSeen     time.Time              `json:"lastSeen"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// GroupSession represents an active multiplayer conversation
type GroupSession struct {
	ID               string                 `json:"id"`
	EventName        string                 `json:"eventName"`
	Participants     []string               `json:"participants"`        // Peer IDs
	InitiatorID      string                 `json:"initiatorId"`         // Who started the session
	CurrentState     string                 `json:"currentState"`        // "waiting", "active", "voting", "completed"
	StateData        map[string]interface{} `json:"stateData,omitempty"` // Session-specific data
	StartTime        time.Time              `json:"startTime"`
	LastActivity     time.Time              `json:"lastActivity"`
	MaxParticipants  int                    `json:"maxParticipants"`
	VoteChoices      map[string]int         `json:"voteChoices,omitempty"`      // Choice index -> vote count
	ParticipantVotes map[string]int         `json:"participantVotes,omitempty"` // Peer ID -> choice index
}

// PeerEventType represents different peer state changes
type PeerEventType string

const (
	PeerEventJoined       PeerEventType = "peer_joined"
	PeerEventLeft         PeerEventType = "peer_left"
	PeerEventCapabilities PeerEventType = "peer_capabilities_changed"
	PeerEventDisconnected PeerEventType = "peer_disconnected"
)

// PeerEventCallback handles peer state changes
type PeerEventCallback func(eventType PeerEventType, peerID string, peerInfo *PeerInfo)

// Network message types for multiplayer events
const (
	MessageTypeNetworkEvent = "network_event"
	MessageTypeGroupSession = "group_session"
	MessageTypePeerUpdate   = "peer_update"
)

// NetworkEventPayload represents network event data
type NetworkEventPayload struct {
	Type        string                 `json:"type"` // "peer_joined", "peer_left", "event_invite"
	EventName   string                 `json:"eventName,omitempty"`
	InitiatorID string                 `json:"initiatorId,omitempty"`
	SessionID   string                 `json:"sessionId,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// GroupSessionPayload represents group conversation data
type GroupSessionPayload struct {
	SessionID     string                 `json:"sessionId"`
	Action        string                 `json:"action"` // "start", "join", "vote", "response", "end"
	ParticipantID string                 `json:"participantId"`
	ChoiceIndex   int                    `json:"choiceIndex,omitempty"`  // For voting actions
	ResponseText  string                 `json:"responseText,omitempty"` // For dialog responses
	StateUpdate   map[string]interface{} `json:"stateUpdate,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
}

// NewNetworkEventManager creates a new network-aware event manager
// Wraps existing GeneralEventManager following the wrapper pattern
func NewNetworkEventManager(
	baseManager *GeneralEventManager,
	networkInterface NetworkInterface,
	peerManager PeerManagerInterface,
	enabled bool,
) *NetworkEventManager {
	nem := &NetworkEventManager{
		GeneralEventManager: baseManager,
		networkInterface:    networkInterface,
		peerManager:         peerManager,
		groupSessions:       make(map[string]*GroupSession),
		peerEventCallbacks:  make(map[PeerEventType][]PeerEventCallback),
		enabled:             enabled && baseManager != nil,
	}

	// Register network message handlers if enabled
	if nem.enabled && networkInterface != nil {
		networkInterface.RegisterMessageHandler(MessageTypeNetworkEvent, nem.handleNetworkEventMessage)
		networkInterface.RegisterMessageHandler(MessageTypeGroupSession, nem.handleGroupSessionMessage)
		networkInterface.RegisterMessageHandler(MessageTypePeerUpdate, nem.handlePeerUpdateMessage)
	}

	// Register for peer state changes if peer manager available
	if peerManager != nil {
		peerManager.AddPeerEventListener(nem.handlePeerStateChange)
	}

	return nem
}

// TriggerNetworkEvent initiates a multiplayer event that can involve other peers
// Extends the base TriggerEvent functionality with network coordination
func (nem *NetworkEventManager) TriggerNetworkEvent(eventName string, gameState *GameState, invitePeers []string) (*GeneralDialogEvent, error) {
	if !nem.enabled || nem.GeneralEventManager == nil {
		return nem.GeneralEventManager.TriggerEvent(eventName, gameState)
	}

	nem.mu.Lock()
	defer nem.mu.Unlock()

	// Find the event
	event := nem.findEventByName(eventName)
	if event == nil {
		return nil, fmt.Errorf("event not found: %s", eventName)
	}

	// Check if this is a network-capable event
	if !nem.isNetworkEvent(event) {
		// Fall back to regular event handling
		return nem.GeneralEventManager.TriggerEvent(eventName, gameState)
	}

	// Create group session for multiplayer events
	sessionID := nem.generateSessionID()
	groupSession := &GroupSession{
		ID:               sessionID,
		EventName:        eventName,
		Participants:     []string{nem.networkInterface.GetLocalPeerID()},
		InitiatorID:      nem.networkInterface.GetLocalPeerID(),
		CurrentState:     "waiting",
		StateData:        make(map[string]interface{}),
		StartTime:        time.Now(),
		LastActivity:     time.Now(),
		MaxParticipants:  nem.getMaxParticipants(event),
		VoteChoices:      make(map[string]int),
		ParticipantVotes: make(map[string]int),
	}

	nem.groupSessions[sessionID] = groupSession

	// Send invitations to specified peers
	if len(invitePeers) > 0 {
		err := nem.sendEventInvitations(sessionID, eventName, invitePeers)
		if err != nil {
			delete(nem.groupSessions, sessionID)
			return nil, fmt.Errorf("failed to send invitations: %w", err)
		}
	}

	// Store reference to active group session
	nem.activeEvent = event
	return event, nil
}

// JoinGroupSession allows a peer to join an ongoing group conversation
func (nem *NetworkEventManager) JoinGroupSession(sessionID, participantID string) error {
	if !nem.enabled {
		return fmt.Errorf("network events not enabled")
	}

	nem.mu.Lock()
	defer nem.mu.Unlock()

	session, exists := nem.groupSessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// Check if session can accept more participants
	if len(session.Participants) >= session.MaxParticipants {
		return fmt.Errorf("session at maximum capacity")
	}

	// Check if already participant
	for _, pid := range session.Participants {
		if pid == participantID {
			return nil // Already joined
		}
	}

	// Add participant
	session.Participants = append(session.Participants, participantID)
	session.LastActivity = time.Now()

	// Notify other participants
	return nem.broadcastSessionUpdate(sessionID, "join", participantID)
}

// SubmitGroupChoice handles voting in group conversations
func (nem *NetworkEventManager) SubmitGroupChoice(sessionID, participantID string, choiceIndex int) error {
	if !nem.enabled {
		return fmt.Errorf("network events not enabled")
	}

	nem.mu.Lock()
	defer nem.mu.Unlock()

	session, exists := nem.groupSessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// Validate participant
	if !nem.isParticipant(session, participantID) {
		return fmt.Errorf("not a participant in session")
	}

	// Record vote
	session.ParticipantVotes[participantID] = choiceIndex
	session.VoteChoices[fmt.Sprintf("%d", choiceIndex)]++
	session.LastActivity = time.Now()

	// Check if all participants have voted
	if len(session.ParticipantVotes) >= len(session.Participants) {
		session.CurrentState = "completed"
		nem.processGroupVoteResults(session)
	}

	// Broadcast vote update
	return nem.broadcastSessionUpdate(sessionID, "vote", participantID)
}

// GetActiveGroupSessions returns currently active group conversations
func (nem *NetworkEventManager) GetActiveGroupSessions() map[string]*GroupSession {
	if !nem.enabled {
		return make(map[string]*GroupSession)
	}

	nem.mu.RLock()
	defer nem.mu.RUnlock()

	// Return copy to prevent external modification
	result := make(map[string]*GroupSession)
	for id, session := range nem.groupSessions {
		sessionCopy := *session
		result[id] = &sessionCopy
	}
	return result
}

// AddPeerEventListener registers a callback for peer state changes
func (nem *NetworkEventManager) AddPeerEventListener(eventType PeerEventType, callback PeerEventCallback) {
	if !nem.enabled {
		return
	}

	nem.mu.Lock()
	defer nem.mu.Unlock()

	nem.peerEventCallbacks[eventType] = append(nem.peerEventCallbacks[eventType], callback)
}

// Internal helper methods

// isNetworkEvent checks if an event supports network multiplayer
func (nem *NetworkEventManager) isNetworkEvent(event *GeneralDialogEvent) bool {
	if event == nil {
		return false
	}

	// Check for network-specific keywords or categories
	for _, keyword := range event.Keywords {
		if keyword == "multiplayer" || keyword == "group" || keyword == "collaborative" {
			return true
		}
	}

	// Check category
	return event.Category == "group" || event.Category == "multiplayer"
}

// generateSessionID creates a unique session identifier
func (nem *NetworkEventManager) generateSessionID() string {
	// Use timestamp + local peer ID for uniqueness (following simple approach)
	timestamp := time.Now().UnixNano()
	localID := nem.networkInterface.GetLocalPeerID()
	return fmt.Sprintf("%s_%d", localID, timestamp)
}

// getMaxParticipants determines maximum participants for an event
func (nem *NetworkEventManager) getMaxParticipants(event *GeneralDialogEvent) int {
	// Default to reasonable group size, could be made configurable
	defaultMax := 4

	// Check if event has specific participant limit in effects
	if event.Effects != nil {
		if maxParticipants, exists := event.Effects["maxParticipants"]; exists && maxParticipants > 0 {
			return int(maxParticipants)
		}
	}

	return defaultMax
}

// sendEventInvitations sends invitations to specified peers
func (nem *NetworkEventManager) sendEventInvitations(sessionID, eventName string, peerIDs []string) error {
	payload := NetworkEventPayload{
		Type:        "event_invite",
		EventName:   eventName,
		InitiatorID: nem.networkInterface.GetLocalPeerID(),
		SessionID:   sessionID,
		Timestamp:   time.Now(),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal invitation payload: %w", err)
	}

	// Send to each invited peer
	for _, peerID := range peerIDs {
		if nem.peerManager.IsValidPeer(peerID) {
			err := nem.networkInterface.SendMessage(MessageTypeNetworkEvent, payloadBytes, peerID)
			if err != nil {
				// Log error but continue with other peers
				continue
			}
		}
	}

	return nil
}

// broadcastSessionUpdate notifies participants of session changes
func (nem *NetworkEventManager) broadcastSessionUpdate(sessionID, action, participantID string) error {
	payload := GroupSessionPayload{
		SessionID:     sessionID,
		Action:        action,
		ParticipantID: participantID,
		Timestamp:     time.Now(),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal session update: %w", err)
	}

	// Broadcast to all connected peers
	return nem.networkInterface.BroadcastMessage(MessageTypeGroupSession, payloadBytes)
}

// isParticipant checks if a peer ID is in the session participants
func (nem *NetworkEventManager) isParticipant(session *GroupSession, peerID string) bool {
	for _, pid := range session.Participants {
		if pid == peerID {
			return true
		}
	}
	return false
}

// processGroupVoteResults handles the completion of group voting
func (nem *NetworkEventManager) processGroupVoteResults(session *GroupSession) {
	// Find the choice with the most votes
	maxVotes := 0
	winningChoice := -1

	for choiceStr, votes := range session.VoteChoices {
		if votes > maxVotes {
			maxVotes = votes
			// Convert string back to int (we stored as string for JSON compatibility)
			if choice, err := fmt.Sscanf(choiceStr, "%d", &winningChoice); err == nil && choice == 1 {
				// Successfully parsed
			}
		}
	}

	// Initialize StateData if nil
	if session.StateData == nil {
		session.StateData = make(map[string]interface{})
	}

	// Store winning choice for reference
	session.StateData["winningChoice"] = winningChoice
	session.StateData["totalVotes"] = len(session.ParticipantVotes)
}

// Network message handlers

// handleNetworkEventMessage processes incoming network event messages
func (nem *NetworkEventManager) handleNetworkEventMessage(payload []byte, fromPeerID string) error {
	if !nem.enabled {
		return nil
	}

	var eventPayload NetworkEventPayload
	if err := json.Unmarshal(payload, &eventPayload); err != nil {
		return fmt.Errorf("failed to unmarshal network event payload: %w", err)
	}

	switch eventPayload.Type {
	case "event_invite":
		return nem.handleEventInvitation(eventPayload, fromPeerID)
	case "peer_joined":
		return nem.handlePeerJoinedEvent(eventPayload, fromPeerID)
	case "peer_left":
		return nem.handlePeerLeftEvent(eventPayload, fromPeerID)
	default:
		return fmt.Errorf("unknown network event type: %s", eventPayload.Type)
	}
}

// handleGroupSessionMessage processes group session messages
func (nem *NetworkEventManager) handleGroupSessionMessage(payload []byte, fromPeerID string) error {
	if !nem.enabled {
		return nil
	}

	var sessionPayload GroupSessionPayload
	if err := json.Unmarshal(payload, &sessionPayload); err != nil {
		return fmt.Errorf("failed to unmarshal group session payload: %w", err)
	}

	nem.mu.Lock()
	defer nem.mu.Unlock()

	session, exists := nem.groupSessions[sessionPayload.SessionID]
	if !exists {
		// Session not found - might be a new session we should join
		if sessionPayload.Action == "start" {
			return nem.handleNewGroupSession(sessionPayload, fromPeerID)
		}
		return fmt.Errorf("session not found: %s", sessionPayload.SessionID)
	}

	// Update last activity
	session.LastActivity = time.Now()

	switch sessionPayload.Action {
	case "join":
		return nem.handleSessionJoin(session, sessionPayload, fromPeerID)
	case "vote":
		return nem.handleSessionVote(session, sessionPayload, fromPeerID)
	case "end":
		return nem.handleSessionEnd(session, sessionPayload, fromPeerID)
	default:
		return fmt.Errorf("unknown session action: %s", sessionPayload.Action)
	}
}

// handlePeerUpdateMessage processes peer status updates
func (nem *NetworkEventManager) handlePeerUpdateMessage(payload []byte, fromPeerID string) error {
	if !nem.enabled {
		return nil
	}

	// Trigger peer event callbacks
	nem.mu.RLock()
	callbacks := nem.peerEventCallbacks[PeerEventCapabilities]
	nem.mu.RUnlock()

	if peerInfo, err := nem.peerManager.GetPeerInfo(fromPeerID); err == nil {
		for _, callback := range callbacks {
			callback(PeerEventCapabilities, fromPeerID, peerInfo)
		}
	}

	return nil
}

// Event-specific handlers

// handleEventInvitation processes incoming event invitations
func (nem *NetworkEventManager) handleEventInvitation(payload NetworkEventPayload, fromPeerID string) error {
	// Validate the invitation
	if !nem.peerManager.IsValidPeer(fromPeerID) {
		return fmt.Errorf("invitation from unknown peer: %s", fromPeerID)
	}

	// Check if we have the requested event
	event := nem.findEventByName(payload.EventName)
	if event == nil {
		return fmt.Errorf("unknown event: %s", payload.EventName)
	}

	// Auto-join if this is a public event, otherwise user decision required
	// For now, implement simple auto-join logic
	if nem.shouldAutoJoinEvent(event, fromPeerID) {
		// Create local session reference if it doesn't exist
		nem.mu.Lock()
		if _, exists := nem.groupSessions[payload.SessionID]; !exists {
			nem.groupSessions[payload.SessionID] = &GroupSession{
				ID:               payload.SessionID,
				EventName:        payload.EventName,
				Participants:     []string{payload.InitiatorID},
				InitiatorID:      payload.InitiatorID,
				CurrentState:     "waiting",
				StateData:        make(map[string]interface{}),
				StartTime:        time.Now(),
				LastActivity:     time.Now(),
				MaxParticipants:  nem.getMaxParticipants(event),
				VoteChoices:      make(map[string]int),
				ParticipantVotes: make(map[string]int),
			}
		}
		nem.mu.Unlock()

		return nem.JoinGroupSession(payload.SessionID, nem.networkInterface.GetLocalPeerID())
	}

	return nil
}

// handlePeerJoinedEvent processes peer join notifications
func (nem *NetworkEventManager) handlePeerJoinedEvent(payload NetworkEventPayload, fromPeerID string) error {
	// Trigger peer joined callbacks
	nem.mu.RLock()
	callbacks := nem.peerEventCallbacks[PeerEventJoined]
	nem.mu.RUnlock()

	if peerInfo, err := nem.peerManager.GetPeerInfo(fromPeerID); err == nil {
		for _, callback := range callbacks {
			callback(PeerEventJoined, fromPeerID, peerInfo)
		}
	}

	return nil
}

// handlePeerLeftEvent processes peer leave notifications
func (nem *NetworkEventManager) handlePeerLeftEvent(payload NetworkEventPayload, fromPeerID string) error {
	// Remove peer from any active sessions
	nem.mu.Lock()
	for sessionID, session := range nem.groupSessions {
		for i, participantID := range session.Participants {
			if participantID == fromPeerID {
				// Remove participant
				session.Participants = append(session.Participants[:i], session.Participants[i+1:]...)
				session.LastActivity = time.Now()

				// End session if no participants left
				if len(session.Participants) == 0 {
					delete(nem.groupSessions, sessionID)
				}
				break
			}
		}
	}
	nem.mu.Unlock()

	// Trigger peer left callbacks
	nem.mu.RLock()
	callbacks := nem.peerEventCallbacks[PeerEventLeft]
	nem.mu.RUnlock()

	if peerInfo, err := nem.peerManager.GetPeerInfo(fromPeerID); err == nil {
		for _, callback := range callbacks {
			callback(PeerEventLeft, fromPeerID, peerInfo)
		}
	}

	return nil
}

// Session-specific handlers

// handleNewGroupSession processes new group session notifications
func (nem *NetworkEventManager) handleNewGroupSession(payload GroupSessionPayload, fromPeerID string) error {
	// This would typically require user confirmation, but for MVP auto-join
	return nem.JoinGroupSession(payload.SessionID, nem.networkInterface.GetLocalPeerID())
}

// handleSessionJoin processes session join notifications
func (nem *NetworkEventManager) handleSessionJoin(session *GroupSession, payload GroupSessionPayload, fromPeerID string) error {
	// Add participant if not already present
	if !nem.isParticipant(session, payload.ParticipantID) {
		session.Participants = append(session.Participants, payload.ParticipantID)
	}
	return nil
}

// handleSessionVote processes voting updates
func (nem *NetworkEventManager) handleSessionVote(session *GroupSession, payload GroupSessionPayload, fromPeerID string) error {
	// Update vote counts
	choiceKey := fmt.Sprintf("%d", payload.ChoiceIndex)
	session.VoteChoices[choiceKey]++
	session.ParticipantVotes[payload.ParticipantID] = payload.ChoiceIndex
	return nil
}

// handleSessionEnd processes session termination
func (nem *NetworkEventManager) handleSessionEnd(session *GroupSession, payload GroupSessionPayload, fromPeerID string) error {
	session.CurrentState = "completed"
	// Session cleanup will be handled by periodic cleanup
	return nil
}

// handlePeerStateChange processes peer state changes from the peer manager
func (nem *NetworkEventManager) handlePeerStateChange(eventType PeerEventType, peerID string) {
	if !nem.enabled {
		return
	}

	// Get peer info for the callback
	peerInfo, err := nem.peerManager.GetPeerInfo(peerID)
	if err != nil {
		peerInfo = &PeerInfo{ID: peerID, LastSeen: time.Now()}
	}

	// Trigger registered callbacks
	nem.mu.RLock()
	callbacks := nem.peerEventCallbacks[eventType]
	nem.mu.RUnlock()

	for _, callback := range callbacks {
		callback(eventType, peerID, peerInfo)
	}
}

// shouldAutoJoinEvent determines if we should automatically join an event invitation
func (nem *NetworkEventManager) shouldAutoJoinEvent(event *GeneralDialogEvent, fromPeerID string) bool {
	// Simple policy: auto-join from valid peers for public events
	// Could be made more sophisticated with user preferences
	return nem.peerManager.IsValidPeer(fromPeerID) &&
		(event.Category == "conversation" || event.Category == "humor")
}
