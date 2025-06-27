# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Building and Running
```bash
# Build the application (with version info from git)
make build

# Build for development (dynamic version detection)
make dev

# Simple go build (also uses dynamic version detection)
go build -o t212-taxes ./cmd/t212-taxes

# Run the application
go run ./cmd/t212-taxes

# Build for multiple platforms
make build-all
```

### Version Information
The application automatically detects version information:
- **With ldflags**: Uses build-time injected values (releases, make build)
- **Without ldflags**: Dynamically detects from git (development, go build)
- **All local builds**: Show "Built by: lizz"

```bash
# Check version info
./t212-taxes version
./t212-taxes version --format json
```

### Testing

#### Quick Testing (Recommended)
```bash
# Run complete test suite - use this for comprehensive testing
./scripts/test-all.sh

# Quick development testing (faster feedback)
./scripts/test-all.sh --quick

# Verbose output for debugging
./scripts/test-all.sh --verbose
```

#### Individual Testing Commands
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out

# Run benchmarks
go test -bench=. ./...

# Run specific test suite
go test ./internal/domain/calculator/...
go test ./internal/domain/parser/...
```

📖 **See [Testing Guide](docs/TESTING.md) for comprehensive testing documentation**
📖 **See [CI/CD Integration](docs/CI_CD_INTEGRATION.md) for pipeline details**

### Code Quality
```bash
# Format code
go fmt ./...

# Lint code with extended timeout (requires golangci-lint)
golangci-lint run --timeout=10m

# Check for security issues
gosec ./...

# Install dependencies
go mod download
go mod tidy

# Run full quality checks
make lint-all
```

## Architecture Overview

This Go application follows Clean Architecture principles with clear separation of concerns:

### Core Architecture Layers
- **Domain Layer** (`internal/domain/`): Business logic, entities, and interfaces
  - `types/`: Core domain types (Transaction, TaxCalculation, ProcessingOptions)
  - `calculator/`: Tax calculation engines with jurisdiction-specific rules
  - `parser/`: CSV parsing and validation interfaces
- **Application Layer** (`internal/app/`): Orchestration and configuration
  - `cli/`: Command-line interface handlers
  - `config/`: Configuration management
  - `tui/`: Terminal UI components using Bubble Tea
- **Infrastructure Layer** (`internal/infrastructure/`): External dependencies
  - `csv/`: CSV file handling implementations
  - `logger/`: Logging implementations
  - `storage/`: Data persistence
- **Package Layer** (`internal/pkg/`): Shared utilities
  - `currency/`: Currency conversion utilities
  - `date/`: Date manipulation helpers
  - `validation/`: Input validation helpers

### Key Domain Types

**Transaction**: Represents T212 transactions with fields for action, time, ISIN, ticker, shares, prices, currency, and tax information.

**TaxCalculation**: Results of tax calculations including gains, losses, dividend income, and estimated taxes.

**ProcessingOptions**: Configuration for processing including tax year, jurisdiction, and currency preferences.

### Tax Jurisdictions

The calculator supports multiple jurisdictions with different tax rates and allowances:
- **US**: 15% capital gains and dividend tax rates
- **UK**: 10% capital gains, 7.5% dividend tax with £6,000 and £2,000 allowances respectively
- **BG**: 10% capital gains, 5% dividend tax

### CSV Format

The application processes Trading 212 CSV exports with the following structure:
```
Action,Time,ISIN,Ticker,Name,No. of shares,Price / share,Currency (Price / share),Exchange rate,Result (USD),Total (USD),Withholding tax,Notes
```

Supported transaction types: Market buy/sell, Limit buy/sell, Stop buy/sell, Dividend, Interest, Deposit, Withdrawal.

### Key Design Patterns

- **Interface-based design**: All major components (Parser, Calculator) are defined as interfaces
- **Strategy Pattern**: Different tax calculation strategies for various jurisdictions
- **Dependency Injection**: Components are injected through interfaces
- **Repository Pattern**: Abstract data access (planned for storage layer)

### Development Status

**Status**: Fully functional CLI application with comprehensive features:
- Complete CSV parsing with full Trading 212 format support
- Financial calculations including yearly and overall reports  
- Interactive TUI and command-line interfaces
- Comprehensive test coverage
- Production-ready codebase

### Usage Examples

```bash
# Interactive TUI mode
./t212-taxes

# Process CSV files and show table output
./t212-taxes process --dir ./exports

# Analyze with interactive TUI
./t212-taxes analyze --dir ./exports

# Validate CSV file structure
./t212-taxes validate --dir ./exports

# Process with JSON output
./t212-taxes process --dir ./exports --format json --output results.json
```

### Testing Strategy

- **Unit Tests**: Individual function testing
- **Integration Tests**: Component interaction testing  
- **Table-Driven Tests**: Multiple scenario testing
- **Benchmark Tests**: Performance testing for critical paths
- **Golden File Tests**: Snapshot testing for complex outputs

### Configuration

The application uses Viper for configuration management supporting:
- YAML configuration files
- Environment variables (T212_*)
- Command line flags
- Multiple output formats (JSON, table)

### Dependencies

Key external dependencies:
- **Cobra**: CLI framework
- **Viper**: Configuration management
- **Bubble Tea**: Terminal UI framework
- **Lip Gloss**: Terminal styling
- **golang.org/x/text**: Text processing utilities