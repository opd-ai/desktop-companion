package dialog

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// TestLLMDialogBackend_Integration tests the complete LLM backend integration
func TestLLMDialogBackend_Integration(t *testing.T) {
	backend := NewLLMDialogBackend()

	// Test initial state
	if backend.IsHealthy() {
		t.Error("Backend should not be healthy before initialization")
	}

	// Test initialization with disabled LLM
	config := LLMDialogConfig{
		Enabled:  false,
		MockMode: true,
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	if err := backend.Initialize(configJSON); err != nil {
		t.Fatalf("Failed to initialize disabled backend: %v", err)
	}

	// Backend should be initialized but not enabled
	if !backend.initialized {
		t.Error("Backend should be initialized")
	}
	if backend.enabled {
		t.Error("Backend should not be enabled when disabled in config")
	}
}

// TestLLMDialogBackend_MockMode tests LLM backend in mock mode
func TestLLMDialogBackend_MockMode(t *testing.T) {
	backend := NewLLMDialogBackend()

	// Initialize with mock mode enabled
	config := LLMDialogConfig{
		Enabled:           true,
		MockMode:          true,
		MaxTokens:         30,
		Temperature:       0.8,
		SystemPrompt:      "You are a test character.",
		PersonalityPrompt: "Be friendly and helpful.",
		FallbackResponses: []string{"Test fallback response"},
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	if err := backend.Initialize(configJSON); err != nil {
		t.Fatalf("Failed to initialize mock backend: %v", err)
	}

	// Test backend info
	info := backend.GetBackendInfo()
	if info.Name != "llm_dialog" {
		t.Errorf("Expected backend name 'llm_dialog', got '%s'", info.Name)
	}
	if !contains(info.Capabilities, "mock_mode") {
		t.Error("Backend should report mock_mode capability")
	}
}

// TestLLMDialogBackend_CanHandle tests the CanHandle method
func TestLLMDialogBackend_CanHandle(t *testing.T) {
	backend := NewLLMDialogBackend()

	// Uninitialized backend should not handle anything
	context := DialogContext{Trigger: "click"}
	if backend.CanHandle(context) {
		t.Error("Uninitialized backend should not handle requests")
	}

	// Initialize but disable
	config := LLMDialogConfig{Enabled: false}
	configJSON, _ := json.Marshal(config)
	backend.Initialize(configJSON)

	if backend.CanHandle(context) {
		t.Error("Disabled backend should not handle requests")
	}

	// Enable backend
	config.Enabled = true
	config.MockMode = true
	configJSON, _ = json.Marshal(config)
	backend.Initialize(configJSON)

	if !backend.CanHandle(context) {
		t.Error("Enabled backend should handle valid requests")
	}

	// Test with empty trigger
	emptyContext := DialogContext{Trigger: ""}
	if backend.CanHandle(emptyContext) {
		t.Error("Backend should not handle requests with empty trigger")
	}
}

// TestLLMDialogBackend_FallbackResponse tests fallback response generation
func TestLLMDialogBackend_FallbackResponse(t *testing.T) {
	backend := NewLLMDialogBackend()

	// Test fallback with default responses
	context := DialogContext{
		Trigger:           "click",
		FallbackResponses: []string{"Custom fallback 1", "Custom fallback 2"},
		FallbackAnimation: "talking",
	}

	response := backend.createFallbackResponse(context)

	if response.Text == "" {
		t.Error("Fallback response should have text")
	}
	if response.Animation != "talking" {
		t.Errorf("Expected fallback animation 'talking', got '%s'", response.Animation)
	}
	if response.Confidence != 0.3 {
		t.Errorf("Expected fallback confidence 0.3, got %f", response.Confidence)
	}
	if response.ResponseType != "fallback" {
		t.Errorf("Expected response type 'fallback', got '%s'", response.ResponseType)
	}

	// Check metadata
	if response.Metadata == nil {
		t.Error("Fallback response should have metadata")
	}
	if response.Metadata["fallback_used"] != true {
		t.Error("Fallback metadata should indicate fallback was used")
	}
}

// TestLLMDialogBackend_PersonalityExtraction tests personality prompt building
func TestLLMDialogBackend_PersonalityExtraction(t *testing.T) {
	backend := NewLLMDialogBackend()

	// Initialize with personality-aware configuration
	config := LLMDialogConfig{
		Enabled:           true,
		MockMode:          true,
		PersonalityWeight: 1.5,
		MoodInfluence:     1.0,
		UseRelationship:   true,
		UseSituation:      true,
		SystemPrompt:      "You are a test character.",
		PersonalityPrompt: "Custom personality context.",
	}
	configJSON, _ := json.Marshal(config)
	backend.Initialize(configJSON)

	// Test context with personality traits
	context := DialogContext{
		Trigger: "click",
		PersonalityTraits: map[string]float64{
			"shyness":      0.8,
			"friendliness": 0.9,
			"intelligence": 0.7,
		},
		CurrentMood:       75.0,
		RelationshipLevel: "friend",
		TimeOfDay:         "evening",
	}

	prompt := backend.buildPersonalityPrompt(context)

	// Check that personality information is included
	if !strings.Contains(prompt, "shyness") && !strings.Contains(prompt, "Strong personality traits") {
		t.Error("Prompt should include personality information")
	}
	if !strings.Contains(prompt, "mood: happy") {
		t.Error("Prompt should include mood information")
	}
	if !strings.Contains(prompt, "Relationship level: friend") {
		t.Error("Prompt should include relationship information")
	}
	if !strings.Contains(prompt, "User click the character") {
		t.Error("Prompt should include situation information")
	}
	if !strings.Contains(prompt, "during evening") {
		t.Error("Prompt should include time of day information")
	}
}

// TestLLMDialogBackend_GenerateResponse tests response generation with error scenarios
func TestLLMDialogBackend_GenerateResponse(t *testing.T) {
	backend := NewLLMDialogBackend()

	// Test uninitialized backend
	context := DialogContext{Trigger: "click"}
	_, err := backend.GenerateResponse(context)
	if err == nil {
		t.Error("Uninitialized backend should return error")
	}

	// Test disabled backend
	config := LLMDialogConfig{Enabled: false}
	configJSON, _ := json.Marshal(config)
	backend.Initialize(configJSON)

	_, err = backend.GenerateResponse(context)
	if err == nil {
		t.Error("Disabled backend should return error")
	}
	if !strings.Contains(err.Error(), "not available") {
		t.Errorf("Error should indicate backend not available, got: %v", err)
	}
}

// TestLLMDialogBackend_UpdateMemory tests memory update functionality
func TestLLMDialogBackend_UpdateMemory(t *testing.T) {
	backend := NewLLMDialogBackend()

	// Initialize in mock mode
	config := LLMDialogConfig{
		Enabled:  true,
		MockMode: true,
	}
	configJSON, _ := json.Marshal(config)
	backend.Initialize(configJSON)

	context := DialogContext{Trigger: "click"}
	response := DialogResponse{Text: "Test response", Confidence: 0.8}
	feedback := &UserFeedback{
		Positive:   true,
		Engagement: 0.9,
	}

	// Should not error even if miniLM memory update fails
	err := backend.UpdateMemory(context, response, feedback)
	if err != nil {
		t.Errorf("Memory update should not error in mock mode: %v", err)
	}

	// Test with disabled backend
	backend.enabled = false
	err = backend.UpdateMemory(context, response, feedback)
	if err != nil {
		t.Errorf("Memory update should silently succeed for disabled backend: %v", err)
	}
}

// TestLLMDialogBackend_ErrorHandling tests error handling and recovery
func TestLLMDialogBackend_ErrorHandling(t *testing.T) {
	backend := NewLLMDialogBackend()

	// Initialize backend
	config := LLMDialogConfig{
		Enabled:  true,
		MockMode: true,
	}
	configJSON, _ := json.Marshal(config)
	backend.Initialize(configJSON)

	// Test critical error handling
	criticalErr := fmt.Errorf("model not found")
	backend.HandleError(criticalErr)

	if backend.enabled {
		t.Error("Backend should be disabled after critical error")
	}

	// Test recovery
	err := backend.RecoverFromError()
	if err != nil {
		t.Errorf("Recovery should succeed in mock mode: %v", err)
	}

	if !backend.enabled {
		t.Error("Backend should be enabled after successful recovery")
	}
}

// TestLLMDialogBackend_ConfigValidation tests configuration validation
func TestLLMDialogBackend_ConfigValidation(t *testing.T) {
	backend := NewLLMDialogBackend()

	// Test invalid JSON
	invalidJSON := []byte(`{"invalid": json`)
	err := backend.Initialize(invalidJSON)
	if err == nil {
		t.Error("Invalid JSON should cause initialization error")
	}

	// Test valid configuration
	config := LLMDialogConfig{
		Enabled:           true,
		MockMode:          true,
		MaxTokens:         50,
		Temperature:       0.8,
		PersonalityWeight: 1.0,
		MoodInfluence:     0.7,
	}
	validJSON, _ := json.Marshal(config)
	err = backend.Initialize(validJSON)
	if err != nil {
		t.Errorf("Valid configuration should not cause error: %v", err)
	}
}

// TestLLMDialogBackend_Shutdown tests clean shutdown
func TestLLMDialogBackend_Shutdown(t *testing.T) {
	backend := NewLLMDialogBackend()

	// Initialize backend
	config := LLMDialogConfig{
		Enabled:  true,
		MockMode: true,
	}
	configJSON, _ := json.Marshal(config)
	backend.Initialize(configJSON)

	// Shutdown
	err := backend.Shutdown()
	if err != nil {
		t.Errorf("Shutdown should not error: %v", err)
	}

	// Check state after shutdown
	if backend.enabled {
		t.Error("Backend should not be enabled after shutdown")
	}
	if backend.initialized {
		t.Error("Backend should not be initialized after shutdown")
	}
	if backend.llmBackend != nil {
		t.Error("LLM backend reference should be nil after shutdown")
	}
	if backend.manager != nil {
		t.Error("Manager reference should be nil after shutdown")
	}
}

// TestLLMDialogBackend_ModelInfo tests model information retrieval
func TestLLMDialogBackend_ModelInfo(t *testing.T) {
	backend := NewLLMDialogBackend()

	// Test uninitialized backend
	info := backend.GetModelInfo()
	if info["enabled"].(bool) {
		t.Error("Uninitialized backend should not be enabled")
	}
	if info["initialized"].(bool) {
		t.Error("Uninitialized backend should not be initialized")
	}

	// Initialize backend
	config := LLMDialogConfig{
		Enabled:   true,
		MockMode:  true,
		ModelPath: "/test/model.gguf",
		MaxTokens: 100,
	}
	configJSON, _ := json.Marshal(config)
	backend.Initialize(configJSON)

	info = backend.GetModelInfo()
	if !info["enabled"].(bool) {
		t.Error("Initialized backend should be enabled")
	}
	if !info["initialized"].(bool) {
		t.Error("Initialized backend should be initialized")
	}
	if info["model_path"].(string) != "/test/model.gguf" {
		t.Error("Model info should include correct model path")
	}
	if info["backend_type"].(string) != "minilm" {
		t.Error("Model info should indicate miniLM backend type")
	}
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
