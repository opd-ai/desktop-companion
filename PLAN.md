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

## Feature 4: Romance Memory Highlights (1.4 hours) - 🚀 READY TO IMPLEMENT

### Overview
Add context menu option to view recent romance interactions and relationship milestones.

### Files to Modify
- `internal/ui/menu.go` - Add "View Romance History" to context menu
- `internal/ui/romance_history_dialog.go` - New formatted romance memory display
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

## Feature 5: Friendship Compatibility Scoring (1.6 hours)

### Overview
Calculate and display compatibility percentages between network characters based on personality traits.

### Files to Modify
- `internal/character/compatibility.go` - New compatibility calculation logic
- `internal/ui/network_overlay.go` - Add compatibility display
- `internal/character/card.go` - Extend personality system

### Implementation Steps

1. **Create Compatibility Calculator (45 minutes)**
   ```go
   // New file: internal/character/compatibility.go
   type CompatibilityCalculator struct {
       localPersonality    *PersonalityConfig
       networkPersonalities map[string]*PersonalityConfig
   }

   func NewCompatibilityCalculator(localChar *Character) *CompatibilityCalculator {
       // Initialize with local character's personality
   }

   func (cc *CompatibilityCalculator) CalculateCompatibility(peerPersonality *PersonalityConfig) float64 {
       if cc.localPersonality == nil || peerPersonality == nil {
           return 0.5 // Neutral compatibility
       }

       // Calculate compatibility based on personality trait differences
       totalScore := 0.0
       traitCount := 0

       for trait, localValue := range cc.localPersonality.Traits {
           if peerValue, exists := peerPersonality.Traits[trait]; exists {
               // Calculate compatibility score for this trait
               difference := math.Abs(localValue - peerValue)
               score := 1.0 - difference // Closer values = higher compatibility
               totalScore += score
               traitCount++
           }
       }

       if traitCount == 0 {
           return 0.5
       }

       return totalScore / float64(traitCount)
   }
   ```

2. **Extend Network Overlay (35 minutes)**
   ```go
   // In network_overlay.go
   type NetworkOverlay struct {
       // ... existing fields ...
       compatibilityCalculator *character.CompatibilityCalculator
   }

   func (no *NetworkOverlay) updateCharacterCompatibility() {
       if no.compatibilityCalculator == nil {
           return
       }

       // Update compatibility scores for all network characters
       for peerID, charInfo := range no.networkCharacters {
           if charInfo.Personality != nil {
               compatibility := no.compatibilityCalculator.CalculateCompatibility(charInfo.Personality)
               no.displayCompatibility(peerID, compatibility)
           }
       }
   }

   func (no *NetworkOverlay) displayCompatibility(peerID string, score float64) {
       percentage := int(score * 100)
       compatibilityText := fmt.Sprintf("Compatibility: %d%%", percentage)
       
       // Add to character display with color coding
       color := no.getCompatibilityColor(score)
       // Update UI to show compatibility score
   }

   func (no *NetworkOverlay) getCompatibilityColor(score float64) color.Color {
       switch {
       case score >= 0.8:
           return color.RGBA{R: 0, G: 255, B: 0, A: 255} // Green
       case score >= 0.6:
           return color.RGBA{R: 255, G: 255, B: 0, A: 255} // Yellow
       case score >= 0.4:
           return color.RGBA{R: 255, G: 165, B: 0, A: 255} // Orange
       default:
           return color.RGBA{R: 255, G: 0, B: 0, A: 255} // Red
       }
   }
   ```

3. **Integration and Updates (20 minutes)**
   ```go
   // In window.go
   func initializeNetworkFeatures(dw *DesktopWindow, networkMode bool, networkManager NetworkManagerInterface, showNetwork bool, char *character.Character) {
       if networkMode && networkManager != nil {
           dw.networkOverlay = NewNetworkOverlay(char, networkManager)
           dw.networkMode = true
           
           // Initialize compatibility calculator
           if char.GetCard().HasRomanceFeatures() {
               dw.networkOverlay.SetCompatibilityCalculator(
                   character.NewCompatibilityCalculator(char))
           }
       }
   }
   ```

### Testing Requirements
- Test compatibility calculation accuracy
- Verify UI color coding
- Test network character updates

---

## Feature 6: Random Event Frequency Tuning (1.3 hours)

### Overview
Allow users to adjust random event frequency through context menu settings.

### Files to Modify
- `internal/character/behavior.go` - Add frequency multiplier field
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

## Feature 7: Gift Giving Cooldown Indicators (1.7 hours)

### Overview
Add visual cooldown timers to the gift interface to prevent spam clicking.

### Files to Modify
- `internal/ui/gift_dialog.go` - Add cooldown display to gift interface
- `internal/character/game_state.go` - Extend gift memory tracking
- `internal/ui/cooldown_timer.go` - New timer widget

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

## Feature 9: Network Peer Activity Feed (1.9 hours)

### Overview
Display a scrollable log of recent network peer actions within the network overlay.

### Files to Modify
- `internal/ui/network_overlay.go` - Add activity feed component
- `internal/network/activity_tracker.go` - New activity tracking
- `internal/ui/activity_feed.go` - New feed widget

### Implementation Steps

1. **Create Activity Tracker (35 minutes)**
   ```go
   // New file: internal/network/activity_tracker.go
   type ActivityType int

   const (
       ActivityJoined ActivityType = iota
       ActivityLeft
       ActivityInteraction
       ActivityStateChange
       ActivityChat
   )

   type ActivityEvent struct {
       Type        ActivityType  `json:"type"`
       PeerID      string        `json:"peerID"`
       CharacterName string      `json:"characterName"`
       Description string        `json:"description"`
       Timestamp   time.Time     `json:"timestamp"`
       Details     interface{}   `json:"details,omitempty"`
   }

   type ActivityTracker struct {
       mu         sync.RWMutex
       events     []ActivityEvent
       maxEvents  int
       listeners  []func(ActivityEvent)
   }

   func NewActivityTracker(maxEvents int) *ActivityTracker {
       return &ActivityTracker{
           events:    make([]ActivityEvent, 0),
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

## Feature 10: Dialog Response Favorites (1.6 hours)

### Overview
Allow users to mark favorite dialog responses with higher AI selection probability.

### Files to Modify
- `internal/character/game_state.go` - Extend DialogMemory with favorites
- `internal/ui/chatbot_interface.go` - Add favorite star buttons
- `internal/dialog/backend.go` - Modify selection logic

### Implementation Steps

1. **Extend Dialog Memory System (25 minutes)**
   ```go
   // In game_state.go
   type DialogMemory struct {
       // ... existing fields ...
       IsFavorite       bool    `json:"isFavorite"`
       FavoriteWeight   float64 `json:"favoriteWeight"`
   }

   func (gs *GameState) MarkDialogAsFavorite(memoryIndex int, isFavorite bool) error {
       gs.mu.Lock()
       defer gs.mu.Unlock()
       
       if memoryIndex < 0 || memoryIndex >= len(gs.DialogMemories) {
           return fmt.Errorf("invalid memory index")
       }
       
       gs.DialogMemories[memoryIndex].IsFavorite = isFavorite
       if isFavorite {
           gs.DialogMemories[memoryIndex].FavoriteWeight = 2.0 // Double weight
       } else {
           gs.DialogMemories[memoryIndex].FavoriteWeight = 1.0
       }
       
       return nil
   }

   func (gs *GameState) GetFavoriteDialogs() []DialogMemory {
       gs.mu.RLock()
       defer gs.mu.RUnlock()
       
       favorites := make([]DialogMemory, 0)
       for _, memory := range gs.DialogMemories {
           if memory.IsFavorite {
               favorites = append(favorites, memory)
           }
       }
       
       return favorites
   }
   ```

2. **Enhance Chatbot Interface (50 minutes)**
   ```go
   // In chatbot_interface.go
   type ChatbotInterface struct {
       // ... existing fields ...
       favoriteButtons map[int]*widget.Button
   }

   func (ci *ChatbotInterface) addMessage(sender, message string, memoryIndex int) {
       // ... existing message display logic ...
       
       // Add favorite button for character responses
       if sender != "You" && memoryIndex >= 0 {
           favoriteButton := ci.createFavoriteButton(memoryIndex, message)
           messageContainer := container.NewHBox(messageWidget, favoriteButton)
           ci.content.Add(messageContainer)
           
           ci.favoriteButtons[memoryIndex] = favoriteButton
       } else {
           ci.content.Add(messageWidget)
       }
   }

   func (ci *ChatbotInterface) createFavoriteButton(memoryIndex int, message string) *widget.Button {
       gameState := ci.character.GetGameState()
       if gameState == nil {
           return nil
       }
       
       // Check if already favorited
       memories := gameState.GetDialogMemories()
       isFavorite := false
       if memoryIndex < len(memories) {
           isFavorite = memories[memoryIndex].IsFavorite
       }
       
       // Create star button
       starIcon := "☆" // Empty star
       if isFavorite {
           starIcon = "★" // Filled star
       }
       
       button := widget.NewButton(starIcon, func() {
           ci.toggleFavorite(memoryIndex)
       })
       
       button.Resize(fyne.NewSize(24, 24))
       return button
   }

   func (ci *ChatbotInterface) toggleFavorite(memoryIndex int) {
       gameState := ci.character.GetGameState()
       if gameState == nil {
           return
       }
       
       memories := gameState.GetDialogMemories()
       if memoryIndex >= len(memories) {
           return
       }
       
       newFavoriteState := !memories[memoryIndex].IsFavorite
       err := gameState.MarkDialogAsFavorite(memoryIndex, newFavoriteState)
       if err != nil {
           log.Printf("Error toggling favorite: %v", err)
           return
       }
       
       // Update button appearance
       if button, exists := ci.favoriteButtons[memoryIndex]; exists {
           if newFavoriteState {
               button.SetText("★")
           } else {
               button.SetText("☆")
           }
           button.Refresh()
       }
   }
   ```

3. **Modify Dialog Backend Selection (25 minutes)**
   ```go
   // In dialog/backend.go (or wherever dialog selection occurs)
   func (db *DialogBackend) SelectResponse(context DialogContext, gameState *GameState) string {
       // Get favorite responses for context
       favorites := gameState.GetFavoriteDialogs()
       
       // Build weighted response pool
       responsePool := make([]WeightedResponse, 0)
       
       // Add favorites with higher weight
       for _, favorite := range favorites {
           if db.matchesContext(favorite, context) {
               responsePool = append(responsePool, WeightedResponse{
                   Text:   favorite.Response,
                   Weight: favorite.FavoriteWeight,
               })
           }
       }
       
       // Add regular responses with normal weight
       for _, response := range db.getContextResponses(context) {
           responsePool = append(responsePool, WeightedResponse{
               Text:   response,
               Weight: 1.0,
           })
       }
       
       // Select weighted random response
       return db.selectWeightedResponse(responsePool)
   }

   type WeightedResponse struct {
       Text   string
       Weight float64
   }

   func (db *DialogBackend) selectWeightedResponse(responses []WeightedResponse) string {
       if len(responses) == 0 {
           return "I'm not sure what to say..."
       }
       
       totalWeight := 0.0
       for _, response := range responses {
           totalWeight += response.Weight
       }
       
       random := rand.Float64() * totalWeight
       currentWeight := 0.0
       
       for _, response := range responses {
           currentWeight += response.Weight
           if random <= currentWeight {
               return response.Text
           }
       }
       
       // Fallback
       return responses[0].Text
   }
   ```

### Testing Requirements
- Test favorite marking and unmarking
- Verify weighted selection algorithm
- Test UI button state synchronization

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