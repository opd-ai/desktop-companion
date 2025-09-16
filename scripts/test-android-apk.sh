#!/bin/bash

# DEPRECATED: Legacy wrapper for test-android-apk.sh
# This script is maintained for backward compatibility.
# New usage: ./scripts/dds-scripts.sh android test-apk
# Direct usage: ./scripts/android/test-apk-build.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Forward all arguments to the new refactored script
exec "$SCRIPT_DIR/android/test-apk-build.sh" "$@"
