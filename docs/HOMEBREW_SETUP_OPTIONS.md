# Homebrew Tap Setup Options

This document outlines different approaches for setting up Homebrew distribution.

## Option 1: Use Main Repository (Current Setup) ✅

**How it works:**
- GoReleaser commits Homebrew formula to your main repository
- Creates `Formula/t212-taxes.rb` in your repo
- Users install with: `brew tap Lizzergas/go-t212-taxes && brew install t212-taxes`

**Pros:**
- ✅ No additional repositories needed
- ✅ Simple setup, uses existing `GITHUB_TOKEN`
- ✅ Formula is versioned with your code
- ✅ Works immediately

**Cons:**
- ⚠️ Adds Homebrew files to main repo
- ⚠️ Each project needs separate tap command

**Setup:** Already configured in `.goreleaser.yml`

## Option 2: Multi-Project Tap Repository

**How it works:**
- Create one `homebrew-tap` repository for all your projects
- GoReleaser commits formulas for all projects to this single repo
- Users install with: `brew tap Lizzergas/tap && brew install t212-taxes`

**Pros:**
- ✅ Clean separation of concerns
- ✅ One tap for all your projects
- ✅ Professional setup (like `brew tap hashicorp/tap`)
- ✅ Scalable for multiple projects

**Setup Steps:**

### 1. Create the Tap Repository
```bash
# Create new repository
gh repo create Lizzergas/homebrew-tap --public --description "Homebrew tap for Lizzergas projects"

# Clone and set up basic structure
git clone https://github.com/Lizzergas/homebrew-tap.git
cd homebrew-tap

# Create initial README
echo "# Lizzergas Homebrew Tap

This tap contains Homebrew formulas for Lizzergas projects.

## Installation

\`\`\`bash
brew tap Lizzergas/tap
brew install t212-taxes
\`\`\`

## Available Formulas

- **t212-taxes**: Trading 212 CSV processor and tax calculator

" > README.md

git add README.md
git commit -m "Initial tap setup"
git push origin main
```

### 2. Update GoReleaser Configuration
```yaml
# In .goreleaser.yml
brews:
  - name: t212-taxes
    repository:
      owner: Lizzergas
      name: homebrew-tap  # Multi-project tap
      branch: main
      token: "{{ .Env.HOMEBREW_TAP_GITHUB_TOKEN }}"
    # ... rest of config
```

### 3. Add GitHub Token
```bash
# Create personal access token with repo permissions
# Add to repository secrets as HOMEBREW_TAP_GITHUB_TOKEN
gh secret set HOMEBREW_TAP_GITHUB_TOKEN --body "your_token_here"
```

### 4. Future Projects
For additional projects, just add more `brews` entries pointing to the same `homebrew-tap` repository.

## Option 3: Official Homebrew Core

**For the future** when your project is mature:

**How it works:**
- Submit formula to official Homebrew repository
- Users install with: `brew install t212-taxes` (no tap needed)

**Requirements:**
- Notable project (1000+ GitHub stars typically)
- Stable releases for 30+ days
- No dependencies on other taps
- Open source with good documentation

**Benefits:**
- Maximum discoverability
- Official Homebrew support
- No maintenance of tap required

## Recommendation

**Start with Option 1** (current setup):
- It's already configured and working
- No additional setup required
- Perfect for getting started

**Migrate to Option 2** when you have multiple projects:
- Create `homebrew-tap` repository
- Update GoReleaser config
- Add GitHub token secret

**Consider Option 3** when project is mature and widely used.

## Comparison

| Approach | Setup Complexity | Maintenance | User Experience | Scalability |
|----------|------------------|-------------|-----------------|-------------|
| Main Repo | ⭐ Very Easy | ⭐ Low | ⭐⭐ Good | ⭐ One project |
| Multi-Tap | ⭐⭐ Easy | ⭐⭐ Medium | ⭐⭐⭐ Great | ⭐⭐⭐ Many projects |
| Official | ⭐⭐⭐ Hard | ⭐ None | ⭐⭐⭐ Perfect | ⭐⭐⭐ Unlimited |

## Migration Path

```bash
# Current (Option 1)
brew tap Lizzergas/go-t212-taxes
brew install t212-taxes

# Future (Option 2) 
brew tap Lizzergas/tap
brew install t212-taxes

# Future (Option 3)
brew install t212-taxes
```

The current setup gives you Homebrew distribution immediately, and you can easily migrate to a multi-project tap later when needed. 