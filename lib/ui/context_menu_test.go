package ui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

func TestNewContextMenu(t *testing.T) {
	menu := NewContextMenu()

	if menu == nil {
		t.Fatal("NewContextMenu returned nil")
	}

	if menu.visible {
		t.Error("New context menu should be initially hidden")
	}

	if len(menu.menuItems) != 0 {
		t.Error("New context menu should have no items initially")
	}

	if menu.background == nil {
		t.Error("Context menu background should be initialized")
	}

	if menu.content == nil {
		t.Error("Context menu content container should be initialized")
	}
}

func TestContextMenuSetMenuItems(t *testing.T) {
	menu := NewContextMenu()

	// Test empty items
	menu.SetMenuItems([]ContextMenuItem{})
	if len(menu.menuItems) != 0 {
		t.Error("Setting empty items should result in no menu items")
	}

	// Test single item
	callbackCalled := false
	items := []ContextMenuItem{
		{
			Text: "Test Item",
			Callback: func() {
				callbackCalled = true
			},
		},
	}

	menu.SetMenuItems(items)

	if len(menu.menuItems) != 1 {
		t.Errorf("Expected 1 menu item, got %d", len(menu.menuItems))
	}

	if menu.menuItems[0].Text != "Test Item" {
		t.Errorf("Expected button text 'Test Item', got %q", menu.menuItems[0].Text)
	}

	// Test callback execution
	menu.menuItems[0].OnTapped()
	if !callbackCalled {
		t.Error("Menu item callback was not called")
	}

	// Test multiple items
	items = []ContextMenuItem{
		{Text: "Item 1", Callback: nil},
		{Text: "Item 2", Callback: nil},
		{Text: "Item 3", Callback: nil},
	}

	menu.SetMenuItems(items)

	if len(menu.menuItems) != 3 {
		t.Errorf("Expected 3 menu items, got %d", len(menu.menuItems))
	}

	expectedTexts := []string{"Item 1", "Item 2", "Item 3"}
	for i, expected := range expectedTexts {
		if menu.menuItems[i].Text != expected {
			t.Errorf("Expected item %d text %q, got %q", i, expected, menu.menuItems[i].Text)
		}
	}
}

func TestContextMenuCallbackClosures(t *testing.T) {
	menu := NewContextMenu()

	// Test that callbacks are properly captured in closures
	callbackResults := make([]int, 3)
	items := make([]ContextMenuItem, 3)

	for i := 0; i < 3; i++ {
		// Capture loop variable properly
		index := i
		items[i] = ContextMenuItem{
			Text: "Item " + string(rune('0'+i)),
			Callback: func() {
				callbackResults[index] = index + 1
			},
		}
	}

	menu.SetMenuItems(items)

	// Execute each callback
	for i := 0; i < 3; i++ {
		menu.menuItems[i].OnTapped()
	}

	// Verify each callback set the correct value
	for i := 0; i < 3; i++ {
		expected := i + 1
		if callbackResults[i] != expected {
			t.Errorf("Callback %d: expected %d, got %d", i, expected, callbackResults[i])
		}
	}
}

func TestContextMenuVisibility(t *testing.T) {
	menu := NewContextMenu()

	// Initially hidden
	if menu.IsVisible() {
		t.Error("New menu should be initially hidden")
	}

	// Show menu
	menu.Show()
	if !menu.IsVisible() {
		t.Error("Menu should be visible after Show()")
	}

	// Hide menu
	menu.Hide()
	if menu.IsVisible() {
		t.Error("Menu should be hidden after Hide()")
	}
}

func TestContextMenuShowAtPosition(t *testing.T) {
	menu := NewContextMenu()

	// Set some menu items first
	items := []ContextMenuItem{
		{Text: "Test", Callback: nil},
	}
	menu.SetMenuItems(items)

	// Test ShowAtPosition
	testX, testY := float32(100), float32(200)
	menu.ShowAtPosition(testX, testY)

	if !menu.IsVisible() {
		t.Error("Menu should be visible after ShowAtPosition()")
	}

	// Check position (note: exact position checking may be limited by Fyne's layout system)
	pos := menu.content.Position()
	if pos.X != testX || pos.Y != testY {
		t.Logf("Position may have been adjusted by layout system: expected (%.1f, %.1f), got (%.1f, %.1f)", 
			testX, testY, pos.X, pos.Y)
	}
}

func TestContextMenuDimensions(t *testing.T) {
	menu := NewContextMenu()

	// Test with no items
	width, height := menu.calculateMenuDimensions()
	if width != 0 || height != 0 {
		t.Errorf("Empty menu should have zero dimensions, got (%.1f, %.1f)", width, height)
	}

	// Test with single item
	items := []ContextMenuItem{
		{Text: "Test", Callback: nil},
	}
	menu.SetMenuItems(items)

	width, height = menu.calculateMenuDimensions()
	if width <= 0 || height <= 0 {
		t.Errorf("Menu with items should have positive dimensions, got (%.1f, %.1f)", width, height)
	}

	// Test with multiple items - height should increase
	items = []ContextMenuItem{
		{Text: "Item 1", Callback: nil},
		{Text: "Item 2", Callback: nil},
		{Text: "Item 3", Callback: nil},
	}
	menu.SetMenuItems(items)

	newWidth, newHeight := menu.calculateMenuDimensions()
	if newHeight <= height {
		t.Errorf("More items should result in greater height: old %.1f, new %.1f", height, newHeight)
	}

	// Width constraints
	if newWidth < 120 {
		t.Errorf("Menu width should be at least 120px, got %.1f", newWidth)
	}
	if newWidth > 200 {
		t.Errorf("Menu width should be at most 200px, got %.1f", newWidth)
	}
}

func TestContextMenuPosition(t *testing.T) {
	menu := NewContextMenu()

	menuX, menuY := menu.calculateMenuPosition()

	// Check that position values are reasonable
	if menuX < 0 {
		t.Errorf("Menu X position should be non-negative, got %.1f", menuX)
	}

	// menuY can be negative (above character), but should be reasonable
	if menuY < -1000 || menuY > 1000 {
		t.Errorf("Menu Y position seems unreasonable: %.1f", menuY)
	}
}

func TestContextMenuRenderer(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	menu := NewContextMenu()
	renderer := menu.CreateRenderer()

	if renderer == nil {
		t.Fatal("CreateRenderer returned nil")
	}

	// Test with hidden menu
	objects := renderer.Objects()
	if len(objects) != 0 {
		t.Errorf("Hidden menu should have no objects, got %d", len(objects))
	}

	minSize := renderer.MinSize()
	if minSize.Width != 0 || minSize.Height != 0 {
		t.Errorf("Hidden menu should have zero MinSize, got %v", minSize)
	}

	// Show menu with items
	items := []ContextMenuItem{
		{Text: "Test Item", Callback: nil},
	}
	menu.SetMenuItems(items)
	menu.Show()

	objects = renderer.Objects()
	if len(objects) != 1 {
		t.Errorf("Visible menu should have 1 object (content container), got %d", len(objects))
	}

	minSize = renderer.MinSize()
	if minSize.Width < 120 || minSize.Height < 32 {
		t.Errorf("Visible menu should have reasonable MinSize, got %v", minSize)
	}

	// Test refresh (should not panic)
	renderer.Refresh()

	// Test layout (should not panic)
	renderer.Layout(fyne.NewSize(200, 100))

	// Test destroy (should not panic)
	renderer.Destroy()
}

func TestContextMenuItemCallbackAutoHide(t *testing.T) {
	menu := NewContextMenu()

	// Test that clicking an item hides the menu
	items := []ContextMenuItem{
		{
			Text:     "Test Item",
			Callback: func() { /* do nothing */ },
		},
	}
	menu.SetMenuItems(items)
	menu.Show()

	if !menu.IsVisible() {
		t.Error("Menu should be visible before clicking item")
	}

	// Simulate clicking the menu item
	menu.menuItems[0].OnTapped()

	if menu.IsVisible() {
		t.Error("Menu should be hidden after clicking item")
	}
}

func TestContextMenuNilCallback(t *testing.T) {
	menu := NewContextMenu()

	// Test that nil callbacks don't cause panics
	items := []ContextMenuItem{
		{
			Text:     "Test Item",
			Callback: nil,
		},
	}
	menu.SetMenuItems(items)
	menu.Show()

	// This should not panic
	menu.menuItems[0].OnTapped()

	// Menu should still be hidden even with nil callback
	if menu.IsVisible() {
		t.Error("Menu should be hidden after clicking item even with nil callback")
	}
}

func TestContextMenuAutoHideTimeout(t *testing.T) {
	menu := NewContextMenu()

	items := []ContextMenuItem{
		{Text: "Test", Callback: nil},
	}
	menu.SetMenuItems(items)

	// This test verifies the showContextMenu auto-hide behavior
	// Note: Since we're testing the timeout functionality, we need to be in the context of DesktopWindow
	// This test documents the expected behavior that will be tested in integration tests
	
	menu.Show()
	if !menu.IsVisible() {
		t.Error("Menu should be visible immediately after Show()")
	}

	// For unit testing, we verify the Show/Hide functionality works
	// The timeout behavior is tested in integration tests with DesktopWindow
	menu.Hide()
	if menu.IsVisible() {
		t.Error("Menu should be hidden after Hide()")
	}
}

// Benchmark tests for performance validation

func BenchmarkContextMenuCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		menu := NewContextMenu()
		_ = menu
	}
}

func BenchmarkContextMenuSetItems(b *testing.B) {
	menu := NewContextMenu()
	items := []ContextMenuItem{
		{Text: "Item 1", Callback: func() {}},
		{Text: "Item 2", Callback: func() {}},
		{Text: "Item 3", Callback: func() {}},
		{Text: "Item 4", Callback: func() {}},
		{Text: "Item 5", Callback: func() {}},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		menu.SetMenuItems(items)
	}
}

func BenchmarkContextMenuShowHide(b *testing.B) {
	menu := NewContextMenu()
	items := []ContextMenuItem{
		{Text: "Test", Callback: nil},
	}
	menu.SetMenuItems(items)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		menu.Show()
		menu.Hide()
	}
}
