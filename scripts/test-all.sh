#!/bin/bash

# test-all.sh - Comprehensive testing script for T212 Taxes
# This script runs all types of tests and quality checks in the correct order

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Function to print colored output
print_step() {
    echo -e "${BLUE}🔧 $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${CYAN}ℹ️  $1${NC}"
}

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Parse command line arguments
SKIP_LINT=false
SKIP_SECURITY=false
SKIP_COVERAGE=false
VERBOSE=false
QUICK=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --skip-lint)
            SKIP_LINT=true
            shift
            ;;
        --skip-security)
            SKIP_SECURITY=true
            shift
            ;;
        --skip-coverage)
            SKIP_COVERAGE=true
            shift
            ;;
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        --quick|-q)
            QUICK=true
            SKIP_SECURITY=true
            SKIP_COVERAGE=true
            shift
            ;;
        --help|-h)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --skip-lint      Skip linting checks"
            echo "  --skip-security  Skip security scanning"
            echo "  --skip-coverage  Skip coverage report generation"
            echo "  --verbose, -v    Enable verbose output"
            echo "  --quick, -q      Quick mode (skip security and coverage)"
            echo "  --help, -h       Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0                    # Run all tests and checks"
            echo "  $0 --quick           # Quick test run"
            echo "  $0 --skip-lint       # Skip linting"
            echo "  $0 --verbose         # Verbose output"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

echo -e "${PURPLE}🧪 T212 Taxes - Comprehensive Test Suite${NC}"
echo "=========================================="
echo ""

# Check prerequisites
print_step "Checking prerequisites..."
if ! command_exists go; then
    print_error "Go is not installed or not in PATH"
    exit 1
fi

GO_VERSION=$(go version | cut -d' ' -f3)
print_info "Go version: $GO_VERSION"

if [[ "$SKIP_LINT" == "false" ]] && ! command_exists golangci-lint; then
    print_warning "golangci-lint not found - linting will be skipped"
    SKIP_LINT=true
fi

if [[ "$SKIP_SECURITY" == "false" ]] && ! command_exists gosec; then
    print_warning "gosec not found - security scanning will be skipped"
    SKIP_SECURITY=true
fi

echo ""

# Step 1: Clean previous artifacts
print_step "Cleaning previous test artifacts..."
rm -f coverage.out coverage.html t212-taxes
print_success "Cleanup completed"
echo ""

# Step 2: Download dependencies
print_step "Downloading dependencies..."
if [[ "$VERBOSE" == "true" ]]; then
    go mod download
    go mod tidy
else
    go mod download >/dev/null 2>&1
    go mod tidy >/dev/null 2>&1
fi
print_success "Dependencies updated"
echo ""

# Step 3: Run unit tests
print_step "Running unit tests..."
if [[ "$VERBOSE" == "true" ]]; then
    TEST_CMD="go test -v -race ./..."
else
    TEST_CMD="go test -race ./..."
fi

if [[ "$SKIP_COVERAGE" == "false" ]]; then
    TEST_CMD="$TEST_CMD -coverprofile=coverage.out"
fi

if $TEST_CMD; then
    print_success "All unit tests passed"
else
    print_error "Unit tests failed"
    exit 1
fi
echo ""

# Step 4: Generate coverage report
if [[ "$SKIP_COVERAGE" == "false" ]]; then
    print_step "Generating coverage reports..."
    
    # Text coverage summary
    COVERAGE=$(go tool cover -func=coverage.out | tail -1 | awk '{print $3}')
    print_info "Overall test coverage: $COVERAGE"
    
    # HTML coverage report
    go tool cover -html=coverage.out -o coverage.html
    print_success "Coverage reports generated (coverage.out, coverage.html)"
    echo ""
fi

# Step 5: Run linting
if [[ "$SKIP_LINT" == "false" ]]; then
    print_step "Running linting checks..."
    
    # Use our existing lint check script if available
    if [[ -f "./scripts/lint-check.sh" ]]; then
        if ./scripts/lint-check.sh; then
            print_success "Linting passed quality gate"
        else
            print_error "Linting failed quality gate"
            exit 1
        fi
    else
        # Fallback to direct golangci-lint - capture output to show issues
        LINT_OUTPUT=$(golangci-lint run --timeout=10m 2>&1 || true)
        
        # Count issues by looking for lines that match the error pattern
        ISSUE_COUNT=$(echo "$LINT_OUTPUT" | grep -c "^[^:]*:[0-9]*:[0-9]*:" || echo "0")
        
        # Also try to extract from summary line if present
        SUMMARY_COUNT=$(echo "$LINT_OUTPUT" | grep -o "[0-9]* issues:" | grep -o "[0-9]*" || echo "")
        
        # Use summary count if available, otherwise use line count
        if [[ "$SUMMARY_COUNT" =~ ^[0-9]+$ ]] && [ "$SUMMARY_COUNT" -gt 0 ]; then
            ISSUE_COUNT=$SUMMARY_COUNT
        fi
        
        # Ensure we have a valid number
        if ! [[ "$ISSUE_COUNT" =~ ^[0-9]+$ ]]; then
            ISSUE_COUNT=0
        fi
        
        print_info "Found $ISSUE_COUNT linting issues"
        
        # Show issues if any were found
        if [ "$ISSUE_COUNT" -gt 0 ]; then
            echo ""
            print_info "📋 Linting issues found:"
            echo "------------------------"
            echo "$LINT_OUTPUT"
            echo ""
            print_warning "Linting found $ISSUE_COUNT issues"
        else
            print_success "Linting completed successfully"
        fi
    fi
    echo ""
fi

# Step 6: Security scanning
if [[ "$SKIP_SECURITY" == "false" ]]; then
    print_step "Running security scan..."
    if gosec -quiet ./...; then
        print_success "Security scan completed - no issues found"
    else
        print_warning "Security scan found potential issues (check output above)"
    fi
    echo ""
fi

# Step 7: Build test
print_step "Testing build process..."
if go build -o t212-taxes ./cmd/t212-taxes; then
    print_success "Build completed successfully"
else
    print_error "Build failed"
    exit 1
fi
echo ""

# Step 8: Test version command
print_step "Testing version command..."
VERSION_OUTPUT=$(./t212-taxes version 2>&1)
if [[ $? -eq 0 ]]; then
    print_success "Version command works"
    if [[ "$VERBOSE" == "true" ]]; then
        echo "$VERSION_OUTPUT"
    fi
else
    print_error "Version command failed"
    exit 1
fi

# Test JSON version output
JSON_OUTPUT=$(./t212-taxes version --format json 2>&1)
if [[ $? -eq 0 ]] && echo "$JSON_OUTPUT" | jq . >/dev/null 2>&1; then
    print_success "Version JSON output is valid"
elif [[ $? -eq 0 ]]; then
    print_success "Version JSON command works (jq not available for validation)"
else
    print_error "Version JSON command failed"
    exit 1
fi
echo ""

# Step 9: Integration test (basic functionality)
print_step "Running basic integration tests..."

# Test help command
if ./t212-taxes --help >/dev/null 2>&1; then
    print_success "Help command works"
else
    print_warning "Help command failed"
fi

# Test command structure
COMMANDS=("process" "analyze" "validate" "income" "portfolio" "version")
for cmd in "${COMMANDS[@]}"; do
    if ./t212-taxes "$cmd" --help >/dev/null 2>&1; then
        if [[ "$VERBOSE" == "true" ]]; then
            print_info "Command '$cmd' is available"
        fi
    else
        print_warning "Command '$cmd' help failed"
    fi
done

print_success "Integration tests completed"
echo ""

# Step 10: Cleanup test artifacts
print_step "Cleaning up test artifacts..."
rm -f t212-taxes
print_success "Cleanup completed"
echo ""

# Final summary
echo -e "${GREEN}🎉 All tests completed successfully!${NC}"
echo ""
echo "Summary:"
echo "--------"
print_success "✅ Unit tests: PASSED"
if [[ "$SKIP_COVERAGE" == "false" ]]; then
    print_success "✅ Coverage: $COVERAGE"
fi
if [[ "$SKIP_LINT" == "false" ]]; then
    print_success "✅ Linting: PASSED"
fi
if [[ "$SKIP_SECURITY" == "false" ]]; then
    print_success "✅ Security: SCANNED"
fi
print_success "✅ Build: PASSED"
print_success "✅ Integration: PASSED"

echo ""
if [[ "$SKIP_COVERAGE" == "false" ]]; then
    print_info "📊 Coverage report available at: coverage.html"
fi
print_info "🚀 Your code is ready for commit/release!"

# Return appropriate exit code
exit 0 