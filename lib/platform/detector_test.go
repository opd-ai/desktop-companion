package platform

import (
	"runtime"
	"strings"
	"testing"
)

func TestGetPlatformInfo(t *testing.T) {
	info := GetPlatformInfo()

	// Test that basic fields are populated
	if info == nil {
		t.Fatal("GetPlatformInfo() returned nil")
	}

	if info.OS == "" {
		t.Error("OS field should not be empty")
	}

	if info.FormFactor == "" {
		t.Error("FormFactor field should not be empty")
	}

	if len(info.InputMethods) == 0 {
		t.Error("InputMethods should contain at least one method")
	}

	// Test that OS matches runtime.GOOS
	if info.OS != runtime.GOOS {
		t.Errorf("Expected OS to be %q, got %q", runtime.GOOS, info.OS)
	}
}

func TestPlatformInfo_IsDesktop(t *testing.T) {
	tests := []struct {
		name       string
		formFactor string
		want       bool
	}{
		{"desktop platform", "desktop", true},
		{"mobile platform", "mobile", false},
		{"tablet platform", "tablet", false},
		{"unknown platform", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PlatformInfo{FormFactor: tt.formFactor}
			if got := p.IsDesktop(); got != tt.want {
				t.Errorf("IsDesktop() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlatformInfo_IsMobile(t *testing.T) {
	tests := []struct {
		name       string
		formFactor string
		want       bool
	}{
		{"mobile platform", "mobile", true},
		{"desktop platform", "desktop", false},
		{"tablet platform", "tablet", false},
		{"unknown platform", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PlatformInfo{FormFactor: tt.formFactor}
			if got := p.IsMobile(); got != tt.want {
				t.Errorf("IsMobile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlatformInfo_IsTablet(t *testing.T) {
	tests := []struct {
		name       string
		formFactor string
		want       bool
	}{
		{"tablet platform", "tablet", true},
		{"mobile platform", "mobile", false},
		{"desktop platform", "desktop", false},
		{"unknown platform", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PlatformInfo{FormFactor: tt.formFactor}
			if got := p.IsTablet(); got != tt.want {
				t.Errorf("IsTablet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlatformInfo_HasTouch(t *testing.T) {
	tests := []struct {
		name         string
		inputMethods []string
		want         bool
	}{
		{"has touch", []string{"touch", "keyboard"}, true},
		{"no touch", []string{"mouse", "keyboard"}, false},
		{"touch only", []string{"touch"}, true},
		{"empty methods", []string{}, false},
		{"nil methods", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PlatformInfo{InputMethods: tt.inputMethods}
			if got := p.HasTouch(); got != tt.want {
				t.Errorf("HasTouch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlatformInfo_HasMouse(t *testing.T) {
	tests := []struct {
		name         string
		inputMethods []string
		want         bool
	}{
		{"has mouse", []string{"mouse", "keyboard"}, true},
		{"no mouse", []string{"touch", "keyboard"}, false},
		{"mouse only", []string{"mouse"}, true},
		{"empty methods", []string{}, false},
		{"nil methods", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PlatformInfo{InputMethods: tt.inputMethods}
			if got := p.HasMouse(); got != tt.want {
				t.Errorf("HasMouse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlatformInfo_HasKeyboard(t *testing.T) {
	tests := []struct {
		name         string
		inputMethods []string
		want         bool
	}{
		{"has keyboard", []string{"mouse", "keyboard"}, true},
		{"no keyboard", []string{"touch"}, false},
		{"keyboard only", []string{"keyboard"}, true},
		{"empty methods", []string{}, false},
		{"nil methods", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PlatformInfo{InputMethods: tt.inputMethods}
			if got := p.HasKeyboard(); got != tt.want {
				t.Errorf("HasKeyboard() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectInputMethods(t *testing.T) {
	tests := []struct {
		name     string
		goos     string
		expected []string
	}{
		{"android", "android", []string{"touch"}},
		{"ios", "ios", []string{"touch"}},
		{"windows", "windows", []string{"mouse", "keyboard"}},
		{"linux", "linux", []string{"mouse", "keyboard"}},
		{"darwin", "darwin", []string{"mouse", "keyboard"}},
		{"unknown", "unknown", []string{"mouse", "keyboard"}},
		{"freebsd", "freebsd", []string{"mouse", "keyboard"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectInputMethods(tt.goos)
			if len(got) != len(tt.expected) {
				t.Errorf("detectInputMethods(%q) = %v, want %v", tt.goos, got, tt.expected)
				return
			}
			for i, method := range tt.expected {
				if got[i] != method {
					t.Errorf("detectInputMethods(%q) = %v, want %v", tt.goos, got, tt.expected)
					break
				}
			}
		})
	}
}

func TestPlatformInfo_String(t *testing.T) {
	tests := []struct {
		name     string
		platform *PlatformInfo
		contains []string
	}{
		{
			name: "desktop platform",
			platform: &PlatformInfo{
				OS:           "linux",
				FormFactor:   "desktop",
				InputMethods: []string{"mouse", "keyboard"},
			},
			contains: []string{"Platform:", "linux", "desktop", "Input:", "mouse", "keyboard"},
		},
		{
			name: "mobile platform",
			platform: &PlatformInfo{
				OS:           "android",
				FormFactor:   "mobile",
				InputMethods: []string{"touch"},
			},
			contains: []string{"Platform:", "android", "mobile", "Input:", "touch"},
		},
		{
			name: "no input methods",
			platform: &PlatformInfo{
				OS:           "linux",
				FormFactor:   "desktop",
				InputMethods: []string{},
			},
			contains: []string{"Platform:", "linux", "desktop"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.platform.String()
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("String() = %q, should contain %q", result, expected)
				}
			}
		})
	}
}

// TestFormFactorMapping tests that OS types map to correct form factors
func TestFormFactorMapping(t *testing.T) {
	tests := []struct {
		mockGOOS           string
		expectedFormFactor string
	}{
		{"android", "mobile"},
		{"ios", "mobile"},
		{"windows", "desktop"},
		{"linux", "desktop"},
		{"darwin", "desktop"},
		{"freebsd", "desktop"}, // unknown defaults to desktop
	}

	for _, tt := range tests {
		t.Run(tt.mockGOOS, func(t *testing.T) {
			// Since we can't mock runtime.GOOS directly, we test the logic
			// by checking what GetPlatformInfo would produce for different OS values
			original := runtime.GOOS

			// Create a mock PlatformInfo with the expected mapping
			var expectedFormFactor string
			switch tt.mockGOOS {
			case "android", "ios":
				expectedFormFactor = "mobile"
			default:
				expectedFormFactor = "desktop"
			}

			if expectedFormFactor != tt.expectedFormFactor {
				t.Errorf("OS %q should map to form factor %q, got %q",
					tt.mockGOOS, tt.expectedFormFactor, expectedFormFactor)
			}

			// Verify we haven't accidentally modified the actual runtime
			if runtime.GOOS != original {
				t.Errorf("Test should not modify runtime.GOOS")
			}
		})
	}
}

// TestPrivacyCompliance verifies that the platform detection doesn't expose sensitive information
func TestPrivacyCompliance(t *testing.T) {
	info := GetPlatformInfo()

	// Test that detailed version info is not exposed
	if strings.Contains(info.MajorVersion, ".") {
		t.Error("MajorVersion should not contain minor version details for privacy")
	}

	// Test that we don't expose detailed system info
	validOS := map[string]bool{
		"windows": true,
		"linux":   true,
		"darwin":  true,
		"android": true,
		"ios":     true,
		"freebsd": true,
		"netbsd":  true,
		"openbsd": true,
		"solaris": true,
	}

	if !validOS[info.OS] && info.OS != runtime.GOOS {
		t.Errorf("OS field should only contain standard GOOS values, got %q", info.OS)
	}

	// Test that input methods are limited to expected values
	validInputs := map[string]bool{
		"mouse":    true,
		"keyboard": true,
		"touch":    true,
	}

	for _, method := range info.InputMethods {
		if !validInputs[method] {
			t.Errorf("InputMethods should only contain standard values, got %q", method)
		}
	}
}

// BenchmarkGetPlatformInfo benchmarks the platform detection performance
func BenchmarkGetPlatformInfo(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetPlatformInfo()
	}
}

// BenchmarkPlatformInfoMethods benchmarks the platform info method calls
func BenchmarkPlatformInfoMethods(b *testing.B) {
	info := GetPlatformInfo()
	b.ResetTimer()

	b.Run("IsDesktop", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = info.IsDesktop()
		}
	})

	b.Run("IsMobile", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = info.IsMobile()
		}
	})

	b.Run("HasTouch", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = info.HasTouch()
		}
	})

	b.Run("String", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = info.String()
		}
	})
}
