#!/bin/bash

# DEPRECATED: Legacy wrapper for validate-pipeline.sh
# This script is maintained for backward compatibility.
# New usage: ./scripts/dds-scripts.sh validation pipeline
# Direct usage: ./scripts/validation/validate-pipeline.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Forward all arguments to the new refactored script
exec "$SCRIPT_DIR/validation/validate-pipeline.sh" "$@"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
QUICK_MODE=${QUICK_MODE:-false}
ANDROID_TESTS=${ANDROID_TESTS:-true}
PARALLEL_TESTS=${PARALLEL_TESTS:-4}

# Test results tracking
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_SKIPPED=0

log() {
    echo -e "${BLUE}[$(date +'%H:%M:%S')]${NC} $1"
}

success() {
    echo -e "${GREEN}✓${NC} $1"
    ((TESTS_PASSED++))
}

warning() {
    echo -e "${YELLOW}⚠${NC} $1"
    ((TESTS_SKIPPED++))
}

error() {
    echo -e "${RED}✗${NC} $1" >&2
    ((TESTS_FAILED++))
}

# Initialize validation environment
init_validation() {
    log "Initializing pipeline validation environment"
    
    # Create validation directories
    mkdir -p "$VALIDATION_DIR"
    mkdir -p "$TEMP_DIR"
    
    # Clean previous results
    rm -rf "$VALIDATION_DIR"/*
    rm -rf "$TEMP_DIR"/*
    
    # Create test report structure
    mkdir -p "$VALIDATION_DIR/reports"
    mkdir -p "$VALIDATION_DIR/artifacts"
    mkdir -p "$VALIDATION_DIR/logs"
    
    success "Validation environment initialized"
}

# Validate Go environment and dependencies
validate_environment() {
    log "Validating build environment"
    
    # Check Go installation
    if ! command -v go &> /dev/null; then
        error "Go is not installed or not in PATH"
        return 1
    fi
    
    local go_version
    go_version=$(go version | awk '{print $3}' | sed 's/go//')
    if [[ "$(printf '%s\n' "1.21" "$go_version" | sort -V | head -n1)" != "1.21" ]]; then
        error "Go version $go_version is below required minimum 1.21"
        return 1
    fi
    success "Go $go_version installed and compatible"
    
    # Check project structure
    if [[ ! -f "$PROJECT_ROOT/go.mod" ]]; then
        error "go.mod not found in project root"
        return 1
    fi
    success "Project structure valid"
    
    # Validate required scripts exist
    local required_scripts=("embed-character.go" "build-characters.sh")
    for script in "${required_scripts[@]}"; do
        if [[ ! -f "$SCRIPT_DIR/$script" ]]; then
            error "Required script missing: $script"
            return 1
        fi
    done
    success "Required scripts present"
    
    # Check fyne CLI for Android tests
    if [[ "$ANDROID_TESTS" == "true" ]]; then
        if command -v fyne &> /dev/null; then
            success "Fyne CLI available for Android testing"
        else
            warning "Fyne CLI not found - Android tests will be skipped"
            ANDROID_TESTS=false
        fi
    fi
    
    return 0
}

# Get list of available characters for testing
get_test_characters() {
    local all_characters
    all_characters=$(find "$PROJECT_ROOT/assets/characters" -maxdepth 1 -type d \
        -not -path "*/assets/characters" \
        -not -path "*/examples" \
        -not -path "*/templates" \
        -exec basename {} \; | sort)
    
    if [[ "$QUICK_MODE" == "true" ]]; then
        # Use only a subset for quick testing
        echo "$all_characters" | head -3
    else
        echo "$all_characters"
    fi
}

# Test character card validation
test_character_cards() {
    log "Testing character card validation"
    
    local characters
    characters=$(get_test_characters)
    local card_errors=0
    
    for char in $characters; do
        local card_path="$PROJECT_ROOT/assets/characters/$char/character.json"
        
        if [[ ! -f "$card_path" ]]; then
            error "Character card not found: $char"
            ((card_errors++))
            continue
        fi
        
        # Validate JSON syntax
        if ! jq empty "$card_path" 2>/dev/null; then
            error "Invalid JSON in character card: $char"
            ((card_errors++))
            continue
        fi
        
        # Validate required fields
        local required_fields=("name" "description" "animations")
        for field in "${required_fields[@]}"; do
            if ! jq -e ".$field" "$card_path" >/dev/null 2>&1; then
                error "Missing required field '$field' in character: $char"
                ((card_errors++))
                continue 2
            fi
        done
        
        # Validate animations exist
        local animations
        animations=$(jq -r '.animations | to_entries[] | .value' "$card_path" 2>/dev/null || echo "")
        
        for anim_path in $animations; do
            if [[ -n "$anim_path" ]]; then
                local full_anim_path="$PROJECT_ROOT/assets/characters/$char/$anim_path"
                if [[ ! -f "$full_anim_path" ]]; then
                    warning "Animation file missing for $char: $anim_path"
                fi
            fi
        done
        
        success "Character card valid: $char"
    done
    
    if [[ $card_errors -eq 0 ]]; then
        success "All character cards validated successfully"
    else
        error "$card_errors character card validation errors"
        return 1
    fi
}

# Test embed-character.go script
test_embed_script() {
    log "Testing character embedding script"
    
    local test_char="default"
    local test_output="$TEMP_DIR/embed-test"
    
    # Check if default character exists
    if [[ ! -d "$PROJECT_ROOT/assets/characters/$test_char" ]]; then
        # Find any available character for testing
        test_char=$(get_test_characters | head -1)
        if [[ -z "$test_char" ]]; then
            error "No characters available for embed testing"
            return 1
        fi
    fi
    
    log "Testing embed script with character: $test_char"
    
    # Test the embed script
    if go run "$SCRIPT_DIR/embed-character.go" \
        -character "$test_char" \
        -output "$test_output" 2>"$VALIDATION_DIR/logs/embed-test.log"; then
        success "Embed script executed successfully"
    else
        error "Embed script failed for character: $test_char"
        return 1
    fi
    
    # Validate generated files
    if [[ ! -f "$test_output/main.go" ]]; then
        error "Embed script did not generate main.go"
        return 1
    fi
    
    # Validate generated Go code compiles
    if go build -o "$TEMP_DIR/test-binary" "$test_output/main.go" 2>"$VALIDATION_DIR/logs/embed-compile.log"; then
        success "Generated embedded code compiles successfully"
    else
        error "Generated embedded code failed to compile"
        return 1
    fi
    
    # Cleanup
    rm -rf "$test_output"
    rm -f "$TEMP_DIR/test-binary"
    
    return 0
}

# Test platform-specific builds
test_platform_builds() {
    local platform="$1"
    local test_chars="$2"
    
    log "Testing builds for platform: $platform"
    
    local goos="${platform%/*}"
    local goarch="${platform#*/}"
    local build_errors=0
    
    for char in $test_chars; do
        log "Building $char for $platform"
        
        # Use the build script to test the actual build process
        if PLATFORMS="$platform" "$SCRIPT_DIR/build-characters.sh" build "$char" \
            >"$VALIDATION_DIR/logs/build-${char}-${goos}-${goarch}.log" 2>&1; then
            
            # Verify binary was created
            local expected_binary
            if [[ "$goos" == "android" ]]; then
                expected_binary="$BUILD_DIR/${char}_android_${goarch}.apk"
            else
                local ext=""
                [[ "$goos" == "windows" ]] && ext=".exe"
                expected_binary="$BUILD_DIR/${char}_${goos}_${goarch}${ext}"
            fi
            
            if [[ -f "$expected_binary" ]]; then
                success "Build successful: $char → $platform"
                
                # Store artifact for later validation
                cp "$expected_binary" "$VALIDATION_DIR/artifacts/" 2>/dev/null || true
            else
                error "Build claimed success but binary not found: $expected_binary"
                ((build_errors++))
            fi
        else
            error "Build failed: $char → $platform"
            ((build_errors++))
        fi
    done
    
    return $build_errors
}

# Test Android APK builds specifically
test_android_builds() {
    if [[ "$ANDROID_TESTS" != "true" ]]; then
        warning "Android tests skipped (fyne CLI not available)"
        return 0
    fi
    
    log "Testing Android APK builds"
    
    local test_chars
    test_chars=$(get_test_characters | head -2)  # Test with 2 characters
    
    # Test both ARM architectures
    local platforms=("android/arm64" "android/arm")
    local total_errors=0
    
    for platform in "${platforms[@]}"; do
        local errors
        errors=$(test_platform_builds "$platform" "$test_chars")
        ((total_errors += errors))
    done
    
    if [[ $total_errors -eq 0 ]]; then
        success "All Android builds completed successfully"
    else
        error "$total_errors Android build failures"
    fi
    
    return $total_errors
}

# Test cross-platform builds
test_cross_platform_builds() {
    log "Testing cross-platform builds"
    
    local current_os
    current_os=$(go env GOOS)
    
    # Test current platform first (should always work)
    local current_platform="${current_os}/$(go env GOARCH)"
    local test_chars
    test_chars=$(get_test_characters | head -2)
    
    log "Testing native platform: $current_platform"
    local native_errors
    native_errors=$(test_platform_builds "$current_platform" "$test_chars")
    
    if [[ $native_errors -eq 0 ]]; then
        success "Native platform builds successful"
    else
        error "$native_errors native platform build failures"
        return $native_errors
    fi
    
    # For cross-platform, just validate the detection logic works
    # (actual cross-compilation may fail due to CGO/Fyne constraints)
    local cross_platforms=()
    case "$current_os" in
        "linux")
            cross_platforms=("windows/amd64" "darwin/amd64")
            ;;
        "darwin")
            cross_platforms=("linux/amd64" "windows/amd64")
            ;;
        "windows")
            cross_platforms=("linux/amd64" "darwin/amd64")
            ;;
    esac
    
    for platform in "${cross_platforms[@]}"; do
        log "Validating cross-compilation detection for: $platform"
        
        # Test the validation function (should warn about cross-compilation)
        if PLATFORMS="$platform" "$SCRIPT_DIR/build-characters.sh" platforms \
            >"$VALIDATION_DIR/logs/cross-platform-${platform//\//-}.log" 2>&1; then
            success "Cross-platform validation works for: $platform"
        else
            warning "Cross-platform validation issues for: $platform"
        fi
    done
    
    return 0
}

# Test artifact management
test_artifact_management() {
    log "Testing artifact management system"
    
    # Check if artifact manager exists and builds
    if [[ ! -f "$PROJECT_ROOT/cmd/artifact-manager/main.go" ]]; then
        error "Artifact manager not found at cmd/artifact-manager/main.go"
        return 1
    fi
    
    # Build artifact manager
    if go build -o "$TEMP_DIR/artifact-manager" "$PROJECT_ROOT/cmd/artifact-manager/main.go" \
        2>"$VALIDATION_DIR/logs/artifact-manager-build.log"; then
        success "Artifact manager builds successfully"
    else
        error "Artifact manager failed to build"
        return 1
    fi
    
    # Test artifact manager commands
    local test_artifact="$VALIDATION_DIR/artifacts/test_binary"
    echo "test binary content" > "$test_artifact"
    
    # Test storing an artifact
    if "$TEMP_DIR/artifact-manager" -dir "$TEMP_DIR/test-artifacts" \
        store "test" "linux" "amd64" "$test_artifact" \
        >"$VALIDATION_DIR/logs/artifact-store.log" 2>&1; then
        success "Artifact storage test passed"
    else
        error "Artifact storage test failed"
        return 1
    fi
    
    # Test listing artifacts
    if "$TEMP_DIR/artifact-manager" -dir "$TEMP_DIR/test-artifacts" \
        list >"$VALIDATION_DIR/logs/artifact-list.log" 2>&1; then
        success "Artifact listing test passed"
    else
        error "Artifact listing test failed"
        return 1
    fi
    
    # Test statistics
    if "$TEMP_DIR/artifact-manager" -dir "$TEMP_DIR/test-artifacts" \
        stats >"$VALIDATION_DIR/logs/artifact-stats.log" 2>&1; then
        success "Artifact statistics test passed"
    else
        error "Artifact statistics test failed"
        return 1
    fi
    
    return 0
}

# Validate GitHub Actions workflow syntax
test_github_actions_workflow() {
    log "Validating GitHub Actions workflow syntax"
    
    local workflow_file="$PROJECT_ROOT/.github/workflows/build-character-binaries.yml"
    
    if [[ ! -f "$workflow_file" ]]; then
        error "GitHub Actions workflow not found"
        return 1
    fi
    
    # Check YAML syntax using Python (available in most environments)
    if command -v python3 &> /dev/null; then
        if python3 -c "import yaml; yaml.safe_load(open('$workflow_file'))" 2>/dev/null; then
            success "GitHub Actions workflow YAML syntax valid"
        else
            error "GitHub Actions workflow YAML syntax invalid"
            return 1
        fi
    elif command -v yq &> /dev/null; then
        if yq eval '.' "$workflow_file" >/dev/null 2>&1; then
            success "GitHub Actions workflow YAML syntax valid"
        else
            error "GitHub Actions workflow YAML syntax invalid"
            return 1
        fi
    else
        warning "Cannot validate YAML syntax (no python3 or yq available)"
    fi
    
    # Check for required workflow components
    local required_jobs=("generate-matrix" "build-binaries" "package-releases")
    
    for job in "${required_jobs[@]}"; do
        if grep -q "^  $job:" "$workflow_file"; then
            success "Required job present: $job"
        else
            error "Required job missing: $job"
            return 1
        fi
    done
    
    # Check for Android platform support
    if grep -q "android" "$workflow_file"; then
        success "Android platform support present in workflow"
    else
        error "Android platform support missing from workflow"
        return 1
    fi
    
    return 0
}

# Generate validation report
generate_report() {
    local report_file="$VALIDATION_DIR/validation-report.md"
    
    log "Generating validation report"
    
    cat > "$report_file" << EOF
# Pipeline Validation Report

**Generated:** $(date -u '+%Y-%m-%d %H:%M:%S UTC')
**Project:** Desktop Companion (DDS) Character Binary Pipeline
**Validation Script:** validate-pipeline.sh

## Summary

- ✅ Tests Passed: $TESTS_PASSED
- ❌ Tests Failed: $TESTS_FAILED
- ⚠️  Tests Skipped: $TESTS_SKIPPED

## Test Results

### Environment Validation
$(if [[ -f "$VALIDATION_DIR/logs/environment.log" ]]; then echo "✅ Environment validation completed"; else echo "❌ Environment validation issues"; fi)

### Character Card Validation
$(if [[ -f "$VALIDATION_DIR/logs/character-cards.log" ]]; then echo "✅ Character cards validated"; else echo "❌ Character card validation issues"; fi)

### Build System Tests
$(if [[ -f "$VALIDATION_DIR/logs/embed-test.log" ]]; then echo "✅ Embed script tested"; else echo "❌ Embed script issues"; fi)

### Platform Build Tests
$(ls "$VALIDATION_DIR/logs"/build-*.log 2>/dev/null | wc -l) build tests executed

### Android APK Tests
$(if [[ "$ANDROID_TESTS" == "true" ]]; then echo "✅ Android tests enabled"; else echo "⚠️ Android tests skipped"; fi)

### Artifact Management Tests
$(if [[ -f "$VALIDATION_DIR/logs/artifact-store.log" ]]; then echo "✅ Artifact management tested"; else echo "❌ Artifact management issues"; fi)

### GitHub Actions Workflow
$(if [[ -f "$VALIDATION_DIR/logs/workflow.log" ]]; then echo "✅ Workflow validated"; else echo "❌ Workflow validation issues"; fi)

## Artifacts Generated

$(find "$VALIDATION_DIR/artifacts" -type f 2>/dev/null | wc -l) test artifacts generated

## Logs Available

$(find "$VALIDATION_DIR/logs" -name "*.log" -exec basename {} \; | sort)

## Recommendations

EOF

    if [[ $TESTS_FAILED -eq 0 ]]; then
        echo "✅ All critical tests passed. Pipeline is ready for production use." >> "$report_file"
    else
        echo "❌ $TESTS_FAILED tests failed. Review logs and address issues before deployment." >> "$report_file"
    fi
    
    if [[ "$ANDROID_TESTS" != "true" ]]; then
        echo "" >> "$report_file"
        echo "⚠️ Android tests were skipped. Install fyne CLI tool to enable full testing:" >> "$report_file"
        echo "   \`go install fyne.io/tools/cmd/fyne@latest\`" >> "$report_file"
    fi
    
    success "Validation report generated: $report_file"
}

# Main validation runner
run_validation() {
    local start_time
    start_time=$(date +%s)
    
    echo "======================================"
    echo "DDS Pipeline Validation Suite"
    echo "======================================"
    echo
    
    init_validation
    
    log "Running validation tests..."
    
    # Core validation tests (run sequentially for clear output)
    validate_environment
    test_character_cards
    test_embed_script
    test_artifact_management
    test_github_actions_workflow
    
    # Platform-specific tests
    test_cross_platform_builds
    test_android_builds
    
    # Generate final report
    generate_report
    
    local end_time
    end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    echo
    echo "======================================"
    echo "Validation Complete"
    echo "======================================"
    echo "Duration: ${duration}s"
    echo "Results: $TESTS_PASSED passed, $TESTS_FAILED failed, $TESTS_SKIPPED skipped"
    echo "Report: $VALIDATION_DIR/validation-report.md"
    echo
    
    if [[ $TESTS_FAILED -eq 0 ]]; then
        success "Pipeline validation successful! ✅"
        return 0
    else
        error "Pipeline validation failed with $TESTS_FAILED errors ❌"
        return 1
    fi
}

# Script entry point
main() {
    case "${1:-validate}" in
        "validate"|"")
            run_validation
            ;;
        "quick")
            QUICK_MODE=true
            run_validation
            ;;
        "no-android")
            ANDROID_TESTS=false
            run_validation
            ;;
        "environment")
            init_validation
            validate_environment
            if [[ $? -eq 0 ]]; then
                success "Environment validation completed successfully"
            else
                error "Environment validation failed"
                exit 1
            fi
            ;;
        "help")
            echo "Usage: $0 [validate|quick|no-android|environment|help]"
            echo
            echo "Commands:"
            echo "  validate    - Run full pipeline validation (default)"
            echo "  quick       - Run quick validation with subset of characters"
            echo "  no-android  - Skip Android APK tests"
            echo "  environment - Validate environment only"
            echo "  help        - Show this help"
            echo
            echo "Environment variables:"
            echo "  QUICK_MODE=true     - Enable quick testing mode"
            echo "  ANDROID_TESTS=false - Disable Android testing"
            echo "  PARALLEL_TESTS=N    - Set parallel test workers"
            ;;
        *)
            error "Unknown command: $1"
            echo "Use '$0 help' for usage information"
            exit 1
            ;;
    esac
}

# Execute main function with all arguments
main "$@"
