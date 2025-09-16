package assets

// generator.go implements the core asset generation logic that integrates
// character.json assetGeneration configurations with the ComfyUI pipeline.
// This follows the "lazy programmer" philosophy by leveraging existing
// ComfyUI workflows and minimal custom integration code.

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/opd-ai/desktop-companion/lib/character"
	"github.com/opd-ai/desktop-companion/lib/comfyui"
)

// AssetGenerator orchestrates the generation of character animation assets
// from assetGeneration configurations in character.json files.
type AssetGenerator struct {
	comfyClient   comfyui.Client
	config        *GeneratorConfig
	validator     *AssetValidator
	backupManager *BackupManager
}

// GeneratorConfig defines asset generator configuration.
type GeneratorConfig struct {
	// ComfyUI server configuration
	ComfyUIURL string

	// Default generation settings
	DefaultSettings *character.GenerationSettings

	// Workflow templates directory
	WorkflowsPath string

	// Output and temporary directories
	OutputDir string
	TempDir   string

	// Quality validation settings
	ValidateOutput bool

	// Backup settings
	BackupExisting bool
	BackupDir      string
}

// NewAssetGenerator creates a new asset generator with the given configuration.
func NewAssetGenerator(config *GeneratorConfig) (*AssetGenerator, error) {
	if config == nil {
		config = DefaultGeneratorConfig()
	}

	// Create ComfyUI client
	clientConfig := comfyui.DefaultConfig()
	clientConfig.ServerURL = config.ComfyUIURL

	client, err := comfyui.New(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("create ComfyUI client: %w", err)
	}

	// Create validator and backup manager
	validator := NewAssetValidator(nil)    // Use default validation config
	backupManager := NewBackupManager(nil) // Use default backup config

	return &AssetGenerator{
		comfyClient:   client,
		config:        config,
		validator:     validator,
		backupManager: backupManager,
	}, nil
}

// DefaultGeneratorConfig returns sensible default configuration.
func DefaultGeneratorConfig() *GeneratorConfig {
	return &GeneratorConfig{
		ComfyUIURL: "http://localhost:8188",
		DefaultSettings: &character.GenerationSettings{
			Model:    "flux1d",
			ArtStyle: "anime",
			Resolution: character.ImageResolution{
				Width:  128,
				Height: 128,
			},
			QualitySettings: character.QualitySettings{
				Steps:     25,
				CFGScale:  7.0,
				Sampler:   "euler_a",
				Scheduler: "normal",
			},
			AnimationSettings: character.AnimationSettings{
				FrameRate:           12,
				Duration:            2.0,
				LoopType:            "seamless",
				Optimization:        "balanced",
				MaxFileSize:         500,
				TransparencyEnabled: true,
				ColorPalette:        "adaptive",
			},
		},
		WorkflowsPath:  "templates/workflows",
		ValidateOutput: true,
		BackupExisting: true,
		BackupDir:      "backups",
	}
}

// GenerateResult contains the results of asset generation.
type GenerateResult struct {
	// Success indicates if generation completed successfully
	Success bool

	// GeneratedAssets maps animation state to generated file path
	GeneratedAssets map[string]string

	// BackupPath contains the path to backed up original assets (if any)
	BackupPath string

	// Metadata tracks generation information
	Metadata *character.AssetMetadata

	// Errors encountered during generation
	Errors []GenerationError

	// Processing time
	Duration time.Duration
}

// GenerationError represents an error that occurred during generation.
type GenerationError struct {
	Stage   string // Stage where error occurred
	State   string // Animation state (if applicable)
	Message string // Error message
	Err     error  // Underlying error
}

// GenerateAssetsFromCharacter generates animation assets from a character configuration.
func (g *AssetGenerator) GenerateAssetsFromCharacter(ctx context.Context, card *character.CharacterCard, basePath string) (*GenerateResult, error) {
	startTime := time.Now()

	result := &GenerateResult{
		GeneratedAssets: make(map[string]string),
		Errors:          []GenerationError{},
	}

	// Check if character has asset generation configuration
	if card.AssetGeneration == nil {
		return nil, fmt.Errorf("character does not have assetGeneration configuration")
	}

	assetConfig := card.AssetGeneration

	// Validate asset generation configuration
	if err := character.ValidateAssetGenerationConfig(assetConfig); err != nil {
		return nil, fmt.Errorf("invalid assetGeneration config: %w", err)
	}

	// Create backup if enabled
	if assetConfig.BackupSettings.Enabled {
		backupResult, err := g.backupManager.CreateBackup(basePath, card.Animations)
		if err != nil {
			result.Errors = append(result.Errors, GenerationError{
				Stage:   "backup",
				Message: "Failed to create backup",
				Err:     err,
			})
		} else if !backupResult.Success {
			result.Errors = append(result.Errors, GenerationError{
				Stage:   "backup",
				Message: fmt.Sprintf("Backup failed: %v", backupResult.Errors),
				Err:     fmt.Errorf("backup creation failed"),
			})
		} else {
			result.BackupPath = backupResult.BackupPath
		}
	}

	// Generate assets for each animation state
	for state, mapping := range assetConfig.AnimationMappings {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		assetPath, err := g.generateAnimationAsset(ctx, state, mapping, assetConfig, basePath)
		if err != nil {
			result.Errors = append(result.Errors, GenerationError{
				Stage:   "generation",
				State:   state,
				Message: fmt.Sprintf("Failed to generate %s animation", state),
				Err:     err,
			})
			continue
		}

		result.GeneratedAssets[state] = assetPath
	}

	// Validate generated assets if enabled
	if g.config.ValidateOutput {
		validationResults, err := g.validator.ValidateCharacterAssets(card, basePath)
		if err != nil {
			result.Errors = append(result.Errors, GenerationError{
				Stage:   "validation",
				Message: "Asset validation failed",
				Err:     err,
			})
		} else {
			// Check if any assets failed validation
			for state, validationResult := range validationResults {
				if !validationResult.Valid {
					result.Errors = append(result.Errors, GenerationError{
						Stage:   "validation",
						State:   state,
						Message: fmt.Sprintf("Asset validation failed: %v", validationResult.Errors),
						Err:     fmt.Errorf("validation failed for %s", state),
					})
				}
			}
		}
	}

	// Update character.json with generation metadata
	result.Metadata = g.createAssetMetadata(assetConfig, result)
	if err := g.updateCharacterMetadata(card, result.Metadata, basePath); err != nil {
		result.Errors = append(result.Errors, GenerationError{
			Stage:   "metadata",
			Message: "Failed to update character metadata",
			Err:     err,
		})
	}

	result.Success = len(result.Errors) == 0
	result.Duration = time.Since(startTime)

	return result, nil
}

// generateAnimationAsset generates a single animation asset for the given state.
func (g *AssetGenerator) generateAnimationAsset(ctx context.Context, state string, mapping character.AnimationMapping, config *character.AssetGenerationConfig, basePath string) (string, error) {
	// Build the complete prompt
	prompt := g.buildPrompt(config.BasePrompt, mapping)

	// Create ComfyUI workflow
	workflow, err := g.createWorkflow(prompt, mapping, config)
	if err != nil {
		return "", fmt.Errorf("create workflow: %w", err)
	}

	// Submit workflow to ComfyUI
	job, err := g.comfyClient.SubmitWorkflow(ctx, workflow)
	if err != nil {
		return "", fmt.Errorf("submit workflow: %w", err)
	}

	// Monitor job progress
	progressChan, err := g.comfyClient.MonitorJob(ctx, job.ID)
	if err != nil {
		return "", fmt.Errorf("monitor job: %w", err)
	}

	// Wait for completion
	var finalProgress comfyui.JobProgress
	for progress := range progressChan {
		finalProgress = progress
		if progress.Err != nil {
			return "", fmt.Errorf("job failed: %w", progress.Err)
		}
	}

	if finalProgress.Status != "completed" {
		return "", fmt.Errorf("job did not complete successfully, status: %s", finalProgress.Status)
	}

	// Get job results
	result, err := g.comfyClient.GetResult(ctx, job.ID)
	if err != nil {
		return "", fmt.Errorf("get result: %w", err)
	}

	// Process and save the generated asset
	outputPath := filepath.Join(basePath, "animations", fmt.Sprintf("%s.gif", state))
	if err := g.processJobResult(result, outputPath, mapping, config); err != nil {
		return "", fmt.Errorf("process result: %w", err)
	}

	return outputPath, nil
}

// buildPrompt constructs the complete prompt for generation.
func (g *AssetGenerator) buildPrompt(basePrompt string, mapping character.AnimationMapping) string {
	prompt := basePrompt
	if mapping.PromptModifier != "" {
		prompt += ", " + mapping.PromptModifier
	}
	return prompt
}

// createWorkflow creates a ComfyUI workflow for the given parameters.
func (g *AssetGenerator) createWorkflow(prompt string, mapping character.AnimationMapping, config *character.AssetGenerationConfig) (*comfyui.Workflow, error) {
	// This is a simplified workflow creation
	// In a full implementation, this would load and customize workflow templates

	settings := config.GenerationSettings
	if mapping.CustomSettings != nil {
		// Apply custom settings override
		settings = *mapping.CustomSettings
	}

	workflow := &comfyui.Workflow{
		ID: fmt.Sprintf("character_gen_%d", time.Now().Unix()),
		Nodes: map[string]interface{}{
			"text_prompt": map[string]interface{}{
				"class_type": "CLIPTextEncode",
				"inputs": map[string]interface{}{
					"text": prompt,
				},
			},
			"negative_prompt": map[string]interface{}{
				"class_type": "CLIPTextEncode",
				"inputs": map[string]interface{}{
					"text": mapping.NegativePrompt,
				},
			},
			"sampler": map[string]interface{}{
				"class_type": "KSampler",
				"inputs": map[string]interface{}{
					"seed":      settings.QualitySettings.Seed,
					"steps":     settings.QualitySettings.Steps,
					"cfg":       settings.QualitySettings.CFGScale,
					"sampler":   settings.QualitySettings.Sampler,
					"scheduler": settings.QualitySettings.Scheduler,
					"width":     settings.Resolution.Width,
					"height":    settings.Resolution.Height,
				},
			},
		},
	}

	return workflow, nil
}

// processJobResult processes the ComfyUI job result and saves the asset.
func (g *AssetGenerator) processJobResult(result *comfyui.JobResult, outputPath string, mapping character.AnimationMapping, config *character.AssetGenerationConfig) error {
	// This is a simplified implementation
	// In a full implementation, this would:
	// 1. Extract generated images from the job result
	// 2. Create animation frames if needed
	// 3. Assemble frames into a GIF
	// 4. Apply optimization and compression
	// 5. Validate the output

	if len(result.Artifacts) == 0 {
		return fmt.Errorf("no artifacts generated")
	}

	// For now, just save the first artifact
	artifact := result.Artifacts[0]

	// Ensure output directory exists
	if err := ensureDir(filepath.Dir(outputPath)); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	// Save the artifact
	if err := saveArtifactToFile(artifact, outputPath); err != nil {
		return fmt.Errorf("save artifact: %w", err)
	}

	return nil
}

// createAssetMetadata creates metadata for the generation process.
func (g *AssetGenerator) createAssetMetadata(config *character.AssetGenerationConfig, result *GenerateResult) *character.AssetMetadata {
	metadata := &character.AssetMetadata{
		Version:     "1.0.0",
		GeneratedAt: time.Now(),
		GeneratedBy: "gif-generator v1.0.0",
		GenerationHistory: []character.GenerationRecord{
			{
				Timestamp:      time.Now(),
				Settings:       config.GenerationSettings,
				Success:        result.Success,
				GeneratedFiles: make([]string, 0, len(result.GeneratedAssets)),
				Duration:       result.Duration,
			},
		},
		AssetHashes: make(map[string]string),
	}

	// Add generated files to history
	for _, path := range result.GeneratedAssets {
		metadata.GenerationHistory[0].GeneratedFiles = append(
			metadata.GenerationHistory[0].GeneratedFiles,
			filepath.Base(path),
		)
	}

	// Add error information if any
	if len(result.Errors) > 0 {
		metadata.GenerationHistory[0].Error = fmt.Sprintf("%d errors occurred during generation", len(result.Errors))
	}

	return metadata
}

// updateCharacterMetadata updates the character.json file with generation metadata.
func (g *AssetGenerator) updateCharacterMetadata(card *character.CharacterCard, metadata *character.AssetMetadata, basePath string) error {
	// Update the asset generation metadata
	if card.AssetGeneration != nil {
		card.AssetGeneration.AssetMetadata = *metadata
	}

	// TODO: Implement character.json file update logic
	// This would involve:
	// 1. Reading the existing character.json
	// 2. Updating the assetGeneration.assetMetadata section
	// 3. Writing the updated JSON back to disk

	return nil
}

// ensureDir creates a directory if it doesn't exist.
func ensureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

// saveArtifactToFile saves a ComfyUI artifact to the specified file path.
func saveArtifactToFile(artifact comfyui.Artifact, outputPath string) error {
	return os.WriteFile(outputPath, artifact.Data, 0o644)
}
