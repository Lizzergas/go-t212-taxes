# Release Process

This document outlines the complete release process for T212 Taxes, from creating a release to distributing it across multiple platforms.

## Overview

The release process is designed to be:
- **Automated**: Most steps are handled by CI/CD
- **Safe**: Multiple validation checks prevent broken releases
- **Multi-platform**: Supports Windows, macOS, and Linux
- **Traceable**: Full audit trail of what was built and when

## Quick Release

For maintainers who want to cut a release quickly:

```bash
# 1. Ensure you're on main with latest changes
git checkout main && git pull origin main

# 2. Test the release script (dry run)
./scripts/release.sh v1.0.0 --dry-run

# 3. Create and push the release tag
./scripts/release.sh v1.0.0
```

## Detailed Process

### 1. Pre-Release Preparation

#### Code Quality Check
Before any release, ensure code quality is good:

```bash
# Run tests
go test ./...

# Run quality checks
./scripts/lint-check.sh

# Full CI simulation
./scripts/ci-test.sh
```

#### Version Planning
Follow [Semantic Versioning](https://semver.org/):
- **Major** (v2.0.0): Breaking changes
- **Minor** (v1.1.0): New features, backward compatible
- **Patch** (v1.0.1): Bug fixes only
- **Pre-release** (v1.0.0-beta.1): Early releases

### 2. Release Creation

#### Using the Release Script (Recommended)

The `scripts/release.sh` script automates the entire process:

```bash
# Basic release
./scripts/release.sh v1.0.0

# Pre-release
./scripts/release.sh v1.0.0-beta.1

# Dry run to see what would happen
./scripts/release.sh v1.0.0 --dry-run

# Force push (use with caution)
./scripts/release.sh v1.0.1 --force
```

The script will:
1. ✅ Validate version format
2. ✅ Check git status (clean, on main, up to date)
3. ✅ Verify tag doesn't exist
4. ✅ Run tests
5. ✅ Run quality checks
6. ✅ Create annotated git tag
7. ✅ Push tag to trigger CI/CD

#### Manual Process

If you need to create a release manually:

```bash
# 1. Create annotated tag
git tag -a v1.0.0 -m "Release v1.0.0

Features:
- New feature X
- Improved Y
- Fixed Z

View full changelog at:
https://github.com/Lizzergas/go-t212-taxes/releases/tag/v1.0.0"

# 2. Push tag
git push origin v1.0.0
```

### 3. Automated CI/CD Process

Once a tag is pushed, GitHub Actions automatically:

#### Quality Gate (5-10 minutes)
- ✅ Runs full test suite with race detection
- ✅ Runs quality checks (allows up to 15 linting issues)
- ✅ Validates code coverage

#### Build Phase (10-15 minutes)
- ✅ Builds binaries for all platforms:
  - Linux (amd64, arm64)
  - macOS (amd64, arm64) 
  - Windows (amd64)
- ✅ Creates platform-specific archives (.tar.gz, .zip)
- ✅ Generates SHA256 checksums
- ✅ Uploads release assets to GitHub

#### Docker Phase (5-10 minutes)
- ✅ Builds multi-platform Docker images
- ✅ Pushes to GitHub Container Registry (ghcr.io)
- ✅ Tags with version and latest

#### Package Distribution Phase (5-10 minutes)
- ✅ Updates Homebrew formula in repository
- ✅ Generates changelog from git commits
- ✅ Updates release with installation instructions
- ✅ Adds download links and verification info

### 4. Distribution Channels

#### Current Channels

**GitHub Releases** ✅ Automatic
- Binary downloads for all platforms
- Source code archives
- SHA256 checksums
- Installation instructions

**GitHub Container Registry** ✅ Automatic
- Multi-platform Docker images
- Semantic version tags
- Latest tag for stable releases

**Go Modules** ✅ Automatic
- Available via `go install`
- Semantic import versioning
- Module proxy cached

**Homebrew** ✅ Automatic
```bash
# Users can install via:
brew tap Lizzergas/t212-taxes https://github.com/Lizzergas/go-t212-taxes
brew install --cask t212-taxes
```
- Cask automatically updated on each release
- Handles unsigned binary quarantine removal
- Multi-platform support (Intel and Apple Silicon)

**Scoop (Windows)** 🔄 Prepared (needs setup)
```bash
# When ready, users will be able to:
scoop bucket add t212-taxes https://github.com/Lizzergas/scoop-bucket
scoop install t212-taxes
```

**Snap Store (Ubuntu)** 🔄 Prepared (needs approval)
```bash
# When ready, users will be able to:
sudo snap install t212-taxes
```

## Version Management

### Build-time Information

Every release includes build-time information:

```bash
$ t212-taxes version
T212 Taxes v1.0.0
Commit:    abc123def456
Built:     2024-01-15T10:30:00Z
Built by:  github-actions
```

This information is injected during build via ldflags:
- `main.version`: Git tag (e.g., v1.0.0)
- `main.commit`: Git commit hash
- `main.date`: Build timestamp
- `main.builtBy`: Build environment (github-actions, docker, source)

### Release Artifacts

Each release produces:

**Binary Archives:**
- `t212-taxes-linux-amd64.tar.gz` + `.sha256`
- `t212-taxes-linux-arm64.tar.gz` + `.sha256`
- `t212-taxes-darwin-amd64.tar.gz` + `.sha256`
- `t212-taxes-darwin-arm64.tar.gz` + `.sha256`
- `t212-taxes-windows-amd64.zip` + `.sha256`

**Docker Images:**
- `ghcr.io/lizzergas/go-t212-taxes:v1.0.0`
- `ghcr.io/lizzergas/go-t212-taxes:v1.0`
- `ghcr.io/lizzergas/go-t212-taxes:v1`
- `ghcr.io/lizzergas/go-t212-taxes:latest` (for stable releases)

## Troubleshooting

### Common Issues

**"Tag already exists"**
```bash
# Check existing tags
git tag | sort -V

# Delete local tag if needed
git tag -d v1.0.0

# Delete remote tag if needed (careful!)
git push origin --delete v1.0.0
```

**"Not on main branch"**
```bash
git checkout main
git pull origin main
```

**"Working directory has uncommitted changes"**
```bash
# Commit changes
git add .
git commit -m "Prepare for release"

# Or stash temporarily
git stash
```

**"Local main branch not up to date"**
```bash
git pull origin main
```

**"Tests failed"**
```bash
# Run tests locally to debug
go test -v ./...

# Fix issues and try again
```

**"Quality gate failed"**
```bash
# Check what's failing
./scripts/lint-check.sh

# Fix issues (you have a budget of 15 total issues)
# Current status: 5/15 issues
```

### CI/CD Failures

**Test Phase Failure:**
- Check GitHub Actions logs
- Run `go test ./...` locally
- Fix failing tests and push fix

**Build Phase Failure:**
- Usually indicates build environment issues
- Check if all required files exist (go.mod, main.go, etc.)
- Verify build tags and constraints

**Docker Phase Failure:**
- Check Dockerfile syntax
- Verify GitHub Container Registry permissions
- Ensure secrets.GITHUB_TOKEN has packages:write permission

### Recovery Procedures

**Failed Release Recovery:**
```bash
# 1. Delete the problematic tag
git push origin --delete v1.0.0
git tag -d v1.0.0

# 2. Fix the issues
# ... make necessary changes ...

# 3. Try release again
./scripts/release.sh v1.0.0
```

**Rollback Release:**
```bash
# Mark release as draft in GitHub UI or delete entirely
# Docker images can't be easily rolled back, but new release will override 'latest'
```

## Security

### Supply Chain Security

- All binaries are built in GitHub Actions with reproducible builds
- SHA256 checksums provided for all downloads
- Docker images built from source with security scanning
- No third-party dependencies in release pipeline

### Verification

Users can verify downloads:

```bash
# Download binary and checksum
curl -L https://github.com/Lizzergas/go-t212-taxes/releases/download/v1.0.0/t212-taxes-linux-amd64.tar.gz
curl -L https://github.com/Lizzergas/go-t212-taxes/releases/download/v1.0.0/t212-taxes-linux-amd64.tar.gz.sha256

# Verify checksum
sha256sum -c t212-taxes-linux-amd64.tar.gz.sha256
```

## Future Improvements

### Planned Enhancements

1. **GoReleaser Migration**: Switch from GitHub Actions to GoReleaser for more advanced features
2. **Homebrew Tap**: Set up official Homebrew tap for easier macOS installation
3. **Windows Package Manager**: Add support for Scoop and/or Chocolatey
4. **Linux Package Repositories**: Add .deb and .rpm packages
5. **Release Automation**: Automatic releases on main branch with sufficient changes
6. **Release Notes Enhancement**: Better changelog generation from conventional commits

### Setting Up Future Channels

**Homebrew Tap Setup:**
1. Create `homebrew-t212-taxes` repository
2. Add `HOMEBREW_TAP_GITHUB_TOKEN` secret
3. Enable GoReleaser or manual Homebrew formula updates

**Scoop Bucket Setup:**
1. Create `scoop-bucket` repository  
2. Add `SCOOP_TAP_GITHUB_TOKEN` secret
3. Configure Scoop manifest generation

**Snap Store Setup:**
1. Register developer account on Snapcraft
2. Configure snapcraft.yaml
3. Set up automated publishing

## Monitoring

### Release Health

Monitor these metrics after each release:
- Download counts from GitHub Releases
- Docker image pulls from GHCR
- `go install` statistics (if available)
- User feedback and issue reports

### Key Performance Indicators

- **Release Frequency**: Target monthly releases
- **Time to Release**: From tag to available < 30 minutes
- **Release Quality**: Zero critical bugs in first 48 hours
- **Adoption Rate**: Downloads increasing release over release

## 🎯 Release Best Practices

### Commit Discipline for Releases

To ensure clean release notes, follow these guidelines:

**✅ Good Release Practices:**
- **Squash debug commits** before releasing
- **Use descriptive commit prefixes** that will be filtered appropriately:
  - `debug:` - Debug/troubleshooting commits (filtered out)
  - `chore:` - Maintenance tasks (filtered out)  
  - `refactor:` - Code restructuring (filtered out)
  - `feat:` - New features (included in release notes)
  - `fix:` - Bug fixes (included in release notes)

**❌ Avoid These Patterns:**
- Releasing during active debugging sessions
- Including multiple unrelated commits in a single release
- Adding commits with generic messages like "fix stuff"

**🔧 Emergency Fix Protocol:**
If you need to release during debugging:
1. **Squash commits** with `git rebase -i HEAD~n`
2. **Use `[skip changelog]`** prefix for commits that shouldn't appear
3. **Consider patch releases** (v1.0.21 → v1.0.22) instead of jumping versions

**📝 Commit Message Examples:**
```bash
# These will be INCLUDED in release notes:
git commit -m "feat: add new portfolio analysis feature"
git commit -m "fix: resolve CSV parsing error for dividend records"

# These will be EXCLUDED from release notes:
git commit -m "debug: add logging for CI compilation issue"
git commit -m "chore: update linting configuration" 
git commit -m "refactor: reorganize calculator package structure"
git commit -m "[skip changelog] temporary debugging commit"
```

---

For questions about the release process, see:
- [Contributing Guide](../CONTRIBUTING.md)
- [GitHub Issues](https://github.com/Lizzergas/go-t212-taxes/issues)
- [GitHub Discussions](https://github.com/Lizzergas/go-t212-taxes/discussions) 