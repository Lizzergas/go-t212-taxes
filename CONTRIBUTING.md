# Contributing to T212 Taxes

Thank you for your interest in contributing to T212 Taxes! We welcome contributions from everyone.

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Code Style](#code-style)
- [Commit Messages](#commit-messages)
- [Issue Guidelines](#issue-guidelines)

## 📜 Code of Conduct

This project adheres to a code of conduct. By participating, you are expected to uphold this code:

- **Be respectful**: Use welcoming and inclusive language
- **Be collaborative**: Disagreements happen, discuss them constructively
- **Be patient**: Remember that people have varying communication styles and technical experience
- **Be thoughtful**: Consider how your contribution affects others

## 🚀 Getting Started

### Prerequisites

- **Go 1.21+**: [Download and install Go](https://golang.org/dl/)
- **Git**: [Download and install Git](https://git-scm.com/downloads)
- **golangci-lint**: [Installation guide](https://golangci-lint.run/usage/install/)

### First Time Setup

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/Lizzergas/go-t212-taxes.git
   cd t212-taxes
   ```
3. **Add the upstream remote**:
   ```bash
   git remote add upstream https://github.com/ORIGINAL_OWNER/go-t212-taxes.git
   ```
4. **Install dependencies**:
   ```bash
   go mod download
   ```
5. **Verify your setup**:
   ```bash
   go test ./...
   go build ./cmd/t212-taxes
   ```

## 🛠️ Development Setup

### Environment Setup

```bash
# Set up your development environment
export GO111MODULE=on
export CGO_ENABLED=0

# Optional: Set up pre-commit hooks
git config core.hooksPath .githooks
```

### Development Commands

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Format code
go fmt ./...

# Lint code
golangci-lint run

# Security scan
gosec ./...

# Build application
go build -o t212-taxes ./cmd/t212-taxes

# Run with sample data
./t212-taxes analyze --dir ./data/sample
```

### Project Structure

Understanding the project structure will help you navigate and contribute effectively:

```
├── cmd/t212-taxes/          # Application entry point
├── internal/
│   ├── app/                 # Application layer
│   │   ├── cli/            # CLI commands
│   │   ├── config/         # Configuration
│   │   └── tui/            # Terminal UI
│   ├── domain/             # Business logic
│   │   ├── calculator/     # Core calculations
│   │   ├── parser/         # CSV parsing
│   │   └── types/          # Domain types
│   ├── infrastructure/     # External dependencies
│   └── pkg/                # Shared utilities
├── data/sample/            # Sample CSV files
├── docs/                   # Documentation
└── test/                   # Integration tests
```

## 🔄 Making Changes

### Branching Strategy

1. **Sync your fork** with upstream:
   ```bash
   git checkout main
   git fetch upstream
   git merge upstream/main
   ```

2. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Make your changes** following our coding standards

4. **Test your changes**:
   ```bash
   go test ./...
   golangci-lint run
   ```

### Types of Contributions

#### 🐛 Bug Fixes
- Include a clear description of the bug
- Add regression tests
- Reference the GitHub issue number

#### ✨ New Features
- Discuss the feature in an issue first
- Follow the existing architecture patterns
- Include comprehensive tests
- Update documentation

#### 📚 Documentation
- Fix typos, clarify explanations
- Add examples and use cases
- Update API documentation

#### 🧪 Tests
- Increase test coverage
- Add integration tests
- Performance benchmarks

## 🧪 Testing

### Test Categories

1. **Unit Tests**: Test individual functions
   ```bash
   go test ./internal/domain/calculator/
   ```

2. **Integration Tests**: Test component interactions
   ```bash
   go test ./internal/app/cli/
   ```

3. **End-to-End Tests**: Test complete workflows
   ```bash
   go test ./test/
   ```

### Writing Tests

- Use table-driven tests for multiple scenarios
- Follow the naming convention: `TestFunction_Scenario`
- Include both positive and negative test cases
- Mock external dependencies
- Aim for >90% test coverage

### Example Test Structure

```go
func TestCalculator_CalculateGains(t *testing.T) {
    tests := []struct {
        name     string
        input    []Transaction
        want     float64
        wantErr  bool
    }{
        {
            name: "positive gains",
            input: []Transaction{
                {Action: "buy", Shares: 10, Price: 100},
                {Action: "sell", Shares: 10, Price: 110},
            },
            want:    100.0,
            wantErr: false,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            calc := NewCalculator()
            got, err := calc.CalculateGains(tt.input)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("CalculateGains() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if got != tt.want {
                t.Errorf("CalculateGains() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## 📝 Submitting Changes

### Pull Request Process

1. **Push your branch** to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

2. **Create a Pull Request** on GitHub with:
   - Clear title describing the change
   - Detailed description of what and why
   - Reference to related issues
   - Screenshots for UI changes

3. **Ensure CI passes**:
   - All tests pass
   - Code coverage maintained
   - Linting passes
   - Security scans pass

4. **Address review feedback** promptly

5. **Squash commits** if requested

### Pull Request Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Documentation update
- [ ] Performance improvement
- [ ] Code refactoring

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing completed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Tests added for new functionality
- [ ] All tests pass
```

## 🎨 Code Style

### Go Standards
- Follow official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` for formatting
- Use `golangci-lint` for linting
- Follow Go naming conventions

### Specific Guidelines

#### Variable Naming
```go
// Good
userCount := 10
maxRetries := 5

// Avoid
n := 10
x := 5
```

#### Function Documentation
```go
// CalculateCapitalGains calculates total capital gains from transactions.
// It returns the gain amount and any error encountered during calculation.
func CalculateCapitalGains(transactions []Transaction) (float64, error) {
    // Implementation
}
```

#### Error Handling
```go
// Good
result, err := someFunction()
if err != nil {
    return fmt.Errorf("failed to process data: %w", err)
}

// Avoid
result, _ := someFunction() // Never ignore errors
```

#### Interface Design
```go
// Good - small, focused interfaces
type Calculator interface {
    Calculate([]Transaction) (*Report, error)
}

// Avoid - large interfaces
type MegaInterface interface {
    Calculate()
    Parse()
    Validate()
    Format()
    // ... 10 more methods
}
```

## 📝 Commit Messages

### Format
```
type(scope): subject

body

footer
```

### Types
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

### Examples
```bash
feat(calculator): add support for FIFO tax calculation

Implement First-In-First-Out calculation method for capital gains.
This provides more accurate tax calculations for US users.

Closes #123
```

```bash
fix(parser): handle malformed CSV dates gracefully

Previously, malformed dates would cause the parser to crash.
Now it logs a warning and skips the invalid row.

Fixes #456
```

## 🐛 Issue Guidelines

### Bug Reports

Include:
- **Description**: Clear description of the bug
- **Steps to reproduce**: Detailed steps
- **Expected behavior**: What should happen
- **Actual behavior**: What actually happens
- **Environment**: OS, Go version, etc.
- **Sample data**: CSV file (anonymized)

### Feature Requests

Include:
- **Use case**: Why is this needed?
- **Proposed solution**: How should it work?
- **Alternatives**: Other approaches considered
- **Implementation details**: Technical considerations

### Questions

- Check existing issues first
- Use GitHub Discussions for general questions
- Provide context and examples

## 🏆 Recognition

Contributors are recognized in:
- Release notes
- Contributors section
- Special mentions for significant contributions

## 📞 Getting Help

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: Questions and general discussion
- **Code Reviews**: Learning opportunity and mentorship

Thank you for contributing to T212 Taxes! 🙏 