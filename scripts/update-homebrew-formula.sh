#!/bin/bash

# Script to update Homebrew formula with actual SHA256 checksums from GitHub release

set -e

VERSION=${1:-"v1.0.0"}
REPO="Lizzergas/go-t212-taxes"
FORMULA_FILE="Formula/t212-taxes.rb"

echo "🔍 Fetching checksums for ${VERSION}..."

# Download checksums file
curl -L -o checksums.txt "https://github.com/${REPO}/releases/download/${VERSION}/checksums.txt"

# Extract SHA256 checksums for each platform
DARWIN_AMD64_SHA=$(grep "go-t212-taxes-darwin-x86_64.tar.gz" checksums.txt | cut -d' ' -f1)
DARWIN_ARM64_SHA=$(grep "go-t212-taxes-darwin-arm64.tar.gz" checksums.txt | cut -d' ' -f1)
LINUX_AMD64_SHA=$(grep "go-t212-taxes-linux-x86_64.tar.gz" checksums.txt | cut -d' ' -f1)
LINUX_ARM64_SHA=$(grep "go-t212-taxes-linux-arm64.tar.gz" checksums.txt | cut -d' ' -f1)

echo "✅ Found checksums:"
echo "  Darwin AMD64: ${DARWIN_AMD64_SHA}"
echo "  Darwin ARM64: ${DARWIN_ARM64_SHA}"
echo "  Linux AMD64:  ${LINUX_AMD64_SHA}"
echo "  Linux ARM64:  ${LINUX_ARM64_SHA}"

# Update the formula file
echo "📝 Updating ${FORMULA_FILE}..."

# Use the first checksum as the default
sed -i.bak "s/PLACEHOLDER_SHA256\"/${DARWIN_AMD64_SHA}\"/g" "${FORMULA_FILE}"
sed -i.bak "s/PLACEHOLDER_SHA256_INTEL\"/${DARWIN_AMD64_SHA}\"/g" "${FORMULA_FILE}"
sed -i.bak "s/PLACEHOLDER_SHA256_ARM\"/${DARWIN_ARM64_SHA}\"/g" "${FORMULA_FILE}"
sed -i.bak "s/PLACEHOLDER_SHA256_LINUX_INTEL\"/${LINUX_AMD64_SHA}\"/g" "${FORMULA_FILE}"
sed -i.bak "s/PLACEHOLDER_SHA256_LINUX_ARM\"/${LINUX_ARM64_SHA}\"/g" "${FORMULA_FILE}"

# Clean up
rm -f "${FORMULA_FILE}.bak" checksums.txt

echo "✅ Formula updated successfully!"
echo ""
echo "📋 To install locally:"
echo "  brew tap Lizzergas/t212-taxes https://github.com/${REPO}"
echo "  brew install t212-taxes"
echo ""
echo "🚀 Don't forget to commit and push the updated formula!" 