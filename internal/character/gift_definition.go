package character

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	UnlockRequirements map[string]interface{} `json:"unlockRequirements"`
}

// GiftEffects defines immediate and memory effects of giving a gift
type GiftEffects struct {
	Immediate ImmediateEffects `json:"immediate"`
	Memory    MemoryEffects    `json:"memory"`
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
	// Required field validation
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

	// Category validation
	validCategories := []string{"food", "flowers", "books", "jewelry", "toys", "electronics", "clothing", "art", "practical", "expensive"}
	if !sliceContains(validCategories, g.Category) {
		return fmt.Errorf("invalid gift category '%s', must be one of: %s", g.Category, strings.Join(validCategories, ", "))
	}

	// Rarity validation
	validRarities := []string{"common", "uncommon", "rare", "epic", "legendary"}
	if !sliceContains(validRarities, g.Rarity) {
		return fmt.Errorf("invalid gift rarity '%s', must be one of: %s", g.Rarity, strings.Join(validRarities, ", "))
	}

	// Properties validation
	if err := g.validateProperties(); err != nil {
		return fmt.Errorf("gift properties validation failed: %w", err)
	}

	// Effects validation
	if err := g.validateEffects(); err != nil {
		return fmt.Errorf("gift effects validation failed: %w", err)
	}

	// Notes validation
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

	// Check if gifts directory exists
	if _, err := os.Stat(giftsPath); os.IsNotExist(err) {
		// Return empty catalog if directory doesn't exist (gifts are optional)
		return catalog, nil
	}

	// Read directory contents
	entries, err := os.ReadDir(giftsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read gifts directory %s: %w", giftsPath, err)
	}

	// Load each JSON file
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		giftPath := filepath.Join(giftsPath, entry.Name())
		gift, err := LoadGiftDefinition(giftPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load gift from %s: %w", giftPath, err)
		}

		// Check for duplicate IDs
		if _, exists := catalog[gift.ID]; exists {
			return nil, fmt.Errorf("duplicate gift ID '%s' found in %s", gift.ID, giftPath)
		}

		catalog[gift.ID] = gift
	}

	return catalog, nil
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
