package dialog

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/jdkato/prose/v2"
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

	allResponses := n.combineResponses(localResponse, peerResponses)

	switch priority {
	case "first":
		return n.selectFirstValidResponse(allResponses)
	case "personality":
		return n.selectByPersonality(context, localResponse, peerResponses)
	case "random":
		return n.selectRandomResponse(allResponses)
	case "confidence":
		return n.selectHighestConfidenceResponse(localResponse, peerResponses)
	default:
		return n.selectHighestConfidenceResponse(localResponse, peerResponses)
	}
}

// combineResponses creates a unified list of all available responses
func (n *NetworkDialogBackend) combineResponses(localResponse DialogResponse, peerResponses []PeerDialogResponse) []DialogResponse {
	allResponses := []DialogResponse{localResponse}
	for _, peer := range peerResponses {
		allResponses = append(allResponses, peer.Response)
	}
	return allResponses
}

// selectFirstValidResponse returns the first response with non-empty text
func (n *NetworkDialogBackend) selectFirstValidResponse(allResponses []DialogResponse) DialogResponse {
	for _, response := range allResponses {
		if response.Text != "" {
			return response
		}
	}
	// Fallback to first response even if empty
	if len(allResponses) > 0 {
		return allResponses[0]
	}
	return DialogResponse{}
}

// selectRandomResponse chooses randomly from valid responses
func (n *NetworkDialogBackend) selectRandomResponse(allResponses []DialogResponse) DialogResponse {
	validResponses := n.filterValidResponses(allResponses)
	if len(validResponses) > 0 {
		return validResponses[rand.Intn(len(validResponses))]
	}
	// Fallback to first response if no valid ones
	if len(allResponses) > 0 {
		return allResponses[0]
	}
	return DialogResponse{}
}

// filterValidResponses returns only responses with non-empty text
func (n *NetworkDialogBackend) filterValidResponses(responses []DialogResponse) []DialogResponse {
	validResponses := make([]DialogResponse, 0)
	for _, response := range responses {
		if response.Text != "" {
			validResponses = append(validResponses, response)
		}
	}
	return validResponses
}

// selectHighestConfidenceResponse chooses the response with the highest confidence score
func (n *NetworkDialogBackend) selectHighestConfidenceResponse(localResponse DialogResponse, peerResponses []PeerDialogResponse) DialogResponse {
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

// selectByPersonality selects response based on personality compatibility
func (n *NetworkDialogBackend) selectByPersonality(context DialogContext, localResponse DialogResponse, peerResponses []PeerDialogResponse) DialogResponse {
	// Advanced personality-based selection with multi-factor scoring
	bestResponse := localResponse
	bestScore := n.calculatePersonalityScore(context, localResponse, "")

	// Evaluate peer responses with personality compatibility scoring
	for _, peer := range peerResponses {
		score := n.calculatePersonalityScore(context, peer.Response, peer.PeerID)

		// Add conversation flow scoring to avoid repetitive selections
		score += n.calculateConversationFlowScore(context, peer.Response)

		// Add emotional tone matching using NLP analysis
		score += n.calculateEmotionalToneScore(context, peer.Response)

		if score > bestScore {
			bestScore = score
			bestResponse = peer.Response
		}
	}

	return bestResponse
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

// calculatePersonalityScore computes personality compatibility score for a response
func (n *NetworkDialogBackend) calculatePersonalityScore(context DialogContext, response DialogResponse, peerID string) float64 {
	score := 0.0

	// Define trait weights for personality scoring
	traitWeights := map[string]float64{
		"shyness":     0.25,
		"openness":    0.20,
		"chattiness":  0.15,
		"empathy":     0.15,
		"creativity":  0.10,
		"playfulness": 0.10,
		"enthusiasm":  0.05,
	}

	// Calculate compatibility based on personality traits
	for trait, weight := range traitWeights {
		if value, exists := context.PersonalityTraits[trait]; exists {
			// Score based on response characteristics matching personality
			traitScore := n.scoreResponseForTrait(response, trait, value)
			score += traitScore * weight
		}
	}

	return score
}

// scoreResponseForTrait evaluates how well a response matches a personality trait
func (n *NetworkDialogBackend) scoreResponseForTrait(response DialogResponse, trait string, traitValue float64) float64 {
	text := response.Text
	textLen := float64(len(text))

	switch trait {
	case "shyness":
		// Shy characters prefer shorter, less assertive responses
		if traitValue > 0.7 {
			return 1.0 - (textLen / 200.0) // Prefer shorter responses
		} else if traitValue < 0.3 {
			return textLen / 200.0 // Prefer longer responses
		}
		return 0.5

	case "chattiness":
		// Chatty characters prefer longer responses
		if traitValue > 0.7 {
			return textLen / 150.0
		} else if traitValue < 0.3 {
			return 1.0 - (textLen / 150.0)
		}
		return 0.5

	case "empathy":
		// Empathetic characters prefer responses with emotional words
		emotionalWords := []string{"feel", "understand", "care", "love", "hurt", "happy", "sad"}
		lowerText := strings.ToLower(text)
		emotionalCount := 0
		for _, word := range emotionalWords {
			if strings.Contains(lowerText, word) {
				emotionalCount++
			}
		}
		if traitValue > 0.7 {
			return float64(emotionalCount) / 3.0 // Max score for 3+ emotional words
		}
		return 0.5

	case "playfulness":
		// Playful characters prefer responses with exclamation marks, humor indicators
		playfulIndicators := []string{"!", "haha", "lol", "fun", "play", "silly"}
		lowerText := strings.ToLower(text)
		playfulCount := 0
		for _, indicator := range playfulIndicators {
			if strings.Contains(lowerText, indicator) {
				playfulCount++
			}
		}
		if traitValue > 0.7 {
			return float64(playfulCount) / 2.0 // Max score for 2+ playful indicators
		}
		return 0.5

	default:
		return 0.5 // Neutral score for unknown traits
	}
}

// calculateConversationFlowScore evaluates response diversity and turn-taking
func (n *NetworkDialogBackend) calculateConversationFlowScore(context DialogContext, response DialogResponse) float64 {
	// Basic conversation flow scoring to avoid repetitive responses
	// Check if this response is too similar to recent responses

	if len(context.InteractionHistory) == 0 {
		return 0.5 // Neutral score for first response
	}

	// Check similarity with recent responses (simple length and keyword comparison)
	recentCount := len(context.InteractionHistory)
	if recentCount > 3 {
		recentCount = 3 // Only check last 3 interactions
	}

	similarityPenalty := 0.0
	for i := len(context.InteractionHistory) - recentCount; i < len(context.InteractionHistory); i++ {
		recent := context.InteractionHistory[i].Response // Get the response from interaction record

		// Length similarity penalty
		lengthDiff := float64(abs(len(response.Text) - len(recent)))
		if lengthDiff < 10 {
			similarityPenalty += 0.2
		}

		// Simple keyword overlap penalty
		if n.hasSignificantOverlap(response.Text, recent) {
			similarityPenalty += 0.3
		}
	}

	return 1.0 - similarityPenalty // Higher score for more diverse responses
}

// calculateEmotionalToneScore uses NLP analysis for emotional tone matching
func (n *NetworkDialogBackend) calculateEmotionalToneScore(context DialogContext, response DialogResponse) float64 {
	// Use prose NLP library for sentiment analysis
	doc, err := prose.NewDocument(response.Text)
	if err != nil {
		return 0.5 // Neutral score on NLP error
	}

	// Analyze entities and part-of-speech for emotional content
	entities := doc.Entities()
	tokens := doc.Tokens()

	emotionalScore := 0.0

	// Count emotional indicators
	emotionalPOSTags := []string{"JJ", "RB", "VB"} // Adjectives, adverbs, verbs
	emotionalCount := 0

	for _, token := range tokens {
		for _, tag := range emotionalPOSTags {
			if strings.HasPrefix(token.Tag, tag) {
				emotionalCount++
				break
			}
		}
	}

	// Score based on emotional richness
	if emotionalCount > 0 {
		emotionalScore = float64(emotionalCount) / float64(len(tokens))
		if emotionalScore > 1.0 {
			emotionalScore = 1.0
		}
	}

	// Bonus for named entities (more engaging responses)
	if len(entities) > 0 {
		emotionalScore += 0.2
	}

	return emotionalScore
}

// hasSignificantOverlap checks if two text strings have significant word overlap
func (n *NetworkDialogBackend) hasSignificantOverlap(text1, text2 string) bool {
	words1 := strings.Fields(strings.ToLower(text1))
	words2 := strings.Fields(strings.ToLower(text2))

	if len(words1) == 0 || len(words2) == 0 {
		return false
	}

	overlap := 0
	for _, word1 := range words1 {
		for _, word2 := range words2 {
			if word1 == word2 && len(word1) > 3 { // Only count meaningful words
				overlap++
				break
			}
		}
	}

	// Significant overlap if more than 30% of words match
	return float64(overlap)/float64(len(words1)) > 0.3
}

// abs returns absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
