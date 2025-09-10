// Package embedding provides functionality for creating character-specific binaries
// with embedded assets using Go's standard library capabilities.
package embedding

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/gif"
	"log"
	"os"
	"path/filepath"
	"text/template"
)

// mainTemplate generates a standalone Go application with embedded character assets
// With no internal packages, this imports directly from the main module without replace directives
const mainTemplate = `package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"image/gif"
	"log"

	"fyne.io/fyne/v2/app"
	
	"github.com/opd-ai/desktop-companion/lib/character"
	"github.com/opd-ai/desktop-companion/lib/monitoring"
	"github.com/opd-ai/desktop-companion/lib/ui"
)

var (
	version = flag.Bool("version", false, "Show version information")
)

// Embedded character data - JSON configuration
var embeddedCharacterData = ` + "`" + `{{.CharacterJSON}}` + "`" + `

// Embedded animations - binary GIF data encoded as Go byte slices
// This follows the "library-first" approach using standard library byte embedding
var embeddedAnimations = map[string][]byte{
{{range $name, $data := .Animations}}	"{{$name}}": {
{{range $i, $b := $data}}{{if gt $i 0}}, {{end}}{{printf "0x%02x" $b}}{{end}}},
{{end}}}

// Application metadata
const appVersion = "1.0.0-{{.CharacterName}}"
const appID = "com.opdai.{{.CharacterName}}-companion"

// showVersionInfo displays application version information.
func showVersionInfo() {
	fmt.Printf("Desktop Companion ({{.CharacterName}}) v%s\n", appVersion)
	fmt.Println("Built with Go and Fyne - Cross-platform desktop pet")
}

func main() {
	flag.Parse()

	if *version {
		showVersionInfo()
		return
	}
	// Parse embedded character data using standard library JSON
	var card character.CharacterCard
	if err := json.Unmarshal([]byte(embeddedCharacterData), &card); err != nil {
		log.Fatalf("Failed to parse embedded character data: %v", err)
	}

	// Create embedded animation manager
	animManager, err := createEmbeddedAnimationManager()
	if err != nil {
		log.Fatalf("Failed to create animation manager: %v", err)
	}

	// Initialize performance profiler (following project standards)
	profiler := monitoring.NewProfiler(50)
	if err := profiler.Start("", "", false); err != nil {
		log.Fatalf("Failed to start profiler: %v", err)
	}
	defer profiler.Stop("", false)

	// Create Fyne application with character-specific ID
	myApp := app.NewWithID(appID)
	
	// Create character with embedded assets (no filesystem dependencies)
	char, err := character.NewEmbedded(&card, animManager)
	if err != nil {
		log.Fatalf("Failed to create character: %v", err)
	}

	// Create and show UI (reusing existing UI components)
	window := ui.NewDesktopWindow(myApp, char, false, profiler, false, false, nil, false, false, false)
	window.Show()
	myApp.Run()
}

// createEmbeddedAnimationManager creates an animation manager from embedded data
// Uses standard library image/gif package for decoding
func createEmbeddedAnimationManager() (*character.AnimationManager, error) {
	animManager := character.NewAnimationManager()
	
	for name, data := range embeddedAnimations {
		reader := bytes.NewReader(data)
		gifData, err := gif.DecodeAll(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to decode embedded animation %s: %w", name, err)
		}
		
		if err := animManager.LoadEmbeddedAnimation(name, gifData); err != nil {
			return nil, fmt.Errorf("failed to load embedded animation %s: %w", name, err)
		}
	}
	
	return animManager, nil
}
`

// TemplateData holds the data for generating the embedded character application
type TemplateData struct {
	CharacterName string
	CharacterJSON string
	Animations    map[string][]byte
}

// GenerateEmbeddedCharacter creates a standalone character application with embedded assets
func GenerateEmbeddedCharacter(characterName, outputDir string) error {
	// Load character card using standard library JSON parsing
	cardPath := fmt.Sprintf("assets/characters/%s/character.json", characterName)
	cardData, err := os.ReadFile(cardPath)
	if err != nil {
		return fmt.Errorf("failed to read character card %s: %w", cardPath, err)
	}

	// Parse character card to extract animation paths
	var card map[string]interface{}
	if err := json.Unmarshal(cardData, &card); err != nil {
		return fmt.Errorf("failed to parse character card: %w", err)
	}

	// Load and embed animations using standard library
	animations, err := LoadAnimations(card, filepath.Dir(cardPath))
	if err != nil {
		return fmt.Errorf("failed to load animations: %w", err)
	}

	// Generate standalone application
	if err := generateEmbeddedApp(characterName, string(cardData), animations, outputDir); err != nil {
		return fmt.Errorf("failed to generate embedded application: %w", err)
	}

	fmt.Printf("✓ Generated embedded character application for %s in %s\n", characterName, outputDir)
	return nil
}

// LoadAnimations loads all GIF animations referenced in the character card
// Returns map of animation name to binary GIF data
func LoadAnimations(card map[string]interface{}, characterDir string) (map[string][]byte, error) {
	animations := make(map[string][]byte)

	// Extract animations section from character card
	animsInterface, ok := card["animations"]
	if !ok {
		return animations, nil // No animations defined
	}

	anims, ok := animsInterface.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("animations field must be an object")
	}

	// Load each animation file
	for animName, animPathInterface := range anims {
		animPath, ok := animPathInterface.(string)
		if !ok {
			log.Printf("Warning: Animation %s path is not a string, skipping", animName)
			continue
		}

		fullPath := filepath.Join(characterDir, animPath)
		animData, err := os.ReadFile(fullPath)
		if err != nil {
			log.Printf("Warning: Failed to read animation %s at %s: %v", animName, fullPath, err)
			continue
		}

		// Validate that it's a valid GIF
		if !IsValidGIF(animData) {
			log.Printf("Warning: File %s is not a valid GIF, skipping", fullPath)
			continue
		}

		animations[animName] = animData
		fmt.Printf("  ✓ Embedded animation: %s (%d bytes)\n", animName, len(animData))
	}

	if len(animations) == 0 {
		return nil, fmt.Errorf("no valid animations found for character")
	}

	return animations, nil
}

// IsValidGIF checks if the provided data is a valid GIF file
// Uses standard library image/gif for validation
func IsValidGIF(data []byte) bool {
	reader := bytes.NewReader(data)
	_, err := gif.DecodeAll(reader)
	return err == nil
}

// generateEmbeddedApp creates the embedded Go application file
func generateEmbeddedApp(characterName, characterJSON string, animations map[string][]byte, outputDir string) error {
	// Create output directory
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Prepare template data
	data := TemplateData{
		CharacterName: characterName,
		CharacterJSON: characterJSON,
		Animations:    animations,
	}

	// Generate main.go file
	tmpl := template.Must(template.New("main").Parse(mainTemplate))

	outputFile := filepath.Join(outputDir, "main.go")
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Create go.mod file for the embedded character - simplified approach
	if err := generateGoMod(characterName, outputDir); err != nil {
		return fmt.Errorf("failed to generate go.mod: %w", err)
	}

	// Copy go.sum for dependency validation
	if err := copyGoSum(outputDir); err != nil {
		return fmt.Errorf("failed to copy go.sum: %w", err)
	}

	fmt.Printf("  → To build: cd %s && go build -o %s-companion main.go\n", outputDir, characterName)
	fmt.Printf("  → Simplified approach: Only one replace directive needed (no complex module copying)\n")

	return nil
}

// generateGoMod creates a simplified go.mod file for the embedded character
// Uses a minimal replace directive to point to the local module
func generateGoMod(characterName, outputDir string) error {
	// Get the project root for the replace directive
	projectRoot, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("failed to get project root path: %w", err)
	}

	// Simple go.mod with minimal replace directive - much cleaner than complex module copying
	goModContent := fmt.Sprintf(`module github.com/opd-ai/desktop-companion/cmd/%s-embedded

go 1.21

require (
	fyne.io/fyne/v2 v2.4.5
	github.com/opd-ai/desktop-companion v0.0.0-00010101000000-000000000000
	github.com/jdkato/prose/v2 v2.0.0
	github.com/mmcdole/gofeed v1.3.0
)

// Single replace directive - much simpler than the old approach
replace github.com/opd-ai/desktop-companion => %s
`, characterName, projectRoot)

	goModFile := filepath.Join(outputDir, "go.mod")
	if err := os.WriteFile(goModFile, []byte(goModContent), 0o644); err != nil {
		return fmt.Errorf("failed to write go.mod file: %w", err)
	}

	fmt.Printf("  ✓ Generated simplified go.mod with minimal replace directive\n")
	return nil
}

// copyGoSum copies the main project's go.sum to the embedded character directory
// This ensures all dependencies are properly validated
func copyGoSum(outputDir string) error {
	sourcePath := "go.sum"
	destPath := filepath.Join(outputDir, "go.sum")

	// Check if source go.sum exists
	if _, err := os.Stat(sourcePath); err != nil {
		// If no go.sum exists, that's ok - it will be generated
		return nil
	}

	// Copy go.sum
	sourceData, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to read source go.sum: %w", err)
	}

	if err := os.WriteFile(destPath, sourceData, 0o644); err != nil {
		return fmt.Errorf("failed to write go.sum: %w", err)
	}

	fmt.Printf("  ✓ Copied go.sum for dependency validation\n")
	return nil
}
