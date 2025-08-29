package character

import (
	"os"
	"path/filepath"
	"testing"
)

// TestNewGiftManager tests creating a new gift manager
func TestNewGiftManager(t *testing.T) {
	character := &CharacterCard{
		Name: "Test Character",
	}
	gameState := &GameState{
		Stats: make(map[string]*Stat),
	}

	gm := NewGiftManager(character, gameState)

	if gm.character != character {
		t.Errorf("Expected character to be set correctly")
	}
	if gm.gameState != gameState {
		t.Errorf("Expected gameState to be set correctly")
	}
	if gm.giftCatalog == nil {
		t.Errorf("Expected giftCatalog to be initialized")
	}
}

// TestLoadGiftCatalogIntegration tests loading gift catalog through GiftManager
func TestLoadGiftCatalogIntegration(t *testing.T) {
	character := &CharacterCard{Name: "Test Character"}
	gameState := &GameState{Stats: make(map[string]*Stat)}
	gm := NewGiftManager(character, gameState)

	// Create temporary directory with test gifts
	tmpDir := t.TempDir()
	giftFile := filepath.Join(tmpDir, "test_gift.json")
	giftContent := `{
		"id": "test_gift",
		"name": "Test Gift",
		"description": "A test gift",
		"category": "food",
		"rarity": "common",
		"image": "test.gif",
		"properties": {
			"consumable": true,
			"stackable": false,
			"maxStack": 1,
			"unlockRequirements": {}
		},
		"giftEffects": {
			"immediate": {
				"stats": {"happiness": 10},
				"animations": ["happy"],
				"responses": ["Thank you!"]
			},
			"memory": {
				"importance": 0.5,
				"tags": ["test"],
				"emotionalTone": "happy"
			}
		},
		"personalityModifiers": {},
		"notes": {
			"enabled": true,
			"maxLength": 100,
			"placeholder": "Add a note..."
		}
	}`

	err := os.WriteFile(giftFile, []byte(giftContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test gift file: %v", err)
	}

	// Test loading catalog
	err = gm.LoadGiftCatalog(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load gift catalog: %v", err)
	}

	// Verify catalog contents
	catalog := gm.GetGiftCatalog()
	if len(catalog) != 1 {
		t.Errorf("Expected 1 gift in catalog, got %d", len(catalog))
	}

	gift, exists := catalog["test_gift"]
	if !exists {
		t.Errorf("Expected test_gift to be in catalog")
	}
	if gift.Name != "Test Gift" {
		t.Errorf("Expected gift name to be 'Test Gift', got '%s'", gift.Name)
	}
}

// TestGiveGift tests the complete gift giving process
func TestGiveGift(t *testing.T) {
	character := &CharacterCard{
		Name: "Test Character",
		Personality: &PersonalityConfig{
			Traits: map[string]float64{
				"shy": 0.8,
			},
		},
		GiftSystem: &GiftSystemConfig{
			Enabled: true,
			Preferences: GiftPreferences{
				PersonalityResponses: map[string]PersonalityResponse{
					"shy": {
						GiftReceived: []string{"Oh... thank you..."},
						Animations:   []string{"blushing"},
					},
				},
			},
		},
	}

	gameState := &GameState{
		Stats: map[string]*Stat{
			"happiness": {Current: 50, Max: 100},
			"affection": {Current: 30, Max: 100},
		},
		RelationshipLevel: "Friend",
	}

	gm := NewGiftManager(character, gameState)

	// Create test gift
	testGift := &GiftDefinition{
		ID:   "test_gift",
		Name: "Test Gift",
		Properties: GiftProperties{
			UnlockRequirements: map[string]interface{}{},
		},
		GiftEffects: GiftEffects{
			Immediate: ImmediateEffects{
				Stats:      map[string]float64{"happiness": 20, "affection": 10},
				Responses:  []string{"Thank you so much!"},
				Animations: []string{"happy"},
			},
			Memory: MemoryEffects{
				Importance:    0.7,
				Tags:          []string{"test_gift"},
				EmotionalTone: "grateful",
			},
		},
		PersonalityModifiers: map[string]map[string]float64{
			"shy": {"affection": 1.5},
		},
	}

	gm.giftCatalog["test_gift"] = testGift

	// Test giving gift
	response, err := gm.GiveGift("test_gift", "Hope you like it!")

	if err != nil {
		t.Fatalf("Failed to give gift: %v", err)
	}

	if response.Response == "" {
		t.Errorf("Expected response to be set")
	}

	if response.Animation == "" {
		t.Errorf("Expected animation to be set")
	}

	if !response.MemoryCreated {
		t.Errorf("Expected memory to be created")
	}

	// Verify stat effects were applied
	happinessStat := gameState.Stats["happiness"]
	if happinessStat.Current <= 50 {
		t.Errorf("Expected happiness to increase, current: %f", happinessStat.Current)
	}

	affectionStat := gameState.Stats["affection"]
	if affectionStat.Current <= 30 {
		t.Errorf("Expected affection to increase, current: %f", affectionStat.Current)
	}

	// Verify memory was created
	memories := gm.GetGiftMemories()
	if len(memories) != 1 {
		t.Errorf("Expected 1 gift memory, got %d", len(memories))
	}

	memory := memories[0]
	if memory.GiftID != "test_gift" {
		t.Errorf("Expected gift ID to be 'test_gift', got '%s'", memory.GiftID)
	}
	if memory.Notes != "Hope you like it!" {
		t.Errorf("Expected notes to be preserved, got '%s'", memory.Notes)
	}
}

// TestGiveGiftNotFound tests giving a non-existent gift
func TestGiveGiftNotFound(t *testing.T) {
	character := &CharacterCard{Name: "Test Character"}
	gameState := &GameState{Stats: make(map[string]*Stat)}
	gm := NewGiftManager(character, gameState)

	response, err := gm.GiveGift("nonexistent", "")

	if err == nil {
		t.Errorf("Expected error when giving non-existent gift")
	}

	if response.ErrorMessage == "" {
		t.Errorf("Expected error message to be set")
	}
}

// TestIsGiftSystemEnabled tests gift system availability checking
func TestIsGiftSystemEnabled(t *testing.T) {
	tests := []struct {
		name      string
		character *CharacterCard
		expected  bool
	}{
		{
			name: "gift_system_enabled",
			character: &CharacterCard{
				GiftSystem: &GiftSystemConfig{Enabled: true},
			},
			expected: true,
		},
		{
			name: "gift_system_disabled",
			character: &CharacterCard{
				GiftSystem: &GiftSystemConfig{Enabled: false},
			},
			expected: false,
		},
		{
			name:      "gift_system_nil",
			character: &CharacterCard{},
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gameState := &GameState{Stats: make(map[string]*Stat)}
			gm := NewGiftManager(tt.character, gameState)

			result := gm.IsGiftSystemEnabled()
			if result != tt.expected {
				t.Errorf("Expected IsGiftSystemEnabled() to return %v, got %v", tt.expected, result)
			}
		})
	}
}

// BenchmarkGiveGift benchmarks gift giving performance
func BenchmarkGiveGift(b *testing.B) {
	character := &CharacterCard{
		Name: "Benchmark Character",
		Personality: &PersonalityConfig{
			Traits: map[string]float64{"romantic": 0.5},
		},
	}

	gameState := &GameState{
		Stats: map[string]*Stat{
			"happiness": {Current: 50, Max: 100},
			"affection": {Current: 30, Max: 100},
		},
	}

	gm := NewGiftManager(character, gameState)

	testGift := &GiftDefinition{
		ID:   "benchmark_gift",
		Name: "Benchmark Gift",
		Properties: GiftProperties{
			UnlockRequirements: map[string]interface{}{},
		},
		GiftEffects: GiftEffects{
			Immediate: ImmediateEffects{
				Stats:      map[string]float64{"happiness": 10, "affection": 5},
				Responses:  []string{"Thank you!", "Amazing!", "I love it!"},
				Animations: []string{"happy", "excited"},
			},
			Memory: MemoryEffects{
				Importance:    0.5,
				Tags:          []string{"benchmark", "test"},
				EmotionalTone: "happy",
			},
		},
		PersonalityModifiers: map[string]map[string]float64{
			"romantic": {"affection": 1.2},
		},
	}

	gm.giftCatalog["benchmark_gift"] = testGift

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gm.GiveGift("benchmark_gift", "Benchmark note")
		if err != nil {
			b.Fatalf("Failed to give gift: %v", err)
		}
	}
}
