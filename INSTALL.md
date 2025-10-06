# wt Installation Guide

## Installation Steps

### 1. Install Binary

Install via Go:

```bash
go install github.com/toritori0318/git-wt/cmd/wt@latest
```

Or build from source:

```bash
git clone https://github.com/toritori0318/git-wt.git
cd git-wt
make build
sudo mv wt /usr/local/bin/
```

### 2. Setup Shell Function (Required)

To enable directory navigation with `wt go` and `--cd` flag (`wt new --cd`, `wt pr --cd`), add shell function to your configuration.

This setup dynamically loads the latest shell function from the binary, so updates are automatically reflected.

#### Bash

```bash
echo 'eval "$(wt hook bash)"' >> ~/.bashrc
source ~/.bashrc
```

#### Zsh

```bash
echo 'eval "$(wt hook zsh)"' >> ~/.zshrc
source ~/.zshrc
```

#### Fish

```bash
echo 'wt hook fish | source' >> ~/.config/fish/config.fish
exec fish
```

### 3. Verification

### Check Binary

```bash
# Display help
wt --help

# Check version (verify it's installed in GOPATH)
which wt
```

### Check Shell Function

```bash
# Verify shell function is loaded
type wt

# Expected output:
# wt is a function
# wt () { ... }

# If it displays "wt is /path/to/wt",
# the shell function is not loaded correctly
```

### Try It Out

```bash
# Try in an existing Git repository
cd /path/to/your/git/repo

# List worktrees
git worktree list

# Create a new worktree
wt new feature/test

# Verify worktree
git worktree list

# Try navigating (if shell function is working, it will actually cd)
wt go feature

# Check current location
pwd
# → Should display /path/to/your-repo-feature-test if working correctly
```

## 4. Optional Dependencies

### fzf (Recommended)

Installing fzf enables a comfortable selection UI for `wt go`/`wt clean`/`wt open`.

Works without it (falls back to numbered selection UI), but installation is recommended.

```bash
# macOS
brew install fzf

# Ubuntu/Debian
sudo apt install fzf

# Other Linux
git clone --depth 1 https://github.com/junegunn/fzf.git ~/.fzf
~/.fzf/install
```

### GitHub CLI (for `wt pr`)

Only required if using the PR review feature (`wt pr`).

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

## 7. Editor Configuration (Optional)

The editor used by `wt open` is determined in the following order:

1. `--editor` flag
2. `WT_EDITOR` environment variable
3. `VISUAL` environment variable
4. `EDITOR` environment variable
5. Auto-detection (code, idea, subl, vim, vi)

To set your preferred editor:

```bash
# Add to .bashrc / .zshrc
export WT_EDITOR=code
# or
export EDITOR=vim
```

## 5. Alternative Shell Function Setup Methods

### Using Local Binary (for Development)

If you built from source and want to use the local binary:

```bash
# Bash
cd /path/to/wt
echo 'eval "$(./wt hook bash)"' >> ~/.bashrc
source ~/.bashrc

# Zsh
cd /path/to/wt
echo 'eval "$(./wt hook zsh)"' >> ~/.zshrc
source ~/.zshrc

# Fish
cd /path/to/wt
echo './wt hook fish | source' >> ~/.config/fish/config.fish
exec fish
```

### Static Shell Function Files (Not Recommended)

You can also use static shell function files, but they won't auto-update when the binary changes:

```bash
# Bash/Zsh - Download and source
mkdir -p ~/.wt
curl -fsSL https://raw.githubusercontent.com/toritori0318/git-wt/main/shell/wt.sh -o ~/.wt/wt.sh
echo 'source ~/.wt/wt.sh' >> ~/.zshrc
source ~/.zshrc

# Fish - Download to functions directory
mkdir -p ~/.config/fish/functions
curl -fsSL https://raw.githubusercontent.com/toritori0318/git-wt/main/shell/wt.fish \
  -o ~/.config/fish/functions/wt.fish
exec fish
```

**Note:** Using `eval "$(wt hook <shell>)"` is recommended as it automatically reflects updates.

## 6. Troubleshooting

### `wt: command not found`

**Cause:** `$GOPATH/bin` is not in PATH

**Solution:**

```bash
# Check GOPATH
echo $(go env GOPATH)/bin

# Check if it's in PATH
echo $PATH | grep $(go env GOPATH)/bin

# If not included, add to shell configuration file
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.zshrc
source ~/.zshrc
```

### Can't navigate with `wt go`

**Cause:** Shell function is not loaded

**How to Check:**

```bash
type wt

# Expected output:
# wt is a function

# If it displays like this, shell function is not set:
# wt is /Users/xxx/go/bin/wt
```

**Solution:**

Execute "2. Setup Shell Function".

### `git: command not found`

**Solution:** Install git

```bash
# macOS
xcode-select --install
# or
brew install git

# Ubuntu/Debian
sudo apt install git
```

### `gh: command not found` (when using wt pr)

**Solution:** See "4. Install Optional Dependencies - GitHub CLI"

### fzf doesn't start

Works without fzf using numbered selection UI.
Installing fzf improves the experience (see "4. Install Optional Dependencies - fzf").

## 8. Uninstall

```bash
# Remove binary
rm /usr/local/bin/wt

# Remove shell function configuration
# Edit your shell config file and remove the eval "$(wt hook ...)" line:
# - ~/.bashrc or ~/.bash_profile (Bash)
# - ~/.zshrc (Zsh)
# - ~/.config/fish/config.fish (Fish)

# If using static shell function files, also remove:
rm ~/.wt.sh  # Bash/Zsh
rm ~/.config/fish/functions/wt.fish  # Fish
```

## Support

If you encounter issues, please report them on GitHub Issues:
https://github.com/toritori0318/git-wt/issues
