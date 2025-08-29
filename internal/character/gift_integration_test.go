package character

import (
	"testing"
)

// TestRealGiftCatalogLoading tests loading the actual gift files from assets/gifts/
func TestRealGiftCatalogLoading(t *testing.T) {
	character := &CharacterCard{
		Name: "Test Character",
		GiftSystem: &GiftSystemConfig{
			Enabled: true,
			Preferences: GiftPreferences{
				FavoriteCategories: []string{"food", "flowers"},
			},
		},
		Personality: &PersonalityConfig{
			Traits: map[string]float64{
				"romantic": 0.7,
				"shy":      0.3,
			},
		},
	}

	gameState := &GameState{
		Stats: map[string]*Stat{
			"happiness": {Current: 60, Max: 100},
			"affection": {Current: 40, Max: 100},
			"health":    {Current: 70, Max: 100},
			"trust":     {Current: 35, Max: 100},
			"intimacy":  {Current: 20, Max: 100},
		},
		RelationshipLevel: "Friend",
	}

	gm := NewGiftManager(character, gameState)

	// Test loading real gifts from assets/gifts/ directory
	err := gm.LoadGiftCatalog("/workspaces/DDS/assets/gifts")
	if err != nil {
		t.Fatalf("Failed to load real gift catalog: %v", err)
	}

	catalog := gm.GetGiftCatalog()

	// Should have loaded multiple gifts
	if len(catalog) == 0 {
		t.Errorf("Expected to load gifts from assets/gifts/, got empty catalog")
	}

	// Verify specific known gifts exist
	expectedGifts := []string{"birthday_cake", "chocolate_box", "red_rose", "red_roses", "coffee_book"}
	for _, expectedID := range expectedGifts {
		gift, exists := catalog[expectedID]
		if !exists {
			t.Errorf("Expected gift '%s' to be loaded", expectedID)
			continue
		}

		// Verify gift has required fields
		if gift.Name == "" {
			t.Errorf("Gift '%s' missing name", expectedID)
		}
		if gift.Category == "" {
			t.Errorf("Gift '%s' missing category", expectedID)
		}
		if len(gift.GiftEffects.Immediate.Responses) == 0 {
			t.Errorf("Gift '%s' missing responses", expectedID)
		}
	}

	// Test getting available gifts based on stats and relationship
	available := gm.GetAvailableGifts()

	// Should have some available gifts since we have decent stats
	if len(available) == 0 {
		t.Errorf("Expected some gifts to be available with current stats")
	}

	// Test giving a real gift
	if len(available) > 0 {
		testGift := available[0]
		response, err := gm.GiveGift(testGift.ID, "Testing real gift system!")

		if err != nil {
			t.Errorf("Failed to give available gift '%s': %v", testGift.ID, err)
		} else {
			if response.Response == "" {
				t.Errorf("Expected response when giving gift '%s'", testGift.ID)
			}
			if !response.MemoryCreated {
				t.Errorf("Expected memory to be created for gift '%s'", testGift.ID)
			}

			// Verify memory was actually created
			memories := gm.GetGiftMemories()
			if len(memories) != 1 {
				t.Errorf("Expected 1 gift memory after giving gift, got %d", len(memories))
			} else {
				memory := memories[0]
				if memory.GiftID != testGift.ID {
					t.Errorf("Expected memory for gift '%s', got '%s'", testGift.ID, memory.GiftID)
				}
				if memory.Notes != "Testing real gift system!" {
					t.Errorf("Expected notes to be preserved in memory")
				}
			}
		}
	}
}

// TestGiftSystemIntegrationWithPersonality tests personality-based gift responses
func TestGiftSystemIntegrationWithPersonality(t *testing.T) {
	// Test with shy character
	shyCharacter := &CharacterCard{
		Name: "Shy Character",
		GiftSystem: &GiftSystemConfig{
			Enabled: true,
			Preferences: GiftPreferences{
				PersonalityResponses: map[string]PersonalityResponse{
					"shy": {
						GiftReceived: []string{"Oh... thank you...", "You didn't have to..."},
						Animations:   []string{"blushing", "shy"},
					},
				},
			},
		},
		Personality: &PersonalityConfig{
			Traits: map[string]float64{
				"shy": 0.9, // Very shy
			},
		},
	}

	gameState := &GameState{
		Stats: map[string]*Stat{
			"happiness": {Current: 50, Max: 100},
			"affection": {Current: 30, Max: 100},
		},
	}

	gm := NewGiftManager(shyCharacter, gameState)

	// Create a test gift with personality modifiers
	testGift := &GiftDefinition{
		ID:   "personality_test_gift",
		Name: "Personality Test Gift",
		Properties: GiftProperties{
			UnlockRequirements: map[string]interface{}{},
		},
		GiftEffects: GiftEffects{
			Immediate: ImmediateEffects{
				Stats:      map[string]float64{"affection": 10},
				Responses:  []string{"Generic thank you"},
				Animations: []string{"happy"},
			},
		},
		PersonalityModifiers: map[string]map[string]float64{
			"shy": {"affection": 1.5}, // 50% bonus for shy characters
		},
	}

	gm.giftCatalog["personality_test_gift"] = testGift

	initialAffection := gameState.Stats["affection"].Current

	response, err := gm.GiveGift("personality_test_gift", "")
	if err != nil {
		t.Fatalf("Failed to give gift: %v", err)
	}

	// Should use personality-specific response
	expectedResponses := []string{"Oh... thank you...", "You didn't have to..."}
	responseFound := false
	for _, expected := range expectedResponses {
		if response.Response == expected {
			responseFound = true
			break
		}
	}
	if !responseFound {
		t.Errorf("Expected personality-specific response, got: %s", response.Response)
	}

	// Verify personality modifier was applied
	finalAffection := gameState.Stats["affection"].Current
	increase := finalAffection - initialAffection

	// Should be around 15 (10 * 1.5) due to shy personality modifier
	expectedIncrease := 10.0 * (1.0 + (1.5-1.0)*0.9) // 10 * 1.45 = 14.5
	if increase < expectedIncrease-1 || increase > expectedIncrease+1 {
		t.Errorf("Expected affection increase around %f, got %f", expectedIncrease, increase)
	}
}

// TestGiftRequirementFiltering tests that gifts are properly filtered by requirements
func TestGiftRequirementFiltering(t *testing.T) {
	character := &CharacterCard{Name: "Test Character"}

	// Low stats - should unlock fewer gifts
	gameState := &GameState{
		Stats: map[string]*Stat{
			"affection": {Current: 20, Max: 100},
		},
		RelationshipLevel: "Acquaintance",
	}

	gm := NewGiftManager(character, gameState)

	// Create gifts with different requirements
	lowRequirementGift := &GiftDefinition{
		ID:   "easy_gift",
		Name: "Easy Gift",
		Properties: GiftProperties{
			UnlockRequirements: map[string]interface{}{
				"stats": map[string]interface{}{
					"affection": map[string]interface{}{"min": 10.0},
				},
			},
		},
		GiftEffects: GiftEffects{
			Immediate: ImmediateEffects{
				Stats:     map[string]float64{"affection": 5},
				Responses: []string{"Thanks!"},
			},
		},
	}

	highRequirementGift := &GiftDefinition{
		ID:   "expensive_gift",
		Name: "Expensive Gift",
		Properties: GiftProperties{
			UnlockRequirements: map[string]interface{}{
				"stats": map[string]interface{}{
					"affection": map[string]interface{}{"min": 70.0},
				},
			},
		},
		GiftEffects: GiftEffects{
			Immediate: ImmediateEffects{
				Stats:     map[string]float64{"affection": 20},
				Responses: []string{"Amazing!"},
			},
		},
	}

	relationshipGift := &GiftDefinition{
		ID:   "romantic_gift",
		Name: "Romantic Gift",
		Properties: GiftProperties{
			UnlockRequirements: map[string]interface{}{
				"relationshipLevel": "Romantic Interest",
			},
		},
		GiftEffects: GiftEffects{
			Immediate: ImmediateEffects{
				Stats:     map[string]float64{"intimacy": 15},
				Responses: []string{"So romantic!"},
			},
		},
	}

	gm.giftCatalog["easy_gift"] = lowRequirementGift
	gm.giftCatalog["expensive_gift"] = highRequirementGift
	gm.giftCatalog["romantic_gift"] = relationshipGift

	available := gm.GetAvailableGifts()

	// Should only have the easy gift available
	if len(available) != 1 {
		t.Errorf("Expected 1 available gift, got %d", len(available))
	}

	if len(available) > 0 && available[0].ID != "easy_gift" {
		t.Errorf("Expected easy_gift to be available, got %s", available[0].ID)
	}

	// Increase stats and test again
	gameState.Stats["affection"].Current = 80
	gameState.RelationshipLevel = "Romantic Interest"

	available = gm.GetAvailableGifts()

	// Should now have all gifts available
	if len(available) != 3 {
		t.Errorf("Expected 3 available gifts with high stats, got %d", len(available))
	}
}
