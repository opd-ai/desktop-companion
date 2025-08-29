package character

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadGiftDefinition tests loading gift definitions from JSON files
func TestLoadGiftDefinition(t *testing.T) {
	tests := []struct {
		name          string
		jsonContent   string
		expectError   bool
		errorContains string
	}{
		{
			name: "valid_gift_definition",
			jsonContent: `{
				"id": "test_gift",
				"name": "Test Gift",
				"description": "A test gift for validation",
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
			}`,
			expectError: false,
		},
		{
			name: "missing_required_fields",
			jsonContent: `{
				"name": "Incomplete Gift"
			}`,
			expectError:   true,
			errorContains: "gift ID is required",
		},
		{
			name: "invalid_category",
			jsonContent: `{
				"id": "invalid_gift",
				"name": "Invalid Gift",
				"description": "A gift with invalid category",
				"category": "invalid_category",
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
			}`,
			expectError:   true,
			errorContains: "invalid gift category",
		},
		{
			name: "invalid_rarity",
			jsonContent: `{
				"id": "invalid_rarity_gift",
				"name": "Invalid Rarity Gift",
				"description": "A gift with invalid rarity",
				"category": "food",
				"rarity": "super_rare",
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
			}`,
			expectError:   true,
			errorContains: "invalid gift rarity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test_gift.json")

			err := os.WriteFile(tmpFile, []byte(tt.jsonContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Test loading
			gift, err := LoadGiftDefinition(tmpFile)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorContains != "" && !containsString(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorContains, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if gift == nil {
				t.Error("Expected gift definition, got nil")
				return
			}
		})
	}
}

// TestGiftDefinitionValidation tests gift validation logic
func TestGiftDefinitionValidation(t *testing.T) {
	tests := []struct {
		name          string
		gift          GiftDefinition
		expectError   bool
		errorContains string
	}{
		{
			name: "valid_gift",
			gift: GiftDefinition{
				ID:          "valid_gift",
				Name:        "Valid Gift",
				Description: "A valid gift",
				Category:    "food",
				Rarity:      "common",
				Properties: GiftProperties{
					Consumable: true,
					Stackable:  false,
					MaxStack:   1,
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
				Notes: GiftNotesConfig{
					Enabled:   true,
					MaxLength: 100,
				},
			},
			expectError: false,
		},
		{
			name: "empty_id",
			gift: GiftDefinition{
				ID:          "",
				Name:        "Gift",
				Description: "Description",
				Category:    "food",
				Rarity:      "common",
			},
			expectError:   true,
			errorContains: "gift ID is required",
		},
		{
			name: "id_too_long",
			gift: GiftDefinition{
				ID:          "this_is_a_very_long_gift_id_that_exceeds_fifty_characters_limit",
				Name:        "Gift",
				Description: "Description",
				Category:    "food",
				Rarity:      "common",
			},
			expectError:   true,
			errorContains: "gift ID must be 50 characters or less",
		},
		{
			name: "invalid_max_stack_zero",
			gift: GiftDefinition{
				ID:          "gift",
				Name:        "Gift",
				Description: "Description",
				Category:    "food",
				Rarity:      "common",
				Properties: GiftProperties{
					MaxStack: 0,
				},
			},
			expectError:   true,
			errorContains: "maxStack must be at least 1",
		},
		{
			name: "non_stackable_with_multiple_stack",
			gift: GiftDefinition{
				ID:          "gift",
				Name:        "Gift",
				Description: "Description",
				Category:    "food",
				Rarity:      "common",
				Properties: GiftProperties{
					Stackable: false,
					MaxStack:  5,
				},
			},
			expectError:   true,
			errorContains: "non-stackable gifts cannot have maxStack > 1",
		},
		{
			name: "stat_effect_too_high",
			gift: GiftDefinition{
				ID:          "gift",
				Name:        "Gift",
				Description: "Description",
				Category:    "food",
				Rarity:      "common",
				Properties: GiftProperties{
					MaxStack: 1,
				},
				GiftEffects: GiftEffects{
					Immediate: ImmediateEffects{
						Stats:     map[string]float64{"happiness": 150},
						Responses: []string{"Thank you!"},
					},
					Memory: MemoryEffects{
						Importance: 0.5,
					},
				},
			},
			expectError:   true,
			errorContains: "stat effect for 'happiness' must be between -100 and 100",
		},
		{
			name: "no_responses",
			gift: GiftDefinition{
				ID:          "gift",
				Name:        "Gift",
				Description: "Description",
				Category:    "food",
				Rarity:      "common",
				Properties: GiftProperties{
					MaxStack: 1,
				},
				GiftEffects: GiftEffects{
					Immediate: ImmediateEffects{
						Stats:     map[string]float64{"happiness": 10},
						Responses: []string{},
					},
					Memory: MemoryEffects{
						Importance: 0.5,
					},
				},
			},
			expectError:   true,
			errorContains: "at least one response is required",
		},
		{
			name: "memory_importance_out_of_range",
			gift: GiftDefinition{
				ID:          "gift",
				Name:        "Gift",
				Description: "Description",
				Category:    "food",
				Rarity:      "common",
				Properties: GiftProperties{
					MaxStack: 1,
				},
				GiftEffects: GiftEffects{
					Immediate: ImmediateEffects{
						Stats:     map[string]float64{"happiness": 10},
						Responses: []string{"Thank you!"},
					},
					Memory: MemoryEffects{
						Importance: 1.5,
					},
				},
			},
			expectError:   true,
			errorContains: "memory importance must be between 0 and 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.gift.Validate()

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}
				if tt.errorContains != "" && !containsString(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorContains, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestLoadGiftCatalog tests loading multiple gift definitions from a directory
func TestLoadGiftCatalog(t *testing.T) {
	// Create temporary directory with multiple gift files
	tmpDir := t.TempDir()

	// Create valid gift files
	gift1 := `{
		"id": "gift1",
		"name": "Gift 1",
		"description": "First gift",
		"category": "food",
		"rarity": "common",
		"image": "gift1.gif",
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

	gift2 := `{
		"id": "gift2",
		"name": "Gift 2",
		"description": "Second gift",
		"category": "flowers",
		"rarity": "rare",
		"image": "gift2.gif",
		"properties": {
			"consumable": false,
			"stackable": true,
			"maxStack": 5,
			"unlockRequirements": {}
		},
		"giftEffects": {
			"immediate": {
				"stats": {"affection": 15},
				"animations": ["happy"],
				"responses": ["Beautiful!"]
			},
			"memory": {
				"importance": 0.7,
				"tags": ["romantic"],
				"emotionalTone": "romantic"
			}
		},
		"personalityModifiers": {},
		"notes": {
			"enabled": true,
			"maxLength": 150,
			"placeholder": "Add a romantic note..."
		}
	}`

	err := os.WriteFile(filepath.Join(tmpDir, "gift1.json"), []byte(gift1), 0644)
	if err != nil {
		t.Fatalf("Failed to create gift1.json: %v", err)
	}

	err = os.WriteFile(filepath.Join(tmpDir, "gift2.json"), []byte(gift2), 0644)
	if err != nil {
		t.Fatalf("Failed to create gift2.json: %v", err)
	}

	// Add a non-JSON file that should be ignored
	err = os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("This should be ignored"), 0644)
	if err != nil {
		t.Fatalf("Failed to create readme.txt: %v", err)
	}

	// Test loading catalog
	catalog, err := LoadGiftCatalog(tmpDir)
	if err != nil {
		t.Fatalf("Unexpected error loading catalog: %v", err)
	}

	// Verify catalog contents
	expectedCount := 2
	if len(catalog) != expectedCount {
		t.Errorf("Expected %d gifts in catalog, got %d", expectedCount, len(catalog))
	}

	// Verify specific gifts are loaded
	if _, exists := catalog["gift1"]; !exists {
		t.Error("Expected gift1 to be in catalog")
	}
	if _, exists := catalog["gift2"]; !exists {
		t.Error("Expected gift2 to be in catalog")
	}

	// Verify gift details
	if gift1Def, exists := catalog["gift1"]; exists {
		if gift1Def.Name != "Gift 1" {
			t.Errorf("Expected gift1 name 'Gift 1', got '%s'", gift1Def.Name)
		}
		if gift1Def.Category != "food" {
			t.Errorf("Expected gift1 category 'food', got '%s'", gift1Def.Category)
		}
	}
}

// TestLoadGiftCatalogNonexistentDirectory tests loading from a directory that doesn't exist
func TestLoadGiftCatalogNonexistentDirectory(t *testing.T) {
	catalog, err := LoadGiftCatalog("/nonexistent/directory")
	if err != nil {
		t.Errorf("Expected no error for nonexistent directory, got: %v", err)
	}
	if len(catalog) != 0 {
		t.Errorf("Expected empty catalog for nonexistent directory, got %d items", len(catalog))
	}
}

// TestLoadGiftCatalogDuplicateIDs tests that duplicate gift IDs are rejected
func TestLoadGiftCatalogDuplicateIDs(t *testing.T) {
	tmpDir := t.TempDir()

	giftContent := `{
		"id": "duplicate_id",
		"name": "Gift",
		"description": "A gift",
		"category": "food",
		"rarity": "common",
		"image": "gift.gif",
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

	// Create two files with the same gift ID
	err := os.WriteFile(filepath.Join(tmpDir, "gift1.json"), []byte(giftContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create gift1.json: %v", err)
	}

	err = os.WriteFile(filepath.Join(tmpDir, "gift2.json"), []byte(giftContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create gift2.json: %v", err)
	}

	// Test loading catalog - should fail with duplicate ID error
	_, err = LoadGiftCatalog(tmpDir)
	if err == nil {
		t.Error("Expected error for duplicate gift IDs, got none")
	}
	if !containsString(err.Error(), "duplicate gift ID") {
		t.Errorf("Expected error about duplicate gift ID, got: %v", err)
	}
}

// containsString is a helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsStringRecursive(s, substr))
}

func containsStringRecursive(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	if s[:len(substr)] == substr {
		return true
	}
	return containsStringRecursive(s[1:], substr)
}

// BenchmarkLoadGiftDefinition benchmarks gift definition loading performance
func BenchmarkLoadGiftDefinition(b *testing.B) {
	// Create temporary gift file
	tmpDir := b.TempDir()
	giftFile := filepath.Join(tmpDir, "benchmark_gift.json")

	giftContent := `{
		"id": "benchmark_gift",
		"name": "Benchmark Gift",
		"description": "A gift for benchmarking",
		"category": "food",
		"rarity": "common",
		"image": "benchmark.gif",
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
				"tags": ["benchmark"],
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
		b.Fatalf("Failed to create benchmark gift file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := LoadGiftDefinition(giftFile)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}

// BenchmarkGiftValidation benchmarks gift validation performance
func BenchmarkGiftValidation(b *testing.B) {
	gift := GiftDefinition{
		ID:          "benchmark_gift",
		Name:        "Benchmark Gift",
		Description: "A gift for benchmarking validation",
		Category:    "food",
		Rarity:      "common",
		Properties: GiftProperties{
			Consumable: true,
			Stackable:  false,
			MaxStack:   1,
		},
		GiftEffects: GiftEffects{
			Immediate: ImmediateEffects{
				Stats:      map[string]float64{"happiness": 10, "affection": 5},
				Responses:  []string{"Thank you!", "Amazing!", "I love it!"},
				Animations: []string{"happy", "excited"},
			},
			Memory: MemoryEffects{
				Importance:    0.5,
				Tags:          []string{"benchmark", "test", "gift"},
				EmotionalTone: "happy",
			},
		},
		PersonalityModifiers: map[string]map[string]float64{
			"shy":      {"affection": 1.2, "trust": 1.1},
			"romantic": {"affection": 1.5, "intimacy": 1.3},
		},
		Notes: GiftNotesConfig{
			Enabled:     true,
			MaxLength:   200,
			Placeholder: "Add a personal message...",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := gift.Validate()
		if err != nil {
			b.Fatalf("Validation failed: %v", err)
		}
	}
}
