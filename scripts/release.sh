#!/bin/bash

# Release Management Script for T212 Taxes
# This script helps create and push version tags to trigger releases

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Help function
show_help() {
    echo "T212 Taxes Release Management Script"
    echo ""
    echo "Usage: $0 <version> [options]"
    echo ""
    echo "Arguments:"
    echo "  version         Version to release (e.g., v1.0.0, v1.2.3-beta.1)"
    echo ""
    echo "Options:"
    echo "  -h, --help      Show this help message"
    echo "  -d, --dry-run   Show what would be done without actually doing it"
    echo "  -f, --force     Force push the tag (use with caution)"
    echo ""
    echo "Examples:"
    echo "  $0 v1.0.0                    # Create and push v1.0.0 release"
    echo "  $0 v1.2.0-beta.1            # Create pre-release"
    echo "  $0 v1.1.0 --dry-run         # Preview what would happen"
    echo ""
    echo "This script will:"
    echo "  1. Validate the version format"
    echo "  2. Check that you're on main branch with clean working directory"
    echo "  3. Run tests and quality checks"
    echo "  4. Create and push a git tag"
    echo "  5. Trigger GitHub Actions to build and publish the release"
}

# Validation functions
validate_version() {
    local version=$1
    if [[ ! $version =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.-]+)?$ ]]; then
        echo -e "${RED}❌ Error: Invalid version format '$version'${NC}"
        echo -e "${YELLOW}💡 Expected format: vX.Y.Z or vX.Y.Z-suffix (e.g., v1.0.0, v1.2.3-beta.1)${NC}"
        exit 1
    fi
}

check_git_status() {
    # Check if we're in a git repository
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        echo -e "${RED}❌ Error: Not in a git repository${NC}"
        exit 1
    fi

    # Check if we're on main branch
    local current_branch=$(git branch --show-current)
    if [ "$current_branch" != "main" ]; then
        echo -e "${RED}❌ Error: Not on main branch (currently on: $current_branch)${NC}"
        echo -e "${YELLOW}💡 Switch to main branch: git checkout main${NC}"
        exit 1
    fi

    # Check for uncommitted changes
    if ! git diff-index --quiet HEAD --; then
        echo -e "${RED}❌ Error: Working directory has uncommitted changes${NC}"
        echo -e "${YELLOW}💡 Commit or stash your changes first${NC}"
        git status --porcelain
        exit 1
    fi

    # Check if we're up to date with remote
    git fetch origin main
    local local_commit=$(git rev-parse main)
    local remote_commit=$(git rev-parse origin/main)
    
    if [ "$local_commit" != "$remote_commit" ]; then
        echo -e "${RED}❌ Error: Local main branch is not up to date with origin/main${NC}"
        echo -e "${YELLOW}💡 Pull latest changes: git pull origin main${NC}"
        exit 1
    fi
}

check_tag_exists() {
    local version=$1
    if git tag | grep -q "^$version$"; then
        echo -e "${RED}❌ Error: Tag '$version' already exists${NC}"
        echo -e "${YELLOW}💡 Existing tags:${NC}"
        git tag | sort -V | tail -5
        exit 1
    fi

    # Check if tag exists on remote
    if git ls-remote --tags origin | grep -q "refs/tags/$version$"; then
        echo -e "${RED}❌ Error: Tag '$version' already exists on remote${NC}"
        exit 1
    fi
}

run_tests() {
    echo -e "${BLUE}🧪 Running tests...${NC}"
    if ! go test ./...; then
        echo -e "${RED}❌ Error: Tests failed${NC}"
        exit 1
    fi
    echo -e "${GREEN}✅ Tests passed${NC}"
}

run_lint() {
    echo -e "${BLUE}🔍 Running linter...${NC}"
    
    # Check if golangci-lint is available
    if ! command -v golangci-lint &> /dev/null; then
        echo -e "${YELLOW}⚠️  golangci-lint not found, skipping linting${NC}"
        echo -e "${YELLOW}💡 Install with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$(go env GOPATH)/bin${NC}"
        return
    fi

    # Run our quality gate check
    if ! ./scripts/lint-check.sh; then
        echo -e "${RED}❌ Error: Quality gate failed${NC}"
        exit 1
    fi
    echo -e "${GREEN}✅ Quality checks passed${NC}"
}

create_and_push_tag() {
    local version=$1
    local dry_run=$2
    local force=$3

    echo -e "${BLUE}🏷️  Creating tag '$version'...${NC}"
    
    if [ "$dry_run" = true ]; then
        echo -e "${YELLOW}[DRY RUN] Would create tag: $version${NC}"
        echo -e "${YELLOW}[DRY RUN] Would push tag to origin${NC}"
        return
    fi

    # Create annotated tag
    git tag -a "$version" -m "Release $version

This release was created automatically by the release script.
View the full changelog and download binaries at:
https://github.com/Lizzergas/go-t212-taxes/releases/tag/$version"

    # Push the tag
    local push_args=""
    if [ "$force" = true ]; then
        push_args="--force"
        echo -e "${YELLOW}⚠️  Force pushing tag...${NC}"
    fi

    git push $push_args origin "$version"
    
    echo -e "${GREEN}✅ Tag '$version' created and pushed successfully!${NC}"
}

show_next_steps() {
    local version=$1
    
    echo ""
    echo -e "${GREEN}🎉 Release '$version' has been initiated!${NC}"
    echo ""
    echo -e "${BLUE}📋 What happens next:${NC}"
    echo "  1. GoReleaser will run tests and quality checks"
    echo "  2. Binaries will be built for all platforms (Linux, macOS, Windows)"
    echo "  3. Docker images will be built and pushed to GitHub Container Registry"
    echo "  4. GitHub release will be created with download links"
    echo "  5. Homebrew cask will be created/updated automatically"
    echo "  6. Scoop manifest will be created/updated automatically"
    echo "  7. Release notes will be generated from commit history"
    echo ""
    echo -e "${BLUE}🔗 Monitor progress:${NC}"
    echo "  • Actions: https://github.com/Lizzergas/go-t212-taxes/actions"
    echo "  • Releases: https://github.com/Lizzergas/go-t212-taxes/releases"
    echo ""
    echo -e "${BLUE}📦 When ready, users can install with:${NC}"
    echo "  • Homebrew: brew tap Lizzergas/go-t212-taxes && brew install t212-taxes"
    echo "  • Scoop: scoop bucket add t212-taxes https://github.com/Lizzergas/go-t212-taxes && scoop install t212-taxes"
    echo "  • Go: go install github.com/Lizzergas/go-t212-taxes/cmd/t212-taxes@$version"
    echo "  • Docker: docker run ghcr.io/lizzergas/go-t212-taxes:$version"
    echo "  • Binary: Download from releases page"
}

# Main script
main() {
    local version=""
    local dry_run=false
    local force=false

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -d|--dry-run)
                dry_run=true
                shift
                ;;
            -f|--force)
                force=true
                shift
                ;;
            -*)
                echo -e "${RED}❌ Error: Unknown option $1${NC}"
                show_help
                exit 1
                ;;
            *)
                if [ -z "$version" ]; then
                    version=$1
                else
                    echo -e "${RED}❌ Error: Too many arguments${NC}"
                    show_help
                    exit 1
                fi
                shift
                ;;
        esac
    done

    # Check if version is provided
    if [ -z "$version" ]; then
        echo -e "${RED}❌ Error: Version argument is required${NC}"
        show_help
        exit 1
    fi

    echo -e "${BLUE}🚀 T212 Taxes Release Script${NC}"
    echo -e "${BLUE}📦 Preparing release: $version${NC}"
    
    if [ "$dry_run" = true ]; then
        echo -e "${YELLOW}🔍 DRY RUN MODE - No changes will be made${NC}"
    fi
    
    echo ""

    # Run all checks
    validate_version "$version"
    echo -e "${GREEN}✅ Version format is valid${NC}"

    check_git_status
    echo -e "${GREEN}✅ Git status is clean${NC}"

    check_tag_exists "$version"
    echo -e "${GREEN}✅ Tag doesn't exist yet${NC}"

    # Skip tests and linting in dry-run for speed
    if [ "$dry_run" = false ]; then
        run_tests
        run_lint
    else
        echo -e "${YELLOW}[DRY RUN] Skipping tests and linting${NC}"
    fi

    # Create and push tag
    create_and_push_tag "$version" "$dry_run" "$force"

    if [ "$dry_run" = false ]; then
        show_next_steps "$version"
    fi
}

# Run main function with all arguments
main "$@" 