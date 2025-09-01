#!/bin/bash
set -e

# Test Android APK build simulation
echo "======================================="
echo "Android APK Build Test"
echo "======================================="

# Test character
CHARACTER="default"
ARCH="arm64"

echo "🔧 Simulating Android APK build for $CHARACTER ($ARCH)"

# Create test directory
TEST_DIR="/tmp/android-build-test"
rm -rf "$TEST_DIR"
mkdir -p "$TEST_DIR"

echo "📁 Created test directory: $TEST_DIR"

# Copy embedded character
if [[ -d "cmd/companion-$CHARACTER-test" ]]; then
    cp -r "cmd/companion-$CHARACTER-test"/* "$TEST_DIR/"
    cp go.mod go.sum "$TEST_DIR/"
    echo "✅ Copied embedded character files"
else
    echo "❌ Embedded character not found. Run embed-character first."
    exit 1
fi

# Create FyneApp.toml
cat > "$TEST_DIR/FyneApp.toml" << EOF
[Details]
Icon = "Icon.png"
Name = "$CHARACTER Companion"
ID = "ai.opd.$CHARACTER"
Version = "1.0.0"
Build = "1"

[Development]
AutoInject = true
EOF

echo "✅ Created FyneApp.toml"

# Create test icon
if [[ -f "assets/app/icon.png" ]]; then
    cp "assets/app/icon.png" "$TEST_DIR/Icon.png"
    echo "✅ Copied app icon"
else
    # Create simple test icon
    echo "📱 Creating test icon..."
    echo "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==" | base64 -d > "$TEST_DIR/Icon.png"
    echo "✅ Created test icon"
fi

# Simulate fyne package command (dry run)
cd "$TEST_DIR"
echo "🏗️ Simulating fyne package command..."
echo "Command: fyne package --target android/$ARCH --name '$CHARACTER Companion' --app-id 'ai.opd.$CHARACTER' --app-version '1.0.0' --release"

# Check if we can at least validate the project structure
if [[ -f "main.go" && -f "FyneApp.toml" && -f "Icon.png" ]]; then
    echo "✅ Android APK build structure is valid"
    echo "📦 Ready for fyne package command"
    
    # Show what would be built
    echo ""
    echo "Build Configuration:"
    echo "  Character: $CHARACTER"
    echo "  Target: android/$ARCH"
    echo "  App ID: ai.opd.$CHARACTER"
    echo "  Files: $(ls -1 | wc -l) files ready"
    echo "  Size: $(du -sh . | cut -f1) total"
    
    echo ""
    echo "✅ Android APK build test completed successfully!"
else
    echo "❌ Android APK build structure validation failed"
    exit 1
fi

# Cleanup
rm -rf "$TEST_DIR"
echo "🧹 Cleaned up test directory"
