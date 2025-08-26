package character

import (
	"testing"
)

// TestHoverDialogDebug specifically debugs the hover dialog issue
func TestHoverDialogDebug(t *testing.T) {
	card := &CharacterCard{
		Name:        "Hover Debug Character",
		Description: "Debug character for hover dialogue",
		Animations: map[string]string{
			"idle":     "idle.gif",
			"blushing": "blushing.gif",
		},
		Dialogs: []Dialog{
			{Trigger: "hover", Responses: []string{"Regular hover"}, Animation: "idle", Cooldown: 5},
		},
		Behavior: Behavior{IdleTimeout: 30, DefaultSize: 128},
		Stats: map[string]StatConfig{
			"affection": {Initial: 0, Max: 100, DegradationRate: 0.1, CriticalThreshold: 10},
			"trust":     {Initial: 0, Max: 100, DegradationRate: 0.1, CriticalThreshold: 10},
		},
		Personality: &PersonalityConfig{
			Traits: map[string]float64{"romanticism": 0.8},
		},
		RomanceDialogs: []DialogExtended{
			{
				Dialog: Dialog{
					Trigger:   "hover",
					Responses: []string{"*heart flutters* 💓", "Just being near you makes me happy..."},
					Animation: "blushing",
					Cooldown:  10,
				},
				Requirements: &RomanceRequirement{
					Stats: map[string]map[string]float64{
						"affection": {"min": 30},
						"trust":     {"min": 20},
					},
				},
			},
		},
	}

	char := createTestCharacterInstance(card, true)

	// Set the stats to meet requirements
	char.gameState.Stats["affection"].Current = 35
	char.gameState.Stats["trust"].Current = 25

	t.Logf("Affection: %f (requirement: >= 30)", char.gameState.Stats["affection"].Current)
	t.Logf("Trust: %f (requirement: >= 20)", char.gameState.Stats["trust"].Current)
	t.Logf("HasRomanceFeatures: %v", char.card.HasRomanceFeatures())
	t.Logf("RomanceDialogs count: %d", len(char.card.RomanceDialogs))

	// Test selectRomanceDialog directly
	response := char.selectRomanceDialog("hover")
	t.Logf("selectRomanceDialog(hover) response: '%s'", response)

	// Test HandleHover
	response = char.HandleHover()
	t.Logf("HandleHover response: '%s'", response)

	// Let me check if requirements are satisfied
	if char.card.RomanceDialogs != nil && len(char.card.RomanceDialogs) > 0 {
		for i, dialog := range char.card.RomanceDialogs {
			if dialog.Trigger == "hover" {
				canSatisfy := char.canSatisfyRomanceRequirements(dialog.Requirements)
				t.Logf("Romance dialog %d (hover) requirements satisfied: %v", i, canSatisfy)
				
				if dialog.Requirements != nil && dialog.Requirements.Stats != nil {
					for statName, conditions := range dialog.Requirements.Stats {
						currentStat := char.gameState.GetStat(statName)
						t.Logf("  %s: current=%f, conditions=%v", statName, currentStat, conditions)
						
						// Test the requirements check directly
						canSatisfyThis := char.gameState.CanSatisfyRequirements(map[string]map[string]float64{statName: conditions})
						t.Logf("  %s requirement satisfied: %v", statName, canSatisfyThis)
					}
				}
			}
		}
	}
}
