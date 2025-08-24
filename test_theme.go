package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

func main() {
	// Test what theme constants are available
	fmt.Println("Testing theme constants")

	// Try some likely candidates for color names
	var colorName fyne.ThemeColorName
	colorName = "foreground"
	fmt.Println("Color name:", colorName)

	var sizeName fyne.ThemeSizeName
	sizeName = "text"
	fmt.Println("Size name:", sizeName)

	// Check what functions return colors
	fmt.Println("ForegroundColor:", theme.ForegroundColor())
}
