#!/bin/bash

echo "=== ANDROID BUILD VALIDATION ==="

# Get list of characters from assets/characters directory
characters=($(ls assets/characters/ | grep -v README | grep -v examples | grep -v templates | sort))
expected_count=${#characters[@]}

echo "Expected characters: $expected_count"
echo "Characters: ${characters[*]}"

# Count Android APKs for each architecture
android_arm64_count=0
android_arm_count=0
missing_apks=()

# Check each character for required Android builds
for character in "${characters[@]}"; do
    echo "Checking Android builds for character: $character"
    
    # Check for ARM64 APK
    arm64_artifact="artifacts/${character}-android-arm64-android"
    arm64_apk=$(find "$arm64_artifact" -name "*.apk" 2>/dev/null | head -1)
    if [[ -n "$arm64_apk" && -f "$arm64_apk" ]]; then
        size=$(stat -c%s "$arm64_apk")
        echo "✅ Found ARM64 APK: $arm64_apk ($size bytes)"
        ((android_arm64_count++))
    else
        echo "❌ Missing ARM64 APK for $character"
        missing_apks+=("$character-android-arm64")
    fi
    
    # Check for ARM32 APK
    arm_artifact="artifacts/${character}-android-arm-android32"
    arm_apk=$(find "$arm_artifact" -name "*.apk" 2>/dev/null | head -1)
    if [[ -n "$arm_apk" && -f "$arm_apk" ]]; then
        size=$(stat -c%s "$arm_apk")
        echo "✅ Found ARM32 APK: $arm_apk ($size bytes)"
        ((android_arm_count++))
    else
        echo "❌ Missing ARM32 APK for $character"
        missing_apks+=("$character-android-arm32")
    fi
done

echo ""
echo "=== ANDROID BUILD COUNT SUMMARY ==="
echo "Expected characters: $expected_count"
echo "Android ARM64 APKs found: $android_arm64_count"
echo "Android ARM32 APKs found: $android_arm_count"
echo "Total Android APKs expected: $((expected_count * 2))"
echo "Total Android APKs found: $((android_arm64_count + android_arm_count))"

if [[ ${#missing_apks[@]} -gt 0 ]]; then
    echo ""
    echo "❌ CRITICAL FAILURE: Missing Android APKs:"
    printf '  - %s\n' "${missing_apks[@]}"
    echo ""
    echo "This is a mandatory requirement for Android build validation"
    exit 1
fi

if [[ $android_arm64_count -ne $expected_count ]]; then
    echo "❌ CRITICAL FAILURE: Expected $expected_count ARM64 APKs, found $android_arm64_count"
    exit 1
fi

if [[ $android_arm_count -ne $expected_count ]]; then
    echo "❌ CRITICAL FAILURE: Expected $expected_count ARM32 APKs, found $android_arm_count"
    exit 1
fi

echo "✅ ANDROID BUILD VALIDATION PASSED"
echo "All required Android APKs are present and accounted for"
