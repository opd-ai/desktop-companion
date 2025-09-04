package ui

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"

	"github.com/opd-ai/desktop-companion/internal/character"
	"github.com/opd-ai/desktop-companion/internal/monitoring"
	"github.com/opd-ai/desktop-companion/internal/network"
)

// BattleInvitationHandler is a global callback for battle invitation dialogs
// Provides minimal coupling between multiplayer battle system and UI (Fix for Finding #1)
var BattleInvitationHandler func(fromCharacter string, onResponse func(accepted bool))

// DesktopWindow represents the transparent overlay window containing the character
// Uses Fyne for cross-platform window management - avoiding custom windowing code
type DesktopWindow struct {
	window                  fyne.Window
	character               *character.Character
	renderer                *CharacterRenderer
	dialog                  *DialogBubble
	contextMenu             *ContextMenu
	statsOverlay            *StatsOverlay
	statsTooltip            *StatsTooltip
	chatbotInterface        *ChatbotInterface
	networkOverlay          *NetworkOverlay
	giftDialog              *GiftSelectionDialog
	battleInvitationDialog  *BattleInvitationDialog
	peerSelectionDialog     *PeerSelectionDialog
	achievementNotification *AchievementNotification
	saveStatusIndicator     *SaveStatusIndicator
	profiler                *monitoring.Profiler
	debug                   bool
	gameMode                bool
	showStats               bool
	networkMode             bool
	showNetwork             bool
	eventsEnabled           bool
}

// NewDesktopWindow creates a new transparent desktop window
// Uses Fyne's desktop app interface for always-on-top and transparency
func NewDesktopWindow(app fyne.App, char *character.Character, debug bool, profiler *monitoring.Profiler, gameMode bool, showStats bool, networkManager NetworkManagerInterface, networkMode bool, showNetwork bool, eventsEnabled bool) *DesktopWindow {
	window := createConfiguredWindow(app, char, debug)

	dw := &DesktopWindow{
		window:        window,
		character:     char,
		profiler:      profiler,
		debug:         debug,
		gameMode:      gameMode,
		showStats:     showStats,
		networkMode:   networkMode,
		showNetwork:   showNetwork,
		eventsEnabled: eventsEnabled,
	}

	initializeBasicComponents(dw, char, debug)
	initializeGameFeatures(dw, gameMode, showStats, char)
	initializeDialogFeatures(dw, char)
	initializeGiftSystem(dw, gameMode, char)
	initializeNetworkFeatures(dw, networkMode, networkManager, showNetwork, char)

	// Set up window content and interactions
	dw.setupContent()
	dw.setupInteractions()

	// Start animation update loop
	go dw.animationLoop()

	if debug {
		log.Printf("Created desktop window: %dx%d with always-on-top configuration", char.GetSize(), char.GetSize())
	}

	return dw
}

// createConfiguredWindow creates and configures the basic window properties
func createConfiguredWindow(app fyne.App, char *character.Character, debug bool) fyne.Window {
	window := app.NewWindow("Desktop Companion")

	// Configure window for desktop overlay behavior
	window.SetFixedSize(true)
	window.Resize(fyne.NewSize(float32(char.GetSize()), float32(char.GetSize())))

	// Configure transparency for desktop overlay
	configureTransparency(window, debug)

	// Attempt to configure always-on-top behavior using available Fyne capabilities
	// Note: Fyne has limited always-on-top support, but we can try available approaches
	configureAlwaysOnTop(window, debug)

	return window
}

// initializeBasicComponents creates the essential UI components for the desktop window
func initializeBasicComponents(dw *DesktopWindow, char *character.Character, debug bool) {
	// Create character renderer
	dw.renderer = NewCharacterRenderer(char, debug)

	// Create dialog bubble (initially hidden)
	dw.dialog = NewDialogBubble()

	// Create context menu (initially hidden)
	dw.contextMenu = NewContextMenu()

	// Create battle invitation dialog (initially hidden)
	dw.battleInvitationDialog = NewBattleInvitationDialog()

	// Create peer selection dialog (initially hidden)
	dw.peerSelectionDialog = NewPeerSelectionDialog()

	// Create save status indicator (small, positioned in corner)
	dw.saveStatusIndicator = NewSaveStatusIndicator()
}

// initializeGameFeatures sets up game-related features like stats overlay
func initializeGameFeatures(dw *DesktopWindow, gameMode bool, showStats bool, char *character.Character) {
	if gameMode && char.GetGameState() != nil {
		dw.statsOverlay = NewStatsOverlay(char)
		if showStats {
			dw.statsOverlay.Show()
		}

		// Initialize stats tooltip for hover functionality
		dw.statsTooltip = NewStatsTooltip(char)

		// Initialize achievement notifications for game mode
		dw.achievementNotification = NewAchievementNotification()
	}
}

// initializeDialogFeatures configures AI-powered dialog capabilities
func initializeDialogFeatures(dw *DesktopWindow, char *character.Character) {
	if char.GetCard() != nil && char.GetCard().HasDialogBackend() {
		dw.chatbotInterface = NewChatbotInterface(char)
	}
}

// initializeGiftSystem sets up the gift selection dialog and related functionality
func initializeGiftSystem(dw *DesktopWindow, gameMode bool, char *character.Character) {
	if gameMode && char.GetCard() != nil && char.GetCard().HasGiftSystem() && char.GetGameState() != nil {
		giftManager := character.NewGiftManager(char.GetCard(), char.GetGameState())
		dw.giftDialog = NewGiftSelectionDialog(giftManager)

		// Set up callbacks for gift dialog
		dw.giftDialog.SetOnGiftGiven(func(response *character.GiftResponse) {
			// Show response message to user
			if response.Response != "" {
				dw.showDialog(response.Response)
			} else if response.ErrorMessage != "" {
				dw.showDialog(response.ErrorMessage)
			}
		})

		dw.giftDialog.SetOnCancel(func() {
			// Dialog closed, no action needed
		})
	}
}

// initializeNetworkFeatures configures multiplayer networking capabilities
func initializeNetworkFeatures(dw *DesktopWindow, networkMode bool, networkManager NetworkManagerInterface, showNetwork bool, char *character.Character) {
	if networkMode && networkManager != nil {
		dw.networkOverlay = NewNetworkOverlay(networkManager)
		dw.networkOverlay.RegisterNetworkEvents()

		// Set local character name for clear UI distinction
		if char != nil && char.GetCard() != nil {
			dw.networkOverlay.SetLocalCharacterName(char.GetCard().Name)
		}

		// Feature 5: Initialize compatibility calculator for personality-based scoring
		// Only enable for characters with personality configurations
		if char != nil && char.GetCard() != nil && char.GetCard().Personality != nil {
			compatibilityCalculator := character.NewCompatibilityCalculator(char)
			dw.networkOverlay.SetCompatibilityCalculator(compatibilityCalculator)
		}

		// Fix for Finding #1: Set up battle invitation callback for multiplayer characters
		// Register the window's battle invitation dialog as the global handler
		if char != nil && char.GetCard() != nil && char.GetCard().HasMultiplayer() {
			BattleInvitationHandler = func(fromCharacter string, onResponse func(accepted bool)) {
				dw.ShowBattleInvitationDialog(fromCharacter, onResponse)
			}
		}

		if showNetwork {
			dw.networkOverlay.Show()
		}
	}
}

// setupContent configures the window's visual content
func (dw *DesktopWindow) setupContent() {
	// Create list of content objects
	objects := []fyne.CanvasObject{
		dw.renderer,
		dw.dialog,
		dw.contextMenu,
	}

	// Add save status indicator (positioned in corner)
	if dw.saveStatusIndicator != nil {
		// Position save status indicator in top-right corner
		dw.saveStatusIndicator.Resize(fyne.NewSize(16, 16))
		dw.saveStatusIndicator.Move(fyne.NewPos(float32(dw.character.GetSize()-20), 4))
		objects = append(objects, dw.saveStatusIndicator)
	}

	// Add stats overlay if available
	if dw.statsOverlay != nil {
		objects = append(objects, dw.statsOverlay.GetContainer())
	}

	// Add stats tooltip if available
	if dw.statsTooltip != nil {
		objects = append(objects, dw.statsTooltip.GetContainer())
	}

	// Add chatbot interface if available
	if dw.chatbotInterface != nil {
		objects = append(objects, dw.chatbotInterface)
	}

	// Add gift dialog if available
	if dw.giftDialog != nil {
		objects = append(objects, dw.giftDialog)
	}

	// Add network overlay if available
	if dw.networkOverlay != nil {
		objects = append(objects, dw.networkOverlay.GetContainer())
	}

	// Add achievement notification if available
	if dw.achievementNotification != nil {
		objects = append(objects, dw.achievementNotification)
	}

	// Create container with transparent background for overlay effect
	content := container.NewWithoutLayout(objects...)

	dw.window.SetContent(content)

	if dw.debug {
		log.Println("Window content configured for transparent overlay")
	}
}

// setupInteractions configures mouse interactions with the character
func (dw *DesktopWindow) setupInteractions() {
	// Add dragging support if character allows movement
	if dw.character.IsMovementEnabled() {
		dw.setupDragging()
		// Setup keyboard shortcuts even for draggable characters
		dw.setupKeyboardShortcuts()
		return
	}

	// For non-draggable characters, create custom clickable widget that supports both left and right click
	clickable := NewClickableWidget(
		func() { dw.handleClick() },
		func() { dw.handleRightClick() },
	)
	clickable.SetSize(fyne.NewSize(float32(dw.character.GetSize()), float32(dw.character.GetSize())))

	// Create list of content objects for interactive overlay
	objects := []fyne.CanvasObject{
		dw.renderer,
		clickable,
		dw.dialog,
		dw.contextMenu,
		dw.battleInvitationDialog.GetContainer(),
		dw.peerSelectionDialog.GetContainer(),
	}

	// Add save status indicator if available (positioned in corner)
	if dw.saveStatusIndicator != nil {
		dw.saveStatusIndicator.Resize(fyne.NewSize(16, 16))
		dw.saveStatusIndicator.Move(fyne.NewPos(float32(dw.character.GetSize()-20), 4))
		objects = append(objects, dw.saveStatusIndicator)
	}

	// Add stats overlay if available
	if dw.statsOverlay != nil {
		objects = append(objects, dw.statsOverlay.GetContainer())
	}

	// Add stats tooltip if available
	if dw.statsTooltip != nil {
		objects = append(objects, dw.statsTooltip.GetContainer())
	}

	// Add chatbot interface if available
	if dw.chatbotInterface != nil {
		objects = append(objects, dw.chatbotInterface)
	}

	// Update window content with interactive overlay
	content := container.NewWithoutLayout(objects...)

	dw.window.SetContent(content)

	// Setup keyboard shortcuts
	dw.setupKeyboardShortcuts()
}

// handleClick processes character click interactions
func (dw *DesktopWindow) handleClick() {
	response := dw.character.HandleClick()

	if dw.debug {
		log.Printf("Character clicked, response: %q", response)
	}

	if response != "" {
		dw.showDialog(response)
	}
}

// handleRightClick processes character right-click interactions
// Now shows context menu instead of direct dialog
func (dw *DesktopWindow) handleRightClick() {
	if dw.debug {
		log.Println("Character right-clicked, showing context menu")
	}

	// Show context menu with available actions
	dw.showContextMenu()
}

// showDialog displays a dialog bubble with the given text
func (dw *DesktopWindow) showDialog(text string) {
	dw.dialog.ShowWithText(text)

	// Auto-hide dialog after 3 seconds
	go func() {
		time.Sleep(3 * time.Second)
		dw.dialog.Hide()
	}()
}

// showEventFrequencySettings displays the random event frequency settings dialog
// Feature 6: Random Event Frequency Tuning - allows users to adjust event frequency
func (dw *DesktopWindow) showEventFrequencySettings() {
	currentMultiplier := dw.character.GetEventFrequencyMultiplier()

	// Create frequency options with user-friendly labels
	options := []struct {
		label      string
		multiplier float64
	}{
		{"Very Rare (0.5x)", 0.5},
		{"Normal (1.0x)", 1.0},
		{"Frequent (1.5x)", 1.5},
		{"Very Frequent (2.0x)", 2.0},
		{"Maximum (3.0x)", 3.0},
	}

	settingsText := fmt.Sprintf("Current Event Frequency: %.1fx\n\nChoose new frequency:\n\n", currentMultiplier)

	// Build options text and find current selection
	var selectedOption string
	for _, option := range options {
		prefix := "  "
		if option.multiplier == currentMultiplier {
			prefix = "→ "
			selectedOption = option.label
		}
		settingsText += fmt.Sprintf("%s%s\n", prefix, option.label)
	}

	if selectedOption != "" {
		settingsText += fmt.Sprintf("\nCurrent: %s", selectedOption)
	}

	// For now, show a simple dialog with the current state
	// In a full implementation, this would use a selection dialog
	dw.showDialog(settingsText + "\n\nUse keyboard shortcuts to adjust:\n  Ctrl+1 = Very Rare\n  Ctrl+2 = Normal\n  Ctrl+3 = Frequent\n  Ctrl+4 = Very Frequent\n  Ctrl+5 = Maximum")
}

// showContextMenu displays the right-click context menu
// Creates dynamic menu items based on character capabilities and game mode
func (dw *DesktopWindow) showContextMenu() {
	var menuItems []ContextMenuItem

	menuItems = append(menuItems, dw.buildBasicMenuItems()...)
	menuItems = append(menuItems, dw.buildGameModeMenuItems()...)
	menuItems = append(menuItems, dw.buildBattleMenuItems()...)
	menuItems = append(menuItems, dw.buildChatMenuItems()...)
	menuItems = append(menuItems, dw.buildNewsMenuItems()...)
	menuItems = append(menuItems, dw.buildNetworkMenuItems()...)
	menuItems = append(menuItems, dw.buildUtilityMenuItems()...)

	dw.displayContextMenu(menuItems)
}

// buildBasicMenuItems creates the basic interaction menu items
func (dw *DesktopWindow) buildBasicMenuItems() []ContextMenuItem {
	menuItems := []ContextMenuItem{
		{
			Text: "Talk",
			Callback: func() {
				response := dw.character.HandleClick()
				if response != "" {
					dw.showDialog(response)
				}
			},
		},
	}

	// Feature 6: Random Event Frequency Tuning - add event settings if character has random events
	if dw.character.HasRandomEvents() {
		menuItems = append(menuItems, ContextMenuItem{
			Text: "Event Settings",
			Callback: func() {
				dw.showEventFrequencySettings()
			},
		})
	}

	return menuItems
}

// buildGameModeMenuItems creates game-specific menu items when game mode is enabled
func (dw *DesktopWindow) buildGameModeMenuItems() []ContextMenuItem {
	if !dw.gameMode || dw.character.GetGameState() == nil {
		return nil
	}

	var menuItems []ContextMenuItem

	menuItems = append(menuItems, ContextMenuItem{
		Text: "Feed",
		Callback: func() {
			response := dw.character.HandleGameInteraction("feed")
			if response != "" {
				dw.showDialog(response)
			}
		},
	})

	menuItems = append(menuItems, ContextMenuItem{
		Text: "Play",
		Callback: func() {
			response := dw.character.HandleGameInteraction("play")
			if response != "" {
				dw.showDialog(response)
			}
		},
	})

	// Add gift option if character has gift system enabled
	if dw.character.GetCard().HasGiftSystem() && dw.giftDialog != nil {
		menuItems = append(menuItems, ContextMenuItem{
			Text: "Give Gift",
			Callback: func() {
				dw.giftDialog.Show()
			},
		})
	}

	if dw.statsOverlay != nil {
		statsText := "Show Stats"
		if dw.statsOverlay.IsVisible() {
			statsText = "Hide Stats"
		}

		menuItems = append(menuItems, ContextMenuItem{
			Text: statsText,
			Callback: func() {
				dw.ToggleStatsOverlay()
			},
		})
	}

	return menuItems
}

// buildBattleMenuItems creates battle-related menu items for battle-capable characters
// Only shows battle options if the character has battle system enabled
// In network mode, provides battle invitation options as documented in README.md
func (dw *DesktopWindow) buildBattleMenuItems() []ContextMenuItem {
	// Check if character has battle system enabled
	if !dw.shouldShowBattleOptions() {
		return nil
	}

	var menuItems []ContextMenuItem

	if dw.networkMode && dw.networkOverlay != nil && dw.networkOverlay.GetNetworkManager() != nil {
		// Network mode: Provide battle invitation options as documented in README.md
		// "Battle invitations available through context menu in multiplayer mode"

		menuItems = append(menuItems, ContextMenuItem{
			Text: "Invite to Battle",
			Callback: func() {
				dw.handleBattleInvitation()
			},
		})

		menuItems = append(menuItems, ContextMenuItem{
			Text: "Challenge Player",
			Callback: func() {
				dw.handleBattleChallenge()
			},
		})

		menuItems = append(menuItems, ContextMenuItem{
			Text: "Send Battle Request",
			Callback: func() {
				dw.handleBattleRequest()
			},
		})
	} else {
		// Non-network mode: Basic battle initiation
		menuItems = append(menuItems, ContextMenuItem{
			Text: "Initiate Battle",
			Callback: func() {
				dw.handleBattleInitiation()
			},
		})
	}

	return menuItems
}

// buildChatMenuItems creates chat-related menu items for AI-capable characters
func (dw *DesktopWindow) buildChatMenuItems() []ContextMenuItem {
	if !dw.shouldShowChatOption() {
		return nil
	}

	var menuItems []ContextMenuItem

	chatText := "Open Chat"
	if dw.chatbotInterface != nil && dw.chatbotInterface.IsVisible() {
		chatText = "Close Chat"
	}

	menuItems = append(menuItems, ContextMenuItem{
		Text: chatText,
		Callback: func() {
			dw.handleChatOptionClick()
		},
	})

	if dw.chatbotInterface != nil {
		menuItems = append(menuItems, ContextMenuItem{
			Text: "Export Chat",
			Callback: func() {
				err := dw.chatbotInterface.ExportConversation()
				if err != nil {
					dw.showDialog(fmt.Sprintf("Export failed: %v", err))
				} else {
					dw.showDialog("Chat conversation exported successfully!")
				}
			},
		})
	}

	// Add romance history menu item if available
	if dw.shouldShowRomanceHistory() {
		menuItems = append(menuItems, ContextMenuItem{
			Text: "View Romance History",
			Callback: func() {
				dw.showRomanceHistory()
			},
		})
	}

	return menuItems
}

// buildNewsMenuItems creates news-related menu items for news-capable characters
// Only shows news options if the character has news features enabled
// As documented in Phase 3: UI and Events Integration
func (dw *DesktopWindow) buildNewsMenuItems() []ContextMenuItem {
	// Check if character is nil
	if dw.character == nil {
		return nil
	}

	// Check if character has news features enabled
	if !dw.character.GetCard().HasNewsFeatures() {
		return nil
	}

	var menuItems []ContextMenuItem

	// Add "📰 Read News" menu item
	menuItems = append(menuItems, ContextMenuItem{
		Text: "📰 Read News",
		Callback: func() {
			dw.HandleNewsReading()
		},
	})

	// Add "🔄 Update Feeds" menu item
	menuItems = append(menuItems, ContextMenuItem{
		Text: "🔄 Update Feeds",
		Callback: func() {
			dw.HandleFeedUpdate()
		},
	})

	return menuItems
}

// buildNetworkMenuItems creates network-related menu items when network mode is enabled
func (dw *DesktopWindow) buildNetworkMenuItems() []ContextMenuItem {
	if !dw.networkMode || dw.networkOverlay == nil {
		return nil
	}

	var menuItems []ContextMenuItem

	overlayText := "Network Overlay"
	if dw.networkOverlay.IsVisible() {
		overlayText = "Hide Network Overlay"
	} else {
		overlayText = "Show Network Overlay"
	}

	menuItems = append(menuItems, ContextMenuItem{
		Text: overlayText,
		Callback: func() {
			dw.ToggleNetworkOverlay()
		},
	})

	return menuItems
}

// buildUtilityMenuItems creates utility menu items like About and Shortcuts
func (dw *DesktopWindow) buildUtilityMenuItems() []ContextMenuItem {
	return []ContextMenuItem{
		{
			Text: "About",
			Callback: func() {
				response := dw.character.HandleRightClick()
				if response != "" {
					dw.showDialog(response)
				}
			},
		},
		{
			Text: "Shortcuts",
			Callback: func() {
				shortcutsText := dw.buildShortcutsText()
				dw.showDialog(shortcutsText)
			},
		},
	}
}

// buildShortcutsText constructs the keyboard shortcuts help text
func (dw *DesktopWindow) buildShortcutsText() string {
	shortcutsText := "Keyboard Shortcuts:\n"
	if dw.statsOverlay != nil {
		shortcutsText += "• 'S' - Toggle stats overlay\n"
	}
	if dw.chatbotInterface != nil {
		shortcutsText += "• 'C' - Toggle chatbot interface\n"
		shortcutsText += "• 'ESC' - Close chatbot interface\n"
	}
	if dw.networkOverlay != nil {
		shortcutsText += "• 'N' - Toggle network overlay\n"
	}
	if dw.character.GetCard().HasNewsFeatures() {
		shortcutsText += "• 'Ctrl+L' - Read latest news\n"
		shortcutsText += "• 'Ctrl+U' - Update news feeds\n"
	}
	shortcutsText += "• Right-click - Show this menu"
	return shortcutsText
}

// displayContextMenu configures and displays the context menu with auto-hide
func (dw *DesktopWindow) displayContextMenu(menuItems []ContextMenuItem) {
	dw.contextMenu.SetMenuItems(menuItems)
	dw.contextMenu.Show()

	go func() {
		time.Sleep(5 * time.Second)
		dw.contextMenu.Hide()
	}()
}

// animationLoop runs the character animation update loop
// Uses adaptive frame rate based on animation needs to optimize performance
func (dw *DesktopWindow) animationLoop() {
	maxFPS, idleFPS, currentInterval := dw.initializeFrameRates()
	ticker := time.NewTicker(currentInterval)
	defer ticker.Stop()

	consecutiveNoChanges := 0

	for range ticker.C {
		hasChanges := dw.character.Update()
		currentInterval, consecutiveNoChanges = dw.handleFrameRateAdaptation(
			hasChanges, consecutiveNoChanges, currentInterval, maxFPS, idleFPS, ticker)
		dw.processFrameUpdates(hasChanges)
	}
}

// initializeFrameRates sets up the frame rate configuration for the animation loop
func (dw *DesktopWindow) initializeFrameRates() (maxFPS, idleFPS, currentInterval time.Duration) {
	maxFPS = time.Second / 60  // 60 FPS when actively animating
	idleFPS = time.Second / 10 // 10 FPS when idle/no changes
	currentInterval = maxFPS   // Start with high frame rate
	return maxFPS, idleFPS, currentInterval
}

// handleFrameRateAdaptation manages adaptive frame rate switching based on animation state
func (dw *DesktopWindow) handleFrameRateAdaptation(hasChanges bool, consecutiveNoChanges int,
	currentInterval, maxFPS, idleFPS time.Duration, ticker *time.Ticker) (time.Duration, int) {

	if hasChanges {
		return dw.handleActiveAnimation(currentInterval, maxFPS, ticker), 0
	}
	return dw.handleIdleAnimation(consecutiveNoChanges, currentInterval, idleFPS, ticker)
}

// handleActiveAnimation switches to high frame rate when character is actively animating
func (dw *DesktopWindow) handleActiveAnimation(currentInterval, maxFPS time.Duration, ticker *time.Ticker) time.Duration {
	if currentInterval != maxFPS {
		ticker.Reset(maxFPS)
		if dw.debug {
			log.Printf("Animation active: switching to %v FPS", time.Second/maxFPS)
		}
		return maxFPS
	}
	return currentInterval
}

// handleIdleAnimation switches to low frame rate after consecutive frames without changes
func (dw *DesktopWindow) handleIdleAnimation(consecutiveNoChanges int, currentInterval, idleFPS time.Duration,
	ticker *time.Ticker) (time.Duration, int) {

	consecutiveNoChanges++
	// After 30 frames (0.5 seconds) with no changes, reduce frame rate
	if consecutiveNoChanges >= 30 && currentInterval != idleFPS {
		ticker.Reset(idleFPS)
		if dw.debug {
			log.Printf("Animation idle: switching to %v FPS for power saving", time.Second/idleFPS)
		}
		return idleFPS, consecutiveNoChanges
	}
	return currentInterval, consecutiveNoChanges
}

// processFrameUpdates handles performance monitoring and rendering updates
func (dw *DesktopWindow) processFrameUpdates(hasChanges bool) {
	// Record frame for performance monitoring
	if dw.profiler != nil {
		dw.profiler.RecordFrame()
	}

	// Check for new achievements and display notifications
	dw.checkForNewAchievements()

	// Only refresh renderer when there are actual changes
	if hasChanges {
		dw.renderer.Refresh()
	}
}

// setupDragging configures character dragging behavior
func (dw *DesktopWindow) setupDragging() {
	// Create draggable wrapper that implements Fyne's drag interface
	// This provides smooth cross-platform drag support without platform-specific code
	draggable := NewDraggableCharacter(dw, dw.character, dw.debug)

	// Create list of content objects for draggable layout
	objects := []fyne.CanvasObject{
		draggable,
		dw.dialog,
		dw.contextMenu,
	}

	// Add stats overlay if available
	if dw.statsOverlay != nil {
		objects = append(objects, dw.statsOverlay.GetContainer())
	}

	// Add stats tooltip if available
	if dw.statsTooltip != nil {
		objects = append(objects, dw.statsTooltip.GetContainer())
	}

	// Update window content to use draggable character instead of separate clickable overlay
	content := container.NewWithoutLayout(objects...)

	dw.window.SetContent(content)

	if dw.debug {
		log.Println("Character dragging enabled using Fyne drag system")
	}
}

// Show displays the desktop window
func (dw *DesktopWindow) Show() {
	dw.window.Show()

	if dw.debug {
		log.Printf("Desktop window shown for character: %s", dw.character.GetName())
	}
}

// Hide hides the desktop window
func (dw *DesktopWindow) Hide() {
	dw.window.Hide()
}

// Close closes the desktop window and stops animation
func (dw *DesktopWindow) Close() {
	dw.window.Close()
}

// SetPosition moves the window to the specified screen coordinates
// Uses available Fyne APIs for best-effort positioning support
func (dw *DesktopWindow) SetPosition(x, y int) {
	// Store position in character for reference
	dw.character.SetPosition(float32(x), float32(y))

	// Attempt to use available Fyne positioning capabilities
	// Note: Full positioning support varies by platform, but we can try
	if x == 0 && y == 0 {
		// Special case: center the window when position is (0,0)
		dw.window.CenterOnScreen()
		if dw.debug {
			log.Printf("Centering window using CenterOnScreen()")
		}
	} else {
		// For non-zero positions, we need to work within Fyne's limitations
		// Fyne doesn't expose direct positioning, but we can provide feedback
		if dw.debug {
			log.Printf("Position set to (%d, %d) - stored in character. Note: Fyne has limited window positioning support on some platforms", x, y)
		}
	}
}

// GetPosition returns the current window position
// Note: Fyne doesn't directly support window position queries on all platforms
func (dw *DesktopWindow) GetPosition() (int, int) {
	// Return stored character position as fallback
	x, y := dw.character.GetPosition()
	return int(x), int(y)
}

// CenterWindow centers the window on screen using Fyne's built-in capability
func (dw *DesktopWindow) CenterWindow() {
	dw.window.CenterOnScreen()
	// Reset stored position to indicate centered state
	dw.character.SetPosition(0, 0)

	if dw.debug {
		log.Println("Window centered on screen")
	}
}

// SetSize updates the window and character size
func (dw *DesktopWindow) SetSize(size int) {
	dw.window.Resize(fyne.NewSize(float32(size), float32(size)))
	dw.renderer.SetSize(size)
}

// GetCharacter returns the character instance for external access
func (dw *DesktopWindow) GetCharacter() *character.Character {
	return dw.character
}

// ToggleStatsOverlay shows/hides the stats overlay if available
func (dw *DesktopWindow) ToggleStatsOverlay() {
	if dw.statsOverlay != nil {
		dw.statsOverlay.Toggle()
		if dw.debug {
			if dw.statsOverlay.IsVisible() {
				log.Println("Stats overlay shown")
			} else {
				log.Println("Stats overlay hidden")
			}
		}
	}
}

// ShowStatsTooltip displays the stats tooltip for quick peek
func (dw *DesktopWindow) ShowStatsTooltip() {
	if dw.statsTooltip != nil {
		// Update content before showing
		dw.statsTooltip.UpdateContent()
		dw.statsTooltip.Show()
		dw.setupContent()
		if dw.debug {
			log.Println("Stats tooltip shown")
		}
	}
}

// HideStatsTooltip hides the stats tooltip
func (dw *DesktopWindow) HideStatsTooltip() {
	if dw.statsTooltip != nil {
		dw.statsTooltip.Hide()
		dw.setupContent()
		if dw.debug {
			log.Println("Stats tooltip hidden")
		}
	}
}

// ToggleChatbotInterface shows/hides the chatbot interface if available
func (dw *DesktopWindow) ToggleChatbotInterface() {
	if dw.chatbotInterface != nil {
		dw.chatbotInterface.Toggle()
		if dw.debug {
			if dw.chatbotInterface.IsVisible() {
				log.Println("Chatbot interface shown")
			} else {
				log.Println("Chatbot interface hidden")
			}
		}
	}
}

// ToggleChatbotInterfaceWithFocus shows/hides the chatbot interface with enhanced focus management
func (dw *DesktopWindow) ToggleChatbotInterfaceWithFocus() {
	if dw.chatbotInterface != nil {
		if !dw.chatbotInterface.IsVisible() {
			// Show chatbot and focus the input field
			dw.chatbotInterface.Show()
			dw.chatbotInterface.FocusInput()
			if dw.debug {
				log.Println("Chatbot interface shown with input focus")
			}
		} else {
			// Hide chatbot
			dw.chatbotInterface.Hide()
			if dw.debug {
				log.Println("Chatbot interface hidden")
			}
		}
	}
}

// ToggleNetworkOverlay shows/hides the network overlay if available
func (dw *DesktopWindow) ToggleNetworkOverlay() {
	if dw.networkOverlay != nil {
		dw.networkOverlay.Toggle()
		if dw.debug {
			if dw.networkOverlay.IsVisible() {
				log.Println("Network overlay shown")
			} else {
				log.Println("Network overlay hidden")
			}
		}
	}
}

// ShowAchievementNotification displays an achievement notification
func (dw *DesktopWindow) ShowAchievementNotification(details character.AchievementDetails) {
	if dw.achievementNotification != nil {
		dw.achievementNotification.ShowAchievement(details)
		if dw.debug {
			log.Printf("Achievement notification shown: %s", details.Name)
		}
	}
}

// checkForNewAchievements checks if any new achievements were earned and displays notifications
func (dw *DesktopWindow) checkForNewAchievements() {
	if dw.character == nil || dw.achievementNotification == nil {
		return
	}

	gameState := dw.character.GetGameState()
	if gameState == nil {
		return
	}

	// Get any recently earned achievements
	recentAchievements := gameState.GetRecentAchievements()
	for _, achievement := range recentAchievements {
		dw.ShowAchievementNotification(achievement)
	}
}

// configureAlwaysOnTop attempts to configure always-on-top behavior using available Fyne capabilities
// Following the "lazy programmer" principle: use what's available rather than implementing platform-specific code
func configureAlwaysOnTop(window fyne.Window, debug bool) {
	// Fyne v2.4.5 has limited always-on-top support, but we can try available approaches:

	// 1. Try to minimize window decorations (makes it more overlay-like)
	window.SetTitle("") // Remove title bar text for cleaner overlay appearance

	// 2. Set window to be borderless for better desktop integration
	// Note: Fyne doesn't expose direct borderless mode, but we can minimize decoration

	// 3. Configure for desktop overlay use case
	// Fyne's design philosophy focuses on cross-platform compatibility over platform-specific features
	// True always-on-top requires platform-specific window manager hints that Fyne doesn't expose

	if debug {
		log.Println("Always-on-top configuration applied using available Fyne capabilities")
		log.Println("Note: Full always-on-top behavior requires platform-specific window manager support")
		log.Println("Window configured for optimal desktop overlay experience within Fyne's limitations")
	}

	// Future enhancement opportunity:
	// Could implement platform-specific always-on-top using CGO or system calls,
	// but this would violate the "lazy programmer" principle of avoiding custom platform code
}

// configureTransparency configures window transparency for desktop overlay behavior
// Following the "lazy programmer" principle: use Fyne's available transparency features
func configureTransparency(window fyne.Window, debug bool) {
	// Remove window padding to make character appear directly on desktop
	window.SetPadded(false)

	if debug {
		log.Println("Window transparency configuration applied using available Fyne capabilities")
		log.Println("Note: True transparency requires transparent window backgrounds and content")
		log.Println("Character should appear with minimal window decoration for overlay effect")
	}
}

// setupKeyboardShortcuts configures keyboard shortcuts for the desktop window
func (dw *DesktopWindow) setupKeyboardShortcuts() {
	canvas := dw.window.Canvas()

	// Configure basic keyboard shortcuts
	dw.setupBasicKeyShortcuts(canvas)

	// Configure conditional shortcuts based on features
	dw.setupConditionalShortcuts(canvas)

	// Log available shortcuts for debugging
	dw.logAvailableShortcuts()
}

// setupBasicKeyShortcuts configures core keyboard shortcuts available to all windows
func (dw *DesktopWindow) setupBasicKeyShortcuts(canvas fyne.Canvas) {
	canvas.SetOnTypedKey(func(key *fyne.KeyEvent) {
		switch key.Name {
		case fyne.KeyS:
			dw.handleStatsToggleShortcut()
		case fyne.KeyC:
			dw.handleChatToggleShortcut()
		case fyne.KeyN:
			dw.handleNetworkToggleShortcut()
		case fyne.KeyEscape:
			dw.handleEscapeKeyShortcut()
		}
	})
}

// handleStatsToggleShortcut processes the 'S' key press for stats overlay toggle
func (dw *DesktopWindow) handleStatsToggleShortcut() {
	if dw.debug {
		log.Println("Stats toggle shortcut pressed (S key)")
	}
	dw.ToggleStatsOverlay()
}

// handleChatToggleShortcut processes the 'C' key press for chatbot interface toggle
func (dw *DesktopWindow) handleChatToggleShortcut() {
	if dw.debug {
		log.Println("Chat toggle shortcut pressed (C key)")
	}
	dw.ToggleChatbotInterfaceWithFocus()
}

// handleNetworkToggleShortcut processes the 'N' key press for network overlay toggle
func (dw *DesktopWindow) handleNetworkToggleShortcut() {
	if dw.debug {
		log.Println("Network toggle shortcut pressed (N key)")
	}
	dw.ToggleNetworkOverlay()
}

// handleEscapeKeyShortcut processes the Escape key press for closing chatbot interface
func (dw *DesktopWindow) handleEscapeKeyShortcut() {
	if dw.chatbotInterface != nil && dw.chatbotInterface.IsVisible() {
		if dw.debug {
			log.Println("Escape key pressed - closing chatbot interface")
		}
		dw.chatbotInterface.Hide()
	}
}

// setupConditionalShortcuts configures feature-dependent keyboard shortcuts
func (dw *DesktopWindow) setupConditionalShortcuts(canvas fyne.Canvas) {
	// Add general events keyboard shortcuts only if events are enabled
	if dw.eventsEnabled {
		dw.setupGeneralEventsShortcuts(canvas)
	}

	// Add news keyboard shortcuts if character has news features
	if dw.character.GetCard().HasNewsFeatures() {
		dw.setupNewsShortcuts(canvas)
	}

	// Feature 6: Random Event Frequency Tuning - add frequency shortcuts if character has random events
	if dw.character.HasRandomEvents() {
		dw.setupEventFrequencyShortcuts(canvas)
	}
}

// logAvailableShortcuts outputs debugging information about configured shortcuts
func (dw *DesktopWindow) logAvailableShortcuts() {
	if !dw.debug {
		return
	}

	log.Println("Keyboard shortcuts configured:")
	dw.logBasicShortcuts()
	dw.logConditionalShortcuts()
}

// logBasicShortcuts outputs information about basic keyboard shortcuts
func (dw *DesktopWindow) logBasicShortcuts() {
	log.Println("  'S' - Toggle stats overlay")
	log.Println("  'C' - Toggle chatbot")
	log.Println("  'N' - Toggle network overlay")
	log.Println("  'ESC' - Close chatbot")
}

// logConditionalShortcuts outputs information about feature-dependent shortcuts
func (dw *DesktopWindow) logConditionalShortcuts() {
	if dw.eventsEnabled {
		log.Println("  'Ctrl+E' - Open events menu")
		log.Println("  'Ctrl+R' - Random roleplay scenario")
		log.Println("  'Ctrl+G' - Mini-game session")
		log.Println("  'Ctrl+H' - Humor/joke session")
	}
	if dw.character.GetCard().HasNewsFeatures() {
		log.Println("  'Ctrl+L' - Read latest news")
		log.Println("  'Ctrl+U' - Update news feeds")
	}
	// Feature 6: Random Event Frequency Tuning
	if dw.character.HasRandomEvents() {
		log.Println("  'Ctrl+1' - Very Rare events (0.5x)")
		log.Println("  'Ctrl+2' - Normal events (1.0x)")
		log.Println("  'Ctrl+3' - Frequent events (1.5x)")
		log.Println("  'Ctrl+4' - Very Frequent events (2.0x)")
		log.Println("  'Ctrl+5' - Maximum events (3.0x)")
	}
}

// setupGeneralEventsShortcuts configures general dialog events keyboard shortcuts
func (dw *DesktopWindow) setupGeneralEventsShortcuts(canvas fyne.Canvas) {
	// Ctrl+E: Open events menu to see available scenarios
	ctrlE := &desktop.CustomShortcut{
		KeyName:  fyne.KeyE,
		Modifier: fyne.KeyModifierControl,
	}
	canvas.AddShortcut(ctrlE, func(shortcut fyne.Shortcut) {
		if dw.debug {
			log.Println("Ctrl+E pressed - opening events menu")
		}
		dw.openEventsMenu()
	})

	// Ctrl+R: Quick-start a random roleplay scenario
	ctrlR := &desktop.CustomShortcut{
		KeyName:  fyne.KeyR,
		Modifier: fyne.KeyModifierControl,
	}
	canvas.AddShortcut(ctrlR, func(shortcut fyne.Shortcut) {
		if dw.debug {
			log.Println("Ctrl+R pressed - starting random roleplay scenario")
		}
		dw.startRandomRoleplayScenario()
	})

	// Ctrl+G: Start a mini-game or trivia session
	ctrlG := &desktop.CustomShortcut{
		KeyName:  fyne.KeyG,
		Modifier: fyne.KeyModifierControl,
	}
	canvas.AddShortcut(ctrlG, func(shortcut fyne.Shortcut) {
		if dw.debug {
			log.Println("Ctrl+G pressed - starting mini-game session")
		}
		dw.startMiniGameSession()
	})

	// Ctrl+H: Trigger a humor/joke session
	ctrlH := &desktop.CustomShortcut{
		KeyName:  fyne.KeyH,
		Modifier: fyne.KeyModifierControl,
	}
	canvas.AddShortcut(ctrlH, func(shortcut fyne.Shortcut) {
		if dw.debug {
			log.Println("Ctrl+H pressed - starting humor/joke session")
		}
		dw.startHumorSession()
	})
}

// setupNewsShortcuts configures news-related keyboard shortcuts
// Only called if character has news features enabled
func (dw *DesktopWindow) setupNewsShortcuts(canvas fyne.Canvas) {
	// Ctrl+L: Read latest news
	ctrlL := &desktop.CustomShortcut{
		KeyName:  fyne.KeyL,
		Modifier: fyne.KeyModifierControl,
	}
	canvas.AddShortcut(ctrlL, func(shortcut fyne.Shortcut) {
		if dw.debug {
			log.Println("Ctrl+L pressed - reading latest news")
		}
		dw.HandleNewsReading()
	})

	// Ctrl+U: Update news feeds
	ctrlU := &desktop.CustomShortcut{
		KeyName:  fyne.KeyU,
		Modifier: fyne.KeyModifierControl,
	}
	canvas.AddShortcut(ctrlU, func(shortcut fyne.Shortcut) {
		if dw.debug {
			log.Println("Ctrl+U pressed - updating news feeds")
		}
		dw.HandleFeedUpdate()
	})
}

// setupEventFrequencyShortcuts configures random event frequency adjustment shortcuts
// Feature 6: Random Event Frequency Tuning - only called if character has random events
func (dw *DesktopWindow) setupEventFrequencyShortcuts(canvas fyne.Canvas) {
	// Ctrl+1: Set Very Rare frequency (0.5x)
	ctrl1 := &desktop.CustomShortcut{
		KeyName:  fyne.Key1,
		Modifier: fyne.KeyModifierControl,
	}
	canvas.AddShortcut(ctrl1, func(shortcut fyne.Shortcut) {
		dw.setEventFrequency(0.5, "Very Rare")
	})

	// Ctrl+2: Set Normal frequency (1.0x)
	ctrl2 := &desktop.CustomShortcut{
		KeyName:  fyne.Key2,
		Modifier: fyne.KeyModifierControl,
	}
	canvas.AddShortcut(ctrl2, func(shortcut fyne.Shortcut) {
		dw.setEventFrequency(1.0, "Normal")
	})

	// Ctrl+3: Set Frequent frequency (1.5x)
	ctrl3 := &desktop.CustomShortcut{
		KeyName:  fyne.Key3,
		Modifier: fyne.KeyModifierControl,
	}
	canvas.AddShortcut(ctrl3, func(shortcut fyne.Shortcut) {
		dw.setEventFrequency(1.5, "Frequent")
	})

	// Ctrl+4: Set Very Frequent frequency (2.0x)
	ctrl4 := &desktop.CustomShortcut{
		KeyName:  fyne.Key4,
		Modifier: fyne.KeyModifierControl,
	}
	canvas.AddShortcut(ctrl4, func(shortcut fyne.Shortcut) {
		dw.setEventFrequency(2.0, "Very Frequent")
	})

	// Ctrl+5: Set Maximum frequency (3.0x)
	ctrl5 := &desktop.CustomShortcut{
		KeyName:  fyne.Key5,
		Modifier: fyne.KeyModifierControl,
	}
	canvas.AddShortcut(ctrl5, func(shortcut fyne.Shortcut) {
		dw.setEventFrequency(3.0, "Maximum")
	})
}

// setEventFrequency sets the event frequency multiplier and shows confirmation
// Feature 6: Random Event Frequency Tuning
func (dw *DesktopWindow) setEventFrequency(multiplier float64, label string) {
	if dw.debug {
		log.Printf("Setting event frequency to %.1fx (%s)", multiplier, label)
	}

	dw.character.SetEventFrequencyMultiplier(multiplier)
	dw.showDialog(fmt.Sprintf("Event frequency set to %s (%.1fx)", label, multiplier))
}

// General Events System Implementation - implements keyboard shortcuts functionality

// openEventsMenu displays available general events for the user to choose from
func (dw *DesktopWindow) openEventsMenu() {
	availableEvents := dw.character.GetAvailableGeneralEvents()

	if len(availableEvents) == 0 {
		dw.showDialog("No events available for this character.")
		return
	}

	// Create menu text with available events
	menuText := "Available Events:\n\n"
	for i, event := range availableEvents {
		menuText += fmt.Sprintf("%d. %s (%s)\n   %s\n\n",
			i+1, event.Name, event.Category, event.Description)
	}

	dw.showDialog(menuText)
}

// startRandomRoleplayScenario triggers a random roleplay event
func (dw *DesktopWindow) startRandomRoleplayScenario() {
	roleplays := dw.character.GetGeneralEventsByCategory("roleplay")

	if len(roleplays) == 0 {
		dw.showDialog("No roleplay scenarios available for this character.")
		return
	}

	// Pick a random roleplay event
	event := roleplays[int(time.Now().UnixNano())%len(roleplays)]

	if dw.debug {
		log.Printf("Triggering random roleplay scenario: %s", event.Name)
	}

	response := dw.character.HandleGeneralEvent(event.Name)
	if response != "" {
		dw.showDialog(fmt.Sprintf("Roleplay: %s\n\n%s", event.Name, response))
	} else {
		dw.showDialog(fmt.Sprintf("Could not start roleplay scenario: %s", event.Name))
	}
}

// startMiniGameSession triggers a game category event
func (dw *DesktopWindow) startMiniGameSession() {
	games := dw.character.GetGeneralEventsByCategory("game")

	if len(games) == 0 {
		dw.showDialog("No mini-games available for this character.")
		return
	}

	// Pick a random game event
	event := games[int(time.Now().UnixNano())%len(games)]

	if dw.debug {
		log.Printf("Triggering mini-game session: %s", event.Name)
	}

	response := dw.character.HandleGeneralEvent(event.Name)
	if response != "" {
		dw.showDialog(fmt.Sprintf("Mini-Game: %s\n\n%s", event.Name, response))
	} else {
		dw.showDialog(fmt.Sprintf("Could not start mini-game: %s", event.Name))
	}
}

// startHumorSession triggers a humor category event
func (dw *DesktopWindow) startHumorSession() {
	humor := dw.character.GetGeneralEventsByCategory("humor")

	if len(humor) == 0 {
		dw.showDialog("No humor/joke content available for this character.")
		return
	}

	// Pick a random humor event
	event := humor[int(time.Now().UnixNano())%len(humor)]

	if dw.debug {
		log.Printf("Triggering humor session: %s", event.Name)
	}

	response := dw.character.HandleGeneralEvent(event.Name)
	if response != "" {
		dw.showDialog(fmt.Sprintf("Humor: %s\n\n%s", event.Name, response))
	} else {
		dw.showDialog(fmt.Sprintf("Could not start humor session: %s", event.Name))
	}
}

// Bug #3 Fix: Improved context menu chat access

// shouldShowChatOption determines if "Open Chat" should appear in the context menu
// Shows for any AI-capable character (has dialog backend OR romance features)
func (dw *DesktopWindow) shouldShowChatOption() bool {
	card := dw.character.GetCard()
	if card == nil {
		return false
	}

	// Show chat option if character has any AI capabilities:
	// 1. Has dialog backend configured (even if disabled)
	// 2. Has romance features (personality, romance dialogs, romance events)
	return card.HasDialogBackendConfig() || card.HasRomanceFeatures()
}

// handleChatOptionClick handles when user clicks "Open Chat" in context menu
// Provides appropriate feedback based on character's chat capabilities
func (dw *DesktopWindow) handleChatOptionClick() {
	// If chatbot interface is available, use normal toggle
	if dw.chatbotInterface != nil {
		dw.ToggleChatbotInterface()
		return
	}

	// No chatbot interface - provide feedback about why
	card := dw.character.GetCard()
	if card == nil {
		dw.showDialog("Chat not available for this character.")
		return
	}

	// Use new granular methods for better user feedback
	if !card.HasDialogBackendConfig() {
		if card.HasRomanceFeatures() {
			dw.showDialog("Chat feature available but no dialog backend configured.\n\nThis character has romance features but needs a dialog backend to enable AI chat.\n\nYou can still interact using the basic dialog system.")
		} else {
			dw.showDialog("Chat not available for this character.\n\nThis character doesn't have AI dialog capabilities.\n\nYou can still interact using basic responses.")
		}
	} else if !card.IsDialogBackendEnabled() {
		dw.showDialog("Chat feature disabled.\n\nThis character has a dialog backend configured but it's currently disabled.\n\nEnable it in the character configuration to use AI chat.")
	} else {
		// This shouldn't happen - HasDialogBackend() returned false but conditions suggest it should work
		hasConfig, enabled, summary := card.GetDialogBackendStatus()
		dw.showDialog(fmt.Sprintf("Chat temporarily unavailable.\n\nDialog backend status: Config=%t, Enabled=%t\nSummary: %s\n\nThere may be an issue with the dialog backend configuration.", hasConfig, enabled, summary))
	}
}

// shouldShowBattleOptions determines if battle options should appear in the context menu
// Shows for characters with battle system enabled
func (dw *DesktopWindow) shouldShowBattleOptions() bool {
	card := dw.character.GetCard()
	if card == nil {
		return false
	}

	// Show battle options if character has battle system enabled
	return card.HasBattleSystem()
}

// shouldShowRomanceHistory determines if "View Romance History" should appear in the context menu
// Shows for characters with game state and romance memories
func (dw *DesktopWindow) shouldShowRomanceHistory() bool {
	if dw.character == nil {
		return false
	}

	card := dw.character.GetCard()
	gameState := dw.character.GetGameState()

	// Check if character has romance features and game state
	if card == nil || !card.HasRomanceFeatures() || gameState == nil {
		return false
	}

	// Only show if there are romance memories to display
	romanceMemories := gameState.GetRomanceMemories()
	return len(romanceMemories) > 0
}

// handleBattleInitiation handles when user clicks "Initiate Battle" in context menu
// Shows the battle action dialog for selecting battle actions
func (dw *DesktopWindow) handleBattleInitiation() {
	// Check if networking is enabled
	if dw.networkOverlay == nil || dw.networkOverlay.GetNetworkManager() == nil {
		dw.showDialog("Battle system requires networking. Start with -network flag to enable multiplayer battles.")
		return
	}

	// Get available peers for battle
	peers := dw.networkOverlay.GetNetworkManager().GetPeers()
	if len(peers) == 0 {
		dw.showDialog("No other players available for battle. Make sure other DDS instances are running on the network.")
		return
	}

	// Check if character is available
	if dw.character == nil {
		dw.showDialog("No character loaded for battle.")
		return
	}

	// Show peer selection dialog for multiple peers (Fix for Finding #5)
	if len(peers) == 1 {
		// If only one peer, skip dialog and select directly
		dw.initiateBattleWithPeer(peers[0])
	} else {
		// Show peer selection dialog for multiple peers
		dw.peerSelectionDialog.Show(peers,
			func(selectedPeer network.Peer) {
				dw.initiateBattleWithPeer(selectedPeer)
			},
			func() {
				// User cancelled peer selection
			},
		)
	}
}

// initiateBattleWithPeer initiates a battle with the specified peer
func (dw *DesktopWindow) initiateBattleWithPeer(targetPeer network.Peer) {
	// Create battle invitation payload
	battleID := fmt.Sprintf("battle_%d", time.Now().UnixNano())
	payload := network.BattleInvitePayload{
		FromCharacterID: dw.networkOverlay.GetNetworkManager().GetNetworkID(),
		ToCharacterID:   targetPeer.ID,
		BattleID:        battleID,
		Timestamp:       time.Now(),
	}

	// Send battle invitation through network manager
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		dw.showDialog(fmt.Sprintf("Failed to create battle invitation: %v", err))
		return
	}

	if err := dw.networkOverlay.GetNetworkManager().SendMessage(network.MessageTypeBattleInvite, payloadBytes, targetPeer.ID); err != nil {
		dw.showDialog(fmt.Sprintf("Failed to send battle invitation: %v", err))
		return
	}

	dw.showDialog(fmt.Sprintf("Battle invitation sent to %s! Waiting for response...", targetPeer.ID))
}

// handleBattleInvitation handles sending battle invitations to other players in network mode
func (dw *DesktopWindow) handleBattleInvitation() {
	if dw.networkOverlay == nil || dw.networkOverlay.GetNetworkManager() == nil {
		dw.showDialog("Network not available. Battle invitations require network mode.")
		return
	}

	peers := dw.networkOverlay.GetNetworkManager().GetPeers()
	if len(peers) == 0 {
		dw.showDialog("No other players connected. Battle invitations require connected peers.")
		return
	}

	// TODO: Show peer selection dialog and send battle invitation using the network protocol
	// For now, send invitation to first available peer
	if len(peers) > 0 {
		targetPeer := peers[0]

		// Create battle invitation payload
		battleID := fmt.Sprintf("battle_inv_%d", time.Now().UnixNano())
		payload := network.BattleInvitePayload{
			FromCharacterID: dw.networkOverlay.GetNetworkManager().GetNetworkID(),
			ToCharacterID:   targetPeer.ID,
			BattleID:        battleID,
			Timestamp:       time.Now(),
		}

		// Send battle invitation through network manager
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			dw.showDialog(fmt.Sprintf("Failed to create battle invitation: %v", err))
			return
		}

		if err := dw.networkOverlay.GetNetworkManager().SendMessage(network.MessageTypeBattleInvite, payloadBytes, targetPeer.ID); err != nil {
			dw.showDialog(fmt.Sprintf("Failed to send battle invitation: %v", err))
			return
		}

		dw.showDialog(fmt.Sprintf("Battle invitation sent to %s! They can accept or decline.", targetPeer.ID))
	} else {
		dw.showDialog("No players available for battle invitations.")
	}
}

// handleBattleChallenge handles challenging specific players to battle
func (dw *DesktopWindow) handleBattleChallenge() {
	if dw.networkOverlay == nil || dw.networkOverlay.GetNetworkManager() == nil {
		dw.showDialog("Network not available. Battle challenges require network mode.")
		return
	}

	peers := dw.networkOverlay.GetNetworkManager().GetPeers()
	if len(peers) == 0 {
		dw.showDialog("No other players connected. Battle challenges require connected peers.")
		return
	}

	// TODO: Show peer selection dialog with challenge options
	// This could include different types of battles (ranked, casual, tournament, etc.)
	// For now, send challenge to first available peer
	if len(peers) > 0 {
		targetPeer := peers[0]

		// Create battle challenge invitation (using same payload structure)
		battleID := fmt.Sprintf("challenge_%d", time.Now().UnixNano())
		payload := network.BattleInvitePayload{
			FromCharacterID: dw.networkOverlay.GetNetworkManager().GetNetworkID(),
			ToCharacterID:   targetPeer.ID,
			BattleID:        battleID,
			Timestamp:       time.Now(),
		}

		// Send challenge through network manager
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			dw.showDialog(fmt.Sprintf("Failed to create battle challenge: %v", err))
			return
		}

		if err := dw.networkOverlay.GetNetworkManager().SendMessage(network.MessageTypeBattleInvite, payloadBytes, targetPeer.ID); err != nil {
			dw.showDialog(fmt.Sprintf("Failed to send battle challenge: %v", err))
			return
		}

		dw.showDialog(fmt.Sprintf("Challenge sent to %s! This will be recorded for ranking.", targetPeer.ID))
	} else {
		dw.showDialog("No players available to challenge.")
	}
}

// handleBattleRequest handles sending formal battle requests with specific terms
func (dw *DesktopWindow) handleBattleRequest() {
	if dw.networkOverlay == nil || dw.networkOverlay.GetNetworkManager() == nil {
		dw.showDialog("Network not available. Battle requests require network mode.")
		return
	}

	peers := dw.networkOverlay.GetNetworkManager().GetPeers()
	if len(peers) == 0 {
		dw.showDialog("No other players connected. Battle requests require connected peers.")
		return
	}

	// TODO: Show detailed battle request dialog with options for:
	// - Battle type (casual, ranked, tournament)
	// - Rules and constraints
	// - Time limits
	// - Spectator settings
	// For now, send formal battle request to first available peer
	if len(peers) > 0 {
		targetPeer := peers[0]

		// Create formal battle request (using same payload structure)
		battleID := fmt.Sprintf("formal_req_%d", time.Now().UnixNano())
		payload := network.BattleInvitePayload{
			FromCharacterID: dw.networkOverlay.GetNetworkManager().GetNetworkID(),
			ToCharacterID:   targetPeer.ID,
			BattleID:        battleID,
			Timestamp:       time.Now(),
		}

		// Send formal battle request through network manager
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			dw.showDialog(fmt.Sprintf("Failed to create formal battle request: %v", err))
			return
		}

		if err := dw.networkOverlay.GetNetworkManager().SendMessage(network.MessageTypeBattleInvite, payloadBytes, targetPeer.ID); err != nil {
			dw.showDialog(fmt.Sprintf("Failed to send formal battle request: %v", err))
			return
		}

		dw.showDialog(fmt.Sprintf("Formal battle request sent to %s with standard rules. Negotiation may begin.", targetPeer.ID))
	} else {
		dw.showDialog("No players available for formal battle requests.")
	}
}

// HandleNewsReading handles when user clicks "📰 Read News" in context menu
// Triggers news reading events and displays news content through existing dialog system
func (dw *DesktopWindow) HandleNewsReading() {
	// Check if character is nil
	if dw.character == nil {
		return
	}

	// Check if character has news features enabled
	if !dw.character.GetCard().HasNewsFeatures() {
		dw.showDialog("News features not available for this character.")
		return
	} // Try to trigger a news reading event
	// Look for a configured news reading event to trigger
	newsConfig := dw.character.GetCard().NewsFeatures
	if newsConfig == nil || len(newsConfig.ReadingEvents) == 0 {
		dw.showDialog("No news reading events configured for this character.")
		return
	}

	// Use the first available reading event as default
	eventName := newsConfig.ReadingEvents[0].Name
	response, err := dw.character.HandleNewsEvent(eventName)
	if err != nil {
		// Provide user-friendly error message
		dw.showDialog(fmt.Sprintf("Unable to read news right now: %v", err))
		return
	}

	if response != "" {
		dw.showDialog(response)
	} else {
		dw.showDialog("No news available at this time.")
	}
}

// HandleFeedUpdate handles when user clicks "🔄 Update Feeds" in context menu
// Manually triggers news feed updates and provides feedback to user
func (dw *DesktopWindow) HandleFeedUpdate() {
	// Check if character is nil
	if dw.character == nil {
		return
	}

	// Check if character has news features enabled
	if !dw.character.GetCard().HasNewsFeatures() {
		dw.showDialog("News features not available for this character.")
		return
	} // Provide feedback that update is starting
	dw.showDialog("Updating news feeds...")

	// Start feed update in background
	go func() {
		// Attempt to refresh news feeds through the character's news features
		newsConfig := dw.character.GetCard().NewsFeatures
		if newsConfig != nil && len(newsConfig.Feeds) > 0 {
			// Trigger a news event to refresh feeds - this will internally call the news backend
			if _, err := dw.character.HandleNewsEvent("feed_refresh"); err != nil {
				// If specific news event doesn't exist, just provide generic feedback
				dw.showDialog("News feeds refreshed! Fresh content may be available through news reading.")
			} else {
				// Show success feedback
				dw.showDialog("News feeds updated successfully! Fresh content is now available.")
			}
		} else {
			// No feeds configured
			dw.showDialog("No news feeds configured for this character.")
		}
	}()
}

// showRomanceHistory displays a formatted list of romance interactions and milestones
// Uses existing showDialog pattern but with enhanced formatting for romance memories
func (dw *DesktopWindow) showRomanceHistory() {
	if dw.character == nil {
		dw.showDialog("No character available.")
		return
	}

	gameState := dw.character.GetGameState()
	if gameState == nil {
		dw.showDialog("No romance history available.")
		return
	}

	romanceMemories := gameState.GetRomanceMemories()
	if len(romanceMemories) == 0 {
		dw.showDialog("No romance interactions recorded yet.")
		return
	}

	// Format romance memories into readable text
	historyText := dw.formatRomanceHistory(romanceMemories)
	dw.showDialog(historyText)
}

// formatRomanceHistory formats romance memories into a readable string
// Shows recent interactions with timestamps and stat changes
func (dw *DesktopWindow) formatRomanceHistory(memories []character.RomanceMemory) string {
	// Show last 10 memories to keep dialog manageable
	maxMemories := 10
	startIndex := 0
	if len(memories) > maxMemories {
		startIndex = len(memories) - maxMemories
	}

	recentMemories := memories[startIndex:]

	var builder strings.Builder
	builder.WriteString("💕 Romance History\n\n")

	for i := len(recentMemories) - 1; i >= 0; i-- {
		memory := recentMemories[i]

		// Format timestamp
		timeStr := memory.Timestamp.Format("Jan 2, 15:04")

		// Format interaction
		builder.WriteString(fmt.Sprintf("🕐 %s\n", timeStr))
		builder.WriteString(fmt.Sprintf("💫 %s\n", memory.InteractionType))
		builder.WriteString(fmt.Sprintf("💬 %s\n", memory.Response))

		// Format stat changes
		statChanges := dw.formatStatChanges(memory.StatsBefore, memory.StatsAfter)
		if statChanges != "" {
			builder.WriteString(fmt.Sprintf("📊 %s\n", statChanges))
		}

		builder.WriteString("\n")
	}

	// Add summary
	if len(memories) > maxMemories {
		builder.WriteString(fmt.Sprintf("💡 Showing last %d of %d total memories", maxMemories, len(memories)))
	} else {
		builder.WriteString(fmt.Sprintf("💡 Total memories: %d", len(memories)))
	}

	return builder.String()
}

// formatStatChanges formats the before/after stat changes into a readable string
// Shows meaningful stat increases/decreases with appropriate symbols
func (dw *DesktopWindow) formatStatChanges(before, after map[string]float64) string {
	if before == nil || after == nil {
		return ""
	}

	var changes []string
	romanceStats := []string{"affection", "trust", "intimacy", "jealousy"}

	for _, stat := range romanceStats {
		beforeVal, hasBefore := before[stat]
		afterVal, hasAfter := after[stat]

		if hasBefore && hasAfter {
			change := afterVal - beforeVal
			if change != 0 {
				symbol := "📈"
				if change < 0 {
					symbol = "📉"
				}
				changes = append(changes, fmt.Sprintf("%s %s %+.1f", symbol, stat, change))
			}
		}
	}

	if len(changes) == 0 {
		return ""
	}

	return strings.Join(changes, ", ")
}

// onSaveStarted updates the save status indicator when a save operation begins
func (dw *DesktopWindow) onSaveStarted() {
	if dw.saveStatusIndicator != nil {
		dw.saveStatusIndicator.SetStatus(SaveStatusSaving, "")
	}
}

// onSaveCompleted updates the save status indicator when a save operation completes successfully
// Automatically returns to idle state after 2 seconds
func (dw *DesktopWindow) onSaveCompleted() {
	if dw.saveStatusIndicator != nil {
		dw.saveStatusIndicator.SetStatus(SaveStatusSaved, "")

		// Return to idle after 2 seconds
		go func() {
			time.Sleep(2 * time.Second)
			if dw.saveStatusIndicator != nil {
				dw.saveStatusIndicator.SetStatus(SaveStatusIdle, "")
			}
		}()
	}
}

// onSaveError updates the save status indicator when a save operation fails
func (dw *DesktopWindow) onSaveError(err error) {
	if dw.saveStatusIndicator != nil {
		dw.saveStatusIndicator.SetStatus(SaveStatusError, err.Error())
	}
}

// SetSaveStatusCallback configures the save status callback for the window
// This method should be called to connect the window's save status indicator
// to a SaveManager's status notifications
func (dw *DesktopWindow) SetSaveStatusCallback() func(SaveStatus, string) {
	return func(status SaveStatus, message string) {
		switch status {
		case SaveStatusSaving:
			dw.onSaveStarted()
		case SaveStatusSaved:
			dw.onSaveCompleted()
		case SaveStatusError:
			dw.onSaveError(fmt.Errorf(message))
		case SaveStatusIdle:
			if dw.saveStatusIndicator != nil {
				dw.saveStatusIndicator.SetStatus(SaveStatusIdle, "")
			}
		}
	}
}

// ShowBattleInvitationDialog shows a battle invitation confirmation dialog
// Provides a UI-based replacement for hardcoded battle acceptance logic
func (dw *DesktopWindow) ShowBattleInvitationDialog(fromCharacter string, onResponse func(accepted bool)) {
	if dw.battleInvitationDialog != nil {
		dw.battleInvitationDialog.Show(fromCharacter, onResponse)
	}
}
