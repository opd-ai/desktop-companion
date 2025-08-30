package main

import (
	"desktop-companion/internal/character"
	"fmt"
	"log"
)

func main() {
	// Test loading our example character card
	loader := character.NewPlatformAwareLoader()
	card, err := loader.LoadCharacterCard("assets/characters/examples/cross_platform_character.json")
	if err != nil {
		log.Fatalf("Error loading character: %v", err)
	}

	fmt.Printf("✅ Successfully loaded cross-platform character: %s\n", card.Name)
	fmt.Printf("📝 Description: %s\n", card.Description)
	fmt.Printf("📱 Has platform config: %t\n", card.PlatformConfig != nil)

	if card.PlatformConfig != nil {
		fmt.Printf("🖥️  Desktop config: %t\n", card.PlatformConfig.Desktop != nil)
		fmt.Printf("📱 Mobile config: %t\n", card.PlatformConfig.Mobile != nil)

		// Test desktop platform config
		if card.PlatformConfig.Desktop != nil {
			desktop := card.PlatformConfig.Desktop
			fmt.Printf("🖥️  Desktop window mode: %s\n", desktop.WindowMode)
			if desktop.Behavior != nil {
				fmt.Printf("🖥️  Desktop size: %d\n", desktop.Behavior.DefaultSize)
				fmt.Printf("🖥️  Desktop movement: %t\n", desktop.Behavior.MovementEnabled)
			}
		}

		// Test mobile platform config
		if card.PlatformConfig.Mobile != nil {
			mobile := card.PlatformConfig.Mobile
			fmt.Printf("📱 Mobile window mode: %s\n", mobile.WindowMode)
			fmt.Printf("📱 Touch optimized: %t\n", mobile.TouchOptimized)
			if mobile.Behavior != nil {
				fmt.Printf("📱 Mobile size: %d\n", mobile.Behavior.DefaultSize)
				fmt.Printf("📱 Mobile movement: %t\n", mobile.Behavior.MovementEnabled)
			}
			if mobile.MobileControls != nil {
				fmt.Printf("📱 Bottom bar: %t\n", mobile.MobileControls.ShowBottomBar)
				fmt.Printf("📱 Haptic feedback: %t\n", mobile.MobileControls.HapticFeedback)
			}
		}
	}

	// Test validation
	err = character.ValidatePlatformConfig(card)
	if err != nil {
		log.Fatalf("Validation failed: %v", err)
	}
	fmt.Printf("✅ Platform configuration validation passed\n")

	// Test current platform config retrieval
	currentConfig := loader.GetPlatformConfig(card)
	if currentConfig != nil {
		fmt.Printf("🔧 Current platform config loaded successfully\n")
	}

	fmt.Printf("🎉 All validation tests passed!\n")
}
