// apk_integrity_test.go: Automated APK integrity check for CI/CD pipeline
// Uses Go stdlib and Android SDK tools (apksigner, aapt)
// Usage: go run apk_integrity_test.go <path-to-apk> <expected-package>

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// checkFileExists verifies the APK file exists and is >0 bytes
func checkFileExists(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}
	if fi.Size() == 0 {
		return fmt.Errorf("file is empty: %s", path)
	}
	return nil
}

// checkApkSignature runs 'apksigner verify' and checks for 'Verified'
func checkApkSignature(path string) error {
	cmd := exec.Command("apksigner", "verify", path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("apksigner failed: %w", err)
	}
	if !strings.Contains(string(out), "Verified") {
		return fmt.Errorf("APK signature not verified: %s", out)
	}
	return nil
}

// checkApkPackage runs 'aapt dump badging' and checks for expected package name
func checkApkPackage(path, expected string) error {
	cmd := exec.Command("aapt", "dump", "badging", path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("aapt failed: %w", err)
	}
	if !strings.Contains(string(out), "package: name='"+expected+"'") {
		return fmt.Errorf("APK package name mismatch: got %s, want %s", out, expected)
	}
	return nil
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: go run apk_integrity_test.go <apk-path> <expected-package>")
		os.Exit(2)
	}
	apkPath := os.Args[1]
	pkg := os.Args[2]
	if err := checkFileExists(apkPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := checkApkSignature(apkPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := checkApkPackage(apkPath, pkg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("APK integrity check PASSED")
}
