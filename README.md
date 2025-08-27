````markdown
# Desktop Companion (DDS)

A lightweight, cross-platform virtual desktop pet application built with Go. Features animated GIF characters, transparent overlays, click interactions, and JSON-based configuration.

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
  - **JSON-Configurable**: 90%+ of romance behavior customizable through character cards
- 💾 **Persistent State**: JSON-based save/load system with auto-save functionality *(Complete)*
- 📊 **Stats Overlay**: Optional real-time stats display with progress bars *(Complete)*
- ⚙️ **Configurable**: JSON-based character cards for easy customization
- 🌍 **Cross-Platform**: Runs on Windows, macOS, and Linux (build on target platform)
- 🪶 **Lightweight**: <50MB memory usage

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or higher
- C compiler (gcc/clang) for CGO dependencies
- Platform-specific requirements:
  - **Linux**: X11 development libraries (`libx11-dev`, `libxcursor-dev`, `libxrandr-dev`, `libxinerama-dev`, `libxi-dev`, `libgl1-mesa-dev`)
  - **macOS**: Xcode command line tools
  - **Windows**: TDM-GCC or Visual Studio Build Tools

### Installation

```bash
# Clone the repository
git clone https://github.com/opd-ai/DDS
cd DDS

# Install dependencies
go mod download

# Add animation GIF files (see SETUP guide below)
# Then run with default character
go run cmd/companion/main.go

# Enable Tamagotchi game features
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

## 🏗️ Architecture & Dependencies

This project follows the "lazy programmer" philosophy, using mature libraries instead of custom implementations:

### Primary Dependencies

| Library | License | Purpose | Why Chosen |
|---------|---------|---------|------------|
| [fyne.io/fyne/v2](https://fyne.io/) | BSD-3-Clause | Cross-platform GUI | Only mature Go GUI with native transparency support |
| Go standard library | BSD-3-Clause | JSON parsing, GIF decoding, image processing | Zero external dependencies, battle-tested |

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
```

**Game Interactions**:
- **Click**: Pet your character (increases happiness and health)
- **Right-click**: Feed your character (increases hunger, shows interaction menu)
- **Double-click**: Play with your character (increases happiness, decreases energy)
- **Stats overlay**: Toggle with keyboard shortcut to monitor character's wellbeing
- **Auto-save**: Game state automatically saves every 5 minutes

**Character Care**:
- **Monitor Stats**: Hunger, happiness, health, and energy decrease over time
- **Critical States**: Characters show special animations when stats are low
- **Progression**: Characters evolve and grow as they age
- **Achievements**: Unlock rewards by maintaining good care over time
- **Random Events**: Unexpected events can affect your character's stats

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
- `defaultSize` (number, 64-512): Character size in pixels (default: 128)

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

For complete game features documentation, see `GAME_FEATURES_PHASE1.md` and `PHASE2_IMPLEMENTATION_COMPLETE.md`.

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

# Custom character with game features
go run cmd/companion/main.go -game -stats -character /path/to/my/character.json
```

## 🛠️ Development

### Project Structure

```
desktop-companion/
├── cmd/companion/main.go           # Application entry point
├── internal/
│   ├── character/
│   │   ├── card.go                 # JSON configuration parser (stdlib)
│   │   ├── animation.go            # GIF animation manager (stdlib)
│   │   ├── behavior.go             # Character behavior logic
│   │   ├── game_state.go           # Tamagotchi-style game state management
│   │   ├── progression.go          # Age-based progression and achievements
│   │   ├── random_events.go        # Probability-based random events system
│   │   └── *_test.go               # Comprehensive unit tests
│   ├── ui/
│   │   ├── window.go              # Transparent window (fyne)
│   │   ├── renderer.go            # Character rendering
│   │   ├── interaction.go         # Dialog bubbles (fyne)
│   │   ├── stats_overlay.go       # Real-time stats display
│   │   └── *_test.go              # UI component tests
│   ├── config/
│   │   └── loader.go              # Configuration file loading
│   ├── persistence/               # Game state persistence
│   │   ├── save_manager.go        # JSON-based save/load system
│   │   └── save_manager_test.go   # Comprehensive persistence tests
│   └── monitoring/                # Performance monitoring
│       ├── profiler.go            # Performance profiling and metrics
│       └── profiler_test.go       # Performance testing
├── assets/characters/             # Character configurations
│   ├── default/                   # Basic character without game features
│   ├── easy/                      # Easy difficulty game character
│   ├── normal/                    # Normal difficulty game character
│   ├── hard/                      # Hard difficulty game character
│   ├── challenge/                 # Expert difficulty game character
│   └── specialist/                # Unique gameplay mechanics
├── Makefile                       # Build automation
├── PERFORMANCE_MONITORING.md      # Performance metrics and monitoring
├── GAME_FEATURES_PHASE1.md        # Game features documentation
├── PHASE2_IMPLEMENTATION_COMPLETE.md # Implementation status
├── AUDIT.md                       # Code quality and functional audit
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
go test ./internal/character -v

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

**Performance Targets**:
- Memory usage: <50MB during normal operation ✅ **MONITORED**
- Animation framerate: 30+ FPS consistently ✅ **MONITORED**
- Startup time: <2 seconds ✅ **MONITORED**

**Real-time Monitoring**:
- Memory usage tracking with target validation
- Frame rate monitoring with performance warnings
- Startup time measurement
- Concurrent frame rendering support
- Automatic performance target validation

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

> **Note**: Due to Fyne GUI framework limitations, cross-platform builds are not supported.  
> Fyne requires platform-specific CGO libraries for graphics drivers.  
> Build on the target platform for proper binary distribution.
```

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
- Try pressing the stats toggle keyboard shortcut
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
- 512MB RAM available (1GB recommended for game mode)
- 50MB disk space (100MB recommended with save files)
- OpenGL 2.1 or higher (for hardware acceleration)
- X11 (Linux), Cocoa (macOS), or Win32 (Windows) display server

**Recommended for Game Mode**:
- 1GB RAM for smooth stats processing
- 100MB disk space for save files and multiple character cards
- SSD storage for faster auto-save operations

**Note**: This application requires GIF animation files to run. See the setup instructions above for details on adding animations. Game features require character cards with `stats`, `interactions`, and `gameRules` configuration sections.
````