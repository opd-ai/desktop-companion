// Package character: Multiplayer battle integration for DDS
// Extends MultiplayerCharacter with battle system support
// WHY: Integrates battle system with existing multiplayer infrastructure while maintaining backward compatibility

package character

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/opd-ai/desktop-companion/internal/battle"
	"github.com/opd-ai/desktop-companion/internal/network"
)

// BattleManager interface to avoid circular imports
type BattleManager interface {
	InitiateBattle(opponentID string) error
	PerformAction(action battle.BattleAction, targetID string) (*battle.BattleResult, error)
	GetBattleState() *battle.BattleState
	GetAvailableActions() []battle.BattleActionType
	EndBattle() error
}

// InitiateBattle starts a battle with another multiplayer character
func (mc *MultiplayerCharacter) InitiateBattle(targetPeerID string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if !mc.networkEnabled {
		return fmt.Errorf("network not enabled")
	}

	// Create battle invite payload
	battleID := generateBattleID()
	payload := network.BattleInvitePayload{
		FromCharacterID: mc.characterID,
		ToCharacterID:   targetPeerID,
		BattleID:        battleID,
		Timestamp:       time.Now(),
	}

	// Store battle ID for tracking (Finding #3 fix)
	mc.currentBattleID = battleID

	// Create and send battle invite message
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal battle invite: %w", err)
	}

	msg := NetworkMessage{
		Type:      "battle_invite",
		From:      mc.characterID,
		To:        targetPeerID,
		Payload:   payloadBytes,
		Timestamp: time.Now(),
	}

	return mc.networkManager.Broadcast(msg)
}

// HandleBattleInvite processes an incoming battle invitation
func (mc *MultiplayerCharacter) HandleBattleInvite(invite network.BattleInvitePayload) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Simulate user acceptance logic (minimal fix for audit)
	accepted := true // TODO: Replace with UI dialog in future
	if !accepted {
		return fmt.Errorf("battle invite declined by user")
	}

	// Store battle ID for accepted invite
	mc.currentBattleID = invite.BattleID

	// Initialize battle manager with participants (Finding #2 fix)
	if mc.battleManager != nil {
		// Initialize battle with participant data
		err := mc.battleManager.InitiateBattle(invite.FromCharacterID)
		if err != nil {
			mc.currentBattleID = "" // Clear on failure
			return fmt.Errorf("failed to initialize battle manager: %w", err)
		}
	}

	return nil
}

// PerformBattleAction sends a battle action to the network
func (mc *MultiplayerCharacter) PerformBattleAction(action battle.BattleAction) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if !mc.networkEnabled {
		return fmt.Errorf("network not enabled")
	}

	// Get current battle ID (Finding #3 fix)
	battleID, err := mc.getCurrentBattleID()
	if err != nil {
		return fmt.Errorf("cannot perform battle action: %w", err)
	}

	// Create battle action payload
	payload := network.BattleActionPayload{
		BattleID:   battleID,
		ActionType: string(action.Type),
		ActorID:    action.ActorID,
		TargetID:   action.TargetID,
		ItemUsed:   action.ItemUsed,
		Timestamp:  time.Now(),
	}

	// Create and send battle action message
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal battle action: %w", err)
	}

	msg := NetworkMessage{
		Type:      "battle_action",
		From:      mc.characterID,
		To:        action.TargetID,
		Payload:   payloadBytes,
		Timestamp: time.Now(),
	}

	return mc.networkManager.Broadcast(msg)
}

// setupBattleHandlers registers battle-specific network message handlers
func (mc *MultiplayerCharacter) setupBattleHandlers() error {
	// Register battle invite handler
	mc.networkManager.RegisterHandler("battle_invite", mc.handleBattleInviteMessage)

	// Register battle action handler
	mc.networkManager.RegisterHandler("battle_action", mc.handleBattleActionMessage)

	// Register battle result handler
	mc.networkManager.RegisterHandler("battle_result", mc.handleBattleResultMessage)

	// Register battle end handler
	mc.networkManager.RegisterHandler("battle_end", mc.handleBattleEndMessage)

	return nil
}

// handleBattleInviteMessage processes incoming battle invite messages
func (mc *MultiplayerCharacter) handleBattleInviteMessage(msg NetworkMessage, peer interface{}) error {
	var payload network.BattleInvitePayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal battle invite payload: %w", err)
	}

	return mc.HandleBattleInvite(payload)
}

// handleBattleActionMessage processes incoming battle action messages
func (mc *MultiplayerCharacter) handleBattleActionMessage(msg NetworkMessage, peer interface{}) error {
	var payload network.BattleActionPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal battle action payload: %w", err)
	}

	// TODO: Forward to battle manager for processing
	// This would require integration with the local battle state

	return nil
}

// handleBattleResultMessage processes incoming battle result messages
func (mc *MultiplayerCharacter) handleBattleResultMessage(msg NetworkMessage, peer interface{}) error {
	var payload network.BattleResultPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal battle result payload: %w", err)
	}

	// TODO: Update local battle state with results
	// This would sync the battle state between peers

	return nil
}

// handleBattleEndMessage processes incoming battle end messages
func (mc *MultiplayerCharacter) handleBattleEndMessage(msg NetworkMessage, peer interface{}) error {
	var payload network.BattleEndPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal battle end payload: %w", err)
	}

	// TODO: Clean up battle state and notify user
	// This would end the battle and return to normal character state

	// Clear battle ID when battle ends (Finding #3 fix)
	mc.mu.Lock()
	mc.currentBattleID = ""
	mc.mu.Unlock()

	return nil
}

// generateBattleID creates a unique battle ID
func generateBattleID() string {
	return fmt.Sprintf("battle_%d", time.Now().UnixNano())
}

// getCurrentBattleID returns the current battle ID or error if no active battle
func (mc *MultiplayerCharacter) getCurrentBattleID() (string, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if mc.currentBattleID == "" {
		return "", fmt.Errorf("no active battle")
	}
	return mc.currentBattleID, nil
}
