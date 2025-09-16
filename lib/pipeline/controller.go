package pipeline

// controller.go provides the main pipeline orchestration for the asset generation
// system. This implements the Controller interface outlined in GIF_PLAN.md,
// coordinating ComfyUI integration, asset generation, validation, and deployment.

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/opd-ai/desktop-companion/lib/assets"
	"github.com/opd-ai/desktop-companion/lib/comfyui"
)

// Controller orchestrates the complete generation pipeline.
type Controller interface {
	// ProcessCharacter generates complete asset set for one character
	ProcessCharacter(ctx context.Context, config *CharacterConfig) (*ProcessResult, error)

	// ProcessBatch generates assets for multiple characters
	ProcessBatch(ctx context.Context, configs []*CharacterConfig) (*BatchResult, error)

	// ValidateAssets checks generated assets for compliance
	ValidateAssets(ctx context.Context, assetPath string) (*ValidationResult, error)

	// DeployAssets moves validated assets to target locations
	DeployAssets(ctx context.Context, result *ProcessResult) error
}

// ProcessResult contains the result of processing a single character.
type ProcessResult struct {
	Character        string                     `json:"character"`
	Success          bool                       `json:"success"`
	GeneratedAssets  map[string]*GeneratedAsset `json:"generated_assets"`
	ValidationResult *ValidationResult          `json:"validation_result,omitempty"`
	Errors           []ProcessError             `json:"errors,omitempty"`
	Warnings         []ProcessWarning           `json:"warnings,omitempty"`
	ProcessingTime   time.Duration              `json:"processing_time"`
	Metadata         *ProcessMetadata           `json:"metadata,omitempty"`
}

// BatchResult contains the result of processing multiple characters.
type BatchResult struct {
	Characters     map[string]*ProcessResult `json:"characters"`
	OverallSuccess bool                      `json:"overall_success"`
	Summary        *BatchSummary             `json:"summary"`
	ProcessingTime time.Duration             `json:"processing_time"`
	ConcurrentJobs int                       `json:"concurrent_jobs"`
}

// GeneratedAsset represents a single generated asset file.
type GeneratedAsset struct {
	State          string        `json:"state"`        // Animation state (idle, talking, etc.)
	SourceFiles    []string      `json:"source_files"` // Original frame files
	OutputPath     string        `json:"output_path"`  // Final GIF path
	Metrics        *AssetMetrics `json:"metrics,omitempty"`
	JobID          string        `json:"job_id,omitempty"` // ComfyUI job ID
	GenerationTime time.Duration `json:"generation_time"`
}

// ProcessError represents an error during processing.
type ProcessError struct {
	Stage     string    `json:"stage"`           // Stage where error occurred
	Message   string    `json:"message"`         // Error description
	Cause     string    `json:"cause,omitempty"` // Underlying cause
	Retryable bool      `json:"retryable"`       // Whether error is retryable
	Timestamp time.Time `json:"timestamp"`
}

// ProcessWarning represents a non-critical issue during processing.
type ProcessWarning struct {
	Stage     string    `json:"stage"`            // Stage where warning occurred
	Message   string    `json:"message"`          // Warning description
	Impact    string    `json:"impact,omitempty"` // Potential impact
	Timestamp time.Time `json:"timestamp"`
}

// ProcessMetadata contains additional processing information.
type ProcessMetadata struct {
	ComfyUIVersion   string                 `json:"comfyui_version,omitempty"`
	WorkflowUsed     string                 `json:"workflow_used,omitempty"`
	ModelUsed        string                 `json:"model_used,omitempty"`
	StyleApplied     string                 `json:"style_applied,omitempty"`
	GenerationParams map[string]interface{} `json:"generation_params,omitempty"`
	TempDir          string                 `json:"temp_dir,omitempty"`
}

// BatchSummary provides aggregate statistics for batch processing.
type BatchSummary struct {
	TotalCharacters       int           `json:"total_characters"`
	SuccessfulCharacters  int           `json:"successful_characters"`
	FailedCharacters      int           `json:"failed_characters"`
	TotalAssets           int           `json:"total_assets"`
	SuccessfulAssets      int           `json:"successful_assets"`
	FailedAssets          int           `json:"failed_assets"`
	AverageProcessingTime time.Duration `json:"average_processing_time"`
	SuccessRate           float64       `json:"success_rate"`
}

// pipelineController is the concrete implementation of Controller.
type pipelineController struct {
	config         *PipelineConfig
	comfyuiClient  comfyui.Client
	assetProcessor assets.ArtifactPostProcessor
	validator      Validator
	mu             sync.RWMutex
}

// NewController creates a new pipeline controller instance.
func NewController(config *PipelineConfig, comfyuiClient comfyui.Client) (Controller, error) {
	if config == nil {
		return nil, fmt.Errorf("pipeline config required")
	}
	if comfyuiClient == nil {
		return nil, fmt.Errorf("comfyui client required")
	}

	return &pipelineController{
		config:         config,
		comfyuiClient:  comfyuiClient,
		assetProcessor: &assets.GIFAssembler{},
		validator:      NewValidator(),
	}, nil
}

// ProcessCharacter generates complete asset set for one character.
func (c *pipelineController) ProcessCharacter(ctx context.Context, config *CharacterConfig) (*ProcessResult, error) {
	if config == nil {
		return nil, fmt.Errorf("character config required")
	}

	startTime := time.Now()
	result := &ProcessResult{
		Character:       config.Character.Archetype,
		GeneratedAssets: make(map[string]*GeneratedAsset),
		Metadata: &ProcessMetadata{
			StyleApplied:     config.Character.Style,
			GenerationParams: make(map[string]interface{}),
		},
	}

	// Create temporary directory for processing
	tempDir, err := c.createTempDir(config.Character.Archetype)
	if err != nil {
		return nil, fmt.Errorf("create temp directory: %w", err)
	}
	result.Metadata.TempDir = tempDir
	defer c.cleanupTempDir(tempDir)

	// Generate assets for each required state
	for _, state := range config.States {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		asset, err := c.generateAssetForState(ctx, config, state, tempDir)
		if err != nil {
			result.Errors = append(result.Errors, ProcessError{
				Stage:     "generation",
				Message:   fmt.Sprintf("Failed to generate asset for state %s: %v", state, err),
				Cause:     err.Error(),
				Retryable: c.isRetryableError(err),
				Timestamp: time.Now(),
			})
			continue
		}

		result.GeneratedAssets[state] = asset
	}

	// Validate generated assets
	if len(result.GeneratedAssets) > 0 {
		validationResult, err := c.validateGeneratedAssets(ctx, result.GeneratedAssets, config)
		if err != nil {
			result.Warnings = append(result.Warnings, ProcessWarning{
				Stage:     "validation",
				Message:   fmt.Sprintf("Asset validation failed: %v", err),
				Impact:    "Assets may not meet quality requirements",
				Timestamp: time.Now(),
			})
		} else {
			result.ValidationResult = validationResult
		}
	}

	// Determine overall success
	result.Success = len(result.Errors) == 0 && len(result.GeneratedAssets) == len(config.States)
	result.ProcessingTime = time.Since(startTime)

	return result, nil
}

// ProcessBatch generates assets for multiple characters.
func (c *pipelineController) ProcessBatch(ctx context.Context, configs []*CharacterConfig) (*BatchResult, error) {
	if len(configs) == 0 {
		return nil, fmt.Errorf("no character configs provided")
	}

	startTime := time.Now()
	result := &BatchResult{
		Characters:     make(map[string]*ProcessResult),
		ConcurrentJobs: c.config.Generation.ConcurrentJobs,
	}

	// Process characters with controlled concurrency
	semaphore := make(chan struct{}, c.config.Generation.ConcurrentJobs)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, config := range configs {
		wg.Add(1)
		go func(cfg *CharacterConfig) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Process character
			charResult, err := c.ProcessCharacter(ctx, cfg)
			if err != nil {
				charResult = &ProcessResult{
					Character: cfg.Character.Archetype,
					Success:   false,
					Errors: []ProcessError{{
						Stage:     "pipeline",
						Message:   fmt.Sprintf("Pipeline processing failed: %v", err),
						Cause:     err.Error(),
						Retryable: false,
						Timestamp: time.Now(),
					}},
				}
			}

			// Store result
			mu.Lock()
			result.Characters[cfg.Character.Archetype] = charResult
			mu.Unlock()
		}(config)
	}

	wg.Wait()

	// Generate summary
	result.Summary = c.generateBatchSummary(result.Characters)
	result.OverallSuccess = result.Summary.FailedCharacters == 0
	result.ProcessingTime = time.Since(startTime)

	return result, nil
}

// ValidateAssets checks generated assets for compliance.
func (c *pipelineController) ValidateAssets(ctx context.Context, assetPath string) (*ValidationResult, error) {
	if assetPath == "" {
		return nil, fmt.Errorf("asset path required")
	}

	// Use default validation config if not provided
	validationConfig := &c.config.Validation

	return c.validator.ValidateAsset(ctx, assetPath, validationConfig)
}

// DeployAssets moves validated assets to target locations.
func (c *pipelineController) DeployAssets(ctx context.Context, result *ProcessResult) error {
	if result == nil {
		return fmt.Errorf("process result required")
	}
	if !result.Success {
		return fmt.Errorf("cannot deploy failed processing result")
	}

	// Validate assets before deployment if required
	if c.config.Deployment.ValidateBeforeDeploy {
		for state, asset := range result.GeneratedAssets {
			validationResult, err := c.ValidateAssets(ctx, asset.OutputPath)
			if err != nil {
				return fmt.Errorf("pre-deployment validation failed for %s: %w", state, err)
			}
			if !validationResult.Valid {
				return fmt.Errorf("asset %s failed validation: %d errors", state, len(validationResult.Errors))
			}
		}
	}

	// Create target directory
	targetDir := c.config.Deployment.OutputDir
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("create target directory: %w", err)
	}

	// Deploy each asset
	for state, asset := range result.GeneratedAssets {
		targetPath := filepath.Join(targetDir, state+".gif")

		// Backup existing file if required
		if c.config.Deployment.BackupExisting {
			if err := c.backupExistingAsset(targetPath); err != nil {
				return fmt.Errorf("backup existing asset %s: %w", state, err)
			}
		}

		// Copy asset to target location
		if err := c.copyFile(asset.OutputPath, targetPath); err != nil {
			return fmt.Errorf("deploy asset %s: %w", state, err)
		}
	}

	return nil
}

// generateAssetForState generates a single asset for a character state.
func (c *pipelineController) generateAssetForState(ctx context.Context, config *CharacterConfig, state string, tempDir string) (*GeneratedAsset, error) {
	startTime := time.Now()

	// Create workflow for this state
	workflow, err := c.createWorkflowForState(config, state)
	if err != nil {
		return nil, fmt.Errorf("create workflow: %w", err)
	}

	// Submit workflow to ComfyUI
	job, err := c.comfyuiClient.SubmitWorkflow(ctx, workflow)
	if err != nil {
		return nil, fmt.Errorf("submit workflow: %w", err)
	}

	// Monitor job progress
	progressChan, err := c.comfyuiClient.MonitorJob(ctx, job.ID)
	if err != nil {
		return nil, fmt.Errorf("monitor job: %w", err)
	}

	// Wait for completion
	var finalProgress comfyui.JobProgress
	for progress := range progressChan {
		finalProgress = progress
		if progress.Err != nil {
			return nil, fmt.Errorf("job failed: %w", progress.Err)
		}
	}

	if finalProgress.Status != "completed" {
		return nil, fmt.Errorf("job failed with status: %s", finalProgress.Status)
	}

	// Get job result
	jobResult, err := c.comfyuiClient.GetResult(ctx, job.ID)
	if err != nil {
		return nil, fmt.Errorf("get job result: %w", err)
	}

	// Save artifacts to temp directory
	framesDir := filepath.Join(tempDir, state+"_frames")
	if err := os.MkdirAll(framesDir, 0755); err != nil {
		return nil, fmt.Errorf("create frames directory: %w", err)
	}

	if err := comfyui.SaveArtifacts(jobResult, framesDir); err != nil {
		return nil, fmt.Errorf("save artifacts: %w", err)
	}

	// Collect frame files
	frameFiles, err := c.collectFrameFiles(framesDir)
	if err != nil {
		return nil, fmt.Errorf("collect frame files: %w", err)
	}

	// Create GIF from frames
	outputPath := filepath.Join(tempDir, state+".gif")
	gifConfig := assets.GIFConfig{
		Width:        config.GIFConfig.Width,
		Height:       config.GIFConfig.Height,
		FrameCount:   config.GIFConfig.FrameCount,
		FrameRate:    config.GIFConfig.FrameRate,
		MaxFileSize:  config.GIFConfig.MaxFileSize,
		Transparency: config.GIFConfig.Transparency,
	}

	if err := c.assetProcessor.Process(ctx, frameFiles, outputPath, gifConfig); err != nil {
		return nil, fmt.Errorf("create GIF: %w", err)
	}

	// Extract metrics
	metrics, err := c.extractAssetMetrics(outputPath)
	if err != nil {
		return nil, fmt.Errorf("extract metrics: %w", err)
	}

	return &GeneratedAsset{
		State:          state,
		SourceFiles:    frameFiles,
		OutputPath:     outputPath,
		Metrics:        metrics,
		JobID:          job.ID,
		GenerationTime: time.Since(startTime),
	}, nil
}

// createWorkflowForState creates a ComfyUI workflow for a character state.
func (c *pipelineController) createWorkflowForState(config *CharacterConfig, state string) (*comfyui.Workflow, error) {
	// This is a simplified workflow creation - in a full implementation,
	// this would use the workflow template system and dynamic prompt injection
	workflow := &comfyui.Workflow{
		ID: fmt.Sprintf("%s_%s_%d", config.Character.Archetype, state, time.Now().Unix()),
		Nodes: map[string]interface{}{
			"prompt": map[string]interface{}{
				"positive": c.buildPositivePrompt(config, state),
				"negative": c.buildNegativePrompt(config, state),
			},
			"generation": map[string]interface{}{
				"width":     config.Character.OutputConfig.Width,
				"height":    config.Character.OutputConfig.Height,
				"steps":     c.config.Workflow.Quality.Steps,
				"cfg_scale": c.config.Workflow.Quality.CFGScale,
				"sampler":   c.config.Workflow.Quality.Sampler,
				"scheduler": c.config.Workflow.Quality.Scheduler,
				"seed":      c.config.Workflow.Quality.Seed,
			},
		},
		Meta: map[string]interface{}{
			"archetype": config.Character.Archetype,
			"state":     state,
			"style":     config.Character.Style,
		},
	}

	return workflow, nil
}

// buildPositivePrompt constructs the positive prompt for generation.
func (c *pipelineController) buildPositivePrompt(config *CharacterConfig, state string) string {
	// Base character description
	prompt := config.Character.Description

	// Add style-specific prompts
	if style, exists := c.config.Workflow.Styles[config.Character.Style]; exists {
		prompt += ", " + style.Prompts.Positive
	}

	// Add state-specific modifiers
	stateModifiers := map[string]string{
		"idle":    "relaxed, calm, neutral expression",
		"talking": "speaking, mouth open, expressive",
		"happy":   "smiling, joyful, positive expression",
		"sad":     "downcast, melancholy, subdued",
		"shy":     "blushing, bashful, looking away",
		"flirty":  "winking, playful, confident",
		"loving":  "gentle smile, warm expression, affectionate",
		"jealous": "frowning, crossed arms, possessive",
	}

	if modifier, exists := stateModifiers[state]; exists {
		prompt += ", " + modifier
	}

	return prompt
}

// buildNegativePrompt constructs the negative prompt for generation.
func (c *pipelineController) buildNegativePrompt(config *CharacterConfig, state string) string {
	prompt := "blurry, low quality, distorted, malformed"

	// Add style-specific negative prompts
	if style, exists := c.config.Workflow.Styles[config.Character.Style]; exists {
		if style.Prompts.Negative != "" {
			prompt += ", " + style.Prompts.Negative
		}
	}

	return prompt
}

// collectFrameFiles finds all image files in a directory for GIF creation.
func (c *pipelineController) collectFrameFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read directory: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := filepath.Ext(entry.Name())
		if ext == ".png" || ext == ".jpg" || ext == ".jpeg" {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no image files found in %s", dir)
	}

	return files, nil
}

// extractAssetMetrics extracts metrics from a generated asset.
func (c *pipelineController) extractAssetMetrics(assetPath string) (*AssetMetrics, error) {
	validator := &assetValidator{}
	return validator.extractMetrics(assetPath)
}

// validateGeneratedAssets validates all generated assets for a character.
func (c *pipelineController) validateGeneratedAssets(ctx context.Context, assets map[string]*GeneratedAsset, config *CharacterConfig) (*ValidationResult, error) {
	// Create a temporary character validation
	assetResults := make(map[string]*ValidationResult)

	for state, asset := range assets {
		result, err := c.validator.ValidateAsset(ctx, asset.OutputPath, config.Validation)
		if err != nil {
			return nil, fmt.Errorf("validate asset %s: %w", state, err)
		}
		assetResults[state] = result
	}

	// Create overall validation result
	overall := &ValidationResult{
		AssetPath:        config.Character.Archetype,
		ComplianceChecks: make(map[string]bool),
		Timestamp:        time.Now(),
	}

	// Aggregate results
	totalErrors := 0
	for _, result := range assetResults {
		totalErrors += len(result.Errors)
		overall.Warnings = append(overall.Warnings, result.Warnings...)

		// Merge compliance checks
		for check, passed := range result.ComplianceChecks {
			if existing, exists := overall.ComplianceChecks[check]; !exists || !existing {
				overall.ComplianceChecks[check] = passed
			}
		}
	}

	overall.Valid = totalErrors == 0

	return overall, nil
}

// createTempDir creates a temporary directory for processing.
func (c *pipelineController) createTempDir(character string) (string, error) {
	baseDir := c.config.Generation.TempDir
	if baseDir == "" {
		baseDir = "temp/generation"
	}

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return "", fmt.Errorf("create base temp directory: %w", err)
	}

	tempDir := filepath.Join(baseDir, fmt.Sprintf("%s_%d", character, time.Now().Unix()))
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("create character temp directory: %w", err)
	}

	return tempDir, nil
}

// cleanupTempDir removes temporary processing directory.
func (c *pipelineController) cleanupTempDir(tempDir string) {
	if tempDir != "" {
		os.RemoveAll(tempDir)
	}
}

// backupExistingAsset creates a backup of an existing asset file.
func (c *pipelineController) backupExistingAsset(assetPath string) error {
	if _, err := os.Stat(assetPath); os.IsNotExist(err) {
		return nil // No existing file to backup
	}

	backupPath := assetPath + ".backup." + fmt.Sprintf("%d", time.Now().Unix())
	return c.copyFile(assetPath, backupPath)
}

// copyFile copies a file from source to destination.
func (c *pipelineController) copyFile(src, dst string) error {
	sourceData, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read source file: %w", err)
	}

	if err := os.WriteFile(dst, sourceData, 0644); err != nil {
		return fmt.Errorf("write destination file: %w", err)
	}

	return nil
}

// isRetryableError determines if an error is retryable.
func (c *pipelineController) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Simple heuristic - check error message for retryable conditions
	errStr := err.Error()
	retryableMessages := []string{
		"timeout",
		"connection refused",
		"server error",
		"temporary failure",
		"queue full",
	}

	for _, msg := range retryableMessages {
		if contains(errStr, msg) {
			return true
		}
	}

	return false
}

// generateBatchSummary creates aggregate statistics for batch processing.
func (c *pipelineController) generateBatchSummary(characters map[string]*ProcessResult) *BatchSummary {
	summary := &BatchSummary{
		TotalCharacters: len(characters),
	}

	var totalProcessingTime time.Duration

	for _, result := range characters {
		if result.Success {
			summary.SuccessfulCharacters++
		} else {
			summary.FailedCharacters++
		}

		summary.TotalAssets += len(result.GeneratedAssets)
		for _, asset := range result.GeneratedAssets {
			if asset.Metrics != nil {
				summary.SuccessfulAssets++
			} else {
				summary.FailedAssets++
			}
		}

		totalProcessingTime += result.ProcessingTime
	}

	if summary.TotalCharacters > 0 {
		summary.AverageProcessingTime = totalProcessingTime / time.Duration(summary.TotalCharacters)
		summary.SuccessRate = float64(summary.SuccessfulCharacters) / float64(summary.TotalCharacters)
	}

	return summary
}

// contains checks if a string contains a substring (case-insensitive).
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(s) > len(substr) &&
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
				findSubstring(s, substr)))
}

// findSubstring is a simple substring search helper.
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
