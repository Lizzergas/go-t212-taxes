#!/bin/bash

# CI Test Script - Mimics the CI quality gate logic
# This helps test locally what the CI will do

set -e

echo "🧪 Testing CI quality gate logic locally..."
echo

# Run tests first
echo "📋 Running tests..."
go test ./...
echo "✅ Tests passed"
echo

# Check golangci-lint availability
if ! command -v golangci-lint &> /dev/null; then
    echo "❌ golangci-lint not found. Please install it first."
    echo "Run: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$(go env GOPATH)/bin v2.1.6"
    exit 1
fi

echo "🔍 Running golangci-lint (version: $(golangci-lint version | head -1))..."

# Run lint and capture both stdout and stderr, allow failure
LINT_OUTPUT=$(golangci-lint run --timeout=10m 2>&1 || true)

# Count issues by looking for lines that match the error pattern
# Format: "file:line:col: message (linter)"
ISSUE_COUNT=$(echo "$LINT_OUTPUT" | grep -c "^[^:]*:[0-9]*:[0-9]*:" || echo "0")

# Also try to extract from summary line if present
# Format: "X issues:"
SUMMARY_COUNT=$(echo "$LINT_OUTPUT" | grep -o "[0-9]* issues:" | grep -o "[0-9]*" || echo "")

# Use summary count if available, otherwise use line count
if [[ "$SUMMARY_COUNT" =~ ^[0-9]+$ ]] && [ "$SUMMARY_COUNT" -gt 0 ]; then
  ISSUE_COUNT=$SUMMARY_COUNT
fi

# Ensure we have a valid number
if ! [[ "$ISSUE_COUNT" =~ ^[0-9]+$ ]]; then
  ISSUE_COUNT=0
fi

echo "Found $ISSUE_COUNT linting issues"

# Show issues if any were found
if [ "$ISSUE_COUNT" -gt 0 ]; then
  echo
  echo "📋 Linting issues found:"
  echo "------------------------"
  echo "$LINT_OUTPUT"
  echo
fi

# Quality gate: Allow up to 15 issues (current: 5)
THRESHOLD=15
if [ "$ISSUE_COUNT" -le "$THRESHOLD" ]; then
  echo "✅ Quality gate PASSED: $ISSUE_COUNT issues (≤ $THRESHOLD allowed)"
  echo
  echo "🎉 This commit will pass CI!"
  
  if [ "$ISSUE_COUNT" -gt 0 ]; then
    echo "💡 Consider fixing the $ISSUE_COUNT remaining issues for better code quality"
  fi
  
  exit 0
else
  echo "❌ Quality gate FAILED: $ISSUE_COUNT issues (> $THRESHOLD allowed)" 
  echo
  echo "🚨 This commit will FAIL CI!"
  exit 1
fi 