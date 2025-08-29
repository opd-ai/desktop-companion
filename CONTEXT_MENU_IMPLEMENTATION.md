# Context Menu Implementation Documentation

## Overview

The Context Menu implementation adds right-click context menus to the Desktop Dating Simulator, replacing the previous direct right-click dialog behavior with a more flexible, user-friendly menu system.

## Architecture

### Component Design

The implementation follows the established Fyne widget architecture used throughout the DDS application:

- **Widget Pattern**: Inherits from `widget.BaseWidget` like `DialogBubble` and `StatsOverlay`
- **Renderer Pattern**: Implements `fyne.WidgetRenderer` for custom drawing
- **Container Layout**: Uses `container.NewWithoutLayout()` matching existing overlay patterns
- **Library-First**: Built entirely with standard Fyne components (`widget.Button`, `canvas.Rectangle`)

### Integration Points

1. **DesktopWindow**: Owns and manages the context menu instance
2. **DraggableCharacter**: Routes right-click events to the window's context menu handler
3. **ClickableWidget**: Also supports right-click events through the same pathway
4. **Character**: Provides the underlying capabilities that determine menu items

## Features

### Dynamic Menu Generation

The context menu dynamically generates items based on:

- **Basic Interactions**: Always includes "Talk" and "About" options
- **Game Mode**: Adds "Feed" and "Play" options when game features are enabled
- **Stats Display**: Adds "Show/Hide Stats" toggle when stats overlay is available
- **Character Capabilities**: Menu items adapt to character configuration

### Auto-Hide Behavior

- **Click Selection**: Menu hides automatically when any item is clicked
- **Timeout**: Menu auto-hides after 5 seconds of inactivity
- **Manual Control**: Can be hidden programmatically via `Hide()` method

### Visual Design

- **Consistent Styling**: Matches existing UI component appearance
- **Responsive Layout**: Adapts size based on text content and number of items
- **Position Awareness**: Positioned relative to character, similar to dialog bubbles

## Usage Example

```go
// In DesktopWindow.showContextMenu()
menuItems := []ContextMenuItem{
    {
        Text: "Talk",
        Callback: func() {
            response := dw.character.HandleClick()
            if response != "" {
                dw.showDialog(response)
            }
        },
    },
    {
        Text: "Feed",
        Callback: func() {
            response := dw.character.HandleGameInteraction("feed")
            if response != "" {
                dw.showDialog(response)
            }
        },
    },
}

dw.contextMenu.SetMenuItems(menuItems)
dw.contextMenu.Show()
```

## Testing Coverage

The implementation includes comprehensive test coverage:

### Unit Tests (11 test cases)
- Widget creation and initialization
- Menu item configuration and callbacks
- Visibility state management
- Position and dimension calculations
- Renderer functionality
- Error handling (nil callbacks, empty menus)
- Performance benchmarks

### Integration Tests
- Right-click event routing through DraggableCharacter
- Window integration and overlay management
- Auto-hide timeout behavior (tested in context)

## Backwards Compatibility

The implementation maintains full backwards compatibility:

- **Existing API**: All existing character interaction methods remain unchanged
- **Event Flow**: Right-click events still reach characters, just through menu selection
- **Game Features**: All game mode interactions remain available
- **Configuration**: No changes required to character cards or settings

## Performance Considerations

- **Lazy Loading**: Menu items are created only when needed
- **Memory Efficient**: No persistent state beyond the widget itself
- **Event Handling**: Uses Fyne's built-in event system for optimal performance
- **Layout Optimization**: Minimal layout calculations, cached dimensions

## Design Decisions

### Why Standard Buttons Instead of Custom Menu Items?

- **Simplicity**: Standard `widget.Button` provides all needed functionality
- **Consistency**: Buttons match the application's visual style automatically
- **Maintenance**: Less custom code to maintain and debug
- **Accessibility**: Inherits Fyne's accessibility features

### Why 5-Second Auto-Hide?

- **User Experience**: Long enough to read options, short enough to avoid clutter
- **Balance**: Matches common desktop application behavior
- **Feedback**: Based on existing 3-second dialog timeout pattern

### Why Dynamic Menu Generation?

- **Flexibility**: Adapts to different character configurations
- **Game Mode**: Seamlessly integrates game features when available
- **Extensibility**: Easy to add new menu items in the future

## Future Enhancements

The architecture supports easy extension for:

- **Custom Menu Items**: Character-specific actions from card configuration
- **Keyboard Shortcuts**: Menu items could include keyboard accelerators
- **Icons**: Menu items could include icons for visual clarity
- **Submenus**: Hierarchical menu structure for complex interactions
- **Context Sensitivity**: Different menus based on character state

## Error Handling

The implementation handles common edge cases:

- **Nil Callbacks**: Menu items with nil callbacks don't cause panics
- **Empty Menus**: Zero menu items result in hidden menu (graceful degradation)
- **Rapid Clicks**: Multiple rapid right-clicks don't create multiple menus
- **State Consistency**: Menu visibility state is properly maintained

This implementation successfully completes Phase 1 of the right-click context menu feature, providing a solid foundation for the next phase (Chatbot Interface Extensions).
