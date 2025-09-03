package character

import (
	"github.com/opd-ai/desktop-companion/internal/battle"
	"fmt"
)

// BattleGiftProvider implements the battle.GiftProvider interface
// to integrate the gift system with the battle system
type BattleGiftProvider struct {
	giftManager *GiftManager
}

// NewBattleGiftProvider creates a new battle-compatible gift provider
func NewBattleGiftProvider(giftManager *GiftManager) *BattleGiftProvider {
	return &BattleGiftProvider{
		giftManager: giftManager,
	}
}

// GetGiftDefinition retrieves a gift definition by ID for battle integration
func (bgp *BattleGiftProvider) GetGiftDefinition(giftID string) (battle.GiftDefinition, error) {
	bgp.giftManager.mu.RLock()
	defer bgp.giftManager.mu.RUnlock()

	// Get gift from catalog
	gift, exists := bgp.giftManager.giftCatalog[giftID]
	if !exists {
		return battle.GiftDefinition{}, fmt.Errorf("gift not found: %s", giftID)
	}

	// Convert to battle-compatible format
	battleGift := battle.GiftDefinition{
		ID:   gift.ID,
		Name: gift.Name,
		BattleEffect: battle.BattleItemEffect{
			ActionType:      gift.GiftEffects.Battle.ActionType,
			DamageModifier:  gift.GiftEffects.Battle.DamageModifier,
			DefenseModifier: gift.GiftEffects.Battle.DefenseModifier,
			SpeedModifier:   gift.GiftEffects.Battle.SpeedModifier,
			HealModifier:    gift.GiftEffects.Battle.HealModifier,
			Duration:        gift.GiftEffects.Battle.Duration,
			Consumable:      gift.GiftEffects.Battle.Consumable,
		},
	}

	return battleGift, nil
}

// GetAvailableGifts returns all gifts available for battle use
func (bgp *BattleGiftProvider) GetAvailableGifts() []battle.GiftDefinition {
	availableGifts := bgp.giftManager.GetAvailableGifts()

	// Convert to battle-compatible format
	battleGifts := make([]battle.GiftDefinition, 0, len(availableGifts))

	for _, gift := range availableGifts {
		// Only include gifts that have battle effects
		if bgp.hasBattleEffects(gift) {
			battleGift := battle.GiftDefinition{
				ID:   gift.ID,
				Name: gift.Name,
				BattleEffect: battle.BattleItemEffect{
					ActionType:      gift.GiftEffects.Battle.ActionType,
					DamageModifier:  gift.GiftEffects.Battle.DamageModifier,
					DefenseModifier: gift.GiftEffects.Battle.DefenseModifier,
					SpeedModifier:   gift.GiftEffects.Battle.SpeedModifier,
					HealModifier:    gift.GiftEffects.Battle.HealModifier,
					Duration:        gift.GiftEffects.Battle.Duration,
					Consumable:      gift.GiftEffects.Battle.Consumable,
				},
			}
			battleGifts = append(battleGifts, battleGift)
		}
	}

	return battleGifts
}

// hasBattleEffects checks if a gift has any battle-related effects
func (bgp *BattleGiftProvider) hasBattleEffects(gift *GiftDefinition) bool {
	effect := gift.GiftEffects.Battle

	return effect.DamageModifier > 0 ||
		effect.DefenseModifier > 0 ||
		effect.SpeedModifier > 0 ||
		effect.HealModifier > 0 ||
		effect.Duration > 0 ||
		effect.ActionType != ""
}
