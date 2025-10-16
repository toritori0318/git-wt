# wt - Git worktree helper
# Shell function: wt go / any command --cd executes actual cd

function wt() {
  # Set environment variable to indicate shell function is active
  export WT_SHELL_FUNCTION=1
  if [[ "$1" == "go" ]]; then
    shift
    # Fast-path: delegate help/version directly to binary
    for arg in "$@"; do
      case "$arg" in
        -h|--help|help|--version)
          command wt go "$@"
          return $?
          ;;
      esac
    done

    local out
    out="$(command wt go --quiet "$@")"
    local code=$?

    # If command failed, print output and return code
    if (( code != 0 )); then
      printf '%s\n' "$out"
      return $code
    fi

    # Only cd when output is exactly one line and is a directory
    if [[ -n "$out" && "$out" != *$'\n'* && -d "$out" ]]; then
      builtin cd -- "$out" || return 1
    else
      # Not a path: show the output (help, usage, etc.)
      printf '%s\n' "$out"
    fi
  elif [[ "$*" == *"--cd"* ]]; then
    # If --cd flag exists, get path and cd
    local out
    out="$(command wt "$@")"
    local code=$?

    if (( code != 0 )); then
      printf '%s\n' "$out"
      return $code
    fi

    if [[ -n "$out" && "$out" != *$'\n'* && -d "$out" ]]; then
      builtin cd -- "$out" || return 1
    else
      printf '%s\n' "$out"
    fi
  else
    # Delegate other commands to binary
    command wt "$@"
  fi
}

# Zsh completion
_wt() {
  local -a subcmds
  subcmds=(
    'new:Create new worktree'
    'go:Navigate between worktrees'
    'clean:Remove worktrees'
    'open:Open worktree in editor'
    'pr:Create worktree for PR review'
    'hook:Output shell hook scripts'
    'help:Show help'
  )

  if (( CURRENT == 2 )); then
    _describe 'wt commands' subcmds
  elif (( CURRENT == 3 )) && [[ "${words[2]}" == "go" ]]; then
    local -a branches
    branches=(${(f)"$(git worktree list --porcelain 2>/dev/null | grep '^branch' | awk '{print $2}' | sed 's|refs/heads/||')"})
    _describe 'branches' branches
  fi
}

compdef _wt wt
