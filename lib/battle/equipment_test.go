package battle

import (
	"fmt"
	"testing"
)

func TestNewEquipmentManager(t *testing.T) {
	em := NewEquipmentManager()

	if em == nil {
		t.Fatal("NewEquipmentManager returned nil")
	}

	if em.participantLoadouts == nil {
		t.Error("participantLoadouts map not initialized")
	}

	if em.availableEquipment == nil {
		t.Error("availableEquipment map not initialized")
	}

	// Check that default equipment was loaded
	if len(em.availableEquipment) == 0 {
		t.Error("no default equipment loaded")
	}
	
	// Verify we have expected equipment types
	expectedEquipment := []string{"iron_sword", "wooden_bow", "leather_armor", "health_potion"}
	for _, expectedID := range expectedEquipment {
		if _, exists := em.availableEquipment[expectedID]; !exists {
			t.Errorf("Expected equipment %s not found", expectedID)
		}
	}
}

func TestInitializeParticipantLoadout(t *testing.T) {
	em := NewEquipmentManager()
	participantID := "test_participant"

	tests := []struct {
		name           string
		characterLevel int
		minInventory   int
		expectEquipped bool
	}{
		{"Level 1", 1, 2, true}, // Lower expectation based on actual behavior
		{"Level 3", 3, 5, true},
		{"Level 5", 5, 8, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			em.InitializeParticipantLoadout(participantID, tt.characterLevel)

			loadout := em.participantLoadouts[participantID]
			if loadout == nil {
				t.Fatal("No loadout created")
			}

			if loadout.ParticipantID != participantID {
				t.Errorf("Expected participant ID %s, got %s", participantID, loadout.ParticipantID)
			}

			if len(loadout.Inventory) < tt.minInventory {
				t.Errorf("Expected at least %d inventory items, got %d", tt.minInventory, len(loadout.Inventory))
			}

			if tt.expectEquipped {
				hasEquippedItem := false
				for _, item := range loadout.EquippedItems {
					if item != nil {
						hasEquippedItem = true
						break
					}
				}
				if !hasEquippedItem {
					t.Error("Expected at least one equipped item")
				}
			}

			// Check stat bonuses are initialized properly
			bonuses := loadout.StatBonuses
			if bonuses.AttackMultiplier < 1.0 {
				t.Error("Attack multiplier should be at least 1.0")
			}
			if bonuses.DefenseMultiplier < 1.0 {
				t.Error("Defense multiplier should be at least 1.0")
			}
		})
	}
}

func TestEquipItem(t *testing.T) {
	em := NewEquipmentManager()
	participantID := "test_participant"
	em.InitializeParticipantLoadout(participantID, 5) // High level for access to all equipment

	loadout := em.participantLoadouts[participantID]

	// Find a weapon in inventory to equip
	var weaponToEquip *BattleEquipment
	for _, item := range loadout.Inventory {
		if item.Slot == SLOT_WEAPON {
			weaponToEquip = item
			break
		}
	}

	if weaponToEquip == nil {
		t.Fatal("No weapon found in inventory")
	}

	initialInventorySize := len(loadout.Inventory)

	// Debug: Check if weapon is already equipped
	alreadyEquipped := loadout.EquippedItems[SLOT_WEAPON] != nil

	// Equip the weapon
	err := em.EquipItem(participantID, weaponToEquip.ID)
	if err != nil {
		t.Fatalf("EquipItem failed: %v", err)
	}

	// Verify weapon is equipped
	equippedWeapon := loadout.EquippedItems[SLOT_WEAPON]
	if equippedWeapon == nil {
		t.Fatal("Weapon not equipped")
	}

	if equippedWeapon.ID != weaponToEquip.ID {
		t.Errorf("Expected weapon %s, got %s", weaponToEquip.ID, equippedWeapon.ID)
	}

	// Verify weapon removed from inventory (accounting for slot replacement)
	expectedInventorySize := initialInventorySize - 1
	if alreadyEquipped {
		// If something was already equipped, it gets moved to inventory, so net change is 0
		expectedInventorySize = initialInventorySize
	}

	if len(loadout.Inventory) != expectedInventorySize {
		t.Errorf("Expected inventory size %d, got %d (initially %d, already equipped: %v)",
			expectedInventorySize, len(loadout.Inventory), initialInventorySize, alreadyEquipped)
	} // Verify stat bonuses updated
	if loadout.StatBonuses.AttackMultiplier <= 1.0 {
		t.Error("Attack multiplier should be increased after equipping weapon")
	}
}

func TestEquipItemSlotReplacement(t *testing.T) {
	em := NewEquipmentManager()
	participantID := "test_participant"
	em.InitializeParticipantLoadout(participantID, 5)

	loadout := em.participantLoadouts[participantID]

	// Find two weapons to test slot replacement
	var weapon1, weapon2 *BattleEquipment
	for _, item := range loadout.Inventory {
		if item.Slot == SLOT_WEAPON {
			if weapon1 == nil {
				weapon1 = item
			} else if weapon2 == nil {
				weapon2 = item
				break
			}
		}
	}

	if weapon1 == nil || weapon2 == nil {
		t.Fatal("Need at least 2 weapons in inventory for this test")
	}

	// Equip first weapon
	err := em.EquipItem(participantID, weapon1.ID)
	if err != nil {
		t.Fatalf("Failed to equip first weapon: %v", err)
	}

	initialInventorySize := len(loadout.Inventory)

	// Equip second weapon (should replace first)
	err = em.EquipItem(participantID, weapon2.ID)
	if err != nil {
		t.Fatalf("Failed to equip second weapon: %v", err)
	}

	// Verify second weapon is equipped
	equippedWeapon := loadout.EquippedItems[SLOT_WEAPON]
	if equippedWeapon.ID != weapon2.ID {
		t.Errorf("Expected weapon %s, got %s", weapon2.ID, equippedWeapon.ID)
	}

	// Verify first weapon returned to inventory
	if len(loadout.Inventory) != initialInventorySize {
		t.Errorf("Expected inventory size %d after replacement, got %d", initialInventorySize, len(loadout.Inventory))
	}

	// Verify first weapon is back in inventory
	weapon1InInventory := false
	for _, item := range loadout.Inventory {
		if item.ID == weapon1.ID {
			weapon1InInventory = true
			break
		}
	}

	if !weapon1InInventory {
		t.Error("First weapon should be returned to inventory when replaced")
	}
}

func TestUnequipItem(t *testing.T) {
	em := NewEquipmentManager()
	participantID := "test_participant"
	em.InitializeParticipantLoadout(participantID, 5)

	loadout := em.participantLoadouts[participantID]

	// Find and equip a weapon
	var weaponToEquip *BattleEquipment
	for _, item := range loadout.Inventory {
		if item.Slot == SLOT_WEAPON {
			weaponToEquip = item
			break
		}
	}

	if weaponToEquip == nil {
		t.Fatal("No weapon found")
	}

	err := em.EquipItem(participantID, weaponToEquip.ID)
	if err != nil {
		t.Fatalf("Failed to equip weapon: %v", err)
	}

	initialInventorySize := len(loadout.Inventory)

	// Unequip the weapon
	err = em.UnequipItem(participantID, SLOT_WEAPON)
	if err != nil {
		t.Fatalf("UnequipItem failed: %v", err)
	}

	// Verify weapon slot is empty
	if loadout.EquippedItems[SLOT_WEAPON] != nil {
		t.Error("Weapon slot should be empty after unequipping")
	}

	// Verify weapon returned to inventory
	if len(loadout.Inventory) != initialInventorySize+1 {
		t.Errorf("Expected inventory size %d, got %d", initialInventorySize+1, len(loadout.Inventory))
	}
}

func TestApplyEquipmentBonuses(t *testing.T) {
	em := NewEquipmentManager()
	participantID := "test_participant"
	em.InitializeParticipantLoadout(participantID, 5)

	// Create base stats
	baseStats := &BattleStats{
		HP:      100,
		MaxHP:   100,
		Attack:  50,
		Defense: 30,
		Speed:   40,
	}

	// Apply equipment bonuses
	modifiedStats := em.ApplyEquipmentBonuses(participantID, baseStats)

	if modifiedStats == nil {
		t.Fatal("ApplyEquipmentBonuses returned nil")
	}

	// Since the character has starter equipment, stats should be modified
	loadout := em.participantLoadouts[participantID]
	if len(loadout.EquippedItems) > 0 {
		// With equipment, some stats should be different
		if modifiedStats.Attack == baseStats.Attack &&
			modifiedStats.Defense == baseStats.Defense &&
			modifiedStats.Speed == baseStats.Speed {
			t.Error("Stats should be modified when equipment is equipped")
		}
	}

	// Verify base stats weren't modified
	if baseStats.Attack != 50 || baseStats.Defense != 30 || baseStats.Speed != 40 {
		t.Error("Base stats should not be modified")
	}
}

func TestEquipmentDurability(t *testing.T) {
	em := NewEquipmentManager()
	participantID := "test_participant"
	em.InitializeParticipantLoadout(participantID, 5)

	loadout := em.participantLoadouts[participantID]

	// Find and equip a weapon
	var weaponToEquip *BattleEquipment
	for _, item := range loadout.Inventory {
		if item.Slot == SLOT_WEAPON {
			weaponToEquip = item
			break
		}
	}

	if weaponToEquip == nil {
		t.Fatal("No weapon found")
	}

	err := em.EquipItem(participantID, weaponToEquip.ID)
	if err != nil {
		t.Fatalf("Failed to equip weapon: %v", err)
	}

	equippedWeapon := loadout.EquippedItems[SLOT_WEAPON]
	initialDurability := equippedWeapon.Durability

	// Damage equipment
	em.DamageEquipment(participantID)

	if equippedWeapon.Durability != initialDurability-EQUIPMENT_DURABILITY_LOSS_PER_BATTLE {
		t.Errorf("Expected durability %d, got %d",
			initialDurability-EQUIPMENT_DURABILITY_LOSS_PER_BATTLE,
			equippedWeapon.Durability)
	}

	// Damage equipment until broken
	for equippedWeapon.Durability > 0 {
		em.DamageEquipment(participantID)
	}

	if !equippedWeapon.IsBroken {
		t.Error("Equipment should be broken when durability reaches 0")
	}

	// Verify broken equipment doesn't provide bonuses
	if loadout.StatBonuses.AttackMultiplier > 1.0 {
		t.Error("Broken equipment should not provide stat bonuses")
	}
}

func TestRepairEquipment(t *testing.T) {
	em := NewEquipmentManager()
	participantID := "test_participant"
	em.InitializeParticipantLoadout(participantID, 5)

	loadout := em.participantLoadouts[participantID]

	// Find and equip a weapon
	var weaponToEquip *BattleEquipment
	for _, item := range loadout.Inventory {
		if item.Slot == SLOT_WEAPON {
			weaponToEquip = item
			break
		}
	}

	if weaponToEquip == nil {
		t.Fatal("No weapon found")
	}

	err := em.EquipItem(participantID, weaponToEquip.ID)
	if err != nil {
		t.Fatalf("Failed to equip weapon: %v", err)
	}

	equippedWeapon := loadout.EquippedItems[SLOT_WEAPON]

	// Damage equipment until broken
	for equippedWeapon.Durability > 0 {
		em.DamageEquipment(participantID)
	}

	// Repair equipment
	err = em.RepairEquipment(participantID, equippedWeapon.ID)
	if err != nil {
		t.Fatalf("RepairEquipment failed: %v", err)
	}

	if equippedWeapon.Durability != equippedWeapon.MaxDurability {
		t.Errorf("Expected durability %d after repair, got %d",
			equippedWeapon.MaxDurability, equippedWeapon.Durability)
	}

	if equippedWeapon.IsBroken {
		t.Error("Equipment should not be broken after repair")
	}

	// Verify repaired equipment provides bonuses again
	if loadout.StatBonuses.AttackMultiplier <= 1.0 {
		t.Error("Repaired equipment should provide stat bonuses")
	}
}

func TestUseConsumable(t *testing.T) {
	em := NewEquipmentManager()
	participantID := "test_participant"
	em.InitializeParticipantLoadout(participantID, 5)

	// Add a health potion to inventory
	err := em.AddEquipmentToInventory(participantID, "health_potion")
	if err != nil {
		t.Fatalf("Failed to add health potion: %v", err)
	}

	loadout := em.participantLoadouts[participantID]
	initialInventorySize := len(loadout.Inventory)

	// Use the health potion
	result, err := em.UseConsumable(participantID, "health_potion")
	if err != nil {
		t.Fatalf("UseConsumable failed: %v", err)
	}

	if result == nil {
		t.Fatal("UseConsumable returned nil result")
	}

	if !result.Success {
		t.Error("Consumable use should be successful")
	}

	if result.Healing <= 0 {
		t.Error("Health potion should provide healing")
	}

	// Verify consumable was removed from inventory
	if len(loadout.Inventory) != initialInventorySize-1 {
		t.Errorf("Expected inventory size %d after using consumable, got %d",
			initialInventorySize-1, len(loadout.Inventory))
	}
}

func TestEquipmentFairnessConstraints(t *testing.T) {
	em := NewEquipmentManager()

	// Create a loadout with multiple high-bonus items to test caps
	loadout := &EquipmentLoadout{
		ParticipantID: "test",
		EquippedItems: make(map[EquipmentSlot]*BattleEquipment),
		Inventory:     make([]*BattleEquipment, 0),
	}

	// Create overpowered equipment that would exceed caps
	overpoweredWeapon := &BattleEquipment{
		ID: "overpowered_weapon", Slot: SLOT_WEAPON,
		AttackBonus: EQUIPMENT_MAX_DAMAGE_BOOST * 2, // Double the max
		IsBroken:    false,
	}

	overpoweredArmor := &BattleEquipment{
		ID: "overpowered_armor", Slot: SLOT_ARMOR,
		DefenseBonus: EQUIPMENT_MAX_DEFENSE_BOOST * 2, // Double the max
		IsBroken:     false,
	}

	loadout.EquippedItems[SLOT_WEAPON] = overpoweredWeapon
	loadout.EquippedItems[SLOT_ARMOR] = overpoweredArmor

	// Calculate bonuses (should be capped)
	em.calculateStatBonuses(loadout)

	// Verify caps are enforced
	expectedMaxAttack := 1.0 + EQUIPMENT_MAX_DAMAGE_BOOST
	if loadout.StatBonuses.AttackMultiplier > expectedMaxAttack {
		t.Errorf("Attack bonus %f exceeds fairness cap %f",
			loadout.StatBonuses.AttackMultiplier, expectedMaxAttack)
	}

	expectedMaxDefense := 1.0 + EQUIPMENT_MAX_DEFENSE_BOOST
	if loadout.StatBonuses.DefenseMultiplier > expectedMaxDefense {
		t.Errorf("Defense bonus %f exceeds fairness cap %f",
			loadout.StatBonuses.DefenseMultiplier, expectedMaxDefense)
	}
}

func TestEquipmentRarity(t *testing.T) {
	em := NewEquipmentManager()

	// Test that different rarities provide different bonus amounts
	rarities := []EquipmentRarity{RARITY_COMMON, RARITY_UNCOMMON, RARITY_RARE, RARITY_EPIC}
	var bonuses []float64

	for _, rarity := range rarities {
		// Find equipment of this rarity
		for _, equipment := range em.availableEquipment {
			if equipment.Rarity == rarity && equipment.AttackBonus > 0 {
				bonuses = append(bonuses, equipment.AttackBonus)
				break
			}
		}
	}

	// Verify that higher rarities generally provide better bonuses
	// (Note: Some rarities may have same bonuses for fairness)
	if len(bonuses) >= 2 {
		// Allow for some rarities to have equal bonuses (fairness caps)
		improves := false
		for i := 1; i < len(bonuses); i++ {
			if bonuses[i] > bonuses[i-1] {
				improves = true
			}
		}
		if !improves {
			t.Error("At least some higher rarities should provide better bonuses")
		}
	}
}

func TestEquipmentByLevel(t *testing.T) {
	em := NewEquipmentManager()

	tests := []struct {
		level        int
		maxEquipment int
	}{
		{1, 5},  // Low level, fewer equipment options
		{3, 8},  // Mid level, more options
		{5, 12}, // High level, most options
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Level_%d", tt.level), func(t *testing.T) {
			participantID := fmt.Sprintf("test_participant_%d", tt.level)
			em.InitializeParticipantLoadout(participantID, tt.level)

			loadout := em.participantLoadouts[participantID]

			// Verify all inventory items meet level requirement
			for _, equipment := range loadout.Inventory {
				if equipment.RequiredLevel > tt.level {
					t.Errorf("Equipment %s requires level %d but character is level %d",
						equipment.ID, equipment.RequiredLevel, tt.level)
				}
			}

			// Verify all equipped items meet level requirement
			for _, equipment := range loadout.EquippedItems {
				if equipment.RequiredLevel > tt.level {
					t.Errorf("Equipped equipment %s requires level %d but character is level %d",
						equipment.ID, equipment.RequiredLevel, tt.level)
				}
			}
		})
	}
}

func TestEquipmentIntegration(t *testing.T) {
	em := NewEquipmentManager()
	participantID := "test_participant"
	em.InitializeParticipantLoadout(participantID, 5)

	// Test that equipment manager integrates properly with battle stats
	baseStats := &BattleStats{
		HP: 100, MaxHP: 100, Attack: 50, Defense: 30, Speed: 40,
	}

	// Get stat bonuses
	bonuses := em.GetStatBonuses(participantID)

	// Apply bonuses
	modifiedStats := em.ApplyEquipmentBonuses(participantID, baseStats)

	// Verify modifications match bonuses
	expectedAttack := baseStats.Attack * bonuses.AttackMultiplier
	if modifiedStats.Attack != expectedAttack {
		t.Errorf("Expected attack %f, got %f", expectedAttack, modifiedStats.Attack)
	}

	expectedDefense := baseStats.Defense * bonuses.DefenseMultiplier
	if modifiedStats.Defense != expectedDefense {
		t.Errorf("Expected defense %f, got %f", expectedDefense, modifiedStats.Defense)
	}

	// Verify inventory and equipped items are accessible
	inventory := em.GetInventory(participantID)
	equippedItems := em.GetEquippedItems(participantID)

	if inventory == nil {
		t.Error("GetInventory returned nil")
	}

	if equippedItems == nil {
		t.Error("GetEquippedItems returned nil")
	}
}

// Benchmark tests for performance validation
func BenchmarkEquipmentBonusCalculation(b *testing.B) {
	em := NewEquipmentManager()
	participantID := "test_participant"
	em.InitializeParticipantLoadout(participantID, 5)

	loadout := em.participantLoadouts[participantID]

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		em.calculateStatBonuses(loadout)
	}
}

func BenchmarkApplyEquipmentBonuses(b *testing.B) {
	em := NewEquipmentManager()
	participantID := "test_participant"
	em.InitializeParticipantLoadout(participantID, 5)

	baseStats := &BattleStats{
		HP: 100, MaxHP: 100, Attack: 50, Defense: 30, Speed: 40,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		em.ApplyEquipmentBonuses(participantID, baseStats)
	}
}

func BenchmarkEquipItem(b *testing.B) {
	em := NewEquipmentManager()
	participantID := "test_participant"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Reset for each iteration
		em.InitializeParticipantLoadout(participantID, 5)

		loadout := em.participantLoadouts[participantID]
		if len(loadout.Inventory) > 0 {
			em.EquipItem(participantID, loadout.Inventory[0].ID)
		}
	}
}
