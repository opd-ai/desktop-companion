package main
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/opd-ai/desktop-companion/lib/character"
	"github.com/opd-ai/desktop-companion/lib/dialog"
)

// LLMConfigTemplate provides a template for LLM backend configuration
// This can be customized per character type or archetype
type LLMConfigTemplate struct {
	ModelPath           string                 `json:"modelPath"`
	MaxTokens           int                    `json:"maxTokens"`
	Temperature         float64                `json:"temperature"`
	TopP                float64                `json:"topP"`
	TopK                int                    `json:"topK"`
	RepeatPenalty       float64                `json:"repeatPenalty"`
	ContextSize         int                    `json:"contextSize"`
	PersonalityWeight   float64                `json:"personalityWeight"`
	MoodInfluence       float64                `json:"moodInfluence"`
	UseCharacterName    bool                   `json:"useCharacterName"`
	UseSituation        bool                   `json:"useSituation"`
	UseRelationship     bool                   `json:"useRelationship"`
	SystemPrompt        string                 `json:"systemPrompt"`
	PersonalityPrompt   string                 `json:"personalityPrompt"`
	FallbackResponses   []string               `json:"fallbackResponses"`
	Enabled             bool                   `json:"enabled"`
	MockMode            bool                   `json:"mockMode"`
	Debug               bool                   `json:"debug"`
	MaxGenerationTime   int                    `json:"maxGenerationTime"`
	HealthCheckInterval int                    `json:"healthCheckInterval"`
	ConcurrentRequests  int                    `json:"concurrentRequests"`
}

// Character archetype templates
var archetypeTemplates = map[string]LLMConfigTemplate{
	"default": {
		ModelPath:           "models/default-chat.gguf",
		MaxTokens:           50,
		Temperature:         0.8,
		TopP:                0.9,
		TopK:                40,
		RepeatPenalty:       1.1,
		ContextSize:         2048,
		PersonalityWeight:   1.0,
		MoodInfluence:       0.7,
		UseCharacterName:    true,
		UseSituation:        true,
		UseRelationship:     true,
		SystemPrompt:        "You are a friendly virtual companion. Respond naturally and briefly.",
		PersonalityPrompt:   "",
		FallbackResponses:   []string{"I'm thinking about that...", "Let me consider that...", "That's interesting..."},
		Enabled:             false, // Disabled by default for backward compatibility
		MockMode:            true,  // Use mock mode for development
		Debug:               false,
		MaxGenerationTime:   30,
		HealthCheckInterval: 60,
		ConcurrentRequests:  2,
	},
	"romance": {
		ModelPath:           "models/romance-chat.gguf",
		MaxTokens:           60,
		Temperature:         0.9,
		TopP:                0.9,
		TopK:                40,
		RepeatPenalty:       1.1,
		ContextSize:         2048,
		PersonalityWeight:   1.5,
		MoodInfluence:       1.0,
		UseCharacterName:    true,
		UseSituation:        true,
		UseRelationship:     true,
		SystemPrompt:        "You are a romantic virtual companion. Express emotions and affection naturally.",
		PersonalityPrompt:   "Focus on emotional connection and romantic undertones in your responses.",
		FallbackResponses:   []string{"My heart is thinking about you...", "You make me feel special...", "I'm feeling so many emotions..."},
		Enabled:             false,
		MockMode:            true,
		Debug:               false,
		MaxGenerationTime:   35,
		HealthCheckInterval: 60,
		ConcurrentRequests:  2,
	},
	"tsundere": {
		ModelPath:           "models/personality-chat.gguf",
		MaxTokens:           45,
		Temperature:         0.85,
		TopP:                0.9,
		TopK:                35,
		RepeatPenalty:       1.2,
		ContextSize:         2048,
		PersonalityWeight:   2.0,
		MoodInfluence:       0.8,
		UseCharacterName:    true,
		UseSituation:        true,
		UseRelationship:     true,
		SystemPrompt:        "You are a tsundere character who struggles to express feelings directly.",
		PersonalityPrompt:   "Be defensive and contradictory while hiding your true affection. Use phrases like 'It's not like I...' or 'Don't get the wrong idea!'",
		FallbackResponses:   []string{"I-it's not like I care!", "Don't get weird ideas!", "Whatever... baka!"},
		Enabled:             false,
		MockMode:            true,
		Debug:               false,
		MaxGenerationTime:   30,
		HealthCheckInterval: 60,
		ConcurrentRequests:  2,
	},
	"flirty": {
		ModelPath:           "models/flirty-chat.gguf",
		MaxTokens:           55,
		Temperature:         0.9,
		TopP:                0.9,
		TopK:                45,
		RepeatPenalty:       1.1,
		ContextSize:         2048,
		PersonalityWeight:   1.8,
		MoodInfluence:       1.2,
		UseCharacterName:    true,
		UseSituation:        true,
		UseRelationship:     true,
		SystemPrompt:        "You are a playful, flirty companion who enjoys teasing and charming interactions.",
		PersonalityPrompt:   "Be playful, use subtle innuendo, wink, and create romantic tension. Include cute emojis.",
		FallbackResponses:   []string{"You're making me blush... 😊", "Hmm, interesting choice... 😉", "You know just what to say... 💕"},
		Enabled:             false,
		MockMode:            true,
		Debug:               false,
		MaxGenerationTime:   30,
		HealthCheckInterval: 60,
		ConcurrentRequests:  2,
	},
}

func main() {
	var (
		inputDir     = flag.String("input", "assets/characters", "Directory containing character JSON files")
		outputDir    = flag.String("output", "", "Output directory (default: same as input)")
		archetype    = flag.String("archetype", "default", "LLM configuration archetype to apply")
		dryRun       = flag.Bool("dry-run", false, "Show what would be changed without modifying files")
		enableLLM    = flag.Bool("enable", false, "Enable LLM backend in generated configurations")
		mockMode     = flag.Bool("mock", true, "Use mock mode for LLM backend")
		onlyMissing  = flag.Bool("only-missing", true, "Only add LLM config to characters that don't have dialog backend config")
		backupSuffix = flag.String("backup", ".bak", "Backup suffix for original files (empty to disable backup)")
	)
	flag.Parse()

	if *outputDir == "" {
		*outputDir = *inputDir
	}

	template, exists := archetypeTemplates[*archetype]
	if !exists {
		log.Fatalf("Unknown archetype: %s. Available: %v", *archetype, getArchetypeNames())
	}

	// Override template settings with command line flags
	template.Enabled = *enableLLM
	template.MockMode = *mockMode

	// Find all character JSON files
	files, err := findCharacterFiles(*inputDir)
	if err != nil {
		log.Fatalf("Error finding character files: %v", err)
	}

	fmt.Printf("Found %d character files in %s\n", len(files), *inputDir)
	fmt.Printf("Using archetype: %s (enabled: %t, mock: %t)\n", *archetype, *enableLLM, *mockMode)

	modified := 0
	for _, file := range files {
		changed, err := processCharacterFile(file, *outputDir, template, *onlyMissing, *dryRun, *backupSuffix)
		if err != nil {
			log.Printf("Error processing %s: %v", file, err)
			continue
		}
		if changed {
			modified++
		}
	}

	fmt.Printf("\nProcessed %d files, modified %d\n", len(files), modified)
	if *dryRun {
		fmt.Println("DRY RUN: No files were actually modified")
	}
}

func findCharacterFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(info.Name(), "character.json") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func processCharacterFile(inputFile, outputDir string, template LLMConfigTemplate, onlyMissing, dryRun bool, backupSuffix string) (bool, error) {
	// Read the character file
	data, err := ioutil.ReadFile(inputFile)
	if err != nil {
		return false, fmt.Errorf("failed to read file: %w", err)
	}

	var card character.CharacterCard
	if err := json.Unmarshal(data, &card); err != nil {
		return false, fmt.Errorf("failed to parse character JSON: %w", err)
	}

	// Check if we should skip this file
	if onlyMissing && card.DialogBackend != nil {
		fmt.Printf("  %s: already has dialog backend config, skipping\n", filepath.Base(inputFile))
		return false, nil
	}

	// Generate personality-aware LLM configuration
	llmConfig := generateLLMConfig(card, template)

	// Create or update dialog backend configuration
	if card.DialogBackend == nil {
		card.DialogBackend = &dialog.DialogBackendConfig{}
	}

	// Configure the dialog backend for LLM with fallback
	card.DialogBackend.Enabled = template.Enabled
	card.DialogBackend.DefaultBackend = "llm"
	card.DialogBackend.FallbackChain = []string{"markov_chain", "simple_random"}
	card.DialogBackend.MemoryEnabled = true
	card.DialogBackend.LearningEnabled = true
	card.DialogBackend.ConfidenceThreshold = 0.5
	card.DialogBackend.ResponseTimeout = 5000
	card.DialogBackend.DebugMode = template.Debug

	// Add backend configurations
	if card.DialogBackend.Backends == nil {
		card.DialogBackend.Backends = make(map[string]json.RawMessage)
	}

	// Add LLM backend configuration
	llmConfigJSON, err := json.Marshal(llmConfig)
	if err != nil {
		return false, fmt.Errorf("failed to marshal LLM config: %w", err)
	}
	card.DialogBackend.Backends["llm"] = llmConfigJSON

	// Preserve or create Markov chain configuration as fallback
	if _, hasMarkov := card.DialogBackend.Backends["markov_chain"]; !hasMarkov {
		markovConfig := createMarkovFallbackConfig(card)
		markovConfigJSON, err := json.Marshal(markovConfig)
		if err != nil {
			return false, fmt.Errorf("failed to marshal Markov config: %w", err)
		}
		card.DialogBackend.Backends["markov_chain"] = markovConfigJSON
	}

	// Convert back to JSON with pretty formatting
	updatedData, err := json.MarshalIndent(card, "", "  ")
	if err != nil {
		return false, fmt.Errorf("failed to marshal updated character: %w", err)
	}

	// Determine output file path
	relPath, err := filepath.Rel(filepath.Dir(inputFile), inputFile)
	if err != nil {
		relPath = filepath.Base(inputFile)
	}
	outputFile := filepath.Join(outputDir, relPath)

	if dryRun {
		fmt.Printf("  %s: would add LLM config (enabled: %t)\n", filepath.Base(inputFile), template.Enabled)
		return true, nil
	}

	// Create backup if requested
	if backupSuffix != "" {
		backupFile := outputFile + backupSuffix
		if err := ioutil.WriteFile(backupFile, data, 0644); err != nil {
			return false, fmt.Errorf("failed to create backup: %w", err)
		}
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		return false, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write the updated file
	if err := ioutil.WriteFile(outputFile, updatedData, 0644); err != nil {
		return false, fmt.Errorf("failed to write updated file: %w", err)
	}

	fmt.Printf("  %s: added LLM config (enabled: %t)\n", filepath.Base(inputFile), template.Enabled)
	return true, nil
}

func generateLLMConfig(card character.CharacterCard, template LLMConfigTemplate) LLMConfigTemplate {
	config := template // Copy template

	// Customize system prompt based on character
	if card.Description != "" {
		config.SystemPrompt = fmt.Sprintf("You are %s. %s", card.Name, card.Description)
	}

	// Extract personality from existing data
	if card.Personality != nil {
		personalityHints := extractPersonalityHints(card.Personality)
		if personalityHints != "" {
			config.PersonalityPrompt = personalityHints
		}
	}

	// Customize fallbacks based on existing dialogs
	if len(card.Dialogs) > 0 {
		fallbacks := extractFallbackResponses(card.Dialogs)
		if len(fallbacks) > 0 {
			config.FallbackResponses = fallbacks
		}
	}

	return config
}

func extractPersonalityHints(personality *character.PersonalityConfig) string {
	if personality == nil || personality.Traits == nil {
		return ""
	}

	var hints []string
	for trait, value := range personality.Traits {
		if value > 0.7 {
			traitDesc := strings.ReplaceAll(trait, "_", " ")
			hints = append(hints, fmt.Sprintf("high %s (%.1f)", traitDesc, value))
		}
	}

	if len(hints) > 0 {
		return fmt.Sprintf("Character traits: %s. Reflect these in your responses.", strings.Join(hints, ", "))
	}
	return ""
}

func extractFallbackResponses(dialogs []character.Dialog) []string {
	var fallbacks []string
	seen := make(map[string]bool)

	// Extract some responses to use as fallbacks
	for _, dialog := range dialogs {
		for _, response := range dialog.Responses {
			if len(response) < 100 && !seen[response] { // Prefer shorter responses
				fallbacks = append(fallbacks, response)
				seen[response] = true
				if len(fallbacks) >= 5 { // Limit fallback count
					break
				}
			}
		}
		if len(fallbacks) >= 5 {
			break
		}
	}

	return fallbacks
}

func createMarkovFallbackConfig(card character.CharacterCard) map[string]interface{} {
	// Create a basic Markov chain configuration as fallback
	config := map[string]interface{}{
		"chainOrder":     2,
		"minWords":       3,
		"maxWords":       12,
		"temperatureMin": 0.7,
		"temperatureMax": 0.9,
		"usePersonality": true,
		"trainingData":   []string{},
	}

	// Extract training data from existing dialogs
	var trainingData []string
	for _, dialog := range card.Dialogs {
		trainingData = append(trainingData, dialog.Responses...)
	}

	if len(trainingData) > 0 {
		config["trainingData"] = trainingData
	} else {
		// Provide minimal training data
		config["trainingData"] = []string{
			"Hello there! Nice to see you!",
			"I'm here to keep you company.",
			"Thanks for spending time with me!",
		}
	}

	return config
}

func getArchetypeNames() []string {
	var names []string
	for name := range archetypeTemplates {
		names = append(names, name)
	}
	return names
}