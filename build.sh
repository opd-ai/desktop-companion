#!/bin/bash

# Desktop Companion Build Script
# Cross-platform build automation using Go's built-in tools

set -e

PROJECT_NAME="desktop-companion"
BUILD_DIR="build"
CMD_PATH="cmd/companion/main.go"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Desktop Companion Build Script${NC}"
echo "======================================"

# Clean previous builds
echo -e "${YELLOW}Cleaning previous builds...${NC}"
rm -rf $BUILD_DIR
mkdir -p $BUILD_DIR

# Download dependencies
echo -e "${YELLOW}Downloading dependencies...${NC}"
go mod download
go mod tidy

# Run tests
echo -e "${YELLOW}Running tests...${NC}"
go test ./... -v

# Build for current platform (development)
echo -e "${YELLOW}Building for current platform...${NC}"
go build -ldflags="-s -w" -o $BUILD_DIR/companion $CMD_PATH

# Cross-platform builds
echo -e "${YELLOW}Building for Windows (amd64)...${NC}"
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $BUILD_DIR/companion-windows.exe $CMD_PATH

echo -e "${YELLOW}Building for macOS (amd64)...${NC}"
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $BUILD_DIR/companion-macos $CMD_PATH

echo -e "${YELLOW}Building for macOS (arm64)...${NC}"
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $BUILD_DIR/companion-macos-arm64 $CMD_PATH

echo -e "${YELLOW}Building for Linux (amd64)...${NC}"
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $BUILD_DIR/companion-linux $CMD_PATH

# Display build results
echo -e "${GREEN}Build completed successfully!${NC}"
echo ""
echo "Built binaries:"
ls -la $BUILD_DIR/

# Display binary sizes
echo ""
echo "Binary sizes:"
du -h $BUILD_DIR/*

echo ""
echo -e "${GREEN}Ready to run!${NC}"
echo "Usage examples:"
echo "  ./$BUILD_DIR/companion"
echo "  ./$BUILD_DIR/companion -debug"
echo "  ./$BUILD_DIR/companion -character assets/characters/default/character.json"
