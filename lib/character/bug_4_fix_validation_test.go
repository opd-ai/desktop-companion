// Bug #4 Fix Validation Test
// Tests the new granular dialog backend status methods

package character

import (
	"testing"

	"github.com/opd-ai/desktop-companion/internal/dialog"
)

func TestBug4FixValidation(t *testing.T) {
	t.Log("Testing Bug #4 Fix: Granular Dialog Backend Status Methods")

	// Test case 1: No dialog backend configured
	t.Run("NoBackendConfigured", func(t *testing.T) {
		card := &CharacterCard{
			Name:        "TestCharacter",
			Description: "A test character",
			// No DialogBackend field
		}

		// Test new methods
		hasConfig := card.HasDialogBackendConfig()
		isEnabled := card.IsDialogBackendEnabled()
		hasBackend := card.HasDialogBackend()

		// Validate
		if hasConfig {
			t.Error("HasDialogBackendConfig() should return false for no backend")
		}
		if isEnabled {
			t.Error("IsDialogBackendEnabled() should return false for no backend")
		}
		if hasBackend {
			t.Error("HasDialogBackend() should return false for no backend")
		}

		// Test status method
		configExists, enabled, summary := card.GetDialogBackendStatus()
		if configExists {
			t.Error("GetDialogBackendStatus() config should be false for no backend")
		}
		if enabled {
			t.Error("GetDialogBackendStatus() enabled should be false for no backend")
		}
		expectedSummary := "No dialog backend configured"
		if summary != expectedSummary {
			t.Errorf("GetDialogBackendStatus() summary = %q, expected %q", summary, expectedSummary)
		}

		t.Log("✓ No backend configured: All methods return appropriate values")
	})

	// Test case 2: Dialog backend configured but disabled
	t.Run("BackendConfiguredButDisabled", func(t *testing.T) {
		card := &CharacterCard{
			Name:        "TestCharacter",
			Description: "A test character",
			DialogBackend: &dialog.DialogBackendConfig{
				DefaultBackend: "markov",
				Enabled:        false, // Disabled
			},
		}

		// Test new methods
		hasConfig := card.HasDialogBackendConfig()
		isEnabled := card.IsDialogBackendEnabled()
		hasBackend := card.HasDialogBackend()

		// Validate
		if !hasConfig {
			t.Error("HasDialogBackendConfig() should return true for configured backend")
		}
		if isEnabled {
			t.Error("IsDialogBackendEnabled() should return false for disabled backend")
		}
		if hasBackend {
			t.Error("HasDialogBackend() should return false for disabled backend")
		}

		// Test status method
		configExists, enabled, summary := card.GetDialogBackendStatus()
		if !configExists {
			t.Error("GetDialogBackendStatus() config should be true for configured backend")
		}
		if enabled {
			t.Error("GetDialogBackendStatus() enabled should be false for disabled backend")
		}
		expectedSummary := "Dialog backend configured but disabled"
		if summary != expectedSummary {
			t.Errorf("GetDialogBackendStatus() summary = %q, expected %q", summary, expectedSummary)
		}

		t.Log("✓ Backend configured but disabled: Methods distinguish config vs enabled state")
	})

	// Test case 3: Dialog backend configured and enabled
	t.Run("BackendConfiguredAndEnabled", func(t *testing.T) {
		card := &CharacterCard{
			Name:        "TestCharacter",
			Description: "A test character",
			DialogBackend: &dialog.DialogBackendConfig{
				DefaultBackend: "markov",
				Enabled:        true, // Enabled
			},
		}

		// Test new methods
		hasConfig := card.HasDialogBackendConfig()
		isEnabled := card.IsDialogBackendEnabled()
		hasBackend := card.HasDialogBackend()

		// Validate
		if !hasConfig {
			t.Error("HasDialogBackendConfig() should return true for configured backend")
		}
		if !isEnabled {
			t.Error("IsDialogBackendEnabled() should return true for enabled backend")
		}
		if !hasBackend {
			t.Error("HasDialogBackend() should return true for enabled backend")
		}

		// Test status method
		configExists, enabled, summary := card.GetDialogBackendStatus()
		if !configExists {
			t.Error("GetDialogBackendStatus() config should be true for configured backend")
		}
		if !enabled {
			t.Error("GetDialogBackendStatus() enabled should be true for enabled backend")
		}
		expectedSummary := "Dialog backend configured and enabled"
		if summary != expectedSummary {
			t.Errorf("GetDialogBackendStatus() summary = %q, expected %q", summary, expectedSummary)
		}

		t.Log("✓ Backend configured and enabled: All methods return true/enabled")
	})

	t.Log("Bug #4 Fix Complete: New granular methods provide clear distinction between configuration and enabled state")
}

// TestBug4UserExperienceImprovement tests how the fix improves user experience
func TestBug4UserExperienceImprovement(t *testing.T) {
	t.Log("Testing Bug #4 User Experience Improvement")

	// Scenario: User configures dialog backend but disables it temporarily
	card := &CharacterCard{
		Name:        "TestCharacter",
		Description: "A test character",
		DialogBackend: &dialog.DialogBackendConfig{
			DefaultBackend: "markov",
			Enabled:        false, // User disabled temporarily
		},
	}

	// Before the fix: UI would say "no chat capabilities"
	// After the fix: UI can distinguish and provide helpful message

	hasConfig := card.HasDialogBackendConfig()
	isEnabled := card.IsDialogBackendEnabled()

	// User has configured backend (so chat is conceptually available)
	if !hasConfig {
		t.Error("Should detect that user has configured dialog backend")
	}

	// But it's currently disabled
	if isEnabled {
		t.Error("Should detect that backend is currently disabled")
	}

	// UI can now provide specific guidance:
	// "Chat feature disabled. Enable it in configuration to use AI chat."
	// Instead of generic: "Chat not available for this character."

	t.Log("✓ User Experience Improved: UI can provide specific guidance about disabled vs unconfigured backends")
}
