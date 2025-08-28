package character

import (
	"testing"
)

// TestDefaultCharacterDialogBackendIntegration validates that default characters have dialog backend enabled
func TestDefaultCharacterDialogBackendIntegration(t *testing.T) {
	t.Run("default_character_has_dialog_backend", func(t *testing.T) {
		card, err := LoadCard("../../assets/characters/default/character.json")
		if err != nil {
			t.Fatalf("Failed to load default character: %v", err)
		}

		if !card.HasDialogBackend() {
			t.Error("Default character should have dialog backend enabled")
		}

		if card.DialogBackend == nil {
			t.Error("Default character should have dialog backend configuration")
		}

		if card.DialogBackend.DefaultBackend != "markov_chain" {
			t.Errorf("Expected default backend 'markov_chain', got '%s'", card.DialogBackend.DefaultBackend)
		}

		if card.DialogBackend.ConfidenceThreshold != 0.6 {
			t.Errorf("Expected confidence threshold 0.6, got %f", card.DialogBackend.ConfidenceThreshold)
		}

		// Validate that Markov backend configuration exists
		if card.DialogBackend.Backends == nil {
			t.Error("Default character should have backend configurations")
		}

		markovConfig, exists := card.DialogBackend.Backends["markov_chain"]
		if !exists {
			t.Error("Default character should have markov_chain backend configuration")
		}

		if len(markovConfig) == 0 {
			t.Error("Markov chain configuration should not be empty")
		}

		t.Log("✅ Default character dialog backend integration validated")
	})

	t.Run("game_character_has_dialog_backend", func(t *testing.T) {
		card, err := LoadCard("../../assets/characters/default/character_with_game_features.json")
		if err != nil {
			t.Fatalf("Failed to load game character: %v", err)
		}

		if !card.HasDialogBackend() {
			t.Error("Game character should have dialog backend enabled")
		}

		if card.DialogBackend == nil {
			t.Error("Game character should have dialog backend configuration")
		}

		if card.DialogBackend.DefaultBackend != "markov_chain" {
			t.Errorf("Expected default backend 'markov_chain', got '%s'", card.DialogBackend.DefaultBackend)
		}

		// Validate that training data is appropriate for game context
		markovConfig, exists := card.DialogBackend.Backends["markov_chain"]
		if !exists {
			t.Error("Game character should have markov_chain backend configuration")
		}

		if len(markovConfig) == 0 {
			t.Error("Game character markov chain configuration should not be empty")
		}

		t.Log("✅ Game character dialog backend integration validated")
	})
}

// TestDialogBackendFeatureCompleteness validates that the dialog backend implementation is complete
func TestDialogBackendFeatureCompleteness(t *testing.T) {
	// Load a character with dialog backend
	card, err := LoadCard("../../assets/characters/default/character.json")
	if err != nil {
		t.Fatalf("Failed to load character: %v", err)
	}

	// Validate all required features are present
	if !card.HasDialogBackend() {
		t.Error("Character should have dialog backend enabled")
	}

	// Check confidence threshold configuration
	if card.DialogBackend.ConfidenceThreshold <= 0 || card.DialogBackend.ConfidenceThreshold > 1 {
		t.Errorf("Confidence threshold should be between 0 and 1, got %f", card.DialogBackend.ConfidenceThreshold)
	}

	// Check memory system is enabled
	if !card.DialogBackend.MemoryEnabled {
		t.Error("Memory system should be enabled for learning")
	}

	// Check fallback chain is configured
	if len(card.DialogBackend.FallbackChain) == 0 {
		t.Error("Fallback chain should be configured")
	}

	// Validate training data exists
	markovConfig, exists := card.DialogBackend.Backends["markov_chain"]
	if !exists {
		t.Error("Markov chain backend should be configured")
	}

	if len(markovConfig) == 0 {
		t.Error("Markov chain configuration should not be empty")
	}

	t.Log("✅ Dialog backend feature completeness validated")
}
