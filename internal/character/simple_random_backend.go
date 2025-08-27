package character

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// SimpleRandomBackend implements DialogBackend using existing dialog selection logic
// Provides 1:1 compatibility with the existing system while adding dialog interface compliance
type SimpleRandomBackend struct {
	config    SimpleRandomConfig
	character *Character
}

// SimpleRandomConfig defines JSON configuration for the simple random backend
type SimpleRandomConfig struct {
	Type                 string   `json:"type"`                        // "basic"
	PersonalityInfluence float64  `json:"personalityInfluence"`        // 0-1, how much personality affects selection
	UseDialogHistory     bool     `json:"useDialogHistory"`            // Whether to consider recent dialog history
	ResponseVariation    float64  `json:"responseVariation"`           // 0-1, adds variety to selection
	PreferRomanceDialogs bool     `json:"preferRomanceDialogs"`        // Prefer romance dialogs when available
	FallbackResponses    []string `json:"fallbackResponses,omitempty"` // Custom fallback responses
}

// NewSimpleRandomBackend creates a new simple random dialog backend
func NewSimpleRandomBackend() *SimpleRandomBackend {
	return &SimpleRandomBackend{}
}

// Initialize sets up the simple random backend with JSON configuration
func (s *SimpleRandomBackend) Initialize(config json.RawMessage, character *Character) error {
	// Set defaults
	s.config = SimpleRandomConfig{
		Type:                 "basic",
		PersonalityInfluence: 0.3,
		UseDialogHistory:     false,
		ResponseVariation:    0.2,
		PreferRomanceDialogs: true,
	}

	// Parse configuration
	if len(config) > 0 {
		if err := json.Unmarshal(config, &s.config); err != nil {
			return fmt.Errorf("failed to parse simple random config: %w", err)
		}
	}

	// Validate configuration
	if err := s.validateConfig(); err != nil {
		return fmt.Errorf("invalid simple random config: %w", err)
	}

	s.character = character
	return nil
}

// validateConfig ensures simple random configuration is valid
func (s *SimpleRandomBackend) validateConfig() error {
	if s.config.PersonalityInfluence < 0 || s.config.PersonalityInfluence > 1 {
		return fmt.Errorf("personalityInfluence must be 0-1, got %f", s.config.PersonalityInfluence)
	}

	if s.config.ResponseVariation < 0 || s.config.ResponseVariation > 1 {
		return fmt.Errorf("responseVariation must be 0-1, got %f", s.config.ResponseVariation)
	}

	return nil
}

// GenerateResponse produces a dialog response using existing dialog selection logic
func (s *SimpleRandomBackend) GenerateResponse(context DialogContext) (DialogResponse, error) {
	// Try to use existing character dialog selection logic
	response := s.selectDialogUsingExistingLogic(context)

	if response == "" {
		// Use fallback responses
		response = s.selectFallbackResponse(context)
	}

	// Determine animation based on trigger and response content
	animation := s.selectAnimation(response, context)

	return DialogResponse{
		Text:             response,
		Animation:        animation,
		Confidence:       0.8, // High confidence since we're using existing proven logic
		ResponseType:     "simple",
		EmotionalTone:    s.detectEmotionalTone(response),
		MemoryImportance: s.calculateMemoryImportance(response, context),
	}, nil
}

// selectDialogUsingExistingLogic uses the same logic as the existing HandleClick/HandleRightClick methods
func (s *SimpleRandomBackend) selectDialogUsingExistingLogic(context DialogContext) string {
	// First try romance dialogs if enabled and character has romance features
	if s.config.PreferRomanceDialogs && s.character.card.HasRomanceFeatures() {
		if romanceResponse := s.selectRomanceDialog(context); romanceResponse != "" {
			return romanceResponse
		}
	}

	// Then try basic dialogs
	return s.selectBasicDialog(context)
}

// selectRomanceDialog attempts to select a romance dialog using existing logic
func (s *SimpleRandomBackend) selectRomanceDialog(context DialogContext) string {
	// Replicate the romance dialog selection logic from the existing system
	for _, dialog := range s.character.card.RomanceDialogs {
		if !s.matchesTrigger(dialog.Trigger, context.Trigger) {
			continue
		}

		// Check requirements using existing logic (simplified)
		if s.character.gameState != nil && dialog.Requirements != nil {
			// Use existing requirement checking if game state is available
			if !s.canSatisfyRomanceRequirements(dialog.Requirements) {
				continue
			}
		}

		// Select response with personality influence
		return s.selectResponseWithPersonality(dialog.Responses, context)
	}

	return ""
}

// selectBasicDialog selects from basic dialogs using existing logic
func (s *SimpleRandomBackend) selectBasicDialog(context DialogContext) string {
	for _, dialog := range s.character.card.Dialogs {
		if !s.matchesTrigger(dialog.Trigger, context.Trigger) {
			continue
		}

		// Check cooldown (simplified)
		if s.isOnCooldown(dialog.Trigger, dialog.Cooldown) {
			continue
		}

		// Select response with personality influence
		return s.selectResponseWithPersonality(dialog.Responses, context)
	}

	return ""
}

// matchesTrigger checks if dialog trigger matches context trigger
func (s *SimpleRandomBackend) matchesTrigger(dialogTrigger, contextTrigger string) bool {
	return dialogTrigger == contextTrigger
}

// isOnCooldown checks if a dialog trigger is currently on cooldown
func (s *SimpleRandomBackend) isOnCooldown(trigger string, cooldownSeconds int) bool {
	if s.character.dialogCooldowns == nil || cooldownSeconds <= 0 {
		return false
	}

	lastUsed, exists := s.character.dialogCooldowns[trigger]
	if !exists {
		return false
	}

	cooldownDuration := time.Duration(cooldownSeconds) * time.Second
	return time.Since(lastUsed) < cooldownDuration
}

// canSatisfyRomanceRequirements checks if romance dialog requirements are met (simplified)
func (s *SimpleRandomBackend) canSatisfyRomanceRequirements(requirements *RomanceRequirement) bool {
	if s.character.gameState == nil || requirements == nil {
		return true // No requirements to check
	}

	// Check stats requirements
	if requirements.Stats != nil {
		for statName, conditions := range requirements.Stats {
			currentValue := s.character.gameState.GetStat(statName)

			// Check minimum value
			if minValue, hasMin := conditions["min"]; hasMin && currentValue < minValue {
				return false
			}

			// Check maximum value
			if maxValue, hasMax := conditions["max"]; hasMax && currentValue > maxValue {
				return false
			}
		}
	}

	// Check relationship level requirement
	if requirements.RelationshipLevel != "" {
		currentLevel := s.character.gameState.GetRelationshipLevel()
		if currentLevel != requirements.RelationshipLevel {
			return false
		}
	}

	// Simplified implementation - in practice this would check all requirements
	// including interaction counts and achievements

	return true
}

// selectResponseWithPersonality chooses a response considering personality traits
func (s *SimpleRandomBackend) selectResponseWithPersonality(responses []string, context DialogContext) string {
	if len(responses) == 0 {
		return ""
	}

	if len(responses) == 1 || s.config.PersonalityInfluence == 0 {
		// Simple random selection
		return responses[s.selectRandomIndex(len(responses))]
	}

	// Apply personality influence to response selection
	scores := make([]float64, len(responses))

	for i, response := range responses {
		scores[i] = s.scoreResponseForPersonality(response, context)
	}

	// Select based on weighted scores
	return responses[s.selectWeightedIndex(scores)]
}

// scoreResponseForPersonality scores a response based on personality traits
func (s *SimpleRandomBackend) scoreResponseForPersonality(response string, context DialogContext) float64 {
	baseScore := 1.0

	if s.config.PersonalityInfluence == 0 {
		return baseScore
	}

	// Simple personality-based scoring
	lowerResponse := strings.ToLower(response)

	// Shyness affects response selection
	shyness := context.PersonalityTraits["shyness"]
	if shyness > 0 {
		// Shy characters prefer shorter, less bold responses
		wordCount := len(strings.Fields(response))
		if wordCount > 10 {
			baseScore -= shyness * 0.3
		}
		// Prefer responses without exclamation marks
		if strings.Contains(response, "!") {
			baseScore -= shyness * 0.2
		}
	}

	// Romanticism affects romantic content preference
	romanticism := context.PersonalityTraits["romanticism"]
	if romanticism > 0 {
		romanticWords := []string{"love", "heart", "dear", "darling", "sweet"}
		for _, word := range romanticWords {
			if strings.Contains(lowerResponse, word) {
				baseScore += romanticism * 0.4
				break
			}
		}
	}

	// Apply personality influence strength
	personalityAdjustment := (baseScore - 1.0) * s.config.PersonalityInfluence
	return 1.0 + personalityAdjustment
}

// selectRandomIndex selects a random index with optional variation
func (s *SimpleRandomBackend) selectRandomIndex(length int) int {
	if s.config.ResponseVariation == 0 {
		// Pure time-based selection for consistency
		return int(time.Now().UnixNano()) % length
	}

	// Mix time-based and random selection
	timeIndex := int(time.Now().UnixNano()) % length
	randomIndex := rand.Intn(length)

	// Blend based on variation setting
	if rand.Float64() < s.config.ResponseVariation {
		return randomIndex
	}
	return timeIndex
}

// selectWeightedIndex selects an index based on weighted scores
func (s *SimpleRandomBackend) selectWeightedIndex(scores []float64) int {
	if len(scores) == 0 {
		return 0
	}

	// Find total weight
	totalWeight := 0.0
	for _, score := range scores {
		totalWeight += score
	}

	if totalWeight <= 0 {
		// Fallback to simple random
		return s.selectRandomIndex(len(scores))
	}

	// Select based on weights
	target := rand.Float64() * totalWeight
	current := 0.0

	for i, score := range scores {
		current += score
		if current >= target {
			return i
		}
	}

	// Fallback to last index
	return len(scores) - 1
}

// selectFallbackResponse provides a fallback when no dialogs match
func (s *SimpleRandomBackend) selectFallbackResponse(context DialogContext) string {
	// Use configured fallback responses
	if len(s.config.FallbackResponses) > 0 {
		index := s.selectRandomIndex(len(s.config.FallbackResponses))
		return s.config.FallbackResponses[index]
	}

	// Use context fallback responses
	if len(context.FallbackResponses) > 0 {
		index := s.selectRandomIndex(len(context.FallbackResponses))
		return context.FallbackResponses[index]
	}

	// Hard-coded fallbacks based on trigger
	switch context.Trigger {
	case "click":
		return "Hello! Nice to see you! 👋"
	case "rightclick":
		return "How can I help you today?"
	case "hover":
		return "Hi there!"
	default:
		return "Hello! 😊"
	}
}

// selectAnimation chooses an appropriate animation for the response
func (s *SimpleRandomBackend) selectAnimation(response string, context DialogContext) string {
	// Use context fallback animation if available
	if context.FallbackAnimation != "" {
		return context.FallbackAnimation
	}

	// Simple heuristic-based animation selection
	lowerResponse := strings.ToLower(response)

	// Emotional content detection
	if strings.Contains(lowerResponse, "love") || strings.Contains(lowerResponse, "heart") {
		return "heart_eyes"
	}
	if strings.Contains(lowerResponse, "thank") {
		return "happy"
	}
	if strings.Contains(lowerResponse, "shy") || strings.Contains(lowerResponse, "blush") {
		return "blushing"
	}
	if strings.Contains(lowerResponse, "sad") || strings.Contains(lowerResponse, "sorry") {
		return "sad"
	}

	// Default based on trigger
	switch context.Trigger {
	case "compliment":
		return "blushing"
	case "rightclick":
		return "thinking"
	default:
		return "talking"
	}
}

// detectEmotionalTone analyzes the emotional tone of the response
func (s *SimpleRandomBackend) detectEmotionalTone(response string) string {
	lowerResponse := strings.ToLower(response)

	// Positive emotions
	if strings.Contains(lowerResponse, "happy") || strings.Contains(lowerResponse, "joy") ||
		strings.Contains(lowerResponse, "excited") || strings.Contains(lowerResponse, "wonderful") {
		return "happy"
	}

	// Romantic emotions
	if strings.Contains(lowerResponse, "love") || strings.Contains(lowerResponse, "heart") ||
		strings.Contains(lowerResponse, "dear") || strings.Contains(lowerResponse, "darling") {
		return "romantic"
	}

	// Shy emotions
	if strings.Contains(lowerResponse, "shy") || strings.Contains(lowerResponse, "blush") ||
		strings.Contains(lowerResponse, "nervous") {
		return "shy"
	}

	// Sad emotions
	if strings.Contains(lowerResponse, "sad") || strings.Contains(lowerResponse, "sorry") ||
		strings.Contains(lowerResponse, "miss") {
		return "sad"
	}

	return "neutral"
}

// calculateMemoryImportance determines how important this response is for memory storage
func (s *SimpleRandomBackend) calculateMemoryImportance(response string, context DialogContext) float64 {
	importance := 0.3 // Base importance for simple responses

	// Romance responses are more important
	if s.detectEmotionalTone(response) == "romantic" {
		importance += 0.4
	}

	// Emotional responses are more important
	tone := s.detectEmotionalTone(response)
	if tone != "neutral" {
		importance += 0.2
	}

	// Clamp to valid range
	if importance > 1.0 {
		importance = 1.0
	}

	return importance
}

// GetBackendInfo returns metadata about the simple random backend
func (s *SimpleRandomBackend) GetBackendInfo() BackendInfo {
	return BackendInfo{
		Name:        "simple_random",
		Version:     "1.0.0",
		Description: "Simple random dialog selection with personality influence, compatible with existing system",
		Capabilities: []string{
			"basic_dialog_selection",
			"personality_influence",
			"romance_dialog_support",
			"fallback_responses",
			"existing_system_compatibility",
		},
		Author:  "DDS Development Team",
		License: "MIT",
	}
}

// CanHandle checks if this backend can process the given context
func (s *SimpleRandomBackend) CanHandle(context DialogContext) bool {
	// Simple random backend can always handle any context (fallback capability)
	return true
}

// UpdateMemory records interaction outcomes (simple implementation)
func (s *SimpleRandomBackend) UpdateMemory(context DialogContext, response DialogResponse, feedback *UserFeedback) error {
	// Simple backend doesn't implement learning, but could record basic statistics
	// Future enhancement could track which responses get positive feedback
	return nil
}
