package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/sirupsen/logrus"

	"github.com/opd-ai/desktop-companion/lib/character"
	"github.com/opd-ai/desktop-companion/lib/monitoring"
	"github.com/opd-ai/desktop-companion/lib/network"
	"github.com/opd-ai/desktop-companion/lib/ui"
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
func validateFlagDependencies(gameMode, showStats, networkMode, showNetwork, events bool, triggerEvent string) error {
	caller := getCaller()
	logrus.WithFields(logrus.Fields{
		"caller":       caller,
		"gameMode":     gameMode,
		"showStats":    showStats,
		"networkMode":  networkMode,
		"showNetwork":  showNetwork,
		"events":       events,
		"triggerEvent": triggerEvent,
	}).Info("Validating flag dependencies")

	if showStats && !gameMode {
		logrus.WithFields(logrus.Fields{
			"caller": caller,
			"error":  "-stats flag requires -game flag to be enabled",
		}).Error("Flag validation failed")
		return fmt.Errorf("-stats flag requires -game flag to be enabled")
	}
	if showNetwork && !networkMode {
		logrus.WithFields(logrus.Fields{
			"caller": caller,
			"error":  "-network-ui flag requires -network flag to be enabled",
		}).Error("Flag validation failed")
		return fmt.Errorf("-network-ui flag requires -network flag to be enabled")
	}
	if triggerEvent != "" && !events {
		logrus.WithFields(logrus.Fields{
			"caller": caller,
			"error":  "-trigger-event flag requires -events flag to be enabled",
		}).Error("Flag validation failed")
		return fmt.Errorf("-trigger-event flag requires -events flag to be enabled")
	}

	logrus.WithFields(logrus.Fields{
		"caller": caller,
	}).Info("Flag dependencies validation successful")
	return nil
}

// getCaller returns the calling function name for structured logging
func getCaller() string {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return "unknown"
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}
	return fn.Name()
}

func main() {
	caller := getCaller()
	logrus.WithFields(logrus.Fields{
		"caller": caller,
	}).Info("Starting desktop companion application")

	flag.Parse()

	logrus.WithFields(logrus.Fields{
		"caller": caller,
	}).Info("Command line flags parsed")

	// Validate flag dependencies
	if err := validateFlagDependencies(*gameMode, *showStats, *networkMode, *showNetwork, *events, *triggerEvent); err != nil {
		logrus.WithFields(logrus.Fields{
			"caller": caller,
			"error":  err.Error(),
		}).Error("Flag validation failed")
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Use -help for usage information\n")
		os.Exit(1)
	}

	if *version {
		logrus.WithFields(logrus.Fields{
			"caller": caller,
		}).Info("Version information requested")
		showVersionInfo()
		return
	}

	configureDebugLogging()

	logrus.WithFields(logrus.Fields{
		"caller": caller,
	}).Info("Debug logging configured")

	// Initialize performance profiler
	profiler := monitoring.NewProfiler(50) // 50MB memory target

	logrus.WithFields(logrus.Fields{
		"caller":       caller,
		"memoryTarget": 50,
	}).Info("Performance profiler initialized")

	// Start profiling if requested
	if err := profiler.Start(*memProfile, *cpuProfile, *debug); err != nil {
		logrus.WithFields(logrus.Fields{
			"caller":     caller,
			"memProfile": *memProfile,
			"cpuProfile": *cpuProfile,
			"debug":      *debug,
			"error":      err.Error(),
		}).Fatal("Failed to start profiler")
	}

	logrus.WithFields(logrus.Fields{
		"caller": caller,
	}).Info("Performance profiler started")

	defer func() {
		if err := profiler.Stop(*memProfile, *debug); err != nil {
			logrus.WithFields(logrus.Fields{
				"caller": caller,
				"error":  err.Error(),
			}).Error("Error stopping profiler")
		}
	}()

	// Load character configuration
	card, characterDir := loadCharacterConfiguration()

	logrus.WithFields(logrus.Fields{
		"caller":       caller,
		"characterDir": characterDir,
	}).Info("Character configuration loaded")

	// Record startup completion
	profiler.RecordStartupComplete()

	if *debug {
		stats := profiler.GetStats()
		logrus.WithFields(logrus.Fields{
			"caller":          caller,
			"startupDuration": stats.StartupDuration,
			"memoryUsageMB":   stats.CurrentMemoryMB,
		}).Info("Startup performance metrics")

		if !profiler.IsMemoryTargetMet() {
			logrus.WithFields(logrus.Fields{
				"caller":          caller,
				"currentMemoryMB": stats.CurrentMemoryMB,
				"targetMemoryMB":  50,
			}).Warn("Memory usage exceeds target")
		}
	}

	// Initialize application and show window
	runDesktopApplication(card, characterDir, profiler)

	logrus.WithFields(logrus.Fields{
		"caller": caller,
	}).Info("Desktop companion application completed")
}

// showVersionInfo displays application version information.
func showVersionInfo() {
	caller := getCaller()
	logrus.WithFields(logrus.Fields{
		"caller":     caller,
		"appVersion": appVersion,
	}).Info("Displaying version information")

	fmt.Printf("Desktop Companion v%s\n", appVersion)
	fmt.Println("Built with Go and Fyne - Cross-platform desktop pet")

	logrus.WithFields(logrus.Fields{
		"caller": caller,
	}).Info("Version information displayed")
}

// configureDebugLogging sets up debug logging if enabled.
func configureDebugLogging() {
	caller := getCaller()
	logrus.WithFields(logrus.Fields{
		"caller":       caller,
		"debugEnabled": *debug,
	}).Info("Configuring debug logging")

	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.SetReportCaller(true)
		log.SetFlags(log.LstdFlags | log.Lshortfile)

		logrus.WithFields(logrus.Fields{
			"caller": caller,
		}).Info("Debug mode enabled with caller information")
	} else {
		logrus.SetLevel(logrus.InfoLevel)
		logrus.WithFields(logrus.Fields{
			"caller": caller,
		}).Info("Standard logging level set")
	}
}

// resolveProjectRoot finds the project root by searching upward for go.mod file.
// For deployed binaries, falls back to executable directory if assets/ exists there.
func resolveProjectRoot() string {
	caller := getCaller()
	logrus.WithFields(logrus.Fields{
		"caller": caller,
	}).Info("Resolving project root directory")

	execPath, execErr := os.Executable()
	if execErr != nil {
		logrus.WithFields(logrus.Fields{
			"caller": caller,
			"error":  execErr.Error(),
		}).Fatal("Failed to get executable path")
	}

	logrus.WithFields(logrus.Fields{
		"caller":   caller,
		"execPath": execPath,
	}).Debug("Executable path obtained")

	searchDir := filepath.Dir(execPath)

	logrus.WithFields(logrus.Fields{
		"caller":    caller,
		"searchDir": searchDir,
	}).Debug("Starting upward search for go.mod")

	// Search upward for go.mod file (development environment)
	for {
		goModPath := filepath.Join(searchDir, "go.mod")
		logrus.WithFields(logrus.Fields{
			"caller":    caller,
			"goModPath": goModPath,
		}).Debug("Checking for go.mod file")

		if _, statErr := os.Stat(goModPath); statErr == nil {
			logrus.WithFields(logrus.Fields{
				"caller":      caller,
				"projectRoot": searchDir,
			}).Info("Project root found via go.mod")
			return searchDir
		}
		parent := filepath.Dir(searchDir)
		if parent == searchDir {
			logrus.WithFields(logrus.Fields{
				"caller": caller,
			}).Debug("Reached filesystem root, no go.mod found")
			break // Reached filesystem root
		}
		searchDir = parent
	}

	// No go.mod found - check if this is a deployed binary with assets/ directory
	execDir := filepath.Dir(execPath)
	assetsPath := filepath.Join(execDir, "assets")

	logrus.WithFields(logrus.Fields{
		"caller":     caller,
		"assetsPath": assetsPath,
	}).Debug("Checking for assets directory relative to executable")

	if _, err := os.Stat(assetsPath); err == nil {
		// assets/ directory exists relative to executable - use executable directory
		logrus.WithFields(logrus.Fields{
			"caller":      caller,
			"projectRoot": execDir,
		}).Info("Project root found via assets directory")
		return execDir
	}

	// Fallback to executable directory (preserves existing behavior)
	logrus.WithFields(logrus.Fields{
		"caller":      caller,
		"projectRoot": execDir,
	}).Info("Using executable directory as fallback project root")
	return execDir
}

// resolveCharacterPath converts the character path to an absolute path.
func resolveCharacterPath() string {
	caller := getCaller()
	logrus.WithFields(logrus.Fields{
		"caller":        caller,
		"characterPath": *characterPath,
	}).Info("Resolving character path")

	// For the default relative path, resolve relative to project root (for development)
	// or executable directory (for deployed binaries)
	if *characterPath == "assets/characters/default/character.json" && !filepath.IsAbs(*characterPath) {
		logrus.WithFields(logrus.Fields{
			"caller": caller,
		}).Debug("Using default relative character path")

		projectRoot := resolveProjectRoot()
		absPath := filepath.Join(projectRoot, *characterPath)

		logrus.WithFields(logrus.Fields{
			"caller":      caller,
			"projectRoot": projectRoot,
			"absPath":     absPath,
		}).Info("Character path resolved relative to project root")
		return absPath
	}

	absPath, err := filepath.Abs(*characterPath)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"caller":        caller,
			"characterPath": *characterPath,
			"error":         err.Error(),
		}).Fatal("Failed to resolve character path")
	}

	logrus.WithFields(logrus.Fields{
		"caller":  caller,
		"absPath": absPath,
	}).Info("Character path resolved to absolute path")
	return absPath
}

// loadAndValidateCharacter loads the character card from the given path with debug logging.
func loadAndValidateCharacter(absPath string) *character.CharacterCard {
	caller := getCaller()
	logrus.WithFields(logrus.Fields{
		"caller":  caller,
		"absPath": absPath,
	}).Info("Loading character card from path")

	if *debug {
		logrus.WithFields(logrus.Fields{
			"caller": caller,
			"path":   absPath,
		}).Debug("Debug mode: Loading character from path")
	}

	card, err := character.LoadCard(absPath)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"caller":  caller,
			"absPath": absPath,
			"error":   err.Error(),
		}).Fatal("Failed to load character card")
	}

	logrus.WithFields(logrus.Fields{
		"caller":      caller,
		"name":        card.Name,
		"description": card.Description,
	}).Info("Character card loaded successfully")

	if *debug {
		logrus.WithFields(logrus.Fields{
			"caller":      caller,
			"name":        card.Name,
			"description": card.Description,
		}).Debug("Debug mode: Character details")
	}

	return card
}

// loadCharacterConfiguration loads and validates the character configuration file.
func loadCharacterConfiguration() (*character.CharacterCard, string) {
	caller := getCaller()
	logrus.WithFields(logrus.Fields{
		"caller": caller,
	}).Info("Loading character configuration")

	absPath := resolveCharacterPath()
	card := loadAndValidateCharacter(absPath)
	characterDir := filepath.Dir(absPath)

	logrus.WithFields(logrus.Fields{
		"caller":       caller,
		"characterDir": characterDir,
	}).Info("Character configuration loading completed")

	return card, characterDir
}

// runDesktopApplication creates and runs the desktop companion application.
func runDesktopApplication(card *character.CharacterCard, characterDir string, profiler *monitoring.Profiler) {
	caller := getCaller()
	logrus.WithFields(logrus.Fields{
		"caller":       caller,
		"characterDir": characterDir,
	}).Info("Starting desktop application")

	if err := checkDisplayAvailable(); err != nil {
		logrus.WithFields(logrus.Fields{
			"caller": caller,
			"error":  err.Error(),
		}).Fatal("Cannot run desktop application - display not available")
	}

	logrus.WithFields(logrus.Fields{
		"caller": caller,
	}).Debug("Display availability check passed")

	myApp := app.NewWithID("com.opdai.desktop-companion")

	logrus.WithFields(logrus.Fields{
		"caller": caller,
		"appID":  "com.opdai.desktop-companion",
	}).Info("Fyne application created")

	char := createCharacterInstance(card, characterDir)

	if *triggerEvent != "" {
		logrus.WithFields(logrus.Fields{
			"caller":       caller,
			"triggerEvent": *triggerEvent,
		}).Info("Trigger event mode requested")

		if err := handleTriggerEventFlag(char); err != nil {
			logrus.WithFields(logrus.Fields{
				"caller":       caller,
				"triggerEvent": *triggerEvent,
				"error":        err.Error(),
			}).Fatal("Failed to trigger event")
		}

		logrus.WithFields(logrus.Fields{
			"caller": caller,
		}).Info("Event trigger completed, exiting")
		return
	}

	networkManager := setupNetworkManager(char)
	if networkManager != nil {
		defer func() {
			logrus.WithFields(logrus.Fields{
				"caller": caller,
			}).Info("Stopping network manager")
			networkManager.Stop()
		}()
	}

	window := createDesktopWindow(myApp, char, profiler, networkManager)

	logrus.WithFields(logrus.Fields{
		"caller": caller,
	}).Info("Desktop window created, showing application")

	window.Show()
	myApp.Run()

	logrus.WithFields(logrus.Fields{
		"caller": caller,
	}).Info("Desktop application completed")
}

// createCharacterInstance creates a new character from the given card and directory.
func createCharacterInstance(card *character.CharacterCard, characterDir string) *character.Character {
	caller := getCaller()
	logrus.WithFields(logrus.Fields{
		"caller":        caller,
		"characterDir":  characterDir,
		"characterName": card.Name,
	}).Info("Creating character instance")

	char, err := character.New(card, characterDir)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"caller":        caller,
			"characterDir":  characterDir,
			"characterName": card.Name,
			"error":         err.Error(),
		}).Fatal("Failed to create character")
	}

	logrus.WithFields(logrus.Fields{
		"caller":        caller,
		"characterName": card.Name,
	}).Info("Character instance created successfully")

	return char
}

// setupNetworkManager creates and starts the network manager if networking is enabled.
func setupNetworkManager(char *character.Character) *network.NetworkManager {
	caller := getCaller()
	logrus.WithFields(logrus.Fields{
		"caller":      caller,
		"networkMode": *networkMode,
	}).Info("Setting up network manager")

	if !*networkMode {
		logrus.WithFields(logrus.Fields{
			"caller": caller,
		}).Info("Network mode disabled, skipping network manager setup")
		return nil
	}

	logrus.WithFields(logrus.Fields{
		"caller": caller,
	}).Info("Network mode enabled, configuring network manager")

	networkConfig := buildNetworkConfig(char)

	networkManager, err := network.NewNetworkManager(*networkConfig)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"caller": caller,
			"error":  err.Error(),
		}).Fatal("Failed to create network manager")
	}

	logrus.WithFields(logrus.Fields{
		"caller": caller,
	}).Info("Network manager created, starting networking")

	if err := networkManager.Start(); err != nil {
		logrus.WithFields(logrus.Fields{
			"caller": caller,
			"error":  err.Error(),
		}).Fatal("Failed to start network manager")
	}

	if *debug {
		logrus.WithFields(logrus.Fields{
			"caller":    caller,
			"networkID": networkConfig.NetworkID,
		}).Debug("Network manager started with configuration")
	}

	logrus.WithFields(logrus.Fields{
		"caller":    caller,
		"networkID": networkConfig.NetworkID,
	}).Info("Network manager started successfully")

	return networkManager
}

// buildNetworkConfig creates network configuration using character settings and defaults.
func buildNetworkConfig(char *character.Character) *network.NetworkManagerConfig {
	caller := getCaller()
	logrus.WithFields(logrus.Fields{
		"caller": caller,
	}).Info("Building network configuration")

	networkConfig := &network.NetworkManagerConfig{
		DiscoveryPort: 8080,
		MaxPeers:      10,
		NetworkID:     "default-network",
	}

	logrus.WithFields(logrus.Fields{
		"caller":        caller,
		"discoveryPort": networkConfig.DiscoveryPort,
		"maxPeers":      networkConfig.MaxPeers,
		"networkID":     networkConfig.NetworkID,
	}).Debug("Default network configuration set")

	if char.GetCard() != nil && char.GetCard().HasMultiplayer() {
		logrus.WithFields(logrus.Fields{
			"caller": caller,
		}).Info("Character has multiplayer configuration, applying settings")

		mpConfig := char.GetCard().Multiplayer

		if mpConfig.DiscoveryPort > 0 {
			oldPort := networkConfig.DiscoveryPort
			networkConfig.DiscoveryPort = mpConfig.DiscoveryPort
			logrus.WithFields(logrus.Fields{
				"caller":  caller,
				"oldPort": oldPort,
				"newPort": mpConfig.DiscoveryPort,
			}).Debug("Discovery port updated from character config")
		}

		if mpConfig.MaxPeers > 0 {
			oldMaxPeers := networkConfig.MaxPeers
			networkConfig.MaxPeers = mpConfig.MaxPeers
			logrus.WithFields(logrus.Fields{
				"caller":      caller,
				"oldMaxPeers": oldMaxPeers,
				"newMaxPeers": mpConfig.MaxPeers,
			}).Debug("Max peers updated from character config")
		}

		if mpConfig.NetworkID != "" {
			oldNetworkID := networkConfig.NetworkID
			networkConfig.NetworkID = mpConfig.NetworkID
			logrus.WithFields(logrus.Fields{
				"caller":       caller,
				"oldNetworkID": oldNetworkID,
				"newNetworkID": mpConfig.NetworkID,
			}).Debug("Network ID updated from character config")
		}
	}

	logrus.WithFields(logrus.Fields{
		"caller":        caller,
		"discoveryPort": networkConfig.DiscoveryPort,
		"maxPeers":      networkConfig.MaxPeers,
		"networkID":     networkConfig.NetworkID,
	}).Info("Network configuration built")

	return networkConfig
}

// createDesktopWindow creates the desktop window with all required components.
func createDesktopWindow(myApp fyne.App, char *character.Character, profiler *monitoring.Profiler, networkManager *network.NetworkManager) *ui.DesktopWindow {
	caller := getCaller()
	logrus.WithFields(logrus.Fields{
		"caller":        caller,
		"gameMode":      *gameMode,
		"showStats":     *showStats,
		"networkMode":   *networkMode,
		"showNetwork":   *showNetwork,
		"eventsEnabled": *events,
	}).Info("Creating desktop window")

	window := ui.NewDesktopWindow(myApp, char, *debug, profiler, *gameMode, *showStats, networkManager, *networkMode, *showNetwork, *events)

	logrus.WithFields(logrus.Fields{
		"caller": caller,
	}).Info("Desktop window created successfully")

	if *debug {
		logrus.WithFields(logrus.Fields{
			"caller": caller,
		}).Debug("Debug mode: Desktop window created")

		if *events {
			logrus.WithFields(logrus.Fields{
				"caller": caller,
			}).Debug("Debug mode: General events system enabled")
		}
	}

	return window
}

// handleTriggerEventFlag handles the -trigger-event command line flag
func handleTriggerEventFlag(char *character.Character) error {
	caller := getCaller()
	logrus.WithFields(logrus.Fields{
		"caller":       caller,
		"triggerEvent": *triggerEvent,
	}).Info("Handling trigger event flag")

	if *triggerEvent == "" {
		logrus.WithFields(logrus.Fields{
			"caller": caller,
		}).Debug("No trigger event specified")
		return nil
	}

	if *debug {
		logrus.WithFields(logrus.Fields{
			"caller":       caller,
			"triggerEvent": *triggerEvent,
		}).Debug("Debug mode: Attempting to trigger event")
	}

	// Get the general event manager from the character
	eventManager := char.GetGeneralEventManager()
	if eventManager == nil {
		logrus.WithFields(logrus.Fields{
			"caller": caller,
		}).Error("General events system not available for this character")
		return fmt.Errorf("general events system not available for this character")
	}

	logrus.WithFields(logrus.Fields{
		"caller": caller,
	}).Debug("General event manager obtained")

	// Try to trigger the specified event
	gameState := char.GetGameState()

	logrus.WithFields(logrus.Fields{
		"caller":       caller,
		"triggerEvent": *triggerEvent,
	}).Info("Triggering specified event")

	event, err := eventManager.TriggerEvent(*triggerEvent, gameState)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"caller":       caller,
			"triggerEvent": *triggerEvent,
			"error":        err.Error(),
		}).Error("Failed to trigger event")
		return fmt.Errorf("failed to trigger event '%s': %w", *triggerEvent, err)
	}

	logrus.WithFields(logrus.Fields{
		"caller":      caller,
		"eventName":   event.Name,
		"description": event.Description,
		"category":    event.Category,
	}).Info("Event triggered successfully")

	fmt.Printf("Successfully triggered event: %s\n", event.Name)
	fmt.Printf("Description: %s\n", event.Description)
	fmt.Printf("Category: %s\n", event.Category)

	if len(event.Responses) > 0 {
		logrus.WithFields(logrus.Fields{
			"caller":        caller,
			"responseCount": len(event.Responses),
			"firstResponse": event.Responses[0],
		}).Debug("Event responses available")
		fmt.Printf("Response: %s\n", event.Responses[0])
	}

	return nil
}

// checkDisplayAvailable verifies that a display is available for GUI applications
func checkDisplayAvailable() error {
	caller := getCaller()
	logrus.WithFields(logrus.Fields{
		"caller": caller,
	}).Info("Checking display availability for GUI application")

	// Check for X11 display on Linux/Unix systems
	display := os.Getenv("DISPLAY")
	waylandDisplay := os.Getenv("WAYLAND_DISPLAY")

	logrus.WithFields(logrus.Fields{
		"caller":         caller,
		"display":        display,
		"waylandDisplay": waylandDisplay,
	}).Debug("Environment display variables")

	// Check if any display environment is available
	if display == "" && waylandDisplay == "" {
		logrus.WithFields(logrus.Fields{
			"caller": caller,
		}).Error("No display environment available")
		return fmt.Errorf("no display available - neither X11 (DISPLAY) nor Wayland (WAYLAND_DISPLAY) environment is available.\n" +
			"This application requires a graphical desktop environment to run.\n" +
			"Please run from a desktop session or use X11 forwarding for remote connections")
	}

	logrus.WithFields(logrus.Fields{
		"caller": caller,
	}).Info("Display environment available")

	// Additional detection for headless systems and SSH sessions
	sshConnection := os.Getenv("SSH_CONNECTION")
	sshClient := os.Getenv("SSH_CLIENT")

	if sshConnection != "" || sshClient != "" {
		logrus.WithFields(logrus.Fields{
			"caller":        caller,
			"sshConnection": sshConnection,
			"sshClient":     sshClient,
		}).Warn("Running in SSH session - GUI may not be available")
	}

	// Check for headless system indicators
	xdgDesktop := os.Getenv("XDG_CURRENT_DESKTOP")
	desktopSession := os.Getenv("DESKTOP_SESSION")

	logrus.WithFields(logrus.Fields{
		"caller":         caller,
		"xdgDesktop":     xdgDesktop,
		"desktopSession": desktopSession,
	}).Debug("Desktop environment variables")

	if xdgDesktop == "" && desktopSession == "" && display != "" {
		logrus.WithFields(logrus.Fields{
			"caller": caller,
		}).Warn("No desktop environment detected - running in minimal graphics mode")
	}

	logrus.WithFields(logrus.Fields{
		"caller": caller,
	}).Info("Display availability check completed successfully")

	return nil
}
