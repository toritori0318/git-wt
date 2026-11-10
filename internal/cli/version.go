package cli

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), rootCmd.Version)
		},
	}
}

func init() {
	rootCmd.AddCommand(newVersionCmd())
}

// ResolveVersion fills blanks left by the linker with Go build info.
// A release build sets version, commit, and date with -ldflags.
// go install of a tagged module sets the version through build info instead.
func ResolveVersion(version, commit, date string) (string, string, string) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return version, commit, date
	}
	return fillVersionFromBuildInfo(version, commit, date, info)
}

func fillVersionFromBuildInfo(version, commit, date string, info *debug.BuildInfo) (string, string, string) {
	if info != nil && unsetVersion(version) {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			version = v
		}
	}
	if info == nil {
		return version, commit, date
	}
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			if unsetCommit(commit) && s.Value != "" {
				commit = shortRevision(s.Value)
			}
		case "vcs.time":
			if unsetDate(date) && s.Value != "" {
				date = s.Value
			}
		}
	}
	return version, commit, date
}

func shortRevision(rev string) string {
	if len(rev) > 7 {
		return rev[:7]
	}
	return rev
}

func unsetVersion(v string) bool { return v == "" || v == "dev" }
func unsetCommit(v string) bool  { return v == "" || v == "unknown" }
func unsetDate(v string) bool    { return v == "" || v == "unknown" }

// FormatVersion renders the string printed by `wt version` and `wt --version`.
func FormatVersion(version, commit, date string) string {
	var extra []string
	if !unsetCommit(commit) {
		extra = append(extra, "commit: "+commit)
	}
	if !unsetDate(date) {
		extra = append(extra, "built: "+date)
	}
	if len(extra) == 0 {
		return version
	}
	return fmt.Sprintf("%s (%s)", version, strings.Join(extra, ", "))
}
