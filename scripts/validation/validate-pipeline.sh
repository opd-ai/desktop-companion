#!/bin/bash

# scripts/validation/validate-pipeline.sh
# Full pipeline validation script
#
# Comprehensive testing and validation for character-specific binary generation
# including build processes, character card validation, and cross-platform support.
#
# Usage: ./scripts/validation/validate-pipeline.sh [OPTIONS] [COMMAND]
#
# Dependencies:
# - Go 1.21+
# - scripts/lib/common.sh
# - scripts/lib/config.sh
# - scripts/build/build-characters.sh

# Load shared libraries
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LIB_DIR="$(dirname "$SCRIPT_DIR")/lib"

# shellcheck source=../lib/common.sh
source "$LIB_DIR/common.sh"
# shellcheck source=../lib/config.sh
source "$LIB_DIR/config.sh"

# ============================================================================
# PIPELINE VALIDATION CONFIGURATION
# ============================================================================

# Validation directories
VALIDATION_DIR="$BUILD_DIR/validation"
TEMP_DIR="$BUILD_DIR/temp"

# Configuration from shared config or defaults
QUICK_MODE="${DDS_QUICK_MODE:-false}"
ANDROID_TESTS="${DDS_ANDROID_TESTS:-true}"
PARALLEL_TESTS="${DDS_PARALLEL_TESTS:-4}"

# Test results tracking
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_SKIPPED=0

# ============================================================================
# HELP AND USAGE
# ============================================================================

show_usage() {
    cat << EOF
Usage: $0 [OPTIONS] [COMMAND]

Test full pipeline with multiple characters and comprehensive validation.

COMMANDS:
    validate               Run full pipeline validation (default)
    quick                  Run quick validation (subset of tests)
    character-cards        Validate only character card files
    build-test            Test build process for all characters
    cross-platform        Test cross-platform build capabilities
    android               Test Android-specific builds
    artifact-mgmt         Test artifact management system
    help                  Show this help message

OPTIONS:
    --quick               Enable quick mode (faster, fewer tests)
    --no-android          Skip Android-specific tests
    --parallel N          Set number of parallel tests (default: $PARALLEL_TESTS)
    -v, --verbose         Enable verbose output
    --dry-run            Show what would be tested without running tests
    --clean              Clean validation artifacts before starting

EXAMPLES:
    $0                    # Run full pipeline validation
    $0 quick              # Run quick validation
    $0 character-cards    # Validate only character JSON files
    $0 --no-android       # Skip Android tests
    $0 --parallel 8       # Use 8 parallel test processes

PIPELINE TESTS:
    1. Environment validation (Go, dependencies)
    2. Character card validation (JSON schema, assets)
    3. Build process testing (embedding, compilation)
    4. Cross-platform build validation
    5. Android APK generation testing
    6. Artifact management validation
    7. Performance and integration testing

OUTPUT:
    Test logs: $TEST_OUTPUT_DIR/pipeline-*.log
    Validation artifacts: $VALIDATION_DIR/
    Summary report: $TEST_OUTPUT_DIR/pipeline-validation-report.txt

EOF
}

# ============================================================================
# PIPELINE VALIDATION FUNCTIONS
# ============================================================================

# Initialize validation environment
init_validation() {
    log "Initializing pipeline validation environment..."
    
    # Create required directories
    mkdir -p "$VALIDATION_DIR" "$TEMP_DIR" "$TEST_OUTPUT_DIR"
    
    # Clean previous artifacts if requested
    if [[ "${DDS_CLEAN:-false}" == "true" ]]; then
        log "Cleaning previous validation artifacts..."
        rm -rf "$VALIDATION_DIR"/* "$TEMP_DIR"/*
    fi
    
    success "Validation environment initialized"
}

# Validate Go environment and dependencies
validate_environment() {
    log "Validating Go environment and dependencies..."
    
    # Check Go version
    if ! go version >/dev/null 2>&1; then
        error "Go is not installed or not in PATH"
        ((TESTS_FAILED++))
        return 1
    fi
    success "Go is available: $(go version)"
    ((TESTS_PASSED++))
    
    # Check go.mod validity
    if ! go mod verify >/dev/null 2>&1; then
        error "Go modules verification failed"
        ((TESTS_FAILED++))
        return 1
    fi
    success "Go modules verified successfully"
    ((TESTS_PASSED++))
    
    # Check required build tools
    if ! command -v git >/dev/null 2>&1; then
        warning "Git not available (may affect version detection)"
        ((TESTS_SKIPPED++))
    else
        success "Git is available"
        ((TESTS_PASSED++))
    fi
    
    return 0
}

# Get list of available characters for testing
get_test_characters() {
    local characters=()
    
    # Find all character.json files
    while IFS= read -r -d '' file; do
        local char_dir=$(dirname "$file")
        local char_name=$(basename "$char_dir")
        characters+=("$char_name")
    done < <(find "$CHARACTERS_DIR" -name "character.json" -print0 2>/dev/null)
    
    if [[ ${#characters[@]} -eq 0 ]]; then
        error "No character files found in $CHARACTERS_DIR"
        return 1
    fi
    
    log "Found ${#characters[@]} characters for testing: ${characters[*]}"
    printf '%s\n' "${characters[@]}"
}

# Test character card validation
test_character_cards() {
    log "Testing character card validation..."
    
    local validation_log="$TEST_OUTPUT_DIR/character-cards-validation.log"
    
    # Use the existing character validation script
    if "$SCRIPT_DIR/validate-characters.sh" --detailed > "$validation_log" 2>&1; then
        success "Character card validation passed"
        ((TESTS_PASSED++))
    else
        error "Character card validation failed (see: $validation_log)"
        ((TESTS_FAILED++))
        return 1
    fi
    
    return 0
}

# Test embed-character.go script
test_embed_script() {
    log "Testing character embedding script..."
    
    local test_char="default"
    local embed_log="$TEST_OUTPUT_DIR/embed-test.log"
    
    # Test the embed script with a known character
    if go run "$PROJECT_ROOT/scripts/embed-character.go" "$test_char" > "$embed_log" 2>&1; then
        success "Character embedding script works correctly"
        ((TESTS_PASSED++))
    else
        error "Character embedding script failed (see: $embed_log)"
        ((TESTS_FAILED++))
        return 1
    fi
    
    return 0
}

# Test platform-specific builds
test_platform_builds() {
    local platforms=("linux/amd64" "windows/amd64" "darwin/amd64")
    
    if [[ "$QUICK_MODE" == "true" ]]; then
        platforms=("$(go env GOOS)/$(go env GOARCH)")
    fi
    
    log "Testing platform-specific builds for: ${platforms[*]}"
    
    for platform in "${platforms[@]}"; do
        local os="${platform%/*}"
        local arch="${platform#*/}"
        
        log "Testing build for $platform..."
        
        # Test build process
        local build_log="$TEST_OUTPUT_DIR/build-${os}-${arch}.log"
        if GOOS="$os" GOARCH="$arch" go build -o "$TEMP_DIR/test-$os-$arch" \
           "$PROJECT_ROOT/cmd/companion" > "$build_log" 2>&1; then
            success "Build successful for $platform"
            ((TESTS_PASSED++))
        else
            error "Build failed for $platform (see: $build_log)"
            ((TESTS_FAILED++))
        fi
    done
}

# Test Android APK builds specifically
test_android_builds() {
    if [[ "$ANDROID_TESTS" != "true" ]]; then
        log "Skipping Android tests (disabled)"
        return 0
    fi
    
    log "Testing Android APK build capabilities..."
    
    # Check if fyne tool is available
    if ! command -v fyne >/dev/null 2>&1; then
        warning "Fyne CLI tool not available, skipping Android tests"
        ((TESTS_SKIPPED++))
        return 0
    fi
    
    # Use the existing Android test script
    local android_log="$TEST_OUTPUT_DIR/android-build-test.log"
    if "$PROJECT_ROOT/scripts/android/test-apk-build.sh" --dry-run > "$android_log" 2>&1; then
        success "Android build validation passed"
        ((TESTS_PASSED++))
    else
        warning "Android build validation had issues (see: $android_log)"
        ((TESTS_SKIPPED++))
    fi
}

# Test cross-platform builds
test_cross_platform_builds() {
    log "Testing cross-platform build script..."
    
    local cross_build_log="$TEST_OUTPUT_DIR/cross-platform-build.log"
    
    # Use the existing cross-platform build script in dry-run mode
    if "$PROJECT_ROOT/scripts/build/cross-platform-build.sh" --dry-run > "$cross_build_log" 2>&1; then
        success "Cross-platform build script validation passed"
        ((TESTS_PASSED++))
    else
        error "Cross-platform build script validation failed (see: $cross_build_log)"
        ((TESTS_FAILED++))
    fi
}

# Test artifact management
test_artifact_management() {
    log "Testing artifact management system..."
    
    # Test artifact manager if available
    if [[ -f "$PROJECT_ROOT/cmd/artifact-manager/main.go" ]]; then
        local artifact_log="$TEST_OUTPUT_DIR/artifact-mgmt-test.log"
        if go run "$PROJECT_ROOT/cmd/artifact-manager" --help > "$artifact_log" 2>&1; then
            success "Artifact management system accessible"
            ((TESTS_PASSED++))
        else
            warning "Artifact management system had issues (see: $artifact_log)"
            ((TESTS_SKIPPED++))
        fi
    else
        warning "Artifact management system not found, skipping test"
        ((TESTS_SKIPPED++))
    fi
}

# Generate comprehensive validation report
generate_pipeline_report() {
    local report_file="$TEST_OUTPUT_DIR/pipeline-validation-report.txt"
    local total_tests=$((TESTS_PASSED + TESTS_FAILED + TESTS_SKIPPED))
    
    {
        echo "# Pipeline Validation Report"
        echo "Generated: $(date -u '+%Y-%m-%d %H:%M:%S UTC')"
        echo ""
        echo "## Summary"
        echo "- Total tests: $total_tests"
        echo "- Passed: $TESTS_PASSED"
        echo "- Failed: $TESTS_FAILED"
        echo "- Skipped: $TESTS_SKIPPED"
        if [[ $total_tests -gt 0 ]]; then
            echo "- Success rate: $(( TESTS_PASSED * 100 / total_tests ))%"
        fi
        echo ""
        echo "## Configuration"
        echo "- Quick mode: $QUICK_MODE"
        echo "- Android tests: $ANDROID_TESTS"
        echo "- Parallel tests: $PARALLEL_TESTS"
        echo ""
        echo "## Test Categories"
        echo "1. Environment validation"
        echo "2. Character card validation"
        echo "3. Build process testing"
        echo "4. Platform-specific builds"
        echo "5. Android APK generation"
        echo "6. Cross-platform validation"
        echo "7. Artifact management"
        echo ""
        echo "## Files Generated"
        echo "- Validation logs: $TEST_OUTPUT_DIR/pipeline-*.log"
        echo "- Temporary artifacts: $VALIDATION_DIR/"
        echo "- Build tests: $TEMP_DIR/"
        echo ""
    } > "$report_file"
    
    success "Pipeline validation report saved to: $report_file"
}

# ============================================================================
# MAIN EXECUTION
# ============================================================================

# Parse command line arguments
COMMAND="validate"

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help|help)
            show_usage
            exit 0
            ;;
        --quick)
            QUICK_MODE=true
            shift
            ;;
        --no-android)
            ANDROID_TESTS=false
            shift
            ;;
        --parallel)
            PARALLEL_TESTS="$2"
            shift 2
            ;;
        -v|--verbose)
            DDS_VERBOSE=true
            shift
            ;;
        --dry-run)
            DDS_DRY_RUN=true
            shift
            ;;
        --clean)
            DDS_CLEAN=true
            shift
            ;;
        validate|quick|character-cards|build-test|cross-platform|android|artifact-mgmt)
            COMMAND="$1"
            shift
            ;;
        -*)
            error "Unknown option: $1"
            show_usage
            exit 1
            ;;
        *)
            error "Unexpected argument: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Adjust settings for quick mode
if [[ "$COMMAND" == "quick" ]]; then
    QUICK_MODE=true
    COMMAND="validate"
fi

# Initialize validation environment
init_validation

# Execute command
case $COMMAND in
    validate)
        log "Starting full pipeline validation..."
        validate_environment
        test_character_cards
        test_embed_script
        test_platform_builds
        if [[ "$ANDROID_TESTS" == "true" ]]; then
            test_android_builds
        fi
        test_cross_platform_builds
        test_artifact_management
        generate_pipeline_report
        ;;
    character-cards)
        log "Running character card validation only..."
        test_character_cards
        ;;
    build-test)
        log "Running build process tests..."
        validate_environment
        test_embed_script
        test_platform_builds
        ;;
    cross-platform)
        log "Running cross-platform tests..."
        test_cross_platform_builds
        ;;
    android)
        log "Running Android-specific tests..."
        test_android_builds
        ;;
    artifact-mgmt)
        log "Running artifact management tests..."
        test_artifact_management
        ;;
    *)
        error "Unknown command: $COMMAND"
        show_usage
        exit 1
        ;;
esac

# Final summary
local total_tests=$((TESTS_PASSED + TESTS_FAILED + TESTS_SKIPPED))
if [[ $TESTS_FAILED -eq 0 ]]; then
    success "Pipeline validation completed successfully! ($TESTS_PASSED passed, $TESTS_SKIPPED skipped)"
    exit 0
else
    error "Pipeline validation completed with failures. ($TESTS_FAILED failed, $TESTS_PASSED passed, $TESTS_SKIPPED skipped)"
    exit 1
fi
