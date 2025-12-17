#!/bin/bash

# Build script for cross-platform static binaries
# Usage: ./scripts/build.sh [version]

set -e

# Get version from argument or default to dev
VERSION=${1:-dev}
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build directory
BUILD_DIR="dist"
mkdir -p "$BUILD_DIR"

# Build flags for static binaries and version injection
LDFLAGS="-s -w -X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE"

echo "Building prompter $VERSION (commit: $COMMIT, date: $DATE)"

# Build for macOS (Intel)
echo "Building for macOS (Intel)..."
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o "$BUILD_DIR/prompter-darwin-amd64" ./cmd/prompter

# Build for macOS (Apple Silicon)
echo "Building for macOS (Apple Silicon)..."
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o "$BUILD_DIR/prompter-darwin-arm64" ./cmd/prompter

# Build for Linux (Intel)
echo "Building for Linux (Intel)..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o "$BUILD_DIR/prompter-linux-amd64" ./cmd/prompter

# Build for Linux (ARM64)
echo "Building for Linux (ARM64)..."
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o "$BUILD_DIR/prompter-linux-arm64" ./cmd/prompter

echo "Build complete! Binaries available in $BUILD_DIR/"
ls -la "$BUILD_DIR/"