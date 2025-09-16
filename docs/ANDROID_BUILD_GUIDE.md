# Android Build Guide for DDS

This guide explains how to build the Desktop Dating Simulator (DDS) for Android devices using Fyne's cross-platform capabilities.

## Prerequisites

### Required Software

1. **Go 1.24.5 or higher**
   ```bash
   go version  # Should show 1.24.5 or higher
   ```

2. **Fyne CLI Tool**
   ```bash
   go install fyne.io/tools/cmd/fyne@latest
   fyne version
   ```

3. **Java Development Kit (JDK) 8+**
   ```bash
   java -version
   ```

4. **Android SDK and NDK** (for full functionality)
   - Download from [Android Studio](https://developer.android.com/studio)
   - Set `ANDROID_HOME` environment variable
   - Install Android NDK

### System Dependencies (Linux)

```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install -y gcc pkg-config libgl1-mesa-dev xorg-dev

# Fedora/RHEL
sudo dnf install gcc pkg-config mesa-libGL-devel libXrandr-devel libXcursor-devel libXinerama-devel libXi-devel

# Arch Linux
sudo pacman -S gcc pkg-config libgl libxrandr libxcursor libxinerama libxi
```

## Building Android APK

### Quick Start (Debug Build)

```bash
# Clone and prepare
git clone https://github.com/opd-ai/desktop-companion
cd desktop-companion

# Install dependencies
make deps

# Build debug APK
make android-debug
```

The debug APK will be created in the `build/` directory.

### Production Build

```bash
# Build release APK (requires signing setup)
make android-apk
```

### Manual Build Commands

```bash
# Create build directory
mkdir -p build
cd build

# Build Android APK with fyne
fyne package --target android \
    --app-id ai.opd.dds \
    --name "Desktop Companion" \
    --app-version "1.0.0" \
    --app-build 1 \
    --icon "../assets/characters/default/animations/idle.gif" \
    --src "../cmd/companion" \
    --release
```

## Installation on Android Device

### Via ADB (Android Debug Bridge)

1. **Enable Developer Options** on your Android device:
   - Go to Settings → About Phone
   - Tap "Build Number" 7 times
   - Go back to Settings → Developer Options
   - Enable "USB Debugging"

2. **Install APK**:
   ```bash
   # Connect device via USB
   adb devices  # Verify device is connected
   
   # Install debug version
   make android-install-debug
   
   # Or install manually
   adb install -r build/companion-debug.apk
   ```

### Via File Transfer

1. Copy the APK file to your Android device
2. Use a file manager to locate and tap the APK
3. Allow installation from unknown sources when prompted

## Platform-Specific Features

### Android Optimizations

The Android version includes several mobile-specific optimizations:

1. **Touch-Friendly Interface**
   - Larger touch targets (25% of screen width)
   - Long-press replaces right-click
   - Double-tap replaces double-click
   - Swipe gestures for navigation

2. **Mobile Layout**
   - Fullscreen experience
   - Responsive character sizing
   - Bottom control bar for actions
   - Picture-in-picture mode support

3. **Performance Optimizations**
   - Reduced animation frame rates for battery life
   - Memory-efficient rendering
   - Background processing limitations

### Platform Detection

The app automatically detects it's running on Android and adapts:

```go
// Platform detection in action
platform := platform.GetPlatformInfo()
if platform.IsMobile() {
    // Use mobile-optimized behavior
    character.EnableMobileBehavior()
}
```

## Configuration

### App Configuration (FyneApp.toml)

```toml
[Details]
Icon = "assets/characters/default/animations/idle.gif"
Name = "Desktop Companion"
ID = "ai.opd.dds"
Version = "1.0.0"
Build = 1

[Metadata]
Description = "A cross-platform virtual companion with AI-powered interactions"
DeveloperName = "OPD AI"
DeveloperID = "ai.opd"
Website = "https://github.com/opd-ai/desktop-companion"
```

### Character Configuration

Characters work identically on Android with automatic mobile adaptations:

```json
{
  "name": "Mobile Pet",
  "platformConfig": {
    "mobile": {
      "behavior": {
        "defaultSize": 256,
        "windowMode": "fullscreen"
      },
      "mobileControls": {
        "showBottomBar": true,
        "hapticFeedback": true
      }
    }
  }
}
```

## Troubleshooting

### Common Issues

1. **"Android SDK not found"**
   ```bash
   # Set Android SDK path
   export ANDROID_HOME=/path/to/android/sdk
   export PATH=$PATH:$ANDROID_HOME/tools:$ANDROID_HOME/platform-tools
   ```

2. **"NDK not found"**
   ```bash
   # Install NDK via Android Studio or download manually
   # Set NDK path in Android Studio or via environment variable
   ```

3. **"Build failed with CGO errors"**
   ```bash
   # Install required development tools
   sudo apt-get install build-essential
   ```

4. **"APK won't install"**
   - Enable "Install unknown apps" for your file manager
   - Check if Developer Options are enabled
   - Try `adb install -r` to replace existing version

### Performance Issues

1. **Slow animations**: Normal on older devices, animations automatically adapt
2. **High battery usage**: Enable battery optimization in Android settings
3. **Memory issues**: Close other apps, restart device if necessary

### Development Tips

1. **Use debug builds** during development (faster, no signing required)
2. **Test on real devices** for accurate performance testing
3. **Monitor logcat** for debugging:
   ```bash
   adb logcat | grep DDS
   ```

## CI/CD Integration

The project includes GitHub Actions workflows for automated Android builds:

```yaml
# .github/workflows/build.yml includes Android build job
build-android:
  runs-on: ubuntu-latest
  steps:
    - name: Set up Android SDK
      uses: android-actions/setup-android@v2
    - name: Build APK
      run: make android-debug
```

## Distribution

### Google Play Store

For Play Store distribution:

1. Create a signed release build
2. Follow [Google Play Console guidelines](https://play.google.com/console)
3. Upload AAB (Android App Bundle) format when possible

### F-Droid

For open-source distribution:

1. Ensure all dependencies use compatible licenses
2. Submit to [F-Droid](https://f-droid.org) repository
3. Follow F-Droid build requirements

### Direct APK Distribution

For direct distribution:

1. Sign APK with release keystore
2. Host on GitHub Releases or website
3. Provide installation instructions for users

## Security Considerations

1. **App Signing**: Always sign release builds
2. **Permissions**: Review Android permissions in generated manifest
3. **Network Security**: All network traffic is encrypted
4. **Data Storage**: Character data stored locally only

## Known Limitations

1. **Cross-compilation**: Must build on Linux/macOS with Android SDK
2. **File System**: Limited file system access on Android
3. **Background Processing**: Limited by Android's background restrictions
4. **Hardware Access**: Some desktop features not available on mobile

## Getting Help

- **Build Issues**: Check GitHub Issues or create new issue
- **Android-Specific Problems**: Review Fyne mobile documentation
- **Performance Questions**: Enable debug mode and review logs
- **Feature Requests**: Submit GitHub issue with "android" label

## Next Steps

After successful Android build:

1. Test on various Android devices and versions
2. Optimize performance for target hardware
3. Set up automated testing and deployment
4. Plan for app store submission process
