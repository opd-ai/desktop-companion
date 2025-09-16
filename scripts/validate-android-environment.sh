#!/bin/bash

# DEPRECATED: Legacy wrapper for validate-android-environment.sh
# This script is maintained for backward compatibility.
# New usage: ./scripts/dds-scripts.sh android validate-environment
# Direct usage: ./scripts/android/validate-environment.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Forward all arguments to the new refactored script
exec "$SCRIPT_DIR/android/validate-environment.sh" "$@"
