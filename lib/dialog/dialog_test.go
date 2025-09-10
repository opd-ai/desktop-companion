package dialog

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSimpleRandomBackendInterface(t *testing.T) {
	// Test that SimpleRandomBackend implements DialogBackend interface
	var _ DialogBackend = (*SimpleRandomBackend)(nil)

	backend := NewSimpleRandomBackend()

	// Test initialization
	config := map[string]interface{}{
		"type":                 "basic",
		"personalityInfluence": 0.5,
		"responseVariation":    0.3,
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		t.Fatal("Failed to marshal config:", err)
	}

	if err := backend.Initialize(json.RawMessage(configJSON)); err != nil {
		t.Fatal("Failed to initialize backend:", err)
	}

	// Test CanHandle
	context := DialogContext{
		Trigger:       "click",
		InteractionID: "test_001",
		Timestamp:     time.Now(),
		CurrentStats: map[string]float64{
			"happiness": 75.0,
		},
		PersonalityTraits: map[string]float64{
			"shyness": 0.3,
		},
		FallbackResponses: []string{"Hello!"},
	}

	if !backend.CanHandle(context) {
		t.Error("SimpleRandomBackend should be able to handle any context")
	}

	// Test GenerateResponse
	response, err := backend.GenerateResponse(context)
	if err != nil {
		t.Fatal("Failed to generate response:", err)
	}

	if response.Text == "" {
		t.Error("Generated response should not be empty")
	}

	if response.Confidence <= 0 {
		t.Error("Response confidence should be positive")
	}

	// Test GetBackendInfo
	info := backend.GetBackendInfo()
	if info.Name != "simple_random" {
		t.Errorf("Expected backend name 'simple_random', got %s", info.Name)
	}

	if info.Version == "" {
		t.Error("Backend version should not be empty")
	}
}

func TestMarkovChainBackendInterface(t *testing.T) {
	// Test that MarkovChainBackend implements DialogBackend interface
	var _ DialogBackend = (*MarkovChainBackend)(nil)

	backend := NewMarkovChainBackend()

	// Test initialization
	config := map[string]interface{}{
		"chainOrder":     2,
		"minWords":       3,
		"maxWords":       10,
		"temperatureMin": 0.3,
		"temperatureMax": 0.8,
		"trainingData": []string{
			"Hello there!",
			"How are you doing?",
			"I hope you're well.",
		},
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		t.Fatal("Failed to marshal config:", err)
	}

	if err := backend.Initialize(json.RawMessage(configJSON)); err != nil {
		t.Fatal("Failed to initialize backend:", err)
	}

	// Test GetBackendInfo
	info := backend.GetBackendInfo()
	if info.Name != "markov_chain" {
		t.Errorf("Expected backend name 'markov_chain', got %s", info.Name)
	}
}

func TestDialogManager(t *testing.T) {
	manager := NewDialogManager(false)

	// Create and register backends
	simpleBackend := NewSimpleRandomBackend()
	if err := simpleBackend.Initialize(json.RawMessage(`{}`)); err != nil {
		t.Fatal("Failed to initialize simple backend:", err)
	}

	manager.RegisterBackend("simple", simpleBackend)

	// Test backend registration
	backends := manager.GetRegisteredBackends()
	if len(backends) != 1 || backends[0] != "simple" {
		t.Errorf("Expected ['simple'], got %v", backends)
	}

	// Test setting default backend
	if err := manager.SetDefaultBackend("simple"); err != nil {
		t.Fatal("Failed to set default backend:", err)
	}

	// Test dialog generation
	context := DialogContext{
		Trigger:           "click",
		InteractionID:     "test_001",
		Timestamp:         time.Now(),
		FallbackResponses: []string{"Hello!"},
	}

	response, err := manager.GenerateDialog(context)
	if err != nil {
		t.Fatal("Failed to generate dialog:", err)
	}

	if response.Text == "" {
		t.Error("Generated response should not be empty")
	}
}
