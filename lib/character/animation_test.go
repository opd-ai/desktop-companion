package character

import (
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Helper function to create a test GIF file
func createTestGIF(t *testing.T, filename string, frameCount int, delays []int) string {
	t.Helper()

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "animation_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create simple test images for GIF frames
	images := make([]*image.Paletted, frameCount)
	disposals := make([]byte, frameCount)

	// Use provided delays or default ones
	if delays == nil {
		delays = make([]int, frameCount)
		for i := range delays {
			delays[i] = 10 // 100ms per frame
		}
	}

	for i := 0; i < frameCount; i++ {
		// Create a simple 64x64 paletted image
		img := image.NewPaletted(image.Rect(0, 0, 64, 64), color.Palette{
			color.RGBA{0, 0, 0, 0},     // Transparent
			color.RGBA{255, 0, 0, 255}, // Red
			color.RGBA{0, 255, 0, 255}, // Green
			color.RGBA{0, 0, 255, 255}, // Blue
		})

		// Fill with different color per frame
		colorIndex := byte((i % 3) + 1)
		for y := 0; y < 64; y++ {
			for x := 0; x < 64; x++ {
				img.SetColorIndex(x, y, colorIndex)
			}
		}

		images[i] = img
		disposals[i] = gif.DisposalNone
	}

	// Create GIF structure
	testGIF := &gif.GIF{
		Image:     images,
		Delay:     delays,
		Disposal:  disposals,
		LoopCount: 0, // Infinite loop
	}

	// Write to file
	fullPath := filepath.Join(tempDir, filename)
	file, err := os.Create(fullPath)
	if err != nil {
		t.Fatalf("Failed to create test GIF file: %v", err)
	}
	defer file.Close()

	if err := gif.EncodeAll(file, testGIF); err != nil {
		t.Fatalf("Failed to encode test GIF: %v", err)
	}

	return fullPath
}

func TestNewAnimationManager(t *testing.T) {
	am := NewAnimationManager()

	if am == nil {
		t.Error("NewAnimationManager() returned nil")
	}

	if am.animations == nil {
		t.Error("animations map not initialized")
	}

	if !am.playing {
		t.Error("animation manager should start in playing state")
	}

	if am.currentAnim != "" {
		t.Errorf("currentAnim should be empty initially, got %s", am.currentAnim)
	}

	if am.frameIndex != 0 {
		t.Errorf("frameIndex should be 0 initially, got %d", am.frameIndex)
	}

	// Check that lastUpdate is reasonable (within last second)
	if time.Since(am.lastUpdate) > time.Second {
		t.Error("lastUpdate should be recent")
	}
}

func TestAnimationManager_LoadAnimation(t *testing.T) {
	// Create test GIF files
	tempDir, err := os.MkdirTemp("", "load_animation_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	validGifPath := createTestGIF(t, "valid.gif", 3, []int{10, 20, 30})
	emptyGifPath := filepath.Join(tempDir, "empty.gif")

	// Create an empty GIF file (invalid)
	emptyFile, err := os.Create(emptyGifPath)
	if err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}
	emptyFile.Close()

	// Create a non-GIF file
	nonGifPath := filepath.Join(tempDir, "notgif.txt")
	if err := os.WriteFile(nonGifPath, []byte("not a gif"), 0o644); err != nil {
		t.Fatalf("Failed to create non-GIF file: %v", err)
	}

	am := NewAnimationManager()

	tests := []struct {
		name     string
		animName string
		filepath string
		wantErr  bool
		validate func(t *testing.T, am *AnimationManager)
	}{
		{
			name:     "ValidGIF",
			animName: "test_anim",
			filepath: validGifPath,
			wantErr:  false,
			validate: func(t *testing.T, am *AnimationManager) {
				am.mu.RLock()
				defer am.mu.RUnlock()
				if _, exists := am.animations["test_anim"]; !exists {
					t.Error("Animation should be loaded")
				}
				if am.currentAnim != "test_anim" {
					t.Errorf("Expected currentAnim to be 'test_anim', got '%s'", am.currentAnim)
				}
			},
		},
		{
			name:     "NonExistentFile",
			animName: "missing",
			filepath: "/nonexistent/file.gif",
			wantErr:  true,
			validate: nil,
		},
		{
			name:     "InvalidGIFFile",
			animName: "invalid",
			filepath: nonGifPath,
			wantErr:  true,
			validate: nil,
		},
		{
			name:     "EmptyGIFFile",
			animName: "empty",
			filepath: emptyGifPath,
			wantErr:  true,
			validate: nil,
		},
		{
			name:     "SecondValidGIF",
			animName: "second_anim",
			filepath: validGifPath,
			wantErr:  false,
			validate: func(t *testing.T, am *AnimationManager) {
				am.mu.RLock()
				defer am.mu.RUnlock()
				if len(am.animations) != 2 {
					t.Errorf("Expected 2 animations loaded, got %d", len(am.animations))
				}
				// Current animation should remain the first one
				if am.currentAnim != "test_anim" {
					t.Errorf("Expected currentAnim to remain 'test_anim', got '%s'", am.currentAnim)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := am.LoadAnimation(tt.animName, tt.filepath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadAnimation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.validate != nil && err == nil {
				tt.validate(t, am)
			}
		})
	}
}

func TestAnimationManager_SetCurrentAnimation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "set_current_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	am := NewAnimationManager()

	// Load test animations
	anim1Path := createTestGIF(t, "anim1.gif", 2, nil)
	anim2Path := createTestGIF(t, "anim2.gif", 3, nil)

	if err := am.LoadAnimation("anim1", anim1Path); err != nil {
		t.Fatalf("Failed to load anim1: %v", err)
	}
	if err := am.LoadAnimation("anim2", anim2Path); err != nil {
		t.Fatalf("Failed to load anim2: %v", err)
	}

	tests := []struct {
		name      string
		animName  string
		wantErr   bool
		expectSet bool
	}{
		{
			name:      "SwitchToExistingAnimation",
			animName:  "anim2",
			wantErr:   false,
			expectSet: true,
		},
		{
			name:      "SwitchToNonExistentAnimation",
			animName:  "nonexistent",
			wantErr:   true,
			expectSet: false,
		},
		{
			name:      "SwitchBackToFirstAnimation",
			animName:  "anim1",
			wantErr:   false,
			expectSet: true,
		},
		{
			name:      "EmptyAnimationName",
			animName:  "",
			wantErr:   true,
			expectSet: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			previousAnim := am.GetCurrentAnimationName()
			err := am.SetCurrentAnimation(tt.animName)

			if (err != nil) != tt.wantErr {
				t.Errorf("SetCurrentAnimation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			currentAnim := am.GetCurrentAnimationName()
			if tt.expectSet {
				if currentAnim != tt.animName {
					t.Errorf("Expected current animation '%s', got '%s'", tt.animName, currentAnim)
				}
				// Frame index should be reset
				am.mu.RLock()
				frameIndex := am.frameIndex
				am.mu.RUnlock()
				if frameIndex != 0 {
					t.Errorf("Frame index should be reset to 0, got %d", frameIndex)
				}
			} else {
				// Animation should remain unchanged on error
				if currentAnim != previousAnim {
					t.Errorf("Animation should not change on error, was '%s', now '%s'", previousAnim, currentAnim)
				}
			}
		})
	}
}

func TestAnimationManager_PlaybackControl(t *testing.T) {
	am := NewAnimationManager()

	// Test initial state
	if !am.IsPlaying() {
		t.Error("Animation should be playing by default")
	}

	// Test pause
	am.Pause()
	if am.IsPlaying() {
		t.Error("Animation should be paused after Pause()")
	}

	// Test play
	am.Play()
	if !am.IsPlaying() {
		t.Error("Animation should be playing after Play()")
	}

	// Test multiple pause/play cycles
	for i := 0; i < 3; i++ {
		am.Pause()
		if am.IsPlaying() {
			t.Errorf("Cycle %d: Animation should be paused", i)
		}
		am.Play()
		if !am.IsPlaying() {
			t.Errorf("Cycle %d: Animation should be playing", i)
		}
	}
}

func TestAnimationManager_GetCurrentFrame(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "current_frame_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	am := NewAnimationManager()

	// Test with no animations loaded
	img, hasNewFrame := am.GetCurrentFrame()
	if img != nil {
		t.Error("Should return nil image when no animations loaded")
	}
	if hasNewFrame {
		t.Error("Should not have new frame when no animations loaded")
	}

	// Load test animation
	animPath := createTestGIF(t, "test.gif", 3, []int{1, 1, 1}) // Very fast animation
	if err := am.LoadAnimation("test", animPath); err != nil {
		t.Fatalf("Failed to load animation: %v", err)
	}

	// Test with animation loaded
	img, hasNewFrame = am.GetCurrentFrame()
	if img == nil {
		t.Error("Should return image when animation is loaded")
	}

	// Test paused animation
	am.Pause()
	img, hasNewFrame = am.GetCurrentFrame()
	if img != nil {
		t.Error("Should return nil when animation is paused")
	}
	if hasNewFrame {
		t.Error("Should not have new frame when paused")
	}
}

func TestAnimationManager_GetCurrentFrameImage(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "frame_image_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	am := NewAnimationManager()

	// Test with no animations
	img := am.GetCurrentFrameImage()
	if img != nil {
		t.Error("Should return nil when no animations loaded")
	}

	// Load animation
	animPath := createTestGIF(t, "test.gif", 2, nil)
	if err := am.LoadAnimation("test", animPath); err != nil {
		t.Fatalf("Failed to load animation: %v", err)
	}

	// Test with animation loaded
	img = am.GetCurrentFrameImage()
	if img == nil {
		t.Error("Should return image when animation loaded")
	}

	// Verify image properties
	bounds := img.Bounds()
	if bounds.Dx() != 64 || bounds.Dy() != 64 {
		t.Errorf("Expected 64x64 image, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestAnimationManager_Update(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "update_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	am := NewAnimationManager()

	// Test update with no animations
	if am.Update() {
		t.Error("Update should return false when no animations loaded")
	}

	// Load animation with very short delays for testing
	animPath := createTestGIF(t, "test.gif", 3, []int{1, 1, 1}) // 10ms per frame
	if err := am.LoadAnimation("test", animPath); err != nil {
		t.Fatalf("Failed to load animation: %v", err)
	}

	// Test update with animation
	_ = am.GetCurrentFrameImage() // Get baseline frame

	// Wait for frame advancement opportunity
	time.Sleep(20 * time.Millisecond)

	frameChanged := am.Update()
	if !frameChanged {
		t.Error("Frame should have changed after sufficient time")
	}

	// Test paused animation
	am.Pause()
	if am.Update() {
		t.Error("Paused animation should not update frames")
	}
}

func TestAnimationManager_Reset(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "reset_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	am := NewAnimationManager()

	// Load animation
	animPath := createTestGIF(t, "test.gif", 5, nil)
	if err := am.LoadAnimation("test", animPath); err != nil {
		t.Fatalf("Failed to load animation: %v", err)
	}

	// Advance to a different frame
	am.mu.Lock()
	am.frameIndex = 3
	am.mu.Unlock()

	// Reset
	am.Reset()

	// Verify reset
	am.mu.RLock()
	frameIndex := am.frameIndex
	am.mu.RUnlock()

	if frameIndex != 0 {
		t.Errorf("Frame index should be 0 after reset, got %d", frameIndex)
	}
}

func TestAnimationManager_GetMethods(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "get_methods_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	am := NewAnimationManager()

	// Test empty state
	if len(am.GetLoadedAnimations()) != 0 {
		t.Error("Should have no loaded animations initially")
	}

	if am.GetCurrentAnimationName() != "" {
		t.Error("Current animation name should be empty initially")
	}

	if am.GetAnimationFrameCount("nonexistent") != 0 {
		t.Error("Frame count for non-existent animation should be 0")
	}

	// Load animations
	anim1Path := createTestGIF(t, "anim1.gif", 3, nil)
	anim2Path := createTestGIF(t, "anim2.gif", 5, nil)

	if err := am.LoadAnimation("anim1", anim1Path); err != nil {
		t.Fatalf("Failed to load anim1: %v", err)
	}
	if err := am.LoadAnimation("anim2", anim2Path); err != nil {
		t.Fatalf("Failed to load anim2: %v", err)
	}

	// Test with loaded animations
	loadedAnims := am.GetLoadedAnimations()
	if len(loadedAnims) != 2 {
		t.Errorf("Expected 2 loaded animations, got %d", len(loadedAnims))
	}

	// Verify animation names are present
	foundAnim1, foundAnim2 := false, false
	for _, name := range loadedAnims {
		if name == "anim1" {
			foundAnim1 = true
		}
		if name == "anim2" {
			foundAnim2 = true
		}
	}
	if !foundAnim1 || !foundAnim2 {
		t.Error("Both animation names should be present in loaded animations")
	}

	// Test current animation name
	if am.GetCurrentAnimationName() != "anim1" {
		t.Errorf("Expected current animation 'anim1', got '%s'", am.GetCurrentAnimationName())
	}

	// Test frame counts
	if am.GetAnimationFrameCount("anim1") != 3 {
		t.Errorf("Expected 3 frames for anim1, got %d", am.GetAnimationFrameCount("anim1"))
	}
	if am.GetAnimationFrameCount("anim2") != 5 {
		t.Errorf("Expected 5 frames for anim2, got %d", am.GetAnimationFrameCount("anim2"))
	}
	if am.GetAnimationFrameCount("nonexistent") != 0 {
		t.Error("Non-existent animation should have 0 frames")
	}
}

// Table-driven test for comprehensive animation workflow
func TestAnimationManager_CompleteWorkflow(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "workflow_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	workflowTests := []struct {
		name       string
		frameCount int
		delays     []int
		operations []func(t *testing.T, am *AnimationManager)
	}{
		{
			name:       "BasicWorkflow",
			frameCount: 3,
			delays:     []int{10, 20, 30},
			operations: []func(t *testing.T, am *AnimationManager){
				func(t *testing.T, am *AnimationManager) {
					// Verify initial state
					if !am.IsPlaying() {
						t.Error("Should be playing initially")
					}
				},
				func(t *testing.T, am *AnimationManager) {
					// Test frame retrieval
					img := am.GetCurrentFrameImage()
					if img == nil {
						t.Error("Should have current frame")
					}
				},
				func(t *testing.T, am *AnimationManager) {
					// Test pause/resume
					am.Pause()
					if am.IsPlaying() {
						t.Error("Should be paused")
					}
					am.Play()
					if !am.IsPlaying() {
						t.Error("Should be playing after resume")
					}
				},
			},
		},
		{
			name:       "LongAnimation",
			frameCount: 10,
			delays:     nil, // Use defaults
			operations: []func(t *testing.T, am *AnimationManager){
				func(t *testing.T, am *AnimationManager) {
					// Verify frame count
					if am.GetAnimationFrameCount("test_anim") != 10 {
						t.Error("Should have 10 frames")
					}
				},
				func(t *testing.T, am *AnimationManager) {
					// Test reset functionality
					am.Reset()
					// Should still be at frame 0 after reset
					am.mu.RLock()
					frameIndex := am.frameIndex
					am.mu.RUnlock()
					if frameIndex != 0 {
						t.Error("Should be at frame 0 after reset")
					}
				},
			},
		},
	}

	for _, tt := range workflowTests {
		t.Run(tt.name, func(t *testing.T) {
			am := NewAnimationManager()

			// Create and load animation
			animPath := createTestGIF(t, tt.name+".gif", tt.frameCount, tt.delays)
			animName := "test_anim" // Use consistent name

			if err := am.LoadAnimation(animName, animPath); err != nil {
				t.Fatalf("Failed to load animation: %v", err)
			}

			// Run operations
			for i, operation := range tt.operations {
				t.Run(fmt.Sprintf("Operation%d", i+1), func(t *testing.T) {
					operation(t, am)
				})
			}
		})
	}
}

// Test concurrent access to ensure thread safety
func TestAnimationManager_ConcurrentAccess(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "concurrent_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	am := NewAnimationManager()

	// Load test animation
	animPath := createTestGIF(t, "concurrent.gif", 5, []int{1, 1, 1, 1, 1})
	if err := am.LoadAnimation("test", animPath); err != nil {
		t.Fatalf("Failed to load animation: %v", err)
	}

	// Launch concurrent operations
	done := make(chan bool, 6)

	// Concurrent frame updates
	go func() {
		for i := 0; i < 100; i++ {
			am.Update()
			time.Sleep(time.Millisecond)
		}
		done <- true
	}()

	// Concurrent frame retrieval
	go func() {
		for i := 0; i < 100; i++ {
			am.GetCurrentFrame()
			am.GetCurrentFrameImage()
			time.Sleep(time.Millisecond)
		}
		done <- true
	}()

	// Concurrent playback control
	go func() {
		for i := 0; i < 50; i++ {
			am.Pause()
			time.Sleep(time.Millisecond)
			am.Play()
			time.Sleep(time.Millisecond)
		}
		done <- true
	}()

	// Concurrent information queries
	go func() {
		for i := 0; i < 100; i++ {
			am.IsPlaying()
			am.GetCurrentAnimationName()
			am.GetLoadedAnimations()
			am.GetAnimationFrameCount("test")
			time.Sleep(time.Millisecond)
		}
		done <- true
	}()

	// Concurrent resets
	go func() {
		for i := 0; i < 50; i++ {
			am.Reset()
			time.Sleep(2 * time.Millisecond)
		}
		done <- true
	}()

	// Concurrent animation switching
	go func() {
		for i := 0; i < 50; i++ {
			am.SetCurrentAnimation("test") // Switch to same animation
			time.Sleep(2 * time.Millisecond)
		}
		done <- true
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 6; i++ {
		select {
		case <-done:
			// Good
		case <-time.After(10 * time.Second):
			t.Fatal("Concurrent test timed out")
		}
	}
}

// Helper function for min (available in Go 1.21+)
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
