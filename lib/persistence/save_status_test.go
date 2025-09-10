package persistence

import (
	"sync"
	"testing"
	"time"
)

func TestSaveManager_SetStatusCallback(t *testing.T) {
	sm := NewSaveManager(t.TempDir())

	// Test setting callback
	var callbackCalled bool
	callback := func(status SaveStatus, message string) {
		callbackCalled = true
	}

	sm.SetStatusCallback(callback)

	// Verify callback is set by triggering a save operation
	testData := &GameSaveData{
		CharacterName: "test",
		SaveVersion:   "1.0",
		GameState:     &GameStateData{},
	}

	err := sm.SaveGameState("test", testData)
	if err != nil {
		t.Fatalf("SaveGameState failed: %v", err)
	}

	if !callbackCalled {
		t.Error("Expected status callback to be called")
	}
}

func TestSaveManager_StatusCallback_Nil(t *testing.T) {
	sm := NewSaveManager(t.TempDir())

	// Test with nil callback (should not crash)
	sm.SetStatusCallback(nil)

	testData := &GameSaveData{
		CharacterName: "test",
		SaveVersion:   "1.0",
		GameState:     &GameStateData{},
	}

	err := sm.SaveGameState("test", testData)
	if err != nil {
		t.Fatalf("SaveGameState failed: %v", err)
	}
}

func TestSaveManager_StatusCallback_StatusTransitions(t *testing.T) {
	sm := NewSaveManager(t.TempDir())

	var statusHistory []SaveStatus
	var messageHistory []string
	var mu sync.Mutex

	callback := func(status SaveStatus, message string) {
		mu.Lock()
		defer mu.Unlock()
		statusHistory = append(statusHistory, status)
		messageHistory = append(messageHistory, message)
	}

	sm.SetStatusCallback(callback)

	testData := &GameSaveData{
		CharacterName: "test",
		SaveVersion:   "1.0",
		GameState:     &GameStateData{},
	}

	err := sm.SaveGameState("test", testData)
	if err != nil {
		t.Fatalf("SaveGameState failed: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	// Verify we got the expected status sequence
	if len(statusHistory) < 2 {
		t.Fatalf("Expected at least 2 status updates, got %d", len(statusHistory))
	}

	if statusHistory[0] != SaveStatusSaving {
		t.Errorf("Expected first status to be SaveStatusSaving, got %v", statusHistory[0])
	}

	if statusHistory[len(statusHistory)-1] != SaveStatusSaved {
		t.Errorf("Expected last status to be SaveStatusSaved, got %v", statusHistory[len(statusHistory)-1])
	}

	// First message should be empty (saving), last should be empty (saved)
	if messageHistory[0] != "" {
		t.Errorf("Expected saving message to be empty, got %q", messageHistory[0])
	}

	if messageHistory[len(messageHistory)-1] != "" {
		t.Errorf("Expected saved message to be empty, got %q", messageHistory[len(messageHistory)-1])
	}
}

func TestSaveManager_StatusCallback_SaveError(t *testing.T) {
	// Create save manager with invalid path to trigger error
	sm := NewSaveManager("/dev/null/invalid/path")

	var lastStatus SaveStatus
	var lastMessage string

	callback := func(status SaveStatus, message string) {
		lastStatus = status
		lastMessage = message
	}

	sm.SetStatusCallback(callback)

	testData := &GameSaveData{
		CharacterName: "test",
		SaveVersion:   "1.0",
		GameState:     &GameStateData{},
	}

	err := sm.SaveGameState("test", testData)
	if err == nil {
		t.Fatal("Expected SaveGameState to fail with invalid path")
	}

	// Verify error status was reported
	if lastStatus != SaveStatusError {
		t.Errorf("Expected final status to be SaveStatusError, got %v", lastStatus)
	}

	if lastMessage == "" {
		t.Error("Expected error message to be non-empty")
	}
}

func TestSaveManager_StatusCallback_ThreadSafety(t *testing.T) {
	sm := NewSaveManager(t.TempDir())

	var callCount int
	var mu sync.Mutex

	callback := func(status SaveStatus, message string) {
		mu.Lock()
		callCount++
		mu.Unlock()
	}

	sm.SetStatusCallback(callback)

	// Perform multiple concurrent saves
	var wg sync.WaitGroup
	numGoroutines := 5

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			data := &GameSaveData{
				CharacterName: "test",
				SaveVersion:   "1.0",
				GameState:     &GameStateData{},
			}

			sm.SaveGameState("test", data)
		}(i)
	}

	wg.Wait()

	mu.Lock()
	defer mu.Unlock()

	// Should have at least 2 calls per save operation (saving + saved)
	if callCount < numGoroutines*2 {
		t.Errorf("Expected at least %d status callback calls, got %d", numGoroutines*2, callCount)
	}
}

func TestSaveManager_StatusCallback_CallbackReplacement(t *testing.T) {
	sm := NewSaveManager(t.TempDir())

	var firstCallbackCalled bool
	var secondCallbackCalled bool

	firstCallback := func(status SaveStatus, message string) {
		firstCallbackCalled = true
	}

	secondCallback := func(status SaveStatus, message string) {
		secondCallbackCalled = true
	}

	// Set first callback
	sm.SetStatusCallback(firstCallback)

	testData := &GameSaveData{
		CharacterName: "test",
		SaveVersion:   "1.0",
		GameState:     &GameStateData{},
	}

	err := sm.SaveGameState("test", testData)
	if err != nil {
		t.Fatalf("SaveGameState failed: %v", err)
	}

	if !firstCallbackCalled {
		t.Error("Expected first callback to be called")
	}

	// Replace with second callback
	sm.SetStatusCallback(secondCallback)

	err = sm.SaveGameState("test", testData)
	if err != nil {
		t.Fatalf("SaveGameState failed: %v", err)
	}

	if !secondCallbackCalled {
		t.Error("Expected second callback to be called after replacement")
	}
}

func TestSaveManager_NotifyStatus_ThreadSafety(t *testing.T) {
	sm := NewSaveManager(t.TempDir())

	var callbackExecuted sync.WaitGroup
	numGoroutines := 10
	callbackExecuted.Add(numGoroutines)

	// Set a callback that takes some time
	sm.SetStatusCallback(func(status SaveStatus, message string) {
		time.Sleep(10 * time.Millisecond)
		callbackExecuted.Done()
	})

	// Call notifyStatus from multiple goroutines
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sm.notifyStatus(SaveStatusSaving, "")
		}()
	}

	// Wait for one callback to execute
	callbackExecuted.Wait()

	// This should not hang due to proper lock handling
	wg.Wait()
}

func TestSaveStatus_Constants(t *testing.T) {
	// Test that status constants have expected values
	if SaveStatusIdle != 0 {
		t.Errorf("Expected SaveStatusIdle to be 0, got %d", SaveStatusIdle)
	}

	if SaveStatusSaving != 1 {
		t.Errorf("Expected SaveStatusSaving to be 1, got %d", SaveStatusSaving)
	}

	if SaveStatusSaved != 2 {
		t.Errorf("Expected SaveStatusSaved to be 2, got %d", SaveStatusSaved)
	}

	if SaveStatusError != 3 {
		t.Errorf("Expected SaveStatusError to be 3, got %d", SaveStatusError)
	}
}
