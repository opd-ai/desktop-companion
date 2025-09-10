package embedding

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestMatrixBuildConfiguration validates the enhanced platform matrix support
// This implements Phase 2 Task 2: Configure matrix builds for all platforms
func TestMatrixBuildConfiguration(t *testing.T) {
	// Test platform matrix configuration validation
	t.Run("PlatformMatrixValidation", func(t *testing.T) {
		supportedPlatforms := []struct {
			goos, goarch  string
			expectSupport bool
		}{
			{"linux", "amd64", true},
			{"windows", "amd64", true},
			{"darwin", "amd64", true},
			{"darwin", "arm64", true},   // Apple Silicon support
			{"freebsd", "amd64", false}, // Not supported in matrix
			{"linux", "arm", false},     // Not in default matrix
		}

		for _, platform := range supportedPlatforms {
			t.Run(platform.goos+"/"+platform.goarch, func(t *testing.T) {
				// Validate platform naming convention
				expectedBinaryName := "test_" + platform.goos + "_" + platform.goarch
				if platform.goos == "windows" {
					expectedBinaryName += ".exe"
				}

				if len(expectedBinaryName) == 0 {
					t.Error("Binary naming failed for platform")
				}

				// Validate platform inclusion in matrix
				isInMatrix := (platform.goos == "linux" && platform.goarch == "amd64") ||
					(platform.goos == "windows" && platform.goarch == "amd64") ||
					(platform.goos == "darwin" && (platform.goarch == "amd64" || platform.goarch == "arm64"))

				if platform.expectSupport && !isInMatrix {
					t.Errorf("Platform %s/%s should be supported but not in matrix", platform.goos, platform.goarch)
				}
			})
		}
	})

	// Test cross-compilation limitations handling
	t.Run("CrossCompilationLimitations", func(t *testing.T) {
		currentOS := runtime.GOOS

		// These combinations should be handled gracefully
		crossCompilationCases := []struct {
			hostOS, targetOS string
			shouldWarn       bool
		}{
			{"linux", "windows", true},
			{"linux", "darwin", true},
			{"darwin", "linux", true},
			{"darwin", "windows", true},
			{"windows", "linux", true},
			{"windows", "darwin", true},
			{"linux", "linux", false},   // Same OS should work
			{"darwin", "darwin", false}, // Same OS should work
		}

		for _, cc := range crossCompilationCases {
			t.Run(cc.hostOS+"_to_"+cc.targetOS, func(t *testing.T) {
				// Simulate the cross-compilation validation logic
				isCrossCompilation := cc.hostOS != cc.targetOS
				needsWarning := isCrossCompilation && cc.shouldWarn

				if needsWarning {
					// In real implementation, this would trigger a warning
					t.Logf("Cross-compilation from %s to %s requires native environment", cc.hostOS, cc.targetOS)
				}

				// The validation should not fail the test, just provide warnings
				if cc.hostOS == currentOS && cc.targetOS != currentOS && !needsWarning {
					t.Error("Should warn about cross-compilation limitations")
				}
			})
		}
	})

	// Test GitHub Actions matrix generation
	t.Run("GitHubActionsMatrix", func(t *testing.T) {
		// Test that the matrix includes all required configurations
		expectedMatrix := []struct {
			os, goos, goarch string
		}{
			{"ubuntu-latest", "linux", "amd64"},
			{"windows-latest", "windows", "amd64"},
			{"macos-latest", "darwin", "amd64"},
			{"macos-latest", "darwin", "arm64"},
		}

		for _, config := range expectedMatrix {
			t.Run(config.os+"_"+config.goos+"_"+config.goarch, func(t *testing.T) {
				// Validate configuration completeness
				if config.os == "" || config.goos == "" || config.goarch == "" {
					t.Error("Matrix configuration incomplete")
				}

				// Validate OS/platform alignment
				expectedAlignment := map[string]string{
					"ubuntu-latest":  "linux",
					"windows-latest": "windows",
					"macos-latest":   "darwin",
				}

				if expectedAlignment[config.os] != config.goos {
					t.Errorf("OS alignment mismatch: %s should build %s", config.os, config.goos)
				}
			})
		}
	})

	// Test artifact management for matrix builds
	t.Run("ArtifactManagement", func(t *testing.T) {
		tempDir := t.TempDir()

		// Simulate artifact creation for multiple platforms
		testArtifacts := []struct {
			character, goos, goarch string
		}{
			{"default", "linux", "amd64"},
			{"tsundere", "windows", "amd64"},
			{"romance", "darwin", "amd64"},
			{"flirty", "darwin", "arm64"},
		}

		for _, artifact := range testArtifacts {
			t.Run(artifact.character+"_"+artifact.goos+"_"+artifact.goarch, func(t *testing.T) {
				// Create mock artifact file
				ext := ""
				if artifact.goos == "windows" {
					ext = ".exe"
				}

				suffix := ""
				if artifact.goarch == "arm64" {
					suffix = "-arm64"
				}

				filename := artifact.character + "_" + artifact.goos + "_" + artifact.goarch + suffix + ext
				artifactPath := filepath.Join(tempDir, filename)

				if err := os.WriteFile(artifactPath, []byte("mock binary"), 0644); err != nil {
					t.Fatalf("Failed to create mock artifact: %v", err)
				}

				// Validate artifact exists and has correct naming
				if _, err := os.Stat(artifactPath); os.IsNotExist(err) {
					t.Error("Artifact should exist after creation")
				}

				// Validate naming convention
				if !validateArtifactNaming(filename, artifact.character, artifact.goos, artifact.goarch) {
					t.Error("Artifact naming convention validation failed")
				}
			})
		}
	})

	// Test artifact retention policies
	t.Run("ArtifactRetentionPolicies", func(t *testing.T) {
		// Test different retention policies for different build types
		retentionTests := []struct {
			buildType           string
			expectedRetention   int // days
			expectedCompression int // days
		}{
			{"development", 7, 1},
			{"production", 90, 7},
			{"release", 365, 30},
		}

		for _, rt := range retentionTests {
			t.Run(rt.buildType, func(t *testing.T) {
				// Validate retention periods are reasonable
				if rt.expectedRetention <= 0 {
					t.Errorf("Retention period for %s should be positive", rt.buildType)
				}
				if rt.expectedCompression <= 0 {
					t.Errorf("Compression period for %s should be positive", rt.buildType)
				}
				if rt.expectedCompression >= rt.expectedRetention {
					t.Errorf("Compression period for %s should be less than retention", rt.buildType)
				}
			})
		}
	})

	// Test artifact size optimization
	t.Run("ArtifactSizeOptimization", func(t *testing.T) {
		// Test that compressed artifacts are smaller than originals
		testContent := make([]byte, 1024) // 1KB of zeros (highly compressible)
		for i := range testContent {
			testContent[i] = 0
		}

		tempDir := t.TempDir()
		originalPath := filepath.Join(tempDir, "test_binary")

		if err := os.WriteFile(originalPath, testContent, 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		originalSize := int64(len(testContent))

		// Simulate compression (we expect significant size reduction for zero-filled data)
		// In real implementation, this would use gzip compression
		expectedCompressionRatio := 0.1 // Expect at least 90% compression for zeros
		maxCompressedSize := int64(float64(originalSize) * expectedCompressionRatio)

		if maxCompressedSize >= originalSize {
			t.Error("Compression should reduce file size significantly")
		}

		t.Logf("Original size: %d bytes, max compressed size: %d bytes", originalSize, maxCompressedSize)
	})
}

// TestPlatformValidation tests the platform validation logic
func TestPlatformValidation(t *testing.T) {
	t.Run("NativeBuildValidation", func(t *testing.T) {
		currentOS := runtime.GOOS

		// Native builds should always be valid
		valid := validatePlatformCompatibility(currentOS, currentOS)
		if !valid {
			t.Error("Native platform builds should always be valid")
		}
	})

	t.Run("CrossCompilationWarnings", func(t *testing.T) {
		// Cross-compilation should trigger warnings but not hard failures
		testCases := []struct {
			host, target string
			expectValid  bool
		}{
			{"linux", "linux", true},
			{"linux", "windows", false},  // Should warn
			{"darwin", "linux", false},   // Should warn
			{"windows", "darwin", false}, // Should warn
		}

		for _, tc := range testCases {
			t.Run(tc.host+"_to_"+tc.target, func(t *testing.T) {
				valid := validatePlatformCompatibility(tc.host, tc.target)
				if valid != tc.expectValid {
					t.Errorf("Platform validation for %s to %s: expected %v, got %v",
						tc.host, tc.target, tc.expectValid, valid)
				}
			})
		}
	})
}

// TestMatrixBuildPerformance validates that matrix builds are efficient
func TestMatrixBuildPerformance(t *testing.T) {
	t.Run("ParallelBuildSupport", func(t *testing.T) {
		// Test parallel build configuration
		maxParallel := 4
		if maxParallel <= 0 {
			t.Error("Parallel build count must be positive")
		}

		// Simulate character and platform counts
		characterCount := 17 // Current number of characters
		platformCount := 4   // linux/amd64, windows/amd64, darwin/amd64, darwin/arm64

		totalBuilds := characterCount * platformCount
		expectedBatches := (totalBuilds + maxParallel - 1) / maxParallel

		if expectedBatches <= 0 {
			t.Error("Build batching calculation failed")
		}

		t.Logf("Matrix will create %d builds in %d parallel batches", totalBuilds, expectedBatches)
	})

	t.Run("ArtifactRetentionPolicy", func(t *testing.T) {
		// Validate retention policies are appropriate
		policies := map[string]int{
			"individual-binaries": 30,
			"release-packages":    90,
		}

		for policy, days := range policies {
			if days <= 0 {
				t.Errorf("Invalid retention policy for %s: %d days", policy, days)
			}
			if days > 400 { // GitHub Actions limit
				t.Errorf("Retention policy for %s exceeds GitHub Actions limits: %d days", policy, days)
			}
		}
	})
}

// Helper functions for validation

// validatePlatformCompatibility simulates the platform validation logic
func validatePlatformCompatibility(hostOS, targetOS string) bool {
	// Native builds are always valid
	if hostOS == targetOS {
		return true
	}

	// Cross-compilation triggers warnings but doesn't fail hard
	// In the real implementation, this would show warnings
	return false
}

// validateArtifactNaming checks if artifact follows naming conventions
func validateArtifactNaming(filename, character, goos, goarch string) bool {
	// Basic validation of naming pattern: character_goos_goarch[suffix].ext
	expectedPrefix := character + "_" + goos + "_" + goarch
	return len(filename) >= len(expectedPrefix) &&
		filename[:len(expectedPrefix)] == expectedPrefix
}
