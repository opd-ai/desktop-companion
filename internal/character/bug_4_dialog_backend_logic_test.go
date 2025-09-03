package character

import (
	"testing"

	"github.com/opd-ai/desktop-companion/internal/dialog"
)

// TestBug4HasDialogBackendLogic tests the strict logic in HasDialogBackend method
// This test reproduces the potential user confusion from overly strict backend checking
func TestBug4HasDialogBackendLogic(t *testing.T) {
	t.Log("Testing Bug #4: HasDialogBackend Logic Dependency")

	// Test 1: Character with dialog backend config but disabled
	t.Run("DisabledDialogBackend", func(t *testing.T) {
		card := &CharacterCard{
			Name: "TestCharacter",
			DialogBackend: &dialog.DialogBackendConfig{
				DefaultBackend: "test",
				Enabled:        false, // Explicitly disabled
			},
		}

		// Test current behavior
		hasBackend := card.HasDialogBackend()
		if hasBackend {
			t.Error("HasDialogBackend() should return false for disabled backend")
		}

		// The confusion: user configured a backend but HasDialogBackend() returns false
		// This might be confusing because they DID configure a backend, it's just disabled
		t.Log("Current behavior: HasDialogBackend() returns false for disabled backend")
		t.Log("Potential confusion: User configured backend but method says 'no backend'")

		// What might be more clear: separate methods for different concepts
		hasConfig := card.DialogBackend != nil
		isEnabled := hasConfig && card.DialogBackend.Enabled

		t.Logf("More granular info: HasConfig=%v, IsEnabled=%v", hasConfig, isEnabled)
		t.Log("This would allow better user feedback and decision making")
	})

	// Test 2: Character with no dialog backend at all
	t.Run("NoDialogBackend", func(t *testing.T) {
		card := &CharacterCard{
			Name: "BasicCharacter",
			// No DialogBackend field
		}

		hasBackend := card.HasDialogBackend()
		if hasBackend {
			t.Error("HasDialogBackend() should return false for no backend")
		}

		// This case is clear - no backend configured at all
		t.Log("Clear case: No backend configured, HasDialogBackend() returns false")
	})

	// Test 3: Character with enabled dialog backend
	t.Run("EnabledDialogBackend", func(t *testing.T) {
		card := &CharacterCard{
			Name: "AICharacter",
			DialogBackend: &dialog.DialogBackendConfig{
				DefaultBackend: "test",
				Enabled:        true,
			},
		}

		hasBackend := card.HasDialogBackend()
		if !hasBackend {
			t.Error("HasDialogBackend() should return true for enabled backend")
		}

		// This case is clear - backend configured and enabled
		t.Log("Clear case: Backend configured and enabled, HasDialogBackend() returns true")
	})

	t.Log("Analysis: The logic is technically correct but may benefit from more granular methods")
	t.Log("Suggestion: Add HasDialogBackendConfig() and IsDialogBackendEnabled() for clarity")
}

// TestBug4UserConfusionScenarios tests scenarios that might confuse users
func TestBug4UserConfusionScenarios(t *testing.T) {
	t.Log("Testing potential user confusion scenarios")

	// Scenario: User has AI character with backend configured but disabled
	t.Run("UserExpectsAIFeaturesButBackendDisabled", func(t *testing.T) {
		card := &CharacterCard{
			Name:        "SmartCharacter",
			Description: "An AI character with advanced features",
			Personality: &PersonalityConfig{
				Traits: map[string]float64{
					"intelligence": 0.9,
					"creativity":   0.8,
				},
			},
			DialogBackend: &dialog.DialogBackendConfig{
				DefaultBackend: "advanced_ai",
				Enabled:        false, // User disabled it or it's not working
			},
		}

		// User perspective: "This character looks AI-capable, why no chat?"
		hasRomanceFeatures := card.HasRomanceFeatures()
		hasDialogBackend := card.HasDialogBackend()

		t.Logf("Character has romance features: %v", hasRomanceFeatures)
		t.Logf("Character has dialog backend: %v", hasDialogBackend)

		// The issue: Character has AI features but HasDialogBackend() says no
		// With current logic, chatbot interface won't be created
		// User doesn't understand why their AI character doesn't have chat

		if hasRomanceFeatures && !hasDialogBackend {
			t.Log("CONFUSION: Character has AI features but no dialog backend available")
			t.Log("User might not understand why chat is unavailable")
			t.Log("Better approach: Show chat option with explanation about disabled backend")
		}
	})

	t.Log("Conclusion: Current logic is strict but could benefit from better user communication")
}
