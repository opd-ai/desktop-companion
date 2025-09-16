package assets

import (
	"testing"

	"github.com/opd-ai/desktop-companion/lib/character"
)

func TestAssetGenerationWorkflow(t *testing.T) {
	// Test creating a character with asset generation config
	card := &character.CharacterCard{
		Name:        "Test Character",
		Description: "A test character for asset generation",
		Animations: map[string]string{
			"idle":    "animations/idle.gif",
			"talking": "animations/talking.gif",
			"happy":   "animations/happy.gif",
			"sad":     "animations/sad.gif",
		},
		AssetGeneration: character.DefaultAssetGenerationConfig(),
	}

	// Validate the character card
	if err := card.Validate(); err != nil {
		t.Fatalf("Character card validation failed: %v", err)
	}

	// Test that asset generation config exists and is valid
	if card.AssetGeneration == nil {
		t.Fatal("AssetGeneration config is nil")
	}

	// Test that required animation mappings exist
	expectedStates := []string{"idle", "talking", "happy", "sad"}
	for _, state := range expectedStates {
		if _, exists := card.AssetGeneration.AnimationMappings[state]; !exists {
			t.Errorf("Missing animation mapping for state: %s", state)
		}
	}

	// Test generation settings are valid
	settings := card.AssetGeneration.GenerationSettings
	if settings.Model == "" {
		t.Error("Model not specified in generation settings")
	}
	if settings.ArtStyle == "" {
		t.Error("ArtStyle not specified in generation settings")
	}
	if settings.Resolution.Width <= 0 || settings.Resolution.Height <= 0 {
		t.Error("Invalid resolution in generation settings")
	}

	// Test asset generator creation
	config := DefaultGeneratorConfig()
	generator, err := NewAssetGenerator(config)
	if err != nil {
		t.Fatalf("Failed to create asset generator: %v", err)
	}
	if generator == nil {
		t.Fatal("Asset generator is nil")
	}

	// Test configuration validation
	if err := character.ValidateAssetGenerationConfig(card.AssetGeneration); err != nil {
		t.Fatalf("Asset generation config validation failed: %v", err)
	}

	t.Logf("Asset generation workflow test passed")
	t.Logf("Character: %s", card.Name)
	t.Logf("Model: %s", settings.Model)
	t.Logf("Art Style: %s", settings.ArtStyle)
	t.Logf("Resolution: %dx%d", settings.Resolution.Width, settings.Resolution.Height)
}
