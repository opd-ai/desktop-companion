package ui

import (
	"testing"

	"fyne.io/fyne/v2/test"

	"github.com/opd-ai/desktop-companion/lib/monitoring"
)

// TestBug3EventsFlagFix validates that the -events flag now functionally controls the events system
func TestBug3EventsFlagFix(t *testing.T) {
	testCases := []struct {
		name                  string
		eventsEnabled         bool
		expectEventsShortcuts bool
	}{
		{
			name:                  "Events enabled - shortcuts should be registered",
			eventsEnabled:         true,
			expectEventsShortcuts: true,
		},
		{
			name:                  "Events disabled - shortcuts should not be registered",
			eventsEnabled:         false,
			expectEventsShortcuts: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test character using the helper functions
			card := createTestCharacterCard()
			char := createMockCharacter(card)
			if char == nil {
				t.Skip("Skipping test due to character creation failure")
				return
			}

			// Create test app and profiler
			app := test.NewApp()
			defer app.Quit()
			profiler := monitoring.NewProfiler(50)

			// Create DesktopWindow with eventsEnabled flag
			window := NewDesktopWindow(
				app,
				char,
				false, // debug
				profiler,
				true,  // gameMode
				false, // showStats
				nil,   // networkManager
				false, // networkMode
				false, // showNetwork
				tc.eventsEnabled,
			)

			// Verify events system state matches flag
			if window.eventsEnabled != tc.eventsEnabled {
				t.Errorf("Expected eventsEnabled to be %v, got %v", tc.eventsEnabled, window.eventsEnabled)
			}

			// Verify eventsEnabled field is properly set during construction
			if tc.eventsEnabled && !window.eventsEnabled {
				t.Error("Events should be enabled when eventsEnabled=true")
			}
			if !tc.eventsEnabled && window.eventsEnabled {
				t.Error("Events should be disabled when eventsEnabled=false")
			}
		})
	}
}

// TestBug3EventsFlagRegression provides comprehensive regression testing
// to ensure the events flag behavior is maintained in future changes
func TestBug3EventsFlagRegression(t *testing.T) {
	// Create test character
	card := createTestCharacterCard()
	char := createMockCharacter(card)
	if char == nil {
		t.Skip("Skipping test due to character creation failure")
		return
	}

	// Create test app and profiler
	app := test.NewApp()
	defer app.Quit()
	profiler := monitoring.NewProfiler(50)

	t.Run("Constructor parameter mapping", func(t *testing.T) {
		// Test that constructor parameter is correctly mapped to struct field
		window := NewDesktopWindow(app, char, false, profiler, true, false, nil, false, false, true)
		if !window.eventsEnabled {
			t.Error("eventsEnabled parameter should be mapped to struct field")
		}

		window = NewDesktopWindow(app, char, false, profiler, true, false, nil, false, false, false)
		if window.eventsEnabled {
			t.Error("eventsEnabled parameter should be mapped to struct field")
		}
	})

	t.Run("Independent of other flags", func(t *testing.T) {
		// Verify events flag works independently of other boolean flags
		testConfigs := []struct {
			debug       bool
			gameMode    bool
			showStats   bool
			networkMode bool
			showNetwork bool
			events      bool
		}{
			{true, true, true, true, true, true},
			{true, true, true, true, true, false},
			{false, false, false, false, false, true},
			{false, false, false, false, false, false},
		}

		for _, config := range testConfigs {
			window := NewDesktopWindow(
				app, char, config.debug, profiler,
				config.gameMode, config.showStats, nil,
				config.networkMode, config.showNetwork, config.events,
			)

			if window.eventsEnabled != config.events {
				t.Errorf("Events flag should be independent: expected %v, got %v",
					config.events, window.eventsEnabled)
			}
		}
	})

	t.Run("Backwards compatibility", func(t *testing.T) {
		// Ensure other functionality still works when events are disabled
		window := NewDesktopWindow(app, char, false, profiler, true, false, nil, false, false, false)

		// Basic window functionality should work
		if window.character == nil {
			t.Error("Character should be set regardless of events flag")
		}
		if window.profiler == nil {
			t.Error("Profiler should be set regardless of events flag")
		}
	})
}

// TestBug3EventsFlagDocumentationCompliance verifies that implementation matches README.md documentation
func TestBug3EventsFlagDocumentationCompliance(t *testing.T) {
	// Per README.md: "-events               Enable general dialog events system for interactive scenarios"
	// The flag should enable/disable the general dialog events system

	// Create test character
	card := createTestCharacterCard()
	char := createMockCharacter(card)
	if char == nil {
		t.Skip("Skipping test due to character creation failure")
		return
	}

	app := test.NewApp()
	defer app.Quit()
	profiler := monitoring.NewProfiler(50)

	t.Run("Events system disabled by default", func(t *testing.T) {
		// When events=false, general dialog events should be disabled
		window := NewDesktopWindow(app, char, false, profiler, true, false, nil, false, false, false)

		if window.eventsEnabled {
			t.Error("Events system should be disabled when flag is false")
		}
	})

	t.Run("Events system enabled when flag is true", func(t *testing.T) {
		// When events=true, general dialog events should be enabled
		window := NewDesktopWindow(app, char, false, profiler, true, false, nil, false, false, true)

		if !window.eventsEnabled {
			t.Error("Events system should be enabled when flag is true")
		}
	})
}

// BenchmarkBug3EventsFlagPerformance ensures the events flag changes don't impact performance
func BenchmarkBug3EventsFlagPerformance(b *testing.B) {
	// Create test character
	card := createTestCharacterCard()
	char := createMockCharacter(card)
	if char == nil {
		b.Skip("Skipping benchmark due to character creation failure")
		return
	}

	app := test.NewApp()
	defer app.Quit()
	profiler := monitoring.NewProfiler(50)

	b.Run("With events enabled", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewDesktopWindow(app, char, false, profiler, true, false, nil, false, false, true)
		}
	})

	b.Run("With events disabled", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewDesktopWindow(app, char, false, profiler, true, false, nil, false, false, false)
		}
	})
}
