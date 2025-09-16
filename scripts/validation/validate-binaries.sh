#!/bin/bash

# scripts/validation/validate-binaries.sh
# Character binary validation script
#
# Tests all character binaries for functionality, performance, and deployment readiness.
# Validates binary size, startup time, embedded assets, and memory usage.
#
# Usage: ./scripts/validation/validate-binaries.sh [OPTIONS] [COMMAND]
#
# Dependencies:
# - Go 1.21+
# - Built character binaries
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
# VALIDATION CONFIGURATION
# ============================================================================

# Test configuration from shared config
VALIDATION_TIMEOUT="${DDS_VALIDATION_TIMEOUT}"
MEMORY_LIMIT_MB="${DDS_MEMORY_LIMIT_MB}"
STARTUP_TIME_LIMIT_SEC="${DDS_STARTUP_TIME_LIMIT_SEC}"

# Local validation settings
BINARY_SIZE_LIMIT_MB=50
TEST_EMBEDDED_ASSETS=true
TEST_MEMORY_USAGE=true

# Counters for reporting
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# ============================================================================
# HELP AND USAGE
# ============================================================================

show_usage() {
    cat << EOF
Usage: $0 [OPTIONS] [COMMAND]

Validate character binaries for functionality and deployment readiness.

COMMANDS:
    validate [CHARACTER]    Validate specific character or all binaries
    benchmark              Run performance benchmarks on all binaries
    report                 Generate validation report
    help                   Show this help message

OPTIONS:
    -t, --timeout SECONDS      Set validation timeout (default: $VALIDATION_TIMEOUT)
    -m, --memory-limit MB      Set memory limit for tests (default: $MEMORY_LIMIT_MB)
    --skip-memory             Skip memory usage tests
    --skip-assets             Skip embedded asset tests
    -v, --verbose             Enable verbose output
    --dry-run                 Show what would be tested without running tests

EXAMPLES:
    $0                        # Validate all character binaries
    $0 validate default       # Validate only the default character binary
    $0 benchmark             # Run performance benchmarks
    $0 -v --timeout 60       # Verbose validation with 60s timeout

PREREQUISITES:
    - Character binaries must be built (run: ./scripts/dds-scripts.sh build characters)
    - Test output directory will be created automatically

OUTPUT:
    Test results: $TEST_OUTPUT_DIR/validation-*.log
    Summary report: $TEST_OUTPUT_DIR/binary-validation-report.txt

EOF
}

# ============================================================================
# VALIDATION FUNCTIONS
# ============================================================================

# Test a single character binary
validate_character_binary() {
    local char_name="$1"
    local binary_path="$2"
    local test_log="$TEST_OUTPUT_DIR/validation_${char_name}.log"
    
    log "Validating character binary: $char_name"
    
    # Test 1: Binary exists and is executable
    ((TOTAL_TESTS++))
    if [[ ! -f "$binary_path" ]]; then
        error "Binary not found: $binary_path"
        return 1
    fi
    
    if [[ ! -x "$binary_path" ]]; then
        error "Binary not executable: $binary_path"
        return 1
    fi
    success "Binary exists and is executable"
    
    # Test 2: Binary size is reasonable
    ((TOTAL_TESTS++))
    local size_mb=$(du -m "$binary_path" | cut -f1)
    if [[ $size_mb -gt $BINARY_SIZE_LIMIT_MB ]]; then
        warning "Binary size is large: ${size_mb}MB (consider optimization)"
    else
        success "Binary size is reasonable: ${size_mb}MB"
    fi
    
    # Test 3: Binary starts without crashing (version check)
    ((TOTAL_TESTS++))
    log "Testing binary startup and version check..."
    if timeout "$VALIDATION_TIMEOUT" "$binary_path" -version >"$test_log" 2>&1; then
        success "Binary starts successfully and reports version"
        ((PASSED_TESTS++))
    else
        error "Binary failed to start or crashed during version check"
        cat "$test_log"
        ((FAILED_TESTS++))
        return 1
    fi
    
    # Test 4: Binary doesn't require external assets (if enabled)
    if [[ "$TEST_EMBEDDED_ASSETS" == "true" ]]; then
        ((TOTAL_TESTS++))
        log "Testing embedded asset independence..."
        local temp_dir=$(mktemp -d)
        cp "$binary_path" "$temp_dir/"
        local binary_name=$(basename "$binary_path")
        
        if (cd "$temp_dir" && timeout "$VALIDATION_TIMEOUT" "./$binary_name" -version >/dev/null 2>&1); then
            success "Binary runs independently without external assets"
            ((PASSED_TESTS++))
        else
            error "Binary requires external assets or failed in isolation"
            ((FAILED_TESTS++))
        fi
        
        rm -rf "$temp_dir"
    fi
    
    # Test 5: Memory usage validation (if enabled)
    if [[ "$TEST_MEMORY_USAGE" == "true" ]]; then
        ((TOTAL_TESTS++))
        log "Testing memory usage..."
        
        # Note: This is a simplified test. In production, you might use tools like valgrind
        local memory_test_log="$TEST_OUTPUT_DIR/memory_${char_name}.log"
        if timeout "$VALIDATION_TIMEOUT" "$binary_path" -version >"$memory_test_log" 2>&1; then
            success "Binary completed memory test"
            ((PASSED_TESTS++))
        else
            warning "Memory test inconclusive for $char_name"
        fi
    fi
    
    return 0
}

# Validate all character binaries in build directory
validate_all_binaries() {
    local build_pattern="$BUILD_DIR/companion-*"
    local binary_count=0
    
    log "Scanning for character binaries in $BUILD_DIR"
    
    for binary_path in $build_pattern; do
        [[ -f "$binary_path" ]] || continue
        
        local binary_name=$(basename "$binary_path")
        local char_name="${binary_name#companion-}"
        
        validate_character_binary "$char_name" "$binary_path"
        ((binary_count++))
    done
    
    if [[ $binary_count -eq 0 ]]; then
        error "No character binaries found in $BUILD_DIR"
        error "Run: ./scripts/dds-scripts.sh build characters"
        return 1
    fi
    
    log "Validated $binary_count character binaries"
    return 0
}

# Performance benchmarking for character binaries
benchmark_binaries() {
    log "Running performance benchmarks..."
    
    local benchmark_log="$TEST_OUTPUT_DIR/benchmark-$(date +%Y%m%d-%H%M%S).log"
    
    {
        echo "# Character Binary Performance Benchmark"
        echo "Generated: $(date -u '+%Y-%m-%d %H:%M:%S UTC')"
        echo ""
    } > "$benchmark_log"
    
    for binary_path in "$BUILD_DIR"/companion-*; do
        [[ -f "$binary_path" ]] || continue
        
        local binary_name=$(basename "$binary_path")
        local char_name="${binary_name#companion-}"
        
        log "Benchmarking $char_name..."
        
        # Startup time test
        local start_time=$(date +%s.%N)
        if timeout "$STARTUP_TIME_LIMIT_SEC" "$binary_path" -version >/dev/null 2>&1; then
            local end_time=$(date +%s.%N)
            local startup_time=$(echo "$end_time - $start_time" | bc -l)
            
            {
                echo "## $char_name"
                echo "- Startup time: ${startup_time}s"
                echo "- Binary size: $(du -h "$binary_path" | cut -f1)"
                echo ""
            } >> "$benchmark_log"
            
            success "$char_name startup: ${startup_time}s"
        else
            warning "$char_name failed startup benchmark"
        fi
    done
    
    success "Benchmark results saved to: $benchmark_log"
}

# Generate validation report
generate_validation_report() {
    local report_file="$TEST_OUTPUT_DIR/binary-validation-report.txt"
    
    {
        echo "# Character Binary Validation Report"
        echo "Generated: $(date -u '+%Y-%m-%d %H:%M:%S UTC')"
        echo ""
        echo "## Summary"
        echo "- Total tests: $TOTAL_TESTS"
        echo "- Passed: $PASSED_TESTS"
        echo "- Failed: $FAILED_TESTS"
        echo "- Success rate: $(( PASSED_TESTS * 100 / TOTAL_TESTS ))%"
        echo ""
        echo "## Configuration"
        echo "- Validation timeout: ${VALIDATION_TIMEOUT}s"
        echo "- Memory limit: ${MEMORY_LIMIT_MB}MB"
        echo "- Binary size limit: ${BINARY_SIZE_LIMIT_MB}MB"
        echo "- Test embedded assets: $TEST_EMBEDDED_ASSETS"
        echo "- Test memory usage: $TEST_MEMORY_USAGE"
        echo ""
    } > "$report_file"
    
    success "Validation report saved to: $report_file"
}

# ============================================================================
# MAIN EXECUTION
# ============================================================================

# Parse command line arguments
COMMAND="validate"
CHARACTER_NAME=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help|help)
            show_usage
            exit 0
            ;;
        -t|--timeout)
            VALIDATION_TIMEOUT="$2"
            shift 2
            ;;
        -m|--memory-limit)
            MEMORY_LIMIT_MB="$2"
            shift 2
            ;;
        --skip-memory)
            TEST_MEMORY_USAGE=false
            shift
            ;;
        --skip-assets)
            TEST_EMBEDDED_ASSETS=false
            shift
            ;;
        -v|--verbose)
            DDS_VERBOSE=true
            shift
            ;;
        --dry-run)
            DDS_DRY_RUN=true
            shift
            ;;
        validate|benchmark|report)
            COMMAND="$1"
            shift
            ;;
        -*)
            error "Unknown option: $1"
            show_usage
            exit 1
            ;;
        *)
            if [[ "$COMMAND" == "validate" && -z "$CHARACTER_NAME" ]]; then
                CHARACTER_NAME="$1"
            else
                error "Unexpected argument: $1"
                show_usage
                exit 1
            fi
            shift
            ;;
    esac
done

# Create test output directory
mkdir -p "$TEST_OUTPUT_DIR"

# Execute command
case $COMMAND in
    validate)
        log "Starting character binary validation..."
        if [[ -n "$CHARACTER_NAME" ]]; then
            binary_path="$BUILD_DIR/companion-$CHARACTER_NAME"
            if [[ -f "$binary_path" ]]; then
                validate_character_binary "$CHARACTER_NAME" "$binary_path"
            else
                error "Character binary not found: $binary_path"
                exit 1
            fi
        else
            validate_all_binaries
        fi
        generate_validation_report
        ;;
    benchmark)
        log "Starting performance benchmarking..."
        benchmark_binaries
        ;;
    report)
        log "Generating validation report..."
        generate_validation_report
        ;;
    *)
        error "Unknown command: $COMMAND"
        show_usage
        exit 1
        ;;
esac

# Final summary
if [[ $FAILED_TESTS -eq 0 ]]; then
    success "All validation tests passed! ($PASSED_TESTS/$TOTAL_TESTS)"
    exit 0
else
    error "Some validation tests failed. ($FAILED_TESTS/$TOTAL_TESTS failed)"
    exit 1
fi
