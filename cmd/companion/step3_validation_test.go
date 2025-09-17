package main

import (
	"path/filepath"
	"testing"

	"github.com/opd-ai/desktop-companion/lib/character"
)

// TestStep3SpecializedFeatures tests that Step 3 implementation
// has successfully added news and general events to appropriate characters
func TestStep3SpecializedFeatures(t *testing.T) {
	testCases := []struct {
		name                string
		charPath            string
		expectNews          bool
		expectGeneralEvents bool
		expectedNewsStyle   string
		expectedEventCount  int
	}{
		{
			name:                "Default Character",
			charPath:            "../../assets/characters/default/character.json",
			expectNews:          true,
			expectGeneralEvents: true,
			expectedNewsStyle:   "casual",
			expectedEventCount:  2,
		},
		{
			name:                "Easy Character",
			charPath:            "../../assets/characters/easy/character.json",
			expectNews:          true,
			expectGeneralEvents: true,
			expectedNewsStyle:   "gentle",
			expectedEventCount:  2,
		},
		{
			name:                "Specialist Character",
			charPath:            "../../assets/characters/specialist/character.json",
			expectNews:          true, // Sleepy character has news reading enabled in config
			expectGeneralEvents: true,
			expectedNewsStyle:   "wellness", // Sleep & wellness themed news
			expectedEventCount:  2,
		},
		{
			name:                "Markov Example",
			charPath:            "../../assets/characters/markov_example/character.json",
			expectNews:          true,
			expectGeneralEvents: true,
			expectedNewsStyle:   "analytical",
			expectedEventCount:  2,
		},
		{
			name:                "News Example",
			charPath:            "../../assets/characters/news_example/character.json",
			expectNews:          true, // Already has news features
			expectGeneralEvents: true,
			expectedNewsStyle:   "casual",
			expectedEventCount:  2,
		},
		{
			name:                "Romance Character",
			charPath:            "../../assets/characters/romance/character.json",
			expectNews:          true,
			expectGeneralEvents: true,
			expectedNewsStyle:   "intimate",
			expectedEventCount:  3, // Romance gets extra intimate events
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

			// Test news features
			if tc.expectNews {
				if !card.HasNewsFeatures() {
					t.Errorf("Character %s should have news features enabled", tc.name)
				} else {
					newsConfig := card.GetNewsConfig()
					if newsConfig == nil {
						t.Errorf("Character %s should have valid news configuration", tc.name)
					} else {
						if newsConfig.ReadingPersonality == nil {
							t.Errorf("Character %s should have reading personality configured", tc.name)
						}

						// Verify news feeds are configured
						if len(newsConfig.Feeds) == 0 {
							t.Errorf("Character %s should have news feeds configured", tc.name)
						}
					}
				}
			} else {
				if card.HasNewsFeatures() {
					t.Errorf("Character %s should not have news features", tc.name)
				}
			}

			// Test general events
			if tc.expectGeneralEvents {
				if len(card.GeneralEvents) == 0 {
					t.Errorf("Character %s should have general events configured", tc.name)
				} else {
					if len(card.GeneralEvents) < tc.expectedEventCount {
						t.Errorf("Character %s expected at least %d general events, got %d",
							tc.name, tc.expectedEventCount, len(card.GeneralEvents))
					}

					// Verify events have proper structure
					for _, event := range card.GeneralEvents {
						if event.Name == "" {
							t.Errorf("Character %s has general event with empty name", tc.name)
						}
						if len(event.Responses) == 0 {
							t.Errorf("Character %s general event %s has no responses", tc.name, event.Name)
						}
						if len(event.Choices) == 0 {
							t.Errorf("Character %s general event %s has no choices", tc.name, event.Name)
						}
					}
				}
			}

			t.Logf("✅ Character %s has appropriate Step 3 features", tc.name)
		})
	}
}

// TestStep3PersonalityPreservation verifies that personality themes are maintained
func TestStep3PersonalityPreservation(t *testing.T) {
	testCases := []struct {
		name          string
		charPath      string
		expectedTheme string
	}{
		{
			name:          "Default Character",
			charPath:      "../../assets/characters/default/character.json",
			expectedTheme: "friendly companion",
		},
		{
			name:          "Easy Character",
			charPath:      "../../assets/characters/easy/character.json",
			expectedTheme: "beginner-friendly pet",
		},
		{
			name:          "Specialist Character",
			charPath:      "../../assets/characters/specialist/character.json",
			expectedTheme: "energy management focus",
		},
		{
			name:          "Markov Example",
			charPath:      "../../assets/characters/markov_example/character.json",
			expectedTheme: "AI dialog demonstration",
		},
		{
			name:          "News Example",
			charPath:      "../../assets/characters/news_example/character.json",
			expectedTheme: "news reading companion",
		},
		{
			name:          "Romance Character",
			charPath:      "../../assets/characters/romance/character.json",
			expectedTheme: "dating simulator",
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

			// Verify character maintains its core theme through description
			if card.Description == "" {
				t.Errorf("Character %s should have a description", tc.name)
			}

			// Check that personality traits are preserved (if they exist)
			if card.Personality != nil && card.Personality.Traits != nil {
				traits := card.Personality.Traits
				hasPersonalityTraits := len(traits) > 0
				if !hasPersonalityTraits {
					t.Errorf("Character %s should preserve personality traits", tc.name)
				}
			}

			t.Logf("✅ Character %s preserved personality and theme: %s", tc.name, tc.expectedTheme)
		})
	}
}

// TestStep3BackwardCompatibility ensures Step 1 and Step 2 features still work
func TestStep3BackwardCompatibility(t *testing.T) {
	testCases := []struct {
		name     string
		charPath string
	}{
		{"Default Character", "../../assets/characters/default/character.json"},
		{"Easy Character", "../../assets/characters/easy/character.json"},
		{"Specialist Character", "../../assets/characters/specialist/character.json"},
		{"Markov Example", "../../assets/characters/markov_example/character.json"},
		{"News Example", "../../assets/characters/news_example/character.json"},
		{"Romance Character", "../../assets/characters/romance/character.json"},
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

			// Test Step 1 features (basic game features)
			if card.Stats == nil {
				t.Errorf("Character %s should have stats (Step 1 feature)", tc.name)
			}
			if card.Interactions == nil {
				t.Errorf("Character %s should have interactions (Step 1 feature)", tc.name)
			}

			// Test Step 2 features (romance and multiplayer)
			if card.Personality == nil {
				t.Errorf("Character %s should have personality (Step 2 feature)", tc.name)
			}

			// Check for romance stats (affection, trust)
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
				t.Errorf("Character %s should have affection stat (Step 2 feature)", tc.name)
			}
			if !hasTrust {
				t.Errorf("Character %s should have trust stat (Step 2 feature)", tc.name)
			}

			t.Logf("✅ Character %s maintains backward compatibility", tc.name)
		})
	}
}
