// Package battle - Equipment system for enhanced tactical battle gameplay
//
// This file implements a battle equipment system that allows characters to
// equip items that provide stat modifications, special effects, and tactical
// advantages during combat. Equipment integrates with the existing gift system
// and special abilities for comprehensive battle customization.
//
// Design principles:
// - Equipment effects respect existing fairness constraints
// - Standard library only implementation
// - Integration with existing gift/item system
// - Configurable equipment slots and restrictions
package battle

import (
	"errors"
	"fmt"
)

// Equipment system errors
var (
	ErrEquipmentSlotFull     = errors.New("equipment slot is already occupied")
	ErrEquipmentNotFound     = errors.New("equipment not found in inventory")
	ErrEquipmentRestricted   = errors.New("equipment cannot be equipped by this character")
	ErrInvalidEquipmentSlot  = errors.New("invalid equipment slot")
	ErrEquipmentInUse        = errors.New("equipment is currently equipped and cannot be removed during battle")
)

// EquipmentSlot defines the type of equipment slot
type EquipmentSlot string

const (
	SLOT_WEAPON     EquipmentSlot = "weapon"     // Primary weapon slot
	SLOT_ARMOR      EquipmentSlot = "armor"      // Body armor slot
	SLOT_ACCESSORY  EquipmentSlot = "accessory"  // Accessory slot (rings, amulets)
	SLOT_CONSUMABLE EquipmentSlot = "consumable" // Single-use items
)

// EquipmentRarity affects the power and availability of equipment
type EquipmentRarity string

const (
	RARITY_COMMON    EquipmentRarity = "common"    // Basic equipment
	RARITY_UNCOMMON  EquipmentRarity = "uncommon"  // Enhanced equipment
	RARITY_RARE      EquipmentRarity = "rare"      // Powerful equipment
	RARITY_EPIC      EquipmentRarity = "epic"      // Very powerful equipment
	RARITY_LEGENDARY EquipmentRarity = "legendary" // Extremely powerful equipment
)

// EquipmentType defines specific equipment categories
type EquipmentType string

const (
	// Weapons
	EQUIPMENT_SWORD     EquipmentType = "sword"
	EQUIPMENT_BOW       EquipmentType = "bow"
	EQUIPMENT_STAFF     EquipmentType = "staff"
	EQUIPMENT_DAGGER    EquipmentType = "dagger"
	EQUIPMENT_HAMMER    EquipmentType = "hammer"
	
	// Armor
	EQUIPMENT_HEAVY_ARMOR  EquipmentType = "heavy_armor"
	EQUIPMENT_LIGHT_ARMOR  EquipmentType = "light_armor"
	EQUIPMENT_MAGIC_ROBE   EquipmentType = "magic_robe"
	EQUIPMENT_LEATHER_ARMOR EquipmentType = "leather_armor"
	
	// Accessories
	EQUIPMENT_POWER_RING    EquipmentType = "power_ring"
	EQUIPMENT_DEFENSE_RING  EquipmentType = "defense_ring"
	EQUIPMENT_SPEED_AMULET  EquipmentType = "speed_amulet"
	EQUIPMENT_HEALTH_CHARM  EquipmentType = "health_charm"
	
	// Consumables
	EQUIPMENT_HEALTH_POTION EquipmentType = "health_potion"
	EQUIPMENT_POWER_ELIXIR  EquipmentType = "power_elixir"
	EQUIPMENT_SPEED_BOOST   EquipmentType = "speed_boost"
	EQUIPMENT_SHIELD_SCROLL EquipmentType = "shield_scroll"
)

// Equipment configuration constants respecting fairness constraints
const (
	// Base stat modification caps (aligned with existing fairness system)
	EQUIPMENT_MAX_DAMAGE_BOOST    = 0.15  // 15% maximum damage boost
	EQUIPMENT_MAX_DEFENSE_BOOST   = 0.12  // 12% maximum defense boost  
	EQUIPMENT_MAX_SPEED_BOOST     = 0.08  // 8% maximum speed boost
	EQUIPMENT_MAX_HEALTH_BOOST    = 0.20  // 20% maximum health boost
	
	// Rarity multipliers for equipment effects
	RARITY_COMMON_MULTIPLIER    = 0.5   // 50% of base effect
	RARITY_UNCOMMON_MULTIPLIER  = 0.7   // 70% of base effect
	RARITY_RARE_MULTIPLIER      = 0.9   // 90% of base effect
	RARITY_EPIC_MULTIPLIER      = 1.0   // 100% of base effect
	RARITY_LEGENDARY_MULTIPLIER = 1.0   // 100% of base effect (no bonus to maintain fairness)
	
	// Equipment durability
	EQUIPMENT_MAX_DURABILITY = 100      // Maximum durability points
	EQUIPMENT_DURABILITY_LOSS_PER_BATTLE = 5  // Durability lost per battle
	
	// Equipment slots per character
	MAX_EQUIPMENT_SLOTS = 4  // weapon, armor, accessory, consumable
)

// BattleEquipment represents a piece of equipment with combat effects
type BattleEquipment struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Type         EquipmentType     `json:"type"`
	Slot         EquipmentSlot     `json:"slot"`
	Rarity       EquipmentRarity   `json:"rarity"`
	Description  string            `json:"description"`
	
	// Stat modifications
	AttackBonus   float64 `json:"attackBonus"`   // Damage multiplier bonus
	DefenseBonus  float64 `json:"defenseBonus"`  // Defense multiplier bonus
	SpeedBonus    float64 `json:"speedBonus"`    // Speed multiplier bonus
	HealthBonus   float64 `json:"healthBonus"`   // Health multiplier bonus
	
	// Special effects
	SpecialEffects []string `json:"specialEffects"` // List of special effects
	
	// Equipment properties
	Durability     int  `json:"durability"`     // Current durability
	MaxDurability  int  `json:"maxDurability"`  // Maximum durability
	IsBroken       bool `json:"isBroken"`       // Whether equipment is broken
	RequiredLevel  int  `json:"requiredLevel"`  // Level requirement to equip
	
	// Integration with gift system
	GiftID string `json:"giftID,omitempty"` // Associated gift item ID
}

// EquipmentLoadout represents a character's equipped items
type EquipmentLoadout struct {
	ParticipantID string                        `json:"participantID"`
	EquippedItems map[EquipmentSlot]*BattleEquipment `json:"equippedItems"`
	Inventory     []*BattleEquipment            `json:"inventory"`     // Available equipment
	StatBonuses   EquipmentStatBonuses          `json:"statBonuses"`   // Calculated total bonuses
}

// EquipmentStatBonuses represents the total stat modifications from equipped items
type EquipmentStatBonuses struct {
	AttackMultiplier  float64  `json:"attackMultiplier"`  // Total attack bonus (1.0 = no bonus)
	DefenseMultiplier float64  `json:"defenseMultiplier"` // Total defense bonus
	SpeedMultiplier   float64  `json:"speedMultiplier"`   // Total speed bonus
	HealthMultiplier  float64  `json:"healthMultiplier"`  // Total health bonus
	ActiveEffects     []string `json:"activeEffects"`     // All active special effects
}

// EquipmentManager handles equipment management for battle participants
type EquipmentManager struct {
	participantLoadouts map[string]*EquipmentLoadout `json:"participantLoadouts"`
	availableEquipment  map[string]*BattleEquipment  `json:"availableEquipment"` // Equipment database
}

// NewEquipmentManager creates a new equipment manager with default equipment
func NewEquipmentManager() *EquipmentManager {
	em := &EquipmentManager{
		participantLoadouts: make(map[string]*EquipmentLoadout),
		availableEquipment:  make(map[string]*BattleEquipment),
	}
	
	// Load default equipment database
	em.loadDefaultEquipment()
	
	return em
}

// loadDefaultEquipment populates the equipment database with standard items
func (em *EquipmentManager) loadDefaultEquipment() {
	defaultEquipment := []*BattleEquipment{
		// Common Weapons
		{
			ID: "iron_sword", Name: "Iron Sword", Type: EQUIPMENT_SWORD, Slot: SLOT_WEAPON,
			Rarity: RARITY_COMMON, Description: "A basic iron sword for combat",
			AttackBonus: EQUIPMENT_MAX_DAMAGE_BOOST * RARITY_COMMON_MULTIPLIER,
			Durability: EQUIPMENT_MAX_DURABILITY, MaxDurability: EQUIPMENT_MAX_DURABILITY,
			RequiredLevel: 1,
		},
		{
			ID: "wooden_bow", Name: "Wooden Bow", Type: EQUIPMENT_BOW, Slot: SLOT_WEAPON,
			Rarity: RARITY_COMMON, Description: "A simple wooden bow for ranged attacks",
			AttackBonus: EQUIPMENT_MAX_DAMAGE_BOOST * RARITY_COMMON_MULTIPLIER * 0.8, // Slightly less damage
			SpeedBonus: EQUIPMENT_MAX_SPEED_BOOST * RARITY_COMMON_MULTIPLIER * 0.5,   // Small speed bonus
			Durability: EQUIPMENT_MAX_DURABILITY, MaxDurability: EQUIPMENT_MAX_DURABILITY,
			RequiredLevel: 1,
		},
		
		// Uncommon Weapons
		{
			ID: "steel_sword", Name: "Steel Sword", Type: EQUIPMENT_SWORD, Slot: SLOT_WEAPON,
			Rarity: RARITY_UNCOMMON, Description: "A well-crafted steel sword",
			AttackBonus: EQUIPMENT_MAX_DAMAGE_BOOST * RARITY_UNCOMMON_MULTIPLIER,
			Durability: EQUIPMENT_MAX_DURABILITY, MaxDurability: EQUIPMENT_MAX_DURABILITY,
			RequiredLevel: 3,
		},
		{
			ID: "magic_staff", Name: "Magic Staff", Type: EQUIPMENT_STAFF, Slot: SLOT_WEAPON,
			Rarity: RARITY_UNCOMMON, Description: "A staff that enhances magical abilities",
			AttackBonus: EQUIPMENT_MAX_DAMAGE_BOOST * RARITY_UNCOMMON_MULTIPLIER * 0.7,
			SpecialEffects: []string{"magic_damage", "mana_efficiency"},
			Durability: EQUIPMENT_MAX_DURABILITY, MaxDurability: EQUIPMENT_MAX_DURABILITY,
			RequiredLevel: 3,
		},
		
		// Rare Weapons
		{
			ID: "enchanted_blade", Name: "Enchanted Blade", Type: EQUIPMENT_SWORD, Slot: SLOT_WEAPON,
			Rarity: RARITY_RARE, Description: "A magically enhanced sword",
			AttackBonus: EQUIPMENT_MAX_DAMAGE_BOOST * RARITY_RARE_MULTIPLIER,
			SpecialEffects: []string{"elemental_damage", "critical_chance"},
			Durability: EQUIPMENT_MAX_DURABILITY, MaxDurability: EQUIPMENT_MAX_DURABILITY,
			RequiredLevel: 5,
		},
		
		// Armor - Common
		{
			ID: "leather_armor", Name: "Leather Armor", Type: EQUIPMENT_LEATHER_ARMOR, Slot: SLOT_ARMOR,
			Rarity: RARITY_COMMON, Description: "Basic leather protection",
			DefenseBonus: EQUIPMENT_MAX_DEFENSE_BOOST * RARITY_COMMON_MULTIPLIER,
			Durability: EQUIPMENT_MAX_DURABILITY, MaxDurability: EQUIPMENT_MAX_DURABILITY,
			RequiredLevel: 1,
		},
		
		// Armor - Uncommon
		{
			ID: "chain_mail", Name: "Chain Mail", Type: EQUIPMENT_LIGHT_ARMOR, Slot: SLOT_ARMOR,
			Rarity: RARITY_UNCOMMON, Description: "Flexible metal armor",
			DefenseBonus: EQUIPMENT_MAX_DEFENSE_BOOST * RARITY_UNCOMMON_MULTIPLIER,
			SpeedBonus: -EQUIPMENT_MAX_SPEED_BOOST * 0.3, // Small speed penalty
			Durability: EQUIPMENT_MAX_DURABILITY, MaxDurability: EQUIPMENT_MAX_DURABILITY,
			RequiredLevel: 2,
		},
		
		// Armor - Rare
		{
			ID: "plate_armor", Name: "Plate Armor", Type: EQUIPMENT_HEAVY_ARMOR, Slot: SLOT_ARMOR,
			Rarity: RARITY_RARE, Description: "Heavy metal plate protection",
			DefenseBonus: EQUIPMENT_MAX_DEFENSE_BOOST * RARITY_RARE_MULTIPLIER,
			SpeedBonus: -EQUIPMENT_MAX_SPEED_BOOST * 0.5, // Speed penalty for heavy armor
			SpecialEffects: []string{"damage_reflection", "knockback_resist"},
			Durability: EQUIPMENT_MAX_DURABILITY, MaxDurability: EQUIPMENT_MAX_DURABILITY,
			RequiredLevel: 4,
		},
		
		// Accessories
		{
			ID: "power_ring", Name: "Ring of Power", Type: EQUIPMENT_POWER_RING, Slot: SLOT_ACCESSORY,
			Rarity: RARITY_UNCOMMON, Description: "A ring that enhances attack power",
			AttackBonus: EQUIPMENT_MAX_DAMAGE_BOOST * RARITY_UNCOMMON_MULTIPLIER * 0.6,
			Durability: EQUIPMENT_MAX_DURABILITY, MaxDurability: EQUIPMENT_MAX_DURABILITY,
			RequiredLevel: 2,
		},
		{
			ID: "health_charm", Name: "Charm of Vitality", Type: EQUIPMENT_HEALTH_CHARM, Slot: SLOT_ACCESSORY,
			Rarity: RARITY_RARE, Description: "A charm that increases health",
			HealthBonus: EQUIPMENT_MAX_HEALTH_BOOST * RARITY_RARE_MULTIPLIER,
			SpecialEffects: []string{"health_regeneration"},
			Durability: EQUIPMENT_MAX_DURABILITY, MaxDurability: EQUIPMENT_MAX_DURABILITY,
			RequiredLevel: 3,
		},
		
		// Consumables
		{
			ID: "health_potion", Name: "Health Potion", Type: EQUIPMENT_HEALTH_POTION, Slot: SLOT_CONSUMABLE,
			Rarity: RARITY_COMMON, Description: "Restores health when used",
			SpecialEffects: []string{"instant_heal_50"},
			Durability: 1, MaxDurability: 1, // Single use
			RequiredLevel: 1,
		},
		{
			ID: "power_elixir", Name: "Power Elixir", Type: EQUIPMENT_POWER_ELIXIR, Slot: SLOT_CONSUMABLE,
			Rarity: RARITY_UNCOMMON, Description: "Temporarily boosts attack power",
			AttackBonus: EQUIPMENT_MAX_DAMAGE_BOOST * 0.5, // Temporary boost
			SpecialEffects: []string{"temporary_power_boost"},
			Durability: 1, MaxDurability: 1,
			RequiredLevel: 2,
		},
	}
	
	// Add all equipment to the database
	for _, equipment := range defaultEquipment {
		em.availableEquipment[equipment.ID] = equipment
	}
}

// InitializeParticipantLoadout creates an equipment loadout for a battle participant
func (em *EquipmentManager) InitializeParticipantLoadout(participantID string, characterLevel int) {
	loadout := &EquipmentLoadout{
		ParticipantID: participantID,
		EquippedItems: make(map[EquipmentSlot]*BattleEquipment),
		Inventory:     make([]*BattleEquipment, 0),
		StatBonuses:   EquipmentStatBonuses{
			AttackMultiplier:  1.0,
			DefenseMultiplier: 1.0,
			SpeedMultiplier:   1.0,
			HealthMultiplier:  1.0,
			ActiveEffects:     make([]string, 0),
		},
	}
	
	// Add starter equipment based on character level
	em.addStarterEquipment(loadout, characterLevel)
	
	// Calculate initial stat bonuses
	em.calculateStatBonuses(loadout)
	
	em.participantLoadouts[participantID] = loadout
}

// addStarterEquipment provides basic equipment based on character level
func (em *EquipmentManager) addStarterEquipment(loadout *EquipmentLoadout, characterLevel int) {
	// Add basic equipment that the character can use
	for _, equipment := range em.availableEquipment {
		if equipment.RequiredLevel <= characterLevel {
			// Add a copy to inventory
			equipmentCopy := *equipment // Copy the equipment
			loadout.Inventory = append(loadout.Inventory, &equipmentCopy)
		}
	}
	
	// Auto-equip starter items if inventory has them
	em.autoEquipStarterItems(loadout)
}

// autoEquipStarterItems automatically equips basic items for new characters
func (em *EquipmentManager) autoEquipStarterItems(loadout *EquipmentLoadout) {
	// Try to equip one item per slot with the best available
	slotPriorities := map[EquipmentSlot][]EquipmentType{
		SLOT_WEAPON: {EQUIPMENT_SWORD, EQUIPMENT_BOW, EQUIPMENT_STAFF, EQUIPMENT_DAGGER},
		SLOT_ARMOR:  {EQUIPMENT_LEATHER_ARMOR, EQUIPMENT_LIGHT_ARMOR, EQUIPMENT_HEAVY_ARMOR},
		SLOT_ACCESSORY: {EQUIPMENT_POWER_RING, EQUIPMENT_HEALTH_CHARM, EQUIPMENT_SPEED_AMULET},
	}
	
	for slot, preferredTypes := range slotPriorities {
		for _, equipType := range preferredTypes {
			for i, equipment := range loadout.Inventory {
				if equipment.Type == equipType && equipment.Slot == slot {
					// Equip this item
					loadout.EquippedItems[slot] = equipment
					// Remove from inventory
					loadout.Inventory = append(loadout.Inventory[:i], loadout.Inventory[i+1:]...)
					break
				}
			}
			// Stop after finding first item for this slot
			if loadout.EquippedItems[slot] != nil {
				break
			}
		}
	}
}

// EquipItem equips an item from inventory to the appropriate slot
func (em *EquipmentManager) EquipItem(participantID, equipmentID string) error {
	loadout := em.participantLoadouts[participantID]
	if loadout == nil {
		return errors.New("participant loadout not found")
	}
	
	// Find equipment in inventory
	var equipment *BattleEquipment
	var inventoryIndex int
	for i, item := range loadout.Inventory {
		if item.ID == equipmentID {
			equipment = item
			inventoryIndex = i
			break
		}
	}
	
	if equipment == nil {
		return ErrEquipmentNotFound
	}
	
	// Check if equipment is broken
	if equipment.IsBroken {
		return errors.New("cannot equip broken equipment")
	}
	
	// Check if slot is occupied
	if existing := loadout.EquippedItems[equipment.Slot]; existing != nil {
		// Move existing equipment back to inventory
		loadout.Inventory = append(loadout.Inventory, existing)
	}
	
	// Equip the new item
	loadout.EquippedItems[equipment.Slot] = equipment
	
	// Remove from inventory
	loadout.Inventory = append(loadout.Inventory[:inventoryIndex], loadout.Inventory[inventoryIndex+1:]...)
	
	// Recalculate stat bonuses
	em.calculateStatBonuses(loadout)
	
	return nil
}

// UnequipItem removes an equipped item and returns it to inventory
func (em *EquipmentManager) UnequipItem(participantID string, slot EquipmentSlot) error {
	loadout := em.participantLoadouts[participantID]
	if loadout == nil {
		return errors.New("participant loadout not found")
	}
	
	equipment := loadout.EquippedItems[slot]
	if equipment == nil {
		return errors.New("no equipment in that slot")
	}
	
	// Move equipment back to inventory
	loadout.Inventory = append(loadout.Inventory, equipment)
	
	// Remove from equipped items
	delete(loadout.EquippedItems, slot)
	
	// Recalculate stat bonuses
	em.calculateStatBonuses(loadout)
	
	return nil
}

// calculateStatBonuses computes the total stat modifications from equipped items
func (em *EquipmentManager) calculateStatBonuses(loadout *EquipmentLoadout) {
	bonuses := EquipmentStatBonuses{
		AttackMultiplier:  1.0,
		DefenseMultiplier: 1.0,
		SpeedMultiplier:   1.0,
		HealthMultiplier:  1.0,
		ActiveEffects:     make([]string, 0),
	}
	
	// Sum bonuses from all equipped items
	for _, equipment := range loadout.EquippedItems {
		if equipment.IsBroken {
			continue // Broken equipment provides no bonuses
		}
		
		bonuses.AttackMultiplier += equipment.AttackBonus
		bonuses.DefenseMultiplier += equipment.DefenseBonus
		bonuses.SpeedMultiplier += equipment.SpeedBonus
		bonuses.HealthMultiplier += equipment.HealthBonus
		
		// Add special effects
		bonuses.ActiveEffects = append(bonuses.ActiveEffects, equipment.SpecialEffects...)
	}
	
	// Apply fairness caps to prevent overpowered combinations
	bonuses.AttackMultiplier = em.clampBonus(bonuses.AttackMultiplier, 1.0+EQUIPMENT_MAX_DAMAGE_BOOST)
	bonuses.DefenseMultiplier = em.clampBonus(bonuses.DefenseMultiplier, 1.0+EQUIPMENT_MAX_DEFENSE_BOOST)
	bonuses.SpeedMultiplier = em.clampBonus(bonuses.SpeedMultiplier, 1.0+EQUIPMENT_MAX_SPEED_BOOST)
	bonuses.HealthMultiplier = em.clampBonus(bonuses.HealthMultiplier, 1.0+EQUIPMENT_MAX_HEALTH_BOOST)
	
	loadout.StatBonuses = bonuses
}

// clampBonus ensures bonus values don't exceed fairness constraints
func (em *EquipmentManager) clampBonus(value, maxValue float64) float64 {
	if value > maxValue {
		return maxValue
	}
	if value < 0.5 { // Minimum 50% effectiveness to avoid negative stats
		return 0.5
	}
	return value
}

// ApplyEquipmentBonuses modifies battle stats based on equipped items
func (em *EquipmentManager) ApplyEquipmentBonuses(participantID string, baseStats *BattleStats) *BattleStats {
	loadout := em.participantLoadouts[participantID]
	if loadout == nil {
		return baseStats // No equipment, return original stats
	}
	
	// Create modified stats
	modifiedStats := *baseStats // Copy original stats
	bonuses := loadout.StatBonuses
	
	// Apply stat multipliers
	modifiedStats.Attack *= bonuses.AttackMultiplier
	modifiedStats.Defense *= bonuses.DefenseMultiplier
	modifiedStats.Speed *= bonuses.SpeedMultiplier
	modifiedStats.MaxHP *= bonuses.HealthMultiplier
	
	// Adjust current HP proportionally if MaxHP changed
	if bonuses.HealthMultiplier != 1.0 {
		healthRatio := modifiedStats.HP / baseStats.MaxHP
		modifiedStats.HP = modifiedStats.MaxHP * healthRatio
	}
	
	return &modifiedStats
}

// GetEquippedItems returns the currently equipped items for a participant
func (em *EquipmentManager) GetEquippedItems(participantID string) map[EquipmentSlot]*BattleEquipment {
	loadout := em.participantLoadouts[participantID]
	if loadout == nil {
		return make(map[EquipmentSlot]*BattleEquipment)
	}
	return loadout.EquippedItems
}

// GetInventory returns the inventory items for a participant  
func (em *EquipmentManager) GetInventory(participantID string) []*BattleEquipment {
	loadout := em.participantLoadouts[participantID]
	if loadout == nil {
		return make([]*BattleEquipment, 0)
	}
	return loadout.Inventory
}

// GetStatBonuses returns the calculated stat bonuses for a participant
func (em *EquipmentManager) GetStatBonuses(participantID string) EquipmentStatBonuses {
	loadout := em.participantLoadouts[participantID]
	if loadout == nil {
		return EquipmentStatBonuses{
			AttackMultiplier: 1.0, DefenseMultiplier: 1.0,
			SpeedMultiplier: 1.0, HealthMultiplier: 1.0,
			ActiveEffects: make([]string, 0),
		}
	}
	return loadout.StatBonuses
}

// DamageEquipment reduces durability of equipped items after battle
func (em *EquipmentManager) DamageEquipment(participantID string) {
	loadout := em.participantLoadouts[participantID]
	if loadout == nil {
		return
	}
	
	// Reduce durability of all equipped items
	for _, equipment := range loadout.EquippedItems {
		equipment.Durability -= EQUIPMENT_DURABILITY_LOSS_PER_BATTLE
		if equipment.Durability <= 0 {
			equipment.Durability = 0
			equipment.IsBroken = true
		}
	}
	
	// Recalculate bonuses in case items broke
	em.calculateStatBonuses(loadout)
}

// RepairEquipment restores durability to an item (costs resources in real game)
func (em *EquipmentManager) RepairEquipment(participantID, equipmentID string) error {
	loadout := em.participantLoadouts[participantID]
	if loadout == nil {
		return errors.New("participant loadout not found")
	}
	
	// Find equipment in equipped items or inventory
	var equipment *BattleEquipment
	for _, item := range loadout.EquippedItems {
		if item.ID == equipmentID {
			equipment = item
			break
		}
	}
	if equipment == nil {
		for _, item := range loadout.Inventory {
			if item.ID == equipmentID {
				equipment = item
				break
			}
		}
	}
	
	if equipment == nil {
		return ErrEquipmentNotFound
	}
	
	// Repair the equipment
	equipment.Durability = equipment.MaxDurability
	equipment.IsBroken = false
	
	// Recalculate bonuses if it was equipped
	em.calculateStatBonuses(loadout)
	
	return nil
}

// UseConsumable activates a consumable item's effects
func (em *EquipmentManager) UseConsumable(participantID, equipmentID string) (*BattleResult, error) {
	loadout := em.participantLoadouts[participantID]
	if loadout == nil {
		return nil, errors.New("participant loadout not found")
	}
	
	// Find consumable in inventory or equipped slot
	var equipment *BattleEquipment
	var isEquipped bool
	var inventoryIndex int
	
	// Check equipped consumable slot first
	if equippedConsumable := loadout.EquippedItems[SLOT_CONSUMABLE]; equippedConsumable != nil && equippedConsumable.ID == equipmentID {
		equipment = equippedConsumable
		isEquipped = true
	} else {
		// Check inventory
		for i, item := range loadout.Inventory {
			if item.ID == equipmentID && item.Slot == SLOT_CONSUMABLE {
				equipment = item
				inventoryIndex = i
				break
			}
		}
	}
	
	if equipment == nil {
		return nil, ErrEquipmentNotFound
	}
	
	if equipment.Slot != SLOT_CONSUMABLE {
		return nil, errors.New("item is not consumable")
	}
	
	// Create battle result based on consumable effects
	result := em.createConsumableResult(equipment)
	
	// Reduce durability (consumables typically have durability of 1)
	equipment.Durability--
	if equipment.Durability <= 0 {
		// Remove consumed item
		if isEquipped {
			delete(loadout.EquippedItems, SLOT_CONSUMABLE)
		} else {
			loadout.Inventory = append(loadout.Inventory[:inventoryIndex], loadout.Inventory[inventoryIndex+1:]...)
		}
	}
	
	return result, nil
}

// createConsumableResult generates battle effects for consumable items
func (em *EquipmentManager) createConsumableResult(equipment *BattleEquipment) *BattleResult {
	result := &BattleResult{
		Success:     true,
		Animation:   em.getConsumableAnimation(equipment.Type),
		Response:    fmt.Sprintf("uses %s!", equipment.Name),
		StatusEffects: equipment.SpecialEffects,
	}
	
	// Apply consumable-specific effects
	switch equipment.Type {
	case EQUIPMENT_HEALTH_POTION:
		result.Healing = 50.0 // Fixed healing amount
		
	case EQUIPMENT_POWER_ELIXIR:
		result.ModifiersApplied = []BattleModifier{
			{
				Type:     MODIFIER_DAMAGE,
				Value:    1.0 + equipment.AttackBonus,
				Duration: 3,
				Source:   "power_elixir",
			},
		}
		
	case EQUIPMENT_SPEED_BOOST:
		result.ModifiersApplied = []BattleModifier{
			{
				Type:     MODIFIER_SPEED,
				Value:    1.0 + EQUIPMENT_MAX_SPEED_BOOST,
				Duration: 2,
				Source:   "speed_boost",
			},
		}
		
	case EQUIPMENT_SHIELD_SCROLL:
		result.ModifiersApplied = []BattleModifier{
			{
				Type:     MODIFIER_SHIELD,
				Value:    25.0, // Shield absorption
				Duration: 4,
				Source:   "shield_scroll",
			},
		}
	}
	
	return result
}

// getConsumableAnimation returns animation name for consumable use
func (em *EquipmentManager) getConsumableAnimation(equipType EquipmentType) string {
	animationMap := map[EquipmentType]string{
		EQUIPMENT_HEALTH_POTION: "drink_potion",
		EQUIPMENT_POWER_ELIXIR:  "drink_elixir", 
		EQUIPMENT_SPEED_BOOST:   "use_boost",
		EQUIPMENT_SHIELD_SCROLL: "cast_scroll",
	}
	
	if animation, exists := animationMap[equipType]; exists {
		return animation
	}
	return "use_item"
}

// GetAvailableEquipment returns all equipment in the database
func (em *EquipmentManager) GetAvailableEquipment() map[string]*BattleEquipment {
	return em.availableEquipment
}

// AddEquipmentToInventory adds equipment to a participant's inventory
func (em *EquipmentManager) AddEquipmentToInventory(participantID, equipmentID string) error {
	loadout := em.participantLoadouts[participantID]
	if loadout == nil {
		return errors.New("participant loadout not found")
	}
	
	baseEquipment := em.availableEquipment[equipmentID]
	if baseEquipment == nil {
		return ErrEquipmentNotFound
	}
	
	// Create a copy for the participant
	equipmentCopy := *baseEquipment
	loadout.Inventory = append(loadout.Inventory, &equipmentCopy)
	
	return nil
}

// GetEquipmentInfo returns detailed information about a specific equipment piece
func (em *EquipmentManager) GetEquipmentInfo(equipmentID string) (*BattleEquipment, error) {
	equipment := em.availableEquipment[equipmentID]
	if equipment == nil {
		return nil, ErrEquipmentNotFound
	}
	return equipment, nil
}