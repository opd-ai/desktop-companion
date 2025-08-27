package character

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// MarkovChainBackend implements DialogBackend using Markov chain text generation
// Follows "lazy programmer" approach: uses simple state transitions without complex NLP
type MarkovChainBackend struct {
	config      MarkovConfig
	chains      map[string]*MarkovChain // Per-trigger chain storage
	globalChain *MarkovChain            // Global chain for fallback
	character   *Character              // Character reference for context
	initialized bool
}

// MarkovConfig defines JSON configuration for the Markov chain backend
type MarkovConfig struct {
	// Chain parameters
	ChainOrder     int     `json:"chainOrder"`     // N-gram size (1=unigram, 2=bigram, etc.)
	MinWords       int     `json:"minWords"`       // Minimum words in generated response
	MaxWords       int     `json:"maxWords"`       // Maximum words in generated response
	TemperatureMin float64 `json:"temperatureMin"` // Minimum randomness (0=deterministic, 1=random)
	TemperatureMax float64 `json:"temperatureMax"` // Maximum randomness based on context

	// Training data sources
	TrainingData     []string `json:"trainingData"`            // Direct text samples for training
	TrainingFiles    []string `json:"trainingFiles,omitempty"` // File paths to training text
	UseDialogHistory bool     `json:"useDialogHistory"`        // Include character's dialog history
	UsePersonality   bool     `json:"usePersonality"`          // Adjust responses based on personality

	// Response filtering and enhancement
	ForbiddenWords   []string `json:"forbiddenWords,omitempty"` // Words to avoid in responses
	RequiredWords    []string `json:"requiredWords,omitempty"`  // Words that should appear more often
	PersonalityBoost float64  `json:"personalityBoost"`         // How much personality affects word selection (0-2)
	MoodInfluence    float64  `json:"moodInfluence"`            // How much mood affects response style (0-2)

	// Context awareness
	TriggerSpecific    bool    `json:"triggerSpecific"`    // Train separate chains per trigger
	StatAwareness      float64 `json:"statAwareness"`      // How much stats influence generation (0-1)
	RelationshipWeight float64 `json:"relationshipWeight"` // Weight relationship level in responses (0-2)
	TimeOfDayWeight    float64 `json:"timeOfDayWeight"`    // Weight time of day in responses (0-1)

	// Memory and learning
	MemoryDecay     float64 `json:"memoryDecay"`     // How quickly old training data is forgotten (0-1)
	LearningRate    float64 `json:"learningRate"`    // How quickly to adapt to new interactions (0-1)
	AdaptationSteps int     `json:"adaptationSteps"` // How many interactions before adaptation

	// Quality control
	CoherenceThreshold float64  `json:"coherenceThreshold"` // Minimum coherence for accepting response (0-1)
	SimilarityPenalty  float64  `json:"similarityPenalty"`  // Penalty for responses too similar to recent ones (0-1)
	FallbackPhrases    []string `json:"fallbackPhrases"`    // High-quality fallback responses

	// Advanced quality filters
	QualityFilters struct {
		MinCoherence    float64 `json:"minCoherence"`    // Minimum coherence score for responses (0-1)
		MaxRepetition   float64 `json:"maxRepetition"`   // Maximum word repetition ratio (0-1)
		RequireComplete bool    `json:"requireComplete"` // Require complete sentences
		GrammarCheck    bool    `json:"grammarCheck"`    // Enable basic grammar validation
		MinUniqueWords  int     `json:"minUniqueWords"`  // Minimum unique words in response
		MaxSimilarity   float64 `json:"maxSimilarity"`   // Maximum similarity to recent responses (0-1)
	} `json:"qualityFilters,omitempty"`
}

// MarkovChain represents a single Markov chain for text generation
type MarkovChain struct {
	order       int
	states      map[string][]string // state -> possible next words
	starters    []string            // possible sentence starters
	wordCounts  map[string]int      // word frequency tracking
	totalWords  int                 // total words processed
	lastUpdated time.Time           // when chain was last updated
}

// MarkovState represents a state in the Markov chain (sequence of words)
type MarkovState struct {
	words []string
	key   string
}

// NewMarkovChainBackend creates a new Markov chain dialog backend
func NewMarkovChainBackend() *MarkovChainBackend {
	return &MarkovChainBackend{
		chains: make(map[string]*MarkovChain),
	}
}

// Initialize sets up the Markov backend with JSON configuration and character context
func (m *MarkovChainBackend) Initialize(config json.RawMessage, character *Character) error {
	// Parse configuration
	if err := json.Unmarshal(config, &m.config); err != nil {
		return fmt.Errorf("failed to parse Markov config: %w", err)
	}

	// Validate configuration
	if err := m.validateConfig(); err != nil {
		return fmt.Errorf("invalid Markov config: %w", err)
	}

	m.character = character

	// Create global chain
	m.globalChain = NewMarkovChain(m.config.ChainOrder)

	// Create trigger-specific chains if enabled
	if m.config.TriggerSpecific {
		triggers := []string{"click", "rightclick", "hover", "compliment", "give_gift", "deep_conversation"}
		for _, trigger := range triggers {
			m.chains[trigger] = NewMarkovChain(m.config.ChainOrder)
		}
	}

	// Train chains with initial data
	if err := m.trainWithInitialData(); err != nil {
		return fmt.Errorf("failed to train initial chains: %w", err)
	}

	m.initialized = true
	return nil
}

// validateConfig ensures Markov configuration is valid
func (m *MarkovChainBackend) validateConfig() error {
	if m.config.ChainOrder < 1 || m.config.ChainOrder > 5 {
		return fmt.Errorf("chainOrder must be 1-5, got %d", m.config.ChainOrder)
	}

	if m.config.MinWords < 1 || m.config.MinWords > m.config.MaxWords {
		return fmt.Errorf("minWords (%d) must be >= 1 and <= maxWords (%d)", m.config.MinWords, m.config.MaxWords)
	}

	if m.config.MaxWords < 1 || m.config.MaxWords > 50 {
		return fmt.Errorf("maxWords must be 1-50, got %d", m.config.MaxWords)
	}

	if m.config.TemperatureMin < 0 || m.config.TemperatureMin > m.config.TemperatureMax || m.config.TemperatureMax > 2 {
		return fmt.Errorf("temperature values must be 0 <= min <= max <= 2")
	}

	return nil
}

// trainWithInitialData trains the Markov chains with configuration-provided data
func (m *MarkovChainBackend) trainWithInitialData() error {
	// Train with direct training data
	for _, text := range m.config.TrainingData {
		if err := m.trainWithText(text, "general"); err != nil {
			return fmt.Errorf("failed to train with text: %w", err)
		}
	}

	// Include character's existing dialogs if configured
	if m.config.UseDialogHistory {
		m.trainWithCharacterDialogs()
	}

	// If no training data provided, use minimal defaults
	if len(m.config.TrainingData) == 0 && !m.config.UseDialogHistory {
		m.trainWithDefaults()
	}

	return nil
}

// trainWithText trains the appropriate chain(s) with the given text
func (m *MarkovChainBackend) trainWithText(text, trigger string) error {
	// Clean and validate text
	cleanText := m.cleanTrainingText(text)
	if len(cleanText) < 3 {
		return nil // Skip very short text
	}

	// Train global chain
	m.globalChain.Train(cleanText)

	// Train trigger-specific chain if applicable
	if m.config.TriggerSpecific {
		if chain, exists := m.chains[trigger]; exists {
			chain.Train(cleanText)
		}
	}

	return nil
}

// trainWithCharacterDialogs includes existing character dialogs in training
func (m *MarkovChainBackend) trainWithCharacterDialogs() {
	// Train with basic dialogs
	for _, dialog := range m.character.card.Dialogs {
		for _, response := range dialog.Responses {
			_ = m.trainWithText(response, dialog.Trigger)
		}
	}

	// Train with romance dialogs if available
	for _, dialog := range m.character.card.RomanceDialogs {
		for _, response := range dialog.Responses {
			_ = m.trainWithText(response, dialog.Trigger)
		}
	}

	// Train with interaction responses
	for interactionType, interaction := range m.character.card.Interactions {
		for _, response := range interaction.Responses {
			_ = m.trainWithText(response, interactionType)
		}
	}
}

// trainWithDefaults provides minimal training data if none is configured
func (m *MarkovChainBackend) trainWithDefaults() {
	defaults := []string{
		"Hello there! How are you doing today?",
		"It's nice to see you again! I've been thinking about you.",
		"Thank you for spending time with me. It means a lot.",
		"I hope you're having a wonderful day! You deserve happiness.",
		"What would you like to talk about? I'm here to listen.",
		"You always know how to make me smile. Thank you for that.",
		"I'm feeling grateful for this moment we're sharing together.",
		"Your presence brightens my day. I'm so glad you're here.",
	}

	for _, text := range defaults {
		_ = m.trainWithText(text, "general")
	}
}

// cleanTrainingText preprocesses text for training
func (m *MarkovChainBackend) cleanTrainingText(text string) string {
	// Remove excessive whitespace and emojis for cleaner chains
	cleaned := strings.ReplaceAll(text, "\n", " ")
	cleaned = strings.TrimSpace(cleaned)

	// Remove emoji characters for simpler text processing
	// In a more sophisticated system, we might preserve emotional markers
	emojiMap := map[string]string{
		"👋": "", "😊": "", "💕": "", "💖": "", "💓": "", "💝": "",
		"🎁": "", "😳": "", "🤗": "", "😢": "", "💔": "", "🥺": "",
		"😘": "", "😄": "", "👀": "", "🖱️": "", "💭": "", "💗": "",
	}

	for emoji := range emojiMap {
		cleaned = strings.ReplaceAll(cleaned, emoji, "")
	}

	return strings.TrimSpace(cleaned)
}

// GenerateResponse produces a dialog response using Markov chain generation
func (m *MarkovChainBackend) GenerateResponse(context DialogContext) (DialogResponse, error) {
	if !m.initialized {
		return DialogResponse{}, fmt.Errorf("backend not initialized")
	}

	// Select appropriate chain
	chain := m.selectChain(context.Trigger)
	if chain == nil {
		return DialogResponse{}, fmt.Errorf("no chain available for trigger: %s", context.Trigger)
	}

	// Calculate generation parameters based on context
	temperature := m.calculateTemperature(context)
	targetWords := m.calculateTargetWords(context)

	// Generate response text
	text, confidence := m.generateWithChain(chain, targetWords, temperature, context)

	// Apply personality and mood adjustments
	text = m.applyPersonalityAdjustments(text, context)

	// Validate and filter response
	if !m.validateResponse(text, context) {
		text = m.selectFallbackResponse(context)
		confidence = 0.3
	}

	// Select appropriate animation
	animation := m.selectAnimation(text, context)

	return DialogResponse{
		Text:             text,
		Animation:        animation,
		Confidence:       confidence,
		ResponseType:     m.classifyResponseType(text, context),
		EmotionalTone:    m.detectEmotionalTone(text, context),
		Topics:           m.extractTopics(text),
		MemoryImportance: m.calculateMemoryImportance(text, context),
		LearningValue:    confidence * 0.8, // High confidence responses are more valuable for learning
	}, nil
}

// selectChain chooses the appropriate Markov chain based on trigger and configuration
func (m *MarkovChainBackend) selectChain(trigger string) *MarkovChain {
	if m.config.TriggerSpecific {
		if chain, exists := m.chains[trigger]; exists && chain.hasEnoughData() {
			return chain
		}
	}
	return m.globalChain
}

// calculateTemperature determines generation randomness based on context
func (m *MarkovChainBackend) calculateTemperature(context DialogContext) float64 {
	baseTemp := (m.config.TemperatureMin + m.config.TemperatureMax) / 2

	// Adjust based on personality traits
	if m.config.UsePersonality {
		creativity := context.PersonalityTraits["creativity"]
		spontaneity := context.PersonalityTraits["spontaneity"]
		baseTemp += (creativity + spontaneity - 1.0) * 0.2 // -0.2 to +0.2 adjustment
	}

	// Adjust based on mood
	if m.config.MoodInfluence > 0 {
		moodFactor := (context.CurrentMood - 50) / 100 // -0.5 to +0.5
		baseTemp += moodFactor * m.config.MoodInfluence * 0.1
	}

	// Clamp to configured range
	if baseTemp < m.config.TemperatureMin {
		baseTemp = m.config.TemperatureMin
	}
	if baseTemp > m.config.TemperatureMax {
		baseTemp = m.config.TemperatureMax
	}

	return baseTemp
}

// calculateTargetWords determines ideal response length based on context
func (m *MarkovChainBackend) calculateTargetWords(context DialogContext) int {
	base := (m.config.MinWords + m.config.MaxWords) / 2

	// Adjust based on trigger type
	switch context.Trigger {
	case "hover":
		base = m.config.MinWords // Hover responses should be brief
	case "deep_conversation":
		base = m.config.MaxWords // Deep conversations can be longer
	case "compliment":
		base = (m.config.MinWords + m.config.MaxWords) / 2
	}

	// Adjust based on relationship level
	if context.RelationshipLevel == "Romantic Interest" {
		base = int(float64(base) * 1.2) // 20% longer for romantic relationships
	}

	// Clamp to configured range
	if base < m.config.MinWords {
		base = m.config.MinWords
	}
	if base > m.config.MaxWords {
		base = m.config.MaxWords
	}

	return base
}

// generateWithChain performs the actual Markov chain text generation
func (m *MarkovChainBackend) generateWithChain(chain *MarkovChain, targetWords int, temperature float64, context DialogContext) (string, float64) {
	maxAttempts := 5
	bestText := ""
	bestScore := 0.0

	for attempt := 0; attempt < maxAttempts; attempt++ {
		text, confidence := chain.Generate(targetWords, temperature)
		score := m.scoreResponse(text, context, confidence)

		if score > bestScore {
			bestScore = score
			bestText = text
		}

		// Accept good enough responses early
		if score > 0.8 {
			break
		}
	}

	return bestText, bestScore
}

// scoreResponse evaluates how well a response fits the context
func (m *MarkovChainBackend) scoreResponse(text string, context DialogContext, baseConfidence float64) float64 {
	score := baseConfidence

	// Length appropriateness
	words := strings.Fields(text)
	targetLength := m.calculateTargetWords(context)
	lengthRatio := float64(len(words)) / float64(targetLength)
	if lengthRatio < 0.5 || lengthRatio > 2.0 {
		score *= 0.8 // Penalty for poor length
	}

	// Check for forbidden words
	for _, forbidden := range m.config.ForbiddenWords {
		if strings.Contains(strings.ToLower(text), strings.ToLower(forbidden)) {
			score *= 0.5 // Heavy penalty for forbidden content
		}
	}

	// Bonus for required words (personality-appropriate terms)
	requiredFound := 0
	for _, required := range m.config.RequiredWords {
		if strings.Contains(strings.ToLower(text), strings.ToLower(required)) {
			requiredFound++
		}
	}
	if len(m.config.RequiredWords) > 0 {
		score *= 1.0 + (float64(requiredFound)/float64(len(m.config.RequiredWords)))*0.2
	}

	// Penalty for similarity to recent responses
	similarity := m.calculateSimilarityToRecent(text, context)
	score *= 1.0 - (similarity * m.config.SimilarityPenalty)

	return score
}

// calculateSimilarityToRecent checks how similar the response is to recent interactions
func (m *MarkovChainBackend) calculateSimilarityToRecent(text string, context DialogContext) float64 {
	if len(context.InteractionHistory) == 0 {
		return 0.0
	}

	words := m.normalizeWords(strings.Fields(text))
	maxSimilarity := 0.0

	// Check against last 3 responses
	checkCount := len(context.InteractionHistory)
	if checkCount > 3 {
		checkCount = 3
	}

	for i := 0; i < checkCount; i++ {
		recentResponse := context.InteractionHistory[i].Response
		recentWords := m.normalizeWords(strings.Fields(recentResponse))
		similarity := m.calculateWordOverlap(words, recentWords)
		if similarity > maxSimilarity {
			maxSimilarity = similarity
		}
	}

	return maxSimilarity
}

// normalizeWords converts words to lowercase for comparison
func (m *MarkovChainBackend) normalizeWords(words []string) []string {
	normalized := make([]string, len(words))
	for i, word := range words {
		normalized[i] = strings.ToLower(strings.Trim(word, ".,!?"))
	}
	return normalized
}

// calculateWordOverlap calculates the percentage of word overlap between two word lists
func (m *MarkovChainBackend) calculateWordOverlap(words1, words2 []string) float64 {
	if len(words1) == 0 || len(words2) == 0 {
		return 0.0
	}

	wordSet1 := make(map[string]bool)
	for _, word := range words1 {
		wordSet1[word] = true
	}

	overlap := 0
	for _, word := range words2 {
		if wordSet1[word] {
			overlap++
		}
	}

	totalWords := len(words1) + len(words2)
	return float64(overlap*2) / float64(totalWords)
}

// applyPersonalityAdjustments modifies response based on character personality
func (m *MarkovChainBackend) applyPersonalityAdjustments(text string, context DialogContext) string {
	if !m.config.UsePersonality || m.config.PersonalityBoost == 0 {
		return text
	}

	// For now, return unchanged - future enhancement could:
	// - Add personality-specific words or phrases
	// - Adjust sentence structure based on traits
	// - Modify punctuation/emphasis based on personality
	return text
}

// validateResponse checks if the response meets quality standards
func (m *MarkovChainBackend) validateResponse(text string, context DialogContext) bool {
	// Basic validation checks
	words := strings.Fields(text)

	// Length check
	if len(words) < m.config.MinWords || len(words) > m.config.MaxWords {
		return false
	}

	// Basic coherence check
	if m.config.CoherenceThreshold > 0 {
		coherence := m.calculateCoherence(text)
		if coherence < m.config.CoherenceThreshold {
			return false
		}
	}

	// Forbidden content check
	for _, forbidden := range m.config.ForbiddenWords {
		if strings.Contains(strings.ToLower(text), strings.ToLower(forbidden)) {
			return false
		}
	}

	// Advanced quality filters
	if !m.validateQualityFilters(text, context) {
		return false
	}

	return true
}

// validateQualityFilters applies advanced quality validation
func (m *MarkovChainBackend) validateQualityFilters(text string, context DialogContext) bool {
	// Skip if no quality filters configured
	if m.config.QualityFilters.MinCoherence == 0 && m.config.QualityFilters.MaxRepetition == 0 &&
		!m.config.QualityFilters.RequireComplete && !m.config.QualityFilters.GrammarCheck &&
		m.config.QualityFilters.MinUniqueWords == 0 && m.config.QualityFilters.MaxSimilarity == 0 {
		return true
	}

	words := strings.Fields(text)

	// Enhanced coherence check
	if m.config.QualityFilters.MinCoherence > 0 {
		coherence := m.calculateAdvancedCoherence(text)
		if coherence < m.config.QualityFilters.MinCoherence {
			return false
		}
	}

	// Word repetition check
	if m.config.QualityFilters.MaxRepetition > 0 {
		repetition := m.calculateWordRepetition(words)
		if repetition > m.config.QualityFilters.MaxRepetition {
			return false
		}
	}

	// Complete sentence check
	if m.config.QualityFilters.RequireComplete {
		if !m.isCompleteSentence(text) {
			return false
		}
	}

	// Basic grammar check
	if m.config.QualityFilters.GrammarCheck {
		if !m.passesBasicGrammarCheck(text) {
			return false
		}
	}

	// Unique words check
	if m.config.QualityFilters.MinUniqueWords > 0 {
		uniqueWords := m.countUniqueWords(words)
		if uniqueWords < m.config.QualityFilters.MinUniqueWords {
			return false
		}
	}

	// Similarity check against recent responses
	if m.config.QualityFilters.MaxSimilarity > 0 {
		similarity := m.calculateSimilarityToRecent(text, context)
		if similarity > m.config.QualityFilters.MaxSimilarity {
			return false
		}
	}

	return true
}

// calculateAdvancedCoherence provides enhanced coherence analysis
func (m *MarkovChainBackend) calculateAdvancedCoherence(text string) float64 {
	words := strings.Fields(text)
	if len(words) < 2 {
		return 0.5
	}

	coherenceScore := 1.0

	// Check for excessive repetition
	wordCounts := make(map[string]int)
	for _, word := range words {
		wordCounts[strings.ToLower(word)]++
	}

	totalRepetition := 0
	for _, count := range wordCounts {
		if count > 1 {
			totalRepetition += count - 1
		}
	}

	repetitionPenalty := float64(totalRepetition) / float64(len(words))
	coherenceScore -= repetitionPenalty * 0.3

	// Check for proper word ordering (basic heuristic)
	properOrderBonus := m.checkWordOrdering(words)
	coherenceScore += properOrderBonus * 0.2

	// Clamp result
	if coherenceScore < 0 {
		coherenceScore = 0
	}
	if coherenceScore > 1 {
		coherenceScore = 1
	}

	return coherenceScore
}

// calculateWordRepetition calculates the repetition ratio in the text
func (m *MarkovChainBackend) calculateWordRepetition(words []string) float64 {
	if len(words) == 0 {
		return 0
	}

	wordCounts := make(map[string]int)
	for _, word := range words {
		wordCounts[strings.ToLower(word)]++
	}

	totalRepeated := 0
	for _, count := range wordCounts {
		if count > 1 {
			totalRepeated += count - 1
		}
	}

	return float64(totalRepeated) / float64(len(words))
}

// isCompleteSentence checks if the text forms complete sentences
func (m *MarkovChainBackend) isCompleteSentence(text string) bool {
	text = strings.TrimSpace(text)
	if len(text) == 0 {
		return false
	}

	// Check for sentence-ending punctuation
	lastChar := text[len(text)-1]
	return lastChar == '.' || lastChar == '!' || lastChar == '?'
}

// passesBasicGrammarCheck performs basic grammar validation
func (m *MarkovChainBackend) passesBasicGrammarCheck(text string) bool {
	// Very basic grammar checks:
	// 1. Starts with capital letter (unless it's a special case like "*blushes*")
	// 2. Has proper spacing
	// 3. No double punctuation (except "..." or "!!")

	text = strings.TrimSpace(text)
	if len(text) == 0 {
		return false
	}

	// Allow special formatting like "*blushes*" or "(whispers)"
	if strings.HasPrefix(text, "*") || strings.HasPrefix(text, "(") {
		return true
	}

	// Check capitalization
	firstChar := rune(text[0])
	if !('A' <= firstChar && firstChar <= 'Z') {
		return false
	}

	// Check for excessive punctuation
	if strings.Contains(text, "???") || strings.Contains(text, "!!!") ||
		strings.Contains(text, ".,") || strings.Contains(text, ".!") {
		return false
	}

	return true
}

// countUniqueWords counts unique words in the word list
func (m *MarkovChainBackend) countUniqueWords(words []string) int {
	uniqueWords := make(map[string]bool)
	for _, word := range words {
		cleanWord := strings.ToLower(strings.Trim(word, ".,!?"))
		if len(cleanWord) > 0 {
			uniqueWords[cleanWord] = true
		}
	}
	return len(uniqueWords)
}

// checkWordOrdering provides a basic check for reasonable word ordering
func (m *MarkovChainBackend) checkWordOrdering(words []string) float64 {
	if len(words) < 2 {
		return 0.5
	}

	// Very basic heuristic: check if articles/determiners are followed by nouns
	// This is a simplified implementation for demonstration
	score := 0.0
	checks := 0

	for i := 0; i < len(words)-1; i++ {
		word := strings.ToLower(words[i])
		nextWord := strings.ToLower(words[i+1])

		// Check if articles are followed by reasonable words
		if word == "the" || word == "a" || word == "an" {
			checks++
			// Avoid common grammar mistakes
			if nextWord != "the" && nextWord != "a" && nextWord != "an" &&
				nextWord != "." && nextWord != "," && nextWord != "!" {
				score++
			}
		}
	}

	if checks == 0 {
		return 0.5 // Neutral if no applicable checks
	}

	return score / float64(checks)
}

// calculateCoherence estimates text coherence (simplified implementation)
func (m *MarkovChainBackend) calculateCoherence(text string) float64 {
	words := strings.Fields(text)
	if len(words) < 2 {
		return 0.5 // Neutral for very short text
	}

	// Simple coherence based on word repetition and length consistency
	uniqueWords := make(map[string]int)
	for _, word := range words {
		uniqueWords[strings.ToLower(word)]++
	}

	// High repetition = low coherence
	repetitionPenalty := 0.0
	for _, count := range uniqueWords {
		if count > 1 {
			repetitionPenalty += float64(count-1) * 0.1
		}
	}

	coherence := 1.0 - repetitionPenalty
	if coherence < 0 {
		coherence = 0
	}
	if coherence > 1 {
		coherence = 1
	}

	return coherence
}

// selectFallbackResponse provides a fallback when generation fails
func (m *MarkovChainBackend) selectFallbackResponse(context DialogContext) string {
	// Use configured fallback phrases if available
	if len(m.config.FallbackPhrases) > 0 {
		index := int(time.Now().UnixNano()) % len(m.config.FallbackPhrases)
		return m.config.FallbackPhrases[index]
	}

	// Use context fallback responses
	if len(context.FallbackResponses) > 0 {
		index := int(time.Now().UnixNano()) % len(context.FallbackResponses)
		return context.FallbackResponses[index]
	}

	// Hard-coded final fallback
	return "I'm not sure what to say right now..."
}

// selectAnimation chooses an appropriate animation for the response
func (m *MarkovChainBackend) selectAnimation(text string, context DialogContext) string {
	// Simple heuristic-based animation selection
	lowerText := strings.ToLower(text)

	// Emotional content detection
	if strings.Contains(lowerText, "love") || strings.Contains(lowerText, "heart") {
		return "heart_eyes"
	}
	if strings.Contains(lowerText, "thank") || strings.Contains(lowerText, "grateful") {
		return "happy"
	}
	if strings.Contains(lowerText, "shy") || strings.Contains(lowerText, "blush") {
		return "blushing"
	}
	if strings.Contains(lowerText, "sad") || strings.Contains(lowerText, "sorry") {
		return "sad"
	}

	// Trigger-based animation selection
	switch context.Trigger {
	case "compliment":
		return "blushing"
	case "give_gift":
		return "excited_romance"
	case "deep_conversation":
		return "romantic_idle"
	default:
		return "talking"
	}
}

// classifyResponseType categorizes the type of response generated
func (m *MarkovChainBackend) classifyResponseType(text string, context DialogContext) string {
	lowerText := strings.ToLower(text)

	// Romantic content detection
	romanticWords := []string{"love", "heart", "romance", "kiss", "hug", "together", "forever"}
	for _, word := range romanticWords {
		if strings.Contains(lowerText, word) {
			return "romantic"
		}
	}

	// Casual/friendly detection
	casualWords := []string{"hi", "hello", "hey", "thanks", "nice", "good"}
	for _, word := range casualWords {
		if strings.Contains(lowerText, word) {
			return "casual"
		}
	}

	// Emotional detection
	emotionalWords := []string{"feel", "emotion", "happy", "sad", "excited", "nervous"}
	for _, word := range emotionalWords {
		if strings.Contains(lowerText, word) {
			return "emotional"
		}
	}

	return "general"
}

// detectEmotionalTone analyzes the emotional tone of the response
func (m *MarkovChainBackend) detectEmotionalTone(text string, context DialogContext) string {
	lowerText := strings.ToLower(text)

	// Positive emotions
	if strings.Contains(lowerText, "happy") || strings.Contains(lowerText, "joy") ||
		strings.Contains(lowerText, "excited") || strings.Contains(lowerText, "wonderful") {
		return "happy"
	}

	// Shy/romantic emotions
	if strings.Contains(lowerText, "shy") || strings.Contains(lowerText, "blush") ||
		strings.Contains(lowerText, "nervous") {
		return "shy"
	}

	// Loving/romantic emotions
	if strings.Contains(lowerText, "love") || strings.Contains(lowerText, "adore") ||
		strings.Contains(lowerText, "cherish") {
		return "loving"
	}

	// Sad emotions
	if strings.Contains(lowerText, "sad") || strings.Contains(lowerText, "sorry") ||
		strings.Contains(lowerText, "miss") {
		return "sad"
	}

	// Default to neutral
	return "neutral"
}

// extractTopics identifies topics covered in the response
func (m *MarkovChainBackend) extractTopics(text string) []string {
	lowerText := strings.ToLower(text)
	topics := []string{}

	// Topic detection based on keywords
	topicMap := map[string][]string{
		"relationship": {"love", "relationship", "together", "partner", "couple"},
		"feelings":     {"feel", "emotion", "heart", "mood", "sentiment"},
		"gratitude":    {"thank", "grateful", "appreciate", "thankful"},
		"conversation": {"talk", "chat", "discuss", "conversation", "speak"},
		"future":       {"future", "tomorrow", "plan", "hope", "dream"},
		"past":         {"remember", "memory", "past", "before", "ago"},
	}

	for topic, keywords := range topicMap {
		for _, keyword := range keywords {
			if strings.Contains(lowerText, keyword) {
				topics = append(topics, topic)
				break
			}
		}
	}

	return topics
}

// calculateMemoryImportance determines how important this response is for memory storage
func (m *MarkovChainBackend) calculateMemoryImportance(text string, context DialogContext) float64 {
	importance := 0.5 // Base importance

	// Emotional responses are more important
	tone := m.detectEmotionalTone(text, context)
	if tone != "neutral" {
		importance += 0.2
	}

	// Romantic responses are very important
	if m.classifyResponseType(text, context) == "romantic" {
		importance += 0.3
	}

	// Longer responses tend to be more important
	words := strings.Fields(text)
	if len(words) > m.config.MaxWords/2 {
		importance += 0.1
	}

	// Clamp to valid range
	if importance > 1.0 {
		importance = 1.0
	}

	return importance
}

// GetBackendInfo returns metadata about the Markov chain backend
func (m *MarkovChainBackend) GetBackendInfo() BackendInfo {
	return BackendInfo{
		Name:        "markov_chain",
		Version:     "1.0.0",
		Description: "Markov chain text generation with personality and context awareness",
		Capabilities: []string{
			"text_generation",
			"personality_adaptation",
			"context_awareness",
			"trigger_specific_chains",
			"memory_learning",
			"coherence_filtering",
		},
		Author:  "DDS Development Team",
		License: "MIT",
	}
}

// CanHandle checks if this backend can process the given context
func (m *MarkovChainBackend) CanHandle(context DialogContext) bool {
	if !m.initialized {
		return false
	}

	// Check if we have appropriate training data
	chain := m.selectChain(context.Trigger)
	return chain != nil && chain.hasEnoughData()
}

// UpdateMemory records interaction outcomes for learning and adaptation
func (m *MarkovChainBackend) UpdateMemory(context DialogContext, response DialogResponse, feedback *UserFeedback) error {
	if m.config.LearningRate <= 0 || feedback == nil {
		return nil
	}

	// For positive feedback, reinforce the response patterns
	if feedback.Positive && feedback.Engagement > 0.7 {
		// Add the successful response to training data
		return m.trainWithText(response.Text, context.Trigger)
	}

	// For negative feedback, we could implement negative reinforcement
	// For now, we just record the feedback (no action taken)
	return nil
}

// NewMarkovChain creates a new Markov chain with the specified order
func NewMarkovChain(order int) *MarkovChain {
	return &MarkovChain{
		order:       order,
		states:      make(map[string][]string),
		starters:    []string{},
		wordCounts:  make(map[string]int),
		totalWords:  0,
		lastUpdated: time.Now(),
	}
}

// Train adds text to the Markov chain training data
func (c *MarkovChain) Train(text string) {
	words := strings.Fields(text)
	if len(words) < c.order+1 {
		return // Not enough words for this order
	}

	// Record word frequencies
	for _, word := range words {
		c.wordCounts[word]++
		c.totalWords++
	}

	// Build states and transitions
	for i := 0; i <= len(words)-c.order-1; i++ {
		state := c.createState(words[i : i+c.order])
		nextWord := words[i+c.order]

		c.states[state.key] = append(c.states[state.key], nextWord)

		// Record sentence starters
		if i == 0 {
			c.starters = append(c.starters, state.key)
		}
	}

	c.lastUpdated = time.Now()
}

// Generate creates new text using the Markov chain
func (c *MarkovChain) Generate(targetWords int, temperature float64) (string, float64) {
	if len(c.starters) == 0 {
		return "", 0.0
	}

	// Select starting state
	startKey := c.selectRandomStarter(temperature)
	currentState := strings.Fields(startKey)
	result := make([]string, len(currentState))
	copy(result, currentState)

	confidence := 0.8 // Start with high confidence

	// Generate words until target length
	for len(result) < targetWords {
		stateKey := strings.Join(currentState, " ")
		nextWords := c.states[stateKey]

		if len(nextWords) == 0 {
			// No transitions available, try to find a new starting point
			if len(result) >= targetWords/2 {
				break // We have enough content
			}

			// Restart with new state
			startKey = c.selectRandomStarter(temperature)
			currentState = strings.Fields(startKey)
			confidence *= 0.9 // Reduce confidence for restarts
			continue
		}

		// Select next word based on temperature
		nextWord := c.selectNextWord(nextWords, temperature)
		result = append(result, nextWord)

		// Update current state
		currentState = append(currentState[1:], nextWord)
	}

	// Join result and calculate final confidence
	text := strings.Join(result, " ")
	finalConfidence := confidence * c.calculateGenerationConfidence(text)

	return text, finalConfidence
}

// createState creates a MarkovState from a slice of words
func (c *MarkovChain) createState(words []string) MarkovState {
	return MarkovState{
		words: words,
		key:   strings.Join(words, " "),
	}
}

// selectRandomStarter chooses a random starting state based on temperature
func (c *MarkovChain) selectRandomStarter(temperature float64) string {
	if temperature <= 0.1 {
		// Low temperature: prefer most common starters
		starterCounts := make(map[string]int)
		for _, starter := range c.starters {
			starterCounts[starter]++
		}

		maxCount := 0
		bestStarter := c.starters[0]
		for starter, count := range starterCounts {
			if count > maxCount {
				maxCount = count
				bestStarter = starter
			}
		}
		return bestStarter
	}

	// Higher temperature: more random selection
	index := rand.Intn(len(c.starters))
	return c.starters[index]
}

// selectNextWord chooses the next word based on available transitions and temperature
func (c *MarkovChain) selectNextWord(options []string, temperature float64) string {
	if len(options) == 1 {
		return options[0]
	}

	if temperature <= 0.1 {
		// Low temperature: prefer most frequent words
		wordCounts := make(map[string]int)
		for _, word := range options {
			wordCounts[word]++
		}

		maxCount := 0
		bestWord := options[0]
		for word, count := range wordCounts {
			if count > maxCount {
				maxCount = count
				bestWord = word
			}
		}
		return bestWord
	}

	// Higher temperature: more random selection
	index := rand.Intn(len(options))
	return options[index]
}

// calculateGenerationConfidence estimates confidence in the generated text
func (c *MarkovChain) calculateGenerationConfidence(text string) float64 {
	words := strings.Fields(text)
	if len(words) == 0 {
		return 0.0
	}

	// Base confidence on word frequency in training data
	totalConfidence := 0.0
	for _, word := range words {
		frequency := float64(c.wordCounts[word]) / float64(c.totalWords)
		// Convert frequency to confidence (common words = higher confidence)
		wordConfidence := frequency * 10 // Scale factor
		if wordConfidence > 1.0 {
			wordConfidence = 1.0
		}
		totalConfidence += wordConfidence
	}

	return totalConfidence / float64(len(words))
}

// hasEnoughData checks if the chain has sufficient training data for generation
func (c *MarkovChain) hasEnoughData() bool {
	return len(c.starters) >= 2 && len(c.states) >= 10
}
