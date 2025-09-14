package dialog

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// Test helper functions for MarkovChainBackend testing

// createTestMarkovConfig creates a basic test configuration for MarkovChainBackend
func createTestMarkovConfig() MarkovConfig {
	return MarkovConfig{
		ChainOrder:     2,
		MinWords:       3,
		MaxWords:       15,
		TemperatureMin: 0.1,
		TemperatureMax: 0.8,
		TrainingData: []string{
			"Hello friend, how are you doing today?",
			"I hope you are having a wonderful day",
			"It's so nice to chat with you",
			"What would you like to talk about?",
			"I'm here to keep you company",
			"How has your day been so far?",
			"Thank you for spending time with me",
			"I enjoy our conversations together",
			"Is there anything I can help you with?",
			"You always make me smile",
			"I'm glad we can be friends",
			"Tell me about your favorite things",
		},
		UseDialogHistory:   false,
		UsePersonality:     true,
		PersonalityBoost:   0.5,
		MoodInfluence:      0.3,
		TriggerSpecific:    true,
		StatAwareness:      0.2,
		RelationshipWeight: 0.4,
		TimeOfDayWeight:    0.1,
		MemoryDecay:        0.1,
		LearningRate:       0.2,
		AdaptationSteps:    5,
		CoherenceThreshold: 0.3,
		SimilarityPenalty:  0.2,
		FallbackPhrases:    []string{"I'm here for you", "Let's chat"},
	}
}

// createTestDialogContextMarkov creates a test DialogContext with reasonable defaults
func createTestDialogContextMarkov() DialogContext {
	return DialogContext{
		Trigger:       "click",
		InteractionID: "test_markov_001",
		Timestamp:     time.Now(),
		CurrentStats: map[string]float64{
			"happiness":  75.0,
			"energy":     60.0,
			"affection":  50.0,
			"friendship": 40.0,
			"romance":    20.0,
			"trust":      65.0,
			"excitement": 45.0,
			"relaxation": 55.0,
		},
		PersonalityTraits: map[string]float64{
			"shyness":      0.3,
			"friendliness": 0.7,
			"intelligence": 0.8,
			"humor":        0.5,
			"creativity":   0.6,
		},
		CurrentMood:       70.0,
		CurrentAnimation:  "idle",
		RelationshipLevel: "Friend",
		InteractionHistory: []InteractionRecord{
			{
				Type:      "click",
				Response:  "Hello there!",
				Timestamp: time.Now().Add(-5 * time.Minute),
				Stats: map[string]float64{
					"happiness": 70.0,
				},
				Outcome: "positive",
			},
		},
		TimeOfDay:         "afternoon",
		FallbackResponses: []string{"Hello!", "Hi there!", "How can I help?"},
	}
}

// createInvalidMarkovConfig creates an invalid configuration for testing error handling
func createInvalidMarkovConfig() MarkovConfig {
	return MarkovConfig{
		ChainOrder:     0,          // Invalid: must be >= 1
		MinWords:       -1,         // Invalid: must be positive
		MaxWords:       1,          // Invalid: must be > MinWords
		TemperatureMin: 1.5,        // Invalid: must be <= 1.0
		TemperatureMax: 0.1,        // Invalid: must be >= TemperatureMin
		TrainingData:   []string{}, // Invalid: empty training data
	}
}

// assertValidDialogResponse checks that a DialogResponse has required fields
func assertValidDialogResponse(t *testing.T, response DialogResponse, expectedMinLength int) {
	t.Helper()

	if response.Text == "" {
		t.Error("Response text should not be empty")
	}

	if len(strings.Fields(response.Text)) < expectedMinLength {
		t.Errorf("Response should have at least %d words, got: %q", expectedMinLength, response.Text)
	}

	if response.Confidence < 0.0 || response.Confidence > 1.0 {
		t.Errorf("Confidence should be between 0 and 1, got: %f", response.Confidence)
	}
}

// assertChainNotEmpty verifies that a MarkovChain has been trained with data
func assertChainNotEmpty(t *testing.T, chain *MarkovChain, description string) {
	t.Helper()

	if chain == nil {
		t.Errorf("%s: chain should not be nil", description)
		return
	}

	if len(chain.states) == 0 {
		t.Errorf("%s: chain should have trained states", description)
	}

	if len(chain.starters) == 0 {
		t.Errorf("%s: chain should have starter words", description)
	}

	if chain.totalWords == 0 {
		t.Errorf("%s: chain should have processed words", description)
	}
}

// TestMarkovChainBackend_NewMarkovChainBackend tests the constructor
func TestMarkovChainBackend_NewMarkovChainBackend(t *testing.T) {
	backend := NewMarkovChainBackend()

	if backend == nil {
		t.Fatal("NewMarkovChainBackend should return a non-nil backend")
	}

	if backend.chains == nil {
		t.Error("Backend should have initialized chains map")
	}

	if backend.initialized {
		t.Error("Backend should not be initialized before Initialize() is called")
	}

	if backend.globalChain != nil {
		t.Error("Global chain should be nil before initialization")
	}
}

// TestMarkovChainBackend_Initialize tests configuration initialization
func TestMarkovChainBackend_Initialize(t *testing.T) {
	tests := []struct {
		name        string
		config      interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid basic configuration",
			config:      createTestMarkovConfig(),
			expectError: false,
		},
		{
			name: "Valid minimal configuration",
			config: MarkovConfig{
				ChainOrder:         1,
				MinWords:           1,
				MaxWords:           10,
				TemperatureMin:     0.0,
				TemperatureMax:     1.0,
				TrainingData:       []string{"Hello world"},
				UseDialogHistory:   false,
				UsePersonality:     false,
				TriggerSpecific:    false,
				CoherenceThreshold: 0.0,
			},
			expectError: false,
		},
		{
			name:        "Invalid JSON",
			config:      "invalid json string",
			expectError: true,
			errorMsg:    "failed to parse Markov config",
		},
		{
			name:        "Invalid configuration values",
			config:      createInvalidMarkovConfig(),
			expectError: true,
			errorMsg:    "chainOrder must be 1-5",
		},
		{
			name: "Zero chain order",
			config: MarkovConfig{
				ChainOrder:   0,
				MinWords:     1,
				MaxWords:     10,
				TrainingData: []string{"test"},
			},
			expectError: true,
			errorMsg:    "chainOrder must be 1-5",
		},
		{
			name: "MaxWords less than MinWords",
			config: MarkovConfig{
				ChainOrder:   1,
				MinWords:     10,
				MaxWords:     5,
				TrainingData: []string{"test"},
			},
			expectError: true,
			errorMsg:    "must be >= 1 and <= maxWords",
		},
		{
			name: "Invalid temperature range",
			config: MarkovConfig{
				ChainOrder:     1,
				MinWords:       1,
				MaxWords:       10,
				TemperatureMin: 0.8,
				TemperatureMax: 0.2,
				TrainingData:   []string{"test"},
			},
			expectError: true,
			errorMsg:    "temperature values must be 0 <= min <= max <= 2",
		},
		{
			name: "Empty training data",
			config: MarkovConfig{
				ChainOrder:   1,
				MinWords:     1,
				MaxWords:     10,
				TrainingData: []string{},
			},
			expectError: false, // Empty training data is allowed - backend may use defaults
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend := NewMarkovChainBackend()

			configJSON, err := json.Marshal(tt.config)
			if err != nil && !tt.expectError {
				t.Fatalf("Failed to marshal config: %v", err)
			}

			err = backend.Initialize(json.RawMessage(configJSON))

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing %q but got nil", tt.errorMsg)
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q but got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				if !backend.initialized {
					t.Error("Backend should be marked as initialized after successful Initialize()")
				}

				if backend.globalChain == nil {
					t.Error("Global chain should be created after initialization")
				}

				// Verify training data was processed
				if backend.globalChain.totalWords == 0 {
					t.Error("Global chain should have processed training data")
				}
			}
		})
	}
}

// TestMarkovChainBackend_InitializeTwice tests re-initialization behavior
func TestMarkovChainBackend_InitializeTwice(t *testing.T) {
	backend := NewMarkovChainBackend()
	config := createTestMarkovConfig()

	configJSON, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	// First initialization should succeed
	if err := backend.Initialize(json.RawMessage(configJSON)); err != nil {
		t.Fatalf("First initialization failed: %v", err)
	}

	// Store original state
	originalChainCount := len(backend.chains)
	originalGlobalWords := backend.globalChain.totalWords

	// Second initialization should also succeed and reset state
	if err := backend.Initialize(json.RawMessage(configJSON)); err != nil {
		t.Errorf("Second initialization failed: %v", err)
	}

	// Verify state was reset and re-initialized
	if len(backend.chains) != originalChainCount {
		t.Errorf("Chain count should be consistent: expected %d, got %d",
			originalChainCount, len(backend.chains))
	}

	if backend.globalChain.totalWords != originalGlobalWords {
		t.Errorf("Global chain should be re-trained consistently: expected %d words, got %d",
			originalGlobalWords, backend.globalChain.totalWords)
	}
}

// TestMarkovChainBackend_TrainWithText tests the trainWithText method
func TestMarkovChainBackend_TrainWithText(t *testing.T) {
	backend := NewMarkovChainBackend()
	config := createTestMarkovConfig()
	config.TriggerSpecific = true

	configJSON, _ := json.Marshal(config)
	if err := backend.Initialize(json.RawMessage(configJSON)); err != nil {
		t.Fatalf("Failed to initialize backend: %v", err)
	}

	tests := []struct {
		name        string
		text        string
		trigger     string
		expectError bool
		shouldTrain bool
	}{
		{
			name:        "Valid text with trigger",
			text:        "Hello friend, how are you doing today?",
			trigger:     "greeting",
			expectError: false,
			shouldTrain: true,
		},
		{
			name:        "Empty text",
			text:        "",
			trigger:     "empty",
			expectError: false, // No error, just skipped
			shouldTrain: false,
		},
		{
			name:        "Very short text",
			text:        "Hi",
			trigger:     "short",
			expectError: false, // No error, but may be skipped due to length
			shouldTrain: false,
		},
		{
			name:        "Long text with punctuation",
			text:        "This is a longer text sample with various punctuation marks! It should be processed correctly, don't you think?",
			trigger:     "complex",
			expectError: false,
			shouldTrain: true,
		},
		{
			name:        "Text with numbers and symbols",
			text:        "Meeting at 3:30 PM today @ the office (building #5).",
			trigger:     "symbols",
			expectError: false,
			shouldTrain: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initialGlobalWords := backend.globalChain.totalWords

			err := backend.trainWithText(tt.text, tt.trigger)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				if tt.shouldTrain {
					// Verify global chain was trained
					if backend.globalChain.totalWords <= initialGlobalWords {
						t.Error("Global chain should have been trained with valid text")
					}
				}
			}
		})
	}
}

// TestMarkovChainBackend_CleanTrainingText tests text cleaning functionality
func TestMarkovChainBackend_CleanTrainingText(t *testing.T) {
	backend := NewMarkovChainBackend()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Basic text",
			input:    "Hello world",
			expected: "Hello world",
		},
		{
			name:     "Text with newlines",
			input:    "Hello\nworld",
			expected: "Hello world",
		},
		{
			name:     "Text with leading/trailing spaces",
			input:    "  Hello world  ",
			expected: "Hello world",
		},
		{
			name:     "Text with punctuation",
			input:    "Hello, world! How are you?",
			expected: "Hello, world! How are you?",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Only whitespace",
			input:    "   \n\t\r   ",
			expected: "",
		},
		{
			name:     "Text with emoji (if implemented)",
			input:    "Hello 😊 world 👋",
			expected: "Hello  world", // Emojis removed but spaces may remain
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := backend.cleanTrainingText(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q but got %q", tt.expected, result)
			}
		})
	}
}

// TestMarkovChain_Train tests the Train method on MarkovChain
func TestMarkovChain_Train(t *testing.T) {
	// Create a new chain directly for testing
	chain := &MarkovChain{
		order:      2,
		states:     make(map[string][]string),
		starters:   []string{},
		wordCounts: make(map[string]int),
		totalWords: 0,
	}

	tests := []struct {
		name           string
		text           string
		expectedStates bool
		expectedWords  bool
	}{
		{
			name:           "Simple sentence",
			text:           "Hello world today",
			expectedStates: true,
			expectedWords:  true,
		},
		{
			name:           "Single word",
			text:           "Hello",
			expectedStates: false, // Not enough words for bigram
			expectedWords:  true,
		},
		{
			name:           "Empty text",
			text:           "",
			expectedStates: false,
			expectedWords:  false,
		},
		{
			name:           "Long sentence",
			text:           "This is a longer sentence with many words to test the training functionality",
			expectedStates: true,
			expectedWords:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initialStates := len(chain.states)
			initialWords := chain.totalWords

			chain.Train(tt.text)

			if tt.expectedStates {
				if len(chain.states) <= initialStates {
					t.Error("Expected chain states to increase")
				}
			}

			if tt.expectedWords {
				if chain.totalWords <= initialWords {
					t.Error("Expected total word count to increase")
				}

				// Check that word counts were updated
				words := strings.Fields(tt.text)
				for _, word := range words {
					if _, exists := chain.wordCounts[word]; !exists && len(words) > 0 {
						t.Errorf("Word %q should be in word counts", word)
					}
				}
			}
		})
	}
}

// TestMarkovChain_Generate tests the Generate method
func TestMarkovChain_Generate(t *testing.T) {
	// Create and train a chain
	chain := &MarkovChain{
		order:      2,
		states:     make(map[string][]string),
		starters:   []string{},
		wordCounts: make(map[string]int),
		totalWords: 0,
	}

	// Train with multiple sentences to ensure good coverage
	trainingTexts := []string{
		"Hello world today is beautiful",
		"Today is a wonderful day",
		"The weather is beautiful today",
		"Hello friend how are you today",
		"What a wonderful beautiful day",
	}

	for _, text := range trainingTexts {
		chain.Train(text)
	}

	tests := []struct {
		name        string
		targetWords int
		temperature float64
		expectText  bool
	}{
		{
			name:        "Generate short response",
			targetWords: 3,
			temperature: 0.5,
			expectText:  true,
		},
		{
			name:        "Generate longer response",
			targetWords: 8,
			temperature: 0.3,
			expectText:  true,
		},
		{
			name:        "High temperature",
			targetWords: 5,
			temperature: 0.9,
			expectText:  true,
		},
		{
			name:        "Low temperature",
			targetWords: 5,
			temperature: 0.1,
			expectText:  true,
		},
		{
			name:        "Zero target words",
			targetWords: 0,
			temperature: 0.5,
			expectText:  true, // Still generates minimum response
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text, confidence := chain.Generate(tt.targetWords, tt.temperature)

			if tt.expectText {
				if text == "" {
					t.Error("Expected non-empty text")
				}

				words := strings.Fields(text)
				if len(words) == 0 {
					t.Error("Generated text should contain words")
				}

				// Confidence should be reasonable
				if confidence < 0.0 || confidence > 1.0 {
					t.Errorf("Confidence should be between 0 and 1, got: %f", confidence)
				}
			} else {
				// For zero target words, the generator still produces some output
				// This is the intended behavior - minimum viable response
				if len(strings.Fields(text)) == 0 {
					t.Error("Even zero target should produce some words")
				}
			}
		})
	}
}

// TestMarkovChainBackend_GenerateResponse tests the core response generation
func TestMarkovChainBackend_GenerateResponse(t *testing.T) {
	backend := NewMarkovChainBackend()
	config := createTestMarkovConfig()

	configJSON, _ := json.Marshal(config)
	if err := backend.Initialize(json.RawMessage(configJSON)); err != nil {
		t.Fatalf("Failed to initialize backend: %v", err)
	}

	tests := []struct {
		name           string
		context        DialogContext
		expectResponse bool
		expectError    bool
	}{
		{
			name:           "Basic click interaction",
			context:        createTestDialogContextMarkov(),
			expectResponse: true,
			expectError:    false,
		},
		{
			name: "Greeting trigger",
			context: func() DialogContext {
				ctx := createTestDialogContextMarkov()
				ctx.Trigger = "greeting"
				return ctx
			}(),
			expectResponse: true,
			expectError:    false,
		},
		{
			name: "High mood context",
			context: func() DialogContext {
				ctx := createTestDialogContextMarkov()
				ctx.CurrentMood = 90.0
				ctx.CurrentStats["happiness"] = 85.0
				return ctx
			}(),
			expectResponse: true,
			expectError:    false,
		},
		{
			name: "Low mood context",
			context: func() DialogContext {
				ctx := createTestDialogContextMarkov()
				ctx.CurrentMood = 20.0
				ctx.CurrentStats["happiness"] = 15.0
				return ctx
			}(),
			expectResponse: true,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := backend.GenerateResponse(tt.context)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				if tt.expectResponse {
					assertValidDialogResponse(t, response, 1)

					// Check for reasonable confidence
					if response.Confidence <= 0.0 {
						t.Error("Response confidence should be positive")
					}
				}
			}
		})
	}
}

// TestMarkovChainBackend_CanHandle tests the CanHandle method
func TestMarkovChainBackend_CanHandle(t *testing.T) {
	backend := NewMarkovChainBackend()

	// Test uninitialized backend
	context := createTestDialogContextMarkov()
	if backend.CanHandle(context) {
		t.Error("Uninitialized backend should not handle any context")
	}

	// Initialize with training data
	config := createTestMarkovConfig()
	// Add more training data to ensure chains have enough data
	config.TrainingData = []string{
		"Hello friend, how are you doing today?",
		"I hope you are having a wonderful day",
		"It's so nice to chat with you",
		"What would you like to talk about?",
		"I'm here to keep you company",
		"How has your day been so far?",
		"Thank you for spending time with me",
		"I enjoy our conversations together",
		"Is there anything I can help you with?",
		"You always make me smile",
		"I'm glad we can be friends",
		"Tell me about your favorite things",
	}

	configJSON, _ := json.Marshal(config)
	if err := backend.Initialize(json.RawMessage(configJSON)); err != nil {
		t.Fatalf("Failed to initialize backend: %v", err)
	}

	tests := []struct {
		name     string
		context  DialogContext
		expected bool
	}{
		{
			name:     "Valid context with general trigger",
			context:  createTestDialogContextMarkov(),
			expected: true, // Should handle because global chain has enough data
		},
		{
			name: "Context with different trigger",
			context: func() DialogContext {
				ctx := createTestDialogContextMarkov()
				ctx.Trigger = "greeting"
				return ctx
			}(),
			expected: true, // Should fall back to global chain
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := backend.CanHandle(tt.context)
			if result != tt.expected {
				t.Errorf("Expected CanHandle to return %v but got %v", tt.expected, result)
			}
		})
	}
}

// TestMarkovChainBackend_GetBackendInfo tests backend info retrieval
func TestMarkovChainBackend_GetBackendInfo(t *testing.T) {
	backend := NewMarkovChainBackend()

	info := backend.GetBackendInfo()

	if info.Name == "" {
		t.Error("Backend info should have a name")
	}

	if info.Version == "" {
		t.Error("Backend info should have a version")
	}

	if info.Description == "" {
		t.Error("Backend info should have a description")
	}

	// Check that name contains "markov" in some form
	if !strings.Contains(strings.ToLower(info.Name), "markov") {
		t.Errorf("Backend name should indicate it's a Markov backend, got: %s", info.Name)
	}
}

// TestMarkovChainBackend_ErrorHandling tests error handling and edge cases
func TestMarkovChainBackend_ErrorHandling(t *testing.T) {
	t.Run("Uninitialized backend GenerateResponse", func(t *testing.T) {
		backend := NewMarkovChainBackend()
		context := createTestDialogContextMarkov()

		response, err := backend.GenerateResponse(context)

		// Should error because backend is not initialized
		if err == nil {
			t.Error("Uninitialized backend should return an error")
		}

		// Response should be empty when there's an error
		if response.Text != "" {
			t.Error("Response should be empty when backend errors")
		}
	})

	t.Run("Empty context GenerateResponse", func(t *testing.T) {
		backend := NewMarkovChainBackend()
		config := createTestMarkovConfig()

		configJSON, _ := json.Marshal(config)
		if err := backend.Initialize(json.RawMessage(configJSON)); err != nil {
			t.Fatalf("Failed to initialize backend: %v", err)
		}

		emptyContext := DialogContext{}
		response, err := backend.GenerateResponse(emptyContext)

		if err != nil {
			t.Errorf("Empty context should not cause error: %v", err)
		}

		if response.Text == "" {
			t.Error("Should return some response even with empty context")
		}
	})

	t.Run("UpdateMemory with nil feedback", func(t *testing.T) {
		backend := NewMarkovChainBackend()
		config := createTestMarkovConfig()

		configJSON, _ := json.Marshal(config)
		if err := backend.Initialize(json.RawMessage(configJSON)); err != nil {
			t.Fatalf("Failed to initialize backend: %v", err)
		}

		context := createTestDialogContextMarkov()
		response := DialogResponse{Text: "Hello", Confidence: 0.5}

		err := backend.UpdateMemory(context, response, nil)
		if err != nil {
			t.Errorf("UpdateMemory with nil feedback should not error: %v", err)
		}
	})

	t.Run("Chain with minimal data", func(t *testing.T) {
		backend := NewMarkovChainBackend()
		config := MarkovConfig{
			ChainOrder:      1,
			MinWords:        1,
			MaxWords:        3,
			TemperatureMin:  0.0,
			TemperatureMax:  1.0,
			TrainingData:    []string{"Hi", "Hello", "Hey"}, // Minimal data
			UsePersonality:  false,
			TriggerSpecific: false,
		}

		configJSON, _ := json.Marshal(config)
		if err := backend.Initialize(json.RawMessage(configJSON)); err != nil {
			t.Fatalf("Failed to initialize backend: %v", err)
		}

		context := createTestDialogContextMarkov()
		response, err := backend.GenerateResponse(context)

		if err != nil {
			t.Errorf("Minimal data should not cause error: %v", err)
		}

		if response.Text == "" {
			t.Error("Should generate some response with minimal data")
		}
	})
}

// TestMarkovChainBackend_EdgeCases tests various edge cases
func TestMarkovChainBackend_EdgeCases(t *testing.T) {
	t.Run("Empty chain Generate", func(t *testing.T) {
		chain := &MarkovChain{
			order:      2,
			states:     make(map[string][]string),
			starters:   []string{},
			wordCounts: make(map[string]int),
			totalWords: 0,
		}

		text, confidence := chain.Generate(5, 0.5)

		// Empty chain should return empty text
		if text != "" {
			t.Errorf("Empty chain should return empty text, got: %q", text)
		}

		if confidence != 0.0 {
			t.Errorf("Empty chain should return zero confidence, got: %f", confidence)
		}
	})

	t.Run("Chain hasEnoughData edge cases", func(t *testing.T) {
		tests := []struct {
			name         string
			starters     []string
			states       map[string][]string
			expectEnough bool
		}{
			{
				name:         "Empty chain",
				starters:     []string{},
				states:       make(map[string][]string),
				expectEnough: false,
			},
			{
				name:         "Few starters",
				starters:     []string{"Hello"},
				states:       make(map[string][]string),
				expectEnough: false,
			},
			{
				name:     "Enough starters but few states",
				starters: []string{"Hello", "Hi"},
				states: map[string][]string{
					"Hello": {"world"},
				},
				expectEnough: false,
			},
			{
				name:     "Sufficient data",
				starters: []string{"Hello", "Hi"},
				states: map[string][]string{
					"Hello":  {"world", "there"},
					"Hi":     {"there", "friend"},
					"world":  {"how", "is"},
					"there":  {"friend", "how"},
					"how":    {"are", "is"},
					"are":    {"you", "things"},
					"is":     {"everything", "going"},
					"you":    {"today", "doing"},
					"friend": {"how", "nice"},
					"today":  {"going", "wonderful"},
				},
				expectEnough: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				chain := &MarkovChain{
					order:      2,
					states:     tt.states,
					starters:   tt.starters,
					wordCounts: make(map[string]int),
					totalWords: 0,
				}

				result := chain.hasEnoughData()
				if result != tt.expectEnough {
					t.Errorf("Expected hasEnoughData to return %v but got %v", tt.expectEnough, result)
				}
			})
		}
	})

	t.Run("Extreme temperature values", func(t *testing.T) {
		backend := NewMarkovChainBackend()
		config := createTestMarkovConfig()
		config.TemperatureMin = 0.0
		config.TemperatureMax = 2.0 // Maximum allowed

		configJSON, _ := json.Marshal(config)
		if err := backend.Initialize(json.RawMessage(configJSON)); err != nil {
			t.Fatalf("Failed to initialize backend: %v", err)
		}

		context := createTestDialogContextMarkov()

		// Test with extreme mood values
		context.CurrentMood = 0.0 // Minimum
		temp1 := backend.calculateTemperature(context)

		context.CurrentMood = 100.0 // Maximum
		temp2 := backend.calculateTemperature(context)

		if temp1 < 0.0 || temp1 > 2.0 {
			t.Errorf("Temperature should be in valid range, got: %f", temp1)
		}

		if temp2 < 0.0 || temp2 > 2.0 {
			t.Errorf("Temperature should be in valid range, got: %f", temp2)
		}
	})

	t.Run("Extreme word count values", func(t *testing.T) {
		backend := NewMarkovChainBackend()
		config := createTestMarkovConfig()
		config.MinWords = 1
		config.MaxWords = 50 // Maximum allowed

		configJSON, _ := json.Marshal(config)
		if err := backend.Initialize(json.RawMessage(configJSON)); err != nil {
			t.Fatalf("Failed to initialize backend: %v", err)
		}

		context := createTestDialogContextMarkov()

		// Test with extreme energy values
		context.CurrentStats["energy"] = 0.0 // Minimum
		words1 := backend.calculateTargetWords(context)

		context.CurrentStats["energy"] = 100.0 // Maximum
		words2 := backend.calculateTargetWords(context)

		if words1 < 1 || words1 > 50 {
			t.Errorf("Word count should be in valid range, got: %d", words1)
		}

		if words2 < 1 || words2 > 50 {
			t.Errorf("Word count should be in valid range, got: %d", words2)
		}
	})
}

// BenchmarkMarkovChainBackend_GenerateResponse benchmarks response generation
func BenchmarkMarkovChainBackend_GenerateResponse(b *testing.B) {
	backend := NewMarkovChainBackend()
	config := createTestMarkovConfig()

	configJSON, _ := json.Marshal(config)
	if err := backend.Initialize(json.RawMessage(configJSON)); err != nil {
		b.Fatalf("Failed to initialize backend: %v", err)
	}

	context := createTestDialogContextMarkov()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := backend.GenerateResponse(context)
		if err != nil {
			b.Errorf("Unexpected error: %v", err)
		}
	}
}

// BenchmarkMarkovChain_Train benchmarks chain training
func BenchmarkMarkovChain_Train(b *testing.B) {
	chain := &MarkovChain{
		order:      2,
		states:     make(map[string][]string),
		starters:   []string{},
		wordCounts: make(map[string]int),
		totalWords: 0,
	}

	trainingText := "This is a sample training text with many words to test the performance of chain training functionality"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chain.Train(trainingText)
	}
}

// BenchmarkMarkovChain_Generate benchmarks text generation
func BenchmarkMarkovChain_Generate(b *testing.B) {
	chain := &MarkovChain{
		order:      2,
		states:     make(map[string][]string),
		starters:   []string{},
		wordCounts: make(map[string]int),
		totalWords: 0,
	}

	// Pre-train the chain
	trainingTexts := []string{
		"Hello world today is beautiful and wonderful",
		"Today is a wonderful day for everyone",
		"The weather is beautiful today and tomorrow",
		"Hello friend how are you doing today",
		"What a wonderful beautiful day it is",
		"I hope you are having a beautiful day",
		"Thank you for this wonderful conversation",
		"Hello there how has your day been",
		"Today has been such a wonderful experience",
		"Beautiful weather makes for wonderful days",
	}

	for _, text := range trainingTexts {
		chain.Train(text)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chain.Generate(10, 0.5)
	}
}

// BenchmarkMarkovChainBackend_Initialize benchmarks backend initialization
func BenchmarkMarkovChainBackend_Initialize(b *testing.B) {
	config := createTestMarkovConfig()
	configJSON, _ := json.Marshal(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		backend := NewMarkovChainBackend()
		err := backend.Initialize(json.RawMessage(configJSON))
		if err != nil {
			b.Errorf("Unexpected error: %v", err)
		}
	}
}
