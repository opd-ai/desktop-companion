package persistence

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// SaveManager handles game state persistence using JSON files
// Follows the "lazy programmer" approach using only Go standard library
// Provides atomic writes and safe concurrent access to save operations
type SaveManager struct {
	mu             sync.RWMutex
	savePath       string
	autoSave       bool
	interval       time.Duration
	stopChan       chan struct{}
	autoSaveTicker *time.Ticker
}

// GameSaveData represents the complete save state for a character
// This struct is JSON-serializable and contains all persistent game data
type GameSaveData struct {
	CharacterName string         `json:"characterName"`
	SaveVersion   string         `json:"saveVersion"`
	GameState     *GameStateData `json:"gameState"`
	Metadata      *SaveMetadata  `json:"metadata"`
}

// GameStateData represents the core game state that needs persistence
// Mirrors the GameState struct but with JSON-friendly types
type GameStateData struct {
	Stats              map[string]*StatData `json:"stats"`
	LastDecayUpdate    time.Time            `json:"lastDecayUpdate"`
	CreationTime       time.Time            `json:"creationTime"`
	TotalPlayTimeNanos int64                `json:"totalPlayTimeNanos"`
}

// StatData represents a single stat's persistent data
type StatData struct {
	Current           float64 `json:"current"`
	Max               float64 `json:"max"`
	DegradationRate   float64 `json:"degradationRate"`
	CriticalThreshold float64 `json:"criticalThreshold"`
}

// SaveMetadata contains additional save file information
type SaveMetadata struct {
	LastSaved         time.Time      `json:"lastSaved"`
	TotalPlayTime     time.Duration  `json:"totalPlayTime"`
	InteractionCounts map[string]int `json:"interactionCounts"`
	Version           string         `json:"version"`
}

// NewSaveManager creates a new save manager instance
// savePath should be the directory where save files will be stored
func NewSaveManager(savePath string) *SaveManager {
	return &SaveManager{
		savePath: savePath,
		autoSave: false,
		interval: 5 * time.Minute, // Default auto-save interval
		stopChan: make(chan struct{}),
	}
}

// EnableAutoSave starts automatic saving at the specified interval
// This runs in a background goroutine and saves when game state changes
func (sm *SaveManager) EnableAutoSave(interval time.Duration, gameStateProvider func() *GameSaveData) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.prepareAutoSaveState(interval)
	sm.startAutoSaveGoroutine(gameStateProvider)
}

// prepareAutoSaveState configures the auto-save state and ticker
func (sm *SaveManager) prepareAutoSaveState(interval time.Duration) {
	if sm.autoSave {
		sm.disableAutoSaveUnsafe() // Stop existing auto-save
	}

	sm.autoSave = true
	sm.interval = interval
	sm.autoSaveTicker = time.NewTicker(interval)
}

// startAutoSaveGoroutine launches the background auto-save process
func (sm *SaveManager) startAutoSaveGoroutine(gameStateProvider func() *GameSaveData) {
	go func() {
		defer sm.recoverFromPanic()
		sm.runAutoSaveLoop(gameStateProvider)
	}()
}

// runAutoSaveLoop executes the main auto-save loop with ticker and stop channel
func (sm *SaveManager) runAutoSaveLoop(gameStateProvider func() *GameSaveData) {
	for {
		select {
		case <-sm.autoSaveTicker.C:
			sm.performAutoSave(gameStateProvider)
		case <-sm.stopChan:
			return
		}
	}
}

// performAutoSave executes a single auto-save operation if game data is available
func (sm *SaveManager) performAutoSave(gameStateProvider func() *GameSaveData) {
	if gameData := gameStateProvider(); gameData != nil {
		if err := sm.SaveGameState(gameData.CharacterName, gameData); err != nil {
			// Log error but don't crash the application
			// In a production app, this might go to a logger
			_ = err
		}
	}
}

// recoverFromPanic handles any panics in the auto-save goroutine
func (sm *SaveManager) recoverFromPanic() {
	if r := recover(); r != nil {
		// Silently handle panics to prevent crashing the application
		// In production, this would be logged
	}
}

// DisableAutoSave stops the automatic saving goroutine
func (sm *SaveManager) DisableAutoSave() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.disableAutoSaveUnsafe()
}

// disableAutoSaveUnsafe stops auto-save without acquiring the mutex
// Must be called with mutex already held
func (sm *SaveManager) disableAutoSaveUnsafe() {
	if !sm.autoSave {
		return
	}

	sm.autoSave = false
	if sm.autoSaveTicker != nil {
		sm.autoSaveTicker.Stop()
		sm.autoSaveTicker = nil
	}

	// Signal the goroutine to stop (non-blocking)
	select {
	case sm.stopChan <- struct{}{}:
	default:
		// Channel might be full or goroutine already stopped, that's okay
	}
}

// SaveGameState saves game state to a JSON file
// Uses atomic write operation to prevent corruption
func (sm *SaveManager) SaveGameState(characterName string, data *GameSaveData) error {
	if data == nil {
		return fmt.Errorf("save data cannot be nil")
	}

	if characterName == "" {
		return fmt.Errorf("character name cannot be empty")
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Ensure save directory exists
	if err := sm.ensureSaveDirectory(); err != nil {
		return fmt.Errorf("failed to create save directory: %w", err)
	}

	// Generate save file path
	fileName := sm.generateSaveFileName(characterName)
	savePath := filepath.Join(sm.savePath, fileName)

	// Update metadata
	data.Metadata = &SaveMetadata{
		LastSaved:     time.Now(),
		TotalPlayTime: time.Duration(data.GameState.TotalPlayTimeNanos),
		Version:       "1.0",
	}
	data.SaveVersion = "1.0"

	// Perform atomic write
	return sm.atomicWriteJSON(savePath, data)
}

// LoadGameState loads game state from a JSON file
// Returns nil if the save file doesn't exist (new game)
func (sm *SaveManager) LoadGameState(characterName string) (*GameSaveData, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	fileName := sm.generateSaveFileName(characterName)
	savePath := filepath.Join(sm.savePath, fileName)

	// Check if save file exists
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		return nil, nil // No save file means new game
	} else if err != nil {
		return nil, fmt.Errorf("failed to access save file: %w", err)
	}

	// Read and parse JSON
	data, err := os.ReadFile(savePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read save file: %w", err)
	}

	var saveData GameSaveData
	if err := json.Unmarshal(data, &saveData); err != nil {
		return nil, fmt.Errorf("failed to parse save file: %w", err)
	}

	// Validate loaded data
	if err := sm.validateSaveData(&saveData); err != nil {
		return nil, fmt.Errorf("invalid save data: %w", err)
	}

	return &saveData, nil
}

// HasSave checks if a save file exists for the given character
func (sm *SaveManager) HasSave(characterName string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	fileName := sm.generateSaveFileName(characterName)
	savePath := filepath.Join(sm.savePath, fileName)

	_, err := os.Stat(savePath)
	return err == nil
}

// DeleteSave removes a save file for the given character
func (sm *SaveManager) DeleteSave(characterName string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	fileName := sm.generateSaveFileName(characterName)
	savePath := filepath.Join(sm.savePath, fileName)

	if err := os.Remove(savePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete save file: %w", err)
	}

	return nil
}

// GetSaveDirectory returns the directory where save files are stored
func (sm *SaveManager) GetSaveDirectory() string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.savePath
}

// ListSaves returns a list of all available save files
func (sm *SaveManager) ListSaves() ([]string, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	entries, err := os.ReadDir(sm.savePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil // No save directory means no saves
		}
		return nil, fmt.Errorf("failed to read save directory: %w", err)
	}

	var saves []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			// Extract character name from filename
			name := entry.Name()
			name = name[:len(name)-5] // Remove .json extension
			saves = append(saves, name)
		}
	}

	return saves, nil
}

// generateSaveFileName creates a safe filename for a character save
func (sm *SaveManager) generateSaveFileName(characterName string) string {
	// Sanitize character name for filename
	safe := filepath.Base(characterName)
	if safe == "." || safe == ".." {
		safe = "character"
	}
	return safe + ".json"
}

// ensureSaveDirectory creates the save directory if it doesn't exist
func (sm *SaveManager) ensureSaveDirectory() error {
	if _, err := os.Stat(sm.savePath); os.IsNotExist(err) {
		return os.MkdirAll(sm.savePath, 0755)
	}
	return nil
}

// atomicWriteJSON performs an atomic write of JSON data to a file
// This prevents corruption if the write is interrupted
func (sm *SaveManager) atomicWriteJSON(filePath string, data interface{}) error {
	// Write to temporary file first
	tempPath := filePath + ".tmp"

	file, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Pretty-print JSON for readability

	if err := encoder.Encode(data); err != nil {
		file.Close()
		os.Remove(tempPath) // Clean up temporary file
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	if err := file.Close(); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to close temporary file: %w", err)
	}

	// Atomic rename - this is the key to atomic writes
	if err := os.Rename(tempPath, filePath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to rename temporary file: %w", err)
	}

	return nil
}

// validateSaveData ensures loaded save data is valid
func (sm *SaveManager) validateSaveData(data *GameSaveData) error {
	if data.CharacterName == "" {
		return fmt.Errorf("character name cannot be empty")
	}

	if data.GameState == nil {
		return fmt.Errorf("game state cannot be nil")
	}

	if len(data.GameState.Stats) == 0 {
		return fmt.Errorf("game state must have at least one stat")
	}

	// Validate each stat
	for name, stat := range data.GameState.Stats {
		if err := sm.validateStatData(name, stat); err != nil {
			return fmt.Errorf("stat '%s': %w", name, err)
		}
	}

	// Validate time fields
	if data.GameState.CreationTime.IsZero() {
		return fmt.Errorf("creation time cannot be zero")
	}

	if data.GameState.LastDecayUpdate.IsZero() {
		return fmt.Errorf("last decay update cannot be zero")
	}

	return nil
}

// validateStatData validates a single stat's data
func (sm *SaveManager) validateStatData(name string, stat *StatData) error {
	if stat == nil {
		return fmt.Errorf("stat data cannot be nil")
	}

	if stat.Max <= 0 {
		return fmt.Errorf("max value must be positive, got %f", stat.Max)
	}

	if stat.Current < 0 {
		return fmt.Errorf("current value cannot be negative, got %f", stat.Current)
	}

	if stat.Current > stat.Max {
		return fmt.Errorf("current value (%f) cannot exceed max (%f)", stat.Current, stat.Max)
	}

	if stat.DegradationRate < 0 {
		return fmt.Errorf("degradation rate cannot be negative, got %f", stat.DegradationRate)
	}

	if stat.CriticalThreshold < 0 || stat.CriticalThreshold > stat.Max {
		return fmt.Errorf("critical threshold (%f) must be between 0 and max (%f)",
			stat.CriticalThreshold, stat.Max)
	}

	return nil
}

// Close shuts down the save manager and stops auto-save
func (sm *SaveManager) Close() {
	sm.DisableAutoSave()
}
