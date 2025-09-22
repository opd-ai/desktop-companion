package main

// main.go provides the CLI entry point for the gif-generator application.
// This implements the command-line interface outlined in GIF_PLAN.md with
// support for batch processing, single character generation, and validation.

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/opd-ai/desktop-companion/lib/backends"
	"github.com/opd-ai/desktop-companion/lib/character"
	"github.com/opd-ai/desktop-companion/lib/comfyui" // For template management
	"github.com/opd-ai/desktop-companion/lib/pipeline"
)

const (
	version = "1.0.0"
	appName = "gif-generator"
)

// Command represents a CLI command.
type Command struct {
	Name        string
	Description string
	Usage       string
	Handler     func(args []string) error
}

// CLIConfig holds CLI application configuration.
type CLIConfig struct {
	ConfigPath string
	Backend    string // "comfyui" or "swarmui"
	ServerURL  string // Backend server URL
	OutputDir  string
	TempDir    string
	Parallel   int
	Verbose    bool
	DryRun     bool
}

var (
	globalConfig CLIConfig
	commands     map[string]Command
)

func init() {
	commands = map[string]Command{
		"character": {
			Name:        "character",
			Description: "Generate assets for a character",
			Usage:       "gif-generator character --file character.json [options] OR --archetype ARCHETYPE [options]",
			Handler:     handleCharacterCommand,
		},
		"batch": {
			Name:        "batch",
			Description: "Generate assets for multiple characters",
			Usage:       "gif-generator batch --config CONFIG [options]",
			Handler:     handleBatchCommand,
		},
		"validate": {
			Name:        "validate",
			Description: "Validate existing assets",
			Usage:       "gif-generator validate --path PATH [options]",
			Handler:     handleValidateCommand,
		},
		"deploy": {
			Name:        "deploy",
			Description: "Deploy generated assets to target location",
			Usage:       "gif-generator deploy --source SOURCE --target TARGET [options]",
			Handler:     handleDeployCommand,
		},
		"list-templates": {
			Name:        "list-templates",
			Description: "List available workflow templates",
			Usage:       "gif-generator list-templates [--templates-dir DIR]",
			Handler:     handleListTemplatesCommand,
		},
		"version": {
			Name:        "version",
			Description: "Show version information",
			Usage:       "gif-generator version",
			Handler:     handleVersionCommand,
		},
		"help": {
			Name:        "help",
			Description: "Show help information",
			Usage:       "gif-generator help [COMMAND]",
			Handler:     handleHelpCommand,
		},
	}
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Parse global flags and find command
	commandIdx := parseGlobalFlags()

	if commandIdx >= len(os.Args) {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[commandIdx]
	args := os.Args[commandIdx+1:]

	if cmd, exists := commands[command]; exists {
		if err := cmd.Handler(args); err != nil {
			log.Fatalf("Command %s failed: %v", command, err)
		}
	} else {
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

// parseGlobalFlags parses global command-line flags and returns the index of the first non-flag argument.
func parseGlobalFlags() int {
	// Note: This is a simplified flag parsing approach.
	// A production CLI would use a more sophisticated library like cobra or cli.

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch arg {
		case "--config":
			if i+1 < len(os.Args) {
				globalConfig.ConfigPath = os.Args[i+1]
				i++ // Skip the value
			}
		case "--backend":
			if i+1 < len(os.Args) {
				globalConfig.Backend = os.Args[i+1]
				i++ // Skip the value
			}
		case "--server-url":
			if i+1 < len(os.Args) {
				globalConfig.ServerURL = os.Args[i+1]
				i++ // Skip the value
			}
		// Legacy support for --comfyui-url
		case "--comfyui-url":
			if i+1 < len(os.Args) {
				globalConfig.ServerURL = os.Args[i+1]
				globalConfig.Backend = "comfyui" // Assume ComfyUI backend
				i++                              // Skip the value
			}
		case "--output":
			if i+1 < len(os.Args) {
				globalConfig.OutputDir = os.Args[i+1]
				i++ // Skip the value
			}
		case "--temp-dir":
			if i+1 < len(os.Args) {
				globalConfig.TempDir = os.Args[i+1]
				i++ // Skip the value
			}
		case "--parallel":
			if i+1 < len(os.Args) {
				fmt.Sscanf(os.Args[i+1], "%d", &globalConfig.Parallel)
				i++ // Skip the value
			}
		case "--verbose", "-v":
			globalConfig.Verbose = true
		case "--dry-run":
			globalConfig.DryRun = true
		default:
			// This is not a global flag, must be the command
			setDefaults()
			return i
		}
	}

	// If we get here, there were only flags and no command
	setDefaults()
	return len(os.Args)
}

func setDefaults() {
	// Set defaults
	if globalConfig.ConfigPath == "" {
		globalConfig.ConfigPath = "config.json"
	}
	if globalConfig.Backend == "" {
		globalConfig.Backend = "comfyui" // Default to ComfyUI
	} else {
		// Validate backend type
		validBackends := []string{"comfyui", "swarmui"}
		isValid := false
		for _, valid := range validBackends {
			if globalConfig.Backend == valid {
				isValid = true
				break
			}
		}
		if !isValid {
			fmt.Fprintf(os.Stderr, "Error: Invalid backend type '%s'. Valid options are: %s\n", 
				globalConfig.Backend, strings.Join(validBackends, ", "))
			fmt.Fprintf(os.Stderr, "Use 'gif-generator help' for more information.\n")
			os.Exit(1)
		}
	}
	if globalConfig.ServerURL == "" {
		switch globalConfig.Backend {
		case "comfyui":
			globalConfig.ServerURL = "http://localhost:8188"
		case "swarmui":
			globalConfig.ServerURL = "http://localhost:7801"
		default:
			globalConfig.ServerURL = "http://localhost:8188" // Default to ComfyUI
		}
	}
	if globalConfig.Parallel == 0 {
		globalConfig.Parallel = 2
	}
}

// printUsage prints the main usage information.
func printUsage() {
	fmt.Printf("%s v%s - Desktop Companion GIF Asset Generator\n\n", appName, version)
	fmt.Println("Usage:")
	fmt.Printf("  %s [GLOBAL OPTIONS] COMMAND [COMMAND OPTIONS]\n\n", appName)
	fmt.Println("Commands:")

	for _, cmd := range commands {
		fmt.Printf("  %-15s %s\n", cmd.Name, cmd.Description)
	}

	fmt.Println("\nGlobal Options:")
	fmt.Println("  --config PATH        Pipeline configuration file (default: config.json)")
	fmt.Println("  --backend TYPE       Backend type: comfyui, swarmui (default: comfyui)")
	fmt.Println("  --server-url URL     Backend server URL")
	fmt.Println("  --comfyui-url URL    ComfyUI server URL (legacy, use --server-url with --backend=comfyui)")
	fmt.Println("  --output DIR         Output directory")
	fmt.Println("  --temp-dir DIR       Temporary directory")
	fmt.Println("  --parallel N         Number of parallel jobs (default: 2)")
	fmt.Println("  --verbose, -v        Verbose output")
	fmt.Println("  --dry-run            Show what would be done without executing")
	
	fmt.Println("\nBackend Configuration:")
	fmt.Println("  ComfyUI (default):   --backend comfyui --server-url http://localhost:8188")
	fmt.Println("  SwarmUI:             --backend swarmui --server-url http://localhost:7801")
	fmt.Println("  Legacy ComfyUI:      --comfyui-url http://localhost:8188")
	
	fmt.Println("\nQuick Start Examples:")
	fmt.Println("  # Generate default character with ComfyUI")
	fmt.Printf("  %s character --archetype default\n", appName)
	fmt.Println()
	fmt.Println("  # Generate romance character with SwarmUI")
	fmt.Printf("  %s --backend swarmui character --archetype romance_tsundere\n", appName)
	fmt.Println()
	fmt.Println("  # Process multiple characters")
	fmt.Printf("  %s batch --config characters.txt\n", appName)
	
	fmt.Println("\nNote:")
	fmt.Println("  Global flags must be specified before the command.")
	fmt.Println("  Use space syntax: --flag value (not --flag=value)")
	
	fmt.Printf("\nUse '%s help COMMAND' for detailed command information.\n", appName)
}

// handleCharacterCommand generates assets for a single character.
func handleCharacterCommand(args []string) error {
	fs := flag.NewFlagSet("character", flag.ExitOnError)
	characterFile := fs.String("file", "", "Character JSON file path")
	archetype := fs.String("archetype", "", "Character archetype (alternative to --file)")
	model := fs.String("model", "flux1d", "AI model to use (sdxl, flux1d, flux1s)")
	style := fs.String("style", "pixel_art", "Art style (when using --archetype)")
	description := fs.String("description", "", "Character description (when using --archetype)")
	states := fs.String("states", "", "Comma-separated animation states (optional)")
	output := fs.String("output", "", "Output directory (overrides default)")
	validate := fs.Bool("validate", false, "Validate generated assets")
	backup := fs.Bool("backup", false, "Backup existing assets before generation")
	
	// Backend configuration flags
	timeout := fs.String("timeout", "", "Backend timeout (e.g., 30s, 1m)")
	retryAttempts := fs.Int("retry-attempts", 0, "Number of retry attempts (0 = use backend default)")
	retryBackoff := fs.String("retry-backoff", "", "Retry backoff duration (e.g., 500ms, 1s)")

	fs.Parse(args)

	// Require either --file or --archetype
	if *characterFile == "" && *archetype == "" {
		return fmt.Errorf("either --file or --archetype is required")
	}
	if *characterFile != "" && *archetype != "" {
		return fmt.Errorf("cannot specify both --file and --archetype")
	}

	if globalConfig.Verbose {
		if *characterFile != "" {
			fmt.Printf("Generating assets from character file: %s\n", *characterFile)
		} else {
			fmt.Printf("Generating assets for character archetype: %s\n", *archetype)
		}
	}

	// Load pipeline configuration
	config, err := loadPipelineConfig()
	if err != nil {
		return fmt.Errorf("load pipeline config: %w", err)
	}

	// Apply command-line backend configuration overrides
	if err := applyBackendOverrides(config, *timeout, *retryAttempts, *retryBackoff); err != nil {
		return fmt.Errorf("apply backend overrides: %w", err)
	}

	var charConfig *pipeline.CharacterConfig

	if *characterFile != "" {
		// Load from character.json file
		charConfig, err = loadCharacterConfigFromFile(*characterFile, *model)
		if err != nil {
			return fmt.Errorf("load character config from file: %w", err)
		}
	} else {
		// Create from archetype
		charConfig = pipeline.DefaultCharacterConfig(*archetype)
		charConfig.Character.Style = *style
		if *description != "" {
			charConfig.Character.Description = *description
		}
		// Store model info in traits for archetype-based generation
		if charConfig.Character.Traits == nil {
			charConfig.Character.Traits = make(map[string]string)
		}
		charConfig.Character.Traits["model"] = *model
	}

	// Apply command-line overrides
	if *states != "" {
		charConfig.States = strings.Split(*states, ",")
		for i, state := range charConfig.States {
			charConfig.States[i] = strings.TrimSpace(state)
		}
	}
	if *output != "" {
		charConfig.Deployment.OutputDir = *output
	} else if globalConfig.OutputDir != "" {
		charConfig.Deployment.OutputDir = globalConfig.OutputDir
	}

	// Apply validation and backup settings
	// Note: These are applied to the deployment config since ValidationConfig doesn't have an Enabled field
	charConfig.Deployment.ValidateBeforeDeploy = *validate
	charConfig.Deployment.BackupExisting = *backup

	if globalConfig.DryRun {
		if *characterFile != "" {
			fmt.Printf("Would generate assets from file: %s\n", *characterFile)
		} else {
			fmt.Printf("Would generate assets for character %s with style %s\n", *archetype, *style)
		}
		fmt.Printf("Model: %s\n", charConfig.Character.Traits["model"])
		fmt.Printf("States: %v\n", charConfig.States)
		fmt.Printf("Output: %s\n", charConfig.Deployment.OutputDir)
		fmt.Printf("Validation: %t\n", charConfig.Deployment.ValidateBeforeDeploy)
		fmt.Printf("Backup: %t\n", charConfig.Deployment.BackupExisting)
		return nil
	}

	// Create pipeline controller
	controller, err := createController(config)
	if err != nil {
		return fmt.Errorf("create controller: %w", err)
	}

	// Process character
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	result, err := controller.ProcessCharacter(ctx, charConfig)
	if err != nil {
		return fmt.Errorf("process character: %w", err)
	}

	// Print results
	printProcessResult(result)

	// Deploy if successful
	if result.Success {
		if err := controller.DeployAssets(ctx, result); err != nil {
			return fmt.Errorf("deploy assets: %w", err)
		}
		fmt.Printf("Assets deployed to: %s\n", charConfig.Deployment.OutputDir)
	}

	return nil
}

// handleBatchCommand generates assets for multiple characters.
func handleBatchCommand(args []string) error {
	fs := flag.NewFlagSet("batch", flag.ExitOnError)
	configPath := fs.String("config", "", "Batch configuration file (required)")
	parallel := fs.Int("parallel", globalConfig.Parallel, "Number of parallel jobs")
	output := fs.String("output", "", "Output directory (overrides config)")
	
	// Backend configuration flags
	timeout := fs.String("timeout", "", "Backend timeout (e.g., 30s, 1m)")
	retryAttempts := fs.Int("retry-attempts", 0, "Number of retry attempts (0 = use backend default)")
	retryBackoff := fs.String("retry-backoff", "", "Retry backoff duration (e.g., 500ms, 1s)")

	fs.Parse(args)

	if *configPath == "" {
		return fmt.Errorf("--config is required")
	}

	if globalConfig.Verbose {
		fmt.Printf("Starting batch processing with %d parallel jobs\n", *parallel)
	}

	// Load batch configuration
	batchConfigs, err := loadBatchConfigs(*configPath)
	if err != nil {
		return fmt.Errorf("load batch configs: %w", err)
	}

	// Override output directory if specified
	if *output != "" {
		for _, config := range batchConfigs {
			config.Deployment.OutputDir = filepath.Join(*output, config.Character.Archetype)
		}
	}

	if globalConfig.DryRun {
		fmt.Printf("Would process %d characters:\n", len(batchConfigs))
		for _, config := range batchConfigs {
			fmt.Printf("  - %s (%s style, %d states)\n",
				config.Character.Archetype, config.Character.Style, len(config.States))
		}
		return nil
	}

	// Load pipeline configuration
	pipelineConfig, err := loadPipelineConfig()
	if err != nil {
		return fmt.Errorf("load pipeline config: %w", err)
	}

	// Apply command-line backend configuration overrides
	if err := applyBackendOverrides(pipelineConfig, *timeout, *retryAttempts, *retryBackoff); err != nil {
		return fmt.Errorf("apply backend overrides: %w", err)
	}

	// Override concurrent jobs
	pipelineConfig.Generation.ConcurrentJobs = *parallel

	// Create pipeline controller
	controller, err := createController(pipelineConfig)
	if err != nil {
		return fmt.Errorf("create controller: %w", err)
	}

	// Process batch
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	result, err := controller.ProcessBatch(ctx, batchConfigs)
	if err != nil {
		return fmt.Errorf("process batch: %w", err)
	}

	// Print results
	printBatchResult(result)

	return nil
}

// handleValidateCommand validates existing assets.
func handleValidateCommand(args []string) error {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	path := fs.String("path", "", "Path to validate (required)")
	recursive := fs.Bool("recursive", false, "Validate recursively")

	fs.Parse(args)

	if *path == "" {
		return fmt.Errorf("--path is required")
	}

	if globalConfig.Verbose {
		fmt.Printf("Validating assets at: %s\n", *path)
	}

	// Load pipeline configuration for validation settings
	config, err := loadPipelineConfig()
	if err != nil {
		return fmt.Errorf("load pipeline config: %w", err)
	}

	// Create validator
	validator := pipeline.NewValidator()

	ctx := context.Background()

	if *recursive {
		// Validate all character directories
		entries, err := os.ReadDir(*path)
		if err != nil {
			return fmt.Errorf("read directory: %w", err)
		}

		var allResults []*pipeline.CharacterValidationResult
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			charDir := filepath.Join(*path, entry.Name())
			charConfig := pipeline.DefaultCharacterConfig(entry.Name())
			charConfig.Deployment.OutputDir = charDir

			result, err := validator.ValidateCharacterSet(ctx, charDir, charConfig)
			if err != nil {
				if globalConfig.Verbose {
					fmt.Printf("Skipping %s: %v\n", entry.Name(), err)
				}
				continue
			}

			allResults = append(allResults, result)
			printCharacterValidationResult(result)
		}

		// Print summary
		printValidationSummary(allResults)
	} else {
		// Validate single asset or character directory
		stat, err := os.Stat(*path)
		if err != nil {
			return fmt.Errorf("stat path: %w", err)
		}

		if stat.IsDir() {
			// Validate character directory
			charName := filepath.Base(*path)
			charConfig := pipeline.DefaultCharacterConfig(charName)
			charConfig.Deployment.OutputDir = *path

			result, err := validator.ValidateCharacterSet(ctx, *path, charConfig)
			if err != nil {
				return fmt.Errorf("validate character set: %w", err)
			}

			printCharacterValidationResult(result)
		} else {
			// Validate single asset
			result, err := validator.ValidateAsset(ctx, *path, &config.Validation)
			if err != nil {
				return fmt.Errorf("validate asset: %w", err)
			}

			printAssetValidationResult(result)
		}
	}

	return nil
}

// handleDeployCommand deploys generated assets to target location.
func handleDeployCommand(args []string) error {
	fs := flag.NewFlagSet("deploy", flag.ExitOnError)
	source := fs.String("source", "", "Source directory (required)")
	target := fs.String("target", "", "Target directory (required)")
	backup := fs.Bool("backup", true, "Backup existing files")

	fs.Parse(args)

	if *source == "" {
		return fmt.Errorf("--source is required")
	}
	if *target == "" {
		return fmt.Errorf("--target is required")
	}

	if globalConfig.Verbose {
		fmt.Printf("Deploying assets from %s to %s\n", *source, *target)
	}

	if globalConfig.DryRun {
		fmt.Printf("Would deploy assets from %s to %s (backup: %v)\n", *source, *target, *backup)
		return nil
	}

	// Simple deployment implementation
	return deployAssets(*source, *target, *backup)
}

// handleListTemplatesCommand lists available workflow templates.
func handleListTemplatesCommand(args []string) error {
	fs := flag.NewFlagSet("list-templates", flag.ExitOnError)
	templatesDir := fs.String("templates-dir", "templates/workflows", "Templates directory")

	fs.Parse(args)

	if globalConfig.Verbose {
		fmt.Printf("Listing templates in: %s\n", *templatesDir)
	}

	// Create template manager
	manager := comfyui.NewTemplateManager()

	templates, err := manager.ListTemplates(*templatesDir)
	if err != nil {
		return fmt.Errorf("list templates: %w", err)
	}

	if len(templates) == 0 {
		fmt.Println("No templates found")
		return nil
	}

	fmt.Printf("Found %d templates:\n\n", len(templates))
	for _, tmpl := range templates {
		fmt.Printf("ID: %s\n", tmpl.ID)
		fmt.Printf("Name: %s\n", tmpl.Name)
		fmt.Printf("Description: %s\n", tmpl.Description)
		fmt.Printf("Version: %s\n", tmpl.Version)
		fmt.Printf("Parameters: %d\n", len(tmpl.Parameters))
		fmt.Printf("Prompt Slots: %d\n", len(tmpl.PromptSlots))
		if tmpl.Metadata.Category != "" {
			fmt.Printf("Category: %s\n", tmpl.Metadata.Category)
		}
		if len(tmpl.Metadata.Tags) > 0 {
			fmt.Printf("Tags: %s\n", strings.Join(tmpl.Metadata.Tags, ", "))
		}
		fmt.Println()
	}

	return nil
}

// handleVersionCommand shows version information.
func handleVersionCommand(args []string) error {
	fmt.Printf("%s version %s\n", appName, version)
	return nil
}

// handleHelpCommand shows help information.
func handleHelpCommand(args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	command := args[0]
	if cmd, exists := commands[command]; exists {
		fmt.Printf("%s\n\n", cmd.Description)
		fmt.Printf("Usage: %s\n", cmd.Usage)

		// Print command-specific help based on command
		switch command {
		case "character":
			fmt.Println("\nOptions:")
			fmt.Println("  --file FILE          Character JSON file path")
			fmt.Println("  --archetype TYPE     Character archetype (alternative to --file)")
			fmt.Println("  --style STYLE        Art style (default: pixel_art)")
			fmt.Println("  --model MODEL        AI model to use (default: flux1d)")
			fmt.Println("  --description TEXT   Character description")
			fmt.Println("  --states LIST        Comma-separated animation states")
			fmt.Println("  --output DIR         Output directory")
			fmt.Println("  --validate           Validate generated assets")
			fmt.Println("  --backup             Backup existing assets")
			fmt.Println("  --timeout DURATION   Backend timeout (e.g., 30s, 1m)")
			fmt.Println("  --retry-attempts N   Number of retry attempts")
			fmt.Println("  --retry-backoff DUR  Retry backoff duration (e.g., 500ms)")
			fmt.Println()
			fmt.Println("Examples:")
			fmt.Println("  # Generate from archetype with ComfyUI")
			fmt.Println("  gif-generator --backend comfyui character --archetype romance_tsundere")
			fmt.Println()
			fmt.Println("  # Generate from archetype with SwarmUI and custom timeout")
			fmt.Println("  gif-generator --backend swarmui --server-url http://localhost:7801 \\")
			fmt.Println("    character --archetype default --timeout 45s")
			fmt.Println()
			fmt.Println("  # Generate from character file")
			fmt.Println("  gif-generator character --file characters/my_character.json")

		case "batch":
			fmt.Println("\nOptions:")
			fmt.Println("  --config FILE        Batch configuration file (required)")
			fmt.Println("  --parallel N         Number of parallel jobs")
			fmt.Println("  --output DIR         Output directory base")
			fmt.Println("  --timeout DURATION   Backend timeout (e.g., 30s, 1m)")
			fmt.Println("  --retry-attempts N   Number of retry attempts")
			fmt.Println("  --retry-backoff DUR  Retry backoff duration (e.g., 500ms)")
			fmt.Println()
			fmt.Println("Examples:")
			fmt.Println("  # Process batch with ComfyUI")
			fmt.Println("  gif-generator --backend comfyui batch --config batch.txt")
			fmt.Println()
			fmt.Println("  # Process batch with SwarmUI and custom settings")
			fmt.Println("  gif-generator --backend swarmui --server-url http://localhost:7801 \\")
			fmt.Println("    batch --config batch.txt --timeout 60s --parallel 4")

		case "validate":
			fmt.Println("\nOptions:")
			fmt.Println("  --path PATH          Path to validate (required)")
			fmt.Println("  --recursive          Validate recursively")

		case "deploy":
			fmt.Println("\nOptions:")
			fmt.Println("  --source DIR         Source directory (required)")
			fmt.Println("  --target DIR         Target directory (required)")
			fmt.Println("  --backup             Backup existing files (default: true)")
		}
	} else {
		return fmt.Errorf("unknown command: %s", command)
	}

	return nil
}

// loadPipelineConfig loads the pipeline configuration.
func loadPipelineConfig() (*pipeline.PipelineConfig, error) {
	if _, err := os.Stat(globalConfig.ConfigPath); os.IsNotExist(err) {
		if globalConfig.Verbose {
			fmt.Printf("Config file %s not found, using defaults\n", globalConfig.ConfigPath)
		}
		config := pipeline.DefaultPipelineConfig()

		// Override with command-line settings
		if globalConfig.ServerURL != "" {
			// New backend system
			backendType := globalConfig.Backend
			if backendType == "" {
				backendType = "comfyui" // Default
			}

			switch backendType {
			case "comfyui":
				if config.Backend.ComfyUI == nil {
					config.Backend.ComfyUI = &backends.ComfyUIConfig{}
				}
				config.Backend.Type = backends.BackendTypeComfyUI
				config.Backend.ComfyUI.ServerURL = globalConfig.ServerURL
			case "swarmui":
				if config.Backend.SwarmUI == nil {
					config.Backend.SwarmUI = &backends.SwarmUIConfig{}
				}
				config.Backend.Type = backends.BackendTypeSwarmUI
				config.Backend.SwarmUI.ServerURL = globalConfig.ServerURL
			}
		}
		if globalConfig.TempDir != "" {
			config.Generation.TempDir = globalConfig.TempDir
		}

		return config, nil
	}

	config, err := pipeline.LoadConfig(globalConfig.ConfigPath)
	if err != nil {
		return nil, err
	}

	// Override with command-line settings
	if globalConfig.ServerURL != "" {
		// New backend system
		backendType := globalConfig.Backend
		if backendType == "" {
			backendType = "comfyui" // Default
		}

		switch backendType {
		case "comfyui":
			if config.Backend.ComfyUI == nil {
				config.Backend.ComfyUI = &backends.ComfyUIConfig{}
			}
			config.Backend.Type = backends.BackendTypeComfyUI
			config.Backend.ComfyUI.ServerURL = globalConfig.ServerURL
		case "swarmui":
			if config.Backend.SwarmUI == nil {
				config.Backend.SwarmUI = &backends.SwarmUIConfig{}
			}
			config.Backend.Type = backends.BackendTypeSwarmUI
			config.Backend.SwarmUI.ServerURL = globalConfig.ServerURL
		}
	}
	if globalConfig.TempDir != "" {
		config.Generation.TempDir = globalConfig.TempDir
	}

	return config, nil
}

// applyBackendOverrides applies command-line backend configuration overrides to the pipeline config.
func applyBackendOverrides(config *pipeline.PipelineConfig, timeoutStr string, retryAttempts int, retryBackoffStr string) error {
	if timeoutStr != "" {
		timeout, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return fmt.Errorf("invalid timeout duration %q: %w", timeoutStr, err)
		}
		
		switch config.Backend.Type {
		case backends.BackendTypeComfyUI:
			if config.Backend.ComfyUI != nil {
				config.Backend.ComfyUI.Timeout = timeout
			}
		case backends.BackendTypeSwarmUI:
			if config.Backend.SwarmUI != nil {
				config.Backend.SwarmUI.Timeout = timeout
			}
		}
	}
	
	if retryAttempts > 0 {
		switch config.Backend.Type {
		case backends.BackendTypeComfyUI:
			if config.Backend.ComfyUI != nil {
				config.Backend.ComfyUI.RetryAttempts = retryAttempts
			}
		case backends.BackendTypeSwarmUI:
			if config.Backend.SwarmUI != nil {
				config.Backend.SwarmUI.RetryAttempts = retryAttempts
			}
		}
	}
	
	if retryBackoffStr != "" {
		retryBackoff, err := time.ParseDuration(retryBackoffStr)
		if err != nil {
			return fmt.Errorf("invalid retry backoff duration %q: %w", retryBackoffStr, err)
		}
		
		switch config.Backend.Type {
		case backends.BackendTypeComfyUI:
			if config.Backend.ComfyUI != nil {
				config.Backend.ComfyUI.RetryBackoff = retryBackoff
			}
		case backends.BackendTypeSwarmUI:
			if config.Backend.SwarmUI != nil {
				config.Backend.SwarmUI.RetryBackoff = retryBackoff
			}
		}
	}
	
	return nil
}

// loadBatchConfigs loads batch processing configurations.
func loadBatchConfigs(path string) ([]*pipeline.CharacterConfig, error) {
	// This is a simplified implementation. A full version would support
	// various batch configuration formats (JSON, YAML, etc.)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read batch config: %w", err)
	}

	// For now, assume it's a simple list of archetype names
	lines := strings.Split(string(data), "\n")
	var configs []*pipeline.CharacterConfig

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		config := pipeline.DefaultCharacterConfig(line)
		configs = append(configs, config)
	}

	return configs, nil
}

// createController creates a pipeline controller from configuration.
func createController(config *pipeline.PipelineConfig) (pipeline.Controller, error) {
	// The pipeline controller now handles backend creation internally
	return pipeline.NewController(config)
}

// Utility functions for printing results would go here...
// (Simplified for brevity - full implementation would have detailed formatting)

func printProcessResult(result *pipeline.ProcessResult) {
	fmt.Printf("Character: %s\n", result.Character)
	fmt.Printf("Success: %v\n", result.Success)
	fmt.Printf("Generated Assets: %d\n", len(result.GeneratedAssets))
	fmt.Printf("Errors: %d\n", len(result.Errors))
	fmt.Printf("Warnings: %d\n", len(result.Warnings))
	fmt.Printf("Processing Time: %v\n", result.ProcessingTime)

	if globalConfig.Verbose {
		for state, asset := range result.GeneratedAssets {
			fmt.Printf("  %s: %s (%v)\n", state, asset.OutputPath, asset.GenerationTime)
		}

		for _, err := range result.Errors {
			fmt.Printf("  Error [%s]: %s\n", err.Stage, err.Message)
		}
	}
}

func printBatchResult(result *pipeline.BatchResult) {
	fmt.Printf("Batch Processing Complete\n")
	fmt.Printf("Overall Success: %v\n", result.OverallSuccess)
	fmt.Printf("Characters Processed: %d\n", len(result.Characters))
	fmt.Printf("Processing Time: %v\n", result.ProcessingTime)

	if result.Summary != nil {
		fmt.Printf("Summary:\n")
		fmt.Printf("  Total Characters: %d\n", result.Summary.TotalCharacters)
		fmt.Printf("  Successful: %d\n", result.Summary.SuccessfulCharacters)
		fmt.Printf("  Failed: %d\n", result.Summary.FailedCharacters)
		fmt.Printf("  Success Rate: %.1f%%\n", result.Summary.SuccessRate*100)
		fmt.Printf("  Average Processing Time: %v\n", result.Summary.AverageProcessingTime)
	}
}

func printCharacterValidationResult(result *pipeline.CharacterValidationResult) {
	fmt.Printf("Character: %s\n", result.Character)
	fmt.Printf("Valid: %v\n", result.Valid)
	fmt.Printf("Assets: %d\n", len(result.AssetResults))
	fmt.Printf("Missing States: %d\n", len(result.MissingStates))

	if globalConfig.Verbose {
		for state, assetResult := range result.AssetResults {
			fmt.Printf("  %s: %v (%d errors, %d warnings)\n",
				state, assetResult.Valid, len(assetResult.Errors), len(assetResult.Warnings))
		}
	}
	fmt.Println()
}

func printAssetValidationResult(result *pipeline.ValidationResult) {
	fmt.Printf("Asset: %s\n", result.AssetPath)
	fmt.Printf("Valid: %v\n", result.Valid)
	fmt.Printf("Errors: %d\n", len(result.Errors))
	fmt.Printf("Warnings: %d\n", len(result.Warnings))

	if result.Metrics != nil {
		fmt.Printf("Metrics:\n")
		fmt.Printf("  File Size: %d bytes\n", result.Metrics.FileSize)
		fmt.Printf("  Dimensions: %dx%d\n", result.Metrics.Dimensions[0], result.Metrics.Dimensions[1])
		fmt.Printf("  Frame Count: %d\n", result.Metrics.FrameCount)
		fmt.Printf("  Frame Rate: %.1f fps\n", result.Metrics.FrameRate)
		fmt.Printf("  Transparency: %v\n", result.Metrics.HasTransparency)
	}
}

func printValidationSummary(results []*pipeline.CharacterValidationResult) {
	valid := 0
	total := len(results)

	for _, result := range results {
		if result.Valid {
			valid++
		}
	}

	fmt.Printf("Validation Summary:\n")
	fmt.Printf("  Total Characters: %d\n", total)
	fmt.Printf("  Valid: %d\n", valid)
	fmt.Printf("  Invalid: %d\n", total-valid)
	fmt.Printf("  Success Rate: %.1f%%\n", float64(valid)/float64(total)*100)
}

// deployAssets performs simple asset deployment.
func deployAssets(source, target string, backup bool) error {
	// This is a simplified deployment implementation
	return fmt.Errorf("deployment not yet implemented - would copy from %s to %s", source, target)
}

// loadCharacterConfigFromFile loads a character.json file and creates a pipeline configuration
func loadCharacterConfigFromFile(filePath, model string) (*pipeline.CharacterConfig, error) {
	// Read the character.json file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read character file: %w", err)
	}

	// Parse character card
	var card character.CharacterCard
	if err := json.Unmarshal(data, &card); err != nil {
		return nil, fmt.Errorf("parse character JSON: %w", err)
	}

	// Validate character card
	if err := card.Validate(); err != nil {
		return nil, fmt.Errorf("invalid character card: %w", err)
	}

	// Check if character has asset generation configuration
	if card.AssetGeneration == nil {
		return nil, fmt.Errorf("character file does not contain assetGeneration configuration")
	}

	// Convert to pipeline configuration
	charConfig := &pipeline.CharacterConfig{
		Character: &pipeline.CharacterRequest{
			Archetype:   card.Name,
			Description: card.AssetGeneration.BasePrompt,
			Style:       card.AssetGeneration.GenerationSettings.ArtStyle,
			Traits:      make(map[string]string),
			OutputConfig: &pipeline.OutputConfig{
				Width:      card.AssetGeneration.GenerationSettings.Resolution.Width,
				Height:     card.AssetGeneration.GenerationSettings.Resolution.Height,
				Format:     "png",
				Background: "transparent",
			},
		},
		States: extractAnimationStates(card.AssetGeneration.AnimationMappings),
		GIFConfig: &pipeline.ExtendedGIFConfig{
			GIFConfig: pipeline.GIFConfig{
				FrameCount:   card.AssetGeneration.GenerationSettings.AnimationSettings.FrameRate,
				MaxFileSize:  card.AssetGeneration.GenerationSettings.AnimationSettings.MaxFileSize * 1024, // Convert KB to bytes
				Transparency: card.AssetGeneration.GenerationSettings.AnimationSettings.TransparencyEnabled,
			},
			Width:        card.AssetGeneration.GenerationSettings.Resolution.Width,
			Height:       card.AssetGeneration.GenerationSettings.Resolution.Height,
			FrameRate:    card.AssetGeneration.GenerationSettings.AnimationSettings.FrameRate,
			Colors:       256, // Default adaptive palette
			Optimization: card.AssetGeneration.GenerationSettings.AnimationSettings.Optimization,
		},
		Validation: &pipeline.ValidationConfig{
			MaxFileSize:          card.AssetGeneration.GenerationSettings.AnimationSettings.MaxFileSize * 1024,
			MinFrameRate:         5,                                           // Minimum acceptable frame rate
			RequiredStates:       []string{"idle", "talking", "happy", "sad"}, // Core states
			StyleConsistency:     true,
			ArchetypeCompliance:  true,
			TransparencyRequired: card.AssetGeneration.GenerationSettings.AnimationSettings.TransparencyEnabled,
		},
		Deployment: &pipeline.DeploymentConfig{
			OutputDir:            filepath.Dir(filePath), // Output to same directory as character.json
			BackupExisting:       card.AssetGeneration.BackupSettings.Enabled,
			UpdateCharacterJSON:  true,
			ValidateBeforeDeploy: true,
		},
	}

	// Override model if specified
	if model != "" {
		// Store model info in traits for now (pipeline doesn't have direct model field)
		charConfig.Character.Traits["model"] = model
	}

	return charConfig, nil
}

// extractAnimationStates extracts animation state names from the asset generation mappings
func extractAnimationStates(mappings map[string]character.AnimationMapping) []string {
	states := make([]string, 0, len(mappings))
	for state := range mappings {
		states = append(states, state)
	}
	return states
}
