# Desktop Companion Right-Click Context Menus & Chatbot Interface Implementation Plan

## 1. Architecture Summary

### Current Component Relationships
The DDS application follows a modular, library-first architecture built on Fyne v2.4.5:

**Core UI Components:**
- `DesktopWindow` - Main transparent overlay window managing all UI components
- `CharacterRenderer` - Handles GIF animation rendering and character display  
- `DialogBubble` - Speech bubble widget for text responses
- `DraggableCharacter` - Wrapper implementing Fyne's drag/click interfaces
- `ClickableWidget` - Invisible overlay for non-draggable character interactions
- `StatsOverlay` - Optional stats display using progress bars

**Interaction Flow:**
1. Mouse events captured by `DraggableCharacter` or `ClickableWidget`
2. Events routed to `DesktopWindow.handleClick()` or `handleRightClick()`
3. Character behavior triggered via `Character.HandleClick()`/`HandleRightClick()`
4. Response text displayed via `DialogBubble.ShowWithText()`

**AI Integration Points:**
- `DialogBackend` interface supports pluggable AI systems (Markov chains, etc.)
- `GeneralEventManager` handles interactive scenarios with choice-based dialogs
- Advanced dialog system operates in parallel with basic interactions
- Character cards configure AI features via JSON `dialogBackend` section

**Existing Right-Click Support:**
- Both `DraggableCharacter.TappedSecondary()` and `ClickableWidget.TappedSecondary()` already handle right-clicks
- Current right-click behavior triggers `Character.HandleRightClick()` → shows dialog bubble
- Game mode enhances right-click with "feed" interaction for Tamagotchi features

## 2. Integration Analysis  

### Files Requiring Modification

**A. Core UI Extensions (3 files, minimal changes):**

**`internal/ui/context_menu.go` (COMPLETED) ✓**
- New `ContextMenu` widget following existing `DialogBubble` pattern
- Inherits from `widget.BaseWidget` with custom renderer
- Menu items as `widget.Button` array with click handlers
- Positioned relative to character like `DialogBubble`
- Comprehensive test coverage with >90% coverage
- Auto-hide functionality with 5-second timeout

**`internal/ui/window.go` (COMPLETED) ✓**
- Added `contextMenu *ContextMenu` field to `DesktopWindow` struct  
- Initialized in `NewDesktopWindow()` after dialog bubble creation
- Added `showContextMenu()` method following `showDialog()` pattern
- Updated `setupContent()` to include context menu in overlay objects
- Dynamic menu generation based on character capabilities and game mode

**`internal/ui/draggable.go` (COMPLETED) ✓**
- Modified `TappedSecondary()` to call `window.showContextMenu()` instead of direct character handling
- Preserves existing click delegation pattern
- Maintains backwards compatibility

**B. Chatbot Interface Extensions (2 files, conditional features):**

**`internal/ui/chatbot_interface.go` (NEW FILE)**
- New `ChatbotInterface` widget for AI-enabled characters only
- Multi-line text input widget with send button
- Conversation history display using scrollable container
- Conditional activation based on `dialogBackend.enabled` in character card

**`internal/character/behavior.go` (1 method addition)**
- Add `HandleChatMessage(message string)` method for chatbot interactions
- Reuses existing dialog backend infrastructure (`dialogManager.GenerateDialog()`)
- Integrates with memory system if enabled

### Rationale for Minimal Changes

**Leverages Existing Patterns:**
- Context menu follows same widget architecture as `DialogBubble` and `StatsOverlay`
- Event routing uses established `DraggableCharacter`/`ClickableWidget` → `DesktopWindow` → `Character` flow
- AI integration reuses existing `DialogBackend` and `GeneralEventManager` systems

**Preserves Fyne Architecture:**
- All new widgets inherit from `widget.BaseWidget` following project standards
- Custom renderers implement `fyne.WidgetRenderer` interface consistently
- Container layouts use `container.NewWithoutLayout()` matching existing overlay pattern

**Maintains Library-First Philosophy:**
- Context menu uses standard `widget.Button` components, not custom implementations
- Chatbot interface built from `widget.Entry`, `widget.Button`, and `container.Scroll`
- No platform-specific code required - pure Fyne implementation

---

## Implementation Status

### COMPLETED ✓

**Core UI Extensions - Context Menu (Phase 1):**
- ✅ `internal/ui/context_menu.go` - Complete context menu widget implementation
- ✅ `internal/ui/window.go` - Integration with DesktopWindow and dynamic menu generation  
- ✅ `internal/ui/draggable.go` - Right-click event routing to context menu
- ✅ Comprehensive test coverage (>90%) with 11 test cases
- ✅ Full GoDoc documentation with usage examples
- ✅ Auto-hide functionality (5-second timeout)
- ✅ Dynamic menu items based on character capabilities and game mode

**Implementation Details:**
- **Lines of Code:** ~290 new lines (context_menu.go), ~60 lines modified (window.go, draggable.go)
- **Test Coverage:** 11 test cases covering functionality, edge cases, and performance
- **Architecture:** Follows existing Fyne widget patterns, zero breaking changes
- **Integration:** Seamless with existing DialogBubble and StatsOverlay patterns

### IN PROGRESS

**Chatbot Interface Extensions (Phase 2):**
- ⏳ `internal/ui/chatbot_interface.go` (NEW FILE) - Next planned implementation
- ⏳ `internal/character/behavior.go` (1 method addition) - HandleChatMessage method

**Implementation Timeline:** 2-3 weeks
**Lines of Code Impact:** ~300-400 new lines across 3 new files, <20 lines modified in existing files
**Risk Level:** LOW - Additive changes with extensive backwards compatibility measures
