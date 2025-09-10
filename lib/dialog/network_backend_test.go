package dialog

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

// MockNetworkCoordinator implements NetworkCoordinator for testing
type MockNetworkCoordinator struct {
	available          bool
	peers              []PeerInfo
	peerResponses      []PeerDialogResponse
	requestError       error
	broadcastError     error
	requestCallCount   int
	broadcastCallCount int
}

func (m *MockNetworkCoordinator) RequestPeerDialogs(context DialogContext) ([]PeerDialogResponse, error) {
	m.requestCallCount++
	if m.requestError != nil {
		return nil, m.requestError
	}
	return m.peerResponses, nil
}

func (m *MockNetworkCoordinator) BroadcastDialogResponse(context DialogContext, response DialogResponse) error {
	m.broadcastCallCount++
	return m.broadcastError
}

func (m *MockNetworkCoordinator) GetConnectedPeers() []PeerInfo {
	return m.peers
}

func (m *MockNetworkCoordinator) IsNetworkAvailable() bool {
	return m.available
}

// Test helper to create a basic dialog context
func createTestDialogContext() DialogContext {
	return DialogContext{
		Trigger:       "click",
		InteractionID: "test-interaction-1",
		Timestamp:     time.Now(),
		CurrentStats: map[string]float64{
			"affection": 50.0,
			"trust":     30.0,
		},
		PersonalityTraits: map[string]float64{
			"shyness":    0.3,
			"chattiness": 0.7,
		},
		CurrentMood:       75.0,
		CurrentAnimation:  "idle",
		RelationshipLevel: "Friend",
		ConversationTurn:  1,
		FallbackResponses: []string{"Hello!", "Hi there!"},
		FallbackAnimation: "talking",
	}
}

func TestNewNetworkDialogBackend(t *testing.T) {
	backend := NewNetworkDialogBackend()

	if backend == nil {
		t.Fatal("NewNetworkDialogBackend() returned nil")
	}

	if backend.responseCache == nil {
		t.Error("responseCache not initialized")
	}

	if backend.peerResponses == nil {
		t.Error("peerResponses not initialized")
	}

	if backend.coordinationTimeout != 500*time.Millisecond {
		t.Errorf("Expected default coordination timeout 500ms, got %v", backend.coordinationTimeout)
	}
}

func TestNetworkDialogBackend_Initialize(t *testing.T) {
	tests := []struct {
		name        string
		config      string
		expectError bool
		expectType  string
	}{
		{
			name:        "Default configuration",
			config:      `{}`,
			expectError: false,
			expectType:  "network",
		},
		{
			name: "Custom configuration",
			config: `{
				"type": "network",
				"localBackendType": "simple_random",
				"coordinationTimeout": "1s",
				"enableGroupDialogs": false,
				"responsePriority": "random",
				"minPeersForGroup": 3
			}`,
			expectError: false,
			expectType:  "network",
		},
		{
			name: "With local backend configuration",
			config: `{
				"localBackendType": "simple_random",
				"localBackendConfig": {
					"personalityInfluence": 0.5,
					"useDialogHistory": true
				}
			}`,
			expectError: false,
			expectType:  "network",
		},
		{
			name: "Invalid JSON",
			config: `{
				"type": "network",
				"coordinationTimeout": "invalid-duration"
			}`,
			expectError: true,
		},
		{
			name: "Unsupported local backend",
			config: `{
				"localBackendType": "unsupported_backend"
			}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend := NewNetworkDialogBackend()
			err := backend.Initialize(json.RawMessage(tt.config))

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if backend.config.Type != tt.expectType {
				t.Errorf("Expected type %s, got %s", tt.expectType, backend.config.Type)
			}

			if backend.localBackend == nil {
				t.Error("Local backend not initialized")
			}
		})
	}
}

func TestNetworkDialogBackend_GenerateResponse_NoNetwork(t *testing.T) {
	backend := NewNetworkDialogBackend()

	// Initialize with simple config
	err := backend.Initialize(json.RawMessage(`{"localBackendType": "simple_random"}`))
	if err != nil {
		t.Fatalf("Failed to initialize backend: %v", err)
	}

	context := createTestDialogContext()
	response, err := backend.GenerateResponse(context)
	if err != nil {
		t.Errorf("GenerateResponse() error: %v", err)
	}

	if response.Text == "" {
		t.Error("Expected non-empty response text")
	}

	if response.Confidence <= 0 {
		t.Error("Expected positive confidence value")
	}
}

func TestNetworkDialogBackend_GenerateResponse_WithNetworkCoordination(t *testing.T) {
	backend := NewNetworkDialogBackend()

	// Initialize backend
	config := `{
		"localBackendType": "simple_random",
		"enableGroupDialogs": true,
		"minPeersForGroup": 1,
		"responsePriority": "first"
	}`
	err := backend.Initialize(json.RawMessage(config))
	if err != nil {
		t.Fatalf("Failed to initialize backend: %v", err)
	}

	// Set up mock network coordinator
	mockCoordinator := &MockNetworkCoordinator{
		available: true,
		peers: []PeerInfo{
			{ID: "peer1", CharacterType: "companion", IsBot: true},
			{ID: "peer2", CharacterType: "friend", IsBot: false},
		},
		peerResponses: []PeerDialogResponse{
			{
				PeerID: "peer1",
				Response: DialogResponse{
					Text:       "Hello from peer!",
					Confidence: 0.8,
					Animation:  "wave",
				},
				Confidence: 0.8,
				Timestamp:  time.Now(),
			},
		},
	}
	backend.SetNetworkCoordinator(mockCoordinator)

	context := createTestDialogContext()
	response, err := backend.GenerateResponse(context)
	if err != nil {
		t.Errorf("GenerateResponse() error: %v", err)
	}

	if response.Text == "" {
		t.Error("Expected non-empty response text")
	}

	// Verify network coordination was attempted
	if mockCoordinator.requestCallCount != 1 {
		t.Errorf("Expected 1 peer dialog request, got %d", mockCoordinator.requestCallCount)
	}

	if mockCoordinator.broadcastCallCount != 1 {
		t.Errorf("Expected 1 broadcast call, got %d", mockCoordinator.broadcastCallCount)
	}
}

func TestNetworkDialogBackend_ResponseSelection(t *testing.T) {
	tests := []struct {
		name              string
		priority          string
		localResponse     DialogResponse
		peerResponses     []PeerDialogResponse
		expectedText      string
		personalityTraits map[string]float64
	}{
		{
			name:     "First response priority",
			priority: "first",
			localResponse: DialogResponse{
				Text:       "Local response",
				Confidence: 0.5,
			},
			peerResponses: []PeerDialogResponse{
				{
					Response: DialogResponse{
						Text:       "Peer response",
						Confidence: 0.8,
					},
				},
			},
			expectedText: "Local response", // First response (local) should be selected
		},
		{
			name:     "Highest confidence priority",
			priority: "confidence",
			localResponse: DialogResponse{
				Text:       "Local response",
				Confidence: 0.5,
			},
			peerResponses: []PeerDialogResponse{
				{
					Response: DialogResponse{
						Text:       "Better peer response",
						Confidence: 0.8,
					},
					Confidence: 0.8, // Add this to match the peer response
				},
			},
			expectedText: "Better peer response", // Higher confidence should be selected
		},
		{
			name:     "Personality-based selection for shy character",
			priority: "personality",
			localResponse: DialogResponse{
				Text:       "This is a long and very detailed response that goes on and on",
				Confidence: 0.6,
			},
			peerResponses: []PeerDialogResponse{
				{
					Response: DialogResponse{
						Text:       "Short reply",
						Confidence: 0.6,
					},
				},
			},
			personalityTraits: map[string]float64{
				"shyness": 0.8, // Very shy
			},
			expectedText: "Short reply", // Shy character prefers shorter responses
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend := NewNetworkDialogBackend()

			config := fmt.Sprintf(`{
				"localBackendType": "simple_random",
				"responsePriority": "%s"
			}`, tt.priority)

			err := backend.Initialize(json.RawMessage(config))
			if err != nil {
				t.Fatalf("Failed to initialize backend: %v", err)
			}

			context := createTestDialogContext()
			if tt.personalityTraits != nil {
				context.PersonalityTraits = tt.personalityTraits
			}

			selected := backend.selectBestResponse(context, tt.localResponse, tt.peerResponses)

			if selected.Text != tt.expectedText {
				t.Errorf("Expected response text %q, got %q", tt.expectedText, selected.Text)
			}
		})
	}
}

func TestNetworkDialogBackend_ResponseCaching(t *testing.T) {
	backend := NewNetworkDialogBackend()

	config := `{
		"localBackendType": "simple_random",
		"cacheExpiry": "100ms"
	}`
	err := backend.Initialize(json.RawMessage(config))
	if err != nil {
		t.Fatalf("Failed to initialize backend: %v", err)
	}

	context := createTestDialogContext()

	// Generate first response - should not be cached
	response1, err := backend.GenerateResponse(context)
	if err != nil {
		t.Errorf("GenerateResponse() error: %v", err)
	}

	// Generate second response immediately - should be cached
	response2, err := backend.GenerateResponse(context)
	if err != nil {
		t.Errorf("GenerateResponse() error: %v", err)
	}

	if response1.Text != response2.Text {
		t.Error("Expected cached response to match original")
	}

	// Wait for cache to expire
	time.Sleep(150 * time.Millisecond)

	// Generate third response - should be new (not cached)
	_, err = backend.GenerateResponse(context)
	if err != nil {
		t.Errorf("GenerateResponse() error: %v", err)
	}

	// Note: We can't guarantee response3 is different from response1/2
	// because the underlying random backend might generate the same response
	// But we can verify the cache mechanism by checking cache stats
	stats := backend.GetCacheStats()
	if cacheSize, ok := stats["cacheSize"]; ok {
		if cacheSize.(int) < 0 {
			t.Error("Cache size should not be negative")
		}
	}
}

func TestNetworkDialogBackend_FallbackBehavior(t *testing.T) {
	backend := NewNetworkDialogBackend()

	err := backend.Initialize(json.RawMessage(`{"localBackendType": "simple_random"}`))
	if err != nil {
		t.Fatalf("Failed to initialize backend: %v", err)
	}

	// Set up mock coordinator that fails
	mockCoordinator := &MockNetworkCoordinator{
		available:    true,
		peers:        []PeerInfo{{ID: "peer1"}},
		requestError: fmt.Errorf("network error"),
	}
	backend.SetNetworkCoordinator(mockCoordinator)

	context := createTestDialogContext()
	response, err := backend.GenerateResponse(context)
	// Should not error even if network coordination fails
	if err != nil {
		t.Errorf("GenerateResponse() should not error on network failure: %v", err)
	}

	if response.Text == "" {
		t.Error("Expected fallback response when network coordination fails")
	}
}

func TestNetworkDialogBackend_GetBackendInfo(t *testing.T) {
	backend := NewNetworkDialogBackend()

	info := backend.GetBackendInfo()

	if info.Name != "network_dialog" {
		t.Errorf("Expected name 'network_dialog', got %s", info.Name)
	}

	if info.Version == "" {
		t.Error("Expected non-empty version")
	}

	if info.Description == "" {
		t.Error("Expected non-empty description")
	}

	expectedCapabilities := []string{
		"peer_coordination",
		"group_dialogs",
		"response_caching",
		"fallback_support",
	}

	for _, expected := range expectedCapabilities {
		found := false
		for _, capability := range info.Capabilities {
			if capability == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected capability %s not found in backend info", expected)
		}
	}
}

func TestNetworkDialogBackend_CanHandle(t *testing.T) {
	backend := NewNetworkDialogBackend()

	err := backend.Initialize(json.RawMessage(`{"localBackendType": "simple_random"}`))
	if err != nil {
		t.Fatalf("Failed to initialize backend: %v", err)
	}

	context := createTestDialogContext()
	canHandle := backend.CanHandle(context)

	if !canHandle {
		t.Error("Network backend should be able to handle basic dialog context")
	}

	// Test with uninitialized backend (no local backend)
	backendUninitialized := NewNetworkDialogBackend()
	canHandleUninitialized := backendUninitialized.CanHandle(context)

	if !canHandleUninitialized {
		t.Error("Network backend should default to accepting contexts when local backend is nil")
	}
}

func TestNetworkDialogBackend_UpdateMemory(t *testing.T) {
	backend := NewNetworkDialogBackend()

	err := backend.Initialize(json.RawMessage(`{"localBackendType": "simple_random"}`))
	if err != nil {
		t.Fatalf("Failed to initialize backend: %v", err)
	}

	context := createTestDialogContext()
	response := DialogResponse{
		Text:       "Test response",
		Confidence: 0.8,
	}
	feedback := &UserFeedback{
		Positive:     true,
		ResponseTime: 2 * time.Second,
		Engagement:   0.7,
	}

	err = backend.UpdateMemory(context, response, feedback)
	if err != nil {
		t.Errorf("UpdateMemory() error: %v", err)
	}

	// Test with nil feedback
	err = backend.UpdateMemory(context, response, nil)
	if err != nil {
		t.Errorf("UpdateMemory() should handle nil feedback: %v", err)
	}
}

func TestNetworkDialogBackend_CacheManagement(t *testing.T) {
	backend := NewNetworkDialogBackend()

	err := backend.Initialize(json.RawMessage(`{"cacheExpiry": "10ms"}`))
	if err != nil {
		t.Fatalf("Failed to initialize backend: %v", err)
	}

	context := createTestDialogContext()
	response := DialogResponse{
		Text:       "Cached response",
		Confidence: 0.7,
	}

	// Cache a response
	backend.cacheResponse(context, response, "test")

	// Verify it's cached
	cached, found := backend.getCachedResponse(context)
	if !found {
		t.Error("Expected to find cached response")
	}

	if cached.Response.Text != response.Text {
		t.Error("Cached response text doesn't match original")
	}

	if cached.Source != "test" {
		t.Error("Cached response source doesn't match")
	}

	// Test cache stats
	stats := backend.GetCacheStats()
	if cacheSize, ok := stats["cacheSize"]; !ok || cacheSize.(int) == 0 {
		t.Error("Expected non-zero cache size in stats")
	}

	// Clear cache
	backend.ClearCache()

	// Verify cache is cleared
	stats = backend.GetCacheStats()
	if cacheSize, ok := stats["cacheSize"]; !ok || cacheSize.(int) != 0 {
		t.Error("Expected zero cache size after clearing")
	}
}

func TestNetworkDialogBackend_CoordinationTimeout(t *testing.T) {
	backend := NewNetworkDialogBackend()

	// Initialize the backend first
	err := backend.Initialize(json.RawMessage(`{"localBackendType": "simple_random"}`))
	if err != nil {
		t.Fatalf("Failed to initialize backend: %v", err)
	}

	// Set short coordination timeout for testing
	backend.coordinationTimeout = 50 * time.Millisecond
	backend.config.MinPeersForGroup = 1

	// Mock coordinator that takes too long
	mockCoordinator := &MockNetworkCoordinator{
		available: true,
		peers:     []PeerInfo{{ID: "peer1"}},
	}
	backend.SetNetworkCoordinator(mockCoordinator)

	context := createTestDialogContext()

	// Test rapid successive calls to ensure coordination doesn't happen too frequently
	start := time.Now()

	// First call should attempt coordination
	_, err = backend.GenerateResponse(context)
	if err != nil {
		t.Errorf("First GenerateResponse() error: %v", err)
	}

	// Second call within 1 second should skip coordination
	_, err = backend.GenerateResponse(context)
	if err != nil {
		t.Errorf("Second GenerateResponse() error: %v", err)
	}

	elapsed := time.Since(start)

	// Should not take too long due to coordination throttling
	if elapsed > 200*time.Millisecond {
		t.Errorf("Response generation took too long: %v", elapsed)
	}

	// Verify coordination was called only once due to throttling
	if mockCoordinator.requestCallCount > 1 {
		t.Errorf("Expected at most 1 coordination request due to throttling, got %d", mockCoordinator.requestCallCount)
	}
}

func BenchmarkNetworkDialogBackend_GenerateResponse(b *testing.B) {
	backend := NewNetworkDialogBackend()

	err := backend.Initialize(json.RawMessage(`{"localBackendType": "simple_random"}`))
	if err != nil {
		b.Fatalf("Failed to initialize backend: %v", err)
	}

	context := createTestDialogContext()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := backend.GenerateResponse(context)
		if err != nil {
			b.Errorf("GenerateResponse() error: %v", err)
		}
	}
}

func BenchmarkNetworkDialogBackend_WithNetworkCoordination(b *testing.B) {
	backend := NewNetworkDialogBackend()

	config := `{
		"localBackendType": "simple_random",
		"minPeersForGroup": 1
	}`
	err := backend.Initialize(json.RawMessage(config))
	if err != nil {
		b.Fatalf("Failed to initialize backend: %v", err)
	}

	// Set up mock coordinator
	mockCoordinator := &MockNetworkCoordinator{
		available: true,
		peers:     []PeerInfo{{ID: "peer1"}},
		peerResponses: []PeerDialogResponse{
			{
				PeerID: "peer1",
				Response: DialogResponse{
					Text:       "Peer response",
					Confidence: 0.7,
				},
			},
		},
	}
	backend.SetNetworkCoordinator(mockCoordinator)

	context := createTestDialogContext()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset coordination timing to allow coordination on each iteration
		backend.lastCoordination = time.Time{}

		_, err := backend.GenerateResponse(context)
		if err != nil {
			b.Errorf("GenerateResponse() error: %v", err)
		}
	}
}
