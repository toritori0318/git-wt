# wt - Git worktree helper
# Shell function: wt go, wt new, and wt pr cd into the worktree.
# wt new and wt pr stay put when --no-cd is passed.

# Return 0 when wt new / wt pr should print a path and the shell should cd.
# Help, version, and --no-cd stay in the current directory.
_wt_should_navigate() {
  local cmd=""
  local skip_value=0
  local arg
  for arg in "$@"; do
    if (( skip_value )); then
      skip_value=0
      continue
    fi
    case "$arg" in
      --)
        break
        ;;
      -h|--help|help|--version|--no-cd)
        return 1
        ;;
      --repo)
        skip_value=1
        ;;
      --repo=*)
        ;;
      -*)
        ;;
      *)
        if [[ -z "$cmd" ]]; then
          cmd="$arg"
        fi
        ;;
    esac
  done
  [[ "$cmd" == "new" || "$cmd" == "pr" ]]
}

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
  elif _wt_should_navigate "$@"; then
    # new/pr: capture the path and cd
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
