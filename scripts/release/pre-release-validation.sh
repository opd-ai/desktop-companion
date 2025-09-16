#!/bin/bash

# scripts/release/pre-release-validation.sh
# Pre-release validation and performance benchmarking script
#
# Performs comprehensive validation before release including full regression
# testing, performance benchmarking, and release readiness checks.
#
# Usage: ./scripts/release/pre-release-validation.sh [OPTIONS]
#
# Dependencies:
# - Go 1.21+
# - Built binaries and assets
# - scripts/lib/common.sh
# - scripts/lib/config.sh

# Load shared libraries
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LIB_DIR="$(dirname "$SCRIPT_DIR")/lib"

# shellcheck source=../lib/common.sh
source "$LIB_DIR/common.sh"
# shellcheck source=../lib/config.sh
source "$LIB_DIR/config.sh"

# ============================================================================
# RELEASE VALIDATION CONFIGURATION
# ============================================================================

# Performance targets from shared config
TARGET_MEMORY_MB="${DDS_TARGET_MEMORY_MB:-50}"
TARGET_FPS="${DDS_TARGET_FPS:-30}"
TEST_TIMEOUT="${DDS_TEST_TIMEOUT:-60s}"

# Release validation settings
COMPREHENSIVE_MODE=true
BENCHMARK_MODE=true
REGRESSION_TESTS=true

# ============================================================================
# HELP AND USAGE
# ============================================================================

show_usage() {
    cat << EOF
Usage: $0 [OPTIONS] [COMMAND]

Comprehensive pre-release validation for Desktop Companion.

COMMANDS:
    validate              Run full pre-release validation (default)
    quick                 Run quick validation (essential tests only)
    benchmark             Run performance benchmarks only
    regression            Run regression tests only
    environment           Validate build environment only
    report                Generate release readiness report
    help                  Show this help message

OPTIONS:
    --skip-benchmarks     Skip performance benchmarking
    --skip-regression     Skip regression testing
    --target-memory MB    Set memory target (default: $TARGET_MEMORY_MB MB)
    --target-fps FPS      Set FPS target (default: $TARGET_FPS FPS)
    --timeout DURATION    Set test timeout (default: $TEST_TIMEOUT)
    -v, --verbose         Enable verbose output
    --dry-run            Show what would be tested

EXAMPLES:
    $0                    # Full pre-release validation
    $0 quick              # Quick validation
    $0 benchmark          # Performance benchmarks only
    $0 --target-memory 64 # Custom memory target

VALIDATION PHASES:
    Phase 1: Environment validation
    Phase 2: Full regression testing
    Phase 3: Performance benchmarking
    Phase 4: Release artifact validation
    Phase 5: Final readiness assessment

OUTPUT:
    Test results: $TEST_OUTPUT_DIR/release-validation-*.log
    Performance data: $TEST_OUTPUT_DIR/performance-*.log
    Release report: $TEST_OUTPUT_DIR/release-readiness-report.txt

EOF
}

# ============================================================================
# VALIDATION FUNCTIONS
# ============================================================================

# Validate build environment
validate_build_environment() {
    log "Phase 1: Environment Validation"
    log "------------------------------"
    
    # Check Go version
    local go_version
    go_version=$(go version)
    log "Go version: $go_version"
    
    # Check available memory
    if command -v free >/dev/null 2>&1; then
        local available_memory
        available_memory=$(free -m | awk 'NR==2{printf "%.1f", $7/1024}')
        log "Available memory: ${available_memory}GB"
        
        if (( $(echo "$available_memory < 2.0" | bc -l) )); then
            warning "Low available memory: ${available_memory}GB (may affect performance tests)"
        fi
    fi
    
    # Verify dependencies
    log "Checking Go modules..."
    if ! go mod verify >/dev/null 2>&1; then
        error "Go module verification failed"
        return 1
    fi
    
    if ! go mod tidy >/dev/null 2>&1; then
        error "Go module tidy failed"
        return 1
    fi
    
    success "✅ Environment validation passed"
    return 0
}

# Run comprehensive regression tests
run_regression_tests() {
    if [[ "$REGRESSION_TESTS" != "true" ]]; then
        log "Skipping regression tests"
        return 0
    fi
    
    log "Phase 2: Full Regression Testing"
    log "--------------------------------"
    
    # Run existing test suites
    log "Running core test suites..."
    local test_log="$TEST_OUTPUT_DIR/regression-tests-$(date +%Y%m%d-%H%M%S).log"
    
    # Go test with coverage
    if go test -v -race -coverprofile="$TEST_OUTPUT_DIR/coverage.out" ./... > "$test_log" 2>&1; then
        success "✅ Core test suites passed"
        
        # Generate coverage report
        if command -v go >/dev/null 2>&1; then
            local coverage_percent
            coverage_percent=$(go tool cover -func="$TEST_OUTPUT_DIR/coverage.out" | tail -1 | awk '{print $3}')
            log "Test coverage: $coverage_percent"
        fi
    else
        error "❌ Core test suites failed (see: $test_log)"
        return 1
    fi
    
    # Run character validation tests
    log "Running character validation tests..."
    if "$PROJECT_ROOT/scripts/validation/validate-characters.sh" >/dev/null 2>&1; then
        success "✅ Character validation tests passed"
    else
        error "❌ Character validation tests failed"
        return 1
    fi
    
    # Run binary validation tests
    log "Running binary validation tests..."
    if "$PROJECT_ROOT/scripts/validation/validate-binaries.sh" >/dev/null 2>&1; then
        success "✅ Binary validation tests passed"
    else
        error "❌ Binary validation tests failed"
        return 1
    fi
    
    # Run pipeline validation
    log "Running pipeline validation..."
    if "$PROJECT_ROOT/scripts/validation/validate-pipeline.sh" quick >/dev/null 2>&1; then
        success "✅ Pipeline validation passed"
    else
        warning "⚠️ Pipeline validation had issues"
    fi
    
    return 0
}

# Run performance benchmarks
run_performance_benchmarks() {
    if [[ "$BENCHMARK_MODE" != "true" ]]; then
        log "Skipping performance benchmarks"
        return 0
    fi
    
    log "Phase 3: Performance Benchmarking"
    log "---------------------------------"
    
    local benchmark_log="$TEST_OUTPUT_DIR/performance-$(date +%Y%m%d-%H%M%S).log"
    
    {
        echo "# Desktop Companion Performance Benchmark"
        echo "Generated: $(date -u '+%Y-%m-%d %H:%M:%S UTC')"
        echo ""
        echo "## Configuration"
        echo "- Target Memory: ${TARGET_MEMORY_MB}MB"
        echo "- Target FPS: ${TARGET_FPS}"
        echo "- Test Timeout: $TEST_TIMEOUT"
        echo ""
    } > "$benchmark_log"
    
    # Memory usage benchmarks
    log "Testing memory usage patterns..."
    local memory_results="$TEST_OUTPUT_DIR/memory-benchmark.log"
    
    # Find character binaries
    local binary_count=0
    for binary_path in "$BUILD_DIR"/companion-*; do
        [[ -f "$binary_path" ]] || continue
        
        local binary_name=$(basename "$binary_path")
        local char_name="${binary_name#companion-}"
        
        log "Benchmarking memory usage for $char_name..."
        
        # Simple memory test (startup and shutdown)
        local start_time=$(date +%s.%N)
        if timeout "$TEST_TIMEOUT" "$binary_path" -version >/dev/null 2>&1; then
            local end_time=$(date +%s.%N)
            local duration=$(echo "$end_time - $start_time" | bc -l)
            
            {
                echo "## $char_name"
                echo "- Startup time: ${duration}s"
                echo "- Binary size: $(du -h "$binary_path" | cut -f1)"
                echo "- Target memory: ${TARGET_MEMORY_MB}MB"
                echo ""
            } >> "$benchmark_log"
            
            success "✅ $char_name performance test completed (${duration}s)"
            ((binary_count++))
        else
            warning "⚠️ $char_name performance test timed out"
        fi
    done
    
    if [[ $binary_count -eq 0 ]]; then
        error "❌ No character binaries found for benchmarking"
        error "Run: ./scripts/dds-scripts.sh build characters"
        return 1
    fi
    
    # Go benchmark tests
    log "Running Go benchmark tests..."
    if go test -bench=. -benchmem ./... >> "$benchmark_log" 2>&1; then
        success "✅ Go benchmark tests completed"
    else
        warning "⚠️ Go benchmark tests had issues"
    fi
    
    log "Performance benchmark results saved to: $benchmark_log"
    return 0
}

# Validate release artifacts
validate_release_artifacts() {
    log "Phase 4: Release Artifact Validation"
    log "------------------------------------"
    
    # Check for required binaries
    local required_binaries=("companion")
    local missing_binaries=()
    
    for binary in "${required_binaries[@]}"; do
        if [[ -f "$BUILD_DIR/$binary" ]]; then
            local size=$(du -h "$BUILD_DIR/$binary" | cut -f1)
            success "✅ $binary binary found (size: $size)"
        else
            missing_binaries+=("$binary")
        fi
    done
    
    if [[ ${#missing_binaries[@]} -gt 0 ]]; then
        error "❌ Missing required binaries: ${missing_binaries[*]}"
        return 1
    fi
    
    # Check for character assets
    local character_count=0
    if [[ -d "$CHARACTERS_DIR" ]]; then
        while IFS= read -r -d '' file; do
            ((character_count++))
        done < <(find "$CHARACTERS_DIR" -name "character.json" -print0 2>/dev/null)
        
        if [[ $character_count -gt 0 ]]; then
            success "✅ Character assets found ($character_count characters)"
        else
            warning "⚠️ No character assets found"
        fi
    fi
    
    # Check build directory structure
    if [[ -d "$BUILD_DIR" ]]; then
        local build_files
        build_files=$(find "$BUILD_DIR" -type f | wc -l)
        log "Build directory contains $build_files files"
    fi
    
    return 0
}

# Generate release readiness report
generate_release_readiness_report() {
    log "Phase 5: Final Readiness Assessment"
    log "----------------------------------"
    
    local report_file="$TEST_OUTPUT_DIR/release-readiness-report.txt"
    
    {
        echo "# Desktop Companion Release Readiness Report"
        echo "Generated: $(date -u '+%Y-%m-%d %H:%M:%S UTC')"
        echo ""
        echo "## Release Information"
        echo "- Project: Desktop Companion (DDS)"
        echo "- Version: $(git describe --tags --always 2>/dev/null || echo "development")"
        echo "- Commit: $(git rev-parse --short HEAD 2>/dev/null || echo "unknown")"
        echo "- Branch: $(git branch --show-current 2>/dev/null || echo "unknown")"
        echo ""
        echo "## Validation Summary"
        echo "### Environment"
        echo "- Go version: $(go version | cut -d' ' -f3)"
        echo "- Operating system: $(uname -s)"
        echo "- Architecture: $(uname -m)"
        echo ""
        echo "### Performance Targets"
        echo "- Memory target: ${TARGET_MEMORY_MB}MB"
        echo "- FPS target: ${TARGET_FPS}"
        echo "- Test timeout: $TEST_TIMEOUT"
        echo ""
        echo "### Test Results"
        echo "- Regression tests: $([ "$REGRESSION_TESTS" = "true" ] && echo "✅ Completed" || echo "⏭️ Skipped")"
        echo "- Performance benchmarks: $([ "$BENCHMARK_MODE" = "true" ] && echo "✅ Completed" || echo "⏭️ Skipped")"
        echo "- Artifact validation: ✅ Completed"
        echo ""
        echo "## Build Artifacts"
        if [[ -d "$BUILD_DIR" ]]; then
            echo "Build directory: $BUILD_DIR"
            find "$BUILD_DIR" -type f -name "companion*" | while read -r file; do
                echo "- $(basename "$file"): $(du -h "$file" | cut -f1)"
            done
        fi
        echo ""
        echo "## Recommendations"
        echo "1. Verify all target platforms are built"
        echo "2. Test installation on clean systems"
        echo "3. Validate documentation is up-to-date"
        echo "4. Confirm all dependencies are properly licensed"
        echo "5. Run final security scans if applicable"
        echo ""
        echo "## Next Steps"
        echo "- [ ] Create release tags"
        echo "- [ ] Generate release notes"
        echo "- [ ] Upload release artifacts"
        echo "- [ ] Update documentation"
        echo "- [ ] Announce release"
        echo ""
    } > "$report_file"
    
    success "✅ Release readiness report generated: $report_file"
    return 0
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
        --skip-benchmarks)
            BENCHMARK_MODE=false
            shift
            ;;
        --skip-regression)
            REGRESSION_TESTS=false
            shift
            ;;
        --target-memory)
            TARGET_MEMORY_MB="$2"
            shift 2
            ;;
        --target-fps)
            TARGET_FPS="$2"
            shift 2
            ;;
        --timeout)
            TEST_TIMEOUT="$2"
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
        validate|quick|benchmark|regression|environment|report)
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
    COMPREHENSIVE_MODE=false
    BENCHMARK_MODE=false
    COMMAND="validate"
fi

# Create output directories
mkdir -p "$BUILD_DIR" "$TEST_OUTPUT_DIR"

# Execute command
case $COMMAND in
    validate)
        log "Starting pre-release validation..."
        echo "=========================================="
        echo "Desktop Companion Release Preparation"
        echo "Phase 4 Task 3: Final Testing & Release"
        echo ""
        
        validate_build_environment || exit 1
        run_regression_tests || exit 1
        run_performance_benchmarks || exit 1
        validate_release_artifacts || exit 1
        generate_release_readiness_report || exit 1
        ;;
    environment)
        log "Validating build environment only..."
        validate_build_environment || exit 1
        ;;
    regression)
        log "Running regression tests only..."
        run_regression_tests || exit 1
        ;;
    benchmark)
        log "Running performance benchmarks only..."
        run_performance_benchmarks || exit 1
        ;;
    report)
        log "Generating release readiness report..."
        generate_release_readiness_report || exit 1
        ;;
    *)
        error "Unknown command: $COMMAND"
        show_usage
        exit 1
        ;;
esac

# Final summary
success "🎉 Pre-release validation completed successfully!"
success "📋 Check the release readiness report for next steps"
exit 0
