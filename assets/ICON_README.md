# DDS Application Icon

For proper cross-platform deployment, the application needs a high-resolution icon.
Currently using: assets/characters/default/animations/idle.gif

For production builds, this should be replaced with:
- icon.png (512x512 for Android)
- icon.ico (for Windows)
- icon.icns (for macOS)

The Fyne packaging tool will automatically generate platform-specific icons from the base PNG file.
