package testing

import (
	"testing"

	"desktop-companion/internal/character"
)

// TestBug4CharacterNameValidation - Investigation shows this is NOT a bug
// The AUDIT.md misunderstood the validation logic
func TestBug4CharacterNameValidationInvestigation(t *testing.T) {
	t.Log("Investigating Bug #4: Character Name Length Validation")

	// The AUDIT.md claimed empty names pass validation, but this is incorrect
	t.Run("EmptyNameCorrectlyFails", func(t *testing.T) {
		card := &character.CharacterCard{
			Name:        "", // 0 characters
			Description: "Valid description",
			Animations: map[string]string{
				"idle":    "idle.gif",
				"talking": "talking.gif",
			},
			Dialogs: []character.Dialog{
				{
					Trigger:   "click",
					Responses: []string{"Hello!"},
					Animation: "talking",
					Cooldown:  5,
				},
			},
			Behavior: character.Behavior{
				IdleTimeout: 30,
				DefaultSize: 128,
			},
		}

		err := card.Validate()
		if err != nil {
			t.Logf("CORRECT: Empty name fails validation as expected: %v", err)
		} else {
			t.Error("UNEXPECTED: Empty name should fail validation")
		}
	})

	// Verify the validation works correctly for all edge cases
	t.Run("ValidationBoundaryTests", func(t *testing.T) {
		testCases := []struct {
			name        string
			input       string
			shouldPass  bool
			description string
		}{
			{"", "", false, "Empty name should fail"},
			{"A", "A", true, "Single character should pass"},
			{"Valid Name", "Valid Name", true, "Normal name should pass"},
			{"ExactlyFiftyCharactersLongNameHereForTesting12", "12345678901234567890123456789012345678901234567890", true, "50 characters should pass"},
			{"TooLong", "123456789012345678901234567890123456789012345678901", false, "51 characters should fail"},
		}

		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				card := &character.CharacterCard{
					Name:        tc.input,
					Description: "Valid description",
					Animations: map[string]string{
						"idle":    "idle.gif",
						"talking": "talking.gif",
					},
					Dialogs: []character.Dialog{
						{
							Trigger:   "click",
							Responses: []string{"Hello!"},
							Animation: "talking",
							Cooldown:  5,
						},
					},
					Behavior: character.Behavior{
						IdleTimeout: 30,
						DefaultSize: 128,
					},
				}

				err := card.Validate()
				if tc.shouldPass && err != nil {
					t.Errorf("Expected '%s' (len=%d) to pass but got error: %v", tc.input, len(tc.input), err)
				} else if !tc.shouldPass && err == nil {
					t.Errorf("Expected '%s' (len=%d) to fail but validation passed", tc.input, len(tc.input))
				} else {
					t.Logf("CORRECT: '%s' (len=%d) behaved as expected", tc.input, len(tc.input))
				}
			})
		}
	})

	t.Log("CONCLUSION: Bug #4 is NOT a real bug - validation logic is correct")
	t.Log("The AUDIT.md misunderstood that 'len(c.Name) == 0' returns an ERROR, not allows empty names")
}
