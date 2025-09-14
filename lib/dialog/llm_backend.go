package dialog

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/opd-ai/minilm/dialog"
	"github.com/sirupsen/logrus"
)

// LLMDialogBackend adapts miniLM's LLMBackend to implement our DialogBackend interface
// This provides optional LLM-powered responses while maintaining full compatibility
// with the existing dialog system architecture.
type LLMDialogBackend struct {
	llmBackend  *dialog.LLMBackend
	config      LLMDialogConfig
	manager     *dialog.DialogManager
	enabled     bool
	initialized bool
}

// LLMDialogConfig extends the miniLM LLMConfig with DDS-specific options
type LLMDialogConfig struct {
	// Core LLM configuration (passed through to miniLM)
	ModelPath     string  `json:"modelPath"`     // Path to GGUF model file
	MaxTokens     int     `json:"maxTokens"`     // Maximum tokens to generate
	Temperature   float64 `json:"temperature"`   // Randomness in generation (0.0-2.0)
	TopP          float64 `json:"topP"`          // Nucleus sampling threshold
	TopK          int     `json:"topK"`          // Top-K sampling parameter
	RepeatPenalty float64 `json:"repeatPenalty"` // Penalty for token repetition
	ContextSize   int     `json:"contextSize"`   // Maximum context window size

	// DDS-specific configuration
	PersonalityWeight float64 `json:"personalityWeight"` // How much personality affects responses (0.0-2.0)
	MoodInfluence     float64 `json:"moodInfluence"`     // How much mood affects responses (0.0-2.0)
	UseCharacterName  bool    `json:"useCharacterName"`  // Include character name in prompts
	UseSituation      bool    `json:"useSituation"`      // Include current situation context
	UseRelationship   bool    `json:"useRelationship"`   // Include relationship level in prompts

	// System prompts and personality extraction
	SystemPrompt      string   `json:"systemPrompt"`      // Base system prompt template
	PersonalityPrompt string   `json:"personalityPrompt"` // Personality-specific prompt additions
	FallbackResponses []string `json:"fallbackResponses"` // Responses if LLM fails

	// Feature toggles
	Enabled  bool `json:"enabled"`  // Master enable/disable switch
	MockMode bool `json:"mockMode"` // Use mock responses for development
	Debug    bool `json:"debug"`    // Enable debug logging

	// Performance settings
	MaxGenerationTime   int `json:"maxGenerationTime"`   // Max time in seconds for generation
	HealthCheckInterval int `json:"healthCheckInterval"` // How often to check LLM health (seconds)
	ConcurrentRequests  int `json:"concurrentRequests"`  // Max concurrent LLM requests
}

// NewLLMDialogBackend creates a new LLM dialog backend adapter
func NewLLMDialogBackend() *LLMDialogBackend {
	return &LLMDialogBackend{
		enabled:     false,
		initialized: false,
		config: LLMDialogConfig{
			// Sensible defaults
			MaxTokens:           50,
			Temperature:         0.8,
			TopP:                0.9,
			TopK:                40,
			RepeatPenalty:       1.1,
			ContextSize:         2048,
			PersonalityWeight:   1.0,
			MoodInfluence:       0.7,
			UseCharacterName:    true,
			UseSituation:        true,
			UseRelationship:     true,
			MaxGenerationTime:   30,
			HealthCheckInterval: 60,
			ConcurrentRequests:  2,
			SystemPrompt:        "You are a virtual companion character. Respond in character with a short, natural response.",
			PersonalityPrompt:   "",
			FallbackResponses: []string{
				"I'm thinking about that...",
				"Let me process that for a moment...",
				"That's interesting to consider...",
			},
		},
	}
}

// Initialize implements DialogBackend.Initialize
func (llm *LLMDialogBackend) Initialize(configData json.RawMessage) error {
	// Parse our extended configuration
	if err := json.Unmarshal(configData, &llm.config); err != nil {
		return fmt.Errorf("failed to parse LLM dialog config: %w", err)
	}

	// Check if LLM backend is enabled
	if !llm.config.Enabled {
		logrus.Info("LLM dialog backend disabled via configuration")
		llm.enabled = false
		llm.initialized = true
		return nil
	}

	// Create miniLM DialogManager and LLMBackend
	llm.manager = dialog.NewDialogManager(llm.config.Debug)
	llm.llmBackend = dialog.NewLLMBackend()

	// Convert our config to miniLM's LLMConfig format
	miniLMConfig := dialog.LLMConfig{
		ModelPath:   llm.config.ModelPath,
		MaxTokens:   llm.config.MaxTokens,
		Temperature: float32(llm.config.Temperature),
		TopP:        float32(llm.config.TopP),
		ContextSize: llm.config.ContextSize,
	}

	// Initialize the miniLM backend
	miniLMConfigJSON, err := json.Marshal(miniLMConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal miniLM config: %w", err)
	}

	if err := llm.llmBackend.Initialize(miniLMConfigJSON); err != nil {
		logrus.WithError(err).Warn("LLM backend initialization failed, will use fallbacks")
		llm.enabled = false
		llm.initialized = true
		return nil // Don't fail completely, just disable LLM features
	}

	// Register the LLM backend with the manager
	llm.manager.RegisterBackend("llm", llm.llmBackend)
	if err := llm.manager.SetDefaultBackend("llm"); err != nil {
		return fmt.Errorf("failed to set LLM as default backend: %w", err)
	}

	llm.enabled = true
	llm.initialized = true

	logrus.WithFields(logrus.Fields{
		"modelPath":   llm.config.ModelPath,
		"maxTokens":   llm.config.MaxTokens,
		"temperature": llm.config.Temperature,
		"mockMode":    llm.config.MockMode,
	}).Info("LLM dialog backend initialized successfully")

	return nil
}

// CanHandle implements DialogBackend.CanHandle
func (llm *LLMDialogBackend) CanHandle(context DialogContext) bool {
	if !llm.initialized {
		return false
	}

	// If disabled, we can't handle anything
	if !llm.enabled {
		return false
	}

	// LLM backend can handle any text-based dialog request
	// We'll let the actual generation logic decide if it's appropriate
	return context.Trigger != "" // Just need a valid trigger
}

// GenerateResponse implements DialogBackend.GenerateResponse
func (llm *LLMDialogBackend) GenerateResponse(context DialogContext) (DialogResponse, error) {
	if !llm.initialized || !llm.enabled {
		return DialogResponse{}, fmt.Errorf("LLM backend not available")
	}

	// Check health before attempting generation
	if !llm.IsHealthy() {
		logrus.Warn("LLM backend health check failed, triggering fallback")
		return llm.createFallbackResponse(context), fmt.Errorf("LLM backend unhealthy")
	}

	// Convert our DialogContext to miniLM's format
	miniLMContext := llm.buildMiniLMContext(context)

	// Generate response using miniLM with timeout
	responseChan := make(chan DialogResponse, 1)
	errorChan := make(chan error, 1)

	go func() {
		response, err := llm.manager.GenerateDialog(miniLMContext)
		if err != nil {
			errorChan <- err
			return
		}

		// Convert miniLM response back to our format
		dialogResponse := DialogResponse{
			Text:             response.Text,
			Animation:        response.Animation,
			Duration:         response.Duration,
			Confidence:       response.Confidence,
			ResponseType:     response.ResponseType,
			EmotionalTone:    response.EmotionalTone,
			Topics:           response.Topics,
			MemoryImportance: response.MemoryImportance,
			LearningValue:    response.LearningValue,
			Metadata:         response.Metadata,
		}

		responseChan <- dialogResponse
	}()

	// Wait for response with timeout
	timeout := time.Duration(llm.config.MaxGenerationTime) * time.Second
	select {
	case response := <-responseChan:
		// Validate response quality
		if response.Text == "" || response.Confidence < 0.3 {
			logrus.WithFields(logrus.Fields{
				"text":       response.Text,
				"confidence": response.Confidence,
			}).Debug("LLM response quality too low, triggering fallback")
			return llm.createFallbackResponse(context), fmt.Errorf("LLM response quality insufficient")
		}

		// Add our own metadata
		if response.Metadata == nil {
			response.Metadata = make(map[string]interface{})
		}
		response.Metadata["backend"] = "llm"
		response.Metadata["model_path"] = llm.config.ModelPath
		response.Metadata["generation_time"] = time.Now().Format(time.RFC3339)

		logrus.WithFields(logrus.Fields{
			"trigger":    context.Trigger,
			"text":       response.Text,
			"confidence": response.Confidence,
			"animation":  response.Animation,
		}).Debug("LLM response generated successfully")

		return response, nil

	case err := <-errorChan:
		logrus.WithError(err).Debug("LLM generation failed, returning error for fallback")
		return llm.createFallbackResponse(context), fmt.Errorf("LLM generation failed: %w", err)

	case <-time.After(timeout):
		logrus.WithField("timeout", timeout).Warn("LLM generation timed out, triggering fallback")
		return llm.createFallbackResponse(context), fmt.Errorf("LLM generation timed out after %v", timeout)
	}
}

// buildMiniLMContext converts our DialogContext to miniLM's format
func (llm *LLMDialogBackend) buildMiniLMContext(context DialogContext) dialog.DialogContext {
	// Note: miniLM doesn't support custom SystemPrompt in DialogContext
	// We'll need to handle personality prompting differently

	return dialog.DialogContext{
		Trigger:            context.Trigger,
		InteractionID:      context.InteractionID,
		Timestamp:          context.Timestamp,
		CurrentStats:       context.CurrentStats,
		PersonalityTraits:  context.PersonalityTraits,
		CurrentMood:        context.CurrentMood,
		CurrentAnimation:   context.CurrentAnimation,
		RelationshipLevel:  context.RelationshipLevel,
		InteractionHistory: convertInteractionHistory(context.InteractionHistory),
		AchievementStatus:  context.AchievementStatus,
		TimeOfDay:          context.TimeOfDay,
		LastResponse:       context.LastResponse,
		ConversationTurn:   context.ConversationTurn,
		TopicContext:       context.TopicContext,
		FallbackResponses:  context.FallbackResponses,
		FallbackAnimation:  context.FallbackAnimation,
	}
}

// buildPersonalityPrompt creates a personality-aware prompt for the LLM
func (llm *LLMDialogBackend) buildPersonalityPrompt(context DialogContext) string {
	// Create personality extractor
	extractor := NewPersonalityExtractor("", "") // Character name/description would come from context

	// Extract personality from traits if available
	var personalityPrompt PersonalityPrompt
	if len(context.PersonalityTraits) > 0 {
		personalityPrompt = extractor.ExtractFromTraits(context.PersonalityTraits)
	} else {
		// Use default personality
		personalityPrompt = PersonalityPrompt{
			SystemPrompt:     llm.config.SystemPrompt,
			PersonalityHints: llm.config.PersonalityPrompt,
			EmotionalTone:    "neutral",
			ResponseStyle:    "casual",
		}
	}

	// Customize prompt with context
	prompt := personalityPrompt.SystemPrompt
	if prompt == "" {
		prompt = llm.config.SystemPrompt
	}

	// Add personality context if enabled
	if llm.config.PersonalityWeight > 0 && personalityPrompt.PersonalityHints != "" {
		prompt += "\n\n" + personalityPrompt.PersonalityHints
	}

	// Add mood context if enabled
	if llm.config.MoodInfluence > 0 {
		moodDesc := "neutral"
		if context.CurrentMood > 80 {
			moodDesc = "very happy"
		} else if context.CurrentMood > 60 {
			moodDesc = "happy"
		} else if context.CurrentMood < 20 {
			moodDesc = "sad"
		} else if context.CurrentMood < 40 {
			moodDesc = "somewhat sad"
		}
		prompt += fmt.Sprintf("\n\nCurrent mood: %s (%.0f/100)", moodDesc, context.CurrentMood)
	}

	// Add relationship context if enabled
	if llm.config.UseRelationship && context.RelationshipLevel != "" {
		prompt += fmt.Sprintf("\n\nRelationship level: %s", context.RelationshipLevel)
	}

	// Add situation context
	if llm.config.UseSituation {
		prompt += fmt.Sprintf("\n\nSituation: User %s the character", context.Trigger)
		if context.TimeOfDay != "" {
			prompt += fmt.Sprintf(" during %s", context.TimeOfDay)
		}
	}

	// Add speech patterns if available
	if len(personalityPrompt.SpeechPatterns) > 0 {
		prompt += fmt.Sprintf("\n\nSpeech patterns: %s", strings.Join(personalityPrompt.SpeechPatterns, ", "))
	}

	// Add custom personality prompt
	if llm.config.PersonalityPrompt != "" && personalityPrompt.PersonalityHints == "" {
		prompt += "\n\n" + llm.config.PersonalityPrompt
	}

	prompt += "\n\nRespond with a short, natural response (1-2 sentences) that fits the character and situation."

	return prompt
}

// convertInteractionHistory converts our interaction records to miniLM format
func convertInteractionHistory(history []InteractionRecord) []dialog.InteractionRecord {
	converted := make([]dialog.InteractionRecord, len(history))
	for i, record := range history {
		converted[i] = dialog.InteractionRecord{
			Type:      record.Type,
			Response:  record.Response,
			Timestamp: record.Timestamp,
			Stats:     record.Stats,
			Outcome:   record.Outcome,
		}
	}
	return converted
}

// createFallbackResponse creates a fallback response when LLM fails
func (llm *LLMDialogBackend) createFallbackResponse(context DialogContext) DialogResponse {
	// Use configured fallback responses or context fallbacks
	responses := llm.config.FallbackResponses
	if len(context.FallbackResponses) > 0 {
		responses = context.FallbackResponses
	}

	// Pick a random fallback response
	responseText := "I'm processing that..."
	if len(responses) > 0 {
		// Use timestamp-based selection for deterministic but varying responses
		responseText = responses[int(time.Now().Unix())%len(responses)]
	}

	animation := context.FallbackAnimation
	if animation == "" {
		animation = "talking"
	}

	return DialogResponse{
		Text:         responseText,
		Animation:    animation,
		Confidence:   0.3, // Low confidence to indicate fallback
		ResponseType: "fallback",
		Metadata: map[string]interface{}{
			"backend":         "llm",
			"fallback_used":   true,
			"fallback_reason": "llm_generation_failed",
		},
	}
}

// GetBackendInfo implements DialogBackend.GetBackendInfo
func (llm *LLMDialogBackend) GetBackendInfo() BackendInfo {
	version := "1.0.0"
	capabilities := []string{"text_generation", "personality_aware", "context_aware", "mood_responsive"}

	if llm.config.MockMode {
		capabilities = append(capabilities, "mock_mode")
	}

	if !llm.enabled {
		capabilities = append(capabilities, "disabled")
	}

	return BackendInfo{
		Name:         "llm_dialog",
		Version:      version,
		Description:  "LLM-powered dialog generation using miniLM with personality and context awareness",
		Capabilities: capabilities,
		Author:       "Desktop Companion System",
		License:      "MIT",
	}
}

// UpdateMemory implements DialogBackend.UpdateMemory
func (llm *LLMDialogBackend) UpdateMemory(context DialogContext, response DialogResponse, feedback *UserFeedback) error {
	if !llm.initialized || !llm.enabled || llm.manager == nil {
		return nil // Silently ignore if not available
	}

	// Convert to miniLM format and update memory
	miniLMContext := llm.buildMiniLMContext(context)

	miniLMResponse := dialog.DialogResponse{
		Text:             response.Text,
		Animation:        response.Animation,
		Duration:         response.Duration,
		Confidence:       response.Confidence,
		ResponseType:     response.ResponseType,
		EmotionalTone:    response.EmotionalTone,
		Topics:           response.Topics,
		MemoryImportance: response.MemoryImportance,
		LearningValue:    response.LearningValue,
		Metadata:         response.Metadata,
	}

	var miniLMFeedback *dialog.UserFeedback
	if feedback != nil {
		miniLMFeedback = &dialog.UserFeedback{
			Positive:     feedback.Positive,
			ResponseTime: feedback.ResponseTime,
			FollowUpType: feedback.FollowUpType,
			Engagement:   feedback.Engagement,
			CustomData:   feedback.CustomData,
		}
	}

	// Use miniLM's memory update function
	dialog.UpdateBackendMemory(llm.manager, miniLMContext, miniLMResponse, miniLMFeedback)

	logrus.WithFields(logrus.Fields{
		"trigger":  context.Trigger,
		"positive": feedback != nil && feedback.Positive,
		"engagement": func() float64 {
			if feedback != nil {
				return feedback.Engagement
			}
			return 0.0
		}(),
	}).Debug("Updated LLM backend memory")

	return nil
}

// Health checking and management methods

// IsHealthy checks if the LLM backend is functioning properly
func (llm *LLMDialogBackend) IsHealthy() bool {
	if !llm.initialized || !llm.enabled {
		return false
	}

	if llm.manager == nil || llm.llmBackend == nil {
		return false
	}

	// TODO: Implement actual health check with miniLM
	// For now, just return our enabled status
	return llm.enabled
}

// GetModelInfo returns information about the loaded model
func (llm *LLMDialogBackend) GetModelInfo() map[string]interface{} {
	info := map[string]interface{}{
		"enabled":     llm.enabled,
		"initialized": llm.initialized,
		"model_path":  llm.config.ModelPath,
		"mock_mode":   llm.config.MockMode,
	}

	if llm.enabled && llm.llmBackend != nil {
		// TODO: Get additional model info from miniLM if available
		info["backend_type"] = "minilm"
		info["max_tokens"] = llm.config.MaxTokens
		info["temperature"] = llm.config.Temperature
		info["context_size"] = llm.config.ContextSize
	}

	return info
}

// Shutdown cleanly shuts down the LLM backend
func (llm *LLMDialogBackend) Shutdown() error {
	if llm.llmBackend != nil {
		// TODO: Implement shutdown method in miniLM if available
		logrus.Info("Shutting down LLM dialog backend")
	}

	llm.enabled = false
	llm.initialized = false
	llm.llmBackend = nil
	llm.manager = nil

	return nil
}

// HandleError processes errors and updates backend state for automatic recovery
func (llm *LLMDialogBackend) HandleError(err error) {
	if err == nil {
		return
	}

	logrus.WithError(err).Debug("LLM backend error occurred")

	// Check for critical errors that should disable the backend
	errorStr := err.Error()
	criticalErrors := []string{
		"model not found",
		"model load failed",
		"insufficient memory",
		"cuda error",
	}

	for _, critical := range criticalErrors {
		if strings.Contains(strings.ToLower(errorStr), critical) {
			logrus.WithError(err).Warn("Critical LLM error detected, disabling backend")
			llm.enabled = false
			break
		}
	}
}

// RecoverFromError attempts to recover from errors and re-enable the backend
func (llm *LLMDialogBackend) RecoverFromError() error {
	if llm.enabled {
		return nil // Already enabled
	}

	if !llm.initialized {
		return fmt.Errorf("backend not initialized")
	}

	// Attempt to re-initialize the miniLM backend
	logrus.Info("Attempting to recover LLM backend...")

	// Reset the backend
	llm.llmBackend = dialog.NewLLMBackend()
	llm.manager = dialog.NewDialogManager(llm.config.Debug)

	// Re-initialize with current config
	miniLMConfig := dialog.LLMConfig{
		ModelPath:   llm.config.ModelPath,
		MaxTokens:   llm.config.MaxTokens,
		Temperature: float32(llm.config.Temperature),
		TopP:        float32(llm.config.TopP),
		ContextSize: llm.config.ContextSize,
	}

	miniLMConfigJSON, err := json.Marshal(miniLMConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal recovery config: %w", err)
	}

	if err := llm.llmBackend.Initialize(miniLMConfigJSON); err != nil {
		return fmt.Errorf("failed to reinitialize LLM backend: %w", err)
	}

	// Register the backend with the manager
	llm.manager.RegisterBackend("llm", llm.llmBackend)
	if err := llm.manager.SetDefaultBackend("llm"); err != nil {
		return fmt.Errorf("failed to set LLM as default backend during recovery: %w", err)
	}

	llm.enabled = true

	logrus.Info("LLM backend recovered successfully")
	return nil
}
