# Quick Release Guide 🚀

## TL;DR - Create a Release

```bash
# 1. Ensure clean state on main
git checkout main && git pull origin main

# 2. Test (optional but recommended)
./scripts/release.sh v1.0.0 --dry-run

# 3. Create release
./scripts/release.sh v1.0.0
```

That's it! 🎉 GitHub Actions will handle the rest.

## What Happens Next?

1. **GoReleaser triggers** (~15-20 minutes total)
   - ✅ Tests and quality checks (5-10 min)
   - ✅ Build binaries for all platforms (5-8 min)
   - ✅ Build and push Docker images (3-5 min)
   - ✅ Create Homebrew cask automatically (1-2 min)
   - ✅ Create Scoop manifest automatically (1-2 min)
   - ✅ Update release with download links (1-2 min)

2. **Artifacts created:**
   - GitHub Release with binaries and checksums
   - Docker images on `ghcr.io/lizzergas/go-t212-taxes:v1.0.0`
   - Homebrew cask in `Casks/t212-taxes.rb`
   - Scoop manifest in `bucket/t212-taxes.json`
   - Go module available via `go install ...@v1.0.0`

3. **Users can install with:**
   ```bash
   # Go install
   go install github.com/Lizzergas/go-t212-taxes/cmd/t212-taxes@v1.0.0
   
   # Homebrew (when using GoReleaser)
   brew tap Lizzergas/go-t212-taxes
   brew install t212-taxes
   
   # Scoop - Windows (when using GoReleaser)
   scoop bucket add t212-taxes https://github.com/Lizzergas/go-t212-taxes
   scoop install t212-taxes
   
   # Docker
   docker run ghcr.io/lizzergas/go-t212-taxes:v1.0.0 --help
   
   # Binary download
   # Links will be in the GitHub release page
   ```

## Version Format

- **Stable**: `v1.0.0`, `v1.2.3`
- **Pre-release**: `v1.0.0-beta.1`, `v2.0.0-alpha.1`

## Troubleshooting

**"Not on main branch"**: `git checkout main`
**"Uncommitted changes"**: `git add . && git commit -m "message"`
**"Not up to date"**: `git pull origin main`
**"Tag exists"**: Choose a different version number

## Need More Details?

See [docs/RELEASE_PROCESS.md](docs/RELEASE_PROCESS.md) for the complete guide.

## Monitor Progress

- **Actions**: https://github.com/Lizzergas/go-t212-taxes/actions
- **Releases**: https://github.com/Lizzergas/go-t212-taxes/releases 