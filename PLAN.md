# PROJECT: Cross-Platform Virtual Desktop Companion

## OBJECTIVE:
Create a lightweight desktop pet application in Go that displays an interactive animated character with transparent background, configurable through JSON files, supporting click interactions and animated responses across Windows, macOS, and Linux platforms.

## TECHNICAL SPECIFICATIONS:
- Language: Go 1.21+
- Type: Cross-platform desktop application
- Key Features: 
  - Transparent window overlay with GIF animations
  - Click/drag interactions with dialog responses
  - JSON-based character configuration
  - Always-on-top positioning
- Performance Requirements: 
  - <50MB memory usage
  - 30+ FPS animation
  - <2 second startup time
  - <10MB binary size per platform

## ARCHITECTURE GUIDELINES:

### Preferred Libraries:
| Library | Use Case | Justification |
|---------|----------|---------------|
| fyne.io/fyne/v2 | Cross-platform GUI | Mature, Go-native, transparency support |
| gopkg.in/yaml.v3 | Config parsing | Standard library alternative for YAML |
| encoding/json | JSON config (primary) | Standard library, zero dependencies |
| image/gif | GIF decoding | Standard library, built-in support |
| golang.org/x/image | Image processing | Official extended library |

LIBRARY SELECTION PROCESS:
1. Prioritize standard library (encoding/json, image/gif)
2. Use fyne.io/fyne/v2 for GUI (most mature Go GUI with transparency)
3. Avoid custom windowing implementations - leverage fyne's platform abstraction
4. Document why each library was chosen in README

### Project Structure:
```
desktop-companion/
├── cmd/
│   └── companion/
│       └── main.go              # Application entry point
├── internal/
│   ├── character/
│   │   ├── card.go             # Character configuration loader
│   │   ├── animation.go        # GIF animation manager
│   │   └── behavior.go         # Character behavior logic
│   ├── ui/
│   │   ├── window.go           # Transparent window management
│   │   ├── renderer.go         # Character rendering
│   │   └── interaction.go      # Mouse event handling
│   └── config/
│       └── loader.go           # Configuration file loading
├── assets/
│   ├── characters/
│   │   └── default/
│   │       ├── character.json  # Default character card
│   │       └── animations/     # GIF files
├── build/
│   └── scripts/               # Cross-platform build scripts
├── go.mod
├── go.sum
└── README.md
```

### Design Patterns:
- **Observer Pattern**: For animation state changes and user interactions
- **Strategy Pattern**: For platform-specific window management
- **Factory Pattern**: For creating character instances from configs
- **Singleton Pattern**: For application state management

## IMPLEMENTATION PHASES:

### Phase 1: Foundation (Days 1-2)
**Tasks:**
- [ ] Set up Go module with fyne dependency
- [ ] Create basic transparent window with fyne
- [ ] Implement character card JSON schema and parser
- [ ] Load and display static GIF image

**Acceptance Criteria:**
- Window appears with transparency on all platforms
- JSON config loads successfully with validation
- Single GIF displays correctly in window
- Error handling for missing files and invalid configs

### Phase 2: Core Features (Days 3-5)
**Tasks:**
- [ ] Implement GIF animation cycling system
- [ ] Add mouse click detection and window dragging
- [ ] Create dialog bubble system with text rendering
- [ ] Implement animation state machine (idle → talking → idle)
- [ ] Add always-on-top window positioning

**Acceptance Criteria:**
- Smooth GIF animation transitions at 30+ FPS
- Click interactions trigger appropriate animations and dialogs
- Window maintains desktop overlay behavior
- Dialog bubbles appear/disappear with timing controls

### Phase 3: Testing & Documentation (Days 6-7)
**Tasks:**
- [ ] Create unit tests for character card parsing
- [ ] Add integration tests for animation system
- [ ] Write comprehensive README with setup instructions
- [ ] Create example character cards
- [ ] Implement cross-platform build scripts

**Acceptance Criteria:**
- 80% test coverage on business logic functions
- All example character cards load successfully
- Build scripts produce working binaries for all platforms
- Memory profiling confirms <50MB usage

## CODE STANDARDS:

SIMPLICITY RULES:
- If a solution requires more than 3 levels of abstraction, redesign it
- Prefer explicit code over implicit magic
- Use dependency injection over global state
- Choose boring technology over cutting-edge when possible

### Good vs Bad Examples:

❌ AVOID - Complex animation state management:
```go
func (c *Character) UpdateState(newState string) {
    if c.stateTransitions[c.currentState][newState] && 
       time.Since(c.lastTransition) > c.cooldowns[newState] {
        c.performTransition(newState)
    }
}
```

✅ PREFER - Explicit state transitions:
```go
// Update character animation state with validation
func (c *Character) UpdateState(newState string) error {
    if !c.canTransitionTo(newState) {
        return fmt.Errorf("invalid transition from %s to %s", c.currentState, newState)
    }
    
    if time.Since(c.lastTransition) < c.getStateCooldown(newState) {
        return fmt.Errorf("state change too soon, need %v cooldown", c.getStateCooldown(newState))
    }
    
    return c.setAnimationState(newState)
}
```

### Function Complexity Limits:
- Maximum 30 lines per function (excluding comments)
- Maximum 3 parameters per function
- Maximum 10 cyclomatic complexity
- No nested functions beyond 2 levels

### Error Handling Requirements:
```go
// All file operations must handle errors explicitly
func loadCharacterCard(path string) (*CharacterCard, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read character card %s: %w", path, err)
    }
    
    var card CharacterCard
    if err := json.Unmarshal(data, &card); err != nil {
        return nil, fmt.Errorf("failed to parse character card %s: %w", path, err)
    }
    
    if err := card.Validate(); err != nil {
        return nil, fmt.Errorf("invalid character card %s: %w", path, err)
    }
    
    return &card, nil
}
```

## CHARACTER CARD SCHEMA:

### Required JSON Structure:
```json
{
  "name": "string (required, 1-50 chars)",
  "description": "string (required, 1-200 chars)",
  "animations": {
    "idle": "string (required, path to GIF)",
    "talking": "string (required, path to GIF)",
    "happy": "string (optional, path to GIF)",
    "sad": "string (optional, path to GIF)"
  },
  "dialogs": [
    {
      "trigger": "click|rightclick|hover",
      "responses": ["string array, 1-10 items"],
      "animation": "string (must match animations key)",
      "cooldown": "number (seconds, default 5)"
    }
  ],
  "behavior": {
    "idleTimeout": "number (seconds, 10-300)",
    "movementEnabled": "boolean (default false)",
    "defaultSize": "number (pixels, 64-512)"
  }
}
```

### Validation Rules:
- All file paths must be relative to character card directory
- Animation files must be valid GIF format with transparency
- Dialog responses must be non-empty strings
- Behavior values must be within specified ranges

## VALIDATION CHECKLIST:
- [ ] All functions under 30 lines
- [ ] 80%+ test coverage on business logic
- [ ] No custom windowing when fyne provides functionality
- [ ] All file operations handle errors explicitly
- [ ] Character card validation prevents runtime crashes
- [ ] Memory profiling confirms <50MB usage
- [ ] Cross-platform builds work on Windows/macOS/Linux
- [ ] Animation performance maintains 30+ FPS
- [ ] Documentation covers installation and character creation
- [ ] Example character cards demonstrate all features

## BUILD INSTRUCTIONS:

### Development Setup:
```bash
# Install dependencies
go mod download

# Run locally (requires X11 on Linux)
go run cmd/companion/main.go

# Run tests
go test ./... -v -cover
```

### Cross-Platform Compilation:
```bash
# Windows
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o build/companion-windows.exe cmd/companion/main.go

# macOS 
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o build/companion-macos cmd/companion/main.go

# Linux
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o build/companion-linux cmd/companion/main.go
```

MAINTAINABILITY REQUIREMENTS:
- All code must be readable without extensive context
- Complex algorithms require step-by-step comments
- Configuration must be centralized and documented
- Dependencies must be explicitly versioned in go.mod
- Include migration guides for character card format changes

## PERFORMANCE MONITORING:
Include these benchmarks in testing:
- Memory usage during 1-hour runtime
- Animation frame timing consistency
- GIF loading performance for various file sizes
- Window rendering performance on different screen resolutions