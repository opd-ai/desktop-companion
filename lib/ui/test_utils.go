package ui

import (
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

// Global mutex to protect Fyne test app creation from race conditions
var fyneTestMutex sync.Mutex

// SafeNewTestApp creates a new Fyne test app in a thread-safe manner
// This prevents race conditions in Fyne's internal font cache
func SafeNewTestApp() fyne.App {
	fyneTestMutex.Lock()
	defer fyneTestMutex.Unlock()
	return test.NewApp()
}
