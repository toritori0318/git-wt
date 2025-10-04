# gwt Installation Guide

## Installation Steps

### 1. Install Binary

Install via Go:

```bash
go install github.com/toritsuyo/gwt@latest
```

Or build from source:

```bash
git clone https://github.com/toritsuyo/gwt.git
cd gwt
go build -o gwt ./cmd/gwt
sudo mv gwt /usr/local/bin/
```

### 2. Setup Shell Function (Required)

To enable directory navigation with `gwt go`, add shell function to your configuration.

This setup dynamically loads the latest shell function from the binary, so updates are automatically reflected.

#### Bash

```bash
echo 'eval "$(gwt hook bash)"' >> ~/.bashrc
source ~/.bashrc
```

#### Zsh

```bash
echo 'eval "$(gwt hook zsh)"' >> ~/.zshrc
source ~/.zshrc
```

#### Fish

```bash
echo 'gwt hook fish | source' >> ~/.config/fish/config.fish
exec fish
```

### 3. Verification

### Check Binary

```bash
# Display help
gwt --help

# Check version (verify it's installed in GOPATH)
which gwt
```

### Check Shell Function

```bash
# Verify shell function is loaded
type gwt

# Expected output:
# gwt is a function
# gwt () { ... }

# If it displays "gwt is /path/to/gwt",
# the shell function is not loaded correctly
```

### Try It Out

```bash
# Try in an existing Git repository
cd /path/to/your/git/repo

# List worktrees
git worktree list

# Create a new worktree
gwt new feature/test

# Verify worktree
git worktree list

# Try navigating (if shell function is working, it will actually cd)
gwt go feature

# Check current location
pwd
# → Should display /path/to/your-repo-feature-test if working correctly
```

## 4. Optional Dependencies

### fzf (Recommended)

Installing fzf enables a comfortable selection UI for `gwt go`/`gwt clean`/`gwt open`.

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

### GitHub CLI (for `gwt pr`)

Only required if using the PR review feature (`gwt pr`).

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

The editor used by `gwt open` is determined in the following order:

1. `--editor` flag
2. `GWT_EDITOR` environment variable
3. `VISUAL` environment variable
4. `EDITOR` environment variable
5. Auto-detection (code, idea, subl, vim, vi)

To set your preferred editor:

```bash
# Add to .bashrc / .zshrc
export GWT_EDITOR=code
# or
export EDITOR=vim
```

## 5. Alternative Shell Function Setup Methods

### Using Local Binary (for Development)

If you built from source and want to use the local binary:

```bash
# Bash
cd /path/to/gwt
echo 'eval "$(./gwt hook bash)"' >> ~/.bashrc
source ~/.bashrc

# Zsh
cd /path/to/gwt
echo 'eval "$(./gwt hook zsh)"' >> ~/.zshrc
source ~/.zshrc

# Fish
cd /path/to/gwt
echo './gwt hook fish | source' >> ~/.config/fish/config.fish
exec fish
```

### Static Shell Function Files (Not Recommended)

You can also use static shell function files, but they won't auto-update when the binary changes:

```bash
# Bash/Zsh - Download and source
mkdir -p ~/.gwt
curl -fsSL https://raw.githubusercontent.com/toritsuyo/gwt/main/shell/gwt.sh -o ~/.gwt/gwt.sh
echo 'source ~/.gwt/gwt.sh' >> ~/.zshrc
source ~/.zshrc

# Fish - Download to functions directory
mkdir -p ~/.config/fish/functions
curl -fsSL https://raw.githubusercontent.com/toritsuyo/gwt/main/shell/gwt.fish \
  -o ~/.config/fish/functions/gwt.fish
exec fish
```

**Note:** Using `eval "$(gwt hook <shell>)"` is recommended as it automatically reflects updates.

## 6. Troubleshooting

### `gwt: command not found`

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

### Can't navigate with `gwt go`

**Cause:** Shell function is not loaded

**How to Check:**

```bash
type gwt

# Expected output:
# gwt is a function

# If it displays like this, shell function is not set:
# gwt is /Users/xxx/go/bin/gwt
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

### `gh: command not found` (when using gwt pr)

**Solution:** See "4. Install Optional Dependencies - GitHub CLI"

### fzf doesn't start

Works without fzf using numbered selection UI.
Installing fzf improves the experience (see "4. Install Optional Dependencies - fzf").

## 8. Uninstall

```bash
# Remove binary
rm /usr/local/bin/gwt

# Remove shell function configuration
# Edit your shell config file and remove the eval "$(gwt hook ...)" line:
# - ~/.bashrc or ~/.bash_profile (Bash)
# - ~/.zshrc (Zsh)
# - ~/.config/fish/config.fish (Fish)

# If using static shell function files, also remove:
rm ~/.gwt.sh  # Bash/Zsh
rm ~/.config/fish/functions/gwt.fish  # Fish
```

## Support

If you encounter issues, please report them on GitHub Issues:
https://github.com/toritsuyo/gwt/issues
