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
      --repo|--base-dir|--branch|--remote)
        skip_value=1
        ;;
      --repo=*|--base-dir=*|--branch=*|--remote=*)
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

# Bash completion
_wt_completion() {
  local cur="${COMP_WORDS[COMP_CWORD]}"
  local prev="${COMP_WORDS[COMP_CWORD-1]}"

  # Subcommands
  if [[ $COMP_CWORD -eq 1 ]]; then
    COMPREPLY=($(compgen -W "new go clean open pr hook version help" -- "$cur"))
    return
  fi

  # Complete branch names for wt go
  if [[ "${COMP_WORDS[1]}" == "go" ]] && [[ $COMP_CWORD -eq 2 ]]; then
    local branches=$(git worktree list --porcelain 2>/dev/null | grep "^branch" | awk '{print $2}' | sed 's|refs/heads/||')
    COMPREPLY=($(compgen -W "$branches" -- "$cur"))
    return
  fi
}
complete -F _wt_completion wt
