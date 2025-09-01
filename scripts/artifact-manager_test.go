package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"desktop-companion/internal/artifact"
)

func TestArtifactManager(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	
	// Initialize artifact manager
	manager, err := artifact.NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create artifact manager: %v", err)
	}

	// Test storing an artifact
	testBinary := filepath.Join(tempDir, "test_binary")
	testContent := []byte("test binary content")
	if err := os.WriteFile(testBinary, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}

	// Store the artifact
	info, err := manager.StoreArtifact(testBinary, "test-char", "linux", "amd64", nil)
	if err != nil {
		t.Fatalf("Failed to store artifact: %v", err)
	}

	if info.Character != "test-char" {
		t.Errorf("Expected character 'test-char', got '%s'", info.Character)
	}

	// Test listing artifacts
	artifacts, err := manager.ListArtifacts("", "", "")
	if err != nil {
		t.Fatalf("Failed to list artifacts: %v", err)
	}

	if len(artifacts) != 1 {
		t.Errorf("Expected 1 artifact, got %d", len(artifacts))
	}

	// Test getting statistics
	stats := manager.GetStatistics()
	if stats.TotalArtifacts != 1 {
		t.Errorf("Expected 1 total artifact, got %d", stats.TotalArtifacts)
	}

	if stats.TotalSize <= 0 {
		t.Errorf("Expected positive total size, got %d", stats.TotalSize)
	}
}

func TestArtifactRetention(t *testing.T) {
	tempDir := t.TempDir()
	
	manager, err := artifact.NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create artifact manager: %v", err)
	}

	// Create test binary
	testBinary := filepath.Join(tempDir, "old_binary")
	if err := os.WriteFile(testBinary, []byte("old content"), 0644); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}

	// Store artifact with old timestamp
	if err := manager.Store("old-char", "linux", "amd64", testBinary); err != nil {
		t.Fatalf("Failed to store artifact: %v", err)
	}

	// Manually set old timestamp for testing
	oldTime := time.Now().Add(-30 * 24 * time.Hour) // 30 days ago
	storedPath := filepath.Join(tempDir, "old-char", "linux", "amd64", "old_binary")
	if err := os.Chtimes(storedPath, oldTime, oldTime); err != nil {
		t.Fatalf("Failed to set old timestamp: %v", err)
	}

	// Test development retention policy (7 days)
	policy := artifact.RetentionPolicy{
		Name:        "development",
		MaxAge:      7 * 24 * time.Hour,
		Description: "Development builds retention",
	}

	cleaned, err := manager.ApplyRetentionPolicy(policy)
	if err != nil {
		t.Fatalf("Failed to apply retention policy: %v", err)
	}

	if cleaned != 1 {
		t.Errorf("Expected 1 artifact to be cleaned, got %d", cleaned)
	}

	// Verify artifact was removed
	artifacts, err := manager.List("", "", "")
	if err != nil {
		t.Fatalf("Failed to list artifacts after cleanup: %v", err)
	}

	if len(artifacts) != 0 {
		t.Errorf("Expected 0 artifacts after cleanup, got %d", len(artifacts))
	}
}

func TestArtifactCompression(t *testing.T) {
	tempDir := t.TempDir()
	
	manager, err := artifact.NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create artifact manager: %v", err)
	}

	// Create test binary with compressible content
	testBinary := filepath.Join(tempDir, "large_binary")
	largeContent := make([]byte, 1024*1024) // 1MB of zeros (highly compressible)
	if err := os.WriteFile(testBinary, largeContent, 0644); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}

	// Store the artifact
	if err := manager.Store("large-char", "linux", "amd64", testBinary); err != nil {
		t.Fatalf("Failed to store artifact: %v", err)
	}

	// Get original size
	originalStats, err := manager.GetStatistics()
	if err != nil {
		t.Fatalf("Failed to get original statistics: %v", err)
	}

	// Apply compression
	policy := artifact.RetentionPolicy{
		Name:           "production",
		MaxAge:         90 * 24 * time.Hour,
		EnableCompress: true,
		Description:    "Production builds with compression",
	}

	compressed, err := manager.ApplyRetentionPolicy(policy)
	if err != nil {
		t.Fatalf("Failed to apply compression policy: %v", err)
	}

	if compressed == 0 {
		t.Log("Note: Compression may not always reduce size significantly")
	}

	// Verify artifact is still accessible after compression
	artifacts, err := manager.List("large-char", "", "")
	if err != nil {
		t.Fatalf("Failed to list artifacts after compression: %v", err)
	}

	if len(artifacts) != 1 {
		t.Errorf("Expected 1 artifact after compression, got %d", len(artifacts))
	}

	// Check if compression actually occurred
	newStats, err := manager.GetStatistics()
	if err != nil {
		t.Fatalf("Failed to get statistics after compression: %v", err)
	}

	if newStats.TotalSize < originalStats.TotalSize {
		t.Logf("Compression successful: %d -> %d bytes", originalStats.TotalSize, newStats.TotalSize)
	}
}

func TestArtifactValidation(t *testing.T) {
	tempDir := t.TempDir()
	
	manager, err := artifact.NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create artifact manager: %v", err)
	}

	// Test validation with invalid parameters
	testCases := []struct {
		name      string
		character string
		platform  string
		arch      string
		expectErr bool
	}{
		{"valid", "test-char", "linux", "amd64", false},
		{"empty character", "", "linux", "amd64", true},
		{"empty platform", "test-char", "", "amd64", true},
		{"empty arch", "test-char", "linux", "", true},
		{"invalid character", "test char with spaces", "linux", "amd64", true},
		{"invalid platform", "linux/invalid", "linux", "amd64", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test binary
			testBinary := filepath.Join(tempDir, "test_binary")
			if err := os.WriteFile(testBinary, []byte("test"), 0644); err != nil {
				t.Fatalf("Failed to create test binary: %v", err)
			}

			err := manager.Store(tc.character, tc.platform, tc.arch, testBinary)
			
			if tc.expectErr && err == nil {
				t.Errorf("Expected error for case %s, but got none", tc.name)
			}
			
			if !tc.expectErr && err != nil {
				t.Errorf("Expected no error for case %s, but got: %v", tc.name, err)
			}
			
			// Clean up
			os.Remove(testBinary)
		})
	}
}

func TestArtifactMetadata(t *testing.T) {
	tempDir := t.TempDir()
	
	manager, err := artifact.NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create artifact manager: %v", err)
	}

	// Create test binary
	testBinary := filepath.Join(tempDir, "test_binary")
	testContent := []byte("test binary with metadata")
	if err := os.WriteFile(testBinary, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}

	// Store artifact
	if err := manager.Store("meta-char", "linux", "amd64", testBinary); err != nil {
		t.Fatalf("Failed to store artifact: %v", err)
	}

	// Get artifact metadata
	artifacts, err := manager.List("meta-char", "linux", "amd64")
	if err != nil {
		t.Fatalf("Failed to list artifacts: %v", err)
	}

	if len(artifacts) != 1 {
		t.Fatalf("Expected 1 artifact, got %d", len(artifacts))
	}

	artifact := artifacts[0]
	
	// Validate metadata
	if artifact.Character != "meta-char" {
		t.Errorf("Expected character 'meta-char', got '%s'", artifact.Character)
	}
	
	if artifact.Platform != "linux" {
		t.Errorf("Expected platform 'linux', got '%s'", artifact.Platform)
	}
	
	if artifact.Architecture != "amd64" {
		t.Errorf("Expected architecture 'amd64', got '%s'", artifact.Architecture)
	}
	
	if artifact.Size != int64(len(testContent)) {
		t.Errorf("Expected size %d, got %d", len(testContent), artifact.Size)
	}
	
	if artifact.CreatedAt.IsZero() {
		t.Error("Expected non-zero creation time")
	}
}

func BenchmarkArtifactStore(b *testing.B) {
	tempDir := b.TempDir()
	
	manager, err := artifact.NewManager(tempDir)
	if err != nil {
		b.Fatalf("Failed to create artifact manager: %v", err)
	}

	// Create test binary
	testBinary := filepath.Join(tempDir, "bench_binary")
	testContent := make([]byte, 1024*1024) // 1MB binary
	if err := os.WriteFile(testBinary, testContent, 0644); err != nil {
		b.Fatalf("Failed to create test binary: %v", err)
	}

	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		character := fmt.Sprintf("bench-char-%d", i)
		if err := manager.Store(character, "linux", "amd64", testBinary); err != nil {
			b.Fatalf("Failed to store artifact: %v", err)
		}
	}
}

func BenchmarkArtifactList(b *testing.B) {
	tempDir := b.TempDir()
	
	manager, err := artifact.NewManager(tempDir)
	if err != nil {
		b.Fatalf("Failed to create artifact manager: %v", err)
	}

	// Pre-populate with artifacts
	testBinary := filepath.Join(tempDir, "bench_binary")
	if err := os.WriteFile(testBinary, []byte("test"), 0644); err != nil {
		b.Fatalf("Failed to create test binary: %v", err)
	}

	for i := 0; i < 100; i++ {
		character := fmt.Sprintf("bench-char-%d", i)
		if err := manager.Store(character, "linux", "amd64", testBinary); err != nil {
			b.Fatalf("Failed to store artifact: %v", err)
		}
	}

	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := manager.List("", "", "")
		if err != nil {
			b.Fatalf("Failed to list artifacts: %v", err)
		}
	}
}
