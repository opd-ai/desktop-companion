package character

import (
	"fmt"
	"time"
)

// AssetGenerationConfig defines the configuration for AI-generated character assets
// This integrates with the gif-generator tool and ComfyUI pipeline to automatically
// create GIF animations from text prompts optimized for SDXL/Flux.1d models.
type AssetGenerationConfig struct {
	// BasePrompt is the comprehensive description optimized for SDXL/Flux.1d
	// that describes the character's visual appearance, art style, and key characteristics
	BasePrompt string `json:"basePrompt"`

	// AnimationMappings defines state-specific prompt modifications for each animation
	AnimationMappings map[string]AnimationMapping `json:"animationMappings"`

	// GenerationSettings contains technical parameters for the AI image generation process
	GenerationSettings GenerationSettings `json:"generationSettings"`

	// AssetMetadata tracks version and generation history for the automated pipeline
	AssetMetadata AssetMetadata `json:"assetMetadata,omitempty"`

	// BackupSettings controls asset backup behavior before regeneration
	BackupSettings BackupSettings `json:"backupSettings,omitempty"`
}

// AnimationMapping defines how to modify the base prompt for specific animation states
type AnimationMapping struct {
	// PromptModifier is the text to append/modify the base prompt for this animation state
	PromptModifier string `json:"promptModifier"`

	// NegativePrompt defines what to avoid in generation for this state
	NegativePrompt string `json:"negativePrompt,omitempty"`

	// StateDescription provides human-readable description of the animation state
	StateDescription string `json:"stateDescription,omitempty"`

	// FrameCount specifies number of frames for this animation (default: 4-8)
	FrameCount int `json:"frameCount,omitempty"`

	// CustomSettings allows per-animation override of generation settings
	CustomSettings *GenerationSettings `json:"customSettings,omitempty"`
}

// GenerationSettings contains technical parameters for AI image generation
type GenerationSettings struct {
	// Model specifies the AI model to use ("sdxl", "flux1d", "flux1s")
	Model string `json:"model"`

	// ArtStyle defines the artistic style ("anime", "pixel_art", "realistic", "cartoon", "chibi")
	ArtStyle string `json:"artStyle"`

	// Resolution defines output image dimensions
	Resolution ImageResolution `json:"resolution"`

	// QualitySettings controls generation quality parameters
	QualitySettings QualitySettings `json:"qualitySettings"`

	// AnimationSettings controls GIF-specific parameters
	AnimationSettings AnimationSettings `json:"animationSettings"`

	// ComfyUISettings contains ComfyUI-specific workflow parameters
	ComfyUISettings ComfyUISettings `json:"comfyUISettings,omitempty"`
}

// ImageResolution defines the target image dimensions
type ImageResolution struct {
	Width  int `json:"width"`  // Target width (64-512px)
	Height int `json:"height"` // Target height (64-512px)
}

// QualitySettings controls generation quality parameters
type QualitySettings struct {
	// Steps defines number of diffusion steps (10-100, higher = better quality)
	Steps int `json:"steps"`

	// CFGScale controls adherence to prompt (1.0-20.0, typical: 7.0-12.0)
	CFGScale float64 `json:"cfgScale"`

	// Seed for reproducible generation (-1 = random)
	Seed int64 `json:"seed,omitempty"`

	// Sampler defines the sampling method ("euler_a", "dpmpp_2m", "heun", etc.)
	Sampler string `json:"sampler,omitempty"`

	// Scheduler defines the noise schedule ("normal", "karras", "exponential")
	Scheduler string `json:"scheduler,omitempty"`
}

// AnimationSettings controls GIF-specific parameters
type AnimationSettings struct {
	// FrameRate defines FPS for GIF animation (5-30)
	FrameRate int `json:"frameRate"`

	// Duration defines animation duration in seconds (1-5)
	Duration float64 `json:"duration"`

	// LoopType defines how animation loops ("seamless", "bounce", "linear")
	LoopType string `json:"loopType"`

	// Optimization controls GIF file size optimization ("size", "quality", "balanced")
	Optimization string `json:"optimization"`

	// MaxFileSize defines maximum GIF file size in KB (default: 500)
	MaxFileSize int `json:"maxFileSize,omitempty"`

	// TransparencyEnabled enables alpha channel support
	TransparencyEnabled bool `json:"transparencyEnabled"`

	// ColorPalette defines color reduction strategy ("adaptive", "web", "grayscale")
	ColorPalette string `json:"colorPalette,omitempty"`
}

// ComfyUISettings contains ComfyUI-specific workflow parameters
type ComfyUISettings struct {
	// WorkflowTemplate defines the ComfyUI workflow template to use
	WorkflowTemplate string `json:"workflowTemplate,omitempty"`

	// CustomNodes defines any required custom nodes
	CustomNodes []string `json:"customNodes,omitempty"`

	// BatchSize for processing multiple frames
	BatchSize int `json:"batchSize,omitempty"`

	// VAE model override
	VAE string `json:"vae,omitempty"`

	// ControlNet settings for pose/composition control
	ControlNet *ControlNetSettings `json:"controlNet,omitempty"`
}

// ControlNetSettings defines ControlNet parameters for pose/composition control
type ControlNetSettings struct {
	// Model defines the ControlNet model to use
	Model string `json:"model"`

	// Strength controls ControlNet influence (0.0-1.0)
	Strength float64 `json:"strength"`

	// StartStep defines when ControlNet starts affecting generation
	StartStep float64 `json:"startStep,omitempty"`

	// EndStep defines when ControlNet stops affecting generation
	EndStep float64 `json:"endStep,omitempty"`

	// Preprocessor defines the preprocessing method
	Preprocessor string `json:"preprocessor,omitempty"`
}

// AssetMetadata tracks version and generation history
type AssetMetadata struct {
	// Version tracks the asset generation schema version
	Version string `json:"version"`

	// GeneratedAt tracks when assets were last generated
	GeneratedAt time.Time `json:"generatedAt,omitempty"`

	// GeneratedBy tracks what tool/version generated the assets
	GeneratedBy string `json:"generatedBy,omitempty"`

	// GenerationHistory tracks previous generation attempts
	GenerationHistory []GenerationRecord `json:"generationHistory,omitempty"`

	// AssetHashes tracks file hashes for change detection
	AssetHashes map[string]string `json:"assetHashes,omitempty"`

	// ValidationResults tracks quality validation results
	ValidationResults *ValidationResults `json:"validationResults,omitempty"`
}

// GenerationRecord tracks a single generation attempt
type GenerationRecord struct {
	// Timestamp of generation attempt
	Timestamp time.Time `json:"timestamp"`

	// Settings used for generation
	Settings GenerationSettings `json:"settings"`

	// Success indicates if generation completed successfully
	Success bool `json:"success"`

	// Error message if generation failed
	Error string `json:"error,omitempty"`

	// GeneratedFiles lists the files that were generated
	GeneratedFiles []string `json:"generatedFiles,omitempty"`

	// Duration of generation process
	Duration time.Duration `json:"duration,omitempty"`
}

// ValidationResults tracks quality validation results
type ValidationResults struct {
	// OverallScore is the overall quality score (0.0-1.0)
	OverallScore float64 `json:"overallScore"`

	// FileSize validation results
	FileSizeValid bool `json:"fileSizeValid"`

	// Animation quality metrics
	AnimationQuality map[string]float64 `json:"animationQuality,omitempty"`

	// Visual consistency score across animations
	ConsistencyScore float64 `json:"consistencyScore,omitempty"`

	// Transparency validation
	TransparencyValid bool `json:"transparencyValid,omitempty"`

	// Validation timestamp
	ValidatedAt time.Time `json:"validatedAt"`
}

// BackupSettings controls asset backup behavior
type BackupSettings struct {
	// Enabled controls whether to backup existing assets before regeneration
	Enabled bool `json:"enabled"`

	// BackupPath defines where to store backups (relative to character folder)
	BackupPath string `json:"backupPath,omitempty"`

	// MaxBackups defines how many backup sets to keep (0 = unlimited)
	MaxBackups int `json:"maxBackups,omitempty"`

	// CompressBackups controls whether to compress backup archives
	CompressBackups bool `json:"compressBackups,omitempty"`
}

// DefaultAssetGenerationConfig returns a sensible default configuration
func DefaultAssetGenerationConfig() *AssetGenerationConfig {
	return &AssetGenerationConfig{
		BasePrompt: "A cute anime character, digital art, transparent background, simple design suitable for desktop companion",
		AnimationMappings: map[string]AnimationMapping{
			"idle": {
				PromptModifier:   "standing calmly, neutral expression, slight smile",
				StateDescription: "Default calm state",
				FrameCount:       6,
			},
			"talking": {
				PromptModifier:   "speaking, mouth open, expressive face, animated gesture",
				StateDescription: "Speaking or interacting",
				FrameCount:       8,
			},
			"happy": {
				PromptModifier:   "smiling brightly, cheerful expression, positive energy",
				StateDescription: "Happy or excited state",
				FrameCount:       6,
			},
			"sad": {
				PromptModifier:   "sad expression, downcast eyes, melancholy mood",
				StateDescription: "Sad or disappointed state",
				FrameCount:       4,
			},
		},
		GenerationSettings: GenerationSettings{
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
				ColorPalette:        "adaptive",
			},
		},
		AssetMetadata: AssetMetadata{
			Version: "1.0.0",
		},
		BackupSettings: BackupSettings{
			Enabled:         true,
			BackupPath:      "backups",
			MaxBackups:      5,
			CompressBackups: true,
		},
	}
}

// ValidateAssetGenerationConfig validates the asset generation configuration
func ValidateAssetGenerationConfig(config *AssetGenerationConfig) error {
	if config == nil {
		return nil // Optional field, nil is valid
	}

	if config.BasePrompt == "" {
		return fmt.Errorf("basePrompt cannot be empty")
	}

	// Validate model
	validModels := []string{"sdxl", "flux1d", "flux1s"}
	modelValid := false
	for _, model := range validModels {
		if config.GenerationSettings.Model == model {
			modelValid = true
			break
		}
	}
	if !modelValid {
		return fmt.Errorf("invalid model %q, must be one of: %v", config.GenerationSettings.Model, validModels)
	}

	// Validate art style
	validStyles := []string{"anime", "pixel_art", "realistic", "cartoon", "chibi"}
	styleValid := false
	for _, style := range validStyles {
		if config.GenerationSettings.ArtStyle == style {
			styleValid = true
			break
		}
	}
	if !styleValid {
		return fmt.Errorf("invalid artStyle %q, must be one of: %v", config.GenerationSettings.ArtStyle, validStyles)
	}

	// Validate resolution
	res := config.GenerationSettings.Resolution
	if res.Width < 64 || res.Width > 512 {
		return fmt.Errorf("invalid width %d, must be between 64-512", res.Width)
	}
	if res.Height < 64 || res.Height > 512 {
		return fmt.Errorf("invalid height %d, must be between 64-512", res.Height)
	}

	// Validate quality settings
	quality := config.GenerationSettings.QualitySettings
	if quality.Steps < 10 || quality.Steps > 100 {
		return fmt.Errorf("invalid steps %d, must be between 10-100", quality.Steps)
	}
	if quality.CFGScale < 1.0 || quality.CFGScale > 20.0 {
		return fmt.Errorf("invalid cfgScale %.1f, must be between 1.0-20.0", quality.CFGScale)
	}

	// Validate animation settings
	anim := config.GenerationSettings.AnimationSettings
	if anim.FrameRate < 5 || anim.FrameRate > 30 {
		return fmt.Errorf("invalid frameRate %d, must be between 5-30", anim.FrameRate)
	}
	if anim.Duration < 1.0 || anim.Duration > 10.0 {
		return fmt.Errorf("invalid duration %.1f, must be between 1.0-10.0", anim.Duration)
	}

	return nil
}
