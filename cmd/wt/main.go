package main

import (
	"errors"
	"os"

	"github.com/toritori0318/git-wt/internal/cli"
)

var (
	// Version information (set by -ldflags during build)
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	cli.SetVersionInfo(version, commit, date)
	if err := cli.Execute(); err != nil {
		var exitCodeErr *cli.ExitCodeError
		if errors.As(err, &exitCodeErr) {
			os.Exit(exitCodeErr.Code)
		}
		os.Exit(1)
	}
}
