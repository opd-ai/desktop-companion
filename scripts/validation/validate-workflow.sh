#!/bin/bash

# scripts/validation/validate-workflow.sh
# GitHub Actions workflow validation script
#
# Validates the character-specific binary generation workflow for completeness,
# correctness, and security best practices.
#
# Usage: ./scripts/validation/validate-workflow.sh [OPTIONS]
#
# Dependencies:
# - Python 3.x (for YAML validation)
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
# WORKFLOW VALIDATION CONFIGURATION
# ============================================================================

# Workflow files to validate
WORKFLOW_FILE="$PROJECT_ROOT/.github/workflows/build-character-binaries.yml"
WORKFLOW_DIR="$PROJECT_ROOT/.github/workflows"

# Validation settings
VALIDATE_YAML_SYNTAX=true
VALIDATE_SECURITY=true
VALIDATE_MATRIX=true
VALIDATE_ARTIFACTS=true

# ============================================================================
# HELP AND USAGE
# ============================================================================

show_usage() {
    cat << EOF
Usage: $0 [OPTIONS] [COMMAND]

Validate GitHub Actions workflows for correctness and best practices.

COMMANDS:
    validate           Validate all workflow aspects (default)
    syntax            Validate only YAML syntax
    security          Validate security configurations
    matrix            Validate platform matrix configuration
    artifacts         Validate artifact management setup
    report            Generate workflow validation report
    help              Show this help message

OPTIONS:
    --skip-syntax     Skip YAML syntax validation
    --skip-security   Skip security validation
    --skip-matrix     Skip matrix validation
    --skip-artifacts  Skip artifact validation
    -v, --verbose     Enable verbose output
    --dry-run        Show what would be validated

EXAMPLES:
    $0                # Validate all workflow aspects
    $0 syntax         # Validate only YAML syntax
    $0 security       # Validate only security settings
    $0 --skip-syntax  # Skip YAML validation

VALIDATION CHECKS:
    1. YAML syntax and structure
    2. Required jobs and steps
    3. Platform matrix configuration
    4. Security settings and permissions
    5. Artifact management setup
    6. Environment variable usage
    7. Android-specific configuration

OUTPUT:
    Validation report: $TEST_OUTPUT_DIR/workflow-validation-report.txt

EOF
}

# ============================================================================
# WORKFLOW VALIDATION FUNCTIONS
# ============================================================================

# Validate workflow file exists
validate_workflow_exists() {
    log "Checking workflow file existence..."
    
    if [[ ! -f "$WORKFLOW_FILE" ]]; then
        error "GitHub Actions workflow not found: $WORKFLOW_FILE"
        return 1
    fi
    
    success "Workflow file found: $WORKFLOW_FILE"
    return 0
}

# Validate YAML syntax
validate_yaml_syntax() {
    if [[ "$VALIDATE_YAML_SYNTAX" != "true" ]]; then
        log "Skipping YAML syntax validation"
        return 0
    fi
    
    log "Validating YAML syntax..."
    
    # Try to parse YAML using Python
    if command -v python3 >/dev/null 2>&1; then
        if python3 -c "
import sys, yaml
try:
    with open('$WORKFLOW_FILE', 'r') as f:
        yaml.safe_load(f)
    print('✓ YAML syntax valid')
except Exception as e:
    print(f'✗ YAML syntax error: {e}')
    sys.exit(1)
" 2>/dev/null; then
            success "YAML syntax is valid"
            return 0
        else
            error "YAML syntax validation failed"
            return 1
        fi
    else
        warning "Python3 not available, skipping YAML syntax validation"
        return 0
    fi
}

# Validate required jobs exist
validate_required_jobs() {
    log "Validating required workflow jobs..."
    
    # Check for essential jobs
    local required_jobs=("build" "test")
    local workflow_content
    workflow_content=$(cat "$WORKFLOW_FILE")
    
    for job in "${required_jobs[@]}"; do
        if echo "$workflow_content" | grep -q "^[[:space:]]*${job}:"; then
            success "Required job found: $job"
        else
            warning "Recommended job missing: $job"
        fi
    done
    
    return 0
}

# Validate platform matrix configuration
validate_platform_matrix() {
    if [[ "$VALIDATE_MATRIX" != "true" ]]; then
        log "Skipping platform matrix validation"
        return 0
    fi
    
    log "Validating platform matrix configuration..."
    
    local workflow_content
    workflow_content=$(cat "$WORKFLOW_FILE")
    
    # Check for matrix strategy
    if echo "$workflow_content" | grep -q "strategy:"; then
        success "Matrix strategy found"
        
        # Check for common platforms
        local platforms=("ubuntu-latest" "windows-latest" "macos-latest")
        for platform in "${platforms[@]}"; do
            if echo "$workflow_content" | grep -q "$platform"; then
                success "Platform included: $platform"
            else
                warning "Platform not included: $platform"
            fi
        done
    else
        warning "No matrix strategy found in workflow"
    fi
    
    return 0
}

# Validate artifact management
validate_artifact_management() {
    if [[ "$VALIDATE_ARTIFACTS" != "true" ]]; then
        log "Skipping artifact management validation"
        return 0
    fi
    
    log "Validating artifact management setup..."
    
    local workflow_content
    workflow_content=$(cat "$WORKFLOW_FILE")
    
    # Check for artifact upload/download actions
    if echo "$workflow_content" | grep -q "upload-artifact"; then
        success "Artifact upload action found"
    else
        warning "No artifact upload action found"
    fi
    
    if echo "$workflow_content" | grep -q "download-artifact"; then
        success "Artifact download action found"
    else
        warning "No artifact download action found"
    fi
    
    return 0
}

# Validate security configurations
validate_security() {
    if [[ "$VALIDATE_SECURITY" != "true" ]]; then
        log "Skipping security validation"
        return 0
    fi
    
    log "Validating security configurations..."
    
    local workflow_content
    workflow_content=$(cat "$WORKFLOW_FILE")
    
    # Check for permissions configuration
    if echo "$workflow_content" | grep -q "permissions:"; then
        success "Permissions configuration found"
        
        # Check for minimal permissions
        if echo "$workflow_content" | grep -A 10 "permissions:" | grep -q "contents: read"; then
            success "Minimal content permissions configured"
        else
            warning "Consider using minimal 'contents: read' permission"
        fi
    else
        warning "No explicit permissions configuration (will use default)"
    fi
    
    # Check for secrets usage
    if echo "$workflow_content" | grep -q "\${{ secrets\."; then
        warning "Secrets usage detected - ensure proper secret management"
    fi
    
    # Check for pull request triggers with write access
    if echo "$workflow_content" | grep -q "pull_request:" && echo "$workflow_content" | grep -q "contents: write"; then
        warning "Pull request triggers with write access may be a security risk"
    fi
    
    return 0
}

# Validate environment setup
validate_environment_setup() {
    log "Validating environment setup..."
    
    local workflow_content
    workflow_content=$(cat "$WORKFLOW_FILE")
    
    # Check for Go setup
    if echo "$workflow_content" | grep -q "setup-go"; then
        success "Go setup action found"
        
        # Check for Go version specification
        if echo "$workflow_content" | grep -A 5 "setup-go" | grep -q "go-version"; then
            success "Go version specified"
        else
            warning "Go version not explicitly specified"
        fi
    else
        warning "Go setup action not found"
    fi
    
    # Check for cache usage
    if echo "$workflow_content" | grep -q "cache"; then
        success "Caching configuration found"
    else
        warning "No caching configuration found (consider adding for performance)"
    fi
    
    return 0
}

# Validate Android-specific configuration
validate_android_configuration() {
    log "Validating Android-specific configuration..."
    
    local workflow_content
    workflow_content=$(cat "$WORKFLOW_FILE")
    
    # Check for Android-related steps
    if echo "$workflow_content" | grep -qi "android\|fyne"; then
        success "Android-related configuration found"
        
        # Check for Android SDK setup
        if echo "$workflow_content" | grep -qi "android.*sdk\|setup.*android"; then
            success "Android SDK setup found"
        else
            warning "Android SDK setup not found (may be needed for APK builds)"
        fi
    else
        log "No Android-specific configuration found (may not be needed)"
    fi
    
    return 0
}

# Generate workflow validation report
generate_workflow_report() {
    local report_file="$TEST_OUTPUT_DIR/workflow-validation-report.txt"
    
    {
        echo "# GitHub Actions Workflow Validation Report"
        echo "Generated: $(date -u '+%Y-%m-%d %H:%M:%S UTC')"
        echo ""
        echo "## Workflow File"
        echo "- Location: $WORKFLOW_FILE"
        echo "- Exists: $([ -f "$WORKFLOW_FILE" ] && echo "Yes" || echo "No")"
        if [[ -f "$WORKFLOW_FILE" ]]; then
            echo "- Size: $(wc -l < "$WORKFLOW_FILE") lines"
            echo "- Last modified: $(stat -c %y "$WORKFLOW_FILE" 2>/dev/null || stat -f %Sm "$WORKFLOW_FILE" 2>/dev/null || echo "Unknown")"
        fi
        echo ""
        echo "## Validation Checks Performed"
        echo "- YAML syntax: $VALIDATE_YAML_SYNTAX"
        echo "- Security configuration: $VALIDATE_SECURITY"
        echo "- Platform matrix: $VALIDATE_MATRIX"
        echo "- Artifact management: $VALIDATE_ARTIFACTS"
        echo ""
        echo "## Recommendations"
        echo "1. Ensure Go version is explicitly specified"
        echo "2. Use minimal permissions where possible"
        echo "3. Enable caching for Go modules to improve performance"
        echo "4. Consider adding security scanning steps"
        echo "5. Validate that all target platforms are included in matrix"
        echo ""
        echo "## Additional Files"
        if [[ -d "$WORKFLOW_DIR" ]]; then
            echo "Other workflow files:"
            find "$WORKFLOW_DIR" -name "*.yml" -o -name "*.yaml" | while read -r file; do
                echo "- $(basename "$file")"
            done
        fi
        echo ""
    } > "$report_file"
    
    success "Workflow validation report saved to: $report_file"
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
        --skip-syntax)
            VALIDATE_YAML_SYNTAX=false
            shift
            ;;
        --skip-security)
            VALIDATE_SECURITY=false
            shift
            ;;
        --skip-matrix)
            VALIDATE_MATRIX=false
            shift
            ;;
        --skip-artifacts)
            VALIDATE_ARTIFACTS=false
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
        validate|syntax|security|matrix|artifacts|report)
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

# Create test output directory
mkdir -p "$TEST_OUTPUT_DIR"

# Validate workflow file exists first
if ! validate_workflow_exists; then
    error "Cannot proceed without workflow file"
    exit 1
fi

# Execute command
case $COMMAND in
    validate)
        log "Starting complete workflow validation..."
        validate_yaml_syntax
        validate_required_jobs
        validate_platform_matrix
        validate_artifact_management
        validate_security
        validate_environment_setup
        validate_android_configuration
        generate_workflow_report
        ;;
    syntax)
        log "Validating YAML syntax only..."
        validate_yaml_syntax
        ;;
    security)
        log "Validating security configuration..."
        validate_security
        ;;
    matrix)
        log "Validating platform matrix..."
        validate_platform_matrix
        ;;
    artifacts)
        log "Validating artifact management..."
        validate_artifact_management
        ;;
    report)
        log "Generating workflow validation report..."
        generate_workflow_report
        ;;
    *)
        error "Unknown command: $COMMAND"
        show_usage
        exit 1
        ;;
esac

success "Workflow validation completed successfully!"
exit 0
