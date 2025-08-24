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
}

// New creates a new character instance from a character card
// Loads all animations and initializes behavior state
func New(card *CharacterCard, basePath string) (*Character, error) {
	char := &Character{
		card:             card,
		animationManager: NewAnimationManager(),
		basePath:         basePath,
		currentState:     "idle",
		lastStateChange:  time.Now(),
		lastInteraction:  time.Now(),
		dialogCooldowns:  make(map[string]time.Time),
		idleTimeout:      time.Duration(card.Behavior.IdleTimeout) * time.Second,
		movementEnabled:  card.Behavior.MovementEnabled,
		size:             card.Behavior.DefaultSize,
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

// Update updates character behavior and animations
// Call this regularly (e.g., 60 FPS) to maintain responsive behavior
func (c *Character) Update() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Update animation frames
	c.animationManager.Update()

	// Check if we should return to idle state
	if c.currentState != "idle" && time.Since(c.lastStateChange) >= c.idleTimeout {
		c.setState("idle")
	}
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
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Only process hover if not recently interacted
	if time.Since(c.lastInteraction) < 2*time.Second {
		return ""
	}

	for _, dialog := range c.card.Dialogs {
		if dialog.Trigger == "hover" {
			lastTrigger, exists := c.dialogCooldowns[dialog.Trigger]
			if !exists || dialog.CanTrigger(lastTrigger) {
				// Note: We don't update cooldown here to avoid write lock
				// Hover should be less aggressive than clicks
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
