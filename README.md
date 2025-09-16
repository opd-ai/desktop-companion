# Desktop Companion (DDS)

## ComfyUI GIF Asset Pipeline

The DDS project includes an automated pipeline for generating GIF character animations using a local ComfyUI instance. This system supports all 19 character archetypes and ensures consistent quality and style across assets.

### Usage

1. Start ComfyUI locally (`http://localhost:8188`).
2. Configure generation parameters in `config.json`.
3. Run asset generation:
  ```bash
  gif-generator batch --config batch_config.json --parallel 4
  ```
4. Validate and deploy assets:
  ```bash
  gif-generator validate --path assets/characters/
  gif-generator deploy --source generated/ --target assets/characters/
  ```

See [GIF_PLAN.md](GIF_PLAN.md) for technical details and troubleshooting.

## Android Build Testing

Automated APK integrity testing is provided for CI/CD validation. The script `scripts/apk_integrity/apk_integrity_test.go` checks APK existence, signature, and package name using Android SDK tools (`apksigner`, `aapt`).

### Usage

1. Build your APK (see Fyne mobile docs).
2. Run the test:
  ```bash
  ./scripts/test-android-apk.sh path/to/app.apk com.example.package
  ```
3. The script will fail if the APK is missing, unsigned, or has the wrong package name.

See `scripts/apk_integrity/apk_integrity_test.go` for details.

## Network Connection Recovery

The multiplayer networking system uses a `ConnectionManager` to handle connection lifecycle, error recovery, and reconnection with exponential backoff. All network operations use Go's standard library and interface types for testability.

### Features

- Automatic reconnection on failure (with exponential backoff)
- Explicit error handling for connection loss and timeouts
- Thread-safe connection state
- Unit tests for recovery logic and error scenarios

See `lib/network/connection.go` for implementation details.

## Performance Optimization

The application includes memory optimization features to reduce garbage collection pressure and improve performance. The `lib/performance` package provides object pools using `sync.Pool` for frequently allocated types.

### Memory Pools

- **Character State Pool**: Reuses character state objects during updates
- **Animation Frame Pool**: Optimizes memory usage during animation processing  
- **Network Message Pool**: Reduces allocations for network communication

Usage example:
```go
// Get from pool, use, then return
cs := performance.GetCharacterState()
cs.Health = 100
// ... use character state ...
performance.PutCharacterState(cs) // Must return to prevent leaks
```

See `lib/performance/pool.go` for details and benchmarks.

## Advanced Dialog Context Enhancement

The dialog system includes advanced conversation context tracking for more natural and contextually aware responses. The `ConversationContext` tracks topics, emotional state, and conversation history to generate better responses.

### Features

- **Topic Detection**: Automatically identifies conversation topics (weather, feelings, activities, food, health)
- **Emotional State Tracking**: Monitors valence, arousal, and dominance in conversations
- **Message History**: Maintains recent conversation context for better continuity
- **Context-Aware Responses**: Markov chain backend uses conversation context to enhance response quality

### Integration

The Markov chain backend automatically:
- Updates conversation context with each message
- Enhances responses with detected topics and emotional state
- Provides context summaries for debugging and analysis

Usage example:
```go
// Context is automatically managed by dialog backends
context := dialog.NewConversationContext()
context.AddMessage(ctx, "I feel happy about the sunny weather")

// Get current state
activeTopics := context.GetActiveTopics() // ["weather", "feelings"]
summary := context.GetContextSummary()   // "discussing weather"
emotional := context.EmotionalState      // Positive valence, moderate arousal
```

See `lib/dialog/context.go` for implementation details and `context_test.go` for usage examples.

A lightweight, platform-native virtual desktop pet application built with Go. Features animated GIF characters, transparent overlays, click interactions, and JSON-based configuration.

## ✨ Features

- 🎭 **Animated Characters**: Support for multi-frame GIF animations with proper timing
- 🪟 **Transparent Overlay**: Always-on-top window with system transparency 
- 🖱️ **Interactive**: Click and drag interactions with animated responses
- 🎮 **Tamagotchi Game Features**: Complete virtual pet system with stats, progression, and achievements *(All Phases Complete)*
- 💕 **Dating Simulator Features**: Complete romance system with relationship progression, personality-driven interactions, and memory-based storytelling *(Phase 3 Complete)*
  - **Stats System**: Hunger, happiness, health, energy with time-based degradation
  - **Game Interactions**: Feed, play, pet with stat effects and cooldowns
  - **Progression System**: Age-based evolution with size changes and animation overrides
  - **Achievement System**: Track milestones with stat-based requirements and rewards
  - **Random Events**: Probability-based events affecting character stats
  - **Critical State Handling**: Special animations and responses for low stats
  - **Mood-Based Animation**: Dynamic animation selection based on character's overall mood
- 💕 **Romance Features**: Complete dating simulator mechanics *(Phase 3 Complete)*
  - **Romance Stats**: Affection, trust, intimacy, jealousy with personality-driven interactions
  - **Relationship Progression**: Stranger → Friend → Close Friend → Romantic Interest with progressive unlocking
  - **Personality System**: Sophisticated personality traits affecting all interactions and responses
  - **Romance Events**: Memory-based random events that respond to interaction history and relationship milestones
  - **Advanced Features**: Jealousy mechanics, compatibility analysis, and relationship crisis recovery systems
  - **JSON-Configurable**: Extensive romance behavior customizable through character cards
- 💾 **Persistent State**: JSON-based save/load system with auto-save functionality *(Complete)*
- 📊 **Stats Overlay**: Optional real-time stats display with progress bars *(Complete)*
- 🤖 **AI-Powered Dialog**: Advanced Markov chain text generation with personality-driven responses *(Complete)*
  - **Intelligent Backends**: Configurable dialog systems with multiple AI backends
  - **Personality Integration**: Responses adapt to character traits, mood, and relationship state
  - **Memory System**: Characters learn and reference past interactions
  - **Context Awareness**: Dialog varies based on triggers, relationship level, and character stats
  - **Quality Control**: Multi-layered filtering ensures coherent, character-appropriate responses
- � **RSS/Atom News Integration**: Real-time news feed parsing and dialog integration *(Complete)*
  - **Feed Management**: Multi-source RSS/Atom feed fetching with smart scheduling
  - **News Backend**: AI dialog backend specializing in news-based conversations  
  - **Caching System**: Memory-efficient news item caching with deduplication
  - **Keyword Filtering**: Optional content filtering per feed source
  - **Background Updates**: Automatic feed updates with error recovery
- �💬 **Interactive Chatbot Interface**: Real-time conversation system with AI characters *(Complete)*
  - **Keyboard Integration**: Press 'C' to toggle chatbot interface instantly
  - **Context Menu Access**: Right-click → "Open Chat" for menu-driven access
  - **Multi-line Input**: Advanced text input with send button for natural conversations
  - **Conversation History**: Scrollable chat history with message persistence
  - **Smart Activation**: Only available for characters with AI dialog backend enabled
  - **Seamless Integration**: Embedded in main desktop window with overlay positioning
- 🎯 **General Dialog Events**: Interactive scenarios and conversations *(Phase 4 Complete)*
  - **Interactive Scenarios**: Multi-choice conversations, roleplay, and story events
  - **Event Categories**: Conversation, roleplay, mini-games, and humor scenarios
  - **User-Initiated Events**: Trigger custom scenarios through keyboard shortcuts or menu
  - **Choice Consequences**: User decisions affect character stats and relationship progression
  - **Event Chaining**: Complex scenarios with multiple phases and branching narratives
  - **Backward Compatible**: Seamlessly integrates with existing dialog and game systems
- 🎁 **Gift System**: Interactive item giving and relationship building *(Complete)*
  - **Gift Categories**: Food, toys, accessories, and special items with stat effects
  - **Relationship Impact**: Gifts affect affection, trust, and character mood
  - **Inventory Management**: Track given gifts and character preferences
  - **Gift UI**: Dedicated interface for browsing and giving gifts
  - **Integration**: Works with both single-player and multiplayer modes
- 🌐 **Multiplayer Networking**: Peer-to-peer networking infrastructure *(Phase 1 Complete)*
  - **Peer Discovery**: UDP-based automatic discovery of other DDS instances on local network
  - **Foundation Ready**: Core infrastructure for AI-controlled multiplayer companions
- 🤖 **Bot Framework**: Autonomous AI character behavior system *(Phase 2 Complete)*
  - **Personality-Driven Behavior**: Configurable traits drive all autonomous decisions
  - **Natural Timing**: Human-like delays and variations prevent mechanical behavior
  - **Character Integration**: Seamless integration with existing Character.Update() cycle
  - **Network Coordination**: Bot characters can interact with peers in multiplayer mode
  - **Performance Optimized**: ~81ns per Update() call, suitable for 60 FPS real-time operation
  - **Rate Limiting**: Prevents excessive actions that would feel unnatural
- ⚔️ **Battle System**: Turn-based tactical combat system *(Complete)*
  - **Fair Combat**: Balanced turn-based mechanics with timeout protection
  - **Multiplayer Ready**: Seamless integration with network multiplayer sessions
  - **AI Opponents**: Intelligent AI-driven battle decisions with personality traits
  - **Battle Actions**: Attack, defend, and special abilities with strategic depth
  - **Battle UI**: Complete interface for battle management and visualization
  - **Cryptographic Security**: Ed25519-signed battle messages for cheat prevention
  - **Performance Optimized**: Sub-millisecond action processing for real-time play
- 🎮 **Multiplayer UI**: Complete network overlay interface *(Phase 3 Complete)*
  - **Character Distinction**: Clear visual separation of local (🏠) vs network (🌐) characters
  - **Peer Management**: Real-time peer discovery and connection status
  - **Network Chat**: Integrated chat system for multiplayer communication
  - **Activity Monitoring**: Live status indicators for all characters and peers
  - **Reliable Messaging**: TCP-based message delivery with JSON protocol
  - **Cryptographic Security**: Ed25519 signature verification for message integrity
  - **Character Configuration**: Optional multiplayer settings in character cards
  - **Standard Library**: Zero external dependencies using Go's built-in networking
  - **Foundation Ready**: Core infrastructure for AI-controlled multiplayer companions
- ⚙️ **Configurable**: JSON-based character cards for easy customization
- 🌍 **Platform-Native**: Runs on Windows, macOS, and Linux (requires building on target platform)
- 🪶 **Lightweight**: Efficient resource usage with built-in monitoring

## 🚀 Quick Start

### Prerequisites

- Go 1.24.5 or higher
- C compiler (gcc/clang) for CGO dependencies
- Platform-specific requirements:
  - **Linux**: X11 or Wayland display environment
  - **macOS**: Xcode command line tools
  - **Windows**: TDM-GCC or Visual Studio Build Tools

### Installation

```bash
```bash
# Quick start
git clone https://github.com/opd-ai/desktop-companion
cd desktop-companion

# Install dependencies
go mod download

# Add animation GIF files (see SETUP guide below)
# Run with default character (includes AI-powered dialog)
go run cmd/companion/main.go

# Enable Tamagotchi game features with AI dialog
go run cmd/companion/main.go -game -character assets/characters/default/character_with_game_features.json

# Enable complete romance features
go run cmd/companion/main.go -game -stats -character assets/characters/romance/character.json

# Show stats overlay for game mode
go run cmd/companion/main.go -game -stats -character assets/characters/default/character_with_game_features.json
```

### 🎬 Animation Setup (Required)

Before running, you need to add GIF animation files:

1. **Create GIF files** in `assets/characters/default/animations/`:
   - `idle.gif` - Default character animation
   - `talking.gif` - Speaking animation  
   - `happy.gif` - Happy/excited animation
   - `sad.gif` - Sad/disappointed animation
   - `hungry.gif` - Hungry animation (for game features)
   - `eating.gif` - Eating animation (for game features)

2. **GIF Requirements**:
   - Format: Animated GIF with transparency
   - Size: 64x64 to 256x256 pixels  
   - File size: <1MB each for best performance
   - Frames: 2-10 frames per animation

3. **Quick Test Setup**:
   - Download sample pixel art GIFs from Tenor or Giphy
   - Or create simple test animations using online GIF makers
   - See `assets/characters/default/animations/SETUP.md` for details

### Building from Source

```bash
# Development build
go build -o companion cmd/companion/main.go

# Optimized release build  
go build -ldflags="-s -w" -o companion cmd/companion/main.go

# Native build for current platform
make build  # Builds for current platform only
```

### Building for Android

DDS supports Android devices through Fyne's cross-platform capabilities:

```bash
# Install fyne CLI tool (if not already installed)
go install fyne.io/tools/cmd/fyne@latest

# Build Android APK (debug version)
make android-debug

# Build Android APK (release version, requires Android SDK)
make android-apk

# Install to connected Android device
make android-install-debug
```

**Android Requirements:**
- Java 8+ installed
- Android SDK (optional for basic builds)
- Android NDK (for full functionality)

For detailed Android setup instructions, see [`docs/ANDROID_BUILD_GUIDE.md`](docs/ANDROID_BUILD_GUIDE.md).

### Cross-Platform Release Build

```bash
# Automated cross-platform build (includes Android)
./scripts/cross_platform_build.sh

# Build for specific platforms
make release-linux    # Linux binary
make release-windows  # Windows binary (requires Windows or cross-compilation)
make release-macos    # macOS binary (requires macOS)
```

### Character-Specific Binary Generation

DDS supports generating standalone, zero-configuration binaries for individual characters with embedded assets:

```bash
# List all available characters
make list-characters

# Build character for current platform
make build-character CHAR=default
make build-character CHAR=tsundere
make build-character CHAR=romance_flirty

# Build all characters for current platform
make build-characters

# Validate all character binaries
make validate-characters

# Benchmark character binary performance
make benchmark-characters

# Clean character build artifacts
make clean-characters

# Build character for specific platforms (including Android)
PLATFORMS=linux/amd64,windows/amd64,darwin/amd64,android/arm64 ./scripts/build-characters.sh build default

# Android-specific character builds (requires fyne CLI)
go install fyne.io/tools/cmd/fyne@latest
PLATFORMS=android/arm64 ./scripts/build-characters.sh build default
PLATFORMS=android/arm ./scripts/build-characters.sh build tsundere
```

**Binary Validation and Testing:**
```bash
# Comprehensive binary validation
make validate-characters                    # Validate all built binaries

# Performance benchmarking
make benchmark-characters                   # Generate performance reports

# Custom validation options
./scripts/validate-character-binaries.sh --timeout 60 validate
./scripts/validate-character-binaries.sh benchmark

# Integration testing
go test scripts/pipeline_integration_test.go -v    # Full pipeline tests
go test scripts/validate-character-binaries_test.go -v    # Validation tests
```

**Generated Output:**
```
build/
├── default_linux_amd64           # Linux executable
├── default_windows_amd64.exe     # Windows executable  
├── default_darwin_amd64          # macOS executable
├── default_android_arm64.apk     # Android APK (64-bit)
├── tsundere_android_arm.apk      # Android APK (32-bit)
└── romance_flirty_android_arm64.apk  # Character-specific APK

test_output/
├── validation_default.log        # Individual validation logs
├── benchmark_results.log         # Performance benchmark data
└── ...                          # Additional test artifacts
```

## 🏗️ Architecture & Dependencies

This project follows the "lazy programmer" philosophy, using mature libraries instead of custom implementations:

### Primary Dependencies

| Library | License | Purpose | Why Chosen |
|---------|---------|---------|------------|
| [fyne.io/fyne/v2](https://fyne.io/) | BSD-3-Clause | Cross-platform GUI | Only mature Go GUI with native transparency support |
| [github.com/mmcdole/gofeed](https://github.com/mmcdole/gofeed) | MIT | RSS/Atom feed parsing | Mature RSS parser with 2.4k+ stars, handles multiple feed formats |
| [github.com/jdkato/prose/v2](https://github.com/jdkato/prose) | MIT | Natural language processing | Advanced text analysis for dialog context enhancement |
| [github.com/opd-ai/minilm](https://github.com/opd-ai/minilm) | MIT | Sentence embedding | Lightweight sentence similarity for dialog matching |
| [github.com/sirupsen/logrus](https://github.com/sirupsen/logrus) | MIT | Structured logging | Production-grade logging with customizable output formats |
| [nhooyr.io/websocket](https://nhooyr.io/websocket) | MIT | WebSocket client | ComfyUI real-time communication for GIF generation pipeline |
| Go standard library | BSD-3-Clause | JSON parsing, GIF decoding, image processing, networking | Zero external dependencies, battle-tested |

*Note: Full dependency list available in `go.mod` - all dependencies use permissive licenses compatible with commercial use.*

### License Compliance

All dependencies use permissive licenses (BSD-3-Clause) that allow commercial use without attribution requirements. No license files need to be bundled with binaries, but this project includes `LICENSES.md` for transparency.

## 📖 Usage

### Basic Usage

1. **Launch**: Run the executable or `go run cmd/companion/main.go`
2. **Interact**: Click on the character to trigger dialog responses
3. **Move**: Drag the character around your desktop (if enabled)
4. **Configure**: Edit `assets/characters/default/character.json` to customize behavior

### Game Mode Usage

Enable Tamagotchi-style game features with the `-game` flag:

```bash
# Basic game mode
go run cmd/companion/main.go -game -character assets/characters/default/character_with_game_features.json

# Game mode with stats overlay
go run cmd/companion/main.go -game -stats -character assets/characters/default/character_with_game_features.json

# Choose difficulty level
go run cmd/companion/main.go -game -stats -character assets/characters/easy/character.json      # Beginner
go run cmd/companion/main.go -game -stats -character assets/characters/normal/character.json    # Normal
go run cmd/companion/main.go -game -stats -character assets/characters/hard/character.json      # Hard
go run cmd/companion/main.go -game -stats -character assets/characters/challenge/character.json # Expert

# Experience romance features
go run cmd/companion/main.go -game -stats -character assets/characters/romance/character.json   # Romance simulator

# Romance Character Archetypes (Phase 4 Complete!)
go run cmd/companion/main.go -game -stats -character assets/characters/tsundere/character.json   # Shy, defensive, slow-burn
go run cmd/companion/main.go -game -stats -character assets/characters/flirty/character.json     # Outgoing, playful, fast-paced  
go run cmd/companion/main.go -game -stats -character assets/characters/slow_burn/character.json  # Thoughtful, realistic, long-term

# AI-Powered Dialog Examples
go run cmd/companion/main.go -character assets/characters/markov_example/character.json         # Basic Markov dialog
go run cmd/companion/main.go -character assets/characters/examples/markov_dialog_example.json  # Advanced dialog system

# RSS/Atom News Integration Examples
go run cmd/companion/main.go -character assets/characters/news_example/character.json          # News-enabled character

# Interactive Chatbot Interface Examples  
go run cmd/companion/main.go -character assets/characters/markov_example/character.json         # AI chat with basic character
# Press 'C' key or right-click → "Open Chat" to start AI conversations
# Type messages and receive AI-generated responses based on character personality

# General Dialog Events Examples
go run cmd/companion/main.go -character assets/characters/examples/interactive_events.json     # Interactive conversations
go run cmd/companion/main.go -character assets/characters/examples/roleplay_character.json    # Roleplay scenarios

# Multiplayer Networking Examples (New in Phase 3!)
go run cmd/companion/main.go -network -character assets/characters/multiplayer/social_bot.json        # Social bot with networking
go run cmd/companion/main.go -network -network-ui -character assets/characters/multiplayer/helper_bot.json  # Helper bot with UI overlay
go run cmd/companion/main.go -network -network-ui -character assets/characters/default/character.json  # Regular character in network mode
# Press 'N' key or right-click → "Network Overlay" to toggle network UI (shows local 🏠 vs network 🌐 characters)

# Battle System Examples (Complete!)
go run cmd/companion/main.go -network -character assets/characters/multiplayer/social_bot.json        # Enable battle-capable character
# Battle invitations available through context menu in multiplayer mode
# Turn-based combat with AI opponents and strategic decision making
```

**General Dialog Events**:
- **Trigger Events**: Use keyboard shortcuts (Ctrl+E, Ctrl+R, Ctrl+G) to initiate scenarios
- **Interactive Choices**: Click on choice buttons during events to make decisions
- **Event Categories**: 
  - **Conversation**: Daily check-ins, deep discussions, life advice
  - **Roleplay**: Fantasy adventures, detective mysteries, sci-fi scenarios
  - **Games**: Trivia questions, word games, creative challenges
  - **Humor**: Jokes, puns, funny stories, and silly interactions
- **Progress Tracking**: Events affect relationship stats and unlock new scenarios
- **Event Memory**: Characters remember your choices and reference them later

**Game Interactions**:
- **Click**: Pet your character (increases happiness and health)
- **Right-click**: Feed your character (increases hunger with dialog response)
- **Double-click**: Play with your character (increases happiness, decreases energy)
- **Stats overlay**: Toggle with 'S' key to monitor character's wellbeing
- **Chatbot interface**: Toggle with 'C' key for AI-powered conversations (AI characters only)
- **Context menu**: Right-click for advanced options including "Open Chat" for AI characters, "Give Gift" for gift system, and "Battle Invite" for multiplayer combat
- **Network overlay**: Toggle with 'N' key to show multiplayer status (network mode only)
- **Gift giving**: Access gift interface through context menu to give items and build relationships
- **Battle system**: Initiate turn-based combat through multiplayer context menu options
- **Auto-save**: Game state automatically saves at intervals that vary by difficulty:
  - Easy: 10 minutes (600 seconds)
  - Normal/Romance: 5 minutes (300 seconds)  
  - Specialist: 10 minutes (600 seconds)
  - Hard: 2 minutes (120 seconds)
  - Challenge: 1 minute (60 seconds)

**Multiplayer Interactions** (network mode):
- **Network overlay**: Shows local (🏠) vs network (🌐) characters with activity status
- **Activity feed**: Real-time scrollable log of all network peer actions and events
- **Peer chat**: Send messages to other players through the network overlay
- **Character visibility**: See all characters connected to the network session
- **Real-time sync**: Character actions and status updates shared across peers
- **Battle system**: Challenge other players to turn-based combat matches
- **Gift exchange**: Share gifts between characters in multiplayer sessions

**Character Care**:
- **Monitor Stats**: Hunger, happiness, health, and energy decrease over time
- **Critical States**: Characters show special animations when stats are low
- **Progression**: Characters evolve and grow as they age
- **Achievements**: Unlock rewards by maintaining good care over time
- **Random Events**: Unexpected events can affect your character's stats

### Romance Character Archetypes

The dating simulator includes multiple distinct character personalities and variants, each offering a unique romantic experience:

| Archetype | Difficulty | Progression | Best For |
|-----------|------------|-------------|----------|
| **Tsundere** | Hard | Slow (8+ days) | Patient players who enjoy character development |
| **Flirty Extrovert** | Easy | Fast (4+ days) | Players wanting immediate gratification |
| **Slow Burn** | Expert | Very Slow (16+ days) | Long-term commitment and realistic pacing |

**Romance Variants Available:**
- `romance/` - Balanced romance character
- `romance_flirty/` - Romance with flirty personality traits
- `romance_slowburn/` - Romance with slow-burn characteristics  
- `romance_supportive/` - Romance with supportive personality
- `romance_tsundere/` - Romance with tsundere elements

See `CHARACTER_ARCHETYPES.md` for detailed personality profiles, strategy guides, and customization options.

### Character Cards

Characters are defined using JSON configuration files with this structure:

```json
{
  "name": "My Desktop Pet",
  "description": "A friendly companion for your desktop",
  "animations": {
    "idle": "animations/idle.gif",
    "talking": "animations/talking.gif",
    "happy": "animations/happy.gif"
  },
  "dialogs": [
    {
      "trigger": "click",
      "responses": [
        "Hello there!",
        "How can I help you today?",
        "Nice to see you again!"
      ],
      "animation": "talking",
      "cooldown": 5
    }
  ],
  "behavior": {
    "idleTimeout": 30,
    "movementEnabled": true,
    "defaultSize": 128
  },
  "dialogBackend": {
    "enabled": true,
    "defaultBackend": "markov_chain",
    "confidenceThreshold": 0.6,
    "backends": {
      "markov_chain": {
        "chainOrder": 2,
        "minWords": 3,
        "maxWords": 12,
        "temperatureMin": 0.4,
        "temperatureMax": 0.7,
        "usePersonality": true,
        "trainingData": [
          "Character-specific training phrases here",
          "Include personality-appropriate language",
          "Mix different emotional tones"
        ]
      }
    }
  }
}
```

### Configuration Schema

#### Required Fields

- `name` (string, 1-50 chars): Character display name
- `description` (string, 1-200 chars): Character description
- `animations` (object): Animation file mappings
  - `idle` (string, required): Default animation GIF path
  - `talking` (string, required): Speaking animation GIF path
  - Additional animations (optional): `happy`, `sad`, `excited`, etc.

#### Dialog Configuration

- `trigger` (string): `click`, `rightclick`, or `hover`
- `responses` (array): 1-10 response text strings
- `animation` (string): Animation to play (must exist in `animations`)
- `cooldown` (number): Seconds between dialog triggers (default: 5)

#### Behavior Settings

- `idleTimeout` (number, 10-300): Seconds before returning to idle animation
- `movementEnabled` (boolean): Allow dragging the character (default: false)
- `defaultSize` (number, 64-512): Character size in pixels (uses 128 when value is 0 or negative)

#### Multiplayer Configuration (Optional)

Characters can be configured for peer-to-peer networking and multiplayer features:

```json
{
  "multiplayer": {
    "enabled": true,
    "botCapable": true,
    "networkID": "my_character_v1",
    "maxPeers": 8,
    "discoveryPort": 8080
  }
}
```

**Multiplayer Fields:**
- `enabled` (boolean): Enable multiplayer networking features
- `botCapable` (boolean): Allow this character to run autonomously as a bot
- `networkID` (string, required if enabled): Unique identifier for this character type (alphanumeric, underscore, dash only)
- `maxPeers` (number, 0-16): Maximum number of peers to connect to (default: 8)
- `discoveryPort` (number, 1024-65535): UDP port for peer discovery (default: 8080)

**Security Notes:**
- All network messages are cryptographically signed with Ed25519 (authenticated, but not encrypted; messages are readable by network intermediaries)
- Only characters with matching `networkID` can communicate
- Discovery ports below 1024 are restricted to avoid system conflicts

### General Dialog Events

Characters support rich interactive scenarios through the general events system:

```json
{
  "generalEvents": [
    {
      "name": "daily_check_in",
      "category": "conversation",
      "description": "A daily conversation about how things are going",
      "trigger": "daily_check_in",
      "probability": 1.0,
      "interactive": true,
      "responses": [
        "How has your day been going? I'd love to hear about it! 😊"
      ],
      "animations": ["talking"],
      "choices": [
        {
          "text": "It's been great!",
          "effects": {"happiness": 5, "affection": 2},
          "nextEvent": "celebrate_good_day"
        },
        {
          "text": "Pretty challenging...",
          "effects": {"trust": 3},
          "nextEvent": "supportive_conversation"
        },
        {
          "text": "Just the usual.",
          "effects": {"affection": 1}
        }
      ],
      "cooldown": 86400,
      "conditions": {"affection": {"min": 10}}
    }
  ]
}
```

#### General Event Properties

- **`name`** (string): Unique identifier for the event
- **`category`** (string): Event type - "conversation", "roleplay", "game", "humor"
- **`trigger`** (string): How to initiate the event (keyboard shortcut or auto-trigger)
- **`interactive`** (boolean): Whether the event supports user choices
- **`choices`** (array): User interaction options with stat effects and follow-ups
- **`followUpEvents`** (array): Events that can chain after this one
- **`cooldown`** (number): Seconds before event can trigger again
- **`conditions`** (object): Stat requirements to access the event

#### Dialog Backend Configuration (Optional)

**Conversation Events**: Daily check-ins, advice sessions, life discussions
**Roleplay Events**: Fantasy adventures, mystery scenarios, sci-fi explorations  
**Game Events**: Trivia questions, word games, creative challenges
**Humor Events**: Joke sessions, pun competitions, funny stories

- `dialogBackend.enabled` (boolean): Enable AI-powered dialog generation
- `dialogBackend.defaultBackend` (string): Primary backend to use ("markov_chain", "simple_random")
- `dialogBackend.confidenceThreshold` (number, 0-1): Minimum confidence for generated responses
- `dialogBackend.backends` (object): Backend-specific configuration
  - `markov_chain.chainOrder` (number, 1-5): Complexity of text generation (2 recommended)
  - `markov_chain.temperatureMin/Max` (number, 0-2): Randomness range for responses
  - `markov_chain.usePersonality` (boolean): Enable personality-driven generation
  - `markov_chain.trainingData` (array): Character-specific training phrases

#### Game Features (Complete Implementation)

Character cards can include comprehensive Tamagotchi-style game features:

```json
{
  "stats": {
    "hunger": {
      "initial": 100,
      "max": 100,
      "degradationRate": 1.0,
      "criticalThreshold": 20
    },
    "happiness": {
      "initial": 100,
      "max": 100,
      "degradationRate": 0.8,
      "criticalThreshold": 15
    },
    "health": {
      "initial": 100,
      "max": 100,
      "degradationRate": 0.3,
      "criticalThreshold": 10
    },
    "energy": {
      "initial": 100,
      "max": 100,
      "degradationRate": 1.5,
      "criticalThreshold": 25
    }
  },
  "gameRules": {
    "statsDecayInterval": 60,
    "autoSaveInterval": 300,
    "criticalStateAnimationPriority": true,
    "deathEnabled": false,
    "evolutionEnabled": true,
    "moodBasedAnimations": true
  },
  "interactions": {
    "feed": {
      "triggers": ["rightclick"],
      "effects": {"hunger": 25, "happiness": 5},
      "animations": ["eating", "happy"],
      "responses": ["Yum! Thank you!", "That was delicious!", "I feel much better now!"],
      "cooldown": 30,
      "requirements": {"hunger": {"max": 80}}
    },
    "play": {
      "triggers": ["doubleclick"],
      "effects": {"happiness": 20, "energy": -15},
      "animations": ["happy"],
      "responses": ["This is fun!", "I love playing with you!", "Let's play more!"],
      "cooldown": 45,
      "requirements": {"energy": {"min": 20}}
    },
    "pet": {
      "triggers": ["click"],
      "effects": {"happiness": 10, "health": 2},
      "animations": ["happy"],
      "responses": ["That feels nice!", "I love attention!", "Pet me more!"],
      "cooldown": 15
    }
  },
  "progression": {
    "levels": [
      {
        "name": "Baby",
        "requirement": {"age": 0},
        "size": 64,
        "animations": {}
      },
      {
        "name": "Child",
        "requirement": {"age": 86400},
        "size": 96,
        "animations": {}
      },
      {
        "name": "Adult",
        "requirement": {"age": 259200},
        "size": 128,
        "animations": {}
      }
    ],
    "achievements": [
      {
        "name": "Well Fed",
        "requirement": {
          "hunger": {"maintainAbove": 80},
          "maintainAbove": {"duration": 86400}
        },
        "reward": {
          "statBoosts": {"hunger": 10}
        }
      },
      {
        "name": "Happy Pet",
        "requirement": {
          "happiness": {"maintainAbove": 90},
          "maintainAbove": {"duration": 43200}
        },
        "reward": {
          "statBoosts": {"happiness": 10}
        }
      }
    ]
  }
}
```

**Game Feature Configuration**:

- **Stats System**: Define character stats (hunger, happiness, health, energy) with individual degradation rates and critical thresholds
- **Game Rules**: Configure game mechanics including decay intervals, auto-save frequency, and feature toggles
- **Interactions**: Define game interactions (feed, play, pet) with stat effects, requirements, cooldowns, and animations
- **Progression System**: Age-based evolution with size changes and animation overrides
- **Achievement System**: Track milestones with complex stat-based requirements and reward stat boosts
- **Random Events**: Probability-based events that can positively or negatively affect character stats
- **Critical State Handling**: Special animations and responses when stats drop below thresholds
- **Mood-Based Animation Selection**: Dynamic idle animation selection based on overall character mood

**Available Difficulty Levels**:
- **Easy** (`assets/characters/easy/`): Slower stat degradation, easier requirements
- **Normal** (`assets/characters/normal/`): Balanced gameplay for regular users
- **Hard** (`assets/characters/hard/`): Faster stat degradation, more challenging requirements
- **Challenge** (`assets/characters/challenge/`): Extreme difficulty for expert players
- **Specialist** (`assets/characters/specialist/`): Unique gameplay mechanics and requirements

**📚 Documentation Suite:**
- **`SCHEMA_DOCUMENTATION.md`**: Complete JSON schema reference with all properties and validation rules
- **`ROMANCE_SCENARIOS.md`**: Example romance progression scenarios and strategies  
- **`CHARACTER_ARCHETYPES.md`**: Detailed comparison of the three romance archetypes
- **`DIALOG_BACKEND_GUIDE.md`**: Complete guide to AI-powered dialog configuration
- **`MARKOV_DIALOG_CONFIGURATION_GUIDE.md`**: Detailed Markov chain setup and customization
- **`GENERAL_EVENTS_GUIDE.md`**: Comprehensive guide to interactive dialog events and scenarios

### Creating Custom Characters

1. **Create character directory**:
   ```bash
   mkdir -p assets/characters/mycharacter/animations
   ```

2. **Add GIF animations**:
   - Ensure GIFs have transparency for best results
   - Recommended size: 64x64 to 256x256 pixels
   - Keep file sizes under 1MB each for performance

3. **Create character.json**:
   ```bash
   cp assets/characters/default/character.json assets/characters/mycharacter/
   # Edit the configuration file
   ```

4. **Load custom character**:
   ```bash
   go run cmd/companion/main.go -character assets/characters/mycharacter/character.json
   ```

## 🎮 Command-Line Options

The companion supports various command-line flags for different modes and configurations:

```bash
# Basic options
go run cmd/companion/main.go [options]

-character <path>     Path to character configuration file (default: "assets/characters/default/character.json")
-debug               Enable debug logging for troubleshooting
-version             Show version information

# Game features (Tamagotchi mode)
-game                Enable Tamagotchi game features (stats, interactions, progression)
-stats               Show real-time stats overlay (requires -game)

# Multiplayer networking features (New in Phase 3!)
-network             Enable multiplayer networking features
-network-ui          Show network overlay UI (requires -network)

# General dialog events system
-events               Enable general dialog events system for interactive scenarios
-trigger-event <name> Manually trigger a specific event by name

# Performance profiling
-memprofile <file>   Write memory profile to file for analysis
-cpuprofile <file>   Write CPU profile to file for analysis
```

**Example Usage**:
```bash
# Standard desktop pet
go run cmd/companion/main.go

# Game mode with easy difficulty
go run cmd/companion/main.go -game -stats -character assets/characters/easy/character.json

# Expert challenge mode
go run cmd/companion/main.go -game -stats -character assets/characters/challenge/character.json

# Debug mode with performance profiling
go run cmd/companion/main.go -debug -memprofile=mem.prof -cpuprofile=cpu.prof

# Interactive dialog events examples  
go run cmd/companion/main.go -events -character assets/characters/examples/interactive_events.json
go run cmd/companion/main.go -events -trigger-event daily_check_in -character assets/characters/examples/interactive_events.json

# Example usage with events
go run cmd/companion/main.go -game -stats -events -character assets/characters/examples/interactive_events.json
```

**General Event Interactions**:
- **Ctrl+E**: Open events menu to see available scenarios
- **Ctrl+R**: Quick-start a random roleplay scenario
- **Ctrl+G**: Start a mini-game or trivia session
- **Ctrl+H**: Trigger a humor/joke session
- **During Events**: Click choice buttons to make decisions and progress the story

## 🛠️ Development

### Project Structure

```
DDS/
├── cmd/companion/main.go           # Application entry point
├── lib/
│   ├── character/
│   │   ├── card.go                 # JSON configuration parser (stdlib)
│   │   ├── animation.go            # GIF animation manager (stdlib)
│   │   ├── behavior.go             # Main character implementation and logic
│   │   ├── game_state.go           # Tamagotchi-style game state management
│   │   ├── progression.go          # Age-based progression and achievements
│   │   ├── random_events.go        # Probability-based random events system
│   │   ├── compatibility.go        # Advanced compatibility analysis
│   │   ├── crisis_recovery.go      # Relationship crisis management
│   │   ├── jealousy.go             # Jealousy mechanics
│   │   ├── gift_definition.go      # Gift system and item management
│   │   ├── gift_manager.go         # Gift giving mechanics
│   │   ├── general_events.go       # Interactive dialog events system
│   │   ├── multiplayer.go          # Multiplayer character support
│   │   ├── multiplayer_battle.go   # Battle system character integration
│   │   ├── network_events.go       # Network-based character events
│   │   └── *_test.go               # Comprehensive unit tests (45+ files)
│   ├── ui/
│   │   ├── window.go              # Transparent window (fyne)
│   │   ├── renderer.go            # Character rendering
│   │   ├── interaction.go         # Dialog bubbles (fyne)
│   │   ├── stats_overlay.go       # Real-time stats display
│   │   ├── chatbot_interface.go   # AI chatbot interface
│   │   ├── context_menu.go        # Right-click context menu
│   │   ├── network_overlay.go     # Multiplayer network UI
│   │   ├── gift_dialog.go         # Gift giving interface
│   │   ├── battle_ui.go           # Battle system interface
│   │   └── *_test.go              # UI component tests
│   ├── config/
│   │   └── loader.go              # Configuration file loading
│   ├── dialog/                    # AI-powered dialog system
│   │   ├── interface.go           # Dialog backend interface
│   │   ├── markov_backend.go      # Markov chain text generation
│   │   └── simple_random_backend.go # Simple random response backend
│   ├── persistence/               # Game state persistence
│   │   ├── save_manager.go        # JSON-based save/load system
│   │   └── save_manager_test.go   # Comprehensive persistence tests
│   ├── monitoring/                # Performance monitoring
│   │   ├── profiler.go            # Performance profiling and metrics
│   │   └── profiler_test.go       # Performance testing
│   ├── battle/                    # Turn-based battle system
│   │   ├── manager.go             # Battle state management and coordination
│   │   ├── actions.go             # Battle action processing and validation
│   │   ├── fairness.go            # Fairness constraint enforcement
│   │   ├── ai.go                  # AI battle decision making
│   │   └── *_test.go              # Battle system tests
│   ├── bot/                       # Autonomous AI character behavior
│   │   ├── controller.go          # Bot behavior coordination
│   │   ├── actions.go             # Autonomous action processing
│   │   ├── personality.go         # Personality-driven AI decisions
│   │   └── *_test.go              # Bot system tests
│   ├── network/                   # Multiplayer networking infrastructure
│   │   ├── manager.go             # Network session management
│   │   ├── protocol.go            # Network protocol and message handling
│   │   ├── sync.go                # State synchronization between peers
│   │   ├── group_events.go        # Multiplayer group event coordination
│   │   └── *_test.go              # Network system tests
│   └── testing/                   # Shared testing utilities
│       └── helpers.go             # Common test infrastructure
├── assets/characters/             # Character configurations
│   ├── default/                   # Basic character without game features
│   ├── easy/                      # Easy difficulty game character
│   ├── normal/                    # Normal difficulty game character
│   ├── hard/                      # Hard difficulty game character
│   ├── challenge/                 # Expert difficulty game character
│   ├── specialist/                # Unique gameplay mechanics
│   ├── romance/                   # Balanced romance character
│   ├── tsundere/                  # Tsundere romance archetype
│   ├── flirty/                    # Flirty extrovert archetype
│   ├── slow_burn/                 # Slow burn romance archetype
│   ├── romance_flirty/            # Romance with flirty traits
│   ├── romance_slowburn/          # Romance with slow burn traits
│   ├── romance_supportive/        # Romance with supportive traits
│   ├── romance_tsundere/          # Romance with tsundere traits
│   ├── markov_example/            # AI dialog demonstration
│   ├── multiplayer/               # Multiplayer networking characters
│   ├── news_example/              # RSS/Atom news integration
│   ├── examples/                  # Example configurations
│   └── templates/                 # Character creation templates
├── Makefile                       # Build automation
├── CHARACTER_ARCHETYPES.md        # Romance archetype comparison guide
├── SCHEMA_DOCUMENTATION.md        # Complete JSON schema reference
├── ROMANCE_SCENARIOS.md           # Example romance progression scenarios
└── LICENSES.md                    # License information
```

### Design Principles

1. **Library-First**: Use existing solutions instead of custom implementations
2. **Standard Library Preference**: Leverage Go's stdlib (json, image/gif) when possible
3. **Minimal Custom Code**: Write only glue code and business logic
4. **Interface-Based**: Use standard Go patterns for testability
5. **Proper Concurrency**: Mutex protection for all shared state

### Running Tests

```bash
# Run all tests with coverage
go test ./... -v -cover

# Run specific package tests
go test ./lib/character -v

# Run with race detection
go test ./... -race

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Performance Monitoring

The application includes built-in performance monitoring and profiling:

```bash
# Run with memory profiling
go run cmd/companion/main.go -memprofile=mem.prof

# Run with CPU profiling  
go run cmd/companion/main.go -cpuprofile=cpu.prof

# Run with both profiles and debug output
go run cmd/companion/main.go -memprofile=mem.prof -cpuprofile=cpu.prof -debug

# Analyze profiles
go tool pprof mem.prof
go tool pprof cpu.prof
```

**Real-time Monitoring**:
- Memory usage tracking and reporting
- Frame rate monitoring with performance warnings
- Startup time measurement
- Concurrent frame rendering support
- Built-in performance profiling capabilities

## 🔨 Building and Distribution

### Development Builds

```bash
# Local development
go run cmd/companion/main.go

# With custom character
go run cmd/companion/main.go -character path/to/character.json

# With debug logging
go run cmd/companion/main.go -debug
```

### Release Builds

```bash
# Native build for current platform
go build -ldflags="-s -w" -o companion cmd/companion/main.go

# Using Makefile
make build  # Creates build/companion
```

> **Note**: Due to Fyne GUI framework limitations, cross-compilation is not supported.  
> Fyne requires platform-specific CGO libraries for graphics drivers.  
> Build on the target platform for proper binary distribution.

## 🔧 Troubleshooting

### Common Issues

**"failed to initialize display" (Linux)**:
```bash
# Install required X11 libraries
sudo apt-get install libx11-dev libxcursor-dev libxrandr-dev libxinerama-dev libxi-dev libgl1-mesa-dev
```

**"character.json not found"**:
- Ensure the character directory contains a valid `character.json` file
- Check file paths are relative to the character.json location
- Verify all referenced GIF files exist

**Poor animation performance**:
- Reduce GIF file sizes (optimize with tools like `gifsicle`)
- Lower GIF frame rates to 10-15 FPS
- Ensure character size is reasonable (64-256 pixels)

**Window not staying on top (Linux)**:
- Some window managers don't support always-on-top hints
- Try different desktop environments (GNOME, KDE, XFCE)
- Check window manager documentation for overlay support

**Game features not working**:
- Ensure you're using the `-game` flag to enable Tamagotchi features
- Use character cards with game features (e.g., `character_with_game_features.json`)
- Check that the character card has `stats`, `interactions`, and `gameRules` sections

**Stats overlay not visible**:
- Ensure both `-game` and `-stats` flags are enabled
- Try pressing the 'S' key to toggle stats overlay
- Verify the character card has valid stats configuration

**Save game not loading**:
- Check permissions on save directory (`~/.local/share/desktop-companion/`)
- Ensure character name matches between save file and character card
- Try starting fresh if save file is corrupted (will auto-regenerate)

### Debug Mode

Enable debug logging for troubleshooting:

```bash
go run cmd/companion/main.go -debug
```

This provides detailed output about:
- Character card loading and validation
- Animation file processing
- Window creation and positioning
- Performance metrics and memory usage
- Game state updates and stat changes (in game mode)
- Save/load operations and persistence
- Achievement progress and random events (in game mode)

For additional help, see the debug output above or check `AUDIT.md` for known issues and resolutions.

## 🤝 Contributing

1. **Fork** the repository
2. **Create** a feature branch: `git checkout -b feature/amazing-feature`
3. **Follow** the code standards in the project documentation
4. **Add** tests for new functionality
5. **Ensure** all tests pass: `go test ./...`
6. **Commit** changes: `git commit -m 'Add amazing feature'`
7. **Push** to branch: `git push origin feature/amazing-feature`
8. **Create** a Pull Request

### Code Standards

- Maximum 30 lines per function
- Use standard library when possible
- Implement proper error handling (no ignored errors)
- Add mutex protection for shared state
- Include unit tests for business logic (target 70% coverage)

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

### Third-Party Licenses

- **Fyne**: BSD-3-Clause License
- **Go Standard Library**: BSD-3-Clause License

See [LICENSES.md](LICENSES.md) for complete license information.

## 🙏 Acknowledgments

- **Fyne Project**: For providing excellent cross-platform GUI capabilities
- **Go Team**: For the robust standard library that handles GIF decoding and JSON parsing
- **Desktop Pet Community**: For inspiration and character art examples

---

**Minimum System Requirements**:
- 512MB RAM available
- 50MB disk space (100MB recommended with save files)
- OpenGL 2.1 or higher (for hardware acceleration)
- X11 (Linux), Cocoa (macOS), or Win32 (Windows) display server

**Recommended for Game Mode**:
- 1GB RAM for smooth performance
- 100MB disk space for save files and multiple character cards
- SSD storage for faster auto-save operations

**Note**: This application requires GIF animation files to run. See the setup instructions above for details on adding animations. Game features require character cards with `stats`, `interactions`, and `gameRules` configuration sections.
