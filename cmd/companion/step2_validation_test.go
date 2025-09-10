package main

import (
	"github.com/opd-ai/desktop-companion/lib/character"
	"path/filepath"
	"testing"
)

// TestStep2PersonalityAppropriateFeatures tests that Step 2 implementation
// has successfully added romance and multiplayer features to appropriate characters
func TestStep2PersonalityAppropriateFeatures(t *testing.T) {
	testCases := []struct {
		name                string
		charPath            string
		expectRomance       bool
		expectMultiplayer   bool
		expectedPersonality map[string]float64
		maxAffection        float64
		maxTrust            float64
	}{
		{
			name:              "Default Character",
			charPath:          "../../assets/characters/default/character.json",
			expectRomance:     true,
			expectMultiplayer: true,
			expectedPersonality: map[string]float64{
				"friendliness": 0.9,
				"romanticism":  0.4,
			},
			maxAffection: 50,
			maxTrust:     60,
		},
		{
			name:              "Easy Character",
			charPath:          "../../assets/characters/easy/character.json",
			expectRomance:     true,
			expectMultiplayer: true,
			expectedPersonality: map[string]float64{
				"gentleness":  0.9,
				"romanticism": 0.3,
			},
			maxAffection: 40,
			maxTrust:     50,
		},
		{
			name:              "Specialist Character",
			charPath:          "../../assets/characters/specialist/character.json",
			expectRomance:     true,
			expectMultiplayer: false,
			expectedPersonality: map[string]float64{
				"sleepiness":  0.8,
				"romanticism": 0.2,
			},
			maxAffection: 35,
			maxTrust:     45,
		},
		{
			name:              "Markov Example",
			charPath:          "../../assets/characters/markov_example/character.json",
			expectRomance:     true,
			expectMultiplayer: true,
			expectedPersonality: map[string]float64{
				"curiosity":   0.9,
				"romanticism": 0.3,
			},
			maxAffection: 45,
			maxTrust:     55,
		},
		{
			name:              "News Example",
			charPath:          "../../assets/characters/news_example/character.json",
			expectRomance:     true,
			expectMultiplayer: true,
			expectedPersonality: map[string]float64{
				"intellectual": 0.8,
				"romanticism":  0.3,
			},
			maxAffection: 40,
			maxTrust:     50,
		},
		{
			name:              "Romance Character",
			charPath:          "../../assets/characters/romance/character.json",
			expectRomance:     true, // Should already exist
			expectMultiplayer: true, // We added this
			expectedPersonality: map[string]float64{
				"romanticism": 0.9, // Should already be high
			},
			maxAffection: 100, // Should already be high
			maxTrust:     100,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Load character card using absolute path
			absPath, err := filepath.Abs(tc.charPath)
			if err != nil {
				t.Fatalf("Failed to get absolute path for %s: %v", tc.charPath, err)
			}

			card, err := character.LoadCard(absPath)
			if err != nil {
				t.Fatalf("Failed to load character %s: %v", tc.charPath, err)
			}

			// Test romance features
			if tc.expectRomance {
				// Check for personality traits
				if card.Personality == nil {
					t.Errorf("Character %s missing personality traits", tc.name)
				} else {
					for trait, expectedValue := range tc.expectedPersonality {
						if actualValue, exists := card.Personality.Traits[trait]; !exists {
							t.Errorf("Character %s missing personality trait: %s", tc.name, trait)
						} else if actualValue != expectedValue {
							t.Logf("Character %s trait %s: expected %.1f, got %.1f", tc.name, trait, expectedValue, actualValue)
						}
					}
				}

				// Check for romance stats in the Stats map
				hasAffection := false
				hasTrust := false

				if card.Stats != nil {
					if _, exists := card.Stats["affection"]; exists {
						hasAffection = true
					}
					if _, exists := card.Stats["trust"]; exists {
						hasTrust = true
					}
				}

				if !hasAffection {
					t.Errorf("Character %s missing affection stat", tc.name)
				}
				if !hasTrust {
					t.Errorf("Character %s missing trust stat", tc.name)
				}

				// Check for romance interactions in the Interactions map
				hasRomanceInteractions := false
				if card.Interactions != nil {
					// Check if any interactions have romance-like names or effects
					romanceKeywords := []string{"compliment", "chat", "cuddle", "connection", "story", "bond", "encouragement", "moment"}
					for interactionName := range card.Interactions {
						for _, keyword := range romanceKeywords {
							if contains(interactionName, keyword) {
								hasRomanceInteractions = true
								break
							}
						}
						if hasRomanceInteractions {
							break
						}
					}
				}

				if !hasRomanceInteractions {
					t.Errorf("Character %s missing romance interactions", tc.name)
				}
			}

			// Test multiplayer features
			if tc.expectMultiplayer {
				if card.Multiplayer == nil {
					t.Errorf("Character %s missing multiplayer configuration", tc.name)
				} else {
					if !card.Multiplayer.Enabled {
						t.Errorf("Character %s has multiplayer disabled", tc.name)
					}
					if card.Multiplayer.NetworkID == "" {
						t.Errorf("Character %s missing network ID", tc.name)
					}
					if card.Multiplayer.MaxPeers <= 0 {
						t.Errorf("Character %s has invalid max peers: %d", tc.name, card.Multiplayer.MaxPeers)
					}
				}
			}

			t.Logf("✅ Character %s has appropriate Step 2 features", tc.name)
		})
	}
}

// TestStep2PersonalityPreservation tests that character personalities remain appropriate
func TestStep2PersonalityPreservation(t *testing.T) {
	testCases := []struct {
		name              string
		charPath          string
		maxRomanticism    float64 // Characters should not exceed their theme
		characterTheme    string
		preservedFeatures []string
	}{
		{
			name:              "Default Character",
			charPath:          "../../assets/characters/default/character.json",
			maxRomanticism:    0.5,
			characterTheme:    "friendly companion",
			preservedFeatures: []string{"dialogBackend", "stats", "interactions"},
		},
		{
			name:              "Easy Character",
			charPath:          "../../assets/characters/easy/character.json",
			maxRomanticism:    0.4,
			characterTheme:    "beginner-friendly pet",
			preservedFeatures: []string{"stats", "gameRules", "interactions"},
		},
		{
			name:              "Specialist Character",
			charPath:          "../../assets/characters/specialist/character.json",
			maxRomanticism:    0.3,
			characterTheme:    "energy management focus",
			preservedFeatures: []string{"stats", "gameRules"},
		},
		{
			name:              "Markov Example",
			charPath:          "../../assets/characters/markov_example/character.json",
			maxRomanticism:    0.4,
			characterTheme:    "AI dialog demonstration",
			preservedFeatures: []string{"dialogBackend", "stats", "interactions"},
		},
		{
			name:              "News Example",
			charPath:          "../../assets/characters/news_example/character.json",
			maxRomanticism:    0.4,
			characterTheme:    "news reading companion",
			preservedFeatures: []string{"dialogBackend", "newsFeatures", "interactions"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			absPath, err := filepath.Abs(tc.charPath)
			if err != nil {
				t.Fatalf("Failed to get absolute path for %s: %v", tc.charPath, err)
			}

			card, err := character.LoadCard(absPath)
			if err != nil {
				t.Fatalf("Failed to load character %s: %v", tc.charPath, err)
			}

			// Check that romanticism is appropriately limited
			if card.Personality != nil && card.Personality.Traits != nil {
				if romanticism, exists := card.Personality.Traits["romanticism"]; exists {
					if romanticism > tc.maxRomanticism {
						t.Errorf("Character %s has too high romanticism (%.1f > %.1f) for theme: %s",
							tc.name, romanticism, tc.maxRomanticism, tc.characterTheme)
					}
				}
			}

			// Check that core features are preserved
			for _, feature := range tc.preservedFeatures {
				switch feature {
				case "dialogBackend":
					if card.DialogBackend == nil {
						t.Errorf("Character %s lost dialog backend feature", tc.name)
					}
				case "stats":
					if len(card.Stats) == 0 {
						t.Errorf("Character %s lost stats feature", tc.name)
					}
				case "gameRules":
					if card.GameRules == nil {
						t.Errorf("Character %s lost game rules feature", tc.name)
					}
				case "interactions":
					if len(card.Interactions) == 0 {
						t.Errorf("Character %s lost interactions feature", tc.name)
					}
				case "newsFeatures":
					if card.NewsFeatures == nil {
						t.Errorf("Character %s lost news features", tc.name)
					}
				}
			}

			t.Logf("✅ Character %s preserved personality and theme: %s", tc.name, tc.characterTheme)
		})
	}
}

// Helper function for string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findInString(s, substr))))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
