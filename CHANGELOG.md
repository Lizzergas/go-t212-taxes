# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased] - 2024-12-XX

### Changed
- **Linting Infrastructure**: Major update to `.golangci.yml` configuration
  - Updated GitHub Actions to use latest versions (upload-artifact v4, codeql-action v3)
  - Replaced deprecated linters (deadcode, varcheck, structcheck, etc.) with modern equivalents
  - Added new linters: copyloopvar (Go 1.22+), mnd (magic number detection)
  - Increased timeout to 10 minutes for comprehensive checks
  - Enhanced error handling with proper `//nolint:errcheck` comments

### Fixed  
- **Code Quality**: Addressed all critical linting issues
  - Fixed magic number usage by introducing comprehensive constants
  - Eliminated code duplication in portfolio calculator with helper functions
  - Added missing package documentation comments
  - Resolved function complexity issues and line length violations
  - Fixed builtin shadowing issues (renamed `max`/`min` to `maxInt`/`minInt`)
  - Improved string constant usage throughout TUI and CLI components

### Technical Debt
- Cleaned up unused functions and dead code
- Enhanced type safety and error handling patterns
- Improved code readability and maintainability

## [Unreleased]

### Added
- Initial project setup with comprehensive CI/CD pipeline
- Docker support with multi-stage builds
- Complete test coverage with automated testing
- Security scanning with Gosec

## [1.0.0] - 2025-01-XX

### Added
- **Core Features**
  - CSV processing for Trading 212 exports with robust error handling
  - Tax calculations for multiple jurisdictions (US, UK, EU, Bulgaria)
  - Portfolio valuation with market values and P&L tracking
  - Income analysis with dividend and interest calculations

- **User Interface**
  - Beautiful terminal UI built with Bubble Tea and Lip Gloss
  - Interactive CLI with Cobra framework
  - Scrollable portfolio navigation with arrow keys
  - Browser integration for opening Yahoo Finance quotes
  - Responsive design that adapts to terminal size

- **Portfolio Features**
  - Real-time portfolio analysis with market values
  - Unrealized gains/losses calculations with percentages
  - Position ranking by market value
  - Expand/collapse functionality for position lists
  - Cursor highlighting for selected positions
  - Yearly activity summaries (deposits, dividends, interest)

- **Technical Features**
  - Clean Architecture with dependency injection
  - Comprehensive test suite with >90% coverage
  - Type-safe error handling and validation
  - Multi-currency support with exchange rate handling
  - Performance optimizations for large CSV files
  - Cross-platform support (Windows, macOS, Linux)

- **Commands**
  - `analyze` - Interactive TUI mode for data exploration
  - `process` - Command-line processing with table output
  - `portfolio` - Portfolio valuation reports
  - `income` - Dividend and interest analysis
  - `validate` - CSV file validation

### Technical Details
- **Architecture**: Clean Architecture with clear separation of concerns
- **Testing**: Unit, integration, and benchmark tests
- **Dependencies**: Minimal external dependencies with careful selection
- **Performance**: Streaming CSV processing for memory efficiency
- **Security**: Input validation and secure handling of financial data

### Browser Integration
- Open Yahoo Finance quotes directly from TUI
- Cross-platform browser launching (Windows/macOS/Linux)
- URL format: `https://finance.yahoo.com/quote/{TICKER}`

### Supported Calculations
- Capital gains/losses with FIFO/LIFO methods
- Dividend tax calculations with withholding tax credits
- Multi-currency support with proper exchange rate handling
- Tax year boundary handling
- Detailed audit trails for compliance

[Unreleased]: https://github.com/Lizzergas/go-t212-taxes/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/Lizzergas/go-t212-taxes/releases/tag/v1.0.0 