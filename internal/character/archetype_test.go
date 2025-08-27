package character

import (
	"path/filepath"
	"testing"
)

func TestCharacterArchetypes(t *testing.T) {
	tests := []struct {
		name                string
		characterFile       string
		expectedTraits      map[string]float64
		expectedProgression int // Number of progression levels
	}{
		{
			name:          "Tsundere Character",
			characterFile: "../../assets/characters/tsundere/character.json",
			expectedTraits: map[string]float64{
				"shyness":          0.9,
				"romanticism":      0.8,
				"jealousy_prone":   0.7,
				"trust_difficulty": 0.8,
			},
			expectedProgression: 5,
		},
		{
			name:          "Flirty Extrovert Character",
			characterFile: "../../assets/characters/flirty/character.json",
			expectedTraits: map[string]float64{
				"shyness":     0.1,
				"romanticism": 0.9,
				"flirtiness":  0.9,
			},
			expectedProgression: 5,
		},
		{
			name:          "Slow Burn Character",
			characterFile: "../../assets/characters/slow_burn/character.json",
			expectedTraits: map[string]float64{
				"trust_difficulty":         0.9,
				"affection_responsiveness": 0.4,
				"jealousy_prone":           0.1,
			},
			expectedProgression: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Load character card
			card, err := LoadCard(tt.characterFile)
			if err != nil {
				t.Fatalf("Failed to load character card: %v", err)
			}

			// Verify romance features are enabled
			if !card.HasRomanceFeatures() {
				t.Error("Character should have romance features enabled")
			}

			// Check personality traits
			if card.Personality == nil {
				t.Fatal("Character should have personality configuration")
			}

			for trait, expected := range tt.expectedTraits {
				actual := card.GetPersonalityTrait(trait)
				if actual != expected {
					t.Errorf("Expected %s trait %.1f, got %.1f", trait, expected, actual)
				}
			}

			// Check progression levels
			if card.Progression == nil {
				t.Fatal("Character should have progression configuration")
			}

			if len(card.Progression.Levels) != tt.expectedProgression {
				t.Errorf("Expected %d progression levels, got %d",
					tt.expectedProgression, len(card.Progression.Levels))
			}

			// Verify romance stats exist
			romanceStats := []string{"affection", "trust", "intimacy", "jealousy"}
			for _, stat := range romanceStats {
				if _, exists := card.Stats[stat]; !exists {
					t.Errorf("Romance stat '%s' should be configured", stat)
				}
			}

			// Verify romance interactions exist
			romanceInteractions := []string{"compliment", "give_gift", "deep_conversation"}
			for _, interaction := range romanceInteractions {
				if _, exists := card.Interactions[interaction]; !exists {
					t.Errorf("Romance interaction '%s' should be configured", interaction)
				}
			}

			// Test character creation with this card
			basePath := filepath.Dir(tt.characterFile)
			character, err := New(card, basePath)
			if err != nil {
				t.Fatalf("Failed to create character: %v", err)
			}

			// Verify character has game state
			if character.GetGameState() == nil {
				t.Error("Character should have game state initialized")
			}

			// Test romance interaction
			response := character.HandleRomanceInteraction("compliment")
			// Note: Response might be empty if requirements not met, which is fine
			if response != "" {
				t.Logf("Romance interaction response: %s", response)
			}
		})
	}
}

func TestArchetypePersonalityDifferences(t *testing.T) {
	// Load all three archetypes
	tsundere, err := LoadCard("../../assets/characters/tsundere/character.json")
	if err != nil {
		t.Fatalf("Failed to load tsundere: %v", err)
	}

	flirty, err := LoadCard("../../assets/characters/flirty/character.json")
	if err != nil {
		t.Fatalf("Failed to load flirty: %v", err)
	}

	slowBurn, err := LoadCard("../../assets/characters/slow_burn/character.json")
	if err != nil {
		t.Fatalf("Failed to load slow burn: %v", err)
	}

	// Test that personalities are distinctly different
	t.Run("shyness_differences", func(t *testing.T) {
		tsundereShyness := tsundere.GetPersonalityTrait("shyness")
		flirtyShyness := flirty.GetPersonalityTrait("shyness")
		slowBurnShyness := slowBurn.GetPersonalityTrait("shyness")

		// Tsundere should be shyest, flirty least shy
		if tsundereShyness <= flirtyShyness {
			t.Error("Tsundere should be more shy than flirty extrovert")
		}
		if flirtyShyness >= slowBurnShyness {
			t.Error("Flirty extrovert should be less shy than slow burn")
		}
	})

	t.Run("trust_difficulty_differences", func(t *testing.T) {
		tsunsdereTrust := tsundere.GetPersonalityTrait("trust_difficulty")
		flirtyTrust := flirty.GetPersonalityTrait("trust_difficulty")
		slowBurnTrust := slowBurn.GetPersonalityTrait("trust_difficulty")

		// Slow burn should have highest trust difficulty
		if slowBurnTrust <= tsunsdereTrust || slowBurnTrust <= flirtyTrust {
			t.Error("Slow burn should have highest trust difficulty")
		}

		// Flirty should have lowest trust difficulty
		if flirtyTrust >= tsunsdereTrust || flirtyTrust >= slowBurnTrust {
			t.Error("Flirty extrovert should have lowest trust difficulty")
		}
	})

	t.Run("compatibility_bonuses", func(t *testing.T) {
		// Check that slow burn has highest conversation bonus
		slowBurnConversation := slowBurn.GetCompatibilityModifier("conversation_lover")
		flirtyConversation := flirty.GetCompatibilityModifier("conversation_lover")

		if slowBurnConversation <= flirtyConversation {
			t.Error("Slow burn should have higher conversation compatibility than flirty")
		}

		// Check that flirty has highest gift appreciation
		flirtyGifts := flirty.GetCompatibilityModifier("gift_appreciation")
		tsundereGifts := tsundere.GetCompatibilityModifier("gift_appreciation")

		if flirtyGifts <= tsundereGifts {
			t.Error("Flirty extrovert should have higher gift appreciation than tsundere")
		}
	})
}

func TestArchetypeStartingStats(t *testing.T) {
	tests := []struct {
		name          string
		characterFile string
		statChecks    map[string]struct {
			expectedValue float64
			description   string
		}
	}{
		{
			name:          "Tsundere Starting Stats",
			characterFile: "../../assets/characters/tsundere/character.json",
			statChecks: map[string]struct {
				expectedValue float64
				description   string
			}{
				"affection": {0, "should start with no affection"},
				"trust":     {5, "should start with minimal trust"},
				"jealousy":  {10, "should start with some jealousy"},
			},
		},
		{
			name:          "Flirty Starting Stats",
			characterFile: "../../assets/characters/flirty/character.json",
			statChecks: map[string]struct {
				expectedValue float64
				description   string
			}{
				"affection": {25, "should start with some affection"},
				"trust":     {40, "should start with moderate trust"},
				"jealousy":  {0, "should start with no jealousy"},
			},
		},
		{
			name:          "Slow Burn Starting Stats",
			characterFile: "../../assets/characters/slow_burn/character.json",
			statChecks: map[string]struct {
				expectedValue float64
				description   string
			}{
				"affection": {0, "should start with no affection"},
				"trust":     {10, "should start with minimal trust"},
				"jealousy":  {0, "should start with no jealousy"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card, err := LoadCard(tt.characterFile)
			if err != nil {
				t.Fatalf("Failed to load character card: %v", err)
			}

			basePath := filepath.Dir(tt.characterFile)
			character, err := New(card, basePath)
			if err != nil {
				t.Fatalf("Failed to create character: %v", err)
			}

			gameState := character.GetGameState()
			if gameState == nil {
				t.Fatal("Character should have game state")
			}

			for statName, check := range tt.statChecks {
				actual := gameState.GetStat(statName)
				if actual != check.expectedValue {
					t.Errorf("Stat %s %s: expected %.0f, got %.0f",
						statName, check.description, check.expectedValue, actual)
				}
			}
		})
	}
}

func TestArchetypeInteractionCooldowns(t *testing.T) {
	// Test that different archetypes have different interaction pacing
	tsundere, err := LoadCard("../../assets/characters/tsundere/character.json")
	if err != nil {
		t.Fatalf("Failed to load tsundere: %v", err)
	}

	flirty, err := LoadCard("../../assets/characters/flirty/character.json")
	if err != nil {
		t.Fatalf("Failed to load flirty: %v", err)
	}

	slowBurn, err := LoadCard("../../assets/characters/slow_burn/character.json")
	if err != nil {
		t.Fatalf("Failed to load slow burn: %v", err)
	}

	// Check compliment cooldowns (all should have this interaction)
	tsundereCompliment := tsundere.Interactions["compliment"].Cooldown
	flirtyCompliment := flirty.Interactions["compliment"].Cooldown
	slowBurnCompliment := slowBurn.Interactions["compliment"].Cooldown

	// Flirty should have shortest cooldown, slow burn should have longest
	if flirtyCompliment >= tsundereCompliment {
		t.Error("Flirty extrovert should have shorter compliment cooldown than tsundere")
	}
	if slowBurnCompliment <= flirtyCompliment {
		t.Error("Slow burn should have longer compliment cooldown than flirty")
	}

	t.Logf("Compliment cooldowns - Tsundere: %ds, Flirty: %ds, Slow Burn: %ds",
		tsundereCompliment, flirtyCompliment, slowBurnCompliment)
}
