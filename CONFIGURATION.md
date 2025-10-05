# Configuration Guide

## Configuration File Location

```
~/.config/wt/config.yaml
```

The configuration file follows the XDG Base Directory specification. If the `XDG_CONFIG_HOME` environment variable is set, that directory will be used.

## Basic Usage

```bash
# Display current settings
wt config list

# Get a specific setting value
wt config get worktree.directory_format

# Change a setting value
wt config set worktree.directory_format sibling

# Reset to defaults
wt config reset
```

## Configuration Options

### worktree.directory_format

Specifies how worktree directories are organized.

**Available values:**
- `subdirectory` (default)
- `sibling`

**Default value:** `subdirectory`

### worktree.subdirectory_suffix

Specifies the suffix to use in `subdirectory` mode.

**Default value:** `-wt`

**Constraint:** Must start with a hyphen `-`

## Directory Organization Modes

### Subdirectory Mode (Recommended, Default)

Places worktrees in a dedicated subdirectory. Multiple branches are easy to organize and manage.

**Directory structure:**

```
work/
├── myproject/                 # Main repository (main branch)
│   ├── .git/
│   ├── src/
│   └── README.md
└── myproject-wt/              # Worktrees directory
    ├── feature-login/         # feature/login branch
    │   ├── .git
    │   ├── src/
    │   └── README.md
    ├── bugfix-123/            # bugfix/123 branch
    │   ├── .git
    │   ├── src/
    │   └── README.md
    └── develop/               # develop branch
        ├── .git
        ├── src/
        └── README.md
```

**Benefits:**
- All worktrees are grouped in one directory
- Clear separation from the main repository
- Easy to navigate in IDEs and file managers

**Usage examples:**

```bash
# Use default suffix (-wt)
wt new feature/login
# → Creates ~/work/myproject-wt/feature-login

# Set custom suffix
wt config set worktree.subdirectory_suffix -worktrees
wt new feature/login
# → Creates ~/work/myproject-worktrees/feature-login

# Another custom suffix
wt config set worktree.subdirectory_suffix -branches
wt new develop
# → Creates ~/work/myproject-branches/develop
```

### Sibling Mode (Legacy)

Places worktrees at the same level as the main repository. This mode is closer to traditional git worktree behavior.

**Directory structure:**

```
work/
├── myproject/                 # Main repository (main branch)
│   ├── .git/
│   ├── src/
│   └── README.md
├── myproject-feature-login/   # feature/login branch
│   ├── .git
│   ├── src/
│   └── README.md
├── myproject-bugfix-123/      # bugfix/123 branch
│   ├── .git
│   ├── src/
│   └── README.md
└── myproject-develop/         # develop branch
    ├── .git
    ├── src/
    └── README.md
```

**Benefits:**
- Simple flat structure
- Similar to traditional git worktree behavior

**Drawbacks:**
- Directories can become cluttered as branches increase
- Hard to distinguish main repository from worktrees

**Usage examples:**

```bash
# Switch to sibling mode
wt config set worktree.directory_format sibling

wt new feature/login
# → Creates ~/work/myproject-feature-login

wt new develop
# → Creates ~/work/myproject-develop
```

## Switching Between Modes

### Switch to Subdirectory Mode

```bash
# Reset to default settings
wt config reset

# Or explicitly set
wt config set worktree.directory_format subdirectory
```

### Switch to Sibling Mode (Legacy)

```bash
wt config set worktree.directory_format sibling
```

## Manual Configuration File Editing

You can also edit the configuration file (`~/.config/wt/config.yaml`) directly.

**Example configuration file:**

```yaml
worktree:
  directory_format: subdirectory
  subdirectory_suffix: -wt
```

**Customization example:**

```yaml
worktree:
  directory_format: subdirectory
  subdirectory_suffix: -worktrees
```

**Legacy mode example:**

```yaml
worktree:
  directory_format: sibling
  subdirectory_suffix: -wt  # Not used in sibling mode
```

## Best Practices

### Choosing Based on Project Type

- **Personal projects (few branches)**: Sibling mode works well for simplicity
- **Team development (many branches)**: Subdirectory mode is easier to manage and organize

### Choosing a Custom Suffix

Choose a suffix that's short and clearly distinguishes from the project name:

```bash
# Recommended
-wt          # Default, short and clear
-worktrees   # Explicit and understandable
-branches    # Simple

# Not recommended
-my-worktrees-directory  # Too long
```

## Troubleshooting

### Configuration File Not Found

```bash
# Check configuration file path
wt config list
# → Configuration file: /Users/username/.config/wt/config.yaml (not found (using defaults))
```

The file is automatically created when you change a setting:

```bash
wt config set worktree.directory_format subdirectory
# → File is created
```

### Settings Not Applied

Verify that the YAML syntax in the configuration file is correct:

```bash
# Check current settings
wt config list

# Reset if there are issues
wt config reset
```

### Setting Values That Start with a Hyphen

`worktree.subdirectory_suffix` must start with `-`:

```bash
# Correct
wt config set worktree.subdirectory_suffix -wt
wt config set worktree.subdirectory_suffix -worktrees

# Error (no hyphen)
wt config set worktree.subdirectory_suffix wt
# → Error: subdirectory_suffix must start with '-'
```

## Additional Resources

- [README](README.md) - Basic usage guide
- [git worktree official documentation](https://git-scm.com/docs/git-worktree)
