#!/bin/bash

# Release script for prompter CLI
# Usage: ./scripts/release.sh <version>

set -e

if [ $# -ne 1 ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 1.0.0"
    exit 1
fi

VERSION=$1
BUILD_DIR="dist"

echo "Preparing release for version $VERSION"

# Build all platforms
echo "Building cross-platform binaries..."
./scripts/build.sh "$VERSION"

# Generate checksums
echo "Generating checksums..."
cd "$BUILD_DIR"

# Calculate SHA256 checksums
SHA256_DARWIN_AMD64=$(shasum -a 256 prompter-darwin-amd64 | cut -d' ' -f1)
SHA256_DARWIN_ARM64=$(shasum -a 256 prompter-darwin-arm64 | cut -d' ' -f1)
SHA256_LINUX_AMD64=$(shasum -a 256 prompter-linux-amd64 | cut -d' ' -f1)
SHA256_LINUX_ARM64=$(shasum -a 256 prompter-linux-arm64 | cut -d' ' -f1)

echo "Checksums:"
echo "  Darwin AMD64: $SHA256_DARWIN_AMD64"
echo "  Darwin ARM64: $SHA256_DARWIN_ARM64"
echo "  Linux AMD64:  $SHA256_LINUX_AMD64"
echo "  Linux ARM64:  $SHA256_LINUX_ARM64"

# Generate Homebrew formula
echo "Generating Homebrew formula..."
cd ..
cp homebrew/prompter.rb.template "homebrew/prompter-$VERSION.rb"

# Replace placeholders in Homebrew formula
sed -i.bak "s/{{VERSION}}/$VERSION/g" "homebrew/prompter-$VERSION.rb"
sed -i.bak "s/{{SHA256_DARWIN_AMD64}}/$SHA256_DARWIN_AMD64/g" "homebrew/prompter-$VERSION.rb"
sed -i.bak "s/{{SHA256_DARWIN_ARM64}}/$SHA256_DARWIN_ARM64/g" "homebrew/prompter-$VERSION.rb"
sed -i.bak "s/{{SHA256_LINUX_AMD64}}/$SHA256_LINUX_AMD64/g" "homebrew/prompter-$VERSION.rb"
sed -i.bak "s/{{SHA256_LINUX_ARM64}}/$SHA256_LINUX_ARM64/g" "homebrew/prompter-$VERSION.rb"

# Clean up backup files
rm "homebrew/prompter-$VERSION.rb.bak"

echo "Release preparation complete!"
echo ""
echo "Files generated:"
echo "  Binaries: $BUILD_DIR/prompter-*"
echo "  Homebrew formula: homebrew/prompter-$VERSION.rb"
echo ""
echo "Next steps:"
echo "1. Create a GitHub release with tag v$VERSION"
echo "2. Upload the binaries from $BUILD_DIR/ to the release"
echo "3. Update your Homebrew tap with the generated formula"
echo "4. Test the installation: brew install your-tap/prompter"