package character

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
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

// TestGetAvailableGifts tests gift availability filtering
func TestGetAvailableGifts(t *testing.T) {
	character := &CharacterCard{
		Name: "Test Character",
		Personality: &PersonalityConfig{
			Traits: map[string]float64{
				"shy": 0.7,
			},
		},
	}
	
	gameState := &GameState{
		Stats: map[string]*Stat{
			"affection": {Current: 60, Max: 100},
			"happiness": {Current: 50, Max: 100},
		},
		RelationshipLevel: "Friend",
	}
	
	gm := NewGiftManager(character, gameState)

	// Create test gifts with different requirements
	gm.giftCatalog = map[string]*GiftDefinition{
		"available_gift": {
			ID:   "available_gift",
			Name: "Available Gift",
			Properties: GiftProperties{
				UnlockRequirements: map[string]interface{}{
					"stats": map[string]interface{}{
						"affection": map[string]interface{}{"min": 50.0},
					},
				},
			},
		},
		"locked_gift": {
			ID:   "locked_gift",
			Name: "Locked Gift",
			Properties: GiftProperties{
				UnlockRequirements: map[string]interface{}{
					"stats": map[string]interface{}{
						"affection": map[string]interface{}{"min": 80.0},
					},
				},
			},
		},
		"relationship_locked": {
			ID:   "relationship_locked",
			Name: "Relationship Locked Gift",
			Properties: GiftProperties{
				UnlockRequirements: map[string]interface{}{
					"relationshipLevel": "Partner",
				},
			},
		},
	}

	available := gm.GetAvailableGifts()
	
	// Should only return available_gift
	if len(available) != 1 {
		t.Errorf("Expected 1 available gift, got %d", len(available))
	}
	
	if available[0].ID != "available_gift" {
		t.Errorf("Expected available_gift to be available, got %s", available[0].ID)
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
				Tags:          []string{"test_gift"],
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

// TestGiveGiftRequirementsNotMet tests giving a gift when requirements aren't met
func TestGiveGiftRequirementsNotMet(t *testing.T) {
	character := &CharacterCard{Name: "Test Character"}
	gameState := &GameState{
		Stats: map[string]*Stat{
			"affection": {Current: 20, Max: 100},
		},
	}
	gm := NewGiftManager(character, gameState)

	// Create gift with high requirements
	testGift := &GiftDefinition{
		ID:   "locked_gift",
		Name: "Locked Gift",
		Properties: GiftProperties{
			UnlockRequirements: map[string]interface{}{
				"stats": map[string]interface{}{
					"affection": map[string]interface{}{"min": 80.0},
				},
			},
		},
		GiftEffects: GiftEffects{
			Immediate: ImmediateEffects{
				Stats:     map[string]float64{"happiness": 10},
				Responses: []string{"Thank you!"},
			},
		},
	}
	
	gm.giftCatalog["locked_gift"] = testGift

	response, err := gm.GiveGift("locked_gift", "")
	
	if err == nil {
		t.Errorf("Expected error when gift requirements not met")
	}
	
	if response.ErrorMessage == "" {
		t.Errorf("Expected error message to be set")
	}
}

// TestPersonalityModifiers tests that personality modifiers are applied correctly
func TestPersonalityModifiers(t *testing.T) {
	character := &CharacterCard{
		Name: "Test Character",
		Personality: &PersonalityConfig{
			Traits: map[string]float64{
				"romantic": 0.9,
			},
		},
	}
	
	gameState := &GameState{
		Stats: map[string]*Stat{
			"affection": {Current: 50, Max: 100},
		},
	}
	
	gm := NewGiftManager(character, gameState)

	testGift := &GiftDefinition{
		ID:   "romantic_gift",
		Name: "Romantic Gift",
		Properties: GiftProperties{
			UnlockRequirements: map[string]interface{}{},
		},
		GiftEffects: GiftEffects{
			Immediate: ImmediateEffects{
				Stats:     map[string]float64{"affection": 10},
				Responses: []string{"So romantic!"},
			},
		},
		PersonalityModifiers: map[string]map[string]float64{
			"romantic": {"affection": 2.0}, // Double effect for romantic characters
		},
	}
	
	gm.giftCatalog["romantic_gift"] = testGift

	initialAffection := gameState.Stats["affection"].Current
	
	_, err := gm.GiveGift("romantic_gift", "")
	if err != nil {
		t.Fatalf("Failed to give gift: %v", err)
	}
	
	finalAffection := gameState.Stats["affection"].Current
	increase := finalAffection - initialAffection
	
	// Should be close to 19 (10 * (1 + (2-1) * 0.9))
	expectedIncrease := 10.0 * (1.0 + (2.0-1.0)*0.9)
	if increase < expectedIncrease-1 || increase > expectedIncrease+1 {
		t.Errorf("Expected affection increase around %f, got %f", expectedIncrease, increase)
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

// TestMemoryLimiting tests that gift memories are limited to prevent unbounded growth
func TestMemoryLimiting(t *testing.T) {
	character := &CharacterCard{Name: "Test Character"}
	gameState := &GameState{
		Stats: map[string]*Stat{
			"happiness": {Current: 50, Max: 100},
		},
	}
	gm := NewGiftManager(character, gameState)

	testGift := &GiftDefinition{
		ID:   "memory_test_gift",
		Name: "Memory Test Gift",
		Properties: GiftProperties{
			UnlockRequirements: map[string]interface{}{},
		},
		GiftEffects: GiftEffects{
			Immediate: ImmediateEffects{
				Stats:     map[string]float64{"happiness": 1},
				Responses: []string{"Thanks!"},
			},
			Memory: MemoryEffects{
				Importance: 0.5,
				Tags:       []string{"test"},
			},
		},
	}
	
	gm.giftCatalog["memory_test_gift"] = testGift

	// Give gift many times to test memory limiting
	for i := 0; i < 150; i++ {
		_, err := gm.GiveGift("memory_test_gift", fmt.Sprintf("Gift %d", i))
		if err != nil {
			t.Fatalf("Failed to give gift %d: %v", i, err)
		}
	}

	memories := gm.GetGiftMemories()
	if len(memories) > 100 {
		t.Errorf("Expected memory count to be limited to 100, got %d", len(memories))
	}
	
	// Verify most recent memories are kept
	if len(memories) == 100 {
		lastMemory := memories[len(memories)-1]
		if lastMemory.Notes != "Gift 149" {
			t.Errorf("Expected last memory to be most recent, got notes: %s", lastMemory.Notes)
		}
	}
}

// TestConcurrentAccess tests thread safety of GiftManager
func TestConcurrentAccess(t *testing.T) {
	character := &CharacterCard{Name: "Test Character"}
	gameState := &GameState{
		Stats: map[string]*Stat{
			"happiness": {Current: 50, Max: 100},
		},
	}
	gm := NewGiftManager(character, gameState)

	testGift := &GiftDefinition{
		ID:   "concurrent_test_gift",
		Name: "Concurrent Test Gift",
		Properties: GiftProperties{
			UnlockRequirements: map[string]interface{}{},
		},
		GiftEffects: GiftEffects{
			Immediate: ImmediateEffects{
				Stats:     map[string]float64{"happiness": 1},
				Responses: []string{"Thanks!"},
			},
		},
	}
	
	gm.giftCatalog["concurrent_test_gift"] = testGift

	// Test concurrent access
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()
			
			// Perform multiple operations concurrently
			gm.GetAvailableGifts()
			gm.GetGiftCatalog()
			gm.IsGiftSystemEnabled()
			
			if id%2 == 0 {
				gm.GiveGift("concurrent_test_gift", fmt.Sprintf("Concurrent gift %d", id))
			}
			
			gm.GetGiftMemories()
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// If we reach here without deadlock or panic, test passes
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
