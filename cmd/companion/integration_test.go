package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/opd-ai/desktop-companion/internal/character"
	"github.com/opd-ai/desktop-companion/internal/monitoring"
)

// TestMainIntegration tests the complete application integration
func TestMainIntegration(t *testing.T) {
	// Create temporary character for testing
	tmpDir, err := os.MkdirTemp("", "integration_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test character configuration
	characterConfig := `{
		"name": "Test Pet",
		"description": "A test character for integration testing",
		"animations": {
			"idle": "idle.gif",
			"talking": "talking.gif"
		},
		"dialogs": [
			{
				"trigger": "click",
				"responses": ["Hello test!"],
				"animation": "talking",
				"cooldown": 1
			}
		],
		"behavior": {
			"idleTimeout": 10,
			"movementEnabled": false,
			"defaultSize": 64
		}
	}`

	characterPath := filepath.Join(tmpDir, "character.json")
	err = os.WriteFile(characterPath, []byte(characterConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write character config: %v", err)
	}

	// Create dummy GIF files (minimal valid GIF data created with Go's standard library)
	// This is a 1x1 white pixel GIF that decodes correctly
	gifData := []byte{71, 73, 70, 56, 57, 97, 1, 0, 1, 0, 128, 0, 0, 255, 255, 255, 0, 0, 0, 44, 0, 0, 0, 0, 1, 0, 1, 0, 0, 2, 2, 68, 1, 0, 59}

	idlePath := filepath.Join(tmpDir, "idle.gif")
	talkingPath := filepath.Join(tmpDir, "talking.gif")

	err = os.WriteFile(idlePath, gifData, 0644)
	if err != nil {
		t.Fatalf("Failed to write idle.gif: %v", err)
	}

	err = os.WriteFile(talkingPath, gifData, 0644)
	if err != nil {
		t.Fatalf("Failed to write talking.gif: %v", err)
	}

	// Test character loading
	card, err := character.LoadCard(characterPath)
	if err != nil {
		t.Fatalf("Failed to load character card: %v", err)
	}

	if card.Name != "Test Pet" {
		t.Errorf("Expected character name 'Test Pet', got '%s'", card.Name)
	}

	// Test character creation
	char, err := character.New(card, tmpDir)
	if err != nil {
		t.Fatalf("Failed to create character: %v", err)
	}

	if char.GetName() != "Test Pet" {
		t.Errorf("Expected character name 'Test Pet', got '%s'", char.GetName())
	}

	// Test profiler integration
	profiler := monitoring.NewProfiler(50)
	err = profiler.Start("", "", true) // Enable debug mode to enable monitoring
	if err != nil {
		t.Fatalf("Failed to start profiler: %v", err)
	}
	defer profiler.Stop("", false)

	// Simulate application runtime
	profiler.RecordStartupComplete()

	// Simulate frame rendering
	for i := 0; i < 60; i++ { // 1 second at 60 FPS
		char.Update()
		profiler.RecordFrame()
		time.Sleep(16 * time.Millisecond) // ~60 FPS
	}

	// Validate performance metrics
	stats := profiler.GetStats()

	if stats.TotalFrames != 60 {
		t.Errorf("Expected 60 frames, got %d", stats.TotalFrames)
	}

	if stats.StartupDuration == 0 {
		t.Error("Startup duration should be recorded")
	}

	if stats.CurrentMemoryMB <= 0 {
		t.Error("Memory usage should be tracked")
	}

	// Test interaction
	response := char.HandleClick()
	if response == "" {
		t.Error("Character should respond to click")
	}

	if response != "Hello test!" {
		t.Errorf("Expected 'Hello test!', got '%s'", response)
	}
}

// TestPerformanceTargets validates that the application meets performance targets
func TestPerformanceTargets(t *testing.T) {
	profiler := monitoring.NewProfiler(50)
	err := profiler.Start("", "", true) // Enable debug mode to enable monitoring
	if err != nil {
		t.Fatalf("Failed to start profiler: %v", err)
	}
	defer profiler.Stop("", false)

	// Let profiler run for a bit to collect data
	time.Sleep(2 * time.Second)

	// Simulate frame rendering
	startTime := time.Now()
	frameCount := 0
	for time.Since(startTime) < 1*time.Second {
		profiler.RecordFrame()
		frameCount++
		time.Sleep(16 * time.Millisecond) // Target 60 FPS
	}

	stats := profiler.GetStats()

	// Validate memory target (should be well under 50MB for basic operation)
	if stats.CurrentMemoryMB > 50 {
		t.Errorf("Memory usage %f MB exceeds 50MB target", stats.CurrentMemoryMB)
	}

	// Validate that frames are being recorded
	if stats.TotalFrames == 0 {
		t.Error("No frames recorded during test")
	}

	// Frame rate test (should be close to 60 FPS)
	expectedFrames := uint64(frameCount)
	actualFrames := stats.TotalFrames

	// Allow some variance in frame counting
	if actualFrames < expectedFrames-5 || actualFrames > expectedFrames+5 {
		t.Errorf("Frame count mismatch: expected ~%d, got %d", expectedFrames, actualFrames)
	}
}

// TestConcurrentPerformanceMonitoring tests performance monitoring under concurrent load
func TestConcurrentPerformanceMonitoring(t *testing.T) {
	profiler := monitoring.NewProfiler(50)
	err := profiler.Start("", "", true) // Enable debug mode to enable monitoring
	if err != nil {
		t.Fatalf("Failed to start profiler: %v", err)
	}
	defer profiler.Stop("", false)

	// Simulate concurrent frame rendering from multiple goroutines
	done := make(chan bool, 5)

	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				profiler.RecordFrame()
				time.Sleep(1 * time.Millisecond)
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 5; i++ {
		<-done
	}

	stats := profiler.GetStats()

	// Should have recorded 500 frames total
	if stats.TotalFrames != 500 {
		t.Errorf("Expected 500 frames from concurrent access, got %d", stats.TotalFrames)
	}

	// Memory usage should still be reasonable
	if stats.CurrentMemoryMB > 50 {
		t.Errorf("Memory usage %f MB exceeds target under concurrent load", stats.CurrentMemoryMB)
	}
}

// TestProfilerFileOutput tests that profiler can create output files
func TestProfilerFileOutput(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "profiler_output_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cpuProfilePath := filepath.Join(tmpDir, "cpu.prof")
	memProfilePath := filepath.Join(tmpDir, "mem.prof")

	profiler := monitoring.NewProfiler(50)

	// Start with CPU profiling
	err = profiler.Start("", cpuProfilePath, false)
	if err != nil {
		t.Fatalf("Failed to start profiler with CPU profiling: %v", err)
	}

	// Run some workload
	for i := 0; i < 1000; i++ {
		profiler.RecordFrame()
	}

	// Stop with memory profiling
	err = profiler.Stop(memProfilePath, false)
	if err != nil {
		t.Fatalf("Failed to stop profiler with memory profiling: %v", err)
	}

	// Verify files were created
	if _, err := os.Stat(cpuProfilePath); os.IsNotExist(err) {
		t.Error("CPU profile file was not created")
	}

	if _, err := os.Stat(memProfilePath); os.IsNotExist(err) {
		t.Error("Memory profile file was not created")
	}

	// Verify files have content
	cpuInfo, err := os.Stat(cpuProfilePath)
	if err != nil {
		t.Fatalf("Failed to stat CPU profile: %v", err)
	}
	if cpuInfo.Size() == 0 {
		t.Error("CPU profile file is empty")
	}

	memInfo, err := os.Stat(memProfilePath)
	if err != nil {
		t.Fatalf("Failed to stat memory profile: %v", err)
	}
	if memInfo.Size() == 0 {
		t.Error("Memory profile file is empty")
	}
}
