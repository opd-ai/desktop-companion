// context.go: Conversation context tracking for enhanced dialog system
// Tracks topics, emotional state, and conversation history for better response generation
// Uses Go stdlib only for string analysis and context management

package dialog

import (
	"context"
	"strings"
	"time"
)

// ConversationTopic represents a detected conversation topic with confidence
type ConversationTopic struct {
	Name       string    `json:"name"`       // Topic name (e.g., "weather", "feelings", "activities")
	Confidence float64   `json:"confidence"` // Confidence score 0.0-1.0
	LastSeen   time.Time `json:"last_seen"`  // When this topic was last detected
}

// EmotionalState tracks the emotional context of conversation
type EmotionalState struct {
	Valence   float64   `json:"valence"`   // Positive/negative emotion (-1.0 to 1.0)
	Arousal   float64   `json:"arousal"`   // Energy level (0.0 to 1.0)
	Dominance float64   `json:"dominance"` // Control/confidence (0.0 to 1.0)
	Updated   time.Time `json:"updated"`   // When emotional state was last updated
}

// ConversationContext maintains conversation state for enhanced dialog generation
// Tracks topics, emotional state, and recent message history for context-aware responses
type ConversationContext struct {
	Topics         []ConversationTopic `json:"topics"`          // Detected conversation topics
	EmotionalState EmotionalState      `json:"emotional_state"` // Current emotional context
	RecentMessages []string            `json:"recent_messages"` // Last N messages for context
	MaxHistory     int                 `json:"max_history"`     // Maximum messages to remember
}

// NewConversationContext creates a new conversation context with default settings
func NewConversationContext() *ConversationContext {
	return &ConversationContext{
		Topics:         make([]ConversationTopic, 0),
		EmotionalState: EmotionalState{Updated: time.Now()},
		RecentMessages: make([]string, 0),
		MaxHistory:     10, // Keep last 10 messages
	}
}

// AddMessage processes a new message and updates conversation context
func (cc *ConversationContext) AddMessage(ctx context.Context, message string) error {
	if message == "" {
		return nil // Ignore empty messages
	}

	// Add to recent messages history
	cc.RecentMessages = append(cc.RecentMessages, message)
	if len(cc.RecentMessages) > cc.MaxHistory {
		cc.RecentMessages = cc.RecentMessages[1:] // Remove oldest
	}

	// Update topics based on message content
	cc.updateTopics(message)

	// Update emotional state based on message sentiment
	cc.updateEmotionalState(message)

	return nil
}

// GetActiveTopics returns topics seen recently (within last 5 minutes)
func (cc *ConversationContext) GetActiveTopics() []ConversationTopic {
	cutoff := time.Now().Add(-5 * time.Minute)
	active := make([]ConversationTopic, 0)

	for _, topic := range cc.Topics {
		if topic.LastSeen.After(cutoff) {
			active = append(active, topic)
		}
	}

	return active
}

// GetContextSummary returns a brief summary of current conversation state
func (cc *ConversationContext) GetContextSummary() string {
	if len(cc.RecentMessages) == 0 {
		return "new conversation"
	}

	activeTopics := cc.GetActiveTopics()
	if len(activeTopics) == 0 {
		return "general conversation"
	}

	// Return highest confidence topic
	topTopic := activeTopics[0]
	for _, topic := range activeTopics {
		if topic.Confidence > topTopic.Confidence {
			topTopic = topic
		}
	}

	return "discussing " + topTopic.Name
}

// updateTopics analyzes message content and updates topic tracking
func (cc *ConversationContext) updateTopics(message string) {
	lower := strings.ToLower(message)
	now := time.Now()

	// Simple keyword-based topic detection
	topicKeywords := map[string][]string{
		"weather":    {"weather", "rain", "sunny", "cloudy", "temperature", "hot", "cold"},
		"feelings":   {"feel", "happy", "sad", "angry", "excited", "worried", "love", "hate"},
		"activities": {"doing", "playing", "working", "studying", "watching", "reading"},
		"food":       {"eat", "hungry", "food", "cook", "meal", "drink", "taste"},
		"health":     {"sick", "tired", "energy", "sleep", "pain", "doctor", "medicine"},
	}

	for topicName, keywords := range topicKeywords {
		confidence := 0.0
		for _, keyword := range keywords {
			if strings.Contains(lower, keyword) {
				confidence += 0.2 // Simple scoring
			}
		}

		if confidence > 0 {
			cc.updateTopic(topicName, confidence, now)
		}
	}
}

// updateTopic updates or adds a topic with new confidence score
func (cc *ConversationContext) updateTopic(name string, confidence float64, timestamp time.Time) {
	// Find existing topic
	for i, topic := range cc.Topics {
		if topic.Name == name {
			// Update existing topic with decayed confidence
			decay := 0.8 // Previous confidence decay factor
			cc.Topics[i].Confidence = decay*topic.Confidence + confidence
			if cc.Topics[i].Confidence > 1.0 {
				cc.Topics[i].Confidence = 1.0
			}
			cc.Topics[i].LastSeen = timestamp
			return
		}
	}

	// Add new topic
	cc.Topics = append(cc.Topics, ConversationTopic{
		Name:       name,
		Confidence: confidence,
		LastSeen:   timestamp,
	})
}

// updateEmotionalState analyzes message sentiment and updates emotional context
func (cc *ConversationContext) updateEmotionalState(message string) {
	lower := strings.ToLower(message)

	// Simple sentiment analysis using keyword lists
	positive := []string{"happy", "great", "good", "love", "amazing", "wonderful", "excited"}
	negative := []string{"sad", "bad", "hate", "terrible", "awful", "angry", "worried"}
	energetic := []string{"excited", "energetic", "hyper", "active", "running", "dancing"}
	calm := []string{"calm", "peaceful", "relaxed", "quiet", "sleepy", "tired"}

	valenceChange := 0.0
	arousalChange := 0.0

	for _, word := range positive {
		if strings.Contains(lower, word) {
			valenceChange += 0.1
		}
	}
	for _, word := range negative {
		if strings.Contains(lower, word) {
			valenceChange -= 0.1
		}
	}
	for _, word := range energetic {
		if strings.Contains(lower, word) {
			arousalChange += 0.1
		}
	}
	for _, word := range calm {
		if strings.Contains(lower, word) {
			arousalChange -= 0.1
		}
	}

	// Apply changes with decay
	decay := 0.9
	cc.EmotionalState.Valence = decay*cc.EmotionalState.Valence + valenceChange
	cc.EmotionalState.Arousal = decay*cc.EmotionalState.Arousal + arousalChange

	// Clamp values to valid ranges
	if cc.EmotionalState.Valence > 1.0 {
		cc.EmotionalState.Valence = 1.0
	}
	if cc.EmotionalState.Valence < -1.0 {
		cc.EmotionalState.Valence = -1.0
	}
	if cc.EmotionalState.Arousal > 1.0 {
		cc.EmotionalState.Arousal = 1.0
	}
	if cc.EmotionalState.Arousal < 0.0 {
		cc.EmotionalState.Arousal = 0.0
	}

	cc.EmotionalState.Updated = time.Now()
}
