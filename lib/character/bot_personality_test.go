package character

import (
	"github.com/opd-ai/desktop-companion/internal/bot"
	"testing"
)

// testCharacterBase returns a basic valid character card structure
func testCharacterBase(name, description string) CharacterCard {
	return CharacterCard{
		Name:        name,
		Description: description,
		Animations: map[string]string{
			"idle":    "idle.gif",
			"talking": "talking.gif",
		},
		Dialogs:  []Dialog{{Trigger: "click", Responses: []string{"Hello"}, Animation: "idle"}},
		Behavior: Behavior{IdleTimeout: 30},
	}
}

func TestCharacterCard_BotPersonalityIntegration(t *testing.T) {
	tests := []struct {
		name              string
		setupCard         func() CharacterCard
		expectBot         bool
		expectPersonality bool
	}{
		{
			name: "Valid character with bot personality",
			setupCard: func() CharacterCard {
				card := testCharacterBase("Test Bot", "Test bot character")
				card.Multiplayer = &MultiplayerConfig{
					Enabled:    true,
					BotCapable: true,
					NetworkID:  "test_bot",
					BotPersonality: &bot.PersonalityArchetype{
						Name:        "Test",
						Description: "Test personality",
						Traits: map[string]float64{
							"chattiness":  0.8,
							"helpfulness": 0.7,
						},
						Behavior: bot.PersonalityBehavior{
							ResponseDelay:       "2s-4s",
							InteractionRate:     3.0,
							Attention:           0.8,
							MaxActionsPerMinute: 6,
							MinTimeBetweenSame:  10,
							PreferredActions:    []string{"chat", "click"},
						},
					},
				}
				return card
			},
			expectBot:         true,
			expectPersonality: true,
		},
		{
			name: "Bot capable but no personality defined",
			setupCard: func() CharacterCard {
				card := testCharacterBase("Test Bot", "Test bot character")
				card.Multiplayer = &MultiplayerConfig{
					Enabled:        true,
					BotCapable:     true,
					NetworkID:      "test_bot",
					BotPersonality: nil, // No personality defined
				}
				return card
			},
			expectBot:         true,
			expectPersonality: false,
		},
		{
			name: "Character not bot capable",
			setupCard: func() CharacterCard {
				card := testCharacterBase("Test Character", "Test non-bot character")
				card.Multiplayer = &MultiplayerConfig{
					Enabled:    true,
					BotCapable: false,
					NetworkID:  "test_char",
				}
				return card
			},
			expectBot:         false,
			expectPersonality: false,
		},
		{
			name: "No multiplayer configuration",
			setupCard: func() CharacterCard {
				card := testCharacterBase("Test Character", "Test character")
				card.Multiplayer = nil
				return card
			},
			expectBot:         false,
			expectPersonality: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := tt.setupCard()

			// Test IsBotCapable method
			isBotCapable := card.IsBotCapable()
			if isBotCapable != tt.expectBot {
				t.Errorf("IsBotCapable() = %v, expected %v", isBotCapable, tt.expectBot)
			}

			// Test GetBotPersonality method
			personality, err := card.GetBotPersonality()
			if err != nil {
				t.Errorf("GetBotPersonality() error: %v", err)
				return
			}

			if tt.expectPersonality {
				if personality == nil {
					t.Error("Expected personality but got nil")
					return
				}

				// Validate personality structure
				if personality.SocialTendencies == nil {
					t.Error("SocialTendencies map is nil")
				}

				if personality.EmotionalProfile == nil {
					t.Error("EmotionalProfile map is nil")
				}

				// Check some trait values were transferred
				if len(personality.SocialTendencies) == 0 && len(personality.EmotionalProfile) == 0 {
					t.Error("No personality traits found")
				}
			} else {
				if personality != nil {
					t.Error("Expected nil personality but got one")
				}
			}
		})
	}
}

func TestCharacterCard_BotPersonalityBuiltinArchetypes(t *testing.T) {
	manager := bot.NewPersonalityManager()
	archetypeNames := manager.ListArchetypes()

	for _, name := range archetypeNames {
		t.Run("archetype_"+name, func(t *testing.T) {
			// Get the archetype
			archetype, err := manager.GetArchetype(name)
			if err != nil {
				t.Fatalf("Failed to get archetype %s: %v", name, err)
			}

			// Create character card with this archetype
			card := testCharacterBase("Test "+archetype.Name, "Test character with "+name+" archetype")
			card.Multiplayer = &MultiplayerConfig{
				Enabled:        true,
				BotCapable:     true,
				NetworkID:      "test_" + name,
				BotPersonality: archetype,
			}

			// Test that the personality can be retrieved correctly
			personality, err := card.GetBotPersonality()
			if err != nil {
				t.Errorf("Failed to get bot personality for %s archetype: %v", name, err)
				return
			}

			if personality == nil {
				t.Errorf("Got nil personality for %s archetype", name)
				return
			}

			// Verify that personality has some expected structure
			if len(personality.SocialTendencies) == 0 && len(personality.EmotionalProfile) == 0 {
				t.Errorf("No personality traits found for %s archetype", name)
			}

			// Verify the archetype traits are properly converted
			if len(archetype.Traits) > 0 {
				totalTraits := len(personality.SocialTendencies) + len(personality.EmotionalProfile)
				if totalTraits == 0 {
					t.Errorf("Archetype %s has traits but personality conversion resulted in no traits", name)
				}
			}
		})
	}
}
