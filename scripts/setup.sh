#!/bin/bash

# T212 Taxes - Development Setup Script
# This script sets up the development environment for new contributors

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 T212 Taxes Development Setup${NC}"
echo "================================="

# Check if Go is installed
check_go() {
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version | cut -d' ' -f3 | cut -d'o' -f2)
        echo -e "${GREEN}✓ Go is installed: $GO_VERSION${NC}"
        
        # Check if Go version is 1.21 or later
        if [[ $(printf '%s\n' "1.21" "$GO_VERSION" | sort -V | head -n1) == "1.21" ]]; then
            echo -e "${GREEN}✓ Go version is compatible${NC}"
        else
            echo -e "${YELLOW}⚠ Warning: Go 1.21+ recommended, you have $GO_VERSION${NC}"
        fi
    else
        echo -e "${RED}✗ Go is not installed${NC}"
        echo -e "${YELLOW}Please install Go 1.21+ from https://golang.org/dl/${NC}"
        exit 1
    fi
}

# Check if Git is installed
check_git() {
    if command -v git &> /dev/null; then
        echo -e "${GREEN}✓ Git is installed${NC}"
    else
        echo -e "${RED}✗ Git is not installed${NC}"
        echo -e "${YELLOW}Please install Git from https://git-scm.com/downloads${NC}"
        exit 1
    fi
}

# Install golangci-lint
install_golangci_lint() {
    if command -v golangci-lint &> /dev/null; then
        echo -e "${GREEN}✓ golangci-lint is already installed${NC}"
    else
        echo -e "${YELLOW}📦 Installing golangci-lint...${NC}"
        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2
        echo -e "${GREEN}✓ golangci-lint installed${NC}"
    fi
}

# Install gosec
install_gosec() {
    if command -v gosec &> /dev/null; then
        echo -e "${GREEN}✓ gosec is already installed${NC}"
    else
        echo -e "${YELLOW}📦 Installing gosec...${NC}"
        go install github.com/securego/gosec/v2/cmd/gosec@latest
        echo -e "${GREEN}✓ gosec installed${NC}"
    fi
}

# Download dependencies
download_deps() {
    echo -e "${YELLOW}📦 Downloading Go dependencies...${NC}"
    go mod download
    go mod verify
    echo -e "${GREEN}✓ Dependencies downloaded${NC}"
}

# Run tests to verify setup
run_tests() {
    echo -e "${YELLOW}🧪 Running tests to verify setup...${NC}"
    if go test ./...; then
        echo -e "${GREEN}✓ All tests pass${NC}"
    else
        echo -e "${RED}✗ Some tests failed${NC}"
        echo -e "${YELLOW}This might be normal if you haven't set up sample data yet${NC}"
    fi
}

# Build the application
build_app() {
    echo -e "${YELLOW}🔨 Building application...${NC}"
    if go build -o t212-taxes ./cmd/t212-taxes; then
        echo -e "${GREEN}✓ Build successful${NC}"
        echo -e "${BLUE}You can now run: ./t212-taxes --help${NC}"
    else
        echo -e "${RED}✗ Build failed${NC}"
        exit 1
    fi
}

# Set up pre-commit hooks (optional)
setup_hooks() {
    read -p "Do you want to set up pre-commit hooks? [y/N]: " setup_hooks
    if [[ $setup_hooks =~ ^[Yy]$ ]]; then
        mkdir -p .git/hooks
        cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash
echo "Running pre-commit checks..."

# Format code
go fmt ./...

# Run linter
if command -v golangci-lint &> /dev/null; then
    golangci-lint run
else
    echo "Warning: golangci-lint not found, skipping linting"
fi

# Run tests
go test ./...

echo "Pre-commit checks complete!"
EOF
        chmod +x .git/hooks/pre-commit
        echo -e "${GREEN}✓ Pre-commit hooks set up${NC}"
    fi
}

# Main setup process
main() {
    echo ""
    echo -e "${BLUE}Checking prerequisites...${NC}"
    check_go
    check_git
    
    echo ""
    echo -e "${BLUE}Installing development tools...${NC}"
    install_golangci_lint
    install_gosec
    
    echo ""
    echo -e "${BLUE}Setting up project...${NC}"
    download_deps
    run_tests
    build_app
    
    echo ""
    echo -e "${BLUE}Optional setup...${NC}"
    setup_hooks
    
    echo ""
    echo -e "${GREEN}🎉 Development environment setup complete!${NC}"
    echo ""
    echo -e "${BLUE}Next steps:${NC}"
    echo "1. Run: ./t212-taxes --help"
    echo "2. Read: CONTRIBUTING.md"
    echo "3. Check: make help"
    echo ""
    echo -e "${YELLOW}Happy coding! 🚀${NC}"
}

# Run main function
main "$@" 