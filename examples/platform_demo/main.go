// Package main demonstrates the platform detection system.
//
// This example shows how to use the platform package to detect
// the current operating system and adapt behavior accordingly.
//
// Usage:
//
//	go run examples/platform_demo.go
package main

import (
	"fmt"

	"desktop-companion/internal/platform"
)

func main() {
	fmt.Println("Desktop Dating Simulator - Platform Detection Demo")
	fmt.Println("=================================================")

	// Get platform information
	info := platform.GetPlatformInfo()
	fmt.Printf("\n%s\n", info.String())

	// Demonstrate platform-specific behavior
	fmt.Println("\nPlatform Capabilities:")
	fmt.Printf("- Desktop Platform: %t\n", info.IsDesktop())
	fmt.Printf("- Mobile Platform: %t\n", info.IsMobile())
	fmt.Printf("- Tablet Platform: %t\n", info.IsTablet())
	fmt.Printf("- Touch Support: %t\n", info.HasTouch())
	fmt.Printf("- Mouse Support: %t\n", info.HasMouse())
	fmt.Printf("- Keyboard Support: %t\n", info.HasKeyboard())

	// Example adaptive behavior
	fmt.Println("\nRecommended Configuration:")
	if info.IsDesktop() {
		fmt.Println("- Enable window dragging")
		fmt.Println("- Use overlay mode")
		fmt.Println("- Enable keyboard shortcuts")
		fmt.Println("- Default character size: 128px")
	} else if info.IsMobile() {
		fmt.Println("- Disable window dragging")
		fmt.Println("- Use fullscreen mode")
		fmt.Println("- Show touch controls")
		fmt.Println("- Larger character size: 256px")
	}

	fmt.Println("\nInput Method Adaptations:")
	if info.HasTouch() {
		fmt.Println("- Single tap → click interaction")
		fmt.Println("- Long press → context menu")
		fmt.Println("- Double tap → play interaction")
		fmt.Println("- Enable haptic feedback")
	}
	if info.HasMouse() {
		fmt.Println("- Left click → pet interaction")
		fmt.Println("- Right click → context menu")
		fmt.Println("- Double click → play interaction")
		fmt.Println("- Hover → proximity dialog")
	}
	if info.HasKeyboard() {
		fmt.Println("- S key → toggle stats")
		fmt.Println("- C key → chat interface")
		fmt.Println("- N key → network overlay")
		fmt.Println("- ESC key → close dialogs")
	}
}
