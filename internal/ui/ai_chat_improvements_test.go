package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"desktop-companion/internal/character"
)

// TestAIChatImprovements tests the enhanced AI chat features
func TestAIChatImprovements(t *testing.T) {
	// Simple test of export functionality without full character setup
	t.Run("ChatExportFunctionality", func(t *testing.T) {
		// Create a minimal chatbot interface for testing
		chatbot := &ChatbotInterface{
			conversationLog: []ChatMessage{
				{
					Text:      "Hello!",
					IsUser:    true,
					Timestamp: time.Now().Add(-2 * time.Minute),
				},
				{
					Text:      "Hi there! How can I help you?",
					IsUser:    false,
					Timestamp: time.Now().Add(-1 * time.Minute),
				},
			},
			character: &character.Character{}, // Minimal character for GetName()
		}

		// Test export functionality
		err := chatbot.ExportConversation()
		if err != nil {
			t.Errorf("Failed to export conversation: %v", err)
			return
		}

		// Check if file was created
		homeDir, err := os.UserHomeDir()
		if err != nil {
			t.Errorf("Failed to get home directory: %v", err)
			return
		}

		// Look for the exported file
		files, err := filepath.Glob(filepath.Join(homeDir, "*_chat_*.txt"))
		if err != nil {
			t.Errorf("Failed to search for exported files: %v", err)
			return
		}

		if len(files) == 0 {
			t.Error("No exported chat files found")
			return
		}

		// Read and verify the exported file content
		latestFile := files[len(files)-1] // Get the most recent file
		content, err := os.ReadFile(latestFile)
		if err != nil {
			t.Errorf("Failed to read exported file: %v", err)
			return
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, "Hello!") {
			t.Error("Exported file doesn't contain expected user message")
		}
		if !strings.Contains(contentStr, "Hi there! How can I help you?") {
			t.Error("Exported file doesn't contain expected character response")
		}

		t.Logf("Successfully exported conversation to: %s", latestFile)

		// Clean up - remove test file
		os.Remove(latestFile)
	})

	// Test memory integration APIs
	t.Run("MemoryAPIAvailability", func(t *testing.T) {
		char := &character.Character{}

		// Test that the memory methods exist and can be called safely
		memories := char.GetRecentDialogMemories(5)
		if memories == nil {
			t.Error("GetRecentDialogMemories returned nil instead of empty slice")
		}

		// Test that RecordChatMemory can be called safely
		char.RecordChatMemory("test message", "test response")
		// No error expected since it handles nil gameState gracefully
	})
}
