package persistence

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestSaveManagerBasicOperations tests basic save and load functionality
func TestSaveManagerBasicOperations(t *testing.T) {
	// Create temporary directory for tests
	tmpDir, err := os.MkdirTemp("", "save_manager_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sm := NewSaveManager(tmpDir)

	// Create test save data
	testData := &GameSaveData{
		CharacterName: "TestPet",
		SaveVersion:   "1.0",
		GameState: &GameStateData{
			Stats: map[string]*StatData{
				"hunger": {
					Current:           75.5,
					Max:               100.0,
					DegradationRate:   1.0,
					CriticalThreshold: 20.0,
				},
				"happiness": {
					Current:           60.0,
					Max:               100.0,
					DegradationRate:   0.8,
					CriticalThreshold: 15.0,
				},
			},
			LastDecayUpdate:    time.Now().Add(-5 * time.Minute),
			CreationTime:       time.Now().Add(-2 * time.Hour),
			TotalPlayTimeNanos: int64(30 * time.Minute),
		},
	}

	// Test saving
	err = sm.SaveGameState("TestPet", testData)
	if err != nil {
		t.Fatalf("Failed to save game state: %v", err)
	}

	// Verify save file was created
	if !sm.HasSave("TestPet") {
		t.Error("Save file should exist after saving")
	}

	// Test loading
	loadedData, err := sm.LoadGameState("TestPet")
	if err != nil {
		t.Fatalf("Failed to load game state: %v", err)
	}

	if loadedData == nil {
		t.Fatal("Loaded data should not be nil")
	}

	// Verify loaded data matches saved data
	if loadedData.CharacterName != testData.CharacterName {
		t.Errorf("Expected character name %s, got %s",
			testData.CharacterName, loadedData.CharacterName)
	}

	// Verify stats were loaded correctly
	hungerStat := loadedData.GameState.Stats["hunger"]
	if hungerStat == nil {
		t.Fatal("Hunger stat should exist in loaded data")
	}

	if hungerStat.Current != 75.5 {
		t.Errorf("Expected hunger current to be 75.5, got %f", hungerStat.Current)
	}

	if hungerStat.Max != 100.0 {
		t.Errorf("Expected hunger max to be 100.0, got %f", hungerStat.Max)
	}

	// Verify metadata was added
	if loadedData.Metadata == nil {
		t.Error("Metadata should be set after saving")
	}
}

// TestSaveManagerNewGame tests behavior when no save file exists
func TestSaveManagerNewGame(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "save_manager_new_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sm := NewSaveManager(tmpDir)

	// Try to load non-existent save
	loadedData, err := sm.LoadGameState("NonExistentPet")
	if err != nil {
		t.Fatalf("Loading non-existent save should not error: %v", err)
	}

	if loadedData != nil {
		t.Error("Loading non-existent save should return nil")
	}

	// Verify HasSave returns false
	if sm.HasSave("NonExistentPet") {
		t.Error("HasSave should return false for non-existent save")
	}
}

// TestSaveManagerDeleteSave tests save file deletion
func TestSaveManagerDeleteSave(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "save_manager_delete_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sm := NewSaveManager(tmpDir)

	// Create and save test data
	testData := createTestSaveData("DeleteTest")
	err = sm.SaveGameState("DeleteTest", testData)
	if err != nil {
		t.Fatalf("Failed to save game state: %v", err)
	}

	// Verify save exists
	if !sm.HasSave("DeleteTest") {
		t.Error("Save should exist before deletion")
	}

	// Delete save
	err = sm.DeleteSave("DeleteTest")
	if err != nil {
		t.Fatalf("Failed to delete save: %v", err)
	}

	// Verify save no longer exists
	if sm.HasSave("DeleteTest") {
		t.Error("Save should not exist after deletion")
	}

	// Try to delete non-existent save (should not error)
	err = sm.DeleteSave("NonExistent")
	if err != nil {
		t.Errorf("Deleting non-existent save should not error: %v", err)
	}
}

// TestSaveManagerListSaves tests listing available saves
func TestSaveManagerListSaves(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "save_manager_list_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sm := NewSaveManager(tmpDir)

	// Initially should have no saves
	saves, err := sm.ListSaves()
	if err != nil {
		t.Fatalf("Failed to list saves: %v", err)
	}

	if len(saves) != 0 {
		t.Errorf("Expected 0 saves initially, got %d", len(saves))
	}

	// Create multiple saves
	characters := []string{"Pet1", "Pet2", "Pet3"}
	for _, char := range characters {
		testData := createTestSaveData(char)
		err = sm.SaveGameState(char, testData)
		if err != nil {
			t.Fatalf("Failed to save %s: %v", char, err)
		}
	}

	// List saves again
	saves, err = sm.ListSaves()
	if err != nil {
		t.Fatalf("Failed to list saves: %v", err)
	}

	if len(saves) != len(characters) {
		t.Errorf("Expected %d saves, got %d", len(characters), len(saves))
	}

	// Verify all character names are present
	saveMap := make(map[string]bool)
	for _, save := range saves {
		saveMap[save] = true
	}

	for _, char := range characters {
		if !saveMap[char] {
			t.Errorf("Character %s not found in save list", char)
		}
	}
}

// TestSaveManagerValidation tests save data validation
func TestSaveManagerValidation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "save_manager_validation_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sm := NewSaveManager(tmpDir)

	// Test saving nil data
	err = sm.SaveGameState("TestPet", nil)
	if err == nil {
		t.Error("Saving nil data should error")
	}

	// Test invalid save data by creating and manipulating a file directly
	invalidData := &GameSaveData{
		CharacterName: "ValidName", // Valid name for saving
		GameState: &GameStateData{
			Stats: map[string]*StatData{
				"hunger": {Current: 50, Max: 100, DegradationRate: 1.0, CriticalThreshold: 20},
			},
			CreationTime:    time.Now(),
			LastDecayUpdate: time.Now(),
		},
	}

	// This should fail validation during save due to empty character name parameter
	err = sm.SaveGameState("", invalidData)
	if err == nil {
		t.Error("Saving data with empty character name should error")
	}

	// Test loading corrupted save file
	savePath := filepath.Join(tmpDir, "corrupted.json")
	err = os.WriteFile(savePath, []byte("invalid json"), 0o644)
	if err != nil {
		t.Fatalf("Failed to create corrupted file: %v", err)
	}

	_, err = sm.LoadGameState("corrupted")
	if err == nil {
		t.Error("Loading corrupted save file should error")
	}
}

// TestSaveManagerAutoSave tests automatic saving functionality
func TestSaveManagerAutoSave(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "save_manager_autosave_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sm := NewSaveManager(tmpDir)

	// Create test data
	testData := createTestSaveData("AutoSaveTest")
	callCount := 0

	// Mock game state provider
	gameStateProvider := func() *GameSaveData {
		callCount++
		return testData
	}

	// Enable auto-save with short interval for testing
	sm.EnableAutoSave(100*time.Millisecond, gameStateProvider)

	// Wait for at least one auto-save
	time.Sleep(250 * time.Millisecond)

	// Disable auto-save
	sm.DisableAutoSave()

	// Verify auto-save was called
	if callCount == 0 {
		t.Error("Auto-save should have been called at least once")
	}

	// Verify save file was created
	if !sm.HasSave("AutoSaveTest") {
		t.Error("Auto-save should have created save file")
	}

	// Verify we can load the auto-saved data
	loadedData, err := sm.LoadGameState("AutoSaveTest")
	if err != nil {
		t.Fatalf("Failed to load auto-saved data: %v", err)
	}

	if loadedData == nil {
		t.Error("Auto-saved data should not be nil")
	}
}

// TestSaveManagerAtomicWrites tests that saves are atomic and don't corrupt
func TestSaveManagerAtomicWrites(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "save_manager_atomic_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sm := NewSaveManager(tmpDir)

	// Create initial save
	testData := createTestSaveData("AtomicTest")
	err = sm.SaveGameState("AtomicTest", testData)
	if err != nil {
		t.Fatalf("Failed to create initial save: %v", err)
	}

	// Verify no temporary files remain
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read save directory: %v", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".tmp" {
			t.Error("Temporary file should not remain after successful save")
		}
	}

	// Verify save file is valid JSON
	savePath := filepath.Join(tmpDir, "AtomicTest.json")
	data, err := os.ReadFile(savePath)
	if err != nil {
		t.Fatalf("Failed to read save file: %v", err)
	}

	var saveData GameSaveData
	err = json.Unmarshal(data, &saveData)
	if err != nil {
		t.Errorf("Save file should contain valid JSON: %v", err)
	}
}

// TestSaveManagerConcurrency tests concurrent access to save operations
func TestSaveManagerConcurrency(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "save_manager_concurrent_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sm := NewSaveManager(tmpDir)

	// Run multiple concurrent save operations
	numGoroutines := 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			testData := createTestSaveData("ConcurrentTest")
			testData.CharacterName = "Pet" + string(rune('0'+id))

			err := sm.SaveGameState(testData.CharacterName, testData)
			if err != nil {
				t.Errorf("Concurrent save %d failed: %v", id, err)
				return
			}

			// Try to load immediately
			_, err = sm.LoadGameState(testData.CharacterName)
			if err != nil {
				t.Errorf("Concurrent load %d failed: %v", id, err)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify all saves exist
	saves, err := sm.ListSaves()
	if err != nil {
		t.Fatalf("Failed to list saves after concurrent operations: %v", err)
	}

	if len(saves) != numGoroutines {
		t.Errorf("Expected %d saves after concurrent operations, got %d",
			numGoroutines, len(saves))
	}
}

// createTestSaveData creates valid test save data
func createTestSaveData(characterName string) *GameSaveData {
	return &GameSaveData{
		CharacterName: characterName,
		SaveVersion:   "1.0",
		GameState: &GameStateData{
			Stats: map[string]*StatData{
				"hunger": {
					Current:           80.0,
					Max:               100.0,
					DegradationRate:   1.0,
					CriticalThreshold: 20.0,
				},
				"happiness": {
					Current:           65.0,
					Max:               100.0,
					DegradationRate:   0.8,
					CriticalThreshold: 15.0,
				},
			},
			LastDecayUpdate:    time.Now().Add(-3 * time.Minute),
			CreationTime:       time.Now().Add(-1 * time.Hour),
			TotalPlayTimeNanos: int64(45 * time.Minute),
		},
	}
}

// TestSaveManagerDirectoryCreation tests that save directory is created when needed
func TestSaveManagerDirectoryCreation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "save_manager_dir_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Use a subdirectory that doesn't exist yet
	savePath := filepath.Join(tmpDir, "saves", "subdirectory")
	sm := NewSaveManager(savePath)

	// Save should create the directory
	testData := createTestSaveData("DirectoryTest")
	err = sm.SaveGameState("DirectoryTest", testData)
	if err != nil {
		t.Fatalf("Failed to save when directory doesn't exist: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		t.Error("Save directory should have been created")
	}

	// Verify save file exists
	if !sm.HasSave("DirectoryTest") {
		t.Error("Save file should exist after directory creation")
	}
}

// TestSaveManagerFilenameSanitization tests that character names are sanitized for filenames
func TestSaveManagerFilenameSanitization(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "save_manager_sanitize_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sm := NewSaveManager(tmpDir)

	// Test various problematic character names
	problematicNames := []string{
		"Pet with spaces",
		"Pet/with/slashes",
		"Pet:with:colons",
		"Pet*with*asterisks",
		".",
		"..",
	}

	for _, name := range problematicNames {
		testData := createTestSaveData(name)
		testData.CharacterName = name

		err := sm.SaveGameState(name, testData)
		if err != nil {
			t.Errorf("Failed to save character with name '%s': %v", name, err)
			continue
		}

		// Verify we can load it back
		_, err = sm.LoadGameState(name)
		if err != nil {
			t.Errorf("Failed to load character with name '%s': %v", name, err)
		}
	}

	// Test empty name separately (should error)
	testData := createTestSaveData("ValidName")
	err = sm.SaveGameState("", testData)
	if err == nil {
		t.Error("Saving with empty character name should error")
	}
}

// TestSaveManagerCloseMethod tests the Close method functionality
func TestSaveManagerCloseMethod(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "save_manager_close_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sm := NewSaveManager(tmpDir)

	// Enable auto-save
	testData := createTestSaveData("CloseTest")
	gameStateProvider := func() *GameSaveData {
		return testData
	}

	sm.EnableAutoSave(100*time.Millisecond, gameStateProvider)

	// Close should stop auto-save
	sm.Close()

	// Verify auto-save is disabled
	sm.mu.RLock()
	autoSaveEnabled := sm.autoSave
	sm.mu.RUnlock()

	if autoSaveEnabled {
		t.Error("Auto-save should be disabled after Close()")
	}
}

// TestSaveManagerStatValidationEdgeCases tests edge cases in stat validation
func TestSaveManagerStatValidationEdgeCases(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "save_manager_stat_validation_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sm := NewSaveManager(tmpDir)

	// Create save data with invalid stats
	invalidStatsData := &GameSaveData{
		CharacterName: "InvalidStatsTest",
		SaveVersion:   "1.0",
		GameState: &GameStateData{
			Stats: map[string]*StatData{
				"invalid_current": {
					Current:           -10.0, // Invalid: negative
					Max:               100.0,
					DegradationRate:   1.0,
					CriticalThreshold: 20.0,
				},
			},
			LastDecayUpdate: time.Now(),
			CreationTime:    time.Now(),
		},
	}

	// This should pass saving but fail on loading due to validation
	savePath := filepath.Join(tmpDir, "InvalidStatsTest.json")
	data, _ := json.Marshal(invalidStatsData)
	os.WriteFile(savePath, data, 0o644)

	_, err = sm.LoadGameState("InvalidStatsTest")
	if err == nil {
		t.Error("Loading data with invalid stats should error")
	}
}

// TestSaveManagerMoreEdgeCases tests additional edge cases for better coverage
func TestSaveManagerMoreEdgeCases(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "save_manager_edge_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sm := NewSaveManager(tmpDir)

	// Test GetSaveDirectory
	if sm.GetSaveDirectory() != tmpDir {
		t.Errorf("Expected save directory %s, got %s", tmpDir, sm.GetSaveDirectory())
	}

	// Test listing saves in non-existent directory
	newSM := NewSaveManager(filepath.Join(tmpDir, "nonexistent"))
	saves, err := newSM.ListSaves()
	if err != nil {
		t.Errorf("ListSaves should not error for non-existent directory: %v", err)
	}
	if len(saves) != 0 {
		t.Errorf("Expected 0 saves in non-existent directory, got %d", len(saves))
	}

	// Test auto-save with nil game state provider
	sm.EnableAutoSave(50*time.Millisecond, func() *GameSaveData { return nil })
	time.Sleep(100 * time.Millisecond)
	sm.DisableAutoSave()

	// Test enable auto-save twice (should stop the first one)
	sm.EnableAutoSave(50*time.Millisecond, func() *GameSaveData { return createTestSaveData("Test1") })
	sm.EnableAutoSave(50*time.Millisecond, func() *GameSaveData { return createTestSaveData("Test2") })
	sm.DisableAutoSave()

	// Test invalid stat data scenarios - create invalid save file to test loading validation
	invalidSaveData := &GameSaveData{
		CharacterName: "EdgeTest",
		SaveVersion:   "1.0",
		GameState: &GameStateData{
			Stats: map[string]*StatData{
				"invalid": nil, // Nil stat should trigger validation error
			},
			CreationTime:    time.Now(),
			LastDecayUpdate: time.Now(),
		},
	}

	// Write invalid data directly to file
	savePath := filepath.Join(tmpDir, "EdgeTest.json")
	data, _ := json.Marshal(invalidSaveData)
	os.WriteFile(savePath, data, 0o644)

	_, err = sm.LoadGameState("EdgeTest")
	if err == nil {
		t.Error("Loading save with nil stat should error")
	}
}
