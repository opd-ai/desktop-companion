# Touch Gesture System

This package provides cross-platform touch gesture translation for the DDS application, enabling seamless interaction on both desktop and mobile platforms.

## Overview

The gesture system translates touch gestures to traditional mouse events, allowing existing desktop interaction patterns to work naturally on mobile devices without code changes.

## Architecture

### Core Components

1. **GestureHandler** (`handler.go`): Core gesture detection and translation logic
2. **TouchAwareWidget** (`widget.go`): Fyne widget with integrated gesture support  
3. **PlatformAwareClickableWidget** (`../platform_clickable.go`): Drop-in replacement for ClickableWidget

### Gesture Translation

| Touch Gesture | Desktop Equivalent | Implementation |
|--------------|-------------------|----------------|
| **Single Tap** | Left Click | Direct callback invocation |
| **Long Press** | Right Click | Timer-based detection (600ms default) |
| **Double Tap** | Double Click | Time window detection (500ms default) |
| **Pan** | Drag | Movement threshold detection (10px default) |

## Usage

### Basic Integration

```go
// Replace existing ClickableWidget
widget := NewPlatformAwareClickableWidget(
    func() { /* left click/tap */ },
    func() { /* right click/long press */ },
)

// Set widget size
widget.SetSize(fyne.NewSize(128, 128))

// Add to container
content := container.NewWithoutLayout(widget)
```

### Advanced Features

```go
// With double tap support
widget := NewPlatformAwareClickableWidgetWithDoubleTap(
    func() { /* single tap */ },
    func() { /* long press */ },
    func() { /* double tap */ },
)

// Add drag support
widget.SetDragHandlers(
    func() { /* drag start */ },
    func(event *fyne.DragEvent) { /* drag move */ },
    func() { /* drag end */ },
)

// Check if gesture translation is active
if widget.IsGestureTranslationActive() {
    // Platform has touch capabilities
}
```

### Custom Gesture Configuration

```go
// Create custom gesture timing
config := &GestureConfig{
    DoubleTapWindow:   400 * time.Millisecond, // Faster double tap
    LongPressDuration: 800 * time.Millisecond, // Longer press required
    DragThreshold:     15.0,                   // Larger drag threshold
}

// Create handler with custom config
platform := platform.GetPlatformInfo()
handler := NewGestureHandler(platform, config)
```

## Platform Detection

The system automatically detects platform capabilities:

- **Desktop Platforms** (Windows, macOS, Linux): Use traditional mouse events directly
- **Mobile Platforms** (Android, iOS): Apply gesture translation  
- **Hybrid Platforms** (Linux with touch): Enable gesture translation

## Performance Characteristics

- **Memory Usage**: Minimal overhead (~100 bytes per widget)
- **CPU Usage**: Gesture detection uses efficient timer-based approach
- **Latency**: 
  - Single tap: <50ms
  - Long press: 600ms (configurable)
  - Double tap: 500ms detection window
  - Drag: 10px threshold detection

## Configuration

### Default Timing Values

```go
DoubleTapWindow:   500 * time.Millisecond  // Standard double-click timing
LongPressDuration: 600 * time.Millisecond  // Slightly longer than iOS (500ms)
DragThreshold:     10.0                    // 10 pixels minimum drag distance
```

### Platform-Specific Behavior

- **Desktop**: No gesture translation overhead
- **Mobile**: Full gesture translation with haptic feedback support (future)
- **Touch-enabled Desktop**: Gesture translation for touch input, mouse events unchanged

## Testing

The gesture system includes comprehensive tests covering:

- Gesture detection accuracy
- Platform-specific behavior
- Timing edge cases  
- Error conditions
- Performance characteristics

Run tests with:
```bash
go test ./lib/ui/gestures/... -v -cover
```

## Integration Notes

### Backward Compatibility

The system maintains 100% backward compatibility:
- Existing `ClickableWidget` usage unchanged on desktop
- New `PlatformAwareClickableWidget` is a drop-in replacement
- No changes required to existing interaction handlers

### Migration Path

1. Replace `NewClickableWidget` with `NewPlatformAwareClickableWidget`
2. No other code changes required
3. Touch gesture support automatically enabled on mobile platforms

### Performance Considerations

- Desktop platforms have zero performance impact
- Mobile platforms use efficient timer-based gesture detection
- Memory allocation minimized using object pooling patterns
- No background goroutines or continuous polling

## Future Enhancements

Planned features for subsequent phases:

1. **Haptic Feedback**: Native mobile haptic response integration
2. **Multi-touch Gestures**: Pinch-to-zoom, rotation support
3. **Gesture Customization**: User-configurable gesture timing
4. **Advanced Touch**: Force touch, 3D touch support
5. **Platform-Specific UX**: Native Android/iOS interaction patterns

## Dependencies

- `fyne.io/fyne/v2`: GUI framework and event handling
- `desktop-companion/lib/platform`: Platform detection system
- Go standard library: `time` package for gesture timing

All dependencies use permissive licenses compatible with commercial use.
