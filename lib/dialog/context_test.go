package dialog

import (
	"context"
	"testing"
	"time"
)

func TestNewConversationContext(t *testing.T) {
	ctx := NewConversationContext()

	if ctx == nil {
		t.Fatal("NewConversationContext returned nil")
	}

	if len(ctx.Topics) != 0 {
		t.Errorf("Expected empty topics, got %d", len(ctx.Topics))
	}

	if len(ctx.RecentMessages) != 0 {
		t.Errorf("Expected empty messages, got %d", len(ctx.RecentMessages))
	}

	if ctx.MaxHistory != 10 {
		t.Errorf("Expected MaxHistory=10, got %d", ctx.MaxHistory)
	}

	if ctx.EmotionalState.Updated.IsZero() {
		t.Error("Expected EmotionalState.Updated to be set")
	}
}

func TestAddMessage(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		wantErr  bool
		checkLen int
	}{
		{
			name:     "empty message",
			message:  "",
			wantErr:  false,
			checkLen: 0,
		},
		{
			name:     "normal message",
			message:  "Hello there!",
			wantErr:  false,
			checkLen: 1,
		},
		{
			name:     "weather message",
			message:  "It's really sunny today",
			wantErr:  false,
			checkLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewConversationContext()
			err := ctx.AddMessage(context.Background(), tt.message)

			if (err != nil) != tt.wantErr {
				t.Errorf("AddMessage() error = %v, wantErr %v", err, tt.wantErr)
			}

			if len(ctx.RecentMessages) != tt.checkLen {
				t.Errorf("Expected %d messages, got %d", tt.checkLen, len(ctx.RecentMessages))
			}

			if tt.checkLen > 0 && ctx.RecentMessages[0] != tt.message {
				t.Errorf("Expected message %q, got %q", tt.message, ctx.RecentMessages[0])
			}
		})
	}
}

func TestMessageHistory(t *testing.T) {
	ctx := NewConversationContext()
	ctx.MaxHistory = 3 // Set small limit for testing

	messages := []string{"msg1", "msg2", "msg3", "msg4", "msg5"}

	for _, msg := range messages {
		err := ctx.AddMessage(context.Background(), msg)
		if err != nil {
			t.Fatalf("AddMessage failed: %v", err)
		}
	}

	if len(ctx.RecentMessages) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(ctx.RecentMessages))
	}

	// Should contain last 3 messages
	expected := []string{"msg3", "msg4", "msg5"}
	for i, expected := range expected {
		if ctx.RecentMessages[i] != expected {
			t.Errorf("Expected message[%d]=%q, got %q", i, expected, ctx.RecentMessages[i])
		}
	}
}

func TestTopicDetection(t *testing.T) {
	tests := []struct {
		name          string
		message       string
		expectedTopic string
	}{
		{
			name:          "weather topic",
			message:       "The weather is really sunny today",
			expectedTopic: "weather",
		},
		{
			name:          "feelings topic",
			message:       "I feel really happy about this",
			expectedTopic: "feelings",
		},
		{
			name:          "activities topic",
			message:       "I've been working on a project",
			expectedTopic: "activities",
		},
		{
			name:          "food topic",
			message:       "I'm hungry, let's eat something",
			expectedTopic: "food",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewConversationContext()
			err := ctx.AddMessage(context.Background(), tt.message)
			if err != nil {
				t.Fatalf("AddMessage failed: %v", err)
			}

			found := false
			for _, topic := range ctx.Topics {
				if topic.Name == tt.expectedTopic && topic.Confidence > 0 {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Expected topic %q not found in %+v", tt.expectedTopic, ctx.Topics)
			}
		})
	}
}

func TestGetActiveTopics(t *testing.T) {
	ctx := NewConversationContext()

	// Add message to create topic
	err := ctx.AddMessage(context.Background(), "The weather is sunny")
	if err != nil {
		t.Fatalf("AddMessage failed: %v", err)
	}

	// Should have active topics
	active := ctx.GetActiveTopics()
	if len(active) == 0 {
		t.Error("Expected active topics, got none")
	}

	// Make topic old by manipulating timestamp
	if len(ctx.Topics) > 0 {
		ctx.Topics[0].LastSeen = time.Now().Add(-10 * time.Minute)
	}

	// Should have no active topics now
	active = ctx.GetActiveTopics()
	if len(active) != 0 {
		t.Errorf("Expected no active topics, got %d", len(active))
	}
}

func TestGetContextSummary(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*ConversationContext)
		expected string
	}{
		{
			name:     "new conversation",
			setup:    func(ctx *ConversationContext) {},
			expected: "new conversation",
		},
		{
			name: "general conversation",
			setup: func(ctx *ConversationContext) {
				ctx.RecentMessages = []string{"hello"}
			},
			expected: "general conversation",
		},
		{
			name: "weather discussion",
			setup: func(ctx *ConversationContext) {
				ctx.AddMessage(context.Background(), "The weather is sunny")
			},
			expected: "discussing weather",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewConversationContext()
			tt.setup(ctx)

			summary := ctx.GetContextSummary()
			if summary != tt.expected {
				t.Errorf("Expected summary %q, got %q", tt.expected, summary)
			}
		})
	}
}

func TestEmotionalStateUpdate(t *testing.T) {
	tests := []struct {
		name           string
		message        string
		expectPositive bool
		expectNegative bool
	}{
		{
			name:           "positive message",
			message:        "I'm so happy and excited!",
			expectPositive: true,
			expectNegative: false,
		},
		{
			name:           "negative message",
			message:        "I feel sad and terrible",
			expectPositive: false,
			expectNegative: true,
		},
		{
			name:           "neutral message",
			message:        "The sky is blue",
			expectPositive: false,
			expectNegative: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewConversationContext()
			initialValence := ctx.EmotionalState.Valence

			err := ctx.AddMessage(context.Background(), tt.message)
			if err != nil {
				t.Fatalf("AddMessage failed: %v", err)
			}

			if tt.expectPositive && ctx.EmotionalState.Valence <= initialValence {
				t.Error("Expected positive valence increase")
			}
			if tt.expectNegative && ctx.EmotionalState.Valence >= initialValence {
				t.Error("Expected negative valence decrease")
			}

			if !ctx.EmotionalState.Updated.After(time.Now().Add(-1 * time.Second)) {
				t.Error("Expected EmotionalState.Updated to be recent")
			}
		})
	}
}

func TestTopicConfidenceUpdates(t *testing.T) {
	ctx := NewConversationContext()

	// Add multiple weather messages
	messages := []string{
		"The weather is sunny",
		"It's really hot today",
		"The temperature is rising",
	}

	var weatherConfidence float64
	for i, msg := range messages {
		err := ctx.AddMessage(context.Background(), msg)
		if err != nil {
			t.Fatalf("AddMessage %d failed: %v", i, err)
		}

		// Find weather topic
		for _, topic := range ctx.Topics {
			if topic.Name == "weather" {
				if i > 0 && topic.Confidence <= weatherConfidence {
					t.Errorf("Expected confidence to increase, got %f <= %f", topic.Confidence, weatherConfidence)
				}
				weatherConfidence = topic.Confidence
				break
			}
		}
	}

	if weatherConfidence == 0 {
		t.Error("Weather topic not found or has zero confidence")
	}
}

func TestEmotionalStateClampingAndRanges(t *testing.T) {
	ctx := NewConversationContext()

	// Test valence clamping
	ctx.EmotionalState.Valence = 2.0 // Above max
	ctx.updateEmotionalState("happy")
	if ctx.EmotionalState.Valence > 1.0 {
		t.Errorf("Valence not clamped to max: %f", ctx.EmotionalState.Valence)
	}

	ctx.EmotionalState.Valence = -2.0 // Below min
	ctx.updateEmotionalState("sad")
	if ctx.EmotionalState.Valence < -1.0 {
		t.Errorf("Valence not clamped to min: %f", ctx.EmotionalState.Valence)
	}

	// Test arousal clamping
	ctx.EmotionalState.Arousal = 2.0 // Above max
	ctx.updateEmotionalState("excited")
	if ctx.EmotionalState.Arousal > 1.0 {
		t.Errorf("Arousal not clamped to max: %f", ctx.EmotionalState.Arousal)
	}

	ctx.EmotionalState.Arousal = -1.0 // Below min
	ctx.updateEmotionalState("calm")
	if ctx.EmotionalState.Arousal < 0.0 {
		t.Errorf("Arousal not clamped to min: %f", ctx.EmotionalState.Arousal)
	}
}

// Benchmark tests for performance validation
func BenchmarkAddMessage(b *testing.B) {
	ctx := NewConversationContext()
	message := "I feel really happy about the sunny weather today"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.AddMessage(context.Background(), message)
	}
}

func BenchmarkGetActiveTopics(b *testing.B) {
	ctx := NewConversationContext()
	ctx.AddMessage(context.Background(), "The weather is sunny and I feel happy")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.GetActiveTopics()
	}
}

func BenchmarkGetContextSummary(b *testing.B) {
	ctx := NewConversationContext()
	ctx.AddMessage(context.Background(), "The weather is really nice today")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.GetContextSummary()
	}
}
