# Battle UI Components

This document describes the battle-related user interface components implemented for the JRPG battle system.

## Overview

The battle UI system provides the following components:

1. **Battle Action Dialog** - For selecting battle actions during combat
2. **Battle Result Overlay** - For displaying battle action results
3. **Context Menu Integration** - Battle initiation option in right-click menu

## Components

### BattleActionDialog

A dialog that allows players to select battle actions during combat.

**Features:**
- 11 battle actions: Attack, Defend, Stun, Heal, Boost, Counter, Drain, Shield, Charge, Evade, Taunt
- Optional turn timer with countdown display
- 3-column grid layout for efficient action selection
- Auto-hide and cancel functionality
- Callback system for action selection and cancellation

**Usage:**
```go
dialog := NewBattleActionDialog(30 * time.Second) // 30-second turn timer
dialog.SetOnActionSelect(func(action BattleActionType) {
    // Handle the selected action
})
dialog.SetOnCancel(func() {
    // Handle timeout or cancellation
})
dialog.Show()
```

### BattleResultOverlay

A temporary overlay that displays the results of battle actions.

**Features:**
- Color-coded display (green for success, red for failure)
- Detailed information about damage, healing, and status effects
- Auto-hide after 3 seconds for results, 2 seconds for simple messages
- Support for both detailed battle results and simple messages

**Usage:**
```go
overlay := NewBattleResultOverlay()

// Show detailed battle result
result := BattleResult{
    Success:    true,
    ActionType: ActionAttack,
    Damage:     25.0,
    Response:   "Take that!",
}
overlay.ShowResult(result)

// Or show simple message
overlay.ShowMessage("Battle Started", "Get ready for combat!")
```

### Context Menu Integration

Battle functionality is integrated into the character's right-click context menu.

**Features:**
- Only shows for characters with battle system enabled (`HasBattleSystem()`)
- "Initiate Battle" option in context menu
- Placeholder implementation ready for full battle system integration

## Battle Action Types

The following battle actions are supported:

| Action | Description |
|--------|-------------|
| Attack | Deal damage to opponent |
| Defend | Reduce incoming damage |
| Stun | Disable opponent temporarily |
| Heal | Restore hit points |
| Boost | Increase attack power |
| Counter | Reactive counter-attack |
| Drain | Absorb opponent's energy |
| Shield | Create protective barrier |
| Charge | Build energy for next attack |
| Evade | Avoid next attack |
| Taunt | Provoke opponent |

## Design Philosophy

### Consistency
- Follows the same widget architecture as existing UI components (DialogBubble, ContextMenu)
- Uses standard Fyne components rather than custom implementations
- Maintains consistent styling and behavior patterns

### User Experience
- Clear visual feedback for all actions
- Intuitive action selection with descriptions
- Non-blocking auto-hide overlays
- Responsive timer displays

### Library-First Approach
- Uses only standard library and well-maintained Fyne components
- No external dependencies beyond the project's existing requirements
- Simple, maintainable code structure

## Testing

The battle UI components have comprehensive test coverage:

- **Unit Tests**: All components have >80% test coverage
- **Integration Tests**: Context menu integration verified
- **Error Handling**: All error paths tested
- **Concurrency**: Timer functionality tested for race conditions

## Integration Points

### Character System
- Integrates with `CharacterCard.HasBattleSystem()` for feature detection
- Uses battle animation constants from the character system

### Battle System
- Ready for integration with the core battle manager
- Action types match the battle system's action definitions
- Result structure aligns with battle system output

### Network System
- Prepared for multiplayer battle scenarios
- Context menu shows appropriate messaging for network availability

## Future Enhancements

The battle UI system is designed to support future enhancements:

1. **Animation Integration**: Battle animations during action selection
2. **Sound Effects**: Audio feedback for actions and results
3. **Multiplayer UI**: Opponent information and turn indicators
4. **Item Selection**: Integration with the gift/item system
5. **Status Effects**: Visual indicators for ongoing effects

## Files

- `battle_dialog.go` - Battle action selection dialog
- `battle_result.go` - Battle result display overlay
- `battle_ui_test.go` - Comprehensive unit tests
- `battle_integration_test.go` - Integration tests
- Window integration in `window.go` (context menu methods)
