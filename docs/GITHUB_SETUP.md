# GitHub Setup Guide

This document explains the comprehensive GitHub setup for the T212 Taxes project, including CI/CD pipelines, automation, and best practices.

## 📁 Repository Structure

The project is now GitHub-ready with the following structure:

```
.
├── .github/                    # GitHub-specific configurations
│   ├── workflows/             # GitHub Actions workflows
│   │   ├── ci.yml            # Main CI/CD pipeline
│   │   └── release.yml       # Release automation
│   ├── ISSUE_TEMPLATE/       # Issue templates
│   │   ├── bug_report.md     # Bug report template
│   │   └── feature_request.md # Feature request template
│   ├── pull_request_template.md # PR template
│   └── dependabot.yml        # Dependency updates
├── .golangci.yml              # Go linting configuration
├── .gitignore                 # Git ignore rules
├── .dockerignore              # Docker ignore rules
├── Dockerfile                 # Multi-stage Docker build
├── Makefile                   # Development commands
├── README.md                  # Project documentation
├── CONTRIBUTING.md            # Contribution guidelines
├── LICENSE                    # MIT license
├── CHANGELOG.md               # Version history
└── scripts/
    └── setup.sh              # Development setup script
```

## 🔧 CI/CD Pipeline

### Main CI Workflow (`.github/workflows/ci.yml`)

The CI pipeline runs on every push and pull request, featuring:

#### **Testing Phase**
- ✅ Go 1.21 compatibility testing
- ✅ Unit tests with race detection
- ✅ Code coverage reporting (Codecov integration)
- ✅ Test artifacts collection

#### **Code Quality Phase**
- ✅ Linting with golangci-lint
- ✅ Security scanning with Gosec
- ✅ SARIF reporting for security issues

#### **Build Phase**
- ✅ Multi-platform builds (Linux, macOS, Windows)
- ✅ Multiple architectures (amd64, arm64)
- ✅ Build artifacts collection

#### **Release Phase** (on tags)
- ✅ Automated GitHub releases
- ✅ Release notes generation
- ✅ Multi-platform binary distribution

#### **Docker Phase**
- ✅ Docker image building
- ✅ GitHub Container Registry publishing
- ✅ Multi-architecture support

### Release Workflow (`.github/workflows/release.yml`)

Triggered on version tags (`v*`):

1. **Create Release**: Generates changelog from git commits
2. **Build Assets**: Creates platform-specific archives
3. **Upload Assets**: Attaches binaries to GitHub release
4. **Update Notes**: Adds installation instructions

## 🔍 Code Quality

### Linting Configuration (`.golangci.yml`)

Comprehensive linting with 30+ enabled linters:

- **Code Quality**: cyclomatic complexity, duplicate detection
- **Style**: formatting, naming conventions
- **Security**: potential vulnerabilities
- **Performance**: inefficient code patterns
- **Best Practices**: Go idioms and patterns

### Pre-commit Hooks

Optional pre-commit hooks for developers:

```bash
# Set up hooks
./scripts/setup.sh
# or manually
git config core.hooksPath .githooks
```

## 🐳 Docker Support

### Multi-stage Dockerfile

- **Builder Stage**: Go 1.21 Alpine with build dependencies
- **Runtime Stage**: Minimal Alpine with security hardening
- **Features**: Non-root user, health checks, timezone support

### Container Registry

Images published to GitHub Container Registry:
- `ghcr.io/YOUR_USERNAME/t212-taxes:latest`
- `ghcr.io/YOUR_USERNAME/t212-taxes:v1.0.0`

## 🔄 Dependency Management

### Dependabot Configuration

Automated dependency updates for:
- **Go modules**: Weekly updates on Mondays
- **GitHub Actions**: Weekly updates
- **Docker base images**: Weekly updates

### Update Strategy

- Maximum 5 Go module PRs open simultaneously
- Maximum 3 GitHub Actions PRs open simultaneously
- Automatic labeling and assignment

## 📋 Issue and PR Templates

### Bug Report Template
- Environment details
- Reproduction steps
- Expected vs actual behavior
- Sample data (anonymized)

### Feature Request Template
- Problem statement
- Proposed solution
- Use cases
- Technical considerations

### Pull Request Template
- Change description
- Testing checklist
- Documentation updates
- Security considerations

## 📊 Badges and Metrics

The README includes badges for:
- ✅ CI/CD status
- ✅ Go Report Card
- ✅ Code coverage
- ✅ License
- ✅ Go version
- ✅ Latest release
- ✅ Docker availability

## 🚀 Release Process

### Automated Releases

1. **Create Tag**: `git tag v1.0.0 && git push origin v1.0.0`
2. **CI Builds**: Automatic multi-platform builds
3. **Release Creation**: GitHub release with changelog
4. **Asset Upload**: Binaries for all platforms
5. **Docker Publish**: Container images to registry

### Manual Steps

1. Update `CHANGELOG.md`
2. Update version references in README
3. Test locally with `make release-check`
4. Create and push git tag

## 🛠️ Development Commands

### Makefile Commands

```bash
# Development
make help          # Show all commands
make dev-setup     # Set up development environment
make deps          # Download dependencies
make build         # Build application
make run-dev       # Run in development mode

# Testing
make test          # Run tests
make test-coverage # Run tests with coverage
make benchmark     # Run benchmarks

# Code Quality
make fmt           # Format code
make lint          # Run linter
make security      # Security scan
make check         # Run all checks

# Building
make build-all     # Build for all platforms
make docker-build  # Build Docker image

# Release
make release-check # Prepare for release
```

### Setup Script

New contributors can use:

```bash
chmod +x scripts/setup.sh
./scripts/setup.sh
```

This script:
- Checks prerequisites (Go, Git)
- Installs development tools
- Downloads dependencies
- Runs tests
- Builds application
- Sets up pre-commit hooks (optional)

## 🔒 Security

### Security Measures

- **Dependency Scanning**: Automated vulnerability detection
- **Code Scanning**: Gosec security analysis
- **Container Security**: Non-root user, minimal base image
- **Secrets Management**: No sensitive data in repository

### Security Workflow

1. **Gosec Scan**: Every CI run
2. **SARIF Upload**: Results to GitHub Security tab
3. **Dependabot Alerts**: Vulnerability notifications
4. **Security Updates**: Automated PR creation

## 📈 Monitoring and Analytics

### Code Coverage

- **Tool**: Go's built-in coverage
- **Reporting**: Codecov integration
- **Target**: >90% coverage maintained
- **Visibility**: Badge in README

### Performance

- **Benchmarks**: Automated in CI
- **Memory Profiling**: Available via `go test -bench`
- **Build Performance**: Cached dependencies

## 📱 Platform Support

### Supported Platforms

| Platform | Architecture | Binary | Docker |
|----------|--------------|--------|--------|
| Linux | amd64 | ✅ | ✅ |
| Linux | arm64 | ✅ | ✅ |
| macOS | amd64 (Intel) | ✅ | - |
| macOS | arm64 (Apple Silicon) | ✅ | - |
| Windows | amd64 | ✅ | - |

### Installation Methods

1. **Pre-built Binaries**: GitHub Releases
2. **Go Install**: `go install github.com/...`
3. **Docker**: GitHub Container Registry
4. **Source**: Git clone and build

## 🤝 Contributing

### Contribution Workflow

1. **Fork** repository
2. **Clone** your fork
3. **Create** feature branch
4. **Develop** with tests
5. **Test** locally (`make check`)
6. **Submit** pull request

### Code Standards

- **Go Standards**: Official Go guidelines
- **Testing**: >90% coverage required
- **Documentation**: Comprehensive comments
- **Formatting**: `gofmt` compliance
- **Linting**: All golangci-lint checks pass

## 📚 Documentation

### Available Documentation

- **README.md**: Project overview and usage
- **CONTRIBUTING.md**: Contribution guidelines
- **CHANGELOG.md**: Version history
- **LICENSE**: MIT license
- **CLAUDE.md**: Development guidance

### Demo Integration

The README automatically displays `demo.gif` when uploaded to the repository root. The image reference is already configured.

## 🔮 Future Enhancements

### Planned Improvements

- [ ] **Web Interface**: Browser-based UI
- [ ] **API Mode**: REST API server
- [ ] **Additional Platforms**: More OS/arch combinations
- [ ] **Performance Metrics**: Detailed benchmarking
- [ ] **Integration Tests**: End-to-end testing

### Community Features

- [ ] **Discussion Board**: GitHub Discussions
- [ ] **Wiki**: Comprehensive documentation
- [ ] **Sponsors**: GitHub Sponsors integration
- [ ] **Contributors**: Recognition system

---

This GitHub setup provides a production-ready foundation for open-source development with comprehensive automation, quality assurance, and community support. 