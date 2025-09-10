package artifact

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestArtifactManager tests the core artifact management functionality
func TestArtifactManager(t *testing.T) {
	tempDir := t.TempDir()

	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create artifact manager: %v", err)
	}

	t.Run("StoreArtifact", func(t *testing.T) {
		// Create a test artifact file
		testFile := filepath.Join(tempDir, "test_binary")
		testContent := []byte("mock binary content")
		if err := os.WriteFile(testFile, testContent, 0o644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Store the artifact
		metadata := map[string]string{
			"version": "1.0.0",
			"type":    "test",
		}

		info, err := manager.StoreArtifact(testFile, "default", "linux", "amd64", metadata)
		if err != nil {
			t.Fatalf("Failed to store artifact: %v", err)
		}

		// Validate artifact info
		if info.Character != "default" {
			t.Errorf("Expected character 'default', got %q", info.Character)
		}
		if info.Platform != "linux" {
			t.Errorf("Expected platform 'linux', got %q", info.Platform)
		}
		if info.Architecture != "amd64" {
			t.Errorf("Expected architecture 'amd64', got %q", info.Architecture)
		}
		if info.Size != int64(len(testContent)) {
			t.Errorf("Expected size %d, got %d", len(testContent), info.Size)
		}
		if info.Checksum == "" {
			t.Error("Checksum should not be empty")
		}
		if info.Metadata["version"] != "1.0.0" {
			t.Errorf("Expected version '1.0.0', got %q", info.Metadata["version"])
		}
	})

	t.Run("ListArtifacts", func(t *testing.T) {
		// Store multiple test artifacts
		testCases := []struct {
			character, platform, arch string
		}{
			{"default", "linux", "amd64"},
			{"tsundere", "windows", "amd64"},
			{"default", "darwin", "amd64"},
		}

		for i, tc := range testCases {
			testFile := filepath.Join(tempDir, "test_binary_"+string(rune('a'+i)))
			if err := os.WriteFile(testFile, []byte("content"), 0o644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			_, err := manager.StoreArtifact(testFile, tc.character, tc.platform, tc.arch, nil)
			if err != nil {
				t.Fatalf("Failed to store artifact: %v", err)
			}
		}

		// List all artifacts for default character
		artifacts, err := manager.ListArtifacts("default", "", "")
		if err != nil {
			t.Fatalf("Failed to list artifacts: %v", err)
		}

		// Should have 2 artifacts for default character
		expectedCount := 2
		if len(artifacts) != expectedCount {
			t.Errorf("Expected %d artifacts for default character, got %d", expectedCount, len(artifacts))
		}

		// List artifacts for specific platform
		artifacts, err = manager.ListArtifacts("default", "linux", "amd64")
		if err != nil {
			t.Fatalf("Failed to list platform-specific artifacts: %v", err)
		}

		if len(artifacts) != 1 {
			t.Errorf("Expected 1 artifact for default/linux/amd64, got %d", len(artifacts))
		}
	})

	t.Run("GetArtifactStats", func(t *testing.T) {
		stats, err := manager.GetArtifactStats()
		if err != nil {
			t.Fatalf("Failed to get artifact stats: %v", err)
		}

		// Validate stats structure
		if stats["total_artifacts"].(int) < 1 {
			t.Error("Expected at least 1 total artifact")
		}
		if stats["total_size"].(int64) <= 0 {
			t.Error("Expected positive total size")
		}

		characters := stats["characters"].(map[string]int)
		if characters["default"] < 1 {
			t.Error("Expected at least 1 artifact for default character")
		}
	})
}

// TestRetentionPolicies tests artifact retention and cleanup functionality
func TestRetentionPolicies(t *testing.T) {
	tempDir := t.TempDir()

	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create artifact manager: %v", err)
	}

	t.Run("DefaultPolicies", func(t *testing.T) {
		policies := DefaultRetentionPolicies()

		// Validate that default policies exist and are reasonable
		expectedPolicies := []string{"development", "production", "release"}
		for _, policyName := range expectedPolicies {
			policy, exists := policies[policyName]
			if !exists {
				t.Errorf("Expected default policy %q not found", policyName)
				continue
			}

			if policy.RetentionPeriod <= 0 {
				t.Errorf("Policy %q has invalid retention period: %v", policyName, policy.RetentionPeriod)
			}
			if policy.CleanupInterval <= 0 {
				t.Errorf("Policy %q has invalid cleanup interval: %v", policyName, policy.CleanupInterval)
			}
		}

		// Validate policy progression (development < production < release)
		devRetention := policies["development"].RetentionPeriod
		prodRetention := policies["production"].RetentionPeriod
		releaseRetention := policies["release"].RetentionPeriod

		if devRetention >= prodRetention {
			t.Error("Development retention should be shorter than production")
		}
		if prodRetention >= releaseRetention {
			t.Error("Production retention should be shorter than release")
		}
	})

	t.Run("SetRetentionPolicy", func(t *testing.T) {
		testPolicy := RetentionPolicy{
			Name:            "test",
			RetentionPeriod: 1 * time.Hour,
			MaxCount:        10,
			CompressAfter:   30 * time.Minute,
			CleanupInterval: 15 * time.Minute,
		}

		manager.SetRetentionPolicy("test", testPolicy)

		// Verify policy was set
		if storedPolicy, exists := manager.policies["test"]; !exists {
			t.Error("Test policy was not stored")
		} else if storedPolicy.RetentionPeriod != testPolicy.RetentionPeriod {
			t.Errorf("Expected retention period %v, got %v", testPolicy.RetentionPeriod, storedPolicy.RetentionPeriod)
		}
	})

	t.Run("CleanupArtifacts", func(t *testing.T) {
		// Create test artifact with old modification time
		testFile := filepath.Join(tempDir, "old_binary")
		if err := os.WriteFile(testFile, []byte("old content"), 0o644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Store artifact
		_, err := manager.StoreArtifact(testFile, "test", "linux", "amd64", nil)
		if err != nil {
			t.Fatalf("Failed to store artifact: %v", err)
		}

		// Set a very short retention policy for testing
		shortPolicy := RetentionPolicy{
			Name:            "short",
			RetentionPeriod: 1 * time.Nanosecond, // Very short for immediate cleanup
			MaxCount:        1,
			CompressAfter:   1 * time.Nanosecond,
			CleanupInterval: 1 * time.Minute,
		}
		manager.SetRetentionPolicy("short", shortPolicy)

		// Wait a moment to ensure artifact is "old"
		time.Sleep(10 * time.Millisecond)

		// Run cleanup
		if err := manager.CleanupArtifacts("short"); err != nil {
			t.Fatalf("Failed to cleanup artifacts: %v", err)
		}

		// Verify artifacts were cleaned up
		artifacts, err := manager.ListArtifacts("test", "", "")
		if err != nil {
			t.Fatalf("Failed to list artifacts: %v", err)
		}

		if len(artifacts) > 0 {
			t.Errorf("Expected 0 artifacts after cleanup, got %d", len(artifacts))
		}
	})
}

// TestArtifactCompression tests the compression functionality
func TestArtifactCompression(t *testing.T) {
	tempDir := t.TempDir()

	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create artifact manager: %v", err)
	}

	t.Run("CompressOldArtifacts", func(t *testing.T) {
		// Create and store test artifact
		testFile := filepath.Join(tempDir, "compress_test")
		testContent := []byte("content to compress " + string(make([]byte, 1000))) // Make it larger
		if err := os.WriteFile(testFile, testContent, 0o644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		info, err := manager.StoreArtifact(testFile, "compress", "linux", "amd64", nil)
		if err != nil {
			t.Fatalf("Failed to store artifact: %v", err)
		}

		// Set policy that compresses immediately
		compressPolicy := RetentionPolicy{
			Name:            "compress",
			RetentionPeriod: 1 * time.Hour,
			MaxCount:        10,
			CompressAfter:   1 * time.Nanosecond, // Compress immediately
			CleanupInterval: 1 * time.Minute,
		}
		manager.SetRetentionPolicy("compress", compressPolicy)

		// Wait a moment and run compression
		time.Sleep(10 * time.Millisecond)
		if err := manager.CompressOldArtifacts("compress"); err != nil {
			t.Fatalf("Failed to compress artifacts: %v", err)
		}

		// Check that compressed file exists
		compressedPath := filepath.Join(manager.artifactsDir, "compress", "linux_amd64", info.Name+".gz")
		if _, err := os.Stat(compressedPath); os.IsNotExist(err) {
			t.Error("Compressed artifact file should exist")
		}

		// Check that original file was removed
		originalPath := filepath.Join(manager.artifactsDir, "compress", "linux_amd64", info.Name)
		if _, err := os.Stat(originalPath); !os.IsNotExist(err) {
			t.Error("Original artifact file should be removed after compression")
		}
	})
}

// TestErrorHandling tests error cases and edge conditions
func TestErrorHandling(t *testing.T) {
	t.Run("InvalidArtifactsDirectory", func(t *testing.T) {
		// Try to create manager with invalid directory
		_, err := NewManager("/invalid/path/that/cannot/be/created")
		if err == nil {
			t.Error("Expected error when creating manager with invalid directory")
		}
	})

	t.Run("NonexistentRetentionPolicy", func(t *testing.T) {
		tempDir := t.TempDir()
		manager, err := NewManager(tempDir)
		if err != nil {
			t.Fatalf("Failed to create artifact manager: %v", err)
		}

		// Try to cleanup with nonexistent policy
		err = manager.CleanupArtifacts("nonexistent")
		if err == nil {
			t.Error("Expected error when using nonexistent retention policy")
		}

		// Try to compress with nonexistent policy
		err = manager.CompressOldArtifacts("nonexistent")
		if err == nil {
			t.Error("Expected error when using nonexistent retention policy")
		}
	})

	t.Run("InvalidArtifactFile", func(t *testing.T) {
		tempDir := t.TempDir()
		manager, err := NewManager(tempDir)
		if err != nil {
			t.Fatalf("Failed to create artifact manager: %v", err)
		}

		// Try to store nonexistent file
		_, err = manager.StoreArtifact("/nonexistent/file", "test", "linux", "amd64", nil)
		if err == nil {
			t.Error("Expected error when storing nonexistent artifact")
		}
	})
}

// BenchmarkArtifactOperations benchmarks key artifact operations
func BenchmarkArtifactOperations(b *testing.B) {
	tempDir := b.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		b.Fatalf("Failed to create artifact manager: %v", err)
	}

	// Create test file
	testFile := filepath.Join(tempDir, "bench_test")
	testContent := make([]byte, 1024*1024) // 1MB test file
	if err := os.WriteFile(testFile, testContent, 0o644); err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	b.Run("StoreArtifact", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := manager.StoreArtifact(testFile, "bench", "linux", "amd64", nil)
			if err != nil {
				b.Fatalf("Failed to store artifact: %v", err)
			}
		}
	})

	b.Run("ListArtifacts", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := manager.ListArtifacts("bench", "", "")
			if err != nil {
				b.Fatalf("Failed to list artifacts: %v", err)
			}
		}
	})

	b.Run("GetArtifactStats", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := manager.GetArtifactStats()
			if err != nil {
				b.Fatalf("Failed to get artifact stats: %v", err)
			}
		}
	})
}
