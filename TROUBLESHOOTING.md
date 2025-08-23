# Troubleshooting Guide

## Common Issues and Solutions

### Build Issues

**Error: `go: cannot find main module`**
```bash
# Solution: Ensure you're in the project root directory
cd /path/to/desktop-companion
go mod tidy
```

**Error: `package fyne.io/fyne/v2 is not in GOROOT`**
```bash
# Solution: Download dependencies
go mod download
go mod tidy
```

**Error: CGO compilation errors**
```bash
# Linux: Install required libraries
sudo apt-get install libx11-dev libxcursor-dev libxrandr-dev libxinerama-dev libxi-dev libgl1-mesa-dev

# macOS: Install Xcode command line tools  
xcode-select --install

# Windows: Install TDM-GCC or Visual Studio Build Tools
```

### Runtime Issues

**Error: `failed to load character card`**
- Check that `character.json` exists in the specified path
- Verify JSON syntax is valid (use `jsonlint` or online validator)
- Ensure all required fields are present

**Error: `failed to load animation 'idle'`**
- Add the required GIF files to the animations directory
- See `assets/characters/default/animations/SETUP.md` for details
- Verify GIF files are valid using an image viewer

**Error: `failed to initialize display`**
```bash
# Linux: Ensure X11 is running
echo $DISPLAY  # Should show something like :0

# Check if GUI libraries are available
ldd ./companion  # Look for missing libraries

# For headless servers, use Xvfb
sudo apt-get install xvfb
xvfb-run -a ./companion
```

**Error: Window not appearing or not staying on top**
- Some window managers don't support always-on-top
- Try different desktop environments (GNOME, KDE, XFCE)
- Check window manager settings for overlay permissions

### Performance Issues

**High memory usage**
- Check GIF file sizes (should be <1MB each)
- Monitor using: `go run cmd/companion/main.go -memprofile=mem.prof`
- Analyze with: `go tool pprof mem.prof`

**Poor animation performance**
- Reduce GIF frame rates to 10-15 FPS
- Optimize GIF files using tools like `gifsicle`
- Lower character size in character.json

**Slow startup**
- Ensure GIF files are optimized
- Check disk I/O performance
- Verify no antivirus interference

### Development Issues

**Tests failing**
```bash
# Run tests with verbose output
go test ./... -v

# Check specific package
go test ./internal/character -v

# Run with race detection
go test ./... -race
```

**Import errors in IDE**
```bash
# Regenerate Go module cache
go clean -modcache
go mod download

# For VS Code, reload window
# For GoLand, invalidate caches
```

### Platform-Specific Issues

#### Linux
**Wayland compatibility issues**
```bash
# Force X11 mode
export GDK_BACKEND=x11
./companion
```

**Permission denied errors**
```bash
# Make binary executable
chmod +x ./companion

# Check SELinux/AppArmor policies if applicable
```

#### macOS
**"Developer cannot be verified" error**
```bash
# Allow unsigned applications
sudo spctl --master-disable

# Or right-click app → Open → Allow
```

**Gatekeeper blocking execution**
```bash
# Remove quarantine attribute
xattr -d com.apple.quarantine ./companion-macos
```

#### Windows
**Windows Defender blocking execution**
- Add exception for the application folder
- Temporarily disable real-time protection for testing

**Missing DLL errors**
- Install Visual C++ Redistributable
- Ensure Windows version is supported (Windows 10+)

## Debug Mode

Enable debug logging for detailed troubleshooting:

```bash
go run cmd/companion/main.go -debug
```

Debug output includes:
- Character card loading details
- Animation file processing
- Window creation events
- Interaction handling
- Performance metrics

## Performance Monitoring

### Memory Profiling
```bash
# Generate memory profile
go run cmd/companion/main.go -memprofile=mem.prof

# Analyze profile
go tool pprof mem.prof
(pprof) top10
(pprof) list main.main
```

### CPU Profiling
```bash
# Generate CPU profile
go run cmd/companion/main.go -cpuprofile=cpu.prof

# Analyze profile
go tool pprof cpu.prof
(pprof) top10
(pprof) web  # Opens browser visualization
```

## Getting Help

### Check System Requirements
- Go 1.21 or higher
- C compiler (for CGO dependencies)
- OpenGL 2.1+ support
- X11 (Linux), Cocoa (macOS), or Win32 (Windows)

### Verify Installation
```bash
# Check Go version
go version

# Check if CGO is available
go env CGO_ENABLED

# Test basic compilation
go build -o test cmd/companion/main.go
```

### Log Analysis
Look for specific error patterns:

- `permission denied`: Check file permissions
- `no such file`: Verify paths and file existence  
- `invalid character`: JSON syntax errors
- `failed to decode`: Invalid GIF files
- `display not found`: X11/display issues

### Community Resources
- GitHub Issues: Report bugs and feature requests
- Go Forums: General Go development questions
- Fyne Community: GUI framework specific issues

### Creating Minimal Reproduction
For bug reports, create a minimal example:
1. Use the default character configuration
2. Add minimal GIF files (even 1x1 pixel animations)
3. Run with `-debug` flag
4. Include full error output and system information
