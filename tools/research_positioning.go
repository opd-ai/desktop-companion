// This file is for researching Fyne window positioning capabilities
package main

import (
	"fmt"
	"reflect"

	"fyne.io/fyne/v2/app"
)

func researchFynePositioning() {
	// Create a test app and window to inspect available methods
	testApp := app.New()
	testWindow := testApp.NewWindow("Research")
	
	// Get the type information for the window
	windowType := reflect.TypeOf(testWindow)
	fmt.Printf("Window type: %v\n", windowType)
	
	// List all methods available on the window
	fmt.Println("\nAvailable methods:")
	for i := 0; i < windowType.NumMethod(); i++ {
		method := windowType.Method(i)
		fmt.Printf("- %s\n", method.Name)
	}
	
	// Check if window implements any desktop-specific interfaces
	fmt.Println("\nChecking for desktop-specific interfaces...")
	
	// Try to access the underlying driver
	canvas := testWindow.Canvas()
	canvasType := reflect.TypeOf(canvas)
	fmt.Printf("Canvas type: %v\n", canvasType)
	
	// Check canvas methods for positioning
	fmt.Println("\nCanvas methods:")
	for i := 0; i < canvasType.NumMethod(); i++ {
		method := canvasType.Method(i)
		fmt.Printf("- %s\n", method.Name)
	}
}
