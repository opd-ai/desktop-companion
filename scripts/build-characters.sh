#!/bin/bash

# DEPRECATED: Legacy wrapper for build-characters.sh
# This script is maintained for backward compatibility.
# New usage: ./scripts/dds-scripts.sh build characters
# Direct usage: ./scripts/build/build-characters.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Forward all arguments to the new refactored script
exec "$SCRIPT_DIR/build/build-characters.sh" "$@"
