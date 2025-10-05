# wt - Git Worktree CLI

A CLI tool that makes git worktree management intuitive through conventions and interactive selection.

## What is wt?

`wt` simplifies working with multiple git branches simultaneously by:
- **Auto-organizing worktrees** in subdirectories (`myproject-wt/feature-name`)
- **Interactive selection** with fzf for navigation and cleanup
- **Configurable directory structure** - subdirectory or sibling mode
- **Single binary** - no runtime dependencies

`wt` is a complete wrapper around git worktree, meaning all git worktree commands work through wt:
- `wt list` → `git worktree list`
- `wt add <path> <ref>` → `git worktree add <path> <ref>`
- Any unknown command is passed through to git worktree

### Quick Example

```bash
# Traditional git worktree
cd /work/myproject
git worktree add ../myproject-feature-login -b feature/login
cd ../myproject-feature-login

# With wt (subdirectory mode)
wt new feature/login  # Creates myproject-wt/feature-login and navigates automatically
```

## Requirements

- **Git** 2.5 or later (for worktree support)
- **macOS** or **Linux** (Windows not yet supported)
- **Go** 1.20+ (only for building from source)

## Installation

### Step 1: Install Binary

```bash
# Option A: Install via Go (recommended)
go install github.com/toritsuyo/wt@latest

# Option B: Build from source
git clone https://github.com/toritsuyo/wt.git
cd wt
go build -o wt ./cmd/wt
sudo mv wt /usr/local/bin/
```

Ensure `$GOPATH/bin` is in your PATH:
```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

### Step 2: Enable Shell Integration

**Why needed?** The wt binary cannot change your current directory directly (child processes can't modify parent shell state). Shell integration provides a wrapper function that enables `wt go` to actually navigate.

Choose your shell:

**Bash:**
```bash
echo 'eval "$(wt hook bash)"' >> ~/.bashrc
source ~/.bashrc
```

**Zsh:**
```bash
echo 'eval "$(wt hook zsh)"' >> ~/.zshrc
source ~/.zshrc
```

**Fish:**
```bash
echo 'wt hook fish | source' >> ~/.config/fish/config.fish
exec fish
```

### Step 3: Verify Installation

```bash
# Check binary
wt --help
wt --version

# Verify shell function is loaded
type wt  # Should show "wt is a function"
```

## Core Commands

### Create Worktree
```bash
wt new feature/new-ui              # Creates ../myproject-wt/feature-new-ui (subdirectory mode)
wt new feature/fix --cd            # Create and navigate immediately
wt new bugfix/123 main             # Create from specific branch/commit
```

By default, worktrees are organized in subdirectories (`<repo>-wt/<branch>`). You can customize this behavior using `wt config` (see Configuration section).

The `--cd` flag outputs only the path (for shell function navigation) instead of user-friendly messages.

### Navigate Between Worktrees
```bash
wt go                # Interactive selection (uses fzf if available)
wt go feature        # Filter by keyword (partial match), auto-select if only one match
```

**Selection UI:**
- **fzf installed**: Uses fzf for fuzzy-finding with real-time filtering
- **fzf not installed**: Automatically falls back to numbered selection menu

**How filtering works:** Searches for substring matches (case-insensitive). If multiple matches found, shows selection UI. If only one match, navigates immediately.

**Note:** Without shell integration, this only displays the path without navigating.

### Remove Worktree
```bash
wt clean                      # Interactive removal
wt clean feature              # Filter and select
wt clean --force              # Force remove even with uncommitted changes (WARNING: may lose work)
wt clean --keep-branch        # Remove worktree but keep the branch
wt clean --yes                # Skip all confirmations
```

### Review GitHub PRs
```bash
wt pr 123                          # Checkout PR #123 for review
wt pr 123 --branch review/pr-123   # Specify custom local branch name
wt pr 123 --cd                     # Navigate immediately after creation
```

**Prerequisites:** Requires GitHub CLI (`gh`) and authentication:
```bash
brew install gh  # or apt install gh
gh auth login
```

### Open in Editor
```bash
wt open              # Select worktree and open with default editor
wt open feature      # Filter and open
wt open --editor code main   # Open with specific editor
```

Editor priority: `--editor` flag → `WT_EDITOR` → `VISUAL` → `EDITOR` → auto-detect (code, idea, subl, vim, vi).

### Configuration

```bash
wt config list    # Show all settings
wt config get worktree.directory_format
wt config set worktree.directory_format sibling
wt config reset   # Reset to defaults
```

**Configuration file:** `~/.config/wt/config.yaml`

**Directory modes:**
- `subdirectory` (default): Organizes worktrees in `<repo>-wt/<branch>` structure
- `sibling`: Places worktrees as `<repo>-<branch>` (legacy mode)

For detailed configuration options, directory structure examples, and best practices, see [CONFIGURATION.md](CONFIGURATION.md).

### Shell Integration Setup
```bash
wt hook bash    # Output bash shell function
wt hook zsh     # Output zsh shell function
wt hook fish    # Output fish shell function
```

See Installation section for setup instructions.

### Passthrough Commands
All unknown commands are passed through to `git worktree`:
```bash
wt list           # → git worktree list
wt lock <path>    # → git worktree lock <path>
wt prune          # → git worktree prune
```

### Global Flags

Available for all commands:
- `--debug` - Show git command execution
- `--quiet` - Minimal output
- `--repo <path>` - Manually specify repository root
- `-h, --help` - Show help for any command

## Optional Dependencies

**fzf (recommended):**
Installing fzf enables interactive selection UI for `wt go`/`wt clean`/`wt open`. Without it, falls back to numbered selection.

```bash
# macOS
brew install fzf

# Ubuntu/Debian
sudo apt install fzf

# Other Linux
git clone --depth 1 https://github.com/junegunn/fzf.git ~/.fzf
~/.fzf/install
```

**GitHub CLI (for `wt pr`):**
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

### Can't navigate with `wt go`

**Symptom:** `wt go` prints a path but doesn't change directory.

**Check:**
```bash
type wt  # Should show "wt is a function"
```

**If it shows:** `wt is /path/to/wt`

**Solution:** Shell function is not loaded. Re-run Step 2 of Installation.

### Command not found

**Symptom:** `wt: command not found`

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
wt go feature/xxx  # Navigate to existing worktree
```

### GitHub authentication error (wt pr)

**Symptom:** `gh: command not found` or authentication errors

**Solution:**
```bash
# Install GitHub CLI (see Optional Dependencies section)
brew install gh

# Authenticate
gh auth login
```

### fzf not found

**Note:** wt works without fzf using numbered selection UI. Installing fzf improves the experience (see Optional Dependencies).

## Project Structure

### Shell Integration (`shell/` directory)

The `shell/` directory contains standalone shell function files that can be sourced directly or used as reference:

- **`wt.sh`** - Bash/Zsh shell function (standalone version)
- **`wt.fish`** - Fish shell function (standalone version)

**Note:** These files are for reference or manual installation. The recommended method is using the `wt hook` command:

```bash
# Recommended (auto-updates with binary)
eval "$(wt hook zsh)"

# Alternative (static, requires manual updates)
source ~/.wt/wt.sh
```

The embedded versions in `internal/cli/hook_*.sh` and `internal/cli/hook_*.fish` are the canonical sources and are embedded into the binary via `//go:embed` directives.

## Documentation

- [Configuration Guide](CONFIGURATION.md) - Worktree directory settings and configuration options
- [Detailed Installation Guide](INSTALL.md)

## License

MIT
