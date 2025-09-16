package assets

// validation.go implements comprehensive validation for generated assets
// including file size, image quality, animation consistency, and compatibility checks.

import (
	"fmt"
	"image"
	"image/gif"
	"os"
	"path/filepath"
	"strings"

	"github.com/opd-ai/desktop-companion/lib/character"
)

// AssetValidator provides comprehensive validation for generated character assets.
type AssetValidator struct {
	config *ValidationConfig
}

// ValidationConfig defines validation parameters.
type ValidationConfig struct {
	// File size limits
	MaxFileSizeKB int
	MinFileSizeKB int

	// Image dimension requirements
	MinWidth  int
	MaxWidth  int
	MinHeight int
	MaxHeight int

	// Animation requirements
	MinFrames int
	MaxFrames int
	MinFPS    int
	MaxFPS    int

	// Quality thresholds
	MinQualityScore float64

	// Compatibility checks
	RequireTransparency bool
	AllowedFormats      []string
}

// ValidationResult contains detailed validation results.
type ValidationResult struct {
	// Overall validation status
	Valid bool
	Score float64

	// File validation
	FileExists bool
	FileSizeOK bool
	FileSizeKB int
	FormatOK   bool
	Format     string

	// Image validation
	DimensionsOK bool
	Width        int
	Height       int
	HasAlpha     bool

	// Animation validation (for GIFs)
	IsAnimation  bool
	FrameCount   int
	FramesOK     bool
	EstimatedFPS float64
	FPSOK        bool
	LoopingOK    bool

	// Quality metrics
	QualityMetrics QualityMetrics

	// Issues found
	Warnings []string
	Errors   []string
}

// QualityMetrics contains computed quality metrics.
type QualityMetrics struct {
	// Visual quality
	Sharpness  float64
	Contrast   float64
	Brightness float64
	ColorRange float64

	// Animation quality (for GIFs)
	FrameStability   float64
	MotionSmoothness float64
	LoopSeamlessness float64

	// Compression efficiency
	CompressionRatio float64
	FileEfficiency   float64
}

// NewAssetValidator creates a new asset validator with the given configuration.
func NewAssetValidator(config *ValidationConfig) *AssetValidator {
	if config == nil {
		config = DefaultValidationConfig()
	}
	return &AssetValidator{config: config}
}

// DefaultValidationConfig returns sensible default validation settings.
func DefaultValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		MaxFileSizeKB:       500,
		MinFileSizeKB:       10,
		MinWidth:            64,
		MaxWidth:            512,
		MinHeight:           64,
		MaxHeight:           512,
		MinFrames:           2,
		MaxFrames:           30,
		MinFPS:              5,
		MaxFPS:              30,
		MinQualityScore:     0.7,
		RequireTransparency: true,
		AllowedFormats:      []string{"gif", "png", "webp"},
	}
}

// ValidateAsset performs comprehensive validation of a single asset file.
func (v *AssetValidator) ValidateAsset(assetPath string) (*ValidationResult, error) {
	result := &ValidationResult{
		Warnings: []string{},
		Errors:   []string{},
	}

	// Check file existence
	info, err := os.Stat(assetPath)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("File not found: %v", err))
		return result, nil
	}
	result.FileExists = true

	// Check file size
	fileSizeKB := int(info.Size() / 1024)
	result.FileSizeKB = fileSizeKB
	if fileSizeKB > v.config.MaxFileSizeKB {
		result.Errors = append(result.Errors, fmt.Sprintf("File too large: %dKB > %dKB", fileSizeKB, v.config.MaxFileSizeKB))
	} else if fileSizeKB < v.config.MinFileSizeKB {
		result.Warnings = append(result.Warnings, fmt.Sprintf("File very small: %dKB < %dKB", fileSizeKB, v.config.MinFileSizeKB))
	} else {
		result.FileSizeOK = true
	}

	// Check file format
	ext := strings.ToLower(filepath.Ext(assetPath)[1:])
	result.Format = ext
	for _, allowed := range v.config.AllowedFormats {
		if ext == allowed {
			result.FormatOK = true
			break
		}
	}
	if !result.FormatOK {
		result.Errors = append(result.Errors, fmt.Sprintf("Unsupported format: %s", ext))
	}

	// Validate image content based on format
	switch ext {
	case "gif":
		if err := v.validateGIF(assetPath, result); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("GIF validation failed: %v", err))
		}
	case "png":
		if err := v.validatePNG(assetPath, result); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("PNG validation failed: %v", err))
		}
	default:
		result.Warnings = append(result.Warnings, fmt.Sprintf("Cannot validate format: %s", ext))
	}

	// Calculate overall score and validity
	result.Score = v.calculateQualityScore(result)
	result.Valid = len(result.Errors) == 0 && result.Score >= v.config.MinQualityScore

	return result, nil
}

// validateGIF performs GIF-specific validation.
func (v *AssetValidator) validateGIF(assetPath string, result *ValidationResult) error {
	file, err := os.Open(assetPath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	gifImg, err := gif.DecodeAll(file)
	if err != nil {
		return fmt.Errorf("decode GIF: %w", err)
	}

	result.IsAnimation = true
	result.FrameCount = len(gifImg.Image)

	// Validate frame count
	if result.FrameCount < v.config.MinFrames {
		result.Errors = append(result.Errors, fmt.Sprintf("Too few frames: %d < %d", result.FrameCount, v.config.MinFrames))
	} else if result.FrameCount > v.config.MaxFrames {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Many frames: %d > %d", result.FrameCount, v.config.MaxFrames))
	} else {
		result.FramesOK = true
	}

	// Get dimensions from first frame
	if len(gifImg.Image) > 0 {
		bounds := gifImg.Image[0].Bounds()
		result.Width = bounds.Dx()
		result.Height = bounds.Dy()

		// Validate dimensions
		if result.Width < v.config.MinWidth || result.Width > v.config.MaxWidth {
			result.Errors = append(result.Errors, fmt.Sprintf("Invalid width: %d (must be %d-%d)", result.Width, v.config.MinWidth, v.config.MaxWidth))
		} else if result.Height < v.config.MinHeight || result.Height > v.config.MaxHeight {
			result.Errors = append(result.Errors, fmt.Sprintf("Invalid height: %d (must be %d-%d)", result.Height, v.config.MinHeight, v.config.MaxHeight))
		} else {
			result.DimensionsOK = true
		}

		// Check for transparency (simplified check)
		result.HasAlpha = v.hasTransparency(gifImg.Image[0])
		if v.config.RequireTransparency && !result.HasAlpha {
			result.Warnings = append(result.Warnings, "No transparency detected")
		}
	}

	// Estimate FPS from delays
	if len(gifImg.Delay) > 0 {
		totalDelay := 0
		for _, delay := range gifImg.Delay {
			totalDelay += delay
		}
		if totalDelay > 0 {
			result.EstimatedFPS = float64(len(gifImg.Delay)*100) / float64(totalDelay) // GIF delays are in 1/100 seconds
			if result.EstimatedFPS < float64(v.config.MinFPS) || result.EstimatedFPS > float64(v.config.MaxFPS) {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Unusual FPS: %.1f (recommended %d-%d)", result.EstimatedFPS, v.config.MinFPS, v.config.MaxFPS))
			} else {
				result.FPSOK = true
			}
		}
	}

	// Check looping (GIF LoopCount: 0 = infinite, >0 = finite)
	result.LoopingOK = gifImg.LoopCount == 0 // Most desktop companions should loop infinitely
	if !result.LoopingOK {
		result.Warnings = append(result.Warnings, fmt.Sprintf("GIF loops %d times (infinite looping recommended)", gifImg.LoopCount))
	}

	return nil
}

// validatePNG performs PNG-specific validation.
func (v *AssetValidator) validatePNG(assetPath string, result *ValidationResult) error {
	file, err := os.Open(assetPath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("decode PNG: %w", err)
	}

	result.IsAnimation = false
	result.FrameCount = 1
	result.FramesOK = true

	// Get dimensions
	bounds := img.Bounds()
	result.Width = bounds.Dx()
	result.Height = bounds.Dy()

	// Validate dimensions
	if result.Width < v.config.MinWidth || result.Width > v.config.MaxWidth {
		result.Errors = append(result.Errors, fmt.Sprintf("Invalid width: %d (must be %d-%d)", result.Width, v.config.MinWidth, v.config.MaxWidth))
	} else if result.Height < v.config.MinHeight || result.Height > v.config.MaxHeight {
		result.Errors = append(result.Errors, fmt.Sprintf("Invalid height: %d (must be %d-%d)", result.Height, v.config.MinHeight, v.config.MaxHeight))
	} else {
		result.DimensionsOK = true
	}

	// Check for transparency
	result.HasAlpha = v.hasImageTransparency(img)
	if v.config.RequireTransparency && !result.HasAlpha {
		result.Warnings = append(result.Warnings, "No transparency detected")
	}

	return nil
}

// hasTransparency checks if a paletted image (from GIF) has transparency.
func (v *AssetValidator) hasTransparency(img image.Image) bool {
	// This is a simplified transparency check
	// A more comprehensive check would examine the actual palette
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y += 10 { // Sample every 10th pixel for performance
		for x := bounds.Min.X; x < bounds.Max.X; x += 10 {
			_, _, _, a := img.At(x, y).RGBA()
			if a < 65535 { // Not fully opaque
				return true
			}
		}
	}
	return false
}

// hasImageTransparency checks if a general image has transparency.
func (v *AssetValidator) hasImageTransparency(img image.Image) bool {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y += 10 { // Sample for performance
		for x := bounds.Min.X; x < bounds.Max.X; x += 10 {
			_, _, _, a := img.At(x, y).RGBA()
			if a < 65535 { // Not fully opaque
				return true
			}
		}
	}
	return false
}

// calculateQualityScore computes an overall quality score based on validation results.
func (v *AssetValidator) calculateQualityScore(result *ValidationResult) float64 {
	score := 0.0
	maxScore := 0.0

	// File validation (20% of score)
	if result.FileExists {
		score += 0.1
	}
	if result.FileSizeOK {
		score += 0.05
	}
	if result.FormatOK {
		score += 0.05
	}
	maxScore += 0.2

	// Dimensions validation (20% of score)
	if result.DimensionsOK {
		score += 0.2
	}
	maxScore += 0.2

	// Animation validation (30% of score for animated content)
	if result.IsAnimation {
		if result.FramesOK {
			score += 0.1
		}
		if result.FPSOK {
			score += 0.1
		}
		if result.LoopingOK {
			score += 0.1
		}
		maxScore += 0.3
	} else {
		maxScore += 0.3 // Static images get full score for this section
		score += 0.3
	}

	// Transparency (10% of score)
	if !v.config.RequireTransparency || result.HasAlpha {
		score += 0.1
	}
	maxScore += 0.1

	// Error penalty (20% of score)
	errorPenalty := float64(len(result.Errors)) * 0.05
	score += 0.2 - errorPenalty
	if score < 0 {
		score = 0
	}
	maxScore += 0.2

	return score / maxScore
}

// ValidateCharacterAssets validates all assets for a character configuration.
func (v *AssetValidator) ValidateCharacterAssets(card *character.CharacterCard, basePath string) (map[string]*ValidationResult, error) {
	results := make(map[string]*ValidationResult)

	// Validate existing animation files
	for state, relativePath := range card.Animations {
		fullPath := filepath.Join(basePath, relativePath)
		result, err := v.ValidateAsset(fullPath)
		if err != nil {
			return nil, fmt.Errorf("validate %s: %w", state, err)
		}
		results[state] = result
	}

	return results, nil
}

// GenerateValidationReport creates a human-readable validation report.
func (v *AssetValidator) GenerateValidationReport(results map[string]*ValidationResult) string {
	var report strings.Builder

	totalAssets := len(results)
	validAssets := 0
	totalScore := 0.0

	report.WriteString("Asset Validation Report\n")
	report.WriteString("=======================\n\n")

	for state, result := range results {
		report.WriteString(fmt.Sprintf("Animation: %s\n", state))
		report.WriteString(fmt.Sprintf("  Valid: %t\n", result.Valid))
		report.WriteString(fmt.Sprintf("  Score: %.2f\n", result.Score))
		report.WriteString(fmt.Sprintf("  Size: %dKB (%dx%d)\n", result.FileSizeKB, result.Width, result.Height))

		if result.IsAnimation {
			report.WriteString(fmt.Sprintf("  Animation: %d frames at %.1f FPS\n", result.FrameCount, result.EstimatedFPS))
		}

		if len(result.Errors) > 0 {
			report.WriteString("  Errors:\n")
			for _, err := range result.Errors {
				report.WriteString(fmt.Sprintf("    - %s\n", err))
			}
		}

		if len(result.Warnings) > 0 {
			report.WriteString("  Warnings:\n")
			for _, warning := range result.Warnings {
				report.WriteString(fmt.Sprintf("    - %s\n", warning))
			}
		}

		report.WriteString("\n")

		if result.Valid {
			validAssets++
		}
		totalScore += result.Score
	}

	avgScore := totalScore / float64(totalAssets)
	report.WriteString(fmt.Sprintf("Summary: %d/%d assets valid (%.1f%% success rate)\n", validAssets, totalAssets, float64(validAssets)/float64(totalAssets)*100))
	report.WriteString(fmt.Sprintf("Average Quality Score: %.2f\n", avgScore))

	return report.String()
}
