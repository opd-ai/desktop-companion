// quality.go: Dialog quality scoring and improvement feedback system
// Provides conversation summary generation and quality assessment for better dialog experiences
// Uses Go stdlib only for scoring algorithms and text analysis

package dialog

import (
	"fmt"
	"strings"
	"time"
)

// MemoryEntry represents a dialog memory entry for summary generation
// This avoids circular dependency with character package
type MemoryEntry struct {
	Timestamp        time.Time `json:"timestamp"`
	Trigger          string    `json:"trigger"`
	Response         string    `json:"response"`
	EmotionalTone    string    `json:"emotional_tone"`
	Topics           []string  `json:"topics"`
	MemoryImportance float64   `json:"memory_importance"`
	BackendUsed      string    `json:"backend_used"`
	Confidence       float64   `json:"confidence"`
}
type QualityMetrics struct {
	Coherence       float64 `json:"coherence"`        // How well the response makes sense in context (0-1)
	Relevance       float64 `json:"relevance"`        // How relevant the response is to the input (0-1)
	Engagement      float64 `json:"engagement"`       // How engaging the response is (0-1)
	Personality     float64 `json:"personality"`      // How well it matches character personality (0-1)
	OverallQuality  float64 `json:"overall_quality"`  // Weighted average of all metrics (0-1)
	ResponseLength  int     `json:"response_length"`  // Length of response in words
	TopicContinuity bool    `json:"topic_continuity"` // Whether response continues conversation topic
	Updated         time.Time `json:"updated"`        // When metrics were calculated
}

// ConversationSummary provides a high-level summary of conversation content and quality
type ConversationSummary struct {
	StartTime       time.Time       `json:"start_time"`
	EndTime         time.Time       `json:"end_time"`
	MessageCount    int             `json:"message_count"`
	DominantTopic   string          `json:"dominant_topic"`
	TopicDistribution map[string]int `json:"topic_distribution"`  // Count of messages per topic
	EmotionalJourney []EmotionalState `json:"emotional_journey"` // Emotional state snapshots
	AverageQuality  QualityMetrics  `json:"average_quality"`
	HighlightMoments []string       `json:"highlight_moments"`   // Best or most important exchanges
	Summary         string          `json:"summary"`             // Natural language summary
}

// QualityAssessment provides real-time dialog quality assessment and improvement suggestions
type QualityAssessment struct {
	context     *ConversationContext
	minMessages int // Minimum messages needed for meaningful assessment
}

// NewQualityAssessment creates a new quality assessment system
func NewQualityAssessment(context *ConversationContext) *QualityAssessment {
	return &QualityAssessment{
		context:     context,
		minMessages: 3, // Need at least 3 messages for context
	}
}

// ScoreResponse evaluates the quality of a dialog response given the conversation context
// Uses simple heuristics to score coherence, relevance, engagement, and personality fit
func (qa *QualityAssessment) ScoreResponse(response DialogResponse, userInput string, personalityTraits map[string]float64) QualityMetrics {
	metrics := QualityMetrics{
		Updated: time.Now(),
	}

	// Ensure we have minimum required data
	if response.Text == "" || userInput == "" {
		return metrics // Return zero scores for invalid input
	}

	responseWords := strings.Fields(response.Text)
	inputWords := strings.Fields(userInput)
	
	metrics.ResponseLength = len(responseWords)

	// Score coherence (0-1): Does the response make grammatical and logical sense?
	metrics.Coherence = qa.scoreCoherence(response.Text)

	// Score relevance (0-1): How relevant is the response to the user input?
	metrics.Relevance = qa.scoreRelevance(response.Text, userInput, inputWords, responseWords)

	// Score engagement (0-1): How engaging and interesting is the response?
	metrics.Engagement = qa.scoreEngagement(response.Text, responseWords)

	// Score personality (0-1): How well does this fit the character personality?
	metrics.Personality = qa.scorePersonality(response.Text, personalityTraits)

	// Check topic continuity
	metrics.TopicContinuity = qa.checkTopicContinuity(response.Topics)

	// Calculate overall quality as weighted average
	metrics.OverallQuality = qa.calculateOverallQuality(metrics)

	return metrics
}

// scoreCoherence evaluates grammatical correctness and logical flow
func (qa *QualityAssessment) scoreCoherence(text string) float64 {
	if text == "" {
		return 0.0
	}

	score := 0.7 // Base score for non-empty text

	// Check for basic sentence structure
	if strings.Contains(text, ".") || strings.Contains(text, "!") || strings.Contains(text, "?") {
		score += 0.1 // Has proper sentence endings
	}

	// Check for reasonable length (not too short or too long)
	words := len(strings.Fields(text))
	if words >= 3 && words <= 25 {
		score += 0.1 // Reasonable length
	}

	// Penalty for repetitive words
	wordCount := make(map[string]int)
	for _, word := range strings.Fields(strings.ToLower(text)) {
		wordCount[word]++
	}
	
	repetitive := false
	for _, count := range wordCount {
		if count > 3 { // Same word repeated more than 3 times
			repetitive = true
			break
		}
	}
	
	if repetitive {
		score -= 0.2
	}

	// Clamp to valid range
	if score > 1.0 {
		score = 1.0
	}
	if score < 0.0 {
		score = 0.0
	}

	return score
}

// scoreRelevance evaluates how well the response addresses the user input
func (qa *QualityAssessment) scoreRelevance(response, input string, inputWords, responseWords []string) float64 {
	if len(inputWords) == 0 || len(responseWords) == 0 {
		return 0.0
	}

	score := 0.3 // Base score for having a response

	// Check for shared words (simple keyword matching)
	inputWordsMap := make(map[string]bool)
	for _, word := range inputWords {
		inputWordsMap[strings.ToLower(word)] = true
	}

	sharedWords := 0
	for _, word := range responseWords {
		if inputWordsMap[strings.ToLower(word)] {
			sharedWords++
		}
	}

	// Score based on percentage of shared words
	if len(inputWords) > 0 {
		sharedRatio := float64(sharedWords) / float64(len(inputWords))
		score += sharedRatio * 0.4
	}

	// Bonus for question responses
	if strings.Contains(input, "?") && len(response) > 10 {
		score += 0.2 // Responding to questions with substantial answers
	}

	// Check if response acknowledges key input concepts
	lowerInput := strings.ToLower(input)
	lowerResponse := strings.ToLower(response)
	
	keyTopics := []string{"weather", "feeling", "love", "happy", "sad", "food", "work", "play"}
	for _, topic := range keyTopics {
		if strings.Contains(lowerInput, topic) && strings.Contains(lowerResponse, topic) {
			score += 0.1
			break
		}
	}

	// Clamp to valid range
	if score > 1.0 {
		score = 1.0
	}
	if score < 0.0 {
		score = 0.0
	}

	return score
}

// scoreEngagement evaluates how engaging and interesting the response is
func (qa *QualityAssessment) scoreEngagement(text string, words []string) float64 {
	if len(words) == 0 {
		return 0.0
	}

	score := 0.4 // Base score

	// Longer responses tend to be more engaging (up to a point)
	wordCount := len(words)
	if wordCount >= 5 && wordCount <= 20 {
		score += 0.2
	} else if wordCount > 20 && wordCount <= 30 {
		score += 0.1
	}

	// Check for engaging language elements
	lowerText := strings.ToLower(text)
	
	// Emotional expressions
	emotions := []string{"excited", "love", "amazing", "wonderful", "fantastic", "great"}
	for _, emotion := range emotions {
		if strings.Contains(lowerText, emotion) {
			score += 0.1
			break
		}
	}

	// Questions engage the user
	if strings.Contains(text, "?") {
		score += 0.1
	}

	// Exclamation points show energy
	if strings.Contains(text, "!") {
		score += 0.1
	}

	// Personal pronouns create connection
	pronouns := []string{"you", "we", "us", "your"}
	for _, pronoun := range pronouns {
		if strings.Contains(lowerText, pronoun) {
			score += 0.1
			break
		}
	}

	// Clamp to valid range
	if score > 1.0 {
		score = 1.0
	}
	if score < 0.0 {
		score = 0.0
	}

	return score
}

// scorePersonality evaluates how well the response matches character personality traits
func (qa *QualityAssessment) scorePersonality(text string, traits map[string]float64) float64 {
	if len(traits) == 0 {
		return 0.5 // Neutral score if no personality data
	}

	score := 0.5 // Base neutral score
	lowerText := strings.ToLower(text)

	// Check personality trait expression
	if shyness, exists := traits["shyness"]; exists {
		if shyness > 0.7 { // Very shy character
			// Shy characters use shorter, less bold responses
			words := len(strings.Fields(text))
			if words <= 8 {
				score += 0.2
			}
			// Shy characters avoid exclamations
			if !strings.Contains(text, "!") {
				score += 0.1
			}
		} else if shyness < 0.3 { // Outgoing character
			// Outgoing characters use longer, more expressive responses
			words := len(strings.Fields(text))
			if words > 8 {
				score += 0.1
			}
			if strings.Contains(text, "!") {
				score += 0.1
			}
		}
	}

	if romanticism, exists := traits["romanticism"]; exists {
		if romanticism > 0.7 { // Very romantic character
			romanticWords := []string{"love", "heart", "dear", "darling", "sweet", "beautiful"}
			for _, word := range romanticWords {
				if strings.Contains(lowerText, word) {
					score += 0.1
					break
				}
			}
		}
	}

	if flirtiness, exists := traits["flirtiness"]; exists {
		if flirtiness > 0.7 { // Very flirty character
			flirtyWords := []string{"cute", "handsome", "gorgeous", "charming", "wink"}
			for _, word := range flirtyWords {
				if strings.Contains(lowerText, word) {
					score += 0.1
					break
				}
			}
		}
	}

	// Clamp to valid range
	if score > 1.0 {
		score = 1.0
	}
	if score < 0.0 {
		score = 0.0
	}

	return score
}

// checkTopicContinuity determines if the response continues the current conversation topic
func (qa *QualityAssessment) checkTopicContinuity(responseTopics []string) bool {
	if qa.context == nil || len(responseTopics) == 0 {
		return false
	}

	activeTopics := qa.context.GetActiveTopics()
	if len(activeTopics) == 0 {
		return false // No active topics to continue
	}

	// Check if any response topic matches an active topic
	activeTopicNames := make(map[string]bool)
	for _, topic := range activeTopics {
		activeTopicNames[topic.Name] = true
	}

	for _, responseTopic := range responseTopics {
		if activeTopicNames[responseTopic] {
			return true
		}
	}

	return false
}

// calculateOverallQuality computes weighted average of all quality metrics
func (qa *QualityAssessment) calculateOverallQuality(metrics QualityMetrics) float64 {
	// Weighted average with emphasis on coherence and relevance
	weights := map[string]float64{
		"coherence":  0.3,
		"relevance":  0.3,
		"engagement": 0.2,
		"personality": 0.2,
	}

	overall := weights["coherence"]*metrics.Coherence +
		weights["relevance"]*metrics.Relevance +
		weights["engagement"]*metrics.Engagement +
		weights["personality"]*metrics.Personality

	// Bonus for topic continuity
	if metrics.TopicContinuity {
		overall += 0.05
	}

	// Clamp to valid range
	if overall > 1.0 {
		overall = 1.0
	}
	if overall < 0.0 {
		overall = 0.0
	}

	return overall
}

// GenerateConversationSummary creates a comprehensive summary of the conversation
func (qa *QualityAssessment) GenerateConversationSummary(dialogMemories []MemoryEntry) ConversationSummary {
	summary := ConversationSummary{
		TopicDistribution: make(map[string]int),
		EmotionalJourney:  make([]EmotionalState, 0),
		HighlightMoments:  make([]string, 0),
	}

	if len(dialogMemories) == 0 {
		summary.Summary = "No conversation recorded"
		return summary
	}

	// Set time boundaries
	summary.StartTime = dialogMemories[0].Timestamp
	summary.EndTime = dialogMemories[len(dialogMemories)-1].Timestamp
	summary.MessageCount = len(dialogMemories)

	// Analyze topic distribution
	topicCounts := make(map[string]int)
	for _, memory := range dialogMemories {
		for _, topic := range memory.Topics {
			topicCounts[topic]++
			summary.TopicDistribution[topic]++
		}
	}

	// Find dominant topic
	maxCount := 0
	for topic, count := range topicCounts {
		if count > maxCount {
			maxCount = count
			summary.DominantTopic = topic
		}
	}

	// Collect highlight moments (high-quality responses)
	for _, memory := range dialogMemories {
		if memory.Confidence > 0.8 || memory.MemoryImportance > 0.8 {
			highlight := fmt.Sprintf("%s: %s", memory.Trigger, memory.Response)
			if len(highlight) <= 100 { // Keep highlights concise
				summary.HighlightMoments = append(summary.HighlightMoments, highlight)
			}
		}
	}

	// Limit highlights to prevent overflow
	if len(summary.HighlightMoments) > 5 {
		summary.HighlightMoments = summary.HighlightMoments[:5]
	}

	// Generate natural language summary
	summary.Summary = qa.generateNaturalLanguageSummary(summary)

	return summary
}

// generateNaturalLanguageSummary creates a human-readable conversation summary
func (qa *QualityAssessment) generateNaturalLanguageSummary(summary ConversationSummary) string {
	if summary.MessageCount == 0 {
		return "No conversation recorded"
	}

	var parts []string

	// Basic conversation info
	duration := summary.EndTime.Sub(summary.StartTime)
	if duration > time.Minute {
		parts = append(parts, fmt.Sprintf("Conversation lasted %v with %d exchanges", 
			duration.Round(time.Minute), summary.MessageCount))
	} else {
		parts = append(parts, fmt.Sprintf("Brief conversation with %d exchanges", summary.MessageCount))
	}

	// Dominant topic
	if summary.DominantTopic != "" {
		parts = append(parts, fmt.Sprintf("Mainly discussed %s", summary.DominantTopic))
	}

	// Topic variety
	if len(summary.TopicDistribution) > 2 {
		parts = append(parts, fmt.Sprintf("Covered %d different topics", len(summary.TopicDistribution)))
	}

	// Highlights
	if len(summary.HighlightMoments) > 0 {
		parts = append(parts, fmt.Sprintf("Featured %d memorable moments", len(summary.HighlightMoments)))
	}

	return strings.Join(parts, ". ") + "."
}

// GetImprovementSuggestions provides feedback for improving dialog quality
func (qa *QualityAssessment) GetImprovementSuggestions(metrics QualityMetrics) []string {
	suggestions := make([]string, 0)

	if metrics.Coherence < 0.6 {
		suggestions = append(suggestions, "Improve response clarity and grammatical structure")
	}

	if metrics.Relevance < 0.6 {
		suggestions = append(suggestions, "Focus more on addressing the user's input directly")
	}

	if metrics.Engagement < 0.6 {
		suggestions = append(suggestions, "Make responses more engaging with questions or emotional expressions")
	}

	if metrics.Personality < 0.6 {
		suggestions = append(suggestions, "Better express character personality traits in responses")
	}

	if !metrics.TopicContinuity {
		suggestions = append(suggestions, "Try to maintain conversation topic continuity")
	}

	if metrics.ResponseLength < 3 {
		suggestions = append(suggestions, "Consider providing more detailed responses")
	} else if metrics.ResponseLength > 30 {
		suggestions = append(suggestions, "Consider making responses more concise")
	}

	if len(suggestions) == 0 {
		suggestions = append(suggestions, "Quality is good - maintain current response style")
	}

	return suggestions
}
