package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"fyne.io/fyne/v2/app"

	"desktop-companion/internal/character"
	"desktop-companion/internal/ui"
)

var (
	characterPath = flag.String("character", "assets/characters/default/character.json", "Path to character configuration file")
	debug         = flag.Bool("debug", false, "Enable debug logging")
	version       = flag.Bool("version", false, "Show version information")
)

const appVersion = "1.0.0"

func main() {
	flag.Parse()

	if *version {
		showVersionInfo()
		return
	}

	configureDebugLogging()

	// Load character configuration
	card, characterDir := loadCharacterConfiguration()

	// Initialize application and show window
	runDesktopApplication(card, characterDir)
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
	absPath, err := filepath.Abs(*characterPath)
	if err != nil {
		log.Fatalf("Failed to resolve character path: %v", err)
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
func runDesktopApplication(card *character.CharacterCard, characterDir string) {
	// Create Fyne application
	myApp := app.NewWithID("com.opdai.desktop-companion")

	// Create character instance
	char, err := character.New(card, characterDir)
	if err != nil {
		log.Fatalf("Failed to create character: %v", err)
	}

	// Create and show desktop window
	window := ui.NewDesktopWindow(myApp, char, *debug)

	if *debug {
		log.Println("Created desktop window")
	}

	// Show window and start event loop
	window.Show()
	myApp.Run()
}
