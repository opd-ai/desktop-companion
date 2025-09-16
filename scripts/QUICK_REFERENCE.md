# Desktop Companion Scripts - Quick Reference

## 🚀 Master Entry Point

```bash
./scripts/dds-scripts.sh [CATEGORY] [COMMAND] [OPTIONS]
```

## 📋 Quick Commands

```bash
# Most common operations
./scripts/dds-scripts.sh build            # Build all characters
./scripts/dds-scripts.sh validate         # Validate all characters
./scripts/dds-scripts.sh fix              # Fix validation issues
./scripts/dds-scripts.sh android          # Check Android environment

# Get help
./scripts/dds-scripts.sh --help           # Main help
./scripts/dds-scripts.sh build --help     # Category help
```

## 🗂️ Categories

### BUILD
```bash
./scripts/dds-scripts.sh build characters          # Build character binaries
./scripts/dds-scripts.sh build cross-platform      # CI/CD builds
```

### VALIDATION
```bash
./scripts/dds-scripts.sh validation characters     # Validate JSON files
./scripts/dds-scripts.sh validation animations     # Validate animations  
./scripts/dds-scripts.sh validation binaries       # Test built binaries
./scripts/dds-scripts.sh validation pipeline       # Full pipeline test
./scripts/dds-scripts.sh validation workflow       # GitHub Actions test
```

### ANDROID
```bash
./scripts/dds-scripts.sh android validate-env      # Check environment
./scripts/dds-scripts.sh android test-apk default  # Test APK build
```

### CHARACTER MANAGEMENT
```bash
./scripts/dds-scripts.sh character fix-validation  # Fix JSON issues
```

### ASSET GENERATION
```bash
./scripts/dds-scripts.sh asset-generation generate # Generate all assets
./scripts/dds-scripts.sh asset-generation simple   # Quick generation
```

### RELEASE
```bash
./scripts/dds-scripts.sh release validate          # Pre-release validation
./scripts/dds-scripts.sh release benchmark         # Performance tests
```

## 🛠️ Direct Script Access

For power users who prefer direct access:

```bash
# Build scripts
./scripts/build/build-characters.sh
./scripts/build/cross-platform-build.sh

# Validation scripts  
./scripts/validation/validate-characters.sh
./scripts/validation/validate-animations.sh
./scripts/validation/validate-binaries.sh
./scripts/validation/validate-pipeline.sh
./scripts/validation/validate-workflow.sh

# Android scripts
./scripts/android/validate-environment.sh
./scripts/android/test-apk-build.sh

# Character management
./scripts/character-management/fix-validation-issues.sh

# Asset generation
./scripts/asset-generation/generate-character-assets.sh

# Release preparation
./scripts/release/pre-release-validation.sh
```

## 🔧 Configuration

```bash
./scripts/dds-scripts.sh config show               # Show current config
./scripts/dds-scripts.sh config save config.env   # Save config
./scripts/dds-scripts.sh config load config.env   # Load config
```

## 📚 Information Commands

```bash
./scripts/dds-scripts.sh --version                 # Version info
./scripts/dds-scripts.sh --list-scripts            # All scripts
./scripts/dds-scripts.sh --show-config             # Current config
```

## 🚨 Legacy Scripts (DEPRECATED)

Legacy wrapper scripts in the root are deprecated:
- `build-characters.sh` → Use `dds-scripts.sh build characters`
- `validate-characters.sh` → Use `dds-scripts.sh validation characters`
- `test-android-*.sh` → Use `dds-scripts.sh android test-apk`

Run `./scripts/cleanup-legacy-wrappers.sh --dry-run` to see cleanup plan.

## 🏗️ Project Structure

```
scripts/
├── dds-scripts.sh                    # 🎯 Master entry point
├── lib/                              # 📚 Shared utilities
│   ├── common.sh                     # Logging, paths, utilities
│   └── config.sh                     # Configuration management
├── build/                            # 🔨 Build scripts
├── validation/                       # ✅ Testing & validation
├── android/                          # 📱 Android-specific
├── character-management/             # 👤 Character operations
├── asset-generation/                 # 🎨 Asset pipeline
└── release/                          # 🚀 Release preparation
```

## Master Script Usage

```bash
# Master entry point - unified interface to all scripts
./scripts/dds-scripts.sh [CATEGORY] [COMMAND] [OPTIONS]
```

## Quick Commands

| Command | Equivalent | Description |
|---------|------------|-------------|
| `./scripts/dds-scripts.sh build` | `build characters` | Build all character binaries |
| `./scripts/dds-scripts.sh validate` | `validation characters` | Validate all characters |
| `./scripts/dds-scripts.sh fix` | `character fix-validation` | Fix validation issues |
| `./scripts/dds-scripts.sh android` | `android validate-environment` | Validate Android environment |

## Category Commands

### build/
```bash
./scripts/dds-scripts.sh build characters [OPTIONS]     # Build character binaries
./scripts/dds-scripts.sh build cross-platform [OPTIONS] # Cross-platform CI builds
```

### validation/
```bash
./scripts/dds-scripts.sh validation characters [OPTIONS]     # Validate character JSON
./scripts/dds-scripts.sh validation animations [OPTIONS]     # Validate animations
./scripts/dds-scripts.sh validation binaries [OPTIONS]       # Validate binaries
./scripts/dds-scripts.sh validation pipeline [OPTIONS]       # Full pipeline validation
./scripts/dds-scripts.sh validation workflow [OPTIONS]       # GitHub Actions validation
```

### android/
```bash
./scripts/dds-scripts.sh android validate-environment [OPTIONS] # Check environment
./scripts/dds-scripts.sh android test-apk [OPTIONS]             # Test APK build
./scripts/dds-scripts.sh android test-integrity [OPTIONS]      # APK integrity check
```

### character-management/
```bash
./scripts/dds-scripts.sh character fix-validation [OPTIONS]   # Fix validation issues
```

### asset-generation/
```bash
./scripts/dds-scripts.sh asset-generation generate [OPTIONS]    # Generate all assets
./scripts/dds-scripts.sh asset-generation simple [OPTIONS]      # Simple generation
./scripts/dds-scripts.sh asset-generation validate [OPTIONS]    # Validate assets
./scripts/dds-scripts.sh asset-generation rebuild [OPTIONS]     # Force rebuild assets
```

### release/
```bash
./scripts/dds-scripts.sh release validate [OPTIONS]            # Full pre-release validation
./scripts/dds-scripts.sh release quick [OPTIONS]               # Quick validation
./scripts/dds-scripts.sh release benchmark [OPTIONS]           # Performance benchmarks
./scripts/dds-scripts.sh release environment [OPTIONS]         # Environment validation
```

## Configuration Commands

```bash
./scripts/dds-scripts.sh config show           # Show current configuration
./scripts/dds-scripts.sh config save           # Save configuration to file
./scripts/dds-scripts.sh config load           # Load configuration from file
./scripts/dds-scripts.sh config reset          # Reset to defaults
```

## Help Commands

```bash
./scripts/dds-scripts.sh --help                    # Master script help
./scripts/dds-scripts.sh [CATEGORY] --help         # Category-specific help
./scripts/dds-scripts.sh [CATEGORY] [COMMAND] --help # Command-specific help
```

## Environment Variables

### Build Configuration
```bash
export DDS_MAX_PARALLEL=8                    # Parallel build jobs (default: 4)
export DDS_PLATFORMS="linux/amd64,android/arm64" # Target platforms
export DDS_LDFLAGS="-s -w"                   # Go linker flags
export DDS_VERBOSE=true                      # Verbose output
export DDS_DRY_RUN=true                      # Simulate operations
```

### Android Configuration
```bash
export DDS_ANDROID_HOME=/path/to/android/sdk # Android SDK path
export DDS_APP_ID=ai.opd.dds                # Android app ID
export DDS_MIN_SDK=21                       # Minimum Android SDK
export DDS_TARGET_SDK=34                    # Target Android SDK
```

### Character Configuration
```bash
export DDS_CHARACTERS_DIR=/path/to/characters # Characters directory
export DDS_ANIMATIONS_REQUIRED=true          # Require animations
export DDS_VALIDATION_STRICT=true           # Strict validation mode
```

## Direct Script Access

For backward compatibility, all original scripts remain available:

### Build Scripts
```bash
./scripts/build/build-characters.sh [OPTIONS]
./scripts/build/cross-platform-build.sh [OPTIONS]
```

### Validation Scripts  
```bash
./scripts/validation/validate-characters.sh [OPTIONS]
./scripts/validation/validate-animations.sh [OPTIONS]
./scripts/validation/validate-character-binaries.sh [OPTIONS]
./scripts/validation/validate-pipeline.sh [OPTIONS]
./scripts/validation/validate-workflow.sh [OPTIONS]
./scripts/validation/release-validation.sh [OPTIONS]
```

### Android Scripts
```bash
./scripts/android/validate-environment.sh [OPTIONS]
./scripts/android/test-apk-build.sh [OPTIONS]
./scripts/android/test-apk-integrity.sh [OPTIONS]
```

### Character Management Scripts
```bash
./scripts/character-management/fix-validation-issues.sh [OPTIONS]
./scripts/character-management/generate-assets-simple.sh [OPTIONS]
./scripts/character-management/generate-assets-full.sh [OPTIONS]
```

## Common Options

Most scripts support these common options:

| Option | Description |
|--------|-------------|
| `--help` | Show help information |
| `--verbose` | Enable verbose output |
| `--dry-run` | Simulate operations without making changes |
| `--parallel N` | Set number of parallel jobs |
| `--platform PLATFORM` | Target specific platform |
| `--config FILE` | Use specific configuration file |
| `--debug` | Enable debug output |
| `--quiet` | Suppress non-error output |

## Examples

### Development Workflow
```bash
# Validate all characters
./scripts/dds-scripts.sh validate

# Fix any validation issues
./scripts/dds-scripts.sh fix

# Build for current platform
./scripts/dds-scripts.sh build

# Test Android environment
./scripts/dds-scripts.sh android
```

### CI/CD Workflow
```bash
# Full validation pipeline
./scripts/dds-scripts.sh validation pipeline

# Cross-platform builds
export DDS_PLATFORMS="linux/amd64,darwin/amd64,windows/amd64,android/arm64"
./scripts/dds-scripts.sh build cross-platform

# Release validation
./scripts/dds-scripts.sh validation release
```

### Character Development
```bash
# Generate character assets
./scripts/dds-scripts.sh character generate-simple

# Validate specific character
./scripts/dds-scripts.sh validation characters --character klippy

# Fix validation issues for specific character
./scripts/dds-scripts.sh character fix-validation --character klippy

# Build character binary
./scripts/dds-scripts.sh build --character klippy
```

### Android Development
```bash
# Check Android environment
./scripts/dds-scripts.sh android validate-environment

# Test APK build process
./scripts/dds-scripts.sh android test-apk

# Validate APK integrity
./scripts/dds-scripts.sh android test-integrity
```

## Troubleshooting

### Common Issues

1. **Permission denied**: Ensure scripts are executable
   ```bash
   chmod +x scripts/dds-scripts.sh
   chmod +x scripts/**/*.sh
   ```

2. **Command not found**: Use full path to scripts
   ```bash
   ./scripts/dds-scripts.sh instead of dds-scripts.sh
   ```

3. **Missing dependencies**: Check shared libraries
   ```bash
   source scripts/lib/common.sh
   source scripts/lib/config.sh
   ```

4. **Configuration issues**: Reset and reconfigure
   ```bash
   ./scripts/dds-scripts.sh config reset
   ./scripts/dds-scripts.sh config show
   ```

### Debug Mode

Enable debug output for troubleshooting:
```bash
export DDS_DEBUG=true
./scripts/dds-scripts.sh --debug [COMMAND]
```

### Getting Help

1. **General help**: `./scripts/dds-scripts.sh --help`
2. **Category help**: `./scripts/dds-scripts.sh [CATEGORY] --help`
3. **Command help**: `./scripts/dds-scripts.sh [CATEGORY] [COMMAND] --help`
4. **Configuration help**: `./scripts/dds-scripts.sh config --help`
5. **Version info**: `./scripts/dds-scripts.sh --version`
