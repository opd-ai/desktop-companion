package monitoring

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestNewProfiler verifies profiler initialization
func TestNewProfiler(t *testing.T) {
	profiler := NewProfiler(50)

	if profiler == nil {
		t.Fatal("NewProfiler returned nil")
	}

	if profiler.targetMemoryMB != 50 {
		t.Errorf("Expected memory target 50MB, got %d", profiler.targetMemoryMB)
	}

	if profiler.enabled {
		t.Error("Profiler should not be enabled initially")
	}
}

// TestStartStopProfiler tests basic profiler lifecycle
func TestStartStopProfiler(t *testing.T) {
	profiler := NewProfiler(50)

	// Test start
	err := profiler.Start("", "", false)
	if err != nil {
		t.Fatalf("Failed to start profiler: %v", err)
	}

	if !profiler.enabled {
		t.Error("Profiler should be enabled after Start()")
	}

	// Wait briefly for monitoring to start
	time.Sleep(100 * time.Millisecond)

	// Test stop
	err = profiler.Stop("", false)
	if err != nil {
		t.Fatalf("Failed to stop profiler: %v", err)
	}

	if profiler.enabled {
		t.Error("Profiler should be disabled after Stop()")
	}
}

// TestCPUProfiling tests CPU profile file creation and cleanup
func TestCPUProfiling(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "profiler_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cpuProfilePath := filepath.Join(tmpDir, "cpu.prof")
	profiler := NewProfiler(50)

	// Start with CPU profiling
	err = profiler.Start("", cpuProfilePath, false)
	if err != nil {
		t.Fatalf("Failed to start profiler with CPU profiling: %v", err)
	}

	// Let it run briefly
	time.Sleep(100 * time.Millisecond)

	// Stop profiling
	err = profiler.Stop("", false)
	if err != nil {
		t.Fatalf("Failed to stop profiler: %v", err)
	}

	// Check if CPU profile file was created
	if _, err := os.Stat(cpuProfilePath); os.IsNotExist(err) {
		t.Error("CPU profile file was not created")
	}
}

// TestMemoryProfiling tests memory profile creation
func TestMemoryProfiling(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "profiler_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	memProfilePath := filepath.Join(tmpDir, "mem.prof")
	profiler := NewProfiler(50)

	// Start profiler
	err = profiler.Start("", "", false)
	if err != nil {
		t.Fatalf("Failed to start profiler: %v", err)
	}

	// Let it run briefly
	time.Sleep(100 * time.Millisecond)

	// Stop with memory profiling
	err = profiler.Stop(memProfilePath, false)
	if err != nil {
		t.Fatalf("Failed to stop profiler with memory profiling: %v", err)
	}

	// Check if memory profile file was created
	if _, err := os.Stat(memProfilePath); os.IsNotExist(err) {
		t.Error("Memory profile file was not created")
	}
}

// TestRecordFrame tests frame recording functionality
func TestRecordFrame(t *testing.T) {
	profiler := NewProfiler(50)

	// Start profiler
	err := profiler.Start("", "", false)
	if err != nil {
		t.Fatalf("Failed to start profiler: %v", err)
	}
	defer profiler.Stop("", false)

	// Record some frames
	initialFrames := profiler.GetTotalFrames()
	profiler.RecordFrame()
	profiler.RecordFrame()
	profiler.RecordFrame()

	finalFrames := profiler.GetTotalFrames()
	expectedFrames := initialFrames + 3

	if finalFrames != expectedFrames {
		t.Errorf("Expected %d frames, got %d", expectedFrames, finalFrames)
	}
}

// TestStartupRecording tests startup time recording
func TestStartupRecording(t *testing.T) {
	profiler := NewProfiler(50)

	// Start profiler
	err := profiler.Start("", "", false)
	if err != nil {
		t.Fatalf("Failed to start profiler: %v", err)
	}
	defer profiler.Stop("", false)

	// Simulate startup completion
	time.Sleep(50 * time.Millisecond)
	profiler.RecordStartupComplete()

	stats := profiler.GetStats()
	if stats.StartupDuration == 0 {
		t.Error("Startup duration should be recorded")
	}

	if stats.StartupDuration < 50*time.Millisecond {
		t.Error("Startup duration seems too short")
	}
}

// TestMemoryMonitoring tests memory usage tracking
func TestMemoryMonitoring(t *testing.T) {
	profiler := NewProfiler(50)

	// Start profiler
	err := profiler.Start("", "", false)
	if err != nil {
		t.Fatalf("Failed to start profiler: %v", err)
	}
	defer profiler.Stop("", false)

	// Wait for at least one memory update
	time.Sleep(1100 * time.Millisecond)

	memoryUsage := profiler.GetMemoryUsage()
	if memoryUsage <= 0 {
		t.Error("Memory usage should be greater than 0")
	}

	stats := profiler.GetStats()
	if stats.CurrentMemoryMB <= 0 {
		t.Error("Current memory usage should be tracked")
	}

	if stats.PeakMemoryMB < stats.CurrentMemoryMB {
		t.Error("Peak memory should be at least current memory")
	}
}

// TestTargetValidation tests performance target checking
func TestTargetValidation(t *testing.T) {
	profiler := NewProfiler(1) // Very low memory target for testing

	// Start profiler
	err := profiler.Start("", "", false)
	if err != nil {
		t.Fatalf("Failed to start profiler: %v", err)
	}
	defer profiler.Stop("", false)

	// Wait for memory monitoring
	time.Sleep(1100 * time.Millisecond)

	// Memory target should not be met (1MB is too low)
	if profiler.IsMemoryTargetMet() {
		t.Error("Memory target should not be met with 1MB limit")
	}
}

// TestConcurrentAccess tests thread safety of stats access
func TestConcurrentAccess(t *testing.T) {
	profiler := NewProfiler(50)

	// Start profiler
	err := profiler.Start("", "", false)
	if err != nil {
		t.Fatalf("Failed to start profiler: %v", err)
	}
	defer profiler.Stop("", false)

	// Simulate concurrent access
	done := make(chan bool, 10)

	// Start multiple goroutines accessing stats
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				profiler.RecordFrame()
				profiler.GetStats()
				profiler.GetMemoryUsage()
				profiler.IsMemoryTargetMet()
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have recorded frames
	if profiler.GetTotalFrames() == 0 {
		t.Error("Should have recorded frames from concurrent access")
	}
}

// BenchmarkRecordFrame benchmarks frame recording performance
func BenchmarkRecordFrame(b *testing.B) {
	profiler := NewProfiler(50)
	profiler.Start("", "", false)
	defer profiler.Stop("", false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		profiler.RecordFrame()
	}
}

// BenchmarkGetStats benchmarks stats retrieval performance
func BenchmarkGetStats(b *testing.B) {
	profiler := NewProfiler(50)
	profiler.Start("", "", false)
	defer profiler.Stop("", false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		profiler.GetStats()
	}
}
