# gwt - Git worktree helper
# Shell function: gwt go / any command --cd executes actual cd

function gwt() {
  if [[ "$1" == "go" ]]; then
    shift
    # Fast-path: delegate help/version directly to binary
    for arg in "$@"; do
      case "$arg" in
        -h|--help|help|--version)
          command gwt go "$@"
          return $?
          ;;
      esac
    done

    local out
    out="$(command gwt go --quiet "$@")"
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
    out="$(command gwt "$@")"
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
    command gwt "$@"
  fi
}

# Bash completion
_gwt_completion() {
  local cur="${COMP_WORDS[COMP_CWORD]}"
  local prev="${COMP_WORDS[COMP_CWORD-1]}"

  # Subcommands
  if [[ $COMP_CWORD -eq 1 ]]; then
    COMPREPLY=($(compgen -W "new go clean open pr hook help" -- "$cur"))
    return
  fi

  # Complete branch names for gwt go
  if [[ "${COMP_WORDS[1]}" == "go" ]] && [[ $COMP_CWORD -eq 2 ]]; then
    local branches=$(git worktree list --porcelain 2>/dev/null | grep "^branch" | awk '{print $2}' | sed 's|refs/heads/||')
    COMPREPLY=($(compgen -W "$branches" -- "$cur"))
    return
  fi
}
complete -F _gwt_completion gwt
