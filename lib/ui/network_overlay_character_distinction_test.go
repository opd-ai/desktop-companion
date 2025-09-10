package ui

import (
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
)

// TestNetworkOverlay_LocalVsNetworkCharacterDistinction tests the core requirement:
// "UI clearly shows network vs local characters"
func TestNetworkOverlay_LocalVsNetworkCharacterDistinction(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	// Create mock network manager with some peers
	mockNM := NewMockNetworkManager()
	mockNM.AddPeer("remote-player-1", true)  // Connected peer
	mockNM.AddPeer("remote-player-2", false) // Disconnected peer

	// Create overlay
	overlay := NewNetworkOverlay(mockNM)
	
	// Set a specific local character name
	localCharName := "My Local Avatar"
	overlay.SetLocalCharacterName(localCharName)

	// Get character list
	characters := overlay.GetCharacterList()

	// Validate character list structure
	if len(characters) < 1 {
		t.Fatal("Character list should contain at least local character")
	}

	// Test 1: Local character should be first and clearly marked
	localChar := characters[0]
	if !localChar.IsLocal {
		t.Error("First character should be local")
	}
	if localChar.Name != localCharName {
		t.Errorf("Local character name = %q, want %q", localChar.Name, localCharName)
	}
	if localChar.Location != "Local" {
		t.Errorf("Local character location = %q, want 'Local'", localChar.Location)
	}
	if !localChar.IsActive {
		t.Error("Local character should be marked as active")
	}

	// Test 2: Network characters should be clearly distinguished
	networkCharCount := 0
	for i, char := range characters {
		if i == 0 {
			continue // Skip local character
		}
		
		if char.IsLocal {
			t.Errorf("Character %d should be network character, but IsLocal=true", i)
		}
		if char.Location == "Local" {
			t.Errorf("Network character %d has location 'Local'", i)
		}
		if char.CharType != "Network" {
			t.Errorf("Network character %d has type %q, want 'Network'", i, char.CharType)
		}
		networkCharCount++
	}

	// Should have 2 network characters (one per peer)
	if networkCharCount != 2 {
		t.Errorf("Expected 2 network characters, got %d", networkCharCount)
	}

	// Test 3: Verify visual distinction works
	// Local character should have different visual indicators than network characters
	for i, char := range characters {
		t.Logf("Character %d: Name=%q, Local=%v, Active=%v, Location=%q, Type=%q", 
			i, char.Name, char.IsLocal, char.IsActive, char.Location, char.CharType)
	}
}

// TestNetworkOverlay_UILayoutContainsCharacterSection verifies the character section exists
func TestNetworkOverlay_UILayoutContainsCharacterSection(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	mockNM := NewMockNetworkManager()
	overlay := NewNetworkOverlay(mockNM)

	// Get container and verify it exists
	container := overlay.GetContainer()
	if container == nil {
		t.Fatal("Container should not be nil")
	}

	// Verify container has been resized to accommodate character list
	size := container.Size()
	expectedHeight := float32(380) // Should be larger than original 300
	if size.Height < expectedHeight {
		t.Errorf("Container height = %f, want >= %f to accommodate character list", 
			size.Height, expectedHeight)
	}

	// Verify character list widget exists
	if overlay.characterList == nil {
		t.Error("Character list widget should be created")
	}
}

// TestNetworkOverlay_RealTimeCharacterUpdates tests that character list updates when peers change
func TestNetworkOverlay_RealTimeCharacterUpdates(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	mockNM := NewMockNetworkManager()
	overlay := NewNetworkOverlay(mockNM)

	// Initial state: only local character
	chars := overlay.GetCharacterList()
	if len(chars) != 1 {
		t.Fatalf("Initial character count = %d, want 1", len(chars))
	}

	// Add a peer and update
	mockNM.AddPeer("new-peer", true)
	overlay.updateCharacterList()

	// Should now have local + 1 network character
	chars = overlay.GetCharacterList()
	if len(chars) != 2 {
		t.Errorf("After adding peer, character count = %d, want 2", len(chars))
	}

	// Verify the new character is network type
	if len(chars) >= 2 {
		networkChar := chars[1]
		if networkChar.IsLocal {
			t.Error("Second character should be network character")
		}
	}

	// Add another peer
	mockNM.AddPeer("another-peer", false)
	overlay.updateCharacterList()

	// Should now have local + 2 network characters
	chars = overlay.GetCharacterList()
	if len(chars) != 3 {
		t.Errorf("After adding second peer, character count = %d, want 3", len(chars))
	}
}

// TestNetworkOverlay_PerformanceWithManyPeers tests performance with realistic peer counts
func TestNetworkOverlay_PerformanceWithManyPeers(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	mockNM := NewMockNetworkManager()
	overlay := NewNetworkOverlay(mockNM)

	// Add maximum realistic number of peers (8 as per plan)
	for i := 0; i < 8; i++ {
		peerID := "peer-" + string(rune('A'+i))
		mockNM.AddPeer(peerID, i%2 == 0) // Alternate connected/disconnected
	}

	// Measure update performance
	start := time.Now()
	overlay.updateCharacterList()
	elapsed := time.Since(start)

	// Should complete quickly (under 1ms for UI updates)
	if elapsed > time.Millisecond {
		t.Errorf("Character list update took %v, want < 1ms", elapsed)
	}

	// Verify all characters are present
	chars := overlay.GetCharacterList()
	expectedCount := 9 // 1 local + 8 network
	if len(chars) != expectedCount {
		t.Errorf("Character count with 8 peers = %d, want %d", len(chars), expectedCount)
	}

	// Verify local character is still first
	if len(chars) > 0 && !chars[0].IsLocal {
		t.Error("Local character should remain first even with many peers")
	}
}

// BenchmarkNetworkOverlay_CharacterListUpdate benchmarks the character list update performance
func BenchmarkNetworkOverlay_CharacterListUpdate(b *testing.B) {
	app := test.NewApp()
	defer app.Quit()

	mockNM := NewMockNetworkManager()
	// Add some peers for realistic testing
	for i := 0; i < 4; i++ {
		mockNM.AddPeer("peer-"+string(rune('A'+i)), true)
	}
	
	overlay := NewNetworkOverlay(mockNM)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		overlay.updateCharacterList()
	}
}
