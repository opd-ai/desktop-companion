// Package main demonstrates the platform-aware character behavior system
// This example shows how characters adapt their behavior based on the detected platform.
//
// Usage:
//
//	go run examples/platform_demo/main.go
package main

import (
	"fmt"
	"log"

	"github.com/opd-ai/desktop-companion/internal/character"
	"github.com/opd-ai/desktop-companion/internal/platform"
)

func main() {
	fmt.Println("=== Platform-Aware Character Behavior Demo ===")

	// Demonstrate platform detection
	platformInfo := platform.GetPlatformInfo()
	fmt.Printf("Detected Platform: %s (%s)\n", platformInfo.OS, platformInfo.FormFactor)
	fmt.Printf("Input Methods: %v\n\n", platformInfo.InputMethods)

	// Demonstrate platform-aware character creation
	fmt.Println("=== Desktop Character Behavior ===")
	desktopPlatform := &platform.PlatformInfo{
		OS:           "windows",
		FormFactor:   "desktop",
		InputMethods: []string{"mouse", "keyboard"},
	}

	// Note: In a real application, you would create the character with actual animation files
	// For this demo, we'll just create the platform adapter directly
	desktopAdapter := character.NewPlatformBehaviorAdapter(desktopPlatform)
	desktopConfig := desktopAdapter.GetBehaviorConfig()

	fmt.Printf("Movement Enabled: %v\n", desktopConfig.MovementEnabled)
	fmt.Printf("Animation Frame Rate: %d FPS\n", desktopConfig.AnimationFrameRate)
	fmt.Printf("Idle Timeout: %v\n", desktopConfig.IdleTimeout)
	fmt.Printf("Haptic Feedback: %v\n", desktopConfig.HapticFeedback)
	fmt.Printf("Battery Optimization: %v\n", desktopConfig.BatteryOptimization)

	fmt.Printf("Optimal Character Size (1920px width): %d pixels\n",
		desktopAdapter.GetOptimalCharacterSize(1920, 128))

	fmt.Println()

	// Demonstrate mobile behavior
	fmt.Println("=== Mobile Character Behavior ===")
	mobilePlatform := &platform.PlatformInfo{
		OS:           "android",
		FormFactor:   "mobile",
		InputMethods: []string{"touch"},
	}

	mobileAdapter := character.NewPlatformBehaviorAdapter(mobilePlatform)
	mobileConfig := mobileAdapter.GetBehaviorConfig()

	fmt.Printf("Movement Enabled: %v\n", mobileConfig.MovementEnabled)
	fmt.Printf("Animation Frame Rate: %d FPS\n", mobileConfig.AnimationFrameRate)
	fmt.Printf("Idle Timeout: %v\n", mobileConfig.IdleTimeout)
	fmt.Printf("Haptic Feedback: %v\n", mobileConfig.HapticFeedback)
	fmt.Printf("Battery Optimization: %v\n", mobileConfig.BatteryOptimization)

	fmt.Printf("Optimal Character Size (400px width): %d pixels\n",
		mobileAdapter.GetOptimalCharacterSize(400, 128))

	fmt.Println()

	// Demonstrate feature detection
	fmt.Println("=== Platform Feature Comparison ===")
	features := []string{"movement", "haptic", "audio", "battery_optimization", "memory_optimization"}

	fmt.Printf("%-20s %-10s %-10s\n", "Feature", "Desktop", "Mobile")
	fmt.Println("----------------------------------------")
	for _, feature := range features {
		desktopEnabled := desktopAdapter.ShouldEnableFeature(feature)
		mobileEnabled := mobileAdapter.ShouldEnableFeature(feature)
		fmt.Printf("%-20s %-10v %-10v\n", feature, desktopEnabled, mobileEnabled)
	}

	fmt.Println()

	// Demonstrate interaction delays
	fmt.Println("=== Interaction Delay Comparison ===")
	interactions := []string{"click", "rightclick", "doubleclick"}

	fmt.Printf("%-15s %-15s %-15s\n", "Interaction", "Desktop", "Mobile")
	fmt.Println("-----------------------------------------------")
	for _, interaction := range interactions {
		desktopDelay := desktopAdapter.GetInteractionDelayForEvent(interaction)
		mobileDelay := mobileAdapter.GetInteractionDelayForEvent(interaction)
		fmt.Printf("%-15s %-15v %-15v\n", interaction, desktopDelay, mobileDelay)
	}

	fmt.Println()
	fmt.Println("=== Integration with Character Card ===")

	// Show how this would work with actual character creation (if we had animation files)
	fmt.Println("When creating a character with NewWithPlatform():")
	fmt.Println("- Desktop: Movement enabled, 60 FPS, 128px size")
	fmt.Println("- Mobile: Movement disabled, 30 FPS, 100px size (25% of 400px screen)")
	fmt.Println("- Interaction cooldowns automatically adapted")
	fmt.Println("- Animation frame rates optimized for platform")
	fmt.Println("- Memory and battery optimizations applied on mobile")

	fmt.Println("\n=== Backward Compatibility ===")
	fmt.Println("- Existing character cards work unchanged")
	fmt.Println("- Old constructor New() defaults to desktop behavior")
	fmt.Println("- Nil platform info gracefully handled")
	fmt.Println("- No breaking changes to existing API")

	log.Println("Platform-aware character behavior demo completed successfully!")
}
