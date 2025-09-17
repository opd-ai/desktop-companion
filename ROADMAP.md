# Desktop Dating Simulator (DDS) - Feature Roadmap

## Overview

This roadmap outlines 10 small, safe feature enhancements for the Desktop Dating Simulator that can be implemented without architectural changes. All features leverage existing interfaces and maintain full backward compatibility.

## Codebase Analysis Summary

The DDS codebase follows a well-structured "lazy programmer" philosophy using mature libraries and JSON-first configuration. Key extension points include:

- **CharacterCard** struct with extensive optional fields for JSON configuration
- **GameState** system with configurable stats/interactions and memory tracking
- **Dialog** system with pluggable backends and AI-powered responses
- **UI components** (StatsOverlay, ContextMenu, DialogBubble) built with Fyne widgets
- **Event systems** (RandomEvents, GeneralEvents, GiftSystem) supporting JSON configuration
- Clean interfaces for **HandleClick/HandleRightClick/HandleHover**, **Update()** loops, and **ApplyInteractionEffects()**

## Feature Implementation Plan

### 🎯 Quality of Life Features

#### [1] Quick Action Tooltips
- **Description:** Add hover tooltips showing available interaction shortcuts (click, right-click, double-click) and current cooldown status
- **Implementation:** Extend `StatsOverlay` component to include interaction hints, leverage existing `HandleHover()` and cooldown tracking in `Character.dialogCooldowns`
- **Files Modified:** `lib/ui/stats_overlay.go`, `lib/character/behavior.go`
- **Time Estimate:** 1.5 hours
- **Priority:** Medium
- **Impact:** Quality of Life

#### [2] Interaction History Popup
- **Description:** Add context menu option to view recent interaction history with timestamps and character responses
- **Implementation:** Leverage existing `GameState.InteractionHistory` and `ContextMenu` system, add new menu item that displays `DialogMemory` entries in scrollable dialog
- **Files Modified:** `lib/ui/context_menu.go`, `lib/character/game_state.go`
- **Time Estimate:** 1.2 hours
- **Priority:** Medium
- **Impact:** Quality of Life

#### [3] Smart Cooldown Indicators
- **Description:** Show visual cooldown indicators on character when interactions are temporarily unavailable
- **Implementation:** Extend `CharacterRenderer` to overlay small progress circles using existing `Character.dialogCooldowns` and `Character.gameInteractionCooldowns` data
- **Files Modified:** `lib/ui/renderer.go`, `lib/character/behavior.go`
- **Time Estimate:** 1.6 hours
- **Priority:** High
- **Impact:** Quality of Life

#### [4] Quick Stats Peek
- **Description:** Add keyboard shortcut (Tab key) to temporarily show stats overlay for 2 seconds without toggling permanent display
- **Implementation:** Use existing `StatsOverlay.Show()` mechanism with timer auto-hide, integrate with current keyboard handler in `DesktopWindow.setupKeyboardShortcuts()`
- **Files Modified:** `lib/ui/window.go`, `lib/ui/stats_overlay.go`
- **Time Estimate:** 0.8 hours
- **Priority:** Low
- **Impact:** Quality of Life

### 🎮 Gameplay Enhancement Features

#### [5] Mood-Based Idle Animations
- **Description:** Automatically cycle through different idle animations based on character's current mood/stats without user input
- **Implementation:** Extend `Behavior.MoodAnimationPreferences` JSON configuration, use existing `GameState.GetStats()` in `Character.Update()` loop to select appropriate idle animation
- **Files Modified:** `lib/character/behavior.go`, `lib/character/card.go`
- **JSON Schema Extension:** Add `moodAnimationMap` to `Behavior` config
- **Time Estimate:** 1.8 hours
- **Priority:** High
- **Impact:** Gameplay

#### [6] Achievement Toast Notifications
- **Description:** Show brief pop-up notifications when achievements are unlocked with achievement name and reward details
- **Implementation:** Use existing `GameState.GetRecentAchievements()` in `Character.Update()`, create new `AchievementToast` widget similar to `DialogBubble` with auto-hide
- **Files Modified:** `lib/ui/achievement_toast.go` (new), `lib/character/behavior.go`
- **Time Estimate:** 1.7 hours
- **Priority:** Medium
- **Impact:** Gameplay

#### [7] Context-Aware Random Events
- **Description:** Adjust random event probability based on time of day and recent user activity patterns
- **Implementation:** Extend `RandomEventConfig.Probability` calculation using existing `RandomEventManager.Update()`, add time-based modifiers to `Character.eventFrequencyMultiplier`
- **Files Modified:** `lib/character/random_events.go`, `lib/character/behavior.go`
- **JSON Schema Extension:** Add `timeBasedModifiers` to `RandomEventConfig`
- **Time Estimate:** 1.9 hours
- **Priority:** Medium
- **Impact:** Gameplay

### 💬 Social Interaction Features

#### [8] Personality-Driven Greeting Variations
- **Description:** Generate different daily greeting messages based on character personality traits and current relationship level
- **Implementation:** Add new dialog trigger "daily_greeting" to existing `Dialog` system, use `PersonalityConfig.Traits` and `GameState.RelationshipLevel` for response selection
- **Files Modified:** `lib/character/behavior.go`, `lib/character/card.go`
- **JSON Schema Extension:** Add `daily_greeting` trigger support to `Dialog` config
- **Time Estimate:** 1.4 hours
- **Priority:** High
- **Impact:** Social

#### [9] Dialog Memory Favorites
- **Description:** Allow users to mark dialog responses as favorites and replay them from context menu
- **Implementation:** Use existing `DialogMemory.IsFavorite` field, add context menu option to browse/replay favorite responses via `Character.dialogManager`
- **Files Modified:** `lib/ui/context_menu.go`, `lib/character/game_state.go`
- **Time Estimate:** 1.3 hours
- **Priority:** Low
- **Impact:** Social

### 🌐 Integration Features

#### [10] Multi-Language Dialog Support
- **Description:** Add JSON-configurable dialog response translations for basic phrases in multiple languages
- **Implementation:** Extend `Dialog.Responses` to support language maps, add optional `Language` field to `Behavior` config, implement fallback to English if translation missing
- **Files Modified:** `lib/character/card.go`, `lib/character/behavior.go`
- **JSON Schema Extension:** Add `language` field to `Behavior`, extend `Dialog.Responses` to support translation maps
- **Time Estimate:** 1.9 hours
- **Priority:** Low
- **Impact:** Integration

## Implementation Strategy

### Phase 1: High Priority QoL (4.7 hours)
1. Smart Cooldown Indicators (1.6h)
2. Mood-Based Idle Animations (1.8h)
3. Personality-Driven Greeting Variations (1.4h)

### Phase 2: Medium Priority Enhancements (5.6 hours)
4. Quick Action Tooltips (1.5h)
5. Interaction History Popup (1.2h)
6. Achievement Toast Notifications (1.7h)
7. Context-Aware Random Events (1.9h)

### Phase 3: Low Priority Polish (3.9 hours)
8. Quick Stats Peek (0.8h)
9. Dialog Memory Favorites (1.3h)
10. Multi-Language Dialog Support (1.9h)

**Total Estimated Time:** 15.1 hours

## Technical Guidelines

### Implementation Principles
- **Backward Compatibility:** All features must work with existing character cards
- **JSON-First:** New behavior should be configurable through character card JSON
- **Interface Reuse:** Leverage existing interfaces (`HandleClick`, `Update()`, etc.)
- **Library-First:** Use existing Fyne widgets and Go standard library
- **No Architecture Changes:** Work within current system boundaries

### Testing Requirements
- Unit tests for new functionality
- Integration tests with existing character cards
- Validation that features work with all character archetypes
- Performance testing to ensure no FPS impact

### Documentation Updates
- Update `SCHEMA_DOCUMENTATION.md` for any JSON schema extensions
- Add examples to character cards demonstrating new features
- Update README.md with new keyboard shortcuts and interactions

## Success Criteria

✅ All features implementable without modifying core systems  
✅ Each feature references specific existing code elements  
✅ Combined implementation time under 20 hours  
✅ Balanced mix of gameplay, social, QoL, and integration improvements  
✅ No breaking changes to JSON schema or animations  
✅ Maintains the "lazy programmer" philosophy of leveraging existing libraries

## Future Considerations

These features provide a foundation for more advanced enhancements:
- Advanced AI-driven mood analysis for animation selection
- Community-driven translation contributions
- Analytics dashboard for interaction patterns
- Advanced achievement system with branching requirements
- Real-time multiplayer event coordination

---

*Last Updated: September 16, 2025*  
*Status: Planning Phase*  
*Next Review: Upon completion of Phase 1 features*
