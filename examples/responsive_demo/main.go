package main

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/opd-ai/desktop-companion/internal/platform"
	"github.com/opd-ai/desktop-companion/internal/ui/responsive"
)

// ResponsiveDemo demonstrates the responsive layout system
func main() {
	log.Println("Starting Responsive Layout Demo...")

	// Create Fyne application
	myApp := app.New()
	myApp.SetIcon(nil)

	// Get platform information
	platformInfo := platform.GetPlatformInfo()
	log.Printf("Platform: %s, Form Factor: %s, Input Methods: %v",
		platformInfo.OS, platformInfo.FormFactor, platformInfo.InputMethods)

	// Create responsive layout
	layout := responsive.NewLayout(platformInfo, myApp)
	config := layout.GetWindowConfig(128)
	log.Printf("Window mode: %s, Character size: %d", config.Mode, config.CharacterSize)

	// Create main window with responsive configuration
	window := myApp.NewWindow("Responsive Layout Demo")
	window.Resize(config.WindowSize)
	window.SetFixedSize(!config.Resizable)

	// Create character placeholder
	characterSize := layout.GetCharacterSize(128)
	character := widget.NewButton("🐱", func() {
		log.Println("Character tapped!")
	})
	character.Resize(fyne.NewSize(float32(characterSize), float32(characterSize)))

	// Create info display
	infoText := widget.NewRichTextFromMarkdown(`
# Responsive Layout Demo

**Platform:** ` + platformInfo.OS + `  
**Form Factor:** ` + platformInfo.FormFactor + `  
**Window Mode:** ` + string(config.Mode) + `  
**Character Size:** ` + formatInt(config.CharacterSize) + `px  
**Touch Target Size:** ` + formatInt(layout.GetTouchTargetSize()) + `px  

` + getLayoutDescription(platformInfo, config))

	// Create mobile window manager for mobile platforms
	if platformInfo.IsMobile() {
		log.Println("Configuring mobile window manager...")
		mwm := responsive.NewMobileWindowManager(platformInfo, layout)

		// Configure window for mobile
		err := mwm.ConfigureWindow(window)
		if err != nil {
			log.Printf("Error configuring mobile window: %v", err)
		}

		// Create mobile controls
		controlBar := responsive.NewMobileControlBar(platformInfo)
		controlBar.SetStatsCallback(func() {
			log.Println("Stats button tapped!")
		})
		controlBar.SetChatCallback(func() {
			log.Println("Chat button tapped!")
		})
		controlBar.SetNetworkCallback(func() {
			log.Println("Network button tapped!")
		})
		controlBar.SetMenuCallback(func() {
			log.Println("Menu button tapped!")
		})

		// Layout for mobile with controls
		content := container.NewVBox(
			container.NewCenter(character),
			infoText,
			controlBar.GetContainer(),
		)

		mwm.SetContent(content)
		log.Printf("Mobile window manager mode: %s", mwm.GetCurrentMode())
	} else {
		// Desktop layout
		log.Println("Configuring desktop layout...")
		content := container.NewVBox(
			container.NewCenter(character),
			infoText,
		)
		window.SetContent(content)

		// Apply desktop-specific settings
		if config.AlwaysOnTop {
			log.Println("Setting always on top (desktop overlay mode)")
		}
	}

	// Position window optimally
	optimalPos := layout.GetOptimalPosition(config.WindowSize)
	log.Printf("Optimal window position: %.0f,%.0f", optimalPos.X, optimalPos.Y)

	// Show mobile controls information if applicable
	if layout.ShouldShowMobileControls() {
		log.Println("Mobile controls enabled")
	}

	log.Println("Demo ready! Check the window for responsive behavior.")
	window.ShowAndRun()
}

// Helper function to format integers as strings
func formatInt(i int) string {
	return fmt.Sprintf("%d", i)
}

// GetLayoutDescription provides human-readable layout information
func getLayoutDescription(platform *platform.PlatformInfo, config *responsive.WindowConfig) string {
	if platform.IsMobile() {
		return `
**Mobile Layout Features:**
- Fullscreen application mode
- Touch-friendly control buttons
- 25% screen width character sizing
- Picture-in-Picture support ready
- Haptic feedback capable`
	} else {
		return `
**Desktop Layout Features:**
- Overlay window mode
- Always-on-top behavior
- Fixed character sizing
- Mouse and keyboard optimized
- Transparent background support`
	}
}
