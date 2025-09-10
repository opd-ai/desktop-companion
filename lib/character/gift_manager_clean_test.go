package character

import (
	"os"
	"path/filepath"
	"testing"
)

// TestGiftManagerBasic tests basic gift manager functionality
func TestGiftManagerBasic(t *testing.T) {
	character := &CharacterCard{
		Name: "Test Character",
		GiftSystem: &GiftSystemConfig{
			Enabled: true,
		},
	}
	gameState := &GameState{
		Stats: map[string]*Stat{
			"happiness": {Current: 50, Max: 100},
		},
	}

	gm := NewGiftManager(character, gameState)

	// Test creation
	if gm.character != character {
		t.Errorf("Expected character to be set correctly")
	}
	if gm.gameState != gameState {
		t.Errorf("Expected gameState to be set correctly")
	}

	// Test gift system status
	if !gm.IsGiftSystemEnabled() {
		t.Errorf("Expected gift system to be enabled")
	}

	// Test empty catalog initially
	catalog := gm.GetGiftCatalog()
	if len(catalog) != 0 {
		t.Errorf("Expected empty catalog initially, got %d items", len(catalog))
	}

	available := gm.GetAvailableGifts()
	if len(available) != 0 {
		t.Errorf("Expected no available gifts initially, got %d", len(available))
	}
}

// TestGiftManagerWithCatalog tests gift manager with loaded catalog
func TestGiftManagerWithCatalog(t *testing.T) {
	// Create temporary directory with test gift
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

	err := os.WriteFile(giftFile, []byte(giftContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to create test gift file: %v", err)
	}

	character := &CharacterCard{Name: "Test Character"}
	gameState := &GameState{
		Stats: map[string]*Stat{
			"happiness": {Current: 50, Max: 100},
		},
	}
	gm := NewGiftManager(character, gameState)

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

	// Test giving gift
	response, err := gm.GiveGift("test_gift", "Test note")
	if err != nil {
		t.Fatalf("Failed to give gift: %v", err)
	}

	if response.Response == "" {
		t.Errorf("Expected response to be set")
	}

	if !response.MemoryCreated {
		t.Errorf("Expected memory to be created")
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
	if memory.Notes != "Test note" {
		t.Errorf("Expected notes to be preserved, got '%s'", memory.Notes)
	}
}

// TestGiftGivingErrors tests error conditions in gift giving
func TestGiftGivingErrors(t *testing.T) {
	character := &CharacterCard{Name: "Test Character"}
	gameState := &GameState{Stats: make(map[string]*Stat)}
	gm := NewGiftManager(character, gameState)

	// Test giving non-existent gift
	response, err := gm.GiveGift("nonexistent", "")
	if err == nil {
		t.Errorf("Expected error when giving non-existent gift")
	}
	if response.ErrorMessage == "" {
		t.Errorf("Expected error message to be set")
	}
}

// BenchmarkGiftManagerOperations benchmarks gift manager performance
func BenchmarkGiftManagerOperations(b *testing.B) {
	character := &CharacterCard{
		Name:       "Benchmark Character",
		GiftSystem: &GiftSystemConfig{Enabled: true},
	}
	gameState := &GameState{
		Stats: map[string]*Stat{
			"happiness": {Current: 50, Max: 100},
		},
	}
	gm := NewGiftManager(character, gameState)

	// Add test gift to catalog
	testGift := &GiftDefinition{
		ID:   "benchmark_gift",
		Name: "Benchmark Gift",
		Properties: GiftProperties{
			UnlockRequirements: map[string]interface{}{},
		},
		GiftEffects: GiftEffects{
			Immediate: ImmediateEffects{
				Stats:     map[string]float64{"happiness": 10},
				Responses: []string{"Thank you!"},
			},
			Memory: MemoryEffects{
				Importance: 0.5,
			},
		},
	}
	gm.giftCatalog["benchmark_gift"] = testGift

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gm.GetAvailableGifts()
	}
}
