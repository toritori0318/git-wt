# wt - Git worktree helper
# Shell function: wt go / any command --cd executes actual cd

function wt
    # Set environment variable to indicate shell function is active
    set -gx WT_SHELL_FUNCTION 1
    if test (count $argv) -gt 0; and test $argv[1] = "go"
        set -e argv[1]
        # Fast-path: delegate help/version directly to binary
        for arg in $argv
            switch $arg
                case -h --help help --version
                    command wt go $argv
                    return $status
            end
        end

        set -l out (command wt go --quiet $argv)
        set -l code $status

        # If command failed, print output and return code
        if test $code -ne 0
            printf '%s\n' "$out"
            return $code
        end

        # Only cd when output is exactly one line and is a directory
        if test -n "$out"; and not string match -q '*\n*' "$out"; and test -d "$out"
            cd "$out"
        else
            # Not a path: show the output
            printf '%s\n' "$out"
        end
    else if contains -- --cd $argv
        # If --cd flag exists, get path and cd
        set -l out (command wt $argv)
        set -l code $status

        if test $code -ne 0
            printf '%s\n' "$out"
            return $code
        end

        if test -n "$out"; and not string match -q '*\n*' "$out"; and test -d "$out"
            cd "$out"
        else
            printf '%s\n' "$out"
        end
    else
        # Delegate other commands to binary
        command wt $argv
    end
end

# Completion configuration
complete -c wt -f

# Subcommand completion
complete -c wt -n "__fish_use_subcommand" -a "new" -d "Create new worktree"
complete -c wt -n "__fish_use_subcommand" -a "go" -d "Navigate between worktrees"
complete -c wt -n "__fish_use_subcommand" -a "clean" -d "Remove worktrees"
complete -c wt -n "__fish_use_subcommand" -a "open" -d "Open worktree in editor"
complete -c wt -n "__fish_use_subcommand" -a "pr" -d "Create worktree for PR review"
complete -c wt -n "__fish_use_subcommand" -a "hook" -d "Output shell hook scripts"
complete -c wt -n "__fish_use_subcommand" -a "help" -d "Show help"

# Branch name completion for wt go
complete -c wt -n "__fish_seen_subcommand_from go" -a "(git worktree list --porcelain 2>/dev/null | grep '^branch' | awk '{print \$2}' | sed 's|refs/heads/||')"
