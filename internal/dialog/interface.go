package dialog

import (
	"encoding/json"
	"fmt"
	"time"
)

// DialogBackend defines the interface for pluggable dialog generation systems
// This interface matches the one in internal/character but provides clean separation
type DialogBackend interface {
	// Initialize sets up the backend with JSON configuration
	Initialize(config json.RawMessage) error

	// GenerateResponse produces a dialog response for the given context
	// Returns the response text and any animation to trigger
	GenerateResponse(context DialogContext) (DialogResponse, error)

	// GetBackendInfo returns metadata about this backend implementation
	GetBackendInfo() BackendInfo

	// CanHandle checks if this backend can process the given trigger/context
	CanHandle(context DialogContext) bool

	// UpdateMemory allows the backend to record interaction outcomes for learning
	// This enables backends to adapt based on user interactions
	UpdateMemory(context DialogContext, response DialogResponse, userFeedback *UserFeedback) error
}

// DialogContext provides complete context for dialog generation
type DialogContext struct {
	// Basic interaction details
	Trigger       string    `json:"trigger"`       // "click", "rightclick", "hover", etc.
	InteractionID string    `json:"interactionId"` // Unique identifier for this interaction
	Timestamp     time.Time `json:"timestamp"`

	// Character state context
	CurrentStats      map[string]float64 `json:"currentStats"`      // Current stat values
	PersonalityTraits map[string]float64 `json:"personalityTraits"` // Character personality
	CurrentMood       float64            `json:"currentMood"`       // Overall mood (0-100)
	CurrentAnimation  string             `json:"currentAnimation"`  // Current character state

	// Relationship/game context
	RelationshipLevel  string              `json:"relationshipLevel,omitempty"`  // Current relationship stage
	InteractionHistory []InteractionRecord `json:"interactionHistory,omitempty"` // Recent interactions
	AchievementStatus  map[string]bool     `json:"achievementStatus,omitempty"`  // Unlocked achievements
	TimeOfDay          string              `json:"timeOfDay,omitempty"`          // "morning", "afternoon", "evening", "night"

	// Conversation context
	LastResponse     string                 `json:"lastResponse,omitempty"` // Previous dialog response
	ConversationTurn int                    `json:"conversationTurn"`       // Turn number in current conversation
	TopicContext     map[string]interface{} `json:"topicContext,omitempty"` // Current conversation topics

	// Fallback configuration
	FallbackResponses []string `json:"fallbackResponses"` // Default responses if backend fails
	FallbackAnimation string   `json:"fallbackAnimation"` // Default animation if backend fails
}

// DialogResponse contains the generated response and associated metadata
type DialogResponse struct {
	// Response content
	Text      string `json:"text"`                // The dialog text to display
	Animation string `json:"animation,omitempty"` // Animation to trigger with response
	Duration  int    `json:"duration,omitempty"`  // Display duration in seconds (0 = default)

	// Response metadata
	Confidence    float64                `json:"confidence"`              // Backend confidence in response (0-1)
	ResponseType  string                 `json:"responseType,omitempty"`  // "casual", "romantic", "informative", etc.
	EmotionalTone string                 `json:"emotionalTone,omitempty"` // "happy", "sad", "flirty", "shy", etc.
	Topics        []string               `json:"topics,omitempty"`        // Topics covered in this response
	Metadata      map[string]interface{} `json:"metadata,omitempty"`      // Backend-specific metadata

	// Memory and learning
	MemoryImportance float64 `json:"memoryImportance,omitempty"` // How important is this for memory (0-1)
	LearningValue    float64 `json:"learningValue,omitempty"`    // Value for backend learning (0-1)
}

// UserFeedback captures user response to dialog for backend learning
type UserFeedback struct {
	Positive     bool                   `json:"positive"`               // Whether user responded positively
	ResponseTime time.Duration          `json:"responseTime"`           // Time until user's next interaction
	FollowUpType string                 `json:"followUpType,omitempty"` // Type of follow-up interaction
	Engagement   float64                `json:"engagement"`             // Engagement score (0-1)
	CustomData   map[string]interface{} `json:"customData,omitempty"`   // Backend-specific feedback data
}

// InteractionRecord captures a single interaction for context building
type InteractionRecord struct {
	Type      string             `json:"type"`     // "click", "feed", "compliment", etc.
	Response  string             `json:"response"` // What the character said
	Timestamp time.Time          `json:"timestamp"`
	Stats     map[string]float64 `json:"stats"`   // Stats at time of interaction
	Outcome   string             `json:"outcome"` // "positive", "negative", "neutral"
}

// BackendInfo provides metadata about a dialog backend
type BackendInfo struct {
	Name         string   `json:"name"`         // Backend name (e.g., "markov_chain", "rule_based")
	Version      string   `json:"version"`      // Backend version
	Description  string   `json:"description"`  // Human-readable description
	Capabilities []string `json:"capabilities"` // List of features supported
	Author       string   `json:"author"`       // Backend author/maintainer
	License      string   `json:"license"`      // License information
}

// DialogManager orchestrates multiple backends and handles fallbacks
type DialogManager struct {
	backends       map[string]DialogBackend
	defaultBackend string
	fallbackChain  []string
	debug          bool
}

// NewDialogManager creates a new dialog manager with no backends registered
func NewDialogManager(debug bool) *DialogManager {
	return &DialogManager{
		backends:      make(map[string]DialogBackend),
		fallbackChain: []string{},
		debug:         debug,
	}
}

// RegisterBackend adds a new dialog backend to the manager
func (dm *DialogManager) RegisterBackend(name string, backend DialogBackend) {
	dm.backends[name] = backend
}

// SetDefaultBackend sets the primary backend to use for dialog generation
func (dm *DialogManager) SetDefaultBackend(name string) error {
	if _, exists := dm.backends[name]; !exists {
		return fmt.Errorf("backend '%s' not registered", name)
	}
	dm.defaultBackend = name
	return nil
}

// SetFallbackChain configures the order of backends to try if primary fails
func (dm *DialogManager) SetFallbackChain(backends []string) error {
	for _, name := range backends {
		if _, exists := dm.backends[name]; !exists {
			return fmt.Errorf("fallback backend '%s' not registered", name)
		}
	}
	dm.fallbackChain = backends
	return nil
}

// GenerateDialog produces a dialog response using the configured backend chain
func (dm *DialogManager) GenerateDialog(context DialogContext) (DialogResponse, error) {
	// Attempt response generation using default backend first
	if response, success := dm.tryDefaultBackend(context); success {
		return response, nil
	}

	// Try fallback chain if default backend fails
	if response, success := dm.tryFallbackChain(context); success {
		return response, nil
	}

	// Final fallback: use provided fallback responses
	return dm.createFallbackResponse(context), nil
}

// tryDefaultBackend attempts to generate response using the configured default backend
func (dm *DialogManager) tryDefaultBackend(context DialogContext) (DialogResponse, bool) {
	if dm.defaultBackend == "" {
		return DialogResponse{}, false
	}

	backend, exists := dm.backends[dm.defaultBackend]
	if !exists {
		return DialogResponse{}, false
	}

	if !backend.CanHandle(context) {
		return DialogResponse{}, false
	}

	response, err := backend.GenerateResponse(context)
	if err != nil || response.Confidence <= 0.5 {
		return DialogResponse{}, false
	}

	return response, true
}

// tryFallbackChain attempts to generate response using the fallback backend chain
func (dm *DialogManager) tryFallbackChain(context DialogContext) (DialogResponse, bool) {
	for _, backendName := range dm.fallbackChain {
		if response, success := dm.tryFallbackBackend(backendName, context); success {
			return response, true
		}
	}
	return DialogResponse{}, false
}

// tryFallbackBackend attempts to generate response using a specific fallback backend
func (dm *DialogManager) tryFallbackBackend(backendName string, context DialogContext) (DialogResponse, bool) {
	backend, exists := dm.backends[backendName]
	if !exists {
		return DialogResponse{}, false
	}

	if !backend.CanHandle(context) {
		return DialogResponse{}, false
	}

	response, err := backend.GenerateResponse(context)
	if err != nil {
		return DialogResponse{}, false
	}

	return response, true
}

// createFallbackResponse generates a basic response when all backends fail
func (dm *DialogManager) createFallbackResponse(context DialogContext) DialogResponse {
	response := "Hello! 👋"
	animation := "talking"

	if len(context.FallbackResponses) > 0 {
		// Simple time-based selection for fallback
		index := int(time.Now().UnixNano()) % len(context.FallbackResponses)
		response = context.FallbackResponses[index]
	}

	if context.FallbackAnimation != "" {
		animation = context.FallbackAnimation
	}

	return DialogResponse{
		Text:         response,
		Animation:    animation,
		Confidence:   0.1, // Very low confidence for fallback
		ResponseType: "fallback",
	}
}

// GetRegisteredBackends returns a list of all registered backend names
func (dm *DialogManager) GetRegisteredBackends() []string {
	names := make([]string, 0, len(dm.backends))
	for name := range dm.backends {
		names = append(names, name)
	}
	return names
}

// GetBackendInfo returns information about a specific backend
func (dm *DialogManager) GetBackendInfo(name string) (BackendInfo, error) {
	backend, exists := dm.backends[name]
	if !exists {
		return BackendInfo{}, fmt.Errorf("backend '%s' not found", name)
	}
	return backend.GetBackendInfo(), nil
}

// UpdateBackendMemory records interaction outcomes for backend learning
func (dm *DialogManager) UpdateBackendMemory(context DialogContext, response DialogResponse, feedback *UserFeedback) {
	// Update memory for the backend that generated this response
	for _, backend := range dm.backends {
		if backend.CanHandle(context) {
			_ = backend.UpdateMemory(context, response, feedback)
			break
		}
	}
}

// GetBackend returns a specific registered backend by name
func (dm *DialogManager) GetBackend(name string) (DialogBackend, bool) {
	backend, exists := dm.backends[name]
	return backend, exists
}

// DialogBackendConfig represents JSON configuration for dialog backends
type DialogBackendConfig struct {
	// Backend selection
	DefaultBackend string   `json:"defaultBackend"`          // Primary backend to use
	FallbackChain  []string `json:"fallbackChain,omitempty"` // Ordered list of fallback backends
	Enabled        bool     `json:"enabled"`                 // Whether to use advanced dialog system

	// Backend-specific configurations
	Backends map[string]json.RawMessage `json:"backends,omitempty"` // Backend-specific config

	// Global settings
	MemoryEnabled       bool    `json:"memoryEnabled"`             // Enable interaction memory
	LearningEnabled     bool    `json:"learningEnabled"`           // Enable backend learning
	ConfidenceThreshold float64 `json:"confidenceThreshold"`       // Minimum confidence to accept response
	ResponseTimeout     int     `json:"responseTimeout,omitempty"` // Max time to wait for response (ms)
	DebugMode           bool    `json:"debugMode,omitempty"`       // Enable debug logging
}

// ValidateBackendConfig ensures the backend configuration is valid
func ValidateBackendConfig(config DialogBackendConfig) error {
	if !config.Enabled {
		return nil // Skip validation if disabled
	}

	if config.DefaultBackend == "" {
		return fmt.Errorf("defaultBackend is required when dialog system is enabled")
	}

	if config.ConfidenceThreshold < 0 || config.ConfidenceThreshold > 1 {
		return fmt.Errorf("confidenceThreshold must be between 0 and 1, got %f", config.ConfidenceThreshold)
	}

	if config.ResponseTimeout < 0 {
		return fmt.Errorf("responseTimeout must be non-negative, got %d", config.ResponseTimeout)
	}

	return nil
}

// LoadDialogBackendConfig loads backend configuration from JSON
func LoadDialogBackendConfig(data []byte) (DialogBackendConfig, error) {
	var config DialogBackendConfig

	// Set defaults
	config.ConfidenceThreshold = 0.5
	config.ResponseTimeout = 1000
	config.MemoryEnabled = true
	config.LearningEnabled = false

	if err := json.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("failed to parse dialog backend config: %w", err)
	}

	if err := ValidateBackendConfig(config); err != nil {
		return config, fmt.Errorf("invalid dialog backend config: %w", err)
	}

	return config, nil
}
