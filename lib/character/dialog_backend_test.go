package character

import (
	"github.com/opd-ai/desktop-companion/internal/dialog"
	"encoding/json"
	"testing"
)

func TestDialogBackendValidation(t *testing.T) {
	tests := []struct {
		name    string
		card    CharacterCard
		wantErr bool
	}{
		{
			name: "character without dialog backend",
			card: CharacterCard{
				Name:        "Test",
				Description: "Test character",
				Animations:  map[string]string{"idle": "idle.gif", "talking": "talking.gif"},
				Dialogs:     []Dialog{{Trigger: "click", Responses: []string{"Hi!"}, Animation: "idle"}},
				Behavior:    Behavior{IdleTimeout: 10, DefaultSize: 128},
			},
			wantErr: false,
		},
		{
			name: "character with valid dialog backend",
			card: CharacterCard{
				Name:        "Enhanced Test",
				Description: "Test character with dialog backend",
				Animations:  map[string]string{"idle": "idle.gif", "talking": "talking.gif"},
				Dialogs:     []Dialog{{Trigger: "click", Responses: []string{"Hi!"}, Animation: "idle"}},
				Behavior:    Behavior{IdleTimeout: 10, DefaultSize: 128},
				DialogBackend: &dialog.DialogBackendConfig{
					Enabled:             true,
					DefaultBackend:      "simple_random",
					ConfidenceThreshold: 0.5,
					Backends: map[string]json.RawMessage{
						"simple_random": json.RawMessage(`{"type": "basic"}`),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "character with invalid dialog backend (missing default)",
			card: CharacterCard{
				Name:        "Invalid Test",
				Description: "Test character with invalid dialog backend",
				Animations:  map[string]string{"idle": "idle.gif", "talking": "talking.gif"},
				Dialogs:     []Dialog{{Trigger: "click", Responses: []string{"Hi!"}, Animation: "idle"}},
				Behavior:    Behavior{IdleTimeout: 10, DefaultSize: 128},
				DialogBackend: &dialog.DialogBackendConfig{
					Enabled:             true,
					ConfidenceThreshold: 0.5,
					// Missing DefaultBackend
				},
			},
			wantErr: true,
		},
		{
			name: "character with invalid confidence threshold",
			card: CharacterCard{
				Name:        "Invalid Confidence",
				Description: "Test character with invalid confidence threshold",
				Animations:  map[string]string{"idle": "idle.gif", "talking": "talking.gif"},
				Dialogs:     []Dialog{{Trigger: "click", Responses: []string{"Hi!"}, Animation: "idle"}},
				Behavior:    Behavior{IdleTimeout: 10, DefaultSize: 128},
				DialogBackend: &dialog.DialogBackendConfig{
					Enabled:             true,
					DefaultBackend:      "simple_random",
					ConfidenceThreshold: 1.5, // Invalid: > 1.0
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.card.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("CharacterCard.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHasDialogBackend(t *testing.T) {
	tests := []struct {
		name     string
		card     CharacterCard
		expected bool
	}{
		{
			name:     "no dialog backend",
			card:     CharacterCard{},
			expected: false,
		},
		{
			name: "dialog backend disabled",
			card: CharacterCard{
				DialogBackend: &dialog.DialogBackendConfig{
					Enabled: false,
				},
			},
			expected: false,
		},
		{
			name: "dialog backend enabled",
			card: CharacterCard{
				DialogBackend: &dialog.DialogBackendConfig{
					Enabled: true,
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.card.HasDialogBackend()
			if result != tt.expected {
				t.Errorf("CharacterCard.HasDialogBackend() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestSimpleRandomBackend(t *testing.T) {
	backend := dialog.NewSimpleRandomBackend()

	// Test configuration parsing
	config := json.RawMessage(`{
		"type": "basic",
		"personalityInfluence": 0.3,
		"responseVariation": 0.2
	}`)

	err := backend.Initialize(config)
	if err != nil {
		t.Fatalf("Failed to initialize backend: %v", err)
	}

	// Test backend info
	info := backend.GetBackendInfo()
	if info.Name != "simple_random" {
		t.Errorf("Expected backend name 'simple_random', got '%s'", info.Name)
	}

	// Test CanHandle (should always return true for simple random)
	context := dialog.DialogContext{Trigger: "click"}
	if !backend.CanHandle(context) {
		t.Error("Simple random backend should handle any context")
	}

	// Test response generation with fallback
	context.FallbackResponses = []string{"Hello!", "Hi there!"}
	response, err := backend.GenerateResponse(context)
	if err != nil {
		t.Fatalf("Failed to generate response: %v", err)
	}

	if response.Text == "" {
		t.Error("Expected non-empty response text")
	}

	if response.Confidence <= 0 || response.Confidence > 1 {
		t.Errorf("Expected confidence between 0 and 1, got %f", response.Confidence)
	}
}

func TestSimpleRandomBackendValidation(t *testing.T) {
	backend := dialog.NewSimpleRandomBackend()

	// Test invalid configuration
	invalidConfig := json.RawMessage(`{
		"personalityInfluence": 1.5
	}`)

	err := backend.Initialize(invalidConfig)
	if err == nil {
		t.Error("Expected error for invalid configuration, got nil")
	}

	// Test valid configuration
	validConfig := json.RawMessage(`{
		"type": "basic",
		"personalityInfluence": 0.5,
		"responseVariation": 0.3
	}`)

	err = backend.Initialize(validConfig)
	if err != nil {
		t.Errorf("Expected no error for valid configuration, got: %v", err)
	}
}

func TestMarkovBackendInitialization(t *testing.T) {
	backend := dialog.NewMarkovChainBackend()

	// Test basic configuration
	config := json.RawMessage(`{
		"chainOrder": 2,
		"minWords": 3,
		"maxWords": 10,
		"temperatureMin": 0.4,
		"temperatureMax": 0.7,
		"trainingData": [
			"Hello! How are you today?",
			"I'm happy to see you!"
		]
	}`)

	err := backend.Initialize(config)
	if err != nil {
		t.Fatalf("Failed to initialize Markov backend: %v", err)
	}

	// Test backend info
	info := backend.GetBackendInfo()
	if info.Name != "markov_chain" {
		t.Errorf("Expected backend name 'markov_chain', got '%s'", info.Name)
	}

	// Test CanHandle (should return true if backend has data)
	context := dialog.DialogContext{Trigger: "click"}
	canHandle := backend.CanHandle(context)
	t.Logf("Markov backend can handle context: %v", canHandle)
}

func TestDialogManagerBasicOperation(t *testing.T) {
	manager := dialog.NewDialogManager(false)

	// Register backends
	manager.RegisterBackend("simple_random", dialog.NewSimpleRandomBackend())
	manager.RegisterBackend("markov_chain", dialog.NewMarkovChainBackend())

	// Test getting registered backends
	backends := manager.GetRegisteredBackends()
	if len(backends) != 2 {
		t.Errorf("Expected 2 registered backends, got %d", len(backends))
	}

	// Set default backend
	err := manager.SetDefaultBackend("simple_random")
	if err != nil {
		t.Fatalf("Failed to set default backend: %v", err)
	}

	// Test setting invalid default backend
	err = manager.SetDefaultBackend("nonexistent")
	if err == nil {
		t.Error("Expected error when setting nonexistent backend as default")
	}

	// Set fallback chain
	err = manager.SetFallbackChain([]string{"markov_chain"})
	if err != nil {
		t.Fatalf("Failed to set fallback chain: %v", err)
	}

	// Test setting invalid fallback chain
	err = manager.SetFallbackChain([]string{"nonexistent"})
	if err == nil {
		t.Error("Expected error when setting nonexistent backend in fallback chain")
	}
}

func TestDialogBackendConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  dialog.DialogBackendConfig
		wantErr bool
	}{
		{
			name: "disabled backend (should be valid)",
			config: dialog.DialogBackendConfig{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "valid enabled backend",
			config: dialog.DialogBackendConfig{
				Enabled:             true,
				DefaultBackend:      "simple_random",
				ConfidenceThreshold: 0.5,
			},
			wantErr: false,
		},
		{
			name: "missing default backend",
			config: dialog.DialogBackendConfig{
				Enabled: true,
				// Missing DefaultBackend
				ConfidenceThreshold: 0.5,
			},
			wantErr: true,
		},
		{
			name: "invalid confidence threshold (negative)",
			config: dialog.DialogBackendConfig{
				Enabled:             true,
				DefaultBackend:      "simple_random",
				ConfidenceThreshold: -0.1,
			},
			wantErr: true,
		},
		{
			name: "invalid confidence threshold (too high)",
			config: dialog.DialogBackendConfig{
				Enabled:             true,
				DefaultBackend:      "simple_random",
				ConfidenceThreshold: 1.1,
			},
			wantErr: true,
		},
		{
			name: "invalid response timeout (negative)",
			config: dialog.DialogBackendConfig{
				Enabled:             true,
				DefaultBackend:      "simple_random",
				ConfidenceThreshold: 0.5,
				ResponseTimeout:     -100,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dialog.ValidateBackendConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBackendConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
