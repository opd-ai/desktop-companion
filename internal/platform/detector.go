// Package platform provides cross-platform detection and capability information
// for adaptive behavior between desktop and mobile environments.
package platform

import (
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
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
// Uses multiple methods including environment variables and system properties.
func detectAndroidMajorVersion() string {
	// Check environment variables first for privacy compliance
	for _, envVar := range []string{"ANDROID_VERSION", "ANDROID_API_LEVEL", "API_LEVEL"} {
		if version := os.Getenv(envVar); version != "" {
			// Extract major version from version string if set
			parts := strings.Split(version, ".")
			if len(parts) > 0 && parts[0] != "" {
				return parts[0]
			}
		}
	}

	// Try to read Android system properties if available
	if version := readAndroidProperty("ro.build.version.release"); version != "unknown" {
		parts := strings.Split(version, ".")
		if len(parts) > 0 && parts[0] != "" {
			return parts[0]
		}
	}

	// Try API level as fallback
	if apiLevel := readAndroidProperty("ro.build.version.sdk"); apiLevel != "unknown" {
		if level, err := strconv.Atoi(apiLevel); err == nil {
			// Convert API level to Android version (approximate mapping)
			return androidAPIToVersion(level)
		}
	}

	// Fallback to unknown for privacy compliance
	return "unknown"
}

// readAndroidProperty reads an Android system property using getprop command
func readAndroidProperty(property string) string {
	cmd := exec.Command("getprop", property)
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

// androidAPIToVersion maps Android API levels to major version numbers
func androidAPIToVersion(apiLevel int) string {
	switch {
	case apiLevel >= 34:
		return "14"
	case apiLevel >= 33:
		return "13"
	case apiLevel >= 31:
		return "12"
	case apiLevel >= 30:
		return "11"
	case apiLevel >= 29:
		return "10"
	case apiLevel >= 28:
		return "9"
	case apiLevel >= 26:
		return "8"
	case apiLevel >= 24:
		return "7"
	case apiLevel >= 23:
		return "6"
	default:
		return "unknown"
	}
}

// detectIOSMajorVersion attempts to detect iOS major version.
// Uses multiple methods including environment variables and system calls.
func detectIOSMajorVersion() string {
	// Check environment variables first for privacy compliance
	for _, envVar := range []string{"IOS_VERSION", "IPHONEOS_DEPLOYMENT_TARGET", "IOS_DEPLOYMENT_TARGET"} {
		if version := os.Getenv(envVar); version != "" {
			// Extract major version from version string if set
			parts := strings.Split(version, ".")
			if len(parts) > 0 && parts[0] != "" {
				return parts[0]
			}
		}
	}

	// Try to use uname or system calls for iOS version if available
	cmd := exec.Command("uname", "-r")
	output, err := cmd.Output()
	if err == nil {
		kernelVersion := strings.TrimSpace(string(output))
		// iOS kernel versions can be mapped to iOS versions (approximate)
		if iosVersion := kernelToIOSVersion(kernelVersion); iosVersion != "unknown" {
			return iosVersion
		}
	}

	// Fallback to unknown for privacy compliance
	return "unknown"
}

// kernelToIOSVersion maps kernel versions to iOS major versions (approximate)
func kernelToIOSVersion(kernelVersion string) string {
	// Parse kernel version like "21.6.0" -> iOS 15
	parts := strings.Split(kernelVersion, ".")
	if len(parts) == 0 {
		return "unknown"
	}

	majorKernel, err := strconv.Atoi(parts[0])
	if err != nil {
		return "unknown"
	}

	// Approximate mapping based on known iOS kernel versions
	switch {
	case majorKernel >= 23:
		return "17"
	case majorKernel >= 22:
		return "16"
	case majorKernel >= 21:
		return "15"
	case majorKernel >= 20:
		return "14"
	case majorKernel >= 19:
		return "13"
	case majorKernel >= 18:
		return "12"
	default:
		return "unknown"
	}
}

// detectDesktopMajorVersion attempts to detect desktop OS major version.
// Uses system-specific detection methods with privacy-conscious fallbacks.
func detectDesktopMajorVersion(goos string) string {
	// Try environment variables first for privacy compliance
	envVersion := detectVersionFromEnv(goos)
	if envVersion != "unknown" {
		return envVersion
	}

	// Try system-specific detection methods
	switch goos {
	case "windows":
		return detectWindowsVersion()
	case "darwin":
		return detectMacOSVersion()
	case "linux":
		return detectLinuxVersion()
	default:
		return "unknown"
	}
}

// detectVersionFromEnv checks environment variables for version information
func detectVersionFromEnv(goos string) string {
	var envVars []string
	switch goos {
	case "windows":
		envVars = []string{"WINDOWS_VERSION", "OS_VERSION", "WINVER"}
	case "darwin":
		envVars = []string{"MACOS_VERSION", "OSX_VERSION", "DARWIN_VERSION"}
	case "linux":
		envVars = []string{"LINUX_VERSION", "DISTRIB_RELEASE", "VERSION_ID"}
	default:
		return "unknown"
	}

	for _, envVar := range envVars {
		if version := os.Getenv(envVar); version != "" {
			parts := strings.Split(version, ".")
			if len(parts) > 0 && parts[0] != "" {
				return parts[0]
			}
		}
	}
	return "unknown"
}

// detectWindowsVersion detects Windows major version using registry or system info
func detectWindowsVersion() string {
	// Try using systeminfo command for version detection
	cmd := exec.Command("cmd", "/c", "ver")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	// Parse version from output like "Microsoft Windows [Version 10.0.19044.1889]"
	versionRegex := regexp.MustCompile(`Version\s+(\d+)\.`)
	matches := versionRegex.FindStringSubmatch(string(output))
	if len(matches) > 1 {
		return matches[1]
	}

	return "unknown"
}

// detectMacOSVersion detects macOS major version using sw_vers
func detectMacOSVersion() string {
	cmd := exec.Command("sw_vers", "-productVersion")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	version := strings.TrimSpace(string(output))
	parts := strings.Split(version, ".")
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}

	return "unknown"
}

// detectLinuxVersion detects Linux distribution major version using various methods
func detectLinuxVersion() string {
	// Try /etc/os-release first (most common)
	if version := parseOSRelease(); version != "unknown" {
		return version
	}

	// Try lsb_release command
	cmd := exec.Command("lsb_release", "-rs")
	output, err := cmd.Output()
	if err == nil {
		version := strings.TrimSpace(string(output))
		parts := strings.Split(version, ".")
		if len(parts) > 0 && parts[0] != "" {
			return parts[0]
		}
	}

	return "unknown"
}

// parseOSRelease parses /etc/os-release for version information
func parseOSRelease() string {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "unknown"
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "VERSION_ID=") {
			version := strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), "\"")
			parts := strings.Split(version, ".")
			if len(parts) > 0 && parts[0] != "" {
				return parts[0]
			}
		}
	}

	return "unknown"
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
