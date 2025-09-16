#!/bin/bash

# DEPRECATED: Legacy wrapper for cross_platform_build.sh
# This script is maintained for backward compatibility.
# New usage: ./scripts/dds-scripts.sh build cross-platform
# Direct usage: ./scripts/build/cross-platform-build.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Forward all arguments to the new refactored script
exec "$SCRIPT_DIR/build/cross-platform-build.sh" "$@"
