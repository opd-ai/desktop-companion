package character

import (
	"fmt"
	"image"
	"sync"
	"time"
)

// Character represents a desktop companion with behavior, animations, and interactions
// Follows the "lazy programmer" approach by combining existing components
type Character struct {
	mu               sync.RWMutex
	card             *CharacterCard
	animationManager *AnimationManager
	basePath         string

	// State management
	currentState    string
	lastStateChange time.Time
	lastInteraction time.Time
	dialogCooldowns map[string]time.Time

	// Behavior settings
	idleTimeout     time.Duration
	movementEnabled bool
	size            int

	// Position (for draggable characters)
	x, y float32

	// Game features (added for Phase 2)
	gameState                *GameState
	gameInteractionCooldowns map[string]time.Time
	randomEventManager       *RandomEventManager // Added for Phase 3 - random events
}

// New creates a new character instance from a character card
// Loads all animations and initializes behavior state
func New(card *CharacterCard, basePath string) (*Character, error) {
	char := &Character{
		card:                     card,
		animationManager:         NewAnimationManager(),
		basePath:                 basePath,
		currentState:             "idle",
		lastStateChange:          time.Now(),
		lastInteraction:          time.Now(),
		dialogCooldowns:          make(map[string]time.Time),
		gameInteractionCooldowns: make(map[string]time.Time),
		idleTimeout:              time.Duration(card.Behavior.IdleTimeout) * time.Second,
		movementEnabled:          card.Behavior.MovementEnabled,
		size:                     card.Behavior.DefaultSize,
	}

	// Initialize game features if the character card has them
	if card.HasGameFeatures() {
		char.initializeGameFeatures()
	}

	// Load all animations from the character card
	for name := range card.Animations {
		fullPath, err := card.GetAnimationPath(basePath, name)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve animation path for '%s': %w", name, err)
		}

		if err := char.animationManager.LoadAnimation(name, fullPath); err != nil {
			return nil, fmt.Errorf("failed to load animation '%s': %w", name, err)
		}
	}

	// Set initial animation to idle
	if err := char.animationManager.SetCurrentAnimation("idle"); err != nil {
		return nil, fmt.Errorf("failed to set initial animation: %w", err)
	}

	return char, nil
}

// initializeGameFeatures sets up game state from character card configuration
// This method is called during character creation if game features are enabled
func (c *Character) initializeGameFeatures() {
	if c.card.Stats == nil || len(c.card.Stats) == 0 {
		return
	}

	// Convert GameRules to GameConfig for the GameState
	var gameConfig *GameConfig
	if c.card.GameRules != nil {
		gameConfig = &GameConfig{
			StatsDecayInterval:             time.Duration(c.card.GameRules.StatsDecayInterval) * time.Second,
			CriticalStateAnimationPriority: c.card.GameRules.CriticalStateAnimationPriority,
			MoodBasedAnimations:            c.card.GameRules.MoodBasedAnimations,
		}
	}

	// Initialize game state with stats from character card
	c.gameState = NewGameState(c.card.Stats, gameConfig)

	// Initialize random events manager if random events are configured
	randomEventsEnabled := len(c.card.RandomEvents) > 0
	checkInterval := 30 * time.Second // Default 30 second check interval
	if c.card.GameRules != nil && c.card.GameRules.StatsDecayInterval > 0 {
		// Use the same interval as stats decay for efficiency
		checkInterval = time.Duration(c.card.GameRules.StatsDecayInterval) * time.Second
	}
	c.randomEventManager = NewRandomEventManager(c.card.RandomEvents, randomEventsEnabled, checkInterval)
}

// Update updates character behavior and animations
// Call this regularly (e.g., 60 FPS) to maintain responsive behavior
// Returns true if visual changes occurred (animation frame changed or state changed)
func (c *Character) Update() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Update animation frames
	frameChanged := c.animationManager.Update()
	stateChanged := false

	// Update game state if enabled
	if c.gameState != nil {
		elapsed := time.Since(c.lastStateChange)
		triggeredStates := c.gameState.Update(elapsed)

		// Check for random events
		var triggeredEvent *TriggeredEvent
		if c.randomEventManager != nil {
			triggeredEvent = c.randomEventManager.Update(elapsed, c.gameState)
		}

		// Handle triggered random event
		if triggeredEvent != nil {
			// Apply stat effects
			if triggeredEvent.HasEffects() {
				c.gameState.ApplyInteractionEffects(triggeredEvent.Effects)
			}

			// Trigger animation if specified
			if triggeredEvent.HasAnimations() {
				animation := triggeredEvent.GetRandomAnimation()
				if animation != "" && animation != c.currentState {
					c.setState(animation)
					stateChanged = true
				}
			}

			// Note: Dialog responses from random events would be handled by the UI layer
			// The character behavior only manages state and animation changes
		}

		// Handle critical states with higher priority (if no random event animation)
		if !stateChanged && len(triggeredStates) > 0 {
			newState := c.selectAnimationFromTriggeredStates(triggeredStates)
			if newState != "" && newState != c.currentState {
				c.setState(newState)
				stateChanged = true
			}
		}
	}

	// Check if we should return to idle state (lower priority than game states)
	if !stateChanged && c.currentState != "idle" && time.Since(c.lastStateChange) >= c.idleTimeout {
		idleAnimation := c.selectIdleAnimation()
		c.setState(idleAnimation)
		stateChanged = true
	}

	return frameChanged || stateChanged
}

// GetCurrentFrame returns the current animation frame for rendering
func (c *Character) GetCurrentFrame() image.Image {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.animationManager.GetCurrentFrameImage()
}

// HandleClick processes a click interaction on the character
// Returns dialog text to display, or empty string if no dialog should show
func (c *Character) HandleClick() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.lastInteraction = time.Now()

	// Find click dialog with available cooldown
	for _, dialog := range c.card.Dialogs {
		if dialog.Trigger == "click" {
			lastTrigger, exists := c.dialogCooldowns[dialog.Trigger]
			if !exists || dialog.CanTrigger(lastTrigger) {
				// Trigger this dialog
				c.dialogCooldowns[dialog.Trigger] = time.Now()
				c.setState(dialog.Animation)
				return dialog.GetRandomResponse()
			}
		}
	}

	return "" // No dialog available due to cooldowns
}

// HandleRightClick processes a right-click interaction
func (c *Character) HandleRightClick() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.lastInteraction = time.Now()

	for _, dialog := range c.card.Dialogs {
		if dialog.Trigger == "rightclick" {
			lastTrigger, exists := c.dialogCooldowns[dialog.Trigger]
			if !exists || dialog.CanTrigger(lastTrigger) {
				c.dialogCooldowns[dialog.Trigger] = time.Now()
				c.setState(dialog.Animation)
				return dialog.GetRandomResponse()
			}
		}
	}

	return ""
}

// HandleHover processes a hover interaction
func (c *Character) HandleHover() string {
	c.mu.Lock() // Use write lock to properly synchronize cooldown updates
	defer c.mu.Unlock()

	// Only process hover if not recently interacted
	if time.Since(c.lastInteraction) < 2*time.Second {
		return ""
	}

	for _, dialog := range c.card.Dialogs {
		if dialog.Trigger == "hover" {
			lastTrigger, exists := c.dialogCooldowns[dialog.Trigger]
			if !exists || dialog.CanTrigger(lastTrigger) {
				// Update cooldown to prevent rapid hover spam
				c.dialogCooldowns[dialog.Trigger] = time.Now()
				return dialog.GetRandomResponse()
			}
		}
	}

	return ""
}

// SetPosition updates character position (for draggable characters)
func (c *Character) SetPosition(x, y float32) {
	if !c.movementEnabled {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.x = x
	c.y = y
}

// GetPosition returns current character position
func (c *Character) GetPosition() (float32, float32) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.x, c.y
}

// GetSize returns character display size
func (c *Character) GetSize() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.size
}

// GetName returns character name
func (c *Character) GetName() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.card.Name
}

// GetDescription returns character description
func (c *Character) GetDescription() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.card.Description
}

// IsMovementEnabled returns whether the character can be dragged
func (c *Character) IsMovementEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.movementEnabled
}

// GetCurrentState returns the current animation state
func (c *Character) GetCurrentState() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentState
}

// setState changes the character's animation state (internal method)
func (c *Character) setState(state string) {
	if c.currentState == state {
		return
	}

	// Only change state if the animation exists
	if err := c.animationManager.SetCurrentAnimation(state); err == nil {
		c.currentState = state
		c.lastStateChange = time.Now()
	}
}

// ForceState allows external code to force a specific animation state
// Useful for testing or special behaviors
func (c *Character) ForceState(state string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.animationManager.SetCurrentAnimation(state); err != nil {
		return fmt.Errorf("failed to set state '%s': %w", state, err)
	}

	c.currentState = state
	c.lastStateChange = time.Now()
	return nil
}

// GetAvailableAnimations returns all loaded animation names
func (c *Character) GetAvailableAnimations() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.animationManager.GetLoadedAnimations()
}

// GetDialogCooldownStatus returns cooldown information for debugging
func (c *Character) GetDialogCooldownStatus() map[string]time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()

	status := make(map[string]time.Duration)
	now := time.Now()

	for trigger, lastTime := range c.dialogCooldowns {
		// Find the cooldown duration for this trigger
		for _, dialog := range c.card.Dialogs {
			if dialog.Trigger == trigger {
				cooldownDuration := time.Duration(dialog.Cooldown) * time.Second
				remaining := cooldownDuration - now.Sub(lastTime)
				if remaining < 0 {
					remaining = 0
				}
				status[trigger] = remaining
				break
			}
		}
	}

	return status
}

// EnableGameMode initializes game features for this character
// saveManager can be nil for testing, loadSave specifies save file to load
func (c *Character) EnableGameMode(saveManager interface{}, loadSave string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Only enable if character card has game features
	if !c.card.HasGameFeatures() {
		return fmt.Errorf("character card does not have game features")
	}

	// Initialize game state from character card configuration
	gameConfig := &GameConfig{
		StatsDecayInterval:             time.Duration(c.card.GameRules.StatsDecayInterval) * time.Second,
		CriticalStateAnimationPriority: c.card.GameRules.CriticalStateAnimationPriority,
		MoodBasedAnimations:            c.card.GameRules.MoodBasedAnimations,
	}

	c.gameState = NewGameState(c.card.Stats, gameConfig)

	// Initialize interaction cooldowns for game interactions
	for interactionName := range c.card.Interactions {
		c.gameInteractionCooldowns[interactionName] = time.Time{}
	}

	return nil
}

// HandleGameInteraction processes game-specific interactions (feed, play, pet, etc.)
// Returns response text to display, or empty string if interaction is not available
func (c *Character) HandleGameInteraction(interactionType string) string {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if game mode is enabled
	if c.gameState == nil {
		return ""
	}

	// Find the interaction configuration
	interaction, exists := c.card.Interactions[interactionType]
	if !exists {
		return ""
	}

	// Check cooldown
	lastUsed, exists := c.gameInteractionCooldowns[interactionType]
	if exists && time.Since(lastUsed) < time.Duration(interaction.Cooldown)*time.Second {
		return "" // Still on cooldown
	}

	// Check requirements
	if !c.gameState.CanSatisfyRequirements(interaction.Requirements) {
		return "" // Requirements not met
	}

	// Apply effects
	c.gameState.ApplyInteractionEffects(interaction.Effects)

	// Set cooldown
	c.gameInteractionCooldowns[interactionType] = time.Now()

	// Update last interaction time
	c.lastInteraction = time.Now()

	// Set animation if specified
	if len(interaction.Animations) > 0 {
		// Use first animation for simplicity
		c.setState(interaction.Animations[0])
	}

	// Return random response
	if len(interaction.Responses) > 0 {
		index := int(time.Now().UnixNano()) % len(interaction.Responses)
		return interaction.Responses[index]
	}

	return ""
}

// GetGameState returns the current game state (for testing and UI)
func (c *Character) GetGameState() *GameState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.gameState
}

// selectAnimationFromStates chooses the best animation from triggered states
// Prioritizes critical states if configured to do so
func (c *Character) selectAnimationFromStates(triggeredStates []string) string {
	if len(triggeredStates) == 0 {
		return c.currentState
	}

	// If critical state animation priority is enabled, check for critical states first
	if c.gameState != nil && c.gameState.Config != nil && c.gameState.Config.CriticalStateAnimationPriority {
		for _, state := range triggeredStates {
			if _, exists := c.card.Animations[state]; exists {
				return state
			}
		}
	}

	// Otherwise, use the first available animation
	for _, state := range triggeredStates {
		if _, exists := c.card.Animations[state]; exists {
			return state
		}
	}

	return c.currentState
}

// selectAnimationFromTriggeredStates chooses the best animation based on triggered game states
// Prioritizes critical states and follows configuration priorities
func (c *Character) selectAnimationFromTriggeredStates(triggeredStates []string) string {
	if len(triggeredStates) == 0 {
		return ""
	}

	// Priority mapping for game states to animations
	// Critical states have higher priority
	statePriority := map[string]int{
		"hungry":             2,
		"sad":                2,
		"sick":               3, // Highest priority
		"tired":              1,
		"hunger_critical":    4,
		"happiness_critical": 4,
		"health_critical":    5, // Highest critical priority
		"energy_critical":    3,
	}

	// Find the highest priority state that has an available animation
	bestState := ""
	bestPriority := 0

	for _, state := range triggeredStates {
		priority := statePriority[state]
		if priority > bestPriority {
			// Check if we have an animation for this state
			animationName := c.getAnimationForGameState(state)
			if animationName != "" {
				bestState = animationName
				bestPriority = priority
			}
		}
	}

	return bestState
}

// selectIdleAnimation chooses appropriate idle animation based on mood when moodBasedAnimations is enabled
// Returns "idle" by default, or mood-based animation if configured and available
func (c *Character) selectIdleAnimation() string {
	// Default to idle animation
	defaultIdle := "idle"

	// Check if mood-based animations are enabled
	if c.gameState == nil || c.gameState.Config == nil || !c.gameState.Config.MoodBasedAnimations {
		return defaultIdle
	}

	// Get overall mood (0-100 scale)
	mood := c.gameState.GetOverallMood()

	// Select animation based on mood thresholds
	var moodAnimation string
	switch {
	case mood >= 80:
		moodAnimation = "happy" // Very good mood
	case mood >= 60:
		moodAnimation = "idle" // Normal mood
	case mood >= 40:
		moodAnimation = "idle" // Slightly below normal
	case mood >= 20:
		moodAnimation = "sad" // Low mood
	default:
		moodAnimation = "sad" // Very low mood
	}

	// Check if the selected mood animation exists, fallback to idle if not
	if _, exists := c.card.Animations[moodAnimation]; exists {
		return moodAnimation
	}

	return defaultIdle
}

// getAnimationForGameState maps game states to animation names
// Returns the animation name if available, empty string otherwise
func (c *Character) getAnimationForGameState(state string) string {
	// Map game states to animation names
	stateToAnimation := map[string]string{
		"hungry":             "hungry",
		"sad":                "sad",
		"sick":               "sick",
		"tired":              "tired",
		"hunger_critical":    "hungry",
		"happiness_critical": "sad",
		"health_critical":    "sick",
		"energy_critical":    "tired",
	}

	animationName, exists := stateToAnimation[state]
	if !exists {
		return ""
	}

	// Check if this animation actually exists in the character
	if _, hasAnimation := c.card.Animations[animationName]; hasAnimation {
		return animationName
	}

	return ""
}

// GetGameInteractionCooldowns returns cooldown status for game interactions
func (c *Character) GetGameInteractionCooldowns() map[string]time.Duration {
	if c.gameState == nil {
		return nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	status := make(map[string]time.Duration)
	now := time.Now()

	for interactionType, lastTime := range c.gameInteractionCooldowns {
		if interaction, exists := c.card.Interactions[interactionType]; exists {
			cooldownDuration := time.Duration(interaction.Cooldown) * time.Second
			remaining := cooldownDuration - now.Sub(lastTime)
			if remaining < 0 {
				remaining = 0
			}
			status[interactionType] = remaining
		}
	}

	return status
}

// CanUseGameInteraction checks if a game interaction is currently available
// Returns false if on cooldown, requirements not met, or game features disabled
func (c *Character) CanUseGameInteraction(interactionType string) bool {
	if c.gameState == nil {
		return false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	interaction, exists := c.card.Interactions[interactionType]
	if !exists {
		return false
	}

	// Check cooldown
	if lastUsed, hasCooldown := c.gameInteractionCooldowns[interactionType]; hasCooldown {
		cooldownDuration := time.Duration(interaction.Cooldown) * time.Second
		if time.Since(lastUsed) < cooldownDuration {
			return false
		}
	}

	// Check requirements
	return c.gameState.CanSatisfyRequirements(interaction.Requirements)
}
