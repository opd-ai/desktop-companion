package dialog

import (
	"fmt"
	"strings"
)

// PersonalityExtractor provides utilities to extract personality information
// from various sources (Markov training data, character traits, dialog history)
// and convert them to LLM-compatible prompts
type PersonalityExtractor struct {
	characterName string
	description   string
}

// NewPersonalityExtractor creates a new personality extraction service
func NewPersonalityExtractor(characterName, description string) *PersonalityExtractor {
	return &PersonalityExtractor{
		characterName: characterName,
		description:   description,
	}
}

// ExtractFromTrainingData analyzes Markov training data to extract personality indicators
// This provides a bridge between existing Markov configurations and LLM prompts
func (pe *PersonalityExtractor) ExtractFromTrainingData(trainingData []string) PersonalityPrompt {
	if len(trainingData) == 0 {
		return PersonalityPrompt{
			SystemPrompt:     pe.buildDefaultSystemPrompt(),
			PersonalityHints: "",
			SpeechPatterns:   []string{},
			EmotionalTone:    "neutral",
			ResponseStyle:    "casual",
		}
	}

	// Analyze training data for patterns
	speechPatterns := pe.analyzeSpeechPatterns(trainingData)
	emotionalTone := pe.analyzeEmotionalTone(trainingData)
	responseStyle := pe.analyzeResponseStyle(trainingData)
	personalityHints := pe.extractPersonalityFromText(trainingData)

	return PersonalityPrompt{
		SystemPrompt:     pe.buildSystemPromptFromData(trainingData),
		PersonalityHints: personalityHints,
		SpeechPatterns:   speechPatterns,
		EmotionalTone:    emotionalTone,
		ResponseStyle:    responseStyle,
	}
}

// ExtractFromTraits converts character trait scores to LLM personality prompts
func (pe *PersonalityExtractor) ExtractFromTraits(traits map[string]float64) PersonalityPrompt {
	if len(traits) == 0 {
		return PersonalityPrompt{
			SystemPrompt:     pe.buildDefaultSystemPrompt(),
			PersonalityHints: "",
			SpeechPatterns:   []string{},
			EmotionalTone:    "neutral",
			ResponseStyle:    "casual",
		}
	}

	personalityHints := pe.buildTraitDescription(traits)
	emotionalTone := pe.inferEmotionalToneFromTraits(traits)
	responseStyle := pe.inferResponseStyleFromTraits(traits)
	speechPatterns := pe.inferSpeechPatternsFromTraits(traits)

	systemPrompt := pe.buildSystemPromptFromTraits(traits)

	return PersonalityPrompt{
		SystemPrompt:     systemPrompt,
		PersonalityHints: personalityHints,
		SpeechPatterns:   speechPatterns,
		EmotionalTone:    emotionalTone,
		ResponseStyle:    responseStyle,
	}
}

// CombinePrompts merges multiple personality prompts into a unified prompt
func (pe *PersonalityExtractor) CombinePrompts(prompts ...PersonalityPrompt) PersonalityPrompt {
	if len(prompts) == 0 {
		return PersonalityPrompt{}
	}

	if len(prompts) == 1 {
		return prompts[0]
	}

	// Combine system prompts
	systemPrompt := prompts[0].SystemPrompt
	for i := 1; i < len(prompts); i++ {
		if prompts[i].SystemPrompt != "" && prompts[i].SystemPrompt != systemPrompt {
			systemPrompt += " " + prompts[i].SystemPrompt
		}
	}

	// Combine personality hints
	var allHints []string
	for _, prompt := range prompts {
		if prompt.PersonalityHints != "" {
			allHints = append(allHints, prompt.PersonalityHints)
		}
	}

	// Combine speech patterns (deduplicate)
	speechPatterns := make(map[string]bool)
	for _, prompt := range prompts {
		for _, pattern := range prompt.SpeechPatterns {
			speechPatterns[pattern] = true
		}
	}
	var uniquePatterns []string
	for pattern := range speechPatterns {
		uniquePatterns = append(uniquePatterns, pattern)
	}

	// Use first non-empty emotional tone and response style
	emotionalTone := "neutral"
	responseStyle := "casual"
	for _, prompt := range prompts {
		if prompt.EmotionalTone != "" && emotionalTone == "neutral" {
			emotionalTone = prompt.EmotionalTone
		}
		if prompt.ResponseStyle != "" && responseStyle == "casual" {
			responseStyle = prompt.ResponseStyle
		}
	}

	return PersonalityPrompt{
		SystemPrompt:     systemPrompt,
		PersonalityHints: strings.Join(allHints, " "),
		SpeechPatterns:   uniquePatterns,
		EmotionalTone:    emotionalTone,
		ResponseStyle:    responseStyle,
	}
}

// PersonalityPrompt contains extracted personality information for LLM prompting
type PersonalityPrompt struct {
	SystemPrompt     string   `json:"systemPrompt"`     // Base system prompt
	PersonalityHints string   `json:"personalityHints"` // Personality description for LLM
	SpeechPatterns   []string `json:"speechPatterns"`   // Common phrases or speech patterns
	EmotionalTone    string   `json:"emotionalTone"`    // Dominant emotional tone
	ResponseStyle    string   `json:"responseStyle"`    // Communication style
}

// ToLLMPrompt converts the personality prompt to a complete LLM system prompt
func (pp PersonalityPrompt) ToLLMPrompt() string {
	prompt := pp.SystemPrompt

	if pp.PersonalityHints != "" {
		prompt += "\n\nPersonality: " + pp.PersonalityHints
	}

	if pp.EmotionalTone != "" && pp.EmotionalTone != "neutral" {
		prompt += "\n\nEmotional tone: " + pp.EmotionalTone
	}

	if pp.ResponseStyle != "" && pp.ResponseStyle != "casual" {
		prompt += "\n\nResponse style: " + pp.ResponseStyle
	}

	if len(pp.SpeechPatterns) > 0 {
		prompt += "\n\nSpeech patterns to consider: " + strings.Join(pp.SpeechPatterns, ", ")
	}

	prompt += "\n\nRespond briefly and naturally, staying in character."

	return prompt
}

// Internal analysis methods

func (pe *PersonalityExtractor) buildDefaultSystemPrompt() string {
	if pe.characterName != "" {
		return fmt.Sprintf("You are %s, a virtual companion.", pe.characterName)
	}
	return "You are a friendly virtual companion."
}

func (pe *PersonalityExtractor) buildSystemPromptFromData(trainingData []string) string {
	basePrompt := pe.buildDefaultSystemPrompt()

	if pe.description != "" {
		basePrompt += " " + pe.description
	}

	// Add context from training data analysis
	if len(trainingData) > 0 {
		basePrompt += " Based on your previous conversations, maintain your established personality and speech patterns."
	}

	return basePrompt
}

func (pe *PersonalityExtractor) buildSystemPromptFromTraits(traits map[string]float64) string {
	basePrompt := pe.buildDefaultSystemPrompt()

	if pe.description != "" {
		basePrompt += " " + pe.description
	}

	// Add dominant traits to system prompt
	dominantTraits := pe.getDominantTraits(traits, 0.7)
	if len(dominantTraits) > 0 {
		basePrompt += fmt.Sprintf(" You have these notable traits: %s.", strings.Join(dominantTraits, ", "))
	}

	return basePrompt
}

func (pe *PersonalityExtractor) analyzeSpeechPatterns(trainingData []string) []string {
	patterns := make(map[string]int)

	// Look for common phrases or patterns
	for _, text := range trainingData {
		// Extract short phrases (2-4 words)
		words := strings.Fields(text)
		for i := 0; i < len(words)-1; i++ {
			if i+2 < len(words) {
				phrase := strings.Join(words[i:i+2], " ")
				if len(phrase) > 3 && !isCommonPhrase(phrase) {
					patterns[phrase]++
				}
			}
		}
	}

	// Return most common patterns
	var result []string
	for phrase, count := range patterns {
		if count >= 2 { // Appears at least twice
			result = append(result, phrase)
		}
		if len(result) >= 5 { // Limit to 5 patterns
			break
		}
	}

	return result
}

func (pe *PersonalityExtractor) analyzeEmotionalTone(trainingData []string) string {
	positiveWords := []string{"happy", "joy", "love", "excited", "wonderful", "amazing", "great", "awesome", "💕", "😊", "😄", "💖"}
	negativeWords := []string{"sad", "angry", "upset", "disappointed", "terrible", "awful", "hate", "😢", "😠", "😞"}
	shyWords := []string{"maybe", "perhaps", "I guess", "kinda", "sorta", "um", "uh", "shy", "nervous"}
	flirtyWords := []string{"cute", "sweet", "darling", "honey", "wink", "tease", "😉", "😘", "💋"}

	positiveCount := countWordOccurrences(trainingData, positiveWords)
	negativeCount := countWordOccurrences(trainingData, negativeWords)
	shyCount := countWordOccurrences(trainingData, shyWords)
	flirtyCount := countWordOccurrences(trainingData, flirtyWords)

	// Determine dominant emotional tone
	maxCount := positiveCount
	tone := "happy"

	if negativeCount > maxCount {
		maxCount = negativeCount
		tone = "melancholic"
	}
	if shyCount > maxCount {
		maxCount = shyCount
		tone = "shy"
	}
	if flirtyCount > maxCount {
		tone = "flirty"
	}

	// Default to neutral if no clear pattern
	if maxCount == 0 {
		tone = "neutral"
	}

	return tone
}

func (pe *PersonalityExtractor) analyzeResponseStyle(trainingData []string) string {
	formalWords := []string{"please", "thank you", "certainly", "indeed", "quite", "rather"}
	casualWords := []string{"yeah", "yep", "nah", "hey", "okay", "cool", "awesome"}
	cutWords := []string{"...", "~", "nya", "uwu", "owo", ">.<", "^^"}

	formalCount := countWordOccurrences(trainingData, formalWords)
	casualCount := countWordOccurrences(trainingData, casualWords)
	cuteCount := countWordOccurrences(trainingData, cutWords)

	if cuteCount > casualCount && cuteCount > formalCount {
		return "cute"
	}
	if formalCount > casualCount {
		return "formal"
	}
	if casualCount > 0 {
		return "casual"
	}

	return "casual" // Default
}

func (pe *PersonalityExtractor) extractPersonalityFromText(trainingData []string) string {
	// Analyze text for personality indicators
	combinedText := strings.ToLower(strings.Join(trainingData, " "))

	var traits []string

	if strings.Contains(combinedText, "shy") || strings.Contains(combinedText, "nervous") {
		traits = append(traits, "shy")
	}
	if strings.Contains(combinedText, "confident") || strings.Contains(combinedText, "sure") {
		traits = append(traits, "confident")
	}
	if strings.Contains(combinedText, "playful") || strings.Contains(combinedText, "fun") {
		traits = append(traits, "playful")
	}
	if strings.Contains(combinedText, "serious") || strings.Contains(combinedText, "important") {
		traits = append(traits, "serious")
	}

	if len(traits) > 0 {
		return fmt.Sprintf("Character exhibits these traits: %s", strings.Join(traits, ", "))
	}

	return ""
}

func (pe *PersonalityExtractor) buildTraitDescription(traits map[string]float64) string {
	dominantTraits := pe.getDominantTraits(traits, 0.6)

	if len(dominantTraits) == 0 {
		return ""
	}

	return fmt.Sprintf("Strong personality traits: %s", strings.Join(dominantTraits, ", "))
}

func (pe *PersonalityExtractor) getDominantTraits(traits map[string]float64, threshold float64) []string {
	var dominant []string
	for trait, value := range traits {
		if value >= threshold {
			traitName := strings.ReplaceAll(trait, "_", " ")
			dominant = append(dominant, traitName)
		}
	}
	return dominant
}

func (pe *PersonalityExtractor) inferEmotionalToneFromTraits(traits map[string]float64) string {
	if traits["happiness"] > 0.7 || traits["joy"] > 0.7 {
		return "happy"
	}
	if traits["shyness"] > 0.7 {
		return "shy"
	}
	if traits["flirtiness"] > 0.7 || traits["romanticism"] > 0.7 {
		return "flirty"
	}
	if traits["seriousness"] > 0.7 {
		return "serious"
	}

	return "neutral"
}

func (pe *PersonalityExtractor) inferResponseStyleFromTraits(traits map[string]float64) string {
	if traits["formality"] > 0.7 {
		return "formal"
	}
	if traits["cuteness"] > 0.7 || traits["playfulness"] > 0.7 {
		return "cute"
	}

	return "casual"
}

func (pe *PersonalityExtractor) inferSpeechPatternsFromTraits(traits map[string]float64) []string {
	var patterns []string

	if traits["shyness"] > 0.7 {
		patterns = append(patterns, "um", "maybe", "I think")
	}
	if traits["flirtiness"] > 0.7 {
		patterns = append(patterns, "😉", "cute", "sweetie")
	}
	if traits["tsundere"] > 0.7 {
		patterns = append(patterns, "It's not like", "Don't get the wrong idea", "baka")
	}
	if traits["enthusiasm"] > 0.7 {
		patterns = append(patterns, "awesome", "amazing", "so cool")
	}

	return patterns
}

// Helper functions

func countWordOccurrences(texts []string, words []string) int {
	count := 0
	combinedText := strings.ToLower(strings.Join(texts, " "))

	for _, word := range words {
		count += strings.Count(combinedText, strings.ToLower(word))
	}

	return count
}

func isCommonPhrase(phrase string) bool {
	commonPhrases := []string{"I am", "you are", "this is", "that is", "and the", "of the", "to the", "in the"}
	lowerPhrase := strings.ToLower(phrase)

	for _, common := range commonPhrases {
		if lowerPhrase == common {
			return true
		}
	}

	return false
}
