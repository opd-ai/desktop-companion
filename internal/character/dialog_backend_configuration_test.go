package character

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDialogBackendDefaultConfigurationMismatch reproduces the bug where
// most default character cards don't include dialog backend configuration,
// causing the AI-powered dialog features to remain unused
func TestDialogBackendDefaultConfigurationMismatch(t *testing.T) {
	// Test that characters that should demonstrate AI features actually have dialogBackend
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	// Define characters that should demonstrate the AI dialog system
	// (at minimum, common demonstration characters)
	testCases := []struct {
		name         string
		characterDir string
		expectDialog bool
		reason       string
	}{
		{
			name:         "default",
			characterDir: "assets/characters/default",
			expectDialog: true,
			reason:       "Default character should demonstrate all major features",
		},
		{
			name:         "normal",
			characterDir: "assets/characters/normal",
			expectDialog: true,
			reason:       "Normal character is a common archetype that users will try",
		},
		{
			name:         "markov_example",
			characterDir: "assets/characters/markov_example",
			expectDialog: true,
			reason:       "Character specifically named to demonstrate Markov chain dialogs",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cardPath := filepath.Join(projectRoot, tc.characterDir, "character.json")

			// Load and validate the character card
			card, err := LoadCard(cardPath)
			if err != nil {
				t.Fatalf("Failed to load character card %s: %v", tc.name, err)
			}

			// Check if character has dialog backend configuration
			hasDialogBackend := card.DialogBackend != nil

			if tc.expectDialog && !hasDialogBackend {
				t.Errorf("Character %s should have dialogBackend configuration (%s), but doesn't",
					tc.name, tc.reason)
			}

			// If it has dialog backend, verify it's properly configured
			if hasDialogBackend {
				if !card.DialogBackend.Enabled {
					t.Errorf("Character %s has dialogBackend but it's disabled", tc.name)
				}

				if card.DialogBackend.DefaultBackend == "" {
					t.Errorf("Character %s has dialogBackend but no defaultBackend specified", tc.name)
				}

				if card.DialogBackend.Backends == nil || len(card.DialogBackend.Backends) == 0 {
					t.Errorf("Character %s has dialogBackend but no backends configured", tc.name)
				}
			}
		})
	}
}

// findProjectRoot finds the project root by looking for go.mod
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}
