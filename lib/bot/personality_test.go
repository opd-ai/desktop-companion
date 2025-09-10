package bot

import (
	"encoding/json"
	"testing"
	"time"
)

func TestPersonalityTraits(t *testing.T) {
	// Test that all personality traits are properly defined
	traits := []string{
		PersonalityTraits.Chattiness,
		PersonalityTraits.Helpfulness,
		PersonalityTraits.Playfulness,
		PersonalityTraits.Curiosity,
		PersonalityTraits.Empathy,
		PersonalityTraits.Assertiveness,
		PersonalityTraits.Patience,
		PersonalityTraits.Enthusiasm,
		PersonalityTraits.Independence,
		PersonalityTraits.Creativity,
		PersonalityTraits.Analytical,
		PersonalityTraits.Spontaneity,
	}

	for _, trait := range traits {
		if trait == "" {
			t.Errorf("Personality trait is empty")
		}
		if len(trait) < 3 {
			t.Errorf("Personality trait '%s' is too short", trait)
		}
	}

	// Test no duplicate trait names
	seen := make(map[string]bool)
	for _, trait := range traits {
		if seen[trait] {
			t.Errorf("Duplicate personality trait: %s", trait)
		}
		seen[trait] = true
	}
}

func TestPersonalityBehavior_ParseResponseDelay(t *testing.T) {
	tests := []struct {
		name        string
		delay       string
		expectMin   time.Duration
		expectMax   time.Duration
		expectError bool
	}{
		{
			name:      "Empty delay uses default",
			delay:     "",
			expectMin: 2 * time.Second,
			expectMax: 4 * time.Second,
		},
		{
			name:      "Single value with variation",
			delay:     "2s",
			expectMin: 1500 * time.Millisecond, // 2s - 25%
			expectMax: 2500 * time.Millisecond, // 2s + 25%
		},
		{
			name:      "Range format seconds",
			delay:     "1s-3s",
			expectMin: 1 * time.Second,
			expectMax: 3 * time.Second,
		},
		{
			name:      "Range format milliseconds",
			delay:     "500ms-2s",
			expectMin: 500 * time.Millisecond,
			expectMax: 2 * time.Second,
		},
		{
			name:      "Range with spaces",
			delay:     " 1s - 4s ",
			expectMin: 1 * time.Second,
			expectMax: 4 * time.Second,
		},
		{
			name:        "Invalid range format",
			delay:       "1-2-3s",
			expectError: true,
		},
		{
			name:        "Invalid duration",
			delay:       "invalid",
			expectError: true,
		},
		{
			name:        "Min greater than max",
			delay:       "5s-2s",
			expectError: true,
		},
		{
			name:        "Invalid range min",
			delay:       "invalid-2s",
			expectError: true,
		},
		{
			name:        "Invalid range max",
			delay:       "1s-invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			behavior := PersonalityBehavior{ResponseDelay: tt.delay}
			min, max, err := behavior.ParseResponseDelay()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if min != tt.expectMin {
				t.Errorf("Expected min %v, got %v", tt.expectMin, min)
			}

			if max != tt.expectMax {
				t.Errorf("Expected max %v, got %v", tt.expectMax, max)
			}

			// Ensure min <= max
			if min > max {
				t.Errorf("Min delay (%v) should not exceed max delay (%v)", min, max)
			}
		})
	}
}

func TestNewPersonalityManager(t *testing.T) {
	pm := NewPersonalityManager()

	if pm == nil {
		t.Fatal("NewPersonalityManager returned nil")
	}

	if pm.archetypes == nil {
		t.Fatal("Archetypes map not initialized")
	}

	// Check that built-in archetypes are loaded
	archetypes := pm.ListArchetypes()
	if len(archetypes) == 0 {
		t.Error("No built-in archetypes loaded")
	}

	expectedArchetypes := []string{"social", "shy", "playful", "helper", "balanced"}
	for _, expected := range expectedArchetypes {
		found := false
		for _, actual := range archetypes {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected archetype '%s' not found in: %v", expected, archetypes)
		}
	}
}

func TestPersonalityManager_GetArchetype(t *testing.T) {
	pm := NewPersonalityManager()

	tests := []struct {
		name        string
		archetype   string
		expectError bool
	}{
		{
			name:      "Get social archetype",
			archetype: "social",
		},
		{
			name:      "Get balanced archetype",
			archetype: "balanced",
		},
		{
			name:      "Case insensitive lookup",
			archetype: "SOCIAL",
		},
		{
			name:        "Nonexistent archetype",
			archetype:   "nonexistent",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archetype, err := pm.GetArchetype(tt.archetype)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if archetype != nil {
					t.Error("Expected nil archetype on error")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if archetype == nil {
				t.Error("Expected archetype but got nil")
				return
			}

			if archetype.Name == "" {
				t.Error("Archetype name is empty")
			}

			if archetype.Description == "" {
				t.Error("Archetype description is empty")
			}

			if len(archetype.Traits) == 0 {
				t.Error("Archetype has no traits")
			}
		})
	}
}

func TestPersonalityManager_CreatePersonality(t *testing.T) {
	pm := NewPersonalityManager()

	tests := []struct {
		name        string
		archetype   *PersonalityArchetype
		expectError bool
	}{
		{
			name: "Valid archetype",
			archetype: &PersonalityArchetype{
				Name:        "Test",
				Description: "Test archetype",
				Traits: map[string]float64{
					PersonalityTraits.Chattiness:  0.8,
					PersonalityTraits.Helpfulness: 0.7,
					PersonalityTraits.Empathy:     0.9,
				},
				Behavior: PersonalityBehavior{
					ResponseDelay:       "2s-4s",
					InteractionRate:     3.0,
					Attention:           0.8,
					MaxActionsPerMinute: 6,
					MinTimeBetweenSame:  10,
					PreferredActions:    []string{"chat", "click"},
				},
			},
		},
		{
			name:        "Nil archetype",
			archetype:   nil,
			expectError: true,
		},
		{
			name: "Invalid response delay",
			archetype: &PersonalityArchetype{
				Name:        "Test",
				Description: "Test archetype",
				Traits:      map[string]float64{PersonalityTraits.Chattiness: 0.8},
				Behavior: PersonalityBehavior{
					ResponseDelay:   "invalid",
					InteractionRate: 3.0,
					Attention:       0.8,
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			personality, err := pm.CreatePersonality(tt.archetype)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if personality != nil {
					t.Error("Expected nil personality on error")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

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

			// Check trait categorization
			if tt.archetype != nil {
				for trait, value := range tt.archetype.Traits {
					found := false

					// Check social tendencies
					if socialValue, exists := personality.SocialTendencies[trait]; exists {
						if socialValue != value {
							t.Errorf("Social trait %s: expected %f, got %f", trait, value, socialValue)
						}
						found = true
					}

					// Check emotional profile
					if emotionalValue, exists := personality.EmotionalProfile[trait]; exists {
						if emotionalValue != value {
							t.Errorf("Emotional trait %s: expected %f, got %f", trait, value, emotionalValue)
						}
						found = true
					}

					if !found {
						t.Errorf("Trait %s not found in either social tendencies or emotional profile", trait)
					}
				}
			}
		})
	}
}

func TestPersonalityManager_LoadFromJSON(t *testing.T) {
	pm := NewPersonalityManager()

	tests := []struct {
		name        string
		jsonData    string
		expectError bool
	}{
		{
			name: "Valid JSON archetype",
			jsonData: `{
				"name": "Custom",
				"description": "Custom test archetype",
				"traits": {
					"chattiness": 0.8,
					"helpfulness": 0.7
				},
				"behavior": {
					"responseDelay": "2s-4s",
					"interactionRate": 3.0,
					"attention": 0.8,
					"maxActionsPerMinute": 6,
					"minTimeBetweenSame": 10,
					"preferredActions": ["chat", "click"]
				}
			}`,
		},
		{
			name:        "Invalid JSON",
			jsonData:    `{"invalid": json}`,
			expectError: true,
		},
		{
			name: "Missing required fields",
			jsonData: `{
				"description": "Missing name"
			}`,
			expectError: true,
		},
		{
			name: "Trait value out of range",
			jsonData: `{
				"name": "Invalid",
				"description": "Invalid trait values",
				"traits": {
					"chattiness": 1.5
				},
				"behavior": {
					"responseDelay": "2s",
					"interactionRate": 3.0,
					"attention": 0.8,
					"maxActionsPerMinute": 6,
					"minTimeBetweenSame": 10
				}
			}`,
			expectError: true,
		},
		{
			name: "Invalid interaction rate",
			jsonData: `{
				"name": "Invalid",
				"description": "Invalid interaction rate",
				"traits": {
					"chattiness": 0.8
				},
				"behavior": {
					"responseDelay": "2s",
					"interactionRate": 15.0,
					"attention": 0.8,
					"maxActionsPerMinute": 6,
					"minTimeBetweenSame": 10
				}
			}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archetype, err := pm.LoadFromJSON([]byte(tt.jsonData))

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if archetype != nil {
					t.Error("Expected nil archetype on error")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if archetype == nil {
				t.Error("Expected archetype but got nil")
				return
			}

			if archetype.Name == "" {
				t.Error("Archetype name is empty")
			}
		})
	}
}

func TestBuiltinArchetypeValidation(t *testing.T) {
	pm := NewPersonalityManager()
	archetypes := pm.ListArchetypes()

	for _, name := range archetypes {
		t.Run("archetype_"+name, func(t *testing.T) {
			archetype, err := pm.GetArchetype(name)
			if err != nil {
				t.Fatalf("Failed to get archetype %s: %v", name, err)
			}

			// Test that archetype can be converted to personality
			personality, err := pm.CreatePersonality(archetype)
			if err != nil {
				t.Errorf("Failed to create personality from archetype %s: %v", name, err)
				return
			}

			// Validate personality fields
			if personality.ResponseDelay <= 0 {
				t.Errorf("Archetype %s has invalid response delay: %v", name, personality.ResponseDelay)
			}

			if personality.InteractionRate <= 0 {
				t.Errorf("Archetype %s has invalid interaction rate: %f", name, personality.InteractionRate)
			}

			if personality.Attention < 0 || personality.Attention > 1 {
				t.Errorf("Archetype %s has invalid attention value: %f", name, personality.Attention)
			}

			// Validate trait values
			for trait, value := range personality.SocialTendencies {
				if value < 0 || value > 1 {
					t.Errorf("Archetype %s social trait %s has invalid value: %f", name, trait, value)
				}
			}

			for trait, value := range personality.EmotionalProfile {
				if value < 0 || value > 1 {
					t.Errorf("Archetype %s emotional trait %s has invalid value: %f", name, trait, value)
				}
			}
		})
	}
}

func TestPersonalityArchetypeCharacteristics(t *testing.T) {
	pm := NewPersonalityManager()

	// Test that different archetypes have distinct characteristics
	social, _ := pm.GetArchetype("social")
	shy, _ := pm.GetArchetype("shy")
	playful, _ := pm.GetArchetype("playful")

	// Social should be more chatty than shy
	socialChattiness := social.Traits[PersonalityTraits.Chattiness]
	shyChattiness := shy.Traits[PersonalityTraits.Chattiness]
	if socialChattiness <= shyChattiness {
		t.Errorf("Social archetype should be more chatty than shy: social=%.2f, shy=%.2f",
			socialChattiness, shyChattiness)
	}

	// Playful should be more playful than others
	playfulPlayfulness := playful.Traits[PersonalityTraits.Playfulness]
	socialPlayfulness := social.Traits[PersonalityTraits.Playfulness]
	if playfulPlayfulness <= socialPlayfulness {
		t.Errorf("Playful archetype should be more playful than social: playful=%.2f, social=%.2f",
			playfulPlayfulness, socialPlayfulness)
	}

	// Test interaction rate differences
	if social.Behavior.InteractionRate <= shy.Behavior.InteractionRate {
		t.Errorf("Social archetype should have higher interaction rate than shy: social=%.2f, shy=%.2f",
			social.Behavior.InteractionRate, shy.Behavior.InteractionRate)
	}
}

func TestPersonalityArchetypeJSON(t *testing.T) {
	pm := NewPersonalityManager()

	// Test that archetypes can be round-tripped through JSON
	archetype, err := pm.GetArchetype("balanced")
	if err != nil {
		t.Fatalf("Failed to get balanced archetype: %v", err)
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(archetype)
	if err != nil {
		t.Fatalf("Failed to marshal archetype to JSON: %v", err)
	}

	// Unmarshal back
	recreated, err := pm.LoadFromJSON(jsonData)
	if err != nil {
		t.Fatalf("Failed to load archetype from JSON: %v", err)
	}

	// Compare key fields
	if recreated.Name != archetype.Name {
		t.Errorf("Name mismatch: expected %s, got %s", archetype.Name, recreated.Name)
	}

	if recreated.Description != archetype.Description {
		t.Errorf("Description mismatch: expected %s, got %s", archetype.Description, recreated.Description)
	}

	if len(recreated.Traits) != len(archetype.Traits) {
		t.Errorf("Traits count mismatch: expected %d, got %d", len(archetype.Traits), len(recreated.Traits))
	}
}

// Benchmarks to ensure personality operations are efficient
func BenchmarkPersonalityManager_GetArchetype(b *testing.B) {
	pm := NewPersonalityManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pm.GetArchetype("social")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPersonalityManager_CreatePersonality(b *testing.B) {
	pm := NewPersonalityManager()
	archetype, _ := pm.GetArchetype("balanced")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pm.CreatePersonality(archetype)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPersonalityBehavior_ParseResponseDelay(b *testing.B) {
	behavior := PersonalityBehavior{ResponseDelay: "2-4s"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := behavior.ParseResponseDelay()
		if err != nil {
			b.Fatal(err)
		}
	}
}
