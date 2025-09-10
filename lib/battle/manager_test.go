package battle

import (
	"testing"
	"time"
)

// TestBattleManager_NewBattleManager verifies manager initialization
func TestBattleManager_NewBattleManager(t *testing.T) {
	bm := NewBattleManager()
	if bm == nil {
		t.Fatal("NewBattleManager returned nil")
	}
	if bm.currentBattle != nil {
		t.Error("New manager should not have an active battle")
	}
}

// TestBattleManager_InitiateBattle tests battle creation
func TestBattleManager_InitiateBattle(t *testing.T) {
	bm := NewBattleManager()

	err := bm.InitiateBattle("opponent_1")
	if err != nil {
		t.Fatalf("Failed to initiate battle: %v", err)
	}

	state := bm.GetBattleState()
	if state == nil {
		t.Fatal("Battle state should not be nil after initiation")
	}
	if state.Phase != PHASE_WAITING {
		t.Errorf("Expected phase %s, got %s", PHASE_WAITING, state.Phase)
	}
	if state.BattleID == "" {
		t.Error("Battle ID should not be empty")
	}
}

// TestBattleManager_InitiateBattle_AlreadyActive tests preventing multiple battles
func TestBattleManager_InitiateBattle_AlreadyActive(t *testing.T) {
	bm := NewBattleManager()

	// Start first battle
	err := bm.InitiateBattle("opponent_1")
	if err != nil {
		t.Fatalf("Failed to initiate first battle: %v", err)
	}

	// Set battle to active
	bm.mu.Lock()
	bm.currentBattle.Phase = PHASE_ACTIVE
	bm.mu.Unlock()

	// Try to start second battle
	err = bm.InitiateBattle("opponent_2")
	if err == nil {
		t.Error("Should not allow initiating battle when one is already active")
	}
}

// TestBattleManager_AddParticipant tests adding participants to battle
func TestBattleManager_AddParticipant(t *testing.T) {
	bm := NewBattleManager()
	bm.InitiateBattle("opponent_1")

	stats := BattleStats{
		HP:      100,
		MaxHP:   100,
		Attack:  20,
		Defense: 15,
		Speed:   10,
	}

	err := bm.AddParticipant("char_1", "peer_1", true, stats)
	if err != nil {
		t.Fatalf("Failed to add participant: %v", err)
	}

	state := bm.GetBattleState()
	if len(state.Participants) != 1 {
		t.Errorf("Expected 1 participant, got %d", len(state.Participants))
	}

	participant := state.Participants["char_1"]
	if participant == nil {
		t.Fatal("Participant not found")
	}
	if participant.CharacterID != "char_1" {
		t.Errorf("Expected character ID 'char_1', got '%s'", participant.CharacterID)
	}
	if participant.Stats.HP != 100 {
		t.Errorf("Expected HP 100, got %.1f", participant.Stats.HP)
	}
}

// TestBattleManager_AddParticipant_NoBattle tests error when no battle active
func TestBattleManager_AddParticipant_NoBattle(t *testing.T) {
	bm := NewBattleManager()

	stats := BattleStats{HP: 100, MaxHP: 100}
	err := bm.AddParticipant("char_1", "", true, stats)
	if err != ErrBattleNotActive {
		t.Errorf("Expected ErrBattleNotActive, got %v", err)
	}
}

// TestBattleManager_GetAvailableActions tests action availability
func TestBattleManager_GetAvailableActions(t *testing.T) {
	bm := NewBattleManager()
	bm.InitiateBattle("opponent_1")

	// No actions when battle is waiting
	actions := bm.GetAvailableActions()
	if len(actions) != 0 {
		t.Errorf("Expected no actions in waiting phase, got %d", len(actions))
	}

	// Set battle to active
	bm.mu.Lock()
	bm.currentBattle.Phase = PHASE_ACTIVE
	bm.mu.Unlock()

	actions = bm.GetAvailableActions()
	expectedCount := 11 // All action types
	if len(actions) != expectedCount {
		t.Errorf("Expected %d actions in active phase, got %d", expectedCount, len(actions))
	}

	// Verify all expected actions are present
	actionMap := make(map[BattleActionType]bool)
	for _, action := range actions {
		actionMap[action] = true
	}

	expectedActions := []BattleActionType{
		ACTION_ATTACK, ACTION_DEFEND, ACTION_HEAL, ACTION_STUN,
		ACTION_BOOST, ACTION_COUNTER, ACTION_DRAIN, ACTION_SHIELD,
		ACTION_CHARGE, ACTION_EVADE, ACTION_TAUNT,
	}

	for _, expected := range expectedActions {
		if !actionMap[expected] {
			t.Errorf("Missing expected action: %s", expected)
		}
	}
}

// TestBattleManager_EndBattle tests battle termination
func TestBattleManager_EndBattle(t *testing.T) {
	bm := NewBattleManager()
	bm.InitiateBattle("opponent_1")

	err := bm.EndBattle()
	if err != nil {
		t.Fatalf("Failed to end battle: %v", err)
	}

	state := bm.GetBattleState()
	if state.Phase != PHASE_FINISHED {
		t.Errorf("Expected phase %s, got %s", PHASE_FINISHED, state.Phase)
	}
}

// TestBattleManager_EndBattle_NoBattle tests error when no battle to end
func TestBattleManager_EndBattle_NoBattle(t *testing.T) {
	bm := NewBattleManager()

	err := bm.EndBattle()
	if err != ErrBattleNotActive {
		t.Errorf("Expected ErrBattleNotActive, got %v", err)
	}
}

// TestBattleManager_GetCurrentTurnParticipant tests turn tracking
func TestBattleManager_GetCurrentTurnParticipant(t *testing.T) {
	bm := NewBattleManager()
	bm.InitiateBattle("opponent_1")

	// No current participant when no turn order
	current := bm.GetCurrentTurnParticipant()
	if current != "" {
		t.Errorf("Expected empty current participant, got '%s'", current)
	}

	// Add participants to create turn order
	stats := BattleStats{HP: 100, MaxHP: 100}
	bm.AddParticipant("char_1", "", true, stats)
	bm.AddParticipant("char_2", "", false, stats)

	current = bm.GetCurrentTurnParticipant()
	if current != "char_1" {
		t.Errorf("Expected 'char_1' as first participant, got '%s'", current)
	}
}

// TestBattleManager_IsParticipantDefeated tests defeat detection
func TestBattleManager_IsParticipantDefeated(t *testing.T) {
	bm := NewBattleManager()
	bm.InitiateBattle("opponent_1")

	stats := BattleStats{HP: 100, MaxHP: 100}
	bm.AddParticipant("char_1", "", true, stats)

	// Not defeated with positive HP
	if bm.IsParticipantDefeated("char_1") {
		t.Error("Participant should not be defeated with positive HP")
	}

	// Set HP to 0
	bm.mu.Lock()
	bm.currentBattle.Participants["char_1"].Stats.HP = 0
	bm.mu.Unlock()

	// Now should be defeated
	if !bm.IsParticipantDefeated("char_1") {
		t.Error("Participant should be defeated with 0 HP")
	}

	// Non-existent participant should return false
	if bm.IsParticipantDefeated("nonexistent") {
		t.Error("Non-existent participant should not be defeated")
	}
}

// TestBattleManager_GetWinner tests winner determination
func TestBattleManager_GetWinner(t *testing.T) {
	bm := NewBattleManager()
	bm.InitiateBattle("opponent_1")

	// No winner when battle not finished
	winner := bm.GetWinner()
	if winner != "" {
		t.Errorf("Expected no winner in active battle, got '%s'", winner)
	}

	// Add participants
	stats1 := BattleStats{HP: 100, MaxHP: 100}
	stats2 := BattleStats{HP: 0, MaxHP: 100} // Defeated
	bm.AddParticipant("char_1", "", true, stats1)
	bm.AddParticipant("char_2", "", false, stats2)

	// End battle
	bm.EndBattle()

	winner = bm.GetWinner()
	if winner != "char_1" {
		t.Errorf("Expected 'char_1' as winner, got '%s'", winner)
	}
}

// TestBattleState_CopyBattleState tests thread-safe state copying
func TestBattleState_CopyBattleState(t *testing.T) {
	bm := NewBattleManager()
	bm.InitiateBattle("opponent_1")

	stats := BattleStats{
		HP:    100,
		MaxHP: 100,
		Modifiers: []BattleModifier{
			{Type: MODIFIER_DAMAGE, Value: 1.2, Duration: 3, Source: "test"},
		},
	}
	bm.AddParticipant("char_1", "", true, stats)

	original := bm.currentBattle
	copy := bm.copyBattleState(original)

	// Verify copy is separate object
	if copy == original {
		t.Error("Copy should be a different object")
	}

	// Verify values are copied correctly
	if copy.BattleID != original.BattleID {
		t.Error("Battle ID not copied correctly")
	}
	if copy.Phase != original.Phase {
		t.Error("Phase not copied correctly")
	}

	// Verify participants are deeply copied
	if copy.Participants["char_1"] == original.Participants["char_1"] {
		t.Error("Participants should be deeply copied")
	}

	// Modify original and verify copy is unaffected
	original.Phase = PHASE_FINISHED
	if copy.Phase == PHASE_FINISHED {
		t.Error("Copy should not be affected by original modification")
	}
}

// TestBattleManager_ThreadSafety tests concurrent access
func TestBattleManager_ThreadSafety(t *testing.T) {
	bm := NewBattleManager()
	bm.InitiateBattle("opponent_1")

	stats := BattleStats{HP: 100, MaxHP: 100}
	bm.AddParticipant("char_1", "", true, stats)

	done := make(chan bool, 10)

	// Multiple goroutines reading state
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				state := bm.GetBattleState()
				if state == nil {
					t.Error("State should not be nil")
				}
			}
			done <- true
		}()
	}

	// Multiple goroutines modifying state
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				bm.IsParticipantDefeated("char_1")
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("Test timed out - possible deadlock")
		}
	}
}
