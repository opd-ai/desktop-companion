#!/bin/bash

# DEPRECATED: Legacy wrapper for validate-animations.sh
# This script is maintained for backward compatibility.
# New usage: ./scripts/dds-scripts.sh validation animations
# Direct usage: ./scripts/validation/validate-animations.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Forward all arguments to the new refactored script
exec "$SCRIPT_DIR/validation/validate-animations.sh" "$@"
