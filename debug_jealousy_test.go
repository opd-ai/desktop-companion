package main

import (
	"fmt"
	"testing"

	"desktop-companion/internal/character"
)

func TestJealousyDebug(t *testing.T) {
	// Create minimal character card with jealousy traits
	card := &character.CharacterCard{
		Name:        "Debug Character",
		Description: "Debug character for jealousy testing",
		Animations: map[string]string{
			"idle":    "test.gif",
			"talking": "test.gif",
		},
		Dialogs: []character.Dialog{
			{
				Trigger:   "click",
				Responses: []string{"Hello!"},
				Animation: "talking",
				Cooldown:  5,
			},
		},
		Behavior: character.Behavior{
			IdleTimeout:     30,
			MovementEnabled: true,
			DefaultSize:     128,
		},
		Stats: map[string]character.StatConfig{
			"affection": {Initial: 50, Max: 100, DegradationRate: 0.1, CriticalThreshold: 10},
			"trust":     {Initial: 50, Max: 100, DegradationRate: 0.05, CriticalThreshold: 5},
			"happiness": {Initial: 50, Max: 100, DegradationRate: 0.8, CriticalThreshold: 15},
			"jealousy":  {Initial: 0, Max: 100, DegradationRate: 2.0, CriticalThreshold: 80},
		},
		GameRules: &character.GameRulesConfig{
			StatsDecayInterval: 60,
			AutoSaveInterval:   300,
		},
		Personality: &character.PersonalityConfig{
			Traits: map[string]float64{
				"jealousy_prone": 0.6,
			},
		},
	}

	// Create character
	char := &character.Character{}
	err := char.EnableGameMode(nil, "")
	if err != nil {
		t.Fatalf("Failed to enable game mode: %v", err)
	}

	gameState := char.GetGameState()
	if gameState == nil {
		t.Fatal("Game state should not be nil")
	}

	fmt.Printf("Initial stats: %+v\n", gameState.GetStats())

	// Set jealousy high
	gameState.ApplyInteractionEffects(map[string]float64{
		"jealousy": 85.0,
	})

	fmt.Printf("After setting jealousy: %+v\n", gameState.GetStats())

	// Force update
	char.Update()

	fmt.Printf("After update: %+v\n", gameState.GetStats())
}
