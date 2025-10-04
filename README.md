# gwt - Git Worktree CLI

A CLI tool that makes git worktree management intuitive through conventions and interactive selection.

## What is gwt?

`gwt` simplifies working with multiple git branches simultaneously by:
- **Auto-organizing worktrees** in sibling directories (`myproject-feature-name`)
- **Interactive selection** with fzf for navigation and cleanup
- **Single binary** - no runtime dependencies

`gwt` is a complete wrapper around git worktree, meaning all git worktree commands work through gwt:
- `gwt list` → `git worktree list`
- `gwt add <path> <ref>` → `git worktree add <path> <ref>`
- Any unknown command is passed through to git worktree

### Quick Example

```bash
# Traditional git worktree
cd /work/myproject
git worktree add ../myproject-feature-login -b feature/login
cd ../myproject-feature-login

# With gwt
gwt new feature/login  # Creates and navigates automatically
```

## Requirements

- **Git** 2.5 or later (for worktree support)
- **macOS** or **Linux** (Windows not yet supported)
- **Go** 1.20+ (only for building from source)

## Installation

### Step 1: Install Binary

```bash
# Option A: Install via Go (recommended)
go install github.com/toritsuyo/gwt@latest

# Option B: Build from source
git clone https://github.com/toritsuyo/gwt.git
cd gwt
go build -o gwt ./cmd/gwt
sudo mv gwt /usr/local/bin/
```

Ensure `$GOPATH/bin` is in your PATH:
```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

### Step 2: Enable Shell Integration

**Why needed?** The gwt binary cannot change your current directory directly (child processes can't modify parent shell state). Shell integration provides a wrapper function that enables `gwt go` to actually navigate.

Choose your shell:

**Bash:**
```bash
echo 'eval "$(gwt hook bash)"' >> ~/.bashrc
source ~/.bashrc
```

**Zsh:**
```bash
echo 'eval "$(gwt hook zsh)"' >> ~/.zshrc
source ~/.zshrc
```

**Fish:**
```bash
echo 'gwt hook fish | source' >> ~/.config/fish/config.fish
exec fish
```

### Step 3: Verify Installation

```bash
# Check binary
gwt --help
gwt --version

# Verify shell function is loaded
type gwt  # Should show "gwt is a function"
```

## Core Commands

### Create Worktree
```bash
gwt new feature/new-ui              # Creates ../myproject-feature-new-ui
gwt new feature/fix --cd            # Create and navigate immediately
gwt new bugfix/123 main             # Create from specific branch/commit
```

The `--cd` flag outputs only the path (for shell function navigation) instead of user-friendly messages.

### Navigate Between Worktrees
```bash
gwt go                # Interactive selection (uses fzf if available)
gwt go feature        # Filter by keyword (partial match), auto-select if only one match
gwt go --no-fzf       # Use numbered selection instead of fzf
```

**How filtering works:** Searches for substring matches (case-insensitive). If multiple matches found, shows selection UI. If only one match, navigates immediately.

**Note:** Without shell integration, this only displays the path without navigating.

### Remove Worktree
```bash
gwt clean                      # Interactive removal
gwt clean feature              # Filter and select
gwt clean --force              # Force remove even with uncommitted changes (WARNING: may lose work)
gwt clean --keep-branch        # Remove worktree but keep the branch
gwt clean --yes                # Skip all confirmations
```

### Review GitHub PRs
```bash
gwt pr 123                          # Checkout PR #123 for review
gwt pr 123 --branch review/pr-123   # Specify custom local branch name
gwt pr 123 --cd                     # Navigate immediately after creation
```

**Prerequisites:** Requires GitHub CLI (`gh`) and authentication:
```bash
brew install gh  # or apt install gh
gh auth login
```

### Open in Editor
```bash
gwt open              # Select worktree and open with default editor
gwt open feature      # Filter and open
gwt open --editor code main   # Open with specific editor
```

Editor priority: `--editor` flag → `GWT_EDITOR` → `VISUAL` → `EDITOR` → auto-detect (code, idea, subl, vim, vi).

### Shell Integration Setup
```bash
gwt hook bash    # Output bash shell function
gwt hook zsh     # Output zsh shell function
gwt hook fish    # Output fish shell function
```

See Installation section for setup instructions.

### Passthrough Commands
All unknown commands are passed through to `git worktree`:
```bash
gwt list           # → git worktree list
gwt lock <path>    # → git worktree lock <path>
gwt prune          # → git worktree prune
```

### Global Flags

Available for all commands:
- `--debug` - Show git command execution
- `--quiet` - Minimal output
- `--repo <path>` - Manually specify repository root
- `-h, --help` - Show help for any command

## Optional Dependencies

**fzf (recommended):**
Installing fzf enables interactive selection UI for `gwt go`/`gwt clean`/`gwt open`. Without it, falls back to numbered selection.

```bash
# macOS
brew install fzf

# Ubuntu/Debian
sudo apt install fzf

# Other Linux
git clone --depth 1 https://github.com/junegunn/fzf.git ~/.fzf
~/.fzf/install
```

**GitHub CLI (for `gwt pr`):**
Only required if using the PR review feature.

```bash
# macOS
brew install gh

# Ubuntu/Debian
curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | \
  sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | \
  sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null
sudo apt update
sudo apt install gh

# Authenticate
gh auth login
```

## Troubleshooting

### Can't navigate with `gwt go`

**Symptom:** `gwt go` prints a path but doesn't change directory.

**Check:**
```bash
type gwt  # Should show "gwt is a function"
```

**If it shows:** `gwt is /path/to/gwt`

**Solution:** Shell function is not loaded. Re-run Step 2 of Installation.

### Command not found

**Symptom:** `gwt: command not found`

**Solution:** Ensure `$GOPATH/bin` is in your PATH:
```bash
# Check if GOPATH/bin is in PATH
echo $PATH | grep $(go env GOPATH)/bin

# If not, add to your shell config
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.zshrc
source ~/.zshrc
```

### Branch already in use

**Symptom:** Error when creating worktree: "branch 'feature/xxx' is already checked out"

**Solution:** A worktree can only check out each branch once. Use a different branch name or navigate to the existing worktree:
```bash
gwt go feature/xxx  # Navigate to existing worktree
```

### GitHub authentication error (gwt pr)

**Symptom:** `gh: command not found` or authentication errors

**Solution:**
```bash
# Install GitHub CLI (see Optional Dependencies section)
brew install gh

# Authenticate
gh auth login
```

### fzf not found

**Note:** gwt works without fzf using numbered selection UI. Installing fzf improves the experience (see Optional Dependencies).

## Project Structure

### Shell Integration (`shell/` directory)

The `shell/` directory contains standalone shell function files that can be sourced directly or used as reference:

- **`gwt.sh`** - Bash/Zsh shell function (standalone version)
- **`gwt.fish`** - Fish shell function (standalone version)

**Note:** These files are for reference or manual installation. The recommended method is using the `gwt hook` command:

```bash
# Recommended (auto-updates with binary)
eval "$(gwt hook zsh)"

# Alternative (static, requires manual updates)
source ~/.gwt/gwt.sh
```

The embedded versions in `internal/cli/hook_*.sh` and `internal/cli/hook_*.fish` are the canonical sources and are embedded into the binary via `//go:embed` directives.

## Documentation

- [Detailed Installation Guide](INSTALL.md)

## License

MIT
