package dialog

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// NetworkDialogBackend implements DialogBackend for multiplayer coordination.
// It coordinates responses with peer characters and handles network-aware dialog generation.
type NetworkDialogBackend struct {
	mu     sync.RWMutex
	config NetworkDialogConfig

	// Dialog coordination
	localBackend  DialogBackend // Fallback to local dialog generation
	responseCache map[string]CachedResponse
	peerResponses map[string]PeerDialogState

	// Network coordination interface
	networkCoordinator NetworkCoordinator

	// Response timing
	coordinationTimeout time.Duration
	cacheExpiry         time.Duration
	lastCoordination    time.Time
}

// NetworkDialogConfig defines configuration for network dialog coordination
type NetworkDialogConfig struct {
	Type                string          `json:"type"`                // "network"
	LocalBackendType    string          `json:"localBackendType"`    // "simple_random", "markov", etc.
	LocalBackendConfig  json.RawMessage `json:"localBackendConfig"`  // Configuration for local backend
	CoordinationTimeout string          `json:"coordinationTimeout"` // Max time to wait for peer responses (e.g. "500ms")
	EnableGroupDialogs  bool            `json:"enableGroupDialogs"`  // Allow multi-character conversations
	ResponsePriority    string          `json:"responsePriority"`    // "first", "personality", "random", "confidence"
	CacheExpiry         string          `json:"cacheExpiry"`         // How long to cache responses (e.g. "5m")
	MinPeersForGroup    int             `json:"minPeersForGroup"`    // Minimum peers needed for group dialogs
}

// CachedResponse stores a dialog response with expiry
type CachedResponse struct {
	Response  DialogResponse `json:"response"`
	ExpiresAt time.Time      `json:"expiresAt"`
	Source    string         `json:"source"` // "local", "peer", "coordinated"
}

// PeerDialogState tracks peer dialog capabilities and recent responses
type PeerDialogState struct {
	PeerID            string    `json:"peerId"`
	CanGenerateDialog bool      `json:"canGenerateDialog"`
	LastResponse      string    `json:"lastResponse"`
	LastResponseTime  time.Time `json:"lastResponseTime"`
	PersonalityType   string    `json:"personalityType"`
	ResponseStyle     string    `json:"responseStyle"` // "talkative", "quiet", "supportive", etc.
}

// NetworkCoordinator interface for network communication
// This allows for dependency injection and testing
type NetworkCoordinator interface {
	// RequestPeerDialogs asks connected peers for dialog suggestions
	RequestPeerDialogs(context DialogContext) ([]PeerDialogResponse, error)

	// BroadcastDialogResponse informs peers of our response choice
	BroadcastDialogResponse(context DialogContext, response DialogResponse) error

	// GetConnectedPeers returns currently connected peer information
	GetConnectedPeers() []PeerInfo

	// IsNetworkAvailable checks if network coordination is possible
	IsNetworkAvailable() bool
}

// PeerDialogResponse represents a dialog suggestion from a peer
type PeerDialogResponse struct {
	PeerID     string         `json:"peerId"`
	Response   DialogResponse `json:"response"`
	Confidence float64        `json:"confidence"`
	Timestamp  time.Time      `json:"timestamp"`
}

// PeerInfo contains basic information about a connected peer
type PeerInfo struct {
	ID              string `json:"id"`
	CharacterType   string `json:"characterType"`
	PersonalityType string `json:"personalityType"`
	IsBot           bool   `json:"isBot"`
}

// NewNetworkDialogBackend creates a new network-aware dialog backend
func NewNetworkDialogBackend() *NetworkDialogBackend {
	return &NetworkDialogBackend{
		responseCache:       make(map[string]CachedResponse),
		peerResponses:       make(map[string]PeerDialogState),
		coordinationTimeout: 500 * time.Millisecond, // Default 500ms timeout
	}
}

// Initialize sets up the network dialog backend with configuration
func (n *NetworkDialogBackend) Initialize(config json.RawMessage) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Set defaults
	n.config = NetworkDialogConfig{
		Type:                "network",
		LocalBackendType:    "simple_random",
		CoordinationTimeout: "500ms",
		EnableGroupDialogs:  true,
		ResponsePriority:    "personality",
		CacheExpiry:         "5m",
		MinPeersForGroup:    2,
	}

	// Parse configuration
	if len(config) > 0 {
		if err := json.Unmarshal(config, &n.config); err != nil {
			return fmt.Errorf("failed to parse network dialog config: %w", err)
		}
	}

	// Parse coordination timeout
	if timeout, err := time.ParseDuration(n.config.CoordinationTimeout); err != nil {
		return fmt.Errorf("invalid coordination timeout: %w", err)
	} else {
		n.coordinationTimeout = timeout
	}

	// Parse cache expiry
	if expiry, err := time.ParseDuration(n.config.CacheExpiry); err != nil {
		return fmt.Errorf("invalid cache expiry: %w", err)
	} else {
		n.cacheExpiry = expiry
	}

	// Initialize local backend as fallback
	if err := n.initializeLocalBackend(); err != nil {
		return fmt.Errorf("failed to initialize local backend: %w", err)
	}

	return nil
}

// initializeLocalBackend sets up the fallback local dialog backend
func (n *NetworkDialogBackend) initializeLocalBackend() error {
	switch n.config.LocalBackendType {
	case "simple_random":
		n.localBackend = NewSimpleRandomBackend()
	case "markov":
		n.localBackend = NewMarkovChainBackend()
	default:
		return fmt.Errorf("unsupported local backend type: %s", n.config.LocalBackendType)
	}

	return n.localBackend.Initialize(n.config.LocalBackendConfig)
}

// GenerateResponse produces a dialog response using network coordination
func (n *NetworkDialogBackend) GenerateResponse(context DialogContext) (DialogResponse, error) {
	// Check cache first
	if cachedResponse, found := n.getCachedResponse(context); found {
		return cachedResponse.Response, nil
	}

	// If network coordination is available and enabled, try it
	if n.shouldUseNetworkCoordination(context) {
		if response, err := n.generateCoordinatedResponse(context); err == nil {
			n.cacheResponse(context, response, "coordinated")
			return response, nil
		}
		// Network coordination failed, fall back to local
	}

	// Use local backend as fallback
	response, err := n.localBackend.GenerateResponse(context)
	if err != nil {
		return n.generateFallbackResponse(context), nil
	}

	n.cacheResponse(context, response, "local")
	return response, nil
}

// shouldUseNetworkCoordination determines if network coordination should be attempted
func (n *NetworkDialogBackend) shouldUseNetworkCoordination(context DialogContext) bool {
	n.mu.RLock()
	defer n.mu.RUnlock()

	// Check if network coordinator is available
	if n.networkCoordinator == nil || !n.networkCoordinator.IsNetworkAvailable() {
		return false
	}

	// Don't coordinate too frequently to avoid spam
	if time.Since(n.lastCoordination) < 1*time.Second {
		return false
	}

	// Check if we have enough peers for group coordination
	if n.config.EnableGroupDialogs {
		peers := n.networkCoordinator.GetConnectedPeers()
		return len(peers) >= n.config.MinPeersForGroup
	}

	return true
}

// generateCoordinatedResponse coordinates with peers to generate a response
func (n *NetworkDialogBackend) generateCoordinatedResponse(context DialogContext) (DialogResponse, error) {
	n.mu.Lock()
	n.lastCoordination = time.Now()
	n.mu.Unlock()

	// Request peer dialog suggestions
	peerResponses, err := n.networkCoordinator.RequestPeerDialogs(context)
	if err != nil {
		return DialogResponse{}, err
	}

	// Generate local response as well
	localResponse, err := n.localBackend.GenerateResponse(context)
	if err != nil {
		return DialogResponse{}, err
	}

	// Select best response based on configuration
	selectedResponse := n.selectBestResponse(context, localResponse, peerResponses)

	// Broadcast our chosen response to peers
	if err := n.networkCoordinator.BroadcastDialogResponse(context, selectedResponse); err != nil {
		// Log error but don't fail - response is still valid
	}

	return selectedResponse, nil
}

// selectBestResponse chooses the best response from local and peer suggestions
func (n *NetworkDialogBackend) selectBestResponse(context DialogContext, localResponse DialogResponse, peerResponses []PeerDialogResponse) DialogResponse {
	n.mu.RLock()
	priority := n.config.ResponsePriority
	n.mu.RUnlock()

	allResponses := []DialogResponse{localResponse}
	for _, peer := range peerResponses {
		allResponses = append(allResponses, peer.Response)
	}

	switch priority {
	case "first":
		// Return first valid response
		for _, response := range allResponses {
			if response.Text != "" {
				return response
			}
		}

	case "personality":
		// Select based on personality compatibility
		return n.selectByPersonality(context, localResponse, peerResponses)

	case "random":
		// Random selection from valid responses
		validResponses := make([]DialogResponse, 0)
		for _, response := range allResponses {
			if response.Text != "" {
				validResponses = append(validResponses, response)
			}
		}
		if len(validResponses) > 0 {
			return validResponses[rand.Intn(len(validResponses))]
		}

	case "confidence":
		// Select highest confidence response
		bestResponse := localResponse
		bestConfidence := localResponse.Confidence

		for _, peer := range peerResponses {
			if peer.Confidence > bestConfidence {
				bestResponse = peer.Response
				bestConfidence = peer.Confidence
			}
		}
		return bestResponse

	default:
		// Default to highest confidence
		bestResponse := localResponse
		bestConfidence := localResponse.Confidence

		for _, peer := range peerResponses {
			if peer.Confidence > bestConfidence {
				bestResponse = peer.Response
				bestConfidence = peer.Confidence
			}
		}
		return bestResponse
	}

	return localResponse // Fallback to local response
}

// selectByPersonality selects response based on personality compatibility
func (n *NetworkDialogBackend) selectByPersonality(context DialogContext, localResponse DialogResponse, peerResponses []PeerDialogResponse) DialogResponse {
	// Simple personality-based selection
	// In a full implementation, this would consider:
	// - Character personality traits from context
	// - Peer personality types
	// - Response emotional tone matching
	// - Conversation flow and turn-taking

	// For now, prefer responses that match the character's personality
	if shyness, exists := context.PersonalityTraits["shyness"]; exists {
		if shyness > 0.7 {
			// Shy characters prefer shorter, less bold responses
			for _, peer := range peerResponses {
				if len(peer.Response.Text) < len(localResponse.Text) {
					return peer.Response
				}
			}
		} else if shyness < 0.3 {
			// Bold characters prefer longer, more expressive responses
			for _, peer := range peerResponses {
				if len(peer.Response.Text) > len(localResponse.Text) {
					return peer.Response
				}
			}
		}
	}

	return localResponse // Default to local response
}

// getCachedResponse retrieves a cached response if available and not expired
func (n *NetworkDialogBackend) getCachedResponse(context DialogContext) (CachedResponse, bool) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	key := n.generateCacheKey(context)
	cached, exists := n.responseCache[key]

	if !exists || time.Now().After(cached.ExpiresAt) {
		return CachedResponse{}, false
	}

	return cached, true
}

// cacheResponse stores a response in the cache
func (n *NetworkDialogBackend) cacheResponse(context DialogContext, response DialogResponse, source string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	key := n.generateCacheKey(context)
	n.responseCache[key] = CachedResponse{
		Response:  response,
		ExpiresAt: time.Now().Add(n.cacheExpiry),
		Source:    source,
	}

	// Clean up expired cache entries
	n.cleanupCache()
}

// generateCacheKey creates a cache key from dialog context
func (n *NetworkDialogBackend) generateCacheKey(context DialogContext) string {
	// Simple cache key based on trigger and basic context
	// In production, this might include more context elements
	return fmt.Sprintf("%s:%s:%d", context.Trigger, context.RelationshipLevel, context.ConversationTurn)
}

// cleanupCache removes expired entries from the response cache
func (n *NetworkDialogBackend) cleanupCache() {
	now := time.Now()
	for key, cached := range n.responseCache {
		if now.After(cached.ExpiresAt) {
			delete(n.responseCache, key)
		}
	}
}

// generateFallbackResponse creates a basic fallback response
func (n *NetworkDialogBackend) generateFallbackResponse(context DialogContext) DialogResponse {
	fallbackTexts := context.FallbackResponses
	if len(fallbackTexts) == 0 {
		fallbackTexts = []string{
			"I'm thinking...",
			"Hmm, let me consider that.",
			"That's interesting!",
			"I need a moment to process that.",
		}
	}

	text := fallbackTexts[rand.Intn(len(fallbackTexts))]
	animation := context.FallbackAnimation
	if animation == "" {
		animation = "idle"
	}

	return DialogResponse{
		Text:         text,
		Animation:    animation,
		Confidence:   0.1, // Low confidence for fallback
		ResponseType: "fallback",
		Topics:       []string{"system"},
	}
}

// SetNetworkCoordinator injects the network coordinator dependency
func (n *NetworkDialogBackend) SetNetworkCoordinator(coordinator NetworkCoordinator) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.networkCoordinator = coordinator
}

// GetBackendInfo returns metadata about this backend
func (n *NetworkDialogBackend) GetBackendInfo() BackendInfo {
	return BackendInfo{
		Name:        "network_dialog",
		Version:     "1.0.0",
		Description: "Network-aware dialog backend that coordinates responses with peer characters",
		Capabilities: []string{
			"peer_coordination",
			"group_dialogs",
			"response_caching",
			"fallback_support",
		},
		Author:  "DDS Multiplayer System",
		License: "MIT",
	}
}

// CanHandle checks if this backend can process the given context
func (n *NetworkDialogBackend) CanHandle(context DialogContext) bool {
	// Network backend can handle any context that the local backend can handle
	n.mu.RLock()
	localBackend := n.localBackend
	n.mu.RUnlock()

	if localBackend == nil {
		return true // Default to accepting all contexts
	}

	return localBackend.CanHandle(context)
}

// UpdateMemory updates both local and network memory systems
func (n *NetworkDialogBackend) UpdateMemory(context DialogContext, response DialogResponse, userFeedback *UserFeedback) error {
	// Update local backend memory
	n.mu.RLock()
	localBackend := n.localBackend
	n.mu.RUnlock()

	if localBackend != nil {
		if err := localBackend.UpdateMemory(context, response, userFeedback); err != nil {
			return fmt.Errorf("failed to update local memory: %w", err)
		}
	}

	// Update peer dialog state tracking
	n.updatePeerDialogState(context, response, userFeedback)

	return nil
}

// updatePeerDialogState updates tracking of peer dialog capabilities
func (n *NetworkDialogBackend) updatePeerDialogState(context DialogContext, response DialogResponse, userFeedback *UserFeedback) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Track our own dialog patterns for peer coordination
	// This helps peers understand our dialog style and preferences

	// Update last response tracking
	for peerID := range n.peerResponses {
		state := n.peerResponses[peerID]
		// Update peer state based on their recent activity
		// This is a simplified implementation - in practice would track more details
		n.peerResponses[peerID] = state
	}
}

// GetCacheStats returns statistics about the response cache
func (n *NetworkDialogBackend) GetCacheStats() map[string]interface{} {
	n.mu.RLock()
	defer n.mu.RUnlock()

	stats := map[string]interface{}{
		"cacheSize":        len(n.responseCache),
		"peerCount":        len(n.peerResponses),
		"cacheExpiry":      n.cacheExpiry.String(),
		"lastCoordination": n.lastCoordination,
	}

	return stats
}

// ClearCache removes all cached responses
func (n *NetworkDialogBackend) ClearCache() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.responseCache = make(map[string]CachedResponse)
}
