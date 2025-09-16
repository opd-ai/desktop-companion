#!/bin/bash

# scripts/dds-scripts.sh
# Desktop Companion Scripts Master Entry Point
#
# This script provides a unified interface to all Desktop Companion
# development and build scripts. It routes commands to the appropriate
# specialized scripts in the organized directory structure.
#
# Usage: ./scripts/dds-scripts.sh [CATEGORY] [COMMAND] [OPTIONS]
#
# Dependencies:
# - scripts/lib/common.sh
# - scripts/lib/config.sh

# Load shared libraries
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LIB_DIR="$SCRIPT_DIR/lib"

# shellcheck source=lib/common.sh
source "$LIB_DIR/common.sh"
# shellcheck source=lib/config.sh
source "$LIB_DIR/config.sh"

# ============================================================================
# HELP AND USAGE
# ============================================================================

show_main_usage() {
    cat << EOF
Desktop Companion Scripts - Master Entry Point
==============================================

Usage: $0 [CATEGORY] [COMMAND] [OPTIONS]

CATEGORIES:
    build              Character build and compilation scripts
    validation         Character and environment validation scripts  
    android            Android-specific build and testing scripts
    character          Character management and fixing scripts
    asset-generation   Character asset generation scripts
    release            Release preparation and validation scripts
    config             Configuration management utilities

GLOBAL OPTIONS:
    --help, -h         Show this help message
    --list-scripts     List all available scripts
    --show-config      Show current configuration
    --version          Show version information

EXAMPLES:
    $0 build characters                    # Build all characters
    $0 validation characters               # Validate all characters
    $0 android test-apk default           # Test Android APK build
    $0 character fix-validation           # Fix character validation issues
    $0 config show                        # Show current configuration

CATEGORY HELP:
    $0 [CATEGORY] --help                  # Show help for specific category

QUICK COMMANDS:
    $0 build                              # Same as: build characters
    $0 validate                           # Same as: validation characters  
    $0 fix                                # Same as: character fix-validation
    $0 android                            # Same as: android validate-environment

See individual script documentation for detailed usage information.

PROJECT STRUCTURE:
    scripts/
    ├── lib/                    # Shared utilities and configuration
    ├── build/                  # Build and compilation scripts
    ├── validation/             # Validation and testing scripts  
    ├── android/                # Android-specific scripts
    ├── character-management/   # Character management scripts
    ├── asset-generation/       # Asset generation scripts
    ├── release/               # Release preparation scripts
    └── dds-scripts.sh         # This master script

EOF
}

show_version_info() {
    cat << EOF
Desktop Companion Scripts
========================

Version: $(git describe --tags --always 2>/dev/null || echo "development")
Commit:  $(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
Branch:  $(git branch --show-current 2>/dev/null || echo "unknown")
Date:    $(date -u '+%Y-%m-%d %H:%M:%S UTC')

Project: $(basename "$PROJECT_ROOT")
Scripts: $SCRIPT_DIR

Environment:
- Go:    $(go version 2>/dev/null | awk '{print $3}' || echo "not found")
- Fyne:  $(fyne version 2>/dev/null || echo "not found")
- OS:    $(uname -s)/$(uname -m)

EOF
}

list_all_scripts() {
    cat << EOF
Available Scripts by Category
============================

BUILD SCRIPTS:
    build-characters           Build character-specific binaries
    cross-platform-build       Cross-platform CI/CD builds

VALIDATION SCRIPTS:
    validate-characters        Validate character JSON files
    validate-animations        Validate character animation files
    validate-binaries          Validate built character binaries
    validate-pipeline          Validate complete build pipeline
    validate-workflow          Validate GitHub Actions workflow

ASSET GENERATION SCRIPTS:
    generate-character-assets  Generate assets for all characters

RELEASE SCRIPTS:
    pre-release-validation     Pre-release validation suite

ANDROID SCRIPTS:
    validate-environment       Validate Android build environment
    test-apk-build             Test Android APK build process

CHARACTER MANAGEMENT:
    fix-validation-issues      Fix character validation issues

CONFIGURATION:
    show-config               Show current configuration
    save-config               Save configuration to file
    load-config               Load configuration from file

LEGACY SCRIPTS (maintained for compatibility):
EOF
    find "$SCRIPT_DIR" -maxdepth 1 -name "*.sh" -not -name "dds-scripts.sh" | sed 's|.*/||' | sort | sed 's/^/    /'
    echo ""
}

# ============================================================================
# SCRIPT ROUTING
# ============================================================================

# Route to build scripts
route_build_command() {
    local command="${1:-characters}"
    shift || true
    
    case "$command" in
        characters|character)
            exec "$SCRIPT_DIR/build/build-characters.sh" "$@"
            ;;
        cross-platform|ci|cicd)
            exec "$SCRIPT_DIR/build/cross-platform-build.sh" "$@"
            ;;
        --help|-h)
            echo "Build Scripts:"
            echo "  characters        Build character-specific binaries"
            echo "  cross-platform    Cross-platform CI/CD builds"
            echo ""
            echo "Usage: $0 build [COMMAND] [OPTIONS]"
            ;;
        *)
            error "Unknown build command: $command"
            echo "Available: characters, cross-platform"
            exit 1
            ;;
    esac
}

# Route to validation scripts
route_validation_command() {
    local command="${1:-characters}"
    shift || true
    
    case "$command" in
        characters|character)
            exec "$SCRIPT_DIR/validation/validate-characters.sh" "$@"
            ;;
        animations|animation)
            exec "$SCRIPT_DIR/validation/validate-animations.sh" "$@"
            ;;
        binaries|binary)
            exec "$SCRIPT_DIR/validation/validate-character-binaries.sh" "$@"
            ;;
        pipeline)
            exec "$SCRIPT_DIR/validation/validate-pipeline.sh" "$@"
            ;;
        workflow)
            exec "$SCRIPT_DIR/validation/validate-workflow.sh" "$@"
            ;;
        release)
            exec "$SCRIPT_DIR/release_validation.sh" "$@"
            ;;
        --help|-h)
            echo "Validation Scripts:"
            echo "  characters        Validate character JSON files"
            echo "  animations        Validate character animation files"
            echo "  binaries          Validate built character binaries"
            echo "  pipeline          Validate complete build pipeline"
            echo "  workflow          Validate GitHub Actions workflow"
            echo "  release           Pre-release validation suite"
            echo ""
            echo "Usage: $0 validation [COMMAND] [OPTIONS]"
            ;;
        *)
            error "Unknown validation command: $command"
            echo "Available: characters, animations, binaries, pipeline, workflow, release"
            exit 1
            ;;
    esac
}

# Route to Android scripts
route_android_command() {
    local command="${1:-validate-environment}"
    shift || true
    
    case "$command" in
        validate-environment|validate|env)
            exec "$SCRIPT_DIR/android/validate-environment.sh" "$@"
            ;;
        test-apk|test|apk)
            exec "$SCRIPT_DIR/android/test-apk-build.sh" "$@"
            ;;
        test-integrity|integrity)
            exec "$SCRIPT_DIR/test-android-apk.sh" "$@"
            ;;
        --help|-h)
            echo "Android Scripts:"
            echo "  validate-environment  Validate Android build environment"
            echo "  test-apk             Test Android APK build process"
            echo "  test-integrity       Test APK file integrity"
            echo ""
            echo "Usage: $0 android [COMMAND] [OPTIONS]"
            ;;
        *)
            error "Unknown Android command: $command"
            echo "Available: validate-environment, test-apk, test-integrity"
            exit 1
            ;;
    esac
}

# Route to character management scripts
route_character_command() {
    local command="${1:-fix-validation}"
    shift || true
    
    case "$command" in
        fix-validation|fix)
            exec "$SCRIPT_DIR/character-management/fix-validation-issues.sh" "$@"
            ;;
        --help|-h)
            echo "Character Management Scripts:"
            echo "  fix-validation       Fix character validation issues"
            echo ""
            echo "Usage: $0 character [COMMAND] [OPTIONS]"
            ;;
        *)
            error "Unknown character command: $command"
            echo "Available: fix-validation"
            exit 1
            ;;
    esac
}

# Route to asset generation scripts
route_asset_generation_command() {
    local command="${1:-generate}"
    shift || true
    
    case "$command" in
        generate)
            exec "$SCRIPT_DIR/asset-generation/generate-character-assets.sh" "$@"
            ;;
        simple)
            exec "$SCRIPT_DIR/asset-generation/generate-character-assets.sh" simple "$@"
            ;;
        validate)
            exec "$SCRIPT_DIR/asset-generation/generate-character-assets.sh" validate "$@"
            ;;
        rebuild)
            exec "$SCRIPT_DIR/asset-generation/generate-character-assets.sh" rebuild "$@"
            ;;
        --help|-h)
            echo "Asset Generation Scripts:"
            echo "  generate             Generate assets for all characters"
            echo "  simple               Use simplified generation (faster)"
            echo "  validate             Validate existing assets"
            echo "  rebuild              Force rebuild all assets"
            echo ""
            echo "Usage: $0 asset-generation [COMMAND] [OPTIONS]"
            ;;
        *)
            error "Unknown asset-generation command: $command"
            echo "Available: generate, simple, validate, rebuild"
            exit 1
            ;;
    esac
}

# Route to release scripts
route_release_command() {
    local command="${1:-validate}"
    shift || true
    
    case "$command" in
        validate|validation)
            exec "$SCRIPT_DIR/release/pre-release-validation.sh" "$@"
            ;;
        quick)
            exec "$SCRIPT_DIR/release/pre-release-validation.sh" quick "$@"
            ;;
        benchmark)
            exec "$SCRIPT_DIR/release/pre-release-validation.sh" benchmark "$@"
            ;;
        environment)
            exec "$SCRIPT_DIR/release/pre-release-validation.sh" environment "$@"
            ;;
        --help|-h)
            echo "Release Scripts:"
            echo "  validate             Full pre-release validation"
            echo "  quick                Quick validation (essential tests)"
            echo "  benchmark            Performance benchmarks only"
            echo "  environment          Environment validation only"
            echo ""
            echo "Usage: $0 release [COMMAND] [OPTIONS]"
            ;;
        *)
            error "Unknown release command: $command"
            echo "Available: validate, quick, benchmark, environment"
            exit 1
            ;;
    esac
}

# Route to configuration commands
route_config_command() {
    local command="${1:-show}"
    shift || true
    
    case "$command" in
        show|display)
            show_config
            ;;
        save)
            save_config_file "$@"
            ;;
        load)
            load_config_file "$@"
            ;;
        --help|-h)
            echo "Configuration Commands:"
            echo "  show              Show current configuration"
            echo "  save [FILE]       Save configuration to file"
            echo "  load [FILE]       Load configuration from file"
            echo ""
            echo "Usage: $0 config [COMMAND] [OPTIONS]"
            ;;
        *)
            error "Unknown config command: $command"
            echo "Available: show, save, load"
            exit 1
            ;;
    esac
}

# ============================================================================
# QUICK COMMANDS
# ============================================================================

# Handle quick command shortcuts
handle_quick_commands() {
    local command="$1"
    shift
    
    case "$command" in
        build)
            route_build_command characters "$@"
            ;;
        validate)
            route_validation_command characters "$@"
            ;;
        fix)
            route_character_command fix-validation "$@"
            ;;
        android)
            route_android_command validate-environment "$@"
            ;;
        *)
            return 1
            ;;
    esac
}

# ============================================================================
# MAIN ROUTING
# ============================================================================

# Main command router
main() {
    # Handle no arguments
    if [[ $# -eq 0 ]]; then
        show_main_usage
        exit 0
    fi
    
    # Handle global options first
    case "$1" in
        --help|-h)
            show_main_usage
            exit 0
            ;;
        --version|-v)
            show_version_info
            exit 0
            ;;
        --list-scripts)
            list_all_scripts
            exit 0
            ;;
        --show-config)
            show_config
            exit 0
            ;;
    esac
    
    # Try quick commands first
    if handle_quick_commands "$@"; then
        return
    fi
    
    # Route to category-specific handlers
    local category="$1"
    shift
    
    case "$category" in
        build)
            route_build_command "$@"
            ;;
        validation|validate)
            route_validation_command "$@"
            ;;
        android)
            route_android_command "$@"
            ;;
        character|char)
            route_character_command "$@"
            ;;
        asset-generation|assets|generate)
            route_asset_generation_command "$@"
            ;;
        release)
            route_release_command "$@"
            ;;
        config|configuration)
            route_config_command "$@"
            ;;
        *)
            error "Unknown category: $category"
            echo ""
            echo "Available categories: build, validation, android, character, asset-generation, release, config"
            echo "Quick commands: build, validate, fix, android"
            echo ""
            echo "Use '$0 --help' for more information."
            exit 1
            ;;
    esac
}

# Initialize and run
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    init_common
    main "$@"
fi
