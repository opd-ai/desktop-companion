// Unit tests for battle protocol extensions
// Tests battle message creation and parsing
// WHY: Ensures battle messages are correctly structured and verified

package network

import (
	"encoding/json"
	"testing"
	"time"
)

func TestProtocolManager_CreateBattleInviteMessage(t *testing.T) {
	pm, err := NewProtocolManager()
	if err != nil {
		t.Fatalf("Failed to create protocol manager: %v", err)
	}

	payload := BattleInvitePayload{
		FromCharacterID: "char1",
		ToCharacterID:   "char2",
		BattleID:        "battle123",
		Timestamp:       time.Now(),
	}

	signedMsg, err := pm.CreateBattleInviteMessage("peer1", "peer2", payload)
	if err != nil {
		t.Fatalf("Failed to create battle invite message: %v", err)
	}

	if signedMsg.Message.Type != MessageTypeBattleInvite {
		t.Errorf("Expected message type %s, got %s", MessageTypeBattleInvite, signedMsg.Message.Type)
	}

	if signedMsg.Message.From != "peer1" {
		t.Errorf("Expected from 'peer1', got '%s'", signedMsg.Message.From)
	}

	if signedMsg.Message.To != "peer2" {
		t.Errorf("Expected to 'peer2', got '%s'", signedMsg.Message.To)
	}

	// Verify payload can be unmarshaled
	var unmarshaled BattleInvitePayload
	err = json.Unmarshal(signedMsg.Message.Payload, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}

	if unmarshaled.FromCharacterID != payload.FromCharacterID {
		t.Errorf("Expected FromCharacterID '%s', got '%s'", payload.FromCharacterID, unmarshaled.FromCharacterID)
	}
}

func TestProtocolManager_CreateBattleActionMessage(t *testing.T) {
	pm, err := NewProtocolManager()
	if err != nil {
		t.Fatalf("Failed to create protocol manager: %v", err)
	}

	payload := BattleActionPayload{
		BattleID:   "battle123",
		ActionType: "attack",
		ActorID:    "char1",
		TargetID:   "char2",
		ItemUsed:   "sword",
		Timestamp:  time.Now(),
	}

	signedMsg, err := pm.CreateBattleActionMessage("peer1", "peer2", payload)
	if err != nil {
		t.Fatalf("Failed to create battle action message: %v", err)
	}

	if signedMsg.Message.Type != MessageTypeBattleAction {
		t.Errorf("Expected message type %s, got %s", MessageTypeBattleAction, signedMsg.Message.Type)
	}

	// Verify payload can be unmarshaled
	var unmarshaled BattleActionPayload
	err = json.Unmarshal(signedMsg.Message.Payload, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}

	if unmarshaled.ActionType != payload.ActionType {
		t.Errorf("Expected ActionType '%s', got '%s'", payload.ActionType, unmarshaled.ActionType)
	}
}

func TestProtocolManager_CreateBattleResultMessage(t *testing.T) {
	pm, err := NewProtocolManager()
	if err != nil {
		t.Fatalf("Failed to create protocol manager: %v", err)
	}

	payload := BattleResultPayload{
		BattleID:   "battle123",
		ActionType: "attack",
		ActorID:    "char1",
		TargetID:   "char2",
		Success:    true,
		Damage:     25.0,
		Healing:    0.0,
		Animation:  "attack_animation",
		Response:   "Ouch!",
		ParticipantStats: map[string]BattleParticipantStats{
			"char1": {HP: 100, MaxHP: 100, Attack: 20, Defense: 15, Speed: 10},
			"char2": {HP: 75, MaxHP: 100, Attack: 18, Defense: 12, Speed: 12},
		},
		Timestamp: time.Now(),
	}

	signedMsg, err := pm.CreateBattleResultMessage("peer1", "peer2", payload)
	if err != nil {
		t.Fatalf("Failed to create battle result message: %v", err)
	}

	if signedMsg.Message.Type != MessageTypeBattleResult {
		t.Errorf("Expected message type %s, got %s", MessageTypeBattleResult, signedMsg.Message.Type)
	}

	// Verify payload can be unmarshaled
	var unmarshaled BattleResultPayload
	err = json.Unmarshal(signedMsg.Message.Payload, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}

	if unmarshaled.Damage != payload.Damage {
		t.Errorf("Expected Damage %f, got %f", payload.Damage, unmarshaled.Damage)
	}

	if len(unmarshaled.ParticipantStats) != len(payload.ParticipantStats) {
		t.Errorf("Expected %d participant stats, got %d", len(payload.ParticipantStats), len(unmarshaled.ParticipantStats))
	}
}

func TestProtocolManager_CreateBattleEndMessage(t *testing.T) {
	pm, err := NewProtocolManager()
	if err != nil {
		t.Fatalf("Failed to create protocol manager: %v", err)
	}

	payload := BattleEndPayload{
		BattleID:  "battle123",
		Winner:    "char1",
		Reason:    "defeat",
		Timestamp: time.Now(),
	}

	signedMsg, err := pm.CreateBattleEndMessage("peer1", "peer2", payload)
	if err != nil {
		t.Fatalf("Failed to create battle end message: %v", err)
	}

	if signedMsg.Message.Type != MessageTypeBattleEnd {
		t.Errorf("Expected message type %s, got %s", MessageTypeBattleEnd, signedMsg.Message.Type)
	}

	// Verify payload can be unmarshaled
	var unmarshaled BattleEndPayload
	err = json.Unmarshal(signedMsg.Message.Payload, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}

	if unmarshaled.Winner != payload.Winner {
		t.Errorf("Expected Winner '%s', got '%s'", payload.Winner, unmarshaled.Winner)
	}

	if unmarshaled.Reason != payload.Reason {
		t.Errorf("Expected Reason '%s', got '%s'", payload.Reason, unmarshaled.Reason)
	}
}

func TestBattlePayloads_Serialization(t *testing.T) {
	testCases := []struct {
		name    string
		payload interface{}
	}{
		{
			name: "BattleInvitePayload",
			payload: BattleInvitePayload{
				FromCharacterID: "char1",
				ToCharacterID:   "char2",
				BattleID:        "battle123",
				Timestamp:       time.Now(),
			},
		},
		{
			name: "BattleActionPayload",
			payload: BattleActionPayload{
				BattleID:   "battle123",
				ActionType: "heal",
				ActorID:    "char1",
				TargetID:   "char1",
				ItemUsed:   "potion",
				Timestamp:  time.Now(),
			},
		},
		{
			name: "BattleResultPayload",
			payload: BattleResultPayload{
				BattleID:   "battle123",
				ActionType: "defend",
				ActorID:    "char2",
				TargetID:   "char2",
				Success:    true,
				Damage:     0.0,
				Healing:    15.0,
				Animation:  "defend_animation",
				Response:   "I'm ready!",
				ParticipantStats: map[string]BattleParticipantStats{
					"char2": {HP: 90, MaxHP: 100, Attack: 18, Defense: 20, Speed: 12},
				},
				Timestamp: time.Now(),
			},
		},
		{
			name: "BattleEndPayload",
			payload: BattleEndPayload{
				BattleID:  "battle123",
				Winner:    "",
				Reason:    "draw",
				Timestamp: time.Now(),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Marshal to JSON
			data, err := json.Marshal(tc.payload)
			if err != nil {
				t.Fatalf("Failed to marshal %s: %v", tc.name, err)
			}

			// Unmarshal back
			switch tc.name {
			case "BattleInvitePayload":
				var unmarshaled BattleInvitePayload
				err = json.Unmarshal(data, &unmarshaled)
				if err != nil {
					t.Fatalf("Failed to unmarshal %s: %v", tc.name, err)
				}
			case "BattleActionPayload":
				var unmarshaled BattleActionPayload
				err = json.Unmarshal(data, &unmarshaled)
				if err != nil {
					t.Fatalf("Failed to unmarshal %s: %v", tc.name, err)
				}
			case "BattleResultPayload":
				var unmarshaled BattleResultPayload
				err = json.Unmarshal(data, &unmarshaled)
				if err != nil {
					t.Fatalf("Failed to unmarshal %s: %v", tc.name, err)
				}
			case "BattleEndPayload":
				var unmarshaled BattleEndPayload
				err = json.Unmarshal(data, &unmarshaled)
				if err != nil {
					t.Fatalf("Failed to unmarshal %s: %v", tc.name, err)
				}
			}
		})
	}
}

func TestBattleMessageTypes_Constants(t *testing.T) {
	expectedTypes := map[MessageType]string{
		MessageTypeBattleInvite: "battle_invite",
		MessageTypeBattleAction: "battle_action",
		MessageTypeBattleResult: "battle_result",
		MessageTypeBattleEnd:    "battle_end",
	}

	for msgType, expectedValue := range expectedTypes {
		if string(msgType) != expectedValue {
			t.Errorf("Expected %s = '%s', got '%s'", msgType, expectedValue, string(msgType))
		}
	}
}
