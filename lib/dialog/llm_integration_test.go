package dialog

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// TestLLMDialogSystem_FullIntegration tests the complete LLM dialog system integration
// with the existing dialog manager and fallback mechanisms
func TestLLMDialogSystem_FullIntegration(t *testing.T) {
	// Create dialog manager
	manager := NewDialogManager(true) // Enable debug mode

	// Register all backends including LLM
	manager.RegisterBackend("simple_random", NewSimpleRandomBackend())
	manager.RegisterBackend("markov_chain", NewMarkovChainBackend())
	manager.RegisterBackend("llm", NewLLMDialogBackend())

	// Configure LLM backend
	llmConfig := LLMDialogConfig{
		Enabled:           true,
		MockMode:          true,
		MaxTokens:         30,
		Temperature:       0.8,
		PersonalityWeight: 1.0,
		MoodInfluence:     0.7,
		UseCharacterName:  true,
		UseSituation:      true,
		UseRelationship:   true,
		SystemPrompt:      "You are a friendly virtual companion.",
		PersonalityPrompt: "Be helpful and engaging.",
		FallbackResponses: []string{"I'm processing that...", "Let me think..."},
		MaxGenerationTime: 5, // Short timeout for testing
	}
	llmConfigJSON, _ := json.Marshal(llmConfig)

	// Configure Markov backend as fallback
	markovConfig := map[string]interface{}{
		"chainOrder": 2,
		"minWords":   3,
		"maxWords":   10,
		"trainingData": []string{
			"Hello there! Nice to see you!",
			"I'm here to keep you company.",
			"Thanks for spending time with me!",
		},
	}
	markovConfigJSON, _ := json.Marshal(markovConfig)

	// Initialize backends
	llmBackend, _ := manager.GetBackend("llm")
	if err := llmBackend.Initialize(llmConfigJSON); err != nil {
		t.Fatalf("Failed to initialize LLM backend: %v", err)
	}

	markovBackend, _ := manager.GetBackend("markov_chain")
	if err := markovBackend.Initialize(markovConfigJSON); err != nil {
		t.Fatalf("Failed to initialize Markov backend: %v", err)
	}

	// Set LLM as default with Markov as fallback
	if err := manager.SetDefaultBackend("llm"); err != nil {
		t.Fatalf("Failed to set LLM as default backend: %v", err)
	}

	if err := manager.SetFallbackChain([]string{"markov_chain", "simple_random"}); err != nil {
		t.Fatalf("Failed to set fallback chain: %v", err)
	}

	// Test normal dialog generation
	context := DialogContext{
		Trigger:       "click",
		InteractionID: "test-1",
		Timestamp:     time.Now(),
		PersonalityTraits: map[string]float64{
			"friendliness": 0.8,
			"helpfulness":  0.9,
		},
		CurrentMood:       70.0,
		RelationshipLevel: "friend",
		TimeOfDay:         "afternoon",
		FallbackResponses: []string{"Default fallback"},
		FallbackAnimation: "talking",
	}

	response, err := manager.GenerateDialog(context)
	if err != nil {
		t.Fatalf("Dialog generation should not fail: %v", err)
	}

	if response.Text == "" {
		t.Error("Generated response should have text")
	}

	// Response should come from LLM backend (or fallback)
	t.Logf("Generated response: %s (confidence: %.2f)", response.Text, response.Confidence)

	// Verify metadata indicates which backend was used
	if response.Metadata == nil {
		t.Error("Response should have metadata")
	}

	backend, hasBackend := response.Metadata["backend"]
	if !hasBackend {
		t.Error("Response metadata should indicate which backend was used")
	}

	// Should be either "llm" or a fallback backend
	if backend != "llm" && backend != "markov_chain" && backend != "simple_random" {
		t.Errorf("Unexpected backend in metadata: %v", backend)
	}
}

// TestLLMDialogSystem_FallbackMechanism tests automatic fallback when LLM fails
func TestLLMDialogSystem_FallbackMechanism(t *testing.T) {
	manager := NewDialogManager(true)

	// Register backends
	manager.RegisterBackend("llm", NewLLMDialogBackend())
	manager.RegisterBackend("markov_chain", NewMarkovChainBackend())
	manager.RegisterBackend("simple_random", NewSimpleRandomBackend())

	// Configure LLM backend to be disabled (simulates failure)
	llmConfig := LLMDialogConfig{
		Enabled: false, // Disabled to force fallback
	}
	llmConfigJSON, _ := json.Marshal(llmConfig)

	markovConfig := map[string]interface{}{
		"chainOrder": 2,
		"trainingData": []string{
			"Fallback response from Markov",
			"This is a Markov generated response",
		},
	}
	markovConfigJSON, _ := json.Marshal(markovConfig)

	// Initialize backends
	llmBackend, _ := manager.GetBackend("llm")
	llmBackend.Initialize(llmConfigJSON)

	markovBackend, _ := manager.GetBackend("markov_chain")
	markovBackend.Initialize(markovConfigJSON)

	// Set LLM as default with fallback chain
	manager.SetDefaultBackend("llm")
	manager.SetFallbackChain([]string{"markov_chain", "simple_random"})

	// Generate dialog - should fallback to Markov
	context := DialogContext{
		Trigger:           "click",
		InteractionID:     "test-fallback",
		Timestamp:         time.Now(),
		FallbackResponses: []string{"Final fallback"},
		FallbackAnimation: "talking",
	}

	response, err := manager.GenerateDialog(context)
	if err != nil {
		t.Fatalf("Dialog generation should not fail with fallbacks: %v", err)
	}

	if response.Text == "" {
		t.Error("Fallback response should have text")
	}

	// Should not be from LLM backend
	if response.Metadata != nil {
		backend := response.Metadata["backend"]
		if backend == "llm" {
			t.Error("Response should not come from disabled LLM backend")
		}
		t.Logf("Fallback response from backend: %v", backend)
	}
}

// TestLLMDialogSystem_PersonalityConsistency tests that personality traits
// are consistently handled between LLM and fallback systems
func TestLLMDialogSystem_PersonalityConsistency(t *testing.T) {
	// Test that the same character configuration produces consistent
	// personality handling across different backends

	commonTraits := map[string]float64{
		"shyness":      0.8,
		"friendliness": 0.6,
		"intelligence": 0.9,
	}

	// Test LLM backend personality extraction
	llmBackend := NewLLMDialogBackend()
	llmConfig := LLMDialogConfig{
		Enabled:           true,
		MockMode:          true,
		PersonalityWeight: 1.5,
	}
	llmConfigJSON, _ := json.Marshal(llmConfig)
	llmBackend.Initialize(llmConfigJSON)

	context := DialogContext{
		Trigger:           "click",
		PersonalityTraits: commonTraits,
		CurrentMood:       80.0,
	}

	// Test that LLM backend can handle the context
	if !llmBackend.CanHandle(context) {
		t.Error("LLM backend should handle context with personality traits")
	}

	// Test personality prompt generation
	prompt := llmBackend.buildPersonalityPrompt(context)
	if prompt == "" {
		t.Error("Personality prompt should not be empty")
	}

	// Should contain personality information
	if !containsAny(prompt, []string{"shy", "friendly", "intelligent", "personality", "traits"}) {
		t.Error("Personality prompt should contain personality information")
	}

	t.Logf("Generated personality prompt: %s", prompt)
}

// TestLLMDialogSystem_BackwardCompatibility tests that existing character
// configurations continue to work without modification
func TestLLMDialogSystem_BackwardCompatibility(t *testing.T) {
	// Simulate existing character configuration without LLM backend
	manager := NewDialogManager(false)

	// Register only traditional backends
	manager.RegisterBackend("simple_random", NewSimpleRandomBackend())
	manager.RegisterBackend("markov_chain", NewMarkovChainBackend())

	// Set up traditional configuration
	markovConfig := map[string]interface{}{
		"chainOrder": 2,
		"trainingData": []string{
			"Hello! I'm your companion.",
			"Nice to see you again!",
			"How are you doing today?",
		},
	}
	markovConfigJSON, _ := json.Marshal(markovConfig)

	markovBackend, _ := manager.GetBackend("markov_chain")
	markovBackend.Initialize(markovConfigJSON)

	manager.SetDefaultBackend("markov_chain")
	manager.SetFallbackChain([]string{"simple_random"})

	// Test that existing dialog generation still works
	context := DialogContext{
		Trigger:           "click",
		InteractionID:     "compat-test",
		Timestamp:         time.Now(),
		FallbackResponses: []string{"Compatibility fallback"},
		FallbackAnimation: "talking",
	}

	response, err := manager.GenerateDialog(context)
	if err != nil {
		t.Fatalf("Backward compatibility dialog generation failed: %v", err)
	}

	if response.Text == "" {
		t.Error("Backward compatibility response should have text")
	}

	t.Logf("Backward compatibility response: %s", response.Text)

	// Now add LLM backend to existing system (simulates upgrade)
	manager.RegisterBackend("llm", NewLLMDialogBackend())

	// LLM backend should be available but not interfere with existing configuration
	llmBackend, exists := manager.GetBackend("llm")
	if !exists {
		t.Error("LLM backend should be registered")
	}

	// LLM backend should be uninitialized and not interfere
	if llmBackend.CanHandle(context) {
		t.Error("Uninitialized LLM backend should not handle requests")
	}

	// Original configuration should still work
	response2, err := manager.GenerateDialog(context)
	if err != nil {
		t.Fatalf("Dialog generation should still work after LLM backend registration: %v", err)
	}

	if response2.Text == "" {
		t.Error("Response should still have text after LLM backend registration")
	}
}

// TestLLMDialogSystem_PerformanceImpact tests that LLM integration
// doesn't negatively impact performance when disabled
func TestLLMDialogSystem_PerformanceImpact(t *testing.T) {
	// Measure performance with and without LLM backend registered

	// Test without LLM backend
	manager1 := NewDialogManager(false)
	manager1.RegisterBackend("simple_random", NewSimpleRandomBackend())
	manager1.SetDefaultBackend("simple_random")

	// Test with LLM backend (disabled)
	manager2 := NewDialogManager(false)
	manager2.RegisterBackend("simple_random", NewSimpleRandomBackend())
	manager2.RegisterBackend("llm", NewLLMDialogBackend())

	llmConfig := LLMDialogConfig{Enabled: false}
	llmConfigJSON, _ := json.Marshal(llmConfig)
	llmBackend, _ := manager2.GetBackend("llm")
	llmBackend.Initialize(llmConfigJSON)

	manager2.SetDefaultBackend("simple_random")

	context := DialogContext{
		Trigger:           "click",
		FallbackResponses: []string{"Test response"},
	}

	// Measure performance (simple timing test)
	iterations := 100

	start1 := time.Now()
	for i := 0; i < iterations; i++ {
		manager1.GenerateDialog(context)
	}
	duration1 := time.Since(start1)

	start2 := time.Now()
	for i := 0; i < iterations; i++ {
		manager2.GenerateDialog(context)
	}
	duration2 := time.Since(start2)

	// Performance impact should be minimal (within 50% difference)
	ratio := float64(duration2) / float64(duration1)
	if ratio > 1.5 {
		t.Errorf("Performance impact too high: %.2fx slower with LLM backend", ratio)
	}

	t.Logf("Performance: without LLM: %v, with LLM (disabled): %v (ratio: %.2f)",
		duration1, duration2, ratio)
}

// Helper functions

func containsAny(text string, substrings []string) bool {
	for _, substr := range substrings {
		if strings.Contains(strings.ToLower(text), strings.ToLower(substr)) {
			return true
		}
	}
	return false
}
