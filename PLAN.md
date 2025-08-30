# RSS/Atom Newsfeed Integration Plan for Desktop Companion

## 1. Architecture Analysis

### Current Codebase Structure

The Desktop Companion application is built with a modular, extensible architecture using Go 1.21+ and Fyne UI framework:

#### Core Components
- **Character System** (`internal/character/`): JSON-configured character cards with behavior, animations, and dialog systems
- **Dialog System** (`internal/dialog/`): Pluggable backend architecture supporting Markov chain generation and custom backends
- **UI System** (`internal/ui/`): Fyne-based transparent overlay windows with interactive components
- **Config System** (`internal/config/`): JSON configuration loading using Go standard library
- **Game State** (`internal/character/game_state.go`): Comprehensive state management with stats, progression, and interactions

#### Key Extension Points
1. **Dialog Backend Interface**: Pluggable system for generating responses
2. **Character Card Schema**: JSON-based configuration with optional features
3. **General Events System**: User-initiated interactive scenarios
4. **Animation System**: GIF-based character animations with state management
5. **UI Overlay System**: Transparent window components with context menus

#### Existing Integration Patterns
- **Markov Chain Backend**: Advanced text generation with personality integration
- **Chatbot Interface**: Real-time conversation system (`internal/ui/chatbot_interface.go`)
- **General Dialog Events**: Interactive scenarios with choice consequences
- **Context Menu System**: Right-click extensibility
- **Stats Overlay**: Real-time information display

### Architecture Strengths for RSS Integration
1. **JSON-first Configuration**: Easy to extend character cards with newsfeed settings
2. **Plugin Architecture**: Dialog backends can be extended for news summarization
3. **Event System**: Existing framework for triggering news-related scenarios
4. **Animation System**: Support for news-reading animations
5. **UI Components**: Reusable overlay patterns for news display

## 2. Integration Strategy

### Minimal Disruption Approach

The RSS integration will leverage existing architectures instead of creating new systems:

1. **Extend Dialog Backend Interface** for news summarization
2. **Add News Events** to existing General Events system
3. **Reuse UI Components** for news display (dialog bubbles, overlays)
4. **Extend Character Cards** with optional news configuration
5. **Utilize Animation System** for news-reading behaviors

### Core Principles
- **Zero Breaking Changes**: All existing character cards continue working unchanged
- **Optional Feature**: News functionality is opt-in through character configuration
- **Library-First**: Use Go standard library (`net/http`, `encoding/xml`) and mature RSS libraries
- **Fyne Integration**: Reuse existing UI components and patterns
- **Performance Conscious**: Async feed fetching with caching

## 3. Component Design

### 3.1 News Backend (Dialog System Extension)

```go
// NewsBlogBackend implements DialogBackend interface for news summarization
type NewsBlogBackend struct {
    feeds        []RSSFeed
    cache        *NewsCache
    summarizer   *NewsSummarizer
    updateTimer  *time.Timer
    debug        bool
}

// RSSFeed represents a single news source
type RSSFeed struct {
    URL           string   `json:"url"`
    Name          string   `json:"name"`
    Category      string   `json:"category"`      // "tech", "gaming", "general"
    UpdateFreq    int      `json:"updateFreq"`    // Minutes between updates
    MaxItems      int      `json:"maxItems"`      // Maximum items to store
    Keywords      []string `json:"keywords"`      // Filter keywords
    Enabled       bool     `json:"enabled"`
}

// NewsItem represents a single news article
type NewsItem struct {
    Title       string    `json:"title"`
    Summary     string    `json:"summary"`
    URL         string    `json:"url"`
    Published   time.Time `json:"published"`
    Category    string    `json:"category"`
    Source      string    `json:"source"`
    ReadStatus  bool      `json:"read"`
}
```

### 3.2 News Events (General Events Extension)

```go
// NewsEventConfig extends GeneralDialogEvent for news-specific scenarios
type NewsEventConfig struct {
    GeneralDialogEvent                 // Embed existing event structure
    NewsCategory       string          `json:"newsCategory"`       // "headlines", "tech", "gaming"
    MaxNews           int             `json:"maxNews"`            // Number of articles to mention
    IncludeSummary    bool            `json:"includeSummary"`     // Include article summaries
    ReadingStyle      string          `json:"readingStyle"`       // "casual", "formal", "enthusiastic"
    TriggerFrequency  string          `json:"triggerFrequency"`   // "daily", "hourly", "manual"
}
```

### 3.3 Character Card Extensions

```json
{
  "newsFeatures": {
    "enabled": true,
    "updateInterval": 30,
    "maxStoredItems": 50,
    "readingPersonality": "casual",
    "preferredCategories": ["tech", "gaming"],
    "feeds": [
      {
        "url": "https://feeds.feedburner.com/TechCrunch",
        "name": "TechCrunch",
        "category": "tech",
        "enabled": true
      }
    ],
    "readingEvents": [
      {
        "name": "morning_news",
        "trigger": "daily_news",
        "newsCategory": "headlines",
        "maxNews": 3,
        "responses": [
          "Good morning! Here's what's happening today: {NEWS_SUMMARY}",
          "I've been reading the news! {NEWS_HEADLINES}"
        ]
      }
    ]
  }
}
```

### 3.4 UI Components (Reuse Existing)

```go
// Extend existing DialogBubble for news display
type NewsDialog struct {
    *DialogBubble                // Reuse existing dialog bubble
    newsItems    []NewsItem      // Current news items
    currentIndex int             // Current news item being displayed
}

// Extend existing context menu for news actions
type NewsContextMenu struct {
    *ContextMenu                 // Reuse existing context menu
    newsManager  *NewsManager    // Reference to news system
}
```

## 4. Modification Points

### 4.1 Character Card Schema (`internal/character/card.go`)

**Addition (Lines ~60-65):**
```go
// News feature extensions (RSS/Atom integration)
NewsFeatures *NewsConfig `json:"newsFeatures,omitempty"`
```

**New Structure:**
```go
// NewsConfig defines RSS/Atom newsfeed configuration
type NewsConfig struct {
    Enabled              bool           `json:"enabled"`
    UpdateInterval       int            `json:"updateInterval"`       // Minutes
    MaxStoredItems       int            `json:"maxStoredItems"`
    ReadingPersonality   string         `json:"readingPersonality"`   // "casual", "formal", "enthusiastic"
    PreferredCategories  []string       `json:"preferredCategories"`
    Feeds               []RSSFeed       `json:"feeds"`
    ReadingEvents       []NewsEvent     `json:"readingEvents"`
}
```

### 4.2 Dialog Backend Registration (`internal/character/behavior.go`)

**Addition (Lines ~170-175):**
```go
// Register news backend if news features are enabled
if c.card.HasNewsFeatures() {
    newsBackend := dialog.NewNewsBlogBackend()
    c.dialogManager.RegisterBackend("news_blog", newsBackend)
}
```

### 4.3 Context Menu Extension (`internal/ui/context_menu.go`)

**Addition to menu items:**
```go
if character.GetCard().HasNewsFeatures() {
    menu.AddItem("📰 Read News", func() {
        window.HandleNewsReading()
    })
    menu.AddItem("🔄 Update Feeds", func() {
        window.HandleFeedUpdate()
    })
}
```

### 4.4 Window Interface (`internal/ui/window.go`)

**New Methods:**
```go
func (dw *DesktopWindow) HandleNewsReading() {
    if newsDialog := dw.createNewsDialog(); newsDialog != nil {
        newsDialog.Show()
    }
}

func (dw *DesktopWindow) HandleFeedUpdate() {
    // Trigger manual feed update
    response := dw.character.HandleNewsUpdate()
    if response != "" {
        dw.showDialog(response)
    }
}
```

## 5. Schema Extensions

### 5.1 Character Card News Configuration

```json
{
  "newsFeatures": {
    "enabled": true,
    "updateInterval": 30,
    "maxStoredItems": 50,
    "readingPersonality": "casual",
    "preferredCategories": ["tech", "gaming", "general"],
    "feeds": [
      {
        "url": "https://feeds.feedburner.com/TechCrunch",
        "name": "TechCrunch", 
        "category": "tech",
        "updateFreq": 60,
        "maxItems": 10,
        "keywords": ["AI", "startup", "programming"],
        "enabled": true
      },
      {
        "url": "https://www.reddit.com/r/programming/.rss",
        "name": "r/programming",
        "category": "tech",
        "updateFreq": 30,
        "maxItems": 15,
        "enabled": true
      }
    ],
    "readingEvents": [
      {
        "name": "morning_headlines",
        "category": "conversation",
        "trigger": "daily_news",
        "newsCategory": "headlines",
        "maxNews": 3,
        "includeSummary": false,
        "readingStyle": "casual",
        "responses": [
          "Good morning! Here's what caught my attention: {NEWS_HEADLINES}",
          "I've been browsing the news! Check this out: {NEWS_SUMMARY}"
        ],
        "animations": ["talking", "thinking"],
        "cooldown": 3600
      },
      {
        "name": "tech_deep_dive",
        "category": "conversation", 
        "trigger": "tech_news",
        "newsCategory": "tech",
        "maxNews": 2,
        "includeSummary": true,
        "readingStyle": "enthusiastic",
        "responses": [
          "Oh! I found some fascinating tech news: {NEWS_SUMMARY}",
          "You'll love this tech update: {NEWS_HEADLINES}"
        ],
        "animations": ["excited", "talking"],
        "cooldown": 1800
      }
    ]
  }
}
```

### 5.2 Dialog Backend Configuration

```json
{
  "dialogBackend": {
    "enabled": true,
    "defaultBackend": "markov_chain",
    "fallbackChain": ["markov_chain", "news_blog", "simple_random"],
    "backends": {
      "news_blog": {
        "enabled": true,
        "summaryLength": 100,
        "personalityInfluence": true,
        "cacheTimeout": 1800,
        "debugMode": false
      }
    }
  }
}
```

## 6. Implementation Phases

### Phase 1: Core News Infrastructure (Week 1) ✅ COMPLETED
**Goal**: Basic RSS fetching and parsing foundation

**Components**:
- ✅ `internal/news/` package with RSS/Atom parsing
- ✅ Basic `NewsBlogBackend` implementing `DialogBackend` interface
- ✅ News item storage and caching system
- ✅ Character card schema extensions

**Library Selection**:
```
Library: github.com/mmcdole/gofeed
License: MIT License  
Import: "github.com/mmcdole/gofeed"
Why: Mature RSS/Atom parser with 2.4k+ stars, handles all major feed formats
```

**Deliverables**:
- ✅ Working RSS feed parsing with FeedFetcher
- ✅ Basic news item storage with NewsCache (100% test coverage)
- ✅ Character card validation with news config (HasNewsFeatures method)
- ✅ Unit tests for core functionality (13 tests passing)
- ✅ Example character configuration (`assets/characters/news_example/`)

**Implementation Details**:
- Created `internal/news/types.go` with core data structures
- Created `internal/news/fetcher.go` with RSS/Atom parsing using gofeed library
- Created `internal/news/backend.go` implementing DialogBackend interface
- Extended `internal/character/card.go` with NewsFeatures field
- Added comprehensive test suite with 100% test coverage

### Phase 2: Dialog Integration (Week 2)
**Goal**: News summarization through existing dialog system

**Components**:
- News dialog backend with personality integration
- Template-based news response generation
- Integration with existing character dialog system
- News event triggers and cooldowns

**Implementation**:
```go
// Extend existing dialog context for news
type NewsDialogContext struct {
    dialog.DialogContext          // Embed existing context
    NewsItems       []NewsItem    `json:"newsItems"`
    RequestedCategory string      `json:"requestedCategory"`
    MaxItems        int           `json:"maxItems"`
}
```

**Deliverables**:
- Functional news summarization
- Personality-driven news reading styles
- Integration with existing character personalities
- Comprehensive testing

### Phase 3: UI and Events Integration (Week 3)
**Goal**: User-facing news features through existing UI systems

**Components**:
- News reading through general events system
- Context menu integration
- News dialog display using existing bubble system
- Manual and automatic news triggers

**UI Enhancements**:
- Extend existing `DialogBubble` for news display
- Add news options to existing context menu
- Reuse existing keyboard shortcuts for news events
- Integrate with existing stats overlay for news status

**Deliverables**:
- Complete user experience for news features
- Context menu news options
- Keyboard shortcuts for news reading
- Visual feedback for news updates

### Phase 4: Polish and Optimization (Week 4)
**Goal**: Production-ready news features with performance optimization

**Components**:
- Background feed updating with goroutines
- News item deduplication and filtering
- Error handling and graceful degradation
- Performance optimization and caching

**Enhancements**:
- Smart feed update scheduling
- Bandwidth-conscious update policies
- Comprehensive error recovery
- Memory usage optimization

**Deliverables**:
- Production-ready news system
- Comprehensive documentation
- Performance benchmarks
- Example character configurations

## 7. Risk Assessment

### 7.1 Compatibility Risks

**Low Risk**: Character card schema extensions are optional
- **Mitigation**: All existing characters continue working unchanged
- **Validation**: Extensive backward compatibility testing

**Low Risk**: Dialog backend additions don't affect existing backends
- **Mitigation**: News backend is additional, not replacement
- **Validation**: Existing dialog tests continue passing

### 7.2 Performance Considerations

**Medium Risk**: RSS feed fetching could impact responsiveness
- **Mitigation**: Async background updates with caching
- **Implementation**: Use goroutines and configurable update intervals
- **Monitoring**: Performance metrics for feed update time

**Low Risk**: Memory usage from storing news items
- **Mitigation**: Configurable maximum items per feed
- **Implementation**: LRU cache with automatic cleanup
- **Monitoring**: Memory usage tracking in existing profiler

### 7.3 Network Dependencies

**Medium Risk**: Internet connectivity required for news features
- **Mitigation**: Graceful degradation when feeds unavailable
- **Implementation**: Fallback to existing dialog when no news
- **Error Handling**: User-friendly error messages

**Low Risk**: External feed reliability
- **Mitigation**: Multiple feed sources and retry logic
- **Implementation**: Feed health monitoring and automatic disabling
- **Fallback**: Character functions normally without news

### 7.4 Library Dependencies

**Low Risk**: Additional dependency for RSS parsing
- **Selected**: `github.com/mmcdole/gofeed` (MIT license, 2.4k+ stars)
- **Justification**: Mature, well-maintained, permissive license
- **Alternatives**: Go standard library `encoding/xml` (more work, no licensing issues)

### 7.5 User Experience Risks

**Low Risk**: News features are optional and user-controlled
- **Design**: Opt-in through character configuration
- **Controls**: Manual and automatic trigger options
- **Feedback**: Clear indication when news features are active

**Low Risk**: Information overload
- **Mitigation**: Configurable update frequency and item limits
- **Implementation**: Personality-based filtering and summarization
- **Controls**: User can disable or configure news frequency

## 8. Success Criteria

### 8.1 Functional Requirements
- [ ] RSS/Atom feeds parse correctly
- [ ] News items integrate with existing dialog system
- [ ] Character personalities influence news reading style
- [ ] News features are completely optional
- [ ] Existing characters continue working unchanged
- [ ] News events trigger through existing event system

### 8.2 Performance Requirements
- [ ] Feed updates complete within 5 seconds
- [ ] Memory usage increase <50MB for 100 news items
- [ ] No impact on character animation frame rate
- [ ] Graceful degradation without internet connectivity

### 8.3 Integration Requirements
- [ ] Zero breaking changes to existing APIs
- [ ] Reuse existing UI components where possible
- [ ] Follow existing JSON configuration patterns
- [ ] Maintain existing code style and conventions
- [ ] Pass all existing test suites

### 8.4 User Experience Requirements
- [ ] News reading feels natural for character personality
- [ ] Clear visual feedback for news availability
- [ ] Intuitive controls for enabling/disabling news
- [ ] Helpful error messages for configuration issues

This plan provides a comprehensive roadmap for integrating RSS/Atom newsfeed functionality while maintaining the application's architecture principles and ensuring minimal disruption to existing systems.

## 📋 Implementation Status

### ✅ Phase 1: COMPLETED (August 30, 2025)
**RSS/Atom Core Infrastructure**

**What was implemented:**
1. **News Package (`internal/news/`)**: Complete RSS/Atom parsing infrastructure
   - `types.go`: Core data structures (RSSFeed, NewsItem, NewsConfig, NewsCache)
   - `fetcher.go`: RSS/Atom parsing using github.com/mmcdole/gofeed library  
   - `backend.go`: NewsBlogBackend implementing DialogBackend interface
   - Full test coverage with 13 comprehensive unit tests

2. **Character Card Extensions**: Added NewsFeatures field to CharacterCard struct
   - Optional news configuration with zero breaking changes
   - HasNewsFeatures() method for feature detection
   - Backward compatibility maintained (all existing tests pass)

3. **Example Configuration**: Created `assets/characters/news_example/`
   - Complete character card with news features enabled
   - Example RSS feeds (TechCrunch, Hacker News, Reddit Programming)
   - Documentation and setup instructions

**Technical Achievements:**
- ✅ Zero breaking changes to existing codebase
- ✅ 100% test coverage for news package (13/13 tests passing)
- ✅ Proper concurrency safety with mutex protection
- ✅ Library-first approach using mature dependencies
- ✅ Following existing architectural patterns

**Next Phase Ready**: Phase 2 can now begin with solid foundation in place.

### 🚀 Phase 2: Dialog Integration (Next Priority)
**Goal**: News summarization through existing dialog system

The foundation is now ready for Phase 2 implementation. The news backend can be registered with the dialog manager and begin generating personality-driven news responses.
