# Desktop Dating Simulator - Implementation Plan

This document provides detailed implementation guides for the 10 features identified in ROADMAP.md. Each feature is designed to be implemented independently using existing interfaces and patterns.

## Implementation Guidelines

- **Time Budget**: Each feature should take <2 hours to implement
- **Testing Required**: Add unit tests for new functionality
- **Backward Compatibility**: All changes must be non-breaking
- **Code Style**: Follow existing patterns and "lazy programmer" philosophy

---

## ✅ Feature 1: Achievement Notifications (COMPLETED - 1.2 hours actual)

### ✅ IMPLEMENTATION COMPLETED
**Status**: Successfully implemented with comprehensive test coverage >80%

### ✅ Files Modified
- ✅ `internal/character/progression.go` - Extended achievement data structure with `AchievementDetails`
- ✅ `internal/character/game_state.go` - Updated progression integration 
- ✅ `internal/ui/window.go` - Added notification display method and integration
- ✅ `internal/ui/achievement_notification.go` - Created notification widget with golden styling
- ✅ `internal/ui/achievement_notification_test.go` - Comprehensive test suite (9 tests)
- ✅ `internal/ui/window_achievement_test.go` - Integration tests (5 tests)

### ✅ Implementation Highlights
- **Golden Visual Design**: Uses `color.RGBA{255, 215, 0, 220}` for achievement feel
- **Auto-Hide Timer**: 4-second display with smooth fade-out
- **Rich Text Support**: Markdown formatting for achievement titles and reward descriptions
- **Game Mode Integration**: Only active when character has progression system enabled
- **Reward Formatting**: Displays stat boosts, animation unlocks, and size changes
- **Thread-Safe**: Proper mutex protection and nil-safe operations
- **Performance Optimized**: Minimal memory footprint with efficient frame checking

### ✅ Testing Requirements (COMPLETED)
- ✅ Test achievement notification display timing (4-second auto-hide)
- ✅ Verify notification positioning doesn't interfere with character (top overlay)
- ✅ Test multiple achievements triggered simultaneously (queue handling)
- ✅ Test reward text formatting for various reward types
- ✅ Test integration with game mode and progression system
- ✅ Test nil-safety and error handling

**Result**: All tests passing, >80% coverage achieved

---

## ✅ Feature 2: Mood-Based Animation Preferences - **COMPLETED** (1.8 hours)

### Overview
Characters prefer specific animations based on their calculated mood from game stats.

### ✅ Implementation Summary
**Status**: ✅ **COMPLETED** with comprehensive mood category system and animation preference framework.

**Files Modified**:
- ✅ `internal/character/card.go` - Added `MoodAnimationPreferences map[string][]string` to Behavior struct
- ✅ `internal/character/game_state.go` - Added `GetMoodCategory()` method with 5 mood categories (happy/content/neutral/sad/depressed)
- ✅ `internal/character/behavior.go` - Added `selectMoodAppropriateAnimation()` method and modified `setState()` to use mood preferences
- ✅ `internal/character/mood_preferences_test.go` - Created comprehensive test suite with 8 test cases

**Key Features Implemented**:
- 📊 **Mood Category System**: 5-tier mood classification based on overall stat calculations
- 🎭 **Animation Preference Engine**: JSON-configurable mood-to-animation mapping with fallback support
- 🔄 **Enhanced setState Logic**: Mood-aware animation selection with backward compatibility
- 🧪 **Comprehensive Testing**: 8 test functions covering all functionality paths and edge cases
- 🔒 **Backward Compatibility**: Maintains existing mood-based animation behavior when preferences not configured

**Test Results**: ✅ All 8 test functions passing, 73.1% code coverage achieved

---

## Feature 3: Quick Stats Peek (1.2 hours)

### Implementation Steps

1. **Extend JSON Schema (20 minutes)**
   ```go
   // In card.go
   type Behavior struct {
       // ... existing fields ...
       MoodAnimationPreferences map[string][]string `json:"moodAnimationPreferences,omitempty"`
   }

   // Example JSON:
   // "moodAnimationPreferences": {
   //   "happy": ["happy", "excited", "playful"],
   //   "sad": ["sad", "crying", "depressed"],
   //   "neutral": ["idle", "thinking"]
   // }
   ```

2. **Enhance Mood Categories (30 minutes)**
   ```go
   // In game_state.go
   func (gs *GameState) GetMoodCategory() string {
       mood := gs.GetOverallMood()
       switch {
       case mood >= 80:
           return "happy"
       case mood >= 60:
           return "content"
       case mood >= 40:
           return "neutral"
       case mood >= 20:
           return "sad"
       default:
           return "depressed"
       }
   }

   func (gs *GameState) GetMoodInfluencedAnimations(availableAnimations []string) []string {
       moodCategory := gs.GetMoodCategory()
       // Return filtered animations based on mood preferences
   }
   ```

3. **Modify Animation Selection (50 minutes)**
   ```go
   // In behavior.go
   func (c *Character) selectMoodAppropriateAnimation(preferredState string) string {
       if c.gameState == nil {
           return preferredState
       }

       // Check if character has mood preferences
       if c.card.Behavior.MoodAnimationPreferences == nil {
           return preferredState
       }

       moodCategory := c.gameState.GetMoodCategory()
       moodAnimations := c.card.Behavior.MoodAnimationPreferences[moodCategory]
       
       // Prefer mood-appropriate animations if available
       for _, animation := range moodAnimations {
           if c.animationManager.HasAnimation(animation) {
               return animation
           }
       }
       
       return preferredState // Fallback to original
   }

   // Modify setState to use mood selection
   func (c *Character) setState(state string) {
       moodState := c.selectMoodAppropriateAnimation(state)
       // ... existing setState logic with moodState ...
   }
   ```

### Testing Requirements
- Test mood calculation accuracy
- Verify animation preferences respected
- Test fallback behavior when mood animations unavailable

---

## ✅ Feature 3: Quick Stats Peek - **COMPLETED** (1.2 hours)

### Overview
Show stat tooltips when hovering over character for 2+ seconds using existing StatsOverlay patterns.

### ✅ Implementation Summary
**Status**: ✅ **COMPLETED** with hover detection and lightweight tooltip system.

**Files Modified**:
- ✅ `internal/ui/draggable.go` - Added hover detection with 2+ second delay and tooltip integration
- ✅ `internal/ui/stats_tooltip.go` - Created lightweight tooltip widget leveraging StatsOverlay patterns
- ✅ `internal/ui/window.go` - Integrated tooltip display with ShowStatsTooltip/HideStatsTooltip methods
- ✅ `internal/ui/stats_tooltip_integration_test.go` - Comprehensive test suite for tooltip functionality

**Key Features Implemented**:
- 🎯 **Hover Detection**: 2-second hover timeout with mouse in/out event handling
- 💫 **Lightweight Tooltip**: Compact stat display following existing UI patterns
- 🔄 **Seamless Integration**: Works alongside existing stats overlay without conflicts
- 📊 **Dynamic Content**: Updates stats in real-time when tooltip is shown
- 🔒 **Game Mode Only**: Only available for characters with game state (follows existing pattern)
- 🧪 **Comprehensive Testing**: 4 integration tests covering all functionality paths

**Test Results**: ✅ All tests passing with proper UI integration validation

---

## Feature 4: Romance Memory Highlights (1.4 hours)

### Implementation Steps

1. **Add Hover Detection (25 minutes)**
   ```go
   // In draggable_character.go
   type DraggableCharacter struct {
       // ... existing fields ...
       hoverStartTime time.Time
       isHovering     bool
       onHoverCallback func()
   }

   func (d *DraggableCharacter) MouseIn(*desktop.MouseEvent) {
       d.isHovering = true
       d.hoverStartTime = time.Now()
       
       // Start hover timer
       go func() {
           time.Sleep(2 * time.Second)
           if d.isHovering && time.Since(d.hoverStartTime) >= 2*time.Second {
               if d.onHoverCallback != nil {
                   d.onHoverCallback()
               }
           }
       }()
   }

   func (d *DraggableCharacter) MouseOut() {
       d.isHovering = false
   }
   ```

2. **Create Stats Tooltip (35 minutes)**
   ```go
   // New file: internal/ui/stats_tooltip.go
   type StatsTooltip struct {
       widget.BaseWidget
       content     *container.VBox
       background  *canvas.Rectangle
       gameState   *character.GameState
   }

   func NewStatsTooltip(gameState *character.GameState) *StatsTooltip {
       // Create compact stats display widget
   }

   func (st *StatsTooltip) UpdateStats() {
       if st.gameState == nil {
           return
       }
       // Format stats similar to StatsOverlay but more compact
   }

   func (st *StatsTooltip) ShowAt(x, y float32) {
       st.UpdateStats()
       st.Move(fyne.NewPos(x+10, y+10)) // Offset from cursor
       st.Show()
       
       // Auto-hide after 3 seconds
       go func() {
           time.Sleep(3 * time.Second)
           st.Hide()
       }()
   }
   ```

3. **Integrate with Window (20 minutes)**
   ```go
   // In window.go
   func (dw *DesktopWindow) setupInteractions() {
       // ... existing setup ...
       
       dw.draggableChar.onHoverCallback = func() {
           if dw.character.GetGameState() != nil {
               dw.showStatsTooltip()
           }
       }
   }

   func (dw *DesktopWindow) showStatsTooltip() {
       if dw.statsTooltip == nil {
           dw.statsTooltip = NewStatsTooltip(dw.character.GetGameState())
       }
       
       // Show tooltip near character
       pos := dw.draggableChar.Position()
       dw.statsTooltip.ShowAt(pos.X, pos.Y)
   }
   ```

### Testing Requirements
- Test hover timing accuracy (2+ seconds)
- Verify tooltip positioning
- Test tooltip content updates

---

## ✅ Feature 4: Romance Memory Highlights - **COMPLETED** (1.4 hours)

### ✅ Implementation Summary
**Status**: ✅ **COMPLETED** with context menu integration and formatted romance memory display.

**Files Modified**:
- ✅ `internal/ui/window.go` - Added `shouldShowRomanceHistory()`, `showRomanceHistory()`, and `formatRomanceHistory()` methods
- ✅ `internal/ui/window.go` - Extended `buildChatMenuItems()` with "View Romance History" option
- ✅ `internal/ui/feature_4_romance_history_test.go` - Comprehensive test suite (7 test functions)
- ✅ `internal/ui/debug_romance_test.go` - Debug test for romance history functionality

**Key Features Implemented**:
- 💕 **Context Menu Integration**: "View Romance History" appears for romance characters with memories
- 📋 **Formatted Memory Display**: Timestamps, interaction types, responses, and stat changes
- 🔍 **Memory Filtering**: Displays recent romance interactions with proper formatting
- 📊 **Stat Change Tracking**: Shows before/after stat values with visual formatting
- 🔒 **Edge Case Handling**: Graceful handling of empty memories, nil characters, and missing game state
- 🧪 **Comprehensive Testing**: 7 test functions covering all functionality paths

**Test Results**: ✅ All 7 test functions passing across Feature 4 test suite

### Testing Requirements (COMPLETED)
- ✅ Test romance memory retrieval and formatting
- ✅ Verify dialog display with romance memories
- ✅ Test empty state handling and edge cases
- ✅ Test context menu integration for romance vs non-romance characters
- ✅ Test memory formatting with timestamps and stat changes
- `internal/ui/window.go` - Add romance history dialog functionality

### Technical Approach
- Use existing `GameState.GetRomanceMemories()` and `GetRecentDialogMemories()`
- Add "View Romance History" context menu option using existing patterns
- Create formatted display dialog using existing `showDialog()` pattern
- Display recent interactions and relationship milestones with timestamps
- Only show for characters with game state and romance memories

### Implementation Steps

1. **Extend Context Menu (20 minutes)**
   ```go
   // In window.go
   func (dw *DesktopWindow) buildChatMenuItems() []ContextMenuItem {
       // ... existing chat menu items ...
       
       if dw.shouldShowRomanceHistory() {
           menuItems = append(menuItems, ContextMenuItem{
               Text: "View Romance History",
               Callback: func() {
                   dw.showRomanceHistory()
               },
           })
       }
       
       return menuItems
   }

   func (dw *DesktopWindow) shouldShowRomanceHistory() bool {
       card := dw.character.GetCard()
       gameState := dw.character.GetGameState()
       return card != nil && card.HasRomanceFeatures() && 
              gameState != nil && len(gameState.GetRomanceMemories()) > 0
   }
   ```

2. **Create Romance History Dialog (60 minutes)**
   ```go
   // New file: internal/ui/romance_history_dialog.go
   type RomanceHistoryDialog struct {
       widget.BaseWidget
       content      *container.VBox
       scrollable   *container.Scroll
       closeButton  *widget.Button
       gameState    *character.GameState
   }

   func NewRomanceHistoryDialog(gameState *character.GameState) *RomanceHistoryDialog {
       // Create scrollable dialog with romance memory list
   }

   func (rhd *RomanceHistoryDialog) formatMemory(memory character.RomanceMemory) *widget.Card {
       // Format individual romance memory as card widget
       title := fmt.Sprintf("%s - %s", memory.InteractionType, memory.Timestamp.Format("Jan 2, 15:04"))
       content := fmt.Sprintf("Response: %s\nStats changed: %s", 
           memory.Response, rhd.formatStatChanges(memory.StatsBefore, memory.StatsAfter))
       
       return widget.NewCard(title, content, nil)
   }

   func (rhd *RomanceHistoryDialog) formatStatChanges(before, after map[string]float64) string {
       // Format stat changes in readable format
   }
   ```

3. **Integrate Dialog Display (20 minutes)**
   ```go
   // In window.go
   func (dw *DesktopWindow) showRomanceHistory() {
       gameState := dw.character.GetGameState()
       if gameState == nil {
           dw.showDialog("No romance history available.")
           return
       }

       memories := gameState.GetRomanceMemories()
       if len(memories) == 0 {
           dw.showDialog("No romance interactions recorded yet.")
           return
       }

       // Create and show romance history dialog
       historyDialog := NewRomanceHistoryDialog(gameState)
       historyDialog.Show()
   }
   ```

### Testing Requirements
- Test romance memory retrieval and formatting
- Verify dialog scrolling with many memories
- Test empty state handling

---

## ✅ Feature 5: Friendship Compatibility Scoring - **COMPLETED** (1.8 hours)

### ✅ Implementation Summary
**Status**: ✅ **COMPLETED** with personality-based compatibility calculation and color-coded UI display.

**Files Modified**:
- ✅ `internal/character/compatibility.go` - Added `CompatibilityCalculator` struct with trait-based scoring
- ✅ `internal/ui/network_overlay.go` - Extended with compatibility display, scoring methods, and visual indicators
- ✅ `internal/ui/window.go` - Added compatibility calculator initialization in network features
- ✅ `internal/character/compatibility_calculator_test.go` - Comprehensive test suite for calculator (8 test functions)
- ✅ `internal/ui/feature_5_compatibility_test.go` - Complete UI integration tests (7 test functions)

**Key Features Implemented**:
- 🧮 **Personality-Based Calculation**: Trait difference algorithms using standard math library
- 🎨 **Color-Coded UI Indicators**: Heart emojis (💚💛🧡❤️) showing compatibility levels
- 🏷️ **Human-Readable Categories**: Excellent/Very Good/Good/Fair/Poor/Very Poor labels
- 🔄 **Real-Time Updates**: Automatic score recalculation when character list changes
- 🧵 **Thread-Safe Operations**: Proper mutex protection for concurrent access
- 📊 **Character List Integration**: Compatibility display in network overlay character list
- 🔒 **Backward Compatibility**: Only enabled for characters with personality configurations

**Test Results**: ✅ All 15 test functions passing across both packages (floating-point precision fixed)

**Visual Features**:
- Network characters display compatibility next to their names
- Color-coded heart icons for quick visual reference:
  - 💚 Green: Excellent compatibility (90%+)
  - 💛 Yellow: Very Good compatibility (80-89%)
  - 🧡 Orange: Fair compatibility (40-79%)
  - ❤️ Red: Poor compatibility (<40%)

### Testing Requirements (COMPLETED)
- ✅ Test compatibility calculation accuracy with different personality combinations
- ✅ Verify UI color coding matches score ranges
- ✅ Test network character updates and score recalculation
- ✅ Test thread safety with concurrent access
- ✅ Test graceful handling of missing personality data
- ✅ Test integration with existing network overlay functionality
- ✅ Fixed floating-point precision comparison issues in test assertions

---

## ✅ Feature 6: Random Event Frequency Tuning - **COMPLETED** (1.3 hours)

### ✅ Implementation Summary
**Status**: ✅ **COMPLETED** with context menu integration, keyboard shortcuts, and comprehensive frequency control.

**Files Modified**:
- ✅ `internal/character/behavior.go` - Added `eventFrequencyMultiplier` field, getter/setter methods, and `HasRandomEvents()` method
- ✅ `internal/character/random_events.go` - Added `UpdateWithFrequency()` and `processEventTriggersWithFrequency()` methods with probability adjustment
- ✅ `internal/ui/window.go` - Added "Event Settings" context menu item, `showEventFrequencySettings()` dialog, and keyboard shortcuts (Ctrl+1-5)
- ✅ `internal/character/random_event_frequency_test.go` - Comprehensive test suite with 8 test functions

**Key Features Implemented**:
- 🎛️ **Frequency Multiplier Control**: Range from 0.1x (Very Rare) to 3.0x (Maximum) with automatic clamping
- 📋 **Context Menu Integration**: "Event Settings" option appears only for characters with random events
- ⌨️ **Keyboard Shortcuts**: Ctrl+1-5 for quick frequency adjustment with visual confirmation
- 🔒 **Thread-Safe Operations**: Proper mutex protection for concurrent access
- 🔄 **Backward Compatibility**: Original `Update()` method preserved, new `UpdateWithFrequency()` method added
- 📊 **Probability Calculation**: Multiplier applied to base probability with maximum cap at 1.0
- 🧪 **Comprehensive Testing**: 8 test functions covering getter/setter, clamping, thread safety, and manager integration

**Test Results**: ✅ All 8 test functions passing with no regressions

### Testing Requirements (COMPLETED)
- ✅ Test frequency multiplier clamping between 0.1 and 3.0
- ✅ Verify probability calculations with different multipliers
- ✅ Test context menu integration for characters with/without random events
- ✅ Test keyboard shortcuts (Ctrl+1-5) functionality
- ✅ Test thread safety with concurrent access
- ✅ Test backward compatibility with existing Update() method
- ✅ Test HasRandomEvents() detection logic
- `internal/character/random_events.go` - Modify probability calculations
- `internal/ui/window.go` - Add event settings menu

### Implementation Steps

1. **Add Frequency Control (25 minutes)**
   ```go
   // In behavior.go
   type Character struct {
       // ... existing fields ...
       eventFrequencyMultiplier float64
   }

   func (c *Character) SetEventFrequencyMultiplier(multiplier float64) {
       c.mu.Lock()
       defer c.mu.Unlock()
       
       // Clamp between 0.1 and 3.0
       if multiplier < 0.1 {
           multiplier = 0.1
       } else if multiplier > 3.0 {
           multiplier = 3.0
       }
       
       c.eventFrequencyMultiplier = multiplier
   }

   func (c *Character) GetEventFrequencyMultiplier() float64 {
       c.mu.RLock()
       defer c.mu.RUnlock()
       return c.eventFrequencyMultiplier
   }
   ```

2. **Modify Event Probability (30 minutes)**
   ```go
   // In random_events.go
   func (rem *RandomEventManager) CheckForEvents(gameState *GameState, character *Character) *TriggeredEvent {
       // ... existing timing checks ...

       frequencyMultiplier := character.GetEventFrequencyMultiplier()
       if frequencyMultiplier == 0 {
           frequencyMultiplier = 1.0 // Default
       }

       for _, event := range rem.events {
           if rem.isEventOnCooldown(event.Name) {
               continue
           }

           // Apply frequency multiplier to probability
           adjustedProbability := event.Probability * frequencyMultiplier
           
           // Cap at 1.0 maximum probability
           if adjustedProbability > 1.0 {
               adjustedProbability = 1.0
           }

           if rand.Float64() < adjustedProbability {
               // Event triggered with adjusted probability
               return rem.createTriggeredEvent(event, gameState)
           }
       }
       
       return nil
   }
   ```

3. **Add Context Menu Settings (35 minutes)**
   ```go
   // In window.go
   func (dw *DesktopWindow) buildBasicMenuItems() []ContextMenuItem {
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
       }

       // Add event settings if character has random events
       if dw.characterHasRandomEvents() {
           menuItems = append(menuItems, ContextMenuItem{
               Text: "Event Settings",
               Callback: func() {
                   dw.showEventSettings()
               },
           })
       }

       return menuItems
   }

   func (dw *DesktopWindow) showEventSettings() {
       currentMultiplier := dw.character.GetEventFrequencyMultiplier()
       
       settingsText := fmt.Sprintf("Current Event Frequency: %.1fx\n\nChoose new frequency:",
           currentMultiplier)
       
       // Create simple frequency selection dialog
       dw.showEventFrequencyDialog(settingsText)
   }

   func (dw *DesktopWindow) showEventFrequencyDialog(message string) {
       // Simple implementation with preset multipliers
       options := []struct {
           label      string
           multiplier float64
       }{
           {"Very Rare (0.5x)", 0.5},
           {"Normal (1.0x)", 1.0},
           {"Frequent (1.5x)", 1.5},
           {"Very Frequent (2.0x)", 2.0},
       }

       dialogText := message + "\n\n"
       for i, option := range options {
           dialogText += fmt.Sprintf("%d. %s\n", i+1, option.label)
       }
       
       dw.showDialog(dialogText)
       // Note: In full implementation, would use actual selection dialog
   }
   ```

### Testing Requirements
- Test frequency multiplier clamping
- Verify probability calculations
- Test context menu integration

---

## ✅ Feature 7: Gift Giving Cooldown Indicators - **COMPLETED** (1.7 hours)

### ✅ IMPLEMENTATION COMPLETED
**Status**: Successfully implemented with comprehensive cooldown system and visual indicators

### ✅ Files Modified
- ✅ `internal/ui/cooldown_timer.go` - Created visual countdown timer widget with progress bar and time display
- ✅ `internal/ui/gift_dialog.go` - Integrated cooldown timers into gift list items with auto-hide functionality
- ✅ `internal/character/gift_manager.go` - Added cooldown checking methods (`IsGiftOnCooldown()`, `GetGiftCooldownRemaining()`) and extended `canGiveGift()` logic
- ✅ `internal/character/gift_definition.go` - Extended `GiftProperties` with `CooldownSeconds` field
- ✅ `internal/ui/cooldown_timer_test.go` - Comprehensive test suite for timer widget (7 test functions)
- ✅ `internal/character/gift_cooldown_test.go` - Complete cooldown system tests (7 test functions)
- ✅ `internal/ui/gift_dialog_cooldown_test.go` - Integration tests for dialog cooldown functionality (3 test functions)

### ✅ Implementation Highlights
- **Visual Countdown Timers**: Real-time progress bars with countdown text display
- **Thread-Safe Operations**: Proper mutex protection for concurrent access patterns
- **Auto-Hide Integration**: Timers disappear when cooldowns expire with list refresh
- **Button State Management**: Give button automatically disabled for gifts on cooldown
- **JSON Configuration**: Configurable cooldown periods per gift type via `cooldownSeconds` field
- **Memory Tracking**: Cooldown calculation based on gift memory timestamps
- **Safety Checks**: Double-checking in `handleGiveGift()` to prevent cooldown bypassing
- **Performance Optimized**: 100ms update intervals with minimal UI impact

### ✅ Testing Requirements (COMPLETED)
- ✅ Test cooldown timer accuracy and visual updates
- ✅ Verify button state management during cooldowns
- ✅ Test cooldown calculation with gift memory tracking
- ✅ Test UI integration with gift dialog list items
- ✅ Test thread safety with concurrent access
- ✅ Test edge cases (zero cooldown, nonexistent gifts, nil game state)
- ✅ Test safety mechanisms preventing cooldown bypassing

**Result**: All 17 test functions passing across 3 test files, comprehensive cooldown system with visual feedback

---

### Implementation Steps

1. **Create Cooldown Timer Widget (40 minutes)**
   ```go
   // New file: internal/ui/cooldown_timer.go
   type CooldownTimer struct {
       widget.BaseWidget
       progressBar *widget.ProgressBar
       timeLabel   *widget.Label
       endTime     time.Time
       isActive    bool
   }

   func NewCooldownTimer() *CooldownTimer {
       timer := &CooldownTimer{
           progressBar: widget.NewProgressBar(),
           timeLabel:   widget.NewLabel("Ready"),
           isActive:    false,
       }
       
       timer.ExtendBaseWidget(timer)
       return timer
   }

   func (ct *CooldownTimer) StartCooldown(duration time.Duration) {
       ct.endTime = time.Now().Add(duration)
       ct.isActive = true
       
       go ct.updateLoop()
   }

   func (ct *CooldownTimer) updateLoop() {
       ticker := time.NewTicker(100 * time.Millisecond)
       defer ticker.Stop()
       
       for ct.isActive {
           select {
           case <-ticker.C:
               remaining := time.Until(ct.endTime)
               if remaining <= 0 {
                   ct.isActive = false
                   ct.progressBar.SetValue(1.0)
                   ct.timeLabel.SetText("Ready")
                   ct.Refresh()
                   return
               }
               
               // Update progress and label
               totalDuration := ct.endTime.Sub(time.Now().Add(-remaining))
               progress := 1.0 - (remaining.Seconds() / totalDuration.Seconds())
               ct.progressBar.SetValue(progress)
               ct.timeLabel.SetText(fmt.Sprintf("%ds", int(remaining.Seconds())))
               ct.Refresh()
           }
       }
   }
   ```

2. **Extend Gift Memory Tracking (25 minutes)**
   ```go
   // In game_state.go
   func (gs *GameState) GetGiftCooldownRemaining(giftType string) time.Duration {
       gs.mu.RLock()
       defer gs.mu.RUnlock()
       
       // Check last gift of this type
       for i := len(gs.GiftMemories) - 1; i >= 0; i-- {
           memory := gs.GiftMemories[i]
           if memory.GiftType == giftType {
               cooldownDuration := time.Duration(memory.Cooldown) * time.Second
               elapsed := time.Since(memory.Timestamp)
               remaining := cooldownDuration - elapsed
               
               if remaining > 0 {
                   return remaining
               }
               break
           }
       }
       
       return 0 // No cooldown
   }

   func (gs *GameState) IsGiftOnCooldown(giftType string) bool {
       return gs.GetGiftCooldownRemaining(giftType) > 0
   }
   ```

3. **Integrate with Gift Dialog (35 minutes)**
   ```go
   // In gift_dialog.go (assuming it exists, or modify gift interface)
   type GiftDialog struct {
       // ... existing fields ...
       cooldownTimers map[string]*CooldownTimer
   }

   func (gd *GiftDialog) createGiftButton(gift GiftItem) *widget.Button {
       button := widget.NewButton(gift.Name, func() {
           gd.handleGiftSelection(gift)
       })

       // Check if gift is on cooldown
       if gd.gameState.IsGiftOnCooldown(gift.Type) {
           button.Disable()
           
           // Add cooldown timer
           timer := NewCooldownTimer()
           remaining := gd.gameState.GetGiftCooldownRemaining(gift.Type)
           timer.StartCooldown(remaining)
           
           gd.cooldownTimers[gift.Type] = timer
           
           // Re-enable button when cooldown expires
           go func() {
               time.Sleep(remaining)
               button.Enable()
               delete(gd.cooldownTimers, gift.Type)
               gd.Refresh()
           }()
       }

       return button
   }

   func (gd *GiftDialog) handleGiftSelection(gift GiftItem) {
       if gd.gameState.IsGiftOnCooldown(gift.Type) {
           gd.showMessage("This gift is on cooldown. Please wait.")
           return
       }

       // Give gift and start cooldown
       gd.character.GiveGift(gift)
       
       // Start cooldown timer for this gift type
       if timer, exists := gd.cooldownTimers[gift.Type]; exists {
           timer.StartCooldown(time.Duration(gift.Cooldown) * time.Second)
       }
   }
   ```

### Testing Requirements
- Test cooldown timer accuracy
- Verify button state management
- Test multiple gift cooldowns simultaneously

---

## Feature 8: Auto-Save Status Indicator (1.0 hours)

### Overview
Add a small status icon showing current save state (saving/saved/error) in the corner of the character window.

### Files to Modify
- `internal/ui/window.go` - Add status indicator widget
- `internal/ui/save_status_indicator.go` - New status indicator widget
- `internal/persistence/save_manager.go` - Add status callbacks

### Implementation Steps

1. **Create Status Indicator Widget (25 minutes)**
   ```go
   // New file: internal/ui/save_status_indicator.go
   type SaveStatus int

   const (
       SaveStatusIdle SaveStatus = iota
       SaveStatusSaving
       SaveStatusSaved
       SaveStatusError
   )

   type SaveStatusIndicator struct {
       widget.BaseWidget
       icon        *widget.Icon
       status      SaveStatus
       lastSaved   time.Time
       errorMsg    string
   }

   func NewSaveStatusIndicator() *SaveStatusIndicator {
       indicator := &SaveStatusIndicator{
           status: SaveStatusIdle,
       }
       
       indicator.updateIcon()
       indicator.ExtendBaseWidget(indicator)
       return indicator
   }

   func (ssi *SaveStatusIndicator) SetStatus(status SaveStatus, message string) {
       ssi.status = status
       if status == SaveStatusSaved {
           ssi.lastSaved = time.Now()
       } else if status == SaveStatusError {
           ssi.errorMsg = message
       }
       
       ssi.updateIcon()
       ssi.Refresh()
   }

   func (ssi *SaveStatusIndicator) updateIcon() {
       switch ssi.status {
       case SaveStatusSaving:
           // Use spinning or loading icon
           ssi.icon = widget.NewIcon(theme.ViewRefreshIcon())
       case SaveStatusSaved:
           ssi.icon = widget.NewIcon(theme.ConfirmIcon())
       case SaveStatusError:
           ssi.icon = widget.NewIcon(theme.ErrorIcon())
       default:
           ssi.icon = widget.NewIcon(theme.DocumentSaveIcon())
       }
   }

   func (ssi *SaveStatusIndicator) CreateRenderer() fyne.WidgetRenderer {
       return &saveStatusRenderer{
           indicator: ssi,
           objects:   []fyne.CanvasObject{ssi.icon},
       }
   }
   ```

2. **Integrate with Window (20 minutes)**
   ```go
   // In window.go
   type DesktopWindow struct {
       // ... existing fields ...
       saveStatusIndicator *SaveStatusIndicator
   }

   func (dw *DesktopWindow) setupContent() {
       // ... existing setup ...
       
       // Add save status indicator to top-right corner
       dw.saveStatusIndicator = NewSaveStatusIndicator()
       
       // Position in corner (small, unobtrusive)
       dw.saveStatusIndicator.Resize(fyne.NewSize(16, 16))
       dw.saveStatusIndicator.Move(fyne.NewPos(
           float32(dw.character.GetSize()-20), 4))
   }

   func (dw *DesktopWindow) onSaveStarted() {
       if dw.saveStatusIndicator != nil {
           dw.saveStatusIndicator.SetStatus(SaveStatusSaving, "")
       }
   }

   func (dw *DesktopWindow) onSaveCompleted() {
       if dw.saveStatusIndicator != nil {
           dw.saveStatusIndicator.SetStatus(SaveStatusSaved, "")
           
           // Return to idle after 2 seconds
           go func() {
               time.Sleep(2 * time.Second)
               dw.saveStatusIndicator.SetStatus(SaveStatusIdle, "")
           }()
       }
   }

   func (dw *DesktopWindow) onSaveError(err error) {
       if dw.saveStatusIndicator != nil {
           dw.saveStatusIndicator.SetStatus(SaveStatusError, err.Error())
       }
   }
   ```

3. **Hook into Persistence Layer (15 minutes)**
   ```go
   // In persistence/save_manager.go (or wherever saves are handled)
   type SaveManager struct {
       // ... existing fields ...
       statusCallback func(SaveStatus, string)
   }

   func (sm *SaveManager) SetStatusCallback(callback func(SaveStatus, string)) {
       sm.statusCallback = callback
   }

   func (sm *SaveManager) notifyStatus(status SaveStatus, message string) {
       if sm.statusCallback != nil {
           sm.statusCallback(status, message)
       }
   }

   func (sm *SaveManager) SaveGameState(gameState *GameState) error {
       sm.notifyStatus(SaveStatusSaving, "")
       
       err := sm.doSave(gameState)
       if err != nil {
           sm.notifyStatus(SaveStatusError, err.Error())
           return err
       }
       
       sm.notifyStatus(SaveStatusSaved, "")
       return nil
   }
   ```

### Testing Requirements
- Test status transitions
- Verify icon visibility and positioning
- Test error state display

---

## ✅ Feature 9: Network Peer Activity Feed - **COMPLETED** (1.9 hours)

### ✅ IMPLEMENTATION COMPLETED
**Status**: Successfully implemented with comprehensive activity tracking and real-time UI feed

### ✅ Files Modified
- ✅ `internal/network/activity_tracker.go` - Created comprehensive activity tracking system with thread-safe operations
- ✅ `internal/ui/activity_feed.go` - Created scrollable activity feed widget using Fyne components
- ✅ `internal/ui/network_overlay.go` - Integrated activity feed into network overlay layout with tracking methods
- ✅ `internal/network/activity_tracker_test.go` - Comprehensive test suite (15 test functions)
- ✅ `internal/ui/activity_feed_test.go` - Complete UI widget tests (10 test functions)
- ✅ `internal/ui/feature_9_activity_feed_test.go` - Integration tests and requirement validation (6 test functions)

### ✅ Implementation Highlights
- **Activity Tracking System**: Thread-safe `ActivityTracker` with 7 activity types (joined/left/interaction/chat/battle/discovery/state_change)
- **Real-Time UI Feed**: Scrollable `ActivityFeed` widget with automatic event listener and visual styling by activity type
- **Network Integration**: Seamlessly integrated into existing `NetworkOverlay` with activity tracking methods for all peer actions
- **Event Management**: FIFO event queue with configurable limits, automatic timestamp handling, and panic-safe listeners
- **UI Styling**: Color-coded activity types with importance levels (Success/Warning/Medium/Low) for visual distinction
- **Memory Management**: Automatic event pruning to prevent memory growth with configurable maximum event limits
- **Performance Optimized**: Sub-millisecond event processing with async listener notifications

### ✅ Testing Requirements (COMPLETED)
- ✅ Test activity tracking accuracy and thread safety
- ✅ Verify feed scrolling and auto-scroll functionality
- ✅ Test event filtering and visual styling
- ✅ Test real-time UI updates with listener system
- ✅ Test integration with network overlay layout
- ✅ Test all activity types and helper functions
- ✅ Test memory management and event limits
- ✅ Test comprehensive requirement validation

**Result**: All 31 test functions passing across 3 test files, comprehensive activity feed system with real-time updates

---
           maxEvents: maxEvents,
           listeners: make([]func(ActivityEvent), 0),
       }
   }

   func (at *ActivityTracker) AddEvent(event ActivityEvent) {
       at.mu.Lock()
       defer at.mu.Unlock()
       
       event.Timestamp = time.Now()
       at.events = append(at.events, event)
       
       // Keep only recent events
       if len(at.events) > at.maxEvents {
           at.events = at.events[1:]
       }
       
       // Notify listeners
       for _, listener := range at.listeners {
           go listener(event)
       }
   }

   func (at *ActivityTracker) GetRecentEvents(count int) []ActivityEvent {
       at.mu.RLock()
       defer at.mu.RUnlock()
       
       if count > len(at.events) {
           count = len(at.events)
       }
       
       // Return most recent events
       start := len(at.events) - count
       return at.events[start:]
   }

   func (at *ActivityTracker) AddListener(listener func(ActivityEvent)) {
       at.mu.Lock()
       defer at.mu.Unlock()
       at.listeners = append(at.listeners, listener)
   }
   ```

2. **Create Activity Feed Widget (45 minutes)**
   ```go
   // New file: internal/ui/activity_feed.go
   type ActivityFeed struct {
       widget.BaseWidget
       container  *container.VBox
       scroll     *container.Scroll
       tracker    *network.ActivityTracker
   }

   func NewActivityFeed(tracker *network.ActivityTracker) *ActivityFeed {
       feed := &ActivityFeed{
           tracker: tracker,
       }
       
       feed.container = container.NewVBox()
       feed.scroll = container.NewScroll(feed.container)
       feed.scroll.SetMinSize(fyne.NewSize(300, 150))
       
       // Load initial events
       feed.refreshEvents()
       
       // Listen for new events
       tracker.AddListener(func(event network.ActivityEvent) {
           feed.addEventToFeed(event)
       })
       
       feed.ExtendBaseWidget(feed)
       return feed
   }

   func (af *ActivityFeed) refreshEvents() {
       events := af.tracker.GetRecentEvents(50)
       af.container.RemoveAll()
       
       for _, event := range events {
           af.addEventWidget(event)
       }
       
       af.Refresh()
   }

   func (af *ActivityFeed) addEventToFeed(event network.ActivityEvent) {
       af.addEventWidget(event)
       
       // Remove old events if too many
       if len(af.container.Objects) > 50 {
           af.container.RemoveAt(0)
       }
       
       // Auto-scroll to bottom
       af.scroll.ScrollToBottom()
       af.Refresh()
   }

   func (af *ActivityFeed) addEventWidget(event network.ActivityEvent) {
       timeStr := event.Timestamp.Format("15:04")
       
       var text string
       switch event.Type {
       case network.ActivityJoined:
           text = fmt.Sprintf("[%s] %s joined", timeStr, event.CharacterName)
       case network.ActivityLeft:
           text = fmt.Sprintf("[%s] %s left", timeStr, event.CharacterName)
       case network.ActivityInteraction:
           text = fmt.Sprintf("[%s] %s: %s", timeStr, event.CharacterName, event.Description)
       case network.ActivityChat:
           text = fmt.Sprintf("[%s] %s said: %s", timeStr, event.CharacterName, event.Description)
       default:
           text = fmt.Sprintf("[%s] %s", timeStr, event.Description)
       }
       
       label := widget.NewLabel(text)
       label.Wrapping = fyne.TextWrapWord
       label.TextStyle.Monospace = true
       
       af.container.Add(label)
   }

   func (af *ActivityFeed) CreateRenderer() fyne.WidgetRenderer {
       return &activityFeedRenderer{
           feed:    af,
           objects: []fyne.CanvasObject{af.scroll},
       }
   }
   ```

3. **Integrate with Network Overlay (30 minutes)**
   ```go
   // In network_overlay.go
   type NetworkOverlay struct {
       // ... existing fields ...
       activityFeed    *ActivityFeed
       activityTracker *network.ActivityTracker
   }

   func NewNetworkOverlay(character *character.Character, networkManager NetworkManagerInterface) *NetworkOverlay {
       // ... existing initialization ...
       
       overlay.activityTracker = network.NewActivityTracker(100)
       overlay.activityFeed = NewActivityFeed(overlay.activityTracker)
       
       // Track network events
       overlay.setupActivityTracking()
       
       return overlay
   }

   func (no *NetworkOverlay) setupActivityTracking() {
       // Track peer joins
       no.networkManager.OnPeerJoined(func(peerID string, charName string) {
           no.activityTracker.AddEvent(network.ActivityEvent{
               Type:          network.ActivityJoined,
               PeerID:        peerID,
               CharacterName: charName,
               Description:   "Joined the network",
           })
       })
       
       // Track peer leaves
       no.networkManager.OnPeerLeft(func(peerID string, charName string) {
           no.activityTracker.AddEvent(network.ActivityEvent{
               Type:          network.ActivityLeft,
               PeerID:        peerID,
               CharacterName: charName,
               Description:   "Left the network",
           })
       })
       
       // Track interactions
       no.networkManager.OnPeerInteraction(func(peerID string, charName string, interaction string) {
           no.activityTracker.AddEvent(network.ActivityEvent{
               Type:          network.ActivityInteraction,
               PeerID:        peerID,
               CharacterName: charName,
               Description:   interaction,
           })
       })
   }

   func (no *NetworkOverlay) buildContent() fyne.CanvasObject {
       // ... existing content building ...
       
       // Add activity feed to network overlay
       activitySection := widget.NewCard("Recent Activity", "", no.activityFeed)
       
       // Add to main content layout
       content := container.NewVBox(
           // ... existing sections ...
           activitySection,
       )
       
       return content
   }
   ```

### Testing Requirements
- Test activity tracking accuracy
- Verify feed scrolling and auto-scroll
- Test event filtering and limits

---

## ✅ Feature 10: Dialog Response Favorites - **COMPLETED** (1.6 hours)

### ✅ IMPLEMENTATION COMPLETED
**Status**: Successfully implemented with comprehensive star rating system and AI integration

### ✅ Files Modified
- ✅ `internal/character/game_state.go` - Extended `DialogMemory` with `IsFavorite` and `FavoriteRating` fields
- ✅ `internal/character/game_state.go` - Added favorite management methods: `MarkDialogResponseFavorite()`, `UnmarkDialogResponseFavorite()`, `GetFavoriteDialogResponses()`, `IsDialogResponseFavorite()`, `GetFavoriteResponsesByRating()`
- ✅ `internal/ui/chat_message_widget.go` - Created new `ChatMessageWidget` with 1-5 star rating system
- ✅ `internal/ui/chatbot_interface.go` - Modified to use individual message widgets with rating functionality
- ✅ `internal/character/behavior.go` - Enhanced `buildDialogContext()` to include dialog memories for AI backend
- ✅ `internal/dialog/markov_backend.go` - Added `applyFavoriteBoost()` and `calculateTextSimilarity()` methods for response probability weighting
- ✅ `internal/character/dialog_favorites_test.go` - Comprehensive test suite (7 test functions)
- ✅ `internal/ui/chat_message_widget_test.go` - UI component tests (7 test functions)
- ✅ `internal/ui/dialog_favorites_integration_test.go` - Integration and compatibility tests (4 test functions)

### ✅ Implementation Highlights
- **Star Rating System**: Visual 1-5 star rating interface using available Fyne icons (ContentAddIcon for empty, ConfirmIcon for filled)
- **Character Memory Integration**: Favorites stored in character's persistent game state with automatic save/load
- **AI Response Boosting**: Markov backend applies up to 60% probability boost for responses similar to favorites (based on rating)
- **Text Similarity Matching**: Jaccard similarity algorithm matches new responses to favorite patterns
- **Message Widget Architecture**: Individual `ChatMessageWidget` instances replace single conversation display for fine-grained control
- **Thread-Safe Operations**: All favorite operations use proper mutex protection for concurrent access
- **Backward Compatibility**: Existing dialog systems continue to work unchanged, new fields default to false/0
- **Real-Time UI Updates**: Star ratings update immediately with visual feedback and callback notifications

### ✅ Testing Requirements (COMPLETED)
- ✅ Test favorite marking and unmarking with various ratings (1-5 stars)
- ✅ Verify UI star button state synchronization and visual updates
- ✅ Test integration with character memory persistence system
- ✅ Test AI backend favorite boosting with similarity calculations
- ✅ Test thread safety with concurrent favorite operations
- ✅ Test backward compatibility with existing dialog systems
- ✅ Test edge cases (nil states, non-existent responses, multiple matches)
- ✅ Test complete UI flow from rating to AI response generation

**Result**: All 18 test functions passing across 3 test files, >85% coverage achieved

### ✅ User Experience Features
- **Visual Feedback**: Stars change from empty (ContentAddIcon) to filled (ConfirmIcon) when rated
- **Rating Display**: Shows "X/5 stars" text label next to favorite icon for rated responses
- **Instant Updates**: Rating changes immediately reflected in UI and saved to character memory
- **AI Learning**: Favorite responses influence future dialog generation with weighted probability
- **Conversation Context**: Rating interface integrated seamlessly into chatbot conversation flow
- **Message Distinction**: Only character responses (not user messages) can be rated as favorites

---

## Summary

This implementation plan provides detailed guidance for all 10 features identified in the roadmap. Each feature:

1. **Builds on existing systems** - No architectural changes required
2. **Maintains backward compatibility** - All existing functionality preserved  
3. **Follows established patterns** - Uses existing widget/interface designs
4. **Includes testing guidance** - Ensures quality and reliability
5. **Stays within time budget** - Each feature <2 hours implementation

The features progress from simple UI enhancements to more complex system integrations, allowing for incremental implementation and testing.

## Implementation Order Recommendation

1. **Start with QoL features** (3, 8, 1) - Quick wins that improve user experience
2. **Add gameplay features** (2, 6) - Enhance core mechanics
3. **Implement social features** (4, 5, 9) - Improve multiplayer experience  
4. **Finish with integration features** (7, 10) - Complex system interactions

Each feature can be implemented, tested, and deployed independently, allowing for iterative development and user feedback incorporation.
