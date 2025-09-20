#!/bin/bash
set -e

echo "🚀 Setting up Desktop Companion development environment..."

# Update package lists
echo "📦 Updating package lists..."
sudo apt-get update

# Install CGO dependencies for Fyne GUI framework
echo "🎨 Installing Fyne GUI dependencies..."
sudo apt-get install -y \
    pkg-config \
    libgl1-mesa-dev \
    xorg-dev \
    libx11-dev \
    libxcursor-dev \
    libxrandr-dev \
    libxinerama-dev \
    libxi-dev \
    libasound2-dev

# Install additional utilities
echo "🔧 Installing additional development tools..."
sudo apt-get install -y \
    wget \
    unzip \
    build-essential

# Install Go dependencies
echo "📖 Downloading Go dependencies..."
go mod download

# Install Fyne CLI tool
echo "📱 Installing Fyne CLI tool..."
go install fyne.io/tools/cmd/fyne@latest

# Install Go development tools
echo "🛠️ Installing Go development tools..."
go install golang.org/x/tools/cmd/goimports@latest
go install honnef.co/go/tools/cmd/staticcheck@latest

# Create Android SDK directory
echo "📱 Setting up Android SDK..."
sudo mkdir -p /opt/android-sdk
sudo chown -R vscode:vscode /opt/android-sdk

# Download and install Android command line tools
CMDTOOLS_VERSION="11076708"
CMDTOOLS_URL="https://dl.google.com/android/repository/commandlinetools-linux-${CMDTOOLS_VERSION}_latest.zip"

echo "⬇️ Downloading Android command line tools..."
cd /tmp
wget -q "${CMDTOOLS_URL}" -O cmdtools.zip
unzip -q cmdtools.zip
mkdir -p /opt/android-sdk/cmdline-tools
mv cmdline-tools /opt/android-sdk/cmdline-tools/latest

# Set up Android SDK environment
export ANDROID_HOME="/opt/android-sdk"
export PATH="${PATH}:${ANDROID_HOME}/cmdline-tools/latest/bin:${ANDROID_HOME}/platform-tools"

# Accept Android SDK licenses
echo "📄 Accepting Android SDK licenses..."
yes | sdkmanager --licenses > /dev/null 2>&1 || true

# Install essential Android SDK components
echo "📦 Installing Android SDK components..."
sdkmanager --install \
    "platform-tools" \
    "platforms;android-30" \
    "platforms;android-33" \
    "build-tools;30.0.3" \
    "build-tools;33.0.2"

# Install Android NDK
echo "🔨 Installing Android NDK..."
NDK_VERSION="25.2.9519653"
sdkmanager --install "ndk;${NDK_VERSION}"

# Create NDK symlink for easier access
sudo ln -sf "/opt/android-sdk/ndk/${NDK_VERSION}" /opt/android-ndk
sudo chown -R vscode:vscode /opt/android-sdk

# Verify installation
echo "✅ Verifying installations..."

# Check Go version
echo "Go version: $(go version)"

# Check Fyne CLI
if command -v fyne >/dev/null 2>&1; then
    echo "Fyne CLI: $(fyne version)"
else
    echo "⚠️ Fyne CLI not found in PATH"
fi

# Check Android SDK
if [ -d "/opt/android-sdk/platform-tools" ]; then
    echo "Android SDK: Installed at /opt/android-sdk"
    echo "Platform tools: $(ls /opt/android-sdk/platform-tools/adb 2>/dev/null && echo "Available" || echo "Missing")"
else
    echo "⚠️ Android SDK installation incomplete"
fi

# Check Android NDK
if [ -d "/opt/android-ndk" ]; then
    echo "Android NDK: Installed at /opt/android-ndk"
else
    echo "⚠️ Android NDK installation incomplete"
fi

# Test build capability
echo "🧪 Testing build capability..."
cd /workspaces/desktop-companion
if make deps > /dev/null 2>&1; then
    echo "✅ Makefile deps target works"
else
    echo "⚠️ Makefile deps target failed"
fi

echo ""
echo "🎉 Development environment setup complete!"
echo ""
echo "Available commands:"
echo "  make build              - Build for current platform"
echo "  make test               - Run unit tests"
echo "  make android-apk        - Build Android APK"
echo "  make run                - Run application locally"
echo ""
echo "Android environment:"
echo "  ANDROID_HOME: ${ANDROID_HOME:-/opt/android-sdk}"
echo "  ANDROID_NDK_HOME: ${ANDROID_NDK_HOME:-/opt/android-ndk}"
echo ""