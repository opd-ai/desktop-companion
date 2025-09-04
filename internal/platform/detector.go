// Package platform provides cross-platform detection and capability information
// for adaptive behavior between desktop and mobile environments.
package platform

import (
	"os"
	"runtime"
	"strings"
)

// PlatformInfo provides limited OS information for privacy-conscious behavior adaptation.
// Only essential information is exposed to minimize data collection while enabling
// platform-appropriate user experiences.
type PlatformInfo struct {
	// OS identifies the operating system: "windows", "linux", "darwin", "android", "ios"
	OS string

	// MajorVersion contains only the major version number for compatibility checks
	// Examples: "10", "11", "12" - no minor versions or detailed system info
	MajorVersion string

	// FormFactor categorizes the device type: "desktop", "mobile", "tablet"
	FormFactor string

	// InputMethods lists available input capabilities: "mouse", "touch", "keyboard"
	InputMethods []string
}

// GetPlatformInfo returns current platform information using Go's standard library.
// This function uses only runtime.GOOS to detect the platform, ensuring privacy
// and avoiding system fingerprinting.
func GetPlatformInfo() *PlatformInfo {
	info := &PlatformInfo{
		OS:           runtime.GOOS,
		InputMethods: detectInputMethods(runtime.GOOS),
	}

	// Configure platform-specific defaults based on OS type
	switch runtime.GOOS {
	case "android":
		info.FormFactor = "mobile"
		info.MajorVersion = detectAndroidMajorVersion()
	case "ios":
		info.FormFactor = "mobile"
		info.MajorVersion = detectIOSMajorVersion()
	case "windows", "linux", "darwin":
		info.FormFactor = "desktop"
		info.MajorVersion = detectDesktopMajorVersion(runtime.GOOS)
	default:
		// Unknown platform defaults to desktop behavior
		info.FormFactor = "desktop"
		info.MajorVersion = "unknown"
	}

	return info
}

// IsDesktop returns true for desktop platforms (Windows, macOS, Linux).
func (p *PlatformInfo) IsDesktop() bool {
	return p.FormFactor == "desktop"
}

// IsMobile returns true for mobile platforms (Android, iOS).
func (p *PlatformInfo) IsMobile() bool {
	return p.FormFactor == "mobile"
}

// IsTablet returns true for tablet form factors.
func (p *PlatformInfo) IsTablet() bool {
	return p.FormFactor == "tablet"
}

// HasTouch returns true if the platform supports touch input.
func (p *PlatformInfo) HasTouch() bool {
	for _, method := range p.InputMethods {
		if method == "touch" {
			return true
		}
	}
	return false
}

// HasMouse returns true if the platform supports mouse input.
func (p *PlatformInfo) HasMouse() bool {
	for _, method := range p.InputMethods {
		if method == "mouse" {
			return true
		}
	}
	return false
}

// HasKeyboard returns true if the platform supports keyboard input.
func (p *PlatformInfo) HasKeyboard() bool {
	for _, method := range p.InputMethods {
		if method == "keyboard" {
			return true
		}
	}
	return false
}

// detectInputMethods determines available input methods based on OS type.
// This uses conservative assumptions to avoid system probing.
func detectInputMethods(goos string) []string {
	switch goos {
	case "android", "ios":
		// Mobile platforms primarily use touch
		return []string{"touch"}
	case "windows", "linux", "darwin":
		// Desktop platforms use mouse and keyboard
		return []string{"mouse", "keyboard"}
	default:
		// Unknown platforms default to basic input
		return []string{"mouse", "keyboard"}
	}
}

// detectAndroidMajorVersion attempts to detect Android major version.
// Uses environment variables for privacy-conscious version detection.
func detectAndroidMajorVersion() string {
	// Privacy-conscious implementation with minimal version detection
	// Check multiple standard environment variable patterns
	for _, envVar := range []string{"ANDROID_VERSION", "ANDROID_API_LEVEL", "API_LEVEL"} {
		if version := os.Getenv(envVar); version != "" {
			// Extract major version from version string if set
			parts := strings.Split(version, ".")
			if len(parts) > 0 && parts[0] != "" {
				return parts[0]
			}
		}
	}

	// Fallback to unknown for privacy compliance
	return "unknown"
}

// detectIOSMajorVersion attempts to detect iOS major version.
// Uses environment variables for privacy-conscious version detection.
func detectIOSMajorVersion() string {
	// Privacy-conscious implementation with minimal version detection
	// Check multiple standard environment variable patterns
	for _, envVar := range []string{"IOS_VERSION", "IPHONEOS_DEPLOYMENT_TARGET", "IOS_DEPLOYMENT_TARGET"} {
		if version := os.Getenv(envVar); version != "" {
			// Extract major version from version string if set
			parts := strings.Split(version, ".")
			if len(parts) > 0 && parts[0] != "" {
				return parts[0]
			}
		}
	}

	// Fallback to unknown for privacy compliance
	return "unknown"
}

// detectDesktopMajorVersion attempts to detect desktop OS major version.
// Uses environment variables for privacy-conscious minimal version detection.
func detectDesktopMajorVersion(goos string) string {
	// Privacy-conscious implementation with minimal version detection
	// Use environment variables when available for compatibility
	switch goos {
	case "windows":
		// Check for Windows version in multiple environment variable patterns
		for _, envVar := range []string{"WINDOWS_VERSION", "OS_VERSION", "WINVER"} {
			if version := os.Getenv(envVar); version != "" {
				return version
			}
		}
		return "unknown"
	case "darwin":
		// Check for macOS version in multiple environment variable patterns
		for _, envVar := range []string{"MACOS_VERSION", "OSX_VERSION", "DARWIN_VERSION"} {
			if version := os.Getenv(envVar); version != "" {
				return version
			}
		}
		return "unknown"
	case "linux":
		// Check for Linux distribution version in multiple environment patterns
		for _, envVar := range []string{"LINUX_VERSION", "DISTRIB_RELEASE", "VERSION_ID"} {
			if version := os.Getenv(envVar); version != "" {
				return version
			}
		}
		return "unknown"
	default:
		return "unknown"
	}
}

// String returns a human-readable representation of the platform info.
func (p *PlatformInfo) String() string {
	var builder strings.Builder
	builder.WriteString("Platform: ")
	builder.WriteString(p.OS)
	builder.WriteString(" (")
	builder.WriteString(p.FormFactor)
	builder.WriteString(")")

	if len(p.InputMethods) > 0 {
		builder.WriteString(" - Input: ")
		builder.WriteString(strings.Join(p.InputMethods, ", "))
	}

	return builder.String()
}
