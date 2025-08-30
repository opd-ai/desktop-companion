// Package battle implements a fair, turn-based battle system for DDS characters.
//
// This package provides the core battle framework with strict fairness constraints,
// timeout-driven AI decisions, and multiplayer synchronization. It leverages existing
// DDS infrastructure (character stats, networking, gift system) to create an engaging
// battle experience while maintaining backward compatibility.
//
// Design Philosophy:
// - Use standard library for core functionality
// - Leverage existing DDS interfaces for integration
// - Maintain strict fairness with capability caps
// - Support both local AI and multiplayer battles
package battle

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// Battle system configuration constants
// All pets have identical base action values to ensure fairness
const (
	BASE_ATTACK_DAMAGE     = 20.0
	BASE_DEFEND_REDUCTION  = 0.5  // 50% damage reduction
	BASE_HEAL_AMOUNT       = 25.0
	BASE_STUN_DURATION     = 1    // 1 turn
	BASE_BOOST_AMOUNT      = 15.0 // +15 to attack for 3 turns
	BASE_DRAIN_RATIO       = 0.3  // 30% of damage dealt as heal
	BASE_SHIELD_ABSORPTION = 30.0
	BASE_CHARGE_MULTIPLIER = 1.5 // 50% damage boost when charged

	// Fairness constraints - item effects cannot exceed these caps
	MAX_DAMAGE_MODIFIER  = 1.20 // +20% max damage
	MAX_DEFENSE_MODIFIER = 1.15 // +15% max defense
	MAX_SPEED_MODIFIER   = 1.10 // +10% max speed
	MAX_HEAL_MODIFIER    = 1.25 // +25% max healing
	MAX_EFFECT_STACKING  = 3    // Maximum 3 item effects

	// Turn timing
	DEFAULT_TURN_TIMEOUT = 30 * time.Second
	AI_EMERGENCY_TIMEOUT = 5 * time.Second
)

// Common battle system errors
var (
	ErrBattleNotActive     = errors.New("battle is not currently active")
	ErrInvalidParticipant  = errors.New("participant not found in battle")
	ErrActionNotAllowed    = errors.New("action not allowed in current state")
	ErrTurnTimeout         = errors.New("turn timed out")
	ErrMaxModifiersReached = errors.New("maximum number of active modifiers reached")
	ErrIllegalAction       = errors.New("action violates battle rules")
)

// BattleActionType defines the available battle actions
type BattleActionType string

const (
	ACTION_ATTACK  BattleActionType = "attack"
	ACTION_DEFEND  BattleActionType = "defend"
	ACTION_STUN    BattleActionType = "stun"
	ACTION_HEAL    BattleActionType = "heal"
	ACTION_BOOST   BattleActionType = "boost"
	ACTION_COUNTER BattleActionType = "counter"
	ACTION_DRAIN   BattleActionType = "drain"
	ACTION_SHIELD  BattleActionType = "shield"
	ACTION_CHARGE  BattleActionType = "charge"
	ACTION_EVADE   BattleActionType = "evade"
	ACTION_TAUNT   BattleActionType = "taunt"
)

// BattlePhase represents the current state of a battle
type BattlePhase string

const (
	PHASE_WAITING   BattlePhase = "waiting"   // Waiting for participants
	PHASE_ACTIVE    BattlePhase = "active"    // Battle in progress
	PHASE_FINISHED  BattlePhase = "finished" // Battle completed
	PHASE_CANCELLED BattlePhase = "cancelled" // Battle cancelled
)

// ModifierType categorizes different types of battle modifiers
type ModifierType string

const (
	MODIFIER_DAMAGE  ModifierType = "damage"
	MODIFIER_DEFENSE ModifierType = "defense"
	MODIFIER_SPEED   ModifierType = "speed"
	MODIFIER_HEALING ModifierType = "healing"
	MODIFIER_STUN    ModifierType = "stun"
	MODIFIER_SHIELD  ModifierType = "shield"
)

// BattleModifier represents a temporary effect applied to a participant
type BattleModifier struct {
	Type     ModifierType `json:"type"`
	Value    float64      `json:"value"`
	Duration int          `json:"duration"` // Turns remaining
	Source   string       `json:"source"`   // Item/action that created it
}

// BattleStats represents a participant's combat statistics
type BattleStats struct {
	HP        float64          `json:"hp"`        // Current hit points
	MaxHP     float64          `json:"maxHP"`     // Maximum hit points
	Attack    float64          `json:"attack"`    // Base attack power
	Defense   float64          `json:"defense"`   // Base defense rating
	Speed     float64          `json:"speed"`     // Turn order priority
	Modifiers []BattleModifier `json:"modifiers"` // Active modifiers
}

// BattleAction represents a single action taken during battle
type BattleAction struct {
	Type      BattleActionType `json:"type"`
	ActorID   string           `json:"actorID"`
	TargetID  string           `json:"targetID"`
	ItemUsed  string           `json:"itemUsed,omitempty"` // Optional item enhancement
	Timestamp time.Time        `json:"timestamp"`
	Result    *BattleResult    `json:"result,omitempty"`
}

// BattleResult contains the outcome of a battle action
type BattleResult struct {
	Success          bool             `json:"success"`
	Damage           float64          `json:"damage"`
	Healing          float64          `json:"healing"`
	StatusEffects    []string         `json:"statusEffects"`
	Animation        string           `json:"animation"`
	Response         string           `json:"response"`
	ModifiersApplied []BattleModifier `json:"modifiersApplied"`
}

// BattleParticipant represents a character participating in battle
type BattleParticipant struct {
	CharacterID    string         `json:"characterID"`
	PeerID         string         `json:"peerID,omitempty"` // For multiplayer battles
	IsLocal        bool           `json:"isLocal"`
	Stats          BattleStats    `json:"stats"`
	ActiveItems    []string       `json:"activeItems"`    // Currently equipped items
	ActionHistory  []BattleAction `json:"actionHistory"`  // Previous actions
	LastActionTime time.Time      `json:"lastActionTime"`
	IsReady        bool           `json:"isReady"`
}

// BattleState maintains the complete state of an active battle
type BattleState struct {
	BattleID     string                       `json:"battleID"`
	Participants map[string]*BattleParticipant `json:"participants"`
	TurnOrder    []string                     `json:"turnOrder"`
	CurrentTurn  int                          `json:"currentTurn"`
	Phase        BattlePhase                  `json:"phase"`
	TurnTimeout  time.Duration                `json:"turnTimeout"`
	Started      time.Time                    `json:"started"`
	LastAction   *BattleAction                `json:"lastAction,omitempty"`
	mu           sync.RWMutex                 // Protects concurrent access
}

// BattleManager handles battle state management and coordination
type BattleManager struct {
	currentBattle *BattleState
	mu            sync.RWMutex
	// Will be extended with dependencies in future PRs
}

// NewBattleManager creates a new battle manager instance
func NewBattleManager() *BattleManager {
	return &BattleManager{}
}

// InitiateBattle starts a new battle between participants
func (bm *BattleManager) InitiateBattle(opponentID string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if bm.currentBattle != nil && bm.currentBattle.Phase == PHASE_ACTIVE {
		return errors.New("battle already in progress")
	}

	battleID := fmt.Sprintf("battle_%d", time.Now().Unix())
	bm.currentBattle = &BattleState{
		BattleID:     battleID,
		Participants: make(map[string]*BattleParticipant),
		Phase:        PHASE_WAITING,
		TurnTimeout:  DEFAULT_TURN_TIMEOUT,
		Started:      time.Now(),
	}

	return nil
}

// GetBattleState returns the current battle state (thread-safe)
func (bm *BattleManager) GetBattleState() *BattleState {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	if bm.currentBattle == nil {
		return nil
	}

	// Return a copy to prevent external modification
	return bm.copyBattleState(bm.currentBattle)
}

// GetAvailableActions returns actions available to the current turn participant
func (bm *BattleManager) GetAvailableActions() []BattleActionType {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	if bm.currentBattle == nil || bm.currentBattle.Phase != PHASE_ACTIVE {
		return nil
	}

	// All participants have access to the same base actions (fairness)
	return []BattleActionType{
		ACTION_ATTACK, ACTION_DEFEND, ACTION_HEAL,
		ACTION_STUN, ACTION_BOOST, ACTION_COUNTER,
		ACTION_DRAIN, ACTION_SHIELD, ACTION_CHARGE,
		ACTION_EVADE, ACTION_TAUNT,
	}
}

// EndBattle terminates the current battle
func (bm *BattleManager) EndBattle() error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if bm.currentBattle == nil {
		return ErrBattleNotActive
	}

	bm.currentBattle.Phase = PHASE_FINISHED
	return nil
}

// copyBattleState creates a deep copy of battle state for thread safety
func (bm *BattleManager) copyBattleState(original *BattleState) *BattleState {
	stateCopy := &BattleState{
		BattleID:    original.BattleID,
		TurnOrder:   make([]string, len(original.TurnOrder)),
		CurrentTurn: original.CurrentTurn,
		Phase:       original.Phase,
		TurnTimeout: original.TurnTimeout,
		Started:     original.Started,
	}

	copy(stateCopy.TurnOrder, original.TurnOrder)

	stateCopy.Participants = make(map[string]*BattleParticipant)
	for id, participant := range original.Participants {
		stateCopy.Participants[id] = &BattleParticipant{
			CharacterID:    participant.CharacterID,
			PeerID:         participant.PeerID,
			IsLocal:        participant.IsLocal,
			Stats:          participant.Stats,
			ActiveItems:    make([]string, len(participant.ActiveItems)),
			ActionHistory:  make([]BattleAction, len(participant.ActionHistory)),
			LastActionTime: participant.LastActionTime,
			IsReady:        participant.IsReady,
		}
		copy(stateCopy.Participants[id].ActiveItems, participant.ActiveItems)
		copy(stateCopy.Participants[id].ActionHistory, participant.ActionHistory)
	}

	if original.LastAction != nil {
		stateCopy.LastAction = &BattleAction{
			Type:      original.LastAction.Type,
			ActorID:   original.LastAction.ActorID,
			TargetID:  original.LastAction.TargetID,
			ItemUsed:  original.LastAction.ItemUsed,
			Timestamp: original.LastAction.Timestamp,
		}
		if original.LastAction.Result != nil {
			stateCopy.LastAction.Result = &BattleResult{
				Success:       original.LastAction.Result.Success,
				Damage:        original.LastAction.Result.Damage,
				Healing:       original.LastAction.Result.Healing,
				StatusEffects: make([]string, len(original.LastAction.Result.StatusEffects)),
				Animation:     original.LastAction.Result.Animation,
				Response:      original.LastAction.Result.Response,
			}
			copy(stateCopy.LastAction.Result.StatusEffects, original.LastAction.Result.StatusEffects)
		}
	}

	return stateCopy
}
