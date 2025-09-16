#!/bin/bash

# DEPRECATED: Legacy wrapper for validate-characters.sh
# This script is maintained for backward compatibility.
# New usage: ./scripts/dds-scripts.sh validation characters
# Direct usage: ./scripts/validation/validate-characters.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Forward all arguments to the new refactored script
exec "$SCRIPT_DIR/validation/validate-characters.sh" "$@"
