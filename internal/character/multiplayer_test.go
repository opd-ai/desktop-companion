package character

import (
	"testing"
)

// TestMultiplayerConfigValidation tests multiplayer configuration validation
func TestMultiplayerConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		multiplayer *MultiplayerConfig
		expectError bool
		errorText   string
	}{
		{
			name:        "nil multiplayer config",
			multiplayer: nil,
			expectError: false,
		},
		{
			name: "valid multiplayer config",
			multiplayer: &MultiplayerConfig{
				Enabled:       true,
				BotCapable:    true,
				NetworkID:     "test_character_v1",
				MaxPeers:      8,
				DiscoveryPort: 8080,
			},
			expectError: false,
		},
		{
			name: "disabled multiplayer with incomplete config",
			multiplayer: &MultiplayerConfig{
				Enabled:   false,
				NetworkID: "", // Empty but allowed when disabled
			},
			expectError: false,
		},
		{
			name: "enabled multiplayer missing networkID",
			multiplayer: &MultiplayerConfig{
				Enabled:    true,
				BotCapable: true,
				// Missing NetworkID
			},
			expectError: true,
			errorText:   "networkID is required",
		},
		{
			name: "networkID too long",
			multiplayer: &MultiplayerConfig{
				Enabled:   true,
				NetworkID: "this_network_id_is_way_too_long_and_exceeds_the_fifty_character_limit_by_far",
			},
			expectError: true,
			errorText:   "networkID too long",
		},
		{
			name: "networkID with invalid characters",
			multiplayer: &MultiplayerConfig{
				Enabled:   true,
				NetworkID: "test@character#1",
			},
			expectError: true,
			errorText:   "networkID contains invalid character",
		},
		{
			name: "valid networkID with allowed characters",
			multiplayer: &MultiplayerConfig{
				Enabled:   true,
				NetworkID: "Test_Character-v1_2024",
			},
			expectError: false,
		},
		{
			name: "maxPeers too high",
			multiplayer: &MultiplayerConfig{
				Enabled:   true,
				NetworkID: "test_char",
				MaxPeers:  20, // Above limit of 16
			},
			expectError: true,
			errorText:   "maxPeers cannot exceed 16",
		},
		{
			name: "valid maxPeers at boundary",
			multiplayer: &MultiplayerConfig{
				Enabled:   true,
				NetworkID: "test_char",
				MaxPeers:  16, // At limit
			},
			expectError: false,
		},
		{
			name: "discoveryPort too low",
			multiplayer: &MultiplayerConfig{
				Enabled:       true,
				NetworkID:     "test_char",
				DiscoveryPort: 80, // Below 1024
			},
			expectError: true,
			errorText:   "discoveryPort must be >= 1024",
		},
		{
			name: "discoveryPort too high",
			multiplayer: &MultiplayerConfig{
				Enabled:       true,
				NetworkID:     "test_char",
				DiscoveryPort: 70000, // Above 65535
			},
			expectError: true,
			errorText:   "discoveryPort cannot exceed 65535",
		},
		{
			name: "valid discoveryPort at boundaries",
			multiplayer: &MultiplayerConfig{
				Enabled:       true,
				NetworkID:     "test_char",
				DiscoveryPort: 1024, // At lower boundary
			},
			expectError: false,
		},
		{
			name: "valid high discoveryPort",
			multiplayer: &MultiplayerConfig{
				Enabled:       true,
				NetworkID:     "test_char",
				DiscoveryPort: 65535, // At upper boundary
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := getValidCharacterCard()
			card.Multiplayer = tt.multiplayer

			err := card.Validate()

			if tt.expectError {
				if err == nil {
					t.Error("expected validation error, but got none")
				} else if tt.errorText != "" && !containsSubstring(err.Error(), tt.errorText) {
					t.Errorf("expected error to contain '%s', got: %v", tt.errorText, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no validation error, got: %v", err)
				}
			}
		})
	}
}

// TestHasMultiplayer tests the HasMultiplayer helper method
func TestHasMultiplayer(t *testing.T) {
	tests := []struct {
		name        string
		multiplayer *MultiplayerConfig
		expected    bool
	}{
		{
			name:        "nil multiplayer config",
			multiplayer: nil,
			expected:    false,
		},
		{
			name: "disabled multiplayer",
			multiplayer: &MultiplayerConfig{
				Enabled: false,
			},
			expected: false,
		},
		{
			name: "enabled multiplayer",
			multiplayer: &MultiplayerConfig{
				Enabled:   true,
				NetworkID: "test_char",
			},
			expected: true,
		},
		{
			name: "enabled multiplayer with all features",
			multiplayer: &MultiplayerConfig{
				Enabled:       true,
				BotCapable:    true,
				NetworkID:     "advanced_bot",
				MaxPeers:      12,
				DiscoveryPort: 9090,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := getValidCharacterCard()
			card.Multiplayer = tt.multiplayer

			result := card.HasMultiplayer()

			if result != tt.expected {
				t.Errorf("HasMultiplayer() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestMultiplayerConfigDefaults tests that the validation works with default values
func TestMultiplayerConfigDefaults(t *testing.T) {
	tests := []struct {
		name        string
		multiplayer *MultiplayerConfig
		expectError bool
	}{
		{
			name: "minimal valid config",
			multiplayer: &MultiplayerConfig{
				Enabled:   true,
				NetworkID: "minimal_char",
				// Default values for MaxPeers (0) and DiscoveryPort (0) should be valid
			},
			expectError: false,
		},
		{
			name: "bot capable config",
			multiplayer: &MultiplayerConfig{
				Enabled:    true,
				BotCapable: true,
				NetworkID:  "bot_char",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := getValidCharacterCard()
			card.Multiplayer = tt.multiplayer

			err := card.Validate()

			if tt.expectError {
				if err == nil {
					t.Error("expected validation error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("expected no validation error, got: %v", err)
				}
			}
		})
	}
}
