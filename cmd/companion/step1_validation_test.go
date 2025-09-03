package main

import (
	"github.com/opd-ai/desktop-companion/internal/character"
	"path/filepath"
	"testing"
)

// TestStep1BasicGameFeatures tests that the Step 1 implementation (Basic Game Features)
// has been successfully added to Default, Markov Example, and News Example characters
func TestStep1BasicGameFeatures(t *testing.T) {
	testCases := []struct {
		name     string
		charPath string
	}{
		{"Default Character", "../../assets/characters/default/character.json"},
		{"Markov Example", "../../assets/characters/markov_example/character.json"},
		{"News Example", "../../assets/characters/news_example/character.json"},
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

			// Verify basic game features are present
			if card.Stats == nil {
				t.Errorf("Character %s missing stats system", tc.name)
			} else {
				// Check required stats
				if _, hasHappiness := card.Stats["happiness"]; !hasHappiness {
					t.Errorf("Character %s missing happiness stat", tc.name)
				}
				if _, hasEnergy := card.Stats["energy"]; !hasEnergy {
					t.Errorf("Character %s missing energy stat", tc.name)
				}
			}

			// Verify game rules are present
			if card.GameRules == nil {
				t.Errorf("Character %s missing game rules", tc.name)
			} else {
				// Check required game rules
				if card.GameRules.StatsDecayInterval == 0 {
					t.Errorf("Character %s missing stats decay interval", tc.name)
				}
				if card.GameRules.AutoSaveInterval == 0 {
					t.Errorf("Character %s missing auto save interval", tc.name)
				}
				if card.GameRules.DeathEnabled {
					t.Errorf("Character %s should have death disabled for Step 1", tc.name)
				}
			}

			// Verify interactions are present
			if card.Interactions == nil {
				t.Errorf("Character %s missing interactions system", tc.name)
			} else {
				// Check for pet interaction
				if _, hasPet := card.Interactions["pet"]; !hasPet {
					t.Errorf("Character %s missing pet interaction", tc.name)
				}
			}

			// Verify dialog backend is present (all characters should have this)
			if card.DialogBackend == nil {
				t.Errorf("Character %s missing dialog backend", tc.name)
			}

			t.Logf("✅ Character %s successfully has basic game features", tc.name)
		})
	}
}

// TestStep1BackwardCompatibility tests that existing functionality is preserved
func TestStep1BackwardCompatibility(t *testing.T) {
	testCases := []struct {
		name     string
		charPath string
	}{
		{"Default Character", "../../assets/characters/default/character.json"},
		{"Markov Example", "../../assets/characters/markov_example/character.json"},
		{"News Example", "../../assets/characters/news_example/character.json"},
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

			// Verify core features are preserved
			if card.Name == "" {
				t.Errorf("Character %s missing name", tc.name)
			}
			if card.Description == "" {
				t.Errorf("Character %s missing description", tc.name)
			}
			if len(card.Animations) == 0 {
				t.Errorf("Character %s missing animations", tc.name)
			}
			if len(card.Dialogs) == 0 {
				t.Errorf("Character %s missing dialogs", tc.name)
			}
			if card.Behavior.DefaultSize == 0 {
				t.Errorf("Character %s missing default size in behavior", tc.name)
			}

			t.Logf("✅ Character %s maintains backward compatibility", tc.name)
		})
	}
}
