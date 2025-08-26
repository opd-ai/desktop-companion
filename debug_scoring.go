package main

import (
	"fmt"
	"strings"
)

func calculateDialogScore(response string, shyness, romanticism, flirtiness, affection float64) float64 {
	baseScore := 1.0

	// Romantic content preference
	if romanticism > 0.6 && (len(response) > 30 || strings.Contains(response, "💕") || strings.Contains(response, "💖")) {
		bonus := romanticism * 0.5
		baseScore += bonus
		fmt.Printf("Romanticism bonus for '%s': +%.1f\n", response, bonus)
	}

	// Shy characters prefer shorter, less dramatic responses
	if shyness > 0.6 {
		if len(response) < 25 {
			baseScore += shyness
			fmt.Printf("Shyness bonus for '%s': +%.1f\n", response, shyness)
		} else {
			penalty := shyness * 0.5
			baseScore -= penalty
			fmt.Printf("Shyness penalty for '%s': -%.1f\n", response, penalty)
		}
	}

	// Flirty characters prefer bold responses, shy characters avoid them
	if strings.Contains(response, "*boldly*") || strings.Contains(response, "😘") {
		if flirtiness > 0.6 {
			baseScore += flirtiness
			fmt.Printf("Flirtiness bonus for '%s': +%.1f\n", response, flirtiness)
		} else if shyness > 0.6 {
			penalty := shyness * 0.5
			baseScore -= penalty
			fmt.Printf("Shyness boldness penalty for '%s': -%.1f\n", response, penalty)
		}
	}

	// Adjust based on current affection level
	if affection > 50 && romanticism > 0.5 {
		baseScore += 0.5
		fmt.Printf("High affection romantic bonus for '%s': +0.5\n", response)
	}

	fmt.Printf("Final score for '%s' (len=%d): %.1f\n\n", response, len(response), baseScore)
	return baseScore
}

func main() {
	// Test character personality
	shyness := 0.8
	romanticism := 0.9
	flirtiness := 0.2
	affection := 60.0

	fmt.Printf("Character: shyness=%.1f, romanticism=%.1f, flirtiness=%.1f, affection=%.1f\n\n",
		shyness, romanticism, flirtiness, affection)

	// Test responses
	boldResponse := "*boldly* Hey there, gorgeous! 😘"
	shyResponse := "Hi... *blushes softly*"

	fmt.Println("Bold response:")
	boldScore := calculateDialogScore(boldResponse, shyness, romanticism, flirtiness, affection)

	fmt.Println("Shy response:")
	shyScore := calculateDialogScore(shyResponse, shyness, romanticism, flirtiness, affection)

	fmt.Printf("Winner: %s (%.1f vs %.1f)\n",
		func() string {
			if shyScore > boldScore {
				return "shy"
			}
			return "bold"
		}(), shyScore, boldScore)
}
