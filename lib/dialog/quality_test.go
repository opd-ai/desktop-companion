package dialog

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestNewQualityAssessment(t *testing.T) {
	context := NewConversationContext()
	qa := NewQualityAssessment(context)

	if qa == nil {
		t.Fatal("NewQualityAssessment returned nil")
	}

	if qa.context != context {
		t.Error("Context not properly assigned")
	}

	if qa.minMessages != 3 {
		t.Errorf("Expected minMessages=3, got %d", qa.minMessages)
	}
}

func TestScoreResponse(t *testing.T) {
	context := NewConversationContext()
	qa := NewQualityAssessment(context)

	tests := []struct {
		name              string
		response          DialogResponse
		userInput         string
		personalityTraits map[string]float64
		expectNonZero     bool
	}{
		{
			name: "empty response",
			response: DialogResponse{
				Text: "",
			},
			userInput:     "Hello",
			expectNonZero: false,
		},
		{
			name: "good response",
			response: DialogResponse{
				Text:   "Hello there! How are you feeling today?",
				Topics: []string{"feelings"},
			},
			userInput:     "Hello",
			expectNonZero: true,
		},
		{
			name: "personality-driven response",
			response: DialogResponse{
				Text:   "Oh, um... hello there...",
				Topics: []string{"greeting"},
			},
			userInput: "Hello",
			personalityTraits: map[string]float64{
				"shyness": 0.8,
			},
			expectNonZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := qa.ScoreResponse(tt.response, tt.userInput, tt.personalityTraits)

			if tt.expectNonZero {
				if metrics.OverallQuality == 0.0 {
					t.Error("Expected non-zero overall quality")
				}
				if metrics.Updated.IsZero() {
					t.Error("Expected Updated timestamp to be set")
				}
			}

			// Validate ranges
			if metrics.Coherence < 0.0 || metrics.Coherence > 1.0 {
				t.Errorf("Coherence out of range: %f", metrics.Coherence)
			}
			if metrics.Relevance < 0.0 || metrics.Relevance > 1.0 {
				t.Errorf("Relevance out of range: %f", metrics.Relevance)
			}
			if metrics.Engagement < 0.0 || metrics.Engagement > 1.0 {
				t.Errorf("Engagement out of range: %f", metrics.Engagement)
			}
			if metrics.Personality < 0.0 || metrics.Personality > 1.0 {
				t.Errorf("Personality out of range: %f", metrics.Personality)
			}
			if metrics.OverallQuality < 0.0 || metrics.OverallQuality > 1.0 {
				t.Errorf("OverallQuality out of range: %f", metrics.OverallQuality)
			}
		})
	}
}

func TestScoreCoherence(t *testing.T) {
	context := NewConversationContext()
	qa := NewQualityAssessment(context)

	tests := []struct {
		name     string
		text     string
		expected float64
		operator string // "gt", "lt", "eq"
	}{
		{
			name:     "empty text",
			text:     "",
			expected: 0.0,
			operator: "eq",
		},
		{
			name:     "good sentence",
			text:     "Hello! How are you doing today?",
			expected: 0.8,
			operator: "gt",
		},
		{
			name:     "repetitive text",
			text:     "hello hello hello hello hello",
			expected: 0.7,
			operator: "lt",
		},
		{
			name:     "reasonable length",
			text:     "This is a well-formed sentence with good structure.",
			expected: 0.8,
			operator: "gt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := qa.scoreCoherence(tt.text)

			switch tt.operator {
			case "gt":
				if score <= tt.expected {
					t.Errorf("Expected score > %f, got %f", tt.expected, score)
				}
			case "lt":
				if score >= tt.expected {
					t.Errorf("Expected score < %f, got %f", tt.expected, score)
				}
			case "eq":
				if score != tt.expected {
					t.Errorf("Expected score = %f, got %f", tt.expected, score)
				}
			}

			// Always validate range
			if score < 0.0 || score > 1.0 {
				t.Errorf("Score out of range: %f", score)
			}
		})
	}
}

func TestScoreRelevance(t *testing.T) {
	context := NewConversationContext()
	qa := NewQualityAssessment(context)

	tests := []struct {
		name     string
		response string
		input    string
		expected float64
		operator string
	}{
		{
			name:     "no input",
			response: "Hello",
			input:    "",
			expected: 0.0,
			operator: "eq",
		},
		{
			name:     "shared keywords",
			response: "I love the sunny weather too!",
			input:    "The weather is sunny today",
			expected: 0.3,
			operator: "gt",
		},
		{
			name:     "question response",
			response: "I'm doing great, thanks for asking!",
			input:    "How are you?",
			expected: 0.4,
			operator: "gt",
		},
		{
			name:     "topic acknowledgment",
			response: "Yes, I love that food too!",
			input:    "Do you like this food?",
			expected: 0.5,
			operator: "gt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputWords := []string{}
			responseWords := []string{}
			if tt.input != "" {
				inputWords = []string{tt.input}
			}
			if tt.response != "" {
				responseWords = []string{tt.response}
			}

			score := qa.scoreRelevance(tt.response, tt.input, inputWords, responseWords)

			switch tt.operator {
			case "gt":
				if score <= tt.expected {
					t.Errorf("Expected score > %f, got %f", tt.expected, score)
				}
			case "lt":
				if score >= tt.expected {
					t.Errorf("Expected score < %f, got %f", tt.expected, score)
				}
			case "eq":
				if score != tt.expected {
					t.Errorf("Expected score = %f, got %f", tt.expected, score)
				}
			}

			// Always validate range
			if score < 0.0 || score > 1.0 {
				t.Errorf("Score out of range: %f", score)
			}
		})
	}
}

func TestScoreEngagement(t *testing.T) {
	context := NewConversationContext()
	qa := NewQualityAssessment(context)

	tests := []struct {
		name     string
		text     string
		expected float64
		operator string
	}{
		{
			name:     "empty text",
			text:     "",
			expected: 0.0,
			operator: "eq",
		},
		{
			name:     "engaging with emotion",
			text:     "I'm so excited to see you!",
			expected: 0.6,
			operator: "gt",
		},
		{
			name:     "question engagement",
			text:     "How was your day?",
			expected: 0.5,
			operator: "gt",
		},
		{
			name:     "personal pronouns",
			text:     "You and I should hang out together!",
			expected: 0.5,
			operator: "gt",
		},
		{
			name:     "optimal length",
			text:     "This is a well-sized response that should score well for engagement.",
			expected: 0.4,
			operator: "gt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			words := []string{}
			if tt.text != "" {
				words = []string{tt.text}
			}

			score := qa.scoreEngagement(tt.text, words)

			switch tt.operator {
			case "gt":
				if score <= tt.expected {
					t.Errorf("Expected score > %f, got %f", tt.expected, score)
				}
			case "lt":
				if score >= tt.expected {
					t.Errorf("Expected score < %f, got %f", tt.expected, score)
				}
			case "eq":
				if score != tt.expected {
					t.Errorf("Expected score = %f, got %f", tt.expected, score)
				}
			}

			// Always validate range
			if score < 0.0 || score > 1.0 {
				t.Errorf("Score out of range: %f", score)
			}
		})
	}
}

func TestScorePersonality(t *testing.T) {
	context := NewConversationContext()
	qa := NewQualityAssessment(context)

	tests := []struct {
		name     string
		text     string
		traits   map[string]float64
		expected float64
		operator string
	}{
		{
			name:     "no traits",
			text:     "Hello there!",
			traits:   map[string]float64{},
			expected: 0.5,
			operator: "eq",
		},
		{
			name: "shy character - short response",
			text: "Hi...",
			traits: map[string]float64{
				"shyness": 0.8,
			},
			expected: 0.6,
			operator: "gt",
		},
		{
			name: "romantic character",
			text: "You're so beautiful, my love!",
			traits: map[string]float64{
				"romanticism": 0.8,
			},
			expected: 0.5,
			operator: "gt",
		},
		{
			name: "flirty character",
			text: "You look cute today! *wink*",
			traits: map[string]float64{
				"flirtiness": 0.9,
			},
			expected: 0.5,
			operator: "gt",
		},
		{
			name: "outgoing character - long response",
			text: "Hello there! I'm so excited to see you today! How are you doing?",
			traits: map[string]float64{
				"shyness": 0.2,
			},
			expected: 0.6,
			operator: "gt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := qa.scorePersonality(tt.text, tt.traits)

			switch tt.operator {
			case "gt":
				if score <= tt.expected {
					t.Errorf("Expected score > %f, got %f", tt.expected, score)
				}
			case "lt":
				if score >= tt.expected {
					t.Errorf("Expected score < %f, got %f", tt.expected, score)
				}
			case "eq":
				if score != tt.expected {
					t.Errorf("Expected score = %f, got %f", tt.expected, score)
				}
			}

			// Always validate range
			if score < 0.0 || score > 1.0 {
				t.Errorf("Score out of range: %f", score)
			}
		})
	}
}

func TestCheckTopicContinuity(t *testing.T) {
	context := NewConversationContext()
	qa := NewQualityAssessment(context)

	// Add some topics to context
	context.AddMessage(nil, "The weather is really sunny today")

	tests := []struct {
		name           string
		responseTopics []string
		expected       bool
	}{
		{
			name:           "no response topics",
			responseTopics: []string{},
			expected:       false,
		},
		{
			name:           "matching topic",
			responseTopics: []string{"weather"},
			expected:       true,
		},
		{
			name:           "non-matching topic",
			responseTopics: []string{"food"},
			expected:       false,
		},
		{
			name:           "mixed topics",
			responseTopics: []string{"food", "weather", "activities"},
			expected:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := qa.checkTopicContinuity(tt.responseTopics)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGenerateConversationSummary(t *testing.T) {
	context := NewConversationContext()
	qa := NewQualityAssessment(context)

	now := time.Now()
	memories := []MemoryEntry{
		{
			Timestamp:        now.Add(-10 * time.Minute),
			Trigger:          "click",
			Response:         "Hello! How are you?",
			EmotionalTone:    "happy",
			Topics:           []string{"greeting"},
			MemoryImportance: 0.5,
			BackendUsed:      "markov_chain",
			Confidence:       0.8,
		},
		{
			Timestamp:        now.Add(-8 * time.Minute),
			Trigger:          "chat",
			Response:         "The weather is beautiful today!",
			EmotionalTone:    "excited",
			Topics:           []string{"weather"},
			MemoryImportance: 0.9,
			BackendUsed:      "markov_chain",
			Confidence:       0.9,
		},
		{
			Timestamp:        now.Add(-5 * time.Minute),
			Trigger:          "chat",
			Response:         "I love sunny days like this!",
			EmotionalTone:    "happy",
			Topics:           []string{"weather", "feelings"},
			MemoryImportance: 0.7,
			BackendUsed:      "markov_chain",
			Confidence:       0.85,
		},
	}

	summary := qa.GenerateConversationSummary(memories)

	// Test basic properties
	if summary.MessageCount != 3 {
		t.Errorf("Expected MessageCount=3, got %d", summary.MessageCount)
	}

	if summary.DominantTopic != "weather" {
		t.Errorf("Expected DominantTopic=weather, got %s", summary.DominantTopic)
	}

	if len(summary.TopicDistribution) == 0 {
		t.Error("Expected topic distribution to be populated")
	}

	if summary.TopicDistribution["weather"] != 2 {
		t.Errorf("Expected weather count=2, got %d", summary.TopicDistribution["weather"])
	}

	// Test highlight moments (high confidence responses)
	if len(summary.HighlightMoments) == 0 {
		t.Error("Expected some highlight moments")
	}

	// Test summary text
	if summary.Summary == "" {
		t.Error("Expected non-empty summary text")
	}

	if !strings.Contains(summary.Summary, "weather") {
		t.Error("Expected summary to mention dominant topic 'weather'")
	}

	// Test time boundaries
	expectedDuration := memories[2].Timestamp.Sub(memories[0].Timestamp)
	actualDuration := summary.EndTime.Sub(summary.StartTime)
	if actualDuration != expectedDuration {
		t.Errorf("Expected duration %v, got %v", expectedDuration, actualDuration)
	}
}

func TestGenerateConversationSummaryEmpty(t *testing.T) {
	context := NewConversationContext()
	qa := NewQualityAssessment(context)

	summary := qa.GenerateConversationSummary([]MemoryEntry{})

	if summary.MessageCount != 0 {
		t.Errorf("Expected MessageCount=0, got %d", summary.MessageCount)
	}

	if summary.Summary != "No conversation recorded" {
		t.Errorf("Expected empty summary message, got %s", summary.Summary)
	}
}

func TestGetImprovementSuggestions(t *testing.T) {
	context := NewConversationContext()
	qa := NewQualityAssessment(context)

	tests := []struct {
		name     string
		metrics  QualityMetrics
		expected []string
	}{
		{
			name: "good quality",
			metrics: QualityMetrics{
				Coherence:       0.8,
				Relevance:       0.8,
				Engagement:      0.8,
				Personality:     0.8,
				ResponseLength:  10,
				TopicContinuity: true,
			},
			expected: []string{"Quality is good - maintain current response style"},
		},
		{
			name: "low coherence",
			metrics: QualityMetrics{
				Coherence:       0.4,
				Relevance:       0.8,
				Engagement:      0.8,
				Personality:     0.8,
				ResponseLength:  10,
				TopicContinuity: true,
			},
			expected: []string{"Improve response clarity and grammatical structure"},
		},
		{
			name: "low relevance",
			metrics: QualityMetrics{
				Coherence:       0.8,
				Relevance:       0.4,
				Engagement:      0.8,
				Personality:     0.8,
				ResponseLength:  10,
				TopicContinuity: true,
			},
			expected: []string{"Focus more on addressing the user's input directly"},
		},
		{
			name: "short response",
			metrics: QualityMetrics{
				Coherence:       0.8,
				Relevance:       0.8,
				Engagement:      0.8,
				Personality:     0.8,
				ResponseLength:  2,
				TopicContinuity: true,
			},
			expected: []string{"Consider providing more detailed responses"},
		},
		{
			name: "long response",
			metrics: QualityMetrics{
				Coherence:       0.8,
				Relevance:       0.8,
				Engagement:      0.8,
				Personality:     0.8,
				ResponseLength:  35,
				TopicContinuity: true,
			},
			expected: []string{"Consider making responses more concise"},
		},
		{
			name: "multiple issues",
			metrics: QualityMetrics{
				Coherence:       0.4,
				Relevance:       0.4,
				Engagement:      0.4,
				Personality:     0.4,
				ResponseLength:  10,
				TopicContinuity: false,
			},
			expected: []string{
				"Improve response clarity and grammatical structure",
				"Focus more on addressing the user's input directly",
				"Make responses more engaging with questions or emotional expressions",
				"Better express character personality traits in responses",
				"Try to maintain conversation topic continuity",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := qa.GetImprovementSuggestions(tt.metrics)

			// Check that we have the expected number of suggestions
			if len(suggestions) != len(tt.expected) {
				t.Errorf("Expected %d suggestions, got %d", len(tt.expected), len(suggestions))
			}

			// Check for presence of each expected suggestion
			suggestionMap := make(map[string]bool)
			for _, suggestion := range suggestions {
				suggestionMap[suggestion] = true
			}

			for _, expected := range tt.expected {
				if !suggestionMap[expected] {
					t.Errorf("Expected suggestion '%s' not found in %v", expected, suggestions)
				}
			}
		})
	}
}

// Benchmark tests for performance validation
func BenchmarkScoreResponse(b *testing.B) {
	context := NewConversationContext()
	qa := NewQualityAssessment(context)

	response := DialogResponse{
		Text:   "Hello there! How are you feeling today? I hope you're having a wonderful time!",
		Topics: []string{"greeting", "feelings"},
	}
	userInput := "Hello, how are you?"
	traits := map[string]float64{
		"shyness":     0.3,
		"romanticism": 0.7,
		"flirtiness":  0.5,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		qa.ScoreResponse(response, userInput, traits)
	}
}

func BenchmarkGenerateConversationSummary(b *testing.B) {
	context := NewConversationContext()
	qa := NewQualityAssessment(context)

	now := time.Now()
	memories := make([]MemoryEntry, 50) // Larger dataset
	for i := 0; i < 50; i++ {
		memories[i] = MemoryEntry{
			Timestamp:        now.Add(time.Duration(-i) * time.Minute),
			Trigger:          "chat",
			Response:         fmt.Sprintf("Response %d about weather and feelings", i),
			EmotionalTone:    "happy",
			Topics:           []string{"weather", "feelings"},
			MemoryImportance: 0.5 + float64(i%5)/10.0,
			BackendUsed:      "markov_chain",
			Confidence:       0.7 + float64(i%3)/10.0,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		qa.GenerateConversationSummary(memories)
	}
}

func BenchmarkGetImprovementSuggestions(b *testing.B) {
	context := NewConversationContext()
	qa := NewQualityAssessment(context)

	metrics := QualityMetrics{
		Coherence:       0.4,
		Relevance:       0.6,
		Engagement:      0.3,
		Personality:     0.8,
		ResponseLength:  15,
		TopicContinuity: false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		qa.GetImprovementSuggestions(metrics)
	}
}
