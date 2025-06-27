# T212 Taxes

A comprehensive Go application for processing Trading 212 CSV exports and calculating tax obligations with modern tooling and scalable architecture.

## Author

**lizz**

## Features

- 📊 **CSV Processing**: Parse and validate Trading 212 export files with robust error handling
- 🧮 **Tax Calculations**: Calculate various tax scenarios and obligations for multiple jurisdictions
- 🎨 **Modern TUI**: Beautiful terminal user interface using Bubble Tea and Lip Gloss
- 🔧 **Type Safety**: Comprehensive struct validation and error handling
- 🧪 **Testing**: Extensive test suite with table-driven tests and benchmarks
- 📦 **CLI Tool**: Feature-rich command-line interface with Cobra
- 🚀 **Scalable**: Clean architecture with dependency injection and interfaces
- ⚡ **Performance**: Optimized for handling large CSV files with streaming processing

## Installation

### From Source

```bash
git clone https://github.com/lizz/t212-taxes.git
cd t212-taxes
go build -o t212-taxes ./cmd/t212-taxes
```

### Using Go Install

```bash
go install github.com/lizz/t212-taxes/cmd/t212-taxes@latest
```

## Development

### Prerequisites

- Go 1.21 or later
- Make (optional, for convenience commands)

### Setup

```bash
# Clone the repository
git clone https://github.com/lizz/t212-taxes.git
cd t212-taxes

# Install dependencies
go mod download

# Run tests
go test ./...

# Build the application
go build -o t212-taxes ./cmd/t212-taxes
```

### Development Commands

```bash
# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out

# Run benchmarks
go test -bench=. ./...

# Format code
go fmt ./...

# Lint code (requires golangci-lint)
golangci-lint run

# Build for multiple platforms
make build-all

# Run the application in development
go run ./cmd/t212-taxes
```

## Usage

### Basic Usage

```bash
# Interactive mode (default)
./t212-taxes

# Process a specific CSV file
./t212-taxes process --file ./data/transactions.csv

# Calculate taxes for a specific year
./t212-taxes calculate --year 2024 --jurisdiction US

# Export results in different formats
./t212-taxes export --format json --output results.json
```

### Command Line Options

```bash
# Show help
./t212-taxes --help

# Version information
./t212-taxes version

# Enable verbose logging
./t212-taxes --verbose process --file data.csv

# Use different configuration file
./t212-taxes --config ./config.yaml process --file data.csv
```

## Project Structure

```
.
├── cmd/
│   └── t212-taxes/          # Main application entry point
│       └── main.go
├── internal/
│   ├── app/                 # Application layer
│   │   ├── cli/            # CLI commands and handlers
│   │   ├── config/         # Configuration management
│   │   └── tui/            # Terminal UI components
│   ├── domain/             # Business logic and entities
│   │   ├── calculator/     # Tax calculation engines
│   │   ├── parser/         # CSV parsing and validation
│   │   └── types/          # Core domain types
│   ├── infrastructure/     # External concerns
│   │   ├── csv/           # CSV file handling
│   │   ├── logger/        # Logging implementation
│   │   └── storage/       # Data persistence
│   └── pkg/               # Shared utilities
│       ├── currency/      # Currency conversion utilities
│       ├── date/          # Date manipulation helpers
│       └── validation/    # Input validation helpers
├── api/                   # API definitions (if needed)
├── configs/               # Configuration files
├── data/                  # Sample data and test files
│   ├── sample/           # Example CSV files
│   └── schemas/          # Validation schemas
├── docs/                 # Documentation
├── scripts/              # Build and deployment scripts
├── test/                 # Integration and end-to-end tests
└── web/                  # Web interface (future)
```

## Architecture

This project follows Clean Architecture principles with clear separation of concerns:

- **Domain Layer**: Core business logic, entities, and use cases
- **Application Layer**: Orchestration of domain logic and external interfaces
- **Infrastructure Layer**: External dependencies (file systems, databases, APIs)
- **Interface Layer**: CLI, TUI, and future web interfaces

### Key Design Patterns

- **Dependency Injection**: All dependencies are injected through interfaces
- **Repository Pattern**: Abstract data access with pluggable implementations
- **Command Pattern**: CLI commands with consistent structure and error handling
- **Strategy Pattern**: Different tax calculation strategies for various jurisdictions
- **Observer Pattern**: Event-driven processing for real-time updates

## CSV Format Support

The application supports various Trading 212 CSV export formats:

### Transaction History
- Market orders (buy/sell)
- Limit orders
- Stop orders
- Dividend payments
- Interest payments
- Deposits and withdrawals

### Data Validation
- ISIN code validation
- Ticker symbol validation
- Currency code validation
- Date format handling
- Numeric precision handling

## Tax Calculations

### Supported Jurisdictions
- **United States**: Federal tax calculations with state considerations
- **United Kingdom**: Capital gains and dividend tax with allowances
- **European Union**: General EU tax framework
- **Bulgaria**: Local tax rules and regulations

### Calculation Features
- Capital gains/losses with FIFO/LIFO methods
- Dividend tax calculations with withholding tax credits
- Wash sale rule applications
- Multi-currency support with exchange rate handling
- Tax year boundary handling
- Detailed audit trails

## Configuration

The application supports multiple configuration methods:

### Configuration File (config.yaml)
```yaml
app:
  log_level: "info"
  output_format: "table"

tax:
  default_jurisdiction: "US"
  default_year: 2024
  use_fifo_method: true

csv:
  delimiter: ","
  skip_invalid_rows: true
  date_format: "2006-01-02 15:04:05"
```

### Environment Variables
```bash
export T212_LOG_LEVEL=debug
export T212_TAX_JURISDICTION=UK
export T212_CSV_DELIMITER=";"
```

### Command Line Flags
All configuration options can be overridden via command line flags.

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# Run specific test suite
go test ./internal/domain/calculator/...

# Run benchmarks
go test -bench=. ./internal/domain/parser/
```

### Test Structure

- **Unit Tests**: Test individual functions and methods
- **Integration Tests**: Test component interactions
- **Table-Driven Tests**: Comprehensive test cases with multiple scenarios
- **Benchmark Tests**: Performance testing for critical paths
- **Golden File Tests**: Snapshot testing for complex outputs

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes following the coding standards
4. Add tests for new functionality
5. Run the full test suite (`go test ./...`)
6. Format your code (`go fmt ./...`)
7. Commit your changes (`git commit -m 'Add amazing feature'`)
8. Push to the branch (`git push origin feature/amazing-feature`)
9. Open a Pull Request

### Coding Standards

- Follow Go best practices and idioms
- Use meaningful variable and function names
- Write comprehensive tests for new functionality
- Document public APIs with Go doc comments
- Keep functions small and focused
- Use interfaces for abstraction
- Handle errors explicitly and appropriately

## Performance Considerations

- **Streaming Processing**: Handle large CSV files without loading everything into memory
- **Concurrent Processing**: Utilize goroutines for CPU-intensive calculations
- **Memory Optimization**: Efficient data structures and garbage collection awareness
- **Caching**: Cache exchange rates and frequently accessed data
- **Profiling**: Built-in profiling support for performance analysis

## Security

- Input validation for all user-provided data
- Secure handling of financial information
- No storage of sensitive data in logs
- Configuration file security best practices
- Dependency vulnerability scanning

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Support

- **Issues**: Report bugs and feature requests on GitHub Issues
- **Documentation**: Comprehensive documentation in the `docs/` directory
- **Examples**: Sample CSV files and usage examples in `data/sample/`

## Roadmap

- [ ] Web interface for non-technical users
- [ ] API server mode for integration with other tools
- [ ] Additional tax jurisdictions
- [ ] Real-time exchange rate integration
- [ ] Advanced reporting and visualization
- [ ] Plugin system for custom calculations
- [ ] Database storage for historical data
- [ ] Multi-user support with authentication
