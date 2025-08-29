package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2/app"

	"desktop-companion/internal/character"
	"desktop-companion/internal/monitoring"
	"desktop-companion/internal/ui"
)

var (
	characterPath = flag.String("character", "assets/characters/default/character.json", "Path to character configuration file")
	debug         = flag.Bool("debug", false, "Enable debug logging")
	version       = flag.Bool("version", false, "Show version information")
	memProfile    = flag.String("memprofile", "", "Write memory profile to file")
	cpuProfile    = flag.String("cpuprofile", "", "Write CPU profile to file")
	gameMode      = flag.Bool("game", false, "Enable Tamagotchi game features")
	showStats     = flag.Bool("stats", false, "Show stats overlay")
)

const appVersion = "1.0.0"

// validateFlagDependencies checks that flag combinations are valid
func validateFlagDependencies(gameMode, showStats bool) error {
	if showStats && !gameMode {
		return fmt.Errorf("-stats flag requires -game flag to be enabled")
	}
	return nil
}

func main() {
	flag.Parse()

	// Validate flag dependencies
	if err := validateFlagDependencies(*gameMode, *showStats); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Use -help for usage information\n")
		os.Exit(1)
	}

	if *version {
		showVersionInfo()
		return
	}

	configureDebugLogging()

	// Initialize performance profiler
	profiler := monitoring.NewProfiler(50) // 50MB memory target

	// Start profiling if requested
	if err := profiler.Start(*memProfile, *cpuProfile, *debug); err != nil {
		log.Fatalf("Failed to start profiler: %v", err)
	}
	defer func() {
		if err := profiler.Stop(*memProfile, *debug); err != nil {
			log.Printf("Error stopping profiler: %v", err)
		}
	}()

	// Load character configuration
	card, characterDir := loadCharacterConfiguration()

	// Record startup completion
	profiler.RecordStartupComplete()

	if *debug {
		stats := profiler.GetStats()
		log.Printf("Startup completed in %v", stats.StartupDuration)
		log.Printf("Current memory usage: %.2f MB", stats.CurrentMemoryMB)

		if !profiler.IsMemoryTargetMet() {
			log.Printf("WARNING: Memory usage exceeds 50MB target")
		}
	}

	// Initialize application and show window
	runDesktopApplication(card, characterDir, profiler)
}

// showVersionInfo displays application version information.
func showVersionInfo() {
	fmt.Printf("Desktop Companion v%s\n", appVersion)
	fmt.Println("Built with Go and Fyne - Cross-platform desktop pet")
}

// configureDebugLogging sets up debug logging if enabled.
func configureDebugLogging() {
	if *debug {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Println("Debug mode enabled")
	}
}

// loadCharacterConfiguration loads and validates the character configuration file.
func loadCharacterConfiguration() (*character.CharacterCard, string) {
	var absPath string
	var err error

	// For the default relative path, resolve relative to project root (for development)
	// or executable directory (for deployed binaries)
	if *characterPath == "assets/characters/default/character.json" && !filepath.IsAbs(*characterPath) {
		// First try to find project root by looking for go.mod
		execPath, execErr := os.Executable()
		if execErr != nil {
			log.Fatalf("Failed to get executable path: %v", execErr)
		}

		searchDir := filepath.Dir(execPath)
		projectRoot := ""

		// Search upward for go.mod file
		for {
			if _, statErr := os.Stat(filepath.Join(searchDir, "go.mod")); statErr == nil {
				projectRoot = searchDir
				break
			}
			parent := filepath.Dir(searchDir)
			if parent == searchDir {
				// Reached filesystem root, use executable directory
				projectRoot = filepath.Dir(execPath)
				break
			}
			searchDir = parent
		}

		absPath = filepath.Join(projectRoot, *characterPath)
	} else {
		absPath, err = filepath.Abs(*characterPath)
		if err != nil {
			log.Fatalf("Failed to resolve character path: %v", err)
		}
	}

	if *debug {
		log.Printf("Loading character from: %s", absPath)
	}

	card, err := character.LoadCard(absPath)
	if err != nil {
		log.Fatalf("Failed to load character card: %v", err)
	}

	if *debug {
		log.Printf("Loaded character: %s - %s", card.Name, card.Description)
	}

	return card, filepath.Dir(absPath)
}

// runDesktopApplication creates and runs the desktop companion application.
func runDesktopApplication(card *character.CharacterCard, characterDir string, profiler *monitoring.Profiler) {
	// Check if we're in a headless environment before attempting to create UI
	if err := checkDisplayAvailable(); err != nil {
		log.Fatalf("Cannot run desktop application: %v", err)
	}

	// Create Fyne application
	myApp := app.NewWithID("com.opdai.desktop-companion")

	// Create character instance
	char, err := character.New(card, characterDir)
	if err != nil {
		log.Fatalf("Failed to create character: %v", err)
	}

	// Create and show desktop window with profiler integration
	window := ui.NewDesktopWindow(myApp, char, *debug, profiler, *gameMode, *showStats)

	if *debug {
		log.Println("Created desktop window")
	}

	// Show window and start event loop
	window.Show()
	myApp.Run()
}

// checkDisplayAvailable verifies that a display is available for GUI applications
func checkDisplayAvailable() error {
	// Check for X11 display on Linux/Unix systems
	display := os.Getenv("DISPLAY")
	if display == "" {
		return fmt.Errorf("no display available - DISPLAY environment variable is not set.\n" +
			"This application requires a graphical desktop environment to run.\n" +
			"Please run from a desktop session or use X11 forwarding for remote connections")
	}

	// For additional safety, we could also check for Wayland
	waylandDisplay := os.Getenv("WAYLAND_DISPLAY")
	if display == "" && waylandDisplay == "" {
		return fmt.Errorf("no display available - neither X11 (DISPLAY) nor Wayland (WAYLAND_DISPLAY) environment is available.\n" +
			"This application requires a graphical desktop environment to run")
	}

	return nil
}
