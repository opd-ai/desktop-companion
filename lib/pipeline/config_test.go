package pipeline

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultPipelineConfig(t *testing.T) {
	config := DefaultPipelineConfig()

	if err := config.Validate(); err != nil {
		t.Fatalf("default config validation failed: %v", err)
	}

	// Test ComfyUI defaults
	if config.ComfyUI.ServerURL != "http://localhost:8188" {
		t.Errorf("expected default server URL, got %s", config.ComfyUI.ServerURL)
	}
	if config.ComfyUI.Timeout != 30*time.Second {
		t.Errorf("expected 30s timeout, got %v", config.ComfyUI.Timeout)
	}

	// Test generation defaults
	if config.Generation.DefaultStyle != "pixel_art" {
		t.Errorf("expected pixel_art style, got %s", config.Generation.DefaultStyle)
	}
	if config.Generation.FrameCount != 6 {
		t.Errorf("expected 6 frames, got %d", config.Generation.FrameCount)
	}

	// Test validation defaults
	if config.Validation.MaxFileSize != 500000 {
		t.Errorf("expected 500KB max file size, got %d", config.Validation.MaxFileSize)
	}
	if len(config.Validation.RequiredStates) != 4 {
		t.Errorf("expected 4 required states, got %d", len(config.Validation.RequiredStates))
	}
}

func TestConfigSaveLoad(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Create and save config
	original := DefaultPipelineConfig()
	original.ComfyUI.ServerURL = "http://test:8188"
	original.Generation.DefaultStyle = "anime"

	if err := SaveConfig(original, configPath); err != nil {
		t.Fatalf("save config failed: %v", err)
	}

	// Load config
	loaded, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}

	// Verify values
	if loaded.ComfyUI.ServerURL != "http://test:8188" {
		t.Errorf("expected test server URL, got %s", loaded.ComfyUI.ServerURL)
	}
	if loaded.Generation.DefaultStyle != "anime" {
		t.Errorf("expected anime style, got %s", loaded.Generation.DefaultStyle)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *PipelineConfig
		wantErr bool
	}{
		{
			name:    "valid default config",
			config:  DefaultPipelineConfig(),
			wantErr: false,
		},
		{
			name: "empty server URL",
			config: func() *PipelineConfig {
				c := DefaultPipelineConfig()
				c.ComfyUI.ServerURL = ""
				return c
			}(),
			wantErr: true,
		},
		{
			name: "invalid timeout",
			config: func() *PipelineConfig {
				c := DefaultPipelineConfig()
				c.ComfyUI.Timeout = 0
				return c
			}(),
			wantErr: true,
		},
		{
			name: "invalid frame count",
			config: func() *PipelineConfig {
				c := DefaultPipelineConfig()
				c.Generation.FrameCount = 10
				return c
			}(),
			wantErr: true,
		},
		{
			name: "empty required states",
			config: func() *PipelineConfig {
				c := DefaultPipelineConfig()
				c.Validation.RequiredStates = []string{}
				return c
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetArchetypeStates(t *testing.T) {
	tests := []struct {
		archetype     string
		expectedMin   int
		shouldContain []string
	}{
		{
			archetype:     "default",
			expectedMin:   4,
			shouldContain: []string{"idle", "talking", "happy", "sad"},
		},
		{
			archetype:     "romance_tsundere",
			expectedMin:   8,
			shouldContain: []string{"idle", "talking", "happy", "sad", "shy", "flirty", "loving", "jealous"},
		},
		{
			archetype:     "challenge",
			expectedMin:   7,
			shouldContain: []string{"idle", "talking", "happy", "sad", "angry", "frustrated", "determined"},
		},
		{
			archetype:     "easy",
			expectedMin:   7,
			shouldContain: []string{"idle", "talking", "happy", "sad", "encouraging", "cheerful", "caring"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.archetype, func(t *testing.T) {
			states := GetArchetypeStates(tt.archetype)

			if len(states) < tt.expectedMin {
				t.Errorf("expected at least %d states, got %d", tt.expectedMin, len(states))
			}

			for _, expected := range tt.shouldContain {
				found := false
				for _, state := range states {
					if state == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected state %q not found in %v", expected, states)
				}
			}
		})
	}
}

func TestDefaultCharacterConfig(t *testing.T) {
	archetype := "romance_tsundere"
	config := DefaultCharacterConfig(archetype)

	if config.Character.Archetype != archetype {
		t.Errorf("expected archetype %s, got %s", archetype, config.Character.Archetype)
	}

	if config.Character.Style != "pixel_art" {
		t.Errorf("expected pixel_art style, got %s", config.Character.Style)
	}

	if config.GIFConfig.Width != 128 {
		t.Errorf("expected width 128, got %d", config.GIFConfig.Width)
	}

	if config.GIFConfig.FrameCount != 6 {
		t.Errorf("expected 6 frames, got %d", config.GIFConfig.FrameCount)
	}

	// Check romance-specific states
	expectedStates := []string{"shy", "flirty", "loving", "jealous"}
	for _, expected := range expectedStates {
		found := false
		for _, state := range config.States {
			if state == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected romance state %q not found in %v", expected, config.States)
		}
	}
}

func TestDefaultStyles(t *testing.T) {
	styles := defaultStyles()

	expectedStyles := []string{"pixel_art", "anime", "chibi"}
	for _, expected := range expectedStyles {
		style, exists := styles[expected]
		if !exists {
			t.Errorf("expected style %q not found", expected)
			continue
		}

		if style.Name == "" {
			t.Errorf("style %q has empty name", expected)
		}
		if style.Prompts.Positive == "" {
			t.Errorf("style %q has empty positive prompts", expected)
		}
		if style.Prompts.Negative == "" {
			t.Errorf("style %q has empty negative prompts", expected)
		}
	}
}

func TestJSONMarshaling(t *testing.T) {
	config := DefaultPipelineConfig()

	// Test marshaling
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("marshal config failed: %v", err)
	}

	// Test unmarshaling
	var unmarshaled PipelineConfig
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("unmarshal config failed: %v", err)
	}

	// Verify some key values
	if unmarshaled.ComfyUI.ServerURL != config.ComfyUI.ServerURL {
		t.Errorf("server URL mismatch after marshal/unmarshal")
	}
	if unmarshaled.Generation.DefaultStyle != config.Generation.DefaultStyle {
		t.Errorf("default style mismatch after marshal/unmarshal")
	}
}

func TestLoadConfigNonexistent(t *testing.T) {
	_, err := LoadConfig("nonexistent.json")
	if err == nil {
		t.Error("expected error for nonexistent config file")
	}
}

func TestSaveConfigInvalidPath(t *testing.T) {
	config := DefaultPipelineConfig()
	err := SaveConfig(config, "")
	if err == nil {
		t.Error("expected error for empty config path")
	}
}

func TestExtendedGIFConfig(t *testing.T) {
	config := &ExtendedGIFConfig{
		GIFConfig: GIFConfig{
			FrameCount:   6,
			Transparency: true,
			MaxFileSize:  500000,
		},
		Width:        128,
		Height:       128,
		FrameRate:    12,
		Colors:       256,
		Optimization: "size",
	}

	// Test basic GIF config fields
	if config.FrameCount != 6 {
		t.Errorf("expected 6 frames, got %d", config.FrameCount)
	}
	if !config.Transparency {
		t.Error("expected transparency enabled")
	}
	if config.MaxFileSize != 500000 {
		t.Errorf("expected 500KB max size, got %d", config.MaxFileSize)
	}

	// Test extended fields
	if config.Width != 128 {
		t.Errorf("expected width 128, got %d", config.Width)
	}
	if config.Height != 128 {
		t.Errorf("expected height 128, got %d", config.Height)
	}
	if config.FrameRate != 12 {
		t.Errorf("expected frame rate 12, got %d", config.FrameRate)
	}
	if config.Colors != 256 {
		t.Errorf("expected 256 colors, got %d", config.Colors)
	}
	if config.Optimization != "size" {
		t.Errorf("expected size optimization, got %s", config.Optimization)
	}
}
