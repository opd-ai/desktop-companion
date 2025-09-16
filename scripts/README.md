# Character Asset Generation

This directory contains scripts for generating GIF assets for all character JSON files in the Desktop Companion project.

## Scripts

### `generate-assets.sh` (Main Entry Point)
Interactive wrapper script that provides quick access to asset generation features.

**Usage:**
```bash
# Interactive menu
./generate-assets.sh

# Direct usage
./generate-assets.sh --dry-run --verbose
```

### `scripts/generate-character-assets-simple.sh` (Recommended)
Simple, reliable script for sequential processing of all character files.

**Features:**
- Processes exactly 20 character archetypes
- Sequential execution (no concurrency issues)
- Clean output and error handling
- Supports dry-run mode for testing

**Usage:**
```bash
# Basic generation
./scripts/generate-character-assets-simple.sh

# Dry run to preview
./scripts/generate-character-assets-simple.sh --dry-run --verbose

# Custom style and model
./scripts/generate-character-assets-simple.sh --style realistic --model sdxl

# Help
./scripts/generate-character-assets-simple.sh --help
```

### `scripts/generate-all-character-assets.sh` (Advanced)
Full-featured script with parallel processing and comprehensive logging.

**Features:**
- Parallel job execution (configurable)
- Comprehensive logging with timestamps
- Advanced error handling and recovery
- Extensive configuration options

**Usage:**
```bash
# Parallel processing
./scripts/generate-all-character-assets.sh --jobs 4 --verbose

# Full feature set
./scripts/generate-all-character-assets.sh --help
```

## Character Files Processed

The scripts automatically find and process these 20 character archetypes:

1. `aria_luna` - Custom character example
2. `challenge` - High difficulty archetype
3. `default` - Basic companion archetype
4. `easy` - Low difficulty archetype
5. `flirty` - Playful personality archetype
6. `hard` - Medium-high difficulty archetype
7. `klippy` - Klipper/3D printing themed character
8. `llm_example` - LLM integration example
9. `markov_example` - Markov chain dialog example
10. `multiplayer` - Network multiplayer archetype
11. `news_example` - News/information archetype
12. `normal` - Standard difficulty archetype
13. `romance` - Basic romance archetype
14. `romance_flirty` - Flirty romance variant
15. `romance_slowburn` - Slow burn romance variant
16. `romance_supportive` - Supportive romance variant
17. `romance_tsundere` - Tsundere romance variant
18. `slow_burn` - Gradual relationship archetype
19. `specialist` - Advanced feature archetype
20. `tsundere` - Tsundere personality archetype

## Generated Assets

For each character, the gif-generator creates:
- `animations/idle.gif` - Default animation
- `animations/talking.gif` - Dialog animation
- `animations/happy.gif` - Positive emotion animation
- `animations/sad.gif` - Negative emotion animation
- Additional state-specific animations based on character configuration

## Configuration

Default settings:
- **Style**: `anime` (alternatives: realistic, cartoon, pixel)
- **Model**: `sd15` (alternatives: sdxl, dall-e)
- **Backup**: Enabled (preserves existing assets)
- **Validation**: Enabled (checks generated assets)

## Requirements

1. **Go 1.21+** - For building gif-generator
2. **gif-generator** - Built automatically from `cmd/gif-generator/`
3. **Character files** - JSON configurations in `assets/characters/*/character.json`

## Troubleshooting

### Build Issues
```bash
# Manually build gif-generator
cd /path/to/project
go build -ldflags="-s -w" -o build/gif-generator cmd/gif-generator/main.go
```

### Character File Issues
```bash
# List all character files found
find assets/characters -mindepth 2 -maxdepth 2 -name "character.json" -type f | sort
```

### Asset Validation
```bash
# Validate specific character
./build/gif-generator validate --path assets/characters/default --recursive
```

## Examples

### Quick Start
```bash
# Generate all assets with defaults
./generate-assets.sh
# Choose option 1

# Preview what will be generated
./generate-assets.sh
# Choose option 2
```

### Custom Generation
```bash
# Realistic style with SDXL model
./scripts/generate-character-assets-simple.sh --style realistic --model sdxl --verbose

# Dry run with custom settings
./scripts/generate-character-assets-simple.sh --dry-run --style cartoon --model sd15
```

### Performance Testing
```bash
# Fast parallel generation (advanced script)
./generate-assets.sh --jobs 4 --verbose

# Or directly:
./scripts/generate-all-character-assets.sh --jobs 4 --verbose --no-backup
```

## Output

Generated assets are placed in each character's directory:
```
assets/characters/default/
├── character.json
└── animations/
    ├── idle.gif
    ├── talking.gif
    ├── happy.gif
    └── sad.gif
```

Log files are created in `test_output/asset-generation-TIMESTAMP.log` for debugging and monitoring progress.
