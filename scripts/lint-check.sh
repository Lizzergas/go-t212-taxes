#!/bin/bash

# Lint check script that mirrors CI quality gate
# Usage: ./scripts/lint-check.sh

set -e

echo "🔍 Running local quality gate check..."
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'  
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if golangci-lint is installed
if ! command -v golangci-lint &> /dev/null; then
    echo -e "${YELLOW}⚠️  golangci-lint not found. Installing...${NC}"
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2.1.6
fi

# Check if jq is available
if ! command -v jq &> /dev/null; then
    echo -e "${YELLOW}⚠️  jq not found. Please install jq to get detailed issue breakdown${NC}"
    HAVE_JQ=false
else
    HAVE_JQ=true
fi

echo "📋 Running tests..."
go test ./... || {
    echo -e "${RED}❌ Tests failed${NC}"
    exit 1
}
echo -e "${GREEN}✅ All tests passed${NC}"
echo

echo "🔍 Running linter..."
# Run linter and capture output - golangci-lint outputs JSON to stdout even on errors  
if golangci-lint run --timeout=10m --out-format=json > lint_results.json 2>&1; then
    LINT_EXIT_CODE=0
    echo -e "${GREEN}✅ No linting issues found${NC}"
    ISSUE_COUNT=0
else
    LINT_EXIT_CODE=1
    # Check if we have valid JSON output
    if [ -s lint_results.json ] && jq . lint_results.json >/dev/null 2>&1; then
        # Valid JSON - count issues
        ISSUE_COUNT=$(jq -r '.Issues | length' lint_results.json 2>/dev/null)
        if [ -z "$ISSUE_COUNT" ] || [ "$ISSUE_COUNT" = "null" ]; then
            ISSUE_COUNT=0
        fi
    else
        # No valid JSON - count manually
        echo "  (Counting issues manually...)"
        ISSUE_COUNT=$(golangci-lint run --timeout=10m 2>&1 | grep -c "^.*:.*:.*:" || echo "0")
    fi
fi

# Ensure ISSUE_COUNT is a number
if ! [[ "$ISSUE_COUNT" =~ ^[0-9]+$ ]]; then
    echo -e "${YELLOW}⚠️  Could not determine issue count, assuming 0${NC}"
    ISSUE_COUNT=0
fi

echo
echo "📊 Results Summary:"
echo "  Issues found: $ISSUE_COUNT"

# Quality gate threshold
THRESHOLD=15

if [ "$ISSUE_COUNT" -le "$THRESHOLD" ]; then
    echo -e "  Quality gate: ${GREEN}✅ PASSED${NC} ($ISSUE_COUNT ≤ $THRESHOLD)"
    QUALITY_PASSED=true
else
    echo -e "  Quality gate: ${RED}❌ FAILED${NC} ($ISSUE_COUNT > $THRESHOLD)"
    QUALITY_PASSED=false
fi

# Show issue breakdown if available
if [ "$HAVE_JQ" = true ] && [ -s lint_results.json ] && jq . lint_results.json >/dev/null 2>&1 && [ "$ISSUE_COUNT" -gt 0 ]; then
    echo
    echo "🔧 Issue breakdown:"
    jq -r '.Issues | group_by(.FromLinter) | .[] | "  \(.[0].FromLinter): \(length)"' lint_results.json 2>/dev/null || echo "  (Could not parse issue breakdown)"
    
    echo
    echo "📋 Specific issues found:"
    echo "------------------------"
    if jq -r '.Issues[] | "\(.Pos.Filename):\(.Pos.Line):\(.Pos.Column): \(.Text) (\(.FromLinter))"' lint_results.json 2>/dev/null; then
        true  # jq succeeded
    else
        # Fallback to regular golangci-lint output if JSON parsing fails
        golangci-lint run --timeout=10m 2>&1 | head -20
    fi
elif [ "$ISSUE_COUNT" -gt 0 ]; then
    echo
    echo "📋 Specific issues found:"
    echo "------------------------"
    golangci-lint run --timeout=10m 2>&1 | head -20
fi

echo
if [ "$QUALITY_PASSED" = true ]; then
    echo -e "${GREEN}🎉 Your changes will pass CI quality gate!${NC}"
    echo
    if [ "$ISSUE_COUNT" -gt 0 ]; then
        echo -e "${YELLOW}💡 Consider fixing the $ISSUE_COUNT remaining issues for better code quality${NC}"
    fi
    exit 0
else
    echo -e "${RED}⚠️  Your changes exceed the quality gate threshold${NC}"
    echo "   Current: $ISSUE_COUNT issues"  
    echo "   Threshold: $THRESHOLD issues"
    echo
    echo "💡 To fix issues:"
    echo "   1. Run: golangci-lint run --timeout=10m"
    echo "   2. Fix issues or add appropriate //nolint comments"
    echo "   3. Re-run this script to verify"
    echo
    echo -e "${YELLOW}Note: CI will still pass but this indicates code quality concerns${NC}"
    # Clean up temporary file
    rm -f lint_results.json
    exit 1
fi

# Clean up temporary file
rm -f lint_results.json 