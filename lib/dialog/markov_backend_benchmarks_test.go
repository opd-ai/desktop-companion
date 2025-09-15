package dialog

import (
	"encoding/json"
	"testing"
	"time"
)

// BenchmarkMarkovGeneration benchmarks dialog generation performance
func BenchmarkMarkovGeneration(b *testing.B) {
	backend := NewMarkovChainBackend()
	config := createTestMarkovConfig()

	// Add more training data for realistic benchmark
	config.TrainingData = append(config.TrainingData,
		"This is additional training data for performance testing",
		"We need enough data to create realistic generation scenarios",
		"The more training data we have the better the chains will be",
		"Performance testing requires substantial training corpus",
		"Benchmarking helps us optimize generation speed",
		"Dialog generation should be fast and efficient",
		"Users expect quick responses from virtual companions",
		"Memory usage should remain reasonable during generation",
		"Quality and speed must both be optimized for good UX",
		"Markov chains provide good balance of speed and variety",
	)

	configData, _ := json.Marshal(config)
	backend.Initialize(configData)

	context := DialogContext{
		Trigger: "click",
		CurrentStats: map[string]float64{
			"happiness": 75.0,
			"energy":    60.0,
		},
		PersonalityTraits: map[string]float64{
			"talkativeness": 0.6,
		},
		CurrentMood: 70.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := backend.GenerateResponse(context)
		if err != nil {
			b.Fatalf("Generation failed: %v", err)
		}
	}
}

// BenchmarkMarkovTraining benchmarks training performance
func BenchmarkMarkovTraining(b *testing.B) {
	backend := NewMarkovChainBackend()
	config := createTestMarkovConfig()
	config.TrainingData = []string{} // Start with empty data
	configData, _ := json.Marshal(config)
	backend.Initialize(configData)

	trainingText := "This is a test sentence for training the Markov chain with realistic content"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := backend.trainWithText(trainingText, "benchmark")
		if err != nil {
			b.Fatalf("Training failed: %v", err)
		}
	}
}

// BenchmarkMarkovChainGeneration benchmarks the core chain generation
func BenchmarkMarkovChainGeneration(b *testing.B) {
	chain := &MarkovChain{
		order:      2,
		states:     make(map[string][]string),
		starters:   []string{"hello world", "good morning", "nice day"},
		wordCounts: make(map[string]int),
		totalWords: 100,
	}

	// Set up realistic chain data
	chain.states["hello world"] = []string{"today", "friend", "there"}
	chain.states["good morning"] = []string{"sunshine", "friend", "everyone"}
	chain.states["nice day"] = []string{"today", "isn't", "for"}
	chain.states["world today"] = []string{"is", "seems", "looks"}
	chain.states["morning sunshine"] = []string{"is", "feels", "brings"}
	chain.states["day today"] = []string{"is", "has", "feels"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		text, _ := chain.Generate(8, 0.5)
		if text == "" {
			b.Fatal("Generation should produce text")
		}
	}
}

// BenchmarkMarkovMemoryProcessing benchmarks memory-related operations
func BenchmarkMarkovMemoryProcessing(b *testing.B) {
	backend := NewMarkovChainBackend()
	config := createTestMarkovConfig()
	configData, _ := json.Marshal(config)
	backend.Initialize(configData)

	context := DialogContext{
		Trigger: "click",
		TopicContext: map[string]interface{}{
			"dialogMemories": []interface{}{
				map[string]interface{}{
					"favorite_topics":    []string{"music", "games", "books", "movies"},
					"positive_responses": []string{"I love that!", "That's amazing!", "So cool!"},
				},
			},
		},
		PersonalityTraits: map[string]float64{
			"friendliness": 0.8,
		},
	}

	response := DialogResponse{
		Text:       "I love talking about music!",
		Confidence: 0.8,
	}

	feedback := &UserFeedback{
		Positive:     true,
		Engagement:   0.9,
		ResponseTime: 1 * time.Second,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := backend.UpdateMemory(context, response, feedback)
		if err != nil {
			b.Fatalf("Memory update failed: %v", err)
		}
	}
}

// TestMarkovCoverageCore focuses on improving coverage for key methods
func TestMarkovCoverageCore(t *testing.T) {
	backend := NewMarkovChainBackend()
	config := createTestMarkovConfig()
	configData, _ := json.Marshal(config)
	backend.Initialize(configData)

	// Test various methods to improve coverage
	t.Run("TemperatureCalculation", func(t *testing.T) {
		context := DialogContext{
			CurrentMood: 80.0,
			PersonalityTraits: map[string]float64{
				"creativity":  0.7,
				"spontaneity": 0.6,
			},
		}
		temp := backend.calculateTemperature(context)
		if temp < 0 || temp > 2 {
			t.Errorf("Temperature out of range: %f", temp)
		}
	})

	t.Run("TargetWordsCalculation", func(t *testing.T) {
		context := DialogContext{
			CurrentStats: map[string]float64{
				"energy": 60.0,
			},
			PersonalityTraits: map[string]float64{
				"talkativeness": 0.8,
			},
		}
		words := backend.calculateTargetWords(context)
		if words < config.MinWords || words > config.MaxWords {
			t.Errorf("Target words out of range: %d", words)
		}
	})

	t.Run("ResponseScoring", func(t *testing.T) {
		context := DialogContext{
			CurrentStats: map[string]float64{
				"happiness": 85.0,
			},
		}
		score := backend.scoreResponse("I'm so happy to see you!", context, 0.5)
		if score <= 0 {
			t.Error("Score should be positive")
		}
	})

	t.Run("ChainSelection", func(t *testing.T) {
		// Set up a basic global chain
		backend.globalChain = &MarkovChain{
			order:      2,
			states:     make(map[string][]string),
			starters:   []string{"hello", "hi"},
			wordCounts: make(map[string]int),
			totalWords: 10,
		}

		chain := backend.selectChain("unknown_trigger")
		if chain != backend.globalChain {
			t.Error("Should fall back to global chain")
		}
	})

	t.Run("CleanTrainingText", func(t *testing.T) {
		dirtyText := "Hello 😊 world!\n\nThis has   extra spaces."
		cleaned := backend.cleanTrainingText(dirtyText)
		if cleaned == "" {
			t.Error("Cleaned text should not be empty")
		}
		if cleaned == dirtyText {
			t.Error("Text should be cleaned")
		}
	})
}
