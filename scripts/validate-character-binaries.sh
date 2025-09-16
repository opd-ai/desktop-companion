#!/bin/bash

# DEPRECATED: Legacy wrapper for validate-character-binaries.sh
# This script is maintained for backward compatibility.
# New usage: ./scripts/dds-scripts.sh validation binaries
# Direct usage: ./scripts/validation/validate-binaries.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Forward all arguments to the new refactored script
exec "$SCRIPT_DIR/validation/validate-binaries.sh" "$@"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
VALIDATION_TIMEOUT=30
MEMORY_LIMIT_MB=100
STARTUP_TIME_LIMIT_SEC=5

# Counters for reporting
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Print colored output
log() {
    echo -e "${BLUE}[$(date +'%H:%M:%S')]${NC} $1"
}

success() {
    echo -e "${GREEN}✓${NC} $1"
    ((PASSED_TESTS++))
}

warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

error() {
    echo -e "${RED}✗${NC} $1" >&2
    ((FAILED_TESTS++))
}

# Create test output directory
mkdir -p "$TEST_OUTPUT_DIR"

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
    
    # Test 2: Binary size is reasonable (should be <50MB for embedded assets)
    ((TOTAL_TESTS++))
    local size_mb=$(du -m "$binary_path" | cut -f1)
    if [[ $size_mb -gt 50 ]]; then
        warning "Binary size is large: ${size_mb}MB (consider optimization)"
    else
        success "Binary size is reasonable: ${size_mb}MB"
    fi
    
    # Test 3: Binary starts without crashing (version check)
    ((TOTAL_TESTS++))
    log "Testing binary startup and version check..."
    if timeout $VALIDATION_TIMEOUT "$binary_path" -version >"$test_log" 2>&1; then
        success "Binary starts successfully and reports version"
    else
        error "Binary failed to start or crashed during version check"
        cat "$test_log"
        return 1
    fi
    
    # Test 4: Binary doesn't require external assets
    ((TOTAL_TESTS++))
    log "Testing embedded asset independence..."
    local temp_dir=$(mktemp -d)
    cp "$binary_path" "$temp_dir/"
    local binary_name=$(basename "$binary_path")
    
    # Try to run in isolated environment
    if (cd "$temp_dir" && timeout $VALIDATION_TIMEOUT "./$binary_name" -version >"$test_log" 2>&1); then
        success "Binary runs independently without external assets"
    else
        error "Binary requires external assets or fails in isolation"
        cat "$test_log"
        rm -rf "$temp_dir"
        return 1
    fi
    rm -rf "$temp_dir"
    
    # Test 5: Memory usage validation (if possible)
    ((TOTAL_TESTS++))
    log "Testing memory usage..."
    # Start binary in background with a short timeout and measure memory
    "$binary_path" -version &
    local pid=$!
    sleep 1
    
    if kill -0 $pid 2>/dev/null; then
        # Process is still running, check memory usage
        local memory_kb=$(ps -o rss= -p $pid 2>/dev/null || echo "0")
        local memory_mb=$((memory_kb / 1024))
        
        kill $pid 2>/dev/null || true
        wait $pid 2>/dev/null || true
        
        if [[ $memory_mb -lt $MEMORY_LIMIT_MB ]]; then
            success "Memory usage acceptable: ${memory_mb}MB"
        else
            warning "Memory usage high: ${memory_mb}MB (limit: ${MEMORY_LIMIT_MB}MB)"
        fi
    else
        success "Binary completed quickly (good for version check)"
    fi
    
    log "Validation complete for $char_name"
    return 0
}

# Validate all character binaries in build directory
validate_all_binaries() {
    log "Starting comprehensive binary validation..."
    
    local platform=$(go env GOOS)
    local arch=$(go env GOARCH)
    local ext=""
    
    if [[ "$platform" == "windows" ]]; then
        ext=".exe"
    fi
    
    # Find all character binaries
    local binaries=()
    while IFS= read -r -d '' binary; do
        binaries+=("$binary")
    done < <(find "$BUILD_DIR" -name "*_${platform}_${arch}${ext}" -type f -print0 2>/dev/null)
    
    if [[ ${#binaries[@]} -eq 0 ]]; then
        error "No character binaries found in $BUILD_DIR"
        error "Run 'make build-characters' first to generate binaries"
        return 1
    fi
    
    log "Found ${#binaries[@]} character binaries to validate"
    
    local validation_failures=0
    
    for binary in "${binaries[@]}"; do
        local filename=$(basename "$binary")
        local char_name="${filename%_${platform}_${arch}${ext}}"
        
        echo
        if validate_character_binary "$char_name" "$binary"; then
            success "✓ $char_name validation passed"
        else
            error "✗ $char_name validation failed"
            ((validation_failures++))
        fi
    done
    
    echo
    log "Validation Summary:"
    log "  Total Tests: $TOTAL_TESTS"
    log "  Passed: $PASSED_TESTS"
    log "  Failed: $FAILED_TESTS"
    log "  Binaries Validated: ${#binaries[@]}"
    log "  Validation Failures: $validation_failures"
    
    if [[ $validation_failures -eq 0 && $FAILED_TESTS -eq 0 ]]; then
        success "All character binaries passed validation! 🎉"
        return 0
    else
        error "Some validations failed. Check logs in $TEST_OUTPUT_DIR"
        return 1
    fi
}

# Performance benchmarking for character binaries
benchmark_binaries() {
    log "Running performance benchmarks..."
    
    local platform=$(go env GOOS)
    local arch=$(go env GOARCH)
    local ext=""
    
    if [[ "$platform" == "windows" ]]; then
        ext=".exe"
    fi
    
    local benchmark_log="$TEST_OUTPUT_DIR/benchmark_results.log"
    echo "# Character Binary Performance Benchmark" > "$benchmark_log"
    echo "# Generated: $(date)" >> "$benchmark_log"
    echo "# Platform: ${platform}/${arch}" >> "$benchmark_log"
    echo >> "$benchmark_log"
    
    # Find all character binaries
    local binaries=()
    while IFS= read -r -d '' binary; do
        binaries+=("$binary")
    done < <(find "$BUILD_DIR" -name "*_${platform}_${arch}${ext}" -type f -print0 2>/dev/null)
    
    printf "%-20s %-10s %-15s %-15s\n" "Character" "Size (MB)" "Startup (ms)" "Memory (MB)" | tee -a "$benchmark_log"
    printf "%-20s %-10s %-15s %-15s\n" "----------" "--------" "-----------" "-----------" | tee -a "$benchmark_log"
    
    for binary in "${binaries[@]}"; do
        local filename=$(basename "$binary")
        local char_name="${filename%_${platform}_${arch}${ext}}"
        
        # Measure binary size
        local size_mb=$(du -m "$binary" | cut -f1)
        
        # Measure startup time
        local start_time=$(date +%s%N)
        timeout $STARTUP_TIME_LIMIT_SEC "$binary" -version >/dev/null 2>&1 || true
        local end_time=$(date +%s%N)
        local startup_ms=$(( (end_time - start_time) / 1000000 ))
        
        # Measure memory usage (approximate)
        "$binary" -version &
        local pid=$!
        sleep 0.5
        local memory_kb=$(ps -o rss= -p $pid 2>/dev/null || echo "0")
        local memory_mb=$((memory_kb / 1024))
        kill $pid 2>/dev/null || true
        wait $pid 2>/dev/null || true
        
        printf "%-20s %-10s %-15s %-15s\n" "$char_name" "${size_mb}" "${startup_ms}" "${memory_mb}" | tee -a "$benchmark_log"
    done
    
    echo >> "$benchmark_log"
    success "Benchmark results saved to $benchmark_log"
}

# Show usage information
show_usage() {
    cat << EOF
Usage: $0 [OPTIONS] [COMMAND]

COMMANDS:
  validate          Validate all character binaries (default)
  benchmark         Run performance benchmarks
  help              Show this help message

OPTIONS:
  --timeout SECONDS     Set validation timeout (default: $VALIDATION_TIMEOUT)
  --memory-limit MB     Set memory limit warning threshold (default: $MEMORY_LIMIT_MB)
  --startup-limit SEC   Set startup time limit (default: $STARTUP_TIME_LIMIT_SEC)

EXAMPLES:
  $0                              # Validate all binaries
  $0 validate                     # Same as above
  $0 benchmark                    # Run performance benchmarks
  $0 --timeout 60 validate        # Validate with 60s timeout

PREREQUISITES:
  - Character binaries must be built first: make build-characters
  - Binaries should be in $BUILD_DIR
  - Current platform: $(go env GOOS)/$(go env GOARCH)

EOF
}

# Parse command line arguments
COMMAND="validate"

while [[ $# -gt 0 ]]; do
    case $1 in
        --timeout)
            VALIDATION_TIMEOUT="$2"
            shift 2
            ;;
        --memory-limit)
            MEMORY_LIMIT_MB="$2"
            shift 2
            ;;
        --startup-limit)
            STARTUP_TIME_LIMIT_SEC="$2"
            shift 2
            ;;
        validate|benchmark|help)
            COMMAND="$1"
            shift
            ;;
        *)
            error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Execute command
case $COMMAND in
    validate)
        validate_all_binaries
        ;;
    benchmark)
        benchmark_binaries
        ;;
    help)
        show_usage
        ;;
    *)
        error "Unknown command: $COMMAND"
        show_usage
        exit 1
        ;;
esac
