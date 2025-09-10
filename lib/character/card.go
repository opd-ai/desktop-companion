package character

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/opd-ai/desktop-companion/lib/bot"
	"github.com/opd-ai/desktop-companion/lib/dialog"
	"github.com/opd-ai/desktop-companion/lib/news"
)

// Battle animation constants based on JRPG Battle System plan
// These animations are optional for backward compatibility
const (
	// Core required animations (existing)
	AnimationIdle    = "idle"
	AnimationTalking = "talking"

	// Battle animations (optional)
	AnimationAttack  = "attack"  // Aggressive forward motion
	AnimationDefend  = "defend"  // Protective blocking stance
	AnimationStun    = "stun"    // Dizzied/stunned state
	AnimationHeal    = "heal"    // Glowing recovery animation
	AnimationBoost   = "boost"   // Power-up energy effect
	AnimationCounter = "counter" // Reactive counter-attack
	AnimationDrain   = "drain"   // Energy absorption visual
	AnimationShield  = "shield"  // Barrier/shield formation
	AnimationCharge  = "charge"  // Building energy/power
	AnimationEvade   = "evade"   // Quick dodge movement
	AnimationTaunt   = "taunt"   // Provocative gesture
	AnimationVictory = "victory" // Battle won celebration
)

// CharacterCard represents the JSON configuration for a desktop companion character
// This follows the "lazy programmer" approach - leveraging Go's built-in JSON package
// instead of writing custom parsers
type CharacterCard struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Animations  map[string]string `json:"animations"`
	Dialogs     []Dialog          `json:"dialogs"`
	Behavior    Behavior          `json:"behavior"`
	// Game feature extensions (Phase 1 implementation)
	Stats        map[string]StatConfig        `json:"stats,omitempty"`
	GameRules    *GameRulesConfig             `json:"gameRules,omitempty"`
	Interactions map[string]InteractionConfig `json:"interactions,omitempty"`
	// Progression features (Phase 3 implementation)
	Progression *ProgressionConfig `json:"progression,omitempty"`
	// Random events (Phase 3 implementation)
	RandomEvents []RandomEventConfig `json:"randomEvents,omitempty"`
	// Romance feature extensions (Dating Simulator Phase 1)
	Personality    *PersonalityConfig  `json:"personality,omitempty"`
	RomanceDialogs []DialogExtended    `json:"romanceDialogs,omitempty"`
	RomanceEvents  []RandomEventConfig `json:"romanceEvents,omitempty"`
	// Advanced dialog system (Phase 1)
	DialogBackend *dialog.DialogBackendConfig `json:"dialogBackend,omitempty"`
	// General dialog events system (Phase 4)
	GeneralEvents []GeneralDialogEvent `json:"generalEvents,omitempty"`
	// Gift system (optional feature - maintains backward compatibility)
	GiftSystem *GiftSystemConfig `json:"giftSystem,omitempty"`
	// Multiplayer networking (Phase 1 - Networking Infrastructure)
	Multiplayer *MultiplayerConfig `json:"multiplayer,omitempty"`
	// Battle system (Phase 3 - Animation & UI Integration)
	BattleSystem *BattleSystemConfig `json:"battleSystem,omitempty"`
	// News feature extensions (RSS/Atom integration)
	NewsFeatures *news.NewsConfig `json:"newsFeatures,omitempty"`
	// Platform-specific configuration (Phase 5.1 - JSON Schema Extensions)
	PlatformConfig *PlatformConfig `json:"platformConfig,omitempty"`
}

// Dialog represents an interaction trigger and response configuration
type Dialog struct {
	Trigger   string   `json:"trigger"`   // "click", "rightclick", "hover"
	Responses []string `json:"responses"` // 1-10 response strings
	Animation string   `json:"animation"` // Must match an animation key
	Cooldown  int      `json:"cooldown"`  // Seconds between triggers (default: 5)
}

// Behavior defines character behavior settings
type Behavior struct {
	IdleTimeout              int                 `json:"idleTimeout"`                        // Seconds before returning to idle
	MovementEnabled          bool                `json:"movementEnabled"`                    // Allow dragging
	DefaultSize              int                 `json:"defaultSize"`                        // Character size in pixels
	MoodAnimationPreferences map[string][]string `json:"moodAnimationPreferences,omitempty"` // Mood-based animation preferences
}

// GameRulesConfig defines game-wide settings for Tamagotchi-style features
type GameRulesConfig struct {
	StatsDecayInterval             int  `json:"statsDecayInterval"`             // Seconds between stat degradation
	AutoSaveInterval               int  `json:"autoSaveInterval"`               // Seconds between auto-saves
	CriticalStateAnimationPriority bool `json:"criticalStateAnimationPriority"` // Priority for critical animations
	DeathEnabled                   bool `json:"deathEnabled"`                   // Whether character can die
	EvolutionEnabled               bool `json:"evolutionEnabled"`               // Whether character evolves
	MoodBasedAnimations            bool `json:"moodBasedAnimations"`            // Use mood for animation selection
}

// InteractionConfig defines a game interaction (feed, play, etc.)
type InteractionConfig struct {
	Triggers     []string                      `json:"triggers"`     // Input triggers (rightclick, doubleclick, etc.)
	Effects      map[string]float64            `json:"effects"`      // Stat changes to apply
	Animations   []string                      `json:"animations"`   // Animations to play
	Responses    []string                      `json:"responses"`    // Dialog responses
	Cooldown     int                           `json:"cooldown"`     // Seconds between uses
	Duration     int                           `json:"duration"`     // Duration of effect (for sleep, etc.)
	Requirements map[string]map[string]float64 `json:"requirements"` // Stat requirements to use interaction
}

// RandomEventConfig defines a random event that can affect character stats
// Events are triggered based on probability and conditions, following "lazy programmer" approach
type RandomEventConfig struct {
	Name        string                        `json:"name"`        // Event name for identification
	Description string                        `json:"description"` // Human-readable description
	Probability float64                       `json:"probability"` // 0.0-1.0 chance of triggering per check
	Effects     map[string]float64            `json:"effects"`     // Stat changes to apply when triggered
	Animations  []string                      `json:"animations"`  // Animations to play when triggered
	Responses   []string                      `json:"responses"`   // Dialog responses to show
	Cooldown    int                           `json:"cooldown"`    // Minimum seconds between triggers
	Duration    int                           `json:"duration"`    // Duration in seconds (0 = instant)
	Conditions  map[string]map[string]float64 `json:"conditions"`  // Stat conditions required to trigger
}

// Romance-specific configuration structures (Dating Simulator Phase 1)
// PersonalityConfig defines character personality traits that affect romance interactions
type PersonalityConfig struct {
	Traits        map[string]float64 `json:"traits"`        // Personality traits (0.0-1.0 values)
	Compatibility map[string]float64 `json:"compatibility"` // Behavior compatibility modifiers
}

// RomanceRequirement defines complex requirements for romance features
type RomanceRequirement struct {
	Stats               map[string]map[string]float64 `json:"stats,omitempty"`               // Stat-based requirements
	RelationshipLevel   string                        `json:"relationshipLevel,omitempty"`   // Required relationship level
	InteractionCount    map[string]map[string]int     `json:"interactionCount,omitempty"`    // Interaction count requirements
	AchievementUnlocked []string                      `json:"achievementUnlocked,omitempty"` // Required achievements
}

// DialogExtended extends the basic Dialog with romance-specific features
type DialogExtended struct {
	Dialog                           // Embed existing Dialog struct
	Requirements *RomanceRequirement `json:"requirements,omitempty"` // Romance-specific requirements
	RomanceLevel string              `json:"romanceLevel,omitempty"` // Associated romance level
}

// InteractionConfigExtended extends basic InteractionConfig with romance features
type InteractionConfigExtended struct {
	InteractionConfig                      // Embed existing InteractionConfig struct
	UnlockRequirements *RomanceRequirement `json:"unlockRequirements,omitempty"` // Romance unlock requirements
	RomanceCategory    string              `json:"romanceCategory,omitempty"`    // Romance interaction category
}

// MultiplayerConfig defines multiplayer networking configuration for character cards
// Enables peer-to-peer networking features while maintaining backward compatibility
type MultiplayerConfig struct {
	Enabled        bool                      `json:"enabled"`                  // Enable multiplayer networking features
	BotCapable     bool                      `json:"botCapable"`               // Can this character run autonomously as a bot
	NetworkID      string                    `json:"networkID"`                // Unique identifier for this character type
	MaxPeers       int                       `json:"maxPeers,omitempty"`       // Maximum number of peers to connect to (default: 8)
	DiscoveryPort  int                       `json:"discoveryPort,omitempty"`  // UDP port for peer discovery (default: 8080)
	BotPersonality *bot.PersonalityArchetype `json:"botPersonality,omitempty"` // Personality configuration for bot behavior
}

// BattleSystemConfig configures JRPG-style battle features for a character
// This enables turn-based combat with animation integration
type BattleSystemConfig struct {
	Enabled           bool                  `json:"enabled"`                     // Enable battle system features
	BattleStats       map[string]BattleStat `json:"battleStats,omitempty"`       // HP, Attack, Defense, Speed stats
	AIDifficulty      string                `json:"aiDifficulty,omitempty"`      // "easy", "normal", "hard" for bot opponents
	PreferredActions  []string              `json:"preferredActions,omitempty"`  // AI preferred action types
	RequireAnimations bool                  `json:"requireAnimations,omitempty"` // Require battle animations for validation
}

// BattleStat represents a battle-specific stat with base and max values
type BattleStat struct {
	Base float64 `json:"base"` // Base value for the stat
	Max  float64 `json:"max"`  // Maximum value for the stat
}

// PlatformConfig enables platform-specific behavior customization for cross-platform compatibility.
// This provides adaptive configuration for desktop vs mobile environments while maintaining
// backward compatibility with existing character cards.
type PlatformConfig struct {
	Desktop *PlatformSpecificConfig `json:"desktop,omitempty"` // Desktop-specific configuration
	Mobile  *PlatformSpecificConfig `json:"mobile,omitempty"`  // Mobile-specific configuration
}

// PlatformSpecificConfig defines platform-specific behavior and interaction overrides.
// Follows the JSON-first approach for maximum configurability without code changes.
type PlatformSpecificConfig struct {
	// Behavior overrides for platform-specific customization
	Behavior *Behavior `json:"behavior,omitempty"`

	// Platform-specific interaction overrides
	Interactions map[string]PlatformInteractionConfig `json:"interactions,omitempty"`

	// Mobile-specific control configuration
	MobileControls *MobileControlsConfig `json:"mobileControls,omitempty"`

	// Window and display configuration
	WindowMode     string `json:"windowMode,omitempty"`     // "overlay", "fullscreen", "pip" (picture-in-picture)
	DefaultSize    int    `json:"defaultSize,omitempty"`    // Platform-specific default size override
	TouchOptimized bool   `json:"touchOptimized,omitempty"` // Enable touch-optimized UI elements
}

// PlatformInteractionConfig extends InteractionConfig with platform-specific features.
// Enables different interaction patterns for desktop (mouse) vs mobile (touch) environments.
type PlatformInteractionConfig struct {
	InteractionConfig          // Embed base interaction config
	Triggers          []string `json:"triggers,omitempty"`       // Platform-specific triggers ("tap", "longpress", etc.)
	HapticPattern     string   `json:"hapticPattern,omitempty"`  // Haptic feedback pattern ("light", "medium", "heavy")
	TouchFeedback     bool     `json:"touchFeedback,omitempty"`  // Enable visual touch feedback
	GestureEnabled    bool     `json:"gestureEnabled,omitempty"` // Enable gesture-based interactions
}

// MobileControlsConfig defines mobile-specific UI control settings.
// Provides configuration for touch-friendly interface elements and navigation patterns.
type MobileControlsConfig struct {
	ShowBottomBar        bool   `json:"showBottomBar,omitempty"`        // Show bottom control bar
	SwipeGesturesEnabled bool   `json:"swipeGesturesEnabled,omitempty"` // Enable swipe gesture navigation
	HapticFeedback       bool   `json:"hapticFeedback,omitempty"`       // Global haptic feedback setting
	LargeButtons         bool   `json:"largeButtons,omitempty"`         // Use larger, touch-friendly buttons
	ContextMenuStyle     string `json:"contextMenuStyle,omitempty"`     // "bottomsheet", "popup", "fullscreen"
}

// LoadCard loads and validates a character card from a JSON file
// Uses standard library encoding/json - no external dependencies needed
func LoadCard(path string) (*CharacterCard, error) {
	// Resolve path if it's the default relative path (Finding #22 fix)
	resolvedPath := path
	if path == "assets/characters/default/character.json" && !filepath.IsAbs(path) {
		// Find project root by looking for go.mod
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}

		projectRoot := wd
		for {
			if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); err == nil {
				// Found go.mod, use this as project root
				resolvedPath = filepath.Join(projectRoot, path)
				break
			}
			parent := filepath.Dir(projectRoot)
			if parent == projectRoot {
				// Reached filesystem root without finding go.mod, use path as-is
				break
			}
			projectRoot = parent
		}
	}

	data, err := os.ReadFile(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read character card %s: %w", resolvedPath, err)
	}

	var card CharacterCard
	if err := json.Unmarshal(data, &card); err != nil {
		return nil, fmt.Errorf("failed to parse character card %s: %w", resolvedPath, err)
	}

	// Get character directory for animation file validation
	characterDir := filepath.Dir(resolvedPath)
	if err := card.ValidateWithBasePath(characterDir); err != nil {
		return nil, fmt.Errorf("invalid character card %s: %w", resolvedPath, err)
	}

	return &card, nil
}

// Validate ensures the character card has valid configuration
// Implements comprehensive validation to prevent runtime errors
// Validate ensures the character card has valid configuration
func (c *CharacterCard) Validate() error {
	if err := c.validateCoreElements(); err != nil {
		return err
	}

	if err := c.validateGameSystems(); err != nil {
		return err
	}

	if err := c.validateRomanceSystems(); err != nil {
		return err
	}

	if err := c.validateMultiplayerSystems(); err != nil {
		return err
	}

	if err := c.validateBattleSystems(); err != nil {
		return err
	}

	if err := c.validatePlatformSystems(); err != nil {
		return err
	}

	return nil
}

// validateCoreElements validates basic character configuration and assets
func (c *CharacterCard) validateCoreElements() error {
	if err := c.validateBasicFields(); err != nil {
		return err
	}

	if err := c.validateAnimations(); err != nil {
		return err
	}

	if err := c.validateDialogs(); err != nil {
		return err
	}

	if err := c.Behavior.Validate(); err != nil {
		return fmt.Errorf("behavior: %w", err)
	}

	return nil
}

// validateGameSystems validates game mechanics and progression systems
func (c *CharacterCard) validateGameSystems() error {
	if err := c.validateGameFeatures(); err != nil {
		return fmt.Errorf("game features: %w", err)
	}

	if err := c.validateProgression(); err != nil {
		return fmt.Errorf("progression: %w", err)
	}

	if err := c.validateRandomEvents(); err != nil {
		return fmt.Errorf("random events: %w", err)
	}

	return nil
}

// validateRomanceSystems validates romance features and dialog backend
func (c *CharacterCard) validateRomanceSystems() error {
	if err := c.validateRomanceFeatures(); err != nil {
		return fmt.Errorf("romance features: %w", err)
	}

	if err := c.validateDialogBackend(); err != nil {
		return fmt.Errorf("dialog backend: %w", err)
	}

	if err := c.validateGeneralEvents(); err != nil {
		return fmt.Errorf("general events: %w", err)
	}

	return nil
}

// validateMultiplayerSystems validates multiplayer networking configuration
func (c *CharacterCard) validateMultiplayerSystems() error {
	if err := c.validateMultiplayerConfig(); err != nil {
		return fmt.Errorf("multiplayer config: %w", err)
	}

	return nil
}

// validateBattleSystems validates battle system configuration and animations
func (c *CharacterCard) validateBattleSystems() error {
	if err := c.validateBattleConfig(); err != nil {
		return fmt.Errorf("battle config: %w", err)
	}

	return nil
}

// validatePlatformSystems validates platform-specific configuration for cross-platform compatibility.
// Ensures platform configs are consistent and don't conflict with base configuration.
func (c *CharacterCard) validatePlatformSystems() error {
	if err := ValidatePlatformConfig(c); err != nil {
		return fmt.Errorf("platform config: %w", err)
	}

	return nil
}

// ValidateWithBasePath ensures the character card has valid configuration including file existence checks
func (c *CharacterCard) ValidateWithBasePath(basePath string) error {
	if err := c.validateCoreFields(basePath); err != nil {
		return err
	}

	if err := c.validateFeatureSections(); err != nil {
		return err
	}

	// Additional validation for battle animations with file path checks
	if err := c.validateBattleSystemWithBasePath(basePath); err != nil {
		return fmt.Errorf("battle system: %w", err)
	}

	// Validate platform-specific configurations
	if err := c.validatePlatformSystems(); err != nil {
		return err
	}

	return nil
}

// validateCoreFields validates essential card fields and animations with file system checks
func (c *CharacterCard) validateCoreFields(basePath string) error {
	if err := c.validateBasicFields(); err != nil {
		return err
	}

	if err := c.validateAnimationsWithBasePath(basePath); err != nil {
		return err
	}

	if err := c.validateDialogs(); err != nil {
		return err
	}

	if err := c.Behavior.Validate(); err != nil {
		return fmt.Errorf("behavior: %w", err)
	}

	return nil
}

// validateFeatureSections validates all optional feature configurations
func (c *CharacterCard) validateFeatureSections() error {
	if err := c.validateGameFeatures(); err != nil {
		return fmt.Errorf("game features: %w", err)
	}

	if err := c.validateProgression(); err != nil {
		return fmt.Errorf("progression: %w", err)
	}

	if err := c.validateRandomEvents(); err != nil {
		return fmt.Errorf("random events: %w", err)
	}

	if err := c.validateRomanceFeatures(); err != nil {
		return fmt.Errorf("romance features: %w", err)
	}

	if err := c.validateDialogBackend(); err != nil {
		return fmt.Errorf("dialog backend: %w", err)
	}

	if err := c.validateGiftSystem(); err != nil {
		return fmt.Errorf("gift system: %w", err)
	}

	return nil
}

// validateBasicFields checks name and description field constraints
func (c *CharacterCard) validateBasicFields() error {
	if len(c.Name) == 0 || len(c.Name) > 50 {
		return fmt.Errorf("name must be 1-50 characters, got %d", len(c.Name))
	}

	if len(c.Description) == 0 || len(c.Description) > 200 {
		return fmt.Errorf("description must be 1-200 characters, got %d", len(c.Description))
	}

	return nil
}

// validateAnimations ensures required animations exist and have valid file paths
func (c *CharacterCard) validateAnimations() error {
	if c.Animations == nil {
		return fmt.Errorf("animations map is required")
	}

	if err := c.validateRequiredAnimations(); err != nil {
		return err
	}

	return c.validateAnimationPaths()
}

// validateAnimationsWithBasePath ensures required animations exist and all animation files are accessible
func (c *CharacterCard) validateAnimationsWithBasePath(basePath string) error {
	if c.Animations == nil {
		return fmt.Errorf("animations map is required")
	}

	if err := c.validateRequiredAnimations(); err != nil {
		return err
	}

	return c.validateAnimationPathsWithBasePath(basePath)
}

// validateRequiredAnimations checks that mandatory animation keys are present
func (c *CharacterCard) validateRequiredAnimations() error {
	requiredAnimations := []string{"idle", "talking"}
	for _, required := range requiredAnimations {
		if _, exists := c.Animations[required]; !exists {
			return fmt.Errorf("required animation '%s' not found", required)
		}
	}
	return nil
}

// validateAnimationPaths ensures all animation files have proper GIF extensions
func (c *CharacterCard) validateAnimationPaths() error {
	for name, path := range c.Animations {
		if !strings.HasSuffix(strings.ToLower(path), ".gif") {
			return fmt.Errorf("animation '%s' must be a GIF file, got: %s", name, path)
		}
	}
	return nil
}

// validateAnimationPathsWithBasePath ensures all animation files exist and are accessible
func (c *CharacterCard) validateAnimationPathsWithBasePath(basePath string) error {
	for name, path := range c.Animations {
		if !strings.HasSuffix(strings.ToLower(path), ".gif") {
			return fmt.Errorf("animation '%s' must be a GIF file, got: %s", name, path)
		}

		// Check if the animation file actually exists and is readable
		fullPath := filepath.Join(basePath, path)
		if _, err := os.Stat(fullPath); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("animation file '%s' not found: %s", name, fullPath)
			}
			return fmt.Errorf("animation file '%s' not accessible: %s (%v)", name, fullPath, err)
		}
	}
	return nil
}

// validateDialogs ensures dialog configurations are valid and reference existing animations
func (c *CharacterCard) validateDialogs() error {
	if len(c.Dialogs) == 0 {
		return fmt.Errorf("at least one dialog configuration is required")
	}

	for i, dialog := range c.Dialogs {
		if err := dialog.Validate(c.Animations); err != nil {
			return fmt.Errorf("dialog %d: %w", i, err)
		}
	}

	return nil
}

// Validate ensures dialog configuration is valid
func (d *Dialog) Validate(animations map[string]string) error {
	if err := d.validateTrigger(); err != nil {
		return err
	}

	if err := d.validateResponses(); err != nil {
		return err
	}

	if err := d.validateAnimationReference(animations); err != nil {
		return err
	}

	d.setDefaultCooldown()
	return nil
}

// validateTrigger checks if the trigger type is valid
func (d *Dialog) validateTrigger() error {
	validTriggers := []string{"click", "rightclick", "hover"}
	for _, trigger := range validTriggers {
		if d.Trigger == trigger {
			return nil
		}
	}
	return fmt.Errorf("trigger must be one of %v, got: %s", validTriggers, d.Trigger)
}

// validateResponses ensures responses are within limits and not empty
func (d *Dialog) validateResponses() error {
	if len(d.Responses) == 0 || len(d.Responses) > 10 {
		return fmt.Errorf("must have 1-10 responses, got %d", len(d.Responses))
	}

	for i, response := range d.Responses {
		if len(strings.TrimSpace(response)) == 0 {
			return fmt.Errorf("response %d cannot be empty", i)
		}
	}

	return nil
}

// validateAnimationReference checks if the referenced animation exists
func (d *Dialog) validateAnimationReference(animations map[string]string) error {
	if d.Animation == "" {
		return fmt.Errorf("animation field cannot be empty")
	}
	if _, exists := animations[d.Animation]; !exists {
		return fmt.Errorf("animation '%s' not found in animations map", d.Animation)
	}
	return nil
}

// setDefaultCooldown applies default cooldown value if not specified
func (d *Dialog) setDefaultCooldown() {
	if d.Cooldown <= 0 {
		d.Cooldown = 5 // 5 second default
	}
}

// Validate ensures behavior settings are within acceptable ranges
func (b *Behavior) Validate() error {
	if b.IdleTimeout < 10 || b.IdleTimeout > 300 {
		return fmt.Errorf("idleTimeout must be 10-300 seconds, got %d", b.IdleTimeout)
	}

	if b.DefaultSize < 64 || b.DefaultSize > 512 {
		return fmt.Errorf("defaultSize must be 64-512 pixels, got %d", b.DefaultSize)
	}

	return nil
}

// GetAnimationPath returns the full path to an animation file
// Resolves relative paths from the character card directory
func (c *CharacterCard) GetAnimationPath(basePath, animationName string) (string, error) {
	animationFile, exists := c.Animations[animationName]
	if !exists {
		return "", fmt.Errorf("animation '%s' not found", animationName)
	}

	fullPath := filepath.Join(basePath, animationFile)

	// Verify file exists
	if _, err := os.Stat(fullPath); err != nil {
		return "", fmt.Errorf("animation file not found: %s", fullPath)
	}

	return fullPath, nil
}

// validateGameFeatures validates game-specific configuration fields
// This ensures game stats, rules, and interactions are properly configured
func (c *CharacterCard) validateGameFeatures() error {
	if err := c.validateStatsConfig(); err != nil {
		return err
	}

	if err := c.validateGameRulesConfig(); err != nil {
		return err
	}

	if err := c.validateInteractionsConfig(); err != nil {
		return err
	}

	return nil
}

// validateStatsConfig validates all character stats configuration
func (c *CharacterCard) validateStatsConfig() error {
	if c.Stats == nil {
		return nil
	}

	for name, stat := range c.Stats {
		if err := c.validateStatConfig(name, stat); err != nil {
			return fmt.Errorf("stat '%s': %w", name, err)
		}
	}
	return nil
}

// validateGameRulesConfig validates game rules configuration
func (c *CharacterCard) validateGameRulesConfig() error {
	if c.GameRules == nil {
		return nil
	}

	if err := c.validateGameRules(); err != nil {
		return fmt.Errorf("game rules: %w", err)
	}
	return nil
}

// validateInteractionsConfig validates all interaction configurations
func (c *CharacterCard) validateInteractionsConfig() error {
	if c.Interactions == nil {
		return nil
	}

	for name, interaction := range c.Interactions {
		if err := c.validateInteractionConfig(name, interaction); err != nil {
			return fmt.Errorf("interaction '%s': %w", name, err)
		}
	}
	return nil
}

// validateStatConfig ensures a stat configuration is valid
func (c *CharacterCard) validateStatConfig(name string, stat StatConfig) error {
	if stat.Max <= 0 {
		return fmt.Errorf("max value must be positive, got %f", stat.Max)
	}

	if stat.Initial < 0 || stat.Initial > stat.Max {
		return fmt.Errorf("initial value (%f) must be between 0 and max (%f)", stat.Initial, stat.Max)
	}

	if stat.DegradationRate < 0 {
		return fmt.Errorf("degradation rate cannot be negative, got %f", stat.DegradationRate)
	}

	if stat.CriticalThreshold < 0 || stat.CriticalThreshold > stat.Max {
		return fmt.Errorf("critical threshold (%f) must be between 0 and max (%f)", stat.CriticalThreshold, stat.Max)
	}

	return nil
}

// validateGameRules ensures game rules configuration is valid
func (c *CharacterCard) validateGameRules() error {
	if c.GameRules.StatsDecayInterval < 10 || c.GameRules.StatsDecayInterval > 3600 {
		return fmt.Errorf("stats decay interval must be 10-3600 seconds, got %d", c.GameRules.StatsDecayInterval)
	}

	if c.GameRules.AutoSaveInterval < 60 || c.GameRules.AutoSaveInterval > 7200 {
		return fmt.Errorf("auto save interval must be 60-7200 seconds, got %d", c.GameRules.AutoSaveInterval)
	}

	return nil
}

// validateInteractionConfig ensures an interaction configuration is valid
func (c *CharacterCard) validateInteractionConfig(name string, interaction InteractionConfig) error {
	if err := c.validateInteractionTriggers(interaction.Triggers); err != nil {
		return err
	}

	if err := c.validateInteractionAnimations(interaction.Animations); err != nil {
		return err
	}

	if err := c.validateInteractionResponses(interaction.Responses); err != nil {
		return err
	}

	if err := c.validateInteractionCooldown(interaction.Cooldown, interaction.Triggers); err != nil {
		return err
	}

	return nil
}

// validateInteractionTriggers validates that interaction triggers are valid and non-empty
func (c *CharacterCard) validateInteractionTriggers(triggers []string) error {
	if len(triggers) == 0 {
		return fmt.Errorf("must have at least one trigger")
	}

	validTriggers := []string{
		"click", "rightclick", "doubleclick", "shift+click", "hover",
		"ctrl+shift+click", "alt+shift+click", "daily_interaction_bonus",
	}

	for _, trigger := range triggers {
		if !c.isValidTrigger(trigger, validTriggers) {
			return fmt.Errorf("invalid trigger '%s', must be one of %v", trigger, validTriggers)
		}
	}

	return nil
}

// validateInteractionAnimations validates that all referenced animations exist in the animations map
func (c *CharacterCard) validateInteractionAnimations(animations []string) error {
	for _, animation := range animations {
		if _, exists := c.Animations[animation]; !exists {
			return fmt.Errorf("animation '%s' not found in animations map", animation)
		}
	}
	return nil
}

// validateInteractionResponses validates that responses count is within acceptable limits
func (c *CharacterCard) validateInteractionResponses(responses []string) error {
	if len(responses) == 0 || len(responses) > 10 {
		return fmt.Errorf("must have 1-10 responses, got %d", len(responses))
	}
	return nil
}

// validateInteractionCooldown validates cooldown duration based on trigger types
func (c *CharacterCard) validateInteractionCooldown(cooldown int, triggers []string) error {
	maxCooldown := c.calculateMaxCooldown(triggers)

	if cooldown < 0 || cooldown > maxCooldown {
		return fmt.Errorf("cooldown must be 0-%d seconds, got %d", maxCooldown, cooldown)
	}

	return nil
}

// calculateMaxCooldown determines the maximum allowed cooldown based on trigger types
func (c *CharacterCard) calculateMaxCooldown(triggers []string) int {
	maxCooldown := 3600 // Default 1 hour for most interactions

	for _, trigger := range triggers {
		if trigger == "daily_interaction_bonus" {
			return 86400 // Allow 24 hours for daily interactions
		}
	}

	return maxCooldown
}

// isValidTrigger checks if a trigger is in the valid list
func (c *CharacterCard) isValidTrigger(trigger string, validTriggers []string) bool {
	for _, valid := range validTriggers {
		if trigger == valid {
			return true
		}
	}
	return false
}

// validateProgression validates progression configuration (levels and achievements)
func (c *CharacterCard) validateProgression() error {
	if c.Progression == nil {
		return nil // Progression is optional
	}

	// Validate levels
	if len(c.Progression.Levels) == 0 {
		return fmt.Errorf("must have at least one level")
	}

	for i, level := range c.Progression.Levels {
		if err := c.validateProgressionLevel(level, i); err != nil {
			return fmt.Errorf("level %d (%s): %w", i, level.Name, err)
		}
	}

	// Validate achievements
	for i, achievement := range c.Progression.Achievements {
		if err := c.validateProgressionAchievement(achievement, i); err != nil {
			return fmt.Errorf("achievement %d (%s): %w", i, achievement.Name, err)
		}
	}

	return nil
}

// validateProgressionLevel validates a single level configuration
func (c *CharacterCard) validateProgressionLevel(level LevelConfig, index int) error {
	if len(level.Name) == 0 {
		return fmt.Errorf("name cannot be empty")
	}

	if level.Size < 32 || level.Size > 1024 {
		return fmt.Errorf("size must be 32-1024 pixels, got %d", level.Size)
	}

	// First level should have age requirement of 0
	if index == 0 {
		if ageReq, hasAge := level.Requirement["age"]; hasAge && ageReq != 0 {
			return fmt.Errorf("first level must have age requirement of 0, got %d", ageReq)
		}
	}

	// Validate level-specific animations exist in main animations map
	for animName := range level.Animations {
		if _, exists := c.Animations[animName]; !exists {
			return fmt.Errorf("level animation '%s' not found in main animations map", animName)
		}
	}

	return nil
}

// validateProgressionAchievement validates a single achievement configuration
func (c *CharacterCard) validateProgressionAchievement(achievement AchievementConfig, index int) error {
	if len(achievement.Name) == 0 {
		return fmt.Errorf("name cannot be empty")
	}

	if len(achievement.Requirement) == 0 {
		return fmt.Errorf("must have at least one requirement")
	}

	// Validate that required stats exist in character stats
	if c.Stats != nil {
		for statName := range achievement.Requirement {
			if statName == "maintainAbove" {
				continue // Special requirement type
			}
			if _, exists := c.Stats[statName]; !exists {
				return fmt.Errorf("achievement requires stat '%s' which is not defined", statName)
			}
		}
	}

	return nil
}

// HasGameFeatures returns true if this character card includes game features
func (c *CharacterCard) HasGameFeatures() bool {
	return len(c.Stats) > 0
}

// GetRandomResponse returns a random response from a dialog's response list
// Uses time-based seeding for pseudo-randomness - good enough for desktop pets
func (d *Dialog) GetRandomResponse() string {
	if len(d.Responses) == 0 {
		return ""
	}

	// Simple time-based pseudo-random selection
	// For a desktop pet, this provides sufficient randomness
	index := int(time.Now().UnixNano()) % len(d.Responses)
	return d.Responses[index]
}

// CanTrigger checks if enough time has passed since the last trigger
func (d *Dialog) CanTrigger(lastTriggerTime time.Time) bool {
	return time.Since(lastTriggerTime) >= time.Duration(d.Cooldown)*time.Second
}

// validateRandomEvents validates the random events configuration
func (c *CharacterCard) validateRandomEvents() error {
	if len(c.RandomEvents) == 0 {
		return nil // Random events are optional
	}

	for i, event := range c.RandomEvents {
		if err := c.validateRandomEventConfig(event, i); err != nil {
			return fmt.Errorf("event %d (%s): %w", i, event.Name, err)
		}
	}

	return nil
}

// validateRandomEventConfig validates a single random event configuration
func (c *CharacterCard) validateRandomEventConfig(event RandomEventConfig, index int) error {
	if err := c.validateEventBasicFields(event); err != nil {
		return err
	}

	if err := c.validateEventNumericRanges(event); err != nil {
		return err
	}

	if err := c.validateEventAnimations(event); err != nil {
		return err
	}

	if err := c.validateEventResponses(event); err != nil {
		return err
	}

	if err := c.validateEventStatReferences(event); err != nil {
		return err
	}

	return nil
}

// validateEventBasicFields validates required string fields are not empty
func (c *CharacterCard) validateEventBasicFields(event RandomEventConfig) error {
	if len(event.Name) == 0 {
		return fmt.Errorf("name cannot be empty")
	}

	if len(event.Description) == 0 {
		return fmt.Errorf("description cannot be empty")
	}

	return nil
}

// validateEventNumericRanges validates numeric fields are within acceptable ranges
func (c *CharacterCard) validateEventNumericRanges(event RandomEventConfig) error {
	if event.Probability < 0.0 || event.Probability > 1.0 {
		return fmt.Errorf("probability must be 0.0-1.0, got %f", event.Probability)
	}

	if event.Cooldown < 0 || event.Cooldown > 86400 {
		return fmt.Errorf("cooldown must be 0-86400 seconds, got %d", event.Cooldown)
	}

	if event.Duration < 0 || event.Duration > 3600 {
		return fmt.Errorf("duration must be 0-3600 seconds, got %d", event.Duration)
	}

	return nil
}

// validateEventAnimations validates that all referenced animations exist
func (c *CharacterCard) validateEventAnimations(event RandomEventConfig) error {
	for _, animation := range event.Animations {
		if _, exists := c.Animations[animation]; !exists {
			return fmt.Errorf("animation '%s' not found in animations map", animation)
		}
	}

	return nil
}

// validateEventResponses validates response count is within limits
func (c *CharacterCard) validateEventResponses(event RandomEventConfig) error {
	if len(event.Responses) > 10 {
		return fmt.Errorf("must have 0-10 responses, got %d", len(event.Responses))
	}

	return nil
}

// validateEventStatReferences validates that referenced stats exist when character has stats defined
func (c *CharacterCard) validateEventStatReferences(event RandomEventConfig) error {
	// If no stats are defined, stat references are allowed but will be ignored at runtime
	if len(c.Stats) == 0 {
		return nil
	}

	for statName := range event.Effects {
		if _, exists := c.Stats[statName]; !exists {
			return fmt.Errorf("event effects reference stat '%s' which is not defined", statName)
		}
	}

	for statName := range event.Conditions {
		// Allow special romance condition types
		if c.isSpecialRomanceCondition(statName) {
			continue
		}

		if _, exists := c.Stats[statName]; !exists {
			return fmt.Errorf("event conditions reference stat '%s' which is not defined", statName)
		}
	}

	return nil
}

// isSpecialRomanceCondition checks if a condition name is a special romance condition type
func (c *CharacterCard) isSpecialRomanceCondition(conditionName string) bool {
	specialConditions := []string{
		"relationshipLevel",
		"interactionCount",
		"memoryCount",
	}

	for _, special := range specialConditions {
		if conditionName == special {
			return true
		}
	}

	return false
}

// Romance feature validation methods (Dating Simulator Phase 1)

// validateRomanceFeatures validates romance-specific configuration fields
func (c *CharacterCard) validateRomanceFeatures() error {
	if err := c.validatePersonalitySection(); err != nil {
		return err
	}

	if err := c.validateRomanceDialogsConfig(); err != nil {
		return err
	}

	if err := c.validateRomanceEventsConfig(); err != nil {
		return err
	}

	return nil
}

// validatePersonalitySection validates personality configuration if present
func (c *CharacterCard) validatePersonalitySection() error {
	if c.Personality == nil {
		return nil
	}

	if err := c.validatePersonalityConfig(); err != nil {
		return fmt.Errorf("personality: %w", err)
	}
	return nil
}

// validateRomanceDialogsConfig validates all romance dialog configurations
func (c *CharacterCard) validateRomanceDialogsConfig() error {
	if len(c.RomanceDialogs) == 0 {
		return nil
	}

	for i, dialog := range c.RomanceDialogs {
		if err := c.validateRomanceDialog(dialog, i); err != nil {
			return fmt.Errorf("romance dialog %d: %w", i, err)
		}
	}
	return nil
}

// validateRomanceEventsConfig validates all romance event configurations
func (c *CharacterCard) validateRomanceEventsConfig() error {
	if len(c.RomanceEvents) == 0 {
		return nil
	}

	for i, event := range c.RomanceEvents {
		if err := c.validateRandomEventConfig(event, i); err != nil {
			return fmt.Errorf("romance event %d (%s): %w", i, event.Name, err)
		}
	}
	return nil
}

// validatePersonalityConfig ensures personality configuration is valid
func (c *CharacterCard) validatePersonalityConfig() error {
	if c.Personality.Traits != nil {
		for name, value := range c.Personality.Traits {
			if value < 0.0 || value > 1.0 {
				return fmt.Errorf("trait '%s' must be 0.0-1.0, got %f", name, value)
			}
		}
	}

	if c.Personality.Compatibility != nil {
		for name, value := range c.Personality.Compatibility {
			if value < 0.0 || value > 5.0 {
				return fmt.Errorf("compatibility modifier '%s' must be 0.0-5.0, got %f", name, value)
			}
		}
	}

	return nil
}

// validateRomanceDialog validates an extended dialog configuration
func (c *CharacterCard) validateRomanceDialog(dialog DialogExtended, index int) error {
	// Validate base dialog
	if err := dialog.Dialog.Validate(c.Animations); err != nil {
		return err
	}

	// Validate romance requirements if present
	if dialog.Requirements != nil {
		if err := c.validateRomanceRequirements(dialog.Requirements); err != nil {
			return fmt.Errorf("requirements: %w", err)
		}
	}

	return nil
}

// validateRomanceRequirements validates romance requirement configuration
func (c *CharacterCard) validateRomanceRequirements(req *RomanceRequirement) error {
	// Validate stat requirements reference existing stats (if stats are defined)
	if len(c.Stats) > 0 && req.Stats != nil {
		for statName := range req.Stats {
			if _, exists := c.Stats[statName]; !exists {
				return fmt.Errorf("requirement references stat '%s' which is not defined", statName)
			}
		}
	}

	// Validate interaction count requirements
	if req.InteractionCount != nil {
		for interactionName := range req.InteractionCount {
			if c.Interactions != nil {
				if _, exists := c.Interactions[interactionName]; !exists {
					return fmt.Errorf("requirement references interaction '%s' which is not defined", interactionName)
				}
			}
		}
	}

	return nil
}

// validateDialogBackend ensures dialog backend configuration is valid when enabled
func (c *CharacterCard) validateDialogBackend() error {
	if c.DialogBackend == nil {
		return nil // Optional feature, validation not required when absent
	}

	return dialog.ValidateBackendConfig(*c.DialogBackend)
}

// HasRomanceFeatures returns true if this character card includes romance features
func (c *CharacterCard) HasRomanceFeatures() bool {
	return c.Personality != nil || len(c.RomanceDialogs) > 0 || len(c.RomanceEvents) > 0
}

// HasDialogBackend returns true if this character card has dialog backend configuration enabled
func (c *CharacterCard) HasDialogBackend() bool {
	return c.DialogBackend != nil && c.DialogBackend.Enabled
}

// Bug #4 Fix: Additional methods for more granular dialog backend state checking

// HasDialogBackendConfig returns true if this character has dialog backend configuration
// regardless of whether it's enabled or disabled. Useful for determining if a character
// was intended to have AI capabilities.
func (c *CharacterCard) HasDialogBackendConfig() bool {
	return c.DialogBackend != nil
}

// IsDialogBackendEnabled returns true if dialog backend is both configured and enabled
// This is equivalent to HasDialogBackend() but with a more descriptive name.
func (c *CharacterCard) IsDialogBackendEnabled() bool {
	return c.DialogBackend != nil && c.DialogBackend.Enabled
}

// GetDialogBackendStatus returns detailed information about dialog backend availability
// Returns: (hasConfig, isEnabled, reason) where reason explains the current state
func (c *CharacterCard) GetDialogBackendStatus() (bool, bool, string) {
	if c.DialogBackend == nil {
		return false, false, "No dialog backend configured"
	}

	if !c.DialogBackend.Enabled {
		return true, false, "Dialog backend configured but disabled"
	}

	return true, true, "Dialog backend configured and enabled"
}

// GetPersonalityTrait returns the value of a personality trait, defaulting to 0.5 if not found
func (c *CharacterCard) GetPersonalityTrait(trait string) float64 {
	if c.Personality == nil || c.Personality.Traits == nil {
		return 0.5 // Default neutral value
	}
	if value, exists := c.Personality.Traits[trait]; exists {
		return value
	}
	return 0.5
}

// GetCompatibilityModifier returns the compatibility modifier for a behavior, defaulting to 1.0
func (c *CharacterCard) GetCompatibilityModifier(behavior string) float64 {
	if c.Personality == nil || c.Personality.Compatibility == nil {
		return 1.0 // Default no modifier
	}
	if modifier, exists := c.Personality.Compatibility[behavior]; exists {
		return modifier
	}
	return 1.0
}

// validateGeneralEvents validates general dialog events configuration
func (c *CharacterCard) validateGeneralEvents() error {
	if len(c.GeneralEvents) == 0 {
		return nil // No general events to validate
	}

	for i, event := range c.GeneralEvents {
		if err := ValidateGeneralEvent(event); err != nil {
			return fmt.Errorf("event %d (%s): %w", i, event.Name, err)
		}
	}

	return nil
}

// validateGiftSystem validates gift system configuration
// Maintains backward compatibility by treating nil GiftSystem as valid
// validateGiftSystem validates the optional gift system configuration.
func (c *CharacterCard) validateGiftSystem() error {
	if c.GiftSystem == nil {
		return nil // Optional feature, validation not required when absent
	}

	if err := c.validateInventorySettings(); err != nil {
		return err
	}

	if err := c.validateGiftCategories(); err != nil {
		return err
	}

	if err := c.validatePersonalityResponses(); err != nil {
		return err
	}

	return nil
}

// validateInventorySettings checks that inventory configuration is within valid bounds.
func (c *CharacterCard) validateInventorySettings() error {
	maxSlots := c.GiftSystem.InventorySettings.MaxSlots
	if maxSlots < 1 {
		return fmt.Errorf("inventory maxSlots must be at least 1, got %d", maxSlots)
	}
	if maxSlots > 100 {
		return fmt.Errorf("inventory maxSlots cannot exceed 100, got %d", maxSlots)
	}
	return nil
}

// validateGiftCategories ensures all gift categories are from the allowed set.
func (c *CharacterCard) validateGiftCategories() error {
	validCategories := []string{"food", "flowers", "books", "jewelry", "toys", "electronics", "clothing", "art", "practical", "expensive"}

	if err := c.validateCategoryList(c.GiftSystem.Preferences.FavoriteCategories, "favorite", validCategories); err != nil {
		return err
	}

	if err := c.validateCategoryList(c.GiftSystem.Preferences.DislikedCategories, "disliked", validCategories); err != nil {
		return err
	}

	return nil
}

// validateCategoryList validates a category list against valid categories.
func (c *CharacterCard) validateCategoryList(categories []string, categoryType string, validCategories []string) error {
	for _, category := range categories {
		if !sliceContains(validCategories, category) {
			return fmt.Errorf("invalid %s category '%s', must be one of: %v", categoryType, category, validCategories)
		}
	}
	return nil
}

// validatePersonalityResponses ensures personality response configuration is valid.
func (c *CharacterCard) validatePersonalityResponses() error {
	for personality, response := range c.GiftSystem.Preferences.PersonalityResponses {
		if len(response.GiftReceived) == 0 {
			return fmt.Errorf("personality '%s' must have at least one gift received response", personality)
		}
		if len(response.GiftReceived) > 10 {
			return fmt.Errorf("personality '%s' cannot have more than 10 gift received responses, got %d", personality, len(response.GiftReceived))
		}
	}
	return nil
}

// HasGiftSystem returns true if this character card includes gift system configuration
func (c *CharacterCard) HasGiftSystem() bool {
	return c.GiftSystem != nil && c.GiftSystem.Enabled
}

// validateMultiplayerConfig validates multiplayer networking configuration
// Ensures multiplayer settings are valid when enabled
func (c *CharacterCard) validateMultiplayerConfig() error {
	// Skip validation if multiplayer is not configured
	if c.Multiplayer == nil {
		return nil
	}

	mp := c.Multiplayer

	// Validate NetworkID when enabled
	if err := c.validateNetworkID(mp); err != nil {
		return err
	}

	// Validate MaxPeers range
	if err := c.validateMaxPeers(mp.MaxPeers); err != nil {
		return err
	}

	// Validate DiscoveryPort range
	if err := c.validateDiscoveryPort(mp.DiscoveryPort); err != nil {
		return err
	}

	// Validate BotPersonality when bot capabilities are enabled
	if err := c.validateBotPersonality(mp); err != nil {
		return err
	}

	return nil
}

// validateNetworkID validates the network ID configuration when multiplayer is enabled
func (c *CharacterCard) validateNetworkID(mp *MultiplayerConfig) error {
	if !mp.Enabled {
		return nil
	}

	if mp.NetworkID == "" {
		return fmt.Errorf("networkID is required when multiplayer is enabled")
	}
	if len(mp.NetworkID) > 50 {
		return fmt.Errorf("networkID too long: %d characters, maximum 50 allowed", len(mp.NetworkID))
	}

	// NetworkID should contain only alphanumeric, underscore, and dash characters
	for _, char := range mp.NetworkID {
		if !c.isValidNetworkIDChar(char) {
			return fmt.Errorf("networkID contains invalid character '%c', only alphanumeric, underscore, and dash allowed", char)
		}
	}

	return nil
}

// isValidNetworkIDChar checks if a character is valid for network ID
func (c *CharacterCard) isValidNetworkIDChar(char rune) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') || char == '_' || char == '-'
}

// validateMaxPeers validates the maximum number of peers configuration
func (c *CharacterCard) validateMaxPeers(maxPeers int) error {
	if maxPeers > 0 && maxPeers > 16 {
		return fmt.Errorf("maxPeers cannot exceed 16, got %d", maxPeers)
	}
	return nil
}

// validateDiscoveryPort validates the discovery port configuration
func (c *CharacterCard) validateDiscoveryPort(discoveryPort int) error {
	if discoveryPort <= 0 {
		return nil
	}

	if discoveryPort < 1024 {
		return fmt.Errorf("discoveryPort must be >= 1024 to avoid system ports, got %d", discoveryPort)
	}
	if discoveryPort > 65535 {
		return fmt.Errorf("discoveryPort cannot exceed 65535, got %d", discoveryPort)
	}

	return nil
}

// validateBotPersonality validates bot personality configuration when bot capabilities are enabled
func (c *CharacterCard) validateBotPersonality(mp *MultiplayerConfig) error {
	// Skip validation if bot capabilities are not enabled
	if !mp.BotCapable {
		return nil
	}

	// BotPersonality is optional - skip validation if not provided
	if mp.BotPersonality == nil {
		return nil
	}

	// Use PersonalityManager to validate the personality archetype
	pm := bot.NewPersonalityManager()

	// Create a JSON representation to validate through the standard path
	jsonData, err := json.Marshal(mp.BotPersonality)
	if err != nil {
		return fmt.Errorf("failed to marshal bot personality: %w", err)
	}

	// Validate using PersonalityManager
	_, err = pm.LoadFromJSON(jsonData)
	if err != nil {
		return fmt.Errorf("invalid bot personality: %w", err)
	}

	return nil
}

// validateBattleConfig validates battle system configuration
// Ensures battle settings are valid when enabled
func (c *CharacterCard) validateBattleConfig() error {
	// Skip validation if battle system is not configured
	if c.BattleSystem == nil {
		return nil
	}

	bs := c.BattleSystem

	// Validate battle stats when provided
	if err := c.validateBattleStats(bs.BattleStats); err != nil {
		return err
	}

	// Validate AI difficulty setting
	if err := c.validateAIDifficulty(bs.AIDifficulty); err != nil {
		return err
	}

	// Validate preferred actions
	if err := c.validatePreferredActions(bs.PreferredActions); err != nil {
		return err
	}

	// Validate battle animations when required
	if bs.RequireAnimations || bs.Enabled {
		if err := c.validateBattleAnimations(); err != nil {
			return fmt.Errorf("battle animations: %w", err)
		}
	}

	return nil
}

// validateBattleStats validates battle stat configurations
func (c *CharacterCard) validateBattleStats(stats map[string]BattleStat) error {
	if len(stats) == 0 {
		return nil // Optional
	}

	for statName, stat := range stats {
		if stat.Base < 0 {
			return fmt.Errorf("battle stat '%s' base value cannot be negative: %f", statName, stat.Base)
		}
		if stat.Max < stat.Base {
			return fmt.Errorf("battle stat '%s' max value (%f) cannot be less than base value (%f)", statName, stat.Max, stat.Base)
		}
	}
	return nil
}

// validateAIDifficulty validates AI difficulty setting
func (c *CharacterCard) validateAIDifficulty(difficulty string) error {
	if difficulty == "" {
		return nil // Optional
	}

	validDifficulties := []string{"easy", "normal", "hard"}
	for _, valid := range validDifficulties {
		if difficulty == valid {
			return nil
		}
	}

	return fmt.Errorf("invalid AI difficulty '%s', must be one of: %v", difficulty, validDifficulties)
}

// validatePreferredActions validates preferred action configuration
func (c *CharacterCard) validatePreferredActions(actions []string) error {
	if len(actions) == 0 {
		return nil // Optional
	}

	validActions := []string{"attack", "defend", "heal", "stun", "boost", "counter", "drain", "shield", "charge", "evade", "taunt"}
	for _, action := range actions {
		valid := false
		for _, validAction := range validActions {
			if action == validAction {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid preferred action '%s', must be one of: %v", action, validActions)
		}
	}

	return nil
}

// validateBattleAnimations validates that battle animations are present when required
func (c *CharacterCard) validateBattleAnimations() error {
	if c.Animations == nil {
		return fmt.Errorf("animations map is required for battle system")
	}

	// Get list of available battle animations
	battleAnimations := c.getAvailableBattleAnimations()

	// Require at least one battle animation
	if len(battleAnimations) == 0 {
		return fmt.Errorf("at least one battle animation is required (attack, defend, heal, etc.)")
	}

	return nil
}

// getAvailableBattleAnimations returns a list of battle animations present in the character
func (c *CharacterCard) getAvailableBattleAnimations() []string {
	battleAnims := []string{
		AnimationAttack, AnimationDefend, AnimationStun, AnimationHeal,
		AnimationBoost, AnimationCounter, AnimationDrain, AnimationShield,
		AnimationCharge, AnimationEvade, AnimationTaunt, AnimationVictory,
	}

	var available []string
	for _, anim := range battleAnims {
		if _, exists := c.Animations[anim]; exists {
			available = append(available, anim)
		}
	}

	return available
}

// HasBattleSystem returns true if this character card has battle system enabled
func (c *CharacterCard) HasBattleSystem() bool {
	return c.BattleSystem != nil && c.BattleSystem.Enabled
}

// validateBattleSystemWithBasePath validates battle system including animation file existence
func (c *CharacterCard) validateBattleSystemWithBasePath(basePath string) error {
	// Skip validation if battle system is not configured
	if c.BattleSystem == nil {
		return nil
	}

	// Validate battle animations with file existence checks when required
	if c.BattleSystem.RequireAnimations || c.BattleSystem.Enabled {
		if err := c.validateBattleAnimationsWithBasePath(basePath); err != nil {
			return fmt.Errorf("battle animations: %w", err)
		}
	}

	return nil
}

// validateBattleAnimationsWithBasePath validates battle animation files exist and are accessible
func (c *CharacterCard) validateBattleAnimationsWithBasePath(basePath string) error {
	if c.Animations == nil {
		return fmt.Errorf("animations map is required for battle system")
	}

	// Get list of available battle animations
	battleAnimations := c.getAvailableBattleAnimations()

	// Require at least one battle animation
	if len(battleAnimations) == 0 {
		return fmt.Errorf("at least one battle animation is required (attack, defend, heal, etc.)")
	}

	// Check that all battle animation files exist and are accessible
	for _, animName := range battleAnimations {
		animPath := c.Animations[animName]
		if !strings.HasSuffix(strings.ToLower(animPath), ".gif") {
			return fmt.Errorf("battle animation '%s' must be a GIF file, got: %s", animName, animPath)
		}

		fullPath := filepath.Join(basePath, animPath)
		if _, err := os.Stat(fullPath); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("battle animation file '%s' not found: %s", animName, fullPath)
			}
			return fmt.Errorf("battle animation file '%s' not accessible: %s (%v)", animName, fullPath, err)
		}
	}

	return nil
}

// HasMultiplayer returns true if this character card has multiplayer networking enabled
func (c *CharacterCard) HasMultiplayer() bool {
	return c.Multiplayer != nil && c.Multiplayer.Enabled
}

// IsBotCapable returns true if this character can run autonomously as a bot
func (c *CharacterCard) IsBotCapable() bool {
	return c.Multiplayer != nil && c.Multiplayer.BotCapable
}

// GetBotPersonality returns the bot personality for this character, or nil if not configured
func (c *CharacterCard) GetBotPersonality() (*bot.BotPersonality, error) {
	if !c.IsBotCapable() || c.Multiplayer.BotPersonality == nil {
		return nil, nil
	}

	pm := bot.NewPersonalityManager()
	return pm.CreatePersonality(c.Multiplayer.BotPersonality)
}

// HasNewsFeatures returns true if this character has news features enabled
func (c *CharacterCard) HasNewsFeatures() bool {
	return c.NewsFeatures != nil && c.NewsFeatures.Enabled
}

// GetNewsConfig returns the news configuration for this character, or nil if not configured
func (c *CharacterCard) GetNewsConfig() *news.NewsConfig {
	return c.NewsFeatures
}
