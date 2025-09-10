package character

import (
	"github.com/opd-ai/desktop-companion/lib/platform"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPlatformAwareLoader_NewLoader tests the creation of platform-aware loaders.
func TestPlatformAwareLoader_NewLoader(t *testing.T) {
	loader := NewPlatformAwareLoader()
	if loader == nil {
		t.Fatal("expected non-nil loader")
	}
	if loader.platform == nil {
		t.Fatal("expected platform info to be initialized")
	}
}

// TestPlatformAwareLoader_LoadCharacterCard tests loading character cards with platform adaptations.
func TestPlatformAwareLoader_LoadCharacterCard(t *testing.T) {
	tests := []struct {
		name          string
		cardData      string
		expectError   bool
		expectAdapted bool
	}{
		{
			name: "basic card without platform config",
			cardData: `{
				"name": "Test Character",
				"description": "A test character",
				"animations": {"idle": "idle.gif", "talking": "talking.gif"},
				"dialogs": [{"trigger": "click", "responses": ["Hello!"], "animation": "idle", "cooldown": 5}],
				"behavior": {"idleTimeout": 30, "movementEnabled": true, "defaultSize": 128}
			}`,
			expectError:   false,
			expectAdapted: false,
		},
		{
			name: "card with platform config",
			cardData: `{
				"name": "Cross-Platform Character",
				"description": "A character with platform-specific settings",
				"animations": {"idle": "idle.gif", "talking": "talking.gif"},
				"dialogs": [{"trigger": "click", "responses": ["Hello!"], "animation": "idle", "cooldown": 5}],
				"behavior": {"idleTimeout": 30, "movementEnabled": true, "defaultSize": 128},
				"platformConfig": {
					"desktop": {
						"behavior": {"defaultSize": 128, "movementEnabled": true},
						"windowMode": "overlay"
					},
					"mobile": {
						"behavior": {"defaultSize": 256, "movementEnabled": false},
						"windowMode": "fullscreen",
						"touchOptimized": true
					}
				}
			}`,
			expectError:   false,
			expectAdapted: true,
		},
		{
			name: "invalid platform config",
			cardData: `{
				"name": "Invalid Character",
				"description": "A character with invalid platform settings",
				"animations": {"idle": "idle.gif", "talking": "talking.gif"},
				"dialogs": [{"trigger": "click", "responses": ["Hello!"], "animation": "idle", "cooldown": 5}],
				"behavior": {"idleTimeout": 30, "movementEnabled": true, "defaultSize": 128},
				"platformConfig": {
					"desktop": {
						"behavior": {"defaultSize": -1},
						"windowMode": "invalid"
					}
				}
			}`,
			expectError:   true,
			expectAdapted: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary test file
			tmpDir := t.TempDir()
			cardPath := filepath.Join(tmpDir, "character.json")
			animPath := filepath.Join(tmpDir, "idle.gif")
			talkingPath := filepath.Join(tmpDir, "talking.gif")

			// Write test data
			err := os.WriteFile(cardPath, []byte(tt.cardData), 0644)
			if err != nil {
				t.Fatalf("failed to write test card: %v", err)
			}

			// Create dummy animation files
			err = os.WriteFile(animPath, []byte("GIF89a"), 0644)
			if err != nil {
				t.Fatalf("failed to write test animation: %v", err)
			}

			err = os.WriteFile(talkingPath, []byte("GIF89a"), 0644)
			if err != nil {
				t.Fatalf("failed to write talking animation: %v", err)
			}

			loader := NewPlatformAwareLoader()
			card, err := loader.LoadCharacterCard(cardPath)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if card == nil {
				t.Fatal("expected non-nil card")
			}

			// Verify basic properties are preserved
			if card.Name == "" {
				t.Error("expected name to be preserved")
			}

			if len(card.Animations) == 0 {
				t.Error("expected animations to be preserved")
			}

			// Check if platform adaptations were applied when expected
			if tt.expectAdapted {
				platformConfig := loader.GetPlatformConfig(card)
				if platformConfig == nil && card.PlatformConfig != nil {
					t.Error("expected platform config to be accessible")
				}
			}
		})
	}
}

// TestPlatformAwareLoader_ApplyPlatformConfig tests platform configuration application logic.
func TestPlatformAwareLoader_ApplyPlatformConfig(t *testing.T) {
	tests := []struct {
		name           string
		platformType   string
		baseCard       *CharacterCard
		expectedSize   int
		expectedMoving bool
	}{
		{
			name:         "desktop platform config",
			platformType: "desktop",
			baseCard: &CharacterCard{
				Name:        "Test",
				Description: "Test character",
				Animations:  map[string]string{"idle": "idle.gif"},
				Behavior:    Behavior{DefaultSize: 100, MovementEnabled: false},
				PlatformConfig: &PlatformConfig{
					Desktop: &PlatformSpecificConfig{
						Behavior: &Behavior{DefaultSize: 150, MovementEnabled: true},
					},
				},
			},
			expectedSize:   150,
			expectedMoving: true,
		},
		{
			name:         "mobile platform config",
			platformType: "mobile",
			baseCard: &CharacterCard{
				Name:        "Test",
				Description: "Test character",
				Animations:  map[string]string{"idle": "idle.gif"},
				Behavior:    Behavior{DefaultSize: 100, MovementEnabled: true},
				PlatformConfig: &PlatformConfig{
					Mobile: &PlatformSpecificConfig{
						Behavior: &Behavior{DefaultSize: 200, MovementEnabled: false},
					},
				},
			},
			expectedSize:   200,
			expectedMoving: false,
		},
		{
			name:         "no platform config",
			platformType: "desktop",
			baseCard: &CharacterCard{
				Name:        "Test",
				Description: "Test character",
				Animations:  map[string]string{"idle": "idle.gif"},
				Behavior:    Behavior{DefaultSize: 100, MovementEnabled: true},
			},
			expectedSize:   100,
			expectedMoving: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create loader with mock platform info
			loader := &PlatformAwareLoader{
				platform: &platform.PlatformInfo{
					OS:         tt.platformType,
					FormFactor: tt.platformType,
				},
			}

			adaptedCard := loader.applyPlatformConfig(tt.baseCard)

			if adaptedCard.Behavior.DefaultSize != tt.expectedSize {
				t.Errorf("expected size %d, got %d", tt.expectedSize, adaptedCard.Behavior.DefaultSize)
			}

			if adaptedCard.Behavior.MovementEnabled != tt.expectedMoving {
				t.Errorf("expected movementEnabled %v, got %v", tt.expectedMoving, adaptedCard.Behavior.MovementEnabled)
			}

			// Verify original card is not modified
			if tt.baseCard.PlatformConfig != nil {
				// Original should remain unchanged
				if tt.baseCard.Behavior.DefaultSize != 100 {
					t.Error("original card was modified")
				}
			}
		})
	}
}

// TestPlatformAwareLoader_AdaptTriggers tests trigger adaptation for different platforms.
func TestPlatformAwareLoader_AdaptTriggers(t *testing.T) {
	tests := []struct {
		name           string
		platformType   string
		inputTriggers  []string
		expectedOutput []string
	}{
		{
			name:           "mobile touch triggers",
			platformType:   "mobile",
			inputTriggers:  []string{"tap", "longpress", "doubletap"},
			expectedOutput: []string{"click", "rightclick", "doubleclick"},
		},
		{
			name:           "desktop mouse triggers",
			platformType:   "desktop",
			inputTriggers:  []string{"tap", "longpress", "doubletap"},
			expectedOutput: []string{"click", "rightclick", "doubleclick"},
		},
		{
			name:           "standard triggers pass through",
			platformType:   "desktop",
			inputTriggers:  []string{"click", "rightclick", "hover"},
			expectedOutput: []string{"click", "rightclick", "hover"},
		},
		{
			name:           "swipe gesture fallback",
			platformType:   "mobile",
			inputTriggers:  []string{"swipe"},
			expectedOutput: []string{"click"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := &PlatformAwareLoader{
				platform: &platform.PlatformInfo{
					OS:         tt.platformType,
					FormFactor: tt.platformType,
				},
			}

			result := loader.adaptTriggers(tt.inputTriggers)

			if len(result) != len(tt.expectedOutput) {
				t.Fatalf("expected %d triggers, got %d", len(tt.expectedOutput), len(result))
			}

			for i, expected := range tt.expectedOutput {
				if result[i] != expected {
					t.Errorf("trigger %d: expected %s, got %s", i, expected, result[i])
				}
			}
		})
	}
}

// TestValidatePlatformConfig tests platform configuration validation.
func TestValidatePlatformConfig(t *testing.T) {
	tests := []struct {
		name        string
		card        *CharacterCard
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid platform config",
			card: &CharacterCard{
				PlatformConfig: &PlatformConfig{
					Desktop: &PlatformSpecificConfig{
						Behavior:   &Behavior{DefaultSize: 128, IdleTimeout: 30},
						WindowMode: "overlay",
					},
					Mobile: &PlatformSpecificConfig{
						Behavior:       &Behavior{DefaultSize: 256, IdleTimeout: 60},
						WindowMode:     "fullscreen",
						TouchOptimized: true,
						MobileControls: &MobileControlsConfig{
							ShowBottomBar:  true,
							HapticFeedback: true,
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "invalid window mode",
			card: &CharacterCard{
				PlatformConfig: &PlatformConfig{
					Desktop: &PlatformSpecificConfig{
						WindowMode: "invalid_mode",
					},
				},
			},
			expectError: true,
			errorMsg:    "invalid window mode",
		},
		{
			name: "negative idle timeout",
			card: &CharacterCard{
				PlatformConfig: &PlatformConfig{
					Mobile: &PlatformSpecificConfig{
						Behavior: &Behavior{IdleTimeout: -1},
					},
				},
			},
			expectError: true,
			errorMsg:    "idle timeout cannot be negative",
		},
		{
			name: "invalid default size",
			card: &CharacterCard{
				PlatformConfig: &PlatformConfig{
					Desktop: &PlatformSpecificConfig{
						Behavior: &Behavior{DefaultSize: 2000},
					},
				},
			},
			expectError: true,
			errorMsg:    "default size must be between 32 and 1024 pixels",
		},
		{
			name: "mobile controls on desktop",
			card: &CharacterCard{
				PlatformConfig: &PlatformConfig{
					Desktop: &PlatformSpecificConfig{
						MobileControls: &MobileControlsConfig{
							ShowBottomBar: true,
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "mobile controls configuration only valid for mobile platform",
		},
		{
			name: "no platform config",
			card: &CharacterCard{
				Name: "Simple Character",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePlatformConfig(tt.card)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestCreateExamplePlatformConfig tests the example configuration generator.
func TestCreateExamplePlatformConfig(t *testing.T) {
	config := CreateExamplePlatformConfig()

	if config == nil {
		t.Fatal("expected non-nil config")
	}

	if config.Desktop == nil {
		t.Error("expected desktop config")
	}

	if config.Mobile == nil {
		t.Error("expected mobile config")
	}

	// Test desktop config
	if config.Desktop.Behavior == nil {
		t.Error("expected desktop behavior config")
	} else {
		if config.Desktop.Behavior.DefaultSize != 128 {
			t.Errorf("expected desktop size 128, got %d", config.Desktop.Behavior.DefaultSize)
		}
		if !config.Desktop.Behavior.MovementEnabled {
			t.Error("expected desktop movement enabled")
		}
	}

	// Test mobile config
	if config.Mobile.Behavior == nil {
		t.Error("expected mobile behavior config")
	} else {
		if config.Mobile.Behavior.DefaultSize != 256 {
			t.Errorf("expected mobile size 256, got %d", config.Mobile.Behavior.DefaultSize)
		}
		if config.Mobile.Behavior.MovementEnabled {
			t.Error("expected mobile movement disabled")
		}
	}

	if config.Mobile.MobileControls == nil {
		t.Error("expected mobile controls config")
	}

	// Validate the example config
	testCard := &CharacterCard{
		PlatformConfig: config,
	}

	if err := ValidatePlatformConfig(testCard); err != nil {
		t.Errorf("example config failed validation: %v", err)
	}
}

// TestPlatformAwareLoader_GetPlatformConfig tests getting platform-specific configuration.
func TestPlatformAwareLoader_GetPlatformConfig(t *testing.T) {
	config := &PlatformConfig{
		Desktop: &PlatformSpecificConfig{WindowMode: "overlay"},
		Mobile:  &PlatformSpecificConfig{WindowMode: "fullscreen"},
	}

	card := &CharacterCard{PlatformConfig: config}

	tests := []struct {
		name         string
		platformType string
		expectedMode string
	}{
		{
			name:         "desktop platform",
			platformType: "desktop",
			expectedMode: "overlay",
		},
		{
			name:         "mobile platform",
			platformType: "mobile",
			expectedMode: "fullscreen",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := &PlatformAwareLoader{
				platform: &platform.PlatformInfo{
					FormFactor: tt.platformType,
				},
			}

			platformConfig := loader.GetPlatformConfig(card)

			if platformConfig == nil {
				t.Fatal("expected non-nil platform config")
			}

			if platformConfig.WindowMode != tt.expectedMode {
				t.Errorf("expected window mode %s, got %s", tt.expectedMode, platformConfig.WindowMode)
			}
		})
	}

	// Test card without platform config
	cardWithoutConfig := &CharacterCard{}
	loader := NewPlatformAwareLoader()
	result := loader.GetPlatformConfig(cardWithoutConfig)
	if result != nil {
		t.Error("expected nil for card without platform config")
	}
}

// TestPlatformAwareLoader_MergeBehavior tests behavior merging logic.
func TestPlatformAwareLoader_MergeBehavior(t *testing.T) {
	loader := NewPlatformAwareLoader()

	base := Behavior{
		IdleTimeout:     30,
		MovementEnabled: true,
		DefaultSize:     100,
	}

	override := Behavior{
		IdleTimeout:     60,
		MovementEnabled: false,
		DefaultSize:     200,
	}

	merged := loader.mergeBehavior(base, override)

	if merged.IdleTimeout != 60 {
		t.Errorf("expected idle timeout 60, got %d", merged.IdleTimeout)
	}

	if merged.DefaultSize != 200 {
		t.Errorf("expected default size 200, got %d", merged.DefaultSize)
	}

	// Movement enabled logic depends on platform
	// This test assumes desktop platform
	expectedMovement := false
	if loader.platform.IsDesktop() {
		expectedMovement = true // Desktop preserves override || base
	}

	if merged.MovementEnabled != expectedMovement {
		t.Errorf("expected movement enabled %v, got %v", expectedMovement, merged.MovementEnabled)
	}
}

// TestPlatformAwareLoader_MergeInteractions tests interaction merging logic.
func TestPlatformAwareLoader_MergeInteractions(t *testing.T) {
	loader := NewPlatformAwareLoader()

	base := map[string]InteractionConfig{
		"pet": {
			Triggers: []string{"click"},
			Effects:  map[string]float64{"happiness": 10},
			Cooldown: 5,
		},
	}

	overrides := map[string]PlatformInteractionConfig{
		"pet": {
			InteractionConfig: InteractionConfig{
				Effects:  map[string]float64{"happiness": 15},
				Cooldown: 3,
			},
			Triggers:      []string{"tap"},
			HapticPattern: "light",
		},
		"feed": {
			InteractionConfig: InteractionConfig{
				Effects:  map[string]float64{"hunger": 20},
				Cooldown: 10,
			},
			Triggers: []string{"longpress"},
		},
	}

	merged := loader.mergeInteractions(base, overrides)

	// Check that base interaction was overridden
	if pet, exists := merged["pet"]; exists {
		if pet.Effects["happiness"] != 15 {
			t.Errorf("expected happiness effect 15, got %f", pet.Effects["happiness"])
		}
		if pet.Cooldown != 3 {
			t.Errorf("expected cooldown 3, got %d", pet.Cooldown)
		}
		// Triggers should be adapted
		expectedTrigger := "click" // tap -> click adaptation
		if len(pet.Triggers) > 0 && pet.Triggers[0] != expectedTrigger {
			t.Errorf("expected trigger %s, got %s", expectedTrigger, pet.Triggers[0])
		}
	} else {
		t.Error("expected pet interaction to exist")
	}

	// Check that new interaction was added
	if feed, exists := merged["feed"]; exists {
		if feed.Effects["hunger"] != 20 {
			t.Errorf("expected hunger effect 20, got %f", feed.Effects["hunger"])
		}
	} else {
		t.Error("expected feed interaction to be added")
	}
}
