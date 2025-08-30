package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"desktop-companion/internal/character"
	"desktop-companion/internal/monitoring"
	"desktop-companion/internal/network"
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
	events        = flag.Bool("events", false, "Enable general dialog events system")
	triggerEvent  = flag.String("trigger-event", "", "Manually trigger a specific event by name")
	networkMode   = flag.Bool("network", false, "Enable multiplayer networking features")
	showNetwork   = flag.Bool("network-ui", false, "Show network overlay UI")
)

const appVersion = "1.0.0"

// validateFlagDependencies checks that flag combinations are valid
func validateFlagDependencies(gameMode, showStats, networkMode, showNetwork bool) error {
	if showStats && !gameMode {
		return fmt.Errorf("-stats flag requires -game flag to be enabled")
	}
	if showNetwork && !networkMode {
		return fmt.Errorf("-network-ui flag requires -network flag to be enabled")
	}
	return nil
}

func main() {
	flag.Parse()

	// Validate flag dependencies
	if err := validateFlagDependencies(*gameMode, *showStats, *networkMode, *showNetwork); err != nil {
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

// resolveProjectRoot finds the project root by searching upward for go.mod file.
func resolveProjectRoot() string {
	execPath, execErr := os.Executable()
	if execErr != nil {
		log.Fatalf("Failed to get executable path: %v", execErr)
	}

	searchDir := filepath.Dir(execPath)

	// Search upward for go.mod file
	for {
		if _, statErr := os.Stat(filepath.Join(searchDir, "go.mod")); statErr == nil {
			return searchDir
		}
		parent := filepath.Dir(searchDir)
		if parent == searchDir {
			// Reached filesystem root, use executable directory
			return filepath.Dir(execPath)
		}
		searchDir = parent
	}
}

// resolveCharacterPath converts the character path to an absolute path.
func resolveCharacterPath() string {
	// For the default relative path, resolve relative to project root (for development)
	// or executable directory (for deployed binaries)
	if *characterPath == "assets/characters/default/character.json" && !filepath.IsAbs(*characterPath) {
		projectRoot := resolveProjectRoot()
		return filepath.Join(projectRoot, *characterPath)
	}

	absPath, err := filepath.Abs(*characterPath)
	if err != nil {
		log.Fatalf("Failed to resolve character path: %v", err)
	}
	return absPath
}

// loadAndValidateCharacter loads the character card from the given path with debug logging.
func loadAndValidateCharacter(absPath string) *character.CharacterCard {
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

	return card
}

// loadCharacterConfiguration loads and validates the character configuration file.
func loadCharacterConfiguration() (*character.CharacterCard, string) {
	absPath := resolveCharacterPath()
	card := loadAndValidateCharacter(absPath)
	return card, filepath.Dir(absPath)
}

// runDesktopApplication creates and runs the desktop companion application.
func runDesktopApplication(card *character.CharacterCard, characterDir string, profiler *monitoring.Profiler) {
	if err := checkDisplayAvailable(); err != nil {
		log.Fatalf("Cannot run desktop application: %v", err)
	}

	myApp := app.NewWithID("com.opdai.desktop-companion")

	char := createCharacterInstance(card, characterDir)

	if *triggerEvent != "" {
		if err := handleTriggerEventFlag(char); err != nil {
			log.Fatalf("Failed to trigger event: %v", err)
		}
		return
	}

	networkManager := setupNetworkManager(char)
	if networkManager != nil {
		defer networkManager.Stop()
	}

	window := createDesktopWindow(myApp, char, profiler, networkManager)

	window.Show()
	myApp.Run()
}

// createCharacterInstance creates a new character from the given card and directory.
func createCharacterInstance(card *character.CharacterCard, characterDir string) *character.Character {
	char, err := character.New(card, characterDir)
	if err != nil {
		log.Fatalf("Failed to create character: %v", err)
	}
	return char
}

// setupNetworkManager creates and starts the network manager if networking is enabled.
func setupNetworkManager(char *character.Character) *network.NetworkManager {
	if !*networkMode {
		return nil
	}

	networkConfig := buildNetworkConfig(char)

	networkManager, err := network.NewNetworkManager(networkConfig)
	if err != nil {
		log.Fatalf("Failed to create network manager: %v", err)
	}

	if err := networkManager.Start(); err != nil {
		log.Fatalf("Failed to start network manager: %v", err)
	}

	if *debug {
		log.Printf("Network manager started with ID: %s", networkConfig.NetworkID)
	}

	return networkManager
}

// buildNetworkConfig creates network configuration using character settings and defaults.
func buildNetworkConfig(char *character.Character) network.NetworkManagerConfig {
	networkConfig := network.NetworkManagerConfig{
		DiscoveryPort:     8080, // Default port
		MaxPeers:          8,    // Default max peers
		NetworkID:         "dds-default-network",
		DiscoveryInterval: 5 * time.Second,
	}

	if char.GetCard() != nil && char.GetCard().HasMultiplayer() {
		mpConfig := char.GetCard().Multiplayer
		if mpConfig.DiscoveryPort > 0 {
			networkConfig.DiscoveryPort = mpConfig.DiscoveryPort
		}
		if mpConfig.MaxPeers > 0 {
			networkConfig.MaxPeers = mpConfig.MaxPeers
		}
		networkConfig.NetworkID = mpConfig.NetworkID
	}

	return networkConfig
}

// createDesktopWindow creates the desktop window with all required components.
func createDesktopWindow(myApp fyne.App, char *character.Character, profiler *monitoring.Profiler, networkManager *network.NetworkManager) *ui.DesktopWindow {
	window := ui.NewDesktopWindow(myApp, char, *debug, profiler, *gameMode, *showStats, networkManager, *networkMode, *showNetwork, *events)

	if *debug {
		log.Println("Created desktop window")
		if *events {
			log.Println("General events system enabled")
		}
	}

	return window
}

// handleTriggerEventFlag handles the -trigger-event command line flag
func handleTriggerEventFlag(char *character.Character) error {
	if *triggerEvent == "" {
		return nil
	}

	if *debug {
		log.Printf("Attempting to trigger event: %s", *triggerEvent)
	}

	// Get the general event manager from the character
	eventManager := char.GetGeneralEventManager()
	if eventManager == nil {
		return fmt.Errorf("general events system not available for this character")
	}

	// Try to trigger the specified event
	gameState := char.GetGameState()
	event, err := eventManager.TriggerEvent(*triggerEvent, gameState)
	if err != nil {
		return fmt.Errorf("failed to trigger event '%s': %w", *triggerEvent, err)
	}

	fmt.Printf("Successfully triggered event: %s\n", event.Name)
	fmt.Printf("Description: %s\n", event.Description)
	fmt.Printf("Category: %s\n", event.Category)

	if len(event.Responses) > 0 {
		fmt.Printf("Response: %s\n", event.Responses[0])
	}

	return nil
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
