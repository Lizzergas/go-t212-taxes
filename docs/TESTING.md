# Testing Guide 🧪

This document provides comprehensive guidance for testing the T212 Taxes application. It covers all types of tests, tools, and best practices for maintaining code quality.

## Quick Start

### One-Command Testing

The simplest way to run all tests:

```bash
./scripts/test-all.sh
```

This runs the complete test suite including unit tests, linting, security scanning, coverage reports, build verification, and integration tests.

### Quick Testing (CI-like)

For faster feedback during development:

```bash
./scripts/test-all.sh --quick
```

## Testing Scripts

### 🚀 `test-all.sh` - Comprehensive Test Suite

The main testing script that runs all quality checks in the correct order:

```bash
# Full test suite
./scripts/test-all.sh

# Available options
./scripts/test-all.sh --help
./scripts/test-all.sh --quick           # Skip security and coverage
./scripts/test-all.sh --verbose         # Detailed output
./scripts/test-all.sh --skip-lint       # Skip linting
./scripts/test-all.sh --skip-security   # Skip security scan
./scripts/test-all.sh --skip-coverage   # Skip coverage reports
```

**What it does:**
1. ✅ Checks prerequisites (Go, linting tools)
2. 🧹 Cleans previous test artifacts
3. 📦 Updates dependencies
4. 🧪 Runs unit tests with race detection
5. 📊 Generates coverage reports (text + HTML)
6. 🔍 Runs linting with quality gate
7. 🔒 Performs security scanning
8. 🏗️ Tests build process
9. ⚙️ Tests version commands
10. 🔗 Runs basic integration tests
11. 🧹 Cleans up test artifacts

### 🔍 `lint-check.sh` - Quality Gate

Runs linting with the project's quality gate (max 15 issues):

```bash
./scripts/lint-check.sh
```

### 🧪 `ci-test.sh` - CI Simulation

Simulates the full CI pipeline locally:

```bash
./scripts/ci-test.sh
```

## Individual Testing Commands

### Unit Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with race detection
go test -race ./...

# Run specific package
go test ./internal/domain/calculator/...
go test ./internal/domain/parser/...
go test ./internal/app/cli/...
```

### Test Coverage

```bash
# Generate coverage profile
go test -coverprofile=coverage.out ./...

# View coverage summary
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# View coverage for specific package
go test -cover ./internal/domain/calculator/...
```

### Benchmarks

```bash
# Run all benchmarks
go test -bench=. ./...

# Run benchmarks for specific package
go test -bench=. ./internal/domain/calculator/...

# Run benchmarks with memory allocation stats
go test -bench=. -benchmem ./...
```

### Linting

```bash
# Run full linting suite
golangci-lint run --timeout=10m

# Run specific linters
golangci-lint run --disable-all --enable=govet
golangci-lint run --disable-all --enable=staticcheck

# Fix auto-fixable issues
golangci-lint run --fix
```

### Security Scanning

```bash
# Run security scan
gosec ./...

# Quiet mode (only show issues)
gosec -quiet ./...

# Generate JSON report
gosec -fmt json -out gosec-report.json ./...
```

### Build Testing

```bash
# Test build
go build -o t212-taxes ./cmd/t212-taxes

# Test cross-platform builds
GOOS=windows GOARCH=amd64 go build -o t212-taxes.exe ./cmd/t212-taxes
GOOS=darwin GOARCH=arm64 go build -o t212-taxes-mac ./cmd/t212-taxes
GOOS=linux GOARCH=amd64 go build -o t212-taxes-linux ./cmd/t212-taxes

# Test with build flags (like release builds)
go build -ldflags="-s -w" -o t212-taxes ./cmd/t212-taxes
```

## Test Categories

### 🧪 Unit Tests

**Location**: `*_test.go` files alongside source code

**Coverage Areas:**
- **Domain Logic** (`internal/domain/`): Business rules, calculations, parsing
- **Application Logic** (`internal/app/`): CLI commands, TUI components
- **Types** (`internal/domain/types/`): Data structure validation

**Key Test Suites:**
- `calculator/financial_calculator_test.go` - Tax calculations
- `calculator/income_calculator_test.go` - Income analysis
- `calculator/portfolio_calculator_test.go` - Portfolio valuation
- `parser/parser_test.go` - CSV parsing and validation
- `cli/commands_test.go` - CLI command structure
- `tui/tui_test.go` - TUI component behavior

### 🔗 Integration Tests

**What's Tested:**
- Command-line interface functionality
- End-to-end CSV processing workflows
- Configuration loading and validation
- Error handling across components

**Example Integration Tests:**
```bash
# Test complete processing workflow
./t212-taxes process --dir ./data/sample/

# Test TUI launch (requires manual verification)
./t212-taxes analyze --dir ./data/sample/

# Test all commands have help
./t212-taxes process --help
./t212-taxes analyze --help
./t212-taxes validate --help
```

### 📊 Performance Tests

```bash
# Benchmark critical paths
go test -bench=BenchmarkCalculateYearlyReports ./internal/domain/calculator/
go test -bench=BenchmarkParseCSV ./internal/domain/parser/

# Memory profiling
go test -bench=. -memprofile=mem.prof ./internal/domain/calculator/
go tool pprof mem.prof

# CPU profiling
go test -bench=. -cpuprofile=cpu.prof ./internal/domain/calculator/
go tool pprof cpu.prof
```

### 🔒 Security Tests

**Tools:**
- `gosec` - Static security analyzer
- `nancy` - Dependency vulnerability scanner (optional)

```bash
# Security scan
gosec ./...

# Check for vulnerable dependencies (if nancy is installed)
go list -json -deps ./... | nancy sleuth
```

## Quality Gates

### Coverage Requirements

- **Overall Coverage**: Target 40%+ (currently 44.1%)
- **Domain Logic**: Target 70%+ (currently 75.2%)
- **Parser Logic**: Target 70%+ (currently 71.0%)
- **Critical Paths**: Target 90%+

### Linting Standards

- **Quality Gate**: Maximum 15 linting issues
- **Current Status**: 5 issues (well within limits)
- **Zero Tolerance**: Security issues, data races, nil pointer dereferences

### Performance Benchmarks

- **CSV Parsing**: <100ms for typical files (<10MB)
- **Tax Calculations**: <50ms for yearly reports
- **Memory Usage**: <100MB for large datasets

## Testing Best Practices

### Writing Good Tests

```go
// ✅ Good: Table-driven tests
func TestCalculateCapitalGains(t *testing.T) {
    tests := []struct {
        name     string
        input    []Transaction
        expected float64
        wantErr  bool
    }{
        {
            name: "simple buy-sell scenario",
            input: []Transaction{
                {Action: "Market buy", Shares: 100, PricePerShare: 10.0},
                {Action: "Market sell", Shares: 100, PricePerShare: 15.0},
            },
            expected: 500.0,
            wantErr:  false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := CalculateCapitalGains(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("CalculateCapitalGains() error = %v, wantErr %v", err, tt.wantErr)
            }
            if result != tt.expected {
                t.Errorf("CalculateCapitalGains() = %v, want %v", result, tt.expected)
            }
        })
    }
}
```

### Test Organization

```
internal/
├── app/
│   ├── cli/
│   │   ├── commands.go
│   │   └── commands_test.go        # CLI command tests
│   └── tui/
│       ├── app.go
│       └── tui_test.go            # TUI component tests
├── domain/
│   ├── calculator/
│   │   ├── financial_calculator.go
│   │   ├── financial_calculator_test.go  # Business logic tests
│   │   ├── income_calculator.go
│   │   └── income_calculator_test.go
│   └── parser/
│       ├── parser.go
│       └── parser_test.go         # CSV parsing tests
```

### Mocking and Test Data

```go
// Use interfaces for testability
type Calculator interface {
    Calculate(transactions []Transaction) (*Report, error)
}

// Test with sample data
func loadTestData(t *testing.T, filename string) []Transaction {
    data, err := os.ReadFile(filepath.Join("testdata", filename))
    if err != nil {
        t.Fatalf("Failed to load test data: %v", err)
    }
    // Parse and return test transactions
}
```

## Continuous Integration

### GitHub Actions Integration

The `test-all.sh` script is designed to work seamlessly with CI/CD:

```yaml
# Example GitHub Actions usage
- name: Run comprehensive tests
  run: ./scripts/test-all.sh --verbose

# Or for faster CI feedback
- name: Quick test run
  run: ./scripts/test-all.sh --quick
```

### Local Pre-commit Hooks

Set up automatic testing before commits:

```bash
# .git/hooks/pre-commit
#!/bin/bash
exec ./scripts/test-all.sh --quick
```

## Troubleshooting

### Common Issues

**Tests fail with "command not found":**
```bash
# Install missing tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/securego/gosec/v2/cmd/gosec@latest
```

**Race condition errors:**
```bash
# Run with race detection to identify issues
go test -race ./...
```

**Coverage reports not generating:**
```bash
# Ensure you have write permissions
chmod 755 .
go test -coverprofile=coverage.out ./...
```

**Build fails on different platforms:**
```bash
# Test cross-compilation
GOOS=windows go build ./cmd/t212-taxes
GOOS=darwin go build ./cmd/t212-taxes
GOOS=linux go build ./cmd/t212-taxes
```

### Performance Issues

**Slow test execution:**
```bash
# Run tests in parallel
go test -parallel 4 ./...

# Skip slow integration tests during development
go test -short ./...
```

**High memory usage:**
```bash
# Profile memory usage
go test -memprofile=mem.prof -bench=. ./...
go tool pprof mem.prof
```

## AI Assistant Guidelines

When working with this codebase, AI assistants should:

### 🤖 Quick Testing Commands

```bash
# Always run this first to check everything
./scripts/test-all.sh --quick

# For comprehensive analysis
./scripts/test-all.sh --verbose

# If you need to skip certain checks
./scripts/test-all.sh --skip-security --skip-coverage
```

### 🔧 Development Workflow

1. **Before making changes**: `./scripts/test-all.sh --quick`
2. **After making changes**: `./scripts/test-all.sh`
3. **Before committing**: `./scripts/lint-check.sh`

### 📋 Test Categories to Consider

- **Unit Tests**: Test individual functions and methods
- **Integration Tests**: Test component interactions
- **Build Tests**: Ensure code compiles correctly
- **CLI Tests**: Verify command-line interface works
- **Version Tests**: Ensure version command functions

### 🎯 Focus Areas

When adding new features, ensure tests cover:
- **Happy path scenarios**
- **Error conditions**
- **Edge cases** (empty data, invalid input)
- **Performance** (for data processing functions)
- **Security** (input validation, data sanitization)

## Contributing

### Test Requirements for PRs

All pull requests must:
1. ✅ Pass all existing tests (`./scripts/test-all.sh`)
2. ✅ Include tests for new functionality
3. ✅ Maintain or improve coverage percentages
4. ✅ Pass linting quality gate
5. ✅ Include integration tests for new commands/features

### Adding New Tests

1. **Unit Tests**: Add alongside source code (`*_test.go`)
2. **Integration Tests**: Add to `scripts/test-all.sh`
3. **Test Data**: Place in `testdata/` directories
4. **Documentation**: Update this guide for new test categories

---

## Quick Reference

| Command | Purpose |
|---------|---------|
| `./scripts/test-all.sh` | Complete test suite |
| `./scripts/test-all.sh --quick` | Fast development testing |
| `./scripts/lint-check.sh` | Quality gate check |
| `go test ./...` | Unit tests only |
| `go test -cover ./...` | Tests with coverage |
| `go test -race ./...` | Race condition detection |
| `golangci-lint run` | Linting only |
| `gosec ./...` | Security scan only |

**Remember**: Always run `./scripts/test-all.sh` before submitting code changes! 