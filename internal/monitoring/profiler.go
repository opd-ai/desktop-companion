package monitoring

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"
)

// Profiler handles performance monitoring and memory profiling
// Uses Go's built-in runtime/pprof package - following "lazy programmer" principle
type Profiler struct {
	mu             sync.RWMutex
	enabled        bool
	startTime      time.Time
	cpuProfile     *os.File
	ctx            context.Context
	cancel         context.CancelFunc
	stats          *PerformanceStats
	targetMemoryMB int // Target <50MB
}

// PerformanceStats tracks real-time performance metrics
type PerformanceStats struct {
	mu                sync.RWMutex
	StartTime         time.Time     `json:"start_time"`
	CurrentMemoryMB   float64       `json:"current_memory_mb"`
	PeakMemoryMB      float64       `json:"peak_memory_mb"`
	StartupDuration   time.Duration `json:"startup_duration"`
	FrameRate         float64       `json:"frame_rate"`
	LastFrameUpdate   time.Time     `json:"last_frame_update"`
	TotalFrames       uint64        `json:"total_frames"`
	MemoryAllocations uint64        `json:"memory_allocations"`
	GCRuns            uint32        `json:"gc_runs"`
}

// NewProfiler creates a new performance profiler
func NewProfiler(memoryTargetMB int) *Profiler {
	ctx, cancel := context.WithCancel(context.Background())

	return &Profiler{
		enabled:        false,
		targetMemoryMB: memoryTargetMB,
		ctx:            ctx,
		cancel:         cancel,
		stats: &PerformanceStats{
			StartTime: time.Now(),
		},
	}
}

// Start begins performance monitoring and optional file-based profiling
func (p *Profiler) Start(memProfilePath, cpuProfilePath string, debug bool) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.validateStartConditions(); err != nil {
		return err
	}

	// Enable profiler if any profiling is requested OR debug mode is enabled
	// In production with debug=false and no profile paths, no monitoring overhead
	profilingRequested := memProfilePath != "" || cpuProfilePath != "" || debug

	if profilingRequested {
		p.initializeProfiler()

		if err := p.startCPUProfilingIfEnabled(cpuProfilePath, debug); err != nil {
			return err
		}

		p.startMonitoring(debug)
		p.logStartupIfDebug(debug)
	}

	return nil
}

// StartWithMonitoring forces monitoring to be enabled regardless of profiling settings.
// This is primarily for testing purposes where monitoring behavior needs to be validated.
func (p *Profiler) StartWithMonitoring() error {
	return p.Start("", "", true) // Force debug=true to enable monitoring
}

// validateStartConditions checks if profiler can be started
func (p *Profiler) validateStartConditions() error {
	if p.enabled {
		return fmt.Errorf("profiler already started")
	}
	return nil
}

// initializeProfiler sets up initial profiler state
func (p *Profiler) initializeProfiler() {
	p.enabled = true
	p.startTime = time.Now()
}

// startCPUProfilingIfEnabled starts CPU profiling when path is provided
func (p *Profiler) startCPUProfilingIfEnabled(cpuProfilePath string, debug bool) error {
	if cpuProfilePath != "" {
		if err := p.startCPUProfiling(cpuProfilePath); err != nil {
			return fmt.Errorf("failed to start CPU profiling: %w", err)
		}
		if debug {
			log.Printf("CPU profiling started: %s", cpuProfilePath)
		}
	}
	return nil
}

// startMonitoring launches memory and frame rate monitoring goroutines
func (p *Profiler) startMonitoring(debug bool) {
	// Perform initial memory reading to avoid race condition in tests
	p.updateMemoryStats(debug)
	go p.monitorMemory(debug)
	go p.monitorFrameRate(debug)
}

// logStartupIfDebug logs startup message when debug mode is enabled
func (p *Profiler) logStartupIfDebug(debug bool) {
	if debug {
		log.Printf("Performance monitoring started (target: %dMB memory)",
			p.targetMemoryMB)
	}
}

// Stop ends profiling and saves memory profile if enabled
func (p *Profiler) Stop(memProfilePath string, debug bool) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.validateProfilerState(); err != nil {
		return err
	}

	p.shutdownProfiler()

	if err := p.stopCPUProfilingIfActive(debug); err != nil {
		return err
	}

	if err := p.saveMemoryProfileIfRequested(memProfilePath, debug); err != nil {
		return err
	}

	return nil
}

// validateProfilerState checks if profiler is in valid state for stopping
func (p *Profiler) validateProfilerState() error {
	// Allow stopping even if profiler was never started (for clean shutdown)
	return nil
}

// shutdownProfiler performs core profiler shutdown operations
func (p *Profiler) shutdownProfiler() {
	// Cancel context first to stop monitoring goroutines
	if p.cancel != nil {
		p.cancel()
	}

	// Set enabled to false without locking since context cancellation
	// provides the primary shutdown signal
	p.enabled = false
}

// stopCPUProfilingIfActive stops CPU profiling if currently active
func (p *Profiler) stopCPUProfilingIfActive(debug bool) error {
	if p.cpuProfile != nil {
		pprof.StopCPUProfile()
		p.cpuProfile.Close()
		if debug {
			log.Printf("CPU profiling stopped")
		}
	}
	return nil
}

// saveMemoryProfileIfRequested saves memory profile when path is provided
func (p *Profiler) saveMemoryProfileIfRequested(memProfilePath string, debug bool) error {
	if memProfilePath != "" {
		if err := p.saveMemoryProfile(memProfilePath); err != nil {
			return fmt.Errorf("failed to save memory profile: %w", err)
		}
		if debug {
			log.Printf("Memory profile saved: %s", memProfilePath)
		}
	}
	return nil
}

// startCPUProfiling begins CPU profile collection
func (p *Profiler) startCPUProfiling(filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create CPU profile file: %w", err)
	}

	if err := pprof.StartCPUProfile(file); err != nil {
		file.Close()
		return fmt.Errorf("failed to start CPU profiling: %w", err)
	}

	p.cpuProfile = file
	return nil
}

// saveMemoryProfile saves current memory profile to file
func (p *Profiler) saveMemoryProfile(filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create memory profile file: %w", err)
	}
	defer file.Close()

	// Force garbage collection for accurate memory profile
	runtime.GC()

	if err := pprof.WriteHeapProfile(file); err != nil {
		return fmt.Errorf("failed to write memory profile: %w", err)
	}

	return nil
}

// monitorMemory continuously monitors memory usage
func (p *Profiler) monitorMemory(debug bool) {
	ticker := time.NewTicker(1 * time.Second) // Check every second
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.updateMemoryStats(debug)
		}
	}
}

// updateMemoryStats reads current memory statistics
func (p *Profiler) updateMemoryStats(debug bool) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	currentMB := float64(m.Alloc) / 1024 / 1024

	p.stats.mu.Lock()
	p.stats.CurrentMemoryMB = currentMB
	if currentMB > p.stats.PeakMemoryMB {
		p.stats.PeakMemoryMB = currentMB
	}
	p.stats.MemoryAllocations = m.Mallocs
	p.stats.GCRuns = m.NumGC
	p.stats.mu.Unlock()

	// Log warning if memory exceeds target
	if currentMB > float64(p.targetMemoryMB) && debug {
		log.Printf("WARNING: Memory usage %.1fMB exceeds target %dMB",
			currentMB, p.targetMemoryMB)
	}
}

// monitorFrameRate tracks animation frame rate performance
func (p *Profiler) monitorFrameRate(debug bool) {
	ticker := time.NewTicker(5 * time.Second) // Calculate FPS every 5 seconds
	defer ticker.Stop()

	var lastFrameCount uint64

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.calculateFrameRate(lastFrameCount, debug)
			lastFrameCount = p.GetTotalFrames()
		}
	}
}

// calculateFrameRate computes current frame rate
func (p *Profiler) calculateFrameRate(lastFrameCount uint64, debug bool) {
	p.stats.mu.RLock()
	currentFrames := p.stats.TotalFrames
	p.stats.mu.RUnlock()

	framesDelta := currentFrames - lastFrameCount
	fps := float64(framesDelta) / 5.0 // 5-second window

	p.stats.mu.Lock()
	p.stats.FrameRate = fps
	p.stats.mu.Unlock()

	// Log warning if FPS drops below 30
	if fps < 30.0 && fps > 0 && debug {
		log.Printf("WARNING: Frame rate %.1f FPS below target 30 FPS", fps)
	}
}

// RecordFrame should be called each time a frame is rendered
func (p *Profiler) RecordFrame() {
	// Use a non-blocking check to avoid deadlocks during shutdown
	select {
	case <-p.ctx.Done():
		// Context is cancelled, profiler is shutting down
		return
	default:
		// Context still active, continue
	}

	p.mu.RLock()
	enabled := p.enabled
	p.mu.RUnlock()

	if !enabled {
		return
	}

	p.stats.mu.Lock()
	p.stats.TotalFrames++
	p.stats.LastFrameUpdate = time.Now()
	p.stats.mu.Unlock()
}

// RecordStartupComplete marks the end of application startup
func (p *Profiler) RecordStartupComplete() {
	if !p.enabled {
		return
	}

	p.stats.mu.Lock()
	p.stats.StartupDuration = time.Since(p.stats.StartTime)
	p.stats.mu.Unlock()
}

// GetStats returns current performance statistics (thread-safe copy)
func (p *Profiler) GetStats() PerformanceStats {
	p.stats.mu.RLock()
	defer p.stats.mu.RUnlock()

	return PerformanceStats{
		StartTime:         p.stats.StartTime,
		CurrentMemoryMB:   p.stats.CurrentMemoryMB,
		PeakMemoryMB:      p.stats.PeakMemoryMB,
		StartupDuration:   p.stats.StartupDuration,
		FrameRate:         p.stats.FrameRate,
		LastFrameUpdate:   p.stats.LastFrameUpdate,
		TotalFrames:       p.stats.TotalFrames,
		MemoryAllocations: p.stats.MemoryAllocations,
		GCRuns:            p.stats.GCRuns,
	}
}

// GetTotalFrames returns total frames rendered (thread-safe)
func (p *Profiler) GetTotalFrames() uint64 {
	p.stats.mu.RLock()
	defer p.stats.mu.RUnlock()
	return p.stats.TotalFrames
}

// IsMemoryTargetMet checks if memory usage is within target
func (p *Profiler) IsMemoryTargetMet() bool {
	p.stats.mu.RLock()
	defer p.stats.mu.RUnlock()
	return p.stats.CurrentMemoryMB <= float64(p.targetMemoryMB)
}

// IsFrameRateTargetMet checks if frame rate meets performance target
func (p *Profiler) IsFrameRateTargetMet() bool {
	p.stats.mu.RLock()
	defer p.stats.mu.RUnlock()
	return p.stats.FrameRate >= 30.0
}

// GetMemoryUsage returns current memory usage in MB
func (p *Profiler) GetMemoryUsage() float64 {
	p.stats.mu.RLock()
	defer p.stats.mu.RUnlock()
	return p.stats.CurrentMemoryMB
}

// GetStartupTime returns application startup duration
func (p *Profiler) GetStartupTime() time.Duration {
	p.stats.mu.RLock()
	defer p.stats.mu.RUnlock()
	return p.stats.StartupDuration
}
