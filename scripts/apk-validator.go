package main

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// APKValidator provides validation functionality for Android APK files
type APKValidator struct {
	apkPath string
}

// AndroidManifest represents the structure of an Android manifest file
type AndroidManifest struct {
	XMLName     xml.Name `xml:"manifest"`
	Package     string   `xml:"package,attr"`
	VersionCode string   `xml:"android:versionCode,attr"`
	VersionName string   `xml:"android:versionName,attr"`
	Application struct {
		Icon  string `xml:"android:icon,attr"`
		Label string `xml:"android:label,attr"`
	} `xml:"application"`
}

// NewAPKValidator creates a new APK validator for the given file
func NewAPKValidator(apkPath string) *APKValidator {
	return &APKValidator{apkPath: apkPath}
}

// ValidateAPK performs comprehensive validation of an APK file
func (v *APKValidator) ValidateAPK() error {
	// Check if file exists and has correct extension
	if err := v.validateFileExists(); err != nil {
		return fmt.Errorf("file validation failed: %w", err)
	}

	// Validate ZIP structure (APK is a ZIP file)
	if err := v.validateZipStructure(); err != nil {
		return fmt.Errorf("ZIP structure validation failed: %w", err)
	}

	// Validate required APK components
	if err := v.validateAPKComponents(); err != nil {
		return fmt.Errorf("APK components validation failed: %w", err)
	}

	// Validate manifest
	if err := v.validateManifest(); err != nil {
		return fmt.Errorf("manifest validation failed: %w", err)
	}

	return nil
}

// validateFileExists checks if the APK file exists and has correct extension
func (v *APKValidator) validateFileExists() error {
	if _, err := os.Stat(v.apkPath); os.IsNotExist(err) {
		return fmt.Errorf("APK file does not exist: %s", v.apkPath)
	}

	if !strings.HasSuffix(strings.ToLower(v.apkPath), ".apk") {
		return fmt.Errorf("file does not have .apk extension: %s", v.apkPath)
	}

	return nil
}

// validateZipStructure validates that the APK is a valid ZIP file
func (v *APKValidator) validateZipStructure() error {
	reader, err := zip.OpenReader(v.apkPath)
	if err != nil {
		return fmt.Errorf("failed to open APK as ZIP: %w", err)
	}
	defer reader.Close()

	if len(reader.File) == 0 {
		return fmt.Errorf("APK appears to be empty")
	}

	return nil
}

// validateAPKComponents checks for required APK components
func (v *APKValidator) validateAPKComponents() error {
	reader, err := zip.OpenReader(v.apkPath)
	if err != nil {
		return fmt.Errorf("failed to open APK: %w", err)
	}
	defer reader.Close()

	requiredFiles := []string{
		"AndroidManifest.xml",
		"classes.dex",
		"META-INF/",
	}

	foundFiles := make(map[string]bool)

	for _, file := range reader.File {
		for _, required := range requiredFiles {
			if strings.HasPrefix(file.Name, required) {
				foundFiles[required] = true
			}
		}
	}

	var missingFiles []string
	for _, required := range requiredFiles {
		if !foundFiles[required] {
			missingFiles = append(missingFiles, required)
		}
	}

	if len(missingFiles) > 0 {
		return fmt.Errorf("missing required APK components: %v", missingFiles)
	}

	return nil
}

// validateManifest validates the Android manifest file
func (v *APKValidator) validateManifest() error {
	reader, err := zip.OpenReader(v.apkPath)
	if err != nil {
		return fmt.Errorf("failed to open APK: %w", err)
	}
	defer reader.Close()

	var manifestFile *zip.File
	for _, file := range reader.File {
		if file.Name == "AndroidManifest.xml" {
			manifestFile = file
			break
		}
	}

	if manifestFile == nil {
		return fmt.Errorf("AndroidManifest.xml not found in APK")
	}

	// Note: Android manifest is typically in binary XML format
	// For a complete validator, we would need to parse binary XML
	// For now, we just verify the file exists and is not empty
	if manifestFile.UncompressedSize64 == 0 {
		return fmt.Errorf("AndroidManifest.xml is empty")
	}

	return nil
}

// GetAPKInfo extracts basic information about the APK
func (v *APKValidator) GetAPKInfo() (*APKInfo, error) {
	reader, err := zip.OpenReader(v.apkPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open APK: %w", err)
	}
	defer reader.Close()

	info := &APKInfo{
		FilePath: v.apkPath,
	}

	// Get file size
	fileInfo, err := os.Stat(v.apkPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	info.SizeBytes = fileInfo.Size()

	// Count files in APK
	info.FileCount = len(reader.File)

	// Look for assets and resources
	for _, file := range reader.File {
		switch {
		case strings.HasPrefix(file.Name, "assets/"):
			info.HasAssets = true
		case strings.HasPrefix(file.Name, "res/"):
			info.HasResources = true
		case strings.HasSuffix(file.Name, ".dex"):
			info.DexFiles++
		case strings.HasPrefix(file.Name, "lib/"):
			info.HasNativeLibs = true
		}
	}

	return info, nil
}

// APKInfo contains information about an APK file
type APKInfo struct {
	FilePath      string
	SizeBytes     int64
	FileCount     int
	HasAssets     bool
	HasResources  bool
	HasNativeLibs bool
	DexFiles      int
}

// String returns a human-readable representation of APK info
func (info *APKInfo) String() string {
	return fmt.Sprintf("APK: %s (%.1f KB, %d files, %d DEX files, assets=%t, resources=%t, native=%t)",
		filepath.Base(info.FilePath),
		float64(info.SizeBytes)/1024,
		info.FileCount,
		info.DexFiles,
		info.HasAssets,
		info.HasResources,
		info.HasNativeLibs,
	)
}

// ValidateCharacterAPK validates a character-specific APK file
func ValidateCharacterAPK(apkPath, expectedCharacter string) error {
	validator := NewAPKValidator(apkPath)

	// Basic APK validation
	if err := validator.ValidateAPK(); err != nil {
		return fmt.Errorf("APK validation failed: %w", err)
	}

	// Get APK info for additional checks
	info, err := validator.GetAPKInfo()
	if err != nil {
		return fmt.Errorf("failed to get APK info: %w", err)
	}

	// Validate file size (should be reasonable)
	minSize := int64(1024 * 1024)      // 1MB minimum
	maxSize := int64(50 * 1024 * 1024) // 50MB maximum

	if info.SizeBytes < minSize {
		return fmt.Errorf("APK file too small (%d bytes), minimum %d bytes", info.SizeBytes, minSize)
	}

	if info.SizeBytes > maxSize {
		return fmt.Errorf("APK file too large (%d bytes), maximum %d bytes", info.SizeBytes, maxSize)
	}

	// Validate DEX files exist (compiled Java code)
	if info.DexFiles == 0 {
		return fmt.Errorf("no DEX files found in APK")
	}

	// Character-specific APK should have embedded assets
	if !info.HasAssets && !info.HasResources {
		return fmt.Errorf("no assets or resources found in character APK")
	}

	return nil
}

// main function for standalone APK validation tool
func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <apk-file> [character-name]\n", os.Args[0])
		fmt.Println("\nValidates Android APK files generated by the DDS character build system.")
		fmt.Println("\nExamples:")
		fmt.Println("  go run apk-validator.go build/default_android_arm64.apk")
		fmt.Println("  go run apk-validator.go build/tsundere_android_arm64.apk tsundere")
		os.Exit(1)
	}

	apkPath := os.Args[1]
	var character string
	if len(os.Args) > 2 {
		character = os.Args[2]
	}

	fmt.Printf("Validating APK: %s\n", apkPath)
	if character != "" {
		fmt.Printf("Expected character: %s\n", character)
	}
	fmt.Println()

	// Perform validation
	var err error
	if character != "" {
		err = ValidateCharacterAPK(apkPath, character)
	} else {
		validator := NewAPKValidator(apkPath)
		err = validator.ValidateAPK()
	}

	if err != nil {
		fmt.Printf("❌ Validation failed: %v\n", err)
		os.Exit(1)
	}

	// Get and display APK info
	validator := NewAPKValidator(apkPath)
	info, err := validator.GetAPKInfo()
	if err != nil {
		fmt.Printf("⚠️  Warning: Could not get APK info: %v\n", err)
	} else {
		fmt.Printf("📱 %s\n", info.String())
	}

	fmt.Println("✅ APK validation successful!")
}
