package character

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GiftDefinition represents a loadable gift with properties and effects
// Follows the existing JSON configuration pattern used in CharacterCard
type GiftDefinition struct {
	ID                   string                        `json:"id"`
	Name                 string                        `json:"name"`
	Description          string                        `json:"description"`
	Category             string                        `json:"category"`
	Rarity               string                        `json:"rarity"`
	Image                string                        `json:"image"`
	Properties           GiftProperties                `json:"properties"`
	GiftEffects          GiftEffects                   `json:"giftEffects"`
	PersonalityModifiers map[string]map[string]float64 `json:"personalityModifiers"`
	Notes                GiftNotesConfig               `json:"notes"`
}

// GiftProperties defines gift behavior and unlock requirements
type GiftProperties struct {
	Consumable         bool                   `json:"consumable"`
	Stackable          bool                   `json:"stackable"`
	MaxStack           int                    `json:"maxStack"`
	CooldownSeconds    int                    `json:"cooldownSeconds,omitempty"` // Cooldown time in seconds (0 = no cooldown)
	UnlockRequirements map[string]interface{} `json:"unlockRequirements"`
}

// GiftEffects defines immediate and memory effects of giving a gift
type GiftEffects struct {
	Immediate ImmediateEffects `json:"immediate"`
	Memory    MemoryEffects    `json:"memory"`
	Battle    BattleItemEffect `json:"battle,omitempty"` // Battle-specific effects
}

// ImmediateEffects represents stat changes, animations, and responses
type ImmediateEffects struct {
	Stats      map[string]float64 `json:"stats"`
	Animations []string           `json:"animations"`
	Responses  []string           `json:"responses"`
}

// MemoryEffects represents how the gift affects character memory
type MemoryEffects struct {
	Importance    float64  `json:"importance"`
	Tags          []string `json:"tags"`
	EmotionalTone string   `json:"emotionalTone"`
}

// BattleItemEffect defines battle-specific effects when items are used in combat
// All modifiers are capped by the battle system's fairness constraints
type BattleItemEffect struct {
	ActionType      string  `json:"actionType,omitempty"`      // Specific action this enhances ("attack", "heal", etc.)
	DamageModifier  float64 `json:"damageModifier,omitempty"`  // Multiplier for damage (capped at MAX_DAMAGE_MODIFIER)
	DefenseModifier float64 `json:"defenseModifier,omitempty"` // Multiplier for defense (capped at MAX_DEFENSE_MODIFIER)
	SpeedModifier   float64 `json:"speedModifier,omitempty"`   // Multiplier for speed (capped at MAX_SPEED_MODIFIER)
	HealModifier    float64 `json:"healModifier,omitempty"`    // Multiplier for healing (capped at MAX_HEAL_MODIFIER)
	Duration        int     `json:"duration,omitempty"`        // Turns the effect lasts (0 = single use)
	Consumable      bool    `json:"consumable"`                // Whether item is consumed on use
}

// GiftNotesConfig defines note attachment settings
type GiftNotesConfig struct {
	Enabled     bool   `json:"enabled"`
	MaxLength   int    `json:"maxLength"`
	Placeholder string `json:"placeholder"`
}

// GiftSystemConfig represents character-specific gift system settings
// This extends the CharacterCard schema in a backward-compatible way
type GiftSystemConfig struct {
	Enabled           bool              `json:"enabled"`
	Preferences       GiftPreferences   `json:"preferences"`
	InventorySettings InventorySettings `json:"inventorySettings"`
}

// GiftPreferences defines character preferences for gift categories
type GiftPreferences struct {
	FavoriteCategories   []string                       `json:"favoriteCategories"`
	DislikedCategories   []string                       `json:"dislikedCategories"`
	PersonalityResponses map[string]PersonalityResponse `json:"personalityResponses"`
}

// PersonalityResponse defines personality-specific gift responses
type PersonalityResponse struct {
	GiftReceived []string `json:"giftReceived"`
	Animations   []string `json:"animations"`
}

// InventorySettings defines inventory UI and behavior settings
type InventorySettings struct {
	MaxSlots   int  `json:"maxSlots"`
	AutoSort   bool `json:"autoSort"`
	ShowRarity bool `json:"showRarity"`
}

// GiftResponse represents the result of giving a gift
type GiftResponse struct {
	Response      string             `json:"response"`
	Animation     string             `json:"animation"`
	StatEffects   map[string]float64 `json:"statEffects"`
	MemoryCreated bool               `json:"memoryCreated"`
	ErrorMessage  string             `json:"errorMessage,omitempty"`
}

// GiftMemory represents a memory of a gift interaction for tracking and learning
// Extends existing memory system with gift-specific fields
type GiftMemory struct {
	Timestamp        time.Time          `json:"timestamp"`
	GiftID           string             `json:"giftId"`
	GiftName         string             `json:"giftName"`
	Notes            string             `json:"notes"`
	Response         string             `json:"response"`
	StatEffects      map[string]float64 `json:"statEffects"`
	MemoryImportance float64            `json:"memoryImportance"`
	Tags             []string           `json:"tags"`
	EmotionalTone    string             `json:"emotionalTone"`
}

// LoadGiftDefinition loads a gift definition from a JSON file
// Reuses the existing LoadCard pattern for consistency
func LoadGiftDefinition(filePath string) (*GiftDefinition, error) {
	// Validate file path exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("gift definition file not found: %s", filePath)
	}

	// Read file content using standard library
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read gift definition file %s: %w", filePath, err)
	}

	// Parse JSON using standard library
	var gift GiftDefinition
	if err := json.Unmarshal(data, &gift); err != nil {
		return nil, fmt.Errorf("failed to parse gift definition JSON in %s: %w", filePath, err)
	}

	// Validate gift definition using existing validation patterns
	if err := gift.Validate(); err != nil {
		return nil, fmt.Errorf("gift definition validation failed for %s: %w", filePath, err)
	}

	return &gift, nil
}

// Validate validates a gift definition using existing validation patterns
// Follows the same approach as CharacterCard.Validate()
func (g *GiftDefinition) Validate() error {
	if err := g.validateRequiredFields(); err != nil {
		return err
	}

	if err := g.validateCategoryAndRarity(); err != nil {
		return err
	}

	if err := g.validateGiftComponents(); err != nil {
		return err
	}

	return nil
}

// validateRequiredFields checks that all required string fields meet length constraints
func (g *GiftDefinition) validateRequiredFields() error {
	if g.ID == "" {
		return fmt.Errorf("gift ID is required")
	}
	if len(g.ID) > 50 {
		return fmt.Errorf("gift ID must be 50 characters or less, got %d", len(g.ID))
	}
	if g.Name == "" {
		return fmt.Errorf("gift name is required")
	}
	if len(g.Name) > 100 {
		return fmt.Errorf("gift name must be 100 characters or less, got %d", len(g.Name))
	}
	if g.Description == "" {
		return fmt.Errorf("gift description is required")
	}
	if len(g.Description) > 500 {
		return fmt.Errorf("gift description must be 500 characters or less, got %d", len(g.Description))
	}
	return nil
}

// validateCategoryAndRarity checks that category and rarity values are from valid sets
func (g *GiftDefinition) validateCategoryAndRarity() error {
	validCategories := []string{"food", "flowers", "books", "jewelry", "toys", "electronics", "clothing", "art", "practical", "expensive"}
	if !sliceContains(validCategories, g.Category) {
		return fmt.Errorf("invalid gift category '%s', must be one of: %s", g.Category, strings.Join(validCategories, ", "))
	}

	validRarities := []string{"common", "uncommon", "rare", "epic", "legendary"}
	if !sliceContains(validRarities, g.Rarity) {
		return fmt.Errorf("invalid gift rarity '%s', must be one of: %s", g.Rarity, strings.Join(validRarities, ", "))
	}
	return nil
}

// validateGiftComponents validates properties, effects, and notes configurations
func (g *GiftDefinition) validateGiftComponents() error {
	if err := g.validateProperties(); err != nil {
		return fmt.Errorf("gift properties validation failed: %w", err)
	}

	if err := g.validateEffects(); err != nil {
		return fmt.Errorf("gift effects validation failed: %w", err)
	}

	if err := g.validateNotes(); err != nil {
		return fmt.Errorf("gift notes validation failed: %w", err)
	}

	return nil
}

// validateProperties validates gift properties
func (g *GiftDefinition) validateProperties() error {
	if g.Properties.MaxStack < 1 {
		return fmt.Errorf("maxStack must be at least 1, got %d", g.Properties.MaxStack)
	}
	if g.Properties.MaxStack > 999 {
		return fmt.Errorf("maxStack must be 999 or less, got %d", g.Properties.MaxStack)
	}
	if !g.Properties.Stackable && g.Properties.MaxStack > 1 {
		return fmt.Errorf("non-stackable gifts cannot have maxStack > 1")
	}
	return nil
}

// validateEffects validates gift effects
func (g *GiftDefinition) validateEffects() error {
	// Validate stat effects are reasonable
	for statName, value := range g.GiftEffects.Immediate.Stats {
		if value < -100 || value > 100 {
			return fmt.Errorf("stat effect for '%s' must be between -100 and 100, got %f", statName, value)
		}
	}

	// Validate responses exist
	if len(g.GiftEffects.Immediate.Responses) == 0 {
		return fmt.Errorf("at least one response is required")
	}
	if len(g.GiftEffects.Immediate.Responses) > 10 {
		return fmt.Errorf("maximum 10 responses allowed, got %d", len(g.GiftEffects.Immediate.Responses))
	}

	// Validate memory importance
	if g.GiftEffects.Memory.Importance < 0 || g.GiftEffects.Memory.Importance > 1 {
		return fmt.Errorf("memory importance must be between 0 and 1, got %f", g.GiftEffects.Memory.Importance)
	}

	return nil
}

// validateNotes validates notes configuration
func (g *GiftDefinition) validateNotes() error {
	if g.Notes.MaxLength < 0 {
		return fmt.Errorf("notes maxLength cannot be negative, got %d", g.Notes.MaxLength)
	}
	if g.Notes.MaxLength > 1000 {
		return fmt.Errorf("notes maxLength cannot exceed 1000, got %d", g.Notes.MaxLength)
	}
	return nil
}

// LoadGiftCatalog loads all gift definitions from a directory
// Follows the existing pattern used for loading character configurations
func LoadGiftCatalog(giftsPath string) (map[string]*GiftDefinition, error) {
	catalog := make(map[string]*GiftDefinition)

	if !directoryExists(giftsPath) {
		// Return empty catalog if directory doesn't exist (gifts are optional)
		return catalog, nil
	}

	entries, err := readGiftDirectory(giftsPath)
	if err != nil {
		return nil, err
	}

	if err := processGiftFiles(entries, giftsPath, catalog); err != nil {
		return nil, err
	}

	return catalog, nil
}

// directoryExists checks if the gifts directory exists
func directoryExists(giftsPath string) bool {
	_, err := os.Stat(giftsPath)
	return !os.IsNotExist(err)
}

// readGiftDirectory reads the contents of the gifts directory
func readGiftDirectory(giftsPath string) ([]os.DirEntry, error) {
	entries, err := os.ReadDir(giftsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read gifts directory %s: %w", giftsPath, err)
	}
	return entries, nil
}

// processGiftFiles processes each JSON file in the directory and loads gift definitions
func processGiftFiles(entries []os.DirEntry, giftsPath string, catalog map[string]*GiftDefinition) error {
	for _, entry := range entries {
		if !isValidGiftFile(entry) {
			continue
		}

		giftPath := filepath.Join(giftsPath, entry.Name())
		if err := loadAndValidateGift(giftPath, catalog); err != nil {
			return err
		}
	}
	return nil
}

// isValidGiftFile checks if a directory entry is a valid JSON gift file
func isValidGiftFile(entry os.DirEntry) bool {
	return !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json")
}

// loadAndValidateGift loads a gift definition and validates for duplicates
func loadAndValidateGift(giftPath string, catalog map[string]*GiftDefinition) error {
	gift, err := LoadGiftDefinition(giftPath)
	if err != nil {
		return fmt.Errorf("failed to load gift from %s: %w", giftPath, err)
	}

	if err := checkDuplicateGiftID(gift.ID, giftPath, catalog); err != nil {
		return err
	}

	catalog[gift.ID] = gift
	return nil
}

// checkDuplicateGiftID validates that a gift ID is unique in the catalog
func checkDuplicateGiftID(giftID, giftPath string, catalog map[string]*GiftDefinition) error {
	if _, exists := catalog[giftID]; exists {
		return fmt.Errorf("duplicate gift ID '%s' found in %s", giftID, giftPath)
	}
	return nil
}

// contains is a utility function to check if a slice contains a string
func sliceContains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
