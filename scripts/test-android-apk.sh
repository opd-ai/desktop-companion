#!/bin/bash
# test-android-apk.sh: Run APK integrity test for CI/CD
# Usage: ./test-android-apk.sh <apk-path> <expected-package>
set -e
APK_PATH="$1"
PKG="$2"
if [ -z "$APK_PATH" ] || [ -z "$PKG" ]; then
  echo "Usage: $0 <apk-path> <expected-package>"
  exit 2
fi
# Run Go integrity test
GO_TEST="$(dirname "$0")/apk_integrity/apk_integrity_test.go"
go run "$GO_TEST" "$APK_PATH" "$PKG"
