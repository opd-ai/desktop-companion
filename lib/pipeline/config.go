package pipeline

// config.go provides configuration management for the asset generation pipeline.
// This implements the CharacterConfig, GIFConfig, and related structures outlined
// in GIF_PLAN.md, following the project's JSON-first configuration philosophy.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CharacterConfig defines complete character processing configuration.
type CharacterConfig struct {
	Character  *CharacterRequest  `json:"character"`
	States     []string           `json:"states"`     // Required animation states
	GIFConfig  *ExtendedGIFConfig `json:"gif_config"` // GIF generation settings
	Validation *ValidationConfig  `json:"validation"` // Quality requirements
	Deployment *DeploymentConfig  `json:"deployment"` // Output configuration
}

// CharacterRequest defines character generation parameters.
type CharacterRequest struct {
	Archetype    string            `json:"archetype"`     // e.g., "romance_tsundere"
	Description  string            `json:"description"`   // Character description
	Style        string            `json:"style"`         // Art style (pixel, anime, etc.)
	Traits       map[string]string `json:"traits"`        // Visual traits
	OutputConfig *OutputConfig     `json:"output_config"` // Size, format settings
}

// OutputConfig specifies image output parameters.
type OutputConfig struct {
	Width      int    `json:"width"`      // Target width in pixels
	Height     int    `json:"height"`     // Target height in pixels
	Format     string `json:"format"`     // Output format (png, jpg)
	Background string `json:"background"` // Background color/transparency
}

// ExtendedGIFConfig specifies comprehensive GIF output parameters.
// This extends the basic GIFConfig from deployer.go with additional generation settings.
type ExtendedGIFConfig struct {
	GIFConfig           // Embed basic GIF config (FrameCount, MaxFileSize, Transparency)
	Width        int    `json:"width"`        // Target width (64-256px)
	Height       int    `json:"height"`       // Target height (64-256px)
	FrameRate    int    `json:"frame_rate"`   // 10-15 FPS
	Colors       int    `json:"colors"`       // Indexed color count (256 max)
	Optimization string `json:"optimization"` // "size" or "quality"
}

// ValidationConfig defines quality requirements.
type ValidationConfig struct {
	MaxFileSize          int      `json:"max_file_size"`         // Maximum file size in bytes
	MinFrameRate         int      `json:"min_frame_rate"`        // Minimum frame rate
	RequiredStates       []string `json:"required_states"`       // Core animation states
	StyleConsistency     bool     `json:"style_consistency"`     // Cross-state consistency check
	ArchetypeCompliance  bool     `json:"archetype_compliance"`  // Personality accuracy check
	TransparencyRequired bool     `json:"transparency_required"` // Transparency validation
}

// DeploymentConfig specifies output and deployment settings.
type DeploymentConfig struct {
	OutputDir            string `json:"output_dir"`             // Target directory
	BackupExisting       bool   `json:"backup_existing"`        // Backup existing assets
	UpdateCharacterJSON  bool   `json:"update_character_json"`  // Update character.json files
	ValidateBeforeDeploy bool   `json:"validate_before_deploy"` // Validate before deployment
}

// ComfyUIConfig holds ComfyUI server configuration.
type ComfyUIConfig struct {
	ServerURL     string        `json:"server_url"`     // "http://localhost:8188"
	APIKey        string        `json:"api_key"`        // Optional authentication
	Timeout       time.Duration `json:"timeout"`        // Request timeout
	RetryAttempts int           `json:"retry_attempts"` // Failed request retries
	QueueLimit    int           `json:"queue_limit"`    // Max concurrent jobs
}

// WorkflowConfig defines workflow template configuration.
type WorkflowConfig struct {
	TemplatesPath string                 `json:"templates_path"` // Workflow JSON files
	Models        map[string]ModelConfig `json:"models"`         // Model configurations
	Styles        map[string]StyleConfig `json:"styles"`         // Style presets
	Quality       QualityConfig          `json:"quality"`        // Output quality settings
}

// ModelConfig specifies AI model parameters.
type ModelConfig struct {
	Name       string                 `json:"name"`       // Model name/path
	Type       string                 `json:"type"`       // Model type (checkpoint, lora, etc.)
	Parameters map[string]interface{} `json:"parameters"` // Model-specific parameters
}

// StyleConfig defines art style presets.
type StyleConfig struct {
	Name        string                 `json:"name"`        // Style name
	Description string                 `json:"description"` // Style description
	Prompts     StylePrompts           `json:"prompts"`     // Style-specific prompts
	Parameters  map[string]interface{} `json:"parameters"`  // Style parameters
}

// StylePrompts contains positive and negative prompts for a style.
type StylePrompts struct {
	Positive string `json:"positive"` // Positive style prompts
	Negative string `json:"negative"` // Negative style prompts
}

// QualityConfig defines output quality settings.
type QualityConfig struct {
	Steps     int     `json:"steps"`     // Generation steps
	CFGScale  float64 `json:"cfg_scale"` // CFG scale
	Sampler   string  `json:"sampler"`   // Sampling method
	Scheduler string  `json:"scheduler"` // Scheduler type
	Seed      int64   `json:"seed"`      // Random seed (-1 for random)
}

// PipelineConfig is the root configuration structure.
type PipelineConfig struct {
	ComfyUI    ComfyUIConfig    `json:"comfyui"`
	Workflow   WorkflowConfig   `json:"workflow"`
	Generation GenerationConfig `json:"generation"`
	Validation ValidationConfig `json:"validation"`
	Deployment DeploymentConfig `json:"deployment"`
}

// GenerationConfig defines default generation settings.
type GenerationConfig struct {
	DefaultStyle      string        `json:"default_style"`      // Default art style
	BaseResolution    [2]int        `json:"base_resolution"`    // [width, height]
	FrameCount        int           `json:"frame_count"`        // Default frame count
	AnimationDuration time.Duration `json:"animation_duration"` // Default animation length
	ConcurrentJobs    int           `json:"concurrent_jobs"`    // Parallel processing limit
	TempDir           string        `json:"temp_dir"`           // Temporary file directory
}

// ArchetypeMapping defines character archetype to prompt mappings.
type ArchetypeMapping struct {
	Archetype    string                      `json:"archetype"`     // Character archetype name
	BasePrompts  ArchetypePrompts            `json:"base_prompts"`  // Base character prompts
	StatePrompts map[string]ArchetypePrompts `json:"state_prompts"` // State-specific prompts
	Traits       map[string]string           `json:"traits"`        // Visual trait mappings
}

// ArchetypePrompts contains prompts for character generation.
type ArchetypePrompts struct {
	Positive  string   `json:"positive"`  // Positive prompts
	Negative  string   `json:"negative"`  // Negative prompts
	Keywords  []string `json:"keywords"`  // Additional keywords
	Modifiers []string `json:"modifiers"` // Prompt modifiers
}

// DefaultPipelineConfig returns a conservative default configuration.
func DefaultPipelineConfig() *PipelineConfig {
	return &PipelineConfig{
		ComfyUI: ComfyUIConfig{
			ServerURL:     "http://localhost:8188",
			Timeout:       30 * time.Second,
			RetryAttempts: 3,
			QueueLimit:    4,
		},
		Workflow: WorkflowConfig{
			TemplatesPath: "templates/workflows",
			Models:        make(map[string]ModelConfig),
			Styles:        defaultStyles(),
			Quality: QualityConfig{
				Steps:     20,
				CFGScale:  7.0,
				Sampler:   "euler_a",
				Scheduler: "normal",
				Seed:      -1,
			},
		},
		Generation: GenerationConfig{
			DefaultStyle:      "pixel_art",
			BaseResolution:    [2]int{128, 128},
			FrameCount:        6,
			AnimationDuration: 1 * time.Second,
			ConcurrentJobs:    2,
			TempDir:           "temp/generation",
		},
		Validation: ValidationConfig{
			MaxFileSize:          500000, // 500KB
			MinFrameRate:         10,
			RequiredStates:       []string{"idle", "talking", "happy", "sad"},
			StyleConsistency:     true,
			ArchetypeCompliance:  true,
			TransparencyRequired: true,
		},
		Deployment: DeploymentConfig{
			OutputDir:            "assets/characters",
			BackupExisting:       true,
			UpdateCharacterJSON:  true,
			ValidateBeforeDeploy: true,
		},
	}
}

// defaultStyles returns default art style configurations.
func defaultStyles() map[string]StyleConfig {
	return map[string]StyleConfig{
		"pixel_art": {
			Name:        "Pixel Art",
			Description: "8-bit pixel art style",
			Prompts: StylePrompts{
				Positive: "pixel art, 8bit, retro game style, clean pixels, sharp edges",
				Negative: "blurry, anti-aliased, high resolution, realistic",
			},
			Parameters: map[string]interface{}{
				"scale_factor": 1,
				"pixel_size":   1,
			},
		},
		"anime": {
			Name:        "Anime",
			Description: "Anime/manga art style",
			Prompts: StylePrompts{
				Positive: "anime style, manga, cel shading, clean lines, vibrant colors",
				Negative: "realistic, photograph, 3d render, western cartoon",
			},
			Parameters: map[string]interface{}{
				"saturation": 1.2,
				"contrast":   1.1,
			},
		},
		"chibi": {
			Name:        "Chibi",
			Description: "Cute chibi character style",
			Prompts: StylePrompts{
				Positive: "chibi, cute, kawaii, big eyes, small body, adorable, soft colors",
				Negative: "realistic proportions, mature, serious, dark colors",
			},
			Parameters: map[string]interface{}{
				"head_ratio": 0.4,
				"eye_size":   1.5,
			},
		},
	}
}

// LoadConfig loads pipeline configuration from a JSON file.
func LoadConfig(path string) (*PipelineConfig, error) {
	if path == "" {
		return nil, errors.New("config path required")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var config PipelineConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parse config JSON: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

// SaveConfig saves pipeline configuration to a JSON file.
func SaveConfig(config *PipelineConfig, path string) error {
	if config == nil {
		return errors.New("config is nil")
	}
	if path == "" {
		return errors.New("config path required")
	}

	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// Create directory if it doesn't exist
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create config directory: %w", err)
		}
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config JSON: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}

// Validate ensures configuration values are sensible.
func (c *PipelineConfig) Validate() error {
	// Validate ComfyUI config
	if c.ComfyUI.ServerURL == "" {
		return errors.New("comfyui server URL required")
	}
	if c.ComfyUI.Timeout <= 0 {
		return errors.New("comfyui timeout must be positive")
	}
	if c.ComfyUI.RetryAttempts < 0 {
		return errors.New("comfyui retry attempts cannot be negative")
	}

	// Validate generation config
	if c.Generation.BaseResolution[0] <= 0 || c.Generation.BaseResolution[1] <= 0 {
		return errors.New("base resolution must be positive")
	}
	if c.Generation.FrameCount < 4 || c.Generation.FrameCount > 8 {
		return errors.New("frame count must be between 4 and 8")
	}
	if c.Generation.ConcurrentJobs <= 0 {
		return errors.New("concurrent jobs must be positive")
	}

	// Validate validation config
	if c.Validation.MaxFileSize <= 0 {
		return errors.New("max file size must be positive")
	}
	if c.Validation.MinFrameRate <= 0 {
		return errors.New("min frame rate must be positive")
	}
	if len(c.Validation.RequiredStates) == 0 {
		return errors.New("required states cannot be empty")
	}

	// Validate deployment config
	if c.Deployment.OutputDir == "" {
		return errors.New("output directory required")
	}

	return nil
}

// LoadArchetypeMapping loads character archetype mappings from a JSON file.
func LoadArchetypeMapping(path string) ([]ArchetypeMapping, error) {
	if path == "" {
		return nil, errors.New("archetype mapping path required")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read archetype mapping file: %w", err)
	}

	var mappings []ArchetypeMapping
	if err := json.Unmarshal(data, &mappings); err != nil {
		return nil, fmt.Errorf("parse archetype mapping JSON: %w", err)
	}

	return mappings, nil
}

// GetArchetypeStates returns the required animation states for a character archetype.
func GetArchetypeStates(archetype string) []string {
	baseStates := []string{"idle", "talking", "happy", "sad"}

	switch archetype {
	case "romance_tsundere", "romance_flirty", "romance_slowburn", "romance_supportive":
		return append(baseStates, "shy", "flirty", "loving", "jealous")
	case "challenge", "hard":
		return append(baseStates, "angry", "frustrated", "determined")
	case "easy", "supportive":
		return append(baseStates, "encouraging", "cheerful", "caring")
	case "specialist":
		return append(baseStates, "focused", "excited", "proud")
	default:
		return baseStates
	}
}

// DefaultCharacterConfig creates a default character configuration for an archetype.
func DefaultCharacterConfig(archetype string) *CharacterConfig {
	states := GetArchetypeStates(archetype)

	return &CharacterConfig{
		Character: &CharacterRequest{
			Archetype:   archetype,
			Description: fmt.Sprintf("A %s character for the desktop companion", archetype),
			Style:       "pixel_art",
			Traits:      make(map[string]string),
			OutputConfig: &OutputConfig{
				Width:      128,
				Height:     128,
				Format:     "png",
				Background: "transparent",
			},
		},
		States: states,
		GIFConfig: &ExtendedGIFConfig{
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
		},
		Validation: &ValidationConfig{
			MaxFileSize:          500000,
			MinFrameRate:         10,
			RequiredStates:       states,
			StyleConsistency:     true,
			ArchetypeCompliance:  true,
			TransparencyRequired: true,
		},
		Deployment: &DeploymentConfig{
			OutputDir:            fmt.Sprintf("assets/characters/%s", archetype),
			BackupExisting:       true,
			UpdateCharacterJSON:  true,
			ValidateBeforeDeploy: true,
		},
	}
}
