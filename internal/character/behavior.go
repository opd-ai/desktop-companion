package character

import (
	"fmt"
	"image"
	"strings"
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
	randomEventManager       *RandomEventManager  // Added for Phase 3 - random events
	romanceEventManager      *RandomEventManager  // Added for Phase 3 Task 2 - romance events
	lastRomanceEventCheck    time.Time            // Last time romance events were checked
	romanceEventCooldowns    map[string]time.Time // Romance event cooldown tracking

	// Advanced features (added for Phase 3 Task 3)
	jealousyManager       *JealousyManager       // Jealousy mechanics and consequences
	compatibilityAnalyzer *CompatibilityAnalyzer // Advanced compatibility algorithms
	crisisRecoveryManager *CrisisRecoveryManager // Relationship crisis and recovery systems

	// Dialog backend integration (Phase 1)
	dialogManager      *DialogManager // Advanced dialog system manager
	useAdvancedDialogs bool           // Whether to use advanced dialog system
	debug              bool           // Debug logging for dialog system
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
		romanceEventCooldowns:    make(map[string]time.Time),
		lastRomanceEventCheck:    time.Now().Add(-30 * time.Second), // Allow immediate first check
		idleTimeout:              time.Duration(card.Behavior.IdleTimeout) * time.Second,
		movementEnabled:          card.Behavior.MovementEnabled,
		size:                     card.Behavior.DefaultSize,
	}

	// Initialize game features if the character card has them
	if card.HasGameFeatures() {
		char.initializeGameFeatures()
	}

	// Initialize dialog system if the character card has backend configuration
	if card.HasDialogBackend() {
		if err := char.initializeDialogSystem(); err != nil {
			return nil, fmt.Errorf("failed to initialize dialog system: %w", err)
		}
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

	// Initialize progression system if configured
	if c.card.Progression != nil {
		c.gameState.SetProgression(c.card.Progression)
	}

	// Initialize random events manager if random events are configured
	randomEventsEnabled := len(c.card.RandomEvents) > 0
	checkInterval := 30 * time.Second // Default 30 second check interval
	if c.card.GameRules != nil && c.card.GameRules.StatsDecayInterval > 0 {
		// Use the same interval as stats decay for efficiency
		checkInterval = time.Duration(c.card.GameRules.StatsDecayInterval) * time.Second
	}
	c.randomEventManager = NewRandomEventManager(c.card.RandomEvents, randomEventsEnabled, checkInterval)

	// Initialize romance events manager if romance events are configured
	romanceEventsEnabled := len(c.card.RomanceEvents) > 0
	c.romanceEventManager = NewRandomEventManager(c.card.RomanceEvents, romanceEventsEnabled, checkInterval)

	// Initialize advanced features if romance features are enabled
	if c.card.HasRomanceFeatures() {
		c.initializeAdvancedFeatures()
	}
}

// initializeAdvancedFeatures sets up Phase 3 Task 3 advanced romance systems
// Called only for characters with romance features enabled
func (c *Character) initializeAdvancedFeatures() {
	// Initialize jealousy mechanics if jealousy-prone personality trait exists
	jealousyProne := c.card.GetPersonalityTrait("jealousy_prone")
	jealousyEnabled := jealousyProne > 0.3 // Enable if character is somewhat jealousy-prone

	// Create default jealousy triggers based on personality
	jealousyTriggers := c.createDefaultJealousyTriggers(jealousyProne)
	jealousyThreshold := 70.0 + (jealousyProne * 20.0) // Threshold from 70-90 based on trait

	c.jealousyManager = NewJealousyManager(jealousyTriggers, jealousyEnabled, jealousyThreshold)

	// Initialize compatibility analyzer with adaptation strength based on personality
	adaptationStrength := 0.5 // Default moderate adaptation
	if affectionResponsiveness := c.card.GetPersonalityTrait("affection_responsiveness"); affectionResponsiveness > 0 {
		adaptationStrength = affectionResponsiveness * 0.8 // Scale to 0-0.8 range
	}

	c.compatibilityAnalyzer = NewCompatibilityAnalyzer(true, adaptationStrength)

	// Initialize crisis recovery manager with personality-based thresholds
	crisisThresholds := c.createPersonalityBasedCrisisThresholds()
	c.crisisRecoveryManager = NewCrisisRecoveryManager(true, crisisThresholds)
}

// createDefaultJealousyTriggers creates jealousy triggers based on personality traits
// Uses JSON-configurable approach but provides sensible defaults
func (c *Character) createDefaultJealousyTriggers(jealousyProne float64) []JealousyTrigger {
	if jealousyProne < 0.3 {
		return []JealousyTrigger{} // Not jealousy-prone enough for triggers
	}

	// Base trigger timing - more jealous characters trigger sooner
	baseTriggerTime := time.Duration(120-int(jealousyProne*60)) * time.Minute // 60-120 minutes

	triggers := []JealousyTrigger{
		{
			Name:              "neglect_jealousy",
			Description:       "Character feels neglected due to lack of interaction",
			InteractionGap:    baseTriggerTime,
			JealousyIncrement: 10.0 + (jealousyProne * 15.0), // 10-25 jealousy based on trait
			TrustPenalty:      2.0 + (jealousyProne * 3.0),   // 2-5 trust penalty
			Conditions: map[string]float64{
				"affection": 20.0, // Must have some affection to feel jealous
			},
			Responses: []string{
				"Where have you been? I've been waiting for you... 😔",
				"Are you spending time with someone else?",
				"I feel like you're ignoring me... 💔",
			},
			Animations:  []string{"jealous", "sad"},
			Probability: 0.3 + (jealousyProne * 0.4), // 30-70% based on trait
		},
		{
			Name:              "attention_jealousy",
			Description:       "Character wants more attention and feels insecure",
			InteractionGap:    baseTriggerTime / 2,          // Triggers more frequently
			JealousyIncrement: 5.0 + (jealousyProne * 10.0), // 5-15 jealousy
			TrustPenalty:      1.0 + (jealousyProne * 2.0),  // 1-3 trust penalty
			Conditions: map[string]float64{
				"affection": 30.0, // Requires moderate affection
				"jealousy":  10.0, // Some existing jealousy
			},
			Responses: []string{
				"Do you still find me interesting? 🥺",
				"I need to know you care about me...",
				"Am I not enough for you anymore? 😢",
			},
			Animations:  []string{"shy", "sad"},
			Probability: 0.2 + (jealousyProne * 0.3), // 20-50% based on trait
		},
	}

	return triggers
}

// createPersonalityBasedCrisisThresholds creates crisis thresholds based on personality
// More sensitive personalities have higher thresholds (trigger crises easier)
func (c *Character) createPersonalityBasedCrisisThresholds() map[string]float64 {
	// Base thresholds
	thresholds := map[string]float64{
		"jealousy":  80.0,
		"trust":     15.0,
		"affection": 10.0,
		"happiness": 20.0,
	}

	// Adjust based on personality traits
	jealousyProne := c.card.GetPersonalityTrait("jealousy_prone")
	trustDifficulty := c.card.GetPersonalityTrait("trust_difficulty")
	affectionResponsiveness := c.card.GetPersonalityTrait("affection_responsiveness")

	// More jealousy-prone characters trigger jealousy crises easier
	thresholds["jealousy"] = 80.0 - (jealousyProne * 20.0) // 60-80 range

	// Characters with trust difficulty trigger trust crises easier
	thresholds["trust"] = 15.0 + (trustDifficulty * 10.0) // 15-25 range

	// Highly affection-responsive characters trigger affection crises easier
	thresholds["affection"] = 10.0 + (affectionResponsiveness * 10.0) // 10-20 range

	return thresholds
}

// initializeDialogSystem sets up the advanced dialog system with configured backends
// Called during character creation if dialog backend configuration is enabled
func (c *Character) initializeDialogSystem() error {
	if c.card.DialogBackend == nil || !c.card.DialogBackend.Enabled {
		return nil
	}

	// Enable debug mode if configured
	c.debug = c.card.DialogBackend.DebugMode

	// Create dialog manager
	c.dialogManager = NewDialogManager(c.debug)
	c.useAdvancedDialogs = true

	// Register available backends
	c.dialogManager.RegisterBackend("simple_random", NewSimpleRandomBackend())
	c.dialogManager.RegisterBackend("markov_chain", NewMarkovChainBackend())

	// Set default backend
	if err := c.dialogManager.SetDefaultBackend(c.card.DialogBackend.DefaultBackend); err != nil {
		return fmt.Errorf("failed to set default backend: %w", err)
	}

	// Set fallback chain if configured
	if len(c.card.DialogBackend.FallbackChain) > 0 {
		if err := c.dialogManager.SetFallbackChain(c.card.DialogBackend.FallbackChain); err != nil {
			return fmt.Errorf("failed to set fallback chain: %w", err)
		}
	}

	// Initialize configured backends with their JSON configurations
	return c.configureBackends()
}

// configureBackends initializes each configured backend with its JSON configuration
func (c *Character) configureBackends() error {
	if c.card.DialogBackend == nil || c.card.DialogBackend.Backends == nil {
		return nil
	}

	for backendName, config := range c.card.DialogBackend.Backends {
		backend, exists := c.dialogManager.backends[backendName]
		if !exists {
			continue // Skip unknown backends
		}

		if err := backend.Initialize(config, c); err != nil {
			return fmt.Errorf("failed to initialize backend '%s': %w", backendName, err)
		}
	}

	return nil
}

// Update updates character behavior and animations
// Call this regularly (e.g., 60 FPS) to maintain responsive behavior
// Returns true if visual changes occurred (animation frame changed or state changed)
func (c *Character) Update() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Update animation frames
	frameChanged := c.animationManager.Update()

	// Process game state updates and check for state changes
	stateChanged := c.processGameStateUpdates()

	// Check for idle timeout if no other state changes occurred
	if !stateChanged {
		stateChanged = c.checkIdleTimeout()
	}

	return frameChanged || stateChanged
}

// processGameStateUpdates handles all game state related updates and returns true if state changed
func (c *Character) processGameStateUpdates() bool {
	if c.gameState == nil {
		return false
	}

	elapsed := time.Since(c.lastStateChange)
	triggeredStates := c.gameState.Update(elapsed)

	// Process random events first (highest priority)
	if c.processRandomEvents(elapsed) {
		return true
	}

	// Handle critical states if no random event occurred
	return c.processCriticalStates(triggeredStates)
}

// processRandomEvents checks and handles random events, returns true if state changed
func (c *Character) processRandomEvents(elapsed time.Duration) bool {
	stateChanged := false

	// Process regular random events
	if c.randomEventManager != nil {
		triggeredEvent := c.randomEventManager.Update(elapsed, c.gameState)
		if triggeredEvent != nil {
			stateChanged = c.handleTriggeredEvent(triggeredEvent) || stateChanged
		}
	}

	// Process romance events with memory-based triggering
	if c.romanceEventManager != nil {
		triggeredEvent := c.processRomanceEvents(elapsed)
		if triggeredEvent != nil {
			stateChanged = c.handleTriggeredEvent(triggeredEvent) || stateChanged
		}
	}

	// Process advanced romance features (Phase 3 Task 3)
	if c.card.HasRomanceFeatures() {
		stateChanged = c.processAdvancedRomanceFeatures() || stateChanged
	}

	return stateChanged
}

// processAdvancedRomanceFeatures handles Phase 3 Task 3 advanced systems
// Processes jealousy, compatibility analysis, and crisis management
func (c *Character) processAdvancedRomanceFeatures() bool {
	stateChanged := false

	// Process jealousy mechanics
	if c.jealousyManager != nil {
		jealousyEvent := c.jealousyManager.Update(c.gameState, c.lastInteraction)
		if jealousyEvent != nil {
			stateChanged = c.handleTriggeredEvent(jealousyEvent) || stateChanged
		}
	}

	// Process compatibility analysis and adaptation
	if c.compatibilityAnalyzer != nil {
		compatibilityModifiers := c.compatibilityAnalyzer.Update(c.gameState)
		if len(compatibilityModifiers) > 0 {
			// Apply compatibility modifiers to future interactions
			// This affects how personality traits influence interaction outcomes
			c.applyCompatibilityModifiers(compatibilityModifiers)
		}
	}

	// Process crisis management
	if c.crisisRecoveryManager != nil {
		crisisEvent, inCrisis := c.crisisRecoveryManager.Update(c.gameState)
		if crisisEvent != nil {
			stateChanged = c.handleTriggeredEvent(crisisEvent) || stateChanged
		}

		// Store crisis state for other systems to use
		c.setInCrisisMode(inCrisis)
	}

	return stateChanged
}

// applyCompatibilityModifiers updates character behavior based on compatibility analysis
// Modifies interaction effectiveness based on learned player preferences
func (c *Character) applyCompatibilityModifiers(modifiers []CompatibilityModifier) {
	// Store modifiers for use in personality calculations
	// This is a simple approach - in a more complex system, these could be
	// stored in the game state or character card for persistence

	// For now, we apply them by temporarily adjusting personality traits
	// This demonstrates the concept without requiring complex state management
	for _, modifier := range modifiers {
		switch modifier.StatName {
		case "consistent_interaction_bonus":
			// Boost consistent interaction compatibility
			if c.card.Personality != nil && c.card.Personality.Compatibility != nil {
				c.card.Personality.Compatibility["consistent_interaction"] *= modifier.ModifierValue
			}
		case "variety_preference_bonus":
			// Boost variety preference
			if c.card.Personality != nil && c.card.Personality.Compatibility != nil {
				c.card.Personality.Compatibility["variety_preference"] *= modifier.ModifierValue
			}
		case "interaction_responsiveness_bonus":
			// General interaction responsiveness boost
			if c.card.Personality != nil && c.card.Personality.Traits != nil {
				c.card.Personality.Traits["affection_responsiveness"] *= modifier.ModifierValue
			}
		}
	}
}

// setInCrisisMode updates character state to reflect crisis mode
// Can be used by other systems to adjust behavior during crises
func (c *Character) setInCrisisMode(inCrisis bool) {
	// For now, this is just a placeholder for crisis state management
	// In a more complex system, this could affect dialogue selection,
	// animation priorities, interaction availability, etc.

	// The crisis state is already being handled by the crisis manager's
	// ongoing effects and event generation
	_ = inCrisis // Placeholder to prevent unused variable warning
}

// processRomanceEvents handles romance-specific random events with memory-based triggering
func (c *Character) processRomanceEvents(elapsed time.Duration) *TriggeredEvent {
	if c.romanceEventManager == nil || c.gameState == nil {
		return nil
	}

	// Use custom romance event processing that supports enhanced conditions
	return c.checkAndTriggerRomanceEvent(elapsed)
}

// checkAndTriggerRomanceEvent implements romance-specific event logic with memory-based conditions
func (c *Character) checkAndTriggerRomanceEvent(elapsed time.Duration) *TriggeredEvent {
	// Note: c.mu is already locked by the caller (Update method)

	// Check if enough time has passed since last romance event check
	now := time.Now()
	checkInterval := 30 * time.Second // Same as random events
	if c.lastRomanceEventCheck.Add(checkInterval).After(now) {
		return nil
	}
	c.lastRomanceEventCheck = now

	// Iterate through romance events and try to trigger one
	for _, event := range c.card.RomanceEvents {
		if c.canTriggerRomanceEvent(event, now) {
			if c.rollEventProbability(event.Probability) {
				return c.createTriggeredRomanceEvent(event, now)
			}
		}
	}

	return nil
}

// canTriggerRomanceEvent checks if a romance event can trigger using enhanced condition checking
func (c *Character) canTriggerRomanceEvent(event RandomEventConfig, now time.Time) bool {
	// Check event-specific cooldown
	if lastTrigger, exists := c.romanceEventCooldowns[event.Name]; exists {
		cooldownDuration := time.Duration(event.Cooldown) * time.Second
		if now.Sub(lastTrigger) < cooldownDuration {
			return false
		}
	}

	// Use enhanced romance condition checking
	if len(event.Conditions) > 0 {
		return c.gameState.CanSatisfyRomanceRequirements(event.Conditions)
	}

	return true
}

// rollEventProbability performs probability check for romance event triggering
func (c *Character) rollEventProbability(probability float64) bool {
	// Simple probability check using time-based pseudo-randomness
	randomValue := float64((time.Now().UnixNano() % 10000)) / 10000.0
	return randomValue <= probability
}

// createTriggeredRomanceEvent creates a triggered romance event and records the cooldown
func (c *Character) createTriggeredRomanceEvent(event RandomEventConfig, now time.Time) *TriggeredEvent {
	// Initialize cooldowns map if needed
	if c.romanceEventCooldowns == nil {
		c.romanceEventCooldowns = make(map[string]time.Time)
	}

	// Record the cooldown
	c.romanceEventCooldowns[event.Name] = now

	return &TriggeredEvent{
		Name:        event.Name,
		Description: event.Description,
		Effects:     event.Effects,
		Animations:  event.Animations,
		Responses:   event.Responses,
		Duration:    time.Duration(event.Duration) * time.Second,
	}
}

// createEnhancedGameStateForRomanceEvents creates a game state with romance context
func (c *Character) createEnhancedGameStateForRomanceEvents() *GameState {
	// For romance events, we use the same game state but the romance event manager
	// will use CanSatisfyRomanceRequirements instead of CanSatisfyRequirements
	// This allows memory-based and relationship-aware condition checking
	return c.gameState
}

// handleTriggeredEvent processes a triggered random event and returns true if state changed
func (c *Character) handleTriggeredEvent(triggeredEvent *TriggeredEvent) bool {
	// Apply stat effects
	if triggeredEvent.HasEffects() {
		c.gameState.ApplyInteractionEffects(triggeredEvent.Effects)
	}

	// Trigger animation if specified
	if triggeredEvent.HasAnimations() {
		animation := triggeredEvent.GetRandomAnimation()
		if animation != "" && animation != c.currentState {
			c.setState(animation)
			return true
		}
	}

	return false
}

// processCriticalStates handles critical game states and returns true if state changed
func (c *Character) processCriticalStates(triggeredStates []string) bool {
	if len(triggeredStates) == 0 {
		return false
	}

	newState := c.selectAnimationFromTriggeredStates(triggeredStates)
	if newState != "" && newState != c.currentState {
		c.setState(newState)
		return true
	}

	return false
}

// checkIdleTimeout checks if character should return to idle state
func (c *Character) checkIdleTimeout() bool {
	if c.currentState == "idle" || time.Since(c.lastStateChange) < c.idleTimeout {
		return false
	}

	idleAnimation := c.selectIdleAnimation()
	c.setState(idleAnimation)
	return true
}

// GetCurrentFrame returns the current animation frame for rendering
func (c *Character) GetCurrentFrame() image.Image {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.animationManager.GetCurrentFrameImage()
}

// HandleClick processes a click interaction on the character
// Returns dialog text to display, or empty string if no dialog should show
// HandleClick processes a click interaction, using advanced dialog system if enabled
func (c *Character) HandleClick() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.lastInteraction = time.Now()

	// Try advanced dialog system first
	if c.useAdvancedDialogs && c.dialogManager != nil {
		context := c.buildDialogContext("click")
		response, err := c.dialogManager.GenerateDialog(context)
		if err == nil && response.Confidence >= c.card.DialogBackend.ConfidenceThreshold {
			c.setState(response.Animation)
			// Update dialog memory for learning if enabled
			if c.card.DialogBackend.MemoryEnabled {
				c.updateDialogMemory(response, context)
			}
			return response.Text
		}
	}

	// Fallback to existing logic
	return c.handleClickFallback()
}

// handleClickFallback implements the original click handling logic
func (c *Character) handleClickFallback() string {
	// First check romance dialogs if romance features are enabled
	if c.card.HasRomanceFeatures() && c.gameState != nil {
		response := c.selectRomanceDialog("click")
		if response != "" {
			return response
		}
	}

	// Fall back to regular dialogs
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

// HandleRightClick processes a right-click interaction, using advanced dialog system if enabled
func (c *Character) HandleRightClick() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.lastInteraction = time.Now()

	// Try advanced dialog system first
	if c.useAdvancedDialogs && c.dialogManager != nil {
		context := c.buildDialogContext("rightclick")
		response, err := c.dialogManager.GenerateDialog(context)
		if err == nil && response.Confidence >= c.card.DialogBackend.ConfidenceThreshold {
			c.setState(response.Animation)
			// Update dialog memory for learning if enabled
			if c.card.DialogBackend.MemoryEnabled {
				c.updateDialogMemory(response, context)
			}
			return response.Text
		}
	}

	// Fallback to existing logic
	return c.handleRightClickFallback()
}

// handleRightClickFallback implements the original right-click handling logic
func (c *Character) handleRightClickFallback() string {
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

	// First check romance dialogs if romance features are enabled
	if c.card.HasRomanceFeatures() && c.gameState != nil {
		response := c.selectRomanceDialog("hover")
		if response != "" {
			return response
		}
	}

	// Fall back to regular dialogs
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

// HandleRomanceInteraction processes romance-specific interactions (compliment, gift, conversation, etc.)
// Returns response text to display, or empty string if interaction is not available
// This implements the missing runtime functionality for the JSON-configured romance system
func (c *Character) HandleRomanceInteraction(interactionType string) string {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Validate interaction preconditions
	interaction, ok := c.validateRomanceInteraction(interactionType)
	if !ok {
		return ""
	}

	// Check interaction requirements
	if !c.checkRomanceRequirements(interaction, interactionType) {
		return c.getFailureResponse(interactionType)
	}

	// Process the interaction effects and record stats
	response := c.processRomanceEffects(interaction, interactionType)

	// Handle post-interaction updates
	c.handlePostRomanceInteraction(interaction, interactionType)

	// Check for crisis recovery and return appropriate response
	return c.checkCrisisRecoveryResponse(interaction, interactionType, response)
}

// validateRomanceInteraction checks if the romance interaction is valid and available
func (c *Character) validateRomanceInteraction(interactionType string) (InteractionConfig, bool) {
	// Check if game mode is enabled and romance features are available
	if c.gameState == nil || !c.card.HasRomanceFeatures() {
		return InteractionConfig{}, false
	}

	// Find the interaction configuration
	interaction, exists := c.card.Interactions[interactionType]
	if !exists {
		return InteractionConfig{}, false
	}

	// Check if this is a romance interaction by examining the effects
	// Romance interactions should affect romance-specific stats
	if !c.isRomanceInteraction(interaction) {
		return InteractionConfig{}, false
	}

	return interaction, true
}

// checkRomanceRequirements verifies cooldown and prerequisite requirements
func (c *Character) checkRomanceRequirements(interaction InteractionConfig, interactionType string) bool {
	// Check cooldown
	lastUsed, exists := c.gameInteractionCooldowns[interactionType]
	if exists && time.Since(lastUsed) < time.Duration(interaction.Cooldown)*time.Second {
		return false
	}

	// Check requirements
	return c.gameState.CanSatisfyRequirements(interaction.Requirements)
}

// processRomanceEffects handles personality modification, effect application, and response generation
func (c *Character) processRomanceEffects(interaction InteractionConfig, interactionType string) string {
	// Calculate personality modifier for effects
	personalityModifier := c.calculatePersonalityModifier(interactionType)

	// Apply personality-modified effects
	modifiedEffects := c.applyPersonalityToEffects(interaction.Effects, personalityModifier)

	// Record stats before interaction for memory system
	statsBefore := c.gameState.GetStats()

	// Apply effects
	c.gameState.ApplyInteractionEffects(modifiedEffects)

	// Record stats after interaction
	statsAfter := c.gameState.GetStats()

	// Record the romance interaction for memory system
	response := c.selectContextualResponse(interaction.Responses, interactionType)
	c.recordRomanceInteraction(interactionType, response, statsBefore, statsAfter)

	return response
}

// handlePostRomanceInteraction manages progression, cooldowns, and animations after interaction
func (c *Character) handlePostRomanceInteraction(interaction InteractionConfig, interactionType string) {
	// Check for relationship level progression
	c.updateRelationshipProgression()

	// Set cooldown and update interaction time
	c.updateInteractionCooldown(interactionType)

	// Set appropriate animation
	c.setRomanceAnimation(interaction)
}

// updateRelationshipProgression checks and handles relationship level changes
func (c *Character) updateRelationshipProgression() {
	if c.card.Progression != nil {
		levelChanged := c.gameState.UpdateRelationshipLevel(c.card.Progression)
		if levelChanged {
			// Log level change for debugging
			newLevel := c.gameState.GetRelationshipLevel()
			c.setState("level_up") // Could trigger special animation

			// You could also trigger a special level-up response here
			_ = newLevel // Use for potential level-up dialogue
		}
	}
}

// updateInteractionCooldown sets cooldown and updates last interaction time
func (c *Character) updateInteractionCooldown(interactionType string) {
	// Set cooldown
	c.gameInteractionCooldowns[interactionType] = time.Now()

	// Update last interaction time
	c.lastInteraction = time.Now()
}

// setRomanceAnimation selects and sets animation based on interaction configuration
func (c *Character) setRomanceAnimation(interaction InteractionConfig) {
	// Select animation based on personality and context
	if len(interaction.Animations) > 0 {
		animationIndex := c.selectRomanceAnimation(interaction.Animations)
		c.setState(interaction.Animations[animationIndex])
	}
}

// checkCrisisRecoveryResponse handles crisis recovery and returns the final response
func (c *Character) checkCrisisRecoveryResponse(interaction InteractionConfig, interactionType, defaultResponse string) string {
	// Check for crisis recovery (Phase 3 Task 3)
	if c.crisisRecoveryManager != nil {
		recoveryEvent := c.crisisRecoveryManager.CheckRecovery(c.gameState, interactionType)
		if recoveryEvent != nil {
			// Crisis was resolved! Override normal response with recovery response
			c.handleTriggeredEvent(recoveryEvent)
			if len(recoveryEvent.Responses) > 0 {
				recoveryResponse := recoveryEvent.Responses[int(time.Now().UnixNano())%len(recoveryEvent.Responses)]
				return recoveryResponse
			}
		}
	}

	// Return contextual response based on personality and current relationship level
	if len(interaction.Responses) > 0 {
		return c.selectContextualResponse(interaction.Responses, interactionType)
	}

	return defaultResponse
}

// isRomanceInteraction determines if an interaction is romance-related by checking its effects
// Romance interactions are identified by affecting romance-specific stats
func (c *Character) isRomanceInteraction(interaction InteractionConfig) bool {
	romanceStats := map[string]bool{
		"affection": true,
		"trust":     true,
		"intimacy":  true,
		"jealousy":  true,
	}

	// Check if any of the interaction effects target romance stats
	for statName := range interaction.Effects {
		if romanceStats[statName] {
			return true
		}
	}

	return false
}

// calculatePersonalityModifier calculates how personality traits affect interaction effects
// Uses existing personality configuration from character card
func (c *Character) calculatePersonalityModifier(interactionType string) float64 {
	baseModifier := 1.0

	// Get compatibility modifier from character card
	compatibilityModifier := c.card.GetCompatibilityModifier(interactionType)
	baseModifier *= compatibilityModifier

	// Apply trait-specific modifiers based on interaction type
	switch interactionType {
	case "compliment":
		// Shy characters are less responsive to compliments initially
		shyness := c.card.GetPersonalityTrait("shyness")
		affectionResponsiveness := c.card.GetPersonalityTrait("affection_responsiveness")
		baseModifier *= (1.0 - shyness*0.3) * affectionResponsiveness

	case "give_gift":
		// Gift appreciation trait directly affects gift interactions
		giftAppreciation := c.card.GetCompatibilityModifier("gift_appreciation")
		baseModifier *= giftAppreciation

	case "deep_conversation":
		// Conversation lovers get more benefit from deep talks
		conversationLover := c.card.GetCompatibilityModifier("conversation_lover")
		baseModifier *= conversationLover

	default:
		// Use general affection responsiveness for other romance interactions
		affectionResponsiveness := c.card.GetPersonalityTrait("affection_responsiveness")
		baseModifier *= affectionResponsiveness
	}

	return baseModifier
}

// applyPersonalityToEffects applies personality modifiers to stat effects
// Ensures personality traits influence the actual stat changes
func (c *Character) applyPersonalityToEffects(effects map[string]float64, modifier float64) map[string]float64 {
	modifiedEffects := make(map[string]float64)

	for statName, value := range effects {
		// Romance stats get personality modifiers, basic stats don't
		if statName == "affection" || statName == "trust" || statName == "intimacy" {
			modifiedEffects[statName] = value * modifier
		} else {
			modifiedEffects[statName] = value
		}
	}

	return modifiedEffects
}

// selectRomanceAnimation chooses an animation based on personality and context
// Uses personality traits to influence animation selection for more character-consistent behavior
func (c *Character) selectRomanceAnimation(animations []string) int {
	if len(animations) == 0 {
		return 0
	}

	// Simple personality-influenced selection
	shyness := c.card.GetPersonalityTrait("shyness")
	flirtiness := c.card.GetPersonalityTrait("flirtiness")

	// Shy characters prefer subtle animations, flirty characters prefer bold ones
	for i, animation := range animations {
		if shyness > 0.7 && (animation == "shy" || animation == "blushing") {
			return i
		}
		if flirtiness > 0.7 && (animation == "flirty" || animation == "heart_eyes") {
			return i
		}
	}

	// Default to time-based pseudo-random selection
	return int(time.Now().UnixNano()) % len(animations)
}

// selectContextualResponse chooses a response based on personality and relationship context
// Provides more character-consistent dialogue based on current stats and personality
func (c *Character) selectContextualResponse(responses []string, interactionType string) string {
	if len(responses) == 0 {
		return ""
	}

	// For romance interactions, consider relationship level
	if c.gameState != nil && len(c.gameState.Stats) > 0 {
		affection := 0.0
		if affectionStat, exists := c.gameState.Stats["affection"]; exists {
			affection = affectionStat.Current
		}

		// Higher affection characters might have different response styles
		romanticism := c.card.GetPersonalityTrait("romanticism")

		// Romantic characters with high affection use sweeter responses
		if romanticism > 0.6 && affection > 40 && len(responses) > 1 {
			// Prefer responses with romantic language (rough heuristic)
			for i, response := range responses {
				if len(response) > 20 { // Longer responses tend to be more romantic
					return response
				}
				_ = i // Use index if needed for more sophisticated selection
			}
		}
	}

	// Default pseudo-random selection
	index := int(time.Now().UnixNano()) % len(responses)
	return responses[index]
}

// getFailureResponse returns an appropriate response when romance interaction fails
// Provides personality-consistent feedback for failed interactions
func (c *Character) getFailureResponse(interactionType string) string {
	shyness := c.card.GetPersonalityTrait("shyness")
	trustDifficulty := c.card.GetPersonalityTrait("trust_difficulty")

	// Customize failure messages based on personality
	if shyness > 0.7 {
		return "I'm... I'm not quite ready for that yet... 😳"
	}

	if trustDifficulty > 0.6 {
		return "I need to feel more comfortable first..."
	}

	// Default failure responses
	failureResponses := map[string][]string{
		"compliment":        {"I'm not sure how to respond to that right now...", "Maybe when I know you better?"},
		"give_gift":         {"That's very thoughtful, but I can't accept that yet.", "Perhaps when we're closer?"},
		"deep_conversation": {"I'm not ready for such deep talks yet.", "Let's start with lighter conversation?"},
	}

	if responses, exists := failureResponses[interactionType]; exists && len(responses) > 0 {
		index := int(time.Now().UnixNano()) % len(responses)
		return responses[index]
	}

	return "I'm not ready for that right now..."
}

// recordRomanceInteraction records romance interactions for memory and progression tracking
// Implements the romance memory system outlined in the plan
func (c *Character) recordRomanceInteraction(interactionType, response string, statsBefore, statsAfter map[string]float64) {
	if c.gameState == nil {
		return
	}

	// Record detailed interaction memory
	c.gameState.RecordRomanceInteraction(interactionType, response, statsBefore, statsAfter)

	// Record interaction for progression tracking
	c.gameState.RecordInteraction(interactionType)
}

// zeroOutAllCooldowns resets all interaction cooldowns to zero for testing purposes
// This allows tests to perform multiple interactions rapidly without cooldown restrictions
func (c *Character) zeroOutAllCooldowns() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Clear dialog cooldowns
	c.dialogCooldowns = make(map[string]time.Time)

	// Clear game interaction cooldowns
	c.gameInteractionCooldowns = make(map[string]time.Time)
}

// GetGameState returns the current game state (for testing and UI)
func (c *Character) GetGameState() *GameState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.gameState
}

// selectRomanceDialog selects an appropriate romance dialog based on relationship context
// This implements the enhanced dialogue system from Phase 2 of the dating simulator plan
func (c *Character) selectRomanceDialog(trigger string) string {
	if c.card.RomanceDialogs == nil || len(c.card.RomanceDialogs) == 0 {
		return ""
	}

	// Find romance dialogs matching the trigger
	var availableDialogs []DialogExtended
	for _, dialog := range c.card.RomanceDialogs {
		if dialog.Trigger == trigger {
			// Check if requirements are satisfied
			if c.canSatisfyRomanceRequirements(dialog.Requirements) {
				// Check cooldown
				lastTrigger, exists := c.dialogCooldowns[dialog.Trigger]
				if !exists || dialog.CanTrigger(lastTrigger) {
					availableDialogs = append(availableDialogs, dialog)
				}
			}
		}
	}

	// If no romance dialogs are available, return empty
	if len(availableDialogs) == 0 {
		return ""
	}

	// Select the best dialog based on relationship context
	selectedDialog := c.selectBestRomanceDialog(availableDialogs)

	// Trigger the selected dialog
	c.dialogCooldowns[selectedDialog.Trigger] = time.Now()
	c.setState(selectedDialog.Animation)
	return selectedDialog.GetRandomResponse()
}

// canSatisfyRomanceRequirements checks if the current game state satisfies romance requirements
// Uses existing requirements system but specifically for romance dialogs
func (c *Character) canSatisfyRomanceRequirements(requirements *RomanceRequirement) bool {
	if requirements == nil {
		return true // No requirements means always available
	}

	return c.checkStatRequirements(requirements.Stats) &&
		c.checkRelationshipLevel(requirements.RelationshipLevel) &&
		c.checkInteractionCounts(requirements.InteractionCount)
}

// checkStatRequirements validates that all stat requirements are satisfied
func (c *Character) checkStatRequirements(stats map[string]map[string]float64) bool {
	if stats == nil {
		return true
	}

	for statName, conditions := range stats {
		if !c.gameState.CanSatisfyRequirements(map[string]map[string]float64{statName: conditions}) {
			return false
		}
	}
	return true
}

// checkRelationshipLevel validates the relationship level requirement
func (c *Character) checkRelationshipLevel(requiredLevel string) bool {
	if requiredLevel == "" {
		return true
	}

	// For Phase 2, we'll use progression level as relationship level
	// This can be enhanced later with dedicated relationship level tracking
	if c.gameState.Progression != nil {
		currentLevel := c.gameState.Progression.CurrentLevel
		return currentLevel == requiredLevel
	}
	return true
}

// checkInteractionCounts validates all interaction count requirements
func (c *Character) checkInteractionCounts(interactionRequirements map[string]map[string]int) bool {
	if interactionRequirements == nil {
		return true
	}

	// Use progression system's interaction counts
	if c.gameState.Progression == nil {
		return true
	}

	interactionCounts := c.gameState.Progression.GetInteractionCounts()
	for interactionType, conditions := range interactionRequirements {
		if !c.validateInteractionConditions(interactionCounts[interactionType], conditions) {
			return false
		}
	}
	return true
}

// validateInteractionConditions checks if interaction count meets all specified conditions
func (c *Character) validateInteractionConditions(currentCount int, conditions map[string]int) bool {
	for conditionType, threshold := range conditions {
		switch conditionType {
		case "min":
			if currentCount < threshold {
				return false
			}
		case "max":
			if currentCount > threshold {
				return false
			}
		}
	}
	return true
}

// selectBestRomanceDialog chooses the most appropriate dialog from available options
// Considers personality traits and current relationship context
func (c *Character) selectBestRomanceDialog(availableDialogs []DialogExtended) DialogExtended {
	if len(availableDialogs) == 1 {
		return availableDialogs[0]
	}

	// Preference scoring based on personality and relationship state
	bestDialog := availableDialogs[0]
	bestScore := 0.0

	for _, dialog := range availableDialogs {
		score := c.calculateDialogScore(dialog)
		if score > bestScore {
			bestScore = score
			bestDialog = dialog
		}
	}

	return bestDialog
}

// calculateDialogScore calculates a score for a dialog based on personality and context
// Higher scores indicate better matches for the current character state
func (c *Character) calculateDialogScore(dialog DialogExtended) float64 {
	baseScore := 1.0

	// Extract personality traits needed for scoring
	shyness := c.card.GetPersonalityTrait("shyness")
	romanticism := c.card.GetPersonalityTrait("romanticism")
	flirtiness := c.card.GetPersonalityTrait("flirtiness")

	// Get current affection level for context
	affection := c.extractCurrentAffection()

	// Apply personality-based scoring adjustments
	baseScore = c.applyPersonalityScoring(dialog, baseScore, shyness, romanticism, flirtiness)

	// Apply affection-based scoring adjustments
	baseScore = c.applyAffectionScoring(baseScore, affection, romanticism)

	return baseScore
}

// extractCurrentAffection retrieves the current affection level from game state
func (c *Character) extractCurrentAffection() float64 {
	if affectionStat, exists := c.gameState.Stats["affection"]; exists {
		return affectionStat.Current
	}
	return 0.0
}

// applyPersonalityScoring adjusts dialog score based on personality traits and response content
func (c *Character) applyPersonalityScoring(dialog DialogExtended, baseScore, shyness, romanticism, flirtiness float64) float64 {
	responses := dialog.Responses
	if len(responses) == 0 {
		return baseScore
	}

	response := responses[0] // Use first response as representative

	// Apply romantic content preferences
	baseScore = c.applyRomanticContentScoring(response, baseScore, romanticism)

	// Apply shyness-based response length preferences
	baseScore = c.applyShynessScoring(response, baseScore, shyness)

	// Apply flirtiness and boldness preferences
	baseScore = c.applyFlirtinessScoring(response, baseScore, flirtiness, shyness)

	return baseScore
}

// applyRomanticContentScoring adjusts score based on romantic content preference
func (c *Character) applyRomanticContentScoring(response string, baseScore, romanticism float64) float64 {
	if romanticism > 0.6 && (len(response) > 30 || strings.Contains(response, "💕") || strings.Contains(response, "💖")) {
		baseScore += romanticism * 0.5 // Reduce romanticism bonus to balance with shyness
	}
	return baseScore
}

// applyShynessScoring adjusts score based on response length preferences for shy characters
func (c *Character) applyShynessScoring(response string, baseScore, shyness float64) float64 {
	if shyness > 0.6 {
		if len(response) < 25 {
			baseScore += shyness // Bonus for short responses
		} else {
			baseScore -= shyness * 0.5 // Penalty for long responses
		}
	}
	return baseScore
}

// applyFlirtinessScoring adjusts score based on bold expression preferences
func (c *Character) applyFlirtinessScoring(response string, baseScore, flirtiness, shyness float64) float64 {
	if strings.Contains(response, "*boldly*") || strings.Contains(response, "😘") {
		if flirtiness > 0.6 {
			baseScore += flirtiness
		} else if shyness > 0.6 {
			baseScore -= shyness * 0.5 // Shy characters avoid bold expressions
		}
	}
	return baseScore
}

// applyAffectionScoring adjusts score based on current affection level and romantic character traits
func (c *Character) applyAffectionScoring(baseScore, affection, romanticism float64) float64 {
	if affection > 50 && romanticism > 0.5 {
		baseScore += 0.5 // Boost romantic dialogs for high-affection romantic characters
	}
	return baseScore
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

// buildDialogContext creates a comprehensive context for dialog generation
func (c *Character) buildDialogContext(trigger string) DialogContext {
	context := DialogContext{
		Trigger:       trigger,
		InteractionID: fmt.Sprintf("%s_%d", trigger, time.Now().UnixNano()),
		Timestamp:     time.Now(),
	}

	// Add character state context
	if c.gameState != nil {
		context.CurrentStats = c.gameState.GetStats()
		context.CurrentMood = c.gameState.GetOverallMood()
		context.RelationshipLevel = c.gameState.GetRelationshipLevel()
		context.InteractionHistory = c.buildInteractionHistory()
	}

	// Add personality traits
	if c.card.Personality != nil && c.card.Personality.Traits != nil {
		context.PersonalityTraits = make(map[string]float64)
		for trait, value := range c.card.Personality.Traits {
			context.PersonalityTraits[trait] = value
		}
	}

	// Set current animation
	context.CurrentAnimation = c.currentState

	// Add time of day context
	context.TimeOfDay = c.getTimeOfDay()

	// Add fallback responses from existing dialogs
	context.FallbackResponses = c.getFallbackResponses(trigger)
	context.FallbackAnimation = c.getFallbackAnimation(trigger)

	return context
}

// buildInteractionHistory builds a recent interaction history for context
func (c *Character) buildInteractionHistory() []InteractionRecord {
	// For now, return empty history - future enhancement could track interactions
	// This would integrate with the existing game state memory system
	return []InteractionRecord{}
}

// getTimeOfDay returns a simple time of day categorization
func (c *Character) getTimeOfDay() string {
	hour := time.Now().Hour()
	switch {
	case hour >= 6 && hour < 12:
		return "morning"
	case hour >= 12 && hour < 18:
		return "afternoon"
	case hour >= 18 && hour < 22:
		return "evening"
	default:
		return "night"
	}
}

// getFallbackResponses gets fallback responses from existing dialogs for the trigger
func (c *Character) getFallbackResponses(trigger string) []string {
	var responses []string

	// Collect from basic dialogs
	for _, dialog := range c.card.Dialogs {
		if dialog.Trigger == trigger {
			responses = append(responses, dialog.Responses...)
		}
	}

	// Collect from romance dialogs if available
	if c.card.HasRomanceFeatures() {
		for _, dialog := range c.card.RomanceDialogs {
			if dialog.Trigger == trigger {
				responses = append(responses, dialog.Responses...)
			}
		}
	}

	return responses
}

// getFallbackAnimation gets the fallback animation for a trigger
func (c *Character) getFallbackAnimation(trigger string) string {
	// Check basic dialogs
	for _, dialog := range c.card.Dialogs {
		if dialog.Trigger == trigger && dialog.Animation != "" {
			return dialog.Animation
		}
	}

	// Check romance dialogs
	for _, dialog := range c.card.RomanceDialogs {
		if dialog.Trigger == trigger && dialog.Animation != "" {
			return dialog.Animation
		}
	}

	// Default animation based on trigger
	switch trigger {
	case "click":
		return "talking"
	case "rightclick":
		return "thinking"
	case "hover":
		return "idle"
	default:
		return "talking"
	}
}

// updateDialogMemory records dialog interactions for learning and adaptation
func (c *Character) updateDialogMemory(response DialogResponse, context DialogContext) {
	if c.gameState == nil || !c.card.DialogBackend.MemoryEnabled {
		return
	}

	// Record high-importance responses in character memory
	if response.MemoryImportance > 0.7 {
		// This would integrate with the existing memory system
		// For now, we could record in the game state or a separate dialog memory system

		// Future enhancement: implement DialogMemory struct and storage
		// c.gameState.RecordDialogMemory(DialogMemory{
		//     Text: response.Text,
		//     Context: context.Trigger,
		//     Timestamp: time.Now(),
		//     EmotionalTone: response.EmotionalTone,
		//     Topics: response.Topics,
		// })
	}

	// Update backend memory for learning if enabled
	if c.card.DialogBackend.LearningEnabled && c.dialogManager != nil {
		// For now, we don't have user feedback, so we pass nil
		// Future enhancement could track user interactions to determine positive/negative feedback
		c.dialogManager.UpdateBackendMemory(context, response, nil)
	}
}
