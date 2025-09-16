package character

import (
	"testing"
	"time"
)

func TestDefaultAssetGenerationConfig(t *testing.T) {
	config := DefaultAssetGenerationConfig()

	if config == nil {
		t.Fatal("DefaultAssetGenerationConfig returned nil")
	}

	// Test required fields
	if config.BasePrompt == "" {
		t.Error("BasePrompt should not be empty")
	}

	if config.GenerationSettings.Model == "" {
		t.Error("Model should not be empty")
	}

	if config.GenerationSettings.ArtStyle == "" {
		t.Error("ArtStyle should not be empty")
	}

	// Test animation mappings
	expectedStates := []string{"idle", "talking", "happy", "sad"}
	for _, state := range expectedStates {
		if _, exists := config.AnimationMappings[state]; !exists {
			t.Errorf("Missing required animation state: %s", state)
		}
	}

	// Test resolution
	res := config.GenerationSettings.Resolution
	if res.Width <= 0 || res.Height <= 0 {
		t.Error("Resolution dimensions must be positive")
	}

	// Test backup settings
	if !config.BackupSettings.Enabled {
		t.Error("Backup should be enabled by default")
	}
}

func TestValidateAssetGenerationConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *AssetGenerationConfig
		wantErr bool
	}{
		{
			name:    "nil config should pass",
			config:  nil,
			wantErr: false,
		},
		{
			name:    "default config should pass",
			config:  DefaultAssetGenerationConfig(),
			wantErr: false,
		},
		{
			name: "empty base prompt should fail",
			config: &AssetGenerationConfig{
				BasePrompt: "",
				GenerationSettings: GenerationSettings{
					Model:    "flux1d",
					ArtStyle: "anime",
					Resolution: ImageResolution{
						Width:  128,
						Height: 128,
					},
					QualitySettings: QualitySettings{
						Steps:    25,
						CFGScale: 7.0,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid model should fail",
			config: &AssetGenerationConfig{
				BasePrompt: "test prompt",
				GenerationSettings: GenerationSettings{
					Model:    "invalid_model",
					ArtStyle: "anime",
					Resolution: ImageResolution{
						Width:  128,
						Height: 128,
					},
					QualitySettings: QualitySettings{
						Steps:    25,
						CFGScale: 7.0,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid art style should fail",
			config: &AssetGenerationConfig{
				BasePrompt: "test prompt",
				GenerationSettings: GenerationSettings{
					Model:    "flux1d",
					ArtStyle: "invalid_style",
					Resolution: ImageResolution{
						Width:  128,
						Height: 128,
					},
					QualitySettings: QualitySettings{
						Steps:    25,
						CFGScale: 7.0,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid resolution should fail",
			config: &AssetGenerationConfig{
				BasePrompt: "test prompt",
				GenerationSettings: GenerationSettings{
					Model:    "flux1d",
					ArtStyle: "anime",
					Resolution: ImageResolution{
						Width:  32, // Too small
						Height: 128,
					},
					QualitySettings: QualitySettings{
						Steps:    25,
						CFGScale: 7.0,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid quality settings should fail",
			config: &AssetGenerationConfig{
				BasePrompt: "test prompt",
				GenerationSettings: GenerationSettings{
					Model:    "flux1d",
					ArtStyle: "anime",
					Resolution: ImageResolution{
						Width:  128,
						Height: 128,
					},
					QualitySettings: QualitySettings{
						Steps:    5, // Too few
						CFGScale: 7.0,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid animation settings should fail",
			config: &AssetGenerationConfig{
				BasePrompt: "test prompt",
				GenerationSettings: GenerationSettings{
					Model:    "flux1d",
					ArtStyle: "anime",
					Resolution: ImageResolution{
						Width:  128,
						Height: 128,
					},
					QualitySettings: QualitySettings{
						Steps:    25,
						CFGScale: 7.0,
					},
					AnimationSettings: AnimationSettings{
						FrameRate: 50, // Too high
						Duration:  2.0,
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAssetGenerationConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAssetGenerationConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAnimationMapping(t *testing.T) {
	mapping := AnimationMapping{
		PromptModifier:   "happy expression, smiling",
		NegativePrompt:   "sad, angry",
		StateDescription: "Happy state",
		FrameCount:       6,
	}

	if mapping.PromptModifier == "" {
		t.Error("PromptModifier should not be empty")
	}

	if mapping.FrameCount <= 0 {
		t.Error("FrameCount should be positive")
	}
}

func TestGenerationSettings(t *testing.T) {
	settings := GenerationSettings{
		Model:    "flux1d",
		ArtStyle: "anime",
		Resolution: ImageResolution{
			Width:  128,
			Height: 128,
		},
		QualitySettings: QualitySettings{
			Steps:     25,
			CFGScale:  7.0,
			Sampler:   "euler_a",
			Scheduler: "normal",
		},
		AnimationSettings: AnimationSettings{
			FrameRate:           12,
			Duration:            2.0,
			LoopType:            "seamless",
			Optimization:        "balanced",
			MaxFileSize:         500,
			TransparencyEnabled: true,
		},
	}

	// Test model validation
	validModels := []string{"sdxl", "flux1d", "flux1s"}
	found := false
	for _, model := range validModels {
		if settings.Model == model {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Model %s should be valid", settings.Model)
	}

	// Test art style validation
	validStyles := []string{"anime", "pixel_art", "realistic", "cartoon", "chibi"}
	found = false
	for _, style := range validStyles {
		if settings.ArtStyle == style {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("ArtStyle %s should be valid", settings.ArtStyle)
	}
}

func TestAssetMetadata(t *testing.T) {
	metadata := AssetMetadata{
		Version:     "1.0.0",
		GeneratedAt: time.Now(),
		GeneratedBy: "gif-generator v1.0.0",
		GenerationHistory: []GenerationRecord{
			{
				Timestamp: time.Now(),
				Success:   true,
				Duration:  5 * time.Minute,
			},
		},
		AssetHashes: map[string]string{
			"idle.gif": "sha256:abc123",
		},
	}

	if metadata.Version == "" {
		t.Error("Version should not be empty")
	}

	if metadata.GeneratedBy == "" {
		t.Error("GeneratedBy should not be empty")
	}

	if len(metadata.GenerationHistory) == 0 {
		t.Error("GenerationHistory should not be empty")
	}

	if len(metadata.AssetHashes) == 0 {
		t.Error("AssetHashes should not be empty")
	}
}

func TestBackupSettings(t *testing.T) {
	settings := BackupSettings{
		Enabled:         true,
		BackupPath:      "backups",
		MaxBackups:      5,
		CompressBackups: true,
	}

	if !settings.Enabled {
		t.Error("Backup should be enabled")
	}

	if settings.BackupPath == "" {
		t.Error("BackupPath should not be empty")
	}

	if settings.MaxBackups <= 0 {
		t.Error("MaxBackups should be positive")
	}
}

func TestControlNetSettings(t *testing.T) {
	settings := ControlNetSettings{
		Model:        "control_v11p_sd15_openpose",
		Strength:     0.8,
		StartStep:    0.0,
		EndStep:      1.0,
		Preprocessor: "openpose",
	}

	if settings.Model == "" {
		t.Error("Model should not be empty")
	}

	if settings.Strength < 0.0 || settings.Strength > 1.0 {
		t.Error("Strength should be between 0.0 and 1.0")
	}

	if settings.StartStep < 0.0 || settings.StartStep > 1.0 {
		t.Error("StartStep should be between 0.0 and 1.0")
	}

	if settings.EndStep < 0.0 || settings.EndStep > 1.0 {
		t.Error("EndStep should be between 0.0 and 1.0")
	}

	if settings.StartStep >= settings.EndStep {
		t.Error("StartStep should be less than EndStep")
	}
}
