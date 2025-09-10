package character

import (
	"testing"
)

// TestRomanceFeatureValidation tests the validation of romance-specific features
func TestRomanceFeatureValidation(t *testing.T) {
	tests := []struct {
		name    string
		card    CharacterCard
		wantErr bool
	}{
		{
			name: "valid personality config",
			card: CharacterCard{
				Name:        "Test Romance",
				Description: "Test romance character",
				Animations:  map[string]string{"idle": "idle.gif", "talking": "talking.gif"},
				Dialogs: []Dialog{
					{Trigger: "click", Responses: []string{"Hello"}, Animation: "talking"},
				},
				Behavior: Behavior{IdleTimeout: 30, DefaultSize: 128},
				Personality: &PersonalityConfig{
					Traits: map[string]float64{
						"shyness":     0.6,
						"romanticism": 0.8,
					},
					Compatibility: map[string]float64{
						"gift_appreciation":  1.5,
						"conversation_lover": 1.3,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid personality trait value too high",
			card: CharacterCard{
				Name:        "Test Romance",
				Description: "Test romance character",
				Animations:  map[string]string{"idle": "idle.gif", "talking": "talking.gif"},
				Dialogs: []Dialog{
					{Trigger: "click", Responses: []string{"Hello"}, Animation: "talking"},
				},
				Behavior: Behavior{IdleTimeout: 30, DefaultSize: 128},
				Personality: &PersonalityConfig{
					Traits: map[string]float64{
						"shyness": 1.5, // Invalid: > 1.0
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid personality trait value too low",
			card: CharacterCard{
				Name:        "Test Romance",
				Description: "Test romance character",
				Animations:  map[string]string{"idle": "idle.gif", "talking": "talking.gif"},
				Dialogs: []Dialog{
					{Trigger: "click", Responses: []string{"Hello"}, Animation: "talking"},
				},
				Behavior: Behavior{IdleTimeout: 30, DefaultSize: 128},
				Personality: &PersonalityConfig{
					Traits: map[string]float64{
						"shyness": -0.1, // Invalid: < 0.0
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid compatibility modifier too high",
			card: CharacterCard{
				Name:        "Test Romance",
				Description: "Test romance character",
				Animations:  map[string]string{"idle": "idle.gif", "talking": "talking.gif"},
				Dialogs: []Dialog{
					{Trigger: "click", Responses: []string{"Hello"}, Animation: "talking"},
				},
				Behavior: Behavior{IdleTimeout: 30, DefaultSize: 128},
				Personality: &PersonalityConfig{
					Compatibility: map[string]float64{
						"gift_appreciation": 6.0, // Invalid: > 5.0
					},
				},
			},
			wantErr: true,
		},
		{
			name: "valid romance dialog with requirements",
			card: CharacterCard{
				Name:        "Test Romance",
				Description: "Test romance character",
				Animations:  map[string]string{"idle": "idle.gif", "talking": "talking.gif", "romantic": "romantic.gif"},
				Dialogs: []Dialog{
					{Trigger: "click", Responses: []string{"Hello"}, Animation: "talking"},
				},
				Behavior: Behavior{IdleTimeout: 30, DefaultSize: 128},
				Stats: map[string]StatConfig{
					"affection": {Initial: 0, Max: 100, DegradationRate: 0.1, CriticalThreshold: 10},
				},
				RomanceDialogs: []DialogExtended{
					{
						Dialog: Dialog{
							Trigger:   "click",
							Responses: []string{"Hello sweetheart!"},
							Animation: "romantic",
						},
						Requirements: &RomanceRequirement{
							Stats: map[string]map[string]float64{
								"affection": {"min": 50},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid romance dialog - references undefined stat",
			card: CharacterCard{
				Name:        "Test Romance",
				Description: "Test romance character",
				Animations:  map[string]string{"idle": "idle.gif", "talking": "talking.gif", "romantic": "romantic.gif"},
				Dialogs: []Dialog{
					{Trigger: "click", Responses: []string{"Hello"}, Animation: "talking"},
				},
				Behavior: Behavior{IdleTimeout: 30, DefaultSize: 128},
				Stats: map[string]StatConfig{
					"affection": {Initial: 0, Max: 100, DegradationRate: 0.1, CriticalThreshold: 10},
				},
				RomanceDialogs: []DialogExtended{
					{
						Dialog: Dialog{
							Trigger:   "click",
							Responses: []string{"Hello sweetheart!"},
							Animation: "romantic",
						},
						Requirements: &RomanceRequirement{
							Stats: map[string]map[string]float64{
								"love": {"min": 50}, // Invalid: stat doesn't exist
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "character without romance features should validate normally",
			card: CharacterCard{
				Name:        "Test Normal",
				Description: "Test normal character",
				Animations:  map[string]string{"idle": "idle.gif", "talking": "talking.gif"},
				Dialogs: []Dialog{
					{Trigger: "click", Responses: []string{"Hello"}, Animation: "talking"},
				},
				Behavior: Behavior{IdleTimeout: 30, DefaultSize: 128},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.card.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("CharacterCard.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestHasRomanceFeatures tests the HasRomanceFeatures method
func TestHasRomanceFeatures(t *testing.T) {
	tests := []struct {
		name string
		card CharacterCard
		want bool
	}{
		{
			name: "character with personality has romance features",
			card: CharacterCard{
				Personality: &PersonalityConfig{
					Traits: map[string]float64{"shyness": 0.5},
				},
			},
			want: true,
		},
		{
			name: "character with romance dialogs has romance features",
			card: CharacterCard{
				RomanceDialogs: []DialogExtended{
					{Dialog: Dialog{Trigger: "click", Responses: []string{"Hello"}}},
				},
			},
			want: true,
		},
		{
			name: "character with romance events has romance features",
			card: CharacterCard{
				RomanceEvents: []RandomEventConfig{
					{Name: "test", Description: "test event", Probability: 0.1},
				},
			},
			want: true,
		},
		{
			name: "character without romance features",
			card: CharacterCard{},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.card.HasRomanceFeatures(); got != tt.want {
				t.Errorf("CharacterCard.HasRomanceFeatures() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGetPersonalityTrait tests the GetPersonalityTrait method
func TestGetPersonalityTrait(t *testing.T) {
	tests := []struct {
		name  string
		card  CharacterCard
		trait string
		want  float64
	}{
		{
			name: "existing trait returns correct value",
			card: CharacterCard{
				Personality: &PersonalityConfig{
					Traits: map[string]float64{"shyness": 0.7},
				},
			},
			trait: "shyness",
			want:  0.7,
		},
		{
			name: "non-existing trait returns default",
			card: CharacterCard{
				Personality: &PersonalityConfig{
					Traits: map[string]float64{"shyness": 0.7},
				},
			},
			trait: "confidence",
			want:  0.5, // Default value
		},
		{
			name:  "nil personality returns default",
			card:  CharacterCard{},
			trait: "shyness",
			want:  0.5, // Default value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.card.GetPersonalityTrait(tt.trait); got != tt.want {
				t.Errorf("CharacterCard.GetPersonalityTrait() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGetCompatibilityModifier tests the GetCompatibilityModifier method
func TestGetCompatibilityModifier(t *testing.T) {
	tests := []struct {
		name     string
		card     CharacterCard
		behavior string
		want     float64
	}{
		{
			name: "existing modifier returns correct value",
			card: CharacterCard{
				Personality: &PersonalityConfig{
					Compatibility: map[string]float64{"gift_appreciation": 1.5},
				},
			},
			behavior: "gift_appreciation",
			want:     1.5,
		},
		{
			name: "non-existing modifier returns default",
			card: CharacterCard{
				Personality: &PersonalityConfig{
					Compatibility: map[string]float64{"gift_appreciation": 1.5},
				},
			},
			behavior: "conversation_lover",
			want:     1.0, // Default value
		},
		{
			name:     "nil personality returns default",
			card:     CharacterCard{},
			behavior: "gift_appreciation",
			want:     1.0, // Default value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.card.GetCompatibilityModifier(tt.behavior); got != tt.want {
				t.Errorf("CharacterCard.GetCompatibilityModifier() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestRomanceCharacterLoading tests loading the actual romance character configuration
func TestRomanceCharacterLoading(t *testing.T) {
	// Test that our romance character config is valid
	card := CharacterCard{
		Name:        "Romance Companion",
		Description: "A virtual dating companion with romance mechanics",
		Animations: map[string]string{
			"idle":     "animations/idle.gif",
			"talking":  "animations/talking.gif",
			"blushing": "animations/blushing.gif",
		},
		Dialogs: []Dialog{
			{Trigger: "click", Responses: []string{"Hello"}, Animation: "talking"},
		},
		Behavior: Behavior{IdleTimeout: 30, DefaultSize: 128},
		Stats: map[string]StatConfig{
			"affection": {Initial: 0, Max: 100, DegradationRate: 0.1, CriticalThreshold: 10},
			"trust":     {Initial: 20, Max: 100, DegradationRate: 0.05, CriticalThreshold: 5},
		},
		Personality: &PersonalityConfig{
			Traits: map[string]float64{
				"shyness":     0.6,
				"romanticism": 0.8,
			},
			Compatibility: map[string]float64{
				"gift_appreciation":  1.5,
				"conversation_lover": 1.3,
			},
		},
		RomanceDialogs: []DialogExtended{
			{
				Dialog: Dialog{
					Trigger:   "click",
					Responses: []string{"Hi sweetheart!"},
					Animation: "talking",
				},
				Requirements: &RomanceRequirement{
					Stats: map[string]map[string]float64{
						"affection": {"min": 50},
					},
				},
			},
		},
	}

	err := card.Validate()
	if err != nil {
		t.Errorf("Romance character validation failed: %v", err)
	}

	// Test romance feature detection
	if !card.HasRomanceFeatures() {
		t.Error("Romance character should have romance features")
	}

	// Test personality trait access
	shyness := card.GetPersonalityTrait("shyness")
	if shyness != 0.6 {
		t.Errorf("Expected shyness 0.6, got %f", shyness)
	}

	// Test compatibility modifier access
	giftMod := card.GetCompatibilityModifier("gift_appreciation")
	if giftMod != 1.5 {
		t.Errorf("Expected gift appreciation modifier 1.5, got %f", giftMod)
	}
}
