// Package ui provides desktop companion user interface components
// This file implements a context menu widget for right-click interactions
package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ContextMenu displays a right-click context menu for character interactions.
// 
// Design Philosophy:
// - Follows the same widget architecture as DialogBubble and StatsOverlay for consistency
// - Uses standard Fyne components (widget.Button) rather than custom implementations
// - Maintains library-first philosophy by avoiding platform-specific code
// - Provides auto-hide functionality to prevent UI clutter
//
// Usage:
//   menu := NewContextMenu()
//   menu.SetMenuItems([]ContextMenuItem{
//       {Text: "Action 1", Callback: func() { /* handle action */ }},
//       {Text: "Action 2", Callback: func() { /* handle action */ }},
//   })
//   menu.Show()
type ContextMenu struct {
	widget.BaseWidget
	background *canvas.Rectangle
	content    *fyne.Container
	visible    bool
	menuItems  []*widget.Button
	callbacks  []func()
}

// ContextMenuItem represents a single menu item with text and callback.
// The callback function is executed when the menu item is selected.
// If Callback is nil, the menu will still hide when the item is clicked.
type ContextMenuItem struct {
	Text     string
	Callback func()
}

// NewContextMenu creates a new context menu widget.
// The menu is initially hidden and has no items.
// 
// Returns a fully initialized ContextMenu widget that follows Fyne's
// widget pattern and can be added to any container layout.
func NewContextMenu() *ContextMenu {
	menu := &ContextMenu{
		menuItems: make([]*widget.Button, 0),
		callbacks: make([]func(), 0),
		visible:   false,
	}

	// Create background rectangle with menu styling
	menu.background = canvas.NewRectangle(color.RGBA{R: 240, G: 240, B: 240, A: 245})
	menu.background.StrokeColor = color.RGBA{R: 120, G: 120, B: 120, A: 255}
	menu.background.StrokeWidth = 1

	// Create container for menu items (initially empty)
	menu.content = container.NewVBox()

	// Initially hidden
	menu.visible = false

	menu.ExtendBaseWidget(menu)
	return menu
}

// SetMenuItems configures the menu items to display
// Takes a slice of ContextMenuItem for flexibility
func (m *ContextMenu) SetMenuItems(items []ContextMenuItem) {
	// Clear existing items
	m.menuItems = make([]*widget.Button, 0, len(items))
	m.callbacks = make([]func(), 0, len(items))

	// Create buttons for each menu item
	for _, item := range items {
		// Capture the callback in a closure to avoid loop variable issues
		callback := item.Callback
		
		// Create button with consistent styling
		btn := widget.NewButton(item.Text, func() {
			// Hide menu when item is clicked
			m.Hide()
			// Execute the callback
			if callback != nil {
				callback()
			}
		})

		// Style the button for menu appearance
		btn.Importance = widget.LowImportance

		m.menuItems = append(m.menuItems, btn)
		m.callbacks = append(m.callbacks, callback)
	}

	// Rebuild the content container
	m.rebuildContent()
}

// rebuildContent recreates the content container with current menu items
func (m *ContextMenu) rebuildContent() {
	// Create objects slice with background first
	objects := []fyne.CanvasObject{m.background}

	// Add all menu buttons
	for _, btn := range m.menuItems {
		objects = append(objects, btn)
	}

	// Create new container with border layout
	// Background fills the entire area, buttons are arranged vertically
	buttonContainer := container.NewVBox()
	for _, btn := range m.menuItems {
		buttonContainer.Add(btn)
	}

	m.content = container.NewBorder(nil, nil, nil, nil, m.background, buttonContainer)
	
	// Update size based on content
	m.updateSize()
}

// updateSize calculates appropriate menu size for the items
// Following the same pattern as DialogBubble.updateSize()
func (m *ContextMenu) updateSize() {
	width, height := m.calculateMenuDimensions()
	menuX, menuY := m.calculateMenuPosition()
	m.applyMenuLayout(width, height, menuX, menuY)
}

// calculateMenuDimensions computes the menu width and height based on items
func (m *ContextMenu) calculateMenuDimensions() (float32, float32) {
	if len(m.menuItems) == 0 {
		return 0, 0
	}

	// Calculate width based on longest text
	minWidth := float32(120)
	maxWidth := float32(200)
	
	// Estimate width from button text (rough calculation)
	width := minWidth
	for _, btn := range m.menuItems {
		textWidth := float32(len(btn.Text)) * 8 + 20 // 8px per char + padding
		if textWidth > width {
			width = textWidth
		}
	}
	
	if width > maxWidth {
		width = maxWidth
	}

	// Calculate height based on number of items
	itemHeight := float32(32) // Standard button height
	padding := float32(4)     // Padding between items
	height := float32(len(m.menuItems))*itemHeight + padding*2

	return width, height
}

// calculateMenuPosition determines the menu position relative to the character
// Positioned to the right of the character, similar to DialogBubble positioning
func (m *ContextMenu) calculateMenuPosition() (float32, float32) {
	menuX := float32(10)  // Small offset from character
	menuY := float32(-20) // Slight offset above character center
	return menuX, menuY
}

// applyMenuLayout applies the calculated dimensions and position to UI components
func (m *ContextMenu) applyMenuLayout(width, height, menuX, menuY float32) {
	if width == 0 || height == 0 {
		return
	}

	// Update container size and position
	m.content.Resize(fyne.NewSize(width, height))
	m.content.Move(fyne.NewPos(menuX, menuY))

	// Update background to match container
	m.background.Resize(fyne.NewSize(width, height))

	// Update button sizes to fit within the menu
	buttonWidth := width - 8 // Account for padding
	buttonHeight := float32(28)
	
	for i, btn := range m.menuItems {
		btn.Resize(fyne.NewSize(buttonWidth, buttonHeight))
		btn.Move(fyne.NewPos(4, 4+float32(i)*32)) // 4px padding, 32px spacing
	}
}

// Show displays the context menu
// Following the same pattern as DialogBubble.Show()
func (m *ContextMenu) Show() {
	m.visible = true
	m.content.Show()
	m.Refresh()
}

// ShowAtPosition displays the context menu at a specific position
// Additional convenience method for right-click positioning
func (m *ContextMenu) ShowAtPosition(x, y float32) {
	// Update position before showing
	m.content.Move(fyne.NewPos(x, y))
	m.Show()
}

// Hide hides the context menu
func (m *ContextMenu) Hide() {
	m.visible = false
	m.content.Hide()
	m.Refresh()
}

// IsVisible returns whether the menu is currently visible
func (m *ContextMenu) IsVisible() bool {
	return m.visible
}

// CreateRenderer creates the Fyne renderer for the context menu
func (m *ContextMenu) CreateRenderer() fyne.WidgetRenderer {
	return &contextMenuRenderer{
		menu:    m,
		content: m.content,
	}
}

// contextMenuRenderer implements fyne.WidgetRenderer for context menus
type contextMenuRenderer struct {
	menu    *ContextMenu
	content *fyne.Container
}

// Layout arranges the context menu components
func (r *contextMenuRenderer) Layout(size fyne.Size) {
	if r.menu.visible && r.content != nil {
		r.content.Resize(r.content.Size())
		r.content.Move(r.content.Position())
	}
}

// MinSize returns the minimum size for the context menu
func (r *contextMenuRenderer) MinSize() fyne.Size {
	if r.menu.visible {
		return fyne.NewSize(120, 32)
	}
	return fyne.NewSize(0, 0)
}

// Objects returns the canvas objects for rendering
func (r *contextMenuRenderer) Objects() []fyne.CanvasObject {
	if r.menu.visible && r.content != nil {
		return []fyne.CanvasObject{r.content}
	}
	return []fyne.CanvasObject{}
}

// Refresh redraws the context menu
func (r *contextMenuRenderer) Refresh() {
	if r.menu.visible && r.content != nil {
		r.content.Refresh()
	}
}

// Destroy cleans up context menu resources
func (r *contextMenuRenderer) Destroy() {
	// No special cleanup needed
}
