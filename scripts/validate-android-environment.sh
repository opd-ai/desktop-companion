#!/bin/bash
# Android Build Environment Validation Script
# Validates compatibility between Go, Fyne, and Android NDK versions

set -e

echo "=== Android Build Environment Validation ==="

# Check Go version compatibility
GO_VERSION=$(go version | grep -o 'go[0-9]\+\.[0-9]\+' | sed 's/go//')
echo "Go Version: $GO_VERSION"

# Check Fyne version
FYNE_VERSION=$(go list -m fyne.io/fyne/v2 | awk '{print $2}')
echo "Fyne Version: $FYNE_VERSION"

# Check NDK version
if [[ -n "$ANDROID_NDK_ROOT" && -d "$ANDROID_NDK_ROOT" ]]; then
    NDK_VERSION=$(basename "$ANDROID_NDK_ROOT")
    echo "NDK Version: $NDK_VERSION"
else
    echo "❌ ANDROID_NDK_ROOT not set or directory not found"
    exit 1
fi

# Validate compatibility matrix
validate_compatibility() {
    local go_ver="$1"
    local fyne_ver="$2" 
    local ndk_ver="$3"
    
    # Known compatible combinations
    if [[ "$fyne_ver" =~ ^v2\.5\. ]] && [[ "$ndk_ver" =~ ^(25|26|27)\. ]]; then
        echo "✅ Fyne v2.5.x is compatible with NDK 25.x-27.x"
        return 0
    elif [[ "$fyne_ver" =~ ^v2\.4\. ]] && [[ "$ndk_ver" =~ ^25\. ]]; then
        echo "✅ Fyne v2.4.x is compatible with NDK 25.x"
        return 0
    else
        echo "⚠️ Untested compatibility: Fyne $fyne_ver with NDK $ndk_ver"
        return 1
    fi
}

if validate_compatibility "$GO_VERSION" "$FYNE_VERSION" "$NDK_VERSION"; then
    echo "✅ Environment validation passed"
else
    echo "❌ Environment validation failed - consider updating versions"
    exit 1
fi
