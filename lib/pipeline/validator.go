package pipeline

// validator.go provides comprehensive asset validation, quality metrics, and
// compatibility testing for generated GIF assets. This implements the validation
// requirements outlined in GIF_PLAN.md with the project's interface-first design.

import (
	"context"
	"errors"
	"fmt"
	"image/gif"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Validator interface defines asset validation operations.
type Validator interface {
	// ValidateAsset checks a single asset file for compliance with requirements
	ValidateAsset(ctx context.Context, assetPath string, config *ValidationConfig) (*ValidationResult, error)

	// ValidateCharacterSet checks all assets for a character archetype
	ValidateCharacterSet(ctx context.Context, characterDir string, config *CharacterConfig) (*CharacterValidationResult, error)

	// ValidateBatch checks multiple character sets in parallel
	ValidateBatch(ctx context.Context, configs []*CharacterConfig) (*BatchValidationResult, error)
}

// ValidationResult contains the result of validating a single asset.
type ValidationResult struct {
	AssetPath        string              `json:"asset_path"`
	Valid            bool                `json:"valid"`
	Errors           []ValidationError   `json:"errors,omitempty"`
	Warnings         []ValidationWarning `json:"warnings,omitempty"`
	Metrics          *AssetMetrics       `json:"metrics"`
	ComplianceChecks map[string]bool     `json:"compliance_checks"`
	Timestamp        time.Time           `json:"timestamp"`
}

// CharacterValidationResult contains validation results for a complete character set.
type CharacterValidationResult struct {
	Character        string                       `json:"character"`
	Valid            bool                         `json:"valid"`
	AssetResults     map[string]*ValidationResult `json:"asset_results"`
	MissingStates    []string                     `json:"missing_states,omitempty"`
	StyleConsistency *StyleConsistencyResult      `json:"style_consistency,omitempty"`
	Overall          *ValidationResult            `json:"overall"`
}

// BatchValidationResult contains validation results for multiple characters.
type BatchValidationResult struct {
	Characters     map[string]*CharacterValidationResult `json:"characters"`
	OverallValid   bool                                  `json:"overall_valid"`
	Summary        *ValidationSummary                    `json:"summary"`
	ProcessingTime time.Duration                         `json:"processing_time"`
}

// ValidationError represents a validation failure.
type ValidationError struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	Severity string `json:"severity"` // "error", "warning", "info"
	Field    string `json:"field,omitempty"`
	Expected string `json:"expected,omitempty"`
	Actual   string `json:"actual,omitempty"`
}

// ValidationWarning represents a non-critical validation issue.
type ValidationWarning struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
}

// AssetMetrics contains measurable properties of an asset.
type AssetMetrics struct {
	FileSize         int64         `json:"file_size"`                   // File size in bytes
	Dimensions       [2]int        `json:"dimensions"`                  // [width, height]
	FrameCount       int           `json:"frame_count"`                 // Number of frames
	Duration         time.Duration `json:"duration"`                    // Animation duration
	FrameRate        float64       `json:"frame_rate"`                  // Effective frame rate
	Colors           int           `json:"colors"`                      // Color count
	HasTransparency  bool          `json:"has_transparency"`            // Transparency support
	CompressionRatio float64       `json:"compression_ratio,omitempty"` // Size efficiency
}

// StyleConsistencyResult contains style consistency analysis.
type StyleConsistencyResult struct {
	Consistent      bool                 `json:"consistent"`
	Score           float64              `json:"score"` // 0.0-1.0 consistency score
	Inconsistencies []StyleInconsistency `json:"inconsistencies,omitempty"`
	Recommendations []string             `json:"recommendations,omitempty"`
}

// StyleInconsistency represents a style consistency issue.
type StyleInconsistency struct {
	States   []string `json:"states"`   // Affected animation states
	Issue    string   `json:"issue"`    // Description of inconsistency
	Severity string   `json:"severity"` // "major", "minor"
}

// ValidationSummary provides aggregate validation statistics.
type ValidationSummary struct {
	TotalAssets   int     `json:"total_assets"`
	ValidAssets   int     `json:"valid_assets"`
	InvalidAssets int     `json:"invalid_assets"`
	ErrorCount    int     `json:"error_count"`
	WarningCount  int     `json:"warning_count"`
	SuccessRate   float64 `json:"success_rate"`
}

// assetValidator is the concrete implementation of Validator.
type assetValidator struct{}

// NewValidator creates a new asset validator instance.
func NewValidator() Validator {
	return &assetValidator{}
}

// ValidateAsset checks a single asset file for compliance with requirements.
func (v *assetValidator) ValidateAsset(ctx context.Context, assetPath string, config *ValidationConfig) (*ValidationResult, error) {
	if assetPath == "" {
		return nil, errors.New("asset path required")
	}
	if config == nil {
		return nil, errors.New("validation config required")
	}

	result := &ValidationResult{
		AssetPath:        assetPath,
		ComplianceChecks: make(map[string]bool),
		Timestamp:        time.Now(),
	}

	// Check if file exists
	if _, err := os.Stat(assetPath); err != nil {
		result.Errors = append(result.Errors, ValidationError{
			Code:     "FILE_NOT_FOUND",
			Message:  fmt.Sprintf("Asset file not found: %s", assetPath),
			Severity: "error",
		})
		result.Valid = false
		return result, nil
	}

	// Extract and validate metrics
	metrics, err := v.extractMetrics(assetPath)
	if err != nil {
		result.Errors = append(result.Errors, ValidationError{
			Code:     "METRICS_EXTRACTION_FAILED",
			Message:  fmt.Sprintf("Failed to extract asset metrics: %v", err),
			Severity: "error",
		})
		result.Valid = false
		return result, nil
	}
	result.Metrics = metrics

	// Run validation checks
	v.checkFileSize(result, config)
	v.checkFrameCount(result, config)
	v.checkFrameRate(result, config)
	v.checkTransparency(result, config)
	v.checkDimensions(result, config)
	v.checkFormat(result, assetPath)

	// Determine overall validity
	result.Valid = len(result.Errors) == 0

	return result, nil
}

// ValidateCharacterSet checks all assets for a character archetype.
func (v *assetValidator) ValidateCharacterSet(ctx context.Context, characterDir string, config *CharacterConfig) (*CharacterValidationResult, error) {
	if characterDir == "" {
		return nil, errors.New("character directory required")
	}
	if config == nil {
		return nil, errors.New("character config required")
	}

	result := &CharacterValidationResult{
		Character:    config.Character.Archetype,
		AssetResults: make(map[string]*ValidationResult),
	}

	// Check for required animation states
	animationsDir := filepath.Join(characterDir, "animations")
	for _, state := range config.States {
		gifPath := filepath.Join(animationsDir, state+".gif")

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if _, err := os.Stat(gifPath); os.IsNotExist(err) {
			result.MissingStates = append(result.MissingStates, state)
			continue
		}

		// Validate individual asset
		assetResult, err := v.ValidateAsset(ctx, gifPath, config.Validation)
		if err != nil {
			return nil, fmt.Errorf("validate asset %s: %w", state, err)
		}
		result.AssetResults[state] = assetResult
	}

	// Check style consistency across assets
	if config.Validation.StyleConsistency {
		consistency := v.checkStyleConsistency(result.AssetResults)
		result.StyleConsistency = consistency
	}

	// Create overall result
	result.Overall = v.createOverallResult(result)
	result.Valid = result.Overall.Valid && len(result.MissingStates) == 0

	return result, nil
}

// ValidateBatch checks multiple character sets in parallel.
func (v *assetValidator) ValidateBatch(ctx context.Context, configs []*CharacterConfig) (*BatchValidationResult, error) {
	if len(configs) == 0 {
		return nil, errors.New("no character configs provided")
	}

	startTime := time.Now()
	result := &BatchValidationResult{
		Characters: make(map[string]*CharacterValidationResult),
	}

	// Process characters in parallel
	type charResult struct {
		archetype string
		result    *CharacterValidationResult
		err       error
	}

	resultChan := make(chan charResult, len(configs))

	for _, config := range configs {
		go func(cfg *CharacterConfig) {
			characterDir := cfg.Deployment.OutputDir
			result, err := v.ValidateCharacterSet(ctx, characterDir, cfg)
			resultChan <- charResult{
				archetype: cfg.Character.Archetype,
				result:    result,
				err:       err,
			}
		}(config)
	}

	// Collect results
	for i := 0; i < len(configs); i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case res := <-resultChan:
			if res.err != nil {
				return nil, fmt.Errorf("validate character %s: %w", res.archetype, res.err)
			}
			result.Characters[res.archetype] = res.result
		}
	}

	// Generate summary
	result.Summary = v.generateSummary(result.Characters)
	result.OverallValid = result.Summary.InvalidAssets == 0
	result.ProcessingTime = time.Since(startTime)

	return result, nil
}

// extractMetrics extracts measurable properties from a GIF asset.
func (v *assetValidator) extractMetrics(assetPath string) (*AssetMetrics, error) {
	file, err := os.Open(assetPath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	// Get file size
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat file: %w", err)
	}

	// Decode GIF
	gifData, err := gif.DecodeAll(file)
	if err != nil {
		return nil, fmt.Errorf("decode GIF: %w", err)
	}

	if len(gifData.Image) == 0 {
		return nil, errors.New("GIF has no frames")
	}

	// Calculate metrics
	bounds := gifData.Image[0].Bounds()
	frameCount := len(gifData.Image)

	// Calculate duration and frame rate
	totalDelay := 0
	for _, delay := range gifData.Delay {
		totalDelay += delay
	}
	duration := time.Duration(totalDelay) * time.Millisecond * 10 // GIF delay is in 1/100s
	frameRate := 0.0
	if duration > 0 {
		frameRate = float64(frameCount) / duration.Seconds()
	}

	// Check transparency
	hasTransparency := v.checkGIFTransparency(gifData)

	// Count colors (approximate)
	colors := v.countColors(gifData)

	return &AssetMetrics{
		FileSize:        stat.Size(),
		Dimensions:      [2]int{bounds.Dx(), bounds.Dy()},
		FrameCount:      frameCount,
		Duration:        duration,
		FrameRate:       frameRate,
		Colors:          colors,
		HasTransparency: hasTransparency,
	}, nil
}

// checkFileSize validates file size against requirements.
func (v *assetValidator) checkFileSize(result *ValidationResult, config *ValidationConfig) {
	if result.Metrics.FileSize > int64(config.MaxFileSize) {
		result.Errors = append(result.Errors, ValidationError{
			Code:     "FILE_SIZE_EXCEEDED",
			Message:  "File size exceeds maximum allowed",
			Severity: "error",
			Expected: fmt.Sprintf("%d bytes", config.MaxFileSize),
			Actual:   fmt.Sprintf("%d bytes", result.Metrics.FileSize),
		})
		result.ComplianceChecks["file_size"] = false
	} else {
		result.ComplianceChecks["file_size"] = true
	}
}

// checkFrameCount validates frame count against requirements.
func (v *assetValidator) checkFrameCount(result *ValidationResult, config *ValidationConfig) {
	frameCount := result.Metrics.FrameCount
	if frameCount < 4 || frameCount > 8 {
		result.Errors = append(result.Errors, ValidationError{
			Code:     "INVALID_FRAME_COUNT",
			Message:  "Frame count must be between 4 and 8",
			Severity: "error",
			Expected: "4-8 frames",
			Actual:   fmt.Sprintf("%d frames", frameCount),
		})
		result.ComplianceChecks["frame_count"] = false
	} else {
		result.ComplianceChecks["frame_count"] = true
	}
}

// checkFrameRate validates frame rate against requirements.
func (v *assetValidator) checkFrameRate(result *ValidationResult, config *ValidationConfig) {
	frameRate := result.Metrics.FrameRate
	if frameRate < float64(config.MinFrameRate) {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Code:       "LOW_FRAME_RATE",
			Message:    fmt.Sprintf("Frame rate %.1f fps is below recommended minimum %d fps", frameRate, config.MinFrameRate),
			Suggestion: "Consider increasing frame rate for smoother animation",
		})
		result.ComplianceChecks["frame_rate"] = false
	} else {
		result.ComplianceChecks["frame_rate"] = true
	}
}

// checkTransparency validates transparency support.
func (v *assetValidator) checkTransparency(result *ValidationResult, config *ValidationConfig) {
	if config.TransparencyRequired && !result.Metrics.HasTransparency {
		result.Errors = append(result.Errors, ValidationError{
			Code:     "TRANSPARENCY_REQUIRED",
			Message:  "Asset must support transparency for desktop overlay",
			Severity: "error",
		})
		result.ComplianceChecks["transparency"] = false
	} else {
		result.ComplianceChecks["transparency"] = true
	}
}

// checkDimensions validates image dimensions.
func (v *assetValidator) checkDimensions(result *ValidationResult, config *ValidationConfig) {
	width, height := result.Metrics.Dimensions[0], result.Metrics.Dimensions[1]

	if width < 64 || width > 256 || height < 64 || height > 256 {
		result.Errors = append(result.Errors, ValidationError{
			Code:     "INVALID_DIMENSIONS",
			Message:  "Dimensions must be between 64x64 and 256x256 pixels",
			Severity: "error",
			Expected: "64x64 to 256x256",
			Actual:   fmt.Sprintf("%dx%d", width, height),
		})
		result.ComplianceChecks["dimensions"] = false
	} else {
		result.ComplianceChecks["dimensions"] = true
	}
}

// checkFormat validates file format.
func (v *assetValidator) checkFormat(result *ValidationResult, assetPath string) {
	if !strings.HasSuffix(strings.ToLower(assetPath), ".gif") {
		result.Errors = append(result.Errors, ValidationError{
			Code:     "INVALID_FORMAT",
			Message:  "Asset must be in GIF format",
			Severity: "error",
			Expected: ".gif",
			Actual:   filepath.Ext(assetPath),
		})
		result.ComplianceChecks["format"] = false
	} else {
		result.ComplianceChecks["format"] = true
	}
}

// checkGIFTransparency checks if a GIF has transparent pixels.
func (v *assetValidator) checkGIFTransparency(gifData *gif.GIF) bool {
	for _, img := range gifData.Image {
		if img.Palette != nil && len(img.Palette) > 0 {
			for _, color := range img.Palette {
				_, _, _, a := color.RGBA()
				if a == 0 {
					return true
				}
			}
		}
	}
	return false
}

// countColors approximates the color count in a GIF.
func (v *assetValidator) countColors(gifData *gif.GIF) int {
	if len(gifData.Image) == 0 {
		return 0
	}

	colorMap := make(map[uint32]bool)
	for _, img := range gifData.Image {
		bounds := img.Bounds()
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				r, g, b, a := img.At(x, y).RGBA()
				color := (r&0xFF00)<<16 | (g&0xFF00)<<8 | (b & 0xFF00) | (a >> 8)
				colorMap[color] = true
			}
		}
	}

	return len(colorMap)
}

// checkStyleConsistency analyzes style consistency across character assets.
func (v *assetValidator) checkStyleConsistency(assetResults map[string]*ValidationResult) *StyleConsistencyResult {
	if len(assetResults) < 2 {
		return &StyleConsistencyResult{
			Consistent: true,
			Score:      1.0,
		}
	}

	result := &StyleConsistencyResult{
		Consistent: true,
		Score:      1.0,
	}

	// Simple consistency check based on dimensions and color count
	var refDimensions [2]int
	var refColors int
	first := true

	for state, assetResult := range assetResults {
		if !assetResult.Valid {
			continue
		}

		if first {
			refDimensions = assetResult.Metrics.Dimensions
			refColors = assetResult.Metrics.Colors
			first = false
			continue
		}

		// Check dimension consistency
		if assetResult.Metrics.Dimensions != refDimensions {
			result.Inconsistencies = append(result.Inconsistencies, StyleInconsistency{
				States: []string{state},
				Issue: fmt.Sprintf("Inconsistent dimensions: %dx%d vs reference %dx%d",
					assetResult.Metrics.Dimensions[0], assetResult.Metrics.Dimensions[1],
					refDimensions[0], refDimensions[1]),
				Severity: "major",
			})
			result.Score -= 0.3
		}

		// Check color consistency (allow some variation)
		colorDiff := float64(abs(assetResult.Metrics.Colors-refColors)) / float64(refColors)
		if colorDiff > 0.5 { // More than 50% difference
			result.Inconsistencies = append(result.Inconsistencies, StyleInconsistency{
				States: []string{state},
				Issue: fmt.Sprintf("Inconsistent color count: %d vs reference %d",
					assetResult.Metrics.Colors, refColors),
				Severity: "minor",
			})
			result.Score -= 0.1
		}
	}

	result.Consistent = result.Score > 0.7

	if !result.Consistent {
		result.Recommendations = []string{
			"Ensure all animation states use consistent dimensions",
			"Use consistent color palettes across all states",
			"Apply the same art style parameters to all generations",
		}
	}

	return result
}

// createOverallResult creates an aggregate result for a character set.
func (v *assetValidator) createOverallResult(charResult *CharacterValidationResult) *ValidationResult {
	overall := &ValidationResult{
		AssetPath:        charResult.Character,
		ComplianceChecks: make(map[string]bool),
		Timestamp:        time.Now(),
	}

	// Aggregate metrics and compliance checks
	totalErrors := 0
	totalWarnings := 0
	validAssets := 0

	for _, assetResult := range charResult.AssetResults {
		totalErrors += len(assetResult.Errors)
		totalWarnings += len(assetResult.Warnings)

		if assetResult.Valid {
			validAssets++
		}

		// Aggregate compliance checks
		for check, passed := range assetResult.ComplianceChecks {
			if existing, exists := overall.ComplianceChecks[check]; !exists || !existing {
				overall.ComplianceChecks[check] = passed
			}
		}
	}

	// Add missing states as errors
	for _, missingState := range charResult.MissingStates {
		overall.Errors = append(overall.Errors, ValidationError{
			Code:     "MISSING_STATE",
			Message:  fmt.Sprintf("Required animation state not found: %s", missingState),
			Severity: "error",
			Field:    "states",
		})
		totalErrors++
	}

	// Add style consistency issues
	if charResult.StyleConsistency != nil && !charResult.StyleConsistency.Consistent {
		for _, inconsistency := range charResult.StyleConsistency.Inconsistencies {
			severity := "warning"
			if inconsistency.Severity == "major" {
				severity = "error"
				totalErrors++
			} else {
				totalWarnings++
			}

			overall.Errors = append(overall.Errors, ValidationError{
				Code:     "STYLE_INCONSISTENCY",
				Message:  inconsistency.Issue,
				Severity: severity,
				Field:    "style",
			})
		}
	}

	overall.Valid = totalErrors == 0

	return overall
}

// generateSummary creates aggregate statistics for batch validation.
func (v *assetValidator) generateSummary(characters map[string]*CharacterValidationResult) *ValidationSummary {
	summary := &ValidationSummary{}

	for _, charResult := range characters {
		for _, assetResult := range charResult.AssetResults {
			summary.TotalAssets++
			if assetResult.Valid {
				summary.ValidAssets++
			} else {
				summary.InvalidAssets++
			}
			summary.ErrorCount += len(assetResult.Errors)
			summary.WarningCount += len(assetResult.Warnings)
		}
	}

	if summary.TotalAssets > 0 {
		summary.SuccessRate = float64(summary.ValidAssets) / float64(summary.TotalAssets)
	}

	return summary
}

// abs returns the absolute value of an integer.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
