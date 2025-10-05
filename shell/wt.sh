#!/bin/bash
# wt shell function (for bash/zsh)
#
# Installation:
#   1. Copy this file to ~/.wt.sh
#   2. Add the following to ~/.bashrc or ~/.zshrc:
#      source ~/.wt.sh
#   3. Restart shell: exec $SHELL
#
# Note: It's recommended to use `eval "$(wt hook bash)"` or `eval "$(wt hook zsh)"`
# instead of sourcing this file, as it ensures you always get the latest version.

# wt function
# Acts as shell function for wt go / any command with --cd to execute actual cd
# Other commands are delegated to binary
function wt() {
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

# Completion configuration (optional)
# For bash
if [[ -n "$BASH_VERSION" ]]; then
  _wt_completion() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    local prev="${COMP_WORDS[COMP_CWORD-1]}"

    # Subcommands
    if [[ $COMP_CWORD -eq 1 ]]; then
      COMPREPLY=($(compgen -W "new go clean open pr hook help" -- "$cur"))
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
fi

# For zsh
if [[ -n "$ZSH_VERSION" ]]; then
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
fi
