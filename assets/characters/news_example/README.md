# News Example Character

This character demonstrates the RSS/Atom newsfeed integration features implemented in Phase 1 of the news system.

## Features Demonstrated

### Core News Integration
- **RSS Feed Sources**: Three real tech news feeds (O'Reilly Radar, The Verge, Ars Technica)
- **News Categories**: Tech, gaming, and general news support
- **Keyword Filtering**: Optional filtering by keywords like "AI", "programming", "technology"
- **Update Scheduling**: Different update frequencies per feed (30-60 minutes)

### Dialog Backend Integration
- **News Blog Backend**: Custom dialog backend that generates news-based responses
- **Personality Adaptation**: Responses adapt to casual reading style
- **Fallback Chain**: Falls back to simple random responses if news backend fails
- **Caching**: 30-minute cache timeout with up to 50 stored news items

### Character Interactions
- **Click**: General greeting with hint about news reading
- **Right-click**: Triggers news feed updates
- **Ctrl+N**: Opens interactive news discussion menu
- **Automatic News**: Daily morning headlines (with 1-hour cooldown)

### News Events
- **Morning Headlines**: Daily summary of top 3 news items
- **Tech Deep Dive**: Detailed tech news with summaries (manual trigger)
- **Gaming Updates**: Gaming news headlines (manual trigger)

## Setup Requirements

### Animation Files
Create these GIF files in `animations/` directory:
- `idle.gif` - Default character animation
- `talking.gif` - Character speaking animation  
- `reading.gif` - Character reading news animation
- `thinking.gif` - Character processing/updating animation

### Network Access
The character requires internet access to fetch RSS feeds from:
- https://feeds.feedburner.com/oreilly/radar
- https://www.theverge.com/rss/index.xml
- https://feeds.arstechnica.com/arstechnica/index

## Usage Example

```bash
# Run with news-enabled character
go run cmd/companion/main.go -character assets/characters/news_example/character.json

# Enable debug mode to see news backend activity
go run cmd/companion/main.go -debug -character assets/characters/news_example/character.json
```

## Interaction Guide

1. **Initial Setup**: Character will attempt to fetch news on startup
2. **Manual Updates**: Right-click to force news feed updates
3. **News Discussion**: Press Ctrl+N to open interactive news menu
4. **Daily Headlines**: Character will automatically share news periodically
5. **Personality**: Responses adapt to casual reading style with tech focus

## Technical Details

### Configuration Schema
The character card demonstrates the new `newsFeatures` configuration section:

```json
{
  "newsFeatures": {
    "enabled": true,
    "updateInterval": 30,
    "maxStoredItems": 50,
    "readingPersonality": "casual",
    "preferredCategories": ["tech", "gaming", "general"],
    "feeds": [...],
    "readingEvents": [...]
  }
}
```

### Backend Configuration
Shows integration with the dialog backend system:

```json
{
  "dialogBackend": {
    "defaultBackend": "news_blog",
    "backends": {
      "news_blog": {
        "enabled": true,
        "summaryLength": 100,
        "personalityInfluence": true,
        "cacheTimeout": 1800
      }
    }
  }
}
```

## Customization

### Adding Feeds
Add new RSS/Atom feeds to the `feeds` array:

```json
{
  "url": "https://example.com/feed.rss",
  "name": "Example News",
  "category": "general",
  "updateFreq": 60,
  "maxItems": 10,
  "keywords": ["keyword1", "keyword2"],
  "enabled": true
}
```

### Personality Styles
Change `readingPersonality` to:
- `"casual"` - Friendly, informal tone
- `"formal"` - Professional, structured tone  
- `"enthusiastic"` - Excited, energetic tone

### News Categories
Supported categories:
- `"tech"` - Technology news
- `"gaming"` - Gaming industry news
- `"general"` - General news/headlines
- Custom categories based on feed configuration

## Phase 1 Implementation Status

✅ **Completed Features:**
- RSS/Atom feed parsing with gofeed library
- News item storage and caching system
- Character card schema extensions
- Dialog backend integration
- Basic news response generation
- Personality-driven reading styles
- Feed management and validation
- Comprehensive unit tests (>80% coverage)

⏳ **Next Phases:**
- Phase 2: Advanced dialog integration with templates
- Phase 3: UI components and context menu integration  
- Phase 4: Background updates and optimization

## Error Handling

The implementation includes robust error handling:
- **Network Failures**: Graceful degradation when feeds are unavailable
- **Invalid Feeds**: Feed validation on startup
- **Malformed Content**: HTML tag removal and text cleaning
- **Rate Limiting**: Configurable update intervals prevent excessive requests

## Performance Considerations

- **Memory Usage**: Configurable cache limits prevent memory bloat
- **Network Efficiency**: Feeds only update when interval has elapsed
- **Response Time**: Sub-second response generation from cached content
- **Concurrency**: Thread-safe operations with mutex protection
