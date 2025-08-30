package network

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// GroupEventManager handles multi-character scenarios and collaborative events
// Uses standard library for JSON marshaling and time management
type GroupEventManager struct {
	activeEvents    map[string]*GroupEvent        // sessionID -> event
	eventTemplates  []GroupEventTemplate          // Available event templates
	networkManager  GroupNetworkManagerInterface  // Network communication
	participants    map[string]map[string]bool    // sessionID -> participantID -> joined
	eventHistory    map[string][]CompletedEvent   // participantID -> completed events
	mu              sync.RWMutex                  // Protects concurrent access
	minParticipants int                           // Minimum participants for group events
	maxParticipants int                           // Maximum participants for group events
}

// GroupNetworkManagerInterface defines required network operations for group events
// Uses interface pattern for testability following project standards
type GroupNetworkManagerInterface interface {
	BroadcastMessage(msgType string, payload []byte) error
	SendMessage(msgType string, payload []byte, targetPeerID string) error
	RegisterMessageHandler(msgType string, handler func([]byte, string) error)
	GetConnectedPeers() []string
	GetLocalPeerID() string
}

// GroupEvent represents an active multi-character scenario
type GroupEvent struct {
	SessionID       string                 `json:"sessionId"`
	Template        GroupEventTemplate     `json:"template"`
	Participants    []string               `json:"participants"`     // Peer IDs
	InitiatorID     string                 `json:"initiatorId"`      // Who started the event
	CurrentPhase    string                 `json:"currentPhase"`     // Current phase name
	PhaseData       map[string]interface{} `json:"phaseData"`        // Phase-specific data
	StartTime       time.Time              `json:"startTime"`
	LastActivity    time.Time              `json:"lastActivity"`
	Votes           map[string]int         `json:"votes"`            // choiceID -> vote count
	ParticipantVotes map[string]string     `json:"participantVotes"` // participantID -> choiceID
	Scores          map[string]int         `json:"scores"`           // participantID -> score
	CompletedPhases []string               `json:"completedPhases"`
}

// GroupEventTemplate defines the structure of group events
type GroupEventTemplate struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Category        string                 `json:"category"`        // "scenario", "minigame", "decision"
	MinParticipants int                    `json:"minParticipants"` // 2-8 participants
	MaxParticipants int                    `json:"maxParticipants"`
	EstimatedTime   time.Duration          `json:"estimatedTime"`   // Expected duration
	Phases          []EventPhase           `json:"phases"`          // Sequential phases
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// EventPhase represents a stage within a group event
type EventPhase struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Type        string        `json:"type"`        // "intro", "choice", "vote", "result", "minigame"
	Duration    time.Duration `json:"duration"`    // Maximum phase duration
	Choices     []EventChoice `json:"choices"`     // Available choices for participants
	MinVotes    int           `json:"minVotes"`    // Minimum votes to proceed
	AutoAdvance bool          `json:"autoAdvance"` // Auto-advance when all voted
}

// EventChoice represents a choice in a group decision
type EventChoice struct {
	ID          string `json:"id"`
	Text        string `json:"text"`
	Description string `json:"description,omitempty"`
	Points      int    `json:"points"` // Points awarded for this choice
}

// CompletedEvent tracks group event history
type CompletedEvent struct {
	SessionID    string    `json:"sessionId"`
	TemplateID   string    `json:"templateId"`
	Participants []string  `json:"participants"`
	CompletedAt  time.Time `json:"completedAt"`
	FinalScore   int       `json:"finalScore"`
	Duration     time.Duration `json:"duration"`
}

// GroupEventMessage represents network messages for group events
type GroupEventMessage struct {
	Type        string                 `json:"type"` // "invite", "join", "vote", "advance", "end"
	SessionID   string                 `json:"sessionId"`
	Sender      string                 `json:"sender"`
	Data        map[string]interface{} `json:"data"`
	Timestamp   time.Time              `json:"timestamp"`
}

// NewGroupEventManager creates a new group event manager
// Follows project pattern of simple constructors with validation
func NewGroupEventManager(networkManager GroupNetworkManagerInterface, templates []GroupEventTemplate) *GroupEventManager {
	gem := &GroupEventManager{
		activeEvents:    make(map[string]*GroupEvent),
		eventTemplates:  templates,
		networkManager:  networkManager,
		participants:    make(map[string]map[string]bool),
		eventHistory:    make(map[string][]CompletedEvent),
		minParticipants: 2,
		maxParticipants: 8,
	}

	// Register network message handlers
	if networkManager != nil {
		networkManager.RegisterMessageHandler("group_event", gem.handleGroupEventMessage)
	}

	return gem
}

// StartGroupEvent initiates a new group event scenario
// Returns sessionID and error following Go error conventions
func (gem *GroupEventManager) StartGroupEvent(templateID string, initiatorID string) (string, error) {
	gem.mu.Lock()
	defer gem.mu.Unlock()

	// Find template by ID
	var template *GroupEventTemplate
	for _, t := range gem.eventTemplates {
		if t.ID == templateID {
			template = &t
			break
		}
	}
	if template == nil {
		return "", fmt.Errorf("group event template not found: %s", templateID)
	}

	// Validate minimum participants available
	connectedPeers := gem.networkManager.GetConnectedPeers()
	if len(connectedPeers)+1 < template.MinParticipants {
		return "", fmt.Errorf("insufficient participants: need %d, have %d", 
			template.MinParticipants, len(connectedPeers)+1)
	}

	// Generate unique session ID using timestamp and random component
	sessionID := fmt.Sprintf("group_%d_%d", time.Now().Unix(), rand.Intn(10000))

	// Create new group event
	event := &GroupEvent{
		SessionID:        sessionID,
		Template:         *template,
		Participants:     []string{initiatorID},
		InitiatorID:      initiatorID,
		CurrentPhase:     template.Phases[0].Name,
		PhaseData:        make(map[string]interface{}),
		StartTime:        time.Now(),
		LastActivity:     time.Now(),
		Votes:            make(map[string]int),
		ParticipantVotes: make(map[string]string),
		Scores:           make(map[string]int),
		CompletedPhases:  []string{},
	}

	gem.activeEvents[sessionID] = event
	gem.participants[sessionID] = map[string]bool{initiatorID: true}

	// Broadcast invitation to connected peers
	inviteMessage := GroupEventMessage{
		Type:      "invite",
		SessionID: sessionID,
		Sender:    initiatorID,
		Data: map[string]interface{}{
			"templateId":      templateID,
			"templateName":    template.Name,
			"description":     template.Description,
			"minParticipants": template.MinParticipants,
			"maxParticipants": template.MaxParticipants,
			"estimatedTime":   template.EstimatedTime.String(),
		},
		Timestamp: time.Now(),
	}

	if err := gem.broadcastGroupEventMessage(inviteMessage); err != nil {
		delete(gem.activeEvents, sessionID)
		delete(gem.participants, sessionID)
		return "", fmt.Errorf("failed to broadcast invitation: %w", err)
	}

	return sessionID, nil
}

// JoinGroupEvent allows a participant to join an active group event
func (gem *GroupEventManager) JoinGroupEvent(sessionID, participantID string) error {
	gem.mu.Lock()
	defer gem.mu.Unlock()

	event, exists := gem.activeEvents[sessionID]
	if !exists {
		return fmt.Errorf("group event not found: %s", sessionID)
	}

	// Check if already participant
	if gem.participants[sessionID][participantID] {
		return fmt.Errorf("already participant in event: %s", sessionID)
	}

	// Check participant limits
	if len(event.Participants) >= event.Template.MaxParticipants {
		return fmt.Errorf("event full: %d/%d participants", 
			len(event.Participants), event.Template.MaxParticipants)
	}

	// Add participant
	event.Participants = append(event.Participants, participantID)
	gem.participants[sessionID][participantID] = true
	event.LastActivity = time.Now()

	// Broadcast join message
	joinMessage := GroupEventMessage{
		Type:      "join",
		SessionID: sessionID,
		Sender:    participantID,
		Data: map[string]interface{}{
			"participantCount": len(event.Participants),
			"participants":     event.Participants,
		},
		Timestamp: time.Now(),
	}

	return gem.broadcastGroupEventMessage(joinMessage)
}

// SubmitVote allows participants to vote on choices in the current phase
func (gem *GroupEventManager) SubmitVote(sessionID, participantID, choiceID string) error {
	gem.mu.Lock()
	defer gem.mu.Unlock()

	event, exists := gem.activeEvents[sessionID]
	if !exists {
		return fmt.Errorf("group event not found: %s", sessionID)
	}

	// Verify participant
	if !gem.participants[sessionID][participantID] {
		return fmt.Errorf("not a participant in event: %s", sessionID)
	}

	// Find current phase
	var currentPhase *EventPhase
	for _, phase := range event.Template.Phases {
		if phase.Name == event.CurrentPhase {
			currentPhase = &phase
			break
		}
	}
	if currentPhase == nil {
		return fmt.Errorf("invalid phase: %s", event.CurrentPhase)
	}

	// Validate choice ID
	validChoice := false
	for _, choice := range currentPhase.Choices {
		if choice.ID == choiceID {
			validChoice = true
			break
		}
	}
	if !validChoice {
		return fmt.Errorf("invalid choice: %s", choiceID)
	}

	// Initialize vote count if not exists
	if _, exists := event.Votes[choiceID]; !exists {
		event.Votes[choiceID] = 0
	}

	// Record vote
	if previousChoice, existed := event.ParticipantVotes[participantID]; existed {
		if event.Votes[previousChoice] > 0 {
			event.Votes[previousChoice]-- // Remove previous vote
		}
	}
	event.ParticipantVotes[participantID] = choiceID
	event.Votes[choiceID]++
	event.LastActivity = time.Now()

	// Check if we can advance (all participants voted or minimum reached)
	totalVotes := len(event.ParticipantVotes)
	canAdvance := totalVotes >= currentPhase.MinVotes || 
		(currentPhase.AutoAdvance && totalVotes == len(event.Participants))

	voteMessage := GroupEventMessage{
		Type:      "vote",
		SessionID: sessionID,
		Sender:    participantID,
		Data: map[string]interface{}{
			"choiceId":   choiceID,
			"totalVotes": totalVotes,
			"canAdvance": canAdvance,
			"votes":      event.Votes,
		},
		Timestamp: time.Now(),
	}

	if err := gem.broadcastGroupEventMessage(voteMessage); err != nil {
		return fmt.Errorf("failed to broadcast vote: %w", err)
	}

	// Auto-advance if conditions met
	if canAdvance {
		return gem.advancePhase(sessionID)
	}

	return nil
}

// advancePhase moves the event to the next phase
// Internal method - no mutex lock as called from locked contexts
func (gem *GroupEventManager) advancePhase(sessionID string) error {
	event := gem.activeEvents[sessionID]
	if event == nil {
		return fmt.Errorf("event not found: %s", sessionID)
	}

	// Find current phase index
	currentPhaseIndex := -1
	for i, phase := range event.Template.Phases {
		if phase.Name == event.CurrentPhase {
			currentPhaseIndex = i
			break
		}
	}

	if currentPhaseIndex == -1 {
		return fmt.Errorf("current phase not found: %s", event.CurrentPhase)
	}

	// Mark current phase as completed
	event.CompletedPhases = append(event.CompletedPhases, event.CurrentPhase)

	// Calculate scores for completed phase
	gem.calculatePhaseScores(event)

	// Check if event is complete
	if currentPhaseIndex >= len(event.Template.Phases)-1 {
		return gem.completeEvent(sessionID)
	}

	// Advance to next phase
	nextPhase := event.Template.Phases[currentPhaseIndex+1]
	event.CurrentPhase = nextPhase.Name
	event.Votes = make(map[string]int)
	event.ParticipantVotes = make(map[string]string)
	event.LastActivity = time.Now()

	// Broadcast phase advance
	advanceMessage := GroupEventMessage{
		Type:      "advance",
		SessionID: sessionID,
		Sender:    event.InitiatorID,
		Data: map[string]interface{}{
			"newPhase":        nextPhase.Name,
			"phaseDescription": nextPhase.Description,
			"choices":         nextPhase.Choices,
			"scores":          event.Scores,
		},
		Timestamp: time.Now(),
	}

	return gem.broadcastGroupEventMessage(advanceMessage)
}

// calculatePhaseScores awards points based on current phase results
// Internal scoring logic - uses simple point allocation
func (gem *GroupEventManager) calculatePhaseScores(event *GroupEvent) {
	// Award points based on votes
	for participantID, choiceID := range event.ParticipantVotes {
		// Find choice points
		for _, phase := range event.Template.Phases {
			if phase.Name == event.CurrentPhase {
				for _, choice := range phase.Choices {
					if choice.ID == choiceID {
						event.Scores[participantID] += choice.Points
						break
					}
				}
				break
			}
		}
	}
}

// completeEvent finishes the group event and records history
func (gem *GroupEventManager) completeEvent(sessionID string) error {
	event := gem.activeEvents[sessionID]
	if event == nil {
		return fmt.Errorf("event not found: %s", sessionID)
	}

	// Create completion record
	completedEvent := CompletedEvent{
		SessionID:    sessionID,
		TemplateID:   event.Template.ID,
		Participants: event.Participants,
		CompletedAt:  time.Now(),
		Duration:     time.Since(event.StartTime),
	}

	// Record in participant history
	for _, participantID := range event.Participants {
		completedEvent.FinalScore = event.Scores[participantID]
		gem.eventHistory[participantID] = append(gem.eventHistory[participantID], completedEvent)
	}

	// Broadcast completion
	endMessage := GroupEventMessage{
		Type:      "end",
		SessionID: sessionID,
		Sender:    event.InitiatorID,
		Data: map[string]interface{}{
			"finalScores": event.Scores,
			"duration":    completedEvent.Duration.String(),
			"completed":   true,
		},
		Timestamp: time.Now(),
	}

	if err := gem.broadcastGroupEventMessage(endMessage); err != nil {
		return fmt.Errorf("failed to broadcast completion: %w", err)
	}

	// Clean up active event
	delete(gem.activeEvents, sessionID)
	delete(gem.participants, sessionID)

	return nil
}

// GetActiveEvents returns list of currently active group events
func (gem *GroupEventManager) GetActiveEvents() []GroupEvent {
	gem.mu.RLock()
	defer gem.mu.RUnlock()

	events := make([]GroupEvent, 0, len(gem.activeEvents))
	for _, event := range gem.activeEvents {
		events = append(events, *event)
	}
	return events
}

// GetEventTemplates returns available group event templates
func (gem *GroupEventManager) GetEventTemplates() []GroupEventTemplate {
	gem.mu.RLock()
	defer gem.mu.RUnlock()

	return gem.eventTemplates
}

// GetParticipantHistory returns completed events for a participant
func (gem *GroupEventManager) GetParticipantHistory(participantID string) []CompletedEvent {
	gem.mu.RLock()
	defer gem.mu.RUnlock()

	history, exists := gem.eventHistory[participantID]
	if !exists {
		return []CompletedEvent{}
	}
	return history
}

// handleGroupEventMessage processes incoming network messages
// Uses Go error conventions for network message handling
func (gem *GroupEventManager) handleGroupEventMessage(data []byte, senderID string) error {
	var message GroupEventMessage
	if err := json.Unmarshal(data, &message); err != nil {
		return fmt.Errorf("failed to unmarshal group event message: %w", err)
	}

	switch message.Type {
	case "invite":
		// Handle invitation (could trigger UI notification)
		return gem.handleInvitation(message, senderID)
	case "join":
		// Handle participant joining
		return gem.handleParticipantJoin(message, senderID)
	case "vote":
		// Handle vote submission
		return gem.handleVoteSubmission(message, senderID)
	case "advance":
		// Handle phase advancement
		return gem.handlePhaseAdvance(message, senderID)
	case "end":
		// Handle event completion
		return gem.handleEventCompletion(message, senderID)
	default:
		return fmt.Errorf("unknown group event message type: %s", message.Type)
	}
}

// handleInvitation processes group event invitations
func (gem *GroupEventManager) handleInvitation(message GroupEventMessage, senderID string) error {
	// For now, just log the invitation
	// In a full implementation, this would trigger UI notifications
	templateName, _ := message.Data["templateName"].(string)
	fmt.Printf("Received group event invitation: %s from %s\n", templateName, senderID)
	return nil
}

// handleParticipantJoin processes participant join notifications
func (gem *GroupEventManager) handleParticipantJoin(message GroupEventMessage, senderID string) error {
	gem.mu.Lock()
	defer gem.mu.Unlock()

	if event, exists := gem.activeEvents[message.SessionID]; exists {
		event.LastActivity = time.Now()
		if participantCount, ok := message.Data["participantCount"].(float64); ok {
			fmt.Printf("Group event %s now has %d participants\n", message.SessionID, int(participantCount))
		}
	}
	return nil
}

// handleVoteSubmission processes vote notifications
func (gem *GroupEventManager) handleVoteSubmission(message GroupEventMessage, senderID string) error {
	gem.mu.Lock()
	defer gem.mu.Unlock()

	if event, exists := gem.activeEvents[message.SessionID]; exists {
		event.LastActivity = time.Now()
		if choiceID, ok := message.Data["choiceId"].(string); ok {
			fmt.Printf("Vote submitted in %s: %s chose %s\n", message.SessionID, senderID, choiceID)
		}
	}
	return nil
}

// handlePhaseAdvance processes phase advancement notifications
func (gem *GroupEventManager) handlePhaseAdvance(message GroupEventMessage, senderID string) error {
	gem.mu.Lock()
	defer gem.mu.Unlock()

	if event, exists := gem.activeEvents[message.SessionID]; exists {
		if newPhase, ok := message.Data["newPhase"].(string); ok {
			event.CurrentPhase = newPhase
			event.LastActivity = time.Now()
			fmt.Printf("Group event %s advanced to phase: %s\n", message.SessionID, newPhase)
		}
	}
	return nil
}

// handleEventCompletion processes event completion notifications
func (gem *GroupEventManager) handleEventCompletion(message GroupEventMessage, senderID string) error {
	gem.mu.Lock()
	defer gem.mu.Unlock()

	if _, exists := gem.activeEvents[message.SessionID]; exists {
		delete(gem.activeEvents, message.SessionID)
		delete(gem.participants, message.SessionID)
		
		if duration, ok := message.Data["duration"].(string); ok {
			fmt.Printf("Group event %s completed in %s\n", message.SessionID, duration)
		}
	}
	return nil
}

// broadcastGroupEventMessage sends a message to all connected peers
// Uses network manager's broadcast functionality
func (gem *GroupEventManager) broadcastGroupEventMessage(message GroupEventMessage) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal group event message: %w", err)
	}

	return gem.networkManager.BroadcastMessage("group_event", data)
}
