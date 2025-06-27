#!/bin/bash

# Script to update version and checksums in Homebrew formula

set -e

if [ $# -eq 0 ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 v1.0.1"
    exit 1
fi

VERSION=$1
FORMULA_FILE="Formula/t212-taxes.rb"

# Remove 'v' prefix if present
VERSION_NUM=${VERSION#v}

echo "🔄 Updating Homebrew formula to version ${VERSION_NUM}..."

# Update version in formula
sed -i.bak "s/version \".*\"/version \"${VERSION_NUM}\"/g" "${FORMULA_FILE}"

# Update URLs to point to new version
sed -i.bak "s|releases/download/v[^/]*/|releases/download/${VERSION}/|g" "${FORMULA_FILE}"

echo "📥 Fetching new checksums..."

# Run the checksum update script
./scripts/update-homebrew-formula.sh "${VERSION}"

echo "✅ Formula updated to version ${VERSION_NUM}!"
echo ""
echo "📋 Next steps:"
echo "  1. Commit the changes: git add Formula/t212-taxes.rb && git commit -m 'feat: update Homebrew formula to ${VERSION}'"
echo "  2. Push changes: git push origin main"
echo "  3. Users can update with: brew upgrade t212-taxes" 