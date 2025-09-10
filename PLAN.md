# Character-Specific Binary Generation Plan

## Overview

This plan outlines the implementation of standalone, zero-configuration binary executables for each individual pet/companion character in the Desktop Dating Simulator (DDS) application. The solution leverages automated GitHub Actions workflows to generate character-specific binaries with embedded assets.

## 1. Codebase Analysis Summary

### Current Architecture

**Core Structure:**
- **Main Entry Point**: `cmd/companion/main.go` - Uses flag-based configuration with character path loading
- **Character System**: `lib/character/` - Modular character loading via JSON cards and GIF animations
- **Asset Loading**: Runtime asset resolution via `resolveProjectRoot()` and `LoadCard()` functions
- **JSON Schema**: Comprehensive character cards with 19+ archetypes supporting game mechanics, romance systems, AI dialogs, and multiplayer features

**Key Interfaces and Structs:**
```go
type CharacterCard struct {
    Name        string            `json:"name"`
    Description string            `json:"description"`
    Animations  map[string]string `json:"animations"` // Relative paths to GIF files
    Dialogs     []Dialog          `json:"dialogs"`
    Behavior    Behavior          `json:"behavior"`
    // ... extensive configuration options
}

type Character struct {
    card             *CharacterCard
    animationManager *AnimationManager
    basePath         string
    // ... state management and features
}
```

**Animation System:**
- Uses Go's standard `image/gif` package via `AnimationManager`
- Loads GIF files from filesystem at runtime using `os.Open()`
- Supports frame-by-frame playback with proper timing

**Configuration Flow:**
1. Command-line flag parsing (`-character` path)
2. Asset path resolution (development vs deployed binary detection)
3. Character card loading via `character.LoadCard()`
4. Animation loading from relative paths in character directory

**Available Characters:**
- 19+ character archetypes including: default, easy, normal, hard, challenge, specialist, romance variants (tsundere, flirty, slow_burn, supportive), multiplayer bots, and specialized examples

## 2. GitHub Actions Workflow Design

### `.github/workflows/build-character-binaries.yml`

```yaml
name: Build Character-Specific Binaries

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

env:
  GO_VERSION: '1.21'
  
jobs:
  generate-matrix:
    runs-on: ubuntu-latest
    outputs:
      characters: ${{ steps.characters.outputs.matrix }}
    steps:
      - uses: actions/checkout@v4
      - name: Generate character matrix
        id: characters
        run: |
          # Extract character names from directory structure
          CHARS=$(find assets/characters -name "character.json" -exec dirname {} \; | \
                  grep -v examples | grep -v templates | \
                  xargs -I {} basename {} | \
                  jq -R -s -c 'split("\n")[:-1]')
          echo "matrix=$CHARS" >> $GITHUB_OUTPUT

  build-binaries:
    needs: generate-matrix
    strategy:
      matrix:
        character: ${{ fromJson(needs.generate-matrix.outputs.characters) }}
        os: [ubuntu-latest, windows-latest, macos-latest]
        include:
          - os: ubuntu-latest
            goos: linux
            goarch: amd64
            ext: ""
          - os: windows-latest
            goos: windows
            goarch: amd64
            ext: ".exe"
          - os: macos-latest
            goos: darwin
            goarch: amd64
            ext: ""
    
    runs-on: ${{ matrix.os }}
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: ${{ env.GO_VERSION }}
      
      - name: Install dependencies
        run: |
          go mod download
          go mod tidy
      
      - name: Install platform dependencies
        shell: bash
        run: |
          if [[ "${{ matrix.os }}" == "ubuntu-latest" ]]; then
            sudo apt-get update
            sudo apt-get install -y gcc pkg-config libgl1-mesa-dev libxi-dev libxcursor-dev libxrandr-dev libxinerama-dev
          elif [[ "${{ matrix.os }}" == "macos-latest" ]]; then
            # macOS has necessary dependencies built-in
            echo "macOS dependencies ready"
          elif [[ "${{ matrix.os }}" == "windows-latest" ]]; then
            # Windows Go includes necessary CGO support
            echo "Windows dependencies ready"
          fi
      
      - name: Generate embedded character
        run: |
          go run scripts/embed-character.go \
            -character ${{ matrix.character }} \
            -output cmd/companion-${{ matrix.character }}
      
      - name: Build character binary
        env:
          CGO_ENABLED: 1
          GOOS: ${{ matrix.goos }}
          GOARCH: ${{ matrix.goarch }}
        run: |
          go build -ldflags="-s -w" \
            -o build/${{ matrix.character }}_${{ matrix.goos }}_${{ matrix.goarch }}${{ matrix.ext }} \
            ./cmd/companion-${{ matrix.character }}
      
      - name: Upload artifacts
        uses: actions/upload-artifact@v4
        with:
          name: ${{ matrix.character }}-${{ matrix.goos }}-${{ matrix.goarch }}
          path: build/${{ matrix.character }}_${{ matrix.goos }}_${{ matrix.goarch }}${{ matrix.ext }}
          retention-days: 30

  package-releases:
    needs: [generate-matrix, build-binaries]
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Download all artifacts
        uses: actions/download-artifact@v4
        with:
          path: artifacts/
      
      - name: Create release packages
        run: |
          mkdir -p releases/
          for char in $(echo '${{ needs.generate-matrix.outputs.characters }}' | jq -r '.[]'); do
            mkdir -p releases/$char/
            cp artifacts/$char-*/* releases/$char/
            # Create platform-specific archives
            for os in linux windows darwin; do
              if [[ "$os" == "windows" ]]; then
                ext=".exe"
              else
                ext=""
              fi
              if [[ -f "releases/$char/${char}_${os}_amd64${ext}" ]]; then
                cd releases/$char/
                if [[ "$os" == "windows" ]]; then
                  zip -r "../${char}_${os}_amd64.zip" "${char}_${os}_amd64${ext}"
                else
                  tar -czf "../${char}_${os}_amd64.tar.gz" "${char}_${os}_amd64${ext}"
                fi
                cd ../../
              fi
            done
          done
      
      - name: Upload release packages
        uses: actions/upload-artifact@v4
        with:
          name: character-releases
          path: releases/*.{tar.gz,zip}
          retention-days: 90
```

## 3. Required Code Modifications

### 3.1 Asset Embedding System

**Create `scripts/embed-character.go`:**

```go
//go:build ignore

package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "os"
    "path/filepath"
    "text/template"
)

var (
    characterName = flag.String("character", "", "Character name to embed")
    outputDir     = flag.String("output", "", "Output directory for generated code")
)

const mainTemplate = `package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "image/gif"
    "log"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/app"
    
    "github.com/opd-ai/desktop-companion/lib/character"
    "github.com/opd-ai/desktop-companion/lib/monitoring"
    "github.com/opd-ai/desktop-companion/lib/ui"
)

// Embedded character data
var embeddedCharacterData = ` + "`" + `{{.CharacterJSON}}` + "`" + `

// Embedded animations
var embeddedAnimations = map[string][]byte{
{{range $name, $data := .Animations}}    "{{$name}}": {
{{range $data}}        {{.}},
{{end}}    },
{{end}}}

const appVersion = "1.0.0-{{.CharacterName}}"

func main() {
    // Parse embedded character data
    var card character.CharacterCard
    if err := json.Unmarshal([]byte(embeddedCharacterData), &card); err != nil {
        log.Fatalf("Failed to parse embedded character data: %v", err)
    }

    // Create embedded animation manager
    animManager, err := createEmbeddedAnimationManager()
    if err != nil {
        log.Fatalf("Failed to create animation manager: %v", err)
    }

    // Initialize profiler
    profiler := monitoring.NewProfiler(50)
    if err := profiler.Start("", "", false); err != nil {
        log.Fatalf("Failed to start profiler: %v", err)
    }
    defer profiler.Stop("", false)

    // Create Fyne application
    myApp := app.NewWithID("com.opdai.{{.CharacterName}}-companion")
    
    // Create character with embedded assets
    char, err := character.NewEmbedded(&card, animManager)
    if err != nil {
        log.Fatalf("Failed to create character: %v", err)
    }

    // Create and show UI
    window := ui.NewCompanionWindow(myApp, char, false, false, false, false)
    window.ShowAndRun()
}

func createEmbeddedAnimationManager() (*character.AnimationManager, error) {
    animManager := character.NewAnimationManager()
    
    for name, data := range embeddedAnimations {
        reader := bytes.NewReader(data)
        gifData, err := gif.DecodeAll(reader)
        if err != nil {
            return nil, fmt.Errorf("failed to decode embedded animation %s: %w", name, err)
        }
        
        if err := animManager.LoadEmbeddedAnimation(name, gifData); err != nil {
            return nil, fmt.Errorf("failed to load embedded animation %s: %w", name, err)
        }
    }
    
    return animManager, nil
}
`

type TemplateData struct {
    CharacterName string
    CharacterJSON string
    Animations    map[string][]string
}

func main() {
    flag.Parse()
    
    if *characterName == "" || *outputDir == "" {
        log.Fatal("Both -character and -output flags are required")
    }
    
    // Load character card
    cardPath := fmt.Sprintf("assets/characters/%s/character.json", *characterName)
    cardData, err := os.ReadFile(cardPath)
    if err != nil {
        log.Fatalf("Failed to read character card: %v", err)
    }
    
    // Parse character card to get animation paths
    var card map[string]interface{}
    if err := json.Unmarshal(cardData, &card); err != nil {
        log.Fatalf("Failed to parse character card: %v", err)
    }
    
    // Load and embed animations
    animations := make(map[string][]string)
    characterDir := filepath.Dir(cardPath)
    
    if animsInterface, ok := card["animations"]; ok {
        if anims, ok := animsInterface.(map[string]interface{}); ok {
            for animName, animPathInterface := range anims {
                if animPath, ok := animPathInterface.(string); ok {
                    fullPath := filepath.Join(characterDir, animPath)
                    animData, err := os.ReadFile(fullPath)
                    if err != nil {
                        log.Printf("Warning: Failed to read animation %s: %v", animName, err)
                        continue
                    }
                    
                    // Convert bytes to Go code representation
                    var byteStrings []string
                    for _, b := range animData {
                        byteStrings = append(byteStrings, fmt.Sprintf("0x%02x", b))
                    }
                    animations[animName] = byteStrings
                }
            }
        }
    }
    
    // Create output directory
    if err := os.MkdirAll(*outputDir, 0755); err != nil {
        log.Fatalf("Failed to create output directory: %v", err)
    }
    
    // Generate main.go
    tmpl := template.Must(template.New("main").Parse(mainTemplate))
    
    data := TemplateData{
        CharacterName: *characterName,
        CharacterJSON: string(cardData),
        Animations:    animations,
    }
    
    outputFile := filepath.Join(*outputDir, "main.go")
    file, err := os.Create(outputFile)
    if err != nil {
        log.Fatalf("Failed to create output file: %v", err)
    }
    defer file.Close()
    
    if err := tmpl.Execute(file, data); err != nil {
        log.Fatalf("Failed to execute template: %v", err)
    }
    
    fmt.Printf("Generated %s for character %s\n", outputFile, *characterName)
}
```

### 3.2 Character Package Extensions

**Extend `lib/character/animation.go`:**

Add the following method to support embedded animations:

```go
// LoadEmbeddedAnimation loads a pre-decoded GIF animation into the manager
func (am *AnimationManager) LoadEmbeddedAnimation(name string, gifData *gif.GIF) error {
    am.mu.Lock()
    defer am.mu.Unlock()

    if len(gifData.Image) == 0 {
        return fmt.Errorf("embedded GIF animation %s contains no frames", name)
    }

    am.animations[name] = gifData

    // Set as current animation if this is the first one loaded
    if am.currentAnim == "" {
        am.currentAnim = name
    }

    return nil
}
```

**Extend `lib/character/behavior.go`:**

Add constructor for embedded character creation:

```go
// NewEmbedded creates a character instance with embedded assets (no filesystem dependencies)
func NewEmbedded(card *CharacterCard, animManager *AnimationManager) (*Character, error) {
    char := &Character{
        card:                     card,
        animationManager:        animManager,
        basePath:                "", // No filesystem base path needed
        currentState:            AnimationIdle,
        lastStateChange:         time.Now(),
        lastInteraction:         time.Now(),
        dialogCooldowns:         make(map[string]time.Time),
        gameInteractionCooldowns: make(map[string]time.Time),
        romanceEventCooldowns:   make(map[string]time.Time),
        idleTimeout:             time.Duration(card.Behavior.IdleTimeout) * time.Second,
        movementEnabled:         card.Behavior.MovementEnabled,
        size:                    card.Behavior.DefaultSize,
    }

    if err := initializeCharacterSystems(char); err != nil {
        return nil, err
    }

    return char, nil
}
```

## 4. Build Automation Scripts

### 4.1 Character Enumeration Script

**Create `scripts/list-characters.sh`:**

```bash
#!/bin/bash

# List all available character directories (excluding examples and templates)
find assets/characters -maxdepth 1 -type d \
    -not -path "assets/characters" \
    -not -path "*/examples" \
    -not -path "*/templates" \
    -exec basename {} \; | \
    sort
```

### 4.2 Parallel Build Script

**Create `scripts/build-all-characters.sh`:**

```bash
#!/bin/bash

set -e

PLATFORMS=("linux/amd64" "windows/amd64" "darwin/amd64")
BUILD_DIR="build"
MAX_PARALLEL=4

# Create build directory
mkdir -p "$BUILD_DIR"

# Get character list
CHARACTERS=($(scripts/list-characters.sh))

echo "Building ${#CHARACTERS[@]} characters for ${#PLATFORMS[@]} platforms..."

# Function to build a single character for a platform
build_character() {
    local char="$1"
    local platform="$2"
    local goos="${platform%/*}"
    local goarch="${platform#*/}"
    local ext=""
    
    if [[ "$goos" == "windows" ]]; then
        ext=".exe"
    fi
    
    echo "Building $char for $platform..."
    
    # Generate embedded character code
    go run scripts/embed-character.go \
        -character "$char" \
        -output "cmd/companion-$char"
    
    # Build binary
    CGO_ENABLED=1 GOOS="$goos" GOARCH="$goarch" \
        go build -ldflags="-s -w" \
        -o "$BUILD_DIR/${char}_${goos}_${goarch}${ext}" \
        "./cmd/companion-$char"
    
    # Cleanup generated code
    rm -rf "cmd/companion-$char"
    
    echo "✓ Built $char for $platform"
}

# Export function for parallel execution
export -f build_character
export BUILD_DIR

# Build all characters in parallel
for char in "${CHARACTERS[@]}"; do
    for platform in "${PLATFORMS[@]}"; do
        echo "$char $platform"
    done
done | xargs -n 2 -P "$MAX_PARALLEL" bash -c 'build_character "$@"' _

echo "Build complete! Binaries available in $BUILD_DIR/"
```

## 5. Artifact Management Strategy

### 5.1 Storage and Naming Conventions

**Binary Naming Pattern:**
```
{CharacterName}_{OS}_{Architecture}{Extension}

Examples:
- default_linux_amd64
- tsundere_windows_amd64.exe  
- romance_darwin_amd64
```

**Artifact Organization:**
```
artifacts/
├── character-releases/
│   ├── default_linux_amd64.tar.gz
│   ├── default_windows_amd64.zip
│   ├── default_darwin_amd64.tar.gz
│   ├── tsundere_linux_amd64.tar.gz
│   └── ...
└── individual/
    ├── default-linux-amd64/
    ├── default-windows-amd64/
    └── ...
```

### 5.2 Retention Policies

**GitHub Actions Artifacts:**
- **Individual binaries**: 30 days retention
- **Release packages**: 90 days retention
- **Development builds**: 7 days retention (PR builds)

### 5.3 Enhanced Makefile Integration

**Add to existing `Makefile`:**

```makefile
# Character-specific binary generation
.PHONY: build-characters list-characters clean-characters

# List available characters
list-characters:
	@scripts/list-characters.sh

# Build all character binaries
build-characters: $(BUILD_DIR)
	@echo "Building character-specific binaries..."
	@scripts/build-all-characters.sh

# Build single character for current platform  
build-character: $(BUILD_DIR)
	@if [ -z "$(CHAR)" ]; then echo "Usage: make build-character CHAR=character_name"; exit 1; fi
	@go run scripts/embed-character.go -character $(CHAR) -output cmd/companion-$(CHAR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(CHAR)_$(shell go env GOOS)_$(shell go env GOARCH) cmd/companion-$(CHAR)/main.go
	@rm -rf cmd/companion-$(CHAR)
	@echo "✓ Built $(CHAR) for $(shell go env GOOS)/$(shell go env GOARCH)"

# Clean character build artifacts
clean-characters:
	rm -rf cmd/companion-*
	rm -f $(BUILD_DIR)/*_*_*

# Help for character builds
help-characters:
	@echo "Character-specific build targets:"
	@echo "  list-characters    - List all available character archetypes"
	@echo "  build-characters   - Build all characters for all platforms"
	@echo "  build-character    - Build single character (specify CHAR=name)"
	@echo "  clean-characters   - Remove character build artifacts"
	@echo ""
	@echo "Examples:"
	@echo "  make build-character CHAR=default"
	@echo "  make build-character CHAR=tsundere"
```

## 6. Implementation Benefits

### 6.1 Zero-Configuration Distribution
- **No external dependencies**: All assets embedded at build time
- **Single-file distribution**: Each character becomes a standalone executable
- **Cross-platform compatibility**: Native builds for Windows, macOS, and Linux
- **Simplified deployment**: Users download one file and run immediately

### 6.2 Automated CI/CD Pipeline  
- **Matrix builds**: Parallel compilation across platforms and characters
- **Asset validation**: Build-time verification of character cards and animations
- **Artifact management**: Organized storage with appropriate retention policies
- **Quality assurance**: Automated testing of embedded asset loading

### 6.3 Developer Experience
- **Minimal code changes**: Leverages existing character system architecture
- **Library-first approach**: Uses Go's standard `image/gif` package and byte embedding
- **Backward compatibility**: Original character loading system remains unchanged
- **Build automation**: Complete pipeline automation with parallel execution

## 7. Implementation Timeline

### Phase 1: Core Infrastructure (Week 1) ✅ COMPLETE
1. ✅ **COMPLETED**: Create asset embedding script (`scripts/embed-character.go`)
2. ✅ **COMPLETED**: Extend character package with embedded asset support
3. ✅ **COMPLETED**: Create build automation scripts (`scripts/build-characters.sh`)
4. ✅ **COMPLETED**: Test local builds for multiple characters (default, flirty validated)

### Phase 2: CI/CD Pipeline (Week 2)
1. ✅ **COMPLETED**: Implement GitHub Actions workflow (`build-character-binaries.yml`)
2. ✅ **COMPLETED**: Configure matrix builds for all platforms (Linux, Windows, macOS + Apple Silicon)
3. ✅ **COMPLETED**: Set up artifact management and retention
4. ✅ **COMPLETED**: Test full pipeline with multiple characters

### Phase 3: Integration and Testing (Week 3)
1. ✅ **COMPLETED**: Integrate with existing Makefile
2. ✅ **COMPLETED**: Validate all character binaries (comprehensive validation system with performance benchmarks)
3. Performance testing and optimization
4. Documentation and user guides

### Phase 4: Release and Monitoring (Week 4)
1. Deploy to production CI/CD
2. Monitor build performance and artifact sizes
3. User feedback collection and iteration
4. Maintenance documentation

## 8. Quality Assurance

### 8.1 Testing Strategy
- **Unit tests**: Embedded asset loading functionality
- **Integration tests**: Full character binary validation
- **Performance tests**: Binary size and startup time benchmarks
- **Cross-platform tests**: Functionality verification on all target platforms

### 8.2 Validation Checks
- Character card JSON schema validation
- Animation GIF format verification
- Binary functionality testing
- Memory usage profiling

This implementation plan transforms the DDS application from a runtime-asset-dependent application into a collection of standalone, zero-configuration executables while maintaining the existing codebase architecture and following the project's "lazy programmer" philosophy of leveraging standard library capabilities.