# Platform Configuration Guide

## Overview

The Platform Configuration system enables Desktop Companion characters to adapt their behavior, appearance, and interactions based on the target platform (desktop vs mobile). This system maintains full backward compatibility while providing sophisticated cross-platform customization capabilities.

## Key Features

- **Cross-Platform Compatibility**: Single character card works on both desktop and mobile
- **Platform-Specific Behavior**: Different interactions, sizing, and controls per platform
- **Touch Optimization**: Mobile-specific UI elements and haptic feedback
- **Backward Compatibility**: Existing character cards work unchanged
- **Validation**: Comprehensive validation prevents configuration conflicts

## JSON Schema Structure

### Platform Configuration Root

```json
{
  "platformConfig": {
    "desktop": { /* PlatformSpecificConfig */ },
    "mobile": { /* PlatformSpecificConfig */ }
  }
}
```

### Platform-Specific Configuration

```json
{
  "behavior": {
    "movementEnabled": true|false,
    "defaultSize": 32-1024,
    "idleTimeout": 0+
  },
  "windowMode": "overlay"|"fullscreen"|"pip",
  "touchOptimized": true|false,
  "interactions": {
    "interactionName": {
      // Standard InteractionConfig fields
      "triggers": ["tap", "longpress", "doubletap", "swipe"],
      "effects": {"statName": value},
      "animations": ["animationName"],
      "responses": ["response text"],
      "cooldown": 0+,
      
      // Platform-specific extensions
      "hapticPattern": "light"|"medium"|"heavy",
      "touchFeedback": true|false,
      "gestureEnabled": true|false
    }
  },
  "mobileControls": {
    "showBottomBar": true|false,
    "swipeGesturesEnabled": true|false,
    "hapticFeedback": true|false,
    "largeButtons": true|false,
    "contextMenuStyle": "bottomsheet"|"popup"|"fullscreen"
  }
}
```

## Platform Differences

### Desktop Platform

**Optimizations:**
- Mouse-driven interactions (click, rightclick, doubleclick, hover)
- Keyboard shortcuts available
- Precise cursor positioning
- Window overlay mode
- Smaller character sizes (64-256px typical)

**Default Behavior:**
- Movement enabled (draggable character)
- Standard context menus
- Keyboard accessibility
- Multi-window support

**Example Desktop Config:**
```json
{
  "desktop": {
    "behavior": {
      "movementEnabled": true,
      "defaultSize": 128,
      "idleTimeout": 30
    },
    "windowMode": "overlay",
    "interactions": {
      "pet": {
        "triggers": ["click"],
        "effects": {"happiness": 10},
        "cooldown": 5
      }
    }
  }
}
```

### Mobile Platform

**Optimizations:**
- Touch-driven interactions (tap, longpress, doubletap, swipe)
- Haptic feedback integration
- Larger touch targets (256-512px typical)
- Fullscreen or picture-in-picture mode
- Battery usage considerations

**Default Behavior:**
- Movement disabled (fixed positioning)
- Touch-optimized controls
- Bottom sheet menus
- Gesture navigation

**Example Mobile Config:**
```json
{
  "mobile": {
    "behavior": {
      "movementEnabled": false,
      "defaultSize": 256,
      "idleTimeout": 60
    },
    "windowMode": "fullscreen",
    "touchOptimized": true,
    "mobileControls": {
      "showBottomBar": true,
      "swipeGesturesEnabled": true,
      "hapticFeedback": true,
      "largeButtons": true,
      "contextMenuStyle": "bottomsheet"
    },
    "interactions": {
      "pet": {
        "triggers": ["tap"],
        "effects": {"happiness": 15},
        "cooldown": 3,
        "hapticPattern": "light",
        "touchFeedback": true
      }
    }
  }
}
```

## Trigger Adaptation

The platform loader automatically adapts interaction triggers between platforms:

| Touch Trigger | Desktop Equivalent | Description |
|---------------|-------------------|-------------|
| `tap` | `click` | Primary interaction |
| `longpress` | `rightclick` | Secondary/context interaction |
| `doubletap` | `doubleclick` | Play/special interaction |
| `swipe` | `click` | Fallback to primary interaction |

**Standard triggers** (`click`, `rightclick`, `doubleclick`, `hover`) pass through unchanged on all platforms.

## Implementation Usage

### Loading Platform-Aware Characters

```go
import "desktop-companion/lib/character"

// Create platform-aware loader
loader := character.NewPlatformAwareLoader()

// Load character with platform adaptations
card, err := loader.LoadCharacterCard("path/to/character.json")
if err != nil {
    log.Fatal(err)
}

// Character automatically adapts to current platform
```

### Getting Platform-Specific Configuration

```go
// Get current platform configuration
platformConfig := loader.GetPlatformConfig(card)
if platformConfig != nil {
    // Use platform-specific settings
    windowMode := platformConfig.WindowMode
    touchOptimized := platformConfig.TouchOptimized
}
```

### Validation

```go
// Validate platform configuration
err := character.ValidatePlatformConfig(card)
if err != nil {
    log.Printf("Platform config validation failed: %v", err)
}
```

## Best Practices

### 1. Progressive Enhancement

Start with a solid base configuration, then add platform-specific enhancements:

```json
{
  "name": "My Character",
  "behavior": {
    "defaultSize": 128,
    "movementEnabled": true
  },
  "platformConfig": {
    "mobile": {
      "behavior": {
        "defaultSize": 256,
        "movementEnabled": false
      }
    }
  }
}
```

### 2. Touch-Friendly Mobile Design

Mobile configurations should prioritize touch accessibility:

```json
{
  "mobile": {
    "behavior": {"defaultSize": 256},
    "touchOptimized": true,
    "mobileControls": {
      "largeButtons": true,
      "hapticFeedback": true
    },
    "interactions": {
      "pet": {
        "triggers": ["tap"],
        "hapticPattern": "light",
        "touchFeedback": true
      }
    }
  }
}
```

### 3. Battery Optimization

Mobile platforms should use longer cooldowns and timeouts:

```json
{
  "mobile": {
    "behavior": {
      "idleTimeout": 60
    },
    "interactions": {
      "pet": {
        "cooldown": 8
      }
    }
  }
}
```

### 4. Interaction Effect Balancing

Mobile interactions can have stronger effects due to less frequent usage:

```json
{
  "desktop": {
    "interactions": {
      "pet": {"effects": {"happiness": 10}}
    }
  },
  "mobile": {
    "interactions": {
      "pet": {"effects": {"happiness": 15}}
    }
  }
}
```

## Validation Rules

### Behavior Validation
- `idleTimeout`: Must be ≥ 0
- `defaultSize`: Must be 32-1024 pixels
- `movementEnabled`: Boolean

### Window Mode Validation
- Valid values: `"overlay"`, `"fullscreen"`, `"pip"`
- `"overlay"` may not work well on mobile (warning only)

### Mobile Controls Validation
- `mobileControls` only valid in mobile platform configuration
- All boolean fields default to `false` if not specified

### Interaction Validation
- `hapticPattern`: Must be `"light"`, `"medium"`, or `"heavy"`
- `contextMenuStyle`: Must be `"bottomsheet"`, `"popup"`, or `"fullscreen"`
- All trigger adaptations are validated against known trigger types

## Example Character Cards

### Basic Cross-Platform Character

See `assets/characters/examples/cross_platform_character.json` for a complete example demonstrating:
- Desktop overlay mode with mouse interactions
- Mobile fullscreen mode with touch optimization
- Platform-specific interaction effects
- Mobile controls configuration
- Haptic feedback patterns

### Migration from Existing Character

To add platform support to an existing character:

1. **Keep existing configuration** as the base
2. **Add platformConfig section** with specific overrides
3. **Test on both platforms** to ensure expected behavior
4. **Validate configuration** using the validation functions

```json
{
  // Existing character configuration...
  "name": "Existing Character",
  "behavior": {"defaultSize": 128},
  
  // Add platform configuration
  "platformConfig": {
    "mobile": {
      "behavior": {"defaultSize": 256},
      "touchOptimized": true
    }
  }
}
```

## Error Handling

The platform configuration system provides detailed error messages for common issues:

- **Invalid window mode**: Clear error with valid options
- **Size out of range**: Specific bounds checking with recommended values
- **Platform mismatch**: Mobile-only features used in desktop config
- **Invalid triggers**: Unknown trigger types with suggestions

All validation errors include the specific configuration path and recommended fixes.

## Backward Compatibility

- **Existing character cards work unchanged** - no platform config required
- **Default behavior preserved** - desktop behavior remains identical
- **Graceful degradation** - missing platform configs use base configuration
- **Version-agnostic** - platform features are purely additive

The platform configuration system is designed to enhance existing characters without breaking changes, ensuring a smooth transition for existing users while enabling new cross-platform capabilities.
